package codegen

import (
	"go/parser"
	"go/token"
	"testing"

	"ghp/src/internal/ast"
)

func TestGenerateIfWithoutElse(t *testing.T) {
	nodes := []ast.Node{
		ast.NewIf("a == b", []ast.Node{ast.NewText("sim", 2)}, nil, 1),
	}

	got, err := Generate("p.ghp", nodes)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	want := "//line p.ghp:1\n" +
		"if a == b {\n" +
		"//line p.ghp:2\n" +
		"io.WriteString(w, \"sim\")\n" +
		"}\n"
	if got != want {
		t.Errorf("Generate() = %q, want %q", got, want)
	}
}

func TestGenerateIfWithElse(t *testing.T) {
	nodes := []ast.Node{
		ast.NewIf("a == b",
			[]ast.Node{ast.NewText("sim", 2)},
			[]ast.Node{ast.NewText("nao", 3)},
			1),
	}

	got, err := Generate("p.ghp", nodes)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	want := "//line p.ghp:1\n" +
		"if a == b {\n" +
		"//line p.ghp:2\n" +
		"io.WriteString(w, \"sim\")\n" +
		"} else {\n" +
		"//line p.ghp:3\n" +
		"io.WriteString(w, \"nao\")\n" +
		"}\n"
	if got != want {
		t.Errorf("Generate() = %q, want %q", got, want)
	}
}

func TestGenerateIfPropagatesErrorFromBody(t *testing.T) {
	// go:for ainda nao tem gerador (GHP-10) - um go:if cujo corpo
	// contem um go:for tem que propagar esse erro, nao engolir ou gerar
	// Go quebrado.
	nodes := []ast.Node{
		ast.NewIf("a", []ast.Node{ast.NewFor("i := range xs", nil, 2)}, nil, 1),
	}

	_, err := Generate("p.ghp", nodes)
	if err == nil {
		t.Fatal("Generate() = nil error, want erro propagado de dentro do corpo do if")
	}
}

func TestGenerateIfPropagatesErrorFromElseBody(t *testing.T) {
	// Mesma checagem do teste anterior, mas com o erro vindo do corpo
	// do else em vez do then - sao dois returns distintos em genIf.
	nodes := []ast.Node{
		ast.NewIf("a", []ast.Node{ast.NewText("sim", 2)}, []ast.Node{ast.NewFor("i := range xs", nil, 3)}, 1),
	}

	_, err := Generate("p.ghp", nodes)
	if err == nil {
		t.Fatal("Generate() = nil error, want erro propagado de dentro do corpo do else")
	}
}

func TestGenerateNestedIfProducesValidGo(t *testing.T) {
	// Prova de ponta a ponta: go:if dentro de go:if (mesmo recurso que
	// o parser ja testa em TestParseNesting) tem que virar Go
	// sintaticamente valido de verdade, nao so bater como string.
	inner := ast.NewIf("b", []ast.Node{ast.NewText("x", 2)}, nil, 2)
	outer := ast.NewIf("a", []ast.Node{inner}, []ast.Node{ast.NewText("y", 4)}, 1)

	body, err := Generate("p.ghp", []ast.Node{outer})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	src := "package p\n\nimport \"io\"\n\nfunc f(w io.Writer, a, b bool) {\n" + body + "}\n"

	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "gerado.go", src, 0); err != nil {
		t.Fatalf("codigo gerado nao e Go valido: %v\n---\n%s", err, src)
	}
}
