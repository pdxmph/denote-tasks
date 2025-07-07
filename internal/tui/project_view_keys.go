package tui

import (
	"fmt"
	"os"
	"strings"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pdxmph/denote-tasks/internal/denote"
)

func (m Model) handleProjectViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.mode = ModeNormal
		m.viewingProject = nil
		m.viewingFile = nil
		m.projectTasks = nil
		m.projectTasksCursor = 0
		m.projectViewTab = 0
		
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
		
	// Keys for overview tab
	case "p":
		if m.projectViewTab == 0 && m.editingField != "p" {
			m.editingField = "p"
			m.editBuffer = m.viewingProject.ProjectMetadata.Priority
			m.statusMsg = "Edit priority (p1/p2/p3): "
		}
		
	case "s":
		if m.projectViewTab == 0 && m.editingField != "s" {
			m.editingField = "s"
			m.editBuffer = m.viewingProject.ProjectMetadata.Status
			m.statusMsg = "Edit status (active/completed/paused/cancelled): "
		} else if m.projectViewTab == 1 && len(m.projectTasks) > 0 && m.editingField == "" {
			// In tasks tab, 's' opens state menu
			m.mode = ModeStateMenu
			return m, nil
		}
		
	case "d":
		if m.projectViewTab == 0 && m.editingField != "d" {
			m.editingField = "d"
			m.editBuffer = m.viewingProject.ProjectMetadata.DueDate
			m.statusMsg = "Edit due date (YYYY-MM-DD or relative: 1d, 1w, tomorrow): "
		}
		
	case "a":
		if m.projectViewTab == 0 && m.editingField != "a" {
			m.editingField = "a"
			m.editBuffer = m.viewingProject.ProjectMetadata.Area
			m.statusMsg = "Edit area: "
		}
		
	case "g":
		if m.projectViewTab == 0 && m.editingField != "g" {
			m.editingField = "g"
			if len(m.viewingProject.ProjectMetadata.Tags) > 0 {
				m.editBuffer = strings.Join(m.viewingProject.ProjectMetadata.Tags, " ")
			} else {
				m.editBuffer = ""
			}
			m.statusMsg = "Edit tags (space-separated): "
		}
		
	// Keys for tasks tab
	case "j", "down":
		if m.projectViewTab == 1 && m.projectTasksCursor < len(m.projectTasks)-1 {
			m.projectTasksCursor++
		}
		
	case "k", "up":
		if m.projectViewTab == 1 && m.projectTasksCursor > 0 {
			m.projectTasksCursor--
		}
		
	case "enter":
		if m.projectViewTab == 1 && len(m.projectTasks) > 0 && m.projectTasksCursor < len(m.projectTasks) {
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
		
	case "1", "2", "3":
		if m.projectViewTab == 1 && len(m.projectTasks) > 0 {
			// Set priority on selected task
			task := &m.projectTasks[m.projectTasksCursor]
			priority := "p" + msg.String()
			if err := m.updateTaskPriorityFromProject(task, priority); err != nil {
				m.statusMsg = fmt.Sprintf("Error: %v", err)
			} else {
				m.statusMsg = fmt.Sprintf("Task priority updated to %s", priority)
			}
		}
		
	case "x":
		// Delete task (in tasks tab)
		if m.projectViewTab == 1 && len(m.projectTasks) > 0 {
			m.mode = ModeConfirmDelete
			return m, nil
		}
	}
	
	// Handle editing mode
	if m.editingField != "" {
		switch msg.String() {
		case "esc":
			m.editingField = ""
			m.editBuffer = ""
			m.statusMsg = ""
			
		case "enter":
			if err := m.updateProjectFieldValue(m.editingField, m.editBuffer); err != nil {
				m.statusMsg = fmt.Sprintf("Error: %v", err)
			} else {
				m.statusMsg = "Updated"
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

// updateProjectFieldValue updates a specific field in the project metadata
func (m *Model) updateProjectFieldValue(field, value string) error {
	if m.viewingProject == nil || m.viewingFile == nil {
		return fmt.Errorf("no project selected")
	}
	
	// Read the file content
	content, err := os.ReadFile(m.viewingFile.Path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	
	// Parse existing frontmatter
	fm, err := denote.ParseFrontmatterFile(content)
	if err != nil {
		return fmt.Errorf("failed to parse frontmatter: %w", err)
	}
	
	// Update the specific field
	if projMeta, ok := fm.Metadata.(denote.ProjectMetadata); ok {
		switch field {
		case "p":
			projMeta.Priority = value
		case "s":
			projMeta.Status = value
		case "d":
			// Parse the date
			if value != "" {
				parsedDate, err := denote.ParseNaturalDate(value)
				if err != nil {
					return fmt.Errorf("invalid date: %w", err)
				}
				projMeta.DueDate = parsedDate
			} else {
				projMeta.DueDate = ""
			}
		case "a":
			projMeta.Area = value
		case "g":
			if value == "" {
				projMeta.Tags = []string{}
			} else {
				projMeta.Tags = strings.Fields(value)
			}
		default:
			return fmt.Errorf("unknown field: %s", field)
		}
		
		// Write updated content
		newContent, err := denote.WriteFrontmatterFile(projMeta, fm.Content)
		if err != nil {
			return fmt.Errorf("failed to write frontmatter: %w", err)
		}
		
		// Write to file
		if err := os.WriteFile(m.viewingFile.Path, newContent, 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		
		// Update our in-memory copy
		m.viewingProject.ProjectMetadata = projMeta
		
		// Update in cache
		if cached, ok := m.projectMetadata[m.viewingFile.Path]; ok {
			cached.ProjectMetadata = projMeta
		}
	}
	
	return nil
}