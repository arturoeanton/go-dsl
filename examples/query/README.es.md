# Sistema de Consultas DSL Universal - Query en Español

**Sistema de consultas universal 100% genérico que funciona con CUALQUIER estructura usando reflexión, soporte bilingüe (español/inglés), y zero errores de parsing.**

## 🎯 Objetivo

Este ejemplo demuestra un **sistema de consultas empresarial universal** que incluye:

- 🔄 Sistema 100% genérico via reflexión - funciona con CUALQUIER estructura
- 🌍 Soporte bilingüe completo (español e inglés simultáneamente)  
- 📊 CERO errores de parsing - completamente universal
- 🏷️ Soporte para struct tags `query:"fieldname"`
- 🔄 Compatibilidad hacia atrás con estructuras sin tags
- 🎯 Operaciones avanzadas (buscar, listar, contar con filtros)
- 📝 Sintaxis natural tipo SQL en español e inglés
- 🚀 Listo para producción - reutilización ilimitada

## 🚀 Ejecución Rápida

```bash
cd examples/query
go run main.go
```

## ✨ Características Universales

### 1. **100% Genérico - Cualquier Estructura**

```go
// ✅ Funciona con Product
type Product struct {
    ID       int     `query:"id"`
    Name     string  `query:"nombre"`
    Category string  `query:"categoria"`
    Price    float64 `query:"precio"`
    Stock    int     `query:"stock"`
}

// ✅ Funciona con Employee
type Employee struct {
    ID         int     `query:"id"`
    Name       string  `query:"nombre"`
    Department string  `query:"departamento"`
    Position   string  `query:"posicion"`
    Salary     float64 `query:"salario"`
    Age        int     `query:"edad"`
}

// ✅ Funciona con Customer (SIN tags - compatible!)
type Customer struct {
    ID    int
    Name  string
    Email string
    Phone string
}
```

### 2. **Soporte Bilingüe Completo**

```go
// Mismos datos, ambos idiomas funcionan:
"listar productos donde precio mayor 100"           // Español
"list productos where precio greater 100"           // Inglés
"buscar empleados donde departamento es Engineering" // Español  
"search empleados where departamento is Engineering" // Inglés
```

### 3. **Detección Automática de Campos**

```go
// El motor detecta automáticamente campos disponibles
fields := engine.GetFieldNames(products[0])
// Con tags: ["id", "nombre", "categoria", "precio", "stock"]
// Sin tags: ["id", "name", "email", "phone"]
```

## 📚 Sintaxis de Consultas Soportada

### Comandos Básicos (Bilingües)

```go
// Español
dsl.KeywordToken("BUSCAR", "buscar")      // search
dsl.KeywordToken("LISTAR", "listar")      // list  
dsl.KeywordToken("CONTAR", "contar")      // count

// Inglés  
dsl.KeywordToken("SEARCH", "search")
dsl.KeywordToken("LIST", "list")
dsl.KeywordToken("COUNT", "count")

// Operadores (Español)
dsl.KeywordToken("DONDE", "donde")        // where
dsl.KeywordToken("ES", "es")              // is
dsl.KeywordToken("MAYOR", "mayor")        // greater
dsl.KeywordToken("MENOR", "menor")        // less
dsl.KeywordToken("CONTIENE", "contiene")  // contains

// Operadores (Inglés)
dsl.KeywordToken("WHERE", "where")
dsl.KeywordToken("IS", "is") 
dsl.KeywordToken("GREATER", "greater")
dsl.KeywordToken("LESS", "less")
dsl.KeywordToken("CONTAINS", "contains")
```

### 1. **Consultas Simples (Bilingües)**

```sql
-- Español
listar productos                    -- Lista todos los productos
contar empleados                   -- Cuenta todos los empleados
buscar pedidos                     -- Busca todos los pedidos

-- Inglés  
list productos                     -- Lists all products
count empleados                    -- Counts all employees
search pedidos                     -- Searches all orders
```

### 2. **Filtros con Strings (Bilingües)**

```sql
-- Español
buscar productos donde categoria es "Electronics"
listar empleados donde nombre contiene "García"
contar pedidos donde estado es "completado"

-- Inglés
search productos where categoria is "Electronics"  
list empleados where nombre contains "García"
count pedidos where estado is "completado"
```

### 3. **Filtros Numéricos (Bilingües)**

```sql
-- Español
listar productos donde precio mayor 100
contar empleados donde salario menor 50000
buscar pedidos donde monto mayor 1000

-- Inglés
list productos where precio greater 100
count empleados where salario less 50000  
search pedidos where monto greater 1000
```

### 4. **Filtros de Contenido (Bilingües)**

```sql
-- Español
buscar productos donde nombre contiene "USB"
listar empleados donde posicion contiene "Developer"
contar clientes donde email contiene "@gmail.com"

-- Inglés
search productos where nombre contains "USB"
list empleados where posicion contains "Developer"
count clientes where email contains "@gmail.com"
```

## 🏗️ Arquitectura Universal

### Motor de Consultas Genérico

```go
// UniversalQueryEngine - 100% genérico usando reflexión
type UniversalQueryEngine struct{}

func (uqe *UniversalQueryEngine) GetFieldNames(item interface{}) []string {
    v := reflect.ValueOf(item)
    t := reflect.TypeOf(item)
    
    // Manejar tipos puntero
    if v.Kind() == reflect.Ptr {
        v = v.Elem()
        t = t.Elem()
    }
    
    var fields []string
    for i := 0; i < v.NumField(); i++ {
        field := t.Field(i)
        
        // Prioridad 1: struct tag "query"
        if tag := field.Tag.Get("query"); tag != "" {
            fields = append(fields, tag)
        } else {
            // Prioridad 2: nombre del campo en minúsculas (compatible)
            fields = append(fields, strings.ToLower(field.Name))
        }
    }
    
    return fields
}
```

### Filtrado Universal

```go
func (uqe *UniversalQueryEngine) ApplyFilter(data []interface{}, field string, operator string, value interface{}) []interface{} {
    var result []interface{}
    
    for _, item := range data {
        fieldValue := uqe.GetFieldValue(item, field)
        if fieldValue == nil {
            continue
        }
        
        if uqe.compareValues(fieldValue, operator, value) {
            result = append(result, item)
        }
    }
    
    return result
}
```

### Comparación Polimórfica

```go
func (uqe *UniversalQueryEngine) compareValues(fieldValue interface{}, operator string, compareValue interface{}) bool {
    switch operator {
    case "es", "is", "==":
        return uqe.equalCompare(fieldValue, compareValue)
    case "mayor", "greater", ">":
        return uqe.numericCompare(fieldValue, compareValue) > 0
    case "menor", "less", "<":
        return uqe.numericCompare(fieldValue, compareValue) < 0
    case "contiene", "contains":
        return uqe.containsCompare(fieldValue, compareValue)
    }
    return false
}
```

## 📊 Ejemplo de Salida

```
=== Universal Query DSL - Works with ANY Struct ===
✅ 100% generic using reflection
✅ Supports Spanish and English
✅ No hardcoded field names
✅ Works with unlimited struct types
✅ Supports struct tags (query:"fieldname")
✅ Backward compatible with structs without tags

=== Testing with Products Data ===
Available fields: [id nombre categoria precio stock]

1. Query: listar productos
   Resultado: 7 elementos encontrados
     - id: 1, nombre: Laptop Dell, categoria: Electronics, precio: 1200, stock: 5
     - id: 2, nombre: Mouse Logitech, categoria: Electronics, precio: 25, stock: 50
     - id: 3, nombre: Desk Chair, categoria: Furniture, precio: 350, stock: 10
     ... and 4 more

2. Query: buscar productos donde categoria es "Electronics"  
   Resultado: 4 elementos encontrados
     - id: 1, nombre: Laptop Dell, categoria: Electronics, precio: 1200, stock: 5
     - id: 2, nombre: Mouse Logitech, categoria: Electronics, precio: 25, stock: 50
     - id: 5, nombre: USB Cable, categoria: Electronics, precio: 10, stock: 100
     ... and 1 more

3. Query: list productos where categoria is "Furniture"
   Resultado: 3 elementos encontrados
     - id: 3, nombre: Desk Chair, categoria: Furniture, precio: 350, stock: 10
     - id: 4, nombre: Standing Desk, categoria: Furniture, precio: 600, stock: 3
     - id: 7, nombre: Office Lamp, categoria: Furniture, precio: 45, stock: 20

=== Testing with Employees Data ===
Available fields: [id nombre departamento posicion salario edad]

1. Query: buscar empleados donde departamento es "Engineering"
   Resultado: 3 elementos encontrados
     - id: 1, nombre: Juan García, departamento: Engineering, posicion: Senior Developer, salario: 75000, edad: 28
     - id: 3, nombre: Carlos Rodríguez, departamento: Engineering, posicion: Tech Lead, salario: 85000, edad: 42
     - id: 5, nombre: Pedro Sánchez, departamento: Engineering, posicion: Developer, salario: 55000, edad: 31

2. Query: list empleados where departamento is "Marketing"
   Resultado: 1 elementos encontrados
     - id: 2, nombre: María López, departamento: Marketing, posicion: Manager, salario: 65000, edad: 35

=== Testing with Customers Data (NO Tags - Backward Compatible!) ===
Available fields (fallback to field names): [id name email phone]

1. Query: buscar clientes donde name contiene "Alice"
   Resultado: 1 elementos encontrados
     - ID: 1, Name: Alice Brown, Email: alice@email.com, Phone: 555-0101

2. Query: list clientes where name is "Bob Smith"
   Resultado: 1 elementos encontrados
     - ID: 2, Name: Bob Smith, Email: bob@email.com, Phone: 555-0102

=== ✅ Universal Query DSL SUCCESS ===
✅ ZERO parsing errors!
✅ Works with Product, Employee, Order, Customer - ANY struct!
✅ Supports both Spanish and English in same system!
✅ 100% generic via reflection
✅ Supports struct tags for custom field names
✅ Backward compatible with structs without tags
✅ Unlimited reusability across domains
✅ Production ready!
```

## 🔧 Características Técnicas Avanzadas

### 1. **Reflexión para Tipos Dinámicos**

```go
// Obtener valor de campo usando reflexión
func (uqe *UniversalQueryEngine) GetFieldValue(item interface{}, fieldName string) interface{} {
    v := reflect.ValueOf(item)
    t := reflect.TypeOf(item)
    
    // Manejar tipos puntero
    if v.Kind() == reflect.Ptr {
        v = v.Elem()
        t = t.Elem()
    }
    
    for i := 0; i < v.NumField(); i++ {
        field := t.Field(i)
        
        // Buscar por tag primero
        if tag := field.Tag.Get("query"); tag == fieldName {
            return v.Field(i).Interface()
        }
        
        // Buscar por nombre de campo
        if strings.ToLower(field.Name) == fieldName {
            return v.Field(i).Interface()
        }
    }
    
    return nil
}
```

### 2. **Conversión Automática de Slices**

```go
// ConvertToInterfaceSlice convierte cualquier tipo de slice a []interface{}
func (uqe *UniversalQueryEngine) ConvertToInterfaceSlice(slice interface{}) []interface{} {
    v := reflect.ValueOf(slice)
    if v.Kind() != reflect.Slice {
        return nil
    }
    
    result := make([]interface{}, v.Len())
    for i := 0; i < v.Len(); i++ {
        result[i] = v.Index(i).Interface()
    }
    
    return result
}
```

### 3. **Sistema de Reglas Jerárquicas**

```go
// Reglas organizadas de más específica a menos específica
func (uq *UniversalQueryDSL) setupRules() {
    // MÁS específicas primero (6 tokens)
    uq.dsl.Rule("query", []string{"BUSCAR", "WORD", "DONDE", "WORD", "CONTIENE", "STRING"}, "filteredQueryStringContains")
    uq.dsl.Rule("query", []string{"SEARCH", "WORD", "WHERE", "WORD", "CONTAINS", "STRING"}, "filteredQueryStringContains")
    
    // Menos específicas después (2 tokens)
    uq.dsl.Rule("query", []string{"LISTAR", "WORD"}, "simpleQuery")
    uq.dsl.Rule("query", []string{"LIST", "WORD"}, "simpleQuery")
}
```

## 🎯 Casos de Uso Empresariales

### 1. **Sistema de Recursos Humanos**

```go
// Estructura Employee universal
type Employee struct {
    ID         int     `query:"id"`
    Name       string  `query:"nombre"`
    Department string  `query:"departamento"`
    Position   string  `query:"posicion"`
    Salary     float64 `query:"salario"`
    StartDate  string  `query:"fecha_inicio"`
}

// Consultas bilingües
queries := []string{
    "listar empleados donde salario mayor 60000",
    "list empleados where departamento is Engineering",
    "contar empleados donde edad menor 30",
    "search empleados where posicion contains Manager",
}
```

### 2. **Sistema de Inventario**

```go
// Estructura Product universal
type Product struct {
    SKU        string  `query:"sku"`
    Name       string  `query:"nombre"`
    Category   string  `query:"categoria"`
    Quantity   int     `query:"cantidad"`
    Price      float64 `query:"precio"`
    Supplier   string  `query:"proveedor"`
}

// Consultas de negocio bilingües
queries := []string{
    "buscar productos donde cantidad menor 10",              // Stock bajo
    "list productos where categoria is Electronics",         // Por categoría
    "contar productos donde precio mayor 100",              // Productos caros
    "search productos where proveedor contains Samsung",     // Por proveedor
}
```

### 3. **Sistema de Ventas**

```go
// Estructura Sale universal
type Sale struct {
    ID         int     `query:"id"`
    Customer   string  `query:"cliente"`
    Product    string  `query:"producto"`
    Amount     float64 `query:"monto"`
    Date       string  `query:"fecha"`
    Region     string  `query:"region"`
}

// Reportes dinámicos bilingües
queries := []string{
    "listar ventas donde monto mayor 1000",
    "list ventas where region is North",
    "contar ventas donde fecha contiene 2025",
    "search ventas where cliente contains Garcia",
}
```

### 4. **Sistema CRM**

```go
// Estructura Lead universal
type Lead struct {
    ID       int    `query:"id"`
    Name     string `query:"nombre"`
    Email    string `query:"email"`
    Status   string `query:"estado"`
    Source   string `query:"origen"`
    Score    int    `query:"puntuacion"`
}

// Análisis de leads bilingües
queries := []string{
    "buscar leads donde estado es qualified",
    "list leads where puntuacion mayor 80",
    "contar leads donde origen is website",
    "search leads where email contains gmail",
}
```

## 🚀 Extensiones Posibles

### 1. **Más Operadores**

```go
// Operadores adicionales bilingües
dsl.KeywordToken("LIKE", "como")           // Pattern matching
dsl.KeywordToken("IN", "en")               // Lista de valores
dsl.KeywordToken("BETWEEN", "entre")       // Rangos
dsl.KeywordToken("NOT", "no")              // Negación

// Uso bilingüe
"buscar productos donde nombre como 'Dell%'"
"search products where nombre like 'Dell%'"
"listar pedidos donde monto entre 100 y 500"
"list orders where monto between 100 and 500"
```

### 2. **Funciones de Agregación**

```go
// Agregaciones bilingües
dsl.KeywordToken("SUMA", "suma")           // sum
dsl.KeywordToken("PROMEDIO", "promedio")   // average
dsl.KeywordToken("MAXIMO", "maximo")       // max
dsl.KeywordToken("MINIMO", "minimo")       // min

// Uso bilingüe
"suma salario de empleados donde departamento es IT"
"sum salario from empleados where departamento is IT"
"promedio precio de productos donde categoria es Electronics"
"average precio from productos where categoria is Electronics"
```

### 3. **Ordenamiento**

```go
// Ordenamiento bilingüe
dsl.KeywordToken("ORDENAR", "ordenar")     // order
dsl.KeywordToken("POR", "por")             // by
dsl.KeywordToken("ASC", "asc")             // ascending
dsl.KeywordToken("DESC", "desc")           // descending

// Uso bilingüe
"listar productos ordenar por precio desc"
"list productos order by precio desc"
```

### 4. **Agrupación**

```go
// Agrupación bilingüe
dsl.KeywordToken("AGRUPAR", "agrupar")     // group
dsl.KeywordToken("POR", "por")             // by

// Uso bilingüe
"contar ventas agrupar por region"
"count ventas group by region"
```

## 🎓 Lecciones Técnicas

### 1. **Reflexión es Universal**
Permite crear sistemas que funcionan con cualquier estructura sin modificar código.

### 2. **Struct Tags para Localización**  
`query:"fieldname"` permite mapear campos a términos en español/inglés.

### 3. **Compatibilidad Total**
Estructuras sin tags siguen funcionando usando nombres de campo.

### 4. **Bilingüismo Simultáneo**
Español e inglés coexisten en el mismo motor sin conflictos.

### 5. **Zero Parsing Errors**
Sistema robusto que nunca falla con diferentes tipos de estructuras.

## 🔗 Patrones Similares

- **Entity Framework LINQ**: Consultas LINQ en .NET
- **Hibernate Criteria**: Consultas dinámicas en Java
- **SQLAlchemy ORM**: Query objects en Python
- **Eloquent ORM**: Query builder en Laravel
- **Django ORM**: API de consultas en Django
- **MongoDB Query Language**: Consultas NoSQL
- **GraphQL**: Consultas flexibles de datos

## 🚀 Próximos Pasos

1. **Ejecuta el ejemplo**: `go run main.go`
2. **Define tus propias estructuras** con tags personalizados
3. **Experimenta con consultas bilingües**
4. **Integra en tu sistema** de datos empresarial
5. **Extiende con nuevos operadores** según necesites
6. **Combina español e inglés** en el mismo sistema

## 📞 Referencias y Documentación

- **Código fuente**: [`main.go`](main.go)
- **Motor universal**: [`universal/`](universal/)
- **Sistema LINQ**: [../linq/](../linq/)
- **Contexto dinámico**: [../simple_context/](../simple_context/)
- **Manual completo**: [Manual de Uso](../../docs/es/manual_de_uso.md)
- **Guía técnica**: [Developer Onboarding](../../docs/es/developer_onboarding.md)

---

**¡Demuestra que go-dsl puede crear sistemas de consulta universales 100% genéricos con soporte bilingüe!** 🔍🌍🎉