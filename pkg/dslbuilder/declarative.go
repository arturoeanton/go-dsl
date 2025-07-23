// Package dslbuilder provides declarative configuration support for DSLs.
// This file enables defining DSLs using YAML or JSON configuration files
// instead of programmatic API calls, making it easier to maintain and
// version control grammar definitions.
package dslbuilder

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// DSLConfig represents the declarative configuration for a DSL.
// This structure can be marshaled to/from YAML or JSON for easy
// grammar definition without writing Go code.
//
// Example YAML:
//
//	name: calculator
//	tokens:
//	  NUMBER: "[0-9]+"
//	  PLUS: "\\+"
//	  TIMES: "\\*"
//	rules:
//	  - name: expr
//	    pattern: [expr, PLUS, expr]
//	    action: add
//	  - name: expr
//	    pattern: [NUMBER]
//	    action: number
//	context:
//	  debug: true
type DSLConfig struct {
	Name    string                 `yaml:"name" json:"name"`                           // DSL identifier
	Tokens  map[string]string      `yaml:"tokens" json:"tokens"`                       // Token definitions
	Rules   []RuleConfig           `yaml:"rules" json:"rules"`                         // Grammar rules
	Context map[string]interface{} `yaml:"context,omitempty" json:"context,omitempty"` // Runtime context
}

// RuleConfig represents a rule in the declarative configuration.
// Each rule defines a pattern to match and an action to execute.
//
// Fields:
//   - Name: Rule identifier (can have multiple rules with same name)
//   - Pattern: Sequence of tokens/rules to match
//   - Action: Name of the action function to execute
//
// Multiple RuleConfig entries with the same Name create alternatives.
type RuleConfig struct {
	Name    string   `yaml:"name" json:"name"`       // Rule identifier
	Pattern []string `yaml:"pattern" json:"pattern"` // Symbol sequence
	Action  string   `yaml:"action" json:"action"`   // Action name
}

// LoadFromYAML creates a DSL from a YAML configuration.
// This allows defining grammars in YAML files for better maintainability.
//
// The YAML data should conform to the DSLConfig structure.
// Actions must be registered separately after loading.
//
// Example:
//
//	dsl, err := LoadFromYAML(yamlData)
//	if err != nil {
//	    return err
//	}
//	// Register actions
//	dsl.Action("add", addFunc)
//	dsl.Action("number", numberFunc)
func LoadFromYAML(yamlData []byte) (*DSL, error) {
	var config DSLConfig
	if err := yaml.Unmarshal(yamlData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return createDSLFromConfig(config)
}

// LoadFromYAMLFile creates a DSL from a YAML file.
// Convenience function that reads the file and calls LoadFromYAML.
//
// Example:
//
//	dsl, err := LoadFromYAMLFile("grammar/calculator.yaml")
//	if err != nil {
//	    return err
//	}
func LoadFromYAMLFile(filename string) (*DSL, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}

	return LoadFromYAML(data)
}

// LoadFromJSON creates a DSL from a JSON configuration.
// Similar to LoadFromYAML but uses JSON format.
//
// The JSON data should conform to the DSLConfig structure.
// Actions must be registered separately after loading.
//
// Example JSON:
//
//	{
//	  "name": "calculator",
//	  "tokens": {
//	    "NUMBER": "[0-9]+",
//	    "PLUS": "\\+"
//	  },
//	  "rules": [
//	    {
//	      "name": "expr",
//	      "pattern": ["NUMBER"],
//	      "action": "number"
//	    }
//	  ]
//	}
func LoadFromJSON(jsonData []byte) (*DSL, error) {
	var config DSLConfig
	if err := json.Unmarshal(jsonData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return createDSLFromConfig(config)
}

// LoadFromJSONFile creates a DSL from a JSON file.
// Convenience function that reads the file and calls LoadFromJSON.
//
// Example:
//
//	dsl, err := LoadFromJSONFile("grammar/calculator.json")
//	if err != nil {
//	    return err
//	}
func LoadFromJSONFile(filename string) (*DSL, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON file: %w", err)
	}

	return LoadFromJSON(data)
}

// createDSLFromConfig creates a DSL instance from a configuration.
// This is the core function that transforms declarative configuration
// into a working DSL instance.
//
// Process:
//  1. Create new DSL with the specified name
//  2. Add all tokens (detects keywords automatically)
//  3. Add all rules in order
//  4. Set context values if provided
//
// The function intelligently detects keyword tokens that were saved
// with word boundary patterns and extracts the actual keyword.
func createDSLFromConfig(config DSLConfig) (*DSL, error) {
	// Create DSL instance
	dsl := New(config.Name)

	// Add tokens
	for name, pattern := range config.Tokens {
		// Check if this is a keyword token saved with word boundaries
		if isKeywordTokenPattern(pattern) {
			// Extract the actual keyword from the pattern
			keyword := extractKeywordFromPattern(pattern)
			if err := dsl.KeywordToken(name, keyword); err != nil {
				return nil, fmt.Errorf("failed to add keyword token %s: %w", name, err)
			}
		} else if isKeywordToken(pattern) {
			// If pattern is a simple word without regex, treat as keyword
			if err := dsl.KeywordToken(name, pattern); err != nil {
				return nil, fmt.Errorf("failed to add keyword token %s: %w", name, err)
			}
		} else {
			if err := dsl.Token(name, pattern); err != nil {
				return nil, fmt.Errorf("failed to add token %s: %w", name, err)
			}
		}
	}

	// Add rules
	for _, rule := range config.Rules {
		dsl.Rule(rule.Name, rule.Pattern, rule.Action)
	}

	// Set context
	for key, value := range config.Context {
		dsl.SetContext(key, value)
	}

	return dsl, nil
}

// isKeywordToken checks if a pattern is likely a keyword (simple word without regex).
// Keywords are simple alphanumeric words without regex metacharacters.
//
// A pattern is considered a keyword if:
//   - Contains only letters, digits, underscores, or hyphens
//   - Has at least one letter
//   - No regex special characters
//
// Examples:
//   - "if", "while", "return" -> true (keywords)
//   - "[0-9]+", "\\+", "a*" -> false (regex patterns)
func isKeywordToken(pattern string) bool {
	// Keywords should be simple alphanumeric words
	// Check if the pattern matches a simple word pattern
	for _, r := range pattern {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '_' || r == '-') {
			return false
		}
	}
	// Must have at least one letter to be a keyword
	hasLetter := false
	for _, r := range pattern {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			hasLetter = true
			break
		}
	}
	return hasLetter && len(pattern) > 0
}

// isKeywordTokenPattern checks if the pattern is a keyword token with word boundaries.
// This detects patterns generated by KeywordToken() when saving/loading configs.
//
// Pattern format: (?i)\b<keyword>\b
//   - (?i) = case insensitive
//   - \b = word boundary
//
// This allows proper round-trip conversion of keyword tokens.
func isKeywordTokenPattern(pattern string) bool {
	// Check for the pattern that KeywordToken generates: (?i)\b<word>\b
	return strings.HasPrefix(pattern, "(?i)\\b") && strings.HasSuffix(pattern, "\\b")
}

// extractKeywordFromPattern extracts the keyword from a pattern like (?i)\bword\b.
// Used when loading saved configurations to restore keyword tokens properly.
//
// Examples:
//   - "(?i)\\bif\\b" -> "if"
//   - "(?i)\\breturn\\b" -> "return"
//   - "[0-9]+" -> "[0-9]+" (unchanged)
func extractKeywordFromPattern(pattern string) string {
	// Remove (?i)\b from start and \b from end
	if isKeywordTokenPattern(pattern) {
		pattern = strings.TrimPrefix(pattern, "(?i)\\b")
		pattern = strings.TrimSuffix(pattern, "\\b")
		return pattern
	}
	return pattern
}

// SaveToYAML exports the DSL configuration to YAML format.
// This allows saving a programmatically created DSL for reuse.
//
// The exported YAML can be loaded back with LoadFromYAML.
// Note: Actions are not exported and must be re-registered.
//
// Example:
//
//	yamlData, err := dsl.SaveToYAML()
//	if err != nil {
//	    return err
//	}
//	// Save to file or transmit...
func (d *DSL) SaveToYAML() ([]byte, error) {
	config := d.toConfig()
	return yaml.Marshal(config)
}

// SaveToYAMLFile exports the DSL configuration to a YAML file.
// Convenience method that combines SaveToYAML with file writing.
//
// File is created with 0644 permissions (readable by all, writable by owner).
//
// Example:
//
//	err := dsl.SaveToYAMLFile("grammar/my-dsl.yaml")
//	if err != nil {
//	    return err
//	}
func (d *DSL) SaveToYAMLFile(filename string) error {
	data, err := d.SaveToYAML()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
}

// SaveToJSON exports the DSL configuration to JSON format.
// Similar to SaveToYAML but uses JSON encoding with indentation.
//
// The exported JSON can be loaded back with LoadFromJSON.
// Note: Actions are not exported and must be re-registered.
//
// The output is formatted with 2-space indentation for readability.
func (d *DSL) SaveToJSON() ([]byte, error) {
	config := d.toConfig()
	return json.MarshalIndent(config, "", "  ")
}

// SaveToJSONFile exports the DSL configuration to a JSON file.
// Convenience method that combines SaveToJSON with file writing.
//
// File is created with 0644 permissions (readable by all, writable by owner).
//
// Example:
//
//	err := dsl.SaveToJSONFile("grammar/my-dsl.json")
//	if err != nil {
//	    return err
//	}
func (d *DSL) SaveToJSONFile(filename string) error {
	data, err := d.SaveToJSON()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
}

// toConfig converts a DSL instance to a configuration struct.
// This is the inverse of createDSLFromConfig, enabling round-trip conversion.
//
// The configuration includes:
//   - DSL name
//   - All token definitions with their patterns
//   - All rules with their alternatives
//   - Context variables
//
// Note: Action functions are not included as they cannot be serialized.
// They must be re-registered when loading the configuration.
func (d *DSL) toConfig() DSLConfig {
	config := DSLConfig{
		Name:    d.name,
		Tokens:  make(map[string]string),
		Rules:   []RuleConfig{},
		Context: d.context,
	}

	// Export tokens
	for name, token := range d.grammar.tokens {
		config.Tokens[name] = token.pattern
	}

	// Export rules
	for name, rule := range d.grammar.rules {
		for _, alt := range rule.alternatives {
			config.Rules = append(config.Rules, RuleConfig{
				Name:    name,
				Pattern: alt.sequence,
				Action:  alt.action,
			})
		}
	}

	return config
}
