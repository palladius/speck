// Package inspire scans a directory of supporting notes to feed into a
// speck prompt as secondary context, alongside (not instead of) the idea.
package inspire

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var allowedExt = map[string]bool{
	".txt":  true,
	".md":   true,
	".html": true,
}

// MaxTotalChars caps how much inspiration content is fed into a prompt.
const MaxTotalChars = 50_000

// Result is the outcome of scanning a directory for inspiration material.
type Result struct {
	// Files lists the paths (relative to the scanned directory) actually
	// read, in the order they were included.
	Files []string
	// Context is the concatenated file contents, each under a header
	// naming its relative path.
	Context string
	// Truncated is true if MaxTotalChars was hit before every matching
	// file could be included.
	Truncated bool
}

// Scan walks dir recursively, reading every .txt/.md/.html file it finds
// (in sorted order) until MaxTotalChars is reached.
func Scan(dir string) (Result, error) {
	var res Result

	var paths []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !allowedExt[strings.ToLower(filepath.Ext(path))] {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return res, fmt.Errorf("scan %s: %w", dir, err)
	}
	sort.Strings(paths)

	var b strings.Builder
	for _, p := range paths {
		content, err := os.ReadFile(p)
		if err != nil {
			return res, fmt.Errorf("read %s: %w", p, err)
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			rel = p
		}
		header := fmt.Sprintf("--- %s ---\n", rel)
		if b.Len()+len(header)+len(content) > MaxTotalChars {
			res.Truncated = true
			break
		}
		b.WriteString(header)
		b.Write(content)
		b.WriteString("\n\n")
		res.Files = append(res.Files, rel)
	}
	res.Context = b.String()
	return res, nil
}
