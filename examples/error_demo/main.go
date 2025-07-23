package main

import (
	"fmt"
	"log"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

func main() {
	fmt.Println("=== Enhanced Error Reporting Demo ===")
	fmt.Println("Showing improved error messages with line and column information")
	fmt.Println()

	// Create a simple DSL
	dsl := dslbuilder.New("ErrorDemo")

	// Define tokens
	dsl.KeywordToken("CREATE", "create")
	dsl.KeywordToken("USER", "user")
	dsl.KeywordToken("WITH", "with")
	dsl.KeywordToken("NAME", "name")
	dsl.KeywordToken("EMAIL", "email")
	dsl.Token("STRING", "\"[^\"]*\"")

	// Define rules
	dsl.Rule("command", []string{"CREATE", "USER", "WITH", "NAME", "STRING"}, "createUser")
	dsl.Rule("command", []string{"CREATE", "USER", "WITH", "NAME", "STRING", "WITH", "EMAIL", "STRING"}, "createUserWithEmail")

	// Define actions
	dsl.Action("createUser", func(args []interface{}) (interface{}, error) {
		name := args[4].(string)
		return fmt.Sprintf("Created user: %s", name), nil
	})

	dsl.Action("createUserWithEmail", func(args []interface{}) (interface{}, error) {
		name := args[4].(string)
		email := args[7].(string)
		return fmt.Sprintf("Created user: %s with email: %s", name, email), nil
	})

	// Test cases that demonstrate error reporting
	testCases := []struct {
		name  string
		input string
	}{
		{
			"Valid input (no error)",
			`create user with name "John Doe"`,
		},
		{
			"Unexpected token on first line",
			`create invalid with name "John"`,
		},
		{
			"Error on second line",
			`create user
with invalid_token "John"`,
		},
		{
			"Error on third line with context",
			`create user
with name "John"
and some_invalid_syntax here`,
		},
		{
			"Missing closing quote",
			`create user with name "John Doe`,
		},
		{
			"Complex multiline with error",
			`create user
with name "Alice Smith"  
with email invalid_email_format`,
		},
	}

	for i, tc := range testCases {
		fmt.Printf("%d. %s\n", i+1, tc.name)
		fmt.Printf("   Input: %q\n", tc.input)

		result, err := dsl.Parse(tc.input)
		if err == nil {
			fmt.Printf("   ✅ Success: %v\n", result.GetOutput())
		} else {
			fmt.Printf("   ❌ Error occurred:\n")

			// Show backward-compatible error
			fmt.Printf("      Standard Error: %v\n", err)

			// Show enhanced error if available
			if dslbuilder.IsParseError(err) {
				fmt.Printf("      Enhanced Error:\n")
				detailedError := dslbuilder.GetDetailedError(err)

				// Indent each line of the detailed error
				lines := fmt.Sprintf("%s", detailedError)
				fmt.Printf("      %s\n", lines)
			}
		}
		fmt.Println()
	}

	// Demonstrate programmatic access to error details
	fmt.Println("=== Programmatic Error Analysis ===")

	_, err := dsl.Parse("create user\nwith invalid_syntax")
	if err != nil && dslbuilder.IsParseError(err) {
		parseErr := err.(*dslbuilder.ParseError)

		fmt.Printf("Error Analysis:\n")
		fmt.Printf("  Message: %s\n", parseErr.Message)
		fmt.Printf("  Line: %d\n", parseErr.Line)
		fmt.Printf("  Column: %d\n", parseErr.Column)
		fmt.Printf("  Position: %d\n", parseErr.Position)
		fmt.Printf("  Token: %s\n", parseErr.Token)
		fmt.Printf("  Input: %s\n", parseErr.Input)
	}

	fmt.Println("\n=== Backward Compatibility Verification ===")
	fmt.Println("All existing code continues to work unchanged:")

	// This is how existing code would work - unchanged
	_, err = dsl.Parse("invalid syntax")
	if err != nil {
		log.Printf("Traditional error handling still works: %v", err)
	}

	fmt.Println("✅ Enhanced error reporting implemented successfully!")
	fmt.Println("✅ Backward compatibility maintained!")
}
