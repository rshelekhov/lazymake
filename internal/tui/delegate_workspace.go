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
	RelPath   string // Relative path for display
}

// FilterValue returns the string value to filter against
func (w WorkspaceItem) FilterValue() string {
	return w.Workspace.Path
}

// WorkspaceItemDelegate is a custom delegate for rendering workspace list items
type WorkspaceItemDelegate struct{}

func (d WorkspaceItemDelegate) Height() int  { return 2 }
func (d WorkspaceItemDelegate) Spacing() int { return 1 }
func (d WorkspaceItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d WorkspaceItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	ws, ok := item.(WorkspaceItem)
	if !ok {
		return
	}

	// Determine if this item is selected
	isSelected := index == m.Index()

	// Line 1: Favorite indicator + path
	var prefix string
	if ws.Workspace.IsFavorite {
		favoriteIcon := lipgloss.NewStyle().
			Foreground(WarningColor).
			Render(IconFavorite)
		prefix = favoriteIcon + " "
	} else {
		prefix = "  "
	}

	// Style the path
	pathStyle := lipgloss.NewStyle().Foreground(PrimaryColor)
	if isSelected {
		pathStyle = pathStyle.Bold(true)
	}

	line1 := pathStyle.Render(prefix + ws.RelPath)

	// Line 2: Metadata (last accessed time, access count)
	metaStyle := lipgloss.NewStyle().Foreground(TextMuted)

	var line2 string
	if ws.Workspace.AccessCount > 0 {
		// Recently accessed workspace - show access info
		timeAgo := formatTimeAgo(ws.Workspace.LastAccessed)
		accessInfo := fmt.Sprintf("%d times", ws.Workspace.AccessCount)
		line2 = metaStyle.Render(fmt.Sprintf("   Last used: %s • %s", timeAgo, accessInfo))
	} else {
		// Discovered but never accessed - show "discovered" label
		line2 = metaStyle.Render("   Discovered in project")
	}

	fmt.Fprintf(w, "%s\n%s", line1, line2)
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
