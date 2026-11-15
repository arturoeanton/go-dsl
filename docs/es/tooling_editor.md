# Tooling de Editor y Experiencia de Desarrollo

go-dsl incluye todo lo necesario para darle a un DSL propio una experiencia de
desarrollo profesional: **DSL Studio** (IDE visual), **servidor LSP** completo,
documentos incrementales, autocompletado derivado de la gramática, formateador
universal, testing y fuzzing declarativo, generación de código Go tipado,
scaffolding de extensión VS Code, servidor MCP y bridge WebAssembly.

## Laboratorio Visual (`cmd/studio`)

Para diseñar gramáticas interactivamente, inspeccionar AST y trazas paso a paso,
fuzzear reglas y trabajar con modelos locales en **Ollama** o **LM Studio** vía
GBNF:

```bash
go run ./cmd/studio -open
go run ./cmd/studio -dsl grammar.yaml -open
```

Consultá la [Guía completa de DSL Studio](dsl_studio.md).

## El Servidor LSP (`cmd/lsp`)

```bash
go install github.com/arturoeanton/go-dsl/cmd/lsp@latest
lsp -dsl grammar.yaml     # language server por stdio
```

Conectalo a cualquier editor compatible con LSP (VS Code, Zed, Neovim, Helix, ...):

| Capacidad | Comportamiento |
|---|---|
| **Diagnósticos** | Diagnósticos multi-error con conversión exacta UTF-16 al abrir/editar |
| **Completado** | Sugerencias calculadas por el propio parser en el cursor: keywords/literales como texto directo y tokens libres (`NUMBER`, `STRING`) como snippets (`${1:NUMBER}`) |
| **Hover** | Nodo del AST bajo el cursor: tipo y valor del token, o regla + acción + fragmento fuente |
| **Formateo** | `textDocument/formatting` respaldado por `DSL.Format` (idempotente, preserva comentarios) |
| **Semantic Tokens** | `textDocument/semanticTokens/full` para coloreado semántico de comentarios, keywords, números, strings, operadores e identificadores |

## Extensión VS Code en 1 Clic y Codegen Tipado (`cmd/dslgen`)

```bash
# Generar constructor Go + interfaz de acciones fuertemente tipada
dslgen -dsl grammar.yaml -package midsl -func NewMiDSL -typed -o midsl_gen.go

# Generar extensión completa para VS Code (TextMate + cliente LSP)
dslgen -dsl grammar.yaml -vscode ./vscode-midsl
```

## Formateador (`cmd/dslfmt`) y Tests Declarativos (`cmd/dsltest`)

```bash
# Formatear archivos in-place o verificar formato en CI
dslfmt -dsl grammar.yaml -w script.dsl
dslfmt -dsl grammar.yaml -check script.dsl

# Ejecutar tests embebidos en YAML + sintetizar muestras + fuzzing
dsltest -dsl grammar.yaml -fuzz 250 -samples 5
```

## Servidor MCP (`cmd/mcp`) y WebAssembly (`cmd/wasm`)

```bash
# Exponer validación, completado, formato y GBNF por stdio MCP
mcp -dsl grammar.yaml

# Compilar binario WebAssembly para navegador/playgrounds
GOOS=js GOARCH=wasm go build -o go-dsl.wasm ./cmd/wasm
```

## Las APIs de Librería Detrás

Todo lo que ejecutan los CLIs es API pública reutilizable en tus propios servicios:

- **Documentos incrementales**: `doc := dsl.NewDocument()` re-parsea únicamente las sentencias afectadas por cada edición (`Update`, `Diagnostics`, `Statements`, `NodeAt`, `Stats`).
- **Completado y Formateo**: `dsl.Completions(text, offset)` y `dsl.Format(text)`.
- **Gramáticas de atributos**: `dslbuilder.NewAttributeGrammar()` con pasadas `Inherited` (top-down) y `Synthesized` (bottom-up).

## El Patrón Producto: `apiflow`

`examples/http_dsl` muestra cómo empaquetar un DSL como producto CLI propio:

```bash
cd examples/http_dsl
go run ./cmd/apiflow run script.http       # ejecutar
go run ./cmd/apiflow check script.http     # validar — solo fase AST, cero efectos secundarios
```
