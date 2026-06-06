package cli

import (
	"reqmd/internal/model"
	"reqmd/internal/reporter"
)

// buildDocSummaries converts parsed documents into a slice of DocumentSummary
// for use by list and stats commands.
func buildDocSummaries(docs []model.Document) []reporter.DocumentSummary {
	summaries := make([]reporter.DocumentSummary, 0, len(docs))
	for _, doc := range docs {
		props := doc.Properties
		rows := make([]reporter.ReqRow, 0, len(doc.Requirements))
		for _, req := range doc.Requirements {
			rows = append(rows, reporter.ReqRow{
				ID:    req.ID,
				Title: req.Title,
				Attrs: req.Attrs,
			})
		}
		summaries = append(summaries, reporter.DocumentSummary{
			Path:       doc.Path,
			Properties: props,
			Rows:       rows,
		})
	}
	return summaries
}
