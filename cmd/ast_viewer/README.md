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
  "type": "root",
  "value": "30",
  "children": [
    {
      "type": "arg_0",
      "value": "10"
    },
    {
      "type": "arg_1",
      "value": "+"
    },
    {
      "type": "arg_2",
      "value": "20"
    }
  ]
}
```

### Tree Format
```
root: 30
  ├─ arg_0: 10
  ├─ arg_1: +
  └─ arg_2: 20
```

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

## Limitations

- The current implementation shows a simplified AST structure
- Full parse tree information requires deeper integration with the parser
- Actions are executed but their internal logic is not shown

## Future Enhancements

- [ ] Full parse tree with all intermediate nodes
- [ ] Token position tracking (line, column)
- [ ] Rule matching visualization
- [ ] Interactive web-based viewer
- [ ] Export to GraphViz format