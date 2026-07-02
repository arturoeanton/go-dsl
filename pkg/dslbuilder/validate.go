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
//
// Errors (returned as a joined error, grammar should not be used):
//   - no rules defined / no start rule
//   - a rule references an unknown symbol (not a token, rule, or expression rule)
//   - an expression rule references unknown tokens or an unknown inner rule
//   - errors deferred from fluent builder helpers (e.g. invalid token regex)
//
// Warnings (returned as strings, grammar still works):
//   - an action referenced by a rule is not registered
//   - a rule is unreachable from the start rule
//   - indirect left recursion (NOT supported by the parser; restructure the
//     grammar or use Expression() for operators)
//   - a rule can never derive a terminal string (non-productive cycle)
//   - precedence declared on alternatives (informative: RuleWithPrecedence
//     metadata is not used to reorder parses; use Expression() instead)
func (d *DSL) Validate() ([]string, error) {
	var warnings []string
	var errs []error

	g := d.grammar

	errs = append(errs, d.deferredErrors...)

	if len(g.rules) == 0 && len(g.exprRules) == 0 {
		errs = append(errs, errors.New("grammar has no rules"))
		return warnings, errors.Join(errs...)
	}
	if g.startRule == "" {
		errs = append(errs, errors.New("grammar has no start rule"))
	}

	// Unknown symbols in rule sequences.
	for _, name := range sortedRuleNames(g) {
		rule := g.rules[name]
		for _, alt := range rule.alternatives {
			for _, symbol := range alt.sequence {
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

	// Expression rule references.
	for _, name := range sortedExprRuleNames(g) {
		spec := g.exprRules[name]
		checkTok := func(tok string) {
			if _, ok := g.tokens[tok]; !ok {
				errs = append(errs, fmt.Errorf("expression rule %s references unknown token %s", name, tok))
			}
		}
		for _, a := range spec.atoms {
			checkTok(a.token)
			if a.action != "" && !d.hasAction(a.action) {
				warnings = append(warnings, fmt.Sprintf("expression rule %s references unregistered action %q", name, a.action))
			}
		}
		for _, gr := range spec.groups {
			checkTok(gr.open)
			checkTok(gr.close)
			if gr.inner != name && !g.isSymbolDefined(gr.inner) {
				errs = append(errs, fmt.Errorf("expression rule %s group references unknown symbol %s", name, gr.inner))
			}
		}
		ops := make([]string, 0, len(spec.prefix)+len(spec.infix))
		for tok := range spec.prefix {
			ops = append(ops, tok)
		}
		for tok := range spec.infix {
			ops = append(ops, tok)
		}
		sort.Strings(ops)
		for _, tok := range ops {
			checkTok(tok)
		}
		if len(spec.atoms) == 0 && len(spec.groups) == 0 {
			errs = append(errs, fmt.Errorf("expression rule %s has no atoms or groups; it can never match", name))
		}
	}

	// Ordered-choice shadowing: a shorter alternative declared before a
	// longer one with the same prefix means the longer one can never match
	// (the classic PEG surprise). Left-recursive longer alternatives are
	// exempt: the growing-seed algorithm extends the seed past the short
	// match.
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

	// Indirect left recursion: supported via generalized growing, but
	// memoization is disabled while those rules parse, so grammars are
	// usually clearer and faster restructured or expressed with Expression().
	for _, cycle := range indirectLeftRecursionCycles(g) {
		warnings = append(warnings, fmt.Sprintf("indirect left recursion detected (%s); supported, but consider restructuring the grammar or using Expression() for clarity and performance", cycle))
	}

	// Non-productive rules.
	productive := productiveRules(g)
	for _, name := range sortedRuleNames(g) {
		if !productive[name] {
			warnings = append(warnings, fmt.Sprintf("rule %s can never derive a terminal string (non-productive cycle)", name))
		}
	}

	return warnings, errors.Join(errs...)
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

// hasAction reports whether an action name is registered either as a
// regular Action or as a lazy NodeAction.
func (d *DSL) hasAction(name string) bool {
	if _, ok := d.actions[name]; ok {
		return true
	}
	_, ok := d.nodeActions[name]
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
					if _, isToken := g.tokens[symbol]; !isToken {
						visit(symbol)
					}
				}
			}
			return
		}
		if spec, ok := g.exprRules[name]; ok {
			reachable[name] = true
			for _, gr := range spec.groups {
				if _, isToken := g.tokens[gr.inner]; !isToken {
					visit(gr.inner)
				}
			}
		}
	}
	visit(g.startRule)
	return reachable
}

// indirectLeftRecursionCycles finds cycles of length > 1 in the
// "leftmost symbol" graph: rule → first symbol of each alternative.
// Direct left recursion (self loops) is supported and excluded.
func indirectLeftRecursionCycles(g *Grammar) []string {
	// Build leftmost graph between rules.
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
				// Extract the cycle from the stack.
				start := -1
				for i, s := range stack {
					if s == next {
						start = i
						break
					}
				}
				if start >= 0 {
					cycle := append(append([]string(nil), stack[start:]...), next)
					// Canonical key: sorted members, to dedupe rotations.
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
// at least one terminal string. Expression rules with atoms or groups are
// considered productive.
func productiveRules(g *Grammar) map[string]bool {
	productive := make(map[string]bool)
	for name, spec := range g.exprRules {
		if len(spec.atoms) > 0 || len(spec.groups) > 0 {
			productive[name] = true
		}
	}

	changed := true
	for changed {
		changed = false
		for name, rule := range g.rules {
			if productive[name] {
				continue
			}
			for _, alt := range rule.alternatives {
				all := true
				for _, symbol := range alt.sequence {
					if _, isToken := g.tokens[symbol]; isToken {
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
