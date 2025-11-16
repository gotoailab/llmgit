package commands

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gotoailab/llmgit/internal/ai"
	"github.com/gotoailab/llmgit/internal/git"
	"github.com/gotoailab/llmgit/internal/i18n"
	"github.com/gotoailab/llmgit/internal/utils"
	"github.com/gotoailab/llmhub"
)

// HandlePR handles the pr command
func HandlePR(client *llmhub.Client, args []string) {
	// Parse arguments
	var baseBranch string = "main"
	var targetBranch string = "HEAD"
	var copyToClipboard bool = false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--base" || arg == "-b" {
			if i+1 < len(args) {
				baseBranch = args[i+1]
				i++
			}
		} else if arg == "--copy" || arg == "-c" {
			copyToClipboard = true
		} else if !strings.HasPrefix(arg, "--") {
			if targetBranch == "HEAD" {
				targetBranch = arg
			} else {
				baseBranch = arg
			}
		}
	}

	// Get branch diff
	diff, err := git.GetBranchDiffFull(baseBranch, targetBranch)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), fmt.Sprintf(i18n.T("error_get_branch_diff"), err))
		os.Exit(1)
	}

	if diff == "" {
		fmt.Println(i18n.T("pr_no_changes"))
		return
	}

	// Get commit list
	commits, err := git.GetBranchCommits(baseBranch, targetBranch)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), fmt.Sprintf(i18n.T("error_get_commits"), err))
		os.Exit(1)
	}

	// Use AI to generate PR description
	fmt.Println(i18n.T("pr_generating"))
	ctx := context.Background()
	prDesc, err := ai.GeneratePRDescription(ctx, client, baseBranch, targetBranch, commits, diff)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), fmt.Sprintf(i18n.T("error_generate_pr"), err))
		os.Exit(1)
	}

	// Output or copy to clipboard
	if copyToClipboard {
		// Try to copy to clipboard (requires system support)
		err := utils.CopyToClipboard(prDesc)
		if err != nil {
			fmt.Println(prDesc)
			fmt.Fprintf(os.Stderr, "\n%s%s\n", i18n.T("error_prefix"), fmt.Sprintf(i18n.T("error_copy_clipboard"), err))
		} else {
			fmt.Println(i18n.T("pr_copied"))
		}
	} else {
		fmt.Println(prDesc)
	}
}

