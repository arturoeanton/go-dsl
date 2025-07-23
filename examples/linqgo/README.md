# LinqGo - Universal LINQ Engine for Go

A complete and universal LINQ engine for Go, **100% compatible with .NET LINQ**, working with both Go structs and `map[string]interface{}` using reflection.

## 🎯 Features

- **100% .NET LINQ Compatible**: Identical syntax and operations to .NET LINQ
- **Dual Data Support**: Works with Go structs and `map[string]interface{}`
- **English-Only DSL**: Simplified with comprehensive English query syntax
- **Fluent API**: Chaining syntax like .NET LINQ
- **Complete Operations**: Where, Select, OrderBy, GroupBy, Take, Skip, etc.
- **Aggregations**: Count, Sum, Average, Min, Max
- **Set Operations**: Union, Intersect, Except
- **Quantifiers**: Any, All, Contains
- **DSL Queries**: SQL-like syntax for text queries
- **High Performance**: In-memory execution with optimized reflection

## 🚀 Usage

```bash
cd examples/linqgo
go run main.go
```

## 📖 DSL Query Syntax

### Basic Query Patterns
```sql
from ENTITY select FIELD
from ENTITY select *
from ENTITY where FIELD > VALUE select FIELD
from ENTITY where FIELD > VALUE select *
from ENTITY order by FIELD select FIELD
from ENTITY order by FIELD desc select FIELD
from ENTITY where FIELD > VALUE order by FIELD select FIELD
from ENTITY group by FIELD
from ENTITY group by FIELD select key
from ENTITY group by FIELD select count
```

## 🔧 DSL Query Examples

### Basic Selection
```sql
from employee select name
from customer select *
from product order by price select name
```

### Filtering
```sql
from employee where salary > 70000 select name
from customer where balance > 15000 select name
from product where price < 500 select name
from employee where department == Engineering select *
from customer where age >= 30 select name
from product where stock < 20 select name
```

### Ordering
```sql
from employee order by salary select name
from customer order by balance desc select name
from product order by price desc select name
from employee where salary > 60000 order by age select name
```

### Aggregations
```sql
from employee count
from employee sum salary
from customer avg balance
from product min price
from order max amount
from employee where department == Engineering sum salary
from customer where category == Premium avg balance
```

### Grouping
```sql
from employee group by department
from customer group by country select key
from order group by status select count
from product group by category select key
```

### Pagination
```sql
from employee take 5 select name
from customer skip 3 select name
from product skip 2 take 5 select *
from employee order by salary desc take 10 select name
```

### Distinct Queries
```sql
from employee select distinct department
from customer select distinct country
from product distinct
from employee select distinct position
```

### First/Last
```sql
from employee first
from customer last
from product where price > 1000 first
from employee where department == Engineering first
```

### Complex Combined Queries
```sql
from employee where salary > 70000 order by salary desc select name
from customer where balance > 10000 order by balance desc take 5 select *
from product where price < 1000 order by rating desc select name
from employee where age > 30 order by salary desc take 3 select name
```

## 📊 Fluent API (Programmatic)

### Examples with Go Structs
```go
// Import
import "github.com/arturoeliasanton/go-dsl/examples/linqgo/universal"

// Define struct
type Employee struct {
    ID         int     `linq:"id"`
    Name       string  `linq:"name"`
    Department string  `linq:"department"`
    Salary     float64 `linq:"salary"`
    Age        int     `linq:"age"`
}

// Use fluent LINQ
employees := []*Employee{...}

// Complex chained query
highEarners := universal.From(employees).
    WhereField("salary", ">", 70000).
    OrderByFieldDescending("salary").
    SelectField("name").
    Take(3).
    ToSlice()

// Aggregations
avgSalary := universal.From(employees).
    WhereField("department", "==", "Engineering").
    AverageField("salary")

// Grouping
groupedByDept := universal.From(employees).
    GroupByField("department")
```

### Examples with map[string]interface{}
```go
// Data as maps
projects := []map[string]interface{}{
    {"id": 1, "name": "Alpha", "budget": 100000.0, "status": "Active"},
    {"id": 2, "name": "Beta", "budget": 75000.0, "status": "Completed"},
}

// Convert to interface{}
var projectsInterface []interface{}
for _, project := range projects {
    projectsInterface = append(projectsInterface, project)
}

// Use LINQ
activeProjects := universal.From(projectsInterface).
    WhereField("status", "==", "Active").
    SumField("budget")
```

## 🎭 Complete Supported Operations

### Filtering Operations
- **Where** / **WhereField** - Filter elements
- **Take** - Take first N elements
- **Skip** - Skip first N elements
- **TakeWhile** - Take while condition is true
- **SkipWhile** - Skip while condition is true
- **Distinct** / **DistinctBy** / **DistinctByField** - Unique elements

### Projection Operations
- **Select** / **SelectField** / **SelectFields** - Select/transform elements

### Sorting Operations
- **OrderBy** / **OrderByField** - Sort ascending
- **OrderByDescending** / **OrderByFieldDescending** - Sort descending
- **Reverse** - Reverse order

### Grouping Operations
- **GroupBy** / **GroupByField** - Group elements

### Set Operations
- **Union** - Union of two sequences (no duplicates)
- **Intersect** - Intersection of two sequences
- **Except** - Difference of two sequences

### Aggregation Operations
- **Count** / **CountWhere** - Count elements
- **Sum** / **SumField** - Sum numeric values
- **Average** / **AverageField** - Calculate average
- **Min** / **MinField** - Find minimum
- **Max** / **MaxField** - Find maximum
- **Aggregate** - Custom aggregation

### Quantification Operations
- **Any** - Does any element meet condition?
- **All** - Do all elements meet condition?
- **Contains** - Contains specific element?

### Element Operations
- **First** / **FirstWhere** / **FirstOrDefault** - First element
- **Last** / **LastWhere** / **LastOrDefault** - Last element
- **Single** / **SingleOrDefault** - Single element

## 🏗️ Supported Data Types

### Go Structs with Tags
```go
type Customer struct {
    ID       int     `linq:"id"`
    Name     string  `linq:"name"`
    Email    string  `linq:"email"`
    Balance  float64 `linq:"balance"`
    Category string  `linq:"category"`
}
```

### Interface Maps
```go
data := []map[string]interface{}{
    {"id": 1, "name": "John", "salary": 75000.0},
    {"id": 2, "name": "Jane", "salary": 85000.0},
}
```

### Any Slice Type
```go
// LinqGo works with any []interface{}
var anyData []interface{}
anyData = append(anyData, customer1, customer2, customer3)

result := universal.From(anyData).
    WhereField("category", "==", "Premium").
    ToSlice()
```

## ⚙️ Supported Operators

### Comparison Operators
- `==`, `equals`, `eq` - Equality
- `!=`, `not_equals`, `ne` - Inequality
- `>`, `greater`, `gt` - Greater than
- `>=`, `greater_equal`, `ge` - Greater or equal
- `<`, `less`, `lt` - Less than
- `<=`, `less_equal`, `le` - Less or equal

### Text Operators
- `contains` - Contains text
- `starts_with` - Starts with
- `ends_with` - Ends with

## 🎯 Enterprise Use Cases

- **Data Analysis**: Process large enterprise datasets
- **Reporting**: Generate complex reports with aggregations
- **REST APIs**: Filter and paginate API results
- **Business Intelligence**: Business data analysis
- **ETL Processes**: Data transformation between systems
- **Data Mining**: Data mining with complex operations
- **Dashboards**: Prepare data for visualizations
- **Microservices**: Data processing between services

## 🚀 Performance and Features

### Performance Advantages
- **In-Memory Execution**: All operations execute in-memory
- **Optimized Reflection**: Efficient use of Go reflection
- **Lazy Evaluation**: Lazy evaluation where possible
- **Zero Dependencies**: Only depends on go-dsl

### Enterprise Features
- ✅ **Type Safety** - Type safety with error handling
- ✅ **Thread Safe** - Safe for concurrent use
- ✅ **Memory Efficient** - Efficient memory usage
- ✅ **Error Handling** - Robust error handling
- ✅ **Extensible** - Easy to extend with new operations
- ✅ **Production Ready** - Ready for production

## 🌟 Comparison with .NET LINQ

| Feature | .NET LINQ | LinqGo | Status |
|---------|-----------|--------|--------|
| Where | ✅ | ✅ | Complete |
| Select | ✅ | ✅ | Complete |
| OrderBy | ✅ | ✅ | Complete |
| GroupBy | ✅ | ✅ | Complete |
| Take/Skip | ✅ | ✅ | Complete |
| Distinct | ✅ | ✅ | Complete |
| Union/Intersect | ✅ | ✅ | Complete |
| Any/All | ✅ | ✅ | Complete |
| Count/Sum/Avg | ✅ | ✅ | Complete |
| First/Last | ✅ | ✅ | Complete |
| Aggregate | ✅ | ✅ | Complete |
| Join | ✅ | 🚧 | In development |
| DSL Syntax | ❌ | ✅ | LinqGo advantage |

## 📈 Performance Examples

```go
// Processing 10,000 employees
employees := make([]*Employee, 10000)
// ... fill data

// Complex query in one line
result := universal.From(employees).
    WhereField("department", "==", "Engineering").
    WhereField("salary", ">", 70000).
    OrderByFieldDescending("salary").
    Take(100).
    SelectFields("name", "salary", "department").
    ToSlice()

// Statistics by department
stats := universal.From(employees).
    GroupByField("department")

for _, group := range stats {
    avgSalary := universal.From(group.Items).AverageField("salary")
    fmt.Printf("%s: %d employees, avg salary: %.2f\n", 
        group.Key, group.Count, avgSalary)
}
```

The most complete and competitive LINQ engine for Go! 🚀