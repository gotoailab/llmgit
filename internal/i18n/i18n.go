package i18n

import (
	"os"
	"strings"
)

// Language represents the supported languages
type Language string

const (
	LangEN Language = "en"
	LangZH Language = "zh"
)

var currentLang Language = LangEN

// SetLanguage sets the current language
func SetLanguage(lang string) {
	lang = strings.ToLower(lang)
	if lang == "zh" || lang == "cn" || lang == "chinese" {
		currentLang = LangZH
	} else {
		currentLang = LangEN
	}
}

// GetLanguage returns the current language
func GetLanguage() Language {
	return currentLang
}

// InitLanguage initializes language from environment variable or default to English
func InitLanguage() {
	// Check environment variable first
	if lang := os.Getenv("LLMGIT_LANG"); lang != "" {
		SetLanguage(lang)
		return
	}
	// Default to English
	currentLang = LangEN
}

// T returns the translated string based on current language
func T(key string) string {
	if currentLang == LangZH {
		return zhMessages[key]
	}
	return enMessages[key]
}

// Messages map
var enMessages = map[string]string{
	// General
	"app_name":     "llmgit - AI-Enhanced Git Tool",
	"usage":        "Usage",
	"examples":     "Examples",
	"error":        "Error",
	"error_prefix": "Error: ",

	// Help command
	"help_usage_desc": "Show this help message",

	// Init command
	"init_usage":                  "llmgit ai init <provider> <api-key> [model]",
	"init_usage_desc":             "Initialize project configuration",
	"init_supported_providers":    "Supported providers:",
	"init_error_invalid_provider": "Unsupported provider: %s",
	"init_error_save_failed":      "Failed to save configuration: %v",
	"init_success":                "Configuration saved to ~/.llmgit/config.json",
	"init_provider":               "Provider: %s",
	"init_model":                  "Model: %s",
	"init_model_default":          "Model: (using default model)",

	// Providers command
	"providers_usage":      "llmgit ai providers",
	"providers_usage_desc": "List all supported providers",
	"providers_title":      "Supported providers:",

	// Version command
	"version_usage":      "llmgit ai version",
	"version_usage_desc": "Show version and build information",

	// Set prompt command
	"set_prompt_usage":      "llmgit ai set-prompt [prompt|file]",
	"set_prompt_usage_desc": "Set custom commit message prompt template",
	"prompt_current":        "Current commit prompt template:",
	"prompt_not_set":        "No custom commit prompt template set. Using default template.",
	"prompt_usage":          "Usage: llmgit ai set-prompt \"<template>\" or llmgit ai set-prompt <file>",
	"prompt_saved":          "Commit prompt template saved successfully",
	"error_read_file":       "Failed to read file %s: %v",
	"error_save_prompt":     "Failed to save prompt template: %v",

	// Changelog command
	"changelog_usage":          "llmgit ai changelog [range] [--output file] [--format markdown|json|yaml]",
	"changelog_usage_desc":     "Auto-generate CHANGELOG from commit history",
	"changelog_generating":     "Generating CHANGELOG...",
	"changelog_no_commits":     "No commits found in the specified range",
	"changelog_saved":          "CHANGELOG saved to %s",
	"error_get_commits":        "Failed to get commit history: %v",
	"error_generate_changelog": "Failed to generate CHANGELOG: %v",
	"error_write_file":         "Failed to write file %s: %v",

	// PR command
	"pr_usage":              "llmgit ai pr [target-branch] [--base base-branch] [--copy]",
	"pr_usage_desc":         "Generate Pull Request description",
	"pr_generating":         "Generating PR description...",
	"pr_no_changes":         "No changes found between branches",
	"pr_copied":             "PR description copied to clipboard",
	"error_get_branch_diff": "Failed to get branch diff: %v",
	"error_generate_pr":     "Failed to generate PR description: %v",
	"error_copy_clipboard":  "Failed to copy to clipboard: %v",

	// Common errors
	"error_not_initialized":      "Project not initialized, please run 'llmgit ai init' first",
	"error_unknown_command":      "Unknown command",
	"error_run_help":             "Run 'llmgit help' for usage information",
	"error_create_client":        "Failed to create LLM client: %v",
	"error_get_changes":          "Failed to get changes: %v",
	"error_get_commit_info":      "Failed to get commit information: %v",
	"error_review_commit":        "Failed to review commit: %v",
	"error_generate_message":     "Failed to generate commit message: %v",
	"error_generate_explanation": "Failed to generate explanation: %v",
	"error_no_response":          "No response received",

	// Commit command
	"commit_usage":       "llmgit ai commit [options]",
	"commit_usage_desc":  "Generate commit message with AI and commit",
	"commit_lang_option": "--lang, -l <lang>",
	"commit_lang_desc":   "Specify language (en/zh, default: en)",
	"commit_no_staged":   "No staged changes",
	"commit_generating":  "Generating commit message with AI...",
	"commit_generated":   "Generated commit message:",
	"commit_confirm":     "Use this commit message? (y/n): ",
	"commit_cancelled":   "Cancelled",

	// Review command
	"review_usage":      "llmgit ai review [commit]",
	"review_usage_desc": "Review commit with AI",
	"review_analyzing":  "Analyzing commit: %s...",
	"review_result":     "=== AI Review Result ===",

	// Diff command
	"diff_usage":       "llmgit ai diff [options]",
	"diff_usage_desc":  "Show diff with AI explanation",
	"diff_explanation": "--- AI Explanation ---",

	// Explain command
	"explain_usage":             "llmgit ai explain [file]",
	"explain_usage_desc":        "Explain file changes with AI",
	"explain_analyzing_file":    "Analyzing file: %s...",
	"explain_analyzing_workdir": "Analyzing working directory changes...",
	"explain_no_changes":        "No changes",
	"explain_result":            "=== AI Explanation ===",

	// Other git commands
	"git_command_desc": "Execute other git commands",
}

var zhMessages = map[string]string{
	// General
	"app_name":     "llmgit - AI 增强的 Git 工具",
	"usage":        "用法",
	"examples":     "示例",
	"error":        "错误",
	"error_prefix": "错误: ",

	// Help command
	"help_usage_desc": "显示帮助信息",

	// Init command
	"init_usage":                  "llmgit ai init <provider> <api-key> [model]",
	"init_usage_desc":             "初始化项目配置",
	"init_supported_providers":    "支持的 provider:",
	"init_error_invalid_provider": "不支持的 provider: %s",
	"init_error_save_failed":      "无法保存配置: %v",
	"init_success":                "✓ 配置已保存到 ~/.llmgit/config.json",
	"init_provider":               "  Provider: %s",
	"init_model":                  "  Model: %s",
	"init_model_default":          "  Model: (使用默认模型)",

	// Providers command
	"providers_usage":      "llmgit ai providers",
	"providers_usage_desc": "列出所有支持的 provider",
	"providers_title":      "支持的 provider:",

	// Version command
	"version_usage":      "llmgit ai version",
	"version_usage_desc": "显示版本和构建信息",

	// Set prompt command
	"set_prompt_usage":      "llmgit ai set-prompt [prompt|file]",
	"set_prompt_usage_desc": "设置自定义 commit message prompt 模板",
	"prompt_current":        "当前 commit prompt 模板:",
	"prompt_not_set":        "未设置自定义 commit prompt 模板，使用默认模板。",
	"prompt_usage":          "用法: llmgit ai set-prompt \"<模板>\" 或 llmgit ai set-prompt <文件>",
	"prompt_saved":          "Commit prompt 模板已保存",
	"error_read_file":       "无法读取文件 %s: %v",
	"error_save_prompt":     "无法保存 prompt 模板: %v",

	// Changelog command
	"changelog_usage":          "llmgit ai changelog [范围] [--output 文件] [--format markdown|json|yaml]",
	"changelog_usage_desc":     "基于 commit 历史自动生成 CHANGELOG",
	"changelog_generating":     "正在生成 CHANGELOG...",
	"changelog_no_commits":     "指定范围内没有找到 commit",
	"changelog_saved":          "CHANGELOG 已保存到 %s",
	"error_get_commits":        "无法获取 commit 历史: %v",
	"error_generate_changelog": "无法生成 CHANGELOG: %v",
	"error_write_file":         "无法写入文件 %s: %v",

	// PR command
	"pr_usage":              "llmgit ai pr [目标分支] [--base 基础分支] [--copy]",
	"pr_usage_desc":         "生成 Pull Request 描述",
	"pr_generating":         "正在生成 PR 描述...",
	"pr_no_changes":         "分支之间没有变更",
	"pr_copied":             "PR 描述已复制到剪贴板",
	"error_get_branch_diff": "无法获取分支差异: %v",
	"error_generate_pr":     "无法生成 PR 描述: %v",
	"error_copy_clipboard":  "无法复制到剪贴板: %v",

	// Common errors
	"error_not_initialized":      "项目未初始化，请先运行 'llmgit ai init'",
	"error_unknown_command":      "未知命令",
	"error_run_help":             "运行 'llmgit help' 查看使用说明",
	"error_create_client":        "无法创建 LLM 客户端: %v",
	"error_get_changes":          "无法获取变更: %v",
	"error_get_commit_info":      "无法获取 commit 信息: %v",
	"error_review_commit":        "无法审查 commit: %v",
	"error_generate_message":     "无法生成 commit message: %v",
	"error_generate_explanation": "无法生成解释: %v",
	"error_no_response":          "没有收到响应",

	// Commit command
	"commit_usage":       "llmgit ai commit [options]",
	"commit_usage_desc":  "AI 生成 commit message 并提交",
	"commit_lang_option": "--lang, -l <lang>",
	"commit_lang_desc":   "指定语言 (en/zh, 默认: en)",
	"commit_no_staged":   "没有暂存的变更",
	"commit_generating":  "正在使用 AI 生成 commit message...",
	"commit_generated":   "生成的 commit message:",
	"commit_confirm":     "是否使用此 commit message? (y/n): ",
	"commit_cancelled":   "已取消",

	// Review command
	"review_usage":      "llmgit ai review [commit]",
	"review_usage_desc": "AI 审查 commit",
	"review_analyzing":  "正在分析 commit: %s...",
	"review_result":     "=== AI 审查结果 ===",

	// Diff command
	"diff_usage":       "llmgit ai diff [options]",
	"diff_usage_desc":  "显示 diff 并 AI 解释",
	"diff_explanation": "--- AI 解释 ---",

	// Explain command
	"explain_usage":             "llmgit ai explain [file]",
	"explain_usage_desc":        "AI 语义化解释文件变更",
	"explain_analyzing_file":    "正在分析文件: %s...",
	"explain_analyzing_workdir": "正在分析工作区变更...",
	"explain_no_changes":        "没有变更",
	"explain_result":            "=== AI 解释 ===",

	// Other git commands
	"git_command_desc": "执行其他 git 命令",
}
