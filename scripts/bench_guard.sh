#!/bin/sh
# Benchmark regression guard: runs the benchmark suite and compares it with
# the committed baseline using benchstat. A significant slowdown shows up as
# a positive delta in the report — review before merging.
#
# Requires: go install golang.org/x/perf/cmd/benchstat@latest
# Note: compare on the same machine that produced the baseline
# (see bench_baseline.sh). Numbers across different machines are meaningless.
#
# Usage: ./scripts/bench_guard.sh [count]
set -e

cd "$(dirname "$0")/.."
COUNT="${1:-6}"

if [ ! -f benchmarks/baseline.txt ]; then
    echo "no baseline found — run ./scripts/bench_baseline.sh first" >&2
    exit 1
fi

if ! command -v benchstat >/dev/null 2>&1; then
    echo "benchstat not installed: go install golang.org/x/perf/cmd/benchstat@latest" >&2
    exit 1
fi

NEW="$(mktemp)"
trap 'rm -f "$NEW"' EXIT

go test -run xxx -bench . -benchmem -count "$COUNT" ./pkg/dslbuilder > "$NEW"

echo ""
echo "==> benchstat baseline vs current"
benchstat benchmarks/baseline.txt "$NEW"
