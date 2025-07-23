package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
	"gopkg.in/yaml.v3"
)

type ValidationResult struct {
	Valid    bool
	Errors   []ValidationError
	Warnings []ValidationWarning
	Info     DSLInfo
}

type ValidationError struct {
	Type    string
	Message string
	Details string
}

type ValidationWarning struct {
	Type    string
	Message string
	Details string
}

type DSLInfo struct {
	Name       string
	TokenCount int
	RuleCount  int
	Tokens     []TokenInfo
	Rules      []RuleInfo
}

type TokenInfo struct {
	Name     string
	Pattern  string
	Priority int
	Valid    bool
	Error    string
}

type RuleInfo struct {
	Name    string
	Pattern []string
	Action  string
	Valid   bool
	Error   string
}

func main() {
	var (
		dslFile    string
		verbose    bool
		format     string
		testInput  string
		showInfo   bool
		strictMode bool
	)

	flag.StringVar(&dslFile, "dsl", "", "DSL configuration file to validate (YAML or JSON)")
	flag.BoolVar(&verbose, "verbose", false, "Show detailed validation information")
	flag.StringVar(&format, "format", "text", "Output format: text, json, or yaml")
	flag.StringVar(&testInput, "test", "", "Test input string to validate against the DSL")
	flag.BoolVar(&showInfo, "info", false, "Show DSL information summary")
	flag.BoolVar(&strictMode, "strict", false, "Enable strict validation mode")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "DSL Validator - Validate DSL grammar and detect potential issues\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -dsl calculator.yaml\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -dsl query.json -verbose -info\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -dsl accounting.yaml -test \"venta de 1000\" -strict\n", os.Args[0])
	}

	flag.Parse()

	if dslFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Validate DSL
	result := validateDSL(dslFile, verbose, strictMode)

	// Test input if provided
	if testInput != "" {
		testResult := testDSLInput(dslFile, testInput)
		if !testResult {
			result.Errors = append(result.Errors, ValidationError{
				Type:    "ParseError",
				Message: fmt.Sprintf("Failed to parse test input: %s", testInput),
				Details: "The provided test input could not be parsed with this DSL",
			})
			result.Valid = false
		} else if verbose {
			fmt.Printf("✓ Test input parsed successfully: %s\n", testInput)
		}
	}

	// Output results
	switch format {
	case "json":
		outputJSON(result, showInfo)
	case "yaml":
		outputYAML(result, showInfo)
	default:
		outputText(result, showInfo, verbose)
	}

	// Exit with appropriate code
	if !result.Valid {
		os.Exit(1)
	}
}

func validateDSL(filename string, verbose bool, strict bool) ValidationResult {
	result := ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
	}

	// Load DSL configuration
	config, err := loadDSLConfig(filename)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Type:    "LoadError",
			Message: "Failed to load DSL configuration",
			Details: err.Error(),
		})
		return result
	}

	result.Info.Name = config.Name

	// Validate tokens
	validateTokens(config, &result, strict)

	// Validate rules
	validateRules(config, &result, strict)

	// Check for common issues
	checkCommonIssues(config, &result, strict)

	// Try to create DSL instance
	if result.Valid {
		if _, err := createDSLFromConfig(config); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Type:    "CreationError",
				Message: "Failed to create DSL instance",
				Details: err.Error(),
			})
		}
	}

	return result
}

func loadDSLConfig(filename string) (*DSLConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config DSLConfig
	ext := strings.ToLower(filename[strings.LastIndex(filename, ".")+1:])

	switch ext {
	case "yaml", "yml":
		err = yaml.Unmarshal(data, &config)
	case "json":
		err = json.Unmarshal(data, &config)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}

	if err != nil {
		return nil, err
	}

	return &config, nil
}

type DSLConfig struct {
	Name    string                 `yaml:"name" json:"name"`
	Tokens  map[string]string      `yaml:"tokens" json:"tokens"`
	Rules   []RuleConfig           `yaml:"rules" json:"rules"`
	Context map[string]interface{} `yaml:"context,omitempty" json:"context,omitempty"`
}

type RuleConfig struct {
	Name    string   `yaml:"name" json:"name"`
	Pattern []string `yaml:"pattern" json:"pattern"`
	Action  string   `yaml:"action" json:"action"`
}

func validateTokens(config *DSLConfig, result *ValidationResult, strict bool) {
	tokenCount := 0
	for name, pattern := range config.Tokens {
		tokenCount++
		tokenInfo := TokenInfo{
			Name:    name,
			Pattern: pattern,
			Valid:   true,
		}

		// Validate regex pattern
		if _, err := regexp.Compile(pattern); err != nil {
			tokenInfo.Valid = false
			tokenInfo.Error = fmt.Sprintf("Invalid regex pattern: %v", err)
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Type:    "TokenError",
				Message: fmt.Sprintf("Invalid regex pattern for token %s", name),
				Details: err.Error(),
			})
		}

		// Check for common regex issues
		if tokenInfo.Valid {
			checkTokenPattern(name, pattern, result, strict)
		}

		// Determine priority (simplified - would need DSL internals for accurate priority)
		if strings.Contains(pattern, "[") || strings.Contains(pattern, "\\") {
			tokenInfo.Priority = 0 // Regular pattern
		} else {
			tokenInfo.Priority = 90 // Keyword pattern
		}

		result.Info.Tokens = append(result.Info.Tokens, tokenInfo)
	}
	result.Info.TokenCount = tokenCount
}

func checkTokenPattern(name, pattern string, result *ValidationResult, strict bool) {
	// Check for overly broad patterns
	if pattern == ".*" || pattern == ".+" {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Type:    "BroadPattern",
			Message: fmt.Sprintf("Token %s has overly broad pattern: %s", name, pattern),
			Details: "This pattern will match everything and may cause parsing issues",
		})
	}

	// Check for unescaped special characters
	specialChars := []string{"+", "*", "?", "(", ")", "[", "]", "{", "}"}
	for _, char := range specialChars {
		if strings.Contains(pattern, char) && !strings.Contains(pattern, "\\"+char) {
			// Check if it's actually a regex construct
			if !isRegexConstruct(pattern, char) && strict {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Type:    "UnescapedChar",
					Message: fmt.Sprintf("Token %s may have unescaped special character: %s", name, char),
					Details: "Consider escaping special characters if they should be matched literally",
				})
			}
		}
	}
}

func isRegexConstruct(pattern, char string) bool {
	// Simplified check for valid regex constructs
	switch char {
	case "+", "*", "?":
		return true // Quantifiers
	case "[", "]":
		return strings.Contains(pattern, "[") && strings.Contains(pattern, "]")
	case "(", ")":
		return strings.Contains(pattern, "(") && strings.Contains(pattern, ")")
	case "{", "}":
		return strings.Contains(pattern, "{") && strings.Contains(pattern, "}")
	}
	return false
}

func validateRules(config *DSLConfig, result *ValidationResult, strict bool) {
	ruleCount := 0
	ruleNames := make(map[string]bool)
	actionNames := make(map[string]bool)

	for _, rule := range config.Rules {
		ruleCount++
		ruleInfo := RuleInfo{
			Name:    rule.Name,
			Pattern: rule.Pattern,
			Action:  rule.Action,
			Valid:   true,
		}

		// Check for duplicate rule names (allowed but may be confusing)
		if ruleNames[rule.Name] && strict {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Type:    "DuplicateRule",
				Message: fmt.Sprintf("Multiple rules with name: %s", rule.Name),
				Details: "Multiple rules with the same name are allowed but may be confusing",
			})
		}
		ruleNames[rule.Name] = true

		// Track action names
		actionNames[rule.Action] = true

		// Validate pattern tokens exist
		for _, token := range rule.Pattern {
			if _, exists := config.Tokens[token]; !exists {
				// Check if it's a rule reference (for recursive grammars)
				if !ruleNames[token] {
					ruleInfo.Valid = false
					ruleInfo.Error = fmt.Sprintf("Unknown token or rule: %s", token)
					result.Valid = false
					result.Errors = append(result.Errors, ValidationError{
						Type:    "RuleError",
						Message: fmt.Sprintf("Rule %s references unknown token/rule: %s", rule.Name, token),
						Details: "All tokens in rule patterns must be defined",
					})
				}
			}
		}

		// Check for empty patterns
		if len(rule.Pattern) == 0 {
			ruleInfo.Valid = false
			ruleInfo.Error = "Empty pattern"
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Type:    "RuleError",
				Message: fmt.Sprintf("Rule %s has empty pattern", rule.Name),
				Details: "Rules must have at least one token in their pattern",
			})
		}

		result.Info.Rules = append(result.Info.Rules, ruleInfo)
	}
	result.Info.RuleCount = ruleCount

	// Warn about unused actions (actions without implementations)
	if strict {
		for action := range actionNames {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Type:    "UnimplementedAction",
				Message: fmt.Sprintf("Action %s is referenced but not implemented", action),
				Details: "Make sure to implement all actions referenced in rules",
			})
		}
	}
}

func checkCommonIssues(config *DSLConfig, result *ValidationResult, strict bool) {
	// Check for left recursion
	for _, rule := range config.Rules {
		if len(rule.Pattern) > 0 && rule.Pattern[0] == rule.Name {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Type:    "LeftRecursion",
				Message: fmt.Sprintf("Rule %s has left recursion", rule.Name),
				Details: "Left recursion is supported but may impact performance",
			})
		}
	}

	// Check for ambiguous token patterns
	tokenPatterns := make(map[string][]string)
	for name, pattern := range config.Tokens {
		if existing, exists := tokenPatterns[pattern]; exists {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Type:    "DuplicatePattern",
				Message: fmt.Sprintf("Tokens %s and %s have identical patterns", name, existing[0]),
				Details: "Multiple tokens with the same pattern may cause confusion",
			})
		}
		tokenPatterns[pattern] = append(tokenPatterns[pattern], name)
	}

	// Check for missing start rule
	if _, hasExpr := findStartRule(config); !hasExpr && strict {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Type:    "NoStartRule",
			Message: "No obvious start rule found",
			Details: "Consider having a clear entry point rule like 'expression' or 'program'",
		})
	}
}

func findStartRule(config *DSLConfig) (string, bool) {
	// Common start rule names
	commonStarts := []string{"expression", "expr", "program", "start", "command", "statement"}

	for _, rule := range config.Rules {
		for _, start := range commonStarts {
			if strings.ToLower(rule.Name) == start {
				return rule.Name, true
			}
		}
	}

	// If no common start rule, use the first rule
	if len(config.Rules) > 0 {
		return config.Rules[0].Name, true
	}

	return "", false
}

func createDSLFromConfig(config *DSLConfig) (*dslbuilder.DSL, error) {
	dsl := dslbuilder.New(config.Name)

	// Add tokens
	for name, pattern := range config.Tokens {
		// Determine if it's a keyword token (simplified heuristic)
		if !strings.ContainsAny(pattern, "[\\+*?(){}") {
			dsl.KeywordToken(name, pattern)
		} else {
			dsl.Token(name, pattern)
		}
	}

	// Add rules
	for _, rule := range config.Rules {
		dsl.Rule(rule.Name, rule.Pattern, rule.Action)
	}

	// Add dummy actions
	actions := make(map[string]bool)
	for _, rule := range config.Rules {
		actions[rule.Action] = true
	}
	for action := range actions {
		dsl.Action(action, func(args []interface{}) (interface{}, error) {
			return args, nil
		})
	}

	return dsl, nil
}

func testDSLInput(filename, input string) bool {
	config, err := loadDSLConfig(filename)
	if err != nil {
		return false
	}

	dsl, err := createDSLFromConfig(config)
	if err != nil {
		return false
	}

	_, err = dsl.Parse(input)
	return err == nil
}

func outputText(result ValidationResult, showInfo, verbose bool) {
	if result.Valid {
		fmt.Println("✓ DSL validation passed")
	} else {
		fmt.Println("✗ DSL validation failed")
	}

	if showInfo {
		fmt.Printf("\nDSL Information:\n")
		fmt.Printf("  Name: %s\n", result.Info.Name)
		fmt.Printf("  Tokens: %d\n", result.Info.TokenCount)
		fmt.Printf("  Rules: %d\n", result.Info.RuleCount)
	}

	if len(result.Errors) > 0 {
		fmt.Printf("\nErrors (%d):\n", len(result.Errors))
		for _, err := range result.Errors {
			fmt.Printf("  ✗ [%s] %s\n", err.Type, err.Message)
			if verbose && err.Details != "" {
				fmt.Printf("    Details: %s\n", err.Details)
			}
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Printf("\nWarnings (%d):\n", len(result.Warnings))
		for _, warn := range result.Warnings {
			fmt.Printf("  ⚠ [%s] %s\n", warn.Type, warn.Message)
			if verbose && warn.Details != "" {
				fmt.Printf("    Details: %s\n", warn.Details)
			}
		}
	}

	if verbose && showInfo {
		if len(result.Info.Tokens) > 0 {
			fmt.Printf("\nToken Details:\n")
			for _, token := range result.Info.Tokens {
				status := "✓"
				if !token.Valid {
					status = "✗"
				}
				fmt.Printf("  %s %s: %s (priority: %d)\n", status, token.Name, token.Pattern, token.Priority)
				if token.Error != "" {
					fmt.Printf("    Error: %s\n", token.Error)
				}
			}
		}

		if len(result.Info.Rules) > 0 {
			fmt.Printf("\nRule Details:\n")
			for _, rule := range result.Info.Rules {
				status := "✓"
				if !rule.Valid {
					status = "✗"
				}
				fmt.Printf("  %s %s: %v -> %s\n", status, rule.Name, rule.Pattern, rule.Action)
				if rule.Error != "" {
					fmt.Printf("    Error: %s\n", rule.Error)
				}
			}
		}
	}
}

func outputJSON(result ValidationResult, showInfo bool) {
	output := result
	if !showInfo {
		output.Info = DSLInfo{}
	}

	data, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println(string(data))
}

func outputYAML(result ValidationResult, showInfo bool) {
	output := result
	if !showInfo {
		output.Info = DSLInfo{}
	}

	data, _ := yaml.Marshal(output)
	fmt.Print(string(data))
}
