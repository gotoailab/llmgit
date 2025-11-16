package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/gotoailab/llmgit/internal/i18n"
	"github.com/gotoailab/llmgit/internal/utils"
	"github.com/gotoailab/llmhub"
)

// GenerateChangelog generates a CHANGELOG using AI
func GenerateChangelog(ctx context.Context, client *llmhub.Client, commitLog string, format string) (string, error) {
	lang := i18n.GetLanguage()
	languageInstruction := "Use English"
	if lang == i18n.LangZH {
		languageInstruction = "Use Chinese"
	}

	formatInstruction := "Markdown format"
	if format == "json" {
		formatInstruction = "JSON format"
	} else if format == "yaml" {
		formatInstruction = "YAML format"
	}

	prompt := fmt.Sprintf(`Generate a CHANGELOG based on the following commit history. The commits are in the format: <commit-hash>|<subject>|<author>|<date>

Requirements:
1. %s
2. %s
3. Organize commits by type (feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert)
4. Group by version if version tags are present, otherwise group by date
5. Include commit subject and author for each entry
6. Use clear and concise descriptions

Commit history:
%s

Generate a well-formatted CHANGELOG.`, languageInstruction, formatInstruction, commitLog)

	resp, err := client.ChatCompletions(ctx, llmhub.ChatCompletionRequest{
		Messages: []llmhub.ChatMessage{
			{Role: "user", Content: prompt},
		},
		Temperature: utils.FloatPtr(0.7),
		MaxTokens:   utils.IntPtr(2000),
	})
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf(i18n.T("error_no_response"))
	}

	changelog := strings.TrimSpace(utils.GetContent(resp.Choices[0].Message.Content))
	return changelog, nil
}

