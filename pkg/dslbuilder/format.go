// Package dslbuilder - Universal AST-guided source code formatter.
package dslbuilder

import (
	"fmt"
	"strings"
)

// RuleLayout provides optional formatting hints for a grammar rule.
type RuleLayout struct {
	IndentBlock   bool     `yaml:"indent_block,omitempty" json:"indent_block,omitempty"`       // Indent children between opening and closing tokens
	NewlineAfter  []string `yaml:"newline_after,omitempty" json:"newline_after,omitempty"`     // Emit a newline after these token types
	NoSpaceBefore []string `yaml:"no_space_before,omitempty" json:"no_space_before,omitempty"` // Suppress space before these token types
	NoSpaceAfter  []string `yaml:"no_space_after,omitempty" json:"no_space_after,omitempty"`   // Suppress space after these token types
}

// SetRuleLayout attaches optional formatting layout metadata to a rule.
func (d *DSL) SetRuleLayout(ruleName string, layout RuleLayout) {
	if d.grammar.frozen {
		d.addDeferredError(fmt.Errorf("grammar is frozen (Build() was called); cannot set layout for rule %s", ruleName))
		return
	}
	if d.grammar.layouts == nil {
		d.grammar.layouts = make(map[string]RuleLayout)
	}
	d.grammar.layouts[ruleName] = layout
}

// WithRuleLayout attaches a RuleLayout and returns the DSL for chaining.
func (d *DSL) WithRuleLayout(ruleName string, layout RuleLayout) *DSL {
	d.SetRuleLayout(ruleName, layout)
	return d
}

type fmtAnnotations struct {
	indentAfter   map[int]int  // keyed by token Start: delta indent after this token
	indentBefore  map[int]int  // keyed by token Start: delta indent before this token
	newlineAfter  map[int]bool // keyed by token Start
	newlineBefore map[int]bool // keyed by token Start
	noSpaceBefore map[int]bool // keyed by token Start
	noSpaceAfter  map[int]bool // keyed by token Start
}

func newFmtAnnotations() *fmtAnnotations {
	return &fmtAnnotations{
		indentAfter:   make(map[int]int),
		indentBefore:  make(map[int]int),
		newlineAfter:  make(map[int]bool),
		newlineBefore: make(map[int]bool),
		noSpaceBefore: make(map[int]bool),
		noSpaceAfter:  make(map[int]bool),
	}
}

func firstLeafToken(n *Node) *TokenMatch {
	if n == nil {
		return nil
	}
	if n.Token != nil {
		return n.Token
	}
	for _, ch := range n.Children {
		if t := firstLeafToken(ch); t != nil {
			return t
		}
	}
	return nil
}

func lastLeafToken(n *Node) *TokenMatch {
	if n == nil {
		return nil
	}
	if n.Token != nil {
		return n.Token
	}
	for i := len(n.Children) - 1; i >= 0; i-- {
		if t := lastLeafToken(n.Children[i]); t != nil {
			return t
		}
	}
	return nil
}

func (d *DSL) collectFmtAnnotations(n *Node, ann *fmtAnnotations) {
	if n == nil || n.Token != nil {
		return
	}

	// Pratt Call node: callee(arg1, arg2) -> no space before '('
	if n.callStep > 0 && len(n.Children) >= 3 {
		if openTok := firstLeafToken(n.Children[1]); openTok != nil {
			ann.noSpaceBefore[openTok.Start] = true
		}
	}

	// Pratt Index node: target[index] -> no space before '['
	if len(n.Children) == 4 && n.Children[1] != nil && n.Children[1].Token != nil && n.Children[1].Token.Value == "[" {
		ann.noSpaceBefore[n.Children[1].Token.Start] = true
	}

	if layout, ok := d.grammar.layouts[n.Rule]; ok {
		if layout.IndentBlock && len(n.Children) >= 2 {
			openTok := lastLeafToken(n.Children[0])
			closeTok := firstLeafToken(n.Children[len(n.Children)-1])
			if openTok != nil && closeTok != nil && openTok.Start < closeTok.Start {
				ann.indentAfter[openTok.Start]++
				ann.newlineAfter[openTok.Start] = true
				ann.indentBefore[closeTok.Start]--
				ann.newlineBefore[closeTok.Start] = true
			}
		}

		nlSet := toBoolSet(layout.NewlineAfter)
		nsBefore := toBoolSet(layout.NoSpaceBefore)
		nsAfter := toBoolSet(layout.NoSpaceAfter)

		for _, ch := range n.Children {
			if ch != nil && ch.Token != nil {
				tt := ch.Token.TokenType
				if nlSet[tt] {
					ann.newlineAfter[ch.Token.Start] = true
				}
				if nsBefore[tt] {
					ann.noSpaceBefore[ch.Token.Start] = true
				}
				if nsAfter[tt] {
					ann.noSpaceAfter[ch.Token.Start] = true
				}
			}
		}
	}

	for _, ch := range n.Children {
		d.collectFmtAnnotations(ch, ann)
	}
}

func toBoolSet(items []string) map[string]bool {
	if len(items) == 0 {
		return nil
	}
	m := make(map[string]bool, len(items))
	for _, s := range items {
		m[s] = true
	}
	return m
}

// Format canonically formats DSL source code using the grammar's AST, layout
// rules, and trivia comments. Returns an error if the code has lexical or
// syntax errors, guaranteeing invalid code is never corrupted.
func (d *DSL) Format(code string) (string, error) {
	tokens, trivia, err := d.TokenizeWithTrivia(code)
	if err != nil {
		return "", err
	}
	if len(tokens) == 0 && len(trivia) == 0 {
		return "", nil
	}

	ann := newFmtAnnotations()
	if len(tokens) > 0 {
		if root, parseErr := d.ParseAST(code); parseErr == nil {
			d.collectFmtAnnotations(root, ann)
		} else {
			diags := d.Diagnostics(code)
			if len(diags) > 0 {
				return "", diags[0]
			}
			doc := d.NewDocument()
			doc.Update(code)
			stmts := doc.Statements()
			for i, st := range stmts {
				d.collectFmtAnnotations(st, ann)
				if i+1 < len(stmts) {
					if endTok := lastLeafToken(st); endTok != nil {
						ann.newlineAfter[endTok.Start] = true
					}
				}
			}
		}
	}

	lineOf := func(offset int) int {
		ln, _ := calculateLineColumn(code, offset)
		return ln
	}

	var out strings.Builder
	indent := 0
	atLineStart := true
	prevEnd := 0
	prevLine := 1
	var prevTok *TokenMatch

	writeIndent := func() {
		if indent > 0 {
			out.WriteString(strings.Repeat("  ", indent))
		}
	}

	emitNewline := func() {
		if !atLineStart {
			out.WriteByte('\n')
			atLineStart = true
		}
	}

	trivIdx := 0
	emitTriviaBefore := func(limitOffset int) {
		for trivIdx < len(trivia) && trivia[trivIdx].Start < limitOffset {
			tr := trivia[trivIdx]
			trivIdx++
			trLine := lineOf(tr.Start)

			if !atLineStart {
				if trLine > prevLine {
					emitNewline()
					if trLine-prevLine > 1 {
						out.WriteByte('\n')
					}
					writeIndent()
				} else {
					out.WriteByte(' ')
				}
			} else {
				if prevEnd > 0 && trLine-prevLine > 1 {
					out.WriteByte('\n')
				}
				writeIndent()
			}

			out.WriteString(strings.TrimRight(tr.Value, " \t\r"))
			atLineStart = false
			prevEnd = tr.End
			prevLine, _ = calculateLineColumn(code, tr.End)

			// Line comments always terminate the current line.
			if !strings.HasPrefix(tr.Value, "/*") {
				emitNewline()
			}
			prevTok = nil
		}
	}

	for i := range tokens {
		tok := tokens[i]
		emitTriviaBefore(tok.Start)

		if tok.TokenType == "NEWLINE" {
			emitNewline()
			prevEnd = tok.End
			prevLine, _ = calculateLineColumn(code, tok.End)
			prevTok = nil
			continue
		}

		if delta := ann.indentBefore[tok.Start]; delta != 0 {
			indent += delta
			if indent < 0 {
				indent = 0
			}
		}
		if ann.newlineBefore[tok.Start] {
			emitNewline()
		}

		tokLine := lineOf(tok.Start)
		if atLineStart {
			if prevEnd > 0 && tokLine-prevLine > 1 {
				out.WriteByte('\n')
			}
			writeIndent()
		} else if prevTok != nil {
			if shouldInsertSpace(*prevTok, tok, ann) {
				out.WriteByte(' ')
			}
		} else {
			out.WriteByte(' ')
		}

		out.WriteString(tok.Value)
		atLineStart = false
		prevEnd = tok.End
		prevLine, _ = calculateLineColumn(code, tok.End)
		tokCopy := tok
		prevTok = &tokCopy

		if delta := ann.indentAfter[tok.Start]; delta != 0 {
			indent += delta
			if indent < 0 {
				indent = 0
			}
		}
		if ann.newlineAfter[tok.Start] {
			// Wait to emit newline only if an immediate trailing comment sits on
			// this same line BEFORE the next grammar token!
			hasSameLineTrivia := trivIdx < len(trivia) &&
				(i+1 >= len(tokens) || trivia[trivIdx].Start < tokens[i+1].Start) &&
				lineOf(trivia[trivIdx].Start) == prevLine
			if !hasSameLineTrivia {
				emitNewline()
				prevTok = nil
			}
		}
	}

	emitTriviaBefore(len(code) + 1)

	formatted := strings.TrimRight(out.String(), " \t\n")
	if strings.HasSuffix(code, "\n") && formatted != "" {
		formatted += "\n"
	}
	return formatted, nil
}

func shouldInsertSpace(prev, curr TokenMatch, ann *fmtAnnotations) bool {
	if ann.noSpaceAfter[prev.Start] || ann.noSpaceBefore[curr.Start] {
		return false
	}
	pv := prev.Value
	cv := curr.Value

	// Opening brackets/parens or member dot: no space after.
	if pv == "(" || pv == "[" || pv == "." {
		return false
	}
	// Closing brackets/parens, comma, semicolon, member dot: no space before.
	if cv == ")" || cv == "]" || cv == "," || cv == ";" || cv == "." {
		return false
	}
	return true
}
