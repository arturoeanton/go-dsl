// Package dslbuilder - Pratt expression parser.
//
// Generic left recursion is a poor fit for operator expressions: encoding
// precedence with rule layers is verbose and fragile. Expression() lets you
// declare operators with explicit binding power and associativity, and the
// parser resolves them with precedence climbing (a Pratt parser), which is
// simple, fast, and predictable.
//
//	dsl.Expression("expr").
//	    Atom("NUMBER", "number").
//	    Group("LPAREN", "expr", "RPAREN").
//	    Prefix("MINUS", 70, "neg").
//	    InfixLeft("PLUS", 10, "add").
//	    InfixLeft("MINUS", 10, "sub").
//	    InfixLeft("STAR", 20, "mul").
//	    InfixLeft("SLASH", 20, "div").
//	    InfixRight("POW", 30, "pow")
//
// Guarantees (with the usual actions):
//
//	1 + 2 * 3    => 7
//	(1 + 2) * 3  => 9
//	10 - 3 - 2   => 5
//	2 ^ 3 ^ 2    => 512
//	-1 * 2       => -2
//
// An expression rule can be referenced from regular rules like any other
// rule name, and its actions receive arguments in the familiar positional
// form: infix → [left, operator, right], prefix → [operator, operand],
// atom → [tokenValue].
package dslbuilder

import "fmt"

// exprOp describes one operator of an expression rule.
type exprOp struct {
	token      string // Operator token name
	power      int    // Binding power (higher binds tighter)
	action     string // Action executed on evaluation
	rightAssoc bool   // Right associativity (for infix)
}

// exprAtom describes a terminal of an expression rule.
type exprAtom struct {
	token  string
	action string
}

// exprGroup describes a grouping construct like ( expr ).
type exprGroup struct {
	open   string // Opening token (e.g. LPAREN)
	inner  string // Rule to parse inside (usually the expression itself)
	close  string // Closing token (e.g. RPAREN)
	action string // Optional action; default passes the inner value through
}

// exprSpec is the full specification of an expression rule.
type exprSpec struct {
	name   string
	atoms  []exprAtom
	groups []exprGroup
	prefix map[string]exprOp
	infix  map[string]exprOp
}

// ExpressionBuilder declares a Pratt-parsed expression rule.
// Obtain one with DSL.Expression(name); all methods chain.
type ExpressionBuilder struct {
	dsl  *DSL
	spec *exprSpec
}

// Expression declares (or extends) an expression rule parsed with a Pratt
// parser. If it is the first rule declared, it becomes the start rule.
//
// The rule can be referenced from regular grammar rules by name.
func (d *DSL) Expression(name string) *ExpressionBuilder {
	if d.grammar.frozen {
		d.deferredErrors = append(d.deferredErrors,
			fmt.Errorf("grammar is frozen (Build() was called); cannot add expression rule %s", name))
		// Return a detached builder so chained calls stay safe no-ops.
		return &ExpressionBuilder{dsl: d, spec: &exprSpec{
			name:   name,
			prefix: map[string]exprOp{},
			infix:  map[string]exprOp{},
		}}
	}

	spec, exists := d.grammar.exprRules[name]
	if !exists {
		spec = &exprSpec{
			name:   name,
			prefix: map[string]exprOp{},
			infix:  map[string]exprOp{},
		}
		d.grammar.exprRules[name] = spec
		if d.grammar.startRule == "" {
			d.grammar.startRule = name
		}
	}
	return &ExpressionBuilder{dsl: d, spec: spec}
}

// Atom declares a terminal of the expression (e.g. a number or identifier).
// The action receives [tokenValue].
func (b *ExpressionBuilder) Atom(token, action string) *ExpressionBuilder {
	b.spec.atoms = append(b.spec.atoms, exprAtom{token: token, action: action})
	return b
}

// Group declares a grouping construct such as ( expr ). Without an action
// the group evaluates to the inner value; with SetGroupAction (or a later
// Group call with the same tokens) you can override that.
func (b *ExpressionBuilder) Group(open, inner, close string) *ExpressionBuilder {
	b.spec.groups = append(b.spec.groups, exprGroup{open: open, inner: inner, close: close})
	return b
}

// Prefix declares a prefix (unary) operator. The action receives
// [operator, operand]. power is the binding power of the operand parse.
func (b *ExpressionBuilder) Prefix(token string, power int, action string) *ExpressionBuilder {
	b.spec.prefix[token] = exprOp{token: token, power: power, action: action}
	return b
}

// InfixLeft declares a left-associative binary operator: a-b-c = (a-b)-c.
// The action receives [left, operator, right].
func (b *ExpressionBuilder) InfixLeft(token string, power int, action string) *ExpressionBuilder {
	b.spec.infix[token] = exprOp{token: token, power: power, action: action}
	return b
}

// InfixRight declares a right-associative binary operator: a^b^c = a^(b^c).
// The action receives [left, operator, right].
func (b *ExpressionBuilder) InfixRight(token string, power int, action string) *ExpressionBuilder {
	b.spec.infix[token] = exprOp{token: token, power: power, action: action, rightAssoc: true}
	return b
}

// parseExpressionRule is the entry point used by the AST parser when a
// rule name resolves to an expression rule.
func (p *astParser) parseExpressionRule(spec *exprSpec) (*Node, error) {
	p.ruleStack = append(p.ruleStack, spec.name)
	node, err := p.parseExprBP(spec, 0)
	p.ruleStack = p.ruleStack[:len(p.ruleStack)-1]
	return node, err
}

// parseExprBP implements precedence climbing. minBP is the minimum binding
// power an infix operator needs to continue extending the left operand.
func (p *astParser) parseExprBP(spec *exprSpec, minBP int) (*Node, error) {
	if p.depth >= maxParseDepth {
		return nil, errParseDepthExceeded
	}
	p.depth++
	defer func() { p.depth-- }()

	left, err := p.parseExprOperand(spec)
	if err != nil {
		return nil, err
	}

	for p.pos < len(p.tokens) {
		tok := p.tokens[p.pos]
		op, ok := spec.infix[tok.TokenType]
		if !ok || op.power < minBP {
			break
		}

		opNode := tokenNode(tok)
		p.pos++

		// Left associativity requires the right side to bind strictly
		// tighter; right associativity allows equal binding power.
		nextMin := op.power + 1
		if op.rightAssoc {
			nextMin = op.power
		}

		right, err := p.parseExprBP(spec, nextMin)
		if err != nil {
			return nil, err
		}

		left = &Node{
			Rule:     spec.name,
			Action:   op.action,
			Children: []*Node{left, opNode, right},
			Span:     Span{Start: left.Span.Start, End: right.Span.End},
		}
	}

	return left, nil
}

// parseExprOperand parses a primary expression: a prefix operator applied
// to an operand, a group, or an atom.
func (p *astParser) parseExprOperand(spec *exprSpec) (*Node, error) {
	if p.pos >= len(p.tokens) {
		p.recordExprExpectations(spec)
		return nil, errNoMatch
	}

	tok := p.tokens[p.pos]

	// Prefix operator
	if op, ok := spec.prefix[tok.TokenType]; ok {
		opNode := tokenNode(tok)
		p.pos++
		operand, err := p.parseExprBP(spec, op.power)
		if err != nil {
			return nil, err
		}
		return &Node{
			Rule:     spec.name,
			Action:   op.action,
			Children: []*Node{opNode, operand},
			Span:     Span{Start: opNode.Span.Start, End: operand.Span.End},
		}, nil
	}

	// Group: open inner close
	for _, g := range spec.groups {
		if tok.TokenType != g.open {
			continue
		}
		openNode := tokenNode(tok)
		p.pos++

		var inner *Node
		var err error
		if g.inner == spec.name {
			inner, err = p.parseExprBP(spec, 0)
		} else {
			inner, err = p.parseSymbol(g.inner)
		}
		if err != nil {
			return nil, err
		}

		if p.pos >= len(p.tokens) || p.tokens[p.pos].TokenType != g.close {
			p.recordFailure(g.close)
			return nil, errNoMatch
		}
		closeNode := tokenNode(p.tokens[p.pos])
		p.pos++

		return &Node{
			Rule:     spec.name,
			Action:   g.action,
			Children: []*Node{openNode, inner, closeNode},
			Span:     Span{Start: openNode.Span.Start, End: closeNode.Span.End},
			group:    true,
		}, nil
	}

	// Atom
	for _, atom := range spec.atoms {
		if tok.TokenType != atom.token {
			continue
		}
		leaf := tokenNode(tok)
		p.pos++
		return &Node{
			Rule:     spec.name,
			Action:   atom.action,
			Children: []*Node{leaf},
			Span:     leaf.Span,
		}, nil
	}

	p.recordExprExpectations(spec)
	return nil, errNoMatch
}

// recordExprExpectations records every token that could start an operand,
// so farthest-failure errors can say what was expected.
func (p *astParser) recordExprExpectations(spec *exprSpec) {
	for _, atom := range spec.atoms {
		p.recordFailure(atom.token)
	}
	for _, g := range spec.groups {
		p.recordFailure(g.open)
	}
	for tokName := range spec.prefix {
		p.recordFailure(tokName)
	}
}
