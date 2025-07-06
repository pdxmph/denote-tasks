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

## Usage

```bash
# Add a task
denote-tasks add "Fix search bug"

# List tasks  
denote-tasks list --area work

# Interactive TUI
denote-tasks --tui
```

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