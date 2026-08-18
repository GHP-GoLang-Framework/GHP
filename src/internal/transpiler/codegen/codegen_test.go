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

	got, err := Generate(nodes)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	want := "io.WriteString(w, \"hi\")\n"
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

	body, err := Generate(nodes)
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

// TestGenerateErrorsOnUnsupportedNode exercises the defensive error
// branches in Generate and the gen_* emitters. A nil node can't come
// from the parser - every ast.Node it produces is handled by
// generateNode's switch (and ast.Node can only be implemented inside
// package ast), so these paths only fire if a future node type is
// added to ast without a matching case here. Each one must fail loudly
// instead of being silently dropped.
//
// Ex: a nil node nested in a <go:if> body -> Generate reports an error.
func TestGenerateErrorsOnUnsupportedNode(t *testing.T) {
	tests := []struct {
		name  string
		nodes []ast.Node
	}{
		{"top level", []ast.Node{nil}},
		{"if then body", []ast.Node{ast.NewIf("true", []ast.Node{nil}, nil, nil, 1)}},
		{"if else body", []ast.Node{ast.NewIf("true", nil, nil, []ast.Node{nil}, 1)}},
		{"if elif body", []ast.Node{ast.NewIf("true", nil, []ast.ElseIf{{Cond: "false", Body: []ast.Node{nil}, Line: 1}}, nil, 1)}},
		{"for body", []ast.Node{ast.NewFor("i := 0; i < 1; i++", []ast.Node{nil}, 1)}},
		{"switch case body", []ast.Node{ast.NewSwitch("v", []ast.Case{{Value: "1", Body: []ast.Node{nil}, Line: 1}}, nil, 1)}},
		{"switch default body", []ast.Node{ast.NewSwitch("v", nil, []ast.Node{nil}, 1)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Generate(tt.nodes); err == nil {
				t.Errorf("Generate(%s) = nil error, want unsupported node type error", tt.name)
			}
		})
	}
}
