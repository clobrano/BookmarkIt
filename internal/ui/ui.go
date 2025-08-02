package ui

import (
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	// DialogTitle is the title for the YAD dialog.
	DialogTitle = "BookmarkIt"
	// DialogText is the text prompt for the YAD dialog.
	DialogText = "Please enter your bookmark and a key"
	// Separator is used by YAD to separate form fields in its output.
	Separator = "___"
)

// init is a special Go function that runs before main.
// We use it to set a global tview theme.
func init() {
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault
	tview.Styles.ContrastBackgroundColor = tcell.ColorDefault
	tview.Styles.PrimaryTextColor = tcell.ColorWhite
	tview.Styles.InverseTextColor = tcell.ColorWhite
	tview.Styles.BorderColor = tcell.ColorWhite
	//tview.Styles.FocusColor = tcell.ColorBlue
	tview.Styles.TitleColor = tcell.ColorWhite
}

// GetYADInput prompts the user for a key and a bookmark link using YAD.
// It pre-fills the bookmark field with clipboardContent.
func GetYADInput(clipboardContent string) (key, bookmark string, err error) {
	// First, check if YAD is installed.
	if !isCommandAvailable("yad") {
		return getTUIInput(clipboardContent)
	}

	cmd := exec.Command("yad",
		"--form",
		"--title", DialogTitle,
		"--text", DialogText,
		"--width=650",
		"--height=150",
		fmt.Sprintf("--separator=%s", Separator),
		"--field=Key", "",
		"--field=Bookmark", fmt.Sprintf("%s", clipboardContent))

	out, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 { // YAD returns 1 for cancel
			return "", "", fmt.Errorf("YAD input cancelled: %w", exitErr)
		}
		return "", "", fmt.Errorf("YAD error: %w, output: %s", err, string(out))
	}

	result := strings.TrimSpace(string(out))

	separatorIndex := strings.Index(result, Separator)
	if separatorIndex == -1 {
		return "", "", fmt.Errorf("separator '%s' not found in YAD result: '%s'", Separator, result)
	}

	key = result[:separatorIndex]
	bookmark = strings.TrimSuffix(result[separatorIndex+len(Separator):], Separator)

	return key, bookmark, nil
}

// getTUIInput provides a fallback TUI for adding bookmarks if YAD is not available.
// This version uses a global theme for consistent colors.
func getTUIInput(clipboardContent string) (key, bookmark string, err error) {
	app := tview.NewApplication()

	// The form will inherit the global theme, so we only need to set
	// the specific colors for the input fields to be black background / white text.
	form := tview.NewForm().
		SetFieldTextColor(tcell.ColorWhite).
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetButtonTextColor(tcell.ColorWhite).
		SetButtonBackgroundColor(tcell.ColorDefault)

	keyInput := tview.NewInputField().SetLabel("Key")
	bookmarkInput := tview.NewInputField().SetLabel("Bookmark").SetText(clipboardContent)

	form.AddFormItem(keyInput)
	form.AddFormItem(bookmarkInput)

	form.AddButton("Save", func() {
		key = keyInput.GetText()
		bookmark = bookmarkInput.GetText()
		app.Stop()
	})
	form.AddButton("Cancel", func() {
		key = ""
		bookmark = ""
		err = fmt.Errorf("TUI input cancelled")
		app.Stop()
	})

	form.SetBorder(true).SetTitle("Add New Bookmark").SetTitleAlign(tview.AlignLeft)

	// The application's background will be set by the global theme to ColorDefault.
	if err := app.SetRoot(form, true).SetFocus(form).Run(); err != nil {
		return "", "", fmt.Errorf("TUI application error: %w", err)
	}

	return key, bookmark, err
}

// ShowFZF presents a list of options to the user via FZF for fuzzy searching.
func ShowFZF(options []string, query string) (string, error) {
	// First, check if fzf is installed.
	if !isCommandAvailable("fzf") {
		return "", fmt.Errorf("fzf is not installed. Please install it to use this feature")
	}

	fzfCmd := exec.Command("fzf",
		"--prompt", "Search bookmark > ",
		"--layout", "reverse",
		"--height=70%",
		"--bind", "ctrl-y:execute(echo {} | cut -d' ' -f1 | xclip -selection clipboard)+abort")

	if query != "" {
		fzfCmd.Args = append(fzfCmd.Args, "--query="+query)
	}

	fzfIn, err := fzfCmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("failed to get stdin pipe for fzf: %w", err)
	}
	fzfOut, err := fzfCmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to get stdout pipe for fzf: %w", err)
	}

	err = fzfCmd.Start()
	if err != nil {
		return "", fmt.Errorf("failed to start fzf: %w", err)
	}

	for _, opt := range options {
		fmt.Fprintln(fzfIn, opt)
	}
	fzfIn.Close()

	outputBytes, err := io.ReadAll(fzfOut)
	if err != nil {
		return "", fmt.Errorf("error reading from fzf output: %w", err)
	}

	err = fzfCmd.Wait()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 130 {
			// fzf returns 130 on Ctrl-C
			return "", nil // No selection, not an error
		}
		return "", fmt.Errorf("fzf command failed: %w", err)
	}

	return strings.TrimSpace(string(outputBytes)), nil
}

// isCommandAvailable checks if a command exists in the system's PATH.
func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

