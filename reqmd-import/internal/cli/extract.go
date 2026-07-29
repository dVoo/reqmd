package cli

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/spf13/cobra"

	"reqmd-import/internal/extractor"
	"reqmd-import/internal/lang"
	"reqmd-import/internal/model"
	"reqmd-import/internal/writer"
)

// Register installs all subcommands exposed by this package onto root.
// main() calls this once after constructing the root command.
//
// Keeping the registration entry point in one place (rather than relying
// on init() side effects) makes the command tree explicit and easy to
// inspect from main.go.
func Register(root *cobra.Command) {
	root.AddCommand(newExtractCmd())
}

// extractFlags carries the CLI flags for the extract subcommand. Using
// a struct (rather than free-floating variables) keeps the flag state
// self-contained and easy to pass to runExtract.
type extractFlags struct {
	lang            string
	heuristicTraces bool
	idPrefix        string
}

func newExtractCmd() *cobra.Command {
	f := &extractFlags{}

	cmd := &cobra.Command{
		Use:   "extract <source-dir> <target-dir>",
		Short: "Extract source-code symbols into reqmd .md requirement files",
		Long: `Walks <source-dir> for files matching any registered language plugin,
parses each file via tree-sitter, groups symbols by package, and writes
per-package .md files (plus the standard schema.yaml) to <target-dir>.

The generated .md files are valid reqmd input and can be validated with
'reqmd check <target-dir>' from the reqmd CLI.

Examples:
  reqmd-import extract ./internal ./spec/imported
  reqmd-import extract --lang go --heuristic-traces ./pkg ./spec/imported-go
  reqmd-import extract --id-prefix MY- ./src ./spec/mine`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExtract(cmd, args, f)
		},
	}

	cmd.Flags().StringVar(&f.lang, "lang", "", "restrict to one language (e.g. 'go', 'python'); default: all registered")
	cmd.Flags().BoolVar(&f.heuristicTraces, "heuristic-traces", false, "also extract bare requirement IDs from prose (not just explicit reqmd:trace markers)")
	cmd.Flags().StringVar(&f.idPrefix, "id-prefix", "IMP-", "prefix for generated requirement IDs (forward-compat; the embedded schema uses 'IMP-')")

	return cmd
}

// runExtract is the cobra RunE body. Splitting it out of the constructor
// keeps the flag wiring in newExtractCmd and the actual work here, so
// this function is straightforward to call from a test if we ever want
// to exercise the pipeline without going through cobra.
func runExtract(cmd *cobra.Command, args []string, f *extractFlags) error {
	srcDir, targetDir := args[0], args[1]

	// Validate the source directory up front with a clear message —
	// failing later in the walk with a less obvious error is a worse
	// user experience.
	srcInfo, err := os.Stat(srcDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("source directory %q does not exist", srcDir)
		}
		return fmt.Errorf("stat source directory %q: %w", srcDir, err)
	}
	if !srcInfo.IsDir() {
		return fmt.Errorf("source path %q is not a directory", srcDir)
	}

	// Resolve --lang up front. An unknown language name is a user error,
	// not a silent fallback, so we surface it here with the list of
	// registered languages for discoverability.
	restrictTo := func(l lang.Language) bool { return true }
	if f.lang != "" {
		plug, ok := lang.Get(f.lang)
		if !ok {
			names := make([]string, 0)
			for _, l := range lang.All() {
				names = append(names, l.Name())
			}
			sort.Strings(names)
			return fmt.Errorf("unknown --lang %q (registered: %s)", f.lang, strings.Join(names, ", "))
		}
		want := extSet(plug.Extensions())
		restrictTo = func(l lang.Language) bool {
			if l.Name() == plug.Name() {
				return true
			}
			// Also allow a different plugin if it shares at least one
			// extension with the requested plugin — this keeps a
			// hypothetical future ".go" support in another language
			// usable. In practice, only the Go plugin claims ".go".
			for _, e := range l.Extensions() {
				if want[e] {
					return true
				}
			}
			return false
		}
	}

	// Walk the source tree. We collect (file, plugin) pairs first, then
	// parse in a second pass; this decouples filesystem I/O from tree-
	// sitter parsing and keeps the loop body small.
	type job struct {
		path string
		plug lang.Language
	}
	var jobs []job
	skipped := 0
	walkErr := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			// Don't abort the whole walk on a single unreadable entry;
			// surface it and continue. Aborting would leave the user
			// with no partial output, which is usually less useful than
			// "everything I could read is in <target-dir>".
			fmt.Fprintf(cmd.ErrOrStderr(), "warn: skipping %s: %v\n", path, walkErr)
			return nil
		}
		if d.IsDir() {
			// Skip common noise directories that are never source.
			name := d.Name()
			if path != srcDir && (name == ".git" || name == "node_modules" || name == "vendor" || name == ".venv") {
				return fs.SkipDir
			}
			return nil
		}
		plug, ok := lang.ForExtension(filepath.Ext(path))
		if !ok {
			skipped++
			return nil
		}
		if !restrictTo(plug) {
			skipped++
			return nil
		}
		jobs = append(jobs, job{path: path, plug: plug})
		return nil
	})
	if walkErr != nil {
		return fmt.Errorf("walk source directory: %w", walkErr)
	}

	// Stable order over jobs so the output is byte-deterministic across
	// runs (matters for the writer's idempotent write-if-changed path).
	sort.Slice(jobs, func(i, j int) bool { return jobs[i].path < jobs[j].path })

	// Parse every file in parallel. Per-file errors are non-fatal: log
	// and continue so one bad file doesn't lose the rest of the run.
	//
	// Fan out to runtime.NumCPU() workers. Each worker reads + parses
	// one file at a time using a fresh tree-sitter parser (the lang
	// plugins create a new parser per Parse call, so there is no
	// shared mutable state). The `results` and `errs` channels are
	// buffered to len(jobs) so workers never block on send.
	//
	// Jobs are pre-sorted by path above; the per-package sort by
	// StartLine downstream is unchanged. Worker-completion order is
	// not deterministic, but it doesn't matter: symbols are bucketed
	// by Package, and within each package they are re-sorted by
	// StartLine before being written. The final on-disk output is
	// byte-identical to the sequential implementation.
	parseFailures := 0
	var allSymbols []model.Symbol
	if len(jobs) > 0 {
		workers := runtime.NumCPU()
		if workers < 1 {
			workers = 1
		}
		if workers > len(jobs) {
			workers = len(jobs)
		}

		jobCh := make(chan job)
		results := make(chan []model.Symbol, len(jobs))
		errs := make(chan string, len(jobs))

		var wg sync.WaitGroup
		for w := 0; w < workers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := range jobCh {
					src, err := os.ReadFile(j.path)
					if err != nil {
						errs <- fmt.Sprintf("warn: read %s: %v", j.path, err)
						results <- nil
						continue
					}
					syms, err := j.plug.Parse(j.path, src)
					if err != nil {
						errs <- fmt.Sprintf("warn: parse %s: %v", j.path, err)
						results <- nil
						continue
					}
					results <- syms
				}
			}()
		}

		for _, j := range jobs {
			jobCh <- j
		}
		close(jobCh)
		wg.Wait()
		close(results)
		close(errs)

		allSymbols = make([]model.Symbol, 0, len(jobs))
		for syms := range results {
			if syms == nil {
				parseFailures++
				continue
			}
			allSymbols = append(allSymbols, syms...)
		}
		for e := range errs {
			fmt.Fprintln(cmd.ErrOrStderr(), e)
		}
	}

	// Bucket by package, then build per-package GeneratedReqs. We bucket
	// first (not per-file) so a method on a struct in the same package
	// can find its parent type's ID.
	pkgBuckets := make(map[string][]model.Symbol)
	for _, s := range allSymbols {
		pkgBuckets[s.Package] = append(pkgBuckets[s.Package], s)
	}

	packages := make([]writer.Package, 0, len(pkgBuckets))
	for pkgName, syms := range pkgBuckets {
		// Stable, line-free ID: hash over (package, name) for top-level
		// symbols, and (package, receiver.name) for methods. The
		// receiver-aware key prevents a method named `multiply` in class
		// `Calculator` and a top-level function named `multiply` from
		// collapsing to the same ID in the same package.
		//
		// Two symbols with the exact same (package, name, receiver) —
		// which a well-formed source shouldn't produce — would still
		// collide; we don't try to disambiguate, the writer's stable
		// sort is the only stable guarantee at that point.
		idFor := make(map[string]string, len(syms))
		for _, s := range syms {
			hashKey := s.Name
			if s.Receiver != "" {
				hashKey = s.Receiver + "." + s.Name
			}
			idFor[hashKey] = f.idPrefix + pkgName + "-" + s.Name + "-" + model.ShortHash(pkgName, hashKey)
		}

		// Sort symbols by start line so the per-package output is
		// deterministic and methods follow their types in source order.
		sort.SliceStable(syms, func(i, j int) bool { return syms[i].StartLine < syms[j].StartLine })

		reqs := make([]model.GeneratedReq, 0, len(syms))
		for _, s := range syms {
			explicit, heuristic := extractor.ExtractTraces(s.Doc, f.heuristicTraces)
			// Methods are stored under "Receiver.Name" in idFor to
			// disambiguate from a same-named top-level function; look
			// up the right key. Top-level symbols (functions, types,
			// consts) are keyed by their bare name.
			selfKey := s.Name
			if s.Receiver != "" {
				selfKey = s.Receiver + "." + s.Name
			}
			req := model.GeneratedReq{
				ID:             idFor[selfKey],
				Title:          s.Name,
				Status:         "approved",
				Body:           s.Doc,
				SourceFile:     s.File,
				SourceLine:     s.StartLine,
				SymbolKind:     string(s.Kind),
				Trace:          explicit,
				HeuristicTrace: heuristic,
				Symbol:         s,
			}
			if s.Kind == model.SymbolMethod && s.Receiver != "" {
				// The parent (a type/struct/class) is a top-level symbol,
				// so its key in idFor is just its name. Methods themselves
				// are keyed by "Receiver.Name" above; look up the parent
				// by its bare name.
				if pid, ok := idFor[model.StripReceiverName(s.Receiver)]; ok {
					req.ParentID = pid
				}
			}
			reqs = append(reqs, req)
		}

		packages = append(packages, writer.Package{
			Name: pkgName,
			Reqs: reqs,
		})
	}

	// Write the output tree: top-level schema.yaml + per-package .md files.
	if err := writer.WriteTarget(targetDir, packages, f.idPrefix); err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	// Summary: keep it short, one line per metric. Emitted to stdout
	// (not stderr) because the command has succeeded and the user
	// should see the result inline with their next command.
	reqCount := 0
	for _, p := range packages {
		reqCount += len(p.Reqs)
	}
	fmt.Fprintf(cmd.OutOrStdout(),
		"reqmd-import: wrote %d requirement(s) across %d package(s) to %s (skipped %d file(s) with no language plugin",
		reqCount, len(packages), targetDir, skipped)
	if parseFailures > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), ", %d parse failure(s)", parseFailures)
	}
	fmt.Fprintln(cmd.OutOrStdout(), ")")
	return nil
}

// extSet is a tiny helper for the --lang filter: build a set of the
// plugin's claimed extensions once, then look up by extension in O(1).
// Keeping it local to this file avoids a public utility for one caller.
func extSet(exts []string) map[string]bool {
	out := make(map[string]bool, len(exts))
	for _, e := range exts {
		out[e] = true
	}
	return out
}
