package universal

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

// UniversalLinqDSL creates a universal LINQ DSL (English only for better stability)
type UniversalLinqDSL struct {
	dsl    *dslbuilder.DSL
	engine *UniversalLinqEngine
}

// NewUniversalLinqDSL creates a new universal LINQ DSL
func NewUniversalLinqDSL() *UniversalLinqDSL {
	dsl := dslbuilder.New("UniversalLinqDSL")

	ul := &UniversalLinqDSL{
		dsl:    dsl,
		engine: nil,
	}

	ul.setupTokens()
	ul.setupRules()
	ul.setupActions()

	return ul
}

// setupTokens defines all the tokens for the LINQ DSL (English only)
func (ul *UniversalLinqDSL) setupTokens() {
	// Value tokens must be defined FIRST, before keywords
	ul.dsl.Token("STRING", `"[^"]*"`)
	ul.dsl.Token("NUMBER", `[0-9]+\.?[0-9]*`)
	ul.dsl.Token("WORD", `[a-zA-Z][a-zA-Z0-9_]*`)
	ul.dsl.Token("ASTERISK", `\*`)

	// LINQ keywords - ORDER must be after WORD is defined
	ul.dsl.KeywordToken("FROM", "from")
	ul.dsl.KeywordToken("WHERE", "where")
	ul.dsl.KeywordToken("SELECT", "select")
	ul.dsl.KeywordToken("ORDER", "order")
	ul.dsl.KeywordToken("BY", "by")
	ul.dsl.KeywordToken("GROUP", "group")
	ul.dsl.KeywordToken("JOIN", "join")
	ul.dsl.KeywordToken("ON", "on")
	ul.dsl.KeywordToken("INTO", "into")
	ul.dsl.KeywordToken("LET", "let")
	ul.dsl.KeywordToken("DISTINCT", "distinct")
	ul.dsl.KeywordToken("TAKE", "take")
	ul.dsl.KeywordToken("SKIP", "skip")
	ul.dsl.KeywordToken("FIRST", "first")
	ul.dsl.KeywordToken("LAST", "last")
	ul.dsl.KeywordToken("SINGLE", "single")
	ul.dsl.KeywordToken("COUNT", "count")
	ul.dsl.KeywordToken("SUM", "sum")
	ul.dsl.KeywordToken("AVG", "avg")
	ul.dsl.KeywordToken("AVERAGE", "average")
	ul.dsl.KeywordToken("MIN", "min")
	ul.dsl.KeywordToken("MAX", "max")
	ul.dsl.KeywordToken("ANY", "any")
	ul.dsl.KeywordToken("ALL", "all")
	ul.dsl.KeywordToken("CONTAINS", "contains")
	ul.dsl.KeywordToken("REVERSE", "reverse")
	ul.dsl.KeywordToken("UNION", "union")
	ul.dsl.KeywordToken("INTERSECT", "intersect")
	ul.dsl.KeywordToken("EXCEPT", "except")

	// Sort direction
	ul.dsl.KeywordToken("ASC", "asc")
	ul.dsl.KeywordToken("DESC", "desc")
	ul.dsl.KeywordToken("ASCENDING", "ascending")
	ul.dsl.KeywordToken("DESCENDING", "descending")

	// Operators
	ul.dsl.KeywordToken("EQ", "==")
	ul.dsl.KeywordToken("NE", "!=")
	ul.dsl.KeywordToken("GT", ">")
	ul.dsl.KeywordToken("GE", ">=")
	ul.dsl.KeywordToken("LT", "<")
	ul.dsl.KeywordToken("LE", "<=")
	ul.dsl.KeywordToken("AND", "and")
	ul.dsl.KeywordToken("OR", "or")
	ul.dsl.KeywordToken("NOT", "not")
	ul.dsl.KeywordToken("IN", "in")
	ul.dsl.KeywordToken("LIKE", "like")

	// Additional operators for completeness
	ul.dsl.KeywordToken("EQUALS", "equals")
	ul.dsl.KeywordToken("GREATER", "greater")
	ul.dsl.KeywordToken("LESS", "less")
	ul.dsl.KeywordToken("THAN", "than")
	ul.dsl.KeywordToken("STARTSWITH", "startswith")
	ul.dsl.KeywordToken("ENDSWITH", "endswith")
}

// setupRules defines comprehensive grammar rules for LINQ-like syntax
func (ul *UniversalLinqDSL) setupRules() {
	// Basic SELECT queries
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "SELECT", "WORD"}, "basicSelectQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "SELECT", "ASTERISK"}, "basicSelectAllQuery")

	// WHERE queries with different operators
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "WHERE", "WORD", "GT", "NUMBER", "SELECT", "WORD"}, "whereSelectQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "WHERE", "WORD", "LT", "NUMBER", "SELECT", "WORD"}, "whereSelectQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "WHERE", "WORD", "GE", "NUMBER", "SELECT", "WORD"}, "whereSelectQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "WHERE", "WORD", "LE", "NUMBER", "SELECT", "WORD"}, "whereSelectQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "WHERE", "WORD", "EQ", "STRING", "SELECT", "WORD"}, "whereSelectQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "WHERE", "WORD", "EQ", "NUMBER", "SELECT", "WORD"}, "whereSelectQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "WHERE", "WORD", "NE", "STRING", "SELECT", "WORD"}, "whereSelectQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "WHERE", "WORD", "NE", "NUMBER", "SELECT", "WORD"}, "whereSelectQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "WHERE", "WORD", "CONTAINS", "STRING", "SELECT", "WORD"}, "whereSelectQuery")

	// WHERE queries with SELECT *
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "WHERE", "WORD", "GT", "NUMBER", "SELECT", "ASTERISK"}, "whereSelectAllQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "WHERE", "WORD", "LT", "NUMBER", "SELECT", "ASTERISK"}, "whereSelectAllQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "WHERE", "WORD", "EQ", "STRING", "SELECT", "ASTERISK"}, "whereSelectAllQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "WHERE", "WORD", "EQ", "NUMBER", "SELECT", "ASTERISK"}, "whereSelectAllQuery")

	// ORDER BY queries
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "ORDER", "BY", "WORD", "SELECT", "WORD"}, "orderBySelectQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "ORDER", "BY", "WORD", "ASC", "SELECT", "WORD"}, "orderBySelectQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "ORDER", "BY", "WORD", "DESC", "SELECT", "WORD"}, "orderByDescSelectQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "ORDER", "BY", "WORD", "ASCENDING", "SELECT", "WORD"}, "orderBySelectQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "ORDER", "BY", "WORD", "DESCENDING", "SELECT", "WORD"}, "orderByDescSelectQuery")

	// ORDER BY with SELECT *
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "ORDER", "BY", "WORD", "SELECT", "ASTERISK"}, "orderBySelectAllQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "ORDER", "BY", "WORD", "DESC", "SELECT", "ASTERISK"}, "orderByDescSelectAllQuery")

	// Combined WHERE and ORDER BY
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "WHERE", "WORD", "GT", "NUMBER", "ORDER", "BY", "WORD", "SELECT", "WORD"}, "whereOrderBySelectQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "WHERE", "WORD", "EQ", "STRING", "ORDER", "BY", "WORD", "SELECT", "WORD"}, "whereOrderBySelectQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "WHERE", "WORD", "GT", "NUMBER", "ORDER", "BY", "WORD", "DESC", "SELECT", "WORD"}, "whereOrderByDescSelectQuery")

	// Aggregation queries
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "COUNT"}, "countQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "WHERE", "WORD", "GT", "NUMBER", "COUNT"}, "whereCountQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "WHERE", "WORD", "EQ", "STRING", "COUNT"}, "whereCountQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "SUM", "WORD"}, "sumQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "WHERE", "WORD", "GT", "NUMBER", "SUM", "WORD"}, "whereSumQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "AVG", "WORD"}, "avgQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "AVERAGE", "WORD"}, "avgQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "WHERE", "WORD", "GT", "NUMBER", "AVG", "WORD"}, "whereAvgQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "MIN", "WORD"}, "minQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "MAX", "WORD"}, "maxQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "WHERE", "WORD", "GT", "NUMBER", "MIN", "WORD"}, "whereMinQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "WHERE", "WORD", "GT", "NUMBER", "MAX", "WORD"}, "whereMaxQuery")

	// Group by queries
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "GROUP", "BY", "WORD"}, "groupByQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "GROUP", "BY", "WORD", "SELECT", "WORD"}, "groupBySelectQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "WHERE", "WORD", "GT", "NUMBER", "GROUP", "BY", "WORD"}, "whereGroupByQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "WHERE", "WORD", "EQ", "STRING", "GROUP", "BY", "WORD"}, "whereGroupByQuery")

	// Take/Skip queries
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "TAKE", "NUMBER", "SELECT", "WORD"}, "takeSelectQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "TAKE", "NUMBER", "SELECT", "ASTERISK"}, "takeSelectAllQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "SKIP", "NUMBER", "SELECT", "WORD"}, "skipSelectQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "SKIP", "NUMBER", "SELECT", "ASTERISK"}, "skipSelectAllQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "SKIP", "NUMBER", "TAKE", "NUMBER", "SELECT", "WORD"}, "skipTakeSelectQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "SKIP", "NUMBER", "TAKE", "NUMBER", "SELECT", "ASTERISK"}, "skipTakeSelectAllQuery")

	// Take/Skip with WHERE
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "WHERE", "WORD", "GT", "NUMBER", "TAKE", "NUMBER", "SELECT", "WORD"}, "whereTakeSelectQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "WHERE", "WORD", "EQ", "STRING", "TAKE", "NUMBER", "SELECT", "WORD"}, "whereTakeSelectQuery")

	// Distinct queries
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "SELECT", "DISTINCT", "WORD"}, "distinctSelectQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "DISTINCT"}, "distinctQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "WHERE", "WORD", "GT", "NUMBER", "DISTINCT"}, "whereDistinctQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "WHERE", "WORD", "EQ", "STRING", "SELECT", "DISTINCT", "WORD"}, "whereDistinctSelectQuery")

	// First/Last queries
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "FIRST"}, "firstQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "LAST"}, "lastQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "SINGLE"}, "singleQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "WHERE", "WORD", "GT", "NUMBER", "FIRST"}, "whereFirstQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "WHERE", "WORD", "EQ", "STRING", "FIRST"}, "whereFirstQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "WHERE", "WORD", "GT", "NUMBER", "LAST"}, "whereLastQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "WHERE", "WORD", "EQ", "STRING", "LAST"}, "whereLastQuery")

	// Reverse queries
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "REVERSE", "SELECT", "WORD"}, "reverseSelectQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "REVERSE", "SELECT", "ASTERISK"}, "reverseSelectAllQuery")

	// Any/All queries
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "ANY"}, "anyQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "ALL"}, "allQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "WHERE", "WORD", "GT", "NUMBER", "ANY"}, "whereAnyQuery")
	ul.dsl.Rule("linq_query", []string{"FROM", "WORD", "WHERE", "WORD", "EQ", "STRING", "ANY"}, "whereAnyQuery")
}

// setupActions defines all the action handlers
func (ul *UniversalLinqDSL) setupActions() {
	// Basic select queries
	ul.dsl.Action("basicSelectQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("insufficient arguments for basic select query")
		}
		entityName := args[1].(string)
		selectField := args[3].(string)
		return ul.executeBasicSelect(entityName, selectField)
	})

	ul.dsl.Action("basicSelectAllQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("insufficient arguments for basic select all query")
		}
		entityName := args[1].(string)
		return ul.executeBasicSelect(entityName, "*")
	})

	// Where select queries
	ul.dsl.Action("whereSelectQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 8 {
			return nil, fmt.Errorf("insufficient arguments for where select query")
		}
		entityName := args[1].(string)
		whereField := args[3].(string)
		operator := args[4].(string)
		var value interface{}
		if strings.Contains(args[5].(string), `"`) {
			value = strings.Trim(args[5].(string), `"`)
		} else {
			value = args[5].(string)
		}
		selectField := args[7].(string)
		return ul.executeWhereSelect(entityName, whereField, operator, value, selectField)
	})

	ul.dsl.Action("whereSelectAllQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 8 {
			return nil, fmt.Errorf("insufficient arguments for where select all query")
		}
		entityName := args[1].(string)
		whereField := args[3].(string)
		operator := args[4].(string)
		var value interface{}
		if strings.Contains(args[5].(string), `"`) {
			value = strings.Trim(args[5].(string), `"`)
		} else {
			value = args[5].(string)
		}
		return ul.executeWhereSelect(entityName, whereField, operator, value, "*")
	})

	// Order by queries
	ul.dsl.Action("orderBySelectQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 7 {
			return nil, fmt.Errorf("insufficient arguments for order by select query")
		}
		entityName := args[1].(string)
		orderField := args[4].(string)
		selectField := args[6].(string)
		return ul.executeOrderBySelect(entityName, orderField, "asc", selectField)
	})

	ul.dsl.Action("orderByDescSelectQuery", func(args []interface{}) (interface{}, error) {
		var entityName, orderField, selectField string
		if len(args) >= 8 {
			entityName = args[1].(string)
			orderField = args[4].(string)
			selectField = args[7].(string)
		} else if len(args) >= 7 {
			entityName = args[1].(string)
			orderField = args[4].(string)
			selectField = args[6].(string)
		} else {
			return nil, fmt.Errorf("insufficient arguments for order by desc select query")
		}
		return ul.executeOrderBySelect(entityName, orderField, "desc", selectField)
	})

	ul.dsl.Action("orderBySelectAllQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 7 {
			return nil, fmt.Errorf("insufficient arguments for order by select all query")
		}
		entityName := args[1].(string)
		orderField := args[4].(string)
		return ul.executeOrderBySelect(entityName, orderField, "asc", "*")
	})

	ul.dsl.Action("orderByDescSelectAllQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 8 {
			return nil, fmt.Errorf("insufficient arguments for order by desc select all query")
		}
		entityName := args[1].(string)
		orderField := args[4].(string)
		return ul.executeOrderBySelect(entityName, orderField, "desc", "*")
	})

	// Combined WHERE and ORDER BY
	ul.dsl.Action("whereOrderBySelectQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 11 {
			return nil, fmt.Errorf("insufficient arguments for where order by select query")
		}
		entityName := args[1].(string)
		whereField := args[3].(string)
		operator := args[4].(string)
		var value interface{}
		if strings.Contains(args[5].(string), `"`) {
			value = strings.Trim(args[5].(string), `"`)
		} else {
			value = args[5].(string)
		}
		orderField := args[8].(string)
		selectField := args[10].(string)
		return ul.executeWhereOrderBySelect(entityName, whereField, operator, value, orderField, "asc", selectField)
	})

	ul.dsl.Action("whereOrderByDescSelectQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 12 {
			return nil, fmt.Errorf("insufficient arguments for where order by desc select query")
		}
		entityName := args[1].(string)
		whereField := args[3].(string)
		operator := args[4].(string)
		var value interface{}
		if strings.Contains(args[5].(string), `"`) {
			value = strings.Trim(args[5].(string), `"`)
		} else {
			value = args[5].(string)
		}
		orderField := args[8].(string)
		selectField := args[11].(string)
		return ul.executeWhereOrderBySelect(entityName, whereField, operator, value, orderField, "desc", selectField)
	})

	// Count queries
	ul.dsl.Action("countQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("insufficient arguments for count query")
		}
		entityName := args[1].(string)
		return ul.executeCount(entityName)
	})

	ul.dsl.Action("whereCountQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 7 {
			return nil, fmt.Errorf("insufficient arguments for where count query")
		}
		entityName := args[1].(string)
		whereField := args[3].(string)
		operator := args[4].(string)
		var value interface{}
		if strings.Contains(args[5].(string), `"`) {
			value = strings.Trim(args[5].(string), `"`)
		} else {
			value = args[5].(string)
		}
		return ul.executeWhereCount(entityName, whereField, operator, value)
	})

	// Aggregation queries
	ul.dsl.Action("sumQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("insufficient arguments for sum query")
		}
		entityName := args[1].(string)
		field := args[3].(string)
		return ul.executeSum(entityName, field)
	})

	ul.dsl.Action("whereSumQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 8 {
			return nil, fmt.Errorf("insufficient arguments for where sum query")
		}
		entityName := args[1].(string)
		whereField := args[3].(string)
		operator := args[4].(string)
		var value interface{}
		if strings.Contains(args[5].(string), `"`) {
			value = strings.Trim(args[5].(string), `"`)
		} else {
			value = args[5].(string)
		}
		sumField := args[7].(string)
		return ul.executeWhereSum(entityName, whereField, operator, value, sumField)
	})

	ul.dsl.Action("avgQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("insufficient arguments for avg query")
		}
		entityName := args[1].(string)
		field := args[3].(string)
		return ul.executeAvg(entityName, field)
	})

	ul.dsl.Action("whereAvgQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 8 {
			return nil, fmt.Errorf("insufficient arguments for where avg query")
		}
		entityName := args[1].(string)
		whereField := args[3].(string)
		operator := args[4].(string)
		var value interface{}
		if strings.Contains(args[5].(string), `"`) {
			value = strings.Trim(args[5].(string), `"`)
		} else {
			value = args[5].(string)
		}
		avgField := args[7].(string)
		return ul.executeWhereAvg(entityName, whereField, operator, value, avgField)
	})

	ul.dsl.Action("minQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("insufficient arguments for min query")
		}
		entityName := args[1].(string)
		field := args[3].(string)
		return ul.executeMin(entityName, field)
	})

	ul.dsl.Action("whereMinQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 8 {
			return nil, fmt.Errorf("insufficient arguments for where min query")
		}
		entityName := args[1].(string)
		whereField := args[3].(string)
		operator := args[4].(string)
		var value interface{}
		if strings.Contains(args[5].(string), `"`) {
			value = strings.Trim(args[5].(string), `"`)
		} else {
			value = args[5].(string)
		}
		minField := args[7].(string)
		return ul.executeWhereMin(entityName, whereField, operator, value, minField)
	})

	ul.dsl.Action("maxQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("insufficient arguments for max query")
		}
		entityName := args[1].(string)
		field := args[3].(string)
		return ul.executeMax(entityName, field)
	})

	ul.dsl.Action("whereMaxQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 8 {
			return nil, fmt.Errorf("insufficient arguments for where max query")
		}
		entityName := args[1].(string)
		whereField := args[3].(string)
		operator := args[4].(string)
		var value interface{}
		if strings.Contains(args[5].(string), `"`) {
			value = strings.Trim(args[5].(string), `"`)
		} else {
			value = args[5].(string)
		}
		maxField := args[7].(string)
		return ul.executeWhereMax(entityName, whereField, operator, value, maxField)
	})

	// Group by queries
	ul.dsl.Action("groupByQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 5 {
			return nil, fmt.Errorf("insufficient arguments for group by query")
		}
		entityName := args[1].(string)
		groupField := args[4].(string)
		return ul.executeGroupBy(entityName, groupField)
	})

	ul.dsl.Action("groupBySelectQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 7 {
			return nil, fmt.Errorf("insufficient arguments for group by select query")
		}
		entityName := args[1].(string)
		groupField := args[4].(string)
		selectField := args[6].(string)
		return ul.executeGroupBySelect(entityName, groupField, selectField)
	})

	ul.dsl.Action("whereGroupByQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 9 {
			return nil, fmt.Errorf("insufficient arguments for where group by query")
		}
		entityName := args[1].(string)
		whereField := args[3].(string)
		operator := args[4].(string)
		var value interface{}
		if strings.Contains(args[5].(string), `"`) {
			value = strings.Trim(args[5].(string), `"`)
		} else {
			value = args[5].(string)
		}
		groupField := args[8].(string)
		return ul.executeWhereGroupBy(entityName, whereField, operator, value, groupField)
	})

	// Take/Skip queries
	ul.dsl.Action("takeSelectQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 6 {
			return nil, fmt.Errorf("insufficient arguments for take select query")
		}
		entityName := args[1].(string)
		countStr := args[3].(string)
		selectField := args[5].(string)
		count, err := strconv.Atoi(countStr)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %s", countStr)
		}
		return ul.executeTakeSelect(entityName, count, selectField)
	})

	ul.dsl.Action("takeSelectAllQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 6 {
			return nil, fmt.Errorf("insufficient arguments for take select all query")
		}
		entityName := args[1].(string)
		countStr := args[3].(string)
		count, err := strconv.Atoi(countStr)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %s", countStr)
		}
		return ul.executeTakeSelect(entityName, count, "*")
	})

	ul.dsl.Action("skipSelectQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 6 {
			return nil, fmt.Errorf("insufficient arguments for skip select query")
		}
		entityName := args[1].(string)
		countStr := args[3].(string)
		selectField := args[5].(string)
		count, err := strconv.Atoi(countStr)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %s", countStr)
		}
		return ul.executeSkipSelect(entityName, count, selectField)
	})

	ul.dsl.Action("skipSelectAllQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 6 {
			return nil, fmt.Errorf("insufficient arguments for skip select all query")
		}
		entityName := args[1].(string)
		countStr := args[3].(string)
		count, err := strconv.Atoi(countStr)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %s", countStr)
		}
		return ul.executeSkipSelect(entityName, count, "*")
	})

	ul.dsl.Action("skipTakeSelectQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 8 {
			return nil, fmt.Errorf("insufficient arguments for skip take select query")
		}
		entityName := args[1].(string)
		skipCountStr := args[3].(string)
		takeCountStr := args[5].(string)
		selectField := args[7].(string)
		skipCount, err := strconv.Atoi(skipCountStr)
		if err != nil {
			return nil, fmt.Errorf("invalid skip number: %s", skipCountStr)
		}
		takeCount, err := strconv.Atoi(takeCountStr)
		if err != nil {
			return nil, fmt.Errorf("invalid take number: %s", takeCountStr)
		}
		return ul.executeSkipTakeSelect(entityName, skipCount, takeCount, selectField)
	})

	ul.dsl.Action("skipTakeSelectAllQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 8 {
			return nil, fmt.Errorf("insufficient arguments for skip take select all query")
		}
		entityName := args[1].(string)
		skipCountStr := args[3].(string)
		takeCountStr := args[5].(string)
		skipCount, err := strconv.Atoi(skipCountStr)
		if err != nil {
			return nil, fmt.Errorf("invalid skip number: %s", skipCountStr)
		}
		takeCount, err := strconv.Atoi(takeCountStr)
		if err != nil {
			return nil, fmt.Errorf("invalid take number: %s", takeCountStr)
		}
		return ul.executeSkipTakeSelect(entityName, skipCount, takeCount, "*")
	})

	ul.dsl.Action("whereTakeSelectQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 10 {
			return nil, fmt.Errorf("insufficient arguments for where take select query")
		}
		entityName := args[1].(string)
		whereField := args[3].(string)
		operator := args[4].(string)
		var value interface{}
		if strings.Contains(args[5].(string), `"`) {
			value = strings.Trim(args[5].(string), `"`)
		} else {
			value = args[5].(string)
		}
		countStr := args[7].(string)
		selectField := args[9].(string)
		count, err := strconv.Atoi(countStr)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %s", countStr)
		}
		return ul.executeWhereTakeSelect(entityName, whereField, operator, value, count, selectField)
	})

	// Distinct queries
	ul.dsl.Action("distinctSelectQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 5 {
			return nil, fmt.Errorf("insufficient arguments for distinct select query")
		}
		entityName := args[1].(string)
		selectField := args[4].(string)
		return ul.executeDistinctSelect(entityName, selectField)
	})

	ul.dsl.Action("distinctQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("insufficient arguments for distinct query")
		}
		entityName := args[1].(string)
		return ul.executeDistinct(entityName)
	})

	ul.dsl.Action("whereDistinctQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 7 {
			return nil, fmt.Errorf("insufficient arguments for where distinct query")
		}
		entityName := args[1].(string)
		whereField := args[3].(string)
		operator := args[4].(string)
		var value interface{}
		if strings.Contains(args[5].(string), `"`) {
			value = strings.Trim(args[5].(string), `"`)
		} else {
			value = args[5].(string)
		}
		return ul.executeWhereDistinct(entityName, whereField, operator, value)
	})

	ul.dsl.Action("whereDistinctSelectQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 9 {
			return nil, fmt.Errorf("insufficient arguments for where distinct select query")
		}
		entityName := args[1].(string)
		whereField := args[3].(string)
		operator := args[4].(string)
		var value interface{}
		if strings.Contains(args[5].(string), `"`) {
			value = strings.Trim(args[5].(string), `"`)
		} else {
			value = args[5].(string)
		}
		selectField := args[8].(string)
		return ul.executeWhereDistinctSelect(entityName, whereField, operator, value, selectField)
	})

	// First/Last/Single queries
	ul.dsl.Action("firstQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("insufficient arguments for first query")
		}
		entityName := args[1].(string)
		return ul.executeFirst(entityName)
	})

	ul.dsl.Action("lastQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("insufficient arguments for last query")
		}
		entityName := args[1].(string)
		return ul.executeLast(entityName)
	})

	ul.dsl.Action("singleQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("insufficient arguments for single query")
		}
		entityName := args[1].(string)
		return ul.executeSingle(entityName)
	})

	ul.dsl.Action("whereFirstQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 7 {
			return nil, fmt.Errorf("insufficient arguments for where first query")
		}
		entityName := args[1].(string)
		whereField := args[3].(string)
		operator := args[4].(string)
		var value interface{}
		if strings.Contains(args[5].(string), `"`) {
			value = strings.Trim(args[5].(string), `"`)
		} else {
			value = args[5].(string)
		}
		return ul.executeWhereFirst(entityName, whereField, operator, value)
	})

	ul.dsl.Action("whereLastQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 7 {
			return nil, fmt.Errorf("insufficient arguments for where last query")
		}
		entityName := args[1].(string)
		whereField := args[3].(string)
		operator := args[4].(string)
		var value interface{}
		if strings.Contains(args[5].(string), `"`) {
			value = strings.Trim(args[5].(string), `"`)
		} else {
			value = args[5].(string)
		}
		return ul.executeWhereLast(entityName, whereField, operator, value)
	})

	// Reverse queries
	ul.dsl.Action("reverseSelectQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 5 {
			return nil, fmt.Errorf("insufficient arguments for reverse select query")
		}
		entityName := args[1].(string)
		selectField := args[4].(string)
		return ul.executeReverseSelect(entityName, selectField)
	})

	ul.dsl.Action("reverseSelectAllQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 5 {
			return nil, fmt.Errorf("insufficient arguments for reverse select all query")
		}
		entityName := args[1].(string)
		return ul.executeReverseSelect(entityName, "*")
	})

	// Any/All queries
	ul.dsl.Action("anyQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("insufficient arguments for any query")
		}
		entityName := args[1].(string)
		return ul.executeAny(entityName)
	})

	ul.dsl.Action("allQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("insufficient arguments for all query")
		}
		entityName := args[1].(string)
		return ul.executeAll(entityName)
	})

	ul.dsl.Action("whereAnyQuery", func(args []interface{}) (interface{}, error) {
		if len(args) < 7 {
			return nil, fmt.Errorf("insufficient arguments for where any query")
		}
		entityName := args[1].(string)
		whereField := args[3].(string)
		operator := args[4].(string)
		var value interface{}
		if strings.Contains(args[5].(string), `"`) {
			value = strings.Trim(args[5].(string), `"`)
		} else {
			value = args[5].(string)
		}
		return ul.executeWhereAny(entityName, whereField, operator, value)
	})
}

// Execute methods - All the actual implementations

func (ul *UniversalLinqDSL) executeBasicSelect(entityName, selectField string) (interface{}, error) {
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entityName)
	}

	linq := From(data)
	if selectField == "*" {
		return linq.ToSlice(), nil
	}

	result := linq.SelectField(selectField).ToSlice()
	return result, nil
}

func (ul *UniversalLinqDSL) executeWhereSelect(entityName, whereField, operator string, value interface{}, selectField string) (interface{}, error) {
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entityName)
	}

	linq := From(data).WhereField(whereField, operator, value)
	if selectField == "*" {
		return linq.ToSlice(), nil
	}

	result := linq.SelectField(selectField).ToSlice()
	return result, nil
}

func (ul *UniversalLinqDSL) executeOrderBySelect(entityName, orderField, direction, selectField string) (interface{}, error) {
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entityName)
	}

	var linq *LinqResult
	if direction == "desc" {
		linq = From(data).OrderByFieldDescending(orderField)
	} else {
		linq = From(data).OrderByField(orderField)
	}

	if selectField == "*" {
		return linq.ToSlice(), nil
	}

	result := linq.SelectField(selectField).ToSlice()
	return result, nil
}

func (ul *UniversalLinqDSL) executeWhereOrderBySelect(entityName, whereField, operator string, value interface{}, orderField, direction, selectField string) (interface{}, error) {
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entityName)
	}

	linq := From(data).WhereField(whereField, operator, value)
	if direction == "desc" {
		linq = linq.OrderByFieldDescending(orderField)
	} else {
		linq = linq.OrderByField(orderField)
	}

	if selectField == "*" {
		return linq.ToSlice(), nil
	}

	result := linq.SelectField(selectField).ToSlice()
	return result, nil
}

func (ul *UniversalLinqDSL) executeCount(entityName string) (interface{}, error) {
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entityName)
	}

	result := From(data).Count()
	return result, nil
}

func (ul *UniversalLinqDSL) executeWhereCount(entityName, whereField, operator string, value interface{}) (interface{}, error) {
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entityName)
	}

	result := From(data).WhereField(whereField, operator, value).Count()
	return result, nil
}

func (ul *UniversalLinqDSL) executeSum(entityName, field string) (interface{}, error) {
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entityName)
	}

	result := From(data).SumField(field)
	return result, nil
}

func (ul *UniversalLinqDSL) executeWhereSum(entityName, whereField, operator string, value interface{}, sumField string) (interface{}, error) {
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entityName)
	}

	result := From(data).WhereField(whereField, operator, value).SumField(sumField)
	return result, nil
}

func (ul *UniversalLinqDSL) executeAvg(entityName, field string) (interface{}, error) {
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entityName)
	}

	result := From(data).AverageField(field)
	return result, nil
}

func (ul *UniversalLinqDSL) executeWhereAvg(entityName, whereField, operator string, value interface{}, avgField string) (interface{}, error) {
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entityName)
	}

	result := From(data).WhereField(whereField, operator, value).AverageField(avgField)
	return result, nil
}

func (ul *UniversalLinqDSL) executeMin(entityName, field string) (interface{}, error) {
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entityName)
	}

	result := From(data).MinField(field)
	return result, nil
}

func (ul *UniversalLinqDSL) executeWhereMin(entityName, whereField, operator string, value interface{}, minField string) (interface{}, error) {
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entityName)
	}

	result := From(data).WhereField(whereField, operator, value).MinField(minField)
	return result, nil
}

func (ul *UniversalLinqDSL) executeMax(entityName, field string) (interface{}, error) {
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entityName)
	}

	result := From(data).MaxField(field)
	return result, nil
}

func (ul *UniversalLinqDSL) executeWhereMax(entityName, whereField, operator string, value interface{}, maxField string) (interface{}, error) {
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entityName)
	}

	result := From(data).WhereField(whereField, operator, value).MaxField(maxField)
	return result, nil
}

func (ul *UniversalLinqDSL) executeGroupBy(entityName, groupField string) (interface{}, error) {
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entityName)
	}

	result := From(data).GroupByField(groupField)
	return result, nil
}

func (ul *UniversalLinqDSL) executeGroupBySelect(entityName, groupField, selectField string) (interface{}, error) {
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entityName)
	}

	groups := From(data).GroupByField(groupField)
	var result []interface{}

	for _, group := range groups {
		if selectField == "key" {
			result = append(result, group.Key)
		} else if selectField == "count" {
			result = append(result, group.Count)
		} else {
			// Select field from first item in group
			if len(group.Items) > 0 {
				fieldValue := getFieldValue(group.Items[0], selectField)
				if fieldValue != nil {
					result = append(result, fieldValue)
				}
			}
		}
	}

	return result, nil
}

func (ul *UniversalLinqDSL) executeWhereGroupBy(entityName, whereField, operator string, value interface{}, groupField string) (interface{}, error) {
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entityName)
	}

	result := From(data).WhereField(whereField, operator, value).GroupByField(groupField)
	return result, nil
}

func (ul *UniversalLinqDSL) executeTakeSelect(entityName string, count int, selectField string) (interface{}, error) {
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entityName)
	}

	linq := From(data).Take(count)
	if selectField == "*" {
		return linq.ToSlice(), nil
	}

	result := linq.SelectField(selectField).ToSlice()
	return result, nil
}

func (ul *UniversalLinqDSL) executeSkipSelect(entityName string, count int, selectField string) (interface{}, error) {
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entityName)
	}

	linq := From(data).Skip(count)
	if selectField == "*" {
		return linq.ToSlice(), nil
	}

	result := linq.SelectField(selectField).ToSlice()
	return result, nil
}

func (ul *UniversalLinqDSL) executeSkipTakeSelect(entityName string, skipCount, takeCount int, selectField string) (interface{}, error) {
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entityName)
	}

	linq := From(data).Skip(skipCount).Take(takeCount)
	if selectField == "*" {
		return linq.ToSlice(), nil
	}

	result := linq.SelectField(selectField).ToSlice()
	return result, nil
}

func (ul *UniversalLinqDSL) executeWhereTakeSelect(entityName, whereField, operator string, value interface{}, count int, selectField string) (interface{}, error) {
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entityName)
	}

	linq := From(data).WhereField(whereField, operator, value).Take(count)
	if selectField == "*" {
		return linq.ToSlice(), nil
	}

	result := linq.SelectField(selectField).ToSlice()
	return result, nil
}

func (ul *UniversalLinqDSL) executeDistinctSelect(entityName, selectField string) (interface{}, error) {
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entityName)
	}

	if selectField == "*" {
		result := From(data).Distinct().ToSlice()
		return result, nil
	}

	result := From(data).SelectField(selectField).Distinct().ToSlice()
	return result, nil
}

func (ul *UniversalLinqDSL) executeDistinct(entityName string) (interface{}, error) {
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entityName)
	}

	result := From(data).Distinct().ToSlice()
	return result, nil
}

func (ul *UniversalLinqDSL) executeWhereDistinct(entityName, whereField, operator string, value interface{}) (interface{}, error) {
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entityName)
	}

	result := From(data).WhereField(whereField, operator, value).Distinct().ToSlice()
	return result, nil
}

func (ul *UniversalLinqDSL) executeWhereDistinctSelect(entityName, whereField, operator string, value interface{}, selectField string) (interface{}, error) {
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entityName)
	}

	linq := From(data).WhereField(whereField, operator, value)
	if selectField == "*" {
		result := linq.Distinct().ToSlice()
		return result, nil
	}

	result := linq.SelectField(selectField).Distinct().ToSlice()
	return result, nil
}

func (ul *UniversalLinqDSL) executeFirst(entityName string) (interface{}, error) {
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entityName)
	}

	result := From(data).First()
	return result, nil
}

func (ul *UniversalLinqDSL) executeLast(entityName string) (interface{}, error) {
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entityName)
	}

	result := From(data).Last()
	return result, nil
}

func (ul *UniversalLinqDSL) executeSingle(entityName string) (interface{}, error) {
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entityName)
	}

	result := From(data).Single()
	return result, nil
}

func (ul *UniversalLinqDSL) executeWhereFirst(entityName, whereField, operator string, value interface{}) (interface{}, error) {
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entityName)
	}

	result := From(data).WhereField(whereField, operator, value).First()
	return result, nil
}

func (ul *UniversalLinqDSL) executeWhereLast(entityName, whereField, operator string, value interface{}) (interface{}, error) {
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entityName)
	}

	result := From(data).WhereField(whereField, operator, value).Last()
	return result, nil
}

func (ul *UniversalLinqDSL) executeReverseSelect(entityName, selectField string) (interface{}, error) {
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entityName)
	}

	linq := From(data).Reverse()
	if selectField == "*" {
		return linq.ToSlice(), nil
	}

	result := linq.SelectField(selectField).ToSlice()
	return result, nil
}

func (ul *UniversalLinqDSL) executeAny(entityName string) (interface{}, error) {
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entityName)
	}

	result := From(data).Count() > 0
	return result, nil
}

func (ul *UniversalLinqDSL) executeAll(entityName string) (interface{}, error) {
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entityName)
	}

	// For simplicity, "all" without condition returns true if any elements exist
	result := From(data).Count() > 0
	return result, nil
}

func (ul *UniversalLinqDSL) executeWhereAny(entityName, whereField, operator string, value interface{}) (interface{}, error) {
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		return nil, fmt.Errorf("entity '%s' not found in context", entityName)
	}

	result := From(data).WhereField(whereField, operator, value).Count() > 0
	return result, nil
}

// Parse executes a LINQ query
func (ul *UniversalLinqDSL) Parse(query string) (*dslbuilder.Result, error) {
	return ul.dsl.Parse(query)
}

// Use executes a LINQ query with context
func (ul *UniversalLinqDSL) Use(query string, context map[string]interface{}) (*dslbuilder.Result, error) {
	return ul.dsl.Use(query, context)
}

// SetContext sets a context value
func (ul *UniversalLinqDSL) SetContext(key string, value interface{}) {
	ul.dsl.SetContext(key, value)
}
