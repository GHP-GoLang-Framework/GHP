// Package integration exercises multiple packages together (parser+codegen+project)
// or calls real `go build`, without needing the compiled `ghp` binary.
package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"ghp/src/internal/parser"
	"ghp/src/internal/transpiler/codegen"
)

// TestParseAssembleBuild is the GHP-11 end-to-end acceptance criterion:
// a page combining all 7 tag kinds has to produce a .go file that
// actually compiles with `go build`, not just parse cleanly on its own.
func TestParseAssembleBuild(t *testing.T) {
	if testing.Short() {
		t.Skip("skipped in short mode (go test -short)")
	}

	src := `<go:import ("strings")/>
<go
	name := "World"
/>
<h1>Hello, <go= strings.ToUpper(name) />!</h1>
<go:if (len(name) > 3)/>
<p>Long name</p>
<go:else/>
<p>Short name</p>
<go:endif/>
<go:for i := 0; i < 3; i++/>
<go:switch i/>
<go:case 0/>
<p>zero</p>
<go:default/>
<p>other</p>
<go:endswitch/>
<go:endfor/>
`

	prog, err := parser.Parse(src)
	if err != nil {
		t.Fatalf("parser.Parse: %v", err)
	}

	out, err := codegen.Assemble("pages", "Index", prog)
	if err != nil {
		t.Fatalf("codegen.Assemble: %v", err)
	}

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.go"), []byte(out), 0o644); err != nil {
		t.Fatalf("write index.go: %v", err)
	}
	goMod := "module pages\n\ngo 1.26\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = dir
	if buildOut, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s\n---\ngenerated file:\n%s", err, buildOut, out)
	}
}
