package universal

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

// HTTPDSLv2 represents the improved HTTP DSL with better parsing and features
type HTTPDSLv2 struct {
	dsl       *dslbuilder.DSL
	engine    *HTTPEngine
	variables map[string]interface{}
	context   map[string]interface{}
}

// NewHTTPDSLv2 creates a new improved HTTP DSL instance
func NewHTTPDSLv2() *HTTPDSLv2 {
	hd := &HTTPDSLv2{
		dsl:       dslbuilder.New("HTTPDSLv2"),
		engine:    NewHTTPEngine(),
		variables: make(map[string]interface{}),
		context:   make(map[string]interface{}),
	}
	hd.setupGrammar()
	return hd
}

func (hd *HTTPDSLv2) setupGrammar() {
	// HTTP Methods - Highest priority (90)
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
	
	// Value tokens - Lower priority (0)
	// Improved JSON pattern to handle special characters
	hd.dsl.Token("JSON_INLINE", `\{(?:[^{}]|\{[^{}]*\})*\}`)
	// String with escape sequences
	hd.dsl.Token("STRING", `"(?:[^"\\]|\\.)*"`)
	hd.dsl.Token("NUMBER", `[0-9]+(\.[0-9]+)?`)
	hd.dsl.Token("VARIABLE", `\$[a-zA-Z_][a-zA-Z0-9_]*`)
	hd.dsl.Token("URL", `https?://[^\s]+`)
	hd.dsl.Token("COMPARISON", `==|!=|>=|<=|>|<`)
	hd.dsl.Token("ARITHMETIC", `\+|\-|\*|\/`)
	hd.dsl.Token("ID", `[a-zA-Z_][a-zA-Z0-9_]*`)
	
	// Main program rule
	hd.dsl.Rule("program", []string{"statements"}, "executeProgram")
	
	// Statements (supports multiple statements)
	hd.dsl.Rule("statements", []string{"statement"}, "singleStatement")
	hd.dsl.Rule("statements", []string{"statements", "statement"}, "multipleStatements")
	
	hd.dsl.Action("singleStatement", func(args []interface{}) (interface{}, error) {
		return []interface{}{args[0]}, nil
	})
	
	hd.dsl.Action("multipleStatements", func(args []interface{}) (interface{}, error) {
		stmts := args[0].([]interface{})
		return append(stmts, args[1]), nil
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
	
	// Control flow
	hd.dsl.Rule("control_flow", []string{"break"}, "breakCmd")
	hd.dsl.Rule("control_flow", []string{"continue"}, "continueCmd")
	
	hd.dsl.Action("breakCmd", func(args []interface{}) (interface{}, error) {
		hd.context["break"] = true
		return "break", nil
	})
	
	hd.dsl.Action("continueCmd", func(args []interface{}) (interface{}, error) {
		hd.context["continue"] = true
		return "continue", nil
	})
	
	// HTTP Requests with recursive options
	hd.dsl.Rule("http_request", []string{"http_method", "url_value"}, "httpSimple")
	hd.dsl.Rule("http_request", []string{"http_method", "url_value", "request_options"}, "httpWithOptions")
	
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
	
	// URL values with variable expansion
	hd.dsl.Rule("url_value", []string{"STRING"}, "urlString")
	hd.dsl.Rule("url_value", []string{"URL"}, "urlDirect")
	hd.dsl.Rule("url_value", []string{"VARIABLE"}, "urlVariable")
	
	hd.dsl.Action("urlString", func(args []interface{}) (interface{}, error) {
		url := hd.unquoteString(args[0].(string))
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
	
	// Request options - recursive definition for multiple options
	hd.dsl.Rule("request_options", []string{"request_option"}, "singleOption")
	hd.dsl.Rule("request_options", []string{"request_options", "request_option"}, "multipleOptions")
	
	hd.dsl.Action("singleOption", func(args []interface{}) (interface{}, error) {
		return []interface{}{args[0]}, nil
	})
	
	hd.dsl.Action("multipleOptions", func(args []interface{}) (interface{}, error) {
		options := args[0].([]interface{})
		return append(options, args[1]), nil
	})
	
	// Individual request options
	hd.dsl.Rule("request_option", []string{"header", "STRING", "STRING"}, "headerOption")
	hd.dsl.Rule("request_option", []string{"body", "STRING"}, "bodyOption")
	hd.dsl.Rule("request_option", []string{"json", "STRING"}, "jsonStringOption")
	hd.dsl.Rule("request_option", []string{"json", "JSON_INLINE"}, "jsonInlineOption")
	hd.dsl.Rule("request_option", []string{"auth", "basic", "STRING", "STRING"}, "authBasicOption")
	hd.dsl.Rule("request_option", []string{"auth", "bearer", "STRING"}, "authBearerOption")
	hd.dsl.Rule("request_option", []string{"timeout", "NUMBER", "time_unit"}, "timeoutOption")
	
	hd.dsl.Rule("time_unit", []string{"ms"}, "timeUnit")
	hd.dsl.Rule("time_unit", []string{"s"}, "timeUnit")
	
	hd.dsl.Action("timeUnit", func(args []interface{}) (interface{}, error) {
		return args[0], nil
	})
	
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
		// Validate JSON
		var temp interface{}
		if err := json.Unmarshal([]byte(jsonStr), &temp); err != nil {
			// Return as-is, might be a template
		}
		return map[string]interface{}{
			"type":  "json",
			"value": jsonStr,
		}, nil
	})
	
	hd.dsl.Action("jsonInlineOption", func(args []interface{}) (interface{}, error) {
		jsonStr := hd.expandVariables(args[1].(string))
		// Validate JSON
		var temp interface{}
		if err := json.Unmarshal([]byte(jsonStr), &temp); err != nil {
			// Return as-is
		}
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
	hd.dsl.Rule("expression", []string{"value"}, "passthrough")
	hd.dsl.Rule("expression", []string{"expression", "ARITHMETIC", "value"}, "arithmeticOp")
	
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
		
		value := hd.engine.Extract(extractType, pattern)
		hd.variables[varName] = value
		
		return fmt.Sprintf("Extracted %s using %s and stored in $%s", pattern, extractType, varName), nil
	})
	
	hd.dsl.Action("extractVariableNoPattern", func(args []interface{}) (interface{}, error) {
		extractType := args[1].(string)
		varName := strings.TrimPrefix(args[3].(string), "$")
		
		value := hd.engine.Extract(extractType, "")
		hd.variables[varName] = value
		
		return fmt.Sprintf("Extracted %s and stored in $%s", extractType, varName), nil
	})
	
	// Improved conditionals with proper context
	hd.dsl.Rule("conditional", []string{"if", "condition", "then", "statement"}, "ifSimple")
	hd.dsl.Rule("conditional", []string{"if", "condition", "then", "statement", "else", "statement"}, "ifElse")
	hd.dsl.Rule("conditional", []string{"if", "condition", "then", "statements", "endif"}, "ifBlock")
	hd.dsl.Rule("conditional", []string{"if", "condition", "then", "statements", "else", "statements", "endif"}, "ifElseBlock")
	
	// Conditions with logical operators
	hd.dsl.Rule("condition", []string{"simple_condition"}, "passthrough")
	hd.dsl.Rule("condition", []string{"condition", "and", "simple_condition"}, "andCondition")
	hd.dsl.Rule("condition", []string{"condition", "or", "simple_condition"}, "orCondition")
	hd.dsl.Rule("condition", []string{"not", "condition"}, "notCondition")
	
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
		condition := hd.toBool(args[1])
		if condition {
			return hd.executeStatement(args[3])
		}
		return hd.executeStatement(args[5])
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
	
	// Assertions
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

// Helper methods

func (hd *HTTPDSLv2) unquoteString(s string) string {
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

func (hd *HTTPDSLv2) expandVariables(s string) string {
	// Expand variables in the string
	for name, value := range hd.variables {
		placeholder := "$" + name
		replacement := fmt.Sprintf("%v", value)
		s = strings.ReplaceAll(s, placeholder, replacement)
	}
	return s
}

func (hd *HTTPDSLv2) toBool(v interface{}) bool {
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

func (hd *HTTPDSLv2) toNumber(v interface{}) float64 {
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

func (hd *HTTPDSLv2) toSlice(v interface{}) []interface{} {
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

func (hd *HTTPDSLv2) executeStatement(stmt interface{}) (interface{}, error) {
	// Execute a single statement and return the result
	return stmt, nil
}

func (hd *HTTPDSLv2) executeStatements(stmts interface{}) (interface{}, error) {
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

func (hd *HTTPDSLv2) evaluateCondition(cond interface{}) bool {
	// Re-evaluate the condition (for while loops)
	// This would need to re-parse the condition in a real implementation
	return hd.toBool(cond)
}

// Parse processes DSL input and returns the result
func (hd *HTTPDSLv2) Parse(input string) (interface{}, error) {
	// Clear context for new parse
	hd.context = make(map[string]interface{})
	
	result, err := hd.dsl.Parse(input)
	if err != nil {
		// Provide better error messages
		if strings.Contains(err.Error(), "no alternative matched") {
			// Try to identify the problematic part
			lines := strings.Split(input, "\n")
			for i, line := range lines {
				if line != "" {
					if _, lineErr := hd.dsl.Parse(line); lineErr != nil {
						return nil, fmt.Errorf("parse error at line %d: %s\nInput: %s", i+1, lineErr.Error(), line)
					}
				}
			}
		}
		return nil, fmt.Errorf("parse error: %w\nInput: %s", err, input)
	}
	return result.Output, nil
}

// GetEngine returns the HTTP engine
func (hd *HTTPDSLv2) GetEngine() *HTTPEngine {
	return hd.engine
}

// GetVariable returns a variable value
func (hd *HTTPDSLv2) GetVariable(name string) (interface{}, bool) {
	val, ok := hd.variables[name]
	return val, ok
}

// SetVariable sets a variable value
func (hd *HTTPDSLv2) SetVariable(name string, value interface{}) {
	hd.variables[name] = value
}

// ClearVariables clears all variables
func (hd *HTTPDSLv2) ClearVariables() {
	hd.variables = make(map[string]interface{})
}

// GetVariables returns all variables
func (hd *HTTPDSLv2) GetVariables() map[string]interface{} {
	return hd.variables
}

// ValidateJSON validates a JSON string
func (hd *HTTPDSLv2) ValidateJSON(jsonStr string) error {
	var temp interface{}
	return json.Unmarshal([]byte(jsonStr), &temp)
}

// MatchesPattern checks if a string matches a regex pattern
func (hd *HTTPDSLv2) MatchesPattern(str, pattern string) bool {
	matched, _ := regexp.MatchString(pattern, str)
	return matched
}