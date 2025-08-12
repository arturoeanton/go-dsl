package main

import (
	"fmt"
	"log"
	"net/http"
	"encoding/json"
	"time"

	"github.com/arturoeanton/go-dsl/examples/http_dsl/universal"
)

func main() {
	// Start a demo server for testing
	go startDemoServer()
	time.Sleep(100 * time.Millisecond) // Give server time to start

	fmt.Println("HTTP DSL Examples")
	fmt.Println("=================\n")

	// Create DSL instance
	dsl := universal.NewHTTPDSL()
	dsl.GetEngine().SetDebug(false) // Set to true to see debug output

	// Example 1: Basic GET request
	fmt.Println("Example 1: Basic GET Request")
	fmt.Println("-----------------------------")
	runExample(dsl, `GET "http://localhost:8080/api/users"`)

	// Example 2: POST with JSON body
	fmt.Println("\nExample 2: POST with JSON Body")
	fmt.Println("-------------------------------")
	runExample(dsl, `POST "http://localhost:8080/api/users" json "{"name":"Alice","email":"alice@example.com"}"`)

	// Example 3: Using variables
	fmt.Println("\nExample 3: Using Variables")
	fmt.Println("---------------------------")
	runExample(dsl, `set $base_url "http://localhost:8080"`)
	runExample(dsl, `set $user_id 123`)
	runExample(dsl, `GET "$base_url/api/users/$user_id"`)
	runExample(dsl, `print $user_id`)

	// Example 4: Request with headers
	fmt.Println("\nExample 4: Request with Headers")
	fmt.Println("--------------------------------")
	runExample(dsl, `GET "http://localhost:8080/api/protected" header "Authorization" "Bearer token123"`)

	// Example 5: Authentication
	fmt.Println("\nExample 5: Authentication")
	fmt.Println("--------------------------")
	runExample(dsl, `POST "http://localhost:8080/api/login" json "{"username":"admin","password":"secret"}"`)
	runExample(dsl, `extract jsonpath "$.token" as $auth_token`)
	runExample(dsl, `print $auth_token`)

	// Example 6: Extracting from responses
	fmt.Println("\nExample 6: Extracting from Responses")
	fmt.Println("-------------------------------------")
	runExample(dsl, `GET "http://localhost:8080/api/users/1"`)
	runExample(dsl, `extract jsonpath "$.name" as $username`)
	runExample(dsl, `extract jsonpath "$.email" as $email`)
	runExample(dsl, `print $username`)
	runExample(dsl, `print $email`)

	// Example 7: Conditional execution
	fmt.Println("\nExample 7: Conditional Execution")
	fmt.Println("---------------------------------")
	runExample(dsl, `GET "http://localhost:8080/api/health"`)
	runExample(dsl, `extract status "" as $status`)
	runExample(dsl, `if $status == 200 then set $health "OK" else set $health "ERROR"`)
	runExample(dsl, `print $health`)

	// Example 8: Loops
	fmt.Println("\nExample 8: Loops")
	fmt.Println("-----------------")
	runExample(dsl, `repeat 3 times do set $counter 1 endloop`)

	// Example 9: Assertions
	fmt.Println("\nExample 9: Assertions")
	fmt.Println("----------------------")
	runExample(dsl, `GET "http://localhost:8080/api/users"`)
	runExample(dsl, `assert status 200`)
	runExample(dsl, `assert response contains "Bob"`)

	// Example 10: Complex workflow
	fmt.Println("\nExample 10: Complex Workflow - User Registration and Profile Update")
	fmt.Println("--------------------------------------------------------------------")
	
	// Reset for clean state
	runExample(dsl, `reset`)
	runExample(dsl, `base url "http://localhost:8080"`)
	
	// Register new user
	runExample(dsl, `POST "/api/register" json "{"username":"newuser","password":"pass123","email":"new@example.com"}"`)
	runExample(dsl, `extract jsonpath "$.userId" as $new_user_id`)
	runExample(dsl, `extract jsonpath "$.token" as $auth_token`)
	
	// Get user profile
	runExample(dsl, `GET "/api/users/$new_user_id" header "Authorization" "Bearer $auth_token"`)
	runExample(dsl, `extract jsonpath "$.email" as $current_email`)
	runExample(dsl, `print $current_email`)
	
	// Update profile
	runExample(dsl, `PATCH "/api/users/$new_user_id" header "Authorization" "Bearer $auth_token" json "{"bio":"Software Developer"}"`)
	runExample(dsl, `assert status 200`)
	
	// Verify update
	runExample(dsl, `GET "/api/users/$new_user_id" header "Authorization" "Bearer $auth_token"`)
	runExample(dsl, `assert response contains "Software Developer"`)

	// Example 11: Error handling
	fmt.Println("\nExample 11: Error Handling")
	fmt.Println("---------------------------")
	runExample(dsl, `GET "http://localhost:8080/api/notfound"`)
	runExample(dsl, `extract status "" as $error_status`)
	runExample(dsl, `if $error_status == 404 then log "Resource not found"`)

	// Example 12: Performance testing with loops
	fmt.Println("\nExample 12: Performance Testing")
	fmt.Println("--------------------------------")
	runExample(dsl, `set $total_time 0`)
	runExample(dsl, `repeat 5 times do GET "http://localhost:8080/api/ping" endloop`)
	
	fmt.Println("\n✅ All examples completed!")
}

func runExample(dsl *universal.HTTPDSL, command string) {
	fmt.Printf("Command: %s\n", command)
	result, err := dsl.Parse(command)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else if result != nil {
		// Format output based on type
		switch v := result.(type) {
		case map[string]interface{}:
			// HTTP response
			if status, ok := v["status"]; ok {
				fmt.Printf("Response: Status=%v", status)
				if body, ok := v["body"]; ok {
					bodyStr := body.(string)
					if len(bodyStr) > 100 {
						fmt.Printf(", Body=%s...\n", bodyStr[:100])
					} else {
						fmt.Printf(", Body=%s\n", bodyStr)
					}
				} else {
					fmt.Println()
				}
			}
		case string:
			if v != "" {
				fmt.Printf("Result: %s\n", v)
			}
		default:
			if v != nil {
				fmt.Printf("Result: %v\n", v)
			}
		}
	}
}

// Demo server for testing
func startDemoServer() {
	// In-memory data store
	users := map[string]map[string]interface{}{
		"1": {
			"id":    "1",
			"name":  "Bob",
			"email": "bob@example.com",
		},
		"123": {
			"id":    "123",
			"name":  "John Doe",
			"email": "john@example.com",
		},
	}
	
	tokens := map[string]string{
		"token123": "user1",
		"auth-token-456": "newuser",
	}

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	// Ping endpoint
	mux.HandleFunc("/api/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"pong": time.Now().Format(time.RFC3339)})
	})

	// Users endpoint
	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(users)
		case "POST":
			var newUser map[string]interface{}
			json.NewDecoder(r.Body).Decode(&newUser)
			userId := fmt.Sprintf("%d", len(users)+1)
			newUser["id"] = userId
			users[userId] = newUser
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(newUser)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// User by ID endpoint
	mux.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
		userId := r.URL.Path[len("/api/users/"):]
		
		switch r.Method {
		case "GET":
			if user, ok := users[userId]; ok {
				// Check authorization for some users
				if userId == "newuser" {
					auth := r.Header.Get("Authorization")
					if auth != "Bearer auth-token-456" {
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					user["bio"] = "Software Developer"
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(user)
			} else {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": "User not found"})
			}
			
		case "PUT", "PATCH":
			auth := r.Header.Get("Authorization")
			if auth == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			
			var updates map[string]interface{}
			json.NewDecoder(r.Body).Decode(&updates)
			
			if user, ok := users[userId]; ok {
				for k, v := range updates {
					user[k] = v
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(user)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
			
		case "DELETE":
			if _, ok := users[userId]; ok {
				delete(users, userId)
				w.WriteHeader(http.StatusNoContent)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
			
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Protected endpoint
	mux.HandleFunc("/api/protected", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer token123" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Access granted"})
	})

	// Login endpoint
	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		var creds map[string]string
		json.NewDecoder(r.Body).Decode(&creds)
		
		if creds["username"] == "admin" && creds["password"] == "secret" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{
				"token": "auth-token-123",
				"userId": "admin",
			})
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid credentials"})
		}
	})

	// Register endpoint
	mux.HandleFunc("/api/register", func(w http.ResponseWriter, r *http.Request) {
		var newUser map[string]interface{}
		json.NewDecoder(r.Body).Decode(&newUser)
		
		userId := "newuser"
		newUser["id"] = userId
		users[userId] = newUser
		
		token := "auth-token-456"
		tokens[token] = userId
		
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"userId": userId,
			"token": token,
		})
	})

	// 404 handler
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
	})

	fmt.Println("Demo server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}