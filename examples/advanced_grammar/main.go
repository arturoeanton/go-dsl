package main

import (
	"fmt"
	"strings"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

func main() {
	// Example 1: Operator Precedence and Associativity
	fmt.Println("=== Example 1: Operator Precedence and Associativity ===")
	demonstratePrecedence()

	// Example 2: Kleene Star (Zero or More)
	fmt.Println("\n=== Example 2: Kleene Star (Zero or More) ===")
	demonstrateKleeneStar()

	// Example 3: Kleene Plus (One or More)
	fmt.Println("\n=== Example 3: Kleene Plus (One or More) ===")
	demonstrateKleenePlus()

	// Example 4: Lookahead/Lookbehind
	fmt.Println("\n=== Example 4: Lookahead/Lookbehind ===")
	demonstrateLookaround()
}

func demonstratePrecedence() {
	calc := dslbuilder.New("PrecedenceCalc")

	// Define tokens
	calc.Token("NUMBER", "[0-9]+")
	calc.Token("PLUS", "\\+")
	calc.Token("MINUS", "-")
	calc.Token("MULTIPLY", "\\*")
	calc.Token("DIVIDE", "/")
	calc.Token("POWER", "\\^")
	calc.Token("LPAREN", "\\(")
	calc.Token("RPAREN", "\\)")

	// Define rules with precedence (higher number = higher precedence)
	// Level 1: Addition and Subtraction (lowest precedence, left associative)
	calc.RuleWithPrecedence("expr", []string{"expr", "PLUS", "term"}, "add", 1, "left")
	calc.RuleWithPrecedence("expr", []string{"expr", "MINUS", "term"}, "subtract", 1, "left")
	calc.Rule("expr", []string{"term"}, "passthrough")

	// Level 2: Multiplication and Division (medium precedence, left associative)
	calc.RuleWithPrecedence("term", []string{"term", "MULTIPLY", "factor"}, "multiply", 2, "left")
	calc.RuleWithPrecedence("term", []string{"term", "DIVIDE", "factor"}, "divide", 2, "left")
	calc.Rule("term", []string{"factor"}, "passthrough")

	// Level 3: Exponentiation (highest precedence, right associative)
	calc.RuleWithPrecedence("factor", []string{"primary", "POWER", "factor"}, "power", 3, "right")
	calc.Rule("factor", []string{"primary"}, "passthrough")

	// Primary expressions
	calc.Rule("primary", []string{"NUMBER"}, "number")
	calc.Rule("primary", []string{"LPAREN", "expr", "RPAREN"}, "paren")

	// Define actions
	calc.Action("add", func(args []interface{}) (interface{}, error) {
		left := toFloat(args[0])
		right := toFloat(args[2])
		return left + right, nil
	})

	calc.Action("subtract", func(args []interface{}) (interface{}, error) {
		left := toFloat(args[0])
		right := toFloat(args[2])
		return left - right, nil
	})

	calc.Action("multiply", func(args []interface{}) (interface{}, error) {
		left := toFloat(args[0])
		right := toFloat(args[2])
		return left * right, nil
	})

	calc.Action("divide", func(args []interface{}) (interface{}, error) {
		left := toFloat(args[0])
		right := toFloat(args[2])
		if right == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return left / right, nil
	})

	calc.Action("power", func(args []interface{}) (interface{}, error) {
		base := toFloat(args[0])
		exp := toFloat(args[2])
		result := 1.0
		for i := 0; i < int(exp); i++ {
			result *= base
		}
		return result, nil
	})

	calc.Action("number", func(args []interface{}) (interface{}, error) {
		return toFloat(args[0]), nil
	})

	calc.Action("passthrough", func(args []interface{}) (interface{}, error) {
		return args[0], nil
	})

	calc.Action("paren", func(args []interface{}) (interface{}, error) {
		return args[1], nil
	})

	// Test expressions
	testExpressions := []string{
		"2 + 3 * 4",   // Should be 14 (not 20)
		"2 * 3 + 4",   // Should be 10
		"2 ^ 3 * 4",   // Should be 32 (8 * 4)
		"2 * 3 ^ 2",   // Should be 18 (2 * 9)
		"2 ^ 3 ^ 2",   // Should be 512 (2^9, right associative)
		"(2 + 3) * 4", // Should be 20
	}

	for _, expr := range testExpressions {
		result, err := calc.Parse(expr)
		if err != nil {
			fmt.Printf("Error parsing '%s': %v\n", expr, err)
		} else {
			fmt.Printf("%s = %v\n", expr, result.GetOutput())
		}
	}
}

func demonstrateKleeneStar() {
	// Create a DSL for space-separated words (zero or more)
	wordsDSL := dslbuilder.New("WordsDSL")

	// Tokens
	wordsDSL.Token("WORD", "[a-zA-Z]+")

	// Rules - start with sentence
	wordsDSL.Rule("sentence", []string{"words"}, "makeSentence")
	wordsDSL.Rule("sentence", []string{}, "emptySentence") // Allow empty

	// Words with Kleene star (zero or more)
	wordsDSL.RuleWithRepetition("words", "WORD", "words")

	// Actions
	wordsDSL.Action("makeSentence", func(args []interface{}) (interface{}, error) {
		return args[0], nil
	})

	wordsDSL.Action("emptySentence", func(args []interface{}) (interface{}, error) {
		return []string{}, nil
	})

	wordsDSL.Action("words_empty", func(args []interface{}) (interface{}, error) {
		return []string{}, nil
	})

	wordsDSL.Action("words_append", func(args []interface{}) (interface{}, error) {
		list := args[0].([]string)
		word := args[1].(string)
		return append(list, word), nil
	})

	// Test cases
	testCases := []string{
		"",                    // Empty
		"hello",               // Single word
		"hello world",         // Two words
		"the quick brown fox", // Multiple words
	}

	for _, test := range testCases {
		result, err := wordsDSL.Parse(test)
		if err != nil {
			fmt.Printf("Error parsing '%s': %v\n", test, err)
		} else {
			fmt.Printf("Words: %v\n", result.GetOutput())
		}
	}
}

func demonstrateKleenePlus() {
	// Create a DSL for identifier lists (one or more required)
	idDSL := dslbuilder.New("IdentifierDSL")

	// Tokens
	idDSL.Token("ID", "[a-zA-Z_][a-zA-Z0-9_]*")
	idDSL.Token("DOT", "\\.")

	// Rules - qualified name must have at least one identifier
	idDSL.Rule("qualified_name", []string{"identifiers"}, "makeName")

	// One or more identifiers separated by dots
	idDSL.RuleWithPlusRepetition("identifiers", "id_part", "ids")

	idDSL.Rule("id_part", []string{"ID"}, "id")
	idDSL.Rule("id_part", []string{"DOT", "ID"}, "dotId")

	// Actions
	idDSL.Action("makeName", func(args []interface{}) (interface{}, error) {
		parts := args[0].([]string)
		return strings.Join(parts, ""), nil
	})

	idDSL.Action("ids_single", func(args []interface{}) (interface{}, error) {
		return []string{args[0].(string)}, nil
	})

	idDSL.Action("ids_append", func(args []interface{}) (interface{}, error) {
		list := args[0].([]string)
		item := args[1].(string)
		return append(list, item), nil
	})

	idDSL.Action("id", func(args []interface{}) (interface{}, error) {
		return args[0].(string), nil
	})

	idDSL.Action("dotId", func(args []interface{}) (interface{}, error) {
		return "." + args[1].(string), nil
	})

	// Test cases
	testCases := []string{
		"variable",          // Single identifier
		"object.property",   // Two parts
		"com.example.Class", // Three parts
		"a.b.c.d.e",         // Many parts
	}

	for _, test := range testCases {
		result, err := idDSL.Parse(test)
		if err != nil {
			fmt.Printf("Error parsing '%s': %v\n", test, err)
		} else {
			fmt.Printf("Qualified name: %v\n", result.GetOutput())
		}
	}
}

func demonstrateLookaround() {
	// Create a DSL that demonstrates lookahead concept
	// Since Go doesn't support regex lookbehind, we'll demonstrate priority-based tokenization
	fmt.Println("Note: Go's regexp package doesn't support lookbehind assertions.")
	fmt.Println("Demonstrating priority-based token matching instead:")

	keywordDSL := dslbuilder.New("KeywordDSL")

	// Higher priority tokens (keywords) take precedence
	keywordDSL.KeywordToken("IF", "if")
	keywordDSL.KeywordToken("WHILE", "while")
	keywordDSL.KeywordToken("FOR", "for")

	// Lower priority generic identifier
	keywordDSL.Token("IDENTIFIER", "[a-zA-Z][a-zA-Z0-9]*")
	keywordDSL.Token("NUMBER", "[0-9]+")

	// Rules
	keywordDSL.Rule("statement", []string{"IF", "NUMBER"}, "ifStmt")
	keywordDSL.Rule("statement", []string{"WHILE", "NUMBER"}, "whileStmt")
	keywordDSL.Rule("statement", []string{"IDENTIFIER", "NUMBER"}, "assignment")

	// Actions
	keywordDSL.Action("ifStmt", func(args []interface{}) (interface{}, error) {
		return fmt.Sprintf("If statement with condition %s", args[1]), nil
	})

	keywordDSL.Action("whileStmt", func(args []interface{}) (interface{}, error) {
		return fmt.Sprintf("While loop with condition %s", args[1]), nil
	})

	keywordDSL.Action("assignment", func(args []interface{}) (interface{}, error) {
		return fmt.Sprintf("Assignment: %s = %s", args[0], args[1]), nil
	})

	// Test cases
	testCases := []string{
		"if 42",       // Should match as IF token (keyword)
		"while 10",    // Should match as WHILE token (keyword)
		"ifx 5",       // Should match as IDENTIFIER (not keyword)
		"counter 100", // Should match as IDENTIFIER
	}

	for _, test := range testCases {
		result, err := keywordDSL.Parse(test)
		if err != nil {
			fmt.Printf("Error parsing '%s': %v\n", test, err)
		} else {
			fmt.Printf("'%s' -> %v\n", test, result.GetOutput())
		}
	}
}

func toFloat(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case string:
		var f float64
		fmt.Sscanf(n, "%f", &f)
		return f
	default:
		return 0
	}
}
