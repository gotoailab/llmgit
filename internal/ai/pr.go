package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/gotoailab/llmgit/internal/i18n"
	"github.com/gotoailab/llmgit/internal/utils"
	"github.com/gotoailab/llmhub"
)

// GeneratePRDescription generates a PR description using AI
func GeneratePRDescription(ctx context.Context, client *llmhub.Client, baseBranch, targetBranch, commits, diff string) (string, error) {
	lang := i18n.GetLanguage()
	languageInstruction := "Use English"
	if lang == i18n.LangZH {
		languageInstruction = "Use Chinese"
	}

	prompt := fmt.Sprintf(`Generate a Pull Request description based on the following information.

Requirements:
1. %s
2. Summarize the changes clearly
3. Include impact scope and testing suggestions
4. Use a formatted PR template structure
5. Be concise but comprehensive

Base branch: %s
Target branch: %s

Commits (format: <hash>|<subject>|<author>|<date>):
%s

Code changes (diff stat):
%s

Generate a professional PR description.`, languageInstruction, baseBranch, targetBranch, commits, diff)

	resp, err := client.ChatCompletions(ctx, llmhub.ChatCompletionRequest{
		Messages: []llmhub.ChatMessage{
			{Role: "user", Content: prompt},
		},
		Temperature: utils.FloatPtr(0.7),
		MaxTokens:   utils.IntPtr(1500),
	})
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf(i18n.T("error_no_response"))
	}

	prDesc := strings.TrimSpace(utils.GetContent(resp.Choices[0].Message.Content))
	return prDesc, nil
}

