package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"

	"reqmd/internal/exporter"
	"reqmd/internal/graph"
	"reqmd/internal/parser"
)

// pollInterval is the file polling interval used when fsnotify is unavailable.
const pollInterval = 2 * time.Second

func newServeCmd() *cobra.Command {
	var (
		addr     string
		debounce time.Duration
		headless bool
		noOpen   bool
	)

	cmd := &cobra.Command{
		Use:   "serve <root>",
		Short: "Watch requirements and serve live-reloading HTML",
		Long: `Watch a requirements directory tree and serve a live-reloading
HTML preview via HTTP.

Whenever a .md or schema.yaml file changes, reqmd re-parses,
re-validates, and re-exports all documents. The browser auto-reloads
via Server-Sent Events (SSE).

Use --headless for terminal-only re-check output without the HTTP server.`,
		Example: `  reqmd serve spec/reqs/
  reqmd serve spec/reqs/ --addr :9090
  reqmd serve spec/reqs/ --headless`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe(args[0], addr, debounce, headless, noOpen)
		},
	}

	cmd.Flags().StringVar(&addr, "addr", "localhost:8080", "HTTP server address")
	cmd.Flags().DurationVar(&debounce, "debounce", 500*time.Millisecond, "Debounce window for file events")
	cmd.Flags().BoolVar(&headless, "headless", false, "Terminal-only mode (no HTTP server)")
	cmd.Flags().BoolVar(&noOpen, "no-open", false, "Do not open a browser on start")
	return cmd
}

func runServe(root string, addr string, debounce time.Duration, headless bool, noOpen bool) error {
	root, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolving root: %w", err)
	}

	// Ensure root exists.
	if _, err := os.Stat(root); err != nil {
		return fmt.Errorf("root %s does not exist: %w", root, err)
	}

	// Watcher + HTTP server
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	srv := &serveData{
		root:     root,
		addr:     addr,
		debounce: debounce,
		headless: headless,
	}

	if err := srv.rebuild(ctx); err != nil {
		return fmt.Errorf("initial build: %w", err)
	}

	if !headless {
		// Start HTTP server
		mux := http.NewServeMux()
		mux.HandleFunc("/events", srv.sseHandler)
		mux.HandleFunc("/", srv.fileHandler)

		httpSrv := &http.Server{Addr: addr, Handler: mux}
		go func() {
			<-ctx.Done()
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			httpSrv.Shutdown(shutdownCtx)
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
	root     string
	addr     string
	debounce time.Duration
	headless bool

	// Latest rendered HTML per output path (relative path → content)
	mu    sync.RWMutex
	pages map[string][]byte

	// SSE clients
	sseMu      sync.Mutex
	sseClients map[chan struct{}]struct{}
}

func (s *serveData) rebuild(ctx context.Context) error {
	docs, err := parser.Discover(s.root)
	if err != nil {
		return fmt.Errorf("discovering documents: %w", err)
	}

	g, err := graph.New(docs)
	if err != nil {
		return fmt.Errorf("building graph: %w", err)
	}

	// Build ID→title map across all documents for trace link labels
	titleMap := make(map[string]string)
	for _, d := range docs {
		for _, req := range d.Requirements {
			if req.Title != "" {
				titleMap[req.ID] = req.Title
			}
		}
	}

	var exp exporter.HTML

	// Build doc chain and cross-file link resolution
	rctx, err := exporter.NewRenderContext(docs, s.root)
	if err != nil {
		return fmt.Errorf("building render context: %w", err)
	}

	boundaries := exporter.ComputeDocBoundaries(docs)

	upstreamFn := g.UpstreamNeighbors
	downstreamFn := g.DownstreamNeighbors

	pages := make(map[string][]byte)
	var totalReqs int
	results := g.CheckResults()
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
		tc := exporter.TraceResolver{
			Upstream:   upstreamFn,
			Downstream: downstreamFn,
			ResolveLink: func(reqID string) string {
				href := rawResolveLink(reqID)
				// Convert filesystem relative path to URL path (strip directory prefix)
				return filepath.Base(href)
			},
			TitleOf: func(id string) string { return titleMap[id] },
		}
		if err := exp.ExportWithTraces(&buf, doc, props, tc); err != nil {
			return fmt.Errorf("exporting %s: %w", doc.Path, err)
		}
		pages[outPath] = buf.Bytes()

		totalReqs += len(doc.Requirements)
	}

	// Count valid (no ERROR-level checks)
	errCount := 0
	warnCount := 0
	for _, cr := range results {
		switch cr.Level {
		case "ERROR":
			errCount++
		case "WARNING":
			warnCount++
		}
	}
	validReqs := totalReqs - errCount

	// Update pages atomically
	s.mu.Lock()
	s.pages = pages
	s.mu.Unlock()

	// Print status
	var status string
	if errCount == 0 {
		status = fmt.Sprintf("✅  all valid (%d reqs, %d warnings)", totalReqs, warnCount)
	} else {
		status = fmt.Sprintf("❌  %d/%d valid, %d errors, %d warnings", validReqs, totalReqs, errCount, warnCount)
	}
	fmt.Fprintf(os.Stderr, "\r[%s] %s\n", time.Now().Format("15:04:05"), status)

	return nil
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
	defer watcher.Close()

	// Only watch root + doc dirs. Recursive watching exhausts inotify limits.
	docDirs := findDocDirsFlat(s.root)
	watchDirs := append([]string{s.root}, docDirs...)
	for _, dir := range watchDirs {
		if err := watcher.Add(dir); err != nil {
			return errFsnotifyUnavailable
		}
	}

	rebuildAndNotify := s.makeRebuildAndNotify(ctx)

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
				watcher.Add(filepath.Dir(event.Name))
			}
			// If a new directory was created, check for schema.yaml
			if event.Op&fsnotify.Create != 0 {
				if fi, err := os.Stat(event.Name); err == nil && fi.IsDir() {
					if _, err := os.Stat(filepath.Join(event.Name, "schema.yaml")); err == nil {
						watcher.Add(event.Name)
					}
				}
			}
			ext := filepath.Ext(event.Name)
			if ext != ".md" && ext != ".yaml" && ext != ".yml" {
				continue
			}
			if event.Op&(fsnotify.Create|fsnotify.Write) == 0 {
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
	rebuildAndNotify := s.makeRebuildAndNotify(ctx)

	// Track last modification time per doc dir
	type docState struct {
		paths map[string]time.Time // file → mtime
	}
	state := make(map[string]*docState)

	// Collect file mtimes for all doc dirs
	snapshot := func() {
		docDirs := findDocDirsFlat(s.root)
		for _, dir := range docDirs {
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
				if ext != ".md" && ext != ".yaml" && ext != ".yml" {
					continue
				}
				info, err := e.Info()
				if err != nil {
					continue
				}
				newState.paths[filepath.Join(dir, e.Name())] = info.ModTime()
			}
			oldState := state[dir]
			state[dir] = newState

			// Compare with old state
			if oldState == nil {
				continue // first snapshot, no comparison
			}
			for path, mtime := range newState.paths {
				if oldMtime, ok := oldState.paths[path]; !ok || !mtime.Equal(oldMtime) {
					rebuildAndNotify()
					return
				}
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

func (s *serveData) makeRebuildAndNotify(ctx context.Context) func() {
	return func() {
		if err := s.rebuild(ctx); err != nil {
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
	switch runtime.GOOS {
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		return fmt.Errorf("unsupported platform %q to open browser", runtime.GOOS)
	}
}

// findDocDirsFlat finds all directories under root that contain a schema.yaml.
func findDocDirsFlat(root string) []string {
	var dirs []string
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && info.Name() == "schema.yaml" {
			dirs = append(dirs, filepath.Dir(path))
		}
		return nil
	})
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
	fmt.Fprintf(w, "event: connected\ndata: \n\n")
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-client:
			fmt.Fprintf(w, "event: reload\ndata: \n\n")
			flusher.Flush()
		}
	}
}

// fileHandler serves the in-memory rendered HTML pages.
func (s *serveData) fileHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/" || path == "" {
		// Serve an index page listing available documents
		s.mu.RLock()
		var names []string
		for p := range s.pages {
			names = append(names, p)
		}
		s.mu.RUnlock()

		if len(names) == 0 {
			http.Error(w, "no documents", http.StatusNotFound)
			return
		}
		// Redirect to first doc
		http.Redirect(w, r, "/"+names[0], http.StatusFound)
		return
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
		w.Write(content[:bodyEnd])
		w.Write(lrScript)
		w.Write(content[bodyEnd:])
	} else {
		w.Write(content)
		w.Write(lrScript)
	}
}
