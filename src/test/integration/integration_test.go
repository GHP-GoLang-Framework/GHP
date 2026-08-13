// Package integration exercises multiple packages together (parser+codegen+project)
// or calls real `go build`, without needing the compiled `ghp` binary.
package integration

import "testing"

func TestPlaceholder(t *testing.T) {
	if testing.Short() {
		t.Skip("skipped in short mode (go test -short)")
	}
	t.Skip("no integration tests yet — see GHP-13 on Linear")
}
