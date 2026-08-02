package model

// Requirement is a single parsed requirement from a Markdown document.
type Requirement struct {
	ID           string // from the ## heading (before optional colon+title)
	Title        string // optional title text after "ID: " in heading
	ParentID     string // set from heading hierarchy when this is a sub-requirement
	Attrs        map[string]any
	Body         string   // prose after the attr block
	Rationale    string   // optional *Rationale:* text, if present
	Source       string   // filepath for error messages
	Suppressions []string // check names to suppress, from reqmd-suppress attr
}

// XReqmd holds reqmd-specific directory metadata from schema.yaml.
// All fields map directly to the x-reqmd extension block.
type XReqmd struct {
	Level                string         `yaml:"level,omitempty" json:"level,omitempty"`
	DocumentID           string         `yaml:"document-id,omitempty" json:"document-id,omitempty"`
	Upstream             *TraceUpstream `yaml:"upstream,omitempty" json:"upstream,omitempty"`
	MandatoryDisposition bool           `yaml:"mandatory-disposition,omitempty" json:"mandatory-disposition,omitempty"`
	External             bool           `yaml:"external,omitempty" json:"external,omitempty"`
	URL                  string         `yaml:"url,omitempty" json:"url,omitempty"`
	Source               *SourceConfig  `yaml:"source,omitempty" json:"source,omitempty"`
	IDPrefix             string         `yaml:"id-prefix,omitempty" json:"id-prefix,omitempty"`
	// AdditionalStatusValues extends the built-in status enum with extra
	// lowercase values, e.g. `additional-status-values: [review]`. Validated
	// by internal/schema.Compile; this field is populated from
	// `x-reqmd.additional-status-values` in schema.yaml.
	AdditionalStatusValues []string `yaml:"additional-status-values,omitempty" json:"additional-status-values,omitempty"`
	// IgnoreStatus opts the document out of the status lifecycle. All
	// requirements in an ignore-status directory count as coverage
	// providers regardless of their `status` value.
	IgnoreStatus bool `yaml:"ignore-status,omitempty" json:"ignore-status,omitempty"`
	// DisjointCheck names one or more array-typed attributes whose values
	// must overlap between every trace-linked source and target requirement.
	// Populated from `x-reqmd.disjoint-check` in schema.yaml (string or
	// array form). Empty means no disjoint check for this document.
	DisjointCheck []string `yaml:"disjoint-check,omitempty" json:"disjoint-check,omitempty"`
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

// Document groups all requirements sharing a single schema.yaml.
type Document struct {
	Path         string // document directory path
	Schema       any    // parsed schema.yaml (map[string]any)
	Requirements []Requirement
	XReqmd       *XReqmd        // parsed x-reqmd block, nil if absent
	Properties   []string       // property order: required first, then optional (schema definition order)
	Meta         map[string]any // YAML frontmatter merged across .md files in this document
}
