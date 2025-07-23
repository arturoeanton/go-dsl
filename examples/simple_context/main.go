package main

import (
	"fmt"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

func main() {
	fmt.Println("=== go-dsl Context Examples ===")
	fmt.Println("Equivalent to r2lang's: q.use(\"query\", context)")
	fmt.Println()

	// Example 1: Simple variable access from context
	fmt.Println("1. Simple Variable Access")
	fmt.Println("------------------------")

	varDSL := dslbuilder.New("Variables")
	varDSL.KeywordToken("GET", "get")
	varDSL.Token("VAR", "[a-zA-Z_][a-zA-Z0-9_]*")
	varDSL.Rule("command", []string{"GET", "VAR"}, "getVariable")

	varDSL.Action("getVariable", func(args []interface{}) (interface{}, error) {
		varName := args[1].(string)
		value := varDSL.GetContext(varName)
		if value == nil {
			return fmt.Sprintf("Variable '%s' not found", varName), nil
		}
		return value, nil
	})

	// Using Use() method - equivalent to r2lang's q.use()
	context := map[string]interface{}{
		"name":  "Juan García",
		"age":   30,
		"city":  "Madrid",
		"score": 95.5,
	}

	queries := []string{"get name", "get age", "get city", "get score", "get missing"}
	for _, query := range queries {
		result, err := varDSL.Use(query, context)
		if err != nil {
			fmt.Printf("  %s -> Error: %v\n", query, err)
		} else {
			fmt.Printf("  %s -> %v\n", query, result.GetOutput())
		}
	}

	fmt.Println()

	// Example 2: Working with data arrays
	fmt.Println("2. Data Array Processing")
	fmt.Println("------------------------")

	dataDSL := dslbuilder.New("DataProcessor")
	dataDSL.KeywordToken("COUNT", "count")
	dataDSL.KeywordToken("LIST", "list")
	dataDSL.KeywordToken("SUM", "sum")
	dataDSL.Token("TABLE", "[a-zA-Z_][a-zA-Z0-9_]*")
	dataDSL.Rule("command", []string{"COUNT", "TABLE"}, "countRecords")
	dataDSL.Rule("command", []string{"LIST", "TABLE"}, "listRecords")
	dataDSL.Rule("command", []string{"SUM", "TABLE"}, "sumRecords")

	dataDSL.Action("countRecords", func(args []interface{}) (interface{}, error) {
		tableName := args[1].(string)
		data := dataDSL.GetContext(tableName)
		if data == nil {
			return 0, nil
		}

		if slice, ok := data.([]interface{}); ok {
			return len(slice), nil
		}
		if slice, ok := data.([]int); ok {
			return len(slice), nil
		}
		if slice, ok := data.([]string); ok {
			return len(slice), nil
		}
		return 0, fmt.Errorf("invalid data type for %s", tableName)
	})

	dataDSL.Action("listRecords", func(args []interface{}) (interface{}, error) {
		tableName := args[1].(string)
		data := dataDSL.GetContext(tableName)
		if data == nil {
			return []interface{}{}, nil
		}
		return data, nil
	})

	dataDSL.Action("sumRecords", func(args []interface{}) (interface{}, error) {
		tableName := args[1].(string)
		data := dataDSL.GetContext(tableName)
		if data == nil {
			return 0, nil
		}

		if numbers, ok := data.([]int); ok {
			sum := 0
			for _, n := range numbers {
				sum += n
			}
			return sum, nil
		}
		return 0, fmt.Errorf("can only sum integer arrays")
	})

	// Sample data
	users := []string{"Juan", "María", "Carlos", "Ana"}
	scores := []int{85, 92, 78, 95}
	ages := []int{28, 35, 42, 29}

	// Test with different datasets
	testCases := []struct {
		query   string
		context map[string]interface{}
		desc    string
	}{
		{"count users", map[string]interface{}{"users": users}, "Count users"},
		{"list users", map[string]interface{}{"users": users}, "List all users"},
		{"sum scores", map[string]interface{}{"scores": scores}, "Sum all scores"},
		{"count ages", map[string]interface{}{"ages": ages}, "Count ages"},
		{"sum ages", map[string]interface{}{"ages": ages}, "Sum all ages"},
	}

	for _, tc := range testCases {
		result, err := dataDSL.Use(tc.query, tc.context)
		if err != nil {
			fmt.Printf("  %s: Error: %v\n", tc.desc, err)
		} else {
			fmt.Printf("  %s: %v\n", tc.desc, result.GetOutput())
		}
	}

	fmt.Println()

	// Example 3: Complex data structures
	fmt.Println("3. Complex Data Structures")
	fmt.Println("--------------------------")

	type Person struct {
		Name string
		Age  int
		City string
	}

	complexDSL := dslbuilder.New("ComplexData")
	complexDSL.KeywordToken("FIND", "find")
	complexDSL.KeywordToken("NAME", "name")
	complexDSL.KeywordToken("AGE", "age")
	complexDSL.KeywordToken("CITY", "city")
	complexDSL.KeywordToken("IN", "in")
	complexDSL.Token("DATASET", "[a-zA-Z_][a-zA-Z0-9_]*")
	complexDSL.Rule("command", []string{"FIND", "NAME", "IN", "DATASET"}, "findName")
	complexDSL.Rule("command", []string{"FIND", "AGE", "IN", "DATASET"}, "findAge")
	complexDSL.Rule("command", []string{"FIND", "CITY", "IN", "DATASET"}, "findCity")

	complexDSL.Action("findName", func(args []interface{}) (interface{}, error) {
		dataset := args[3].(string)

		data := complexDSL.GetContext(dataset)
		if data == nil {
			return nil, fmt.Errorf("dataset '%s' not found", dataset)
		}

		people, ok := data.([]Person)
		if !ok {
			return nil, fmt.Errorf("invalid data type for %s", dataset)
		}

		var results []interface{}
		for _, person := range people {
			results = append(results, person.Name)
		}
		return results, nil
	})

	complexDSL.Action("findAge", func(args []interface{}) (interface{}, error) {
		dataset := args[3].(string)

		data := complexDSL.GetContext(dataset)
		if data == nil {
			return nil, fmt.Errorf("dataset '%s' not found", dataset)
		}

		people, ok := data.([]Person)
		if !ok {
			return nil, fmt.Errorf("invalid data type for %s", dataset)
		}

		var results []interface{}
		for _, person := range people {
			results = append(results, person.Age)
		}
		return results, nil
	})

	complexDSL.Action("findCity", func(args []interface{}) (interface{}, error) {
		dataset := args[3].(string)

		data := complexDSL.GetContext(dataset)
		if data == nil {
			return nil, fmt.Errorf("dataset '%s' not found", dataset)
		}

		people, ok := data.([]Person)
		if !ok {
			return nil, fmt.Errorf("invalid data type for %s", dataset)
		}

		var results []interface{}
		for _, person := range people {
			results = append(results, person.City)
		}
		return results, nil
	})

	// Complex data
	people := []Person{
		{"Juan García", 28, "Madrid"},
		{"María López", 35, "Barcelona"},
		{"Carlos Rodríguez", 42, "Madrid"},
	}

	complexTests := []struct {
		query   string
		context map[string]interface{}
		desc    string
	}{
		{"find name in people", map[string]interface{}{"people": people}, "All names"},
		{"find age in people", map[string]interface{}{"people": people}, "All ages"},
		{"find city in people", map[string]interface{}{"people": people}, "All cities"},
	}

	for _, tc := range complexTests {
		result, err := complexDSL.Use(tc.query, tc.context)
		if err != nil {
			fmt.Printf("  %s: Error: %v\n", tc.desc, err)
		} else {
			fmt.Printf("  %s: %v\n", tc.desc, result.GetOutput())
		}
	}

	fmt.Println()

	// Example 4: SetContext vs Use() methods
	fmt.Println("4. SetContext vs Use() methods")
	fmt.Println("------------------------------")

	methodDSL := dslbuilder.New("Methods")
	methodDSL.KeywordToken("SHOW", "show")
	methodDSL.Token("KEY", "[a-zA-Z_][a-zA-Z0-9_]*")
	methodDSL.Rule("command", []string{"SHOW", "KEY"}, "showValue")

	methodDSL.Action("showValue", func(args []interface{}) (interface{}, error) {
		key := args[1].(string)
		value := methodDSL.GetContext(key)
		return fmt.Sprintf("%s = %v", key, value), nil
	})

	// Method 1: Using SetContext
	fmt.Println("  Method 1: Using SetContext()")
	methodDSL.SetContext("user", "Alice")
	methodDSL.SetContext("role", "admin")

	result1, _ := methodDSL.Parse("show user")
	result2, _ := methodDSL.Parse("show role")
	fmt.Printf("    %s\n", result1.GetOutput())
	fmt.Printf("    %s\n", result2.GetOutput())

	// Method 2: Using Use() - like r2lang's q.use()
	fmt.Println("  Method 2: Using Use() - equivalent to r2lang's q.use()")
	context2 := map[string]interface{}{
		"user": "Bob",
		"role": "user",
		"temp": "override",
	}

	result3, _ := methodDSL.Use("show user", context2)
	result4, _ := methodDSL.Use("show role", context2)
	result5, _ := methodDSL.Use("show temp", context2)
	fmt.Printf("    %s\n", result3.GetOutput())
	fmt.Printf("    %s\n", result4.GetOutput())
	fmt.Printf("    %s\n", result5.GetOutput())

	fmt.Println()
	fmt.Println("=== Summary ===")
	fmt.Println("go-dsl equivalent to r2lang:")
	fmt.Println("  r2lang: q.use(\"query\", {key: value})")
	fmt.Println("  go-dsl: dsl.Use(\"query\", map[string]interface{}{\"key\": value})")
	fmt.Println()
	fmt.Println("Alternative method:")
	fmt.Println("  dsl.SetContext(\"key\", value)")
	fmt.Println("  dsl.Parse(\"query\")")
}
