package dslbuilder

import (
	"strconv"
	"strings"
	"testing"
)

// fuzzCalc builds a calculator DSL with total actions (no panics on any
// input, including division by zero).
func fuzzCalc() *DSL {
	d := New("fuzzcalc")
	d.Token("NUMBER", `[0-9]+`)
	d.Token("PLUS", `\+`)
	d.Token("MINUS", `-`)
	d.Token("STAR", `\*`)
	d.Token("SLASH", `/`)
	d.Token("POW", `\^`)
	d.Token("LPAREN", `\(`)
	d.Token("RPAREN", `\)`)

	d.Expression("expr").
		Atom("NUMBER", "number").
		Group("LPAREN", "expr", "RPAREN").
		Prefix("MINUS", 70, "neg").
		InfixLeft("PLUS", 10, "add").
		InfixLeft("MINUS", 10, "sub").
		InfixLeft("STAR", 20, "mul").
		InfixLeft("SLASH", 20, "div").
		InfixRight("POW", 30, "pow")

	d.Action("number", func(args []interface{}) (interface{}, error) {
		n, err := strconv.Atoi(args[0].(string))
		if err != nil {
			return 0, nil // overflow-sized literals evaluate to 0
		}
		return n, nil
	})
	d.Action("neg", func(args []interface{}) (interface{}, error) {
		n, _ := args[1].(int)
		return -n, nil
	})
	binop := func(fn func(a, b int) int) ActionFunc {
		return func(args []interface{}) (interface{}, error) {
			a, _ := args[0].(int)
			b, _ := args[2].(int)
			return fn(a, b), nil
		}
	}
	d.Action("add", binop(func(a, b int) int { return a + b }))
	d.Action("sub", binop(func(a, b int) int { return a - b }))
	d.Action("mul", binop(func(a, b int) int { return a * b }))
	d.Action("div", binop(func(a, b int) int {
		if b == 0 {
			return 0
		}
		return a / b
	}))
	d.Action("pow", binop(func(a, b int) int {
		result := 1
		for i := 0; i < b && i < 63; i++ {
			result *= a
		}
		return result
	}))
	return d
}

// FuzzExpressionParse asserts the Pratt engine never panics: any input
// either parses and evaluates, or returns an error.
func FuzzExpressionParse(f *testing.F) {
	seeds := []string{
		"1 + 2 * 3",
		"(1 + 2) * 3",
		"10 - 3 - 2",
		"2 ^ 3 ^ 2",
		"-1 * 2",
		"--3",
		"((((7))))",
		"1 +",
		"* 2",
		"(1 + 2",
		")(",
		"",
		"   ",
		"1)))))",
		strings.Repeat("(", 500) + "1" + strings.Repeat(")", 500),
		strings.Repeat("(", 20000), // must error cleanly, not overflow
		strings.Repeat("-", 20000) + "1",
		"1" + strings.Repeat("+1", 5000),
	}
	for _, s := range seeds {
		f.Add(s)
	}

	d := fuzzCalc()
	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > 1<<16 {
			return // keep iterations fast
		}
		result, err := d.Parse(input)
		if err == nil && result == nil {
			t.Fatalf("nil result without error for %q", input)
		}
	})
}

// FuzzLeftRecursiveParse asserts the growing-seed algorithm never panics
// or loops forever on arbitrary input.
func FuzzLeftRecursiveParse(f *testing.F) {
	seeds := []string{
		"b",
		"b a",
		"b a a a a",
		"a",
		"a b",
		"",
		"b " + strings.Repeat("a ", 2000),
	}
	for _, s := range seeds {
		f.Add(s)
	}

	d := New("fuzzrec")
	d.Token("A", "a")
	d.Token("B", "b")
	d.Rule("list", []string{"list", "A"}, "append")
	d.Rule("list", []string{"B"}, "base")
	d.Action("append", func(args []interface{}) (interface{}, error) {
		return args[0].(string) + "+a", nil
	})
	d.Action("base", func(args []interface{}) (interface{}, error) {
		return "b", nil
	})

	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > 1<<16 {
			return
		}
		d.Parse(input) // must not panic or hang
	})
}

// FuzzGrammarRules asserts arbitrary token/rule text can be parsed
// against a small keyword grammar without panics.
func FuzzGrammarRules(f *testing.F) {
	seeds := []string{
		"if x then y",
		"if then",
		"if if if",
		"x == 10",
		"\x00\x01\x02",
		"ñandú 🦤",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	d := New("fuzzgrammar")
	d.KeywordToken("IF", "if")
	d.KeywordToken("THEN", "then")
	d.Token("EQ", "==")
	d.Token("NUMBER", `[0-9]+`)
	d.Token("IDENT", `[a-z]+`)
	d.Rule("stmt", []string{"IF", "cond", "THEN", "IDENT"}, "")
	d.Rule("cond", []string{"IDENT", "EQ", "value"}, "")
	d.Rule("value", []string{"NUMBER"}, "")
	d.Rule("value", []string{"IDENT"}, "")

	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > 1<<16 {
			return
		}
		d.Parse(input) // must not panic
	})
}

// FuzzIndirectLeftRecursion asserts the generalized growing algorithm
// (multi-rule leftmost cycles) never panics or hangs.
func FuzzIndirectLeftRecursion(f *testing.F) {
	seeds := []string{
		"1",
		"1 y x",
		"1 y x y x",
		"1 y",
		"y x",
		"",
		"1 " + strings.Repeat("y x ", 500),
		strings.Repeat("1 y x ", 100),
	}
	for _, s := range seeds {
		f.Add(s)
	}

	// a -> b X | NUM ; b -> a Y  (indirect left recursion)
	d := New("fuzzindirect")
	d.Token("NUM", `[0-9]+`)
	d.Token("X", "x")
	d.Token("Y", "y")
	d.Rule("a", []string{"b", "X"}, "")
	d.Rule("a", []string{"NUM"}, "")
	d.Rule("b", []string{"a", "Y"}, "")

	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > 1<<14 {
			return
		}
		d.Parse(input) // must not panic or hang
	})
}
