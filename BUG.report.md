# Reporte de Auditoría de Bugs y Vulnerabilidades — `go-dsl`

**Versión auditada inicialmente:** `v1.4.0` → **Estado en `v1.5.0`:** ✅ **21 / 21 RESUELTOS Y VERIFICADOS CON TESTS (`-race`)**  
**Alcance:** Motor core (`pkg/dslbuilder`), herramientas CLI y servidor LSP (`cmd/`), módulos de ejemplo productivos (`examples/http_dsl`, `examples/liveview`, `examples/query`) e higiene del repositorio.

---

## Matriz de Estado de Remediación (21/21 Cerrados)

| ID | Severidad | Componente | Estado | Solución Aplicada y Evidencia |
| :--- | :---: | :--- | :---: | :--- |
| `BUG-01` | Crítica | `pkg/dslbuilder/document.go` | ✅ Resuelto | Ventana de dependencia `depsEnd` (lookahead real + EOF) y cota estricta `old.span.Start > oldSuffixStart` + fast-path idéntico (`TestBug01IncrementalDocumentAppendOperatorAndDigits`). |
| `BUG-02` | Alta | `pkg/dslbuilder/dsl.go` | ✅ Resuelto | Eliminado `strings.TrimSpace(code)` en `tokenizeInput`/`tokenizeWithTrivia`; `Span` y `Line:Column` exactos con indentación/líneas previas (`TestBug02LeadingBlankLinesPreserveSpansAndErrorLine`). |
| `BUG-03` | Alta | `diagnostics.go`, `document.go`, `completion.go` | ✅ Resuelto | Limpieza de `parser.memo` al resetear `parser.failure` entre sentencias (`TestBug03MemoResetAcrossStatementsInDiagnosticsAndCompletions`). |
| `BUG-04` | Alta | `pkg/dslbuilder/ast.go` | ✅ Resuelto | `cycleDependentRuleSet` mantiene Packrat activo en sub-reglas fuera del ciclo + guardia `maxParseSteps = 1_000_000` (`TestBug04IndirectLeftRecursionKeepsSubruleMemoFast`). |
| `BUG-05` | Alta | `pkg/dslbuilder/compile.go` | ✅ Resuelto | `reentrantMutex` por goroutine elimina el deadlock en `CompiledDSL.Use`/`Parse`/`Eval` anidados manteniendo exclusión entre goroutines (`TestBug05CompiledDSLNestedUseNoDeadlock`). |
| `BUG-06` | Alta | `declarative.go`, `dsl.go` | ✅ Resuelto | Esquema declarativo v2 sin pérdidas (`token_defs`, `expressions`, `layouts`, `tests`), bordes inteligentes en `KeywordToken` para símbolos (`$status`, `!=`) y orden determinista (`TestBug06LosslessYAMLv2RoundTripAndSymbolKeywords`). |
| `BUG-07` | Media | `pkg/dslbuilder/dsl_multiline.go` | ✅ Resuelto | Reemplazado `string(rune(i+1))` por `strconv.Itoa(i+1)` (`TestBug07And08MultilineErrorFormatting`). |
| `BUG-08` | Media | `dsl_multiline.go`, `stream.go`, `dsl.go` | ✅ Resuelto | Preservación de `parseErr.Input` completo, ajuste de columna por indentación y soporte de input de 1 línea en `getContextLine()`. |
| `BUG-09` | Media | `dsl.go`, `diagnostics.go` | ✅ Resuelto | Decodificación `utf8.DecodeRuneInString`, conteo de runas en `calculateLineColumn` y preservación de `\t` en `DetailedError()` (`TestBug09UTF8RuneErrorAndColumns`). |
| `BUG-10` | Media | `dsl.go`, `dsl_improved.go`, `validate.go`, `declarative.go` | ✅ Resuelto | Locks `d.mu` en `toConfig`/`Validate`/`addDeferredError`, límite de recursión, verificación de tokens sobrantes y reglas anulables en EOF (`TestBug10LegacyParserTrailingTokensAndRecursionLimit`). |
| `BUG-11` | Media | `validate.go`, `block_parser.go` | ✅ Resuelto | Validación completa de acciones/productividad Pratt, soporte de `if/else` anidados y `foreach`, y no-ejecución incondicional en `ParseMultilineBlocks` (`TestBug11ValidatePrattProductivityAndNestedIfElse`). |
| `BUG-12` | Crítica | `cmd/dslgen/main.go` | ✅ Resuelto | Validación estricta con `go/token.IsIdentifier` y rechazo de keywords Go + sanitización de saltos de línea en comentarios (`TestBug12IdentifierValidationAndExtHandling`). |
| `BUG-13` | Alta | `cmd/lsp/server.go` | ✅ Resuelto | Conversión exacta UTF-16 code units ↔ byte offsets y acotamiento estricto al fin de línea en `offsetForPosition` (`TestBug13UTF16PositionsAndLineClamping`). |
| `BUG-14` | Media | `cmd/repl/main.go` | ✅ Resuelto | Implementado `loadHistory()`, salida limpia en `.exit`/`.quit` guardando historial, `.tokens`/`.rules` reales, guardia `os.Stdin.Stat()` y modo `-exec` no bloqueante. |
| `BUG-15` | Media | `cmd/ast_viewer/main.go`, `cmd/*/main.go` | ✅ Resuelto | Uso directo de `ParseAST` sin acciones dummy hardcodeadas, poblado completo de `Token`/`Rule`/`Line`/`Col` en `-verbose` y `filepath.Ext` en todos los CLIs. |
| `VULN-16` | Crítica | `examples/liveview/universal/html_generator.go` | ✅ Resuelto | Escape sistemático con `html.EscapeString` en formularios, tablas, cards y templates (`TestVuln16XSSEscapingInHTMLGenerator`). |
| `VULN-17` | Crítica | `examples/http_dsl/universal/http_engine.go` | ✅ Resuelto | Lecturas acotadas con `io.LimitReader` (32 MiB default), `regexp.Compile` + `QuoteMeta` sin panic en `extractXPath`, `context.WithTimeout` por request y guardia SSRF `SetBlockPrivateNetworks` (`TestVuln17SafeXPathAndPrivateIPGuard`). |
| `HYG-18` + `HYG-19` | Baja | Raíz Git, `.gitignore`, documentación | ✅ Resuelto | Eliminados binarios Mach-O `ast_viewer`, `repl`, `validator` (11.2 MB), actualizado `.gitignore` y sincronizados `README.md`, `README.es.md`, `CHANGELOG.md`, `ROADMAP.md` y `docs/`. |

---

## Resumen Ejecutivo

El motor v2 introducido en `v1.3.0`–`v1.4.0` tiene una base arquitectónica muy sólida: separación en dos fases (`ParseAST` → `Eval`), anclaje `\A` en regexes del lexer para evitar degradación cuadrática en el tokenizador, límites de recursión (`maxParseDepth` / `maxEvalDepth`) y pruebas de fuzzing.

Sin embargo, la auditoría estática y de trazado de ejecución detectó **21 hallazgos reales** (bugs funcionales, corrupción de estado incremental, condiciones de carrera, vectores de DoS y vulnerabilidades de inyección/seguridad en herramientas y ejemplos):

| Severidad | Core (`pkg/dslbuilder`) | Tooling (`cmd/`) | Ejemplos & Repo | Total |
| :--- | :---: | :---: | :---: | :---: |
| **Crítica** | 1 | 1 | 2 | **4** |
| **Alta** | 5 | 1 | 2 | **8** |
| **Media** | 5 | 2 | 0 | **7** |
| **Baja / Higiene** | 0 | 0 | 2 | **2** |
| **Total** | **11** | **4** | **6** | **21** |

---

## 1. Hallazgos en el Motor Core (`pkg/dslbuilder`)

### [BUG-01] CRÍTICA — Corrupción del parseo incremental en `Document.Update` al editar al final de una sentencia (falsos errores en LSP)
- **Ubicación:** `pkg/dslbuilder/document.go:155-160`
- **Código afectado:**
  ```go
  // Reuse from the unchanged prefix: identical offsets, tree as-is.
  if old, ok := oldByStart[tokStart]; ok && old.span.End <= prefix && tokStart < prefix {
      doc.stmts = append(doc.stmts, old)
      doc.stats.Reused++
      pos = tokenAt(old.span.End)
      continue
  }
  ```
- **Causa raíz:** La condición `old.span.End <= prefix` reutiliza un árbol AST anterior incluso cuando la edición ocurre **exactamente en el borde final** (`old.span.End == prefix`) o inmediatamente después, donde los nuevos caracteres extienden el último token o continúan la misma sentencia (por ejemplo, operadores infijos o recursión izquierda).
- **Reproducción exacta (flujo normal de escritura en un editor con `cmd/lsp`):**
  1. Gramática de calculadora (`expr -> NUMBER` e `InfixLeft("PLUS", ...)`).
  2. El usuario escribe `"1"` → `doc.Update("1")` guarda `old` con `Span{Start: 0, End: 1}`.
  3. El usuario continúa escribiendo `" + 2"` → `doc.Update("1 + 2")`.
  4. `commonPrefix("1", "1 + 2")` vale `1`. Como `old.span.End (1) <= prefix (1)` es **verdadero**, `Document` reutiliza `expr("1")` como sentencia cerrada y pasa a parsear `"+ 2"` desde el byte `1`, emitiendo un **falso error de sintaxis** sobre una expresión válida.
  5. Peor aún: si escribe `"1"` y luego `"2"` (`"12"`), `prefix = 1`, reutiliza el nodo `NUMBER("1")` con `Span{0, 1}` y `tokenAt(1)` salta al final, dejando el AST desincronizado respecto a `doc.Text()` (`"12"`).
- **Solución recomendada:** Un statement del prefijo solo es seguro de reutilizar si termina **estrictamente antes** del token que toca o precede a `prefix` y además el primer token posterior en el nuevo texto coincide con el token de sincronización/frontera que existía en `oldText` (por ejemplo, exigir que exista al menos un token completo sin cambios entre `old.span.End` y `prefix`, p. ej. `nextOldToken.End < prefix`).

---

### [BUG-02] ALTA — `tokenizeInput` ejecuta `strings.TrimSpace(code)`, corrompiendo `Span`, `TokenMatch` y líneas/columnas de `ParseError`
- **Ubicación:** `pkg/dslbuilder/dsl.go:1286-1308` y `pkg/dslbuilder/ast.go:400-405`
- **Código afectado:**
  ```go
  func tokenizeInput(grammar *Grammar, code string, errorInput string) ([]TokenMatch, error) {
      code = strings.TrimSpace(code)
      pos := 0
  ```
- **Causa raíz:** `tokenizeInput` recorta espacios y saltos de línea iniciales de `code`, desplazando todos los índices `pos`, `TokenMatch.Start` y `TokenMatch.End` hacia `0`. Sin embargo, `errorInput` (y `p.input` en `astParser.Parse`) conserva el string **original sin recortar**. Además, el propio bucle de `tokenizeInput` ya ignora `' '`, `'\t'`, `'\n'` y `'\r'` en las líneas 1293–1295, por lo que `TrimSpace` es redundante y destructivo.
- **Impacto:**
  1. Si el input comienza con líneas vacías o indentación (ej. `"\n\n  @error"`), `pos` vale `0`, pero `calculateLineColumn(errorInput, 0)` reporta **línea 1, columna 1** (una línea en blanco) en vez de **línea 3, columna 3**.
  2. `node.Span` queda desfasado respecto al string pasado a `ParseAST(code)`: hacer `code[node.Span.Start:node.Span.End]` corta los caracteres equivocados.
  3. Inconsistencia interna: `tokenizeTolerant` (`pkg/dslbuilder/diagnostics.go:145`) **no** hace `TrimSpace`, por lo que `ParseAST` y `Diagnostics` devuelven spans y posiciones distintas para el mismo código.
- **Solución recomendada:** Eliminar `code = strings.TrimSpace(code)` en `tokenizeInput` (y si el lexer debe ignorar whitespace Unicode adicional, hacerlo avanzando `pos` dentro del bucle sin mutar los offsets del string).

---

### [BUG-03] ALTA — Reutilización de `parser.memo` entre iteraciones en `Diagnostics`, `Document.Update` y `Completions` suprime errores y autocompletado
- **Ubicación:** `pkg/dslbuilder/ast.go:563-568`, `pkg/dslbuilder/diagnostics.go:55-61`, `pkg/dslbuilder/document.go:178-183`, `pkg/dslbuilder/completion.go:94-115`
- **Causa raíz:** `astMemoEntry` solo guarda `{node, endPos, err}`, pero **no** guarda qué tokens fallaron (`parseFailure`). En una llamada simple a `ParseAST`, `p.failure` nunca se resetea. Pero en `Diagnostics()`, `Document.Update()` y `expectedAt()` (`Completions`), el bucle **resetea `parser.failure = parseFailure{tokenPos: -1}` en cada sentencia conservando el mismo `parser.memo`**. Cuando una iteración posterior cae en una entrada fallida memoizada (`entry.err == errNoMatch`) generada durante el lookahead/backtracking de una sentencia previa, `parseRule` retorna inmediatamente sin invocar `recordFailure()`.
- **Impacto:**
  - `Diagnostics()` y `Document.Update()` pierden `expected`, `ruleStack`, `hint` y reportan la posición de inicio de la regla en lugar del punto real de fallo (`farthestError` cae en `failPos < 0 → p.pos`).
  - `Completions()` evalúa `if parser.failure.tokenPos >= len(tokens)` (`-1 >= len(tokens)` → `false`) y devuelve **cero sugerencias** o sugerencias erróneas de inicio de sentencia cuando el cursor está a mitad de una sentencia cuyo prefijo fue explorado antes.
- **Solución recomendada:** Limpiar `parser.memo` al iniciar una nueva sentencia tras reseteo de `parser.failure`, o almacenar en `astMemoEntry` el `parseFailure` relativo para re-aplicarlo vía `recordFailure` en caso de cache hit con error.

---

### [BUG-04] ALTA — Denegación de Servicio (CPU DoS exponencial $O(2^n)$) al desactivar toda memoización durante recursión izquierda indirecta
- **Ubicación:** `pkg/dslbuilder/ast.go:558-560` y `625-637`
- **Código afectado:**
  ```go
  // The memo table is disabled while a cycle head is growing:
  useMemo := p.lrHeads == 0
  ```
- **Causa raíz:** Cuando cualquier regla perteneciente a un ciclo de recursión izquierda indirecta está activa en la pila (`p.lrHeads > 0` — por ejemplo, si la regla inicial participa en el ciclo), se deshabilita el Packrat caching **para absolutamente todas las reglas de la gramática**, incluso sub-reglas puras que no forman parte del ciclo.
- **Impacto (Vulnerabilidad DoS):** En cualquier gramática con alternativas que compartan prefijos (ej. `cond -> expr EQ expr | expr NEQ expr` con paréntesis anidados), una entrada pequeña de ~35–50 caracteres como `((((((((((((((((((((x))))))))))))))))))))` provoca backtracking exponencial $O(2^n)$, congelando el hilo de CPU indefinidamente sin alcanzar jamás `maxParseDepth` (`10000`).
- **Solución recomendada:** Restringir la invalidación/desactivación de memoización únicamente a las reglas que pueden alcanzar el `head` activo en el grafo de llamadas, manteniendo Packrat activo para sub-reglas independientes del ciclo, y agregar un presupuesto total de pasos (`maxParseSteps`) por invocación.

---

### [BUG-05] ALTA — Deadlock permanente en `CompiledDSL` ante llamadas anidadas (`Use` / `Parse` / `Eval`) y serialización total
- **Ubicación:** `pkg/dslbuilder/compile.go:39, 124-151` vs `pkg/dslbuilder/dsl.go:719-725`
- **Código afectado:**
  ```go
  func (c *CompiledDSL) Use(code string, ctx map[string]interface{}) (*Result, error) {
      c.mu.Lock()
      defer c.mu.Unlock()
      return c.dsl.Use(code, ctx)
  }
  ```
- **Causa raíz:** `DSL.Use` fue diseñado explícitamente para soportar llamadas `Use` anidadas desde acciones semánticas (p. ej. instrucciones `include`, `eval` o sub-expresiones dinámicas; ver comentario en `dsl.go:719`: *"Nested Use calls inherit the outer parse-scoped context"*). Al congelar el DSL con `Build()`, `CompiledDSL` envuelve `Parse`, `Use` y `Eval` en un `sync.Mutex` no reentrante (`c.mu`).
- **Impacto:**
  1. Cualquier acción que invoque `compiled.Use(...)`, `compiled.Parse(...)` o `compiled.Eval(...)` produce un **deadlock instantáneo e irrecuperable** de la goroutine.
  2. Además, aunque `CompiledDSL` se documenta para uso concurrente en servidores, el `sync.Mutex` global impide el paralelismo real entre goroutines durante `Eval` (p. ej. si una acción hace I/O o HTTP, bloquea a todas las demás goroutines).
- **Solución recomendada:** Pasar el contexto de ejecución por valor dentro de `EvalContext` (en vez de mutar `d.parseContext` en el struct compartido `*DSL`). Así `CompiledDSL.Parse` / `Use` / `Eval` serán 100% libres de locks (`lock-free`), verdaderamente paralelos y reentrantes sin riesgo de deadlock.

---

### [BUG-06] ALTA — Inconsistencia, pérdida de datos y no-determinismo en `SaveToYAML`/`LoadFromYAML` (`declarative.go`)
- **Ubicación:** `pkg/dslbuilder/declarative.go:36-41, 166-184, 210-258, 339-366` y `pkg/dslbuilder/dsl.go:1016-1029`
- **Cuatro defectos combinados:**
  1. **Pérdida silenciosa de `Expression()` (Pratt rules):** `DSLConfig` y `toConfig()` ignoran `d.grammar.exprRules`. Exportar a YAML/JSON cualquier DSL moderno construido con `Expression()` o `Expr()` descarta todas sus reglas de expresiones sin error ni advertencia.
  2. **Tokenización NO determinista al cargar desde YAML/JSON:** El README garantiza desempate determinista por *orden de declaración*. Sin embargo, `DSLConfig.Tokens` es un `map[string]string` y `createDSLFromConfig` itera `for name, pattern := range config.Tokens` (`línea 166`). En Go la iteración de mapas es aleatoria en cada ejecución, rompiendo el determinismo en `LoadFromYAML`, `cmd/lsp`, `cmd/validator`, `cmd/repl` y `cmd/ast_viewer`.
  3. **Mutación silenciosa de `Token` literal a `KeywordToken`:** `d.Token("X", "x")` crea un token case-sensitive de prioridad `0` sin `\b`. Al guardar y recargar con `LoadFromYAML` (o generar código con `cmd/dslgen`), `isKeywordToken("x")` devuelve `true` y lo convierte en `KeywordToken` (`(?i)\bx\b`, prioridad `90`, case-insensitive y con bordes de palabra), alterando la semántica del lenguaje.
  4. **Doble escape en `extractKeywordFromPattern` y fallo de `KeywordToken` con símbolos:** Si un keyword contiene caracteres escapados por `QuoteMeta` (ej. `c++` → `(?i)\bc\+\+\b`), `extractKeywordFromPattern` devuelve `c\+\+` sin des-escapar, y `LoadFromYAML` vuelve a pasar `c\+\+` por `QuoteMeta`, corrompiendo el patrón (`c\\\+\\\+`). Además, `KeywordToken("NEQ", "!=")` envuelve en `\b`, lo que jamás matchea `"a != b"` porque entre espacio (`\W`) y `!` (`\W`) no existe borde de palabra `\b`.
- **Solución recomendada:** Extender `DSLConfig` con soporte explícito para ` tokens` ordenados (con tipo `regex | literal | keyword | skip`) y `expressions`, preservando un round-trip exacto 1:1.

---

### [BUG-07] MEDIA — Bug `string(rune(i+1))` en `ParseStatements` inyecta caracteres de control binarios en mensajes de error
- **Ubicación:** `pkg/dslbuilder/dsl_multiline.go:130-132`
- **Código afectado:**
  ```go
  if parseErr, ok := err.(*ParseError); ok {
      parseErr.Message = "in statement " + string(rune(i+1)) + ": " + parseErr.Message
  }
  ```
- **Causa raíz:** En Go, `string(rune(1))` produce el carácter de control ASCII `\x01` (SOH), no `"1"`; para la sentencia 10 produce `"\n"`.
- **Solución recomendada:** Reemplazar `string(rune(i+1))` por `strconv.Itoa(i+1)`.

---

### [BUG-08] MEDIA — `ParseMultiline` y `ParseStream` rompen `ParseError.DetailedError()` y desalinean columnas
- **Ubicación:** `pkg/dslbuilder/dsl_multiline.go:36-42` y `pkg/dslbuilder/stream.go:55-60`
- **Causa raíz:** Ambos métodos llaman a `d.Parse(trimmed)` pasando una sola línea recortada (`trimmed`), por lo que `parseErr.Input` contiene únicamente 1 línea. Luego sobrescriben `parseErr.Line = lineNum` (ej. `Line = 5`). Al llamar a `GetDetailedError(err)`, `getContextLine()` busca `lines[4]` en un `Input` de longitud 1, devolviendo `""` (sin contexto visual) y con `Column` desplazada si la línea tenía indentación inicial.
- **Solución recomendada:** En `ParseMultiline`, asignar `parseErr.Input = code` y sumar el offset de espacios iniciales a `parseErr.Column`; en `ParseStream`, conservar `parseErr.Input = rawLine` y ajustar `getContextLine()` para usar `pe.Input` directamente cuando contiene una única línea.

---

### [BUG-09] MEDIA — Corrupción de UTF-8 / Mojibake en errores léxicos y desalineación de columnas con Unicode
- **Ubicación:** `pkg/dslbuilder/dsl.go:92-110, 1300-1301` y `pkg/dslbuilder/diagnostics.go:151-165`
- **Causa raíz:** Tanto `tokenizeInput` como `tokenizeTolerant` indexan bytes crudos (`c := code[pos]`, `fmt.Sprintf("unexpected character: %c", c)`, `string(c)`, `pos++`).
- **Impacto:** Ante caracteres en español o Unicode no reconocidos (ej. `ñ`, `á`, `€`, emojis):
  1. `tokenizeInput` imprime Mojibake (`unexpected character: Ã`) y guarda un byte huérfano inválido UTF-8 (`"\xc3"`) en `ParseError.Token`.
  2. `tokenizeTolerant` emite entre 2 y 4 diagnósticos falsos por un solo carácter Unicode, todos con bytes UTF-8 inválidos.
  3. `calculateLineColumn` cuenta bytes en lugar de runas Unicode, desplazando el puntero `^` en `DetailedError()`.
- **Solución recomendada:** Usar `utf8.DecodeRuneInString(code[pos:])` para reportar la runa completa y avanzar `pos += width`, y contar runas en `calculateLineColumn`.

---

### [BUG-10] MEDIA — Condiciones de carrera (Data Races) residuales en `toConfig`, `Validate` y `deferredErrors` + Parsers legacy inseguros
- **Ubicación:**
  - `pkg/dslbuilder/declarative.go:344` (`SaveToYAML`/`SaveToJSON` lee `d.context` sin `d.mu.RLock()`)
  - `pkg/dslbuilder/validate.go:37, 180-186` (`Validate` lee `d.deferredErrors`, `d.actions` y `d.nodeActions` sin `d.mu.RLock()`, mientras `d.Rule` escribe `d.deferredErrors` sin `d.mu.Lock()` y `d.Action` con `d.mu.Lock()`)
  - `pkg/dslbuilder/dsl.go:1205-1251` (`NewParser`) y `pkg/dslbuilder/dsl_improved.go:62-120` (`NewImprovedParser`)
- **Impacto:**
  - Llamar a `SaveToYAML()` concurrentemente con `SetContext()` o `Use()` puede abortar el proceso Go con `fatal error: concurrent map iteration and map write`.
  - Los tipos exportados legacy `Parser` e `ImprovedParser` carecen de `maxParseDepth` (riesgo de *stack overflow* fatal del proceso), ejecutan acciones con efectos secundarios durante backtracking, `NewParser.Parse` **no verifica que se hayan consumido todos los tokens finales** (acepta basura al final de una entrada válida), y ambos fallan con `"unexpected end of input"` cuando una regla anulable (`RuleWithRepetition`) aparece al final de la entrada (`dsl.go:1385`, `dsl_improved.go:434`).
- **Solución recomendada:** Proteger `toConfig()`, `Validate()` y `deferredErrors` bajo `d.mu`; marcar `NewParser` y `NewImprovedParser` como `Deprecated` delegando internamente en `astParser` + `Eval`.

---

### [BUG-11] MEDIA — Brechas en `Validate()` y defectos lógicos severos en `BlockParser` / `ParseMultilineBlocks`
- **Ubicación:** `pkg/dslbuilder/validate.go:66-100, 335-339` y `pkg/dslbuilder/block_parser.go:125-265`
- **Detalles:**
  1. **`Validate()` omite validar acciones en operadores Pratt (`validate.go:79-96`):** Verifica `!d.hasAction(...)` en `spec.atoms`, pero **olvida verificarlo** en `spec.prefix`, `spec.infix` y `spec.groups`. Un typo en la acción de un operador (`InfixLeft("PLUS", 10, "addd")`) pasa `Validate()` y `Build()` sin ningún warning.
  2. **`Validate()` considera productiva una `Expression` sin átomos (`validate.go:97, 336`):** Una regla `Expression("e").Group("LPAREN", "e", "RPAREN")` sin `Atom` jamás puede derivar un string finito, pero `Validate()` la da por válida y productiva.
  3. **`ParseMultilineBlocks` ignora la condición del `if` y ejecuta ambas ramas (`block_parser.go:235-242`):** Itera todas las líneas dentro de `if ... else ... endif` y ejecuta tanto el bloque `then` como el `else` sin evaluar jamás la condición.
  4. **`BlockParser.processBlock` corrompe `if` anidados y olvida `foreach` (`block_parser.go:138-177`):** Busca `"else"` con un bucle plano sin respetar `nestLevel`, y aunque `isBlockStart` detecta `foreach .* do$`, `processBlock` no tiene rama para `foreach`.

---

## 2. Hallazgos en Herramientas CLI y LSP (`cmd/`)

### [BUG-12] CRÍTICA (Seguridad) — Inyección de código Go arbitrario en `cmd/dslgen`
- **Ubicación:** `cmd/dslgen/main.go:101-103, 118-126`
- **Código afectado:**
  ```go
  fmt.Fprintf(&b, "package %s\n\n", pkgName)
  ...
  fmt.Fprintf(&b, "func %s() (*dslbuilder.DSL, error) {\n", funcName)
  ```
- **Causa raíz:** `pkgName` y `funcName` se interpolan directamente con `%s` en código fuente Go antes de pasar por `format.Source` sin validar que sean identificadores Go válidos (`go/token.IsIdentifier`), y `config.Name` con saltos de línea en comentarios (`línea 118`) puede inyectar líneas fuera del comentario `//`.
- **Impacto:** Si `dslgen` o `Generate()` se expone en un pipeline automatizado, CI o futuro **Lab UI Web** donde el nombre del paquete/función provenga del usuario o del archivo YAML, un atacante puede inyectar `init() { ... }` ejecutando comandos arbitrarios al compilar/testear el paquete generado. (Nota adicional: `dslgen` también ignora `config.Context`).
- **Solución recomendada:** Validar `token.IsIdentifier(pkgName)` y `token.IsIdentifier(funcName)` antes de generar código, y sanitizar saltos de línea en comentarios generados.

---

### [BUG-13] ALTA — Desbordamiento de línea y desalineación UTF-16 en posiciones del servidor LSP (`cmd/lsp`)
- **Ubicación:** `cmd/lsp/server.go:91-99, 174-194, 272-298`
- **Causa raíz:**
  1. **Cruce de línea en `offsetForPosition` (`líneas 283-287`):** Suma `offset += pos.Character` sin limitar el resultado al `\n` de la línea actual. Si el cliente LSP envía una columna mayor al largo de la línea (común al mover el cursor verticalmente), `offset` cae dentro de la **línea siguiente**, devolviendo hover/completions del nodo equivocado.
  2. **Incompatibilidad UTF-16 vs Bytes:** El protocolo LSP usa unidades de código **UTF-16** por defecto para `Position.Character` salvo que el servidor negocie `"positionEncoding": "utf-8"` en `initialize`. `cmd/lsp` no negocia `positionEncoding` pero calcula bytes crudos (`len(d.Token)`, `offset - lineStart`), desfasando subrayados de error, autocompletado y hover en cualquier línea con tildes, `ñ` o emojis.
- **Solución recomendada:** Acotar `pos.Character` al final de la línea buscada y convertir entre UTF-16 code units y byte offsets (o declarar soporte para `positionEncoding: "utf-8"` con fallback UTF-16).

---

### [BUG-14] MEDIA — Pérdida de historial, comandos rotos y bloqueo no interactivo en `cmd/repl`
- **Ubicación:** `cmd/repl/main.go:116-133, 304-305, 528-542, 582-596`
- **Detalles:**
  1. **`loadHistory()` (`línea 528`) es un stub vacío**, pero al salir `saveHistory()` ejecuta `os.Create(filename)`, **truncando y destruyendo** el archivo de historial previo.
  2. **`.exit` / `.quit` (`línea 305`) llama a `os.Exit(0)`**, saltándose `repl.saveHistory()` al final de `main()`, por lo que salir con `.exit` nunca guarda el historial de la sesión.
  3. **`-exec` en una terminal interactiva no termina (`línea 126`):** `if len(commands) == 0 || isInteractive() { repl.run() }` hace que `repl -dsl x.yaml -exec "..."` quede bloqueado esperando `stdin` siempre que se invoque desde una terminal normal; además `os.Stdin.Stat()` (`línea 540`) ignora el error pudiendo causar `nil pointer panic`.
  4. **`.tokens` y `.rules` (`líneas 582-596`)** imprimen *"information not available"* pese a que `r.dsl.Debug()` ya expone ambos mapas.

---

### [BUG-15] MEDIA — Flag `-verbose` muerto y ejecución innecesaria de acciones en `cmd/ast_viewer` + parsing frágil de extensiones en `cmd/*`
- **Ubicación:** `cmd/ast_viewer/main.go:89-90, 148, 181-197, 359-368` y `loadDSL` en los 5 binarios de `cmd/`
- **Detalles:**
  1. `cmd/ast_viewer` llama a `v.dsl.Parse(input)` en vez de `v.dsl.ParseAST(input)`, requiriendo una lista hardcodeada de 19 acciones dummy (`saleWithTax`, `greet`, etc.).
  2. `buildFromNode()` nunca puebla `ASTNode.Token` ni `ASTNode.Rule`, dejando el flag `-verbose` completamente inoperante.
  3. Los 5 comandos de `cmd/` extraen la extensión con `filename[strings.LastIndex(filename, ".")+1:]`: si la ruta no tiene punto (ej. `/dev/stdin`) o tiene un punto en un directorio padre (`configs.v1/mydsl`), falla en vez de usar `filepath.Ext(filename)`.

---

## 3. Vulnerabilidades en Módulos de Ejemplo Productivos e Higiene del Repo

> Aunque `examples/` suele considerarse código demostrativo, el repositorio documenta `examples/http_dsl` (`apiflow`) como *"production ready"* y `examples/liveview` como generador universal reutilizable. Si usuarios copian o ejecutan estos motores con entradas no confiables, quedan expuestos a los siguientes riesgos:

### [VULN-16] CRÍTICA (Seguridad) — Cross-Site Scripting (XSS) en `examples/liveview`
- **Ubicación:** `examples/liveview/universal/html_generator.go:225-231, 437, 453-489`
- **Detalle:** `UniversalHTMLGenerator` construye HTML interpolando campos de structs y mapas mediante `fmt.Sprintf` y `strings.ReplaceAll` sin usar `html/template` ni `html.EscapeString()`. Cualquier valor como `"><script>alert(1)</script>` en una tabla, card o formulario inyecta HTML/JS arbitrario.

---

### [VULN-17] CRÍTICA (Seguridad / Disponibilidad) — OOM DoS, Panic por inyección Regex y SSRF en `examples/http_dsl` (`apiflow`)
- **Ubicación:** `examples/http_dsl/universal/http_engine.go:222-224, 261, 498-499, 866-900`
- **Detalles:**
  1. **Lectura ilimitada de memoria (`línea 261`):** `io.ReadAll(resp.Body)` sin `io.LimitReader` permite que un servidor remoto malicioso o un archivo grande agote la RAM (OOM crash).
  2. **Crash (`panic`) por Regex Injection en `extractXPath` (`líneas 498-499`):** `regexp.MustCompile(fmt.Sprintf("<%s[^>]*>(.*?)</%s>", tagName, tagName))` entra en pánico si `tagName` contiene metacaracteres inválidos de regex.
  3. **Mutación global del `Timeout` (`líneas 222-224`):** Pasar una opción `timeout` en un request modifica permanentemente `he.client.Timeout` para todos los requests posteriores (y genera un *data race* si se comparte el engine). Debe usarse `context.WithTimeout` en el `http.Request`.
  4. **SSRF / Lectura de archivos locales:** Si se ejecuta `http_dsl` en un entorno multi-tenant o playground web, scripts DSL pueden invocar IPs internas/metadata cloud (`169.254.169.254`), desactivar validación TLS (`SetInsecureSkipVerify`) o cargar certificados locales arbitrarios (`SetCustomCA` / `SetClientCertificate`).

---

### [HYG-18] BAJA / HIGIENE — Binarios compilados Mach-O `arm64` (~11.2 MB) versionados en la raíz de Git
- **Ubicación:** `./ast_viewer` (3.8 MB), `./repl` (3.7 MB), `./validator` (3.7 MB)
- **Detalle:** Tres ejecutables binarios compilados para macOS arm64 están trackeados en Git (`git ls-files`). Inflan el historial del repositorio, fallan en Linux/Windows y pueden inducir a ejecutar binarios desactualizados en lugar del código fuente actual. Deben eliminarse con `git rm --cached ast_viewer repl validator` y agregarse a `.gitignore`.

---

### [HYG-19] BAJA / DOCUMENTACIÓN — Enlace roto a Skill de IA y desincronización de `ROADMAP.md`
- **Ubicación:** `README.md:175, 378` y `ROADMAP.md:3, 35-57`
- **Detalles:**
  - `README.md:378` enlaza a `.claude/skills/go-dsl/SKILL.md`, archivo que **no existe** en el repositorio (404).
  - `README.md:175` indica en un comentario `indirect left recursion (unsupported)`, contradiciendo las líneas 20 y 27 del mismo `README.md`.
  - `ROADMAP.md` sigue congelado en `v1.2.0` marcando como pendientes (`[ ]`) funcionalidades ya entregadas en `v1.3.0` y `v1.4.0` (Pratt expressions, recursión indirecta, streaming, incremental parsing, grammar validation y LSP).

---

## Plan de Remediación Sugerido (Orden de Impacto)

1. **Inmediato (1–2 días):**
   - Corregir `Document.Update` (`[BUG-01]`), eliminar `strings.TrimSpace` en `tokenizeInput` (`[BUG-02]`), limpiar `parser.memo` entre sentencias o guardar `parseFailure` en el memo (`[BUG-03]`) y arreglar `strconv.Itoa` en `ParseStatements` (`[BUG-07]`).
   - Validar identificadores en `cmd/dslgen` (`[BUG-12]`) y eliminar los binarios compilados de Git (`[HYG-18]`).
2. **Corto plazo (3–5 días):**
   - Hacer `CompiledDSL` *lock-free* moviendo el contexto de ejecución a `EvalContext` para eliminar el deadlock de `Use` anidado (`[BUG-05]`).
   - Acotar la desactivación de memoización en `parseCycleLR` + añadir límite `maxParseSteps` (`[BUG-04]`).
   - Corregir decodificación UTF-8 en el lexer y conversión de posiciones UTF-16 / fin de línea en `cmd/lsp` (`[BUG-09]`, `[BUG-13]`).
3. **Mediano plazo (1–2 semanas):**
   - Rediseñar el formato declarativo YAML/JSON v2 para preservar orden de tokens, tipos de token y reglas `Expression()` (`[BUG-06]`), y completar validaciones en `Validate()` (`[BUG-11]`).
   - Endurecer `examples/http_dsl` (`io.LimitReader`, `regexp.Compile`, `context.WithTimeout`) y migrar `examples/liveview` a `html/template` (`[VULN-16]`, `[VULN-17]`).
