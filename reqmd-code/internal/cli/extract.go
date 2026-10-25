package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/dVoo/reqmd/reqmd-code/internal/extractor"
	"github.com/dVoo/reqmd/reqmd-code/internal/lang"
	"github.com/dVoo/reqmd/reqmd-code/internal/model"
	"github.com/dVoo/reqmd/reqmd-code/internal/schema"
	"github.com/dVoo/reqmd/reqmd-import/exchange"

	"github.com/spf13/cobra"
)

// extractFlags carries the CLI flags for extraction. Using a struct
// (rather than free-floating variables) keeps the flag state
// self-contained and easy to pass to runExtract.
type extractFlags struct {
	lang            string
	output          string
	idPrefix        string
	heuristicTraces bool
}

// NewRootCmd builds the reqmd-code command tree. Extraction is the tool's
// only job, so it is the root command itself: users run
// `reqmd-code <source-dir>` and get a reqmd-import/v1 exchange bundle.
// version is injected at build time via -ldflags "-X main.version=...".
func NewRootCmd(version string) *cobra.Command {
	f := &extractFlags{}

	cmd := &cobra.Command{
		Use:     "reqmd-code <source-dir>",
		Short:   "Extract source-code symbols into a reqmd-import/v1 exchange bundle",
		Version: version,
		Long: `Walks <source-dir> for files matching any registered language plugin, parses
each file via tree-sitter, groups symbols by package, and emits a
reqmd-import/v1 JSON bundle on stdout (or to -o <file>).

The bundle is rendered into reqmd documents by reqmd-import:

  reqmd-code ./internal | reqmd-import - -o spec/03-software/imported
  reqmd check spec/

Examples:
  reqmd-code ./internal
  reqmd-code --lang go --heuristic-traces ./pkg -o imported.json
  reqmd-code --id-prefix MY- ./src | reqmd-import - -o spec/mine`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExtract(cmd, args, f)
		},
	}

	cmd.Flags().StringVar(&f.lang, "lang", "", "restrict to one language (e.g. 'c', 'cpp', 'go', 'python', 'rust'); default: all registered")
	cmd.Flags().BoolVar(&f.heuristicTraces, "heuristic-traces", false, "also extract bare requirement IDs from prose (not just explicit reqmd:trace markers)")
	cmd.Flags().StringVar(&f.idPrefix, "id-prefix", "IMP-", "prefix for generated requirement IDs")
	cmd.Flags().StringVarP(&f.output, "output", "o", "", "write the JSON bundle to this file instead of stdout")

	return cmd
}

// runExtract is the cobra RunE body. Splitting it out of the constructor
// keeps the flag wiring in NewRootCmd and the actual work here, so
// this function is straightforward to call from a test if we ever want
// to exercise the pipeline without going through cobra.
func runExtract(cmd *cobra.Command, args []string, f *extractFlags) error {
	srcDir := args[0]

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
		plug lang.Language
		path string
	}
	var jobs []job
	skipped := 0
	walkErr := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			// Don't abort the whole walk on a single unreadable entry;
			// surface it and continue. Aborting would leave the user
			// with no partial output, which is usually less useful than
			// "everything I could read is in the bundle".
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

	// Parse every file in parallel, one result slot per job: no channel to
	// drain and no nil sentinel — a failed job just carries its error. Per-file
	// failures are non-fatal so one bad file doesn't lose the rest of the run.
	//
	// Jobs are pre-sorted by path, and symbols are bucketed by package and
	// re-sorted by (file, line) below, so worker completion order never reaches
	// the output.
	type parseResult struct {
		err     error
		symbols []model.Symbol
	}
	parsed := make([]parseResult, len(jobs))
	var wg sync.WaitGroup
	sem := make(chan struct{}, max(min(runtime.NumCPU(), len(jobs)), 1))
	for i, j := range jobs {
		sem <- struct{}{}
		wg.Go(func() {
			defer func() { <-sem }()
			src, err := os.ReadFile(j.path)
			if err != nil {
				parsed[i] = parseResult{err: fmt.Errorf("read %s: %w", j.path, err)}
				return
			}
			syms, err := j.plug.Parse(j.path, src)
			if err != nil {
				parsed[i] = parseResult{err: fmt.Errorf("parse %s: %w", j.path, err)}
				return
			}
			parsed[i] = parseResult{symbols: syms}
		})
	}
	wg.Wait()

	parseFailures := 0
	allSymbols := make([]model.Symbol, 0, len(jobs))
	for _, r := range parsed {
		if r.err != nil {
			parseFailures++
			fmt.Fprintf(cmd.ErrOrStderr(), "warn: %v\n", r.err)
			continue
		}
		allSymbols = append(allSymbols, r.symbols...)
	}

	// Bucket by package, then build per-package exchange documents. We bucket
	// first (not per-file) so a method on a struct in the same package
	// can find its parent type's native ID.
	pkgBuckets := make(map[string][]model.Symbol)
	for _, s := range allSymbols {
		pkgBuckets[s.Package] = append(pkgBuckets[s.Package], s)
	}

	// The standard document template: schema header, attr properties, level
	// and upstream. Every document in the bundle shares it; only the output
	// location, title and ID prefix differ per package.
	base, err := schema.Document()
	if err != nil {
		return fmt.Errorf("load standard schema: %w", err)
	}
	base.IDPrefix = f.idPrefix

	// The root document carries only the umbrella schema.yaml (no .md): it
	// declares the tree's level and the shared id-prefix, and is the parent
	// every per-package document points upstream at.
	docs := []exchange.Document{base}

	pkgNames := make([]string, 0, len(pkgBuckets))
	for pkgName := range pkgBuckets {
		pkgNames = append(pkgNames, pkgName)
	}
	sort.Strings(pkgNames)

	for _, pkgName := range pkgNames {
		syms := pkgBuckets[pkgName]

		// Stable, line-free native ID: hash over (package, name) for
		// top-level symbols, and (package, receiver.name) for methods. The
		// receiver-aware key prevents a method named `multiply` in class
		// `Calculator` and a top-level function named `multiply` from
		// collapsing to the same ID in the same package.
		//
		// Two symbols with the exact same (package, name, receiver) —
		// which a well-formed source shouldn't produce — would still
		// collide; we don't try to disambiguate, the deterministic sort is
		// the only stable guarantee at that point.
		nativeFor := make(map[string]string, len(syms))
		for _, s := range syms {
			key := nativeKey(s)
			nativeFor[key] = s.Name + "-" + model.ShortHash(pkgName, key)
		}

		// Sort symbols by (file, start line, name) so the per-package output is
		// deterministic (workers finish out of order) and methods follow their
		// types in source order. It also makes the duplicate-symbol dedup below
		// deterministic.
		sort.SliceStable(syms, func(i, j int) bool {
			if syms[i].File != syms[j].File {
				return syms[i].File < syms[j].File
			}
			if syms[i].StartLine != syms[j].StartLine {
				return syms[i].StartLine < syms[j].StartLine
			}
			return syms[i].Name < syms[j].Name
		})

		// Build-tag alternatives (e.g. graph.go guarded by `ladybug` and
		// graph_stub.go guarded by `!ladybug`) declare the same symbol twice;
		// tree-sitter sees both files. They are one logical requirement, so
		// the first in source order wins and the duplicate is skipped.
		emitted := make(map[string]struct{}, len(syms))
		reqs := make([]exchange.Requirement, 0, len(syms))
		for _, s := range syms {
			explicit, heuristic := extractor.ExtractTraces(s.Doc, f.heuristicTraces)
			// Methods are stored under "Receiver.Name" in nativeFor to
			// disambiguate from a same-named top-level function; top-level
			// symbols (functions, types, consts) are keyed by their bare name.
			nativeID := nativeFor[nativeKey(s)]
			if _, dup := emitted[nativeID]; dup {
				fmt.Fprintf(cmd.ErrOrStderr(), "warn: skipping duplicate symbol %s (%s) in package %s\n", s.Name, s.File, pkgName)
				continue
			}
			emitted[nativeID] = struct{}{}
			attrs := map[string]any{
				"x-reqmd.imported":    true,
				"x-reqmd.source-file": s.File,
				"x-reqmd.symbol-kind": string(s.Kind),
			}
			if s.StartLine > 0 {
				attrs["x-reqmd.source-line"] = s.StartLine
			}
			if len(heuristic) > 0 {
				attrs["x-reqmd.heuristic-trace"] = heuristic
			}
			req := exchange.Requirement{
				NativeID: nativeID,
				Title:    s.Name,
				Status:   "approved",
				Body:     s.Doc,
				Trace:    explicit,
				Attrs:    attrs,
			}
			if s.Kind == model.SymbolMethod && s.Receiver != "" {
				// The parent (a type/struct/class) is a top-level symbol,
				// so its key in nativeFor is just its name. Methods
				// themselves are keyed by "Receiver.Name" above; look up
				// the parent by its bare name.
				if pid, ok := nativeFor[model.StripReceiverName(s.Receiver)]; ok {
					req.Parent = pid
				}
			}
			reqs = append(reqs, req)
		}

		doc := base
		doc.Dir = pkgName
		doc.File = "package.md"
		doc.Title = "Package: " + pkgName
		doc.Notice = "This file is auto-generated by reqmd-code. Do not edit by hand."
		// Per-package prefix so reqmd's graph checker sees distinct ID
		// prefixes for every package and does not flag a collision. The
		// native IDs above carry the package name, so the final ID is
		// <idPrefix><pkg>-<name>-<hash>.
		doc.IDPrefix = f.idPrefix + pkgName + "-"
		doc.Requirements = reqs
		docs = append(docs, doc)
	}

	bundle := exchange.Bundle{Format: exchange.Format, Documents: docs}
	out, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return fmt.Errorf("encode exchange bundle: %w", err)
	}
	out = append(out, '\n')

	if f.output == "" || f.output == "-" {
		if _, err := cmd.OutOrStdout().Write(out); err != nil {
			return fmt.Errorf("write bundle: %w", err)
		}
	} else if err := os.WriteFile(f.output, out, 0o644); err != nil {
		return fmt.Errorf("write bundle to %s: %w", f.output, err)
	}

	// Summary goes to stderr so stdout stays a clean JSON stream.
	reqCount := 0
	for _, d := range docs {
		reqCount += len(d.Requirements)
	}
	dest := "stdout"
	if f.output != "" && f.output != "-" {
		dest = f.output
	}
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
		"reqmd-code: extracted %d requirement(s) across %d package(s) to %s (skipped %d file(s) with no language plugin",
		reqCount, len(pkgNames), dest, skipped)
	if parseFailures > 0 {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), ", %d parse failure(s)", parseFailures)
	}
	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), ")")
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

// nativeKey is a symbol's disambiguation key inside its package. Methods are
// keyed "Receiver.Name" so a method and a same-named top-level function don't
// collapse into one requirement ID; everything else is keyed by name.
func nativeKey(s model.Symbol) string {
	if s.Receiver != "" {
		return s.Receiver + "." + s.Name
	}
	return s.Name
}
