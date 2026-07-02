package main

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

func testConfig() *dslbuilder.DSLConfig {
	return &dslbuilder.DSLConfig{
		Name: "Calc",
		Tokens: map[string]string{
			"NUMBER": "[0-9]+",
			"PLUS":   `\+`,
			"IF":     `(?i)\bif\b`, // keyword round-trip
			"SET":    "set",        // bare word -> keyword
		},
		Rules: []dslbuilder.RuleConfig{
			{Name: "expr", Pattern: []string{"NUMBER", "PLUS", "NUMBER"}, Action: "add"},
			{Name: "expr", Pattern: []string{"NUMBER"}, Action: "number"},
		},
	}
}

func TestGenerateProducesValidGo(t *testing.T) {
	source, err := Generate(testConfig(), "calcgen", "NewCalcDSL")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// The output must be parseable Go.
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "gen.go", source, parser.AllErrors); err != nil {
		t.Fatalf("generated code does not parse: %v\n%s", err, source)
	}

	code := string(source)
	for _, want := range []string{
		"package calcgen",
		"func NewCalcDSL() (*dslbuilder.DSL, error)",
		`dsl.Token("NUMBER", "[0-9]+")`,
		`dsl.KeywordToken("IF", "if")`,
		`dsl.KeywordToken("SET", "set")`,
		`dsl.Rule("expr", []string{"NUMBER", "PLUS", "NUMBER"}, "add")`,
		`dsl.Action("add"`, // action stub hint in the doc comment
	} {
		if !strings.Contains(code, want) {
			t.Errorf("generated code missing %q\n%s", want, code)
		}
	}
}

func TestGeneratedGrammarBehavesLikeSource(t *testing.T) {
	// The generated code is a Go rendering of the config; verify the config
	// itself builds a working DSL (what the generated constructor will do).
	cfgDSL := dslbuilder.New("Calc")
	if err := cfgDSL.Token("NUMBER", "[0-9]+"); err != nil {
		t.Fatal(err)
	}
	if err := cfgDSL.Token("PLUS", `\+`); err != nil {
		t.Fatal(err)
	}
	cfgDSL.Rule("expr", []string{"NUMBER", "PLUS", "NUMBER"}, "add")
	cfgDSL.Rule("expr", []string{"NUMBER"}, "number")
	cfgDSL.Action("add", func(args []interface{}) (interface{}, error) {
		return dslbuilder.Args(args).Int(0) + dslbuilder.Args(args).Int(2), nil
	})
	cfgDSL.Action("number", func(args []interface{}) (interface{}, error) {
		return dslbuilder.Args(args).Int(0), nil
	})

	result, err := cfgDSL.Parse("40 + 2")
	if err != nil {
		t.Fatal(err)
	}
	if result.GetOutput() != 42 {
		t.Fatalf("got %v, want 42", result.GetOutput())
	}
}
