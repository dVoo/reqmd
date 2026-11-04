// Package parser discovers reqmd document directories, parses the
// Markdown content tree (requirements, containers, and info items), and
// loads document trees from git tags.
package parser

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/dVoo/reqmd/internal/model"
	"github.com/dVoo/reqmd/internal/schema"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/extension"
	"github.com/yuin/goldmark/v2/parser"
	meta "github.com/yuin/goldmark-meta/v2"

	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

const attrFenceInfo = "attr"

// Package-level patterns reused in the hot per-line scan helpers. Hoisted
// so the hot paths don't allocate a []byte pattern on every call.
var (
	attrFenceBytes = []byte("```attr")
)

// mdParser is the shared goldmark parser used by parseMD. It only parses: the
// content tree is built by walking the AST, so no renderer is attached.
//
// Structural extensions are the only ones registered here — tables and task
// lists must become real block nodes or their markers would be flattened into
// paragraph text — plus the YAML frontmatter parser, which stores the document's
// metadata on the AST. Purely presentational extensions (emoji, syntax
// highlighting, alerts, diagrams) live on the exporter's renderer instead: the
// body is stored as Markdown, so they apply when that Markdown is rendered.
var mdParser = parser.New(parser.WithExtensions(
	extension.NewGFMParser(),
	meta.NewParser(),
))

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
		return nil, fmt.Errorf("loading documents: %w", err)
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
		return model.Document{}, fmt.Errorf("reading %s: %w", dir, err)
	}
	var mdFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			mdFiles = append(mdFiles, filepath.Join(dir, e.Name()))
		}
	}

	nodes, title, frontmatter, errs := parseFiles(mdFiles)
	if len(errs) > 0 {
		return model.Document{}, fmt.Errorf("parse errors: %v", errs)
	}

	xReqmd := schema.ExtractXReqmd(schemaData)

	return model.Document{
		Path:       dir,
		Schema:     schemaData,
		Title:      title,
		Nodes:      nodes,
		XReqmd:     xReqmd,
		Properties: schema.ExtractProperties(schemaData),
		Meta:       frontmatter,
	}, nil
}

// parseFiles parses all .md files in a document directory and returns the
// concatenated content tree roots in mdFiles order (deterministic), the
// document title (the first file's h1 heading, if any), and any per-file
// errors. Files are parsed in parallel;
// results are collected by file index so the returned order never depends
// on goroutine completion order.
func parseFiles(files []string) ([]*model.Node, string, map[string]any, []error) {
	if len(files) == 0 {
		return nil, "", nil, nil
	}

	type result struct {
		err   error
		meta  map[string]any
		title string
		nodes []*model.Node
	}
	results := make([]result, len(files))
	sem := make(chan struct{}, runtime.NumCPU())
	var wg sync.WaitGroup

	for i, f := range files {
		sem <- struct{}{}
		wg.Go(func() {
			defer func() { <-sem }()
			src, err := os.ReadFile(f)
			if err != nil {
				results[i] = result{err: fmt.Errorf("reading %s: %w", f, err)}
				return
			}
			nodes, title, fm, err := parseMD(src, f)
			if err != nil {
				results[i] = result{err: fmt.Errorf("parsing %s: %w", f, err)}
				return
			}
			results[i] = result{nodes: nodes, title: title, meta: fm}
		})
	}

	wg.Wait()

	var all []*model.Node
	var errs []error
	mergedMeta := make(map[string]any)
	title := ""
	for i, r := range results {
		if r.err != nil {
			errs = append(errs, r.err)
			continue
		}
		all = append(all, r.nodes...)
		// Document title = the first file's h1 heading (file order).
		if i == 0 && title == "" {
			title = r.title
		}
		// Merge frontmatter in file order; the first file's value wins per key.
		for k, v := range r.meta {
			if _, exists := mergedMeta[k]; !exists {
				mergedMeta[k] = v
			}
		}
	}
	return all, title, mergedMeta, errs
}

// parseParallelThreshold is the minimum file size (in bytes) for which the
// root-split pre-scan + intra-file chunk parallelism is worth it. Below
// this size the file is parsed serially: the pre-scan is an extra full
// line scan, and a small file parses in
// well under a millisecond so chunking cannot help. Cross-file parallelism
// (the worker pool in parseFiles) still applies.
const parseParallelThreshold = 256 << 10

// parseMD parses a single .md file and returns the assembled content tree
// roots, the file title (the first h1 heading, when present), and any YAML
// frontmatter as a map.
//
// Level-1 headings are document titles, not content: they are never nodes
// and can never be requirements (only h2+ can). The first h1 in the file is
// captured as the file title; the tree assembled from the remaining nodes
// starts at the content heading level.
//
// For large files with multiple root-level requirements (parent=nil), the
// file is split at root requirement boundaries and chunks are parsed in
// parallel. Each chunk yields a flat node stream in source order; the
// streams are concatenated and assembled into the tree afterwards so
// nesting across chunk boundaries (e.g. a container in chunk 0 with a
// requirement in chunk 1) is preserved. The pre-scan to find split
// points is a cheap line scan that replicates the stack logic without
// building a goldmark AST.
func parseMD(src []byte, sourcePath string) ([]*model.Node, string, map[string]any, error) {
	// A file that opens a frontmatter block and never closes it would have the
	// rest of its content swallowed by the metadata extension, silently and with
	// no diagnostic. Fail instead: this is the one shape where nothing else in
	// the pipeline can notice the loss.
	if _, ok := frontmatterEnd(src); !ok {
		return nil, "", nil, fmt.Errorf("%s: frontmatter opened on the first line is never closed; add a closing --- line or remove the leading dashes", sourcePath)
	}

	// Small files: skip the root-split pre-scan entirely. The serial path
	// produces the same tree (assembleTree on one flat stream).
	if len(src) < parseParallelThreshold {
		nodes, title, fm, err := parseMDChunk(src, sourcePath)
		if err != nil {
			return nil, "", fm, err
		}
		return assembleTree(nodes), title, fm, nil
	}

	splits := findRootSplits(src)

	// No splits or a single root → parse the whole file serially.
	nRoots := len(splits)
	if nRoots <= 1 {
		nodes, title, fm, err := parseMDChunk(src, sourcePath)
		if err != nil {
			return nil, "", fm, err
		}
		return assembleTree(nodes), title, fm, nil
	}

	// Batch root requirements into ~NumCPU chunks: repeated goldmark
	// parses of smaller chunks run in parallel, which pays off only for
	// files large enough to split (see parseParallelThreshold).
	// Chunks are balanced by root count, not byte size.
	nCPU := runtime.NumCPU()
	nChunks := min(nRoots, nCPU)

	// Compute chunk boundaries: which split indices start/end each chunk.
	// Chunk 0 additionally covers the file preamble [0, splits[0]) so
	// leading containers and prose (including the h1 title) are never
	// dropped.
	type chunkResult struct {
		err   error
		meta  map[string]any
		nodes []*model.Node
	}
	results := make([]chunkResult, nChunks)
	var title string
	var frontmatter map[string]any
	var wg sync.WaitGroup

	for c := range nChunks {
		// Root index range [rootStart, rootEnd) for this chunk.
		rootStart := c * nRoots / nChunks
		rootEnd := min((c+1)*nRoots/nChunks, nRoots)
		// Byte range: from the start of the first root in this chunk
		// to the start of the first root in the next chunk (or EOF).
		byteStart := splits[rootStart]
		if rootStart == 0 {
			byteStart = 0 // include the preamble in chunk 0
		}
		byteEnd := len(src)
		if rootEnd < nRoots {
			byteEnd = splits[rootEnd]
		}
		chunk := src[byteStart:byteEnd]
		idx := c
		wg.Go(func() {
			nodes, chunkTitle, chunkMeta, err := parseMDChunk(chunk, sourcePath)
			results[idx] = chunkResult{nodes: nodes, err: err, meta: chunkMeta}
			// Frontmatter can only live in chunk 0: it is the file preamble.
			if idx == 0 {
				frontmatter = chunkMeta
				if chunkTitle != "" {
					title = chunkTitle
				}
			}
		})
	}
	wg.Wait()

	var all []*model.Node
	for _, r := range results {
		if r.err != nil {
			return nil, "", frontmatter, r.err
		}
		all = append(all, r.nodes...)
	}
	return assembleTree(all), title, frontmatter, nil
}

// parseMDChunk parses a single chunk of markdown (possibly the whole file)
// using goldmark and returns the flat node stream in source order, the file
// title (first h1 heading), and any YAML frontmatter. Every heading except a
// level-1 document title becomes a node; prose and other blocks attach to the
// most recently opened node's Body. Nodes are NOT assembled into a tree here —
// callers must assemble the concatenated streams so nesting across chunk
// boundaries is preserved.
func parseMDChunk(src []byte, sourcePath string) ([]*model.Node, string, map[string]any, error) {
	pc := parser.NewContext()
	doc := mdParser.Parse(src, parser.WithContext(pc))

	// Frontmatter is parsed by the meta extension into the document node; it is
	// metadata, never content, so it does not open a node of its own.
	var frontmatter map[string]any
	if d, ok := doc.(*ast.Document); ok {
		frontmatter = d.Metadata()
	}

	var nodes []*model.Node
	var cur *model.Node
	var waitingForAttr bool
	var seenHeading bool
	var title string
	var bodyBuf strings.Builder
	// ensureOpenNode creates a headingless info node (level 0) when prose
	// or other blocks appear before the first heading, so nothing is dropped.
	ensureOpenNode := func() {
		if cur == nil {
			cur = &model.Node{Kind: model.KindSection, Level: 0, Source: sourcePath}
			nodes = append(nodes, cur)
		}
	}
	// closeNode flushes the open node's body and resets the cursor so the
	// next content starts a fresh node.
	closeNode := func() {
		if cur != nil {
			cur.Body = bodyBuf.String()
			bodyBuf.Reset()
			cur = nil
			waitingForAttr = false
		}
	}
	for n := doc.FirstChild(); n != nil; n = n.NextSibling() {
		switch v := n.(type) {
		case *ast.Heading:
			level := v.Level
			hasAdjacentAttr := isNextAttrBlock(n, src)
			heading := collectInlineText(v, src)

			// Level 1 is the document title, never content: it cannot be a
			// requirement, and it does not become a section node. It must also
			// be the first heading in the file — a second h1 would silently
			// lose its text, so it is rejected instead.
			if level == 1 {
				if hasAdjacentAttr {
					return nil, "", frontmatter, fmt.Errorf("%s: heading %q is a level-1 requirement; requirement headings must be level 2 or higher (h1 is the document title)", sourcePath, heading)
				}
				if seenHeading {
					return nil, "", frontmatter, fmt.Errorf("%s: heading %q is a second level-1 heading; the h1 is the document title and must be the first heading in the file", sourcePath, heading)
				}
				seenHeading = true
				closeNode()
				if title == "" {
					title = strings.TrimSpace(heading)
				}
				continue
			}

			seenHeading = true
			closeNode()
			kind := model.KindSection
			if hasAdjacentAttr {
				kind = model.KindRequirement
			}
			cur = &model.Node{
				Kind:   kind,
				Level:  level,
				Source: sourcePath,
			}
			if hasAdjacentAttr {
				cur.ID, cur.Title = splitHeadingID(heading)
			} else {
				cur.Title = heading
			}
			nodes = append(nodes, cur)
			waitingForAttr = hasAdjacentAttr

		case *ast.CodeBlock:
			// v2 merges fenced and indented code blocks into one node kind.
			// An indented block is never an attr fence; store it as a fence so
			// the body keeps its code semantics without its indentation.
			if v.CodeBlockKind != ast.CodeBlockKindFenced {
				ensureOpenNode()
				appendBodyBlock(&bodyBuf, renderFence("", v, src))
				continue
			}
			if cur != nil && waitingForAttr && codeInfo(v, src) == attrFenceInfo {
				attrs, err := decodeAttrs(string(v.Value.Bytes(src)))
				if err != nil {
					return nil, "", frontmatter, fmt.Errorf("%s: req %s: invalid attr YAML: %w", sourcePath, cur.ID, err)
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
			} else {
				ensureOpenNode()
				// Non-attr code block inside a node's body
				appendBodyBlock(&bodyBuf, renderFence(codeInfo(v, src), v, src))
			}

		case *ast.Paragraph:
			text := paragraphText(v, src)
			if text == "" {
				continue
			}
			ensureOpenNode()
			if trimmed, ok := strings.CutPrefix(text, "*Rationale:"); ok {
				// Handle both *Rationale: and *Rationale:* (markdown italic)
				trimmed = strings.TrimLeft(trimmed, "* ")
				cur.Rationale = strings.TrimSpace(trimmed)
			} else {
				appendBodyBlock(&bodyBuf, text)
			}

		case *meta.MetaBlock:
			// YAML frontmatter is document metadata, not content: the meta
			// extension already stored it on the AST, so it opens no node.
			continue

		case *ast.ThematicBreak:
			// Thematic breaks are content: keep them in the open node's body.
			ensureOpenNode()
			appendBodyBlock(&bodyBuf, "---")

		case *ast.HTMLBlock:
			// The generated-content marker is tool metadata, not content: it
			// would otherwise open a phantom headingless node in every file the
			// renderer writes. Both markers are recognized (the second is the
			// pre-rename one). The renderer lives in another module, so the
			// literals are duplicated here; TestE2E_* proves they still agree.
			if isGeneratedMarker(v, src) {
				continue
			}
			ensureOpenNode()
			appendBodyBlock(&bodyBuf, renderBlock(src, n))

		default:
			// Any other block (blockquote, list, table, raw HTML) → body text if
			// inside a node.
			ensureOpenNode()
			appendBodyBlock(&bodyBuf, renderBlock(src, n))
		}
	}

	if cur != nil {
		cur.Body = bodyBuf.String()
	}

	return nodes, title, frontmatter, nil
}

// reqmd:trace SW-PAR-002
// assembleTree builds the document content tree from a flat, ordered node
// stream. A node's parent is the nearest preceding heading with a
// shallower level; requirements additionally get ParentID pointing at the
// nearest preceding requirement heading. Sections keep KindSection whether
// they end up with children or not: folder-ness is derived from Children, so
// adding a subsection never changes a node's kind.
func assembleTree(nodes []*model.Node) []*model.Node {
	var roots []*model.Node
	var stack []*model.Node    // open headings by level
	var reqStack []*model.Node // open requirements by level (for ParentID)
	for _, n := range nodes {
		stack = popByLevel(stack, n.Level)
		reqStack = popByLevel(reqStack, n.Level)
		if n.Level == 0 || len(stack) == 0 {
			roots = append(roots, n)
		} else {
			stack[len(stack)-1].Children = append(stack[len(stack)-1].Children, n)
		}
		if n.Kind == model.KindRequirement {
			if len(reqStack) > 0 {
				n.ParentID = reqStack[len(reqStack)-1].ID
			}
			reqStack = append(reqStack, n)
		}
		if n.Level > 0 {
			stack = append(stack, n)
		}
	}
	return roots
}

// popByLevel removes stack entries whose level is >= target, mirroring
// the heading-nesting rule: a heading at level L closes everything at
// level L or deeper.
func popByLevel(stack []*model.Node, target int) []*model.Node {
	for len(stack) > 0 && stack[len(stack)-1].Level >= target {
		stack = stack[:len(stack)-1]
	}
	return stack
}

// reqStackEntry tracks an open requirement at a heading level during the
// root-split pre-scan.
type reqStackEntry struct {
	level int
}

// popStackToLevel removes entries whose level is >= targetLevel.
// Used by the root-split pre-scan to replicate requirement nesting.
func popStackToLevel(stack []reqStackEntry, targetLevel int) []reqStackEntry {
	for len(stack) > 0 && stack[len(stack)-1].level >= targetLevel {
		stack = stack[:len(stack)-1]
	}
	return stack
}

// findRootSplits scans src line-by-line to find byte offsets where root-level
// requirements begin. A root requirement is one where the requirement stack
// is empty when the requirement heading is encountered. The scan replicates
// the same stack logic as the assembly step but without building a goldmark
// AST.
//
// Returns the byte offsets of root requirement starts (including the heading
// line) and any YAML frontmatter parsed from the file header.
//
// If only 0 or 1 roots are found, the caller parses the whole file serially.
func findRootSplits(src []byte) (splits []int) {
	// Skip a leading frontmatter block so its lines are never scanned for
	// headings; an unterminated block is rejected by parseMD before we get here.
	pos, _ := frontmatterEnd(src)

	var stack []reqStackEntry

	for pos < len(src) {
		lineStart := pos
		line, _ := nextLine(src, pos)
		lineEnd := pos + len(line)

		if level, ok := headingLevelBytes(line); ok {
			// Level 1 is the document title: it can never be a requirement,
			// so it only closes any open headings (matching the chunk parse).
			if level == 1 {
				stack = popStackToLevel(stack, level)
			} else if hasAttr := nextNonBlankIsAttrBytes(src, lineEnd+1); hasAttr {
				stack = popStackToLevel(stack, level)
				if len(stack) == 0 {
					splits = append(splits, lineStart)
				}
				stack = append(stack, reqStackEntry{level: level})
			} else {
				stack = popStackToLevel(stack, level)
			}
		}

		pos = lineEnd + 1 // skip \n
	}

	return splits
}

// frontmatterEnd returns the offset just past a leading YAML frontmatter block,
// or 0 when the file does not start with one. ok is false when the block is
// opened on the first line and never closed.
func frontmatterEnd(src []byte) (end int, ok bool) {
	firstLine, _ := nextLine(src, 0)
	if !isDashLine(firstLine) {
		return 0, true
	}
	for pos := len(firstLine) + 1; pos < len(src); {
		line, _ := nextLine(src, pos)
		if isDashLine(line) {
			return min(pos+len(line)+1, len(src)), true
		}
		pos += len(line) + 1
	}
	return 0, false
}

// isDashLine reports whether line is a non-empty run of dashes, surrounding
// spaces ignored. It mirrors the frontmatter extension's own separator rule so
// the split scan and the parser agree on where a metadata block starts and ends
// (the extension accepts "--", "----", and any longer run, not just "---").
func isDashLine(line []byte) bool {
	line = bytes.TrimSpace(line)
	if len(line) == 0 {
		return false
	}
	for _, c := range line {
		if c != '-' {
			return false
		}
	}
	return true
}


// nextLine returns the line at byte offset pos (without the trailing \n)
// and whether there was a newline. The returned slice references src —
// no allocation.
func nextLine(src []byte, pos int) (line []byte, hasNL bool) {
	if pos >= len(src) {
		return nil, false
	}
	line, _, hasNL = bytes.Cut(src[pos:], []byte{'\n'})
	return line, hasNL
}

// headingLevelBytes returns the heading level (1-6) and true if the line
// is a markdown heading (e.g. "## Title"). Returns 0, false otherwise.
// Works on []byte to avoid string allocation.
func headingLevelBytes(line []byte) (int, bool) {
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

// nextNonBlankIsAttrBytes checks if the next non-blank line after byte
// offset pos is a ```attr fence. Scans src directly — no allocation.
func nextNonBlankIsAttrBytes(src []byte, pos int) bool {
	for pos < len(src) {
		line, hasNL := nextLine(src, pos)
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) == 0 {
			pos += len(line) + 1
			if !hasNL {
				break
			}
			continue
		}
		return bytes.HasPrefix(trimmed, attrFenceBytes)
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
	fence, ok := next.(*ast.CodeBlock)
	if !ok || fence.CodeBlockKind != ast.CodeBlockKindFenced {
		return false
	}
	return codeInfo(fence, src) == attrFenceInfo
}

// splitHeadingID splits a heading like "REQ-001: This is a title"
// into ("REQ-001", "This is a title"). If there's no colon+space,
// returns (heading, "").
func splitHeadingID(heading string) (id string, title string) {
	if id, title, found := strings.Cut(heading, ": "); found {
		return strings.TrimSpace(id), strings.TrimSpace(title)
	}
	return strings.TrimSpace(heading), ""
}

// --- text extraction helpers ---

func collectInlineText(n ast.Node, src []byte) string {
	var b strings.Builder
	_ = ast.Walk(n, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if tn, ok := n.(*ast.Text); ok {
			b.WriteString(tn.Value.Value(src))
		}
		return ast.WalkContinue, nil
	})
	return strings.TrimSpace(b.String())
}

// codeInfo returns the fenced block's full info string ("go title=\"x\""), not
// just its language token: the body keeps its Markdown, so anything after the
// language must survive the round-trip.
func codeInfo(c *ast.CodeBlock, src []byte) string {
	return strings.TrimSpace(c.Info.Value(src))
}

// appendBodyBlock appends a block to the open node's body, separated from any
// preceding prose by a blank line. Without that separator a fenced code block or
// a thematic break glues onto the paragraph above it, and the fence stops being
// a fence once the body Markdown is rendered again.
func appendBodyBlock(b *strings.Builder, block string) {
	if block == "" {
		return
	}
	if b.Len() > 0 {
		b.WriteString("\n\n")
	}
	b.WriteString(block)
}

// renderFence re-emits a code block as a fenced block, so the stored body keeps
// the delimiters the exporter needs to render it again.
func renderFence(info string, c *ast.CodeBlock, src []byte) string {
	body := c.Value.Bytes(src)
	var b strings.Builder
	b.WriteString("```")
	b.WriteString(info)
	b.WriteString("\n")
	b.Write(body)
	if len(body) > 0 && body[len(body)-1] != '\n' {
		b.WriteString("\n")
	}
	b.WriteString("```\n\n")
	return b.String()
}

func paragraphText(p *ast.Paragraph, src []byte) string {
	var b strings.Builder
	for _, seg := range p.Source() {
		// Write the segment bytes directly (trimmed in place) instead of
		// converting each line to a string first — the conversion copies.
		_, _ = b.Write(bytes.TrimRight(seg.Bytes(src), "\r\n"))
	}
	return strings.TrimSpace(b.String())
}

// renderBlock returns the source markdown of a block that has no dedicated
// case: lists, tables, blockquotes, and raw HTML. The body of a node is
// markdown, so the block's original source is the faithful representation —
// concatenating the line segments instead would drop the list markers and the
// line breaks between items, turning a list into one run-on paragraph.
func renderBlock(src []byte, n ast.Node) string {
	start, stop := -1, -1
	_ = ast.Walk(n, func(m ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		// Only block nodes carry source segments — inline nodes (e.g. italic
		// text) have none.
		bn, ok := m.(ast.BlockNode)
		if !ok {
			return ast.WalkContinue, nil
		}
		for _, seg := range bn.Source() {
			if start < 0 || seg.Start < start {
				start = seg.Start
			}
			if seg.Stop > stop {
				stop = seg.Stop
			}
		}
		return ast.WalkContinue, nil
	})
	if start < 0 || stop <= start || stop > len(src) {
		return ""
	}
	// Widen to whole lines: a leaf segment starts after the list marker (or the
	// blockquote marker, or the indentation), so slicing at seg.Start would drop
	// the marker and glue the item onto whatever precedes it.
	lineStart := bytes.LastIndexByte(src[:start], '\n') + 1
	lineEnd := len(src)
	if nl := bytes.IndexByte(src[stop:], '\n'); nl >= 0 {
		lineEnd = stop + nl
	}
	return strings.TrimRight(string(src[lineStart:lineEnd]), " \t")
}

// ParseSingleFile is used by `reqmd check <file> --schema <schema>`.
func ParseSingleFile(path string) ([]*model.Node, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	nodes, _, _, err := parseMD(src, path)
	return nodes, err
}

// generatedMarker is the sentinel the renderer writes into every file it owns.
// Its pre-rename spelling is still recognized so trees generated before the
// renderer moved to reqmd-import keep parsing cleanly.
const (
	generatedMarker       = "<!-- reqmd: generated -->"
	legacyGeneratedMarker = "<!-- reqmd-code: generated -->"
)

// isGeneratedMarker reports whether an HTML block is one of the renderer's
// generated-content markers. Such a block is tool metadata: keeping it would
// open a phantom headingless content node in every generated document.
func isGeneratedMarker(block *ast.HTMLBlock, src []byte) bool {
	switch strings.TrimSpace(string(block.Value.Bytes(src))) {
	case generatedMarker, legacyGeneratedMarker:
		return true
	}
	return false
}
