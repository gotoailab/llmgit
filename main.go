package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/gotoailab/llmgit/internal/config"
	"github.com/gotoailab/llmgit/internal/git"
	"github.com/gotoailab/llmhub"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// 初始化命令
	if command == "init" {
		handleInit()
		return
	}

	// 检查是否已初始化
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 项目未初始化，请先运行 'llmgit init'\n")
		os.Exit(1)
	}

	// 创建 LLM 客户端
	client, err := llmhub.NewClient(llmhub.ClientConfig{
		APIKey:   cfg.APIKey,
		Provider: llmhub.Provider(cfg.Provider),
		Model:    cfg.Model,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法创建 LLM 客户端: %v\n", err)
		os.Exit(1)
	}

	// 处理特殊命令
	switch command {
	case "commit":
		handleCommit(client, cfg, os.Args[2:])
	case "review":
		handleReview(client, cfg, os.Args[2:])
	case "diff":
		handleDiff(client, cfg, os.Args[2:])
	case "explain":
		handleExplain(client, cfg, os.Args[2:])
	default:
		// 其他命令直接转发给 git
		handleGitCommand(os.Args[1:])
	}
}

func printUsage() {
	fmt.Println("llmgit - AI 增强的 Git 工具")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  llmgit init <provider> <api-key> [model]  - 初始化项目配置")
	fmt.Println("  llmgit commit [options]                   - AI 生成 commit message 并提交")
	fmt.Println("    --lang, -l <lang>                       - 指定语言 (en/zh, 默认: en)")
	fmt.Println("  llmgit review [commit]                    - AI 审查 commit")
	fmt.Println("  llmgit diff [options]                     - 显示 diff 并 AI 解释")
	fmt.Println("  llmgit explain [file]                    - AI 语义化解释文件变更")
	fmt.Println("  llmgit <git-command>                      - 执行其他 git 命令")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  llmgit init openai sk-xxx gpt-4")
	fmt.Println("  llmgit commit -a")
	fmt.Println("  llmgit commit --lang zh -a                # 使用中文生成 commit message")
	fmt.Println("  llmgit review HEAD")
	fmt.Println("  llmgit diff")
}

func handleInit() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "用法: llmgit init <provider> <api-key> [model]\n")
		fmt.Fprintf(os.Stderr, "\n支持的 provider:\n")
		for _, p := range llmhub.AllProviders() {
			fmt.Fprintf(os.Stderr, "  - %s\n", p)
		}
		os.Exit(1)
	}

	provider := os.Args[2]
	apiKey := os.Args[3]
	model := ""
	if len(os.Args) > 4 {
		model = os.Args[4]
	}

	// 验证 provider
	if !llmhub.Provider(provider).IsValid() {
		fmt.Fprintf(os.Stderr, "错误: 不支持的 provider: %s\n", provider)
		os.Exit(1)
	}

	cfg := &config.Config{
		Provider: provider,
		APIKey:   apiKey,
		Model:    model,
	}

	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法保存配置: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ 配置已保存到 ~/.llmgit/config.json\n")
	fmt.Printf("  Provider: %s\n", provider)
	if model != "" {
		fmt.Printf("  Model: %s\n", model)
	} else {
		fmt.Printf("  Model: (使用默认模型)\n")
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
		fmt.Fprintf(os.Stderr, "错误: 无法获取变更: %v\n", err)
		os.Exit(1)
	}

	if diff == "" {
		fmt.Println("没有暂存的变更")
		os.Exit(0)
	}

	// 使用 AI 生成 commit message
	fmt.Println("正在使用 AI 生成 commit message...")
	ctx := context.Background()
	message, err := generateCommitMessage(ctx, client, diff, lang)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法生成 commit message: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n生成的 commit message:\n\n%s\n\n", message)

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
		fmt.Print("是否使用此 commit message? (y/n): ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "y" && confirm != "Y" {
			fmt.Println("已取消")
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
	fmt.Printf("正在分析 commit: %s...\n", commit)
	commitInfo, err := git.GetCommitInfo(commit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法获取 commit 信息: %v\n", err)
		os.Exit(1)
	}

	// 使用 AI 审查
	ctx := context.Background()
	review, err := reviewCommit(ctx, client, commitInfo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法审查 commit: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n=== AI 审查结果 ===\n\n%s\n", review)
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
		fmt.Println("\n--- AI 解释 ---")
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
		fmt.Printf("正在分析文件: %s...\n", args[0])
		diff, err = git.GetFileDiff(args[0])
	} else {
		// 解释工作区的变更
		fmt.Println("正在分析工作区变更...")
		diff, err = git.GetDiff()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法获取变更: %v\n", err)
		os.Exit(1)
	}

	if diff == "" {
		fmt.Println("没有变更")
		os.Exit(0)
	}

	ctx := context.Background()
	explanation, err := explainDiff(ctx, client, diff)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法生成解释: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n=== AI 解释 ===\n\n%s\n", explanation)
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
		languageInstruction = "使用中文"
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
		return "", fmt.Errorf("没有收到响应")
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
	prompt := fmt.Sprintf(`你是一个专业的代码审查助手。请审查以下 Git commit，提供详细的审查意见。

要求：
1. 使用中文
2. 评估代码质量、潜在问题、最佳实践
3. 提供建设性的建议
4. 如果发现问题，请明确指出

Commit 信息：
%s

请提供详细的审查报告。`, commitInfo)

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
		return "", fmt.Errorf("没有收到响应")
	}

	return getContent(resp.Choices[0].Message.Content), nil
}

func explainDiff(ctx context.Context, client *llmhub.Client, diff string) (string, error) {
	prompt := fmt.Sprintf(`你是一个专业的代码解释助手。请用通俗易懂的语言解释以下代码变更的含义。

要求：
1. 使用中文
2. 解释变更的目的和影响
3. 指出关键的变化点
4. 如果可能，说明为什么这样修改

代码变更：
%s

请提供清晰的解释。`, diff)

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
		return "", fmt.Errorf("没有收到响应")
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
