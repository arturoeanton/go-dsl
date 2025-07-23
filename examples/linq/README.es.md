# LINQ Universal DSL - Sistema LINQ en Español

**Sistema LINQ 100% genérico que funciona con CUALQUIER estructura usando reflexión, sin errores de parsing, compatible con struct tags y sin tags.**

## 🎯 Objetivo

Este ejemplo demuestra un **sistema LINQ universal** que incluye:

- 🔄 LINQ 100% genérico via reflexión - funciona con CUALQUIER estructura
- 📊 CERO errores de parsing - completamente universal
- 🏷️ Soporte para struct tags `linq:"fieldname"`
- 🔄 Compatibilidad hacia atrás con estructuras sin tags
- 🎯 Operaciones SELECT, WHERE, ORDERBY, TOP
- 📝 Sintaxis natural tipo SQL/LINQ
- 🚀 Listo para producción - reutilización ilimitada

## 🚀 Ejecución Rápida

```bash
cd examples/linq
go run main.go
```

## ✨ Características Universales

### 1. **100% Genérico - Cualquier Estructura**

```go
// ✅ Funciona con Person
type Person struct {
    ID         int     `linq:"id"`
    Name       string  `linq:"name"`
    Age        int     `linq:"age"`
    Department string  `linq:"department"`
    Salary     float64 `linq:"salary"`
}

// ✅ Funciona con Product
type Product struct {
    ID       int     `linq:"id"`
    Name     string  `linq:"name"`
    Category string  `linq:"category"`
    Price    float64 `linq:"price"`
    Stock    int     `linq:"stock"`
}

// ✅ Funciona con Customer (SIN tags - compatible!)
type Customer struct {
    ID    int
    Name  string
    Email string
    Phone string
}
```

### 2. **Soporte para Struct Tags**

```go
// Con tags personalizados
type Order struct {
    ID       int     `linq:"id"`
    Customer string  `linq:"customer"`  // Campo "Customer" → query "customer"
    Amount   float64 `linq:"amount"`
    Status   string  `linq:"status"`
    Date     string  `linq:"date"`
}
```

### 3. **Detección Automática de Campos**

```go
// El motor detecta automáticamente campos disponibles
fields := queryEngine.GetFieldNames(people[0])
// Con tags: ["id", "name", "age", "department", "salary"]
// Sin tags: ["ID", "Name", "Email", "Phone"]
```

## 📚 Sintaxis de Consultas Soportada

### Comandos Básicos

```go
// Tokens del DSL
dsl.KeywordToken("FROM", "from")
dsl.KeywordToken("SELECT", "select")
dsl.KeywordToken("WHERE", "where")
dsl.KeywordToken("ORDERBY", "orderby")
dsl.KeywordToken("TOP", "top")
dsl.KeywordToken("ASC", "asc")
dsl.KeywordToken("DESC", "desc")
```

### 1. **Consultas SELECT**

```sql
from people select *                    -- Todos los campos
from people select name                 -- Campo específico
from products select name               -- Funciona con cualquier tabla/estructura
```

### 2. **Filtros WHERE**

```sql
from people where age > 30 select name
from people where department == "Engineering" select name
from products where price > 100 select name
from products where category == "Electronics" select name
```

### 3. **Ordenamiento ORDERBY**

```sql
from people where salary > 50000 select name orderby salary desc
from products where stock < 20 select name orderby price desc
```

### 4. **Límite TOP**

```sql
from people top 3 select name
from products top 2 select name
from orders top 5 select customer
```

### 5. **Combinaciones Complejas**

```sql
from people where age > 30 and department == "Engineering" select name orderby salary desc top 2
```

## 🏗️ Arquitectura Universal

### Motor de Consultas Genérico

```go
// QueryEngine - 100% genérico usando reflexión
type QueryEngine struct{}

func (qe *QueryEngine) GetFieldNames(item interface{}) []string {
    v := reflect.ValueOf(item)
    t := reflect.TypeOf(item)
    
    var fields []string
    for i := 0; i < v.NumField(); i++ {
        field := t.Field(i)
        
        // Prioridad 1: struct tag "linq"
        if tag := field.Tag.Get("linq"); tag != "" {
            fields = append(fields, tag)
        } else {
            // Prioridad 2: nombre del campo (compatible)
            fields = append(fields, strings.ToLower(field.Name))
        }
    }
    
    return fields
}
```

### Filtrado Universal

```go
func (qe *QueryEngine) ApplyFilter(data []interface{}, field string, operator string, value interface{}) []interface{} {
    var result []interface{}
    
    for _, item := range data {
        v := reflect.ValueOf(item)
        t := reflect.TypeOf(item)
        
        fieldValue := qe.getFieldValue(v, t, field)
        if fieldValue == nil {
            continue
        }
        
        if qe.compareValues(fieldValue, operator, value) {
            result = append(result, item)
        }
    }
    
    return result
}
```

## 📊 Ejemplo de Salida

```
=== Universal LINQ DSL - Works with ANY Struct ===
✅ 100% generic using reflection
✅ No hardcoded field names
✅ No parsing errors
✅ Works with unlimited struct types
✅ Supports struct tags (linq:"fieldname")
✅ Backward compatible with structs without tags

=== Testing with People Data ===
Available fields: [id name age city department salary]

1. Query: from people select *
   Results (5 items):
   ID: 1, Name: Juan García, Age: 28, City: Madrid, Department: Engineering, Salary: 45000.00
   ID: 2, Name: María López, Age: 35, City: Barcelona, Department: Marketing, Salary: 52000.00
   ID: 3, Name: Carlos Rodríguez, Age: 42, City: Madrid, Department: Engineering, Salary: 68000.00
   ID: 4, Name: Ana Martínez, Age: 29, City: Valencia, Department: Sales, Salary: 38000.00
   ID: 5, Name: Pedro Sánchez, Age: 31, City: Barcelona, Department: Engineering, Salary: 48000.00

2. Query: from people where age > 30 select name
   Results (3 items):
   María López
   Carlos Rodríguez
   Pedro Sánchez

3. Query: from people where department == "Engineering" select name
   Results (3 items):
   Juan García
   Carlos Rodríguez
   Pedro Sánchez

=== Testing with Product Data ===
Available fields: [id name category price stock]

1. Query: from products where category == "Electronics" select name
   Results (3 items):
   Laptop Dell
   Mouse Logitech
   USB Cable

=== Testing with Customers Data (NO Tags - Backward Compatible!) ===
Available fields (fallback to field names): [id name email phone]

1. Query: from customers select name
   Results (3 items):
   Alice Brown
   Bob Smith
   Carol Johnson

=== ✅ Universal LINQ DSL SUCCESS ===
✅ ZERO parsing errors!
✅ Works with Person, Product, Order, Customer - ANY struct!
✅ 100% generic via reflection
✅ Supports struct tags for custom field names
✅ Backward compatible with structs without tags
✅ Unlimited reusability
✅ Production ready!
```

## 🔧 Características Técnicas Avanzadas

### 1. **Reflexión para Tipos Dinámicos**

```go
// Obtener valor de campo usando reflexión
func (qe *QueryEngine) getFieldValue(v reflect.Value, t reflect.Type, fieldName string) interface{} {
    for i := 0; i < v.NumField(); i++ {
        field := t.Field(i)
        
        // Buscar por tag primero
        if tag := field.Tag.Get("linq"); tag == fieldName {
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

### 2. **Comparación Polimórfica**

```go
func (qe *QueryEngine) compareValues(fieldValue interface{}, operator string, compareValue interface{}) bool {
    switch operator {
    case "==":
        return fmt.Sprintf("%v", fieldValue) == fmt.Sprintf("%v", compareValue)
    case ">":
        return qe.numericCompare(fieldValue, compareValue) > 0
    case "<":
        return qe.numericCompare(fieldValue, compareValue) < 0
    case ">=":
        return qe.numericCompare(fieldValue, compareValue) >= 0
    case "<=":
        return qe.numericCompare(fieldValue, compareValue) <= 0
    }
    return false
}
```

### 3. **Ordenamiento Genérico**

```go
func (qe *QueryEngine) OrderBy(data []interface{}, field string, desc bool) []interface{} {
    sort.Slice(data, func(i, j int) bool {
        vi := reflect.ValueOf(data[i])
        vj := reflect.ValueOf(data[j])
        ti := reflect.TypeOf(data[i])
        
        fieldI := qe.getFieldValue(vi, ti, field)
        fieldJ := qe.getFieldValue(vj, ti, field)
        
        if desc {
            return qe.compareGeneric(fieldI, fieldJ) > 0
        }
        return qe.compareGeneric(fieldI, fieldJ) < 0
    })
    return data
}
```

## 🎯 Casos de Uso Empresariales

### 1. **Sistema de Recursos Humanos**

```go
// Estructura Employee
type Employee struct {
    ID         int     `linq:"id"`
    Name       string  `linq:"name"`
    Department string  `linq:"dept"`
    Position   string  `linq:"position"`
    Salary     float64 `linq:"salary"`
    StartDate  string  `linq:"start_date"`
}

// Consultas
queries := []string{
    "from employees select *",
    "from employees where dept == \"IT\" select name",
    "from employees where salary > 50000 select name orderby salary desc",
    "from employees top 10 select name",
}
```

### 2. **Sistema de Inventario**

```go
// Estructura Inventory
type InventoryItem struct {
    SKU        string  `linq:"sku"`
    Name       string  `linq:"name"`
    Category   string  `linq:"category"`
    Quantity   int     `linq:"qty"`
    Price      float64 `linq:"price"`
    Supplier   string  `linq:"supplier"`
}

// Consultas de negocio
queries := []string{
    "from inventory where qty < 10 select name",              // Stock bajo
    "from inventory where category == \"Electronics\" select name", // Por categoría
    "from inventory where price > 100 select name orderby price desc", // Productos caros
}
```

### 3. **Análisis de Ventas**

```go
// Estructura Sales
type SaleRecord struct {
    ID         int     `linq:"id"`
    Customer   string  `linq:"customer"`
    Product    string  `linq:"product"`
    Amount     float64 `linq:"amount"`
    Date       string  `linq:"date"`
    Region     string  `linq:"region"`
}

// Reportes dinámicos
queries := []string{
    "from sales where amount > 1000 select customer",
    "from sales where region == \"North\" select customer orderby amount desc",
    "from sales top 5 select customer",
}
```

### 4. **Sistema CRM**

```go
// Estructura Lead
type Lead struct {
    ID       int    `linq:"id"`
    Name     string `linq:"name"`
    Email    string `linq:"email"`
    Status   string `linq:"status"`
    Source   string `linq:"source"`
    Score    int    `linq:"score"`
}

// Análisis de leads
queries := []string{
    "from leads where status == \"qualified\" select name",
    "from leads where score > 80 select name orderby score desc",
    "from leads where source == \"website\" select name",
}
```

## 🚀 Extensiones Posibles

### 1. **Más Operadores**

```go
// Operadores adicionales
dsl.KeywordToken("LIKE", "like")        // Pattern matching
dsl.KeywordToken("IN", "in")            // Lista de valores
dsl.KeywordToken("BETWEEN", "between")  // Rangos
dsl.KeywordToken("NOT", "not")          // Negación

// Uso
"from people where name like \"Juan%\" select name"
"from products where category in [\"Electronics\", \"Furniture\"] select name"
"from orders where amount between 100 and 500 select customer"
```

### 2. **Funciones de Agregación**

```go
// Agregaciones
dsl.KeywordToken("COUNT", "count")
dsl.KeywordToken("SUM", "sum")
dsl.KeywordToken("AVG", "avg")
dsl.KeywordToken("MAX", "max")
dsl.KeywordToken("MIN", "min")

// Uso
"count from people where department == \"Engineering\""
"sum salary from people where department == \"Sales\""
"avg price from products where category == \"Electronics\""
```

### 3. **JOIN Entre Estructuras**

```go
// JOIN syntax
"from people p join orders o on p.id == o.customer_id select p.name, o.amount"
```

### 4. **GROUP BY**

```go
// Agrupación
"from sales group by region select region, count(*)"
"from people group by department select department, avg(salary)"
```

## 🎓 Lecciones Técnicas

### 1. **Reflexión es Poderosa**
Permite crear sistemas 100% genéricos que funcionan con cualquier estructura.

### 2. **Struct Tags para Flexibilidad**
`linq:"fieldname"` permite mapeo personalizado de campos.

### 3. **Compatibilidad hacia Atrás**
Estructuras sin tags siguen funcionando usando nombres de campo.

### 4. **Parsing Zero Errors**
Sistema robusto que no falla con diferentes tipos de estructuras.

## 🔗 Patrones Similares

- **Entity Framework LINQ**: LINQ to Entities en .NET
- **Hibernate Criteria**: Query dinámico en Java
- **SQLAlchemy ORM**: Query object en Python
- **Eloquent ORM**: Query builder en Laravel
- **Django ORM**: Query API en Django

## 🚀 Próximos Pasos

1. **Ejecuta el ejemplo**: `go run main.go`
2. **Define tus propias estructuras** con tags personalizados
3. **Experimenta con consultas complejas**
4. **Integra en tu sistema** de datos
5. **Extiende con nuevos operadores** según necesites

## 📞 Referencias y Documentación

- **Código fuente**: [`main.go`](main.go)
- **Motor LINQ**: [`dsllinq/`](dsllinq/)
- **Manual completo**: [Manual de Uso](../../docs/es/manual_de_uso.md)
- **Guía técnica**: [Developer Onboarding](../../docs/es/developer_onboarding.md)

---

**¡Demuestra que go-dsl puede crear un sistema LINQ universal 100% genérico!** 🔍🎉