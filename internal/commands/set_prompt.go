package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/gotoailab/llmgit/internal/config"
	"github.com/gotoailab/llmgit/internal/i18n"
)

// HandleSetPrompt handles the set-prompt command
func HandleSetPrompt(args []string) {
	// Check if initialized
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), i18n.T("error_not_initialized"))
		os.Exit(1)
	}

	if len(args) == 0 {
		// Show current prompt
		if cfg.CommitPrompt != "" {
			fmt.Println(i18n.T("prompt_current"))
			fmt.Println(cfg.CommitPrompt)
		} else {
			fmt.Println(i18n.T("prompt_not_set"))
			fmt.Println(i18n.T("prompt_usage"))
		}
		return
	}

	// Set prompt
	// Support reading from file or direct input
	prompt := strings.Join(args, " ")

	// If argument is a file path, read file content
	if len(args) == 1 {
		if _, err := os.Stat(args[0]); err == nil {
			// File exists, read file content
			data, err := os.ReadFile(args[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), fmt.Sprintf(i18n.T("error_read_file"), args[0], err))
				os.Exit(1)
			}
			prompt = strings.TrimSpace(string(data))
		}
	}

	cfg.CommitPrompt = prompt
	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), fmt.Sprintf(i18n.T("error_save_prompt"), err))
		os.Exit(1)
	}

	fmt.Println(i18n.T("prompt_saved"))
}

