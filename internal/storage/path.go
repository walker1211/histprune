package storage

import (
	"os"
	"path/filepath"
)

// DefaultHistoryPath returns the zsh history file path from HISTFILE or HOME.
func DefaultHistoryPath() (string, error) {
	if histfile := os.Getenv("HISTFILE"); histfile != "" {
		return histfile, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".zsh_history"), nil
}
