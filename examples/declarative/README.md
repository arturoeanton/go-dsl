# Declarative DSL Example

This example demonstrates the new declarative features added to go-dsl:
- **Builder Pattern** for fluent API
- **YAML/JSON configuration** for DSL definition
- **100% backward compatibility** with existing code

## Features Demonstrated

### 1. Builder Pattern (Fluent API)

```go
dsl := dslbuilder.New("Calculator").
    WithToken("NUMBER", "[0-9]+").
    WithToken("PLUS", "\\+").
    WithRule("expr", []string{"NUMBER", "PLUS", "NUMBER"}, "add").
    WithAction("add", func(args []interface{}) (interface{}, error) {
        // Implementation
    })
```

### 2. YAML Configuration

```yaml
name: "Calculator"
tokens:
  NUMBER: "[0-9]+"
  PLUS: "+"
  MINUS: "-"
rules:
  - name: "expr"
    pattern: ["NUMBER", "PLUS", "NUMBER"]
    action: "add"
```

### 3. JSON Export/Import

```go
// Save DSL to JSON
jsonData, _ := dsl.SaveToJSON()
dsl.SaveToJSONFile("calculator.json")

// Load DSL from JSON
loadedDSL, _ := dslbuilder.LoadFromJSONFile("calculator.json")
```

## Running the Example

```bash
cd examples/declarative
go run main.go
```

## Output

The example will:
1. Load a calculator DSL from `calculator.yaml`
2. Create the same DSL using the Builder Pattern
3. Export the DSL to JSON format
4. Test backward compatibility with traditional API

## Files

- `main.go` - Example implementation
- `calculator.yaml` - YAML DSL definition
- `calculator.json` - Generated JSON configuration (after running)

## Backward Compatibility

All existing code continues to work:

```go
// Traditional API still works
dsl := dslbuilder.New("Calculator")
dsl.Token("NUMBER", "[0-9]+")
dsl.Rule("expr", []string{"NUMBER", "PLUS", "NUMBER"}, "add")
dsl.Action("add", addFunction)
```