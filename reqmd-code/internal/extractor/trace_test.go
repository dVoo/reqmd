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
		wantReqIDs    []string
		wantSymbols   []string
		wantMalformed []string
		wantOptOut    bool
		wantHeuristic []string
	}{
		{
			name: "empty input",
			text: "",
		},
		{
			name:       "authored ID",
			text:       "reqmd:trace REQ-ARITH-001",
			wantReqIDs: []string{"REQ-ARITH-001"},
		},
		{
			name:       "all-letters authored ID",
			text:       "reqmd:trace REQ-MATH-ADD",
			wantReqIDs: []string{"REQ-MATH-ADD"},
		},
		{
			name:       "qualified ID",
			text:       "reqmd:trace system/REQ-FEAT-007",
			wantReqIDs: []string{"system/REQ-FEAT-007"},
		},
		{
			name:       "version pin",
			text:       "reqmd:trace REQ-XYZ-001~3",
			wantReqIDs: []string{"REQ-XYZ-001~3"},
		},
		{
			name:       "generated proxy ID",
			text:       "reqmd:trace IMP-clang-Parse-479c6b0",
			wantReqIDs: []string{"IMP-clang-Parse-479c6b0"},
		},
		{
			name:       "qualified generated ID",
			text:       "reqmd:trace graph/IMP-graph-New-1234567",
			wantReqIDs: []string{"graph/IMP-graph-New-1234567"},
		},
		{
			name:        "bare symbol",
			text:        "reqmd:trace Parse",
			wantSymbols: []string{"Parse"},
		},
		{
			name:        "receiver-qualified symbol",
			text:        "reqmd:trace Parser.Parse",
			wantSymbols: []string{"Parser.Parse"},
		},
		{
			name:        "c++ scope separator",
			text:        "reqmd:trace Parser::Parse",
			wantSymbols: []string{"Parser::Parse"},
		},
		{
			name:        "package-qualified symbol",
			text:        "reqmd:trace clang/Parse",
			wantSymbols: []string{"clang/Parse"},
		},
		{
			name:        "lowercase typo is a symbol ref, not silently dropped",
			text:        "reqmd:trace SW-imp-001",
			wantSymbols: []string{"SW-imp-001"},
		},
		{
			name:        "hyphenless uppercase token is a symbol",
			text:        "reqmd:trace HTTP",
			wantSymbols: []string{"HTTP"},
		},
		{
			name:       "none is an opt-out",
			text:       "reqmd:trace none",
			wantOptOut: true,
		},
		{
			name:        "qualified none is a symbol",
			text:        "reqmd:trace pkg/none",
			wantSymbols: []string{"pkg/none"},
		},
		{
			name:          "prose on the directive is malformed",
			text:          "reqmd:trace see the parser",
			wantMalformed: []string{"see the parser"},
		},
		{
			name:        "mixed targets",
			text:        "reqmd:trace REQ-AUTH-001, Parse, none, Parser.Parse",
			wantReqIDs:  []string{"REQ-AUTH-001"},
			wantSymbols: []string{"Parse", "Parser.Parse"},
			wantOptOut:  true,
		},
		{
			name:       "single stray space is tolerated",
			text:       " reqmd:trace REQ-001",
			wantReqIDs: []string{"REQ-001"},
		},
		{
			name: "indented code example is not a marker",
			text: "    reqmd:trace REQ-DEMO-001\n\treqmd:trace REQ-DEMO-002",
		},
		{
			name:       "CRLF line ending",
			text:       "reqmd:trace REQ-001\r\n",
			wantReqIDs: []string{"REQ-001"},
		},
		{
			name:        "deduplication",
			text:        "reqmd:trace REQ-001\nreqmd:trace REQ-001, Parse\nreqmd:trace Parse",
			wantReqIDs:  []string{"REQ-001"},
			wantSymbols: []string{"Parse"},
		},
		{
			name:          "heuristic off by default",
			text:          "See REQ-AUTH-001 for details",
			wantHeuristic: nil,
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
			wantReqIDs:    []string{"REQ-AUTH-001"},
			wantHeuristic: nil,
		},
		{
			name:          "heuristic requires at least 3 digits",
			text:          "UTF-8 and HTTP-2 are protocols",
			useHeuristic:  true,
			wantHeuristic: nil,
		},
		{
			name: "marker must start a line",
			text: "see also reqmd:trace REQ-001",
		},
		{
			name:        "order preserved across lines and commas",
			text:        "reqmd:trace REQ-Z, Parse\nreqmd:trace REQ-M",
			wantReqIDs:  []string{"REQ-Z", "REQ-M"},
			wantSymbols: []string{"Parse"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ExtractTraces(tc.text, tc.useHeuristic)
			if !reflect.DeepEqual(got.ReqIDs, tc.wantReqIDs) {
				t.Errorf("ReqIDs: got %v, want %v", got.ReqIDs, tc.wantReqIDs)
			}
			if !reflect.DeepEqual(got.SymbolRefs, tc.wantSymbols) {
				t.Errorf("SymbolRefs: got %v, want %v", got.SymbolRefs, tc.wantSymbols)
			}
			if !reflect.DeepEqual(got.Malformed, tc.wantMalformed) {
				t.Errorf("Malformed: got %v, want %v", got.Malformed, tc.wantMalformed)
			}
			if got.OptOut != tc.wantOptOut {
				t.Errorf("OptOut: got %v, want %v", got.OptOut, tc.wantOptOut)
			}
			if !reflect.DeepEqual(got.Heuristic, tc.wantHeuristic) {
				t.Errorf("Heuristic: got %v, want %v", got.Heuristic, tc.wantHeuristic)
			}
		})
	}
}
