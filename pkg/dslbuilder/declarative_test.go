package dslbuilder

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test YAML loading and saving
func TestYAMLOperations(t *testing.T) {
	// Create a DSL to save
	originalDSL := New("YAMLTest")
	require.NoError(t, originalDSL.Token("NUM", "[0-9]+"))
	require.NoError(t, originalDSL.Token("PLUS", "\\+"))
	require.NoError(t, originalDSL.KeywordToken("IF", "if"))

	originalDSL.Rule("expr", []string{"NUM"}, "number")
	originalDSL.Rule("expr", []string{"expr", "PLUS", "NUM"}, "add")

	// Test SaveToYAML
	yamlData, err := originalDSL.SaveToYAML()
	assert.NoError(t, err)
	assert.Contains(t, string(yamlData), "name: YAMLTest")
	assert.Contains(t, string(yamlData), "NUM")
	assert.Contains(t, string(yamlData), "PLUS")

	// Test LoadFromYAML
	loadedDSL, err := LoadFromYAML(yamlData)
	assert.NoError(t, err)
	assert.Equal(t, "YAMLTest", loadedDSL.name)

	// Add actions to loaded DSL
	loadedDSL.Action("number", func(args []interface{}) (interface{}, error) {
		return args[0], nil
	})
	loadedDSL.Action("add", func(args []interface{}) (interface{}, error) {
		return fmt.Sprintf("%v+%v", args[0], args[2]), nil
	})

	// Test parsing with loaded DSL
	result, err := loadedDSL.Parse("123 + 456")
	if assert.NoError(t, err) && assert.NotNil(t, result) {
		assert.Equal(t, "123+456", result.GetOutput())
	}

	// Test file operations
	tempDir := t.TempDir()
	yamlFile := filepath.Join(tempDir, "test.yaml")

	// Test SaveToYAMLFile
	err = originalDSL.SaveToYAMLFile(yamlFile)
	assert.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(yamlFile)
	assert.NoError(t, err)

	// Test LoadFromYAMLFile
	fileDSL, err := LoadFromYAMLFile(yamlFile)
	assert.NoError(t, err)
	assert.Equal(t, "YAMLTest", fileDSL.name)
}

// Test JSON loading and saving
func TestJSONOperations(t *testing.T) {
	// Create a DSL to save
	originalDSL := New("JSONTest")
	require.NoError(t, originalDSL.Token("ID", "[a-zA-Z]+"))
	require.NoError(t, originalDSL.Token("ASSIGN", "\\="))
	require.NoError(t, originalDSL.Token("NUM", "[0-9]+"))
	require.NoError(t, originalDSL.KeywordToken("LET", "let"))

	originalDSL.Rule("assign", []string{"LET", "ID", "ASSIGN", "NUM"}, "assignment")

	// Test SaveToJSON
	jsonData, err := originalDSL.SaveToJSON()
	assert.NoError(t, err)

	// Verify JSON structure
	var config DSLConfig
	err = json.Unmarshal(jsonData, &config)
	assert.NoError(t, err)
	assert.Equal(t, "JSONTest", config.Name)
	assert.Len(t, config.Tokens, 4)
	assert.Len(t, config.Rules, 1)

	// Test LoadFromJSON
	loadedDSL, err := LoadFromJSON(jsonData)
	assert.NoError(t, err)
	assert.Equal(t, "JSONTest", loadedDSL.name)

	// Add action to loaded DSL
	loadedDSL.Action("assignment", func(args []interface{}) (interface{}, error) {
		return fmt.Sprintf("let %s = %s", args[1], args[3]), nil
	})

	// Test parsing with loaded DSL
	result, err := loadedDSL.Parse("let x = 42")
	if assert.NoError(t, err) && assert.NotNil(t, result) {
		assert.Equal(t, "let x = 42", result.GetOutput())
	}

	// Test file operations
	tempDir := t.TempDir()
	jsonFile := filepath.Join(tempDir, "test.json")

	// Test SaveToJSONFile
	err = originalDSL.SaveToJSONFile(jsonFile)
	assert.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(jsonFile)
	assert.NoError(t, err)

	// Test LoadFromJSONFile
	fileDSL, err := LoadFromJSONFile(jsonFile)
	assert.NoError(t, err)
	assert.Equal(t, "JSONTest", fileDSL.name)
}

// Test complex declarative configurations
func TestComplexDeclarativeConfig(t *testing.T) {
	config := DSLConfig{
		Name: "ComplexDSL",
		Tokens: map[string]string{
			"NUM":    "[0-9]+",
			"ID":     "[a-zA-Z_][a-zA-Z0-9_]*",
			"IF":     "if",
			"ELSE":   "else",
			"STRING": "\"[^\"]*\"",
			"PLUS":   "\\+",
			"TIMES":  "\\*",
		},
		Rules: []RuleConfig{
			{
				Name:    "stmt",
				Pattern: []string{"IF", "expr", "block"},
				Action:  "ifStmt",
			},
			{
				Name:    "expr",
				Pattern: []string{"expr", "PLUS", "expr"},
				Action:  "add",
			},
			{
				Name:    "expr",
				Pattern: []string{"expr", "TIMES", "expr"},
				Action:  "mul",
			},
			{
				Name:    "expr",
				Pattern: []string{"NUM"},
				Action:  "number",
			},
		},
		Context: map[string]interface{}{
			"debug": true,
			"level": 5,
		},
	}

	// Convert to JSON and load
	jsonData, err := json.Marshal(config)
	assert.NoError(t, err)

	dsl, err := LoadFromJSON(jsonData)
	assert.NoError(t, err)
	assert.Equal(t, "ComplexDSL", dsl.name)

	// Verify context was loaded
	assert.Equal(t, true, dsl.GetContext("debug"))
	assert.Equal(t, float64(5), dsl.GetContext("level")) // JSON numbers are float64
}

// Test isKeywordToken helper function
func TestIsKeywordToken(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected bool
	}{
		{"Keyword if", "if", true},
		{"Keyword else", "else", true},
		{"Simple word", "hello", true},
		{"With underscore", "test_keyword", true},
		{"With hyphen", "test-keyword", true},
		{"Regex pattern", "[a-zA-Z]+", false},
		{"Number pattern", "[0-9]+", false},
		{"With star", "test*", false},
		{"With plus", "test+", false},
		{"With question", "test?", false},
		{"With dot", "test.word", false},
		{"With caret", "^test", false},
		{"With dollar", "test$", false},
		{"With pipe", "if|then", false},
		{"With backslash", "test\\n", false},
		{"With parentheses", "(test)", false},
		{"With curly braces", "{test}", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isKeywordToken(tt.pattern)
			assert.Equal(t, tt.expected, result, "Pattern: %s", tt.pattern)
		})
	}
}

// Test error handling in declarative loading
func TestDeclarativeErrorHandling(t *testing.T) {
	// Test invalid YAML
	_, err := LoadFromYAML([]byte("invalid: yaml: content:"))
	assert.Error(t, err)

	// Test invalid JSON
	_, err = LoadFromJSON([]byte("{invalid json"))
	assert.Error(t, err)

	// Test non-existent file
	_, err = LoadFromYAMLFile("/non/existent/file.yaml")
	assert.Error(t, err)

	_, err = LoadFromJSONFile("/non/existent/file.json")
	assert.Error(t, err)

	// Test invalid token pattern in config
	config := DSLConfig{
		Name: "ErrorDSL",
		Tokens: map[string]string{
			"BAD": "[", // Invalid regex
		},
	}

	jsonData, err := json.Marshal(config)
	assert.NoError(t, err)

	_, err = LoadFromJSON(jsonData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add token")
}

// Test toConfig method
func TestToConfig(t *testing.T) {
	dsl := New("ConfigTest")

	// Add various token types
	require.NoError(t, dsl.Token("NUM", "[0-9]+"))
	require.NoError(t, dsl.KeywordToken("IF", "if"))
	require.NoError(t, dsl.Token("STRING", "\"[^\"]*\""))

	// Add various rules
	dsl.Rule("expr", []string{"NUM"}, "number")
	dsl.Rule("expr", []string{"expr", "+", "expr"}, "add")
	dsl.Rule("stmt", []string{"IF", "expr"}, "ifStmt")

	// Set context
	dsl.SetContext("version", "1.0")
	dsl.SetContext("debug", false)

	// Convert to config
	config := dsl.toConfig()

	assert.Equal(t, "ConfigTest", config.Name)
	assert.Len(t, config.Tokens, 3)
	assert.Len(t, config.Rules, 3)

	// Verify tokens
	assert.Equal(t, "[0-9]+", config.Tokens["NUM"])
	// KeywordToken adds word boundaries
	assert.Contains(t, config.Tokens["IF"], "if")
	assert.Equal(t, "\"[^\"]*\"", config.Tokens["STRING"])

	// Verify context
	assert.Equal(t, "1.0", config.Context["version"])
	assert.Equal(t, false, config.Context["debug"])
}

// Test round-trip conversion (DSL -> Config -> DSL)
func TestRoundTripConversion(t *testing.T) {
	// Create original DSL
	original := New("RoundTripTest")

	// Add configuration
	require.NoError(t, original.Token("NUM", "[0-9]+"))
	require.NoError(t, original.Token("ID", "[a-zA-Z]+"))
	require.NoError(t, original.KeywordToken("FOR", "for"))
	require.NoError(t, original.KeywordToken("IN", "in"))

	original.Rule("loop", []string{"FOR", "ID", "IN", "list"}, "forLoop")
	original.Rule("list", []string{"ID"}, "idList")

	original.SetContext("lang", "en")

	// Save to YAML
	yamlData, err := original.SaveToYAML()
	assert.NoError(t, err)

	// Load from YAML
	loaded, err := LoadFromYAML(yamlData)
	assert.NoError(t, err)

	// Compare names
	assert.Equal(t, original.name, loaded.name)

	// Add actions to loaded DSL to test parsing
	loaded.Action("forLoop", func(args []interface{}) (interface{}, error) {
		return fmt.Sprintf("for %s in %s", args[1], args[3]), nil
	})
	loaded.Action("idList", func(args []interface{}) (interface{}, error) {
		return args[0], nil
	})

	// Test parsing works
	result, err := loaded.Parse("for x in y")
	if assert.NoError(t, err) && assert.NotNil(t, result) {
		assert.Equal(t, "for x in y", result.GetOutput())
	}

	// Save loaded DSL to JSON
	jsonData, err := loaded.SaveToJSON()
	assert.NoError(t, err)

	// Load from JSON
	loaded2, err := LoadFromJSON(jsonData)
	assert.NoError(t, err)

	// Final comparison
	assert.Equal(t, original.name, loaded2.name)
}

// Test empty and minimal configurations
func TestMinimalConfigurations(t *testing.T) {
	// Test empty config
	emptyConfig := DSLConfig{Name: "Empty"}
	jsonData, err := json.Marshal(emptyConfig)
	assert.NoError(t, err)

	dsl, err := LoadFromJSON(jsonData)
	assert.NoError(t, err)
	assert.Equal(t, "Empty", dsl.name)

	// Test minimal config with one token
	minConfig := DSLConfig{
		Name: "Minimal",
		Tokens: map[string]string{
			"A": "a",
		},
	}

	jsonData, err = json.Marshal(minConfig)
	assert.NoError(t, err)

	dsl, err = LoadFromJSON(jsonData)
	assert.NoError(t, err)
	assert.Equal(t, "Minimal", dsl.name)
}

// Test SetContext method from DSL
func TestDSLSetContext(t *testing.T) {
	dsl := New("ContextDSL")

	// Test SetContext with string key
	dsl.SetContext("key1", "value1")
	val := dsl.GetContext("key1")
	assert.Equal(t, "value1", val)

	// Test adding more context values
	dsl.SetContext("key2", "value2")
	dsl.SetContext("key3", 123)

	// All context values should be present
	assert.Equal(t, "value1", dsl.GetContext("key1"))
	assert.Equal(t, "value2", dsl.GetContext("key2"))
	assert.Equal(t, 123, dsl.GetContext("key3"))

	// Test nil value
	assert.Nil(t, dsl.GetContext("nonexistent"))
}

// Test edge cases for declarative loading
func TestDeclarativeEdgeCases(t *testing.T) {
	// Test loading with nil/empty context
	config := DSLConfig{
		Name: "NoContext",
		Tokens: map[string]string{
			"A": "a",
		},
		Context: nil,
	}

	jsonData, err := json.Marshal(config)
	assert.NoError(t, err)

	dsl, err := LoadFromJSON(jsonData)
	assert.NoError(t, err)
	assert.NotNil(t, dsl)

	// Test with special characters in token names
	config2 := DSLConfig{
		Name: "SpecialChars",
		Tokens: map[string]string{
			"TOKEN_1": "token1",
			"TOKEN-2": "token2",
			"TOKEN.3": "token3",
		},
	}

	jsonData, err = json.Marshal(config2)
	assert.NoError(t, err)

	dsl, err = LoadFromJSON(jsonData)
	assert.NoError(t, err)
	assert.NotNil(t, dsl)
}
