# CLAUDE.md - Project Context for denote-tasks

This file contains important context about the denote-tasks project to help AI assistants understand the codebase, architecture decisions, and current state.

## Project Overview

**denote-tasks** is a task management system built on the Denote file naming convention. It extends Denote with task-specific metadata while maintaining 100% compatibility with Emacs Denote.

### Current Status (2025-01-13)

- Basic CLI working (add, list, done, show commands)
- TUI implementation started but NOT YET TESTED
- Following Denote Task Spec v2.0.0
- Issue #1 (basic TUI) partially implemented

## Important Documents

**ALWAYS READ THESE FIRST:**
- `PROJECT_CHARTER.md` - Vision, principles, and roadmap
- `docs/DENOTE_TASK_SPEC.md` - File format specification (v2.0.0)
- `PROGRESS.md` - Current session status

## Architecture Principles

1. **Denote Purity** - Only Denote files, no other formats
2. **Single Directory** - Notes and tasks coexist
3. **Two Interfaces** - CLI and TUI share core logic
4. **No External Dependencies** - No TaskWarrior, etc.

## Testing Guidelines

### CRITICAL RULE: NEVER MARK FEATURES AS COMPLETE WITHOUT HUMAN TESTING

**STOP AND READ:** Any feature implementation MUST be marked as "IMPLEMENTED BUT NOT TESTED" until the human has confirmed it works. This includes:
- Never marking issues as "✅ Completed" without human confirmation
- Never updating todo lists to "completed" for untested features
- Always use phrases like "implemented but needs testing" or "code complete, awaiting manual testing"
- NEVER assume code that compiles successfully actually works

### For TUI Development

Since TUI applications can't be tested in this environment:

1. **Implement features completely** before declaring done
2. **Document what needs manual testing** in PROGRESS.md
3. **Create test configurations** (never modify user configs)
4. **List specific test cases** for human testing

### Manual Testing Checklist

When implementing TUI features, provide:
- Step-by-step testing instructions
- Expected behavior for each feature
- Edge cases to verify
- Sample test data if needed

## Current Implementation Status

### Completed
- ✅ Core denote package (parser, scanner, types)
- ✅ Basic CLI commands
- ✅ Configuration system
- ✅ Task creation with frontmatter

### In Progress
- 🚧 TUI for browsing notes (Issue #1)
  - Model and views created
  - Needs manual testing
  - See PROGRESS.md for details

### Not Started
- ❌ Task mode in TUI
- ❌ Project management
- ❌ Advanced filtering
- ❌ External editor integration

## Code Organization

```
denote-tasks/
├── cmd/denote-tasks/     # Main entry point
├── internal/
│   ├── cli/              # CLI implementation
│   ├── config/           # Configuration
│   ├── core/             # Business logic (filters)
│   ├── denote/           # Denote file operations
│   ├── task/             # Task-specific logic
│   └── tui/              # TUI implementation (NEW)
└── docs/                 # Specifications
```

## Development Workflow

1. **Read the spec first** - docs/DENOTE_TASK_SPEC.md
2. **Check PROJECT_CHARTER.md** for principles
3. **Update PROGRESS.md** during work
4. **Use test configs** for testing
5. **Document manual test needs**

## Key Differences from Legacy notes-tui

This is NOT a fork of notes-tui. Key differences:
- Denote-only (no plain markdown support)
- Spec v2.0.0 compliant (uses Denote IDs for projects)
- Simpler architecture
- Unified CLI/TUI codebase

## Testing the TUI

```bash
# Build
go build ./cmd/denote-tasks

# Test with provided config
./denote-tasks --config test-config.toml --tui

# Create your own test environment
mkdir ~/denote-test
# Add some .md files with Denote naming
# Update test-config.toml
```

## Common Pitfalls to Avoid

1. **Don't modify user configs** - Always use test configs
2. **Don't assume TUI works** - It needs terminal testing
3. **Don't add non-Denote features** - Stay pure to the spec
4. **Don't forget PROGRESS.md** - Update it regularly

## Questions/Decisions

- **Why Bubble Tea?** - Modern, well-maintained, good docs
- **Why not fork notes-tui?** - Too much legacy, want clean start
- **Why unified architecture?** - Easier maintenance, consistent behavior

---

Last updated: 2025-01-13 by Claude (Session on TUI implementation)