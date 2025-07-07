package tui

import (
	"fmt"
	
	"github.com/charmbracelet/bubbletea"
	"github.com/pdxmph/denote-tasks/internal/denote"
)

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case ModeHelp:
		return m.handleHelpKeys(msg)
	case ModeSearch:
		return m.handleSearchKeys(msg)
	case ModePreview:
		return m.handlePreviewKeys(msg)
	case ModeCreate:
		return m.handleCreateKeys(msg)
	case ModeCreateTags:
		return m.handleCreateTagsKeys(msg)
	case ModeTaskView:
		return m.handleTaskViewKeys(msg)
	case ModeProjectView:
		return m.handleProjectViewKeys(msg)
	case ModeAreaFilter:
		return m.handleAreaFilterKeys(msg)
	case ModeSort:
		return m.handleSortKeys(msg)
	case ModeStateMenu:
		return m.handleStateMenuKeys(msg)
	case ModeConfirmDelete:
		return m.handleConfirmDeleteKeys(msg)
	case ModeFilterMenu:
		return m.handleFilterMenuKeys(msg)
	case ModePriorityFilter:
		return m.handlePriorityFilterKeys(msg)
	case ModeStateFilter:
		return m.handleStateFilterKeys(msg)
	default:
		return m.handleNormalKeys(msg)
	}
}

func (m Model) handleNormalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Clear status message on any key press
	m.statusMsg = ""
	
	// Check if we're in task mode and need different key handling
	if m.viewMode == ViewModeTasks {
		return m.handleTaskModeKeys(msg)
	}
	
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
		
	case "j", "down":
		if m.cursor < len(m.filtered)-1 {
			m.cursor++
		}
		
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
		
	case "g":
		if m.lastKey == "g" {
			m.cursor = 0
			m.lastKey = ""
		} else {
			m.lastKey = "g"
		}
		
	case "G":
		if len(m.filtered) > 0 {
			m.cursor = len(m.filtered) - 1
		}
		
	case "/":
		m.mode = ModeSearch
		m.searchInput = m.searchQuery
		
	case "enter", "o":
		if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
			m.mode = ModePreview
			m.previewFile = &m.filtered[m.cursor]
			m.previewScroll = 0
		}
		
	case "e":
		if m.config.Editor != "" && m.cursor < len(m.filtered) {
			file := m.filtered[m.cursor]
			return m, m.editFile(file.Path)
		} else {
			m.statusMsg = "No editor configured"
		}
		
	case "?":
		m.mode = ModeHelp
		
	case "n":
		m.mode = ModeCreate
		m.createTitle = ""
		m.createTags = ""
		
	case "s":
		// Enter sort mode
		m.mode = ModeSort
		
	case "r":
		// Toggle reverse sort
		m.reverseSort = !m.reverseSort
		m.sortFiles()
		
	case "t":
		// Toggle to Task mode
		m.viewMode = ViewModeTasks
		m.statusMsg = "Task Mode"
		m.cursor = 0
		// Use configured defaults for tasks
		m.sortBy = m.config.Tasks.SortBy
		if m.sortBy == "" {
			m.sortBy = "due"
		}
		m.reverseSort = m.config.Tasks.SortOrder == "reverse"
		// Default to "active" filter (open + delegated)
		m.stateFilter = "active"
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
		
	case "x", "delete":
		// Delete note confirmation
		if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
			m.mode = ModeConfirmDelete
		}
		
	default:
		m.lastKey = ""
	}
	
	return m, nil
}

func (m Model) handleHelpKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "?":
		m.mode = ModeNormal
	}
	return m, nil
}

func (m Model) handleSearchKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Cancel search and restore previous state
		m.mode = ModeNormal
		m.searchInput = ""
		m.searchQuery = ""
		m.cursor = 0
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
		
	case "ctrl+c":
		// Exit search and keep current results
		m.mode = ModeNormal
		
	case "enter":
		// Confirm search and exit search mode
		m.mode = ModeNormal
		
	case "backspace":
		if len(m.searchInput) > 0 {
			m.searchInput = m.searchInput[:len(m.searchInput)-1]
			m.searchQuery = m.searchInput
			m.cursor = 0
			m.applyFilters()
			m.sortFiles()
			m.loadVisibleMetadata()
		}
		
	default:
		if len(msg.String()) == 1 {
			m.searchInput += msg.String()
			m.searchQuery = m.searchInput
			m.cursor = 0
			m.applyFilters()
			m.sortFiles()
			m.loadVisibleMetadata()
		}
	}
	
	return m, nil
}

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
		
	case "e":
		if m.config.Editor != "" && m.previewFile != nil {
			return m, m.editFile(m.previewFile.Path)
		}
	}
	
	return m, nil
}

func (m Model) handleCreateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		m.mode = ModeNormal
		m.createTitle = ""
		
	case "enter":
		if m.createTitle != "" {
			m.mode = ModeCreateTags
		}
		
	case "backspace":
		if len(m.createTitle) > 0 {
			m.createTitle = m.createTitle[:len(m.createTitle)-1]
		}
		
	default:
		if len(msg.String()) == 1 {
			m.createTitle += msg.String()
		}
	}
	
	return m, nil
}

func (m Model) handleCreateTagsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		m.mode = ModeNormal
		m.createTitle = ""
		m.createTags = ""
		
	case "enter":
		m.mode = ModeNormal
		return m, m.create()
		
	case "backspace":
		if len(m.createTags) > 0 {
			m.createTags = m.createTags[:len(m.createTags)-1]
		}
		
	default:
		if len(msg.String()) == 1 {
			m.createTags += msg.String()
		}
	}
	
	return m, nil
}

func (m Model) handleTaskModeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	// Common navigation keys work the same
	case "q", "ctrl+c":
		return m, tea.Quit
		
	case "j", "down":
		if m.cursor < len(m.filtered)-1 {
			m.cursor++
			m.loadVisibleMetadata()
		}
		
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
			m.loadVisibleMetadata()
		}
		
	case "g":
		if m.lastKey == "g" {
			m.cursor = 0
			m.lastKey = ""
		} else {
			m.lastKey = "g"
		}
		
	case "G":
		if len(m.filtered) > 0 {
			m.cursor = len(m.filtered) - 1
		}
		
	case "/":
		m.mode = ModeSearch
		m.searchInput = m.searchQuery
		
	case "enter":
		if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
			file := m.filtered[m.cursor]
			
			// Load metadata if not already loaded
			if file.IsTask() {
				task, ok := m.taskMetadata[file.Path]
				if !ok {
					if t, err := denote.ParseTaskFile(file.Path); err == nil {
						m.taskMetadata[file.Path] = t
						task = t
						ok = true
					}
				}
				if ok {
					m.mode = ModeTaskView
					m.viewingTask = task
					m.viewingProject = nil
					m.viewingFile = &file
					m.editingField = ""
					m.editBuffer = ""
				}
			} else if file.IsProject() {
				project, ok := m.projectMetadata[file.Path]
				if !ok {
					if p, err := denote.ParseProjectFile(file.Path); err == nil {
						m.projectMetadata[file.Path] = p
						project = p
						ok = true
					}
				}
				if ok {
					m.mode = ModeProjectView
					m.viewingTask = nil
					m.viewingProject = project
					m.viewingFile = &file
					m.editingField = ""
					m.editBuffer = ""
					m.projectViewTab = 0 // Start on overview tab
					m.loadProjectTasks() // Load tasks assigned to this project
				}
			} else {
				// Fall back to preview for non-task files
				m.mode = ModePreview
				m.previewFile = &file
			}
		}
		
	case "?":
		m.mode = ModeHelp
		
	// Task-specific keys
	case "1", "2", "3":
		// Set priority
		priority := "p" + msg.String()
		if err := m.updateTaskPriority(priority); err != nil {
			m.statusMsg = fmt.Sprintf("Error: %v", err)
		}
		
	case "e":
		// Edit task in external editor
		if m.config.Editor != "" && m.cursor < len(m.filtered) {
			file := m.filtered[m.cursor]
			return m, m.editFile(file.Path)
		} else {
			m.statusMsg = "No editor configured"
		}
		
	case "u":
		// Update task metadata - same as enter
		if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
			file := m.filtered[m.cursor]
			
			// Load metadata if not already loaded
			if file.IsTask() {
				task, ok := m.taskMetadata[file.Path]
				if !ok {
					if t, err := denote.ParseTaskFile(file.Path); err == nil {
						m.taskMetadata[file.Path] = t
						task = t
						ok = true
					}
				}
				if ok {
					m.mode = ModeTaskView
					m.viewingTask = task
					m.viewingProject = nil
					m.viewingFile = &file
					m.editingField = ""
					m.editBuffer = ""
				}
			} else if file.IsProject() {
				project, ok := m.projectMetadata[file.Path]
				if !ok {
					if p, err := denote.ParseProjectFile(file.Path); err == nil {
						m.projectMetadata[file.Path] = p
						project = p
						ok = true
					}
				}
				if ok {
					m.mode = ModeProjectView
					m.viewingTask = nil
					m.viewingProject = project
					m.viewingFile = &file
					m.editingField = ""
					m.editBuffer = ""
					m.projectViewTab = 0 // Start on overview tab
					m.loadProjectTasks() // Load tasks assigned to this project
				}
			}
		}
		
	case "c":
		// Create new task
		m.mode = ModeCreate
		m.createTitle = ""
		m.createTags = ""
		
	case "f":
		// Filter menu
		m.mode = ModeFilterMenu
		
	case "s":
		// State change menu - only for tasks, not projects
		if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
			file := m.filtered[m.cursor]
			if file.IsTask() {
				m.mode = ModeStateMenu
			} else {
				m.statusMsg = "State change only available for tasks"
			}
		}
		
	case "S":
		// Sort mode (uppercase S since lowercase is now for state)
		m.mode = ModeSort
		
	case "x", "delete":
		// Delete task confirmation
		if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
			file := m.filtered[m.cursor]
			if file.IsTask() || file.IsProject() {
				m.mode = ModeConfirmDelete
			}
		}
		
	case "r":
		// Toggle reverse sort
		m.reverseSort = !m.reverseSort
		m.sortFiles()
		
	case "p":
		// Toggle project filter
		m.projectFilter = !m.projectFilter
		if m.projectFilter {
			m.statusMsg = "Showing projects only"
			// Clear state filter when switching to projects
			m.stateFilter = ""
		} else {
			m.statusMsg = "Showing all tasks"
			// Restore active filter when going back to tasks
			m.stateFilter = "active"
		}
		m.cursor = 0
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
		
	case "t":
		// Toggle back to Notes mode
		m.viewMode = ViewModeNotes
		m.statusMsg = "Notes Mode"
		m.cursor = 0
		// Use configured defaults for notes
		m.sortBy = m.config.Notes.SortBy
		if m.sortBy == "modified" || m.sortBy == "created" {
			m.sortBy = "date"
		}
		if m.sortBy == "" {
			m.sortBy = "date"
		}
		m.reverseSort = m.config.Notes.SortOrder == "reverse"
		// Clear task-specific filters when leaving task mode
		m.stateFilter = ""
		m.projectFilter = false
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
		
	default:
		m.lastKey = ""
	}
	
	return m, nil
}

func (m Model) handleAreaFilterKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		m.mode = ModeNormal
		m.searchInput = ""
		
	case "enter":
		m.areaFilter = m.searchInput
		m.mode = ModeNormal
		m.searchInput = ""
		if m.areaFilter != "" {
			m.statusMsg = fmt.Sprintf("Filtering by area: %s", m.areaFilter)
		} else {
			m.statusMsg = "Area filter cleared"
		}
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
		
	case "backspace":
		if len(m.searchInput) > 0 {
			m.searchInput = m.searchInput[:len(m.searchInput)-1]
		}
		
	default:
		if len(msg.String()) == 1 {
			m.searchInput += msg.String()
		}
	}
	
	return m, nil
}

func (m Model) handleSortKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c", "q":
		m.mode = ModeNormal
		
	case "d":
		// Sort by due date
		m.sortBy = "due"
		m.mode = ModeNormal
		m.sortFiles()
		m.loadVisibleMetadata()
		m.statusMsg = "Sorted by due date"
		
	case "p":
		// Sort by priority
		m.sortBy = "priority"
		m.mode = ModeNormal
		m.sortFiles()
		m.loadVisibleMetadata()
		m.statusMsg = "Sorted by priority"
		
	case "e":
		// Sort by estimate
		m.sortBy = "estimate"
		m.mode = ModeNormal
		m.sortFiles()
		m.loadVisibleMetadata()
		m.statusMsg = "Sorted by estimate"
		
	case "t":
		// Sort by title
		m.sortBy = "title"
		m.mode = ModeNormal
		m.sortFiles()
		m.loadVisibleMetadata()
		m.statusMsg = "Sorted by title"
		
	case "c":
		// Sort by created date (default date sort)
		m.sortBy = "date"
		m.mode = ModeNormal
		m.sortFiles()
		m.loadVisibleMetadata()
		m.statusMsg = "Sorted by created date"
		
	case "r":
		// Toggle reverse sort
		m.reverseSort = !m.reverseSort
		m.mode = ModeNormal
		m.sortFiles()
		m.loadVisibleMetadata()
		if m.reverseSort {
			m.statusMsg = fmt.Sprintf("Sorted by %s (descending)", m.sortBy)
		} else {
			m.statusMsg = fmt.Sprintf("Sorted by %s (ascending)", m.sortBy)
		}
	}
	
	return m, nil
}

func (m Model) handleStateMenuKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Determine which mode to return to
	returnMode := ModeNormal
	if m.projectViewTab == 1 && m.viewingProject != nil {
		returnMode = ModeProjectView
	}
	
	switch msg.String() {
	case "esc", "ctrl+c", "q":
		m.mode = returnMode
		
	case "o":
		// Open
		var err error
		if returnMode == ModeProjectView {
			err = m.updateProjectTaskStatus(denote.TaskStatusOpen)
		} else {
			err = m.updateCurrentTaskStatus(denote.TaskStatusOpen)
		}
		if err != nil {
			m.statusMsg = fmt.Sprintf("Error: %v", err)
		} else {
			m.statusMsg = "Task status changed to open"
		}
		m.mode = returnMode
		
	case "p":
		// Paused
		var err error
		if returnMode == ModeProjectView {
			err = m.updateProjectTaskStatus(denote.TaskStatusPaused)
		} else {
			err = m.updateCurrentTaskStatus(denote.TaskStatusPaused)
		}
		if err != nil {
			m.statusMsg = fmt.Sprintf("Error: %v", err)
		} else {
			m.statusMsg = "Task status changed to paused"
		}
		m.mode = returnMode
		
	case "d":
		// Done
		var err error
		if returnMode == ModeProjectView {
			err = m.updateProjectTaskStatus(denote.TaskStatusDone)
		} else {
			err = m.updateCurrentTaskStatus(denote.TaskStatusDone)
		}
		if err != nil {
			m.statusMsg = fmt.Sprintf("Error: %v", err)
		} else {
			m.statusMsg = "Task status changed to done"
		}
		m.mode = returnMode
		
	case "e":
		// Delegated
		var err error
		if returnMode == ModeProjectView {
			err = m.updateProjectTaskStatus(denote.TaskStatusDelegated)
		} else {
			err = m.updateCurrentTaskStatus(denote.TaskStatusDelegated)
		}
		if err != nil {
			m.statusMsg = fmt.Sprintf("Error: %v", err)
		} else {
			m.statusMsg = "Task status changed to delegated"
		}
		m.mode = returnMode
		
	case "r":
		// Dropped
		var err error
		if returnMode == ModeProjectView {
			err = m.updateProjectTaskStatus(denote.TaskStatusDropped)
		} else {
			err = m.updateCurrentTaskStatus(denote.TaskStatusDropped)
		}
		if err != nil {
			m.statusMsg = fmt.Sprintf("Error: %v", err)
		} else {
			m.statusMsg = "Task status changed to dropped"
		}
		m.mode = returnMode
	}
	
	return m, nil
}

func (m Model) handleConfirmDeleteKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Determine which mode to return to
	returnMode := ModeNormal
	if m.viewingProject != nil {
		returnMode = ModeProjectView
	}
	
	switch msg.String() {
	case "y", "Y":
		// Handle project deletion specially
		if m.projectViewTab == 0 && m.viewingProject != nil {
			// Delete the project and clear project_id from affected tasks
			projectPath := m.viewingFile.Path
			projectTitle := m.viewingProject.ProjectMetadata.Title
			
			// First, clear project_id from all affected tasks
			for _, task := range m.affectedTasks {
				if err := m.clearProjectFromTask(task.File.Path); err != nil {
					m.statusMsg = fmt.Sprintf("Error updating task: %v", err)
					m.mode = returnMode
					return m, nil
				}
			}
			
			// Then delete the project file
			if err := m.deleteFile(projectPath); err != nil {
				m.statusMsg = fmt.Sprintf("Error deleting project: %v", err)
			} else {
				m.statusMsg = fmt.Sprintf("Deleted project: %s", projectTitle)
				if len(m.affectedTasks) > 0 {
					m.statusMsg += fmt.Sprintf(" (cleared from %d tasks)", len(m.affectedTasks))
				}
				// Go back to main task list
				m.mode = ModeNormal
				m.viewingProject = nil
				m.viewingFile = nil
				m.affectedTasks = nil
				// Rescan files
				m.scanFiles()
				return m, nil
			}
		} else if m.projectViewTab == 1 && m.viewingProject != nil {
			// Delete task from project view
			if m.projectTasksCursor < len(m.projectTasks) {
				task := m.projectTasks[m.projectTasksCursor]
				filePath := task.File.Path
				fileTitle := task.TaskMetadata.Title
				
				if err := m.deleteFile(filePath); err != nil {
					m.statusMsg = fmt.Sprintf("Error deleting: %v", err)
				} else {
					m.statusMsg = fmt.Sprintf("Deleted: %s", fileTitle)
					// Rescan files after deletion
					m.scanFiles()
					// Reload project tasks
					m.loadProjectTasks()
				}
			}
		} else {
			// Delete from normal view
			if m.cursor < len(m.filtered) {
				file := m.filtered[m.cursor]
				filePath := file.Path
				fileTitle := file.Title
				
				if err := m.deleteFile(filePath); err != nil {
					m.statusMsg = fmt.Sprintf("Error deleting: %v", err)
				} else {
					m.statusMsg = fmt.Sprintf("Deleted: %s", fileTitle)
					// Rescan files after deletion
					m.scanFiles()
				}
			}
		}
		m.mode = returnMode
		
	case "n", "N", "esc", "ctrl+c":
		// Cancel delete
		m.mode = returnMode
		m.statusMsg = "Delete cancelled"
		// Clear affected tasks list
		m.affectedTasks = nil
	}
	
	return m, nil
}

func (m Model) handleFilterMenuKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c", "q":
		m.mode = ModeNormal
		
	case "a":
		// Area filter
		m.mode = ModeAreaFilter
		m.searchInput = m.areaFilter
		
	case "p":
		// Priority filter
		m.mode = ModePriorityFilter
		
	case "s":
		// State filter  
		m.mode = ModeStateFilter
		
	case "d":
		// Soon filter toggle
		m.soonFilter = !m.soonFilter
		m.mode = ModeNormal
		if m.soonFilter {
			m.statusMsg = fmt.Sprintf("Showing tasks due within %d days", m.config.SoonHorizon)
		} else {
			m.statusMsg = "Soon filter disabled"
		}
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
		
	case "c":
		// Clear all filters
		m.areaFilter = ""
		m.priorityFilter = ""
		m.stateFilter = ""
		m.soonFilter = false
		m.mode = ModeNormal
		m.statusMsg = "All filters cleared"
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
	}
	
	return m, nil
}

func (m Model) handlePriorityFilterKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		m.mode = ModeNormal
		
	case "1":
		m.priorityFilter = "p1"
		m.mode = ModeNormal
		m.statusMsg = "Filtering by priority: p1"
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
		
	case "2":
		m.priorityFilter = "p2"
		m.mode = ModeNormal
		m.statusMsg = "Filtering by priority: p2"
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
		
	case "3":
		m.priorityFilter = "p3"
		m.mode = ModeNormal
		m.statusMsg = "Filtering by priority: p3"
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
		
	case "c", "x":
		// Clear priority filter
		m.priorityFilter = ""
		m.mode = ModeNormal
		m.statusMsg = "Priority filter cleared"
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
	}
	
	return m, nil
}

func (m Model) handleStateFilterKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		m.mode = ModeNormal
		
	case "a":
		// Active (open + delegated)
		m.stateFilter = "active"
		m.mode = ModeNormal
		m.statusMsg = "Filtering by state: active (open + delegated)"
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
		
	case "o":
		// Open
		m.stateFilter = denote.TaskStatusOpen
		m.mode = ModeNormal
		m.statusMsg = "Filtering by state: open"
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
		
	case "p":
		// Paused
		m.stateFilter = denote.TaskStatusPaused
		m.mode = ModeNormal
		m.statusMsg = "Filtering by state: paused"
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
		
	case "d":
		// Done
		m.stateFilter = denote.TaskStatusDone
		m.mode = ModeNormal
		m.statusMsg = "Filtering by state: done"
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
		
	case "e":
		// Delegated
		m.stateFilter = denote.TaskStatusDelegated
		m.mode = ModeNormal
		m.statusMsg = "Filtering by state: delegated"
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
		
	case "r":
		// Dropped
		m.stateFilter = denote.TaskStatusDropped
		m.mode = ModeNormal
		m.statusMsg = "Filtering by state: dropped"
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
		
	case "c", "x":
		// Clear state filter
		m.stateFilter = ""
		m.mode = ModeNormal
		m.statusMsg = "State filter cleared"
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
	}
	
	return m, nil
}