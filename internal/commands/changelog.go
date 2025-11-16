package commands

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gotoailab/llmgit/internal/ai"
	"github.com/gotoailab/llmgit/internal/git"
	"github.com/gotoailab/llmgit/internal/i18n"
	"github.com/gotoailab/llmhub"
)

// HandleChangelog handles the changelog command
func HandleChangelog(client *llmhub.Client, args []string) {
	// Parse arguments
	var rangeSpec string
	var outputFile string
	var format string = "markdown"

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--output" || arg == "-o" {
			if i+1 < len(args) {
				outputFile = args[i+1]
				i++
			}
		} else if arg == "--format" || arg == "-f" {
			if i+1 < len(args) {
				format = args[i+1]
				i++
			}
		} else if !strings.HasPrefix(arg, "--") {
			rangeSpec = arg
		}
	}

	// If no range specified, try to get commits since last tag
	if rangeSpec == "" {
		lastTag, err := git.GetLastTag()
		if err == nil && lastTag != "" {
			rangeSpec = fmt.Sprintf("%s..HEAD", lastTag)
		}
	}

	// Get commit history
	commitLog, err := git.GetCommitLog(rangeSpec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), fmt.Sprintf(i18n.T("error_get_commits"), err))
		os.Exit(1)
	}

	if commitLog == "" {
		fmt.Println(i18n.T("changelog_no_commits"))
		return
	}

	// Use AI to generate CHANGELOG
	fmt.Println(i18n.T("changelog_generating"))
	ctx := context.Background()
	changelog, err := ai.GenerateChangelog(ctx, client, commitLog, format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), fmt.Sprintf(i18n.T("error_generate_changelog"), err))
		os.Exit(1)
	}

	// Output or save to file
	if outputFile != "" {
		err := os.WriteFile(outputFile, []byte(changelog), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), fmt.Sprintf(i18n.T("error_write_file"), outputFile, err))
			os.Exit(1)
		}
		fmt.Printf(i18n.T("changelog_saved")+"\n", outputFile)
	} else {
		fmt.Println(changelog)
	}
}

