// Package dslbuilder provides a dynamic DSL (Domain Specific Language) builder for Go.
// This file contains the improved parser implementation that handles left recursion
// using memoization (Packrat parsing) for better performance and more flexible grammars.
package dslbuilder

import (
	"fmt"
	"strings"
)

// ImprovedParser represents an improved DSL parser that handles left recursion.
// It uses memoization (Packrat parsing) to efficiently parse grammars with
// left-recursive rules and achieve linear time complexity.
//
// Key improvements over basic parser:
//   - Handles left recursion (e.g., expr -> expr '+' term)
//   - Memoization prevents exponential backtracking
//   - Better error reporting with position tracking
//   - Support for operator precedence and associativity
//
// Fields:
//   - grammar: Language grammar definition
//   - tokens: Tokenized input
//   - pos: Current token position
//   - dsl: Parent DSL for function/context access
//   - memo: Memoization table for Packrat parsing
//   - input: Original input for error messages
//   - leftRecStack: Stack for detecting left recursion cycles
//   - growing: Track rules currently being grown
type ImprovedParser struct {
	grammar      *Grammar
	tokens       []TokenMatch
	pos          int
	dsl          *DSL
	memo         map[string]map[int]memoEntry // Memoization for packrat parsing
	input        string                       // Original input for error reporting
	leftRecStack []string                     // Stack to detect left recursion
	growing      map[string]bool              // Rules currently being grown
}

// memoEntry stores the cached result of parsing a rule at a specific position.
// This is the core data structure for Packrat parsing that enables linear time
// parsing even with backtracking and left recursion.
//
// Fields:
//   - result: The parsed value if successful
//   - endPos: Token position after successful parse
//   - err: Error if parsing failed
type memoEntry struct {
	result interface{} // Parsed result value
	endPos int         // Position after parsing
	err    error       // Error if failed
}

// NewImprovedParser creates a new improved parser with memoization support.
// This parser can handle left-recursive grammars and provides better performance
// than the basic recursive descent parser.
//
// Example:
//
//	parser := NewImprovedParser(grammar)
//	result, err := parser.Parse("x = 1 + 2 * 3")
func NewImprovedParser(grammar *Grammar) *ImprovedParser {
	return &ImprovedParser{
		grammar:      grammar,
		tokens:       []TokenMatch{},
		pos:          0,
		memo:         make(map[string]map[int]memoEntry),
		input:        "", // Will be set during parsing
		leftRecStack: []string{},
		growing:      make(map[string]bool),
	}
}

// Parse parses DSL code with left recursion handling using memoization.
// This is the main entry point for the improved parser.
//
// The parsing process:
//  1. Tokenization: Convert input to tokens
//  2. Memoized parsing: Parse with caching to handle left recursion
//  3. Completeness check: Ensure all input was consumed
//
// Returns:
//   - Parsed result from the start rule's action
//   - ParseError with detailed position info on failure
//
// Example:
//
//	result, err := parser.Parse("2 + 3 * 4")
//	if err != nil {
//	    if parseErr, ok := err.(*ParseError); ok {
//	        fmt.Println(parseErr.DetailedError())
//	    }
//	}
func (p *ImprovedParser) Parse(code string) (interface{}, error) {
	// Reset parser state
	p.tokens = []TokenMatch{}
	p.pos = 0
	p.memo = make(map[string]map[int]memoEntry)
	p.input = code // Store input for error reporting
	p.leftRecStack = []string{}
	p.growing = make(map[string]bool)

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
		message := fmt.Sprintf("unexpected token: %s", p.tokens[p.pos].Value)
		return nil, createParseError(message, p.tokens[p.pos].Start, p.tokens[p.pos].Value, p.input)
	}

	return result, err
}

// tokenize converts code into tokens (lexical analysis).
// Uses the same algorithm as the basic parser but with
// input tracking for better error messages.
//
// Token matching priority:
//  1. Higher priority value wins (keywords > regular)
//  2. For same priority, longest match wins
//  3. Whitespace is automatically skipped
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
			message := fmt.Sprintf("unexpected character: %c", code[pos])
			return createParseError(message, pos, string(code[pos]), p.input)
		}
	}

	return nil
}

// parseRuleWithMemo parses a rule with memoization to handle left recursion.
// This is the core of the Packrat parsing algorithm that enables efficient
// parsing of left-recursive grammars.
//
// The memoization table prevents:
//   - Exponential backtracking in ambiguous grammars
//   - Infinite recursion in left-recursive rules
//   - Redundant parsing of the same rule at the same position
//
// Algorithm:
//  1. Check if result is already memoized
//  2. If left-recursive, use iterative algorithm
//  3. Otherwise, use standard recursive descent
//  4. Cache the result for future use
//
// Returns the parsed result and updates the position.
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

// isLeftRecursive checks if a rule is directly left-recursive.
// A rule is left-recursive if it has an alternative that starts
// with the rule itself.
//
// Example of left-recursive rule:
//
//	expr → expr '+' term  (left-recursive)
//	expr → term           (base case)
//
// This detection is used to choose the appropriate parsing strategy.
// Note: This only detects direct left recursion, not indirect.
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

// parseLeftRecursive handles left-recursive rules using an improved growing seed algorithm.
// This prevents stack overflow and enables parsing of left-associative operators.
//
// Algorithm (growing seed approach):
//  1. Parse base case first (non-recursive alternatives)
//  2. Use the base as a "seed" for growing the parse
//  3. Iteratively try to extend the seed with recursive alternatives
//  4. Continue until no more growth is possible
//
// This naturally produces left-associative parse trees and handles
// complex cases like multiple options in HTTP requests.
//
// Parameters:
//   - ruleName: The left-recursive rule to parse
//
// Returns:
//   - The final parsed result after all possible extensions
//   - Error if no alternatives match
func (p *ImprovedParser) parseLeftRecursive(ruleName string) (interface{}, error) {
	rule, exists := p.grammar.rules[ruleName]
	if !exists {
		return nil, fmt.Errorf("rule %s not found", ruleName)
	}

	startPos := p.pos
	
	// Mark this rule as growing to prevent infinite recursion
	growKey := fmt.Sprintf("%s_%d", ruleName, startPos)
	if p.growing[growKey] {
		// Already growing this rule at this position - return failure
		return nil, fmt.Errorf("left recursion cycle detected")
	}
	p.growing[growKey] = true
	defer delete(p.growing, growKey)

	// Step 1: Try all alternatives to find an initial seed
	var seed interface{}
	seedPos := startPos
	foundSeed := false

	// First pass: try non-recursive alternatives
	for _, alt := range rule.alternatives {
		// Skip left-recursive alternatives in first pass
		if len(alt.sequence) > 0 && alt.sequence[0] == ruleName {
			continue
		}

		p.pos = startPos
		result, err := p.parseAlternative(alt)
		if err == nil {
			seed = result
			seedPos = p.pos
			foundSeed = true
			break
		}
	}

	// If no non-recursive alternative matched, try recursive ones with nil seed
	if !foundSeed {
		// This handles cases where all alternatives are left-recursive
		// We'll return an error but cache it to prevent re-parsing
		return nil, fmt.Errorf("no alternative matched for rule %s", ruleName)
	}

	// Step 2: Growing phase - iteratively extend the seed
	for {
		improved := false
		bestResult := seed
		bestPos := seedPos

		// Try each left-recursive alternative with current seed
		for _, alt := range rule.alternatives {
			// Only process left-recursive alternatives
			if len(alt.sequence) == 0 || alt.sequence[0] != ruleName {
				continue
			}

			// Reset position for this attempt
			p.pos = startPos
			
			// Temporarily install the seed in memo for this rule
			if p.memo[ruleName] == nil {
				p.memo[ruleName] = make(map[int]memoEntry)
			}
			p.memo[ruleName][startPos] = memoEntry{
				result: seed,
				endPos: seedPos,
				err:    nil,
			}

			// Try to parse this alternative
			var results []interface{}
			results = append(results, seed) // Use seed as first element

			// Parse remaining symbols after the recursive call
			p.pos = seedPos // Start after the seed
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
					// Symbol is a rule - parse it
					result, err := p.parseRuleWithMemo(symbol)
					if err != nil {
						success = false
						break
					}
					results = append(results, result)
				}
			}

			if success && p.pos > bestPos {
				// We found a longer match - apply action
				if alt.action != "" {
					if action, exists := p.grammar.actions[alt.action]; exists {
						actionResult, err := action(results)
						if err == nil {
							bestResult = actionResult
							bestPos = p.pos
							improved = true
						}
					}
				} else {
					bestResult = results
					bestPos = p.pos
					improved = true
				}
			}
		}

		if !improved {
			// No more growth possible
			p.pos = bestPos
			return bestResult, nil
		}

		// Update seed for next iteration
		seed = bestResult
		seedPos = bestPos
	}
}

// parseRuleRegular handles non-left-recursive rules using standard recursive descent.
// This is the traditional parsing approach for rules without left recursion.
//
// The parser tries each alternative in order:
//  1. Save current position
//  2. Try to parse the alternative
//  3. If successful, return the result
//  4. If failed, restore position and try next
//
// This provides ordered choice (PEG-like behavior) where the first
// matching alternative wins.
//
// Returns:
//   - Result of the first successful alternative
//   - ParseError with position info if no alternatives match
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

	// Create detailed error with current token position
	var token string
	var position int
	if p.pos < len(p.tokens) {
		token = p.tokens[p.pos].Value
		position = p.tokens[p.pos].Start
	} else {
		token = "<end of input>"
		position = len(p.input)
	}

	message := fmt.Sprintf("no alternative matched for rule %s", ruleName)
	return nil, createParseError(message, position, token, p.input)
}

// parseAlternative parses a specific alternative (sequence of symbols).
// This is shared by both regular and left-recursive parsing.
//
// Process:
//  1. Match each symbol in the sequence
//  2. For tokens: match exact token type
//  3. For rules: recursively parse with memoization
//  4. Collect all matched values
//  5. Apply action function if defined
//
// The results array passed to actions contains:
//   - Token values as strings
//   - Rule results as returned by their actions
//
// Example:
//
//	Alternative: ["IF", "expr", "THEN", "stmt"]
//	Results: ["if", exprValue, "then", stmtValue]
func (p *ImprovedParser) parseAlternative(alt *Alternative) (interface{}, error) {
	var results []interface{}

	for _, symbol := range alt.sequence {
		if p.pos >= len(p.tokens) {
			message := "unexpected end of input"
			position := len(p.input)
			return nil, createParseError(message, position, "<end of input>", p.input)
		}

		// Check if symbol is a token
		if _, isToken := p.grammar.tokens[symbol]; isToken {
			if p.tokens[p.pos].TokenType == symbol {
				results = append(results, p.tokens[p.pos].Value)
				p.pos++
			} else {
				message := fmt.Sprintf("expected token %s, got %s", symbol, p.tokens[p.pos].TokenType)
				return nil, createParseError(message, p.tokens[p.pos].Start, p.tokens[p.pos].Value, p.input)
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

// ParseWithContext enables parsing with additional context that can be accessed
// by action functions. This is useful for passing runtime configuration or
// variables to the DSL execution.
//
// Note: This appears to be a stub. The actual implementation would merge
// the provided context with the DSL's context before parsing.
