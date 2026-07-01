// Package input resolves the idea text speck should turn into a spec.
package input

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"speck/internal/ui"
)

// ideaFallbackFile is checked in the current directory when no idea is
// given via argument, --file, or piped stdin.
const ideaFallbackFile = "IDEA.md"

// Resolve determines the idea text, trying in order: the positional CLI
// arg, --file, piped stdin, ./IDEA.md, and finally an interactive
// multi-line prompt.
func Resolve(argIdea, filePath string) (string, error) {
	if s := strings.TrimSpace(argIdea); s != "" {
		return s, nil
	}

	if filePath != "" {
		b, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("read --file %s: %w", filePath, err)
		}
		if s := strings.TrimSpace(string(b)); s != "" {
			return s, nil
		}
		return "", fmt.Errorf("--file %s is empty", filePath)
	}

	if piped(os.Stdin) {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("read stdin: %w", err)
		}
		if s := strings.TrimSpace(string(b)); s != "" {
			return s, nil
		}
	}

	if b, err := os.ReadFile(ideaFallbackFile); err == nil {
		if s := strings.TrimSpace(string(b)); s != "" {
			return s, nil
		}
	}

	return promptInteractive(os.Stdin)
}

// piped reports whether f looks like non-interactive input (a pipe or
// redirected file) rather than an attached terminal.
func piped(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice == 0
}

// promptInteractive reads multi-line input from r, terminated by a blank
// line (after some content) or EOF (e.g. Ctrl-D).
func promptInteractive(r io.Reader) (string, error) {
	ui.Info("Enter your idea, then press Enter twice (or Ctrl-D) to finish:")
	scanner := bufio.NewScanner(r)
	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			if len(lines) > 0 {
				break
			}
			continue
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read interactive input: %w", err)
	}
	idea := strings.TrimSpace(strings.Join(lines, "\n"))
	if idea == "" {
		return "", fmt.Errorf("no idea provided")
	}
	return idea, nil
}
