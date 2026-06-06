package exporter

import (
	"io"

	"reqmd/internal/model"
)

// Exporter writes parsed documents to a writer.
type Exporter interface {
	Export(w io.Writer, doc model.Document, propOrder []string) error
}
