# 🗺️ go-dsl Roadmap

## Current Version: v1.5.0
_Production-Ready Deterministic DSL Platform for Go, WebAssembly & AI Guardrails_

## ✅ Completed Features (v1.0 – v1.5)

### Core Parsing & Evaluation Engine
- [x] Deterministic anchored regex tokenizer (`SkipToken`, `KeywordToken`, `LiteralToken`, `Token`)
- [x] Significant newlines (`SetSignificantNewlines`) and comment/trivia preservation (`TokenizeWithTrivia`)
- [x] Two-phase AST-first Packrat parser (`ParseAST` → `Eval`)
- [x] Direct & indirect left recursion with cycle-scoped Packrat memoization and `maxParseSteps` DoS protection
- [x] Full Pratt expression engine (`Prefix`, `Postfix`, `Left`, `Right`, `InfixNonAssoc`, `Call`, `Index`, `Member`, `Ternary`, `GroupWithAction`)
- [x] Lazy control-flow evaluation (`NodeAction`) & cancellation/sandbox-aware actions (`ContextAction`)
- [x] Execution sandbox (`EvalOptions`: `Context`, `MaxSteps`, `MaxDepth`, isolated `Vars`)
- [x] Step-by-step execution trace & explain mode (`ParseTrace`, `EvalTrace`, `[SKIPPED]` branch tracking)
- [x] Static grammar validator (`Validate`) & reentrant immutable builds (`Build` → `CompiledDSL`)
- [x] Incremental per-statement document re-parsing (`NewDocument`)
- [x] Attribute grammars (`NewAttributeGrammar`: inherited + synthesized passes)
- [x] Streaming (`ParseStream`) and multiline/block helpers

### Tooling, Editor & AI Ecosystem
- [x] Lossless YAML/JSON declarative schema v2 (`token_defs`, `expressions`, `layouts`, `tests`)
- [x] Strongly-typed Go code generator (`cmd/dslgen -typed`)
- [x] 1-click VS Code / Cursor extension generator (`cmd/dslgen -vscode`)
- [x] Full LSP server (`cmd/lsp`): incremental diagnostics, UTF-16 positions, completions, hover, formatting, and semantic tokens
- [x] Universal comment-preserving AST formatter (`DSL.Format`, `cmd/dslfmt`)
- [x] Declarative grammar test runner, sample synthesizer & fuzzer (`cmd/dsltest`)
- [x] AI Guardrails (`ToGBNF`, `AIPromptSpec`, `RepairPrompt`) & stdio MCP server (`cmd/mcp`)
- [x] WebAssembly runtime bridge (`pkg/wasm`, `cmd/wasm`) for browser playgrounds and IDEs
- [x] **DSL Studio (`cmd/studio`)** — Local-First Visual & Organic Language Lab IDE with native **Ollama** & **LM Studio** integration (`ToGBNF()` constrained decoding + closed-loop `RepairPrompt`), Visual Grammar Map, interactive AST/Trace inspector, Fuzzer, and 1-click Go / VS Code `.zip` export

## 🚀 Next Horizons (v1.6+)

- [ ] **ANTLR4 / EBNF Importer** — Bootstrap `DSLConfig` v2 grammars from existing `.g4` and `.ebnf` specifications
- [ ] **Cross-file Module Imports** — Multi-file workspace resolution in `Document` and `cmd/lsp`
- [ ] **Bytecode VM Compilation Target** — Optional stack bytecode compiler over `*Node` for ultra-high-throughput rule engines
