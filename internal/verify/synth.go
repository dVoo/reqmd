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

// Synthesize builds one model.Document holding one pseudo-requirement per
// merged result. Each pseudo-requirement:
//   - ID: RESULT:<measure-id-stripped> (e.g. RESULT:TST-UNT-001)
//   - trace: [<measure-id-with-pin>]  — so the existing graph builds a
//     result→measure edge and version-pin checks apply to the pin.
//   - outcome: <pass|fail|skipped|inconclusive>
//   - status: approved (so they are coverage providers for the
//     missing-verdict check)
//
// The returned Document is appended to the spec docs before graph.New so
// all existing graph machinery (broken-ref, circular, version-pin) works
// on result→measure edges with no new traversal code.
func Synthesize(merged map[string]Result) model.Document {
	// Deterministic order for stable output.
	ids := make([]string, 0, len(merged))
	for id := range merged {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	nodes := make([]*model.Node, 0, len(ids))
	for _, id := range ids {
		r := merged[id]
		nodes = append(nodes, synthesizeOne(id, r))
	}

	return model.Document{
		Path:       "<verify-results>", // synthetic; not a real dir
		Nodes:      nodes,
		XReqmd:     &model.XReqmd{Level: resultDocLevel},
		Properties: []string{"outcome", model.AttrTrace, model.AttrStatus, "verifier", "evidence", "verified-at"},
	}
}

// synthesizeOne builds a single pseudo-requirement node from a result.
// The measure ID (with its `~N` pin preserved) becomes the trace target.
func synthesizeOne(measureID string, r Result) *model.Node {
	attrs := map[string]any{
		"outcome": string(r.Outcome),
		// trace with the original (pinned) measure ID so version-pin
		// checks fire on stale pins.
		model.AttrTrace:  []any{r.MeasureID},
		model.AttrStatus: model.StatusApproved,
	}
	if r.Verifier != "" {
		attrs["verifier"] = r.Verifier
	}
	if r.Evidence != "" {
		attrs["evidence"] = r.Evidence
	}
	if !r.VerifiedAt.IsZero() {
		attrs["verified-at"] = r.VerifiedAt.Format(time.DateOnly)
	}

	return &model.Node{
		Kind:   model.KindRequirement,
		ID:     synthID(measureID),
		Title:  outcomeTitle(r),
		Attrs:  attrs,
		Source: r.Source,
	}
}

// synthID builds the synthetic node ID for a result. The measure ID is
// stripped of its `~N` pin so multiple results for the same measure
// collapse to one pseudo-requirement.
func synthID(measureID string) string {
	bare, _, _ := model.StripPin(measureID)
	return "RESULT:" + bare
}

// outcomeTitle returns a short human-readable title for the pseudo-req.
func outcomeTitle(r Result) string {
	bare, _, _ := model.StripPin(r.MeasureID)
	return fmt.Sprintf("verification result: %s = %s", bare, r.Outcome)
}

// IsResultNode reports whether a requirement ID belongs to a synthesized
// verification-result pseudo-requirement (i.e. produced by Synthesize).
// Used by graph checks to distinguish result nodes from authored ones.
func IsResultNode(reqID string) bool {
	return strings.HasPrefix(reqID, "RESULT:")
}

// ResultOutcome extracts the outcome attribute from a synthesized result
// node's requirement attrs. Returns "" if not present.
func ResultOutcome(attrs map[string]any) Outcome {
	v, ok := attrs["outcome"].(string)
	if !ok {
		return ""
	}
	return Outcome(v)
}
