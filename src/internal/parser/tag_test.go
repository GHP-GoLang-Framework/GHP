package parser

import "testing"

func TestMatchTagHead(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantKind    tagKind
		wantHeadLen int
	}{
		{"import", `go:import ("fmt")>`, tagImport, len("go:import")},
		{"echo", `go= expressao >`, tagEcho, len("go=")},
		{"statement", `go codigo >`, tagStatement, len("go")},
		{"if", `go:if a == b>`, tagIf, len("go:if")},
		{"else", `go:else>`, tagElse, len("go:else")},
		{"switch", `go:switch v>`, tagSwitch, len("go:switch")},
		{"case", `go:case x>`, tagCase, len("go:case")},
		{"default", `go:default>`, tagDefault, len("go:default")},
		{"for", `go:for expr>`, tagFor, len("go:for")},
		{"close if", `/go:if>`, tagCloseIf, 1 + len("go:if")},
		{"close switch", `/go:switch>`, tagCloseSwitch, 1 + len("go:switch")},
		{"close for", `/go:for>`, tagCloseFor, 1 + len("go:for")},

		// Cabecalho que termina exatamente no fim da string, sem nada
		// depois (nem o '>') - cobre o ramo `rest == ""` de boundary e
		// statementBoundary, que o '>' presente em todo caso acima nunca
		// exercita.
		{"else no fim absoluto da string", `go:else`, tagElse, len("go:else")},
		{"go no fim absoluto da string", `go`, tagStatement, len("go")},

		// Casos que NÃO são tags GHP — são o motivo de existir a checagem
		// de fronteira, não só o "acontece de funcionar".
		{"html element starting with go", `google-maps>`, tagNone, 0},
		{"identifier longer than go:if", `go:iffy>`, tagNone, 0},
		{"identifier longer than go:for", `go:format>`, tagNone, 0},
		{"bare go followed by letter", `gopher>`, tagNone, 0},
		{"dangling colon", `go:`, tagNone, 0},
		{"empty string", ``, tagNone, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotKind, gotHeadLen := matchTagHead(tt.input)
			if gotKind != tt.wantKind {
				t.Errorf("matchTagHead(%q) kind = %v, want %v", tt.input, gotKind, tt.wantKind)
			}
			if gotHeadLen != tt.wantHeadLen {
				t.Errorf("matchTagHead(%q) headLen = %d, want %d", tt.input, gotHeadLen, tt.wantHeadLen)
			}
		})
	}
}
