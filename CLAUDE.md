# CLAUDE.md - Project Context for denote-tasks

This file contains important context about the denote-tasks project to help AI assistants understand the codebase, architecture decisions, and current state.

## Project Overview

**denote-tasks** is a focused task management tool built on the Denote file naming convention. It uses Denote format for consistent file identification and backward compatibility while providing powerful task and project management features.

### Current Status (2025-01-14)

- Basic CLI working (task and project commands)
- TUI implementation for tasks (needs testing)
- Following Denote Task Spec v2.0.0
- Refactoring to task-only focus in progress

## Important Documents

**ALWAYS READ THESE FIRST:**
- `REFACTORING_PLAN.md` - Current refactoring to task-only focus
- `PROJECT_CHARTER.md` - Vision, principles, and roadmap
- `docs/DENOTE_TASK_SPEC.md` - File format specification (v2.0.0)
- `PROGRESS.md` - Current session status

## Architecture Principles

1. **Denote Format** - Use Denote naming for consistent IDs
2. **Task Focus** - Only task and project files, no general notes
3. **Two Interfaces** - CLI and TUI for task management
4. **No External Dependencies** - No TaskWarrior, etc.
5. **Simplicity** - Focused functionality over feature creep

## Testing Guidelines

### CRITICAL RULE: NEVER MARK FEATURES AS COMPLETE WITHOUT HUMAN TESTING

**STOP AND READ:** Any feature implementation MUST be marked as "IMPLEMENTED BUT NOT TESTED" until the human has confirmed it works. This includes:
- Never marking issues as "✅ Completed" without human confirmation
- Never updating todo lists to "completed" for untested features
- Always use phrases like "implemented but needs testing" or "code complete, awaiting manual testing"
- NEVER assume code that compiles successfully actually works

### For TUI Development

**IMPORTANT FOR AI ASSISTANTS:** It is IMPOSSIBLE to test TUI applications in this environment. NEVER attempt to run or test the TUI. Instead, always ask the user to test the features and provide feedback. TUI applications require an interactive terminal which is not available in this context.

Since TUI applications can't be tested in this environment:

1. **Implement features completely** before declaring done
2. **Document what needs manual testing** in PROGRESS.md
3. **Create test configurations** (never modify user configs)
4. **List specific test cases** for human testing
5. **Ask the user to test** rather than attempting to test yourself

### Manual Testing Checklist

When implementing TUI features, provide:
- Step-by-step testing instructions
- Expected behavior for each feature
- Edge cases to verify
- Sample test data if needed

## Current Implementation Status

### Completed
- ✅ Core denote package (parser, scanner, types)
- ✅ Basic CLI task commands
- ✅ Configuration system
- ✅ Task creation with frontmatter
- ✅ Project file support

### In Progress
- 🚧 TUI for task management
  - Task and project views implemented
  - Needs manual testing
  - Being refactored for task-only focus

### Refactoring Tasks
- 🔄 Remove notes functionality from CLI
- 🔄 Remove notes views from TUI
- 🔄 Simplify data models to task-only
- 🔄 Update documentation for task focus

## Code Organization

```
denote-tasks/
├── cmd/denote-tasks/     # Main entry point
├── internal/
│   ├── cli/              # CLI implementation (task/project commands)
│   ├── config/           # Configuration
│   ├── core/             # Business logic (task filters)
│   ├── denote/           # Denote file operations (keep for IDs)
│   ├── task/             # Task and project logic
│   └── tui/              # TUI implementation (task-focused)
└── docs/                 # Specifications and plans
```

## Development Workflow

1. **Read REFACTORING_PLAN.md** for current focus
2. **Check docs/DENOTE_TASK_SPEC.md** for file format
3. **Review PROJECT_CHARTER.md** for principles
4. **Update PROGRESS.md** during work
5. **Use test configs** for testing
6. **Document manual test needs**

## Key Differences from Legacy notes-tui

This is NOT a fork of notes-tui. Key differences:
- Task-only focus (no general notes support)
- Denote format for consistent IDs
- Spec v2.0.0 compliant (uses Denote IDs for projects)
- Simpler, focused architecture
- Unified CLI/TUI codebase for tasks

## Testing the TUI

```bash
# Build
go build ./cmd/denote-tasks

# Test with provided config
./denote-tasks --config test-config.toml --tui

# Create test task environment
mkdir ~/denote-tasks-test
# Add task files: *__task*.md
# Add project files: *__project*.md
# Update test-config.toml

# Test task operations
./denote-tasks task list
./denote-tasks task new "Test task" --priority p1
./denote-tasks project list
```

## Common Pitfalls to Avoid

1. **Don't modify user configs** - Always use test configs
2. **Don't assume TUI works** - It needs terminal testing
3. **Don't add non-task features** - This is a task management tool only
4. **Don't add general notes support** - Task and project files only
5. **Don't forget PROGRESS.md** - Update it regularly
6. **Don't add caching** - We're working with small text files. Always question any caching you find and ask if we can remove it. Caching causes staleness bugs without meaningful performance benefits.
7. **Don't stray from focus** - If it's not about tasks/projects, it doesn't belong

## Questions/Decisions

- **Why Bubble Tea?** - Modern, well-maintained, good docs
- **Why not fork notes-tui?** - Too much legacy, want clean start
- **Why task-only?** - Focus and simplicity beat feature creep
- **Why keep Denote format?** - Consistent IDs, backward compatibility
- **Why remove notes?** - Clear purpose, simpler codebase
- **Why unified architecture?** - Easier maintenance, consistent behavior

## Performance Philosophy

We prioritize simplicity and correctness over premature optimization:
- Always read files fresh from disk - no caching
- These are small markdown files (typically < 200 lines)
- File I/O is negligible compared to user interaction time
- Eliminating cache eliminates an entire class of staleness bugs

## Task Management Focus

This project is specifically designed for task management:
- **Tasks**: Files with `__task` tag and task metadata
- **Projects**: Files with `__project` tag and project metadata
- **No General Notes**: We don't support browsing or managing general markdown files
- **Clear Purpose**: If you need notes management, use a different tool

The Denote format is used purely for its excellent file naming convention and ID system, not because we're trying to be a general Denote file manager.

---

Last updated: 2025-01-14 by Claude (Session on refactoring to task-only focus)