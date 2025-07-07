package denote

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// CounterData represents the counter file per spec v2.0.1
type CounterData struct {
	NextIndexID int    `json:"next_index_id"`
	SpecVersion string `json:"spec_version"`
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
			// Counter doesn't exist, scan for highest ID
			maxID := findMaxIndexID(dir)
			
			counter := &IDCounter{
				CounterData: CounterData{
					NextIndexID: maxID + 1,
					SpecVersion: "2.0.1",
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
	
	// First try to parse as new format
	var counterData CounterData
	if err := json.Unmarshal(data, &counterData); err == nil && counterData.NextIndexID > 0 {
		counter := &IDCounter{
			CounterData: counterData,
			filePath:    counterFile,
		}
		
		// Ensure spec version is set
		if counter.SpecVersion == "" {
			counter.SpecVersion = "2.0.1"
		}
		
		return counter, nil
	}
	
	// Try to parse as old format for migration
	var oldCounterData struct {
		NextTaskID    int    `json:"next_task_id"`
		NextProjectID int    `json:"next_project_id"`
		SpecVersion   string `json:"spec_version"`
	}
	
	if err := json.Unmarshal(data, &oldCounterData); err != nil {
		return nil, fmt.Errorf("failed to parse counter file: %w", err)
	}
	
	// Migrate from old format
	nextID := oldCounterData.NextTaskID
	if oldCounterData.NextProjectID > nextID {
		nextID = oldCounterData.NextProjectID
	}
	
	counter := &IDCounter{
		CounterData: CounterData{
			NextIndexID: nextID,
			SpecVersion: "2.0.1",
		},
		filePath: counterFile,
	}
	
	// Save the migrated counter
	if err := counter.save(); err != nil {
		return nil, fmt.Errorf("failed to save migrated counter: %w", err)
	}
	
	return counter, nil
}

// findMaxIndexID scans the directory for the highest index ID
func findMaxIndexID(dir string) int {
	maxID := 0
	
	// Check task files
	taskPattern := filepath.Join(dir, "*__task*.md")
	taskFiles, _ := filepath.Glob(taskPattern)
	
	for _, file := range taskFiles {
		task, err := ParseTaskFile(file)
		if err != nil {
			continue
		}
		if task.IndexID > maxID {
			maxID = task.IndexID
		}
	}
	
	// Check project files
	projPattern := filepath.Join(dir, "*__project*.md")
	projFiles, _ := filepath.Glob(projPattern)
	
	for _, file := range projFiles {
		project, err := ParseProjectFile(file)
		if err != nil {
			continue
		}
		if project.IndexID > maxID {
			maxID = project.IndexID
		}
	}
	
	return maxID
}

// NextIndexID returns the next index ID and increments the counter
func (c *IDCounter) NextIndexID() (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	id := c.CounterData.NextIndexID
	c.CounterData.NextIndexID++
	
	if err := c.save(); err != nil {
		// Rollback on save failure
		c.CounterData.NextIndexID--
		return 0, fmt.Errorf("failed to save counter: %w", err)
	}
	
	return id, nil
}

// Deprecated: Use NextIndexID instead
func (c *IDCounter) NextTaskID() (int, error) {
	return c.NextIndexID()
}

// Deprecated: Use NextIndexID instead
func (c *IDCounter) NextProjectID() (int, error) {
	return c.NextIndexID()
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