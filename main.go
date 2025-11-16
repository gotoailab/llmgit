package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/gotoailab/llmgit/internal/config"
	"github.com/gotoailab/llmgit/internal/git"
	"github.com/gotoailab/llmgit/internal/i18n"
	"github.com/gotoailab/llmhub"
)

func main() {
	// Initialize i18n (default to English)
	i18n.InitLanguage()

	// Parse global language option if present
	args := parseGlobalLangOption(os.Args[1:])

	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	// Check if it's an AI command (starts with "ai")
	if args[0] == "ai" {
		if len(args) < 2 {
			printUsage()
			os.Exit(1)
		}
		handleAICommand(args[1:])
		return
	}

	// Check for help command (without ai prefix for convenience)
	if args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		printUsage()
		return
	}

	// All other commands are forwarded to git
	handleGitCommand(args)
}

func handleAICommand(args []string) {
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	command := args[0]

	// 初始化命令
	if command == "init" {
		handleInit(args[1:])
		return
	}

	// 帮助命令（不需要初始化）
	if command == "help" || command == "--help" || command == "-h" {
		printUsage()
		return
	}

	// 列出所有支持的 provider（不需要初始化）
	if command == "providers" || command == "list-providers" {
		handleProviders()
		return
	}

	// 检查是否已初始化
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), i18n.T("error_not_initialized"))
		os.Exit(1)
	}

	// 创建 LLM 客户端
	client, err := llmhub.NewClient(llmhub.ClientConfig{
		APIKey:   cfg.APIKey,
		Provider: llmhub.Provider(cfg.Provider),
		Model:    cfg.Model,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), fmt.Sprintf(i18n.T("error_create_client"), err))
		os.Exit(1)
	}

	// 处理特殊命令
	switch command {
	case "commit":
		handleCommit(client, cfg, args[1:])
	case "review":
		handleReview(client, cfg, args[1:])
	case "diff":
		handleDiff(client, cfg, args[1:])
	case "explain":
		handleExplain(client, cfg, args[1:])
	default:
		fmt.Fprintf(os.Stderr, "%s%s: %s\n", i18n.T("error_prefix"), i18n.T("error_unknown_command"), command)
		fmt.Fprintf(os.Stderr, "%s\n", i18n.T("error_run_help"))
		os.Exit(1)
	}
}

// parseGlobalLangOption parses --lang global option and returns remaining args
func parseGlobalLangOption(args []string) []string {
	result := []string{}
	skipNext := false
	for i, arg := range args {
		if skipNext {
			skipNext = false
			continue
		}
		if arg == "--lang" || arg == "-l" {
			if i+1 < len(args) {
				i18n.SetLanguage(args[i+1])
				skipNext = true
				continue
			}
		} else if strings.HasPrefix(arg, "--lang=") {
			lang := strings.TrimPrefix(arg, "--lang=")
			i18n.SetLanguage(lang)
		} else {
			result = append(result, arg)
		}
	}
	return result
}

func printUsage() {
	fmt.Println(i18n.T("app_name"))
	fmt.Println()
	fmt.Printf("%s:\n", i18n.T("usage"))
	fmt.Printf("  llmgit help  - %s\n", i18n.T("help_usage_desc"))
	fmt.Printf("  %s  - %s\n", i18n.T("init_usage"), i18n.T("init_usage_desc"))
	fmt.Printf("  %s  - %s\n", i18n.T("commit_usage"), i18n.T("commit_usage_desc"))
	fmt.Printf("    %s  - %s\n", i18n.T("commit_lang_option"), i18n.T("commit_lang_desc"))
	fmt.Printf("  %s  - %s\n", i18n.T("review_usage"), i18n.T("review_usage_desc"))
	fmt.Printf("  %s  - %s\n", i18n.T("diff_usage"), i18n.T("diff_usage_desc"))
	fmt.Printf("  %s  - %s\n", i18n.T("explain_usage"), i18n.T("explain_usage_desc"))
	fmt.Printf("  %s  - %s\n", i18n.T("providers_usage"), i18n.T("providers_usage_desc"))
	fmt.Printf("  llmgit <git-command>  - %s\n", i18n.T("git_command_desc"))
	fmt.Println()
	fmt.Printf("%s:\n", i18n.T("examples"))
	fmt.Println("  llmgit ai init openai sk-xxx gpt-4")
	fmt.Println("  llmgit ai commit -a")
	if i18n.GetLanguage() == i18n.LangZH {
		fmt.Println("  llmgit ai commit --lang zh -a                # 使用中文生成 commit message")
	} else {
		fmt.Println("  llmgit ai commit --lang zh -a                # Generate commit message in Chinese")
	}
	fmt.Println("  llmgit ai review HEAD")
	fmt.Println("  llmgit ai diff")
}

func handleInit(args []string) {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "%s: %s\n", i18n.T("usage"), i18n.T("init_usage"))
		fmt.Fprintf(os.Stderr, "\n%s\n", i18n.T("init_supported_providers"))
		for _, p := range llmhub.AllProviders() {
			fmt.Fprintf(os.Stderr, "  - %s\n", p)
		}
		os.Exit(1)
	}

	provider := args[0]
	apiKey := args[1]
	model := ""
	if len(args) > 2 {
		model = args[2]
	}

	// 验证 provider
	if !llmhub.Provider(provider).IsValid() {
		fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), fmt.Sprintf(i18n.T("init_error_invalid_provider"), provider))
		os.Exit(1)
	}

	cfg := &config.Config{
		Provider: provider,
		APIKey:   apiKey,
		Model:    model,
	}

	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), fmt.Sprintf(i18n.T("init_error_save_failed"), err))
		os.Exit(1)
	}

	fmt.Println(i18n.T("init_success"))
	fmt.Printf(i18n.T("init_provider")+"\n", provider)
	if model != "" {
		fmt.Printf(i18n.T("init_model")+"\n", model)
	} else {
		fmt.Println(i18n.T("init_model_default"))
	}
}

func handleProviders() {
	fmt.Println(i18n.T("providers_title"))
	for _, p := range llmhub.AllProviders() {
		fmt.Println(p)
	}
}

func handleCommit(client *llmhub.Client, cfg *config.Config, args []string) {
	// 检查是否已经有 -m 参数（用户自己提供了 message）
	hasMessage := false
	for i, arg := range args {
		if arg == "-m" && i+1 < len(args) {
			hasMessage = true
			break
		}
	}

	// 如果用户已经提供了 message，直接执行 git commit
	if hasMessage {
		gitArgs := []string{"commit"}
		gitArgs = append(gitArgs, args...)
		handleGitCommand(gitArgs)
		return
	}

	// 解析语言选项（--lang 或 -l）
	lang := "en" // 默认英文
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

	// 获取暂存区的变更
	diff, err := git.GetStagedDiff()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), fmt.Sprintf(i18n.T("error_get_changes"), err))
		os.Exit(1)
	}

	if diff == "" {
		fmt.Println(i18n.T("commit_no_staged"))
		os.Exit(0)
	}

	// 使用 AI 生成 commit message
	fmt.Println(i18n.T("commit_generating"))
	ctx := context.Background()
	message, err := generateCommitMessage(ctx, client, diff, lang)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), fmt.Sprintf(i18n.T("error_generate_message"), err))
		os.Exit(1)
	}

	fmt.Printf("\n%s\n\n%s\n\n", i18n.T("commit_generated"), message)

	// 检查是否有 --no-edit 参数（跳过确认）
	skipConfirm := false
	for _, arg := range filteredArgs {
		if arg == "--no-edit" {
			skipConfirm = true
			break
		}
	}

	// 询问是否确认
	if !skipConfirm {
		fmt.Print(i18n.T("commit_confirm"))
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "y" && confirm != "Y" {
			fmt.Println(i18n.T("commit_cancelled"))
			os.Exit(0)
		}
	}

	// 执行 git commit
	gitArgs := []string{"commit", "-m", message}
	// 过滤掉 --no-edit 和 --lang 相关参数
	finalArgs := []string{}
	for _, arg := range filteredArgs {
		if arg != "--no-edit" {
			finalArgs = append(finalArgs, arg)
		}
	}
	gitArgs = append(gitArgs, finalArgs...)
	handleGitCommand(gitArgs)
}

func handleReview(client *llmhub.Client, cfg *config.Config, args []string) {
	commit := "HEAD"
	if len(args) > 0 {
		commit = args[0]
	}

	// 获取 commit 信息
	fmt.Printf(i18n.T("review_analyzing")+"\n", commit)
	commitInfo, err := git.GetCommitInfo(commit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), fmt.Sprintf(i18n.T("error_get_commit_info"), err))
		os.Exit(1)
	}

	// 使用 AI 审查
	ctx := context.Background()
	review, err := reviewCommit(ctx, client, commitInfo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), fmt.Sprintf(i18n.T("error_review_commit"), err))
		os.Exit(1)
	}

	fmt.Printf("\n%s\n\n%s\n", i18n.T("review_result"), review)
}

func handleDiff(client *llmhub.Client, cfg *config.Config, args []string) {
	// 先执行普通的 git diff
	gitArgs := []string{"diff"}
	gitArgs = append(gitArgs, args...)

	cmd := exec.Command("git", gitArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}

	// 获取 diff 内容用于 AI 解释
	diff, err := git.GetDiff(args...)
	if err != nil {
		return // 忽略错误，只显示普通 diff
	}

	if diff != "" {
		fmt.Println("\n" + i18n.T("diff_explanation"))
		ctx := context.Background()
		explanation, err := explainDiff(ctx, client, diff)
		if err == nil {
			fmt.Println(explanation)
		}
	}
}

func handleExplain(client *llmhub.Client, cfg *config.Config, args []string) {
	var diff string
	var err error

	if len(args) > 0 {
		// 解释特定文件的变更
		fmt.Printf(i18n.T("explain_analyzing_file")+"\n", args[0])
		diff, err = git.GetFileDiff(args[0])
	} else {
		// 解释工作区的变更
		fmt.Println(i18n.T("explain_analyzing_workdir"))
		diff, err = git.GetDiff()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), fmt.Sprintf(i18n.T("error_get_changes"), err))
		os.Exit(1)
	}

	if diff == "" {
		fmt.Println(i18n.T("explain_no_changes"))
		os.Exit(0)
	}

	ctx := context.Background()
	explanation, err := explainDiff(ctx, client, diff)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), fmt.Sprintf(i18n.T("error_generate_explanation"), err))
		os.Exit(1)
	}

	fmt.Printf("\n%s\n\n%s\n", i18n.T("explain_result"), explanation)
}

func handleGitCommand(args []string) {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}

func generateCommitMessage(ctx context.Context, client *llmhub.Client, diff string, lang string) (string, error) {
	// 确定语言和提示词
	languageInstruction := "Use English"
	if lang == "zh" || lang == "cn" || lang == "chinese" {
		languageInstruction = "Use Chinese"
	}

	prompt := fmt.Sprintf(`You are a professional Git commit message generator. Please generate a concise and clear commit message based on the following code changes.

Requirements:
1. %s
2. Follow the conventional commit format: <type>(<scope>): <short summary>
3. Type must be one of: feat, chore, docs, fix, perf, test, build, ci, revert, refactor, style
4. Scope is optional (can be omitted)
5. Short summary should be concise (no more than 50 characters)
6. If needed, add a detailed description after a blank line
7. At the end, add a note: "generated by llmgit"

Code changes:
%s

Please return only the commit message, without any additional explanation.`, languageInstruction, diff)

	resp, err := client.ChatCompletions(ctx, llmhub.ChatCompletionRequest{
		Messages: []llmhub.ChatMessage{
			{Role: "user", Content: prompt},
		},
		Temperature: floatPtr(0.7),
		MaxTokens:   intPtr(500),
	})
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf(i18n.T("error_no_response"))
	}

	message := strings.TrimSpace(getContent(resp.Choices[0].Message.Content))

	// 确保最后有 "generated by llmgit" 备注
	// 移除可能已经存在的 "generated by llmgit"（避免重复）
	message = strings.TrimSpace(message)
	lowerMessage := strings.ToLower(message)
	if strings.Contains(lowerMessage, "generated by llmgit") {
		// 如果已经存在，移除它（包括前后的空行）
		lines := strings.Split(message, "\n")
		filteredLines := []string{}
		found := false
		for _, line := range lines {
			if strings.Contains(strings.ToLower(line), "generated by llmgit") {
				found = true
				continue
			}
			// 如果找到了，跳过后续的空行
			if found && strings.TrimSpace(line) == "" {
				continue
			}
			filteredLines = append(filteredLines, line)
		}
		message = strings.Join(filteredLines, "\n")
		message = strings.TrimSpace(message)
	}

	// 在最后添加 "generated by llmgit"
	if message != "" {
		message = message + "\n\ngenerated by llmgit"
	} else {
		message = "generated by llmgit"
	}

	return message, nil
}

func reviewCommit(ctx context.Context, client *llmhub.Client, commitInfo string) (string, error) {
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
		Temperature: floatPtr(0.7),
		MaxTokens:   intPtr(1000),
	})
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf(i18n.T("error_no_response"))
	}

	return getContent(resp.Choices[0].Message.Content), nil
}

func explainDiff(ctx context.Context, client *llmhub.Client, diff string) (string, error) {
	lang := i18n.GetLanguage()
	languageInstruction := "Use English"
	if lang == i18n.LangZH {
		languageInstruction = "Use Chinese"
	}

	prompt := fmt.Sprintf(`You are a professional code explanation assistant. Please explain the meaning of the following code changes in plain language.

Requirements:
1. %s
2. Explain the purpose and impact of the changes
3. Point out key changes
4. If possible, explain why these changes were made

Code changes:
%s

Please provide a clear explanation.`, languageInstruction, diff)

	resp, err := client.ChatCompletions(ctx, llmhub.ChatCompletionRequest{
		Messages: []llmhub.ChatMessage{
			{Role: "user", Content: prompt},
		},
		Temperature: floatPtr(0.7),
		MaxTokens:   intPtr(1000),
	})
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf(i18n.T("error_no_response"))
	}

	return getContent(resp.Choices[0].Message.Content), nil
}

func getContent(content interface{}) string {
	switch v := content.(type) {
	case string:
		return v
	case []interface{}:
		var parts []string
		for _, part := range v {
			if m, ok := part.(map[string]interface{}); ok {
				if text, ok := m["text"].(string); ok {
					parts = append(parts, text)
				}
			}
		}
		return strings.Join(parts, "\n")
	default:
		return fmt.Sprintf("%v", v)
	}
}

func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}
