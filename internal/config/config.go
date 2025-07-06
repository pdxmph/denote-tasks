package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config represents the application configuration
type Config struct {
	NotesDirectory string    `toml:"notes_directory"`
	Editor         string    `toml:"editor"`
	DefaultArea    string    `toml:"default_area"`
	TUI            TUIConfig `toml:"tui"`
}

// TUIConfig represents TUI-specific settings
type TUIConfig struct {
	Theme string `toml:"theme"`
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		NotesDirectory: filepath.Join(homeDir, "notes"),
		Editor:         "vim",
		DefaultArea:    "",
		TUI: TUIConfig{
			Theme: "default",
		},
	}
}

// Load reads configuration from file
func Load(path string) (*Config, error) {
	// Start with defaults
	cfg := DefaultConfig()

	// If no path specified, try default locations
	if path == "" {
		path = findConfigFile()
	}

	// If still no path, return defaults
	if path == "" {
		return cfg, nil
	}

	// Read config file
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Config file doesn't exist, use defaults
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	// Parse TOML
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Expand home directory in paths
	cfg.NotesDirectory = expandHome(cfg.NotesDirectory)

	// Validate config
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save writes configuration to file
func (c *Config) Save(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Open file for writing
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	// Write TOML
	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// Validate checks if configuration is valid
func (c *Config) Validate() error {
	// Check notes directory
	if c.NotesDirectory == "" {
		return fmt.Errorf("notes_directory cannot be empty")
	}

	// Check if notes directory exists
	if info, err := os.Stat(c.NotesDirectory); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("notes_directory does not exist: %s", c.NotesDirectory)
		}
		return fmt.Errorf("failed to check notes_directory: %w", err)
	} else if !info.IsDir() {
		return fmt.Errorf("notes_directory is not a directory: %s", c.NotesDirectory)
	}

	// Validate theme
	validThemes := []string{"default", "dark", "light", "high-contrast", "minimal"}
	themeValid := false
	for _, theme := range validThemes {
		if c.TUI.Theme == theme {
			themeValid = true
			break
		}
	}
	if !themeValid {
		return fmt.Errorf("invalid theme: %s", c.TUI.Theme)
	}

	return nil
}

// findConfigFile looks for config in standard locations
func findConfigFile() string {
	// Check XDG_CONFIG_HOME first
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		path := filepath.Join(xdgConfig, "denote-tasks", "config.toml")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Check ~/.config/denote-tasks/config.toml
	if homeDir, err := os.UserHomeDir(); err == nil {
		path := filepath.Join(homeDir, ".config", "denote-tasks", "config.toml")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// expandHome expands ~ to home directory
func expandHome(path string) string {
	if path == "" {
		return path
	}

	if path[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(homeDir, path[1:])
	}

	return path
}

// ConfigPath returns the default config file path
func ConfigPath() string {
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "denote-tasks", "config.toml")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(homeDir, ".config", "denote-tasks", "config.toml")
}