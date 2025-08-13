package universal

import (
	"fmt"
	"strings"
)

// EvaluateCondition evaluates a condition string that may contain AND/OR operators
func (hd *HTTPDSLv3) EvaluateCondition(conditionStr string) bool {
	// Handle OR operator (lower precedence)
	if strings.Contains(conditionStr, " OR ") {
		parts := strings.Split(conditionStr, " OR ")
		for _, part := range parts {
			if hd.EvaluateCondition(strings.TrimSpace(part)) {
				return true
			}
		}
		return false
	}
	
	// Handle AND operator (higher precedence)
	if strings.Contains(conditionStr, " AND ") {
		parts := strings.Split(conditionStr, " AND ")
		for _, part := range parts {
			if !hd.EvaluateCondition(strings.TrimSpace(part)) {
				return false
			}
		}
		return true
	}
	
	// Evaluate simple condition
	return hd.EvaluateSimpleCondition(conditionStr)
}

// EvaluateSimpleCondition evaluates a simple condition without AND/OR
func (hd *HTTPDSLv3) EvaluateSimpleCondition(conditionStr string) bool {
	// Parse the condition (e.g., "$x > 3" or "$status == 200")
	parts := strings.Fields(conditionStr)
	
	// Handle single variable check (e.g., "if $var then")
	if len(parts) == 1 {
		varName := strings.TrimPrefix(parts[0], "$")
		if val, ok := hd.variables[varName]; ok {
			// Check if variable exists and is truthy
			switch v := val.(type) {
			case bool:
				return v
			case int:
				return v != 0
			case float64:
				return v != 0
			case string:
				return v != "" && v != "0" && v != "false"
			default:
				return val != nil
			}
		}
		return false
	}
	
	// Handle comparison (e.g., "$x > 3")
	if len(parts) != 3 {
		return false
	}
	
	leftSide := parts[0]
	operator := parts[1]
	rightSide := parts[2]
	
	// Get left value
	var leftVal interface{}
	if strings.HasPrefix(leftSide, "$") {
		varName := strings.TrimPrefix(leftSide, "$")
		if val, ok := hd.variables[varName]; ok {
			leftVal = val
		} else {
			return false
		}
	} else {
		leftVal = leftSide
	}
	
	// Get right value
	var rightVal interface{}
	if strings.HasPrefix(rightSide, "$") {
		varName := strings.TrimPrefix(rightSide, "$")
		if val, ok := hd.variables[varName]; ok {
			rightVal = val
		} else {
			return false
		}
	} else {
		rightVal = rightSide
	}
	
	// Perform comparison
	return hd.CompareValues(leftVal, operator, rightVal)
}

// CompareValues compares two values with an operator
func (hd *HTTPDSLv3) CompareValues(left interface{}, operator string, right interface{}) bool {
	// Try numeric comparison first
	var leftNum, rightNum float64
	var leftIsNum, rightIsNum bool
	
	// Convert left to number
	switch v := left.(type) {
	case int:
		leftNum = float64(v)
		leftIsNum = true
	case float64:
		leftNum = v
		leftIsNum = true
	case string:
		if _, err := fmt.Sscanf(v, "%f", &leftNum); err == nil {
			leftIsNum = true
		}
	}
	
	// Convert right to number
	switch v := right.(type) {
	case int:
		rightNum = float64(v)
		rightIsNum = true
	case float64:
		rightNum = v
		rightIsNum = true
	case string:
		if _, err := fmt.Sscanf(v, "%f", &rightNum); err == nil {
			rightIsNum = true
		}
	}
	
	// If both are numbers, do numeric comparison
	if leftIsNum && rightIsNum {
		switch operator {
		case ">":
			return leftNum > rightNum
		case "<":
			return leftNum < rightNum
		case ">=":
			return leftNum >= rightNum
		case "<=":
			return leftNum <= rightNum
		case "==":
			return leftNum == rightNum
		case "!=":
			return leftNum != rightNum
		}
	}
	
	// Otherwise do string comparison
	leftStr := fmt.Sprintf("%v", left)
	rightStr := fmt.Sprintf("%v", right)
	
	switch operator {
	case "==":
		return leftStr == rightStr
	case "!=":
		return leftStr != rightStr
	case ">":
		return leftStr > rightStr
	case "<":
		return leftStr < rightStr
	case ">=":
		return leftStr >= rightStr
	case "<=":
		return leftStr <= rightStr
	}
	
	return false
}