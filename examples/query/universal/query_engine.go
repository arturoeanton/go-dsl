package universal

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// UniversalQueryEngine handles generic queries on any struct type using reflection
type UniversalQueryEngine struct{}

// NewQueryEngine creates a new universal query engine
func NewQueryEngine() *UniversalQueryEngine {
	return &UniversalQueryEngine{}
}

// GetFieldNames extracts field names from any struct, supporting query tags
func (uqe *UniversalQueryEngine) GetFieldNames(item interface{}) []string {
	v := reflect.ValueOf(item)
	t := reflect.TypeOf(item)
	
	// Handle pointer types
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}
	
	var fields []string
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		
		// Priority 1: struct tag "query"
		if tag := field.Tag.Get("query"); tag != "" {
			fields = append(fields, tag)
		} else {
			// Priority 2: lowercase field name (compatible)
			fields = append(fields, strings.ToLower(field.Name))
		}
	}
	
	return fields
}

// GetFieldValue gets the value of a specific field from a struct using reflection
func (uqe *UniversalQueryEngine) GetFieldValue(item interface{}, fieldName string) interface{} {
	v := reflect.ValueOf(item)
	t := reflect.TypeOf(item)
	
	// Handle pointer types
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}
	
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		
		// Search by tag first
		if tag := field.Tag.Get("query"); tag == fieldName {
			return v.Field(i).Interface()
		}
		
		// Search by lowercase field name
		if strings.ToLower(field.Name) == fieldName {
			return v.Field(i).Interface()
		}
	}
	
	return nil
}

// ApplyFilter applies a filter condition to a slice of any struct type
func (uqe *UniversalQueryEngine) ApplyFilter(data []interface{}, field string, operator string, value interface{}) []interface{} {
	var result []interface{}
	
	for _, item := range data {
		fieldValue := uqe.GetFieldValue(item, field)
		if fieldValue == nil {
			continue
		}
		
		if uqe.compareValues(fieldValue, operator, value) {
			result = append(result, item)
		}
	}
	
	return result
}

// compareValues compares two values using the given operator
func (uqe *UniversalQueryEngine) compareValues(fieldValue interface{}, operator string, compareValue interface{}) bool {
	switch operator {
	case "es", "is", "==":
		return uqe.equalCompare(fieldValue, compareValue)
	case "mayor", "greater", ">":
		return uqe.numericCompare(fieldValue, compareValue) > 0
	case "menor", "less", "<":
		return uqe.numericCompare(fieldValue, compareValue) < 0
	case "mayor_igual", "greater_equal", ">=":
		return uqe.numericCompare(fieldValue, compareValue) >= 0
	case "menor_igual", "less_equal", "<=":
		return uqe.numericCompare(fieldValue, compareValue) <= 0
	case "contiene", "contains":
		return uqe.containsCompare(fieldValue, compareValue)
	}
	return false
}

// equalCompare compares two values for equality
func (uqe *UniversalQueryEngine) equalCompare(a, b interface{}) bool {
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return strings.EqualFold(aStr, bStr)
}

// numericCompare compares two values numerically
func (uqe *UniversalQueryEngine) numericCompare(a, b interface{}) int {
	aFloat := uqe.toFloat64(a)
	bFloat := uqe.toFloat64(b)
	
	if aFloat > bFloat {
		return 1
	} else if aFloat < bFloat {
		return -1
	}
	return 0
}

// containsCompare checks if a contains b (case insensitive)
func (uqe *UniversalQueryEngine) containsCompare(a, b interface{}) bool {
	aStr := strings.ToLower(fmt.Sprintf("%v", a))
	bStr := strings.ToLower(fmt.Sprintf("%v", b))
	return strings.Contains(aStr, bStr)
}

// toFloat64 converts any numeric type to float64
func (uqe *UniversalQueryEngine) toFloat64(value interface{}) float64 {
	switch v := value.(type) {
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case float32:
		return float64(v)
	case float64:
		return v
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return 0.0
}

// Count returns the count of items in a slice
func (uqe *UniversalQueryEngine) Count(data []interface{}) int {
	return len(data)
}

// FormatResults formats results for display
func (uqe *UniversalQueryEngine) FormatResults(data []interface{}) string {
	if len(data) == 0 {
		return "No results found\n"
	}
	
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Results (%d items):\n", len(data)))
	
	for _, item := range data {
		result.WriteString(fmt.Sprintf("  %s\n", uqe.FormatItem(item)))
	}
	
	return result.String()
}

// FormatItem formats a single item for display (exported method)
func (uqe *UniversalQueryEngine) FormatItem(item interface{}) string {
	v := reflect.ValueOf(item)
	t := reflect.TypeOf(item)
	
	// Handle pointer types
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}
	
	var parts []string
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i).Interface()
		
		// Use tag name if available, otherwise field name
		name := field.Name
		if tag := field.Tag.Get("query"); tag != "" {
			name = tag
		}
		
		parts = append(parts, fmt.Sprintf("%s: %v", name, value))
	}
	
	return strings.Join(parts, ", ")
}

// SelectFields selects specific fields from results
func (uqe *UniversalQueryEngine) SelectFields(data []interface{}, fields []string) []map[string]interface{} {
	var result []map[string]interface{}
	
	for _, item := range data {
		selected := make(map[string]interface{})
		
		for _, field := range fields {
			if value := uqe.GetFieldValue(item, field); value != nil {
				selected[field] = value
			}
		}
		
		if len(selected) > 0 {
			result = append(result, selected)
		}
	}
	
	return result
}

// ConvertToInterfaceSlice converts any slice type to []interface{}
func (uqe *UniversalQueryEngine) ConvertToInterfaceSlice(slice interface{}) []interface{} {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		return nil
	}
	
	result := make([]interface{}, v.Len())
	for i := 0; i < v.Len(); i++ {
		result[i] = v.Index(i).Interface()
	}
	
	return result
}