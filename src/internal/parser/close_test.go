package parser

import "testing"

func TestFindTagClose(t *testing.T) {
	// Aspas com escape: o '"' escapado (\") nao pode ser lido como o fim
	// da string, senao o '>' logo depois dele fecharia a tag cedo demais.
	src := ` x == "a\"b" >resto`
	want := len(` x == "a\"b" `) // ate o '>' logo apos a aspa de fechamento real

	got := findTagClose(src, 0)
	if got != want {
		t.Errorf("findTagClose(%q, 0) = %d, want %d", src, got, want)
	}
}
