// Package dslbuilder - AI Guardrails: GBNF constrained-decoding exporter,
// LLM system prompt specification generator, and diagnostic repair prompts.
package dslbuilder

import (
	"fmt"
	"sort"
	"strings"
)

// ToGBNF exports the DSL grammar in GBNF (GGML BNF) format compatible with
// llama.cpp, Ollama, and LM Studio for 100% syntax-constrained local LLM decoding.
func (d *DSL) ToGBNF() string {
	g := d.grammar
	if g.startRule == "" {
		return "root ::= \"\"\n"
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "root ::= ws %s ws\n", gbnfIdent(g.startRule))

	// Emit standard rules in deterministic order.
	for _, name := range sortedRuleNames(g) {
		rule := g.rules[name]
		if len(rule.alternatives) == 0 {
			continue
		}
		alts := make([]string, 0, len(rule.alternatives))
		for _, alt := range rule.alternatives {
			if len(alt.sequence) == 0 {
				alts = append(alts, `""`)
				continue
			}
			parts := make([]string, 0, len(alt.sequence)*2)
			for i, sym := range alt.sequence {
				if i > 0 {
					parts = append(parts, "ws")
				}
				parts = append(parts, gbnfSymbolRef(g, sym))
			}
			alts = append(alts, strings.Join(parts, " "))
		}
		fmt.Fprintf(&sb, "%s ::= %s\n", gbnfIdent(name), strings.Join(alts, " | "))
	}

	// Emit Pratt expression rules lowered into precedence levels.
	for _, name := range sortedExprRuleNames(g) {
		spec := g.exprRules[name]
		emitExprGBNF(&sb, g, spec)
	}

	// Emit token productions in declaration order.
	var skipAlts []string
	for _, tok := range g.tokenList {
		if tok.skip {
			skipAlts = append(skipAlts, regexToGBNF(tok))
			continue
		}
		fmt.Fprintf(&sb, "%s ::= %s\n", gbnfTokenIdent(tok.name), regexToGBNF(tok))
	}
	if g.significantNewlines {
		if _, explicit := g.tokens["NEWLINE"]; !explicit {
			sb.WriteString("tok-newline ::= \"\\r\"? \"\\n\"\n")
		}
	}

	// Whitespace + skip trivia rule.
	wsChar := `[ \t\n\r]`
	if g.significantNewlines {
		wsChar = `[ \t]`
	}
	if len(skipAlts) > 0 {
		fmt.Fprintf(&sb, "ws ::= (%s | %s)*\n", wsChar, strings.Join(skipAlts, " | "))
	} else {
		fmt.Fprintf(&sb, "ws ::= %s*\n", wsChar)
	}

	return sb.String()
}

func emitExprGBNF(sb *strings.Builder, g *Grammar, spec *exprSpec) {
	base := gbnfIdent(spec.name)
	primaryName := base + "-primary"
	postfixName := base + "-postfix"

	// Collect distinct infix/ternary precedence levels ascending.
	powerSet := make(map[int]bool)
	for _, op := range spec.infix {
		powerSet[op.power] = true
	}
	if spec.ternary != nil {
		powerSet[spec.ternary.power] = true
	}
	powers := make([]int, 0, len(powerSet))
	for p := range powerSet {
		powers = append(powers, p)
	}
	sort.Ints(powers)

	if len(powers) == 0 {
		fmt.Fprintf(sb, "%s ::= %s\n", base, postfixName)
	} else {
		fmt.Fprintf(sb, "%s ::= %s-lvl0\n", base, base)
		for idx, p := range powers {
			currLvl := fmt.Sprintf("%s-lvl%d", base, idx)
			nextLvl := postfixName
			if idx+1 < len(powers) {
				nextLvl = fmt.Sprintf("%s-lvl%d", base, idx+1)
			}

			var alts []string
			if spec.ternary != nil && spec.ternary.power == p {
				alts = append(alts, fmt.Sprintf("%s ws %s ws %s ws %s ws %s",
					nextLvl,
					gbnfSymbolRef(g, spec.ternary.questionToken),
					base,
					gbnfSymbolRef(g, spec.ternary.colonToken),
					currLvl))
			}

			var leftOps, rightOps, nonAssocOps []string
			for _, tok := range sortedMapKeys(spec.infix) {
				op := spec.infix[tok]
				if op.power != p {
					continue
				}
				ref := gbnfSymbolRef(g, op.token)
				if op.rightAssoc {
					rightOps = append(rightOps, ref)
				} else if op.nonAssoc {
					nonAssocOps = append(nonAssocOps, ref)
				} else {
					leftOps = append(leftOps, ref)
				}
			}

			if len(leftOps) > 0 {
				alts = append(alts, fmt.Sprintf("%s (ws (%s) ws %s)*", nextLvl, strings.Join(leftOps, " | "), nextLvl))
			}
			if len(rightOps) > 0 {
				alts = append(alts, fmt.Sprintf("%s ws (%s) ws %s", nextLvl, strings.Join(rightOps, " | "), currLvl))
			}
			if len(nonAssocOps) > 0 {
				alts = append(alts, fmt.Sprintf("%s ws (%s) ws %s", nextLvl, strings.Join(nonAssocOps, " | "), nextLvl))
			}
			if len(leftOps) == 0 {
				alts = append(alts, nextLvl)
			}
			fmt.Fprintf(sb, "%s ::= %s\n", currLvl, strings.Join(alts, " | "))
		}
	}

	// Postfix suffixes: calls, indices, members, unary postfix.
	var suffixes []string
	for _, tok := range sortedCallKeys(spec.calls) {
		c := spec.calls[tok]
		openRef := gbnfSymbolRef(g, c.open)
		closeRef := gbnfSymbolRef(g, c.close)
		if c.sep != "" {
			sepRef := gbnfSymbolRef(g, c.sep)
			suffixes = append(suffixes, fmt.Sprintf("%s ws (%s (ws %s ws %s)* ws)? %s", openRef, base, sepRef, base, closeRef))
		} else {
			suffixes = append(suffixes, fmt.Sprintf("%s ws (%s ws)* %s", openRef, base, closeRef))
		}
	}
	for _, tok := range sortedIndexKeys(spec.indices) {
		idx := spec.indices[tok]
		innerRule := idx.inner
		if innerRule == "" {
			innerRule = spec.name
		}
		suffixes = append(suffixes, fmt.Sprintf("%s ws %s ws %s",
			gbnfSymbolRef(g, idx.open), gbnfSymbolRef(g, innerRule), gbnfSymbolRef(g, idx.close)))
	}
	for _, tok := range sortedMemberKeys(spec.members) {
		m := spec.members[tok]
		suffixes = append(suffixes, fmt.Sprintf("%s ws %s",
			gbnfSymbolRef(g, m.dotToken), gbnfSymbolRef(g, m.fieldToken)))
	}
	for _, tok := range sortedMapKeys(spec.postfix) {
		suffixes = append(suffixes, gbnfSymbolRef(g, tok))
	}

	if len(suffixes) > 0 {
		fmt.Fprintf(sb, "%s ::= %s (ws (%s))*\n", postfixName, primaryName, strings.Join(suffixes, " | "))
	} else {
		fmt.Fprintf(sb, "%s ::= %s\n", postfixName, primaryName)
	}

	// Primary: prefix operators, groups, atoms.
	var primAlts []string
	for _, tok := range sortedMapKeys(spec.prefix) {
		primAlts = append(primAlts, fmt.Sprintf("%s ws %s", gbnfSymbolRef(g, tok), postfixName))
	}
	for _, gr := range spec.groups {
		primAlts = append(primAlts, fmt.Sprintf("%s ws %s ws %s",
			gbnfSymbolRef(g, gr.open), gbnfSymbolRef(g, gr.inner), gbnfSymbolRef(g, gr.close)))
	}
	for _, a := range spec.atoms {
		primAlts = append(primAlts, gbnfSymbolRef(g, a.token))
	}
	if len(primAlts) == 0 {
		primAlts = append(primAlts, `""`)
	}
	fmt.Fprintf(sb, "%s ::= %s\n", primaryName, strings.Join(primAlts, " | "))
}

func gbnfSymbolRef(g *Grammar, sym string) string {
	if g.isToken(sym) {
		return gbnfTokenIdent(sym)
	}
	return gbnfIdent(sym)
}

func gbnfTokenIdent(name string) string {
	return "tok-" + gbnfIdent(name)
}

func gbnfIdent(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteByte('-')
		}
	}
	res := strings.Trim(b.String(), "-")
	if res == "" {
		return "sym"
	}
	return res
}

func regexToGBNF(tok *Token) string {
	if tok.kind == tokenKindKeyword && tok.rawText != "" {
		return caseInsensitiveGBNF(tok.rawText)
	}
	if tok.kind == tokenKindLiteral && tok.rawText != "" {
		return gbnfQuote(tok.rawText)
	}
	if lit, ok := literalText(tok.pattern); ok {
		if isKeywordTokenPattern(tok.pattern) {
			return caseInsensitiveGBNF(lit)
		}
		return gbnfQuote(lit)
	}

	p := tok.pattern
	switch p {
	case `[0-9]+`, `\d+`:
		return `[0-9]+`
	case `[0-9]+(\.[0-9]+)?`, `\d+(\.\d+)?`:
		return `[0-9]+ ("." [0-9]+)?`
	case `[a-zA-Z_][a-zA-Z0-9_]*`:
		return `[a-zA-Z_] [a-zA-Z0-9_]*`
	case `[a-zA-Z][a-zA-Z0-9_]*`:
		return `[a-zA-Z] [a-zA-Z0-9_]*`
	case `[a-zA-Z]+`:
		return `[a-zA-Z]+`
	case `[a-z]+`:
		return `[a-z]+`
	case `"[^"]*"`:
		return `"\"" [^"]* "\""`
	case `"([^"\\]|\\.)*"`:
		return `"\"" ([^"\\] | "\\" .)* "\""`
	}

	// Line comment pattern like //[^\r\n]* or #[^\r\n]*
	if strings.HasSuffix(p, `[^\r\n]*`) {
		prefix := strings.TrimSuffix(p, `[^\r\n]*`)
		unescaped, _ := literalText(prefix)
		if unescaped == "" {
			unescaped = prefix
		}
		return fmt.Sprintf("%s [^\\r\\n]*", gbnfQuote(unescaped))
	}

	// Fallback: sanitize common RE2 shorthands for GBNF compatibility.
	replacer := strings.NewReplacer(
		`\d`, `[0-9]`,
		`\w`, `[a-zA-Z0-9_]`,
		`\s`, `[ \t\n\r]`,
	)
	return replacer.Replace(p)
}

func caseInsensitiveGBNF(word string) string {
	parts := make([]string, 0, len(word))
	for _, r := range word {
		lo := strings.ToLower(string(r))
		up := strings.ToUpper(string(r))
		if lo != up && len(lo) == 1 && len(up) == 1 {
			parts = append(parts, fmt.Sprintf("[%s%s]", lo, up))
		} else {
			parts = append(parts, gbnfQuote(string(r)))
		}
	}
	return strings.Join(parts, " ")
}

func gbnfQuote(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

// AIPromptSpec generates a structured system-prompt specification of the DSL
// (tokens, precedence table, rules, GBNF grammar, and valid examples) so LLMs
// can author valid DSL programs with minimal hallucination.
func (d *DSL) AIPromptSpec() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "# DSL Specification: %s\n\n", d.name)
	fmt.Fprintf(&sb, "Start rule: `%s`\n", d.grammar.startRule)
	if d.grammar.significantNewlines {
		sb.WriteString("Newlines are syntactically significant (`NEWLINE`).\n")
	}
	sb.WriteString("\n## Tokens\n")
	for _, tok := range d.grammar.tokenList {
		if tok.skip {
			fmt.Fprintf(&sb, "- `%s` (comment/skipped): `%s`\n", tok.name, tok.pattern)
			continue
		}
		if (tok.kind == tokenKindKeyword || tok.kind == tokenKindLiteral) && tok.rawText != "" {
			fmt.Fprintf(&sb, "- `%s` (%s): `%s`\n", tok.name, tok.kind, tok.rawText)
		} else {
			fmt.Fprintf(&sb, "- `%s` (regex): `%s`\n", tok.name, tok.pattern)
		}
	}

	if len(d.grammar.rules) > 0 {
		sb.WriteString("\n## Grammar Rules\n")
		for _, name := range d.grammar.ruleOrder {
			rule := d.grammar.rules[name]
			for _, alt := range rule.alternatives {
				seq := "ε"
				if len(alt.sequence) > 0 {
					seq = strings.Join(alt.sequence, " ")
				}
				if alt.errHint != "" {
					fmt.Fprintf(&sb, "- `%s` -> %s  _(hint: %s)_\n", name, seq, alt.errHint)
				} else {
					fmt.Fprintf(&sb, "- `%s` -> %s\n", name, seq)
				}
			}
		}
	}

	if len(d.grammar.exprRules) > 0 {
		sb.WriteString("\n## Expression Rules (Pratt Precedence)\n")
		for _, name := range sortedExprRuleNames(d.grammar) {
			spec := d.grammar.exprRules[name]
			fmt.Fprintf(&sb, "### `%s`\n", name)
			for _, tok := range sortedMapKeys(spec.prefix) {
				fmt.Fprintf(&sb, "- Prefix `%s` (power %d)\n", tok, spec.prefix[tok].power)
			}
			for _, tok := range sortedMapKeys(spec.postfix) {
				fmt.Fprintf(&sb, "- Postfix `%s` (power %d)\n", tok, spec.postfix[tok].power)
			}
			for _, tok := range sortedMapKeys(spec.infix) {
				op := spec.infix[tok]
				assoc := "left"
				if op.rightAssoc {
					assoc = "right"
				} else if op.nonAssoc {
					assoc = "non-associative"
				}
				fmt.Fprintf(&sb, "- Infix `%s` (power %d, %s)\n", tok, op.power, assoc)
			}
			if spec.ternary != nil {
				fmt.Fprintf(&sb, "- Ternary `%s ... %s` (power %d, right)\n", spec.ternary.questionToken, spec.ternary.colonToken, spec.ternary.power)
			}
		}
	}

	samples := d.GenerateSamples(42, 3, 3)
	if len(samples) > 0 {
		sb.WriteString("\n## Valid Examples\n")
		for _, ex := range samples {
			fmt.Fprintf(&sb, "```dsl\n%s\n```\n", ex)
		}
	}

	sb.WriteString("\n## GBNF Grammar\n```gbnf\n")
	sb.WriteString(d.ToGBNF())
	sb.WriteString("```\n")
	return sb.String()
}

// RepairPrompt inspects DSL code with Diagnostics + Completions. If the code
// has zero errors, it returns ("", true). Otherwise it returns a targeted
// self-repair prompt containing exact line/column diagnostics and valid next
// tokens at each failure point, plus clean=false.
func (d *DSL) RepairPrompt(code string) (string, bool) {
	diags := d.Diagnostics(code)
	if len(diags) == 0 {
		return "", true
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "The following `%s` DSL code contains %d syntax/lexical error(s):\n\n```dsl\n%s\n```\n\n", d.name, len(diags), code)
	sb.WriteString("### Diagnostics\n")
	for i, pe := range diags {
		fmt.Fprintf(&sb, "%d. %s\n", i+1, pe.DetailedError())
		if pe.Position >= 0 && pe.Position <= len(code) {
			comps := d.Completions(code, pe.Position)
			if len(comps) > 0 {
				labels := make([]string, 0, len(comps))
				for _, c := range comps {
					labels = append(labels, fmt.Sprintf("`%s`", c.Label))
				}
				fmt.Fprintf(&sb, "   Valid tokens here: %s\n", strings.Join(labels, ", "))
			}
		}
	}
	sb.WriteString("\nReturn ONLY the corrected DSL code conforming strictly to the grammar.\n")
	return sb.String(), false
}
