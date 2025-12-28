package tui

import (
	"os"
	"path/filepath"

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

		case "up", "k":
			// Navigate up, skip headers
			var cmd tea.Cmd
			m.WorkspaceList, cmd = m.WorkspaceList.Update(msg)
			m.skipHeadersUp()
			return m, cmd

		case "down", "j":
			// Navigate down, skip headers
			var cmd tea.Cmd
			m.WorkspaceList, cmd = m.WorkspaceList.Update(msg)
			m.skipHeadersDown()
			return m, cmd
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		// Note: list size is set in render function, not here
	}

	// Delegate to workspace list for other navigation/updates
	var cmd tea.Cmd
	m.WorkspaceList, cmd = m.WorkspaceList.Update(msg)

	return m, cmd
}

// skipHeadersDown moves cursor down past headers
func (m *Model) skipHeadersDown() {
	items := m.WorkspaceList.Items()
	index := m.WorkspaceList.Index()

	// Check if current item is a header, skip to next workspace
	for index < len(items) {
		if _, ok := items[index].(WorkspaceHeaderItem); ok {
			index++
			if index < len(items) {
				m.WorkspaceList.Select(index)
			}
		} else {
			break
		}
	}
}

// skipHeadersUp moves cursor up past headers
func (m *Model) skipHeadersUp() {
	items := m.WorkspaceList.Items()
	index := m.WorkspaceList.Index()

	// Check if current item is a header, skip to previous workspace
	for index >= 0 {
		if _, ok := items[index].(WorkspaceHeaderItem); ok {
			index--
			if index >= 0 {
				m.WorkspaceList.Select(index)
			}
		} else {
			break
		}
	}
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

	m.WorkspaceManager.RecordAccess(m.MakefilePath)
	_ = m.WorkspaceManager.Save()

	cwd, _ := os.Getwd()
	discovered := m.discoverWorkspaces(cwd)
	recent := m.WorkspaceManager.GetRecent(10)

	allWorkspaces := m.buildAllWorkspacesList(recent, discovered, cwd)
	items := buildWorkspaceListWithSections(allWorkspaces)

	m.WorkspaceList = m.createWorkspaceList(items)
}

// discoverWorkspaces finds all Makefiles in the project
func (m *Model) discoverWorkspaces(cwd string) []workspace.DiscoveryResult {
	searchRoot := cwd
	if searchRoot == "" {
		searchRoot = "."
	}

	discovered, err := workspace.DiscoverMakefiles(searchRoot, workspace.DefaultDiscoveryOptions())
	if err != nil {
		return []workspace.DiscoveryResult{}
	}
	return discovered
}

// buildAllWorkspacesList combines recent and discovered workspaces
func (m *Model) buildAllWorkspacesList(recent []workspace.Workspace, discovered []workspace.DiscoveryResult, cwd string) []WorkspaceItem {
	var allWorkspaces []WorkspaceItem
	recentPaths := make(map[string]bool)
	cwdBase := filepath.Base(cwd)

	// Add recent workspaces
	for _, ws := range recent {
		item := m.createWorkspaceItem(ws, cwd, cwdBase)
		allWorkspaces = append(allWorkspaces, item)
		recentPaths[ws.Path] = true
	}

	// Add discovered Makefiles that aren't in recent
	for _, result := range discovered {
		if !recentPaths[result.Path] {
			ws := workspace.Workspace{
				Path:         result.Path,
				LastAccessed: result.ModTime,
				AccessCount:  0,
				IsFavorite:   false,
			}
			item := m.createWorkspaceItem(ws, cwd, cwdBase)
			allWorkspaces = append(allWorkspaces, item)
		}
	}

	return allWorkspaces
}

// createWorkspaceItem creates a WorkspaceItem with formatted paths
func (m *Model) createWorkspaceItem(ws workspace.Workspace, cwd, cwdBase string) WorkspaceItem {
	relPath := m.WorkspaceManager.GetRelativePath(ws.Path, cwd)
	relDir := m.formatRelativeDir(filepath.Dir(relPath), cwdBase)

	return WorkspaceItem{
		Workspace: ws,
		RelPath:   filepath.Base(relPath),
		RelDir:    relDir,
	}
}

// formatRelativeDir formats the relative directory with root directory name
func (m *Model) formatRelativeDir(relDir, cwdBase string) string {
	switch {
	case relDir == ".":
		return "./" + cwdBase
	case len(relDir) > 2 && relDir[:2] == "./":
		return "./" + cwdBase + "/" + relDir[2:]
	case relDir != ".." && !filepath.IsAbs(relDir):
		return "./" + cwdBase + "/" + relDir
	default:
		return relDir
	}
}

// createWorkspaceList creates and configures the workspace list
func (m *Model) createWorkspaceList(items []list.Item) list.Model {
	delegate := NewWorkspaceItemDelegate()
	l := list.New(items, delegate, 0, 0)
	l.Title = "Switch Workspace"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = TitleStyle

	// Position cursor on first actual workspace (skip headers)
	for i, item := range items {
		if _, ok := item.(WorkspaceItem); ok {
			l.Select(i)
			break
		}
	}

	return l
}

// buildWorkspaceListWithSections organizes workspaces into sections with headers
func buildWorkspaceListWithSections(workspaces []WorkspaceItem) []list.Item {
	// Separate favorites from non-favorites
	var favorites []WorkspaceItem
	var nonFavorites []WorkspaceItem

	for _, ws := range workspaces {
		if ws.Workspace.IsFavorite {
			favorites = append(favorites, ws)
		} else {
			nonFavorites = append(nonFavorites, ws)
		}
	}

	// Build items list with sections
	items := make([]list.Item, 0, len(workspaces)+2)

	// Add favorites section if there are any
	if len(favorites) > 0 {
		items = append(items, WorkspaceHeaderItem{Label: "FAVORITES", WithSeparator: false})
		for _, ws := range favorites {
			items = append(items, ws)
		}
	}

	// Add all workspaces section
	if len(nonFavorites) > 0 {
		// Add separator before this section if there were favorites
		withSeparator := len(favorites) > 0
		items = append(items, WorkspaceHeaderItem{Label: "ALL WORKSPACES", WithSeparator: withSeparator})
		for _, ws := range nonFavorites {
			items = append(items, ws)
		}
	}

	return items
}

// refreshWorkspaceList refreshes the workspace list items (e.g., after toggling favorite)
func (m *Model) refreshWorkspaceList() {
	if m.WorkspaceManager == nil {
		return
	}

	// Get recent workspaces
	recent := m.WorkspaceManager.GetRecent(10)

	// Convert to workspace items
	var allWorkspaces []WorkspaceItem
	cwd, _ := os.Getwd()
	cwdBase := filepath.Base(cwd)

	for _, ws := range recent {
		// Compute relative path for display
		relPath := m.WorkspaceManager.GetRelativePath(ws.Path, cwd)
		relDir := filepath.Dir(relPath)

		// Add root directory name for current directory paths
		switch {
		case relDir == ".":
			relDir = "./" + cwdBase
		case len(relDir) > 2 && relDir[:2] == "./":
			relDir = "./" + cwdBase + "/" + relDir[2:]
		case relDir != ".." && !filepath.IsAbs(relDir):
			// Subdirectory without ./ prefix (e.g., "examples")
			relDir = "./" + cwdBase + "/" + relDir
		}

		allWorkspaces = append(allWorkspaces, WorkspaceItem{
			Workspace: ws,
			RelPath:   filepath.Base(relPath), // Just filename
			RelDir:    relDir,                  // Full relative path with root
		})
	}

	// Build list with sections
	items := buildWorkspaceListWithSections(allWorkspaces)

	// Update list items
	m.WorkspaceList.SetItems(items)
}

