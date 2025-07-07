package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/pdxmph/denote-tasks/internal/cli"
	"github.com/pdxmph/denote-tasks/internal/config"
	"github.com/pdxmph/denote-tasks/internal/tui"
)

var version = "0.1.0"
var specVersion = "2.0.0"

func main() {
	var (
		showVersion   = flag.Bool("version", false, "Show version")
		useTUI        = flag.Bool("tui", false, "Use TUI interface")
		configPath    = flag.String("config", "", "Config file path")
		areaFilter    = flag.String("area", "", "Filter by area (TUI mode)")
		showProjects  = flag.Bool("projects", false, "Show projects list (TUI mode)")
	)
	flag.Parse()

	if *showVersion {
		fmt.Printf("denote-tasks v%s (spec v%s)\n", version, specVersion)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Mode detection
	if *useTUI || (len(os.Args) == 1) || *areaFilter != "" || *showProjects {
		// Launch TUI (also launch if area filter or projects flag is specified)
		runTUI(cfg, *areaFilter, *showProjects)
	} else {
		// Run CLI
		runCLI(cfg, flag.Args())
	}
}

func runTUI(cfg *config.Config, areaFilter string, showProjects bool) {
	if err := tui.Run(cfg, areaFilter, showProjects); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runCLI(cfg *config.Config, args []string) {
	if err := cli.Run(cfg, args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}