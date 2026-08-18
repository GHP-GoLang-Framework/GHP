package gen

import (
	"fmt"
	"strings"

	"ghp/src/internal/ast"
)

// NodesFunc renders a slice of AST nodes into Go source. The concrete
// implementation lives in the codegen package and is passed here to
// avoid a circular import between gen/ and codegen/.
type NodesFunc func(b *strings.Builder, nodes []ast.Node) error

// For emits a Go for loop, recursing into Body via nodesFn.
//
// Output:
//
//	for <n.Expr> {
//	  <nodesFn(n.Body)>
//	}
//
//	b       – destination buffer
//	n       – For node; Expr is the loop header, Body is the loop body
//	nodesFn – callback that renders child nodes (breaks circular import)
func For(b *strings.Builder, n *ast.For, nodesFn NodesFunc) error {
	fmt.Fprintf(b, "for %s {\n", n.Expr)

	if err := nodesFn(b, n.Body); err != nil {
		return err
	}

	b.WriteString("}\n")
	return nil
}
