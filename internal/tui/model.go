package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pdxmph/denote-tasks/internal/config"
	"github.com/pdxmph/denote-tasks/internal/denote"
	"github.com/pdxmph/denote-tasks/internal/task"
)

type Model struct {
	// Config
	config *config.Config
	
	// Denote files
	files      []denote.File
	filtered   []denote.File
	cursor     int
	
	// Task/Project metadata (parallel arrays to files/filtered)
	taskMetadata    map[string]*denote.Task    // keyed by file path
	projectMetadata map[string]*denote.Project // keyed by file path
	
	// UI State
	width      int
	height     int
	mode       Mode
	viewMode   ViewMode
	sortBy     string
	reverseSort bool
	
	// Filters
	searchQuery    string
	searchInput    string
	areaFilter     string
	priorityFilter string
	stateFilter    string
	soonFilter     bool
	projectFilter  bool  // Filter to show only projects
	
	// Preview
	previewFile     *denote.File
	previewScroll   int
	previewMaxScroll int
	
	// Create mode
	createTitle    string
	createTags     string
	createPriority string
	createDue      string
	createEstimate string
	createProject  string
	createField    int // Which field is being edited in create mode
	
	// Task view mode
	viewingTask     *denote.Task
	viewingProject  *denote.Project
	viewingFile     *denote.File
	editingField    string // which field is being edited
	editBuffer      string // temporary edit buffer
	returnToProject bool   // whether to return to project view after task view
	
	// Project view mode
	projectViewTab     int // 0 = overview, 1 = tasks
	projectTasks       []denote.Task // tasks assigned to current project
	projectTasksCursor int
	affectedTasks      []denote.Task // tasks affected by project deletion
	
	// Display
	err        error
	statusMsg  string
	lastKey    string
	
	// Log entry mode
	logInput     string // Current log entry being typed
	loggingFile  *denote.File // File we're adding log to
	
	// Project selection mode
	projectSelectList   []*denote.Project
	projectSelectCursor int
	projectSelectFor    string // "create" or "update"
	projectSelectTask   *denote.Task // For update mode
}

type Mode int

const (
	ModeNormal Mode = iota
	ModeSearch
	ModeHelp
	ModePreview
	ModeCreate
	ModeCreateTags
	ModeTaskView
	ModeAreaFilter
	ModeSort
	ModeProjectView
	ModeStateMenu
	ModeConfirmDelete
	ModeFilterMenu
	ModePriorityFilter
	ModeStateFilter
	ModeLogEntry
	ModeProjectSelect
)

type ViewMode int

const (
	ViewModeNotes ViewMode = iota
	ViewModeTasks
)

// noteCreatedMsg is sent when a note is successfully created
type noteCreatedMsg struct {
	path string
}

// taskCreatedMsg is sent when a task is successfully created
type taskCreatedMsg struct {
	path string
}

// fileEditedMsg is sent when returning from external editor
type fileEditedMsg struct {
	path string
}

func NewModel(cfg *config.Config) (*Model, error) {
	// Use configured defaults for notes mode
	reverseSort := cfg.Notes.SortOrder == "reverse"
	sortBy := cfg.Notes.SortBy
	if sortBy == "" {
		sortBy = "created" // Default to created date
	}
	
	m := &Model{
		config:          cfg,
		mode:            ModeNormal,
		viewMode:        ViewModeNotes, // Start in Notes mode
		sortBy:          sortBy,
		reverseSort:     reverseSort,
		taskMetadata:    make(map[string]*denote.Task),
		projectMetadata: make(map[string]*denote.Project),
	}
	
	// Initial scan
	if err := m.scanFiles(); err != nil {
		return nil, err
	}
	
	// Load metadata for initial view
	m.loadVisibleMetadata()
	
	return m, nil
}

func (m *Model) scanFiles() error {
	scanner := denote.NewScanner(m.config.NotesDirectory)
	files, err := scanner.FindAllNotes()
	if err != nil {
		return err
	}
	
	m.files = files
	
	// Don't clear metadata caches - preserve existing loaded metadata
	if m.taskMetadata == nil {
		m.taskMetadata = make(map[string]*denote.Task)
	}
	if m.projectMetadata == nil {
		m.projectMetadata = make(map[string]*denote.Project)
	}
	
	// Remove metadata for files that no longer exist
	fileMap := make(map[string]bool)
	for _, f := range m.files {
		fileMap[f.Path] = true
	}
	
	for path := range m.taskMetadata {
		if !fileMap[path] {
			delete(m.taskMetadata, path)
		}
	}
	for path := range m.projectMetadata {
		if !fileMap[path] {
			delete(m.projectMetadata, path)
		}
	}
	
	m.applyFilters()
	m.sortFiles()
	
	// Load metadata for visible files after filtering
	m.loadVisibleMetadata()
	
	return nil
}

// loadVisibleMetadata loads metadata for currently visible files only
func (m *Model) loadVisibleMetadata() {
	// Calculate visible range
	visibleHeight := m.height - 6
	if visibleHeight < 1 {
		visibleHeight = 20
	}
	
	start := 0
	if m.cursor >= visibleHeight {
		start = m.cursor - visibleHeight + 1
	}
	
	end := start + visibleHeight
	if end > len(m.filtered) {
		end = len(m.filtered)
	}
	
	// Always load fresh metadata for visible files
	for i := start; i < end; i++ {
		file := &m.filtered[i]
		if file.IsTask() {
			if task, err := denote.ParseTaskFile(file.Path); err == nil {
				m.taskMetadata[file.Path] = task
				// Update the title with metadata title
				if task.TaskMetadata.Title != "" {
					file.Title = task.TaskMetadata.Title
				}
			}
		} else if file.IsProject() {
			if project, err := denote.ParseProjectFile(file.Path); err == nil {
				m.projectMetadata[file.Path] = project
				// Update the title with metadata title
				if project.ProjectMetadata.Title != "" {
					file.Title = project.ProjectMetadata.Title
				}
			}
		}
	}
}

func (m *Model) applyFilters() {
	filtered := make([]denote.File, 0, len(m.files))
	
	for _, f := range m.files {
		// Apply view mode filter first
		if m.viewMode == ViewModeTasks {
			// In Task mode, only show tasks and projects
			if !f.IsTask() && !f.IsProject() {
				continue
			}
			
			// Apply project filter if active
			if m.projectFilter {
				if !f.IsProject() {
					continue
				}
			}
		} else {
			// In Notes mode, exclude tasks and projects
			if f.IsTask() || f.IsProject() {
				continue
			}
		}
		
		// Apply search filter
		if m.searchQuery != "" {
			if strings.HasPrefix(m.searchQuery, "#") {
				// Tag search mode
				tagQuery := strings.TrimPrefix(m.searchQuery, "#")
				if tagQuery != "" && !f.MatchesTag(tagQuery) {
					continue
				}
			} else {
				// Regular search mode
				matches := f.MatchesSearch(m.searchQuery)
				
				// For tasks, also search in metadata
				if !matches && f.IsTask() {
					if task, ok := m.taskMetadata[f.Path]; ok {
						if m.taskMatchesSearch(task, m.searchQuery) {
							matches = true
						}
					} else {
						// Load metadata on demand
						if t, err := denote.ParseTaskFile(f.Path); err == nil {
							m.taskMetadata[f.Path] = t
							if m.taskMatchesSearch(t, m.searchQuery) {
								matches = true
							}
						}
					}
				}
				
				if !matches {
					continue
				}
			}
		}
		
		// Apply filters (only in task mode)
		if m.viewMode == ViewModeTasks {
			// Load metadata if needed for filtering
			var taskMeta *denote.Task
			var projectMeta *denote.Project
			
			if f.IsTask() {
				task, ok := m.taskMetadata[f.Path]
				if !ok {
					// Load metadata on demand
					if t, err := denote.ParseTaskFile(f.Path); err == nil {
						m.taskMetadata[f.Path] = t
						task = t
						ok = true
					}
				}
				if ok {
					taskMeta = task
				}
			} else if f.IsProject() {
				project, ok := m.projectMetadata[f.Path]
				if !ok {
					// Load metadata on demand
					if p, err := denote.ParseProjectFile(f.Path); err == nil {
						m.projectMetadata[f.Path] = p
						project = p
						ok = true
					}
				}
				if ok {
					projectMeta = project
				}
			}
			
			// Area filter
			if m.areaFilter != "" {
				if taskMeta != nil && !strings.EqualFold(taskMeta.Area, m.areaFilter) {
					continue
				}
				if projectMeta != nil && !strings.EqualFold(projectMeta.Area, m.areaFilter) {
					continue
				}
			}
			
			// Priority filter (tasks and projects)
			if m.priorityFilter != "" {
				if taskMeta != nil && taskMeta.Priority != m.priorityFilter {
					continue
				}
				if projectMeta != nil && projectMeta.Priority != m.priorityFilter {
					continue
				}
			}
			
			// State filter (tasks only)
			if m.stateFilter != "" && taskMeta != nil {
				if m.stateFilter == "active" {
					// Active means open or delegated
					if taskMeta.Status != denote.TaskStatusOpen && 
					   taskMeta.Status != denote.TaskStatusDelegated {
						continue
					}
				} else if taskMeta.Status != m.stateFilter {
					continue
				}
			}
			
			// Soon filter (tasks and projects with due dates)
			if m.soonFilter {
				isDueSoon := false
				if taskMeta != nil && taskMeta.DueDate != "" {
					isDueSoon = denote.IsDueSoon(taskMeta.DueDate, m.config.SoonHorizon)
				} else if projectMeta != nil && projectMeta.DueDate != "" {
					isDueSoon = denote.IsDueSoon(projectMeta.DueDate, m.config.SoonHorizon)
				}
				if !isDueSoon {
					continue
				}
			}
		}
		
		filtered = append(filtered, f)
	}
	
	m.filtered = filtered
	
	// Reset cursor if out of bounds
	if m.cursor >= len(m.filtered) && len(m.filtered) > 0 {
		m.cursor = len(m.filtered) - 1
	}
}

func (m *Model) sortFiles() {
	if m.viewMode == ViewModeTasks {
		// Use task-specific sorting for task mode
		denote.SortTaskFiles(m.filtered, m.sortBy, m.reverseSort, m.taskMetadata, m.projectMetadata)
	} else {
		// Use note-specific sorting for notes mode
		denote.SortFiles(m.filtered, m.sortBy, m.reverseSort)
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
		
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
		
	case noteCreatedMsg:
		// Rescan files after note creation
		m.scanFiles()
		m.statusMsg = "Note created: " + msg.path
		return m, nil
		
	case taskCreatedMsg:
		// Rescan files after task creation
		m.scanFiles()
		m.statusMsg = "Task created: " + msg.path + " (press 'e' to edit)"
		
		// Reset create fields
		m.resetCreateFields()
		
		// Try to position cursor on the newly created task
		for i, f := range m.filtered {
			if f.Path == msg.path {
				m.cursor = i
				break
			}
		}
		
		return m, nil
		
	case fileEditedMsg:
		// Check if file needs renaming after edit
		oldPath := msg.path
		var newPath string
		
		// Parse the filename to determine file type
		parser := denote.NewParser()
		file, err := parser.ParseFilename(filepath.Base(oldPath))
		if err != nil {
			// Not a Denote file, just rescan
			m.scanFiles()
			return m, nil
		}
		
		// Check if it's a task or project based on tags
		if file.HasTag("task") {
			if task, err := denote.ParseTaskFile(oldPath); err == nil {
				// Build tag list including 'task' tag
				allTags := []string{"task"}
				for _, tag := range task.TaskMetadata.Tags {
					if tag != "task" {
						allTags = append(allTags, tag)
					}
				}
				
				// Check if rename is needed
				if renamed, err := denote.RenameFileForTags(oldPath, allTags); err == nil {
					newPath = renamed
					if newPath != oldPath {
						// Update cached metadata with new path
						if cachedTask, ok := m.taskMetadata[oldPath]; ok {
							delete(m.taskMetadata, oldPath)
							m.taskMetadata[newPath] = cachedTask
						}
						// Update viewing file path if this is the file being viewed
						if m.viewingFile != nil && m.viewingFile.Path == oldPath {
							m.viewingFile.Path = newPath
						}
					}
				}
			}
		} else if file.HasTag("project") {
			if project, err := denote.ParseProjectFile(oldPath); err == nil {
				// Build tag list including 'project' tag
				allTags := []string{"project"}
				for _, tag := range project.ProjectMetadata.Tags {
					if tag != "project" {
						allTags = append(allTags, tag)
					}
				}
				
				// Check if rename is needed
				if renamed, err := denote.RenameFileForTags(oldPath, allTags); err == nil {
					newPath = renamed
					if newPath != oldPath {
						// Update cached metadata with new path
						if cachedProject, ok := m.projectMetadata[oldPath]; ok {
							delete(m.projectMetadata, oldPath)
							m.projectMetadata[newPath] = cachedProject
						}
						// Update viewing file path if this is the file being viewed
						if m.viewingFile != nil && m.viewingFile.Path == oldPath {
							m.viewingFile.Path = newPath
						}
					}
				}
			}
		}
		
		// Always rescan files after editing
		m.scanFiles()
		
		// Re-apply filters and sort to reflect any metadata changes
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
		
		if newPath != "" && newPath != oldPath {
			m.statusMsg = "File renamed to match updated tags"
		}
		
		return m, nil
		
	case error:
		m.err = msg
		return m, nil
	}
	
	return m, nil
}

func (m *Model) resetCreateFields() {
	m.createTitle = ""
	m.createTags = ""
	m.createPriority = ""
	m.createDue = ""
	m.createEstimate = ""
	m.createProject = ""
	m.createField = 0
}

func (m *Model) loadProjectsForSelection() {
	// Get all projects
	m.projectSelectList = make([]*denote.Project, 0)
	
	for _, file := range m.files {
		if file.IsProject() {
			if project, ok := m.projectMetadata[file.Path]; ok {
				m.projectSelectList = append(m.projectSelectList, project)
			}
		}
	}
	
	// Sort by title
	sort.Slice(m.projectSelectList, func(i, j int) bool {
		return m.projectSelectList[i].ProjectMetadata.Title < m.projectSelectList[j].ProjectMetadata.Title
	})
	
	m.projectSelectCursor = 0
}

func (m Model) createTask() tea.Cmd {
	return func() tea.Msg {
		// Parse tags
		tags := []string{}
		if m.createTags != "" {
			tags = strings.Fields(m.createTags)
		}
		
		// Create the task
		newTask, err := task.CreateTask(m.config.NotesDirectory, m.createTitle, "", tags, m.areaFilter)
		if err != nil {
			return err
		}
		
		// Update metadata if provided
		needsUpdate := false
		
		if m.createPriority != "" {
			newTask.TaskMetadata.Priority = m.createPriority
			needsUpdate = true
		}
		
		if m.createDue != "" {
			// Parse due date
			parsedDue, err := denote.ParseNaturalDate(m.createDue)
			if err == nil {
				newTask.TaskMetadata.DueDate = parsedDue
				needsUpdate = true
			}
		}
		
		if m.createProject != "" {
			newTask.TaskMetadata.ProjectID = m.createProject
			needsUpdate = true
		}
		
		if m.createEstimate != "" {
			// Parse estimate as integer
			if estimate, err := strconv.Atoi(m.createEstimate); err == nil {
				newTask.TaskMetadata.Estimate = estimate
				needsUpdate = true
			}
		}
		
		// Write updated metadata if needed
		if needsUpdate {
			if err := task.UpdateTaskFile(newTask.File.Path, newTask.TaskMetadata); err != nil {
				return err
			}
		}
		
		return taskCreatedMsg{path: newTask.File.Path}
	}
}

func (m Model) create() tea.Cmd {
	return func() tea.Msg {
		// Parse tags
		tags := []string{}
		if m.createTags != "" {
			tags = strings.Fields(m.createTags)
		}
		
		if m.viewMode == ViewModeTasks {
			// Use the new createTask for tasks (this path is for the old flow)
			task, err := task.CreateTask(m.config.NotesDirectory, m.createTitle, "", tags, m.areaFilter)
			if err != nil {
				return err
			}
			
			return taskCreatedMsg{path: task.Path}
		} else {
			// Create a note
			filepath, err := denote.CreateNote(m.config.NotesDirectory, m.createTitle, tags)
			if err != nil {
				return err
			}
			
			return noteCreatedMsg{path: filepath}
		}
	}
}

// updateTaskPriority updates the priority of the current task
func (m *Model) updateTaskPriority(priority string) error {
	if m.cursor >= len(m.filtered) {
		return fmt.Errorf("no task selected")
	}
	
	file := m.filtered[m.cursor]
	task, ok := m.taskMetadata[file.Path]
	if !ok {
		return fmt.Errorf("no task metadata found")
	}
	
	// Update the priority
	task.Priority = priority
	
	// Read the file content
	content, err := os.ReadFile(file.Path)
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
		if err := os.WriteFile(file.Path, newContent, 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		
		// Update our in-memory copy
		task.TaskMetadata = taskMeta
		m.statusMsg = fmt.Sprintf("Priority updated to %s", priority)
	}
	
	return nil
}

// updateTaskField updates a specific field in the task metadata
func (m *Model) updateTaskField(field, value string) error {
	if m.viewingTask == nil || m.viewingFile == nil {
		return fmt.Errorf("no task selected")
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
	
	// Update the metadata
	if taskMeta, ok := fm.Metadata.(denote.TaskMetadata); ok {
		switch field {
		case "priority":
			taskMeta.Priority = value
		case "status":
			taskMeta.Status = value
		case "due_date":
			// Parse natural language dates
			if value != "" {
				parsed, err := denote.ParseNaturalDate(value)
				if err != nil {
					return fmt.Errorf("invalid date: %s (try: 2d, 1w, friday, jan 15, 2024-01-15)", value)
				}
				taskMeta.DueDate = parsed
			} else {
				taskMeta.DueDate = ""
			}
		case "area":
			taskMeta.Area = value
		case "estimate":
			// Parse as int
			var est int
			fmt.Sscanf(value, "%d", &est)
			taskMeta.Estimate = est
		case "tags":
			// Split by spaces
			if value == "" {
				taskMeta.Tags = []string{}
			} else {
				taskMeta.Tags = strings.Fields(value)
			}
		}
		
		// Write updated content
		newContent, err := denote.WriteFrontmatterFile(taskMeta, fm.Content)
		if err != nil {
			return fmt.Errorf("failed to write frontmatter: %w", err)
		}
		
		// Check if we need to rename the file (for tag changes)
		oldPath := m.viewingFile.Path
		newPath := oldPath
		
		if field == "tags" {
			// Combine filename tags with metadata tags, excluding 'task'
			allTags := []string{"task"} // Always include task tag
			for _, tag := range taskMeta.Tags {
				if tag != "task" {
					allTags = append(allTags, tag)
				}
			}
			
			// Rename file to reflect new tags
			renamed, err := denote.RenameFileForTags(oldPath, allTags)
			if err != nil {
				return fmt.Errorf("failed to rename file: %w", err)
			}
			newPath = renamed
		}
		
		// Write to file (at potentially new path)
		if err := os.WriteFile(newPath, newContent, 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		
		// Update our in-memory copy
		m.viewingTask.TaskMetadata = taskMeta
		
		// Update path references if file was renamed
		if newPath != oldPath {
			// Update viewing file path
			m.viewingFile.Path = newPath
			
			// Update in main metadata map
			if task, ok := m.taskMetadata[oldPath]; ok {
				delete(m.taskMetadata, oldPath)
				m.taskMetadata[newPath] = task
				task.TaskMetadata = taskMeta
			}
			
			// Trigger a rescan to update the file list
			m.scanFiles()
		} else {
			// Just update metadata if no rename
			if task, ok := m.taskMetadata[m.viewingFile.Path]; ok {
				task.TaskMetadata = taskMeta
			}
		}
		
		m.statusMsg = fmt.Sprintf("Updated %s to %s", field, value)
	}
	
	return nil
}

// updateProjectField updates a specific field in the project metadata
func (m *Model) updateProjectField(field, value string) error {
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
	
	// Update the metadata
	if projectMeta, ok := fm.Metadata.(denote.ProjectMetadata); ok {
		switch field {
		case "title":
			projectMeta.Title = value
		case "priority":
			projectMeta.Priority = value
		case "status":
			projectMeta.Status = value
		case "due_date":
			// Parse natural language dates
			if value != "" {
				parsed, err := denote.ParseNaturalDate(value)
				if err != nil {
					return fmt.Errorf("invalid date: %s (try: 2d, 1w, friday, jan 15, 2024-01-15)", value)
				}
				projectMeta.DueDate = parsed
			} else {
				projectMeta.DueDate = ""
			}
		case "area":
			projectMeta.Area = value
		case "tags":
			// Split by spaces
			if value == "" {
				projectMeta.Tags = []string{}
			} else {
				projectMeta.Tags = strings.Fields(value)
			}
		}
		
		// Write updated content
		newContent, err := denote.WriteFrontmatterFile(projectMeta, fm.Content)
		if err != nil {
			return fmt.Errorf("failed to write frontmatter: %w", err)
		}
		
		// Check if we need to rename the file (for tag or title changes)
		oldPath := m.viewingFile.Path
		newPath := oldPath
		
		if field == "tags" || field == "title" {
			if field == "title" {
				// For title changes, we need to update the slug
				// Parse the current filename to get components
				parser := denote.NewParser()
				oldFile, err := parser.ParseFilename(filepath.Base(oldPath))
				if err != nil {
					return fmt.Errorf("failed to parse filename: %w", err)
				}
				
				// Create new filename with updated title slug
				// Convert title to slug (same logic as titleToSlug in create.go)
				slug := strings.ToLower(projectMeta.Title)
				slug = strings.Map(func(r rune) rune {
					if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
						return r
					}
					return '-'
				}, slug)
				for strings.Contains(slug, "--") {
					slug = strings.ReplaceAll(slug, "--", "-")
				}
				slug = strings.Trim(slug, "-")
				
				newBasename := fmt.Sprintf("%s--%s", oldFile.ID, slug)
				
				// Add tags
				if len(oldFile.Tags) > 0 {
					newBasename += "__" + strings.Join(oldFile.Tags, "__")
				}
				newBasename += ".md"
				
				// Create full path
				dir := filepath.Dir(oldPath)
				newPath = filepath.Join(dir, newBasename)
				
				// Rename the file
				if newPath != oldPath {
					if err := os.Rename(oldPath, newPath); err != nil {
						return fmt.Errorf("failed to rename file: %w", err)
					}
				}
			} else {
				// Tag changes - use existing logic
				allTags := []string{"project"} // Always include project tag
				for _, tag := range projectMeta.Tags {
					if tag != "project" {
						allTags = append(allTags, tag)
					}
				}
				
				// Rename file to reflect new tags
				renamed, err := denote.RenameFileForTags(oldPath, allTags)
				if err != nil {
					return fmt.Errorf("failed to rename file: %w", err)
				}
				newPath = renamed
			}
		}
		
		// Write to file (at potentially new path)
		if err := os.WriteFile(newPath, newContent, 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		
		// Update our in-memory copy
		m.viewingProject.ProjectMetadata = projectMeta
		
		// Update path references if file was renamed
		if newPath != oldPath {
			// Update viewing file path
			m.viewingFile.Path = newPath
			
			// Update in main metadata map
			if project, ok := m.projectMetadata[oldPath]; ok {
				delete(m.projectMetadata, oldPath)
				m.projectMetadata[newPath] = project
				project.ProjectMetadata = projectMeta
			}
			
			// Trigger a rescan to update the file list
			m.scanFiles()
		} else {
			// Just update metadata if no rename
			if project, ok := m.projectMetadata[m.viewingFile.Path]; ok {
				project.ProjectMetadata = projectMeta
			}
		}
		
		m.statusMsg = fmt.Sprintf("Updated %s", field)
		return nil
	} else {
		return fmt.Errorf("file is not a project")
	}
	
	return nil
}

// editFile opens a file in the external editor
func (m Model) editFile(path string) tea.Cmd {
	// Use tea.ExecProcess to properly suspend the TUI
	cmd := exec.Command(m.config.Editor, path)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			return fmt.Errorf("failed to edit file: %w", err)
		}
		// Return a message to trigger file check and potential rename
		return fileEditedMsg{path: path}
	})
}

// loadProjectTasks loads all tasks assigned to the current viewing project
func (m *Model) loadProjectTasks() {
	if m.viewingProject == nil {
		return
	}
	
	m.projectTasks = []denote.Task{}
	projectID := m.viewingProject.File.ID
	
	// Go through all task files and find ones assigned to this project
	for _, file := range m.files {
		if file.IsTask() {
			// Check if we have metadata cached
			if task, ok := m.taskMetadata[file.Path]; ok {
				if task.TaskMetadata.ProjectID == projectID {
					m.projectTasks = append(m.projectTasks, *task)
				}
			} else {
				// Load metadata to check
				if task, err := denote.ParseTaskFile(file.Path); err == nil {
					m.taskMetadata[file.Path] = task
					if task.TaskMetadata.ProjectID == projectID {
						m.projectTasks = append(m.projectTasks, *task)
					}
				}
			}
		}
	}
	
	m.projectTasksCursor = 0
}

// taskMatchesSearch performs fuzzy search on task metadata
func (m *Model) taskMatchesSearch(task *denote.Task, query string) bool {
	query = strings.ToLower(query)
	
	// Helper function for fuzzy matching
	fuzzyMatch := func(text, pattern string) bool {
		if pattern == "" {
			return true
		}
		
		patternIdx := 0
		for _, ch := range text {
			if patternIdx < len(pattern) && ch == rune(pattern[patternIdx]) {
				patternIdx++
			}
		}
		
		return patternIdx == len(pattern)
	}
	
	// Search in status
	if fuzzyMatch(strings.ToLower(task.Status), query) {
		return true
	}
	
	// Search in priority
	if fuzzyMatch(strings.ToLower(task.Priority), query) {
		return true
	}
	
	// Search in area
	if fuzzyMatch(strings.ToLower(task.Area), query) {
		return true
	}
	
	// Search in assignee
	if fuzzyMatch(strings.ToLower(task.Assignee), query) {
		return true
	}
	
	// Search in project name if we have it cached
	if task.ProjectID != "" {
		for _, file := range m.files {
			if file.ID == task.ProjectID && file.IsProject() {
				if fuzzyMatch(strings.ToLower(file.Title), query) {
					return true
				}
				break
			}
		}
	}
	
	return false
}

// updateCurrentTaskStatus updates the status of the currently selected task
func (m *Model) updateCurrentTaskStatus(newStatus string) error {
	if m.cursor >= len(m.filtered) {
		return fmt.Errorf("no task selected")
	}
	
	file := m.filtered[m.cursor]
	if !file.IsTask() {
		return fmt.Errorf("selected item is not a task")
	}
	
	// Update the task status
	err := denote.UpdateTaskStatus(file.Path, newStatus)
	if err != nil {
		return err
	}
	
	// Update the cached metadata
	if task, ok := m.taskMetadata[file.Path]; ok {
		task.TaskMetadata.Status = newStatus
	}
	
	return nil
}

// deleteFile deletes a file from the filesystem
func (m *Model) deleteFile(path string) error {
	return os.Remove(path)
}

// findTasksAffectedByProjectDeletion finds all tasks that reference the current project
func (m *Model) findTasksAffectedByProjectDeletion() {
	if m.viewingProject == nil {
		return
	}
	
	m.affectedTasks = []denote.Task{}
	projectID := m.viewingProject.File.ID
	
	// Go through all task files and find ones assigned to this project
	for _, file := range m.files {
		if file.IsTask() {
			// Check if we have metadata cached
			if task, ok := m.taskMetadata[file.Path]; ok {
				if task.TaskMetadata.ProjectID == projectID {
					m.affectedTasks = append(m.affectedTasks, *task)
				}
			} else {
				// Load metadata to check
				if task, err := denote.ParseTaskFile(file.Path); err == nil {
					m.taskMetadata[file.Path] = task
					if task.TaskMetadata.ProjectID == projectID {
						m.affectedTasks = append(m.affectedTasks, *task)
					}
				}
			}
		}
	}
}

// clearProjectFromTask removes the project_id from a task
func (m *Model) clearProjectFromTask(taskPath string) error {
	// Read the file content
	content, err := os.ReadFile(taskPath)
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
		// Clear the project ID
		taskMeta.ProjectID = ""
		
		// Write updated content
		newContent, err := denote.WriteFrontmatterFile(taskMeta, fm.Content)
		if err != nil {
			return fmt.Errorf("failed to write frontmatter: %w", err)
		}
		
		// Write to file
		if err := os.WriteFile(taskPath, newContent, 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		
		// Update cached metadata
		if task, ok := m.taskMetadata[taskPath]; ok {
			task.TaskMetadata.ProjectID = ""
		}
	}
	
	return nil
}

// updateProjectTaskStatus updates the status of the currently selected task in project view
func (m *Model) updateProjectTaskStatus(newStatus string) error {
	if m.projectTasksCursor >= len(m.projectTasks) {
		return fmt.Errorf("no task selected")
	}
	
	task := &m.projectTasks[m.projectTasksCursor]
	
	// Update the task status
	err := denote.UpdateTaskStatus(task.File.Path, newStatus)
	if err != nil {
		return err
	}
	
	// Update the cached metadata
	task.TaskMetadata.Status = newStatus
	if cachedTask, ok := m.taskMetadata[task.File.Path]; ok {
		cachedTask.TaskMetadata.Status = newStatus
	}
	
	return nil
}

func (m Model) View() string {
	if m.err != nil {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Render("Error: " + m.err.Error())
	}
	
	switch m.mode {
	case ModeHelp:
		return m.renderHelp()
	case ModePreview:
		return m.renderPreview()
	case ModeTaskView:
		return m.renderTaskView()
	case ModeProjectView:
		return m.renderProjectView()
	case ModeCreate:
		return m.renderCreate()
	case ModeCreateTags:
		return m.renderCreateTags()
	case ModeAreaFilter:
		return m.renderAreaFilter()
	case ModeSort:
		return m.renderSort()
	case ModeStateMenu:
		return m.renderStateMenu()
	case ModeConfirmDelete:
		return m.renderConfirmDelete()
	case ModeFilterMenu:
		return m.renderFilterMenu()
	case ModePriorityFilter:
		return m.renderPriorityFilter()
	case ModeStateFilter:
		return m.renderStateFilter()
	case ModeLogEntry:
		return m.renderLogEntry()
	case ModeProjectSelect:
		return m.renderProjectSelect()
	default:
		return m.renderNormal()
	}
}

func (m *Model) addLogEntry() error {
	if m.loggingFile == nil || m.logInput == "" {
		return fmt.Errorf("no file selected or empty log input")
	}
	
	// Read the file
	content, err := os.ReadFile(m.loggingFile.Path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	
	// Find the end of frontmatter
	lines := strings.Split(string(content), "\n")
	frontmatterEnd := -1
	inFrontmatter := false
	
	for i, line := range lines {
		if i == 0 && line == "---" {
			inFrontmatter = true
			continue
		}
		if inFrontmatter && line == "---" {
			frontmatterEnd = i
			break
		}
	}
	
	if frontmatterEnd == -1 {
		return fmt.Errorf("no frontmatter found in file")
	}
	
	// Format the log entry with timestamp
	now := time.Now()
	// Use reference time to get day name: Mon Jan 2 15:04:05 MST 2006
	timestamp := now.Format("[2006-01-02 Mon]")
	logEntry := fmt.Sprintf("%s: %s", timestamp, m.logInput)
	
	// Build the new content
	var newLines []string
	
	// Add frontmatter
	newLines = append(newLines, lines[:frontmatterEnd+1]...)
	
	// Find where to insert the log entry
	insertPos := frontmatterEnd + 1
	
	// Skip any existing blank lines after frontmatter
	for insertPos < len(lines) && lines[insertPos] == "" {
		insertPos++
	}
	
	// Add a blank line if needed
	if insertPos == frontmatterEnd+1 || (insertPos < len(lines) && lines[insertPos-1] != "") {
		newLines = append(newLines, "")
	}
	
	// Add the log entry
	newLines = append(newLines, logEntry)
	
	// Add the rest of the content
	if insertPos < len(lines) {
		// If there's existing content, ensure there's a blank line after our log entry
		if lines[insertPos] != "" {
			newLines = append(newLines, "")
		}
		newLines = append(newLines, lines[insertPos:]...)
	}
	
	// Write back to file
	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(m.loggingFile.Path, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	
	return nil
}