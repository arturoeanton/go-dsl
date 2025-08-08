package universal

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// ValidationRule represents a validation rule for a field
type ValidationRule struct {
	FieldName   string            `yaml:"field"`
	Type        string            `yaml:"type"`
	Required    bool              `yaml:"required"`
	Min         *float64          `yaml:"min,omitempty"`
	Max         *float64          `yaml:"max,omitempty"`
	MinLength   *int              `yaml:"minLength,omitempty"`
	MaxLength   *int              `yaml:"maxLength,omitempty"`
	Pattern     string            `yaml:"pattern,omitempty"`
	Enum        []interface{}     `yaml:"enum,omitempty"`
	Nested      *ValidationSchema `yaml:"nested,omitempty"`
	ArrayItems  *ValidationRule   `yaml:"items,omitempty"`
	Description string            `yaml:"description,omitempty"`
}

// ValidationSchema represents a collection of validation rules
type ValidationSchema struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Rules       []*ValidationRule `yaml:"rules"`
}

// StructValidator validates Go structs and maps against rules
type StructValidator struct {
	schema *ValidationSchema
	rules  map[string]*ValidationRule
}

// NewStructValidator creates a new struct validator
func NewStructValidator() *StructValidator {
	return &StructValidator{
		rules: make(map[string]*ValidationRule),
	}
}

// LoadSchemaFromYAML loads validation rules from a YAML string
func (v *StructValidator) LoadSchemaFromYAML(yamlContent string) error {
	var schema ValidationSchema
	err := yaml.Unmarshal([]byte(yamlContent), &schema)
	if err != nil {
		return fmt.Errorf("error parsing YAML: %v", err)
	}

	v.schema = &schema
	
	// Build rules map
	v.rules = make(map[string]*ValidationRule)
	for _, rule := range schema.Rules {
		v.rules[rule.FieldName] = rule
	}

	return nil
}

// LoadSchemaFromFile loads validation rules from a YAML file
func (v *StructValidator) LoadSchemaFromFile(filepath string) error {
	content, err := ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}
	return v.LoadSchemaFromYAML(string(content))
}

// GetSchema returns the current validation schema
func (v *StructValidator) GetSchema() *ValidationSchema {
	return v.schema
}

// GetRules returns the current validation rules map
func (v *StructValidator) GetRules() map[string]*ValidationRule {
	return v.rules
}

// ValidateStruct validates a Go struct against the rules
func (v *StructValidator) ValidateStruct(data interface{}) (bool, []string) {
	val := reflect.ValueOf(data)
	typ := reflect.TypeOf(data)
	
	// Handle pointer to struct
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}
	
	if val.Kind() != reflect.Struct {
		return false, []string{"Input must be a struct"}
	}
	
	errors := []string{}
	
	// Check each rule
	for fieldName, rule := range v.rules {
		// Find the field in the struct
		fieldVal, _, found := v.findField(val, typ, fieldName)
		
		// Check required fields
		if rule.Required && !found {
			errors = append(errors, fmt.Sprintf("Field '%s' is required", fieldName))
			continue
		}
		
		if !found {
			continue
		}
		
		// Validate the field value
		fieldErrors := v.validateField(fieldName, fieldVal, rule)
		errors = append(errors, fieldErrors...)
	}
	
	return len(errors) == 0, errors
}

// ValidateMap validates a map[string]interface{} against the rules
func (v *StructValidator) ValidateMap(data map[string]interface{}) (bool, []string) {
	errors := []string{}
	
	for fieldName, rule := range v.rules {
		value, exists := data[fieldName]
		
		// Check required fields
		if rule.Required && !exists {
			errors = append(errors, fmt.Sprintf("Field '%s' is required", fieldName))
			continue
		}
		
		if !exists {
			continue
		}
		
		// Validate the field value
		fieldErrors := v.validateValue(fieldName, value, rule)
		errors = append(errors, fieldErrors...)
	}
	
	return len(errors) == 0, errors
}

// findField finds a field in a struct by name (case-insensitive)
func (v *StructValidator) findField(val reflect.Value, typ reflect.Type, fieldName string) (reflect.Value, reflect.Type, bool) {
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		
		// Check field name (case-insensitive)
		if strings.EqualFold(field.Name, fieldName) {
			return val.Field(i), field.Type, true
		}
		
		// Check struct tag
		tag := field.Tag.Get("json")
		if tag != "" {
			tagName := strings.Split(tag, ",")[0]
			if tagName == fieldName {
				return val.Field(i), field.Type, true
			}
		}
		
		// Check yaml tag
		tag = field.Tag.Get("yaml")
		if tag != "" {
			tagName := strings.Split(tag, ",")[0]
			if tagName == fieldName {
				return val.Field(i), field.Type, true
			}
		}
	}
	
	return reflect.Value{}, nil, false
}

// validateField validates a struct field
func (v *StructValidator) validateField(fieldName string, fieldVal reflect.Value, rule *ValidationRule) []string {
	errors := []string{}
	
	// Handle zero values
	if !fieldVal.IsValid() || (fieldVal.Kind() == reflect.Ptr && fieldVal.IsNil()) {
		if rule.Required {
			errors = append(errors, fmt.Sprintf("Field '%s' is required", fieldName))
		}
		return errors
	}
	
	// Dereference pointer if necessary
	if fieldVal.Kind() == reflect.Ptr {
		fieldVal = fieldVal.Elem()
	}
	
	// Type validation
	switch rule.Type {
	case "string":
		if fieldVal.Kind() != reflect.String {
			errors = append(errors, fmt.Sprintf("Field '%s' must be a string", fieldName))
			return errors
		}
		
		str := fieldVal.String()
		
		// String constraints
		if rule.MinLength != nil && len(str) < *rule.MinLength {
			errors = append(errors, fmt.Sprintf("Field '%s' must have at least %d characters", fieldName, *rule.MinLength))
		}
		if rule.MaxLength != nil && len(str) > *rule.MaxLength {
			errors = append(errors, fmt.Sprintf("Field '%s' must have at most %d characters", fieldName, *rule.MaxLength))
		}
		if rule.Pattern != "" {
			if matched, _ := regexp.MatchString(rule.Pattern, str); !matched {
				errors = append(errors, fmt.Sprintf("Field '%s' must match pattern %s", fieldName, rule.Pattern))
			}
		}
		
		// Enum validation
		if len(rule.Enum) > 0 {
			found := false
			for _, allowed := range rule.Enum {
				if fmt.Sprintf("%v", str) == fmt.Sprintf("%v", allowed) {
					found = true
					break
				}
			}
			if !found {
				errors = append(errors, fmt.Sprintf("Field '%s' must be one of %v", fieldName, rule.Enum))
			}
		}
		
	case "number", "integer":
		var num float64
		switch fieldVal.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			num = float64(fieldVal.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			num = float64(fieldVal.Uint())
		case reflect.Float32, reflect.Float64:
			num = fieldVal.Float()
		default:
			errors = append(errors, fmt.Sprintf("Field '%s' must be a number", fieldName))
			return errors
		}
		
		// Integer check
		if rule.Type == "integer" && num != float64(int(num)) {
			errors = append(errors, fmt.Sprintf("Field '%s' must be an integer", fieldName))
		}
		
		// Number constraints
		if rule.Min != nil && num < *rule.Min {
			errors = append(errors, fmt.Sprintf("Field '%s' must be >= %v", fieldName, *rule.Min))
		}
		if rule.Max != nil && num > *rule.Max {
			errors = append(errors, fmt.Sprintf("Field '%s' must be <= %v", fieldName, *rule.Max))
		}
		
	case "boolean":
		if fieldVal.Kind() != reflect.Bool {
			errors = append(errors, fmt.Sprintf("Field '%s' must be a boolean", fieldName))
		}
		
	case "array", "slice":
		if fieldVal.Kind() != reflect.Slice && fieldVal.Kind() != reflect.Array {
			errors = append(errors, fmt.Sprintf("Field '%s' must be an array", fieldName))
			return errors
		}
		
		// Validate array items if rule specified
		if rule.ArrayItems != nil {
			for i := 0; i < fieldVal.Len(); i++ {
				itemErrors := v.validateField(fmt.Sprintf("%s[%d]", fieldName, i), fieldVal.Index(i), rule.ArrayItems)
				errors = append(errors, itemErrors...)
			}
		}
		
	case "object", "struct":
		if fieldVal.Kind() != reflect.Struct && fieldVal.Kind() != reflect.Map {
			errors = append(errors, fmt.Sprintf("Field '%s' must be an object", fieldName))
			return errors
		}
		
		// Validate nested object if schema specified
		if rule.Nested != nil {
			nestedValidator := NewStructValidator()
			nestedValidator.schema = rule.Nested
			nestedValidator.rules = make(map[string]*ValidationRule)
			for _, r := range rule.Nested.Rules {
				nestedValidator.rules[r.FieldName] = r
			}
			
			if fieldVal.Kind() == reflect.Struct {
				valid, nestedErrors := nestedValidator.ValidateStruct(fieldVal.Interface())
				if !valid {
					for _, err := range nestedErrors {
						errors = append(errors, fmt.Sprintf("%s.%s", fieldName, err))
					}
				}
			} else if fieldVal.Kind() == reflect.Map {
				// Convert to map[string]interface{}
				m := make(map[string]interface{})
				for _, key := range fieldVal.MapKeys() {
					m[key.String()] = fieldVal.MapIndex(key).Interface()
				}
				valid, nestedErrors := nestedValidator.ValidateMap(m)
				if !valid {
					for _, err := range nestedErrors {
						errors = append(errors, fmt.Sprintf("%s.%s", fieldName, err))
					}
				}
			}
		}
	}
	
	return errors
}

// validateValue validates a value from a map
func (v *StructValidator) validateValue(fieldName string, value interface{}, rule *ValidationRule) []string {
	errors := []string{}
	
	// Type validation
	switch rule.Type {
	case "string":
		str, ok := value.(string)
		if !ok {
			errors = append(errors, fmt.Sprintf("Field '%s' must be a string", fieldName))
			return errors
		}
		
		// String constraints
		if rule.MinLength != nil && len(str) < *rule.MinLength {
			errors = append(errors, fmt.Sprintf("Field '%s' must have at least %d characters", fieldName, *rule.MinLength))
		}
		if rule.MaxLength != nil && len(str) > *rule.MaxLength {
			errors = append(errors, fmt.Sprintf("Field '%s' must have at most %d characters", fieldName, *rule.MaxLength))
		}
		if rule.Pattern != "" {
			if matched, _ := regexp.MatchString(rule.Pattern, str); !matched {
				errors = append(errors, fmt.Sprintf("Field '%s' must match pattern %s", fieldName, rule.Pattern))
			}
		}
		
	case "number", "integer":
		num, ok := toNumber(value)
		if !ok {
			errors = append(errors, fmt.Sprintf("Field '%s' must be a number", fieldName))
			return errors
		}
		
		// Integer check
		if rule.Type == "integer" && num != float64(int(num)) {
			errors = append(errors, fmt.Sprintf("Field '%s' must be an integer", fieldName))
		}
		
		// Number constraints
		if rule.Min != nil && num < *rule.Min {
			errors = append(errors, fmt.Sprintf("Field '%s' must be >= %v", fieldName, *rule.Min))
		}
		if rule.Max != nil && num > *rule.Max {
			errors = append(errors, fmt.Sprintf("Field '%s' must be <= %v", fieldName, *rule.Max))
		}
		
	case "boolean":
		if _, ok := value.(bool); !ok {
			errors = append(errors, fmt.Sprintf("Field '%s' must be a boolean", fieldName))
		}
		
	case "array", "slice":
		arr, ok := value.([]interface{})
		if !ok {
			errors = append(errors, fmt.Sprintf("Field '%s' must be an array", fieldName))
			return errors
		}
		
		// Validate array items if rule specified
		if rule.ArrayItems != nil {
			for i, item := range arr {
				itemErrors := v.validateValue(fmt.Sprintf("%s[%d]", fieldName, i), item, rule.ArrayItems)
				errors = append(errors, itemErrors...)
			}
		}
		
	case "object", "struct":
		obj, ok := value.(map[string]interface{})
		if !ok {
			errors = append(errors, fmt.Sprintf("Field '%s' must be an object", fieldName))
			return errors
		}
		
		// Validate nested object if schema specified
		if rule.Nested != nil {
			nestedValidator := NewStructValidator()
			nestedValidator.schema = rule.Nested
			nestedValidator.rules = make(map[string]*ValidationRule)
			for _, r := range rule.Nested.Rules {
				nestedValidator.rules[r.FieldName] = r
			}
			
			valid, nestedErrors := nestedValidator.ValidateMap(obj)
			if !valid {
				for _, err := range nestedErrors {
					errors = append(errors, fmt.Sprintf("%s.%s", fieldName, err))
				}
			}
		}
	}
	
	// Enum validation
	if len(rule.Enum) > 0 {
		found := false
		for _, allowed := range rule.Enum {
			if fmt.Sprintf("%v", value) == fmt.Sprintf("%v", allowed) {
				found = true
				break
			}
		}
		if !found {
			errors = append(errors, fmt.Sprintf("Field '%s' must be one of %v", fieldName, rule.Enum))
		}
	}
	
	return errors
}

// toNumber converts various numeric types to float64
func toNumber(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case int32:
		return float64(v), true
	case int8:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint64:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint8:
		return float64(v), true
	default:
		return 0, false
	}
}