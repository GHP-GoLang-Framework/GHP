package gen

import (
	"fmt"
	"strings"

	"ghp/src/internal/ast"
)

// noopNodes is a NodesFunc that renders each node with its minimal output,
// enough to make For/If/Switch body tests pass without importing codegen.
func noopNodes(b *strings.Builder, nodes []ast.Node) error {
	for _, n := range nodes {
		switch node := n.(type) {
		case *ast.Text:
			Text(b, node)
		case *ast.Echo:
			Echo(b, node)
		case *ast.Statement:
			Statement(b, node)
		default:
			return fmt.Errorf("test: unsupported node %T", n)
		}
	}
	return nil
}

// errNodes is a NodesFunc that always returns an error.
func errNodes(_ *strings.Builder, _ []ast.Node) error {
	return fmt.Errorf("nodesFn error")
}
