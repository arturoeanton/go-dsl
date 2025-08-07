package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

// ValidationRule represents a validation rule for a field
type ValidationRule struct {
	FieldName   string        `yaml:"field"`
	Type        string        `yaml:"type"`
	Required    bool          `yaml:"required"`
	Min         *float64      `yaml:"min,omitempty"`
	Max         *float64      `yaml:"max,omitempty"`
	MinLength   *int          `yaml:"minLength,omitempty"`
	MaxLength   *int          `yaml:"maxLength,omitempty"`
	Pattern     string        `yaml:"pattern,omitempty"`
	Enum        []interface{} `yaml:"enum,omitempty"`
	Description string        `yaml:"description,omitempty"`
}

// ValidationSchema represents a collection of validation rules
type ValidationSchema struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Rules       []*ValidationRule `yaml:"rules"`
}

// Validator validates JSON data against rules
type Validator struct {
	schema *ValidationSchema
	rules  map[string]*ValidationRule
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		rules: make(map[string]*ValidationRule),
	}
}

// LoadSchemaFromYAML loads validation rules from a YAML file
func (v *Validator) LoadSchemaFromYAML(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading YAML file: %v", err)
	}

	var schema ValidationSchema
	err = yaml.Unmarshal(data, &schema)
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

// LoadSchemaFromYAMLString loads validation rules from a YAML string
func (v *Validator) LoadSchemaFromYAMLString(yamlContent string) error {
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

// AddRule adds a validation rule programmatically
func (v *Validator) AddRule(rule *ValidationRule) {
	v.rules[rule.FieldName] = rule
}

// GetSchemaInfo returns schema name and description
func (v *Validator) GetSchemaInfo() (string, string) {
	if v.schema != nil {
		return v.schema.Name, v.schema.Description
	}
	return "", ""
}

// Validate validates JSON data against the rules
func (v *Validator) Validate(jsonData string) (bool, []string) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return false, []string{fmt.Sprintf("Invalid JSON: %v", err)}
	}
	
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
		
		// Type validation
		switch rule.Type {
		case "string":
			str, ok := value.(string)
			if !ok {
				errors = append(errors, fmt.Sprintf("Field '%s' must be a string", fieldName))
				continue
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
				continue
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
			
		case "array":
			if _, ok := value.([]interface{}); !ok {
				errors = append(errors, fmt.Sprintf("Field '%s' must be an array", fieldName))
			}
			
		case "object":
			if _, ok := value.(map[string]interface{}); !ok {
				errors = append(errors, fmt.Sprintf("Field '%s' must be an object", fieldName))
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
	}
	
	return len(errors) == 0, errors
}

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
	default:
		return 0, false
	}
}

func main() {
	fmt.Println("=== JSON Validator with YAML Configuration ===\n")

	// Create validator
	validator := NewValidator()

	// Example 1: Load from YAML string (embedded)
	fmt.Println("Example 1: User Registration Schema (from YAML string)")
	fmt.Println("-------------------------------------------------------")
	
	userSchemaYAML := `
name: UserRegistration
description: Schema for user registration validation
rules:
  - field: username
    type: string
    required: true
    minLength: 3
    maxLength: 20
    pattern: "^[a-zA-Z0-9_]+$"
    description: Username for the account
  
  - field: email
    type: string
    required: true
    pattern: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
    description: User email address
  
  - field: age
    type: integer
    required: true
    min: 18
    max: 120
    description: User age
  
  - field: password
    type: string
    required: true
    minLength: 8
    maxLength: 100
    description: Account password
  
  - field: country
    type: string
    required: false
    enum: ["USA", "Canada", "Mexico", "UK", "Germany", "France"]
    description: Country of residence
  
  - field: newsletter
    type: boolean
    required: false
    description: Newsletter subscription
`

	// Load schema from YAML string
	err := validator.LoadSchemaFromYAMLString(userSchemaYAML)
	if err != nil {
		fmt.Printf("Error loading schema: %v\n", err)
		return
	}

	name, desc := validator.GetSchemaInfo()
	fmt.Printf("Loaded Schema: %s\n", name)
	fmt.Printf("Description: %s\n\n", desc)

	// Test valid user
	validUser := `{
		"username": "john_doe",
		"email": "john@example.com",
		"age": 25,
		"password": "SecurePass123",
		"country": "USA",
		"newsletter": true
	}`
	
	fmt.Println("Testing valid user:")
	valid, errors := validator.Validate(validUser)
	if valid {
		fmt.Println("✓ Validation passed!\n")
	} else {
		fmt.Println("✗ Validation failed:")
		for _, err := range errors {
			fmt.Println("  -", err)
		}
		fmt.Println()
	}

	// Test invalid user
	invalidUser := `{
		"username": "jo",
		"email": "invalid-email",
		"age": 15,
		"password": "short"
	}`
	
	fmt.Println("Testing invalid user:")
	valid, errors = validator.Validate(invalidUser)
	if valid {
		fmt.Println("✓ Validation passed!\n")
	} else {
		fmt.Println("✗ Validation failed:")
		for _, err := range errors {
			fmt.Println("  -", err)
		}
		fmt.Println()
	}

	// Example 2: Load from YAML file
	fmt.Println("\nExample 2: Product Schema (from YAML file)")
	fmt.Println("-------------------------------------------")

	// Create a sample YAML file
	productSchemaYAML := `
name: ProductCatalog
description: Schema for product catalog validation
rules:
  - field: id
    type: string
    required: true
    pattern: "^PROD-[0-9]{4,}$"
    description: Product ID

  - field: name
    type: string
    required: true
    minLength: 1
    maxLength: 100
    description: Product name

  - field: price
    type: number
    required: true
    min: 0.01
    max: 999999.99
    description: Product price

  - field: stock
    type: integer
    required: true
    min: 0
    description: Stock quantity

  - field: category
    type: string
    required: true
    enum: ["Electronics", "Clothing", "Food", "Books", "Sports"]
    description: Product category

  - field: tags
    type: array
    required: false
    description: Product tags

  - field: discount
    type: number
    required: false
    min: 0
    max: 100
    description: Discount percentage
`

	// Write to file
	err = ioutil.WriteFile("product_schema.yaml", []byte(productSchemaYAML), 0644)
	if err != nil {
		fmt.Printf("Error writing YAML file: %v\n", err)
		return
	}
	defer os.Remove("product_schema.yaml") // Clean up

	// Create new validator for products
	productValidator := NewValidator()
	err = productValidator.LoadSchemaFromYAML("product_schema.yaml")
	if err != nil {
		fmt.Printf("Error loading schema from file: %v\n", err)
		return
	}

	name, desc = productValidator.GetSchemaInfo()
	fmt.Printf("Loaded Schema: %s\n", name)
	fmt.Printf("Description: %s\n\n", desc)

	// Test valid product
	validProduct := `{
		"id": "PROD-12345",
		"name": "Wireless Headphones",
		"price": 79.99,
		"stock": 150,
		"category": "Electronics",
		"tags": ["wireless", "bluetooth", "audio"],
		"discount": 10
	}`
	
	fmt.Println("Testing valid product:")
	valid, errors = productValidator.Validate(validProduct)
	if valid {
		fmt.Println("✓ Validation passed!\n")
	} else {
		fmt.Println("✗ Validation failed:")
		for _, err := range errors {
			fmt.Println("  -", err)
		}
		fmt.Println()
	}

	// Test invalid product
	invalidProduct := `{
		"id": "12345",
		"name": "",
		"price": -10,
		"stock": -5,
		"category": "InvalidCategory"
	}`
	
	fmt.Println("Testing invalid product:")
	valid, errors = productValidator.Validate(invalidProduct)
	if valid {
		fmt.Println("✓ Validation passed!\n")
	} else {
		fmt.Println("✗ Validation failed:")
		for _, err := range errors {
			fmt.Println("  -", err)
		}
		fmt.Println()
	}

	// Example 3: Mixed approach - YAML base with programmatic additions
	fmt.Println("\nExample 3: API Schema (Mixed YAML + Programmatic)")
	fmt.Println("--------------------------------------------------")

	apiSchemaYAML := `
name: APIConfiguration
description: Base API configuration schema
rules:
  - field: endpoint
    type: string
    required: true
    pattern: "^https?://.*"
    description: API endpoint URL

  - field: timeout
    type: integer
    required: false
    min: 1
    max: 300
    description: Request timeout in seconds

  - field: methods
    type: array
    required: true
    description: Allowed HTTP methods
`

	// Create validator and load base schema
	apiValidator := NewValidator()
	err = apiValidator.LoadSchemaFromYAMLString(apiSchemaYAML)
	if err != nil {
		fmt.Printf("Error loading schema: %v\n", err)
		return
	}

	// Add additional rules programmatically
	minRetries := 0.0
	maxRetries := 10.0
	apiValidator.AddRule(&ValidationRule{
		FieldName:   "retries",
		Type:        "integer",
		Required:    false,
		Min:         &minRetries,
		Max:         &maxRetries,
		Description: "Number of retry attempts (added programmatically)",
	})

	minRateLimit := 1.0
	maxRateLimit := 10000.0
	apiValidator.AddRule(&ValidationRule{
		FieldName:   "rateLimit",
		Type:        "integer",
		Required:    false,
		Min:         &minRateLimit,
		Max:         &maxRateLimit,
		Description: "Requests per minute (added programmatically)",
	})

	fmt.Println("Schema loaded from YAML with programmatic additions")
	fmt.Println()

	// Test valid API config
	validAPI := `{
		"endpoint": "https://api.example.com/v1",
		"timeout": 60,
		"methods": ["GET", "POST", "PUT"],
		"retries": 5,
		"rateLimit": 500
	}`
	
	fmt.Println("Testing valid API configuration:")
	valid, errors = apiValidator.Validate(validAPI)
	if valid {
		fmt.Println("✓ Validation passed!\n")
	} else {
		fmt.Println("✗ Validation failed:")
		for _, err := range errors {
			fmt.Println("  -", err)
		}
		fmt.Println()
	}

	// Show schema export capability
	fmt.Println("\nExample 4: Exporting Schema to YAML")
	fmt.Println("------------------------------------")
	
	// Create a schema programmatically
	exportValidator := NewValidator()
	minNameLen := 2
	maxNameLen := 50
	exportValidator.AddRule(&ValidationRule{
		FieldName:   "firstName",
		Type:        "string",
		Required:    true,
		MinLength:   &minNameLen,
		MaxLength:   &maxNameLen,
		Description: "First name",
	})
	
	exportValidator.AddRule(&ValidationRule{
		FieldName:   "lastName",
		Type:        "string",
		Required:    true,
		MinLength:   &minNameLen,
		MaxLength:   &maxNameLen,
		Description: "Last name",
	})
	
	minSalary := 0.0
	maxSalary := 1000000.0
	exportValidator.AddRule(&ValidationRule{
		FieldName:   "salary",
		Type:        "number",
		Required:    false,
		Min:         &minSalary,
		Max:         &maxSalary,
		Description: "Annual salary",
	})

	// Export to YAML
	schema := ValidationSchema{
		Name:        "EmployeeSchema",
		Description: "Employee data validation schema",
		Rules:       []*ValidationRule{},
	}
	
	for _, rule := range exportValidator.rules {
		schema.Rules = append(schema.Rules, rule)
	}
	
	yamlOutput, err := yaml.Marshal(&schema)
	if err != nil {
		fmt.Printf("Error marshaling to YAML: %v\n", err)
		return
	}
	
	fmt.Println("Exported YAML Schema:")
	fmt.Println("```yaml")
	fmt.Print(string(yamlOutput))
	fmt.Println("```")

	fmt.Println("\n=== Summary ===")
	fmt.Println("This validator supports:")
	fmt.Println("✓ Loading validation rules from YAML files")
	fmt.Println("✓ Loading validation rules from YAML strings")
	fmt.Println("✓ Adding rules programmatically")
	fmt.Println("✓ Mixing YAML and programmatic rule definitions")
	fmt.Println("✓ Exporting schemas to YAML format")
}