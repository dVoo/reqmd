package parser

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"

	"github.com/FurqanSoftware/goldmark-katex"
	"github.com/stefanfritsch/goldmark-fences"
	"github.com/yuin/goldmark-emoji"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark-meta"
	"go.abhg.dev/goldmark/mermaid"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"

	"reqmd/internal/model"
	"reqmd/internal/schema"
)

const attrFenceInfo = "attr"

// mdParser is the shared goldmark instance used by parseMD.
// Constructed once at package init to avoid re-running extension Init for every .md file.
var mdParser = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		&mermaid.Extender{},
		&fences.Extender{},
		highlighting.Highlighting,
		emoji.Emoji,
		meta.Meta,
		&katex.Extender{},
	),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
)

// Discover walks root recursively, locating schema.yaml files.
// Every directory containing schema.yaml is a document directory.
// Returns one Document per document directory.
// Document directories are loaded in parallel for performance.
func Discover(root string) ([]model.Document, error) {
	docDirs, err := findDocDirs(root)
	if err != nil {
		return nil, err
	}

	docs := make([]model.Document, len(docDirs))
	g := new(errgroup.Group)
	for i, dir := range docDirs {
		i, dir := i, dir
		g.Go(func() error {
			doc, err := loadDocument(dir)
			if err != nil {
				return fmt.Errorf("loading %s: %w", dir, err)
			}
			docs[i] = doc
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return docs, nil
}

// findDocDirs walks root and collects all directories containing schema.yaml.
func findDocDirs(root string) ([]string, error) {
	var docDirs []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Name() == "schema.yaml" {
			docDirs = append(docDirs, filepath.Dir(path))
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking %s: %w", root, err)
	}
	if len(docDirs) == 0 {
		return nil, fmt.Errorf("no schema.yaml found under %s", root)
	}
	return docDirs, nil
}

// LoadSchema reads and parses the schema.yaml file from a document directory.
// It returns the raw schema map before any x-reqmd extraction.
func LoadSchema(docDir string) (map[string]any, error) {
	schemaRaw, err := os.ReadFile(filepath.Join(docDir, "schema.yaml"))
	if err != nil {
		return nil, fmt.Errorf("reading schema: %w", err)
	}
	var schema map[string]any
	if err := yaml.Unmarshal(schemaRaw, &schema); err != nil {
		return nil, fmt.Errorf("parsing schema: %w", err)
	}
	return schema, nil
}

func loadDocument(dir string) (model.Document, error) {
	schemaData, err := LoadSchema(dir)
	if err != nil {
		return model.Document{}, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return model.Document{}, err
	}
	var mdFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			mdFiles = append(mdFiles, filepath.Join(dir, e.Name()))
		}
	}

	reqs, frontmatter, errs := parseFiles(mdFiles)
	if len(errs) > 0 {
		return model.Document{}, fmt.Errorf("parse errors: %v", errs)
	}

	xReqmd := schema.ExtractXReqmd(schemaData)

	return model.Document{
		Path:         dir,
		Schema:       schemaData,
		Requirements: reqs,
		XReqmd:       xReqmd,
		Properties:   schema.ExtractProperties(schemaData),
		Meta:         frontmatter,
	}, nil
}

func parseFiles(files []string) ([]model.Requirement, map[string]any, []error) {
	if len(files) == 0 {
		return nil, nil, nil
	}

	type result struct {
		reqs []model.Requirement
		meta map[string]any
		err  error
	}
	results := make(chan result, len(files))
	sem := make(chan struct{}, runtime.NumCPU())
	var wg sync.WaitGroup

	for _, f := range files {
		f := f
		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			src, err := os.ReadFile(f)
			if err != nil {
				results <- result{err: fmt.Errorf("reading %s: %w", f, err)}
				return
			}
			reqs, frontmatter, err := parseMD(src, f)
			if err != nil {
				results <- result{err: fmt.Errorf("parsing %s: %w", f, err)}
				return
			}
			results <- result{reqs: reqs, meta: frontmatter}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var allReqs []model.Requirement
	var errs []error
	mergedMeta := make(map[string]any)
	for r := range results {
		if r.err != nil {
			errs = append(errs, r.err)
		} else {
			allReqs = append(allReqs, r.reqs...)
			// Merge frontmatter in order (first file's value wins per key)
			for k, v := range r.meta {
				if _, exists := mergedMeta[k]; !exists {
					mergedMeta[k] = v
				}
			}
		}
	}
	return allReqs, mergedMeta, errs
}

// reqStackEntry tracks an open requirement at a heading level.
type reqStackEntry struct {
	level int
	id    string
}

// popStackToLevel removes entries whose level is >= targetLevel.
// Used by both requirement and container headings to close requirements
// at their level or deeper.
func popStackToLevel(stack []reqStackEntry, targetLevel int) []reqStackEntry {
	for len(stack) > 0 && stack[len(stack)-1].level >= targetLevel {
		stack = stack[:len(stack)-1]
	}
	return stack
}

// parseMD uses goldmark to extract requirements from a single .md file
// and returns any YAML frontmatter as a map.
//
// For files with multiple root-level requirements (parent=nil), the file
// is split at root requirement boundaries and chunks are parsed in
// parallel. Each chunk is self-contained: a root requirement and all its
// descendants. The pre-scan to find split points is a cheap line scan
// that replicates the stack logic without building a goldmark AST.
func parseMD(src []byte, sourcePath string) ([]model.Requirement, map[string]any, error) {
	splits, frontmatter := findRootSplits(src)

	// No splits or a single root → parse the whole file serially.
	nRoots := len(splits)
	if nRoots <= 1 {
		return parseMDChunk(src, sourcePath)
	}

	// Batch root requirements into ~NumCPU chunks to amortize goldmark
	// init overhead (~490ns per parse) while still getting parallelism.
	// Each chunk contains multiple consecutive root requirements + their
	// descendants. Chunks are balanced by root count, not byte size.
	nCPU := runtime.NumCPU()
	nChunks := nRoots
	if nChunks > nCPU {
		nChunks = nCPU
	}

	// Compute chunk boundaries: which split indices start/end each chunk.
	type chunkResult struct {
		reqs []model.Requirement
		err  error
	}
	results := make([]chunkResult, nChunks)
	var wg sync.WaitGroup

	for c := range nChunks {
		// Root index range [rootStart, rootEnd) for this chunk.
		rootStart := c * nRoots / nChunks
		rootEnd := (c + 1) * nRoots / nChunks
		if rootEnd > nRoots {
			rootEnd = nRoots
		}
		// Byte range: from the start of the first root in this chunk
		// to the start of the first root in the next chunk (or EOF).
		byteStart := splits[rootStart]
		byteEnd := len(src)
		if rootEnd < nRoots {
			byteEnd = splits[rootEnd]
		}
		chunk := src[byteStart:byteEnd]
		idx := c
		wg.Add(1)
		go func() {
			defer wg.Done()
			reqs, _, err := parseMDChunk(chunk, sourcePath)
			results[idx] = chunkResult{reqs: reqs, err: err}
		}()
	}
	wg.Wait()

	var reqs []model.Requirement
	for _, r := range results {
		if r.err != nil {
			return nil, frontmatter, r.err
		}
		reqs = append(reqs, r.reqs...)
	}
	return reqs, frontmatter, nil
}

// parseMDChunk parses a single chunk of markdown (possibly the whole file)
// using goldmark and returns requirements and frontmatter.
func parseMDChunk(src []byte, sourcePath string) ([]model.Requirement, map[string]any, error) {
	p := mdParser.Parser()
	parseCtx := parser.NewContext()
	doc := p.Parse(text.NewReader(src), parser.WithContext(parseCtx))

	// Extract YAML frontmatter (parsed by meta.Meta, stored in context).
	var frontmatter map[string]any
	if fm := meta.Get(parseCtx); fm != nil {
		frontmatter = fm
	}

	var reqs []model.Requirement
	var cur *model.Requirement
	var waitingForAttr bool
	var bodyBuf strings.Builder
	var stack []reqStackEntry // open requirements by heading level
	for n := doc.FirstChild(); n != nil; n = n.NextSibling() {
		switch v := n.(type) {
		case *ast.Heading:
			level := v.Level
			hasAdjacentAttr := isNextAttrBlock(n, src)

			if hasAdjacentAttr {
				// ── Requirement heading ──
				if cur != nil {
					if cur.Attrs == nil {
						return nil, frontmatter, fmt.Errorf("%s: req %s: missing attr block", sourcePath, cur.ID)
					}
					cur.Body = bodyBuf.String()
					reqs = append(reqs, *cur)
					cur = nil
				}
				stack = popStackToLevel(stack, level)
				parentID := ""
				if len(stack) > 0 {
					parentID = stack[len(stack)-1].id
				}
				heading := collectInlineText(v, src)
				id, title := splitHeadingID(heading)
				cur = &model.Requirement{
					ID:       id,
					Title:    title,
					Source:   sourcePath,
					ParentID: parentID,
				}
				bodyBuf.Reset()
				stack = append(stack, reqStackEntry{level: level, id: id})
				waitingForAttr = true
			} else {
				// ── Container heading (no attr) ──
				// Pop requirements at this level or deeper, breaking the
				// parent chain. Deeper headings (level > stack top) are
				// body text and don't pop anything.
				stack = popStackToLevel(stack, level)
				if cur != nil {
					bodyBuf.WriteString(renderHeading(v, src))
				}
			}

		case *ast.FencedCodeBlock:
			if cur != nil && waitingForAttr && fenceInfo(v, src) == attrFenceInfo {
				yamlText := fenceLines(v, src)
				attrs, err := decodeAttrYAML(yamlText)
				if err != nil {
					// Fallback to yaml.v3 for complex syntax the scanner
					// doesn't handle (nested maps, flow sequences, etc.).
					attrs = make(map[string]any, 8)
					if yamlErr := yaml.Unmarshal([]byte(yamlText), &attrs); yamlErr != nil {
						return nil, frontmatter, fmt.Errorf("%s: req %s: invalid attr YAML: %w", sourcePath, cur.ID, yamlErr)
					}
				}
				cur.Attrs = attrs
				// Extract reqmd-suppress before it reaches schema validation
				if suppressRaw, ok := attrs["reqmd-suppress"]; ok {
					if suppressList, ok := suppressRaw.([]any); ok {
						for _, item := range suppressList {
							if s, ok := item.(string); ok {
								cur.Suppressions = append(cur.Suppressions, s)
							}
						}
					}
					delete(attrs, "reqmd-suppress")
				}
				waitingForAttr = false
			} else if cur != nil {
				// Non-attr code block inside requirement body
				bodyBuf.WriteString(renderFence(v, src))
			}

		case *ast.Paragraph:
			if cur != nil {
				text := paragraphText(v, src)
				if text == "" {
					continue
				}
				if strings.HasPrefix(text, "*Rationale:") {
					// Handle both *Rationale: and *Rationale:* (markdown italic)
					trimmed := strings.TrimPrefix(text, "*Rationale:")
					trimmed = strings.TrimLeft(trimmed, "* ")
					cur.Rationale = strings.TrimSpace(trimmed)
				} else {
					if bodyBuf.Len() > 0 {
						bodyBuf.WriteString("\n\n")
					}
					bodyBuf.WriteString(text)
				}
			}

		case *ast.ThematicBreak:
			// ignore separators

		default:
			// Any other block (blockquote, list, etc.) → body text if inside a requirement
			if cur != nil {
				bodyBuf.WriteString(renderBlock(src, n))
			}
		}
	}

	// Finalize last requirement
	if cur != nil {
		if cur.Attrs == nil {
			return nil, frontmatter, fmt.Errorf("%s: req %s: missing attr block", sourcePath, cur.ID)
		}
		cur.Body = bodyBuf.String()
		reqs = append(reqs, *cur)
	}

	return reqs, frontmatter, nil
}

// findRootSplits scans src line-by-line to find byte offsets where root-level
// requirements begin. A root requirement is one where the stack is empty when
// the requirement heading is encountered. The scan replicates the same stack
// logic as parseMDChunk but without building a goldmark AST.
//
// Returns the byte offsets of root requirement starts (including the heading
// line) and any YAML frontmatter parsed from the file header.
//
// If only 0 or 1 roots are found, the caller parses the whole file serially.
func findRootSplits(src []byte) (splits []int, frontmatter map[string]any) {
	lines := splitLines(src)

	// Parse frontmatter if present (--- delimited block at file start).
	lineIdx := 0
	if len(lines) > 0 && strings.TrimSpace(string(lines[0])) == "---" {
		for lineIdx = 1; lineIdx < len(lines); lineIdx++ {
			if strings.TrimSpace(string(lines[lineIdx])) == "---" {
				if lineIdx > 1 {
					fmText := string(bytes.Join(lines[1:lineIdx], []byte("\n")))
					_ = yaml.Unmarshal([]byte(fmText), &frontmatter)
				}
				lineIdx++
				break
			}
		}
	}

	// Track byte offset of the current line.
	byteOffset := 0
	for i := 0; i < lineIdx; i++ {
		byteOffset += len(lines[i]) + 1 // +1 for the \n
	}

	var stack []reqStackEntry

	for lineIdx < len(lines) {
		line := string(lines[lineIdx])
		// Detect heading lines: 1-6 '#' followed by a space.
		if level, ok := headingLevel(line); ok {
			hasAttr := nextNonBlankIsAttr(lines, lineIdx+1)

			if hasAttr {
				// Requirement heading: pop to this level, check if root.
				stack = popStackToLevel(stack, level)
				if len(stack) == 0 {
					// Root requirement — record split point.
					splits = append(splits, byteOffset)
				}
				// Extract ID from heading text (same as splitHeadingID).
				headingText := strings.TrimSpace(line[level:])
				id, _ := splitHeadingID(headingText)
				stack = append(stack, reqStackEntry{level: level, id: id})
			} else {
				// Container heading: pop to this level (breaks chain).
				stack = popStackToLevel(stack, level)
			}
		}

		byteOffset += len(lines[lineIdx]) + 1
		lineIdx++
	}

	return splits, frontmatter
}

// splitLines splits src on \n, returning each line without the trailing newline.
// A trailing empty line (from a final \n) is not included.
func splitLines(src []byte) [][]byte {
	if len(src) == 0 {
		return nil
	}
	var lines [][]byte
	start := 0
	for i, b := range src {
		if b == '\n' {
			lines = append(lines, src[start:i])
			start = i + 1
		}
	}
	if start < len(src) {
		lines = append(lines, src[start:])
	}
	return lines
}

// headingLevel returns the heading level (1-6) and true if the line is a
// markdown heading (e.g. "## Title"). Returns 0, false otherwise.
func headingLevel(line string) (int, bool) {
	if len(line) == 0 || line[0] != '#' {
		return 0, false
	}
	level := 0
	for level < 6 && level < len(line) && line[level] == '#' {
		level++
	}
	if level < len(line) && line[level] == ' ' {
		return level, true
	}
	return 0, false
}

// nextNonBlankIsAttr checks if the next non-blank line after lineIdx is a
// ```attr fence. This mirrors isNextAttrBlock in the goldmark path.
func nextNonBlankIsAttr(lines [][]byte, idx int) bool {
	for idx < len(lines) {
		trimmed := strings.TrimSpace(string(lines[idx]))
		if trimmed == "" {
			idx++
			continue
		}
		return strings.HasPrefix(trimmed, "```attr")
	}
	return false
}

// isNextAttrBlock returns true if n's next sibling is a ```attr fence.
// Blank lines between heading and fence are consumed by goldmark and don't
// create AST nodes, so they don't affect this check.
func isNextAttrBlock(n ast.Node, src []byte) bool {
	next := n.NextSibling()
	if next == nil {
		return false
	}
	fence, ok := next.(*ast.FencedCodeBlock)
	if !ok {
		return false
	}
	return fenceInfo(fence, src) == attrFenceInfo
}

// splitHeadingID splits a heading like "REQ-001: This is a title"
// into ("REQ-001", "This is a title"). If there's no colon+space,
// returns (heading, "").
func splitHeadingID(heading string) (id string, title string) {
	if idx := strings.Index(heading, ": "); idx >= 0 {
		return strings.TrimSpace(heading[:idx]), strings.TrimSpace(heading[idx+2:])
	}
	return strings.TrimSpace(heading), ""
}

// --- text extraction helpers ---

func collectInlineText(n ast.Node, src []byte) string {
	var b strings.Builder
	ast.Walk(n, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if tn, ok := n.(*ast.Text); ok {
			b.Write(tn.Segment.Value(src))
		}
		return ast.WalkContinue, nil
	})
	return strings.TrimSpace(b.String())
}

func fenceInfo(f *ast.FencedCodeBlock, src []byte) string {
	if f.Info == nil {
		return ""
	}
	infoSeg := f.Info
	return strings.TrimSpace(string(infoSeg.Value(src)))
}

func fenceLines(f *ast.FencedCodeBlock, src []byte) string {
	var b strings.Builder
	for i := 0; i < f.Lines().Len(); i++ {
		seg := f.Lines().At(i)
		b.Write(seg.Value(src))
	}
	return b.String()
}

func paragraphText(p *ast.Paragraph, src []byte) string {
	var b strings.Builder
	for i := 0; i < p.Lines().Len(); i++ {
		seg := p.Lines().At(i)
		b.WriteString(strings.TrimRight(string(seg.Value(src)), "\n\r"))
	}
	return strings.TrimSpace(b.String())
}

func renderHeading(h *ast.Heading, src []byte) string {
	var b strings.Builder
	b.WriteString(strings.Repeat("#", h.Level))
	b.WriteString(" ")
	b.WriteString(collectInlineText(h, src))
	b.WriteString("\n\n")
	return b.String()
}

func renderFence(f *ast.FencedCodeBlock, src []byte) string {
	var b strings.Builder
	b.WriteString("```")
	b.WriteString(fenceInfo(f, src))
	b.WriteString("\n")
	b.WriteString(fenceLines(f, src))
	b.WriteString("```\n\n")
	return b.String()
}

func renderBlock(src []byte, n ast.Node) string {
	var b strings.Builder
	// Collect all text content from unrecognized block elements
	ast.Walk(n, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		// Only block nodes have Lines() — inline nodes (e.g., italic text)
		// panic when Lines() is called.
		if n.Type() == ast.TypeInline {
			return ast.WalkContinue, nil
		}
		for i := 0; i < n.Lines().Len(); i++ {
			seg := n.Lines().At(i)
			b.Write(seg.Value(src))
		}
		return ast.WalkContinue, nil
	})
	return b.String()
}

// ParseSingleFile is used by `reqmd check <file> --schema <schema>`.
func ParseSingleFile(path string) ([]model.Requirement, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	reqs, _, err := parseMD(src, path)
	return reqs, err
}
