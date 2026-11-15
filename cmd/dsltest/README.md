# Declarative Test & Fuzz Runner (`cmd/dsltest`)

Executes declarative `tests:` embedded in YAML/JSON grammars, synthesizes valid sample inputs, and fuzzes grammars for panics or hangs.

## Usage

```bash
go run ./cmd/dsltest -dsl grammar.yaml
go run ./cmd/dsltest -dsl grammar.yaml -fuzz 250 -samples 5 -seed 42
go run ./cmd/dsltest -dsl grammar.yaml -format json
```
