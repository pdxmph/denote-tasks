package main

import (
	"fmt"
	"os"

	"github.com/pdxmph/denote-tasks/internal/cli"
	"github.com/pdxmph/denote-tasks/internal/config"
)

var version = "0.1.0"
var specVersion = "2.0.0"

func main() {
	// Check for version flag early
	for _, arg := range os.Args[1:] {
		if arg == "--version" || arg == "-version" {
			fmt.Printf("denote-tasks v%s (spec v%s)\n", version, specVersion)
			os.Exit(0)
		}
	}

	// Load initial config (may be overridden by global flags)
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Run CLI with task-focused commands
	if err := cli.Run(cfg, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}