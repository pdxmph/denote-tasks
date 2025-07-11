package tui

import (
	"fmt"
	"strings"
	
	"github.com/charmbracelet/lipgloss"
	"github.com/pdxmph/denote-tasks/internal/denote"
)

var (
	fieldLabelStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99"))
		
	fieldValueStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("230"))
		
	editingStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Bold(true)
		
	hintStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true)
)

func (m Model) renderTaskView() string {
	if m.viewingTask == nil && m.viewingProject == nil {
		return "No task or project selected"
	}
	
	var sections []string
	
	// Title
	if m.viewingTask != nil {
		title := titleStyle.Render("Task Details")
		sections = append(sections, title)
		sections = append(sections, m.renderTaskDetails())
	} else {
		title := titleStyle.Render("Project Details")
		sections = append(sections, title)
		sections = append(sections, m.renderProjectDetails())
	}
	
	// Status message or edit prompt
	if m.editingField != "" {
		prompt := fmt.Sprintf("\n%s %s", m.statusMsg, m.editBuffer)
		if m.editingField != "" {
			prompt += "█"
		}
		sections = append(sections, editingStyle.Render(prompt))
	} else if m.statusMsg != "" {
		sections = append(sections, "\n"+statusStyle.Render(m.statusMsg))
	}
	
	// Add horizontal rule
	sections = append(sections, "\n" + strings.Repeat("─", 60))
	
	// Add body content
	bodyContent := m.getBodyContent()
	if bodyContent != "" {
		// Process content line by line to highlight log entries
		lines := strings.Split(bodyContent, "\n")
		var styledLines []string
		
		for _, line := range lines {
			// Highlight log entries
			if strings.Contains(line, "]:") && strings.HasPrefix(line, "[") {
				// This looks like a log entry
				styledLines = append(styledLines, lipgloss.NewStyle().Foreground(lipgloss.Color("99")).Render(line))
			} else {
				styledLines = append(styledLines, line)
			}
		}
		
		// Join and wrap for readability
		styledContent := strings.Join(styledLines, "\n")
		maxWidth := 80
		if m.width > 0 && m.width < maxWidth {
			maxWidth = m.width - 4
		}
		wrapped := wrapText(styledContent, maxWidth)
		sections = append(sections, "\n"+wrapped)
	} else {
		sections = append(sections, "\n"+helpStyle.Render("(no notes)"))
	}
	
	// Footer with hints
	hints := []string{
		"q/esc:back",
		"e:edit file",
		"p:priority",
		"s:status",
		"d:due date",
		"a:area",
		"g:tags",
		"r:rename",
	}
	if m.viewingTask != nil {
		hints = append(hints, "j:project")
		hints = append(hints, "t:estimate")
		hints = append(hints, "l:log")
	}
	footer := "\n" + hintStyle.Render(strings.Join(hints, " • "))
	sections = append(sections, footer)
	
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) renderTaskDetails() string {
	task := m.viewingTask
	meta := task.TaskMetadata
	
	var lines []string
	
	// Title
	lines = append(lines, m.renderFieldWithHotkey("Title", meta.Title, "", ""))
	
	// Status with color
	statusValue := meta.Status
	if statusValue == "" {
		statusValue = "open"
	}
	var statusColor string
	switch statusValue {
	case denote.TaskStatusDone:
		statusColor = "70" // green
	case denote.TaskStatusPaused:
		statusColor = "214" // orange
	case denote.TaskStatusDropped:
		statusColor = "241" // gray
	default:
		statusColor = "230" // default
	}
	statusStyled := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor)).Render(statusValue)
	lines = append(lines, m.renderFieldWithHotkey("Status", statusStyled, "open", "s"))
	
	// Priority - use specialized renderer
	isEditing := m.editingField == "p"
	editBuf := ""
	if isEditing {
		editBuf = m.editBuffer
	}
	priorityLine := m.fieldRenderer.RenderPriority(meta.Priority, isEditing, editBuf)
	// Add hotkey hint if not editing
	if !isEditing {
		priorityLine = strings.Replace(priorityLine, "Priority    :", "(p)riority  :", 1)
	}
	lines = append(lines, "  " + priorityLine)
	
	// Due Date - use specialized renderer
	dueDateEdit := m.editingField == "d"
	dueBuf := ""
	if dueDateEdit {
		dueBuf = m.editBuffer
	}
	dueLine := m.fieldRenderer.RenderDueDate(meta.DueDate, dueDateEdit, dueBuf)
	// Add hotkey hint if not editing
	if !dueDateEdit {
		dueLine = strings.Replace(dueLine, "Due Date    :", "(d)ue Date  :", 1)
	}
	lines = append(lines, "  " + dueLine)
	
	// Area
	lines = append(lines, m.renderFieldWithHotkey("Area", meta.Area, "not set", "a"))
	
	// Time Estimate
	if meta.Estimate > 0 {
		lines = append(lines, m.renderFieldWithHotkey("Estimate", fmt.Sprintf("%d", meta.Estimate), "not set", "t"))
	} else {
		lines = append(lines, m.renderFieldWithHotkey("Estimate", "", "not set", "t"))
	}
	
	// Tags (editable)
	tagsDisplay := ""
	if len(meta.Tags) > 0 {
		tagsDisplay = strings.Join(meta.Tags, " ")
	} else if len(task.File.Tags) > 0 {
		// Show file tags if no metadata tags
		tagsDisplay = strings.Join(task.File.Tags, " ") + " (from filename)"
	} else {
		tagsDisplay = ""
	}
	lines = append(lines, m.renderFieldWithHotkey("Tags", tagsDisplay, "not set", "g"))
	
	// Other metadata
	if meta.StartDate != "" {
		lines = append(lines, m.renderFieldWithHotkey("Start Date", meta.StartDate, "not set", ""))
	}
	
	// Project with name lookup
	if meta.ProjectID != "" {
		projectName := meta.ProjectID
		// Look up project name from cached metadata or file list
		for _, f := range m.files {
			if f.ID == meta.ProjectID && f.IsProject() {
				if proj, ok := m.projectMetadata[f.Path]; ok && proj.ProjectMetadata.Title != "" {
					projectName = proj.ProjectMetadata.Title
				} else if f.Title != "" {
					projectName = f.Title
				}
				break
			}
		}
		lines = append(lines, m.renderFieldWithHotkey("Project", projectName, "not set", "j"))
	} else {
		lines = append(lines, m.renderFieldWithHotkey("Project", "", "not set", "j"))
	}
	
	if meta.Assignee != "" {
		lines = append(lines, m.renderFieldWithHotkey("Assignee", meta.Assignee, "not set", ""))
	}
	
	// File info
	lines = append(lines, "")
	lines = append(lines, m.renderFieldWithHotkey("File", m.viewingFile.Path, "", ""))
	lines = append(lines, m.renderFieldWithHotkey("ID", task.File.ID, "", ""))
	
	
	return strings.Join(lines, "\n")
}

func (m Model) renderProjectDetails() string {
	project := m.viewingProject
	meta := project.ProjectMetadata
	
	var lines []string
	
	// Title
	lines = append(lines, m.renderFieldWithHotkey("Title", meta.Title, "not set", ""))
	
	// Status
	statusValue := meta.Status
	if statusValue == "" {
		statusValue = "active"
	}
	lines = append(lines, m.renderFieldWithHotkey("Status", statusValue, "not set", "s"))
	
	// Priority
	if meta.Priority != "" {
		lines = append(lines, m.renderFieldWithHotkey("Priority", meta.Priority, "not set", "p"))
	} else {
		lines = append(lines, m.renderFieldWithHotkey("Priority", "", "not set", "p"))
	}
	
	// Due Date
	if meta.DueDate != "" {
		lines = append(lines, m.renderFieldWithHotkey("Due Date", meta.DueDate, "not set", "d"))
	} else {
		lines = append(lines, m.renderFieldWithHotkey("Due Date", "", "not set", "d"))
	}
	
	// Area
	if meta.Area != "" {
		lines = append(lines, m.renderFieldWithHotkey("Area", meta.Area, "not set", "a"))
	} else {
		lines = append(lines, m.renderFieldWithHotkey("Area", "", "not set", "a"))
	}
	
	// Tags (editable)
	tagsDisplay := ""
	if len(meta.Tags) > 0 {
		tagsDisplay = strings.Join(meta.Tags, " ")
	} else if len(project.File.Tags) > 0 {
		// Show file tags if no metadata tags
		tagsDisplay = strings.Join(project.File.Tags, " ") + " (from filename)"
	} else {
		tagsDisplay = ""
	}
	lines = append(lines, m.renderFieldWithHotkey("Tags", tagsDisplay, "not set", "g"))
	
	// Other metadata
	if meta.StartDate != "" {
		lines = append(lines, m.renderFieldWithHotkey("Start Date", meta.StartDate, "not set", ""))
	}
	
	// File info
	lines = append(lines, "")
	lines = append(lines, m.renderFieldWithHotkey("File", m.viewingFile.Path, "", ""))
	lines = append(lines, m.renderFieldWithHotkey("ID", project.File.ID, "", ""))
	if len(project.File.Tags) > 0 {
		lines = append(lines, m.renderFieldWithHotkey("Tags", strings.Join(project.File.Tags, ", "), "not set", ""))
	}
	
	return strings.Join(lines, "\n")
}

func (m Model) getBodyContent() string {
	if m.viewingTask != nil {
		// Extract body from content (content after frontmatter)
		content := m.viewingTask.Content
		
		// Find the end of frontmatter
		if strings.HasPrefix(content, "---\n") {
			// Find the closing ---
			parts := strings.SplitN(content, "\n---\n", 3)
			if len(parts) >= 2 {
				body := strings.TrimSpace(parts[1])
				if len(parts) == 3 {
					body = strings.TrimSpace(parts[2])
				}
				return body
			}
		}
		
		// If no frontmatter, return the whole content
		return strings.TrimSpace(content)
	} else if m.viewingProject != nil {
		// Same for projects
		content := m.viewingProject.Content
		
		// Find the end of frontmatter
		if strings.HasPrefix(content, "---\n") {
			// Find the closing ---
			parts := strings.SplitN(content, "\n---\n", 3)
			if len(parts) >= 2 {
				body := strings.TrimSpace(parts[1])
				if len(parts) == 3 {
					body = strings.TrimSpace(parts[2])
				}
				return body
			}
		}
		
		// If no frontmatter, return the whole content
		return strings.TrimSpace(content)
	}
	
	return ""
}

// wrapText wraps text to the specified width
func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}
	
	var result []string
	lines := strings.Split(text, "\n")
	
	for _, line := range lines {
		// If line is already short enough, add it
		if len(line) <= width {
			result = append(result, line)
			continue
		}
		
		// Wrap long lines
		words := strings.Fields(line)
		current := ""
		for _, word := range words {
			if current == "" {
				current = word
			} else if len(current)+1+len(word) <= width {
				current += " " + word
			} else {
				result = append(result, current)
				current = word
			}
		}
		if current != "" {
			result = append(result, current)
		}
	}
	
	return strings.Join(result, "\n")
}

func (m Model) renderFieldWithHotkey(label, value, emptyText, hotkey string) string {
	// Special label formatting for hotkeys
	displayLabel := label
	if hotkey != "" && m.editingField != hotkey {
		// Special cases for better display
		switch {
		case hotkey == "t" && label == "Estimate":
			displayLabel = "es(t)imate"
		case hotkey == "g" && label == "Tags":
			displayLabel = "ta(g)s"
		case hotkey == "j" && label == "Project":
			displayLabel = "pro(j)ect"
		default:
			displayLabel = fmt.Sprintf("(%s)%s", hotkey, label[1:])
		}
	}
	
	// Determine if we're editing this field
	isEditing := m.editingField == hotkey && hotkey != ""
	editBuf := ""
	if isEditing {
		editBuf = m.editBuffer
	}
	
	// Use the field renderer for consistent formatting
	rendered := m.fieldRenderer.RenderField(displayLabel, value, emptyText, isEditing, editBuf)
	return "  " + rendered
}