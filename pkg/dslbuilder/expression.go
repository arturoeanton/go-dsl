// Package dslbuilder - Pratt expression parser.
//
// Expression() lets you declare operators with explicit binding power and
// associativity, resolved via precedence climbing (Pratt parsing):
//   - Atoms & grouped sub-expressions: Atom, Group, GroupWithAction
//   - Prefix & Postfix unary operators: Prefix, Postfix
//   - Infix binary operators: InfixLeft, InfixRight, InfixNonAssoc
//   - Call expressions: Call (fn(a, b, ...))
//   - Index / subscript expressions: Index (arr[i])
//   - Member / property access: Member (obj.prop)
//   - Ternary conditional operator: Ternary (cond ? a : b)
package dslbuilder

import "fmt"

// exprOp describes one unary or binary operator of an expression rule.
type exprOp struct {
	token      string // Operator token name
	power      int    // Binding power (higher binds tighter)
	action     string // Action executed on evaluation
	rightAssoc bool   // Right associativity (for infix)
	nonAssoc   bool   // Non-associative (for infix: rejects a < b < c)
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

// exprCall describes function/method call syntax like callee(a, b).
type exprCall struct {
	open   string // Opening token (e.g. LPAREN)
	close  string // Closing token (e.g. RPAREN)
	sep    string // Optional argument separator token (e.g. COMMA; "" for none)
	power  int    // Postfix binding power
	action string // Action receiving [callee, []interface{}{args...}]
}

// exprIndex describes subscript syntax like target[index].
type exprIndex struct {
	open   string // Opening bracket token (e.g. LBRACK)
	inner  string // Rule for index expression ("" or self = this expression)
	close  string // Closing bracket token (e.g. RBRACK)
	power  int    // Postfix binding power
	action string // Action receiving [target, open, index, close]
}

// exprMember describes member access syntax like target.field.
type exprMember struct {
	dotToken   string // Accessor operator token (e.g. DOT)
	fieldToken string // Field identifier token (e.g. IDENT)
	power      int    // Postfix binding power
	action     string // Action receiving [target, dot, field]
}

// exprTernary describes conditional ternary syntax cond ? thenExpr : elseExpr.
type exprTernary struct {
	questionToken string // '?' token
	colonToken    string // ':' token
	power         int    // Binding power (right-associative)
	action        string // Action ([cond, "?", thenVal, ":", elseVal]) or lazy NodeAction
}

// exprSpec is the full specification of an expression rule.
type exprSpec struct {
	name    string
	atoms   []exprAtom
	groups  []exprGroup
	prefix  map[string]exprOp
	postfix map[string]exprOp
	infix   map[string]exprOp
	calls   map[string]exprCall
	indices map[string]exprIndex
	members map[string]exprMember
	ternary *exprTernary
}

func newExprSpec(name string) *exprSpec {
	return &exprSpec{
		name:    name,
		prefix:  map[string]exprOp{},
		postfix: map[string]exprOp{},
		infix:   map[string]exprOp{},
		calls:   map[string]exprCall{},
		indices: map[string]exprIndex{},
		members: map[string]exprMember{},
	}
}

// ExpressionBuilder declares a Pratt-parsed expression rule.
// Obtain one with DSL.Expression(name); all methods chain.
type ExpressionBuilder struct {
	dsl  *DSL
	spec *exprSpec
}

// Expression declares (or extends) an expression rule parsed with a Pratt
// parser. If it is the first rule declared, it becomes the start rule.
func (d *DSL) Expression(name string) *ExpressionBuilder {
	if d.grammar.frozen {
		d.addDeferredError(fmt.Errorf("grammar is frozen (Build() was called); cannot add expression rule %s", name))
		return &ExpressionBuilder{dsl: d, spec: newExprSpec(name)}
	}

	spec, exists := d.grammar.exprRules[name]
	if !exists {
		spec = newExprSpec(name)
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
// the group evaluates to the inner value.
func (b *ExpressionBuilder) Group(open, inner, close string) *ExpressionBuilder {
	b.spec.groups = append(b.spec.groups, exprGroup{open: open, inner: inner, close: close})
	return b
}

// GroupWithAction declares a grouping construct with a custom action receiving
// [openToken, innerValue, closeToken].
func (b *ExpressionBuilder) GroupWithAction(open, inner, close, action string) *ExpressionBuilder {
	b.spec.groups = append(b.spec.groups, exprGroup{open: open, inner: inner, close: close, action: action})
	return b
}

// Prefix declares a prefix (unary) operator. The action receives
// [operator, operand]. power is the binding power of the operand parse.
func (b *ExpressionBuilder) Prefix(token string, power int, action string) *ExpressionBuilder {
	b.spec.prefix[token] = exprOp{token: token, power: power, action: action}
	return b
}

// Postfix declares a postfix (unary) operator such as x++ or n!.
// The action receives [operand, operator].
func (b *ExpressionBuilder) Postfix(token string, power int, action string) *ExpressionBuilder {
	b.spec.postfix[token] = exprOp{token: token, power: power, action: action}
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

// InfixNonAssoc declares a non-associative binary operator such as == or <:
// chaining two non-associative operators at the same precedence (a < b < c)
// is rejected without parentheses. The action receives [left, operator, right].
func (b *ExpressionBuilder) InfixNonAssoc(token string, power int, action string) *ExpressionBuilder {
	b.spec.infix[token] = exprOp{token: token, power: power, action: action, nonAssoc: true}
	return b
}

// Call declares function-call postfix syntax: callee(arg1, arg2, ...).
// sep may be "" for whitespace-separated arguments.
// The action receives [calleeValue, []interface{}{argValues...}].
func (b *ExpressionBuilder) Call(open, close, sep string, power int, action string) *ExpressionBuilder {
	b.spec.calls[open] = exprCall{open: open, close: close, sep: sep, power: power, action: action}
	return b
}

// Index declares subscript syntax: target[index].
// If inner is "" or matches the expression rule's own name, the index is parsed
// using this expression rule. The action receives [target, open, index, close].
func (b *ExpressionBuilder) Index(open, inner, close string, power int, action string) *ExpressionBuilder {
	if inner == "" {
		inner = b.spec.name
	}
	b.spec.indices[open] = exprIndex{open: open, inner: inner, close: close, power: power, action: action}
	return b
}

// Member declares member/property access syntax: target.field.
// The action receives [target, dotToken, fieldToken].
func (b *ExpressionBuilder) Member(dotToken, fieldToken string, power int, action string) *ExpressionBuilder {
	b.spec.members[dotToken] = exprMember{dotToken: dotToken, fieldToken: fieldToken, power: power, action: action}
	return b
}

// Ternary declares a right-associative conditional ternary operator:
// cond ? thenExpr : elseExpr.
// Regular actions receive [cond, question, thenVal, colon, elseVal]; lazy
// NodeActions receive children [0]=cond, [2]=thenNode, [4]=elseNode.
func (b *ExpressionBuilder) Ternary(questionToken, colonToken string, power int, action string) *ExpressionBuilder {
	b.spec.ternary = &exprTernary{
		questionToken: questionToken,
		colonToken:    colonToken,
		power:         power,
		action:        action,
	}
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
// power an infix/postfix operator needs to continue extending the left operand.
func (p *astParser) parseExprBP(spec *exprSpec, minBP int) (*Node, error) {
	p.steps++
	if p.steps > maxParseSteps {
		return nil, errParseStepsExceeded
	}
	if p.depth >= maxParseDepth {
		return nil, errParseDepthExceeded
	}
	p.depth++
	defer func() { p.depth-- }()

	left, err := p.parseExprOperand(spec)
	if err != nil {
		return nil, err
	}

	lastNonAssocPower := -1

	for {
		p.touchPos()
		if p.pos >= len(p.tokens) {
			break
		}
		tok := p.tokens[p.pos]

		// 1. Call syntax: callee(arg1, arg2, ...)
		if callSpec, ok := spec.calls[tok.TokenType]; ok && callSpec.power >= minBP {
			openNode := tokenNode(tok)
			p.pos++
			step := 1
			if callSpec.sep != "" {
				step = 2
			}
			children := []*Node{left, openNode}

			p.touchPos()
			if p.pos >= len(p.tokens) {
				p.recordFailure(callSpec.close)
				return nil, errNoMatch
			}
			if p.tokens[p.pos].TokenType != callSpec.close {
				for {
					argNode, err := p.parseExprBP(spec, 0)
					if err != nil {
						return nil, err
					}
					children = append(children, argNode)

					p.touchPos()
					if p.pos >= len(p.tokens) {
						p.recordFailure(callSpec.close)
						if callSpec.sep != "" {
							p.recordFailure(callSpec.sep)
						}
						return nil, errNoMatch
					}
					if p.tokens[p.pos].TokenType == callSpec.close {
						break
					}
					if callSpec.sep != "" {
						if p.tokens[p.pos].TokenType != callSpec.sep {
							p.recordFailure(callSpec.sep)
							p.recordFailure(callSpec.close)
							return nil, errNoMatch
						}
						children = append(children, tokenNode(p.tokens[p.pos]))
						p.pos++
					}
				}
			}
			closeNode := tokenNode(p.tokens[p.pos])
			p.pos++
			children = append(children, closeNode)

			left = &Node{
				Rule:     spec.name,
				Action:   callSpec.action,
				Children: children,
				Span:     Span{Start: left.Span.Start, End: closeNode.Span.End},
				callStep: step,
			}
			lastNonAssocPower = -1
			continue
		}

		// 2. Index syntax: target[index]
		if idxSpec, ok := spec.indices[tok.TokenType]; ok && idxSpec.power >= minBP {
			openNode := tokenNode(tok)
			p.pos++

			var inner *Node
			if idxSpec.inner == spec.name || idxSpec.inner == "" {
				inner, err = p.parseExprBP(spec, 0)
			} else {
				inner, err = p.parseSymbol(idxSpec.inner)
			}
			if err != nil {
				return nil, err
			}

			p.touchPos()
			if p.pos >= len(p.tokens) || p.tokens[p.pos].TokenType != idxSpec.close {
				p.recordFailure(idxSpec.close)
				return nil, errNoMatch
			}
			closeNode := tokenNode(p.tokens[p.pos])
			p.pos++

			left = &Node{
				Rule:     spec.name,
				Action:   idxSpec.action,
				Children: []*Node{left, openNode, inner, closeNode},
				Span:     Span{Start: left.Span.Start, End: closeNode.Span.End},
			}
			lastNonAssocPower = -1
			continue
		}

		// 3. Member access: target.field
		if memSpec, ok := spec.members[tok.TokenType]; ok && memSpec.power >= minBP {
			dotNode := tokenNode(tok)
			p.pos++

			p.touchPos()
			if p.pos >= len(p.tokens) || p.tokens[p.pos].TokenType != memSpec.fieldToken {
				p.recordFailure(memSpec.fieldToken)
				return nil, errNoMatch
			}
			fieldNode := tokenNode(p.tokens[p.pos])
			p.pos++

			left = &Node{
				Rule:     spec.name,
				Action:   memSpec.action,
				Children: []*Node{left, dotNode, fieldNode},
				Span:     Span{Start: left.Span.Start, End: fieldNode.Span.End},
			}
			lastNonAssocPower = -1
			continue
		}

		// 4. Postfix unary operator: x++, n!
		if postOp, ok := spec.postfix[tok.TokenType]; ok && postOp.power >= minBP {
			opNode := tokenNode(tok)
			p.pos++
			left = &Node{
				Rule:     spec.name,
				Action:   postOp.action,
				Children: []*Node{left, opNode},
				Span:     Span{Start: left.Span.Start, End: opNode.Span.End},
			}
			lastNonAssocPower = -1
			continue
		}

		// 5. Ternary conditional operator: cond ? thenExpr : elseExpr
		if spec.ternary != nil && tok.TokenType == spec.ternary.questionToken && spec.ternary.power >= minBP {
			qNode := tokenNode(tok)
			p.pos++

			thenNode, err := p.parseExprBP(spec, 0)
			if err != nil {
				return nil, err
			}

			p.touchPos()
			if p.pos >= len(p.tokens) || p.tokens[p.pos].TokenType != spec.ternary.colonToken {
				p.recordFailure(spec.ternary.colonToken)
				return nil, errNoMatch
			}
			colonNode := tokenNode(p.tokens[p.pos])
			p.pos++

			// Right-associative: allow nested ternary at equal binding power.
			elseNode, err := p.parseExprBP(spec, spec.ternary.power)
			if err != nil {
				return nil, err
			}

			left = &Node{
				Rule:     spec.name,
				Action:   spec.ternary.action,
				Children: []*Node{left, qNode, thenNode, colonNode, elseNode},
				Span:     Span{Start: left.Span.Start, End: elseNode.Span.End},
			}
			lastNonAssocPower = -1
			continue
		}

		// 6. Binary infix operator
		op, ok := spec.infix[tok.TokenType]
		if !ok || op.power < minBP {
			break
		}
		if lastNonAssocPower >= 0 && op.power == lastNonAssocPower {
			// Disallow chaining non-associative operators at the same precedence level.
			break
		}

		opNode := tokenNode(tok)
		p.pos++

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
		if op.nonAssoc {
			lastNonAssocPower = op.power
		} else {
			lastNonAssocPower = -1
		}
	}

	return left, nil
}

// parseExprOperand parses a primary expression: a prefix operator applied
// to an operand, a group, or an atom.
func (p *astParser) parseExprOperand(spec *exprSpec) (*Node, error) {
	p.touchPos()
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

		p.touchPos()
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
			group:    g.action == "",
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
