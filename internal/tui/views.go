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
	// Title shows current mode
	titleText := "Denote Notes"
	if m.viewMode == ViewModeTasks {
		if m.projectFilter {
			titleText = "Denote Projects"
		} else {
			titleText = "Denote Tasks"
		}
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
		if m.projectFilter {
			itemType = "projects"
		} else {
			itemType = "tasks"
		}
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
			if m.projectFilter {
				msg = "No projects found"
			} else {
				msg = "No tasks found"
			}
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
	todayStr := time.Now().Format("2006-01-02")
	
	// Check if we should show divider in the visible range
	showDividerAt := -1
	if m.viewMode == ViewModeTasks && m.sortBy == "due" && !m.reverseSort {
		// Find where to show the divider in the full list
		for i := 0; i < len(m.filtered); i++ {
			file := m.filtered[i]
			if task, ok := m.taskMetadata[file.Path]; ok {
				// Show divider before first task that is:
				// 1. Due after today, OR
				// 2. Has no due date (and we've seen tasks with due dates)
				if (task.TaskMetadata.DueDate != "" && task.TaskMetadata.DueDate > todayStr) ||
				   (task.TaskMetadata.DueDate == "" && i > 0) {
					showDividerAt = i
					break
				}
			} else if project, ok := m.projectMetadata[file.Path]; ok {
				// Same logic for projects
				if (project.ProjectMetadata.DueDate != "" && project.ProjectMetadata.DueDate > todayStr) ||
				   (project.ProjectMetadata.DueDate == "" && i > 0) {
					showDividerAt = i
					break
				}
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
	
	// Check if we're in task mode and have task/project metadata
	if m.viewMode == ViewModeTasks {
		if task, ok := m.taskMetadata[file.Path]; ok {
			return m.renderTaskLine(index, file, task)
		}
		if project, ok := m.projectMetadata[file.Path]; ok {
			return m.renderProjectLine(index, file, project)
		} else if file.IsProject() {
			// Project without metadata - show debug
			line := fmt.Sprintf("%s %s %-15s [NO METADATA] %-40s", 
				" ", ">", file.ID, truncate(file.Title, 40))
			return baseStyle.Render(line)
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
				
				if proj, ok := m.projectMetadata[f.Path]; ok {
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
		due = fmt.Sprintf("%-12s", due)
	} else {
		// Empty date placeholder for alignment
		due = "            "  // 12 spaces
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
	// Format: selector status priority due title tags area project
	line := fmt.Sprintf("%s %s %s %s  %-50s %-25s %-10s %s", 
		selector,
		status, 
		priority, 
		due,                     // Right after priority
		truncate(title, 50),     // Good room for title (with 2 spaces before)
		truncate(tagStr, 25),    // Tags
		truncate(area, 10),      // Area (truncated for consistency)
		projectName)             // Project at the very end
	
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
	titleTruncated := truncate(title, 50)
	
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
		dueDisplay = "            "  // 12 spaces
		if isActive {
			dueDisplay = cyanStyle.Render(dueDisplay)
		}
	}
	
	// Prepare padded strings BEFORE applying colors
	titlePadded := fmt.Sprintf("%-50s", titleTruncated)
	tagsPadded := fmt.Sprintf("%-25s", truncate(tagStr, 25))
	areaPadded := fmt.Sprintf("%-10s", truncate(area, 10))
	
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
	line := fmt.Sprintf("%s %s %s %s  %s %s %s %s", 
		selector,
		statusDisplay, 
		priority, 
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
			"l:log",
			"f:filter",
			"p:projects",
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
  c       Create new task (full form with metadata)
  1/2/3   Set priority (p1/p2/p3)
  s       Change task state (open/done/etc)
  x       Delete task/project
  e       Edit task in external editor
  u       Update task metadata
  l       Add log entry to task
  /       Fuzzy search (use #tag for tag search)
  f       Filter menu (area/priority/state/soon)
  p       Toggle projects view
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
	if m.viewMode != ViewModeTasks {
		// Simple create for notes
		prompt := titleStyle.Render("Create New Note")
		input := baseStyle.Render(fmt.Sprintf("\nTitle: %s█", m.createTitle))
		help := helpStyle.Render("\nEnter to continue, Esc to cancel")
		return prompt + input + help
	}
	
	// Full task creation form
	prompt := titleStyle.Render("Create New Task")
	
	// Build form with all fields
	var form strings.Builder
	form.WriteString("\n")
	
	// Look up project name if we have a project ID
	projectDisplay := m.createProject
	projectHint := "press Enter to select"
	if m.createProject != "" {
		// Find project name
		for _, f := range m.files {
			if f.ID == m.createProject && f.IsProject() {
				if proj, ok := m.projectMetadata[f.Path]; ok && proj.ProjectMetadata.Title != "" {
					projectDisplay = proj.ProjectMetadata.Title
				} else if f.Title != "" {
					projectDisplay = f.Title
				}
				break
			}
		}
		projectHint = "press Enter to change"
	}
	
	fields := []struct {
		label string
		value string
		hint  string
	}{
		{"Title", m.createTitle, "required"},
		{"Priority", m.createPriority, "p1, p2, p3"},
		{"Due Date", m.createDue, "YYYY-MM-DD or natural language"},
		{"Area", m.areaFilter, "inherited from filter"},
		{"Project", projectDisplay, projectHint},
		{"Estimate", m.createEstimate, "time estimate"},
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
  (j) Project
  (e) Estimate
  (t) Title
  (c) Created date
  (m) Modified date`
	} else {
		options += `
  (t) Title
  (c) Created date
  (m) Modified date`
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
		return "No task selected"
	}
	
	title := m.loggingFile.Title
	if task, ok := m.taskMetadata[m.loggingFile.Path]; ok && task.TaskMetadata.Title != "" {
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