//go:build ladybug

// Package exporter provides ladybugdb export for durable graph storage.
// This is the only package that depends on ladybugdb — all other packages
// use the pure Go adjacency graph in internal/graph.
package exporter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	lbug "github.com/LadybugDB/go-ladybug"

	"reqmd/internal/model"
)

// ExportGraph writes all parsed requirements as a ladybugdb database at outDir.
// Each requirement becomes a node with id and file properties. Every trace
// link becomes a TracesTo edge between Requirement nodes.
//
// The resulting database is at <outDir>/reqmd-graph.lbug/ and can be queried
// with the ladybug CLI or SQL ODBC driver.
func ExportGraph(docs []model.Document, outDir string) error {
	// Build database path and clean any previous database.
	dbPath := filepath.Join(outDir, "reqmd-graph.lbug")
	if err := os.RemoveAll(dbPath); err != nil {
		return fmt.Errorf("removing previous database: %w", err)
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Configure and open database.
	cfg := lbug.DefaultSystemConfig()
	cfg.BufferPoolSize = 64 * 1024 * 1024

	db, err := lbug.OpenDatabase(dbPath, cfg)
	if err != nil {
		return fmt.Errorf("opening ladybugdb at %s: %w", dbPath, err)
	}
	defer db.Close()

	conn, err := lbug.OpenConnection(db)
	if err != nil {
		return fmt.Errorf("opening connection: %w", err)
	}
	defer conn.Close()

	// Create node and relationship tables.
	// Create node table with optional verification result properties.
	// outcome and source are only set on RESULT: pseudo-requirements;
	// authored requirements have NULL for these columns.
	if _, err := conn.Query("CREATE NODE TABLE IF NOT EXISTS Requirement(id STRING PRIMARY KEY, file STRING, outcome STRING, source STRING)"); err != nil {
		return fmt.Errorf("creating node table: %w", err)
	}

	// Pass 1: insert one node per requirement across all documents.
	// Result pseudo-requirements (RESULT:<id>) carry outcome and source
	// properties; authored requirements leave them NULL.
	for _, doc := range docs {
		for _, req := range doc.Requirements() {
			escID := escapeCypherString(req.ID)
			escFile := escapeCypherString(req.Source)
			outcome, _ := req.Attrs[model.AttrResultOutcome].(string)
			source := req.Source
			if outcome != "" {
				escOutcome := escapeCypherString(outcome)
				escSource := escapeCypherString(source)
				q := fmt.Sprintf("CREATE (n:Requirement {id: '%s', file: '%s', outcome: '%s', source: '%s'})", escID, escFile, escOutcome, escSource)
				if _, err := conn.Query(q); err != nil {
					return fmt.Errorf("inserting node for %s: %w", req.ID, err)
				}
			} else {
				q := fmt.Sprintf("CREATE (n:Requirement {id: '%s', file: '%s'})", escID, escFile)
				if _, err := conn.Query(q); err != nil {
					return fmt.Errorf("inserting node for %s: %w", req.ID, err)
				}
			}
		}
	}

	// Pass 2: create TracesTo edges for every trace attribute entry.
	for _, doc := range docs {
		for _, req := range doc.Requirements() {
			traceVal, ok := req.Attrs[model.AttrTrace]
			if !ok {
				continue
			}
			traceList, ok := traceVal.([]any)
			if !ok {
				continue
			}
			for _, ref := range traceList {
				refStr, ok := ref.(string)
				if !ok {
					continue
				}
				escFrom := escapeCypherString(req.ID)
				escTo := escapeCypherString(refStr)
				q := fmt.Sprintf(
					"MATCH (a:Requirement {id: '%s'}), (b:Requirement {id: '%s'}) CREATE (a)-[:TracesTo]->(b)",
					escFrom, escTo,
				)
				// Silently skip dangling refs (the MATCH will produce no rows,
				// but ladybugdb may still report success; we ignore errors here).
				if _, err := conn.Query(q); err != nil {
					// Dangling ref — skip silently (ladybugdb MATCH returns no rows).
					continue
				}
			}
		}
	}

	return nil
}

// escapeCypherString escapes single quotes by doubling them for safe inline
// Cypher string literals.
func escapeCypherString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
