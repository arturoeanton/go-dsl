package universal

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// UniversalRuleEngine handles generic rule processing on any struct type using reflection
type UniversalRuleEngine struct {
	facts    []interface{}
	rules    []*Rule
	executed map[string]bool
}

// Rule represents a business rule with conditions and actions
type Rule struct {
	Name       string
	Priority   int
	Conditions []Condition
	Actions    []Action
	Salience   int // Higher salience = higher priority
}

// Condition represents a rule condition
type Condition struct {
	Field    string
	Operator string
	Value    interface{}
	Entity   string
}

// Action represents a rule action
type Action struct {
	Type   string // "set", "modify", "insert", "retract", "execute"
	Target string
	Field  string
	Value  interface{}
	Func   func(facts []interface{}) error
}

// NewRuleEngine creates a new universal rule engine
func NewRuleEngine() *UniversalRuleEngine {
	return &UniversalRuleEngine{
		facts:    make([]interface{}, 0),
		rules:    make([]*Rule, 0),
		executed: make(map[string]bool),
	}
}

// InsertFact adds a fact to the working memory
func (re *UniversalRuleEngine) InsertFact(fact interface{}) {
	re.facts = append(re.facts, fact)
}

// AddRule adds a rule to the engine
func (re *UniversalRuleEngine) AddRule(rule *Rule) {
	re.rules = append(re.rules, rule)
}

// FireAllRules executes all matching rules
func (re *UniversalRuleEngine) FireAllRules() error {
	// Sort rules by salience (priority)
	re.sortRulesBySalience()

	changed := true
	for changed {
		changed = false
		for _, rule := range re.rules {
			ruleKey := fmt.Sprintf("%s-%v", rule.Name, re.facts)
			if re.executed[ruleKey] {
				continue
			}

			if re.evaluateRule(rule) {
				err := re.executeRule(rule)
				if err != nil {
					return err
				}
				re.executed[ruleKey] = true
				changed = true
			}
		}
	}

	return nil
}

// GetFieldNames extracts field names from any struct, supporting drools tags
func (re *UniversalRuleEngine) GetFieldNames(item interface{}) []string {
	v := reflect.ValueOf(item)
	t := reflect.TypeOf(item)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	var fields []string
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)

		if tag := field.Tag.Get("drools"); tag != "" {
			fields = append(fields, tag)
		} else {
			fields = append(fields, strings.ToLower(field.Name))
		}
	}

	return fields
}

// GetFieldValue gets the value of a specific field from a struct using reflection
func (re *UniversalRuleEngine) GetFieldValue(item interface{}, fieldName string) interface{} {
	v := reflect.ValueOf(item)
	t := reflect.TypeOf(item)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)

		if tag := field.Tag.Get("drools"); tag == fieldName {
			return v.Field(i).Interface()
		}

		if strings.ToLower(field.Name) == fieldName {
			return v.Field(i).Interface()
		}
	}

	return nil
}

// SetFieldValue sets the value of a specific field in a struct using reflection
func (re *UniversalRuleEngine) SetFieldValue(item interface{}, fieldName string, value interface{}) error {
	v := reflect.ValueOf(item)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("item must be a pointer to modify fields")
	}

	v = v.Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)

		match := false
		if tag := field.Tag.Get("drools"); tag == fieldName {
			match = true
		} else if strings.ToLower(field.Name) == fieldName {
			match = true
		}

		if match {
			fieldValue := v.Field(i)
			if !fieldValue.CanSet() {
				return fmt.Errorf("field %s cannot be set", fieldName)
			}

			// Convert value to field type
			convertedValue, err := re.convertValue(value, fieldValue.Type())
			if err != nil {
				return err
			}

			fieldValue.Set(convertedValue)
			return nil
		}
	}

	return fmt.Errorf("field %s not found", fieldName)
}

// FindFactsByType finds all facts of a specific type
func (re *UniversalRuleEngine) FindFactsByType(typeName string) []interface{} {
	var result []interface{}

	for _, fact := range re.facts {
		t := reflect.TypeOf(fact)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		if strings.ToLower(t.Name()) == strings.ToLower(typeName) {
			result = append(result, fact)
		}
	}

	return result
}

// evaluateRule checks if all conditions of a rule are met
func (re *UniversalRuleEngine) evaluateRule(rule *Rule) bool {
	for _, condition := range rule.Conditions {
		if !re.evaluateCondition(condition) {
			return false
		}
	}
	return true
}

// evaluateCondition checks if a single condition is met
func (re *UniversalRuleEngine) evaluateCondition(condition Condition) bool {
	facts := re.FindFactsByType(condition.Entity)

	for _, fact := range facts {
		fieldValue := re.GetFieldValue(fact, condition.Field)
		if fieldValue == nil {
			continue
		}

		if re.compareValues(fieldValue, condition.Operator, condition.Value) {
			return true
		}
	}

	return false
}

// executeRule executes all actions of a rule
func (re *UniversalRuleEngine) executeRule(rule *Rule) error {
	for _, action := range rule.Actions {
		err := re.executeAction(action)
		if err != nil {
			return err
		}
	}
	return nil
}

// executeAction executes a single action
func (re *UniversalRuleEngine) executeAction(action Action) error {
	switch action.Type {
	case "set":
		return re.executeSetAction(action)
	case "modify":
		return re.executeModifyAction(action)
	case "insert":
		return re.executeInsertAction(action)
	case "retract":
		return re.executeRetractAction(action)
	case "execute":
		if action.Func != nil {
			return action.Func(re.facts)
		}
	}
	return nil
}

// executeSetAction sets a field value in matching facts
func (re *UniversalRuleEngine) executeSetAction(action Action) error {
	facts := re.FindFactsByType(action.Target)

	for _, fact := range facts {
		err := re.SetFieldValue(fact, action.Field, action.Value)
		if err != nil {
			return err
		}
	}

	return nil
}

// executeModifyAction modifies matching facts
func (re *UniversalRuleEngine) executeModifyAction(action Action) error {
	return re.executeSetAction(action) // Same as set for now
}

// executeInsertAction inserts a new fact
func (re *UniversalRuleEngine) executeInsertAction(action Action) error {
	re.InsertFact(action.Value)
	return nil
}

// executeRetractAction removes matching facts
func (re *UniversalRuleEngine) executeRetractAction(action Action) error {
	newFacts := make([]interface{}, 0)

	for _, fact := range re.facts {
		t := reflect.TypeOf(fact)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		if strings.ToLower(t.Name()) != strings.ToLower(action.Target) {
			newFacts = append(newFacts, fact)
		}
	}

	re.facts = newFacts
	return nil
}

// compareValues compares two values using the given operator
func (re *UniversalRuleEngine) compareValues(fieldValue interface{}, operator string, compareValue interface{}) bool {
	switch operator {
	case "==", "equals", "es":
		return re.equalCompare(fieldValue, compareValue)
	case ">", "greater", "mayor":
		return re.numericCompare(fieldValue, compareValue) > 0
	case "<", "less", "menor":
		return re.numericCompare(fieldValue, compareValue) < 0
	case ">=", "greater_equal", "mayor_igual":
		return re.numericCompare(fieldValue, compareValue) >= 0
	case "<=", "less_equal", "menor_igual":
		return re.numericCompare(fieldValue, compareValue) <= 0
	case "contains", "contiene":
		return re.containsCompare(fieldValue, compareValue)
	case "matches", "coincide":
		return re.matchesCompare(fieldValue, compareValue)
	}
	return false
}

// equalCompare compares two values for equality
func (re *UniversalRuleEngine) equalCompare(a, b interface{}) bool {
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return strings.EqualFold(aStr, bStr)
}

// numericCompare compares two values numerically
func (re *UniversalRuleEngine) numericCompare(a, b interface{}) int {
	aFloat := re.toFloat64(a)
	bFloat := re.toFloat64(b)

	if aFloat > bFloat {
		return 1
	} else if aFloat < bFloat {
		return -1
	}
	return 0
}

// containsCompare checks if a contains b (case insensitive)
func (re *UniversalRuleEngine) containsCompare(a, b interface{}) bool {
	aStr := strings.ToLower(fmt.Sprintf("%v", a))
	bStr := strings.ToLower(fmt.Sprintf("%v", b))
	return strings.Contains(aStr, bStr)
}

// matchesCompare checks if a matches pattern b
func (re *UniversalRuleEngine) matchesCompare(a, b interface{}) bool {
	// Simple pattern matching for now
	return re.containsCompare(a, b)
}

// toFloat64 converts any numeric type to float64
func (re *UniversalRuleEngine) toFloat64(value interface{}) float64 {
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

// convertValue converts a value to the target type
func (re *UniversalRuleEngine) convertValue(value interface{}, targetType reflect.Type) (reflect.Value, error) {
	switch targetType.Kind() {
	case reflect.String:
		return reflect.ValueOf(fmt.Sprintf("%v", value)), nil
	case reflect.Int:
		if f := re.toFloat64(value); f != 0 {
			return reflect.ValueOf(int(f)), nil
		}
		return reflect.ValueOf(0), nil
	case reflect.Float64:
		return reflect.ValueOf(re.toFloat64(value)), nil
	case reflect.Bool:
		str := strings.ToLower(fmt.Sprintf("%v", value))
		return reflect.ValueOf(str == "true" || str == "1"), nil
	default:
		return reflect.ValueOf(value), nil
	}
}

// sortRulesBySalience sorts rules by salience (priority)
func (re *UniversalRuleEngine) sortRulesBySalience() {
	// Simple bubble sort by salience
	for i := 0; i < len(re.rules); i++ {
		for j := i + 1; j < len(re.rules); j++ {
			if re.rules[i].Salience < re.rules[j].Salience {
				re.rules[i], re.rules[j] = re.rules[j], re.rules[i]
			}
		}
	}
}

// GetFacts returns all facts in working memory
func (re *UniversalRuleEngine) GetFacts() []interface{} {
	return re.facts
}

// ClearFacts removes all facts from working memory
func (re *UniversalRuleEngine) ClearFacts() {
	re.facts = make([]interface{}, 0)
	re.executed = make(map[string]bool)
}

// FormatFact formats a single fact for display
func (re *UniversalRuleEngine) FormatFact(fact interface{}) string {
	v := reflect.ValueOf(fact)
	t := reflect.TypeOf(fact)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	var parts []string
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i).Interface()

		name := field.Name
		if tag := field.Tag.Get("drools"); tag != "" {
			name = tag
		}

		parts = append(parts, fmt.Sprintf("%s: %v", name, value))
	}

	return fmt.Sprintf("%s{%s}", t.Name(), strings.Join(parts, ", "))
}
