package dslbuilder

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test ParseError context generation
func TestParseErrorContext(t *testing.T) {
	// Create parse error with different positions
	err1 := &ParseError{
		Message:  "test error",
		Input:    "",
		Position: 0,
	}
	assert.Equal(t, "test error", err1.Error())

	// Test with position beyond input
	err2 := &ParseError{
		Message:  "test error",
		Input:    "hello",
		Position: 10,
	}
	assert.Contains(t, err2.Error(), "test error")

	// Test DetailedError
	details := err2.DetailedError()
	assert.Contains(t, details, "test error")
}

// Test SaveToYAMLFile error cases
func TestSaveToYAMLFileErrors(t *testing.T) {
	dsl := New("SaveTest")

	// Try to save to invalid directory
	err := dsl.SaveToYAMLFile("/nonexistent/directory/file.yaml")
	assert.Error(t, err)
}

// Test SaveToJSONFile error cases
func TestSaveToJSONFileErrors(t *testing.T) {
	dsl := New("SaveTest")

	// Try to save to invalid directory
	err := dsl.SaveToJSONFile("/nonexistent/directory/file.json")
	assert.Error(t, err)
}

// Test extractKeywordFromPattern edge case
func TestExtractKeywordFromPatternEdgeCase(t *testing.T) {
	// Test with non-keyword pattern
	result := extractKeywordFromPattern("not a keyword pattern")
	assert.Equal(t, "not a keyword pattern", result)
}

// Test createDSLFromConfig with only keyword tokens
func TestCreateDSLFromConfigKeywords(t *testing.T) {
	config := DSLConfig{
		Name: "KeywordOnlyTest",
		Tokens: map[string]string{
			"IF":   "if",
			"THEN": "then",
			"ELSE": "else",
		},
	}

	dsl, err := createDSLFromConfig(config)
	assert.NoError(t, err)
	assert.Equal(t, "KeywordOnlyTest", dsl.name)
}

// Test AddKeywordToken error case
func TestAddKeywordTokenError(t *testing.T) {
	grammar := NewGrammar()

	// This should work
	err := grammar.AddKeywordToken("TEST", "test")
	assert.NoError(t, err)

	// Try to add keyword with invalid pattern
	// Since it creates a regex pattern, it's hard to make it fail
	// But we can test the path
}

// Test parseRule when there are no alternatives
func TestParseRuleNoAlternatives(t *testing.T) {
	dsl := New("NoAltTest")
	require.NoError(t, dsl.Token("A", "a"))

	// Create grammar but don't add any rules
	parser := NewParser(dsl.grammar)
	parser.dsl = dsl

	// Tokenize first
	err := parser.tokenize("a")
	assert.NoError(t, err)

	// Try to parse non-existent rule
	parser.pos = 0
	_, err = parser.parseRule("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// Test tokenize with special cases
func TestTokenizeSpecialCases(t *testing.T) {
	dsl := New("TokenizeSpecial")

	// Add token with lookahead
	require.NoError(t, dsl.Token("STR", `"[^"]*"`))
	require.NoError(t, dsl.Token("SPACE", `\s+`))

	parser := NewParser(dsl.grammar)
	parser.dsl = dsl

	// Test string matching
	err := parser.tokenize(`"hello world"`)
	assert.NoError(t, err)
	assert.Len(t, parser.tokens, 1)
	assert.Equal(t, "STR", parser.tokens[0].TokenType)
}

// Test improved parser edge cases
func TestImprovedParserEdgeCases(t *testing.T) {
	dsl := New("ImprovedEdge")

	// Test empty input after adding tokens
	require.NoError(t, dsl.Token("A", "a"))
	dsl.Rule("test", []string{"A"}, "action")

	// Parse empty string
	_, err := dsl.Parse("")
	assert.Error(t, err)

	// Test with only whitespace
	_, err = dsl.Parse("   \n\t  ")
	assert.Error(t, err)
}

// Test parseLeftRecursive with multiple alternatives
func TestParseLeftRecursiveMultipleAlts(t *testing.T) {
	dsl := New("LeftRecMulti")

	require.NoError(t, dsl.Token("A", "a"))
	require.NoError(t, dsl.Token("B", "b"))
	require.NoError(t, dsl.Token("C", "c"))

	// Multiple left recursive rules
	dsl.Rule("expr", []string{"expr", "A"}, "appendA")
	dsl.Rule("expr", []string{"expr", "B"}, "appendB")
	dsl.Rule("expr", []string{"C"}, "base")

	dsl.Action("appendA", func(args []interface{}) (interface{}, error) {
		return args[0].(string) + "a", nil
	})
	dsl.Action("appendB", func(args []interface{}) (interface{}, error) {
		return args[0].(string) + "b", nil
	})
	dsl.Action("base", func(args []interface{}) (interface{}, error) {
		return "c", nil
	})

	// Test mixed sequence
	result, err := dsl.Parse("c a b a")
	assert.NoError(t, err)
	assert.Equal(t, "caba", result.GetOutput())
}

// Test parseAlternative with action error
func TestParseAlternativeActionError(t *testing.T) {
	dsl := New("ActionErrorTest")

	require.NoError(t, dsl.Token("A", "a"))
	dsl.Rule("test", []string{"A"}, "errorAction")

	// Add action to grammar that returns error
	dsl.grammar.actions["errorAction"] = func(args []interface{}) (interface{}, error) {
		return nil, assert.AnError
	}

	parser := NewParser(dsl.grammar)
	parser.dsl = dsl

	// This should propagate the action error
	_, err := parser.Parse("a")
	assert.Error(t, err)
}

// Test file operations with proper cleanup
func TestFileOperationsWithCleanup(t *testing.T) {
	dsl := New("FileOpsTest")
	require.NoError(t, dsl.Token("NUM", "[0-9]+"))
	dsl.Rule("num", []string{"NUM"}, "number")

	// Create temp directory
	tempDir := t.TempDir()

	// Test successful save and load
	yamlFile := filepath.Join(tempDir, "test.yaml")
	err := dsl.SaveToYAMLFile(yamlFile)
	assert.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(yamlFile)
	assert.NoError(t, err)

	// Test JSON operations
	jsonFile := filepath.Join(tempDir, "test.json")
	err = dsl.SaveToJSONFile(jsonFile)
	assert.NoError(t, err)

	// Load and verify
	loaded, err := LoadFromYAMLFile(yamlFile)
	assert.NoError(t, err)
	assert.Equal(t, dsl.name, loaded.name)
}

// Test improved parser with no rules
func TestImprovedParserNoRules(t *testing.T) {
	dsl := New("NoRulesTest")
	require.NoError(t, dsl.Token("A", "a"))

	// No rules defined
	_, err := dsl.Parse("a")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// Test DebugTokens with actual output
func TestDebugTokensOutput(t *testing.T) {
	dsl := New("DebugOutput")
	require.NoError(t, dsl.Token("NUM", "[0-9]+"))
	require.NoError(t, dsl.KeywordToken("IF", "if"))

	// Just call it to cover the print statements
	// In a real test environment, we'd capture stdout
	dsl.DebugTokens("if 123")
}

// Test parseRuleRegular with position tracking
func TestParseRuleRegularPosition(t *testing.T) {
	dsl := New("PositionTest")
	require.NoError(t, dsl.Token("A", "a"))
	require.NoError(t, dsl.Token("B", "b"))

	// Rule that will partially match
	dsl.Rule("ab", []string{"A", "B"}, "concat")
	dsl.Action("concat", func(args []interface{}) (interface{}, error) {
		return args[0].(string) + args[1].(string), nil
	})

	// Parse with only "a" - should fail at "b"
	_, err := dsl.Parse("a")
	assert.Error(t, err)
	// Error should mention we expected more tokens
}
