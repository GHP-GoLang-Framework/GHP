package codegen

import (
	"fmt"
	"strings"

	"ghp/src/internal/ast"
)

// genText writes n's value to the response verbatim. %q lets fmt produce a
// valid, safely-escaped Go string literal regardless of what the source
// HTML contains (quotes, newlines, unicode) - there's no manual escaping
// to get wrong here.
func genText(b *strings.Builder, n *ast.Text) {
	fmt.Fprintf(b, "io.WriteString(w, %q)\n", n.Value)
}
