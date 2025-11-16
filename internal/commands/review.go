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

// HandleReview handles the review command
func HandleReview(client *llmhub.Client, cfg *config.Config, args []string) {
	commit := "HEAD"
	if len(args) > 0 {
		commit = args[0]
	}

	// Get commit information
	commitInfo, err := git.GetCommitInfo(commit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), fmt.Sprintf(i18n.T("error_get_commit_info"), err))
		os.Exit(1)
	}

	// Use AI to review
	fmt.Println(i18n.T("review_generating"))
	ctx := context.Background()
	review, err := ai.ReviewCommit(ctx, client, commitInfo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), fmt.Sprintf(i18n.T("error_review_commit"), err))
		os.Exit(1)
	}

	fmt.Println(review)
}

