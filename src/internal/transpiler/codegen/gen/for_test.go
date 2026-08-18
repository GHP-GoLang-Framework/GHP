package gen

import (
	"strings"
	"testing"

	"ghp/src/internal/ast"
)

func TestFor(t *testing.T) {
	var b strings.Builder
	err := For(&b, ast.NewFor("i := 0; i < n; i++", nil, 1), noopNodes)
	if err != nil {
		t.Fatalf("For(): %v", err)
	}

	want := "for i := 0; i < n; i++ {\n}\n"
	if got := b.String(); got != want {
		t.Errorf("For() = %q, want %q", got, want)
	}
}

func TestForWithBody(t *testing.T) {
	var b strings.Builder
	body := []ast.Node{ast.NewText("hi", 2)}
	err := For(&b, ast.NewFor("_, item := range items", body, 1), noopNodes)
	if err != nil {
		t.Fatalf("For(): %v", err)
	}

	want := "for _, item := range items {\n" +
		"io.WriteString(w, \"hi\")\n" +
		"}\n"
	if got := b.String(); got != want {
		t.Errorf("For() = %q, want %q", got, want)
	}
}

func TestForBodyError(t *testing.T) {
	var b strings.Builder
	err := For(&b, ast.NewFor("i := 0; i < 1; i++", []ast.Node{nil}, 1), errNodes)
	if err == nil {
		t.Fatal("For() = nil error, want error from nodesFn")
	}
}
