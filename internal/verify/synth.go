package verify

import (
	"fmt"
	"reqmd/internal/model"
	"sort"
	"strings"
	"time"
)

// resultDocLevel is the x-reqmd level assigned to synthesized
// verification-result documents. The graph uses levels to infer V-model
// boundaries; results are the bottom of the V (leaves), so they must not
// be treated as roots.
const resultDocLevel = "verify-results"

// testCasePrefix prefixes the ID of a synthesized test-case node
// ("TC:<case-key>"). Case-keyed CTRF entries (patterns D/E) synthesize one
// such node per case; it carries the `verifies` trace edges to the
// requirements it exercises and the Markdown description body.
const testCasePrefix = "TC:"

// Synthesize builds one model.Document (marked Synthetic) holding the
// pseudo-requirements derived from merged results:
//
//   - One result node per merged (target, case) key:
//     ID RESULT:<target>[#<case>], trace: [<pinned target ref>], plus
//     outcome / status:approved / verifier / evidence / verified-at.
//   - One test-case node per synthesized case target (ID TC:<case>) whose
//     ID starts with testCasePrefix: trace: [verifies...] (pins allowed),
//     Body = the Markdown description of the winning run, Title = the test
//     name.
//
// The returned Document is appended to the spec docs before graph.New so
// all existing graph machinery (broken-ref, circular, version-pin) works on
// result→measure and case→requirement edges with no new traversal code.
func Synthesize(merged map[ResultKey]Result) model.Document {
	keys := make([]ResultKey, 0, len(merged))
	for k := range merged {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Target != keys[j].Target {
			return keys[i].Target < keys[j].Target
		}
		return keys[i].Case < keys[j].Case
	})

	var nodes []*model.Node
	tcInfo := make(map[string]Result) // TC id → result carrying description
	for _, k := range keys {
		if strings.HasPrefix(k.Target, testCasePrefix) {
			r := merged[k]
			if cur, ok := tcInfo[k.Target]; !ok || (cur.Description == "" && r.Description != "") {
				tcInfo[k.Target] = r
			}
		}
	}
	for tcID := range tcInfo {
		nodes = append(nodes, synthesizeTestCase(tcID, tcInfo[tcID]))
	}
	for _, k := range keys {
		nodes = append(nodes, synthesizeResult(k, merged[k]))
	}

	return model.Document{
		Path:       "<verify-results>", // synthetic; not a real dir
		Nodes:      nodes,
		XReqmd:     &model.XReqmd{Level: resultDocLevel},
		Properties: []string{model.AttrResultOutcome, model.AttrTrace, model.AttrStatus, model.AttrResultVerifier, model.AttrResultEvidence, model.AttrResultVerifiedAt},
		Synthetic:  true,
	}
}

// synthesizeTestCase builds the synthesized test-case node for a
// case-keyed CTRF result. Its trace edges go to the requirements it
// exercises (extra.x-reqmd.verifies, pins allowed); its body is the
// Markdown description taken from the winning run.
func synthesizeTestCase(tcID string, r Result) *model.Node {
	attrs := map[string]any{
		model.AttrTrace: toAny(r.Verifies),
	}
	return &model.Node{
		Kind:   model.KindRequirement,
		ID:     tcID,
		Title:  r.Name,
		Body:   r.Description,
		Attrs:  attrs,
		Source: r.Source,
	}
}

// synthesizeResult builds a single pseudo-requirement node from a merged
// result. The binding target (with its `~N` pin preserved for id-bound
// results) becomes the trace target; a case-keyed result gets a distinct
// node ID so several cases can attach to the same authored measure.
func synthesizeResult(k ResultKey, r Result) *model.Node {
	attrs := map[string]any{
		model.AttrResultOutcome: string(r.Outcome),
		// trace with the original (pinned) target so version-pin checks
		// fire on stale pins.
		model.AttrTrace:  []any{r.MeasureID},
		model.AttrStatus: model.StatusApproved,
	}
	if r.Verifier != "" {
		attrs[model.AttrResultVerifier] = r.Verifier
	}
	if r.Evidence != "" {
		attrs[model.AttrResultEvidence] = r.Evidence
	}
	if !r.VerifiedAt.IsZero() {
		attrs[model.AttrResultVerifiedAt] = r.VerifiedAt.Format(time.DateOnly)
	}

	id := resultNodeID(k)
	return &model.Node{
		Kind:   model.KindRequirement,
		ID:     id,
		Title:  outcomeTitle(k),
		Attrs:  attrs,
		Source: r.Source,
	}
}

// resultNodeID is the synthetic ID of a result node:
// RESULT:<target> for the single-result model, RESULT:<target>#<case> when
// the run is case-keyed.
func resultNodeID(k ResultKey) string {
	if k.Case == "" {
		return "RESULT:" + k.Target
	}
	return "RESULT:" + k.Target + "#" + k.Case
}

// outcomeTitle returns a short human-readable title for the pseudo-req.
func outcomeTitle(k ResultKey) string {
	if k.Case == "" {
		return fmt.Sprintf("verification result: %s", k.Target)
	}
	return fmt.Sprintf("verification result: %s#%s", k.Target, k.Case)
}

// toAny converts a []string into a []any for attribute maps.
func toAny(vals []string) []any {
	out := make([]any, 0, len(vals))
	for _, v := range vals {
		out = append(out, v)
	}
	return out
}

// IsResultNode reports whether a requirement ID belongs to a synthesized
// verification-result pseudo-requirement (i.e. produced by Synthesize).
// Used by graph checks to distinguish result nodes from authored ones.
func IsResultNode(reqID string) bool {
	return strings.HasPrefix(reqID, "RESULT:")
}

// IsTestCaseNode reports whether a requirement ID belongs to a synthesized
// test case ("TC:..."). Used to recognize case nodes independently of the
// RESULT prefix.
func IsTestCaseNode(reqID string) bool {
	return strings.HasPrefix(reqID, testCasePrefix)
}

// ResultOutcome extracts the outcome attribute from a synthesized result
// node's requirement attrs. Returns "" if not present.
func ResultOutcome(attrs map[string]any) Outcome {
	v, ok := attrs[model.AttrResultOutcome].(string)
	if !ok {
		return ""
	}
	return Outcome(v)
}
