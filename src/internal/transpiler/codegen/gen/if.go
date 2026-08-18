package gen

import (
	"fmt"
	"strings"

	"ghp/src/internal/ast"
)

// If emits a Go if/else, recursing into Then and (when present) Else
// via nodesFn.
//
// Output:
//
//	if <n.Cond> {
//	  <nodesFn(n.Then)>
//	}
//	// optional:
//	} else {
//	  <nodesFn(n.Else)>
//	}
//
//	b       – destination buffer
//	n       – If node; Cond is the condition, Then/Else are branch bodies
//	nodesFn – callback that renders child nodes (breaks circular import)
func If(b *strings.Builder, n *ast.If, nodesFn NodesFunc) error {
	fmt.Fprintf(b, "if %s {\n", n.Cond)

	if err := nodesFn(b, n.Then); err != nil {
		return err
	}

	if n.Else != nil {
		b.WriteString("} else {\n")
		if err := nodesFn(b, n.Else); err != nil {
			return err
		}
	}

	b.WriteString("}\n")
	return nil
}
