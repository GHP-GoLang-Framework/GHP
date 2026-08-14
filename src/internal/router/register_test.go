package router

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestRegisterProducesValidGo(t *testing.T) {
	pages := []Page{
		{GhpPath: "index.ghp", Route: "/", FuncName: "Index"},
		{GhpPath: "blog/[slug].ghp", Route: "/blog/{slug}", FuncName: "BlogSlug"},
	}

	out, err := Register("pages", pages)
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "gerado.go", out, 0); err != nil {
		t.Fatalf("saida de Register nao e Go valido: %v\n---\n%s", err, out)
	}

	for _, want := range []string{
		`mux.HandleFunc("/", Index)`,
		`mux.HandleFunc("/blog/{slug}", BlogSlug)`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saida nao contem %q:\n%s", want, out)
		}
	}
}

func TestRegisterInvalidPackageNameFails(t *testing.T) {
	_, err := Register("nome invalido", nil)
	if err == nil {
		t.Fatal("Register() = nil error, want erro para pkg invalido")
	}
}

func TestRegisterEmptyPages(t *testing.T) {
	out, err := Register("pages", nil)
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "gerado.go", out, 0); err != nil {
		t.Fatalf("saida de Register nao e Go valido: %v\n---\n%s", err, out)
	}
}
