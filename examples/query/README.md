# Query DSL System - SQL-like Queries in English

**SQL-like query system in English that demonstrates complex grammars, advanced filters, and enterprise data processing.**

## 🎯 Objective

This example demonstrates how to create an **enterprise query system** in English that includes:

- 🔍 Natural query syntax in English
- 📊 Complex filters with multiple conditions
- 🗂️ Processing of complex data structures
- 🎯 Aggregation operations (count, list, search)
- 🔄 Dynamic context for variable datasets
- 📝 LINQ/SQL-like queries in English

## 🚀 Quick Start

```bash
cd examples/query
go run main.go
```

## 📚 DSL Features

### Defined Tokens

```go
// Main commands
dsl.KeywordToken("LIST", "list")           // list
dsl.KeywordToken("SEARCH", "search")       // search
dsl.KeywordToken("COUNT", "count")         // count

// Entities
dsl.KeywordToken("PRODUCTS", "products")   // products

// Filters and conditions  
dsl.KeywordToken("WHERE", "where")         // where
dsl.KeywordToken("CATEGORY", "category")   // category
dsl.KeywordToken("PRICE", "price")         // price
dsl.KeywordToken("STOCK", "stock")         // stock
dsl.KeywordToken("NAME", "name")           // name

// Operators
dsl.KeywordToken("IS", "is")               // is
dsl.KeywordToken("GREATER", "greater")     // greater
dsl.KeywordToken("LESS", "less")           // less
dsl.KeywordToken("CONTAINS", "contains")   // contains

// Values
dsl.Token("STRING", "\"[^\"]*\"")          // strings
dsl.Token("NUMBER", "[0-9]+\\.?[0-9]*")   // numbers
dsl.Token("WORD", "[a-zA-Z]+")             // category names
```

### Supported Commands

#### 1. Basic Queries
```
list products                      # Lists all products
count products                     # Counts total products
```

#### 2. Category Filters
```
search products where category is Electronics
list products where category is Furniture
```

#### 3. Price Filters
```
list products where price greater 100
search products where price less 50
```

#### 4. Stock Filters
```
count products where stock less 10
list products where stock greater 20
```

#### 5. Name Search
```
search products where name contains "Desk"
list products where name contains "USB"
```

#### 6. Complex Combinations
Rules are organized from most specific to least specific for maximum flexibility.

## 🏗️ System Architecture

### Data Structure

```go
type Product struct {
    Name     string
    Category string
    Price    float64
    Stock    int
}

// Example dataset
products := []Product{
    {"Dell Laptop", "Electronics", 1200.00, 5},
    {"Logitech Mouse", "Electronics", 25.00, 50},
    {"Desk Chair", "Furniture", 350.00, 10},
    {"Standing Desk", "Furniture", 600.00, 3},
    {"USB Cable", "Electronics", 10.00, 100},
    {"27\" Monitor", "Electronics", 400.00, 8},
    {"Office Lamp", "Furniture", 45.00, 20},
}
```

### Implemented Filter Types

```go
type Filter struct {
    Field     string      // "category", "price", "stock", "name"
    Operator  string      // "is", "greater", "less", "contains"
    Value     interface{} // Value to compare
}

// Universal filtering function
func applyFilter(products []Product, filter Filter) []Product {
    var result []Product
    
    for _, product := range products {
        matches := false
        
        switch filter.Field {
        case "category":
            matches = (filter.Operator == "is" && product.Category == filter.Value.(string))
            
        case "price":
            switch filter.Operator {
            case "greater":
                matches = product.Price > filter.Value.(float64)
            case "less":
                matches = product.Price < filter.Value.(float64)
            }
            
        case "stock":
            switch filter.Operator {
            case "greater":
                matches = float64(product.Stock) > filter.Value.(float64)
            case "less":
                matches = float64(product.Stock) < filter.Value.(float64)
            }
            
        case "name":
            if filter.Operator == "contains" {
                matches = strings.Contains(product.Name, filter.Value.(string))
            }
        }
        
        if matches {
            result = append(result, product)
        }
    }
    
    return result
}
```

## 🔧 Advanced Technical Features

### 1. Rules Organized by Specificity

```go
// MORE specific first (longer patterns)
query.Rule("query", []string{"SEARCH", "PRODUCTS", "WHERE", "CATEGORY", "IS", "WORD"}, "filterByCategory")
query.Rule("query", []string{"LIST", "PRODUCTS", "WHERE", "PRICE", "GREATER", "NUMBER"}, "filterByPriceGreater")
query.Rule("query", []string{"COUNT", "PRODUCTS", "WHERE", "STOCK", "LESS", "NUMBER"}, "countByStockLess")
query.Rule("query", []string{"SEARCH", "PRODUCTS", "WHERE", "NAME", "CONTAINS", "STRING"}, "filterByNameContains")

// Less specific after (shorter patterns)
query.Rule("query", []string{"LIST", "PRODUCTS"}, "listAll")
query.Rule("query", []string{"COUNT", "PRODUCTS"}, "countAll")
```

**Why this order?**
- go-dsl tries rules in definition order
- More specific rules capture special cases
- General rules capture basic cases
- Prevents simple rules from "capturing" complex commands

### 2. Dynamic Context for Datasets

```go
// Context can change per query
contextDemo := map[string]interface{}{
    "expensiveProducts": getExpensiveProducts(),
    "lowStockProducts":  getLowStockProducts(),
}

// Same syntax, different data
result1, _ := query.Use("list products", context1)
result2, _ := query.Use("list products", context2)
```

### 3. Reusable Actions

```go
// Generic filtering action
query.Action("applyFilter", func(args []interface{}) (interface{}, error) {
    // Build filter from arguments
    filter := Filter{
        Field:    extractField(args),
        Operator: extractOperator(args),
        Value:    extractValue(args),
    }
    
    // Get products from context
    products := query.GetContext("products").([]Product)
    
    // Apply filter
    return applyFilter(products, filter), nil
})
```

## 📊 Example Output

```
Query DSL Demo
==============

Query: list products
Result: 7 products found
  - Dell Laptop (Electronics) $1200.00 [Stock: 5]
  - Logitech Mouse (Electronics) $25.00 [Stock: 50]
  - Desk Chair (Furniture) $350.00 [Stock: 10]
  - Standing Desk (Furniture) $600.00 [Stock: 3]
  - USB Cable (Electronics) $10.00 [Stock: 100]
  - 27" Monitor (Electronics) $400.00 [Stock: 8]
  - Office Lamp (Furniture) $45.00 [Stock: 20]

Query: search products where category is Electronics
Result: 4 products found
  - Dell Laptop (Electronics) $1200.00 [Stock: 5]
  - Logitech Mouse (Electronics) $25.00 [Stock: 50]
  - USB Cable (Electronics) $10.00 [Stock: 100]
  - 27" Monitor (Electronics) $400.00 [Stock: 8]

Query: list products where price greater 100
Result: 4 products found
  - Dell Laptop (Electronics) $1200.00 [Stock: 5]
  - Desk Chair (Furniture) $350.00 [Stock: 10]
  - Standing Desk (Furniture) $600.00 [Stock: 3]
  - 27" Monitor (Electronics) $400.00 [Stock: 8]

Query: search products where name contains "Desk"
Result: 2 products found
  - Desk Chair (Furniture) $350.00 [Stock: 10]
  - Standing Desk (Furniture) $600.00 [Stock: 3]
```

## 🎯 Enterprise Use Cases

### 1. **Inventory Systems**
```
list products where stock less 5          # Low stock products
count products where category is Critical  # Critical products
search products where price greater 1000  # Premium products
```

### 2. **Sales Analysis**
```
list sales where date greater "2025-01-01"  # Recent sales
count customers where city is "Madrid"      # Customers by city
search orders where status is Pending       # Pending orders
```

### 3. **Human Resources Management**
```
list employees where department is IT       # IT employees
count users where active is true           # Active users
search candidates where experience greater 5 # Senior candidates
```

### 4. **Financial Analysis**
```
list expenses where category is Marketing   # Marketing expenses
count invoices where status is Overdue     # Overdue invoices
search transactions where amount greater 10000 # Large transactions
```

## 🔧 Possible Extensions

### 1. **Additional Operators**
```go
// Ranges
dsl.KeywordToken("BETWEEN", "between")
// "price between 100 and 500"

// Multiple values  
dsl.KeywordToken("IN", "in")
// "category in Electronics,Furniture"

// Dates
dsl.KeywordToken("FROM", "from")
dsl.KeywordToken("TO", "to")
// "date from 2025-01-01 to 2025-01-31"
```

### 2. **Advanced Aggregations**
```go
// Statistics
query.Rule("query", []string{"AVERAGE", "PRICE", "PRODUCTS"}, "averagePrice")
query.Rule("query", []string{"SUM", "STOCK", "PRODUCTS"}, "totalStock")
query.Rule("query", []string{"MAX", "PRICE", "PRODUCTS"}, "maxPrice")
```

### 3. **Sorting**
```go
// Order
dsl.KeywordToken("ORDER", "order")
dsl.KeywordToken("BY", "by")
dsl.KeywordToken("ASC", "asc")
dsl.KeywordToken("DESC", "desc")
// "list products order by price desc"
```

### 4. **Grouping**
```go
// Groups
dsl.KeywordToken("GROUP", "group")
// "count products group by category"
```

## 🎓 Technical Lessons

### 1. **KeywordToken Avoids Conflicts**
Without KeywordToken, words like "products" could be captured by generic patterns.

### 2. **Rule Order is Critical**
More specific rules should be defined first to prevent general ones from capturing them.

### 3. **Context Enables Flexibility**
The same DSL can work with different datasets depending on context.

### 4. **Reusable Filters**
Centralized filtering logic allows easy extension and maintenance.

## 🔗 Similar Patterns

- **Search Engines**: Natural language queries
- **BI Systems**: Ad-hoc queries for analysis
- **Filtering APIs**: Complex queries in REST APIs
- **Reporting**: Dynamic report generation
- **Data Discovery**: Exploration of large datasets

## 🚀 Next Steps

1. **Run the example**: `go run main.go`
2. **Modify the queries** in the code
3. **Add new operators** and filters
4. **Experiment with different datasets**
5. **Combine with dynamic context** for real cases

## 📞 References and Documentation

- **Source code**: [`main.go`](main.go)
- **Dynamic context**: [Simple Context Example](../simple_context/)
- **Complete manual**: [Usage Manual](../../docs/es/manual_de_uso.md) (Spanish)
- **Contributor guide**: [Developer Onboarding](../../docs/es/developer_onboarding.md) (Spanish)

---

**Demonstrates that go-dsl can create natural query interfaces in English for enterprise systems!** 🔍🎉