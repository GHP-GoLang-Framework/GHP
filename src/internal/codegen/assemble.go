package codegen

import (
	"fmt"
	"go/format"
	"strings"

	"ghp/src/internal/ast"
)

// Assemble builds a complete, compilable .go source file for one page:
// a package declaration, only the imports the page actually needs, and
// an http.HandlerFunc-shaped function (named funcName) whose body comes
// from Generate.
//
// ghpFile is the original .ghp path, threaded through to Generate for its
// //line directives. The result is run through go/format before being
// returned - both for idiomatic formatting and as a cheap correctness
// check: if the assembled source isn't valid Go (e.g. funcName isn't a
// valid identifier), this reports that as an error here instead of
// handing back something that would only fail later, confusingly, at
// `go build`.
func Assemble(pkg, funcName, ghpFile string, prog *ast.Program) (string, error) {
	body, err := Generate(ghpFile, prog.Nodes)
	if err != nil {
		return "", err
	}

	var need neededImports
	scanNeededImports(prog.Nodes, &need)

	var b strings.Builder
	fmt.Fprintf(&b, "package %s\n\n", pkg)
	writeImports(&b, need, collectImports(prog.Nodes))
	fmt.Fprintf(&b, "func %s(w http.ResponseWriter, r *http.Request) {\n", funcName)
	b.WriteString(body)
	b.WriteString("}\n")

	formatted, err := format.Source([]byte(b.String()))
	if err != nil {
		return "", fmt.Errorf("codegen: arquivo montado nao e Go valido: %w", err)
	}
	return string(formatted), nil
}

// neededImports tracks which of the packages genEcho/genText depend on
// are actually exercised somewhere in the page - importing fmt/html/io
// unconditionally would fail to compile on a page with no <go=...> or
// plain text anywhere in it (an unlikely but possible page made up only
// of <go ...>/<go:for>/etc).
type neededImports struct {
	io, html, fmt bool
}

// scanNeededImports walks nodes (recursing into every tag with a body)
// looking for *ast.Text and *ast.Echo, and sets the corresponding fields
// on need. It never returns early: a single pass has to see the whole
// tree, since any branch might be the only one using io/html/fmt.
func scanNeededImports(nodes []ast.Node, need *neededImports) {
	for _, n := range nodes {
		switch node := n.(type) {
		case *ast.Text:
			need.io = true
		case *ast.Echo:
			need.io, need.html, need.fmt = true, true, true
		case *ast.If:
			scanNeededImports(node.Then, need)
			scanNeededImports(node.Else, need)
		case *ast.Switch:
			for _, c := range node.Cases {
				scanNeededImports(c.Body, need)
			}
			scanNeededImports(node.Default, need)
		case *ast.For:
			scanNeededImports(node.Body, need)
		}
	}
}

// collectImports gathers every ast.ImportPath declared via <go:import>
// at the top level of nodes, deduplicated by Path (the first alias seen
// for a repeated path wins). It doesn't recurse into <go:if>/<go:switch>/
// <go:for> bodies - a conditional import isn't a thing Go supports, so a
// <go:import> nested inside one of those wouldn't mean anything.
func collectImports(nodes []ast.Node) []ast.ImportPath {
	seen := make(map[string]bool)
	var paths []ast.ImportPath

	for _, n := range nodes {
		imp, ok := n.(*ast.Import)
		if !ok {
			continue
		}
		for _, p := range imp.Paths {
			if seen[p.Path] {
				continue
			}
			seen[p.Path] = true
			paths = append(paths, p)
		}
	}

	return paths
}

// writeImports writes the import(...) block: net/http always (the
// handler signature needs it), fmt/html/io only when need says the page
// actually uses them, then every path collected from the page's own
// <go:import> tags - skipping any that duplicate one of the automatic
// ones above, since Go doesn't allow importing the same path twice in
// one file.
func writeImports(b *strings.Builder, need neededImports, userImports []ast.ImportPath) {
	auto := []string{"net/http"}
	if need.fmt {
		auto = append(auto, "fmt")
	}
	if need.html {
		auto = append(auto, "html")
	}
	if need.io {
		auto = append(auto, "io")
	}

	b.WriteString("import (\n")
	autoSet := make(map[string]bool, len(auto))
	for _, path := range auto {
		autoSet[path] = true
		fmt.Fprintf(b, "\t%q\n", path)
	}

	for _, p := range userImports {
		if autoSet[p.Path] {
			continue
		}
		if p.Alias != "" {
			fmt.Fprintf(b, "\t%s %q\n", p.Alias, p.Path)
		} else {
			fmt.Fprintf(b, "\t%q\n", p.Path)
		}
	}
	b.WriteString(")\n\n")
}
