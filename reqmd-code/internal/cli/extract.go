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
	linkTests       bool
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
		// runExtract prints one batched diagnostic; Cobra must not repeat it
		// or dump usage for a data error.
		SilenceErrors: true,
		SilenceUsage:  true,
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
	cmd.Flags().BoolVar(&f.linkTests, "link-tests", false, "link unit tests to the symbols their names refer to (TestParse/test_parse); explicit reqmd:trace markers win")
	cmd.Flags().StringVar(&f.idPrefix, "id-prefix", "IMP-", "prefix for generated requirement IDs")
	cmd.Flags().StringVarP(&f.output, "output", "o", "", "write the JSON bundle to this file instead of stdout")

	return cmd
}

// reqmd:trace SW-IMP-001
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
	base.XReqmd.IDPrefix = f.idPrefix

	// The root document carries only the umbrella schema.yaml (no .md): it
	// declares the tree's level and the shared id-prefix, and is the parent
	// every per-package document points upstream at.
	root := withXReqmd(base)
	root.Key = "root"
	root.Title = "Imported proxies"
	root.XReqmd.Dir = "."
	docs := []exchange.Document{root}

	pkgNames := make([]string, 0, len(pkgBuckets))
	for pkgName := range pkgBuckets {
		pkgNames = append(pkgNames, pkgName)
	}
	sort.Strings(pkgNames)

	// First pass: sort and deduplicate each package's symbols, keeping the
	// stable native ID of every survivor. Build-tag alternatives
	// (graph.go / graph_stub.go) declare the same symbol twice; the first in
	// source order wins. Resolution runs over this deduplicated set so a
	// legitimate build-tag duplicate can never look like an ambiguity.
	pkgSymbols := make(map[string][]packagedSymbol, len(pkgBuckets))
	nativeIDs := make(map[string]map[string]string, len(pkgBuckets))
	for _, pkgName := range pkgNames {
		syms := pkgBuckets[pkgName]
		sortSymbols(syms)
		byKey := nativeIDMap(pkgName, syms)
		nativeIDs[pkgName] = byKey
		emitted := make(map[string]struct{}, len(syms))
		kept := make([]packagedSymbol, 0, len(syms))
		for _, s := range syms {
			id := byKey[nativeKey(s)]
			if _, dup := emitted[id]; dup {
				fmt.Fprintf(cmd.ErrOrStderr(), "warn: skipping duplicate symbol %s (%s) in package %s\n", s.Name, s.File, pkgName)
				continue
			}
			emitted[id] = struct{}{}
			kept = append(kept, packagedSymbol{sym: s, nativeID: id})
		}
		pkgSymbols[pkgName] = kept
	}
	index := newSymbolIndex(pkgSymbols)
	var conventionPool map[string][]conventionTarget
	if f.linkTests {
		conventionPool = buildConventionPool(pkgSymbols)
	}

	var refErrs []traceRefError
	var conventionWarns []string

	for _, pkgName := range pkgNames {
		kept := pkgSymbols[pkgName]

		// childOf collects methods per type key so they can be nested once the
		// parent's position in source order is known.
		childOf := make(map[string][]exchange.Item, len(kept))
		// Symbols are emitted into one section per source file, in source
		// order: the file list is contiguous because syms is sorted by file.
		var sections []exchange.Item
		var fileItems []exchange.Item
		curFile := ""
		flushFile := func() {
			if curFile == "" {
				return
			}
			sections = append(sections, exchange.Item{
				Kind:     exchange.KindSection,
				Title:    relPath(srcDir, curFile),
				Children: nestMethods(fileItems, childOf),
			})
			fileItems = nil
			childOf = make(map[string][]exchange.Item, len(kept))
		}
		for _, ps := range kept {
			s := ps.sym
			if s.File != curFile {
				flushFile()
				curFile = s.File
			}
			traces := extractor.ExtractTraces(s.Doc, f.heuristicTraces)
			attrs := map[string]any{
				"x-reqmd.imported":    true,
				"x-reqmd.source-file": s.File,
				"x-reqmd.symbol-kind": string(s.Kind),
			}
			if s.StartLine > 0 {
				attrs["x-reqmd.source-line"] = s.StartLine
			}
			if len(traces.Heuristic) > 0 {
				attrs["x-reqmd.heuristic-trace"] = traces.Heuristic
			}
			item := exchange.Item{
				Kind:   exchange.KindRequirement,
				Key:    ps.nativeID,
				Title:  s.Name,
				Status: "approved",
				Fields: attrs,
				Refs:   traces.ReqIDs,
			}
			for _, bad := range traces.Malformed {
				refErrs = append(refErrs, traceRefError{
					file: relPath(srcDir, s.File), line: s.StartLine, symbol: s.Name,
					message: fmt.Sprintf("invalid reqmd:trace marker %q (expected a symbol or a requirement ID)", bad),
				})
			}
			seenLinks := make(map[string]bool, len(traces.SymbolRefs))
			for _, ref := range traces.SymbolRefs {
				target, candidates, ok := index.resolve(pkgName, ref)
				if !ok {
					msg := fmt.Sprintf("unresolved symbol trace %q", ref)
					if len(candidates) > 0 {
						msg = fmt.Sprintf("ambiguous symbol trace %q (matches: %s)", ref, strings.Join(candidates, ", "))
					}
					refErrs = append(refErrs, traceRefError{
						file: relPath(srcDir, s.File), line: s.StartLine, symbol: s.Name, message: msg,
					})
					continue
				}
				link := target.nativeID
				if target.pkg != pkgName {
					link = target.pkg + "/" + target.nativeID
				}
				if !seenLinks[link] {
					seenLinks[link] = true
					item.Links = append(item.Links, link)
				}
			}
			// Convention linking is opt-in and never overrides an explicit
			// directive: any marker content (IDs, symbols, none, malformed)
			// suppresses it for that item.
			if f.linkTests && !hasExplicitDirective(traces) {
				if link, candidates, ok := conventionLink(pkgName, s, conventionPool); ok {
					if !seenLinks[link] {
						seenLinks[link] = true
						item.Links = append(item.Links, link)
					}
				} else if len(candidates) > 0 {
					conventionWarns = append(conventionWarns, fmt.Sprintf(
						"%s:%d (%s): ambiguous convention match (candidates: %s)",
						relPath(srcDir, s.File), s.StartLine, s.Name, strings.Join(candidates, ", ")))
				}
			}
			// The doc comment becomes content nodes under the symbol rather
			// than its body: a section body may not contain headings, and a
			// doc comment often does.
			item.Children = append(item.Children, docSections(s.Doc)...)
			// Methods nest under their type as a child item, which is how the
			// renderer expresses heading depth now that items are a tree.
			if s.Kind == model.SymbolMethod && s.Receiver != "" {
				// The parent (a type/struct/class) is a top-level symbol, so its
				// key in nativeIDs is its bare name. Methods themselves are keyed
				// by "Receiver.Name".
				if pid, ok := nativeIDs[pkgName][model.StripReceiverName(s.Receiver)]; ok {
					childOf[pid] = append(childOf[pid], item)
					continue
				}
			}
			fileItems = append(fileItems, item)
		}
		flushFile()
		items := sections

		doc := withXReqmd(base)
		doc.Key = pkgName
		doc.Title = "Package: " + pkgName
		doc.XReqmd.Dir = pkgName
		doc.XReqmd.File = "package.md"
		doc.XReqmd.DocumentID = pkgName
		// Per-package prefix so reqmd's graph checker sees distinct ID
		// prefixes for every package and does not flag a collision. The
		// keys above carry the package name, so the final ID is
		// <idPrefix><pkg>-<name>-<hash>.
		doc.XReqmd.IDPrefix = f.idPrefix + pkgName + "-"
		doc.Items = items
		docs = append(docs, doc)
	}

	// Explicit symbol references are a hard contract: report every bad one
	// before anything is marshaled or written, so stdout stays empty and an
	// `-o` file is never created or truncated on failure.
	if len(refErrs) > 0 {
		return formatTraceRefErrors(refErrs)
	}
	if len(conventionWarns) > 0 {
		sort.Strings(conventionWarns)
		for _, w := range conventionWarns {
			fmt.Fprintf(cmd.ErrOrStderr(), "warn: %s\n", w)
		}
	}

	bundle := exchange.Bundle{Schema: exchange.SchemaURL, Documents: docs}
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
		reqCount += countRequirements(d.Items)
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

// packagedSymbol is a symbol that survived per-package deduplication, with
// its stable native ID.
type packagedSymbol struct {
	sym      model.Symbol
	nativeID string
}

// traceRefError is one explicit `reqmd:trace` failure, addressed to the
// declaring symbol.
type traceRefError struct {
	file    string
	line    int
	symbol  string
	message string
}

// formatTraceRefErrors renders the batched failures deterministically.
func formatTraceRefErrors(errs []traceRefError) error {
	sort.Slice(errs, func(i, j int) bool {
		if errs[i].file != errs[j].file {
			return errs[i].file < errs[j].file
		}
		if errs[i].line != errs[j].line {
			return errs[i].line < errs[j].line
		}
		return errs[i].message < errs[j].message
	})
	const limit = 20
	var b strings.Builder
	b.WriteString("invalid reqmd:trace references:")
	for i, e := range errs {
		if i == limit {
			fmt.Fprintf(&b, "\n  ... and %d more", len(errs)-limit)
			break
		}
		fmt.Fprintf(&b, "\n  - %s:%d (%s): %s", e.file, e.line, e.symbol, e.message)
	}
	return errors.New(b.String())
}

// sortSymbols orders a package's symbols deterministically: by file, then
// start line, then name. Methods therefore follow their parent type.
func sortSymbols(syms []model.Symbol) {
	sort.SliceStable(syms, func(i, j int) bool {
		if syms[i].File != syms[j].File {
			return syms[i].File < syms[j].File
		}
		if syms[i].StartLine != syms[j].StartLine {
			return syms[i].StartLine < syms[j].StartLine
		}
		return syms[i].Name < syms[j].Name
	})
}

// nativeIDMap maps every symbol's native key to its stable item key. The key
// is `<Name>-<hash(pkg, nameOrReceiver.Name)>`: line-free, so edits above the
// symbol do not churn IDs.
func nativeIDMap(pkg string, syms []model.Symbol) map[string]string {
	out := make(map[string]string, len(syms))
	for _, s := range syms {
		key := nativeKey(s)
		out[key] = s.Name + "-" + model.ShortHash(pkg, key)
	}
	return out
}

// symbolTarget is a resolved explicit reference.
type symbolTarget struct {
	pkg      string
	nativeID string
}

// symbolIndex indexes deduplicated symbols by package and by the exact name
// a `reqmd:trace` target can use: the bare name for non-methods, and
// "Type.Name" for methods.
type symbolIndex struct {
	byPkg map[string]map[string][]symbolTarget
}

func newSymbolIndex(pkgs map[string][]packagedSymbol) *symbolIndex {
	ix := &symbolIndex{byPkg: make(map[string]map[string][]symbolTarget, len(pkgs))}
	for pkg, kept := range pkgs {
		entries := make(map[string][]symbolTarget, len(kept))
		for _, ps := range kept {
			key := ps.sym.Name
			if ps.sym.Receiver != "" {
				key = model.StripReceiverName(ps.sym.Receiver) + "." + ps.sym.Name
			}
			entries[key] = append(entries[key], symbolTarget{pkg: pkg, nativeID: ps.nativeID})
		}
		ix.byPkg[pkg] = entries
	}
	return ix
}

// resolve matches an explicit symbol trace exactly: case-sensitive, no prefix
// matching. `pkg/Name` and `pkg/Type.Name` search only that package; an
// unqualified reference tries the same package, the base package of a test
// package (`foo_test` -> `foo`, `test_foo` -> `foo`), then a globally unique
// match. On ambiguity it returns the candidate names.
func (ix *symbolIndex) resolve(fromPkg, ref string) (target symbolTarget, candidates []string, ok bool) {
	ref = strings.ReplaceAll(ref, "::", ".")
	qualPkg, key := "", ref
	if i := strings.IndexByte(ref, '/'); i >= 0 {
		qualPkg, key = ref[:i], ref[i+1:]
		if qualPkg == "" || key == "" {
			return symbolTarget{}, nil, false
		}
	}
	if strings.Count(key, ".") > 1 {
		return symbolTarget{}, nil, false
	}

	lookup := func(pkg string) []symbolTarget {
		if entries := ix.byPkg[pkg]; entries != nil {
			return entries[key]
		}
		return nil
	}
	decide := func(matches []symbolTarget) (symbolTarget, []string, bool) {
		switch len(matches) {
		case 0:
			return symbolTarget{}, nil, false
		case 1:
			return matches[0], nil, true
		default:
			names := make([]string, len(matches))
			for i, m := range matches {
				names[i] = m.pkg + "/" + key
			}
			sort.Strings(names)
			return symbolTarget{}, names, false
		}
	}

	if qualPkg != "" {
		return decide(lookup(qualPkg))
	}
	if matches := lookup(fromPkg); len(matches) > 0 {
		return decide(matches)
	}
	searched := map[string]bool{fromPkg: true}
	if base := basePackage(fromPkg); base != "" {
		searched[base] = true
		if matches := lookup(base); len(matches) > 0 {
			return decide(matches)
		}
	}
	var all []symbolTarget
	for pkg := range ix.byPkg {
		if searched[pkg] {
			continue
		}
		all = append(all, ix.byPkg[pkg][key]...)
	}
	return decide(all)
}

// basePackage returns the non-test package a test package belongs to, or ""
// when the name is not a test package.
func basePackage(pkg string) string {
	if before, ok := strings.CutSuffix(pkg, "_test"); ok {
		return before
	}
	if after, ok := strings.CutPrefix(pkg, "test_"); ok {
		return after
	}
	return ""
}

// conventionTarget is a symbol a test name may refer to by convention.
type conventionTarget struct {
	pkg      string
	nativeID string
	key      string
}

// buildConventionPool collects the symbols convention matching may link to:
// production functions, methods, and types. Constants, variables, and tests
// (including symbols in test files and inline `#[test]` items) are excluded so
// a test can never link to another test or to a constant.
func buildConventionPool(pkgs map[string][]packagedSymbol) map[string][]conventionTarget {
	pool := make(map[string][]conventionTarget, len(pkgs))
	for pkg, kept := range pkgs {
		targets := make([]conventionTarget, 0, len(kept))
		for _, ps := range kept {
			s := ps.sym
			if s.Kind == model.SymbolConst || s.Kind == model.SymbolVar || s.IsTest || lang.IsTestFile(s.File) {
				continue
			}
			key := s.Name
			if s.Receiver != "" {
				key = model.StripReceiverName(s.Receiver) + "_" + s.Name
			}
			targets = append(targets, conventionTarget{pkg: pkg, nativeID: ps.nativeID, key: key})
		}
		pool[pkg] = targets
	}
	return pool
}

// hasExplicitDirective reports whether an item already carries an explicit
// trace directive. Convention linking only fills the gap when nothing was
// written by hand.
func hasExplicitDirective(traces extractor.TraceRefs) bool {
	return len(traces.ReqIDs) > 0 || len(traces.SymbolRefs) > 0 || len(traces.Malformed) > 0 || traces.OptOut
}

// conventionLink resolves a test symbol to the production symbol its name
// refers to. Lookup is scoped to the same package, then the base package of a
// test package — never globally, so unrelated packages cannot produce a link.
func conventionLink(fromPkg string, sym model.Symbol, pool map[string][]conventionTarget) (link string, candidates []string, ok bool) {
	if !sym.IsTest && !lang.IsTestFile(sym.File) {
		return "", nil, false
	}
	stem, camel, ok := testStem(sym)
	if !ok {
		return "", nil, false
	}
	tiers := []string{fromPkg}
	if base := basePackage(fromPkg); base != "" && base != fromPkg {
		tiers = append(tiers, base)
	}
	for _, pkg := range tiers {
		target, cands, found := matchConvention(stem, camel, pool[pkg])
		if found {
			link := target.nativeID
			if target.pkg != fromPkg {
				link = target.pkg + "/" + target.nativeID
			}
			return link, nil, true
		}
		if len(cands) > 0 {
			return "", cands, false
		}
	}
	return "", nil, false
}

// testStem returns the part of a test name that names its target, and whether
// the language uses CamelCase (`TestParse`) or snake_case (`test_parse`).
// Symbol kinds other than functions/methods never yield a stem.
func testStem(sym model.Symbol) (stem string, camel bool, ok bool) {
	if sym.Kind != model.SymbolFunction && sym.Kind != model.SymbolMethod {
		return "", false, false
	}
	name := sym.Name
	if filepath.Ext(sym.File) == ".go" {
		if name == "TestMain" || len(name) <= len("Test") || !strings.HasPrefix(name, "Test") {
			return "", false, false
		}
		rest := name[len("Test"):]
		if !isUpperByte(rest[0]) {
			return "", false, false
		}
		return rest, true, true
	}
	if len(name) <= len("test_") || !strings.HasPrefix(name, "test_") {
		return "", false, false
	}
	return name[len("test_"):], false, true
}

// matchConvention finds the longest candidate key that the test stem starts
// with at a valid boundary. A tie across distinct targets is ambiguous and
// returns the candidate names.
func matchConvention(stem string, camel bool, targets []conventionTarget) (conventionTarget, []string, bool) {
	var best []conventionTarget
	bestLen := -1
	bestExact := false
	consider := func(t conventionTarget, text string, exact bool) {
		if !strings.HasPrefix(stem, text) || !boundaryOK(stem[len(text):], text, camel) {
			return
		}
		switch {
		case len(text) > bestLen || (len(text) == bestLen && exact && !bestExact):
			best = []conventionTarget{t}
			bestLen = len(text)
			bestExact = exact
		case len(text) == bestLen && exact == bestExact:
			for _, b := range best {
				if b.pkg == t.pkg && b.nativeID == t.nativeID {
					return
				}
			}
			best = append(best, t)
		}
	}
	for _, t := range targets {
		consider(t, t.key, true)
		// Go convention often names an unexported function: TestCleanComment
		// refers to cleanComment. Compare a title-cased variant of a lowercase
		// candidate, and let an exact case match outrank it.
		if camel && len(t.key) > 0 && isLowerByte(t.key[0]) {
			folded := string(t.key[0]-'a'+'A') + t.key[1:]
			consider(t, folded, false)
		}
	}
	if bestLen < 0 {
		return conventionTarget{}, nil, false
	}
	if len(best) > 1 {
		names := make([]string, len(best))
		for i, b := range best {
			names[i] = b.pkg + "/" + b.key
		}
		sort.Strings(names)
		return conventionTarget{}, names, false
	}
	return best[0], nil, true
}

// boundaryOK reports whether the text after a matched candidate is a valid
// word boundary: end of name or an underscore for both styles, and a
// CamelCase boundary for CamelCase names.
func boundaryOK(rest, cand string, camel bool) bool {
	if rest == "" {
		return true
	}
	if !camel {
		return rest[0] == '_'
	}
	if rest[0] == '_' {
		return true
	}
	if !isUpperByte(rest[0]) {
		return false
	}
	last := cand[len(cand)-1]
	if !isUpperByte(last) {
		return true
	}
	return len(rest) > 1 && isLowerByte(rest[1])
}

func isUpperByte(c byte) bool { return c >= 'A' && c <= 'Z' }
func isLowerByte(c byte) bool { return c >= 'a' && c <= 'z' }

// nativeKey is a symbol's disambiguation key inside its package. Methods are
// keyed "Receiver.Name" so a method and a same-named top-level function don't
// collapse into one requirement ID; everything else is keyed by name.
func nativeKey(s model.Symbol) string {
	if s.Receiver != "" {
		return s.Receiver + "." + s.Name
	}
	return s.Name
}

// countRequirements counts requirement-kind items in a document tree.
func countRequirements(items []exchange.Item) int {
	n := 0
	for _, it := range items {
		if it.Kind == exchange.KindRequirement {
			n++
		}
		n += countRequirements(it.Children)
	}
	return n
}

// withXReqmd returns a copy of doc whose x-reqmd block can be mutated without
// affecting the other documents built from the same template.
func withXReqmd(doc exchange.Document) exchange.Document {
	if doc.XReqmd == nil {
		doc.XReqmd = &exchange.XReqmd{}
		return doc
	}
	xr := *doc.XReqmd
	if doc.XReqmd.Upstream != nil {
		up := *doc.XReqmd.Upstream
		xr.Upstream = &up
	}
	doc.XReqmd = &xr
	return doc
}

// nestMethods attaches collected methods to their type, preserving source order.
// A method whose type lives in another file stays at its own file's level: the
// file section is the authoritative grouping.
func nestMethods(items []exchange.Item, childOf map[string][]exchange.Item) []exchange.Item {
	if len(childOf) == 0 {
		return items
	}
	for i := range items {
		if kids, ok := childOf[items[i].Key]; ok {
			items[i].Children = append(items[i].Children, kids...)
		}
	}
	return items
}

// relPath renders a source file relative to the scanned root, so a per-file
// section title is "parser.go" rather than an absolute build path. A file
// outside the root keeps its full path.
func relPath(root, file string) string {
	rel, err := filepath.Rel(root, file)
	if err != nil || rel == "" || strings.HasPrefix(rel, "..") {
		return file
	}
	return filepath.ToSlash(rel)
}

// docSections turns a doc comment into section items: the prose before the
// first heading becomes an untitled section (rendered as a plain prose block),
// and every heading becomes a child section carrying the prose beneath it.
// Headings are consumed here so no body ever contains one — the renderer's
// contract is that structure lives in items, never in prose. Fenced code
// blocks are skipped so a commented-out "#include" or a Markdown example is not
// mistaken for a heading.
func docSections(doc string) []exchange.Item {
	if strings.TrimSpace(doc) == "" {
		return nil
	}
	var out []exchange.Item
	var intro []string
	var cur *exchange.Item
	var body []string
	inFence := false

	flushBody := func() {
		if cur == nil {
			return
		}
		cur.Body = strings.TrimSpace(strings.Join(body, "\n"))
		body = nil
	}

	for line := range strings.SplitSeq(doc, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence
		}
		// Godoc separators (a rule of dashes, equals signs, stars or
		// underscores) are decoration. Worse, a rule directly under a text line
		// is a *setext heading* once rendered, so it would silently become a
		// section that swallows the symbols that follow.
		if !inFence && isRuleLine(trimmed) {
			continue
		}
		if !inFence {
			if level, title, ok := atxHeading(trimmed); ok && level >= 1 {
				flushBody()
				if cur != nil {
					out = append(out, *cur)
				}
				cur = &exchange.Item{Kind: exchange.KindSection, Title: title}
				continue
			}
		}
		if cur == nil {
			intro = append(intro, line)
			continue
		}
		body = append(body, line)
	}
	flushBody()
	if cur != nil {
		out = append(out, *cur)
	}
	if text := strings.TrimSpace(strings.Join(intro, "\n")); text != "" {
		// The intro is content, so it is an untitled section rather than a
		// heading: an outline entry per symbol would flood the sidebar.
		out = append([]exchange.Item{{Kind: exchange.KindSection, Body: text}}, out...)
	}
	return out
}

// atxHeading parses "## Title" into its level and text.
func atxHeading(line string) (int, string, bool) {
	n := 0
	for n < len(line) && line[n] == '#' {
		n++
	}
	if n == 0 || n > 6 || n == len(line) || line[n] != ' ' {
		return 0, "", false
	}
	return n, strings.TrimSpace(line[n+1:]), true
}

// isRuleLine reports whether a line is a decorative rule: three or more of the
// same Markdown rule character, nothing else.
func isRuleLine(line string) bool {
	if len(line) < 3 {
		return false
	}
	c := line[0]
	if c != '-' && c != '=' && c != '*' && c != '_' {
		return false
	}
	for i := range len(line) {
		if line[i] != c {
			return false
		}
	}
	return true
}
