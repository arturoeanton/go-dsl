package dsls

import (
	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

// TaxDSL defines the DSL for tax calculations
type TaxDSL struct {
	dsl *dslbuilder.DSL
}

// NewTaxDSL creates a new tax calculation DSL
func NewTaxDSL() *TaxDSL {
	t := &TaxDSL{
		dsl: dslbuilder.New("TaxCalculation"),
	}
	t.defineTokens()
	t.defineRules()
	t.registerActions()
	return t
}

func (t *TaxDSL) defineTokens() {
	// Tax-specific tokens
	t.dsl.Token("CALCULATE", `calculate`)
	t.dsl.Token("TAX_FUNCTION", `(iva|retefuente|reteica|reteiva|impoconsumo)`)
	t.dsl.Token("BASE", `base`)
	t.dsl.Token("RATE", `rate`)
	t.dsl.Token("TO_ACCOUNT", `to_account`)
	t.dsl.Token("WHEN", `when`)
	t.dsl.Token("IF", `if`)
	t.dsl.Token("THEN", `then`)
	t.dsl.Token("ELSE", `else`)
	
	// Common tokens
	t.dsl.Token("NUMBER", `\d+(\.\d+)?`)
	t.dsl.Token("PERCENTAGE", `\d+(\.\d+)?%`)
	t.dsl.Token("ACCOUNT", `\d{4,8}`)
	t.dsl.Token("STRING", `"[^"]*"`)
	t.dsl.Token("OPERATOR", `(==|!=|>|<|>=|<=)`)
}

func (t *TaxDSL) defineRules() {
	t.dsl.Rule("tax_calculation", []string{"CALCULATE", "TAX_FUNCTION", "tax_params"}, "calculateTax")
	t.dsl.Rule("tax_params", []string{"BASE", "amount", "RATE", "rate_expr", "TO_ACCOUNT", "ACCOUNT"}, "setupTaxParams")
	t.dsl.Rule("rate_expr", []string{"PERCENTAGE"}, "fixedRate")
	t.dsl.Rule("rate_expr", []string{"IF", "condition", "THEN", "PERCENTAGE", "ELSE", "rate_expr"}, "conditionalRate")
}

func (t *TaxDSL) registerActions() {
	t.dsl.Action("calculateTax", func(args []interface{}) (interface{}, error) {
		taxType := args[1].(string)
		params := args[2].(map[string]interface{})
		
		// Implementation would calculate the specific tax
		return map[string]interface{}{
			"tax_type": taxType,
			"params": params,
		}, nil
	})
	
	t.dsl.Action("fixedRate", func(args []interface{}) (interface{}, error) {
		rate := args[0].(string)
		// Remove % and convert to decimal
		return rate, nil
	})
}

// Parse executes the tax calculation DSL
func (t *TaxDSL) Parse(input string, context map[string]interface{}) (interface{}, error) {
	return t.dsl.ParseWithContext(input, context)
}

// GetDSL returns the underlying DSL
func (t *TaxDSL) GetDSL() *dslbuilder.DSL {
	return t.dsl
}