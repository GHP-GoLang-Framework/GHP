package gen

import (
	"fmt"
	"strings"

	"ghp/src/internal/ast"
)

// Switch writes n as a Go switch, one case per ast.Case plus an
// optional default, recursing into each branch's body through nodesFn -
// the same dispatch every node goes through, so any tag is allowed
// inside any branch.
//
// Two things come for free from emitting a real Go switch instead of
// something bespoke: a <go:case "a", "b"/> works with zero special
// handling, because Value is emitted verbatim after "case " and Go's own
// switch already accepts a comma-separated value list; and there's no
// fallthrough between cases unless the generated code said so explicitly
// (it never does), matching Go's own switch semantics.
//
// The two error returns only fire if a case's or default's body contains
// a node type nodesFn doesn't handle - with every current ast.Node type
// supported, that can't happen with real data today; they're there so a
// future unsupported tag fails loudly instead of being silently dropped.
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
