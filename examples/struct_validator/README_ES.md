# Validador de Structs y Maps de Go

Una biblioteca de validación flexible y poderosa para structs de Go y `map[string]interface{}` usando definiciones de esquema basadas en YAML.

## Características

- ✅ **Soporte de validación dual**: Valida tanto structs de Go como `map[string]interface{}`
- ✅ **Esquemas basados en YAML**: Define reglas de validación en formato YAML legible
- ✅ **Validación de structs basada en reflexión**: Valida automáticamente los campos de struct usando reflexión
- ✅ **Validación de objetos anidados**: Soporte para estructuras anidadas complejas
- ✅ **Validación de arrays/slices**: Valida elementos de array con reglas personalizadas
- ✅ **Múltiples tipos de validación**:
  - String (minLength, maxLength, pattern, enum)
  - Number/Integer (min, max)
  - Boolean
  - Array/Slice (con validación de elementos)
  - Object/Struct (con validación anidada)
- ✅ **Soporte de etiquetas de campo**: Reconoce etiquetas de struct `json` y `yaml`
- ✅ **Campos opcionales y requeridos**: Configuración flexible de requisitos de campo
- ✅ **Coincidencia de patrones**: Validación de strings basada en regex
- ✅ **Validación de enum**: Restringe valores a conjuntos predefinidos

## Instalación

```bash
go get github.com/arturoeanton/go-dsl
```

## Inicio Rápido

### 1. Define tu Esquema de Validación (YAML)

```yaml
name: ValidacionUsuario
description: Esquema para validación de usuario
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

### 2. Valida un Struct de Go

```go
package main

import (
    "fmt"
    "io/ioutil"
)

type Usuario struct {
    Username string `json:"username"`
    Email    string `json:"email"`
    Age      int    `json:"age"`
}

func main() {
    // Crear validador
    validador := NewStructValidator()
    
    // Cargar esquema desde archivo
    contenidoEsquema, _ := ioutil.ReadFile("schemas/user.yaml")
    validador.LoadSchemaFromYAML(string(contenidoEsquema))
    
    // Validar struct
    usuario := Usuario{
        Username: "juan_perez",
        Email:    "juan@ejemplo.com",
        Age:      25,
    }
    
    valido, errores := validador.ValidateStruct(usuario)
    if valido {
        fmt.Println("✓ ¡Validación exitosa!")
    } else {
        fmt.Println("✗ Validación fallida:")
        for _, err := range errores {
            fmt.Println("  -", err)
        }
    }
}
```

### 3. Valida un Map

```go
// Validar map[string]interface{}
datos := map[string]interface{}{
    "username": "maria_garcia",
    "email":    "maria@ejemplo.com",
    "age":      30,
}

valido, errores := validador.ValidateMap(datos)
```

## Definición de Esquema

### Tipos de Campo Básicos

| Tipo | Descripción | Restricciones |
|------|-------------|---------------|
| `string` | Valores de cadena | minLength, maxLength, pattern, enum |
| `integer` | Números enteros | min, max |
| `number` | Números de punto flotante | min, max |
| `boolean` | Valores booleanos | - |
| `array` | Arrays/Slices | items (para validación de elementos) |
| `object` | Objetos/Structs | nested (para validación anidada) |

### Propiedades de Campo

```yaml
- field: nombreCampo     # Nombre del campo (requerido)
  type: string          # Tipo de campo (requerido)
  required: true        # ¿Es campo requerido? (por defecto: false)
  description: "..."    # Descripción del campo (opcional)
  
  # Restricciones de string
  minLength: 3          # Longitud mínima de string
  maxLength: 50         # Longitud máxima de string
  pattern: "^[A-Z]+$"   # Patrón regex
  enum: ["A", "B", "C"] # Valores permitidos
  
  # Restricciones numéricas
  min: 0                # Valor mínimo
  max: 100              # Valor máximo
  
  # Validación de array
  items:                # Validación para elementos del array
    field: item
    type: string
    minLength: 1
  
  # Validación de objeto anidado
  nested:               # Validación para objetos anidados
    name: EsquemaAnidado
    rules:
      - field: campoAnidado
        type: string
        required: true
```

## Ejemplos Avanzados

### Validación de Objetos Anidados

```yaml
- field: direccion
  type: object
  required: true
  nested:
    name: Direccion
    rules:
      - field: calle
        type: string
        required: true
      - field: ciudad
        type: string
        required: true
      - field: codigoPostal
        type: string
        pattern: "^[0-9]{5}$"
```

### Array con Elementos de Objeto

```yaml
- field: articulos
  type: array
  required: true
  items:
    field: articulo
    type: object
    nested:
      name: ArticuloOrden
      rules:
        - field: productoId
          type: string
          required: true
        - field: cantidad
          type: integer
          min: 1
```

## Archivos de Esquema Disponibles

El directorio `schemas/` contiene varios esquemas de validación preconstruidos:

- **user.yaml**: Validación de registro y perfil de usuario
- **product.yaml**: Validación de catálogo de productos
- **order.yaml**: Validación compleja de órdenes con estructuras anidadas
- **api_config.yaml**: Validación de configuración de API

## Ejecutar Pruebas

```bash
# Ejecutar todas las pruebas
go test ./...

# Ejecutar con cobertura
go test -cover ./...

# Ejecutar prueba específica
go test -run TestStructValidation

# Ejecutar benchmarks
go test -bench=.
```

## Rendimiento

El validador está optimizado para el rendimiento con:
- Uso eficiente de reflexión
- Asignaciones mínimas
- Patrones regex compilados (en caché)

Resultados de benchmark (ejemplo):
```
BenchmarkStructValidation-8    100000    10234 ns/op
BenchmarkMapValidation-8       100000    11456 ns/op
```

## Mensajes de Error

El validador proporciona mensajes de error detallados:

```
El campo 'username' debe tener al menos 3 caracteres
El campo 'email' debe coincidir con el patrón ^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$
El campo 'age' debe ser >= 18
profile.El campo 'country' debe ser uno de [USA Canada Mexico UK Germany France]
items[0].El campo 'quantity' debe ser >= 1
```

## Casos de Uso Comunes

### 1. Validación de Formularios Web

Ideal para validar datos de entrada de formularios antes de procesarlos o almacenarlos en la base de datos.

### 2. Validación de API REST

Valida payloads JSON entrantes en endpoints de API para asegurar la integridad de los datos.

### 3. Configuración de Aplicaciones

Valida archivos de configuración cargados desde YAML o JSON para asegurar que cumplen con los requisitos esperados.

### 4. Procesamiento de Datos por Lotes

Valida grandes conjuntos de datos antes de procesarlos para evitar errores durante el procesamiento.

## Integración con Frameworks

### Gin

```go
func CreateUser(c *gin.Context) {
    var user User
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    validator := NewStructValidator()
    validator.LoadSchemaFromYAML(userSchema)
    
    if valid, errors := validator.ValidateStruct(user); !valid {
        c.JSON(400, gin.H{"errors": errors})
        return
    }
    
    // Procesar usuario válido...
}
```

### Echo

```go
func CreateUser(c echo.Context) error {
    user := new(User)
    if err := c.Bind(user); err != nil {
        return c.JSON(400, map[string]string{"error": err.Error()})
    }
    
    validator := NewStructValidator()
    validator.LoadSchemaFromYAML(userSchema)
    
    if valid, errors := validator.ValidateStruct(*user); !valid {
        return c.JSON(400, map[string][]string{"errors": errors})
    }
    
    // Procesar usuario válido...
    return c.JSON(200, user)
}
```

## Contribuciones

¡Las contribuciones son bienvenidas! Por favor, siéntete libre de enviar un Pull Request.

## Licencia

Este proyecto es parte del framework go-dsl y sigue sus términos de licencia.