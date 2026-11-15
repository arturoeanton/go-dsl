// Package dslbuilder - High-level fluent API.
//
// This layer removes the need to know formal grammar concepts for common
// cases. A complete calculator:
//
//	calc, err := dslbuilder.New("calc").
//	    Tokens(func(t *dslbuilder.TokenSet) {
//	        t.Regex("NUMBER", `\d+`)
//	        t.Literal("PLUS", "+")
//	        t.Literal("STAR", "*")
//	        t.Literal("LPAREN", "(")
//	        t.Literal("RPAREN", ")")
//	    }).
//	    Expr("expr", func(e *dslbuilder.ExpressionBuilder) {
//	        e.Atom("NUMBER", "number")
//	        e.Group("LPAREN", "expr", "RPAREN")
//	        e.InfixLeft("PLUS", 10, "add")
//	        e.InfixLeft("STAR", 20, "mul")
//	    }).
//	    WithAction("number", dslbuilder.Action1(strconv.Atoi)).
//	    WithAction("add", dslbuilder.Action3(func(l int, _ string, r int) (int, error) { return l + r, nil })).
//	    WithAction("mul", dslbuilder.Action3(func(l int, _ string, r int) (int, error) { return l * r, nil })).
//	    Build()
package dslbuilder

import (
	"fmt"
	"regexp"
)

// TokenSet declares tokens inside a Tokens() block. Errors (invalid
// regexes, empty-matching patterns, frozen grammar) are deferred and
// surface in Validate()/Build().
type TokenSet struct {
	dsl *DSL
}

// Regex declares a token from a regular expression.
func (t *TokenSet) Regex(name, pattern string) *TokenSet {
	if err := t.dsl.Token(name, pattern); err != nil {
		t.dsl.addDeferredError(err)
	}
	return t
}

// Literal declares a token that matches an exact text (no regex semantics):
// t.Literal("PLUS", "+") matches the character '+' literally.
func (t *TokenSet) Literal(name, text string) *TokenSet {
	if err := t.dsl.LiteralToken(name, text); err != nil {
		t.dsl.addDeferredError(err)
	}
	return t
}

// Keyword declares a case-insensitive, word-bounded keyword token with
// high matching priority (same as KeywordToken).
func (t *TokenSet) Keyword(name, keyword string) *TokenSet {
	if err := t.dsl.KeywordToken(name, keyword); err != nil {
		t.dsl.addDeferredError(err)
	}
	return t
}

// Skip declares a token whose matches are skipped during normal parsing and
// captured as trivia in TokenizeWithTrivia / Format.
func (t *TokenSet) Skip(name, pattern string) *TokenSet {
	if err := t.dsl.SkipToken(name, pattern); err != nil {
		t.dsl.addDeferredError(err)
	}
	return t
}

// LineComment declares a line-comment trivia token starting with prefix
// (for example, "//" or "#") and running until the end of the line.
func (t *TokenSet) LineComment(prefix string) *TokenSet {
	if prefix == "" {
		t.dsl.addDeferredError(fmt.Errorf("line comment prefix cannot be empty"))
		return t
	}
	name := "LINE_COMMENT"
	if _, exists := t.dsl.grammar.tokens[name]; exists {
		name = "LINE_COMMENT_" + regexp.QuoteMeta(prefix)
	}
	pattern := regexp.QuoteMeta(prefix) + `[^\r\n]*`
	return t.Skip(name, pattern)
}

// BlockComment declares a delimited block-comment trivia token (for example,
// open="/*", close="*/") that may span multiple lines.
func (t *TokenSet) BlockComment(open, close string) *TokenSet {
	if open == "" || close == "" {
		t.dsl.addDeferredError(fmt.Errorf("block comment delimiters cannot be empty"))
		return t
	}
	name := "BLOCK_COMMENT"
	if _, exists := t.dsl.grammar.tokens[name]; exists {
		name = "BLOCK_COMMENT_" + regexp.QuoteMeta(open)
	}
	pattern := `(?s)` + regexp.QuoteMeta(open) + `.*?` + regexp.QuoteMeta(close)
	return t.Skip(name, pattern)
}

// Tokens declares a batch of tokens with a TokenSet and returns the DSL
// for chaining. Errors are deferred to Validate()/Build().
func (d *DSL) Tokens(fn func(t *TokenSet)) *DSL {
	fn(&TokenSet{dsl: d})
	return d
}

// Expr declares a Pratt-parsed expression rule with a builder callback
// and returns the DSL for chaining. Equivalent to configuring
// d.Expression(name) inside the callback.
func (d *DSL) Expr(name string, fn func(e *ExpressionBuilder)) *DSL {
	fn(d.Expression(name))
	return d
}
