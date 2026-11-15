# Referencia de API de go-dsl

Referencia completa de `pkg/dslbuilder`, `pkg/dslgen` y `pkg/wasm`, organizada
por área. Todo lo documentado acá está verificado por la suite de tests; `go doc`
sobre cualquier símbolo muestra la firma completa.

## Construcción y Esquema Declarativo v2

| API | Propósito |
|---|---|
| `New(name) *DSL` | Crear una instancia de DSL |
| `LoadFromYAML(data)` / `LoadFromYAMLFile(path)` | Cargar YAML v1/v2 (`token_defs`, `rules`, `expressions`, `layouts`, `tests`, `significant_newlines`, `start_rule`) |
| `LoadFromJSON(data)` / `LoadFromJSONFile(path)` | Ídem en JSON |
| `SaveToYAML()` / `SaveToJSON()` (+`...File`) | Exportación v2 sin pérdidas (preserva prioridades, expresiones Pratt, layouts, `start_rule` y `tests`) |
| `SetStartRule(name)` | Definir explícitamente la regla inicial de la gramática |

## Tokens, Comentarios y Saltos de Línea

| API | Propósito |
|---|---|
| `Token(name, regex) error` | Token regex libre (prioridad 0). Anclado internamente; patrones vacíos se rechazan |
| `KeywordToken(name, word) error` | Palabra reservada: case-insensitive, bordes inteligentes, prioridad 90 |
| `LiteralToken(name, lit) error` | Operador o puntuación literal (`regexp.QuoteMeta`), prioridad 50 |
| `SkipToken(name, regex) error` | Token de comentario/trivia ignorado en el parseo pero preservado en `TokenizeWithTrivia` y `Format` |
| `SetSignificantNewlines(enabled)` | Emitir tokens `NEWLINE` en lugar de ignorar `\n` como espacio |
| `Tokens(func(*TokenSet)) *DSL` | API fluida: `t.Regex`, `t.Literal`, `t.Keyword`, `t.Skip`, `t.LineComment`, `t.BlockComment` |
| `TokenizeWithTrivia(code)` | Devuelve los `tokens` activos y la `trivia` ignorada (`[]TokenMatch`) |
| `DebugTokens(code)` | Inspeccionar el stream de tokens |

El matching es determinista: **prioridad → match más largo → orden de declaración**.

Deprecado: `TokenWithLookaround` (el lookaround se almacena pero nunca se aplica en RE2).

## Reglas

| API | Propósito |
|---|---|
| `Rule(name, secuencia, action)` | Una alternativa. Múltiples llamadas con el mismo nombre = elección ordenada (declarar primero lo más específico) |
| `RuleWithError(name, seq, action, hint)` | Ídem + pista de error personalizada en el farthest failure |
| `RuleWithRepetition(name, elem, action)` | Kleene star: `name → ε \| name elem` (acciones `_empty`/`_append`) |
| `RuleWithPlusRepetition(name, elem, action)` | Kleene plus (acciones `_single`/`_append`) |

Recursión izquierda: la directa (`list → list item`) y la indirecta
(`a → b…`, `b → a…`) están soportadas. El validador emite warning en la
indirecta (la memoización se desactiva únicamente en las reglas del ciclo durante su expansión).

Deprecado: `RuleWithPrecedence` — usá `Expression()`.

## Expresiones (Motor Pratt Completo)

```go
d.Expression("expr").
    Atom(token, action).                        // terminal              → action([texto])
    Group(open, inner, close).                  // ( expr )              → pasa el valor interno
    GroupWithAction(open, inner, close, act).   // ( expr )              → act([open, val, close])
    Prefix(token, power, action).               // unario prefijo        → action([op, operando])
    Postfix(token, power, action).              // unario postfijo (x!)  → action([operando, op])
    InfixLeft(token, power, action).            // binario asoc. izq.    → action([izq, op, der])
    InfixRight(token, power, action).           // binario asoc. der.    → action([izq, op, der])
    InfixNonAssoc(token, power, action).        // no asociativo (==, <) → rechaza a == b == c
    Call(open, close, sep, power, action).      // fn(a, b)              → action([callee, []args])
    Index(open, inner, close, power, action).   // arr[i]                → action([target, index])
    Member(dot, field, power, action).          // obj.prop              → action([target, field])
    Ternary(qmark, colon, power, action)        // cond ? then : else    → ternario
```

Mayor `power` liga más fuerte (también disponibles los alias `Left`/`Right`).
Las reglas de expresión se referencian por nombre desde cualquier regla normal.

## Acciones y Sandbox de Ejecución

| API | Propósito |
|---|---|
| `Action(name, func([]interface{}) (interface{}, error))` | Acción clásica: hijos ya evaluados en orden gramatical |
| `ContextAction(name, func(context.Context, map[string]interface{}, []interface{}) (interface{}, error))` | Acción consciente de cancelación/timeout (`EvalOptions.Context`) y variables aisladas (`Vars`) |
| `NodeAction(name, func(*EvalContext, *Node) (interface{}, error))` | Acción **lazy**: recibe el nodo sin evaluar; evaluás hijos a demanda con `ctx.Eval(n.Child(i))`. Ideal para `if/else`, `&&`, `\|\|`, loops |
| `Action1/Action2/Action3(fnTipada)` | Adaptadores genéricos sin casts manuales |
| `Args(raw)` | Accesores posicionales: `.Int(i)`, `.Float(i)`, `.String(i)`, `.Get(i)` |

## Parsing, Límites de Sandbox y Traza de Evaluación

| API | Propósito |
|---|---|
| `Parse(code) (*Result, error)` | Ambas fases. `Result.GetOutput()` = valor, `Result.GetAST()` = árbol |
| `ParseWithOptions(code, EvalOptions)` | Parseo + evaluación acotada con `Context`, `MaxSteps`, `MaxDepth` y `Vars` aisladas |
| `ParseTrace(code, EvalOptions)` | Parseo + evaluación devolviendo `*EvalTrace` (pasos, tiempos por nodo, salidas y ramas lazy `[SKIPPED]`) |
| `ParseAST(code) (*Node, error)` | Solo fase 1 — ninguna acción se ejecuta |
| `Eval(node)` / `EvalWithOptions(node, opts)` | Fase 2 — ejecuta acciones una sola vez, bottom-up |
| `Use(code, ctx) (*Result, error)` | Parse con contexto local a la llamada |
| `ParseMultiline` / `ParseAuto` / `ParseStatements` / `ParseWithBlocks` | Helpers multi-sentencia y de bloques |
| `ParseStream(io.Reader, handler) error` | Streaming línea a línea sin cargar todo en memoria |

## Formateador Universal

| API | Propósito |
|---|---|
| `Format(code) (string, error)` | Parsea código + trivia y devuelve código formateado canónico e idempotente preservando comentarios |
| `SetRuleLayout(rule, RuleLayout)` | Configura indentación, saltos de línea y espaciado por regla |

## Testing Declarativo, Síntesis y Fuzzing

| API | Propósito |
|---|---|
| `RunConfigTests() (*DSLTestReport, error)` | Ejecuta los casos `tests:` embebidos en el YAML/JSON |
| `RunTestSuite([]DSLTestCase)` | Ejecuta suites declarativas con aserciones de parseo y `expect_output` |
| `GenerateSamples(seed, count)` | Sintetiza programas válidos a partir de la gramática |
| `FuzzGrammar(seed, iterations)` | Fuzzea entradas sintetizadas y mutadas verificando ausencia de panics y cuelgues |

## Decodificación GBNF y Auto-Reparación

| API | Propósito |
|---|---|
| `ToGBNF() string` | Exporta gramática formal GBNF para decodificación restringida en Ollama, LM Studio y `llama.cpp` |
| `AIPromptSpec() string` | Especificación compacta en Markdown (tokens, reglas, precedencias Pratt, ejemplos válidos y GBNF) |
| `RepairPrompt(code) (string, bool)` | Genera un prompt diagnóstico con línea/columna, puntero visual y tokens válidos siguientes (`Completions`) |

## Errores, Validación y Freeze

| API | Propósito |
|---|---|
| `IsParseError(err)` / `GetDetailedError(err)` | Posición, tokens esperados, rule stack y puntero visual |
| `Diagnostics(code) []*ParseError` | Todos los errores en una pasada (recovery panic-mode con resync por FIRST, tope 50) |
| `Validate() ([]string, error)` | Análisis estático: símbolos desconocidos, tokens vacíos, reglas inalcanzables, acciones faltantes, alternativas ensombrecidas, ciclos no productivos |
| `Build() (*CompiledDSL, error)` | Validate + freeze inmutable, seguro para concurrencia y reentrancia |
| `BuildAllowLateActions()` | Igual que `Build()` permitiendo registrar acciones tardías con lock |

## Paquetes de Editor, Codegen y WASM

| API / Paquete | Propósito |
|---|---|
| `NewDocument() *Document` | Buffer incremental: `Update(text)` re-parsea solo las sentencias tocadas; `Diagnostics()`, `Statements()`, `NodeAt(offset)`, `Stats()` |
| `Completions(text, offset) []Completion` | Tokens válidos siguientes en `offset` calculados por el parser |
| `NewAttributeGrammar()` | Atributos `Inherited` (top-down) y `Synthesized` (bottom-up) sobre el AST |
| `pkg/dslgen` | Generador programático de código Go (`GenerateWithOptions`, `-typed`) y empaquetador `.zip` para VS Code (`GenerateVSCodeExtensionZip`) |
| `pkg/wasm` | Bridge JSON (`Engine`) para WebAssembly (`GOOS=js GOARCH=wasm`) y herramientas embebidas |

## Herramientas CLI (`cmd/`)

| Herramienta | Propósito |
|---|---|
| `studio [-dsl g.yaml] [-open]` | IDE visual DSL Studio + integración local con Ollama y LM Studio (`http://127.0.0.1:8420`) |
| `validator -dsl g.yaml` | `Validate()` desde terminal (`-test`, `-strict`, `-format json`) |
| `dsltest -dsl g.yaml` | Ejecuta `tests:` embebidos, sintetiza ejemplos (`-samples`) y fuzzea (`-fuzz`) |
| `dslfmt -dsl g.yaml` | Formateador canónico (`-w`, `-check`) |
| `dslgen -dsl g.yaml` | Genera código Go (`-typed`) o extensión completa para VS Code (`-vscode <dir>`) |
| `lsp -dsl g.yaml` | Language server: diagnósticos UTF-16, completado, hover, formateo y semantic tokens |
| `mcp -dsl g.yaml` | Servidor stdio Model Context Protocol (`dsl_grammar_info`, `dsl_validate_code`, `dsl_complete`, `dsl_format`, `dsl_gbnf`) |
| `ast_viewer -dsl g.yaml -input "..."` | Visualiza el árbol sintáctico real (`tree`, `json`, `compact`) |
| `repl -dsl g.yaml` | Exploración interactiva (`-exec` para ejecución no interactiva) |
| `cmd/wasm` | Binario WebAssembly (`GOOS=js GOARCH=wasm go build -o go-dsl.wasm ./cmd/wasm`) |

## Quality Gates (`scripts/`, `.github/`)

| Entrada | Propósito |
|---|---|
| `./scripts/check.sh` | Gate local completo: build raíz + `js/wasm`, `vet`, `test`, `-race` y `examples/http_dsl` |
| `./scripts/bench_baseline.sh` / `bench_guard.sh` | Baseline de benchmarks + comparación con `benchstat` |
| `ci.yml` | Gate automático en push/PR; job manual `bench` (HEAD vs main) |
| `release.yml` | Tag `v*` → quality gate completo → GitHub Release |
