package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetDefaultBookmarksFilePath determines the default path for the bookmarks YAML file.
// It follows the XDG Base Directory Specification, falling back to $HOME/.config.
func GetDefaultBookmarksFilePath() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("error getting user home directory: %w", err)
		}
		userConfigDir = filepath.Join(homeDir, ".config")
	}

	return filepath.Join(userConfigDir, "bookmarkit", "bookmarks.yml"), nil
}
