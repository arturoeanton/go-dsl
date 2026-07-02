#!/bin/sh
# Full local quality gate for go-dsl — the same checks CI runs, at zero cost.
# Usage: ./scripts/check.sh
set -e

cd "$(dirname "$0")/.."

echo "==> build (root module)"
go build ./...

echo "==> vet (root module)"
go vet ./...

echo "==> test (root module)"
go test ./...

echo "==> test -race (core)"
go test -race ./pkg/...

echo "==> build + vet + test (examples/http_dsl module)"
(cd examples/http_dsl && go build ./... && go vet ./... && go test ./...)

echo ""
echo "ALL GREEN"
