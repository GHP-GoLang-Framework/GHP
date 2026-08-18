package gen

import (
	"strings"
	"testing"

	"ghp/src/internal/ast"
)

func TestIfWithoutElse(t *testing.T) {
	var b strings.Builder
	err := If(&b, ast.NewIf("a == b", []ast.Node{ast.NewText("yes", 2)}, nil, 1), noopNodes)
	if err != nil {
		t.Fatalf("If(): %v", err)
	}

	want := "if a == b {\n" +
		"io.WriteString(w, \"yes\")\n" +
		"}\n"
	if got := b.String(); got != want {
		t.Errorf("If() = %q, want %q", got, want)
	}
}

func TestIfWithElse(t *testing.T) {
	var b strings.Builder
	err := If(&b, ast.NewIf("a == b",
		[]ast.Node{ast.NewText("yes", 2)},
		[]ast.Node{ast.NewText("no", 3)},
		1), noopNodes)
	if err != nil {
		t.Fatalf("If(): %v", err)
	}

	want := "if a == b {\n" +
		"io.WriteString(w, \"yes\")\n" +
		"} else {\n" +
		"io.WriteString(w, \"no\")\n" +
		"}\n"
	if got := b.String(); got != want {
		t.Errorf("If() = %q, want %q", got, want)
	}
}

func TestIfThenBodyError(t *testing.T) {
	var b strings.Builder
	err := If(&b, ast.NewIf("true", []ast.Node{nil}, nil, 1), errNodes)
	if err == nil {
		t.Fatal("If() = nil error, want error from nodesFn")
	}
}

func TestIfElseBodyError(t *testing.T) {
	var b strings.Builder
	err := If(&b, ast.NewIf("true", nil, []ast.Node{nil}, 1), errNodes)
	if err == nil {
		t.Fatal("If() = nil error, want error from nodesFn")
	}
}
