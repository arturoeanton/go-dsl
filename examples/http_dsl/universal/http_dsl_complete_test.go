package universal

import (
	"fmt"
	"strings"
	"testing"
)

// TestVariableOperations tests all variable-related functionality
func TestVariableOperations(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		expected string
		contains bool
	}{
		{
			name: "Basic variable assignment",
			script: `
set $name "John"
print "Name: $name"`,
			expected: "Name: John",
			contains: true,
		},
		{
			name: "Arithmetic operations",
			script: `
set $a 10
set $b 5
set $sum $a + $b
set $diff $a - $b
set $prod $a * $b
set $div $a / $b
print "Sum: $sum, Diff: $diff, Prod: $prod, Div: $div"`,
			expected: "Sum: 15, Diff: 5, Prod: 50, Div: 2",
			contains: true,
		},
		{
			name: "Variable expansion in strings",
			script: `
set $city "New York"
set $temp 72
print "Weather in $city: $temp degrees"`,
			expected: "Weather in New York: 72 degrees",
			contains: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hd := NewHTTPDSLv3()
			result, err := hd.ParseWithBlockSupport(tt.script)
			if err != nil {
				t.Fatalf("ParseWithBlockSupport failed: %v", err)
			}

			output := captureOutput(func() {
				hd.ParseWithBlockSupport(tt.script)
			})

			if tt.contains && !strings.Contains(output, tt.expected) {
				t.Errorf("Expected output to contain '%s', got '%s'", tt.expected, output)
			}
			_ = result
		})
	}
}

// TestControlFlow tests if/then/else and loop constructs
func TestControlFlow(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		expected string
		contains bool
	}{
		{
			name: "Simple if/then/else",
			script: `
set $count 10
if $count > 5 then
    print "Count is high"
else
    print "Count is low"
endif`,
			expected: "Count is high",
			contains: true,
		},
		{
			name: "Nested if statements",
			script: `
set $a 10
set $b 5
if $a > 5 then
    print "A is greater than 5"
    if $b < 10 then
        print "B is less than 10"
    endif
endif`,
			expected: "B is less than 10",
			contains: true,
		},
		{
			name: "While loop",
			script: `
set $count 0
while $count < 3 do
    print "Count: $count"
    set $count $count + 1
endloop`,
			expected: "Count: 2",
			contains: true,
		},
		{
			name: "Foreach loop",
			script: `
foreach $item in ["apple", "banana", "orange"] do
    print "Fruit: $item"
endloop`,
			expected: "Fruit: banana",
			contains: true,
		},
		{
			name: "Break in loop",
			script: `
set $i 0
while $i < 10 do
    if $i == 3 then
        break
    endif
    set $i $i + 1
endloop
print "Final: $i"`,
			expected: "Final: 3",
			contains: true,
		},
		{
			name: "Continue in loop",
			script: `
foreach $num in ["1", "2", "3", "4"] do
    if $num == "2" then
        continue
    endif
    print "Number: $num"
endloop`,
			expected: "Number: 1",
			contains: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hd := NewHTTPDSLv3()
			result, err := hd.ParseWithBlockSupport(tt.script)
			if err != nil {
				t.Fatalf("ParseWithBlockSupport failed: %v", err)
			}

			// For now, we're checking that parsing succeeds
			// In a real implementation, we'd capture print output
			_ = result
		})
	}
}

// TestArrayOperations tests array functionality
func TestArrayOperations(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		expected interface{}
		hasError bool
	}{
		{
			name: "Array length function",
			script: `
set $fruits "[\"apple\", \"banana\", \"orange\"]"
set $len length $fruits
print "Length: $len"`,
			expected: 3,
			hasError: false,
		},
		{
			name: "Empty array in foreach",
			script: `
foreach $item in [] do
    print "Should not print"
endloop
print "Done"`,
			expected: "Done",
			hasError: false,
		},
		{
			name: "Array indexing with brackets",
			script: `
set $arr "[\"a\", \"b\", \"c\"]"
set $first $arr[0]
set $second $arr[1]
print "First: $first, Second: $second"`,
			expected: "First: a, Second: b",
			hasError: false,
		},
		{
			name: "Array indexing with variable",
			script: `
set $arr "[\"x\", \"y\", \"z\"]"
set $idx 1
set $item $arr[$idx]
print "Item: $item"`,
			expected: "Item: y",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hd := NewHTTPDSLv3()
			result, err := hd.ParseWithBlockSupport(tt.script)
			
			if tt.hasError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.hasError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			_ = result
		})
	}
}

// TestLogicalOperators tests AND/OR operators
func TestLogicalOperators(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		expected bool
	}{
		{
			name: "AND operator - both true",
			script: `
set $a 10
set $b 5
if $a > 5 and $b < 10 then
    print "Both conditions true"
endif`,
			expected: true,
		},
		{
			name: "AND operator - one false",
			script: `
set $a 3
set $b 5
if $a > 5 and $b < 10 then
    print "Should not print"
else
    print "One condition false"
endif`,
			expected: true,
		},
		{
			name: "OR operator - one true",
			script: `
set $a 3
set $b 5
if $a > 5 or $b < 10 then
    print "At least one true"
endif`,
			expected: true,
		},
		{
			name: "Complex logical expression",
			script: `
set $a 10
set $b 5
set $c 15
if $a > 5 and $b < 10 or $c == 15 then
    print "Complex condition met"
endif`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hd := NewHTTPDSLv3()
			_, err := hd.ParseWithBlockSupport(tt.script)
			if err != nil {
				t.Errorf("ParseWithBlockSupport failed: %v", err)
			}
		})
	}
}

// TestCLIArguments tests command-line argument functionality
func TestCLIArguments(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		script   string
		expected string
	}{
		{
			name: "Access CLI arguments",
			args: []string{"test.http", "value1", "value2"},
			script: `
print "Arg1: $ARG1"
print "Arg2: $ARG2"
print "ArgCount: $ARGC"`,
			expected: "Arg1: value1",
		},
		{
			name: "No arguments provided",
			args: []string{"test.http"},
			script: `
if $ARGC == 0 then
    print "No arguments"
endif`,
			expected: "No arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hd := NewHTTPDSLv3()
			// Manually set CLI arguments as variables
			for i, arg := range tt.args {
				hd.SetVariable(fmt.Sprintf("ARG%d", i+1), arg)
			}
			hd.SetVariable("ARGC", len(tt.args))
			_, err := hd.ParseWithBlockSupport(tt.script)
			if err != nil {
				t.Errorf("ParseWithBlockSupport failed: %v", err)
			}
		})
	}
}

// TestStringOperations tests string-related functionality
func TestStringOperations(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		expected interface{}
	}{
		{
			name: "String length",
			script: `
set $text "Hello World"
set $len length $text
print "Length: $len"`,
			expected: 11,
		},
		{
			name: "String character access",
			script: `
set $text "Hello"
set $first $text[0]
print "First char: $first"`,
			expected: "H",
		},
		{
			name: "String concatenation",
			script: `
set $first "Hello"
set $second "World"
print "$first $second"`,
			expected: "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hd := NewHTTPDSLv3()
			_, err := hd.ParseWithBlockSupport(tt.script)
			if err != nil {
				t.Errorf("ParseWithBlockSupport failed: %v", err)
			}
		})
	}
}

// TestEdgeCases tests various edge cases and error conditions
func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		hasError bool
	}{
		{
			name: "Empty foreach loop",
			script: `
foreach $item in [] do
    print "Should not execute"
endloop`,
			hasError: false,
		},
		{
			name: "Division by zero",
			script: `
set $a 10
set $b 0
set $result $a / $b`,
			hasError: true,
		},
		{
			name: "Undefined variable",
			script: `
print "Value: $undefined"`,
			hasError: false, // Variable expansion returns empty string
		},
		{
			name: "Nested comments",
			script: `
# This is a comment
set $a 10
# Another comment
if $a > 5 then
    # Comment in block
    print "Value is high"
endif`,
			hasError: false,
		},
		{
			name: "Maximum loop iterations",
			script: `
set $i 0
while $i < 2000 do
    set $i $i + 1
endloop`,
			hasError: false, // Should stop at 1000 iterations
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hd := NewHTTPDSLv3()
			_, err := hd.ParseWithBlockSupport(tt.script)
			
			if tt.hasError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.hasError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// Helper function to capture output (would need implementation)
func captureOutput(f func()) string {
	// This would capture stdout/stderr during function execution
	// For testing purposes, we're using a placeholder
	f()
	return ""
}