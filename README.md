# go-dsl

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/doc/install)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/arturoeanton/go-dsl)](https://goreportcard.com/report/github.com/arturoeanton/go-dsl)

**A practical Go toolkit for building small-to-medium DSLs with AST-first parsing, lazy evaluation, validation, diagnostics, and editor tooling.**

go-dsl lets you define tokens, grammar rules, and semantic actions at runtime and get a working language: business rules, query filters, calculators, command syntaxes. The engine parses to a real AST first and evaluates actions afterwards (lazily where you need it), so side effects are safe by construction.

## ✨ Features

### Engine (new)
- 🌳 **Two-Phase Engine**: `Parse()` first builds a real AST (`ParseAST`), then executes actions exactly once (`Eval`). Actions can never run on rejected alternatives or during backtracking — side effects are safe by construction
- 🧾 **Real AST**: `*Node` trees with rule, action, children, token, and source `Span`; pretty-printing via `node.Pretty()`
- 🎯 **Deterministic Tokenizer**: token matching resolves by priority → longest match → declaration order; tokens that can match the empty string are rejected at definition time
- 🎚️ **Pratt Expression Parser**: declare operators with binding power and associativity via `Expression()` — no grammar gymnastics for precedence
- 🧭 **Farthest-Failure Errors**: syntax errors report the farthest failure point, the expected tokens, and the rule stack
- 🧊 **Freezable Builder**: `Build()` validates the grammar once and returns an immutable `CompiledDSL`, safe for concurrent use
- ✅ **Integrated Validator**: `dsl.Validate()` detects unknown symbols, unreachable rules, unregistered actions, non-productive cycles, and indirect left-recursive cycles. Indirect LR is supported, but warned because it disables memoization during growth (see below)
- 🦥 **Lazy Node Actions**: `NodeAction()` receives the unevaluated AST node, enabling real control flow (if/else branches that don't execute unless taken)
- 🧰 **Typed Action Helpers**: `Action1/Action2/Action3` generics and the `Args` accessor remove type-assertion boilerplate
- 🔒 **Per-Parse Context**: `Use(code, ctx)` scopes the context to that call — it no longer mutates the DSL's persistent context; `Set/Get/SetContext/GetContext` are mutex-protected
- 🩺 **Multi-Error Diagnostics**: `Diagnostics(code)` recovers after each failure and reports every syntax/lexical error in one pass (powers the LSP server)
- 💬 **Per-Rule Error Hints**: `RuleWithError(...)` attaches a domain-specific message shown when that alternative is the farthest failure
- 🔁 **Indirect Left Recursion**: multi-rule leftmost cycles (`a → b …`, `b → a …`) parse via a generalized growing algorithm
- 🌊 **Streaming**: `ParseStream(io.Reader, handler)` processes line-oriented scripts without loading them into memory
- 🖨️ **Code Generation**: `cmd/dslgen` turns a YAML/JSON grammar into checked-in Go code
- 🧿 **LSP Server**: `cmd/lsp` gives any LSP editor live diagnostics, **completion** (parser expectations at the cursor), and **hover** (AST node under the cursor)
- ⚡ **Incremental Documents**: `NewDocument()` re-parses only the statements an edit touched, reusing untouched subtrees (this is what the LSP runs per keystroke)
- 🧮 **Attribute Grammars**: `NewAttributeGrammar()` evaluates inherited (top-down) and synthesized (bottom-up) attributes over the AST

### Core
- 🚀 **Dynamic DSL Creation**: Build custom languages at runtime
- 🏗️ **Direct Left Recursion**: `movements -> movements movement` handled with the growing seed algorithm (direct only; use `Expression()` for operators)
- ⚡ **Memoization (Packrat)**: linear-time parsing even with backtracking
- 🎨 **KeywordToken Priority**: keywords win over generic identifiers
- 🔨 **Builder Pattern API**: fluent interface, plus a high-level `Tokens()`/`Expr()` layer
- 📄 **Declarative Syntax**: define DSLs in YAML/JSON files
- 🛠️ **Developer Tools**: AST viewer (real parse trees), grammar validator (core-powered), interactive REPL
- 🔁 **Repetition Rules**: Kleene star (*) and plus (+)
- 📐 **Multiline Support**: `ParseMultiline()`, `ParseAuto()`, `ParseWithBlocks()`
- ✅ **Backward Compatible**: the classic `Parse`/`Rule`/`Action` API is unchanged

## 🚀 Quick Start

### Installation

```bash
go get github.com/arturoeanton/go-dsl/pkg/dslbuilder
```

### Basic Example

```go
package main

import (
    "fmt"
    "log"
    "github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

func main() {
    dsl := dslbuilder.New("HelloDSL")
    dsl.KeywordToken("HELLO", "hello")
    dsl.KeywordToken("WORLD", "world")
    dsl.Rule("greeting", []string{"HELLO", "WORLD"}, "greet")
    dsl.Action("greet", func(args []interface{}) (interface{}, error) {
        return "Hello, World!", nil
    })

    result, err := dsl.Parse("hello world")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(result.GetOutput()) // Output: Hello, World!
}
```

### Calculator with the High-Level API (new)

```go
calc, err := dslbuilder.New("calc").
    Tokens(func(t *dslbuilder.TokenSet) {
        t.Regex("NUMBER", `\d+`)
        t.Literal("PLUS", "+")     // literal text, no regex escaping needed
        t.Literal("MINUS", "-")
        t.Literal("STAR", "*")
        t.Literal("SLASH", "/")
        t.Literal("POW", "^")
        t.Literal("LPAREN", "(")
        t.Literal("RPAREN", ")")
    }).
    Expr("expr", func(e *dslbuilder.ExpressionBuilder) {
        e.Atom("NUMBER", "number")
        e.Group("LPAREN", "expr", "RPAREN")
        e.Prefix("MINUS", 70, "neg")
        e.InfixLeft("PLUS", 10, "add")
        e.InfixLeft("MINUS", 10, "sub")
        e.InfixLeft("STAR", 20, "mul")
        e.InfixLeft("SLASH", 20, "div")
        e.InfixRight("POW", 30, "pow")
    }).
    WithAction("number", dslbuilder.Action1(strconv.Atoi)).
    WithAction("add", dslbuilder.Action3(func(l int, _ string, r int) (int, error) { return l + r, nil })).
    // ... sub, mul, div, neg, pow ...
    Build() // validates and freezes the grammar

result, _ := calc.Parse("1 + 2 * 3")   // 7
result, _  = calc.Parse("(1 + 2) * 3") // 9
result, _  = calc.Parse("10 - 3 - 2")  // 5  (left associative)
result, _  = calc.Parse("2 ^ 3 ^ 2")   // 512 (right associative)
result, _  = calc.Parse("-1 * 2")      // -2 (prefix operator)
```

### Two-Phase Parsing: AST first, actions after (new)

```go
// Phase 1: parse only — no action runs, no side effects possible
node, err := dsl.ParseAST(`GET "https://api.example.com" header "X" "Y"`)
if err != nil { /* rich syntax error */ }

fmt.Println(node.Pretty())
// http_request (httpWithOptions)
// ├─ http_method (methodType)
// │  └─ GET "GET"
// ...

// Phase 2: evaluate the final tree exactly once
result, err := dsl.Eval(node)
```

`dsl.Parse()` does both phases for you — but because they are separate, an
alternative that is tried and rejected during parsing can never fire an HTTP
call, write a file, or mutate your context.

### Lazy Control Flow with NodeAction (new)

```go
// Regular actions receive evaluated children. For control flow you want the
// opposite: evaluate ONLY the branch that is taken.
dsl.NodeAction("ifElse", func(ctx *dslbuilder.EvalContext, n *dslbuilder.Node) (interface{}, error) {
    cond, err := ctx.Eval(n.Child(1))
    if err != nil {
        return nil, err
    }
    if cond.(bool) {
        return ctx.Eval(n.Child(3)) // then-branch only
    }
    return ctx.Eval(n.Child(5))     // else-branch only
})
```

### Better Errors (new)

```go
_, err := dsl.Parse("if status == then")
fmt.Println(dslbuilder.GetDetailedError(err))
// no alternative matched for rule condition: expected IDENT or NUMBER, got THEN "then"
// rule stack: if_stmt > condition > value at line 1, column 14:
// if status == then
//              ^
```

The parser tracks the *farthest* failure (not the last local one) and reports
what it expected there, plus the rule stack that led to it.

### Grammar Validation (new)

```go
warnings, err := dsl.Validate()
// err      -> structural problems: unknown symbols, no rules, invalid tokens
// warnings -> unreachable rules, unregistered actions, non-productive cycles,
//             indirect left recursion (unsupported), ...

compiled, err := dsl.Build() // Validate + freeze; mutations are rejected afterwards
```

## 📚 Examples

Three official demos cover the whole surface — start with these:

| Demo | What it shows |
|---|---|
| [`examples/calculator_pratt`](examples/calculator_pratt/) | Expressions: Pratt precedence/associativity, typed actions, AST inspection, Build() |
| [`examples/http_dsl`](examples/http_dsl/) | A full scripting DSL: lazy control flow, side effects done right, API testing (self-contained module) |
| [`examples/scim`](examples/scim/) | Parsing a real-world standard (SCIM 2.0 filters): query DSL over data |

The rest of [`examples/`](examples/) are additional recipes kept for reference.


### 1. Accounting DSL

```go
accounting := dslbuilder.New("Accounting")

accounting.KeywordToken("VENTA", "venta")
accounting.KeywordToken("DE", "de")
accounting.KeywordToken("CON", "con")
accounting.KeywordToken("IVA", "iva")
accounting.Token("IMPORTE", "[0-9]+\\.?[0-9]*")

// Most specific alternatives first (ordered choice)
accounting.Rule("command", []string{"VENTA", "DE", "IMPORTE", "CON", "IVA"}, "saleWithTax")
accounting.Rule("command", []string{"VENTA", "DE", "IMPORTE"}, "simpleSale")

// Direct left recursion is supported
accounting.Rule("movements", []string{"movement"}, "singleMovement")
accounting.Rule("movements", []string{"movements", "movement"}, "multipleMovements")

accounting.Action("saleWithTax", func(args []interface{}) (interface{}, error) {
    amount, _ := strconv.ParseFloat(args[2].(string), 64)
    tax := amount * 0.16
    return Transaction{Amount: amount, Tax: tax, Total: amount + tax}, nil
})
// "venta de 5000 con iva" → Transaction{Amount: 5000, Tax: 800, Total: 5800}
```

### 2. Context-Scoped Evaluation

```go
// The context passed to Use() lives ONLY for that call: it shadows the
// persistent context and is discarded afterwards.
mexContext := map[string]interface{}{"country": "MX"}
result, _ := accounting.Use(`registrar venta de 5000`, mexContext)

colContext := map[string]interface{}{"country": "COL"}
result, _ = accounting.Use(`crear compra de 3000`, colContext)

// accounting.GetContext("country") is unaffected between calls.
```

### 3. Declarative DSL Definition

```yaml
# calculator.yaml
name: "Calculator"
tokens:
  NUMBER: "[0-9]+"
  PLUS: "\\+"
rules:
  - name: "expr"
    pattern: ["NUMBER", "PLUS", "NUMBER"]
    action: "add"
```

```go
calcDSL, _ := dslbuilder.LoadFromYAMLFile("calculator.yaml")
calcDSL.Action("add", addFunc)
warnings, err := calcDSL.Validate() // same checks as cmd/validator
```

### 4. Typed Actions (new)

```go
// Instead of casting args by hand:
dsl.Action("add", dslbuilder.Action3(func(left int, _ string, right int) (int, error) {
    return left + right, nil
}))

// Or with the Args helper:
dsl.Action("add", func(raw []interface{}) (interface{}, error) {
    args := dslbuilder.Args(raw)
    return args.Int(0) + args.Int(2), nil
})
```

## 📏 Left Recursion: Scope and Guarantees

go-dsl supports **direct** left recursion (`expr -> expr PLUS term`) using the
growing seed algorithm, and **indirect** left recursion (`a -> b ...`,
`b -> a ...`) using a generalized growing algorithm. `Validate()` still flags
indirect cycles as a warning because memoization is disabled while those
rules parse — restructured grammars, or `Expression()` (Pratt parser) for
operators, are usually clearer and faster.

Also note that rule alternatives use **ordered choice** (PEG-like): declare
the most specific alternatives first. `Validate()` warns when a shorter
alternative shadows a longer one with the same prefix (the classic mistake).

## 🎯 Use Cases

- **Configuration Languages**: Create domain-specific config file formats
- **Business Rules**: Build rule engines for complex business logic
- **Query Languages**: Develop custom query interfaces for your data
- **Calculators**: Build specialized calculation engines
- **Scripting**: Embed custom scripting capabilities in applications
- **Data Processing**: Create transformation languages for ETL pipelines

## 🏗️ Architecture

go-dsl consists of several key components:

- **Tokenizer**: Deterministic lexer (priority → length → declaration order)
- **AST Parser**: Packrat parser that builds `*Node` trees; direct left recursion; Pratt parsing for expression rules
- **Evaluator**: Executes actions exactly once over the final AST; supports lazy `NodeAction`s
- **Validator**: Static analysis of the grammar (`Validate()` / `Build()`)
- **Context System**: Persistent context + per-call scoped context (`Use`)
- **Builder API**: Fluent interface, high-level `Tokens()`/`Expr()` layer
- **Declarative Loader**: YAML/JSON configuration support

## 🛠️ Command-Line Tools

### AST Viewer
Visualize the real parse tree of your DSL input:

```bash
go install github.com/arturoeanton/go-dsl/cmd/ast_viewer@latest
ast_viewer -dsl calculator.yaml -input "10 + 20 * 30" -format tree
```

### Grammar Validator
Runs the same checks as `dsl.Validate()`:

```bash
go install github.com/arturoeanton/go-dsl/cmd/validator@latest
validator -dsl mydsl.yaml -verbose -info
validator -dsl mydsl.yaml -test "some input"   # parses without needing actions
```

### Interactive REPL

```bash
go install github.com/arturoeanton/go-dsl/cmd/repl@latest
repl -dsl calculator.yaml -context data.json
```

### Code Generator
Turn a declarative grammar into checked-in Go code (constructor + action stubs):

```bash
go install github.com/arturoeanton/go-dsl/cmd/dslgen@latest
dslgen -dsl grammar.yaml -package mydsl -func NewMyDSL -o mydsl_gen.go
```

### LSP Server
Live diagnostics, completion, and hover for your DSL in any LSP-capable
editor. Documents re-parse incrementally (only edited statements), completion
comes from the parser's own expectations at the cursor, and hover shows the
AST node (rule, action, source span) under it:

```bash
go install github.com/arturoeanton/go-dsl/cmd/lsp@latest
lsp -dsl grammar.yaml    # wire as a stdio language server
```

The same capabilities are available as library APIs: `dsl.NewDocument()`
(incremental parse + `NodeAt`), `dsl.Completions(text, offset)`, and
`dslbuilder.NewAttributeGrammar()` for semantic analysis over the AST.

See the [cmd/](cmd/) directory for detailed documentation of each tool.

## 🧪 Testing & CI

```bash
./scripts/check.sh              # the full quality gate, locally and free
go test ./...                   # main module
go test -race ./pkg/...         # the core is race-clean
cd examples/http_dsl && go test ./...  # HTTP DSL module

# Performance regression guard (same machine as the baseline):
./scripts/bench_baseline.sh     # (re)generate benchmarks/baseline.txt
./scripts/bench_guard.sh        # compare current benchmarks via benchstat
```

CI runs the same gate in a single GitHub Actions job (`.github/workflows/ci.yml`),
free for public repositories. Fuzz targets (`go test -fuzz FuzzExpressionParse ./pkg/dslbuilder`)
and their regression corpus live in the test suite.

## 📖 Documentation

### English
- [Quickstart](docs/en/quickstart.md) — zero to a working DSL on the v2 engine
- [API Reference](docs/en/api-reference.md) — every public API, organized by area
- [Editor Tooling](docs/en/editor-tooling.md) — LSP, incremental documents, completion, attribute grammars, codegen
- [GoDoc](pkg/dslbuilder/)
- [AI/agent skill for building DSLs with go-dsl](.claude/skills/go-dsl/SKILL.md) — drop-in guide so coding agents (Claude Code, etc.) use the library correctly
- [Examples](examples/)
- [Command-Line Tools](cmd/)

### Español
- [Guía Rápida](docs/es/guia_rapida.md) - Introducción completa y ejemplos
- [Conceptos Avanzados de DSL](docs/es/introduccion_dsl_segunda_parte.md) - Teoría de gramáticas y conceptos avanzados
- [Limitaciones](docs/es/limitaciones.md) - Limitaciones conocidas y soluciones
- [Propuesta de Mejoras](docs/es/propuesta_de_mejoras.md) - Roadmap y mejoras implementadas

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📜 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🔗 Related Projects

- [r2lang](https://github.com/arturoeanton/go-r2lang) - The inspiration for context functionality

## 👨‍💻 Author

**Arturo Elias Anton**
- GitHub: [@arturoeanton](https://github.com/arturoeanton)

---

⭐ If you find this project useful, please give it a star on GitHub!
