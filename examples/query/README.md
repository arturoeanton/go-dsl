# Universal Query DSL System - Query in English

**Universal query system 100% generic that works with ANY structure using reflection, bilingual support (Spanish/English), and zero parsing errors.**

## 🎯 Objective

This example demonstrates a **universal enterprise query system** that includes:

- 🔄 100% generic system via reflection - works with ANY structure
- 🌍 Complete bilingual support (Spanish and English simultaneously)  
- 📊 ZERO parsing errors - completely universal
- 🏷️ Support for struct tags `query:"fieldname"`
- 🔄 Backward compatibility with structures without tags
- 🎯 Advanced operations (search, list, count with filters)
- 📝 Natural SQL-like syntax in Spanish and English
- 🚀 Production ready - unlimited reusability

## 🚀 Quick Start

```bash
cd examples/query
go run main.go
```

## ✨ Universal Features

### 1. **100% Generic - Any Structure**

```go
// ✅ Works with Product
type Product struct {
    ID       int     `query:"id"`
    Name     string  `query:"nombre"`
    Category string  `query:"categoria"`
    Price    float64 `query:"precio"`
    Stock    int     `query:"stock"`
}

// ✅ Works with Employee
type Employee struct {
    ID         int     `query:"id"`
    Name       string  `query:"nombre"`
    Department string  `query:"departamento"`
    Position   string  `query:"posicion"`
    Salary     float64 `query:"salario"`
    Age        int     `query:"edad"`
}

// ✅ Works with Customer (NO tags - compatible!)
type Customer struct {
    ID    int
    Name  string
    Email string
    Phone string
}
```

### 2. **Complete Bilingual Support**

```go
// Same data, both languages work:
"listar productos donde precio mayor 100"           // Spanish
"list productos where precio greater 100"           // English
"buscar empleados donde departamento es Engineering" // Spanish  
"search empleados where departamento is Engineering" // English
```

### 3. **Automatic Field Detection**

```go
// Engine automatically detects available fields
fields := engine.GetFieldNames(products[0])
// With tags: ["id", "nombre", "categoria", "precio", "stock"]
// Without tags: ["id", "name", "email", "phone"]
```

## 📚 Supported Query Syntax

### Basic Commands (Bilingual)

```go
// Spanish
dsl.KeywordToken("BUSCAR", "buscar")      // search
dsl.KeywordToken("LISTAR", "listar")      // list  
dsl.KeywordToken("CONTAR", "contar")      // count

// English  
dsl.KeywordToken("SEARCH", "search")
dsl.KeywordToken("LIST", "list")
dsl.KeywordToken("COUNT", "count")

// Operators (Spanish)
dsl.KeywordToken("DONDE", "donde")        // where
dsl.KeywordToken("ES", "es")              // is
dsl.KeywordToken("MAYOR", "mayor")        // greater
dsl.KeywordToken("MENOR", "menor")        // less
dsl.KeywordToken("CONTIENE", "contiene")  // contains

// Operators (English)
dsl.KeywordToken("WHERE", "where")
dsl.KeywordToken("IS", "is") 
dsl.KeywordToken("GREATER", "greater")
dsl.KeywordToken("LESS", "less")
dsl.KeywordToken("CONTAINS", "contains")
```

### 1. **Simple Queries (Bilingual)**

```sql
-- Spanish
listar productos                    -- Lists all products
contar empleados                   -- Counts all employees
buscar pedidos                     -- Searches all orders

-- English  
list productos                     -- Lists all products
count empleados                    -- Counts all employees
search pedidos                     -- Searches all orders
```

### 2. **String Filters (Bilingual)**

```sql
-- Spanish
buscar productos donde categoria es "Electronics"
listar empleados donde nombre contiene "García"
contar pedidos donde estado es "completado"

-- English
search productos where categoria is "Electronics"  
list empleados where nombre contains "García"
count pedidos where estado is "completado"
```

### 3. **Numeric Filters (Bilingual)**

```sql
-- Spanish
listar productos donde precio mayor 100
contar empleados donde salario menor 50000
buscar pedidos donde monto mayor 1000

-- English
list productos where precio greater 100
count empleados where salario less 50000  
search pedidos where monto greater 1000
```

### 4. **Content Filters (Bilingual)**

```sql
-- Spanish
buscar productos donde nombre contiene "USB"
listar empleados donde posicion contiene "Developer"
contar clientes donde email contiene "@gmail.com"

-- English
search productos where nombre contains "USB"
list empleados where posicion contains "Developer"
count clientes where email contains "@gmail.com"
```

## 🏗️ Universal Architecture

### Generic Query Engine

```go
// UniversalQueryEngine - 100% generic using reflection
type UniversalQueryEngine struct{}

func (uqe *UniversalQueryEngine) GetFieldNames(item interface{}) []string {
    v := reflect.ValueOf(item)
    t := reflect.TypeOf(item)
    
    // Handle pointer types
    if v.Kind() == reflect.Ptr {
        v = v.Elem()
        t = t.Elem()
    }
    
    var fields []string
    for i := 0; i < v.NumField(); i++ {
        field := t.Field(i)
        
        // Priority 1: struct tag "query"
        if tag := field.Tag.Get("query"); tag != "" {
            fields = append(fields, tag)
        } else {
            // Priority 2: lowercase field name (compatible)
            fields = append(fields, strings.ToLower(field.Name))
        }
    }
    
    return fields
}
```

### Universal Filtering

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

### Polymorphic Comparison

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

## 📊 Example Output

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
   Result: 7 elements found
     - id: 1, nombre: Laptop Dell, categoria: Electronics, precio: 1200, stock: 5
     - id: 2, nombre: Mouse Logitech, categoria: Electronics, precio: 25, stock: 50
     - id: 3, nombre: Desk Chair, categoria: Furniture, precio: 350, stock: 10
     ... and 4 more

2. Query: buscar productos donde categoria es "Electronics"  
   Result: 4 elements found
     - id: 1, nombre: Laptop Dell, categoria: Electronics, precio: 1200, stock: 5
     - id: 2, nombre: Mouse Logitech, categoria: Electronics, precio: 25, stock: 50
     - id: 5, nombre: USB Cable, categoria: Electronics, precio: 10, stock: 100
     ... and 1 more

3. Query: list productos where categoria is "Furniture"
   Result: 3 elements found
     - id: 3, nombre: Desk Chair, categoria: Furniture, precio: 350, stock: 10
     - id: 4, nombre: Standing Desk, categoria: Furniture, precio: 600, stock: 3
     - id: 7, nombre: Office Lamp, categoria: Furniture, precio: 45, stock: 20

=== Testing with Employees Data ===
Available fields: [id nombre departamento posicion salario edad]

1. Query: buscar empleados donde departamento es "Engineering"
   Result: 3 elements found
     - id: 1, nombre: Juan García, departamento: Engineering, posicion: Senior Developer, salario: 75000, edad: 28
     - id: 3, nombre: Carlos Rodríguez, departamento: Engineering, posicion: Tech Lead, salario: 85000, edad: 42
     - id: 5, nombre: Pedro Sánchez, departamento: Engineering, posicion: Developer, salario: 55000, edad: 31

2. Query: list empleados where departamento is "Marketing"
   Result: 1 elements found
     - id: 2, nombre: María López, departamento: Marketing, posicion: Manager, salario: 65000, edad: 35

=== Testing with Customers Data (NO Tags - Backward Compatible!) ===
Available fields (fallback to field names): [id name email phone]

1. Query: buscar clientes donde name contiene "Alice"
   Result: 1 elements found
     - ID: 1, Name: Alice Brown, Email: alice@email.com, Phone: 555-0101

2. Query: list clientes where name is "Bob Smith"
   Result: 1 elements found
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

## 🔧 Advanced Technical Features

### 1. **Reflection for Dynamic Types**

```go
// Get field value using reflection
func (uqe *UniversalQueryEngine) GetFieldValue(item interface{}, fieldName string) interface{} {
    v := reflect.ValueOf(item)
    t := reflect.TypeOf(item)
    
    // Handle pointer types
    if v.Kind() == reflect.Ptr {
        v = v.Elem()
        t = t.Elem()
    }
    
    for i := 0; i < v.NumField(); i++ {
        field := t.Field(i)
        
        // Search by tag first
        if tag := field.Tag.Get("query"); tag == fieldName {
            return v.Field(i).Interface()
        }
        
        // Search by field name
        if strings.ToLower(field.Name) == fieldName {
            return v.Field(i).Interface()
        }
    }
    
    return nil
}
```

### 2. **Automatic Slice Conversion**

```go
// ConvertToInterfaceSlice converts any slice type to []interface{}
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

### 3. **Hierarchical Rule System**

```go
// Rules organized from most specific to least specific
func (uq *UniversalQueryDSL) setupRules() {
    // MOST specific first (6 tokens)
    uq.dsl.Rule("query", []string{"BUSCAR", "WORD", "DONDE", "WORD", "CONTIENE", "STRING"}, "filteredQueryStringContains")
    uq.dsl.Rule("query", []string{"SEARCH", "WORD", "WHERE", "WORD", "CONTAINS", "STRING"}, "filteredQueryStringContains")
    
    // Less specific after (2 tokens)
    uq.dsl.Rule("query", []string{"LISTAR", "WORD"}, "simpleQuery")
    uq.dsl.Rule("query", []string{"LIST", "WORD"}, "simpleQuery")
}
```

## 🎯 Enterprise Use Cases

### 1. **Human Resources System**

```go
// Universal Employee structure
type Employee struct {
    ID         int     `query:"id"`
    Name       string  `query:"nombre"`
    Department string  `query:"departamento"`
    Position   string  `query:"posicion"`
    Salary     float64 `query:"salario"`
    StartDate  string  `query:"fecha_inicio"`
}

// Bilingual queries
queries := []string{
    "listar empleados donde salario mayor 60000",
    "list empleados where departamento is Engineering",
    "contar empleados donde edad menor 30",
    "search empleados where posicion contains Manager",
}
```

### 2. **Inventory System**

```go
// Universal Product structure
type Product struct {
    SKU        string  `query:"sku"`
    Name       string  `query:"nombre"`
    Category   string  `query:"categoria"`
    Quantity   int     `query:"cantidad"`
    Price      float64 `query:"precio"`
    Supplier   string  `query:"proveedor"`
}

// Bilingual business queries
queries := []string{
    "buscar productos donde cantidad menor 10",              // Low stock
    "list productos where categoria is Electronics",         // By category
    "contar productos donde precio mayor 100",              // Expensive products
    "search productos where proveedor contains Samsung",     // By supplier
}
```

### 3. **Sales System**

```go
// Universal Sale structure
type Sale struct {
    ID         int     `query:"id"`
    Customer   string  `query:"cliente"`
    Product    string  `query:"producto"`
    Amount     float64 `query:"monto"`
    Date       string  `query:"fecha"`
    Region     string  `query:"region"`
}

// Bilingual dynamic reports
queries := []string{
    "listar ventas donde monto mayor 1000",
    "list ventas where region is North",
    "contar ventas donde fecha contiene 2025",
    "search ventas where cliente contains Garcia",
}
```

### 4. **CRM System**

```go
// Universal Lead structure
type Lead struct {
    ID       int    `query:"id"`
    Name     string `query:"nombre"`
    Email    string `query:"email"`
    Status   string `query:"estado"`
    Source   string `query:"origen"`
    Score    int    `query:"puntuacion"`
}

// Bilingual lead analysis
queries := []string{
    "buscar leads donde estado es qualified",
    "list leads where puntuacion mayor 80",
    "contar leads donde origen is website",
    "search leads where email contains gmail",
}
```

## 🚀 Possible Extensions

### 1. **More Operators**

```go
// Additional bilingual operators
dsl.KeywordToken("LIKE", "como")           // Pattern matching
dsl.KeywordToken("IN", "en")               // List of values
dsl.KeywordToken("BETWEEN", "entre")       // Ranges
dsl.KeywordToken("NOT", "no")              // Negation

// Bilingual usage
"buscar productos donde nombre como 'Dell%'"
"search products where nombre like 'Dell%'"
"listar pedidos donde monto entre 100 y 500"
"list orders where monto between 100 and 500"
```

### 2. **Aggregation Functions**

```go
// Bilingual aggregations
dsl.KeywordToken("SUMA", "suma")           // sum
dsl.KeywordToken("PROMEDIO", "promedio")   // average
dsl.KeywordToken("MAXIMO", "maximo")       // max
dsl.KeywordToken("MINIMO", "minimo")       // min

// Bilingual usage
"suma salario de empleados donde departamento es IT"
"sum salario from empleados where departamento is IT"
"promedio precio de productos donde categoria es Electronics"
"average precio from productos where categoria is Electronics"
```

### 3. **Sorting**

```go
// Bilingual sorting
dsl.KeywordToken("ORDENAR", "ordenar")     // order
dsl.KeywordToken("POR", "por")             // by
dsl.KeywordToken("ASC", "asc")             // ascending
dsl.KeywordToken("DESC", "desc")           // descending

// Bilingual usage
"listar productos ordenar por precio desc"
"list productos order by precio desc"
```

### 4. **Grouping**

```go
// Bilingual grouping
dsl.KeywordToken("AGRUPAR", "agrupar")     // group
dsl.KeywordToken("POR", "por")             // by

// Bilingual usage
"contar ventas agrupar por region"
"count ventas group by region"
```

## 🎓 Technical Lessons

### 1. **Reflection is Universal**
Enables creating systems that work with any structure without modifying code.

### 2. **Struct Tags for Localization**  
`query:"fieldname"` allows mapping fields to Spanish/English terms.

### 3. **Total Compatibility**
Structures without tags continue working using field names.

### 4. **Simultaneous Bilingualism**
Spanish and English coexist in the same engine without conflicts.

### 5. **Zero Parsing Errors**
Robust system that never fails with different structure types.

## 🔗 Similar Patterns

- **Entity Framework LINQ**: LINQ queries in .NET
- **Hibernate Criteria**: Dynamic queries in Java
- **SQLAlchemy ORM**: Query objects in Python
- **Eloquent ORM**: Query builder in Laravel
- **Django ORM**: Query API in Django
- **MongoDB Query Language**: NoSQL queries
- **GraphQL**: Flexible data queries

## 🚀 Next Steps

1. **Run the example**: `go run main.go`
2. **Define your own structures** with custom tags
3. **Experiment with bilingual queries**
4. **Integrate into your enterprise data system**
5. **Extend with new operators** as needed
6. **Combine Spanish and English** in the same system

## 📞 References and Documentation

- **Source code**: [`main.go`](main.go)
- **Universal engine**: [`universal/`](universal/)
- **LINQ system**: [../linq/](../linq/)
- **Dynamic context**: [../simple_context/](../simple_context/)
- **Complete manual**: [Usage Manual](../../docs/es/manual_de_uso.md) (Spanish)
- **Technical guide**: [Developer Onboarding](../../docs/es/developer_onboarding.md) (Spanish)

---

**Demonstrates that go-dsl can create universal query systems 100% generic with bilingual support!** 🔍🌍🎉