# Advanced Grammar Examples

This example demonstrates the advanced grammar features added to go-dsl:

## Features Demonstrated

### 1. Operator Precedence and Associativity

The example shows how to define operators with different precedence levels and associativity:

- **Level 1**: Addition/Subtraction (lowest precedence, left associative)
- **Level 2**: Multiplication/Division (medium precedence, left associative)  
- **Level 3**: Exponentiation (highest precedence, right associative)

Example results:
```
2 + 3 * 4 = 14      (not 20, multiplication has higher precedence)
2 * 3 + 4 = 10      (multiplication first, then addition)
2 ^ 3 * 4 = 32      (8 * 4, exponentiation has highest precedence)
2 * 3 ^ 2 = 18      (2 * 9, exponentiation first)
2 ^ 3 ^ 2 = 512     (2^9, right associative so 2^(3^2))
(2 + 3) * 4 = 20    (parentheses override precedence)
```

### 2. Kleene Star (Zero or More)

Demonstrates repetition rules for zero or more elements:

```go
// Define a rule that accepts zero or more words
dsl.RuleWithRepetition("words", "WORD", "words")
```

Example results:
```
""                  -> []
"hello"             -> [hello]
"hello world"       -> [hello world]
"the quick brown"   -> [the quick brown fox]
```

### 3. Kleene Plus (One or More)

Demonstrates repetition rules for one or more elements:

```go
// Define a rule that requires at least one identifier
dsl.RuleWithPlusRepetition("identifiers", "ID", "ids")
```

Example results:
```
"variable"          -> variable
"object.property"   -> object.property
"com.example.Class" -> com.example.Class
"a.b.c.d.e"        -> a.b.c.d.e
```

### 4. Priority-Based Token Matching

Since Go's regexp package doesn't support lookbehind assertions, the example demonstrates priority-based token matching:

```go
// Keywords have higher priority than generic identifiers
dsl.KeywordToken("IF", "if")        // Priority: 90
dsl.KeywordToken("WHILE", "while")  // Priority: 90
dsl.Token("IDENTIFIER", "[a-zA-Z][a-zA-Z0-9]*")  // Priority: 0
```

Example results:
```
"if 42"       -> If statement with condition 42
"while 10"    -> While loop with condition 10  
"ifx 5"       -> Assignment: ifx = 5
"counter 100" -> Assignment: counter = 100
```

## Running the Example

```bash
cd examples/advanced_grammar
go run main.go
```

## Implementation Details

### Precedence and Associativity

The `RuleWithPrecedence` method allows you to specify:
- **Precedence**: Higher numbers mean higher priority
- **Associativity**: "left", "right", or "none"

### Repetition Rules

Two helper methods are provided:
- `RuleWithRepetition`: Creates rules for zero or more repetitions (Kleene star)
- `RuleWithPlusRepetition`: Creates rules for one or more repetitions (Kleene plus)

These methods automatically generate the necessary production rules with appropriate actions.

### Token Priority

The priority system ensures that more specific tokens (like keywords) are matched before more general ones (like identifiers). This solves common ambiguity issues in language design.

## Use Cases

These advanced features enable creation of:
- Programming language parsers with proper operator precedence
- Configuration languages with repeated sections
- Query languages with complex expressions
- Template languages with nested structures