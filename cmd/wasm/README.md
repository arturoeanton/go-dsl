# WebAssembly Bridge (`cmd/wasm`)

Compiles `pkg/wasm.Engine` to WebAssembly (`GOOS=js GOARCH=wasm`) and exposes `globalThis.GoDSL` in JavaScript environments.

## Build

```bash
GOOS=js GOARCH=wasm go build -o go-dsl.wasm ./cmd/wasm
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" .
```

## Exposed JS Methods (`globalThis.GoDSL`)

- `loadGrammar(spec, format)`
- `validateGrammar(spec, format)`
- `parse(code, maxSteps)`
- `complete(code, offset)`
- `format(code)`
- `toGBNF()`
- `repairPrompt(code)`
- `runTests()`
- `fuzz(seed, iterations)`
