// Package codegen converts the tree defined by internal/ast into the Go
// source that makes up the body of a page's handler function.
//
// Generate does not produce a compilable file by itself: it only emits the
// statements that go inside the function body. The generated statements
// assume:
//
//   - a variable named w, implementing io.Writer, is already in scope
//   - the packages fmt, html and io are imported in the file that embeds
//     this output
//
// Wiring the function signature, those imports and the final import(...)
// block together (collecting every *ast.Import in the file) is the job of
// whoever assembles the full .go file, not this package.
package codegen

import (
	"fmt"
	"strings"

	"ghp/src/internal/ast"
)

// Generate walks nodes in source order and returns the Go statements that
// render them, one node at a time. file is the original .ghp path, used to
// emit a `//line file:N` directive before each node's output - so a
// compile error in the generated .go, or a debugger stepping through it,
// points back at the line the developer actually wrote.
func Generate(file string, nodes []ast.Node) (string, error) {
	var b strings.Builder
	if err := generateNodes(&b, file, nodes); err != nil {
		return "", err
	}
	return b.String(), nil
}

// generateNodes writes each node in nodes to b, in order. Besides backing
// Generate itself, this is what a tag with a nested body (<go:if>,
// <go:switch>, <go:for>) calls to render its own Then/Else/Cases/Body -
// see gen_if.go and gen_for.go for examples.
func generateNodes(b *strings.Builder, file string, nodes []ast.Node) error {
	for _, n := range nodes {
		if err := generateNode(b, file, n); err != nil {
			return err
		}
	}
	return nil
}

// generateNode dispatches a single node to its generator by concrete type.
// This is the extension point future tags plug into: each adds its own
// case here and implements the generator in its own gen_*.go file, so two
// tags being built at the same time only ever touch this one shared line
// each, not each other's code.
func generateNode(b *strings.Builder, file string, n ast.Node) error {
	switch node := n.(type) {
	case *ast.Import:
		// Imports don't render anything in the function body - GHP-11
		// collects every *ast.Import in the file separately, to build
		// the import(...) block once, deduplicated.
		return nil
	case *ast.Text:
		genText(b, file, node)
	case *ast.Echo:
		genEcho(b, file, node)
	case *ast.Statement:
		genStatement(b, file, node)
	case *ast.If:
		return genIf(b, file, node)
	case *ast.For:
		return genFor(b, file, node)
	default:
		return fmt.Errorf("codegen: no generator registered for %T (line %d)", n, n.Line())
	}
	return nil
}
