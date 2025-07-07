package denote

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CreateNote creates a new Denote-formatted note file
func CreateNote(directory, title string, tags []string) (string, error) {
	// Validate title
	if title == "" {
		return "", fmt.Errorf("title cannot be empty")
	}
	
	// Generate Denote ID (timestamp)
	id := time.Now().Format("20060102T150405")
	
	// Convert title to slug
	slug := titleToSlug(title)
	if slug == "" {
		return "", fmt.Errorf("invalid title - could not generate slug")
	}
	
	// Join tags with underscores
	tagString := ""
	if len(tags) > 0 {
		tagString = "__" + strings.Join(tags, "_")
	}
	
	// Build filename
	filename := fmt.Sprintf("%s--%s%s.md", id, slug, tagString)
	filepath := filepath.Join(directory, filename)
	
	// Create note metadata
	metadata := NoteMetadata{
		Title:   title,
		Type:    "note",
		Created: time.Now().Format("2006-01-02"),
		Tags:    tags,
	}
	
	// Use the strict frontmatter writer
	content, err := WriteFrontmatterFile(metadata, "")
	if err != nil {
		return "", fmt.Errorf("failed to create frontmatter: %w", err)
	}
	
	// Write file
	if err := os.WriteFile(filepath, content, 0644); err != nil {
		return "", fmt.Errorf("failed to create note: %w", err)
	}
	
	return filepath, nil
}

// titleToSlug converts a title to a Denote-compatible slug
func titleToSlug(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)
	
	// Replace spaces and special characters with hyphens
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return '-'
	}, slug)
	
	// Remove multiple consecutive hyphens
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	
	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")
	
	return slug
}

// BuildDenoteFilename builds a Denote filename from components
func BuildDenoteFilename(id, slug string, tags []string) string {
	// Join tags with underscores
	tagString := ""
	if len(tags) > 0 {
		tagString = "__" + strings.Join(tags, "_")
	}
	
	// Build filename
	return fmt.Sprintf("%s--%s%s.md", id, slug, tagString)
}

// RenameFileForTags renames a Denote file to reflect new tags
func RenameFileForTags(oldPath string, newTags []string) (string, error) {
	// Parse the existing filename
	dir := filepath.Dir(oldPath)
	base := filepath.Base(oldPath)
	
	// Extract ID and slug from filename
	// Format: ID--slug__tag1_tag2.md
	parts := strings.SplitN(base, "--", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid Denote filename format")
	}
	
	id := parts[0]
	
	// Extract slug (before tags)
	remainingPart := strings.TrimSuffix(parts[1], ".md")
	slugParts := strings.SplitN(remainingPart, "__", 2)
	slug := slugParts[0]
	
	// Build new filename
	newFilename := BuildDenoteFilename(id, slug, newTags)
	newPath := filepath.Join(dir, newFilename)
	
	// Don't rename if the path hasn't changed
	if newPath == oldPath {
		return oldPath, nil
	}
	
	// Check if target already exists
	if _, err := os.Stat(newPath); err == nil {
		return "", fmt.Errorf("target file already exists: %s", newPath)
	}
	
	// Rename the file
	if err := os.Rename(oldPath, newPath); err != nil {
		return "", fmt.Errorf("failed to rename file: %w", err)
	}
	
	return newPath, nil
}