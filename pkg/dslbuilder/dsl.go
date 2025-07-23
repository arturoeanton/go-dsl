// Package dslbuilder provides a dynamic DSL (Domain Specific Language) builder for Go.
// It allows you to create custom languages with tokens, rules, and actions that can
// call Go functions dynamically.
package dslbuilder

import (
	"fmt"
	"regexp"
	"strings"
)

// ParseError provides detailed error information with line and column.
// It implements the error interface and provides rich error context including
// the exact position where parsing failed, making debugging much easier.
type ParseError struct {
	Message  string // Original error message (for backward compatibility)
	Line     int    // Line number where error occurred (1-based)
	Column   int    // Column number where error occurred (1-based)
	Position int    // Character position in input (0-based)
	Token    string // Token value at error position
	Input    string // Original input for context display
}

// Error implements the error interface, maintaining backward compatibility.
// It returns only the error message without position information.
// For detailed error info with line/column, use DetailedError() instead.
func (pe *ParseError) Error() string {
	return pe.Message
}

// DetailedError returns a formatted error message with line/column information.
// It includes a visual pointer showing exactly where the error occurred in the input.
// Example output:
//
//	unexpected token at line 2, column 15:
//	if x > 10 then @
//	               ^
func (pe *ParseError) DetailedError() string {
	if pe.Line == 0 && pe.Column == 0 {
		return pe.Message // Fallback to simple message
	}

	context := pe.getContextLine()
	pointer := strings.Repeat(" ", pe.Column-1) + "^"

	return fmt.Sprintf("%s at line %d, column %d:\n%s\n%s",
		pe.Message, pe.Line, pe.Column, context, pointer)
}

// getContextLine extracts the line containing the error from the input.
// It's used internally to provide context in error messages.
func (pe *ParseError) getContextLine() string {
	if pe.Input == "" {
		return ""
	}

	lines := strings.Split(pe.Input, "\n")
	if pe.Line > 0 && pe.Line <= len(lines) {
		return lines[pe.Line-1]
	}

	return ""
}

// IsParseError checks if an error is a ParseError with detailed information.
// This is useful for error handling when you want to provide better error messages:
//
//	if IsParseError(err) {
//	    fmt.Println(GetDetailedError(err))
//	} else {
//	    fmt.Println(err)
//	}
func IsParseError(err error) bool {
	_, ok := err.(*ParseError)
	return ok
}

// GetDetailedError returns detailed error information if available.
// For ParseError types, it returns the detailed error with line/column info.
// For other error types, it returns the standard error message.
func GetDetailedError(err error) string {
	if parseErr, ok := err.(*ParseError); ok {
		return parseErr.DetailedError()
	}
	return err.Error()
}

// calculateLineColumn calculates line and column numbers from a character position.
// It takes a 0-based character position and returns 1-based line and column numbers.
// This is used internally to convert absolute positions to human-readable locations.
func calculateLineColumn(input string, position int) (line int, column int) {
	if position < 0 || position > len(input) {
		return 0, 0
	}

	line = 1
	column = 1

	for i := 0; i < position && i < len(input); i++ {
		if input[i] == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}

	return line, column
}

// createParseError creates a ParseError with full context information.
// It automatically calculates line/column from the character position and
// captures the input for error context display.
//
// Parameters:
//   - message: The error description
//   - position: Character position where error occurred (0-based)
//   - token: The token value at the error position
//   - input: The original input being parsed
func createParseError(message string, position int, token string, input string) *ParseError {
	line, column := calculateLineColumn(input, position)

	return &ParseError{
		Message:  message,
		Line:     line,
		Column:   column,
		Position: position,
		Token:    token,
		Input:    input,
	}
}

// DSL represents a Domain Specific Language instance.
// It encapsulates the grammar definition, semantic actions, and runtime context
// needed to parse and execute DSL code.
//
// A DSL consists of:
//   - Grammar: Token definitions and parsing rules
//   - Actions: Functions that execute when rules match
//   - Functions: Go functions exposed to the DSL
//   - Context: Runtime variables accessible during parsing
type DSL struct {
	name      string                 // Name of the DSL for identification
	grammar   *Grammar               // Grammar rules and tokens
	actions   map[string]ActionFunc  // Semantic actions for rules
	functions map[string]interface{} // Go functions available to DSL code
	context   map[string]interface{} // Runtime context variables
}

// ActionFunc is a function that processes parsed tokens and returns a result.
// It receives the matched tokens/values as arguments and can return any value
// or an error if processing fails.
//
// Example:
//
//	dsl.Action("add", func(args []interface{}) (interface{}, error) {
//	    left := args[0].(int)
//	    right := args[2].(int)  // args[1] would be the operator
//	    return left + right, nil
//	})
type ActionFunc func(args []interface{}) (interface{}, error)

// New creates a new DSL instance with the given name.
// The name is used for identification and debugging purposes.
//
// Example:
//
//	calc := dslbuilder.New("calculator")
//	calc.Token("NUM", "[0-9]+")
//	calc.Rule("expr", []string{"NUM"}, "number")
func New(name string) *DSL {
	return &DSL{
		name:      name,
		grammar:   NewGrammar(),
		actions:   make(map[string]ActionFunc),
		functions: make(map[string]interface{}),
		context:   make(map[string]interface{}),
	}
}

// Token defines a token in the DSL with a regular expression pattern.
// Tokens are the basic building blocks of your language (lexemes).
//
// Parameters:
//   - name: Token identifier used in grammar rules (e.g., "NUMBER", "ID")
//   - pattern: Regular expression pattern to match the token
//
// Example:
//
//	dsl.Token("NUMBER", "[0-9]+")          // Matches: 123, 456
//	dsl.Token("ID", "[a-zA-Z_][a-zA-Z0-9_]*") // Matches: var_name, _test
//	dsl.Token("STRING", `"[^"]*"`)        // Matches: "hello world"
//
// Returns an error if the regex pattern is invalid.
func (d *DSL) Token(name, pattern string) error {
	return d.grammar.AddToken(name, pattern)
}

// KeywordToken defines a keyword token with high priority.
// Keywords are matched before regular tokens to avoid conflicts
// (e.g., "if" should be IF token, not an ID token).
//
// The keyword is automatically wrapped with word boundaries (\b)
// and made case-insensitive for convenience.
//
// Parameters:
//   - name: Token identifier (e.g., "IF", "WHILE")
//   - keyword: The actual keyword text (e.g., "if", "while")
//
// Example:
//
//	dsl.KeywordToken("IF", "if")       // Matches: if, IF, If
//	dsl.KeywordToken("RETURN", "return") // Matches: return, RETURN
//
// Keywords have priority 90 (regular tokens have priority 0).
func (d *DSL) KeywordToken(name, keyword string) error {
	return d.grammar.AddKeywordToken(name, keyword)
}

// TokenWithLookaround defines a token with lookahead/lookbehind assertions.
// This allows for context-sensitive tokenization where a pattern matches
// only when specific conditions are met before or after it.
//
// Note: Go's regexp package has limited lookaround support.
// Lookbehind is stored but not enforced by the default implementation.
//
// Parameters:
//   - name: Token identifier
//   - pattern: Main pattern to match
//   - lookahead: Pattern that must follow (positive lookahead)
//   - lookbehind: Pattern that must precede (stored but not enforced)
//
// Example:
//
//	// Match "=" only when followed by "="
//	dsl.TokenWithLookaround("ASSIGN", "=", "=", "")
//	// Match number only when followed by whitespace or EOF
//	dsl.TokenWithLookaround("NUM", "[0-9]+", "\\s|$", "")
//
// Tokens with lookaround have priority 50.
func (d *DSL) TokenWithLookaround(name, pattern string, lookahead, lookbehind string) error {
	return d.grammar.AddTokenWithLookaround(name, pattern, lookahead, lookbehind)
}

// Rule defines a grammar rule that describes how to parse a language construct.
// Rules can reference tokens and other rules to build complex grammars.
//
// Parameters:
//   - name: Rule identifier (e.g., "expression", "statement")
//   - pattern: Sequence of token names or rule names to match
//   - actionName: Name of the action function to execute when matched
//
// Example:
//
//	// expression → NUMBER
//	dsl.Rule("expression", []string{"NUMBER"}, "number")
//	// expression → expression '+' expression
//	dsl.Rule("expression", []string{"expression", "PLUS", "expression"}, "add")
//	// statement → 'if' expression 'then' statement
//	dsl.Rule("statement", []string{"IF", "expression", "THEN", "statement"}, "ifStmt")
//
// Multiple rules with the same name create alternatives (like BNF |).
// The first rule defined becomes the start rule if not otherwise specified.
func (d *DSL) Rule(name string, pattern []string, actionName string) {
	d.grammar.AddRule(name, pattern, actionName)
}

// RuleWithPrecedence defines a grammar rule with precedence and associativity.
// This is essential for parsing expressions with operators correctly.
//
// Parameters:
//   - name: Rule identifier
//   - pattern: Sequence to match
//   - actionName: Action to execute
//   - precedence: Higher numbers = higher priority (tighter binding)
//   - associativity: "left", "right", or "none"
//
// Example:
//
//	// Multiplication has higher precedence than addition
//	dsl.RuleWithPrecedence("expr", []string{"expr", "TIMES", "expr"}, "mul", 20, "left")
//	dsl.RuleWithPrecedence("expr", []string{"expr", "PLUS", "expr"}, "add", 10, "left")
//	// Power operator is right-associative
//	dsl.RuleWithPrecedence("expr", []string{"expr", "POW", "expr"}, "pow", 30, "right")
//
// With left associativity: 1+2+3 = (1+2)+3
// With right associativity: 2^3^4 = 2^(3^4)
func (d *DSL) RuleWithPrecedence(name string, pattern []string, actionName string, precedence int, associativity string) {
	d.grammar.AddRuleWithPrecedence(name, pattern, actionName, precedence, associativity)
}

// RuleWithRepetition defines a rule with Kleene star (zero or more repetitions).
// This creates a rule that matches zero or more occurrences of an element.
//
// Internally creates two rules:
//   - name → ε (empty match)
//   - name → name element (recursive for multiple matches)
//
// Parameters:
//   - name: Rule identifier for the repetition
//   - element: Token or rule to repeat
//   - actionName: Base name for actions ("_empty" and "_append" suffixes added)
//
// Example:
//
//	// Define: arglist → (expression (',' expression)*) ?
//	dsl.RuleWithRepetition("args", "arg", "collectArgs")
//	// You need to define actions:
//	dsl.Action("collectArgs_empty", func(args []interface{}) (interface{}, error) {
//	    return []interface{}{}, nil // Empty list
//	})
//	dsl.Action("collectArgs_append", func(args []interface{}) (interface{}, error) {
//	    list := args[0].([]interface{})
//	    return append(list, args[1]), nil
//	})
func (d *DSL) RuleWithRepetition(name string, element string, actionName string) {
	// Create two rules: one for empty, one for one-or-more
	// name → ε (empty)
	d.grammar.AddRule(name, []string{}, actionName+"_empty")
	// name → name element (left recursive for one or more)
	d.grammar.AddRule(name, []string{name, element}, actionName+"_append")
}

// RuleWithPlusRepetition defines a rule with Kleene plus (one or more repetitions).
// This creates a rule that matches one or more occurrences of an element.
//
// Internally creates two rules:
//   - name → element (single match)
//   - name → name element (recursive for multiple matches)
//
// Parameters:
//   - name: Rule identifier for the repetition
//   - element: Token or rule to repeat (must match at least once)
//   - actionName: Base name for actions ("_single" and "_append" suffixes added)
//
// Example:
//
//	// Define: numbers → NUMBER+
//	dsl.RuleWithPlusRepetition("numbers", "NUMBER", "collectNums")
//	// You need to define actions:
//	dsl.Action("collectNums_single", func(args []interface{}) (interface{}, error) {
//	    return []interface{}{args[0]}, nil // Single item list
//	})
//	dsl.Action("collectNums_append", func(args []interface{}) (interface{}, error) {
//	    list := args[0].([]interface{})
//	    return append(list, args[1]), nil
//	})
func (d *DSL) RuleWithPlusRepetition(name string, element string, actionName string) {
	// name → element
	d.grammar.AddRule(name, []string{element}, actionName+"_single")
	// name → name element (left recursive)
	d.grammar.AddRule(name, []string{name, element}, actionName+"_append")
}

// Action registers an action function that executes when a rule matches.
// Actions transform the parsed tokens/values into meaningful results.
//
// Parameters:
//   - name: Action identifier matching the rule's actionName
//   - fn: Function that processes matched values
//
// The ActionFunc receives matched values in order and returns a result.
// For tokens, args contain the matched string values.
// For rules, args contain the results of their actions.
//
// Example:
//
//	dsl.Action("number", func(args []interface{}) (interface{}, error) {
//	    // args[0] is the NUMBER token value
//	    return strconv.Atoi(args[0].(string))
//	})
//
//	dsl.Action("add", func(args []interface{}) (interface{}, error) {
//	    // args[0] = left expression result
//	    // args[1] = "+" token
//	    // args[2] = right expression result
//	    left := args[0].(int)
//	    right := args[2].(int)
//	    return left + right, nil
//	})
func (d *DSL) Action(name string, fn ActionFunc) {
	d.actions[name] = fn
	d.grammar.actions[name] = fn
}

// Builder Pattern Methods for fluent API

// WithToken adds a token and returns the DSL for chaining
func (d *DSL) WithToken(name, pattern string) *DSL {
	d.Token(name, pattern)
	return d
}

// WithKeywordToken adds a keyword token and returns the DSL for chaining
func (d *DSL) WithKeywordToken(name, keyword string) *DSL {
	d.KeywordToken(name, keyword)
	return d
}

// WithRule adds a rule and returns the DSL for chaining
func (d *DSL) WithRule(name string, pattern []string, actionName string) *DSL {
	d.Rule(name, pattern, actionName)
	return d
}

// WithRulePrecedence adds a rule with precedence and returns the DSL for chaining
func (d *DSL) WithRulePrecedence(name string, pattern []string, actionName string, precedence int, associativity string) *DSL {
	d.RuleWithPrecedence(name, pattern, actionName, precedence, associativity)
	return d
}

// WithRepetition adds a rule with Kleene star and returns the DSL for chaining
func (d *DSL) WithRepetition(name string, element string, actionName string) *DSL {
	d.RuleWithRepetition(name, element, actionName)
	return d
}

// WithPlusRepetition adds a rule with Kleene plus and returns the DSL for chaining
func (d *DSL) WithPlusRepetition(name string, element string, actionName string) *DSL {
	d.RuleWithPlusRepetition(name, element, actionName)
	return d
}

// WithTokenLookaround adds a token with lookaround and returns the DSL for chaining
func (d *DSL) WithTokenLookaround(name, pattern, lookahead, lookbehind string) *DSL {
	d.TokenWithLookaround(name, pattern, lookahead, lookbehind)
	return d
}

// WithAction adds an action and returns the DSL for chaining
func (d *DSL) WithAction(name string, fn ActionFunc) *DSL {
	d.Action(name, fn)
	return d
}

// WithContext sets a context value and returns the DSL for chaining
func (d *DSL) WithContext(key string, value interface{}) *DSL {
	d.SetContext(key, value)
	return d
}

// WithFunction registers a Go function and returns the DSL for chaining
func (d *DSL) WithFunction(name string, fn interface{}) *DSL {
	d.Set(name, fn)
	return d
}

// SetContext sets a context variable that can be accessed during parsing.
// Context provides a way to pass configuration or state to your DSL.
//
// Parameters:
//   - key: Variable name
//   - value: Any value (accessed later via GetContext)
//
// Example:
//
//	dsl.SetContext("debug", true)
//	dsl.SetContext("maxDepth", 10)
//	dsl.SetContext("variables", map[string]int{"x": 5})
//
// Context values persist across Parse calls unless overwritten.
func (d *DSL) SetContext(key string, value interface{}) {
	d.context[key] = value
}

// GetContext retrieves a context variable previously set with SetContext.
// Returns nil if the key doesn't exist.
//
// Example:
//
//	if debug, ok := dsl.GetContext("debug").(bool); ok && debug {
//	    fmt.Println("Debug mode enabled")
//	}
//
//	vars := dsl.GetContext("variables").(map[string]int)
func (d *DSL) GetContext(key string) interface{} {
	return d.context[key]
}

// Set registers a Go function that can be called from DSL code.
// This allows your DSL to invoke external functionality.
//
// Parameters:
//   - name: Function name as it will be called from DSL
//   - fn: Any Go function
//
// Example:
//
//	// Register a math function
//	dsl.Set("sqrt", math.Sqrt)
//
//	// Register a custom function
//	dsl.Set("greet", func(name string) string {
//	    return fmt.Sprintf("Hello, %s!", name)
//	})
//
//	// Use in action:
//	dsl.Action("callFunc", func(args []interface{}) (interface{}, error) {
//	    fnName := args[0].(string)
//	    if fn, ok := dsl.Get(fnName); ok {
//	        // Call function with remaining args...
//	    }
//	})
func (d *DSL) Set(name string, fn interface{}) {
	d.functions[name] = fn
}

// Get retrieves a registered function by name.
// Returns the function and true if found, nil and false otherwise.
//
// Example:
//
//	if fn, ok := dsl.Get("sqrt"); ok {
//	    if sqrtFn, ok := fn.(func(float64) float64); ok {
//	        result := sqrtFn(16.0) // Returns 4.0
//	    }
//	}
func (d *DSL) Get(name string) (interface{}, bool) {
	fn, exists := d.functions[name]
	return fn, exists
}

// Debug returns debug information about the DSL configuration.
// Useful for understanding the grammar structure and troubleshooting.
//
// Returns a map containing:
//   - "name": DSL name
//   - "tokens": Map of token names to their patterns
//   - "rules": Map of rule names to their alternatives
//
// Example output:
//
//	{
//	    "name": "calculator",
//	    "tokens": {
//	        "NUMBER": "[0-9]+",
//	        "PLUS": "\\+"
//	    },
//	    "rules": {
//	        "expr": [
//	            {"sequence": ["NUMBER"], "action": "number"},
//	            {"sequence": ["expr", "PLUS", "expr"], "action": "add"}
//	        ]
//	    }
//	}
func (d *DSL) Debug() map[string]interface{} {
	debug := make(map[string]interface{})
	debug["name"] = d.name
	debug["tokens"] = make(map[string]string)
	debug["rules"] = make(map[string]interface{})

	// Add token info
	for name, token := range d.grammar.tokens {
		debug["tokens"].(map[string]string)[name] = token.pattern
	}

	// Add rule info
	for name, rule := range d.grammar.rules {
		alternatives := make([]map[string]interface{}, 0)
		for _, alt := range rule.alternatives {
			altInfo := map[string]interface{}{
				"sequence": alt.sequence,
				"action":   alt.action,
			}
			alternatives = append(alternatives, altInfo)
		}
		debug["rules"].(map[string]interface{})[name] = alternatives
	}

	return debug
}

// DebugTokens returns the tokens that would be generated for a given code.
// This is useful for understanding how your input is being tokenized.
//
// Parameters:
//   - code: Input string to tokenize
//
// Returns:
//   - []TokenMatch: Array of matched tokens with their positions
//   - error: Tokenization error if any
//
// Example:
//
//	tokens, err := dsl.DebugTokens("x = 42")
//	// Returns: [
//	//   {TokenType: "ID", Value: "x", Start: 0, End: 1},
//	//   {TokenType: "ASSIGN", Value: "=", Start: 2, End: 3},
//	//   {TokenType: "NUMBER", Value: "42", Start: 4, End: 6}
//	// ]
func (d *DSL) DebugTokens(code string) ([]TokenMatch, error) {
	parser := NewParser(d.grammar)
	err := parser.tokenize(code)
	if err != nil {
		return nil, err
	}
	return parser.tokens, nil
}

// Use evaluates DSL code with an optional context override.
// The provided context is merged with existing context for this parse only.
//
// Parameters:
//   - code: DSL code to parse and evaluate
//   - ctx: Additional context variables (merged with existing)
//
// Example:
//
//	result, err := dsl.Use("x + y", map[string]interface{}{
//	    "x": 10,
//	    "y": 20,
//	})
//
// The context values are available to actions during parsing.
func (d *DSL) Use(code string, ctx map[string]interface{}) (*Result, error) {
	// Merge context if provided
	if ctx != nil {
		for k, v := range ctx {
			d.context[k] = v
		}
	}

	return d.Parse(code)
}

// Parse parses and evaluates DSL code using the improved parser.
// This is the main entry point for executing DSL code.
//
// The improved parser includes:
//   - Memoization for better performance
//   - Left recursion support
//   - Enhanced error reporting with line/column info
//
// Parameters:
//   - code: DSL code to parse
//
// Returns:
//   - *Result: Contains AST and output value
//   - error: ParseError with detailed position info if parsing fails
//
// Example:
//
//	result, err := dsl.Parse("2 + 3 * 4")
//	if err != nil {
//	    if IsParseError(err) {
//	        fmt.Println(GetDetailedError(err)) // Shows line/column
//	    }
//	    return err
//	}
//	fmt.Println(result.GetOutput()) // Prints: 14
func (d *DSL) Parse(code string) (*Result, error) {
	parser := NewImprovedParser(d.grammar)
	parser.dsl = d // Give parser access to DSL functions
	ast, err := parser.Parse(code)
	if err != nil {
		// Preserve ParseError type for enhanced error information
		if IsParseError(err) {
			return nil, err
		}
		// Only wrap non-ParseError errors
		return nil, fmt.Errorf("parsing error: %w", err)
	}

	return &Result{
		AST:    ast,
		Code:   code,
		Output: ast,
		DSL:    d,
	}, nil
}

// Grammar represents the grammar of a DSL.
// It contains all the tokens, rules, and actions that define the language.
//
// A grammar consists of:
//   - tokens: Terminal symbols (lexemes) with regex patterns
//   - rules: Non-terminal symbols defined by sequences of symbols
//   - startRule: The root rule to begin parsing
//   - actions: Functions that process matched patterns
type Grammar struct {
	rules     map[string]*Rule      // Named grammar rules
	tokens    map[string]*Token     // Named token definitions
	startRule string                // Entry point for parsing
	actions   map[string]ActionFunc // Semantic actions
}

// Rule represents a grammar rule (non-terminal symbol).
// A rule can have multiple alternatives, similar to BNF notation.
//
// Example rule "expr" with alternatives:
//   - expr → expr '+' term
//   - expr → expr '-' term
//   - expr → term
type Rule struct {
	name         string         // Rule identifier
	alternatives []*Alternative // Different ways to match this rule
}

// Alternative represents one way to match a rule.
// Each alternative has a sequence of symbols and an associated action.
//
// Fields:
//   - sequence: Ordered list of token/rule names to match
//   - action: Name of the function to call when matched
//   - precedence: For operators (higher = tighter binding)
//   - associativity: How operators of same precedence combine
type Alternative struct {
	sequence      []string // Symbol sequence to match
	action        string   // Action function name
	precedence    int      // Operator precedence (higher = higher priority)
	associativity string   // "left", "right", or "none"
}

// Token represents a token (terminal symbol) in the grammar.
// Tokens are matched using regular expressions during lexical analysis.
//
// Fields:
//   - name: Token identifier used in rules
//   - pattern: Original regex pattern string
//   - regex: Compiled regex for matching
//   - priority: Higher priority tokens match first (keywords=90)
//   - lookahead: Pattern that must follow (if using lookaround)
//   - lookbehind: Pattern that must precede (stored but not enforced)
type Token struct {
	name       string         // Token identifier
	pattern    string         // Regex pattern string
	regex      *regexp.Regexp // Compiled pattern
	priority   int            // Matching priority
	lookahead  string         // Positive lookahead pattern
	lookbehind string         // Positive lookbehind pattern
}

// NewGrammar creates a new empty grammar.
// The grammar can be populated with tokens and rules.
func NewGrammar() *Grammar {
	return &Grammar{
		rules:   make(map[string]*Rule),
		tokens:  make(map[string]*Token),
		actions: make(map[string]ActionFunc),
	}
}

// AddToken adds a token to the grammar with the given regex pattern.
// Regular tokens have priority 0 (keywords have priority 90).
//
// Parameters:
//   - name: Token identifier (e.g., "NUMBER", "IDENTIFIER")
//   - pattern: Regular expression pattern
//
// Returns an error if the regex pattern is invalid.
//
// Example:
//
//	g.AddToken("NUMBER", "[0-9]+")           // Integers
//	g.AddToken("FLOAT", "[0-9]+\\.[0-9]+")   // Floats
//	g.AddToken("STRING", `"([^"\\]|\\.)*"`) // Quoted strings
func (g *Grammar) AddToken(name, pattern string) error {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}

	g.tokens[name] = &Token{
		name:     name,
		pattern:  pattern,
		regex:    regex,
		priority: 0,
	}
	return nil
}

// AddKeywordToken adds a keyword token with high priority.
// Keywords are automatically wrapped with word boundaries and made case-insensitive.
//
// The pattern created is: (?i)\b<keyword>\b
//   - (?i) makes it case-insensitive
//   - \b ensures word boundaries
//
// Parameters:
//   - name: Token identifier (e.g., "IF", "WHILE")
//   - keyword: The keyword text (e.g., "if", "while")
//
// Keywords have priority 90 to match before identifiers.
//
// Example:
//
//	g.AddKeywordToken("RETURN", "return") // Matches: return, Return, RETURN
//	g.AddKeywordToken("CLASS", "class")   // Won't match: subclass, classname
func (g *Grammar) AddKeywordToken(name, keyword string) error {
	pattern := "(?i)\\b" + regexp.QuoteMeta(keyword) + "\\b"
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid keyword pattern: %w", err)
	}

	g.tokens[name] = &Token{
		name:     name,
		pattern:  pattern,
		regex:    regex,
		priority: 90, // High priority for keywords
	}
	return nil
}

// AddTokenWithLookaround adds a token with lookahead/lookbehind assertions.
// This provides context-sensitive tokenization.
//
// Note: Go's regexp package has limited lookaround support:
//   - Positive lookahead is supported: (?=pattern)
//   - Lookbehind is stored but not enforced by default parser
//
// Parameters:
//   - name: Token identifier
//   - pattern: Main pattern to match
//   - lookahead: Pattern that must follow
//   - lookbehind: Pattern that must precede (stored only)
//
// Lookaround tokens have priority 50 (between regular and keywords).
//
// Example:
//
//	// Match "=" only when not followed by "="
//	g.AddTokenWithLookaround("ASSIGN", "=", "[^=]", "")
//	// Match word only when followed by "("
//	g.AddTokenWithLookaround("FUNC_CALL", "\\w+", "\\(", "")
func (g *Grammar) AddTokenWithLookaround(name, pattern, lookahead, lookbehind string) error {
	// Note: Go's regexp package doesn't support lookbehind assertions
	// We'll implement a custom solution using context checking

	// For now, just add the basic pattern and store lookaround info
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}

	g.tokens[name] = &Token{
		name:       name,
		pattern:    pattern,
		regex:      regex,
		priority:   50, // Medium priority for lookaround tokens
		lookahead:  lookahead,
		lookbehind: lookbehind,
	}
	return nil
}

// AddRule adds a rule alternative to the grammar.
// Multiple calls with the same rule name create alternatives (like BNF |).
//
// Parameters:
//   - name: Rule identifier
//   - sequence: Array of symbol names to match (tokens or rules)
//   - action: Name of the action function to execute
//
// The first rule added becomes the start rule if not set.
// Default precedence is 0 with left associativity.
//
// Example:
//
//	// expr → expr '+' expr | expr '*' expr | '(' expr ')' | NUMBER
//	g.AddRule("expr", []string{"expr", "PLUS", "expr"}, "add")
//	g.AddRule("expr", []string{"expr", "TIMES", "expr"}, "mul")
//	g.AddRule("expr", []string{"LPAREN", "expr", "RPAREN"}, "paren")
//	g.AddRule("expr", []string{"NUMBER"}, "number")
func (g *Grammar) AddRule(name string, sequence []string, action string) {
	rule, exists := g.rules[name]
	if !exists {
		rule = &Rule{
			name:         name,
			alternatives: []*Alternative{},
		}
		g.rules[name] = rule
		if g.startRule == "" {
			g.startRule = name
		}
	}

	rule.alternatives = append(rule.alternatives, &Alternative{
		sequence:      sequence,
		action:        action,
		precedence:    0,
		associativity: "left", // default
	})
}

// AddRuleWithPrecedence adds a rule with explicit precedence and associativity.
// Essential for correctly parsing expressions with multiple operators.
//
// Parameters:
//   - name: Rule identifier
//   - sequence: Symbol sequence to match
//   - action: Action function name
//   - precedence: Higher = tighter binding (e.g., * > +)
//   - associativity: "left", "right", or "none"
//
// Associativity determines grouping for same-precedence operators:
//   - "left": a+b+c = (a+b)+c
//   - "right": a^b^c = a^(b^c)
//   - "none": a<b<c is an error
//
// Example:
//
//	// Typical arithmetic precedence
//	g.AddRuleWithPrecedence("expr", []string{"expr", "POW", "expr"}, "pow", 30, "right")
//	g.AddRuleWithPrecedence("expr", []string{"expr", "MUL", "expr"}, "mul", 20, "left")
//	g.AddRuleWithPrecedence("expr", []string{"expr", "ADD", "expr"}, "add", 10, "left")
func (g *Grammar) AddRuleWithPrecedence(name string, sequence []string, action string, precedence int, associativity string) {
	rule, exists := g.rules[name]
	if !exists {
		rule = &Rule{
			name:         name,
			alternatives: []*Alternative{},
		}
		g.rules[name] = rule
		if g.startRule == "" {
			g.startRule = name
		}
	}

	// Validate associativity
	if associativity != "left" && associativity != "right" && associativity != "none" {
		associativity = "left" // default
	}

	rule.alternatives = append(rule.alternatives, &Alternative{
		sequence:      sequence,
		action:        action,
		precedence:    precedence,
		associativity: associativity,
	})
}

// Parser represents a DSL parser instance.
// The parser performs lexical analysis (tokenization) and
// syntax analysis (parsing) of input according to the grammar.
//
// Fields:
//   - grammar: The language grammar definition
//   - tokens: Array of tokens from lexical analysis
//   - pos: Current position in token array
//   - dsl: Parent DSL for accessing functions/context
//   - input: Original input for error reporting
type Parser struct {
	grammar *Grammar     // Language grammar
	tokens  []TokenMatch // Tokenized input
	pos     int          // Current token position
	dsl     *DSL         // Reference to parent DSL for function access
	input   string       // Original input for error reporting
}

// TokenMatch represents a matched token during lexical analysis.
// Contains the token type, matched text, and position information.
//
// Fields:
//   - TokenType: Name of the matched token (e.g., "NUMBER", "ID")
//   - Value: The actual matched text
//   - Start: Starting position in input (0-based)
//   - End: Ending position in input (exclusive)
//
// Example:
//
//	Input: "x = 42"
//	Token: {TokenType: "NUMBER", Value: "42", Start: 4, End: 6}
type TokenMatch struct {
	TokenType string // Token name from grammar
	Value     string // Matched text
	Start     int    // Start position in input
	End       int    // End position (exclusive)
}

// NewParser creates a new parser instance with the given grammar.
// The parser uses a basic recursive descent algorithm.
//
// For better performance and left recursion support,
// use NewImprovedParser instead.
func NewParser(grammar *Grammar) *Parser {
	return &Parser{
		grammar: grammar,
		tokens:  []TokenMatch{},
		pos:     0,
		input:   "", // Will be set during parsing
	}
}

// Parse parses DSL code and returns the result.
// This is the main entry point for the basic parser.
//
// The parsing process:
//  1. Tokenization: Convert input into tokens
//  2. Syntax analysis: Match tokens against grammar rules
//  3. Action execution: Run associated functions
//
// Parameters:
//   - code: Input string to parse
//
// Returns:
//   - Result of the start rule's action
//   - ParseError with line/column info on failure
//
// Example:
//
//	parser := NewParser(grammar)
//	result, err := parser.Parse("2 + 3")
//	if err != nil {
//	    return handleError(err)
//	}
func (p *Parser) Parse(code string) (interface{}, error) {
	// Reset parser state
	p.tokens = []TokenMatch{}
	p.pos = 0
	p.input = code // Store input for error reporting

	// Tokenize
	err := p.tokenize(code)
	if err != nil {
		return nil, err
	}

	// Parse from start rule
	p.pos = 0
	return p.parseRule(p.grammar.startRule)
}

// tokenize converts code into tokens (lexical analysis).
// This process converts the input string into a sequence of tokens
// based on the grammar's token definitions.
//
// The tokenizer:
//   - Skips whitespace automatically
//   - Uses token priority (keywords > regular tokens)
//   - For same priority, longest match wins
//   - Returns detailed error with position on failure
//
// Token priority levels:
//   - 90: Keywords (IF, WHILE, etc.)
//   - 50: Tokens with lookaround
//   - 0: Regular tokens
//
// Example tokenization:
//
//	Input: "if x > 10"
//	Output: [IF, ID("x"), GT, NUMBER("10")]
func (p *Parser) tokenize(code string) error {
	code = strings.TrimSpace(code)
	pos := 0

	for pos < len(code) {
		// Skip whitespace
		if code[pos] == ' ' || code[pos] == '\t' || code[pos] == '\n' || code[pos] == '\r' {
			pos++
			continue
		}

		matched := false
		bestMatch := TokenMatch{}
		bestLength := 0
		bestPriority := -1

		// Find best matching token
		for _, token := range p.grammar.tokens {
			if matches := token.regex.FindStringIndex(code[pos:]); matches != nil && matches[0] == 0 {
				matchLength := matches[1]

				// Higher priority or longer match wins
				shouldReplace := false
				if token.priority > bestPriority {
					shouldReplace = true
				} else if token.priority == bestPriority && matchLength > bestLength {
					shouldReplace = true
				}

				if shouldReplace {
					bestLength = matchLength
					bestPriority = token.priority
					bestMatch = TokenMatch{
						TokenType: token.name,
						Value:     code[pos : pos+matchLength],
						Start:     pos,
						End:       pos + matchLength,
					}
					matched = true
				}
			}
		}

		if matched {
			p.tokens = append(p.tokens, bestMatch)
			pos += bestLength
		} else {
			message := fmt.Sprintf("unexpected character: %c", code[pos])
			return createParseError(message, pos, string(code[pos]), p.input)
		}
	}

	return nil
}

// parseRule attempts to parse a specific grammar rule.
// It tries each alternative in order until one succeeds.
//
// Parameters:
//   - ruleName: Name of the rule to parse
//
// Returns:
//   - Result of the successful alternative's action
//   - ParseError if no alternatives match
//
// The parser backtracks on failure, restoring position
// before trying the next alternative.
//
// Example rule with alternatives:
//
//	expr → expr '+' term  (first alternative)
//	expr → term           (second alternative)
func (p *Parser) parseRule(ruleName string) (interface{}, error) {
	rule, exists := p.grammar.rules[ruleName]
	if !exists {
		return nil, fmt.Errorf("rule %s not found", ruleName)
	}

	// Try each alternative
	for _, alt := range rule.alternatives {
		savedPos := p.pos
		result, err := p.parseAlternative(alt)
		if err == nil {
			return result, nil
		}
		// Restore position if failed
		p.pos = savedPos
	}

	// Create detailed error with current token position
	var token string
	var position int
	if p.pos < len(p.tokens) {
		token = p.tokens[p.pos].Value
		position = p.tokens[p.pos].Start
	} else {
		token = "<end of input>"
		position = len(p.input)
	}

	message := fmt.Sprintf("no alternative matched for rule %s", ruleName)
	return nil, createParseError(message, position, token, p.input)
}

// parseAlternative attempts to parse a rule alternative.
// An alternative is a sequence of symbols (tokens or rules) to match.
//
// Parameters:
//   - alt: The alternative to parse
//
// Returns:
//   - Result of the alternative's action (if defined)
//   - Array of matched values (if no action)
//   - ParseError on failure
//
// Process:
//  1. Match each symbol in sequence
//  2. Collect matched values/results
//  3. Execute action function with collected values
//
// Example:
//
//	Alternative: ["IF", "expr", "THEN", "stmt"]
//	Action: "ifStatement"
//	Results passed to action: ["if", exprResult, "then", stmtResult]
func (p *Parser) parseAlternative(alt *Alternative) (interface{}, error) {
	var results []interface{}

	for _, symbol := range alt.sequence {
		if p.pos >= len(p.tokens) {
			message := "unexpected end of input"
			position := len(p.input)
			return nil, createParseError(message, position, "<end of input>", p.input)
		}

		// Check if symbol is a token
		if _, isToken := p.grammar.tokens[symbol]; isToken {
			if p.tokens[p.pos].TokenType == symbol {
				results = append(results, p.tokens[p.pos].Value)
				p.pos++
			} else {
				message := fmt.Sprintf("expected token %s, got %s", symbol, p.tokens[p.pos].TokenType)
				return nil, createParseError(message, p.tokens[p.pos].Start, p.tokens[p.pos].Value, p.input)
			}
		} else {
			// Symbol is a rule
			result, err := p.parseRule(symbol)
			if err != nil {
				return nil, err
			}
			results = append(results, result)
		}
	}

	// Apply action if available
	if alt.action != "" {
		if action, exists := p.grammar.actions[alt.action]; exists {
			result, err := action(results)
			if err != nil {
				return nil, err
			}
			return result, nil
		}
	}

	return results, nil
}

// Result represents the result of parsing DSL code.
// Contains both the abstract syntax tree and final output.
//
// Fields:
//   - AST: Abstract syntax tree (currently same as Output)
//   - Code: Original input code
//   - Output: Final result after action execution
//   - DSL: Reference to parent DSL for accessing functions
type Result struct {
	AST    interface{} // Abstract syntax tree
	Code   string      // Original input
	Output interface{} // Final evaluation result
	DSL    *DSL        // Reference to DSL for function calls
}

// GetOutput returns the final output of parsing and evaluation.
// This is the result after all actions have been executed.
func (r *Result) GetOutput() interface{} {
	return r.Output
}

// String returns a string representation of the result.
// Format: "DSL[<input>] -> <output>"
//
// Examples:
//
//	"DSL[2+3] -> 5"
//	"DSL[invalid] -> <no result>"
func (r *Result) String() string {
	if r.Output == nil {
		return fmt.Sprintf("DSL[%s] -> <no result>", r.Code)
	}
	return fmt.Sprintf("DSL[%s] -> %v", r.Code, r.Output)
}
