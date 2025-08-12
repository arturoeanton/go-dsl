// Package dslbuilder - Block parsing support
package dslbuilder

import (
	"strings"
	"fmt"
	"regexp"
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
				processed := bp.processBlock(lines[i:blockEnd+1])
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

// isBlockStart checks if a line starts a block construct
func (bp *BlockParser) isBlockStart(line string) bool {
	// Check for patterns that start blocks
	blockStarters := []string{
		"if .* then$",      // if condition then (at end of line)
		"repeat .* do$",     // repeat N times do
		"while .* do$",      // while condition do
		"foreach .* do$",    // foreach item in list do
	}
	
	for _, pattern := range blockStarters {
		if matched, _ := regexp.MatchString(pattern, line); matched {
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
		if line == endMarker || strings.HasPrefix(line, endMarker + " ") {
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
		
		// Find else clause if present
		elseIndex := -1
		for i, line := range lines {
			if strings.TrimSpace(line) == "else" {
				elseIndex = i
				break
			}
		}
		
		if elseIndex > 0 {
			// if...then...else...endif
			thenStatements := bp.extractStatements(lines[1:elseIndex])
			elseStatements := bp.extractStatements(lines[elseIndex+1:len(lines)-1])
			
			// Convert to inline format with statement grouping
			return fmt.Sprintf("%s (%s) else (%s)", 
				condition, 
				bp.joinStatements(thenStatements),
				bp.joinStatements(elseStatements))
		} else {
			// if...then...endif (no else)
			statements := bp.extractStatements(lines[1:len(lines)-1])
			
			// Convert to inline format
			return fmt.Sprintf("%s (%s)", condition, bp.joinStatements(statements))
		}
	}
	
	// Handle repeat...do...endloop blocks
	if strings.HasPrefix(firstLine, "repeat ") && strings.HasSuffix(firstLine, " do") {
		statements := bp.extractStatements(lines[1:len(lines)-1])
		return fmt.Sprintf("%s (%s) endloop", firstLine, bp.joinStatements(statements))
	}
	
	// Handle while...do...endloop blocks
	if strings.HasPrefix(firstLine, "while ") && strings.HasSuffix(firstLine, " do") {
		statements := bp.extractStatements(lines[1:len(lines)-1])
		return fmt.Sprintf("%s (%s) endloop", firstLine, bp.joinStatements(statements))
	}
	
	// Default: join all lines
	return strings.Join(lines, " ")
}

// extractStatements gets the statements from block body lines
func (bp *BlockParser) extractStatements(lines []string) []string {
	statements := []string{}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			statements = append(statements, trimmed)
		}
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

// ParseMultilineBlocks combines multiline and block parsing
func (d *DSL) ParseMultilineBlocks(code string) (interface{}, error) {
	// For now, just parse each line that's not part of a block structure
	lines := strings.Split(code, "\n")
	var results []interface{}
	i := 0
	
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			i++
			continue
		}
		
		// Check if this starts a block
		if strings.HasSuffix(line, " then") && strings.HasPrefix(line, "if ") {
			// Find the matching endif
			blockLines := []string{line}
			i++
			nestLevel := 1
			
			for i < len(lines) && nestLevel > 0 {
				innerLine := strings.TrimSpace(lines[i])
				if innerLine == "endif" {
					nestLevel--
					if nestLevel == 0 {
						// Process the complete if block as a single unit
						// For now, execute each line in the block
						for _, bl := range blockLines[1:] {
							if bl != "" && bl != "else" {
								r, err := d.Parse(bl)
								if err != nil {
									return results, err
								}
								results = append(results, r)
							}
						}
						i++
						break
					}
				} else if strings.HasPrefix(innerLine, "if ") && strings.HasSuffix(innerLine, " then") {
					nestLevel++
				}
				blockLines = append(blockLines, innerLine)
				i++
			}
		} else {
			// Regular line - parse it
			result, err := d.Parse(line)
			if err != nil {
				return results, err
			}
			results = append(results, result)
			i++
		}
	}
	
	return results, nil
}