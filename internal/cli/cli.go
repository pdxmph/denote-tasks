package cli

import (
	"fmt"
	"os"

	"github.com/pdxmph/denote-tasks/internal/config"
	"github.com/pdxmph/denote-tasks/internal/tui"
)

// Run executes the CLI with task-focused command structure
func Run(cfg *config.Config, args []string) error {
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
	if len(remaining) == 0 || globalFlags.TUI {
		if globalFlags.TUI || len(os.Args) == 1 {
			if globalFlags.Area != "" {
				return tui.Run(cfg, globalFlags.Area)
			}
			return tui.Run(cfg)
		}
	}

	// Create root command
	root := &Command{
		Name:  "denote-tasks",
		Usage: "denote-tasks <command> [options]",
		Description: `A focused task management tool using Denote file naming convention.

Task Commands (implicit):
  new        Create a new task
  list       List tasks
  show       Show task details
  update     Update task metadata
  done       Mark tasks as done
  log        Add log entry to task

Project Commands:
  project new      Create a new project
  project list     List projects
  project show     Show project details
  project update   Update project metadata
  project tasks    Show tasks for a project

Other Commands:
  completion  Generate shell completions

Global Options:
  --tui, -t      Launch TUI interface
  --area AREA    Filter by area (for TUI or commands)
  --config PATH  Use specific config file
  --dir PATH     Override task directory
  --json         Output in JSON format
  --no-color     Disable color output
  --quiet, -q    Minimal output`,
	}

	// Get task commands and add them directly to root
	taskCmd := TaskCommand(cfg)
	for _, cmd := range taskCmd.Subcommands {
		root.Subcommands = append(root.Subcommands, cmd)
	}
	
	// Add project and completion commands
	root.Subcommands = append(root.Subcommands, 
		ProjectCommand(cfg),
		CompletionCommand(cfg),
	)

	// Execute command
	return root.Execute(remaining)
}