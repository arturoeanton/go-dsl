# Editor Tooling & Developer Experience

go-dsl ships everything needed to give a custom DSL a first-class editor
experience: a language server, incremental documents, completion from the
grammar itself, multi-error diagnostics, and code generation.

## The LSP server (`cmd/lsp`)

```bash
go install github.com/arturoeanton/go-dsl/cmd/lsp@latest
lsp -dsl grammar.yaml     # stdio language server
```

Wire it into any LSP-capable editor (VS Code, Zed, Neovim, Helix, ...) as a
stdio server for your DSL's file extension. Capabilities:

| Capability | Behavior |
|---|---|
| **Diagnostics** | On open/change, every syntax and lexical error in the document is underlined (multi-error recovery, not just the first failure) |
| **Completion** | Suggestions are the parser's own expectations at the cursor: keywords/literals as insertable text, free-form tokens (NUMBER, STRING, ...) as placeholders with their pattern |
| **Hover** | The AST node under the cursor: token type and text, or rule + action + source snippet |

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

## The library APIs behind it

Everything the LSP does is a public API you can embed in your own tools.

### Incremental documents

```go
doc := dsl.NewDocument()
doc.Update(text)                  // full parse the first time
doc.Update(textAfterEdit)         // ONLY edited statements re-parse
doc.Stats()                       // {Reused: 4, Reparsed: 1} — verify it
doc.Diagnostics()                 // all current errors
doc.Statements()                  // parse trees, in order
node := doc.NodeAt(byteOffset)    // innermost node at a position
```

**Granularity — read this**: incrementality is per *top-level parse unit*
(one match of the grammar's start rule — typically "a statement"). It is
deliberate, bounded reuse, not arbitrary incremental magic: the unchanged
prefix keeps its trees as-is, the unchanged suffix keeps its trees with
spans shifted by the edit delta, and the edited unit is **re-parsed in
full**. Lexing remains a single linear pass (it is cheap); the savings are
in parsing and tree building. A grammar whose entire document is one
start-rule match re-parses that unit on every edit.

### Completion

```go
comps := dsl.Completions(text, cursorOffset)
for _, c := range comps {
    // c.Label     "set" or "NUMBER"
    // c.IsKeyword true → insertable text; false → placeholder
    // c.Detail    "keyword" or "pattern: [0-9]+"
}
```

Mid-statement, candidates come from the farthest-failure expectations at the
cursor — i.e. exactly what the parser would accept next. At a statement
boundary, candidates are the tokens that can start a statement (FIRST set).

### Multi-error diagnostics

```go
for _, diag := range dsl.Diagnostics(script) {
    fmt.Println(diag.DetailedError()) // line, column, expected, caret
}
```

Recovery is panic-mode with FIRST-set resynchronization: after an error the
parser skips to the next token that can begin a statement, avoiding
cascading spurious errors. Output is capped at 50 diagnostics.

### Attribute grammars (semantic analysis)

```go
ag := dslbuilder.NewAttributeGrammar()

// Top-down: propagate an environment/scope.
ag.Inherited("env", func(parent dslbuilder.Attrs, n *dslbuilder.Node, i int) interface{} {
    return parent["env"]
})

// Bottom-up: synthesize a type/value per rule.
ag.Synthesized("type", "expr", func(n *dslbuilder.Node, children []dslbuilder.Attrs) (interface{}, error) {
    // children[i]["type"] already computed
    return inferType(n, children)
})

attrs, err := ag.Evaluate(root)         // map[*Node]Attrs
_ = attrs[root]["type"]
```

One top-down pass (inherited) followed by one bottom-up pass (synthesized) —
the L-attributed subset, which covers symbol tables, scoping, and type
synthesis without hand-rolled tree walkers.

## Code generation (`cmd/dslgen`)

Turn a prototyped YAML/JSON grammar into reviewable Go code:

```bash
dslgen -dsl grammar.yaml -package mydsl -func NewMyDSL -o mydsl_gen.go
```

The generated file rebuilds the DSL programmatically (keywords round-trip
correctly) and lists the actions you still need to register in its doc
comment. Zero runtime file loading.

## The product pattern: apiflow

`examples/http_dsl` shows how a DSL graduates into a product: the engine is
go-dsl, the language lives in a library package, and a small CLI is the
user-facing entry point:

```bash
cd examples/http_dsl
go run ./cmd/apiflow run script.http       # execute
go run ./cmd/apiflow check script.http     # validate — AST phase only, zero side effects
```

`check` can never fire an HTTP request: it uses `ParseAST`, which by design
runs no actions.
