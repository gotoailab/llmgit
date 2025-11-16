package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/gotoailab/llmgit/internal/ai"
	"github.com/gotoailab/llmgit/internal/config"
	"github.com/gotoailab/llmgit/internal/git"
	"github.com/gotoailab/llmgit/internal/i18n"
	"github.com/gotoailab/llmhub"
)

// HandleExplain handles the explain command
func HandleExplain(client *llmhub.Client, cfg *config.Config, args []string) {
	// Get diff
	var diff string
	var err error
	if len(args) > 0 {
		diff, err = git.GetFileDiff(args[0])
	} else {
		diff, err = git.GetDiff()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), fmt.Sprintf(i18n.T("error_get_changes"), err))
		os.Exit(1)
	}

	if diff == "" {
		fmt.Println(i18n.T("explain_no_changes"))
		return
	}

	// Use AI to explain
	fmt.Println(i18n.T("explain_generating"))
	ctx := context.Background()
	explanation, err := ai.ExplainDiff(ctx, client, diff)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), fmt.Sprintf(i18n.T("error_generate_explanation"), err))
		os.Exit(1)
	}

	fmt.Println(explanation)
}

