package codegen

import (
	"testing"

	"ghp/src/internal/ast"
)

func TestGenerateStatement(t *testing.T) {
	nodes := []ast.Node{ast.NewStatement("x := 1", 2)}

	got, err := Generate(nodes)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	want := "x := 1\n"
	if got != want {
		t.Errorf("Generate() = %q, want %q", got, want)
	}
}
