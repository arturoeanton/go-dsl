# Enhanced Error System Demo - Error Demo in English

**Example demonstrating go-dsl's enhanced error system with line and column information, maintaining 100% backward compatibility.**

## 🎯 Objective

This example demonstrates go-dsl's **enhanced error system**, showing:

- 🎯 Errors with specific line and column information
- 📊 Complete compatibility with existing code
- 🔧 `ParseError` type with detailed context
- 🔄 Helper functions for error handling
- 📝 Examples of different parsing error types

## 🚀 Quick Start

```bash
cd examples/error_demo
go run main.go
```

## ✨ New Error Features

### ParseError Type

```go
type ParseError struct {
    Message  string  // Descriptive error message
    Line     int     // Line number (1-based)
    Column   int     // Column number (1-based) 
    Position int     // Absolute position in input
    Token    string  // Token that caused the error
    Input    string  // Complete input for context
}
```

### Helper Functions

```go
// Check if an error is ParseError
func IsParseError(err error) bool

// Get detailed error information
func GetDetailedError(err error) string
```

## 📚 Backward Compatibility

### Existing Code Continues Working

```go
// ✅ Existing code - NO CHANGES
result, err := dsl.Parse("invalid command")
if err != nil {
    fmt.Println("Error:", err.Error())
    // Works exactly the same as before
}
```

### New Code Can Use Enhanced Features

```go
// ✅ New code - WITH enhanced features
result, err := dsl.Parse("invalid command")
if err != nil {
    if IsParseError(err) {
        parseErr := err.(*ParseError)
        fmt.Printf("Error at line %d, column %d: %s\n", 
                   parseErr.Line, parseErr.Column, parseErr.Message)
    } else {
        // Regular error (not parsing)
        fmt.Println("Error:", err.Error())
    }
}
```

## 🔧 Demonstrated Error Types

### 1. Unrecognized Token

```
Input: "xyz abc"
Error: Unrecognized token 'xyz' at line 1, column 1
Position: 0
```

### 2. Rule Not Found

```
Input: "HELLO world"
Error: No rule found matching tokens: [HELLO WORLD] at line 1, column 1
Position: 0
```

### 3. Empty Input

```
Input: ""
Error: Empty input at line 1, column 1
Position: 0
```

### 4. Whitespace Only

```
Input: "   "
Error: Empty input (whitespace only) at line 1, column 1
Position: 0
```

## 🏗️ Technical Implementation

### Demo DSL

```go
// Create simple DSL for demonstration
demo := dslbuilder.NewDSL("ErrorDemo")

// Basic tokens
demo.KeywordToken("HELLO", "hello")
demo.Token("WORLD", "world")
demo.Token("NUMBER", "[0-9]+")

// Simple rule
demo.Rule("greeting", []string{"HELLO", "WORLD"}, "sayHello")

// Action
demo.Action("sayHello", func(args []interface{}) (interface{}, error) {
    return "Hello World!", nil
})
```

### Test Cases

```go
testCases := []struct {
    input       string
    description string
}{
    {"hello world", "✅ Valid command"},
    {"xyz abc", "❌ Unrecognized token"},
    {"hello", "❌ Incomplete rule"},
    {"", "❌ Empty input"},
    {"   ", "❌ Whitespace only"},
    {"hello 123", "❌ Unexpected token"},
}
```

## 📊 Example Output

```
=== go-dsl Error Demo ===
Enhanced error system demonstration with line and column information

1. Valid command: 'hello world'
   ✅ Result: Hello World!

2. Unrecognized token: 'xyz abc'
   ❌ Error at line 1, column 1: Unrecognized token 'xyz'
   Context: xyz abc
            ^

3. Incomplete command: 'hello'
   ❌ Error at line 1, column 1: No rule found matching tokens: [HELLO]
   Context: hello
            ^

4. Empty input: ''
   ❌ Error at line 1, column 1: Empty input
   Context: 
            ^

5. Whitespace only: '   '
   ❌ Error at line 1, column 1: Empty input (whitespace only)
   Context:    
            ^

6. Unexpected token: 'hello 123'
   ❌ Error at line 1, column 7: No rule found matching tokens: [HELLO NUMBER]
   Context: hello 123
                  ^

=== Comparison: Standard error vs ParseError ===

Standard error (compatible):
  Unrecognized token 'xyz'

ParseError (enhanced):
  Error at line 1, column 1: Unrecognized token 'xyz'
  Position: 0
  Token: 'xyz'
  Input: 'xyz abc'
  Visual context:
    xyz abc
    ^
```

## 🔄 Gradual Migration

### Recommended Strategy

```go
// Step 1: Keep existing code working
result, err := dsl.Parse(input)
if err != nil {
    // Existing code continues working
    log.Printf("Error: %s", err.Error())
    
    // Step 2: Add enhanced information gradually
    if IsParseError(err) {
        parseErr := err.(*ParseError)
        log.Printf("Details: line %d, column %d", 
                   parseErr.Line, parseErr.Column)
    }
}
```

### Migration Benefits

1. **No Breaking Changes**: All existing code works
2. **Optional Information**: You can use new features when needed
3. **Better Debugging**: More informative errors for development
4. **Better UX**: End users get better error messages

## 🎯 Practical Use Cases

### 1. **IDEs and Editors**
```go
if IsParseError(err) {
    parseErr := err.(*ParseError)
    // Highlight error at specific line and column
    highlightError(parseErr.Line, parseErr.Column)
}
```

### 2. **Web APIs**
```go
// JSON response with detailed information
if IsParseError(err) {
    parseErr := err.(*ParseError)
    return ErrorResponse{
        Message:  parseErr.Message,
        Line:     parseErr.Line,
        Column:   parseErr.Column,
        Context:  parseErr.Input,
    }
}
```

### 3. **CLI Tools**
```go
// Show error with visual context
if IsParseError(err) {
    parseErr := err.(*ParseError)
    showVisualError(parseErr.Input, parseErr.Line, parseErr.Column)
}
```

### 4. **Logging Systems**
```go
// Structured logging with complete information
if IsParseError(err) {
    parseErr := err.(*ParseError)
    logger.WithFields(map[string]interface{}{
        "line":     parseErr.Line,
        "column":   parseErr.Column,
        "position": parseErr.Position,
        "token":    parseErr.Token,
    }).Error(parseErr.Message)
}
```

## 🔧 Technical Features

### 1. **Type Preservation**
```go
// ParseError is preserved through the call chain
func (dsl *DSL) Parse(input string) (*Result, error) {
    // ...
    if IsParseError(err) {
        return nil, err  // Preserves ParseError, doesn't wrap it
    }
    return nil, fmt.Errorf("other error type: %w", err)
}
```

### 2. **Intelligent Detection**
```go
// Helper function for robust detection
func IsParseError(err error) bool {
    if err == nil {
        return false
    }
    _, ok := err.(*ParseError)
    return ok
}
```

### 3. **Visual Context**
```go
// Function to show visual context
func getContextLine(input string, position int) string {
    return input  // Complete input available
}
```

## 🎓 Technical Lessons

### 1. **Compatibility is Key**
New features should not break existing code.

### 2. **Gradual Information**
Users can adopt new features at their own pace.

### 3. **Explicit Types**
`ParseError` is a specific type, not just a generic `error`.

### 4. **Helper Functions**
Functions like `IsParseError()` simplify usage.

## 🔗 Similar Cases

- **Compilers**: Syntax errors with line/column
- **Linters**: Problem reporting with location
- **IDEs**: Error underlining in code
- **APIs**: Structured error responses
- **CLIs**: Informative error messages

## 🚀 Next Steps

1. **Run the example**: `go run main.go`
2. **Modify test inputs** in the code
3. **Experiment with different error types**
4. **Integrate into your own code** gradually
5. **Compare with existing error systems**

## 📞 References and Documentation

- **Source code**: [`main.go`](main.go)
- **ParseError tests**: [../../pkg/dslbuilder/parse_error_test.go](../../pkg/dslbuilder/parse_error_test.go)
- **Complete manual**: [Usage Manual](../../docs/es/manual_de_uso.md) (Spanish)
- **Technical documentation**: [Developer Onboarding](../../docs/es/developer_onboarding.md) (Spanish)

---

**Demonstrates that go-dsl has informative errors without breaking compatibility!** 🔧🎉