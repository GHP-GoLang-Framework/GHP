package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout))
}

// run executes the ghp command and returns the exit code — separated from
// main() so it can be tested without killing the test process with os.Exit.
func run(args []string, stdout io.Writer) int {
	if len(args) < 1 {
		printUsage(stdout)
		return 2
	}

	switch args[0] {
	case "dev":
		fmt.Fprintln(stdout, "Starting development server...")
	case "build":
		fmt.Fprintln(stdout, "Building GHP project...")

		fmt.Fprintln(stdout, "GHP project built")
	case "help":
		printUsage(stdout)
	default:
		fmt.Fprintln(stdout, "Command unknown")
		printUsage(stdout)
		return 2
	}

	return 0
}

func printUsage(w io.Writer) {
	fmt.Fprint(w, `
	GHP - Good Hygiene Practices

	Usage:
	  ghp <command>

	Commands:
	  dev       Start the development environment
	  build     Build the current project
	  help      Show this help message

	Run 'ghp help' for more information.
	`)
}
