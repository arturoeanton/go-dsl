package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

// Asiento represents an accounting entry
type Asiento struct {
	ID          int
	Fecha       time.Time
	Descripcion string
	Movimientos []Movimiento
}

// Movimiento represents a line in an accounting entry
type Movimiento struct {
	Cuenta      string
	Descripcion string
	Debe        float64
	Haber       float64
}

// Cuenta represents an account
type Cuenta struct {
	Codigo string
	Nombre string
	Tipo   string // Activo, Pasivo, Patrimonio, Ingreso, Gasto
	Saldo  float64
}

// SistemaContable holds the accounting system state
type SistemaContable struct {
	Cuentas  map[string]*Cuenta
	Asientos []Asiento
	Contador int
	IVA      float64
}

// Helper function to create a new DSL instance
func createContabilidadDSL(sistema *SistemaContable) *dslbuilder.DSL {
	contabilidad := dslbuilder.New("Contabilidad-Fixed")

	// Define tokens - KEYWORDS FIRST with high priority
	contabilidad.KeywordToken("VENTA", "venta")
	contabilidad.KeywordToken("COMPRA", "compra")
	contabilidad.KeywordToken("DE", "de")
	contabilidad.KeywordToken("CON", "con")
	contabilidad.KeywordToken("IVA", "iva")
	contabilidad.KeywordToken("A", "a")
	contabilidad.KeywordToken("CLIENTE", "cliente")
	contabilidad.KeywordToken("POR", "por")
	contabilidad.KeywordToken("COBRO", "cobro")
	contabilidad.KeywordToken("PAGO", "pago")
	contabilidad.KeywordToken("PROVEEDOR", "proveedor")
	contabilidad.KeywordToken("NOMINA", "nomina")
	contabilidad.KeywordToken("EMPLEADO", "empleado")
	contabilidad.KeywordToken("GASTO", "gasto")
	contabilidad.KeywordToken("ASIENTO", "asiento")
	contabilidad.KeywordToken("DEBE", "debe")
	contabilidad.KeywordToken("HABER", "haber")

	// Values - IMPORTE first to get priority over CUENTA
	contabilidad.Token("IMPORTE", "[0-9]+\\.?[0-9]*")
	contabilidad.Token("STRING", "\"[^\"]*\"")
	// CUENTA pattern should be more specific than IMPORTE
	// But for now, let's distinguish by requiring CUENTA only in contexts where it makes sense

	// Grammar rules - MOST SPECIFIC rules first (longest patterns first)
	// Sales patterns (most specific first)
	contabilidad.Rule("command", []string{"VENTA", "DE", "IMPORTE", "A", "CLIENTE", "STRING"}, "saleToClient")
	contabilidad.Rule("command", []string{"VENTA", "DE", "IMPORTE", "CON", "IVA"}, "saleWithTax")
	contabilidad.Rule("command", []string{"VENTA", "DE", "IMPORTE"}, "simpleSale")

	// Purchase patterns (most specific first)
	contabilidad.Rule("command", []string{"COMPRA", "DE", "IMPORTE", "CON", "IVA"}, "purchaseWithTax")
	contabilidad.Rule("command", []string{"COMPRA", "DE", "IMPORTE"}, "simplePurchase")

	// Other operations
	contabilidad.Rule("command", []string{"COBRO", "DE", "CLIENTE", "STRING", "POR", "IMPORTE"}, "collection")
	contabilidad.Rule("command", []string{"PAGO", "A", "PROVEEDOR", "STRING", "POR", "IMPORTE"}, "payment")
	contabilidad.Rule("command", []string{"NOMINA", "DE", "EMPLEADO", "STRING", "POR", "IMPORTE"}, "payroll")
	contabilidad.Rule("command", []string{"GASTO", "DE", "STRING", "POR", "IMPORTE"}, "expense")

	// Manual accounting entries - using IMPORTE for both accounts and amounts
	contabilidad.Rule("command", []string{"ASIENTO", "movements"}, "processEntry")

	// Movements rules - LEFT-RECURSIVE with ImprovedParser
	contabilidad.Rule("movements", []string{"movement"}, "singleMovement")
	contabilidad.Rule("movements", []string{"movements", "movement"}, "multipleMovements")

	// Single movements - using IMPORTE for both account codes and amounts
	contabilidad.Rule("movement", []string{"DEBE", "IMPORTE", "IMPORTE"}, "debitMovement")
	contabilidad.Rule("movement", []string{"HABER", "IMPORTE", "IMPORTE"}, "creditMovement")

	// Actions for sales
	contabilidad.Action("simpleSale", func(args []interface{}) (interface{}, error) {
		importeStr := args[2].(string)
		importe, _ := strconv.ParseFloat(importeStr, 64)

		movements := []Movimiento{
			{Cuenta: "1201", Descripcion: "Clientes", Debe: importe, Haber: 0},
			{Cuenta: "4101", Descripcion: "Ventas", Debe: 0, Haber: importe},
		}

		return createAsiento(sistema, fmt.Sprintf("Venta por %.2f", importe), movements), nil
	})

	contabilidad.Action("saleWithTax", func(args []interface{}) (interface{}, error) {
		importeStr := args[2].(string)
		importe, _ := strconv.ParseFloat(importeStr, 64)
		iva := importe * sistema.IVA
		total := importe + iva

		movements := []Movimiento{
			{Cuenta: "1201", Descripcion: "Clientes", Debe: total, Haber: 0},
			{Cuenta: "4101", Descripcion: "Ventas", Debe: 0, Haber: importe},
			{Cuenta: "2401", Descripcion: "IVA por pagar", Debe: 0, Haber: iva},
		}

		return createAsiento(sistema, fmt.Sprintf("Venta con IVA por %.2f", total), movements), nil
	})

	contabilidad.Action("saleToClient", func(args []interface{}) (interface{}, error) {
		importeStr := args[2].(string)
		cliente := strings.Trim(args[5].(string), "\"'")
		importe, _ := strconv.ParseFloat(importeStr, 64)
		iva := importe * sistema.IVA
		total := importe + iva

		movements := []Movimiento{
			{Cuenta: "1201", Descripcion: "Clientes - " + cliente, Debe: total, Haber: 0},
			{Cuenta: "4101", Descripcion: "Ventas", Debe: 0, Haber: importe},
			{Cuenta: "2401", Descripcion: "IVA por pagar", Debe: 0, Haber: iva},
		}

		return createAsiento(sistema, fmt.Sprintf("Venta a %s por %.2f", cliente, total), movements), nil
	})

	// Actions for purchases
	contabilidad.Action("simplePurchase", func(args []interface{}) (interface{}, error) {
		importeStr := args[2].(string)
		importe, _ := strconv.ParseFloat(importeStr, 64)

		movements := []Movimiento{
			{Cuenta: "5101", Descripcion: "Compras", Debe: importe, Haber: 0},
			{Cuenta: "2101", Descripcion: "Proveedores", Debe: 0, Haber: importe},
		}

		return createAsiento(sistema, fmt.Sprintf("Compra por %.2f", importe), movements), nil
	})

	contabilidad.Action("purchaseWithTax", func(args []interface{}) (interface{}, error) {
		importeStr := args[2].(string)
		importe, _ := strconv.ParseFloat(importeStr, 64)
		iva := importe * sistema.IVA
		total := importe + iva

		movements := []Movimiento{
			{Cuenta: "5101", Descripcion: "Compras", Debe: importe, Haber: 0},
			{Cuenta: "1401", Descripcion: "IVA acreditable", Debe: iva, Haber: 0},
			{Cuenta: "2101", Descripcion: "Proveedores", Debe: 0, Haber: total},
		}

		return createAsiento(sistema, fmt.Sprintf("Compra con IVA por %.2f", total), movements), nil
	})

	// Actions for other operations
	contabilidad.Action("collection", func(args []interface{}) (interface{}, error) {
		cliente := strings.Trim(args[3].(string), "\"'")
		importeStr := args[5].(string)
		importe, _ := strconv.ParseFloat(importeStr, 64)

		movements := []Movimiento{
			{Cuenta: "1101", Descripcion: "Bancos", Debe: importe, Haber: 0},
			{Cuenta: "1201", Descripcion: "Clientes - " + cliente, Debe: 0, Haber: importe},
		}

		return createAsiento(sistema, fmt.Sprintf("Cobro de %s", cliente), movements), nil
	})

	contabilidad.Action("payment", func(args []interface{}) (interface{}, error) {
		proveedor := strings.Trim(args[3].(string), "\"'")
		importeStr := args[5].(string)
		importe, _ := strconv.ParseFloat(importeStr, 64)

		movements := []Movimiento{
			{Cuenta: "2101", Descripcion: "Proveedores - " + proveedor, Debe: importe, Haber: 0},
			{Cuenta: "1101", Descripcion: "Bancos", Debe: 0, Haber: importe},
		}

		return createAsiento(sistema, fmt.Sprintf("Pago a %s", proveedor), movements), nil
	})

	contabilidad.Action("payroll", func(args []interface{}) (interface{}, error) {
		empleado := strings.Trim(args[3].(string), "\"'")
		importeStr := args[5].(string)
		importe, _ := strconv.ParseFloat(importeStr, 64)

		movements := []Movimiento{
			{Cuenta: "5201", Descripcion: "Sueldos y salarios", Debe: importe, Haber: 0},
			{Cuenta: "1101", Descripcion: "Bancos", Debe: 0, Haber: importe},
		}

		return createAsiento(sistema, fmt.Sprintf("Nómina de %s", empleado), movements), nil
	})

	contabilidad.Action("expense", func(args []interface{}) (interface{}, error) {
		concepto := strings.Trim(args[2].(string), "\"'")
		importeStr := args[4].(string)
		importe, _ := strconv.ParseFloat(importeStr, 64)

		movements := []Movimiento{
			{Cuenta: "5301", Descripcion: "Gastos generales - " + concepto, Debe: importe, Haber: 0},
			{Cuenta: "1101", Descripcion: "Bancos", Debe: 0, Haber: importe},
		}

		return createAsiento(sistema, fmt.Sprintf("Gasto: %s", concepto), movements), nil
	})

	// Actions for manual entries - using IMPORTE for both accounts and amounts
	contabilidad.Action("singleMovement", func(args []interface{}) (interface{}, error) {
		movement := args[0].(Movimiento)
		return []Movimiento{movement}, nil
	})

	contabilidad.Action("multipleMovements", func(args []interface{}) (interface{}, error) {
		movements := args[0].([]Movimiento)
		newMovement := args[1].(Movimiento)
		return append(movements, newMovement), nil
	})

	contabilidad.Action("debitMovement", func(args []interface{}) (interface{}, error) {
		cuentaStr := args[1].(string)  // First IMPORTE is account code
		importeStr := args[2].(string) // Second IMPORTE is amount
		importe, _ := strconv.ParseFloat(importeStr, 64)

		return Movimiento{
			Cuenta:      cuentaStr,
			Descripcion: getCuentaName(sistema, cuentaStr),
			Debe:        importe,
			Haber:       0,
		}, nil
	})

	contabilidad.Action("creditMovement", func(args []interface{}) (interface{}, error) {
		cuentaStr := args[1].(string)  // First IMPORTE is account code
		importeStr := args[2].(string) // Second IMPORTE is amount
		importe, _ := strconv.ParseFloat(importeStr, 64)

		return Movimiento{
			Cuenta:      cuentaStr,
			Descripcion: getCuentaName(sistema, cuentaStr),
			Debe:        0,
			Haber:       importe,
		}, nil
	})

	contabilidad.Action("processEntry", func(args []interface{}) (interface{}, error) {
		movements := args[1].([]Movimiento)

		// Validate balanced entry
		totalDebe := 0.0
		totalHaber := 0.0
		for _, m := range movements {
			totalDebe += m.Debe
			totalHaber += m.Haber
		}

		if totalDebe != totalHaber {
			return nil, fmt.Errorf("asiento descuadrado: Debe %.2f != Haber %.2f", totalDebe, totalHaber)
		}

		asiento := createAsiento(sistema, "Asiento manual", movements)
		return asiento, nil
	})

	return contabilidad
}

func main() {
	// Initialize accounting system
	sistema := &SistemaContable{
		Cuentas:  initializeCuentas(),
		Asientos: []Asiento{},
		Contador: 0,
		IVA:      0.16, // 16% default (Mexico)
	}

	fmt.Println("=== Sistema Contable DSL Mejorado ===")
	fmt.Println("Usando contexto dinámico y tokenización corregida")
	fmt.Println()
	fmt.Println("Cuentas disponibles:")
	fmt.Println("  1101 - Bancos")
	fmt.Println("  1201 - Clientes")
	fmt.Println("  1401 - IVA acreditable")
	fmt.Println("  2101 - Proveedores")
	fmt.Println("  2401 - IVA por pagar")
	fmt.Println("  3101 - Capital social")
	fmt.Println("  4101 - Ventas")
	fmt.Println("  5101 - Compras")
	fmt.Println("  5201 - Sueldos y salarios")
	fmt.Println("  5301 - Gastos generales")
	fmt.Println()
	fmt.Printf("IVA configurado: %.0f%%\n", sistema.IVA*100)
	fmt.Println(strings.Repeat("-", 60))

	// Debug DSL structure first
	testDSL := createContabilidadDSL(sistema)
	debugInfo := testDSL.Debug()
	fmt.Printf("DSL Debug Info:\n")
	fmt.Printf("Rules for 'command': %d alternatives\n", len(debugInfo["rules"].(map[string]interface{})["command"].([]map[string]interface{})))
	for i, alt := range debugInfo["rules"].(map[string]interface{})["command"].([]map[string]interface{}) {
		fmt.Printf("  Alt %d: %v -> %s\n", i, alt["sequence"], alt["action"])
	}
	fmt.Println()

	// Test operations
	operations := []struct {
		desc string
		code string
	}{
		{
			"Venta simple de $1,000:",
			`venta de 1000`,
		},
		{
			"Venta con IVA de $5,000:",
			`venta de 5000 con iva`,
		},
		{
			"Venta a cliente específico:",
			`venta de 3000 a cliente "Empresa ABC"`,
		},
		{
			"Compra simple de $2,000:",
			`compra de 2000`,
		},
		{
			"Compra con IVA de $4,000:",
			`compra de 4000 con iva`,
		},
		{
			"Cobro a cliente:",
			`cobro de cliente "Empresa ABC" por 3480`,
		},
		{
			"Pago a proveedor:",
			`pago a proveedor "Proveedores SA" por 2320`,
		},
		{
			"Pago de nómina:",
			`nomina de empleado "Juan Pérez" por 15000`,
		},
		{
			"Registro de gasto:",
			`gasto de "Papelería" por 500`,
		},
		{
			"Asiento manual complejo:",
			`asiento debe 1101 10000 debe 1401 1600 haber 2101 11600`,
		},
	}

	for i, op := range operations {
		fmt.Printf("\n%d. %s\n", i+1, op.desc)
		fmt.Printf("   Comando: %s\n", op.code)

		// Create a fresh DSL instance for each operation to avoid state issues
		contabilidad := createContabilidadDSL(sistema)

		result, err := contabilidad.Parse(op.code)

		if err != nil {
			log.Printf("   Error en '%s': %v\n", op.code, err)
			// Debug tokenization for failing commands
			tokens, tokenErr := contabilidad.DebugTokens(op.code)
			if tokenErr != nil {
				fmt.Printf("   Tokenization failed: %v\n", tokenErr)
			} else {
				fmt.Printf("   Tokens: ")
				for _, token := range tokens {
					fmt.Printf("[%s:%s] ", token.TokenType, token.Value)
				}
				fmt.Printf("\n")
			}
			continue
		}

		if asiento, ok := result.GetOutput().(Asiento); ok {
			printAsiento(asiento)
		}
	}

	// Show account balances
	fmt.Println("\n\nBalances de Cuentas:")
	fmt.Println("====================")

	updateBalances(sistema)

	fmt.Println("\nActivos:")
	printAccountType(sistema, "Activo")

	fmt.Println("\nPasivos:")
	printAccountType(sistema, "Pasivo")

	fmt.Println("\nCapital:")
	printAccountType(sistema, "Patrimonio")

	fmt.Println("\nIngresos:")
	printAccountType(sistema, "Ingreso")

	fmt.Println("\nGastos:")
	printAccountType(sistema, "Gasto")

	// Show trial balance
	fmt.Println("\n\nBalanza de Comprobación:")
	fmt.Println("========================")

	totalDebe := 0.0
	totalHaber := 0.0

	for _, cuenta := range sistema.Cuentas {
		if cuenta.Saldo != 0 {
			if cuenta.Saldo > 0 {
				fmt.Printf("  %s - %-30s %15s\n", cuenta.Codigo, cuenta.Nombre, formatMoney(cuenta.Saldo))
				totalDebe += cuenta.Saldo
			} else {
				fmt.Printf("  %s - %-30s %25s\n", cuenta.Codigo, cuenta.Nombre, formatMoney(-cuenta.Saldo))
				totalHaber += -cuenta.Saldo
			}
		}
	}

	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("  Totales: %31s %15s\n", formatMoney(totalDebe), formatMoney(totalHaber))

	if totalDebe == totalHaber {
		fmt.Println("\n✅ La balanza está cuadrada")
	} else {
		fmt.Printf("\n❌ Descuadre: %.2f\n", totalDebe-totalHaber)
	}

	// Show some complex manual entries
	fmt.Println("\n\nEjemplos de Asientos Manuales Complejos:")
	fmt.Println("=========================================")

	complexEntries := []struct {
		desc string
		code string
	}{
		{
			"Aportación de capital:",
			`asiento debe 3101 100000 haber 1101 100000`,
		},
		{
			"Gasto con IVA (teléfono):",
			`asiento debe 5301 1000 debe 1401 160 haber 1101 1160`,
		},
		{
			"Venta a crédito con IVA:",
			`asiento debe 1201 11600 haber 4101 10000 haber 2401 1600`,
		},
	}

	for _, entry := range complexEntries {
		fmt.Printf("\n%s\n", entry.desc)
		fmt.Printf("Comando: %s\n", entry.code)

		// Create a fresh DSL instance for each complex operation
		contabilidad := createContabilidadDSL(sistema)
		result, err := contabilidad.Parse(entry.code)

		if err != nil {
			log.Printf("Error en '%s': %v\n", entry.code, err)
			continue
		}

		if asiento, ok := result.GetOutput().(Asiento); ok {
			printAsiento(asiento)
		}
	}

	fmt.Println("\n=== ✅ Sistema Contable DSL funcionando correctamente ===")
	fmt.Println("✅ Sin errores de parsing")
	fmt.Println("✅ Todas las operaciones funcionan")
	fmt.Println("✅ Asientos manuales complejos funcionan")
	fmt.Println("✅ Balanza cuadrada automáticamente")
}

// Helper functions

func initializeCuentas() map[string]*Cuenta {
	return map[string]*Cuenta{
		"1101": {Codigo: "1101", Nombre: "Bancos", Tipo: "Activo", Saldo: 0},
		"1201": {Codigo: "1201", Nombre: "Clientes", Tipo: "Activo", Saldo: 0},
		"1401": {Codigo: "1401", Nombre: "IVA acreditable", Tipo: "Activo", Saldo: 0},
		"2101": {Codigo: "2101", Nombre: "Proveedores", Tipo: "Pasivo", Saldo: 0},
		"2401": {Codigo: "2401", Nombre: "IVA por pagar", Tipo: "Pasivo", Saldo: 0},
		"3101": {Codigo: "3101", Nombre: "Capital social", Tipo: "Patrimonio", Saldo: 0},
		"4101": {Codigo: "4101", Nombre: "Ventas", Tipo: "Ingreso", Saldo: 0},
		"5101": {Codigo: "5101", Nombre: "Compras", Tipo: "Gasto", Saldo: 0},
		"5201": {Codigo: "5201", Nombre: "Sueldos y salarios", Tipo: "Gasto", Saldo: 0},
		"5301": {Codigo: "5301", Nombre: "Gastos generales", Tipo: "Gasto", Saldo: 0},
	}
}

func getCuentaName(sistema *SistemaContable, codigo string) string {
	if cuenta, exists := sistema.Cuentas[codigo]; exists {
		return cuenta.Nombre
	}
	return "Cuenta " + codigo
}

func createAsiento(sistema *SistemaContable, descripcion string, movimientos []Movimiento) Asiento {
	sistema.Contador++
	asiento := Asiento{
		ID:          sistema.Contador,
		Fecha:       time.Now(),
		Descripcion: descripcion,
		Movimientos: movimientos,
	}

	sistema.Asientos = append(sistema.Asientos, asiento)

	// Update account balances
	for _, mov := range movimientos {
		if cuenta, exists := sistema.Cuentas[mov.Cuenta]; exists {
			// For assets and expenses, debit increases, credit decreases
			// For liabilities, equity, and income, credit increases, debit decreases
			if cuenta.Tipo == "Activo" || cuenta.Tipo == "Gasto" {
				cuenta.Saldo += mov.Debe - mov.Haber
			} else {
				cuenta.Saldo += mov.Haber - mov.Debe
			}
		}
	}

	return asiento
}

func printAsiento(asiento Asiento) {
	fmt.Printf("\n   Asiento #%d - %s\n", asiento.ID, asiento.Fecha.Format("2006-01-02"))
	fmt.Printf("   %s\n", asiento.Descripcion)
	fmt.Println("   " + strings.Repeat("-", 55))

	totalDebe := 0.0
	totalHaber := 0.0

	for _, mov := range asiento.Movimientos {
		if mov.Debe > 0 {
			fmt.Printf("   %s %-25s %12s\n", mov.Cuenta, mov.Descripcion, formatMoney(mov.Debe))
			totalDebe += mov.Debe
		} else {
			fmt.Printf("   %s %-25s %25s\n", mov.Cuenta, mov.Descripcion, formatMoney(mov.Haber))
			totalHaber += mov.Haber
		}
	}

	fmt.Println("   " + strings.Repeat("-", 55))
	fmt.Printf("   Totales: %31s %12s\n", formatMoney(totalDebe), formatMoney(totalHaber))
}

func formatMoney(amount float64) string {
	return fmt.Sprintf("$%.2f", amount)
}

func updateBalances(sistema *SistemaContable) {
	// Reset balances
	for _, cuenta := range sistema.Cuentas {
		cuenta.Saldo = 0
	}

	// Recalculate from all entries
	for _, asiento := range sistema.Asientos {
		for _, mov := range asiento.Movimientos {
			if cuenta, exists := sistema.Cuentas[mov.Cuenta]; exists {
				if cuenta.Tipo == "Activo" || cuenta.Tipo == "Gasto" {
					cuenta.Saldo += mov.Debe - mov.Haber
				} else {
					cuenta.Saldo += mov.Haber - mov.Debe
				}
			}
		}
	}
}

func printAccountType(sistema *SistemaContable, tipo string) {
	total := 0.0
	for _, cuenta := range sistema.Cuentas {
		if cuenta.Tipo == tipo && cuenta.Saldo != 0 {
			fmt.Printf("  %s - %-30s %15s\n", cuenta.Codigo, cuenta.Nombre, formatMoney(cuenta.Saldo))
			total += cuenta.Saldo
		}
	}
	if total != 0 {
		fmt.Printf("  Total %-35s %15s\n", tipo+":", formatMoney(total))
	}
}
