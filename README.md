# denote-tasks

A focused task management tool using the Denote file naming convention.

## Features

- **Task-focused** - Built specifically for task management, not general notes
- **Pure Denote** - Uses standard Denote file naming for compatibility
- **Project support** - Organize tasks by project with automatic linking
- **Dual interface** - Both CLI and TUI for different workflows
- **Emacs compatible** - Works seamlessly with Emacs Denote for task files

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
denote-tasks task new "Fix search bug"
denote-tasks task new -p p1 --due tomorrow "Call client"

# List tasks
denote-tasks task list
denote-tasks task list -p p1 --area work

# Update tasks (uses index_id from list)
denote-tasks task update -p p2 28
denote-tasks task done 28,35

# Add log entries
denote-tasks task log 28 "Found root cause"

# Interactive TUI
denote-tasks --tui
```

See [CLI Reference](docs/CLI_REFERENCE.md) for full command documentation.

## Configuration

Create `~/.config/denote-tasks/config.toml`:

```toml
notes_directory = "~/tasks"
editor = "vim"
default_area = "work"

[tui]
theme = "default"
```

## Documentation

- [Project Charter](PROJECT_CHARTER.md) - Vision and goals
- [Denote Task Specification](docs/DENOTE_TASK_SPEC.md) - File format (v2.0.0)
- [Architecture](docs/UNIFIED_ARCHITECTURE.md) - Technical design

## License

MIT License