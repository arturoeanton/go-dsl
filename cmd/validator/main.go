// Command validator validates DSL grammar definitions (YAML/JSON).
//
// The actual analysis lives in the dslbuilder package (DSL.Validate), so
// programmatic users get exactly the same checks:
//
//	warnings, err := dsl.Validate()
//
// This command loads a declarative configuration, runs the core validator,
// and renders the result as text, JSON, or YAML.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
	"gopkg.in/yaml.v3"
)

type ValidationResult struct {
	Valid    bool
	Errors   []ValidationIssue
	Warnings []ValidationIssue
	Info     DSLInfo
}

type ValidationIssue struct {
	Message string
}

type DSLInfo struct {
	Name       string
	TokenCount int
	RuleCount  int
	Tokens     map[string]string
	Rules      []dslbuilder.RuleConfig
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
	flag.BoolVar(&strictMode, "strict", false, "Treat warnings as errors")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "DSL Validator - Validate DSL grammar and detect potential issues\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if dslFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	result := validateDSLFile(dslFile, testInput, strictMode)

	switch format {
	case "json":
		outputJSON(result, showInfo)
	case "yaml":
		outputYAML(result, showInfo)
	default:
		outputText(result, showInfo, verbose)
	}

	if !result.Valid {
		os.Exit(1)
	}
}

// validateDSLFile loads the declarative configuration and delegates
// grammar analysis to the core validator (dslbuilder.DSL.Validate).
func validateDSLFile(filename, testInput string, strict bool) ValidationResult {
	result := ValidationResult{Valid: true}

	data, err := os.ReadFile(filename)
	if err != nil {
		return failure(result, fmt.Sprintf("Failed to read file: %v", err))
	}

	var config dslbuilder.DSLConfig
	var dsl *dslbuilder.DSL

	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(filename)), ".")
	switch ext {
	case "yaml", "yml":
		if err := yaml.Unmarshal(data, &config); err != nil {
			return failure(result, fmt.Sprintf("Invalid YAML: %v", err))
		}
		dsl, err = dslbuilder.LoadFromYAML(data)
	case "json":
		if err := json.Unmarshal(data, &config); err != nil {
			return failure(result, fmt.Sprintf("Invalid JSON: %v", err))
		}
		dsl, err = dslbuilder.LoadFromJSON(data)
	default:
		if ext == "" {
			return failure(result, fmt.Sprintf("Missing file extension in %q (expected .yaml, .yml, or .json)", filename))
		}
		return failure(result, fmt.Sprintf("Unsupported file format: %s", ext))
	}

	tokenCount := len(config.Tokens)
	if len(config.TokenDefs) > tokenCount {
		tokenCount = len(config.TokenDefs)
	}
	result.Info = DSLInfo{
		Name:       config.Name,
		TokenCount: tokenCount,
		RuleCount:  len(config.Rules) + len(config.Expressions),
		Tokens:     config.Tokens,
		Rules:      config.Rules,
	}

	if err != nil {
		return failure(result, fmt.Sprintf("Failed to create DSL: %v", err))
	}

	warnings, err := dsl.Validate()
	if err != nil {
		result.Valid = false
		for _, line := range strings.Split(err.Error(), "\n") {
			if line != "" {
				result.Errors = append(result.Errors, ValidationIssue{Message: line})
			}
		}
	}
	for _, w := range warnings {
		result.Warnings = append(result.Warnings, ValidationIssue{Message: w})
	}

	if strict && len(result.Warnings) > 0 {
		result.Valid = false
	}

	if testInput != "" {
		if _, err := dsl.ParseAST(testInput); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationIssue{
				Message: fmt.Sprintf("Test input failed to parse: %v", dslbuilder.GetDetailedError(err)),
			})
		}
	}

	return result
}

func failure(result ValidationResult, message string) ValidationResult {
	result.Valid = false
	result.Errors = append(result.Errors, ValidationIssue{Message: message})
	return result
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
		for _, e := range result.Errors {
			fmt.Printf("  ✗ %s\n", e.Message)
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Printf("\nWarnings (%d):\n", len(result.Warnings))
		for _, w := range result.Warnings {
			fmt.Printf("  ⚠ %s\n", w.Message)
		}
	}

	if verbose && showInfo {
		if len(result.Info.Tokens) > 0 {
			fmt.Printf("\nToken Details:\n")
			for name, pattern := range result.Info.Tokens {
				fmt.Printf("  %s: %s\n", name, pattern)
			}
		}
		if len(result.Info.Rules) > 0 {
			fmt.Printf("\nRule Details:\n")
			for _, rule := range result.Info.Rules {
				fmt.Printf("  %s: %v -> %s\n", rule.Name, rule.Pattern, rule.Action)
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
