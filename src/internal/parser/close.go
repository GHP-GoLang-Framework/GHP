package parser

// findTagClose scans src starting at start (the byte offset right after a
// tag's head, e.g. right after "go:if") for the '>' that closes the tag.
//
// A '>' inside (), [] or a quoted string doesn't count - this lets
// conditions like <go:if (a > b)> and <go:for i := range []int{1, 2}> use
// those characters without prematurely closing the tag.
//
// '{'/'}' are deliberately NOT tracked here, unlike '(...)'/'[...]': a
// <go ...> statement is allowed to open a brace that only closes in a
// later <go ...> tag, with HTML in between (see internal/codegen, which is
// what actually pairs them - by emitting each tag's code verbatim, in
// order, and letting go build do the matching). If '{' paused closing the
// way '('/'[' do, that first tag would never find its own '>' at all.
func findTagClose(src string, start int) int {
	depth := 0
	var quote byte // 0, '"', '\'', or '`'

	for i := start; i < len(src); i++ {
		c := src[i]

		if quote != 0 {
			switch {
			case c == '\\' && quote != '`' && i+1 < len(src):
				i++ // skip the escaped character
			case c == quote:
				quote = 0
			}
			continue
		}

		switch c {
		case '"', '\'', '`':
			quote = c
		case '(', '[':
			depth++
		case ')', ']':
			if depth > 0 {
				depth--
			}
		case '>':
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}
