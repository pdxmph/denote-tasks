package denote

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// CounterData represents the counter file per spec v2.0.0
type CounterData struct {
	NextTaskID    int    `json:"next_task_id"`
	NextProjectID int    `json:"next_project_id"`
	SpecVersion   string `json:"spec_version"`
}

// IDCounter manages sequential IDs for tasks and projects
type IDCounter struct {
	CounterData
	mu       sync.Mutex
	filePath string
}

var (
	globalCounter     *IDCounter
	globalCounterOnce sync.Once
)

// GetIDCounter returns the singleton ID counter for the given directory
func GetIDCounter(dir string) (*IDCounter, error) {
	var err error
	globalCounterOnce.Do(func() {
		globalCounter, err = loadOrCreateCounter(dir)
	})
	return globalCounter, err
}

// loadOrCreateCounter loads an existing counter or creates a new one
func loadOrCreateCounter(dir string) (*IDCounter, error) {
	counterFile := filepath.Join(dir, ".denote-task-counter.json")
	
	// Try to load existing counter
	data, err := os.ReadFile(counterFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Counter doesn't exist, scan for highest IDs
			maxTaskID := findMaxTaskID(dir)
			maxProjectID := findMaxProjectID(dir)
			
			counter := &IDCounter{
				CounterData: CounterData{
					NextTaskID:    maxTaskID + 1,
					NextProjectID: maxProjectID + 1,
					SpecVersion:   "2.0.0",
				},
				filePath: counterFile,
			}
			
			// Save the initial counter
			if err := counter.save(); err != nil {
				return nil, fmt.Errorf("failed to save initial counter: %w", err)
			}
			
			return counter, nil
		}
		return nil, fmt.Errorf("failed to read counter file: %w", err)
	}
	
	// Parse existing counter
	var counterData CounterData
	if err := json.Unmarshal(data, &counterData); err != nil {
		return nil, fmt.Errorf("failed to parse counter file: %w", err)
	}
	
	counter := &IDCounter{
		CounterData: counterData,
		filePath:    counterFile,
	}
	
	// Ensure spec version is set
	if counter.SpecVersion == "" {
		counter.SpecVersion = "2.0.0"
	}
	
	return counter, nil
}

// findMaxTaskID scans the directory for the highest task ID
func findMaxTaskID(dir string) int {
	maxID := 0
	
	pattern := filepath.Join(dir, "*__task*.md")
	files, _ := filepath.Glob(pattern)
	
	for _, file := range files {
		task, err := ParseTaskFile(file)
		if err != nil {
			continue
		}
		if task.TaskID > maxID {
			maxID = task.TaskID
		}
	}
	
	return maxID
}

// findMaxProjectID scans the directory for the highest project ID
func findMaxProjectID(dir string) int {
	maxID := 0
	
	pattern := filepath.Join(dir, "*__project*.md")
	files, _ := filepath.Glob(pattern)
	
	for _, file := range files {
		project, err := ParseProjectFile(file)
		if err != nil {
			continue
		}
		if project.ProjectID > maxID {
			maxID = project.ProjectID
		}
	}
	
	return maxID
}

// NextTaskID returns the next task ID and increments the counter
func (c *IDCounter) NextTaskID() (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	id := c.CounterData.NextTaskID
	c.CounterData.NextTaskID++
	
	if err := c.save(); err != nil {
		// Rollback on save failure
		c.CounterData.NextTaskID--
		return 0, fmt.Errorf("failed to save counter: %w", err)
	}
	
	return id, nil
}

// NextProjectID returns the next project ID and increments the counter
func (c *IDCounter) NextProjectID() (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	id := c.CounterData.NextProjectID
	c.CounterData.NextProjectID++
	
	if err := c.save(); err != nil {
		// Rollback on save failure
		c.CounterData.NextProjectID--
		return 0, fmt.Errorf("failed to save counter: %w", err)
	}
	
	return id, nil
}

// save writes the counter to disk
func (c *IDCounter) save() error {
	data, err := json.MarshalIndent(c.CounterData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal counter: %w", err)
	}
	
	// Write to temp file first for atomicity
	tempFile := c.filePath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	
	// Rename temp file to actual file (atomic on most systems)
	if err := os.Rename(tempFile, c.filePath); err != nil {
		os.Remove(tempFile) // Clean up temp file
		return fmt.Errorf("failed to rename counter file: %w", err)
	}
	
	return nil
}

// ResetSingleton resets the singleton (useful for testing or config changes)
func ResetSingleton() {
	globalCounterOnce = sync.Once{}
	globalCounter = nil
}