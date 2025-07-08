package denote

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"
	
	"gopkg.in/yaml.v3"
)

// FrontmatterFile represents a file with YAML frontmatter
type FrontmatterFile struct {
	Metadata interface{} // Can be NoteMetadata, TaskMetadata, or ProjectMetadata
	Content  string      // The markdown content after frontmatter
}

// ParseFrontmatterFile reads and validates a file with YAML frontmatter
func ParseFrontmatterFile(content []byte) (*FrontmatterFile, error) {
	contentStr := string(content)
	
	// Must start with ---
	if !strings.HasPrefix(contentStr, "---\n") && !strings.HasPrefix(contentStr, "---\r\n") {
		return nil, fmt.Errorf("file does not start with YAML frontmatter marker")
	}
	
	// Find the closing --- by looking for valid YAML
	lines := strings.Split(contentStr, "\n")
	var frontmatterLines []string
	endIndex := -1
	
	for i := 1; i < len(lines); i++ {
		line := strings.TrimRight(lines[i], "\r")
		
		if line == "---" {
			// Try to parse what we have so far as YAML
			testYAML := strings.Join(frontmatterLines, "\n")
			var test map[string]interface{}
			
			if err := yaml.Unmarshal([]byte(testYAML), &test); err == nil {
				// It's valid YAML, so this is the end marker
				endIndex = i
				break
			}
			// Not valid YAML yet, keep going
		}
		
		frontmatterLines = append(frontmatterLines, line)
	}
	
	if endIndex == -1 {
		return nil, fmt.Errorf("no valid YAML frontmatter found")
	}
	
	// Extract frontmatter and content
	frontmatterStr := strings.Join(frontmatterLines, "\n")
	contentLines := lines[endIndex+1:]
	
	// First check the type field to determine what kind of metadata this is
	var typeCheck struct {
		Type    string `yaml:"type"`
		IndexID int    `yaml:"index_id"`
	}
	if err := yaml.Unmarshal([]byte(frontmatterStr), &typeCheck); err == nil {
		// If type is explicitly set, use that
		if typeCheck.Type == "project" {
			var projectMeta ProjectMetadata
			if err := yaml.Unmarshal([]byte(frontmatterStr), &projectMeta); err == nil {
				return &FrontmatterFile{
					Metadata: projectMeta,
					Content:  strings.Join(contentLines, "\n"),
				}, nil
			}
		} else if typeCheck.Type == "task" {
			var taskMeta TaskMetadata
			if err := yaml.Unmarshal([]byte(frontmatterStr), &taskMeta); err == nil {
				return &FrontmatterFile{
					Metadata: taskMeta,
					Content:  strings.Join(contentLines, "\n"),
				}, nil
			}
		}
		
		// If no type but has index_id, try to determine by other means
		if typeCheck.IndexID > 0 {
			// Try project first (they're less common)
			var projectMeta ProjectMetadata
			if err := yaml.Unmarshal([]byte(frontmatterStr), &projectMeta); err == nil {
				// Check if it has project-specific fields or patterns
				// For now, we'll need another way to distinguish
				// Let's check if the content suggests it's a project
				// This is a bit fragile but necessary without explicit type
				return &FrontmatterFile{
					Metadata: projectMeta,
					Content:  strings.Join(contentLines, "\n"),
				}, nil
			}
			
			// Fall back to task
			var taskMeta TaskMetadata
			if err := yaml.Unmarshal([]byte(frontmatterStr), &taskMeta); err == nil {
				return &FrontmatterFile{
					Metadata: taskMeta,
					Content:  strings.Join(contentLines, "\n"),
				}, nil
			}
		}
	}
	
	// Default to note metadata
	var noteMeta NoteMetadata
	if err := yaml.Unmarshal([]byte(frontmatterStr), &noteMeta); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}
	
	metadata := noteMeta
	
	return &FrontmatterFile{
		Metadata: metadata,
		Content:  strings.Join(contentLines, "\n"),
	}, nil
}

// WriteFrontmatterFile creates file content with validated YAML frontmatter
func WriteFrontmatterFile(metadata interface{}, content string) ([]byte, error) {
	// Validate that metadata has required fields based on type
	switch m := metadata.(type) {
	case NoteMetadata:
		if m.Title == "" {
			return nil, fmt.Errorf("note title is required")
		}
		
	case TaskMetadata:
		if m.Title == "" {
			return nil, fmt.Errorf("task title is required")
		}
		if m.IndexID <= 0 {
			return nil, fmt.Errorf("task index ID must be positive")
		}
		
	case ProjectMetadata:
		if m.Title == "" {
			return nil, fmt.Errorf("project title is required")
		}
		if m.IndexID <= 0 {
			return nil, fmt.Errorf("project index ID must be positive")
		}
		
	default:
		return nil, fmt.Errorf("unsupported metadata type")
	}
	
	// Marshal to YAML
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	
	if err := encoder.Encode(metadata); err != nil {
		return nil, fmt.Errorf("failed to encode metadata: %w", err)
	}
	
	// Build complete file content
	fileContent := fmt.Sprintf("---\n%s---\n\n%s", buf.String(), content)
	
	return []byte(fileContent), nil
}

// ValidateYAMLFrontmatter checks if content has valid YAML frontmatter
func ValidateYAMLFrontmatter(content []byte) error {
	_, err := ParseFrontmatterFile(content)
	return err
}

// AddLogEntry adds a timestamped log entry to a task file
func AddLogEntry(filepath string, message string) error {
	// Read the file
	content, err := os.ReadFile(filepath)
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
	logEntry := fmt.Sprintf("%s: %s", timestamp, message)
	
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
	
	// Check if we need to add blank lines
	if insertPos < len(lines) {
		// Add blank line before log entry if there's content after frontmatter
		newLines = append(newLines, "")
	}
	
	// Add the log entry
	newLines = append(newLines, logEntry)
	
	// Add the rest of the content
	if insertPos < len(lines) {
		// Add a blank line after log entry if there's more content
		newLines = append(newLines, "")
		newLines = append(newLines, lines[insertPos:]...)
	}
	
	// Write back to file
	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(filepath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	
	return nil
}