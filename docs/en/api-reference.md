# go-dsl API Reference

Complete reference for `pkg/dslbuilder`, organized by area. Everything here
is verified by the test suite; `go doc` on any symbol gives the full detail.

## Construction

| API | Purpose |
|---|---|
| `New(name) *DSL` | Create a DSL instance |
| `LoadFromYAML(data)` / `LoadFromYAMLFile(path)` | Build tokens+rules from YAML (actions registered separately) |
| `LoadFromJSON(data)` / `LoadFromJSONFile(path)` | Same, JSON |
| `SaveToYAML()` / `SaveToJSON()` (+`...File`) | Export the grammar (rules in declaration order) |

## Tokens

| API | Purpose |
|---|---|
| `Token(name, regex) error` | Free-form token. Anchored internally; empty-matching patterns rejected |
| `KeywordToken(name, word) error` | Reserved word: case-insensitive, word-bounded, priority 90 |
| `Tokens(func(*TokenSet)) *DSL` | Fluent batch: `t.Regex`, `t.Literal` (exact text, no escaping), `t.Keyword` |
| `DebugTokens(code)` | Inspect the token stream |

Matching is deterministic: **priority → longest match → declaration order**.

Deprecated: `TokenWithLookaround` (lookaround stored, never enforced — RE2).

## Rules

| API | Purpose |
|---|---|
| `Rule(name, sequence, action)` | One alternative. Multiple calls with the same name = ordered choice (first match wins — declare specific alternatives first) |
| `RuleWithError(name, seq, action, hint)` | Same + custom error hint shown at the farthest failure |
| `RuleWithRepetition(name, elem, action)` | Kleene star: `name → ε \| name elem` (actions `_empty`/`_append`) |
| `RuleWithPlusRepetition(name, elem, action)` | Kleene plus (actions `_single`/`_append`) |

Left recursion: direct (`list → list item`) and indirect (`a → b…`, `b → a…`)
both parse. The validator warns on indirect (memoization is disabled while
those rules grow).

Deprecated: `RuleWithPrecedence` — the metadata never reordered parses; use
`Expression()`.

## Expressions (Pratt)

```go
d.Expression("expr").
    Atom(token, action).            // terminal            → action([text])
    Group(open, inner, close).      // ( expr )            → passes inner value
    Prefix(token, power, action).   // unary               → action([op, operand])
    InfixLeft(token, power, action).  // left-assoc binary → action([l, op, r])
    InfixRight(token, power, action)  // right-assoc binary
```

Higher `power` binds tighter. Expression rules are referenced from normal
rules by name.

## Actions

| API | Purpose |
|---|---|
| `Action(name, func([]interface{}) (interface{}, error))` | Classic action: evaluated children in grammar order |
| `NodeAction(name, func(*EvalContext, *Node) (interface{}, error))` | **Lazy** action: receives the unevaluated node; evaluate children on demand with `ctx.Eval(n.Child(i))`. Use for control flow |
| `Action1/Action2/Action3(typedFn)` | Generic adapters — no casts; numeric strings convert |
| `Args(raw)` | Positional typed accessors: `.Int(i)`, `.Float(i)`, `.String(i)`, `.Get(i)` |

Rules without an action evaluate to the slice of child values.

## Parsing & evaluation

| API | Purpose |
|---|---|
| `Parse(code) (*Result, error)` | Both phases. `Result.GetOutput()` = value, `Result.GetAST()` = tree |
| `ParseAST(code) (*Node, error)` | Phase 1 only — no actions run |
| `Eval(node) (interface{}, error)` | Phase 2 — actions once, bottom-up. Repeatable |
| `Use(code, ctx) (*Result, error)` | Parse with a call-scoped context (never persists). **Not concurrency-safe on a bare *DSL** |
| `ParseMultiline` / `ParseAuto` / `ParseStatements` | Line-oriented multi-statement helpers |
| `ParseStream(io.Reader, handler) error` | Streaming: statement per line, without loading the input |

`*Node` has `Rule`, `Action`, `Children`, `Token`, `Span` (byte range), plus
`Child(i)`, `Text()`, `Pretty()`, `IsToken()`.

Depth limits: parse and eval recursion are capped at 10 000 — pathological
input errors cleanly instead of overflowing the stack.

## Errors & diagnostics

| API | Purpose |
|---|---|
| `IsParseError(err)` / `GetDetailedError(err)` | Position, expected tokens, rule stack, caret pointer |
| `Diagnostics(code) []*ParseError` | ALL errors in one pass (panic-mode recovery with FIRST-set resync, capped at 50) |
| `RuleWithError(...)` | Adds `hint: ...` to the farthest-failure message |

Error anatomy:

```
no alternative matched for rule condition: expected IDENT or NUMBER, got THEN "then"
hint: expected a comparison like 'status == 200'
rule stack: if_stmt > condition > value at line 1, column 14:
if status == then
             ^
```

## Validation & freezing

| API | Purpose |
|---|---|
| `Validate() ([]string, error)` | Errors: unknown symbols, invalid/empty tokens, no rules. Warnings: unreachable rules, unregistered actions, shadowed alternatives (short prefix declared before long), non-productive cycles, indirect left recursion |
| `Build() (*CompiledDSL, error)` | Validate + freeze **grammar and actions**. The result is safe for concurrent `Parse`/`Use` |
| `BuildAllowLateActions()` | Same but actions stay registrable via `compiled.Action`/`compiled.NodeAction` (lock-protected) |

## Context

| API | Purpose |
|---|---|
| `SetContext(k, v)` / `GetContext(k)` | Persistent values, mutex-protected |
| `Use(code, ctx)` | Call-scoped overlay; discarded after the call |

## Editor & tooling APIs

| API | Purpose |
|---|---|
| `NewDocument() *Document` | Incremental buffer: `Update(text)` re-parses only edited statements; `Diagnostics()`, `Statements()`, `NodeAt(offset)`, `Stats()` |
| `Completions(text, offset) []Completion` | What can come next at the cursor, from the parser's own expectations |
| `NewAttributeGrammar()` | `Inherited` (top-down) + `Synthesized` (bottom-up) attributes over the AST; `Evaluate(root)` |

## CLI tools (cmd/)

| Tool | Purpose |
|---|---|
| `validator -dsl g.yaml` | Core `Validate()` from the command line (`-test`, `-strict`, `-format json`) |
| `ast_viewer -dsl g.yaml -input "..."` | Render the real parse tree |
| `repl -dsl g.yaml` | Interactive exploration |
| `dslgen -dsl g.yaml -o gen.go` | Grammar → checked-in Go code |
| `lsp -dsl g.yaml` | Language server: diagnostics, completion, hover |

## Quality gates (scripts/, .github/)

| Entry | Purpose |
|---|---|
| `./scripts/check.sh` | Full local gate: build+vet+test+race, both modules |
| `./scripts/bench_baseline.sh` / `bench_guard.sh` | Benchmark baseline + benchstat comparison |
| `ci.yml` | Same gate on push/PR; manual `bench` job (HEAD vs main) |
| `release.yml` | Tag `v*` → gate → GitHub release with notes |
