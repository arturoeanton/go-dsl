package universal

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

// UniversalDroolsDSL creates a universal Drools-like DSL that works with any struct type
type UniversalDroolsDSL struct {
	dsl    *dslbuilder.DSL
	engine *UniversalRuleEngine
}

// NewUniversalDroolsDSL creates a new universal Drools DSL
func NewUniversalDroolsDSL() *UniversalDroolsDSL {
	dsl := dslbuilder.New("UniversalDroolsDSL")
	engine := NewRuleEngine()

	ud := &UniversalDroolsDSL{
		dsl:    dsl,
		engine: engine,
	}

	ud.setupTokens()
	ud.setupRules()
	ud.setupActions()

	return ud
}

// setupTokens defines all the tokens for the Drools DSL
func (ud *UniversalDroolsDSL) setupTokens() {
	// Rule structure keywords
	ud.dsl.KeywordToken("RULE", "rule")
	ud.dsl.KeywordToken("WHEN", "when")
	ud.dsl.KeywordToken("THEN", "then")
	ud.dsl.KeywordToken("END", "end")
	ud.dsl.KeywordToken("SALIENCE", "salience")

	// Condition keywords
	ud.dsl.KeywordToken("EXISTS", "exists")
	ud.dsl.KeywordToken("NOT", "not")
	ud.dsl.KeywordToken("AND", "and")
	ud.dsl.KeywordToken("OR", "or")

	// Operators (Spanish)
	ud.dsl.KeywordToken("ES", "es")
	ud.dsl.KeywordToken("MAYOR", "mayor")
	ud.dsl.KeywordToken("MENOR", "menor")
	ud.dsl.KeywordToken("CONTIENE", "contiene")
	ud.dsl.KeywordToken("COINCIDE", "coincide")

	// Operators (English)
	ud.dsl.KeywordToken("IS", "is")
	ud.dsl.KeywordToken("EQUALS", "equals")
	ud.dsl.KeywordToken("GREATER", "greater")
	ud.dsl.KeywordToken("LESS", "less")
	ud.dsl.KeywordToken("CONTAINS", "contains")
	ud.dsl.KeywordToken("MATCHES", "matches")

	// Actions (Spanish)
	ud.dsl.KeywordToken("MODIFICAR", "modificar")
	ud.dsl.KeywordToken("INSERTAR", "insertar")
	ud.dsl.KeywordToken("RETIRAR", "retirar")
	ud.dsl.KeywordToken("EJECUTAR", "ejecutar")
	ud.dsl.KeywordToken("ESTABLECER", "establecer")

	// Actions (English)
	ud.dsl.KeywordToken("MODIFY", "modify")
	ud.dsl.KeywordToken("INSERT", "insert")
	ud.dsl.KeywordToken("RETRACT", "retract")
	ud.dsl.KeywordToken("EXECUTE", "execute")
	ud.dsl.KeywordToken("SET", "set")

	// Connectors
	ud.dsl.KeywordToken("CON", "con") // with
	ud.dsl.KeywordToken("WITH", "with")
	ud.dsl.KeywordToken("TO", "to")
	ud.dsl.KeywordToken("A", "a") // to (Spanish)

	// Value tokens
	ud.dsl.Token("STRING", `"[^"]*"`)
	ud.dsl.Token("NUMBER", `[0-9]+\.?[0-9]*`)
	ud.dsl.Token("WORD", `[a-zA-Z][a-zA-Z0-9_]*`)
}

// setupRules defines grammar rules for Drools-like syntax
func (ud *UniversalDroolsDSL) setupRules() {
	// Rule definition with salience (Spanish)
	ud.dsl.Rule("rule_def", []string{"RULE", "STRING", "SALIENCE", "NUMBER", "WHEN", "WORD", "WORD", "ES", "STRING", "THEN", "ESTABLECER", "WORD", "A", "STRING", "END"}, "ruleWithSalienceSetSpanish")
	ud.dsl.Rule("rule_def", []string{"RULE", "STRING", "SALIENCE", "NUMBER", "WHEN", "WORD", "WORD", "MAYOR", "NUMBER", "THEN", "ESTABLECER", "WORD", "A", "STRING", "END"}, "ruleWithSalienceSetSpanish")
	ud.dsl.Rule("rule_def", []string{"RULE", "STRING", "SALIENCE", "NUMBER", "WHEN", "WORD", "WORD", "MENOR", "NUMBER", "THEN", "ESTABLECER", "WORD", "A", "STRING", "END"}, "ruleWithSalienceSetSpanish")

	// Rule definition with salience (English)
	ud.dsl.Rule("rule_def", []string{"RULE", "STRING", "SALIENCE", "NUMBER", "WHEN", "WORD", "WORD", "IS", "STRING", "THEN", "SET", "WORD", "TO", "STRING", "END"}, "ruleWithSalienceSetEnglish")
	ud.dsl.Rule("rule_def", []string{"RULE", "STRING", "SALIENCE", "NUMBER", "WHEN", "WORD", "WORD", "GREATER", "NUMBER", "THEN", "SET", "WORD", "TO", "STRING", "END"}, "ruleWithSalienceSetEnglish")
	ud.dsl.Rule("rule_def", []string{"RULE", "STRING", "SALIENCE", "NUMBER", "WHEN", "WORD", "WORD", "LESS", "NUMBER", "THEN", "SET", "WORD", "TO", "STRING", "END"}, "ruleWithSalienceSetEnglish")

	// Simple rules without salience (Spanish)
	ud.dsl.Rule("rule_def", []string{"RULE", "STRING", "WHEN", "WORD", "WORD", "ES", "STRING", "THEN", "ESTABLECER", "WORD", "A", "STRING", "END"}, "simpleRuleSetSpanish")
	ud.dsl.Rule("rule_def", []string{"RULE", "STRING", "WHEN", "WORD", "WORD", "MAYOR", "NUMBER", "THEN", "ESTABLECER", "WORD", "A", "STRING", "END"}, "simpleRuleSetSpanish")
	ud.dsl.Rule("rule_def", []string{"RULE", "STRING", "WHEN", "WORD", "WORD", "MENOR", "NUMBER", "THEN", "ESTABLECER", "WORD", "A", "STRING", "END"}, "simpleRuleSetSpanish")

	// Simple rules without salience (English)
	ud.dsl.Rule("rule_def", []string{"RULE", "STRING", "WHEN", "WORD", "WORD", "IS", "STRING", "THEN", "SET", "WORD", "TO", "STRING", "END"}, "simpleRuleSetEnglish")
	ud.dsl.Rule("rule_def", []string{"RULE", "STRING", "WHEN", "WORD", "WORD", "GREATER", "NUMBER", "THEN", "SET", "WORD", "TO", "STRING", "END"}, "simpleRuleSetEnglish")
	ud.dsl.Rule("rule_def", []string{"RULE", "STRING", "WHEN", "WORD", "WORD", "LESS", "NUMBER", "THEN", "SET", "WORD", "TO", "STRING", "END"}, "simpleRuleSetEnglish")

	// Modify actions (Spanish)
	ud.dsl.Rule("rule_def", []string{"RULE", "STRING", "WHEN", "WORD", "WORD", "ES", "STRING", "THEN", "MODIFICAR", "WORD", "ESTABLECER", "WORD", "A", "STRING", "END"}, "ruleModifySpanish")

	// Modify actions (English)
	ud.dsl.Rule("rule_def", []string{"RULE", "STRING", "WHEN", "WORD", "WORD", "IS", "STRING", "THEN", "MODIFY", "WORD", "SET", "WORD", "TO", "STRING", "END"}, "ruleModifyEnglish")

	// Insert actions (Spanish)
	ud.dsl.Rule("rule_def", []string{"RULE", "STRING", "WHEN", "WORD", "WORD", "ES", "STRING", "THEN", "INSERTAR", "WORD", "END"}, "ruleInsertSpanish")

	// Insert actions (English)
	ud.dsl.Rule("rule_def", []string{"RULE", "STRING", "WHEN", "WORD", "WORD", "IS", "STRING", "THEN", "INSERT", "WORD", "END"}, "ruleInsertEnglish")
}

// setupActions defines all the action handlers
func (ud *UniversalDroolsDSL) setupActions() {
	// Rule with salience and set action (Spanish)
	ud.dsl.Action("ruleWithSalienceSetSpanish", func(args []interface{}) (interface{}, error) {
		return ud.createSetRule(args, true, true) // withSalience=true, isSpanish=true
	})

	// Rule with salience and set action (English)
	ud.dsl.Action("ruleWithSalienceSetEnglish", func(args []interface{}) (interface{}, error) {
		return ud.createSetRule(args, true, false) // withSalience=true, isSpanish=false
	})

	// Simple rule with set action (Spanish)
	ud.dsl.Action("simpleRuleSetSpanish", func(args []interface{}) (interface{}, error) {
		return ud.createSetRule(args, false, true) // withSalience=false, isSpanish=true
	})

	// Simple rule with set action (English)
	ud.dsl.Action("simpleRuleSetEnglish", func(args []interface{}) (interface{}, error) {
		return ud.createSetRule(args, false, false) // withSalience=false, isSpanish=false
	})

	// Modify rule (Spanish)
	ud.dsl.Action("ruleModifySpanish", func(args []interface{}) (interface{}, error) {
		return ud.createModifyRule(args, true)
	})

	// Modify rule (English)
	ud.dsl.Action("ruleModifyEnglish", func(args []interface{}) (interface{}, error) {
		return ud.createModifyRule(args, false)
	})

	// Insert rule (Spanish)
	ud.dsl.Action("ruleInsertSpanish", func(args []interface{}) (interface{}, error) {
		return ud.createInsertRule(args, true)
	})

	// Insert rule (English)
	ud.dsl.Action("ruleInsertEnglish", func(args []interface{}) (interface{}, error) {
		return ud.createInsertRule(args, false)
	})
}

// createSetRule creates a rule with set action
func (ud *UniversalDroolsDSL) createSetRule(args []interface{}, withSalience bool, isSpanish bool) (interface{}, error) {
	var ruleName, entity, field, operator, value, targetField, newValue string
	var salience int

	if withSalience {
		if len(args) < 15 {
			return nil, fmt.Errorf("insufficient arguments for rule with salience")
		}
		ruleName = strings.Trim(args[1].(string), `"`)
		salienceStr := args[3].(string)
		var err error
		salience, err = strconv.Atoi(salienceStr)
		if err != nil {
			return nil, fmt.Errorf("invalid salience: %s", salienceStr)
		}
		entity = args[5].(string)
		field = args[6].(string)
		operator = args[7].(string)

		if strings.Contains(args[8].(string), `"`) {
			value = strings.Trim(args[8].(string), `"`)
		} else {
			value = args[8].(string)
		}

		targetField = args[11].(string)
		newValue = strings.Trim(args[13].(string), `"`)
	} else {
		if len(args) < 13 {
			return nil, fmt.Errorf("insufficient arguments for simple rule")
		}
		ruleName = strings.Trim(args[1].(string), `"`)
		salience = 0
		entity = args[3].(string)
		field = args[4].(string)
		operator = args[5].(string)

		if strings.Contains(args[6].(string), `"`) {
			value = strings.Trim(args[6].(string), `"`)
		} else {
			value = args[6].(string)
		}

		targetField = args[9].(string)
		newValue = strings.Trim(args[11].(string), `"`)
	}

	// Normalize operator
	normalizedOp := ud.normalizeOperator(operator)

	// Create rule
	rule := &Rule{
		Name:     ruleName,
		Priority: salience,
		Salience: salience,
		Conditions: []Condition{
			{
				Entity:   entity,
				Field:    field,
				Operator: normalizedOp,
				Value:    value,
			},
		},
		Actions: []Action{
			{
				Type:   "set",
				Target: entity,
				Field:  targetField,
				Value:  newValue,
			},
		},
	}

	ud.engine.AddRule(rule)
	return fmt.Sprintf("Rule '%s' created successfully", ruleName), nil
}

// createModifyRule creates a rule with modify action
func (ud *UniversalDroolsDSL) createModifyRule(args []interface{}, isSpanish bool) (interface{}, error) {
	if len(args) < 15 {
		return nil, fmt.Errorf("insufficient arguments for modify rule")
	}

	ruleName := strings.Trim(args[1].(string), `"`)
	entity := args[3].(string)
	field := args[4].(string)
	operator := args[5].(string)
	value := strings.Trim(args[6].(string), `"`)
	targetEntity := args[8].(string)
	targetField := args[10].(string)
	newValue := strings.Trim(args[12].(string), `"`)

	normalizedOp := ud.normalizeOperator(operator)

	rule := &Rule{
		Name:     ruleName,
		Priority: 0,
		Salience: 0,
		Conditions: []Condition{
			{
				Entity:   entity,
				Field:    field,
				Operator: normalizedOp,
				Value:    value,
			},
		},
		Actions: []Action{
			{
				Type:   "modify",
				Target: targetEntity,
				Field:  targetField,
				Value:  newValue,
			},
		},
	}

	ud.engine.AddRule(rule)
	return fmt.Sprintf("Modify rule '%s' created successfully", ruleName), nil
}

// createInsertRule creates a rule with insert action
func (ud *UniversalDroolsDSL) createInsertRule(args []interface{}, isSpanish bool) (interface{}, error) {
	if len(args) < 11 {
		return nil, fmt.Errorf("insufficient arguments for insert rule")
	}

	ruleName := strings.Trim(args[1].(string), `"`)
	entity := args[3].(string)
	field := args[4].(string)
	operator := args[5].(string)
	value := strings.Trim(args[6].(string), `"`)
	insertTarget := args[8].(string)

	normalizedOp := ud.normalizeOperator(operator)

	rule := &Rule{
		Name:     ruleName,
		Priority: 0,
		Salience: 0,
		Conditions: []Condition{
			{
				Entity:   entity,
				Field:    field,
				Operator: normalizedOp,
				Value:    value,
			},
		},
		Actions: []Action{
			{
				Type:   "insert",
				Target: insertTarget,
			},
		},
	}

	ud.engine.AddRule(rule)
	return fmt.Sprintf("Insert rule '%s' created successfully", ruleName), nil
}

// normalizeOperator converts operators to internal format
func (ud *UniversalDroolsDSL) normalizeOperator(operator string) string {
	switch strings.ToLower(operator) {
	case "es", "is", "equals":
		return "=="
	case "mayor", "greater":
		return ">"
	case "menor", "less":
		return "<"
	case "contiene", "contains":
		return "contains"
	case "coincide", "matches":
		return "matches"
	default:
		return operator
	}
}

// Parse executes a rule definition string
func (ud *UniversalDroolsDSL) Parse(rule string) (*dslbuilder.Result, error) {
	return ud.dsl.Parse(rule)
}

// Use executes a rule definition string with context
func (ud *UniversalDroolsDSL) Use(rule string, context map[string]interface{}) (*dslbuilder.Result, error) {
	return ud.dsl.Use(rule, context)
}

// GetEngine returns the underlying rule engine
func (ud *UniversalDroolsDSL) GetEngine() *UniversalRuleEngine {
	return ud.engine
}

// InsertFact inserts a fact into the rule engine
func (ud *UniversalDroolsDSL) InsertFact(fact interface{}) {
	ud.engine.InsertFact(fact)
}

// FireAllRules executes all matching rules
func (ud *UniversalDroolsDSL) FireAllRules() error {
	return ud.engine.FireAllRules()
}

// GetFacts returns all facts in working memory
func (ud *UniversalDroolsDSL) GetFacts() []interface{} {
	return ud.engine.GetFacts()
}
