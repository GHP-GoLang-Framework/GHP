package main

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestBuild(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "index.ghp"), []byte("<h1>Home</h1>"), 0o644)
	os.MkdirAll(filepath.Join(dir, "blog"), 0o755)
	os.WriteFile(filepath.Join(dir, "blog", "[slug].ghp"), []byte(`<p><go= r.PathValue("slug") /></p>`), 0o644)
	out := t.TempDir()

	var buf bytes.Buffer
	if code := Build([]string{"--dir", dir, "--out", out}, &buf); code != 0 {
		t.Fatalf("Build exit = %d, want 0\nout:\n%s", code, buf.String())
	}

	for _, want := range []string{"Building GHP project...", "GHP project built"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("output missing %q\ngot: %s", want, buf.String())
		}
	}
	for _, name := range []string{"main.go", "go.mod", "pages/index.go", "pages/blog_slug.go", "pages/register.go"} {
		if _, err := os.Stat(filepath.Join(out, name)); err != nil {
			t.Errorf("missing generated file %s: %v", name, err)
		}
	}
}

func TestBuildError(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "bad.ghp"), []byte(`<go:if x/>`), 0o644)

	var buf bytes.Buffer
	if code := Build([]string{"--dir", dir, "--out", t.TempDir()}, &buf); code != 1 {
		t.Errorf("Build exit = %d, want 1\nout:\n%s", code, buf.String())
	}
	if !strings.Contains(buf.String(), "ghp build: bad.ghp") {
		t.Errorf("output missing file reference\ngot: %s", buf.String())
	}
}

func TestBuildFlagError(t *testing.T) {
	var buf bytes.Buffer
	if code := Build([]string{"--bogus"}, &buf); code != 2 {
		t.Errorf("Build exit = %d, want 2\nout:\n%s", code, buf.String())
	}
}

func TestDevServesSlugRoute(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "index.ghp"), []byte("<h1>Home</h1>"), 0o644)
	os.MkdirAll(filepath.Join(dir, "blog"), 0o755)
	os.WriteFile(filepath.Join(dir, "blog", "[slug].ghp"), []byte(`<p>Slug: <go= r.PathValue("slug") /></p>`), 0o644)

	port := freePort(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var buf bytes.Buffer
	done := make(chan int, 1)
	go func() { done <- runDev(ctx, []string{"--dir", dir, "--port", port}, &buf) }()

	base := "http://127.0.0.1:" + port
	waitFor(t, base+"/")

	resp, err := http.Get(base + "/blog/ola")
	if err != nil {
		t.Fatalf("GET /blog/ola: %v", err)
	}
	body := readBody(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /blog/ola status = %d, want 200", resp.StatusCode)
	}
	if !strings.Contains(body, "Slug: ola") {
		t.Errorf("GET /blog/ola body missing the slug\ngot: %s", body)
	}

	cancel()
	select {
	case code := <-done:
		if code != 0 {
			t.Fatalf("runDev exit = %d, want 0\nout:\n%s", code, buf.String())
		}
	case <-time.After(10 * time.Second):
		t.Fatalf("runDev did not stop after cancel\nout:\n%s", buf.String())
	}
}

// freePort asks the OS for an unused TCP port and hands it back (with a
// small race window between closing and the child binding it, which is
// acceptable for tests).
func freePort(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	return strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
}

// waitFor polls url until the server answers with 200, failing after a
// timeout so the test reports a clear error instead of hanging.
func waitFor(t *testing.T, url string) {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("server never answered %s", url)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	b := new(bytes.Buffer)
	if _, err := b.ReadFrom(resp.Body); err != nil {
		t.Fatalf("read body: %v", err)
	}
	return b.String()
}
