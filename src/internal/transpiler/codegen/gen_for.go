package codegen

import (
	"fmt"
	"strings"

	"ghp/src/internal/ast"
)

// genFor writes n as a Go for loop, recursing into Body through
// generateNodes once per iteration at runtime - the loop itself only
// appears once in the generated source, same as it would in hand-written
// Go.
//
// n.Expr is emitted verbatim as the loop header: whatever form the
// developer wrote - `i := 0; i < n; i++`, `_, item := range slice`,
// `k, v := range m` - becomes the real `for` header in the generated
// code, so any variable it declares (item, k, v...) is naturally in scope
// for the tags inside Body, like <go= item.Name>. There's nothing for
// codegen to do to make that work; it falls out of emitting real Go.
//
// The error return only fires if Body contains a node type generateNode
// doesn't handle - with every current ast.Node type supported, that can't
// happen with real data today; it's there so a future unsupported tag
// fails loudly instead of being silently dropped.
func genFor(b *strings.Builder, file string, n *ast.For) error {
	fmt.Fprintf(b, "//line %s:%d\n", file, n.Line())
	fmt.Fprintf(b, "for %s {\n", n.Expr)

	if err := generateNodes(b, file, n.Body); err != nil {
		return err
	}

	b.WriteString("}\n")
	return nil
}
