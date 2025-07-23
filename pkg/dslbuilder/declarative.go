package dslbuilder

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// DSLConfig represents the declarative configuration for a DSL
type DSLConfig struct {
	Name    string                 `yaml:"name" json:"name"`
	Tokens  map[string]string      `yaml:"tokens" json:"tokens"`
	Rules   []RuleConfig           `yaml:"rules" json:"rules"`
	Context map[string]interface{} `yaml:"context,omitempty" json:"context,omitempty"`
}

// RuleConfig represents a rule in the declarative configuration
type RuleConfig struct {
	Name    string   `yaml:"name" json:"name"`
	Pattern []string `yaml:"pattern" json:"pattern"`
	Action  string   `yaml:"action" json:"action"`
}

// LoadFromYAML creates a DSL from a YAML configuration
func LoadFromYAML(yamlData []byte) (*DSL, error) {
	var config DSLConfig
	if err := yaml.Unmarshal(yamlData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return createDSLFromConfig(config)
}

// LoadFromYAMLFile creates a DSL from a YAML file
func LoadFromYAMLFile(filename string) (*DSL, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}

	return LoadFromYAML(data)
}

// LoadFromJSON creates a DSL from a JSON configuration
func LoadFromJSON(jsonData []byte) (*DSL, error) {
	var config DSLConfig
	if err := json.Unmarshal(jsonData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return createDSLFromConfig(config)
}

// LoadFromJSONFile creates a DSL from a JSON file
func LoadFromJSONFile(filename string) (*DSL, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON file: %w", err)
	}

	return LoadFromJSON(data)
}

// createDSLFromConfig creates a DSL instance from a configuration
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

// isKeywordToken checks if a pattern is likely a keyword (simple word without regex)
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

// isKeywordTokenPattern checks if the pattern is a keyword token with word boundaries
func isKeywordTokenPattern(pattern string) bool {
	// Check for the pattern that KeywordToken generates: (?i)\b<word>\b
	return strings.HasPrefix(pattern, "(?i)\\b") && strings.HasSuffix(pattern, "\\b")
}

// extractKeywordFromPattern extracts the keyword from a pattern like (?i)\bword\b
func extractKeywordFromPattern(pattern string) string {
	// Remove (?i)\b from start and \b from end
	if isKeywordTokenPattern(pattern) {
		pattern = strings.TrimPrefix(pattern, "(?i)\\b")
		pattern = strings.TrimSuffix(pattern, "\\b")
		return pattern
	}
	return pattern
}

// SaveToYAML exports the DSL configuration to YAML format
func (d *DSL) SaveToYAML() ([]byte, error) {
	config := d.toConfig()
	return yaml.Marshal(config)
}

// SaveToYAMLFile exports the DSL configuration to a YAML file
func (d *DSL) SaveToYAMLFile(filename string) error {
	data, err := d.SaveToYAML()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
}

// SaveToJSON exports the DSL configuration to JSON format
func (d *DSL) SaveToJSON() ([]byte, error) {
	config := d.toConfig()
	return json.MarshalIndent(config, "", "  ")
}

// SaveToJSONFile exports the DSL configuration to a JSON file
func (d *DSL) SaveToJSONFile(filename string) error {
	data, err := d.SaveToJSON()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
}

// toConfig converts a DSL instance to a configuration struct
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
