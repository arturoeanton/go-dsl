# DSL Studio — Laboratorio Visual de Lenguajes y Copiloto Local

`cmd/studio` provee un IDE visual local en un único binario Go (`embed.FS`, sin
Node ni npm) para diseñar, probar, depurar, fuzzear y exportar lenguajes creados
con `go-dsl`.

## Inicio Rápido

```bash
# Levantar DSL Studio en http://127.0.0.1:8420 y abrir el navegador
go run ./cmd/studio -open

# Abrir una gramática YAML/JSON existente para editarla y guardarla in-place
go run ./cmd/studio -dsl examples/declarative/configs/math_dsl.yaml -port 8420 -open
```

### Flags de CLI

| Flag | Default | Propósito |
|---|---|---|
| `-host` | `127.0.0.1` | Dirección de escucha |
| `-port` | `8420` | Puerto HTTP |
| `-dsl` | *(preset)* | Ruta a un archivo `.yaml`/`.json` para editar y guardar en disco |
| `-open` | `false` | Abrir automáticamente el navegador al iniciar |
| `-allow-remote-ai` | `false` | Permitir endpoints LLM fuera de loopback (bloqueados por defecto contra SSRF) |

---

## Arquitectura de Tres Paneles

### 1. Panel de Gramática (Izquierda)
- **Editor YAML v2**: edición en vivo de `token_defs`, `rules`, `expressions` (Pratt), `layouts`, `significant_newlines`, `start_rule` y `tests:` con validación estática instantánea (`Validate()`).
- **Mapa Visual del Lenguaje**: vista estructural con chips de tokens por prioridad (`skip`, `keyword`, `literal`, `regex`), reglas BNF con sus acciones y **escalera visual de precedencias Pratt** ordenada por `power` (`prefix`, `postfix`, `left`, `right`, `nonassoc`, `ternary`).
- **Presets Incluidos**: 4 gramáticas v2 completas listas para explorar (*Fintech Pricing Rules*, *SmartQuery Filter*, *Workflow Ops* y *Pratt Scientific Calculator*).

### 2. Sandbox en Vivo y Cinta de Tokens (Centro)
- **Evaluación con Presupuesto (`MaxSteps`)**: ejecuta `ParseTrace` bajo límites configurables de `EvalOptions`.
- **Formateador Universal (`✨ Formatear`)**: aplica `DSL.Format` de manera idempotente preservando comentarios.
- **Autocompletado del Parser**: chips clicables calculados con `DSL.Completions(code, offset)` según los tokens válidos esperados en la posición exacta del cursor.
- **Cinta de Tokens y Trivia**: muestra cada token reconocido junto con comentarios ignorados (`skip`); al hacer clic selecciona su rango exacto (`Span`) en el editor.

### 3. Inspector de Seis Pestañas (Derecha)
1. **Salida y Trace (`⚡`)**: valor resultante y árbol `EvalTrace` con tiempos en microsegundos (`µs`) y ramas lazy no ejecutadas marcadas como `[SKIPPED]`.
2. **AST (`🌳`)**: árbol sintáctico `*Node` interactivo vinculado a las posiciones del código fuente.
3. **Modelos Locales (`🤖`)**: integración nativa con **Ollama** (`http://127.0.0.1:11434`) y **LM Studio** (`http://127.0.0.1:1234`).
4. **Tests y Fuzz (`🧪`)**: ejecuta suites declarativas `tests:` embebidas en el YAML (`RunConfigTests`), `FuzzGrammar` y generación de ejemplos válidos clicables (`GenerateSamples`).
5. **GBNF y Spec (`🧬`)**: exportación en vivo de `dsl.ToGBNF()` y `dsl.AIPromptSpec()` listos para copiar.
6. **Exportar (`📦`)**: generación de código Go estático (incluyendo interfaz tipada `-typed`) y descarga directa en `.zip` de una extensión completa para **VS Code** (`package.json`, gramática TextMate, configuración de lenguaje y cliente LSP).

---

## Integración con Ollama y LM Studio

DSL Studio se comunica desde el backend local en Go hacia **Ollama** y **LM Studio**, evitando bloqueos de CORS del navegador y manteniendo las gramáticas y reglas 100% en `localhost`:

1. **Auto-detección concurrente**: consulta `GET /api/tags` (Ollama) y `GET /v1/models` (LM Studio), destacando modelos orientados a código (`qwen2.5-coder`, `deepseek-coder`, `codestral`, `llama3.1`/`3.3`).
2. **Capa 1 — Decodificación Restringida con GBNF (`ToGBNF`)**: envía la gramática formal `grammar: dsl.ToGBNF()` en `/v1/chat/completions` cuando el backend GGUF/`llama.cpp` lo soporta, con **fallback automático** transparente si el motor utiliza MLX o rechaza el campo `grammar`.
3. **Capa 2 — Bucle Cerrado de Auto-Reparación (`RepairPrompt`)**: verifica cada salida con `dsl.ParseAST`. Si hay error sintáctico, reinyecta `dsl.RepairPrompt(candidato)` (línea, columna, puntero visual y tokens válidos esperados) hasta **3 rondas** y formatea el código final con `dsl.Format` antes de entregarlo al editor.

### Modos Disponibles
- **Lenguaje Natural → Código DSL Verificado**
- **Auto-Reparación en 1 Clic desde el Banner de Diagnósticos**
- **Diseño de Gramáticas YAML v2 Validadas (`LoadFromYAML` + `Validate()`)**
- **Síntesis de Suites Declarativas `tests:` Verificadas contra el Parser**
