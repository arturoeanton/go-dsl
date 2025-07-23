// Package dslbuilder provides a dynamic DSL (Domain Specific Language) builder for Go.
// This is an improved version that handles left recursion properly.
package dslbuilder

import (
	"fmt"
	"strings"
)

// ImprovedParser represents an improved DSL parser that handles left recursion
type ImprovedParser struct {
	grammar *Grammar
	tokens  []TokenMatch
	pos     int
	dsl     *DSL
	memo    map[string]map[int]memoEntry // Memoization for packrat parsing
}

type memoEntry struct {
	result interface{}
	endPos int
	err    error
}

// NewImprovedParser creates a new improved parser
func NewImprovedParser(grammar *Grammar) *ImprovedParser {
	return &ImprovedParser{
		grammar: grammar,
		tokens:  []TokenMatch{},
		pos:     0,
		memo:    make(map[string]map[int]memoEntry),
	}
}

// Parse parses DSL code with left recursion handling
func (p *ImprovedParser) Parse(code string) (interface{}, error) {
	// Reset parser state
	p.tokens = []TokenMatch{}
	p.pos = 0
	p.memo = make(map[string]map[int]memoEntry)

	// Tokenize
	err := p.tokenize(code)
	if err != nil {
		return nil, err
	}
	

	// Parse from start rule
	p.pos = 0
	result, err := p.parseRuleWithMemo(p.grammar.startRule)
	
	// Check if we consumed all tokens
	if err == nil && p.pos < len(p.tokens) {
		return nil, fmt.Errorf("unexpected token at position %d: %s", p.pos, p.tokens[p.pos].Value)
	}
	
	return result, err
}

// tokenize converts code into tokens (same as original)
func (p *ImprovedParser) tokenize(code string) error {
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

// parseRuleWithMemo parses a rule with memoization to handle left recursion
func (p *ImprovedParser) parseRuleWithMemo(ruleName string) (interface{}, error) {
	// Check memo table
	if ruleMemo, exists := p.memo[ruleName]; exists {
		if entry, exists := ruleMemo[p.pos]; exists {
			p.pos = entry.endPos
			return entry.result, entry.err
		}
	} else {
		p.memo[ruleName] = make(map[int]memoEntry)
	}

	startPos := p.pos
	
	// Use iterative approach for left-recursive rules
	if p.isLeftRecursive(ruleName) {
		result, err := p.parseLeftRecursive(ruleName)
		p.memo[ruleName][startPos] = memoEntry{result: result, endPos: p.pos, err: err}
		return result, err
	}

	// Regular recursive parsing for non-left-recursive rules
	result, err := p.parseRuleRegular(ruleName)
	p.memo[ruleName][startPos] = memoEntry{result: result, endPos: p.pos, err: err}
	return result, err
}

// isLeftRecursive checks if a rule is left-recursive
func (p *ImprovedParser) isLeftRecursive(ruleName string) bool {
	rule, exists := p.grammar.rules[ruleName]
	if !exists {
		return false
	}

	// Check if any alternative starts with the rule itself
	for _, alt := range rule.alternatives {
		if len(alt.sequence) > 0 && alt.sequence[0] == ruleName {
			return true
		}
	}
	return false
}

// parseLeftRecursive handles left-recursive rules iteratively
func (p *ImprovedParser) parseLeftRecursive(ruleName string) (interface{}, error) {
	rule, exists := p.grammar.rules[ruleName]
	if !exists {
		return nil, fmt.Errorf("rule %s not found", ruleName)
	}

	// First, try non-left-recursive alternatives to get the base
	var base interface{}
	baseFound := false
	savedPos := p.pos

	for _, alt := range rule.alternatives {
		// Skip left-recursive alternatives
		if len(alt.sequence) > 0 && alt.sequence[0] == ruleName {
			continue
		}

		p.pos = savedPos
		result, err := p.parseAlternative(alt)
		if err == nil {
			base = result
			baseFound = true
			break
		}
	}

	if !baseFound {
		return nil, fmt.Errorf("no base case found for left-recursive rule %s", ruleName)
	}

	// Now iteratively apply left-recursive alternatives
	for {
		improved := false
		savedPos = p.pos

		for _, alt := range rule.alternatives {
			// Only process left-recursive alternatives
			if len(alt.sequence) == 0 || alt.sequence[0] != ruleName {
				continue
			}

			// Try to parse the rest of the sequence
			p.pos = savedPos
			var results []interface{}
			results = append(results, base) // Use current base as first element

			// Parse remaining symbols
			success := true
			for i := 1; i < len(alt.sequence); i++ {
				symbol := alt.sequence[i]
				
				if p.pos >= len(p.tokens) {
					success = false
					break
				}

				// Check if symbol is a token
				if _, isToken := p.grammar.tokens[symbol]; isToken {
					if p.tokens[p.pos].TokenType == symbol {
						results = append(results, p.tokens[p.pos].Value)
						p.pos++
					} else {
						success = false
						break
					}
				} else {
					// Symbol is a rule
					result, err := p.parseRuleWithMemo(symbol)
					if err != nil {
						success = false
						break
					}
					results = append(results, result)
				}
			}

			if success {
				// Apply action if available
				if alt.action != "" {
					if action, exists := p.grammar.actions[alt.action]; exists {
						newBase, err := action(results)
						if err == nil {
							base = newBase
							improved = true
							break
						}
					}
				} else {
					base = results
					improved = true
					break
				}
			}
		}

		if !improved {
			// No more improvements possible
			break
		}
	}

	return base, nil
}

// parseRuleRegular handles non-left-recursive rules
func (p *ImprovedParser) parseRuleRegular(ruleName string) (interface{}, error) {
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

	return nil, fmt.Errorf("no alternative matched for rule %s at position %d", ruleName, p.pos)
}

// parseAlternative parses a specific alternative
func (p *ImprovedParser) parseAlternative(alt *Alternative) (interface{}, error) {
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
			result, err := p.parseRuleWithMemo(symbol)
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

// Parse parses using the improved parser
