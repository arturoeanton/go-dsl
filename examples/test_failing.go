package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/arturoeliasanton/go-dsl/pkg/dslbuilder"
)

func main() {
	fmt.Println("=== Testing Failing Commands ===")
	
	// Create DSL exactly like in contabilidad
	contabilidad := dslbuilder.New("Test-Failing")

	// Define tokens - KEYWORDS FIRST with high priority
	contabilidad.KeywordToken("VENTA", "venta")
	contabilidad.KeywordToken("DE", "de")
	contabilidad.KeywordToken("CON", "con")
	contabilidad.KeywordToken("IVA", "iva")

	// Values - simplified patterns (lower priority)
	contabilidad.Token("IMPORTE", "[0-9]+\\.?[0-9]*")

	// Grammar rules - MOST SPECIFIC rules first
	// Sales patterns (most specific first) 
	contabilidad.Rule("command", []string{"VENTA", "DE", "IMPORTE", "CON", "IVA"}, "saleWithTax")
	contabilidad.Rule("command", []string{"VENTA", "DE", "IMPORTE"}, "simpleSale")

	// Actions for sales
	contabilidad.Action("simpleSale", func(args []interface{}) (interface{}, error) {
		importeStr := args[2].(string)
		importe, _ := strconv.ParseFloat(importeStr, 64)
		return fmt.Sprintf("Venta simple por %.2f", importe), nil
	})

	contabilidad.Action("saleWithTax", func(args []interface{}) (interface{}, error) {
		importeStr := args[2].(string)
		importe, _ := strconv.ParseFloat(importeStr, 64)
		iva := importe * 0.16
		total := importe + iva
		return fmt.Sprintf("Venta con IVA por %.2f (total: %.2f)", importe, total), nil
	})

	// Test the failing commands
	testCommands := []string{
		"venta de 1000",
		"venta de 5000 con iva",
	}

	for i, cmd := range testCommands {
		fmt.Printf("%d. Testing: %s\n", i+1, cmd)
		result, err := contabilidad.Parse(cmd)
		if err != nil {
			log.Printf("   Error: %v\n", err)
		} else {
			fmt.Printf("   Result: %v\n", result.GetOutput())
		}
		fmt.Println()
	}
}