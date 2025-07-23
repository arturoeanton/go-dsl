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
- 🔨 **Builder Pattern API**: Fluent interface for DSL construction
- 📄 **Declarative Syntax**: Define DSLs using YAML/JSON configuration files
- 🛠️ **Developer Tools**: AST viewer, grammar validator, and interactive REPL
- 🎚️ **Operator Precedence**: Configurable precedence and associativity for operators
- 🔁 **Repetition Rules**: Kleene star (*) and plus (+) for zero/one or more patterns
- 🎯 **Advanced Grammar**: Support for complex language constructs and patterns

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
    // Option 1: Traditional API
    dsl := dslbuilder.New("HelloDSL")
    dsl.KeywordToken("HELLO", "hello")
    dsl.KeywordToken("WORLD", "world")
    
    // Option 2: Fluent Builder API
    dsl = dslbuilder.New("HelloDSL").
        WithKeywordToken("HELLO", "hello").
        WithKeywordToken("WORLD", "world").
        WithRule("greeting", []string{"HELLO", "WORLD"}, "greet").
        WithAction("greet", func(args []interface{}) (interface{}, error) {
            return "Hello, World!", nil
        })
        
    // Option 3: Load from YAML
    dsl, _ = dslbuilder.LoadFromYAMLFile("hello.yaml")
    
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

### 4. Declarative DSL Definition

Define your DSL using YAML or JSON:

```yaml
# calculator.yaml
name: "Calculator"
tokens:
  NUMBER: "[0-9]+"
  PLUS: "+"
  MINUS: "-"
  MULTIPLY: "*"
  DIVIDE: "/"
rules:
  - name: "expr"
    pattern: ["NUMBER", "PLUS", "NUMBER"]
    action: "add"
  - name: "expr"
    pattern: ["NUMBER", "MINUS", "NUMBER"]
    action: "subtract"
```

```go
// Load DSL from YAML
calcDSL, _ := dslbuilder.LoadFromYAMLFile("calculator.yaml")

// Register actions
calcDSL.Action("add", func(args []interface{}) (interface{}, error) {
    // Implementation
})

// Export DSL to JSON
calcDSL.SaveToJSONFile("calculator.json")
```

### 5. Advanced Grammar Features

go-dsl now supports advanced grammar features for building sophisticated DSLs:

#### Operator Precedence and Associativity

```go
// Define rules with precedence (higher number = higher priority)
calc := dslbuilder.New("Calculator")

// Level 1: Addition/Subtraction (lowest precedence, left associative)
calc.RuleWithPrecedence("expr", []string{"expr", "PLUS", "term"}, "add", 1, "left")
calc.RuleWithPrecedence("expr", []string{"expr", "MINUS", "term"}, "subtract", 1, "left")

// Level 2: Multiplication/Division (medium precedence, left associative)
calc.RuleWithPrecedence("term", []string{"term", "MULTIPLY", "factor"}, "multiply", 2, "left")
calc.RuleWithPrecedence("term", []string{"term", "DIVIDE", "factor"}, "divide", 2, "left")

// Level 3: Exponentiation (highest precedence, right associative)
calc.RuleWithPrecedence("factor", []string{"base", "POWER", "factor"}, "power", 3, "right")

// Result: "2 + 3 * 4" = 14 (not 20)
// Result: "2 ^ 3 ^ 2" = 512 (right associative: 2^(3^2))
```

#### Repetition Rules (Kleene Star/Plus)

```go
// Kleene Star (*) - Zero or more repetitions
list := dslbuilder.New("ListDSL")
list.RuleWithRepetition("items", "item", "items")  // items -> ε | items item

// Kleene Plus (+) - One or more repetitions  
list.RuleWithPlusRepetition("identifiers", "ID", "ids")  // ids -> ID | ids ID

// Example: Parse "a b c d" as a list of identifiers
```

#### Priority-Based Token Matching

```go
// Keywords have higher priority than generic identifiers
lang := dslbuilder.New("Language")
lang.KeywordToken("IF", "if")        // Priority: 90
lang.KeywordToken("WHILE", "while")  // Priority: 90
lang.Token("ID", "[a-zA-Z]+")        // Priority: 0

// "if" matches as IF token, not ID
// "ifx" matches as ID token
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
- **Builder API**: Fluent interface for DSL construction
- **Declarative Loader**: YAML/JSON configuration support

### Key Concepts

1. **Tokens**: Define the vocabulary of your language using regex patterns
2. **Rules**: Specify how tokens combine to form valid expressions
3. **Actions**: Define what happens when rules are matched
4. **Context**: Pass dynamic data to your DSL operations
5. **Builder Pattern**: Chain methods for fluent DSL construction
6. **Declarative Syntax**: Define DSLs externally in YAML/JSON

## 🛠️ Command-Line Tools

go-dsl includes powerful command-line tools to help you develop and debug your DSLs:

### AST Viewer
Visualize the Abstract Syntax Tree of your DSL parsing results:

```bash
# Install
go install github.com/arturoeanton/go-dsl/cmd/ast_viewer@latest

# Usage
ast_viewer -dsl calculator.yaml -input "10 + 20 * 30" -format tree
```

### Grammar Validator
Validate your DSL grammar and detect potential issues:

```bash
# Install
go install github.com/arturoeanton/go-dsl/cmd/validator@latest

# Usage
validator -dsl mydsl.yaml -verbose -info
```

### Interactive REPL
Test and explore your DSL interactively:

```bash
# Install
go install github.com/arturoeanton/go-dsl/cmd/repl@latest

# Usage
repl -dsl calculator.yaml -context data.json
```

See the [cmd/](cmd/) directory for detailed documentation of each tool.

## 📖 Documentation

### English
- [API Reference](pkg/dslbuilder/)
- [Examples](examples/)
- [Command-Line Tools](cmd/)

### Español
- [Guía Rápida](docs/es/guia_rapida.md) - Introducción completa y ejemplos
- [Conceptos Avanzados de DSL](docs/es/introduccion_dsl_segunda_parte.md) - Teoría de gramáticas y conceptos avanzados
- [Limitaciones](docs/es/limitaciones.md) - Limitaciones conocidas y soluciones
- [Propuesta de Mejoras](docs/es/propuesta_de_mejoras.md) - Roadmap y mejoras implementadas

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

- [r2lang](https://github.com/arturoeanton/go-r2lang) - The inspiration for context functionality

## 👨‍💻 Author

**Arturo Elias Anton**
- GitHub: [@arturoeanton](https://github.com/arturoeanton)

---

⭐ If you find this project useful, please give it a star on GitHub!