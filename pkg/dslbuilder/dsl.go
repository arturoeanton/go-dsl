// Package dslbuilder provides a dynamic DSL (Domain Specific Language) builder for Go.
// It allows you to create custom languages with tokens, rules, and actions that can
// call Go functions dynamically.
package dslbuilder

import (
	"fmt"
	"regexp"
	"strings"
)

// DSL represents a Domain Specific Language instance
type DSL struct {
	name      string
	grammar   *Grammar
	actions   map[string]ActionFunc
	functions map[string]interface{} // Go functions available to DSL
	context   map[string]interface{}
}

// ActionFunc is a function that processes parsed tokens
type ActionFunc func(args []interface{}) (interface{}, error)

// New creates a new DSL instance
func New(name string) *DSL {
	return &DSL{
		name:      name,
		grammar:   NewGrammar(),
		actions:   make(map[string]ActionFunc),
		functions: make(map[string]interface{}),
		context:   make(map[string]interface{}),
	}
}

// Token defines a token in the DSL
func (d *DSL) Token(name, pattern string) error {
	return d.grammar.AddToken(name, pattern)
}

// KeywordToken defines a keyword token with high priority
func (d *DSL) KeywordToken(name, keyword string) error {
	return d.grammar.AddKeywordToken(name, keyword)
}

// Rule defines a grammar rule
func (d *DSL) Rule(name string, pattern []string, actionName string) {
	d.grammar.AddRule(name, pattern, actionName)
}

// Action registers an action function
func (d *DSL) Action(name string, fn ActionFunc) {
	d.actions[name] = fn
	d.grammar.actions[name] = fn
}

// SetContext sets a context variable
func (d *DSL) SetContext(key string, value interface{}) {
	d.context[key] = value
}

// GetContext gets a context variable
func (d *DSL) GetContext(key string) interface{} {
	return d.context[key]
}

// Set registers a Go function that can be called from DSL code
func (d *DSL) Set(name string, fn interface{}) {
	d.functions[name] = fn
}

// Get retrieves a registered function
func (d *DSL) Get(name string) (interface{}, bool) {
	fn, exists := d.functions[name]
	return fn, exists
}

// Debug returns debug information about the DSL
func (d *DSL) Debug() map[string]interface{} {
	debug := make(map[string]interface{})
	debug["name"] = d.name
	debug["tokens"] = make(map[string]string)
	debug["rules"] = make(map[string]interface{})
	
	// Add token info
	for name, token := range d.grammar.tokens {
		debug["tokens"].(map[string]string)[name] = token.pattern
	}
	
	// Add rule info
	for name, rule := range d.grammar.rules {
		alternatives := make([]map[string]interface{}, 0)
		for _, alt := range rule.alternatives {
			altInfo := map[string]interface{}{
				"sequence": alt.sequence,
				"action":   alt.action,
			}
			alternatives = append(alternatives, altInfo)
		}
		debug["rules"].(map[string]interface{})[name] = alternatives
	}
	
	return debug
}

// DebugTokens returns the tokens that would be generated for a given code
func (d *DSL) DebugTokens(code string) ([]TokenMatch, error) {
	parser := NewParser(d.grammar)
	err := parser.tokenize(code)
	if err != nil {
		return nil, err
	}
	return parser.tokens, nil
}

// Use evaluates DSL code with an optional context
func (d *DSL) Use(code string, ctx map[string]interface{}) (*Result, error) {
	// Merge context if provided
	if ctx != nil {
		for k, v := range ctx {
			d.context[k] = v
		}
	}
	
	return d.Parse(code)
}

// Parse parses and evaluates DSL code
func (d *DSL) Parse(code string) (*Result, error) {
	parser := NewImprovedParser(d.grammar)
	parser.dsl = d // Give parser access to DSL functions
	ast, err := parser.Parse(code)
	if err != nil {
		return nil, fmt.Errorf("parsing error: %w", err)
	}

	return &Result{
		AST:    ast,
		Code:   code,
		Output: ast,
		DSL:    d,
	}, nil
}

// Grammar represents the grammar of a DSL
type Grammar struct {
	rules     map[string]*Rule
	tokens    map[string]*Token
	startRule string
	actions   map[string]ActionFunc
}

// Rule represents a grammar rule
type Rule struct {
	name         string
	alternatives []*Alternative
}

// Alternative represents an alternative in a rule
type Alternative struct {
	sequence []string
	action   string
}

// Token represents a token in the grammar
type Token struct {
	name     string
	pattern  string
	regex    *regexp.Regexp
	priority int
}

// NewGrammar creates a new grammar
func NewGrammar() *Grammar {
	return &Grammar{
		rules:   make(map[string]*Rule),
		tokens:  make(map[string]*Token),
		actions: make(map[string]ActionFunc),
	}
}

// AddToken adds a token to the grammar
func (g *Grammar) AddToken(name, pattern string) error {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}

	g.tokens[name] = &Token{
		name:     name,
		pattern:  pattern,
		regex:    regex,
		priority: 0,
	}
	return nil
}

// AddKeywordToken adds a keyword token with high priority
func (g *Grammar) AddKeywordToken(name, keyword string) error {
	pattern := "(?i)\\b" + regexp.QuoteMeta(keyword) + "\\b"
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid keyword pattern: %w", err)
	}

	g.tokens[name] = &Token{
		name:     name,
		pattern:  pattern,
		regex:    regex,
		priority: 90, // High priority for keywords
	}
	return nil
}

// AddRule adds a rule to the grammar
func (g *Grammar) AddRule(name string, sequence []string, action string) {
	rule, exists := g.rules[name]
	if !exists {
		rule = &Rule{
			name:         name,
			alternatives: []*Alternative{},
		}
		g.rules[name] = rule
		if g.startRule == "" {
			g.startRule = name
		}
	}

	rule.alternatives = append(rule.alternatives, &Alternative{
		sequence: sequence,
		action:   action,
	})
}

// Parser represents a DSL parser
type Parser struct {
	grammar *Grammar
	tokens  []TokenMatch
	pos     int
	dsl     *DSL // Reference to parent DSL for function access
}

// TokenMatch represents a matched token
type TokenMatch struct {
	TokenType string
	Value     string
	Start     int
	End       int
}

// NewParser creates a new parser
func NewParser(grammar *Grammar) *Parser {
	return &Parser{
		grammar: grammar,
		tokens:  []TokenMatch{},
		pos:     0,
	}
}

// Parse parses DSL code
func (p *Parser) Parse(code string) (interface{}, error) {
	// Reset parser state
	p.tokens = []TokenMatch{}
	p.pos = 0

	// Tokenize
	err := p.tokenize(code)
	if err != nil {
		return nil, err
	}

	// Parse from start rule
	p.pos = 0
	return p.parseRule(p.grammar.startRule)
}

// tokenize converts code into tokens
func (p *Parser) tokenize(code string) error {
	code = strings.TrimSpace(code)
	pos := 0

	for pos < len(code) {
		// Skip whitespace
		if code[pos] == ' ' || code[pos] == '\t' || code[pos] == '\n' || code[pos] == '\r' {
			pos++
			continue
		}

		matched := false
		bestMatch := TokenMatch{}
		bestLength := 0
		bestPriority := -1

		// Find best matching token
		for _, token := range p.grammar.tokens {
			if matches := token.regex.FindStringIndex(code[pos:]); matches != nil && matches[0] == 0 {
				matchLength := matches[1]

				// Higher priority or longer match wins
				shouldReplace := false
				if token.priority > bestPriority {
					shouldReplace = true
				} else if token.priority == bestPriority && matchLength > bestLength {
					shouldReplace = true
				}

				if shouldReplace {
					bestLength = matchLength
					bestPriority = token.priority
					bestMatch = TokenMatch{
						TokenType: token.name,
						Value:     code[pos : pos+matchLength],
						Start:     pos,
						End:       pos + matchLength,
					}
					matched = true
				}
			}
		}

		if matched {
			p.tokens = append(p.tokens, bestMatch)
			pos += bestLength
		} else {
			return fmt.Errorf("unexpected character at position %d: %c", pos, code[pos])
		}
	}

	return nil
}

// parseRule parses a specific rule
func (p *Parser) parseRule(ruleName string) (interface{}, error) {
	rule, exists := p.grammar.rules[ruleName]
	if !exists {
		return nil, fmt.Errorf("rule %s not found", ruleName)
	}

	// Try each alternative
	for _, alt := range rule.alternatives {
		savedPos := p.pos
		result, err := p.parseAlternative(alt)
		if err == nil {
			return result, nil
		}
		// Restore position if failed
		p.pos = savedPos
	}

	return nil, fmt.Errorf("no alternative matched for rule %s", ruleName)
}

// parseAlternative parses a specific alternative
func (p *Parser) parseAlternative(alt *Alternative) (interface{}, error) {
	var results []interface{}

	for _, symbol := range alt.sequence {
		if p.pos >= len(p.tokens) {
			return nil, fmt.Errorf("unexpected end of input")
		}

		// Check if symbol is a token
		if _, isToken := p.grammar.tokens[symbol]; isToken {
			if p.tokens[p.pos].TokenType == symbol {
				results = append(results, p.tokens[p.pos].Value)
				p.pos++
			} else {
				return nil, fmt.Errorf("expected token %s, got %s", symbol, p.tokens[p.pos].TokenType)
			}
		} else {
			// Symbol is a rule
			result, err := p.parseRule(symbol)
			if err != nil {
				return nil, err
			}
			results = append(results, result)
		}
	}

	// Apply action if available
	if alt.action != "" {
		if action, exists := p.grammar.actions[alt.action]; exists {
			result, err := action(results)
			if err != nil {
				return nil, err
			}
			return result, nil
		}
	}

	return results, nil
}

// Result represents the result of parsing DSL code
type Result struct {
	AST    interface{}
	Code   string
	Output interface{}
	DSL    *DSL // Reference to DSL for function calls
}

// GetOutput returns the final output
func (r *Result) GetOutput() interface{} {
	return r.Output
}

// String returns a string representation of the result
func (r *Result) String() string {
	if r.Output == nil {
		return fmt.Sprintf("DSL[%s] -> <no result>", r.Code)
	}
	return fmt.Sprintf("DSL[%s] -> %v", r.Code, r.Output)
}