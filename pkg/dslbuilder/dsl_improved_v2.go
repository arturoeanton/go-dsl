// Package dslbuilder provides improved left recursion handling.
// This file contains the enhanced version of the ImprovedParser with full
// left recursion support using the growing seed algorithm.
package dslbuilder

import (
	"fmt"
)

// leftRecEntry represents an entry in the left recursion detection stack
type leftRecEntry struct {
	rule     string
	startPos int
	seed     interface{}
	detected bool
}

// ImprovedParserV2 extends ImprovedParser with better left recursion handling
type ImprovedParserV2 struct {
	*ImprovedParser
	leftRecStack []leftRecEntry  // Stack for detecting left recursion
	heads        map[string]head // Head entries for left recursion
}

// head represents the head of a left-recursive parse
type head struct {
	rule        string
	involvedSet map[string]bool
	evalSet     map[string]bool
}

// NewImprovedParserV2 creates a parser with enhanced left recursion support
func NewImprovedParserV2(grammar *Grammar) *ImprovedParserV2 {
	return &ImprovedParserV2{
		ImprovedParser: NewImprovedParser(grammar),
		leftRecStack:   []leftRecEntry{},
		heads:          make(map[string]head),
	}
}

// parseRuleWithMemoV2 implements the growing seed algorithm for left recursion
func (p *ImprovedParserV2) parseRuleWithMemoV2(ruleName string) (interface{}, error) {
	startPos := p.pos
	
	// Check if we're already in a left-recursive call
	for _, entry := range p.leftRecStack {
		if entry.rule == ruleName && entry.startPos == startPos {
			// Left recursion detected - return current seed
			if entry.detected {
				return entry.seed, nil
			}
			// Mark as detected
			entry.detected = true
			return nil, fmt.Errorf("left recursion detected for %s", ruleName)
		}
	}
	
	// Check memo table first
	if ruleMemo, exists := p.memo[ruleName]; exists {
		if entry, exists := ruleMemo[startPos]; exists {
			p.pos = entry.endPos
			return entry.result, entry.err
		}
	} else {
		p.memo[ruleName] = make(map[int]memoEntry)
	}
	
	// Push to left recursion stack
	p.leftRecStack = append(p.leftRecStack, leftRecEntry{
		rule:     ruleName,
		startPos: startPos,
		seed:     nil,
		detected: false,
	})
	defer func() {
		// Pop from stack
		if len(p.leftRecStack) > 0 {
			p.leftRecStack = p.leftRecStack[:len(p.leftRecStack)-1]
		}
	}()
	
	// Try parsing with growing seed algorithm
	var seed interface{}
	var seedErr error
	var lastPos int
	
	// Initial parse attempt
	seed, seedErr = p.parseRuleRegular(ruleName)
	if seedErr != nil {
		// No initial match - not left recursive at this position
		p.memo[ruleName][startPos] = memoEntry{result: nil, endPos: p.pos, err: seedErr}
		return nil, seedErr
	}
	
	lastPos = p.pos
	
	// Growing phase - keep trying to extend the parse
	for {
		// Update seed in stack
		if len(p.leftRecStack) > 0 {
			p.leftRecStack[len(p.leftRecStack)-1].seed = seed
		}
		
		// Reset position for next attempt
		p.pos = startPos
		
		// Clear memo for this position to force re-parse
		delete(p.memo[ruleName], startPos)
		
		// Try to grow the seed
		newSeed, newErr := p.parseRuleWithLeftRec(ruleName, seed)
		
		// Check if we made progress
		if newErr != nil || p.pos <= lastPos {
			// No improvement - we're done
			p.pos = lastPos
			p.memo[ruleName][startPos] = memoEntry{result: seed, endPos: p.pos, err: seedErr}
			return seed, seedErr
		}
		
		// We grew the seed - continue
		seed = newSeed
		seedErr = newErr
		lastPos = p.pos
	}
}

// parseRuleWithLeftRec attempts to parse a rule with a known seed for left recursion
func (p *ImprovedParserV2) parseRuleWithLeftRec(ruleName string, seed interface{}) (interface{}, error) {
	rule, exists := p.grammar.rules[ruleName]
	if !exists {
		return nil, fmt.Errorf("rule %s not found", ruleName)
	}
	
	savedPos := p.pos
	var bestResult interface{}
	var bestErr error
	bestEndPos := savedPos
	
	// Try each alternative
	for _, alt := range rule.alternatives {
		p.pos = savedPos
		
		// For left-recursive alternatives, inject the seed
		if len(alt.sequence) > 0 && alt.sequence[0] == ruleName {
			// This is a left-recursive alternative
			var results []interface{}
			results = append(results, seed) // Use seed as first element
			
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
					// Symbol is a rule - parse recursively
					result, err := p.parseRuleWithMemoV2(symbol)
					if err != nil {
						success = false
						break
					}
					results = append(results, result)
				}
			}
			
			if success && p.pos > bestEndPos {
				// Apply action if available
				if alt.action != "" {
					if action, exists := p.grammar.actions[alt.action]; exists {
						actionResult, err := action(results)
						if err == nil {
							bestResult = actionResult
							bestErr = nil
							bestEndPos = p.pos
						}
					}
				} else {
					bestResult = results
					bestErr = nil
					bestEndPos = p.pos
				}
			}
		}
	}
	
	p.pos = bestEndPos
	if bestErr == nil && bestResult != nil {
		return bestResult, nil
	}
	
	return nil, fmt.Errorf("no alternative matched for rule %s", ruleName)
}

// Parse overrides the base Parse method to use the enhanced algorithm
func (p *ImprovedParserV2) Parse(code string) (interface{}, error) {
	// Reset parser state
	p.tokens = []TokenMatch{}
	p.pos = 0
	p.memo = make(map[string]map[int]memoEntry)
	p.input = code
	p.leftRecStack = []leftRecEntry{}
	p.heads = make(map[string]head)
	
	// Tokenize
	err := p.tokenize(code)
	if err != nil {
		return nil, err
	}
	
	// Parse from start rule using enhanced algorithm
	p.pos = 0
	result, err := p.parseRuleWithMemoV2(p.grammar.startRule)
	
	// Check if we consumed all tokens
	if err == nil && p.pos < len(p.tokens) {
		message := fmt.Sprintf("unexpected token: %s", p.tokens[p.pos].Value)
		return nil, createParseError(message, p.tokens[p.pos].Start, p.tokens[p.pos].Value, p.input)
	}
	
	return result, err
}