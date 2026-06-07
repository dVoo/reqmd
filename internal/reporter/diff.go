package reporter

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"reqmd/internal/diff"
)

const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
)

// FormatDiff produces a human-readable colored text diff.
func FormatDiff(result *diff.Result) string {
	var b strings.Builder

	// Header
	fmt.Fprintf(&b, "%s=== Baseline Diff: %s → %s ===%s\n\n", colorBold, result.Tag1, result.Tag2, colorReset)

	// Per-document diffs
	for _, doc := range result.Documents {
		dirName := doc.Path
		// Just show the dir name, not the full path
		if idx := strings.LastIndex(dirName, "/"); idx >= 0 {
			dirName = dirName[idx+1:]
		}

		fmt.Fprintf(&b, "%s--- %s/ ---%s\n", colorCyan, dirName, colorReset)

		for _, req := range doc.Reqs {
			switch req.Category {
			case diff.Added:
				fmt.Fprintf(&b, "  %s+ %s  (added)%s\n", colorGreen, req.ID, colorReset)
			case diff.Removed:
				fmt.Fprintf(&b, "  %s- %s  (removed)%s\n", colorRed, req.ID, colorReset)
			case diff.Modified:
				fmt.Fprintf(&b, "  %s~ %s  (modified)%s\n", colorYellow, req.ID, colorReset)
				for _, ch := range req.AttrChanges {
					oldStr := formatValue(ch.OldVal)
					newStr := formatValue(ch.NewVal)
					switch {
					case ch.OldVal == nil:
						fmt.Fprintf(&b, "      %s%s: %s (new)%s\n", colorGreen, ch.Key, newStr, colorReset)
					case ch.NewVal == nil:
						fmt.Fprintf(&b, "      %s%s: %s (removed)%s\n", colorRed, ch.Key, oldStr, colorReset)
					default:
						fmt.Fprintf(&b, "      %s%s: %s → %s%s\n", colorYellow, ch.Key, oldStr, newStr, colorReset)
					}
				}
			}
			// Unchanged: skip (not printed)
		}
		// Add blank line between documents
		fmt.Fprintln(&b)
	}

	// Schema diffs
	if len(result.Schemas) > 0 {
		fmt.Fprintf(&b, "%s--- Schema Changes ---%s\n", colorCyan, colorReset)
		for _, sd := range result.Schemas {
			dirName := filepath.Base(sd.Path)
			fmt.Fprintf(&b, "%s%s/%s\n", colorBold, dirName, colorReset)
			for _, key := range sd.Added {
				fmt.Fprintf(&b, "  %s+ %s (new top-level key)%s\n", colorGreen, key, colorReset)
			}
			for _, key := range sd.Removed {
				fmt.Fprintf(&b, "  %s- %s (removed top-level key)%s\n", colorRed, key, colorReset)
			}
			for _, ch := range sd.Changes {
				oldStr := formatValue(ch.OldVal)
				newStr := formatValue(ch.NewVal)
				fmt.Fprintf(&b, "  %s~ %s: %s → %s%s\n", colorYellow, ch.Path, oldStr, newStr, colorReset)
			}
		}
		fmt.Fprintln(&b)
	}

	// Submodule diffs
	if len(result.Submodules) > 0 {
		fmt.Fprintln(&b, renderSubmoduleSection(result.Submodules))
	}

	// Summary
	fmt.Fprintf(&b, "%sSummary:%s ", colorBold, colorReset)
	if result.Added > 0 {
		fmt.Fprintf(&b, "%s%d added%s, ", colorGreen, result.Added, colorReset)
	}
	if result.Removed > 0 {
		fmt.Fprintf(&b, "%s%d removed%s, ", colorRed, result.Removed, colorReset)
	}
	if result.Modified > 0 {
		fmt.Fprintf(&b, "%s%d modified%s, ", colorYellow, result.Modified, colorReset)
	}
	fmt.Fprintf(&b, "%d unchanged\n", result.Unchanged)

	return b.String()
}

// shortHash returns the first 7 characters of a SHA, or the full SHA if shorter.
func shortHash(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}

// formatSubmoduleChange returns the formatted line and color for a submodule change.
func formatSubmoduleChange(ch diff.SubmoduleChange) (string, string) {
	switch ch.Status {
	case "added":
		return fmt.Sprintf("  %s:  (added)   →  %s", ch.Path, shortHash(ch.NewSHA)), colorGreen
	case "removed":
		return fmt.Sprintf("  %s:  %s  →  (removed)", ch.Path, shortHash(ch.OldSHA)), colorRed
	case "updated":
		return fmt.Sprintf("  %s:  %s  →  %s  (updated)", ch.Path, shortHash(ch.OldSHA), shortHash(ch.NewSHA)), colorYellow
	default:
		return fmt.Sprintf("  %s:  %s  →  %s  (%s)", ch.Path, shortHash(ch.OldSHA), shortHash(ch.NewSHA), ch.Status), colorYellow
	}
}

// renderSubmoduleSection renders the submodule changes section.
// Returns an empty string if there are no submodule changes.
func renderSubmoduleSection(submodules []diff.SubmoduleChange) string {
	if len(submodules) == 0 {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s--- Submodule Changes ---%s\n", colorCyan, colorReset)
	for _, ch := range submodules {
		line, color := formatSubmoduleChange(ch)
		fmt.Fprintf(&b, "%s%s%s\n", color, line, colorReset)
	}

	return b.String()
}

// FormatDiffJSON produces JSON output.
func FormatDiffJSON(result *diff.Result) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// formatValue renders an attribute value for display.
func formatValue(v any) string {
	if v == nil {
		return "<nil>"
	}
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("%q", val)
	case []any:
		parts := make([]string, len(val))
		for i, item := range val {
			parts[i] = fmt.Sprintf("%v", item)
		}
		return fmt.Sprintf("[%s]", strings.Join(parts, ", "))
	default:
		return fmt.Sprintf("%v", val)
	}
}
