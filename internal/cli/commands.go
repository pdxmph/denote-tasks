package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// Command represents a CLI command
type Command struct {
	Name        string
	Usage       string
	Description string
	Flags       *flag.FlagSet
	Run         func(cmd *Command, args []string) error
	Subcommands []*Command
}

// Execute runs the command
func (c *Command) Execute(args []string) error {
	// If this command has subcommands, check for them first
	if len(c.Subcommands) > 0 && len(args) > 0 && args[0] != "-h" && args[0] != "--help" && !strings.HasPrefix(args[0], "-") {
		// Look for subcommand
		for _, sub := range c.Subcommands {
			if sub.Name == args[0] {
				return sub.Execute(args[1:])
			}
		}
	}

	// Parse flags
	if c.Flags != nil {
		if err := c.Flags.Parse(args); err != nil {
			return err
		}
		args = c.Flags.Args()
	}

	// Run the command
	if c.Run != nil {
		return c.Run(c, args)
	}

	// No run function, show usage
	c.PrintUsage()
	return nil
}

// PrintUsage prints the command usage
func (c *Command) PrintUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s\n\n", c.Usage)
	if c.Description != "" {
		fmt.Fprintf(os.Stderr, "%s\n\n", c.Description)
	}

	if len(c.Subcommands) > 0 {
		fmt.Fprintf(os.Stderr, "Commands:\n")
		maxLen := 0
		for _, sub := range c.Subcommands {
			if len(sub.Name) > maxLen {
				maxLen = len(sub.Name)
			}
		}
		for _, sub := range c.Subcommands {
			fmt.Fprintf(os.Stderr, "  %-*s  %s\n", maxLen+2, sub.Name, strings.Split(sub.Description, "\n")[0])
		}
		fmt.Fprintf(os.Stderr, "\n")
	}

	if c.Flags != nil {
		fmt.Fprintf(os.Stderr, "Flags:\n")
		c.Flags.PrintDefaults()
	}
}

// Global flags
type GlobalFlags struct {
	Config   string
	Dir      string
	TUI      bool
	NoColor  bool
	JSON     bool
	Quiet    bool
}

var globalFlags GlobalFlags

// ParseGlobalFlags extracts global flags before command parsing
func ParseGlobalFlags(args []string) ([]string, error) {
	// Look for global flags only
	var remaining []string
	i := 0
	for i < len(args) {
		arg := args[i]
		
		// Check if this is a global flag with value
		if (arg == "--config" || arg == "--dir") && i+1 < len(args) {
			switch arg {
			case "--config":
				globalFlags.Config = args[i+1]
			case "--dir":
				globalFlags.Dir = args[i+1]
			}
			i += 2
			continue
		}
		
		// Check if this is a global flag without value
		switch arg {
		case "--tui", "-t":
			globalFlags.TUI = true
			i++
			continue
		case "--no-color":
			globalFlags.NoColor = true
			i++
			continue
		case "--json":
			globalFlags.JSON = true
			i++
			continue
		case "--quiet", "-q":
			globalFlags.Quiet = true
			i++
			continue
		}
		
		// Check for = style flags (e.g., --config=value)
		if strings.HasPrefix(arg, "--config=") {
			globalFlags.Config = strings.TrimPrefix(arg, "--config=")
			i++
			continue
		}
		if strings.HasPrefix(arg, "--dir=") {
			globalFlags.Dir = strings.TrimPrefix(arg, "--dir=")
			i++
			continue
		}
		
		// Not a global flag, keep it
		remaining = append(remaining, arg)
		i++
	}

	return remaining, nil
}