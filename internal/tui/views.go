package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

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
		
	cyanStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("51"))   // Cyan for active projects
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
	// Title shows current mode (always tasks now)
	titleText := "Denote Tasks"
	if m.projectFilter {
		titleText = "Denote Projects"
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
	sortInfo := fmt.Sprintf(SortFormatString, m.sortBy)
	if m.reverseSort {
		sortInfo += " ↓"
	} else {
		sortInfo += " ↑"
	}
	
	// Status line
	itemType := "tasks"
	if m.projectFilter {
		itemType = "projects"
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
		msg := "No tasks found"
		if m.projectFilter {
			msg = "No projects found"
		}
		return helpStyle.Render(msg)
	}
	
	// Calculate visible range
	visibleHeight := m.height - HeaderFooterHeight // Leave room for header and footer
	if visibleHeight < 1 {
		visibleHeight = DefaultVisibleHeight // Default
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
	todayStr := time.Now().Format("2006-01-02")
	
	// Check if we should show divider in the visible range
	showDividerAt := -1
	if m.sortBy == "due" && !m.reverseSort {
		// Find where to show the divider in the full list
		for i := 0; i < len(m.filtered); i++ {
			file := m.filtered[i]
			if file.IsTask() {
				if task, err := denote.ParseTaskFile(file.Path); err == nil {
					// Show divider before first task that is:
					// 1. Due after today, OR
					// 2. Has no due date (and we've seen tasks with due dates)
					if (task.TaskMetadata.DueDate != "" && task.TaskMetadata.DueDate > todayStr) ||
					   (task.TaskMetadata.DueDate == "" && i > 0) {
						showDividerAt = i
						break
					}
				}
			} else if file.IsProject() {
				// Skip projects - don't show divider for them
				continue
			}
		}
	}
	
	for i := start; i < end; i++ {
		// Show divider if this is the position
		if i == showDividerAt {
			// Create a continuous line that matches the total width
			// The format string adds spaces, so we need to account for those
			// Total chars before project: roughly 106-110
			divider := strings.Repeat("─", 106) + "→ due today"
			lines = append(lines, helpStyle.Render(divider))
		}
		
		line := m.renderFileLine(i)
		lines = append(lines, line)
	}
	
	return strings.Join(lines, "\n")
}

func (m Model) renderFileLine(index int) string {
	file := m.filtered[index]
	
	// Always render task/project lines with fresh metadata
	if file.IsTask() {
		if task, err := denote.ParseTaskFile(file.Path); err == nil {
			return m.renderTaskLine(index, file, task)
		}
	} else if file.IsProject() {
		if project, err := denote.ParseProjectFile(file.Path); err == nil {
			return m.renderProjectLine(index, file, project)
		}
		// Project without metadata - show debug
		line := fmt.Sprintf("%s %s %-15s [NO METADATA] %-40s", 
			" ", ">", file.ID, truncate(file.Title, 40))
		return baseStyle.Render(line)
	}
	
	// Fallback for debugging - show basic file info
	selector := " "
	if index == m.cursor {
		selector = ">"
	}
	
	line := fmt.Sprintf("%s %-15s [UNKNOWN] %-40s", selector, file.ID, truncate(file.Title, 40))
	
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
	status := StatusSymbolOpen // open
	isDone := false
	if task.TaskMetadata.Status == denote.TaskStatusDone {
		status = StatusSymbolDone
		isDone = true
	} else if task.TaskMetadata.Status == denote.TaskStatusPaused {
		status = StatusSymbolPaused
	} else if task.TaskMetadata.Status == denote.TaskStatusDelegated {
		status = StatusSymbolDelegated
	} else if task.TaskMetadata.Status == denote.TaskStatusDropped {
		status = StatusSymbolDropped
	}
	
	// Priority with color
	priority := "    " // Default empty space for alignment
	switch task.TaskMetadata.Priority {
	case PriorityLevels[0]:
		priority = priorityHighStyle.Render("[p1]")
	case PriorityLevels[1]:
		priority = priorityMediumStyle.Render("[p2]")
	case PriorityLevels[2]:
		priority = priorityLowStyle.Render("[p3]")
	}
	
	// Estimate
	estimate := "     " // Default empty space for alignment (5 spaces)
	if task.TaskMetadata.Estimate > 0 {
		estimate = fmt.Sprintf("[%3d]", task.TaskMetadata.Estimate)
	}
	
	title := task.TaskMetadata.Title
	if title == "" {
		title = file.Title
	}
	
	area := ""
	// Only show area if we're not filtering by area
	if task.TaskMetadata.Area != "" && m.areaFilter == "" {
		area = fmt.Sprintf("(%s)", task.TaskMetadata.Area)
	}
	
	// Project name
	projectName := ""
	if task.TaskMetadata.ProjectID != "" {
		// Look up project name from cached metadata or file list
		for _, f := range m.files {
			if f.ID == task.TaskMetadata.ProjectID && f.IsProject() {
				var projTitle string
				var isActiveProject bool
				
				// Always read fresh from disk
				if proj, err := denote.ParseProjectFile(f.Path); err == nil {
					projTitle = truncate(proj.ProjectMetadata.Title, 15)
					isActiveProject = (proj.ProjectMetadata.Status == denote.ProjectStatusActive || proj.ProjectMetadata.Status == "")
				} else {
					projTitle = truncate(f.Title, 15)
					isActiveProject = true // Assume active if no metadata
				}
				
				if projTitle != "" {
					if isActiveProject {
						projectName = cyanStyle.Render("→ " + projTitle)
					} else {
						projectName = fmt.Sprintf("→ %s", projTitle)
					}
				}
				break
			}
		}
	}
	
	// Due date with consistent width
	due := ""
	isOverdue := false
	if task.TaskMetadata.DueDate != "" {
		dateStr := fmt.Sprintf("[%s]", task.TaskMetadata.DueDate)
		if denote.IsOverdue(task.TaskMetadata.DueDate) {
			// Red for overdue
			due = overdueStyle.Render(dateStr)
			isOverdue = true
		} else if denote.IsDueSoon(task.TaskMetadata.DueDate, m.config.SoonHorizon) {
			// Orange for soon
			due = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render(dateStr)
		} else {
			// Normal color for future
			due = dateStr
		}
		// Pad to consistent width (dates are typically 12 chars [YYYY-MM-DD])
		due = fmt.Sprintf("%*s", -ColumnWidthDueSpaces, due)
	} else {
		// Empty date placeholder for alignment
		due = strings.Repeat(" ", ColumnWidthDueSpaces)
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
	// Note: priority and due already have color codes, so we use %s instead of fixed width
	// Format: selector status priority estimate due title tags area project
	line := fmt.Sprintf("%s %s %s %s %s  %*s %*s %*s %s", 
		selector,
		status, 
		priority, 
		estimate,                                      // Right after priority
		due,                                           // After estimate
		-ColumnWidthTitle, truncate(title, ColumnWidthTitle),     // Good room for title (with 2 spaces before)
		-ColumnWidthTags, truncate(tagStr, ColumnWidthTags),    // Tags
		-ColumnWidthArea, truncate(area, ColumnWidthArea),      // Area (truncated for consistency)
		projectName)                                   // Project at the very end
	
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
	
	// Use same status indicator style as tasks
	status := "▶" // Project indicator
	isCompleted := false
	isActive := false
	
	// DEBUG: Check exact status matching
	switch project.ProjectMetadata.Status {
	case denote.ProjectStatusCompleted:
		status = "●"
		isCompleted = true
	case denote.ProjectStatusPaused:
		status = "◐"
	case denote.ProjectStatusCancelled:
		status = "⨯"
	case denote.ProjectStatusActive, "":
		// Active or empty status - both treated as active
		isActive = true
		status = "▶" // Keep the project indicator
	default:
		// Unexpected status - debug
		status = "?"
		// Will add debug to title later after it's defined
	}
	
	// Priority - we'll color it later based on active status
	priority := "    " // Default empty space for alignment
	priorityRaw := ""
	switch project.ProjectMetadata.Priority {
	case "p1":
		priorityRaw = "[p1]"
	case "p2":
		priorityRaw = "[p2]"
	case "p3":
		priorityRaw = "[p3]"
	}
	
	title := project.ProjectMetadata.Title
	if title == "" {
		title = file.Title
	}
	
	
	// Truncate title first
	titleTruncated := truncate(title, ColumnWidthTitle)
	
	// No special styling for status - will be handled at line level
	
	area := ""
	// Only show area if we're not filtering by area
	if project.ProjectMetadata.Area != "" && m.areaFilter == "" {
		area = fmt.Sprintf("(%s)", project.ProjectMetadata.Area)
	}
	
	// Check if overdue
	isOverdue := false
	if project.ProjectMetadata.DueDate != "" {
		isOverdue = denote.IsOverdue(project.ProjectMetadata.DueDate)
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
	
	
	// Build the line - exactly matching task format
	// Format: selector status priority due title tags area project
	
	// For priority, apply the color now
	if priorityRaw != "" {
		switch project.ProjectMetadata.Priority {
		case "p1":
			priority = priorityHighStyle.Render(priorityRaw)
		case "p2":
			priority = priorityMediumStyle.Render(priorityRaw)
		case "p3":
			priority = priorityLowStyle.Render(priorityRaw)
		}
	}
	
	// For due date, we need to pad BEFORE coloring
	dueDisplay := ""
	if project.ProjectMetadata.DueDate != "" {
		dateStr := fmt.Sprintf("[%s]", project.ProjectMetadata.DueDate)
		// Pad to 12 chars
		dateStr = fmt.Sprintf("%-12s", dateStr)
		
		if denote.IsOverdue(project.ProjectMetadata.DueDate) {
			dueDisplay = overdueStyle.Render(dateStr)
		} else if denote.IsDueSoon(project.ProjectMetadata.DueDate, m.config.SoonHorizon) {
			dueDisplay = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render(dateStr)
		} else if isActive {
			dueDisplay = cyanStyle.Render(dateStr)
		} else {
			dueDisplay = dateStr
		}
	} else {
		dueDisplay = strings.Repeat(" ", ColumnWidthDueSpaces)
		if isActive {
			dueDisplay = cyanStyle.Render(dueDisplay)
		}
	}
	
	// Prepare padded strings BEFORE applying colors
	titlePadded := fmt.Sprintf("%*s", -ColumnWidthTitle, titleTruncated)
	tagsPadded := fmt.Sprintf("%*s", -ColumnWidthTags, truncate(tagStr, ColumnWidthTags))
	areaPadded := fmt.Sprintf("%*s", -ColumnWidthArea, truncate(area, ColumnWidthArea))
	
	// Apply cyan to components if active
	statusDisplay := status
	titleDisplay := titlePadded
	tagsDisplay := tagsPadded
	areaDisplay := areaPadded
	
	if isActive {
		statusDisplay = cyanStyle.Render(status)
		titleDisplay = cyanStyle.Render(titlePadded)
		tagsDisplay = cyanStyle.Render(tagsPadded)
		areaDisplay = cyanStyle.Render(areaPadded)
	}
	
	// Build line with pre-padded, pre-colored components
	// Include empty estimate column space to align with tasks
	line := fmt.Sprintf("%s %s %s       %s  %s %s %s %s", 
		selector,
		statusDisplay, 
		priority,      // priority
		              // 6 spaces for estimate column + 1 space separator  
		dueDisplay,
		titleDisplay,
		tagsDisplay,
		areaDisplay,
		"")                      // Empty project field
	
	// No project field for projects themselves
	
	// Apply styling
	if index == m.cursor {
		return selectedStyle.Render(line)
	} else if isCompleted {
		return doneStyle.Render(line)
	} else if isOverdue && !isActive {
		// Only apply overdue style if not active
		return overdueStyle.Render(line)
	}
	
	// Default styling for other statuses
	switch project.ProjectMetadata.Status {
	case denote.ProjectStatusPaused:
		return pausedStyle.Render(line)
	case denote.ProjectStatusCancelled:
		return droppedStyle.Render(line)
	}
	
	return baseStyle.Render(line)
}

func (m Model) renderFooter() string {
	if m.mode == ModeSearch {
		// Show search input at bottom when in search mode
		prompt := "Search: " + m.searchInput + "█"
		help := MsgFuzzyMatch
		return "\n" + prompt + helpStyle.Render(help)
	}
	
	// Show appropriate hotkeys based on current view
	var help []string
	if m.projectFilter {
		// Project mode hotkeys
		help = []string{
			"j/k:nav",
			"/:search",
			"enter:view",
			"c:create project",
			"0-3:priority",
			"d:due date",
			"t:tags",
			"x:delete",
			"E:edit",
			"f:filter",
			"T:tasks",
			"S:sort",
			"?:help",
			"q:quit",
		}
	} else {
		// Task mode hotkeys
		help = []string{
			"j/k:nav",
			"/:search",
			"enter:preview",
			"c:create task",
			"0-3:priority",
			"s:state",
			"d:due date",
			"t:tags",
			"x:delete",
			"E:edit",
			"l:log",
			"f:filter",
			"P:projects",
			"S:sort",
			"?:help",
			"q:quit",
		}
	}
	
	return "\n" + helpStyle.Render(strings.Join(help, " • "))
}

func (m Model) renderHelp() string {
	help := `
Denote Tasks - Keyboard Shortcuts

Navigation:
  j/↓     Move down
  k/↑     Move up
  gg      Go to top
  G       Go to bottom

Actions (lowercase):
  Enter   Open task/project details
  c       Create new task/project
  d       Edit due date
  e       Edit estimate (tasks only)
  l       Add log entry (tasks only)
  r       Toggle sort order
  s       Change task state (open/done/etc)
  t       Edit tags
  u       Update task metadata
  x       Delete task/project
  /       Fuzzy search (use #tag for tag search)

Priority:
  0       Clear priority
  1/2/3   Set priority (p1/p2/p3)

Filters & Views (uppercase):
  E       Edit in external editor
  P       Toggle projects view
  T       Toggle tasks view
  S       Sort options menu
  f       Filter menu (area/priority/state/soon)
  
Other:
  ?       Toggle this help
  q       Quit

Press any key to continue...`

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
	// Always in task mode now
	// Full task creation form
	prompt := titleStyle.Render("Create New Task")
	
	// Build form with all fields
	var form strings.Builder
	form.WriteString("\n")
	
	// Look up project name if we have a project ID
	projectDisplay := m.createProject
	projectHint := MsgPressEnterSelect
	if m.createProject != "" {
		// Find project name
		for _, f := range m.files {
			if f.ID == m.createProject && f.IsProject() {
				// Always read fresh from disk
				if proj, err := denote.ParseProjectFile(f.Path); err == nil && proj.ProjectMetadata.Title != "" {
					projectDisplay = proj.ProjectMetadata.Title
				} else if f.Title != "" {
					projectDisplay = f.Title
				}
				break
			}
		}
		projectHint = MsgPressEnterChange
	}
	
	fields := []struct {
		label string
		value string
		hint  string
	}{
		{"Title", m.createTitle, "required"},
		{"Priority", m.createPriority, "p1, p2, p3"},
		{"Due Date", m.createDue, "YYYY-MM-DD or natural language"},
		{"Area", m.createArea, "life context"},
		{"Project", projectDisplay, projectHint},
		{"Estimate", m.createEstimate, "numeric value"},
		{"Tags", m.createTags, "space-separated"},
	}
	
	for i, field := range fields {
		if i == m.createField {
			// Active field with cursor
			if field.label == "Area" && m.areaFilter != "" {
				// Area is read-only when filtered
				form.WriteString(fmt.Sprintf("  %s: %s (inherited)\n", field.label, field.value))
			} else if field.label == "Project" {
				// Project is read-only - selection only
				form.WriteString(fmt.Sprintf("→ %s: %s", field.label, field.value))
				if field.value == "" {
					form.WriteString(fmt.Sprintf(" (%s)", field.hint))
				} else {
					form.WriteString(fmt.Sprintf(" [%s]", field.hint))
				}
				form.WriteString("\n")
			} else {
				form.WriteString(fmt.Sprintf("→ %s: %s█", field.label, field.value))
				if field.hint != "" {
					form.WriteString(fmt.Sprintf(" (%s)", field.hint))
				}
				form.WriteString("\n")
			}
		} else {
			// Inactive field
			if field.label == "Area" && m.areaFilter != "" {
				form.WriteString(fmt.Sprintf("  %s: %s (inherited)\n", field.label, field.value))
			} else if field.value != "" {
				form.WriteString(fmt.Sprintf("  %s: %s\n", field.label, field.value))
			} else {
				form.WriteString(fmt.Sprintf("  %s: (%s)\n", field.label, field.hint))
			}
		}
	}
	
	help := helpStyle.Render("\n↑/↓ to navigate, Enter to save, Esc to cancel")
	
	return prompt + baseStyle.Render(form.String()) + help
}

func (m Model) renderCreateTags() string {
	var itemType string
	if m.projectFilter {
		itemType = "Project"
	} else {
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
	// Always in task mode now
	options += `
  (d) Due date
  (p) Priority
  (j) Project
  (e) Estimate
  (t) Title
  (c) Created date
  (m) Modified date`
	
	options += `
  
  (r) Toggle reverse
  
  Esc to cancel`
	
	return prompt + current + helpStyle.Render(options)
}

func (m Model) renderStateMenu() string {
	if m.cursor >= len(m.filtered) {
		return MsgNoTaskSelected
	}
	
	file := m.filtered[m.cursor]
	// Always read fresh from disk
	task, err := denote.ParseTaskFile(file.Path)
	if err != nil {
		return "Failed to read task"
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
	// Handle project deletion from project view
	if m.viewingProject != nil && m.projectViewTab == 0 && m.mode == ModeConfirmDelete {
		prompt := titleStyle.Render("Confirm Project Deletion")
		
		warning := baseStyle.Render(fmt.Sprintf("\nAre you sure you want to delete project: %s?", m.viewingProject.ProjectMetadata.Title))
		
		// Show affected tasks if any
		affectedInfo := ""
		if len(m.affectedTasks) > 0 {
			affectedInfo = fmt.Sprintf("\n\n⚠️  This will affect %d task(s):", len(m.affectedTasks))
			for i, task := range m.affectedTasks {
				if i < 10 { // Show first 10 tasks
					affectedInfo += fmt.Sprintf("\n  • %s", task.TaskMetadata.Title)
				} else if i == 10 {
					affectedInfo += fmt.Sprintf("\n  ... and %d more", len(m.affectedTasks)-10)
					break
				}
			}
			affectedInfo += "\n\nThe project_id will be removed from these tasks."
		}
		
		options := `

  (y) Yes, delete project and clear task associations
  (n) No, cancel
  
  This action cannot be undone!`
		
		dangerStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
		
		return prompt + warning + affectedInfo + "\n" + dangerStyle.Render(options)
	}
	
	// Handle task deletion from project view
	if m.viewingProject != nil && m.projectViewTab == 1 && len(m.projectTasks) > 0 && m.mode == ModeConfirmDelete {
		task := m.projectTasks[m.projectTasksCursor]
		prompt := titleStyle.Render("Confirm Delete")
		warning := baseStyle.Render("\nAre you sure you want to delete this task?")
		fileName := baseStyle.Render(fmt.Sprintf("\n\nTask: %s", task.TaskMetadata.Title))
		
		options := `

  (y) Yes, delete
  (n) No, cancel
  
  This action cannot be undone!`
		
		dangerStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
		
		return prompt + warning + fileName + "\n" + dangerStyle.Render(options)
	}
	
	// Handle normal deletion
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

func (m Model) renderLogEntry() string {
	if m.loggingFile == nil {
		return MsgNoTaskSelected
	}
	
	title := m.loggingFile.Title
	// Always read fresh from disk
	if task, err := denote.ParseTaskFile(m.loggingFile.Path); err == nil && task.TaskMetadata.Title != "" {
		title = task.TaskMetadata.Title
	}
	
	prompt := titleStyle.Render("Add Log Entry")
	taskInfo := baseStyle.Render(fmt.Sprintf("\nTask: %s", title))
	input := baseStyle.Render(fmt.Sprintf("\n\nLog entry: %s█", m.logInput))
	help := helpStyle.Render("\n\nEnter to save, Esc to cancel")
	
	return prompt + taskInfo + input + help
}

func (m Model) renderProjectSelect() string {
	prompt := titleStyle.Render("Select Project")
	
	if len(m.projectSelectList) == 0 {
		return prompt + "\n\n" + helpStyle.Render("No projects found.\n\nPress Esc to go back")
	}
	
	// Build project list
	var lines []string
	
	// Add "None" option at the top to unassign
	selector := " "
	if m.projectSelectCursor == 0 {
		selector = ">"
	}
	noneLine := fmt.Sprintf("%s 0. ✗ (None - unassign from project)", selector)
	if m.projectSelectCursor == 0 {
		lines = append(lines, selectedStyle.Render(noneLine))
	} else {
		lines = append(lines, helpStyle.Render(noneLine))
	}
	
	// Add projects starting from index 1
	for i, project := range m.projectSelectList {
		// Selection indicator (adjust for None option at position 0)
		selector := " "
		if i+1 == m.projectSelectCursor {
			selector = ">"
		}
		
		// Number for quick selection (1-9)
		number := ""
		if i < 9 {
			number = fmt.Sprintf("%d. ", i+1)
		} else {
			number = "   "
		}
		
		// Project status indicator
		status := ""
		switch project.ProjectMetadata.Status {
		case denote.ProjectStatusActive, "":
			status = "●" // Active
		case denote.ProjectStatusCompleted:
			status = "✓" // Completed
		case denote.ProjectStatusPaused:
			status = "⏸" // Paused
		case denote.ProjectStatusCancelled:
			status = "⨯" // Cancelled
		}
		
		// Project title and area
		title := project.ProjectMetadata.Title
		if title == "" {
			title = project.File.Title
		}
		
		area := ""
		if project.ProjectMetadata.Area != "" {
			area = fmt.Sprintf(" (%s)", project.ProjectMetadata.Area)
		}
		
		// Due date
		due := ""
		if project.ProjectMetadata.DueDate != "" {
			due = fmt.Sprintf(" [%s]", project.ProjectMetadata.DueDate)
		}
		
		// Format line
		line := fmt.Sprintf("%s %s%s %s%s%s", selector, number, status, title, area, due)
		
		if i+1 == m.projectSelectCursor {
			lines = append(lines, selectedStyle.Render(line))
		} else if project.ProjectMetadata.Status == denote.ProjectStatusActive || project.ProjectMetadata.Status == "" {
			lines = append(lines, cyanStyle.Render(line))
		} else {
			lines = append(lines, baseStyle.Render(line))
		}
	}
	
	list := strings.Join(lines, "\n")
	
	help := helpStyle.Render("\n\nj/k or ↑/↓: navigate • 0: unassign • 1-9: quick select • Enter: select • Esc: cancel")
	
	return prompt + "\n\n" + list + help
}

func (m Model) renderCreateProject() string {
	itemType := "Project"
	if m.areaFilter != "" {
		itemType = fmt.Sprintf("Project (Area: %s)", m.areaFilter)
	}
	prompt := titleStyle.Render(fmt.Sprintf("Create New %s", itemType))
	input := baseStyle.Render(fmt.Sprintf("\nTitle: %s█", m.createTitle))
	help := helpStyle.Render("\nEnter to continue, Esc to cancel")
	
	return prompt + input + help
}

func (m Model) renderCreateProjectTags() string {
	itemType := "Project"
	if m.areaFilter != "" {
		itemType = fmt.Sprintf("Project (Area: %s)", m.areaFilter)
	}
	prompt := titleStyle.Render(fmt.Sprintf("Create New %s", itemType))
	titleLine := baseStyle.Render(fmt.Sprintf("\nTitle: %s", m.createTitle))
	input := baseStyle.Render(fmt.Sprintf("\nTags (space-separated): %s█", m.createTags))
	if m.areaFilter != "" {
		help := helpStyle.Render(fmt.Sprintf("\nArea '%s' will be added automatically\nEnter to create, Esc to go back", m.areaFilter))
		return prompt + titleLine + input + help
	}
	help := helpStyle.Render("\nEnter to create, Esc to go back")
	
	return prompt + titleLine + input + help
}