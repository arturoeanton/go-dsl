# go-dsl

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/doc/install)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/arturoeanton/go-dsl)](https://goreportcard.com/report/github.com/arturoeanton/go-dsl)

**A powerful and flexible Domain Specific Language (DSL) builder for Go that enables you to create custom programming languages with enterprise-grade features.**

go-dsl allows you to quickly build domain-specific languages with custom syntax, grammar rules, and semantic actions. Perfect for business rules, accounting systems, query languages, calculators, and complex enterprise applications. **Now with full left-recursive grammar support and production-ready stability.**

## ✨ Features

- 🚀 **Dynamic DSL Creation**: Build custom languages at runtime
- 🎯 **Advanced Grammar System**: Full left-recursive grammar support with memoization
- 🔄 **Context Support**: Pass dynamic data like r2lang's `q.use()` method
- 🧠 **Production-Ready Parser**: Handles complex enterprise scenarios with stability
- 📝 **Rich Examples**: Accounting systems, multi-country tax calculations, LINQ-like syntax
- 🔧 **Easy Integration**: Simple API for embedding in your applications
- ⚡ **High Performance**: Efficient parsing with intelligent token prioritization
- 🌍 **Enterprise Features**: Multi-language support, complex business rules, tax calculations
- 🏗️ **Left-Recursive Rules**: Handle complex patterns like `movements -> movements movement`
- 🎨 **KeywordToken Priority**: Solve token conflicts with priority-based matching

## 🚀 Quick Start

### Installation

```bash
go get github.com/arturoeanton/go-dsl/pkg/dslbuilder
```

### Basic Example

```go
package main

import (
    "fmt"
    "log"
    "github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

func main() {
    // Create a new DSL
    dsl := dslbuilder.New("HelloDSL")
    
    // Define tokens
    dsl.KeywordToken("HELLO", "hello")
    dsl.KeywordToken("WORLD", "world")
    
    // Define grammar rule
    dsl.Rule("greeting", []string{"HELLO", "WORLD"}, "greet")
    
    // Define action
    dsl.Action("greet", func(args []interface{}) (interface{}, error) {
        return "Hello, World!", nil
    })
    
    // Parse and execute
    result, err := dsl.Parse("hello world")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(result.GetOutput()) // Output: Hello, World!
}
```

## 📚 Examples

### 1. Enterprise Accounting DSL (Production Ready)

Create a complete accounting system with tax calculations:

```go
accounting := dslbuilder.New("Accounting")

// Define tokens with KeywordToken for priority
accounting.KeywordToken("VENTA", "venta")
accounting.KeywordToken("DE", "de")
accounting.KeywordToken("CON", "con")
accounting.KeywordToken("IVA", "iva")
accounting.Token("IMPORTE", "[0-9]+\\.?[0-9]*")
accounting.Token("STRING", "\"[^\"]*\"")

// Left-recursive rules for complex entries
accounting.Rule("command", []string{"VENTA", "DE", "IMPORTE", "CON", "IVA"}, "saleWithTax")
accounting.Rule("command", []string{"VENTA", "DE", "IMPORTE"}, "simpleSale")
accounting.Rule("movements", []string{"movement"}, "singleMovement")
accounting.Rule("movements", []string{"movements", "movement"}, "multipleMovements") // Left-recursive!

// Actions with business logic
accounting.Action("saleWithTax", func(args []interface{}) (interface{}, error) {
    amount, _ := strconv.ParseFloat(args[2].(string), 64)
    tax := amount * 0.16 // 16% IVA
    return Transaction{Amount: amount, Tax: tax, Total: amount + tax}, nil
})

// Usage: Parse complex accounting entries
// "venta de 5000 con iva" → Transaction{Amount: 5000, Tax: 800, Total: 5800}
// "asiento debe 1101 10000 debe 1401 1600 haber 2101 11600" → Balanced accounting entry
```

### 2. Multi-Country Tax System DSL

Build a flexible tax calculation system:

```go
accounting := dslbuilder.New("TaxSystem")

// Define tokens with KeywordToken for priority
accounting.KeywordToken("REGISTRAR", "registrar")
accounting.KeywordToken("CREAR", "crear")
accounting.KeywordToken("VENTA", "venta")
accounting.KeywordToken("COMPRA", "compra")
accounting.KeywordToken("DE", "de")
accounting.KeywordToken("CON", "con")
accounting.KeywordToken("DESCRIPCION", "descripcion")

// Most specific rules first
accounting.Rule("transaction", []string{"REGISTRAR", "VENTA", "DE", "AMOUNT", "CON", "DESCRIPCION", "STRING"}, "fullTransaction")
accounting.Rule("transaction", []string{"REGISTRAR", "VENTA", "DE", "AMOUNT"}, "simpleTransaction")

// Multi-country tax calculation
calcIVA := func(amount float64, country string) float64 {
    rates := map[string]float64{"MX": 0.16, "COL": 0.19, "AR": 0.21, "PE": 0.18}
    return amount * rates[country]
}

// Usage with context for different countries
mexContext := map[string]interface{}{"country": "MX"}
result, _ := accounting.Use(`registrar venta de 5000 con descripcion "Laptops"`, mexContext)
// → Transaction with 16% Mexican IVA

colContext := map[string]interface{}{"country": "COL"}  
result, _ := accounting.Use(`crear compra de 3000`, colContext)
// → Transaction with 19% Colombian IVA
```

### 3. LINQ-like DSL with Advanced Context

Create powerful data querying with dynamic context:

```go
linq := dslbuilder.New("LINQ")

// Define comprehensive LINQ-style syntax
linq.KeywordToken("FROM", "from")
linq.KeywordToken("WHERE", "where") 
linq.KeywordToken("SELECT", "select")
linq.KeywordToken("NAME", "name")
linq.KeywordToken("AGE", "age")
linq.KeywordToken("CITY", "city")

// Advanced context-based data access (like r2lang's q.use())
people := []Person{
    {Name: "Juan García", Age: 28, City: "Madrid"},
    {Name: "María López", Age: 35, City: "Barcelona"},
    {Name: "Carlos Rodríguez", Age: 42, City: "Madrid"},
}

// Multiple contexts for different datasets
context1 := map[string]interface{}{"data": people}
context2 := map[string]interface{}{"users": people}

// Execute queries with dynamic context switching
result1, _ := linq.Use(`select name from data where age > 30`, context1)
result2, _ := linq.Use(`select city from users where city == Madrid`, context2)
// → Dynamic queries on different data sources
```

## 🎯 Use Cases

- **Configuration Languages**: Create domain-specific config file formats
- **Business Rules**: Build rule engines for complex business logic
- **Query Languages**: Develop custom query interfaces for your data
- **Calculators**: Build specialized calculation engines
- **Scripting**: Embed custom scripting capabilities in applications
- **Data Processing**: Create transformation languages for ETL pipelines

## 🏗️ Architecture

go-dsl consists of several key components:

- **Tokenizer**: Converts input text into tokens using regex patterns
- **Parser**: Processes tokens according to grammar rules with left-recursion support
- **Actions**: Execute semantic actions when grammar rules match
- **Context System**: Provides dynamic data access during parsing

### Key Concepts

1. **Tokens**: Define the vocabulary of your language using regex patterns
2. **Rules**: Specify how tokens combine to form valid expressions
3. **Actions**: Define what happens when rules are matched
4. **Context**: Pass dynamic data to your DSL operations

## 📖 Documentation

- [Quick Start Guide](docs/es/guia_rapida.md) (Spanish)
- [API Reference](pkg/dslbuilder/)
- [Examples](examples/)

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📜 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🔗 Related Projects

- [r2lang](https://github.com/arturoeanton/r2lang) - The inspiration for context functionality

## 👨‍💻 Author

**Arturo Elias Anton**
- GitHub: [@arturoeanton](https://github.com/arturoeanton)

---

⭐ If you find this project useful, please give it a star on GitHub!