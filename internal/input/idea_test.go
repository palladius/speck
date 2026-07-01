package input

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveArgTakesPriority(t *testing.T) {
	got, err := Resolve("  an idea from the arg  ", "/should/not/be/read")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if got != "an idea from the arg" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "idea.txt")
	if err := os.WriteFile(path, []byte("  idea from file  \n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	got, err := Resolve("", path)
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if got != "idea from file" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveFromIdeaMdFallback(t *testing.T) {
	dir := t.TempDir()
	restoreWD := chdir(t, dir)
	defer restoreWD()

	if err := os.WriteFile(ideaFallbackFile, []byte("idea from IDEA.md"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	restoreStdin := withEmptyStdin(t)
	defer restoreStdin()

	got, err := Resolve("", "")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if got != "idea from IDEA.md" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveFromPipedStdin(t *testing.T) {
	dir := t.TempDir()
	restoreWD := chdir(t, dir)
	defer restoreWD()

	restoreStdin := withStdinContent(t, "idea piped in via stdin\n")
	defer restoreStdin()

	got, err := Resolve("", "")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if got != "idea piped in via stdin" {
		t.Fatalf("got %q", got)
	}
}

func TestPromptInteractiveEndsOnBlankLine(t *testing.T) {
	r := strings.NewReader("line one\nline two\n\nshould not be read\n")
	got, err := promptInteractive(r)
	if err != nil {
		t.Fatalf("promptInteractive failed: %v", err)
	}
	want := "line one\nline two"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestPromptInteractiveEndsOnEOF(t *testing.T) {
	r := strings.NewReader("only line, then ctrl-d")
	got, err := promptInteractive(r)
	if err != nil {
		t.Fatalf("promptInteractive failed: %v", err)
	}
	if got != "only line, then ctrl-d" {
		t.Fatalf("got %q", got)
	}
}

// --- test helpers ---

func chdir(t *testing.T, dir string) func() {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	return func() { _ = os.Chdir(old) }
}

// withStdinContent replaces os.Stdin with a pipe pre-loaded with content,
// so piped()/io.ReadAll(os.Stdin) behave as they would for real piped input.
func withStdinContent(t *testing.T, content string) func() {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	if _, err := w.WriteString(content); err != nil {
		t.Fatalf("write: %v", err)
	}
	w.Close()
	old := os.Stdin
	os.Stdin = r
	return func() { os.Stdin = old; r.Close() }
}

// withEmptyStdin gives an already-closed pipe as stdin: piped() still
// reports true (it's not a char device), but reading it yields no bytes,
// exercising the "piped but empty, fall through to IDEA.md" path.
func withEmptyStdin(t *testing.T) func() {
	return withStdinContent(t, "")
}
