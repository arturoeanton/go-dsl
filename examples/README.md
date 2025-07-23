# go-dsl Examples

This directory contains various examples demonstrating the capabilities of go-dsl.

## Basic Examples

### [simple](simple/)
Basic DSL usage demonstrating tokens, rules, and actions.

### [simple_context](simple_context/)
Shows how to use context variables in your DSL, similar to r2lang's `q.use()` functionality.

### [context_demo](context_demo/)
Advanced context usage with the LINQ-like query DSL, demonstrating data filtering and selection.

## Domain-Specific Examples

### [calculator](calculator/)
A calculator DSL supporting arithmetic operations with proper operator precedence.

### [accounting](accounting/)
Multi-country accounting DSL with support for different Latin American tax systems (Mexico, Colombia, Argentina, Peru).

### [drools](drools/)
Business rules engine implementation inspired by Drools, demonstrating rule-based systems.

### [liveview](liveview/)
Phoenix LiveView-inspired DSL for building reactive web interfaces.

## Advanced Features

### [declarative](declarative/)
Demonstrates declarative DSL configuration using YAML/JSON files and the builder pattern API.

### [advanced_grammar](advanced_grammar/)
Showcases advanced grammar features:
- Operator precedence and associativity
- Kleene star (*) and plus (+) repetition rules
- Priority-based token matching

## Test Files

### [test_failing.go](test_failing.go)
Test cases demonstrating error handling and parsing failures.

## Running Examples

Each example can be run independently:

```bash
# Run a specific example
cd calculator
go run main.go

# Or run from the examples directory
go run calculator/main.go
```

## Learning Path

1. Start with `simple` to understand basic concepts
2. Move to `simple_context` to learn about context usage
3. Try `calculator` for a practical DSL implementation
4. Explore `declarative` for YAML/JSON configuration
5. Study `advanced_grammar` for complex language features
6. Look at domain-specific examples (`accounting`, `drools`) for real-world applications

## Common Patterns

### Token Definition
```go
dsl.Token("NUMBER", "[0-9]+")
dsl.KeywordToken("IF", "if")  // Higher priority
```

### Rule Definition
```go
dsl.Rule("expr", []string{"NUMBER", "PLUS", "NUMBER"}, "add")
dsl.RuleWithPrecedence("expr", []string{"expr", "PLUS", "term"}, "add", 1, "left")
```

### Action Implementation
```go
dsl.Action("add", func(args []interface{}) (interface{}, error) {
    left := toInt(args[0])
    right := toInt(args[2])
    return left + right, nil
})
```

### Context Usage
```go
result, err := dsl.Use("query", map[string]interface{}{
    "data": myData,
})
```

## Contributing

When adding new examples:
1. Create a new directory with a descriptive name
2. Include a `main.go` file with the example code
3. Add a `README.md` explaining the example
4. Update this file with a description of your example