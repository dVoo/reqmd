package exchange

import (
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/dVoo/reqmd/reqmd-import/internal/writer"

	"gopkg.in/yaml.v3"
)

// maxProblems caps how many violations one run reports, so a large broken
// bundle produces a readable list instead of thousands of lines.
const maxProblems = 20

// maxRenderDepth is the deepest nesting the renderer accepts: depth 0 renders
// "##", so depth 4 is "######" and depth 5 would be seven hashes, which
// CommonMark does not treat as a heading.
const maxRenderDepth = 4

// Options carries the renderer's non-format configuration.
type Options struct {
	// StatusMap translates a source system's status vocabulary onto reqmd's
	// ("In Review" -> "review"). An unmapped value falls back to reqmd's
	// default and keeps the original as x-reqmd.source-status.
	StatusMap map[string]string
}

// ValidationError reports every violation found in one pass. Producers fix a
// batch, not one error at a time.
type ValidationError struct {
	Problems []string
}

func (e *ValidationError) Error() string {
	var b strings.Builder
	b.WriteString("invalid exchange bundle:")
	for _, p := range e.Problems {
		b.WriteString("\n  - ")
		b.WriteString(p)
	}
	return b.String()
}

// statusAdditionPattern matches x-reqmd.additional-status-values entries.
var statusAdditionPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

// Build validates the bundle and renders every document to output files. A
// document without items produces only its schema.yaml; one with items also
// produces its .md file. Items keep the order they were given in, at every
// level — the renderer never reorders, normalizes or drops anything.
func (b *Bundle) Build() ([]writer.File, error) {
	return b.BuildWith(Options{})
}

// BuildWith is Build with renderer options (status mapping).
func (b *Bundle) BuildWith(opts Options) ([]writer.File, error) {
	v := &validator{opts: opts}
	v.validate(b)
	if len(v.problems) > 0 {
		return nil, &ValidationError{Problems: v.problems}
	}
	return v.render(b)
}

// validator accumulates violations and the per-document render state.
type validator struct {
	opts     Options
	problems []string
	// documents is the bundle being validated, kept for cross-document link
	// resolution and for rendering after validation succeeds.
	documents []Document
	// docKeys maps a document key to its index, for cross-document links.
	docKeys map[string]int
	// kinds maps a document index to key -> item kind, so link targets can be
	// checked without re-walking the tree.
	kinds map[int]map[string]string
	// additions holds mapped statuses that extend a document's enum, keyed by
	// document index.
	additions map[int][]string
}

func (v *validator) addf(format string, args ...any) {
	if len(v.problems) >= maxProblems {
		return
	}
	v.problems = append(v.problems, fmt.Sprintf(format, args...))
}

func (v *validator) validate(b *Bundle) {
	v.documents = b.Documents
	v.docKeys = make(map[string]int, len(b.Documents))
	v.additions = make(map[int][]string, len(b.Documents))

	if b.Schema != SchemaURL {
		v.addf("bundle $schema is %q, want %q", b.Schema, SchemaURL)
	}
	if len(b.Documents) == 0 {
		v.addf("bundle has no documents")
		return
	}

	// Pass 1: document identity, so cross-document links can resolve.
	for i := range b.Documents {
		d := &b.Documents[i]
		if d.Key == "" {
			v.addf("document %d has no key", i)
			continue
		}
		if strings.ContainsAny(d.Key, "/") || strings.ContainsAny(d.Key, " \t\n\r\v\f") {
			v.addf("document %q: key must not contain '/' or whitespace", d.Key)
			continue
		}
		if prev, dup := v.docKeys[d.Key]; dup {
			v.addf("document key %q is used by both document %d and %d", d.Key, prev+1, i+1)
			continue
		}
		v.docKeys[d.Key] = i
	}

	// Pass 1b: key -> kind per document, so links can be resolved while the
	// content pass walks the tree.
	v.kinds = make(map[int]map[string]string, len(b.Documents))
	for i := range b.Documents {
		v.kinds[i] = map[string]string{}
		var walk func(items []Item)
		walk = func(items []Item) {
			for j := range items {
				if k := items[j].Key; k != "" {
					v.kinds[i][k] = items[j].Kind
				}
				walk(items[j].Children)
			}
		}
		walk(b.Documents[i].Items)
	}

	// Pass 2: per-document content.
	for i := range b.Documents {
		v.validateDocument(i, &b.Documents[i])
	}

	if len(v.problems) == maxProblems && len(v.problems) < v.totalProblems() {
		v.problems = append(v.problems, "… and more (only the first violations are shown)")
	}
}

// totalProblems reports how many violations were seen, including the ones
// dropped by the cap.
func (v *validator) totalProblems() int { return len(v.problems) }

func (v *validator) validateDocument(idx int, d *Document) {
	if err := checkSingleLine("document title", d.Title); err != nil {
		v.addf("document %q: %v", docName(d), err)
	}
	if d.Title == "" {
		v.addf("document %q: title is required", docName(d))
	}
	// Keys are document-local, so uniqueness is checked per document.
	seen := make(map[string]string, len(d.Items))
	v.validateItems(idx, d, d.Items, nil, 0, seen)

	// `required` must actually be satisfied by every requirement.
	requiredSet := make(map[string]struct{}, len(d.Required))
	for _, name := range d.Required {
		requiredSet[name] = struct{}{}
	}
	if len(requiredSet) > 0 {
		var walk func(items []Item, path []string)
		walk = func(items []Item, itemPath []string) {
			for i := range items {
				it := &items[i]
				here := append(append([]string(nil), itemPath...), itemLabel(it))
				if it.Kind == KindRequirement {
					for _, name := range d.Required {
						if _, ok := it.Fields[name]; !ok {
							v.addf("document %q: requirement %q is missing required attribute %q",
								docName(d), strings.Join(here, "/"), name)
						}
					}
				}
				walk(it.Children, here)
			}
		}
		walk(d.Items, []string{docName(d)})
	}
}

func (v *validator) validateItems(idx int, d *Document, items []Item, itemPath []string, depth int, seen map[string]string) {
	for i := range items {
		it := &items[i]
		here := append(append([]string(nil), itemPath...), itemLabel(it))
		where := strings.Join(here, "/")

		switch it.Kind {
		case KindRequirement:
			if it.Key == "" {
				v.addf("document %q: item %q is a requirement without a key", docName(d), where)
			} else if strings.ContainsAny(it.Key, " \t\n\r\v\f") {
				v.addf("document %q: requirement key %q must not contain whitespace", docName(d), where)
			}
			if prev, dup := seen[it.Key]; dup && it.Key != "" {
				v.addf("document %q: key %q is used by both %q and %q", docName(d), it.Key, prev, where)
			}
			seen[it.Key] = where
			if err := checkSingleLine("requirement title", it.Title); err != nil {
				v.addf("document %q: %q: %v", docName(d), where, err)
			}
			v.validateFields(d, it, where)
		case KindSection:
			if len(it.Fields) > 0 {
				v.addf("document %q: %q is a section and must not carry fields", docName(d), where)
			}
			if it.Status != "" {
				v.addf("document %q: %q is a section and must not carry a status", docName(d), where)
			}
			if it.Key != "" {
				if prev, dup := seen[it.Key]; dup {
					v.addf("document %q: key %q is used by both %q and %q", docName(d), it.Key, prev, where)
				}
				seen[it.Key] = where
			}
			if err := checkSingleLine("section title", it.Title); err != nil {
				v.addf("document %q: %q: %v", docName(d), where, err)
			}
		default:
			v.addf("document %q: %q has kind %q, want %q or %q",
				docName(d), where, it.Kind, KindRequirement, KindSection)
		}

		v.validateLinks(idx, d, it, where)
		v.validateStatus(idx, d, it, where)

		if depth > maxRenderDepth {
			v.addf("document %q: %q would render heading level %d; Markdown allows six (flatten in the producer)",
				docName(d), where, 2+depth)
		}
		v.validateItems(idx, d, it.Children, here, depth+1, seen)
	}
}

func (v *validator) validateFields(d *Document, it *Item, where string) {
	for name, val := range it.Fields {
		if err := checkAttrKey(name); err != nil {
			v.addf("document %q: %q: %v", docName(d), where, err)
			continue
		}
		if len(d.Properties) > 0 {
			if _, declared := d.Properties[name]; !declared {
				v.addf("document %q: %q: attribute %q is not declared in the document's properties",
					docName(d), where, name)
			}
		}
		if err := checkScalarValue(name, val); err != nil {
			v.addf("document %q: %q: %v", docName(d), where, err)
		}
	}
}

func (v *validator) validateLinks(idx int, d *Document, it *Item, where string) {
	for _, link := range it.Links {
		targetDoc, targetKey := idx, link
		if docKey, key, qualified := cutDocRef(link); qualified {
			j, ok := v.docKeys[docKey]
			if !ok {
				v.addf("document %q: %q links to unknown document %q", docName(d), where, docKey)
				continue
			}
			targetDoc, targetKey = j, key
		}
		kind, found := v.lookup(targetDoc, targetKey)
		switch {
		case !found:
			v.addf("document %q: %q links to %q, which is not in the bundle", docName(d), where, link)
		case kind != KindRequirement:
			v.addf("document %q: %q links to %q, which is a section; links may only target requirements",
				docName(d), where, link)
		}
	}
}

// lookup finds an item by key in a document and returns its kind.
func (v *validator) lookup(idx int, key string) (string, bool) {
	kind, ok := v.kinds[idx][key]
	return kind, ok
}

func (v *validator) validateStatus(idx int, d *Document, it *Item, where string) {
	if it.Status == "" {
		return
	}
	mapped, ok := v.opts.StatusMap[it.Status]
	if !ok {
		// No mapping: a value that already looks like a reqmd status passes
		// through (and extends the enum when it is not a built-in).
		if isReqmdStatus(it.Status) {
			v.recordAddition(idx, it.Status)
		}
		// Otherwise it is a foreign word ("In Review"): the renderer falls back
		// to the default and keeps the original as provenance, rather than
		// rejecting a producer for using its own vocabulary.
		return
	}
	if !isReqmdStatus(mapped) {
		v.addf("document %q: %q maps status %q to %q, which is not a valid reqmd status value",
			docName(d), where, it.Status, mapped)
		return
	}
	v.recordAddition(idx, mapped)
}

// recordAddition extends a document's status enum for a non-built-in status.
func (v *validator) recordAddition(idx int, status string) {
	if isBuiltinStatus(status) {
		return
	}
	v.additions[idx] = append(v.additions[idx], status)
}

// isBuiltinStatus reports whether a status is one of reqmd's built-ins.
func isBuiltinStatus(s string) bool {
	switch s {
	case "draft", "approved", "rejected", "deferred":
		return true
	}
	return false
}

// isReqmdStatus reports whether a value is a status reqmd can carry: a built-in
// or a valid x-reqmd.additional-status-values token.
func isReqmdStatus(s string) bool {
	return isBuiltinStatus(s) || statusAdditionPattern.MatchString(s)
}

func (v *validator) render(b *Bundle) ([]writer.File, error) {
	var files []writer.File
	written := make(map[string]string)

	for i := range b.Documents {
		d := &b.Documents[i]
		dir, err := cleanRelDir(d.Dir())
		if err != nil {
			return nil, fmt.Errorf("document %q: %w", docName(d), err)
		}
		name, err := cleanFileName(d.FileName())
		if err != nil {
			return nil, fmt.Errorf("document %q: %w", docName(d), err)
		}

		schema, err := v.renderSchema(i, d)
		if err != nil {
			return nil, err
		}
		if err := add(&files, written, dir, "schema.yaml", []byte(writer.GeneratedMarker+"\n"+schema)); err != nil {
			return nil, err
		}
		if len(d.Items) == 0 {
			continue
		}
		items, err := v.renderItems(d, d.Items, 0, nil)
		if err != nil {
			return nil, err
		}
		md, err := writer.RenderDocument(d.Title, d.Body, items)
		if err != nil {
			return nil, err
		}
		if err := add(&files, written, dir, name, []byte(md)); err != nil {
			return nil, err
		}
	}
	return files, nil
}

// renderItems flattens the item tree into render order, resolving IDs, links,
// the status mapping and each item's heading depth.
func (v *validator) renderItems(d *Document, items []Item, depth int, refs refIndex) ([]writer.Item, error) {
	if refs == nil {
		refs = v.buildRefIndex(d)
	}
	out := make([]writer.Item, 0, len(items))
	for i := range items {
		it := &items[i]
		wi := writer.Item{
			Kind:   it.Kind,
			Title:  it.Title,
			Body:   it.Body,
			Fields: it.Fields,
			Depth:  depth,
		}
		if it.Kind == KindRequirement {
			wi.ID = d.prefix() + it.Key
			wi.Trace = append(wi.Trace, it.Refs...)
			for _, link := range it.Links {
				resolved, err := v.resolveLink(d, refs, link)
				if err != nil {
					return nil, err
				}
				wi.Trace = append(wi.Trace, resolved)
			}
			status, original := v.mapStatus(it.Status)
			wi.Status = status
			if original != "" {
				if wi.Fields == nil {
					wi.Fields = make(map[string]any, 1)
				}
				wi.Fields["x-reqmd.source-status"] = original
			}
		}
		out = append(out, wi)
		children, err := v.renderItems(d, it.Children, depth+1, refs)
		if err != nil {
			return nil, err
		}
		out = append(out, children...)
	}
	return out, nil
}

// mapStatus returns the reqmd status to render and, when the source value could
// not be mapped, the original value to preserve as provenance.
func (v *validator) mapStatus(status string) (mapped, original string) {
	if status == "" {
		return "approved", ""
	}
	if m, ok := v.opts.StatusMap[status]; ok {
		return m, ""
	}
	if isReqmdStatus(status) {
		return status, ""
	}
	return "approved", status
}

// refIndex maps a key to the reqmd ID of the requirement it names, per document
// key, so cross-document links resolve.
type refIndex map[string]string

func (v *validator) buildRefIndex(d *Document) refIndex {
	refs := make(refIndex, len(d.Items))
	var walk func(items []Item)
	walk = func(items []Item) {
		for i := range items {
			it := &items[i]
			if it.Kind == KindRequirement {
				refs[it.Key] = d.prefix() + it.Key
			}
			walk(it.Children)
		}
	}
	walk(d.Items)
	return refs
}

func (v *validator) resolveLink(d *Document, local refIndex, link string) (string, error) {
	docKey, key, qualified := cutDocRef(link)
	if !qualified {
		if id, ok := local[key]; ok {
			return id, nil
		}
		return "", fmt.Errorf("document %q: link %q does not resolve", docName(d), link)
	}
	j, ok := v.docKeys[docKey]
	if !ok {
		return "", fmt.Errorf("document %q: link %q names unknown document %q", docName(d), link, docKey)
	}
	target := &v.documents[j]
	if id, ok := v.buildRefIndex(target)[key]; ok {
		return qualifyRef(target, id), nil
	}
	return "", fmt.Errorf("document %q: link %q does not resolve", docName(d), link)
}

// qualifyRef renders a cross-document trace ref. reqmd resolves
// "<document-id>/<id>", falling back to "<dir>/<id>", so the emitted ref uses
// the target's document-id when it has one and its directory otherwise.
func qualifyRef(target *Document, id string) string {
	if target.XReqmd != nil && target.XReqmd.DocumentID != "" {
		return target.XReqmd.DocumentID + "/" + id
	}
	if dir := target.Dir(); dir != "" {
		return dir + "/" + id
	}
	return id
}

func (v *validator) renderSchema(idx int, d *Document) (string, error) {
	m := map[string]any{
		"$schema":              "https://json-schema.org/draft/2020-12/schema",
		"type":                 "object",
		"additionalProperties": false,
	}
	if len(d.Properties) > 0 {
		m["properties"] = d.Properties
	}
	if len(d.Required) > 0 {
		m["required"] = d.Required
	}

	x := map[string]any{}
	if d.XReqmd != nil {
		xr := d.XReqmd
		if xr.Level != "" {
			x["level"] = xr.Level
		}
		if xr.DocumentID != "" {
			x["document-id"] = xr.DocumentID
		}
		if xr.IDPrefix != "" {
			x["id-prefix"] = xr.IDPrefix
		}
		if xr.External {
			x["external"] = true
		}
		if len(xr.RequiresTraceFrom) > 0 {
			x["requires-trace-from"] = xr.RequiresTraceFrom
		}
		if xr.Upstream != nil && (xr.Upstream.Level != "" || len(xr.Upstream.Sources) > 0) {
			u := map[string]any{}
			if xr.Upstream.Level != "" {
				u["level"] = xr.Upstream.Level
			}
			if len(xr.Upstream.Sources) > 0 {
				u["sources"] = xr.Upstream.Sources
			}
			x["upstream"] = u
		}
	}
	if extra := v.additions[idx]; len(extra) > 0 {
		seen := map[string]struct{}{"draft": {}, "approved": {}, "rejected": {}, "deferred": {}}
		out := make([]string, 0, len(extra))
		for _, s := range extra {
			if _, dup := seen[s]; dup {
				continue
			}
			seen[s] = struct{}{}
			out = append(out, s)
		}
		sort.Strings(out)
		x["additional-status-values"] = out
	}
	if len(x) > 0 {
		m["x-reqmd"] = x
	}

	out, err := yaml.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("render schema for %q: %w", docName(d), err)
	}
	return string(out), nil
}

// add appends a rendered file, rejecting a second document claiming the same
// output path.
func add(files *[]writer.File, written map[string]string, dir, name string, data []byte) error {
	rel := path.Join(dir, name)
	if prev, dup := written[rel]; dup {
		return fmt.Errorf("output path %q is claimed by both %q and %q", rel, prev, dir)
	}
	written[rel] = dir
	*files = append(*files, writer.File{Path: rel, Data: data})
	return nil
}

func docName(d *Document) string {
	if d.Key == "" {
		return "(unnamed)"
	}
	return d.Key
}

// Dir is the output directory, from x-reqmd.file's sibling config. It defaults
// to the document key so a neutral producer needs no layout knowledge.
func (d *Document) Dir() string {
	if d.XReqmd != nil && d.XReqmd.Dir != "" {
		return d.XReqmd.Dir
	}
	return d.Key
}

// FileName is the .md filename, defaulting to requirements.md.
func (d *Document) FileName() string {
	if d.XReqmd != nil && d.XReqmd.File != "" {
		return d.XReqmd.File
	}
	return "requirements.md"
}

// prefix is the ID prefix, empty for a document without x-reqmd.
func (d *Document) prefix() string {
	if d.XReqmd == nil {
		return ""
	}
	return d.XReqmd.IDPrefix
}

// itemLabel is how an item is addressed in an error message: its title when it
// has one (that is what a producer or reviewer recognises), else its key, else
// a placeholder for a headingless block.
func itemLabel(it *Item) string {
	if it.Title != "" {
		return it.Title
	}
	if it.Key != "" {
		return it.Key
	}
	return "(untitled)"
}

// cutDocRef splits a "doc-key/key" link. A bare key is document-local.
func cutDocRef(link string) (string, string, bool) {
	before, after, ok := strings.Cut(link, "/")
	if !ok || before == "" || after == "" {
		return "", link, false
	}
	return before, after, true
}

// checkSingleLine rejects a producer string that would break the one-line
// structure of a heading or document title: a line break injects extra Markdown
// lines and can steal the attr block of the next item.
func checkSingleLine(field, value string) error {
	if strings.ContainsAny(value, "\n\r") {
		return fmt.Errorf("%s must not contain a line break", field)
	}
	return nil
}

// checkAttrKey rejects an attribute name that cannot be written as "key: value".
// reqmd's attr scanner splits each line at the first colon and never unquotes
// keys, so a key containing a colon would silently arrive under a different
// name ("a:b: v" is read as attribute "a").
func checkAttrKey(name string) error {
	if name == "" {
		return fmt.Errorf("attribute name must not be empty")
	}
	for _, r := range name {
		switch {
		case r == '_' || r == '-' || r == '.':
		case r >= '0' && r <= '9':
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		default:
			return fmt.Errorf("attribute %q: name may only contain letters, digits, '_', '-' and '.'", name)
		}
	}
	return nil
}

// checkScalarValue rejects an attribute value the renderer cannot write: reqmd
// attr blocks hold scalars and flat lists of scalars, never nested structures.
func checkScalarValue(name string, v any) error {
	switch val := v.(type) {
	case nil, string, bool, int, int64, float64, json.Number:
		return nil
	case []any:
		for _, item := range val {
			switch item.(type) {
			case nil, string, bool, int, int64, float64, json.Number:
			default:
				return fmt.Errorf("attribute %q: list items must be scalars, got %T", name, item)
			}
		}
		return nil
	case []string:
		return nil
	default:
		return fmt.Errorf("attribute %q: nested %T values are not supported", name, v)
	}
}

// cleanRelDir normalizes a document directory and rejects paths that escape the
// output root.
func cleanRelDir(dir string) (string, error) {
	if dir == "" {
		return "", nil
	}
	c := path.Clean(dir)
	if c == "." {
		return "", nil
	}
	if path.IsAbs(c) || c == ".." || len(c) > 3 && c[:3] == "../" {
		return "", fmt.Errorf("dir %q escapes the output root", dir)
	}
	return c, nil
}

// cleanFileName rejects a .md name that is not a plain file name in the
// document directory.
func cleanFileName(name string) (string, error) {
	if name == "" {
		return "requirements.md", nil
	}
	if err := checkSingleLine("file name", name); err != nil {
		return "", err
	}
	if path.Base(name) != name || name == "." || name == ".." {
		return "", fmt.Errorf("file %q must be a plain file name in the document directory", name)
	}
	return name, nil
}
