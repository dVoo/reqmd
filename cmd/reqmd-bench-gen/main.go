// Synthetic reqmd benchmark generator.
//
// Reads a source spec tree (e.g. /home/daniel/projects/mdreq/spec/) and
// emits N copies under <dst>/. Each copy is renamed to have globally
// unique IDs (with a per-copy id-prefix) and rewritten trace references.
//
// Flags:
//
//	-copies N       emit N copies of the source (each is a separate root)
//	-reqsperdoc K  replicate per-doc content to reach ~K reqs per doc
//
// When -reqsperdoc is set, each .md file in each copy is appended with
// (K / source_per_doc) copies of the file's own requirement blocks, each
// renamed to keep IDs unique. This produces a corpus with a target doc
// count (6 × copies) AND a target per-doc req count, separately.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	headingIDRe  = regexp.MustCompile(`(?m)^##\s+([A-Z][A-Z0-9-]+)(\b|:)`)
	idPrefixRe   = regexp.MustCompile(`(?m)^(\s*id-prefix:\s*)([A-Z][A-Z0-9-]+)\s*$`)
	qualifiedRe  = regexp.MustCompile(`\b([A-Za-z][A-Za-z0-9_-]+)/([A-Z][A-Z0-9-]+)(~\d+)?\b`)
	plainIDRe    = regexp.MustCompile(`\b([A-Z][A-Z0-9-]+)(~\d+)?\b`)
	traceFieldRe = regexp.MustCompile(`^(\s*-?\s*(?:trace|requires-trace-from):\s*\[?\s*)(.*?)(\s*\]?\s*)$`)
)

const placeholderFmt = "ZZ_BENCH_PH_%d_ZZ"

// Generated corpora are private to the exporting user.
const (
	outputDirPerm  = 0o750
	outputFilePerm = 0o600
)

func main() {
	src := flag.String("src", "", "source spec dir")
	dst := flag.String("dst", "", "destination root")
	copies := flag.Int("copies", 5, "how many copies to emit under dst")
	reqsPerDoc := flag.Int("reqsperdoc", 0, "if >0, replicate per-doc content to reach this req count per doc")
	flag.Parse()

	if *src == "" || *dst == "" {
		fmt.Fprintln(os.Stderr, "usage: -src <dir> -dst <dir> [-copies N] [-reqsperdoc K]")
		os.Exit(2)
	}

	*src = filepath.Clean(*src)
	*dst = filepath.Clean(*dst)

	if err := os.MkdirAll(*dst, outputDirPerm); err != nil {
		panic(err)
	}

	// Step 1: collect every requirement ID in the source tree. Reads go
	// through an os.Root so a symlink cannot escape the source tree.
	srcRoot, err := os.OpenRoot(*src)
	if err != nil {
		panic(err)
	}
	srcFS := srcRoot.FS()

	ids := map[string]bool{}
	if err := fs.WalkDir(srcFS, ".", func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || (!strings.HasSuffix(p, ".md") && d.Name() != "schema.yaml") {
			return nil
		}
		body, readErr := fs.ReadFile(srcFS, p)
		if readErr != nil {
			return fmt.Errorf("reading %s: %w", p, readErr)
		}
		for _, m := range headingIDRe.FindAllSubmatch(body, -1) {
			ids[string(m[1])] = true
		}
		return nil
	}); err != nil {
		panic(err)
	}

	idList := make([]string, 0, len(ids))
	for id := range ids {
		idList = append(idList, id)
	}
	sort.Slice(idList, func(i, j int) bool { return len(idList[i]) > len(idList[j]) })

	// Count top-level doc dirs (those with a schema.yaml in the original).
	totalSrcReqs := len(ids)
	srcDocDirs := 0
	_ = fs.WalkDir(srcFS, ".", func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fs.SkipDir
		}
		if !d.IsDir() || p == "." {
			return nil
		}
		if path.Dir(p) != "." {
			return nil
		}
		if _, statErr := fs.Stat(srcFS, path.Join(p, "schema.yaml")); statErr == nil {
			srcDocDirs++
		}
		return nil
	})
	srcReqsPerDoc := 0
	if srcDocDirs > 0 {
		srcReqsPerDoc = totalSrcReqs / srcDocDirs
	}
	if closeErr := srcRoot.Close(); closeErr != nil {
		panic(closeErr)
	}

	// Step 2: emit copies.
	for i := range *copies {
		tag := fmt.Sprintf("T%02d", i)
		copyDir := filepath.Join(*dst, fmt.Sprintf("spec_%s", tag))
		if err := os.MkdirAll(copyDir, outputDirPerm); err != nil {
			panic(err)
		}
		rename := map[string]string{}
		for _, id := range idList {
			rename[id] = tag + "-" + id
		}
		rewriteTree(*src, copyDir, tag, rename)
		if *reqsPerDoc > 0 && srcReqsPerDoc > 0 {
			multiplier := max(1, (*reqsPerDoc)/srcReqsPerDoc)
			if multiplier > 1 {
				if repErr := replicatePerDoc(copyDir, multiplier); repErr != nil {
					fmt.Fprintf(os.Stderr, "replicatePerDoc: %v\n", repErr)
					os.Exit(1)
				}
			}
		}
		fmt.Printf("wrote %s (%d IDs rewritten", copyDir, len(rename))
		if *reqsPerDoc > 0 && srcReqsPerDoc > 0 {
			m := max(1, (*reqsPerDoc)/srcReqsPerDoc)
			fmt.Printf(", per-doc content x%d", m)
		}
		fmt.Println(")")
	}
}

func rewriteTree(src, dst, tag string, rename map[string]string) {
	keys := make([]string, 0, len(rename))
	for k := range rename {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return len(keys[i]) > len(keys[j]) })

	srcRoot, err := os.OpenRoot(src)
	if err != nil {
		panic(err)
	}
	defer func() { _ = srcRoot.Close() }()
	dstRoot, err := os.OpenRoot(dst)
	if err != nil {
		panic(err)
	}
	defer func() { _ = dstRoot.Close() }()
	srcFS := srcRoot.FS()

	if err := fs.WalkDir(srcFS, ".", func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if mkdirErr := dstRoot.MkdirAll(p, outputDirPerm); mkdirErr != nil {
				return fmt.Errorf("creating dir %s: %w", p, mkdirErr)
			}
			return nil
		}
		body, readErr := fs.ReadFile(srcFS, p)
		if readErr != nil {
			return fmt.Errorf("reading %s: %w", p, readErr)
		}
		out := applyRewrite(string(body), d.Name(), rename, tag)
		if writeErr := dstRoot.WriteFile(p, []byte(out), outputFilePerm); writeErr != nil {
			return fmt.Errorf("writing %s: %w", p, writeErr)
		}
		return nil
	}); err != nil {
		panic(err)
	}
}

// replicatePerDoc walks each top-level doc dir in `root` and, for every
// .md file, repeats the requirement blocks `multiplier - 1` more times.
// Each repeated block has its IDs and trace references rewritten to keep
// the trace graph connected.
func replicatePerDoc(root string, multiplier int) error {
	rootDir, err := os.OpenRoot(root)
	if err != nil {
		return fmt.Errorf("opening %s: %w", root, err)
	}
	defer func() { _ = rootDir.Close() }()
	fsys := rootDir.FS()

	err = fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() || p == "." {
			return nil
		}
		_, statErr := fs.Stat(fsys, path.Join(p, "schema.yaml"))
		switch {
		case errors.Is(statErr, fs.ErrNotExist):
			return nil // not a doc dir
		case statErr != nil:
			return fmt.Errorf("checking for schema.yaml in %s: %w", p, statErr)
		}
		entries, readDirErr := fs.ReadDir(fsys, p)
		if readDirErr != nil {
			return fmt.Errorf("reading dir %s: %w", p, readDirErr)
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			mdPath := path.Join(p, e.Name())
			b, readErr := fs.ReadFile(fsys, mdPath)
			if readErr != nil {
				fmt.Fprintf(os.Stderr, "warning: replicatePerDoc: reading %s: %v\n", mdPath, readErr)
				continue
			}
			expanded := expandContent(string(b), multiplier)
			if writeErr := rootDir.WriteFile(mdPath, []byte(expanded), outputFilePerm); writeErr != nil {
				fmt.Fprintf(os.Stderr, "warning: replicatePerDoc: writing %s: %v\n", mdPath, writeErr)
				continue
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("replicating %s: %w", root, err)
	}
	return nil
}

// expandContent extracts the frontmatter, the preamble, and the heading
// blocks from `body`, then concatenates `multiplier` copies of the heading
// blocks. The first copy uses the original IDs; subsequent copies have
// their IDs and references rewritten.
func expandContent(body string, multiplier int) string {
	frontmatter := ""
	rest := body
	if strings.HasPrefix(body, "---\n") {
		if idx := strings.Index(body[4:], "\n---\n"); idx >= 0 {
			frontmatter = body[:4+idx+5]
			rest = body[4+idx+5:]
		}
	}
	parts := splitOnH2(rest)
	if len(parts) < 2 {
		return body
	}
	preamble := parts[0]
	headingBlocks := parts[1:]

	var b strings.Builder
	b.WriteString(frontmatter)
	b.WriteString(preamble)
	b.WriteString(strings.Join(headingBlocks, ""))
	for n := 1; n < multiplier; n++ {
		for _, blk := range headingBlocks {
			b.WriteString(rewriteBlockForCopy(blk, n))
		}
	}
	return b.String()
}

// splitOnH2 splits body on lines starting with "## ". The first chunk
// includes the preamble (possibly empty); subsequent chunks start with
// "## " (preserved as their first line).
func splitOnH2(body string) []string {
	lines := strings.Split(body, "\n")
	var out []string
	var current []string
	for _, line := range lines {
		if strings.HasPrefix(line, "## ") && len(current) > 0 {
			out = append(out, strings.Join(current, "\n"))
			current = current[:0]
		}
		current = append(current, line)
	}
	if len(current) > 0 {
		out = append(out, strings.Join(current, "\n"))
	}
	return out
}

// rewriteBlockForCopy returns a copy of `block` with the heading ID
// rewritten to have a "_c<copyIdx>" suffix, and the `trace:` and
// `requires-trace-from:` lines rewritten to point at the renamed IDs.
// Other fields are left untouched (we don't want to corrupt enum values
// like `process: SYS` or `status: approved`).
func rewriteBlockForCopy(block string, copyIdx int) string {
	suffix := fmt.Sprintf("_c%d", copyIdx)
	placeholders := map[string]string{}
	pcount := 0
	makePh := func() string {
		pcount++
		return fmt.Sprintf(placeholderFmt, pcount)
	}

	body := headingIDRe.ReplaceAllStringFunc(block, func(s string) string {
		loc := headingIDRe.FindStringSubmatchIndex(s)
		if loc == nil {
			return s
		}
		id := s[loc[2]:loc[3]]
		rest := s[loc[4]:]
		ph := makePh()
		placeholders[ph] = "## " + id + suffix + rest
		return ph
	})

	// Rewrite trace:/requires-trace-from: lines only.
	lines := strings.Split(body, "\n")
	for li, line := range lines {
		m := traceFieldRe.FindStringSubmatchIndex(line)
		if m == nil {
			continue
		}
		pre := line[m[2]:m[3]]
		list := line[m[4]:m[5]]
		suf := line[m[6]:m[7]]
		list = rewriteIDListInPlace(list, suffix, placeholders, makePh)
		lines[li] = pre + list + suf
	}
	body = strings.Join(lines, "\n")

	phs := make([]string, 0, len(placeholders))
	for ph := range placeholders {
		phs = append(phs, ph)
	}
	sort.Slice(phs, func(i, j int) bool { return len(phs[i]) > len(phs[j]) })
	for _, ph := range phs {
		body = strings.ReplaceAll(body, ph, placeholders[ph])
	}
	return body
}

// rewriteIDListInPlace rewrites each ID token in a comma-separated list
// like "SYS-001, stakeholder/SYS-001~3" by appending `suffix` to the ID
// portion.
func rewriteIDListInPlace(list, suffix string, placeholders map[string]string, makePh func() string) string {
	list = qualifiedRe.ReplaceAllStringFunc(list, func(s string) string {
		loc := qualifiedRe.FindStringSubmatch(s)
		if loc == nil {
			return s
		}
		doc, id, pin := loc[1], loc[2], loc[3]
		ph := makePh()
		placeholders[ph] = doc + "/" + id + suffix + pin
		return ph
	})
	list = plainIDRe.ReplaceAllStringFunc(list, func(s string) string {
		loc := plainIDRe.FindStringSubmatch(s)
		if loc == nil {
			return s
		}
		id, pin := loc[1], loc[2]
		ph := makePh()
		placeholders[ph] = id + suffix + pin
		return ph
	})
	return list
}

// applyRewrite is the per-copy rename pass. Same as rewriteBlockForCopy
// in spirit but operates on the whole file body and uses the per-copy
// rename map (e.g. "SYS-001" -> "T00-SYS-001").
func applyRewrite(body, filename string, rename map[string]string, tag string) string {
	placeholders := map[string]string{}
	pcount := 0
	makePlaceholder := func() string {
		pcount++
		return fmt.Sprintf(placeholderFmt, pcount)
	}

	if strings.HasSuffix(filename, ".md") {
		body = headingIDRe.ReplaceAllStringFunc(body, func(s string) string {
			loc := headingIDRe.FindStringSubmatchIndex(s)
			if loc == nil {
				return s
			}
			id := s[loc[2]:loc[3]]
			rest := s[loc[4]:]
			if newID, ok := rename[id]; ok {
				ph := makePlaceholder()
				placeholders[ph] = "## " + newID + rest
				return ph
			}
			return s
		})
	}

	body = qualifiedRe.ReplaceAllStringFunc(body, func(s string) string {
		loc := qualifiedRe.FindStringSubmatch(s)
		if loc == nil {
			return s
		}
		doc, id, pin := loc[1], loc[2], loc[3]
		if newID, ok := rename[id]; ok {
			ph := makePlaceholder()
			placeholders[ph] = doc + "/" + newID + pin
			return ph
		}
		return s
	})

	body = plainIDRe.ReplaceAllStringFunc(body, func(s string) string {
		loc := plainIDRe.FindStringSubmatch(s)
		if loc == nil {
			return s
		}
		id, pin := loc[1], loc[2]
		if newID, ok := rename[id]; ok {
			ph := makePlaceholder()
			placeholders[ph] = newID + pin
			return ph
		}
		return s
	})

	phs := make([]string, 0, len(placeholders))
	for ph := range placeholders {
		phs = append(phs, ph)
	}
	sort.Slice(phs, func(i, j int) bool { return len(phs[i]) > len(phs[j]) })
	for _, ph := range phs {
		body = strings.ReplaceAll(body, ph, placeholders[ph])
	}

	if filename == "schema.yaml" {
		body = idPrefixRe.ReplaceAllStringFunc(body, func(s string) string {
			loc := idPrefixRe.FindStringSubmatchIndex(s)
			if loc == nil {
				return s
			}
			prefix := s[loc[4]:loc[5]]
			return s[:loc[4]] + tag + "-" + prefix + s[loc[5]:]
		})
	}
	return body
}
