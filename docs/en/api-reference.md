# go-dsl API Reference

Complete reference for `pkg/dslbuilder`, `pkg/dslgen`, and `pkg/wasm`, organized
by area. Everything here is verified by the test suite; `go doc` on any symbol
gives the full detail.

## Construction & Declarative Schema v2

| API | Purpose |
|---|---|
| `New(name) *DSL` | Create a DSL instance |
| `LoadFromYAML(data)` / `LoadFromYAMLFile(path)` | Load v1/v2 YAML (`token_defs`, `rules`, `expressions`, `layouts`, `tests`, `significant_newlines`, `start_rule`) |
| `LoadFromJSON(data)` / `LoadFromJSONFile(path)` | Same, JSON |
| `SaveToYAML()` / `SaveToJSON()` (+`...File`) | Lossless v2 export (preserves token priorities, Pratt expressions, layouts, start rule, and embedded tests) |
| `SetStartRule(name)` | Explicitly override the grammar's start rule |

## Tokens, Comments & Newlines

| API | Purpose |
|---|---|
| `Token(name, regex) error` | Free-form regex token (priority 0). Anchored internally; empty-matching patterns rejected |
| `KeywordToken(name, word) error` | Reserved word: case-insensitive, smart word boundaries, priority 90 |
| `LiteralToken(name, lit) error` | Exact operator/punctuation (`regexp.QuoteMeta`), priority 50 |
| `SkipToken(name, regex) error` | Ignored trivia/comment token (collected in `TokenizeWithTrivia`, skipped during parsing) |
| `SetSignificantNewlines(enabled)` | Emit `NEWLINE` tokens instead of treating `\n` as whitespace |
| `Tokens(func(*TokenSet)) *DSL` | Fluent batch: `t.Regex`, `t.Literal`, `t.Keyword`, `t.Skip`, `t.LineComment`, `t.BlockComment` |
| `TokenizeWithTrivia(code)` | Return both parsed `tokens` and skipped `trivia` (`[]TokenMatch`) |
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
both parse. The validator warns on indirect (memoization is disabled only for
rules participating in the cycle while they grow).

Deprecated: `RuleWithPrecedence` — the metadata never reordered parses; use
`Expression()`.

## Expressions (Full Pratt Engine)

```go
d.Expression("expr").
    Atom(token, action).                        // terminal              → action([text])
    Group(open, inner, close).                  // ( expr )              → passes inner value
    GroupWithAction(open, inner, close, act).   // ( expr )              → act([open, val, close])
    Prefix(token, power, action).               // unary prefix          → action([op, operand])
    Postfix(token, power, action).              // unary postfix (x!)    → action([operand, op])
    InfixLeft(token, power, action).            // left-assoc binary     → action([l, op, r])
    InfixRight(token, power, action).           // right-assoc binary    → action([l, op, r])
    InfixNonAssoc(token, power, action).        // non-assoc (==, <)     → rejects a == b == c
    Call(open, close, sep, power, action).      // fn(a, b)              → action([callee, []args])
    Index(open, inner, close, power, action).   // arr[i]                → action([target, index])
    Member(dot, field, power, action).          // obj.prop              → action([target, field])
    Ternary(qmark, colon, power, action)        // cond ? then : else    → lazy/eager ternary
```

Higher `power` binds tighter (`Left`/`Right` aliases also available). Expression
rules are referenced from normal rules by name.

## Actions & Execution Sandbox

| API | Purpose |
|---|---|
| `Action(name, func([]interface{}) (interface{}, error))` | Classic action: evaluated children in grammar order |
| `ContextAction(name, func(context.Context, map[string]interface{}, []interface{}) (interface{}, error))` | Cancellation- and scope-aware action receiving `EvalOptions.Context` and isolated `Vars` |
| `NodeAction(name, func(*EvalContext, *Node) (interface{}, error))` | **Lazy** action: receives the unevaluated node; evaluate children on demand with `ctx.Eval(n.Child(i))`. Use for `if/else`, `&&`, `\|\|`, loops |
| `Action1/Action2/Action3(typedFn)` | Generic adapters — no casts; numeric strings convert automatically |
| `Args(raw)` | Positional typed accessors: `.Int(i)`, `.Float(i)`, `.String(i)`, `.Get(i)` |

## Parsing, Sandbox Limits & Execution Trace

| API | Purpose |
|---|---|
| `Parse(code) (*Result, error)` | Both phases. `Result.GetOutput()` = value, `Result.GetAST()` = tree |
| `ParseWithOptions(code, EvalOptions)` | Parse + evaluate under `Context`, `MaxSteps`, `MaxDepth`, and isolated `Vars` |
| `ParseTrace(code, EvalOptions)` | Parse + evaluate and return `*EvalTrace` (step count, per-node duration, outputs, and `[SKIPPED]` lazy branches) |
| `ParseAST(code) (*Node, error)` | Phase 1 only — no actions run |
| `Eval(node)` / `EvalWithOptions(node, opts)` | Phase 2 — actions once, bottom-up. Repeatable |
| `Use(code, ctx) (*Result, error)` | Parse with a call-scoped context (never persists) |
| `ParseMultiline` / `ParseAuto` / `ParseStatements` / `ParseWithBlocks` | Multi-statement and block-structured helpers |
| `ParseStream(io.Reader, handler) error` | Streaming: statement per line, without loading the full input |

## Universal Formatter

| API | Purpose |
|---|---|
| `Format(code) (string, error)` | Parse `code` + trivia and emit canonical, idempotent formatted source preserving comments |
| `SetRuleLayout(rule, RuleLayout)` | Customize indentation, newline breaks, and spacing around tokens per grammar rule |

## Declarative Testing, Synthesis & Fuzzing

| API | Purpose |
|---|---|
| `RunConfigTests() (*DSLTestReport, error)` | Execute `tests:` cases embedded in a YAML/JSON grammar |
| `RunTestSuite([]DSLTestCase)` | Run programmatic positive/negative parse and `expect_output` assertions |
| `GenerateSamples(seed, count)` | Synthesize valid sample programs directly from the grammar |
| `FuzzGrammar(seed, iterations)` | Fuzz synthesized + mutated inputs to verify zero panics and bounded execution |

## GBNF Constrained Decoding & Repair Prompts

| API | Purpose |
|---|---|
| `ToGBNF() string` | Export a formal GBNF grammar for constrained decoding in Ollama, LM Studio, and `llama.cpp` |
| `AIPromptSpec() string` | Compact Markdown language specification (tokens, rules, Pratt precedences, valid samples, GBNF) |
| `RepairPrompt(code) (string, bool)` | Build a self-healing diagnostic prompt with line/column caret context and valid next parser tokens (`Completions`) |

## Errors, Validation & Freezing

| API | Purpose |
|---|---|
| `IsParseError(err)` / `GetDetailedError(err)` | Position, expected tokens, rule stack, caret pointer |
| `Diagnostics(code) []*ParseError` | ALL errors in one pass (panic-mode recovery with FIRST-set resync, capped at 50) |
| `Validate() ([]string, error)` | Static checks: unknown symbols, empty tokens, unreachable rules, unregistered actions, shadowed alternatives, non-productive cycles |
| `Build() (*CompiledDSL, error)` | Validate + freeze grammar and actions. Concurrency- and reentrancy-safe |
| `BuildAllowLateActions()` | Same, keeping `compiled.Action`/`NodeAction`/`ContextAction` open |

## Editor, Codegen & WASM Packages

| API / Package | Purpose |
|---|---|
| `NewDocument() *Document` | Incremental buffer: `Update(text)` re-parses only edited statements; `Diagnostics()`, `Statements()`, `NodeAt(offset)`, `Stats()` |
| `Completions(text, offset) []Completion` | Valid next tokens at byte `offset` from parser expectations |
| `NewAttributeGrammar()` | `Inherited` (top-down) + `Synthesized` (bottom-up) attributes over the AST |
| `pkg/dslgen` | Programmatic Go parser generator (`GenerateWithOptions`, `-typed`) + VS Code extension builder (`GenerateVSCodeExtensionZip`) |
| `pkg/wasm` | Browser/Studio JSON bridge (`Engine`) for WASM (`GOOS=js GOARCH=wasm`) and embedded tools |

## CLI Tools (`cmd/`)

| Tool | Purpose |
|---|---|
| `studio [-dsl g.yaml] [-open]` | Visual Language Lab IDE + local Ollama & LM Studio workbench (`http://127.0.0.1:8420`) |
| `validator -dsl g.yaml` | Core `Validate()` from the command line (`-test`, `-strict`, `-format json`) |
| `dsltest -dsl g.yaml` | Run embedded `tests:`, synthesize samples (`-samples`), and fuzz (`-fuzz`) |
| `dslfmt -dsl g.yaml` | Canonical source formatter (`-w`, `-check`) |
| `dslgen -dsl g.yaml` | Generate Go parser code (`-typed`) or a complete VS Code extension (`-vscode <dir>`) |
| `lsp -dsl g.yaml` | Language server: UTF-16 diagnostics, completion, hover, formatting, semantic tokens |
| `mcp -dsl g.yaml` | Stdio Model Context Protocol server (`dsl_grammar_info`, `dsl_validate_code`, `dsl_complete`, `dsl_format`, `dsl_gbnf`) |
| `ast_viewer -dsl g.yaml -input "..."` | Render the real parse tree (`tree`, `json`, `compact`) |
| `repl -dsl g.yaml` | Interactive exploration (`-exec` for non-interactive runs) |
| `cmd/wasm` | WebAssembly entrypoint (`GOOS=js GOARCH=wasm go build -o go-dsl.wasm ./cmd/wasm`) |

## Quality Gates (`scripts/`, `.github/`)

| Entry | Purpose |
|---|---|
| `./scripts/check.sh` | Full local gate: root + `js/wasm` build, `vet`, `test`, `-race`, and `examples/http_dsl` |
| `./scripts/bench_baseline.sh` / `bench_guard.sh` | Benchmark baseline + `benchstat` comparison |
| `ci.yml` | Automated gate on push/PR; manual `bench` job (HEAD vs main) |
| `release.yml` | Tag `v*` → full quality gate → GitHub Release |
