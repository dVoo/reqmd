// m1harness is a one-shot CLI that exercises the extract pipeline end-to-end
// against a single Go source file or directory of .go files. It parses via
// the lang/go plugin, groups symbols by package, and writes per-package
// .md files + the standard schema.yaml via writer.WriteTarget.
//
// Usage:  go run ./cmd/m1harness/ <go-file-or-dir> <output-dir>
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"reqmd-import/internal/lang"
	_ "reqmd-import/internal/lang/go" // register Go plugin via init()
	"reqmd-import/internal/model"
	"reqmd-import/internal/writer"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: m1harness <input> <output-dir>\n")
		os.Exit(2)
	}
	input, outDir := os.Args[1], os.Args[2]

	plug, ok := lang.Get("go")
	if !ok {
		fmt.Fprintln(os.Stderr, "Go language plugin not registered")
		os.Exit(1)
	}

	files, err := discoverGoFiles(input)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var allSymbols []model.Symbol
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read %s: %v\n", f, err)
			os.Exit(1)
		}
		syms, err := plug.Parse(f, src)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse %s: %v\n", f, err)
			os.Exit(1)
		}
		allSymbols = append(allSymbols, syms...)
	}

	pkgMap := make(map[string][]model.Symbol)
	for _, s := range allSymbols {
		pkgMap[s.Package] = append(pkgMap[s.Package], s)
	}

	packages := make([]writer.Package, 0, len(pkgMap))
	for pkgName, syms := range pkgMap {
		// Receiver-aware key: a method (T).multiply and a top-level
		// multiply() in the same package must not collapse to the same ID.
		idFor := make(map[string]string, len(syms))
		for _, s := range syms {
			hashKey := s.Name
			if s.Receiver != "" {
				hashKey = s.Receiver + "." + s.Name
			}
			idFor[hashKey] = "IMP-" + pkgName + "-" + s.Name + "-" + model.ShortHash(pkgName, hashKey)
		}
		sort.SliceStable(syms, func(i, j int) bool { return syms[i].StartLine < syms[j].StartLine })
		var reqs []model.GeneratedReq
		for _, s := range syms {
			selfKey := s.Name
			if s.Receiver != "" {
				selfKey = s.Receiver + "." + s.Name
			}
			req := model.GeneratedReq{
				ID:         idFor[selfKey],
				Title:      s.Name,
				Status:     "approved",
				Body:       s.Doc,
				SourceFile: s.File,
				SourceLine: s.StartLine,
				SymbolKind: string(s.Kind),
				Symbol:     s,
			}
			if s.Kind == model.SymbolMethod && s.Receiver != "" {
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
	sort.Slice(packages, func(i, j int) bool { return packages[i].Name < packages[j].Name })

	if err := writer.WriteTarget(outDir, packages, "IMP-"); err != nil {
		fmt.Fprintf(os.Stderr, "write: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Wrote %d package(s) to %s\n", len(packages), outDir)
}

// discoverGoFiles returns the sorted list of .go files under input. If
// input is a file, it must end in .go and is returned as a single-element
// slice. If input is a directory, all non-directory entries with a .go
// suffix are returned.
func discoverGoFiles(input string) ([]string, error) {
	st, err := os.Stat(input)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", input, err)
	}
	if !st.IsDir() {
		if !strings.HasSuffix(input, ".go") {
			return nil, fmt.Errorf("input %q is not a .go file or directory", input)
		}
		return []string{input}, nil
	}
	entries, err := os.ReadDir(input)
	if err != nil {
		return nil, fmt.Errorf("read dir %s: %w", input, err)
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".go") {
			files = append(files, filepath.Join(input, e.Name()))
		}
	}
	sort.Strings(files)
	return files, nil
}
