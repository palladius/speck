// Package cli wires speck's cobra commands: oneshot and chat.
package cli

import (
	"github.com/spf13/cobra"
)

var (
	flagModel       string
	flagAPIKey      string
	flagBaseDir     string
	flagInspiration string
	flagForce       bool
)

var knownModels = []string{
	"gemini-flash-latest",
	"gemini-pro-latest",
	"gemini-2.5-flash",
	"gemini-2.5-pro",
}

// NewRootCmd builds speck's root cobra command.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "speck",
		Short:         "Turn a rough idea into a structured SPEC.md",
		Long:          "speck turns a rough idea into a structured SPEC.md you can feed to any AI coding assistant.",
		SilenceUsage:  true,
		SilenceErrors: false,
	}

	root.PersistentFlags().StringVar(&flagModel, "model", "", "Gemini model to use (default: gemini-2.5-flash)")
	root.PersistentFlags().StringVar(&flagAPIKey, "api-key", "", "Gemini API key (overrides GEMINI_API_KEY)")
	root.PersistentFlags().StringVar(&flagBaseDir, "base-dir", ".", "Base directory for the <category>/<slug> output")
	root.PersistentFlags().StringVarP(&flagInspiration, "source-of-inspiration", "s", "", "Directory of .txt/.md/.html files to use as supporting context")
	root.PersistentFlags().BoolVar(&flagForce, "force", false, "Overwrite an existing SPEC.md at the resolved path")

	_ = root.RegisterFlagCompletionFunc("model", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return knownModels, cobra.ShellCompDirectiveNoFileComp
	})

	root.AddCommand(newOneshotCmd())
	root.AddCommand(newChatCmd())

	return root
}
