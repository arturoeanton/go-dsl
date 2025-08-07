# JSON Validator with YAML Configuration / Validador JSON con Configuración YAML

[English](#english) | [Español](#español)

---

## English

### JSON Validator with YAML Schema Configuration

An enhanced version of the JSON validator that allows you to define validation rules using YAML files or strings, making schema management more flexible and maintainable.

### Features

- **YAML-based schema definition**: Define validation rules in easy-to-read YAML format
- **Multiple loading options**:
  - Load from YAML files
  - Load from YAML strings
  - Programmatic rule definition
  - Mixed approach (YAML + programmatic)
- **Schema export**: Export programmatically created schemas to YAML
- **All validation features from base validator**:
  - Multiple data types (string, number, integer, boolean, array, object)
  - Required/optional fields
  - Min/max constraints
  - Pattern matching
  - Enumeration values
  - Length constraints

### Installation

```bash
# Install dependencies
go mod init myproject
go get gopkg.in/yaml.v3

# Or use the provided go.mod
cd examples/json_validator_yaml
go mod tidy
```

### How to Use

#### 1. Run the Example

```bash
go run validator_with_yaml.go
```

#### 2. Create a YAML Schema File

```yaml
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
  
  - field: country
    type: string
    required: false
    enum: ["USA", "Canada", "Mexico", "UK"]
    description: Country of residence
```

#### 3. Load Schema from YAML File

```go
validator := NewValidator()
err := validator.LoadSchemaFromYAML("schemas/user_registration.yaml")
if err != nil {
    log.Fatal(err)
}

// Get schema information
name, description := validator.GetSchemaInfo()
fmt.Printf("Schema: %s - %s\n", name, description)
```

#### 4. Load Schema from YAML String

```go
yamlContent := `
name: SimpleSchema
description: A simple validation schema
rules:
  - field: name
    type: string
    required: true
    maxLength: 50
`

validator := NewValidator()
err := validator.LoadSchemaFromYAMLString(yamlContent)
if err != nil {
    log.Fatal(err)
}
```

#### 5. Mixed Approach (YAML + Programmatic)

```go
// Load base schema from YAML
validator := NewValidator()
err := validator.LoadSchemaFromYAML("base_schema.yaml")

// Add additional rules programmatically
minValue := 0.0
maxValue := 100.0
validator.AddRule(&ValidationRule{
    FieldName:   "score",
    Type:        "number",
    Required:    true,
    Min:         &minValue,
    Max:         &maxValue,
    Description: "Score value",
})
```

#### 6. Validate JSON Data

```go
jsonData := `{
    "username": "john_doe",
    "email": "john@example.com",
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

### YAML Schema Structure

```yaml
name: SchemaName                    # Schema identifier
description: Schema description      # Human-readable description
rules:                              # List of validation rules
  - field: fieldName                # JSON field name
    type: string                    # Data type
    required: true                  # Is field required?
    minLength: 3                    # Minimum string length
    maxLength: 50                   # Maximum string length
    pattern: "^[A-Z].*"            # Regex pattern
    enum: ["val1", "val2"]         # Allowed values
    min: 0                         # Minimum numeric value
    max: 100                       # Maximum numeric value
    description: Field description  # Field description
```

### Included Example Schemas

The `schemas/` directory contains ready-to-use validation schemas:

1. **user_registration.yaml**: User registration with email, password, age validation
2. **product_catalog.yaml**: Product information with SKU, price, stock validation
3. **api_config.yaml**: API configuration with endpoint, timeout, rate limiting

### Exporting Schemas to YAML

```go
// Create schema programmatically
schema := ValidationSchema{
    Name:        "MySchema",
    Description: "My validation schema",
    Rules:       []*ValidationRule{
        // ... your rules
    },
}

// Export to YAML
yamlOutput, err := yaml.Marshal(&schema)
if err != nil {
    log.Fatal(err)
}

// Save to file
ioutil.WriteFile("exported_schema.yaml", yamlOutput, 0644)
```

---

## Español

### Validador JSON con Configuración de Esquema YAML

Una versión mejorada del validador JSON que permite definir reglas de validación usando archivos o cadenas YAML, haciendo la gestión de esquemas más flexible y mantenible.

### Características

- **Definición de esquema basada en YAML**: Define reglas de validación en formato YAML fácil de leer
- **Múltiples opciones de carga**:
  - Cargar desde archivos YAML
  - Cargar desde cadenas YAML
  - Definición programática de reglas
  - Enfoque mixto (YAML + programático)
- **Exportación de esquemas**: Exporta esquemas creados programáticamente a YAML
- **Todas las características del validador base**:
  - Múltiples tipos de datos (string, number, integer, boolean, array, object)
  - Campos requeridos/opcionales
  - Restricciones min/max
  - Coincidencia de patrones
  - Valores de enumeración
  - Restricciones de longitud

### Instalación

```bash
# Instalar dependencias
go mod init miproyecto
go get gopkg.in/yaml.v3

# O usar el go.mod proporcionado
cd examples/json_validator_yaml
go mod tidy
```

### Cómo Usar

#### 1. Ejecutar el Ejemplo

```bash
go run validator_with_yaml.go
```

#### 2. Crear un Archivo de Esquema YAML

```yaml
name: RegistroUsuario
description: Esquema para validación de registro de usuario
rules:
  - field: username
    type: string
    required: true
    minLength: 3
    maxLength: 20
    pattern: "^[a-zA-Z0-9_]+$"
    description: Nombre de usuario para la cuenta
  
  - field: email
    type: string
    required: true
    pattern: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
    description: Dirección de email del usuario
  
  - field: age
    type: integer
    required: true
    min: 18
    max: 120
    description: Edad del usuario
  
  - field: country
    type: string
    required: false
    enum: ["México", "España", "Argentina", "Colombia"]
    description: País de residencia
```

#### 3. Cargar Esquema desde Archivo YAML

```go
validator := NewValidator()
err := validator.LoadSchemaFromYAML("schemas/registro_usuario.yaml")
if err != nil {
    log.Fatal(err)
}

// Obtener información del esquema
nombre, descripcion := validator.GetSchemaInfo()
fmt.Printf("Esquema: %s - %s\n", nombre, descripcion)
```

#### 4. Cargar Esquema desde Cadena YAML

```go
contenidoYAML := `
name: EsquemaSimple
description: Un esquema de validación simple
rules:
  - field: nombre
    type: string
    required: true
    maxLength: 50
`

validator := NewValidator()
err := validator.LoadSchemaFromYAMLString(contenidoYAML)
if err != nil {
    log.Fatal(err)
}
```

#### 5. Enfoque Mixto (YAML + Programático)

```go
// Cargar esquema base desde YAML
validator := NewValidator()
err := validator.LoadSchemaFromYAML("esquema_base.yaml")

// Agregar reglas adicionales programáticamente
valorMin := 0.0
valorMax := 100.0
validator.AddRule(&ValidationRule{
    FieldName:   "puntaje",
    Type:        "number",
    Required:    true,
    Min:         &valorMin,
    Max:         &valorMax,
    Description: "Valor del puntaje",
})
```

#### 6. Validar Datos JSON

```go
datosJSON := `{
    "username": "juan_perez",
    "email": "juan@ejemplo.com",
    "age": 25,
    "country": "México"
}`

valido, errores := validator.Validate(datosJSON)
if valido {
    fmt.Println("✓ ¡Validación exitosa!")
} else {
    fmt.Println("✗ Validación fallida:")
    for _, err := range errores {
        fmt.Println("  -", err)
    }
}
```

### Estructura del Esquema YAML

```yaml
name: NombreEsquema                 # Identificador del esquema
description: Descripción del esquema # Descripción legible
rules:                              # Lista de reglas de validación
  - field: nombreCampo              # Nombre del campo JSON
    type: string                    # Tipo de dato
    required: true                  # ¿Es campo requerido?
    minLength: 3                    # Longitud mínima de cadena
    maxLength: 50                   # Longitud máxima de cadena
    pattern: "^[A-Z].*"            # Patrón regex
    enum: ["val1", "val2"]         # Valores permitidos
    min: 0                         # Valor numérico mínimo
    max: 100                       # Valor numérico máximo
    description: Descripción campo  # Descripción del campo
```

### Esquemas de Ejemplo Incluidos

El directorio `schemas/` contiene esquemas de validación listos para usar:

1. **user_registration.yaml**: Registro de usuario con validación de email, contraseña, edad
2. **product_catalog.yaml**: Información de producto con validación de SKU, precio, stock
3. **api_config.yaml**: Configuración de API con endpoint, timeout, límite de tasa

### Exportar Esquemas a YAML

```go
// Crear esquema programáticamente
esquema := ValidationSchema{
    Name:        "MiEsquema",
    Description: "Mi esquema de validación",
    Rules:       []*ValidationRule{
        // ... tus reglas
    },
}

// Exportar a YAML
salidaYAML, err := yaml.Marshal(&esquema)
if err != nil {
    log.Fatal(err)
}

// Guardar en archivo
ioutil.WriteFile("esquema_exportado.yaml", salidaYAML, 0644)
```

---

## Advanced Examples / Ejemplos Avanzados

### Complex Validation Schema / Esquema de Validación Complejo

```yaml
name: OrderValidation
description: E-commerce order validation schema
rules:
  - field: orderId
    type: string
    required: true
    pattern: "^ORD-[0-9]{10}$"
    
  - field: customer
    type: object
    required: true
    description: Customer information
    
  - field: items
    type: array
    required: true
    description: Order items
    
  - field: totalAmount
    type: number
    required: true
    min: 0.01
    max: 1000000
    
  - field: status
    type: string
    required: true
    enum: ["pending", "processing", "shipped", "delivered", "cancelled"]
    
  - field: paymentMethod
    type: string
    required: true
    enum: ["credit_card", "debit_card", "paypal", "bank_transfer"]
    
  - field: shippingAddress
    type: object
    required: true
    
  - field: notes
    type: string
    required: false
    maxLength: 500
```

---

## License / Licencia

This example is part of the go-dsl project. See the main project LICENSE file for details.

Este ejemplo es parte del proyecto go-dsl. Consulte el archivo LICENSE del proyecto principal para más detalles.