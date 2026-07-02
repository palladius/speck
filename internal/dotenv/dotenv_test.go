package dotenv

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSetsUnsetVars(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	content := "# a comment\n\nSPECK_OUTPUT_DIR=\"out/\"\nQUOTED='hello world'\nBARE=plain\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	for _, k := range []string{"SPECK_OUTPUT_DIR", "QUOTED", "BARE"} {
		os.Unsetenv(k)
	}

	if err := Load(path); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	cases := map[string]string{
		"SPECK_OUTPUT_DIR": "out/",
		"QUOTED":           "hello world",
		"BARE":             "plain",
	}
	for k, want := range cases {
		if got := os.Getenv(k); got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}
}

func TestLoadDoesNotOverrideExistingEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("EXISTING=fromfile\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	t.Setenv("EXISTING", "fromenv")
	if err := Load(path); err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if got := os.Getenv("EXISTING"); got != "fromenv" {
		t.Errorf("expected existing env var to win, got %q", got)
	}
}

func TestLoadMissingFileIsNotAnError(t *testing.T) {
	if err := Load(filepath.Join(t.TempDir(), "does-not-exist.env")); err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
}
