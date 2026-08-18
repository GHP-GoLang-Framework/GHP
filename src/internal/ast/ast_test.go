package ast

import "testing"

// TestNodeLine confirms that Line() reports the value passed to the
// constructor for each of the 7 node types, and that each of them actually
// implements Node (by calling node() directly - only the ast package itself
// can, since the method is unexported).
func TestNodeLine(t *testing.T) {
	tests := []struct {
		name string
		node Node
		want int
	}{
		{"Text", NewText("hi", 1), 1},
		{"Import", NewImport(nil, 2), 2},
		{"Statement", NewStatement("x := 1", 3), 3},
		{"Echo", NewEcho("x", 4), 4},
		{"If", NewIf("a", nil, nil, nil, 5), 5},
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
