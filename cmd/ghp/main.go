package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "dev":
		fmt.Println("Starting development server...")
	case "build":
		fmt.Println("Building GHP project...")

		fmt.Println("GHP project built")
	case "help":
		printUsage()
	default:
		fmt.Println("Command unknown")
		printUsage()
		os.Exit(2)
	}
}

func printUsage() {
	fmt.Println(`
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
