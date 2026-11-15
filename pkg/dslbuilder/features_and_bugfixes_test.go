package dslbuilder

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Bug Fix Regressions ([BUG-01] .. [BUG-11]) -----------------------------

func TestBug01DocumentIncrementalBoundaryAndLookahead(t *testing.T) {
	d := New("doc_boundary")
	require.NoError(t, d.KeywordToken("SET", "set"))
	require.NoError(t, d.Token("IDENT", `[a-z]+`))
	require.NoError(t, d.Token("NUMBER", `[0-9]+`))
	d.Rule("stmt", []string{"SET", "IDENT", "NUMBER"}, "")

	doc := d.NewDocument()
	// 1. Appending digits at EOF must NOT keep stale NUMBER "1".
	doc.Update("set a 1")
	require.Len(t, doc.Statements(), 1)
	assert.Equal(t, "1", doc.Statements()[0].Child(2).Token.Value)

	doc.Update("set a 10")
	require.Len(t, doc.Statements(), 1)
	assert.Equal(t, "10", doc.Statements()[0].Child(2).Token.Value)
	assert.Equal(t, 1, doc.Stats().Reparsed)

	// 2. Lookahead / expression extension across lines must reparse the prefix statement.
	exprDSL := New("doc_expr")
	require.NoError(t, exprDSL.Token("IDENT", `[a-z]+`))
	require.NoError(t, exprDSL.Token("NUMBER", `[0-9]+`))
	require.NoError(t, exprDSL.LiteralToken("EQ", "="))
	require.NoError(t, exprDSL.LiteralToken("PLUS", "+"))
	exprDSL.Rule("stmt", []string{"IDENT", "EQ", "expr"}, "")
	exprDSL.Expression("expr").
		Atom("NUMBER", "").
		Atom("IDENT", "").
		InfixLeft("PLUS", 10, "")

	edoc := exprDSL.NewDocument()
	edoc.Update("a = 1\nb = 2")
	require.Len(t, edoc.Statements(), 2)
	assert.Equal(t, "a = 1", edoc.Statements()[0].Text())

	// Replace second line start with "+ 3": statement 1 must extend to "a = 1 + 3"!
	edoc.Update("a = 1\n+ 3\nb = 2")
	require.Len(t, edoc.Statements(), 2)
	assert.Equal(t, "a = 1 + 3", edoc.Statements()[0].Text())
	assert.Equal(t, "b = 2", edoc.Statements()[1].Text())
}

func TestBug02LeadingWhitespacePreservesLineAndColumn(t *testing.T) {
	d := New("ws_offset")
	require.NoError(t, d.KeywordToken("SET", "set"))
	require.NoError(t, d.Token("IDENT", `[a-z]+`))
	d.Rule("stmt", []string{"SET", "IDENT"}, "")

	// Leading newlines/spaces before lexical error must report Line 3, Column 5.
	code := "\n\n    @bad"
	_, err := d.Parse(code)
	require.Error(t, err)
	pe, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, 3, pe.Line)
	assert.Equal(t, 5, pe.Column)

	// Span on valid AST must index the exact slice in the original untrimmed input.
	valid := "\n  set abc\n"
	node, err := d.ParseAST(valid)
	require.NoError(t, err)
	assert.Equal(t, "set abc", valid[node.Span.Start:node.Span.End])
}

func TestBug03DiagnosticsAndCompletionsAcrossRepeatedMemoizedSubrules(t *testing.T) {
	d := New("memo_diag")
	require.NoError(t, d.KeywordToken("IF", "if"))
	require.NoError(t, d.KeywordToken("THEN", "then"))
	require.NoError(t, d.Token("IDENT", `[a-z]+`))
	d.Rule("stmt", []string{"IF", "cond", "THEN", "IDENT"}, "")
	d.Rule("cond", []string{"IDENT"}, "")

	// First statement fails mid-way after caching a failure at 'cond' or 'THEN';
	// second statement also fails and must report precise expected tokens.
	diags := d.Diagnostics("if then x\nif y z")
	require.Len(t, diags, 2)
	assert.Contains(t, diags[0].Message, "expected IDENT")
	assert.Contains(t, diags[1].Message, "expected THEN")
}

func TestBug04IndirectLeftRecursionKeepsSubruleMemoFast(t *testing.T) {
	d := New("indirect_perf")
	require.NoError(t, d.Token("N", `[0-9]+`))
	require.NoError(t, d.LiteralToken("PLUS", "+"))
	require.NoError(t, d.LiteralToken("LPAREN", "("))
	require.NoError(t, d.LiteralToken("RPAREN", ")"))

	// Indirect LR between a <-> b, referencing independent expensive subrule 'atom'.
	d.Rule("a", []string{"b", "PLUS", "atom"}, "")
	d.Rule("a", []string{"atom"}, "")
	d.Rule("b", []string{"a", "PLUS", "atom"}, "")
	d.Rule("b", []string{"atom"}, "")

	d.Rule("atom", []string{"LPAREN", "atom", "RPAREN", "X"}, "")
	d.Rule("atom", []string{"LPAREN", "atom", "RPAREN"}, "")
	d.Rule("atom", []string{"N"}, "")
	require.NoError(t, d.LiteralToken("X", "!"))

	// Previously 22 nested parens inside indirect LR hung exponentially (>30s);
	// with selective cycleDependent memoization it finishes in <5ms.
	nested := strings.Repeat("(", 22) + "1" + strings.Repeat(")", 22) + " + 2 + 3 + 4"
	start := time.Now()
	_, err := d.ParseAST(nested)
	require.NoError(t, err)
	assert.Less(t, time.Since(start), 2*time.Second)
}

func TestBug05CompiledDSLReentrantNestedUseNoDeadlock(t *testing.T) {
	d := New("reentrant")
	require.NoError(t, d.Token("NUM", `[0-9]+`))
	require.NoError(t, d.KeywordToken("EVAL", "eval"))
	d.Rule("expr", []string{"EVAL", "NUM"}, "evalInner")
	d.Rule("expr", []string{"NUM"}, "num")

	var compiled *CompiledDSL
	d.Action("num", func(args []interface{}) (interface{}, error) {
		n, err := strconv.Atoi(args[0].(string))
		if err != nil {
			return nil, err
		}
		if mult, ok := d.GetContext("mult").(int); ok && mult > 0 {
			return n * mult, nil
		}
		return n, nil
	})
	d.Action("evalInner", func(args []interface{}) (interface{}, error) {
		// Re-entering compiled.Use on the same goroutine previously deadlocked!
		inner, err := compiled.Use(args[1].(string), map[string]interface{}{"mult": 10})
		if err != nil {
			return nil, err
		}
		return inner.GetOutput(), nil
	})

	var err error
	compiled, err = d.Build()
	require.NoError(t, err)

	res, err := compiled.Use("eval 7", map[string]interface{}{"mult": 2})
	require.NoError(t, err)
	assert.Equal(t, 70, res.GetOutput())
}

func TestBug06PunctuationKeywordsAndLosslessDeclarativeV2(t *testing.T) {
	d := New("v2_lossless")
	require.NoError(t, d.KeywordToken("STATUS", "$status"))
	require.NoError(t, d.KeywordToken("EQ", "=="))
	require.NoError(t, d.Token("CASE_SENSITIVE_A", "a"))
	require.NoError(t, d.Token("NUM", `[0-9]+`))
	require.NoError(t, d.SkipToken("COMMENT", `//[^\r\n]*`))
	d.SetSignificantNewlines(true)

	d.RuleWithError(" program ", []string{"check"}, "", "")
	// Rebuild cleanly with explicit start rule, expression, and hint:
	d = New("v2_lossless")
	require.NoError(t, d.KeywordToken("STATUS", "$status"))
	require.NoError(t, d.KeywordToken("EQ", "=="))
	require.NoError(t, d.Token("LOWER_A", "a"))
	require.NoError(t, d.Token("NUM", `[0-9]+`))
	require.NoError(t, d.SkipToken("COMMENT", `//[^\r\n]*`))
	d.RuleWithError("stmt", []string{"STATUS", "EQ", "expr", "LOWER_A"}, "check", "use '$status == <num> a'")
	d.Expression("expr").Atom("NUM", "num")
	d.SetStartRule("stmt")

	// "$status==200 a" works (punctuation keyword boundaries!).
	_, err := d.ParseAST("$status==200 a // ok")
	require.NoError(t, err)

	// Export to YAML and reload: LOWER_A must remain case-sensitive regex, NOT become KeywordToken!
	yamlBytes, err := d.SaveToYAML()
	require.NoError(t, err)

	reloaded, err := LoadFromYAML(yamlBytes)
	require.NoError(t, err)

	_, err = reloaded.ParseAST("$status == 200 a // still works")
	require.NoError(t, err)

	// Uppercase "A" must fail because LOWER_A was Token("LOWER_A", "a"), not KeywordToken!
	_, err = reloaded.ParseAST("$status == 200 A")
	assert.Error(t, err)

	// Error hint must also survive YAML round-trip!
	_, err = reloaded.ParseAST("$status 200")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "use '$status == <num> a'")
}

func TestBug07And08MultilineAndStatementsDiagnostics(t *testing.T) {
	d := New("multi_bugs")
	require.NoError(t, d.KeywordToken("SET", "set"))
	require.NoError(t, d.Token("NUM", `[0-9]+`))
	d.Rule("stmt", []string{"SET", "NUM"}, "")

	// Statement 10 error must say "in statement 10:", not "in statement \n:"!
	stmts := make([]string, 10)
	for i := 0; i < 9; i++ {
		stmts[i] = "set 1"
	}
	stmts[9] = "set bad"
	_, err := d.ParseStatements(strings.Join(stmts, "; "))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "in statement 10:")

	// ParseMultiline DetailedError must show the context line and indented column!
	multi := "set 1\n    set bad\nset 3"
	_, err = d.ParseMultiline(multi)
	require.Error(t, err)
	detailed := GetDetailedError(err)
	assert.Contains(t, detailed, "line 2, column 9")
	assert.Contains(t, detailed, "    set bad")

	// ParseStream DetailedError on line > 1 must also display context line!
	err = d.ParseStream(strings.NewReader("set 1\n  set bad\n"), nil)
	require.Error(t, err)
	assert.Contains(t, GetDetailedError(err), "  set bad")
}

func TestBug09UTF8ColumnsAndLexerErrors(t *testing.T) {
	d := New("utf8_test")
	require.NoError(t, d.Token("WORD", `[a-zA-ZáéíóúñÁÉÍÓÚÑ]+`))
	d.Rule("w", []string{"WORD"}, "")

	// Multibyte rune '🚀' must not be corrupted into Latin-1 byte Mojibake.
	_, err := d.Parse("mañana 🚀")
	require.Error(t, err)
	pe, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Contains(t, pe.Message, "🚀")
	assert.Equal(t, 1, pe.Line)
	assert.Equal(t, 8, pe.Column) // m(1) a(2) ñ(3) a(4) n(5) a(6) ' '(7) 🚀(8)
}

func TestBug11BlockParserNestedElseAndSelfGroupProductivity(t *testing.T) {
	d := New("block_nested")
	bp := NewBlockParser(d)
	src := strings.Join([]string{
		"if outer then",
		"  if inner then",
		"    act1",
		"  else",
		"    act2",
		"  endif",
		"else",
		"  act3",
		"endif",
	}, "\n")
	out := bp.ProcessBlocks(src)
	assert.Equal(t, "if outer then (if inner then (act1) else (act2)) else (act3)", out)

	// Self-referencing group with zero atoms must warn as non-productive!
	badExpr := New("self_group")
	require.NoError(t, badExpr.LiteralToken("LP", "("))
	require.NoError(t, badExpr.LiteralToken("RP", ")"))
	badExpr.Expression("expr").Group("LP", "expr", "RP")
	warnings, valErr := badExpr.Validate()
	require.NoError(t, valErr)
	assert.Contains(t, strings.Join(warnings, "\n"), "non-productive")
}

// --- New Features (#1, #3, #4, #5, #7, #8, #9) ------------------------------

func TestFeature4SkipTokensTriviaAndSignificantNewlines(t *testing.T) {
	d := New("py_like").
		Tokens(func(t *TokenSet) {
			t.Keyword("SET", "set")
			t.Regex("IDENT", `[a-z]+`)
			t.Regex("NUM", `[0-9]+`)
			t.LineComment("#")
			t.BlockComment("/*", "*/")
		}).
		WithSignificantNewlines(true)

	d.Rule("program", []string{"stmt"}, "single")
	d.Rule("program", []string{"program", "NEWLINE", "stmt"}, "append")
	d.Rule("stmt", []string{"SET", "IDENT", "NUM"}, "set")

	code := "set a 1 # line comment\nset b /* inline block */ 2"
	toks, trivia, err := d.TokenizeWithTrivia(code)
	require.NoError(t, err)
	assert.Len(t, trivia, 2)
	assert.Equal(t, "# line comment", trivia[0].Value)
	assert.Equal(t, "/* inline block */", trivia[1].Value)
	assert.Len(t, toks, 7) // SET a 1 NEWLINE SET b 2

	_, err = d.ParseAST(code)
	require.NoError(t, err)
}

func TestFeature5FullPrattExpressionEngine(t *testing.T) {
	d := New("pratt_pro").
		Tokens(func(ts *TokenSet) {
			ts.Keyword("TRUE", "true")
			ts.Keyword("FALSE", "false")
			ts.Regex("NUMBER", `[0-9]+`)
			ts.Regex("IDENT", `[a-zA-Z_][a-zA-Z0-9_]*`)
			ts.Literal("PLUS", "+")
			ts.Literal("BANG", "!")
			ts.Literal("EQ", "==")
			ts.Literal("LT", "<")
			ts.Literal("QMARK", "?")
			ts.Literal("COLON", ":")
			ts.Literal("LPAREN", "(")
			ts.Literal("RPAREN", ")")
			ts.Literal("LBRACK", "[")
			ts.Literal("RBRACK", "]")
			ts.Literal("COMMA", ",")
			ts.Literal("DOT", ".")
		})

	d.Expression("expr").
		Atom("NUMBER", "num").
		Atom("TRUE", "boolTrue").
		Atom("FALSE", "boolFalse").
		Atom("IDENT", "ident").
		Group("LPAREN", "expr", "RPAREN").
		Call("LPAREN", "RPAREN", "COMMA", 80, "call").
		Index("LBRACK", "expr", "RBRACK", 80, "index").
		Member("DOT", "IDENT", 80, "member").
		Postfix("BANG", 70, "fact").
		InfixLeft("PLUS", 40, "add").
		InfixNonAssoc("LT", 20, "lt").
		InfixNonAssoc("EQ", 20, "eq").
		Ternary("QMARK", "COLON", 10, "ternary")

	d.Action("num", Action1(strconv.Atoi))
	d.Action("boolTrue", func(_ []interface{}) (interface{}, error) { return true, nil })
	d.Action("boolFalse", func(_ []interface{}) (interface{}, error) { return false, nil })
	d.Action("ident", func(args []interface{}) (interface{}, error) {
		switch args[0].(string) {
		case "arr":
			return []int{10, 20, 30}, nil
		case "obj":
			return map[string]int{"val": 5}, nil
		default:
			return args[0].(string), nil
		}
	})
	d.Action("fact", func(args []interface{}) (interface{}, error) {
		n := args[0].(int)
		res := 1
		for i := 2; i <= n; i++ {
			res *= i
		}
		return res, nil
	})
	d.Action("add", Action3(func(l int, _ string, r int) (int, error) { return l + r, nil }))
	d.Action("lt", Action3(func(l int, _ string, r int) (bool, error) { return l < r, nil }))
	d.Action("eq", func(args []interface{}) (interface{}, error) { return args[0] == args[2], nil })
	d.Action("call", func(args []interface{}) (interface{}, error) {
		fn := args[0].(string)
		callArgs := args[1].([]interface{})
		if fn == "max" {
			best := callArgs[0].(int)
			for _, a := range callArgs[1:] {
				if a.(int) > best {
					best = a.(int)
				}
			}
			return best, nil
		}
		return 0, fmt.Errorf("unknown fn %s", fn)
	})
	d.Action("index", func(args []interface{}) (interface{}, error) {
		slice := args[0].([]int)
		idx := args[2].(int)
		return slice[idx], nil
	})
	d.Action("member", func(args []interface{}) (interface{}, error) {
		m := args[0].(map[string]int)
		field := args[2].(string)
		return m[field], nil
	})
	// Lazy ternary via NodeAction!
	d.NodeAction("ternary", func(ctx *EvalContext, n *Node) (interface{}, error) {
		cond, err := ctx.Eval(n.Child(0))
		if err != nil {
			return nil, err
		}
		if cond.(bool) {
			return ctx.Eval(n.Child(2))
		}
		return ctx.Eval(n.Child(4))
	})

	// Combined expression: max(3!, arr[1]) + obj.val < 30 ? 100 : 200
	// 3! = 6, arr[1] = 20 -> max(6, 20) = 20; + obj.val (5) = 25; 25 < 30 => true => 100!
	res, trace, err := d.ParseTrace("max(3!, arr[1]) + obj.val < 30 ? 100 : 200")
	require.NoError(t, err)
	assert.Equal(t, 100, res.GetOutput())

	// Verify Trace shows the untaken else branch ('200') as [SKIPPED]!
	prettyTrace := trace.Pretty()
	assert.Contains(t, prettyTrace, "[SKIPPED]")

	// Non-associative chaining (1 < 2 < 3) MUST fail!
	_, err = d.Parse("1 < 2 < 3")
	assert.Error(t, err)

	// Parenthesized ((1 < 2) == true) succeeds!
	res2, err := d.Parse("(1 < 2) == true ? 1 : 0")
	require.NoError(t, err)
	assert.Equal(t, 1, res2.GetOutput())
}

func TestFeature7ExecutionSandboxContextAndGasQuota(t *testing.T) {
	d := New("sandbox")
	require.NoError(t, d.Token("NUM", `[0-9]+`))
	require.NoError(t, d.LiteralToken("PLUS", "+"))
	d.Expression("expr").
		Atom("NUM", "num").
		InfixLeft("PLUS", 10, "add")

	d.ContextAction("num", func(ctx *EvalContext, args []interface{}) (interface{}, error) {
		n, _ := strconv.Atoi(args[0].(string))
		if bonus, ok := ctx.Get("bonus").(int); ok {
			n += bonus
		}
		return n, nil
	})
	d.Action("add", Action3(func(l int, _ string, r int) (int, error) { return l + r, nil }))

	// 1. Step quota exceeded returns ErrStepLimitExceeded.
	_, err := d.ParseWithOptions("1 + 2 + 3 + 4 + 5", EvalOptions{MaxSteps: 4})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrStepLimitExceeded))

	// 2. Context cancellation returns context.Canceled.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = d.ParseWithOptions("1 + 2", EvalOptions{Context: ctx})
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))

	// 3. Call-scoped Vars are isolated and work without mutating persistent context.
	res, err := d.ParseWithOptions("10 + 20", EvalOptions{
		MaxSteps: 100,
		Vars:     map[string]interface{}{"bonus": 1},
	})
	require.NoError(t, err)
	assert.Equal(t, 32, res.GetOutput())
	assert.Nil(t, d.GetContext("bonus"))
}

func TestFeature9FormatterIdempotencyAndTrivia(t *testing.T) {
	d := New("fmt_dsl").
		Tokens(func(ts *TokenSet) {
			ts.Keyword("IF", "if")
			ts.Regex("IDENT", `[a-z]+`)
			ts.Regex("NUM", `[0-9]+`)
			ts.Literal("LBRACE", "{")
			ts.Literal("RBRACE", "}")
			ts.Literal("EQ", "=")
			ts.LineComment("//")
		}).
		WithRuleLayout("block", RuleLayout{IndentBlock: true})

	d.Rule("stmt", []string{"IF", "IDENT", "block"}, "")
	d.Rule("stmt", []string{"IDENT", "EQ", "NUM"}, "")
	d.Rule("block", []string{"LBRACE", "stmt", "RBRACE"}, "")

	messy := "if   flag   {   x=42   } // check flag\n"
	formatted, err := d.Format(messy)
	require.NoError(t, err)
	expected := "if flag {\n  x = 42\n} // check flag\n"
	assert.Equal(t, expected, formatted)

	// Idempotency: Format(Format(x)) == Format(x)
	again, err := d.Format(formatted)
	require.NoError(t, err)
	assert.Equal(t, formatted, again)
}

func TestFeature1And3AIGuardrailsAndGrammarTesting(t *testing.T) {
	d := New("guardrails").
		Tokens(func(ts *TokenSet) {
			ts.Keyword("ALLOW", "allow")
			ts.Keyword("DENY", "deny")
			ts.Keyword("WHEN", "when")
			ts.Regex("IDENT", `[a-z]+`)
			ts.Regex("NUM", `[0-9]+`)
			ts.Literal("GT", ">")
		})
	d.Rule("policy", []string{"action_kw", "IDENT", "WHEN", "IDENT", "GT", "NUM"}, "")
	d.Rule("action_kw", []string{"ALLOW"}, "")
	d.Rule("action_kw", []string{"DENY"}, "")

	// 1. GBNF export
	gbnf := d.ToGBNF()
	assert.Contains(t, gbnf, "root ::= ws policy ws")
	assert.Contains(t, gbnf, "tok-allow ::=")

	// 2. AI Prompt Spec + RepairPrompt
	spec := d.AIPromptSpec()
	assert.Contains(t, spec, "## Valid Examples")
	assert.Contains(t, spec, "## GBNF Grammar")

	repair, clean := d.RepairPrompt("allow user when score >")
	assert.False(t, clean)
	assert.Contains(t, repair, "Valid tokens here")
	assert.Contains(t, repair, "NUM")

	_, clean2 := d.RepairPrompt("allow user when score > 80")
	assert.True(t, clean2)

	// 3. GenerateSamples + FuzzGrammar + RunTestSuite
	samples := d.GenerateSamples(7, 5, 4)
	assert.NotEmpty(t, samples)
	for _, s := range samples {
		_, err := d.ParseAST(s)
		assert.NoError(t, err, "sample failed to parse: %q", s)
	}

	fuzzRep := d.FuzzGrammar(99, 80)
	assert.Empty(t, fuzzRep.Failures)

	suiteRep := d.RunTestSuite([]DSLTestCase{
		{Name: "valid allow", Input: "allow admin when level > 5"},
		{Name: "invalid missing rhs", Input: "deny guest when level >", ExpectError: "expected NUM"},
	})
	assert.Equal(t, 2, suiteRep.Passed)
	assert.Equal(t, 0, suiteRep.Failed)
}
