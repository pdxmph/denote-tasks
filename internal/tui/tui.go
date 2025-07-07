package tui

import (
	"fmt"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pdxmph/denote-tasks/internal/config"
)

// Run starts the TUI application
func Run(cfg *config.Config) error {
	model, err := NewModel(cfg)
	if err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}
	
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running program: %w", err)
	}
	
	return nil
}