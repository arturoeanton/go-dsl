# DSL Studio — Visual Language Lab & Local LLM Workbench

`cmd/studio` ships a single-binary, zero-Node/npm local IDE (`embed.FS`) for
designing, testing, debugging, fuzzing, and exporting `go-dsl` languages.

## Quick Launch

```bash
# Start DSL Studio on http://127.0.0.1:8420 and open the browser
go run ./cmd/studio -open

# Open an existing YAML/JSON grammar for live editing and in-place saving
go run ./cmd/studio -dsl examples/declarative/configs/math_dsl.yaml -port 8420 -open
```

### CLI Flags

| Flag | Default | Purpose |
|---|---|---|
| `-host` | `127.0.0.1` | Bind address |
| `-port` | `8420` | HTTP port |
| `-dsl` | *(preset)* | Path to a `.yaml`/`.json` grammar file to edit and save in place |
| `-open` | `false` | Open the default browser automatically on startup |
| `-allow-remote-ai` | `false` | Allow non-loopback LLM endpoints (blocked by default to prevent accidental SSRF) |

---

## Three-Pane Architecture

### 1. Grammar Pane (Left)
- **YAML v2 Editor**: edit `token_defs`, `rules`, `expressions` (Pratt), `layouts`, `significant_newlines`, `start_rule`, and embedded `tests:` with instant `Validate()` feedback.
- **Visual Grammar Map**: switch to the visual view to inspect token priority chips (`skip`, `keyword`, `literal`, `regex`), BNF production sequences, and the **Pratt Precedence Ladder** sorted by binding power (`prefix`, `postfix`, `left`, `right`, `nonassoc`, `ternary`).
- **Curated Presets**: includes 4 complete v2 grammars out of the box (*Fintech Pricing Rules*, *SmartQuery Filter*, *Workflow Ops*, and *Pratt Scientific Calculator*).

### 2. Live Sandbox & Token Stream (Center)
- **Step-Budgeted Evaluation**: runs `ParseTrace` under a configurable `MaxSteps` execution budget (`EvalOptions`).
- **Universal Formatter (`✨ Format`)**: formats the buffer idempotently via `DSL.Format` while preserving comments.
- **Parser-Driven Completions**: clickable completion chips generated from `DSL.Completions(code, cursorByteOffset)` showing exact parser expectations at the cursor.
- **Interactive Token & Trivia Ribbon**: displays matched tokens alongside skipped comments/trivia; clicking any token highlights its exact byte span in the editor.

### 3. Six-Tab Inspector (Right)
1. **Output & Trace (`⚡`)**: displays the evaluated result and the hierarchical `EvalTrace` tree with per-node microsecond timings and `[SKIPPED]` markers on untaken lazy `NodeAction` branches.
2. **AST (`🌳`)**: interactive `*Node` syntax tree; clicking any rule or token node selects `[Start..End]` in the source editor.
3. **Local LLM Workbench (`🤖`)**: native integration with **Ollama** (`http://127.0.0.1:11434`) and **LM Studio** (`http://127.0.0.1:1234`).
4. **Tests & Fuzz (`🧪`)**: runs declarative `tests:` embedded in the grammar YAML (`RunConfigTests`) plus `FuzzGrammar` iterations and clickable valid sample synthesis (`GenerateSamples`).
5. **GBNF & Spec (`🧬`)**: live `dsl.ToGBNF()` grammar and `dsl.AIPromptSpec()` Markdown specification with 1-click copy.
6. **Export (`📦`)**: generates static Go parser code (`-typed` action interface supported) and downloads a ready-to-install `.zip` **VS Code Extension** (`package.json`, TextMate grammar, language configuration, and LSP client).

---

## Local Ollama & LM Studio Integration

DSL Studio connects directly from the local Go server to **Ollama** and **LM Studio** (avoiding browser CORS restrictions while keeping 100% of grammars and business rules on `localhost`):

1. **Auto-Discovery**: probes `GET /api/tags` (Ollama) and `GET /v1/models` (LM Studio) concurrently and highlights code-oriented models (`qwen2.5-coder`, `deepseek-coder`, `codestral`, `llama3.1`/`3.3`).
2. **Layer 1 — Formal GBNF Constrained Decoding**: sends `grammar: dsl.ToGBNF()` on `/v1/chat/completions` when supported by GGUF/`llama.cpp` backends, with transparent automatic fallback to unconstrained chat if the runtime uses MLX or rejects raw GBNF.
3. **Layer 2 — Closed-Loop Self-Repair (`RepairPrompt`)**: validates every candidate with `dsl.ParseAST`. If any syntax error occurs, Studio feeds `dsl.RepairPrompt(candidate)` (line, column, caret context, and valid next tokens) back into the local model for up to **3 rounds**, then formats the verified result with `dsl.Format` before inserting it into the editor.

### Supported Workflows
- **Natural Language → Verified DSL Code**: synthesizes domain scripts guaranteed to pass `ParseAST`.
- **1-Click Syntax Repair**: repairs broken scripts directly from the diagnostics banner.
- **Natural Language → YAML v2 Grammar**: designs complete grammars verified against `LoadFromYAML` + `Validate()`.
- **Declarative Test Synthesis**: generates positive and negative `tests:` cases verified against the active parser.
