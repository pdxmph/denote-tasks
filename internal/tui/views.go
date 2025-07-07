package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/pdxmph/denote-tasks/internal/denote"
)

var (
	// Base style for normal text - brighter for better readability
	baseStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))  // Light gray, almost white
		
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99"))
		
	selectedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")). // Orange for selected
		Bold(true)
		
	helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("248"))  // Brighter for better readability in menus
		
	statusStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))
		
	// Task-specific styles
	doneStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("70"))  // Green for done tasks
		
	pausedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("243"))  // Dim gray for paused tasks
		
	delegatedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("33"))   // Blue for delegated tasks
		
	droppedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))  // Dark gray for dropped tasks
		
	overdueStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")). // Red for overdue
		Bold(true)
		
	priorityHighStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")). // Red for p1
		Bold(true)
		
	priorityMediumStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("214"))  // Orange for p2
		
	priorityLowStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("248"))  // Lighter gray for p3, better contrast
		
	projectStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("135")). // Purple for projects
		Bold(true)
)

func (m Model) renderNormal() string {
	var sections []string
	
	// Header
	header := m.renderHeader()
	sections = append(sections, header)
	
	// File list
	list := m.renderFileList()
	sections = append(sections, list)
	
	// Footer
	footer := m.renderFooter()
	sections = append(sections, footer)
	
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) renderHeader() string {
	// Title shows current mode
	titleText := "Denote Notes"
	if m.viewMode == ViewModeTasks {
		titleText = "Denote Tasks"
	}
	title := titleStyle.Render(titleText)
	
	// Filter info
	filterInfo := []string{}
	if m.searchQuery != "" {
		if strings.HasPrefix(m.searchQuery, "#") {
			filterInfo = append(filterInfo, fmt.Sprintf("Tag: %s", strings.TrimPrefix(m.searchQuery, "#")))
		} else {
			filterInfo = append(filterInfo, fmt.Sprintf("Search: %s", m.searchQuery))
		}
	}
	if m.areaFilter != "" {
		filterInfo = append(filterInfo, fmt.Sprintf("Area: %s", m.areaFilter))
	}
	if m.priorityFilter != "" {
		filterInfo = append(filterInfo, fmt.Sprintf("Priority: %s", m.priorityFilter))
	}
	if m.stateFilter != "" {
		filterInfo = append(filterInfo, fmt.Sprintf("State: %s", m.stateFilter))
	}
	if m.soonFilter {
		filterInfo = append(filterInfo, fmt.Sprintf("Soon: %dd", m.config.SoonHorizon))
	}
	
	// Sort info
	sortInfo := fmt.Sprintf("Sort: %s", m.sortBy)
	if m.reverseSort {
		sortInfo += " ↓"
	} else {
		sortInfo += " ↑"
	}
	
	// Status line
	itemType := "notes"
	if m.viewMode == ViewModeTasks {
		itemType = "tasks"
	}
	status := fmt.Sprintf("%d %s", len(m.filtered), itemType)
	if len(filterInfo) > 0 {
		status += " | " + strings.Join(filterInfo, " | ")
	}
	status += " | " + sortInfo
	if m.statusMsg != "" {
		status += " | " + m.statusMsg
	}
	
	statusLine := statusStyle.Render(status)
	
	return lipgloss.JoinVertical(lipgloss.Left, title, statusLine, "")
}

func (m Model) renderFileList() string {
	if len(m.filtered) == 0 {
		msg := "No notes found"
		if m.viewMode == ViewModeTasks {
			msg = "No tasks found"
		}
		return helpStyle.Render(msg)
	}
	
	// Calculate visible range
	visibleHeight := m.height - 6 // Leave room for header and footer
	if visibleHeight < 1 {
		visibleHeight = 20 // Default
	}
	
	start := 0
	if m.cursor >= visibleHeight {
		start = m.cursor - visibleHeight + 1
	}
	
	end := start + visibleHeight
	if end > len(m.filtered) {
		end = len(m.filtered)
	}
	
	
	var lines []string
	for i := start; i < end; i++ {
		line := m.renderFileLine(i)
		lines = append(lines, line)
	}
	
	return strings.Join(lines, "\n")
}

func (m Model) renderFileLine(index int) string {
	file := m.filtered[index]
	
	// Check if we're in task mode and have task/project metadata
	if m.viewMode == ViewModeTasks {
		if task, ok := m.taskMetadata[file.Path]; ok {
			return m.renderTaskLine(index, file, task)
		}
		if project, ok := m.projectMetadata[file.Path]; ok {
			return m.renderProjectLine(index, file, project)
		}
		// If we're in task mode but no metadata found, still show the item
		// This helps with debugging
	}
	
	// Selection indicator
	selector := " "
	if index == m.cursor {
		selector = ">"
	}
	
	// Format: ID Title [tags]
	id := file.ID
	title := file.Title
	if title == "" {
		title = file.Slug
	}
	
	tags := ""
	if len(file.Tags) > 0 {
		tags = fmt.Sprintf(" [%s]", strings.Join(file.Tags, ", "))
	}
	
	line := fmt.Sprintf("%s %-15s %-40s%s", selector, id, truncate(title, 40), tags)
	
	if index == m.cursor {
		return selectedStyle.Render(line)
	}
	return baseStyle.Render(line)
}

func (m Model) renderTaskLine(index int, file denote.File, task *denote.Task) string {
	// Selection indicator
	selector := " "
	if index == m.cursor {
		selector = ">"
	}
	
	// Format: Status Priority Title (Area) [Due Date]
	status := "○" // open
	isDone := false
	if task.TaskMetadata.Status == denote.TaskStatusDone {
		status = "✓"
		isDone = true
	} else if task.TaskMetadata.Status == denote.TaskStatusPaused {
		status = "⏸"
	} else if task.TaskMetadata.Status == denote.TaskStatusDelegated {
		status = "→"
	} else if task.TaskMetadata.Status == denote.TaskStatusDropped {
		status = "⨯"
	}
	
	// Priority with color
	priority := "    " // Default empty space for alignment
	switch task.TaskMetadata.Priority {
	case "p1":
		priority = priorityHighStyle.Render("[p1]")
	case "p2":
		priority = priorityMediumStyle.Render("[p2]")
	case "p3":
		priority = priorityLowStyle.Render("[p3]")
	}
	
	title := task.TaskMetadata.Title
	if title == "" {
		title = file.Title
	}
	
	area := ""
	if task.TaskMetadata.Area != "" {
		area = fmt.Sprintf("(%s)", task.TaskMetadata.Area)
	}
	
	// Due date with consistent width
	due := ""
	isOverdue := false
	if task.TaskMetadata.DueDate != "" {
		if denote.IsOverdue(task.TaskMetadata.DueDate) {
			due = fmt.Sprintf("[DUE: %s!]", task.TaskMetadata.DueDate)
			isOverdue = true
		} else if denote.IsDueThisWeek(task.TaskMetadata.DueDate) {
			due = fmt.Sprintf("[Due: %s]", task.TaskMetadata.DueDate)
		} else {
			due = fmt.Sprintf("[%s]", task.TaskMetadata.DueDate)
		}
	}
	
	// Tags - filter out 'task' and 'project'
	var displayTags []string
	for _, tag := range file.Tags {
		if tag != "task" && tag != "project" {
			displayTags = append(displayTags, tag)
		}
	}
	tagStr := ""
	if len(displayTags) > 0 {
		tagStr = fmt.Sprintf("[%s]", strings.Join(displayTags, ", "))
	}
	
	// Build the line with proper spacing
	// Note: priority already has color codes, so we use %s instead of fixed width
	line := fmt.Sprintf("%s %s %s %-30s %-10s %-20s %s", 
		selector,
		status, 
		priority, 
		truncate(title, 30), 
		area, 
		due,  // Now has consistent width of 20
		truncate(tagStr, 20))
	
	// Apply overall styling
	if index == m.cursor {
		return selectedStyle.Render(line)
	} else if isDone {
		return doneStyle.Render(line)
	} else if isOverdue {
		return overdueStyle.Render(line)
	}
	
	// Apply status-specific styling
	switch task.TaskMetadata.Status {
	case denote.TaskStatusPaused:
		return pausedStyle.Render(line)
	case denote.TaskStatusDelegated:
		return delegatedStyle.Render(line)
	case denote.TaskStatusDropped:
		return droppedStyle.Render(line)
	}
	
	// No need to apply priority coloring to the whole line anymore
	// since we colored the priority badge directly
	
	// Apply base style for better readability
	return baseStyle.Render(line)
}

func (m Model) renderProjectLine(index int, file denote.File, project *denote.Project) string {
	// Selection indicator
	selector := " "
	if index == m.cursor {
		selector = ">"
	}
	
	// Format: ▶ [Priority] Title (Area) [Due Date] [Status]
	status := "▶" // Project indicator
	
	// Priority with color
	priority := "    " // Default empty space for alignment
	switch project.ProjectMetadata.Priority {
	case "p1":
		priority = priorityHighStyle.Render("[p1]")
	case "p2":
		priority = priorityMediumStyle.Render("[p2]")
	case "p3":
		priority = priorityLowStyle.Render("[p3]")
	}
	
	title := project.ProjectMetadata.Title
	if title == "" {
		title = file.Title
	}
	
	area := ""
	if project.ProjectMetadata.Area != "" {
		area = fmt.Sprintf("(%s)", project.ProjectMetadata.Area)
	}
	
	// Due date with consistent width
	due := ""
	isOverdue := false
	if project.ProjectMetadata.DueDate != "" {
		if denote.IsOverdue(project.ProjectMetadata.DueDate) {
			due = fmt.Sprintf("[DUE: %s!]", project.ProjectMetadata.DueDate)
			isOverdue = true
		} else if denote.IsDueThisWeek(project.ProjectMetadata.DueDate) {
			due = fmt.Sprintf("[Due: %s]", project.ProjectMetadata.DueDate)
		} else {
			due = fmt.Sprintf("[%s]", project.ProjectMetadata.DueDate)
		}
	}
	
	// Status badge - make it shorter to fit with due date
	statusBadge := ""
	switch project.ProjectMetadata.Status {
	case denote.ProjectStatusActive:
		statusBadge = "[ACT]"
	case denote.ProjectStatusCompleted:
		statusBadge = "[DONE]"
	case denote.ProjectStatusPaused:
		statusBadge = "[PAUSE]"
	case denote.ProjectStatusCancelled:
		statusBadge = "[CANCEL]"
	}
	
	// Tags - filter out 'task' and 'project'
	var displayTags []string
	for _, tag := range file.Tags {
		if tag != "task" && tag != "project" {
			displayTags = append(displayTags, tag)
		}
	}
	tagStr := ""
	if len(displayTags) > 0 {
		tagStr = fmt.Sprintf("[%s]", strings.Join(displayTags, ", "))
	}
	
	// Build the line - with consistent spacing for all fields
	line := fmt.Sprintf("%s %s %s %-25s %-8s %-20s %-8s %s", 
		selector,
		status,
		priority,
		truncate(title, 25),
		area,
		due,  // Now has consistent width of 20
		statusBadge,
		truncate(tagStr, 15))
	
	// Apply styling with priority colors
	if index == m.cursor {
		return selectedStyle.Render(line)
	} else if isOverdue {
		return overdueStyle.Render(line)
	}
	
	// No need to apply priority coloring to the whole line anymore
	// since we colored the priority badge directly
	
	return projectStyle.Render(line)
}

func (m Model) renderFooter() string {
	if m.mode == ModeSearch {
		// Show search input at bottom when in search mode
		prompt := "Search: " + m.searchInput + "█"
		help := " (fuzzy match, #tag for tags, Esc to clear)"
		return "\n" + prompt + helpStyle.Render(help)
	}
	
	var help []string
	if m.viewMode == ViewModeTasks {
		// Task mode hotkeys
		help = []string{
			"j/k:nav",
			"/:search",
			"enter:preview",
			"1/2/3:priority",
			"s:state",
			"x:delete",
			"e:edit",
			"f:filter",
			"S:sort",
			"t:notes mode",
			"?:help",
			"q:quit",
		}
	} else {
		// Notes mode hotkeys
		help = []string{
			"j/k:nav",
			"/:search",
			"enter:preview",
			"n:new",
			"x:delete",
			"e:edit",
			"t:task mode",
			"s:sort",
			"r:reverse",
			"?:help",
			"q:quit",
		}
	}
	
	return "\n" + helpStyle.Render(strings.Join(help, " • "))
}

func (m Model) renderHelp() string {
	var help string
	if m.viewMode == ViewModeTasks {
		help = `
Denote Tasks - Keyboard Shortcuts

Navigation:
  j/↓     Move down
  k/↑     Move up
  gg      Go to top
  G       Go to bottom

Task Actions:
  Enter   Preview task
  c       Create new task (inherits area filter)
  1/2/3   Set priority (p1/p2/p3)
  s       Change task state (open/done/etc)
  x       Delete task/project
  e       Edit task in external editor
  u       Update task metadata
  /       Fuzzy search (use #tag for tag search)
  f       Filter menu (area/priority/state/soon)
  S       Sort options menu (uppercase S)
  r       Toggle sort order
  t       Switch to Notes mode
  
Other:
  ?       Toggle this help
  q       Quit

Press any key to continue...`
	} else {
		help = `
Denote Notes - Keyboard Shortcuts

Navigation:
  j/↓     Move down
  k/↑     Move up
  gg      Go to top
  G       Go to bottom

Actions:
  Enter   Preview note
  n       Create new note
  x       Delete note
  e       Edit note in external editor
  t       Switch to Task mode
  /       Fuzzy search (use #tag for tag search)
  s       Sort options menu
  r       Toggle sort order
  
Other:
  ?       Toggle this help
  q       Quit

Press any key to continue...`
	}

	return titleStyle.Render("Help") + help
}



func (m Model) renderPreview() string {
	if m.previewFile == nil {
		return "No file to preview"
	}
	
	// Read file content
	content, err := os.ReadFile(m.previewFile.Path)
	if err != nil {
		return fmt.Sprintf("Error reading file: %v", err)
	}
	
	// Simple preview - just show the content
	// TODO: Add markdown rendering, scrolling, etc.
	preview := string(content)
	
	header := titleStyle.Render(m.previewFile.Title)
	footer := helpStyle.Render("\nPress q or Esc to close")
	
	return header + "\n\n" + preview + footer
}

func (m Model) renderCreate() string {
	itemType := "Note"
	if m.viewMode == ViewModeTasks {
		itemType = "Task"
	}
	prompt := titleStyle.Render(fmt.Sprintf("Create New %s", itemType))
	input := baseStyle.Render(fmt.Sprintf("\nTitle: %s█", m.createTitle))
	help := helpStyle.Render("\nEnter to continue, Esc to cancel")
	
	return prompt + input + help
}

func (m Model) renderCreateTags() string {
	itemType := "Note"
	if m.viewMode == ViewModeTasks {
		itemType = "Task"
		if m.areaFilter != "" {
			itemType = fmt.Sprintf("Task (Area: %s)", m.areaFilter)
		}
	}
	prompt := titleStyle.Render(fmt.Sprintf("Create New %s", itemType))
	titleLine := baseStyle.Render(fmt.Sprintf("\nTitle: %s", m.createTitle))
	input := baseStyle.Render(fmt.Sprintf("\nTags (space-separated): %s█", m.createTags))
	help := helpStyle.Render("\nEnter to create, Esc to go back")
	
	return prompt + titleLine + input + help
}

func (m Model) renderAreaFilter() string {
	prompt := titleStyle.Render("Filter by Area")
	input := baseStyle.Render(fmt.Sprintf("\nArea: %s█", m.searchInput))
	help := helpStyle.Render("\nEnter to apply filter, Esc to cancel, Empty to clear filter")
	
	return prompt + input + help
}

func (m Model) renderSort() string {
	prompt := titleStyle.Render("Sort Options")
	
	// Show current sort
	current := baseStyle.Render(fmt.Sprintf("\nCurrent: %s", m.sortBy))
	if m.reverseSort {
		current = baseStyle.Render(fmt.Sprintf("\nCurrent: %s (descending)", m.sortBy))
	} else {
		current = baseStyle.Render(fmt.Sprintf("\nCurrent: %s (ascending)", m.sortBy))
	}
	
	options := "\n\nSort by:"
	if m.viewMode == ViewModeTasks {
		options += `
  (d) Due date
  (p) Priority  
  (e) Estimate
  (t) Title
  (c) Created date`
	} else {
		options += `
  (t) Title
  (c) Created date`
	}
	
	options += `
  
  (r) Toggle reverse
  
  Esc to cancel`
	
	return prompt + current + helpStyle.Render(options)
}

func (m Model) renderStateMenu() string {
	if m.cursor >= len(m.filtered) {
		return "No task selected"
	}
	
	file := m.filtered[m.cursor]
	task, ok := m.taskMetadata[file.Path]
	if !ok {
		return "Task metadata not loaded"
	}
	
	prompt := titleStyle.Render("Change Task Status")
	
	taskInfo := baseStyle.Render(fmt.Sprintf("\nTask: %s", task.TaskMetadata.Title))
	currentStatus := baseStyle.Render(fmt.Sprintf("\nCurrent status: %s", task.TaskMetadata.Status))
	
	options := `

Change to:
  (o) Open
  (p) Paused
  (d) Done
  (e) Delegated
  (r) Dropped
  
  Esc to cancel`
	
	return prompt + taskInfo + currentStatus + helpStyle.Render(options)
}

func (m Model) renderConfirmDelete() string {
	if m.cursor >= len(m.filtered) {
		return "No item selected"
	}
	
	file := m.filtered[m.cursor]
	
	prompt := titleStyle.Render("Confirm Delete")
	
	itemType := "file"
	if file.IsTask() {
		itemType = "task"
	} else if file.IsProject() {
		itemType = "project"
	}
	
	warning := baseStyle.Render(fmt.Sprintf("\nAre you sure you want to delete this %s?", itemType))
	fileName := baseStyle.Render(fmt.Sprintf("\n\nFile: %s", file.Title))
	
	options := `

  (y) Yes, delete
  (n) No, cancel
  
  This action cannot be undone!`
	
	dangerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)
	
	return prompt + warning + fileName + "\n" + dangerStyle.Render(options)
}

func (m Model) renderFilterMenu() string {
	prompt := titleStyle.Render("Filter Options")
	
	// Show current filters
	activeFilters := []string{}
	if m.areaFilter != "" {
		activeFilters = append(activeFilters, fmt.Sprintf("Area: %s", m.areaFilter))
	}
	if m.priorityFilter != "" {
		activeFilters = append(activeFilters, fmt.Sprintf("Priority: %s", m.priorityFilter))
	}
	if m.stateFilter != "" {
		activeFilters = append(activeFilters, fmt.Sprintf("State: %s", m.stateFilter))
	}
	if m.soonFilter {
		activeFilters = append(activeFilters, fmt.Sprintf("Soon: %d days", m.config.SoonHorizon))
	}
	
	current := "\n\nActive filters:"
	if len(activeFilters) == 0 {
		current += " None"
	} else {
		for _, f := range activeFilters {
			current += "\n  • " + f
		}
	}
	
	// Apply base style to current filters section
	current = baseStyle.Render(current)
	
	options := `

Filter by:
  (a) Area
  (p) Priority
  (s) State
  (d) Due soon (toggle)
  
  (c) Clear all filters
  
  Esc to close`
	
	return prompt + current + helpStyle.Render(options)
}

func (m Model) renderPriorityFilter() string {
	prompt := titleStyle.Render("Filter by Priority")
	
	current := ""
	if m.priorityFilter != "" {
		current = baseStyle.Render(fmt.Sprintf("\n\nCurrent: %s", m.priorityFilter))
	}
	
	options := `

Select priority:
  (1) P1 - High priority
  (2) P2 - Medium priority
  (3) P3 - Low priority
  
  (c) Clear priority filter
  
  Esc to cancel`
	
	return prompt + current + helpStyle.Render(options)
}

func (m Model) renderStateFilter() string {
	prompt := titleStyle.Render("Filter by State")
	
	current := ""
	if m.stateFilter != "" {
		current = baseStyle.Render(fmt.Sprintf("\n\nCurrent: %s", m.stateFilter))
	}
	
	options := `

Select state:
  (a) Active (open + delegated)
  (o) Open only
  (p) Paused
  (d) Done
  (e) Delegated only
  (r) Dropped
  
  (c) Clear state filter
  
  Esc to cancel`
	
	return prompt + current + helpStyle.Render(options)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}