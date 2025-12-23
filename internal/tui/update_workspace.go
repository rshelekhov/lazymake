package tui

import (
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rshelekhov/lazymake/config"
	"github.com/rshelekhov/lazymake/internal/workspace"
)

// workspaceSwitchedMsg is sent when workspace switch completes
type workspaceSwitchedMsg struct {
	newModel Model
	err      error
}

// updateWorkspace handles updates when in workspace picker state
func (m Model) updateWorkspace(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case workspaceSwitchedMsg:
		// Workspace switch completed
		if msg.err != nil {
			// Switch failed - show error and return to list
			m.Err = msg.err
			m.State = StateList
			return m, nil
		}
		// Replace entire model with new model
		return msg.newModel, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "esc", "w":
			// Cancel and return to list view
			m.State = StateList
			return m, nil

		case "f":
			// Toggle favorite
			return m.handleToggleFavorite(), nil

		case "enter":
			// Switch to selected workspace
			return m.handleWorkspaceSelection()
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		// Note: list size is set in render function, not here
	}

	// Delegate to workspace list for navigation
	var cmd tea.Cmd
	m.WorkspaceList, cmd = m.WorkspaceList.Update(msg)

	return m, cmd
}

// handleWorkspaceSelection processes workspace selection and initiates switch
func (m Model) handleWorkspaceSelection() (Model, tea.Cmd) {
	selected := m.WorkspaceList.SelectedItem()
	if ws, ok := selected.(WorkspaceItem); ok {
		// Load config to get current settings
		cfg, err := config.Load()
		if err != nil {
			// Config load failed - show error
			m.Err = err
			m.State = StateList
			return m, nil
		}

		// Trigger workspace switch
		return m, switchWorkspaceCmd(ws.Workspace.Path, cfg, m)
	}
	return m, nil
}

// handleToggleFavorite toggles favorite status for selected workspace
func (m Model) handleToggleFavorite() Model {
	selected := m.WorkspaceList.SelectedItem()
	if ws, ok := selected.(WorkspaceItem); ok {
		if m.WorkspaceManager != nil {
			m.WorkspaceManager.ToggleFavorite(ws.Workspace.Path)
			_ = m.WorkspaceManager.Save()

			// Refresh workspace list to show updated favorite status
			m.refreshWorkspaceList()
		}
	}
	return m
}

// switchWorkspaceCmd performs the workspace switch asynchronously
func switchWorkspaceCmd(newPath string, cfg *config.Config, oldModel Model) tea.Cmd {
	return func() tea.Msg {
		// Create new model with new Makefile
		newModel := oldModel.SwitchWorkspace(newPath, cfg)

		// Return to list state after switch
		newModel.State = StateList

		return workspaceSwitchedMsg{
			newModel: newModel,
			err:      nil,
		}
	}
}

// initWorkspacePicker initializes the workspace picker list
func (m *Model) initWorkspacePicker() {
	if m.WorkspaceManager == nil {
		return
	}

	// Record current Makefile as accessed (ensures it appears in the list)
	m.WorkspaceManager.RecordAccess(m.MakefilePath)
	_ = m.WorkspaceManager.Save() // Async save, ignore errors (non-critical)

	cwd, _ := os.Getwd()

	// Discover Makefiles in the project (starting from cwd or Makefile directory)
	searchRoot := cwd
	if searchRoot == "" {
		searchRoot = "."
	}

	discovered, err := workspace.DiscoverMakefiles(searchRoot, workspace.DefaultDiscoveryOptions())
	if err != nil {
		// Fall back to recent-only if discovery fails
		discovered = []workspace.DiscoveryResult{}
	}

	// Build a map of all discovered Makefiles by path
	discoveredMap := make(map[string]workspace.DiscoveryResult)
	for _, result := range discovered {
		discoveredMap[result.Path] = result
	}

	// Get recent workspaces
	recent := m.WorkspaceManager.GetRecent(10)

	// Build combined list: recent first, then discovered (excluding duplicates)
	var items []list.Item

	// Add recent workspaces first
	recentPaths := make(map[string]bool)
	for _, ws := range recent {
		relPath := m.WorkspaceManager.GetRelativePath(ws.Path, cwd)
		items = append(items, WorkspaceItem{
			Workspace: ws,
			RelPath:   relPath,
		})
		recentPaths[ws.Path] = true
	}

	// Add discovered Makefiles that aren't already in recent
	for _, result := range discovered {
		if !recentPaths[result.Path] {
			// Create a workspace entry for discovered Makefile
			ws := workspace.Workspace{
				Path:         result.Path,
				LastAccessed: result.ModTime,
				AccessCount:  0,
				IsFavorite:   false,
			}
			relPath := m.WorkspaceManager.GetRelativePath(result.Path, cwd)
			items = append(items, WorkspaceItem{
				Workspace: ws,
				RelPath:   relPath,
			})
		}
	}

	// Create list with workspace delegate
	delegate := NewWorkspaceItemDelegate()
	l := list.New(items, delegate, 0, 0)
	l.Title = "Switch Workspace"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = TitleStyle
	// Note: list size is set in render function, not here

	m.WorkspaceList = l
}

// refreshWorkspaceList refreshes the workspace list items (e.g., after toggling favorite)
func (m *Model) refreshWorkspaceList() {
	if m.WorkspaceManager == nil {
		return
	}

	// Get recent workspaces
	recent := m.WorkspaceManager.GetRecent(10)

	// Convert to list items
	items := make([]list.Item, len(recent))
	cwd, _ := os.Getwd()
	for i, ws := range recent {
		// Compute relative path for display
		relPath := m.WorkspaceManager.GetRelativePath(ws.Path, cwd)

		items[i] = WorkspaceItem{
			Workspace: ws,
			RelPath:   relPath,
		}
	}

	// Update list items
	m.WorkspaceList.SetItems(items)
}

