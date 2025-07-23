# Universal LINQ DSL - LINQ System in English

**100% generic LINQ system that works with ANY structure using reflection, zero parsing errors, compatible with struct tags and without tags.**

## 🎯 Objective

This example demonstrates a **universal LINQ system** that includes:

- 🔄 100% generic LINQ via reflection - works with ANY structure
- 📊 ZERO parsing errors - completely universal
- 🏷️ Support for struct tags `linq:"fieldname"`
- 🔄 Backward compatibility with structures without tags
- 🎯 SELECT, WHERE, ORDERBY, TOP operations
- 📝 Natural SQL/LINQ-like syntax
- 🚀 Production ready - unlimited reusability

## 🚀 Quick Start

```bash
cd examples/linq
go run main.go
```

## ✨ Universal Features

### 1. **100% Generic - Any Structure**

```go
// ✅ Works with Person
type Person struct {
    ID         int     `linq:"id"`
    Name       string  `linq:"name"`
    Age        int     `linq:"age"`
    Department string  `linq:"department"`
    Salary     float64 `linq:"salary"`
}

// ✅ Works with Product
type Product struct {
    ID       int     `linq:"id"`
    Name     string  `linq:"name"`
    Category string  `linq:"category"`
    Price    float64 `linq:"price"`
    Stock    int     `linq:"stock"`
}

// ✅ Works with Customer (NO tags - compatible!)
type Customer struct {
    ID    int
    Name  string
    Email string
    Phone string
}
```

### 2. **Struct Tags Support**

```go
// With custom tags
type Order struct {
    ID       int     `linq:"id"`
    Customer string  `linq:"customer"`  // Field "Customer" → query "customer"
    Amount   float64 `linq:"amount"`
    Status   string  `linq:"status"`
    Date     string  `linq:"date"`
}
```

### 3. **Automatic Field Detection**

```go
// Engine automatically detects available fields
fields := queryEngine.GetFieldNames(people[0])
// With tags: ["id", "name", "age", "department", "salary"]
// Without tags: ["ID", "Name", "Email", "Phone"]
```

## 📚 Supported Query Syntax

### Basic Commands

```go
// DSL tokens
dsl.KeywordToken("FROM", "from")
dsl.KeywordToken("SELECT", "select")
dsl.KeywordToken("WHERE", "where")
dsl.KeywordToken("ORDERBY", "orderby")
dsl.KeywordToken("TOP", "top")
dsl.KeywordToken("ASC", "asc")
dsl.KeywordToken("DESC", "desc")
```

### 1. **SELECT Queries**

```sql
from people select *                    -- All fields
from people select name                 -- Specific field
from products select name               -- Works with any table/structure
```

### 2. **WHERE Filters**

```sql
from people where age > 30 select name
from people where department == "Engineering" select name
from products where price > 100 select name
from products where category == "Electronics" select name
```

### 3. **ORDERBY Sorting**

```sql
from people where salary > 50000 select name orderby salary desc
from products where stock < 20 select name orderby price desc
```

### 4. **TOP Limit**

```sql
from people top 3 select name
from products top 2 select name
from orders top 5 select customer
```

### 5. **Complex Combinations**

```sql
from people where age > 30 and department == "Engineering" select name orderby salary desc top 2
```

## 🏗️ Universal Architecture

### Generic Query Engine

```go
// QueryEngine - 100% generic using reflection
type QueryEngine struct{}

func (qe *QueryEngine) GetFieldNames(item interface{}) []string {
    v := reflect.ValueOf(item)
    t := reflect.TypeOf(item)
    
    var fields []string
    for i := 0; i < v.NumField(); i++ {
        field := t.Field(i)
        
        // Priority 1: struct tag "linq"
        if tag := field.Tag.Get("linq"); tag != "" {
            fields = append(fields, tag)
        } else {
            // Priority 2: field name (compatible)
            fields = append(fields, strings.ToLower(field.Name))
        }
    }
    
    return fields
}
```

### Universal Filtering

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

## 📊 Example Output

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
   Dell Laptop
   Logitech Mouse
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

## 🔧 Advanced Technical Features

### 1. **Reflection for Dynamic Types**

```go
// Get field value using reflection
func (qe *QueryEngine) getFieldValue(v reflect.Value, t reflect.Type, fieldName string) interface{} {
    for i := 0; i < v.NumField(); i++ {
        field := t.Field(i)
        
        // Search by tag first
        if tag := field.Tag.Get("linq"); tag == fieldName {
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

### 2. **Polymorphic Comparison**

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

### 3. **Generic Sorting**

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

## 🎯 Enterprise Use Cases

### 1. **Human Resources System**

```go
// Employee structure
type Employee struct {
    ID         int     `linq:"id"`
    Name       string  `linq:"name"`
    Department string  `linq:"dept"`
    Position   string  `linq:"position"`
    Salary     float64 `linq:"salary"`
    StartDate  string  `linq:"start_date"`
}

// Queries
queries := []string{
    "from employees select *",
    "from employees where dept == \"IT\" select name",
    "from employees where salary > 50000 select name orderby salary desc",
    "from employees top 10 select name",
}
```

### 2. **Inventory System**

```go
// Inventory structure
type InventoryItem struct {
    SKU        string  `linq:"sku"`
    Name       string  `linq:"name"`
    Category   string  `linq:"category"`
    Quantity   int     `linq:"qty"`
    Price      float64 `linq:"price"`
    Supplier   string  `linq:"supplier"`
}

// Business queries
queries := []string{
    "from inventory where qty < 10 select name",              // Low stock
    "from inventory where category == \"Electronics\" select name", // By category
    "from inventory where price > 100 select name orderby price desc", // Expensive products
}
```

### 3. **Sales Analysis**

```go
// Sales structure
type SaleRecord struct {
    ID         int     `linq:"id"`
    Customer   string  `linq:"customer"`
    Product    string  `linq:"product"`
    Amount     float64 `linq:"amount"`
    Date       string  `linq:"date"`
    Region     string  `linq:"region"`
}

// Dynamic reports
queries := []string{
    "from sales where amount > 1000 select customer",
    "from sales where region == \"North\" select customer orderby amount desc",
    "from sales top 5 select customer",
}
```

### 4. **CRM System**

```go
// Lead structure
type Lead struct {
    ID       int    `linq:"id"`
    Name     string `linq:"name"`
    Email    string `linq:"email"`
    Status   string `linq:"status"`
    Source   string `linq:"source"`
    Score    int    `linq:"score"`
}

// Lead analysis
queries := []string{
    "from leads where status == \"qualified\" select name",
    "from leads where score > 80 select name orderby score desc",
    "from leads where source == \"website\" select name",
}
```

## 🚀 Possible Extensions

### 1. **More Operators**

```go
// Additional operators
dsl.KeywordToken("LIKE", "like")        // Pattern matching
dsl.KeywordToken("IN", "in")            // List of values
dsl.KeywordToken("BETWEEN", "between")  // Ranges
dsl.KeywordToken("NOT", "not")          // Negation

// Usage
"from people where name like \"Juan%\" select name"
"from products where category in [\"Electronics\", \"Furniture\"] select name"
"from orders where amount between 100 and 500 select customer"
```

### 2. **Aggregation Functions**

```go
// Aggregations
dsl.KeywordToken("COUNT", "count")
dsl.KeywordToken("SUM", "sum")
dsl.KeywordToken("AVG", "avg")
dsl.KeywordToken("MAX", "max")
dsl.KeywordToken("MIN", "min")

// Usage
"count from people where department == \"Engineering\""
"sum salary from people where department == \"Sales\""
"avg price from products where category == \"Electronics\""
```

### 3. **JOIN Between Structures**

```go
// JOIN syntax
"from people p join orders o on p.id == o.customer_id select p.name, o.amount"
```

### 4. **GROUP BY**

```go
// Grouping
"from sales group by region select region, count(*)"
"from people group by department select department, avg(salary)"
```

## 🎓 Technical Lessons

### 1. **Reflection is Powerful**
Enables creating 100% generic systems that work with any structure.

### 2. **Struct Tags for Flexibility**
`linq:"fieldname"` allows custom field mapping.

### 3. **Backward Compatibility**
Structures without tags continue working using field names.

### 4. **Zero Parsing Errors**
Robust system that doesn't fail with different structure types.

## 🔗 Similar Patterns

- **Entity Framework LINQ**: LINQ to Entities in .NET
- **Hibernate Criteria**: Dynamic queries in Java
- **SQLAlchemy ORM**: Query objects in Python
- **Eloquent ORM**: Query builder in Laravel
- **Django ORM**: Query API in Django

## 🚀 Next Steps

1. **Run the example**: `go run main.go`
2. **Define your own structures** with custom tags
3. **Experiment with complex queries**
4. **Integrate into your data system**
5. **Extend with new operators** as needed

## 📞 References and Documentation

- **Source code**: [`main.go`](main.go)
- **LINQ engine**: [`dsllinq/`](dsllinq/)
- **Complete manual**: [Usage Manual](../../docs/es/manual_de_uso.md) (Spanish)
- **Technical guide**: [Developer Onboarding](../../docs/es/developer_onboarding.md) (Spanish)

---

**Demonstrates that go-dsl can create a 100% generic universal LINQ system!** 🔍🎉