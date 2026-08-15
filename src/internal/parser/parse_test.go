package parser

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"ghp/src/internal/ast"
)

func TestParseNodeTypes(t *testing.T) {
	t.Run("text", func(t *testing.T) {
		prog, err := Parse("hello world")
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		if len(prog.Nodes) != 1 {
			t.Fatalf("got %d nodes, want 1", len(prog.Nodes))
		}
		text, ok := prog.Nodes[0].(*ast.Text)
		if !ok {
			t.Fatalf("node type = %T, want *ast.Text", prog.Nodes[0])
		}
		if text.Value != "hello world" {
			t.Errorf("Value = %q, want %q", text.Value, "hello world")
		}
		if text.Line() != 1 {
			t.Errorf("Line() = %d, want 1", text.Line())
		}
	})

	t.Run("import", func(t *testing.T) {
		prog, err := Parse(`<go:import ("fmt")/>`)
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		imp, ok := prog.Nodes[0].(*ast.Import)
		if !ok {
			t.Fatalf("node type = %T, want *ast.Import", prog.Nodes[0])
		}
		want := []ast.ImportPath{{Path: "fmt"}}
		if !reflect.DeepEqual(imp.Paths, want) {
			t.Errorf("Paths = %#v, want %#v", imp.Paths, want)
		}
	})

	t.Run("import with alias and multiple paths", func(t *testing.T) {
		prog, err := Parse("<go:import (\n\tf \"fmt\"\n\t\"strings\"\n)/>")
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		imp := prog.Nodes[0].(*ast.Import)
		want := []ast.ImportPath{
			{Alias: "f", Path: "fmt"},
			{Path: "strings"},
		}
		if !reflect.DeepEqual(imp.Paths, want) {
			t.Errorf("Paths = %#v, want %#v", imp.Paths, want)
		}
	})

	t.Run("import of module-internal package", func(t *testing.T) {
		prog, err := Parse(`<go:import ("myapp/internal/db")/>`)
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		imp := prog.Nodes[0].(*ast.Import)
		want := []ast.ImportPath{{Path: "myapp/internal/db"}}
		if !reflect.DeepEqual(imp.Paths, want) {
			t.Errorf("Paths = %#v, want %#v", imp.Paths, want)
		}
	})

	t.Run("import skips empty item between commas", func(t *testing.T) {
		prog, err := Parse(`<go:import ("fmt", , "strings")/>`)
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		imp := prog.Nodes[0].(*ast.Import)
		want := []ast.ImportPath{{Path: "fmt"}, {Path: "strings"}}
		if !reflect.DeepEqual(imp.Paths, want) {
			t.Errorf("Paths = %#v, want %#v", imp.Paths, want)
		}
	})

	t.Run("statement", func(t *testing.T) {
		prog, err := Parse("<go\n\tx := 1\n/>")
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		st, ok := prog.Nodes[0].(*ast.Statement)
		if !ok {
			t.Fatalf("node type = %T, want *ast.Statement", prog.Nodes[0])
		}
		if st.Code != "x := 1" {
			t.Errorf("Code = %q, want %q", st.Code, "x := 1")
		}
	})

	t.Run("echo", func(t *testing.T) {
		prog, err := Parse(`<go= name />`)
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		echo, ok := prog.Nodes[0].(*ast.Echo)
		if !ok {
			t.Fatalf("node type = %T, want *ast.Echo", prog.Nodes[0])
		}
		if echo.Expr != "name" {
			t.Errorf("Expr = %q, want %q", echo.Expr, "name")
		}
	})

	t.Run("if without else", func(t *testing.T) {
		prog, err := Parse(`<go:if a == b/>yes<go:endif/>`)
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		ifNode, ok := prog.Nodes[0].(*ast.If)
		if !ok {
			t.Fatalf("node type = %T, want *ast.If", prog.Nodes[0])
		}
		if ifNode.Cond != "a == b" {
			t.Errorf("Cond = %q, want %q", ifNode.Cond, "a == b")
		}
		if len(ifNode.Then) != 1 {
			t.Fatalf("len(Then) = %d, want 1", len(ifNode.Then))
		}
		if ifNode.Else != nil {
			t.Errorf("Else = %#v, want nil", ifNode.Else)
		}
	})

	t.Run("if with else", func(t *testing.T) {
		prog, err := Parse(`<go:if a == b/>yes<go:else/>no<go:endif/>`)
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		ifNode := prog.Nodes[0].(*ast.If)
		if len(ifNode.Then) != 1 || len(ifNode.Else) != 1 {
			t.Fatalf("Then=%d Else=%d, want 1 and 1", len(ifNode.Then), len(ifNode.Else))
		}
	})

	t.Run("if with compound condition", func(t *testing.T) {
		prog, err := Parse(`<go:if a && b || c/>x<go:endif/>`)
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		ifNode := prog.Nodes[0].(*ast.If)
		if ifNode.Cond != "a && b || c" {
			t.Errorf("Cond = %q, want %q", ifNode.Cond, "a && b || c")
		}
	})

	t.Run("switch", func(t *testing.T) {
		src := "<go:switch v/>" +
			"<go:case 1/>one" +
			"<go:case 2/>two" +
			"<go:default/>other" +
			"<go:endswitch/>"
		prog, err := Parse(src)
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		sw, ok := prog.Nodes[0].(*ast.Switch)
		if !ok {
			t.Fatalf("node type = %T, want *ast.Switch", prog.Nodes[0])
		}
		if sw.Expr != "v" {
			t.Errorf("Expr = %q, want %q", sw.Expr, "v")
		}
		if len(sw.Cases) != 2 {
			t.Fatalf("len(Cases) = %d, want 2", len(sw.Cases))
		}
		if sw.Cases[0].Value != "1" || sw.Cases[1].Value != "2" {
			t.Errorf("Cases values = %q, %q, want 1, 2", sw.Cases[0].Value, sw.Cases[1].Value)
		}
		if sw.Default == nil {
			t.Error("Default = nil, want populated")
		}
	})

	t.Run("switch with multiple values in the same case", func(t *testing.T) {
		// GHP-9 decision: a go:case accepts several comma-separated
		// values, just like Go's switch (case "a", "b":) - no special
		// handling needed here because Case.Value is already opaque
		// text, passed verbatim to codegen.
		prog, err := Parse(`<go:switch v/><go:case "a", "b"/>x<go:endswitch/>`)
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		sw := prog.Nodes[0].(*ast.Switch)
		if len(sw.Cases) != 1 {
			t.Fatalf("len(Cases) = %d, want 1", len(sw.Cases))
		}
		if sw.Cases[0].Value != `"a", "b"` {
			t.Errorf("Value = %q, want %q", sw.Cases[0].Value, `"a", "b"`)
		}
	})

	t.Run("statement with brace opened and closed in separate tags", func(t *testing.T) {
		// <go .../> does not pair braces between tags - each tag
		// becomes an independent Statement, and it is go build itself
		// that matches the "{" of this tag with the "}" of a later
		// tag, since both end up as literal, sequential Go code in
		// the generated file (see internal/transpiler/codegen). The parser does
		// not need to know about this.
		prog, err := Parse(`<go if user.LoggedIn {/>hi<go }/>`)
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		if len(prog.Nodes) != 3 {
			t.Fatalf("got %d nodes, want 3 (statement, text, statement)", len(prog.Nodes))
		}
		open, ok := prog.Nodes[0].(*ast.Statement)
		if !ok {
			t.Fatalf("node[0] type = %T, want *ast.Statement", prog.Nodes[0])
		}
		if open.Code != "if user.LoggedIn {" {
			t.Errorf("Code = %q, want %q", open.Code, "if user.LoggedIn {")
		}
		close, ok := prog.Nodes[2].(*ast.Statement)
		if !ok {
			t.Fatalf("node[2] type = %T, want *ast.Statement", prog.Nodes[2])
		}
		if close.Code != "}" {
			t.Errorf("Code = %q, want %q", close.Code, "}")
		}
	})

	t.Run("for", func(t *testing.T) {
		prog, err := Parse(`<go:for i := range items/>hi<go:endfor/>`)
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		forNode, ok := prog.Nodes[0].(*ast.For)
		if !ok {
			t.Fatalf("node type = %T, want *ast.For", prog.Nodes[0])
		}
		if forNode.Expr != "i := range items" {
			t.Errorf("Expr = %q, want %q", forNode.Expr, "i := range items")
		}
		if len(forNode.Body) != 1 {
			t.Fatalf("len(Body) = %d, want 1", len(forNode.Body))
		}
	})
}

func TestParseNesting(t *testing.T) {
	t.Run("if inside for", func(t *testing.T) {
		src := `<go:for i := range items/><go:if i != 0/>, <go:endif/><go= i /><go:endfor/>`
		prog, err := Parse(src)
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}

		forNode, ok := prog.Nodes[0].(*ast.For)
		if !ok {
			t.Fatalf("node type = %T, want *ast.For", prog.Nodes[0])
		}
		if len(forNode.Body) != 2 {
			t.Fatalf("len(Body) = %d, want 2 (if + echo)", len(forNode.Body))
		}

		ifNode, ok := forNode.Body[0].(*ast.If)
		if !ok {
			t.Fatalf("Body[0] type = %T, want *ast.If", forNode.Body[0])
		}
		if ifNode.Cond != "i != 0" {
			t.Errorf("Cond = %q, want %q", ifNode.Cond, "i != 0")
		}

		if _, ok := forNode.Body[1].(*ast.Echo); !ok {
			t.Fatalf("Body[1] type = %T, want *ast.Echo", forNode.Body[1])
		}
	})

	t.Run("for inside if", func(t *testing.T) {
		src := `<go:if show/><go:for i := range items/><go= i /><go:endfor/><go:endif/>`
		prog, err := Parse(src)
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}

		ifNode, ok := prog.Nodes[0].(*ast.If)
		if !ok {
			t.Fatalf("node type = %T, want *ast.If", prog.Nodes[0])
		}
		if len(ifNode.Then) != 1 {
			t.Fatalf("len(Then) = %d, want 1", len(ifNode.Then))
		}
		if _, ok := ifNode.Then[0].(*ast.For); !ok {
			t.Fatalf("Then[0] type = %T, want *ast.For", ifNode.Then[0])
		}
	})

	t.Run("switch inside for", func(t *testing.T) {
		src := `<go:for i := range items/><go:switch i/><go:case 0/>zero<go:endswitch/><go:endfor/>`
		prog, err := Parse(src)
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}

		forNode, ok := prog.Nodes[0].(*ast.For)
		if !ok {
			t.Fatalf("node type = %T, want *ast.For", prog.Nodes[0])
		}
		if len(forNode.Body) != 1 {
			t.Fatalf("len(Body) = %d, want 1", len(forNode.Body))
		}
		if _, ok := forNode.Body[0].(*ast.Switch); !ok {
			t.Fatalf("Body[0] type = %T, want *ast.Switch", forNode.Body[0])
		}
	})

	t.Run("for inside for", func(t *testing.T) {
		// GHP-10 acceptance criterion: "nested go:for (loop inside
		// loop)".
		src := `<go:for _, row := range matrix/><go:for _, cell := range row/><go= cell /><go:endfor/><go:endfor/>`
		prog, err := Parse(src)
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}

		outer, ok := prog.Nodes[0].(*ast.For)
		if !ok {
			t.Fatalf("node type = %T, want *ast.For", prog.Nodes[0])
		}
		if len(outer.Body) != 1 {
			t.Fatalf("len(outer.Body) = %d, want 1", len(outer.Body))
		}

		inner, ok := outer.Body[0].(*ast.For)
		if !ok {
			t.Fatalf("outer.Body[0] type = %T, want *ast.For", outer.Body[0])
		}
		if inner.Expr != "_, cell := range row" {
			t.Errorf("Expr = %q, want %q", inner.Expr, "_, cell := range row")
		}
	})
}

func TestParseTemplateGhp(t *testing.T) {
	data, err := os.ReadFile("../../../docs/template.ghp")
	if err != nil {
		t.Fatalf("could not read docs/template.ghp: %v", err)
	}

	prog, err := Parse(string(data))
	if err != nil {
		t.Fatalf("Parse(docs/template.ghp) failed: %v", err)
	}

	wantKinds := []string{
		"*ast.Import", "*ast.Text", "*ast.Statement", "*ast.Text",
		"*ast.Echo", "*ast.Text", "*ast.If", "*ast.Text",
		"*ast.Switch", "*ast.Text", "*ast.For", "*ast.Text",
	}
	if len(prog.Nodes) != len(wantKinds) {
		t.Fatalf("got %d top-level nodes, want %d", len(prog.Nodes), len(wantKinds))
	}
	for i, n := range prog.Nodes {
		got := fmt.Sprintf("%T", n)
		if got != wantKinds[i] {
			t.Errorf("node[%d] = %s, want %s", i, got, wantKinds[i])
		}
	}
}
