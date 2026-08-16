package codegen

import (
	"fmt"
	"strings"

	"ghp/src/internal/ast"
)

// genEcho writes n's expression to the response, escaped for safe HTML
// output by default (GHP-7's acceptance criteria: <go= ...> is never raw
// HTML unless the developer explicitly opts out - opting out isn't
// implemented yet, so today it's simply never raw).
//
// fmt.Sprint accepts any Go value - string, number, struct, the return of
// a function call - and produces its default string form, which
// html.EscapeString then makes HTML-safe. This is an interim, minimal
// escaper: the ghp runtime package (GHP-16) is expected to replace it with
// something context-aware (attribute vs. text vs. URL), the way
// html/template does. Until then, this is the simplest correct choice.
func genEcho(b *strings.Builder, n *ast.Echo) {
	fmt.Fprintf(b, "io.WriteString(w, html.EscapeString(fmt.Sprint(%s)))\n", n.Expr)
}
