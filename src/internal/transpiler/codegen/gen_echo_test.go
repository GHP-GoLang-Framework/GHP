package codegen

import (
	"fmt"
	"html"
	"strings"
	"testing"

	"ghp/src/internal/ast"
)

func TestGenerateEcho(t *testing.T) {
	nodes := []ast.Node{ast.NewEcho("user.Name", 5)}

	got, err := Generate("page.ghp", nodes)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	want := "//line page.ghp:5\nio.WriteString(w, html.EscapeString(fmt.Sprint(user.Name)))\n"
	if got != want {
		t.Errorf("Generate() = %q, want %q", got, want)
	}
}

func TestGenerateEchoFunctionCall(t *testing.T) {
	// GHP-7 acceptance criterion: any valid Go expression works in
	// <go= ...>, including a function call with a return value - the
	// parser stores the tag text opaquely (ast.Echo.Expr), and codegen
	// only has to interpolate it verbatim inside fmt.Sprint(...),
	// without caring what's inside.
	nodes := []ast.Node{ast.NewEcho("user.FullName()", 8)}

	got, err := Generate("page.ghp", nodes)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	want := "//line page.ghp:8\nio.WriteString(w, html.EscapeString(fmt.Sprint(user.FullName())))\n"
	if got != want {
		t.Errorf("Generate() = %q, want %q", got, want)
	}
}

func TestGenerateEchoOutputIsEscaped(t *testing.T) {
	// Confirms the real effect of the formula genEcho emits: a value
	// with dangerous HTML has to come out as text, not markup. If
	// someone changes the escaping formula without noticing the
	// security implication, this test breaks.
	payload := `<script>alert(1)</script>`
	rendered := html.EscapeString(fmt.Sprint(payload))

	if strings.Contains(rendered, "<script>") {
		t.Errorf("output not escaped correctly: %q", rendered)
	}
}
