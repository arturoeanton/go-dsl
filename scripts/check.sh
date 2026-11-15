#!/bin/sh
# Full local quality gate for go-dsl — the same checks CI runs, at zero cost.
# Usage: ./scripts/check.sh
set -e

cd "$(dirname "$0")/.."

echo "==> build (root module)"
go build ./...

echo "==> build (js/wasm target)"
GOOS=js GOARCH=wasm go build -o /tmp/godsl_check.wasm ./cmd/wasm
rm -f /tmp/godsl_check.wasm

echo "==> vet (root module)"
go vet ./...

echo "==> test (root module)"
go test ./...

echo "==> test -race (core + cli)"
go test -race ./pkg/... ./cmd/...

echo "==> build + vet + test (examples/http_dsl module)"
(cd examples/http_dsl && go build ./... && go vet ./... && go test ./...)

echo ""
echo "ALL GREEN"
