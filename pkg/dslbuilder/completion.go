// Package dslbuilder - Completion candidates.
//
// Completions answers "what can come next at this cursor position?" by
// running the parser over the text before the cursor and reading the
// farthest-failure expectations — the same machinery that powers error
// messages, so suggestions are always consistent with what the parser
// actually accepts. cmd/lsp exposes this as textDocument/completion.
package dslbuilder

import (
	"regexp"
	"sort"
	"strings"
)

// Completion is one suggestion for the next token at a cursor position.
type Completion struct {
	Label     string // Text to insert (keyword/literal) or token placeholder
	Token     string // Grammar token name this suggestion came from
	IsKeyword bool   // True when Label is literal text (keyword or literal token)
	Detail    string // Human-readable detail (e.g. the token pattern)
}

// Completions returns the tokens that could validly appear at the given
// byte offset, based on parsing the text before the cursor:
//
//   - If the prefix ends mid-statement, the candidates are the parser's
//     farthest-failure expectations at the cursor.
//   - If the prefix parses cleanly (statement boundary), the candidates
//     are the tokens that can start a new statement (FIRST set).
//
// Keywords and literal tokens are returned as insertable text
// (IsKeyword=true); free-form tokens (numbers, identifiers, strings) are
// returned as placeholders with their pattern in Detail.
func (d *DSL) Completions(text string, offset int) []Completion {
	if offset < 0 {
		offset = 0
	}
	if offset > len(text) {
		offset = len(text)
	}
	prefix := text[:offset]

	startRule := d.grammar.startRule
	if startRule == "" {
		return nil
	}

	expected := d.expectedAt(prefix, startRule)
	if len(expected) == 0 {
		// Statement boundary (or empty input): suggest statement starters.
		for tok := range firstTokens(d.grammar, startRule) {
			expected = append(expected, tok)
		}
	}

	sort.Strings(expected)
	completions := make([]Completion, 0, len(expected))
	for _, tokName := range expected {
		tok, ok := d.grammar.tokens[tokName]
		if !ok {
			continue
		}
		if literal, isLiteral := literalText(tok.pattern); isLiteral {
			completions = append(completions, Completion{
				Label:     literal,
				Token:     tokName,
				IsKeyword: true,
				Detail:    "keyword",
			})
		} else {
			completions = append(completions, Completion{
				Label:  tokName,
				Token:  tokName,
				Detail: "pattern: " + tok.pattern,
			})
		}
	}
	return completions
}

// expectedAt parses the prefix statement by statement and returns the
// expected tokens at its end — empty when the prefix ends exactly at a
// statement boundary.
func (d *DSL) expectedAt(prefix, startRule string) []string {
	tokens, _ := tokenizeTolerant(d.grammar, prefix)
	parser := newASTParser(d.grammar)
	parser.input = prefix
	parser.tokens = tokens
	sync := firstTokens(d.grammar, startRule)

	pos := 0
	rounds := 0
	for pos < len(tokens) && rounds < maxDiagnostics {
		rounds++
		parser.pos = pos
		parser.failure = parseFailure{tokenPos: -1}
		parser.ruleStack = nil
		parser.hintStack = nil

		_, err := parser.parseRule(startRule)
		if err == nil {
			if parser.pos <= pos {
				pos++
			} else {
				pos = parser.pos
			}
			continue
		}

		// Failure at or beyond the end of the prefix: these expectations
		// are exactly what could come next at the cursor.
		if parser.failure.tokenPos >= len(tokens) {
			return parser.failure.expected
		}

		// Failure strictly inside the prefix: recover and keep going, the
		// cursor is further right.
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
	return nil
}

// literalText reports whether a token pattern matches exactly one literal
// text (keywords `(?i)\bword\b` and QuoteMeta-style literals), returning
// that text.
func literalText(pattern string) (string, bool) {
	// KeywordToken pattern.
	if strings.HasPrefix(pattern, `(?i)\b`) && strings.HasSuffix(pattern, `\b`) {
		word := strings.TrimSuffix(strings.TrimPrefix(pattern, `(?i)\b`), `\b`)
		if word != "" && regexp.QuoteMeta(word) == word {
			return word, true
		}
		return "", false
	}
	// Literal token: unescaping the pattern and re-quoting must round-trip.
	unescaped := strings.NewReplacer(
		`\+`, `+`, `\*`, `*`, `\?`, `?`, `\(`, `(`, `\)`, `)`,
		`\[`, `[`, `\]`, `]`, `\{`, `{`, `\}`, `}`, `\.`, `.`,
		`\^`, `^`, `\$`, `$`, `\|`, `|`, `\\`, `\`,
	).Replace(pattern)
	if unescaped != "" && regexp.QuoteMeta(unescaped) == pattern {
		return unescaped, true
	}
	return "", false
}
