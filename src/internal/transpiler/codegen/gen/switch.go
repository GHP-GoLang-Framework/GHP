package gen

import (
	"fmt"
	"strings"

	"ghp/src/internal/ast"
)

// Switch emits a Go switch with one case per ast.Case and an optional
// default, recursing into each branch body via nodesFn.
//
// Output:
//
//	switch <n.Expr> {
//	case <value>:
//	  <nodesFn(body)>
//	// ...
//	default:
//	  <nodesFn(n.Default)>  // when present
//	}
//
//	b       – destination buffer
//	n       – Switch node; Expr is the switched value, Cases/Default are branches
//	nodesFn – callback that renders child nodes (breaks circular import)
func Switch(b *strings.Builder, n *ast.Switch, nodesFn NodesFunc) error {
	fmt.Fprintf(b, "switch %s {\n", n.Expr)

	for _, c := range n.Cases {
		fmt.Fprintf(b, "case %s:\n", c.Value)
		if err := nodesFn(b, c.Body); err != nil {
			return err
		}
	}

	if n.Default != nil {
		b.WriteString("default:\n")
		if err := nodesFn(b, n.Default); err != nil {
			return err
		}
	}

	b.WriteString("}\n")
	return nil
}
