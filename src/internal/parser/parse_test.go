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
		prog, err := Parse("ola mundo")
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
		if text.Value != "ola mundo" {
			t.Errorf("Value = %q, want %q", text.Value, "ola mundo")
		}
		if text.Line() != 1 {
			t.Errorf("Line() = %d, want 1", text.Line())
		}
	})

	t.Run("import", func(t *testing.T) {
		prog, err := Parse(`<go:import ("fmt")>`)
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

	t.Run("import com alias e multiplos paths", func(t *testing.T) {
		prog, err := Parse("<go:import (\n\tf \"fmt\"\n\t\"strings\"\n)>")
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

	t.Run("import de pacote interno do modulo", func(t *testing.T) {
		prog, err := Parse(`<go:import ("meuapp/internal/db")>`)
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		imp := prog.Nodes[0].(*ast.Import)
		want := []ast.ImportPath{{Path: "meuapp/internal/db"}}
		if !reflect.DeepEqual(imp.Paths, want) {
			t.Errorf("Paths = %#v, want %#v", imp.Paths, want)
		}
	})

	t.Run("import ignora item vazio entre virgulas", func(t *testing.T) {
		prog, err := Parse(`<go:import ("fmt", , "strings")>`)
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
		prog, err := Parse("<go\n\tx := 1\n>")
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
		prog, err := Parse(`<go= nome >`)
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		echo, ok := prog.Nodes[0].(*ast.Echo)
		if !ok {
			t.Fatalf("node type = %T, want *ast.Echo", prog.Nodes[0])
		}
		if echo.Expr != "nome" {
			t.Errorf("Expr = %q, want %q", echo.Expr, "nome")
		}
	})

	t.Run("if sem else", func(t *testing.T) {
		prog, err := Parse(`<go:if a == b>sim</go:if>`)
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

	t.Run("if com else", func(t *testing.T) {
		prog, err := Parse(`<go:if a == b>sim<go:else>nao</go:if>`)
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		ifNode := prog.Nodes[0].(*ast.If)
		if len(ifNode.Then) != 1 || len(ifNode.Else) != 1 {
			t.Fatalf("Then=%d Else=%d, want 1 e 1", len(ifNode.Then), len(ifNode.Else))
		}
	})

	t.Run("switch", func(t *testing.T) {
		src := "<go:switch v>" +
			"<go:case 1>um" +
			"<go:case 2>dois" +
			"<go:default>outro" +
			"</go:switch>"
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

	t.Run("switch com multiplos valores no mesmo case", func(t *testing.T) {
		// Decisao do GHP-9: um go:case aceita varios valores separados
		// por virgula, igual ao switch do Go (case "a", "b":) - nao
		// precisa de nenhum tratamento especial aqui porque Case.Value
		// ja e texto opaco, repassado verbatim pro codegen.
		prog, err := Parse(`<go:switch v><go:case "a", "b">x</go:switch>`)
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

	t.Run("statement com chave aberta e fechada em tags separadas", func(t *testing.T) {
		// <go ...> nao pareia chaves entre tags - cada tag vira um
		// Statement independente, e e o proprio go build que vai casar
		// o "{" desta tag com o "}" da tag mais adiante, ja que os dois
		// terminam como codigo Go literal e sequencial no arquivo
		// gerado (ver internal/codegen). O parser nao precisa saber
		// disso.
		prog, err := Parse(`<go if usuario.Logado {>ola<go }>`)
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		if len(prog.Nodes) != 3 {
			t.Fatalf("got %d nodes, want 3 (statement, texto, statement)", len(prog.Nodes))
		}
		open, ok := prog.Nodes[0].(*ast.Statement)
		if !ok {
			t.Fatalf("node[0] type = %T, want *ast.Statement", prog.Nodes[0])
		}
		if open.Code != "if usuario.Logado {" {
			t.Errorf("Code = %q, want %q", open.Code, "if usuario.Logado {")
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
		prog, err := Parse(`<go:for i := range items>oi</go:for>`)
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
	t.Run("if dentro de for", func(t *testing.T) {
		src := `<go:for i := range items><go:if i != 0>, </go:if><go= i ></go:for>`
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

	t.Run("for dentro de if", func(t *testing.T) {
		src := `<go:if show><go:for i := range items><go= i ></go:for></go:if>`
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

	t.Run("switch dentro de for", func(t *testing.T) {
		src := `<go:for i := range items><go:switch i><go:case 0>zero</go:switch></go:for>`
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
}

func TestParseTemplateGhp(t *testing.T) {
	data, err := os.ReadFile("../../../docs/template.ghp")
	if err != nil {
		t.Fatalf("nao consegui ler docs/template.ghp: %v", err)
	}

	prog, err := Parse(string(data))
	if err != nil {
		t.Fatalf("Parse(docs/template.ghp) falhou: %v", err)
	}

	wantKinds := []string{
		"*ast.Import", "*ast.Text", "*ast.Statement", "*ast.Text",
		"*ast.Echo", "*ast.Text", "*ast.If", "*ast.Text",
		"*ast.Switch", "*ast.Text", "*ast.For", "*ast.Text",
	}
	if len(prog.Nodes) != len(wantKinds) {
		t.Fatalf("got %d nos de topo, want %d", len(prog.Nodes), len(wantKinds))
	}
	for i, n := range prog.Nodes {
		got := fmt.Sprintf("%T", n)
		if got != wantKinds[i] {
			t.Errorf("node[%d] = %s, want %s", i, got, wantKinds[i])
		}
	}
}
