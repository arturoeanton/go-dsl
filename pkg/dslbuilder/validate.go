// Package dslbuilder - Grammar validation.
//
// Validate performs static analysis of a DSL definition and reports
// structural errors (which make the grammar unusable) and warnings
// (which usually indicate mistakes). Build() runs it automatically;
// cmd/validator exposes it on the command line.
package dslbuilder

import (
	"errors"
	"fmt"
	"sort"
)

// Validate checks the DSL definition for common problems.
func (d *DSL) Validate() ([]string, error) {
	var warnings []string
	var errs []error

	g := d.grammar

	d.mu.RLock()
	errs = append(errs, d.deferredErrors...)
	d.mu.RUnlock()

	if len(g.rules) == 0 && len(g.exprRules) == 0 {
		errs = append(errs, errors.New("grammar has no rules"))
		return warnings, errors.Join(errs...)
	}
	if g.startRule == "" {
		errs = append(errs, errors.New("grammar has no start rule"))
	} else if !g.isSymbolDefined(g.startRule) {
		errs = append(errs, fmt.Errorf("grammar start rule %s is not defined", g.startRule))
	}

	// Unknown symbols or skip tokens in rule sequences.
	for _, name := range sortedRuleNames(g) {
		rule := g.rules[name]
		for _, alt := range rule.alternatives {
			for _, symbol := range alt.sequence {
				if tok, ok := g.tokens[symbol]; ok && tok.skip {
					errs = append(errs, fmt.Errorf("rule %s references skipped token %s (SkipToken cannot be matched in grammar rules)", name, symbol))
					continue
				}
				if !g.isSymbolDefined(symbol) {
					errs = append(errs, fmt.Errorf("rule %s references unknown symbol %s", name, symbol))
				}
			}
			if alt.action != "" && !d.hasAction(alt.action) {
				warnings = append(warnings, fmt.Sprintf("rule %s references unregistered action %q", name, alt.action))
			}
			if alt.precedence != 0 {
				warnings = append(warnings, fmt.Sprintf("rule %s declares precedence %d, but generic rules do not use precedence to reorder parses; use Expression() for operator precedence", name, alt.precedence))
			}
		}
	}

	// Expression rule references & unregistered actions.
	for _, name := range sortedExprRuleNames(g) {
		spec := g.exprRules[name]
		checkTok := func(tok string) {
			if t, ok := g.tokens[tok]; ok && t.skip {
				errs = append(errs, fmt.Errorf("expression rule %s references skipped token %s", name, tok))
				return
			}
			if !g.isToken(tok) {
				errs = append(errs, fmt.Errorf("expression rule %s references unknown token %s", name, tok))
			}
		}
		checkAction := func(act string) {
			if act != "" && !d.hasAction(act) {
				warnings = append(warnings, fmt.Sprintf("expression rule %s references unregistered action %q", name, act))
			}
		}

		for _, a := range spec.atoms {
			checkTok(a.token)
			checkAction(a.action)
		}
		for _, gr := range spec.groups {
			checkTok(gr.open)
			checkTok(gr.close)
			if gr.inner != name && !g.isSymbolDefined(gr.inner) {
				errs = append(errs, fmt.Errorf("expression rule %s group references unknown symbol %s", name, gr.inner))
			}
			checkAction(gr.action)
		}

		for _, tok := range sortedMapKeys(spec.prefix) {
			checkTok(tok)
			checkAction(spec.prefix[tok].action)
		}
		for _, tok := range sortedMapKeys(spec.postfix) {
			checkTok(tok)
			checkAction(spec.postfix[tok].action)
		}
		for _, tok := range sortedMapKeys(spec.infix) {
			checkTok(tok)
			checkAction(spec.infix[tok].action)
		}
		for _, tok := range sortedCallKeys(spec.calls) {
			c := spec.calls[tok]
			checkTok(c.open)
			checkTok(c.close)
			if c.sep != "" {
				checkTok(c.sep)
			}
			checkAction(c.action)
		}
		for _, tok := range sortedIndexKeys(spec.indices) {
			idx := spec.indices[tok]
			checkTok(idx.open)
			checkTok(idx.close)
			if idx.inner != "" && idx.inner != name && !g.isSymbolDefined(idx.inner) {
				errs = append(errs, fmt.Errorf("expression rule %s index references unknown symbol %s", name, idx.inner))
			}
			checkAction(idx.action)
		}
		for _, tok := range sortedMemberKeys(spec.members) {
			m := spec.members[tok]
			checkTok(m.dotToken)
			checkTok(m.fieldToken)
			checkAction(m.action)
		}
		if spec.ternary != nil {
			checkTok(spec.ternary.questionToken)
			checkTok(spec.ternary.colonToken)
			checkAction(spec.ternary.action)
		}

		if len(spec.atoms) == 0 && len(spec.groups) == 0 {
			errs = append(errs, fmt.Errorf("expression rule %s has no atoms or groups; it can never match", name))
		}
	}

	// Ordered-choice shadowing.
	for _, name := range sortedRuleNames(g) {
		rule := g.rules[name]
		for i := 0; i < len(rule.alternatives); i++ {
			for j := i + 1; j < len(rule.alternatives); j++ {
				short, long := rule.alternatives[i], rule.alternatives[j]
				if len(long.sequence) > 0 && long.sequence[0] == name {
					continue // left recursion grows past the short match
				}
				if len(short.sequence) >= len(long.sequence) {
					if sequencesEqual(short.sequence, long.sequence) {
						warnings = append(warnings, fmt.Sprintf(
							"rule %s: alternatives %d and %d are identical (%v); the second can never match",
							name, i+1, j+1, short.sequence))
					}
					continue
				}
				if sequencesEqual(short.sequence, long.sequence[:len(short.sequence)]) {
					warnings = append(warnings, fmt.Sprintf(
						"rule %s: alternative %d %v is a prefix of later alternative %d %v; with ordered choice the shorter one matches first and the longer can never match — declare the longer alternative first",
						name, i+1, short.sequence, j+1, long.sequence))
				}
			}
		}
	}

	// Unreachable rules.
	if g.startRule != "" {
		reachable := reachableRules(g)
		for _, name := range sortedRuleNames(g) {
			if !reachable[name] {
				warnings = append(warnings, fmt.Sprintf("rule %s is unreachable from start rule %s", name, g.startRule))
			}
		}
		for _, name := range sortedExprRuleNames(g) {
			if !reachable[name] {
				warnings = append(warnings, fmt.Sprintf("expression rule %s is unreachable from start rule %s", name, g.startRule))
			}
		}
	}

	// Indirect left recursion.
	for _, cycle := range indirectLeftRecursionCycles(g) {
		warnings = append(warnings, fmt.Sprintf("indirect left recursion detected (%s); supported, but consider restructuring the grammar or using Expression() for clarity and performance", cycle))
	}

	// Non-productive rules (including self-grouped expression rules with no atoms).
	productive := productiveRules(g)
	for _, name := range sortedRuleNames(g) {
		if !productive[name] {
			warnings = append(warnings, fmt.Sprintf("rule %s can never derive a terminal string (non-productive cycle)", name))
		}
	}
	for _, name := range sortedExprRuleNames(g) {
		if !productive[name] {
			warnings = append(warnings, fmt.Sprintf("expression rule %s can never derive a terminal string (non-productive cycle)", name))
		}
	}

	return warnings, errors.Join(errs...)
}

func sortedMapKeys(m map[string]exprOp) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func sortedCallKeys(m map[string]exprCall) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func sortedIndexKeys(m map[string]exprIndex) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func sortedMemberKeys(m map[string]exprMember) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// sequencesEqual reports whether two symbol sequences are identical.
func sequencesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// hasAction reports whether an action name is registered as Action, NodeAction,
// or ContextAction.
func (d *DSL) hasAction(name string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if _, ok := d.actions[name]; ok {
		return true
	}
	if _, ok := d.nodeActions[name]; ok {
		return true
	}
	_, ok := d.contextActions[name]
	return ok
}

func sortedRuleNames(g *Grammar) []string {
	names := make([]string, 0, len(g.rules))
	for name := range g.rules {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func sortedExprRuleNames(g *Grammar) []string {
	names := make([]string, 0, len(g.exprRules))
	for name := range g.exprRules {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// reachableRules computes the set of rules reachable from the start rule.
func reachableRules(g *Grammar) map[string]bool {
	reachable := make(map[string]bool)
	var visit func(name string)
	visit = func(name string) {
		if reachable[name] {
			return
		}
		if rule, ok := g.rules[name]; ok {
			reachable[name] = true
			for _, alt := range rule.alternatives {
				for _, symbol := range alt.sequence {
					if !g.isToken(symbol) {
						visit(symbol)
					}
				}
			}
			return
		}
		if spec, ok := g.exprRules[name]; ok {
			reachable[name] = true
			for _, gr := range spec.groups {
				if !g.isToken(gr.inner) {
					visit(gr.inner)
				}
			}
			for _, idx := range spec.indices {
				if idx.inner != "" && !g.isToken(idx.inner) {
					visit(idx.inner)
				}
			}
		}
	}
	visit(g.startRule)
	return reachable
}

// indirectLeftRecursionCycles finds cycles of length > 1 in the
// "leftmost symbol" graph: rule -> first symbol of each alternative.
// Direct left recursion (self loops) is supported and excluded.
func indirectLeftRecursionCycles(g *Grammar) []string {
	edges := make(map[string][]string)
	for name, rule := range g.rules {
		seen := make(map[string]bool)
		for _, alt := range rule.alternatives {
			if len(alt.sequence) == 0 {
				continue
			}
			first := alt.sequence[0]
			if first == name {
				continue // direct left recursion is supported
			}
			if _, isRule := g.rules[first]; isRule && !seen[first] {
				edges[name] = append(edges[name], first)
				seen[first] = true
			}
		}
	}

	var cycles []string
	seenCycle := make(map[string]bool)

	const (
		white = 0
		gray  = 1
		black = 2
	)
	color := make(map[string]int)
	var stack []string

	var dfs func(name string)
	dfs = func(name string) {
		color[name] = gray
		stack = append(stack, name)
		for _, next := range edges[name] {
			if color[next] == gray {
				start := -1
				for i, s := range stack {
					if s == next {
						start = i
						break
					}
				}
				if start >= 0 {
					cycle := append(append([]string(nil), stack[start:]...), next)
					members := append([]string(nil), stack[start:]...)
					sort.Strings(members)
					key := fmt.Sprint(members)
					if !seenCycle[key] {
						seenCycle[key] = true
						cycles = append(cycles, joinCycle(cycle))
					}
				}
			} else if color[next] == white {
				dfs(next)
			}
		}
		stack = stack[:len(stack)-1]
		color[name] = black
	}

	names := make([]string, 0, len(g.rules))
	for name := range g.rules {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if color[name] == white {
			dfs(name)
		}
	}
	sort.Strings(cycles)
	return cycles
}

func joinCycle(cycle []string) string {
	out := ""
	for i, s := range cycle {
		if i > 0 {
			out += " -> "
		}
		out += s
	}
	return out
}

// productiveRules computes, via fixpoint iteration, which rules can derive
// at least one terminal string. Expression rules require either at least one
// atom or at least one group whose inner symbol is itself productive.
func productiveRules(g *Grammar) map[string]bool {
	productive := make(map[string]bool)

	changed := true
	for changed {
		changed = false
		for name, spec := range g.exprRules {
			if productive[name] {
				continue
			}
			if len(spec.atoms) > 0 {
				productive[name] = true
				changed = true
				continue
			}
			for _, gr := range spec.groups {
				if g.isToken(gr.inner) || productive[gr.inner] {
					productive[name] = true
					changed = true
					break
				}
			}
		}
		for name, rule := range g.rules {
			if productive[name] {
				continue
			}
			for _, alt := range rule.alternatives {
				all := true
				for _, symbol := range alt.sequence {
					if g.isToken(symbol) {
						continue
					}
					if !productive[symbol] {
						all = false
						break
					}
				}
				if all {
					productive[name] = true
					changed = true
					break
				}
			}
		}
	}
	return productive
}
