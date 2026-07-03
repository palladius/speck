package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/palladius/speck/internal/genai"
	"github.com/palladius/speck/internal/spec"
)

// writeSpec renders and writes SPEC.md, plus a sidecar input_prompt.md
// carrying the exact idea that produced it, to dir.
func writeSpec(dir, specPath string, fm spec.Frontmatter, idea string, result genai.SpecResult) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	fm.IdeaFile = spec.InputPromptFileName
	if err := os.WriteFile(filepath.Join(dir, spec.InputPromptFileName), []byte(idea), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", spec.InputPromptFileName, err)
	}

	doc, err := spec.Render(fm, result.Title, result.SpecBody)
	if err != nil {
		return err
	}
	if err := os.WriteFile(specPath, []byte(doc), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", specPath, err)
	}
	return nil
}
