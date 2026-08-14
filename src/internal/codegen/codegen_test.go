package codegen

import (
	"go/parser"
	"go/token"
	"testing"

	"ghp/src/internal/ast"
)

func TestGenerateSkipsImportNodes(t *testing.T) {
	nodes := []ast.Node{
		ast.NewImport([]ast.ImportPath{{Path: "fmt"}}, 1),
		ast.NewText("ola", 2),
	}

	got, err := Generate("pagina.ghp", nodes)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	want := "//line pagina.ghp:2\nio.WriteString(w, \"ola\")\n"
	if got != want {
		t.Errorf("Generate() = %q, want %q", got, want)
	}
}

func TestGenerateUnsupportedNodeType(t *testing.T) {
	// go:if ainda nao tem gerador proprio (isso e tarefa do GHP-8) -
	// Generate tem que falhar alto e claro em vez de omitir em silencio
	// o bloco condicional do HTML final.
	nodes := []ast.Node{ast.NewIf("true", nil, nil, 7)}

	_, err := Generate("pagina.ghp", nodes)
	if err == nil {
		t.Fatal("Generate() = nil error, want erro para *ast.If (ainda nao suportado)")
	}
}

func TestGenerateBraceSpanningStatementProducesValidGo(t *testing.T) {
	// A tarefa do GHP-7 pede um <go ...> que abre uma chave e so fecha
	// numa tag <go ...> posterior, com HTML no meio. Isso nao exige
	// nenhuma logica especial aqui nem no parser (ver o comentario em
	// gen_statement.go) - este teste prova isso de verdade, compilando
	// o resultado com go/parser em vez de so comparar strings.
	nodes := []ast.Node{
		ast.NewStatement("if usuario.Logado {", 1),
		ast.NewText("<p>Bem vindo, ", 2),
		ast.NewEcho("usuario.Nome", 2),
		ast.NewText("!</p>", 2),
		ast.NewStatement("}", 3),
	}

	body, err := Generate("pagina.ghp", nodes)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	src := "package p\n\n" +
		"import (\n\t\"fmt\"\n\t\"html\"\n\t\"io\"\n)\n\n" +
		"func f(w io.Writer, usuario any) {\n" + body + "}\n"

	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "gerado.go", src, 0); err != nil {
		t.Fatalf("codigo gerado nao e Go valido: %v\n---\n%s", err, src)
	}
}
