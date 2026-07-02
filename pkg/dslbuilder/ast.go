// Package dslbuilder - AST construction and evaluation.
//
// This file implements the two-phase engine used by DSL.Parse:
//
//  1. ParseAST builds a real syntax tree (*Node) without running any
//     semantic action. Backtracking, memoization, and rejected alternatives
//     can therefore never trigger side effects (HTTP calls, file writes,
//     context mutation, ...).
//  2. Eval walks the final tree exactly once, bottom-up, executing actions.
//
// The parser supports direct left recursion via the growing seed algorithm
// and indirect left recursion via a generalized growing algorithm
// (parseCycleLR). Validate still flags indirect cycles as a warning because
// restructured grammars — or Expression(), backed by a Pratt parser — are
// usually clearer and faster.
package dslbuilder

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Span identifies a byte range [Start, End) in the original input.
type Span struct {
	Start int // Byte offset where the node begins (0-based)
	End   int // Byte offset just past the node (exclusive)
}

// Node is a node of the parse tree produced by ParseAST.
//
// There are two kinds of nodes:
//   - Token leaves: Token != nil, no children. Eval returns the token text.
//   - Rule nodes: Rule and Action set, children are the matched symbols.
//     Eval evaluates children and passes them to the action.
type Node struct {
	Rule     string      // Rule name ("" for token leaves)
	Action   string      // Action name to run on evaluation ("" = none)
	Children []*Node     // Matched symbols, in sequence order
	Token    *TokenMatch // Non-nil for token leaves
	Span     Span        // Position of this node in the input

	// group marks parenthesized expression nodes created by
	// ExpressionBuilder.Group without an action: evaluation passes
	// through the inner expression value.
	group bool
}

// IsToken reports whether the node is a token leaf.
func (n *Node) IsToken() bool { return n.Token != nil }

// Child returns the i-th child, or nil if out of range.
func (n *Node) Child(i int) *Node {
	if i < 0 || i >= len(n.Children) {
		return nil
	}
	return n.Children[i]
}

// Text returns the token text for leaves, or the concatenation of the
// children's text for rule nodes.
func (n *Node) Text() string {
	if n == nil {
		return ""
	}
	if n.Token != nil {
		return n.Token.Value
	}
	parts := make([]string, 0, len(n.Children))
	for _, c := range n.Children {
		parts = append(parts, c.Text())
	}
	return strings.Join(parts, " ")
}

// Pretty returns an indented, human-readable representation of the tree.
// Useful for debugging grammars and inspecting what the parser matched.
//
// Example output:
//
//	expr (add)
//	├─ expr (number)
//	│  └─ NUMBER "1"
//	└─ PLUS "+"
//	...
func (n *Node) Pretty() string {
	var sb strings.Builder
	n.pretty(&sb, "", true, true)
	return sb.String()
}

func (n *Node) pretty(sb *strings.Builder, prefix string, isLast, isRoot bool) {
	if !isRoot {
		if isLast {
			sb.WriteString(prefix + "└─ ")
		} else {
			sb.WriteString(prefix + "├─ ")
		}
	}

	if n.Token != nil {
		fmt.Fprintf(sb, "%s %q\n", n.Token.TokenType, n.Token.Value)
		return
	}

	label := n.Rule
	if label == "" {
		label = "<seq>"
	}
	if n.Action != "" {
		fmt.Fprintf(sb, "%s (%s)\n", label, n.Action)
	} else {
		fmt.Fprintf(sb, "%s\n", label)
	}

	childPrefix := prefix
	if !isRoot {
		if isLast {
			childPrefix += "   "
		} else {
			childPrefix += "│  "
		}
	}
	for i, c := range n.Children {
		c.pretty(sb, childPrefix, i == len(n.Children)-1, false)
	}
}

// ParseAST parses code into a syntax tree without executing any action.
// This is phase 1 of the two-phase engine. Combine with Eval to execute:
//
//	node, err := dsl.ParseAST("2 + 3 * 4")
//	if err != nil { ... }
//	fmt.Println(node.Pretty())          // inspect the tree
//	value, err := dsl.Eval(node)        // run actions exactly once
func (d *DSL) ParseAST(code string) (*Node, error) {
	parser := newASTParser(d.grammar)
	return parser.Parse(code)
}

// EvalContext is passed to node actions (see NodeAction). It gives the
// action explicit control over child evaluation, enabling lazy semantics
// such as short-circuiting and conditional branches.
type EvalContext struct {
	dsl   *DSL
	depth int
}

// Eval evaluates a node on demand. Nil nodes evaluate to nil.
func (c *EvalContext) Eval(node *Node) (interface{}, error) {
	return c.dsl.evalNode(node, c.depth)
}

// EvalChildren evaluates every child of a node, in order.
func (c *EvalContext) EvalChildren(node *Node) ([]interface{}, error) {
	args := make([]interface{}, len(node.Children))
	for i, child := range node.Children {
		value, err := c.dsl.evalNode(child, c.depth)
		if err != nil {
			return nil, err
		}
		args[i] = value
	}
	return args, nil
}

// NodeActionFunc is an action that receives the unevaluated AST node
// instead of pre-evaluated arguments. The action decides which children
// to evaluate (and when) via the EvalContext.
type NodeActionFunc func(ctx *EvalContext, node *Node) (interface{}, error)

// NodeAction registers a lazy action: unlike Action, the children of the
// matched rule are NOT evaluated beforehand. This is the right tool for
// control flow, where a branch must not execute unless taken:
//
//	dsl.NodeAction("ifElse", func(ctx *dslbuilder.EvalContext, n *dslbuilder.Node) (interface{}, error) {
//	    cond, err := ctx.Eval(n.Child(1))
//	    if err != nil { return nil, err }
//	    if cond.(bool) {
//	        return ctx.Eval(n.Child(3)) // then-branch only
//	    }
//	    return ctx.Eval(n.Child(5))     // else-branch only
//	})
//
// A NodeAction with a given name takes precedence over an Action with
// the same name.
func (d *DSL) NodeAction(name string, fn NodeActionFunc) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.actionsFrozen {
		d.deferredErrors = append(d.deferredErrors,
			fmt.Errorf("actions are frozen (Build() was called); cannot register node action %s — register it before Build, or use BuildAllowLateActions and CompiledDSL.NodeAction", name))
		return
	}
	if d.nodeActions == nil {
		d.nodeActions = make(map[string]NodeActionFunc)
	}
	d.nodeActions[name] = fn
}

// lookupAction and lookupNodeAction read the action maps under the read
// lock so late registration (BuildAllowLateActions) is race-free.
func (d *DSL) lookupAction(name string) (ActionFunc, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	fn, ok := d.actions[name]
	return fn, ok
}

func (d *DSL) lookupNodeAction(name string) (NodeActionFunc, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	fn, ok := d.nodeActions[name]
	return fn, ok
}

// Eval evaluates a parse tree produced by ParseAST, executing the
// registered actions bottom-up. Token leaves evaluate to their text;
// rule nodes evaluate their children and pass the results to the
// rule's action. Rules without an action evaluate to the slice of
// child values (matching the historical parser behavior).
//
// Rules whose action is registered with NodeAction receive the raw node
// and control child evaluation themselves (lazy evaluation).
//
// Evaluation depth is bounded (maxEvalDepth) so pathological trees return
// an error instead of overflowing the stack.
func (d *DSL) Eval(node *Node) (interface{}, error) {
	return d.evalNode(node, 0)
}

func (d *DSL) evalNode(node *Node, depth int) (interface{}, error) {
	if node == nil {
		return nil, nil
	}
	if depth >= maxEvalDepth {
		return nil, fmt.Errorf("maximum evaluation depth exceeded (%d): tree is nested too deeply", maxEvalDepth)
	}

	if node.Token != nil {
		return node.Token.Value, nil
	}

	// Lazy node actions decide what to evaluate.
	if node.Action != "" {
		if fn, exists := d.lookupNodeAction(node.Action); exists {
			return fn(&EvalContext{dsl: d, depth: depth + 1}, node)
		}
	}

	args := make([]interface{}, len(node.Children))
	for i, child := range node.Children {
		value, err := d.evalNode(child, depth+1)
		if err != nil {
			return nil, err
		}
		args[i] = value
	}

	if node.Action != "" {
		if action, exists := d.lookupAction(node.Action); exists {
			return action(args)
		}
	}

	// Parenthesized group without action: pass through the inner value.
	if node.group && len(args) == 3 {
		return args[1], nil
	}

	return args, nil
}

// parseFailure records the farthest point reached by the parser before
// failing, together with what it expected there. Reporting the farthest
// failure (instead of the last local one) gives far more useful errors
// for complex grammars.
type parseFailure struct {
	tokenPos  int      // Token index of the failure
	expected  []string // Symbols that would have allowed progress
	ruleStack []string // Rule nesting at the farthest failure
	hint      string   // Custom error hint from the failing alternative (RuleWithError)
}

// astParser builds parse trees with memoization (Packrat) and direct
// left recursion support. It never executes semantic actions.
type astParser struct {
	grammar   *Grammar
	tokens    []TokenMatch
	pos       int
	memo      map[string]map[int]astMemoEntry
	input     string
	growing   map[string]bool
	ruleStack []string
	hintStack []string // active RuleWithError hints, innermost last
	failure   parseFailure
	depth     int // current recursion depth (bounded by maxParseDepth)

	// Indirect left recursion support (generalized growing, see
	// parseCycleLR): rules involved in multi-rule leftmost cycles, the
	// active seeds keyed by "rule@pos", and the number of active heads.
	// While a head is active the memo table is disabled, because entries
	// computed under a temporary seed would be unsound afterwards.
	leftCycleRules map[string]bool
	lrSeeds        map[string]lrSeed
	lrHeads        int
}

// lrSeed is the current best parse of a cycle rule at a position while its
// growth is in progress. A nil node means "no seed yet" (re-entries fail).
type lrSeed struct {
	node *Node
	end  int
}

// astMemoEntry caches the parse result of a rule at a token position.
type astMemoEntry struct {
	node   *Node
	endPos int
	err    error
}

func newASTParser(grammar *Grammar) *astParser {
	return &astParser{
		grammar:        grammar,
		memo:           make(map[string]map[int]astMemoEntry),
		growing:        make(map[string]bool),
		failure:        parseFailure{tokenPos: -1},
		leftCycleRules: leftCycleRuleSet(grammar),
		lrSeeds:        make(map[string]lrSeed),
	}
}

// leftCycleRuleSet returns the rules that participate in a multi-rule
// leftmost cycle (indirect left recursion). Pure self-loops (direct left
// recursion) are excluded — they use the dedicated growing-seed path.
func leftCycleRuleSet(g *Grammar) map[string]bool {
	// Leftmost graph between rules, excluding self-loops.
	edges := make(map[string][]string)
	for name, rule := range g.rules {
		seen := make(map[string]bool)
		for _, alt := range rule.alternatives {
			if len(alt.sequence) == 0 {
				continue
			}
			first := alt.sequence[0]
			if first == name {
				continue
			}
			if _, isRule := g.rules[first]; isRule && !seen[first] {
				edges[name] = append(edges[name], first)
				seen[first] = true
			}
		}
	}

	members := make(map[string]bool)
	const (
		white = 0
		gray  = 1
		black = 2
	)
	color := make(map[string]int)
	var stack []string

	var dfs func(name string)
	dfs = func(name string) {
		color[name] = gray
		stack = append(stack, name)
		for _, next := range edges[name] {
			if color[next] == gray {
				for i := len(stack) - 1; i >= 0; i-- {
					members[stack[i]] = true
					if stack[i] == next {
						break
					}
				}
			} else if color[next] == white {
				dfs(next)
			}
		}
		stack = stack[:len(stack)-1]
		color[name] = black
	}
	for name := range g.rules {
		if color[name] == white {
			dfs(name)
		}
	}
	return members
}

// Parse tokenizes the input and parses it from the start rule,
// requiring all tokens to be consumed.
func (p *astParser) Parse(code string) (*Node, error) {
	p.tokens = nil
	p.pos = 0
	p.memo = make(map[string]map[int]astMemoEntry)
	p.input = code
	p.growing = make(map[string]bool)
	p.ruleStack = nil
	p.failure = parseFailure{tokenPos: -1}

	tokens, err := tokenizeInput(p.grammar, code, code)
	if err != nil {
		return nil, err
	}
	p.tokens = tokens

	startRule := p.grammar.startRule
	if startRule == "" {
		return nil, fmt.Errorf("start rule not found (grammar has no rules)")
	}

	node, err := p.parseRule(startRule)
	if err != nil {
		// Grammar-level problems (e.g. missing rules) are not syntax errors;
		// report them as-is instead of a farthest-failure message.
		if err != errNoMatch {
			return nil, err
		}
		return nil, p.farthestError(startRule)
	}

	if p.pos < len(p.tokens) {
		tok := p.tokens[p.pos]
		message := fmt.Sprintf("unexpected token: %s", tok.Value)
		return nil, createParseError(message, tok.Start, tok.Value, p.input)
	}

	return node, nil
}

// recordFailure updates the farthest-failure information. Failures at
// positions before the current farthest point are ignored; failures at
// the same point accumulate their expected symbols.
func (p *astParser) recordFailure(expected string) {
	if p.pos < p.failure.tokenPos {
		return
	}
	if p.pos > p.failure.tokenPos {
		p.failure.tokenPos = p.pos
		p.failure.expected = nil
		p.failure.ruleStack = append([]string(nil), p.ruleStack...)
		p.failure.hint = ""
		if len(p.hintStack) > 0 {
			p.failure.hint = p.hintStack[len(p.hintStack)-1]
		}
	}
	for _, e := range p.failure.expected {
		if e == expected {
			return
		}
	}
	p.failure.expected = append(p.failure.expected, expected)
}

// farthestError builds the final ParseError from the farthest failure.
// The message keeps the historical "no alternative matched for rule X"
// prefix for backward compatibility, then adds expected/got details and
// the rule stack:
//
//	no alternative matched for rule condition: expected NUMBER or IDENT, got THEN
//	rule stack: program > statement > if_stmt > condition
func (p *astParser) farthestError(startRule string) error {
	var got, tokenValue string
	var position int

	failPos := p.failure.tokenPos
	if failPos < 0 {
		failPos = p.pos
	}

	if failPos < len(p.tokens) {
		tok := p.tokens[failPos]
		got = fmt.Sprintf("%s %q", tok.TokenType, tok.Value)
		tokenValue = tok.Value
		position = tok.Start
	} else {
		got = "<end of input>"
		tokenValue = "<end of input>"
		position = len(p.input)
	}

	failedRule := startRule
	if len(p.failure.ruleStack) > 0 {
		failedRule = p.failure.ruleStack[len(p.failure.ruleStack)-1]
	}

	message := fmt.Sprintf("no alternative matched for rule %s", failedRule)
	if len(p.failure.expected) > 0 {
		expected := append([]string(nil), p.failure.expected...)
		sort.Strings(expected)
		message += fmt.Sprintf(": expected %s, got %s", joinExpected(expected), got)
	}
	if p.failure.hint != "" {
		message += "\nhint: " + p.failure.hint
	}
	if len(p.failure.ruleStack) > 1 {
		message += "\nrule stack: " + strings.Join(p.failure.ruleStack, " > ")
	}

	return createParseError(message, position, tokenValue, p.input)
}

func joinExpected(expected []string) string {
	switch len(expected) {
	case 0:
		return ""
	case 1:
		return expected[0]
	default:
		return strings.Join(expected[:len(expected)-1], ", ") + " or " + expected[len(expected)-1]
	}
}

// errNoMatch is a sentinel used internally; the detailed error is always
// reconstructed from the farthest failure in Parse.
var errNoMatch = fmt.Errorf("no match")

// maxParseDepth bounds parser recursion (rule nesting) and maxEvalDepth
// bounds evaluator recursion (node nesting). Both are far above anything a
// real grammar produces; they exist to turn pathological inputs (e.g.
// thousands of nested parentheses) into a clean error instead of a stack
// overflow.
const (
	maxParseDepth = 10000
	maxEvalDepth  = 10000
)

// errParseDepthExceeded aborts parsing immediately: unlike errNoMatch it is
// not swallowed by alternative backtracking.
var errParseDepthExceeded = fmt.Errorf("maximum parse depth exceeded (%d): input is nested too deeply", maxParseDepth)

// parseRule parses a rule (or expression rule) with memoization.
func (p *astParser) parseRule(ruleName string) (*Node, error) {
	if p.depth >= maxParseDepth {
		return nil, errParseDepthExceeded
	}
	p.depth++
	defer func() { p.depth-- }()

	// Expression rules use the Pratt parser (see expression.go).
	if spec, ok := p.grammar.exprRules[ruleName]; ok {
		return p.parseExpressionRule(spec)
	}

	// Rules in a multi-rule leftmost cycle (indirect left recursion) use
	// the generalized growing algorithm and bypass the memo table.
	if p.leftCycleRules[ruleName] {
		p.ruleStack = append(p.ruleStack, ruleName)
		node, err := p.parseCycleLR(ruleName)
		p.ruleStack = p.ruleStack[:len(p.ruleStack)-1]
		return node, err
	}

	// The memo table is disabled while a cycle head is growing: results
	// computed under a temporary seed would be unsound once it changes.
	useMemo := p.lrHeads == 0

	if useMemo {
		if ruleMemo, exists := p.memo[ruleName]; exists {
			if entry, exists := ruleMemo[p.pos]; exists {
				p.pos = entry.endPos
				return entry.node, entry.err
			}
		} else {
			p.memo[ruleName] = make(map[int]astMemoEntry)
		}
	}

	startPos := p.pos
	p.ruleStack = append(p.ruleStack, ruleName)

	var node *Node
	var err error
	if p.isLeftRecursive(ruleName) {
		node, err = p.parseLeftRecursive(ruleName)
	} else {
		node, err = p.parseRuleRegular(ruleName)
	}

	p.ruleStack = p.ruleStack[:len(p.ruleStack)-1]
	if useMemo {
		// A nested cycle head may have replaced the memo table while this
		// rule was parsing; re-ensure the submap before writing. The result
		// itself is sound: any nested growth completed (fixpoint reached)
		// before this rule returned.
		if p.memo[ruleName] == nil {
			p.memo[ruleName] = make(map[int]astMemoEntry)
		}
		p.memo[ruleName][startPos] = astMemoEntry{node: node, endPos: p.pos, err: err}
	}
	return node, err
}

// parseCycleLR parses a rule involved in indirect left recursion using a
// generalized growing algorithm (a simplified form of Warth et al.):
//
//  1. The first parse attempt runs with an empty seed — re-entries into
//     this rule at this position fail, so only non-recursive derivations
//     succeed. That result becomes the seed (PEG ordered choice applies).
//  2. Growth rounds re-parse the rule with the seed installed: re-entries
//     return the seed, letting recursive alternatives extend it. Each
//     round keeps the LONGEST successful alternative; growth stops when
//     the match no longer lengthens.
//
// While any head is active the memo table is disabled (see parseRule) and
// it is cleared on entry/exit of the outermost head, so no seed-dependent
// result survives.
func (p *astParser) parseCycleLR(ruleName string) (*Node, error) {
	startPos := p.pos
	key := ruleName + "@" + strconv.Itoa(startPos)

	// Re-entry during an active growth: return the current seed.
	if seed, ok := p.lrSeeds[key]; ok {
		if seed.node == nil {
			return nil, errNoMatch
		}
		p.pos = seed.end
		return seed.node, nil
	}

	// This call becomes a head.
	if p.lrHeads == 0 {
		p.memo = make(map[string]map[int]astMemoEntry)
	}
	p.lrHeads++
	p.lrSeeds[key] = lrSeed{node: nil, end: startPos}
	defer func() {
		delete(p.lrSeeds, key)
		p.lrHeads--
		if p.lrHeads == 0 {
			p.memo = make(map[string]map[int]astMemoEntry)
		}
	}()

	var bestNode *Node
	bestPos := startPos
	found := false

	for {
		p.pos = startPos

		var node *Node
		var err error
		if !found {
			// Seed round: ordered choice, recursive re-entries fail.
			node, err = p.parseRuleRegular(ruleName)
		} else {
			// Growth round: longest alternative wins so the seed extends.
			node, err = p.parseRuleLongest(ruleName)
		}
		if err != nil {
			break
		}
		if !found || p.pos > bestPos {
			bestNode = node
			bestPos = p.pos
			found = true
			p.lrSeeds[key] = lrSeed{node: bestNode, end: bestPos}
			continue
		}
		break
	}

	if !found {
		p.pos = startPos
		return nil, errNoMatch
	}
	p.pos = bestPos
	return bestNode, nil
}

// parseRuleLongest tries every alternative from the same start position and
// keeps the longest successful match (first declared wins ties). Used for
// the growth rounds of parseCycleLR.
func (p *astParser) parseRuleLongest(ruleName string) (*Node, error) {
	rule, exists := p.grammar.rules[ruleName]
	if !exists {
		return nil, fmt.Errorf("rule %s not found", ruleName)
	}

	startPos := p.pos
	var bestNode *Node
	bestPos := startPos
	found := false

	for _, alt := range rule.alternatives {
		p.pos = startPos
		node, err := p.parseAlternative(ruleName, alt)
		if err == errParseDepthExceeded {
			return nil, err
		}
		if err == nil && (!found || p.pos > bestPos) {
			bestNode = node
			bestPos = p.pos
			found = true
		}
	}

	if !found {
		p.pos = startPos
		return nil, errNoMatch
	}
	p.pos = bestPos
	return bestNode, nil
}

// isLeftRecursive reports whether a rule has a directly left-recursive
// alternative (rule → rule ...). Only DIRECT left recursion is detected
// here; indirect left recursion is handled earlier in parseRule via
// leftCycleRules/parseCycleLR.
func (p *astParser) isLeftRecursive(ruleName string) bool {
	rule, exists := p.grammar.rules[ruleName]
	if !exists {
		return false
	}
	for _, alt := range rule.alternatives {
		if len(alt.sequence) > 0 && alt.sequence[0] == ruleName {
			return true
		}
	}
	return false
}

// parseRuleRegular tries each alternative in order (PEG-style ordered
// choice) and returns the first that matches.
func (p *astParser) parseRuleRegular(ruleName string) (*Node, error) {
	rule, exists := p.grammar.rules[ruleName]
	if !exists {
		return nil, fmt.Errorf("rule %s not found", ruleName)
	}

	for _, alt := range rule.alternatives {
		savedPos := p.pos
		node, err := p.parseAlternative(ruleName, alt)
		if err == nil {
			return node, nil
		}
		if err == errParseDepthExceeded {
			return nil, err // abort: not a normal alternative failure
		}
		p.pos = savedPos
	}

	return nil, errNoMatch
}

// parseAlternative matches one alternative's symbol sequence and builds
// the corresponding rule node. No action is executed here.
func (p *astParser) parseAlternative(ruleName string, alt *Alternative) (*Node, error) {
	if alt.errHint != "" {
		p.hintStack = append(p.hintStack, alt.errHint)
		defer func() { p.hintStack = p.hintStack[:len(p.hintStack)-1] }()
	}

	children := make([]*Node, 0, len(alt.sequence))

	for _, symbol := range alt.sequence {
		child, err := p.parseSymbol(symbol)
		if err != nil {
			return nil, err
		}
		children = append(children, child)
	}

	return p.makeRuleNode(ruleName, alt.action, children), nil
}

// parseSymbol matches a single symbol: a token or a nested rule.
func (p *astParser) parseSymbol(symbol string) (*Node, error) {
	if _, isToken := p.grammar.tokens[symbol]; isToken {
		if p.pos >= len(p.tokens) {
			p.recordFailure(symbol)
			return nil, errNoMatch
		}
		if p.tokens[p.pos].TokenType != symbol {
			p.recordFailure(symbol)
			return nil, errNoMatch
		}
		tok := p.tokens[p.pos]
		p.pos++
		return tokenNode(tok), nil
	}

	if !p.grammar.isSymbolDefined(symbol) {
		return nil, fmt.Errorf("rule %s not found", symbol)
	}
	return p.parseRule(symbol)
}

// parseLeftRecursive implements the growing seed algorithm for direct
// left recursion, building nodes instead of executing actions:
//
//  1. Parse a seed using the non-recursive alternatives.
//  2. Repeatedly try to extend the seed with the recursive alternatives,
//     using the current seed as the leftmost child.
//  3. Stop when no alternative extends the match.
//
// This yields naturally left-associative trees.
func (p *astParser) parseLeftRecursive(ruleName string) (*Node, error) {
	rule := p.grammar.rules[ruleName]
	startPos := p.pos

	growKey := fmt.Sprintf("%s_%d", ruleName, startPos)
	if p.growing[growKey] {
		return nil, errNoMatch
	}
	p.growing[growKey] = true
	defer delete(p.growing, growKey)

	// Step 1: seed from non-recursive alternatives.
	var seed *Node
	seedPos := startPos
	foundSeed := false

	for _, alt := range rule.alternatives {
		if len(alt.sequence) > 0 && alt.sequence[0] == ruleName {
			continue
		}
		p.pos = startPos
		node, err := p.parseAlternative(ruleName, alt)
		if err == nil {
			seed = node
			seedPos = p.pos
			foundSeed = true
			break
		}
	}

	if !foundSeed {
		p.pos = startPos
		return nil, errNoMatch
	}

	// Step 2: grow the seed.
	for {
		improved := false
		bestNode := seed
		bestPos := seedPos

		for _, alt := range rule.alternatives {
			if len(alt.sequence) == 0 || alt.sequence[0] != ruleName {
				continue
			}

			children := []*Node{seed}
			p.pos = seedPos
			success := true

			for i := 1; i < len(alt.sequence); i++ {
				child, err := p.parseSymbol(alt.sequence[i])
				if err != nil {
					success = false
					break
				}
				children = append(children, child)
			}

			if success && p.pos > bestPos {
				bestNode = p.makeRuleNode(ruleName, alt.action, children)
				bestPos = p.pos
				improved = true
			}
		}

		if !improved {
			p.pos = bestPos
			return bestNode, nil
		}

		seed = bestNode
		seedPos = bestPos
	}
}

// makeRuleNode builds a rule node computing its span from the children.
func (p *astParser) makeRuleNode(ruleName, action string, children []*Node) *Node {
	span := Span{}
	if len(children) > 0 {
		span.Start = children[0].Span.Start
		span.End = children[len(children)-1].Span.End
	} else if p.pos < len(p.tokens) {
		span.Start = p.tokens[p.pos].Start
		span.End = span.Start
	} else {
		span.Start = len(p.input)
		span.End = span.Start
	}

	return &Node{
		Rule:     ruleName,
		Action:   action,
		Children: children,
		Span:     span,
	}
}

// tokenNode builds a leaf node for a matched token.
func tokenNode(tok TokenMatch) *Node {
	t := tok
	return &Node{
		Token: &t,
		Span:  Span{Start: tok.Start, End: tok.End},
	}
}
