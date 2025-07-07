package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type CounterFile struct {
	NextTaskID    int    `json:"next_task_id"`
	NextProjectID int    `json:"next_project_id"`
	SpecVersion   string `json:"spec_version"`
}

type Frontmatter struct {
	ID              string      `yaml:"id,omitempty"`
	Title           string      `yaml:"title"`
	TaskID          int         `yaml:"task_id,omitempty"`
	Type            string      `yaml:"type,omitempty"`
	Date            string      `yaml:"date,omitempty"`
	Tags            []string    `yaml:"tags,omitempty"`
	Status          string      `yaml:"status,omitempty"`
	Area            string      `yaml:"area,omitempty"`
	Project         string      `yaml:"project,omitempty"`       // Old field to be migrated
	ProjectIDRaw    interface{} `yaml:"project_id,omitempty"`   // Can be int or string
	Priority        string      `yaml:"priority,omitempty"`
	DueDate         string      `yaml:"due_date,omitempty"`
	StartDate       string      `yaml:"start_date,omitempty"`
	Estimate        int         `yaml:"estimate,omitempty"`
	Assignee        string      `yaml:"assignee,omitempty"`
}

// Helper methods to get/set project_id properly
func (fm *Frontmatter) GetProjectIDInt() int {
	if v, ok := fm.ProjectIDRaw.(int); ok {
		return v
	}
	return 0
}

func (fm *Frontmatter) GetProjectIDString() string {
	if v, ok := fm.ProjectIDRaw.(string); ok {
		return v
	}
	return ""
}

func (fm *Frontmatter) SetProjectIDInt(id int) {
	fm.ProjectIDRaw = id
}

func (fm *Frontmatter) SetProjectIDString(id string) {
	fm.ProjectIDRaw = id
}

type ProjectInfo struct {
	OldID      string
	NewID      string
	Name       string
	Filename   string
	SequenceID int
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run fix-test-data.go <notes-directory>")
		os.Exit(1)
	}

	notesDir := os.Args[1]

	// Step 1: Fix duplicate timestamps in project files
	fmt.Println("=== Step 1: Fixing duplicate timestamps in project files ===")
	projectMap := fixProjectTimestamps(notesDir)

	// Step 2: Update task files with project_id fields
	fmt.Println("\n=== Step 2: Converting project: to project_id: in task files ===")
	taskCount := updateTaskProjectReferences(notesDir, projectMap)

	// Step 3: Add sequential IDs to all files
	fmt.Println("\n=== Step 3: Adding sequential IDs to all files ===")
	nextTaskID, nextProjectID := addSequentialIDs(notesDir)

	// Step 4: Create counter file
	fmt.Println("\n=== Step 4: Creating .denote-task-counter.json ===")
	createCounterFile(notesDir, nextTaskID, nextProjectID)

	fmt.Printf("\nCompleted! Fixed %d projects and %d tasks.\n", len(projectMap), taskCount)
}

func fixProjectTimestamps(notesDir string) map[string]*ProjectInfo {
	projectMap := make(map[string]*ProjectInfo)
	var projects []*ProjectInfo

	// Find all project files
	projectFiles, err := filepath.Glob(filepath.Join(notesDir, "*__project*.md"))
	if err != nil {
		log.Fatal(err)
	}

	// Sort to ensure consistent ordering
	sort.Strings(projectFiles)

	// Parse each project file
	for _, file := range projectFiles {
		basename := filepath.Base(file)
		parts := strings.SplitN(basename, "--", 2)
		if len(parts) < 2 {
			continue
		}

		oldID := parts[0]
		rest := parts[1]
		titlePart := strings.SplitN(rest, "__", 2)[0]

		info := &ProjectInfo{
			OldID:    oldID,
			Name:     titlePart,
			Filename: file,
		}
		projects = append(projects, info)
	}

	// Generate new unique timestamps and assign sequence IDs
	baseTime, _ := time.Parse("20060102T150405", "20250704T120000")
	for i, proj := range projects {
		newTime := baseTime.Add(time.Duration(i) * time.Second)
		proj.NewID = newTime.Format("20060102T150405")
		proj.SequenceID = i + 1  // Assign sequential ID starting from 1
		
		// Map both the title and various variants
		projectMap[proj.Name] = proj
		projectMap[strings.ReplaceAll(proj.Name, "-", "_")] = proj
		projectMap[strings.ReplaceAll(proj.Name, "_", "-")] = proj
		
		// Store in cache for later lookup
		projectInfoCache[proj.NewID] = proj
		
		fmt.Printf("Project: %s\n  Old ID: %s\n  New ID: %s\n  Sequence ID: %d\n", proj.Name, proj.OldID, proj.NewID, proj.SequenceID)
	}

	// Rename files with new timestamps
	for _, proj := range projects {
		oldPath := proj.Filename
		newFilename := strings.Replace(filepath.Base(oldPath), proj.OldID, proj.NewID, 1)
		newPath := filepath.Join(filepath.Dir(oldPath), newFilename)
		
		if oldPath != newPath {
			if err := os.Rename(oldPath, newPath); err != nil {
				log.Printf("Error renaming %s: %v", oldPath, err)
			} else {
				fmt.Printf("  Renamed: %s -> %s\n", filepath.Base(oldPath), filepath.Base(newPath))
				proj.Filename = newPath
			}
		}

		// Update the ID in the frontmatter
		updateFrontmatterID(newPath, proj.NewID)
	}

	return projectMap
}

func updateTaskProjectReferences(notesDir string, projectMap map[string]*ProjectInfo) int {
	taskFiles, err := filepath.Glob(filepath.Join(notesDir, "*__task*.md"))
	if err != nil {
		log.Fatal(err)
	}

	count := 0
	for _, file := range taskFiles {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			log.Printf("Error reading %s: %v", file, err)
			continue
		}

		// Parse frontmatter
		parts := strings.SplitN(string(content), "---", 3)
		if len(parts) < 3 {
			continue
		}

		var fm Frontmatter
		if err := yaml.Unmarshal([]byte(parts[1]), &fm); err != nil {
			log.Printf("Error parsing frontmatter in %s: %v", file, err)
			continue
		}

		// If there's a project field, convert it
		if fm.Project != "" && fm.GetProjectIDString() == "" {
			if proj, ok := projectMap[fm.Project]; ok {
				fm.SetProjectIDString(proj.NewID)
				count++
				
				// Marshal back to YAML
				newFM, err := yaml.Marshal(&fm)
				if err != nil {
					log.Printf("Error marshaling frontmatter: %v", err)
					continue
				}

				// Rebuild file content
				newContent := fmt.Sprintf("---\n%s---\n%s", string(newFM), parts[2])
				
				// Write back
				if err := ioutil.WriteFile(file, []byte(newContent), 0644); err != nil {
					log.Printf("Error writing %s: %v", file, err)
				} else {
					fmt.Printf("Updated task: %s (project: %s -> project_id: %s)\n", 
						filepath.Base(file), fm.Project, proj.NewID)
				}
			} else {
				log.Printf("Warning: Unknown project '%s' in %s", fm.Project, file)
			}
		}
	}

	return count
}

func addSequentialIDs(notesDir string) (int, int) {
	// Get all markdown files
	files, err := filepath.Glob(filepath.Join(notesDir, "*.md"))
	if err != nil {
		log.Fatal(err)
	}

	nextTaskID := 1
	nextProjectID := 1
	maxTaskID := 0
	maxProjectID := 0

	// Sort files by timestamp for consistent ordering
	sort.Strings(files)

	for _, file := range files {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			log.Printf("Error reading %s: %v", file, err)
			continue
		}

		// Parse frontmatter
		parts := strings.SplitN(string(content), "---", 3)
		if len(parts) < 3 {
			continue
		}

		var fm Frontmatter
		if err := yaml.Unmarshal([]byte(parts[1]), &fm); err != nil {
			log.Printf("Error parsing frontmatter in %s: %v", file, err)
			continue
		}

		// Determine if it's a task or project
		isTask := strings.Contains(file, "__task")
		isProject := strings.Contains(file, "__project")

		// Assign sequential ID if missing
		if isTask {
			if fm.TaskID == 0 {
				fm.TaskID = nextTaskID
				nextTaskID++
			}
			if fm.TaskID > maxTaskID {
				maxTaskID = fm.TaskID
			}
			fm.Type = "task"
		} else if isProject {
			if fm.GetProjectIDInt() == 0 {
				// Look up the project in our cache by its ID
				if proj, ok := projectInfoCache[fm.ID]; ok {
					fm.SetProjectIDInt(proj.SequenceID)
				} else {
					fm.SetProjectIDInt(nextProjectID)
					nextProjectID++
				}
			}
			projID := fm.GetProjectIDInt()
			if projID > maxProjectID {
				maxProjectID = projID
			}
			fm.Type = "project"
		}

		// Marshal back to YAML
		newFM, err := yaml.Marshal(&fm)
		if err != nil {
			log.Printf("Error marshaling frontmatter: %v", err)
			continue
		}

		// Rebuild file content
		newContent := fmt.Sprintf("---\n%s---\n%s", string(newFM), parts[2])
		
		// Write back
		if err := ioutil.WriteFile(file, []byte(newContent), 0644); err != nil {
			log.Printf("Error writing %s: %v", file, err)
		} else {
			if isTask {
				fmt.Printf("Added task_id %d to %s\n", fm.TaskID, filepath.Base(file))
			} else if isProject {
				fmt.Printf("Added project_id %d to %s\n", fm.GetProjectIDInt(), filepath.Base(file))
			}
		}
	}

	// Return the next available IDs
	return maxTaskID + 1, maxProjectID + 1
}

func createCounterFile(notesDir string, nextTaskID, nextProjectID int) {
	counter := CounterFile{
		NextTaskID:    nextTaskID,
		NextProjectID: nextProjectID,
		SpecVersion:   "2.0.0",
	}

	data, err := json.MarshalIndent(counter, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	counterPath := filepath.Join(notesDir, ".denote-task-counter.json")
	if err := ioutil.WriteFile(counterPath, data, 0644); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created %s\n", counterPath)
	fmt.Printf("  Next task ID: %d\n", nextTaskID)
	fmt.Printf("  Next project ID: %d\n", nextProjectID)
}

func updateFrontmatterID(file, newID string) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Printf("Error reading %s: %v", file, err)
		return
	}

	// Parse frontmatter
	parts := strings.SplitN(string(content), "---", 3)
	if len(parts) < 3 {
		return
	}

	var fm Frontmatter
	if err := yaml.Unmarshal([]byte(parts[1]), &fm); err != nil {
		log.Printf("Error parsing frontmatter in %s: %v", file, err)
		return
	}

	fm.ID = newID

	// Marshal back to YAML
	newFM, err := yaml.Marshal(&fm)
	if err != nil {
		log.Printf("Error marshaling frontmatter: %v", err)
		return
	}

	// Rebuild file content
	newContent := fmt.Sprintf("---\n%s---\n%s", string(newFM), parts[2])
	
	// Write back
	if err := ioutil.WriteFile(file, []byte(newContent), 0644); err != nil {
		log.Printf("Error writing %s: %v", file, err)
	}
}

var projectInfoCache map[string]*ProjectInfo

func findProjectByNewID(id string) (*ProjectInfo, bool) {
	if projectInfoCache == nil {
		return nil, false
	}
	for _, p := range projectInfoCache {
		if p.NewID == id {
			return p, true
		}
	}
	return nil, false
}

func init() {
	projectInfoCache = make(map[string]*ProjectInfo)
}