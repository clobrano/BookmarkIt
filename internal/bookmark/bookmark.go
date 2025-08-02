package bookmark

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Bookmark represents a single bookmark entry.
type Bookmark struct {
	Key  string `yaml:"key"`
	Link string `yaml:"link"`
}

// Bookmarks holds a collection of Bookmark entries.
type Bookmarks struct {
	Bookmarks []Bookmark `yaml:"bookmarks"`
}

// Load reads bookmarks from the specified YAML file.
// If the file does not exist, it creates an empty bookmarks file.
func Load(filePath string) (*Bookmarks, error) {
	var bookmarks Bookmarks

	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		emptyBookmarks := Bookmarks{Bookmarks: []Bookmark{}}
		err = Save(&emptyBookmarks, filePath) // Use Save to create the empty file
		if err != nil {
			return nil, fmt.Errorf("failed to create initial bookmarks file at %s: %w", filePath, err)
		}
		return &emptyBookmarks, nil // Return the empty bookmarks
	} else if err != nil {
		return nil, fmt.Errorf("failed to check if bookmarks file exists at %s: %w", filePath, err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read bookmarks file at %s: %w", filePath, err)
	}

	err = yaml.Unmarshal(data, &bookmarks)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal bookmarks from %s: %w", filePath, err)
	}

	return &bookmarks, nil
}

// Save writes the given bookmarks to the specified YAML file.
func Save(bookmarks *Bookmarks, filePath string) error {
	data, err := yaml.Marshal(bookmarks)
	if err != nil {
		return fmt.Errorf("failed to marshal bookmarks: %w", err)
	}

	// Ensure the directory exists before writing the file
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory for bookmark file %s: %w", dir, err)
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write bookmarks file to %s: %w", filePath, err)
	}

	return nil
}

// HasLink checks if a bookmark with the given link already exists in the collection.
func (b *Bookmarks) HasLink(link string) bool {
	for _, bm := range b.Bookmarks {
		if bm.Link == link {
			return true
		}
	}
	return false
}

// Add adds a new bookmark to the collection. It does not check for duplicates.
func (b *Bookmarks) Add(key, link string) {
	newBookmark := Bookmark{Key: key, Link: link}
	b.Bookmarks = append(b.Bookmarks, newBookmark)
}
