package ai

import (
	"context"
	"fmt"

	"github.com/gotoailab/llmgit/internal/i18n"
	"github.com/gotoailab/llmgit/internal/utils"
	"github.com/gotoailab/llmhub"
)

// ReviewCommit reviews a commit using AI
func ReviewCommit(ctx context.Context, client *llmhub.Client, commitInfo string) (string, error) {
	lang := i18n.GetLanguage()
	languageInstruction := "Use English"
	if lang == i18n.LangZH {
		languageInstruction = "Use Chinese"
	}

	prompt := fmt.Sprintf(`You are a professional code review assistant. Please review the following Git commit and provide detailed review comments.

Requirements:
1. %s
2. Evaluate code quality, potential issues, and best practices
3. Provide constructive suggestions
4. If issues are found, clearly point them out

Commit information:
%s

Please provide a detailed review report.`, languageInstruction, commitInfo)

	resp, err := client.ChatCompletions(ctx, llmhub.ChatCompletionRequest{
		Messages: []llmhub.ChatMessage{
			{Role: "user", Content: prompt},
		},
		Temperature: utils.FloatPtr(0.7),
		MaxTokens:   utils.IntPtr(1000),
	})
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf(i18n.T("error_no_response"))
	}

	return utils.GetContent(resp.Choices[0].Message.Content), nil
}

