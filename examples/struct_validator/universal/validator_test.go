package universal

import (
	"testing"
)

func TestStructValidation(t *testing.T) {
	tests := []struct {
		name      string
		yaml      string
		data      interface{}
		wantValid bool
		wantErrors []string
	}{
		{
			name: "valid user struct",
			yaml: `
name: UserTest
rules:
  - field: username
    type: string
    required: true
    minLength: 3
    maxLength: 20
  - field: age
    type: integer
    required: true
    min: 18
    max: 120`,
			data: struct {
				Username string
				Age      int
			}{
				Username: "john_doe",
				Age:      25,
			},
			wantValid: true,
			wantErrors: []string{},
		},
		{
			name: "invalid username too short",
			yaml: `
name: UserTest
rules:
  - field: username
    type: string
    required: true
    minLength: 3`,
			data: struct {
				Username string
			}{
				Username: "ab",
			},
			wantValid: false,
			wantErrors: []string{"Field 'username' must have at least 3 characters"},
		},
		{
			name: "missing required field",
			yaml: `
name: UserTest
rules:
  - field: email
    type: string
    required: true`,
			data: struct {
				Username string
			}{
				Username: "john",
			},
			wantValid: false,
			wantErrors: []string{"Field 'email' is required"},
		},
		{
			name: "pattern validation",
			yaml: `
name: EmailTest
rules:
  - field: email
    type: string
    required: true
    pattern: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"`,
			data: struct {
				Email string
			}{
				Email: "invalid-email",
			},
			wantValid: false,
			wantErrors: []string{"Field 'email' must match pattern ^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"},
		},
		{
			name: "enum validation",
			yaml: `
name: StatusTest
rules:
  - field: status
    type: string
    required: true
    enum: ["active", "inactive", "pending"]`,
			data: struct {
				Status string
			}{
				Status: "unknown",
			},
			wantValid: false,
			wantErrors: []string{"Field 'status' must be one of [active inactive pending]"},
		},
		{
			name: "nested struct validation",
			yaml: `
name: UserTest
rules:
  - field: profile
    type: object
    required: true
    nested:
      name: ProfileTest
      rules:
        - field: firstName
          type: string
          required: true
          minLength: 1`,
			data: struct {
				Profile struct {
					FirstName string
				}
			}{
				Profile: struct {
					FirstName string
				}{
					FirstName: "John",
				},
			},
			wantValid: true,
			wantErrors: []string{},
		},
		{
			name: "array validation",
			yaml: `
name: TagsTest
rules:
  - field: tags
    type: array
    required: true
    items:
      field: tag
      type: string
      minLength: 2`,
			data: struct {
				Tags []string
			}{
				Tags: []string{"go", "test"},
			},
			wantValid: true,
			wantErrors: []string{},
		},
		{
			name: "number range validation",
			yaml: `
name: PriceTest
rules:
  - field: price
    type: number
    required: true
    min: 0.01
    max: 999.99`,
			data: struct {
				Price float64
			}{
				Price: 50.25,
			},
			wantValid: true,
			wantErrors: []string{},
		},
		{
			name: "boolean validation",
			yaml: `
name: ActiveTest
rules:
  - field: isActive
    type: boolean
    required: true`,
			data: struct {
				IsActive bool
			}{
				IsActive: true,
			},
			wantValid: true,
			wantErrors: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewStructValidator()
			err := validator.LoadSchemaFromYAML(tt.yaml)
			if err != nil {
				t.Fatalf("Failed to load schema: %v", err)
			}

			valid, errors := validator.ValidateStruct(tt.data)
			
			if valid != tt.wantValid {
				t.Errorf("ValidateStruct() valid = %v, want %v", valid, tt.wantValid)
			}
			
			if len(errors) != len(tt.wantErrors) {
				t.Errorf("ValidateStruct() got %d errors, want %d errors. Got: %v", 
					len(errors), len(tt.wantErrors), errors)
				return
			}
			
			for i, wantError := range tt.wantErrors {
				if i >= len(errors) || errors[i] != wantError {
					t.Errorf("ValidateStruct() error[%d] = %v, want %v", 
						i, errors[i], wantError)
				}
			}
		})
	}
}

func TestMapValidation(t *testing.T) {
	tests := []struct {
		name       string
		yaml       string
		data       map[string]interface{}
		wantValid  bool
		wantErrors []string
	}{
		{
			name: "valid map",
			yaml: `
name: MapTest
rules:
  - field: name
    type: string
    required: true
    minLength: 1
  - field: age
    type: integer
    required: true
    min: 0`,
			data: map[string]interface{}{
				"name": "Alice",
				"age":  30,
			},
			wantValid:  true,
			wantErrors: []string{},
		},
		{
			name: "invalid type in map",
			yaml: `
name: MapTest
rules:
  - field: age
    type: integer
    required: true`,
			data: map[string]interface{}{
				"age": "not a number",
			},
			wantValid:  false,
			wantErrors: []string{"Field 'age' must be a number"},
		},
		{
			name: "nested map validation",
			yaml: `
name: NestedMapTest
rules:
  - field: address
    type: object
    required: true
    nested:
      name: AddressTest
      rules:
        - field: city
          type: string
          required: true
        - field: zipCode
          type: string
          required: true
          pattern: "^[0-9]{5}$"`,
			data: map[string]interface{}{
				"address": map[string]interface{}{
					"city":    "New York",
					"zipCode": "10001",
				},
			},
			wantValid:  true,
			wantErrors: []string{},
		},
		{
			name: "array in map validation",
			yaml: `
name: ArrayMapTest
rules:
  - field: items
    type: array
    required: true
    items:
      field: item
      type: object
      nested:
        name: ItemTest
        rules:
          - field: id
            type: string
            required: true
          - field: quantity
            type: integer
            required: true
            min: 1`,
			data: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{
						"id":       "ITEM-001",
						"quantity": 5,
					},
					map[string]interface{}{
						"id":       "ITEM-002",
						"quantity": 3,
					},
				},
			},
			wantValid:  true,
			wantErrors: []string{},
		},
		{
			name: "optional field validation",
			yaml: `
name: OptionalTest
rules:
  - field: required_field
    type: string
    required: true
  - field: optional_field
    type: string
    required: false
    minLength: 5`,
			data: map[string]interface{}{
				"required_field": "value",
				// optional_field is missing, which should be OK
			},
			wantValid:  true,
			wantErrors: []string{},
		},
		{
			name: "optional field with invalid value",
			yaml: `
name: OptionalTest
rules:
  - field: optional_field
    type: string
    required: false
    minLength: 5`,
			data: map[string]interface{}{
				"optional_field": "abc", // Too short
			},
			wantValid:  false,
			wantErrors: []string{"Field 'optional_field' must have at least 5 characters"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewStructValidator()
			err := validator.LoadSchemaFromYAML(tt.yaml)
			if err != nil {
				t.Fatalf("Failed to load schema: %v", err)
			}

			valid, errors := validator.ValidateMap(tt.data)
			
			if valid != tt.wantValid {
				t.Errorf("ValidateMap() valid = %v, want %v", valid, tt.wantValid)
			}
			
			if len(errors) != len(tt.wantErrors) {
				t.Errorf("ValidateMap() got %d errors, want %d errors. Got: %v", 
					len(errors), len(tt.wantErrors), errors)
				return
			}
			
			for i, wantError := range tt.wantErrors {
				if i >= len(errors) || errors[i] != wantError {
					t.Errorf("ValidateMap() error[%d] = %v, want %v", 
						i, errors[i], wantError)
				}
			}
		})
	}
}

func TestFieldTagRecognition(t *testing.T) {
	yaml := `
name: TagTest
rules:
  - field: user_name
    type: string
    required: true`

	validator := NewStructValidator()
	err := validator.LoadSchemaFromYAML(yaml)
	if err != nil {
		t.Fatalf("Failed to load schema: %v", err)
	}

	// Test with json tag
	jsonStruct := struct {
		Name string `json:"user_name"`
	}{
		Name: "Alice",
	}

	valid, _ := validator.ValidateStruct(jsonStruct)
	if !valid {
		t.Error("Expected struct with json tag to be valid")
	}

	// Test with yaml tag
	yamlStruct := struct {
		Name string `yaml:"user_name"`
	}{
		Name: "Bob",
	}

	valid, _ = validator.ValidateStruct(yamlStruct)
	if !valid {
		t.Error("Expected struct with yaml tag to be valid")
	}

	// Test with field name match (case insensitive)
	fieldStruct := struct {
		User_Name string
	}{
		User_Name: "Charlie",
	}

	valid, _ = validator.ValidateStruct(fieldStruct)
	if !valid {
		t.Error("Expected struct with matching field name to be valid")
	}
}

func TestPointerFields(t *testing.T) {
	yaml := `
name: PointerTest
rules:
  - field: optional
    type: string
    required: false
    minLength: 3
  - field: required
    type: string
    required: true`

	validator := NewStructValidator()
	err := validator.LoadSchemaFromYAML(yaml)
	if err != nil {
		t.Fatalf("Failed to load schema: %v", err)
	}

	// Test with nil pointer (optional field)
	type TestStruct struct {
		Optional *string
		Required string
	}

	validStruct := TestStruct{
		Optional: nil,
		Required: "value",
	}

	valid, errors := validator.ValidateStruct(validStruct)
	if !valid {
		t.Errorf("Expected struct with nil optional pointer to be valid. Errors: %v", errors)
	}

	// Test with valid pointer value
	optionalValue := "valid"
	validStruct2 := TestStruct{
		Optional: &optionalValue,
		Required: "value",
	}

	valid, errors = validator.ValidateStruct(validStruct2)
	if !valid {
		t.Errorf("Expected struct with valid pointer to be valid. Errors: %v", errors)
	}

	// Test with invalid pointer value
	invalidValue := "ab" // Too short
	invalidStruct := TestStruct{
		Optional: &invalidValue,
		Required: "value",
	}

	valid, errors = validator.ValidateStruct(invalidStruct)
	if valid {
		t.Error("Expected struct with invalid pointer value to fail validation")
	}
	if len(errors) != 1 || errors[0] != "Field 'optional' must have at least 3 characters" {
		t.Errorf("Unexpected errors: %v", errors)
	}
}

func BenchmarkStructValidation(b *testing.B) {
	yaml := `
name: BenchTest
rules:
  - field: name
    type: string
    required: true
    minLength: 1
    maxLength: 100
  - field: age
    type: integer
    required: true
    min: 0
    max: 150
  - field: email
    type: string
    required: true
    pattern: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"`

	validator := NewStructValidator()
	validator.LoadSchemaFromYAML(yaml)

	testStruct := struct {
		Name  string
		Age   int
		Email string
	}{
		Name:  "John Doe",
		Age:   30,
		Email: "john@example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidateStruct(testStruct)
	}
}

func BenchmarkMapValidation(b *testing.B) {
	yaml := `
name: BenchTest
rules:
  - field: name
    type: string
    required: true
    minLength: 1
    maxLength: 100
  - field: age
    type: integer
    required: true
    min: 0
    max: 150
  - field: email
    type: string
    required: true
    pattern: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"`

	validator := NewStructValidator()
	validator.LoadSchemaFromYAML(yaml)

	testMap := map[string]interface{}{
		"name":  "John Doe",
		"age":   30,
		"email": "john@example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidateMap(testMap)
	}
}