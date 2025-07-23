package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

// Transaction represents an accounting transaction
type Transaction struct {
	Type        string
	Amount      float64
	Description string
	Tax         float64
	Total       float64
}

func main() {
	// Create an accounting DSL
	accounting := dslbuilder.New("Accounting")

	// Define tokens - KEYWORDS first with high priority
	accounting.KeywordToken("ASIENTO", "asiento")
	accounting.KeywordToken("REGISTRAR", "registrar")
	accounting.KeywordToken("CREAR", "crear")
	accounting.KeywordToken("VENTA", "venta")
	accounting.KeywordToken("COMPRA", "compra")
	accounting.KeywordToken("DE", "de")
	accounting.KeywordToken("POR", "por")
	accounting.KeywordToken("CON", "con")
	accounting.KeywordToken("DESCRIPCION", "descripcion")
	
	// Values - lower priority
	accounting.Token("AMOUNT", "[0-9]+\\.?[0-9]*")
	accounting.Token("STRING", "\"[^\"]*\"")

	// Define grammar rules - MOST SPECIFIC first
	accounting.Rule("transaction", []string{"REGISTRAR", "VENTA", "DE", "AMOUNT", "CON", "DESCRIPCION", "STRING"}, "fullTransaction")
	accounting.Rule("transaction", []string{"REGISTRAR", "COMPRA", "DE", "AMOUNT", "CON", "DESCRIPCION", "STRING"}, "fullTransaction")
	accounting.Rule("transaction", []string{"CREAR", "VENTA", "DE", "AMOUNT", "CON", "DESCRIPCION", "STRING"}, "fullTransaction")
	accounting.Rule("transaction", []string{"CREAR", "COMPRA", "DE", "AMOUNT", "CON", "DESCRIPCION", "STRING"}, "fullTransaction")
	accounting.Rule("transaction", []string{"ASIENTO", "VENTA", "DE", "AMOUNT", "CON", "DESCRIPCION", "STRING"}, "fullTransaction")
	accounting.Rule("transaction", []string{"ASIENTO", "COMPRA", "DE", "AMOUNT", "CON", "DESCRIPCION", "STRING"}, "fullTransaction")
	
	// Simple rules (shorter patterns last)
	accounting.Rule("transaction", []string{"REGISTRAR", "VENTA", "DE", "AMOUNT"}, "simpleTransaction")
	accounting.Rule("transaction", []string{"REGISTRAR", "COMPRA", "DE", "AMOUNT"}, "simpleTransaction")
	accounting.Rule("transaction", []string{"CREAR", "VENTA", "DE", "AMOUNT"}, "simpleTransaction")
	accounting.Rule("transaction", []string{"CREAR", "COMPRA", "DE", "AMOUNT"}, "simpleTransaction")
	accounting.Rule("transaction", []string{"ASIENTO", "VENTA", "DE", "AMOUNT"}, "simpleTransaction")
	accounting.Rule("transaction", []string{"ASIENTO", "COMPRA", "DE", "AMOUNT"}, "simpleTransaction")

	// Create a context to store transactions
	transactions := []Transaction{}

	// Register Go functions that can be called from DSL
	accounting.Set("calcularIVA", func(amount float64, country string) float64 {
		taxRates := map[string]float64{
			"MX":  0.16,
			"COL": 0.19,
			"AR":  0.21,
			"PE":  0.18,
		}

		rate, exists := taxRates[country]
		if !exists {
			rate = 0.16 // Default
		}

		return amount * rate
	})

	accounting.Set("formatMoney", func(amount float64) string {
		return fmt.Sprintf("$%.2f", amount)
	})

	// Define actions
	accounting.Action("simpleTransaction", func(args []interface{}) (interface{}, error) {
		if len(args) >= 4 {
			// args[0] = action (registrar/crear/asiento)
			// args[1] = type (venta/compra)  
			// args[2] = "de"
			// args[3] = amount
			transType := args[1].(string)
			amountStr := args[3].(string)
			amount, _ := strconv.ParseFloat(amountStr, 64)

			// Get current country from context
			country, _ := accounting.GetContext("country").(string)
			if country == "" {
				country = "MX"
			}

			// Calculate tax using registered function
			calcIVA, _ := accounting.Get("calcularIVA")
			taxFn := calcIVA.(func(float64, string) float64)
			tax := taxFn(amount, country)

			trans := Transaction{
				Type:        transType,
				Amount:      amount,
				Description: fmt.Sprintf("Transacción de %s", transType),
				Tax:         tax,
				Total:       amount + tax,
			}

			transactions = append(transactions, trans)

			return trans, nil
		}
		return nil, fmt.Errorf("invalid transaction")
	})

	accounting.Action("fullTransaction", func(args []interface{}) (interface{}, error) {
		if len(args) >= 7 {
			// args[0] = action (registrar/crear/asiento)
			// args[1] = type (venta/compra)
			// args[2] = "de"  
			// args[3] = amount
			// args[4] = "con"
			// args[5] = "descripcion"
			// args[6] = string
			transType := args[1].(string)
			amountStr := args[3].(string)
			description := strings.Trim(args[6].(string), "\"")
			amount, _ := strconv.ParseFloat(amountStr, 64)

			// Get current country from context
			country, _ := accounting.GetContext("country").(string)
			if country == "" {
				country = "MX"
			}

			// Calculate tax using registered function
			calcIVA, _ := accounting.Get("calcularIVA")
			taxFn := calcIVA.(func(float64, string) float64)
			tax := taxFn(amount, country)

			trans := Transaction{
				Type:        transType,
				Amount:      amount,
				Description: description,
				Tax:         tax,
				Total:       amount + tax,
			}

			transactions = append(transactions, trans)

			return trans, nil
		}
		return nil, fmt.Errorf("invalid transaction")
	})

	fmt.Println("Accounting DSL Demo")
	fmt.Println("===================\n")

	// Test DSL with different contexts
	testCases := []struct {
		country string
		code    string
	}{
		{"MX", `registrar venta de 1000`},
		{"MX", `crear venta de 5000 con descripcion "Venta de laptops"`},
		{"COL", `asiento compra de 3000`},
		{"AR", `registrar venta de 10000 con descripcion "Servicios de consultoría"`},
		{"PE", `crear compra de 2500 con descripcion "Materiales de oficina"`},
	}

	for _, tc := range testCases {
		// Use DSL with context
		ctx := map[string]interface{}{
			"country": tc.country,
		}

		result, err := accounting.Use(tc.code, ctx)
		if err != nil {
			log.Printf("Error: %v\n", err)
			continue
		}

		if trans, ok := result.GetOutput().(Transaction); ok {
			formatFn, _ := accounting.Get("formatMoney")
			format := formatFn.(func(float64) string)

			fmt.Printf("País: %s\n", tc.country)
			fmt.Printf("Comando: %s\n", tc.code)
			fmt.Printf("Resultado:\n")
			fmt.Printf("  Tipo: %s\n", trans.Type)
			fmt.Printf("  Monto: %s\n", format(trans.Amount))
			fmt.Printf("  IVA: %s\n", format(trans.Tax))
			fmt.Printf("  Total: %s\n", format(trans.Total))
			fmt.Printf("  Descripción: %s\n", trans.Description)
			fmt.Println()
		}
	}

	// Show all transactions
	fmt.Println("\nResumen de Transacciones:")
	fmt.Println("========================")

	formatFn, _ := accounting.Get("formatMoney")
	format := formatFn.(func(float64) string)

	totalAmount := 0.0
	totalTax := 0.0

	for i, trans := range transactions {
		fmt.Printf("%d. %s - %s (IVA: %s) = %s\n",
			i+1,
			trans.Type,
			format(trans.Amount),
			format(trans.Tax),
			format(trans.Total))
		totalAmount += trans.Amount
		totalTax += trans.Tax
	}

	fmt.Printf("\nTotal: %s + IVA %s = %s\n",
		format(totalAmount),
		format(totalTax),
		format(totalAmount+totalTax))
}
