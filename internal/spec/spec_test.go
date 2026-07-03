package spec

import (
	"os"
	"strings"
	"testing"
)

func TestRenderAndParseRoundTrip(t *testing.T) {
	fm := Frontmatter{
		SpeckVersion:     "0.1",
		Mode:             "oneshot",
		IdeaFile:         InputPromptFileName,
		InspirationDir:   "./notes",
		InspirationFiles: []string{"a.md", "b.txt"},
		CreatedAt:        "2026-07-01T15:20:00Z",
		Model:            "gemini-2.5-flash",
		Tokens:           TokenUsage{Prompt: 100, Output: 50, Total: 150},
	}
	doc, err := Render(fm, "Tetris Clone", "## Problem Statement\nsomething\n")
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if !strings.HasPrefix(doc, "---\n") {
		t.Fatalf("expected document to start with frontmatter delimiter, got %q", doc[:20])
	}
	if !strings.Contains(doc, "# Tetris Clone") {
		t.Fatalf("expected title heading in document")
	}

	parsedFM, body, err := ParseFrontmatter(doc)
	if err != nil {
		t.Fatalf("ParseFrontmatter failed: %v", err)
	}
	if parsedFM.Mode != fm.Mode || parsedFM.Model != fm.Model || parsedFM.IdeaFile != fm.IdeaFile {
		t.Fatalf("round-tripped frontmatter mismatch: %+v", parsedFM)
	}
	if len(parsedFM.InspirationFiles) != 2 {
		t.Fatalf("expected 2 inspiration files, got %v", parsedFM.InspirationFiles)
	}
	if !strings.Contains(body, "Problem Statement") {
		t.Fatalf("expected body to contain section header, got %q", body)
	}
}

func TestSanitize(t *testing.T) {
	cases := map[string]string{
		"Tetris Clone":   "tetris-clone",
		"  weird!! Name": "weird-name",
		"":               "misc",
		"---":            "misc",
		"already-good":   "already-good",
	}
	for in, want := range cases {
		if got := Sanitize(in); got != want {
			t.Errorf("Sanitize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestResolvePath(t *testing.T) {
	dir, specPath := ResolvePath("/base", "Games", "Tetris Clone")
	if dir != "/base/games/tetris-clone" {
		t.Fatalf("unexpected dir: %q", dir)
	}
	if specPath != "/base/games/tetris-clone/SPEC.md" {
		t.Fatalf("unexpected specPath: %q", specPath)
	}
}

func TestCheckOverwrite(t *testing.T) {
	dir := t.TempDir()
	specPath := dir + "/SPEC.md"
	if err := CheckOverwrite(specPath, false); err != nil {
		t.Fatalf("expected no error for nonexistent file, got %v", err)
	}
	if err := os.WriteFile(specPath, []byte("x"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := CheckOverwrite(specPath, false); err != ErrExists {
		t.Fatalf("expected ErrExists, got %v", err)
	}
	if err := CheckOverwrite(specPath, true); err != nil {
		t.Fatalf("expected no error with force, got %v", err)
	}
}
