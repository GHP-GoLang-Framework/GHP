package parser

import "testing"

func TestTagKindStringUnknown(t *testing.T) {
	// tagKind so recebe valores validos vindos de matchTagHead - este
	// teste cobre o fallback defensivo para um valor fora da tabela, que
	// na pratica nunca deveria acontecer.
	var k tagKind = 99
	if got := k.String(); got != "unknown tag" {
		t.Errorf("String() = %q, want %q", got, "unknown tag")
	}
}
