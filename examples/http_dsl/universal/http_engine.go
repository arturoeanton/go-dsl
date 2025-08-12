package universal

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

// LogLevel represents logging verbosity
type LogLevel int

const (
	LogError LogLevel = iota
	LogWarn
	LogInfo
	LogDebug
	LogVerbose
)

// RequestHistory stores request/response pairs
type RequestHistory struct {
	Request     *http.Request
	Response    *http.Response
	RequestBody string
	ResponseBody string
	Duration    time.Duration
	Timestamp   time.Time
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	Multiplier     float64
	RetryOn        []int // Status codes to retry on
}

// HTTPEngine handles HTTP requests and responses
type HTTPEngine struct {
	client           *http.Client
	baseURL          string
	lastResponse     *http.Response
	lastResponseBody string
	lastStatusCode   int
	lastResponseTime float64
	cookies          *cookiejar.Jar
	headers          map[string]string
	debug            bool
	logs             []string
	logLevel         LogLevel
	history          []RequestHistory
	maxHistory       int
	retryPolicy      *RetryPolicy
	proxy            string
	tlsConfig        *tls.Config
	requestHooks     []func(*http.Request) error
	responseHooks    []func(*http.Response) error
	rateLimit        time.Duration
	lastRequestTime  time.Time
	metrics          map[string]interface{}
	metricsLock      sync.RWMutex
	sessions         map[string]*Session
	currentSession   string
	oauth2Config     *OAuth2Config
}

// Session represents a named HTTP session with its own state
type Session struct {
	Name     string
	Cookies  *cookiejar.Jar
	Headers  map[string]string
	Variables map[string]interface{}
	History  []RequestHistory
}

// OAuth2Config holds OAuth 2.0 configuration
type OAuth2Config struct {
	ClientID     string
	ClientSecret string
	TokenURL     string
	AuthURL      string
	RedirectURL  string
	Scopes       []string
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
}

// NewHTTPEngine creates a new HTTP engine instance
func NewHTTPEngine() *HTTPEngine {
	jar, _ := cookiejar.New(nil)
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	
	return &HTTPEngine{
		client: &http.Client{
			Jar:       jar,
			Timeout:   30 * time.Second,
			Transport: transport,
		},
		cookies:        jar,
		headers:        make(map[string]string),
		logs:          make([]string, 0),
		logLevel:      LogInfo,
		history:       make([]RequestHistory, 0),
		maxHistory:    100,
		metrics:       make(map[string]interface{}),
		sessions:      make(map[string]*Session),
		requestHooks:  make([]func(*http.Request) error, 0),
		responseHooks: make([]func(*http.Response) error, 0),
	}
}

// Request performs an HTTP request with the given method, URL, and options
func (he *HTTPEngine) Request(method, urlStr string, options map[string]interface{}) (interface{}, error) {
	// Enforce rate limiting
	he.enforceRateLimit()
	
	// Combine with base URL if it's a relative path
	if he.baseURL != "" && !strings.HasPrefix(urlStr, "http") {
		urlStr = he.baseURL + urlStr
	}

	// Parse the URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		he.LogError("Invalid URL: %s", err)
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Create request body
	var body io.Reader
	var bodyStr string
	if options != nil {
		// Handle body options
		if bs, ok := options["body"].(string); ok {
			bodyStr = bs
			body = strings.NewReader(bs)
		} else if jsonBody, ok := options["json"].(string); ok {
			bodyStr = jsonBody
			body = strings.NewReader(jsonBody)
		} else if formData, ok := options["form"].(map[string]string); ok {
			formValues := url.Values{}
			for key, value := range formData {
				formValues.Add(key, value)
			}
			bodyStr = formValues.Encode()
			body = strings.NewReader(bodyStr)
		}
	}

	// Create the request
	req, err := http.NewRequest(method, parsedURL.String(), body)
	if err != nil {
		he.LogError("Failed to create request: %s", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	req.Header.Set("User-Agent", "HTTPDSL/2.0")

	// Apply global headers
	for key, value := range he.headers {
		req.Header.Set(key, value)
	}

	// Apply request-specific options
	if options != nil {
		// Headers
		if headers, ok := options["header"].(map[string]string); ok {
			for key, value := range headers {
				req.Header.Set(key, value)
			}
		}

		// Content-Type for form data
		if _, hasForm := options["form"]; hasForm {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}

		// Content-Type for JSON
		if _, hasJSON := options["json"]; hasJSON {
			req.Header.Set("Content-Type", "application/json")
		}

		// Authentication
		if auth, ok := options["auth"].(map[string]string); ok {
			if auth["type"] == "basic" {
				req.SetBasicAuth(auth["user"], auth["pass"])
			} else if auth["type"] == "bearer" {
				req.Header.Set("Authorization", "Bearer "+auth["token"])
			}
		}

		// Timeout
		if timeout, ok := options["timeout"].(int); ok {
			he.client.Timeout = time.Duration(timeout) * time.Millisecond
		}
	}

	// Apply request hooks
	for _, hook := range he.requestHooks {
		if err := hook(req); err != nil {
			he.LogError("Request hook failed: %s", err)
			return nil, fmt.Errorf("request hook failed: %w", err)
		}
	}

	// Log the request if debug is enabled
	if he.logLevel >= LogDebug {
		he.logRequest(req)
	}

	// Perform the request
	startTime := time.Now()
	resp, err := he.client.Do(req)
	duration := time.Since(startTime)
	he.lastResponseTime = float64(duration.Milliseconds())

	if err != nil {
		he.LogError("Request failed: %s", err)
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Apply response hooks
	for _, hook := range he.responseHooks {
		if err := hook(resp); err != nil {
			he.LogError("Response hook failed: %s", err)
			return nil, fmt.Errorf("response hook failed: %w", err)
		}
	}

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		he.LogError("Failed to read response: %s", err)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Store response data
	he.lastResponse = resp
	he.lastResponseBody = string(bodyBytes)
	he.lastStatusCode = resp.StatusCode

	// Add to history
	he.addToHistory(req, resp, bodyStr, string(bodyBytes), duration)

	// Record metrics
	he.RecordMetric("last_request_duration_ms", duration.Milliseconds())
	he.RecordMetric("last_status_code", resp.StatusCode)
	he.RecordMetric("last_response_size", len(bodyBytes))

	// Log the response if debug is enabled
	if he.logLevel >= LogDebug {
		he.logResponse(resp, string(bodyBytes))
	}

	he.LogInfo("%s %s - Status: %d, Time: %.2fms, Size: %d bytes", 
		method, urlStr, resp.StatusCode, he.lastResponseTime, len(bodyBytes))

	// Return response data
	return map[string]interface{}{
		"status":  resp.StatusCode,
		"body":    string(bodyBytes),
		"headers": resp.Header,
		"time":    he.lastResponseTime,
		"size":    len(bodyBytes),
	}, nil
}

// Extract extracts data from the last response using the specified method
func (he *HTTPEngine) Extract(extractType, pattern string) interface{} {
	switch extractType {
	case "status":
		return he.lastStatusCode

	case "header":
		if he.lastResponse != nil {
			return he.lastResponse.Header.Get(pattern)
		}

	case "jsonpath":
		return he.extractJSONPath(pattern)

	case "xpath":
		// Simplified XPath-like extraction for demonstration
		return he.extractXPath(pattern)

	case "regex":
		return he.extractRegex(pattern)
	}

	return nil
}

// extractJSONPath extracts data using a simple JSON path
func (he *HTTPEngine) extractJSONPath(path string) interface{} {
	var data interface{}
	if err := json.Unmarshal([]byte(he.lastResponseBody), &data); err != nil {
		return nil
	}

	// Handle array at root with filter (e.g., "$[?(@.userId == 1)].title")
	if strings.HasPrefix(path, "$[?(@.") {
		filterEnd := strings.Index(path, ")]")
		if filterEnd > 6 {
			filterExpr := path[6:filterEnd]
			// Parse filter expression
			var fieldName, operator, compareValue string
			if strings.Contains(filterExpr, " == ") {
				parts := strings.Split(filterExpr, " == ")
				fieldName = parts[0]
				compareValue = strings.Trim(parts[1], "'\"")
				operator = "=="
			} else if strings.Contains(filterExpr, " != ") {
				parts := strings.Split(filterExpr, " != ")
				fieldName = parts[0]
				compareValue = strings.Trim(parts[1], "'\"")
				operator = "!="
			} else if strings.Contains(filterExpr, " > ") {
				parts := strings.Split(filterExpr, " > ")
				fieldName = parts[0]
				compareValue = strings.Trim(parts[1], "'\"")
				operator = ">"
			} else if strings.Contains(filterExpr, " < ") {
				parts := strings.Split(filterExpr, " < ")
				fieldName = parts[0]
				compareValue = strings.Trim(parts[1], "'\"")
				operator = "<"
			}

			// Filter array elements
			if arr, ok := data.([]interface{}); ok {
				var results []interface{}
				for _, item := range arr {
					if obj, ok := item.(map[string]interface{}); ok {
						if fieldValue, exists := obj[fieldName]; exists {
							// Compare values
							match := false
							fieldStr := fmt.Sprintf("%v", fieldValue)
							
							// Try numeric comparison
							fieldNum, fieldErr := strconv.ParseFloat(fieldStr, 64)
							compareNum, compareErr := strconv.ParseFloat(compareValue, 64)
							
							if fieldErr == nil && compareErr == nil {
								switch operator {
								case "==":
									match = fieldNum == compareNum
								case "!=":
									match = fieldNum != compareNum
								case ">":
									match = fieldNum > compareNum
								case "<":
									match = fieldNum < compareNum
								}
							} else {
								// String comparison
								switch operator {
								case "==":
									match = fieldStr == compareValue
								case "!=":
									match = fieldStr != compareValue
								}
							}
							
							if match {
								// Check if there's a field selector after the filter
								if filterEnd+2 < len(path) && path[filterEnd+2] == '.' {
									fieldSelector := path[filterEnd+3:]
									if selectedValue, exists := obj[fieldSelector]; exists {
										results = append(results, selectedValue)
									}
								} else {
									results = append(results, item)
								}
							}
						}
					}
				}
				
				// Return single value if only one result, otherwise return array
				if len(results) == 1 {
					return results[0]
				} else if len(results) > 0 {
					return results
				}
			}
		}
		return nil
	}

	// Handle array at root (e.g., "$[0].id")
	if strings.HasPrefix(path, "$[") {
		indexEnd := strings.Index(path, "]")
		if indexEnd > 2 {
			indexStr := path[2:indexEnd]
			index, err := strconv.Atoi(indexStr)
			if err == nil {
				if arr, ok := data.([]interface{}); ok && index < len(arr) {
					current := arr[index]
					// Check if there's more path after the array index
					if indexEnd+1 < len(path) && path[indexEnd+1] == '.' {
						remainingPath := "$" + path[indexEnd+1:]
						// Recursively extract from the array element
						he.lastResponseBody = mustMarshalJSON(current)
						result := he.extractJSONPath(remainingPath)
						// Restore original response body
						he.lastResponseBody = mustMarshalJSON(data)
						return result
					}
					return current
				}
			}
		}
		return nil
	}

	// Simple JSON path implementation
	parts := strings.Split(strings.TrimPrefix(path, "$."), ".")
	current := data

	for _, part := range parts {
		// Handle array indices
		if strings.Contains(part, "[") && strings.Contains(part, "]") {
			fieldName := part[:strings.Index(part, "[")]
			indexStr := part[strings.Index(part, "[")+1 : strings.Index(part, "]")]
			index, _ := strconv.Atoi(indexStr)

			if m, ok := current.(map[string]interface{}); ok {
				if arr, ok := m[fieldName].([]interface{}); ok && index < len(arr) {
					current = arr[index]
					continue
				}
			}
			return nil
		}

		// Handle object fields
		if m, ok := current.(map[string]interface{}); ok {
			current = m[part]
		} else {
			return nil
		}
	}

	return current
}

// Helper function to marshal JSON (panic-free for internal use)
func mustMarshalJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

// extractXPath extracts data using a simplified XPath-like syntax
func (he *HTTPEngine) extractXPath(path string) interface{} {
	// This is a simplified implementation for demonstration
	// In a real implementation, you'd use a proper HTML/XML parser
	
	// Extract text between tags
	if strings.HasPrefix(path, "//") {
		tagName := strings.TrimPrefix(path, "//")
		if strings.Contains(tagName, "/") {
			tagName = tagName[:strings.Index(tagName, "/")]
		}
		
		pattern := fmt.Sprintf("<%s[^>]*>(.*?)</%s>", tagName, tagName)
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(he.lastResponseBody)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return nil
}

// extractRegex extracts data using a regular expression
func (he *HTTPEngine) extractRegex(pattern string) interface{} {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}

	matches := re.FindStringSubmatch(he.lastResponseBody)
	if len(matches) > 1 {
		return matches[1] // Return first capturing group
	} else if len(matches) == 1 {
		return matches[0] // Return full match
	}

	return nil
}

// Compare performs a comparison operation
func (he *HTTPEngine) Compare(left interface{}, op string, right interface{}) bool {
	// Convert to strings for comparison
	leftStr := fmt.Sprintf("%v", left)
	rightStr := fmt.Sprintf("%v", right)

	// Try numeric comparison first
	leftNum, leftErr := strconv.ParseFloat(leftStr, 64)
	rightNum, rightErr := strconv.ParseFloat(rightStr, 64)

	if leftErr == nil && rightErr == nil {
		// Numeric comparison
		switch op {
		case "==":
			return leftNum == rightNum
		case "!=":
			return leftNum != rightNum
		case ">":
			return leftNum > rightNum
		case ">=":
			return leftNum >= rightNum
		case "<":
			return leftNum < rightNum
		case "<=":
			return leftNum <= rightNum
		}
	}

	// String comparison
	switch op {
	case "==":
		return leftStr == rightStr
	case "!=":
		return leftStr != rightStr
	case ">":
		return leftStr > rightStr
	case ">=":
		return leftStr >= rightStr
	case "<":
		return leftStr < rightStr
	case "<=":
		return leftStr <= rightStr
	}

	return false
}

// Matches checks if a value matches a regular expression pattern
func (he *HTTPEngine) Matches(value, pattern string) bool {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(value)
}

// Wait pauses execution for the specified duration in milliseconds
func (he *HTTPEngine) Wait(milliseconds int) {
	time.Sleep(time.Duration(milliseconds) * time.Millisecond)
}

// Log adds a message to the log
func (he *HTTPEngine) Log(message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] %s", timestamp, message)
	he.logs = append(he.logs, logEntry)
	if he.debug {
		fmt.Println(logEntry)
	}
}

// Debug adds a debug message to the log
func (he *HTTPEngine) Debug(message string) {
	if he.debug {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		debugEntry := fmt.Sprintf("[%s] DEBUG: %s", timestamp, message)
		he.logs = append(he.logs, debugEntry)
		fmt.Println(debugEntry)
	}
}

// ClearCookies clears all cookies
func (he *HTTPEngine) ClearCookies() {
	jar, _ := cookiejar.New(nil)
	he.cookies = jar
	he.client.Jar = jar
}

// Reset resets the engine to its initial state
func (he *HTTPEngine) Reset() {
	he.ClearCookies()
	he.headers = make(map[string]string)
	he.baseURL = ""
	he.lastResponse = nil
	he.lastResponseBody = ""
	he.lastStatusCode = 0
	he.lastResponseTime = 0
	he.logs = make([]string, 0)
	he.client.Timeout = 30 * time.Second
}

// SetBaseURL sets the base URL for relative requests
func (he *HTTPEngine) SetBaseURL(url string) {
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	he.baseURL = url
}

// GetLastStatusCode returns the status code of the last response
func (he *HTTPEngine) GetLastStatusCode() int {
	return he.lastStatusCode
}

// GetLastResponseTime returns the response time of the last request in milliseconds
func (he *HTTPEngine) GetLastResponseTime() float64 {
	return he.lastResponseTime
}

// GetLastResponse returns the body of the last response
func (he *HTTPEngine) GetLastResponse() string {
	return he.lastResponseBody
}

// SetHeader sets a global header for all requests
func (he *HTTPEngine) SetHeader(key, value string) {
	he.headers[key] = value
}

// GetHeader gets a global header value
func (he *HTTPEngine) GetHeader(key string) string {
	return he.headers[key]
}

// SetDebug enables or disables debug mode
func (he *HTTPEngine) SetDebug(enabled bool) {
	he.debug = enabled
}

// GetLogs returns all logged messages
func (he *HTTPEngine) GetLogs() []string {
	return he.logs
}

// logRequest logs request details
func (he *HTTPEngine) logRequest(req *http.Request) {
	he.Debug(fmt.Sprintf("Request: %s %s", req.Method, req.URL.String()))
	for key, values := range req.Header {
		for _, value := range values {
			he.Debug(fmt.Sprintf("  Header: %s: %s", key, value))
		}
	}
	if req.Body != nil {
		bodyBytes, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		if len(bodyBytes) > 0 {
			he.Debug(fmt.Sprintf("  Body: %s", string(bodyBytes)))
		}
	}
}

// logResponse logs response details
func (he *HTTPEngine) logResponse(resp *http.Response, body string) {
	he.Debug(fmt.Sprintf("Response: %d %s", resp.StatusCode, resp.Status))
	for key, values := range resp.Header {
		for _, value := range values {
			he.Debug(fmt.Sprintf("  Header: %s: %s", key, value))
		}
	}
	if len(body) > 500 {
		he.Debug(fmt.Sprintf("  Body: %s... (truncated)", body[:500]))
	} else {
		he.Debug(fmt.Sprintf("  Body: %s", body))
	}
}

// SetTimeout sets the client timeout
func (he *HTTPEngine) SetTimeout(seconds int) {
	he.client.Timeout = time.Duration(seconds) * time.Second
}

// AddCookie adds a cookie to the jar
func (he *HTTPEngine) AddCookie(urlStr, name, value string) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return err
	}
	
	cookie := &http.Cookie{
		Name:  name,
		Value: value,
		Path:  "/",
	}
	
	he.cookies.SetCookies(u, []*http.Cookie{cookie})
	return nil
}

// GetCookies returns all cookies for a URL
func (he *HTTPEngine) GetCookies(urlStr string) ([]*http.Cookie, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	return he.cookies.Cookies(u), nil
}

// SetBasicAuth sets basic authentication credentials
func (he *HTTPEngine) SetBasicAuth(username, password string) {
	auth := username + ":" + password
	encoded := base64.StdEncoding.EncodeToString([]byte(auth))
	he.SetHeader("Authorization", "Basic "+encoded)
}

// SetBearerToken sets a bearer token for authorization
func (he *HTTPEngine) SetBearerToken(token string) {
	he.SetHeader("Authorization", "Bearer "+token)
}

// Advanced Cookie Management

// SetCookie sets a detailed cookie
func (he *HTTPEngine) SetCookie(urlStr string, cookie *http.Cookie) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return err
	}
	he.cookies.SetCookies(u, []*http.Cookie{cookie})
	return nil
}

// DeleteCookie removes a specific cookie
func (he *HTTPEngine) DeleteCookie(urlStr, name string) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return err
	}
	
	cookies := he.cookies.Cookies(u)
	newCookies := make([]*http.Cookie, 0)
	for _, cookie := range cookies {
		if cookie.Name != name {
			newCookies = append(newCookies, cookie)
		}
	}
	
	// Clear and reset cookies
	jar, _ := cookiejar.New(nil)
	jar.SetCookies(u, newCookies)
	he.cookies = jar
	he.client.Jar = jar
	
	return nil
}

// GetCookie retrieves a specific cookie
func (he *HTTPEngine) GetCookie(urlStr, name string) (*http.Cookie, error) {
	cookies, err := he.GetCookies(urlStr)
	if err != nil {
		return nil, err
	}
	
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie, nil
		}
	}
	return nil, fmt.Errorf("cookie %s not found", name)
}

// ExportCookies exports all cookies to JSON
func (he *HTTPEngine) ExportCookies() (string, error) {
	// This would need custom implementation as cookiejar doesn't expose all cookies
	return "{}", nil
}

// ImportCookies imports cookies from JSON
func (he *HTTPEngine) ImportCookies(jsonStr string) error {
	// This would need custom implementation
	return nil
}

// Advanced Logging

// SetLogLevel sets the logging verbosity
func (he *HTTPEngine) SetLogLevel(level LogLevel) {
	he.logLevel = level
}

// LogWithLevel logs a message at a specific level
func (he *HTTPEngine) LogWithLevel(level LogLevel, format string, args ...interface{}) {
	if level <= he.logLevel {
		message := fmt.Sprintf(format, args...)
		timestamp := time.Now().Format("2006-01-02 15:04:05.000")
		levelStr := []string{"ERROR", "WARN", "INFO", "DEBUG", "VERBOSE"}[level]
		logEntry := fmt.Sprintf("[%s] [%s] %s", timestamp, levelStr, message)
		he.logs = append(he.logs, logEntry)
		
		if he.debug || level <= LogWarn {
			fmt.Println(logEntry)
		}
	}
}

// LogError logs an error message
func (he *HTTPEngine) LogError(format string, args ...interface{}) {
	he.LogWithLevel(LogError, format, args...)
}

// LogWarn logs a warning message
func (he *HTTPEngine) LogWarn(format string, args ...interface{}) {
	he.LogWithLevel(LogWarn, format, args...)
}

// LogInfo logs an info message
func (he *HTTPEngine) LogInfo(format string, args ...interface{}) {
	he.LogWithLevel(LogInfo, format, args...)
}

// LogDebug logs a debug message
func (he *HTTPEngine) LogDebug(format string, args ...interface{}) {
	he.LogWithLevel(LogDebug, format, args...)
}

// LogVerbose logs a verbose message
func (he *HTTPEngine) LogVerbose(format string, args ...interface{}) {
	he.LogWithLevel(LogVerbose, format, args...)
}

// SSL/TLS Configuration

// SetTLSConfig sets custom TLS configuration
func (he *HTTPEngine) SetTLSConfig(config *tls.Config) {
	he.tlsConfig = config
	if transport, ok := he.client.Transport.(*http.Transport); ok {
		transport.TLSClientConfig = config
	}
}

// SetInsecureSkipVerify disables SSL certificate verification
func (he *HTTPEngine) SetInsecureSkipVerify(skip bool) {
	if he.tlsConfig == nil {
		he.tlsConfig = &tls.Config{}
	}
	he.tlsConfig.InsecureSkipVerify = skip
	
	if transport, ok := he.client.Transport.(*http.Transport); ok {
		transport.TLSClientConfig = he.tlsConfig
	}
}

// SetClientCertificate sets client certificate for mutual TLS
func (he *HTTPEngine) SetClientCertificate(certFile, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	
	if he.tlsConfig == nil {
		he.tlsConfig = &tls.Config{}
	}
	he.tlsConfig.Certificates = []tls.Certificate{cert}
	
	if transport, ok := he.client.Transport.(*http.Transport); ok {
		transport.TLSClientConfig = he.tlsConfig
	}
	
	return nil
}

// SetCustomCA sets custom CA certificate
func (he *HTTPEngine) SetCustomCA(caFile string) error {
	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		return err
	}
	
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	
	if he.tlsConfig == nil {
		he.tlsConfig = &tls.Config{}
	}
	he.tlsConfig.RootCAs = caCertPool
	
	if transport, ok := he.client.Transport.(*http.Transport); ok {
		transport.TLSClientConfig = he.tlsConfig
	}
	
	return nil
}

// Proxy Support

// SetProxy sets HTTP/HTTPS proxy
func (he *HTTPEngine) SetProxy(proxyURL string) error {
	he.proxy = proxyURL
	
	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return err
	}
	
	if transport, ok := he.client.Transport.(*http.Transport); ok {
		transport.Proxy = http.ProxyURL(parsedURL)
	}
	
	return nil
}

// SetSOCKS5Proxy sets SOCKS5 proxy
func (he *HTTPEngine) SetSOCKS5Proxy(host string, auth *proxy.Auth) error {
	dialer, err := proxy.SOCKS5("tcp", host, auth, proxy.Direct)
	if err != nil {
		return err
	}
	
	if transport, ok := he.client.Transport.(*http.Transport); ok {
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		}
	}
	
	return nil
}

// ClearProxy removes proxy configuration
func (he *HTTPEngine) ClearProxy() {
	he.proxy = ""
	if transport, ok := he.client.Transport.(*http.Transport); ok {
		transport.Proxy = nil
		transport.DialContext = nil
	}
}

// Multipart/Form-Data Support

// RequestWithFile performs a request with file upload
func (he *HTTPEngine) RequestWithFile(method, urlStr string, files map[string]string, fields map[string]string) (interface{}, error) {
	// Create multipart writer
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	
	// Add files
	for fieldName, filePath := range files {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		
		part, err := writer.CreateFormFile(fieldName, filepath.Base(filePath))
		if err != nil {
			return nil, err
		}
		
		_, err = io.Copy(part, file)
		if err != nil {
			return nil, err
		}
	}
	
	// Add fields
	for key, value := range fields {
		writer.WriteField(key, value)
	}
	
	err := writer.Close()
	if err != nil {
		return nil, err
	}
	
	// Create request
	req, err := http.NewRequest(method, urlStr, &buf)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", writer.FormDataContentType())
	
	// Apply headers
	for key, value := range he.headers {
		req.Header.Set(key, value)
	}
	
	// Execute request
	resp, err := he.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	
	return map[string]interface{}{
		"status":  resp.StatusCode,
		"body":    string(body),
		"headers": resp.Header,
	}, nil
}

// Request/Response Interceptors

// AddRequestHook adds a request interceptor
func (he *HTTPEngine) AddRequestHook(hook func(*http.Request) error) {
	he.requestHooks = append(he.requestHooks, hook)
}

// AddResponseHook adds a response interceptor
func (he *HTTPEngine) AddResponseHook(hook func(*http.Response) error) {
	he.responseHooks = append(he.responseHooks, hook)
}

// ClearHooks removes all interceptors
func (he *HTTPEngine) ClearHooks() {
	he.requestHooks = make([]func(*http.Request) error, 0)
	he.responseHooks = make([]func(*http.Response) error, 0)
}

// Retry Policies

// SetRetryPolicy configures retry behavior
func (he *HTTPEngine) SetRetryPolicy(policy *RetryPolicy) {
	he.retryPolicy = policy
}

// RequestWithRetry performs a request with retry logic
func (he *HTTPEngine) RequestWithRetry(method, urlStr string, options map[string]interface{}) (interface{}, error) {
	if he.retryPolicy == nil {
		return he.Request(method, urlStr, options)
	}
	
	var lastErr error
	backoff := he.retryPolicy.InitialBackoff
	
	for attempt := 0; attempt <= he.retryPolicy.MaxRetries; attempt++ {
		if attempt > 0 {
			he.LogInfo("Retry attempt %d/%d after %v", attempt, he.retryPolicy.MaxRetries, backoff)
			time.Sleep(backoff)
			
			// Calculate next backoff
			backoff = time.Duration(float64(backoff) * he.retryPolicy.Multiplier)
			if backoff > he.retryPolicy.MaxBackoff {
				backoff = he.retryPolicy.MaxBackoff
			}
		}
		
		result, err := he.Request(method, urlStr, options)
		if err == nil {
			// Check if status code requires retry
			if response, ok := result.(map[string]interface{}); ok {
				if status, ok := response["status"].(int); ok {
					shouldRetry := false
					for _, retryStatus := range he.retryPolicy.RetryOn {
						if status == retryStatus {
							shouldRetry = true
							break
						}
					}
					if !shouldRetry {
						return result, nil
					}
				}
			}
		}
		
		lastErr = err
	}
	
	return nil, fmt.Errorf("max retries exceeded: %v", lastErr)
}

// Connection Management

// SetMaxIdleConnections sets the maximum number of idle connections
func (he *HTTPEngine) SetMaxIdleConnections(max int) {
	if transport, ok := he.client.Transport.(*http.Transport); ok {
		transport.MaxIdleConns = max
	}
}

// SetMaxConnectionsPerHost sets the maximum connections per host
func (he *HTTPEngine) SetMaxConnectionsPerHost(max int) {
	if transport, ok := he.client.Transport.(*http.Transport); ok {
		transport.MaxIdleConnsPerHost = max
	}
}

// SetKeepAlive enables/disables connection keep-alive
func (he *HTTPEngine) SetKeepAlive(enabled bool) {
	if transport, ok := he.client.Transport.(*http.Transport); ok {
		transport.DisableKeepAlives = !enabled
	}
}

// History Management

// GetHistory returns request/response history
func (he *HTTPEngine) GetHistory() []RequestHistory {
	return he.history
}

// ClearHistory clears the request history
func (he *HTTPEngine) ClearHistory() {
	he.history = make([]RequestHistory, 0)
}

// SetMaxHistory sets the maximum history size
func (he *HTTPEngine) SetMaxHistory(max int) {
	he.maxHistory = max
}

// addToHistory adds a request/response to history
func (he *HTTPEngine) addToHistory(req *http.Request, resp *http.Response, reqBody, respBody string, duration time.Duration) {
	if he.maxHistory <= 0 {
		return
	}
	
	history := RequestHistory{
		Request:      req,
		Response:     resp,
		RequestBody:  reqBody,
		ResponseBody: respBody,
		Duration:     duration,
		Timestamp:    time.Now(),
	}
	
	he.history = append(he.history, history)
	
	// Trim history if needed
	if len(he.history) > he.maxHistory {
		he.history = he.history[len(he.history)-he.maxHistory:]
	}
}

// Session Management

// CreateSession creates a new named session
func (he *HTTPEngine) CreateSession(name string) error {
	if _, exists := he.sessions[name]; exists {
		return fmt.Errorf("session %s already exists", name)
	}
	
	jar, _ := cookiejar.New(nil)
	session := &Session{
		Name:      name,
		Cookies:   jar,
		Headers:   make(map[string]string),
		Variables: make(map[string]interface{}),
		History:   make([]RequestHistory, 0),
	}
	
	he.sessions[name] = session
	return nil
}

// SwitchSession switches to a named session
func (he *HTTPEngine) SwitchSession(name string) error {
	session, exists := he.sessions[name]
	if !exists {
		return fmt.Errorf("session %s not found", name)
	}
	
	// Save current session if exists
	if he.currentSession != "" {
		if current, ok := he.sessions[he.currentSession]; ok {
			current.Cookies = he.cookies
			current.Headers = he.headers
			current.History = he.history
		}
	}
	
	// Load new session
	he.currentSession = name
	he.cookies = session.Cookies
	he.client.Jar = session.Cookies
	he.headers = session.Headers
	he.history = session.History
	
	return nil
}

// DeleteSession removes a session
func (he *HTTPEngine) DeleteSession(name string) error {
	if name == he.currentSession {
		return fmt.Errorf("cannot delete active session")
	}
	
	delete(he.sessions, name)
	return nil
}

// ListSessions returns all session names
func (he *HTTPEngine) ListSessions() []string {
	names := make([]string, 0, len(he.sessions))
	for name := range he.sessions {
		names = append(names, name)
	}
	return names
}

// Rate Limiting

// SetRateLimit sets minimum time between requests
func (he *HTTPEngine) SetRateLimit(duration time.Duration) {
	he.rateLimit = duration
}

// enforceRateLimit waits if necessary to respect rate limit
func (he *HTTPEngine) enforceRateLimit() {
	if he.rateLimit <= 0 {
		return
	}
	
	elapsed := time.Since(he.lastRequestTime)
	if elapsed < he.rateLimit {
		time.Sleep(he.rateLimit - elapsed)
	}
	
	he.lastRequestTime = time.Now()
}

// Metrics and Performance

// GetMetrics returns performance metrics
func (he *HTTPEngine) GetMetrics() map[string]interface{} {
	he.metricsLock.RLock()
	defer he.metricsLock.RUnlock()
	
	metrics := make(map[string]interface{})
	for k, v := range he.metrics {
		metrics[k] = v
	}
	return metrics
}

// RecordMetric records a performance metric
func (he *HTTPEngine) RecordMetric(name string, value interface{}) {
	he.metricsLock.Lock()
	defer he.metricsLock.Unlock()
	
	he.metrics[name] = value
}

// GetAverageResponseTime calculates average response time from history
func (he *HTTPEngine) GetAverageResponseTime() float64 {
	if len(he.history) == 0 {
		return 0
	}
	
	var total time.Duration
	for _, h := range he.history {
		total += h.Duration
	}
	
	return float64(total.Milliseconds()) / float64(len(he.history))
}

// OAuth 2.0 Support

// SetOAuth2Config configures OAuth 2.0
func (he *HTTPEngine) SetOAuth2Config(config *OAuth2Config) {
	he.oauth2Config = config
}

// OAuth2Authorize initiates OAuth 2.0 authorization flow
func (he *HTTPEngine) OAuth2Authorize() string {
	if he.oauth2Config == nil {
		return ""
	}
	
	params := url.Values{}
	params.Set("client_id", he.oauth2Config.ClientID)
	params.Set("redirect_uri", he.oauth2Config.RedirectURL)
	params.Set("response_type", "code")
	params.Set("scope", strings.Join(he.oauth2Config.Scopes, " "))
	
	return he.oauth2Config.AuthURL + "?" + params.Encode()
}

// OAuth2ExchangeCode exchanges authorization code for access token
func (he *HTTPEngine) OAuth2ExchangeCode(code string) error {
	if he.oauth2Config == nil {
		return fmt.Errorf("OAuth2 not configured")
	}
	
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("client_id", he.oauth2Config.ClientID)
	data.Set("client_secret", he.oauth2Config.ClientSecret)
	data.Set("redirect_uri", he.oauth2Config.RedirectURL)
	
	resp, err := http.PostForm(he.oauth2Config.TokenURL, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	
	if token, ok := result["access_token"].(string); ok {
		he.oauth2Config.AccessToken = token
		he.SetBearerToken(token)
	}
	
	if refresh, ok := result["refresh_token"].(string); ok {
		he.oauth2Config.RefreshToken = refresh
	}
	
	if expiresIn, ok := result["expires_in"].(float64); ok {
		he.oauth2Config.Expiry = time.Now().Add(time.Duration(expiresIn) * time.Second)
	}
	
	return nil
}

// OAuth2RefreshToken refreshes the access token
func (he *HTTPEngine) OAuth2RefreshToken() error {
	if he.oauth2Config == nil || he.oauth2Config.RefreshToken == "" {
		return fmt.Errorf("refresh token not available")
	}
	
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", he.oauth2Config.RefreshToken)
	data.Set("client_id", he.oauth2Config.ClientID)
	data.Set("client_secret", he.oauth2Config.ClientSecret)
	
	resp, err := http.PostForm(he.oauth2Config.TokenURL, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	
	if token, ok := result["access_token"].(string); ok {
		he.oauth2Config.AccessToken = token
		he.SetBearerToken(token)
	}
	
	if expiresIn, ok := result["expires_in"].(float64); ok {
		he.oauth2Config.Expiry = time.Now().Add(time.Duration(expiresIn) * time.Second)
	}
	
	return nil
}

// GraphQL Support

// GraphQLQuery performs a GraphQL query
func (he *HTTPEngine) GraphQLQuery(endpoint, query string, variables map[string]interface{}) (interface{}, error) {
	payload := map[string]interface{}{
		"query": query,
	}
	
	if variables != nil {
		payload["variables"] = variables
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	
	return he.Request("POST", endpoint, map[string]interface{}{
		"body": string(jsonData),
		"header": map[string]string{
			"Content-Type": "application/json",
		},
	})
}

// WebSocket Support (simplified)

// WebSocketConnect establishes a WebSocket connection
func (he *HTTPEngine) WebSocketConnect(urlStr string) error {
	// This would require gorilla/websocket or similar
	// Placeholder for WebSocket support
	return fmt.Errorf("WebSocket support not yet implemented")
}

// Streaming Support

// StreamRequest performs a streaming request
func (he *HTTPEngine) StreamRequest(method, urlStr string, callback func([]byte) error) error {
	req, err := http.NewRequest(method, urlStr, nil)
	if err != nil {
		return err
	}
	
	// Apply headers
	for key, value := range he.headers {
		req.Header.Set(key, value)
	}
	
	resp, err := he.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	buffer := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			if err := callback(buffer[:n]); err != nil {
				return err
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	
	return nil
}

// DownloadFile downloads a file to disk
func (he *HTTPEngine) DownloadFile(urlStr, filepath string) error {
	resp, err := http.Get(urlStr)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	_, err = io.Copy(file, resp.Body)
	return err
}

