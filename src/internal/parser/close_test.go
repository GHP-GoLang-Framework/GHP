package parser

import "testing"

func TestFindTagClose(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want int
	}{
		{
			// Escaped quotes: the escaped '"' (\") must not be read as
			// the end of the string, or the '>' right after it would
			// close the tag too early.
			name: "escaped quotes",
			src:  ` x == "a\"b" >resto`,
			want: len(` x == "a\"b" `), // up to the '>' right after the real closing quote
		},
		{
			name: "parentheses protect the inner >",
			src:  ` (a > b) >resto`,
			want: len(` (a > b) `),
		},
		{
			// Unlike '(' and '[', '{' does not pause closing - a
			// <go ...> may open a brace that only closes in a later tag
			// (see the comment in findTagClose), so the '>' right
			// after it must close the tag normally.
			name: "brace does not protect the inner >",
			src:  ` if x {>resto`,
			want: len(` if x {`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findTagClose(tt.src, 0)
			if got != tt.want {
				t.Errorf("findTagClose(%q, 0) = %d, want %d", tt.src, got, tt.want)
			}
		})
	}
}
