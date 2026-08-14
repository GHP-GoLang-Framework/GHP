package router

import (
	"fmt"
	"go/format"
	"strings"
)

// Register generates the source of a Go function that wires every page's
// route to its handler via mux.HandleFunc, one call per Page. pkg has to
// match the package the handler functions themselves live in, since the
// generated code references them by bare name - producing that package
// (with codegen.Assemble, one function per page) isn't this package's
// job.
//
// The result is run through go/format before being returned, same as
// codegen.Assemble: idiomatic formatting, and a cheap correctness check
// along the way.
func Register(pkg string, pages []Page) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "package %s\n\n", pkg)
	b.WriteString("import \"net/http\"\n\n")
	b.WriteString("// Register wires every page's route to its handler.\n")
	b.WriteString("func Register(mux *http.ServeMux) {\n")
	for _, p := range pages {
		fmt.Fprintf(&b, "\tmux.HandleFunc(%q, %s)\n", p.Route, p.FuncName)
	}
	b.WriteString("}\n")

	formatted, err := format.Source([]byte(b.String()))
	if err != nil {
		return "", fmt.Errorf("router: arquivo gerado nao e Go valido: %w", err)
	}
	return string(formatted), nil
}
