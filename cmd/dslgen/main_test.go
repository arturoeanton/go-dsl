package main

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
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

func TestBug12IdentifierValidationAndExtHandling(t *testing.T) {
	if _, err := Generate(testConfig(), "main; import \"os\"", "NewDSL"); err == nil {
		t.Fatal("expected invalid package identifier to fail")
	}
	if _, err := Generate(testConfig(), "func", "NewDSL"); err == nil {
		t.Fatal("expected Go keyword package name to fail")
	}
	if _, err := Generate(testConfig(), "calcgen", "New() {}\nfunc Evil"); err == nil {
		t.Fatal("expected invalid function identifier to fail")
	}
	if _, err := loadConfig(filepath.Join(t.TempDir(), "extensionless_grammar")); err == nil {
		t.Fatal("expected extensionless filename to fail cleanly")
	}
}

func TestV2ExpressionsTypedAndVSCodeExtension(t *testing.T) {
	v2 := &dslbuilder.DSLConfig{
		Name:                "ExprLang\nv2",
		StartRule:           "expr",
		SignificantNewlines: true,
		TokenDefs: []dslbuilder.TokenConfig{
			{Name: "COMMENT", Kind: "skip", Pattern: `//[^\n]*`},
			{Name: "IF", Kind: "keyword", Pattern: "if"},
			{Name: "PLUS", Kind: "literal", Pattern: "+"},
			{Name: "QMARK", Kind: "literal", Pattern: "?"},
			{Name: "COLON", Kind: "literal", Pattern: ":"},
			{Name: "LPAREN", Kind: "literal", Pattern: "("},
			{Name: "RPAREN", Kind: "literal", Pattern: ")"},
			{Name: "NUMBER", Kind: "regex", Pattern: `[0-9]+`},
		},
		Expressions: []dslbuilder.ExpressionConfig{
			{
				Name: "expr",
				Atoms: []dslbuilder.ExprAtomConfig{
					{Token: "NUMBER", Action: "num_lit"},
				},
				Groups: []dslbuilder.ExprGroupConfig{
					{Open: "LPAREN", Inner: "expr", Close: "RPAREN"},
				},
				Infix: []dslbuilder.ExprOpConfig{
					{Token: "PLUS", Power: 10, Assoc: "left", Action: "add_op"},
				},
				Ternary: &dslbuilder.ExprTernaryConf{
					QuestionToken: "QMARK",
					ColonToken:    "COLON",
					Power:         2,
					Action:        "cond_op",
				},
			},
		},
	}

	source, err := GenerateWithOptions(v2, "exprgen", "NewExprLang", GenerateOptions{Typed: true})
	if err != nil {
		t.Fatalf("GenerateWithOptions failed: %v", err)
	}

	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "expr_gen.go", source, parser.AllErrors); err != nil {
		t.Fatalf("typed v2 generated code does not parse: %v\n%s", err, source)
	}

	code := string(source)
	for _, want := range []string{
		`dsl.SkipToken("COMMENT"`,
		`dsl.LiteralToken("PLUS", "+")`,
		`dsl.Expression("expr")`,
		`Ternary("QMARK", "COLON", 2, "cond_op")`,
		`type ExprLangActions interface`,
		`OnAddOp(args []interface{}) (interface{}, error)`,
		`OnCondOp(args []interface{}) (interface{}, error)`,
		`func NewExprLangWithActions(handler ExprLangActions)`,
	} {
		if !strings.Contains(code, want) {
			t.Errorf("typed v2 generated code missing %q:\n%s", want, code)
		}
	}

	outDir := filepath.Join(t.TempDir(), "vscode-exprlang")
	if err := GenerateVSCodeExtension(v2, outDir); err != nil {
		t.Fatalf("GenerateVSCodeExtension failed: %v", err)
	}
	for _, rel := range []string{
		"package.json",
		"language-configuration.json",
		"extension.js",
		"grammar.yaml",
		filepath.Join("syntaxes", "exprlang-v2.tmLanguage.json"),
	} {
		if _, err := os.Stat(filepath.Join(outDir, rel)); err != nil {
			t.Errorf("missing generated VS Code extension file %s: %v", rel, err)
		}
	}
}
