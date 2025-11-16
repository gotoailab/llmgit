package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gotoailab/llmgit/internal/commands"
	"github.com/gotoailab/llmgit/internal/config"
	"github.com/gotoailab/llmgit/internal/i18n"
	"github.com/gotoailab/llmhub"
)

func main() {
	// Initialize i18n (default to English)
	i18n.InitLanguage()

	// Parse global language option if present
	args := parseGlobalLangOption(os.Args[1:])

	if len(args) < 1 {
		commands.PrintUsage()
		os.Exit(1)
	}

	// Check if it's an AI command (starts with "ai")
	if args[0] == "ai" {
		if len(args) < 2 {
			commands.PrintUsage()
			os.Exit(1)
		}
		handleAICommand(args[1:])
		return
	}

	// Check for help command (without ai prefix for convenience)
	if args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		commands.PrintUsage()
		return
	}

	// All other commands are forwarded to git
	commands.HandleGitCommand(args)
}

func handleAICommand(args []string) {
	if len(args) < 1 {
		commands.PrintUsage()
		os.Exit(1)
	}

	command := args[0]

	// Init command
	if command == "init" {
		commands.HandleInit(args[1:])
		return
	}

	// Help command (no initialization needed)
	if command == "help" || command == "--help" || command == "-h" {
		commands.PrintUsage()
		return
	}

	// List all supported providers (no initialization needed)
	if command == "providers" || command == "list-providers" {
		commands.HandleProviders()
		return
	}

	// Show version info (no initialization needed)
	if command == "version" {
		commands.HandleVersion()
		return
	}

	// Set commit prompt (needs initialization first)
	if command == "set-prompt" {
		commands.HandleSetPrompt(args[1:])
		return
	}

	// CHANGELOG generation (needs initialization)
	if command == "changelog" {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), i18n.T("error_not_initialized"))
			os.Exit(1)
		}
		client, err := llmhub.NewClient(llmhub.ClientConfig{
			APIKey:   cfg.APIKey,
			Provider: llmhub.Provider(cfg.Provider),
			Model:    cfg.Model,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), fmt.Sprintf(i18n.T("error_create_client"), err))
			os.Exit(1)
		}
		commands.HandleChangelog(client, args[1:])
		return
	}

	// PR description generation (needs initialization)
	if command == "pr" {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), i18n.T("error_not_initialized"))
			os.Exit(1)
		}
		client, err := llmhub.NewClient(llmhub.ClientConfig{
			APIKey:   cfg.APIKey,
			Provider: llmhub.Provider(cfg.Provider),
			Model:    cfg.Model,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), fmt.Sprintf(i18n.T("error_create_client"), err))
			os.Exit(1)
		}
		commands.HandlePR(client, args[1:])
		return
	}

	// Check if initialized
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), i18n.T("error_not_initialized"))
		os.Exit(1)
	}

	// Create LLM client
	client, err := llmhub.NewClient(llmhub.ClientConfig{
		APIKey:   cfg.APIKey,
		Provider: llmhub.Provider(cfg.Provider),
		Model:    cfg.Model,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), fmt.Sprintf(i18n.T("error_create_client"), err))
		os.Exit(1)
	}

	// Handle special commands
	switch command {
	case "commit":
		commands.HandleCommit(client, cfg, args[1:])
	case "review":
		commands.HandleReview(client, cfg, args[1:])
	case "diff":
		commands.HandleDiff(client, cfg, args[1:])
	case "explain":
		commands.HandleExplain(client, cfg, args[1:])
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
				lang := args[i+1]
				i18n.SetLanguage(lang)
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
