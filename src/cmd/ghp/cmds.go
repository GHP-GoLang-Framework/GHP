package main

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"ghp/src/internal/transpiler"
)

// parseDir resolves the single optional positional directory argument of
// build/dev, defaulting to "." - the current directory. Anything that
// looks like a flag, or a second argument, is rejected.
func parseDir(args []string) (string, error) {
	if len(args) > 1 {
		return "", fmt.Errorf("too many arguments: %v", args)
	}
	if len(args) == 1 {
		if strings.HasPrefix(args[0], "-") {
			return "", fmt.Errorf("unexpected flag %q (expected a directory)", args[0])
		}
		return args[0], nil
	}
	return ".", nil
}

// Build generates the Go module that serves the .ghp files under dir
// (default ".") and writes it to dir/build.
func Build(args []string, stdout io.Writer) int {
	dir, err := parseDir(args)
	if err != nil {
		fmt.Fprintf(stdout, "ghp build: %v\n", err)
		return 2
	}

	out := filepath.Join(dir, "build")

	fmt.Fprintln(stdout, "Building GHP project...")

	files, err := transpiler.Generate(dir)
	if err != nil {
		fmt.Fprintf(stdout, "ghp build: %v\n", err)
		return 1
	}

	if err := writeFiles(out, files); err != nil {
		fmt.Fprintf(stdout, "ghp build: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "GHP project built at %s\n", out)
	return 0
}

// Dev runs the .ghp files under dir (default ".") through a local server
// that restarts itself whenever a page changes, until interrupted.
func Dev(args []string, stdout io.Writer) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return runDev(ctx, args, stdout)
}

// runDev is Dev without the signal wiring, so tests can drive it with
// their own context instead of waiting on a real signal.
func runDev(ctx context.Context, args []string, stdout io.Writer) int {
	dir, err := parseDir(args)
	if err != nil {
		fmt.Fprintf(stdout, "ghp dev: %v\n", err)
		return 2
	}

	fmt.Fprintln(stdout, "Starting development server...")

	tmp, err := os.MkdirTemp("", "ghpapp-*")
	if err != nil {
		fmt.Fprintf(stdout, "ghp dev: %v\n", err)
		return 1
	}
	defer os.RemoveAll(tmp)

	bin := filepath.Join(tmp, "ghpapp")
	build := func() error {
		files, err := transpiler.Generate(dir)
		if err != nil {
			return err
		}
		if err := writeFiles(tmp, files); err != nil {
			return err
		}
		cmd := exec.Command("go", "build", "-o", bin, ".")
		cmd.Dir = tmp
		cmd.Stdout = stdout
		cmd.Stderr = stdout
		return cmd.Run()
	}

	if err := build(); err != nil {
		fmt.Fprintf(stdout, "ghp dev: go build: %v\n", err)
		return 1
	}

	server, err := startApp(ctx, bin)
	if err != nil {
		fmt.Fprintf(stdout, "ghp dev: %v\n", err)
		return 1
	}

	// Rebuild on every .ghp change, swapping the running app only after
	// the new binary builds - a broken edit keeps the current page up.
	poll := time.NewTicker(300 * time.Millisecond)
	defer poll.Stop()
	last := snapshot(dir)

	for {
		select {
		case <-ctx.Done():
			server.Stop()
			return 0
		case <-poll.C:
			now := snapshot(dir)
			if equalSnapshots(now, last) {
				continue
			}
			last = now

			fmt.Fprintln(stdout, "Change detected, rebuilding...")
			if err := build(); err != nil {
				fmt.Fprintf(stdout, "ghp dev: %v (keeping the current server running)\n", err)
				continue
			}
			server.Stop()
			server, err = startApp(ctx, bin)
			if err != nil {
				fmt.Fprintf(stdout, "ghp dev: %v\n", err)
				return 1
			}
		}
	}
}

// appServer wraps the running generated binary and knows how to stop it.
type appServer struct {
	cmd    *exec.Cmd
	cancel context.CancelFunc
}

// startApp launches bin, attaching its output to the process streams and
// letting it inherit the environment - GHP_PORT is what drives which port
// it listens on (the generated main.go defaults it to 8080). The app's
// output goes to os.Stdout/os.Stderr rather than the injectable writer
// because os/exec drains it on its own goroutines, racing any use of that
// writer (reported as a data race under -race).
func startApp(ctx context.Context, bin string) (*appServer, error) {
	childCtx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(childCtx, bin)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, err
	}
	return &appServer{cmd: cmd, cancel: cancel}, nil
}

// Stop kills the app and waits for it to exit.
func (s *appServer) Stop() {
	s.cancel()
	s.cmd.Wait()
}

// snapshot maps every .ghp file under dir to its mtime, keyed by
// slash-separated relative path - the dev server's change detector.
func snapshot(dir string) map[string]int64 {
	mod := make(map[string]int64)
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".ghp" {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return nil
		}
		mod[filepath.ToSlash(rel)] = info.ModTime().UnixNano()
		return nil
	})
	return mod
}

// equalSnapshots reports whether two snapshots describe the same pages.
func equalSnapshots(a, b map[string]int64) bool {
	if len(a) != len(b) {
		return false
	}
	for path, mod := range a {
		if b[path] != mod {
			return false
		}
	}
	return true
}

// writeFiles writes every generated file, creating parent directories.
func writeFiles(root string, files []transpiler.File) error {
	for _, f := range files {
		path := filepath.Join(root, filepath.FromSlash(f.Name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(f.Content), 0o644); err != nil {
			return err
		}
	}
	return nil
}
