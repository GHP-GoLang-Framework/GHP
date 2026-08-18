package gen

import (
	"strings"
	"testing"

	"ghp/src/internal/ast"
)

func TestText(t *testing.T) {
	var b strings.Builder
	Text(&b, ast.NewText("hello world", 1))

	want := "io.WriteString(w, \"hello world\")\n"
	if got := b.String(); got != want {
		t.Errorf("Text() = %q, want %q", got, want)
	}
}
