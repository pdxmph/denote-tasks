package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Generic frontmatter that can handle both tasks and projects
type Frontmatter struct {
	ID        string   `yaml:"id,omitempty"`
	Identifier string   `yaml:"identifier,omitempty"` 
	TaskID    int      `yaml:"task_id,omitempty"`
	ProjectID int      `yaml:"project_id,omitempty"`
	Title     string   `yaml:"title"`
	Type      string   `yaml:"type,omitempty"`
	Date      string   `yaml:"date"`
	Tags      []string `yaml:"tags"`
	Status    string   `yaml:"status"`
	DueDate   string   `yaml:"due_date,omitempty"`
	StartDate string   `yaml:"start_date,omitempty"`
	Area      string   `yaml:"area,omitempty"`
	Project   string   `yaml:"project,omitempty"`
	ProjectRef string  `yaml:"project_id,omitempty"`
	Priority  string   `yaml:"priority,omitempty"`
	Estimate  int      `yaml:"estimate,omitempty"`
	Assignee  string   `yaml:"assignee,omitempty"`
}

// Map of project names to their Denote IDs
var projectMap = map[string]string{
	"teammtg":         "20250704T124525",
	"lyon_planning":   "20250704T124525", // lyon-planning project
	"apex":            "20250704T124524",
	"roadmap":         "20250704T124525",
	"sunsetting":      "20250704T124524",
	"clockwise":       "20250704T124525",
	"otter":           "20250704T124525",
	"postfest":        "20250704T124525",
	"securityreviews": "20250704T124525",
	"crowdstrike":     "20250704T124524",
	"oncall":          "20250704T124524",
	"it":              "20250704T124524",
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run normalize-tasks.go <directory>")
		os.Exit(1)
	}

	dir := os.Args[1]
	
	// First pass: build project map from actual project files
	fmt.Println("Building project map...")
	buildProjectMap(dir)
	
	// Second pass: normalize all files
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".md") {
			fmt.Printf("Processing: %s\n", filepath.Base(path))
			if err := normalizeFile(path); err != nil {
				log.Printf("Error processing %s: %v\n", path, err)
			}
		}
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
}

func buildProjectMap(dir string) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || !strings.HasSuffix(path, ".md") {
			return err
		}
		
		filename := filepath.Base(path)
		if strings.Contains(filename, "__project") {
			// Extract Denote ID from filename
			parts := strings.Split(filename, "--")
			if len(parts) >= 1 {
				denoteID := parts[0]
				
				// Read file to get project name
				content, err := os.ReadFile(path)
				if err == nil {
					var fm Frontmatter
					if err := parseYAML(content, &fm); err == nil && fm.Title != "" {
						// Create various possible project name formats
						projectName := strings.ToLower(fm.Title)
						projectName = strings.ReplaceAll(projectName, " ", "_")
						projectMap[projectName] = denoteID
						
						// Also try with hyphens
						projectNameHyphen := strings.ReplaceAll(strings.ToLower(fm.Title), " ", "-")
						projectMap[projectNameHyphen] = denoteID
						
						// And the original title
						projectMap[strings.ToLower(fm.Title)] = denoteID
					}
				}
			}
		}
		return nil
	})
	
	fmt.Printf("Found %d projects in map\n", len(projectMap))
}

func normalizeFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var fm Frontmatter
	body, err := parseYAMLWithBody(content, &fm)
	if err != nil {
		return err
	}

	// Determine if this is a task or project based on tags
	filename := filepath.Base(path)
	isTask := strings.Contains(filename, "__task")
	isProject := strings.Contains(filename, "__project")

	// Normalize the frontmatter
	
	// Set type field if missing
	if fm.Type == "" {
		if isTask {
			fm.Type = "task"
		} else if isProject {
			fm.Type = "project"
		}
	}
	
	// Use identifier field if id field is present
	if fm.ID != "" && fm.Identifier == "" {
		fm.Identifier = fm.ID
	}
	
	// Convert project name to project_id for tasks
	if isTask && fm.Project != "" && fm.ProjectRef == "" {
		if projectID, ok := projectMap[fm.Project]; ok {
			fm.ProjectRef = projectID
			fm.Project = "" // Clear the old field
		} else {
			fmt.Printf("  Warning: Unknown project '%s' in %s\n", fm.Project, filepath.Base(path))
		}
	}

	// Write back the normalized file
	output, err := generateYAML(&fm)
	if err != nil {
		return err
	}

	normalized := string(output) + "\n" + body
	return os.WriteFile(path, []byte(normalized), 0644)
}

func parseYAML(content []byte, v interface{}) error {
	decoder := yaml.NewDecoder(strings.NewReader(string(content)))
	decoder.KnownFields(false) // Allow unknown fields
	return decoder.Decode(v)
}

func parseYAMLWithBody(content []byte, v interface{}) (string, error) {
	parts := strings.SplitN(string(content), "---", 3)
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid frontmatter format")
	}

	yamlContent := parts[1]
	body := "---" + parts[2]

	decoder := yaml.NewDecoder(strings.NewReader(yamlContent))
	decoder.KnownFields(false)
	err := decoder.Decode(v)
	
	return body, err
}

func generateYAML(fm *Frontmatter) ([]byte, error) {
	var b strings.Builder
	b.WriteString("---\n")
	
	encoder := yaml.NewEncoder(&b)
	encoder.SetIndent(2)
	err := encoder.Encode(fm)
	if err != nil {
		return nil, err
	}
	
	b.WriteString("---")
	return []byte(b.String()), nil
}