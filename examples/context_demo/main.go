package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

// Person represents a person in our dataset
type Person struct {
	Name string
	Age  int
	City string
}

func main() {
	// Sample data - this would be your "data" array
	people := []Person{
		{"Juan García", 28, "Madrid"},
		{"María López", 35, "Barcelona"},
		{"Carlos Rodríguez", 42, "Madrid"},
		{"Ana Martínez", 29, "Valencia"},
		{"Pedro Sánchez", 31, "Barcelona"},
		{"Laura Fernández", 26, "Madrid"},
	}

	// Create a simple query DSL
	query := dslbuilder.New("ContextQueryDSL")

	// Define tokens - keywords first with high priority
	query.KeywordToken("SELECT", "select")
	query.KeywordToken("FROM", "from")
	query.KeywordToken("WHERE", "where")
	query.KeywordToken("AND", "and")
	query.KeywordToken("NAME", "name")
	query.KeywordToken("AGE", "age")
	query.KeywordToken("CITY", "city")
	
	// Operators
	query.Token("GT", ">")
	query.Token("LT", "<")
	query.Token("EQ", "==")
	query.Token("NEQ", "!=")
	
	// Values
	query.Token("NUMBER", "[0-9]+")
	query.Token("IDENTIFIER", "[a-zA-Z][a-zA-Z0-9]*")

	// Define grammar rules - more specific rules first
	query.Rule("query", []string{"SELECT", "field", "FROM", "IDENTIFIER", "WHERE", "condition"}, "selectWithWhere")
	query.Rule("query", []string{"SELECT", "field", "FROM", "IDENTIFIER"}, "simpleSelect")
	
	// Field rules
	query.Rule("field", []string{"NAME"}, "fieldName")
	query.Rule("field", []string{"AGE"}, "fieldAge")
	query.Rule("field", []string{"CITY"}, "fieldCity")
	
	// Condition rules
	query.Rule("condition", []string{"field", "GT", "NUMBER"}, "numberConditionGT")
	query.Rule("condition", []string{"field", "LT", "NUMBER"}, "numberConditionLT")
	query.Rule("condition", []string{"field", "EQ", "NUMBER"}, "numberConditionEQ")
	query.Rule("condition", []string{"field", "NEQ", "NUMBER"}, "numberConditionNEQ")
	query.Rule("condition", []string{"field", "EQ", "IDENTIFIER"}, "stringConditionEQ")
	query.Rule("condition", []string{"field", "NEQ", "IDENTIFIER"}, "stringConditionNEQ")

	// Define actions for fields
	query.Action("fieldName", func(args []interface{}) (interface{}, error) {
		return "name", nil
	})
	query.Action("fieldAge", func(args []interface{}) (interface{}, error) {
		return "age", nil
	})
	query.Action("fieldCity", func(args []interface{}) (interface{}, error) {
		return "city", nil
	})

	// Define actions that use context
	query.Action("simpleSelect", func(args []interface{}) (interface{}, error) {
		field := args[1].(string)
		tableName := args[3].(string)

		// Get data from context - equivalent to context1.data in r2lang
		data := query.GetContext(tableName)
		if data == nil {
			return nil, fmt.Errorf("table '%s' not found in context", tableName)
		}

		people, ok := data.([]Person)
		if !ok {
			return nil, fmt.Errorf("invalid data type for table '%s'", tableName)
		}

		// Extract the requested field from all records
		var results []string
		for _, person := range people {
			switch field {
			case "name":
				results = append(results, person.Name)
			case "age":
				results = append(results, strconv.Itoa(person.Age))
			case "city":
				results = append(results, person.City)
			}
		}

		return results, nil
	})

	query.Action("selectWithWhere", func(args []interface{}) (interface{}, error) {
		field := args[1].(string)
		tableName := args[3].(string)
		condition := args[5]

		// Get data from context
		data := query.GetContext(tableName)
		if data == nil {
			return nil, fmt.Errorf("table '%s' not found in context", tableName)
		}

		people, ok := data.([]Person)
		if !ok {
			return nil, fmt.Errorf("invalid data type for table '%s'", tableName)
		}

		// Apply condition filter
		var filtered []Person
		conditionMap := condition.(map[string]interface{})
		condField := conditionMap["field"].(string)
		operator := conditionMap["operator"].(string)
		value := conditionMap["value"]

		for _, person := range people {
			match := false
			switch condField {
			case "age":
				age := person.Age
				compareValue, _ := strconv.Atoi(value.(string))
				switch operator {
				case ">":
					match = age > compareValue
				case "<":
					match = age < compareValue
				case "==":
					match = age == compareValue
				case "!=":
					match = age != compareValue
				}
			case "city":
				switch operator {
				case "==":
					match = person.City == value.(string)
				case "!=":
					match = person.City != value.(string)
				}
			case "name":
				switch operator {
				case "==":
					match = person.Name == value.(string)
				case "!=":
					match = person.Name != value.(string)
				}
			}

			if match {
				filtered = append(filtered, person)
			}
		}

		// Extract requested field from filtered results
		var results []string
		for _, person := range filtered {
			switch field {
			case "name":
				results = append(results, person.Name)
			case "age":
				results = append(results, strconv.Itoa(person.Age))
			case "city":
				results = append(results, person.City)
			}
		}

		return results, nil
	})

	// Number condition actions
	query.Action("numberConditionGT", func(args []interface{}) (interface{}, error) {
		return map[string]interface{}{
			"field":    args[0].(string),
			"operator": ">",
			"value":    args[2].(string),
		}, nil
	})
	
	query.Action("numberConditionLT", func(args []interface{}) (interface{}, error) {
		return map[string]interface{}{
			"field":    args[0].(string),
			"operator": "<",
			"value":    args[2].(string),
		}, nil
	})
	
	query.Action("numberConditionEQ", func(args []interface{}) (interface{}, error) {
		return map[string]interface{}{
			"field":    args[0].(string),
			"operator": "==",
			"value":    args[2].(string),
		}, nil
	})
	
	query.Action("numberConditionNEQ", func(args []interface{}) (interface{}, error) {
		return map[string]interface{}{
			"field":    args[0].(string),
			"operator": "!=",
			"value":    args[2].(string),
		}, nil
	})

	// String condition actions
	query.Action("stringConditionEQ", func(args []interface{}) (interface{}, error) {
		return map[string]interface{}{
			"field":    args[0].(string),
			"operator": "==",
			"value":    args[2].(string),
		}, nil
	})
	
	query.Action("stringConditionNEQ", func(args []interface{}) (interface{}, error) {
		return map[string]interface{}{
			"field":    args[0].(string),
			"operator": "!=",
			"value":    args[2].(string),
		}, nil
	})

	// Demo 1: Using Use() method with context (equivalent to r2lang's q.use())
	fmt.Println("=== Demo 1: Using Use() method with context ===")
	context1 := map[string]interface{}{
		"data": people, // This is like {data: data} in r2lang
	}

	// This is equivalent to: q.use("select name from data where age > 30", context1)
	result1, err := query.Use("select name from data where age > 30", context1)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Query: select name from data where age > 30\n")
		fmt.Printf("Results: %v\n", result1.GetOutput())
	}

	fmt.Println()

	// Demo 2: Using SetContext() method
	fmt.Println("=== Demo 2: Using SetContext() method ===")
	query.SetContext("users", people)
	query.SetContext("minAge", 25)

	result2, err := query.Parse("select city from users where age > 25")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Query: select city from users where age > 25\n")
		fmt.Printf("Results: %v\n", result2.GetOutput())
	}

	fmt.Println()

	// Demo 3: Multiple contexts
	fmt.Println("=== Demo 3: Multiple contexts and queries ===")
	
	youngPeople := []Person{
		{"Sofia Morales", 22, "Barcelona"},
		{"Diego Ruiz", 24, "Sevilla"},
	}

	queries := []struct {
		query   string
		context map[string]interface{}
		desc    string
	}{
		{
			"select name from people",
			map[string]interface{}{"people": people},
			"All people names",
		},
		{
			"select name from people where age > 30",
			map[string]interface{}{"people": people},
			"People older than 30",
		},
		{
			"select city from young where age > 20",
			map[string]interface{}{"young": youngPeople},
			"Cities of young people older than 20",
		},
		{
			"select name from people where city == Madrid",
			map[string]interface{}{"people": people},
			"People from Madrid",
		},
	}

	for i, q := range queries {
		fmt.Printf("%d. %s\n", i+1, q.desc)
		fmt.Printf("   Query: %s\n", q.query)
		
		result, err := query.Use(q.query, q.context)
		if err != nil {
			fmt.Printf("   Error: %v\n", err)
		} else {
			fmt.Printf("   Results: %v\n", result.GetOutput())
		}
		fmt.Println()
	}

	// Demo 4: Dynamic context manipulation
	fmt.Println("=== Demo 4: Dynamic context manipulation ===")
	
	// Start with empty context
	contextQuery := dslbuilder.New("DynamicContext")
	contextQuery.KeywordToken("GET", "get")
	contextQuery.Token("KEY", "[a-zA-Z]+")
	contextQuery.Rule("command", []string{"GET", "KEY"}, "getValue")
	
	contextQuery.Action("getValue", func(args []interface{}) (interface{}, error) {
		key := args[1].(string)
		value := contextQuery.GetContext(key)
		if value == nil {
			return fmt.Sprintf("Key '%s' not found", key), nil
		}
		return value, nil
	})

	// Set some context values
	contextQuery.SetContext("user", "Juan")
	contextQuery.SetContext("age", 30)
	contextQuery.SetContext("city", "Madrid")

	// Query the context
	testQueries := []string{"get user", "get age", "get city", "get nonexistent"}
	for _, tq := range testQueries {
		result, err := contextQuery.Parse(tq)
		if err != nil {
			fmt.Printf("Query '%s': Error %v\n", tq, err)
		} else {
			fmt.Printf("Query '%s': %v\n", tq, result.GetOutput())
		}
	}
}