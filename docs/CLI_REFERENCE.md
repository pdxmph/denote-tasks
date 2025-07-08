# denote-tasks CLI Reference

## Overview

The denote-tasks CLI uses an entity-first command structure:

```
denote-tasks <entity> <action> [options] [arguments]
```

**Important**: Task operations use `index_id` values from the files, not display positions. When you see task "28" in the list, you use 28 in commands.

## Global Options

These options can be used with any command:

- `--config PATH` - Use specific config file
- `--dir PATH` - Override notes directory  
- `--tui, -t` - Launch TUI interface (not yet implemented)
- `--json` - Output in JSON format (not yet implemented)
- `--no-color` - Disable color output
- `--quiet, -q` - Minimal output

## Task Commands

### task new

Create a new task.

```bash
denote-tasks task new [options] <title>
```

Options:
- `-p, --priority` - Set priority (p1, p2, p3)
- `--due` - Set due date (YYYY-MM-DD or natural language)
- `--area` - Set task area
- `--project` - Set project ID
- `--estimate` - Set time estimate
- `--tags` - Comma-separated tags

Examples:
```bash
denote-tasks task new "Review budget proposal"
denote-tasks task new -p p1 --due tomorrow "Call client"
denote-tasks task new --area work --project 20240315T093000 "Update docs"
```

### task list

List tasks with filtering and sorting.

```bash
denote-tasks task list [options]
```

Options:
- `-a, --all` - Show all tasks (default: open only)
- `--area` - Filter by area
- `--status` - Filter by status
- `-p, --priority` - Filter by priority (p1, p2, p3)
- `--project` - Filter by project ID
- `--overdue` - Show only overdue tasks
- `--soon` - Show tasks due soon
- `-s, --sort` - Sort by: modified (default), priority, due, created
- `-r, --reverse` - Reverse sort order

Examples:
```bash
denote-tasks task list                    # List open tasks
denote-tasks task list --all              # List all tasks
denote-tasks task list -p p1              # List only p1 tasks
denote-tasks task list --area work        # List work tasks
denote-tasks task list --overdue          # List overdue tasks
denote-tasks task list --sort priority    # Sort by priority
```

### task update

Update task metadata. **Note**: Options must come before task IDs.

```bash
denote-tasks task update [options] <task-ids>
```

Options:
- `-p, --priority` - Set priority (p1, p2, p3)
- `--due` - Set due date
- `--area` - Set area
- `--project` - Set project ID
- `--estimate` - Set time estimate
- `--status` - Set status (open, done, paused, delegated, dropped)

Task IDs support:
- Single: `28`
- List: `28,35,61`
- Range: `28-35`
- Mixed: `28,35-40,61`

Examples:
```bash
denote-tasks task update -p p2 28                # Change priority
denote-tasks task update --due "next week" 35    # Set due date
denote-tasks task update --status paused 28,35   # Pause multiple tasks
denote-tasks task update --area personal 10-15   # Update area for range
```

### task done

Mark tasks as done.

```bash
denote-tasks task done <task-ids>
```

Examples:
```bash
denote-tasks task done 28           # Mark single task as done
denote-tasks task done 28,35,61     # Mark multiple tasks as done
denote-tasks task done 10-15        # Mark range as done
```

### task log

Add a timestamped log entry to a task.

```bash
denote-tasks task log <task-id> <message>
```

Examples:
```bash
denote-tasks task log 28 "Discussed with team, waiting for feedback"
denote-tasks task log 35 "Completed first draft"
```

### task edit (not implemented)

Edit task in external editor or TUI.

```bash
denote-tasks task edit <task-id>
```

### task delete (not implemented)

Delete tasks with confirmation.

```bash
denote-tasks task delete <task-ids>
```

## Project Commands (not implemented)

```bash
denote-tasks project new <title>
denote-tasks project list [options]
denote-tasks project update [options] <project-ids>
denote-tasks project archive <project-ids>
```

## Note Commands (not implemented)

```bash
denote-tasks note new <title>
denote-tasks note list [options]
denote-tasks note search <query>
denote-tasks note tag [options] <note-ids>
```

## Legacy Command Aliases

For compatibility with the old CLI:

- `denote-tasks add` → `denote-tasks task new`
- `denote-tasks list` → `denote-tasks task list`
- `denote-tasks done` → `denote-tasks task done`

## Examples

### Daily workflow

```bash
# Check what's due today
denote-tasks task list --overdue
denote-tasks task list --soon

# Add a new urgent task
denote-tasks task new -p p1 --due today "Fix critical bug"

# Update task priority after meeting
denote-tasks task update -p p2 28

# Log progress
denote-tasks task log 28 "Found root cause, working on fix"

# Mark completed
denote-tasks task done 28
```

### Bulk operations

```bash
# Update multiple tasks to a project
denote-tasks task update --project 20240315T093000 28,35,61

# Mark a range as done
denote-tasks task done 10-15

# Change area for all personal tasks
denote-tasks task list --area personal  # See the IDs
denote-tasks task update --area home 4,7,12-15,23
```

## Tips

1. **Index IDs are stable**: The number shown (e.g., 28) is the task's permanent ID, not its position in the list.

2. **Flags before IDs**: Due to Go's flag parsing, options must come before task IDs:
   - ✓ `denote-tasks task update -p p1 28`
   - ✗ `denote-tasks task update 28 -p p1`

3. **Natural date parsing**: The `--due` flag accepts natural language:
   - `today`, `tomorrow`, `next week`
   - `monday`, `next friday`
   - `2025-01-15` (ISO format)

4. **Filtering is additive**: Multiple filters work together:
   ```bash
   denote-tasks task list -p p1 --area work --soon
   ```
   Shows only p1 work tasks due soon.