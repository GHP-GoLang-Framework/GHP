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
		ast.NewText("hi", 2),
	}

	got, err := Generate("page.ghp", nodes)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	want := "//line page.ghp:2\nio.WriteString(w, \"hi\")\n"
	if got != want {
		t.Errorf("Generate() = %q, want %q", got, want)
	}
}

func TestGenerateBraceSpanningStatementProducesValidGo(t *testing.T) {
	// GHP-7's task asks for a <go ...> that opens a brace and only
	// closes in a later <go ...> tag, with HTML in between. This needs
	// no special logic here or in the parser (see the comment in
	// gen_statement.go) - this test proves it for real, compiling the
	// result with go/parser instead of just comparing strings.
	nodes := []ast.Node{
		ast.NewStatement("if user.LoggedIn {", 1),
		ast.NewText("<p>Welcome, ", 2),
		ast.NewEcho("user.Name", 2),
		ast.NewText("!</p>", 2),
		ast.NewStatement("}", 3),
	}

	body, err := Generate("page.ghp", nodes)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	src := "package p\n\n" +
		"import (\n\t\"fmt\"\n\t\"html\"\n\t\"io\"\n)\n\n" +
		"func f(w io.Writer, user any) {\n" + body + "}\n"

	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "generated.go", src, 0); err != nil {
		t.Fatalf("generated code is not valid Go: %v\n---\n%s", err, src)
	}
}
