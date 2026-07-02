// Package main demonstrates the modern go-dsl engine:
//
//   - the high-level fluent API (Tokens / Expr)
//   - the Pratt expression parser (precedence + associativity)
//   - typed actions (Action1 / Action3)
//   - Build() validation + frozen grammar
//   - the two-phase engine (ParseAST + Eval) and AST pretty-printing
//   - farthest-failure error messages
package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

func main() {
	calc, err := dslbuilder.New("calc").
		Tokens(func(t *dslbuilder.TokenSet) {
			t.Regex("NUMBER", `\d+`)
			t.Literal("PLUS", "+")
			t.Literal("MINUS", "-")
			t.Literal("STAR", "*")
			t.Literal("SLASH", "/")
			t.Literal("POW", "^")
			t.Literal("LPAREN", "(")
			t.Literal("RPAREN", ")")
		}).
		Expr("expr", func(e *dslbuilder.ExpressionBuilder) {
			e.Atom("NUMBER", "number")
			e.Group("LPAREN", "expr", "RPAREN")
			e.Prefix("MINUS", 70, "neg")
			e.InfixLeft("PLUS", 10, "add")
			e.InfixLeft("MINUS", 10, "sub")
			e.InfixLeft("STAR", 20, "mul")
			e.InfixLeft("SLASH", 20, "div")
			e.InfixRight("POW", 30, "pow")
		}).
		WithAction("number", dslbuilder.Action1(strconv.Atoi)).
		WithAction("neg", func(args []interface{}) (interface{}, error) {
			return -args[1].(int), nil
		}).
		WithAction("add", dslbuilder.Action3(func(l int, _ string, r int) (int, error) { return l + r, nil })).
		WithAction("sub", dslbuilder.Action3(func(l int, _ string, r int) (int, error) { return l - r, nil })).
		WithAction("mul", dslbuilder.Action3(func(l int, _ string, r int) (int, error) { return l * r, nil })).
		WithAction("div", dslbuilder.Action3(func(l int, _ string, r int) (int, error) {
			if r == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return l / r, nil
		})).
		WithAction("pow", dslbuilder.Action3(func(l int, _ string, r int) (int, error) {
			result := 1
			for i := 0; i < r; i++ {
				result *= l
			}
			return result, nil
		})).
		Build() // validates the grammar and freezes it
	if err != nil {
		log.Fatalf("invalid grammar: %v", err)
	}

	fmt.Println("=== Pratt calculator (precedence + associativity) ===")
	for _, input := range []string{
		"1 + 2 * 3",     // 7   (* binds tighter than +)
		"(1 + 2) * 3",   // 9   (grouping)
		"10 - 3 - 2",    // 5   (left associative)
		"2 ^ 3 ^ 2",     // 512 (right associative)
		"-1 * 2",        // -2  (prefix operator)
		"-(3 + 4) ^ 2",  // 49 (prefix minus binds tighter than ^ at power 70)
		"100 / 10 / 2",  // 5
	} {
		result, err := calc.Parse(input)
		if err != nil {
			log.Fatalf("%s -> %v", input, err)
		}
		fmt.Printf("%-15s = %v\n", input, result.GetOutput())
	}

	fmt.Println("\n=== Two-phase engine: inspect the AST before evaluating ===")
	node, err := calc.ParseAST("1 + 2 * 3")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(node.Pretty())
	value, _ := calc.Eval(node)
	fmt.Printf("evaluates to: %v\n", value)

	fmt.Println("\n=== Farthest-failure errors ===")
	_, err = calc.Parse("1 + * 2")
	fmt.Println(dslbuilder.GetDetailedError(err))
}
