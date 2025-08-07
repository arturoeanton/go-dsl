# JSON Validator / Validador JSON

[English](#english) | [Español](#español)

---

## English

### Generic JSON Data Validator

A flexible and extensible JSON validator that allows you to define validation rules programmatically and apply them to any JSON data structure.

### Features

- **Multiple data types support**: string, number, integer, boolean, array
- **Comprehensive validation rules**:
  - Required/optional fields
  - Min/max values for numbers
  - Min/max length for strings
  - Regular expression patterns
  - Enumeration values
  - Custom error messages
- **Easy to extend** for additional validation requirements

### How to Use

#### 1. Run the Example

```bash
go run working_validator.go
```

#### 2. Create a Validator

```go
validator := NewValidator()
```

#### 3. Define Validation Rules

```go
// String field with pattern validation
validator.AddRule(&ValidationRule{
    FieldName:   "email",
    Type:        "string",
    Required:    true,
    Pattern:     "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
    Description: "User email address",
})

// Integer field with range validation
minAge := 18.0
maxAge := 120.0
validator.AddRule(&ValidationRule{
    FieldName:   "age",
    Type:        "integer",
    Required:    true,
    Min:         &minAge,
    Max:         &maxAge,
    Description: "User age",
})

// String field with enumeration
validator.AddRule(&ValidationRule{
    FieldName:   "country",
    Type:        "string",
    Required:    false,
    Enum:        []interface{}{"USA", "Canada", "Mexico", "UK"},
    Description: "Country of residence",
})
```

#### 4. Validate JSON Data

```go
jsonData := `{
    "email": "user@example.com",
    "age": 25,
    "country": "USA"
}`

valid, errors := validator.Validate(jsonData)
if valid {
    fmt.Println("✓ Validation passed!")
} else {
    fmt.Println("✗ Validation failed:")
    for _, err := range errors {
        fmt.Println("  -", err)
    }
}
```

### Validation Rule Options

| Field | Type | Description |
|-------|------|-------------|
| `FieldName` | string | Name of the JSON field to validate |
| `Type` | string | Data type: "string", "number", "integer", "boolean", "array" |
| `Required` | bool | Whether the field is required |
| `Min` | *float64 | Minimum value for numbers |
| `Max` | *float64 | Maximum value for numbers |
| `MinLength` | *int | Minimum length for strings |
| `MaxLength` | *int | Maximum length for strings |
| `Pattern` | string | Regular expression pattern for strings |
| `Enum` | []interface{} | List of allowed values |
| `Description` | string | Field description |

### Example Use Cases

The example includes three practical validation scenarios:

1. **User Registration**: Validates username, email, age, password, and country
2. **Product Catalog**: Validates product ID, name, price, stock, category, and discount
3. **API Configuration**: Validates endpoint URL, timeout, retries, HTTP methods, and rate limit

### Extending the Validator

You can easily extend the validator to support:
- Nested object validation
- Custom validation functions
- Cross-field validation
- Async validation
- Database constraint validation

---

## Español

### Validador Genérico de Datos JSON

Un validador JSON flexible y extensible que permite definir reglas de validación programáticamente y aplicarlas a cualquier estructura de datos JSON.

### Características

- **Soporte para múltiples tipos de datos**: string, number, integer, boolean, array
- **Reglas de validación completas**:
  - Campos requeridos/opcionales
  - Valores mínimos/máximos para números
  - Longitud mínima/máxima para cadenas
  - Patrones de expresiones regulares
  - Valores de enumeración
  - Mensajes de error personalizados
- **Fácil de extender** para requisitos de validación adicionales

### Cómo Usar

#### 1. Ejecutar el Ejemplo

```bash
go run working_validator.go
```

#### 2. Crear un Validador

```go
validator := NewValidator()
```

#### 3. Definir Reglas de Validación

```go
// Campo string con validación de patrón
validator.AddRule(&ValidationRule{
    FieldName:   "email",
    Type:        "string",
    Required:    true,
    Pattern:     "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
    Description: "Dirección de email del usuario",
})

// Campo entero con validación de rango
minAge := 18.0
maxAge := 120.0
validator.AddRule(&ValidationRule{
    FieldName:   "age",
    Type:        "integer",
    Required:    true,
    Min:         &minAge,
    Max:         &maxAge,
    Description: "Edad del usuario",
})

// Campo string con enumeración
validator.AddRule(&ValidationRule{
    FieldName:   "country",
    Type:        "string",
    Required:    false,
    Enum:        []interface{}{"USA", "Canada", "Mexico", "UK"},
    Description: "País de residencia",
})
```

#### 4. Validar Datos JSON

```go
jsonData := `{
    "email": "usuario@ejemplo.com",
    "age": 25,
    "country": "Mexico"
}`

valid, errors := validator.Validate(jsonData)
if valid {
    fmt.Println("✓ ¡Validación exitosa!")
} else {
    fmt.Println("✗ Validación fallida:")
    for _, err := range errors {
        fmt.Println("  -", err)
    }
}
```

### Opciones de Reglas de Validación

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `FieldName` | string | Nombre del campo JSON a validar |
| `Type` | string | Tipo de dato: "string", "number", "integer", "boolean", "array" |
| `Required` | bool | Si el campo es requerido |
| `Min` | *float64 | Valor mínimo para números |
| `Max` | *float64 | Valor máximo para números |
| `MinLength` | *int | Longitud mínima para cadenas |
| `MaxLength` | *int | Longitud máxima para cadenas |
| `Pattern` | string | Patrón de expresión regular para cadenas |
| `Enum` | []interface{} | Lista de valores permitidos |
| `Description` | string | Descripción del campo |

### Casos de Uso de Ejemplo

El ejemplo incluye tres escenarios prácticos de validación:

1. **Registro de Usuario**: Valida nombre de usuario, email, edad, contraseña y país
2. **Catálogo de Productos**: Valida ID de producto, nombre, precio, stock, categoría y descuento
3. **Configuración de API**: Valida URL del endpoint, timeout, reintentos, métodos HTTP y límite de tasa

### Extender el Validador

Puedes extender fácilmente el validador para soportar:
- Validación de objetos anidados
- Funciones de validación personalizadas
- Validación entre campos
- Validación asíncrona
- Validación de restricciones de base de datos

---

## License / Licencia

This example is part of the go-dsl project. See the main project LICENSE file for details.

Este ejemplo es parte del proyecto go-dsl. Consulte el archivo LICENSE del proyecto principal para más detalles.