package dsllinq

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

var Linq *dslbuilder.DSL
var QueryEngine *GenericQueryEngine

// GenericQueryEngine handles all reflection-based operations
type GenericQueryEngine struct {
	currentData []interface{}
}

func NewGenericQueryEngine() *GenericQueryEngine {
	return &GenericQueryEngine{}
}

func (g *GenericQueryEngine) SetData(data []interface{}) {
	g.currentData = data
}

func (g *GenericQueryEngine) GetFieldValue(item interface{}, fieldName string) (interface{}, error) {
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()
	// First try to find by linq tag
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if tag := field.Tag.Get("linq"); tag != "" {
			if strings.EqualFold(tag, fieldName) {
				fieldValue := v.Field(i)
				if fieldValue.CanInterface() {
					return fieldValue.Interface(), nil
				}
			}
		}
	}

	// Fallback to field name matching (case-insensitive)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if strings.EqualFold(field.Name, fieldName) {
			fieldValue := v.Field(i)
			if fieldValue.CanInterface() {
				return fieldValue.Interface(), nil
			}
		}
	}

	return nil, fmt.Errorf("field '%s' not found", fieldName)
}

func (g *GenericQueryEngine) GetFieldNames(item interface{}) []string {
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()
	var fields []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.IsExported() {
			// Prefer linq tag name, fallback to field name
			if tag := field.Tag.Get("linq"); tag != "" {
				fields = append(fields, tag)
			} else {
				fields = append(fields, strings.ToLower(field.Name))
			}
		}
	}
	return fields
}

func (g *GenericQueryEngine) Filter(fieldName, operator string, compareValue interface{}) []interface{} {
	var filtered []interface{}
	for _, item := range g.currentData {
		fieldValue, err := g.GetFieldValue(item, fieldName)
		if err != nil {
			continue // Skip items that don't have this field
		}

		if g.compareValues(fieldValue, operator, compareValue) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func (g *GenericQueryEngine) compareValues(fieldValue interface{}, operator string, compareValue interface{}) bool {
	switch fv := fieldValue.(type) {
	case int:
		if cv, ok := compareValue.(int); ok {
			return g.compareNumbers(float64(fv), operator, float64(cv))
		}
		if cv, ok := compareValue.(float64); ok {
			return g.compareNumbers(float64(fv), operator, cv)
		}
	case float64:
		if cv, ok := compareValue.(int); ok {
			return g.compareNumbers(fv, operator, float64(cv))
		}
		if cv, ok := compareValue.(float64); ok {
			return g.compareNumbers(fv, operator, cv)
		}
	case string:
		if cv, ok := compareValue.(string); ok {
			return g.compareStrings(fv, operator, cv)
		}
	}
	return false
}

func (g *GenericQueryEngine) compareNumbers(field float64, op string, value float64) bool {
	switch op {
	case ">":
		return field > value
	case "<":
		return field < value
	case "==":
		return field == value
	case "!=":
		return field != value
	case ">=":
		return field >= value
	case "<=":
		return field <= value
	}
	return false
}

func (g *GenericQueryEngine) compareStrings(field string, op string, value string) bool {
	switch op {
	case "==":
		return field == value
	case "!=":
		return field != value
	case "contains":
		return strings.Contains(strings.ToLower(field), strings.ToLower(value))
	}
	return false
}

func (g *GenericQueryEngine) Sort(items []interface{}, fieldName string, descending bool) []interface{} {
	if len(items) <= 1 {
		return items
	}

	// Create a copy to avoid modifying original
	sorted := make([]interface{}, len(items))
	copy(sorted, items)

	// Simple bubble sort using reflection
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			val1, err1 := g.GetFieldValue(sorted[j], fieldName)
			val2, err2 := g.GetFieldValue(sorted[j+1], fieldName)

			if err1 != nil || err2 != nil {
				continue
			}

			shouldSwap := false

			// Compare values based on type using reflection
			switch v1 := val1.(type) {
			case int:
				if v2, ok := val2.(int); ok {
					if descending {
						shouldSwap = v1 < v2
					} else {
						shouldSwap = v1 > v2
					}
				}
			case float64:
				if v2, ok := val2.(float64); ok {
					if descending {
						shouldSwap = v1 < v2
					} else {
						shouldSwap = v1 > v2
					}
				}
			case string:
				if v2, ok := val2.(string); ok {
					if descending {
						shouldSwap = v1 > v2
					} else {
						shouldSwap = v1 < v2
					}
				}
			}

			if shouldSwap {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	return sorted
}

func (g *GenericQueryEngine) FormatResult(items []interface{}, selectField string) string {
	var result strings.Builder

	if len(items) == 0 {
		result.WriteString("No results found\n")
		return result.String()
	}

	if selectField == "*" {
		// Show all fields using reflection
		for i, item := range items {
			result.WriteString(fmt.Sprintf("  [%d] ", i+1))

			v := reflect.ValueOf(item)
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
			}

			t := v.Type()
			var parts []string
			for j := 0; j < t.NumField(); j++ {
				field := t.Field(j)
				if field.IsExported() {
					fieldValue := v.Field(j)
					if fieldValue.CanInterface() {
						parts = append(parts, fmt.Sprintf("%s: %v", field.Name, fieldValue.Interface()))
					}
				}
			}
			result.WriteString(strings.Join(parts, ", "))
			result.WriteString("\n")
		}
	} else {
		// Show specific field using reflection
		for _, item := range items {
			fieldValue, err := g.GetFieldValue(item, selectField)
			if err == nil {
				result.WriteString(fmt.Sprintf("  - %v\n", fieldValue))
			}
		}
	}

	result.WriteString(fmt.Sprintf("\nTotal results: %d\n", len(items)))
	return result.String()
}

func init() {
	// Create Universal LINQ DSL
	linq := dslbuilder.New("UniversalLINQ")

	// Initialize generic query engine
	queryEngine := NewGenericQueryEngine()
	QueryEngine = queryEngine
	linq.Set("engine", queryEngine)

	// Define tokens - completely generic approach
	linq.KeywordToken("FROM", "from")
	linq.KeywordToken("WHERE", "where")
	linq.KeywordToken("SELECT", "select")
	linq.KeywordToken("ORDERBY", "orderby")
	linq.KeywordToken("TOP", "top")
	linq.KeywordToken("AND", "and")
	linq.KeywordToken("OR", "or")
	linq.KeywordToken("ASC", "asc")
	linq.KeywordToken("DESC", "desc")

	// Operators
	linq.Token("GT", ">")
	linq.Token("LT", "<")
	linq.Token("EQ", "==")
	linq.Token("NEQ", "!=")
	linq.Token("GTE", ">=")
	linq.Token("LTE", "<=")
	linq.KeywordToken("CONTAINS", "contains")

	// Generic tokens
	linq.Token("STAR", "\\*")
	linq.Token("NUMBER", "[0-9]+\\.?[0-9]*")
	linq.Token("STRING", "\"[^\"]*\"")
	linq.Token("IDENTIFIER", "[a-zA-Z_][a-zA-Z0-9_]*")

	// Simplified grammar rules - MUCH more generic
	// Basic patterns that cover most use cases
	linq.Rule("query", []string{"FROM", "IDENTIFIER", "SELECT", "STAR"}, "selectAll")
	linq.Rule("query", []string{"FROM", "IDENTIFIER", "SELECT", "IDENTIFIER"}, "selectField")
	linq.Rule("query", []string{"FROM", "IDENTIFIER", "WHERE", "IDENTIFIER", "GT", "NUMBER", "SELECT", "IDENTIFIER"}, "whereGreater")
	linq.Rule("query", []string{"FROM", "IDENTIFIER", "WHERE", "IDENTIFIER", "LT", "NUMBER", "SELECT", "IDENTIFIER"}, "whereLess")
	linq.Rule("query", []string{"FROM", "IDENTIFIER", "WHERE", "IDENTIFIER", "EQ", "STRING", "SELECT", "IDENTIFIER"}, "whereEqualString")
	linq.Rule("query", []string{"FROM", "IDENTIFIER", "WHERE", "IDENTIFIER", "EQ", "NUMBER", "SELECT", "IDENTIFIER"}, "whereEqualNumber")
	linq.Rule("query", []string{"FROM", "IDENTIFIER", "TOP", "NUMBER", "SELECT", "IDENTIFIER"}, "selectTop")
	linq.Rule("query", []string{"FROM", "IDENTIFIER", "SELECT", "IDENTIFIER", "ORDERBY", "IDENTIFIER", "DESC"}, "selectOrderDesc")
	linq.Rule("query", []string{"FROM", "IDENTIFIER", "SELECT", "IDENTIFIER", "ORDERBY", "IDENTIFIER", "ASC"}, "selectOrderAsc")
	linq.Rule("query", []string{"FROM", "IDENTIFIER", "SELECT", "IDENTIFIER", "ORDERBY", "IDENTIFIER"}, "selectOrderDefault")
	linq.Rule("query", []string{"FROM", "IDENTIFIER", "WHERE", "IDENTIFIER", "GT", "NUMBER", "SELECT", "IDENTIFIER", "ORDERBY", "IDENTIFIER", "DESC"}, "whereOrderDesc")

	// Actions using the generic query engine
	linq.Action("selectAll", func(args []interface{}) (interface{}, error) {
		tableName := args[1].(string)

		data := linq.GetContext(tableName)
		if data == nil {
			return nil, fmt.Errorf("table '%s' not found", tableName)
		}

		items := data.([]interface{})
		queryEngine.SetData(items)

		return queryEngine.FormatResult(items, "*"), nil
	})

	linq.Action("selectField", func(args []interface{}) (interface{}, error) {
		tableName := args[1].(string)
		fieldName := args[3].(string)

		data := linq.GetContext(tableName)
		if data == nil {
			return nil, fmt.Errorf("table '%s' not found", tableName)
		}

		items := data.([]interface{})
		queryEngine.SetData(items)

		return queryEngine.FormatResult(items, fieldName), nil
	})

	linq.Action("whereGreater", func(args []interface{}) (interface{}, error) {
		tableName := args[1].(string)
		fieldName := args[3].(string)
		numStr := args[5].(string)
		selectField := args[7].(string)

		threshold, _ := strconv.ParseFloat(numStr, 64)

		data := linq.GetContext(tableName)
		items := data.([]interface{})
		queryEngine.SetData(items)

		filtered := queryEngine.Filter(fieldName, ">", threshold)
		return queryEngine.FormatResult(filtered, selectField), nil
	})

	linq.Action("whereLess", func(args []interface{}) (interface{}, error) {
		tableName := args[1].(string)
		fieldName := args[3].(string)
		numStr := args[5].(string)
		selectField := args[7].(string)

		threshold, _ := strconv.ParseFloat(numStr, 64)

		data := linq.GetContext(tableName)
		items := data.([]interface{})
		queryEngine.SetData(items)

		filtered := queryEngine.Filter(fieldName, "<", threshold)
		return queryEngine.FormatResult(filtered, selectField), nil
	})

	linq.Action("whereEqualString", func(args []interface{}) (interface{}, error) {
		tableName := args[1].(string)
		fieldName := args[3].(string)
		valueStr := strings.Trim(args[5].(string), "\"")
		selectField := args[7].(string)

		data := linq.GetContext(tableName)
		items := data.([]interface{})
		queryEngine.SetData(items)

		filtered := queryEngine.Filter(fieldName, "==", valueStr)
		return queryEngine.FormatResult(filtered, selectField), nil
	})

	linq.Action("whereEqualNumber", func(args []interface{}) (interface{}, error) {
		tableName := args[1].(string)
		fieldName := args[3].(string)
		numStr := args[5].(string)
		selectField := args[7].(string)

		value, _ := strconv.ParseFloat(numStr, 64)

		data := linq.GetContext(tableName)
		items := data.([]interface{})
		queryEngine.SetData(items)

		filtered := queryEngine.Filter(fieldName, "==", value)
		return queryEngine.FormatResult(filtered, selectField), nil
	})

	linq.Action("selectTop", func(args []interface{}) (interface{}, error) {
		tableName := args[1].(string)
		topStr := args[3].(string)
		fieldName := args[5].(string)

		top, _ := strconv.Atoi(topStr)

		data := linq.GetContext(tableName)
		items := data.([]interface{})

		limited := items
		if top < len(limited) {
			limited = limited[:top]
		}

		queryEngine.SetData(limited)
		return queryEngine.FormatResult(limited, fieldName), nil
	})

	linq.Action("selectOrderDesc", func(args []interface{}) (interface{}, error) {
		tableName := args[1].(string)
		selectField := args[3].(string)
		orderField := args[5].(string)

		data := linq.GetContext(tableName)
		items := data.([]interface{})
		queryEngine.SetData(items)

		sorted := queryEngine.Sort(items, orderField, true) // true = descending
		return queryEngine.FormatResult(sorted, selectField), nil
	})

	linq.Action("selectOrderAsc", func(args []interface{}) (interface{}, error) {
		tableName := args[1].(string)
		selectField := args[3].(string)
		orderField := args[5].(string)

		data := linq.GetContext(tableName)
		items := data.([]interface{})
		queryEngine.SetData(items)

		sorted := queryEngine.Sort(items, orderField, false) // false = ascending
		return queryEngine.FormatResult(sorted, selectField), nil
	})

	linq.Action("selectOrderDefault", func(args []interface{}) (interface{}, error) {
		tableName := args[1].(string)
		selectField := args[3].(string)
		orderField := args[5].(string)

		data := linq.GetContext(tableName)
		items := data.([]interface{})
		queryEngine.SetData(items)

		sorted := queryEngine.Sort(items, orderField, false) // default = ascending
		return queryEngine.FormatResult(sorted, selectField), nil
	})

	linq.Action("whereOrderDesc", func(args []interface{}) (interface{}, error) {
		tableName := args[1].(string)
		whereField := args[3].(string)
		numStr := args[5].(string)
		selectField := args[7].(string)
		orderField := args[9].(string)

		threshold, _ := strconv.ParseFloat(numStr, 64)

		data := linq.GetContext(tableName)
		items := data.([]interface{})
		queryEngine.SetData(items)

		filtered := queryEngine.Filter(whereField, ">", threshold)
		sorted := queryEngine.Sort(filtered, orderField, true) // true = descending
		return queryEngine.FormatResult(sorted, selectField), nil
	})
	Linq = linq
}
