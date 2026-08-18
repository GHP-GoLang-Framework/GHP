package parser

import "testing"

func TestParseEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		wantErr string
	}{
		{"unclosed if", `<go:if a == b/>text`, "line 1: <go:if> missing <go:endif/>"},
		{"tag without />", `<go:if a == b>text<go:endif/>`, "line 1: tag not closed with '/>"},
		{"tag never closed", `<go:if a == b`, "line 1: tag not closed with '/>"},
		{"stray close", `text <go:endif/> more text`, "line 1: unexpected <go:endif/> here"},
		{"else without if", `<go:else/>hi<go:endif/>`, "line 1: unexpected <go:else/> here"},
		{"duplicate default", "<go:switch v/>\n<go:default/>a\n<go:default/>b\n<go:endswitch/>", "line 3: duplicate <go:default>"},
		{"if greater-than in cond needs parens", `<go:if (a > b)/>x<go:endif/>`, ""},
		{"unclosed for", `<go:for i := range xs/>text`, "line 1: <go:for> missing <go:endfor/>"},
		{"empty import", `<go:import ()/>`, "line 1: empty <go:import>"},
		{"error propagated from inside the for", `<go:for i/><go:if a/>x`, "line 1: <go:if> missing <go:endif/>"},
		{"error propagated from the if's then body", `<go:if a/><go:for i/>x`, "line 1: <go:for> missing <go:endfor/>"},
		{"error propagated from the if's else body", `<go:if a/>x<go:else/><go:for i/>y`, "line 1: <go:for> missing <go:endfor/>"},
		{"if with else but never closed", `<go:if a/>x<go:else/>y`, "line 1: <go:if> missing <go:endif/>"},
		{"error before the switch's first case", `<go:switch v/><go:if a/>x`, "line 1: <go:if> missing <go:endif/>"},
		{"error inside a case body", `<go:switch v/><go:case 1/><go:if a/>x`, "line 1: <go:if> missing <go:endif/>"},
		{"error inside the default body", `<go:switch v/><go:default/><go:if a/>x`, "line 1: <go:if> missing <go:endif/>"},
		{"switch with no case/default/close", `<go:switch v/>text`, "line 1: <go:switch> missing <go:endswitch/>"},
		{"chained else-if via go:else still rejected", `<go:if a/>x<go:else b/>y<go:endif/>`, "line 1: <go:else> does not take a condition - chained else-if isn't supported, use nested <go:if> instead"},
		{"elif without condition", `<go:if a/>x<go:elif/>y<go:endif/>`, "line 1: <go:elif> requires a condition"},
		{"unclosed if after elif", `<go:if a/>x<go:elif b/>y`, "line 1: <go:if> missing <go:endif/>"},
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
