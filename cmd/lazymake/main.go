package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rshelekhov/lazymake/config"
	"github.com/rshelekhov/lazymake/internal/presets"
	"github.com/rshelekhov/lazymake/internal/tui"
	"github.com/rshelekhov/lazymake/internal/workspace"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "lazymake",
	Short: "A beautiful TUI for running Makefile targets",
	Long:  `Lazymake is a terminal user interface for browsing and executing Makefile targets.`,
	RunE:  run,
}

func init() {
	rootCmd.Flags().StringP("file", "f", "Makefile", "Path to Makefile")
	rootCmd.Flags().Bool("dry-run", false, "Preview commands via 'make -n' without executing them")

	if err := viper.BindPFlag("makefile", rootCmd.Flags().Lookup("file")); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error binding makefile flag: %v\n", err)
		os.Exit(1)
	}

	if err := viper.BindPFlag("dry_run", rootCmd.Flags().Lookup("dry-run")); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error binding dry-run flag: %v\n", err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Initialize workspace manager
	workspaceMgr, err := workspace.Load()
	if err != nil {
		// Graceful degradation: continue with empty workspace manager
		workspaceMgr = workspace.NewEmpty()
	}

	// Initialize saved presets manager (issue #35). Failures here
	// degrade gracefully: an empty manager means "no presets yet",
	// the rest of the TUI works as before.
	presetsMgr, err := presets.Load()
	if err != nil {
		presetsMgr = presets.NewEmpty()
	}

	m := tui.NewModel(cfg)
	m.WorkspaceManager = workspaceMgr
	m.Presets = presetsMgr

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
