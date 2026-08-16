package codegen

import (
	"go/parser"
	"go/token"
	"testing"

	"ghp/src/internal/ast"
)

func TestGenerateIfWithoutElse(t *testing.T) {
	nodes := []ast.Node{
		ast.NewIf("a == b", []ast.Node{ast.NewText("yes", 2)}, nil, 1),
	}

	got, err := Generate(nodes)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	want := "if a == b {\n" +
		"io.WriteString(w, \"yes\")\n" +
		"}\n"
	if got != want {
		t.Errorf("Generate() = %q, want %q", got, want)
	}
}

func TestGenerateIfWithElse(t *testing.T) {
	nodes := []ast.Node{
		ast.NewIf("a == b",
			[]ast.Node{ast.NewText("yes", 2)},
			[]ast.Node{ast.NewText("no", 3)},
			1),
	}

	got, err := Generate(nodes)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	want := "if a == b {\n" +
		"io.WriteString(w, \"yes\")\n" +
		"} else {\n" +
		"io.WriteString(w, \"no\")\n" +
		"}\n"
	if got != want {
		t.Errorf("Generate() = %q, want %q", got, want)
	}
}

func TestGenerateNestedIfProducesValidGo(t *testing.T) {
	// End-to-end proof: go:if inside go:if (the same feature the
	// parser already tests in TestParseNesting) has to come out as
	// real syntactically valid Go, not just match as a string.
	inner := ast.NewIf("b", []ast.Node{ast.NewText("x", 2)}, nil, 2)
	outer := ast.NewIf("a", []ast.Node{inner}, []ast.Node{ast.NewText("y", 4)}, 1)

	body, err := Generate([]ast.Node{outer})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	src := "package p\n\nimport \"io\"\n\nfunc f(w io.Writer, a, b bool) {\n" + body + "}\n"

	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "generated.go", src, 0); err != nil {
		t.Fatalf("generated code is not valid Go: %v\n---\n%s", err, src)
	}
}
