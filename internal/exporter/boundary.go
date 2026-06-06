package exporter

import (
	"path/filepath"

	"reqmd/internal/model"
)

// DocBoundary describes a document directory's position in the V-model chain.
// IsRoot is true when the document has no upstream sources (or is
// declared external). IsLeaf is true when no other document references
// it as an upstream source.
type DocBoundary struct {
	IsRoot bool
	IsLeaf bool
}

// ComputeDocBoundaries returns the boundary for each document directory path.
// A document is a root boundary when it has no upstream sources, or is
// declared external. A document is a leaf boundary when it is not
// referenced as an upstream source by any other document.
func ComputeDocBoundaries(docs []model.Document) map[string]DocBoundary {
	referenced := make(map[string]bool)
	for _, d := range docs {
		if d.XReqmd == nil || d.XReqmd.Upstream == nil {
			continue
		}
		for _, src := range d.XReqmd.Upstream.Sources {
			// Sources are relative paths from the schema's location
			// (e.g. "../00-aspice/"); resolve against this doc's
			// directory to get the absolute path that matches doc.Path.
			referenced[filepath.Clean(filepath.Join(d.Path, src))] = true
		}
	}
	out := make(map[string]DocBoundary, len(docs))
	for _, d := range docs {
		b := DocBoundary{IsRoot: true, IsLeaf: true}
		if d.XReqmd != nil {
			if d.XReqmd.Upstream != nil && len(d.XReqmd.Upstream.Sources) > 0 {
				b.IsRoot = false
			}
			if d.XReqmd.External {
				b.IsRoot = true
			}
		}
		if referenced[d.Path] {
			b.IsLeaf = false
		}
		out[d.Path] = b
	}
	return out
}
