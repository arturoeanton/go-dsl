# Multi-Country Accounting System - Advanced Example

**Enterprise accounting system with multi-country support and automatic tax calculation using dynamic context.**

## 🎯 Objective

This example demonstrates how to create a **multi-country accounting system** that uses go-dsl's dynamic context (equivalent to r2lang's `q.use()`) for:

- 🌍 Support for multiple countries (Mexico, Colombia, Argentina, Peru)
- 💰 Automatic VAT calculation per country
- 🔄 Dynamic context for configuration changes
- 📊 Transactions with custom descriptions
- 🏢 Enterprise-ready international system

## 🚀 Quick Start

```bash
cd examples/accounting
go run main.go
```

## 🌐 Supported Countries and VAT Rates

| Country | Code | VAT Rate | Currency |
|---------|------|----------|----------|
| Mexico | MX | 16% | MXN |
| Colombia | COL | 19% | COP |
| Argentina | AR | 21% | ARS |
| Peru | PE | 18% | PEN |

## 📚 DSL Features

### Defined Tokens

```go
// Actions (KeywordToken with high priority)
accounting.KeywordToken("REGISTRAR", "registrar")  // register
accounting.KeywordToken("CREAR", "crear")          // create
accounting.KeywordToken("ASIENTO", "asiento")      // entry

// Transaction types
accounting.KeywordToken("VENTA", "venta")          // sale
accounting.KeywordToken("COMPRA", "compra")        // purchase

// Connectors
accounting.KeywordToken("DE", "de")                // of/from
accounting.KeywordToken("CON", "con")              // with
accounting.KeywordToken("DESCRIPCION", "descripcion") // description

// Values
accounting.Token("AMOUNT", "[0-9]+\\.?[0-9]*")     // amount
accounting.Token("STRING", "\"[^\"]*\"")           // string
```

### Supported Commands

#### 1. Simple Transactions
```
registrar venta de 1000        # register sale of 1000
crear compra de 3000          # create purchase of 3000
asiento venta de 5000         # entry sale of 5000
```

#### 2. Transactions with Description
```
registrar venta de 5000 con descripcion "Venta de laptops"
crear compra de 3000 con descripcion "Materiales de oficina"
```

#### 3. Multiple Actions
Each action (`registrar`, `crear`, `asiento`) works identically - providing flexibility for different teams or regions.

## 🔄 Multi-Country Dynamic Context

### Equivalence with r2lang

```javascript
// r2lang
q.use("comando", {country: "MX", currency: "MXN"})

// go-dsl
dsl.Use("comando", map[string]interface{}{
    "country": "MX", 
    "currency": "MXN"
})
```

### Implementation in Go

```go
// Multi-country VAT calculation function
dsl.Set("calcularIVA", func(amount float64, country string) float64 {
    taxRates := map[string]float64{
        "MX":  0.16,  // Mexico: 16%
        "COL": 0.19,  // Colombia: 19% 
        "AR":  0.21,  // Argentina: 21%
        "PE":  0.18,  // Peru: 18%
    }
    
    rate, exists := taxRates[country]
    if !exists {
        rate = 0.16 // Default to Mexico
    }
    
    return amount * rate
})

// Usage in action
dsl.Action("fullTransaction", func(args []interface{}) (interface{}, error) {
    amount, _ := strconv.ParseFloat(args[3].(string), 64)
    
    // Get country from context
    country, _ := dsl.GetContext("country").(string)
    if country == "" {
        country = "MX" // Default
    }
    
    // Calculate VAT according to country
    calcIVA, _ := dsl.Get("calcularIVA")
    taxFn := calcIVA.(func(float64, string) float64)
    tax := taxFn(amount, country)
    
    return Transaction{
        Amount: amount,
        Tax:    tax,
        Total:  amount + tax,
    }, nil
})
```

## 🏗️ System Architecture

### Main Data Type

```go
type Transaction struct {
    Type        string  // "venta" (sale) or "compra" (purchase)
    Amount      float64 // Base amount
    Description string  // Optional description
    Tax         float64 // VAT calculated by country
    Total       float64 // Total amount (Amount + Tax)
}
```

### Processing Flow

1. **Command parsing** → Extracts type, amount, and description
2. **Country context** → Gets regional configuration
3. **VAT calculation** → Applies country-specific rate
4. **Result** → Complete Transaction with taxes

## 📊 Example Output

```
Accounting DSL Demo
===================

Country: MX
Command: registrar venta de 1000
Result:
  Type: venta
  Amount: $1000.00
  VAT: $160.00      # 16% Mexico
  Total: $1160.00
  Description: Sale transaction

Country: COL
Command: asiento compra de 3000
Result:
  Type: compra
  Amount: $3000.00
  VAT: $570.00      # 19% Colombia
  Total: $3570.00
  Description: Purchase transaction

Country: AR
Command: registrar venta de 10000 con descripcion "Servicios de consultoría"
Result:
  Type: venta
  Amount: $10000.00
  VAT: $2100.00     # 21% Argentina
  Total: $12100.00
  Description: Consulting services

Transaction Summary:
===================
Total: $21500.00 + VAT $4080.00 = $25580.00
```

## 🔧 Advanced Technical Features

### 1. Specific Rules per Combination

```go
// MORE specific rules first (more tokens)
accounting.Rule("transaction", []string{"REGISTRAR", "VENTA", "DE", "AMOUNT", "CON", "DESCRIPCION", "STRING"}, "fullTransaction")
accounting.Rule("transaction", []string{"CREAR", "COMPRA", "DE", "AMOUNT", "CON", "DESCRIPCION", "STRING"}, "fullTransaction")

// Simpler rules after
accounting.Rule("transaction", []string{"REGISTRAR", "VENTA", "DE", "AMOUNT"}, "simpleTransaction")
accounting.Rule("transaction", []string{"CREAR", "COMPRA", "DE", "AMOUNT"}, "simpleTransaction")
```

### 2. Registered Go Functions

```go
// Register Go function for use in DSL
dsl.Set("formatMoney", func(amount float64) string {
    return fmt.Sprintf("$%.2f", amount)
})

// Use in actions
formatFn, _ := dsl.Get("formatMoney")
format := formatFn.(func(float64) string)
formattedAmount := format(transaction.Amount)
```

### 3. Multi-Dimensional Context

```go
// Complete context for each country
contextMX := map[string]interface{}{
    "country":  "MX",
    "currency": "MXN", 
    "taxRate":  0.16,
    "locale":   "es-MX",
}

contextCOL := map[string]interface{}{
    "country":  "COL",
    "currency": "COP",
    "taxRate":  0.19,
    "locale":   "es-CO",
}
```

## 🌍 Enterprise Use Cases

### 1. **Multi-National ERP**
- Same DSL syntax across all countries
- Automatic regional configuration
- Local tax compliance

### 2. **International Billing**
- Automatic tax calculations
- Multiple currencies
- Consolidated reports

### 3. **Distributed Accounting**
- Offices in different countries
- Same business rules
- Local tax adaptation

### 4. **Multi-Country Auditing**
- Complete traceability
- Regulatory compliance
- Unified reporting

## 🔄 Comparison with r2lang

| Feature | r2lang | go-dsl |
|---------|--------|--------|
| Dynamic context | `q.use("cmd", ctx)` | `dsl.Use("cmd", ctx)` |
| Context access | `context.country` | `dsl.GetContext("country")` |
| Functions | Automatic | `dsl.Set()` / `dsl.Get()` |
| Typing | Dynamic | Type assertions required |
| Performance | JavaScript | Native Go |

## 🚀 Possible Extensions

### 1. **More Countries**
```go
taxRates := map[string]float64{
    "MX":  0.16, "COL": 0.19, "AR": 0.21, "PE": 0.18,
    "CL":  0.19, "EC":  0.12, "UY":  0.22, "PY":  0.10,
}
```

### 2. **Multiple Taxes**
```go
// VAT + Withholdings + Local taxes
type TaxCalculation struct {
    VAT         float64
    Withholding float64
    LocalTax    float64
    Total       float64
}
```

### 3. **Country-Specific Validations**
```go
// Different validations per local regulations
func validateTransaction(country string, transaction Transaction) error {
    switch country {
    case "MX":
        return validateMexicanTransaction(transaction)
    case "COL":
        return validateColombianTransaction(transaction)
    // ...
    }
}
```

## 🎓 Key Lessons

### 1. **Dynamic Context is Powerful**
Allows same syntax with different behaviors based on configuration.

### 2. **KeywordToken for Multiple Actions**
`REGISTRAR`, `CREAR`, `ASIENTO` coexist without conflicts.

### 3. **Registered Go Functions**
Enable complex reusable logic from actions.

### 4. **Specific Rules First**
Longer rules have automatic precedence.

## 📞 Support and References

- **Source code**: [`main.go`](main.go)
- **National system**: [../contabilidad/](../contabilidad/)
- **Documentation**: [Usage Manual](../../docs/es/manual_de_uso.md) (Spanish)
- **API Reference**: [Developer Guide](../../docs/es/developer_onboarding.md) (Spanish)
- **Main README**: [Project Overview](../../README.md)

---

**Demonstrates that go-dsl can handle complex multi-national enterprise systems!** 🌍🎉