package dslbuilder

import (
	"errors"
	"strings"
	"testing"
)

func TestParseErrorBackwardCompatibility(t *testing.T) {
	dsl := New("TestErrorCompatibility")
	
	// Define a simple DSL that will fail
	dsl.KeywordToken("HELLO", "hello")
	dsl.KeywordToken("WORLD", "world")
	dsl.Rule("greeting", []string{"HELLO", "WORLD"}, "greet")
	dsl.Action("greet", func(args []interface{}) (interface{}, error) {
		return "Hello, World!", nil
	})
	
	// Test 1: Error should still work as normal error interface
	_, err := dsl.Parse("hello invalid")
	if err == nil {
		t.Fatal("Expected error but got none")
	}
	
	// Test 2: Error.Error() should return original message format
	errorMessage := err.Error()
	if errorMessage == "" {
		t.Fatal("Error message should not be empty")
	}
	
	// Test 3: Should be able to check if it's a ParseError
	if !IsParseError(err) {
		t.Fatal("Error should be a ParseError")
	}
	
	// Test 4: Should be able to get detailed error
	detailedError := GetDetailedError(err)
	if !strings.Contains(detailedError, "line") || !strings.Contains(detailedError, "column") {
		t.Fatalf("Detailed error should contain line and column info: %s", detailedError)
	}
	
	// Test 5: Normal string errors should still work
	normalErr := errors.New("normal error") // This is not a ParseError
	if IsParseError(normalErr) {
		t.Fatal("Normal error should not be identified as ParseError")
	}
	
	normalDetailed := GetDetailedError(normalErr)
	if normalDetailed != normalErr.Error() {
		t.Fatal("Normal error detailed should be same as Error()")
	}
}

func TestParseErrorLineColumn(t *testing.T) {
	dsl := New("TestLineColumn")
	
	// Define DSL
	dsl.KeywordToken("LINE1", "line1")
	dsl.KeywordToken("LINE2", "line2")
	dsl.Rule("multiline", []string{"LINE1", "LINE2"}, "process")
	dsl.Action("process", func(args []interface{}) (interface{}, error) {
		return "processed", nil
	})
	
	// Test multiline input with error on second line
	input := "line1\ninvalid"
	_, err := dsl.Parse(input)
	
	if err == nil {
		t.Fatal("Expected error but got none")
	}
	
	if !IsParseError(err) {
		t.Fatal("Error should be a ParseError")
	}
	
	parseErr := err.(*ParseError)
	
	// Should point to line 2, column 1 (where "invalid" starts)
	if parseErr.Line != 2 {
		t.Fatalf("Expected line 2, got %d", parseErr.Line)
	}
	
	if parseErr.Column != 1 {
		t.Fatalf("Expected column 1, got %d", parseErr.Column)
	}
	
	// Test detailed error format
	detailed := parseErr.DetailedError()
	if !strings.Contains(detailed, "line 2, column 1") {
		t.Fatalf("Detailed error should contain position info: %s", detailed)
	}
	
	// Should contain context line and pointer
	if !strings.Contains(detailed, "invalid") {
		t.Fatalf("Detailed error should contain context: %s", detailed)
	}
	
	if !strings.Contains(detailed, "^") {
		t.Fatalf("Detailed error should contain pointer: %s", detailed)
	}
}

func TestCalculateLineColumn(t *testing.T) {
	input := "hello\nworld\ntest"
	
	tests := []struct {
		position int
		line     int
		column   int
	}{
		{0, 1, 1},   // 'h' in "hello"
		{4, 1, 5},   // 'o' in "hello"
		{5, 1, 6},   // '\n' after "hello"
		{6, 2, 1},   // 'w' in "world"
		{11, 2, 6},  // '\n' after "world"
		{12, 3, 1},  // 't' in "test"
		{15, 3, 4},  // 't' at end of "test"
	}
	
	for _, test := range tests {
		line, column := calculateLineColumn(input, test.position)
		if line != test.line || column != test.column {
			t.Errorf("Position %d: expected line %d, column %d; got line %d, column %d",
				test.position, test.line, test.column, line, column)
		}
	}
}

func TestCreateParseError(t *testing.T) {
	input := "hello\ninvalid token"
	position := 6 // 'i' in "invalid"
	token := "invalid"
	message := "unexpected token"
	
	err := createParseError(message, position, token, input)
	
	if err.Message != message {
		t.Fatalf("Expected message '%s', got '%s'", message, err.Message)
	}
	
	if err.Line != 2 {
		t.Fatalf("Expected line 2, got %d", err.Line)
	}
	
	if err.Column != 1 {
		t.Fatalf("Expected column 1, got %d", err.Column)
	}
	
	if err.Position != position {
		t.Fatalf("Expected position %d, got %d", position, err.Position)
	}
	
	if err.Token != token {
		t.Fatalf("Expected token '%s', got '%s'", token, err.Token)
	}
	
	if err.Input != input {
		t.Fatalf("Expected input '%s', got '%s'", input, err.Input)
	}
	
	// Test Error() method (backward compatibility)
	if err.Error() != message {
		t.Fatalf("Error() should return original message: got '%s'", err.Error())
	}
}

func TestExistingExamplesStillWork(t *testing.T) {
	// Test that existing examples still work with new error system
	dsl := New("BackwardCompatTest")
	
	// Simple working case
	dsl.KeywordToken("VENTA", "venta")
	dsl.KeywordToken("DE", "de")
	dsl.Token("IMPORTE", "[0-9]+")
	dsl.Rule("command", []string{"VENTA", "DE", "IMPORTE"}, "sale")
	dsl.Action("sale", func(args []interface{}) (interface{}, error) {
		return "Sale processed", nil
	})
	
	// This should work fine
	result, err := dsl.Parse("venta de 5000")
	if err != nil {
		t.Fatalf("Valid input should not error: %v", err)
	}
	
	if result.GetOutput() != "Sale processed" {
		t.Fatalf("Expected 'Sale processed', got %v", result.GetOutput())
	}
	
	// This should fail gracefully with enhanced error info
	_, err = dsl.Parse("venta de invalid_amount")
	if err == nil {
		t.Fatal("Invalid input should error")
	}
	
	// But error should still work as before for backward compatibility
	errorStr := err.Error()
	if errorStr == "" {
		t.Fatal("Error message should not be empty")
	}
	
	// And enhanced error should be available
	if IsParseError(err) {
		detailed := GetDetailedError(err)
		if !strings.Contains(detailed, "line") {
			t.Fatal("Enhanced error should contain line information")
		}
	}
}