# Contexto Simple - Equivalente a r2lang

**Ejemplo fundamental que demuestra cómo go-dsl implementa contexto dinámico equivalente al método `q.use()` de r2lang.**

## 🎯 Objetivo

Este ejemplo demuestra las **características de contexto dinámico** de go-dsl, mostrando:

- 🔄 Equivalencia directa con `q.use()` de r2lang  
- 📊 Acceso a variables y datos del contexto
- 📋 Procesamiento de arrays y estructuras complejas
- 🔄 Dos métodos de manejo de contexto: `SetContext()` vs `Use()`
- 🎯 Operaciones de agregación (Count, Sum, List)

## 🚀 Ejecución Rápida

```bash
cd examples/simple_context
go run main.go
```

## 🔄 Equivalencia con r2lang

### Sintaxis Comparada

| r2lang | go-dsl |
|--------|--------|
| `q.use("get name", {name: "Juan"})` | `dsl.Use("get name", map[string]interface{}{"name": "Juan"})` |
| `context.name` | `dsl.GetContext("name")` |
| Automático | Requiere type assertion: `name.(string)` |

### Ejemplo Directo

```javascript
// r2lang
const result = q.use("get variable", {
    name: "Juan García",
    age: 30,
    city: "Madrid"
});

// go-dsl (equivalente exacto)
result, err := dsl.Use("get variable", map[string]interface{}{
    "name": "Juan García",
    "age":  30,
    "city": "Madrid",
})
```

## 📚 Características del DSL

### Tokens Definidos

```go
// Comandos básicos
dsl.KeywordToken("GET", "get")           // obtener
dsl.KeywordToken("COUNT", "count")       // contar
dsl.KeywordToken("LIST", "list")         // listar
dsl.KeywordToken("SUM", "sum")           // sumar

// Campos específicos
dsl.KeywordToken("USERS", "users")       // usuarios
dsl.KeywordToken("AGES", "ages")         // edades
dsl.KeywordToken("ALL", "all")           // todos/todas

// Variables genéricas
dsl.Token("VARIABLE", "[a-zA-Z_][a-zA-Z0-9_]*") // nombres de variables
```

### Comandos Soportados

#### 1. Acceso a Variables Simples
```
get name      # Obtiene valor de "name" del contexto
get age       # Obtiene valor de "age" del contexto
get city      # Obtiene valor de "city" del contexto
get missing   # Maneja variables no existentes
```

#### 2. Operaciones de Array
```
count users   # Cuenta elementos en array "users"
list all users # Lista todos los usuarios
sum all ages   # Suma todos los valores en "ages"  
count ages     # Cuenta elementos en "ages"
```

#### 3. Variables Dinámicas
```
get [cualquier_variable]  # Acceso dinámico a variables
```

## 🏗️ Arquitectura del Contexto

### Tipos de Datos Soportados

```go
// Variables simples
context := map[string]interface{}{
    "name":  "Juan García",
    "age":   30,
    "city":  "Madrid",
    "score": 95.5,
}

// Arrays simples
context := map[string]interface{}{
    "users":  []string{"Juan", "María", "Carlos", "Ana"},
    "ages":   []int{28, 35, 42, 29},
    "scores": []float64{95.5, 87.2, 92.1, 88.9},
}

// Estructuras complejas
type Person struct {
    Name string
    Age  int
    City string
}

context := map[string]interface{}{
    "people": []Person{
        {"Juan García", 28, "Madrid"},
        {"María López", 35, "Barcelona"},
    },
}
```

### Acciones de Procesamiento

```go
// Acceso simple a variables
dsl.Action("getVariable", func(args []interface{}) (interface{}, error) {
    varName := args[1].(string)
    value := dsl.GetContext(varName)
    
    if value == nil {
        return fmt.Sprintf("Variable '%s' not found", varName), nil
    }
    return value, nil
})

// Conteo de arrays
dsl.Action("countUsers", func(args []interface{}) (interface{}, error) {
    users := dsl.GetContext("users")
    if userArray, ok := users.([]string); ok {
        return len(userArray), nil
    }
    return 0, nil
})

// Suma de valores numéricos
dsl.Action("sumAllAges", func(args []interface{}) (interface{}, error) {
    ages := dsl.GetContext("ages")
    if ageArray, ok := ages.([]int); ok {
        total := 0
        for _, age := range ageArray {
            total += age
        }
        return total, nil
    }
    return 0, nil
})
```

## 🔄 Dos Métodos de Contexto

### Método 1: Use() - Equivalente Directo a r2lang

```go
// Similar a q.use() de r2lang
context := map[string]interface{}{
    "name": "Juan García",
    "age":  30,
    "city": "Madrid",
}

result, err := dsl.Use("get name", context)
// result.GetOutput() → "Juan García"
```

**Ventajas:**
- Equivalencia exacta con r2lang
- Contexto temporal para una operación
- Ideal para datos que cambian frecuentemente

### Método 2: SetContext() - Contexto Persistente

```go
// Establecer contexto que persiste
dsl.SetContext("user", "Alice")
dsl.SetContext("role", "admin")

// Usar en múltiples operaciones
result1, _ := dsl.Parse("get user")  // → "Alice"
result2, _ := dsl.Parse("get role")  // → "admin"
```

**Ventajas:**
- Contexto persiste entre llamadas
- Menos overhead para datos estáticos
- Ideal para configuración global

## 📊 Ejemplo de Salida

```
=== go-dsl Context Examples ===
Equivalent to r2lang's: q.use("query", context)

1. Simple Variable Access
------------------------
  get name -> Juan García
  get age -> 30
  get city -> Madrid
  get score -> 95.5
  get missing -> Variable 'missing' not found

2. Data Array Processing
------------------------
  Count users: 4
  List all users: [Juan María Carlos Ana]
  Sum all scores: 350
  Count ages: 4
  Sum all ages: 134

3. Complex Data Structures
--------------------------
  All names: [Juan García María López Carlos Rodríguez]
  All ages: [28 35 42]
  All cities: [Madrid Barcelona Madrid]

4. SetContext vs Use() methods
------------------------------
  Method 1: Using SetContext()
    user = Alice
    role = admin
  Method 2: Using Use() - equivalent to r2lang's q.use()
    user = Bob
    role = user
    temp = override
```

## 🎯 Casos de Uso Prácticos

### 1. **Configuración de Aplicación**
```go
// Configuración global
dsl.SetContext("environment", "production")
dsl.SetContext("debug", false)
dsl.SetContext("maxConnections", 100)

// Comandos de configuración
result, _ := dsl.Parse("get environment")
```

### 2. **Datos de Usuario Dinámicos**
```go
// Por cada usuario/sesión
userContext := map[string]interface{}{
    "userId":   12345,
    "username": "john.doe",
    "permissions": []string{"read", "write"},
}

result, _ := dsl.Use("check permissions", userContext)
```

### 3. **Procesamiento de Reportes**
```go
// Datos de reporte
reportContext := map[string]interface{}{
    "sales":     []float64{1000, 1500, 2000},
    "customers": []string{"A", "B", "C"},
    "period":    "2025-Q1",
}

totalSales, _ := dsl.Use("sum all sales", reportContext)
```

### 4. **Sistemas Multi-Tenant**
```go
// Por cada tenant
tenantContext := map[string]interface{}{
    "tenantId":   "company-123",
    "plan":       "enterprise",
    "features":   []string{"advanced", "api", "support"},
}

result, _ := dsl.Use("check features", tenantContext)
```

## 🔧 Características Técnicas

### 1. **Type Assertions Necesarias**

```go
// go-dsl requiere type assertions explícitas
name := dsl.GetContext("name").(string)
age := dsl.GetContext("age").(int)

// Verificación segura
if value := dsl.GetContext("optional"); value != nil {
    text := value.(string)
}
```

### 2. **Manejo de Arrays Tipados**

```go
// Arrays requieren type assertion correcta
users := dsl.GetContext("users")
if userArray, ok := users.([]string); ok {
    // Procesar userArray
}
```

### 3. **Contexto Inmutable Durante Parse**

```go
// El contexto no cambia durante un parse individual
result, _ := dsl.Use("comando", context)
// context permanece sin cambios
```

## 🎓 Mejores Prácticas

### 1. **Usa Use() para Datos Dinámicos**
```go
// ✅ Datos que cambian por operación
userCtx := getUserContext(userId)
result, _ := dsl.Use("process user", userCtx)
```

### 2. **Usa SetContext() para Configuración**
```go
// ✅ Configuración global/estática
dsl.SetContext("apiKey", config.ApiKey)
dsl.SetContext("version", "1.2.3")
```

### 3. **Valida Contexto en Acciones**
```go
dsl.Action("safeAccess", func(args []interface{}) (interface{}, error) {
    value := dsl.GetContext("required")
    if value == nil {
        return nil, fmt.Errorf("required context missing")
    }
    return value, nil
})
```

## 🔗 Casos de Uso Similares

- **Sistemas de Plantillas**: Variables dinámicas en templates
- **Motores de Reglas**: Contexto de evaluación de reglas
- **APIs con Estado**: Datos de sesión y usuario
- **Sistemas de Configuración**: Valores dinámicos por entorno
- **Procesamiento ETL**: Datos de contexto para transformaciones

## 🚀 Próximos Pasos

1. **Ejecuta el ejemplo**: `go run main.go`
2. **Modifica el contexto** en el código
3. **Agrega nuevas variables** y comandos
4. **Experimenta con estructuras** más complejas
5. **Combina con otros ejemplos** del proyecto

## 📞 Referencias

- **r2lang comparación**: [Documentación r2lang](https://github.com/arturoeanton/r2lang)
- **Manual completo**: [Manual de Uso](../../docs/es/manual_de_uso.md)
- **Ejemplo avanzado**: [Sistema multi-país](../accounting/)

---

**¡Este ejemplo demuestra que go-dsl es un reemplazo completo y mejorado para r2lang!** 🔄🎉