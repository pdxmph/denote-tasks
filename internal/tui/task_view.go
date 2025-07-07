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
		// Wrap long lines for readability
		maxWidth := 80
		if m.width > 0 && m.width < maxWidth {
			maxWidth = m.width - 4
		}
		wrapped := wrapText(bodyContent, maxWidth)
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
		hints = append(hints, "t:estimate")
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
	lines = append(lines, m.renderField("Title", meta.Title, ""))
	
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
	lines = append(lines, m.renderField("Status", statusStyled, "s"))
	
	// Priority with color
	if meta.Priority != "" {
		var priorityColor string
		switch meta.Priority {
		case "p1":
			priorityColor = "196" // red
		case "p2":
			priorityColor = "214" // orange
		case "p3":
			priorityColor = "245" // gray
		default:
			priorityColor = "230"
		}
		priorityStyled := lipgloss.NewStyle().Foreground(lipgloss.Color(priorityColor)).Bold(true).Render(meta.Priority)
		lines = append(lines, m.renderField("Priority", priorityStyled, "p"))
	} else {
		lines = append(lines, m.renderField("Priority", "(not set)", "p"))
	}
	
	// Due Date with overdue highlighting
	if meta.DueDate != "" {
		dueValue := meta.DueDate
		if denote.IsOverdue(meta.DueDate) {
			dueValue = overdueStyle.Render(dueValue + " (OVERDUE!)")
		} else if denote.IsDueThisWeek(meta.DueDate) {
			dueValue = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render(dueValue + " (this week)")
		}
		lines = append(lines, m.renderField("Due Date", dueValue, "d"))
	} else {
		lines = append(lines, m.renderField("Due Date", "(not set)", "d"))
	}
	
	// Area
	if meta.Area != "" {
		lines = append(lines, m.renderField("Area", meta.Area, "a"))
	} else {
		lines = append(lines, m.renderField("Area", "(not set)", "a"))
	}
	
	// Time Estimate
	if meta.Estimate > 0 {
		lines = append(lines, m.renderField("Estimate", fmt.Sprintf("%d", meta.Estimate), "t"))
	} else {
		lines = append(lines, m.renderField("Estimate", "(not set)", "t"))
	}
	
	// Tags (editable)
	tagsDisplay := ""
	if len(meta.Tags) > 0 {
		tagsDisplay = strings.Join(meta.Tags, " ")
	} else if len(task.File.Tags) > 0 {
		// Show file tags if no metadata tags
		tagsDisplay = strings.Join(task.File.Tags, " ") + " (from filename)"
	} else {
		tagsDisplay = "(not set)"
	}
	lines = append(lines, m.renderField("Tags", tagsDisplay, "g"))
	
	// Other metadata
	if meta.StartDate != "" {
		lines = append(lines, m.renderField("Start Date", meta.StartDate, ""))
	}
	
	if meta.ProjectID != "" {
		lines = append(lines, m.renderField("Project", meta.ProjectID, ""))
	}
	
	if meta.Assignee != "" {
		lines = append(lines, m.renderField("Assignee", meta.Assignee, ""))
	}
	
	// File info
	lines = append(lines, "")
	lines = append(lines, m.renderField("File", m.viewingFile.Path, ""))
	lines = append(lines, m.renderField("ID", task.File.ID, ""))
	
	return strings.Join(lines, "\n")
}

func (m Model) renderProjectDetails() string {
	project := m.viewingProject
	meta := project.ProjectMetadata
	
	var lines []string
	
	// Title
	lines = append(lines, m.renderField("Title", meta.Title, ""))
	
	// Status
	statusValue := meta.Status
	if statusValue == "" {
		statusValue = "active"
	}
	lines = append(lines, m.renderField("Status", statusValue, "s"))
	
	// Priority
	if meta.Priority != "" {
		lines = append(lines, m.renderField("Priority", meta.Priority, "p"))
	} else {
		lines = append(lines, m.renderField("Priority", "(not set)", "p"))
	}
	
	// Due Date
	if meta.DueDate != "" {
		lines = append(lines, m.renderField("Due Date", meta.DueDate, "d"))
	} else {
		lines = append(lines, m.renderField("Due Date", "(not set)", "d"))
	}
	
	// Area
	if meta.Area != "" {
		lines = append(lines, m.renderField("Area", meta.Area, "a"))
	} else {
		lines = append(lines, m.renderField("Area", "(not set)", "a"))
	}
	
	// Tags (editable)
	tagsDisplay := ""
	if len(meta.Tags) > 0 {
		tagsDisplay = strings.Join(meta.Tags, " ")
	} else if len(project.File.Tags) > 0 {
		// Show file tags if no metadata tags
		tagsDisplay = strings.Join(project.File.Tags, " ") + " (from filename)"
	} else {
		tagsDisplay = "(not set)"
	}
	lines = append(lines, m.renderField("Tags", tagsDisplay, "g"))
	
	// Other metadata
	if meta.StartDate != "" {
		lines = append(lines, m.renderField("Start Date", meta.StartDate, ""))
	}
	
	// File info
	lines = append(lines, "")
	lines = append(lines, m.renderField("File", m.viewingFile.Path, ""))
	lines = append(lines, m.renderField("ID", project.File.ID, ""))
	if len(project.File.Tags) > 0 {
		lines = append(lines, m.renderField("Tags", strings.Join(project.File.Tags, ", "), ""))
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

func (m Model) renderField(label, value, hotkey string) string {
	labelText := label + ":"
	if hotkey != "" && m.editingField != hotkey {
		// Special cases for better display
		switch {
		case hotkey == "t" && label == "Estimate":
			labelText = "es(t)imate:"
		case hotkey == "g" && label == "Tags":
			labelText = "ta(g)s:"
		default:
			labelText = fmt.Sprintf("(%s)%s:", hotkey, label[1:])
		}
	}
	
	// Highlight if we're editing this field
	if m.editingField == hotkey && hotkey != "" {
		value = editingStyle.Render(value)
	}
	
	return fmt.Sprintf("  %-15s %s", 
		fieldLabelStyle.Render(labelText),
		value)
}