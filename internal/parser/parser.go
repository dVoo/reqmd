package parser

import (
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

// parseMD uses goldmark to extract requirements from a single .md file
// and returns any YAML frontmatter as a map.
func parseMD(src []byte, sourcePath string) ([]model.Requirement, map[string]any, error) {
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
	var reqLevel int // 0 = unknown, set on first heading+attr pair
	var parentID string
	var waitingForAttr bool

	for n := doc.FirstChild(); n != nil; n = n.NextSibling() {
		switch v := n.(type) {
		case *ast.Heading:
			level := v.Level
			hasAdjacentAttr := isNextAttrBlock(n, src)

			// Discover reqLevel from the first heading+attr pair.
			if reqLevel == 0 && hasAdjacentAttr {
				reqLevel = level
			}

			// Before reqLevel is known, all headings are body text.
			if reqLevel == 0 {
				if cur != nil {
					cur.Body += renderHeading(v, src)
				}
				continue
			}

			switch {
			case level == reqLevel:
				// ── Top-level heading ──
				if hasAdjacentAttr {
					if cur != nil {
						if cur.Attrs == nil {
							return nil, frontmatter, fmt.Errorf("%s: req %s: missing attr block", sourcePath, cur.ID)
						}
						reqs = append(reqs, *cur)
						cur = nil
					}
					heading := collectInlineText(v, src)
					id, title := splitHeadingID(heading)
					cur = &model.Requirement{
						ID:     id,
						Title:  title,
						Source: sourcePath,
					}
					parentID = id
					waitingForAttr = true
				} else if cur != nil {
					// reqLevel heading without attr → body text
					cur.Body += renderHeading(v, src)
				}

			case level == reqLevel+1:
				// ── Sub-requirement level ──
				if hasAdjacentAttr {
					if cur != nil {
						if cur.Attrs == nil {
							return nil, frontmatter, fmt.Errorf("%s: req %s: missing attr block", sourcePath, cur.ID)
						}
						reqs = append(reqs, *cur)
						cur = nil
					}
					heading := collectInlineText(v, src)
					id, title := splitHeadingID(heading)
					if parentID == "" {
						return nil, frontmatter, fmt.Errorf("%s: ### heading %q without preceding %s requirement",
							sourcePath, heading, headingName(reqLevel))
					}
					cur = &model.Requirement{
						ID:       id,
						Title:    title,
						Source:   sourcePath,
						ParentID: parentID,
					}
					waitingForAttr = true
				} else if cur != nil {
					// Sub-req-level heading without attr → body text
					cur.Body += renderHeading(v, src)
				}

			default:
				// Other heading levels → body text
				if cur != nil {
					cur.Body += renderHeading(v, src)
				}
			}

		case *ast.FencedCodeBlock:
			if cur != nil && waitingForAttr && fenceInfo(v, src) == attrFenceInfo {
				yamlText := fenceLines(v, src)
				var attrs map[string]any
				if err := yaml.Unmarshal([]byte(yamlText), &attrs); err != nil {
					return nil, frontmatter, fmt.Errorf("%s: req %s: invalid attr YAML: %w", sourcePath, cur.ID, err)
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
				cur.Body += renderFence(v, src)
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
					if cur.Body != "" {
						cur.Body += "\n\n"
					}
					cur.Body += text
				}
			}

		case *ast.ThematicBreak:
			// ignore separators

		default:
			// Any other block (blockquote, list, etc.) → body text if inside a requirement
			if cur != nil {
				cur.Body += renderBlock(src, n)
			}
		}
	}

	// Finalize last requirement
	if cur != nil {
		if cur.Attrs == nil {
			return nil, frontmatter, fmt.Errorf("%s: req %s: missing attr block", sourcePath, cur.ID)
		}
		reqs = append(reqs, *cur)
	}

	return reqs, frontmatter, nil
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

func headingName(level int) string {
	switch level {
	case 1:
		return "#"
	case 2:
		return "##"
	default:
		return fmt.Sprintf("level %d", level)
	}
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
