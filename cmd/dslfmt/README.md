# Universal DSL Formatter (`cmd/dslfmt`)

Formats source files written in any `go-dsl` grammar idempotently while preserving `SkipToken` comments.

## Usage

```bash
# Print formatted output to stdout
go run ./cmd/dslfmt -dsl grammar.yaml script.dsl

# Format in place
go run ./cmd/dslfmt -dsl grammar.yaml -w script.dsl

# Check formatting in CI (exits non-zero if unformatted)
go run ./cmd/dslfmt -dsl grammar.yaml -check script.dsl
```
