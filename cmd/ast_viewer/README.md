# AST Viewer

A tool to visualize the Abstract Syntax Tree (AST) of your DSL parsing results.

## Overview

The AST Viewer helps you understand how your DSL parses input by displaying the resulting structure in various formats. This is essential for debugging grammar rules and understanding the parsing process.

## Installation

```bash
go install github.com/arturoeanton/go-dsl/cmd/ast_viewer@latest
```

Or build from source:

```bash
cd cmd/ast_viewer
go build -o ast_viewer
```

## Usage

```bash
ast_viewer -dsl <dsl-file> -input <input-string> [options]
```

### Options

- `-dsl` - DSL configuration file (YAML or JSON) **[required]**
- `-input` - Input string to parse
- `-file` - Input file to parse (alternative to -input)
- `-format` - Output format: `json`, `yaml`, or `tree` (default: json)
- `-indent` - Indent output for json/yaml (default: true)
- `-verbose` - Show detailed token and rule information

### Examples

**Basic usage with JSON output:**
```bash
ast_viewer -dsl calculator.yaml -input "10 + 20"
```

**Tree format visualization:**
```bash
ast_viewer -dsl calculator.yaml -input "10 + 20 * 30" -format tree
```

**Parse from file with YAML output:**
```bash
ast_viewer -dsl query.json -file queries.txt -format yaml
```

**Verbose mode for debugging:**
```bash
ast_viewer -dsl accounting.yaml -input "venta de 1000 con iva" -format tree -verbose
```

## Output Formats

### JSON Format
```json
{
  "type": "expression",
  "children": [
    {
      "type": "number",
      "value": "10"
    },
    {
      "type": "operator",
      "value": "+"
    },
    {
      "type": "number",
      "value": "20"
    }
  ]
}
```

### Tree Format (Enhanced)
```
◆ expression
├─ # number 10
├─ ● operator +
└─ # number 20
```

**Tree Symbols:**
- `◆` - Expression/Root nodes
- `●` - Operators
- `#` - Numbers
- `□` - Identifiers
- `{}` - Objects
- `?` - Booleans
- `○` - Generic nodes

**Colors in Tree Mode:**
- Cyan - Numbers
- Yellow - Operators
- Green - Strings/Identifiers
- Magenta - Actions

### YAML Format
```yaml
type: root
value: "30"
children:
  - type: arg_0
    value: "10"
  - type: arg_1
    value: "+"
  - type: arg_2
    value: "20"
```

## Use Cases

1. **Grammar Debugging**: Understand how your rules are being matched
2. **Token Analysis**: See which tokens are being recognized
3. **Rule Optimization**: Identify inefficient parsing patterns
4. **Documentation**: Generate visual representations of parsing results

## New Features

### Enhanced AST Building
- Recursive AST construction for nested structures
- Automatic type detection (numbers, operators, identifiers)
- Support for complex data types (objects, arrays)
- Better representation of parsing results

### Improved Tree Visualization
- Color-coded output for better readability
- Symbolic representation of node types
- Hierarchical structure with proper indentation
- Verbose mode shows token and rule details

## Limitations

- Token and rule information requires DSL introspection API
- Position tracking (line/column) not yet available
- Actions are executed but internal logic is opaque

## Future Enhancements

- [ ] Full parse tree with all intermediate nodes
- [ ] Token position tracking (line, column)
- [ ] Rule matching visualization
- [ ] Interactive web-based viewer
- [ ] Export to GraphViz/DOT format
- [ ] Real-time AST updates
- [ ] Syntax highlighting in tree view