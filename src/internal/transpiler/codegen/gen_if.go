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
//
// The two error returns only fire if Then/Else contain a node type
// generateNode doesn't handle - with every current ast.Node type
// supported, that can't happen with real data today; they're there so a
// future unsupported tag fails loudly instead of being silently dropped.
func genIf(b *strings.Builder, n *ast.If) error {
	fmt.Fprintf(b, "if %s {\n", n.Cond)

	if err := generateNodes(b, n.Then); err != nil {
		return err
	}

	if n.Else != nil {
		b.WriteString("} else {\n")
		if err := generateNodes(b, n.Else); err != nil {
			return err
		}
	}

	b.WriteString("}\n")
	return nil
}
