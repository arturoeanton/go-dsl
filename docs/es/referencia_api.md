# Referencia de API de go-dsl

Referencia completa de `pkg/dslbuilder`, organizada por área. Todo lo que
figura acá está verificado por la suite de tests; `go doc` sobre cualquier
símbolo da el detalle completo.

## Construcción

| API | Propósito |
|---|---|
| `New(name) *DSL` | Crear una instancia de DSL |
| `LoadFromYAML(data)` / `LoadFromYAMLFile(path)` | Tokens+reglas desde YAML (las acciones se registran aparte) |
| `LoadFromJSON(data)` / `LoadFromJSONFile(path)` | Ídem, JSON |
| `SaveToYAML()` / `SaveToJSON()` (+`...File`) | Exportar la gramática (reglas en orden de declaración) |

## Tokens

| API | Propósito |
|---|---|
| `Token(name, regex) error` | Token libre. Anclado internamente; patrones que matchean vacío se rechazan |
| `KeywordToken(name, word) error` | Palabra reservada: case-insensitive, word-bounded, prioridad 90 |
| `Tokens(func(*TokenSet)) *DSL` | Lote fluido: `t.Regex`, `t.Literal` (texto exacto, sin escapar), `t.Keyword` |
| `DebugTokens(code)` | Inspeccionar el stream de tokens |

El matching es determinista: **prioridad → match más largo → orden de declaración**.

Deprecado: `TokenWithLookaround` (el lookaround se almacena pero nunca se aplica — RE2).

## Reglas

| API | Propósito |
|---|---|
| `Rule(name, secuencia, action)` | Una alternativa. Varias llamadas con el mismo nombre = elección ordenada (gana el primer match — declarar lo específico primero) |
| `RuleWithError(name, seq, action, hint)` | Ídem + mensaje personalizado en el farthest failure |
| `RuleWithRepetition(name, elem, action)` | Kleene star: `name → ε \| name elem` (acciones `_empty`/`_append`) |
| `RuleWithPlusRepetition(name, elem, action)` | Kleene plus (acciones `_single`/`_append`) |

Recursión izquierda: la directa (`list → list item`) y la indirecta
(`a → b…`, `b → a…`) parsean. El validador warnea la indirecta (la
memoización se deshabilita mientras esas reglas crecen).

Deprecado: `RuleWithPrecedence` — la metadata nunca reordenó parses; usá `Expression()`.

## Expresiones (Pratt)

```go
d.Expression("expr").
    Atom(token, action).             // terminal              → action([texto])
    Group(open, inner, close).       // ( expr )              → pasa el valor interno
    Prefix(token, power, action).    // unario                → action([op, operando])
    InfixLeft(token, power, action).   // binario asoc. izq.  → action([izq, op, der])
    InfixRight(token, power, action)   // binario asoc. der.
```

Mayor `power` liga más fuerte. Las reglas de expresión se referencian desde
reglas normales por nombre.

## Acciones

| API | Propósito |
|---|---|
| `Action(name, func([]interface{}) (interface{}, error))` | Acción clásica: hijos ya evaluados, en orden de gramática |
| `NodeAction(name, func(*EvalContext, *Node) (interface{}, error))` | Acción **lazy**: recibe el nodo sin evaluar; evaluás hijos a demanda con `ctx.Eval(n.Child(i))`. Para control de flujo |
| `Action1/Action2/Action3(fnTipada)` | Adaptadores genéricos — sin casts; strings numéricos convierten |
| `Args(raw)` | Accesores tipados posicionales: `.Int(i)`, `.Float(i)`, `.String(i)`, `.Get(i)` |

Las reglas sin acción evalúan al slice de valores de sus hijos.

## Parsing y evaluación

| API | Propósito |
|---|---|
| `Parse(code) (*Result, error)` | Ambas fases. `Result.GetOutput()` = valor, `Result.GetAST()` = árbol |
| `ParseAST(code) (*Node, error)` | Solo fase 1 — no corre ninguna acción |
| `Eval(node) (interface{}, error)` | Fase 2 — acciones una sola vez, bottom-up. Repetible |
| `Use(code, ctx) (*Result, error)` | Parse con contexto scoped a la llamada (nunca persiste). **No es concurrency-safe sobre un *DSL pelado** |
| `ParseMultiline` / `ParseAuto` / `ParseStatements` | Helpers multi-statement orientados a líneas |
| `ParseStream(io.Reader, handler) error` | Streaming: statement por línea, sin cargar el input |

`*Node` tiene `Rule`, `Action`, `Children`, `Token`, `Span` (rango de bytes),
más `Child(i)`, `Text()`, `Pretty()`, `IsToken()`.

Límites de profundidad: la recursión de parseo y evaluación está acotada a
10.000 — input patológico devuelve error limpio en vez de stack overflow.

## Errores y diagnósticos

| API | Propósito |
|---|---|
| `IsParseError(err)` / `GetDetailedError(err)` | Posición, tokens esperados, rule stack, puntero |
| `Diagnostics(code) []*ParseError` | TODOS los errores en una pasada (recovery panic-mode con resync por FIRST, tope 50) |
| `RuleWithError(...)` | Agrega `hint: ...` al mensaje del farthest failure |

Anatomía de un error:

```
no alternative matched for rule condition: expected IDENT or NUMBER, got THEN "then"
hint: expected a comparison like 'status == 200'
rule stack: if_stmt > condition > value at line 1, column 14:
if status == then
             ^
```

## Validación y freeze

| API | Propósito |
|---|---|
| `Validate() ([]string, error)` | Errores: símbolos desconocidos, tokens inválidos/vacíos, sin reglas. Warnings: reglas inalcanzables, acciones sin registrar, alternativas ensombrecidas (prefijo corto antes que largo), ciclos no productivos, recursión izquierda indirecta |
| `Build() (*CompiledDSL, error)` | Validate + freeze de **gramática y acciones**. El resultado es seguro para `Parse`/`Use` concurrente |
| `BuildAllowLateActions()` | Ídem pero las acciones siguen registrables vía `compiled.Action`/`compiled.NodeAction` (con lock) |

## Contexto

| API | Propósito |
|---|---|
| `SetContext(k, v)` / `GetContext(k)` | Valores persistentes, protegidos con mutex |
| `Use(code, ctx)` | Overlay scoped a la llamada; se descarta al terminar |

## APIs de editor y tooling

| API | Propósito |
|---|---|
| `NewDocument() *Document` | Buffer incremental: `Update(text)` re-parsea solo los statements editados; `Diagnostics()`, `Statements()`, `NodeAt(offset)`, `Stats()` |
| `Completions(text, offset) []Completion` | Qué puede venir después del cursor, según las expectativas del propio parser |
| `NewAttributeGrammar()` | Atributos `Inherited` (top-down) + `Synthesized` (bottom-up) sobre el AST; `Evaluate(root)` |

## Herramientas CLI (cmd/)

| Herramienta | Propósito |
|---|---|
| `validator -dsl g.yaml` | El `Validate()` del core desde la terminal (`-test`, `-strict`, `-format json`) |
| `ast_viewer -dsl g.yaml -input "..."` | Renderiza el árbol de parseo real |
| `repl -dsl g.yaml` | Exploración interactiva |
| `dslgen -dsl g.yaml -o gen.go` | Gramática → código Go versionable |
| `lsp -dsl g.yaml` | Language server: diagnósticos, completado, hover |

## Quality gates (scripts/, .github/)

| Entrada | Propósito |
|---|---|
| `./scripts/check.sh` | Gate local completo: build+vet+test+race, ambos módulos |
| `./scripts/bench_baseline.sh` / `bench_guard.sh` | Baseline de benchmarks + comparación con benchstat |
| `ci.yml` | Mismo gate en push/PR; job manual `bench` (HEAD vs main) |
| `release.yml` | Tag `v*` → gate → GitHub release con notes |
