package dslbuilder_test

import (
	"fmt"
	"strconv"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

// The classic API: tokens, rules, actions.
func Example() {
	dsl := dslbuilder.New("greeting")
	dsl.KeywordToken("HELLO", "hello")
	dsl.KeywordToken("WORLD", "world")
	dsl.Rule("greeting", []string{"HELLO", "WORLD"}, "greet")
	dsl.Action("greet", func(args []interface{}) (interface{}, error) {
		return "Hello, World!", nil
	})

	result, _ := dsl.Parse("hello world")
	fmt.Println(result.GetOutput())
	// Output: Hello, World!
}

// Expression declares a Pratt-parsed expression rule: operator precedence
// and associativity without grammar gymnastics.
func ExampleDSL_Expression() {
	calc := dslbuilder.New("calc")
	calc.Token("NUMBER", `[0-9]+`)
	calc.Token("PLUS", `\+`)
	calc.Token("STAR", `\*`)
	calc.Token("LPAREN", `\(`)
	calc.Token("RPAREN", `\)`)

	calc.Expression("expr").
		Atom("NUMBER", "number").
		Group("LPAREN", "expr", "RPAREN").
		InfixLeft("PLUS", 10, "add").
		InfixLeft("STAR", 20, "mul")

	calc.Action("number", dslbuilder.Action1(strconv.Atoi))
	calc.Action("add", dslbuilder.Action3(func(l int, _ string, r int) (int, error) {
		return l + r, nil
	}))
	calc.Action("mul", dslbuilder.Action3(func(l int, _ string, r int) (int, error) {
		return l * r, nil
	}))

	r1, _ := calc.Parse("1 + 2 * 3")
	r2, _ := calc.Parse("(1 + 2) * 3")
	fmt.Println(r1.GetOutput(), r2.GetOutput())
	// Output: 7 9
}

// ParseAST and Eval are the two phases behind Parse: build the tree without
// running actions, then evaluate it exactly once.
func ExampleDSL_ParseAST() {
	calc := dslbuilder.New("calc")
	calc.Token("NUMBER", `[0-9]+`)
	calc.Token("PLUS", `\+`)
	calc.Expression("expr").
		Atom("NUMBER", "number").
		InfixLeft("PLUS", 10, "add")
	calc.Action("number", dslbuilder.Action1(strconv.Atoi))
	calc.Action("add", dslbuilder.Action3(func(l int, _ string, r int) (int, error) {
		return l + r, nil
	}))

	node, _ := calc.ParseAST("1 + 2") // phase 1: no actions run
	value, _ := calc.Eval(node)       // phase 2: actions run once

	fmt.Println(node.Rule, node.Action, value)
	// Output: expr add 3
}

// Build validates the grammar and freezes it; the resulting CompiledDSL is
// safe to share across goroutines.
func ExampleDSL_Build() {
	dsl := dslbuilder.New("nums")
	dsl.Token("NUM", "[0-9]+")
	dsl.Rule("expr", []string{"NUM"}, "number")
	dsl.Action("number", dslbuilder.Action1(strconv.Atoi))

	compiled, err := dsl.Build()
	if err != nil {
		fmt.Println("invalid grammar:", err)
		return
	}

	result, _ := compiled.Parse("42")
	fmt.Println(result.GetOutput())
	// Output: 42
}

// Validate reports structural errors and suspicious constructs without
// parsing anything.
func ExampleDSL_Validate() {
	dsl := dslbuilder.New("broken")
	dsl.Token("NUM", "[0-9]+")
	dsl.Rule("expr", []string{"MISSING"}, "x")

	_, err := dsl.Validate()
	fmt.Println(err)
	// Output: rule expr references unknown symbol MISSING
}

// Use passes a context scoped to a single call: it shadows the persistent
// context and is discarded afterwards.
func ExampleDSL_Use() {
	dsl := dslbuilder.New("ctx")
	dsl.KeywordToken("WHO", "who")
	dsl.Rule("q", []string{"WHO"}, "who")
	dsl.Action("who", func(args []interface{}) (interface{}, error) {
		return dsl.GetContext("user"), nil
	})
	dsl.SetContext("user", "default")

	r1, _ := dsl.Use("who", map[string]interface{}{"user": "alice"})
	r2, _ := dsl.Parse("who") // the Use() context did not persist
	fmt.Println(r1.GetOutput(), r2.GetOutput())
	// Output: alice default
}

// NodeAction registers a lazy action that receives the unevaluated node:
// only the branch that is taken executes.
func ExampleDSL_NodeAction() {
	dsl := dslbuilder.New("cond")
	dsl.KeywordToken("IF", "if")
	dsl.KeywordToken("ELSE", "else")
	dsl.Token("NUM", "[0-9]+")

	dsl.Rule("stmt", []string{"IF", "NUM", "value", "ELSE", "value"}, "ifElse")
	dsl.Rule("value", []string{"NUM"}, "echo")

	evaluated := 0
	dsl.Action("echo", func(args []interface{}) (interface{}, error) {
		evaluated++
		return args[0], nil
	})
	dsl.NodeAction("ifElse", func(ctx *dslbuilder.EvalContext, n *dslbuilder.Node) (interface{}, error) {
		cond, _ := ctx.Eval(n.Child(1))
		if cond == "1" {
			return ctx.Eval(n.Child(2)) // then-branch only
		}
		return ctx.Eval(n.Child(4)) // else-branch only
	})

	result, _ := dsl.Parse("if 1 10 else 20")
	fmt.Println(result.GetOutput(), "branches evaluated:", evaluated)
	// Output: 10 branches evaluated: 1
}

// Tokens and Expr form the high-level fluent layer.
func ExampleDSL_Tokens() {
	calc, _ := dslbuilder.New("calc").
		Tokens(func(t *dslbuilder.TokenSet) {
			t.Regex("NUMBER", `\d+`)
			t.Literal("PLUS", "+")
		}).
		Expr("expr", func(e *dslbuilder.ExpressionBuilder) {
			e.Atom("NUMBER", "number")
			e.InfixLeft("PLUS", 10, "add")
		}).
		WithAction("number", dslbuilder.Action1(strconv.Atoi)).
		WithAction("add", dslbuilder.Action3(func(l int, _ string, r int) (int, error) {
			return l + r, nil
		})).
		Build()

	result, _ := calc.Parse("20 + 22")
	fmt.Println(result.GetOutput())
	// Output: 42
}

// GetDetailedError renders farthest-failure syntax errors with position,
// expected tokens, and a visual pointer.
func ExampleGetDetailedError() {
	dsl := dslbuilder.New("mini")
	dsl.KeywordToken("IF", "if")
	dsl.KeywordToken("THEN", "then")
	dsl.Token("NUM", "[0-9]+")
	dsl.Rule("stmt", []string{"IF", "NUM", "THEN", "NUM"}, "")

	_, err := dsl.Parse("if then")
	fmt.Println(dslbuilder.GetDetailedError(err))
	// Output:
	// no alternative matched for rule stmt: expected NUM, got THEN "then" at line 1, column 4:
	// if then
	//    ^
}

// AttributeGrammar evaluates inherited (top-down) and synthesized
// (bottom-up) attributes over the AST. The initial Attrs passed to
// Evaluate act as the root's parent context and propagate down.
func ExampleAttributeGrammar() {
	dsl := dslbuilder.New("calc")
	dsl.Token("NUMBER", `[0-9]+`)
	dsl.Token("PLUS", `\+`)
	dsl.Expression("expr").
		Atom("NUMBER", "number").
		InfixLeft("PLUS", 10, "add")

	root, _ := dsl.ParseAST("1 + 2")

	ag := dslbuilder.NewAttributeGrammar()

	// Inherited: the environment flows from the initial context to every node.
	ag.Inherited("env", func(parent dslbuilder.Attrs, n *dslbuilder.Node, i int) interface{} {
		return parent["env"]
	})

	// Synthesized: compute each expr's value bottom-up.
	ag.Synthesized("value", "expr", func(n *dslbuilder.Node, children []dslbuilder.Attrs) (interface{}, error) {
		switch n.Action {
		case "number":
			return strconv.Atoi(n.Child(0).Token.Value)
		case "add":
			return children[0]["value"].(int) + children[2]["value"].(int), nil
		}
		return nil, nil
	})

	// The initial context is the "parent" of the root.
	attrs, _ := ag.Evaluate(root, dslbuilder.Attrs{"env": "prod"})

	fmt.Println(attrs[root]["value"], attrs[root]["env"], attrs[root.Child(0)]["env"])
	// Output: 3 prod prod
}
