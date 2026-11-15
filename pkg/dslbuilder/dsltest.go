// Package dslbuilder - Declarative grammar test runner, sample generator, and fuzzer.
package dslbuilder

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
)

// DSLTestCase defines one declarative grammar test case (typically in YAML/JSON).
type DSLTestCase struct {
	Name         string                 `yaml:"name" json:"name"`
	Input        string                 `yaml:"input" json:"input"`
	ShouldParse  *bool                  `yaml:"should_parse,omitempty" json:"should_parse,omitempty"`
	ExpectOutput interface{}            `yaml:"expect_output,omitempty" json:"expect_output,omitempty"`
	ExpectAST    string                 `yaml:"expect_ast,omitempty" json:"expect_ast,omitempty"`
	ExpectError  string                 `yaml:"expect_error,omitempty" json:"expect_error,omitempty"`
	Context      map[string]interface{} `yaml:"context,omitempty" json:"context,omitempty"`
}

// DSLTestResult records the outcome of running a single DSLTestCase.
type DSLTestResult struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Message string `json:"message,omitempty"`
}

// DSLTestReport summarizes a suite of declarative grammar tests.
type DSLTestReport struct {
	Total   int             `json:"total"`
	Passed  int             `json:"passed"`
	Failed  int             `json:"failed"`
	Results []DSLTestResult `json:"results"`
}

// FuzzFailure records a panic or unexpected failure discovered during grammar fuzzing.
type FuzzFailure struct {
	Iteration int    `json:"iteration"`
	Input     string `json:"input"`
	Reason    string `json:"reason"`
}

// FuzzReport summarizes a grammar fuzzing run.
type FuzzReport struct {
	Iterations      int           `json:"iterations"`
	ValidParsed     int           `json:"valid_parsed"`
	RejectedCleanly int           `json:"rejected_cleanly"`
	Failures        []FuzzFailure `json:"failures,omitempty"`
}

// Tests returns a copy of the declarative test cases loaded with this DSL.
func (d *DSL) Tests() []DSLTestCase {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return append([]DSLTestCase(nil), d.tests...)
}

// RunConfigTests executes the declarative test cases embedded in the DSL's configuration.
func (d *DSL) RunConfigTests() DSLTestReport {
	return d.RunTestSuite(d.Tests())
}

// RunTestSuite runs a slice of declarative DSLTestCases against this DSL.
func (d *DSL) RunTestSuite(cases []DSLTestCase) DSLTestReport {
	report := DSLTestReport{
		Total:   len(cases),
		Results: make([]DSLTestResult, 0, len(cases)),
	}

	for idx, tc := range cases {
		name := tc.Name
		if name == "" {
			name = fmt.Sprintf("case_%d", idx+1)
		}

		shouldParse := true
		if tc.ShouldParse != nil {
			shouldParse = *tc.ShouldParse
		} else if tc.ExpectError != "" {
			shouldParse = false
		}

		d.mu.RLock()
		noActions := len(d.actions) == 0 && len(d.nodeActions) == 0 && len(d.contextActions) == 0
		d.mu.RUnlock()

		var res *Result
		var err error
		if tc.ExpectOutput == nil && noActions {
			var ast *Node
			ast, err = d.ParseAST(tc.Input)
			if err == nil {
				res = &Result{AST: ast, Code: tc.Input}
			}
		} else if len(tc.Context) > 0 {
			res, err = d.Use(tc.Input, tc.Context)
		} else {
			res, err = d.Parse(tc.Input)
		}

		if shouldParse {
			if err != nil {
				report.Failed++
				report.Results = append(report.Results, DSLTestResult{
					Name:    name,
					Passed:  false,
					Message: fmt.Sprintf("expected parse success, got error: %v", err),
				})
				continue
			}
			if tc.ExpectAST != "" {
				astStr := ""
				if res != nil && res.GetAST() != nil {
					astStr = res.GetAST().Pretty()
				}
				if !strings.Contains(astStr, tc.ExpectAST) {
					report.Failed++
					report.Results = append(report.Results, DSLTestResult{
						Name:    name,
						Passed:  false,
						Message: fmt.Sprintf("AST did not contain %q; got:\n%s", tc.ExpectAST, astStr),
					})
					continue
				}
			}
			if tc.ExpectOutput != nil {
				actual := res.GetOutput()
				if !outputsMatch(tc.ExpectOutput, actual) {
					report.Failed++
					report.Results = append(report.Results, DSLTestResult{
						Name:    name,
						Passed:  false,
						Message: fmt.Sprintf("expected output %v (%T), got %v (%T)", tc.ExpectOutput, tc.ExpectOutput, actual, actual),
					})
					continue
				}
			}
			report.Passed++
			report.Results = append(report.Results, DSLTestResult{Name: name, Passed: true})
		} else {
			if err == nil {
				report.Failed++
				report.Results = append(report.Results, DSLTestResult{
					Name:    name,
					Passed:  false,
					Message: fmt.Sprintf("expected parse failure, but input %q succeeded", tc.Input),
				})
				continue
			}
			if tc.ExpectError != "" && !strings.Contains(err.Error(), tc.ExpectError) {
				report.Failed++
				report.Results = append(report.Results, DSLTestResult{
					Name:    name,
					Passed:  false,
					Message: fmt.Sprintf("expected error containing %q, got %q", tc.ExpectError, err.Error()),
				})
				continue
			}
			report.Passed++
			report.Results = append(report.Results, DSLTestResult{Name: name, Passed: true})
		}
	}

	return report
}

func outputsMatch(expected, actual interface{}) bool {
	if reflect.DeepEqual(expected, actual) {
		return true
	}
	// Normalize numeric comparisons across YAML/JSON int vs float64 decoding.
	ef, eok := toFloat64(expected)
	af, aok := toFloat64(actual)
	if eok && aok {
		return ef == af
	}
	return fmt.Sprintf("%v", expected) == fmt.Sprintf("%v", actual)
}

func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case float64:
		return n, true
	case float32:
		return float64(n), true
	default:
		return 0, false
	}
}

// GenerateSamples synthesizes valid DSL programs directly from the grammar
// rules, Pratt expressions, and token definitions.
func (d *DSL) GenerateSamples(seed int64, count int, maxDepth int) []string {
	if count <= 0 || d.grammar.startRule == "" {
		return nil
	}
	if maxDepth <= 0 {
		maxDepth = 4
	}

	rng := rand.New(rand.NewSource(seed))
	seen := make(map[string]bool)
	samples := make([]string, 0, count)

	maxAttempts := count * 8
	for attempt := 0; attempt < maxAttempts && len(samples) < count; attempt++ {
		s := strings.TrimSpace(d.genSymbol(d.grammar.startRule, 0, maxDepth, rng))
		if s == "" || seen[s] {
			continue
		}
		if _, err := d.ParseAST(s); err == nil {
			seen[s] = true
			samples = append(samples, s)
		}
	}
	return samples
}

func (d *DSL) genSymbol(sym string, depth, maxDepth int, rng *rand.Rand) string {
	g := d.grammar
	if g.isToken(sym) {
		return d.sampleForToken(sym, rng)
	}

	if spec, ok := g.exprRules[sym]; ok {
		return d.genExpr(spec, depth, maxDepth, rng)
	}

	rule, ok := g.rules[sym]
	if !ok || len(rule.alternatives) == 0 {
		return ""
	}

	// Prefer non-recursive / shorter alternatives once depth reaches maxDepth.
	candidates := rule.alternatives
	if depth >= maxDepth {
		var base []*Alternative
		for _, alt := range rule.alternatives {
			recursive := false
			for _, s := range alt.sequence {
				if !g.isToken(s) {
					recursive = true
					break
				}
			}
			if !recursive {
				base = append(base, alt)
			}
		}
		if len(base) == 0 {
			for _, alt := range rule.alternatives {
				if len(alt.sequence) == 0 || alt.sequence[0] != sym {
					base = append(base, alt)
				}
			}
		}
		if len(base) > 0 {
			candidates = base
		}
	}

	alt := candidates[rng.Intn(len(candidates))]
	parts := make([]string, 0, len(alt.sequence))
	for _, item := range alt.sequence {
		piece := d.genSymbol(item, depth+1, maxDepth, rng)
		if piece != "" {
			parts = append(parts, piece)
		}
	}
	if g.significantNewlines {
		var b strings.Builder
		for i, p := range parts {
			if i > 0 && p != "\n" && parts[i-1] != "\n" {
				b.WriteByte(' ')
			}
			b.WriteString(p)
		}
		return b.String()
	}
	return strings.Join(parts, " ")
}

func (d *DSL) genExpr(spec *exprSpec, depth, maxDepth int, rng *rand.Rand) string {
	var primary string
	if len(spec.atoms) > 0 && (depth >= maxDepth || len(spec.groups) == 0 || rng.Intn(3) > 0) {
		atom := spec.atoms[rng.Intn(len(spec.atoms))]
		primary = d.sampleForToken(atom.token, rng)
	} else if len(spec.groups) > 0 {
		gr := spec.groups[rng.Intn(len(spec.groups))]
		primary = d.sampleForToken(gr.open, rng) + " " + d.genSymbol(gr.inner, depth+1, maxDepth, rng) + " " + d.sampleForToken(gr.close, rng)
	}
	if primary == "" {
		return ""
	}

	if depth >= maxDepth {
		return primary
	}

	// Optionally wrap with prefix, postfix, or infix operator.
	choice := rng.Intn(4)
	if choice == 0 && len(spec.prefix) > 0 {
		keys := sortedMapKeys(spec.prefix)
		op := keys[rng.Intn(len(keys))]
		return d.sampleForToken(op, rng) + " " + primary
	}
	if choice == 1 && len(spec.infix) > 0 {
		keys := sortedMapKeys(spec.infix)
		op := keys[rng.Intn(len(keys))]
		rhs := d.genExpr(spec, depth+1, maxDepth, rng)
		if rhs != "" {
			return primary + " " + d.sampleForToken(op, rng) + " " + rhs
		}
	}
	if choice == 2 && len(spec.postfix) > 0 {
		keys := sortedMapKeys(spec.postfix)
		op := keys[rng.Intn(len(keys))]
		return primary + " " + d.sampleForToken(op, rng)
	}
	return primary
}

func (d *DSL) sampleForToken(name string, rng *rand.Rand) string {
	if d.grammar.significantNewlines && name == "NEWLINE" {
		if _, explicit := d.grammar.tokens[name]; !explicit {
			return "\n"
		}
	}
	tok, ok := d.grammar.tokens[name]
	if !ok {
		return ""
	}
	if (tok.kind == tokenKindKeyword || tok.kind == tokenKindLiteral) && tok.rawText != "" {
		return tok.rawText
	}
	if lit, isLit := literalText(tok.pattern); isLit {
		return lit
	}

	// Common regex families + validation against the compiled token regex!
	candidates := []string{
		fmt.Sprintf("%d", rng.Intn(90)+1),
		fmt.Sprintf("%d.%d", rng.Intn(20)+1, rng.Intn(9)+1),
		[]string{"alpha", "beta", "item", "x", "val"}[rng.Intn(5)],
		`"sample"`,
		`'sample'`,
		"$var",
		"http://example.com",
		"true",
		"a",
		"1",
		"+", "-", "*", "/", "=", "==", "!=", "<", ">", "(", ")", "[", "]", "{", "}", ",", ";", ":", ".", "?",
	}
	for _, c := range candidates {
		if loc := tok.regex.FindStringIndex(c); loc != nil && loc[0] == 0 && loc[1] == len(c) {
			return c
		}
	}
	return "x"
}

// FuzzGrammar stress-tests the parser, diagnostics, completion engine, and
// formatter against grammar-generated inputs and adversarial mutations,
// verifying that no input causes a panic.
func (d *DSL) FuzzGrammar(seed int64, iterations int) FuzzReport {
	if iterations <= 0 {
		iterations = 100
	}
	rng := rand.New(rand.NewSource(seed))
	seeds := d.GenerateSamples(seed, 16, 4)
	if len(seeds) == 0 {
		seeds = []string{"", "x", "1 + 2", "( )"}
	}

	report := FuzzReport{Iterations: iterations}

	for i := 0; i < iterations; i++ {
		base := seeds[rng.Intn(len(seeds))]
		input := mutateFuzzInput(base, i, rng)

		func() {
			defer func() {
				if r := recover(); r != nil {
					report.Failures = append(report.Failures, FuzzFailure{
						Iteration: i,
						Input:     input,
						Reason:    fmt.Sprintf("panic: %v", r),
					})
				}
			}()

			_, err := d.ParseAST(input)
			if err == nil {
				report.ValidParsed++
				_, _ = d.Format(input)
			} else {
				report.RejectedCleanly++
			}
			_ = d.Diagnostics(input)
			if len(input) > 0 {
				_ = d.Completions(input, rng.Intn(len(input)+1))
			} else {
				_ = d.Completions("", 0)
			}
		}()
	}

	return report
}

func mutateFuzzInput(base string, iter int, rng *rand.Rand) string {
	switch iter % 6 {
	case 0:
		return base
	case 1:
		if len(base) == 0 {
			return "@@@"
		}
		cut := rng.Intn(len(base))
		return base[:cut]
	case 2:
		return base + " " + []string{"@", "ñ", "🚀", "\xff", "(((", ")))", ";"}[rng.Intn(7)]
	case 3:
		return strings.Repeat("( ", rng.Intn(25)+1) + base
	case 4:
		return base + "\n" + base
	default:
		if len(base) < 2 {
			return "?!#"
		}
		b := []byte(base)
		idx := rng.Intn(len(b))
		b[idx] ^= 0x2A
		return string(b)
	}
}
