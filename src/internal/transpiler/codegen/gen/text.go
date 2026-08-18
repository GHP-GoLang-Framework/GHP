package gen

import (
	"fmt"
	"strings"

	"ghp/src/internal/ast"
)

// Text emits a raw string write. The value is %q-escaped so the
// generated literal is always valid Go regardless of quotes or newlines.
//
// Output: io.WriteString(w, "<escaped value>")
//
//	b   – destination buffer
//	n   – Text node whose Value is written verbatim to the response
func Text(b *strings.Builder, n *ast.Text) {
	fmt.Fprintf(b, "io.WriteString(w, %q)\n", n.Value)
}
