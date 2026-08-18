package gen

import (
	"strings"
	"testing"

	"ghp/src/internal/ast"
)

func TestSwitchWithDefault(t *testing.T) {
	var b strings.Builder
	err := Switch(&b, ast.NewSwitch("v",
		[]ast.Case{
			{Value: "1", Body: []ast.Node{ast.NewText("one", 2)}, Line: 2},
			{Value: "2", Body: []ast.Node{ast.NewText("two", 3)}, Line: 3},
		},
		[]ast.Node{ast.NewText("other", 4)},
		1), noopNodes)
	if err != nil {
		t.Fatalf("Switch(): %v", err)
	}

	want := "switch v {\n" +
		"case 1:\n" +
		"io.WriteString(w, \"one\")\n" +
		"case 2:\n" +
		"io.WriteString(w, \"two\")\n" +
		"default:\n" +
		"io.WriteString(w, \"other\")\n" +
		"}\n"
	if got := b.String(); got != want {
		t.Errorf("Switch() = %q, want %q", got, want)
	}
}

func TestSwitchWithoutDefault(t *testing.T) {
	var b strings.Builder
	err := Switch(&b, ast.NewSwitch("v",
		[]ast.Case{{Value: "1", Body: nil, Line: 2}},
		nil,
		1), noopNodes)
	if err != nil {
		t.Fatalf("Switch(): %v", err)
	}

	want := "switch v {\n" +
		"case 1:\n" +
		"}\n"
	if got := b.String(); got != want {
		t.Errorf("Switch() = %q, want %q", got, want)
	}
}

func TestCaseBodyError(t *testing.T) {
	var b strings.Builder
	err := Switch(&b, ast.NewSwitch("v",
		[]ast.Case{{Value: "1", Body: []ast.Node{nil}, Line: 1}},
		nil, 1), errNodes)
	if err == nil {
		t.Fatal("Switch() = nil error, want error from nodesFn")
	}
}

func TestDefaultBodyError(t *testing.T) {
	var b strings.Builder
	err := Switch(&b, ast.NewSwitch("v", nil, []ast.Node{nil}, 1), errNodes)
	if err == nil {
		t.Fatal("Switch() = nil error, want error from nodesFn")
	}
}
