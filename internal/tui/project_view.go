package tui

import (
	"fmt"
	"strings"
	
	"github.com/charmbracelet/lipgloss"
	"github.com/pdxmph/denote-tasks/internal/denote"
)

var (
	tabStyle = lipgloss.NewStyle().
		Padding(0, 2).
		Border(lipgloss.NormalBorder(), false, false, true, false)
		
	activeTabStyle = tabStyle.Copy().
		Foreground(lipgloss.Color("214")).
		BorderForeground(lipgloss.Color("214")).
		Bold(true)
		
	inactiveTabStyle = tabStyle.Copy().
		Foreground(lipgloss.Color("241"))
)

func (m Model) renderProjectView() string {
	if m.viewingProject == nil {
		return "No project selected"
	}
	
	var sections []string
	
	// Title
	title := titleStyle.Render("Project: " + m.viewingProject.ProjectMetadata.Title)
	sections = append(sections, title)
	
	// Tabs
	tabs := m.renderProjectTabs()
	sections = append(sections, tabs)
	
	// Tab content
	var content string
	switch m.projectViewTab {
	case 0:
		content = m.renderProjectOverview()
	case 1:
		content = m.renderProjectTasks()
	}
	sections = append(sections, content)
	
	// Status message or edit prompt - JUST LIKE TASK VIEW
	if m.editingField != "" {
		prompt := fmt.Sprintf("\n%s %s", m.statusMsg, m.editBuffer)
		if m.editingField != "" {
			prompt += "█"
		}
		sections = append(sections, editingStyle.Render(prompt))
	} else if m.statusMsg != "" {
		sections = append(sections, "\n"+statusStyle.Render(m.statusMsg))
	}
	
	// Footer with hints
	hints := m.getProjectViewHints()
	footer := "\n" + hintStyle.Render(strings.Join(hints, " • "))
	sections = append(sections, footer)
	
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) renderProjectTabs() string {
	tabs := []string{}
	
	// Overview tab
	if m.projectViewTab == 0 {
		tabs = append(tabs, activeTabStyle.Render("Overview"))
	} else {
		tabs = append(tabs, inactiveTabStyle.Render("Overview"))
	}
	
	// Tasks tab with count
	taskCount := len(m.projectTasks)
	taskLabel := fmt.Sprintf("Tasks (%d)", taskCount)
	if m.projectViewTab == 1 {
		tabs = append(tabs, activeTabStyle.Render(taskLabel))
	} else {
		tabs = append(tabs, inactiveTabStyle.Render(taskLabel))
	}
	
	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

func (m Model) renderProjectOverview() string {
	project := m.viewingProject
	meta := project.ProjectMetadata
	
	var lines []string
	
	// Project metadata (similar to task view)
	lines = append(lines, m.renderField("Title", meta.Title, "t"))
	
	// Status with color
	statusValue := meta.Status
	if statusValue == "" {
		statusValue = "active"
	}
	var statusColor string
	switch statusValue {
	case denote.ProjectStatusCompleted:
		statusColor = "70" // green
	case denote.ProjectStatusPaused:
		statusColor = "214" // orange
	case denote.ProjectStatusCancelled:
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
	
	// Tags
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
	
	// Add horizontal rule
	lines = append(lines, "\n" + strings.Repeat("─", 60))
	
	// Add body content
	bodyContent := m.getProjectBodyContent()
	if bodyContent != "" {
		// Wrap long lines for readability
		maxWidth := 80
		if m.width > 0 && m.width < maxWidth {
			maxWidth = m.width - 4
		}
		wrapped := wrapText(bodyContent, maxWidth)
		lines = append(lines, "\n"+wrapped)
	} else {
		lines = append(lines, "\n"+helpStyle.Render("(no notes)"))
	}
	
	return strings.Join(lines, "\n")
}

func (m Model) renderProjectTasks() string {
	if len(m.projectTasks) == 0 {
		return "\n" + helpStyle.Render("No tasks assigned to this project")
	}
	
	var lines []string
	
	// Calculate visible range
	visibleHeight := m.height - 10 // Leave room for header, tabs, and footer
	if visibleHeight < 1 {
		visibleHeight = 10 // Default
	}
	
	start := 0
	if m.projectTasksCursor >= visibleHeight {
		start = m.projectTasksCursor - visibleHeight + 1
	}
	
	end := start + visibleHeight
	if end > len(m.projectTasks) {
		end = len(m.projectTasks)
	}
	
	// Render each task
	for i := start; i < end; i++ {
		task := m.projectTasks[i]
		line := m.renderProjectTaskLine(i, task)
		lines = append(lines, line)
	}
	
	return "\n" + strings.Join(lines, "\n")
}

func (m Model) renderProjectTaskLine(index int, task denote.Task) string {
	// Selection indicator
	selector := " "
	if m.projectViewTab == 1 && index == m.projectTasksCursor {
		selector = ">"
	}
	
	// Status
	status := "○" // open
	isDone := false
	if task.TaskMetadata.Status == denote.TaskStatusDone {
		status = "●"
		isDone = true
	} else if task.TaskMetadata.Status == denote.TaskStatusPaused {
		status = "◐"
	}
	
	// Priority
	priority := ""
	switch task.TaskMetadata.Priority {
	case "p1":
		priority = "[p1]"
	case "p2":
		priority = "[p2]"
	case "p3":
		priority = "[p3]"
	default:
		priority = "    " // Empty space for alignment
	}
	
	title := task.TaskMetadata.Title
	if title == "" {
		title = task.File.Title
	}
	
	// Due date
	due := ""
	isOverdue := false
	if task.TaskMetadata.DueDate != "" {
		if denote.IsOverdue(task.TaskMetadata.DueDate) {
			// Red for overdue
			due = " " + overdueStyle.Render(fmt.Sprintf("[%s]", task.TaskMetadata.DueDate))
			isOverdue = true
		} else if denote.IsDueSoon(task.TaskMetadata.DueDate, m.config.SoonHorizon) {
			// Orange for soon
			due = " " + lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render(fmt.Sprintf("[%s]", task.TaskMetadata.DueDate))
		} else {
			// Normal color for future
			due = fmt.Sprintf(" [%s]", task.TaskMetadata.DueDate)
		}
	}
	
	// Build the line
	line := fmt.Sprintf("%s %s %s %-40s%s", 
		selector,
		status, 
		priority, 
		truncate(title, 40), 
		due)
	
	// Apply styling
	if m.projectViewTab == 1 && index == m.projectTasksCursor {
		return selectedStyle.Render(line)
	} else if isDone {
		return doneStyle.Render(line)
	} else if isOverdue {
		return overdueStyle.Render(line)
	}
	
	// Apply priority coloring
	switch task.TaskMetadata.Priority {
	case "p1":
		return priorityHighStyle.Render(line)
	case "p2":
		return priorityMediumStyle.Render(line)
	case "p3":
		return priorityLowStyle.Render(line)
	}
	
	return line
}

func (m Model) getProjectBodyContent() string {
	if m.viewingProject != nil {
		// Extract body from content (content after frontmatter)
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

func (m Model) getProjectViewHints() []string {
	hints := []string{
		"tab:switch tabs",
		"q/esc:back",
		"e:edit file",
	}
	
	if m.projectViewTab == 0 {
		// Overview tab hints
		hints = append(hints,
			"t:title",
			"p:priority",
			"s:status",
			"d:due date",
			"a:area",
			"g:tags",
			"x:delete",
		)
	} else {
		// Tasks tab hints
		if len(m.projectTasks) > 0 {
			hints = append(hints,
				"j/k:nav",
				"enter:view task",
				"1/2/3:priority",
				"s:state",
				"x:delete",
			)
		}
	}
	
	return hints
}