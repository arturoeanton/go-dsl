package dsls

import (
	"fmt"
	"motor-contable-poc/internal/models"
	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

// ValidationDSL defines the DSL for voucher validation rules
type ValidationDSL struct {
	dsl *dslbuilder.DSL
}

// NewValidationDSL creates a new validation DSL instance
func NewValidationDSL() *ValidationDSL {
	v := &ValidationDSL{
		dsl: dslbuilder.New("VoucherValidation"),
	}
	v.defineTokens()
	v.defineRules()
	v.registerActions()
	return v
}

func (v *ValidationDSL) defineTokens() {
	// Basic tokens
	v.dsl.Token("NUMBER", `\d+(\.\d+)?`)
	v.dsl.Token("STRING", `"[^"]*"`)
	v.dsl.Token("ACCOUNT", `\d{4,8}`)
	v.dsl.Token("COMPARISON", `(>|<|>=|<=|==|!=)`)
	v.dsl.Token("LOGICAL", `(AND|OR)`)
	v.dsl.Token("IF", `if`)
	v.dsl.Token("THEN", `then`)
	v.dsl.Token("ERROR", `error`)
	v.dsl.Token("WARNING", `warning`)
	
	// Functions
	v.dsl.Token("FUNCTION", `(account_type|account_balance|voucher_total|third_party_type|user_role|period_status)`)
	v.dsl.Token("LPAREN", `\(`)
	v.dsl.Token("RPAREN", `\)`)
}

func (v *ValidationDSL) defineRules() {
	v.dsl.Rule("validation", []string{"IF", "condition", "THEN", "action"}, "executeValidation")
	v.dsl.Rule("condition", []string{"expression"}, "evaluateCondition")
	v.dsl.Rule("condition", []string{"expression", "LOGICAL", "condition"}, "combineConditions")
	v.dsl.Rule("expression", []string{"FUNCTION", "LPAREN", "STRING", "RPAREN", "COMPARISON", "value"}, "checkFunction")
	v.dsl.Rule("value", []string{"NUMBER"}, "numericValue")
	v.dsl.Rule("value", []string{"STRING"}, "stringValue")
	v.dsl.Rule("action", []string{"ERROR", "STRING"}, "raiseError")
	v.dsl.Rule("action", []string{"WARNING", "STRING"}, "raiseWarning")
}

func (v *ValidationDSL) registerActions() {
	v.dsl.Action("executeValidation", func(args []interface{}) (interface{}, error) {
		// Implementation would check condition and execute action
		return models.ValidationResult{
			Type: "INFO",
			Message: "Validation executed",
		}, nil
	})
	
	v.dsl.Action("checkFunction", func(args []interface{}) (interface{}, error) {
		funcName := args[0].(string)
		param := args[2].(string)
		comparison := args[4].(string)
		value := args[5]
		
		// Implementation would call the actual function
		return fmt.Sprintf("%s(%s) %s %v", funcName, param, comparison, value), nil
	})
}

// Parse executes the DSL validation rules
func (v *ValidationDSL) Parse(input string, context map[string]interface{}) (interface{}, error) {
	return v.dsl.ParseWithContext(input, context)
}

// GetDSL returns the underlying DSL for advanced usage
func (v *ValidationDSL) GetDSL() *dslbuilder.DSL {
	return v.dsl
}