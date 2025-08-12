package universal

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

// HTTPDSL represents the HTTP DSL with variables, conditionals, and loops
type HTTPDSL struct {
	dsl       *dslbuilder.DSL
	engine    *HTTPEngine
	variables map[string]interface{}
}

// NewHTTPDSL creates a new HTTP DSL instance
func NewHTTPDSL() *HTTPDSL {
	hd := &HTTPDSL{
		dsl:       dslbuilder.New("HTTPDSL"),
		engine:    NewHTTPEngine(),
		variables: make(map[string]interface{}),
	}
	hd.setupGrammar()
	return hd
}

func (hd *HTTPDSL) setupGrammar() {
	// HTTP Methods
	hd.dsl.KeywordToken("GET", "GET")
	hd.dsl.KeywordToken("POST", "POST")
	hd.dsl.KeywordToken("PUT", "PUT")
	hd.dsl.KeywordToken("DELETE", "DELETE")
	hd.dsl.KeywordToken("PATCH", "PATCH")
	hd.dsl.KeywordToken("HEAD", "HEAD")
	hd.dsl.KeywordToken("OPTIONS", "OPTIONS")
	hd.dsl.KeywordToken("CONNECT", "CONNECT")
	hd.dsl.KeywordToken("TRACE", "TRACE")
	
	// Request components
	hd.dsl.KeywordToken("header", "header")
	hd.dsl.KeywordToken("headers", "headers")
	hd.dsl.KeywordToken("body", "body")
	hd.dsl.KeywordToken("json", "json")
	hd.dsl.KeywordToken("form", "form")
	hd.dsl.KeywordToken("query", "query")
	hd.dsl.KeywordToken("timeout", "timeout")
	hd.dsl.KeywordToken("auth", "auth")
	hd.dsl.KeywordToken("basic", "basic")
	hd.dsl.KeywordToken("bearer", "bearer")
	hd.dsl.KeywordToken("cookie", "cookie")
	hd.dsl.KeywordToken("cookies", "cookies")
	hd.dsl.KeywordToken("follow", "follow")
	hd.dsl.KeywordToken("redirects", "redirects")
	hd.dsl.KeywordToken("proxy", "proxy")
	hd.dsl.KeywordToken("retry", "retry")
	hd.dsl.KeywordToken("times", "times")
	hd.dsl.KeywordToken("delay", "delay")
	hd.dsl.KeywordToken("ms", "ms")
	hd.dsl.KeywordToken("s", "s")
	
	// Variable operations
	hd.dsl.KeywordToken("set", "set")
	hd.dsl.KeywordToken("var", "var")
	hd.dsl.KeywordToken("extract", "extract")
	hd.dsl.KeywordToken("from", "from")
	hd.dsl.KeywordToken("response", "response")
	hd.dsl.KeywordToken("status", "status")
	hd.dsl.KeywordToken("jsonpath", "jsonpath")
	hd.dsl.KeywordToken("xpath", "xpath")
	hd.dsl.KeywordToken("regex", "regex")
	hd.dsl.KeywordToken("store", "store")
	hd.dsl.KeywordToken("as", "as")
	hd.dsl.KeywordToken("print", "print")
	
	// Conditional operations
	hd.dsl.KeywordToken("if", "if")
	hd.dsl.KeywordToken("then", "then")
	hd.dsl.KeywordToken("else", "else")
	hd.dsl.KeywordToken("endif", "endif")
	hd.dsl.KeywordToken("equals", "equals")
	hd.dsl.KeywordToken("contains", "contains")
	hd.dsl.KeywordToken("matches", "matches")
	hd.dsl.KeywordToken("greater", "greater")
	hd.dsl.KeywordToken("less", "less")
	hd.dsl.KeywordToken("not", "not")
	hd.dsl.KeywordToken("and", "and")
	hd.dsl.KeywordToken("or", "or")
	hd.dsl.KeywordToken("empty", "empty")
	hd.dsl.KeywordToken("exists", "exists")
	
	// Loop operations
	hd.dsl.KeywordToken("loop", "loop")
	hd.dsl.KeywordToken("repeat", "repeat")
	hd.dsl.KeywordToken("while", "while")
	hd.dsl.KeywordToken("foreach", "foreach")
	hd.dsl.KeywordToken("in", "in")
	hd.dsl.KeywordToken("do", "do")
	hd.dsl.KeywordToken("endloop", "endloop")
	hd.dsl.KeywordToken("break", "break")
	hd.dsl.KeywordToken("continue", "continue")
	hd.dsl.KeywordToken("until", "until")
	
	// Response assertions
	hd.dsl.KeywordToken("assert", "assert")
	hd.dsl.KeywordToken("expect", "expect")
	hd.dsl.KeywordToken("code", "code")
	hd.dsl.KeywordToken("time", "time")
	hd.dsl.KeywordToken("size", "size")
	
	// Utility operations
	hd.dsl.KeywordToken("wait", "wait")
	hd.dsl.KeywordToken("sleep", "sleep")
	hd.dsl.KeywordToken("log", "log")
	hd.dsl.KeywordToken("debug", "debug")
	hd.dsl.KeywordToken("clear", "clear")
	hd.dsl.KeywordToken("reset", "reset")
	hd.dsl.KeywordToken("base", "base")
	hd.dsl.KeywordToken("url", "url")
	
	// Tokens for values
	hd.dsl.Token("STRING", `"[^"]*"`)
	hd.dsl.Token("NUMBER", `[0-9]+(\.[0-9]+)?`)
	hd.dsl.Token("VARIABLE", `\$[a-zA-Z_][a-zA-Z0-9_]*`)
	hd.dsl.Token("URL", `https?://[^\s]+`)
	hd.dsl.Token("ID", `[a-zA-Z_][a-zA-Z0-9_]*`)
	hd.dsl.Token("COMPARISON", `==|!=|>=|<=|>|<`)
	
	// Grammar rules
	hd.dsl.Rule("program", []string{"statement"}, "executeStatement")
	hd.dsl.Rule("program", []string{"block"}, "executeBlock")
	
	hd.dsl.Action("executeStatement", func(args []interface{}) (interface{}, error) {
		if len(args) > 0 {
			return args[0], nil
		}
		return nil, nil
	})
	
	hd.dsl.Action("executeBlock", func(args []interface{}) (interface{}, error) {
		if len(args) > 0 {
			return args[0], nil
		}
		return nil, nil
	})
	
	// Statement types
	hd.dsl.Rule("statement", []string{"http_request"}, "passthrough")
	hd.dsl.Rule("statement", []string{"variable_op"}, "passthrough")
	hd.dsl.Rule("statement", []string{"conditional"}, "passthrough")
	hd.dsl.Rule("statement", []string{"loop_stmt"}, "passthrough")
	hd.dsl.Rule("statement", []string{"assertion"}, "passthrough")
	hd.dsl.Rule("statement", []string{"utility"}, "passthrough")
	
	hd.dsl.Action("passthrough", func(args []interface{}) (interface{}, error) {
		if len(args) > 0 {
			return args[0], nil
		}
		return nil, nil
	})
	
	// HTTP Request rules
	hd.dsl.Rule("http_request", []string{"http_method", "url_value"}, "httpSimple")
	hd.dsl.Rule("http_request", []string{"http_method", "url_value", "request_options"}, "httpWithOptions")
	
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
	
	hd.dsl.Action("httpSimple", func(args []interface{}) (interface{}, error) {
		method := args[0].(string)
		url := args[1].(string)
		return hd.engine.Request(method, url, nil)
	})
	
	hd.dsl.Action("httpWithOptions", func(args []interface{}) (interface{}, error) {
		method := args[0].(string)
		url := args[1].(string)
		options := args[2].(map[string]interface{})
		return hd.engine.Request(method, url, options)
	})
	
	// Request options
	hd.dsl.Rule("request_options", []string{"request_option"}, "singleOption")
	hd.dsl.Rule("request_options", []string{"request_options", "request_option"}, "multipleOptions")
	
	hd.dsl.Action("singleOption", func(args []interface{}) (interface{}, error) {
		return args[0], nil
	})
	
	hd.dsl.Action("multipleOptions", func(args []interface{}) (interface{}, error) {
		options := args[0].(map[string]interface{})
		newOption := args[1].(map[string]interface{})
		for k, v := range newOption {
			options[k] = v
		}
		return options, nil
	})
	
	hd.dsl.Rule("request_option", []string{"header_option"}, "passthrough")
	hd.dsl.Rule("request_option", []string{"body_option"}, "passthrough")
	hd.dsl.Rule("request_option", []string{"auth_option"}, "passthrough")
	hd.dsl.Rule("request_option", []string{"timeout_option"}, "passthrough")
	
	// Header option
	hd.dsl.Rule("header_option", []string{"header", "STRING", "STRING"}, "headerOption")
	hd.dsl.Action("headerOption", func(args []interface{}) (interface{}, error) {
		key := strings.Trim(args[1].(string), "\"")
		value := strings.Trim(args[2].(string), "\"")
		return map[string]interface{}{
			"header": map[string]string{key: value},
		}, nil
	})
	
	// Body options
	hd.dsl.Rule("body_option", []string{"body", "STRING"}, "bodyString")
	hd.dsl.Rule("body_option", []string{"json", "STRING"}, "bodyJSON")
	hd.dsl.Rule("body_option", []string{"form", "form_data"}, "bodyForm")
	
	hd.dsl.Action("bodyString", func(args []interface{}) (interface{}, error) {
		body := strings.Trim(args[1].(string), "\"")
		return map[string]interface{}{"body": body}, nil
	})
	
	hd.dsl.Action("bodyJSON", func(args []interface{}) (interface{}, error) {
		json := strings.Trim(args[1].(string), "\"")
		return map[string]interface{}{"json": json}, nil
	})
	
	hd.dsl.Action("bodyForm", func(args []interface{}) (interface{}, error) {
		return map[string]interface{}{"form": args[1]}, nil
	})
	
	// Form data
	hd.dsl.Rule("form_data", []string{"STRING", "STRING"}, "formPair")
	hd.dsl.Rule("form_data", []string{"form_data", "STRING", "STRING"}, "formMultiple")
	
	hd.dsl.Action("formPair", func(args []interface{}) (interface{}, error) {
		key := strings.Trim(args[0].(string), "\"")
		value := strings.Trim(args[1].(string), "\"")
		return map[string]string{key: value}, nil
	})
	
	hd.dsl.Action("formMultiple", func(args []interface{}) (interface{}, error) {
		form := args[0].(map[string]string)
		key := strings.Trim(args[1].(string), "\"")
		value := strings.Trim(args[2].(string), "\"")
		form[key] = value
		return form, nil
	})
	
	// Auth options
	hd.dsl.Rule("auth_option", []string{"auth", "basic", "STRING", "STRING"}, "authBasic")
	hd.dsl.Rule("auth_option", []string{"auth", "bearer", "STRING"}, "authBearer")
	
	hd.dsl.Action("authBasic", func(args []interface{}) (interface{}, error) {
		user := strings.Trim(args[2].(string), "\"")
		pass := strings.Trim(args[3].(string), "\"")
		return map[string]interface{}{
			"auth": map[string]string{"type": "basic", "user": user, "pass": pass},
		}, nil
	})
	
	hd.dsl.Action("authBearer", func(args []interface{}) (interface{}, error) {
		token := strings.Trim(args[2].(string), "\"")
		return map[string]interface{}{
			"auth": map[string]string{"type": "bearer", "token": token},
		}, nil
	})
	
	// Timeout option
	hd.dsl.Rule("timeout_option", []string{"timeout", "NUMBER", "time_unit"}, "timeoutOption")
	hd.dsl.Rule("time_unit", []string{"ms"}, "timeUnit")
	hd.dsl.Rule("time_unit", []string{"s"}, "timeUnit")
	
	hd.dsl.Action("timeUnit", func(args []interface{}) (interface{}, error) {
		return args[0], nil
	})
	
	hd.dsl.Action("timeoutOption", func(args []interface{}) (interface{}, error) {
		value, _ := strconv.ParseFloat(args[1].(string), 64)
		unit := args[2].(string)
		if unit == "s" {
			value = value * 1000
		}
		return map[string]interface{}{"timeout": int(value)}, nil
	})
	
	// Variable operations
	hd.dsl.Rule("variable_op", []string{"set_var"}, "passthrough")
	hd.dsl.Rule("variable_op", []string{"extract_var"}, "passthrough")
	hd.dsl.Rule("variable_op", []string{"print_var"}, "passthrough")
	
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
	
	// Extract from response
	hd.dsl.Rule("extract_var", []string{"extract", "extraction_type", "STRING", "as", "VARIABLE"}, "extractVariable")
	hd.dsl.Rule("extraction_type", []string{"jsonpath"}, "extractionType")
	hd.dsl.Rule("extraction_type", []string{"xpath"}, "extractionType")
	hd.dsl.Rule("extraction_type", []string{"regex"}, "extractionType")
	hd.dsl.Rule("extraction_type", []string{"header"}, "extractionType")
	hd.dsl.Rule("extraction_type", []string{"status"}, "extractionType")
	
	hd.dsl.Action("extractionType", func(args []interface{}) (interface{}, error) {
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
	
	// Print variable
	hd.dsl.Rule("print_var", []string{"print", "VARIABLE"}, "printVariable")
	hd.dsl.Action("printVariable", func(args []interface{}) (interface{}, error) {
		varName := strings.TrimPrefix(args[1].(string), "$")
		if val, ok := hd.variables[varName]; ok {
			return fmt.Sprintf("$%s = %v", varName, val), nil
		}
		return fmt.Sprintf("Variable $%s not found", varName), nil
	})
	
	// Conditional statements
	hd.dsl.Rule("conditional", []string{"if", "condition", "then", "block"}, "ifStatement")
	hd.dsl.Rule("conditional", []string{"if", "condition", "then", "block", "else", "block"}, "ifElseStatement")
	
	hd.dsl.Rule("condition", []string{"value", "COMPARISON", "value"}, "comparison")
	hd.dsl.Rule("condition", []string{"value", "contains", "value"}, "containsCheck")
	hd.dsl.Rule("condition", []string{"value", "matches", "STRING"}, "matchesCheck")
	hd.dsl.Rule("condition", []string{"value", "exists"}, "existsCheck")
	hd.dsl.Rule("condition", []string{"value", "empty"}, "emptyCheck")
	
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
	
	hd.dsl.Action("matchesCheck", func(args []interface{}) (interface{}, error) {
		value := fmt.Sprintf("%v", args[0])
		pattern := strings.Trim(args[2].(string), "\"")
		return hd.engine.Matches(value, pattern), nil
	})
	
	hd.dsl.Action("existsCheck", func(args []interface{}) (interface{}, error) {
		return args[0] != nil, nil
	})
	
	hd.dsl.Action("emptyCheck", func(args []interface{}) (interface{}, error) {
		val := fmt.Sprintf("%v", args[0])
		return val == "" || val == "0" || val == "false", nil
	})
	
	hd.dsl.Action("ifStatement", func(args []interface{}) (interface{}, error) {
		condition := args[1].(bool)
		if condition {
			return args[3], nil // Execute then block
		}
		return nil, nil
	})
	
	hd.dsl.Action("ifElseStatement", func(args []interface{}) (interface{}, error) {
		condition := args[1].(bool)
		if condition {
			return args[3], nil // Execute then block
		}
		return args[5], nil // Execute else block
	})
	
	// Loop statements
	hd.dsl.Rule("loop_stmt", []string{"repeat_loop"}, "passthrough")
	hd.dsl.Rule("loop_stmt", []string{"while_loop"}, "passthrough")
	hd.dsl.Rule("loop_stmt", []string{"foreach_loop"}, "passthrough")
	
	// Repeat loop
	hd.dsl.Rule("repeat_loop", []string{"repeat", "NUMBER", "times", "do", "block", "endloop"}, "repeatLoop")
	hd.dsl.Action("repeatLoop", func(args []interface{}) (interface{}, error) {
		times, _ := strconv.Atoi(args[1].(string))
		block := args[4]
		
		var results []interface{}
		for i := 0; i < times; i++ {
			hd.variables["_index"] = i
			result := hd.executeBlock(block)
			results = append(results, result)
		}
		
		return fmt.Sprintf("Repeated %d times", times), nil
	})
	
	// While loop
	hd.dsl.Rule("while_loop", []string{"while", "condition", "do", "block", "endloop"}, "whileLoop")
	hd.dsl.Action("whileLoop", func(args []interface{}) (interface{}, error) {
		maxIterations := 1000 // Safety limit
		iterations := 0
		
		for iterations < maxIterations {
			condition := hd.evaluateCondition(args[1])
			if !condition {
				break
			}
			
			hd.executeBlock(args[3])
			iterations++
		}
		
		return fmt.Sprintf("While loop executed %d times", iterations), nil
	})
	
	// Foreach loop
	hd.dsl.Rule("foreach_loop", []string{"foreach", "VARIABLE", "in", "VARIABLE", "do", "block", "endloop"}, "foreachLoop")
	hd.dsl.Action("foreachLoop", func(args []interface{}) (interface{}, error) {
		itemVar := strings.TrimPrefix(args[1].(string), "$")
		listVar := strings.TrimPrefix(args[3].(string), "$")
		
		list, ok := hd.variables[listVar]
		if !ok {
			return nil, fmt.Errorf("list variable $%s not found", listVar)
		}
		
		// Convert to slice if possible
		switch v := list.(type) {
		case []interface{}:
			for _, item := range v {
				hd.variables[itemVar] = item
				hd.executeBlock(args[5])
			}
		case []string:
			for _, item := range v {
				hd.variables[itemVar] = item
				hd.executeBlock(args[5])
			}
		default:
			return nil, fmt.Errorf("variable $%s is not iterable", listVar)
		}
		
		return fmt.Sprintf("Foreach completed for $%s", listVar), nil
	})
	
	// Block of statements
	hd.dsl.Rule("block", []string{"statement"}, "singleStatement")
	hd.dsl.Rule("block", []string{"block", "statement"}, "multipleStatements")
	
	hd.dsl.Action("singleStatement", func(args []interface{}) (interface{}, error) {
		return []interface{}{args[0]}, nil
	})
	
	hd.dsl.Action("multipleStatements", func(args []interface{}) (interface{}, error) {
		block := args[0].([]interface{})
		return append(block, args[1]), nil
	})
	
	// Assertions
	hd.dsl.Rule("assertion", []string{"assert", "assertion_type"}, "assertionCmd")
	hd.dsl.Rule("assertion", []string{"expect", "assertion_type"}, "assertionCmd")
	
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
	
	hd.dsl.Action("assertionCmd", func(args []interface{}) (interface{}, error) {
		return args[1], nil
	})
	
	// Utility operations
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
		return fmt.Sprintf("Waited %.0f%s", duration, unit), nil
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

// Helper methods

func (hd *HTTPDSL) executeBlock(block interface{}) interface{} {
	statements, ok := block.([]interface{})
	if !ok {
		return nil
	}
	
	var lastResult interface{}
	for _, stmt := range statements {
		// Execute each statement
		lastResult = stmt
	}
	return lastResult
}

func (hd *HTTPDSL) evaluateCondition(condition interface{}) bool {
	if cond, ok := condition.(bool); ok {
		return cond
	}
	return false
}

// Parse processes DSL input and returns the result
func (hd *HTTPDSL) Parse(input string) (interface{}, error) {
	result, err := hd.dsl.Parse(input)
	if err != nil {
		return nil, err
	}
	return result.Output, nil
}

// GetEngine returns the HTTP engine
func (hd *HTTPDSL) GetEngine() *HTTPEngine {
	return hd.engine
}

// GetVariable returns a variable value
func (hd *HTTPDSL) GetVariable(name string) (interface{}, bool) {
	val, ok := hd.variables[name]
	return val, ok
}

// SetVariable sets a variable value
func (hd *HTTPDSL) SetVariable(name string, value interface{}) {
	hd.variables[name] = value
}

// ClearVariables clears all variables
func (hd *HTTPDSL) ClearVariables() {
	hd.variables = make(map[string]interface{})
}

// GetVariables returns all variables
func (hd *HTTPDSL) GetVariables() map[string]interface{} {
	return hd.variables
}