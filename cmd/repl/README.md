# DSL REPL

An interactive Read-Eval-Print Loop (REPL) for testing and exploring your DSLs.

## Overview

The DSL REPL provides an interactive environment to test your DSL commands, explore grammar rules, and debug parsing issues in real-time. It supports context variables, command history, and multiple output formats.

## Installation

```bash
go install github.com/arturoeanton/go-dsl/cmd/repl@latest
```

Or build from source:

```bash
cd cmd/repl
go build -o repl
```

## Usage

```bash
repl -dsl <dsl-file> [options]
```

### Options

- `-dsl` - DSL configuration file (YAML or JSON) **[required]**
- `-history` - History file to save/load commands
- `-context` - Context file (JSON) to preload
- `-ast` - Show AST representation of parsed input
- `-time` - Show execution time for each command
- `-multiline` - Enable multiline input mode
- `-exec` - Execute commands (can be used multiple times)

### Examples

**Basic interactive session:**
```bash
repl -dsl calculator.yaml
```

**With context and history:**
```bash
repl -dsl query.json -context data.json -history query_history.txt
```

**Execute commands and exit:**
```bash
repl -dsl accounting.yaml -exec "venta de 1000" -exec "venta de 2000"
```

**Debug mode with AST and timing:**
```bash
repl -dsl mydsl.yaml -ast -time
```

## REPL Commands

| Command | Description |
|---------|-------------|
| `.help` | Show available REPL commands |
| `.exit` | Exit the REPL |
| `.history` | Show command history |
| `.clear` | Clear the screen |
| `.context` | Show current context variables |
| `.set <key> <value>` | Set context variable |
| `.load <file>` | Load and execute commands from file |
| `.save <file>` | Save history to file |
| `.ast on/off` | Toggle AST display |
| `.time on/off` | Toggle execution time display |
| `.multiline` | Toggle multiline input mode |
| `.tokens` | **NEW:** Show available tokens |
| `.rules` | **NEW:** Show available rules |
| `.reset` | **NEW:** Reset context and buffer |
| `.last` | **NEW:** Show last command and result |

## New Features

### Enhanced Error Display
- Color-coded error messages (red)
- Helpful suggestions for common errors
- Token hints for unexpected token errors
- Command suggestions for typos

### Better Output Formatting
- Color-coded output by type:
  - Cyan for numbers
  - Magenta for booleans
  - Gray for nil values
- Pretty-printed arrays and objects
- JSON formatting for complex types

### Improved Commands
- `.tokens` - View available DSL tokens
- `.rules` - View available DSL rules
- `.reset` - Clear context and buffers
- `.last` - Quickly see last command result

## Features

### Interactive Mode
```
DSL> 10 + 20
30
DSL> 5 * (3 + 2)
25
DSL> .time on
Time display: true
DSL> 100 / 4
25
⏱  125µs
```

### Context Variables
```
Query> .set users ["John", "Jane", "Bob"]
Set users = ["John", "Jane", "Bob"]
Query> select name from users
["John", "Jane", "Bob"]
```

### Multiline Input
```
DSL> .multiline
Multiline mode: true
DSL> create function calculate
... with parameters x, y
... return x + y
... 
Function created successfully
```

### Command History
```
DSL> .history
[1] 10 + 20 => 30
[2] 5 * 3 => 15
[3] .set x 100 
[4] x / 2 => 50
```

### Loading Scripts
Create a script file `commands.txt`:
```
# Calculator commands
10 + 20
30 * 2
# Set variables
.set tax 0.16
1000 * tax
```

Load and execute:
```
Calculator> .load commands.txt
[commands.txt:2] 10 + 20
30
[commands.txt:3] 30 * 2
60
[commands.txt:5] .set tax 0.16
Set tax = 0.16
[commands.txt:6] 1000 * tax
160
```

## Output Formats

### Standard Output
Direct display of results:
```
Calculator> 42
42
```

### Array Output
Indexed display for arrays:
```
Query> select all from data
[0] {name: "John", age: 30}
[1] {name: "Jane", age: 25}
[2] {name: "Bob", age: 35}
```

### JSON Output
Pretty-printed for complex objects:
```
DSL> process data
{
  "status": "success",
  "count": 3,
  "results": [...]
}
```

## Use Cases

1. **DSL Testing**: Quickly test grammar rules and actions
2. **Interactive Development**: Develop DSL features interactively
3. **Debugging**: Debug parsing issues with AST visualization
4. **Demonstrations**: Show DSL capabilities to users
5. **Learning**: Explore DSL syntax and features
6. **Batch Processing**: Execute DSL scripts from files

## Tips and Tricks

### Quick Testing
```bash
# Test a single command
echo "10 + 20" | repl -dsl calculator.yaml

# Test multiple commands
repl -dsl calc.yaml -exec "10+20" -exec "30*2" -exec "15/3"
```

### Debugging Parse Errors
```
DSL> .ast on
AST display: true
DSL> invalid syntax here
Error: Unexpected token 'invalid' at position 0
```

### Performance Testing
```
DSL> .time on
Time display: true
DSL> complex calculation here
Result: 42
⏱  2.5ms
```

### Session Recording
```bash
# Start with history recording
repl -dsl mydsl.yaml -history session.log

# Later replay the session
repl -dsl mydsl.yaml -exec "$(cat session.log | grep -v '^#')"
```