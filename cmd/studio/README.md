# DSL Studio (`cmd/studio`)

Single-binary visual Language Lab for `go-dsl` (`embed.FS` web UI, zero Node/npm dependencies).

## Usage

```bash
go run ./cmd/studio -open
go run ./cmd/studio -dsl path/to/grammar.yaml -port 8420 -open
```

## Capabilities

- **YAML v2 Editor + Visual Grammar Map**: edit tokens, BNF rules, Pratt precedence ladders, layouts, and embedded `tests:` with live validation.
- **Live Sandbox & Token Ribbon**: evaluate scripts under `MaxSteps` budgets, format code (`DSL.Format`), insert parser-driven completions, and inspect token/trivia spans.
- **AST & Execution Trace Inspector**: span-linked syntax trees and step-by-step `EvalTrace` with `[SKIPPED]` lazy branches.
- **Local Ollama & LM Studio Workbench**: auto-discovers models on `127.0.0.1:11434` and `127.0.0.1:1234`, sends `ToGBNF()` constrained decoding grammars, and runs up to 3 closed-loop `RepairPrompt()` self-repair rounds.
- **Test & Fuzz Runner**: runs embedded `tests:` and `FuzzGrammar` iterations plus valid sample synthesis.
- **1-Click Export**: generates Go parser code (`-typed`) and ready-to-install VS Code extension `.zip` archives.
