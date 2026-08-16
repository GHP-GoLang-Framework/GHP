package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"

	"ghp/src/internal/transpiler"
)

// Build generates the Go module that serves the .ghp files under --dir
// and writes it to --out.
func Build(args []string, stdout io.Writer) int {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	dir := fs.String("dir", ".", "directory to scan for .ghp files")
	out := fs.String("out", "build", "directory to write the generated module to")
	fs.SetOutput(stdout)
	if err := fs.Parse(args); err != nil {
		return 2
	}

	fmt.Fprintln(stdout, "Building GHP project...")

	files, err := transpiler.Generate(*dir)
	if err != nil {
		fmt.Fprintf(stdout, "ghp build: %v\n", err)
		return 1
	}

	if err := writeFiles(*out, files); err != nil {
		fmt.Fprintf(stdout, "ghp build: %v\n", err)
		return 1
	}

	fmt.Fprintln(stdout, "GHP project built")
	return 0
}

// Dev generates the module, compiles it and runs it, serving the .ghp
// files under --dir on --port until the process is interrupted.
func Dev(args []string, stdout io.Writer) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return runDev(ctx, args, stdout)
}

// runDev is Dev without the signal wiring, so tests can drive it with
// their own context instead of waiting on a real signal.
func runDev(ctx context.Context, args []string, stdout io.Writer) int {
	fs := flag.NewFlagSet("dev", flag.ContinueOnError)
	dir := fs.String("dir", ".", "directory to scan for .ghp files")
	port := fs.String("port", "8080", "port to listen on")
	fs.SetOutput(stdout)
	if err := fs.Parse(args); err != nil {
		return 2
	}

	fmt.Fprintln(stdout, "Starting development server...")

	files, err := transpiler.Generate(*dir)
	if err != nil {
		fmt.Fprintf(stdout, "ghp dev: %v\n", err)
		return 1
	}

	tmp, err := os.MkdirTemp("", "ghp-dev-*")
	if err != nil {
		fmt.Fprintf(stdout, "ghp dev: %v\n", err)
		return 1
	}
	defer os.RemoveAll(tmp)

	if err := writeFiles(tmp, files); err != nil {
		fmt.Fprintf(stdout, "ghp dev: %v\n", err)
		return 1
	}

	bin := filepath.Join(tmp, "ghp-dev")
	build := exec.Command("go", "build", "-o", bin, ".")
	build.Dir = tmp
	build.Stdout = stdout
	build.Stderr = stdout
	if err := build.Run(); err != nil {
		fmt.Fprintf(stdout, "ghp dev: go build: %v\n", err)
		return 1
	}

	server := exec.CommandContext(ctx, bin)
	server.Env = append(os.Environ(), "GHP_PORT="+*port)
	server.Stdout = stdout
	server.Stderr = stdout
	if err := server.Start(); err != nil {
		fmt.Fprintf(stdout, "ghp dev: %v\n", err)
		return 1
	}

	<-ctx.Done()
	server.Wait()
	return 0
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
