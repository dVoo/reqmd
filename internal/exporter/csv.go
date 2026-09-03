package exporter

import (
	"encoding/csv"
	"fmt"
	"io"
	"reqmd/internal/model"
)

// CSV exports document content as comma-separated values. When verdicts
// are set via SetVerdicts, two extra columns are appended per row.
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

	// Header: Type, ID, Title, properties..., Body, Rationale, [Verdict, Verdict Source]
	hasVerdicts := c.verdicts != nil
	header := make([]string, 0, 4+len(propOrder)+2+2)
	header = append(header, "Type", "ID", "Title")
	header = append(header, propOrder...)
	header = append(header, "Body", "Rationale")
	if hasVerdicts {
		header = append(header, "Verdict", "Verdict Source")
	}
	if err := cw.Write(header); err != nil {
		return fmt.Errorf("writing CSV header: %w", err)
	}

	// Rows are emitted in document order for every node: requirements and
	// information items (containers + info blocks) alike.
	var writeNode func(n *model.Node) error
	writeNode = func(n *model.Node) error {
		row := make([]string, 0, len(header))
		row = append(row, n.Kind.String())
		if n.Kind == model.KindRequirement {
			row = append(row, n.ID, n.Title)
		} else {
			row = append(row, "", n.Title)
		}
		for _, prop := range propOrder {
			val := ""
			if n.Kind == model.KindRequirement {
				val = formatAttr(n.Attrs[prop])
			}
			row = append(row, val)
		}
		row = append(row, n.Body, n.Rationale)
		if hasVerdicts {
			if v, ok := c.verdicts[n.ID]; ok {
				row = append(row, v.Outcome, v.Source)
			} else {
				row = append(row, "", "")
			}
		}
		if err := cw.Write(row); err != nil {
			return fmt.Errorf("writing CSV row for %s: %w", n.ID, err)
		}
		for _, child := range n.Children {
			if err := writeNode(child); err != nil {
				return err
			}
		}
		return nil
	}
	for _, n := range doc.Nodes {
		if err := writeNode(n); err != nil {
			return err
		}
	}
	return nil
}
