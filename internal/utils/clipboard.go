package utils

import (
	"fmt"
	"os/exec"
	"strings"
)

// CopyToClipboard copies text to clipboard (cross-platform)
func CopyToClipboard(text string) error {
	// Detect system and execute corresponding copy command
	var cmd *exec.Cmd
	if _, err := exec.LookPath("pbcopy"); err == nil {
		// macOS
		cmd = exec.Command("pbcopy")
	} else if _, err := exec.LookPath("xclip"); err == nil {
		// Linux with xclip
		cmd = exec.Command("xclip", "-selection", "clipboard")
	} else if _, err := exec.LookPath("xsel"); err == nil {
		// Linux with xsel
		cmd = exec.Command("xsel", "--clipboard", "--input")
	} else if _, err := exec.LookPath("clip.exe"); err == nil {
		// Windows
		cmd = exec.Command("clip.exe")
	} else {
		return fmt.Errorf("no clipboard utility found")
	}

	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

