package commands

import (
	"fmt"
	"os"

	"github.com/gotoailab/llmgit/internal/config"
	"github.com/gotoailab/llmgit/internal/i18n"
	"github.com/gotoailab/llmhub"
)

// HandleInit handles the init command
func HandleInit(args []string) {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "%s%s\n", i18n.T("error_prefix"), i18n.T("init_usage"))
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

	// Validate provider
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

