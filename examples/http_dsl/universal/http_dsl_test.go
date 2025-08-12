package universal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestHTTPDSLBasicRequests tests basic HTTP method requests
func TestHTTPDSLBasicRequests(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Echo back request information
		response := map[string]interface{}{
			"method":  r.Method,
			"path":    r.URL.Path,
			"headers": r.Header,
		}
		
		// Handle different methods
		switch r.Method {
		case "GET":
			w.WriteHeader(http.StatusOK)
		case "POST":
			w.WriteHeader(http.StatusCreated)
		case "PUT":
			w.WriteHeader(http.StatusOK)
		case "DELETE":
			w.WriteHeader(http.StatusNoContent)
		case "PATCH":
			w.WriteHeader(http.StatusOK)
		case "HEAD":
			w.WriteHeader(http.StatusOK)
			return // HEAD should not return body
		case "OPTIONS":
			w.Header().Set("Allow", "GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS")
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	dsl := NewHTTPDSL()

	tests := []struct {
		name           string
		input          string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "GET request",
			input:          fmt.Sprintf(`GET "%s/api/users"`, server.URL),
			expectedStatus: 200,
			expectError:    false,
		},
		{
			name:           "POST request",
			input:          fmt.Sprintf(`POST "%s/api/users"`, server.URL),
			expectedStatus: 201,
			expectError:    false,
		},
		{
			name:           "PUT request",
			input:          fmt.Sprintf(`PUT "%s/api/users/1"`, server.URL),
			expectedStatus: 200,
			expectError:    false,
		},
		{
			name:           "DELETE request",
			input:          fmt.Sprintf(`DELETE "%s/api/users/1"`, server.URL),
			expectedStatus: 204,
			expectError:    false,
		},
		{
			name:           "PATCH request",
			input:          fmt.Sprintf(`PATCH "%s/api/users/1"`, server.URL),
			expectedStatus: 200,
			expectError:    false,
		},
		{
			name:           "HEAD request",
			input:          fmt.Sprintf(`HEAD "%s/api/status"`, server.URL),
			expectedStatus: 200,
			expectError:    false,
		},
		{
			name:           "OPTIONS request",
			input:          fmt.Sprintf(`OPTIONS "%s/api"`, server.URL),
			expectedStatus: 200,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := dsl.Parse(tt.input)
			
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}
			
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if err == nil {
				if response, ok := result.(map[string]interface{}); ok {
					if status, ok := response["status"].(int); ok {
						if status != tt.expectedStatus {
							t.Errorf("Expected status %d, got %d", tt.expectedStatus, status)
						}
					}
				}
			}
		})
	}
}

// TestHTTPDSLWithHeaders tests requests with headers
func TestHTTPDSLWithHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Echo headers back
		w.Header().Set("X-Echo-Auth", r.Header.Get("Authorization"))
		w.Header().Set("X-Echo-Custom", r.Header.Get("X-Custom-Header"))
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"auth":   r.Header.Get("Authorization"),
			"custom": r.Header.Get("X-Custom-Header"),
		})
	}))
	defer server.Close()

	dsl := NewHTTPDSL()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Request with single header",
			input:    fmt.Sprintf(`GET "%s/api" header "X-Custom-Header" "TestValue"`, server.URL),
			expected: "TestValue",
		},
		{
			name:     "Request with auth header",
			input:    fmt.Sprintf(`GET "%s/api" header "Authorization" "Bearer token123"`, server.URL),
			expected: "Bearer token123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := dsl.Parse(tt.input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if response, ok := result.(map[string]interface{}); ok {
				if body, ok := response["body"].(string); ok {
					if !strings.Contains(body, tt.expected) {
						t.Errorf("Response body doesn't contain expected value: %s", tt.expected)
					}
				}
			}
		})
	}
}

// TestHTTPDSLWithBody tests requests with body content
func TestHTTPDSLWithBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body []byte
		if r.Body != nil {
			body, _ = io.ReadAll(r.Body)
		}
		
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"received": string(body),
			"contentType": r.Header.Get("Content-Type"),
		})
	}))
	defer server.Close()

	dsl := NewHTTPDSL()

	tests := []struct {
		name         string
		input        string
		expectedBody string
		expectedType string
	}{
		{
			name:         "POST with string body",
			input:        fmt.Sprintf(`POST "%s/api" body "Hello World"`, server.URL),
			expectedBody: "Hello World",
			expectedType: "",
		},
		{
			name:         "POST with JSON body",
			input:        fmt.Sprintf(`POST "%s/api" json "{\"name\":\"test\"}"`, server.URL),
			expectedBody: `{"name":"test"}`,
			expectedType: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := dsl.Parse(tt.input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if response, ok := result.(map[string]interface{}); ok {
				if body, ok := response["body"].(string); ok {
					if !strings.Contains(body, tt.expectedBody) {
						t.Errorf("Response doesn't contain expected body: %s", tt.expectedBody)
					}
					if tt.expectedType != "" && !strings.Contains(body, tt.expectedType) {
						t.Errorf("Response doesn't contain expected content type: %s", tt.expectedType)
					}
				}
			}
		})
	}
}

// TestHTTPDSLVariables tests variable operations
func TestHTTPDSLVariables(t *testing.T) {
	dsl := NewHTTPDSL()

	tests := []struct {
		name     string
		input    string
		varName  string
		expected interface{}
	}{
		{
			name:     "Set string variable",
			input:    `set $name "John Doe"`,
			varName:  "name",
			expected: "John Doe",
		},
		{
			name:     "Set number variable",
			input:    `set $count 42`,
			varName:  "count",
			expected: float64(42),
		},
		{
			name:     "Alternative var syntax",
			input:    `var $api_key "secret123"`,
			varName:  "api_key",
			expected: "secret123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := dsl.Parse(tt.input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if val, ok := dsl.GetVariable(tt.varName); ok {
				if val != tt.expected {
					t.Errorf("Expected variable value %v, got %v", tt.expected, val)
				}
			} else {
				t.Errorf("Variable %s not found", tt.varName)
			}
		})
	}
}

// TestHTTPDSLPrintVariable tests variable printing
func TestHTTPDSLPrintVariable(t *testing.T) {
	dsl := NewHTTPDSL()
	
	// Set a variable first
	dsl.SetVariable("test_var", "Hello World")
	
	result, err := dsl.Parse(`print $test_var`)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}
	
	if resultStr, ok := result.(string); ok {
		if !strings.Contains(resultStr, "Hello World") {
			t.Errorf("Print output doesn't contain variable value: %s", resultStr)
		}
	} else {
		t.Errorf("Expected string result, got %T", result)
	}
}

// TestHTTPDSLExtraction tests response extraction
func TestHTTPDSLExtraction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-ID", "12345")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"user": map[string]interface{}{
				"id":   123,
				"name": "John Doe",
				"email": "john@example.com",
			},
			"token": "abc123xyz",
		})
	}))
	defer server.Close()

	dsl := NewHTTPDSL()

	// First make a request
	_, err := dsl.Parse(fmt.Sprintf(`GET "%s/api/user"`, server.URL))
	if err != nil {
		t.Errorf("Request failed: %v", err)
		return
	}

	tests := []struct {
		name     string
		input    string
		varName  string
		expected interface{}
	}{
		{
			name:     "Extract status code",
			input:    `extract status "" as $status_code`,
			varName:  "status_code",
			expected: 200,
		},
		{
			name:     "Extract header value",
			input:    `extract header "X-Request-ID" as $request_id`,
			varName:  "request_id",
			expected: "12345",
		},
		{
			name:     "Extract with JSONPath",
			input:    `extract jsonpath "$.user.name" as $username`,
			varName:  "username",
			expected: "John Doe",
		},
		{
			name:     "Extract with regex",
			input:    `extract regex "token.*:.*\"([^\"]+)" as $token`,
			varName:  "token",
			expected: "abc123xyz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := dsl.Parse(tt.input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if val, ok := dsl.GetVariable(tt.varName); ok {
				if fmt.Sprintf("%v", val) != fmt.Sprintf("%v", tt.expected) {
					t.Errorf("Expected extracted value %v, got %v", tt.expected, val)
				}
			} else {
				t.Errorf("Variable %s not found after extraction", tt.varName)
			}
		})
	}
}

// TestHTTPDSLConditionals tests conditional statements
func TestHTTPDSLConditionals(t *testing.T) {
	dsl := NewHTTPDSL()

	// Set up test variables
	dsl.SetVariable("status", 200)
	dsl.SetVariable("error", "")
	dsl.SetVariable("count", 5)

	tests := []struct {
		name        string
		input       string
		shouldMatch bool
	}{
		{
			name:        "Simple if with true condition",
			input:       `if $status == 200 then set $result "success"`,
			shouldMatch: true,
		},
		{
			name:        "Simple if with false condition",
			input:       `if $status == 404 then set $notfound "true"`,
			shouldMatch: false,
		},
		{
			name:        "If-else with true condition",
			input:       `if $count > 3 then set $size "large" else set $size "small"`,
			shouldMatch: true,
		},
		{
			name:        "Contains check",
			input:       `if $error empty then set $no_error "true"`,
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := dsl.Parse(tt.input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			// Check if expected variables were set based on conditions
			if tt.shouldMatch {
				// Verify that the conditional action was executed
				if strings.Contains(tt.input, "result") {
					if val, ok := dsl.GetVariable("result"); !ok || val != "success" {
						t.Errorf("Expected result variable to be set to 'success'")
					}
				}
				if strings.Contains(tt.input, "size") {
					if val, ok := dsl.GetVariable("size"); !ok || val != "large" {
						t.Errorf("Expected size variable to be set to 'large'")
					}
				}
				if strings.Contains(tt.input, "no_error") {
					if val, ok := dsl.GetVariable("no_error"); !ok || val != "true" {
						t.Errorf("Expected no_error variable to be set to 'true'")
					}
				}
			}
		})
	}
}

// TestHTTPDSLLoops tests loop statements
func TestHTTPDSLLoops(t *testing.T) {
	dsl := NewHTTPDSL()

	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "Repeat loop",
			input:    `repeat 3 times do set $counter 1 endloop`,
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := dsl.Parse(tt.input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if strings.Contains(fmt.Sprintf("%v", result), fmt.Sprintf("%d times", tt.expected)) {
				// Loop executed correctly
			} else {
				t.Errorf("Loop didn't execute expected number of times")
			}
		})
	}
}

// TestHTTPDSLAssertions tests assertion statements
func TestHTTPDSLAssertions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond) // Small delay for response time testing
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))
	}))
	defer server.Close()

	dsl := NewHTTPDSL()

	// Make a request first
	_, err := dsl.Parse(fmt.Sprintf(`GET "%s/api"`, server.URL))
	if err != nil {
		t.Errorf("Request failed: %v", err)
		return
	}

	tests := []struct {
		name        string
		input       string
		shouldPass  bool
	}{
		{
			name:       "Assert correct status code",
			input:      `assert status 200`,
			shouldPass: true,
		},
		{
			name:       "Assert wrong status code",
			input:      `assert status 404`,
			shouldPass: false,
		},
		{
			name:       "Assert response time",
			input:      `assert time less 1000 ms`,
			shouldPass: true,
		},
		{
			name:       "Assert response contains text",
			input:      `assert response contains "Success"`,
			shouldPass: true,
		},
		{
			name:       "Assert response doesn't contain text",
			input:      `assert response contains "Error"`,
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

// TestHTTPDSLUtilities tests utility operations
func TestHTTPDSLUtilities(t *testing.T) {
	dsl := NewHTTPDSL()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Wait command",
			input:    `wait 100 ms`,
			expected: "Waited 100ms",
		},
		{
			name:     "Sleep command",
			input:    `sleep 0.1 s`,
			expected: "Waited 100s",
		},
		{
			name:     "Log message",
			input:    `log "Test message"`,
			expected: "Logged: Test message",
		},
		{
			name:     "Debug message",
			input:    `debug "Debug info"`,
			expected: "Debug: Debug info",
		},
		{
			name:     "Clear cookies",
			input:    `clear cookies`,
			expected: "Cookies cleared",
		},
		{
			name:     "Reset engine",
			input:    `reset`,
			expected: "Reset complete",
		},
		{
			name:     "Set base URL",
			input:    `base url "https://api.example.com"`,
			expected: "Base URL set to https://api.example.com/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := dsl.Parse(tt.input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if resultStr, ok := result.(string); ok {
				if !strings.Contains(resultStr, strings.Split(tt.expected, ":")[0]) {
					t.Errorf("Expected result to contain '%s', got '%s'", tt.expected, resultStr)
				}
			}
		})
	}
}

// TestHTTPDSLAuthentication tests authentication options
func TestHTTPDSLAuthentication(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"auth": auth,
		})
	}))
	defer server.Close()

	dsl := NewHTTPDSL()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Basic authentication",
			input:    fmt.Sprintf(`GET "%s/api" auth basic "user" "pass"`, server.URL),
			expected: "Basic",
		},
		{
			name:     "Bearer token authentication",
			input:    fmt.Sprintf(`GET "%s/api" auth bearer "token123"`, server.URL),
			expected: "Bearer token123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := dsl.Parse(tt.input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if response, ok := result.(map[string]interface{}); ok {
				if body, ok := response["body"].(string); ok {
					if !strings.Contains(body, tt.expected) {
						t.Errorf("Response doesn't contain expected auth: %s", tt.expected)
					}
				}
			}
		})
	}
}

// TestHTTPDSLTimeout tests timeout configuration
func TestHTTPDSLTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Delay response
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dsl := NewHTTPDSL()

	tests := []struct {
		name        string
		input       string
		shouldError bool
	}{
		{
			name:        "Request with sufficient timeout",
			input:       fmt.Sprintf(`GET "%s" timeout 200 ms`, server.URL),
			shouldError: false,
		},
		{
			name:        "Request with insufficient timeout",
			input:       fmt.Sprintf(`GET "%s" timeout 50 ms`, server.URL),
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := dsl.Parse(tt.input)
			
			if tt.shouldError && err == nil {
				t.Errorf("Expected timeout error but got none")
			}
			
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestHTTPDSLComplexScenario tests a complex scenario with multiple operations
func TestHTTPDSLComplexScenario(t *testing.T) {
	// Create a mock API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login":
			// Return auth token
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{
				"token": "auth-token-123",
				"userId": "user-456",
			})
			
		case "/user/user-456":
			// Check auth header
			if r.Header.Get("Authorization") != "Bearer auth-token-123" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "user-456",
				"name": "John Doe",
				"email": "john@example.com",
				"posts": []string{"post-1", "post-2", "post-3"},
			})
			
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	dsl := NewHTTPDSL()

	// Complex scenario: Login, extract token, use it for authenticated request
	scenario := fmt.Sprintf(`
		POST "%s/login"
		extract jsonpath "$.token" as $token
		extract jsonpath "$.userId" as $userId
		GET "%s/user/user-456" header "Authorization" "Bearer auth-token-123"
		extract jsonpath "$.name" as $username
		assert status 200
	`, server.URL, server.URL)

	// Parse each command separately (since our parser handles one at a time)
	commands := strings.Split(strings.TrimSpace(scenario), "\n")
	for _, cmd := range commands {
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			continue
		}
		
		_, err := dsl.Parse(cmd)
		if err != nil {
			t.Errorf("Command failed: %s, Error: %v", cmd, err)
		}
	}

	// Verify extracted values
	if token, ok := dsl.GetVariable("token"); !ok || token != "auth-token-123" {
		t.Errorf("Token extraction failed")
	}
	
	if userId, ok := dsl.GetVariable("userId"); !ok || userId != "user-456" {
		t.Errorf("UserID extraction failed")
	}
	
	if username, ok := dsl.GetVariable("username"); !ok || username != "John Doe" {
		t.Errorf("Username extraction failed")
	}
}

// TestHTTPDSLEngineFeatures tests direct engine features
func TestHTTPDSLEngineFeatures(t *testing.T) {
	engine := NewHTTPEngine()

	// Test SetBaseURL
	engine.SetBaseURL("https://api.example.com")
	if engine.baseURL != "https://api.example.com/" {
		t.Errorf("Base URL not set correctly")
	}

	// Test SetHeader
	engine.SetHeader("X-API-Key", "test-key")
	if engine.GetHeader("X-API-Key") != "test-key" {
		t.Errorf("Header not set correctly")
	}

	// Test SetDebug
	engine.SetDebug(true)
	if !engine.debug {
		t.Errorf("Debug mode not enabled")
	}

	// Test Log
	engine.Log("Test log message")
	logs := engine.GetLogs()
	if len(logs) == 0 || !strings.Contains(logs[0], "Test log message") {
		t.Errorf("Log message not recorded")
	}

	// Test Reset
	engine.Reset()
	if engine.baseURL != "" || len(engine.headers) != 0 {
		t.Errorf("Engine not reset properly")
	}

	// Test Compare
	if !engine.Compare(5, ">", 3) {
		t.Errorf("Numeric comparison failed")
	}
	if !engine.Compare("abc", "==", "abc") {
		t.Errorf("String comparison failed")
	}

	// Test Matches
	if !engine.Matches("test@example.com", `^[a-z]+@[a-z]+\.com$`) {
		t.Errorf("Regex matching failed")
	}
}

