package codegen

import (
	"testing"

	"ghp/src/internal/ast"
)

func TestGenerateStatement(t *testing.T) {
	nodes := []ast.Node{ast.NewStatement("x := 1", 2)}

	got, err := Generate("page.ghp", nodes)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	want := "//line page.ghp:2\nx := 1\n"
	if got != want {
		t.Errorf("Generate() = %q, want %q", got, want)
	}
}
