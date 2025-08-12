package universal

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

// HTTPDSLFixed represents the fixed HTTP DSL with better parsing
type HTTPDSLFixed struct {
	dsl       *dslbuilder.DSL
	engine    *HTTPEngine
	variables map[string]interface{}
}

// NewHTTPDSLFixed creates a new fixed HTTP DSL instance
func NewHTTPDSLFixed() *HTTPDSLFixed {
	hd := &HTTPDSLFixed{
		dsl:       dslbuilder.New("HTTPDSLFixed"),
		engine:    NewHTTPEngine(),
		variables: make(map[string]interface{}),
	}
	hd.setupGrammar()
	return hd
}

func (hd *HTTPDSLFixed) setupGrammar() {
	// HTTP Methods - High priority
	hd.dsl.KeywordToken("GET", "GET")
	hd.dsl.KeywordToken("POST", "POST")
	hd.dsl.KeywordToken("PUT", "PUT")
	hd.dsl.KeywordToken("DELETE", "DELETE")
	hd.dsl.KeywordToken("PATCH", "PATCH")
	hd.dsl.KeywordToken("HEAD", "HEAD")
	hd.dsl.KeywordToken("OPTIONS", "OPTIONS")
	hd.dsl.KeywordToken("CONNECT", "CONNECT")
	hd.dsl.KeywordToken("TRACE", "TRACE")
	
	// Keywords
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
	hd.dsl.KeywordToken("header", "header")
	
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
	
	// Loops
	hd.dsl.KeywordToken("repeat", "repeat")
	hd.dsl.KeywordToken("times", "times")
	hd.dsl.KeywordToken("do", "do")
	hd.dsl.KeywordToken("endloop", "endloop")
	hd.dsl.KeywordToken("while", "while")
	hd.dsl.KeywordToken("foreach", "foreach")
	hd.dsl.KeywordToken("in", "in")
	
	// Assertions
	hd.dsl.KeywordToken("assert", "assert")
	hd.dsl.KeywordToken("expect", "expect")
	hd.dsl.KeywordToken("less", "less")
	hd.dsl.KeywordToken("greater", "greater")
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
	
	// Value tokens - Lower priority
	hd.dsl.Token("JSON_STRING", `\{[^}]*\}`)  // Simple JSON pattern
	hd.dsl.Token("STRING", `"[^"]*"`)
	hd.dsl.Token("NUMBER", `[0-9]+(\.[0-9]+)?`)
	hd.dsl.Token("VARIABLE", `\$[a-zA-Z_][a-zA-Z0-9_]*`)
	hd.dsl.Token("URL", `https?://[^\s]+`)
	hd.dsl.Token("COMPARISON", `==|!=|>=|<=|>|<`)
	hd.dsl.Token("ID", `[a-zA-Z_][a-zA-Z0-9_]*`)
	
	// Main rules
	hd.dsl.Rule("program", []string{"statement"}, "executeStatement")
	
	// Statement types
	hd.dsl.Rule("statement", []string{"http_request"}, "passthrough")
	hd.dsl.Rule("statement", []string{"variable_op"}, "passthrough")
	hd.dsl.Rule("statement", []string{"print_cmd"}, "passthrough")
	hd.dsl.Rule("statement", []string{"conditional"}, "passthrough")
	hd.dsl.Rule("statement", []string{"loop_stmt"}, "passthrough")
	hd.dsl.Rule("statement", []string{"assertion"}, "passthrough")
	hd.dsl.Rule("statement", []string{"utility"}, "passthrough")
	
	hd.dsl.Action("executeStatement", func(args []interface{}) (interface{}, error) {
		if len(args) > 0 {
			return args[0], nil
		}
		return nil, nil
	})
	
	hd.dsl.Action("passthrough", func(args []interface{}) (interface{}, error) {
		if len(args) > 0 {
			return args[0], nil
		}
		return nil, nil
	})
	
	// HTTP Requests - simplified
	hd.dsl.Rule("http_request", []string{"http_method", "url_value"}, "httpSimple")
	hd.dsl.Rule("http_request", []string{"http_method", "url_value", "options_list"}, "httpWithOptions")
	
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
	
	// URL values
	hd.dsl.Rule("url_value", []string{"STRING"}, "urlString")
	hd.dsl.Rule("url_value", []string{"URL"}, "urlDirect")
	hd.dsl.Rule("url_value", []string{"VARIABLE"}, "urlVariable")
	
	hd.dsl.Action("urlString", func(args []interface{}) (interface{}, error) {
		return strings.Trim(args[0].(string), "\""), nil
	})
	
	hd.dsl.Action("urlDirect", func(args []interface{}) (interface{}, error) {
		return args[0], nil
	})
	
	hd.dsl.Action("urlVariable", func(args []interface{}) (interface{}, error) {
		varName := strings.TrimPrefix(args[0].(string), "$")
		if val, ok := hd.variables[varName]; ok {
			return fmt.Sprintf("%v", val), nil
		}
		return "", fmt.Errorf("variable %s not found", varName)
	})
	
	// Options list - can have multiple options
	hd.dsl.Rule("options_list", []string{"option"}, "singleOption")
	hd.dsl.Rule("options_list", []string{"options_list", "option"}, "multipleOptions")
	
	hd.dsl.Action("singleOption", func(args []interface{}) (interface{}, error) {
		return map[string]interface{}{
			"options": []interface{}{args[0]},
		}, nil
	})
	
	hd.dsl.Action("multipleOptions", func(args []interface{}) (interface{}, error) {
		opts := args[0].(map[string]interface{})
		optsList := opts["options"].([]interface{})
		return map[string]interface{}{
			"options": append(optsList, args[1]),
		}, nil
	})
	
	// Individual options
	hd.dsl.Rule("option", []string{"header", "STRING", "STRING"}, "headerOption")
	hd.dsl.Rule("option", []string{"body", "STRING"}, "bodyOption")
	hd.dsl.Rule("option", []string{"json", "STRING"}, "jsonOption")
	hd.dsl.Rule("option", []string{"json", "JSON_STRING"}, "jsonDirectOption")
	hd.dsl.Rule("option", []string{"auth", "basic", "STRING", "STRING"}, "authBasicOption")
	hd.dsl.Rule("option", []string{"auth", "bearer", "STRING"}, "authBearerOption")
	hd.dsl.Rule("option", []string{"timeout", "NUMBER", "time_unit"}, "timeoutOption")
	
	hd.dsl.Rule("time_unit", []string{"ms"}, "timeUnit")
	hd.dsl.Rule("time_unit", []string{"s"}, "timeUnit")
	
	hd.dsl.Action("timeUnit", func(args []interface{}) (interface{}, error) {
		return args[0], nil
	})
	
	hd.dsl.Action("headerOption", func(args []interface{}) (interface{}, error) {
		return map[string]interface{}{
			"type": "header",
			"key":  strings.Trim(args[1].(string), "\""),
			"value": strings.Trim(args[2].(string), "\""),
		}, nil
	})
	
	hd.dsl.Action("bodyOption", func(args []interface{}) (interface{}, error) {
		return map[string]interface{}{
			"type": "body",
			"value": strings.Trim(args[1].(string), "\""),
		}, nil
	})
	
	hd.dsl.Action("jsonOption", func(args []interface{}) (interface{}, error) {
		jsonStr := strings.Trim(args[1].(string), "\"")
		// Validate JSON
		var temp interface{}
		if err := json.Unmarshal([]byte(jsonStr), &temp); err != nil {
			// If it fails, use as-is (might be a template)
		}
		return map[string]interface{}{
			"type": "json",
			"value": jsonStr,
		}, nil
	})
	
	hd.dsl.Action("jsonDirectOption", func(args []interface{}) (interface{}, error) {
		return map[string]interface{}{
			"type": "json",
			"value": args[1].(string),
		}, nil
	})
	
	hd.dsl.Action("authBasicOption", func(args []interface{}) (interface{}, error) {
		return map[string]interface{}{
			"type": "auth",
			"authType": "basic",
			"user": strings.Trim(args[2].(string), "\""),
			"pass": strings.Trim(args[3].(string), "\""),
		}, nil
	})
	
	hd.dsl.Action("authBearerOption", func(args []interface{}) (interface{}, error) {
		return map[string]interface{}{
			"type": "auth",
			"authType": "bearer",
			"token": strings.Trim(args[2].(string), "\""),
		}, nil
	})
	
	hd.dsl.Action("timeoutOption", func(args []interface{}) (interface{}, error) {
		value, _ := strconv.ParseFloat(args[1].(string), 64)
		unit := args[2].(string)
		if unit == "s" {
			value = value * 1000
		}
		return map[string]interface{}{
			"type": "timeout",
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
		
		// Process options
		optionsMap := args[2].(map[string]interface{})
		optsList := optionsMap["options"].([]interface{})
		
		requestOptions := make(map[string]interface{})
		headers := make(map[string]string)
		
		for _, opt := range optsList {
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
						"type": "bearer",
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
	
	// Set variable
	hd.dsl.Rule("set_var", []string{"set", "VARIABLE", "value"}, "setVariable")
	hd.dsl.Rule("set_var", []string{"var", "VARIABLE", "value"}, "setVariable")
	
	hd.dsl.Rule("value", []string{"STRING"}, "valueString")
	hd.dsl.Rule("value", []string{"NUMBER"}, "valueNumber")
	hd.dsl.Rule("value", []string{"VARIABLE"}, "valueVariable")
	
	hd.dsl.Action("valueString", func(args []interface{}) (interface{}, error) {
		return strings.Trim(args[0].(string), "\""), nil
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
		return nil, fmt.Errorf("variable %s not found", varName)
	})
	
	hd.dsl.Action("setVariable", func(args []interface{}) (interface{}, error) {
		varName := strings.TrimPrefix(args[1].(string), "$")
		value := args[2]
		hd.variables[varName] = value
		return fmt.Sprintf("Variable $%s set to %v", varName, value), nil
	})
	
	// Print command - separate from print variable
	hd.dsl.Rule("print_cmd", []string{"print", "VARIABLE"}, "printVariable")
	
	hd.dsl.Action("printVariable", func(args []interface{}) (interface{}, error) {
		varName := strings.TrimPrefix(args[1].(string), "$")
		if val, ok := hd.variables[varName]; ok {
			return fmt.Sprintf("$%s = %v", varName, val), nil
		}
		return fmt.Sprintf("Variable $%s not found", varName), nil
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
		pattern := strings.Trim(args[2].(string), "\"")
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
	
	// Conditionals - support both single statement and else
	hd.dsl.Rule("conditional", []string{"if", "condition", "then", "statement"}, "ifSimple")
	hd.dsl.Rule("conditional", []string{"if", "condition", "then", "statement", "else", "statement"}, "ifElseSimple")
	
	hd.dsl.Rule("condition", []string{"value", "COMPARISON", "value"}, "comparison")
	hd.dsl.Rule("condition", []string{"value", "contains", "value"}, "containsCheck")
	hd.dsl.Rule("condition", []string{"value", "empty"}, "emptyCheck")
	hd.dsl.Rule("condition", []string{"value", "exists"}, "existsCheck")
	
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
	
	hd.dsl.Action("ifSimple", func(args []interface{}) (interface{}, error) {
		condition := args[1].(bool)
		if condition {
			// Parse and execute the then statement
			return args[3], nil
		}
		return nil, nil
	})
	
	hd.dsl.Action("ifElseSimple", func(args []interface{}) (interface{}, error) {
		condition := args[1].(bool)
		if condition {
			// Execute then statement
			return args[3], nil
		}
		// Execute else statement
		return args[5], nil
	})
	
	// Loops - simplified
	hd.dsl.Rule("loop_stmt", []string{"repeat", "NUMBER", "times", "do"}, "repeatStart")
	hd.dsl.Rule("loop_stmt", []string{"endloop"}, "loopEnd")
	
	hd.dsl.Action("repeatStart", func(args []interface{}) (interface{}, error) {
		times, _ := strconv.Atoi(args[1].(string))
		// Store loop info in context
		hd.variables["_loop_times"] = times
		hd.variables["_loop_count"] = 0
		return fmt.Sprintf("Starting loop for %d times", times), nil
	})
	
	hd.dsl.Action("loopEnd", func(args []interface{}) (interface{}, error) {
		return "Loop ended", nil
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
		return nil, fmt.Errorf("✗ Expected status %d, got %d", expectedCode, actualCode)
	})
	
	hd.dsl.Action("assertTime", func(args []interface{}) (interface{}, error) {
		maxTime, _ := strconv.ParseFloat(args[2].(string), 64)
		actualTime := hd.engine.GetLastResponseTime()
		if actualTime < maxTime {
			return fmt.Sprintf("✓ Response time %.2fms < %.2fms", actualTime, maxTime), nil
		}
		return nil, fmt.Errorf("✗ Response time %.2fms exceeds %.2fms", actualTime, maxTime)
	})
	
	hd.dsl.Action("assertContains", func(args []interface{}) (interface{}, error) {
		expected := strings.Trim(args[2].(string), "\"")
		response := hd.engine.GetLastResponse()
		if strings.Contains(response, expected) {
			return fmt.Sprintf("✓ Response contains '%s'", expected), nil
		}
		return nil, fmt.Errorf("✗ Response does not contain '%s'", expected)
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
		message := strings.Trim(args[1].(string), "\"")
		hd.engine.Log(message)
		return fmt.Sprintf("Logged: %s", message), nil
	})
	
	hd.dsl.Action("debugCmd", func(args []interface{}) (interface{}, error) {
		message := strings.Trim(args[1].(string), "\"")
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
		return "Reset complete", nil
	})
	
	hd.dsl.Action("setBaseURL", func(args []interface{}) (interface{}, error) {
		url := strings.Trim(args[2].(string), "\"")
		hd.engine.SetBaseURL(url)
		return fmt.Sprintf("Base URL set to %s", url), nil
	})
}

// Parse processes DSL input and returns the result
func (hd *HTTPDSLFixed) Parse(input string) (interface{}, error) {
	result, err := hd.dsl.Parse(input)
	if err != nil {
		return nil, err
	}
	return result.Output, nil
}

// GetEngine returns the HTTP engine
func (hd *HTTPDSLFixed) GetEngine() *HTTPEngine {
	return hd.engine
}

// GetVariable returns a variable value
func (hd *HTTPDSLFixed) GetVariable(name string) (interface{}, bool) {
	val, ok := hd.variables[name]
	return val, ok
}

// SetVariable sets a variable value
func (hd *HTTPDSLFixed) SetVariable(name string, value interface{}) {
	hd.variables[name] = value
}

// ClearVariables clears all variables
func (hd *HTTPDSLFixed) ClearVariables() {
	hd.variables = make(map[string]interface{})
}

// GetVariables returns all variables
func (hd *HTTPDSLFixed) GetVariables() map[string]interface{} {
	return hd.variables
}