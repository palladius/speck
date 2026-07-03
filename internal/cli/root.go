// Package cli wires speck's cobra commands: oneshot and chat.
package cli

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/palladius/speck/internal/version"
)

// outputDirEnvVar overrides the default output directory when
// --output-dir/-o isn't passed explicitly.
const outputDirEnvVar = "SPECK_OUTPUT_DIR"

// defaultOutputDir is used when neither --output-dir nor SPECK_OUTPUT_DIR is set.
const defaultOutputDir = "out"

var (
	flagModel       string
	flagAPIKey      string
	flagOutputDir   string
	flagInspiration string
	flagForce       bool
)

// resolveOutputDir returns the effective output base directory: the
// --output-dir flag if set, else $SPECK_OUTPUT_DIR, else defaultOutputDir.
func resolveOutputDir() string {
	if flagOutputDir != "" {
		return flagOutputDir
	}
	if v := os.Getenv(outputDirEnvVar); v != "" {
		return v
	}
	return defaultOutputDir
}

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
		Version:       version.Version,
		SilenceUsage:  true,
		SilenceErrors: false,
	}

	root.PersistentFlags().StringVar(&flagModel, "model", "", "Gemini model to use (default: gemini-flash-latest)")
	root.PersistentFlags().StringVar(&flagAPIKey, "api-key", "", "Gemini API key (overrides GEMINI_API_KEY)")
	root.PersistentFlags().StringVarP(&flagOutputDir, "output-dir", "o", "", "Base directory for the <category>/<slug> output (default: $SPECK_OUTPUT_DIR or ./out)")
	root.PersistentFlags().StringVarP(&flagInspiration, "source-of-inspiration", "s", "", "Directory of .txt/.md/.html files to use as supporting context")
	root.PersistentFlags().BoolVar(&flagForce, "force", false, "Overwrite an existing SPEC.md at the resolved path")

	_ = root.RegisterFlagCompletionFunc("model", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return knownModels, cobra.ShellCompDirectiveNoFileComp
	})

	root.AddCommand(newOneshotCmd())
	root.AddCommand(newChatCmd())

	return root
}
