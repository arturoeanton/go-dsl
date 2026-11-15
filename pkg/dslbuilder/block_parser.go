// Package dslbuilder - Block parsing support
package dslbuilder

import (
	"fmt"
	"regexp"
	"strings"
)

// BlockParser handles parsing of multi-line block constructs
type BlockParser struct {
	dsl *DSL
}

// NewBlockParser creates a new block parser
func NewBlockParser(dsl *DSL) *BlockParser {
	return &BlockParser{dsl: dsl}
}

// ParseWithBlocks processes input that may contain block structures
// It preprocesses block constructs into a format the regular parser can handle
func (d *DSL) ParseWithBlocks(code string) (interface{}, error) {
	bp := NewBlockParser(d)
	processed := bp.ProcessBlocks(code)
	return d.Parse(processed)
}

// ProcessBlocks identifies and transforms block structures
func (bp *BlockParser) ProcessBlocks(code string) string {
	lines := strings.Split(code, "\n")
	result := []string{}
	i := 0

	for i < len(lines) {
		line := strings.TrimSpace(lines[i])

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			i++
			continue
		}

		// Check for block start patterns
		if bp.isBlockStart(line) {
			// Process the entire block
			blockEnd := bp.findBlockEnd(lines, i)
			if blockEnd > i {
				processed := bp.processBlock(lines[i : blockEnd+1])
				result = append(result, processed)
				i = blockEnd + 1
				continue
			}
		}

		// Regular line
		result = append(result, line)
		i++
	}

	return strings.Join(result, "\n")
}

var blockStarterRegexes = []*regexp.Regexp{
	regexp.MustCompile(`^if .* then$`),
	regexp.MustCompile(`^repeat .* do$`),
	regexp.MustCompile(`^while .* do$`),
	regexp.MustCompile(`^foreach .* do$`),
}

// isBlockStart checks if a line starts a block construct
func (bp *BlockParser) isBlockStart(line string) bool {
	for _, re := range blockStarterRegexes {
		if re.MatchString(line) {
			return true
		}
	}
	return false
}

// findBlockEnd finds the matching end statement for a block
func (bp *BlockParser) findBlockEnd(lines []string, start int) int {
	startLine := strings.TrimSpace(lines[start])

	// Determine what end marker to look for
	var endMarker string
	if strings.HasPrefix(startLine, "if ") {
		endMarker = "endif"
	} else if strings.HasPrefix(startLine, "repeat ") ||
		strings.HasPrefix(startLine, "while ") ||
		strings.HasPrefix(startLine, "foreach ") {
		endMarker = "endloop"
	} else {
		return start // No block end needed
	}

	// Look for the end marker
	nestLevel := 1
	for i := start + 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// Check for nested blocks
		if bp.isBlockStart(line) {
			nestLevel++
		}

		// Check for end marker
		if line == endMarker || strings.HasPrefix(line, endMarker+" ") {
			nestLevel--
			if nestLevel == 0 {
				return i
			}
		}

		// Also check for 'else' in if blocks
		if endMarker == "endif" && nestLevel == 1 && line == "else" {
			// Continue looking for endif
		}
	}

	return start // No matching end found
}

// processBlock converts a block into inline format
func (bp *BlockParser) processBlock(lines []string) string {
	if len(lines) < 2 {
		return strings.Join(lines, " ")
	}

	firstLine := strings.TrimSpace(lines[0])

	// Handle if...then...endif blocks
	if strings.HasPrefix(firstLine, "if ") && strings.HasSuffix(firstLine, " then") {
		condition := firstLine

		// Find top-level else clause if present (respecting nested blocks)
		elseIndex := -1
		nestLevel := 1
		for i := 1; i < len(lines)-1; i++ {
			trimmed := strings.TrimSpace(lines[i])
			if bp.isBlockStart(trimmed) {
				nestLevel++
				continue
			}
			if trimmed == "endif" || strings.HasPrefix(trimmed, "endif ") ||
				trimmed == "endloop" || strings.HasPrefix(trimmed, "endloop ") {
				if nestLevel > 1 {
					nestLevel--
				}
				continue
			}
			if nestLevel == 1 && trimmed == "else" {
				elseIndex = i
				break
			}
		}

		if elseIndex > 0 {
			// if...then...else...endif
			thenStatements := bp.extractStatements(lines[1:elseIndex])
			elseStatements := bp.extractStatements(lines[elseIndex+1 : len(lines)-1])

			// Convert to inline format with statement grouping
			return fmt.Sprintf("%s (%s) else (%s)",
				condition,
				bp.joinStatements(thenStatements),
				bp.joinStatements(elseStatements))
		} else {
			// if...then...endif (no else)
			statements := bp.extractStatements(lines[1 : len(lines)-1])

			// Convert to inline format
			return fmt.Sprintf("%s (%s)", condition, bp.joinStatements(statements))
		}
	}

	// Handle repeat/while/foreach...do...endloop blocks
	if (strings.HasPrefix(firstLine, "repeat ") ||
		strings.HasPrefix(firstLine, "while ") ||
		strings.HasPrefix(firstLine, "foreach ")) && strings.HasSuffix(firstLine, " do") {
		statements := bp.extractStatements(lines[1 : len(lines)-1])
		return fmt.Sprintf("%s (%s) endloop", firstLine, bp.joinStatements(statements))
	}

	// Default: join all lines
	return strings.Join(lines, " ")
}

// extractStatements gets the statements from block body lines, recursively
// collapsing nested block constructs.
func (bp *BlockParser) extractStatements(lines []string) []string {
	statements := []string{}
	i := 0
	for i < len(lines) {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			i++
			continue
		}
		if bp.isBlockStart(trimmed) {
			blockEnd := bp.findBlockEnd(lines, i)
			if blockEnd > i {
				statements = append(statements, bp.processBlock(lines[i:blockEnd+1]))
				i = blockEnd + 1
				continue
			}
		}
		statements = append(statements, trimmed)
		i++
	}
	return statements
}

// joinStatements combines multiple statements for sequential execution
func (bp *BlockParser) joinStatements(statements []string) string {
	if len(statements) == 0 {
		return ""
	}
	if len(statements) == 1 {
		return statements[0]
	}
	// For multiple statements, we need a different approach
	// Instead of semicolon, we'll create a compound statement
	return "begin " + strings.Join(statements, " and ") + " end"
}

// ParseMultilineBlocks combines multiline and block parsing, transforming
// multi-line block constructs into single inline statements before parsing
// so conditional branches are never executed unconditionally.
func (d *DSL) ParseMultilineBlocks(code string) (interface{}, error) {
	bp := NewBlockParser(d)
	lines := strings.Split(code, "\n")
	var results []interface{}
	i := 0

	for i < len(lines) {
		line := strings.TrimSpace(lines[i])

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			i++
			continue
		}

		// Check if this starts a block
		if bp.isBlockStart(line) {
			blockEnd := bp.findBlockEnd(lines, i)
			if blockEnd > i {
				processed := bp.processBlock(lines[i : blockEnd+1])
				result, err := d.Parse(processed)
				if err != nil {
					return results, err
				}
				results = append(results, result)
				i = blockEnd + 1
				continue
			}
		}

		// Regular line - parse it
		result, err := d.Parse(line)
		if err != nil {
			return results, err
		}
		results = append(results, result)
		i++
	}

	return results, nil
}
