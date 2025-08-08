# Go Struct & Map Validator

A flexible and powerful validation library for Go structs and `map[string]interface{}` using YAML-based schema definitions.

## Features

- ✅ **Dual validation support**: Validate both Go structs and `map[string]interface{}`
- ✅ **YAML-based schemas**: Define validation rules in readable YAML format
- ✅ **Reflection-based struct validation**: Automatically validates struct fields using reflection
- ✅ **Nested object validation**: Support for complex nested structures
- ✅ **Array/slice validation**: Validate array elements with custom rules
- ✅ **Multiple validation types**:
  - String (minLength, maxLength, pattern, enum)
  - Number/Integer (min, max)
  - Boolean
  - Array/Slice (with item validation)
  - Object/Struct (with nested validation)
- ✅ **Field tag support**: Recognizes `json` and `yaml` struct tags
- ✅ **Optional and required fields**: Flexible field requirement configuration
- ✅ **Pattern matching**: Regex-based string validation
- ✅ **Enum validation**: Restrict values to predefined sets

## Installation

```bash
go get github.com/arturoeanton/go-dsl
```

## Quick Start

### 1. Define Your Validation Schema (YAML)

```yaml
name: UserValidation
description: Schema for user validation
rules:
  - field: username
    type: string
    required: true
    minLength: 3
    maxLength: 20
    pattern: "^[a-zA-Z0-9_]+$"
    
  - field: email
    type: string
    required: true
    pattern: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
    
  - field: age
    type: integer
    required: true
    min: 18
    max: 120
```

### 2. Validate a Go Struct

```go
package main

import (
    "fmt"
    "io/ioutil"
)

type User struct {
    Username string `json:"username"`
    Email    string `json:"email"`
    Age      int    `json:"age"`
}

func main() {
    // Create validator
    validator := NewStructValidator()
    
    // Load schema from file
    schemaContent, _ := ioutil.ReadFile("schemas/user.yaml")
    validator.LoadSchemaFromYAML(string(schemaContent))
    
    // Validate struct
    user := User{
        Username: "john_doe",
        Email:    "john@example.com",
        Age:      25,
    }
    
    valid, errors := validator.ValidateStruct(user)
    if valid {
        fmt.Println("✓ Validation passed!")
    } else {
        fmt.Println("✗ Validation failed:")
        for _, err := range errors {
            fmt.Println("  -", err)
        }
    }
}
```

### 3. Validate a Map

```go
// Validate map[string]interface{}
data := map[string]interface{}{
    "username": "jane_smith",
    "email":    "jane@example.com",
    "age":      30,
}

valid, errors := validator.ValidateMap(data)
```

## Schema Definition

### Basic Field Types

| Type | Description | Constraints |
|------|-------------|------------|
| `string` | String values | minLength, maxLength, pattern, enum |
| `integer` | Integer numbers | min, max |
| `number` | Floating point numbers | min, max |
| `boolean` | Boolean values | - |
| `array` | Arrays/Slices | items (for element validation) |
| `object` | Objects/Structs | nested (for nested validation) |

### Field Properties

```yaml
- field: fieldName        # Field name (required)
  type: string           # Field type (required)
  required: true         # Is field required? (default: false)
  description: "..."     # Field description (optional)
  
  # String constraints
  minLength: 3          # Minimum string length
  maxLength: 50         # Maximum string length
  pattern: "^[A-Z]+$"   # Regex pattern
  enum: ["A", "B", "C"] # Allowed values
  
  # Number constraints  
  min: 0                # Minimum value
  max: 100              # Maximum value
  
  # Array validation
  items:                # Validation for array elements
    field: item
    type: string
    minLength: 1
  
  # Nested object validation
  nested:               # Validation for nested objects
    name: NestedSchema
    rules:
      - field: nestedField
        type: string
        required: true
```

## Advanced Examples

### Nested Object Validation

```yaml
- field: address
  type: object
  required: true
  nested:
    name: Address
    rules:
      - field: street
        type: string
        required: true
      - field: city
        type: string
        required: true
      - field: zipCode
        type: string
        pattern: "^[0-9]{5}$"
```

### Array with Object Elements

```yaml
- field: items
  type: array
  required: true
  items:
    field: item
    type: object
    nested:
      name: OrderItem
      rules:
        - field: productId
          type: string
          required: true
        - field: quantity
          type: integer
          min: 1
```

## Available Schema Files

The `schemas/` directory contains several pre-built validation schemas:

- **user.yaml**: User registration and profile validation
- **product.yaml**: Product catalog validation
- **order.yaml**: Complex order validation with nested structures
- **api_config.yaml**: API configuration validation

## Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -run TestStructValidation

# Run benchmarks
go test -bench=.
```

## Performance

The validator is optimized for performance with:
- Efficient reflection usage
- Minimal allocations
- Compiled regex patterns (cached)

Benchmark results (example):
```
BenchmarkStructValidation-8    100000    10234 ns/op
BenchmarkMapValidation-8       100000    11456 ns/op
```

## Error Messages

The validator provides detailed error messages:

```
Field 'username' must have at least 3 characters
Field 'email' must match pattern ^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$
Field 'age' must be >= 18
profile.Field 'country' must be one of [USA Canada Mexico UK Germany France]
items[0].Field 'quantity' must be >= 1
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is part of the go-dsl framework and follows its licensing terms.