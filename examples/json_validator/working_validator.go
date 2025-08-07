package main

import (
	"encoding/json"
	"fmt"
	"regexp"
)

// ValidationRule represents a validation rule for a field
type ValidationRule struct {
	FieldName   string
	Type        string
	Required    bool
	Min         *float64
	Max         *float64
	MinLength   *int
	MaxLength   *int
	Pattern     string
	Enum        []interface{}
	Description string
}

// Validator validates JSON data against rules
type Validator struct {
	rules map[string]*ValidationRule
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		rules: make(map[string]*ValidationRule),
	}
}

// AddRule adds a validation rule
func (v *Validator) AddRule(rule *ValidationRule) {
	v.rules[rule.FieldName] = rule
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
	fmt.Println("=== Generic JSON Data Validator ===\n")
	
	// Example 1: User Registration
	fmt.Println("Example 1: User Registration Validation")
	fmt.Println("----------------------------------------")
	
	userValidator := NewValidator()
	
	// Define validation rules
	minAge := 18.0
	maxAge := 120.0
	minPassLen := 8
	maxPassLen := 100
	
	userValidator.AddRule(&ValidationRule{
		FieldName:   "username",
		Type:        "string",
		Required:    true,
		MinLength:   intPtr(3),
		MaxLength:   intPtr(20),
		Pattern:     "^[a-zA-Z0-9_]+$",
		Description: "Username for the account",
	})
	
	userValidator.AddRule(&ValidationRule{
		FieldName:   "email",
		Type:        "string",
		Required:    true,
		Pattern:     "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
		Description: "User email address",
	})
	
	userValidator.AddRule(&ValidationRule{
		FieldName:   "age",
		Type:        "integer",
		Required:    true,
		Min:         &minAge,
		Max:         &maxAge,
		Description: "User age",
	})
	
	userValidator.AddRule(&ValidationRule{
		FieldName:   "password",
		Type:        "string",
		Required:    true,
		MinLength:   &minPassLen,
		MaxLength:   &maxPassLen,
		Description: "Account password",
	})
	
	userValidator.AddRule(&ValidationRule{
		FieldName:   "country",
		Type:        "string",
		Required:    false,
		Enum:        []interface{}{"USA", "Canada", "Mexico", "UK", "Germany", "France"},
		Description: "Country of residence",
	})
	
	// Test valid user
	validUser := `{
		"username": "john_doe",
		"email": "john@example.com",
		"age": 25,
		"password": "SecurePass123",
		"country": "USA"
	}`
	
	fmt.Println("Testing valid user:")
	fmt.Println(validUser)
	valid, errors := userValidator.Validate(validUser)
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
	fmt.Println(invalidUser)
	valid, errors = userValidator.Validate(invalidUser)
	if valid {
		fmt.Println("✓ Validation passed!\n")
	} else {
		fmt.Println("✗ Validation failed:")
		for _, err := range errors {
			fmt.Println("  -", err)
		}
		fmt.Println()
	}
	
	// Example 2: Product Catalog
	fmt.Println("\nExample 2: Product Catalog Validation")
	fmt.Println("--------------------------------------")
	
	productValidator := NewValidator()
	
	minPrice := 0.01
	maxPrice := 999999.99
	minStock := 0.0
	maxDiscount := 100.0
	minDiscount := 0.0
	
	productValidator.AddRule(&ValidationRule{
		FieldName:   "id",
		Type:        "string",
		Required:    true,
		Pattern:     "^PROD-[0-9]{4,}$",
		Description: "Product ID",
	})
	
	productValidator.AddRule(&ValidationRule{
		FieldName:   "name",
		Type:        "string",
		Required:    true,
		MinLength:   intPtr(1),
		MaxLength:   intPtr(100),
		Description: "Product name",
	})
	
	productValidator.AddRule(&ValidationRule{
		FieldName:   "price",
		Type:        "number",
		Required:    true,
		Min:         &minPrice,
		Max:         &maxPrice,
		Description: "Product price",
	})
	
	productValidator.AddRule(&ValidationRule{
		FieldName:   "stock",
		Type:        "integer",
		Required:    true,
		Min:         &minStock,
		Description: "Stock quantity",
	})
	
	productValidator.AddRule(&ValidationRule{
		FieldName:   "category",
		Type:        "string",
		Required:    true,
		Enum:        []interface{}{"Electronics", "Clothing", "Food", "Books", "Sports"},
		Description: "Product category",
	})
	
	productValidator.AddRule(&ValidationRule{
		FieldName:   "discount",
		Type:        "number",
		Required:    false,
		Min:         &minDiscount,
		Max:         &maxDiscount,
		Description: "Discount percentage",
	})
	
	// Test valid product
	validProduct := `{
		"id": "PROD-12345",
		"name": "Wireless Headphones",
		"price": 79.99,
		"stock": 150,
		"category": "Electronics",
		"discount": 10
	}`
	
	fmt.Println("Testing valid product:")
	fmt.Println(validProduct)
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
	fmt.Println(invalidProduct)
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
	
	// Example 3: API Configuration
	fmt.Println("\nExample 3: API Configuration Validation")
	fmt.Println("----------------------------------------")
	
	apiValidator := NewValidator()
	
	minTimeout := 1.0
	maxTimeout := 300.0
	minRetries := 0.0
	maxRetries := 10.0
	minRateLimit := 1.0
	maxRateLimit := 10000.0
	
	apiValidator.AddRule(&ValidationRule{
		FieldName:   "endpoint",
		Type:        "string",
		Required:    true,
		Pattern:     "^https?://.*",
		Description: "API endpoint URL",
	})
	
	apiValidator.AddRule(&ValidationRule{
		FieldName:   "timeout",
		Type:        "integer",
		Required:    false,
		Min:         &minTimeout,
		Max:         &maxTimeout,
		Description: "Request timeout in seconds",
	})
	
	apiValidator.AddRule(&ValidationRule{
		FieldName:   "retries",
		Type:        "integer",
		Required:    false,
		Min:         &minRetries,
		Max:         &maxRetries,
		Description: "Number of retry attempts",
	})
	
	apiValidator.AddRule(&ValidationRule{
		FieldName:   "methods",
		Type:        "array",
		Required:    true,
		Description: "Allowed HTTP methods",
	})
	
	apiValidator.AddRule(&ValidationRule{
		FieldName:   "rateLimit",
		Type:        "integer",
		Required:    false,
		Min:         &minRateLimit,
		Max:         &maxRateLimit,
		Description: "Requests per minute",
	})
	
	// Test valid API config
	validAPI := `{
		"endpoint": "https://api.example.com/v1",
		"timeout": 60,
		"retries": 5,
		"methods": ["GET", "POST", "PUT"],
		"rateLimit": 500
	}`
	
	fmt.Println("Testing valid API configuration:")
	fmt.Println(validAPI)
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
	
	// Test invalid API config
	invalidAPI := `{
		"endpoint": "invalid-url",
		"timeout": 500,
		"methods": []
	}`
	
	fmt.Println("Testing invalid API configuration:")
	fmt.Println(invalidAPI)
	valid, errors = apiValidator.Validate(invalidAPI)
	if valid {
		fmt.Println("✓ Validation passed!\n")
	} else {
		fmt.Println("✗ Validation failed:")
		for _, err := range errors {
			fmt.Println("  -", err)
		}
		fmt.Println()
	}
	
	// Demonstrate flexibility
	fmt.Println("\n=== Demonstrating Flexibility ===")
	fmt.Println("This validator can be easily extended to handle:")
	fmt.Println("- Any JSON structure")
	fmt.Println("- Custom validation rules")
	fmt.Println("- Nested objects (with modifications)")
	fmt.Println("- Custom error messages")
	fmt.Println("- Multiple validation strategies")
	fmt.Println("\nThe validator is generic and can be used for any JSON validation needs!")
}

func intPtr(i int) *int {
	return &i
}

func floatPtr(f float64) *float64 {
	return &f
}