# denote-tasks

A focused task management tool using the Denote file naming convention for portability. This project has no association with the Denote project.

## Important consideration before using this code or interacting with this codebase

This application is an experiment in using Claude Code as the primary driver the development of a small, focused app that concerns itself with the owner's particular point of view on the task it is accomplishing.

As such, this is not meant to be what people think of as "an open source project," because I don't have a commitment to building a community around it and don't have the bandwidth to maintain it beyond "fix bugs I find in the process of pushing it in a direction that works for me."

It's important to understand this for a few reasons:

1. If you use this code, you'll be using something largely written by an LLM with all the things we know this entails in 2025: Potential inefficiency, security risks, and the risk of data loss.

2. If you use this code, you'll be using something that works for me the way I would like it to work. If it doesn't do what you want it to do, or if it fails in some way particular to your preferred environment, tools, or use cases, your best option is to take advantage of its very liberal license and fork it.

3. I'll make a best effort to only tag the codebase when it is in a working state with no bugs that functional testing has revealed.

While I appreciate and applaud assorted efforts to certify code and projects AI-free, I think it's also helpful to post commentary like this up front: Yes, this was largely written by an LLM so treat it accordingly. Don't think of it like code you can engage with, think of it like someone's take on how to do a task or solve a problem.

That said:

## What it is

`denote-tasks` is a specialized tool for managing tasks and projects using plain text files with the Denote naming convention. Each task is a Markdown file with YAML frontmatter containing task metadata (priority, due date, project assignment, etc.).

## What it isn't

- **Not a general note-taking app** - Use Denote in Emacs or other tools for that
- **Not a calendar** - Though it tracks due dates
- **Not a time tracker** - Though it supports time estimates
- **For most people?** - It's very aligned with one particular person's idea of just how they wanted CLI/TUI task management to work.

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
- `Enter` - Open task/project details

**Actions (lowercase):**

- `c` - Create new task or project
- `d` - Edit due date
- `l` - Add log entry (tasks only)
- `r` - Toggle sort order
- `s` - Change task state
- `t` - Edit tags
- `u` - Update task metadata
- `x` - Delete task/project
- `/` - Search (use `#tag` for tag search)

**Priority:**

- `0` - Clear priority
- `1/2/3` - Set priority (p1/p2/p3)

**Filters & Views (uppercase):**

- `E` - Edit in external editor
- `P` - Toggle projects view
- `T` - Toggle tasks view
- `S` - Sort options menu
- `f` - Filter menu (area/priority/state/soon)

**General:**

- `?` - Help screen
- `q` - Quit

See [CLI Reference](docs/CLI_REFERENCE.md) for full command documentation.

## Configuration

Create `~/.config/denote-tasks/config.toml`:

```toml
notes_directory = "~/tasks"  # Where task files live (kept for backward compatibility)
editor = "vim"              # External editor for 'E' command
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
