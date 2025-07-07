package tui

import (
	"fmt"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pdxmph/denote-tasks/internal/config"
)

// Run starts the TUI application
func Run(cfg *config.Config, areaFilter string, showProjects bool) error {
	model, err := NewModel(cfg)
	if err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}
	
	// Handle initial flags
	if showProjects {
		// Show projects list
		model.viewMode = ViewModeTasks  // Switch to task mode
		model.projectFilter = true      // Filter to show only projects
		model.sortBy = "due"            // Sort by due date
		model.reverseSort = false       // Ascending order (earliest due dates first)
		
		// Apply area filter if also provided
		if areaFilter != "" {
			model.areaFilter = areaFilter
		}
		
		model.applyFilters()
		model.sortFiles()
		model.loadVisibleMetadata()
	} else if areaFilter != "" {
		// Just area filter (tasks)
		model.areaFilter = areaFilter
		model.viewMode = ViewModeTasks  // Switch to task mode
		model.stateFilter = "active"    // Apply active filter
		model.sortBy = "due"            // Sort by due date
		model.reverseSort = false       // Ascending order (earliest due dates first)
		model.applyFilters()
		model.sortFiles()               // Apply the sort after filters
		model.loadVisibleMetadata()     // Load metadata for visible items
	}
	
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running program: %w", err)
	}
	
	return nil
}