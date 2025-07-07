# Progress Report - TUI Implementation

## CRITICAL REMINDER: NEVER MARK FEATURES AS COMPLETE WITHOUT HUMAN TESTING

Any feature marked as "✅ Completed" means it has been TESTED AND VERIFIED by a human.
Features that compile but haven't been tested must be marked as "IMPLEMENTED BUT NOT TESTED".

## Session: 2025-01-13 - Project View Implementation

### What Was Done

Created tabbed project view interface:

#### Files Created
- `internal/tui/project_view.go` - Project view rendering with tabs
- `internal/tui/project_view_keys.go` - Project view key handling

#### Files Modified
- `internal/tui/model.go` - Added ModeProjectView, project state fields
- `internal/tui/keys.go` - Updated to handle project view mode
- `internal/tui/views.go` - Projects now show due dates and priority

### Current Status

**IMPLEMENTED BUT NOT TESTED:**

1. **Enhanced Project Display in Task List**
   - Projects now show priority [p1/p2/p3] like tasks
   - Projects show due dates with overdue highlighting
   - Priority coloring applied to projects
   - Shortened status badges to fit with due dates

2. **Project View Mode (ModeProjectView)**
   - Opens when pressing Enter on a project
   - Two tabs: Overview and Tasks
   - Tab switching with 'tab' key
   - Different UI than task view (not hybrid edit mode)

3. **Overview Tab Features**
   - Shows all project metadata
   - Inline editing with same keys as task view (p/s/d/a/g)
   - Natural language date parsing for due dates
   - Body content display below metadata
   - Priority and status coloring

4. **Tasks Tab Features**
   - Lists all tasks assigned to this project
   - Shows task count in tab header
   - Task navigation with j/k
   - Quick priority updates with 1/2/3 keys
   - Open task in task view with Enter
   - Task display with status, priority, title, due date

### Testing Needed

1. **Project List Display**
   - Do projects show priority and due dates correctly?
   - Is overdue highlighting working?
   - Are priority colors applied?

2. **Project View - Overview Tab**
   - Does tab key switch between tabs?
   - Can you edit metadata with p/s/d/a/g keys?
   - Does natural language date parsing work?
   - Is the project body content displayed?

3. **Project View - Tasks Tab**
   - Are tasks correctly filtered by project ID?
   - Does task navigation (j/k) work?
   - Can you set task priorities with 1/2/3?
   - Does Enter open the task in task view?

4. **Edge Cases**
   - Projects with no tasks
   - Projects with no metadata
   - Large number of tasks

### How to Test

```bash
# Build the application
go build ./cmd/denote-tasks

# Run with test config
./denote-tasks --config test-config.toml --tui

# Test project features:
# 1. Press 't' to enter task mode
# 2. Navigate to a project (▶ indicator)
# 3. Press Enter to open project view
# 4. Press tab to switch between Overview/Tasks
# 5. Try editing fields with p/s/d/a/g
# 6. In Tasks tab, navigate and set priorities
```

### Known Limitations

1. Task creation from project view not implemented
2. No sorting in project tasks list
3. No filtering within project tasks
4. No bulk operations on project tasks

## Previous Session Progress

### Session: 2025-01-13

### What Was Done

Created basic TUI structure for Issue #1 (Browse all Denote notes):

#### Files Created
- `internal/tui/model.go` - Core Bubble Tea model
- `internal/tui/keys.go` - Keyboard handling  
- `internal/tui/views.go` - View rendering
- `internal/tui/tui.go` - Entry point

#### Files Modified
- `internal/denote/scanner.go` - Added FindAllNotes() method
- `internal/denote/types.go` - Added File methods (HasTag, MatchesSearch)
- `internal/denote/parser.go` - Added Parser struct
- `cmd/denote-tasks/main.go` - Integrated TUI support
- `go.mod` - Added Bubble Tea dependencies

#### Test Setup
- Created `test-notes/` directory with sample Denote files
- Created `test-config.toml` pointing to test directory

### Current Status

✅ **TESTED AND WORKING** - The TUI has been successfully tested in a real terminal environment. The implementation includes:

- File listing with ID, title, and tags
- Keyboard navigation (j/k, arrows, gg, G)
- Search functionality (/)
- Sort by date/title (s key)
- Reverse sort toggle (r key)
- Basic file preview (Enter key)
- Help screen (? key)

### Testing Results (2025-01-13)

All features have been manually tested and verified working:

1. **Basic Navigation**
   - Does the file list display correctly?
   - Do j/k and arrow keys work?
   - Does gg/G jump to top/bottom?

2. **Search Feature**
   - Does / key open search prompt?
   - Does search filter files correctly?
   - Does Esc cancel search?

3. **Sorting**
   - Does s toggle between date/title sort?
   - Does r reverse the sort order?

4. **Preview**
   - Does Enter show file content?
   - Can you exit preview with q/Esc?

5. **Edge Cases**
   - Empty directory
   - Non-Denote files mixed in
   - Large number of files

### Issues Found and Fixed

1. **Low contrast on selection** - Fixed by updating selectedStyle with better colors (blue background, light text, bold)

### Known Limitations (Acceptable for v1)

1. **Preview is basic** - Just shows raw file content, no markdown rendering
2. **No scrolling in preview** - Long files may not display fully
3. **No status bar messages** - User feedback could be improved
4. **No configuration for themes** - Using hardcoded styles

### Completed

✅ Issue #1: Create TUI for browsing all Denote notes
- Basic TUI implementation with Bubble Tea
- File listing, navigation, search, sort, preview
- Improved selection contrast after testing

✅ Issue #2: Implement general note creation  
- Added 'n' key to create new notes
- Two-step dialog for title and tags
- Strict YAML validation using yaml.v3 library
- Protection against markdown --- confusion
- Editor integration support

✅ Issue #3: Add tag filtering for all notes
- In-place fuzzy filtering (no screen change)
- True fuzzy matching algorithm
- '#' prefix for tag-only search
- Real-time results as you type
- Tested and confirmed working

## Issue #4: Task Mode Implementation ✅ TESTED

### Requirements clarified: 2025-01-13

**Scope:** Full task management mode, not just filtering

**Two Distinct Modes:**

1. **Notes Mode** (default)
   - Shows all files EXCEPT those with task/project tags
   - Current interface and hotkeys remain

2. **Task Mode** (toggle with 't')
   - Shows ONLY files with task/project tags
   - Completely different interface with task-specific hotkeys:
     - 1/2/3 - Set priority (p1/p2/p3)
     - e - Edit task file in external editor
     - u - Update task metadata (interactive)
     - Sort options: priority, estimate, due date
     - Filter by "soon" (configurable due date horizon)
     - Filter by area (work/personal/etc)
   - Area filtering is "sticky":
     - Persists across other operations
     - New tasks inherit current area filter

**Implementation Plan:**
1. ✅ Add ViewMode enum (Notes/Tasks) to model
2. ✅ Toggle mode with 't' key in Notes mode
3. ✅ Filter files based on active mode
4. Create separate hotkey handlers for task mode
5. Implement task-specific features incrementally

### Implementation Progress (2025-01-13)

**✅ TESTED AND CONFIRMED BY USER:**

1. **Mode Toggle Infrastructure** ✅
   - Added ViewMode enum (ViewModeNotes, ViewModeTasks)
   - Added viewMode field to Model struct
   - Initialize to ViewModeNotes by default

2. **File Filtering by Mode** ✅
   - Updated applyFilters() to check viewMode
   - Notes mode: Excludes files with task/project tags
   - Task mode: Shows ONLY files with task/project tags
   - Uses IsTask() and IsProject() methods

3. **Toggle Functionality** ✅
   - 't' key toggles between modes
   - Shows status message "Task Mode" or "Notes Mode"
   - Resets cursor and reapplies filters
   - Updated header to show "Denote Tasks" in task mode
   - Updated status line to show "X tasks" vs "X notes"
   - Updated empty message: "No tasks found" vs "No notes found"

4. **UI Updates** ✅
   - Footer shows "t:toggle mode" help text
   - Help screen documents the 't' key
   - Status messages clear on next key press

5. **Task-Specific Hotkeys** ✅
   - Created handleTaskModeKeys() function
   - Different key bindings active in task mode:
     - 1/2/3 keys for priority (shows "not yet implemented" message)
     - e key for edit (shows "not yet implemented" message)
     - u key for update metadata (shows "not yet implemented" message)
     - Common navigation keys still work (j/k, gg, G, /, etc.)
   - 't' key toggles back to Notes mode from Task mode

6. **UI Updates for Task Mode** ✅
   - Footer shows different hotkeys based on mode
   - Help screen shows different shortcuts for each mode
   - Task mode shows task-specific actions
   - Removed "n:new" from task mode (tasks created via CLI)

### Testing Results (2025-01-13)

All features tested and confirmed working:
- Mode toggle with 't' key switches between Notes and Task modes
- File filtering correctly shows/hides task files based on mode
- UI updates (header, status line, footer) reflect current mode
- Task-specific hotkeys show placeholder messages
- Help screen shows mode-specific shortcuts
- Search functionality works correctly in both modes

## Task Metadata Features (2025-01-13)

**✅ TESTED:**

1. **Task Metadata Loading** ✅
   - Added taskMetadata map to Model
   - scanFiles() now loads task metadata using ParseTaskFile
   - Metadata stored in memory for quick access

2. **Task Display in List View** ✅
   - Created renderTaskLine() for task-specific formatting
   - Shows: Status symbol, Priority, Title, Area, Due Date
   - Special formatting for overdue and due-this-week tasks
   - Proper truncation and alignment

3. **Priority Updates (1/2/3 keys)** ✅
   - Implemented updateTaskPriority() method
   - 1/2/3 keys set priority to p1/p2/p3
   - Updates both file and in-memory metadata
   - Shows confirmation message

4. **External Editor Integration** ✅
   - 'e' key opens task/note in configured editor
   - Works in both Notes and Task modes
   - Refreshes display after editing
   - Requires `editor` setting in config

5. **Task Creation ('c' key)** ✅
   - Creates new tasks with proper frontmatter
   - Inherits area filter if active
   - Two-step process: title then tags
   - Positions cursor on new task

6. **Area Filtering ('f' key)** ✅
   - Filter tasks by area (work, personal, etc)
   - Case-insensitive matching
   - Shows active filter in status line
   - Clear filter with empty input
   - New tasks inherit area filter

7. **Enhanced Sorting ('s' key)** ✅
   - Sort mode with visual menu
   - Task-specific options: due date, priority, estimate
   - Direct hotkeys: d, p, e, t, c, r
   - Shows current sort in menu

8. **Performance Optimizations** ✅
   - Lazy loading of metadata
   - Only loads visible files
   - Fixed slow refresh after task creation

9. **Task View Improvements** ✅
   - Shows body content below metadata
   - Text wrapping at 80 characters
   - Horizontal separator
   - "(no notes)" when empty

10. **Natural Language Due Dates** ✅
    - Relative: 1d, 5d, 1w, 1m, 1y
    - Named: today, tomorrow, friday
    - Month/day: jan 15, 15 jan
    - Clear error messages

**Remaining Work for Full Task Mode:**
- None - all major features implemented and tested!

### Potential Future Enhancements

- Better preview with scrolling
- Status messages for user feedback  
- Theme configuration support
- File creation (n key)
- External editor support (e key)
- Task mode (t key) as specified in other issues

### How to Test

```bash
# Build the application
go build ./cmd/denote-tasks

# Run with test config
./denote-tasks --config test-config.toml --tui

# Or set up your own test directory
mkdir my-test-notes
# Copy some Denote files there
# Update test-config.toml to point to your directory
```

### Files That Need Cleanup

None - all changes are permanent additions to the project.

### Important Notes

- This is Phase 1 of the unified architecture
- Focuses only on general note browsing (not task-specific features)
- Follows Denote-first philosophy from PROJECT_CHARTER.md
- Implements Issue #1 requirements exactly

## Issue #3: Tag Filtering - FUZZY SEARCH IMPLEMENTATION ✅ TESTED

### Implementation Date: 2025-01-13
### Revised: 2025-01-13 (after user feedback on menu approach)
### Revised again: 2025-01-13 (added true fuzzy matching)
### Final revision: 2025-01-13 (in-place filtering, no separate screen)

**Status:** ✅ TESTED AND WORKING - In-place fuzzy filtering confirmed by user

**Evolution of approaches:**
1. **First attempt:** Separate tag filter menu with 't' key (removed)
2. **Second attempt:** Search with '#' prefix but substring matching (improved)
3. **Third attempt:** True fuzzy but still separate search screen (improved)
4. **Current:** In-place fuzzy filtering - no screen change!

**Current implementation:**
- Press '/' to start searching - stays on main screen
- Search input appears at bottom, list filters as you type
- TRUE FUZZY MATCHING - letters must appear in order but can be non-consecutive
- Use '#' prefix for tag-only search (e.g., '#dv' for development)
- Real-time filtering - instant results, no Enter required
- Esc clears search and returns to normal navigation

**Files Modified:**
- `internal/tui/model.go` - Removed separate search screen from View()
- `internal/tui/keys.go` - Real-time search updates in handleSearchKeys()
- `internal/tui/views.go` - Search input in footer, removed renderSearch()
- `internal/denote/types.go` - Added fuzzyMatch() algorithm, updated MatchesSearch() and MatchesTag()

**How to Test:**
```bash
go build ./cmd/denote-tasks
./denote-tasks --config test-config.toml --tui
```

**Fuzzy Search Test Cases:**
1. Press '/' to open search
2. Type '#dv' - should match 'development' tag (fuzzy)
3. Type '#apl' - should match 'apple' tag
4. Type '#wrk' - should match 'work' tag
5. Type 'impl' - should match titles with 'implement' (regular fuzzy search)
6. Backspace to see results update in real-time
7. No Enter needed - results update as you type
8. Esc clears search and shows all notes

**Fuzzy Match Examples:**
- 'dv' matches 'development'
- 'apl' matches 'apple'
- 'mtg' matches 'meeting'
- 'prj' matches 'project'
- 'bnna' matches 'banana'

## Session: 2025-01-14 - Spec Updates and UI Improvements

### What Was Done

1. **Updated to Spec v2.0.1 - Universal index_id Field**
   - Changed from separate `task_id`/`project_id` to universal `index_id`
   - Updated all type definitions in `internal/denote/types.go`
   - Modified ID counter to use `next_index_id` instead of separate counters
   - Updated parser, scanner, CLI, and task creation logic
   - Counter file now tracks single `next_index_id` value
   - Maintains backward compatibility during counter file migration
   - Updated `DENOTE_TASK_SPEC.md` to version 2.0.1

2. **Task List Default Sort Changed**
   - CLI: Default sort changed from ID to due date (ascending)
   - TUI: When entering task mode ('t'), automatically sets sort to due date
   - Shows tasks with closest due dates first
   - Overdue tasks naturally appear at top

3. **Implemented Real-time Fuzzy Search in TUI**
   - Fixed regression where search wasn't filtering progressively
   - Search now updates results as you type (no Enter needed)
   - Enhanced to search task metadata (status, priority, area, assignee)
   - Maintains fuzzy matching for all fields
   - Esc cancels search and restores previous view
   - Enter confirms search and exits search mode

4. **Added Tag Display to Task Lists**
   - Tags now display in task/project lists similar to notes view
   - Format: `[tag1, tag2, tag3]` at end of line
   - Automatically filters out "task" and "project" tags
   - Fixed alignment issues with variable-width due date fields
   - Added consistent column spacing for better readability

### Files Modified
- `internal/denote/types.go` - Changed TaskID/ProjectID to IndexID
- `internal/denote/id_counter.go` - Unified counter system
- `internal/denote/parser.go` - Updated to check for index_id
- `internal/denote/scanner.go` - Updated sort to use IndexID
- `internal/denote/frontmatter.go` - Updated validation for index_id
- `internal/task/task.go` - Uses NextIndexID() for both tasks and projects
- `internal/cli/cli.go` - Default sort changed to due date, added fuzzy search
- `internal/tui/keys.go` - Real-time search filtering, auto-sort on task mode
- `internal/tui/model.go` - Enhanced taskMatchesSearch for metadata
- `internal/tui/views.go` - Added tag display with proper alignment
- `docs/DENOTE_TASK_SPEC.md` - Updated to v2.0.1 with index_id
- `scripts/migration/` - Created migration scripts for test data

### Testing Status

**IMPLEMENTED BUT NOT TESTED:**

1. **index_id Migration**
   - Need to verify backward compatibility with old counter files
   - Test that new tasks/projects get sequential IDs correctly
   - Verify counter file migration works smoothly

2. **Fuzzy Search Enhancement**
   - Test that metadata search works (search "p1" for priority)
   - Test area search (search "wrk" for work area)
   - Verify project name search in task metadata

3. **Tag Display**
   - Check tag display alignment with various due date formats
   - Verify long tag lists truncate properly
   - Test with tasks that have no additional tags

### How to Test

```bash
# Build the application
go build ./cmd/denote-tasks

# Test CLI list with new sort
./denote-tasks --config test-config.toml list

# Test CLI fuzzy search
./denote-tasks --config test-config.toml list --search "p1"

# Test TUI improvements
./denote-tasks --config test-config.toml --tui
# Press 't' for task mode (should auto-sort by due date)
# Press '/' and type to test real-time fuzzy search
# Check tag display alignment
```

### Known Issues Fixed
- Fuzzy search now works progressively in TUI
- Due date field alignment is now consistent
- Tags display properly without breaking column alignment

## Session: 2025-01-14 - Task Management Hotkeys

### What Was Done

Added task state management and delete functionality to the TUI:

#### Files Modified
- `internal/tui/model.go` - Added ModeStateMenu and ModeConfirmDelete modes
- `internal/tui/keys.go` - Added handlers for state change and delete keys
- `internal/tui/views.go` - Added renderStateMenu() and renderConfirmDelete() views

### Features Added

**IMPLEMENTED BUT NOT TESTED:**

1. **Task State Changes (s key)**
   - Press 's' on a task to open state change menu
   - Options: (o)pen, (p)aused, (d)one, d(e)legated, d(r)opped
   - Updates task status in file and cached metadata
   - Shows confirmation message after change

2. **Delete Functionality (x key)**
   - Press 'x' on any task, project, or note to delete
   - Works in both Task mode and Notes mode
   - Shows confirmation dialog
   - Press 'y' to confirm, 'n' or Esc to cancel
   - Removes file from filesystem
   - Automatically rescans files after deletion

3. **UI Updates**
   - Footer now shows: "s:state" and "x:delete" in task mode
   - Footer shows "x:delete" in notes mode too
   - Sort moved to uppercase 'S' to make room for state (task mode only)
   - Help screens updated with new hotkeys for both modes

### Testing Needed

1. **State Change Menu**
   - Does 's' key open the state menu for tasks?
   - Do all state options work (o/p/d/e/r)?
   - Is the task list updated after state change?
   - Does it prevent state change on projects?

2. **Delete Functionality**
   - Does 'x' key show delete confirmation?
   - Does 'y' actually delete the file?
   - Does 'n'/Esc cancel properly?
   - Does file list refresh after deletion?

3. **Edge Cases**
   - What happens when deleting the last task?
   - Can you delete projects?
   - Does state change work on all task states?

### How to Test

```bash
# Build the application
go build ./cmd/denote-tasks

# Run with test config
./denote-tasks --config test-config.toml --tui

# Test new features:
# 1. Press 't' to enter task mode
# 2. Navigate to a task
# 3. Press 's' to change state
# 4. Try different state options
# 5. Press 'x' to delete a task
# 6. Confirm with 'y' or cancel with 'n'
```

## Session: 2025-01-14 - Advanced Filtering

### What Was Done

Added comprehensive filtering options to the task view:

#### Files Modified
- `internal/config/config.go` - Added soon_horizon config option (default 3 days)
- `internal/tui/model.go` - Added filter state fields and new modes
- `internal/tui/keys.go` - Added filter menu handlers
- `internal/tui/views.go` - Added filter menu views
- `internal/denote/types.go` - Added IsDueSoon() function

### Features Added

**IMPLEMENTED BUT NOT TESTED:**

1. **Filter Menu (f key)**
   - Press 'f' in task mode to open filter menu
   - Shows all active filters
   - Options:
     - (a) Area filter (existing, now in menu)
     - (p) Priority filter
     - (s) State filter  
     - (d) Due soon toggle
     - (c) Clear all filters

2. **Priority Filter**
   - Filter by p1, p2, or p3
   - Works on both tasks and projects
   - Quick keys: 1/2/3 to set, c to clear

3. **State Filter**
   - Filter by task state: open, paused, done, delegated, dropped
   - Only applies to tasks (not projects)
   - Quick keys: o/p/d/e/r to set, c to clear

4. **Soon Filter**
   - Toggle filter for items due within N days
   - Configurable horizon via soon_horizon in config (default 3 days)
   - Shows tasks/projects due today or within horizon
   - Excludes overdue items

5. **UI Updates**
   - Header shows all active filters
   - Filters are sticky and persist across operations
   - Multiple filters can be active simultaneously
   - Footer shows "f:filter" in task mode

### Configuration

Add to your config.toml:
```toml
soon_horizon = 7  # Show items due within 7 days (default is 3)
```

### Testing Needed

1. **Filter Menu**
   - Does 'f' key open the filter menu?
   - Are active filters displayed correctly?
   - Does 'c' clear all filters?

2. **Individual Filters**
   - Priority: Do 1/2/3 keys filter correctly?
   - State: Do o/p/d/e/r keys filter correctly?
   - Area: Does area filter still work?
   - Soon: Does toggle show only items due soon?

3. **Filter Combinations**
   - Can multiple filters be active at once?
   - Do filters combine correctly (AND logic)?
   - Are filters maintained when sorting/searching?

4. **Soon Filter Edge Cases**
   - Items due today
   - Items due in exactly N days
   - Overdue items (should be excluded)
   - Items with no due date

### How to Test

```bash
# Build the application
go build ./cmd/denote-tasks

# Add soon_horizon to test-config.toml (optional)
echo "soon_horizon = 5" >> test-config.toml

# Run with test config
./denote-tasks --config test-config.toml --tui

# Test filters:
# 1. Press 't' to enter task mode
# 2. Press 'f' to open filter menu
# 3. Try different filter options
# 4. Combine multiple filters
# 5. Use 'c' in filter menu to clear all
```

## Session Summary: 2025-01-14

### Major Accomplishments

This session focused on enhancing task management capabilities in the TUI with two major feature sets:

#### 1. Task State Management & Deletion
- **State Changes**: Added 's' hotkey to quickly change task states between open, paused, done, delegated, and dropped
- **Delete Functionality**: Added 'x' hotkey for deleting tasks/projects/notes with confirmation dialog
- **Safety**: Delete requires 'y' confirmation, preventing accidental deletions

#### 2. Advanced Filtering System
- **Unified Filter Menu**: Replaced single area filter with comprehensive filter menu ('f' key)
- **Multiple Filter Types**:
  - Area: Filter by work context (existing, now in menu)
  - Priority: Filter by p1/p2/p3
  - State: Filter by task state (open/done/etc)
  - Soon: Show items due within configurable days
- **Filter Combinations**: All filters work together with AND logic
- **Sticky Filters**: Filters persist across all operations
- **Configuration**: Added `soon_horizon` setting (default 3 days)

### User Experience Improvements
- Clear visual feedback for active filters in header
- Consistent hotkey patterns (state change and state filter use same keys)
- Sort moved to 'S' (uppercase) in task mode to make room for state
- One-key clearing of all filters ('c' in filter menu)

### Technical Implementation
- Added new modes to state machine for menus
- Extended model with filter fields
- Implemented IsDueSoon() for date calculations
- Maintained clean separation between view modes and filters

All features follow existing codebase patterns and are ready for testing. The implementation provides powerful task management while maintaining the simple, keyboard-driven interface.

### Visual Improvements - Task Status Indicators

Updated task status symbols for better visual clarity:
- **Done**: ✓ (checkmark) - Green color
- **Paused**: ⏸ (pause symbol) - Dim gray color
- **Delegated**: → (right arrow) - Blue color
- **Dropped**: ⨯ (crossed circle) - Dark gray color
- **Open**: ○ (open circle) - Default color

Each status also has distinct coloring to make task states immediately recognizable at a glance. The symbols provide intuitive visual cues that match common UI conventions.

### Refined Priority Coloring

Updated priority display to be more subtle and focused:
- **[p1]** badge only is colored red (not the entire line)
- **[p2]** badge only is colored orange (not the entire line)
- **[p3]** badge only is colored gray (not the entire line)

This change makes the interface cleaner and less overwhelming while still providing clear visual priority indicators. The colored badges stand out without dominating the entire task display.

### Active State Filter

Added a new "active" filter state that shows both open and delegated tasks:
- **Active filter**: Shows tasks with status "open" OR "delegated" 
- **Default behavior**: Task mode now starts with "active" filter applied
- **Menu option**: Press 'a' in state filter menu to select active filter
- **Rationale**: Provides complete view of all "live" work items by default

When entering task mode (pressing 't'), the active filter is automatically applied, showing all tasks that need attention (both those you're working on and those delegated to others). The filter is cleared when returning to notes mode.

### Bug Fix: Denote File Renaming on Tag Changes

Fixed a critical bug where changing tags on a task or project didn't rename the file as required by the Denote specification:
- **Issue**: Tags were only updated in frontmatter, not in filename
- **Fix**: When tags are modified, the file is now renamed to reflect new tags
- **Implementation**: 
  - Added `RenameFileForTags()` function to handle Denote-compliant renaming
  - Preserves the identifier (timestamp) and slug
  - Updates only the tag portion of the filename
  - Updates all internal references after rename
- **Example**: `20240115T093000--project-name__work.md` → `20240115T093000--project-name__work_urgent.md`

This ensures full compliance with Denote's file naming convention where tags must be reflected in both the filename and frontmatter.

## Session Summary: 2025-01-14 - Part 2

### Visual and UX Improvements

This session continued with significant visual and usability enhancements:

#### 1. Enhanced Status Symbols
- **✓** Checkmark for done (was filled circle)
- **⏸** Pause symbol for paused (was half circle)
- **→** Right arrow for delegated (new)
- **⨯** Crossed circle for dropped (new)
- Each status has distinct coloring for quick recognition

#### 2. Refined Visual Hierarchy
- **Priority badges only colored** - [p1] red, [p2] orange, [p3] gray
- **Base text brightened** - Changed from dim gray (245) to near-white (252)
- **Menu text brightened** - All menus now use readable colors
- **Better contrast** throughout the interface

#### 3. Active Filter Enhancement
- **New "active" state** - Shows both open AND delegated tasks
- **Default filter** - Task mode starts with active filter
- **Complete view** - See all "live" work at a glance
- **Smart clearing** - Filter clears when returning to notes mode

#### 4. Critical Bug Fix
- **Denote compliance** - Tags changes now rename files per spec
- **Preserves ID** - Timestamp and slug remain unchanged
- **Updates references** - All internal paths updated correctly
- **Safe operation** - Checks for conflicts before renaming

### Technical Improvements
- Added `RenameFileForTags()` for Denote-compliant file operations
- Improved color system with base styles for consistency
- Enhanced filter logic with OR conditions for active state
- Better state management across mode transitions

### File Rename Feature in Task View
- **Added 'r' key** in task/project view to sync filename with metadata tags
- **Use case**: Fixes files where tags were changed but filename wasn't updated
- **Implementation**: Uses the new `RenameFileForTags()` function
- **Safety**: Checks for conflicts before renaming
- **Feedback**: Shows status message after rename operation

### Auto-Rename on Editor Return
- **Automatic check**: When returning from external editor, checks if tags changed
- **Smart rename**: If metadata tags differ from filename tags, automatically renames
- **Seamless experience**: User edits tags in editor, file is renamed on return
- **File tracking**: Updates all internal references to use new path
- **Status feedback**: Shows "File renamed to match updated tags" when rename occurs

### Bug Fixes
- **Fixed**: Raw filenames showing after rename operation
- **Solution**: Added `m.scanFiles()` call after rename to refresh file list
- **Result**: Task list always shows properly formatted lines after rename

### Enhanced Project View Task Management
- **Task management in project view**: Added full task management capabilities to project task tab
  - 's' key opens state change menu (open/paused/done/delegated/dropped)
  - 'x' key allows task deletion with confirmation
  - 1/2/3 keys for quick priority changes (existing)
- **Smart navigation return**: When viewing/editing a task from project view:
  - System remembers you came from project view
  - Pressing 'q' or 'esc' returns to project task list, not main task list
  - Maintains context and workflow continuity
- **Unified handlers**: State menu and delete confirmation work seamlessly in both contexts
  - Automatically detects whether in project view or main task list
  - Returns to appropriate view after operation
- **Updated UI hints**: Project task tab now shows all available operations

The interface is now more readable, more compliant with Denote specifications, and provides better default views for task management.