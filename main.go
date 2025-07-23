package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	bookmarkFilePath string
	actionFlag       string
	queryFlag        string
)

const (
	bookmarksFile = "$HOME/Documents/bookmarks.yml"
	dialogTitle   = "BookmarkIt"
	dialogText    = "Please enter the your bookmark and a key"
	separator     = "___"
)

type Bookmark struct {
	Key  string `yaml:"key"`
	Link string `yaml:"link"`
}

type Bookmarks struct {
	Bookmarks []Bookmark `yaml:"bookmarks"`
}

func main() {
	defaultBookmarksFile, err := getDefaultBookmarksFilePath()
	if err != nil {
		os.Exit(1)
	}
	flag.StringVar(&bookmarkFilePath, "file", defaultBookmarksFile, "Path to the bookmarks YAML file")
	flag.StringVar(&actionFlag, "action", "find", "Action to perform: 'add' or 'find'")
	flag.StringVar(&queryFlag, "query", "", "Query string for 'find' action")
	flag.Parse()

	// Ensure the directory for the bookmark file exists
	bookmarkDir := filepath.Dir(bookmarkFilePath)
	if _, err := os.Stat(bookmarkDir); os.IsNotExist(err) {
		notify(fmt.Sprintf("Failed to find bookmark directory \"%s\": %v", bookmarkDir, err))
		os.Exit(1)
	}

	run(actionFlag, queryFlag)
}

func getDefaultBookmarksFilePath() (string, error) {
	// Determine default config directory (XDG_CONFIG_HOME or $HOME/.config)
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback if os.UserConfigDir fails
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Error getting home directory:", err)
			return "", err
		}
		userConfigDir = filepath.Join(homeDir, ".config")
	}

	// Construct the default bookmarks file path
	return filepath.Join(userConfigDir, "bookmarkit", "bookmarks.yml"), nil
}

func run(action, query string) {
	switch action {
	case "add":
		if query != "" {
			fmt.Printf("[!] You don't need the query flag with the 'add' command")
		}
		addBookmark()
	case "find":
		findBookmark(query)
	default:
		fmt.Printf("[!] unsupported action \"%s\"", action)
		os.Exit(1)
	}
}

func addBookmark() {
	key, book := getYADInput()
	if key == "" {
		notify("failed: key is empty")
		os.Exit(1)
	}
	if book == "" {
		notify("failed: bookmark is empty")
		os.Exit(1)
	}

	key = strings.ReplaceAll(key, " ", "_")

	bookmarks, err := loadBookmarks()
	if err != nil {
		notify(fmt.Sprintf("Failed to load bookmarks: %v", err))
		os.Exit(1)
	}

	// Check if bookmark already exists
	for _, bm := range bookmarks.Bookmarks {
		if bm.Link == book {
			notify("This URL was already bookmarked")
			return
		}
	}

	newBookmark := Bookmark{Key: key, Link: book}
	bookmarks.Bookmarks = append(bookmarks.Bookmarks, newBookmark)

	err = saveBookmarks(bookmarks)
	if err != nil {
		notify(fmt.Sprintf("Failed to save bookmarks: %v", err))
		os.Exit(1)
	}

	notify(fmt.Sprintf("%s stored in Bookmark", key))
}

func findBookmark(query string) {
	bookmarks, err := loadBookmarks()
	if err != nil {
		notify(fmt.Sprintf("Failed to load bookmarks: %v", err))
		os.Exit(1)
	}

	var options []string
	for _, bm := range bookmarks.Bookmarks {
		options = append(options, bm.Key+" => "+bm.Link)
	}

	selection := showFZF(options, query)
	if selection == "" {
		os.Exit(0)
	}

	parts := strings.Split(selection, " => ")
	if len(parts) != 2 {
		notify("Invalid selection format")
		os.Exit(1)
	}

	link := parts[1]
	if strings.HasPrefix(link, "https://") {
		err := openURL(link)
		if err != nil {
			notify(fmt.Sprintf("Failed to open URL: %v", err))
			os.Exit(1)
		}
	} else {
		copyToClipboard(link)
		notify(fmt.Sprintf("added to clipboard: '%s'", link))
	}
}

func getYADInput() (string, string) {
	clipboardContent := getClipboardContent()

	cmd := exec.Command("yad",
		"--form",
		"--title", dialogTitle,
		"--text", dialogText,
		"--width=650",
		"--height=150",
		fmt.Sprintf("--separator=%s", separator),
		"--field=Key", "",
		"--field=Bookmark", fmt.Sprintf("%s", clipboardContent))

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("YAD error: %v\n", err)
		return "", ""
	}

	result := strings.TrimSpace(string(out))

	for i, r := range result {
		fmt.Printf("%d:%U ", i, r)
	}
	fmt.Println()

	separatorIndex := strings.Index(result, separator)

	if separatorIndex == -1 {
		fmt.Printf("Separator '%s' not found in result\n", separator)
		return "", ""
	}

	key := result[:separatorIndex]
	book := result[separatorIndex+len(separator):]
	book = strings.TrimSuffix(book, separator)

	return key, book
}

func showFZF(options []string, query string) string {
	fzfCmd := exec.Command("fzf",
		"--prompt", "Search bookmark > ",
		"--layout", "reverse",
		"--height=70%",
		"--bind", "ctrl-y:execute(readlink -f {} | echo {} | cut -d'>' -f2 | tr -d '\\n' | xclip -selection clipboard)+abort")

	if query != "" {
		fzfCmd.Args = append(fzfCmd.Args, "--query="+query)
	}

	fzfIn, _ := fzfCmd.StdinPipe()
	fzfOut, _ := fzfCmd.StdoutPipe()

	err := fzfCmd.Start()
	if err != nil {
		notify(fmt.Sprintf("Failed to start fzf: %v", err))
		os.Exit(1)
	}

	for _, opt := range options {
		fmt.Fprintln(fzfIn, opt)
	}
	fzfIn.Close()

	outputBytes, err := io.ReadAll(fzfOut)
	if err != nil {
		fmt.Printf("Error reading from fzf output: %v\n", err)
		return ""
	}
	fzfCmd.Wait()

	return strings.TrimSpace(string(outputBytes))
}

func loadBookmarks() (Bookmarks, error) {
	var bookmarks Bookmarks
	filePath := os.ExpandEnv(bookmarksFile)

	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		// Create the file if it doesn't exist
		emptyBookmarks := Bookmarks{Bookmarks: []Bookmark{}}
		err = saveBookmarks(emptyBookmarks)
		if err != nil {
			return bookmarks, fmt.Errorf("failed to create bookmarks file: %w", err)
		}
	} else if err != nil {
		return bookmarks, fmt.Errorf("failed to check if bookmarks file exists: %w", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return bookmarks, fmt.Errorf("failed to read bookmarks file: %w", err)
	}

	err = yaml.Unmarshal(data, &bookmarks)
	if err != nil {
		return bookmarks, fmt.Errorf("failed to unmarshal bookmarks: %w", err)
	}

	return bookmarks, nil
}

func saveBookmarks(bookmarks Bookmarks) error {
	filePath := os.ExpandEnv(bookmarksFile)
	data, err := yaml.Marshal(bookmarks)
	if err != nil {
		return fmt.Errorf("failed to marshal bookmarks: %w", err)
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write bookmarks file: %w", err)
	}

	return nil
}

func openURL(url string) error {
	cmd := exec.Command("xdg-open", url)
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to open URL: %w", err)
	}
	return nil
}

func copyToClipboard(text string) {
	var cmd *exec.Cmd
	if _, err := exec.LookPath("wl-copy"); err == nil {
		cmd = exec.Command("wl-copy")
	} else if _, err := exec.LookPath("xclip"); err == nil {
		cmd = exec.Command("xclip", "-selection", "clipboard")
	} else {
		notify("Neither wl-copy nor xclip is available.")
		return
	}

	pipe, err := cmd.StdinPipe()
	if err != nil {
		notify(fmt.Sprintf("Error getting stdin pipe: %v", err))
		return
	}
	if err := cmd.Start(); err != nil {
		notify(fmt.Sprintf("Error starting command: %v", err))
		return
	}
	_, err = pipe.Write([]byte(text))
	if err != nil {
		notify(fmt.Sprintf("Error writing to pipe: %v", err))
		return
	}
	pipe.Close()
	err = cmd.Wait()
	if err != nil {
		notify(fmt.Sprintf("Error waiting for command: %v", err))
	}
}

func notify(message string) {
	if _, err := exec.LookPath("notify-send"); err == nil {
		cmd := exec.Command("notify-send", "--app-name", "BookMarkIt", "-i", "dialog-information", message)
		err := cmd.Run()
		if err != nil {
			fmt.Printf("Error sending notification: %v\n", err)
		}
	} else {
		fmt.Println(message)
	}
}

func getClipboardContent() string {
	var cmd *exec.Cmd
	if _, err := exec.LookPath("wl-paste"); err == nil {
		cmd = exec.Command("wl-paste", "--no-newline")
	} else if _, err := exec.LookPath("xclip"); err == nil {
		cmd = exec.Command("xclip", "-selection", "clipboard", "-o")
	} else {
		notify("Neither wl-paste nor xclip is available.")
		return ""
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error getting clipboard content: %v\n", err)
		return ""
	}

	return strings.TrimSpace(string(out))
}
