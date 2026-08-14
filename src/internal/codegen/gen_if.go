package codegen

import (
	"fmt"
	"strings"

	"ghp/src/internal/ast"
)

// genIf writes n as a Go if/else, recursing into Then and (when present)
// Else through generateNodes - the same dispatch every node goes through,
// so any tag is allowed inside either branch, including another <go:if>.
//
// n.Cond is emitted verbatim: whether or not the developer wrapped it in
// parens (e.g. to work around the '>' tag-closing ambiguity), the result
// is valid Go either way - `if (a > b) {` and `if a > b {` mean the same
// thing to the compiler.
func genIf(b *strings.Builder, file string, n *ast.If) error {
	fmt.Fprintf(b, "//line %s:%d\n", file, n.Line())
	fmt.Fprintf(b, "if %s {\n", n.Cond)

	if err := generateNodes(b, file, n.Then); err != nil {
		return err
	}

	if n.Else != nil {
		b.WriteString("} else {\n")
		if err := generateNodes(b, file, n.Else); err != nil {
			return err
		}
	}

	b.WriteString("}\n")
	return nil
}
