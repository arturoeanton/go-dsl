package main

import (
	"fmt"
	"path/filepath"

	"github.com/arturoeanton/go-dsl/examples/struct_validator/universal"
)

// Sample structs for testing
type User struct {
	Username  string `json:"username" yaml:"username"`
	Email     string `json:"email" yaml:"email"`
	Age       int    `json:"age" yaml:"age"`
	IsActive  bool   `json:"isActive" yaml:"isActive"`
	Profile   UserProfile `json:"profile" yaml:"profile"`
	Tags      []string `json:"tags" yaml:"tags"`
}

type UserProfile struct {
	FirstName string `json:"firstName" yaml:"firstName"`
	LastName  string `json:"lastName" yaml:"lastName"`
	Country   string `json:"country" yaml:"country"`
}

type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	Category    string  `json:"category"`
	Tags        []string `json:"tags"`
	Discount    *float64 `json:"discount,omitempty"`
}

func main() {
	fmt.Println("=== Go Struct & Map Validator with YAML Configuration ===")
	fmt.Println()

	// Example 1: Validate a Go struct
	fmt.Println("Example 1: Validating Go Struct (User) - Loading from external YAML")
	fmt.Println("--------------------------------------------------------------------")
	
	// Load schema from external YAML file
	userSchemaPath := filepath.Join("schemas", "user.yaml")
	
	// Create validator
	validator := universal.NewStructValidator()
	err := validator.LoadSchemaFromFile(userSchemaPath)
	if err != nil {
		fmt.Printf("Error loading schema: %v\n", err)
		fmt.Println("Make sure to run this from the struct_validator directory")
		return
	}
	
	fmt.Printf("Loaded schema from: %s\n", userSchemaPath)

	// Test valid user struct
	validUser := User{
		Username: "john_doe",
		Email:    "john@example.com",
		Age:      25,
		IsActive: true,
		Profile: UserProfile{
			FirstName: "John",
			LastName:  "Doe",
			Country:   "USA",
		},
		Tags: []string{"developer", "golang"},
	}
	
	fmt.Println("\nTesting valid user struct:")
	valid, errors := validator.ValidateStruct(validUser)
	if valid {
		fmt.Println("✓ Validation passed!")
	} else {
		fmt.Println("✗ Validation failed:")
		for _, err := range errors {
			fmt.Println("  -", err)
		}
	}

	// Test invalid user struct
	invalidUser := User{
		Username: "jo", // Too short
		Email:    "invalid-email",
		Age:      15, // Below minimum
		IsActive: true,
		Profile: UserProfile{
			FirstName: "",  // Empty required field
			LastName:  "Doe",
			Country:   "InvalidCountry",
		},
		Tags: []string{"ok", ""}, // Empty tag
	}
	
	fmt.Println("\nTesting invalid user struct:")
	valid, errors = validator.ValidateStruct(invalidUser)
	if valid {
		fmt.Println("✓ Validation passed!")
	} else {
		fmt.Println("✗ Validation failed:")
		for _, err := range errors {
			fmt.Println("  -", err)
		}
	}

	// Example 2: Validate a map[string]interface{}
	fmt.Println("\n\nExample 2: Validating map[string]interface{}")
	fmt.Println("---------------------------------------------")
	
	// Test valid map
	validMap := map[string]interface{}{
		"username": "jane_smith",
		"email":    "jane@example.com",
		"age":      30,
		"isActive": true,
		"profile": map[string]interface{}{
			"firstName": "Jane",
			"lastName":  "Smith",
			"country":   "Canada",
		},
		"tags": []interface{}{"manager", "team-lead"},
	}
	
	fmt.Println("\nTesting valid map:")
	valid, errors = validator.ValidateMap(validMap)
	if valid {
		fmt.Println("✓ Validation passed!")
	} else {
		fmt.Println("✗ Validation failed:")
		for _, err := range errors {
			fmt.Println("  -", err)
		}
	}

	// Test invalid map
	invalidMap := map[string]interface{}{
		"username": "a", // Too short
		"email":    "not-an-email",
		"age":      200, // Above maximum
		"isActive": "yes", // Wrong type
		"profile": map[string]interface{}{
			"firstName": "J",
			// lastName missing (required)
			"country": "Japan", // Not in enum
		},
	}
	
	fmt.Println("\nTesting invalid map:")
	valid, errors = validator.ValidateMap(invalidMap)
	if valid {
		fmt.Println("✓ Validation passed!")
	} else {
		fmt.Println("✗ Validation failed:")
		for _, err := range errors {
			fmt.Println("  -", err)
		}
	}

	// Example 3: Product validation
	fmt.Println("\n\nExample 3: Product Struct Validation - Loading from external YAML")
	fmt.Println("------------------------------------------------------------------")
	
	// Load product schema from external YAML file
	productSchemaPath := filepath.Join("schemas", "product.yaml")
	
	// Create product validator
	productValidator := universal.NewStructValidator()
	err = productValidator.LoadSchemaFromFile(productSchemaPath)
	if err != nil {
		fmt.Printf("Error loading product schema: %v\n", err)
		return
	}
	
	fmt.Printf("Loaded schema from: %s\n", productSchemaPath)

	// Test valid product
	discount := 15.0
	validProduct := Product{
		ID:       "PROD-12345",
		Name:     "Wireless Headphones",
		Price:    79.99,
		Stock:    150,
		Category: "Electronics",
		Tags:     []string{"wireless", "bluetooth", "audio"},
		Discount: &discount,
	}
	
	fmt.Println("\nTesting valid product:")
	valid, errors = productValidator.ValidateStruct(validProduct)
	if valid {
		fmt.Println("✓ Validation passed!")
	} else {
		fmt.Println("✗ Validation failed:")
		for _, err := range errors {
			fmt.Println("  -", err)
		}
	}

	// Test invalid product
	invalidDiscount := 150.0
	invalidProduct := Product{
		ID:       "12345", // Wrong pattern
		Name:     "", // Empty
		Price:    -10, // Negative
		Stock:    -5, // Negative
		Category: "InvalidCategory",
		Tags:     []string{""},
		Discount: &invalidDiscount, // Above max
	}
	
	fmt.Println("\nTesting invalid product:")
	valid, errors = productValidator.ValidateStruct(invalidProduct)
	if valid {
		fmt.Println("✓ Validation passed!")
	} else {
		fmt.Println("✗ Validation failed:")
		for _, err := range errors {
			fmt.Println("  -", err)
		}
	}

	// Example 4: Complex nested validation
	fmt.Println("\n\nExample 4: Complex Nested Validation - Loading from external YAML")
	fmt.Println("------------------------------------------------------------------")
	
	// Load order schema from external YAML file
	orderSchemaPath := filepath.Join("schemas", "order.yaml")
	
	// Create complex validator
	complexValidator := universal.NewStructValidator()
	err = complexValidator.LoadSchemaFromFile(orderSchemaPath)
	if err != nil {
		fmt.Printf("Error loading complex schema: %v\n", err)
		return
	}
	
	fmt.Printf("Loaded schema from: %s\n", orderSchemaPath)

	// Test valid complex map
	validOrder := map[string]interface{}{
		"orderId": "ORD-123456",
		"customer": map[string]interface{}{
			"id":    "CUST-001",
			"name":  "Alice Johnson",
			"email": "alice@example.com",
			"phone": "+14155552345",
			"address": map[string]interface{}{
				"street":  "123 Main St",
				"city":    "New York",
				"state":   "NY",
				"zipCode": "10001",
				"country": "US",
			},
		},
		"items": []interface{}{
			map[string]interface{}{
				"productId":   "PROD-0001",
				"productName": "Wireless Mouse",
				"quantity":    2,
				"unitPrice":   29.99,
				"discount":    10.0,
			},
			map[string]interface{}{
				"productId":   "PROD-0002",
				"productName": "USB Keyboard",
				"quantity":    1,
				"unitPrice":   49.99,
			},
		},
		"subtotal":      103.97,
		"tax":           8.32,
		"shipping":      5.99,
		"totalAmount":   118.28,
		"status":        "processing",
		"paymentMethod": "credit_card",
		"notes":         "Please deliver to the front desk",
		"createdAt":     "2024-01-15T10:30:00Z",
	}
	
	fmt.Println("\nTesting valid complex order:")
	valid, errors = complexValidator.ValidateMap(validOrder)
	if valid {
		fmt.Println("✓ Validation passed!")
	} else {
		fmt.Println("✗ Validation failed:")
		for _, err := range errors {
			fmt.Println("  -", err)
		}
	}

	// Test invalid complex map
	invalidOrder := map[string]interface{}{
		"orderId": "123456", // Wrong pattern
		"customer": map[string]interface{}{
			"id":    "001", // Wrong pattern
			"name":  "", // Empty required field
			"email": "invalid-email",
			"phone": "not-a-phone",
			"address": map[string]interface{}{
				"street":  "123 Main St",
				"city":    "New York",
				"zipCode": "ABC", // Invalid pattern
				"country": "USA", // Should be 2 chars
			},
		},
		"items": []interface{}{
			map[string]interface{}{
				"productId":   "001", // Wrong pattern
				"productName": "", // Empty
				"quantity":    0, // Below minimum
				"unitPrice":   -10, // Negative price
				"discount":    150, // Above max
			},
		},
		"subtotal":      -50, // Negative
		"tax":           -5,  // Negative
		"shipping":      -2,  // Negative
		"totalAmount":   -57, // Negative amount
		"status":        "unknown", // Not in enum
		"paymentMethod": "bitcoin", // Not in enum
		"createdAt":     "invalid-date",
	}
	
	fmt.Println("\nTesting invalid complex order:")
	valid, errors = complexValidator.ValidateMap(invalidOrder)
	if valid {
		fmt.Println("✓ Validation passed!")
	} else {
		fmt.Println("✗ Validation failed:")
		for _, err := range errors {
			fmt.Println("  -", err)
		}
	}

	// Example 5: API Configuration validation
	fmt.Println("\n\nExample 5: API Configuration Validation - Loading from external YAML")
	fmt.Println("--------------------------------------------------------------------")
	
	// Load API config schema from external YAML file
	apiSchemaPath := filepath.Join("schemas", "api_config.yaml")
	
	// Create API validator
	apiValidator := universal.NewStructValidator()
	err = apiValidator.LoadSchemaFromFile(apiSchemaPath)
	if err != nil {
		fmt.Printf("Error loading API schema: %v\n", err)
		return
	}
	
	fmt.Printf("Loaded schema from: %s\n", apiSchemaPath)

	// Test valid API configuration
	validAPIConfig := map[string]interface{}{
		"endpoint": "https://api.example.com/v1/data",
		"apiKey":   "sk_test_4eC39HqLyjWDarjtT1zdp7dc_4eC39HqLyjWDarjtT1zdp7dc",
		"timeout":  30,
		"retries":  3,
		"retryDelay": 1000,
		"methods":  []interface{}{"GET", "POST", "PUT"},
		"headers": map[string]interface{}{
			"Content-Type": "application/json",
			"User-Agent":   "MyApp/1.0",
		},
		"rateLimit": map[string]interface{}{
			"requests": 100,
			"window":   60,
			"burst":    10,
		},
		"logging": map[string]interface{}{
			"level":  "info",
			"format": "json",
			"output": "stdout",
		},
	}
	
	fmt.Println("\nTesting valid API configuration:")
	valid, errors = apiValidator.ValidateMap(validAPIConfig)
	if valid {
		fmt.Println("✓ Validation passed!")
	} else {
		fmt.Println("✗ Validation failed:")
		for _, err := range errors {
			fmt.Println("  -", err)
		}
	}

	// Example 6: Programmatic schema creation
	fmt.Println("\n\nExample 6: Programmatic Schema Creation")
	fmt.Println("---------------------------------------")
	
	// Create a validator with programmatic rules
	programmaticValidator := universal.NewStructValidator()
	
	// Load inline YAML
	inlineSchema := `
name: SimpleUser
description: Simple user validation
rules:
  - field: name
    type: string
    required: true
    minLength: 1
    maxLength: 100
  - field: email
    type: string
    required: true
    pattern: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
  - field: age
    type: integer
    required: false
    min: 0
    max: 150`
	
	err = programmaticValidator.LoadSchemaFromYAML(inlineSchema)
	if err != nil {
		fmt.Printf("Error loading inline schema: %v\n", err)
		return
	}
	
	// Test with map
	testData := map[string]interface{}{
		"name":  "Bob Smith",
		"email": "bob@example.com",
		"age":   35,
	}
	
	fmt.Println("\nTesting programmatic schema:")
	valid, errors = programmaticValidator.ValidateMap(testData)
	if valid {
		fmt.Println("✓ Validation passed!")
	} else {
		fmt.Println("✗ Validation failed:")
		for _, err := range errors {
			fmt.Println("  -", err)
		}
	}

	fmt.Println("\n=== Summary ===")
	fmt.Println("This validator supports:")
	fmt.Println("✓ Validating Go structs with reflection")
	fmt.Println("✓ Validating map[string]interface{} structures")
	fmt.Println("✓ Loading validation rules from external YAML files")
	fmt.Println("✓ Loading validation rules from inline YAML strings")
	fmt.Println("✓ Nested object validation (multiple levels)")
	fmt.Println("✓ Array item validation with custom rules")
	fmt.Println("✓ Type checking (string, number, integer, boolean, array, object)")
	fmt.Println("✓ String constraints (minLength, maxLength, pattern, enum)")
	fmt.Println("✓ Number constraints (min, max)")
	fmt.Println("✓ Enum validation for restricted value sets")
	fmt.Println("✓ Required and optional field validation")
	fmt.Println("✓ Support for JSON and YAML struct tags")
	fmt.Println("\nThe validation logic is in the 'universal' package for easy reuse")
	fmt.Println("Schema files are located in the 'schemas/' directory")
}