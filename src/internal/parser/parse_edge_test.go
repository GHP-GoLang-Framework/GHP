package parser

import "testing"

func TestParseEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		wantErr string
	}{
		{"unclosed tag", `<go:if a == b>texto`, "line 1: <go:if> missing </go:if>"},
		{"tag never closed with >", `<go:if a == b`, "line 1: tag not closed with '>'"},
		{"stray close", `texto </go:if> mais texto`, "line 1: unexpected </go:if> here"},
		{"else without if", `<go:else>oi</go:else>`, "line 1: unexpected <go:else> here"},
		{"duplicate default", "<go:switch v>\n<go:default>a\n<go:default>b\n</go:switch>", "line 3: duplicate <go:default>"},
		{"if greater-than in cond needs parens", `<go:if (a > b)>x</go:if>`, ""},
		{"unclosed for", `<go:for i := range xs>texto`, "line 1: <go:for> missing </go:for>"},
		{"import vazio", `<go:import ()>`, "line 1: empty <go:import>"},
		{"erro propagado de dentro do for", `<go:for i><go:if a>x`, "line 1: <go:if> missing </go:if>"},
		{"erro propagado do corpo then do if", `<go:if a><go:for i>x`, "line 1: <go:for> missing </go:for>"},
		{"erro propagado do corpo else do if", `<go:if a>x<go:else><go:for i>y`, "line 1: <go:for> missing </go:for>"},
		{"if com else mas sem fechar", `<go:if a>x<go:else>y`, "line 1: <go:if> missing </go:if>"},
		{"erro antes do primeiro case do switch", `<go:switch v><go:if a>x`, "line 1: <go:if> missing </go:if>"},
		{"erro dentro do corpo de um case", `<go:switch v><go:case 1><go:if a>x`, "line 1: <go:if> missing </go:if>"},
		{"erro dentro do corpo do default", `<go:switch v><go:default><go:if a>x`, "line 1: <go:if> missing </go:if>"},
		{"switch sem nenhum case/default/fechamento", `<go:switch v>texto`, "line 1: <go:switch> missing </go:switch>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.src)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Parse(%q) unexpected error: %v", tt.src, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Parse(%q) = nil error, want %q", tt.src, tt.wantErr)
			}
			if err.Error() != tt.wantErr {
				t.Errorf("Parse(%q) error = %q, want %q", tt.src, err.Error(), tt.wantErr)
			}
		})
	}
}
