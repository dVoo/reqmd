package writer

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"reqmd-import/internal/model"
)

// Package groups GeneratedReqs by their source package.
type Package struct {
	Name string
	// Reqs is the flat list of requirements (top-level symbols + nested methods).
	Reqs []model.GeneratedReq
}

// standardSchema is the canonical schema.yaml content, embedded at compile time
// from standard_schema.yaml (in this directory). The canonical source-of-truth
// for the schema content lives in internal/schema/standard_schema.yaml — this
// package keeps a local copy so //go:embed (which cannot cross package
// boundaries) can pick it up.
//
//go:embed standard_schema.yaml
var standardSchema string

// WriteTarget writes the standard schema.yaml once at targetDir and
// per-package .md files at targetDir/<package>/package.md.
//
// The idPrefix argument is the global --id-prefix flag (e.g. "IMP-"). The
// per-package schema written into each subdirectory declares a *package-
// specific* prefix "<idPrefix><pkgName>-" so that reqmd's graph checker
// sees distinct prefixes for every package and does not flag an ID-PREFIX
// collision. The IDs in package.md already embed the package name (built
// by the extract subcommand as "<idPrefix><pkgName>-<name>-<hash>"), so
// the per-package schema's prefix matches the actual ID prefix exactly.
//
// Idempotent: if a file already exists with the same content, the write is
// skipped. Existing files without the generated marker (i.e., hand-authored
// content) are left untouched so writer does not clobber manual edits.
func WriteTarget(targetDir string, packages []Package, idPrefix string) error {
	if err := writeSchema(targetDir); err != nil {
		return fmt.Errorf("write schema: %w", err)
	}
	for _, pkg := range packages {
		if err := writePackage(targetDir, pkg, idPrefix); err != nil {
			return fmt.Errorf("write package %s: %w", pkg.Name, err)
		}
	}
	return nil
}

func writeSchema(targetDir string) error {
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("mkdir target: %w", err)
	}
	dest := filepath.Join(targetDir, "schema.yaml")
	// Prepend the generated marker so writeIfChanged's marker check
	// recognizes this file as writer output on subsequent runs and
	// overwrites it when the embedded schema changes. Without the
	// marker, a re-run would mistake the schema for hand-authored
	// content and skip the write.
	return writeIfChanged(dest, []byte(GeneratedMarker+"\n"+standardSchema))
}

func writePackage(targetDir string, pkg Package, idPrefix string) error {
	pkgDir := filepath.Join(targetDir, pkg.Name)
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		return fmt.Errorf("mkdir package: %w", err)
	}
	// reqmd treats every directory containing a schema.yaml as an
	// independent document; subdirectories without a schema are not
	// recursed into. We therefore copy the standard schema into every
	// per-package subdir so `reqmd check <targetDir>` discovers each
	// package's package.md as a valid document.
	//
	// The copy is parsed as YAML, the per-package id-prefix is rewritten
	// under x-reqmd, and the result is re-marshaled. This is robust
	// against whitespace/comment/reorder changes in standardSchema; the
	// only invariant is that the schema stays a valid YAML mapping with
	// an x-reqmd block containing id-prefix.
	//
	// Note: yaml.Marshal of a map[string]any sorts keys alphabetically,
	// so the per-package output has a different top-level field order
	// than the canonical (hand-authored) standardSchema. That is
	// acceptable — the per-package output is consumed by the reqmd
	// schema validator, which does not care about field order, and
	// yaml.Marshal is deterministic so repeated runs produce
	// byte-identical output. The top-level writeSchema still embeds
	// the canonical byte-for-byte, so the schema-drift test is
	// unaffected.
	pkgSchema, err := renderPackageSchema(standardSchema, idPrefix, pkg.Name)
	if err != nil {
		return fmt.Errorf("render per-package schema: %w", err)
	}
	schemaDest := filepath.Join(pkgDir, "schema.yaml")
	// Same marker treatment as writeSchema: see writeSchema for why.
	if err := writeIfChanged(schemaDest, []byte(GeneratedMarker+"\n"+pkgSchema)); err != nil {
		return fmt.Errorf("write per-package schema: %w", err)
	}
	dest := filepath.Join(pkgDir, "package.md")
	return writeIfChanged(dest, []byte(RenderPackageMD(pkg)))
}

// renderPackageSchema returns a copy of the standard schema with the
// x-reqmd.id-prefix field rewritten to a package-specific value. The
// input is parsed as YAML, the x-reqmd.id-prefix field is overwritten,
// and the result is re-marshaled. This is robust against whitespace,
// comment, and reordering changes in the canonical schema; the only
// invariant is that the schema stays a valid YAML mapping with an
// x-reqmd block containing id-prefix.
//
// Note: yaml.Marshal of a map[string]any sorts keys alphabetically,
// so the rendered output has alphabetically-sorted top-level fields
// (e.g. $id, $schema, additionalProperties, description, properties,
// title, type, x-reqmd) regardless of the order in the source. This
// is acceptable for the per-package schema: the reqmd validator does
// not care about key order, and yaml.Marshal is deterministic so
// repeated runs produce byte-identical output.
func renderPackageSchema(schema, idPrefix, pkgName string) (string, error) {
	var doc map[string]any
	if err := yaml.Unmarshal([]byte(schema), &doc); err != nil {
		return "", fmt.Errorf("unmarshal standard schema: %w", err)
	}
	xReqmd, ok := doc["x-reqmd"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("standard schema missing x-reqmd block")
	}
	xReqmd["id-prefix"] = idPrefix + pkgName + "-"
	out, err := yaml.Marshal(doc)
	if err != nil {
		return "", fmt.Errorf("marshal per-package schema: %w", err)
	}
	return string(out), nil
}

// writeIfChanged writes data to path only when the file is missing, has the
// wrong content, or is missing the generated marker. Hand-authored files
// (no generated marker) are left alone.
func writeIfChanged(path string, data []byte) error {
	existing, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return os.WriteFile(path, data, 0o644)
	}
	if err != nil {
		return fmt.Errorf("read existing: %w", err)
	}
	if !IsGenerated(existing) {
		// Hand-authored content: do not overwrite.
		return nil
	}
	if bytes.Equal(existing, data) {
		return nil
	}
	return os.WriteFile(path, data, 0o644)
}

// RenderPackageMD returns the .md content for a package.
//
// Layout:
//
//	# Package: <name>
//	<generated marker>
//	<auto-gen notice>
//	(one ## heading + attr block + body per top-level requirement, with
//	 nested ### headings for child requirements such as struct methods)
func RenderPackageMD(pkg Package) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Package: %s\n\n", pkg.Name)
	b.WriteString(GeneratedMarker)
	b.WriteString("\n\n")
	b.WriteString("This file is auto-generated by reqmd-import. Do not edit by hand.\n\n")

	// Group children under their parent (methods inside their struct).
	// Stable, deterministic order: by ID, so repeated runs produce identical
	// output (important for the idempotent write-if-changed path).
	sorted := append([]model.GeneratedReq(nil), pkg.Reqs...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].ID < sorted[j].ID
	})

	var top []model.GeneratedReq
	children := make(map[string][]model.GeneratedReq)
	for _, r := range sorted {
		if r.ParentID == "" {
			top = append(top, r)
		} else {
			children[r.ParentID] = append(children[r.ParentID], r)
		}
	}

	for _, r := range top {
		b.WriteString(RenderRequirement(r))
		for _, c := range children[r.ID] {
			b.WriteString(RenderRequirement(c))
		}
	}
	return b.String()
}

// RenderRequirement returns the heading + attr block + body for a single
// requirement. The heading depth is `##` for top-level requirements and
// `###` for child requirements (ParentID set).
func RenderRequirement(req model.GeneratedReq) string {
	var b strings.Builder
	heading := "##"
	if req.ParentID != "" {
		heading = "###"
	}
	if req.Title != "" {
		fmt.Fprintf(&b, "%s %s: %s\n", heading, req.ID, req.Title)
	} else {
		fmt.Fprintf(&b, "%s %s\n", heading, req.ID)
	}

	b.WriteString("```attr\n")
	writeAttr(&b, req)
	b.WriteString("```\n")

	if req.Body != "" {
		b.WriteString("\n")
		b.WriteString(req.Body)
		b.WriteString("\n")
	}
	b.WriteString("\n")
	return b.String()
}

// writeAttr writes the YAML-ish content of the attr block. The format mirrors
// what the reqmd CLI consumes: simple `key: value` pairs, one per line, in a
// stable order. Quoting is applied only when the value contains characters
// that would confuse the YAML parser.
func writeAttr(b *strings.Builder, req model.GeneratedReq) {
	// Stable field order — keep this list in sync with the schema.
	writeKV(b, "status", req.Status)
	fmt.Fprintln(b, "x-reqmd.imported: true")
	if req.SourceFile != "" {
		fmt.Fprintln(b, "x-reqmd.source-file:", quoteIfNeeded(req.SourceFile))
	}
	if req.SourceLine > 0 {
		fmt.Fprintf(b, "x-reqmd.source-line: %d\n", req.SourceLine)
	}
	if req.SymbolKind != "" {
		fmt.Fprintln(b, "x-reqmd.symbol-kind:", quoteIfNeeded(req.SymbolKind))
	}
	if len(req.Trace) == 0 {
		fmt.Fprintln(b, "trace: []")
	} else {
		fmt.Fprintln(b, "trace:")
		for _, t := range req.Trace {
			fmt.Fprintf(b, "  - %s\n", quoteIfNeeded(t))
		}
	}
	// Heuristic traces go to a separate x-reqmd.heuristic-trace attribute
	// so the developer-asserted `trace:` list stays clean. Only render
	// when there are heuristic matches.
	if len(req.HeuristicTrace) > 0 {
		fmt.Fprintln(b, "x-reqmd.heuristic-trace:")
		for _, t := range req.HeuristicTrace {
			fmt.Fprintf(b, "  - %s\n", quoteIfNeeded(t))
		}
	}
}

// writeKV writes a single key/value pair, quoting the value when needed.
func writeKV(b *strings.Builder, key, value string) {
	fmt.Fprintf(b, "%s: %s\n", key, quoteIfNeeded(value))
}

// quoteIfNeeded quotes a value when it contains characters that would confuse
// the YAML parser. When quoting with double quotes, control characters are
// escaped to their two-character YAML sequences (\n, \t, \\, \") so the
// round-trip preserves the original string content.
func quoteIfNeeded(s string) string {
	if s == "" {
		return `""`
	}
	needsQuote := false
	for _, r := range s {
		if r == ':' || r == '#' || r == '"' || r == '\'' || r == '\n' || r == '\t' {
			needsQuote = true
			break
		}
	}
	leading := len(s) > 0 && (s[0] == ' ' || s[0] == '\t' || s[0] == '-' || s[0] == '*')
	trailing := len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t')
	if leading || trailing {
		needsQuote = true
	}
	if !needsQuote {
		return s
	}
	// Escape backslash first, then double-quote, then \n, \t. Backslash must
	// be first so the backslashes introduced by the later passes are not
	// themselves re-escaped.
	escaped := strings.ReplaceAll(s, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	escaped = strings.ReplaceAll(escaped, "\n", `\n`)
	escaped = strings.ReplaceAll(escaped, "\t", `\t`)
	return `"` + escaped + `"`
}
