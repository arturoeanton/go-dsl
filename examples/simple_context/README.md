# Simple Context - r2lang Equivalent

**Fundamental example demonstrating how go-dsl implements dynamic context equivalent to r2lang's `q.use()` method.**

## 🎯 Objective

This example demonstrates go-dsl's **dynamic context features**, showing:

- 🔄 Direct equivalence with r2lang's `q.use()`
- 📊 Access to context variables and data
- 📋 Array and complex structure processing
- 🔄 Two context management methods: `SetContext()` vs `Use()`
- 🎯 Aggregation operations (Count, Sum, List)

## 🚀 Quick Start

```bash
cd examples/simple_context
go run main.go
```

## 🔄 r2lang Equivalence

### Syntax Comparison

| r2lang | go-dsl |
|--------|--------|
| `q.use("get name", {name: "Juan"})` | `dsl.Use("get name", map[string]interface{}{"name": "Juan"})` |
| `context.name` | `dsl.GetContext("name")` |
| Automatic | Requires type assertion: `name.(string)` |

### Direct Example

```javascript
// r2lang
const result = q.use("get variable", {
    name: "Juan García",
    age: 30,
    city: "Madrid"
});

// go-dsl (exact equivalent)
result, err := dsl.Use("get variable", map[string]interface{}{
    "name": "Juan García",
    "age":  30,
    "city": "Madrid",
})
```

## 📚 DSL Features

### Defined Tokens

```go
// Basic commands
dsl.KeywordToken("GET", "get")           // retrieve
dsl.KeywordToken("COUNT", "count")       // count
dsl.KeywordToken("LIST", "list")         // list
dsl.KeywordToken("SUM", "sum")           // sum

// Specific fields
dsl.KeywordToken("USERS", "users")       // users
dsl.KeywordToken("AGES", "ages")         // ages
dsl.KeywordToken("ALL", "all")           // all

// Generic variables
dsl.Token("VARIABLE", "[a-zA-Z_][a-zA-Z0-9_]*") // variable names
```

### Supported Commands

#### 1. Simple Variable Access
```
get name      # Gets "name" value from context
get age       # Gets "age" value from context
get city      # Gets "city" value from context
get missing   # Handles non-existent variables
```

#### 2. Array Operations
```
count users   # Counts elements in "users" array
list all users # Lists all users
sum all ages   # Sums all values in "ages"  
count ages     # Counts elements in "ages"
```

#### 3. Dynamic Variables
```
get [any_variable]  # Dynamic access to variables
```

## 🏗️ Context Architecture

### Supported Data Types

```go
// Simple variables
context := map[string]interface{}{
    "name":  "Juan García",
    "age":   30,
    "city":  "Madrid",
    "score": 95.5,
}

// Simple arrays
context := map[string]interface{}{
    "users":  []string{"Juan", "María", "Carlos", "Ana"},
    "ages":   []int{28, 35, 42, 29},
    "scores": []float64{95.5, 87.2, 92.1, 88.9},
}

// Complex structures
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

### Processing Actions

```go
// Simple variable access
dsl.Action("getVariable", func(args []interface{}) (interface{}, error) {
    varName := args[1].(string)
    value := dsl.GetContext(varName)
    
    if value == nil {
        return fmt.Sprintf("Variable '%s' not found", varName), nil
    }
    return value, nil
})

// Array counting
dsl.Action("countUsers", func(args []interface{}) (interface{}, error) {
    users := dsl.GetContext("users")
    if userArray, ok := users.([]string); ok {
        return len(userArray), nil
    }
    return 0, nil
})

// Numeric value sum
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

## 🔄 Two Context Methods

### Method 1: Use() - Direct r2lang Equivalent

```go
// Similar to r2lang's q.use()
context := map[string]interface{}{
    "name": "Juan García",
    "age":  30,
    "city": "Madrid",
}

result, err := dsl.Use("get name", context)
// result.GetOutput() → "Juan García"
```

**Advantages:**
- Exact equivalence with r2lang
- Temporary context for one operation
- Ideal for frequently changing data

### Method 2: SetContext() - Persistent Context

```go
// Set persistent context
dsl.SetContext("user", "Alice")
dsl.SetContext("role", "admin")

// Use in multiple operations
result1, _ := dsl.Parse("get user")  // → "Alice"
result2, _ := dsl.Parse("get role")  // → "admin"
```

**Advantages:**
- Context persists between calls
- Less overhead for static data
- Ideal for global configuration

## 📊 Example Output

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

## 🎯 Practical Use Cases

### 1. **Application Configuration**
```go
// Global configuration
dsl.SetContext("environment", "production")
dsl.SetContext("debug", false)
dsl.SetContext("maxConnections", 100)

// Configuration commands
result, _ := dsl.Parse("get environment")
```

### 2. **Dynamic User Data**
```go
// Per user/session
userContext := map[string]interface{}{
    "userId":   12345,
    "username": "john.doe",
    "permissions": []string{"read", "write"},
}

result, _ := dsl.Use("check permissions", userContext)
```

### 3. **Report Processing**
```go
// Report data
reportContext := map[string]interface{}{
    "sales":     []float64{1000, 1500, 2000},
    "customers": []string{"A", "B", "C"},
    "period":    "2025-Q1",
}

totalSales, _ := dsl.Use("sum all sales", reportContext)
```

### 4. **Multi-Tenant Systems**
```go
// Per tenant
tenantContext := map[string]interface{}{
    "tenantId":   "company-123",
    "plan":       "enterprise",
    "features":   []string{"advanced", "api", "support"},
}

result, _ := dsl.Use("check features", tenantContext)
```

## 🔧 Technical Features

### 1. **Type Assertions Required**

```go
// go-dsl requires explicit type assertions
name := dsl.GetContext("name").(string)
age := dsl.GetContext("age").(int)

// Safe checking
if value := dsl.GetContext("optional"); value != nil {
    text := value.(string)
}
```

### 2. **Typed Array Handling**

```go
// Arrays require correct type assertion
users := dsl.GetContext("users")
if userArray, ok := users.([]string); ok {
    // Process userArray
}
```

### 3. **Immutable Context During Parse**

```go
// Context doesn't change during individual parse
result, _ := dsl.Use("command", context)
// context remains unchanged
```

## 🎓 Best Practices

### 1. **Use Use() for Dynamic Data**
```go
// ✅ Data that changes per operation
userCtx := getUserContext(userId)
result, _ := dsl.Use("process user", userCtx)
```

### 2. **Use SetContext() for Configuration**
```go
// ✅ Global/static configuration
dsl.SetContext("apiKey", config.ApiKey)
dsl.SetContext("version", "1.2.3")
```

### 3. **Validate Context in Actions**
```go
dsl.Action("safeAccess", func(args []interface{}) (interface{}, error) {
    value := dsl.GetContext("required")
    if value == nil {
        return nil, fmt.Errorf("required context missing")
    }
    return value, nil
})
```

## 🔗 Similar Use Cases

- **Template Systems**: Dynamic variables in templates
- **Rule Engines**: Rule evaluation context
- **Stateful APIs**: Session and user data
- **Configuration Systems**: Dynamic values per environment
- **ETL Processing**: Context data for transformations

## 🚀 Next Steps

1. **Run the example**: `go run main.go`
2. **Modify the context** in the code
3. **Add new variables** and commands
4. **Experiment with more complex** structures
5. **Combine with other examples** in the project

## 📞 References

- **r2lang comparison**: [r2lang Documentation](https://github.com/arturoeanton/r2lang)
- **Complete manual**: [Usage Manual](../../docs/es/manual_de_uso.md) (Spanish)
- **Advanced example**: [Multi-country system](../accounting/)
- **Main README**: [Project Overview](../../README.md)

---

**This example demonstrates that go-dsl is a complete and improved replacement for r2lang!** 🔄🎉