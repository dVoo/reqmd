package exporter

import "reqmd/internal/graph"

// VerdictInfoFromGraph converts a graph-side rolled-up NodeVerdict into the
// exporter's render type, carrying the outcome, representative source file,
// the contributing case keys (for CSV), and the evidence items (for HTML).
func VerdictInfoFromGraph(nv graph.NodeVerdict) VerdictInfo {
	vi := VerdictInfo{Outcome: nv.Outcome, Source: nv.Source}
	for _, it := range nv.Evidence {
		if it.Case != "" {
			vi.Cases = append(vi.Cases, it.Case)
		}
		vi.Evidence = append(vi.Evidence, EvidenceItem{
			Case:        it.Case,
			Outcome:     it.Outcome,
			File:        it.File,
			Description: it.Description,
		})
	}
	return vi
}
