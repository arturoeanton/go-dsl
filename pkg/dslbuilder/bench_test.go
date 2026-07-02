package dslbuilder

import (
	"strings"
	"testing"
)

// BenchmarkExpressionParse measures the Pratt parser end to end
// (tokenize + parse + eval) on a medium arithmetic expression.
func BenchmarkExpressionParse(b *testing.B) {
	d := fuzzCalc()
	input := "1 + 2 * (3 - 4) / 5 ^ 2 + 6 * 7 - (8 + 9) * 10"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := d.Parse(input); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkExpressionParseLong measures a long flat expression chain
// (500 operators), the worst case for naive precedence handling.
func BenchmarkExpressionParseLong(b *testing.B) {
	d := fuzzCalc()
	input := "1" + strings.Repeat(" + 2 * 3", 500)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := d.Parse(input); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkLeftRecursion measures the growing-seed algorithm on a
// 1000-element left-recursive list.
func BenchmarkLeftRecursion(b *testing.B) {
	d := New("bench")
	d.Token("A", "a")
	d.Token("B", "b")
	d.Rule("list", []string{"list", "A"}, "append")
	d.Rule("list", []string{"B"}, "base")
	d.Action("append", func(args []interface{}) (interface{}, error) { return args[0], nil })
	d.Action("base", func(args []interface{}) (interface{}, error) { return "b", nil })
	input := "b" + strings.Repeat(" a", 1000)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := d.Parse(input); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseASTOnly isolates the parsing phase (no evaluation).
func BenchmarkParseASTOnly(b *testing.B) {
	d := fuzzCalc()
	input := "1 + 2 * (3 - 4) / 5 ^ 2 + 6 * 7 - (8 + 9) * 10"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := d.ParseAST(input); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEvalOnly isolates the evaluation phase over a pre-built tree.
func BenchmarkEvalOnly(b *testing.B) {
	d := fuzzCalc()
	node, err := d.ParseAST("1 + 2 * (3 - 4) / 5 ^ 2 + 6 * 7 - (8 + 9) * 10")
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := d.Eval(node); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTokenizer isolates lexing.
func BenchmarkTokenizer(b *testing.B) {
	d := fuzzCalc()
	input := "1" + strings.Repeat(" + 2 * 3", 500)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := tokenizeInput(d.grammar, input, input); err != nil {
			b.Fatal(err)
		}
	}
}
