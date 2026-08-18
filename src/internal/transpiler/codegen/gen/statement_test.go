package gen

import (
	"strings"
	"testing"

	"ghp/src/internal/ast"
)

func TestStatement(t *testing.T) {
	var b strings.Builder
	Statement(&b, ast.NewStatement("x := 1", 1))

	want := "x := 1\n"
	if got := b.String(); got != want {
		t.Errorf("Statement() = %q, want %q", got, want)
	}
}
