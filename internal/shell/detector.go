package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DetectShell returns the user's shell type based on $SHELL environment variable
func DetectShell() string {
	// Check SHELL environment variable
	shell := os.Getenv("SHELL")

	// Parse shell name from path
	if strings.Contains(shell, "bash") {
		return "bash"
	}
	if strings.Contains(shell, "zsh") {
		return "zsh"
	}
	if strings.Contains(shell, "fish") {
		return "fish"
	}

	// Default to none if unknown
	return "none"
}

// GetHistoryFile returns the default history file path for a shell type
func GetHistoryFile(shellType string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch shellType {
	case "bash":
		return filepath.Join(home, ".bash_history"), nil

	case "zsh":
		// Check $HISTFILE first
		if histfile := os.Getenv("HISTFILE"); histfile != "" {
			return histfile, nil
		}
		return filepath.Join(home, ".zsh_history"), nil

	case "fish":
		return filepath.Join(home, ".local/share/fish/fish_history"), nil

	default:
		return "", fmt.Errorf("unknown shell type: %s", shellType)
	}
}
