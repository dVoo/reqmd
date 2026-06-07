package diff

import (
	"reflect"
	"sort"
	"strings"

	r3diff "github.com/r3labs/diff/v3"

	"reqmd/internal/model"
)

// DiffCategory describes what happened to a requirement between baselines.
type DiffCategory string

const (
	Added     DiffCategory = "added"
	Removed   DiffCategory = "removed"
	Modified  DiffCategory = "modified"
	Unchanged DiffCategory = "unchanged"
)

// AttrChange records a single attribute change within a requirement.
type AttrChange struct {
	Key    string
	OldVal any // nil if attribute was added
	NewVal any // nil if attribute was removed
}

// ReqDiff is the diff result for a single requirement.
type ReqDiff struct {
	ID          string
	Category    DiffCategory
	Title       string
	DocPath     string
	AttrChanges []AttrChange
}

// DocDiff is the diff result for a single document directory.
type DocDiff struct {
	Path      string
	Added     int
	Removed   int
	Modified  int
	Unchanged int
	Reqs      []ReqDiff
}

// SchemaChange describes a single property change in a schema.
type SchemaChange struct {
	Path   string // e.g. "properties.priority.enum"
	OldVal any
	NewVal any
}

// SchemaDiff is the diff result for a single document's schema.
type SchemaDiff struct {
	Path    string
	Changes []SchemaChange
	Added   []string // new top-level keys
	Removed []string // removed top-level keys
}

// SubmoduleChange describes a submodule's pinned-commit change between two baselines.
type SubmoduleChange struct {
	Path   string `json:"path"`
	OldSHA string `json:"old_sha,omitempty"` // empty if submodule was added
	NewSHA string `json:"new_sha,omitempty"` // empty if submodule was removed
	Status string `json:"status"`            // "added" | "removed" | "updated"
}

// Result is the top-level diff result between two baselines.
type Result struct {
	Tag1, Tag2                          string
	Added, Removed, Modified, Unchanged int
	Documents                           []DocDiff
	Schemas                             []SchemaDiff
	Submodules                          []SubmoduleChange
}

// Diff compares two sets of parsed documents and returns structured diff results.
// docs1 is the older baseline (tag1), docs2 is the newer baseline (tag2).
func Diff(docs1, docs2 []model.Document, schemas1, schemas2 map[string]map[string]any, tag1, tag2 string) *Result {
	// Build maps: reqID → Requirement, grouped by doc path
	map1 := buildReqMap(docs1)
	map2 := buildReqMap(docs2)

	result := &Result{Tag1: tag1, Tag2: tag2}

	differ, _ := r3diff.NewDiffer()

	// Collect all document paths
	allPaths := make(map[string]bool)
	for p := range map1 {
		allPaths[p] = true
	}
	for p := range map2 {
		allPaths[p] = true
	}

	// Sort paths for deterministic output
	paths := make([]string, 0, len(allPaths))
	for p := range allPaths {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	for _, docPath := range paths {
		requids1 := map1[docPath]
		requids2 := map2[docPath]

		dd := DocDiff{Path: docPath}

		// All req IDs in this document across both baselines
		allIDs := make(map[string]bool)
		for id := range requids1 {
			allIDs[id] = true
		}
		for id := range requids2 {
			allIDs[id] = true
		}

		ids := make([]string, 0, len(allIDs))
		for id := range allIDs {
			ids = append(ids, id)
		}
		sort.Strings(ids)

		for _, id := range ids {
			r1, ok1 := requids1[id]
			r2, ok2 := requids2[id]

			var rd ReqDiff
			rd.ID = id
			rd.DocPath = docPath

			switch {
			case ok1 && !ok2:
				rd.Category = Removed
				rd.Title = r1.Title
				dd.Removed++
				result.Removed++

			case !ok1 && ok2:
				rd.Category = Added
				rd.Title = r2.Title
				dd.Added++
				result.Added++

			default:
				// Both exist — diff attributes
				rd.Title = r2.Title
				changes := diffAttrs(differ, r1.Attrs, r2.Attrs)
				if len(changes) > 0 {
					rd.Category = Modified
					rd.AttrChanges = changes
					dd.Modified++
					result.Modified++
				} else {
					rd.Category = Unchanged
					dd.Unchanged++
					result.Unchanged++
				}
			}

			dd.Reqs = append(dd.Reqs, rd)
		}

		result.Documents = append(result.Documents, dd)
	}

	result.Schemas = DiffSchemasWithDiffer(differ, schemas1, schemas2)

	return result
}

// DiffSchemas compares two sets of schemas and returns per-document schema diffs.
func DiffSchemas(schemas1, schemas2 map[string]map[string]any) []SchemaDiff {
	differ, _ := r3diff.NewDiffer()
	return DiffSchemasWithDiffer(differ, schemas1, schemas2)
}

// DiffSchemasWithDiffer compares two sets of schemas using the provided differ.
func DiffSchemasWithDiffer(differ *r3diff.Differ, schemas1, schemas2 map[string]map[string]any) []SchemaDiff {
	allPaths := make(map[string]bool)
	for p := range schemas1 {
		allPaths[p] = true
	}
	for p := range schemas2 {
		allPaths[p] = true
	}

	var results []SchemaDiff
	for path := range allPaths {
		s1 := schemas1[path]
		s2 := schemas2[path]

		sd := SchemaDiff{Path: path}

		if s1 == nil && s2 != nil {
			sd.Added = topLevelKeys(s2)
		} else if s1 != nil && s2 == nil {
			sd.Removed = topLevelKeys(s1)
		} else if s1 != nil && s2 != nil {
			if differ == nil {
				continue
			}
			changelog, err := differ.Diff(s1, s2)
			if err == nil {
				for _, c := range changelog {
					if len(c.Path) == 0 {
						continue
					}
					sd.Changes = append(sd.Changes, SchemaChange{
						Path:   strings.Join(c.Path, "."),
						OldVal: c.From,
						NewVal: c.To,
					})
				}
			}
		}

		if len(sd.Changes) > 0 || len(sd.Added) > 0 || len(sd.Removed) > 0 {
			results = append(results, sd)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Path < results[j].Path
	})
	return results
}

func topLevelKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// buildReqMap groups requirements by document path.
func buildReqMap(docs []model.Document) map[string]map[string]model.Requirement {
	result := make(map[string]map[string]model.Requirement)
	for _, doc := range docs {
		byID := make(map[string]model.Requirement)
		for _, req := range doc.Requirements {
			byID[req.ID] = req
		}
		result[doc.Path] = byID
	}
	return result
}

// diffAttrs compares two attribute maps using r3labs/diff and returns changes.
func diffAttrs(differ *r3diff.Differ, old, new map[string]any) []AttrChange {
	if len(old) == 0 && len(new) == 0 {
		return nil
	}

	if differ == nil {
		return diffAttrsManual(old, new)
	}
	changelog, err := differ.Diff(old, new)
	if err != nil {
		// Fallback: manual comparison if r3labs fails on complex types
		return diffAttrsManual(old, new)
	}

	var changes []AttrChange
	for _, c := range changelog {
		key := c.Path[len(c.Path)-1] // last element is the key name
		changes = append(changes, AttrChange{
			Key:    key,
			OldVal: c.From,
			NewVal: c.To,
		})
	}
	return changes
}

// diffAttrsManual is a fallback for when r3labs can't handle the types.
func diffAttrsManual(old, new map[string]any) []AttrChange {
	allKeys := make(map[string]bool)
	for k := range old {
		allKeys[k] = true
	}
	for k := range new {
		allKeys[k] = true
	}

	keys := make([]string, 0, len(allKeys))
	for k := range allKeys {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var changes []AttrChange
	for _, k := range keys {
		oval, ok1 := old[k]
		nval, ok2 := new[k]
		if ok1 != ok2 || !reflect.DeepEqual(oval, nval) {
			changes = append(changes, AttrChange{
				Key:    k,
				OldVal: oval,
				NewVal: nval,
			})
		}
	}
	return changes
}

// DiffSubmodules compares two maps of submodule path → commit SHA and returns
// a sorted list of SubmoduleChange. Submodules with the same SHA at both
// tags are omitted (no noise). Result is sorted by path for deterministic output.
func DiffSubmodules(subs1, subs2 map[string]string) []SubmoduleChange {
	seen := make(map[string]bool, len(subs1)+len(subs2))
	all := make([]string, 0, len(subs1)+len(subs2))

	// Collect all submodule paths
	for path := range subs1 {
		if !seen[path] {
			seen[path] = true
			all = append(all, path)
		}
	}
	for path := range subs2 {
		if !seen[path] {
			seen[path] = true
			all = append(all, path)
		}
	}

	// Sort for deterministic output
	sort.Strings(all)

	changes := make([]SubmoduleChange, 0)
	for _, path := range all {
		oldSHA, ok1 := subs1[path]
		newSHA, ok2 := subs2[path]

		switch {
		case !ok1 && ok2:
			// Submodule added
			changes = append(changes, SubmoduleChange{
				Path:   path,
				NewSHA: newSHA,
				Status: "added",
			})
		case ok1 && !ok2:
			// Submodule removed
			changes = append(changes, SubmoduleChange{
				Path:   path,
				OldSHA: oldSHA,
				Status: "removed",
			})
		case ok1 && ok2 && oldSHA != newSHA:
			// Submodule updated (SHA changed)
			changes = append(changes, SubmoduleChange{
				Path:   path,
				OldSHA: oldSHA,
				NewSHA: newSHA,
				Status: "updated",
			})
		}
		// Same SHA at both tags → omit (no noise)
	}

	return changes
}
