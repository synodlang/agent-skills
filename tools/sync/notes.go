package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// notes returns the body of the CHANGELOG section whose level-2 heading is
// "## <tag>", optionally followed by whitespace and more text (for example
// "## v0.2.0 — 2026-10-22"). The body runs to the next "## " heading, without
// leading and trailing blank lines. A missing or empty section is an error.
// Release pages show this text, never a compare link.
func notes(changelog, tag string) (string, error) {
	if tag == "" {
		return "", fmt.Errorf("empty tag")
	}
	lines := strings.Split(strings.ReplaceAll(changelog, "\r\n", "\n"), "\n")
	start := -1
	for i, l := range lines {
		rest, ok := strings.CutPrefix(strings.TrimRight(l, " \t"), "## "+tag)
		if ok && (rest == "" || rest[0] == ' ' || rest[0] == '\t') {
			start = i + 1
			break
		}
	}
	if start < 0 {
		return "", fmt.Errorf("no section %q in the changelog", "## "+tag)
	}
	end := len(lines)
	for i := start; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			end = i
			break
		}
	}
	body := strings.Trim(strings.Join(lines[start:end], "\n"), "\n \t")
	if body == "" {
		return "", fmt.Errorf("section %q in the changelog is empty", "## "+tag)
	}
	return body + "\n", nil
}

func printNotes(root, tag string) error {
	b, err := os.ReadFile(filepath.Join(root, "CHANGELOG.md"))
	if err != nil {
		return err
	}
	s, err := notes(string(b), tag)
	if err != nil {
		return err
	}
	_, err = fmt.Print(s)
	return err
}
