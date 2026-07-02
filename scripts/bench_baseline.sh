#!/bin/sh
# (Re)generate the benchmark baseline used by bench_guard.sh.
# Run this on the SAME machine you will compare on — benchmark numbers are
# machine-specific. Commit benchmarks/baseline.txt after reviewing it.
#
# Usage: ./scripts/bench_baseline.sh [count]
set -e

cd "$(dirname "$0")/.."
COUNT="${1:-6}"

mkdir -p benchmarks
go test -run xxx -bench . -benchmem -count "$COUNT" ./pkg/dslbuilder | tee benchmarks/baseline.txt
echo ""
echo "baseline written to benchmarks/baseline.txt (count=$COUNT)"
