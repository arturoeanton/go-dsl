package universal

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

// SCIMFilterDSL represents a SCIM (System for Cross-domain Identity Management) filter DSL
type SCIMFilterDSL struct {
	dsl    *dslbuilder.DSL
	engine *SCIMEngine
}

// NewSCIMFilterDSL creates a new SCIM filter DSL instance
func NewSCIMFilterDSL() *SCIMFilterDSL {
	scim := &SCIMFilterDSL{
		dsl:    dslbuilder.New("SCIMFilter"),
		engine: NewSCIMEngine(),
	}

	scim.setupTokens()
	scim.setupRules()
	scim.setupActions()

	return scim
}

// setupTokens defines the SCIM filter tokens
func (s *SCIMFilterDSL) setupTokens() {
	// Comparison operators (keywords with high priority)
	s.dsl.KeywordToken("EQ", "eq")
	s.dsl.KeywordToken("NE", "ne")
	s.dsl.KeywordToken("CO", "co")
	s.dsl.KeywordToken("SW", "sw")
	s.dsl.KeywordToken("EW", "ew")
	s.dsl.KeywordToken("GT", "gt")
	s.dsl.KeywordToken("GE", "ge")
	s.dsl.KeywordToken("LT", "lt")
	s.dsl.KeywordToken("LE", "le")
	s.dsl.KeywordToken("PR", "pr")

	// Logical operators (keywords with high priority)
	s.dsl.KeywordToken("AND", "and")
	s.dsl.KeywordToken("OR", "or")
	s.dsl.KeywordToken("NOT", "not")

	// Structural tokens
	s.dsl.Token("LPAREN", "\\(")
	s.dsl.Token("RPAREN", "\\)")
	s.dsl.Token("LBRACKET", "\\[")
	s.dsl.Token("RBRACKET", "\\]")

	// Value tokens (higher priority for keywords)
	s.dsl.KeywordToken("TRUE", "true")           // Boolean true (high priority)
	s.dsl.KeywordToken("FALSE", "false")         // Boolean false (high priority)
	s.dsl.Token("STRING", "\"[^\"]*\"")           // Quoted strings
	s.dsl.Token("DATETIME", "[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}([.][0-9]+)?Z?") // ISO datetime
	s.dsl.Token("NUMBER", "[0-9]+\\.?[0-9]*")    // Numbers (int or float)
	s.dsl.Token("ATTRIBUTE", "[a-zA-Z][a-zA-Z0-9_.]*") // Attribute paths like "meta.created", "emails.value"
}

// setupRules defines the SCIM filter grammar rules
func (s *SCIMFilterDSL) setupRules() {
	// Main entry point
	s.dsl.Rule("filter", []string{"expression"}, "processFilter")

	// Expression hierarchy (precedence handled by rule order)
	s.dsl.Rule("expression", []string{"or_expression"}, "passthrough")

	// OR expressions (lowest precedence)
	s.dsl.Rule("or_expression", []string{"or_expression", "OR", "and_expression"}, "processLogicalOr")
	s.dsl.Rule("or_expression", []string{"and_expression"}, "passthrough")

	// AND expressions (higher precedence than OR)
	s.dsl.Rule("and_expression", []string{"and_expression", "AND", "not_expression"}, "processLogicalAnd")
	s.dsl.Rule("and_expression", []string{"not_expression"}, "passthrough")

	// NOT expressions (higher precedence than AND)
	s.dsl.Rule("not_expression", []string{"NOT", "primary_expression"}, "processLogicalNot")
	s.dsl.Rule("not_expression", []string{"primary_expression"}, "passthrough")

	// Primary expressions
	s.dsl.Rule("primary_expression", []string{"LPAREN", "expression", "RPAREN"}, "processGrouped")
	s.dsl.Rule("primary_expression", []string{"comparison"}, "passthrough")
	s.dsl.Rule("primary_expression", []string{"presence"}, "passthrough")
	s.dsl.Rule("primary_expression", []string{"complex_filter"}, "passthrough")

	// Comparison expressions
	s.dsl.Rule("comparison", []string{"ATTRIBUTE", "EQ", "value"}, "processComparison")
	s.dsl.Rule("comparison", []string{"ATTRIBUTE", "NE", "value"}, "processComparison")
	s.dsl.Rule("comparison", []string{"ATTRIBUTE", "CO", "value"}, "processComparison")
	s.dsl.Rule("comparison", []string{"ATTRIBUTE", "SW", "value"}, "processComparison")
	s.dsl.Rule("comparison", []string{"ATTRIBUTE", "EW", "value"}, "processComparison")
	s.dsl.Rule("comparison", []string{"ATTRIBUTE", "GT", "value"}, "processComparison")
	s.dsl.Rule("comparison", []string{"ATTRIBUTE", "GE", "value"}, "processComparison")
	s.dsl.Rule("comparison", []string{"ATTRIBUTE", "LT", "value"}, "processComparison")
	s.dsl.Rule("comparison", []string{"ATTRIBUTE", "LE", "value"}, "processComparison")

	// Presence test
	s.dsl.Rule("presence", []string{"ATTRIBUTE", "PR"}, "processPresence")

	// Complex attribute filters (e.g., emails[type eq "work"])
	s.dsl.Rule("complex_filter", []string{"ATTRIBUTE", "LBRACKET", "expression", "RBRACKET"}, "processComplexFilter")

	// Values
	s.dsl.Rule("value", []string{"STRING"}, "processStringValue")
	s.dsl.Rule("value", []string{"NUMBER"}, "processNumberValue")
	s.dsl.Rule("value", []string{"TRUE"}, "processTrueValue")
	s.dsl.Rule("value", []string{"FALSE"}, "processFalseValue")
	s.dsl.Rule("value", []string{"DATETIME"}, "processDateTimeValue")
}

// setupActions defines the semantic actions for SCIM filter rules
func (s *SCIMFilterDSL) setupActions() {
	s.dsl.Action("processFilter", s.processFilter)
	s.dsl.Action("passthrough", s.passthrough)
	s.dsl.Action("processLogicalOr", s.processLogicalOr)
	s.dsl.Action("processLogicalAnd", s.processLogicalAnd)
	s.dsl.Action("processLogicalNot", s.processLogicalNot)
	s.dsl.Action("processGrouped", s.processGrouped)
	s.dsl.Action("processComparison", s.processComparison)
	s.dsl.Action("processPresence", s.processPresence)
	s.dsl.Action("processComplexFilter", s.processComplexFilter)
	s.dsl.Action("processStringValue", s.processStringValue)
	s.dsl.Action("processNumberValue", s.processNumberValue)
	s.dsl.Action("processTrueValue", s.processTrueValue)
	s.dsl.Action("processFalseValue", s.processFalseValue)
	s.dsl.Action("processDateTimeValue", s.processDateTimeValue)
}

// Action implementations

func (s *SCIMFilterDSL) processFilter(args []interface{}) (interface{}, error) {
	return args[0], nil
}

func (s *SCIMFilterDSL) passthrough(args []interface{}) (interface{}, error) {
	return args[0], nil
}

func (s *SCIMFilterDSL) processLogicalOr(args []interface{}) (interface{}, error) {
	left := args[0].(FilterExpression)
	right := args[2].(FilterExpression)
	
	return FilterExpression{
		Type:     "logical",
		Operator: "or",
		Left:     &left,
		Right:    &right,
	}, nil
}

func (s *SCIMFilterDSL) processLogicalAnd(args []interface{}) (interface{}, error) {
	left := args[0].(FilterExpression)
	right := args[2].(FilterExpression)
	
	return FilterExpression{
		Type:     "logical",
		Operator: "and",
		Left:     &left,
		Right:    &right,
	}, nil
}

func (s *SCIMFilterDSL) processLogicalNot(args []interface{}) (interface{}, error) {
	expr := args[1].(FilterExpression)
	
	return FilterExpression{
		Type:       "logical",
		Operator:   "not",
		Expression: &expr,
	}, nil
}

func (s *SCIMFilterDSL) processGrouped(args []interface{}) (interface{}, error) {
	return args[1], nil // Return the expression inside parentheses
}

func (s *SCIMFilterDSL) processComparison(args []interface{}) (interface{}, error) {
	attribute := args[0].(string)
	operator := args[1].(string)
	value := args[2]

	return FilterExpression{
		Type:      "comparison",
		Attribute: attribute,
		Operator:  operator,
		Value:     value,
	}, nil
}

func (s *SCIMFilterDSL) processPresence(args []interface{}) (interface{}, error) {
	attribute := args[0].(string)
	
	return FilterExpression{
		Type:      "presence",
		Attribute: attribute,
		Operator:  "pr",
	}, nil
}

func (s *SCIMFilterDSL) processComplexFilter(args []interface{}) (interface{}, error) {
	attribute := args[0].(string)
	filter := args[2].(FilterExpression)
	
	return FilterExpression{
		Type:       "complex",
		Attribute:  attribute,
		Expression: &filter,
	}, nil
}

func (s *SCIMFilterDSL) processStringValue(args []interface{}) (interface{}, error) {
	str := args[0].(string)
	// Remove quotes
	return str[1 : len(str)-1], nil
}

func (s *SCIMFilterDSL) processNumberValue(args []interface{}) (interface{}, error) {
	str := args[0].(string)
	if strings.Contains(str, ".") {
		return strconv.ParseFloat(str, 64)
	}
	return strconv.Atoi(str)
}

func (s *SCIMFilterDSL) processTrueValue(args []interface{}) (interface{}, error) {
	return true, nil
}

func (s *SCIMFilterDSL) processFalseValue(args []interface{}) (interface{}, error) {
	return false, nil
}

func (s *SCIMFilterDSL) processDateTimeValue(args []interface{}) (interface{}, error) {
	return args[0].(string), nil // Return as string, parsing handled by engine
}

// FilterExpression represents a parsed SCIM filter expression
type FilterExpression struct {
	Type       string            `json:"type"`       // "comparison", "logical", "presence", "complex"
	Attribute  string            `json:"attribute,omitempty"`
	Operator   string            `json:"operator"`
	Value      interface{}       `json:"value,omitempty"`
	Left       *FilterExpression `json:"left,omitempty"`
	Right      *FilterExpression `json:"right,omitempty"`
	Expression *FilterExpression `json:"expression,omitempty"`
}

// Parse parses a SCIM filter string and returns the filter expression
func (s *SCIMFilterDSL) Parse(filter string, context map[string]interface{}) (*FilterExpression, error) {
	result, err := s.dsl.Use(filter, context)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SCIM filter: %w", err)
	}

	expr, ok := result.GetOutput().(FilterExpression)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result.GetOutput())
	}

	return &expr, nil
}

// Use parses and executes a SCIM filter against the provided data
func (s *SCIMFilterDSL) Use(filter string, context map[string]interface{}) (*dslbuilder.Result, error) {
	// Parse the filter
	expr, err := s.Parse(filter, context)
	if err != nil {
		return nil, err
	}

	// Execute the filter using the engine
	results, err := s.engine.ExecuteFilter(expr, context)
	if err != nil {
		return nil, err
	}

	// Return as DSL result
	return &dslbuilder.Result{Output: results}, nil
}

// GetEngine returns the SCIM engine for custom usage
func (s *SCIMFilterDSL) GetEngine() *SCIMEngine {
	return s.engine
}

// SetDataProvider allows injecting custom data provider functions
func (s *SCIMFilterDSL) SetDataProvider(provider DataProvider) {
	s.engine.SetDataProvider(provider)
}

// GetFieldNames extracts field names from a struct using reflection
func (s *SCIMFilterDSL) GetFieldNames(item interface{}) []string {
	return s.engine.GetFieldNames(item)
}

// FormatExpression returns a human-readable representation of a filter expression
func (s *SCIMFilterDSL) FormatExpression(expr *FilterExpression) string {
	switch expr.Type {
	case "comparison":
		return fmt.Sprintf("%s %s %v", expr.Attribute, expr.Operator, expr.Value)
	case "presence":
		return fmt.Sprintf("%s %s", expr.Attribute, expr.Operator)
	case "logical":
		if expr.Operator == "not" {
			return fmt.Sprintf("not (%s)", s.FormatExpression(expr.Expression))
		}
		return fmt.Sprintf("(%s %s %s)", 
			s.FormatExpression(expr.Left), 
			expr.Operator, 
			s.FormatExpression(expr.Right))
	case "complex":
		return fmt.Sprintf("%s[%s]", expr.Attribute, s.FormatExpression(expr.Expression))
	default:
		return fmt.Sprintf("unknown(%+v)", expr)
	}
}

// GetAvailableOperators returns all supported SCIM operators
func (s *SCIMFilterDSL) GetAvailableOperators() map[string]string {
	return map[string]string{
		"eq": "Equal",
		"ne": "Not Equal",
		"co": "Contains",
		"sw": "Starts With",
		"ew": "Ends With",
		"gt": "Greater Than",
		"ge": "Greater Than or Equal",
		"lt": "Less Than",
		"le": "Less Than or Equal",
		"pr": "Present (not null)",
		"and": "Logical AND",
		"or": "Logical OR",
		"not": "Logical NOT",
	}
}