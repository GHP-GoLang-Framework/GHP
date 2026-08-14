package codegen

import (
	goast "go/ast"
	goparser "go/parser"
	"go/token"
	"strings"
	"testing"

	"ghp/src/internal/ast"
)

// parseAssembled confirms out is valid Go on its own (independent of
// Assemble's own internal format.Source check) and returns the parsed
// file so tests can inspect its import list precisely, without being
// tripped up by gofmt reordering imports alphabetically inside a single
// block - a plain string comparison would be brittle here.
func parseAssembled(t *testing.T, out string) *goast.File {
	t.Helper()
	fset := token.NewFileSet()
	f, err := goparser.ParseFile(fset, "gerado.go", out, 0)
	if err != nil {
		t.Fatalf("saida de Assemble nao e Go valido: %v\n---\n%s", err, out)
	}
	return f
}

// assertImports checks that f imports exactly want, regardless of order
// - catches a missing import and a duplicate/extra one with the same
// helper.
func assertImports(t *testing.T, f *goast.File, want ...string) {
	t.Helper()
	var got []string
	for _, imp := range f.Imports {
		got = append(got, imp.Path.Value)
	}
	if len(got) != len(want) {
		t.Fatalf("imports = %v, want exatamente %v", got, want)
	}
	for _, w := range want {
		found := false
		for _, g := range got {
			if g == w {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("import %s ausente; imports = %v", w, got)
		}
	}
}

func TestAssembleOnlyText(t *testing.T) {
	prog := &ast.Program{Nodes: []ast.Node{ast.NewText("ola", 1)}}

	out, err := Assemble("pages", "Index", "index.ghp", prog)
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}

	f := parseAssembled(t, out)
	if f.Name.Name != "pages" {
		t.Errorf("package = %q, want %q", f.Name.Name, "pages")
	}
	// Text so precisa de io - fmt/html ficam de fora porque nao ha
	// nenhum <go=...> na pagina.
	assertImports(t, f, `"net/http"`, `"io"`)

	if !strings.Contains(out, "func Index(w http.ResponseWriter, r *http.Request) {") {
		t.Errorf("assinatura da funcao nao encontrada em:\n%s", out)
	}
}

func TestAssembleEcho(t *testing.T) {
	prog := &ast.Program{Nodes: []ast.Node{ast.NewEcho("nome", 1)}}

	out, err := Assemble("pages", "Index", "index.ghp", prog)
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}

	f := parseAssembled(t, out)
	assertImports(t, f, `"net/http"`, `"fmt"`, `"html"`, `"io"`)
}

func TestAssembleNoOutputNeedsOnlyHTTP(t *testing.T) {
	// Uma pagina sem nenhum <go=...> nem texto puro (so um statement)
	// nao precisa de fmt/html/io - so net/http, pela assinatura da
	// funcao.
	prog := &ast.Program{Nodes: []ast.Node{ast.NewStatement("_ = 1", 1)}}

	out, err := Assemble("pages", "Index", "index.ghp", prog)
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}

	f := parseAssembled(t, out)
	assertImports(t, f, `"net/http"`)
}

func TestAssembleUserImport(t *testing.T) {
	prog := &ast.Program{Nodes: []ast.Node{
		ast.NewImport([]ast.ImportPath{{Path: "strings"}}, 1),
		ast.NewText("ola", 2),
	}}

	out, err := Assemble("pages", "Index", "index.ghp", prog)
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}

	f := parseAssembled(t, out)
	assertImports(t, f, `"net/http"`, `"io"`, `"strings"`)
}

func TestAssembleUserImportWithAlias(t *testing.T) {
	prog := &ast.Program{Nodes: []ast.Node{
		ast.NewImport([]ast.ImportPath{{Alias: "s", Path: "strings"}}, 1),
		ast.NewText("ola", 2),
	}}

	out, err := Assemble("pages", "Index", "index.ghp", prog)
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}

	f := parseAssembled(t, out)
	var found bool
	for _, imp := range f.Imports {
		if imp.Path.Value == `"strings"` {
			found = true
			if imp.Name == nil || imp.Name.Name != "s" {
				t.Errorf("alias do import strings = %v, want \"s\"", imp.Name)
			}
		}
	}
	if !found {
		t.Fatal("import \"strings\" ausente")
	}
}

func TestAssembleDedupesUserImportAgainstAuto(t *testing.T) {
	// A pagina declara <go:import ("fmt")> explicitamente, mas o echo
	// tambem precisa de fmt automaticamente - so pode aparecer uma vez
	// no bloco import, ou o arquivo gerado nem compila.
	prog := &ast.Program{Nodes: []ast.Node{
		ast.NewImport([]ast.ImportPath{{Path: "fmt"}}, 1),
		ast.NewEcho("nome", 2),
	}}

	out, err := Assemble("pages", "Index", "index.ghp", prog)
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}

	f := parseAssembled(t, out)
	assertImports(t, f, `"net/http"`, `"fmt"`, `"html"`, `"io"`)
}

func TestAssembleDedupesRepeatedUserImport(t *testing.T) {
	// Duas tags <go:import> diferentes declarando o mesmo pacote - so
	// pode aparecer uma vez no bloco import.
	prog := &ast.Program{Nodes: []ast.Node{
		ast.NewImport([]ast.ImportPath{{Path: "strings"}}, 1),
		ast.NewImport([]ast.ImportPath{{Path: "strings"}}, 2),
		ast.NewText("ola", 3),
	}}

	out, err := Assemble("pages", "Index", "index.ghp", prog)
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}

	f := parseAssembled(t, out)
	assertImports(t, f, `"net/http"`, `"io"`, `"strings"`)
}

func TestAssembleInvalidFuncNameFails(t *testing.T) {
	prog := &ast.Program{Nodes: []ast.Node{ast.NewText("ola", 1)}}

	_, err := Assemble("pages", "nome invalido", "index.ghp", prog)
	if err == nil {
		t.Fatal("Assemble() = nil error, want erro para funcName invalido")
	}
}
