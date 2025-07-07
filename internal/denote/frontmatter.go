package denote

import (
	"bytes"
	"fmt"
	"strings"
	
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
	
	// Try to parse as different types based on content
	// First try task (check for index_id)
	var taskMeta TaskMetadata
	if err := yaml.Unmarshal([]byte(frontmatterStr), &taskMeta); err == nil && taskMeta.IndexID > 0 {
		return &FrontmatterFile{
			Metadata: taskMeta,
			Content:  strings.Join(contentLines, "\n"),
		}, nil
	}
	
	// Then try project (check for index_id)
	var projectMeta ProjectMetadata
	if err := yaml.Unmarshal([]byte(frontmatterStr), &projectMeta); err == nil && projectMeta.IndexID > 0 {
		return &FrontmatterFile{
			Metadata: projectMeta,
			Content:  strings.Join(contentLines, "\n"),
		}, nil
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