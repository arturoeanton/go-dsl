package universal

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

// UniversalQueryDSL creates a universal query DSL that works with any struct type
type UniversalQueryDSL struct {
	dsl    *dslbuilder.DSL
	engine *UniversalQueryEngine
}

// NewUniversalQueryDSL creates a new universal query DSL
func NewUniversalQueryDSL() *UniversalQueryDSL {
	dsl := dslbuilder.New("UniversalQueryDSL")
	engine := NewQueryEngine()
	
	uq := &UniversalQueryDSL{
		dsl:    dsl,
		engine: engine,
	}
	
	uq.setupTokens()
	uq.setupRules()
	uq.setupActions()
	
	return uq
}

// setupTokens defines all the tokens for the query DSL
func (uq *UniversalQueryDSL) setupTokens() {
	// Spanish commands
	uq.dsl.KeywordToken("BUSCAR", "buscar")
	uq.dsl.KeywordToken("LISTAR", "listar")
	uq.dsl.KeywordToken("CONTAR", "contar")
	
	// English commands (for universal support)
	uq.dsl.KeywordToken("SEARCH", "search")
	uq.dsl.KeywordToken("LIST", "list")
	uq.dsl.KeywordToken("COUNT", "count")
	
	// Connectors (Spanish)
	uq.dsl.KeywordToken("DONDE", "donde")
	uq.dsl.KeywordToken("ES", "es")
	uq.dsl.KeywordToken("MAYOR", "mayor")
	uq.dsl.KeywordToken("MENOR", "menor")
	uq.dsl.KeywordToken("CONTIENE", "contiene")
	
	// Connectors (English)
	uq.dsl.KeywordToken("WHERE", "where")
	uq.dsl.KeywordToken("IS", "is")
	uq.dsl.KeywordToken("GREATER", "greater")
	uq.dsl.KeywordToken("LESS", "less")
	uq.dsl.KeywordToken("CONTAINS", "contains")
	
	// Value tokens
	uq.dsl.Token("STRING", `"[^"]*"`)
	uq.dsl.Token("NUMBER", `[0-9]+\.?[0-9]*`)
	uq.dsl.Token("WORD", `[a-zA-Z][a-zA-Z0-9_]*`)   // Generic word token
}

// setupRules defines grammar rules from most specific to least specific
func (uq *UniversalQueryDSL) setupRules() {
	// Most specific: String comparisons (Spanish)
	uq.dsl.Rule("query", []string{"BUSCAR", "WORD", "DONDE", "WORD", "CONTIENE", "STRING"}, "filteredQueryStringContains")
	uq.dsl.Rule("query", []string{"LISTAR", "WORD", "DONDE", "WORD", "CONTIENE", "STRING"}, "filteredQueryStringContains")
	uq.dsl.Rule("query", []string{"CONTAR", "WORD", "DONDE", "WORD", "CONTIENE", "STRING"}, "filteredQueryStringContains")
	uq.dsl.Rule("query", []string{"BUSCAR", "WORD", "DONDE", "WORD", "ES", "STRING"}, "filteredQueryStringEquals")
	uq.dsl.Rule("query", []string{"LISTAR", "WORD", "DONDE", "WORD", "ES", "STRING"}, "filteredQueryStringEquals")
	uq.dsl.Rule("query", []string{"CONTAR", "WORD", "DONDE", "WORD", "ES", "STRING"}, "filteredQueryStringEquals")
	
	// String comparisons (English)
	uq.dsl.Rule("query", []string{"SEARCH", "WORD", "WHERE", "WORD", "CONTAINS", "STRING"}, "filteredQueryStringContains")
	uq.dsl.Rule("query", []string{"LIST", "WORD", "WHERE", "WORD", "CONTAINS", "STRING"}, "filteredQueryStringContains")
	uq.dsl.Rule("query", []string{"COUNT", "WORD", "WHERE", "WORD", "CONTAINS", "STRING"}, "filteredQueryStringContains")
	uq.dsl.Rule("query", []string{"SEARCH", "WORD", "WHERE", "WORD", "IS", "STRING"}, "filteredQueryStringEquals")
	uq.dsl.Rule("query", []string{"LIST", "WORD", "WHERE", "WORD", "IS", "STRING"}, "filteredQueryStringEquals")
	uq.dsl.Rule("query", []string{"COUNT", "WORD", "WHERE", "WORD", "IS", "STRING"}, "filteredQueryStringEquals")
	
	// Numeric comparisons (Spanish)
	uq.dsl.Rule("query", []string{"BUSCAR", "WORD", "DONDE", "WORD", "MAYOR", "NUMBER"}, "filteredQueryNumberGreater")
	uq.dsl.Rule("query", []string{"LISTAR", "WORD", "DONDE", "WORD", "MAYOR", "NUMBER"}, "filteredQueryNumberGreater")
	uq.dsl.Rule("query", []string{"CONTAR", "WORD", "DONDE", "WORD", "MAYOR", "NUMBER"}, "filteredQueryNumberGreater")
	uq.dsl.Rule("query", []string{"BUSCAR", "WORD", "DONDE", "WORD", "MENOR", "NUMBER"}, "filteredQueryNumberLess")
	uq.dsl.Rule("query", []string{"LISTAR", "WORD", "DONDE", "WORD", "MENOR", "NUMBER"}, "filteredQueryNumberLess")
	uq.dsl.Rule("query", []string{"CONTAR", "WORD", "DONDE", "WORD", "MENOR", "NUMBER"}, "filteredQueryNumberLess")
	
	// Numeric comparisons (English)
	uq.dsl.Rule("query", []string{"SEARCH", "WORD", "WHERE", "WORD", "GREATER", "NUMBER"}, "filteredQueryNumberGreater")
	uq.dsl.Rule("query", []string{"LIST", "WORD", "WHERE", "WORD", "GREATER", "NUMBER"}, "filteredQueryNumberGreater")
	uq.dsl.Rule("query", []string{"COUNT", "WORD", "WHERE", "WORD", "GREATER", "NUMBER"}, "filteredQueryNumberGreater")
	uq.dsl.Rule("query", []string{"SEARCH", "WORD", "WHERE", "WORD", "LESS", "NUMBER"}, "filteredQueryNumberLess")
	uq.dsl.Rule("query", []string{"LIST", "WORD", "WHERE", "WORD", "LESS", "NUMBER"}, "filteredQueryNumberLess")
	uq.dsl.Rule("query", []string{"COUNT", "WORD", "WHERE", "WORD", "LESS", "NUMBER"}, "filteredQueryNumberLess")
	
	// Value comparisons (Spanish)
	uq.dsl.Rule("query", []string{"BUSCAR", "WORD", "DONDE", "WORD", "ES", "WORD"}, "filteredQueryValueEquals")
	uq.dsl.Rule("query", []string{"LISTAR", "WORD", "DONDE", "WORD", "ES", "WORD"}, "filteredQueryValueEquals")
	uq.dsl.Rule("query", []string{"CONTAR", "WORD", "DONDE", "WORD", "ES", "WORD"}, "filteredQueryValueEquals")
	
	// Value comparisons (English)
	uq.dsl.Rule("query", []string{"SEARCH", "WORD", "WHERE", "WORD", "IS", "WORD"}, "filteredQueryValueEquals")
	uq.dsl.Rule("query", []string{"LIST", "WORD", "WHERE", "WORD", "IS", "WORD"}, "filteredQueryValueEquals")
	uq.dsl.Rule("query", []string{"COUNT", "WORD", "WHERE", "WORD", "IS", "WORD"}, "filteredQueryValueEquals")
	
	// Simple queries (least specific - Spanish)
	uq.dsl.Rule("query", []string{"BUSCAR", "WORD"}, "simpleQuery")
	uq.dsl.Rule("query", []string{"LISTAR", "WORD"}, "simpleQuery")
	uq.dsl.Rule("query", []string{"CONTAR", "WORD"}, "simpleQuery")
	
	// Simple queries (English)
	uq.dsl.Rule("query", []string{"SEARCH", "WORD"}, "simpleQuery")
	uq.dsl.Rule("query", []string{"LIST", "WORD"}, "simpleQuery")
	uq.dsl.Rule("query", []string{"COUNT", "WORD"}, "simpleQuery")
}

// setupActions defines all the action handlers
func (uq *UniversalQueryDSL) setupActions() {
	// Simple query action
	uq.dsl.Action("simpleQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("insufficient arguments for simple query")
		}
		
		action := args[0].(string)
		entity := args[1].(string)
		
		return uq.executeSimpleQuery(action, entity)
	})
	
	// String contains action
	uq.dsl.Action("filteredQueryStringContains", func(args []interface{}) (interface{}, error) {
		if len(args) < 6 {
			return nil, fmt.Errorf("insufficient arguments for string contains query")
		}
		
		action := args[0].(string)
		entity := args[1].(string)
		field := args[3].(string)
		value := strings.Trim(args[5].(string), `"`)
		
		return uq.executeFilteredQuery(action, entity, field, "contiene", value)
	})
	
	// String equals action
	uq.dsl.Action("filteredQueryStringEquals", func(args []interface{}) (interface{}, error) {
		if len(args) < 6 {
			return nil, fmt.Errorf("insufficient arguments for string equals query")
		}
		
		action := args[0].(string)
		entity := args[1].(string)
		field := args[3].(string)
		value := strings.Trim(args[5].(string), `"`)
		
		return uq.executeFilteredQuery(action, entity, field, "es", value)
	})
	
	// Number greater action
	uq.dsl.Action("filteredQueryNumberGreater", func(args []interface{}) (interface{}, error) {
		if len(args) < 6 {
			return nil, fmt.Errorf("insufficient arguments for number greater query")
		}
		
		action := args[0].(string)
		entity := args[1].(string)
		field := args[3].(string)
		valueStr := args[5].(string)
		
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %s", valueStr)
		}
		
		return uq.executeFilteredQuery(action, entity, field, "mayor", value)
	})
	
	// Number less action
	uq.dsl.Action("filteredQueryNumberLess", func(args []interface{}) (interface{}, error) {
		if len(args) < 6 {
			return nil, fmt.Errorf("insufficient arguments for number less query")
		}
		
		action := args[0].(string)
		entity := args[1].(string)
		field := args[3].(string)
		valueStr := args[5].(string)
		
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %s", valueStr)
		}
		
		return uq.executeFilteredQuery(action, entity, field, "menor", value)
	})
	
	// Value equals action
	uq.dsl.Action("filteredQueryValueEquals", func(args []interface{}) (interface{}, error) {
		if len(args) < 6 {
			return nil, fmt.Errorf("insufficient arguments for value equals query")
		}
		
		action := args[0].(string)
		entity := args[1].(string)
		field := args[3].(string)
		value := args[5].(string)
		
		return uq.executeFilteredQuery(action, entity, field, "es", value)
	})
}

// executeSimpleQuery executes a simple query without filters
func (uq *UniversalQueryDSL) executeSimpleQuery(action, entity string) (interface{}, error) {
	// Get data from context
	data := uq.dsl.GetContext(entity)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entity)
	}
	
	// Convert to interface slice
	dataSlice := uq.engine.ConvertToInterfaceSlice(data)
	if dataSlice == nil {
		return nil, fmt.Errorf("entity '%s' is not a slice", entity)
	}
	
	switch strings.ToLower(action) {
	case "contar", "count":
		return uq.engine.Count(dataSlice), nil
	case "listar", "list", "buscar", "search":
		return dataSlice, nil
	default:
		return dataSlice, nil
	}
}

// executeFilteredQuery executes a filtered query
func (uq *UniversalQueryDSL) executeFilteredQuery(action, entity, field, operator string, value interface{}) (interface{}, error) {
	// Get data from context
	data := uq.dsl.GetContext(entity)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entity)
	}
	
	// Convert to interface slice
	dataSlice := uq.engine.ConvertToInterfaceSlice(data)
	if dataSlice == nil {
		return nil, fmt.Errorf("entity '%s' is not a slice", entity)
	}
	
	// Apply filter
	filtered := uq.engine.ApplyFilter(dataSlice, field, operator, value)
	
	switch strings.ToLower(action) {
	case "contar", "count":
		return uq.engine.Count(filtered), nil
	case "listar", "list", "buscar", "search":
		return filtered, nil
	default:
		return filtered, nil
	}
}

// Parse executes a query string
func (uq *UniversalQueryDSL) Parse(query string) (*dslbuilder.Result, error) {
	return uq.dsl.Parse(query)
}

// Use executes a query string with context
func (uq *UniversalQueryDSL) Use(query string, context map[string]interface{}) (*dslbuilder.Result, error) {
	return uq.dsl.Use(query, context)
}

// GetEngine returns the underlying query engine
func (uq *UniversalQueryDSL) GetEngine() *UniversalQueryEngine {
	return uq.engine
}

// SetContext sets a context value
func (uq *UniversalQueryDSL) SetContext(key string, value interface{}) {
	uq.dsl.SetContext(key, value)
}