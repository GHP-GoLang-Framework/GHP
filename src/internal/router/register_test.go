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
	if _, err := parser.ParseFile(fset, "generated.go", out, 0); err != nil {
		t.Fatalf("Register output is not valid Go: %v\n---\n%s", err, out)
	}

	for _, want := range []string{
		`mux.HandleFunc("/", Index)`,
		`mux.HandleFunc("/blog/{slug}", BlogSlug)`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output does not contain %q:\n%s", want, out)
		}
	}
}

func TestRegisterInvalidPackageNameFails(t *testing.T) {
	_, err := Register("invalid name", nil)
	if err == nil {
		t.Fatal("Register() = nil error, want error for invalid pkg")
	}
}

func TestRegisterEmptyPages(t *testing.T) {
	out, err := Register("pages", nil)
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "generated.go", out, 0); err != nil {
		t.Fatalf("Register output is not valid Go: %v\n---\n%s", err, out)
	}
}
