package export

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/dVoo/reqmd/internal/exporter"
	"github.com/dVoo/reqmd/internal/filter"
	"github.com/dVoo/reqmd/internal/graph"
	"github.com/dVoo/reqmd/internal/model"
	"github.com/dVoo/reqmd/internal/parser"
	"github.com/dVoo/reqmd/internal/verify"

	"github.com/spf13/cobra"
)

// outputDirPerm keeps output directories readable by the exporting user's
// group, following the CLI's 0750 convention.
const outputDirPerm = 0o750

func newHTMLCmd() *cobra.Command {
	var (
		outputDir     string
		resultsPaths  []string
		filterExpr    string
		relaxedFilter bool
	)

	cmd := &cobra.Command{
		Use:   "html <dir>",
		Short: "Export requirements to HTML format with trace links",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExportHTML(args[0], outputDir, resultsPaths, filterExpr, relaxedFilter)
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory for HTML files")
	cmd.Flags().StringArrayVar(&resultsPaths, "results", nil, "Load ephemeral verification results (CTRF .ctrf.json or manual-results dirs) to render verdict badges on measure cards. Repeatable.")
	cmd.Flags().StringVar(&filterExpr, "filter", "", "Filter requirements using an expr-lang expression. Only matching requirements are exported.")
	cmd.Flags().BoolVar(&relaxedFilter, "relaxed-filter", false, "Allow the filter to reference attributes not declared in any schema.yaml (warn instead of error).")
	return cmd
}

// runExportHTML writes one page per document plus the project home.
func runExportHTML(root, outputDir string, resultsPaths []string, filterExpr string, relaxedFilter bool) error {
	docs, exportDocs, matchedIDs, err := discoverHTMLDocs(root, filterExpr, relaxedFilter)
	if err != nil {
		return err
	}
	analysis, err := analyzeHTMLGraph(docs, resultsPaths, matchedIDs, root, outputDir)
	if err != nil {
		return err
	}
	rctx, err := exporter.NewRenderContext(docs, root)
	if err != nil {
		return fmt.Errorf("building render context: %w", err)
	}
	if outputDir != "" {
		if err := os.MkdirAll(outputDir, outputDirPerm); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}
	}

	out := newHTMLOutput(root, docs, exportDocs, outputDir)
	if err := out.writeDocumentPages(analysis, rctx); err != nil {
		return err
	}
	return out.writeProjectIndex(analysis)
}

// discoverHTMLDocs parses the tree and, when a filter is active, returns the
// matching requirement IDs for graph scoping plus the pruned export set. The
// graph is still built from the full docs so trace links resolve.
func discoverHTMLDocs(root, filterExpr string, relaxedFilter bool) (docs, exportDocs []model.Document, matchedIDs map[string]struct{}, err error) {
	docs, err = parser.Discover(root)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("discovering documents: %w", err)
	}
	exportDocs = docs
	if filterExpr == "" {
		return docs, exportDocs, nil, nil
	}

	f, warnings, ferr := filter.CompileForDocs(docs, filterExpr, relaxedFilter)
	if ferr != nil {
		return nil, nil, nil, fmt.Errorf("compiling filter %q: %w", filterExpr, ferr)
	}
	filter.WriteWarnings(os.Stderr, warnings)
	exportDocs, err = f.FilterDocs(docs)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("applying filter %q: %w", filterExpr, err)
	}
	matchedIDs, err = f.MatchingIDs(docs)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("collecting matching IDs: %w", err)
	}
	return docs, exportDocs, matchedIDs, nil
}

// analyzeHTMLGraph builds the trace graph and adapts it into the presentation
// analysis every HTML view renders from. Ephemeral verification results are
// synthesized into pseudo-requirements before the graph is built.
func analyzeHTMLGraph(docs []model.Document, resultsPaths []string, matchedIDs map[string]struct{}, root, outputDir string) (*exporter.Analysis, error) {
	graphDocs := docs
	if len(resultsPaths) > 0 {
		merged, _, err := verify.LoadMerged(resultsPaths)
		if err != nil {
			return nil, fmt.Errorf("loading results: %w", err)
		}
		graphDocs = append(graphDocs, verify.Synthesize(merged))
	}
	g, err := graph.New(graphDocs)
	if err != nil {
		return nil, fmt.Errorf("building trace graph: %w", err)
	}
	if matchedIDs != nil {
		g.SetFilter(matchedIDs)
	}
	analysis := exporter.NewAnalysis(g, exporter.BuildTitleMap(docs))
	out := newHTMLOutput(root, docs, docs, outputDir)
	analysis.SetHrefResolver(func(id string) string {
		target, ok := out.reqOut[id]
		if !ok {
			return ""
		}
		rel, relErr := filepath.Rel("", target)
		if relErr != nil {
			return ""
		}
		return filepath.ToSlash(rel) + "#" + id
	})
	return analysis, nil
}

// htmlOutput resolves output paths once for a document set and writes the
// document pages and the project home.
type htmlOutput struct {
	outPathFor   func(string) string
	outPathByDir map[string]string
	reqOut       map[string]string
	docs         []model.Document
	exportDocs   []model.Document
	root         string
	outputDir    string
}

// newHTMLOutput computes the output layout: flat files in outputDir when set,
// otherwise one page inside each document directory.
func newHTMLOutput(root string, docs, exportDocs []model.Document, outputDir string) *htmlOutput {
	outPathFor := func(docPath string) string {
		name := filepath.Base(docPath) + "-requirements.html"
		if outputDir != "" {
			return filepath.Join(outputDir, name)
		}
		return filepath.Join(docPath, name)
	}
	o := &htmlOutput{
		docs:         docs,
		exportDocs:   exportDocs,
		root:         root,
		outputDir:    outputDir,
		outPathFor:   outPathFor,
		outPathByDir: make(map[string]string, len(docs)),
		reqOut:       make(map[string]string),
	}
	for _, d := range docs {
		o.outPathByDir[filepath.Base(d.Path)] = outPathFor(d.Path)
		for _, req := range d.Requirements() {
			o.reqOut[req.ID] = outPathFor(d.Path)
		}
	}
	return o
}

// resolveFor maps a requirement ID to an href relative to current's
// directory, so trace links work in both output layouts.
func (o *htmlOutput) resolveFor(current string) func(string) string {
	return func(id string) string {
		target, ok := o.reqOut[id]
		if !ok || target == current {
			return "#" + id
		}
		rel, err := filepath.Rel(filepath.Dir(current), target)
		if err != nil {
			return "#" + id
		}
		return filepath.ToSlash(rel) + "#" + id
	}
}

// writeDocumentPages renders every exported document page.
func (o *htmlOutput) writeDocumentPages(analysis *exporter.Analysis, rctx *exporter.RenderContext) error {
	boundaries := exporter.ComputeDocBoundaries(o.docs)
	var exp exporter.HTML
	exp.SetAnalysis(analysis)
	for _, doc := range o.exportDocs {
		if err := o.writeDocumentPage(&exp, doc, boundaries[doc.Path], rctx); err != nil {
			return err
		}
	}
	return nil
}

// writeDocumentPage renders one document page, wiring its boundary badge,
// document chain, and per-page trace link resolver.
func (o *htmlOutput) writeDocumentPage(exp *exporter.HTML, doc model.Document, boundary exporter.DocBoundary, rctx *exporter.RenderContext) error {
	outPath := o.outPathFor(doc.Path)

	exp.SetBoundary(boundary)
	chain := rctx.BuildChainGraph(doc)
	exporter.ResolveChainOutputs(&chain, func(dirName string) string {
		return o.outPathByDir[dirName]
	})
	exporter.RelativizeChainGraph(&chain, filepath.Dir(outPath))
	exp.SetDocChainGraph(chain)
	exp.SetResolveLink(o.resolveFor(outPath))

	err := writeHTMLFile(outPath, func(w io.Writer) error {
		return exp.Export(w, doc, doc.Properties)
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Wrote %s\n", outPath)
	return nil
}

// writeProjectIndex writes the project home page.
func (o *htmlOutput) writeProjectIndex(analysis *exporter.Analysis) error {
	indexPath := filepath.Join(o.root, "index.html")
	if o.outputDir != "" {
		indexPath = filepath.Join(o.outputDir, "index.html")
	}
	hrefFor := func(docPath string) string {
		rel, err := filepath.Rel(filepath.Dir(indexPath), o.outPathFor(docPath))
		if err != nil {
			return o.outPathFor(docPath)
		}
		return filepath.ToSlash(rel)
	}
	project := exporter.NewProjectIndex(o.exportDocs, analysis, exporter.BuildTitleMap(o.docs), hrefFor)
	if err := writeHTMLFile(indexPath, project.Export); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Wrote %s\n", indexPath)
	return nil
}

// writeHTMLFile creates path, runs write, and reports both the write and close
// error so a truncated export is never reported as written.
func writeHTMLFile(path string, write func(io.Writer) error) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating %s: %w", path, err)
	}
	writeErr := write(f)
	closeErr := f.Close()
	if writeErr != nil {
		return fmt.Errorf("exporting %s: %w", path, writeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("closing %s: %w", path, closeErr)
	}
	return nil
}
