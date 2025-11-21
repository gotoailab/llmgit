package ai

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/gotoailab/llmgit/internal/config"
	"github.com/gotoailab/llmgit/internal/i18n"
	"github.com/gotoailab/llmgit/internal/utils"
	"github.com/gotoailab/llmhub"
)

const (
	// DefaultMaxFiles 默认最大文件数量
	DefaultMaxFiles = 50
	// DefaultMaxLines 默认最大代码行数
	DefaultMaxLines = 5000
	// MaxRetries 最大重试次数
	MaxRetries = 2
)

// ValidCommitTypes 有效的 commit 类型
var ValidCommitTypes = []string{
	"feat", "chore", "docs", "fix", "perf", "test",
	"build", "ci", "revert", "refactor", "style",
}

// AnalyzeChangeType analyzes changed file types and returns type suggestion
func AnalyzeChangeType(files []string) string {
	if len(files) == 0 {
		return ""
	}

	// Document file extensions
	docExtensions := []string{".md", ".markdown", ".txt", ".rst", ".adoc", ".asciidoc", ".org"}
	// Test file extensions
	testExtensions := []string{"_test.go", "_test.py", ".test.js", ".spec.js", ".test.ts", ".spec.ts", ".test.tsx", ".spec.tsx"}
	// Config file extensions
	configExtensions := []string{".json", ".yaml", ".yml", ".toml", ".ini", ".conf", ".config", ".env"}
	// Build file extensions
	buildExtensions := []string{"Makefile", "makefile", "Dockerfile", "docker-compose.yml", ".mk", "CMakeLists.txt", "build.gradle", "pom.xml", "package.json", "go.mod", "requirements.txt", "package-lock.json", "yarn.lock"}
	// CI file paths
	ciPaths := []string{".github/", ".gitlab-ci.yml", ".travis.yml", ".circleci/", "Jenkinsfile", ".jenkins/"}

	allDocs := true
	allTests := true
	allConfig := true
	allBuild := true
	hasCI := false

	for _, file := range files {
		fileLower := strings.ToLower(file)

		// Check if it's a document file
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

		// Check if it's a test file
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

		// Check if it's a config file
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

		// Check if it's a build file
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

		// Check if it's a CI file
		for _, ciPath := range ciPaths {
			if strings.Contains(fileLower, ciPath) {
				hasCI = true
				break
			}
		}
	}

	// Return suggestion based on file types
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

// LimitDiffSize 限制 diff 的大小，防止传输过多上下文
// maxFiles: 最大文件数量，0 表示使用默认值
// maxLines: 最大代码行数，0 表示使用默认值
func LimitDiffSize(diff string, maxFiles, maxLines int) string {
	if diff == "" {
		return ""
	}

	// 使用默认值
	if maxFiles <= 0 {
		maxFiles = DefaultMaxFiles
	}
	if maxLines <= 0 {
		maxLines = DefaultMaxLines
	}

	lines := strings.Split(diff, "\n")
	var resultLines []string
	var fileCount int
	var codeLineCount int

	for i, line := range lines {
		// 检测文件头
		if strings.HasPrefix(line, "diff --git") {
			// 如果已经达到文件数量限制，停止处理
			if fileCount >= maxFiles {
				// 添加截断提示
				if len(resultLines) > 0 {
					resultLines = append(resultLines, fmt.Sprintf("\n... (truncated: reached file limit of %d files)", maxFiles))
				}
				break
			}

			// 提取文件路径
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				fileCount++
			}
			resultLines = append(resultLines, line)
		} else if strings.HasPrefix(line, "--- a/") || strings.HasPrefix(line, "+++ b/") {
			resultLines = append(resultLines, line)
		} else if strings.HasPrefix(line, "@@") {
			// 这是 hunk 头，包含行数信息
			resultLines = append(resultLines, line)
		} else if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
			// 这是代码行
			codeLineCount++
			if codeLineCount > maxLines {
				// 达到行数限制，添加截断提示并停止
				if len(resultLines) > 0 {
					resultLines = append(resultLines, fmt.Sprintf("\n... (truncated: reached line limit of %d lines)", maxLines))
				}
				break
			}
			resultLines = append(resultLines, line)
		} else {
			// 其他行（如索引行、二进制文件提示等）
			// 如果已经达到行数限制，不再添加
			if codeLineCount >= maxLines {
				// 检查是否还有更多文件
				hasMoreFiles := false
				for j := i + 1; j < len(lines); j++ {
					if strings.HasPrefix(lines[j], "diff --git") {
						hasMoreFiles = true
						break
					}
				}
				if hasMoreFiles {
					resultLines = append(resultLines, fmt.Sprintf("\n... (truncated: reached line limit of %d lines)", maxLines))
					break
				}
			}
			resultLines = append(resultLines, line)
		}
	}

	return strings.Join(resultLines, "\n")
}

// CleanCommitMessage removes introductory text and explanations from the commit message
func CleanCommitMessage(message string) string {
	lines := strings.Split(message, "\n")

	// Common intro patterns
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

	// Remove code block markers (keep content inside)
	filteredLines := []string{}
	for _, line := range lines {
		lineTrimmed := strings.TrimSpace(line)
		if strings.HasPrefix(lineTrimmed, "```") {
			continue // Skip code block marker lines
		}
		filteredLines = append(filteredLines, line)
	}
	lines = filteredLines

	// Find first valid commit message line (usually conventional commit format)
	startIdx := 0
	for i, line := range lines {
		lineTrimmed := strings.TrimSpace(line)
		lineLower := strings.ToLower(lineTrimmed)

		// Skip empty lines
		if lineLower == "" {
			continue
		}

		// Skip numbered option lines (like "1.", "2.", "3.")
		if matched, _ := regexp.MatchString(`^\d+\.`, lineTrimmed); matched {
			continue
		}

		// Check if it's an intro
		isIntro := false
		for _, pattern := range introPatterns {
			if strings.Contains(lineLower, pattern) {
				isIntro = true
				break
			}
		}

		// If it looks like conventional commit format (feat:, fix:, etc.), start from here
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

	// If no conventional commit format found, find first non-intro line
	if startIdx == 0 {
		for i, line := range lines {
			lineTrimmed := strings.TrimSpace(line)
			lineLower := strings.ToLower(lineTrimmed)
			if lineLower == "" {
				continue
			}
			// Skip numbered lines
			if matched, _ := regexp.MatchString(`^\d+\.`, lineTrimmed); matched {
				continue
			}
			// Skip intro
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

	// Extract content from startIdx
	if startIdx > 0 {
		lines = lines[startIdx:]
	}

	// Further cleanup: remove option descriptions and explanatory text, ensure only first commit message is kept
	finalLines := []string{}
	commitTypePattern := regexp.MustCompile(`^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\(.+?\))?:`)
	foundFirstCommit := false

	for _, line := range lines {
		lineTrimmed := strings.TrimSpace(line)
		lineLower := strings.ToLower(lineTrimmed)

		// Skip empty lines (but keep first empty line as separator)
		if lineTrimmed == "" {
			if len(finalLines) > 0 && finalLines[len(finalLines)-1] != "" {
				finalLines = append(finalLines, "")
			}
			continue
		}

		// Skip numbered option lines
		if matched, _ := regexp.MatchString(`^\d+\.`, lineTrimmed); matched {
			continue
		}

		// Detect if it's a conventional commit format start
		if commitTypePattern.MatchString(lineTrimmed) {
			if foundFirstCommit {
				break
			}
			foundFirstCommit = true
		}

		// Skip obvious explanatory text
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

		// If explanatory text is encountered, stop processing (reached end of commit message)
		if isExplanation {
			break
		}

		// Only add lines after finding first commit
		if foundFirstCommit {
			finalLines = append(finalLines, line)
		}
	}

	// Remove trailing code block markers
	result := strings.Join(finalLines, "\n")
	result = strings.TrimSpace(result)
	if strings.HasSuffix(result, "```") {
		result = strings.TrimSuffix(result, "```")
		result = strings.TrimSpace(result)
	}

	return result
}

// isValidCommitMessage 检查 commit message 是否有效
// 有效条件：去掉前面的空格后，以 conventional commit 类型开头
func isValidCommitMessage(message string) bool {
	// 去掉 "generated by llmgit" 后检查
	trimmed := strings.TrimSpace(message)
	lowerMessage := strings.ToLower(trimmed)

	// 如果只包含 "generated by llmgit"，视为无效
	if trimmed == "" || lowerMessage == "generated by llmgit" {
		return false
	}

	// 移除 "generated by llmgit" 部分
	if strings.Contains(lowerMessage, "generated by llmgit") {
		lines := strings.Split(trimmed, "\n")
		filteredLines := []string{}
		for _, line := range lines {
			if !strings.Contains(strings.ToLower(line), "generated by llmgit") {
				filteredLines = append(filteredLines, line)
			}
		}
		trimmed = strings.TrimSpace(strings.Join(filteredLines, "\n"))
	}

	// 如果移除后为空，视为无效
	if trimmed == "" {
		return false
	}

	// 去掉前面的空格
	trimmed = strings.TrimLeft(trimmed, " \t\n\r")

	// 检查是否以有效的 commit 类型开头
	for _, commitType := range ValidCommitTypes {
		prefix := commitType + ":"
		if strings.HasPrefix(trimmed, prefix) {
			return true
		}
		// 也检查带 scope 的情况，如 "feat(scope):"
		prefixWithScope := commitType + "("
		if strings.HasPrefix(trimmed, prefixWithScope) {
			return true
		}
	}

	return false
}

// generateCommitMessageOnce 生成一次 commit message（不包含重试逻辑）
func generateCommitMessageOnce(ctx context.Context, client *llmhub.Client, diff string, lang string, cfg *config.Config, typeHint string) (string, error) {
	// Determine language and prompt
	languageInstruction := "Use English"
	if lang == "zh" || lang == "cn" || lang == "chinese" {
		languageInstruction = "Use Chinese"
	}

	var prompt string
	if cfg.CommitPrompt != "" {
		// Use custom prompt template
		// Supports placeholders: {language}, {diff}
		prompt = cfg.CommitPrompt
		prompt = strings.ReplaceAll(prompt, "{language}", languageInstruction)
		prompt = strings.ReplaceAll(prompt, "{diff}", diff)
	} else {
		// Use default prompt
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
- Return EXACTLY ONE commit message, no more, no less
- Do NOT provide multiple commit messages, even if you think there are multiple logical changes
- Do NOT list multiple options or alternatives
- Do NOT include any introductory text, explanations, or phrases like "Based on", "Here's", "For all changes", "Split into", "Alternative", etc.
- Do NOT number your response (no "1.", "2.", "3.", etc.)
- Start DIRECTLY with the commit message (e.g., "feat:", "fix:", etc.)
- If there are multiple changes, combine them into a SINGLE comprehensive commit message
- Return ONLY the commit message itself, nothing else
- Do NOT include multiple commit messages separated by blank lines or any other separator`, languageInstruction, typeGuidance, diff)
	}

	resp, err := client.ChatCompletions(ctx, llmhub.ChatCompletionRequest{
		Messages: []llmhub.ChatMessage{
			{Role: "user", Content: prompt},
		},
		Temperature: utils.FloatPtr(0.7),
		MaxTokens:   utils.IntPtr(500),
	})
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf(i18n.T("error_no_response"))
	}

	message := strings.TrimSpace(utils.GetContent(resp.Choices[0].Message.Content))

	// Clean possible intro parts
	message = CleanCommitMessage(message)

	// Ensure "generated by llmgit" note at the end
	// Remove any existing "generated by llmgit" (avoid duplicates)
	message = strings.TrimSpace(message)
	lowerMessage := strings.ToLower(message)
	if strings.Contains(lowerMessage, "generated by llmgit") {
		// If it exists, remove it (including surrounding blank lines)
		lines := strings.Split(message, "\n")
		filteredLines := []string{}
		found := false
		for _, line := range lines {
			if strings.Contains(strings.ToLower(line), "generated by llmgit") {
				found = true
				continue
			}
			// If found, skip subsequent blank lines
			if found && strings.TrimSpace(line) == "" {
				continue
			}
			filteredLines = append(filteredLines, line)
		}
		message = strings.Join(filteredLines, "\n")
		message = strings.TrimSpace(message)
	}

	// Add "generated by llmgit" at the end
	if message != "" {
		message = message + "\n\ngenerated by llmgit"
	} else {
		message = "generated by llmgit"
	}

	return message, nil
}

// GenerateCommitMessage generates a commit message using AI
// It will retry up to MaxRetries times if the result is invalid
func GenerateCommitMessage(ctx context.Context, client *llmhub.Client, diff string, lang string, cfg *config.Config, typeHint string, debug bool) (string, error) {
	// Limit diff size before processing
	diff = LimitDiffSize(diff, DefaultMaxFiles, DefaultMaxLines)

	var message string
	var err error

	// 尝试生成，最多重试 MaxRetries 次
	for attempt := 0; attempt <= MaxRetries; attempt++ {
		message, err = generateCommitMessageOnce(ctx, client, diff, lang, cfg, typeHint)
		if err != nil {
			// 如果是最后一次尝试，返回错误
			if attempt == MaxRetries {
				fmt.Println("Failed to generate commit message after", MaxRetries, "attempts, error:", err)
				return "", err
			}
			// 否则继续重试
			continue
		}

		// 检查消息是否有效
		if isValidCommitMessage(message) {
			return message, nil
		}

		// 如果无效且开启了 debug 模式，打印无效的消息
		if debug {
			fmt.Println("=== DEBUG: Invalid commit message (attempt", attempt+1, ") ===")
			fmt.Println(message)
			fmt.Println("=== END DEBUG ===")
		}

		// 如果无效且不是最后一次尝试，继续重试
		if attempt < MaxRetries {
			fmt.Println("Invalid commit message, retrying...")
			continue
		}

		fmt.Println("Failed to generate commit message after", MaxRetries, "attempts, error:", err)
	}

	// 如果所有尝试都失败，返回最后一次的结果（即使无效）
	return message, nil
}
