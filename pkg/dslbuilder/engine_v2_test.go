package dslbuilder

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Expression / Pratt parser -------------------------------------------

func newCalcDSL(t *testing.T) *DSL {
	t.Helper()
	d := New("calc")
	require.NoError(t, d.Token("NUMBER", `[0-9]+`))
	require.NoError(t, d.Token("PLUS", `\+`))
	require.NoError(t, d.Token("MINUS", `-`))
	require.NoError(t, d.Token("STAR", `\*`))
	require.NoError(t, d.Token("SLASH", `/`))
	require.NoError(t, d.Token("POW", `\^`))
	require.NoError(t, d.Token("LPAREN", `\(`))
	require.NoError(t, d.Token("RPAREN", `\)`))

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
		return strconv.Atoi(args[0].(string))
	})
	d.Action("neg", func(args []interface{}) (interface{}, error) {
		return -args[1].(int), nil
	})
	binop := func(fn func(a, b int) int) ActionFunc {
		return func(args []interface{}) (interface{}, error) {
			return fn(args[0].(int), args[2].(int)), nil
		}
	}
	d.Action("add", binop(func(a, b int) int { return a + b }))
	d.Action("sub", binop(func(a, b int) int { return a - b }))
	d.Action("mul", binop(func(a, b int) int { return a * b }))
	d.Action("div", binop(func(a, b int) int { return a / b }))
	d.Action("pow", binop(func(a, b int) int {
		return int(math.Pow(float64(a), float64(b)))
	}))
	return d
}

func TestExpressionPrecedence(t *testing.T) {
	d := newCalcDSL(t)

	tests := []struct {
		input string
		want  int
	}{
		{"1 + 2 * 3", 7},
		{"(1 + 2) * 3", 9},
		{"10 - 3 - 2", 5},
		{"2 ^ 3 ^ 2", 512},
		{"-1 * 2", -2},
		{"2 * 3 + 4", 10},
		{"100 / 10 / 2", 5},
		{"-(1 + 2)", -3},
		{"--3", 3},
		{"2 ^ 2 ^ 3", 256},
		{"((((7))))", 7},
	}
	for _, tc := range tests {
		result, err := d.Parse(tc.input)
		require.NoError(t, err, "input: %s", tc.input)
		assert.Equal(t, tc.want, result.GetOutput(), "input: %s", tc.input)
	}
}

func TestExpressionErrors(t *testing.T) {
	d := newCalcDSL(t)

	_, err := d.Parse("1 +")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected")

	_, err = d.Parse("(1 + 2")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "RPAREN")

	_, err = d.Parse("* 2")
	require.Error(t, err)
}

func TestExpressionInsideRegularRule(t *testing.T) {
	d := New("stmt")
	require.NoError(t, d.KeywordToken("PRINT", "print"))
	require.NoError(t, d.Token("NUMBER", `[0-9]+`))
	require.NoError(t, d.Token("PLUS", `\+`))

	d.Rule("statement", []string{"PRINT", "expr"}, "print")
	d.Expression("expr").
		Atom("NUMBER", "number").
		InfixLeft("PLUS", 10, "add")

	d.Action("number", Action1(strconv.Atoi))
	d.Action("add", func(args []interface{}) (interface{}, error) {
		return args[0].(int) + args[2].(int), nil
	})
	d.Action("print", func(args []interface{}) (interface{}, error) {
		return fmt.Sprintf("printed %v", args[1]), nil
	})

	result, err := d.Parse("print 1 + 2 + 3")
	require.NoError(t, err)
	assert.Equal(t, "printed 6", result.GetOutput())
}

// --- Two-phase engine: parsing never runs actions -------------------------

func TestActionsDoNotRunOnRejectedAlternatives(t *testing.T) {
	d := New("sideeffects")
	require.NoError(t, d.KeywordToken("GET", "get"))
	require.NoError(t, d.KeywordToken("POST", "post"))
	require.NoError(t, d.Token("URL", `[a-z:/.]+`))

	sideEffects := 0
	// The GET alternative is declared first, so the parser tries it and
	// backtracks for a POST input. Its action must NOT run in that case.
	d.Rule("request", []string{"GET", "URL"}, "doGet")
	d.Rule("request", []string{"POST", "URL"}, "doPost")
	d.Action("doGet", func(args []interface{}) (interface{}, error) {
		sideEffects++
		return "GET " + args[1].(string), nil
	})
	d.Action("doPost", func(args []interface{}) (interface{}, error) {
		return "POST " + args[1].(string), nil
	})

	result, err := d.Parse("post http://example.com")
	require.NoError(t, err)
	assert.Equal(t, "POST http://example.com", result.GetOutput())
	assert.Equal(t, 0, sideEffects, "rejected alternative executed its action")
}

func TestParseASTAndEval(t *testing.T) {
	d := newCalcDSL(t)

	node, err := d.ParseAST("1 + 2 * 3")
	require.NoError(t, err)
	require.NotNil(t, node)
	assert.Equal(t, "expr", node.Rule)
	assert.Equal(t, "add", node.Action)
	require.Len(t, node.Children, 3)
	assert.Equal(t, "mul", node.Children[2].Action)

	// Spans cover the input.
	assert.Equal(t, 0, node.Span.Start)
	assert.Equal(t, len("1 + 2 * 3"), node.Span.End)

	// Pretty printing works and mentions the structure.
	pretty := node.Pretty()
	assert.Contains(t, pretty, "expr (add)")
	assert.Contains(t, pretty, `NUMBER "3"`)

	// Evaluation is separate and repeatable.
	for i := 0; i < 2; i++ {
		value, err := d.Eval(node)
		require.NoError(t, err)
		assert.Equal(t, 7, value)
	}
}

func TestResultGetAST(t *testing.T) {
	d := newCalcDSL(t)
	result, err := d.Parse("2 + 2")
	require.NoError(t, err)
	require.NotNil(t, result.GetAST())
	assert.Equal(t, "add", result.GetAST().Action)
	assert.Equal(t, "2 + 2", result.GetAST().Text())
}

func TestEvalActionErrorSurfaces(t *testing.T) {
	d := New("evalerr")
	require.NoError(t, d.Token("A", "a"))
	require.NoError(t, d.Token("B", "b"))
	d.Rule("expr", []string{"expr", "A"}, "boom")
	d.Rule("expr", []string{"B"}, "base")
	d.Action("boom", func(args []interface{}) (interface{}, error) {
		return nil, fmt.Errorf("action error")
	})
	d.Action("base", func(args []interface{}) (interface{}, error) {
		return "b", nil
	})

	_, err := d.Parse("b a")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "action error")
}

// --- Deterministic tokenizer ----------------------------------------------

func TestTokenizerDeterministicTieBreak(t *testing.T) {
	// Two tokens with identical priority and pattern length matching the
	// same text: the first declared must always win.
	for i := 0; i < 20; i++ {
		d := New("ties")
		require.NoError(t, d.Token("FIRST", `[a-z]+`))
		require.NoError(t, d.Token("SECOND", `[a-z]+`))
		d.Rule("s", []string{"FIRST"}, "f")
		d.Action("f", func(args []interface{}) (interface{}, error) {
			return "first wins", nil
		})

		result, err := d.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "first wins", result.GetOutput())

		tokens, err := d.DebugTokens("hello")
		require.NoError(t, err)
		require.Len(t, tokens, 1)
		assert.Equal(t, "FIRST", tokens[0].TokenType)
	}
}

func TestTokenizerRejectsEmptyMatch(t *testing.T) {
	d := New("empty")
	err := d.Token("BAD", `[0-9]*`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty string")

	err = d.Token("ALSO_BAD", `a?`)
	require.Error(t, err)

	// Non-empty patterns are fine.
	require.NoError(t, d.Token("GOOD", `[0-9]+`))
}

func TestTokenRedefinitionKeepsOrder(t *testing.T) {
	d := New("redef")
	require.NoError(t, d.Token("A", "aaa"))
	require.NoError(t, d.Token("B", "[a-z]+"))
	// Redefine A: it must keep its first-declared position.
	require.NoError(t, d.Token("A", "[a-z]+"))
	require.Len(t, d.grammar.tokenList, 2)
	assert.Equal(t, "A", d.grammar.tokenList[0].name)

	d.Rule("s", []string{"A"}, "")
	tokens, err := d.DebugTokens("abc")
	require.NoError(t, err)
	require.Len(t, tokens, 1)
	assert.Equal(t, "A", tokens[0].TokenType)
}

// --- Use() context immutability -------------------------------------------

func TestUseContextDoesNotPersist(t *testing.T) {
	d := New("ctx")
	require.NoError(t, d.KeywordToken("SHOW", "show"))
	d.Rule("cmd", []string{"SHOW"}, "show")
	d.Action("show", func(args []interface{}) (interface{}, error) {
		if v := d.GetContext("user"); v != nil {
			return v, nil
		}
		return "anonymous", nil
	})

	d.SetContext("user", "default-user")

	// Use() context shadows the persistent one during the call...
	result, err := d.Use("show", map[string]interface{}{"user": "alice"})
	require.NoError(t, err)
	assert.Equal(t, "alice", result.GetOutput())

	// ...but does not leak into later parses.
	result, err = d.Parse("show")
	require.NoError(t, err)
	assert.Equal(t, "default-user", result.GetOutput())
	assert.Equal(t, "default-user", d.GetContext("user"))
}

func TestSetContextDuringUseStillPersists(t *testing.T) {
	d := New("ctx2")
	require.NoError(t, d.KeywordToken("SET", "set"))
	d.Rule("cmd", []string{"SET"}, "set")
	d.Action("set", func(args []interface{}) (interface{}, error) {
		d.SetContext("stored", "value-from-action")
		return "ok", nil
	})

	_, err := d.Use("set", map[string]interface{}{"tmp": 1})
	require.NoError(t, err)

	// Explicit SetContext from an action persists; the Use() ctx does not.
	assert.Equal(t, "value-from-action", d.GetContext("stored"))
	assert.Nil(t, d.GetContext("tmp"))
}

// --- Build() / frozen grammar ---------------------------------------------

func TestBuildFreezesGrammar(t *testing.T) {
	d := New("frozen")
	require.NoError(t, d.Token("NUM", "[0-9]+"))
	d.Rule("expr", []string{"NUM"}, "num")
	d.Action("num", Action1(strconv.Atoi))

	compiled, err := d.Build()
	require.NoError(t, err)

	result, err := compiled.Parse("42")
	require.NoError(t, err)
	assert.Equal(t, 42, result.GetOutput())

	// Mutations after Build are rejected.
	err = d.Token("LATE", "x")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "frozen")

	d.Rule("late", []string{"NUM"}, "num") // deferred error
	_, err = d.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "frozen")
}

func TestBuildFailsOnInvalidGrammar(t *testing.T) {
	d := New("invalid")
	require.NoError(t, d.Token("NUM", "[0-9]+"))
	d.Rule("expr", []string{"MISSING_SYMBOL"}, "x")

	_, err := d.Build()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown symbol")
}

func TestCompiledDSLConcurrentParse(t *testing.T) {
	d := newCalcDSL(t)
	compiled, err := d.Build()
	require.NoError(t, err)

	done := make(chan error, 8)
	for i := 0; i < 8; i++ {
		go func() {
			for j := 0; j < 50; j++ {
				result, err := compiled.Parse("1 + 2 * 3")
				if err != nil {
					done <- err
					return
				}
				if result.GetOutput() != 7 {
					done <- fmt.Errorf("got %v, want 7", result.GetOutput())
					return
				}
			}
			done <- nil
		}()
	}
	for i := 0; i < 8; i++ {
		require.NoError(t, <-done)
	}
}

// --- Validate() ------------------------------------------------------------

func TestValidateReportsProblems(t *testing.T) {
	d := New("validate")
	require.NoError(t, d.Token("NUM", "[0-9]+"))

	d.Rule("start", []string{"NUM"}, "registered")
	d.Rule("start", []string{"ghost"}, "registered")    // unknown symbol -> error
	d.Rule("orphan", []string{"NUM"}, "missing_action") // unreachable + unregistered action
	d.Action("registered", func(a []interface{}) (interface{}, error) { return a[0], nil })

	warnings, err := d.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown symbol ghost")

	joined := strings.Join(warnings, "\n")
	assert.Contains(t, joined, "unreachable")
	assert.Contains(t, joined, "missing_action")
}

func TestValidateDetectsIndirectLeftRecursion(t *testing.T) {
	d := New("indirect")
	require.NoError(t, d.Token("X", "x"))
	// a -> b ... ; b -> a ... : indirect left recursion (unsupported)
	d.Rule("a", []string{"b", "X"}, "")
	d.Rule("b", []string{"a", "X"}, "")
	d.Rule("b", []string{"X"}, "")

	warnings, err := d.Validate()
	require.NoError(t, err)
	joined := strings.Join(warnings, "\n")
	assert.Contains(t, joined, "indirect left recursion")
}

func TestValidateDirectLeftRecursionIsFine(t *testing.T) {
	d := New("direct")
	require.NoError(t, d.Token("X", "x"))
	d.Rule("list", []string{"list", "X"}, "")
	d.Rule("list", []string{"X"}, "")

	warnings, err := d.Validate()
	require.NoError(t, err)
	for _, w := range warnings {
		assert.NotContains(t, w, "indirect left recursion")
	}
}

func TestValidateNonProductiveRule(t *testing.T) {
	d := New("nonprod")
	require.NoError(t, d.Token("X", "x"))
	d.Rule("start", []string{"X"}, "")
	d.Rule("loop", []string{"loop2"}, "")
	d.Rule("loop2", []string{"loop"}, "")

	warnings, err := d.Validate()
	require.NoError(t, err)
	joined := strings.Join(warnings, "\n")
	assert.Contains(t, joined, "non-productive")
}

// --- Farthest failure errors ------------------------------------------------

func TestFarthestFailureError(t *testing.T) {
	d := New("errors")
	require.NoError(t, d.KeywordToken("IF", "if"))
	require.NoError(t, d.KeywordToken("THEN", "then"))
	require.NoError(t, d.Token("EQ", "=="))
	require.NoError(t, d.Token("NUMBER", `[0-9]+`))
	require.NoError(t, d.Token("IDENT", `[a-z]+`))

	d.Rule("if_stmt", []string{"IF", "condition", "THEN", "IDENT"}, "if")
	d.Rule("condition", []string{"IDENT", "EQ", "value"}, "cond")
	d.Rule("value", []string{"NUMBER"}, "")
	d.Rule("value", []string{"IDENT"}, "")

	_, err := d.Parse("if status == then")
	require.Error(t, err)
	require.True(t, IsParseError(err))

	msg := err.Error()
	// Points at the farthest failure ("then" where a value was expected),
	// lists expectations, and keeps the historical prefix.
	assert.Contains(t, msg, "no alternative matched")
	assert.Contains(t, msg, "expected")
	assert.Contains(t, msg, "NUMBER")
	assert.Contains(t, msg, "rule stack:")
	assert.Contains(t, msg, "if_stmt")

	detailed := GetDetailedError(err)
	assert.Contains(t, detailed, "line 1")
	assert.Contains(t, detailed, "^")
}

// --- Typed actions -----------------------------------------------------------

func TestTypedActions(t *testing.T) {
	d := New("typed")
	require.NoError(t, d.Token("NUMBER", `[0-9]+`))
	require.NoError(t, d.Token("PLUS", `\+`))
	d.Rule("expr", []string{"expr", "PLUS", "term"}, "add")
	d.Rule("expr", []string{"term"}, "pass")
	d.Rule("term", []string{"NUMBER"}, "number")

	d.Action("number", Action1(strconv.Atoi))
	d.Action("pass", func(args []interface{}) (interface{}, error) { return args[0], nil })
	d.Action("add", Action3(func(left int, _ string, right int) (int, error) {
		return left + right, nil
	}))

	result, err := d.Parse("1 + 2 + 39")
	require.NoError(t, err)
	assert.Equal(t, 42, result.GetOutput())
}

func TestTypedActionConversionError(t *testing.T) {
	fn := Action1(func(n int) (int, error) { return n, nil })
	_, err := fn([]interface{}{"not-a-number"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot convert")

	_, err = fn([]interface{}{"1", "2"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected 1 arguments")
}

func TestArgsHelpers(t *testing.T) {
	args := Args{"10", 20, 2.5, "x"}
	assert.Equal(t, 4, args.Len())
	assert.Equal(t, 10, args.Int(0))
	assert.Equal(t, 20, args.Int(1))
	assert.Equal(t, 2.5, args.Float(2))
	assert.Equal(t, "x", args.String(3))
	assert.Equal(t, "20", args.String(1))
	assert.Equal(t, 0, args.Int(3)) // non-numeric -> zero value
	assert.Nil(t, args.Get(99))     // out of range
	assert.Equal(t, "", args.String(99))
}

// --- Nice API ----------------------------------------------------------------

func TestNiceAPICalculator(t *testing.T) {
	calc, err := New("calc").
		Tokens(func(t *TokenSet) {
			t.Regex("NUMBER", `\d+`)
			t.Literal("PLUS", "+")
			t.Literal("STAR", "*")
			t.Literal("LPAREN", "(")
			t.Literal("RPAREN", ")")
		}).
		Expr("expr", func(e *ExpressionBuilder) {
			e.Atom("NUMBER", "number")
			e.Group("LPAREN", "expr", "RPAREN")
			e.InfixLeft("PLUS", 10, "add")
			e.InfixLeft("STAR", 20, "mul")
		}).
		WithAction("number", Action1(strconv.Atoi)).
		WithAction("add", Action3(func(l int, _ string, r int) (int, error) { return l + r, nil })).
		WithAction("mul", Action3(func(l int, _ string, r int) (int, error) { return l * r, nil })).
		Build()
	require.NoError(t, err)

	result, err := calc.Parse("2 + 3 * (4 + 1)")
	require.NoError(t, err)
	assert.Equal(t, 17, result.GetOutput())
}

func TestNiceAPIDeferredTokenErrors(t *testing.T) {
	d := New("bad").
		Tokens(func(t *TokenSet) {
			t.Regex("BROKEN", "[") // invalid regex -> deferred
		})
	d.Rule("s", []string{"BROKEN"}, "")

	_, err := d.Build()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid regex")
}

func TestValidateDetectsShadowedAlternatives(t *testing.T) {
	d := New("shadow")
	require.NoError(t, d.Token("A", "a"))
	require.NoError(t, d.Token("B", "b"))
	// Shorter alternative first: the longer one can never match.
	d.Rule("cmd", []string{"A"}, "")
	d.Rule("cmd", []string{"A", "B"}, "")

	warnings, err := d.Validate()
	require.NoError(t, err)
	joined := strings.Join(warnings, "\n")
	assert.Contains(t, joined, "prefix of later alternative")

	// Longer first: no shadowing warning.
	d2 := New("ok")
	require.NoError(t, d2.Token("A", "a"))
	require.NoError(t, d2.Token("B", "b"))
	d2.Rule("cmd", []string{"A", "B"}, "")
	d2.Rule("cmd", []string{"A"}, "")
	warnings, err = d2.Validate()
	require.NoError(t, err)
	for _, w := range warnings {
		assert.NotContains(t, w, "prefix of later alternative")
	}

	// Left-recursive longer alternative is exempt (growing seed handles it).
	d3 := New("leftrec")
	require.NoError(t, d3.Token("X", "x"))
	d3.Rule("list", []string{"X"}, "")
	d3.Rule("list", []string{"list", "X"}, "")
	warnings, err = d3.Validate()
	require.NoError(t, err)
	for _, w := range warnings {
		assert.NotContains(t, w, "prefix of later alternative")
	}
}

func TestValidateDetectsDuplicateAlternatives(t *testing.T) {
	d := New("dup")
	require.NoError(t, d.Token("A", "a"))
	d.Rule("cmd", []string{"A"}, "x")
	d.Rule("cmd", []string{"A"}, "y")

	warnings, err := d.Validate()
	require.NoError(t, err)
	assert.Contains(t, strings.Join(warnings, "\n"), "identical")
}

func TestRuleWithErrorHint(t *testing.T) {
	d := New("hints")
	require.NoError(t, d.KeywordToken("IF", "if"))
	require.NoError(t, d.KeywordToken("THEN", "then"))
	require.NoError(t, d.Token("EQ", "=="))
	require.NoError(t, d.Token("NUMBER", `[0-9]+`))
	require.NoError(t, d.Token("IDENT", `[a-z]+`))

	d.Rule("if_stmt", []string{"IF", "condition", "THEN", "IDENT"}, "if")
	d.RuleWithError("condition", []string{"IDENT", "EQ", "NUMBER"}, "cmp",
		"expected a comparison like 'status == 200'")

	_, err := d.Parse("if status == then")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "hint: expected a comparison like 'status == 200'")

	// A failure outside the hinted alternative carries no hint.
	_, err = d.Parse("if status == 200 nothen")
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "hint:")
}

func TestDiagnosticsMultipleErrors(t *testing.T) {
	d := New("diag")
	require.NoError(t, d.KeywordToken("SET", "set"))
	require.NoError(t, d.Token("IDENT", `[a-z]+`))
	require.NoError(t, d.Token("NUMBER", `[0-9]+`))
	d.Rule("stmt", []string{"SET", "IDENT", "NUMBER"}, "")

	// Three statements: ok, broken (missing number), broken (bad char), ok.
	code := "set x 1 set y set z 3 set @ set w 4"
	diags := d.Diagnostics(code)

	require.NotEmpty(t, diags)
	// One lexical error for '@'.
	lexical := 0
	for _, diag := range diags {
		if strings.Contains(diag.Message, "unexpected character") {
			lexical++
		}
	}
	assert.Equal(t, 1, lexical, "expected one lexical diagnostic for '@'")
	// And at least one parse error (set y <missing number>).
	assert.GreaterOrEqual(t, len(diags), 2)

	// Clean input yields no diagnostics.
	assert.Empty(t, d.Diagnostics("set x 1"))
	// Multi-statement clean input also yields none.
	assert.Empty(t, d.Diagnostics("set x 1 set y 2 set z 3"))
}

func TestDiagnosticsPositions(t *testing.T) {
	d := New("diagpos")
	require.NoError(t, d.KeywordToken("SET", "set"))
	require.NoError(t, d.Token("IDENT", `[a-z]+`))
	require.NoError(t, d.Token("NUMBER", `[0-9]+`))
	d.Rule("stmt", []string{"SET", "IDENT", "NUMBER"}, "")

	code := "set x 1\nset y\nset z 3"
	diags := d.Diagnostics(code)
	require.Len(t, diags, 1)
	// The error is on line 2 or 3 (the failure is detected where the next
	// token contradicts the expectation).
	assert.GreaterOrEqual(t, diags[0].Line, 2)
	assert.Contains(t, diags[0].Message, "expected")
}

func TestDiagnosticsBounded(t *testing.T) {
	d := New("diagbound")
	require.NoError(t, d.Token("A", "a"))
	d.Rule("stmt", []string{"A"}, "")

	// 200 lexical errors: output must be capped.
	diags := d.Diagnostics(strings.Repeat("@ ", 200))
	assert.LessOrEqual(t, len(diags), 50)
}

func newIndirectLRDSL(t *testing.T) *DSL {
	t.Helper()
	// Classic indirect left recursion: a -> b X | NUM ; b -> a Y
	d := New("indirectlr")
	require.NoError(t, d.Token("NUM", `[0-9]+`))
	require.NoError(t, d.Token("X", "x"))
	require.NoError(t, d.Token("Y", "y"))
	d.Rule("a", []string{"b", "X"}, "growA")
	d.Rule("a", []string{"NUM"}, "num")
	d.Rule("b", []string{"a", "Y"}, "growB")

	d.Action("num", func(args []interface{}) (interface{}, error) {
		return args[0].(string), nil
	})
	d.Action("growA", func(args []interface{}) (interface{}, error) {
		return fmt.Sprintf("A(%v,x)", args[0]), nil
	})
	d.Action("growB", func(args []interface{}) (interface{}, error) {
		return fmt.Sprintf("B(%v,y)", args[0]), nil
	})
	return d
}

func TestIndirectLeftRecursionParses(t *testing.T) {
	d := newIndirectLRDSL(t)

	tests := []struct {
		input string
		want  string
	}{
		{"1", "1"},
		{"1 y x", "A(B(1,y),x)"},
		{"1 y x y x", "A(B(A(B(1,y),x),y),x)"},
		{"7 y x y x y x", "A(B(A(B(A(B(7,y),x),y),x),y),x)"},
	}
	for _, tc := range tests {
		result, err := d.Parse(tc.input)
		require.NoError(t, err, "input: %s", tc.input)
		assert.Equal(t, tc.want, result.GetOutput(), "input: %s", tc.input)
	}

	// Invalid inputs still fail cleanly.
	for _, bad := range []string{"y", "1 y", "1 x", "x y", ""} {
		_, err := d.Parse(bad)
		assert.Error(t, err, "input: %q", bad)
	}
}

func TestIndirectLeftRecursionThreeRuleCycle(t *testing.T) {
	// a -> b P | NUM ; b -> c Q ; c -> a R
	d := New("cycle3")
	require.NoError(t, d.Token("NUM", `[0-9]+`))
	require.NoError(t, d.Token("P", "p"))
	require.NoError(t, d.Token("Q", "q"))
	require.NoError(t, d.Token("R", "r"))
	d.Rule("a", []string{"b", "P"}, "")
	d.Rule("a", []string{"NUM"}, "")
	d.Rule("b", []string{"c", "Q"}, "")
	d.Rule("c", []string{"a", "R"}, "")

	for _, input := range []string{"1", "1 r q p", "1 r q p r q p"} {
		_, err := d.Parse(input)
		require.NoError(t, err, "input: %s", input)
	}
	for _, bad := range []string{"1 r", "1 r q", "r q p"} {
		_, err := d.Parse(bad)
		assert.Error(t, err, "input: %q", bad)
	}
}

func TestDirectLeftRecursionUnaffectedByCycleSupport(t *testing.T) {
	// A grammar with BOTH a direct-recursive rule and an indirect cycle:
	// the direct rule must keep its behavior.
	d := New("mixed")
	require.NoError(t, d.Token("NUM", `[0-9]+`))
	require.NoError(t, d.Token("PLUS", `\+`))
	require.NoError(t, d.Token("X", "x"))
	require.NoError(t, d.Token("Y", "y"))

	d.Rule("start", []string{"sum"}, "pass0")
	d.Rule("start", []string{"a"}, "pass0")
	d.Rule("sum", []string{"sum", "PLUS", "NUM"}, "add")
	d.Rule("sum", []string{"NUM"}, "num")
	d.Rule("a", []string{"b", "X"}, "")
	d.Rule("b", []string{"a", "Y"}, "")
	d.Rule("b", []string{"Y"}, "")

	d.Action("pass0", func(args []interface{}) (interface{}, error) { return args[0], nil })
	d.Action("num", Action1(strconv.Atoi))
	d.Action("add", func(args []interface{}) (interface{}, error) {
		return args[0].(int) + Args(args).Int(2), nil
	})

	result, err := d.Parse("1 + 2 + 39")
	require.NoError(t, err)
	assert.Equal(t, 42, result.GetOutput())

	_, err = d.Parse("y x y x")
	require.NoError(t, err)
}

func newStreamDSL(t *testing.T) *DSL {
	t.Helper()
	d := New("stream")
	require.NoError(t, d.KeywordToken("ADD", "add"))
	require.NoError(t, d.Token("NUMBER", `[0-9]+`))
	d.Rule("stmt", []string{"ADD", "NUMBER"}, "add")
	total := 0
	d.Action("add", func(args []interface{}) (interface{}, error) {
		n := Args(args).Int(1)
		total += n
		return total, nil
	})
	return d
}

func TestParseStream(t *testing.T) {
	d := newStreamDSL(t)
	input := "add 1\n\n# comment\nadd 2\n// other comment\nadd 39\n"

	var lines []int
	var outputs []interface{}
	err := d.ParseStream(strings.NewReader(input), func(line int, result *Result) error {
		lines = append(lines, line)
		outputs = append(outputs, result.GetOutput())
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, []int{1, 4, 6}, lines)
	assert.Equal(t, []interface{}{1, 3, 42}, outputs)
}

func TestParseStreamErrorCarriesLine(t *testing.T) {
	d := newStreamDSL(t)
	err := d.ParseStream(strings.NewReader("add 1\nadd oops\n"), nil)
	require.Error(t, err)
	require.True(t, IsParseError(err))
	assert.Equal(t, 2, err.(*ParseError).Line)
}

func TestParseStreamHandlerAborts(t *testing.T) {
	d := newStreamDSL(t)
	calls := 0
	err := d.ParseStream(strings.NewReader("add 1\nadd 2\nadd 3\n"), func(line int, result *Result) error {
		calls++
		return fmt.Errorf("stop here")
	})
	require.Error(t, err)
	assert.Equal(t, 1, calls)
	assert.Contains(t, err.Error(), "stop here")
}
