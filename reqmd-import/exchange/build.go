package exchange

import (
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/dVoo/reqmd/reqmd-import/internal/writer"

	"gopkg.in/yaml.v3"
)

// Build validates the bundle and renders every document to output files. A
// document without requirements produces only its schema.yaml; one with
// requirements also produces its .md file. Requirements keep bundle order.
func (b *Bundle) Build() ([]writer.File, error) {
	if b.Format != Format {
		return nil, fmt.Errorf("unsupported format %q (want %q)", b.Format, Format)
	}
	if len(b.Documents) == 0 {
		return nil, fmt.Errorf("bundle has no documents")
	}

	// Global native-id -> reqmd ID index. native-trace refs resolve through
	// it, so native ids must be unique across the bundle. This pass is also
	// where producer strings that would corrupt a document are rejected.
	ids := map[string]string{}
	for i := range b.Documents {
		d := &b.Documents[i]
		if len(d.Requirements) > 0 && d.IDPrefix == "" {
			return nil, fmt.Errorf("document %q: id-prefix is required when the document has requirements", docName(d))
		}
		if err := checkSingleLine("document title", d.Title); err != nil {
			return nil, fmt.Errorf("document %q: %w", docName(d), err)
		}
		if err := checkSingleLine("document notice", d.Notice); err != nil {
			return nil, fmt.Errorf("document %q: %w", docName(d), err)
		}
		for j := range d.Requirements {
			r := &d.Requirements[j]
			if r.NativeID == "" {
				return nil, fmt.Errorf("document %q: requirement %d has no native-id", docName(d), j)
			}
			if _, dup := ids[r.NativeID]; dup {
				return nil, fmt.Errorf("document %q: native-id %q is not unique across the bundle", docName(d), r.NativeID)
			}
			id := d.IDPrefix + r.NativeID
			ids[r.NativeID] = id

			if err := checkID(id); err != nil {
				return nil, err
			}
			if err := checkSingleLine("requirement title", r.Title); err != nil {
				return nil, fmt.Errorf("requirement %q: %w", id, err)
			}
			if err := checkBody(r.Body); err != nil {
				return nil, fmt.Errorf("requirement %q: %w", id, err)
			}
			for name := range r.Attrs {
				if err := checkAttrKey(name); err != nil {
					return nil, fmt.Errorf("requirement %q: %w", id, err)
				}
			}
		}
	}

	var files []writer.File
	written := map[string]string{}
	for i := range b.Documents {
		d := &b.Documents[i]
		dir, err := cleanRelDir(d.Dir)
		if err != nil {
			return nil, fmt.Errorf("document %q: %w", docName(d), err)
		}
		if err := checkFileName(d.File); err != nil {
			return nil, fmt.Errorf("document %q: %w", docName(d), err)
		}

		// Parent refs are document-local: heading nesting is per-document.
		localIDs := make(map[string]string, len(d.Requirements))
		for j := range d.Requirements {
			r := &d.Requirements[j]
			localIDs[r.NativeID] = d.IDPrefix + r.NativeID
		}
		parentOf := map[string]string{}
		for j := range d.Requirements {
			r := &d.Requirements[j]
			if r.Parent == "" {
				continue
			}
			pid, ok := localIDs[r.Parent]
			if !ok {
				return nil, fmt.Errorf("requirement %q: parent %q is not in the same document", d.IDPrefix+r.NativeID, r.Parent)
			}
			parentOf[d.IDPrefix+r.NativeID] = pid
		}

		reqs := make([]writer.Req, 0, len(d.Requirements))
		for j := range d.Requirements {
			r := &d.Requirements[j]
			req, err := buildReq(d, r, ids)
			if err != nil {
				return nil, err
			}
			reqs = append(reqs, req)
		}
		reqs, err = orderByDepth(reqs, parentOf)
		if err != nil {
			return nil, err
		}

		schema, err := renderSchema(d)
		if err != nil {
			return nil, err
		}
		if err := add(&files, written, dir, "schema.yaml", []byte(writer.GeneratedMarker+"\n"+schema)); err != nil {
			return nil, err
		}
		if len(reqs) > 0 {
			name := d.File
			if name == "" {
				name = "requirements.md"
			}
			title := d.Title
			if title == "" {
				title = "Imported requirements"
			}
			md, err := writer.RenderDocument(title, d.Notice, reqs)
			if err != nil {
				return nil, err
			}
			if err := add(&files, written, dir, name, []byte(md)); err != nil {
				return nil, err
			}
		}
	}
	return files, nil
}

func buildReq(d *Document, r *Requirement, ids map[string]string) (writer.Req, error) {
	req := writer.Req{
		ID:     d.IDPrefix + r.NativeID,
		Title:  r.Title,
		Status: r.Status,
		Body:   r.Body,
		Attrs:  r.Attrs,
		Trace:  append([]string(nil), r.Trace...),
	}
	if req.Status == "" {
		req.Status = "approved"
	}
	for _, nt := range r.NativeTrace {
		id, ok := ids[nt]
		if !ok {
			return writer.Req{}, fmt.Errorf("requirement %q: native-trace %q is not a native-id in this bundle", req.ID, nt)
		}
		req.Trace = append(req.Trace, id)
	}
	for _, reserved := range []string{"status", "trace"} {
		if _, ok := r.Attrs[reserved]; ok {
			return writer.Req{}, fmt.Errorf("requirement %q: attr %q is reserved (use the dedicated field)", req.ID, reserved)
		}
	}
	return req, nil
}

// orderByDepth returns the requirements in input order but with every child
// moved directly after its parent and its heading depth filled in, so heading
// nesting is valid regardless of the order a producer listed them in.
// parentOf maps a requirement ID to its parent's ID; cycles are rejected
// because they would drop requirements from the output.
func orderByDepth(reqs []writer.Req, parentOf map[string]string) ([]writer.Req, error) {
	children := map[string][]writer.Req{}
	var roots []writer.Req
	for _, r := range reqs {
		parent, isChild := parentOf[r.ID]
		if !isChild {
			roots = append(roots, r)
			continue
		}
		children[parent] = append(children[parent], r)
	}

	out := make([]writer.Req, 0, len(reqs))
	var emit func(writer.Req, int)
	emit = func(r writer.Req, depth int) {
		r.Depth = depth
		out = append(out, r)
		for _, c := range children[r.ID] {
			emit(c, depth+1)
		}
	}
	for _, r := range roots {
		emit(r, 0)
	}
	if len(out) != len(reqs) {
		emitted := make(map[string]bool, len(out))
		for _, r := range out {
			emitted[r.ID] = true
		}
		for _, r := range reqs {
			if !emitted[r.ID] {
				return nil, fmt.Errorf("requirement %q: parent links contain a cycle", r.ID)
			}
		}
	}
	return out, nil
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

// cleanRelDir normalizes a document directory and rejects paths that escape
// the output root.
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

func docName(d *Document) string {
	if d.Dir == "" {
		return d.DocumentID + " (root)"
	}
	return d.Dir
}

// checkSingleLine rejects a producer string that would break the one-line
// structure of a heading, document title, or notice. A line break is the
// dangerous case: it injects extra Markdown lines, and a "# ID" line followed
// by an attr fence would add a requirement of the producer's choosing — or
// steal the attr block of the requirement being rendered.
func checkSingleLine(field, value string) error {
	if strings.ContainsAny(value, "\n\r") {
		return fmt.Errorf("%s must not contain a line break", field)
	}
	return nil
}

// checkID rejects a requirement ID that reqmd can never reference: a trace ref
// must match reqmd's ID pattern, which allows no whitespace. Catching it here
// names the producer's native-id instead of failing later on somebody's trace.
func checkID(id string) error {
	if strings.ContainsAny(id, " \t\n\r\v\f") {
		return fmt.Errorf("requirement id %q must not contain whitespace (check id-prefix and native-id)", id)
	}
	return nil
}

// checkBody rejects a body that would become a requirement. A body is Markdown
// by design: headings in it turn into info items, which is a documented
// feature, and code examples are ordinary content. What is not fine is an attr
// fence directly under a heading — that pair is exactly what makes a
// requirement, so an imported description field could silently add
// requirements or re-key one that already exists.
func checkBody(body string) error {
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) != "```attr" {
			continue
		}
		for j := i - 1; j >= 0; j-- {
			prev := strings.TrimSpace(lines[j])
			if prev == "" {
				continue
			}
			if isATXHeading(prev) {
				return errors.New("body must not put an ```attr fence directly under a heading (that pair would become a requirement)")
			}
			break
		}
	}
	return nil
}

// isATXHeading reports whether a trimmed Markdown line is an ATX heading of
// level 2 or deeper. A level-1 heading is the document title and cannot carry
// an attr block, so it cannot inject a requirement.
func isATXHeading(line string) bool {
	hashes := 0
	for hashes < len(line) && line[hashes] == '#' {
		hashes++
	}
	return hashes >= 2 && hashes < len(line) && line[hashes] == ' '
}

// checkAttrKey rejects an attribute name that cannot be written as
// "key: value". reqmd's attr scanner splits each line at the first colon and
// never unquotes keys, so a key containing a colon would silently arrive under
// a different name ("a:b: v" is read as attribute "a").
func checkAttrKey(name string) error {
	if name == "" {
		return errors.New("attribute name must not be empty")
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

// checkFileName rejects a document .md name that is not a plain file name in
// the document directory. Without this a producer could write outside the
// output root ("../../outside/x.md").
func checkFileName(name string) error {
	if name == "" {
		return nil
	}
	if err := checkSingleLine("file name", name); err != nil {
		return err
	}
	if path.Base(name) != name || name == "." || name == ".." {
		return fmt.Errorf("file %q must be a plain file name in the document directory", name)
	}
	return nil
}

func renderSchema(d *Document) (string, error) {
	m := map[string]any{
		"$schema":              "https://json-schema.org/draft/2020-12/schema",
		"type":                 "object",
		"additionalProperties": false,
	}
	if d.SchemaID != "" {
		m["$id"] = d.SchemaID
	}
	if d.SchemaTitle != "" {
		m["title"] = d.SchemaTitle
	}
	if d.SchemaDescription != "" {
		m["description"] = d.SchemaDescription
	}
	if len(d.Properties) > 0 {
		m["properties"] = d.Properties
	}
	if len(d.Required) > 0 {
		m["required"] = d.Required
	}

	x := map[string]any{}
	if d.Level != "" {
		x["level"] = d.Level
	}
	if d.DocumentID != "" {
		x["document-id"] = d.DocumentID
	}
	if d.IDPrefix != "" {
		x["id-prefix"] = d.IDPrefix
	}
	if d.External {
		x["external"] = true
	}
	if len(d.RequiresTraceFrom) > 0 {
		x["requires-trace-from"] = d.RequiresTraceFrom
	}
	if d.Upstream != nil && (d.Upstream.Level != "" || len(d.Upstream.Sources) > 0) {
		u := map[string]any{}
		if d.Upstream.Level != "" {
			u["level"] = d.Upstream.Level
		}
		if len(d.Upstream.Sources) > 0 {
			u["sources"] = d.Upstream.Sources
		}
		x["upstream"] = u
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
