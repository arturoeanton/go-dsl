package universal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestHTTPDSLv3MultipleHeaders tests the critical fix for multiple headers
func TestHTTPDSLv3MultipleHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Echo all headers back
		headers := make(map[string]string)
		for key := range r.Header {
			headers[key] = r.Header.Get(key)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(headers)
	}))
	defer server.Close()

	dsl := NewHTTPDSLv3()

	tests := []struct {
		name            string
		input           string
		expectedHeaders []string
	}{
		{
			name:            "Two headers",
			input:           fmt.Sprintf(`GET "%s" header "X-Test-1" "Value1" header "X-Test-2" "Value2"`, server.URL),
			expectedHeaders: []string{"X-Test-1", "X-Test-2"},
		},
		{
			name:            "Three headers",
			input:           fmt.Sprintf(`POST "%s" header "Authorization" "Bearer token" header "Content-Type" "application/json" header "X-Custom" "test"`, server.URL),
			expectedHeaders: []string{"Authorization", "Content-Type", "X-Custom"},
		},
		{
			name:            "Headers with body",
			input:           fmt.Sprintf(`PUT "%s" header "X-API-Key" "secret" header "X-Request-ID" "123" body "test data"`, server.URL),
			expectedHeaders: []string{"X-Api-Key", "X-Request-Id"}, // Note: Go canonicalizes header names
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := dsl.Parse(tt.input)
			if err != nil {
				t.Errorf("Parse() error = %v", err)
				return
			}

			if response, ok := result.(map[string]interface{}); ok {
				if body, ok := response["body"].(string); ok {
					// Check that all expected headers are in the response
					for _, header := range tt.expectedHeaders {
						if !strings.Contains(body, header) {
							t.Errorf("Response doesn't contain expected header: %s", header)
						}
					}
				} else {
					t.Errorf("No response body found")
				}
			} else {
				t.Errorf("Result is not a response map")
			}
		})
	}
}

// TestHTTPDSLv3JSONInline tests JSON with special characters like @
func TestHTTPDSLv3JSONInline(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Echo the received JSON back
		var received interface{}
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"received": received,
		})
	}))
	defer server.Close()

	dsl := NewHTTPDSLv3()

	tests := []struct {
		name         string
		input        string
		expectedJSON string
	}{
		{
			name:         "JSON with @ symbol",
			input:        fmt.Sprintf(`POST "%s" json {"email":"test@example.com","name":"Test User"}`, server.URL),
			expectedJSON: "test@example.com",
		},
		{
			name:         "Complex JSON with nested objects",
			input:        fmt.Sprintf(`POST "%s" json {"user":{"email":"admin@test.com","id":123},"active":true}`, server.URL),
			expectedJSON: "admin@test.com",
		},
		{
			name:         "JSON with special characters",
			input:        fmt.Sprintf(`POST "%s" json {"message":"Hello @user #123!","tags":["@mention","#hashtag"]}`, server.URL),
			expectedJSON: "@mention",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := dsl.Parse(tt.input)
			if err != nil {
				t.Errorf("Parse() error = %v", err)
				return
			}

			if response, ok := result.(map[string]interface{}); ok {
				if body, ok := response["body"].(string); ok {
					if !strings.Contains(body, tt.expectedJSON) {
						t.Errorf("Response doesn't contain expected JSON content: %s", tt.expectedJSON)
					}
				}
			}
		})
	}
}

// TestHTTPDSLv3VariableExpansion tests variable expansion in PRINT and URLs
func TestHTTPDSLv3VariableExpansion(t *testing.T) {
	dsl := NewHTTPDSLv3()

	// Set up test variables
	dsl.SetVariable("base_url", "https://api.example.com")
	dsl.SetVariable("version", "v2")
	dsl.SetVariable("endpoint", "users")
	dsl.SetVariable("test_name", "HTTPDSLv3")

	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{
			name:     "Variable in print statement",
			input:    `print "$test_name version $version"`,
			expected: "HTTPDSLv3 version v2",
		},
		{
			name:     "Variable expansion in print",
			input:    `print "API: $base_url/$version/$endpoint"`,
			expected: "API: https://api.example.com/v2/users",
		},
		{
			name:     "Set with variable reference",
			input:    `set $full_url "$base_url/$version"`,
			expected: "https://api.example.com/v2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := dsl.Parse(tt.input)
			if err != nil {
				t.Errorf("Parse() error = %v", err)
				return
			}

			// Check if it's a print result
			if resultStr, ok := result.(string); ok {
				if !strings.Contains(resultStr, tt.expected.(string)) {
					t.Errorf("Expected %q, got %q", tt.expected, resultStr)
				}
			} else if strings.HasPrefix(tt.input, "set ") {
				// For set commands, check the variable was set correctly
				varName := "full_url"
				if val, ok := dsl.GetVariable(varName); ok {
					if val != tt.expected {
						t.Errorf("Variable %s = %v, expected %v", varName, val, tt.expected)
					}
				}
			}
		})
	}
}

// TestHTTPDSLv3Assertions tests assert as standalone statement
func TestHTTPDSLv3Assertions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success","data":"test"}`))
	}))
	defer server.Close()

	dsl := NewHTTPDSLv3()

	// First make a request
	_, err := dsl.Parse(fmt.Sprintf(`GET "%s/api"`, server.URL))
	if err != nil {
		t.Fatalf("Initial request failed: %v", err)
	}

	tests := []struct {
		name       string
		input      string
		shouldPass bool
	}{
		{
			name:       "Assert status as standalone",
			input:      `assert status 200`,
			shouldPass: true,
		},
		{
			name:       "Assert wrong status",
			input:      `assert status 404`,
			shouldPass: false,
		},
		{
			name:       "Assert response time",
			input:      `assert time less 5000 ms`,
			shouldPass: true,
		},
		{
			name:       "Assert response contains",
			input:      `assert response contains "success"`,
			shouldPass: true,
		},
		{
			name:       "Assert response doesn't contain",
			input:      `assert response contains "error"`,
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := dsl.Parse(tt.input)

			if tt.shouldPass && err != nil {
				t.Errorf("Expected assertion to pass but got error: %v", err)
			}

			if !tt.shouldPass && err == nil {
				t.Errorf("Expected assertion to fail but it passed: %v", result)
			}
		})
	}
}

// TestHTTPDSLv3Loops tests loop constructs in DSL
func TestHTTPDSLv3Loops(t *testing.T) {
	dsl := NewHTTPDSLv3()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Simple repeat loop",
			input: `repeat 3 times do
set $counter 1
endloop`,
			expected: "Repeated 3 times",
		},
		{
			name: "While loop",
			input: `set $count 0
while $count < 3 do
set $count $count + 1
endloop`,
			expected: "Loop completed",
		},
		{
			name: "Foreach loop",
			input: `set $items ["a", "b", "c"]
foreach $item in $items do
print $item
endloop`,
			expected: "Loop completed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := dsl.Parse(tt.input)
			if err != nil {
				t.Errorf("Parse() error = %v", err)
				return
			}

			// Loops should complete without error
			t.Logf("Loop executed: %v", result)
		})
	}
}

// TestHTTPDSLv3Conditionals tests if/then/else constructs
func TestHTTPDSLv3Conditionals(t *testing.T) {
	dsl := NewHTTPDSLv3()

	// Set up test variables
	dsl.SetVariable("status", 200)
	dsl.SetVariable("count", 5)
	dsl.SetVariable("error", "")

	tests := []struct {
		name         string
		input        string
		expectedVar  string
		expectedVal  interface{}
	}{
		{
			name:        "Simple if/then without else",
			input:       `if $status == 200 then set $result "OK"`,
			expectedVar: "result",
			expectedVal: "OK",
		},
		{
			name:        "If/then/else",
			input:       `if $count > 10 then set $size "large" else set $size "small"`,
			expectedVar: "size",
			expectedVal: "small",
		},
		{
			name: "Block if statement",
			input: `if $status == 200 then
set $success "true"
set $message "Request successful"
endif`,
			expectedVar: "success",
			expectedVal: "true",
		},
		{
			name:        "Empty check",
			input:       `if $error empty then set $no_error "true"`,
			expectedVar: "no_error",
			expectedVal: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := dsl.Parse(tt.input)
			if err != nil {
				t.Errorf("Parse() error = %v", err)
				return
			}

			if val, ok := dsl.GetVariable(tt.expectedVar); ok {
				if val != tt.expectedVal {
					t.Errorf("Variable %s = %v, expected %v", tt.expectedVar, val, tt.expectedVal)
				}
			} else {
				t.Errorf("Variable %s not found", tt.expectedVar)
			}
		})
	}
}

// TestHTTPDSLv3ExtractWithoutResponse tests extract error handling
func TestHTTPDSLv3ExtractWithoutResponse(t *testing.T) {
	dsl := NewHTTPDSLv3()

	// Clear any previous response
	dsl.GetEngine().Reset()

	// Try to extract without a response
	_, err := dsl.Parse(`extract jsonpath "$.data" as $value`)

	if err == nil {
		t.Errorf("Expected error when extracting without response, but got none")
	}

	if !strings.Contains(err.Error(), "no response") && !strings.Contains(err.Error(), "No response") {
		t.Errorf("Expected 'no response' error, got: %v", err)
	}
}

// TestHTTPDSLv3ArithmeticExpressions tests arithmetic operations
func TestHTTPDSLv3ArithmeticExpressions(t *testing.T) {
	dsl := NewHTTPDSLv3()

	tests := []struct {
		name        string
		input       string
		varName     string
		expectedVal float64
	}{
		{
			name:        "Simple addition",
			input:       `set $result 10 + 5`,
			varName:     "result",
			expectedVal: 15,
		},
		{
			name:        "Subtraction",
			input:       `set $result 20 - 8`,
			varName:     "result",
			expectedVal: 12,
		},
		{
			name:        "Multiplication",
			input:       `set $result 6 * 7`,
			varName:     "result",
			expectedVal: 42,
		},
		{
			name:        "Division",
			input:       `set $result 100 / 4`,
			varName:     "result",
			expectedVal: 25,
		},
		{
			name: "Variable arithmetic",
			input: `set $a 10
set $b 5
set $sum $a + $b`,
			varName:     "sum",
			expectedVal: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse each line if multiline
			lines := strings.Split(tt.input, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" {
					_, err := dsl.Parse(line)
					if err != nil {
						t.Errorf("Parse() error = %v for line: %s", err, line)
						return
					}
				}
			}

			if val, ok := dsl.GetVariable(tt.varName); ok {
				if numVal, ok := val.(float64); ok {
					if numVal != tt.expectedVal {
						t.Errorf("Variable %s = %v, expected %v", tt.varName, numVal, tt.expectedVal)
					}
				} else {
					t.Errorf("Variable %s is not a number: %T", tt.varName, val)
				}
			} else {
				t.Errorf("Variable %s not found", tt.varName)
			}
		})
	}
}

// TestHTTPDSLv3StringEscaping tests string escaping and special characters
func TestHTTPDSLv3StringEscaping(t *testing.T) {
	dsl := NewHTTPDSLv3()

	tests := []struct {
		name        string
		input       string
		varName     string
		expectedVal string
	}{
		{
			name:        "String with quotes",
			input:       `set $msg "He said \"Hello\""`,
			varName:     "msg",
			expectedVal: `He said "Hello"`,
		},
		{
			name:        "String with newline",
			input:       `set $msg "Line 1\nLine 2"`,
			varName:     "msg",
			expectedVal: "Line 1\nLine 2",
		},
		{
			name:        "String with tabs",
			input:       `set $msg "Col1\tCol2\tCol3"`,
			varName:     "msg",
			expectedVal: "Col1\tCol2\tCol3",
		},
		{
			name:        "String with backslash",
			input:       `set $path "C:\\Users\\Test"`,
			varName:     "path",
			expectedVal: `C:\Users\Test`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := dsl.Parse(tt.input)
			if err != nil {
				t.Errorf("Parse() error = %v", err)
				return
			}

			if val, ok := dsl.GetVariable(tt.varName); ok {
				if strVal, ok := val.(string); ok {
					if strVal != tt.expectedVal {
						t.Errorf("Variable %s = %q, expected %q", tt.varName, strVal, tt.expectedVal)
					}
				} else {
					t.Errorf("Variable %s is not a string: %T", tt.varName, val)
				}
			} else {
				t.Errorf("Variable %s not found", tt.varName)
			}
		})
	}
}

// TestHTTPDSLv3CompleteScenario tests a complete real-world scenario
func TestHTTPDSLv3CompleteScenario(t *testing.T) {
	// Create a mock API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/login":
			// Check credentials
			var creds map[string]string
			json.NewDecoder(r.Body).Decode(&creds)
			if creds["username"] == "admin" && creds["password"] == "secret" {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"token":  "jwt-token-12345",
					"userId": "user-789",
					"expires": 3600,
				})
			} else {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "Invalid credentials",
				})
			}

		case "/api/profile":
			// Check auth token
			if r.Header.Get("Authorization") != "Bearer jwt-token-12345" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":    "user-789",
				"name":  "Admin User",
				"email": "admin@example.com",
				"role":  "admin",
			})

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	dsl := NewHTTPDSLv3()

	// Complete scenario with all features
	scenario := []string{
		// Set base URL
		fmt.Sprintf(`base url "%s"`, server.URL),
		
		// Login with credentials
		`set $username "admin"`,
		`set $password "secret"`,
		fmt.Sprintf(`POST "%s/auth/login" json {"username":"admin","password":"secret"}`, server.URL),
		
		// Assert successful login
		`assert status 200`,
		
		// Extract token
		`extract jsonpath "$.token" as $token`,
		`extract jsonpath "$.userId" as $userId`,
		
		// Use token for authenticated request
		fmt.Sprintf(`GET "%s/api/profile" header "Authorization" "Bearer jwt-token-12345"`, server.URL),
		
		// Assert successful profile fetch
		`assert status 200`,
		
		// Extract profile data
		`extract jsonpath "$.email" as $email`,
		`extract jsonpath "$.role" as $role`,
		
		// Conditional based on role
		`if $role == "admin" then set $access_level "full"`,
		
		// Print results
		`print "User $userId has $access_level access"`,
	}

	// Execute scenario
	for i, cmd := range scenario {
		_, err := dsl.Parse(cmd)
		if err != nil {
			t.Errorf("Step %d failed: %s\nError: %v", i+1, cmd, err)
			return
		}
	}

	// Verify extracted values
	expectedVars := map[string]interface{}{
		"token":        "jwt-token-12345",
		"userId":       "user-789",
		"email":        "admin@example.com",
		"role":         "admin",
		"access_level": "full",
	}

	for name, expected := range expectedVars {
		if val, ok := dsl.GetVariable(name); ok {
			if val != expected {
				t.Errorf("Variable %s = %v, expected %v", name, val, expected)
			}
		} else {
			t.Errorf("Variable %s not found", name)
		}
	}
}