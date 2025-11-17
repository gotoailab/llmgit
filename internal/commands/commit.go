package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/gotoailab/llmgit/internal/ai"
	"github.com/gotoailab/llmgit/internal/config"
	"github.com/gotoailab/llmgit/internal/git"
	"github.com/gotoailab/llmgit/internal/i18n"
	"github.com/gotoailab/llmhub"
)

// HandleCommit handles the commit command
func HandleCommit(client *llmhub.Client, cfg *config.Config, args []string) {
	// Check if user already provided -m parameter (user provided message)
	hasMessage := false
	for i, arg := range args {
		if arg == "-m" && i+1 < len(args) {
			hasMessage = true
			break
		}
	}

	// If user already provided message, execute git commit directly
	if hasMessage {
		gitArgs := []string{"commit"}
		gitArgs = append(gitArgs, args...)
		handleGitCommand(gitArgs)
		return
	}

	// Parse language option (--lang or -l)
	lang := "en" // Default English
	filteredArgs := []string{}
	skipNext := false
	for i, arg := range args {
		if skipNext {
			skipNext = false
			continue
		}
		if arg == "--lang" || arg == "-l" {
			if i+1 < len(args) {
				lang = args[i+1]
				skipNext = true
				continue
			}
		} else if strings.HasPrefix(arg, "--lang=") {
			lang = strings.TrimPrefix(arg, "--lang=")
		} else {
			filteredArgs = append(filteredArgs, arg)
		}
	}

	// Get staged changes
	diff, err := git.GetStagedDiff()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), fmt.Sprintf(i18n.T("error_get_changes"), err))
		os.Exit(1)
	}

	if diff == "" {
		fmt.Println(i18n.T("commit_no_staged"))
		os.Exit(0)
	}

	// Filter out vendor and node_modules
	diff = git.FilterDiffExcludes(diff, nil)

	// Analyze changed file types for type suggestion
	files, err := git.GetStagedFiles()
	if err != nil {
		// If getting file list fails, continue without type suggestion
		files = []string{}
	}
	typeHint := ai.AnalyzeChangeType(files)

	// Use AI to generate commit message
	fmt.Println(i18n.T("commit_generating"))
	ctx := context.Background()
	message, err := ai.GenerateCommitMessage(ctx, client, diff, lang, cfg, typeHint)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), fmt.Sprintf(i18n.T("error_generate_message"), err))
		os.Exit(1)
	}

	fmt.Printf("\n%s\n\n%s\n\n", i18n.T("commit_generated"), message)

	// Check for --no-edit parameter (skip confirmation)
	skipConfirm := false
	for _, arg := range filteredArgs {
		if arg == "--no-edit" {
			skipConfirm = true
			break
		}
	}

	// Ask for confirmation
	if !skipConfirm {
		fmt.Print(i18n.T("commit_confirm"))
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "y" && confirm != "Y" {
			fmt.Println(i18n.T("commit_cancelled"))
			os.Exit(0)
		}
	}

	// Execute git commit
	gitArgs := []string{"commit", "-m", message}
	// Filter out --no-edit and --lang related parameters
	for _, arg := range filteredArgs {
		if arg != "--no-edit" && !strings.HasPrefix(arg, "--lang") {
			gitArgs = append(gitArgs, arg)
		}
	}
	handleGitCommand(gitArgs)
}

// handleGitCommand forwards command to native git
func handleGitCommand(args []string) {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}

