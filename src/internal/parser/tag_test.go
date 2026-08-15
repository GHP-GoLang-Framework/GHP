package parser

import "testing"

func TestMatchTagHead(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantKind    tagKind
		wantHeadLen int
	}{
		{"import", `go:import ("fmt")/>`, tagImport, len("go:import")},
		{"echo", `go= expression />`, tagEcho, len("go=")},
		{"statement", `go code />`, tagStatement, len("go")},
		{"if", `go:if a == b/>`, tagIf, len("go:if")},
		{"else", `go:else/>`, tagElse, len("go:else")},
		{"switch", `go:switch v/>`, tagSwitch, len("go:switch")},
		{"case", `go:case x/>`, tagCase, len("go:case")},
		{"default", `go:default/>`, tagDefault, len("go:default")},
		{"for", `go:for expr/>`, tagFor, len("go:for")},
		{"close if", `go:endif/>`, tagCloseIf, len("go:endif")},
		{"close switch", `go:endswitch/>`, tagCloseSwitch, len("go:endswitch")},
		{"close for", `go:endfor/>`, tagCloseFor, len("go:endfor")},

		// Head that ends exactly at the end of the string, with nothing
		// after (not even the '/>') - covers the `rest == ""` branch of
		// boundary and statementBoundary, which the '/>' present in every
		// case above never exercises.
		{"else at the absolute end of the string", `go:else`, tagElse, len("go:else")},
		{"go at the absolute end of the string", `go`, tagStatement, len("go")},

		// Cases that are NOT GHP tags - they are the reason the
		// boundary check exists, not just "it happens to work".
		{"html element starting with go", `google-maps>`, tagNone, 0},
		{"identifier longer than go:if", `go:iffy/>`, tagNone, 0},
		{"identifier longer than go:for", `go:format/>`, tagNone, 0},
		{"identifier longer than go:endif", `go:endifx/>`, tagNone, 0},
		{"bare go followed by letter", `gopher>`, tagNone, 0},
		{"dangling colon", `go:`, tagNone, 0},
		{"old close tag is not a tag", `/go:if>`, tagNone, 0},
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
