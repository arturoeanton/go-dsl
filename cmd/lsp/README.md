# Language Server (`cmd/lsp`)

Stdio Language Server Protocol (LSP) implementation for any `go-dsl` YAML/JSON grammar.

## Usage

```bash
go install github.com/arturoeanton/go-dsl/cmd/lsp@latest
lsp -dsl grammar.yaml
```

## Supported LSP Methods

- `textDocument/didOpen`, `didChange`, `didClose` → UTF-16 multi-error diagnostics via incremental `Document`
- `textDocument/completion` → parser-driven completions with snippet placeholders
- `textDocument/hover` → AST node inspection at cursor
- `textDocument/formatting` → idempotent source formatting (`DSL.Format`)
- `textDocument/semanticTokens/full` → token classification (comments, keywords, numbers, strings, operators, identifiers)
