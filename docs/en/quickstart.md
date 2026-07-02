# go-dsl Quickstart

From zero to a working DSL in a few minutes. This guide covers the modern
engine (v2): two-phase parse/eval, Pratt expressions, validation, and typed
actions.

```bash
go get github.com/arturoeanton/go-dsl/pkg/dslbuilder
```

## 1. The mental model

A go-dsl language has three pieces:

| Piece | What it is | Example |
|---|---|---|
| **Tokens** | Regex terminals (the vocabulary) | `NUMBER = [0-9]+` |
| **Rules** | How tokens combine (the grammar) | `stmt → SET IDENT NUMBER` |
| **Actions** | Go functions that give meaning | `"set" → store the variable` |

Parsing happens in **two phases**:

1. `ParseAST` — tokenize and build a real syntax tree (`*Node`). **No user
   code runs here**, so a rejected alternative can never fire a side effect.
2. `Eval` — run actions exactly once, bottom-up, over the final tree.

`dsl.Parse(code)` does both. You almost always want `Parse`; reach for the
phases separately when you need to inspect or transform the tree.

## 2. Your first DSL

```go
package main

import (
    "fmt"
    "log"
    "strconv"

    "github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

func main() {
    d := dslbuilder.New("orders")

    // Tokens. KeywordToken = reserved word (case-insensitive, word-bounded,
    // wins over free-form tokens). Token = plain regex.
    d.KeywordToken("SELL", "sell")
    d.KeywordToken("OF", "of")
    d.Token("AMOUNT", `[0-9]+\.?[0-9]*`)
    d.Token("STRING", `"(?:[^"\\]|\\.)*"`) // escape-aware string

    // Rules: most specific alternative FIRST (ordered choice, PEG-style).
    d.Rule("cmd", []string{"SELL", "OF", "AMOUNT", "STRING"}, "sellNoted")
    d.Rule("cmd", []string{"SELL", "OF", "AMOUNT"}, "sell")

    // Actions: args arrive in grammar position order.
    d.Action("sell", func(args []interface{}) (interface{}, error) {
        amount, _ := strconv.ParseFloat(args[2].(string), 64)
        return fmt.Sprintf("sold %.2f", amount), nil
    })
    d.Action("sellNoted", func(args []interface{}) (interface{}, error) {
        return fmt.Sprintf("sold %v note %v", args[2], args[3]), nil
    })

    result, err := d.Parse(`sell of 100.50`)
    if err != nil {
        log.Fatal(dslbuilder.GetDetailedError(err)) // line, column, expected, caret
    }
    fmt.Println(result.GetOutput()) // sold 100.50
}
```

## 3. Expressions: always use the Pratt parser

Never encode operator precedence with nested rules — declare it:

```go
d.Token("NUMBER", `\d+`)
d.Token("PLUS", `\+`)
d.Token("STAR", `\*`)
d.Token("POW", `\^`)
d.Token("MINUS", `-`)
d.Token("LPAREN", `\(`)
d.Token("RPAREN", `\)`)

d.Expression("expr").
    Atom("NUMBER", "number").          // action gets [text]
    Group("LPAREN", "expr", "RPAREN"). // parentheses pass through
    Prefix("MINUS", 70, "neg").        // action gets [op, operand]
    InfixLeft("PLUS", 10, "add").      // action gets [left, op, right]
    InfixLeft("STAR", 20, "mul").      // higher power binds tighter
    InfixRight("POW", 30, "pow")       // 2^3^2 = 2^(3^2)

d.Action("number", dslbuilder.Action1(strconv.Atoi))
d.Action("add", dslbuilder.Action3(func(l int, _ string, r int) (int, error) {
    return l + r, nil
}))
// 1 + 2 * 3 = 7 ; (1 + 2) * 3 = 9 ; 10 - 3 - 2 = 5 ; -1 * 2 = -2
```

`Action1/2/3` are typed adapters — no `interface{}` casts. Expression rules
can be referenced from normal rules: `d.Rule("stmt", []string{"PRINT", "expr"}, "print")`.

## 4. Lists: direct left recursion

```go
d.Rule("items", []string{"item"}, "first")
d.Rule("items", []string{"items", "COMMA", "item"}, "append") // left-recursive: fine
```

Indirect left recursion (`a → b …`, `b → a …`) also parses, but `Validate()`
warns about it — restructure or use `Expression()` when you can.

## 5. Control flow: lazy actions

Regular actions receive **already-evaluated** children — both branches of an
`if` would run. Use `NodeAction` for real control flow:

```go
d.NodeAction("ifElse", func(ctx *dslbuilder.EvalContext, n *dslbuilder.Node) (interface{}, error) {
    cond, err := ctx.Eval(n.Child(1))
    if err != nil {
        return nil, err
    }
    if cond.(bool) {
        return ctx.Eval(n.Child(3)) // then-branch ONLY
    }
    return ctx.Eval(n.Child(5))     // else-branch ONLY
})
```

Loops work the same way: re-`Eval` the body node once per iteration.

## 6. Validate, freeze, ship

```go
warnings, err := d.Validate() // static analysis: typos, shadowed alternatives, ...
compiled, err := d.Build()    // Validate + freeze grammar AND actions
result, err := compiled.Parse(input) // safe for concurrent use
```

- A bare `*DSL` is **not** safe for concurrent `Use`/`Parse` — share the
  `CompiledDSL`.
- Need to register actions after building (YAML-loaded grammars)? Use
  `BuildAllowLateActions()` and `compiled.Action(...)`.

## 7. Context: per-call values

```go
d.SetContext("rate", 0.16)                                        // persistent default
result, _ := d.Use(code, map[string]interface{}{"rate": 0.21})    // this call only
// inside an action: d.GetContext("rate")
```

`Use()`'s context never leaks into the DSL — it is scoped to the call.

## 8. Debugging your grammar

```go
node, _ := d.ParseAST(input)
fmt.Println(node.Pretty())      // exactly what matched, as a tree
tokens, _ := d.DebugTokens(input) // the token stream

for _, diag := range d.Diagnostics(script) { // ALL errors, not just the first
    fmt.Println(diag.DetailedError())
}
```

## Where to go next

- [API reference](api-reference.md) — every public API, organized
- [Editor tooling](editor-tooling.md) — LSP, incremental documents, completion, codegen
- [`examples/calculator_pratt`](../../examples/calculator_pratt/) — the whole engine in one file
- [`examples/http_dsl`](../../examples/http_dsl/) — a full scripting DSL (own module)
- [`.claude/skills/go-dsl`](../../.claude/skills/go-dsl/SKILL.md) — drop-in guide for AI coding agents
