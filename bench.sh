#!/usr/bin/env bash
# bench.sh — reproduce the reqmd benchmark from the Hugo /benchmarks/ page.
#
# Builds reqmd, reqmd-bench-gen, and reqmd-bench-run, then measures:
#   1. Baseline: the real bundled spec/ tree (249 requirements, 6 doc dirs).
#   2. Scaling:  synthetic corpora at 71k, 180k, 360k, 720k requirements.
#
# Each command at each scale runs 5 times; the median wall time is reported.
# Results (with CPU/OS/Go metadata) are saved to a timestamped JSON file.
#
# Usage:
#   ./bench.sh                      # full benchmark, default 5 runs
#   ./bench.sh --runs 3             # fewer runs (faster, less stable)
#   ./bench.sh --baseline-only      # just the real spec, no scaling
#   ./bench.sh --scales 2000,5000   # custom scales (reqs/doc targets)
#   ./bench.sh --out results.json   # custom output path
#
# Output: bench-results-YYYYMMDD-HHMMSS.json in the project root (or --out).
set -euo pipefail

cd "$(dirname "$0")"

# --- defaults ---
RUNS=5
SCALES="2000,5000,10000,20000"
OUT=""
EXTRA_ARGS=""
RESULTS_DIR="bench-results"

# --- parse args ---
while [[ $# -gt 0 ]]; do
  case "$1" in
    --runs)         RUNS="$2"; shift 2 ;;
    --scales)       SCALES="$2"; shift 2 ;;
    --out)          OUT="$2"; shift 2 ;;
    --baseline-only) EXTRA_ARGS="$EXTRA_ARGS --skip-scaling"; shift ;;
    --scaling-only)  EXTRA_ARGS="$EXTRA_ARGS --skip-baseline"; shift ;;
    --verbose)       EXTRA_ARGS="$EXTRA_ARGS --verbose"; shift ;;
    -h|--help)
      grep '^#' "$0" | head -20
      exit 0 ;;
    *) echo "unknown flag: $1" >&2; exit 2 ;;
  esac
done

# --- timestamped output path ---
if [[ -z "$OUT" ]]; then
  mkdir -p "$RESULTS_DIR"
  OUT="$RESULTS_DIR/bench-results-$(date +%Y%m%d-%H%M%S).json"
fi

WORK="/tmp/reqmd-bench-$$"
BIN_DIR="/tmp/reqmd-bench-bins-$$"

echo ">>> Building reqmd, reqmd-bench-gen, reqmd-bench-run..." >&2
mkdir -p "$BIN_DIR"
go build -o "$BIN_DIR/reqmd"          ./cmd/reqmd          >&2
go build -o "$BIN_DIR/reqmd-bench-gen" ./cmd/reqmd-bench-gen >&2
go build -o "$BIN_DIR/reqmd-bench-run" ./cmd/reqmd-bench-run >&2

echo ">>> Starting benchmark (runs=$RUNS, scales=$SCALES)" >&2
echo ">>> Output: $OUT" >&2

"$BIN_DIR/reqmd-bench-run" \
  -src spec/ \
  -gen "$BIN_DIR/reqmd-bench-gen" \
  -reqmd "$BIN_DIR/reqmd" \
  -out "$OUT" \
  -work "$WORK" \
  -runs "$RUNS" \
  -scales "$SCALES" \
  $EXTRA_ARGS \
  > /dev/null   # bench-run also prints JSON to stdout; we have it in $OUT

# --- cleanup ---
rm -rf "$WORK" "$BIN_DIR"

echo "" >&2
echo "========================================" >&2
echo " Benchmark complete." >&2
echo " Results saved to: $OUT" >&2
echo "========================================" >&2
echo "" >&2

# --- print a human-readable summary from the JSON ---
if command -v jq &>/dev/null; then
  echo ">>> Summary (from jq):" >&2
  echo "" >&2

  # Hardware
  echo "Hardware:" >&2
  jq -r '.hardware | "  CPU: \(.cpu_model // "unknown")", "  Cores: \(.num_cpu)", "  OS: \(.os)/\(.arch)", "  Go: \(.go_version)", "  Kernel: \(.kernel // "n/a")"' "$OUT" >&2
  echo "" >&2

  # Baseline
  if jq -e '.baseline[0]' "$OUT" >/dev/null 2>&1; then
    echo "Baseline (real spec tree):" >&2
    printf "  %-18s %10s %10s %6s\n" "Command" "Wall(ms)" "Peak(MB)" "Exit" >&2
    jq -r '.baseline[] | "  \(.command)\t\(.wall_ms)\t\(.peak_rss_kb/1024|floor)\t\(.exit_code)"' "$OUT" | \
      awk -F'\t' '{printf "  %-18s %10s %10s %6s\n", $1, $2, $3, $4}' >&2
    echo "" >&2
  fi

  # Scaling
  if jq -e '.scaling[0]' "$OUT" >/dev/null 2>&1; then
    echo "Scaling (synthetic corpora, median of $RUNS runs):" >&2
    printf "  %-14s %-18s %10s %10s %12s %6s\n" "Scale" "Command" "Wall(ms)" "Peak(MB)" "Reqs/sec" "Exit" >&2
    jq -r '.scaling[] | "  \(.scale)\t\(.command)\t\(.wall_ms)\t\(.peak_rss_kb/1024|floor)\t\((.total_reqs/(.wall_ms/1000))|floor)\t\(.exit_code)"' "$OUT" | \
      awk -F'\t' '{printf "  %-14s %-18s %10s %10s %12s %6s\n", $1, $2, $3, $4, $5, $6}' >&2
    echo "" >&2
  fi

  dur=$(jq -r '(.duration_s*10|floor)/10' "$OUT")
  echo "Total benchmark duration: ${dur}s" >&2
else
  echo "(Install jq for a formatted summary. Raw JSON is in $OUT)" >&2
fi