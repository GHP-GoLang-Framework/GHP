package codegen

import (
	"fmt"
	"strings"

	"ghp/src/internal/ast"
)

// genStatement writes n's code to the output verbatim - it's already Go,
// there's nothing to translate.
//
// This is also what makes a <go ...> block able to open a brace that only
// closes in a later <go ...> tag, with HTML in between: each <go ...> tag
// becomes its own independent Statement node (see internal/parser), and
// this function never wraps or reorders them. Emitted one after another in
// source order, an unmatched "{" here and the "}" that closes it two tags
// later land in the generated file exactly where they'd need to for Go
// itself to parse them as one block - the parser doesn't have to know
// about this at all, `go build` does the pairing for free.
func genStatement(b *strings.Builder, file string, n *ast.Statement) {
	fmt.Fprintf(b, "//line %s:%d\n", file, n.Line())
	fmt.Fprintf(b, "%s\n", n.Code)
}
