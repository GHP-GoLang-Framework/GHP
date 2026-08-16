package codegen

import (
	"go/parser"
	"go/token"
	"testing"

	"ghp/src/internal/ast"
)

func TestGenerateForRangeSlice(t *testing.T) {
	nodes := []ast.Node{
		ast.NewFor("_, item := range items",
			[]ast.Node{ast.NewEcho("item.Name", 2)}, 1),
	}

	got, err := Generate("p.ghp", nodes)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	want := "//line p.ghp:1\n" +
		"for _, item := range items {\n" +
		"//line p.ghp:2\n" +
		"io.WriteString(w, html.EscapeString(fmt.Sprint(item.Name)))\n" +
		"}\n"
	if got != want {
		t.Errorf("Generate() = %q, want %q", got, want)
	}
}

func TestGenerateForRangeMap(t *testing.T) {
	nodes := []ast.Node{
		ast.NewFor("key, value := range m",
			[]ast.Node{ast.NewEcho("key", 2), ast.NewEcho("value", 2)}, 1),
	}

	got, err := Generate("p.ghp", nodes)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	want := "//line p.ghp:1\n" +
		"for key, value := range m {\n" +
		"//line p.ghp:2\n" +
		"io.WriteString(w, html.EscapeString(fmt.Sprint(key)))\n" +
		"//line p.ghp:2\n" +
		"io.WriteString(w, html.EscapeString(fmt.Sprint(value)))\n" +
		"}\n"
	if got != want {
		t.Errorf("Generate() = %q, want %q", got, want)
	}
}

func TestGenerateForTraditionalIndex(t *testing.T) {
	nodes := []ast.Node{
		ast.NewFor("i := 0; i < n; i++", nil, 1),
	}

	got, err := Generate("p.ghp", nodes)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	want := "//line p.ghp:1\n" +
		"for i := 0; i < n; i++ {\n" +
		"}\n"
	if got != want {
		t.Errorf("Generate() = %q, want %q", got, want)
	}
}

func TestGenerateNestedForProducesValidGo(t *testing.T) {
	// End-to-end proof: go:for inside go:for (the same scenario the
	// parser covers in TestParseNesting) compiled for real with
	// go/parser, not just compared as a string.
	inner := ast.NewFor("_, cell := range row", []ast.Node{ast.NewEcho("cell", 2)}, 2)
	outer := ast.NewFor("_, row := range matrix", []ast.Node{inner}, 1)

	body, err := Generate("p.ghp", []ast.Node{outer})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	src := "package p\n\n" +
		"import (\n\t\"fmt\"\n\t\"html\"\n\t\"io\"\n)\n\n" +
		"func f(w io.Writer, matrix [][]string) {\n" + body + "}\n"

	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "generated.go", src, 0); err != nil {
		t.Fatalf("generated code is not valid Go: %v\n---\n%s", err, src)
	}
}
