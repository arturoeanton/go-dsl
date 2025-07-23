package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

func main() {
	fmt.Println("=== DSL Declarative Examples ===")
	fmt.Println()

	// Example 1: Load DSL from YAML file
	fmt.Println("1. Loading Calculator DSL from YAML file...")
	yamlDSL, err := dslbuilder.LoadFromYAMLFile("calculator.yaml")
	if err != nil {
		log.Printf("Error loading YAML: %v", err)
	} else {
		// Register actions for the calculator
		yamlDSL.Action("add", func(args []interface{}) (interface{}, error) {
			a, _ := strconv.Atoi(args[0].(string))
			b, _ := strconv.Atoi(args[2].(string))
			return a + b, nil
		})
		yamlDSL.Action("subtract", func(args []interface{}) (interface{}, error) {
			a, _ := strconv.Atoi(args[0].(string))
			b, _ := strconv.Atoi(args[2].(string))
			return a - b, nil
		})
		yamlDSL.Action("multiply", func(args []interface{}) (interface{}, error) {
			a, _ := strconv.Atoi(args[0].(string))
			b, _ := strconv.Atoi(args[2].(string))
			return a * b, nil
		})
		yamlDSL.Action("divide", func(args []interface{}) (interface{}, error) {
			a, _ := strconv.Atoi(args[0].(string))
			b, _ := strconv.Atoi(args[2].(string))
			if b == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			return a / b, nil
		})

		// Test the YAML-loaded DSL
		testCalculator(yamlDSL, "YAML-loaded")
	}

	fmt.Println()

	// Example 2: Create DSL with Builder Pattern
	fmt.Println("2. Creating Calculator DSL with Builder Pattern...")
	builderDSL := dslbuilder.New("CalculatorBuilder").
		WithToken("NUMBER", "[0-9]+").
		WithToken("PLUS", "\\+").
		WithToken("MINUS", "-").
		WithToken("MULTIPLY", "\\*").
		WithToken("DIVIDE", "/").
		WithRule("expr", []string{"NUMBER", "PLUS", "NUMBER"}, "add").
		WithRule("expr", []string{"NUMBER", "MINUS", "NUMBER"}, "subtract").
		WithRule("expr", []string{"NUMBER", "MULTIPLY", "NUMBER"}, "multiply").
		WithRule("expr", []string{"NUMBER", "DIVIDE", "NUMBER"}, "divide").
		WithAction("add", func(args []interface{}) (interface{}, error) {
			a, _ := strconv.Atoi(args[0].(string))
			b, _ := strconv.Atoi(args[2].(string))
			return a + b, nil
		}).
		WithAction("subtract", func(args []interface{}) (interface{}, error) {
			a, _ := strconv.Atoi(args[0].(string))
			b, _ := strconv.Atoi(args[2].(string))
			return a - b, nil
		}).
		WithAction("multiply", func(args []interface{}) (interface{}, error) {
			a, _ := strconv.Atoi(args[0].(string))
			b, _ := strconv.Atoi(args[2].(string))
			return a * b, nil
		}).
		WithAction("divide", func(args []interface{}) (interface{}, error) {
			a, _ := strconv.Atoi(args[0].(string))
			b, _ := strconv.Atoi(args[2].(string))
			if b == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			return a / b, nil
		})

	// Test the Builder Pattern DSL
	testCalculator(builderDSL, "Builder Pattern")

	fmt.Println()

	// Example 3: Save DSL to JSON
	fmt.Println("3. Saving DSL to JSON...")
	jsonData, err := builderDSL.SaveToJSON()
	if err != nil {
		log.Printf("Error saving to JSON: %v", err)
	} else {
		fmt.Println("JSON configuration:")
		fmt.Println(string(jsonData))

		// Save to file
		err = builderDSL.SaveToJSONFile("calculator.json")
		if err != nil {
			log.Printf("Error saving JSON file: %v", err)
		} else {
			fmt.Println("✅ Saved to calculator.json")
		}
	}

	fmt.Println()

	// Example 4: Traditional API (backward compatibility test)
	fmt.Println("4. Testing Traditional API (backward compatibility)...")
	traditionalDSL := dslbuilder.New("TraditionalCalculator")

	// Using traditional methods
	traditionalDSL.Token("NUMBER", "[0-9]+")
	traditionalDSL.Token("PLUS", "\\+")
	traditionalDSL.Rule("expr", []string{"NUMBER", "PLUS", "NUMBER"}, "add")
	traditionalDSL.Action("add", func(args []interface{}) (interface{}, error) {
		a, _ := strconv.Atoi(args[0].(string))
		b, _ := strconv.Atoi(args[2].(string))
		return a + b, nil
	})

	result, err := traditionalDSL.Parse("5 + 3")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("✅ Traditional API: 5 + 3 = %v\n", result.GetOutput())
	}

	fmt.Println()
	fmt.Println("=== All examples completed successfully! ===")
}

func testCalculator(dsl *dslbuilder.DSL, name string) {
	testCases := []string{
		"5 + 3",
		"10 - 4",
		"6 * 7",
		"20 / 4",
	}

	fmt.Printf("\nTesting %s DSL:\n", name)
	for _, expr := range testCases {
		result, err := dsl.Parse(expr)
		if err != nil {
			fmt.Printf("❌ %s: Error: %v\n", expr, err)
		} else {
			fmt.Printf("✅ %s = %v\n", expr, result.GetOutput())
		}
	}
}
