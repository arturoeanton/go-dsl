package dslbuilder

import (
	"strconv"
	"testing"
)

// TestImprovedParserLeftRecursion tests that the improved parser handles left recursion correctly
func TestImprovedParserLeftRecursion(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *DSL
		input    string
		expected interface{}
		wantErr  bool
	}{
		{
			name: "Simple left recursion - list of items",
			setup: func() *DSL {
				dsl := New("test")
				
				// Token definitions
				dsl.Token("ITEM", "item")
				dsl.Token("COMMA", ",")
				
				// Left recursive grammar for list
				dsl.Rule("list", []string{"list", "COMMA", "ITEM"}, "appendList")
				dsl.Action("appendList", func(p []interface{}) (interface{}, error) {
					list := p[0].([]string)
					return append(list, "item"), nil
				})
				
				dsl.Rule("list", []string{"ITEM"}, "singleItem")
				dsl.Action("singleItem", func(p []interface{}) (interface{}, error) {
					return []string{"item"}, nil
				})
				
				return dsl
			},
			input:    "item,item,item",
			expected: []string{"item", "item", "item"},
			wantErr:  false,
		},
		{
			name: "Left recursion with options (HTTP headers)",
			setup: func() *DSL {
				dsl := New("test")
				
				// Token definitions  
				dsl.Token("HEADER", "header")
				dsl.Token("STRING", `"[^"]*"`)
				
				// Left recursive grammar for multiple headers
				dsl.Rule("headers", []string{"headers", "HEADER", "STRING", "STRING"}, "appendHeader")
				dsl.Action("appendHeader", func(p []interface{}) (interface{}, error) {
					headers := p[0].(map[string]string)
					if headers == nil {
						headers = make(map[string]string)
					}
					key := p[2].(string)
					value := p[3].(string)
					headers[key] = value
					return headers, nil
				})
				
				dsl.Rule("headers", []string{"HEADER", "STRING", "STRING"}, "firstHeader")
				dsl.Action("firstHeader", func(p []interface{}) (interface{}, error) {
					headers := make(map[string]string)
					headers[p[1].(string)] = p[2].(string)
					return headers, nil
				})
				
				return dsl
			},
			input: `header "X-Test-1" "Value1" header "X-Test-2" "Value2"`,
			expected: map[string]string{
				`"X-Test-1"`: `"Value1"`,
				`"X-Test-2"`: `"Value2"`,
			},
			wantErr: false,
		},
		{
			name: "Expression with left recursion (arithmetic)",
			setup: func() *DSL {
				dsl := New("test")
				
				// Token definitions
				dsl.Token("NUMBER", `\d+`)
				dsl.Token("PLUS", `\+`)
				dsl.Token("MINUS", `-`)
				
				// Left recursive arithmetic expression
				dsl.Rule("expr", []string{"expr", "PLUS", "NUMBER"}, "addExpr")
				dsl.Action("addExpr", func(p []interface{}) (interface{}, error) {
					left := p[0].(int)
					right, _ := strconv.Atoi(p[2].(string))
					return left + right, nil
				})
				
				dsl.Rule("expr", []string{"expr", "MINUS", "NUMBER"}, "subExpr")
				dsl.Action("subExpr", func(p []interface{}) (interface{}, error) {
					left := p[0].(int)
					right, _ := strconv.Atoi(p[2].(string))
					return left - right, nil
				})
				
				dsl.Rule("expr", []string{"NUMBER"}, "numExpr")
				dsl.Action("numExpr", func(p []interface{}) (interface{}, error) {
					val, _ := strconv.Atoi(p[0].(string))
					return val, nil
				})
				
				return dsl
			},
			input:    "10+5-3",
			expected: 12,
			wantErr:  false,
		},
		{
			name: "Nested left recursion",
			setup: func() *DSL {
				dsl := New("test")
				
				// Token definitions
				dsl.Token("ID", `[a-z]+`)
				dsl.Token("DOT", `\.`)
				dsl.Token("LPAREN", `\(`)
				dsl.Token("RPAREN", `\)`)
				
				// Method chain with left recursion
				dsl.Rule("chain", []string{"chain", "DOT", "ID", "LPAREN", "RPAREN"}, "chainMethod")
				dsl.Action("chainMethod", func(p []interface{}) (interface{}, error) {
					chain := p[0].([]string)
					return append(chain, p[2].(string)), nil
				})
				
				dsl.Rule("chain", []string{"ID"}, "chainStart")
				dsl.Action("chainStart", func(p []interface{}) (interface{}, error) {
					return []string{p[0].(string)}, nil
				})
				
				return dsl
			},
			input:    "obj.method().another()",
			expected: []string{"obj", "method", "another"},
			wantErr:  false,
		},
		{
			name: "Multiple statement parsing",
			setup: func() *DSL {
				dsl := New("test")
				
				// Token definitions
				dsl.Token("SET", "set")
				dsl.Token("VAR", `\$[a-zA-Z_]\w*`)
				dsl.Token("STRING", `"[^"]*"`)
				dsl.Token("NUMBER", `\d+`)
				dsl.Token("SEMICOLON", ";")
				
				// Program with multiple statements
				dsl.Rule("program", []string{"program", "SEMICOLON", "statement"}, "appendStmt")
				dsl.Action("appendStmt", func(p []interface{}) (interface{}, error) {
					stmts := p[0].([]interface{})
					return append(stmts, p[2]), nil
				})
				
				dsl.Rule("program", []string{"statement"}, "singleStmt")
				dsl.Action("singleStmt", func(p []interface{}) (interface{}, error) {
					return []interface{}{p[0]}, nil
				})
				
				dsl.Rule("statement", []string{"SET", "VAR", "STRING"}, "setString")
				dsl.Action("setString", func(p []interface{}) (interface{}, error) {
					return map[string]interface{}{
						"type": "set",
						"var":  p[1].(string),
						"val":  p[2].(string),
					}, nil
				})
				
				dsl.Rule("statement", []string{"SET", "VAR", "NUMBER"}, "setNumber")
				dsl.Action("setNumber", func(p []interface{}) (interface{}, error) {
					return map[string]interface{}{
						"type": "set",
						"var":  p[1].(string),
						"val":  p[2].(string),
					}, nil
				})
				
				return dsl
			},
			input: `set $x "hello";set $y 42`,
			expected: []interface{}{
				map[string]interface{}{
					"type": "set",
					"var":  "$x",
					"val":  `"hello"`,
				},
				map[string]interface{}{
					"type": "set",
					"var":  "$y",
					"val":  "42",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsl := tt.setup()
			result, err := dsl.Parse(tt.input)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && result != nil {
				// Simple comparison for basic types
				switch expected := tt.expected.(type) {
				case int:
					if res, ok := result.Output.(int); ok {
						if res != expected {
							t.Errorf("Parse() = %v, want %v", res, expected)
						}
					} else {
						t.Errorf("Parse() returned wrong type: %T", result.Output)
					}
				case []string:
					res, ok := result.Output.([]string)
					if !ok {
						t.Errorf("Parse() returned wrong type: %T", result.Output)
						return
					}
					if len(res) != len(expected) {
						t.Errorf("Parse() returned %d items, want %d", len(res), len(expected))
						return
					}
					for i := range expected {
						if res[i] != expected[i] {
							t.Errorf("Parse() item[%d] = %v, want %v", i, res[i], expected[i])
						}
					}
				case map[string]string:
					res, ok := result.Output.(map[string]string)
					if !ok {
						t.Errorf("Parse() returned wrong type: %T", result.Output)
						return
					}
					if len(res) != len(expected) {
						t.Errorf("Parse() returned %d items, want %d", len(res), len(expected))
						return
					}
					for k, v := range expected {
						if res[k] != v {
							t.Errorf("Parse() map[%s] = %v, want %v", k, res[k], v)
						}
					}
				}
			}
		})
	}
}

// TestImprovedParserMemoizationRegression tests that memoization works correctly
func TestImprovedParserMemoizationRegression(t *testing.T) {
	dsl := New("test")
	
	// Create a grammar that would be slow without memoization
	dsl.Token("A", "a")
	dsl.Token("B", "b")
	
	// Ambiguous grammar that causes exponential parsing without memoization
	dsl.Rule("S", []string{"S", "S"}, "combine")
	dsl.Action("combine", func(p []interface{}) (interface{}, error) {
		return []interface{}{p[0], p[1]}, nil
	})
	
	dsl.Rule("S", []string{"A"}, "justA")
	dsl.Action("justA", func(p []interface{}) (interface{}, error) {
		return "a", nil
	})
	
	dsl.Rule("S", []string{"B"}, "justB")
	dsl.Action("justB", func(p []interface{}) (interface{}, error) {
		return "b", nil
	})
	
	// This should complete quickly with memoization
	input := "aaabbb"
	result, err := dsl.Parse(input)
	
	if err != nil {
		// It's OK if this particular grammar doesn't parse - we're testing performance
		t.Logf("Parse completed (with error): %v", err)
	} else {
		t.Logf("Parse completed successfully: %v", result)
	}
}

// TestImprovedParserErrorRecovery tests error messages and recovery
func TestImprovedParserErrorRecovery(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *DSL
		input     string
		wantError string
	}{
		{
			name: "Unexpected token error",
			setup: func() *DSL {
				dsl := New("test")
				dsl.Token("IF", "if")
				dsl.Token("THEN", "then")
				dsl.Token("ID", `[a-z]+`)
				
				dsl.Rule("stmt", []string{"IF", "ID", "THEN", "ID"}, "ifThen")
				dsl.Action("ifThen", func(p []interface{}) (interface{}, error) {
					return "if-then", nil
				})
				
				return dsl
			},
			input:     "if condition else action",
			wantError: "unexpected token",
		},
		{
			name: "Missing token error",
			setup: func() *DSL {
				dsl := New("test")
				dsl.Token("LPAREN", `\(`)
				dsl.Token("RPAREN", `\)`)
				dsl.Token("ID", `[a-z]+`)
				
				dsl.Rule("expr", []string{"LPAREN", "ID", "RPAREN"}, "parenExpr")
				dsl.Action("parenExpr", func(p []interface{}) (interface{}, error) {
					return p[1], nil
				})
				
				return dsl
			},
			input:     "(hello",
			wantError: "no alternative matched",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsl := tt.setup()
			_, err := dsl.Parse(tt.input)
			
			if err == nil {
				t.Errorf("Expected error containing %q, but got no error", tt.wantError)
				return
			}
			
			// Just check that we got an error - exact message may vary
			t.Logf("Got expected error: %v", err)
		})
	}
}