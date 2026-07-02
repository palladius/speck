package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/palladius/speck/internal/genai"
	"github.com/palladius/speck/internal/input"
	"github.com/palladius/speck/internal/inspire"
	"github.com/palladius/speck/internal/spec"
	"github.com/palladius/speck/internal/ui"
)

var oneshotFile string

func newOneshotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oneshot [idea]",
		Short: "Generate a SPEC.md from an idea in a single call",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runOneshot,
	}
	cmd.Flags().StringVar(&oneshotFile, "file", "", "Read the idea from a file")
	return cmd
}

func runOneshot(cmd *cobra.Command, args []string) error {
	var argIdea string
	if len(args) > 0 {
		argIdea = args[0]
	}

	idea, err := input.Resolve(argIdea, oneshotFile)
	if err != nil {
		return fmt.Errorf("resolve idea: %w", err)
	}

	var inspiration inspire.Result
	if flagInspiration != "" {
		inspiration, err = inspire.Scan(flagInspiration)
		if err != nil {
			return fmt.Errorf("scan inspiration dir: %w", err)
		}
		if inspiration.Truncated {
			ui.Warn("inspiration material truncated at %d chars", inspire.MaxTotalChars)
		}
	}

	ctx := context.Background()
	client, err := genai.New(ctx, flagAPIKey, flagModel)
	if err != nil {
		return err
	}

	ui.Thinking("generating spec...")
	prompt := buildInitialPrompt(idea, inspiration)
	result, usage, err := client.GenerateSpec(ctx, oneshotSystemPrompt, []*genai.Content{genai.UserText(prompt)})
	if err != nil {
		return fmt.Errorf("generate spec: %w", err)
	}

	dir, specPath := spec.ResolvePath(flagBaseDir, result.Category, result.Slug)
	if err := spec.CheckOverwrite(specPath, flagForce); err != nil {
		return err
	}

	fm := spec.Frontmatter{
		SpeckVersion: "0.1",
		Mode:         "oneshot",
		CreatedAt:    time.Now().UTC().Format(time.RFC3339),
		Model:        client.Model,
		Tokens:       spec.TokenUsage{Prompt: usage.Prompt, Output: usage.Output, Total: usage.Total},
	}
	if len(inspiration.Files) > 0 {
		fm.InspirationDir = flagInspiration
		fm.InspirationFiles = inspiration.Files
	}

	if err := writeSpec(dir, specPath, fm, idea, result); err != nil {
		return err
	}

	ui.Success("wrote %s", specPath)
	return nil
}
