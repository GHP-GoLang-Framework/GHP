package codegen

import (
	"strconv"
	"strings"
	"testing"

	"ghp/src/internal/ast"
)

func TestGenerateText(t *testing.T) {
	nodes := []ast.Node{ast.NewText("ola mundo", 3)}

	got, err := Generate("pagina.ghp", nodes)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	want := "//line pagina.ghp:3\nio.WriteString(w, \"ola mundo\")\n"
	if got != want {
		t.Errorf("Generate() = %q, want %q", got, want)
	}
}

func TestGenerateTextEscapesSpecialChars(t *testing.T) {
	// Aspas, quebra de linha e barra invertida no HTML de origem nao
	// podem vazar pro literal Go sem escapar, ou o .go gerado nem
	// compila. Em vez de escrever a mao a string esperada (fragil e
	// dificil de ler), extraimos de volta o literal gerado com
	// strconv.Unquote e conferimos que bate com o valor original.
	value := "linha 1\n\"aspas\" e \\barra"
	nodes := []ast.Node{ast.NewText(value, 1)}

	got, err := Generate("p.ghp", nodes)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	start := strings.Index(got, "io.WriteString(w, ") + len("io.WriteString(w, ")
	end := strings.LastIndex(got, ")")
	literal := got[start:end]

	decoded, err := strconv.Unquote(literal)
	if err != nil {
		t.Fatalf("literal gerado nao e um Go string literal valido: %v (%q)", err, literal)
	}
	if decoded != value {
		t.Errorf("decoded = %q, want %q", decoded, value)
	}
}
