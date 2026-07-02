package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/palladius/speck/internal/genai"
	"github.com/palladius/speck/internal/input"
	"github.com/palladius/speck/internal/inspire"
	"github.com/palladius/speck/internal/spec"
	"github.com/palladius/speck/internal/ui"
)

var chatFile string

// doneCommand lets the human end an interview phase early, mirroring
// "keep asking clarifying questions until I tell you to stop" from
// docs/META-SPECS.md.
const doneCommand = "/done"

// phaseDoneMarker / readyMarker are emitted by the model to end phase 1
// (problem) and phase 2 (acceptance criteria) respectively.
const (
	phaseDoneMarker = "<<PHASE_DONE>>"
	readyMarker     = "<<READY>>"
)

const phase1SystemPrompt = `You are speck's interviewer for the PROBLEM phase of a spec-writing interview.

Ask exactly one focused clarifying question per turn about the problem being solved: what it is, who it's for, and what constraints matter. Do not write any spec content yet.

Do not be sycophantic. If an answer is vague, underspecified, or self-contradictory, push back and ask a sharper follow-up instead of praising it.

Once you understand the problem well enough to write a solid "Problem Statement", "Goals", and "Non-Goals" section, reply with ONLY the exact text ` + phaseDoneMarker + ` and nothing else.`

const phase2SystemPrompt = `You are speck's interviewer for the ACCEPTANCE CRITERIA phase of a spec-writing interview. The problem is already understood from the earlier turns.

Ask exactly one focused clarifying question per turn about how success will be judged: what "done" looks like, and how it will be verified.

Do not be sycophantic. If a criterion is vague or untestable, push back and ask a sharper follow-up instead of accepting it.

Once acceptance criteria are clear, reply with ONLY the exact text ` + readyMarker + ` and nothing else.`

func newChatCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chat [idea]",
		Short: "Interview you with clarifying questions, then write a SPEC.md",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runChat,
	}
	cmd.Flags().StringVar(&chatFile, "file", "", "Read the initial idea from a file")
	return cmd
}

func runChat(cmd *cobra.Command, args []string) error {
	var argIdea string
	if len(args) > 0 {
		argIdea = args[0]
	}

	idea, err := input.Resolve(argIdea, chatFile)
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

	var usage genai.Usage
	var transcript strings.Builder
	transcript.WriteString("# speck interview transcript\n\n## Idea\n\n" + idea + "\n\n")
	if len(inspiration.Files) > 0 {
		transcript.WriteString("## Inspiration files used\n\n" + strings.Join(inspiration.Files, ", ") + "\n\n")
	}

	history := []*genai.Content{genai.UserText(buildInitialPrompt(idea, inspiration))}
	reader := bufio.NewReader(os.Stdin)

	transcript.WriteString("## Phase 1: Problem\n\n")
	history, err = runInterviewPhase(ctx, client, phase1SystemPrompt, phaseDoneMarker, history, &usage, &transcript, reader)
	if err != nil {
		return err
	}

	transcript.WriteString("\n## Phase 2: Acceptance Criteria\n\n")
	history, err = runInterviewPhase(ctx, client, phase2SystemPrompt, readyMarker, history, &usage, &transcript, reader)
	if err != nil {
		return err
	}

	ui.Thinking("synthesizing spec...")
	result, finalUsage, err := client.GenerateSpec(ctx, finalSynthesisSystemPrompt, history)
	if err != nil {
		return fmt.Errorf("synthesize spec: %w", err)
	}
	usage.Prompt += finalUsage.Prompt
	usage.Output += finalUsage.Output
	usage.Total += finalUsage.Total

	dir, specPath := spec.ResolvePath(flagBaseDir, result.Category, result.Slug)
	if err := spec.CheckOverwrite(specPath, flagForce); err != nil {
		return err
	}

	fm := spec.Frontmatter{
		SpeckVersion:   "0.1",
		Mode:           "interactive",
		CreatedAt:      time.Now().UTC().Format(time.RFC3339),
		Model:          client.Model,
		TranscriptFile: "speck_transcript.md",
		Tokens:         spec.TokenUsage{Prompt: usage.Prompt, Output: usage.Output, Total: usage.Total},
	}
	if len(inspiration.Files) > 0 {
		fm.InspirationDir = flagInspiration
		fm.InspirationFiles = inspiration.Files
	}

	if err := writeSpec(dir, specPath, fm, idea, result); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "speck_transcript.md"), []byte(transcript.String()), 0o644); err != nil {
		return fmt.Errorf("write transcript: %w", err)
	}

	ui.Success("wrote %s", specPath)
	return nil
}

// runInterviewPhase drives one phase of the interview: ask a question,
// print it, read the human's answer (or /done / EOF to force-exit early),
// and repeat until the model emits doneMarker. Returns the updated history.
func runInterviewPhase(ctx context.Context, client *genai.Client, systemPrompt, doneMarker string, history []*genai.Content, usage *genai.Usage, transcript *strings.Builder, reader *bufio.Reader) ([]*genai.Content, error) {
	for {
		reply, turnUsage, err := client.Ask(ctx, systemPrompt, history)
		if err != nil {
			return history, fmt.Errorf("ask model: %w", err)
		}
		usage.Prompt += turnUsage.Prompt
		usage.Output += turnUsage.Output
		usage.Total += turnUsage.Total

		reply = strings.TrimSpace(reply)
		if strings.Contains(reply, doneMarker) {
			return history, nil
		}

		ui.Question(reply)
		fmt.Print("> ")
		answer, err := reader.ReadString('\n')
		eof := errors.Is(err, io.EOF)
		if err != nil && !eof {
			return history, fmt.Errorf("read answer: %w", err)
		}
		answer = strings.TrimSpace(answer)

		if eof || answer == doneCommand {
			return history, nil
		}

		history = append(history, genai.ModelText(reply), genai.UserText(answer))
		fmt.Fprintf(transcript, "**Q:** %s\n\n**A:** %s\n\n", reply, answer)
	}
}
