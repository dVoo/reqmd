// Package model defines the content tree for a parsed reqmd document
// (requirements, containers, and info nodes), the built-in attribute
// keys, and the shared helpers that graph checks and exporters operate on.
package model

// Kind discriminates the node types in a document's content tree.
type Kind int

const (
	// KindRequirement is a requirement: a heading followed by an attr
	// block. It is the zero value so existing Node literals that omit
	// Kind remain requirements.
	KindRequirement Kind = iota
	// KindContainer is a heading without an attr block that has child
	// nodes — a folder-like grouping of requirements and other items.
	KindContainer
	// KindInfo is a heading without an attr block and no children, or a
	// prose block that precedes any heading.
	KindInfo
)

// String returns the lowercase display name used in text and JSON output.
func (k Kind) String() string {
	switch k {
	case KindRequirement:
		return "req"
	case KindContainer:
		return "container"
	case KindInfo:
		return "info"
	default:
		return ""
	}
}

// Node is a single content node in a parsed Markdown document. Every
// heading becomes a node; a node is a requirement iff it has an adjacent
// attr block. Prose and other blocks attach to the nearest open heading's
// Body. Nodes form a tree via Children, preserving document order.
type Node struct {
	Attrs        map[string]any
	ID           string
	Title        string
	ParentID     string
	Body         string
	Rationale    string
	Source       string
	Suppressions []string
	Children     []*Node
	Kind         Kind
	Level        int
}

// IsRequirement reports whether the node is a requirement (as opposed to
// a container or info item).
func (n *Node) IsRequirement() bool {
	return n.Kind == KindRequirement
}

// XReqmd holds reqmd-specific directory metadata from schema.yaml.
// All fields map directly to the x-reqmd extension block.
type XReqmd struct {
	Upstream               *TraceUpstream `yaml:"upstream,omitempty" json:"upstream,omitempty"`
	Source                 *SourceConfig  `yaml:"source,omitempty" json:"source,omitempty"`
	Level                  string         `yaml:"level,omitempty" json:"level,omitempty"`
	DocumentID             string         `yaml:"document-id,omitempty" json:"document-id,omitempty"`
	URL                    string         `yaml:"url,omitempty" json:"url,omitempty"`
	IDPrefix               string         `yaml:"id-prefix,omitempty" json:"id-prefix,omitempty"`
	AdditionalStatusValues []string       `yaml:"additional-status-values,omitempty" json:"additional-status-values,omitempty"`
	DisjointCheck          []string       `yaml:"disjoint-check,omitempty" json:"disjoint-check,omitempty"`
	RequiresTraceFrom      []string       `yaml:"requires-trace-from,omitempty" json:"requires-trace-from,omitempty"`
	MandatoryDisposition   bool           `yaml:"mandatory-disposition,omitempty" json:"mandatory-disposition,omitempty"`
	External               bool           `yaml:"external,omitempty" json:"external,omitempty"`
	IgnoreStatus           bool           `yaml:"ignore-status,omitempty" json:"ignore-status,omitempty"`
}

// TraceUpstream declares the expected upstream layer and source directories.
type TraceUpstream struct {
	Level   string   `yaml:"level,omitempty" json:"level,omitempty"`
	Sources []string `yaml:"sources,omitempty" json:"sources,omitempty"`
}

// SourceConfig describes where to find and how to parse the originating artefact.
type SourceConfig struct {
	Path   string `yaml:"path,omitempty" json:"path,omitempty"`
	Format string `yaml:"format,omitempty" json:"format,omitempty"`
}

// Document groups all content sharing a single schema.yaml. Nodes holds
// the ordered content tree; every heading in every .md file is a node
// except the level-1 document title (see parser). Title is the first
// file's h1 heading, when present.
//
// Synthetic marks documents that do not originate from authored Markdown
// on disk — currently the pseudo-document of verification results
// synthesized by internal/verify, later also synthesized test cases. Nodes
// of a synthetic document carry no schema, no ID prefix, and no source file
// that tooling may rewrite (e.g. repin must skip them).
type Document struct {
	Schema     any
	XReqmd     *XReqmd
	Meta       map[string]any
	Path       string
	Title      string
	Nodes      []*Node
	Properties []string
	Synthetic  bool
}

// Requirements returns all requirement-kind nodes in document order
// (depth-first). The returned pointers reference the tree itself, so
// in-place mutation updates the document.
func (d *Document) Requirements() []*Node {
	return CollectRequirements(d.Nodes)
}

// CollectRequirements walks a node tree depth-first and returns all
// requirement-kind nodes in document order.
func CollectRequirements(nodes []*Node) []*Node {
	var out []*Node
	var walk func(n *Node)
	walk = func(n *Node) {
		if n.Kind == KindRequirement {
			out = append(out, n)
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	for _, n := range nodes {
		walk(n)
	}
	return out
}
