package dsls

import (
	"fmt"
	"time"
	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

// TemplateDSL defines the DSL for journal entry templates
type TemplateDSL struct {
	dsl *dslbuilder.DSL
}

// NewTemplateDSL creates a new template DSL instance
func NewTemplateDSL() *TemplateDSL {
	t := &TemplateDSL{
		dsl: dslbuilder.New("TemplateEngine"),
	}
	t.defineTokens()
	t.defineRules()
	t.registerFunctions()
	return t
}

func (t *TemplateDSL) defineTokens() {
	// Template structure tokens
	t.dsl.Token("TEMPLATE", `template`)
	t.dsl.Token("PARAMS", `params`)
	t.dsl.Token("ENTRY", `entry`)
	t.dsl.Token("LINE", `line`)
	t.dsl.Token("DEBIT", `debit`)
	t.dsl.Token("CREDIT", `credit`)
	t.dsl.Token("ACCOUNT", `account`)
	t.dsl.Token("AMOUNT", `amount`)
	t.dsl.Token("DESCRIPTION", `description`)
	t.dsl.Token("DATE", `date`)
	
	// Variables and literals
	t.dsl.Token("VARIABLE", `\$[a-zA-Z_][a-zA-Z0-9_]*`)
	t.dsl.Token("STRING", `"[^"]*"`)
	t.dsl.Token("NUMBER", `\d+(\.\d+)?`)
	t.dsl.Token("LPAREN", `\(`)
	t.dsl.Token("RPAREN", `\)`)
	
	// Operators
	t.dsl.Token("PLUS", `\+`)
	t.dsl.Token("MINUS", `-`)
	t.dsl.Token("MULTIPLY", `\*`)
	t.dsl.Token("DIVIDE", `/`)
}

func (t *TemplateDSL) defineRules() {
	// Template structure rules
	t.dsl.Rule("template", []string{"TEMPLATE", "name", "PARAMS", "param_list", "ENTRY", "entry_body"}, "createTemplate")
	t.dsl.Rule("entry_body", []string{"entry_metadata", "lines"}, "buildEntry")
	t.dsl.Rule("lines", []string{"line", "lines"}, "appendLine")
	t.dsl.Rule("lines", []string{"line"}, "singleLine")
	
	// Line rules
	t.dsl.Rule("line", []string{"LINE", "DEBIT", "ACCOUNT", "LPAREN", "STRING", "RPAREN", "AMOUNT", "LPAREN", "expression", "RPAREN", "DESCRIPTION", "LPAREN", "STRING", "RPAREN"}, "createDebitLine")
	t.dsl.Rule("line", []string{"LINE", "CREDIT", "ACCOUNT", "LPAREN", "STRING", "RPAREN", "AMOUNT", "LPAREN", "expression", "RPAREN", "DESCRIPTION", "LPAREN", "STRING", "RPAREN"}, "createCreditLine")
	
	// Expression rules
	t.dsl.Rule("expression", []string{"VARIABLE"}, "variableValue")
	t.dsl.Rule("expression", []string{"NUMBER"}, "numberValue")
	t.dsl.Rule("expression", []string{"expression", "MULTIPLY", "NUMBER"}, "multiply")
}

func (t *TemplateDSL) registerFunctions() {
	// Function to get last day of period
	t.dsl.Action("last_day", func(args []interface{}) (interface{}, error) {
		if len(args) == 0 {
			return time.Now().Format("2006-01-02"), nil
		}
		period := args[0].(string)
		// Simple implementation - in real version would parse period properly
		return fmt.Sprintf("%s-31", period), nil
	})
	
	// Function to get current date
	t.dsl.Action("current_date", func(args []interface{}) (interface{}, error) {
		return time.Now().Format("2006-01-02"), nil
	})
	
	// Function to format currency
	t.dsl.Action("format_currency", func(args []interface{}) (interface{}, error) {
		if len(args) == 0 {
			return "$0.00", nil
		}
		amount := args[0].(float64)
		return fmt.Sprintf("$%.2f", amount), nil
	})
}

// Parse executes the template DSL
func (t *TemplateDSL) Parse(input string, params map[string]interface{}) (interface{}, error) {
	return t.dsl.ParseWithContext(input, params)
}

// GetDSL returns the underlying DSL
func (t *TemplateDSL) GetDSL() *dslbuilder.DSL {
	return t.dsl
}