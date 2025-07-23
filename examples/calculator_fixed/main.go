package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

func main() {
	// Create a calculator DSL
	calc := dslbuilder.New("Calculator")

	// Define tokens
	calc.Token("NUMBER", "[0-9]+")
	calc.Token("PLUS", "\\+")
	calc.Token("MINUS", "-")
	calc.Token("MULTIPLY", "\\*")
	calc.Token("DIVIDE", "/")
	calc.Token("LPAREN", "\\(")
	calc.Token("RPAREN", "\\)")

	// Use simplified grammar that works with the improved parser
	calc.Rule("expression", []string{"term"}, "passthrough")
	calc.Rule("expression", []string{"expression", "PLUS", "term"}, "add")
	calc.Rule("expression", []string{"expression", "MINUS", "term"}, "subtract")

	calc.Rule("term", []string{"factor"}, "passthrough")
	calc.Rule("term", []string{"term", "MULTIPLY", "factor"}, "multiply")
	calc.Rule("term", []string{"term", "DIVIDE", "factor"}, "divide")

	calc.Rule("factor", []string{"NUMBER"}, "number")
	calc.Rule("factor", []string{"LPAREN", "expression", "RPAREN"}, "parentheses")

	// Define actions
	calc.Action("passthrough", func(args []interface{}) (interface{}, error) {
		if len(args) > 0 {
			return args[0], nil
		}
		return 0, nil
	})

	calc.Action("add", func(args []interface{}) (interface{}, error) {
		if len(args) >= 3 {
			left, _ := toFloat(args[0])
			right, _ := toFloat(args[2])
			return left + right, nil
		}
		return 0, fmt.Errorf("invalid add operation")
	})

	calc.Action("subtract", func(args []interface{}) (interface{}, error) {
		if len(args) >= 3 {
			left, _ := toFloat(args[0])
			right, _ := toFloat(args[2])
			return left - right, nil
		}
		return 0, fmt.Errorf("invalid subtract operation")
	})

	calc.Action("multiply", func(args []interface{}) (interface{}, error) {
		if len(args) >= 3 {
			left, _ := toFloat(args[0])
			right, _ := toFloat(args[2])
			return left * right, nil
		}
		return 0, fmt.Errorf("invalid multiply operation")
	})

	calc.Action("divide", func(args []interface{}) (interface{}, error) {
		if len(args) >= 3 {
			left, _ := toFloat(args[0])
			right, _ := toFloat(args[2])
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return left / right, nil
		}
		return 0, fmt.Errorf("invalid divide operation")
	})

	calc.Action("number", func(args []interface{}) (interface{}, error) {
		if len(args) > 0 {
			return strconv.ParseFloat(args[0].(string), 64)
		}
		return 0.0, fmt.Errorf("no number provided")
	})

	calc.Action("parentheses", func(args []interface{}) (interface{}, error) {
		if len(args) >= 3 {
			return args[1], nil // Return the expression inside parentheses
		}
		return 0, nil
	})

	// Test expressions
	expressions := []string{
		"5",
		"3 + 2",
		"8 - 3",
		"10 + 5 - 3",
		"2 + 3",
		"10 - 5",
		"4 * 6",
		"20 / 4",
		"2 + 3 * 4",
		"(2 + 3) * 4",
		"10 + 20 - 5",
		"100 / 10 / 2",
	}

	fmt.Println("Calculator DSL Demo (Fixed)")
	fmt.Println("============================")

	for _, expr := range expressions {
		result, err := calc.Parse(expr)
		if err != nil {
			log.Printf("Error parsing '%s': %v\n", expr, err)
			continue
		}
		fmt.Printf("%s = %v\n", expr, result.GetOutput())
	}
}

func toFloat(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case int:
		return float64(val), nil
	case string:
		return strconv.ParseFloat(val, 64)
	default:
		return 0, fmt.Errorf("cannot convert to float: %v", v)
	}
}
