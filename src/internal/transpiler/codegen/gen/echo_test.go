package gen

import (
	"strings"
	"testing"

	"ghp/src/internal/ast"
)

func TestEcho(t *testing.T) {
	var b strings.Builder
	Echo(&b, ast.NewEcho("user.Name", 1))

	want := "io.WriteString(w, html.EscapeString(fmt.Sprint(user.Name)))\n"
	if got := b.String(); got != want {
		t.Errorf("Echo() = %q, want %q", got, want)
	}
}
