# DSL Validator

A comprehensive tool to validate DSL grammar definitions and detect potential issues before runtime.

## Overview

The DSL Validator analyzes your DSL configuration files to identify syntax errors, semantic issues, and potential problems that could cause runtime failures. It provides detailed feedback to help you create robust and efficient DSLs.

## Installation

```bash
go install github.com/arturoeanton/go-dsl/cmd/validator@latest
```

Or build from source:

```bash
cd cmd/validator
go build -o validator
```

## Usage

```bash
validator -dsl <dsl-file> [options]
```

### Options

- `-dsl` - DSL configuration file to validate (YAML or JSON) **[required]**
- `-verbose` - Show detailed validation information
- `-format` - Output format: `text`, `json`, or `yaml` (default: text)
- `-test` - Test input string to validate against the DSL
- `-info` - Show DSL information summary
- `-strict` - Enable strict validation mode

### Examples

**Basic validation:**
```bash
validator -dsl calculator.yaml
```

**Detailed validation with info:**
```bash
validator -dsl query.json -verbose -info
```

**Test with sample input:**
```bash
validator -dsl accounting.yaml -test "venta de 1000" -strict
```

**JSON output for CI/CD integration:**
```bash
validator -dsl mydsl.yaml -format json
```

## Validation Checks

### Token Validation
- ✓ Valid regex patterns
- ✓ Pattern complexity analysis
- ✓ Duplicate pattern detection
- ✓ Overly broad pattern warnings
- ✓ Unescaped special character detection

### Rule Validation
- ✓ Token/rule reference verification
- ✓ Empty pattern detection
- ✓ Duplicate rule warnings
- ✓ Left recursion detection
- ✓ Action reference tracking

### Grammar Analysis
- ✓ Start rule identification
- ✓ Unreachable rule detection
- ✓ Ambiguous grammar patterns
- ✓ Token priority conflicts

### Best Practices
- ✓ Naming convention checks
- ✓ Complexity warnings
- ✓ Performance impact analysis

## Output Examples

### Text Format (Default)
```
✓ DSL validation passed

DSL Information:
  Name: Calculator
  Tokens: 6
  Rules: 8

Warnings (2):
  ⚠ [LeftRecursion] Rule expr has left recursion
    Details: Left recursion is supported but may impact performance
  ⚠ [UnimplementedAction] Action calculate is referenced but not implemented
    Details: Make sure to implement all actions referenced in rules
```

### JSON Format
```json
{
  "valid": true,
  "errors": [],
  "warnings": [
    {
      "type": "LeftRecursion",
      "message": "Rule expr has left recursion",
      "details": "Left recursion is supported but may impact performance"
    }
  ],
  "info": {
    "name": "Calculator",
    "tokenCount": 6,
    "ruleCount": 8
  }
}
```

## Validation Rules

### Errors (Fail Validation)
- Invalid regex patterns in tokens
- References to undefined tokens/rules
- Empty rule patterns
- Failed DSL instantiation

### Warnings (Pass with Cautions)
- Left recursive rules
- Duplicate rule names
- Overly broad token patterns
- Missing action implementations
- No clear start rule

## Use Cases

1. **Pre-deployment Validation**: Check DSL files before production
2. **CI/CD Integration**: Automated grammar validation in pipelines
3. **Development Aid**: Catch issues during DSL development
4. **Documentation**: Generate DSL structure documentation
5. **Migration Validation**: Verify DSL files after updates

## Exit Codes

- `0` - Validation passed
- `1` - Validation failed or error occurred

## Integration Examples

### Git Pre-commit Hook
```bash
#!/bin/sh
validator -dsl mydsl.yaml -strict || exit 1
```

### GitHub Actions
```yaml
- name: Validate DSL
  run: |
    go install github.com/arturoeanton/go-dsl/cmd/validator@latest
    validator -dsl config/mydsl.yaml -format json
```

### Makefile
```makefile
validate:
    @validator -dsl $(DSL_FILE) -verbose -info
```