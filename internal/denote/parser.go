package denote

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	// Denote filename pattern: YYYYMMDDTHHMMSS-title__tags.md or YYYYMMDDTHHMMSS--title__tags.md
	denotePattern = regexp.MustCompile(`^(\d{8}T\d{6})-{1,2}([^_]+)(?:__(.+))?\.md$`)
)

// Parser handles parsing of Denote files
type Parser struct{}

// NewParser creates a new parser
func NewParser() *Parser {
	return &Parser{}
}

// ParseFrontmatter extracts frontmatter metadata from a file
func (p *Parser) ParseFrontmatter(path string) (map[string]interface{}, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	// Use the new strict parser
	file, err := ParseFrontmatterFile(content)
	if err != nil {
		return nil, err
	}
	
	// Convert to generic map for backward compatibility
	yamlData, err := yaml.Marshal(file.Metadata)
	if err != nil {
		return nil, err
	}
	
	var result map[string]interface{}
	if err := yaml.Unmarshal(yamlData, &result); err != nil {
		return nil, err
	}
	
	return result, nil
}

// ParseFilename extracts Denote components from a filename
func (p *Parser) ParseFilename(filename string) (*File, error) {
	base := filepath.Base(filename)
	matches := denotePattern.FindStringSubmatch(base)
	if len(matches) < 3 {
		return nil, fmt.Errorf("not a valid denote filename: %s", base)
	}

	file := &File{
		ID:    matches[1],
		Slug:  matches[2],
		Title: titleFromSlug(matches[2]), // Convert slug to readable title as fallback
		Tags:  []string{},
		Path:  filename,
	}

	// Parse tags if present
	if len(matches) > 3 && matches[3] != "" {
		file.Tags = strings.Split(matches[3], "_")
	}

	return file, nil
}

// ParseTaskFile reads and parses a task file
func ParseTaskFile(path string) (*Task, error) {
	// Parse filename first
	p := NewParser()
	file, err := p.ParseFilename(path)
	if err != nil {
		return nil, err
	}

	// Check if it's a task file
	if !contains(file.Tags, "task") {
		return nil, fmt.Errorf("not a task file: %s", path)
	}

	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Get file info
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	task := &Task{
		File:    *file,
		ModTime: info.ModTime(),
		Content: string(content),
	}

	// Parse frontmatter using strict parser
	if file, err := ParseFrontmatterFile(content); err == nil {
		if taskMeta, ok := file.Metadata.(TaskMetadata); ok {
			task.TaskMetadata = taskMeta
		}
	}

	// Set defaults per spec
	if task.Status == "" {
		task.Status = TaskStatusOpen
	}
	
	// Use metadata title if available, otherwise fall back to filename title
	if task.TaskMetadata.Title != "" {
		task.File.Title = task.TaskMetadata.Title
	}

	return task, nil
}

// ParseProjectFile reads and parses a project file
func ParseProjectFile(path string) (*Project, error) {
	// Parse filename first
	p := NewParser()
	file, err := p.ParseFilename(path)
	if err != nil {
		return nil, err
	}

	// Check if it's a project file
	if !contains(file.Tags, "project") {
		return nil, fmt.Errorf("not a project file: %s", path)
	}

	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Get file info
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	project := &Project{
		File:    *file,
		ModTime: info.ModTime(),
		Content: string(content),
	}

	// Parse frontmatter using strict parser
	if file, err := ParseFrontmatterFile(content); err == nil {
		if projMeta, ok := file.Metadata.(ProjectMetadata); ok {
			project.ProjectMetadata = projMeta
		}
	}

	// Set defaults per spec
	if project.Status == "" {
		project.Status = ProjectStatusActive
	}
	
	// Use metadata title if available, otherwise fall back to filename title
	if project.ProjectMetadata.Title != "" {
		project.File.Title = project.ProjectMetadata.Title
	}

	return project, nil
}

// parseFrontmatter extracts YAML frontmatter from file content
func parseFrontmatter(content []byte) (interface{}, string, error) {
	contentStr := string(content)
	if !strings.HasPrefix(contentStr, "---\n") {
		return nil, contentStr, fmt.Errorf("no frontmatter found")
	}

	// Use robust frontmatter parsing
	lines := strings.Split(contentStr, "\n")
	frontmatterEndLine := -1
	inFrontmatter := false
	
	for i, line := range lines {
		if i == 0 && line == "---" {
			inFrontmatter = true
			continue
		}
		
		if inFrontmatter && line == "---" {
			// Check if this looks like the end of frontmatter
			possibleYAML := strings.Join(lines[1:i], "\n")
			if looksLikeValidFrontmatter(possibleYAML) {
				frontmatterEndLine = i
				break
			}
		}
	}
	
	if frontmatterEndLine == -1 {
		return nil, contentStr, fmt.Errorf("frontmatter not properly closed")
	}

	// Extract frontmatter YAML and remaining content
	frontmatterStr := strings.Join(lines[1:frontmatterEndLine], "\n")
	remaining := strings.Join(lines[frontmatterEndLine+1:], "\n")

	// First check the type field to determine what to unmarshal
	var typeCheck struct {
		Type string `yaml:"type"`
	}
	if err := yaml.Unmarshal([]byte(frontmatterStr), &typeCheck); err == nil {
		switch typeCheck.Type {
		case TypeTask:
			var taskMeta TaskMetadata
			if err := yaml.Unmarshal([]byte(frontmatterStr), &taskMeta); err == nil {
				return &taskMeta, remaining, nil
			}
		case TypeProject:
			var projMeta ProjectMetadata
			if err := yaml.Unmarshal([]byte(frontmatterStr), &projMeta); err == nil {
				return &projMeta, remaining, nil
			}
		}
	}

	// Fall back to checking for index_id
	var taskMeta TaskMetadata
	if err := yaml.Unmarshal([]byte(frontmatterStr), &taskMeta); err == nil && taskMeta.IndexID > 0 {
		return &taskMeta, remaining, nil
	}

	var projMeta ProjectMetadata
	if err := yaml.Unmarshal([]byte(frontmatterStr), &projMeta); err == nil && projMeta.IndexID > 0 {
		return &projMeta, remaining, nil
	}

	return nil, contentStr, fmt.Errorf("could not determine file type from frontmatter")
}

// titleFromSlug converts a kebab-case slug to a title
func titleFromSlug(slug string) string {
	// Simply replace hyphens with spaces, no capitalization
	return strings.ReplaceAll(slug, "-", " ")
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// looksLikeValidFrontmatter checks if content appears to be valid YAML frontmatter
func looksLikeValidFrontmatter(content string) bool {
	if content == "" {
		return true // Empty frontmatter is technically valid
	}
	
	// At least check that it has some YAML-like structure
	// Must contain at least one "key: value" pattern
	lines := strings.Split(content, "\n")
	hasValidLine := false
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		
		// Check for key: value pattern
		if colonIndex := strings.Index(trimmed, ":"); colonIndex > 0 {
			key := strings.TrimSpace(trimmed[:colonIndex])
			// Basic check that key looks valid
			if key != "" && !strings.ContainsAny(key, "{}[]|><") {
				hasValidLine = true
				break
			}
		}
		
		// Also accept list items
		if strings.HasPrefix(trimmed, "- ") {
			hasValidLine = true
			break
		}
	}
	
	return hasValidLine
}