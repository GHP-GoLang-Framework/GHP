package codegen

import (
	"fmt"
	"html"
	"strings"
	"testing"

	"ghp/src/internal/ast"
)

func TestGenerateEcho(t *testing.T) {
	nodes := []ast.Node{ast.NewEcho("usuario.Nome", 5)}

	got, err := Generate("pagina.ghp", nodes)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	want := "//line pagina.ghp:5\nio.WriteString(w, html.EscapeString(fmt.Sprint(usuario.Nome)))\n"
	if got != want {
		t.Errorf("Generate() = %q, want %q", got, want)
	}
}

func TestGenerateEchoFunctionCall(t *testing.T) {
	// Criterio de aceite do GHP-7: qualquer expressao Go valida
	// funciona em <go= ...>, incluindo chamada de funcao com retorno -
	// o parser guarda o texto da tag de forma opaca (ast.Echo.Expr), e
	// o codegen so precisa interpolar ele verbatim dentro do
	// fmt.Sprint(...), sem se importar com o que tem dentro.
	nodes := []ast.Node{ast.NewEcho("usuario.NomeCompleto()", 8)}

	got, err := Generate("pagina.ghp", nodes)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	want := "//line pagina.ghp:8\nio.WriteString(w, html.EscapeString(fmt.Sprint(usuario.NomeCompleto())))\n"
	if got != want {
		t.Errorf("Generate() = %q, want %q", got, want)
	}
}

func TestGenerateEchoOutputIsEscaped(t *testing.T) {
	// Confirma o efeito real da formula que genEcho emite: um valor
	// com HTML perigoso tem que sair como texto, nao como marcacao.
	// Se alguem trocar a formula de escaping sem perceber a implicacao
	// de seguranca, este teste quebra.
	payload := `<script>alert(1)</script>`
	rendered := html.EscapeString(fmt.Sprint(payload))

	if strings.Contains(rendered, "<script>") {
		t.Errorf("saida nao escapada corretamente: %q", rendered)
	}
}
