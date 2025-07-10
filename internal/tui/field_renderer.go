package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/pdxmph/denote-tasks/internal/denote"
)

// FieldRenderer handles consistent field display
type FieldRenderer struct {
	labelStyle lipgloss.Style
	valueStyle lipgloss.Style
	emptyStyle lipgloss.Style
}

// NewFieldRenderer creates a field renderer with default styles
func NewFieldRenderer() *FieldRenderer {
	return &FieldRenderer{
		labelStyle: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("246")),
		valueStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		emptyStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true),
	}
}

// RenderField renders a single field with consistent formatting
func (fr *FieldRenderer) RenderField(label, value, emptyText string, isEditing bool, editBuffer string) string {
	displayValue := value
	if value == "" {
		displayValue = fr.emptyStyle.Render(emptyText)
	} else {
		displayValue = fr.valueStyle.Render(value)
	}

	if isEditing {
		displayValue = fmt.Sprintf("%s█", editBuffer)
	}

	return fmt.Sprintf("%s %s", 
		fr.labelStyle.Render(fmt.Sprintf("%-12s:", label)),
		displayValue,
	)
}

// RenderPriority renders priority with appropriate styling
func (fr *FieldRenderer) RenderPriority(priority string, isEditing bool, editBuffer string) string {
	if isEditing {
		return fr.RenderField("Priority", "", "none", true, editBuffer)
	}

	var style lipgloss.Style
	switch priority {
	case "p1":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	case "p2":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	case "p3":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("248"))
	default:
		return fr.RenderField("Priority", "", "none", false, "")
	}

	return fmt.Sprintf("%s %s",
		fr.labelStyle.Render("Priority    :"),
		style.Render(priority),
	)
}

// RenderStatus renders status with appropriate symbol
func (fr *FieldRenderer) RenderStatus(status string, isEditing bool, editBuffer string) string {
	if isEditing {
		return fr.RenderField("Status", "", status, true, editBuffer)
	}

	symbol := StatusSymbolOpen
	var style lipgloss.Style

	switch status {
	case denote.TaskStatusDone:
		symbol = StatusSymbolDone
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("70"))
	case denote.TaskStatusPaused:
		symbol = StatusSymbolPaused
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	case denote.TaskStatusDelegated:
		symbol = StatusSymbolDelegated
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
	case denote.TaskStatusDropped:
		symbol = StatusSymbolDropped
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	default:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	}

	return fmt.Sprintf("%s %s %s",
		fr.labelStyle.Render("Status      :"),
		style.Render(symbol),
		style.Render(status),
	)
}

// RenderTags renders tags as a space-separated list
func (fr *FieldRenderer) RenderTags(tags []string, isEditing bool, editBuffer string) string {
	if isEditing {
		return fr.RenderField("Tags", "", "none", true, editBuffer)
	}

	if len(tags) == 0 {
		return fr.RenderField("Tags", "", "none", false, "")
	}

	return fr.RenderField("Tags", strings.Join(tags, " "), "", false, "")
}

// RenderDueDate renders due date with overdue highlighting
func (fr *FieldRenderer) RenderDueDate(dueDate string, isEditing bool, editBuffer string) string {
	if isEditing {
		return fr.RenderField("Due Date", "", "not set", true, editBuffer)
	}

	if dueDate == "" {
		return fr.RenderField("Due Date", "", "not set", false, "")
	}

	// Check if overdue
	if denote.IsOverdue(dueDate) {
		return fmt.Sprintf("%s %s",
			fr.labelStyle.Render("Due Date    :"),
			lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Render(dueDate + " (overdue)"),
		)
	}

	return fr.RenderField("Due Date", dueDate, "", false, "")
}