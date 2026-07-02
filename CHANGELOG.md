# Changelog

## Unreleased

### Editor tooling (P1)
- **LSP completion** — `textDocument/completion` backed by the new core API
  `Completions(text, offset)`: suggestions are the parser's own expectations
  at the cursor (keywords/literals as insertable text, free-form tokens as
  placeholders).
- **LSP hover** — `textDocument/hover` shows the AST node under the cursor
  (rule, action, source span) via `Document.NodeAt`.

### Engine
- **Incremental documents** — `DSL.NewDocument()` keeps parse state across
  edits and re-parses only the statements the edit touched (prefix reused
  as-is, suffix reused with shifted spans). Powers the LSP per keystroke.
- **Attribute grammars** — `NewAttributeGrammar()` with `Inherited`
  (top-down) and `Synthesized` (bottom-up) attribute definitions evaluated
  over the AST in two passes.
- **Immutable builds** — `Build()` now freezes actions too;
  `BuildAllowLateActions()` + lock-protected `CompiledDSL.Action/NodeAction`
  for late binding. Deprecated: `RuleWithPrecedence` (use `Expression()`),
  `TokenWithLookaround` (lookaround never enforced).

### Product & CI (P2)
- **apiflow** — product CLI for the HTTP DSL (`examples/http_dsl/cmd/apiflow`):
  `run`, `check` (syntax-validates via the AST phase, zero side effects),
  `version`.
- **Manual benchmark job** — `workflow_dispatch` job compares HEAD vs main
  with benchstat on the same runner.
- **Release automation** — pushing a `v*` tag runs the full gate and
  publishes a GitHub release with generated notes.

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
