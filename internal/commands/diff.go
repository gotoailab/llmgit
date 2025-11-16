package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/gotoailab/llmgit/internal/ai"
	"github.com/gotoailab/llmgit/internal/config"
	"github.com/gotoailab/llmgit/internal/git"
	"github.com/gotoailab/llmgit/internal/i18n"
	"github.com/gotoailab/llmhub"
)

// HandleDiff handles the diff command
func HandleDiff(client *llmhub.Client, cfg *config.Config, args []string) {
	// Get diff
	diff, err := git.GetDiff(args...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), fmt.Sprintf(i18n.T("error_get_changes"), err))
		os.Exit(1)
	}

	if diff == "" {
		fmt.Println(i18n.T("diff_no_changes"))
		return
	}

	// Show diff first
	cmd := exec.Command("git", append([]string{"diff"}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}

	// Use AI to explain
	fmt.Println("\n" + i18n.T("diff_explanation"))
	ctx := context.Background()
	explanation, err := ai.ExplainDiff(ctx, client, diff)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), fmt.Sprintf(i18n.T("error_generate_explanation"), err))
		os.Exit(1)
	}

	fmt.Println(explanation)
}

