# Project Charter: denote-tasks

Version: 1.0.0  
Date: 2025-01-13

## Project Vision

**denote-tasks** is a task management system built on the Denote file naming convention. It extends Denote with task-specific metadata while maintaining 100% compatibility with Emacs Denote.

### Core Philosophy

"Your tasks are just Denote files with extra metadata."

## Design Principles

1. **Denote Purity**
   - Every file follows the Denote naming convention
   - No support for other markdown formats
   - Emacs Denote can read all our files
   - We can read all Emacs Denote files

2. **Single Directory**
   - Notes and tasks coexist naturally
   - Organization through tags, not directories
   - No artificial separation

3. **Minimal Extension**
   - Tasks add YAML frontmatter to Denote files
   - Frontmatter is ignored by Emacs (appears as content)
   - No proprietary formats or lock-in

4. **Two Interfaces, One Tool**
   - CLI for quick operations
   - TUI for interactive work
   - Same filters, same operations

## What We Are Building

### Included Features

- **Denote Compliance**
  - Standard Denote filenames: `YYYYMMDDTHHMMSS--title__tags.md`
  - Denote identifier as primary key
  - Tag-based organization

- **Task Management**
  - Task creation with YAML frontmatter
  - Status tracking (open, done, paused, etc.)
  - Priority levels (p1, p2, p3)
  - Due dates and scheduling
  - Project associations using Denote IDs
  - Time estimates (Fibonacci)
  - Areas for life contexts

- **Project Management**
  - Projects are Denote files with project tag
  - Task-project associations via Denote IDs
  - Project status and timelines
  - Hierarchical project views

- **Dual Interface**
  - CLI: `denote-tasks add "Fix bug"`
  - TUI: `denote-tasks --tui`
  - Shared filtering system

### Explicitly Excluded

- ❌ TaskWarrior integration
- ❌ Plain markdown support  
- ❌ Multiple note formats
- ❌ Separate directories for tasks/notes
- ❌ Import/export from other systems
- ❌ Non-Denote file support

## Technical Architecture

```
denote-tasks/
├── cmd/
│   └── denote-tasks/
│       └── main.go          # Entry point
├── internal/
│   ├── denote/              # Denote operations
│   │   ├── id.go            # ID parsing/generation
│   │   ├── filename.go      # Filename parsing
│   │   └── scanner.go       # Directory scanning
│   ├── task/                # Task-specific logic
│   │   ├── task.go          # Task operations
│   │   ├── project.go       # Project operations
│   │   └── frontmatter.go   # YAML handling
│   ├── core/                # Business logic
│   │   ├── filter.go        # Filtering system
│   │   ├── sort.go          # Sorting logic
│   │   └── cache.go         # Metadata cache
│   ├── cli/                 # CLI interface
│   └── tui/                 # TUI interface
└── docs/
    └── DENOTE_TASK_SPEC.md  # Specification
```

## Data Format (Denote Task Spec v2.0.0)

### Task File Example

```
Filename: 20250113T143022--fix-search-bug__task_bug_urgent.md
```

```yaml
---
title: Fix search bug in project view
task_id: 234
type: task
status: open
priority: p1
due_date: 2025-01-15
project_id: 20250110T091533  # Project's Denote ID
area: work
---

The search function returns duplicate results when filtering by status.

## Steps to Reproduce
1. Open project view
2. Search for "open" tasks
3. See duplicates

[2025-01-13] Reported by user
```

### Project File Example

```
Filename: 20250110T091533--search-improvements__project.md
```

```yaml
---
title: Search Improvements Q1
project_id: 45
type: project  
status: active
priority: p2
due_date: 2025-03-31
area: work
---

Improve search functionality across the application.
```

## Key Differences from Legacy Tools

| Feature | notes-cli/tui | denote-tasks |
|---------|---------------|--------------|
| File formats | Multiple | Denote only |
| Directory structure | Separate tasks/notes | Single directory |
| TaskWarrior | Supported | Not supported |
| Note types | Various markdown | Denote files only |
| Project references | String matching | Denote IDs |
| Complexity | High | Low |

## Development Approach

### Phase 1: Core (Week 1-2)
- Denote file operations
- Task/project data structures
- YAML frontmatter parsing
- Spec v2.0.0 compliance

### Phase 2: Business Logic (Week 3-4)
- Filter system
- Sort operations  
- Cache layer
- Project associations

### Phase 3: Interfaces (Week 5-6)
- CLI commands
- TUI views
- Shared argument parsing
- Configuration

### Phase 4: Polish (Week 7-8)
- Testing
- Documentation
- Performance optimization
- Release preparation

## Configuration

Minimal configuration in `~/.config/denote-tasks/config.toml`:

```toml
# Required
notes_directory = "~/notes"

# Optional
editor = "nvim"
default_area = "work"

[tui]
theme = "default"
```

## Success Criteria

1. **Emacs Compatibility**: Can exchange files with Emacs Denote
2. **Simplicity**: <5000 LOC total
3. **Performance**: Instant operations on 10k+ files
4. **Reliability**: Zero data corruption
5. **Usability**: Intuitive for Denote users

## Migration from Legacy Tools

For users of notes-cli/tui:

1. **File Format**: Already compatible if using Denote
2. **Project References**: Need migration to Denote IDs
3. **Directory**: Move tasks to notes directory
4. **Config**: Simplified configuration

Provide migration tool:
```bash
denote-tasks migrate --from notes-cli ~/old-tasks
```

## Design Decisions

1. **Why Denote Only?**
   - Simplicity and focus
   - Guaranteed uniqueness
   - Emacs ecosystem compatibility
   - Immutable identifiers

2. **Why No TaskWarrior?**
   - Different philosophies
   - Added complexity
   - Denote is sufficient

3. **Why Single Directory?**
   - Natural organization
   - Simpler mental model
   - Better for sync

4. **Why YAML Frontmatter?**
   - Human readable
   - Well-supported
   - Emacs ignores it

## Future Possibilities

Once core is solid:

- Web interface
- Mobile app
- Org-mode export
- CalDAV sync
- Git integration

But always maintaining Denote purity.

## Getting Started

```bash
# Create new task
denote-tasks add "Write project charter"

# View tasks
denote-tasks list

# Interactive mode
denote-tasks --tui

# Filter tasks
denote-tasks list --area work --status open
```

## License

MIT License - Keep it simple and open.

## Summary

denote-tasks is a focused, opinionated task manager that embraces Denote's philosophy of stable, unique identifiers. By constraining ourselves to Denote files only, we gain simplicity, reliability, and perfect interoperability with the Emacs ecosystem.