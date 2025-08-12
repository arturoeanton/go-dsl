# HTTP DSL v3 - Complete API Reference

## Table of Contents
- [HTTP Methods](#http-methods)
- [Variables and State](#variables-and-state)
- [Control Flow](#control-flow)
- [Data Extraction](#data-extraction)
- [Assertions](#assertions)
- [Output](#output)
- [Utility Commands](#utility-commands)

---

## HTTP Methods

### GET
Performs an HTTP GET request to the specified URL.

**Syntax:**
```http
GET "<url>" [header "<name>" "<value>"]...
```

**Parameters:**
- `url` (string): Target URL for the request
- `header` (optional): Add custom headers to the request

**Example:**
```http
GET "https://api.example.com/users"
    header "Authorization" "Bearer token123"
    header "Accept" "application/json"
```

---

### POST
Performs an HTTP POST request with optional body data.

**Syntax:**
```http
POST "<url>" [header "<name>" "<value>"]... [json <json-object>] [body "<raw-body>"]
```

**Parameters:**
- `url` (string): Target URL for the request
- `header` (optional): Add custom headers
- `json` (optional): Send JSON body (automatically sets Content-Type)
- `body` (optional): Send raw body content

**Examples:**
```http
# JSON body
POST "https://api.example.com/users" json {"name":"John","age":30}

# Raw body
POST "https://api.example.com/data" 
    header "Content-Type" "text/plain"
    body "Raw data content"
```

---

### PUT
Performs an HTTP PUT request for updating resources.

**Syntax:**
```http
PUT "<url>" [header "<name>" "<value>"]... [json <json-object>] [body "<raw-body>"]
```

**Parameters:**
Same as POST

**Example:**
```http
PUT "https://api.example.com/users/1" json {"name":"Jane","age":25}
```

---

### PATCH
Performs an HTTP PATCH request for partial updates.

**Syntax:**
```http
PATCH "<url>" [header "<name>" "<value>"]... [json <json-object>]
```

**Parameters:**
Same as POST

**Example:**
```http
PATCH "https://api.example.com/users/1" json {"age":26}
```

---

### DELETE
Performs an HTTP DELETE request.

**Syntax:**
```http
DELETE "<url>" [header "<name>" "<value>"]...
```

**Parameters:**
- `url` (string): Target URL for the request
- `header` (optional): Add custom headers

**Example:**
```http
DELETE "https://api.example.com/users/1"
    header "Authorization" "Bearer token123"
```

---

## Variables and State

### set
Assigns a value to a variable.

**Syntax:**
```http
set $<variable> <value>
set $<variable> [<array>]
```

**Parameters:**
- `variable` (identifier): Variable name (must start with $)
- `value` (any): Value to assign (string, number, array)

**Examples:**
```http
set $api_key "sk-123456789"
set $user_id 42
set $endpoints ["users", "posts", "comments"]
```

---

### Variables in URLs and Values
Variables can be interpolated in strings using `$variable` syntax.

**Example:**
```http
set $base_url "https://api.example.com"
set $user_id 123
GET "$base_url/users/$user_id"
```

---

## Control Flow

### if/then/else/endif
Conditional execution based on variable values or comparisons.

**Syntax:**
```http
if <condition> then
    <statements>
[else
    <statements>]
endif
```

**Conditions:**
- Variable equality: `$var == "value"`
- Variable inequality: `$var != "value"`
- Numeric comparisons: `>`, `<`, `>=`, `<=`
- Variable existence: `$var`

**Example:**
```http
if $status_code == 200 then
    print "Request successful"
else
    print "Request failed with status: $status_code"
endif
```

---

### while/endloop
Execute statements repeatedly while condition is true.

**Syntax:**
```http
while <condition> do
    <statements>
endloop
```

**Parameters:**
- `condition`: Expression that evaluates to true/false
- Maximum iterations: 1000 (safety limit)

**Example:**
```http
set $counter 0
while $counter < 5 do
    print "Iteration: $counter"
    GET "https://api.example.com/poll"
    set $counter $counter + 1
    wait 1000 ms
endloop
```

---

### foreach/endloop
Iterate over array elements.

**Syntax:**
```http
foreach $<variable> in <array> do
    <statements>
endloop
```

**Parameters:**
- `variable`: Loop variable to hold current element
- `array`: Array to iterate over (can be inline or variable)

**Examples:**
```http
# Inline array
foreach $method in ["GET", "POST", "PUT"] do
    print "Testing method: $method"
endloop

# Variable array
set $users ["alice", "bob", "charlie"]
foreach $user in $users do
    GET "https://api.example.com/users/$user"
endloop
```

---

### repeat/endloop
Execute statements a fixed number of times.

**Syntax:**
```http
repeat <count> times do
    <statements>
endloop
```

**Parameters:**
- `count`: Number of iterations (can be literal or variable)

**Examples:**
```http
# Literal count
repeat 3 times do
    print "Ping!"
    wait 1000 ms
endloop

# Variable count
set $retries 5
repeat $retries times do
    GET "https://api.example.com/health"
endloop
```

---

## Data Extraction

### extract jsonpath
Extract data from JSON responses using JSONPath expressions.

**Syntax:**
```http
extract jsonpath "<jsonpath-expression>" as $<variable>
```

**Parameters:**
- `jsonpath-expression`: JSONPath query string
- `variable`: Variable to store extracted value

**JSONPath Examples:**
- `$.field` - Root level field
- `$.data.items[0]` - First item in array
- `$.users[*].name` - All names from users array
- `$.store.book[?(@.price < 10)]` - Books with price < 10
- `$..author` - All authors (recursive descent)

**Example:**
```http
GET "https://api.example.com/user/1"
extract jsonpath "$.data.email" as $user_email
extract jsonpath "$.data.roles[0]" as $primary_role
print "User email: $user_email"
```

---

### extract regex
Extract data using regular expressions.

**Syntax:**
```http
extract regex "<pattern>" as $<variable>
```

**Parameters:**
- `pattern`: Regular expression pattern (first capture group is extracted)
- `variable`: Variable to store extracted value

**Example:**
```http
GET "https://api.example.com/version"
extract regex "version:\\s*(\\d+\\.\\d+\\.\\d+)" as $version
print "API Version: $version"
```

---

### extract header
Extract HTTP response header values.

**Syntax:**
```http
extract header "<header-name>" as $<variable>
```

**Parameters:**
- `header-name`: Name of the header (case-insensitive)
- `variable`: Variable to store header value

**Example:**
```http
GET "https://api.example.com/data"
extract header "content-type" as $content_type
extract header "x-rate-limit-remaining" as $rate_limit
```

---

## Assertions

### assert status
Verify HTTP response status code.

**Syntax:**
```http
assert status <code>
```

**Parameters:**
- `code`: Expected HTTP status code (200, 404, etc.)

**Example:**
```http
GET "https://api.example.com/users"
assert status 200
```

---

### assert header
Verify presence and optionally value of response header.

**Syntax:**
```http
assert header "<name>" [equals "<value>"]
```

**Parameters:**
- `name`: Header name to check
- `value` (optional): Expected header value

**Examples:**
```http
assert header "content-type" equals "application/json"
assert header "x-rate-limit-remaining"
```

---

### assert json
Verify JSON response structure and values.

**Syntax:**
```http
assert json "<jsonpath>" equals <value>
```

**Parameters:**
- `jsonpath`: JSONPath expression
- `value`: Expected value

**Example:**
```http
GET "https://api.example.com/status"
assert json "$.status" equals "healthy"
assert json "$.version" equals "3.0.0"
```

---

## Output

### print
Output text to console with variable interpolation.

**Syntax:**
```http
print "<text>"
```

**Parameters:**
- `text`: Text to output (supports $variable interpolation)

**Examples:**
```http
print "Starting API test..."
print "User ID: $user_id"
print "Status: $status_code - Response: $response_body"
```

---

## Utility Commands

### wait
Pause execution for specified duration.

**Syntax:**
```http
wait <duration> ms
```

**Parameters:**
- `duration`: Time to wait in milliseconds

**Example:**
```http
print "Waiting for rate limit..."
wait 1000 ms
print "Continuing..."
```

---

### sleep
Alias for wait command.

**Syntax:**
```http
sleep <duration>
```

**Parameters:**
- `duration`: Time to sleep in seconds

**Example:**
```http
sleep 2
```

---

## Advanced Features

### Multiline Blocks
HTTP DSL v3 supports multiline blocks for better organization:

```http
if $status == "error" then
    print "Error detected"
    set $retry_count $retry_count + 1
    if $retry_count < 3 then
        print "Retrying..."
        wait 1000 ms
    endif
endif
```

### Nested Loops
Loops can be nested for complex iterations:

```http
foreach $endpoint in ["users", "posts"] do
    foreach $method in ["GET", "POST"] do
        print "Testing $method $endpoint"
    endloop
endloop
```

### Complex JSONPath Filters
Advanced JSONPath queries with filters:

```http
# Get all active users with age > 18
extract jsonpath "$.users[?(@.active == true && @.age > 18)]" as $adult_users

# Get the most expensive product
extract jsonpath "$.products[?(@.price == $.products.max(price))]" as $expensive
```

### Dynamic Headers
Build headers dynamically using variables:

```http
set $auth_token "Bearer xyz123"
set $api_version "v3"

GET "https://api.example.com/data"
    header "Authorization" "$auth_token"
    header "API-Version" "$api_version"
    header "X-Request-ID" "req-$timestamp"
```

---

## Error Handling

### Response Status Checking
Check response status before processing:

```http
GET "https://api.example.com/data"
extract header "status" as $status_code

if $status_code != "200" then
    print "Error: Received status $status_code"
    extract jsonpath "$.error.message" as $error_msg
    print "Error message: $error_msg"
else
    extract jsonpath "$.data" as $result
    print "Success: $result"
endif
```

### Retry Logic
Implement retry mechanisms:

```http
set $max_retries 3
set $retry_count 0
set $success 0

while $retry_count < $max_retries do
    GET "https://api.example.com/unstable"
    
    if $status_code == 200 then
        set $success 1
        set $retry_count $max_retries
    else
        set $retry_count $retry_count + 1
        if $retry_count < $max_retries then
            print "Retry $retry_count of $max_retries"
            wait 2000 ms
        endif
    endif
endloop

if $success == 0 then
    print "Failed after $max_retries retries"
endif
```

---

## Best Practices

1. **Variable Naming**: Use descriptive variable names with $ prefix
   ```http
   set $user_email "test@example.com"  # Good
   set $e "test@example.com"           # Avoid
   ```

2. **Error Checking**: Always check status codes for critical operations
   ```http
   POST "https://api.example.com/critical"
   assert status 200
   ```

3. **Rate Limiting**: Add delays between requests to avoid rate limits
   ```http
   foreach $id in $user_ids do
       GET "https://api.example.com/user/$id"
       wait 500 ms
   endloop
   ```

4. **Secure Credentials**: Use variables for sensitive data
   ```http
   set $api_key "sk-secure-key-here"
   GET "https://api.example.com/data"
       header "X-API-Key" "$api_key"
   ```

5. **Clean Output**: Use print statements for clear test progress
   ```http
   print "=== Starting User API Tests ==="
   print "Test 1: Create User"
   # ... test code ...
   print "✅ Test 1 Passed"
   ```

---

## Complete Example

```http
# API Testing Suite Example
print "🚀 Starting API Test Suite"

# Configuration
set $base_url "https://jsonplaceholder.typicode.com"
set $test_user_id 1

# Test 1: Get User
print "\n📝 Test 1: Fetching user data..."
GET "$base_url/users/$test_user_id"
assert status 200
extract jsonpath "$.name" as $user_name
extract jsonpath "$.email" as $user_email
print "✅ User found: $user_name ($user_email)"

# Test 2: Get User Posts
print "\n📝 Test 2: Fetching user posts..."
GET "$base_url/users/$test_user_id/posts"
assert status 200
extract jsonpath "$[0].title" as $first_post_title
print "✅ First post: $first_post_title"

# Test 3: Create New Post
print "\n📝 Test 3: Creating new post..."
POST "$base_url/posts" json {
    "userId": 1,
    "title": "Test Post",
    "body": "This is a test post created with HTTP DSL v3"
}
assert status 201
extract jsonpath "$.id" as $new_post_id
print "✅ Post created with ID: $new_post_id"

# Test 4: Verify Creation
print "\n📝 Test 4: Verifying post creation..."
if $new_post_id > 0 then
    print "✅ Post ID is valid"
else
    print "❌ Invalid post ID"
endif

# Summary
print "\n📊 Test Summary:"
print "- User tested: $user_name"
print "- Posts retrieved: Yes"
print "- New post created: ID $new_post_id"
print "\n✅ All tests completed successfully!"
```

---

## Version History

- **v3.0.0** (Current)
  - Full multiline block support
  - While loops implementation
  - Foreach loops with inline arrays
  - Repeat loops with variable support
  - Enhanced JSONPath with filters
  - Improved error handling

- **v2.0.0**
  - Basic control flow (if/then/else)
  - JSONPath extraction
  - Regex extraction
  - Assertions

- **v1.0.0**
  - Basic HTTP methods
  - Simple variables
  - Print statements