package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/pdxmph/denote-tasks/internal/config"
	"github.com/pdxmph/denote-tasks/internal/denote"
	"github.com/pdxmph/denote-tasks/internal/task"
)

// fuzzyMatch performs fuzzy matching - query letters must appear in order but can be non-consecutive
func fuzzyMatch(text, pattern string) bool {
	if pattern == "" {
		return true
	}
	
	patternIdx := 0
	for _, ch := range text {
		if patternIdx < len(pattern) && ch == rune(pattern[patternIdx]) {
			patternIdx++
		}
	}
	
	return patternIdx == len(pattern)
}

// taskMatchesSearch performs fuzzy search on task metadata
func taskMatchesSearch(t *denote.Task, query string) bool {
	query = strings.ToLower(query)
	
	// Search in status
	if fuzzyMatch(strings.ToLower(t.Status), query) {
		return true
	}
	
	// Search in priority
	if fuzzyMatch(strings.ToLower(t.Priority), query) {
		return true
	}
	
	// Search in area
	if fuzzyMatch(strings.ToLower(t.Area), query) {
		return true
	}
	
	// Search in assignee
	if fuzzyMatch(strings.ToLower(t.Assignee), query) {
		return true
	}
	
	return false
}

// Run executes the CLI with the given arguments
func Run(cfg *config.Config, args []string) error {
	if len(args) == 0 {
		return showUsage()
	}

	command := args[0]
	commandArgs := args[1:]

	switch command {
	case "add":
		return addCommand(cfg, commandArgs)
	case "list":
		return listCommand(cfg, commandArgs)
	case "done":
		return doneCommand(cfg, commandArgs)
	case "show":
		return showCommand(cfg, commandArgs)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

// addCommand creates a new task
func addCommand(cfg *config.Config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: denote-tasks add <title>")
	}

	title := strings.Join(args, " ")
	
	// Create task with default tags
	createdTask, err := task.CreateTask(cfg.NotesDirectory, title, "", []string{}, "")
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	fmt.Printf("Created task %d: %s\n", createdTask.IndexID, createdTask.TaskMetadata.Title)
	return nil
}

// listCommand lists tasks
func listCommand(cfg *config.Config, args []string) error {
	scanner := denote.NewScanner(cfg.NotesDirectory)
	tasks, err := scanner.FindTasks()
	if err != nil {
		return fmt.Errorf("failed to find tasks: %w", err)
	}

	// Parse arguments for filters
	var area, status, search string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--area":
			if i+1 < len(args) {
				area = args[i+1]
				i++
			}
		case "--status":
			if i+1 < len(args) {
				status = args[i+1]
				i++
			}
		case "--search", "-s":
			if i+1 < len(args) {
				search = args[i+1]
				i++
			}
		}
	}

	// Apply filters
	if area != "" {
		tasks = denote.FilterTasks(tasks, "area", area)
	}
	if status != "" {
		tasks = denote.FilterTasks(tasks, "status", status)
	} else {
		// Default to open tasks
		tasks = denote.FilterTasks(tasks, "open", "")
	}
	
	// Apply search filter
	if search != "" {
		filtered := make([]*denote.Task, 0)
		for _, task := range tasks {
			if task.File.MatchesSearch(search) || taskMatchesSearch(task, search) {
				filtered = append(filtered, task)
			}
		}
		tasks = filtered
	}

	// Sort by due date (closest first)
	denote.SortTasks(tasks, "due", false)

	// Display tasks
	if len(tasks) == 0 {
		fmt.Println("No tasks found")
		return nil
	}

	fmt.Printf("%-4s %-8s %-8s %-10s %s\n", "ID", "Status", "Priority", "Due", "Title")
	fmt.Println(strings.Repeat("-", 60))

	for _, t := range tasks {
		due := ""
		if t.DueDate != "" {
			due = t.DueDate
			if denote.IsOverdue(t.DueDate) {
				due = "!" + due
			}
		}

		priority := t.Priority
		if priority == "" {
			priority = "-"
		}

		fmt.Printf("%-4d %-8s %-8s %-10s %s\n", 
			t.IndexID, t.Status, priority, due, t.TaskMetadata.Title)
	}

	return nil
}

// doneCommand marks a task as done
func doneCommand(cfg *config.Config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: denote-tasks done <task-id>")
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid task ID: %s", args[0])
	}

	// Find task
	t, err := task.FindTaskByID(cfg.NotesDirectory, id)
	if err != nil {
		return err
	}

	// Update status
	if err := denote.UpdateTaskStatus(t.Path, denote.TaskStatusDone); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	fmt.Printf("Marked task %d as done: %s\n", id, t.TaskMetadata.Title)
	return nil
}

// showCommand displays a specific task
func showCommand(cfg *config.Config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: denote-tasks show <task-id>")
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid task ID: %s", args[0])
	}

	// Find task
	t, err := task.FindTaskByID(cfg.NotesDirectory, id)
	if err != nil {
		return err
	}

	// Display task details
	fmt.Printf("Task %d: %s\n", t.IndexID, t.TaskMetadata.Title)
	fmt.Printf("Status: %s\n", t.Status)
	if t.Priority != "" {
		fmt.Printf("Priority: %s\n", t.Priority)
	}
	if t.DueDate != "" {
		fmt.Printf("Due: %s", t.DueDate)
		if denote.IsOverdue(t.DueDate) {
			fmt.Printf(" (OVERDUE)")
		}
		fmt.Println()
	}
	if t.Area != "" {
		fmt.Printf("Area: %s\n", t.Area)
	}
	if t.ProjectID != "" {
		fmt.Printf("Project: %s\n", t.ProjectID)
	}
	
	fmt.Printf("\nFile: %s\n", t.Path)

	return nil
}

// showUsage displays usage information
func showUsage() error {
	fmt.Fprintf(os.Stderr, `Usage: denote-tasks [--tui] [command] [args...]

Commands:
  add <title>          Create a new task
  list [options]       List tasks
    --area <area>      Filter by area
    --status <status>  Filter by status (default: open)
  done <id>           Mark task as done
  show <id>           Show task details

Options:
  --tui               Launch TUI interface
  --area <area>       Launch TUI with area filter
  --projects          Launch TUI showing projects list
  --version           Show version
  --config <path>     Use specific config file

When run without arguments, launches TUI interface.
`)
	return nil
}