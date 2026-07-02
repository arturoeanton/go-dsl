// Package dslbuilder - Error recovery and multi-error diagnostics.
//
// Parse stops at the first syntax error. Diagnostics keeps going: it
// collects lexical errors (skipping the offending characters) and applies
// panic-mode recovery after each parse failure (skipping past the farthest
// failure point), so a single pass reports every problem in the input.
// This is what editor integrations (see cmd/lsp) use to underline all
// errors at once.
package dslbuilder

import "fmt"

// maxDiagnostics bounds how many errors a single Diagnostics pass reports,
// so adversarial input cannot produce unbounded work or output.
const maxDiagnostics = 50

// Diagnostics parses code and returns every syntax and lexical error it
// can find, instead of stopping at the first one.
//
// Recovery strategy:
//   - Lexical errors: the offending character is reported and skipped.
//   - Parse errors: the farthest-failure error is reported, then parsing
//     resumes right after the token where it occurred (panic mode).
//   - Statements: after a successful parse that doesn't consume all input,
//     parsing continues from the next token, so multi-statement inputs are
//     diagnosed statement by statement.
//
// An empty slice means the input is clean. Parse/ParseAST behavior is
// unchanged — they still stop at the first error.
func (d *DSL) Diagnostics(code string) []*ParseError {
	var diags []*ParseError

	tokens, lexErrs := tokenizeTolerant(d.grammar, code)
	diags = append(diags, lexErrs...)
	if len(diags) >= maxDiagnostics {
		return diags[:maxDiagnostics]
	}

	startRule := d.grammar.startRule
	if startRule == "" {
		diags = append(diags, createParseError("start rule not found (grammar has no rules)", 0, "", code))
		return diags
	}

	parser := newASTParser(d.grammar)
	parser.input = code
	parser.tokens = tokens

	// Synchronization set: tokens that can begin the start rule. After an
	// error we resume at the next such token, which avoids cascading
	// spurious errors from resuming mid-statement.
	sync := firstTokens(d.grammar, startRule)

	pos := 0
	for pos < len(tokens) && len(diags) < maxDiagnostics {
		parser.pos = pos
		parser.failure = parseFailure{tokenPos: -1}
		parser.ruleStack = nil
		parser.hintStack = nil

		_, err := parser.parseRule(startRule)
		if err == nil {
			if parser.pos <= pos {
				pos++ // zero-width match: force progress
			} else {
				pos = parser.pos
			}
			continue
		}

		if err != errNoMatch {
			// Grammar-level problem (missing rule, depth exceeded): not
			// recoverable by skipping input.
			diags = append(diags, toParseError(err, code))
			break
		}

		if perr, ok := parser.farthestError(startRule).(*ParseError); ok {
			diags = append(diags, perr)
		}

		// Panic-mode recovery: resume at the next token that can begin the
		// start rule, searching from the farthest failure point. This
		// prevents cascading errors from resuming mid-statement.
		next := parser.failure.tokenPos
		if next <= pos {
			next = pos + 1
		}
		for next < len(tokens) && !sync[tokens[next].TokenType] {
			next++
		}
		if next <= pos {
			next = pos + 1
		}
		pos = next
	}

	return diags
}

// firstTokens computes the set of token types that can appear first in a
// derivation of the given rule (the FIRST set restricted to terminals).
func firstTokens(g *Grammar, ruleName string) map[string]bool {
	first := make(map[string]bool)
	visited := make(map[string]bool)

	var visit func(name string)
	visit = func(name string) {
		if visited[name] {
			return
		}
		visited[name] = true

		if _, isToken := g.tokens[name]; isToken {
			first[name] = true
			return
		}
		if spec, ok := g.exprRules[name]; ok {
			for _, a := range spec.atoms {
				first[a.token] = true
			}
			for _, gr := range spec.groups {
				first[gr.open] = true
			}
			for tok := range spec.prefix {
				first[tok] = true
			}
			return
		}
		if rule, ok := g.rules[name]; ok {
			for _, alt := range rule.alternatives {
				if len(alt.sequence) > 0 {
					visit(alt.sequence[0])
				}
			}
		}
	}
	visit(ruleName)
	return first
}

// tokenizeTolerant is the error-recovering variant of tokenizeInput: instead
// of stopping at the first unmatchable character it records a diagnostic,
// skips one byte, and keeps lexing.
func tokenizeTolerant(grammar *Grammar, code string) ([]TokenMatch, []*ParseError) {
	var tokens []TokenMatch
	var errs []*ParseError

	pos := 0
	for pos < len(code) {
		c := code[pos]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			pos++
			continue
		}

		match, ok := grammar.matchTokenAt(code, pos)
		if !ok {
			if len(errs) < maxDiagnostics {
				message := fmt.Sprintf("unexpected character: %c", c)
				errs = append(errs, createParseError(message, pos, string(c), code))
			}
			pos++
			continue
		}

		tokens = append(tokens, match)
		pos = match.End
	}

	return tokens, errs
}

// toParseError wraps any error as a *ParseError for uniform diagnostics.
func toParseError(err error, input string) *ParseError {
	if perr, ok := err.(*ParseError); ok {
		return perr
	}
	return &ParseError{Message: err.Error(), Input: input}
}
