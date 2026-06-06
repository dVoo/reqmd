package exporter

import (
	"encoding/csv"
	"fmt"
	"io"

	"reqmd/internal/model"
)

type CSV struct{}

func (CSV) Export(w io.Writer, doc model.Document, propOrder []string) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()

	// Header: ID, Title, properties..., Body, Rationale
	header := make([]string, 0, 3+len(propOrder)+2)
	header = append(header, "ID", "Title")
	header = append(header, propOrder...)
	header = append(header, "Body", "Rationale")
	if err := cw.Write(header); err != nil {
		return fmt.Errorf("writing CSV header: %w", err)
	}

	for _, req := range doc.Requirements {
		row := make([]string, 0, len(header))
		row = append(row, req.ID, req.Title)
		for _, prop := range propOrder {
			val := formatAttr(req.Attrs[prop])
			row = append(row, val)
		}
		row = append(row, req.Body, req.Rationale)
		if err := cw.Write(row); err != nil {
			return fmt.Errorf("writing CSV row for %s: %w", req.ID, err)
		}
	}
	return nil
}


