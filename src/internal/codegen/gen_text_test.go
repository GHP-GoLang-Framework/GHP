package codegen

import (
	"strconv"
	"strings"
	"testing"

	"ghp/src/internal/ast"
)

func TestGenerateText(t *testing.T) {
	nodes := []ast.Node{ast.NewText("hello world", 3)}

	got, err := Generate("page.ghp", nodes)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	want := "//line page.ghp:3\nio.WriteString(w, \"hello world\")\n"
	if got != want {
		t.Errorf("Generate() = %q, want %q", got, want)
	}
}

func TestGenerateTextEscapesSpecialChars(t *testing.T) {
	// Quotes, newlines and backslashes in the source HTML cannot leak
	// into the Go literal without escaping, or the generated .go will
	// not even compile. Instead of writing the expected string by hand
	// (fragile and hard to read), we pull the generated literal back
	// out with strconv.Unquote and check that it matches the original
	// value.
	value := "line 1\n\"quotes\" and \\backslash"
	nodes := []ast.Node{ast.NewText(value, 1)}

	got, err := Generate("p.ghp", nodes)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	start := strings.Index(got, "io.WriteString(w, ") + len("io.WriteString(w, ")
	end := strings.LastIndex(got, ")")
	literal := got[start:end]

	decoded, err := strconv.Unquote(literal)
	if err != nil {
		t.Fatalf("generated literal is not a valid Go string literal: %v (%q)", err, literal)
	}
	if decoded != value {
		t.Errorf("decoded = %q, want %q", decoded, value)
	}
}
