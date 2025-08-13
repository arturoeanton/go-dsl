package universal

import (
	"fmt"
	"strings"
)

// LoopResult contains the results of processing a loop body
type LoopResult struct {
	Results      []interface{}
	ShouldBreak  bool
	ShouldContinue bool
}

// ProcessLoopBody processes the body of a loop with full support for nested structures
// This function handles break, continue, and nested if blocks correctly
func (hd *HTTPDSLv3) ProcessLoopBody(body []string) (*LoopResult, error) {
	result := &LoopResult{
		Results: []interface{}{},
		ShouldBreak: false,
		ShouldContinue: false,
	}
	
	for i := 0; i < len(body); i++ {
		line := body[i]
		trimmed := strings.TrimSpace(line)
		
		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		
		// Handle break statement
		if trimmed == "break" {
			result.ShouldBreak = true
			return result, nil
		}
		
		// Handle continue statement
		if trimmed == "continue" {
			result.ShouldContinue = true
			return result, nil
		}
		
		// Handle if blocks
		if strings.HasPrefix(trimmed, "if ") && strings.HasSuffix(trimmed, " then") {
			// Extract the complete if block
			ifBlock, endIdx := hd.ExtractIfBlock(body, i)
			if ifBlock == nil {
				return nil, fmt.Errorf("malformed if block at line %d", i+1)
			}
			
			// Process the if block with special handling for break/continue
			ifResult, err := hd.ProcessIfBlockWithControl(ifBlock)
			if err != nil {
				return nil, err
			}
			
			// Append results
			result.Results = append(result.Results, ifResult.Results...)
			
			// Check for break/continue from the if block
			if ifResult.ShouldBreak {
				result.ShouldBreak = true
				return result, nil
			}
			if ifResult.ShouldContinue {
				result.ShouldContinue = true
				return result, nil
			}
			
			// Skip to the end of the if block
			i = endIdx
			continue
		}
		
		// Handle nested loops (while, foreach, repeat)
		if strings.HasPrefix(trimmed, "while ") && strings.HasSuffix(trimmed, " do") ||
		   strings.HasPrefix(trimmed, "foreach ") && strings.Contains(trimmed, " in ") && strings.HasSuffix(trimmed, " do") ||
		   strings.HasPrefix(trimmed, "repeat ") && strings.Contains(trimmed, " times do") {
			// Extract the nested loop block
			loopBlock, endIdx := hd.ExtractLoopBlock(body, i)
			if loopBlock == nil {
				return nil, fmt.Errorf("malformed loop block at line %d", i+1)
			}
			
			// Process the nested loop recursively
			loopResults := []interface{}{}
			for _, loopLine := range loopBlock {
				res, err := hd.ParseWithBlockSupport(loopLine)
				if err != nil {
					return nil, err
				}
				if res != nil && res != "" {
					loopResults = append(loopResults, res)
				}
			}
			
			result.Results = append(result.Results, loopResults...)
			i = endIdx
			continue
		}
		
		// Process regular line
		lineResult, err := hd.ParseWithContext(trimmed)
		if err != nil {
			return nil, fmt.Errorf("error processing line %d: %v", i+1, err)
		}
		if lineResult != nil && lineResult != "" {
			result.Results = append(result.Results, lineResult)
		}
	}
	
	return result, nil
}

// ExtractIfBlock extracts a complete if/endif block from lines starting at index
func (hd *HTTPDSLv3) ExtractIfBlock(lines []string, startIdx int) ([]string, int) {
	if startIdx >= len(lines) {
		return nil, -1
	}
	
	var block []string
	nestLevel := 0
	endIdx := startIdx
	
	for i := startIdx; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		
		// Track nesting
		if strings.HasPrefix(trimmed, "if ") && strings.HasSuffix(trimmed, " then") {
			nestLevel++
		} else if trimmed == "endif" {
			nestLevel--
		}
		
		block = append(block, line)
		
		// Found matching endif
		if nestLevel == 0 {
			endIdx = i
			break
		}
	}
	
	// Check if we found a complete block
	if nestLevel != 0 {
		return nil, -1
	}
	
	return block, endIdx
}

// ExtractLoopBlock extracts a complete loop/endloop block from lines starting at index
func (hd *HTTPDSLv3) ExtractLoopBlock(lines []string, startIdx int) ([]string, int) {
	if startIdx >= len(lines) {
		return nil, -1
	}
	
	var block []string
	nestLevel := 0
	endIdx := startIdx
	
	for i := startIdx; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		
		// Track nesting
		if strings.HasSuffix(trimmed, " do") {
			nestLevel++
		} else if trimmed == "endloop" {
			nestLevel--
		}
		
		block = append(block, line)
		
		// Found matching endloop
		if nestLevel == 0 {
			endIdx = i
			break
		}
	}
	
	// Check if we found a complete block
	if nestLevel != 0 {
		return nil, -1
	}
	
	return block, endIdx
}

// ProcessIfBlock processes a complete if/endif block
func (hd *HTTPDSLv3) ProcessIfBlock(block []string) ([]interface{}, error) {
	if len(block) < 2 {
		return nil, fmt.Errorf("invalid if block: too short")
	}
	
	// Get the first line (if condition)
	firstLine := strings.TrimSpace(block[0])
	if !strings.HasPrefix(firstLine, "if ") || !strings.HasSuffix(firstLine, " then") {
		return nil, fmt.Errorf("invalid if block: missing if/then")
	}
	
	// Process the if block completely using ParseWithBlockSupport
	// This handles the entire if/then/else/endif as a single unit
	results := []interface{}{}
	
	// Join the block lines and process as a single unit
	blockCode := strings.Join(block, "\n")
	result, err := hd.ParseWithBlockSupport(blockCode)
	
	// Check for loop control errors
	if err != nil {
		if loopErr, ok := err.(*LoopControlError); ok {
			// Propagate loop control signals
			return nil, loopErr
		}
		return nil, err
	}
	
	if result != nil && result != "" {
		results = append(results, result)
	}
	
	return results, nil
}

// LoopControlError represents a break or continue signal from within nested structures
type LoopControlError struct {
	IsBreak    bool
	IsContinue bool
}

func (e *LoopControlError) Error() string {
	if e.IsBreak {
		return "break"
	}
	if e.IsContinue {
		return "continue"
	}
	return "loop control"
}

// CreateLoopControlError creates a new loop control error
func CreateLoopControlError(isBreak, isContinue bool) *LoopControlError {
	return &LoopControlError{
		IsBreak:    isBreak,
		IsContinue: isContinue,
	}
}

// ProcessIfBlockWithControl processes an if block and handles break/continue statements
func (hd *HTTPDSLv3) ProcessIfBlockWithControl(block []string) (*LoopResult, error) {
	result := &LoopResult{
		Results: []interface{}{},
		ShouldBreak: false,
		ShouldContinue: false,
	}
	
	if len(block) < 2 {
		return nil, fmt.Errorf("invalid if block: too short")
	}
	
	// Get the first line (if condition)
	firstLine := strings.TrimSpace(block[0])
	if !strings.HasPrefix(firstLine, "if ") || !strings.HasSuffix(firstLine, " then") {
		return nil, fmt.Errorf("invalid if block: missing if/then")
	}
	
	// Extract condition
	conditionStr := strings.TrimSuffix(strings.TrimPrefix(firstLine, "if "), " then")
	
	// Evaluate condition using the new evaluator that supports AND/OR
	shouldExecute := hd.EvaluateCondition(conditionStr)
	
	// Parse the if block to find then/else sections
	var thenBlock []string
	var elseBlock []string
	inElse := false
	nestLevel := 0
	
	for i := 1; i < len(block); i++ {
		line := strings.TrimSpace(block[i])
		
		// Track nesting for nested if blocks
		if strings.HasPrefix(line, "if ") && strings.HasSuffix(line, " then") {
			nestLevel++
			// Add line to appropriate block (include nested if/endif/else)
			if inElse {
				elseBlock = append(elseBlock, line)
			} else {
				thenBlock = append(thenBlock, line)
			}
		} else if line == "endif" {
			if nestLevel == 0 {
				break // End of our if block
			}
			nestLevel--
			// Add line to appropriate block (include nested if/endif/else)
			if inElse {
				elseBlock = append(elseBlock, line)
			} else {
				thenBlock = append(thenBlock, line)
			}
		} else if line == "else" && nestLevel == 0 {
			inElse = true
			continue
		} else if line == "else" && nestLevel > 0 {
			// This else belongs to a nested if
			if inElse {
				elseBlock = append(elseBlock, line)
			} else {
				thenBlock = append(thenBlock, line)
			}
		} else if line != "" && !strings.HasPrefix(line, "#") {
			// Add regular lines to appropriate block
			if inElse {
				elseBlock = append(elseBlock, line)
			} else {
				thenBlock = append(thenBlock, line)
			}
		}
	}
	
	// Execute the appropriate block
	var blockToExecute []string
	if shouldExecute {
		blockToExecute = thenBlock
	} else {
		blockToExecute = elseBlock
	}
	
	// Process each line in the selected block
	for i := 0; i < len(blockToExecute); i++ {
		blockLine := blockToExecute[i]
		trimmed := strings.TrimSpace(blockLine)
		
		// Skip empty lines (already processed)
		if trimmed == "" {
			continue
		}
		
		// Check for break
		if trimmed == "break" {
			result.ShouldBreak = true
			return result, nil
		}
		
		// Check for continue
		if trimmed == "continue" {
			result.ShouldContinue = true
			return result, nil
		}
		
		// Process nested if blocks using ParseWithBlockSupport
		if strings.HasPrefix(trimmed, "if ") && strings.HasSuffix(trimmed, " then") {
			// Find the complete nested if block including all its content
			nestedLines := []string{blockLine}
			nestCount := 1
			
			// Collect all lines of the nested if block
			for j := i + 1; j < len(blockToExecute) && nestCount > 0; j++ {
				nestedLine := strings.TrimSpace(blockToExecute[j])
				nestedLines = append(nestedLines, blockToExecute[j])
				
				if strings.HasPrefix(nestedLine, "if ") && strings.HasSuffix(nestedLine, " then") {
					nestCount++
				} else if nestedLine == "endif" {
					nestCount--
					if nestCount == 0 {
						// Process the complete nested if block using ParseWithBlockSupport
						nestedCode := strings.Join(nestedLines, "\n")
						nestedResult, err := hd.ParseWithBlockSupport(nestedCode)
						if err != nil {
							return nil, err
						}
						if nestedResult != nil {
							// Check if it's a slice of results
							if results, ok := nestedResult.([]interface{}); ok {
								result.Results = append(result.Results, results...)
							} else if nestedResult != "" {
								result.Results = append(result.Results, nestedResult)
							}
						}
						
						// Skip the lines we've processed
						i = j
						break
					}
				}
			}
			
			continue
		}
		
		// Execute normal line
		lineResult, err := hd.ParseWithContext(trimmed)
		if err != nil {
			return nil, fmt.Errorf("error processing line '%s': %v", trimmed, err)
		}
		if lineResult != nil && lineResult != "" {
			result.Results = append(result.Results, lineResult)
		}
	}
	
	return result, nil
}