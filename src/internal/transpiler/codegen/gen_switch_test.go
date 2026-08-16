package codegen

import (
	"go/parser"
	"go/token"
	"testing"

	"ghp/src/internal/ast"
)

func TestGenerateSwitchWithDefault(t *testing.T) {
	nodes := []ast.Node{
		ast.NewSwitch("v",
			[]ast.Case{
				{Value: "1", Body: []ast.Node{ast.NewText("one", 2)}, Line: 2},
				{Value: "2", Body: []ast.Node{ast.NewText("two", 3)}, Line: 3},
			},
			[]ast.Node{ast.NewText("other", 4)},
			1),
	}

	got, err := Generate(nodes)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	want := "switch v {\n" +
		"case 1:\n" +
		"io.WriteString(w, \"one\")\n" +
		"case 2:\n" +
		"io.WriteString(w, \"two\")\n" +
		"default:\n" +
		"io.WriteString(w, \"other\")\n" +
		"}\n"
	if got != want {
		t.Errorf("Generate() = %q, want %q", got, want)
	}
}

func TestGenerateSwitchWithoutDefault(t *testing.T) {
	nodes := []ast.Node{
		ast.NewSwitch("v",
			[]ast.Case{{Value: "1", Body: nil, Line: 2}},
			nil,
			1),
	}

	got, err := Generate(nodes)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	want := "switch v {\n" +
		"case 1:\n" +
		"}\n"
	if got != want {
		t.Errorf("Generate() = %q, want %q", got, want)
	}
}

func TestGenerateSwitchMultiValueCase(t *testing.T) {
	// GHP-9 decision: a case accepts several comma-separated values,
	// just like Go's switch - Value already comes ready from the
	// parser (see parse_test.go), genSwitch only has to place it after
	// "case ".
	nodes := []ast.Node{
		ast.NewSwitch("v", []ast.Case{{Value: `"a", "b"`, Body: nil, Line: 1}}, nil, 1),
	}

	got, err := Generate(nodes)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	want := "switch v {\n" +
		"case \"a\", \"b\":\n" +
		"}\n"
	if got != want {
		t.Errorf("Generate() = %q, want %q", got, want)
	}
}

func TestGenerateSwitchProducesValidGo(t *testing.T) {
	// End-to-end proof, in the same spirit as the equivalent test for
	// go:if: the generated Go is really compiled with go/parser, not
	// just compared as a string.
	nodes := []ast.Node{
		ast.NewSwitch("v",
			[]ast.Case{
				{Value: `"a", "b"`, Body: []ast.Node{ast.NewText("one", 2)}, Line: 2},
			},
			[]ast.Node{ast.NewEcho("v", 3)},
			1),
	}

	body, err := Generate(nodes)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	src := "package p\n\n" +
		"import (\n\t\"fmt\"\n\t\"html\"\n\t\"io\"\n)\n\n" +
		"func f(w io.Writer, v string) {\n" + body + "}\n"

	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "generated.go", src, 0); err != nil {
		t.Fatalf("generated code is not valid Go: %v\n---\n%s", err, src)
	}
}
