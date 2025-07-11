package tui

import (
	"fmt"
	"os"
	"strings"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pdxmph/denote-tasks/internal/denote"
)

func (m Model) handleProjectViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle editing mode first - EXACTLY like task view
	if m.editingField != "" {
		switch msg.String() {
		case "esc":
			m.editingField = ""
			m.editBuffer = ""
			
		case "enter":
			// Handle project field updates - map single letters to field names
			fieldMap := map[string]string{
				"t": "title",
				"p": "priority",
				"s": "status",
				"d": "due_date",
				"a": "area",
				"g": "tags",
			}
			
			if fieldName, ok := fieldMap[m.editingField]; ok {
				var updateErr error
				var updateValue string
				
				switch m.editingField {
				case "p":
					if m.editBuffer == "1" || m.editBuffer == "2" || m.editBuffer == "3" {
						updateValue = "p" + m.editBuffer
					} else {
						m.statusMsg = "Priority must be 1, 2, or 3"
						m.editingField = ""
						m.editBuffer = ""
						return m, nil
					}
				default:
					updateValue = m.editBuffer
				}
				
				updateErr = m.updateProjectField(fieldName, updateValue)
				
				if updateErr != nil {
					m.statusMsg = fmt.Sprintf(ErrorFormat, updateErr)
				} else {
					m.statusMsg = fmt.Sprintf("Updated %s", fieldName)
					
					// Force reload the project from disk to ensure we show the updated data
					if project, err := denote.ParseProjectFile(m.viewingFile.Path); err == nil {
						m.viewingProject = project
						m.projectMetadata[m.viewingFile.Path] = project
						// Update the file title if it changed
						if project.ProjectMetadata.Title != "" {
							m.viewingFile.Title = project.ProjectMetadata.Title
						}
					}
					
					// Re-sort if we updated a field that could affect order
					if m.editingField == "d" || m.editingField == "p" {
						m.applyFilters()
						m.sortFiles()
						m.loadVisibleMetadata()
					}
				}
			}
			m.editingField = ""
			m.editBuffer = ""
			
		case "backspace":
			if len(m.editBuffer) > 0 {
				m.editBuffer = m.editBuffer[:len(m.editBuffer)-1]
			}
			
		default:
			if len(msg.String()) == 1 {
				m.editBuffer += msg.String()
			}
		}
		return m, nil
	}
	
	// Normal navigation when not editing
	switch msg.String() {
	case "q", "esc":
		m.mode = ModeNormal
		m.viewingProject = nil
		m.viewingFile = nil
		m.projectTasks = nil
		m.projectTasksCursor = 0
		m.projectViewTab = 0
		m.affectedTasks = nil
		// Re-sort the list in case metadata changed
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
		
	case "tab":
		// Switch between tabs
		if m.projectViewTab == 0 {
			m.projectViewTab = 1
		} else {
			m.projectViewTab = 0
		}
		
	case "e":
		// Edit project file in external editor
		if m.config.Editor != "" && m.viewingFile != nil {
			return m, m.editFile(m.viewingFile.Path)
		} else {
			m.statusMsg = "No editor configured"
		}
		
	// Field edit hotkeys - work on overview tab
	case "t":
		if m.projectViewTab == 0 {
			m.editingField = "t"  // Title
			m.editBuffer = m.viewingProject.ProjectMetadata.Title
			m.statusMsg = "Enter title:"
		}
		
	case "p":
		if m.projectViewTab == 0 {
			m.editingField = "p"  // Use single letter like renderField expects
			m.editBuffer = strings.TrimPrefix(m.viewingProject.ProjectMetadata.Priority, "p")
			m.statusMsg = "Enter priority (1/2/3):"
		}
		
	case "s":
		if m.projectViewTab == 0 {
			m.editingField = "s"  // Use single letter
			m.editBuffer = m.viewingProject.ProjectMetadata.Status
			m.statusMsg = "Enter status (active/completed/paused/cancelled):"
		} else if m.projectViewTab == 0 && len(m.projectTasks) > 0 {
			// In tasks tab, 's' opens state menu
			m.mode = ModeStateMenu
			return m, nil
		}
		
	case "d":
		if m.projectViewTab == 0 {
			m.editingField = "d"  // Use single letter
			m.editBuffer = m.viewingProject.ProjectMetadata.DueDate
			m.statusMsg = "Enter due date (YYYY-MM-DD or relative: 1d, 1w, tomorrow):"
		}
		
	case "a":
		if m.projectViewTab == 0 {
			m.editingField = "a"  // Use single letter
			m.editBuffer = m.viewingProject.ProjectMetadata.Area
			m.statusMsg = "Enter area:"
		}
		
	case "g":
		if m.projectViewTab == 0 {
			m.editingField = "g"  // Use single letter
			if len(m.viewingProject.ProjectMetadata.Tags) > 0 {
				m.editBuffer = strings.Join(m.viewingProject.ProjectMetadata.Tags, " ")
			} else {
				m.editBuffer = ""
			}
			m.statusMsg = "Enter tags (space-separated):"
		}
		
	case "n":
		// Create new task with this project pre-selected
		m.mode = ModeCreate
		m.resetCreateFields()
		// Pre-fill the project ID from the current project
		if m.viewingProject != nil {
			m.createProject = m.viewingProject.File.ID
		}
		// Pre-fill area if the project has one
		if m.viewingProject != nil && m.viewingProject.ProjectMetadata.Area != "" {
			m.areaFilter = m.viewingProject.ProjectMetadata.Area
		}
		m.creatingFromProject = true
		m.statusMsg = "Creating new task for project"
		return m, nil
		
	// Keys for task navigation (on main tab)
	case "j", "down", "k", "up", "G", "ctrl+d", "ctrl+u":
		if m.projectViewTab == 0 && len(m.projectTasks) > 0 {
			// Use navigation handler for task list in project view
			nav := NewNavigationHandler(len(m.projectTasks), false)
			nav.cursor = m.projectTasksCursor
			m.projectTasksCursor = nav.HandleKey(msg.String())
		}
		
	case "enter":
		if m.projectViewTab == 0 && len(m.projectTasks) > 0 && m.projectTasksCursor < len(m.projectTasks) {
			// Open the selected task
			task := m.projectTasks[m.projectTasksCursor]
			m.mode = ModeTaskView
			m.viewingTask = &task
			m.viewingFile = &task.File
			m.editingField = ""
			m.editBuffer = ""
			m.returnToProject = true // Remember to return to project view
			// Keep the project reference!
		}
		
	case "0":
		if m.projectViewTab == 0 && len(m.projectTasks) > 0 {
			// Clear priority on selected task
			task := &m.projectTasks[m.projectTasksCursor]
			if err := m.updateTaskPriorityFromProject(task, ""); err != nil {
				m.statusMsg = fmt.Sprintf(ErrorFormat, err)
			} else {
				m.statusMsg = "Task priority cleared"
			}
		}
		
	case "1", "2", "3":
		if m.projectViewTab == 0 && len(m.projectTasks) > 0 {
			// Set priority on selected task
			task := &m.projectTasks[m.projectTasksCursor]
			priority := "p" + msg.String()
			if err := m.updateTaskPriorityFromProject(task, priority); err != nil {
				m.statusMsg = fmt.Sprintf(ErrorFormat, err)
			} else {
				m.statusMsg = fmt.Sprintf("Task priority updated to %s", priority)
			}
		}
		
	case "x":
		// On main tab (0): delete selected task if cursor is on a task
		if m.projectViewTab == 0 && len(m.projectTasks) > 0 && m.projectTasksCursor < len(m.projectTasks) {
			// Deleting a task
			m.mode = ModeConfirmDelete
			return m, nil
		}
		
	case "X":
		// Capital X to delete the project itself (only on main tab)
		if m.projectViewTab == 0 {
			// Deleting the project itself - find affected tasks first
			m.findTasksAffectedByProjectDeletion()
			m.mode = ModeConfirmDelete
			return m, nil
		}
		
	// Sorting keys (same as task mode - uppercase S since lowercase s is for status)
	case "S":
		// Open sort menu
		m.mode = ModeSort
		return m, nil
	}
	
	return m, nil
}

// updateTaskPriorityFromProject updates a task priority from the project view
func (m *Model) updateTaskPriorityFromProject(task *denote.Task, priority string) error {
	// Read the file content
	content, err := os.ReadFile(task.File.Path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	
	// Parse existing frontmatter
	fm, err := denote.ParseFrontmatterFile(content)
	if err != nil {
		return fmt.Errorf("failed to parse frontmatter: %w", err)
	}
	
	// Update the metadata
	if taskMeta, ok := fm.Metadata.(denote.TaskMetadata); ok {
		taskMeta.Priority = priority
		
		// Write updated content
		newContent, err := denote.WriteFrontmatterFile(taskMeta, fm.Content)
		if err != nil {
			return fmt.Errorf("failed to write frontmatter: %w", err)
		}
		
		// Write to file
		if err := os.WriteFile(task.File.Path, newContent, 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		
		// Update our in-memory copy
		task.TaskMetadata = taskMeta
		// Also update in the metadata cache
		if cachedTask, ok := m.taskMetadata[task.File.Path]; ok {
			cachedTask.TaskMetadata.Priority = priority
		}
	}
	
	return nil
}

