package extractor

import (
	"reflect"
	"testing"
)

func TestExtractTraces(t *testing.T) {
	tests := []struct {
		name          string
		text          string
		useHeuristic  bool
		wantExplicit  []string
		wantHeuristic []string
	}{
		{
			name:          "empty input",
			text:          "",
			wantExplicit:  nil,
			wantHeuristic: nil,
		},
		{
			name:         "single explicit",
			text:         "reqmd:trace REQ-ARITH-001",
			wantExplicit: []string{"REQ-ARITH-001"},
		},
		{
			name:         "multiple explicit lines",
			text:         "reqmd:trace REQ-AUTH-001\nreqmd:trace REQ-AUTH-002",
			wantExplicit: []string{"REQ-AUTH-001", "REQ-AUTH-002"},
		},
		{
			name:         "qualified ref",
			text:         "reqmd:trace system/REQ-FEAT-007",
			wantExplicit: []string{"system/REQ-FEAT-007"},
		},
		{
			name:         "comma-separated fallback",
			text:         "reqmd:trace REQ-1, REQ-2, REQ-3",
			wantExplicit: []string{"REQ-1", "REQ-2", "REQ-3"},
		},
		{
			name:         "version pin",
			text:         "reqmd:trace REQ-XYZ-001~3",
			wantExplicit: []string{"REQ-XYZ-001~3"},
		},
		{
			name:         "invalid ID is silently skipped",
			text:         "reqmd:trace REQ-foo-bar\nreqmd:trace REQ-001",
			wantExplicit: []string{"REQ-001"},
		},
		{
			name:         "deduplication",
			text:         "reqmd:trace REQ-001\nreqmd:trace REQ-001",
			wantExplicit: []string{"REQ-001"},
		},
		{
			name:         "heuristic off by default",
			text:         "See REQ-AUTH-001 for details",
			wantExplicit: nil,
		},
		{
			name:          "heuristic on, no explicit",
			text:          "See REQ-AUTH-001 and SYS-LOG-002 for details",
			useHeuristic:  true,
			wantHeuristic: []string{"REQ-AUTH-001", "SYS-LOG-002"},
		},
		{
			name:          "heuristic does not double-count explicit",
			text:          "reqmd:trace REQ-AUTH-001\nSee REQ-AUTH-001 for details",
			useHeuristic:  true,
			wantExplicit:  []string{"REQ-AUTH-001"},
			wantHeuristic: nil, // dedup: explicit wins
		},
		{
			name:          "heuristic requires at least 3 digits",
			text:          "UTF-8 and HTTP-2 are protocols",
			useHeuristic:  true,
			wantHeuristic: nil, // UTF-8 and HTTP-2 have only 1 digit
		},
		{
			name:         "marker must be at line start",
			text:         "see also reqmd:trace REQ-001",
			wantExplicit: nil, // not at start of line, regex requires ^reqmd:trace
		},
		{
			name:         "prose in docstring with markers",
			text:         "Implements the auth flow.\n\nreqmd:trace REQ-AUTH-001\nreqmd:trace REQ-AUTH-002",
			wantExplicit: []string{"REQ-AUTH-001", "REQ-AUTH-002"},
		},
		{
			// All-letters ID (no digit segment) is accepted by the
			// permissive pre-filter, even though the strict canonical
			// pattern would reject it (a broken-ref WARNING at
			// `reqmd check` time is the right outcome for that).
			name:         "all-letters ID is accepted by loose pre-filter",
			text:         "reqmd:trace REQ-MATH-ADD",
			wantExplicit: []string{"REQ-MATH-ADD"},
		},
		{
			// Digit-ending ID still works (regression coverage for the
			// pre-existing canonical shape after the regex was loosened).
			name:         "digit-ending ID is still accepted",
			text:         "reqmd:trace REQ-AUTH-001",
			wantExplicit: []string{"REQ-AUTH-001"},
		},
		{
			name:         "multi-line explicit traces are all extracted in order",
			text:         "reqmd:trace REQ-AUTH-001\nreqmd:trace REQ-AUTH-002\nreqmd:trace REQ-AUTH-003",
			wantExplicit: []string{"REQ-AUTH-001", "REQ-AUTH-002", "REQ-AUTH-003"},
		},
		{
			name:         "comma-separated traces split correctly",
			text:         "reqmd:trace REQ-AUTH-001, REQ-AUTH-002, REQ-AUTH-003",
			wantExplicit: []string{"REQ-AUTH-001", "REQ-AUTH-002", "REQ-AUTH-003"},
		},
		{
			name:          "mix of multi-line and heuristic, explicit wins",
			text:          "reqmd:trace REQ-AUTH-001\nreqmd:trace REQ-AUTH-002\nThis doc also mentions REQ-AUTH-001 in prose.",
			useHeuristic:  true,
			wantExplicit:  []string{"REQ-AUTH-001", "REQ-AUTH-002"},
			wantHeuristic: nil, // REQ-AUTH-001 is already explicit; no heuristic duplicates
		},
		{
			name:         "order preserved across multi-line and comma",
			text:         "reqmd:trace REQ-Z, REQ-A\nreqmd:trace REQ-M",
			wantExplicit: []string{"REQ-Z", "REQ-A", "REQ-M"},
		},
		{
			name:         "duplicate across multi-line and comma is deduplicated",
			text:         "reqmd:trace REQ-AUTH-001\nreqmd:trace REQ-AUTH-001, REQ-AUTH-002",
			wantExplicit: []string{"REQ-AUTH-001", "REQ-AUTH-002"},
		},
		{
			name:         "qualified refs in multi-trace are all extracted",
			text:         "reqmd:trace system/REQ-FEAT-001\nreqmd:trace system/REQ-FEAT-002",
			wantExplicit: []string{"system/REQ-FEAT-001", "system/REQ-FEAT-002"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotExp, gotHeur := ExtractTraces(tc.text, tc.useHeuristic)
			if !reflect.DeepEqual(gotExp, tc.wantExplicit) {
				t.Errorf("explicit: got %v, want %v", gotExp, tc.wantExplicit)
			}
			if !reflect.DeepEqual(gotHeur, tc.wantHeuristic) {
				t.Errorf("heuristic: got %v, want %v", gotHeur, tc.wantHeuristic)
			}
		})
	}
}
