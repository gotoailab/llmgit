package commands

import (
	"fmt"

	"github.com/gotoailab/llmgit/internal/i18n"
)

// PrintUsage prints the usage information
func PrintUsage() {
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
	fmt.Printf("  %s  - %s\n", i18n.T("changelog_usage"), i18n.T("changelog_usage_desc"))
	fmt.Printf("  %s  - %s\n", i18n.T("pr_usage"), i18n.T("pr_usage_desc"))
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

