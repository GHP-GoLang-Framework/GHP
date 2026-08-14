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
				{Value: "1", Body: []ast.Node{ast.NewText("um", 2)}, Line: 2},
				{Value: "2", Body: []ast.Node{ast.NewText("dois", 3)}, Line: 3},
			},
			[]ast.Node{ast.NewText("outro", 4)},
			1),
	}

	got, err := Generate("p.ghp", nodes)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	want := "//line p.ghp:1\n" +
		"switch v {\n" +
		"//line p.ghp:2\n" +
		"case 1:\n" +
		"//line p.ghp:2\n" +
		"io.WriteString(w, \"um\")\n" +
		"//line p.ghp:3\n" +
		"case 2:\n" +
		"//line p.ghp:3\n" +
		"io.WriteString(w, \"dois\")\n" +
		"default:\n" +
		"//line p.ghp:4\n" +
		"io.WriteString(w, \"outro\")\n" +
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

	got, err := Generate("p.ghp", nodes)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	want := "//line p.ghp:1\n" +
		"switch v {\n" +
		"//line p.ghp:2\n" +
		"case 1:\n" +
		"}\n"
	if got != want {
		t.Errorf("Generate() = %q, want %q", got, want)
	}
}

func TestGenerateSwitchMultiValueCase(t *testing.T) {
	// Decisao do GHP-9: case aceita varios valores separados por
	// virgula, igual ao switch do Go - o Value ja vem pronto do parser
	// (ver parse_test.go), genSwitch so precisa colocar depois de
	// "case ".
	nodes := []ast.Node{
		ast.NewSwitch("v", []ast.Case{{Value: `"a", "b"`, Body: nil, Line: 1}}, nil, 1),
	}

	got, err := Generate("p.ghp", nodes)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	want := "//line p.ghp:1\n" +
		"switch v {\n" +
		"//line p.ghp:1\n" +
		"case \"a\", \"b\":\n" +
		"}\n"
	if got != want {
		t.Errorf("Generate() = %q, want %q", got, want)
	}
}

func TestGenerateSwitchProducesValidGo(t *testing.T) {
	// Prova de ponta a ponta, mesmo espirito do teste equivalente pro
	// go:if: o Go gerado e compilado de verdade com go/parser, nao so
	// comparado como string.
	nodes := []ast.Node{
		ast.NewSwitch("v",
			[]ast.Case{
				{Value: `"a", "b"`, Body: []ast.Node{ast.NewText("um", 2)}, Line: 2},
			},
			[]ast.Node{ast.NewEcho("v", 3)},
			1),
	}

	body, err := Generate("p.ghp", nodes)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	src := "package p\n\n" +
		"import (\n\t\"fmt\"\n\t\"html\"\n\t\"io\"\n)\n\n" +
		"func f(w io.Writer, v string) {\n" + body + "}\n"

	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "gerado.go", src, 0); err != nil {
		t.Fatalf("codigo gerado nao e Go valido: %v\n---\n%s", err, src)
	}
}
