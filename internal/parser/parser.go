// Package parser discovers reqmd document directories, parses the
// Markdown content tree (requirements, containers, and info items), and
// loads document trees from git tags.
package parser

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reqmd/internal/model"
	"reqmd/internal/schema"
	"runtime"
	"strings"
	"sync"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"

	katex "github.com/FurqanSoftware/goldmark-katex"
	fences "github.com/stefanfritsch/goldmark-fences"
	emoji "github.com/yuin/goldmark-emoji"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	meta "github.com/yuin/goldmark-meta"
	"go.abhg.dev/goldmark/mermaid"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

const attrFenceInfo = "attr"

// Package-level patterns reused in the hot per-line scan helpers. Hoisted
// so the hot paths don't allocate a []byte pattern on every call.
var (
	attrFenceBytes = []byte("```attr")
	frontmatterSep = []byte("---")
)

// mdParser is the shared goldmark instance used by parseMD.
// Constructed once at package init to avoid re-running extension Init for every .md file.
var mdParser = goldmark.New(
	goldmark.WithExtensions(
		extension.Table,
		extension.Strikethrough,
		extension.TaskList,
		&mermaid.Extender{},
		&fences.Extender{},
		highlighting.Highlighting,
		emoji.Emoji,
		meta.Meta,
		&katex.Extender{},
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
// document title (the first file's h1 heading, if any), merged
// frontmatter, and any per-file errors. Files are parsed in parallel;
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
			nodes, title, frontmatter, err := parseMD(src, f)
			if err != nil {
				results[i] = result{err: fmt.Errorf("parsing %s: %w", f, err)}
				return
			}
			results[i] = result{nodes: nodes, title: title, meta: frontmatter}
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
		// Merge frontmatter in order (first file's value wins per key)
		for k, v := range r.meta {
			if _, exists := mergedMeta[k]; !exists {
				mergedMeta[k] = v
			}
		}
	}
	return all, title, mergedMeta, errs
}

// parseMD parses a single .md file and returns the assembled content tree
// roots, the file title (the first h1 heading, when present), and any YAML
// frontmatter as a map.
//
// Level-1 headings are document titles, not content: they are never nodes
// and can never be requirements (only h2+ can). The first h1 in the file is
// captured as the file title; the tree assembled from the remaining nodes
// starts at the content heading level.
//
// For files with multiple root-level requirements (parent=nil), the file
// is split at root requirement boundaries and chunks are parsed in
// parallel. Each chunk yields a flat node stream in source order; the
// streams are concatenated and assembled into the tree afterwards so
// nesting across chunk boundaries (e.g. a container in chunk 0 with a
// requirement in chunk 1) is preserved. The pre-scan to find split
// points is a cheap line scan that replicates the stack logic without
// building a goldmark AST.
func parseMD(src []byte, sourcePath string) ([]*model.Node, string, map[string]any, error) {
	splits, frontmatter := findRootSplits(src)

	// No splits or a single root → parse the whole file serially.
	nRoots := len(splits)
	if nRoots <= 1 {
		nodes, title, fm, err := parseMDChunk(src, sourcePath)
		if err != nil {
			return nil, "", fm, err
		}
		return assembleTree(nodes), title, fm, nil
	}

	// Batch root requirements into ~NumCPU chunks to amortize goldmark
	// init overhead (~490ns per parse) while still getting parallelism.
	// Each chunk contains multiple consecutive root requirements + their
	// descendants. Chunks are balanced by root count, not byte size.
	nCPU := runtime.NumCPU()
	nChunks := min(nRoots, nCPU)

	// Compute chunk boundaries: which split indices start/end each chunk.
	// Chunk 0 additionally covers the file preamble [0, splits[0]) so
	// leading containers and prose (including the h1 title) are never
	// dropped.
	type chunkResult struct {
		err   error
		nodes []*model.Node
	}
	results := make([]chunkResult, nChunks)
	var title string
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
			nodes, chunkTitle, _, err := parseMDChunk(chunk, sourcePath)
			results[idx] = chunkResult{nodes: nodes, err: err}
			if idx == 0 && chunkTitle != "" {
				title = chunkTitle
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
// title (first h1 heading), and any YAML frontmatter. Every heading except
// a level-1 document title becomes a node; prose and other blocks attach to
// the most recently opened node's Body. Nodes are NOT assembled into a tree
// here — callers must assemble the concatenated streams so nesting across
// chunk boundaries is preserved.
func parseMDChunk(src []byte, sourcePath string) ([]*model.Node, string, map[string]any, error) {
	p := mdParser.Parser()
	parseCtx := parser.NewContext()
	doc := p.Parse(text.NewReader(src), parser.WithContext(parseCtx))

	// Extract YAML frontmatter (parsed by meta.Meta, stored in context).
	var frontmatter map[string]any
	if fm := meta.Get(parseCtx); fm != nil {
		frontmatter = fm
	}

	var nodes []*model.Node
	var cur *model.Node
	var waitingForAttr bool
	var title string
	var bodyBuf strings.Builder
	// ensureOpenNode creates a headingless info node (level 0) when prose
	// or other blocks appear before the first heading, so nothing is dropped.
	ensureOpenNode := func() {
		if cur == nil {
			cur = &model.Node{Kind: model.KindInfo, Level: 0, Source: sourcePath}
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
			// requirement, and it does not become a container or info node.
			if level == 1 {
				if hasAdjacentAttr {
					return nil, "", frontmatter, fmt.Errorf("%s: heading %q is a level-1 requirement; requirement headings must be level 2 or higher (h1 is the document title)", sourcePath, heading)
				}
				closeNode()
				if title == "" {
					title = strings.TrimSpace(heading)
				}
				continue
			}

			closeNode()
			kind := model.KindInfo
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

		case *ast.FencedCodeBlock:
			if cur != nil && waitingForAttr && fenceInfo(v, src) == attrFenceInfo {
				yamlText := fenceLines(v, src)
				attrs, err := decodeAttrYAML(yamlText)
				if err != nil {
					// Fallback to yaml.v3 for complex syntax the scanner
					// doesn't handle (nested maps, flow sequences, etc.).
					attrs = make(map[string]any, 8)
					if yamlErr := yaml.Unmarshal([]byte(yamlText), &attrs); yamlErr != nil {
						return nil, "", frontmatter, fmt.Errorf("%s: req %s: invalid attr YAML: %w", sourcePath, cur.ID, yamlErr)
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
			} else {
				ensureOpenNode()
				// Non-attr code block inside a node's body
				bodyBuf.WriteString(renderFence(v, src))
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
				if bodyBuf.Len() > 0 {
					bodyBuf.WriteString("\n\n")
				}
				bodyBuf.WriteString(text)
			}

		case *ast.ThematicBreak:
			// Thematic breaks are content: keep them in the open node's body.
			ensureOpenNode()
			if bodyBuf.Len() > 0 {
				bodyBuf.WriteString("\n\n")
			}
			bodyBuf.WriteString("---")

		default:
			// Any other block (blockquote, list, etc.) → body text if inside a node
			ensureOpenNode()
			bodyBuf.WriteString(renderBlock(src, n))
		}
	}

	if cur != nil {
		cur.Body = bodyBuf.String()
	}

	return nodes, title, frontmatter, nil
}

// assembleTree builds the document content tree from a flat, ordered node
// stream. A node's parent is the nearest preceding heading with a
// shallower level; requirements additionally get ParentID pointing at the
// nearest preceding requirement heading. Info nodes with children are
// promoted to KindContainer.
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
	promoteContainers(roots)
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

// promoteContainers upgrades info nodes that have children to containers.
func promoteContainers(nodes []*model.Node) {
	for _, n := range nodes {
		promoteContainers(n.Children)
		if n.Kind == model.KindInfo && len(n.Children) > 0 {
			n.Kind = model.KindContainer
		}
	}
}

// reqStackEntry tracks an open requirement at a heading level.
type reqStackEntry struct {
	id    string
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
func findRootSplits(src []byte) (splits []int, frontmatter map[string]any) {
	// Parse frontmatter if present (--- delimited block at file start).
	pos := 0 // byte offset of current line start

	// Check for frontmatter delimiter on the first line.
	if firstLine, _ := nextLine(src, 0); bytes.Equal(bytes.TrimSpace(firstLine), frontmatterSep) {
		pos += len(firstLine) + 1 // skip the --- line
		for pos < len(src) {
			line, nl := nextLine(src, pos)
			if bytes.Equal(bytes.TrimSpace(line), frontmatterSep) {
				// Found closing ---. Parse frontmatter between markers.
				if pos > len(firstLine)+1 {
					fmStart := len(firstLine) + 1
					fmEnd := pos
					_ = yaml.Unmarshal(src[fmStart:fmEnd], &frontmatter)
				}
				pos += len(line) + 1 // skip closing ---
				break
			}
			pos += len(line) + 1
			_ = nl
		}
	}

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
				headingText := bytes.TrimSpace(line[level:])
				id, _ := splitHeadingIDBytes(headingText)
				stack = append(stack, reqStackEntry{level: level, id: id})
			} else {
				stack = popStackToLevel(stack, level)
			}
		}

		pos = lineEnd + 1 // skip \n
	}

	return splits, frontmatter
}

// nextLine returns the line at byte offset pos (without the trailing \n)
// and whether there was a newline. The returned slice references src —
// no allocation.
func nextLine(src []byte, pos int) (line []byte, hasNL bool) {
	if pos >= len(src) {
		return nil, false
	}
	rest := src[pos:]
	line, _, hasNL = bytes.Cut(rest, []byte{'\n'})
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

// splitHeadingIDBytes splits a heading like "REQ-001: This is a title"
// into ("REQ-001", "This is a title"). Works on []byte.
func splitHeadingIDBytes(heading []byte) (id string, title string) {
	if idx := bytes.IndexByte(heading, ':'); idx >= 0 && idx+1 < len(heading) && heading[idx+1] == ' ' {
		return string(bytes.TrimSpace(heading[:idx])), string(bytes.TrimSpace(heading[idx+2:]))
	}
	return string(bytes.TrimSpace(heading)), ""
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
	for i := range f.Lines().Len() {
		seg := f.Lines().At(i)
		b.Write(seg.Value(src))
	}
	return b.String()
}

func paragraphText(p *ast.Paragraph, src []byte) string {
	var b strings.Builder
	for i := range p.Lines().Len() {
		seg := p.Lines().At(i)
		b.WriteString(strings.TrimRight(string(seg.Value(src)), "\n\r"))
	}
	return strings.TrimSpace(b.String())
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
	_ = ast.Walk(n, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		// Only block nodes have Lines() — inline nodes (e.g., italic text)
		// panic when Lines() is called.
		if n.Type() == ast.TypeInline {
			return ast.WalkContinue, nil
		}
		for i := range n.Lines().Len() {
			seg := n.Lines().At(i)
			b.Write(seg.Value(src))
		}
		return ast.WalkContinue, nil
	})
	return b.String()
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
