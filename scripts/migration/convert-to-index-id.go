package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type OldCounterFile struct {
	NextTaskID    int    `json:"next_task_id"`
	NextProjectID int    `json:"next_project_id"`
	SpecVersion   string `json:"spec_version"`
}

type NewCounterFile struct {
	NextIndexID int    `json:"next_index_id"`
	SpecVersion string `json:"spec_version"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run convert-to-index-id.go <notes-directory>")
		os.Exit(1)
	}

	notesDir := os.Args[1]

	// Step 1: Convert all files to use index_id
	fmt.Println("=== Converting task_id/project_id to index_id ===")
	convertFiles(notesDir)

	// Step 2: Update counter file
	fmt.Println("\n=== Updating counter file ===")
	updateCounterFile(notesDir)

	fmt.Println("\nConversion completed!")
}

func convertFiles(notesDir string) {
	files, err := filepath.Glob(filepath.Join(notesDir, "*.md"))
	if err != nil {
		log.Fatal(err)
	}

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

		// Parse YAML as raw map to handle dynamic fields
		var fm map[string]interface{}
		if err := yaml.Unmarshal([]byte(parts[1]), &fm); err != nil {
			log.Printf("Error parsing frontmatter in %s: %v", file, err)
			continue
		}

		changed := false
		
		// Check for task_id field
		if taskID, ok := fm["task_id"]; ok {
			fm["index_id"] = taskID
			delete(fm, "task_id")
			changed = true
			fmt.Printf("Converting task_id to index_id in %s\n", filepath.Base(file))
		}
		
		// Check for project_id field (as integer for projects)
		if projectID, ok := fm["project_id"]; ok {
			// Only convert if it's an integer (project's own ID, not task's reference)
			switch v := projectID.(type) {
			case int:
				fm["index_id"] = v
				delete(fm, "project_id")
				changed = true
				fmt.Printf("Converting project_id to index_id in %s\n", filepath.Base(file))
			case float64:
				fm["index_id"] = int(v)
				delete(fm, "project_id")
				changed = true
				fmt.Printf("Converting project_id to index_id in %s\n", filepath.Base(file))
			}
			// If it's a string, it's a task's reference to a project, leave it alone
		}

		if changed {
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
			}
		}
	}
}

func updateCounterFile(notesDir string) {
	counterPath := filepath.Join(notesDir, ".denote-task-counter.json")
	
	// Read old counter file
	data, err := ioutil.ReadFile(counterPath)
	if err != nil {
		log.Printf("Error reading counter file: %v", err)
		return
	}

	var oldCounter OldCounterFile
	if err := json.Unmarshal(data, &oldCounter); err != nil {
		log.Printf("Error parsing counter file: %v", err)
		return
	}

	// Use the higher of the two values for the new index
	nextIndex := oldCounter.NextTaskID
	if oldCounter.NextProjectID > nextIndex {
		nextIndex = oldCounter.NextProjectID
	}

	// Create new counter structure
	newCounter := NewCounterFile{
		NextIndexID: nextIndex,
		SpecVersion: "2.0.1", // Bump patch version for this change
	}

	// Write new counter file
	newData, err := json.MarshalIndent(newCounter, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	if err := ioutil.WriteFile(counterPath, newData, 0644); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Updated counter file:\n")
	fmt.Printf("  Old: next_task_id=%d, next_project_id=%d\n", oldCounter.NextTaskID, oldCounter.NextProjectID)
	fmt.Printf("  New: next_index_id=%d\n", newCounter.NextIndexID)
}