# go-dsl Command Line Tools

go-dsl provides powerful command-line tools to help you develop, test, and debug your Domain Specific Languages.

## Available Tools

### 🔍 AST Viewer (`ast_viewer`)
Visualize the Abstract Syntax Tree of your DSL parsing results with enhanced tree representation.

**Features:**
- Color-coded tree visualization
- Multiple output formats (tree, json, debug)
- Support for both YAML and JSON DSL configurations
- Recursive AST structure display

[Detailed Documentation](ast_viewer/README.md) | [Documentación en Español](ast_viewer/README.es.md)

### ✅ Grammar Validator (`validator`)
Validate your DSL grammar configuration and detect potential issues before runtime.

**Features:**
- Comprehensive grammar validation
- Token conflict detection
- Circular dependency analysis
- Left recursion warnings
- Detailed error and warning reports

[Detailed Documentation](validator/README.md) | [Documentación en Español](validator/README.es.md)

### 💻 Interactive REPL (`repl`)
Test and explore your DSL interactively with a Read-Eval-Print Loop.

**Features:**
- Interactive DSL testing
- Command history with arrow key navigation
- Colored output for better readability
- Built-in commands (.help, .tokens, .rules, .reset, .last, .exit)
- Context data support from JSON files

[Detailed Documentation](repl/README.md) | [Documentación en Español](repl/README.es.md)

## Installation

You can install all tools at once or individually:

```bash
# Install all tools
go install github.com/arturoeanton/go-dsl/cmd/...@latest

# Or install individually
go install github.com/arturoeanton/go-dsl/cmd/ast_viewer@latest
go install github.com/arturoeanton/go-dsl/cmd/validator@latest
go install github.com/arturoeanton/go-dsl/cmd/repl@latest
```

## Quick Examples

### AST Viewer
```bash
# View AST for a calculator expression
ast_viewer -dsl calculator.yaml -input "10 + 20 * 30" -format tree

# Debug token generation
ast_viewer -dsl mydsl.json -input "test input" -format debug
```

### Grammar Validator
```bash
# Basic validation
validator -dsl mydsl.yaml

# Detailed validation with warnings
validator -dsl mydsl.yaml -verbose -strict -info
```

### Interactive REPL
```bash
# Start REPL with a DSL
repl -dsl calculator.yaml

# Start REPL with context data
repl -dsl query.yaml -context data.json
```

## Common Use Cases

1. **DSL Development**: Use the validator to check your grammar, the REPL to test expressions, and the AST viewer to understand how your DSL parses input.

2. **Debugging**: The AST viewer's debug mode shows token generation, while the validator helps identify grammar issues.

3. **Documentation**: Use the AST viewer to generate examples of how your DSL structures are parsed.

4. **Testing**: The REPL provides an interactive environment for testing DSL expressions before integrating them into your application.

## Tips

- Start with the validator to ensure your grammar is correct
- Use the REPL to interactively test your DSL expressions
- Use the AST viewer to understand how your input is being parsed
- Enable verbose mode in all tools for detailed output

## Contributing

Contributions to improve these tools are welcome! Each tool has its own README with specific details about implementation and potential improvements.