# Model Context Protocol Server (`cmd/mcp`)

Zero-dependency JSON-RPC 2.0 stdio MCP server exposing any `go-dsl` grammar as deterministic tools.

## Usage

```bash
go install github.com/arturoeanton/go-dsl/cmd/mcp@latest
mcp -dsl grammar.yaml
```

## Tools Exposed

- `dsl_grammar_info` — Markdown specification + valid sample programs (`AIPromptSpec`)
- `dsl_validate_code` — syntax validation + diagnostic self-repair prompt (`RepairPrompt`)
- `dsl_complete` — valid next tokens at byte offset (`Completions`)
- `dsl_format` — canonical formatting (`Format`)
- `dsl_gbnf` — formal GBNF grammar for constrained decoding (`ToGBNF`)
