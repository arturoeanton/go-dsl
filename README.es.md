# go-dsl

**Un toolkit práctico en Go para construir DSLs chicos y medianos, con parsing AST-first, evaluación lazy, validación, diagnósticos y tooling de editor.**

Definís tokens, reglas y acciones en runtime y obtenés un lenguaje funcionando: reglas de negocio, filtros de consulta, calculadoras, sintaxis de comandos. El motor parsea a un AST real primero y evalúa después (lazy donde hace falta), así los efectos secundarios son seguros por construcción.

## ✨ Características

### Motor (nuevo)
- 🌳 **Motor en Dos Fases** - `Parse()` primero construye un AST real (`ParseAST`) y después ejecuta las acciones exactamente una vez (`Eval`). Las acciones nunca corren sobre alternativas rechazadas ni durante backtracking: los efectos secundarios son seguros por construcción
- 🧾 **AST Real** - Árboles `*Node` con regla, acción, hijos, token y `Span` (posición en el código fuente); impresión con `node.Pretty()`
- 🎯 **Tokenizer Determinista** - Resolución por prioridad → match más largo → orden de declaración; los tokens que pueden matchear vacío se rechazan al definirlos
- 🎚️ **Parser de Expresiones Pratt** - Declarás operadores con binding power y asociatividad vía `Expression()`; sin malabares de gramática para la precedencia
- 🧭 **Errores de "Farthest Failure"** - Los errores de sintaxis reportan el punto más lejano alcanzado, los tokens esperados y el stack de reglas
- 🧊 **Builder Congelable** - `Build()` valida la gramática una sola vez y devuelve un `CompiledDSL` inmutable, seguro para uso concurrente
- ✅ **Validador Integrado** - `dsl.Validate()` detecta símbolos desconocidos, reglas inalcanzables, acciones sin registrar, ciclos no productivos y ciclos de recursión izquierda indirecta. La RI indirecta está soportada, pero se warnea porque deshabilita la memoización durante el growth (ver abajo)
- 🦥 **Node Actions (lazy)** - `NodeAction()` recibe el nodo sin evaluar: control de flujo real (la rama del if que no se toma, no se ejecuta)
- 🧰 **Acciones Tipadas** - Genéricos `Action1/Action2/Action3` y el helper `Args` eliminan los casts repetitivos
- 🔒 **Contexto por Parse** - `Use(code, ctx)` limita el contexto a esa llamada: ya no muta el contexto persistente del DSL; `Set/Get/SetContext/GetContext` protegidos con mutex (libre de data races)
- 🩺 **Diagnósticos Multi-Error** - `Diagnostics(code)` se recupera tras cada fallo y reporta TODOS los errores de sintaxis/léxicos en una pasada (alimenta al servidor LSP)
- 💬 **Errores por Regla** - `RuleWithError(...)` agrega un mensaje del dominio cuando esa alternativa es el farthest failure
- 🔁 **Recursión Izquierda Indirecta** - Los ciclos leftmost multi-regla (`a → b …`, `b → a …`) parsean con growing generalizado
- 🌊 **Streaming** - `ParseStream(io.Reader, handler)` procesa scripts orientados a líneas sin cargarlos en memoria
- 🖨️ **Generación de Código** - `cmd/dslgen` convierte una gramática YAML/JSON en código Go versionable
- 🧿 **Servidor LSP** - `cmd/lsp` da a cualquier editor LSP diagnósticos en vivo, **autocompletado** (expectativas del parser en el cursor) y **hover** (nodo del AST bajo el cursor)
- ⚡ **Documentos Incrementales** - `NewDocument()` re-parsea solo los statements tocados por el edit, reusando los subárboles intactos (es lo que corre el LSP por tecla)
- 🧮 **Gramáticas de Atributos** - `NewAttributeGrammar()` evalúa atributos heredados (top-down) y sintetizados (bottom-up) sobre el AST

### Núcleo
- **Recursión Izquierda Directa** - Algoritmo de semilla creciente (solo directa; para operadores usá `Expression()`)
- **Memoización (Packrat Parsing)** - Rendimiento lineal incluso con retroceso
- **Sistema de Tokens con Prioridad** - Keywords sobre patrones regulares
- **Soporte Multiline** - `ParseMultiline()`, `ParseAuto()`, `ParseWithBlocks()`
- **Configuración Declarativa** - Define DSLs con YAML/JSON
- **Herramientas CLI** - AST viewer (árbol real), validador (usa el core), REPL interactivo
- **100% Retrocompatible** - La API clásica `Parse`/`Rule`/`Action` no cambia

## 🚀 Inicio Rápido

### Instalación

```bash
go get github.com/arturoeanton/go-dsl/pkg/dslbuilder
```

### Ejemplo Básico

```go
package main

import (
    "fmt"
    "log"
    "strconv"

    "github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

func main() {
    dsl := dslbuilder.New("Calculadora")

    dsl.Token("NUMBER", `\d+`)
    dsl.Token("PLUS", `\+`)
    dsl.Token("MINUS", `-`)

    // Recursión izquierda directa: soportada
    dsl.Rule("expr", []string{"expr", "PLUS", "term"}, "add")
    dsl.Rule("expr", []string{"expr", "MINUS", "term"}, "subtract")
    dsl.Rule("expr", []string{"term"}, "pass")
    dsl.Rule("term", []string{"NUMBER"}, "number")

    dsl.Action("number", dslbuilder.Action1(strconv.Atoi))
    dsl.Action("pass", func(args []interface{}) (interface{}, error) { return args[0], nil })
    dsl.Action("add", dslbuilder.Action3(func(l int, _ string, r int) (int, error) {
        return l + r, nil
    }))
    dsl.Action("subtract", dslbuilder.Action3(func(l int, _ string, r int) (int, error) {
        return l - r, nil
    }))

    result, err := dsl.Parse("10 + 20 - 5")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(result.GetOutput()) // Output: 25
}
```

### Calculadora con la API de Alto Nivel (nuevo)

```go
calc, err := dslbuilder.New("calc").
    Tokens(func(t *dslbuilder.TokenSet) {
        t.Regex("NUMBER", `\d+`)
        t.Literal("PLUS", "+")    // texto literal, sin escapar regex
        t.Literal("STAR", "*")
        t.Literal("POW", "^")
        t.Literal("LPAREN", "(")
        t.Literal("RPAREN", ")")
        t.Literal("MINUS", "-")
    }).
    Expr("expr", func(e *dslbuilder.ExpressionBuilder) {
        e.Atom("NUMBER", "number")
        e.Group("LPAREN", "expr", "RPAREN")
        e.Prefix("MINUS", 70, "neg")
        e.InfixLeft("PLUS", 10, "add")
        e.InfixLeft("STAR", 20, "mul")
        e.InfixRight("POW", 30, "pow")
    }).
    WithAction("number", dslbuilder.Action1(strconv.Atoi)).
    // ... add, mul, pow, neg ...
    Build() // valida y congela la gramática

calc.Parse("1 + 2 * 3")   // 7
calc.Parse("(1 + 2) * 3") // 9
calc.Parse("10 - 3 - 2")  // 5   (asociativa a izquierda)
calc.Parse("2 ^ 3 ^ 2")   // 512 (asociativa a derecha)
calc.Parse("-1 * 2")      // -2  (operador prefijo)
```

### Dos Fases: AST primero, acciones después (nuevo)

```go
// Fase 1: solo parsing — ninguna acción corre, imposible tener efectos secundarios
node, err := dsl.ParseAST(`venta de 5000 con iva`)
fmt.Println(node.Pretty()) // inspeccioná el árbol

// Fase 2: evaluación, exactamente una vez sobre el árbol final
result, err := dsl.Eval(node)
```

### Control de Flujo Lazy con NodeAction (nuevo)

```go
dsl.NodeAction("ifElse", func(ctx *dslbuilder.EvalContext, n *dslbuilder.Node) (interface{}, error) {
    cond, err := ctx.Eval(n.Child(1))
    if err != nil {
        return nil, err
    }
    if cond.(bool) {
        return ctx.Eval(n.Child(3)) // solo la rama then
    }
    return ctx.Eval(n.Child(5))     // solo la rama else
})
```

### Errores Buenos de Verdad (nuevo)

```go
_, err := dsl.Parse("if status == then")
fmt.Println(dslbuilder.GetDetailedError(err))
// no alternative matched for rule condition: expected IDENT or NUMBER, got THEN "then"
// rule stack: if_stmt > condition > value at line 1, column 14:
// if status == then
//              ^
```

### Validación de Gramática (nuevo)

```go
warnings, err := dsl.Validate()
// err      -> problemas estructurales (símbolos desconocidos, tokens inválidos, ...)
// warnings -> reglas inalcanzables, acciones sin registrar, ciclos no productivos,
//             recursión izquierda indirecta (no soportada), ...

compiled, err := dsl.Build() // Validate + freeze: después de Build no se puede mutar
```

### Soporte Multiline

```go
script := `
set x 10
set y 20
print x + y
`

result, err := dsl.ParseAuto(script)        // detección automática
results, err := dsl.ParseMultiline(script)  // multiline explícito
results, err := dsl.ParseWithBlocks(script) // con soporte de bloques
```

## 📚 Ejemplos Incluidos

Tres demos oficiales cubren toda la superficie — empezá por acá:

| Demo | Qué muestra |
|---|---|
| [`examples/calculator_pratt`](examples/calculator_pratt/) | Expresiones: precedencia/asociatividad Pratt, acciones tipadas, inspección de AST, Build() |
| [`examples/http_dsl`](examples/http_dsl/) | Un DSL de scripting completo: control de flujo lazy, side effects bien hechos, testing de APIs (módulo autocontenido) |
| [`examples/scim`](examples/scim/) | Parseo de un estándar real (filtros SCIM 2.0): query DSL sobre datos |

El resto de [`examples/`](examples/) son recetas adicionales de referencia.


### 1. **Calculadora** (`examples/calculator/`)
```bash
go run examples/calculator/main.go
```

### 2. **HTTP DSL v3** (`examples/http_dsl/`)
DSL completo para operaciones HTTP con bloques (módulo Go independiente con su propia suite de tests):
```http
if $status == 200 then
    set $result "success"
    print "Operation completed"
else
    set $result "error"
endif
```

### 3. **Contabilidad multi-país** (`examples/accounting/`, `examples/contabilidad/`)
### 4. **Filtro SCIM** (`examples/scim/`)
### 5. **Query/LINQ** (`examples/query/`, `examples/linq/`, `examples/linqgo/`)

## 🔧 Recursión Izquierda: Alcance

go-dsl soporta recursión izquierda **directa** con el algoritmo de semilla creciente:

```go
dsl.Rule("list", []string{"item"}, "single")
dsl.Rule("list", []string{"list", "COMMA", "item"}, "append") // ✅ directa
```

La recursión izquierda **indirecta** (`a -> b ...` y `b -> a ...`) también
está **soportada** mediante un algoritmo de growing generalizado. `Validate()`
la sigue marcando como warning porque la memoización se deshabilita mientras
esas reglas parsean: una gramática reestructurada — o `Expression()` (Pratt)
para operadores — suele ser más clara y más rápida.

Además, las alternativas de una regla usan **elección ordenada** (estilo PEG):
declará primero las alternativas más específicas. `Validate()` warnea cuando
una alternativa corta ensombrece a una más larga con el mismo prefijo (el
error clásico).

## 🛠️ Herramientas CLI

### AST Viewer (ahora muestra el árbol real)
```bash
go install github.com/arturoeanton/go-dsl/cmd/ast_viewer@latest
ast_viewer -dsl calculator.yaml -input "2 + 3" -format tree
```

### Validador (usa `dsl.Validate()` del core)
```bash
go install github.com/arturoeanton/go-dsl/cmd/validator@latest
validator -dsl grammar.yaml -verbose -info
validator -dsl grammar.yaml -test "entrada de prueba"
```

### REPL
```bash
go install github.com/arturoeanton/go-dsl/cmd/repl@latest
repl -dsl calculator.yaml -context data.json
```

### Generador de Código
Convierte una gramática declarativa en código Go versionable (constructor + stubs de acciones):
```bash
go install github.com/arturoeanton/go-dsl/cmd/dslgen@latest
dslgen -dsl grammar.yaml -package midsl -func NewMiDSL -o midsl_gen.go
```

### Servidor LSP
Diagnósticos en vivo, autocompletado y hover en cualquier editor con LSP.
Los documentos se re-parsean incrementalmente (solo los statements editados),
el completado sale de las expectativas del propio parser y el hover muestra
el nodo del AST (regla, acción, span) bajo el cursor:
```bash
go install github.com/arturoeanton/go-dsl/cmd/lsp@latest
lsp -dsl grammar.yaml    # conectar como language server por stdio
```

Las mismas capacidades existen como API de librería: `dsl.NewDocument()`
(parse incremental + `NodeAt`), `dsl.Completions(text, offset)` y
`dslbuilder.NewAttributeGrammar()` para análisis semántico sobre el AST.

## 📊 Comparación con Otras Herramientas

| Característica | go-dsl | ANTLR | PEG | Yacc |
|---------------|--------|-------|-----|------|
| Recursión Izquierda (directa e indirecta) | ✅ | ✅ | ❌ | ✅ |
| Parser Pratt Integrado | ✅ | ❌ | ❌ | ❌ |
| AST + Eval Separados | ✅ | ✅ | ⚠️ | ⚠️ |
| Sin Generación de Código | ✅ | ❌ | ❌ | ❌ |
| Configuración Runtime | ✅ | ❌ | ❌ | ❌ |
| Memoización | ✅ | ❌ | ✅ | ❌ |
| YAML/JSON Config | ✅ | ❌ | ❌ | ❌ |

## 🧪 Testing y CI

```bash
./scripts/check.sh               # el quality gate completo, local y gratis
go test ./...                    # módulo principal
go test -race ./pkg/dslbuilder/...  # el core es race-clean
cd examples/http_dsl && go test ./...  # módulo del HTTP DSL

# Guardia de regresión de performance (misma máquina que el baseline):
./scripts/bench_baseline.sh      # (re)genera benchmarks/baseline.txt
./scripts/bench_guard.sh         # compara benchmarks actuales con benchstat
```

El CI corre el mismo gate en un único job de GitHub Actions
(`.github/workflows/ci.yml`), gratis para repos públicos. Los targets de
fuzzing (`go test -fuzz FuzzExpressionParse ./pkg/dslbuilder`) y su corpus de
regresión viven en la suite.

## 📖 Documentación

- [Guía Rápida](docs/es/guia_rapida.md)
- [Skill para IAs/agentes](.claude/skills/go-dsl/SKILL.md) — guía lista para que agentes de código (Claude Code, etc.) usen go-dsl correctamente en proyectos Go
- [Conceptos Avanzados](docs/es/introduccion_dsl_segunda_parte.md)
- [Limitaciones](docs/es/limitaciones.md)
- [Ejemplos](examples/)

## 🤝 Contribuir

¡Las contribuciones son bienvenidas! Áreas prioritarias:

1. Optimizaciones de rendimiento
2. Nuevos ejemplos de DSL
3. Mejoras en documentación
4. Corrección de bugs y tests

## 📜 Licencia

Apache License 2.0 - Ver [LICENSE](LICENSE) para detalles.

## 🙏 Agradecimientos

- Inspirado en parsers PEG y Packrat
- Precedencia de operadores con precedence climbing (Pratt)
- Comunidad Go por el feedback y contribuciones

---

Hecho con ❤️ por la comunidad go-dsl
