# Accounting System DSL - Complete Example

**A complete enterprise-level accounting system built with go-dsl demonstrating all advanced framework features.**

## 🎯 Objective

This example demonstrates how to create an **enterprise-level accounting system** using go-dsl, including:

- ✅ Left-recursive grammars for complex accounting entries
- ✅ KeywordToken to resolve tokenization conflicts
- ✅ Automatic accounting balance validation
- ✅ Tax calculations and transactions with taxes
- ✅ Complete chart of accounts system
- ✅ Complex manual and balanced entries
- ✅ Production-ready stability (no intermittent errors)

## 🚀 Quick Start

```bash
cd examples/contabilidad
go run main.go
```

## 📚 DSL Features

### Defined Tokens

```go
// Keywords with high priority (90)
contabilidad.KeywordToken("VENTA", "venta")     // sale
contabilidad.KeywordToken("COMPRA", "compra")   // purchase
contabilidad.KeywordToken("DE", "de")           // of/from
contabilidad.KeywordToken("CON", "con")         // with
contabilidad.KeywordToken("IVA", "iva")         // VAT/tax
contabilidad.KeywordToken("ASIENTO", "asiento") // entry
contabilidad.KeywordToken("DEBE", "debe")       // debit
contabilidad.KeywordToken("HABER", "haber")     // credit

// Values with normal priority (0)
contabilidad.Token("IMPORTE", "[0-9]+\\.?[0-9]*") // amount
contabilidad.Token("STRING", "\"[^\"]*\"")         // string
```

### Supported Commands

#### 1. Basic Sales Operations
```
venta de 1000                    # Simple sale
venta de 5000 con iva           # Sale with VAT (16%)
venta de 3000 a cliente "ABC"   # Sale to specific client
```

#### 2. Purchase Operations
```
compra de 2000                  # Simple purchase
compra de 4000 con iva         # Purchase with creditable VAT
```

#### 3. Treasury Operations
```
cobro de cliente "ABC" por 3480      # Client collection
pago a proveedor "XYZ" por 2320      # Supplier payment
nomina de empleado "Juan" por 15000  # Payroll payment
gasto de "Papelería" por 500         # Expense recording
```

#### 4. Complex Manual Entries (Left Recursion)
```
asiento debe 1101 10000 debe 1401 1600 haber 2101 11600
```

**This left-recursive grammar was impossible before!** Now it works perfectly thanks to ImprovedParser with memoization.

## 🏗️ System Architecture

### Data Types

```go
type Asiento struct { // AccountingEntry
    ID          int
    Fecha       time.Time  // Date
    Descripcion string     // Description
    Movimientos []Movimiento // Movements
}

type Movimiento struct { // Movement
    Cuenta      string  // Account
    Descripcion string  // Description
    Debe        float64 // Debit
    Haber       float64 // Credit
}

type SistemaContable struct { // AccountingSystem
    Cuentas  map[string]*Cuenta // Accounts
    Asientos []Asiento          // Entries
    Contador int                // Counter
    IVA      float64            // VAT (16% default for Mexico)
}
```

### Chart of Accounts

| Code | Description | Type |
|------|-------------|------|
| 1101 | Banks | Asset |
| 1201 | Clients | Asset |
| 1401 | Creditable VAT | Asset |
| 2101 | Suppliers | Liability |
| 2401 | VAT Payable | Liability |
| 3101 | Share Capital | Equity |
| 4101 | Sales | Income |
| 5101 | Purchases | Expense |
| 5201 | Wages and Salaries | Expense |
| 5301 | General Expenses | Expense |

## 🔧 Advanced Technical Features

### 1. Fresh DSL Instances for Stability

```go
// ✅ RECOMMENDED PATTERN: New instance for each operation
func procesarOperacion(comando string) (interface{}, error) {
    contabilidad := createContabilidadDSL(sistema)  // New instance
    return contabilidad.Parse(comando)
}
```

**Why is this important?**
- Eliminates race conditions
- Guarantees stability in concurrent systems
- Avoids intermittent errors
- Recommended for production systems

### 2. Left-Recursive Grammars

```go
// These rules previously caused stack overflow
contabilidad.Rule("movements", []string{"movement"}, "singleMovement")
contabilidad.Rule("movements", []string{"movements", "movement"}, "multipleMovements")
```

**How it works:**
- ImprovedParser with memoization (packrat parsing)
- Automatic left-recursion detection
- Partial result caching to avoid infinite loops

### 3. KeywordToken vs Token

```go
// ✅ CORRECT: Keywords with high priority
contabilidad.KeywordToken("DEBE", "debe")  // Priority 90

// Generic token with low priority
contabilidad.Token("IMPORTE", "[0-9]+")    // Priority 0
```

**Advantages:**
- Resolves conflicts automatically
- Doesn't depend on definition order
- Works 100% of the time without exceptions

### 4. Automatic Accounting Validation

```go
contabilidad.Action("processEntry", func(args []interface{}) (interface{}, error) {
    movements := args[1].([]Movimiento)
    
    // Validate that Debit = Credit
    totalDebe := 0.0
    totalHaber := 0.0
    for _, m := range movements {
        totalDebe += m.Debe
        totalHaber += m.Haber
    }
    
    if totalDebe != totalHaber {
        return nil, fmt.Errorf("unbalanced entry: %.2f != %.2f", totalDebe, totalHaber)
    }
    
    return createAsiento(sistema, "Manual entry", movements), nil
})
```

## 📊 Example Output

```
=== Enhanced Accounting DSL System ===

1. Simple sale of $1,000:
   Command: venta de 1000

   Entry #1 - 2025-07-23
   Sale for 1000.00
   -------------------------------------------------------
   1201 Clients                       $1000.00
   4101 Sales                                      $1000.00
   -------------------------------------------------------
   Totals:                         $1000.00     $1000.00

10. Complex manual entry:
   Command: asiento debe 1101 10000 debe 1401 1600 haber 2101 11600

   Entry #10 - 2025-07-23
   Manual entry
   -------------------------------------------------------
   1101 Banks                        $10000.00
   1401 Creditable VAT                $1600.00
   2101 Suppliers                                 $11600.00
   -------------------------------------------------------
   Totals:                        $11600.00    $11600.00

Trial Balance:
==============
✅ The balance is balanced
```

## 🎓 Lessons Learned

### 1. **KeywordToken is Essential**
Without KeywordToken, words like "debe" (debit) and "haber" (credit) can be captured by generic tokens, causing intermittent errors.

### 2. **Fresh Instances for Critical Systems**
For maximum stability, create new DSL instances for each operation.

### 3. **Left Recursion Works**
go-dsl now perfectly handles complex grammars that were previously impossible.

### 4. **Business Rule Validation**
Actions can implement complex validations (like accounting balances).

## 🔗 Similar Use Cases

This pattern can be adapted for:

- **Billing Systems**: With different document types
- **Inventory**: With in/out movements
- **Payroll**: With complex deduction calculations
- **Budgets**: With allocations and transfers
- **Auditing**: With complete transaction traceability

## 🚀 Next Steps

1. **Try the example**: `go run main.go`
2. **Modify commands** in the code
3. **Add new accounting operations**
4. **Implement your own chart of accounts**
5. **Integrate with a real database**

## 📞 Support

- **Documentation**: [Quick Guide](../../docs/es/guia_rapida.md) (Spanish)
- **Complete Manual**: [Usage Manual](../../docs/es/manual_de_uso.md) (Spanish)
- **For Contributors**: [Developer Onboarding](../../docs/es/developer_onboarding.md) (Spanish)
- **English README**: [Main README](../../README.md)

---

**This example demonstrates that go-dsl is ready for production enterprise systems!** 🎉