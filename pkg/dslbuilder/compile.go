// Package dslbuilder - Freezable builder.
//
// Build() validates the grammar once and freezes it: after Build, attempts
// to add tokens or rules are rejected, so a compiled DSL behaves as an
// immutable artifact that can be handed around safely.
package dslbuilder

import (
	"fmt"
	"sync"
)

// CompiledDSL is a validated, frozen DSL ready for parsing.
//
// Obtain one with DSL.Build():
//
//	calc, err := dslbuilder.New("calc").
//	    WithToken("NUMBER", `[0-9]+`).
//	    WithRule("expr", []string{"NUMBER"}, "number").
//	    WithAction("number", numberAction).
//	    Build()
//	if err != nil { ... }
//	result, err := calc.Parse("42")
//
// Build() freezes both the grammar AND the actions, so the compiled DSL
// behaves as an immutable artifact. If you need to register actions after
// building (late binding), use BuildAllowLateActions and register them
// through CompiledDSL.Action / CompiledDSL.NodeAction, which are
// lock-protected.
//
// Parse/Use/Eval on a CompiledDSL are serialized with an internal mutex,
// so a CompiledDSL may be shared across goroutines. (Full lock-free
// concurrency is not possible because actions receive the parent *DSL
// and may read its context.)
type CompiledDSL struct {
	dsl         *DSL
	warnings    []string
	lateActions bool
	mu          sync.Mutex
}

// Build validates the DSL and freezes its grammar and actions.
//
// It returns an error if validation finds structural problems (unknown
// symbols, no rules, deferred token errors, ...). Validation warnings do
// not fail the build; they are available via Warnings().
//
// After Build:
//   - Token/Rule/Expression additions are rejected
//   - Action/NodeAction registrations are rejected
//
// Register everything before building. If you genuinely need to register
// actions after the build, use BuildAllowLateActions.
func (d *DSL) Build() (*CompiledDSL, error) {
	warnings, err := d.Validate()
	if err != nil {
		return nil, fmt.Errorf("cannot build DSL %s: %w", d.name, err)
	}

	d.grammar.frozen = true
	d.mu.Lock()
	d.actionsFrozen = true
	d.mu.Unlock()
	return &CompiledDSL{dsl: d, warnings: warnings}, nil
}

// BuildAllowLateActions is Build() minus the action freeze: the grammar is
// validated and frozen, but actions may still be registered afterwards via
// CompiledDSL.Action and CompiledDSL.NodeAction (lock-protected).
//
// Use this when actions are wired up after the grammar is loaded — e.g.
// grammars from YAML/JSON where the Go actions are registered by a later
// initialization step.
func (d *DSL) BuildAllowLateActions() (*CompiledDSL, error) {
	warnings, err := d.Validate()
	if err != nil {
		return nil, fmt.Errorf("cannot build DSL %s: %w", d.name, err)
	}

	d.grammar.frozen = true
	return &CompiledDSL{dsl: d, warnings: warnings, lateActions: true}, nil
}

// Action registers an action on a compiled DSL built with
// BuildAllowLateActions. It returns an error on a fully frozen build.
func (c *CompiledDSL) Action(name string, fn ActionFunc) error {
	if !c.lateActions {
		return fmt.Errorf("DSL %s was built with Build(): actions are frozen — register actions before Build, or build with BuildAllowLateActions", c.dsl.name)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.dsl.Action(name, fn)
	return nil
}

// NodeAction registers a lazy node action on a compiled DSL built with
// BuildAllowLateActions. It returns an error on a fully frozen build.
func (c *CompiledDSL) NodeAction(name string, fn NodeActionFunc) error {
	if !c.lateActions {
		return fmt.Errorf("DSL %s was built with Build(): actions are frozen — register actions before Build, or build with BuildAllowLateActions", c.dsl.name)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.dsl.NodeAction(name, fn)
	return nil
}

// Warnings returns the validation warnings collected by Build.
func (c *CompiledDSL) Warnings() []string {
	return c.warnings
}

// Name returns the name of the underlying DSL.
func (c *CompiledDSL) Name() string {
	return c.dsl.name
}

// DSL returns the underlying DSL (e.g. to read context or debug info).
// The grammar stays frozen; with Build() the actions are frozen too.
func (c *CompiledDSL) DSL() *DSL {
	return c.dsl
}

// Parse parses and evaluates code. Safe for concurrent use.
func (c *CompiledDSL) Parse(code string) (*Result, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.dsl.Parse(code)
}

// Use parses and evaluates code with a call-scoped context.
// The context is discarded afterwards. Safe for concurrent use.
func (c *CompiledDSL) Use(code string, ctx map[string]interface{}) (*Result, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.dsl.Use(code, ctx)
}

// ParseAST parses code into a syntax tree without executing actions.
// Safe for concurrent use.
func (c *CompiledDSL) ParseAST(code string) (*Node, error) {
	// ParseAST never touches DSL state (no actions run), so no lock needed.
	return c.dsl.ParseAST(code)
}

// Eval evaluates a tree produced by ParseAST. Safe for concurrent use.
func (c *CompiledDSL) Eval(node *Node) (interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.dsl.Eval(node)
}
