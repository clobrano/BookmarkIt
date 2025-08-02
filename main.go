package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	osexec "os/exec"
	"path/filepath"
	"strings"

	"github.com/clobrano/BookmarkIt/internal/bookmark"
	"github.com/clobrano/BookmarkIt/internal/config"
	"github.com/clobrano/BookmarkIt/internal/system"
	"github.com/clobrano/BookmarkIt/internal/ui"
)

var (
	bookmarkFilePath string
	actionFlag       string
	queryFlag        string
)

func main() {
	// 1. Get default bookmark file path
	defaultBookmarksFile, err := config.GetDefaultBookmarksFilePath()
	if err != nil {
		system.Notify(fmt.Sprintf("Error getting default bookmark file path: %v", err))
		os.Exit(1)
	}

	// 2. Parse flags
	flag.StringVar(&bookmarkFilePath, "file", defaultBookmarksFile, "Path to the bookmarks YAML file")
	flag.StringVar(&actionFlag, "action", "find", "Action to perform: 'add' or 'find'")
	flag.StringVar(&queryFlag, "query", "", "Query string for 'find' action")
	flag.Parse()

	// 3. Ensure bookmark directory exists, create if not
	bookmarkDir := filepath.Dir(bookmarkFilePath)
	if _, err := os.Stat(bookmarkDir); os.IsNotExist(err) {
		if err := os.MkdirAll(bookmarkDir, 0755); err != nil {
			system.Notify(fmt.Sprintf("Failed to create bookmark directory \"%s\": %v", bookmarkDir, err))
			os.Exit(1)
		}
	} else if err != nil {
		system.Notify(fmt.Sprintf("Failed to check bookmark directory \"%s\": %v", bookmarkDir, err))
		os.Exit(1)
	}

	// 4. Run the main action
	if err := run(bookmarkFilePath, actionFlag, queryFlag); err != nil {
		system.Notify(fmt.Sprintf("Operation failed: %v", err))
		os.Exit(1)
	}
}

func run(filePath, action, query string) error {
	switch action {
	case "add":
		if query != "" {
			system.Notify("[!] You don't need the query flag with the 'add' command")
		}
		return addBookmark(filePath)
	case "find":
		return findBookmark(filePath, query)
	default:
		return fmt.Errorf("[!] unsupported action \"%s\"", action)
	}
}

func addBookmark(filePath string) error {
	clipboardContent, err := system.GetClipboardContent()
	if err != nil {
		system.Notify(fmt.Sprintf("Could not get clipboard content: %v", err))
		// Continue without clipboard content
		clipboardContent = ""
	}

	key, link, err := ui.GetYADInput(clipboardContent)
	if err != nil {
		return fmt.Errorf("failed to get YAD input: %w", err)
	}
	if key == "" {
		return fmt.Errorf("failed: key is empty")
	}
	if link == "" {
		return fmt.Errorf("failed: bookmark is empty")
	}

	key = strings.ReplaceAll(key, " ", "_")

	bookmarks, err := bookmark.Load(filePath)
	if err != nil {
		return fmt.Errorf("failed to load bookmarks: %w", err)
	}

	if bookmarks.HasLink(link) {
		system.Notify("This URL was already bookmarked")
		return nil
	}

	bookmarks.Add(key, link)

	err = bookmark.Save(bookmarks, filePath)
	if err != nil {
		return fmt.Errorf("failed to save bookmarks: %w", err)
	}

	system.Notify(fmt.Sprintf("'%s' stored in Bookmark", key))
	return nil
}

func findBookmark(filePath, query string) error {
	bookmarks, err := bookmark.Load(filePath)
	if err != nil {
		return fmt.Errorf("failed to load bookmarks: %w", err)
	}

	var options []string
	for _, bm := range bookmarks.Bookmarks {
		options = append(options, bm.Key+" => "+bm.Link)
	}

	selection, err := ui.ShowFZF(options, query)
	if err != nil {
		var exitErr *osexec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 130 { // 130 is common for Ctrl+C exit
			return nil // User cancelled FZF, not an error
		}
		return fmt.Errorf("failed to get FZF selection: %w", err)
	}
	if selection == "" {
		return nil // User selected nothing or FZF returned empty
	}

	parts := strings.Split(selection, " => ")
	if len(parts) != 2 {
		return fmt.Errorf("invalid selection format from FZF: %s", selection)
	}

	link := parts[1]
	if strings.HasPrefix(link, "https://") || strings.HasPrefix(link, "http://") {
		err := system.OpenURL(link)
		if err != nil {
			return fmt.Errorf("failed to open URL: %w", err)
		}
		system.Notify(fmt.Sprintf("Opened URL: '%s'", link))
	} else {
		err := system.CopyToClipboard(link)
		if err != nil {
			return fmt.Errorf("failed to copy to clipboard: %w", err)
		}
		system.Notify(fmt.Sprintf("Copied to clipboard: '%s'", link))
	}
	return nil
}
