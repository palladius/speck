package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/palladius/speck/internal/genai"
	"github.com/palladius/speck/internal/spec"
)

// writeSpec resolves the idea's inline-vs-sidecar storage, then renders and
// writes SPEC.md (and original_idea.md, if the idea didn't fit inline) to dir.
func writeSpec(dir, specPath string, fm spec.Frontmatter, idea string, result genai.SpecResult) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	sidecar, needsSidecar := fm.SetIdea(idea, "original_idea.md")
	if needsSidecar {
		if err := os.WriteFile(filepath.Join(dir, "original_idea.md"), []byte(sidecar), 0o644); err != nil {
			return fmt.Errorf("write original_idea.md: %w", err)
		}
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
