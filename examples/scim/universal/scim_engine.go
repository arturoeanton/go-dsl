package universal

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// DataProvider interface allows injecting custom data access functions
type DataProvider interface {
	// GetUsers returns all users from the data source
	GetUsers() ([]interface{}, error)
	
	// FilterUsers applies a specific filter criteria
	FilterUsers(attribute, operator, value string) ([]interface{}, error)
	
	// ApplyLogicalOperator combines results using logical operators
	ApplyLogicalOperator(operator string, left, right []interface{}) ([]interface{}, error)
}

// SCIMEngine handles the execution of SCIM filter expressions
type SCIMEngine struct {
	dataProvider DataProvider
}

// NewSCIMEngine creates a new SCIM execution engine
func NewSCIMEngine() *SCIMEngine {
	engine := &SCIMEngine{}
	
	// Set default memory-based data provider
	engine.dataProvider = &MemoryDataProvider{}
	
	return engine
}

// SetDataProvider allows injecting a custom data provider
func (e *SCIMEngine) SetDataProvider(provider DataProvider) {
	e.dataProvider = provider
}

// ExecuteFilter executes a parsed SCIM filter expression against the provided data
func (e *SCIMEngine) ExecuteFilter(expr *FilterExpression, context map[string]interface{}) ([]interface{}, error) {
	switch expr.Type {
	case "comparison":
		return e.executeComparison(expr, context)
	case "presence":
		return e.executePresence(expr, context)
	case "logical":
		return e.executeLogical(expr, context)
	case "complex":
		return e.executeComplex(expr, context)
	default:
		return nil, fmt.Errorf("unknown expression type: %s", expr.Type)
	}
}

// executeComparison handles comparison operations (eq, ne, co, sw, ew, gt, ge, lt, le)
func (e *SCIMEngine) executeComparison(expr *FilterExpression, context map[string]interface{}) ([]interface{}, error) {
	// Get data from context (users, groups, etc.)
	data, err := e.getDataFromContext(context)
	if err != nil {
		return nil, err
	}

	var results []interface{}
	
	for _, item := range data {
		match, err := e.evaluateComparison(item, expr.Attribute, expr.Operator, expr.Value)
		if err != nil {
			continue // Skip items that can't be evaluated
		}
		if match {
			results = append(results, item)
		}
	}
	
	return results, nil
}

// executePresence handles presence tests (pr operator)
func (e *SCIMEngine) executePresence(expr *FilterExpression, context map[string]interface{}) ([]interface{}, error) {
	data, err := e.getDataFromContext(context)
	if err != nil {
		return nil, err
	}

	var results []interface{}
	
	for _, item := range data {
		if e.hasAttribute(item, expr.Attribute) {
			results = append(results, item)
		}
	}
	
	return results, nil
}

// executeLogical handles logical operations (and, or, not)
func (e *SCIMEngine) executeLogical(expr *FilterExpression, context map[string]interface{}) ([]interface{}, error) {
	switch expr.Operator {
	case "and":
		leftResults, err := e.ExecuteFilter(expr.Left, context)
		if err != nil {
			return nil, err
		}
		rightResults, err := e.ExecuteFilter(expr.Right, context)
		if err != nil {
			return nil, err
		}
		return e.intersect(leftResults, rightResults), nil
		
	case "or":
		leftResults, err := e.ExecuteFilter(expr.Left, context)
		if err != nil {
			return nil, err
		}
		rightResults, err := e.ExecuteFilter(expr.Right, context)
		if err != nil {
			return nil, err
		}
		return e.union(leftResults, rightResults), nil
		
	case "not":
		allData, err := e.getDataFromContext(context)
		if err != nil {
			return nil, err
		}
		exprResults, err := e.ExecuteFilter(expr.Expression, context)
		if err != nil {
			return nil, err
		}
		return e.difference(allData, exprResults), nil
		
	default:
		return nil, fmt.Errorf("unknown logical operator: %s", expr.Operator)
	}
}

// executeComplex handles complex attribute filters like emails[type eq "work"]
func (e *SCIMEngine) executeComplex(expr *FilterExpression, context map[string]interface{}) ([]interface{}, error) {
	data, err := e.getDataFromContext(context)
	if err != nil {
		return nil, err
	}

	var results []interface{}
	
	for _, item := range data {
		// Get the complex attribute (e.g., emails, phoneNumbers)
		complexAttr := e.getAttributeValue(item, expr.Attribute)
		if complexAttr == nil {
			continue
		}
		
		// Check if it's a slice/array
		rv := reflect.ValueOf(complexAttr)
		if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
			continue
		}
		
		// Apply the sub-filter to each element
		for i := 0; i < rv.Len(); i++ {
			element := rv.Index(i).Interface()
			subContext := map[string]interface{}{"element": []interface{}{element}}
			
			subResults, err := e.ExecuteFilter(expr.Expression, subContext)
			if err != nil {
				continue
			}
			
			// If any element matches, include the parent item
			if len(subResults) > 0 {
				results = append(results, item)
				break
			}
		}
	}
	
	return results, nil
}

// evaluateComparison evaluates a comparison operation on a single item
func (e *SCIMEngine) evaluateComparison(item interface{}, attribute, operator string, value interface{}) (bool, error) {
	itemValue := e.getAttributeValue(item, attribute)
	if itemValue == nil {
		return false, nil
	}
	
	switch operator {
	case "eq":
		return e.equals(itemValue, value), nil
	case "ne":
		return !e.equals(itemValue, value), nil
	case "co":
		return e.contains(itemValue, value), nil
	case "sw":
		return e.startsWith(itemValue, value), nil
	case "ew":
		return e.endsWith(itemValue, value), nil
	case "gt":
		return e.greaterThan(itemValue, value), nil
	case "ge":
		return e.greaterThanOrEqual(itemValue, value), nil
	case "lt":
		return e.lessThan(itemValue, value), nil
	case "le":
		return e.lessThanOrEqual(itemValue, value), nil
	default:
		return false, fmt.Errorf("unknown comparison operator: %s", operator)
	}
}

// getAttributeValue extracts an attribute value from an item using reflection
func (e *SCIMEngine) getAttributeValue(item interface{}, attribute string) interface{} {
	// Handle nested attributes like "meta.created" or "name.familyName"
	parts := strings.Split(attribute, ".")
	
	current := item
	for _, part := range parts {
		rv := reflect.ValueOf(current)
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}
		
		if rv.Kind() != reflect.Struct {
			return nil
		}
		
		// Try to find field by name (case-insensitive)
		field := e.findField(rv, part)
		if !field.IsValid() {
			return nil
		}
		
		current = field.Interface()
	}
	
	return current
}

// findField finds a struct field by name (case-insensitive)
func (e *SCIMEngine) findField(rv reflect.Value, fieldName string) reflect.Value {
	rt := rv.Type()
	
	// First try exact match
	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		if field.Name == fieldName {
			return rv.Field(i)
		}
		
		// Check SCIM tag
		if scimTag := field.Tag.Get("scim"); scimTag == fieldName {
			return rv.Field(i)
		}
		
		// Check JSON tag
		if jsonTag := field.Tag.Get("json"); strings.Split(jsonTag, ",")[0] == fieldName {
			return rv.Field(i)
		}
	}
	
	// Try case-insensitive match
	lowerFieldName := strings.ToLower(fieldName)
	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		if strings.ToLower(field.Name) == lowerFieldName {
			return rv.Field(i)
		}
	}
	
	return reflect.Value{}
}

// hasAttribute checks if an item has a specific attribute
func (e *SCIMEngine) hasAttribute(item interface{}, attribute string) bool {
	value := e.getAttributeValue(item, attribute)
	return value != nil && !reflect.ValueOf(value).IsZero()
}

// Comparison helper functions

func (e *SCIMEngine) equals(a, b interface{}) bool {
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

func (e *SCIMEngine) contains(a, b interface{}) bool {
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return strings.Contains(strings.ToLower(aStr), strings.ToLower(bStr))
}

func (e *SCIMEngine) startsWith(a, b interface{}) bool {
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return strings.HasPrefix(strings.ToLower(aStr), strings.ToLower(bStr))
}

func (e *SCIMEngine) endsWith(a, b interface{}) bool {
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return strings.HasSuffix(strings.ToLower(aStr), strings.ToLower(bStr))
}

func (e *SCIMEngine) greaterThan(a, b interface{}) bool {
	return e.compareValues(a, b) > 0
}

func (e *SCIMEngine) greaterThanOrEqual(a, b interface{}) bool {
	return e.compareValues(a, b) >= 0
}

func (e *SCIMEngine) lessThan(a, b interface{}) bool {
	return e.compareValues(a, b) < 0
}

func (e *SCIMEngine) lessThanOrEqual(a, b interface{}) bool {
	return e.compareValues(a, b) <= 0
}

// compareValues compares two values numerically or lexicographically
func (e *SCIMEngine) compareValues(a, b interface{}) int {
	// Try numeric comparison first
	aFloat, aErr := e.toFloat64(a)
	bFloat, bErr := e.toFloat64(b)
	
	if aErr == nil && bErr == nil {
		if aFloat < bFloat {
			return -1
		} else if aFloat > bFloat {
			return 1
		}
		return 0
	}
	
	// Try datetime comparison
	aTime, aErr := e.parseDateTime(a)
	bTime, bErr := e.parseDateTime(b)
	
	if aErr == nil && bErr == nil {
		if aTime.Before(bTime) {
			return -1
		} else if aTime.After(bTime) {
			return 1
		}
		return 0
	}
	
	// Fall back to string comparison
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return strings.Compare(aStr, bStr)
}

// toFloat64 converts a value to float64
func (e *SCIMEngine) toFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}

// parseDateTime parses various datetime formats
func (e *SCIMEngine) parseDateTime(value interface{}) (time.Time, error) {
	str := fmt.Sprintf("%v", value)
	
	// Try RFC3339 format (ISO 8601)
	if t, err := time.Parse(time.RFC3339, str); err == nil {
		return t, nil
	}
	
	// Try RFC3339 without timezone
	if t, err := time.Parse("2006-01-02T15:04:05", str); err == nil {
		return t, nil
	}
	
	// Try date only
	if t, err := time.Parse("2006-01-02", str); err == nil {
		return t, nil
	}
	
	return time.Time{}, fmt.Errorf("invalid datetime format: %s", str)
}

// itemsEqual compares two items for equality using reflection
func (e *SCIMEngine) itemsEqual(a, b interface{}) bool {
	// Use reflect.DeepEqual for safe comparison of all types
	return reflect.DeepEqual(a, b)
}

// Set operations for logical operators

func (e *SCIMEngine) intersect(a, b []interface{}) []interface{} {
	result := make([]interface{}, 0)
	
	// Use linear search instead of map to avoid hash issues with complex types
	for _, itemA := range a {
		for _, itemB := range b {
			if e.itemsEqual(itemA, itemB) {
				result = append(result, itemA)
				break
			}
		}
	}
	
	return result
}

func (e *SCIMEngine) union(a, b []interface{}) []interface{} {
	result := make([]interface{}, 0)
	
	// Add all items from a
	for _, item := range a {
		result = append(result, item)
	}
	
	// Add items from b that are not already in result
	for _, itemB := range b {
		found := false
		for _, itemA := range a {
			if e.itemsEqual(itemA, itemB) {
				found = true
				break
			}
		}
		if !found {
			result = append(result, itemB)
		}
	}
	
	return result
}

func (e *SCIMEngine) difference(a, b []interface{}) []interface{} {
	result := make([]interface{}, 0)
	
	// Add items from a that are not in b
	for _, itemA := range a {
		found := false
		for _, itemB := range b {
			if e.itemsEqual(itemA, itemB) {
				found = true
				break
			}
		}
		if !found {
			result = append(result, itemA)
		}
	}
	
	return result
}

// getDataFromContext extracts data from the context (users, groups, etc.)
func (e *SCIMEngine) getDataFromContext(context map[string]interface{}) ([]interface{}, error) {
	// Try different common keys
	keys := []string{"users", "data", "items", "element", "resources"}
	
	for _, key := range keys {
		if data, exists := context[key]; exists {
			if slice, ok := data.([]interface{}); ok {
				return slice, nil
			}
			// Try to convert single item to slice
			return []interface{}{data}, nil
		}
	}
	
	return nil, fmt.Errorf("no data found in context (looking for keys: %v)", keys)
}

// GetFieldNames extracts field names from a struct using reflection
func (e *SCIMEngine) GetFieldNames(item interface{}) []string {
	var fields []string
	
	rv := reflect.ValueOf(item)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	
	if rv.Kind() != reflect.Struct {
		return fields
	}
	
	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		
		// Check for SCIM tag first
		if scimTag := field.Tag.Get("scim"); scimTag != "" {
			fields = append(fields, scimTag)
			continue
		}
		
		// Check for JSON tag
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			tagName := strings.Split(jsonTag, ",")[0]
			if tagName != "-" {
				fields = append(fields, tagName)
				continue
			}
		}
		
		// Use field name as fallback
		fields = append(fields, strings.ToLower(field.Name))
	}
	
	return fields
}

// FormatItem returns a string representation of an item
func (e *SCIMEngine) FormatItem(item interface{}) string {
	rv := reflect.ValueOf(item)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	
	if rv.Kind() != reflect.Struct {
		return fmt.Sprintf("%v", item)
	}
	
	// Try to find identifying fields
	idFields := []string{"id", "userName", "name", "displayName"}
	rt := rv.Type()
	
	for _, idField := range idFields {
		for i := 0; i < rv.NumField(); i++ {
			field := rt.Field(i)
			fieldName := strings.ToLower(field.Name)
			
			if fieldName == idField {
				value := rv.Field(i).Interface()
				return fmt.Sprintf("%s: %v", field.Name, value)
			}
		}
	}
	
	// Fallback to struct representation
	return fmt.Sprintf("%+v", item)
}

// MemoryDataProvider is a default implementation that works with in-memory data
type MemoryDataProvider struct{}

func (m *MemoryDataProvider) GetUsers() ([]interface{}, error) {
	return []interface{}{}, nil
}

func (m *MemoryDataProvider) FilterUsers(attribute, operator, value string) ([]interface{}, error) {
	return []interface{}{}, nil
}

func (m *MemoryDataProvider) ApplyLogicalOperator(operator string, left, right []interface{}) ([]interface{}, error) {
	return []interface{}{}, nil
}