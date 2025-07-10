package tui

// Example of refactored handlePreviewKeys using new helpers
// This shows how the improvements reduce code and increase clarity

import tea "github.com/charmbracelet/bubbletea"

// BEFORE: Original handlePreviewKeys (30+ lines of navigation code)
/*
func (m Model) handlePreviewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	const scrollAmount = 5
	
	switch msg.String() {
	case "q", "esc", "enter":
		m.mode = ModeNormal
		m.previewFile = nil
		
	case "j", "down":
		if m.previewScroll < m.previewMaxScroll {
			m.previewScroll++
		}
		
	case "k", "up":
		if m.previewScroll > 0 {
			m.previewScroll--
		}
		
	case "ctrl+d":
		// Page down
		m.previewScroll += scrollAmount
		if m.previewScroll > m.previewMaxScroll {
			m.previewScroll = m.previewMaxScroll
		}
		
	case "ctrl+u":
		// Page up
		m.previewScroll -= scrollAmount
		if m.previewScroll < 0 {
			m.previewScroll = 0
		}
	// ... more cases
	}
}
*/

// AFTER: Refactored using NavigationHandler
func (m Model) handlePreviewKeysRefactored(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "enter":
		m.mode = ModeNormal
		m.previewFile = nil
		
	case "e":
		if m.config.Editor != "" && m.previewFile != nil {
			return m, m.editFile(m.previewFile.Path)
		}
		m.statusMsg = MsgNoEditor
		
	default:
		// Handle all navigation with one line
		nav := NewNavigationHandler(m.previewMaxScroll+1, false)
		m.previewScroll = nav.HandleKey(msg.String())
	}
	
	return m, nil
}

// Example of using constants for status messages
func (m Model) handleTaskActionsRefactored(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.cursor >= len(m.filtered) {
		m.statusMsg = MsgNoTaskSelected
		return m, nil
	}
	
	file := m.filtered[m.cursor]
	
	switch msg.String() {
	case "0":
		// Clear priority using constants
		if err := m.updateTaskPriority(""); err != nil {
			m.statusMsg = formatError("clear priority", err)
		} else {
			m.statusMsg = "Priority removed"
		}
		
	case "1", "2", "3":
		// Set priority using constants
		priority := "p" + msg.String()
		if err := m.updateTaskPriority(priority); err != nil {
			m.statusMsg = formatError("set priority", err)
		} else {
			m.statusMsg = fmt.Sprintf("Priority set to %s", priority)
		}
	}
	
	return m, nil
}

// Helper for consistent error formatting
func formatError(operation string, err error) string {
	return fmt.Sprintf("Failed to %s: %v", operation, err)
}

// Example of using FieldRenderer in task view
func (m Model) renderTaskFieldsRefactored(task *denote.Task) []string {
	renderer := NewFieldRenderer()
	meta := task.TaskMetadata
	
	var lines []string
	
	// Each field is now one clean line instead of 10-20 lines
	lines = append(lines, renderer.RenderStatus(meta.Status, 
		m.editingField == string(FieldStatus), m.editBuffer))
	
	lines = append(lines, renderer.RenderPriority(meta.Priority,
		m.editingField == string(FieldPriority), m.editBuffer))
	
	lines = append(lines, renderer.RenderDueDate(meta.DueDate,
		m.editingField == string(FieldDueDate), m.editBuffer))
		
	lines = append(lines, renderer.RenderField("Area", meta.Area, "not set",
		m.editingField == string(FieldArea), m.editBuffer))
		
	lines = append(lines, renderer.RenderTags(meta.Tags,
		m.editingField == string(FieldTags), m.editBuffer))
	
	return lines
}

// Example of mode handler map pattern
var modeHandlers = map[Mode]struct {
	handleKeys func(Model, tea.KeyMsg) (tea.Model, tea.Cmd)
	render     func(Model) string
}{
	ModeNormal:   {handleNormalKeys, renderNormal},
	ModeHelp:     {handleHelpKeys, renderHelp},
	ModeSearch:   {handleSearchKeys, renderSearch},
	ModePreview:  {handlePreviewKeysRefactored, renderPreview},
	ModeTaskView: {handleTaskViewKeys, renderTaskView},
	// ... etc
}

// Simplified main handler using map
func (m Model) handleKeyPressRefactored(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if handler, ok := modeHandlers[m.mode]; ok {
		return handler.handleKeys(m, msg)
	}
	return m, nil
}