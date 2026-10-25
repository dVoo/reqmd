package parser

import (
	"testing"

	"gopkg.in/yaml.v3"
)

// benchAttr is a representative minimal attr block: two scalars, one int, one
// list with pins. Both benchmarks decode it so the scanner's advantage stays
// measurable — it is ~15x faster than yaml.v3 (1784 vs 27608 ns/op, 14 vs 140
// allocs) and that is why the fast path exists at all. If the gap ever closes,
// delete the scanner and keep yaml.v3.
const benchAttr = "status: approved\nversion: 3\nverify: test\ndisposition: implemented\ntrace:\n  - STK-001~2\n  - SYS-014\n  - SWE-207\ntitle: Some requirement title here\n"

func BenchmarkDecodeAttrs(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		if _, err := decodeAttrs(benchAttr); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeAttrsYAMLv3(b *testing.B) {
	b.ReportAllocs()
	src := []byte(benchAttr)
	for b.Loop() {
		var m map[string]any
		if err := yaml.Unmarshal(src, &m); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseMD measures the whole file→content-tree path, which is what
// `reqmd check`/`ls` spend nearly all their time in.
func BenchmarkParseMD(b *testing.B) {
	for _, size := range []struct {
		name  string
		bytes int
	}{
		{"small-64KiB", 64 << 10},
		{"chunked-1MiB", 1 << 20},
	} {
		src := syntheticDoc(size.bytes)
		b.Run(size.name, func(b *testing.B) {
			b.SetBytes(int64(len(src)))
			b.ReportAllocs()
			for b.Loop() {
				if _, _, _, err := parseMD(src, "bench.md"); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// syntheticDoc builds a document of roughly n bytes made of root-level
// requirements with attr blocks — the shape that exercises both the
// root-split pre-scan (above parseParallelThreshold) and the chunked
// parallel parse.
func syntheticDoc(n int) []byte {
	var b []byte
	b = append(b, "# Bench document\n\n"...)
	for i := len(b); i < n; i = len(b) {
		b = append(b, "## BENCH-000: A requirement heading used for benchmarking\n```attr\nstatus: approved\nversion: 1\ntrace:\n  - STK-001\n```\n\nShort prose body for the requirement.\n\n"...)
	}
	return b
}
