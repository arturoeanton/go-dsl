// Package wasm provides a JSON-serializable Playground/Studio bridge around
// dslbuilder.DSL so that WebAssembly (cmd/wasm) and DSL Studio (cmd/studio)
// can load grammars, parse ASTs, inspect execution traces, compute diagnostics
// and completions, format code, fuzz grammars, and export GBNF.
package wasm

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
	"github.com/arturoeanton/go-dsl/pkg/dslgen"
)

// Engine wraps a mutable active *dslbuilder.DSL for interactive browser/WASM and Studio sessions.
type Engine struct {
	mu  sync.RWMutex
	dsl *dslbuilder.DSL
}

// NewEngine creates an empty WASM/Studio engine instance.
func NewEngine() *Engine {
	return &Engine{}
}

// Response is the uniform JSON envelope returned by Engine methods.
type Response struct {
	OK           bool                      `json:"ok"`
	Error        string                    `json:"error,omitempty"`
	Warnings     []string                  `json:"warnings,omitempty"`
	Config       *dslbuilder.DSLConfig     `json:"config,omitempty"`
	AST          *ASTView                  `json:"ast,omitempty"`
	ASTPretty    string                    `json:"ast_pretty,omitempty"`
	Output       interface{}               `json:"output,omitempty"`
	Trace        *dslbuilder.EvalTrace     `json:"trace,omitempty"`
	Tokens       []dslbuilder.TokenMatch   `json:"tokens,omitempty"`
	Trivia       []dslbuilder.TokenMatch   `json:"trivia,omitempty"`
	Diagnostics  []DiagnosticItem          `json:"diagnostics,omitempty"`
	Completions  []dslbuilder.Completion   `json:"completions,omitempty"`
	Formatted    string                    `json:"formatted,omitempty"`
	GBNF         string                    `json:"gbnf,omitempty"`
	PromptSpec   string                    `json:"prompt_spec,omitempty"`
	RepairPrompt string                    `json:"repair_prompt,omitempty"`
	Clean        bool                      `json:"clean,omitempty"`
	Samples      []string                  `json:"samples,omitempty"`
	TestReport   *dslbuilder.DSLTestReport `json:"test_report,omitempty"`
	FuzzReport   *dslbuilder.FuzzReport    `json:"fuzz_report,omitempty"`
}

// ASTView is a clean JSON representation of a dslbuilder.Node.
type ASTView struct {
	Rule     string     `json:"rule,omitempty"`
	Action   string     `json:"action,omitempty"`
	Token    string     `json:"token,omitempty"`
	TokenVal string     `json:"value,omitempty"`
	Start    int        `json:"start"`
	End      int        `json:"end"`
	Children []*ASTView `json:"children,omitempty"`
}

// DiagnosticItem is a JSON-friendly representation of a dslbuilder.ParseError.
type DiagnosticItem struct {
	Message  string `json:"message"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Position int    `json:"position"`
	Token    string `json:"token,omitempty"`
	Detailed string `json:"detailed,omitempty"`
}

func encode(r Response) string {
	b, err := json.Marshal(r)
	if err != nil {
		return fmt.Sprintf(`{"ok":false,"error":%q}`, err.Error())
	}
	return string(b)
}

func toASTView(n *dslbuilder.Node) *ASTView {
	if n == nil {
		return nil
	}
	v := &ASTView{
		Rule:   n.Rule,
		Action: n.Action,
		Start:  n.Span.Start,
		End:    n.Span.End,
	}
	if n.Token != nil {
		v.Token = n.Token.TokenType
		v.TokenVal = n.Token.Value
	}
	if len(n.Children) > 0 {
		v.Children = make([]*ASTView, 0, len(n.Children))
		for _, ch := range n.Children {
			v.Children = append(v.Children, toASTView(ch))
		}
	}
	return v
}

// DSL returns a snapshot pointer to the currently loaded *dslbuilder.DSL (or nil).
func (e *Engine) DSL() *dslbuilder.DSL {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.dsl
}

// LoadGrammar loads a YAML or JSON grammar definition into the engine.
// If format is "" or "auto", JSON is detected when trimmed input starts with "{".
func (e *Engine) LoadGrammar(spec string, format string) string {
	trimmed := strings.TrimSpace(spec)
	if trimmed == "" {
		return encode(Response{OK: false, Error: "grammar specification is empty"})
	}

	useJSON := strings.EqualFold(format, "json") || (format == "" || strings.EqualFold(format, "auto")) && strings.HasPrefix(trimmed, "{")

	var (
		d   *dslbuilder.DSL
		err error
	)
	if useJSON {
		d, err = dslbuilder.LoadFromJSON([]byte(spec))
	} else {
		d, err = dslbuilder.LoadFromYAML([]byte(spec))
	}
	if err != nil {
		return encode(Response{OK: false, Error: err.Error()})
	}

	// Capture static grammar validation before attaching Studio interactive actions.
	warnings, valErr := d.Validate()
	registerStudioActions(d)

	cfg := d.ToConfig()
	e.mu.Lock()
	e.dsl = d
	e.mu.Unlock()

	if valErr != nil {
		return encode(Response{OK: false, Error: valErr.Error(), Warnings: warnings, Config: &cfg})
	}
	return encode(Response{OK: true, Warnings: warnings, Config: &cfg})
}

// registerStudioActions wires built-in arithmetic/logic/comparison evaluators plus
// structured AST node evaluation for custom domain actions so interactive Studio
// sessions execute live without requiring Go compilation.
func registerStudioActions(d *dslbuilder.DSL) {
	cfg := d.ToConfig()
	allActions := dslgen.CollectAllActions(&cfg)

	for _, name := range allActions {
		actName := name
		if fn, ok := builtinStudioAction(actName); ok {
			d.Action(actName, fn)
		} else {
			d.Action(actName, func(args []interface{}) (interface{}, error) {
				if len(args) == 1 {
					return args[0], nil
				}
				return map[string]interface{}{
					"$action": actName,
					"args":    args,
				}, nil
			})
		}
	}
}

func builtinStudioAction(name string) (dslbuilder.ActionFunc, bool) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "number", "num", "num_lit", "float", "int":
		return func(args []interface{}) (interface{}, error) {
			if len(args) == 0 {
				return 0, nil
			}
			return parseNumOrPass(args[0]), nil
		}, true
	case "string", "str", "str_lit":
		return func(args []interface{}) (interface{}, error) {
			if len(args) == 0 {
				return "", nil
			}
			return strings.Trim(fmt.Sprintf("%v", args[0]), `"'`), nil
		}, true
	case "bool", "boolean":
		return func(args []interface{}) (interface{}, error) {
			if len(args) == 0 {
				return false, nil
			}
			return toBool(args[0]), nil
		}, true
	case "passthrough", "pass", "group", "paren", "ident", "id":
		return func(args []interface{}) (interface{}, error) {
			if len(args) == 1 {
				return args[0], nil
			}
			if len(args) == 3 {
				return args[1], nil
			}
			return args, nil
		}, true
	case "neg", "negate", "unary_minus":
		return func(args []interface{}) (interface{}, error) {
			if len(args) >= 2 {
				return -toFloat(args[1]), nil
			}
			if len(args) == 1 {
				return -toFloat(args[0]), nil
			}
			return 0.0, nil
		}, true
	case "fact", "factorial":
		return func(args []interface{}) (interface{}, error) {
			n := int(toFloat(args[0]))
			res := 1
			for i := 2; i <= n && i <= 20; i++ {
				res *= i
			}
			return res, nil
		}, true
	case "add", "plus", "add_op":
		return func(args []interface{}) (interface{}, error) {
			if len(args) >= 3 {
				l, lok := tryFloat(args[0])
				r, rok := tryFloat(args[2])
				if lok && rok {
					return cleanNumber(l + r), nil
				}
				return fmt.Sprintf("%v%v", args[0], args[2]), nil
			}
			return args, nil
		}, true
	case "sub", "subtract", "minus", "sub_op":
		return func(args []interface{}) (interface{}, error) {
			if len(args) >= 3 {
				return cleanNumber(toFloat(args[0]) - toFloat(args[2])), nil
			}
			return args, nil
		}, true
	case "mul", "multiply", "times", "mul_op":
		return func(args []interface{}) (interface{}, error) {
			if len(args) >= 3 {
				return cleanNumber(toFloat(args[0]) * toFloat(args[2])), nil
			}
			return args, nil
		}, true
	case "div", "divide", "div_op":
		return func(args []interface{}) (interface{}, error) {
			if len(args) >= 3 {
				r := toFloat(args[2])
				if r == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				return cleanNumber(toFloat(args[0]) / r), nil
			}
			return args, nil
		}, true
	case "pow", "power", "exp":
		return func(args []interface{}) (interface{}, error) {
			if len(args) >= 3 {
				return cleanNumber(math.Pow(toFloat(args[0]), toFloat(args[2]))), nil
			}
			return args, nil
		}, true
	case "percent", "pct":
		return func(args []interface{}) (interface{}, error) {
			if len(args) >= 1 {
				return cleanNumber(toFloat(args[0]) / 100.0), nil
			}
			return 0.0, nil
		}, true
	case "eq", "equals":
		return func(args []interface{}) (interface{}, error) {
			if len(args) >= 3 {
				l, lok := tryFloat(args[0])
				r, rok := tryFloat(args[2])
				if lok && rok {
					return l == r, nil
				}
				return strings.Trim(fmt.Sprintf("%v", args[0]), `"'`) == strings.Trim(fmt.Sprintf("%v", args[2]), `"'`), nil
			}
			return false, nil
		}, true
	case "neq", "not_equals":
		return func(args []interface{}) (interface{}, error) {
			if len(args) >= 3 {
				l, lok := tryFloat(args[0])
				r, rok := tryFloat(args[2])
				if lok && rok {
					return l != r, nil
				}
				return strings.Trim(fmt.Sprintf("%v", args[0]), `"'`) != strings.Trim(fmt.Sprintf("%v", args[2]), `"'`), nil
			}
			return false, nil
		}, true
	case "lt":
		return func(args []interface{}) (interface{}, error) {
			return len(args) >= 3 && toFloat(args[0]) < toFloat(args[2]), nil
		}, true
	case "lte", "le":
		return func(args []interface{}) (interface{}, error) {
			return len(args) >= 3 && toFloat(args[0]) <= toFloat(args[2]), nil
		}, true
	case "gt":
		return func(args []interface{}) (interface{}, error) {
			return len(args) >= 3 && toFloat(args[0]) > toFloat(args[2]), nil
		}, true
	case "gte", "ge":
		return func(args []interface{}) (interface{}, error) {
			return len(args) >= 3 && toFloat(args[0]) >= toFloat(args[2]), nil
		}, true
	case "and":
		return func(args []interface{}) (interface{}, error) {
			return len(args) >= 3 && toBool(args[0]) && toBool(args[2]), nil
		}, true
	case "or":
		return func(args []interface{}) (interface{}, error) {
			return len(args) >= 3 && (toBool(args[0]) || toBool(args[2])), nil
		}, true
	case "not":
		return func(args []interface{}) (interface{}, error) {
			if len(args) >= 2 {
				return !toBool(args[1]), nil
			}
			return !toBool(args[0]), nil
		}, true
	case "ternary", "cond", "cond_op":
		return func(args []interface{}) (interface{}, error) {
			// [cond, "?", thenVal, ":", elseVal]
			if len(args) >= 5 {
				if toBool(args[0]) {
					return args[2], nil
				}
				return args[4], nil
			}
			return args, nil
		}, true
	}
	return nil, false
}

func tryFloat(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(n), 64)
		return f, err == nil
	default:
		return 0, false
	}
}

func toFloat(v interface{}) float64 {
	f, _ := tryFloat(v)
	return f
}

func parseNumOrPass(v interface{}) interface{} {
	if s, ok := v.(string); ok {
		if i, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
			return i
		}
		if f, err := strconv.ParseFloat(strings.TrimSpace(s), 64); err == nil {
			return f
		}
	}
	return v
}

func cleanNumber(f float64) interface{} {
	if f == math.Trunc(f) && math.Abs(f) < 1e15 {
		return int(f)
	}
	return f
}

func toBool(v interface{}) bool {
	switch b := v.(type) {
	case bool:
		return b
	case int:
		return b != 0
	case float64:
		return b != 0
	case string:
		s := strings.ToLower(strings.TrimSpace(b))
		return s == "true" || s == "1" || s == "yes"
	default:
		return v != nil
	}
}

// ValidateGrammar validates the currently loaded grammar.
func (e *Engine) ValidateGrammar() string {
	e.mu.RLock()
	d := e.dsl
	e.mu.RUnlock()
	if d == nil {
		return encode(Response{OK: false, Error: "no grammar loaded"})
	}
	warnings, err := d.Validate()
	cfg := d.ToConfig()
	if err != nil {
		return encode(Response{OK: false, Error: err.Error(), Warnings: warnings, Config: &cfg})
	}
	return encode(Response{OK: true, Warnings: warnings, Config: &cfg})
}

// Parse parses and evaluates code with a default gas quota (10,000 steps if maxSteps <= 0)
// and returns the AST, evaluation output, tokens, and execution trace.
func (e *Engine) Parse(code string, maxSteps int) string {
	e.mu.RLock()
	d := e.dsl
	e.mu.RUnlock()
	if d == nil {
		return encode(Response{OK: false, Error: "no grammar loaded"})
	}
	if maxSteps <= 0 {
		maxSteps = 10000
	}

	toks, trivia, _ := d.TokenizeWithTrivia(code)
	res, trace, err := d.ParseTrace(code, dslbuilder.EvalOptions{
		MaxSteps: maxSteps,
		Trace:    true,
	})
	if err != nil {
		diags := e.collectDiags(d, code)
		repair, _ := d.RepairPrompt(code)
		return encode(Response{
			OK:           false,
			Error:        dslbuilder.GetDetailedError(err),
			Trace:        trace,
			Tokens:       toks,
			Trivia:       trivia,
			Diagnostics:  diags,
			RepairPrompt: repair,
		})
	}

	astNode := res.GetAST()
	pretty := ""
	if astNode != nil {
		pretty = astNode.Pretty()
	}
	return encode(Response{
		OK:        true,
		AST:       toASTView(astNode),
		ASTPretty: pretty,
		Output:    res.GetOutput(),
		Trace:     trace,
		Tokens:    toks,
		Trivia:    trivia,
	})
}

// Diagnostics returns all lexical and syntax errors in code.
func (e *Engine) Diagnostics(code string) string {
	e.mu.RLock()
	d := e.dsl
	e.mu.RUnlock()
	if d == nil {
		return encode(Response{OK: false, Error: "no grammar loaded"})
	}
	return encode(Response{
		OK:          true,
		Diagnostics: e.collectDiags(d, code),
	})
}

func (e *Engine) collectDiags(d *dslbuilder.DSL, code string) []DiagnosticItem {
	raw := d.Diagnostics(code)
	items := make([]DiagnosticItem, 0, len(raw))
	for _, pe := range raw {
		items = append(items, DiagnosticItem{
			Message:  pe.Message,
			Line:     pe.Line,
			Column:   pe.Column,
			Position: pe.Position,
			Token:    pe.Token,
			Detailed: pe.DetailedError(),
		})
	}
	return items
}

// Completions returns valid token/keyword suggestions at byte offset.
func (e *Engine) Completions(code string, offset int) string {
	e.mu.RLock()
	d := e.dsl
	e.mu.RUnlock()
	if d == nil {
		return encode(Response{OK: false, Error: "no grammar loaded"})
	}
	return encode(Response{
		OK:          true,
		Completions: d.Completions(code, offset),
	})
}

// Format formats DSL source code canonically.
func (e *Engine) Format(code string) string {
	e.mu.RLock()
	d := e.dsl
	e.mu.RUnlock()
	if d == nil {
		return encode(Response{OK: false, Error: "no grammar loaded"})
	}
	formatted, err := d.Format(code)
	if err != nil {
		return encode(Response{OK: false, Error: dslbuilder.GetDetailedError(err)})
	}
	return encode(Response{OK: true, Formatted: formatted})
}

// ToGBNF exports the active grammar in GBNF format.
func (e *Engine) ToGBNF() string {
	e.mu.RLock()
	d := e.dsl
	e.mu.RUnlock()
	if d == nil {
		return encode(Response{OK: false, Error: "no grammar loaded"})
	}
	return encode(Response{OK: true, GBNF: d.ToGBNF()})
}

// AIPromptSpec exports an LLM system prompt specification for the active grammar.
func (e *Engine) AIPromptSpec() string {
	e.mu.RLock()
	d := e.dsl
	e.mu.RUnlock()
	if d == nil {
		return encode(Response{OK: false, Error: "no grammar loaded"})
	}
	return encode(Response{OK: true, PromptSpec: d.AIPromptSpec()})
}

// RepairPrompt inspects code and returns a targeted LLM repair prompt if errors exist.
func (e *Engine) RepairPrompt(code string) string {
	e.mu.RLock()
	d := e.dsl
	e.mu.RUnlock()
	if d == nil {
		return encode(Response{OK: false, Error: "no grammar loaded"})
	}
	prompt, clean := d.RepairPrompt(code)
	return encode(Response{OK: true, RepairPrompt: prompt, Clean: clean})
}

// GenerateSamples synthesizes valid sample inputs from the active grammar.
func (e *Engine) GenerateSamples(seed int64, count int) string {
	e.mu.RLock()
	d := e.dsl
	e.mu.RUnlock()
	if d == nil {
		return encode(Response{OK: false, Error: "no grammar loaded"})
	}
	if count <= 0 {
		count = 5
	}
	return encode(Response{OK: true, Samples: d.GenerateSamples(seed, count, 4)})
}

// RunTests executes the declarative tests embedded in the loaded grammar config.
func (e *Engine) RunTests() string {
	e.mu.RLock()
	d := e.dsl
	e.mu.RUnlock()
	if d == nil {
		return encode(Response{OK: false, Error: "no grammar loaded"})
	}
	rep := d.RunConfigTests()
	return encode(Response{OK: rep.Failed == 0, TestReport: &rep})
}

// Fuzz runs deterministic grammar mutation fuzzing and reports panics or hangs.
func (e *Engine) Fuzz(seed int64, iterations int) string {
	e.mu.RLock()
	d := e.dsl
	e.mu.RUnlock()
	if d == nil {
		return encode(Response{OK: false, Error: "no grammar loaded"})
	}
	if iterations <= 0 {
		iterations = 200
	}
	rep := d.FuzzGrammar(seed, iterations)
	return encode(Response{OK: len(rep.Failures) == 0, FuzzReport: &rep})
}
