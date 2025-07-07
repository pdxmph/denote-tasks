package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/pdxmph/denote-tasks/internal/config"
	"github.com/pdxmph/denote-tasks/internal/tui"
)

func main() {
	// Use test config
	cfg, err := config.Load("test-config.toml")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Create a test project if it doesn't exist
	projectPath := filepath.Join(cfg.NotesDirectory, "20250113T120000--test-project__project.md")
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		content := `---
title: Test Project
type: project
status: active
priority: p2
area: test
due_date: 2025-01-20
tags: [test, demo]
---

# Test Project

This is a test project for debugging hotkeys.`

		if err := os.WriteFile(projectPath, []byte(content), 0644); err != nil {
			log.Fatal("Failed to create test project:", err)
		}
		fmt.Println("Created test project:", projectPath)
	}

	fmt.Println("\nProject View Hotkey Test")
	fmt.Println("========================")
	fmt.Println("1. Start TUI with: ./denote-tasks --config test-config.toml --tui")
	fmt.Println("2. Press 't' to enter Task Mode")
	fmt.Println("3. Press 'p' to show projects only")
	fmt.Println("4. Navigate to 'Test Project' and press Enter")
	fmt.Println("5. You should be on the Overview tab")
	fmt.Println("\nTest these hotkeys on the Overview tab:")
	fmt.Println("- p: Edit priority (should show prompt)")
	fmt.Println("- s: Edit status (should show prompt)")
	fmt.Println("- d: Edit due date (should show prompt)")
	fmt.Println("- a: Edit area (should show prompt)")
	fmt.Println("- g: Edit tags (should show prompt)")
	fmt.Println("\nPress Tab to switch to Tasks tab, then:")
	fmt.Println("- The above hotkeys should NOT work")
	fmt.Println("- Press 's' to see state menu instead")
	fmt.Println("\nPress Tab again to return to Overview tab")
	fmt.Println("- The hotkeys should work again")
}