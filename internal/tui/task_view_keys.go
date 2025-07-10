package tui

import (
	"fmt"
	"strings"
	
	"github.com/charmbracelet/bubbletea"
	"github.com/pdxmph/denote-tasks/internal/denote"
)

func (m Model) handleTaskViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If we're editing a field, handle input
	if m.editingField != "" {
		switch msg.String() {
		case "esc":
			m.editingField = ""
			m.editBuffer = ""
			
		case "enter":
			// Save the field
			if m.viewingTask != nil {
				switch m.editingField {
				case "priority":
					if m.editBuffer == "" || m.editBuffer == "0" {
						// Clear priority
						if err := m.updateTaskField("priority", ""); err != nil {
							m.statusMsg = fmt.Sprintf("Error: %v", err)
						} else {
							m.statusMsg = "Priority removed"
						}
					} else if m.editBuffer == "1" || m.editBuffer == "2" || m.editBuffer == "3" {
						if err := m.updateTaskField("priority", "p"+m.editBuffer); err != nil {
							m.statusMsg = fmt.Sprintf("Error: %v", err)
						} else {
							m.statusMsg = fmt.Sprintf("Priority set to p%s", m.editBuffer)
						}
					} else {
						m.statusMsg = "Priority must be 0 (clear), 1, 2, or 3"
					}
				case "status":
					if err := m.updateTaskField("status", m.editBuffer); err != nil {
						m.statusMsg = fmt.Sprintf("Error: %v", err)
					}
				case "due":
					if err := m.updateTaskField("due_date", m.editBuffer); err != nil {
						m.statusMsg = fmt.Sprintf("Error: %v", err)
					}
				case "area":
					if err := m.updateTaskField("area", m.editBuffer); err != nil {
						m.statusMsg = fmt.Sprintf("Error: %v", err)
					}
				case "estimate":
					if err := m.updateTaskField("estimate", m.editBuffer); err != nil {
						m.statusMsg = fmt.Sprintf("Error: %v", err)
					}
				case "tags":
					if err := m.updateTaskField("tags", m.editBuffer); err != nil {
						m.statusMsg = fmt.Sprintf("Error: %v", err)
					}
				}
			} else if m.viewingProject != nil {
				// Handle project updates
				switch m.editingField {
				case "priority":
					if m.editBuffer == "" || m.editBuffer == "0" {
						// Clear priority
						if err := m.updateProjectField("priority", ""); err != nil {
							m.statusMsg = fmt.Sprintf("Error: %v", err)
						} else {
							m.statusMsg = "Priority removed"
						}
					} else if m.editBuffer == "1" || m.editBuffer == "2" || m.editBuffer == "3" {
						if err := m.updateProjectField("priority", "p"+m.editBuffer); err != nil {
							m.statusMsg = fmt.Sprintf("Error: %v", err)
						} else {
							m.statusMsg = fmt.Sprintf("Priority set to p%s", m.editBuffer)
						}
					} else {
						m.statusMsg = "Priority must be 0 (clear), 1, 2, or 3"
					}
				case "status":
					if err := m.updateProjectField("status", m.editBuffer); err != nil {
						m.statusMsg = fmt.Sprintf("Error: %v", err)
					}
				case "due":
					if err := m.updateProjectField("due_date", m.editBuffer); err != nil {
						m.statusMsg = fmt.Sprintf("Error: %v", err)
					}
				case "area":
					if err := m.updateProjectField("area", m.editBuffer); err != nil {
						m.statusMsg = fmt.Sprintf("Error: %v", err)
					}
				case "tags":
					if err := m.updateProjectField("tags", m.editBuffer); err != nil {
						m.statusMsg = fmt.Sprintf("Error: %v", err)
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
	
	// Normal task view navigation
	switch msg.String() {
	case "q", "esc":
		if m.returnToProject && m.viewingProject != nil {
			// Return to project view
			m.mode = ModeProjectView
			m.viewingTask = nil
			m.viewingFile = &m.viewingProject.File
			m.returnToProject = false
			// Re-sort project tasks in case metadata changed
			m.loadProjectTasks()
		} else {
			// Return to normal task list
			m.mode = ModeNormal
			m.viewingTask = nil
			m.viewingProject = nil
			m.viewingFile = nil
			// Re-sort the list in case metadata changed
			m.applyFilters()
			m.sortFiles()
			m.loadVisibleMetadata()
		}
		
	case "e":
		// Edit in external editor
		if m.config.Editor != "" && m.viewingFile != nil {
			return m, m.editFile(m.viewingFile.Path)
		}
		
	case "l":
		// Log entry - only for tasks
		if m.viewingTask != nil && m.viewingFile != nil {
			m.mode = ModeLogEntry
			m.logInput = ""
			m.loggingFile = m.viewingFile
		}
		
	// Field edit hotkeys
	case "p":
		m.editingField = "priority"
		m.editBuffer = ""
		m.statusMsg = "Enter priority (1/2/3):"
		
	case "s":
		m.editingField = "status"
		m.editBuffer = ""
		if m.viewingTask != nil {
			m.statusMsg = "Enter status (open/done/paused/delegated/dropped):"
		} else {
			m.statusMsg = "Enter status (active/completed/paused/cancelled):"
		}
		
	case "d":
		m.editingField = "due"
		m.editBuffer = ""
		m.statusMsg = "Enter due date (e.g. 2d, 1w, friday, jan 15, 2024-01-15):"
		
	case "a":
		m.editingField = "area"
		m.editBuffer = ""
		m.statusMsg = "Enter area (work/personal/etc):"
		
	case "t":
		if m.viewingTask != nil {
			m.editingField = "estimate"
			m.editBuffer = ""
			m.statusMsg = "Enter time estimate (1/2/3/5/8/13):"
		}
		
	case "j":
		// Project selection - only for tasks
		if m.viewingTask != nil {
			// Load projects and switch to selection mode
			m.loadProjectsForSelection()
			if len(m.projectSelectList) > 0 {
				m.projectSelectFor = "update"
				m.projectSelectTask = m.viewingTask
				m.mode = ModeProjectSelect
			} else {
				m.statusMsg = "No projects found"
			}
		}
		
	case "g":
		m.editingField = "tags"
		// Pre-fill with current tags
		if m.viewingTask != nil && len(m.viewingTask.TaskMetadata.Tags) > 0 {
			m.editBuffer = strings.Join(m.viewingTask.TaskMetadata.Tags, " ")
		} else if m.viewingProject != nil && len(m.viewingProject.ProjectMetadata.Tags) > 0 {
			m.editBuffer = strings.Join(m.viewingProject.ProjectMetadata.Tags, " ")
		} else {
			m.editBuffer = ""
		}
		m.statusMsg = "Enter tags (space-separated):"
		
	case "r":
		// Rename file to match current metadata tags
		if m.viewingTask != nil {
			// Build tag list including 'task' tag
			allTags := []string{"task"}
			for _, tag := range m.viewingTask.TaskMetadata.Tags {
				if tag != "task" {
					allTags = append(allTags, tag)
				}
			}
			
			// Rename file
			oldPath := m.viewingFile.Path
			newPath, err := denote.RenameFileForTags(oldPath, allTags)
			if err != nil {
				m.statusMsg = fmt.Sprintf("Error renaming: %v", err)
			} else if newPath != oldPath {
				// Update references
				m.viewingFile.Path = newPath
				// Trigger a rescan to update the file list
				m.scanFiles()
				m.statusMsg = "File renamed to match tags"
			} else {
				m.statusMsg = "Filename already matches tags"
			}
		} else if m.viewingProject != nil {
			// Build tag list including 'project' tag
			allTags := []string{"project"}
			for _, tag := range m.viewingProject.ProjectMetadata.Tags {
				if tag != "project" {
					allTags = append(allTags, tag)
				}
			}
			
			// Rename file
			oldPath := m.viewingFile.Path
			newPath, err := denote.RenameFileForTags(oldPath, allTags)
			if err != nil {
				m.statusMsg = fmt.Sprintf("Error renaming: %v", err)
			} else if newPath != oldPath {
				// Update references
				m.viewingFile.Path = newPath
				if project, ok := m.projectMetadata[oldPath]; ok {
					delete(m.projectMetadata, oldPath)
					m.projectMetadata[newPath] = project
				}
				// Trigger a rescan to update the file list
				m.scanFiles()
				m.statusMsg = "File renamed to match tags"
			} else {
				m.statusMsg = "Filename already matches tags"
			}
		}
		
	case "?":
		m.mode = ModeHelp
	}
	
	return m, nil
}