package system

import (
	"fmt"
	"os/exec"
	"strings"
)

// Notify sends a desktop notification using notify-send or prints to console as a fallback.
func Notify(message string) {
	if _, err := exec.LookPath("notify-send"); err == nil {
		cmd := exec.Command("notify-send", "--app-name", "BookMarkIt", "-i", "dialog-information", message)
		err := cmd.Run()
		if err != nil {
			fmt.Printf("Error sending notification via notify-send: %v\n", err)
		}
	} else {
		fmt.Println(message)
	}
}

// OpenURL attempts to open the given URL using xdg-open.
func OpenURL(url string) error {
	cmd := exec.Command("xdg-open", url)
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start xdg-open: %w", err)
	}
	return nil
}

// CopyToClipboard copies the given text to the system clipboard using wl-copy or xclip.
func CopyToClipboard(text string) error {
	var cmd *exec.Cmd
	if _, err := exec.LookPath("wl-copy"); err == nil {
		cmd = exec.Command("wl-copy")
	} else if _, err := exec.LookPath("xclip"); err == nil {
		cmd = exec.Command("xclip", "-selection", "clipboard")
	} else {
		return fmt.Errorf("neither wl-copy nor xclip is available to copy to clipboard")
	}

	pipe, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("error getting stdin pipe for clipboard command: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting clipboard command: %w", err)
	}
	_, err = pipe.Write([]byte(text))
	if err != nil {
		pipe.Close()
		return fmt.Errorf("error writing to clipboard pipe: %w", err)
	}
	pipe.Close()
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("error waiting for clipboard command: %w", err)
	}
	return nil
}

// GetClipboardContent retrieves text from the system clipboard using wl-paste or xclip.
func GetClipboardContent() (string, error) {
	var cmd *exec.Cmd
	if _, err := exec.LookPath("wl-paste"); err == nil {
		cmd = exec.Command("wl-paste", "--no-newline")
	} else if _, err := exec.LookPath("xclip"); err == nil {
		cmd = exec.Command("xclip", "-selection", "clipboard", "-o")
	} else {
		return "", fmt.Errorf("neither wl-paste nor xclip is available to get clipboard content")
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error executing clipboard content command: %w, output: %s", err, string(out))
	}

	return strings.TrimSpace(string(out)), nil
}
