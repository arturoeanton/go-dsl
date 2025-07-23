package universal

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// UniversalLinqEngine handles LINQ operations on any data type using reflection
type UniversalLinqEngine struct {
	data []interface{}
}

// LinqResult represents the result of a LINQ operation
type LinqResult struct {
	data   []interface{}
	engine *UniversalLinqEngine
}

// AggregateResult represents aggregation results
type AggregateResult struct {
	Count   int
	Sum     float64
	Average float64
	Min     interface{}
	Max     interface{}
}

// GroupResult represents a grouped result
type GroupResult struct {
	Key   interface{}
	Items []interface{}
	Count int
}

// NewLinqEngine creates a new universal LINQ engine
func NewLinqEngine(data []interface{}) *UniversalLinqEngine {
	return &UniversalLinqEngine{
		data: data,
	}
}

// From creates a new LINQ query from data
func From(data interface{}) *LinqResult {
	sliceData := convertToInterfaceSlice(data)
	engine := NewLinqEngine(sliceData)
	return &LinqResult{
		data:   sliceData,
		engine: engine,
	}
}

// Where filters elements based on a predicate
func (lr *LinqResult) Where(predicate func(interface{}) bool) *LinqResult {
	var result []interface{}
	for _, item := range lr.data {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return &LinqResult{data: result, engine: lr.engine}
}

// WhereField filters elements based on field value
func (lr *LinqResult) WhereField(fieldName string, operator string, value interface{}) *LinqResult {
	var result []interface{}
	for _, item := range lr.data {
		fieldValue := getFieldValue(item, fieldName)
		if fieldValue != nil && compareValues(fieldValue, operator, value) {
			result = append(result, item)
		}
	}
	return &LinqResult{data: result, engine: lr.engine}
}

// Select transforms elements using a selector function
func (lr *LinqResult) Select(selector func(interface{}) interface{}) *LinqResult {
	var result []interface{}
	for _, item := range lr.data {
		result = append(result, selector(item))
	}
	return &LinqResult{data: result, engine: lr.engine}
}

// SelectField selects specific fields from elements
func (lr *LinqResult) SelectField(fieldName string) *LinqResult {
	var result []interface{}
	for _, item := range lr.data {
		fieldValue := getFieldValue(item, fieldName)
		if fieldValue != nil {
			result = append(result, fieldValue)
		}
	}
	return &LinqResult{data: result, engine: lr.engine}
}

// SelectFields selects multiple fields and returns as map
func (lr *LinqResult) SelectFields(fieldNames ...string) *LinqResult {
	var result []interface{}
	for _, item := range lr.data {
		selected := make(map[string]interface{})
		for _, fieldName := range fieldNames {
			fieldValue := getFieldValue(item, fieldName)
			if fieldValue != nil {
				selected[fieldName] = fieldValue
			}
		}
		if len(selected) > 0 {
			result = append(result, selected)
		}
	}
	return &LinqResult{data: result, engine: lr.engine}
}

// OrderBy sorts elements in ascending order
func (lr *LinqResult) OrderBy(keySelector func(interface{}) interface{}) *LinqResult {
	result := make([]interface{}, len(lr.data))
	copy(result, lr.data)
	
	sort.Slice(result, func(i, j int) bool {
		key1 := keySelector(result[i])
		key2 := keySelector(result[j])
		return compareForSort(key1, key2) < 0
	})
	
	return &LinqResult{data: result, engine: lr.engine}
}

// OrderByField sorts elements by field name in ascending order
func (lr *LinqResult) OrderByField(fieldName string) *LinqResult {
	result := make([]interface{}, len(lr.data))
	copy(result, lr.data)
	
	sort.Slice(result, func(i, j int) bool {
		val1 := getFieldValue(result[i], fieldName)
		val2 := getFieldValue(result[j], fieldName)
		return compareForSort(val1, val2) < 0
	})
	
	return &LinqResult{data: result, engine: lr.engine}
}

// OrderByDescending sorts elements in descending order
func (lr *LinqResult) OrderByDescending(keySelector func(interface{}) interface{}) *LinqResult {
	result := make([]interface{}, len(lr.data))
	copy(result, lr.data)
	
	sort.Slice(result, func(i, j int) bool {
		key1 := keySelector(result[i])
		key2 := keySelector(result[j])
		return compareForSort(key1, key2) > 0
	})
	
	return &LinqResult{data: result, engine: lr.engine}
}

// OrderByFieldDescending sorts elements by field name in descending order
func (lr *LinqResult) OrderByFieldDescending(fieldName string) *LinqResult {
	result := make([]interface{}, len(lr.data))
	copy(result, lr.data)
	
	sort.Slice(result, func(i, j int) bool {
		val1 := getFieldValue(result[i], fieldName)
		val2 := getFieldValue(result[j], fieldName)
		return compareForSort(val1, val2) > 0
	})
	
	return &LinqResult{data: result, engine: lr.engine}
}

// GroupBy groups elements by a key selector
func (lr *LinqResult) GroupBy(keySelector func(interface{}) interface{}) []*GroupResult {
	groups := make(map[interface{}][]interface{})
	
	for _, item := range lr.data {
		key := keySelector(item)
		groups[key] = append(groups[key], item)
	}
	
	var result []*GroupResult
	for key, items := range groups {
		result = append(result, &GroupResult{
			Key:   key,
			Items: items,
			Count: len(items),
		})
	}
	
	return result
}

// GroupByField groups elements by field name
func (lr *LinqResult) GroupByField(fieldName string) []*GroupResult {
	groups := make(map[interface{}][]interface{})
	
	for _, item := range lr.data {
		key := getFieldValue(item, fieldName)
		if key != nil {
			groups[key] = append(groups[key], item)
		}
	}
	
	var result []*GroupResult
	for key, items := range groups {
		result = append(result, &GroupResult{
			Key:   key,
			Items: items,
			Count: len(items),
		})
	}
	
	return result
}

// Take returns the first n elements
func (lr *LinqResult) Take(count int) *LinqResult {
	if count > len(lr.data) {
		count = len(lr.data)
	}
	result := make([]interface{}, count)
	copy(result, lr.data[:count])
	return &LinqResult{data: result, engine: lr.engine}
}

// Skip skips the first n elements
func (lr *LinqResult) Skip(count int) *LinqResult {
	if count >= len(lr.data) {
		return &LinqResult{data: []interface{}{}, engine: lr.engine}
	}
	result := make([]interface{}, len(lr.data)-count)
	copy(result, lr.data[count:])
	return &LinqResult{data: result, engine: lr.engine}
}

// TakeWhile takes elements while condition is true
func (lr *LinqResult) TakeWhile(predicate func(interface{}) bool) *LinqResult {
	var result []interface{}
	for _, item := range lr.data {
		if predicate(item) {
			result = append(result, item)
		} else {
			break
		}
	}
	return &LinqResult{data: result, engine: lr.engine}
}

// SkipWhile skips elements while condition is true
func (lr *LinqResult) SkipWhile(predicate func(interface{}) bool) *LinqResult {
	var result []interface{}
	skipping := true
	for _, item := range lr.data {
		if skipping && predicate(item) {
			continue
		}
		skipping = false
		result = append(result, item)
	}
	return &LinqResult{data: result, engine: lr.engine}
}

// Distinct returns unique elements
func (lr *LinqResult) Distinct() *LinqResult {
	seen := make(map[interface{}]bool)
	var result []interface{}
	
	for _, item := range lr.data {
		key := fmt.Sprintf("%v", item)
		if !seen[key] {
			seen[key] = true
			result = append(result, item)
		}
	}
	
	return &LinqResult{data: result, engine: lr.engine}
}

// DistinctBy returns unique elements based on key selector
func (lr *LinqResult) DistinctBy(keySelector func(interface{}) interface{}) *LinqResult {
	seen := make(map[interface{}]bool)
	var result []interface{}
	
	for _, item := range lr.data {
		key := keySelector(item)
		if !seen[key] {
			seen[key] = true
			result = append(result, item)
		}
	}
	
	return &LinqResult{data: result, engine: lr.engine}
}

// DistinctByField returns unique elements based on field value
func (lr *LinqResult) DistinctByField(fieldName string) *LinqResult {
	seen := make(map[interface{}]bool)
	var result []interface{}
	
	for _, item := range lr.data {
		key := getFieldValue(item, fieldName)
		if key != nil && !seen[key] {
			seen[key] = true
			result = append(result, item)
		}
	}
	
	return &LinqResult{data: result, engine: lr.engine}
}

// Union combines two sequences and removes duplicates
func (lr *LinqResult) Union(other *LinqResult) *LinqResult {
	combined := append(lr.data, other.data...)
	combinedResult := &LinqResult{data: combined, engine: lr.engine}
	return combinedResult.Distinct()
}

// Intersect returns elements that exist in both sequences
func (lr *LinqResult) Intersect(other *LinqResult) *LinqResult {
	otherSet := make(map[string]bool)
	for _, item := range other.data {
		key := fmt.Sprintf("%v", item)
		otherSet[key] = true
	}
	
	var result []interface{}
	seen := make(map[string]bool)
	
	for _, item := range lr.data {
		key := fmt.Sprintf("%v", item)
		if otherSet[key] && !seen[key] {
			seen[key] = true
			result = append(result, item)
		}
	}
	
	return &LinqResult{data: result, engine: lr.engine}
}

// Except returns elements that don't exist in the other sequence
func (lr *LinqResult) Except(other *LinqResult) *LinqResult {
	otherSet := make(map[string]bool)
	for _, item := range other.data {
		key := fmt.Sprintf("%v", item)
		otherSet[key] = true
	}
	
	var result []interface{}
	
	for _, item := range lr.data {
		key := fmt.Sprintf("%v", item)
		if !otherSet[key] {
			result = append(result, item)
		}
	}
	
	return &LinqResult{data: result, engine: lr.engine}
}

// Reverse reverses the order of elements
func (lr *LinqResult) Reverse() *LinqResult {
	result := make([]interface{}, len(lr.data))
	for i, j := 0, len(lr.data)-1; i <= j; i, j = i+1, j-1 {
		result[i], result[j] = lr.data[j], lr.data[i]
	}
	return &LinqResult{data: result, engine: lr.engine}
}

// Aggregate functions

// Count returns the number of elements
func (lr *LinqResult) Count() int {
	return len(lr.data)
}

// CountWhere returns the number of elements that satisfy a condition
func (lr *LinqResult) CountWhere(predicate func(interface{}) bool) int {
	count := 0
	for _, item := range lr.data {
		if predicate(item) {
			count++
		}
	}
	return count
}

// Any returns true if any element satisfies the condition
func (lr *LinqResult) Any(predicate func(interface{}) bool) bool {
	for _, item := range lr.data {
		if predicate(item) {
			return true
		}
	}
	return false
}

// All returns true if all elements satisfy the condition
func (lr *LinqResult) All(predicate func(interface{}) bool) bool {
	for _, item := range lr.data {
		if !predicate(item) {
			return false
		}
	}
	return true
}

// First returns the first element
func (lr *LinqResult) First() interface{} {
	if len(lr.data) == 0 {
		return nil
	}
	return lr.data[0]
}

// FirstOrDefault returns the first element or default value
func (lr *LinqResult) FirstOrDefault(defaultValue interface{}) interface{} {
	if len(lr.data) == 0 {
		return defaultValue
	}
	return lr.data[0]
}

// FirstWhere returns the first element that satisfies the condition
func (lr *LinqResult) FirstWhere(predicate func(interface{}) bool) interface{} {
	for _, item := range lr.data {
		if predicate(item) {
			return item
		}
	}
	return nil
}

// Last returns the last element
func (lr *LinqResult) Last() interface{} {
	if len(lr.data) == 0 {
		return nil
	}
	return lr.data[len(lr.data)-1]
}

// LastOrDefault returns the last element or default value
func (lr *LinqResult) LastOrDefault(defaultValue interface{}) interface{} {
	if len(lr.data) == 0 {
		return defaultValue
	}
	return lr.data[len(lr.data)-1]
}

// LastWhere returns the last element that satisfies the condition
func (lr *LinqResult) LastWhere(predicate func(interface{}) bool) interface{} {
	for i := len(lr.data) - 1; i >= 0; i-- {
		if predicate(lr.data[i]) {
			return lr.data[i]
		}
	}
	return nil
}

// Single returns the only element (throws error if more than one)
func (lr *LinqResult) Single() interface{} {
	if len(lr.data) == 0 {
		return nil
	}
	if len(lr.data) > 1 {
		panic("sequence contains more than one element")
	}
	return lr.data[0]
}

// SingleOrDefault returns the only element or default value
func (lr *LinqResult) SingleOrDefault(defaultValue interface{}) interface{} {
	if len(lr.data) == 0 {
		return defaultValue
	}
	if len(lr.data) > 1 {
		panic("sequence contains more than one element")
	}
	return lr.data[0]
}

// Sum calculates the sum of numeric field
func (lr *LinqResult) Sum(selector func(interface{}) float64) float64 {
	sum := 0.0
	for _, item := range lr.data {
		sum += selector(item)
	}
	return sum
}

// SumField calculates the sum of a numeric field
func (lr *LinqResult) SumField(fieldName string) float64 {
	sum := 0.0
	for _, item := range lr.data {
		value := getFieldValue(item, fieldName)
		if value != nil {
			sum += toFloat64(value)
		}
	}
	return sum
}

// Average calculates the average of numeric field
func (lr *LinqResult) Average(selector func(interface{}) float64) float64 {
	if len(lr.data) == 0 {
		return 0
	}
	return lr.Sum(selector) / float64(len(lr.data))
}

// AverageField calculates the average of a numeric field
func (lr *LinqResult) AverageField(fieldName string) float64 {
	if len(lr.data) == 0 {
		return 0
	}
	return lr.SumField(fieldName) / float64(len(lr.data))
}

// Min finds the minimum value
func (lr *LinqResult) Min(selector func(interface{}) interface{}) interface{} {
	if len(lr.data) == 0 {
		return nil
	}
	
	min := selector(lr.data[0])
	for i := 1; i < len(lr.data); i++ {
		val := selector(lr.data[i])
		if compareForSort(val, min) < 0 {
			min = val
		}
	}
	return min
}

// MinField finds the minimum value of a field
func (lr *LinqResult) MinField(fieldName string) interface{} {
	if len(lr.data) == 0 {
		return nil
	}
	
	min := getFieldValue(lr.data[0], fieldName)
	for i := 1; i < len(lr.data); i++ {
		val := getFieldValue(lr.data[i], fieldName)
		if val != nil && compareForSort(val, min) < 0 {
			min = val
		}
	}
	return min
}

// Max finds the maximum value
func (lr *LinqResult) Max(selector func(interface{}) interface{}) interface{} {
	if len(lr.data) == 0 {
		return nil
	}
	
	max := selector(lr.data[0])
	for i := 1; i < len(lr.data); i++ {
		val := selector(lr.data[i])
		if compareForSort(val, max) > 0 {
			max = val
		}
	}
	return max
}

// MaxField finds the maximum value of a field
func (lr *LinqResult) MaxField(fieldName string) interface{} {
	if len(lr.data) == 0 {
		return nil
	}
	
	max := getFieldValue(lr.data[0], fieldName)
	for i := 1; i < len(lr.data); i++ {
		val := getFieldValue(lr.data[i], fieldName)
		if val != nil && compareForSort(val, max) > 0 {
			max = val
		}
	}
	return max
}

// Aggregate performs aggregation with accumulator
func (lr *LinqResult) Aggregate(seed interface{}, accumulator func(interface{}, interface{}) interface{}) interface{} {
	result := seed
	for _, item := range lr.data {
		result = accumulator(result, item)
	}
	return result
}

// ToSlice returns the result as a slice
func (lr *LinqResult) ToSlice() []interface{} {
	result := make([]interface{}, len(lr.data))
	copy(result, lr.data)
	return result
}

// ToMap converts result to map using key and value selectors
func (lr *LinqResult) ToMap(keySelector func(interface{}) interface{}, valueSelector func(interface{}) interface{}) map[interface{}]interface{} {
	result := make(map[interface{}]interface{})
	for _, item := range lr.data {
		key := keySelector(item)
		value := valueSelector(item)
		result[key] = value
	}
	return result
}

// Helper functions

func convertToInterfaceSlice(data interface{}) []interface{} {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Slice {
		return []interface{}{data}
	}
	
	result := make([]interface{}, v.Len())
	for i := 0; i < v.Len(); i++ {
		result[i] = v.Index(i).Interface()
	}
	
	return result
}

func getFieldValue(item interface{}, fieldName string) interface{} {
	v := reflect.ValueOf(item)
	t := reflect.TypeOf(item)
	
	// Handle map[string]interface{}
	if v.Kind() == reflect.Map {
		mapValue := v.MapIndex(reflect.ValueOf(fieldName))
		if mapValue.IsValid() {
			return mapValue.Interface()
		}
		return nil
	}
	
	// Handle pointer types
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}
	
	// Handle structs
	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			
			// Check tag first
			if tag := field.Tag.Get("linq"); tag == fieldName {
				return v.Field(i).Interface()
			}
			
			// Check field name (case insensitive)
			if strings.EqualFold(field.Name, fieldName) {
				return v.Field(i).Interface()
			}
		}
	}
	
	return nil
}

func compareValues(fieldValue interface{}, operator string, compareValue interface{}) bool {
	switch strings.ToLower(operator) {
	case "==", "equals", "eq":
		return fmt.Sprintf("%v", fieldValue) == fmt.Sprintf("%v", compareValue)
	case "!=", "not_equals", "ne":
		return fmt.Sprintf("%v", fieldValue) != fmt.Sprintf("%v", compareValue)
	case ">", "greater", "gt":
		return compareForSort(fieldValue, compareValue) > 0
	case ">=", "greater_equal", "ge":
		return compareForSort(fieldValue, compareValue) >= 0
	case "<", "less", "lt":
		return compareForSort(fieldValue, compareValue) < 0
	case "<=", "less_equal", "le":
		return compareForSort(fieldValue, compareValue) <= 0
	case "contains":
		return strings.Contains(strings.ToLower(fmt.Sprintf("%v", fieldValue)), strings.ToLower(fmt.Sprintf("%v", compareValue)))
	case "starts_with":
		return strings.HasPrefix(strings.ToLower(fmt.Sprintf("%v", fieldValue)), strings.ToLower(fmt.Sprintf("%v", compareValue)))
	case "ends_with":
		return strings.HasSuffix(strings.ToLower(fmt.Sprintf("%v", fieldValue)), strings.ToLower(fmt.Sprintf("%v", compareValue)))
	}
	return false
}

func compareForSort(a, b interface{}) int {
	aFloat := toFloat64(a)
	bFloat := toFloat64(b)
	
	if aFloat != 0 || bFloat != 0 {
		if aFloat > bFloat {
			return 1
		} else if aFloat < bFloat {
			return -1
		}
		return 0
	}
	
	// String comparison
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return strings.Compare(aStr, bStr)
}

func toFloat64(value interface{}) float64 {
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