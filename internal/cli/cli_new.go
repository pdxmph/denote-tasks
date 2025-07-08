package cli

import (
	"fmt"
	"os"

	"github.com/pdxmph/denote-tasks/internal/config"
	"github.com/pdxmph/denote-tasks/internal/tui"
)

// Run executes the CLI with the new command structure
func RunNew(cfg *config.Config, args []string) error {
	// Parse global flags first
	remaining, err := ParseGlobalFlags(args)
	if err != nil {
		return err
	}

	// Override config with global flags
	if globalFlags.Config != "" {
		// Reload config from specified file
		newCfg, err := config.Load(globalFlags.Config)
		if err != nil {
			return fmt.Errorf("failed to load config: %v", err)
		}
		cfg = newCfg
	}

	if globalFlags.Dir != "" {
		cfg.NotesDirectory = globalFlags.Dir
	}

	// If no arguments or just --tui, launch TUI
	if len(remaining) == 0 || (len(remaining) == 0 && globalFlags.TUI) {
		if globalFlags.TUI || len(os.Args) == 1 {
			return tui.Run(cfg)
		}
	}

	// Create root command
	root := &Command{
		Name:  "denote-tasks",
		Usage: "denote-tasks <command> [options]",
		Description: `A task management system built on the Denote file naming convention.

Commands:
  task      Manage tasks
  project   Manage projects  
  note      Manage notes

Global Options:
  --tui, -t      Launch TUI interface
  --config PATH  Use specific config file
  --dir PATH     Override notes directory
  --json         Output in JSON format
  --no-color     Disable color output
  --quiet, -q    Minimal output`,
	}

	// Add subcommands
	root.Subcommands = []*Command{
		TaskCommand(cfg),
		// TODO: ProjectCommand(cfg),
		// TODO: NoteCommand(cfg),
	}

	// Add legacy command aliases for compatibility
	if len(remaining) > 0 {
		switch remaining[0] {
		case "add":
			// Redirect to task new
			return TaskCommand(cfg).Subcommands[0].Execute(remaining[1:])
		case "list":
			// Redirect to task list
			return TaskCommand(cfg).Subcommands[1].Execute(remaining[1:])
		case "done":
			// Redirect to task done
			return TaskCommand(cfg).Subcommands[3].Execute(remaining[1:])
		}
	}

	// Execute command
	return root.Execute(remaining)
}