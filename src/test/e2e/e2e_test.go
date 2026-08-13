// Package e2e runs the compiled `ghp` binary against sample projects
// (ghp init/dev/build) and validates real HTTP responses.
package e2e

import "testing"

func TestPlaceholder(t *testing.T) {
	if testing.Short() {
		t.Skip("skipped in short mode (go test -short)")
	}
	t.Skip("no e2e tests yet — see GHP-14 / GHP-15 on Linear")
}
