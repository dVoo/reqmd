// Package exchange defines the reqmd exchange format v1: the JSON contract
// between producers (reqmd-code, third-party importer scripts) and the single
// renderer that turns it into reqmd documents (schema.yaml + .md).
//
// The format is neutral by design. A producer supplies its own stable keys and
// its own vocabulary; every reqmd-specific decision (ID prefixes, V-model
// levels, upstream wiring) lives in the document's x-reqmd block, which a
// producer that is not reqmd-aware can omit entirely.
package exchange

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// SchemaURL is the value a bundle must declare in its "$schema" field. It is
// also the identity of the format: a breaking change moves to a new URL
// instead of growing an in-place version.
const SchemaURL = "https://reqmd.dev/exchange/v1/schema.json"

// Item kinds. A requirement renders a heading plus a reqmd block and takes part
// in validation, tracing and coverage; a section renders structure only.
const (
	KindRequirement = "requirement"
	KindSection     = "section"
)

// maxDepth bounds bundle nesting at decode time. The renderer separately
// rejects anything deeper than six heading levels; this guard only keeps a
// hostile bundle from exhausting the stack.
const maxDepth = 32

// Bundle is the top-level exchange document: the documents to render, in
// output order.
type Bundle struct {
	Schema    string     `json:"$schema"`
	Documents []Document `json:"documents"`
}

// Document describes one reqmd document directory: one schema.yaml plus, when
// it has items, one .md file.
type Document struct {
	// Key is the document's stable identity, unique across the bundle, without
	// "/" or whitespace. It qualifies cross-document links.
	Key string `json:"key"`
	// Title is the document title (the level-1 heading). Required.
	Title string `json:"title"`
	// Body is preamble prose rendered under the title, before the first item.
	Body string `json:"body,omitempty"`
	// Items is the document content, in document order.
	Items []Item `json:"items,omitempty"`
	// Properties and Required describe the reqmd-block JSON Schema for
	// requirements. Sections never carry fields, so both apply to requirements
	// only.
	Properties map[string]any `json:"properties,omitempty"`
	Required   []string       `json:"required,omitempty"`
	// XReqmd holds the renderer-specific configuration.
	XReqmd *XReqmd `json:"x-reqmd,omitempty"`
}

// XReqmd is every reqmd-specific decision for one document. A producer that
// only knows the neutral part of the format omits it.
type XReqmd struct {
	// IDPrefix is prepended to every requirement key to form the reqmd ID.
	// Without it the ID is the bare key.
	IDPrefix string `json:"id-prefix,omitempty"`
	// Level is x-reqmd.level.
	Level string `json:"level,omitempty"`
	// DocumentID is x-reqmd.document-id; it qualifies emitted trace refs.
	DocumentID string `json:"document-id,omitempty"`
	// Upstream is x-reqmd.upstream.
	Upstream *Upstream `json:"upstream,omitempty"`
	// RequiresTraceFrom is x-reqmd.requires-trace-from.
	RequiresTraceFrom []string `json:"requires-trace-from,omitempty"`
	// External marks a mirror of content authored elsewhere.
	External bool `json:"external,omitempty"`
	// Dir is the output directory relative to the output root; it defaults to
	// the document key. It must not escape the root.
	Dir string `json:"dir,omitempty"`
	// File is the .md filename inside the document directory.
	File string `json:"file,omitempty"`
}

// Upstream is a document's x-reqmd.upstream block.
type Upstream struct {
	Level   string   `json:"level,omitempty"`
	Sources []string `json:"sources,omitempty"`
}

// Item is one content item. Nesting replaces a parent pointer entirely: a child
// cannot reference a missing parent, cannot be listed before it, and cannot
// form a cycle.
type Item struct {
	// Kind is "requirement" or "section". Mandatory.
	Kind string `json:"kind"`
	// Key is the producer's stable identity, unique within the document and
	// without whitespace. Required on requirements; sections are linked to by
	// position, never by key.
	Key string `json:"key,omitempty"`
	// Title is the heading text. Omit it for a headingless prose block (the
	// document preamble, or an untitled section).
	Title string `json:"title,omitempty"`
	// Body is Markdown prose. It must not contain headings.
	Body string `json:"body,omitempty"`
	// Fields are the remaining reqmd-block entries. Requirements only.
	Fields map[string]any `json:"fields,omitempty"`
	// Status is the source system's own status value; the renderer maps it.
	Status string `json:"status,omitempty"`
	// Links names other items by key ("17" in this document, "brakes/17"
	// across documents). Every entry must resolve to a requirement.
	Links []string `json:"links,omitempty"`
	// Refs are external requirement IDs passed through verbatim. They are never
	// resolved or validated.
	Refs []string `json:"refs,omitempty"`
	// Children are nested items, in document order.
	Children []Item `json:"children,omitempty"`
}

// Decode reads one exchange bundle. Unknown fields are rejected so a producer
// typo fails loudly instead of silently dropping data, and nesting deeper than
// maxDepth is refused before it can exhaust the stack.
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
	for i := range b.Documents {
		if err := checkNesting(b.Documents[i].Items, 1); err != nil {
			return nil, fmt.Errorf("document %q: %w", b.Documents[i].Key, err)
		}
	}
	return &b, nil
}

// checkNesting rejects a bundle nested deeper than maxDepth.
func checkNesting(items []Item, depth int) error {
	if depth > maxDepth {
		return fmt.Errorf("nesting exceeds %d levels", maxDepth)
	}
	for i := range items {
		if err := checkNesting(items[i].Children, depth+1); err != nil {
			return err
		}
	}
	return nil
}
