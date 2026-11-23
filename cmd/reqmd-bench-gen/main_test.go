package main

import "testing"

// The source spec keeps most requirement IDs on h3/h4 headings, so every
// heading level the parser accepts as a requirement must be discovered.
func TestHeadingIDReMatchesRequirementHeadingLevels(t *testing.T) {
	cases := []struct {
		line string
		id   string
	}{
		{"## SYS-001: top level", "SYS-001"},
		{"### SW-EXP-004: nested", "SW-EXP-004"},
		{"#### DEEP-1 — nested deeper", "DEEP-1"},
		{"##### TST-UNT-001: five", "TST-UNT-001"},
		{"###### SIX-1: six", "SIX-1"},
		// The generated code proxies carry mixed-case, underscore IDs.
		{"#### IMP-cli-extract: symbol", "IMP-cli-extract"},
		{"#### IMP-lang_test-1a2b3c: symbol", "IMP-lang_test-1a2b3c"},
	}
	for _, c := range cases {
		m := headingIDRe.FindStringSubmatch(c.line)
		if m == nil {
			t.Errorf("no match for %q", c.line)
			continue
		}
		if m[1] != c.id {
			t.Errorf("%q: id = %q, want %q", c.line, m[1], c.id)
		}
	}

	for _, line := range []string{
		"# TITLE-1: document title", // h1 is the doc title, never a requirement
		"####### SEVEN-1: not a heading",
		"### lowercase-1: not an ID",
		"###NoSpace-1: needs whitespace",
		"### SYS.1: dotted section heading", // no dash, so not an ID
	} {
		if m := headingIDRe.FindStringSubmatch(line); m != nil {
			t.Errorf("%q: unexpected match %q", line, m[0])
		}
	}
}

func TestApplyRewritePreservesHeadingLevel(t *testing.T) {
	body := "### SYS-002: child\n```attr\ntrace: [SYS-001]\n```\n"
	got := applyRewrite(body, "doc.md", map[string]string{"SYS-002": "T00-SYS-002", "SYS-001": "T00-SYS-001"}, "T00")
	want := "### T00-SYS-002: child\n```attr\ntrace: [T00-SYS-001]\n```\n"
	if got != want {
		t.Fatalf("applyRewrite =\n%q\nwant\n%q", got, want)
	}
}

// A document's declared prefix must receive the copy tag exactly once. Tagging
// it twice makes every requirement in that document fail validation.
func TestApplyRewriteTagsSchemaPrefixOnce(t *testing.T) {
	body := "x-reqmd:\n  id-prefix: SYS-\n"
	got := applyRewrite(body, "schema.yaml", map[string]string{"SYS": "T00-SYS"}, "T00")
	want := "x-reqmd:\n  id-prefix: T00-SYS-\n"
	if got != want {
		t.Fatalf("applyRewrite(schema.yaml) =\n%q\nwant\n%q", got, want)
	}
}

func TestRewriteBlockForCopyPreservesHeadingLevel(t *testing.T) {
	block := "#### SW-EXP-004: diagram\n```attr\ntrace: [SYS-001~2]\n```\n"
	reqIDs := map[string]bool{"SW-EXP-004": true, "SYS-001": true}
	got := rewriteBlockForCopy(block, 1, reqIDs)
	want := "#### SW-EXP-004_c1: diagram\n```attr\ntrace: [SYS-001_c1~2]\n```\n"
	if got != want {
		t.Fatalf("rewriteBlockForCopy =\n%q\nwant\n%q", got, want)
	}
}

// A prose heading that happens to contain a dash is not a requirement and must
// be repeated verbatim.
func TestRewriteBlockForCopySkipsProseHeading(t *testing.T) {
	block := "## V-Model overview\n\nProse only.\n"
	got := rewriteBlockForCopy(block, 1, map[string]bool{"SYS-001": true})
	if got != block {
		t.Fatalf("rewriteBlockForCopy =\n%q\nwant unchanged\n%q", got, block)
	}
}
