package transpiler

import (
	"fmt"
	"go/format"
	"os"
	"sort"
	"strconv"
	"strings"
)

type block struct {
	kind    string
	hasCase bool
}

type gen struct {
	body    strings.Builder
	imports map[string]struct{}
	blocks  []block
}

func Parse(src, funcName string) ([]byte, error) {
	g := newGen()

	if err := g.scan(src); err != nil {
		return nil, err
	}

	if len(g.blocks) > 0 {
		return nil, fmt.Errorf(
			"unclosed <go:%s> block",
			g.blocks[len(g.blocks)-1].kind,
		)
	}

	return g.generate(funcName)
}

func ParseFile(path, funcName string) ([]byte, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return Parse(string(src), funcName)
}

func newGen() *gen {
	return &gen{
		imports: map[string]struct{}{
			"io":       {},
			"net/http": {},
		},
	}
}

func (g *gen) scan(src string) error {
	for pos := 0; pos < len(src); {
		tag := strings.Index(src[pos:], "<go")

		if tag == -1 {
			g.writeHTML(src[pos:])
			return nil
		}

		tag += pos

		// "<go" pode simplesmente fazer parte do HTML:
		// <google>, <gopher>, etc.
		if !isGHPTag(src, tag) {
			g.writeHTML(src[pos : tag+1])
			pos = tag + 1
			continue
		}

		g.writeHTML(src[pos:tag])

		end, err := g.parseTag(src, tag)
		if err != nil {
			return fmt.Errorf("line %d: %w", lineAt(src, tag), err)
		}

		pos = end
	}

	return nil
}

func isGHPTag(src string, pos int) bool {
	if pos+3 > len(src) || src[pos:pos+3] != "<go" {
		return false
	}

	if pos+3 == len(src) {
		return true
	}

	switch src[pos+3] {
	case ' ', '=', ':':
		return true
	default:
		return false
	}
}

func (g *gen) parseTag(src string, start int) (int, error) {
	end := strings.Index(src[start:], "/>")

	if end == -1 {
		return 0, fmt.Errorf("tag not closed with '/>'")
	}

	end += start

	raw := strings.TrimSpace(src[start : end+2])
	tag, payload := splitTag(raw)

	if err := g.handleTag(tag, payload); err != nil {
		return 0, err
	}

	return end + 2, nil
}

func splitTag(tag string) (string, string) {
	tag = strings.TrimSuffix(tag, "/>")

	switch {
	case strings.HasPrefix(tag, "<go:"):
		i := strings.IndexByte(tag, ' ')
		if i == -1 {
			return tag, ""
		}

		return tag[:i], strings.TrimSpace(tag[i+1:])

	case strings.HasPrefix(tag, "<go="):
		return "<go=", strings.TrimSpace(strings.TrimPrefix(tag, "<go="))

	case strings.HasPrefix(tag, "<go "):
		return "<go", strings.TrimSpace(strings.TrimPrefix(tag, "<go "))

	default:
		return tag, ""
	}
}

func (g *gen) handleTag(tag, payload string) error {
	switch tag {
	case "<go":
		g.writeGo(payload)

	case "<go=":
		g.writeExpression(payload)

	case "<go:import":
		g.addImports(payload)

	case "<go:if":
		return g.openBlock("if", payload)

	case "<go:elif":
		return g.elif(payload)

	case "<go:else":
		return g.elseBlock()

	case "<go:endif":
		return g.closeBlock("if")

	case "<go:for":
		return g.openBlock("for", payload)

	case "<go:endfor":
		return g.closeBlock("for")

	case "<go:switch":
		return g.openSwitch(payload)

	case "<go:case":
		return g.caseBlock(payload)

	case "<go:default":
		return g.defaultBlock()

	case "<go:endswitch":
		return g.closeBlock("switch")

	default:
		return fmt.Errorf("unknown tag %q", tag)
	}

	return nil
}

func (g *gen) writeHTML(text string) {
	if text == "" || g.inSwitchBeforeCase() {
		return
	}

	text = cleanText(text)
	if text == "" {
		return
	}

	fmt.Fprintf(
		&g.body,
		"io.WriteString(w, %s)\n",
		strconv.Quote(text),
	)
}

func (g *gen) writeGo(code string) {
	if g.inSwitchBeforeCase() {
		return
	}

	g.body.WriteString(code)
	g.body.WriteByte('\n')
}

func (g *gen) writeExpression(expression string) {
	if g.inSwitchBeforeCase() {
		return
	}

	g.imports["fmt"] = struct{}{}
	g.imports["html"] = struct{}{}

	fmt.Fprintf(
		&g.body,
		"io.WriteString(w, html.EscapeString(fmt.Sprint(%s)))\n",
		expression,
	)
}

func (g *gen) addImports(payload string) {
	payload = strings.TrimSpace(payload)
	payload = strings.TrimPrefix(payload, "(")
	payload = strings.TrimSuffix(payload, ")")

	for _, path := range strings.Split(payload, ",") {
		path = strings.Trim(strings.TrimSpace(path), `"'`)

		if path != "" {
			g.imports[path] = struct{}{}
		}
	}
}

func (g *gen) openBlock(kind, condition string) error {
	if g.inSwitchBeforeCase() {
		return fmt.Errorf(
			"<go:%s> before the first <go:case> of a <go:switch>",
			kind,
		)
	}

	fmt.Fprintf(&g.body, "%s %s {\n", kind, condition)

	g.blocks = append(g.blocks, block{
		kind: kind,
	})

	return nil
}

func (g *gen) elif(condition string) error {
	if _, err := g.requireBlock("if"); err != nil {
		return err
	}

	fmt.Fprintf(&g.body, "} else if %s {\n", condition)
	return nil
}

func (g *gen) elseBlock() error {
	if _, err := g.requireBlock("if"); err != nil {
		return err
	}

	g.body.WriteString("} else {\n")
	return nil
}

func (g *gen) openSwitch(expression string) error {
	if g.inSwitchBeforeCase() {
		return fmt.Errorf(
			"<go:switch> before the first <go:case> of a <go:switch>",
		)
	}

	fmt.Fprintf(&g.body, "switch %s {\n", expression)

	g.blocks = append(g.blocks, block{
		kind: "switch",
	})

	return nil
}

func (g *gen) caseBlock(expression string) error {
	i, err := g.requireBlock("switch")
	if err != nil {
		return err
	}

	g.blocks[i].hasCase = true

	fmt.Fprintf(&g.body, "case %s:\n", expression)
	return nil
}

func (g *gen) defaultBlock() error {
	i, err := g.requireBlock("switch")
	if err != nil {
		return err
	}

	g.blocks[i].hasCase = true
	g.body.WriteString("default:\n")

	return nil
}

func (g *gen) closeBlock(kind string) error {
	if _, err := g.requireBlock(kind); err != nil {
		return err
	}

	g.blocks = g.blocks[:len(g.blocks)-1]
	g.body.WriteString("}\n")

	return nil
}

func (g *gen) requireBlock(kind string) (int, error) {
	if len(g.blocks) == 0 {
		return -1, fmt.Errorf(
			"no open <go:%s> block",
			kind,
		)
	}

	i := len(g.blocks) - 1

	if g.blocks[i].kind != kind {
		return -1, fmt.Errorf(
			"expected <go:%s>, found <go:%s>",
			g.blocks[i].kind,
			kind,
		)
	}

	return i, nil
}

func (g *gen) inSwitchBeforeCase() bool {
	if len(g.blocks) == 0 {
		return false
	}

	b := g.blocks[len(g.blocks)-1]

	return b.kind == "switch" && !b.hasCase
}

func (g *gen) generate(funcName string) ([]byte, error) {
	var src strings.Builder

	fmt.Fprintln(&src, "package pages")
	fmt.Fprintln(&src)
	fmt.Fprintln(&src, "import (")

	imports := make([]string, 0, len(g.imports))
	for path := range g.imports {
		imports = append(imports, path)
	}

	sort.Strings(imports)

	for _, path := range imports {
		fmt.Fprintf(&src, "\t%q\n", path)
	}

	fmt.Fprintln(&src, ")")
	fmt.Fprintln(&src)

	fmt.Fprintf(
		&src,
		"func %s(w http.ResponseWriter, r *http.Request) {\n",
		funcName,
	)

	src.WriteString(g.body.String())
	src.WriteString("}\n")

	formatted, err := format.Source([]byte(src.String()))
	if err != nil {
		return nil, fmt.Errorf(
			"codegen: generated invalid Go: %w\n%s",
			err,
			src.String(),
		)
	}

	return formatted, nil
}

var replacer = strings.NewReplacer("\t", "", "\n", "")

func cleanText(text string) string {
	text = replacer.Replace(text)

	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}

	return text
}

func lineAt(src string, offset int) int {
	return strings.Count(src[:offset], "\n") + 1
}
