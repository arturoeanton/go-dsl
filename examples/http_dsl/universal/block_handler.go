package universal

import (
	"strings"
	"fmt"
)

// Helper function to check if a line starts with an HTTP method
func isHTTPMethod(line string) bool {
	methods := []string{"GET ", "POST ", "PUT ", "DELETE ", "PATCH ", "HEAD ", "OPTIONS ", "CONNECT ", "TRACE "}
	for _, method := range methods {
		if strings.HasPrefix(line, method) {
			return true
		}
	}
	return false
}

// ParseWithBlockSupport handles multiline blocks properly
func (hd *HTTPDSLv3) ParseWithBlockSupport(code string) (interface{}, error) {
	lines := strings.Split(code, "\n")
	var results []interface{}
	i := 0
	
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			i++
			continue
		}
		
		// Check if this is an HTTP request with multiple headers
		if isHTTPMethod(line) {
			// Collect the request line and any following headers
			requestParts := []string{line}
			j := i + 1
			
			// Look ahead for indented headers
			for j < len(lines) {
				nextLine := lines[j]
				trimmedNext := strings.TrimSpace(nextLine)
				
				// Check if this is an indented header
				if strings.HasPrefix(nextLine, "    ") && strings.HasPrefix(trimmedNext, "header ") {
					// Add this header to the request (inline)
					requestParts = append(requestParts, trimmedNext)
					j++
				} else {
					break
				}
			}
			
			// Build the complete request on one line
			fullRequest := strings.Join(requestParts, " ")
			
			// Parse the full request
			result, err := hd.ParseWithContext(fullRequest)
			if err != nil {
				return results, fmt.Errorf("error parsing HTTP request: %v", err)
			}
			if result != nil && result != "" {
				results = append(results, result)
			}
			
			i = j
			continue
		}
		
		// Check if this is an if block
		if strings.HasPrefix(line, "if ") && strings.HasSuffix(line, " then") {
			// Extract and evaluate the condition
			conditionStr := strings.TrimSuffix(strings.TrimPrefix(line, "if "), " then")
			
			// Evaluate condition using the new evaluator that supports AND/OR
			shouldExecute := hd.EvaluateCondition(conditionStr)
			
			// Collect the block lines
			i++
			var thenBlock []string
			var elseBlock []string
			inElse := false
			nestLevel := 1
			
			for i < len(lines) && nestLevel > 0 {
				innerLine := strings.TrimSpace(lines[i])
				
				if innerLine == "endif" {
					nestLevel--
					if nestLevel == 0 {
						break
					}
				} else if strings.HasPrefix(innerLine, "if ") && strings.HasSuffix(innerLine, " then") {
					nestLevel++
				} else if innerLine == "else" && nestLevel == 1 {
					inElse = true
					i++
					continue
				}
				
				if innerLine != "" && innerLine != "endif" && !strings.HasPrefix(innerLine, "#") {
					if inElse {
						elseBlock = append(elseBlock, innerLine)
					} else {
						thenBlock = append(thenBlock, innerLine)
					}
				}
				i++
			}
			
			// Execute the appropriate block
			var blockToExecute []string
			if shouldExecute {
				blockToExecute = thenBlock
			} else {
				blockToExecute = elseBlock
			}
			
			// Execute each line in the selected block
			for j := 0; j < len(blockToExecute); j++ {
				blockLine := blockToExecute[j]
				trimmedLine := strings.TrimSpace(blockLine)
				
				// Handle nested if blocks
				if strings.HasPrefix(trimmedLine, "if ") && strings.HasSuffix(trimmedLine, " then") {
					// Find the complete nested if block
					nestedBlock := []string{blockLine}
					nestCount := 1
					for k := j + 1; k < len(blockToExecute) && nestCount > 0; k++ {
						nestedLine := strings.TrimSpace(blockToExecute[k])
						nestedBlock = append(nestedBlock, blockToExecute[k])
						
						if strings.HasPrefix(nestedLine, "if ") && strings.HasSuffix(nestedLine, " then") {
							nestCount++
						} else if nestedLine == "endif" {
							nestCount--
							if nestCount == 0 {
								// Process the complete nested if block
								nestedCode := strings.Join(nestedBlock, "\n")
								result, err := hd.ParseWithBlockSupport(nestedCode)
								if err != nil {
									return results, fmt.Errorf("error in nested if block: %v", err)
								}
								if result != nil {
									// Add results from nested if
									if nestedResults, ok := result.([]interface{}); ok {
										results = append(results, nestedResults...)
									} else if result != "" {
										results = append(results, result)
									}
								}
								// Skip the lines we've processed
								j = k
								break
							}
						}
					}
				} else {
					// Regular line - parse normally
					result, err := hd.ParseWithContext(blockLine)
					if err != nil {
						return results, fmt.Errorf("error in block line '%s': %v", blockLine, err)
					}
					// Only add non-nil results
					if result != nil && result != "" {
						results = append(results, result)
					}
				}
			}
			
			// Don't add the temp variable result
			i++ // Skip the endif
			
		} else if strings.HasPrefix(line, "repeat ") && strings.HasSuffix(line, " do") {
			// Handle repeat blocks
			// Extract repeat count
			parts := strings.Fields(line)
			if len(parts) < 4 {
				return results, fmt.Errorf("invalid repeat syntax: %s", line)
			}
			
			// Parse the repeat count
			countStr := parts[1]
			var count int
			
			// Check if it's a variable
			if strings.HasPrefix(countStr, "$") {
				varName := strings.TrimPrefix(countStr, "$")
				if val, ok := hd.variables[varName]; ok {
					switch v := val.(type) {
					case int:
						count = v
					case float64:
						count = int(v)
					case string:
						fmt.Sscanf(v, "%d", &count)
					default:
						count = 0
					}
				}
			} else {
				// It's a literal number
				fmt.Sscanf(countStr, "%d", &count)
			}
			
			// Collect the loop body
			i++
			var loopBody []string
			nestLevel := 1
			
			for i < len(lines) && nestLevel > 0 {
				innerLine := strings.TrimSpace(lines[i])
				
				if innerLine == "endloop" {
					nestLevel--
					if nestLevel == 0 {
						break
					}
				} else if strings.HasSuffix(innerLine, " do") {
					nestLevel++
				}
				
				if innerLine != "" && innerLine != "endloop" && !strings.HasPrefix(innerLine, "#") {
					loopBody = append(loopBody, innerLine)
				}
				i++
			}
			
			// Execute the loop
			actualIterations := 0
			for iteration := 0; iteration < count; iteration++ {
				hd.SetVariable("_index", iteration)
				hd.SetVariable("_iteration", iteration + 1)
				
				// Use the new ProcessLoopBody function
				loopResult, err := hd.ProcessLoopBody(loopBody)
				if err != nil {
					return results, fmt.Errorf("error in loop iteration %d: %v", iteration+1, err)
				}
				
				// Append results
				for _, res := range loopResult.Results {
					if res != nil && res != "" {
						results = append(results, res)
					}
				}
				
				actualIterations++
				
				// Handle continue (skip to next iteration)
				if loopResult.ShouldContinue {
					continue
				}
				
				// Handle break
				if loopResult.ShouldBreak {
					break // Exit the repeat loop
				}
			}
			
			results = append(results, fmt.Sprintf("Repeated %d times", actualIterations))
			i++ // Skip the endloop
			
		} else if strings.HasPrefix(line, "while ") && strings.HasSuffix(line, " do") {
			// Handle while blocks
			// Extract condition
			conditionStr := strings.TrimSuffix(strings.TrimPrefix(line, "while "), " do")
			
			// Collect the loop body
			i++
			var loopBody []string
			nestLevel := 1
			
			for i < len(lines) && nestLevel > 0 {
				innerLine := strings.TrimSpace(lines[i])
				
				if innerLine == "endloop" {
					nestLevel--
					if nestLevel == 0 {
						break
					}
				} else if strings.HasSuffix(innerLine, " do") {
					nestLevel++
				}
				
				if innerLine != "" && innerLine != "endloop" && !strings.HasPrefix(innerLine, "#") {
					loopBody = append(loopBody, innerLine)
				}
				i++
			}
			
			// Execute the while loop
			maxIterations := 1000 // Safety limit
			iterations := 0
			
			for iterations < maxIterations {
				// Evaluate condition
				shouldContinue := false
				
				// Parse the condition (e.g., "$count < 10")
				parts := strings.Fields(conditionStr)
				if len(parts) == 3 {
					varName := strings.TrimPrefix(parts[0], "$")
					operator := parts[1]
					compareToStr := parts[2]
					
					if val, ok := hd.variables[varName]; ok {
						var numVal, compareVal float64
						switch v := val.(type) {
						case int:
							numVal = float64(v)
						case float64:
							numVal = v
						case string:
							fmt.Sscanf(v, "%f", &numVal)
						default:
							numVal = 0
						}
						fmt.Sscanf(compareToStr, "%f", &compareVal)
						
						switch operator {
						case "<":
							shouldContinue = numVal < compareVal
						case ">":
							shouldContinue = numVal > compareVal
						case "<=":
							shouldContinue = numVal <= compareVal
						case ">=":
							shouldContinue = numVal >= compareVal
						case "==":
							shouldContinue = numVal == compareVal
						case "!=":
							shouldContinue = numVal != compareVal
						}
					}
				}
				
				if !shouldContinue {
					break
				}
				
				hd.SetVariable("_iteration", iterations + 1)
				
				// Use the new ProcessLoopBody function
				loopResult, err := hd.ProcessLoopBody(loopBody)
				if err != nil {
					return results, fmt.Errorf("error in while loop iteration %d: %v", iterations+1, err)
				}
				
				// Append results
				for _, res := range loopResult.Results {
					if res != nil && res != "" {
						results = append(results, res)
					}
				}
				
				// Handle continue
				if loopResult.ShouldContinue {
					continue // Skip to next iteration
				}
				
				// Handle break
				if loopResult.ShouldBreak {
					break // Exit the while loop
				}
				
				iterations++
			}
			
			if iterations >= maxIterations {
				return results, fmt.Errorf("while loop exceeded maximum iterations (%d)", maxIterations)
			}
			
			results = append(results, fmt.Sprintf("While loop executed %d times", iterations))
			i++ // Skip the endloop
			
		} else if strings.HasPrefix(line, "foreach ") && strings.Contains(line, " in ") && strings.HasSuffix(line, " do") {
			// Handle foreach blocks
			// Extract item variable and list
			parts := strings.Split(line, " in ")
			if len(parts) != 2 {
				return results, fmt.Errorf("invalid foreach syntax: %s", line)
			}
			
			itemVar := strings.TrimPrefix(strings.TrimPrefix(parts[0], "foreach "), "$")
			listPart := strings.TrimSuffix(parts[1], " do")
			
			// Collect the loop body
			i++
			var loopBody []string
			nestLevel := 1
			
			for i < len(lines) && nestLevel > 0 {
				innerLine := strings.TrimSpace(lines[i])
				
				if innerLine == "endloop" {
					nestLevel--
					if nestLevel == 0 {
						break
					}
				} else if strings.HasSuffix(innerLine, " do") {
					nestLevel++
				}
				
				if innerLine != "" && innerLine != "endloop" && !strings.HasPrefix(innerLine, "#") {
					loopBody = append(loopBody, innerLine)
				}
				i++
			}
			
			// Get the list to iterate over
			var items []interface{}
			
			// Check if it's a JSON array literal
			if strings.HasPrefix(listPart, "[") && strings.HasSuffix(listPart, "]") {
				// Parse JSON array
				listStr := listPart
				// Simple parsing for string arrays like ["apple", "banana", "orange"]
				listStr = strings.Trim(listStr, "[]")
				parts := strings.Split(listStr, ",")
				for _, part := range parts {
					item := strings.TrimSpace(part)
					item = strings.Trim(item, "\"'")
					items = append(items, item)
				}
			} else if strings.HasPrefix(listPart, "$") {
				// It's a variable reference
				varName := strings.TrimPrefix(listPart, "$")
				if val, ok := hd.variables[varName]; ok {
					switch v := val.(type) {
					case []interface{}:
						items = v
					case []string:
						for _, s := range v {
							items = append(items, s)
						}
					case string:
						// Try to parse as JSON array
						if strings.HasPrefix(v, "[") {
							v = strings.Trim(v, "[]")
							parts := strings.Split(v, ",")
							for _, part := range parts {
								item := strings.TrimSpace(part)
								item = strings.Trim(item, "\"'")
								items = append(items, item)
							}
						}
					}
				}
			}
			
			// Execute the foreach loop
			actualIterations := 0
			for idx, item := range items {
				hd.SetVariable(itemVar, item)
				hd.SetVariable("_index", idx)
				hd.SetVariable("_iteration", idx + 1)
				
				// Use the new ProcessLoopBody function
				loopResult, err := hd.ProcessLoopBody(loopBody)
				if err != nil {
					return results, fmt.Errorf("error in foreach iteration %d: %v", idx+1, err)
				}
				
				// Append results
				for _, res := range loopResult.Results {
					if res != nil && res != "" {
						results = append(results, res)
					}
				}
				
				actualIterations++
				
				// Handle continue (skip to next iteration)
				if loopResult.ShouldContinue {
					continue
				}
				
				// Handle break
				if loopResult.ShouldBreak {
					break // Exit the foreach loop
				}
			}
			
			results = append(results, fmt.Sprintf("Foreach executed for %d items", actualIterations))
			i++ // Skip the endloop
			
		} else {
			// Special handling for single-line if/then/else to avoid double execution
			if strings.HasPrefix(line, "if ") && strings.Contains(line, " then ") && strings.Contains(line, " else ") && !strings.Contains(line, "endif") {
				// Parse if/then/else manually to avoid both branches executing
				// Find the positions of "then" and "else"
				parts := strings.SplitN(line, " then ", 2)
				if len(parts) == 2 {
					conditionPart := strings.TrimPrefix(parts[0], "if ")
					restParts := strings.SplitN(parts[1], " else ", 2)
					if len(restParts) == 2 {
						thenStatement := restParts[0]
						elseStatement := restParts[1]
						
						// Evaluate the condition directly
						shouldExecuteThen := false
						
						// Parse the condition (e.g., "$x > 10")
						condParts := strings.Fields(conditionPart)
						if len(condParts) == 3 {
							// Simple comparison like "$x > 10"
							varName := strings.TrimPrefix(condParts[0], "$")
							operator := condParts[1]
							compareToStr := condParts[2]
							
							if val, ok := hd.variables[varName]; ok {
								var numVal, compareVal float64
								// Convert to numbers
								switch v := val.(type) {
								case int:
									numVal = float64(v)
								case float64:
									numVal = v
								case string:
									fmt.Sscanf(v, "%f", &numVal)
								default:
									numVal = 0
								}
								fmt.Sscanf(compareToStr, "%f", &compareVal)
								
								// Evaluate comparison
								switch operator {
								case ">":
									shouldExecuteThen = numVal > compareVal
								case "<":
									shouldExecuteThen = numVal < compareVal
								case ">=":
									shouldExecuteThen = numVal >= compareVal
								case "<=":
									shouldExecuteThen = numVal <= compareVal
								case "==":
									shouldExecuteThen = numVal == compareVal
								case "!=":
									shouldExecuteThen = numVal != compareVal
								}
							}
						}
						
						// Execute the appropriate branch
						if shouldExecuteThen {
							// Execute ONLY the then branch
							result, err := hd.ParseWithContext(thenStatement)
							if err != nil {
								return results, fmt.Errorf("error in then statement: %v", err)
							}
							results = append(results, result)
						} else {
							// Execute ONLY the else branch
							result, err := hd.ParseWithContext(elseStatement)
							if err != nil {
								return results, fmt.Errorf("error in else statement: %v", err)
							}
							results = append(results, result)
						}
						i++
						continue
					}
				}
			}
			
			// Regular line - parse normally
			result, err := hd.ParseWithContext(line)
			if err != nil {
				return results, fmt.Errorf("error at line %d: %v", i+1, err)
			}
			results = append(results, result)
			i++
		}
	}
	
	return results, nil
}