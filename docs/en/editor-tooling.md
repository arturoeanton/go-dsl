# Editor Tooling & Developer Experience

go-dsl ships everything needed to give a custom DSL a first-class authoring
experience: **DSL Studio** (visual IDE), a full **LSP server**, incremental
documents, parser-driven completion, universal formatting, declarative testing
and fuzzing, typed Go code generation, 1-click VS Code extension scaffolding,
MCP server support, and a WebAssembly bridge.

## Visual Language Lab (`cmd/studio`)

For interactive grammar design, AST/trace inspection, fuzzing, and local
**Ollama** / **LM Studio** GBNF workflows, launch DSL Studio:

```bash
go run ./cmd/studio -open
go run ./cmd/studio -dsl grammar.yaml -open
```

See the dedicated [DSL Studio Guide](dsl-studio.md) for full details.

## The LSP Server (`cmd/lsp`)

```bash
go install github.com/arturoeanton/go-dsl/cmd/lsp@latest
lsp -dsl grammar.yaml     # stdio language server
```

Wire it into any LSP-capable editor (VS Code, Zed, Neovim, Helix, ...) as a
stdio server for your DSL's file extension. Capabilities:

| Capability | Behavior |
|---|---|
| **Diagnostics** | UTF-16 accurate multi-error diagnostics on open/change (panic-mode recovery, not just the first failure) |
| **Completion** | Suggestions are the parser's own expectations at the cursor: keywords/literals as insertable text, free-form tokens (`NUMBER`, `STRING`, ...) as snippet placeholders (`${1:NUMBER}`) |
| **Hover** | The AST node under the cursor: token type and text, or rule + action + source snippet |
| **Formatting** | `textDocument/formatting` powered by `DSL.Format` (idempotent, preserves comments) |
| **Semantic Tokens** | `textDocument/semanticTokens/full` classifying comments, keywords, numbers, strings, operators, and identifiers |

Documents re-parse **incrementally**: only the statements touched by an edit
are parsed again (see below), so the server stays fast per keystroke.

Zed example (`settings.json` of your extension/project):

```json
{
  "lsp": {
    "go-dsl": { "binary": { "path": "lsp", "arguments": ["-dsl", "grammar.yaml"] } }
  }
}
```

## 1-Click VS Code Extension & Typed Codegen (`cmd/dslgen`)

Turn a YAML/JSON grammar into checked-in Go code (with optional strongly-typed
action interfaces) or generate a complete VS Code extension directory:

```bash
# Generate Go parser + typed Actions interface
dslgen -dsl grammar.yaml -package mydsl -func NewMyDSL -typed -o mydsl_gen.go

# Generate a ready-to-launch VS Code extension (TextMate + LSP client)
dslgen -dsl grammar.yaml -vscode ./vscode-mydsl
```

## Universal Formatter (`cmd/dslfmt`) & Declarative Tests (`cmd/dsltest`)

```bash
# Format DSL files in place or verify formatting in CI
dslfmt -dsl grammar.yaml -w script.dsl
dslfmt -dsl grammar.yaml -check script.dsl

# Run embedded YAML tests + synthesize valid samples + fuzz the grammar
dsltest -dsl grammar.yaml -fuzz 250 -samples 5
```

## MCP Server (`cmd/mcp`) & WebAssembly Bridge (`cmd/wasm`)

```bash
# Expose validation, completions, formatting, and GBNF over stdio MCP
mcp -dsl grammar.yaml

# Compile the browser/playground WebAssembly binary
GOOS=js GOARCH=wasm go build -o go-dsl.wasm ./cmd/wasm
```

## The Library APIs Behind It

Everything the CLI tools run is a public Go API you can embed directly:

### Incremental Documents

```go
doc := dsl.NewDocument()
doc.Update(text)                  // full parse the first time
doc.Update(textAfterEdit)         // ONLY edited statements re-parse
doc.Stats()                       // {Reused: 4, Reparsed: 1}
doc.Diagnostics()                 // all current errors
doc.Statements()                  // parse trees, in order
node := doc.NodeAt(byteOffset)    // innermost node at a position
```

**Granularity**: incrementality is per *top-level parse unit* (one match of the
grammar's start rule — typically a statement). The unchanged prefix keeps its
trees as-is, the unchanged suffix keeps its trees with spans shifted by the edit
delta, and the edited unit (plus any boundary token that could extend into the
edit) re-parses in full.

### Completion, Formatting & Attribute Grammars

```go
comps := dsl.Completions(text, cursorOffset)
formatted, err := dsl.Format(text)

ag := dslbuilder.NewAttributeGrammar()
ag.Inherited("env", func(parent dslbuilder.Attrs, n *dslbuilder.Node, i int) interface{} {
    return parent["env"]
})
ag.Synthesized("type", "expr", func(n *dslbuilder.Node, children []dslbuilder.Attrs) (interface{}, error) {
    return inferType(n, children)
})
attrs, err := ag.Evaluate(root)
```

## The Product Pattern: `apiflow`

`examples/http_dsl` shows how a DSL graduates into a standalone product:

```bash
cd examples/http_dsl
go run ./cmd/apiflow run script.http       # execute
go run ./cmd/apiflow check script.http     # validate — AST phase only, zero side effects
```
