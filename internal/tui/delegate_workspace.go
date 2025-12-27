package tui

import (
	"fmt"
	"io"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rshelekhov/lazymake/internal/workspace"
)

// WorkspaceItem implements list.Item for workspace picker
type WorkspaceItem struct {
	Workspace workspace.Workspace
	RelPath   string // Relative path to Makefile for display
	RelDir    string // Relative path to directory for description
}

// FilterValue returns the string value to filter against
func (w WorkspaceItem) FilterValue() string {
	return w.Workspace.Path
}

// WorkspaceItemDelegate is a custom delegate for rendering workspace list items
type WorkspaceItemDelegate struct {
	list.DefaultDelegate
}

// NewWorkspaceItemDelegate creates a new delegate with custom styling matching target list
func NewWorkspaceItemDelegate() WorkspaceItemDelegate {
	d := list.NewDefaultDelegate()

	// Apply the same GitHub-inspired colors as target list
	// Selected item (highlighted) - with left border line
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		Foreground(PrimaryColor).
		BorderForeground(PrimaryColor)

	d.Styles.SelectedDesc = d.Styles.SelectedDesc.
		Foreground(SecondaryColor).
		BorderForeground(PrimaryColor)

	// Normal items
	d.Styles.NormalTitle = d.Styles.NormalTitle.
		Foreground(TextPrimary)

	d.Styles.NormalDesc = d.Styles.NormalDesc.
		Foreground(TextSecondary)

	// Dimmed (when filtering)
	d.Styles.DimmedTitle = d.Styles.DimmedTitle.
		Foreground(TextMuted)

	d.Styles.DimmedDesc = d.Styles.DimmedDesc.
		Foreground(TextMuted)

	// Filter match highlighting
	d.Styles.FilterMatch = d.Styles.FilterMatch.
		Foreground(WarningColor).
		Bold(true)

	return WorkspaceItemDelegate{DefaultDelegate: d}
}

func (d WorkspaceItemDelegate) Height() int  { return 2 }
func (d WorkspaceItemDelegate) Spacing() int { return 1 }
func (d WorkspaceItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return d.DefaultDelegate.Update(msg, m)
}

func (d WorkspaceItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	ws, ok := item.(WorkspaceItem)
	if !ok {
		return
	}

	// Determine if this item is selected
	isSelected := index == m.Index()

	// Build title: Favorite indicator + path
	var titleParts []string
	if ws.Workspace.IsFavorite {
		favoriteIcon := lipgloss.NewStyle().
			Foreground(WarningColor).
			Render(IconFavorite)
		titleParts = append(titleParts, favoriteIcon)
	}
	titleParts = append(titleParts, ws.RelPath)
	title := fmt.Sprintf("%s", titleParts[len(titleParts)-1])
	if len(titleParts) > 1 {
		title = titleParts[0] + " " + titleParts[1]
	}

	// Build description: directory path + last accessed time
	var desc string
	if ws.Workspace.AccessCount > 0 {
		// Recently accessed workspace - show directory and last used time
		timeAgo := formatTimeAgo(ws.Workspace.LastAccessed)
		desc = fmt.Sprintf("%s • Last used: %s", ws.RelDir, timeAgo)
	} else {
		// Discovered but never accessed - show directory and "discovered" label
		desc = fmt.Sprintf("%s • Discovered", ws.RelDir)
	}

	// Select appropriate styles based on selection state
	var titleStyle, descStyle lipgloss.Style
	if isSelected {
		titleStyle = d.Styles.SelectedTitle
		// Description always uses muted color, even when selected
		descStyle = d.Styles.SelectedDesc.UnsetWidth().Foreground(TextMuted)
	} else {
		titleStyle = d.Styles.NormalTitle
		descStyle = d.Styles.NormalDesc.UnsetWidth()
	}

	// Render title and description
	fmt.Fprintf(w, "%s\n%s", titleStyle.Render(title), descStyle.Render(desc))
}

// formatTimeAgo formats a timestamp as a human-readable "time ago" string
func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)

	switch {
	case duration < time.Minute:
		return "just now"
	case duration < time.Hour:
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case duration < 24*time.Hour:
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case duration < 7*24*time.Hour:
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	case duration < 30*24*time.Hour:
		weeks := int(duration.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	case duration < 365*24*time.Hour:
		months := int(duration.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		years := int(duration.Hours() / 24 / 365)
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}
