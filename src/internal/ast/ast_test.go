package ast

import "testing"

// TestNodeLine confirma que Line() reporta o valor passado ao construtor
// para cada um dos 7 tipos de no, e que cada um deles de fato implementa
// Node (chamando node() diretamente - so o proprio pacote ast pode, ja
// que o metodo e nao exportado).
func TestNodeLine(t *testing.T) {
	tests := []struct {
		name string
		node Node
		want int
	}{
		{"Text", NewText("ola", 1), 1},
		{"Import", NewImport(nil, 2), 2},
		{"Statement", NewStatement("x := 1", 3), 3},
		{"Echo", NewEcho("x", 4), 4},
		{"If", NewIf("a", nil, nil, 5), 5},
		{"Switch", NewSwitch("a", nil, nil, 6), 6},
		{"For", NewFor("a", nil, 7), 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.node.Line(); got != tt.want {
				t.Errorf("Line() = %d, want %d", got, tt.want)
			}
			tt.node.node()
		})
	}
}
