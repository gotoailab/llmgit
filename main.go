package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
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

	// 显示版本信息（不需要初始化）
	if command == "version" {
		handleVersion()
		return
	}

	// 设置 commit prompt（需要先初始化）
	if command == "set-prompt" {
		handleSetPrompt(args[1:])
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
	fmt.Printf("  %s  - %s\n", i18n.T("version_usage"), i18n.T("version_usage_desc"))
	fmt.Printf("  %s  - %s\n", i18n.T("set_prompt_usage"), i18n.T("set_prompt_usage_desc"))
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

func handleVersion() {
	fmt.Print(GetVersionInfo())
}

func handleSetPrompt(args []string) {
	// 检查是否已初始化
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), i18n.T("error_not_initialized"))
		os.Exit(1)
	}

	if len(args) == 0 {
		// 显示当前 prompt
		if cfg.CommitPrompt != "" {
			fmt.Println(i18n.T("prompt_current"))
			fmt.Println(cfg.CommitPrompt)
		} else {
			fmt.Println(i18n.T("prompt_not_set"))
			fmt.Println(i18n.T("prompt_usage"))
		}
		return
	}

	// 设置 prompt
	// 支持从文件读取或直接输入
	prompt := strings.Join(args, " ")

	// 如果参数是文件路径，读取文件内容
	if len(args) == 1 {
		if _, err := os.Stat(args[0]); err == nil {
			// 文件存在，读取文件内容
			data, err := os.ReadFile(args[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), fmt.Sprintf(i18n.T("error_read_file"), args[0], err))
				os.Exit(1)
			}
			prompt = strings.TrimSpace(string(data))
		}
	}

	cfg.CommitPrompt = prompt
	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), fmt.Sprintf(i18n.T("error_save_prompt"), err))
		os.Exit(1)
	}

	fmt.Println(i18n.T("prompt_saved"))
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

	// 分析变更的文件类型，用于类型建议
	files, err := git.GetStagedFiles()
	if err != nil {
		// 如果获取文件列表失败，继续执行但不提供类型建议
		files = []string{}
	}
	typeHint := analyzeChangeType(files)

	// 使用 AI 生成 commit message
	fmt.Println(i18n.T("commit_generating"))
	ctx := context.Background()
	message, err := generateCommitMessage(ctx, client, diff, lang, cfg, typeHint)
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

// analyzeChangeType 分析变更的文件类型，返回类型建议
func analyzeChangeType(files []string) string {
	if len(files) == 0 {
		return ""
	}

	// 文档文件扩展名
	docExtensions := []string{".md", ".markdown", ".txt", ".rst", ".adoc", ".asciidoc", ".org"}
	// 测试文件扩展名
	testExtensions := []string{"_test.go", "_test.py", ".test.js", ".spec.js", ".test.ts", ".spec.ts", ".test.tsx", ".spec.tsx"}
	// 配置文件扩展名
	configExtensions := []string{".json", ".yaml", ".yml", ".toml", ".ini", ".conf", ".config", ".env"}
	// 构建文件扩展名
	buildExtensions := []string{"Makefile", "makefile", "Dockerfile", "docker-compose.yml", ".mk", "CMakeLists.txt", "build.gradle", "pom.xml", "package.json", "go.mod", "requirements.txt", "package-lock.json", "yarn.lock"}
	// CI 文件路径
	ciPaths := []string{".github/", ".gitlab-ci.yml", ".travis.yml", ".circleci/", "Jenkinsfile", ".jenkins/"}

	allDocs := true
	allTests := true
	allConfig := true
	allBuild := true
	hasCI := false

	for _, file := range files {
		fileLower := strings.ToLower(file)

		// 检查是否是文档文件
		isDoc := false
		for _, ext := range docExtensions {
			if strings.HasSuffix(fileLower, ext) {
				isDoc = true
				break
			}
		}
		if !isDoc {
			allDocs = false
		}

		// 检查是否是测试文件
		isTest := false
		for _, ext := range testExtensions {
			if strings.Contains(fileLower, ext) {
				isTest = true
				break
			}
		}
		if !isTest {
			allTests = false
		}

		// 检查是否是配置文件
		isConfig := false
		for _, ext := range configExtensions {
			if strings.HasSuffix(fileLower, ext) {
				isConfig = true
				break
			}
		}
		if !isConfig {
			allConfig = false
		}

		// 检查是否是构建文件
		isBuild := false
		for _, ext := range buildExtensions {
			if strings.Contains(fileLower, ext) || strings.HasSuffix(fileLower, ext) {
				isBuild = true
				break
			}
		}
		if !isBuild {
			allBuild = false
		}

		// 检查是否是 CI 文件
		for _, ciPath := range ciPaths {
			if strings.Contains(fileLower, ciPath) {
				hasCI = true
				break
			}
		}
	}

	// 根据文件类型返回建议
	if allDocs {
		return "docs"
	}
	if allTests {
		return "test"
	}
	if hasCI {
		return "ci"
	}
	if allBuild {
		return "build"
	}
	if allConfig {
		return "chore"
	}

	return ""
}

func generateCommitMessage(ctx context.Context, client *llmhub.Client, diff string, lang string, cfg *config.Config, typeHint string) (string, error) {
	// 确定语言和提示词
	languageInstruction := "Use English"
	if lang == "zh" || lang == "cn" || lang == "chinese" {
		languageInstruction = "Use Chinese"
	}

	var prompt string
	if cfg.CommitPrompt != "" {
		// 使用自定义 prompt 模板
		// 支持占位符: {language}, {diff}
		prompt = cfg.CommitPrompt
		prompt = strings.ReplaceAll(prompt, "{language}", languageInstruction)
		prompt = strings.ReplaceAll(prompt, "{diff}", diff)
	} else {
		// 使用默认 prompt
		typeGuidance := ""
		if typeHint != "" {
			typeGuidance = fmt.Sprintf("\nIMPORTANT TYPE GUIDANCE: Based on the changed files, this change appears to be primarily %s-related. Please use type '%s' unless the changes clearly indicate otherwise (e.g., if markdown files contain code examples that are part of a feature implementation, use 'feat' instead).", typeHint, typeHint)
		}

		prompt = fmt.Sprintf(`Generate a SINGLE Git commit message based on the code changes below.

Requirements:
1. %s
2. Follow conventional commit format: <type>(<scope>): <short summary>
3. Type must be one of: feat, chore, docs, fix, perf, test, build, ci, revert, refactor, style
4. Scope is optional (can be omitted)
5. Short summary should be concise (no more than 50 characters)
6. If needed, add a detailed description after a blank line
7. At the end, add a note: "generated by llmgit"%s

Code changes:
`+"```"+`
%s
`+"```"+`

CRITICAL REQUIREMENTS:
- Return ONLY ONE commit message
- Do NOT provide multiple options or alternatives
- Do NOT include any introductory text, explanations, or phrases like "Based on", "Here's", "For all changes", "Split into", "Alternative", etc.
- Do NOT number your response (no "1.", "2.", "3.", etc.)
- Start DIRECTLY with the commit message (e.g., "feat:", "fix:", etc.)
- Return ONLY the commit message, nothing else`, languageInstruction, typeGuidance, diff)
	}

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

	// 清理可能的前言部分
	message = cleanCommitMessage(message)

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

// cleanCommitMessage removes introductory text and explanations from the commit message
func cleanCommitMessage(message string) string {
	lines := strings.Split(message, "\n")

	// 常见的前言模式
	introPatterns := []string{
		"based on",
		"here's",
		"here is",
		"here are",
		"the following",
		"below is",
		"below are",
		"this commit",
		"commit message",
		"summarizing",
		"summarizes",
		"for all changes",
		"split into",
		"alternative",
		"the most",
		"would be",
		"option",
		"comprehensive",
	}

	// 移除代码块标记（保留代码块内的内容）
	filteredLines := []string{}
	for _, line := range lines {
		lineTrimmed := strings.TrimSpace(line)
		if strings.HasPrefix(lineTrimmed, "```") {
			continue // 跳过代码块标记行
		}
		filteredLines = append(filteredLines, line)
	}
	lines = filteredLines

	// 查找第一个有效的 commit message 行（通常是 conventional commit 格式）
	startIdx := 0
	for i, line := range lines {
		lineTrimmed := strings.TrimSpace(line)
		lineLower := strings.ToLower(lineTrimmed)

		// 跳过空行
		if lineLower == "" {
			continue
		}

		// 跳过数字开头的行（如 "1.", "2.", "3." 等选项）
		if matched, _ := regexp.MatchString(`^\d+\.`, lineTrimmed); matched {
			continue
		}

		// 检查是否是前言
		isIntro := false
		for _, pattern := range introPatterns {
			if strings.Contains(lineLower, pattern) {
				isIntro = true
				break
			}
		}

		// 如果看起来像 conventional commit 格式（feat:, fix: 等），从这里开始
		if !isIntro && (strings.HasPrefix(lineLower, "feat:") ||
			strings.HasPrefix(lineLower, "fix:") ||
			strings.HasPrefix(lineLower, "docs:") ||
			strings.HasPrefix(lineLower, "style:") ||
			strings.HasPrefix(lineLower, "refactor:") ||
			strings.HasPrefix(lineLower, "perf:") ||
			strings.HasPrefix(lineLower, "test:") ||
			strings.HasPrefix(lineLower, "build:") ||
			strings.HasPrefix(lineLower, "ci:") ||
			strings.HasPrefix(lineLower, "chore:") ||
			strings.HasPrefix(lineLower, "revert:")) {
			startIdx = i
			break
		}
	}

	// 如果没找到 conventional commit 格式，查找第一个非前言行
	if startIdx == 0 {
		for i, line := range lines {
			lineTrimmed := strings.TrimSpace(line)
			lineLower := strings.ToLower(lineTrimmed)
			if lineLower == "" {
				continue
			}
			// 跳过数字开头的行
			if matched, _ := regexp.MatchString(`^\d+\.`, lineTrimmed); matched {
				continue
			}
			// 跳过前言
			isIntro := false
			for _, pattern := range introPatterns {
				if strings.Contains(lineLower, pattern) {
					isIntro = true
					break
				}
			}
			if !isIntro {
				startIdx = i
				break
			}
		}
	}

	// 提取从 startIdx 开始的内容
	if startIdx > 0 {
		lines = lines[startIdx:]
	}

	// 进一步清理：移除选项说明和解释性文字
	finalLines := []string{}
	for _, line := range lines {
		lineTrimmed := strings.TrimSpace(line)
		lineLower := strings.ToLower(lineTrimmed)

		// 跳过空行（但保留第一个空行作为分隔符）
		if lineTrimmed == "" {
			if len(finalLines) > 0 && finalLines[len(finalLines)-1] != "" {
				finalLines = append(finalLines, "")
			}
			continue
		}

		// 跳过数字开头的选项行
		if matched, _ := regexp.MatchString(`^\d+\.`, lineTrimmed); matched {
			continue
		}

		// 跳过明显的解释性文字
		isExplanation := false
		explanationPatterns := []string{
			"the most",
			"would be",
			"option",
			"comprehensive",
			"as it",
			"while maintaining",
			"the changes include",
		}
		for _, pattern := range explanationPatterns {
			if strings.Contains(lineLower, pattern) {
				isExplanation = true
				break
			}
		}

		// 如果遇到解释性文字，停止处理（说明已经到 commit message 的末尾了）
		if isExplanation {
			break
		}

		finalLines = append(finalLines, line)
	}

	// 移除末尾的代码块标记
	result := strings.Join(finalLines, "\n")
	result = strings.TrimSpace(result)
	if strings.HasSuffix(result, "```") {
		result = strings.TrimSuffix(result, "```")
		result = strings.TrimSpace(result)
	}

	return result
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
