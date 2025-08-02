package ui

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
)

const (
	// DialogTitle is the title for the YAD dialog.
	DialogTitle = "BookmarkIt"
	// DialogText is the text prompt for the YAD dialog.
	DialogText = "Please enter your bookmark and a key"
	// Separator is used by YAD to separate form fields in its output.
	Separator = "___"
)

// GetYADInput prompts the user for a key and a bookmark link using YAD.
// It pre-fills the bookmark field with clipboardContent.
func GetYADInput(clipboardContent string) (key, bookmark string, err error) {
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

// ShowFZF presents a list of options to the user via FZF for fuzzy searching.
func ShowFZF(options []string, query string) (string, error) {
	fzfCmd := exec.Command("fzf",
		"--prompt", "Search bookmark > ",
		"--layout", "reverse",
		"--height=70%",
		"--bind", "ctrl-y:execute(readlink -f {} | echo {} | cut -d'>' -f2 | tr -d '\\n' | xclip -selection clipboard)+abort")

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
		return "", fmt.Errorf("fzf command failed: %w", err)
	}

	return strings.TrimSpace(string(outputBytes)), nil
}
