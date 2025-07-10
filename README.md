# denote-tasks

A focused task management tool using the Denote file naming convention.

## What it is

`denote-tasks` is a specialized tool for managing tasks and projects using plain text files with the Denote naming convention. Each task is a Markdown file with YAML frontmatter containing task metadata (priority, due date, project assignment, etc.).

## What it isn't

- **Not a general note-taking app** - Use Denote in Emacs or other tools for that
- **Not a calendar** - Though it tracks due dates
- **Not a time tracker** - Though it supports time estimates

## Quick Start

```bash
# Install
go install github.com/pdxmph/denote-tasks@latest

# Create your first task
denote-tasks new -p p1 --due tomorrow "Review pull request"

# Launch the TUI
denote-tasks --tui

# Filter by area in TUI
denote-tasks --tui --area work
```

## Features

- **Task-focused** - Built specifically for task management, not general notes
- **Works with Denote** - Uses standard Denote file naming for compatibility
- **Project support** - Organize tasks by project with automatic linking
- **Dual interface** - Both CLI and TUI for different workflows

## Installation

```bash
go install github.com/pdxmph/denote-tasks@latest
```

Or for a specific version:

```bash
go install github.com/pdxmph/denote-tasks@v0.2.0
```

## Usage

```bash
# Create a new task
denote-tasks new "Fix search bug"
denote-tasks new -p p1 --due tomorrow "Call client"

# List tasks
denote-tasks list
denote-tasks list -p p1 --area work

# Update tasks (uses index_id from list)
denote-tasks update -p p2 28
denote-tasks done 28,35

# Add log entries
denote-tasks log 28 "Found root cause"

# Interactive TUI
denote-tasks --tui
denote-tasks --tui --area work  # Start filtered by area

# Project management
denote-tasks project new "Q1 Planning"
denote-tasks project list
denote-tasks project tasks 15  # Show tasks for project
```

### TUI Hotkeys

**Navigation:**

- `j/k` or `↓/↑` - Move down/up
- `g g` - Go to top
- `G` - Go to bottom
- `p` - Switch to projects view
- `t` - Switch to tasks view

**Task Actions:**

- `Enter` - Preview/view item
- `c` - Create new task or project
- `0` - Clear priority
- `1/2/3` - Set task priority (p1/p2/p3)
- `s` - Change task state menu
- `x` - Delete task/project
- `e` - Edit in external editor
- `u` - Update task metadata
- `l` - Add log entry to task

**Organization:**

- `/` - Search (use `#tag` for tag search)
- `f` - Filter menu (area/priority/state/soon)
- `S` - Sort options menu
- `r` - Toggle sort order

**General:**

- `?` - Help screen
- `q` - Quit

See [CLI Reference](docs/CLI_REFERENCE.md) for full command documentation.

## Configuration

Create `~/.config/denote-tasks/config.toml`:

```toml
notes_directory = "~/tasks"  # Where task files live (kept for backward compatibility)
editor = "vim"              # External editor for 'e' command
default_area = "work"       # Default area for new tasks
soon_horizon = 3            # Days ahead for "soon" filter

[tui]
theme = "default"           # UI theme

[tasks]
sort_by = "due"             # Default sort: due, priority, project, title, created
sort_order = "normal"       # normal or reverse
```

## Documentation

- [Project Charter](PROJECT_CHARTER.md) - Vision and goals
- [Denote Task Specification](docs/DENOTE_TASK_SPEC.md) - File format (v2.0.0)
- [Architecture](docs/UNIFIED_ARCHITECTURE.md) - Technical design

## Task File Format

Tasks are stored as markdown files with Denote naming:

```
20240315T093000--fix-search-bug__task_p1_work.md
┗──────────────┘ ┗─────────────┘┗───────────┘
   Denote ID       Title slug        Tags
```

With YAML frontmatter:

```yaml
---
title: Fix search bug in task list
index_id: 28
type: task
status: open
priority: p1
due_date: 2024-03-16
project_id: 20240301T100000
area: work
---
## Description
Search is not filtering tasks correctly when...
```

## License

MIT License
