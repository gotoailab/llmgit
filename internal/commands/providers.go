package commands

import (
	"fmt"

	"github.com/gotoailab/llmgit/internal/i18n"
	"github.com/gotoailab/llmhub"
)

// HandleProviders handles the providers command
func HandleProviders() {
	fmt.Println(i18n.T("providers_title"))
	for _, p := range llmhub.AllProviders() {
		fmt.Println(p)
	}
}

