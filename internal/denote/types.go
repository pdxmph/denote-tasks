package denote

import (
	"time"
)

// File represents the basic Denote file structure
type File struct {
	ID    string   // Denote timestamp ID (e.g., "20250704T151739")
	Title string   // Human-readable title from slug
	Tags  []string // Tags from filename
	Path  string   // Full file path
}

// TaskMetadata represents task-specific frontmatter per spec v2.0.0
type TaskMetadata struct {
	Title     string   `yaml:"title"`      // Required: Human-readable title
	TaskID    int      `yaml:"task_id"`    // Required: Sequential ID for CLI
	Type      string   `yaml:"type"`       // Required: "task"
	Status    string   `yaml:"status"`     // Default: "open"
	Priority  string   `yaml:"priority"`   // p1, p2, p3
	DueDate   string   `yaml:"due_date"`   // YYYY-MM-DD format
	StartDate string   `yaml:"start_date"` // YYYY-MM-DD format
	Estimate  int      `yaml:"estimate"`   // Fibonacci: 1,2,3,5,8,13
	ProjectID string   `yaml:"project_id"` // Denote ID of project (v2.0.0)
	Area      string   `yaml:"area"`       // Life context
	Assignee  string   `yaml:"assignee"`   // Person responsible
	Tags      []string `yaml:"tags"`       // Additional tags beyond filename
}

// ProjectMetadata represents project-specific frontmatter per spec v2.0.0
type ProjectMetadata struct {
	Title     string   `yaml:"title"`      // Required: Human-readable title
	ProjectID int      `yaml:"project_id"` // Required: Sequential ID for CLI
	Type      string   `yaml:"type"`       // Required: "project"
	Status    string   `yaml:"status"`     // Default: "active"
	Priority  string   `yaml:"priority"`   // p1, p2, p3
	DueDate   string   `yaml:"due_date"`   // YYYY-MM-DD format
	StartDate string   `yaml:"start_date"` // YYYY-MM-DD format
	Area      string   `yaml:"area"`       // Life context
	Tags      []string `yaml:"tags"`       // Additional tags beyond filename
}

// Task combines File info with TaskMetadata
type Task struct {
	File
	TaskMetadata
	ModTime time.Time
	Content string // Full file content
}

// Project combines File info with ProjectMetadata
type Project struct {
	File
	ProjectMetadata
	ModTime time.Time
	Content string
}

// Common status values
const (
	// Task statuses
	TaskStatusOpen      = "open"
	TaskStatusDone      = "done"
	TaskStatusPaused    = "paused"
	TaskStatusDelegated = "delegated"
	TaskStatusDropped   = "dropped"

	// Project statuses
	ProjectStatusActive    = "active"
	ProjectStatusCompleted = "completed"
	ProjectStatusPaused    = "paused"
	ProjectStatusCancelled = "cancelled"

	// Priority levels
	PriorityP1 = "p1"
	PriorityP2 = "p2"
	PriorityP3 = "p3"

	// File types
	TypeTask    = "task"
	TypeProject = "project"
)

// IsValidTaskStatus checks if a status is valid for tasks
func IsValidTaskStatus(status string) bool {
	switch status {
	case TaskStatusOpen, TaskStatusDone, TaskStatusPaused, TaskStatusDelegated, TaskStatusDropped:
		return true
	}
	return false
}

// IsValidProjectStatus checks if a status is valid for projects
func IsValidProjectStatus(status string) bool {
	switch status {
	case ProjectStatusActive, ProjectStatusCompleted, ProjectStatusPaused, ProjectStatusCancelled:
		return true
	}
	return false
}

// IsValidPriority checks if a priority is valid
func IsValidPriority(priority string) bool {
	switch priority {
	case PriorityP1, PriorityP2, PriorityP3:
		return true
	}
	return false
}

// IsOverdue checks if a task/project is overdue
func IsOverdue(dueDateStr string) bool {
	if dueDateStr == "" {
		return false
	}
	// Parse date in local timezone to avoid timezone issues
	loc := time.Now().Location()
	dueDate, err := time.ParseInLocation("2006-01-02", dueDateStr, loc)
	if err != nil {
		return false
	}
	// Get current time at start of day in local timezone
	now := time.Now().In(loc)
	nowStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	dueStart := time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 0, 0, 0, 0, loc)
	
	return dueStart.Before(nowStart)
}

// DaysUntilDue returns the number of days until the due date
func DaysUntilDue(dueDateStr string) int {
	if dueDateStr == "" {
		return 0
	}
	// Parse date in local timezone to avoid timezone issues
	loc := time.Now().Location()
	dueDate, err := time.ParseInLocation("2006-01-02", dueDateStr, loc)
	if err != nil {
		return 0
	}
	// Get current time at start of day in local timezone
	now := time.Now().In(loc)
	nowStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	dueStart := time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 0, 0, 0, 0, loc)
	
	return int(dueStart.Sub(nowStart).Hours() / 24)
}

// IsDueThisWeek checks if a task is due within the next 7 days
func IsDueThisWeek(dueDateStr string) bool {
	days := DaysUntilDue(dueDateStr)
	return days >= 0 && days <= 7
}

// GetParsedStartDate returns the parsed start date
func (t *Task) GetParsedStartDate() *time.Time {
	if t.StartDate == "" {
		return nil
	}
	parsed, err := time.Parse("2006-01-02", t.StartDate)
	if err != nil {
		return nil
	}
	return &parsed
}

// GetParsedDueDate returns the parsed due date
func (t *Task) GetParsedDueDate() *time.Time {
	if t.DueDate == "" {
		return nil
	}
	parsed, err := time.Parse("2006-01-02", t.DueDate)
	if err != nil {
		return nil
	}
	return &parsed
}

// GetParsedStartDate returns the parsed start date for a project
func (p *Project) GetParsedStartDate() *time.Time {
	if p.StartDate == "" {
		return nil
	}
	parsed, err := time.Parse("2006-01-02", p.StartDate)
	if err != nil {
		return nil
	}
	return &parsed
}

// GetParsedDueDate returns the parsed due date for a project
func (p *Project) GetParsedDueDate() *time.Time {
	if p.DueDate == "" {
		return nil
	}
	parsed, err := time.Parse("2006-01-02", p.DueDate)
	if err != nil {
		return nil
	}
	return &parsed
}