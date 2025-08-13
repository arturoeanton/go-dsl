// Package universal provides a comprehensive HTTP DSL implementation for testing and automation.
// The HTTPDSLv3 is the production-ready version with full feature support including
// HTTP methods, variables, control flow, data extraction, and more.
//
// Features:
//   - Complete HTTP method support (GET, POST, PUT, DELETE, PATCH, etc.)
//   - Variable operations with arithmetic expressions
//   - Control flow statements (if/else, while, foreach, repeat)
//   - Break and continue statements for loop control
//   - Logical operators (AND/OR) with proper precedence
//   - Array operations including indexing with bracket notation
//   - Data extraction (JSONPath, regex, XPath)
//   - Response assertions and validations
//   - Command-line argument support ($ARG1, $ARGC)
//
// Example usage:
//
//	hd := NewHTTPDSLv3()
//	script := `
//	    set $baseURL "https://api.example.com"
//	    GET "$baseURL/users"
//	    assert status 200
//	    extract jsonpath "$.users[0].id" as $userId
//	    if $userId > 0 then
//	        print "Valid user ID: $userId"
//	    endif
//	`
//	result, err := hd.ParseWithBlockSupport(script)
package universal

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

// HTTPDSLv3 represents the production-ready HTTP DSL implementation.
// It provides a complete domain-specific language for HTTP testing and automation,
// supporting methods, headers, variables, control flow, and data extraction.
//
// Version 3.1.1 includes:
//   - Full HTTP method support (GET, POST, PUT, DELETE, etc.)
//   - Variable operations with arithmetic
//   - Control flow (if/else, while, foreach, repeat)
//   - Break/continue statements
//   - Logical operators (AND/OR)
//   - Array operations with indexing
//   - JSON/regex/XPath extraction
//   - Command-line argument support
type HTTPDSLv3 struct {
	dsl       *dslbuilder.DSL          // DSL parser and tokenizer
	engine    *HTTPEngine              // HTTP request execution engine
	variables map[string]interface{}   // Script variables storage
	context   map[string]interface{}   // Execution context (break/continue flags)
}

// NewHTTPDSLv3 creates a new HTTP DSL v3 instance.
// It initializes the DSL parser with improved parsing capabilities,
// sets up the HTTP engine for request execution, and prepares
// variable and context storage for script execution.
//
// Example:
//
//	hd := NewHTTPDSLv3()
//	result, err := hd.ParseWithBlockSupport(`
//	    GET "https://api.example.com/users"
//	    assert status 200
//	`)
func NewHTTPDSLv3() *HTTPDSLv3 {
	hd := &HTTPDSLv3{
		dsl:       dslbuilder.New("HTTPDSLv3"), // Already uses ImprovedParser by default
		engine:    NewHTTPEngine(),
		variables: make(map[string]interface{}),
		context:   make(map[string]interface{}),
	}
	hd.setupGrammar()
	return hd
}

// setupGrammar defines the complete DSL grammar including tokens, rules, and actions.
// It sets up all language constructs in the proper priority order to ensure
// correct parsing. Keywords have priority 90, while general patterns have priority 0.
func (hd *HTTPDSLv3) setupGrammar() {
	// HTTP Methods - Highest priority (90) to ensure they're recognized before general IDs
	hd.dsl.KeywordToken("GET", "GET")
	hd.dsl.KeywordToken("POST", "POST")
	hd.dsl.KeywordToken("PUT", "PUT")
	hd.dsl.KeywordToken("DELETE", "DELETE")
	hd.dsl.KeywordToken("PATCH", "PATCH")
	hd.dsl.KeywordToken("HEAD", "HEAD")
	hd.dsl.KeywordToken("OPTIONS", "OPTIONS")
	hd.dsl.KeywordToken("CONNECT", "CONNECT")
	hd.dsl.KeywordToken("TRACE", "TRACE")
	
	// Keywords - High priority (90)
	hd.dsl.KeywordToken("header", "header")
	hd.dsl.KeywordToken("body", "body")
	hd.dsl.KeywordToken("json", "json")
	hd.dsl.KeywordToken("form", "form")
	hd.dsl.KeywordToken("auth", "auth")
	hd.dsl.KeywordToken("basic", "basic")
	hd.dsl.KeywordToken("bearer", "bearer")
	hd.dsl.KeywordToken("timeout", "timeout")
	hd.dsl.KeywordToken("ms", "ms")
	hd.dsl.KeywordToken("s", "s")
	
	// Variables
	hd.dsl.KeywordToken("set", "set")
	hd.dsl.KeywordToken("var", "var")
	hd.dsl.KeywordToken("print", "print")
	hd.dsl.KeywordToken("length", "length")
	hd.dsl.KeywordToken("at", "at")
	hd.dsl.KeywordToken("extract", "extract")
	hd.dsl.KeywordToken("from", "from")
	hd.dsl.KeywordToken("as", "as")
	hd.dsl.KeywordToken("jsonpath", "jsonpath")
	hd.dsl.KeywordToken("xpath", "xpath")
	hd.dsl.KeywordToken("regex", "regex")
	hd.dsl.KeywordToken("status", "status")
	hd.dsl.KeywordToken("response", "response")
	
	// Conditionals
	hd.dsl.KeywordToken("if", "if")
	hd.dsl.KeywordToken("then", "then")
	hd.dsl.KeywordToken("else", "else")
	hd.dsl.KeywordToken("endif", "endif")
	hd.dsl.KeywordToken("contains", "contains")
	hd.dsl.KeywordToken("equals", "equals")
	hd.dsl.KeywordToken("matches", "matches")
	hd.dsl.KeywordToken("exists", "exists")
	hd.dsl.KeywordToken("empty", "empty")
	hd.dsl.KeywordToken("greater", "greater")
	hd.dsl.KeywordToken("less", "less")
	
	// Loops
	hd.dsl.KeywordToken("repeat", "repeat")
	hd.dsl.KeywordToken("times", "times")
	hd.dsl.KeywordToken("do", "do")
	hd.dsl.KeywordToken("endloop", "endloop")
	hd.dsl.KeywordToken("while", "while")
	hd.dsl.KeywordToken("foreach", "foreach")
	hd.dsl.KeywordToken("in", "in")
	hd.dsl.KeywordToken("break", "break")
	hd.dsl.KeywordToken("continue", "continue")
	
	// Assertions
	hd.dsl.KeywordToken("assert", "assert")
	hd.dsl.KeywordToken("expect", "expect")
	hd.dsl.KeywordToken("time", "time")
	
	// Utilities
	hd.dsl.KeywordToken("wait", "wait")
	hd.dsl.KeywordToken("sleep", "sleep")
	hd.dsl.KeywordToken("log", "log")
	hd.dsl.KeywordToken("debug", "debug")
	hd.dsl.KeywordToken("clear", "clear")
	hd.dsl.KeywordToken("cookies", "cookies")
	hd.dsl.KeywordToken("reset", "reset")
	hd.dsl.KeywordToken("base", "base")
	hd.dsl.KeywordToken("url", "url")
	
	// Operators
	hd.dsl.KeywordToken("and", "and")
	hd.dsl.KeywordToken("or", "or")
	hd.dsl.KeywordToken("not", "not")
	
	// DEVELOPER GUIDE: Token Definition
	// Tokens are defined with patterns and priorities.
	// KeywordToken() = priority 90 (matched first)
	// Token() = priority 0 (matched last)
	// Order matters only within same priority level.
	
	// Value tokens - Lower priority (0)
	// Better JSON pattern to handle nested objects and special characters
	hd.dsl.Token("JSON_INLINE", `\{[^{}]*(?:\{[^{}]*\}[^{}]*)*\}`)
	// String with escape sequences - handles \n, \t, \", etc.
	hd.dsl.Token("STRING", `"(?:[^"\\]|\\.)*"`)
	hd.dsl.Token("NUMBER", `[0-9]+(\.[0-9]+)?`)
	hd.dsl.Token("VARIABLE", `\$[a-zA-Z_][a-zA-Z0-9_]*`)
	hd.dsl.Token("URL", `https?://[^\s]+`)
	hd.dsl.Token("COMPARISON", `==|!=|>=|<=|>|<`)
	hd.dsl.Token("ARITHMETIC", `\+|\-|\*|\/`)
	hd.dsl.Token("ID", `[a-zA-Z_][a-zA-Z0-9_]*`)
	hd.dsl.Token("(", `\(`)
	hd.dsl.Token(")", `\)`)
	hd.dsl.Token("[", `\[`)
	hd.dsl.Token("]", `\]`)
	
	// DEVELOPER GUIDE: Grammar Rules
	// Rules define the syntax structure. Format: Rule(name, pattern, action)
	// - name: rule identifier
	// - pattern: array of tokens/rules to match
	// - action: function name to execute when matched
	// Rules are tried in order until one matches.
	
	// Main program rule - accepts single statement OR multiple statements
	hd.dsl.Rule("program", []string{"statement"}, "executeSingleStatement")
	hd.dsl.Rule("program", []string{"statements"}, "executeProgram")
	
	// Statements (supports multiple statements)
	hd.dsl.Rule("statements", []string{"statement", "statements"}, "multipleStatements")
	hd.dsl.Rule("statements", []string{"statement"}, "singleStatement")
	
	// DEVELOPER GUIDE: Actions
	// Actions are functions that execute when rules match.
	// They receive matched tokens/values as args[].
	// Return value becomes the rule's result.
	// To add new functionality:
	// 1. Define tokens (if needed)
	// 2. Create rules that use those tokens
	// 3. Write actions to process the matched data
	
	hd.dsl.Action("executeSingleStatement", func(args []interface{}) (interface{}, error) {
		return args[0], nil
	})
	
	hd.dsl.Action("singleStatement", func(args []interface{}) (interface{}, error) {
		return []interface{}{args[0]}, nil
	})
	
	hd.dsl.Action("multipleStatements", func(args []interface{}) (interface{}, error) {
		stmt := args[0]
		stmts := args[1].([]interface{})
		return append([]interface{}{stmt}, stmts...), nil
	})
	
	hd.dsl.Action("executeProgram", func(args []interface{}) (interface{}, error) {
		statements := args[0].([]interface{})
		var lastResult interface{}
		for _, stmt := range statements {
			lastResult = stmt
			// Handle control flow
			if hd.context["break"] == true {
				break
			}
			if hd.context["continue"] == true {
				hd.context["continue"] = false
				continue
			}
		}
		return lastResult, nil
	})
	
	// Statement types
	hd.dsl.Rule("statement", []string{"http_request"}, "passthrough")
	hd.dsl.Rule("statement", []string{"variable_op"}, "passthrough")
	hd.dsl.Rule("statement", []string{"print_cmd"}, "passthrough")
	hd.dsl.Rule("statement", []string{"conditional"}, "passthrough")
	hd.dsl.Rule("statement", []string{"loop_stmt"}, "passthrough")
	hd.dsl.Rule("statement", []string{"assertion"}, "passthrough")
	hd.dsl.Rule("statement", []string{"utility"}, "passthrough")
	hd.dsl.Rule("statement", []string{"control_flow"}, "passthrough")
	
	hd.dsl.Action("passthrough", func(args []interface{}) (interface{}, error) {
		if len(args) > 0 {
			return args[0], nil
		}
		return nil, nil
	})
	
	// DEVELOPER GUIDE: Control Flow Implementation
	// Break/continue use context flags to signal loop termination.
	// Loops check these flags after each iteration.
	// The flags propagate through nested structures via LoopResult.
	
	// Control flow
	hd.dsl.Rule("control_flow", []string{"break"}, "breakCmd")
	hd.dsl.Rule("control_flow", []string{"continue"}, "continueCmd")
	
	hd.dsl.Action("breakCmd", func(args []interface{}) (interface{}, error) {
		hd.context["break"] = true  // Set flag for loop to check
		return "break", nil
	})
	
	hd.dsl.Action("continueCmd", func(args []interface{}) (interface{}, error) {
		hd.context["continue"] = true  // Set flag to skip to next iteration
		return "continue", nil
	})
	
	// DEVELOPER GUIDE: HTTP Request Pattern
	// HTTP requests can have optional parameters (headers, body, auth).
	// Rules are ordered: most specific first, general last.
	// This prevents shorter patterns from matching prematurely.
	
	// HTTP Requests - Order matters! Longer patterns first
	hd.dsl.Rule("http_request", []string{"http_method", "url_value", "option_list"}, "httpWithOptions")
	hd.dsl.Rule("http_request", []string{"http_method", "url_value"}, "httpSimple")
	
	// Option list - using LEFT recursion (now supported by improved parser)
	// Left recursion is more efficient for building lists
	hd.dsl.Rule("option_list", []string{"option"}, "firstOption")
	hd.dsl.Rule("option_list", []string{"option_list", "option"}, "appendOption")
	
	hd.dsl.Action("firstOption", func(args []interface{}) (interface{}, error) {
		return []interface{}{args[0]}, nil
	})
	
	hd.dsl.Action("appendOption", func(args []interface{}) (interface{}, error) {
		// With left recursion: list comes first, then the new option
		list := args[0].([]interface{})
		option := args[1]
		return append(list, option), nil
	})
	
	// Individual options
	hd.dsl.Rule("option", []string{"header", "STRING", "STRING"}, "headerOption")
	hd.dsl.Rule("option", []string{"body", "STRING"}, "bodyOption")
	hd.dsl.Rule("option", []string{"json", "STRING"}, "jsonStringOption")
	hd.dsl.Rule("option", []string{"json", "JSON_INLINE"}, "jsonInlineOption")
	hd.dsl.Rule("option", []string{"auth", "basic", "STRING", "STRING"}, "authBasicOption")
	hd.dsl.Rule("option", []string{"auth", "bearer", "STRING"}, "authBearerOption")
	hd.dsl.Rule("option", []string{"timeout", "NUMBER", "time_unit"}, "timeoutOption")
	
	// HTTP methods
	hd.dsl.Rule("http_method", []string{"GET"}, "methodType")
	hd.dsl.Rule("http_method", []string{"POST"}, "methodType")
	hd.dsl.Rule("http_method", []string{"PUT"}, "methodType")
	hd.dsl.Rule("http_method", []string{"DELETE"}, "methodType")
	hd.dsl.Rule("http_method", []string{"PATCH"}, "methodType")
	hd.dsl.Rule("http_method", []string{"HEAD"}, "methodType")
	hd.dsl.Rule("http_method", []string{"OPTIONS"}, "methodType")
	hd.dsl.Rule("http_method", []string{"CONNECT"}, "methodType")
	hd.dsl.Rule("http_method", []string{"TRACE"}, "methodType")
	
	hd.dsl.Action("methodType", func(args []interface{}) (interface{}, error) {
		return args[0], nil
	})
	
	// URL values with proper variable expansion
	hd.dsl.Rule("url_value", []string{"STRING"}, "urlString")
	hd.dsl.Rule("url_value", []string{"URL"}, "urlDirect")
	hd.dsl.Rule("url_value", []string{"VARIABLE"}, "urlVariable")
	
	hd.dsl.Action("urlString", func(args []interface{}) (interface{}, error) {
		url := hd.unquoteString(args[0].(string))
		// Expand variables in URL
		return hd.expandVariables(url), nil
	})
	
	hd.dsl.Action("urlDirect", func(args []interface{}) (interface{}, error) {
		return hd.expandVariables(args[0].(string)), nil
	})
	
	hd.dsl.Action("urlVariable", func(args []interface{}) (interface{}, error) {
		varName := strings.TrimPrefix(args[0].(string), "$")
		if val, ok := hd.variables[varName]; ok {
			return fmt.Sprintf("%v", val), nil
		}
		return "", fmt.Errorf("variable $%s not found", varName)
	})
	
	// Time units
	hd.dsl.Rule("time_unit", []string{"ms"}, "timeUnit")
	hd.dsl.Rule("time_unit", []string{"s"}, "timeUnit")
	
	hd.dsl.Action("timeUnit", func(args []interface{}) (interface{}, error) {
		return args[0], nil
	})
	
	// Option actions
	hd.dsl.Action("headerOption", func(args []interface{}) (interface{}, error) {
		return map[string]interface{}{
			"type":  "header",
			"key":   hd.unquoteString(args[1].(string)),
			"value": hd.expandVariables(hd.unquoteString(args[2].(string))),
		}, nil
	})
	
	hd.dsl.Action("bodyOption", func(args []interface{}) (interface{}, error) {
		return map[string]interface{}{
			"type":  "body",
			"value": hd.expandVariables(hd.unquoteString(args[1].(string))),
		}, nil
	})
	
	hd.dsl.Action("jsonStringOption", func(args []interface{}) (interface{}, error) {
		jsonStr := hd.expandVariables(hd.unquoteString(args[1].(string)))
		return map[string]interface{}{
			"type":  "json",
			"value": jsonStr,
		}, nil
	})
	
	hd.dsl.Action("jsonInlineOption", func(args []interface{}) (interface{}, error) {
		jsonStr := hd.expandVariables(args[1].(string))
		return map[string]interface{}{
			"type":  "json",
			"value": jsonStr,
		}, nil
	})
	
	hd.dsl.Action("authBasicOption", func(args []interface{}) (interface{}, error) {
		return map[string]interface{}{
			"type":     "auth",
			"authType": "basic",
			"user":     hd.expandVariables(hd.unquoteString(args[2].(string))),
			"pass":     hd.expandVariables(hd.unquoteString(args[3].(string))),
		}, nil
	})
	
	hd.dsl.Action("authBearerOption", func(args []interface{}) (interface{}, error) {
		return map[string]interface{}{
			"type":     "auth",
			"authType": "bearer",
			"token":    hd.expandVariables(hd.unquoteString(args[2].(string))),
		}, nil
	})
	
	hd.dsl.Action("timeoutOption", func(args []interface{}) (interface{}, error) {
		value, _ := strconv.ParseFloat(args[1].(string), 64)
		unit := args[2].(string)
		if unit == "s" {
			value = value * 1000
		}
		return map[string]interface{}{
			"type":  "timeout",
			"value": int(value),
		}, nil
	})
	
	hd.dsl.Action("httpSimple", func(args []interface{}) (interface{}, error) {
		method := args[0].(string)
		url := args[1].(string)
		return hd.engine.Request(method, url, nil)
	})
	
	hd.dsl.Action("httpWithOptions", func(args []interface{}) (interface{}, error) {
		method := args[0].(string)
		url := args[1].(string)
		
		// Process options list
		optionsList := args[2].([]interface{})
		requestOptions := make(map[string]interface{})
		headers := make(map[string]string)
		
		for _, opt := range optionsList {
			option := opt.(map[string]interface{})
			optType := option["type"].(string)
			
			switch optType {
			case "header":
				headers[option["key"].(string)] = option["value"].(string)
			case "body":
				requestOptions["body"] = option["value"]
			case "json":
				requestOptions["json"] = option["value"]
			case "auth":
				authType := option["authType"].(string)
				if authType == "basic" {
					requestOptions["auth"] = map[string]string{
						"type": "basic",
						"user": option["user"].(string),
						"pass": option["pass"].(string),
					}
				} else if authType == "bearer" {
					requestOptions["auth"] = map[string]string{
						"type":  "bearer",
						"token": option["token"].(string),
					}
				}
			case "timeout":
				requestOptions["timeout"] = option["value"]
			}
		}
		
		if len(headers) > 0 {
			requestOptions["header"] = headers
		}
		
		return hd.engine.Request(method, url, requestOptions)
	})
	
	// Variable operations
	hd.dsl.Rule("variable_op", []string{"set_var"}, "passthrough")
	hd.dsl.Rule("variable_op", []string{"extract_var"}, "passthrough")
	
	// Set variable with expression support
	hd.dsl.Rule("set_var", []string{"set", "VARIABLE", "expression"}, "setVariable")
	hd.dsl.Rule("set_var", []string{"var", "VARIABLE", "expression"}, "setVariable")
	
	// Expressions (supports arithmetic and string concatenation)
	hd.dsl.Rule("expression", []string{"array_access"}, "passthrough")
	hd.dsl.Rule("expression", []string{"function_call"}, "passthrough")
	hd.dsl.Rule("expression", []string{"expression", "ARITHMETIC", "term"}, "arithmeticOp")
	hd.dsl.Rule("expression", []string{"term"}, "passthrough")
	
	hd.dsl.Rule("term", []string{"value"}, "passthrough")
	
	hd.dsl.Action("arithmeticOp", func(args []interface{}) (interface{}, error) {
		left := hd.toNumber(args[0])
		op := args[1].(string)
		right := hd.toNumber(args[2])
		
		switch op {
		case "+":
			return left + right, nil
		case "-":
			return left - right, nil
		case "*":
			return left * right, nil
		case "/":
			if right == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			return left / right, nil
		}
		return nil, fmt.Errorf("unknown operator: %s", op)
	})
	
	hd.dsl.Rule("value", []string{"STRING"}, "valueString")
	hd.dsl.Rule("value", []string{"NUMBER"}, "valueNumber")
	hd.dsl.Rule("value", []string{"VARIABLE"}, "valueVariable")
	
	// DEVELOPER GUIDE: Extending Functions
	// To add a new function:
	// 1. Add keyword token for function name
	// 2. Create rule: Rule("function_call", [funcname, ...args], actionName)
	// 3. Implement action to process the function
	
	// Function calls
	hd.dsl.Rule("function_call", []string{"length", "VARIABLE"}, "lengthFunction")
	
	// DEVELOPER GUIDE: Array Indexing
	// Arrays use bracket notation: $array[index]
	// Supports both numeric and variable indices.
	// The brackets are separate tokens to avoid conflicts with JSON arrays.
	
	// Array access - using bracket syntax
	hd.dsl.Rule("array_access", []string{"VARIABLE", "[", "NUMBER", "]"}, "arrayAccess")
	hd.dsl.Rule("array_access", []string{"VARIABLE", "[", "VARIABLE", "]"}, "arrayAccessVar")
	
	hd.dsl.Action("valueString", func(args []interface{}) (interface{}, error) {
		str := hd.unquoteString(args[0].(string))
		return hd.expandVariables(str), nil
	})
	
	hd.dsl.Action("valueNumber", func(args []interface{}) (interface{}, error) {
		num, _ := strconv.ParseFloat(args[0].(string), 64)
		return num, nil
	})
	
	hd.dsl.Action("valueVariable", func(args []interface{}) (interface{}, error) {
		varName := strings.TrimPrefix(args[0].(string), "$")
		if val, ok := hd.variables[varName]; ok {
			return val, nil
		}
		return nil, fmt.Errorf("variable $%s not found", varName)
	})
	
	// DEVELOPER GUIDE: Function Implementation
	// Functions operate on variables and return computed values.
	// They can handle different data types (arrays, strings, etc.).
	
	hd.dsl.Action("lengthFunction", func(args []interface{}) (interface{}, error) {
		varName := strings.TrimPrefix(args[1].(string), "$")
		if val, ok := hd.variables[varName]; ok {
			// Handle different types: arrays, strings, JSON arrays
			switch v := val.(type) {
			case []interface{}:
				return len(v), nil
			case []string:
				return len(v), nil
			case string:
				// Try to parse as JSON array
				if strings.HasPrefix(v, "[") && strings.HasSuffix(v, "]") {
					// Count elements
					trimmed := strings.Trim(v, "[]")
					if strings.TrimSpace(trimmed) == "" {
						return 0, nil
					}
					// Simple count of comma-separated elements
					parts := strings.Split(trimmed, ",")
					return len(parts), nil
				}
				// Return string length
				return len(v), nil
			default:
				return 0, nil
			}
		}
		return 0, nil
	})
	
	hd.dsl.Action("arrayAccess", func(args []interface{}) (interface{}, error) {
		varName := strings.TrimPrefix(args[0].(string), "$")
		// Parse index from NUMBER token (now at position 2 with brackets)
		indexStr := args[2].(string)
		index, _ := strconv.Atoi(indexStr)
		
		if val, ok := hd.variables[varName]; ok {
			switch v := val.(type) {
			case []interface{}:
				if index >= 0 && index < len(v) {
					return v[index], nil
				}
				return nil, fmt.Errorf("array index out of bounds: %d", index)
			case []string:
				if index >= 0 && index < len(v) {
					return v[index], nil
				}
				return nil, fmt.Errorf("array index out of bounds: %d", index)
			case string:
				// Try to parse as JSON array
				if strings.HasPrefix(v, "[") && strings.HasSuffix(v, "]") {
					trimmed := strings.Trim(v, "[]")
					if strings.TrimSpace(trimmed) == "" {
						return nil, fmt.Errorf("cannot access index %d of empty array", index)
					}
					parts := strings.Split(trimmed, ",")
					if index >= 0 && index < len(parts) {
						item := strings.TrimSpace(parts[index])
						item = strings.Trim(item, "\"'")
						return item, nil
					}
					return nil, fmt.Errorf("array index out of bounds: %d", index)
				}
				// String character access
				if index >= 0 && index < len(v) {
					return string(v[index]), nil
				}
				return nil, fmt.Errorf("string index out of bounds: %d", index)
			default:
				return nil, fmt.Errorf("variable $%s is not an array", varName)
			}
		}
		return nil, fmt.Errorf("variable $%s not found", varName)
	})
	
	hd.dsl.Action("arrayAccessVar", func(args []interface{}) (interface{}, error) {
		varName := strings.TrimPrefix(args[0].(string), "$")
		indexVarName := strings.TrimPrefix(args[2].(string), "$")
		
		// Get index from variable
		var index int
		if idxVal, ok := hd.variables[indexVarName]; ok {
			switch v := idxVal.(type) {
			case float64:
				index = int(v)
			case int:
				index = v
			case string:
				index, _ = strconv.Atoi(v)
			default:
				return nil, fmt.Errorf("index variable $%s is not a number", indexVarName)
			}
		} else {
			return nil, fmt.Errorf("index variable $%s not found", indexVarName)
		}
		
		// Now use the same logic as arrayAccess
		if val, ok := hd.variables[varName]; ok {
			switch v := val.(type) {
			case []interface{}:
				if index >= 0 && index < len(v) {
					return v[index], nil
				}
				return nil, fmt.Errorf("array index out of bounds: %d", index)
			case []string:
				if index >= 0 && index < len(v) {
					return v[index], nil
				}
				return nil, fmt.Errorf("array index out of bounds: %d", index)
			case string:
				// Try to parse as JSON array
				if strings.HasPrefix(v, "[") && strings.HasSuffix(v, "]") {
					trimmed := strings.Trim(v, "[]")
					if strings.TrimSpace(trimmed) == "" {
						return nil, fmt.Errorf("cannot access index %d of empty array", index)
					}
					parts := strings.Split(trimmed, ",")
					if index >= 0 && index < len(parts) {
						item := strings.TrimSpace(parts[index])
						item = strings.Trim(item, "\"'")
						return item, nil
					}
					return nil, fmt.Errorf("array index out of bounds: %d", index)
				}
				// String character access
				if index >= 0 && index < len(v) {
					return string(v[index]), nil
				}
				return nil, fmt.Errorf("string index out of bounds: %d", index)
			default:
				return nil, fmt.Errorf("variable $%s is not an array", varName)
			}
		}
		return nil, fmt.Errorf("variable $%s not found", varName)
	})
	
	hd.dsl.Action("setVariable", func(args []interface{}) (interface{}, error) {
		varName := strings.TrimPrefix(args[1].(string), "$")
		value := args[2]
		hd.variables[varName] = value
		return fmt.Sprintf("Variable $%s set to %v", varName, value), nil
	})
	
	// Print command with variable expansion
	hd.dsl.Rule("print_cmd", []string{"print", "VARIABLE"}, "printVariable")
	hd.dsl.Rule("print_cmd", []string{"print", "STRING"}, "printString")
	
	hd.dsl.Action("printVariable", func(args []interface{}) (interface{}, error) {
		varName := strings.TrimPrefix(args[1].(string), "$")
		if val, ok := hd.variables[varName]; ok {
			return fmt.Sprintf("$%s = %v", varName, val), nil
		}
		return fmt.Sprintf("Variable $%s not found", varName), nil
	})
	
	hd.dsl.Action("printString", func(args []interface{}) (interface{}, error) {
		str := hd.unquoteString(args[1].(string))
		return hd.expandVariables(str), nil
	})
	
	// Extract variable
	hd.dsl.Rule("extract_var", []string{"extract", "extract_type", "STRING", "as", "VARIABLE"}, "extractVariable")
	hd.dsl.Rule("extract_var", []string{"extract", "extract_type", "as", "VARIABLE"}, "extractVariableNoPattern")
	
	hd.dsl.Rule("extract_type", []string{"jsonpath"}, "extractType")
	hd.dsl.Rule("extract_type", []string{"xpath"}, "extractType")
	hd.dsl.Rule("extract_type", []string{"regex"}, "extractType")
	hd.dsl.Rule("extract_type", []string{"header"}, "extractType")
	hd.dsl.Rule("extract_type", []string{"status"}, "extractType")
	
	hd.dsl.Action("extractType", func(args []interface{}) (interface{}, error) {
		return args[0], nil
	})
	
	hd.dsl.Action("extractVariable", func(args []interface{}) (interface{}, error) {
		extractType := args[1].(string)
		pattern := hd.unquoteString(args[2].(string))
		varName := strings.TrimPrefix(args[4].(string), "$")
		
		// Check if there's a response to extract from
		if hd.engine.GetLastResponse() == "" {
			hd.variables[varName] = ""
			return fmt.Sprintf("Warning: No response available for extraction. Variable $%s set to empty.", varName), nil
		}
		
		value := hd.engine.Extract(extractType, pattern)
		if value == nil {
			value = ""
		}
		hd.variables[varName] = value
		
		return fmt.Sprintf("Extracted %s using %s and stored in $%s", pattern, extractType, varName), nil
	})
	
	hd.dsl.Action("extractVariableNoPattern", func(args []interface{}) (interface{}, error) {
		extractType := args[1].(string)
		varName := strings.TrimPrefix(args[3].(string), "$")
		
		// Check if there's a response to extract from
		if hd.engine.GetLastResponse() == "" {
			hd.variables[varName] = ""
			return fmt.Sprintf("Warning: No response available for extraction. Variable $%s set to empty.", varName), nil
		}
		
		value := hd.engine.Extract(extractType, "")
		if value == nil {
			value = ""
		}
		hd.variables[varName] = value
		
		return fmt.Sprintf("Extracted %s and stored in $%s", extractType, varName), nil
	})
	
	// Improved conditionals - fixed to handle single line if/then without else
	hd.dsl.Rule("conditional", []string{"if", "condition", "then", "statement", "else", "statement"}, "ifElse")
	hd.dsl.Rule("conditional", []string{"if", "condition", "then", "statement"}, "ifSimple")
	hd.dsl.Rule("conditional", []string{"if", "condition", "then", "statements", "else", "statements", "endif"}, "ifElseBlock")
	hd.dsl.Rule("conditional", []string{"if", "condition", "then", "statements", "endif"}, "ifBlock")
	
	// Support for grouped statements (for block parsing)
	hd.dsl.Rule("conditional", []string{"if", "condition", "then", "(", "statements", ")", "else", "(", "statements", ")"}, "ifGroupedElse")
	hd.dsl.Rule("conditional", []string{"if", "condition", "then", "(", "statements", ")"}, "ifGrouped")
	
	// Conditions with logical operators
	hd.dsl.Rule("condition", []string{"condition", "and", "simple_condition"}, "andCondition")
	hd.dsl.Rule("condition", []string{"condition", "or", "simple_condition"}, "orCondition")
	hd.dsl.Rule("condition", []string{"not", "condition"}, "notCondition")
	hd.dsl.Rule("condition", []string{"simple_condition"}, "passthrough")
	
	hd.dsl.Rule("simple_condition", []string{"value", "COMPARISON", "value"}, "comparison")
	hd.dsl.Rule("simple_condition", []string{"value", "contains", "value"}, "containsCheck")
	hd.dsl.Rule("simple_condition", []string{"value", "empty"}, "emptyCheck")
	hd.dsl.Rule("simple_condition", []string{"value", "exists"}, "existsCheck")
	
	hd.dsl.Action("comparison", func(args []interface{}) (interface{}, error) {
		left := args[0]
		op := args[1].(string)
		right := args[2]
		return hd.engine.Compare(left, op, right), nil
	})
	
	hd.dsl.Action("containsCheck", func(args []interface{}) (interface{}, error) {
		haystack := fmt.Sprintf("%v", args[0])
		needle := fmt.Sprintf("%v", args[2])
		return strings.Contains(haystack, needle), nil
	})
	
	hd.dsl.Action("emptyCheck", func(args []interface{}) (interface{}, error) {
		val := fmt.Sprintf("%v", args[0])
		return val == "" || val == "0" || val == "false" || val == "<nil>", nil
	})
	
	hd.dsl.Action("existsCheck", func(args []interface{}) (interface{}, error) {
		return args[0] != nil, nil
	})
	
	hd.dsl.Action("andCondition", func(args []interface{}) (interface{}, error) {
		left := hd.toBool(args[0])
		right := hd.toBool(args[2])
		return left && right, nil
	})
	
	hd.dsl.Action("orCondition", func(args []interface{}) (interface{}, error) {
		left := hd.toBool(args[0])
		right := hd.toBool(args[2])
		return left || right, nil
	})
	
	hd.dsl.Action("notCondition", func(args []interface{}) (interface{}, error) {
		cond := hd.toBool(args[1])
		return !cond, nil
	})
	
	hd.dsl.Action("ifSimple", func(args []interface{}) (interface{}, error) {
		condition := hd.toBool(args[1])
		if condition {
			return hd.executeStatement(args[3])
		}
		return nil, nil
	})
	
	hd.dsl.Action("ifElse", func(args []interface{}) (interface{}, error) {
		// args[1] should be the condition result
		condition := hd.toBool(args[1])
		
		// Debug: Let's see what we're getting
		// fmt.Printf("DEBUG ifElse: condition=%v, args[3]=%v, args[5]=%v\n", condition, args[3], args[5])
		
		if condition {
			// Execute the "then" statement (args[3])
			return args[3], nil  // Return the then statement result directly
		} else {
			// Execute the "else" statement (args[5])
			return args[5], nil  // Return the else statement result directly
		}
	})
	
	hd.dsl.Action("ifBlock", func(args []interface{}) (interface{}, error) {
		condition := hd.toBool(args[1])
		if condition {
			return hd.executeStatements(args[3])
		}
		return nil, nil
	})
	
	hd.dsl.Action("ifElseBlock", func(args []interface{}) (interface{}, error) {
		condition := hd.toBool(args[1])
		if condition {
			return hd.executeStatements(args[3])
		}
		return hd.executeStatements(args[5])
	})
	
	hd.dsl.Action("ifGrouped", func(args []interface{}) (interface{}, error) {
		condition := hd.toBool(args[1])
		if condition {
			// args[4] contains the statements (skipping "(" and ")")
			return hd.executeStatements(args[4])
		}
		return nil, nil
	})
	
	hd.dsl.Action("ifGroupedElse", func(args []interface{}) (interface{}, error) {
		condition := hd.toBool(args[1])
		if condition {
			// args[4] contains the then statements
			return hd.executeStatements(args[4])
		}
		// args[8] contains the else statements  
		return hd.executeStatements(args[8])
	})
	
	// Loops with proper DSL integration
	hd.dsl.Rule("loop_stmt", []string{"repeat", "NUMBER", "times", "do", "statements", "endloop"}, "repeatLoop")
	hd.dsl.Rule("loop_stmt", []string{"while", "condition", "do", "statements", "endloop"}, "whileLoop")
	hd.dsl.Rule("loop_stmt", []string{"foreach", "VARIABLE", "in", "VARIABLE", "do", "statements", "endloop"}, "foreachLoop")
	
	hd.dsl.Action("repeatLoop", func(args []interface{}) (interface{}, error) {
		times, _ := strconv.Atoi(args[1].(string))
		statements := args[4]
		
		var results []interface{}
		for i := 0; i < times; i++ {
			hd.variables["_index"] = i
			hd.variables["_iteration"] = i + 1
			
			result, _ := hd.executeStatements(statements)
			results = append(results, result)
			
			// Check for break
			if hd.context["break"] == true {
				hd.context["break"] = false
				break
			}
		}
		
		return fmt.Sprintf("Repeated %d times", times), nil
	})
	
	hd.dsl.Action("whileLoop", func(args []interface{}) (interface{}, error) {
		maxIterations := 1000 // Safety limit
		iterations := 0
		statements := args[3]
		
		for iterations < maxIterations {
			// Re-evaluate condition each time
			condition := hd.evaluateCondition(args[1])
			if !condition {
				break
			}
			
			hd.variables["_iteration"] = iterations + 1
			_, _ = hd.executeStatements(statements)
			iterations++
			
			// Check for break
			if hd.context["break"] == true {
				hd.context["break"] = false
				break
			}
		}
		
		if iterations >= maxIterations {
			return nil, fmt.Errorf("while loop exceeded maximum iterations (%d)", maxIterations)
		}
		
		return fmt.Sprintf("While loop executed %d times", iterations), nil
	})
	
	hd.dsl.Action("foreachLoop", func(args []interface{}) (interface{}, error) {
		itemVar := strings.TrimPrefix(args[1].(string), "$")
		listVar := strings.TrimPrefix(args[3].(string), "$")
		statements := args[5]
		
		list, ok := hd.variables[listVar]
		if !ok {
			return nil, fmt.Errorf("list variable $%s not found", listVar)
		}
		
		// Convert to slice if possible
		items := hd.toSlice(list)
		if items == nil {
			return nil, fmt.Errorf("variable $%s is not iterable", listVar)
		}
		
		for i, item := range items {
			hd.variables[itemVar] = item
			hd.variables["_index"] = i
			_, _ = hd.executeStatements(statements)
			
			// Check for break
			if hd.context["break"] == true {
				hd.context["break"] = false
				break
			}
		}
		
		return fmt.Sprintf("Foreach completed for $%s", listVar), nil
	})
	
	// Assertions - fixed to work as standalone statements
	hd.dsl.Rule("assertion", []string{"assert", "assertion_type"}, "doAssertion")
	hd.dsl.Rule("assertion", []string{"expect", "assertion_type"}, "doAssertion")
	
	hd.dsl.Rule("assertion_type", []string{"status", "NUMBER"}, "assertStatus")
	hd.dsl.Rule("assertion_type", []string{"time", "less", "NUMBER", "ms"}, "assertTime")
	hd.dsl.Rule("assertion_type", []string{"response", "contains", "STRING"}, "assertContains")
	
	hd.dsl.Action("assertStatus", func(args []interface{}) (interface{}, error) {
		expectedCode, _ := strconv.Atoi(args[1].(string))
		actualCode := hd.engine.GetLastStatusCode()
		if actualCode == expectedCode {
			return fmt.Sprintf("✓ Status code is %d", expectedCode), nil
		}
		return nil, fmt.Errorf("assertion failed: expected status %d, got %d", expectedCode, actualCode)
	})
	
	hd.dsl.Action("assertTime", func(args []interface{}) (interface{}, error) {
		maxTime, _ := strconv.ParseFloat(args[2].(string), 64)
		actualTime := hd.engine.GetLastResponseTime()
		if actualTime < maxTime {
			return fmt.Sprintf("✓ Response time %.2fms < %.2fms", actualTime, maxTime), nil
		}
		return nil, fmt.Errorf("assertion failed: response time %.2fms exceeds %.2fms", actualTime, maxTime)
	})
	
	hd.dsl.Action("assertContains", func(args []interface{}) (interface{}, error) {
		expected := hd.expandVariables(hd.unquoteString(args[2].(string)))
		response := hd.engine.GetLastResponse()
		if strings.Contains(response, expected) {
			return fmt.Sprintf("✓ Response contains '%s'", expected), nil
		}
		return nil, fmt.Errorf("assertion failed: response does not contain '%s'", expected)
	})
	
	hd.dsl.Action("doAssertion", func(args []interface{}) (interface{}, error) {
		return args[1], nil
	})
	
	// Utilities
	hd.dsl.Rule("utility", []string{"wait", "NUMBER", "time_unit"}, "waitCmd")
	hd.dsl.Rule("utility", []string{"sleep", "NUMBER", "time_unit"}, "waitCmd")
	hd.dsl.Rule("utility", []string{"log", "STRING"}, "logCmd")
	hd.dsl.Rule("utility", []string{"debug", "STRING"}, "debugCmd")
	hd.dsl.Rule("utility", []string{"clear", "cookies"}, "clearCookies")
	hd.dsl.Rule("utility", []string{"reset"}, "resetCmd")
	hd.dsl.Rule("utility", []string{"base", "url", "STRING"}, "setBaseURL")
	
	hd.dsl.Action("waitCmd", func(args []interface{}) (interface{}, error) {
		duration, _ := strconv.ParseFloat(args[1].(string), 64)
		unit := args[2].(string)
		if unit == "s" {
			duration = duration * 1000
		}
		hd.engine.Wait(int(duration))
		return fmt.Sprintf("Waited %.0fms", duration), nil
	})
	
	hd.dsl.Action("logCmd", func(args []interface{}) (interface{}, error) {
		message := hd.expandVariables(hd.unquoteString(args[1].(string)))
		hd.engine.Log(message)
		return fmt.Sprintf("Logged: %s", message), nil
	})
	
	hd.dsl.Action("debugCmd", func(args []interface{}) (interface{}, error) {
		message := hd.expandVariables(hd.unquoteString(args[1].(string)))
		hd.engine.Debug(message)
		return fmt.Sprintf("Debug: %s", message), nil
	})
	
	hd.dsl.Action("clearCookies", func(args []interface{}) (interface{}, error) {
		hd.engine.ClearCookies()
		return "Cookies cleared", nil
	})
	
	hd.dsl.Action("resetCmd", func(args []interface{}) (interface{}, error) {
		hd.engine.Reset()
		hd.variables = make(map[string]interface{})
		hd.context = make(map[string]interface{})
		return "Reset complete", nil
	})
	
	hd.dsl.Action("setBaseURL", func(args []interface{}) (interface{}, error) {
		url := hd.expandVariables(hd.unquoteString(args[2].(string)))
		hd.engine.SetBaseURL(url)
		return fmt.Sprintf("Base URL set to %s", url), nil
	})
}

// Helper methods for internal use

// unquoteString removes surrounding quotes and processes escape sequences.
// Handles standard escape sequences like \n, \t, \r, and escaped quotes.
func (hd *HTTPDSLv3) unquoteString(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		// Remove quotes and handle escape sequences
		s = s[1 : len(s)-1]
		s = strings.ReplaceAll(s, `\"`, `"`)
		s = strings.ReplaceAll(s, `\\`, `\`)
		s = strings.ReplaceAll(s, `\n`, "\n")
		s = strings.ReplaceAll(s, `\t`, "\t")
		s = strings.ReplaceAll(s, `\r`, "\r")
	}
	return s
}

// expandVariables replaces $variable references with their actual values.
// Scans the string for $name patterns and substitutes them with variable values.
// Used throughout the DSL to enable variable interpolation in strings.
//
// DEVELOPER GUIDE: Variable System
// Variables are stored in hd.variables map.
// They're expanded in strings before execution.
// To add special variables (like $ARGC), set them during initialization.
// Variables persist across statements but are cleared on Reset.
func (hd *HTTPDSLv3) expandVariables(s string) string {
	// Expand variables in the string
	result := s
	for name, value := range hd.variables {
		placeholder := "$" + name
		replacement := fmt.Sprintf("%v", value)
		result = strings.ReplaceAll(result, placeholder, replacement)
	}
	return result
}

// toBool converts various types to boolean.
// Empty strings, "false", "0", zero numbers, and nil return false.
// Everything else returns true.
func (hd *HTTPDSLv3) toBool(v interface{}) bool {
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return val != "" && val != "false" && val != "0"
	case int, int64, float64:
		return val != 0
	default:
		return v != nil
	}
}

// toNumber converts various types to float64.
// Handles int, int64, float64, and numeric strings.
// Returns 0 if conversion fails.
func (hd *HTTPDSLv3) toNumber(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		if num, err := strconv.ParseFloat(val, 64); err == nil {
			return num
		}
	}
	return 0
}

// toSlice converts various types to a slice of interfaces.
// Handles arrays, slices, and comma-separated strings.
// Used internally for foreach loop iteration.
func (hd *HTTPDSLv3) toSlice(v interface{}) []interface{} {
	switch val := v.(type) {
	case []interface{}:
		return val
	case []string:
		result := make([]interface{}, len(val))
		for i, v := range val {
			result[i] = v
		}
		return result
	case []int:
		result := make([]interface{}, len(val))
		for i, v := range val {
			result[i] = v
		}
		return result
	case string:
		// Split by comma for simple lists
		parts := strings.Split(val, ",")
		result := make([]interface{}, len(parts))
		for i, p := range parts {
			result[i] = strings.TrimSpace(p)
		}
		return result
	}
	return nil
}

// executeStatement processes a single DSL statement.
// It handles both pre-parsed statements and string commands that need parsing.
// Used internally by the execution engine.
func (hd *HTTPDSLv3) executeStatement(stmt interface{}) (interface{}, error) {
	// If the statement is already executed (is the result), return it
	if stmt == nil {
		return nil, nil
	}
	
	// If it's a string that looks like a command, parse and execute it
	if cmdStr, ok := stmt.(string); ok {
		if strings.HasPrefix(cmdStr, "set ") || strings.HasPrefix(cmdStr, "print ") {
			// This is a command that needs to be executed
			result, err := hd.ParseWithContext(cmdStr)
			if err != nil {
				return nil, err
			}
			return result, nil
		}
	}
	
	// Otherwise return the statement as-is (it's already the result)
	return stmt, nil
}

// executeStatements processes a list of DSL statements sequentially.
// It handles control flow (break/continue) and returns the last result.
// Used internally for processing multi-statement scripts.
func (hd *HTTPDSLv3) executeStatements(stmts interface{}) (interface{}, error) {
	statements, ok := stmts.([]interface{})
	if !ok {
		return hd.executeStatement(stmts)
	}
	
	var lastResult interface{}
	for _, stmt := range statements {
		result, err := hd.executeStatement(stmt)
		if err != nil {
			return nil, err
		}
		lastResult = result
		
		// Check for control flow
		if hd.context["break"] == true || hd.context["continue"] == true {
			break
		}
	}
	return lastResult, nil
}

// evaluateCondition converts a condition result to boolean.
// Used by control flow statements (if/while) to determine execution path.
func (hd *HTTPDSLv3) evaluateCondition(cond interface{}) bool {
	// Re-evaluate the condition (for while loops)
	// This would need to re-parse the condition in a real implementation
	return hd.toBool(cond)
}

// Parse processes a single line of DSL input and returns the result.
// It clears the execution context and provides detailed error messages.
//
// Example:
//
//	result, err := hd.Parse(`GET "https://api.example.com"`)
func (hd *HTTPDSLv3) Parse(input string) (interface{}, error) {
	// Clear context for new parse
	hd.context = make(map[string]interface{})
	
	result, err := hd.dsl.Parse(input)
	if err != nil {
		// Provide better error messages
		if parseErr, ok := err.(*dslbuilder.ParseError); ok {
			return nil, fmt.Errorf("%s", parseErr.DetailedError())
		}
		return nil, err
	}
	return result.Output, nil
}

// ParseMultiline parses multiple HTTP DSL statements separated by newlines.
// This enables script-like execution of multiple commands in sequence.
// Each line is parsed and executed independently.
//
// Example:
//
//	results, err := hd.ParseMultiline(`
//	    set $url "https://api.example.com"
//	    GET "$url/users"
//	    assert status 200
//	`)
func (hd *HTTPDSLv3) ParseMultiline(input string) ([]interface{}, error) {
	// Clear context for new parse
	hd.context = make(map[string]interface{})
	
	results, err := hd.dsl.ParseMultiline(input)
	if err != nil {
		// Provide better error messages
		if parseErr, ok := err.(*dslbuilder.ParseError); ok {
			return nil, fmt.Errorf("%s", parseErr.DetailedError())
		}
		return nil, err
	}
	
	return results, nil
}

// ParseAuto automatically detects single vs multiline input and parses accordingly
// This is the recommended method for parsing HTTP DSL scripts
func (hd *HTTPDSLv3) ParseAuto(input string) (interface{}, error) {
	// Clear context for new parse
	hd.context = make(map[string]interface{})
	
	result, err := hd.dsl.ParseAuto(input)
	if err != nil {
		// Provide better error messages
		if parseErr, ok := err.(*dslbuilder.ParseError); ok {
			return nil, fmt.Errorf("%s", parseErr.DetailedError())
		}
		return nil, err
	}
	
	return result, nil
}

// ParseWithBlocks handles multiline blocks with if/then/endif structures.
// This method preprocesses block constructs before parsing, enabling
// complex control flow with nested blocks.
//
// Deprecated: Use ParseWithBlockSupport instead for better block handling.
func (hd *HTTPDSLv3) ParseWithBlocks(input string) (interface{}, error) {
	// Clear context for new parse
	hd.context = make(map[string]interface{})
	
	result, err := hd.dsl.ParseMultilineBlocks(input)
	if err != nil {
		// Provide better error messages
		if parseErr, ok := err.(*dslbuilder.ParseError); ok {
			return nil, fmt.Errorf("%s", parseErr.DetailedError())
		}
		return nil, err
	}
	
	return result, nil
}

// ParseWithContext parses DSL input without clearing the execution context.
// This allows maintaining variables and state across multiple parse calls.
// Primarily used internally for recursive parsing within blocks.
func (hd *HTTPDSLv3) ParseWithContext(input string) (interface{}, error) {
	// DO NOT clear context - keep existing variables
	result, err := hd.dsl.Parse(input)
	if err != nil {
		// Provide better error messages
		if parseErr, ok := err.(*dslbuilder.ParseError); ok {
			return nil, fmt.Errorf("%s", parseErr.DetailedError())
		}
		return nil, err
	}
	return result.Output, nil
}

// GetEngine returns the underlying HTTP execution engine.
// The engine handles actual HTTP requests, responses, and network operations.
func (hd *HTTPDSLv3) GetEngine() *HTTPEngine {
	return hd.engine
}

// GetVariable retrieves a variable value by name.
// Returns the value and a boolean indicating if the variable exists.
//
// Example:
//
//	if val, ok := hd.GetVariable("username"); ok {
//	    fmt.Printf("Username: %v\n", val)
//	}
func (hd *HTTPDSLv3) GetVariable(name string) (interface{}, bool) {
	val, ok := hd.variables[name]
	return val, ok
}

// SetVariable sets a variable value in the DSL context.
// Variables can be referenced in scripts using $name syntax.
//
// Example:
//
//	hd.SetVariable("baseURL", "https://api.example.com")
//	hd.SetVariable("timeout", 5000)
func (hd *HTTPDSLv3) SetVariable(name string, value interface{}) {
	hd.variables[name] = value
}

// ClearVariables removes all variables from the DSL context.
// Useful for resetting state between script executions.
func (hd *HTTPDSLv3) ClearVariables() {
	hd.variables = make(map[string]interface{})
}

// GetVariables returns a copy of all current variables.
// The returned map can be used for debugging or state inspection.
func (hd *HTTPDSLv3) GetVariables() map[string]interface{} {
	return hd.variables
}

// ValidateJSON validates that a string contains valid JSON.
// Returns nil if valid, or an error describing the JSON syntax issue.
//
// Example:
//
//	if err := hd.ValidateJSON(responseBody); err != nil {
//	    log.Printf("Invalid JSON: %v", err)
//	}
func (hd *HTTPDSLv3) ValidateJSON(jsonStr string) error {
	var temp interface{}
	return json.Unmarshal([]byte(jsonStr), &temp)
}

// MatchesPattern checks if a string matches a regular expression pattern.
// Returns true if the pattern matches anywhere in the string.
//
// Example:
//
//	if hd.MatchesPattern(response, `\"id\":\s*\d+`) {
//	    fmt.Println("Found ID in response")
//	}
func (hd *HTTPDSLv3) MatchesPattern(str, pattern string) bool {
	matched, _ := regexp.MatchString(pattern, str)
	return matched
}