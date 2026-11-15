//go:build js && wasm

// Command wasm compiles go-dsl into a zero-backend WebAssembly module exposing
// globalThis.GoDSL to browser JavaScript / TypeScript applications.
//
// Build with:
//
//	GOOS=js GOARCH=wasm go build -o godsl.wasm ./cmd/wasm
package main

import (
	"syscall/js"

	"github.com/arturoeanton/go-dsl/pkg/wasm"
)

func main() {
	engine := wasm.NewEngine()

	api := map[string]interface{}{
		"loadGrammar": js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
			if len(args) < 1 {
				return `{"ok":false,"error":"loadGrammar(spec, [format]) requires at least 1 argument"}`
			}
			format := "auto"
			if len(args) >= 2 {
				format = args[1].String()
			}
			return engine.LoadGrammar(args[0].String(), format)
		}),
		"validateGrammar": js.FuncOf(func(_ js.Value, _ []js.Value) interface{} {
			return engine.ValidateGrammar()
		}),
		"parse": js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
			if len(args) < 1 {
				return `{"ok":false,"error":"parse(code, [maxSteps]) requires at least 1 argument"}`
			}
			maxSteps := 10000
			if len(args) >= 2 && args[1].Type() == js.TypeNumber {
				maxSteps = args[1].Int()
			}
			return engine.Parse(args[0].String(), maxSteps)
		}),
		"diagnostics": js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
			if len(args) < 1 {
				return `{"ok":false,"error":"diagnostics(code) requires 1 argument"}`
			}
			return engine.Diagnostics(args[0].String())
		}),
		"completions": js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
			if len(args) < 2 {
				return `{"ok":false,"error":"completions(code, offset) requires 2 arguments"}`
			}
			return engine.Completions(args[0].String(), args[1].Int())
		}),
		"format": js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
			if len(args) < 1 {
				return `{"ok":false,"error":"format(code) requires 1 argument"}`
			}
			return engine.Format(args[0].String())
		}),
		"toGBNF": js.FuncOf(func(_ js.Value, _ []js.Value) interface{} {
			return engine.ToGBNF()
		}),
		"aiPromptSpec": js.FuncOf(func(_ js.Value, _ []js.Value) interface{} {
			return engine.AIPromptSpec()
		}),
		"repairPrompt": js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
			if len(args) < 1 {
				return `{"ok":false,"error":"repairPrompt(code) requires 1 argument"}`
			}
			return engine.RepairPrompt(args[0].String())
		}),
		"generateSamples": js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
			seed := int64(1)
			count := 5
			if len(args) >= 1 && args[0].Type() == js.TypeNumber {
				seed = int64(args[0].Int())
			}
			if len(args) >= 2 && args[1].Type() == js.TypeNumber {
				count = args[1].Int()
			}
			return engine.GenerateSamples(seed, count)
		}),
		"runTests": js.FuncOf(func(_ js.Value, _ []js.Value) interface{} {
			return engine.RunTests()
		}),
		"fuzz": js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
			seed := int64(42)
			iters := 200
			if len(args) >= 1 && args[0].Type() == js.TypeNumber {
				seed = int64(args[0].Int())
			}
			if len(args) >= 2 && args[1].Type() == js.TypeNumber {
				iters = args[1].Int()
			}
			return engine.Fuzz(seed, iters)
		}),
	}

	js.Global().Set("GoDSL", js.ValueOf(api))
	select {}
}
