package gen

import (
	"fmt"
	"strings"

	"ghp/src/internal/ast"
)

// Echo emits an HTML-escaped expression write.
//
// Output: io.WriteString(w, html.EscapeString(fmt.Sprint(<expr>)))
//
//	b   – destination buffer
//	n   – Echo node containing the Go expression to render
func Echo(b *strings.Builder, n *ast.Echo) {
	fmt.Fprintf(b, "io.WriteString(w, html.EscapeString(fmt.Sprint(%s)))\n", n.Expr)
}
