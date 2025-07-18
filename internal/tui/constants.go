package tui

// UI Layout Constants
const (
	HeaderFooterHeight = 6
	MinVisibleHeight   = 5
	ScrollPageSize     = 10
	MaxHelpWidth       = 80
)

// Status Symbols
const (
	StatusSymbolOpen      = "○"
	StatusSymbolDone      = "✓"
	StatusSymbolPaused    = "⏸"
	StatusSymbolDelegated = "→"
	StatusSymbolDropped   = "⨯"
	StatusSymbolActive    = "●"
)

// Priority Levels
var PriorityLevels = []string{"p1", "p2", "p3"}

// Field Types for editing
type EditableField string

const (
	FieldTitle    EditableField = "title"
	FieldPriority EditableField = "priority"
	FieldStatus   EditableField = "status"
	FieldDueDate  EditableField = "due_date"
	FieldArea     EditableField = "area"
	FieldTags     EditableField = "tags"
	FieldEstimate EditableField = "estimate"
	FieldProject  EditableField = "project"
)

// Sort Fields
type SortField string

const (
	SortByDue      SortField = "due"
	SortByPriority SortField = "priority"
	SortByStatus   SortField = "status"
	SortByTitle    SortField = "title"
	SortByCreated  SortField = "created"
	SortByModified SortField = "modified"
	SortByProject  SortField = "project"
	SortByEstimate SortField = "estimate"
)

// Common Messages
const (
	MsgNoTaskSelected    = "No task selected"
	MsgNoProjectSelected = "No project selected"
	MsgNoEditor          = "No editor configured"
	MsgTasksOnly         = "Task management only - notes mode removed"
	MsgAlreadyInTaskMode = "Already showing tasks"
	MsgFuzzyMatch        = " (fuzzy match, #tag for tags, Esc to clear)"
	MsgPressEnterSelect  = "press Enter to select"
	MsgPressEnterChange  = "press Enter to change"
	MsgSpaceSeparated    = "space-separated"
)

// Column Widths
const (
	ColumnWidthTitle       = 40
	ColumnWidthTitleAlt    = 30
	ColumnWidthTags        = 20
	ColumnWidthArea        = 10
	ColumnWidthProject     = 15
	ColumnWidthID          = 15
	ColumnWidthDueSpaces   = 12
	ColumnWidthEstimate    = 5
	ProjectViewHeaderHeight = 10
	DefaultVisibleHeight   = 20
)

// Error Formats
const (
	ErrorFormat   = "Error: %v"
	ErrorFailedTo = "failed to %s: %w"
)

// Date Formats
const (
	DateFormatSimple  = "2006-01-02"
	DateFormatWithDay = "[2006-01-02 Mon]"
)

// UI Format Strings
const (
	SortFormatString = "Sort: %s"
)