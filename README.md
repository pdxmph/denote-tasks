# denote-tasks

A task management system built on the Denote file naming convention.

## Features

- Pure Denote implementation
- Task and project management
- CLI and TUI interfaces
- 100% compatible with Emacs Denote

## Installation

```bash
go install github.com/pdxmph/denote-tasks/cmd/denote-tasks@latest
```

Or for a specific version:
```bash
go install github.com/pdxmph/denote-tasks/cmd/denote-tasks@v0.2.0
```

**Note**: This is a binary application, not a library. Do not use `go get` to install it.

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
notes_directory = "~/notes"
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