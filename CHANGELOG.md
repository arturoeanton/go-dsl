# Changelog

## v1.3.0 — the "go-dsl v2" engine

A ground-up rework of the parsing engine. Fully backward compatible with the
existing `Parse`/`Rule`/`Action` API.

### Engine
- **Deterministic tokenizer** — priority → longest match → declaration order;
  anchored patterns (linear lexing); empty-matching tokens rejected.
- **AST-first parsing** — `ParseAST` builds a real tree (`*Node`, spans,
  `Pretty()`); `Eval` runs actions exactly once over the final tree. No action
  can fire on a rejected alternative or during backtracking.
- **Lazy evaluation** — `NodeAction` receives the unevaluated node: real
  control flow where the untaken branch never runs.
- **Pratt expressions** — `Expression()` with binding power and associativity
  for operators.
- **Left recursion** — direct (growing seed) and indirect (generalized growing
  over multi-rule leftmost cycles).
- **Validator** — `Validate()`/`Build()`: unknown symbols, unreachable rules,
  shadowed alternatives (ordered-choice prefix trap), non-productive cycles,
  unregistered actions; `Build()` freezes into a concurrency-safe `CompiledDSL`.
- **Diagnostics & recovery** — farthest-failure errors with expected tokens,
  rule stack, and per-rule hints (`RuleWithError`); `Diagnostics()` reports
  every error in one pass with FIRST-set resynchronization.
- **Hardening** — parse/eval depth limits, race-clean public API, per-call
  context in `Use()`, fuzz targets with regression corpus.

### Tooling
- **Generated parsers** — `cmd/dslgen` turns YAML/JSON grammars into
  checked-in Go code.
- **LSP support** — `cmd/lsp` publishes live diagnostics in any LSP editor.
- **Benchmark guard** — `scripts/bench_guard.sh` + committed baseline
  (benchstat); single-job CI plus `scripts/check.sh` local gate.
- `cmd/validator` rewritten on top of core `Validate()`; `cmd/ast_viewer`
  renders the real parse tree.

### API additions
`ParseAST`, `Eval`, `NodeAction`, `Expression`, `RuleWithError`,
`Diagnostics`, `ParseStream`, `Validate`, `Build`, `Action1/2/3`, `Args`,
`Tokens`/`Expr` fluent layer.

### Removed
- `ImprovedParserV2` (unused, known defect).
