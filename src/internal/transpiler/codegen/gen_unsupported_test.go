package codegen

import (
	"testing"

	"ghp/src/internal/ast"
)

// TestGenerateErrorsOnUnsupportedNode exercises the defensive error
// branches in Generate and the gen_* emitters. A nil node can't come
// from the parser - every ast.Node it produces is handled by
// generateNode's switch (and ast.Node can only be implemented inside
// package ast), so these paths only fire if a future node type is
// added to ast without a matching case here. Each one must fail loudly
// instead of being silently dropped.
//
// Ex: a nil node nested in a <go:if> body -> Generate reports an error.
func TestGenerateErrorsOnUnsupportedNode(t *testing.T) {
	tests := []struct {
		name  string
		nodes []ast.Node
	}{
		{"top level", []ast.Node{nil}},
		{"if then body", []ast.Node{ast.NewIf("true", []ast.Node{nil}, nil, 1)}},
		{"if else body", []ast.Node{ast.NewIf("true", nil, []ast.Node{nil}, 1)}},
		{"for body", []ast.Node{ast.NewFor("i := 0; i < 1; i++", []ast.Node{nil}, 1)}},
		{"switch case body", []ast.Node{ast.NewSwitch("v", []ast.Case{{Value: "1", Body: []ast.Node{nil}, Line: 1}}, nil, 1)}},
		{"switch default body", []ast.Node{ast.NewSwitch("v", nil, []ast.Node{nil}, 1)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Generate(tt.nodes); err == nil {
				t.Errorf("Generate(%s) = nil error, want unsupported node type error", tt.name)
			}
		})
	}
}
