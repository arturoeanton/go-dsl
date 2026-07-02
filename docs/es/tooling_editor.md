# Tooling de Editor y Experiencia de Desarrollo

go-dsl trae todo lo necesario para darle a un DSL propio una experiencia de
editor de primera: language server, documentos incrementales, completado
derivado de la propia gramática, diagnósticos multi-error y generación de
código.

## El servidor LSP (`cmd/lsp`)

```bash
go install github.com/arturoeanton/go-dsl/cmd/lsp@latest
lsp -dsl grammar.yaml     # language server por stdio
```

Conectalo a cualquier editor con LSP (VS Code, Zed, Neovim, Helix, ...) como
servidor stdio para la extensión de archivo de tu DSL. Capacidades:

| Capacidad | Comportamiento |
|---|---|
| **Diagnósticos** | Al abrir/editar se subrayan TODOS los errores del documento (recovery multi-error, no solo el primero) |
| **Completado** | Las sugerencias son las expectativas del propio parser en el cursor: keywords/literales como texto insertable, tokens libres (NUMBER, STRING, ...) como snippets con placeholder |
| **Hover** | El nodo del AST bajo el cursor: tipo y texto del token, o regla + acción + snippet del código |

Los documentos se re-parsean **incrementalmente**: solo los statements
tocados por el edit se parsean de nuevo (ver abajo).

## Las APIs de librería detrás

Todo lo que hace el LSP es API pública que podés embeber en tus herramientas.

### Documentos incrementales

```go
doc := dsl.NewDocument()
doc.Update(text)                  // primer parse completo
doc.Update(textEditado)           // SOLO los statements editados se re-parsean
doc.Stats()                       // {Reused: 4, Reparsed: 1} — verificable
doc.Diagnostics()                 // todos los errores actuales
doc.Statements()                  // árboles de parseo, en orden
node := doc.NodeAt(byteOffset)    // nodo más interno en una posición
```

**Granularidad — importante**: la incrementalidad es por *unidad top-level
de parseo* (cada match de la regla start, típicamente "un statement"). No es
magia incremental arbitraria: el prefijo intacto reusa sus árboles tal cual,
el sufijo intacto los reusa con spans desplazados, y **la unidad editada se
re-parsea completa**. El lexing sigue siendo una pasada lineal (es barato);
el ahorro está en el parseo y la construcción de árboles. Para gramáticas
donde todo el documento es una sola unidad top-level, cada edit re-parsea
esa unidad entera.

### Completado

```go
comps := dsl.Completions(text, cursorOffset)
for _, c := range comps {
    // c.Label     "set" o "NUMBER"
    // c.IsKeyword true → texto insertable; false → placeholder/snippet
    // c.Detail    "keyword" o "pattern: [0-9]+"
}
```

A mitad de statement, los candidatos salen de las expectativas del farthest
failure en el cursor — exactamente lo que el parser aceptaría después. En un
límite de statement, los candidatos son los tokens que pueden iniciar uno
(conjunto FIRST).

### Diagnósticos multi-error

```go
for _, diag := range dsl.Diagnostics(script) {
    fmt.Println(diag.DetailedError()) // línea, columna, expected, puntero
}
```

Recovery panic-mode con resincronización por FIRST: tras un error, el parser
salta al próximo token que puede iniciar un statement, evitando cascadas de
errores espurios. Tope: 50 diagnósticos.

### Gramáticas de atributos (análisis semántico)

```go
ag := dslbuilder.NewAttributeGrammar()

// Top-down: propagar un entorno/scope.
ag.Inherited("env", func(parent dslbuilder.Attrs, n *dslbuilder.Node, i int) interface{} {
    return parent["env"]
})

// Bottom-up: sintetizar un tipo/valor por regla.
ag.Synthesized("tipo", "expr", func(n *dslbuilder.Node, children []dslbuilder.Attrs) (interface{}, error) {
    return inferirTipo(n, children)
})

// El contexto inicial actúa como "atributos del padre" de la raíz:
attrs, err := ag.Evaluate(root, dslbuilder.Attrs{"env": "prod"})
_ = attrs[root]["env"]  // "prod", propagado a todo el árbol
```

Una pasada top-down (heredados) seguida de una bottom-up (sintetizados) —
el subconjunto L-attributed, que cubre tablas de símbolos, scoping y
síntesis de tipos sin walkers a mano.

## Generación de código (`cmd/dslgen`)

```bash
dslgen -dsl grammar.yaml -package midsl -func NewMiDSL -o midsl_gen.go
```

El archivo generado reconstruye el DSL programáticamente (los keywords
hacen round-trip correcto) y lista en su doc comment las acciones que falta
registrar. Cero carga de archivos en runtime.

## El patrón producto: apiflow

`examples/http_dsl` muestra cómo un DSL se gradúa a producto: el motor es
go-dsl, el lenguaje vive en un paquete de librería, y un CLI chico es la
cara al usuario:

```bash
cd examples/http_dsl
go run ./cmd/apiflow run script.http       # ejecutar
go run ./cmd/apiflow check script.http     # validar — solo fase AST, cero side effects
```

`check` no puede disparar un request jamás: usa `ParseAST`, que por diseño
no ejecuta acciones. Valida además el balance estructural de bloques
(if/endif, loops/endloop); la semántica interna de un bloque se verifica
statement a statement.
