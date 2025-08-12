// Package dslbuilder - Multiline support extension
package dslbuilder

import (
	"strings"
)

// ParseMultiline parses multiple DSL statements separated by newlines.
// This method provides backward-compatible multiline support without
// modifying the core parsing logic.
//
// The method processes each non-empty line as a separate statement
// and returns all results. If any line fails to parse, it returns
// the error with line information.
//
// Example:
//
//	code := `
//	set $x 10
//	if $x > 5 then set $result "big"
//	print "Result: $result"
//	`
//	results, err := dsl.ParseMultiline(code)
func (d *DSL) ParseMultiline(code string) ([]interface{}, error) {
	lines := strings.Split(code, "\n")
	var results []interface{}
	
	for lineNum, line := range lines {
		// Skip empty lines and comments
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") {
			continue
		}
		
		// Parse the line
		result, err := d.Parse(trimmed)
		if err != nil {
			// Enhance error with line number
			if parseErr, ok := err.(*ParseError); ok {
				parseErr.Line = lineNum + 1
				return results, parseErr
			}
			return results, &ParseError{
				Message:  err.Error(),
				Line:     lineNum + 1,
				Column:   1,
				Token:    trimmed,
				Input:    code,
			}
		}
		
		results = append(results, result)
	}
	
	return results, nil
}

// ParseAuto automatically detects if input is multiline and parses accordingly.
// This provides a smart parsing mode that handles both single and multiline
// inputs transparently.
//
// Detection logic:
//   - If input contains newlines and parsing as single statement fails,
//     it automatically retries as multiline
//   - If input has no newlines, parses as single statement
//
// Returns:
//   - Single result for single-line input
//   - Array of results for multi-line input
//   - Error if parsing fails
func (d *DSL) ParseAuto(code string) (interface{}, error) {
	// First, try parsing as a single statement
	result, err := d.Parse(code)
	
	// If successful or no newlines, return as-is
	if err == nil || !strings.Contains(code, "\n") {
		return result, err
	}
	
	// If failed and has newlines, try multiline parsing
	results, multiErr := d.ParseMultiline(code)
	if multiErr != nil {
		// Return original error if multiline also fails
		return nil, err
	}
	
	// Return results from multiline parsing
	if len(results) == 1 {
		return results[0], nil
	}
	return results, nil
}

// ParseStatements parses input that may contain multiple statements.
// Statements can be separated by newlines or semicolons.
// This is useful for script-like DSLs where multiple operations
// need to be executed in sequence.
//
// Separator priority:
//  1. Semicolons (;) if present
//  2. Newlines (\n) otherwise
//
// Example:
//
//	// With semicolons
//	code := "set $x 10; set $y 20; print $x"
//	
//	// With newlines
//	code := `
//	set $x 10
//	set $y 20
//	print $x
//	`
func (d *DSL) ParseStatements(code string) ([]interface{}, error) {
	// Check if code uses semicolons as separator
	if strings.Contains(code, ";") {
		// Split by semicolon and process each
		statements := strings.Split(code, ";")
		var results []interface{}
		
		for i, stmt := range statements {
			trimmed := strings.TrimSpace(stmt)
			if trimmed == "" {
				continue
			}
			
			result, err := d.Parse(trimmed)
			if err != nil {
				// Add position context
				if parseErr, ok := err.(*ParseError); ok {
					parseErr.Message = "in statement " + string(rune(i+1)) + ": " + parseErr.Message
				}
				return results, err
			}
			results = append(results, result)
		}
		return results, nil
	}
	
	// Otherwise use newline separation
	return d.ParseMultiline(code)
}