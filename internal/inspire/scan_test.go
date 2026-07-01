package inspire

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanFiltersByExtension(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "notes.md", "# notes")
	write(t, dir, "readme.txt", "plain text")
	write(t, dir, "page.html", "<p>hi</p>")
	write(t, dir, "ignored.go", "package main")
	write(t, dir, "ignored.png", "binary")

	res, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(res.Files) != 3 {
		t.Fatalf("expected 3 files, got %v", res.Files)
	}
	for _, want := range []string{"notes.md", "readme.txt", "page.html"} {
		if !contains(res.Files, want) {
			t.Errorf("expected %q to be included, got %v", want, res.Files)
		}
	}
	if strings.Contains(res.Context, "package main") {
		t.Errorf("expected .go file content to be excluded")
	}
	if res.Truncated {
		t.Errorf("did not expect truncation for small fixture")
	}
}

func TestScanTruncatesAtSizeCap(t *testing.T) {
	dir := t.TempDir()
	big := strings.Repeat("a", MaxTotalChars)
	write(t, dir, "big.txt", big)
	write(t, dir, "small.txt", "small")

	res, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if !res.Truncated {
		t.Errorf("expected truncation when content exceeds MaxTotalChars")
	}
}

func TestScanNestedDirs(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	write(t, sub, "nested.md", "nested content")

	res, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(res.Files) != 1 || !strings.Contains(res.Files[0], "nested.md") {
		t.Fatalf("expected nested file to be found, got %v", res.Files)
	}
}

func write(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func contains(list []string, want string) bool {
	for _, v := range list {
		if v == want {
			return true
		}
	}
	return false
}
