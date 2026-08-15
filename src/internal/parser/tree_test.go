package parser

import "testing"

func TestTagKindStringUnknown(t *testing.T) {
	// tagKind only ever receives valid values from matchTagHead - this
	// test covers the defensive fallback for a value out of the table,
	// which in practice should never happen.
	var k tagKind = 99
	if got := k.String(); got != "unknown tag" {
		t.Errorf("String() = %q, want %q", got, "unknown tag")
	}
}
