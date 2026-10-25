// Package exchange defines the reqmd-import/v1 JSON interchange format: the
// contract between producers (reqmd-code, third-party importer scripts) and
// the single renderer that turns it into reqmd documents (schema.yaml + .md).
//
// Producers emit the tool's own native, stable IDs (Doors object numbers,
// codeBeamer item IDs, code symbol hashes); this package stamps the reqmd IDs
// and resolves refs, so no producer has to know the markdown/YAML rules.
package exchange

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// Format is the value a bundle must declare in its "format" field.
const Format = "github.com/dVoo/reqmd/reqmd-import/v1"

// Bundle is the top-level exchange document: the list of reqmd document
// directories to render, in output order.
type Bundle struct {
	Format    string     `json:"format"`
	Documents []Document `json:"documents"`
}

// Document describes one reqmd document directory: one schema.yaml plus,
// when it has requirements, one .md file.
//
// Field order below is the reading order of the format, not a memory layout:
// this is a wire DTO decoded once per run. Reordering the fields (which is what
// govet's fieldalignment wants) also drops every field doc comment.
type Document struct {
	// Dir is the output directory relative to the output root. "" is the
	// root itself. Must not escape the root.
	Dir string `json:"dir,omitempty"`
	// File is the .md filename inside Dir. Defaults to "requirements.md".
	// Ignored when the document has no requirements (schema-only). Must be a
	// plain file name: no path separator, no "..".
	File string `json:"file,omitempty"`
	// Title is the document title (the level-1 heading). Defaults to
	// "Imported requirements". Must be a single line.
	Title string `json:"title,omitempty"`
	// Notice is an optional banner line rendered under the generated marker.
	// Must be a single line.
	Notice string `json:"notice,omitempty"`

	// IDPrefix is prepended to every requirement's native-id to form its reqmd
	// ID. Required when the document has requirements.
	IDPrefix string `json:"id-prefix"`
	// Level is x-reqmd.level.
	Level string `json:"level,omitempty"`
	// DocumentID is x-reqmd.document-id.
	DocumentID string `json:"document-id,omitempty"`
	// Upstream is x-reqmd.upstream.
	Upstream *Upstream `json:"upstream,omitempty"`
	// RequiresTraceFrom is x-reqmd.requires-trace-from.
	RequiresTraceFrom []string `json:"requires-trace-from,omitempty"`
	// External sets x-reqmd.external for proxy documents authored elsewhere.
	External bool `json:"external,omitempty"`

	// SchemaID, SchemaTitle and SchemaDescription are the JSON-Schema header
	// fields of the generated schema.yaml.
	SchemaID          string `json:"schema-id,omitempty"`
	SchemaTitle       string `json:"schema-title,omitempty"`
	SchemaDescription string `json:"schema-description,omitempty"`
	// Properties are the attr-block JSON-Schema property definitions;
	// Required lists the required ones. An attribute a requirement emits must
	// appear here, or the generated schema rejects it.
	Properties map[string]any `json:"properties,omitempty"`
	Required   []string       `json:"required,omitempty"`

	// Requirements are the document's requirements, in render order.
	Requirements []Requirement `json:"requirements,omitempty"`
}

// Upstream is a document's x-reqmd.upstream block.
type Upstream struct {
	Level   string   `json:"level,omitempty"`
	Sources []string `json:"sources,omitempty"`
}

// Requirement is one imported requirement, identified by the source tool's
// own stable ID. See Document for why fieldalignment is disabled here.
type Requirement struct {
	// NativeID is the producer's stable key for this item. The rendered reqmd
	// ID is <document id-prefix><native-id>. Must be unique across the bundle
	// (native-trace refs resolve by it) and must not contain whitespace: a
	// reqmd trace ref cannot match an ID with a space, so it would be
	// unreachable.
	NativeID string `json:"native-id"`
	// Title is the human-readable heading text. Empty renders the heading
	// without a title suffix. Must be a single line.
	Title string `json:"title,omitempty"`
	// Body is the markdown body following the attr block. Headings in it become
	// info items, which is intended; it must not put a fenced block whose info
	// string is "attr" directly under a heading, because that pair would make a
	// requirement of its own.
	Body string `json:"body,omitempty"`
	// Status is the reqmd built-in status. Defaults to "approved" so imported
	// items count as upstream coverage providers.
	Status string `json:"status,omitempty"`
	// Parent is the native-id of this requirement's parent in the same
	// document; it renders one heading level deeper.
	Parent string `json:"parent,omitempty"`
	// Trace holds fully-qualified reqmd IDs, passed through verbatim.
	Trace []string `json:"trace,omitempty"`
	// NativeTrace holds native-ids resolved against the bundle and rewritten to
	// <that document's id-prefix><native-id>.
	NativeTrace []string `json:"native-trace,omitempty"`
	// Attrs are the remaining attr-block entries. Names may contain only
	// letters, digits, '_', '-' and '.' (reqmd reads the block line by line and
	// would silently file a name containing ':' under a different attribute).
	// Values may be strings, numbers, booleans, or flat lists of those; a string
	// is always rendered as a string, even when it reads like a number or null.
	// The built-ins "status" and "trace" are reserved.
	Attrs map[string]any `json:"attrs,omitempty"`
}

// Decode reads one exchange bundle. Unknown fields are rejected so a producer
// typo fails loudly instead of silently dropping data.
func Decode(r io.Reader) (*Bundle, error) {
	dec := json.NewDecoder(r)
	dec.UseNumber()
	dec.DisallowUnknownFields()
	var b Bundle
	if err := dec.Decode(&b); err != nil {
		return nil, fmt.Errorf("decode exchange bundle: %w", err)
	}
	if err := dec.Decode(new(json.RawMessage)); err == nil {
		return nil, errors.New("decode exchange bundle: trailing data after the JSON document")
	} else if !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("decode exchange bundle: %w", err)
	}
	return &b, nil
}
