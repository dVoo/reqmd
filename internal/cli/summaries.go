package cli

import (
	"reqmd/internal/model"
	"reqmd/internal/reporter"
)

// buildDocSummaries converts parsed documents into a slice of DocumentSummary
// for use by list and stats commands. Rows contain every node — requirements
// and information items (containers + info blocks) — in document order.
func buildDocSummaries(docs []model.Document) []reporter.DocumentSummary {
	summaries := make([]reporter.DocumentSummary, 0, len(docs))
	for _, doc := range docs {
		summary := reporter.DocumentSummary{
			Path:       doc.Path,
			Properties: doc.Properties,
			Rows:       make([]reporter.NodeRow, 0, len(doc.Nodes)),
		}
		var walk func(nodes []*model.Node)
		walk = func(nodes []*model.Node) {
			for _, n := range nodes {
				summary.Rows = append(summary.Rows, reporter.NodeRow{
					Kind:  n.Kind,
					ID:    n.ID,
					Title: n.Title,
					Attrs: n.Attrs,
				})
				if n.Kind == model.KindRequirement {
					summary.ReqCount++
				} else {
					summary.ItemCount++
				}
				walk(n.Children)
			}
		}
		walk(doc.Nodes)
		summaries = append(summaries, summary)
	}
	return summaries
}
