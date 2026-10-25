package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/dVoo/reqmd/internal/exporter"
	"github.com/dVoo/reqmd/internal/filter"
	"github.com/dVoo/reqmd/internal/graph"
	"github.com/dVoo/reqmd/internal/model"
	"github.com/dVoo/reqmd/internal/parser"
	"github.com/dVoo/reqmd/internal/verify"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

// pollInterval is the file polling interval used when fsnotify is unavailable.
const pollInterval = 2 * time.Second

func newServeCmd() *cobra.Command {
	var (
		addr          string
		debounce      time.Duration
		headless      bool
		noOpen        bool
		resultsPaths  []string
		filterExpr    string
		relaxedFilter bool
	)

	cmd := &cobra.Command{
		Use:   "serve <root>",
		Short: "Watch requirements and serve live-reloading HTML",
		Long: `Watch a requirements directory tree and serve a live-reloading
HTML preview via HTTP.

Whenever a .md or schema.yaml file changes, reqmd re-parses, rebuilds the
trace graph, re-checks graph-level invariants, and re-exports all documents.
The root URL serves the project home (catalog, project-wide health, exception
queues, and cross-document search); each document has Document, Table, Trace,
and Coverage views. The browser auto-reloads via Server-Sent Events (SSE).

With --results, verification results (CTRF or manual) are loaded and
rendered as verdict badges on measure cards. Result file changes also
trigger a rebuild.

Note: serve runs graph-level checks (trace refs, cycles, coverage) but does
not re-run JSON Schema validation. Use ` + "`reqmd check`" + ` for full validation.

Use --headless for terminal-only re-check output without the HTTP server.`,
		Example: `  reqmd serve spec/
  reqmd serve spec/ --addr :9090
  reqmd serve spec/ --headless
  reqmd serve spec/ --results tests/ --results ci/ctrf.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe(args[0], addr, debounce, headless, noOpen, resultsPaths, filterExpr, relaxedFilter)
		},
	}

	cmd.Flags().StringVar(&addr, "addr", "localhost:8080", "HTTP server address")
	cmd.Flags().DurationVar(&debounce, "debounce", 500*time.Millisecond, "Debounce window for file events")
	cmd.Flags().BoolVar(&headless, "headless", false, "Terminal-only mode (no HTTP server)")
	cmd.Flags().StringArrayVar(&resultsPaths, "results", nil, "Load ephemeral verification results (CTRF .ctrf.json or manual-results dirs) and render verdict badges. Repeatable. Results paths are also watched for changes.")
	cmd.Flags().BoolVar(&noOpen, "no-open", false, "Do not open a browser on start")
	cmd.Flags().StringVar(&filterExpr, "filter", "", "Filter requirements using an expr-lang expression. Only matching requirements are served.")
	addRelaxedFilterFlag(cmd, &relaxedFilter)
	return cmd
}

func runServe(root string, addr string, debounce time.Duration, headless bool, noOpen bool, resultsPaths []string, filterExpr string, relaxedFilter bool) error {
	root, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolving root: %w", err)
	}

	// Ensure root exists.
	if _, err := os.Stat(root); err != nil {
		return fmt.Errorf("root %s does not exist: %w", root, err)
	}

	// Resolve results paths to absolute and validate existence.
	var absResults []string
	for _, rp := range resultsPaths {
		abs, err := filepath.Abs(rp)
		if err != nil {
			return fmt.Errorf("resolving results path %s: %w", rp, err)
		}
		if _, err := os.Stat(abs); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: results path %s does not exist: %v\n", abs, err)
		}
		absResults = append(absResults, abs)
	}

	// Watcher + HTTP server
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	srv := &serveData{
		root:          root,
		addr:          addr,
		debounce:      debounce,
		headless:      headless,
		resultsPaths:  absResults,
		filterExpr:    filterExpr,
		relaxedFilter: relaxedFilter,
	}

	if err := srv.rebuild(); err != nil {
		return fmt.Errorf("initial build: %w", err)
	}

	if !headless {
		// Start HTTP server
		mux := http.NewServeMux()
		mux.HandleFunc("/events", srv.sseHandler)
		mux.HandleFunc("/", srv.fileHandler)

		httpSrv := &http.Server{
			Addr:              addr,
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
		}
		go func() {
			<-ctx.Done()

			// Shutdown must outlive the canceled parent, so derive a context
			// that ignores the parent's cancellation.
			shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
			defer cancel()
			_ = httpSrv.Shutdown(shutdownCtx)
		}()

		go func() {
			fmt.Fprintf(os.Stderr, "Serving at http://%s\n", addr)
			fmt.Fprintf(os.Stderr, "Press Ctrl+C to stop\n")
			if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				fmt.Fprintf(os.Stderr, "HTTP server error: %v\n", err)
			}
		}()

		// Open browser on start
		if !noOpen {
			url := "http://" + addr
			if err := openBrowser(url); err != nil {
				fmt.Fprintf(os.Stderr, "Open browser at %s\n", url)
			}
		}
	} else {
		fmt.Fprintf(os.Stderr, "Watching %s... (headless mode)\n", root)
		fmt.Fprintf(os.Stderr, "Press Ctrl+C to stop\n")
	}

	return srv.watchLoop(ctx)
}

// serveData holds shared state for the serve command.
type serveData struct {
	pages         map[string][]byte
	sseClients    map[chan struct{}]struct{}
	root          string
	addr          string
	filterExpr    string
	resultsPaths  []string
	debounce      time.Duration
	mu            sync.RWMutex
	sseMu         sync.Mutex
	headless      bool
	relaxedFilter bool
}

func (s *serveData) rebuild() error {
	docs, analysis, g, err := discoverAndBuildGraph(s.root, s.resultsPaths, s.filterExpr, s.relaxedFilter)
	if err != nil {
		return err
	}

	titleMap := exporter.BuildTitleMap(docs)

	rctx, err := exporter.NewRenderContext(docs, s.root)
	if err != nil {
		return fmt.Errorf("building render context: %w", err)
	}
	boundaries := exporter.ComputeDocBoundaries(docs)

	pages, totalReqs, results, err := exportServePages(docs, analysis, g, rctx, boundaries, titleMap)
	if err != nil {
		return err
	}

	publishPages(s, pages)

	// Count ERROR/WARNING-level graph checks. serve does not run Pass 1
	// schema validation, so a per-requirement "valid" count would be
	// misleading (one req can yield multiple ERRORs, or none). Report
	// check counts instead.
	errCount := 0
	warnCount := 0
	for _, cr := range results {
		switch cr.Level {
		case graph.LevelError:
			errCount++
		case graph.LevelWarning:
			warnCount++
		}
	}

	// Print status
	var status string
	if errCount == 0 {
		status = fmt.Sprintf("✅  no graph errors (%d reqs, %d warnings)", totalReqs, warnCount)
	} else {
		status = fmt.Sprintf("❌  %d reqs, %d errors, %d warnings", totalReqs, errCount, warnCount)
	}
	fmt.Fprintf(os.Stderr, "\r[%s] %s\n", time.Now().Format("15:04:05"), status)

	return nil
}

// discoverAndBuildGraph runs Pass 0 document discovery, loads ephemeral
// verification results (if any), synthesizes them into pseudo-requirements,
// and builds the requirement graph from the combined document set.
// Returns the original docs (without synthesized results), the verdicts
// map for badge rendering, and the graph.
func discoverAndBuildGraph(root string, resultsPaths []string, filterExpr string, relaxedFilter bool) ([]model.Document, *exporter.Analysis, *graph.Graph, error) {
	docs, err := parser.Discover(root)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("discovering documents: %w", err)
	}

	graphDocs := docs
	if len(resultsPaths) > 0 {
		merged, _, loadErr := verify.LoadMerged(resultsPaths)
		if loadErr != nil {
			return nil, nil, nil, fmt.Errorf("loading results: %w", loadErr)
		}
		graphDocs = append(graphDocs, verify.Synthesize(merged))
	}

	g, err := graph.New(graphDocs)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("building graph: %w", err)
	}

	// Apply filter to the graph (filter-aware checks) and to the exported docs.
	if filterExpr != "" {
		f, warnings, err := filter.CompileForDocs(docs, filterExpr, relaxedFilter)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("compiling filter %q: %w", filterExpr, err)
		}
		filter.WriteWarnings(os.Stderr, warnings)
		filterSet, err := f.MatchingIDs(docs)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("collecting matching IDs: %w", err)
		}
		g.SetFilter(filterSet)
		docs, err = f.FilterDocs(docs)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("applying filter %q: %w", filterExpr, err)
		}
	}

	titleMap := exporter.BuildTitleMap(docs)
	analysis := exporter.NewAnalysis(g, titleMap)
	// Every page is served at the HTTP root, so a cross-document link is just
	// the target page's basename.
	analysis.SetHrefResolver(func(id string) string {
		for _, doc := range docs {
			for _, req := range doc.Requirements() {
				if req.ID == id {
					return filepath.Base(doc.Path) + "-requirements.html#" + id
				}
			}
		}
		return ""
	})
	return docs, analysis, g, nil
}

// exportServePages renders every document to HTML and counts graph checks.
// It returns the rendered pages, the total requirement count, and the graph
// check results so the caller can report status.
func exportServePages(docs []model.Document, analysis *exporter.Analysis, g *graph.Graph, rctx *exporter.RenderContext, boundaries map[string]exporter.DocBoundary, titleMap map[string]string) (map[string][]byte, int, []graph.CheckResult, error) {
	var exp exporter.HTML
	exp.SetAnalysis(analysis)

	pages := make(map[string][]byte)
	var totalReqs int
	for _, doc := range docs {
		props := doc.Properties
		dirName := filepath.Base(doc.Path)
		outPath := dirName + "-requirements.html"

		exp.SetBoundary(boundaries[doc.Path])

		var buf bytes.Buffer
		chainGraph := rctx.BuildChainGraph(doc)
		// In serve mode all files are served at HTTP root, so collapse
		// every card path to its basename.
		exporter.BasenameChainGraph(&chainGraph)
		exp.SetDocChainGraph(chainGraph)

		// ResolveLink uses filesystem-relative paths; convert to URL paths
		// (just the filename) since all files are at HTTP root in serve mode.
		absOutPath := filepath.Join(doc.Path, outPath)
		rawResolveLink := rctx.ResolveLink(absOutPath)
		exp.SetResolveLink(func(reqID string) string {
			// Convert the filesystem-relative path to a URL path: every
			// page is served at the HTTP root.
			return filepath.Base(rawResolveLink(reqID))
		})
		if err := exp.Export(&buf, doc, props); err != nil {
			return nil, 0, nil, fmt.Errorf("exporting %s: %w", doc.Path, err)
		}
		pages[outPath] = buf.Bytes()

		totalReqs += len(doc.Requirements())
	}

	// Project home: / serves the catalog instead of redirecting to the
	// first document.
	project := exporter.NewProjectIndex(docs, analysis, titleMap, func(docPath string) string {
		return filepath.Base(docPath) + "-requirements.html"
	})
	var index bytes.Buffer
	if err := project.Export(&index); err != nil {
		return nil, 0, nil, fmt.Errorf("exporting project index: %w", err)
	}
	pages["index.html"] = index.Bytes()

	return pages, totalReqs, g.CheckResults(), nil
}

// publishPages atomically swaps the in-memory rendered pages.
func publishPages(s *serveData, pages map[string][]byte) {
	s.mu.Lock()
	s.pages = pages
	s.mu.Unlock()
}

func (s *serveData) watchLoop(ctx context.Context) error {
	// Try fsnotify first; fall back to polling if it fails.
	if err := s.watchFsnotify(ctx); err != nil {
		if errors.Is(err, errFsnotifyUnavailable) {
			fmt.Fprintf(os.Stderr, "fsnotify unavailable, falling back to polling (every %v)\n", pollInterval)
			return s.watchPolling(ctx)
		}
		return err
	}
	return nil
}

// errFsnotifyUnavailable is returned when fsnotify cannot be used.
var errFsnotifyUnavailable = fmt.Errorf("fsnotify unavailable")

func (s *serveData) watchFsnotify(ctx context.Context) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return errFsnotifyUnavailable
	}
	defer func() { _ = watcher.Close() }()

	// Watch root + doc dirs. Recursive watching exhausts inotify limits.
	docDirs := findDocDirsFlat(s.root)
	watchDirs := append([]string{s.root}, docDirs...)
	// Also watch directories containing results files.
	for _, rp := range s.resultsPaths {
		watchDirs = append(watchDirs, findResultWatchDirs(rp)...)
	}
	// Deduplicate watch dirs.
	seen := make(map[string]bool)
	uniqueDirs := watchDirs[:0]
	for _, dir := range watchDirs {
		if dir != "" && !seen[dir] {
			seen[dir] = true
			uniqueDirs = append(uniqueDirs, dir)
		}
	}
	watchDirs = uniqueDirs
	for _, dir := range watchDirs {
		if err := watcher.Add(dir); err != nil {
			return errFsnotifyUnavailable
		}
	}

	rebuildAndNotify := s.makeRebuildAndNotify()

	var debounceTimer *time.Timer
	var debounceCh chan struct{}

	for {
		select {
		case <-ctx.Done():
			return nil

		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			// If a new schema.yaml was created, add its parent dir
			if event.Op&fsnotify.Create != 0 && filepath.Base(event.Name) == "schema.yaml" {
				_ = watcher.Add(filepath.Dir(event.Name))
			}
			// If a new directory was created, check for schema.yaml
			if event.Op&fsnotify.Create != 0 {
				if fi, err := os.Stat(event.Name); err == nil && fi.IsDir() {
					if _, err := os.Stat(filepath.Join(event.Name, "schema.yaml")); err == nil {
						_ = watcher.Add(event.Name)
					}
				}
			}
			ext := filepath.Ext(event.Name)
			// Accept .md/.yaml/.yml for spec files, .json for CTRF results.
			if ext != ".md" && ext != ".yaml" && ext != ".yml" && ext != ".json" {
				continue
			}
			if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) == 0 {
				continue
			}

			// Debounce
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			debounceTimer = time.NewTimer(s.debounce)
			debounceCh = make(chan struct{}, 1)
			go func() {
				<-debounceTimer.C
				close(debounceCh)
			}()

		case <-debounceCh:
			debounceCh = nil
			rebuildAndNotify()

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Fprintf(os.Stderr, "watcher error: %v\n", err)
		}
	}
}

func (s *serveData) watchPolling(ctx context.Context) error {
	rebuildAndNotify := s.makeRebuildAndNotify()

	// Track last modification time per doc dir
	type docState struct {
		paths map[string]time.Time // file → mtime
	}
	state := make(map[string]*docState)

	// Collect file mtimes for all watched dirs (spec dirs + result dirs).
	// Result dirs also track .json files (CTRF reports).
	snapshot := func() {
		// Build the set of dirs to check: spec doc dirs + result watch dirs.
		docDirs := findDocDirsFlat(s.root)
		watchDirs := docDirs
		for _, rp := range s.resultsPaths {
			watchDirs = append(watchDirs, findResultWatchDirs(rp)...)
		}
		// Deduplicate.
		seen := make(map[string]bool)
		uniqueDirs := watchDirs[:0]
		for _, dir := range watchDirs {
			if dir != "" && !seen[dir] {
				seen[dir] = true
				uniqueDirs = append(uniqueDirs, dir)
			}
		}
		watchDirs = uniqueDirs

		for _, dir := range watchDirs {
			entries, err := os.ReadDir(dir)
			if err != nil {
				continue
			}
			newState := &docState{paths: make(map[string]time.Time)}
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				ext := filepath.Ext(e.Name())
				if ext != ".md" && ext != ".yaml" && ext != ".yml" && ext != ".json" {
					continue
				}
				info, err := e.Info()
				if err != nil {
					continue
				}
				newState.paths[filepath.Join(dir, e.Name())] = info.ModTime()
			}
			oldState := state[dir]
			// Compare with old state: detect added/modified and deleted files.
			if oldState == nil {
				continue // first snapshot, no comparison
			}
			changed := false
			for path, mtime := range newState.paths {
				if oldMtime, ok := oldState.paths[path]; !ok || !mtime.Equal(oldMtime) {
					changed = true
					break
				}
			}
			if !changed {
				for path := range oldState.paths {
					if _, ok := newState.paths[path]; !ok {
						changed = true // file was deleted
						break
					}
				}
			}
			if changed {
				rebuildAndNotify()
				return
			}
		}
	}

	// Initial snapshot (no rebuild)
	snapshot()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			snapshot()
		}
	}
}

func (s *serveData) makeRebuildAndNotify() func() {
	return func() {
		if err := s.rebuild(); err != nil {
			fmt.Fprintf(os.Stderr, "[%s] rebuild error: %v\n", time.Now().Format("15:04:05"), err)
			return
		}
		if !s.headless {
			s.sseMu.Lock()
			for client := range s.sseClients {
				select {
				case client <- struct{}{}:
				default:
				}
			}
			s.sseMu.Unlock()
		}
	}
}

// openBrowser opens the given URL in the default system browser.
func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.CommandContext(context.Background(), "xdg-open", url)
	case "darwin":
		cmd = exec.CommandContext(context.Background(), "open", url)
	case "windows":
		cmd = exec.CommandContext(context.Background(), "rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform %q to open browser", runtime.GOOS)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("opening browser: %w", err)
	}
	return nil
}

// findDocDirsFlat finds all directories under root that contain a schema.yaml.
// Entry-level walk errors are skipped: a watcher that sees the readable part
// of the tree is more useful than none.
func findDocDirsFlat(root string) []string {
	var dirs []string
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			// Best-effort discovery: skip the unreadable entry/subtree.
			return fs.SkipDir
		}
		if !d.IsDir() && d.Name() == "schema.yaml" {
			dirs = append(dirs, filepath.Dir(path))
		}
		return nil
	})
	return dirs
}

// findResultWatchDirs returns the directories that should be watched to
// detect changes in the given results path. For a file path, the parent
// directory is watched. For a directory path, the directory itself is
// watched (for CTRF files) plus any subdirectories containing a schema.yaml
// (manual-results dirs).
func findResultWatchDirs(resultsPath string) []string {
	info, err := os.Stat(resultsPath)
	if err != nil {
		return nil
	}
	if !info.IsDir() {
		return []string{filepath.Dir(resultsPath)}
	}
	dirs := []string{resultsPath}
	dirs = append(dirs, findDocDirsFlat(resultsPath)...)
	return dirs
}

// sseHandler implements Server-Sent Events for live-reload.
func (s *serveData) sseHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	client := make(chan struct{}, 1)
	s.sseMu.Lock()
	if s.sseClients == nil {
		s.sseClients = make(map[chan struct{}]struct{})
	}
	s.sseClients[client] = struct{}{}
	s.sseMu.Unlock()

	defer func() {
		s.sseMu.Lock()
		delete(s.sseClients, client)
		s.sseMu.Unlock()
	}()

	// Send initial connected event
	_, _ = fmt.Fprintf(w, "event: connected\ndata: \n\n")
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-client:
			_, _ = fmt.Fprintf(w, "event: reload\ndata: \n\n")
			flusher.Flush()
		}
	}
}

// fileHandler serves the in-memory rendered HTML pages.
func (s *serveData) fileHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/" || path == "" {
		// The project home page: document catalog, project-wide health,
		// actionable queues, and cross-document search.
		path = "/index.html"
	}

	// Strip leading /
	if path[0] == '/' {
		path = path[1:]
	}

	s.mu.RLock()
	content, ok := s.pages[path]
	s.mu.RUnlock()

	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Inject live-reload script before </body>
	lrScript := []byte(`<script>new EventSource('/events').addEventListener('reload',()=>location.reload());</script>`)
	bodyEnd := bytes.LastIndex(content, []byte("</body>"))
	if bodyEnd >= 0 {
		_, _ = w.Write(content[:bodyEnd])
		_, _ = w.Write(lrScript)
		_, _ = w.Write(content[bodyEnd:])
	} else {
		_, _ = w.Write(content)
		_, _ = w.Write(lrScript)
	}
}
