// Package spec renders SPEC.md documents: YAML frontmatter for provenance,
// a default section structure for the body, and output path resolution.
package spec

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// InputPromptFileName is the sidecar file that always carries the exact
// prompt (idea, plus any inspiration context) that produced a spec.
const InputPromptFileName = "input_prompt.md"

// TokenUsage records token accounting for a generation, for provenance.
type TokenUsage struct {
	Prompt int32 `yaml:"prompt"`
	Output int32 `yaml:"output"`
	Total  int32 `yaml:"total"`
}

// Frontmatter is the YAML metadata block at the top of every generated SPEC.md.
type Frontmatter struct {
	SpeckVersion     string     `yaml:"speck_version"`
	Mode             string     `yaml:"mode"`
	IdeaFile         string     `yaml:"idea_file"`
	InspirationDir   string     `yaml:"inspiration_dir,omitempty"`
	InspirationFiles []string   `yaml:"inspiration_files,omitempty"`
	TranscriptFile   string     `yaml:"transcript_file,omitempty"`
	CreatedAt        string     `yaml:"created_at"`
	Model            string     `yaml:"model"`
	Tokens           TokenUsage `yaml:"tokens"`
}

// Render renders the full SPEC.md document: YAML frontmatter, a title, and the body.
func Render(fm Frontmatter, title, body string) (string, error) {
	yamlBytes, err := yaml.Marshal(fm)
	if err != nil {
		return "", fmt.Errorf("marshal frontmatter: %w", err)
	}
	var b strings.Builder
	b.WriteString("---\n")
	b.Write(yamlBytes)
	b.WriteString("---\n\n")
	if title != "" {
		b.WriteString("# " + title + "\n\n")
	}
	b.WriteString(strings.TrimRight(body, "\n"))
	b.WriteString("\n")
	return b.String(), nil
}

// ParseFrontmatter splits a rendered SPEC.md document back into its
// frontmatter and body. Used by tests to round-trip Render.
func ParseFrontmatter(doc string) (Frontmatter, string, error) {
	var fm Frontmatter
	if !strings.HasPrefix(doc, "---\n") {
		return fm, "", fmt.Errorf("document has no frontmatter delimiter")
	}
	rest := doc[len("---\n"):]
	end := strings.Index(rest, "\n---\n")
	if end == -1 {
		return fm, "", fmt.Errorf("document has no closing frontmatter delimiter")
	}
	yamlPart := rest[:end]
	body := strings.TrimPrefix(rest[end+len("\n---\n"):], "\n")
	if err := yaml.Unmarshal([]byte(yamlPart), &fm); err != nil {
		return fm, "", fmt.Errorf("unmarshal frontmatter: %w", err)
	}
	return fm, body, nil
}
