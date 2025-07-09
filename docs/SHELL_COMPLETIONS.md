# Shell Completions

denote-tasks includes comprehensive shell completion support for both bash and zsh.

## Installation

### Quick Install

Run the included installation script:

```bash
./install-completions.sh
```

This will detect your shell and install completions to the appropriate location.

### Manual Installation

#### Bash

Copy the completion file to one of these locations:
- `/usr/local/etc/bash_completion.d/denote-tasks`
- `/etc/bash_completion.d/denote-tasks`
- `~/.local/share/bash-completion/completions/denote-tasks`

Or source it in your `.bashrc`:
```bash
source /path/to/denote-tasks/completions/denote-tasks.bash
```

#### Zsh

Copy the completion file to a directory in your `fpath`:
- `/usr/local/share/zsh/site-functions/_denote-tasks`
- `~/.zsh/completions/_denote-tasks`

Or add the completions directory to your `fpath` in `.zshrc`:
```bash
fpath=(~/path/to/denote-tasks/completions $fpath)
autoload -Uz compinit && compinit
```

## Features

### Command Completion
- Main commands: `task`, `project`, `note`
- Subcommands for each entity type
- Legacy command aliases

### Smart Argument Completion
- **Task IDs**: Dynamic completion of existing task IDs
- **Project IDs**: Shows both ID and project name
- **Areas**: Completes from existing areas in your notes
- **Tags**: Completes from existing tags

### Flag Completion
- All flags and options with descriptions
- Context-aware value suggestions:
  - Priority: `p1`, `p2`, `p3`
  - Status: `open`, `done`, `paused`, `delegated`, `dropped`
  - Sort options: `modified`, `priority`, `due`, `created`
  - Due dates: `today`, `tomorrow`, weekday names

### Examples

```bash
# Complete commands
denote-tasks <TAB>
# Shows: task project note --tui --help --version

# Complete task subcommands
denote-tasks task <TAB>
# Shows: new list update done edit delete log

# Complete task IDs for done command
denote-tasks task done <TAB>
# Shows: 1 2 3 5 8 13 (actual task IDs)

# Complete priority values
denote-tasks task new "Fix bug" -p <TAB>
# Shows: p1 p2 p3

# Complete areas
denote-tasks task list -area <TAB>
# Shows: work personal hobby (from your existing tasks)

# Complete due date shortcuts
denote-tasks task new "Meeting" -due <TAB>
# Shows: today tomorrow monday tuesday wednesday...
```

## Advanced Usage

### Range and List Completion

The completions understand the smart task argument format:
- Single IDs: `3`
- Ranges: `3-5`
- Lists: `3,5,7`
- Mixed: `3,5-7,10`

### Dynamic Data

The completion system uses a special `completion` command to get dynamic data:

```bash
# Get all task IDs
denote-tasks completion task-ids

# Get all project IDs with names
denote-tasks completion project-ids

# Get all areas
denote-tasks completion areas

# Get all tags
denote-tasks completion tags
```

This ensures completions always reflect your current data.

## Troubleshooting

### Completions Not Working

1. **Reload your shell**: `source ~/.bashrc` or `source ~/.zshrc`
2. **Clear zsh cache**: `rm -f ~/.zcompdump && compinit`
3. **Check installation**: Verify the completion file is in the right location
4. **Test manually**: Try running `denote-tasks completion task-ids` to ensure it works

### Performance

Completions are cached per shell session. If you add new tasks/projects/areas, you may need to start a new shell to see them in completions.