package cli

import (
	"strings"

	"github.com/palladius/speck/internal/inspire"
)

// buildInitialPrompt combines the idea with any inspiration material,
// clearly marking the latter as secondary context.
func buildInitialPrompt(idea string, inspiration inspire.Result) string {
	var b strings.Builder
	b.WriteString("Idea:\n")
	b.WriteString(idea)
	if inspiration.Context != "" {
		b.WriteString("\n\nSupporting inspiration material (secondary context, not the primary ask):\n")
		b.WriteString(inspiration.Context)
	}
	return b.String()
}

const oneshotSystemPrompt = `You are speck, a tool that turns a rough software idea into a complete, well-structured spec.

Produce a spec_body in markdown using exactly these "## " section headers, in this order:
Problem Statement, Goals, Non-Goals, Technical Plan / Approach, Alternatives Considered, Implementation Plan, Open Questions.

Also choose a short lowercase-hyphenated "category" (e.g. games, tools, web-apps) and "slug" (e.g. tetris-clone) for where this spec should live, and a human-readable "title".

Be concrete and specific. Do not pad sections with filler — if there is genuinely nothing to say in a section (e.g. no real alternatives were considered), say so briefly rather than inventing content.`

const finalSynthesisSystemPrompt = `You are speck. Based on the full interview transcript so far (the original idea plus every clarifying question and answer), produce the structured spec JSON.

Produce a spec_body in markdown using exactly these "## " section headers, in this order:
Problem Statement, Goals, Non-Goals, Technical Plan / Approach, Alternatives Considered, Implementation Plan, Open Questions.

Ground every section in what was actually discussed in the transcript. Also choose a short lowercase-hyphenated "category", "slug", and a human-readable "title".`
