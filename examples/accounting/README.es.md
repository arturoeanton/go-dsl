# Sistema Contable Multi-País - Ejemplo Avanzado

**Sistema contable empresarial con soporte para múltiples países y cálculo automático de impuestos usando contexto dinámico.**

## 🎯 Objetivo

Este ejemplo demuestra cómo crear un **sistema contable multi-país** que utiliza el contexto dinámico de go-dsl (equivalente a `q.use()` de r2lang) para:

- 🌍 Soporte para múltiples países (México, Colombia, Argentina, Perú)
- 💰 Cálculo automático de IVA según el país
- 🔄 Contexto dinámico para cambiar configuraciones
- 📊 Transacciones con descripciones personalizadas
- 🏢 Sistema listo para uso empresarial internacional

## 🚀 Ejecución Rápida

```bash
cd examples/accounting
go run main.go
```

## 🌐 Países y Tasas de IVA Soportadas

| País | Código | Tasa IVA | Moneda |
|------|--------|----------|--------|
| México | MX | 16% | MXN |
| Colombia | COL | 19% | COP |
| Argentina | AR | 21% | ARS |
| Perú | PE | 18% | PEN |

## 📚 Características del DSL

### Tokens Definidos

```go
// Acciones (KeywordToken con prioridad alta)
accounting.KeywordToken("REGISTRAR", "registrar")  // register
accounting.KeywordToken("CREAR", "crear")          // create
accounting.KeywordToken("ASIENTO", "asiento")      // entry

// Tipos de transacción
accounting.KeywordToken("VENTA", "venta")          // sale
accounting.KeywordToken("COMPRA", "compra")        // purchase

// Conectores
accounting.KeywordToken("DE", "de")                // of/from
accounting.KeywordToken("CON", "con")              // with
accounting.KeywordToken("DESCRIPCION", "descripcion") // description

// Valores
accounting.Token("AMOUNT", "[0-9]+\\.?[0-9]*")     // amount
accounting.Token("STRING", "\"[^\"]*\"")           // string
```

### Comandos Soportados

#### 1. Transacciones Simples
```
registrar venta de 1000
crear compra de 3000
asiento venta de 5000
```

#### 2. Transacciones con Descripción
```
registrar venta de 5000 con descripcion "Venta de laptops"
crear compra de 3000 con descripcion "Materiales de oficina"
```

#### 3. Múltiples Acciones
Cada acción (`registrar`, `crear`, `asiento`) funciona idénticamente - flexibilidad para diferentes equipos o regiones.

## 🔄 Contexto Dinámico Multi-País

### Equivalencia con r2lang

```javascript
// r2lang
q.use("comando", {country: "MX", currency: "MXN"})

// go-dsl
dsl.Use("comando", map[string]interface{}{
    "country": "MX", 
    "currency": "MXN"
})
```

### Implementación en Go

```go
// Función de cálculo de IVA multi-país
dsl.Set("calcularIVA", func(amount float64, country string) float64 {
    taxRates := map[string]float64{
        "MX":  0.16,  // México: 16%
        "COL": 0.19,  // Colombia: 19% 
        "AR":  0.21,  // Argentina: 21%
        "PE":  0.18,  // Perú: 18%
    }
    
    rate, exists := taxRates[country]
    if !exists {
        rate = 0.16 // Default a México
    }
    
    return amount * rate
})

// Uso en acción
dsl.Action("fullTransaction", func(args []interface{}) (interface{}, error) {
    amount, _ := strconv.ParseFloat(args[3].(string), 64)
    
    // Obtener país del contexto
    country, _ := dsl.GetContext("country").(string)
    if country == "" {
        country = "MX" // Default
    }
    
    // Calcular IVA según el país
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

## 🏗️ Arquitectura del Sistema

### Tipo de Datos Principal

```go
type Transaction struct {
    Type        string  // "venta" o "compra"
    Amount      float64 // Monto base
    Description string  // Descripción opcional
    Tax         float64 // IVA calculado según país
    Total       float64 // Monto total (Amount + Tax)
}
```

### Flujo de Procesamiento

1. **Parse del comando** → Extrae tipo, monto y descripción
2. **Contexto del país** → Obtiene configuración regional
3. **Cálculo de IVA** → Aplica tasa específica del país
4. **Resultado** → Transaction completa con impuestos

## 📊 Ejemplo de Salida

```
Accounting DSL Demo
===================

País: MX
Comando: registrar venta de 1000
Resultado:
  Tipo: venta
  Monto: $1000.00
  IVA: $160.00      # 16% México
  Total: $1160.00
  Descripción: Transacción de venta

País: COL
Comando: asiento compra de 3000
Resultado:
  Tipo: compra
  Monto: $3000.00
  IVA: $570.00      # 19% Colombia
  Total: $3570.00
  Descripción: Transacción de compra

País: AR
Comando: registrar venta de 10000 con descripcion "Servicios de consultoría"
Resultado:
  Tipo: venta
  Monto: $10000.00
  IVA: $2100.00     # 21% Argentina
  Total: $12100.00
  Descripción: Servicios de consultoría

Resumen de Transacciones:
========================
Total: $21500.00 + IVA $4080.00 = $25580.00
```

## 🔧 Características Técnicas Avanzadas

### 1. Reglas Específicas por Combinación

```go
// Reglas MÁS específicas primero (más tokens)
accounting.Rule("transaction", []string{"REGISTRAR", "VENTA", "DE", "AMOUNT", "CON", "DESCRIPCION", "STRING"}, "fullTransaction")
accounting.Rule("transaction", []string{"CREAR", "COMPRA", "DE", "AMOUNT", "CON", "DESCRIPCION", "STRING"}, "fullTransaction")

// Reglas más simples después
accounting.Rule("transaction", []string{"REGISTRAR", "VENTA", "DE", "AMOUNT"}, "simpleTransaction")
accounting.Rule("transaction", []string{"CREAR", "COMPRA", "DE", "AMOUNT"}, "simpleTransaction")
```

### 2. Funciones Go Registradas

```go
// Registrar función Go para usar en DSL
dsl.Set("formatMoney", func(amount float64) string {
    return fmt.Sprintf("$%.2f", amount)
})

// Usar en acciones
formatFn, _ := dsl.Get("formatMoney")
format := formatFn.(func(float64) string)
formattedAmount := format(transaction.Amount)
```

### 3. Contexto Multi-Dimensional

```go
// Contexto completo para cada país
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

## 🌍 Casos de Uso Empresariales

### 1. **ERP Multi-Nacional**
- Misma sintaxis DSL en todos los países
- Configuración automática por región
- Cumplimiento fiscal local

### 2. **Facturación Internacional**
- Cálculos de impuestos automáticos
- Múltiples monedas
- Reportes consolidados

### 3. **Contabilidad Distribuida**
- Oficinas en diferentes países
- Mismas reglas de negocio
- Adaptación fiscal local

### 4. **Auditoría Multi-País**
- Trazabilidad completa
- Cumplimiento normativo
- Reportes unificados

## 🔄 Comparación con r2lang

| Característica | r2lang | go-dsl |
|----------------|--------|--------|
| Contexto dinámico | `q.use("cmd", ctx)` | `dsl.Use("cmd", ctx)` |
| Acceso a contexto | `context.country` | `dsl.GetContext("country")` |
| Funciones | Automático | `dsl.Set()` / `dsl.Get()` |
| Tipado | Dinámico | Type assertions necesarias |
| Rendimiento | JavaScript | Go nativo |

## 🚀 Extensiones Posibles

### 1. **Más Países**
```go
taxRates := map[string]float64{
    "MX":  0.16, "COL": 0.19, "AR": 0.21, "PE": 0.18,
    "CL":  0.19, "EC":  0.12, "UY":  0.22, "PY":  0.10,
}
```

### 2. **Múltiples Impuestos**
```go
// IVA + Retenciones + Impuestos locales
type TaxCalculation struct {
    IVA         float64
    Retention   float64
    LocalTax    float64
    Total       float64
}
```

### 3. **Validaciones por País**
```go
// Diferentes validaciones según normativa local
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

## 🎓 Lecciones Clave

### 1. **Contexto Dinámico es Poderoso**
Permite misma sintaxis con comportamientos diferentes según configuración.

### 2. **KeywordToken para Múltiples Acciones**
`REGISTRAR`, `CREAR`, `ASIENTO` coexisten sin conflictos.

### 3. **Funciones Go Registradas**
Permiten lógica compleja reutilizable desde las acciones.

### 4. **Reglas Específicas Primero**
Reglas más largas tienen precedencia automática.

## 📞 Soporte y Referencias

- **Código fuente**: [`main.go`](main.go)
- **Sistema nacional**: [../contabilidad/](../contabilidad/)
- **Documentación**: [Manual de Uso](../../docs/es/manual_de_uso.md)
- **API Reference**: [Developer Guide](../../docs/es/developer_onboarding.md)

---

**¡Demuestra que go-dsl puede manejar sistemas empresariales multi-nacionales complejos!** 🌍🎉