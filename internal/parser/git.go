package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"reqmd/internal/model"
)

// DiscoverAtTagWithSchemas extracts the repo at the given git tag into a temp
// directory, runs the standard Discover pipeline, and returns the parsed
// documents plus the raw schema.yaml content for each document directory.
func DiscoverAtTagWithSchemas(root, tag string) ([]model.Document, map[string]map[string]any, error) {
	return discoverAtTag(root, tag)
}

// discoverAtTag extracts the repo at the given git tag and returns both the
// parsed documents and the raw schema maps per document path.
func discoverAtTag(root, tag string) ([]model.Document, map[string]map[string]any, error) {
	// Validate tag exists.
	verifyCmd := exec.Command("git", "-C", root, "rev-parse", "--verify", tag)
	if out, err := verifyCmd.CombinedOutput(); err != nil {
		return nil, nil, fmt.Errorf("tag %q not found: %w: %s", tag, err, strings.TrimSpace(string(out)))
	}

	// Create temp dir.
	tmpDir, err := os.MkdirTemp("", "reqmd-baseline-*")
	if err != nil {
		return nil, nil, fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Extract repo at tag: git archive tag | tar -xf - -C tmpDir.
	archive := exec.Command("git", "-C", root, "archive", tag)
	tar := exec.Command("tar", "-xf", "-", "-C", tmpDir)

	tar.Stdin, err = archive.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("piping git archive: %w", err)
	}

	if err := tar.Start(); err != nil {
		return nil, nil, fmt.Errorf("starting tar: %w", err)
	}

	// Start the git archive process (its stdout is piped to tar)
	if err := archive.Start(); err != nil {
		tar.Wait() //nolint:errcheck
		return nil, nil, fmt.Errorf("starting git archive: %w", err)
	}

	// Wait for both processes
	if err := archive.Wait(); err != nil {
		tar.Wait() //nolint:errcheck
		return nil, nil, fmt.Errorf("git archive %q failed: %w", tag, err)
	}
	if err := tar.Wait(); err != nil {
		return nil, nil, fmt.Errorf("tar extraction failed: %w", err)
	}

	// Discover docs in extracted tree.
	docs, err := Discover(tmpDir)
	if err != nil {
		return nil, nil, fmt.Errorf("discovering docs at tag %q: %w", tag, err)
	}

	// Rewrite paths to be relative to the repo root. git archive stores
	// files with paths relative to the repo root, so stripping the temp
	// dir prefix yields the original repo-relative path.
	for i := range docs {
		docs[i].Path = relToTmpDir(tmpDir, docs[i].Path)
		for j := range docs[i].Requirements {
			docs[i].Requirements[j].Source = relToTmpDir(tmpDir, docs[i].Requirements[j].Source)
		}
	}

	// Load raw schemas for each discovered document.
	schemas := make(map[string]map[string]any, len(docs))
	for _, doc := range docs {
		schema, err := LoadSchema(filepath.Join(tmpDir, doc.Path))
		if err != nil {
			return nil, nil, fmt.Errorf("loading schema for %s: %w", doc.Path, err)
		}
		schemas[doc.Path] = schema
	}

	return docs, schemas, nil
}

// relToTmpDir returns path relative to the temp extraction directory.
func relToTmpDir(tmpDir, path string) string {
	if path == "" {
		return path
	}
	rel, err := filepath.Rel(tmpDir, path)
	if err != nil {
		return strings.TrimPrefix(path, tmpDir+string(filepath.Separator))
	}
	return rel
}

// ListSubmodulesAtTag returns a map of submodule path → pinned commit SHA
// at the given git tag. Uses `git ls-tree -t <tag>` and filters for
// submodule entries (mode 160000, type commit). Returns an empty map
// if the tag has no submodules. Returns an error if the tag is invalid
// or the git command fails.
func ListSubmodulesAtTag(root, tag string) (map[string]string, error) {
	// 1. Validate tag exists
	//    git -C root rev-parse --verify <tag>^{commit}
	if err := exec.Command("git", "-C", root, "rev-parse", "--verify", tag+"^{commit}").Run(); err != nil {
		return nil, fmt.Errorf("tag %q not found", tag)
	}

	// 2. Run git ls-tree -t <tag>
	//    Output format per line (space-separated, 4 fields):
	//      <mode> <type> <sha>\t<path>
	//    For submodules: mode=160000, type=commit
	out, err := exec.Command("git", "-C", root, "ls-tree", "-t", tag).Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-tree %s: %w", tag, err)
	}

	return parseLsTreeOutput(string(out)), nil
}

// parseLsTreeOutput parses the output of `git ls-tree -t <tag>` and
// returns only submodule entries (mode 160000, type commit) as
// path → SHA. Exposed as a separate function for unit testability.
//
// git ls-tree -t output format (one entry per line):
//
//	<mode> <type> <object>\t<file>
//
// Example:
//
//	160000 commit abc123def456...\tvendor/spec-a
//	100644 blob   789ghi012jkl...\tREADME.md
func parseLsTreeOutput(output string) map[string]string {
	result := make(map[string]string)
	if output == "" {
		return result
	}

	scanner := bufio.NewScanner(strings.NewReader(output))
	// Increase buffer size for very long paths (default 64KB may not be enough)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		// Find tab separator between metadata and path
		tabIdx := bytes.IndexByte(line, '\t')
		if tabIdx < 0 {
			continue
		}
		meta := string(line[:tabIdx])
		path := string(line[tabIdx+1:])

		// Parse mode type sha
		var mode, typeStr, sha string
		if _, err := fmt.Sscanf(meta, "%s %s %s", &mode, &typeStr, &sha); err != nil {
			continue
		}

		// Filter: only submodules (mode 160000, type commit)
		if mode == "160000" && typeStr == "commit" {
			result[path] = sha
		}
	}
	return result
}
