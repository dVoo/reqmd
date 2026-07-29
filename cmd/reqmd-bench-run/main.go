// Synthetic benchmark runner for reqmd.
//
// Generates a scaled corpus from the spec/ tree and measures wall time,
// peak RSS, and throughput for `check`, `check --json`, `ls`, `stats`,
// and `export html` at each scale.
//
// Each scale is N copies of the source merged into one root. With the
// spec/ tree (6 doc dirs), 6 copies give 36 documents. Per-doc content
// is replicated to hit a target reqs/doc count.
//
// Output: JSON on stdout for downstream parsing.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
)

type scale struct {
	Label      string
	Copies     int
	ReqsPerDoc int
}

type measurement struct {
	Scale      string  `json:"scale"`
	Copies     int     `json:"copies"`
	ReqsPerDoc int     `json:"reqs_per_doc"`
	TotalReqs  int     `json:"total_reqs"`
	Docs       int     `json:"docs"`
	Command    string  `json:"command"`
	WallMs     int64   `json:"wall_ms"`
	PeakRSSKB  int64   `json:"peak_rss_kb"`
	ReqsPerSec float64 `json:"reqs_per_sec"`
	ExitCode   int     `json:"exit_code"`
}

var peakRSS atomic.Int64

func main() {
	src := flag.String("src", "", "source spec dir (the real one)")
	gen := flag.String("gen", "", "path to reqmd-bench-gen binary")
	reqmd := flag.String("reqmd", "", "path to reqmd binary")
	out := flag.String("out", "", "output JSON file")
	work := flag.String("work", "/tmp/reqmd-bench", "scratch directory")
	scalesFlag := flag.String("scales", "2000,5000,10000,20000", "comma-separated reqs/doc targets to measure at")
	flag.Parse()

	if *src == "" || *gen == "" || *reqmd == "" {
		fmt.Fprintln(os.Stderr, "usage: -src <spec> -gen <bench-gen> -reqmd <reqmd> [-out file.json] [-work dir] [-scales 2000,5000,...]")
		os.Exit(2)
	}

	var reqsPerDoc []int
	for _, s := range strings.Split(*scalesFlag, ",") {
		var n int
		fmt.Sscanf(strings.TrimSpace(s), "%d", &n)
		if n > 0 {
			reqsPerDoc = append(reqsPerDoc, n)
		}
	}
	if len(reqsPerDoc) == 0 {
		reqsPerDoc = []int{2000, 5000, 10000, 20000}
	}

	scales := make([]scale, len(reqsPerDoc))
	for i, n := range reqsPerDoc {
		scales[i] = scale{
			Label:      fmt.Sprintf("%dk reqs/doc", n/1000),
			Copies:     6, // 6 × spec doc dirs = 36 documents
			ReqsPerDoc: n,
		}
	}

	os.RemoveAll(*work)
	if err := os.MkdirAll(*work, 0o755); err != nil {
		panic(err)
	}

	var results []measurement
	for _, s := range scales {
		fmt.Fprintf(os.Stderr, "\n>>> scale %s: %d copies, %d reqs/doc\n", s.Label, s.Copies, s.ReqsPerDoc)
		root := filepath.Join(*work, fmt.Sprintf("reqs%d", s.ReqsPerDoc))
		if err := os.MkdirAll(root, 0o755); err != nil {
			panic(err)
		}
		genArgs := []string{"-src", *src, "-dst", root, "-copies", fmt.Sprintf("%d", s.Copies)}
		if s.ReqsPerDoc > 0 {
			genArgs = append(genArgs, "-reqsperdoc", fmt.Sprintf("%d", s.ReqsPerDoc))
		}
		genCmd := exec.Command(*gen, genArgs...)
		if genOut, err := genCmd.CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "gen error: %v\n%s\n", err, genOut)
			panic(err)
		}
		totalReqs, _ := countReqs(root)
		docs := countDirs(root)
		fmt.Fprintf(os.Stderr, "    -> %d requirements, %d doc dirs\n", totalReqs, docs)

		cmds := []struct {
			name string
			args []string
		}{
			{"check", []string{"check", root}},
			{"check --json", []string{"check", "--json", root}},
			{"ls", []string{"ls", root}},
			{"stats", []string{"stats", root}},
			{"export html", []string{"export", "html", root, "-o", filepath.Join(*work, fmt.Sprintf("html-%d", s.ReqsPerDoc))}},
		}
		for _, c := range cmds {
			m := measure(*reqmd, c.name, s.Label, s.Copies, s.ReqsPerDoc, totalReqs, docs, c.args)
			results = append(results, m)
			fmt.Fprintf(os.Stderr, "    %-15s %6dms  peak=%5dMB  %8.0f req/s  exit=%d\n",
				c.name, m.WallMs, m.PeakRSSKB/1024, m.ReqsPerSec, m.ExitCode)
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(results); err != nil {
		panic(err)
	}
	if *out != "" {
		f, err := os.Create(*out)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")
		enc.Encode(results)
	}
}

func countReqs(root string) (int, error) {
	count := 0
	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(p, ".md") {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return nil
		}
		count += strings.Count(string(b), "\n## ")
		return nil
	})
	return count, err
}

func countDirs(root string) int {
	count := 0
	filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() {
			count++
		}
		return nil
	})
	return count
}

func measure(reqmdBin, cmdName, scaleLabel string, copies, reqsPerDoc, totalReqs, docs int, args []string) measurement {
	peakRSS.Store(0)
	m := measurement{
		Scale:      scaleLabel,
		Copies:     copies,
		ReqsPerDoc: reqsPerDoc,
		TotalReqs:  totalReqs,
		Docs:       docs,
		Command:    cmdName,
	}
	c := exec.Command(reqmdBin, args...)
	c.Stdout = nil
	c.Stderr = nil
	start := time.Now()
	if err := c.Start(); err != nil {
		m.ExitCode = -1
		return m
	}
	doneCh := make(chan struct{})
	go sampleRSS(c.Process.Pid, doneCh)
	_ = c.Wait()
	elapsed := time.Since(start)
	close(doneCh)
	m.WallMs = elapsed.Milliseconds()
	if m.WallMs > 0 {
		m.ReqsPerSec = float64(totalReqs) / (float64(m.WallMs) / 1000.0)
	}
	m.PeakRSSKB = peakRSS.Load()
	if c.ProcessState != nil {
		m.ExitCode = c.ProcessState.ExitCode()
	}
	return m
}

func sampleRSS(pid int, done <-chan struct{}) {
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-done:
			return
		case <-tick.C:
			kb := readRSSKB(pid)
			if kb > peakRSS.Load() {
				peakRSS.Store(kb)
			}
		}
	}
}

func readRSSKB(pid int) int64 {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "VmRSS:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				var kb int64
				fmt.Sscanf(fields[1], "%d", &kb)
				return kb
			}
		}
	}
	return 0
}
