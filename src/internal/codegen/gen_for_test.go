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
			[]ast.Node{ast.NewEcho("item.Nome", 2)}, 1),
	}

	got, err := Generate("p.ghp", nodes)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	want := "//line p.ghp:1\n" +
		"for _, item := range items {\n" +
		"//line p.ghp:2\n" +
		"io.WriteString(w, html.EscapeString(fmt.Sprint(item.Nome)))\n" +
		"}\n"
	if got != want {
		t.Errorf("Generate() = %q, want %q", got, want)
	}
}

func TestGenerateForRangeMap(t *testing.T) {
	nodes := []ast.Node{
		ast.NewFor("chave, valor := range mapa",
			[]ast.Node{ast.NewEcho("chave", 2), ast.NewEcho("valor", 2)}, 1),
	}

	got, err := Generate("p.ghp", nodes)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	want := "//line p.ghp:1\n" +
		"for chave, valor := range mapa {\n" +
		"//line p.ghp:2\n" +
		"io.WriteString(w, html.EscapeString(fmt.Sprint(chave)))\n" +
		"//line p.ghp:2\n" +
		"io.WriteString(w, html.EscapeString(fmt.Sprint(valor)))\n" +
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

func TestGenerateForPropagatesErrorFromBody(t *testing.T) {
	// go:switch ainda nao tem gerador nesta branch (GHP-9) - um go:for
	// cujo corpo contem um go:switch tem que propagar o erro.
	nodes := []ast.Node{
		ast.NewFor("_, x := range xs", []ast.Node{ast.NewSwitch("x", nil, nil, 2)}, 1),
	}

	_, err := Generate("p.ghp", nodes)
	if err == nil {
		t.Fatal("Generate() = nil error, want erro propagado de dentro do corpo do for")
	}
}

func TestGenerateNestedForProducesValidGo(t *testing.T) {
	// Prova de ponta a ponta: go:for dentro de go:for (o mesmo cenario
	// que o parser cobre em TestParseNesting) compilado de verdade com
	// go/parser, nao so comparado como string.
	inner := ast.NewFor("_, cel := range linha", []ast.Node{ast.NewEcho("cel", 2)}, 2)
	outer := ast.NewFor("_, linha := range matriz", []ast.Node{inner}, 1)

	body, err := Generate("p.ghp", []ast.Node{outer})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	src := "package p\n\n" +
		"import (\n\t\"fmt\"\n\t\"html\"\n\t\"io\"\n)\n\n" +
		"func f(w io.Writer, matriz [][]string) {\n" + body + "}\n"

	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "gerado.go", src, 0); err != nil {
		t.Fatalf("codigo gerado nao e Go valido: %v\n---\n%s", err, src)
	}
}
