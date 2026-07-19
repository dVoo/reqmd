package exporter

import (
	"encoding/csv"
	"fmt"
	"io"

	"reqmd/internal/model"
)

type CSV struct {
	verdicts map[string]VerdictInfo
}

// SetVerdicts stores verification verdicts for CSV export. When non-nil,
// two extra columns (Verdict, Verdict Source) are appended to the output.
func (c *CSV) SetVerdicts(v map[string]VerdictInfo) {
	c.verdicts = v
}
func (c *CSV) Export(w io.Writer, doc model.Document, propOrder []string) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()

	// Header: ID, Title, properties..., Body, Rationale, [Verdict, Verdict Source]
	hasVerdicts := c.verdicts != nil
	header := make([]string, 0, 3+len(propOrder)+2+2)
	header = append(header, "ID", "Title")
	header = append(header, propOrder...)
	header = append(header, "Body", "Rationale")
	if hasVerdicts {
		header = append(header, "Verdict", "Verdict Source")
	}
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
		if hasVerdicts {
			if v, ok := c.verdicts[req.ID]; ok {
				row = append(row, v.Outcome, v.Source)
			} else {
				row = append(row, "", "")
			}
		}
		if err := cw.Write(row); err != nil {
			return fmt.Errorf("writing CSV row for %s: %w", req.ID, err)
		}
	}
	return nil
}
