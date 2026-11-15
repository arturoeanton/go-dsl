// Package dslbuilder - Execution sandbox (context.Context, step/depth quotas,
// call-scoped variables) and structured execution tracing / explainability.
package dslbuilder

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrStepLimitExceeded is returned when AST evaluation exceeds EvalOptions.MaxSteps.
var ErrStepLimitExceeded = errors.New("evaluation step quota exceeded")

// EvalOptions configures sandboxed and traceable AST evaluation.
type EvalOptions struct {
	Context  context.Context        // Cancellation / deadline propagation
	MaxSteps int                    // Maximum AST nodes evaluated (0 = unlimited)
	MaxDepth int                    // Maximum evaluation recursion depth (0 = default maxEvalDepth)
	Vars     map[string]interface{} // Call-scoped variables isolated to this evaluation
	Trace    bool                   // Capture an EvalTrace tree during evaluation
}

// TraceNode records the evaluation (or lazy skipping) of a single AST node.
type TraceNode struct {
	Step       int           `json:"step,omitempty"`
	Rule       string        `json:"rule,omitempty"`
	Action     string        `json:"action,omitempty"`
	Token      string        `json:"token,omitempty"`
	Span       Span          `json:"span"`
	Executed   bool          `json:"executed"`
	Skipped    bool          `json:"skipped,omitempty"`
	Args       []interface{} `json:"args,omitempty"`
	Output     interface{}   `json:"output,omitempty"`
	Error      string        `json:"error,omitempty"`
	DurationNs int64         `json:"duration_ns,omitempty"`
	Children   []*TraceNode  `json:"children,omitempty"`
}

// EvalTrace captures the full step-by-step execution audit of an AST evaluation,
// including intermediate inputs/outputs, timings, and untaken lazy branches.
type EvalTrace struct {
	Steps      int         `json:"steps"`
	DurationNs int64       `json:"duration_ns"`
	Output     interface{} `json:"output,omitempty"`
	Error      string      `json:"error,omitempty"`
	Root       *TraceNode  `json:"root,omitempty"`
}

// ToJSON serializes the execution trace as indented JSON.
func (t *EvalTrace) ToJSON() ([]byte, error) {
	if t == nil {
		return []byte("null"), nil
	}
	return json.MarshalIndent(t, "", "  ")
}

// Pretty renders a human-readable execution tree highlighting executed steps,
// outputs, timings, and skipped lazy branches.
func (t *EvalTrace) Pretty() string {
	if t == nil || t.Root == nil {
		return "<empty trace>"
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "Trace (%d steps, %s)\n", t.Steps, time.Duration(t.DurationNs))
	t.Root.pretty(&sb, "", true, true)
	return sb.String()
}

func (tn *TraceNode) pretty(sb *strings.Builder, prefix string, isLast, isRoot bool) {
	if tn == nil {
		return
	}
	if !isRoot {
		if isLast {
			sb.WriteString(prefix + "└─ ")
		} else {
			sb.WriteString(prefix + "├─ ")
		}
	}

	if tn.Skipped && !tn.Executed {
		label := tn.Rule
		if label == "" && tn.Token != "" {
			label = fmt.Sprintf("TOKEN %q", tn.Token)
		}
		if tn.Action != "" {
			fmt.Fprintf(sb, "[SKIPPED] %s (%s)\n", label, tn.Action)
		} else {
			fmt.Fprintf(sb, "[SKIPPED] %s\n", label)
		}
		return
	}

	if tn.Token != "" && tn.Rule == "" {
		fmt.Fprintf(sb, "[#%d] %q\n", tn.Step, tn.Token)
		return
	}

	label := tn.Rule
	if label == "" {
		label = "<node>"
	}
	if tn.Action != "" {
		label = fmt.Sprintf("%s (%s)", label, tn.Action)
	}
	if tn.Error != "" {
		fmt.Fprintf(sb, "[#%d] %s => ERROR: %s (%s)\n", tn.Step, label, tn.Error, time.Duration(tn.DurationNs))
	} else {
		fmt.Fprintf(sb, "[#%d] %s => %v (%s)\n", tn.Step, label, tn.Output, time.Duration(tn.DurationNs))
	}

	childPrefix := prefix
	if !isRoot {
		if isLast {
			childPrefix += "   "
		} else {
			childPrefix += "│  "
		}
	}
	for i, ch := range tn.Children {
		ch.pretty(sb, childPrefix, i == len(tn.Children)-1, false)
	}
}

type evalState struct {
	ctx        context.Context
	maxSteps   int
	maxDepth   int
	steps      int
	vars       map[string]interface{}
	trace      bool
	childSlots map[*Node]*TraceNode
}

// Context returns the context.Context governing this evaluation (never nil).
func (c *EvalContext) Context() context.Context {
	if c != nil && c.state != nil && c.state.ctx != nil {
		return c.state.ctx
	}
	return context.Background()
}

// DSL returns the DSL instance running the evaluation.
func (c *EvalContext) DSL() *DSL {
	if c == nil {
		return nil
	}
	return c.dsl
}

// Steps returns the number of AST evaluation steps executed so far.
func (c *EvalContext) Steps() int {
	if c != nil && c.state != nil {
		return c.state.steps
	}
	return 0
}

// Get reads a call-scoped variable from EvalOptions.Vars, falling back to
// DSL.GetContext if not overridden in Vars.
func (c *EvalContext) Get(key string) interface{} {
	if c == nil {
		return nil
	}
	if c.state != nil && c.state.vars != nil {
		if v, ok := c.state.vars[key]; ok {
			return v
		}
	}
	if c.dsl != nil {
		return c.dsl.GetContext(key)
	}
	return nil
}

// Set stores a variable in the call-scoped EvalOptions.Vars map when evaluating
// with options, or in DSL context when no EvalOptions state is active.
func (c *EvalContext) Set(key string, value interface{}) {
	if c == nil {
		return
	}
	if c.state != nil {
		if c.state.vars == nil {
			c.state.vars = make(map[string]interface{})
		}
		c.state.vars[key] = value
		return
	}
	if c.dsl != nil {
		c.dsl.SetContext(key, value)
	}
}

// ContextAction registers an action that receives an *EvalContext alongside
// pre-evaluated child arguments.
func (d *DSL) ContextAction(name string, fn ContextActionFunc) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.actionsFrozen {
		d.deferredErrors = append(d.deferredErrors,
			fmt.Errorf("actions are frozen (Build() was called); cannot register context action %s — register it before Build, or use BuildAllowLateActions and CompiledDSL.ContextAction", name))
		return
	}
	if d.contextActions == nil {
		d.contextActions = make(map[string]ContextActionFunc)
	}
	d.contextActions[name] = fn
}

// WithContextAction registers a ContextAction and returns the DSL for chaining.
func (d *DSL) WithContextAction(name string, fn ContextActionFunc) *DSL {
	d.ContextAction(name, fn)
	return d
}

// EvalWithOptions evaluates an AST node under the given sandbox options.
func (d *DSL) EvalWithOptions(node *Node, opts EvalOptions) (interface{}, error) {
	out, _, err := d.runEval(node, opts)
	return out, err
}

// ParseWithOptions parses and evaluates DSL code under the given sandbox options.
// If opts.Trace is true, Result.Trace contains the execution trace.
func (d *DSL) ParseWithOptions(code string, opts EvalOptions) (*Result, error) {
	if opts.Context != nil {
		if err := opts.Context.Err(); err != nil {
			return nil, err
		}
	}
	node, err := d.ParseAST(code)
	if err != nil {
		if IsParseError(err) {
			return nil, err
		}
		return nil, fmt.Errorf("parsing error: %w", err)
	}

	output, trace, err := d.runEval(node, opts)
	if err != nil {
		return nil, fmt.Errorf("parsing error: %w", err)
	}

	return &Result{
		AST:    node,
		Code:   code,
		Output: output,
		Trace:  trace,
		DSL:    d,
	}, nil
}

// EvalTrace evaluates an AST node and returns a complete structured EvalTrace.
func (d *DSL) EvalTrace(node *Node, opts ...EvalOptions) (*EvalTrace, error) {
	var o EvalOptions
	if len(opts) > 0 {
		o = opts[0]
	}
	o.Trace = true
	_, trace, err := d.runEval(node, o)
	return trace, err
}

// ParseTrace parses and evaluates DSL code, returning both the Result and the
// structured EvalTrace (even if evaluation returns an error, the partial trace
// up to the failure point is returned).
func (d *DSL) ParseTrace(code string, opts ...EvalOptions) (*Result, *EvalTrace, error) {
	var o EvalOptions
	if len(opts) > 0 {
		o = opts[0]
	}
	o.Trace = true
	if o.Context != nil {
		if err := o.Context.Err(); err != nil {
			return nil, nil, err
		}
	}

	node, err := d.ParseAST(code)
	if err != nil {
		if IsParseError(err) {
			return nil, nil, err
		}
		return nil, nil, fmt.Errorf("parsing error: %w", err)
	}

	output, trace, evalErr := d.runEval(node, o)
	if evalErr != nil {
		return nil, trace, fmt.Errorf("parsing error: %w", evalErr)
	}

	res := &Result{
		AST:    node,
		Code:   code,
		Output: output,
		Trace:  trace,
		DSL:    d,
	}
	return res, trace, nil
}

func (d *DSL) runEval(node *Node, opts EvalOptions) (interface{}, *EvalTrace, error) {
	maxDepth := maxEvalDepth
	if opts.MaxDepth > 0 {
		maxDepth = opts.MaxDepth
	}

	var varsCopy map[string]interface{}
	if len(opts.Vars) > 0 {
		varsCopy = make(map[string]interface{}, len(opts.Vars))
		for k, v := range opts.Vars {
			varsCopy[k] = v
		}
	}

	state := &evalState{
		ctx:      opts.Context,
		maxSteps: opts.MaxSteps,
		maxDepth: maxDepth,
		vars:     varsCopy,
		trace:    opts.Trace,
	}

	var rootTrace *TraceNode
	if opts.Trace && node != nil {
		rootTrace = newSkippedTraceNode(node)
	}

	start := time.Now()
	out, err := d.evalNodeWithState(node, 0, state, rootTrace)
	elapsed := time.Since(start)

	if !opts.Trace {
		return out, nil, err
	}

	tr := &EvalTrace{
		Steps:      state.steps,
		DurationNs: elapsed.Nanoseconds(),
		Output:     out,
		Root:       rootTrace,
	}
	if err != nil {
		tr.Error = err.Error()
	}
	return out, tr, err
}

func newSkippedTraceNode(n *Node) *TraceNode {
	if n == nil {
		return nil
	}
	tn := &TraceNode{
		Rule:     n.Rule,
		Action:   n.Action,
		Span:     n.Span,
		Executed: false,
		Skipped:  true,
	}
	if n.Token != nil {
		tn.Token = n.Token.Value
	}
	return tn
}

func (d *DSL) evalNodeWithState(node *Node, depth int, state *evalState, traceSlot *TraceNode) (interface{}, error) {
	if node == nil {
		return nil, nil
	}

	limitDepth := maxEvalDepth
	if state != nil {
		if state.ctx != nil {
			if err := state.ctx.Err(); err != nil {
				return nil, err
			}
		}
		state.steps++
		if state.maxSteps > 0 && state.steps > state.maxSteps {
			return nil, fmt.Errorf("%w: limit of %d steps reached", ErrStepLimitExceeded, state.maxSteps)
		}
		if state.maxDepth > 0 {
			limitDepth = state.maxDepth
		}
	}

	if depth >= limitDepth {
		return nil, fmt.Errorf("maximum evaluation depth exceeded (%d): tree is nested too deeply", limitDepth)
	}

	var stepNum int
	var start time.Time
	if traceSlot != nil && state != nil {
		stepNum = state.steps
		start = time.Now()
		traceSlot.Step = stepNum
		traceSlot.Executed = true
		traceSlot.Skipped = false
	}

	finishTrace := func(args []interface{}, out interface{}, err error) {
		if traceSlot == nil {
			return
		}
		traceSlot.DurationNs = time.Since(start).Nanoseconds()
		if len(args) > 0 {
			traceSlot.Args = args
		}
		if err != nil {
			traceSlot.Error = err.Error()
		} else {
			traceSlot.Output = out
		}
	}

	if node.Token != nil {
		finishTrace(nil, node.Token.Value, nil)
		return node.Token.Value, nil
	}

	// Lazy node actions decide which children to evaluate; pre-populate skipped
	// child trace slots so untaken branches remain visible in the trace!
	if node.Action != "" {
		if fn, exists := d.lookupNodeAction(node.Action); exists {
			var prevSlots map[*Node]*TraceNode
			if traceSlot != nil && state != nil {
				prevSlots = state.childSlots
				slots := make(map[*Node]*TraceNode, len(node.Children))
				for k, v := range prevSlots {
					slots[k] = v
				}
				traceSlot.Children = make([]*TraceNode, len(node.Children))
				for i, ch := range node.Children {
					chTrace := newSkippedTraceNode(ch)
					traceSlot.Children[i] = chTrace
					if ch != nil {
						slots[ch] = chTrace
					}
				}
				state.childSlots = slots
			}
			out, err := fn(&EvalContext{dsl: d, depth: depth + 1, state: state}, node)
			if traceSlot != nil && state != nil {
				state.childSlots = prevSlots
			}
			finishTrace(nil, out, err)
			return out, err
		}
	}

	args := make([]interface{}, len(node.Children))
	if traceSlot != nil {
		traceSlot.Children = make([]*TraceNode, len(node.Children))
		for i, ch := range node.Children {
			traceSlot.Children[i] = newSkippedTraceNode(ch)
		}
	}
	for i, child := range node.Children {
		var childTrace *TraceNode
		if traceSlot != nil {
			childTrace = traceSlot.Children[i]
		}
		value, err := d.evalNodeWithState(child, depth+1, state, childTrace)
		if err != nil {
			finishTrace(args[:i], nil, err)
			return nil, err
		}
		args[i] = value
	}

	actionArgs := args
	if node.callStep > 0 && len(args) >= 3 {
		callSlice := make([]interface{}, 0)
		for i := 2; i < len(args)-1; i += node.callStep {
			callSlice = append(callSlice, args[i])
		}
		actionArgs = []interface{}{args[0], callSlice}
	}

	if node.Action != "" {
		if ctxFn, exists := d.lookupContextAction(node.Action); exists {
			out, err := ctxFn(&EvalContext{dsl: d, depth: depth + 1, state: state}, actionArgs)
			finishTrace(actionArgs, out, err)
			return out, err
		}
		if action, exists := d.lookupAction(node.Action); exists {
			out, err := action(actionArgs)
			finishTrace(actionArgs, out, err)
			return out, err
		}
	}

	if node.callStep > 0 && len(args) >= 3 {
		finishTrace(actionArgs, actionArgs, nil)
		return actionArgs, nil
	}

	// Parenthesized group without action: pass through the inner value.
	if node.group && len(args) == 3 {
		finishTrace(args, args[1], nil)
		return args[1], nil
	}

	finishTrace(args, args, nil)
	return args, nil
}
