# Manual de Uso - go-dsl

**Guía completa para usuarios: Desde conceptos básicos hasta sistemas empresariales complejos.**

Este manual te guiará a través de todas las características de go-dsl, con ejemplos prácticos y casos de uso reales. Perfecto para desarrolladores que quieren dominar la creación de DSLs empresariales.

## 📋 Tabla de Contenidos

1. [Conceptos Fundamentales](#-conceptos-fundamentales)
2. [API Completa](#-api-completa)
3. [Casos de Uso por Dominio](#-casos-de-uso-por-dominio)
4. [Características Avanzadas](#-características-avanzadas)
5. [Mejores Prácticas](#-mejores-prácticas)
6. [Solución de Problemas](#-solución-de-problemas)
7. [Casos de Estudio](#-casos-de-estudio)

## 🎯 Conceptos Fundamentales

### ¿Qué es un DSL?

Un **Domain Specific Language** (Lenguaje Específico de Dominio) es un lenguaje de programación especializado para un dominio particular. go-dsl te permite crear estos lenguajes fácilmente.

**Ejemplos de DSLs exitosos:**
- **SQL**: Para consultas de bases de datos
- **CSS**: Para estilos web
- **Makefile**: Para automatización de builds
- **go-dsl**: Para cualquier dominio que imagines

### Componentes de un DSL en go-dsl

```
📝 ENTRADA: "venta de 5000 con iva"
     ↓
🔍 TOKENIZER: [VENTA:venta] [DE:de] [IMPORTE:5000] [CON:con] [IVA:iva]
     ↓
🎯 PARSER: Aplica reglas gramaticales
     ↓
⚙️ ACTIONS: Ejecuta lógica de negocio
     ↓
📊 RESULTADO: Transaction{Amount: 5000, Tax: 800, Total: 5800}
```

## 🔧 API Completa

### Crear DSL

```go
import "github.com/arturoeliasanton/go-dsl/pkg/dslbuilder"

// Crear nueva instancia
dsl := dslbuilder.New("NombreDSL")
```

### Definir Tokens

#### KeywordToken (Recomendado)
```go
// Para palabras clave específicas - SIEMPRE úsalo para keywords
dsl.KeywordToken("VENTA", "venta")
dsl.KeywordToken("COMPRA", "compra")
dsl.KeywordToken("CON", "con")
```

#### Token Regular
```go
// Para patrones generales
dsl.Token("IMPORTE", "[0-9]+\\.?[0-9]*")
dsl.Token("STRING", "\"[^\"]*\"")
dsl.Token("ID", "[a-zA-Z_][a-zA-Z0-9_]*")
```

**🎯 Regla de Oro**: KeywordToken para palabras específicas, Token para patrones.

### Definir Reglas Gramaticales

```go
// Sintaxis: Rule(nombre, secuencia_de_tokens, acción)
dsl.Rule("command", []string{"VENTA", "DE", "IMPORTE"}, "simpleSale")
dsl.Rule("command", []string{"VENTA", "DE", "IMPORTE", "CON", "IVA"}, "saleWithTax")

// ✨ Gramáticas recursivas por la izquierda (¡ahora funciona!)
dsl.Rule("movements", []string{"movement"}, "singleMovement")
dsl.Rule("movements", []string{"movements", "movement"}, "multipleMovements")
```

**💡 Tip**: Reglas más específicas (más tokens) primero.

### Definir Acciones

```go
dsl.Action("simpleSale", func(args []interface{}) (interface{}, error) {
    // args[0] = "VENTA", args[1] = "DE", args[2] = importe
    amount, _ := strconv.ParseFloat(args[2].(string), 64)
    return Transaction{Amount: amount}, nil
})

dsl.Action("saleWithTax", func(args []interface{}) (interface{}, error) {
    amount, _ := strconv.ParseFloat(args[2].(string), 64)
    tax := amount * 0.16  // 16% IVA
    return Transaction{Amount: amount, Tax: tax, Total: amount + tax}, nil
})
```

### Ejecutar DSL

#### Parsing Básico
```go
result, err := dsl.Parse("venta de 5000")
if err != nil {
    log.Fatal(err)
}
transaction := result.GetOutput().(Transaction)
```

#### Con Contexto Dinámico (Como r2lang)
```go
context := map[string]interface{}{
    "country": "MX",
    "currency": "MXN",
    "taxRate": 0.16,
}

result, err := dsl.Use("venta de 5000 con iva", context)
// El contexto está disponible en las acciones via dsl.GetContext()
```

### Manejo de Contexto

#### Establecer Contexto Persistente
```go
dsl.SetContext("company", "Mi Empresa SA")
dsl.SetContext("defaultTax", 0.16)
```

#### Obtener Contexto en Acciones
```go
dsl.Action("processTransaction", func(args []interface{}) (interface{}, error) {
    company := dsl.GetContext("company").(string)
    taxRate := dsl.GetContext("defaultTax").(float64)
    
    return fmt.Sprintf("Procesado por %s con IVA %.0f%%", company, taxRate*100), nil
})
```

#### Registrar Funciones Go
```go
// Registrar función Go para usar en acciones
dsl.Set("calculateTax", func(amount float64, country string) float64 {
    rates := map[string]float64{"MX": 0.16, "COL": 0.19, "AR": 0.21}
    return amount * rates[country]
})

// Usar en acciones
dsl.Action("taxCalculation", func(args []interface{}) (interface{}, error) {
    calcTax, _ := dsl.Get("calculateTax")
    taxFunc := calcTax.(func(float64, string) float64)
    
    amount := parseFloat(args[1])
    country := dsl.GetContext("country").(string)
    
    return taxFunc(amount, country), nil
})
```

## 🏢 Casos de Uso por Dominio

### 1. Sistema Contable Empresarial

```go
func createAccountingDSL() *dslbuilder.DSL {
    accounting := dslbuilder.New("Accounting")
    
    // Tokens con prioridad KeywordToken
    accounting.KeywordToken("ASIENTO", "asiento")
    accounting.KeywordToken("DEBE", "debe")
    accounting.KeywordToken("HABER", "haber")
    accounting.KeywordToken("VENTA", "venta")
    accounting.KeywordToken("DE", "de")
    accounting.KeywordToken("CON", "con")
    accounting.KeywordToken("IVA", "iva")
    
    accounting.Token("IMPORTE", "[0-9]+\\.?[0-9]*")
    accounting.Token("STRING", "\"[^\"]*\"")
    
    // Reglas - más específicas primero
    accounting.Rule("command", []string{"VENTA", "DE", "IMPORTE", "CON", "IVA"}, "saleWithTax")
    accounting.Rule("command", []string{"VENTA", "DE", "IMPORTE"}, "simpleSale")
    
    // Gramáticas recursivas para asientos complejos
    accounting.Rule("command", []string{"ASIENTO", "movements"}, "balancedEntry")
    accounting.Rule("movements", []string{"movement"}, "singleMovement")
    accounting.Rule("movements", []string{"movements", "movement"}, "multipleMovements")
    accounting.Rule("movement", []string{"DEBE", "IMPORTE", "IMPORTE"}, "debitEntry")
    accounting.Rule("movement", []string{"HABER", "IMPORTE", "IMPORTE"}, "creditEntry")
    
    // Acciones con validación de negocio
    accounting.Action("balancedEntry", func(args []interface{}) (interface{}, error) {
        movements := args[1].([]Movement)
        
        totalDebit := 0.0
        totalCredit := 0.0
        for _, m := range movements {
            totalDebit += m.Debit
            totalCredit += m.Credit
        }
        
        if totalDebit != totalCredit {
            return nil, fmt.Errorf("asiento descuadrado: %.2f != %.2f", totalDebit, totalCredit)
        }
        
        return AccountingEntry{Movements: movements, Balanced: true}, nil
    })
    
    return accounting
}

// Uso del DSL contable
func main() {
    accounting := createAccountingDSL()
    
    // Casos de uso
    examples := []string{
        "venta de 5000",
        "venta de 5000 con iva",
        "asiento debe 1101 10000 debe 1401 1600 haber 2101 11600",
    }
    
    for _, example := range examples {
        result, err := accounting.Parse(example)
        if err != nil {
            fmt.Printf("Error: %v\n", err)
            continue
        }
        fmt.Printf("✅ %s → %+v\n", example, result.GetOutput())
    }
}
```

### 2. Sistema de Consultas LINQ en Español

```go
func createLinqDSL() *dslbuilder.DSL {
    linq := dslbuilder.New("LINQ-ES")
    
    // Tokens para consultas
    linq.KeywordToken("SELECCIONAR", "seleccionar")
    linq.KeywordToken("DE", "de")
    linq.KeywordToken("DONDE", "donde")
    linq.KeywordToken("ORDENAR", "ordenar")
    linq.KeywordToken("POR", "por")
    linq.KeywordToken("AGRUPAR", "agrupar")
    
    // Campos y operadores
    linq.KeywordToken("NOMBRE", "nombre")
    linq.KeywordToken("EDAD", "edad")
    linq.KeywordToken("CIUDAD", "ciudad")
    linq.KeywordToken("MAYOR", "mayor")
    linq.KeywordToken("MENOR", "menor")
    linq.KeywordToken("IGUAL", "igual")
    
    linq.Token("NUMBER", "[0-9]+")
    linq.Token("STRING", "\"[^\"]*\"")
    linq.Token("ID", "[a-zA-Z_][a-zA-Z0-9_]*")
    
    // Reglas de consulta
    linq.Rule("query", []string{"SELECCIONAR", "field", "DE", "ID"}, "simpleSelect")
    linq.Rule("query", []string{"SELECCIONAR", "field", "DE", "ID", "DONDE", "condition"}, "selectWhere")
    linq.Rule("query", []string{"SELECCIONAR", "field", "DE", "ID", "DONDE", "condition", "ORDENAR", "POR", "field"}, "selectWhereOrder")
    
    linq.Rule("field", []string{"NOMBRE"}, "nameField")
    linq.Rule("field", []string{"EDAD"}, "ageField")
    linq.Rule("field", []string{"CIUDAD"}, "cityField")
    
    linq.Rule("condition", []string{"EDAD", "MAYOR", "NUMBER"}, "ageGreater")
    linq.Rule("condition", []string{"CIUDAD", "IGUAL", "STRING"}, "cityEquals")
    
    // Acciones con acceso a datos
    linq.Action("selectWhere", func(args []interface{}) (interface{}, error) {
        field := args[1].(string)
        dataset := args[3].(string)
        condition := args[5].(FilterCondition)
        
        // Obtener datos del contexto
        data := linq.GetContext(dataset).([]Person)
        
        // Aplicar filtro
        filtered := applyFilter(data, condition)
        
        // Seleccionar campo
        return selectField(filtered, field), nil
    })
    
    return linq
}

// Uso con datos dinámicos
func main() {
    linq := createLinqDSL()
    
    // Datos de ejemplo
    people := []Person{
        {"Juan García", 28, "Madrid"},
        {"María López", 35, "Barcelona"},
        {"Carlos Rodríguez", 42, "Madrid"},
    }
    
    // Contexto con datos
    context := map[string]interface{}{
        "personas": people,
        "empleados": people,
    }
    
    // Consultas
    queries := []string{
        "seleccionar nombre de personas",
        "seleccionar edad de personas donde edad mayor 30",
        "seleccionar ciudad de empleados donde ciudad igual \"Madrid\"",
    }
    
    for _, query := range queries {
        result, err := linq.Use(query, context)
        if err != nil {
            fmt.Printf("Error: %v\n", err)
            continue
        }
        fmt.Printf("✅ %s\n   → %v\n\n", query, result.GetOutput())
    }
}
```

### 3. Calculadora Científica

```go
func createScientificCalculator() *dslbuilder.DSL {
    calc := dslbuilder.New("ScientificCalc")
    
    // Tokens matemáticos
    calc.Token("NUMBER", "[0-9]+\\.?[0-9]*")
    calc.KeywordToken("PI", "pi")
    calc.KeywordToken("E", "e")
    
    // Operadores básicos
    calc.Token("PLUS", "\\+")
    calc.Token("MINUS", "-")
    calc.Token("MULTIPLY", "\\*")
    calc.Token("DIVIDE", "/")
    calc.Token("POWER", "\\^")
    
    // Funciones científicas
    calc.KeywordToken("SIN", "sin")
    calc.KeywordToken("COS", "cos")
    calc.KeywordToken("TAN", "tan")
    calc.KeywordToken("LOG", "log")
    calc.KeywordToken("SQRT", "sqrt")
    
    // Paréntesis
    calc.Token("LPAREN", "\\(")
    calc.Token("RPAREN", "\\)")
    
    // Gramática con precedencia implícita
    calc.Rule("expression", []string{"term"}, "passthrough")
    calc.Rule("expression", []string{"expression", "PLUS", "term"}, "add")
    calc.Rule("expression", []string{"expression", "MINUS", "term"}, "subtract")
    
    calc.Rule("term", []string{"factor"}, "passthrough")
    calc.Rule("term", []string{"term", "MULTIPLY", "factor"}, "multiply")
    calc.Rule("term", []string{"term", "DIVIDE", "factor"}, "divide")
    
    calc.Rule("factor", []string{"power"}, "passthrough")
    calc.Rule("factor", []string{"factor", "POWER", "power"}, "power")
    
    calc.Rule("power", []string{"atom"}, "passthrough")
    calc.Rule("power", []string{"function"}, "passthrough")
    calc.Rule("power", []string{"LPAREN", "expression", "RPAREN"}, "parentheses")
    
    calc.Rule("atom", []string{"NUMBER"}, "number")
    calc.Rule("atom", []string{"PI"}, "pi")
    calc.Rule("atom", []string{"E"}, "e")
    
    calc.Rule("function", []string{"SIN", "LPAREN", "expression", "RPAREN"}, "sin")
    calc.Rule("function", []string{"COS", "LPAREN", "expression", "RPAREN"}, "cos")
    calc.Rule("function", []string{"SQRT", "LPAREN", "expression", "RPAREN"}, "sqrt")
    
    // Acciones matemáticas
    calc.Action("add", func(args []interface{}) (interface{}, error) {
        left := args[0].(float64)
        right := args[2].(float64)
        return left + right, nil
    })
    
    calc.Action("sin", func(args []interface{}) (interface{}, error) {
        value := args[2].(float64)
        return math.Sin(value), nil
    })
    
    calc.Action("sqrt", func(args []interface{}) (interface{}, error) {
        value := args[2].(float64)
        if value < 0 {
            return nil, fmt.Errorf("no se puede calcular raíz cuadrada de número negativo")
        }
        return math.Sqrt(value), nil
    })
    
    // ... más acciones matemáticas
    
    return calc
}

// Uso de calculadora científica
func main() {
    calc := createScientificCalculator()
    
    expressions := []string{
        "2 + 3 * 4",
        "sin(pi / 2)",
        "sqrt(16) + log(e)",
        "(2 + 3) * (4 - 1)",
        "2 ^ 3 ^ 2",  // 2^(3^2) = 2^9 = 512
    }
    
    for _, expr := range expressions {
        result, err := calc.Parse(expr)
        if err != nil {
            fmt.Printf("Error en '%s': %v\n", expr, err)
            continue
        }
        fmt.Printf("📊 %s = %.6f\n", expr, result.GetOutput().(float64))
    }
}
```

### 4. DSL para Reglas de Negocio

```go
func createBusinessRulesDSL() *dslbuilder.DSL {
    rules := dslbuilder.New("BusinessRules")
    
    // Tokens para reglas
    rules.KeywordToken("SI", "si")
    rules.KeywordToken("ENTONCES", "entonces")
    rules.KeywordToken("SINO", "sino")
    rules.KeywordToken("CLIENTE", "cliente")
    rules.KeywordToken("VIP", "vip")
    rules.KeywordToken("REGULAR", "regular")
    rules.KeywordToken("COMPRA", "compra")
    rules.KeywordToken("MAYOR", "mayor")
    rules.KeywordToken("DESCUENTO", "descuento")
    rules.KeywordToken("ENVIO", "envio")
    rules.KeywordToken("GRATIS", "gratis")
    
    rules.Token("PERCENT", "[0-9]+%")
    rules.Token("AMOUNT", "[0-9]+\\.?[0-9]*")
    
    // Reglas de negocio
    rules.Rule("rule", []string{"SI", "condition", "ENTONCES", "action"}, "simpleRule")
    rules.Rule("rule", []string{"SI", "condition", "ENTONCES", "action", "SINO", "action"}, "conditionalRule")
    
    rules.Rule("condition", []string{"CLIENTE", "VIP"}, "isVipClient")
    rules.Rule("condition", []string{"COMPRA", "MAYOR", "AMOUNT"}, "purchaseGreater")
    
    rules.Rule("action", []string{"DESCUENTO", "PERCENT"}, "applyDiscount")
    rules.Rule("action", []string{"ENVIO", "GRATIS"}, "freeShipping")
    
    // Acciones de reglas de negocio
    rules.Action("simpleRule", func(args []interface{}) (interface{}, error) {
        condition := args[1].(BusinessCondition)
        action := args[3].(BusinessAction)
        
        return BusinessRule{
            Condition: condition,
            Action:    action,
            Type:      "simple",
        }, nil
    })
    
    rules.Action("isVipClient", func(args []interface{}) (interface{}, error) {
        // Verificar en contexto si el cliente es VIP
        clientType := rules.GetContext("clientType").(string)
        return BusinessCondition{
            Type:   "clientType",
            Value:  "VIP",
            Result: clientType == "VIP",
        }, nil
    })
    
    rules.Action("applyDiscount", func(args []interface{}) (interface{}, error) {
        percentStr := strings.TrimSuffix(args[1].(string), "%")
        percent, _ := strconv.Atoi(percentStr)
        
        return BusinessAction{
            Type:  "discount",
            Value: float64(percent) / 100,
        }, nil
    })
    
    return rules
}

// Motor de reglas de negocio
type BusinessRuleEngine struct {
    dsl   *dslbuilder.DSL
    rules []BusinessRule
}

func (bre *BusinessRuleEngine) AddRule(ruleText string) error {
    result, err := bre.dsl.Parse(ruleText)
    if err != nil {
        return err
    }
    
    rule := result.GetOutput().(BusinessRule)
    bre.rules = append(bre.rules, rule)
    return nil
}

func (bre *BusinessRuleEngine) ExecuteRules(context map[string]interface{}) []BusinessAction {
    var actions []BusinessAction
    
    for _, rule := range bre.rules {
        // Establecer contexto para evaluación
        for k, v := range context {
            bre.dsl.SetContext(k, v)
        }
        
        // Evaluar condición (esto requeriría más lógica)
        if rule.Condition.Result {
            actions = append(actions, rule.Action)
        }
    }
    
    return actions
}

// Uso del motor de reglas
func main() {
    ruleEngine := &BusinessRuleEngine{
        dsl: createBusinessRulesDSL(),
    }
    
    // Definir reglas de negocio
    businessRules := []string{
        "si cliente vip entonces descuento 20%",
        "si compra mayor 1000 entonces envio gratis",
        "si cliente regular entonces descuento 5%",
    }
    
    for _, rule := range businessRules {
        err := ruleEngine.AddRule(rule)
        if err != nil {
            fmt.Printf("Error en regla '%s': %v\n", rule, err)
            continue
        }
        fmt.Printf("✅ Regla agregada: %s\n", rule)
    }
    
    // Contexto de cliente
    clientContext := map[string]interface{}{
        "clientType":    "VIP",
        "purchaseAmount": 1500.0,
        "clientHistory": "good",
    }
    
    // Ejecutar reglas
    actions := ruleEngine.ExecuteRules(clientContext)
    
    fmt.Println("\n🎯 Acciones aplicables:")
    for _, action := range actions {
        fmt.Printf("  - %s: %.0f%%\n", action.Type, action.Value*100)
    }
}
```

## 🚀 Características Avanzadas

### Gramáticas Recursivas por la Izquierda

**Problema Clásico**: Las gramáticas como `A → A B` causan stack overflow en parsers descendentes.

**Solución go-dsl**: Usa ImprovedParser automáticamente cuando detecta recursión.

```go
// ✅ Esto ahora funciona perfectamente
dsl.Rule("list", []string{"item"}, "singleItem")
dsl.Rule("list", []string{"list", "COMMA", "item"}, "multipleItems")

// Ejemplo: "item1, item2, item3, item4"
result, _ := dsl.Parse("item1, item2, item3, item4")
items := result.GetOutput().([]Item)  // 4 elementos
```

**Casos de Uso Reales:**
- Listas de elementos
- Expresiones matemáticas anidadas
- Asientos contables complejos
- Comandos con múltiples parámetros

### Contexto Dinámico Avanzado

#### Contexto con Múltiples Fuentes
```go
// Combinar contextos
baseContext := map[string]interface{}{
    "company": "Mi Empresa",
    "version": "1.0",
}

sessionContext := map[string]interface{}{
    "user": "juan.perez",
    "role": "admin",
}

// Fusionar contextos
mergedContext := mergeMaps(baseContext, sessionContext)
result, _ := dsl.Use("comando", mergedContext)
```

#### Contexto con Funciones
```go
// Registrar funciones complejas
dsl.Set("calculateTax", func(amount float64, country string, clientType string) TaxResult {
    base := amount * getTaxRate(country)
    
    if clientType == "VIP" {
        base *= 0.95  // 5% descuento VIP
    }
    
    return TaxResult{
        BaseAmount: amount,
        TaxAmount:  base,
        Total:      amount + base,
        Country:    country,
    }
})

// Usar en acciones
dsl.Action("complexTax", func(args []interface{}) (interface{}, error) {
    taxFunc := dsl.Get("calculateTax").(func(float64, string, string) TaxResult)
    
    amount := parseFloat(args[1])
    country := dsl.GetContext("country").(string)
    clientType := dsl.GetContext("clientType").(string)
    
    return taxFunc(amount, country, clientType), nil
})
```

### Debug y Herramientas de Desarrollo

#### Debug de Tokenización
```go
// Ver cómo se tokeniza una entrada
tokens, err := dsl.DebugTokens("venta de 5000 con iva")
if err != nil {
    log.Fatal(err)
}

for _, token := range tokens {
    fmt.Printf("[%s:%s] ", token.TokenType, token.Value)
}
// Output: [VENTA:venta] [DE:de] [IMPORTE:5000] [CON:con] [IVA:iva]
```

#### Debug de Gramática
```go
// Ver estructura interna del DSL
debugInfo := dsl.Debug()
fmt.Printf("Tokens: %+v\n", debugInfo["tokens"])
fmt.Printf("Rules: %+v\n", debugInfo["rules"])
fmt.Printf("Actions: %+v\n", debugInfo["actions"])
```

#### Logging Personalizado
```go
// Agregar logging a acciones
dsl.Action("loggedAction", func(args []interface{}) (interface{}, error) {
    start := time.Now()
    defer func() {
        fmt.Printf("Action completed in %v\n", time.Since(start))
    }()
    
    // Lógica de la acción
    result := processBusinessLogic(args)
    
    // Log del resultado
    fmt.Printf("Action result: %+v\n", result)
    return result, nil
})
```

## 💡 Mejores Prácticas

### 1. Diseño de Tokens

#### ✅ Hacer
```go
// Palabras clave específicas con KeywordToken
dsl.KeywordToken("REGISTRAR", "registrar")
dsl.KeywordToken("CREAR", "crear")
dsl.KeywordToken("ELIMINAR", "eliminar")

// Patrones genéricos con Token
dsl.Token("AMOUNT", "[0-9]+\\.?[0-9]*")
dsl.Token("STRING", "\"[^\"]*\"")
dsl.Token("ID", "[a-zA-Z_][a-zA-Z0-9_]*")
```

#### ❌ Evitar
```go
// No uses Token para keywords - causará conflictos
dsl.Token("WORDS", "[a-zA-Z]+")  // Capturaría "registrar", "crear", etc.
dsl.Token("REGISTRAR", "registrar")  // Nunca se ejecutaría
```

### 2. Diseño de Reglas

#### ✅ Hacer
```go
// Reglas más específicas primero
dsl.Rule("command", []string{"REGISTRAR", "VENTA", "DE", "AMOUNT", "CON", "CLIENTE", "STRING"}, "fullSale")
dsl.Rule("command", []string{"REGISTRAR", "VENTA", "DE", "AMOUNT"}, "simpleSale")
dsl.Rule("command", []string{"REGISTRAR", "VENTA"}, "basicSale")
```

#### ❌ Evitar
```go
// Reglas genéricas primero - capturarán casos específicos
dsl.Rule("command", []string{"REGISTRAR", "VENTA"}, "basicSale")  // Muy genérica
dsl.Rule("command", []string{"REGISTRAR", "VENTA", "DE", "AMOUNT"}, "simpleSale")  // Nunca se ejecutará
```

### 3. Manejo de Errores

#### ✅ Hacer
```go
dsl.Action("safeAction", func(args []interface{}) (interface{}, error) {
    // Validar argumentos
    if len(args) < 3 {
        return nil, fmt.Errorf("se requieren al menos 3 argumentos, recibidos %d", len(args))
    }
    
    // Type assertions seguras
    amountStr, ok := args[2].(string)
    if !ok {
        return nil, fmt.Errorf("argumento 2 debe ser string, recibido %T", args[2])
    }
    
    // Validación de datos
    amount, err := strconv.ParseFloat(amountStr, 64)
    if err != nil {
        return nil, fmt.Errorf("importe inválido '%s': %v", amountStr, err)
    }
    
    if amount <= 0 {
        return nil, fmt.Errorf("el importe debe ser positivo, recibido %.2f", amount)
    }
    
    return Transaction{Amount: amount}, nil
})
```

### 4. Performance

#### Para DSLs Simples
```go
// Una instancia global está bien
var globalDSL = createSimpleDSL()

func parseCommand(input string) (interface{}, error) {
    return globalDSL.Parse(input)
}
```

#### Para DSLs Complejos/Empresariales
```go
// Instancias frescas para máxima estabilidad
func parseAccountingCommand(input string, context map[string]interface{}) (interface{}, error) {
    dsl := createAccountingDSL()  // Nueva instancia
    return dsl.Use(input, context)
}
```

### 5. Testing

#### Test Básico
```go
func TestBasicParsing(t *testing.T) {
    dsl := createTestDSL()
    
    testCases := []struct {
        input    string
        expected interface{}
        hasError bool
    }{
        {"valid command", ExpectedResult{}, false},
        {"invalid syntax", nil, true},
    }
    
    for _, tc := range testCases {
        result, err := dsl.Parse(tc.input)
        
        if tc.hasError {
            assert.Error(t, err, "esperaba error para: %s", tc.input)
        } else {
            assert.NoError(t, err, "no esperaba error para: %s", tc.input)
            assert.Equal(t, tc.expected, result.GetOutput())
        }
    }
}
```

#### Test con Contexto
```go
func TestContextualParsing(t *testing.T) {
    dsl := createTestDSL()
    
    context := map[string]interface{}{
        "country": "MX",
        "taxRate": 0.16,
    }
    
    result, err := dsl.Use("venta de 1000 con iva", context)
    assert.NoError(t, err)
    
    transaction := result.GetOutput().(Transaction)
    assert.Equal(t, 1000.0, transaction.Amount)
    assert.Equal(t, 160.0, transaction.Tax)  // 16% de 1000
}
```

## 🔧 Solución de Problemas

### Error: "parsing error: no alternative matched"

**Causa**: Ninguna regla coincide con la entrada.

**Solución**:
```go
// 1. Debug de tokenización primero
tokens, err := dsl.DebugTokens("tu entrada aquí")
if err != nil {
    fmt.Printf("Error de tokenización: %v\n", err)
} else {
    fmt.Printf("Tokens: %+v\n", tokens)
}

// 2. Verificar que tienes reglas para esa secuencia
// 3. Verificar orden de reglas (más específicas primero)
```

### Error: "unexpected token at position X"

**Causa**: Token no esperado en esa posición.

**Soluciones**:
```go
// 1. Verificar conflictos de tokens
dsl.KeywordToken("PALABRA", "palabra")  // En lugar de Token

// 2. Verificar patrones de regex
dsl.Token("NUMBER", "[0-9]+")  // Asegurar que el patrón es correcto

// 3. Debug la posición específica
tokens, _ := dsl.DebugTokens("entrada")
fmt.Printf("Token en posición %d: %+v\n", X, tokens[X])
```

### Error: Stack Overflow en Parsing

**Causa**: Gramática recursiva por la izquierda sin ImprovedParser.

**Solución**: go-dsl maneja esto automáticamente, pero si ocurre:
```go
// Verificar que tu gramática es correcta
dsl.Rule("list", []string{"item"}, "single")           // ✅ Caso base
dsl.Rule("list", []string{"list", "COMMA", "item"}, "multiple")  // ✅ Recursión

// No hacer recursión directa infinita
// dsl.Rule("bad", []string{"bad"}, "infinite")  // ❌ Esto causaría problemas
```

### Errores de Contexto

**Problema**: `panic: interface conversion: <nil> is not string`

**Solución**:
```go
// Verificación segura de contexto
func safeGetContext(dsl *dslbuilder.DSL, key string, defaultValue interface{}) interface{} {
    if value := dsl.GetContext(key); value != nil {
        return value
    }
    return defaultValue
}

// Uso en acciones
dsl.Action("safeAction", func(args []interface{}) (interface{}, error) {
    country := safeGetContext(dsl, "country", "MX").(string)
    taxRate := safeGetContext(dsl, "taxRate", 0.16).(float64)
    
    // ... resto de la lógica
})
```

## 📊 Casos de Estudio

### Caso 1: Sistema de Facturación Multi-País

**Desafío**: Mismo DSL, diferentes reglas fiscales por país.

**Solución**:
```go
type TaxSystem struct {
    dsl     *dslbuilder.DSL
    country string
    rates   map[string]float64
}

func NewTaxSystem(country string) *TaxSystem {
    rates := map[string]float64{
        "MX":  0.16,
        "COL": 0.19,
        "AR":  0.21,
        "PE":  0.18,
    }
    
    ts := &TaxSystem{
        dsl:     createBillingDSL(),
        country: country,
        rates:   rates,
    }
    
    // Configurar contexto del país
    ts.dsl.SetContext("country", country)
    ts.dsl.SetContext("taxRate", rates[country])
    
    return ts
}

func (ts *TaxSystem) ProcessInvoice(command string) (*Invoice, error) {
    result, err := ts.dsl.Parse(command)
    if err != nil {
        return nil, err
    }
    
    invoice := result.GetOutput().(*Invoice)
    invoice.Country = ts.country
    invoice.Currency = getCurrencyForCountry(ts.country)
    
    return invoice, nil
}

// Uso
func main() {
    // Diferentes países, mismo DSL
    mexicanSystem := NewTaxSystem("MX")
    colombianSystem := NewTaxSystem("COL")
    
    command := "facturar venta de 1000 con iva"
    
    mexInvoice, _ := mexicanSystem.ProcessInvoice(command)
    colInvoice, _ := colombianSystem.ProcessInvoice(command)
    
    fmt.Printf("México: %.2f (IVA: %.0f%%)\n", mexInvoice.Total, mexInvoice.TaxAmount/mexInvoice.BaseAmount*100)
    fmt.Printf("Colombia: %.2f (IVA: %.0f%%)\n", colInvoice.Total, colInvoice.TaxAmount/colInvoice.BaseAmount*100)
}
```

### Caso 2: DSL para Configuración de CI/CD

**Desafío**: Crear un DSL para pipelines de CI/CD en español.

**Solución**:
```go
func createCIPipelineDSL() *dslbuilder.DSL {
    ci := dslbuilder.New("CI-Pipeline")
    
    // Tokens para CI/CD
    ci.KeywordToken("PIPELINE", "pipeline")
    ci.KeywordToken("ETAPA", "etapa")
    ci.KeywordToken("TRABAJO", "trabajo")
    ci.KeywordToken("EJECUTAR", "ejecutar")
    ci.KeywordToken("CONSTRUIR", "construir")
    ci.KeywordToken("PROBAR", "probar")
    ci.KeywordToken("DESPLEGAR", "desplegar")
    ci.KeywordToken("EN", "en")
    ci.KeywordToken("CON", "con")
    
    ci.Token("STRING", "\"[^\"]*\"")
    ci.Token("ID", "[a-zA-Z_][a-zA-Z0-9_]*")
    
    // Estructura jerárquica
    ci.Rule("pipeline", []string{"PIPELINE", "ID", "stages"}, "createPipeline")
    ci.Rule("stages", []string{"stage"}, "singleStage")
    ci.Rule("stages", []string{"stages", "stage"}, "multipleStages")
    ci.Rule("stage", []string{"ETAPA", "STRING", "jobs"}, "createStage")
    ci.Rule("jobs", []string{"job"}, "singleJob")
    ci.Rule("jobs", []string{"jobs", "job"}, "multipleJobs")
    ci.Rule("job", []string{"TRABAJO", "action", "EN", "ID"}, "createJob")
    ci.Rule("action", []string{"CONSTRUIR"}, "buildAction")
    ci.Rule("action", []string{"PROBAR"}, "testAction")
    ci.Rule("action", []string{"DESPLEGAR"}, "deployAction")
    
    // Acciones para CI/CD
    ci.Action("createPipeline", func(args []interface{}) (interface{}, error) {
        name := args[1].(string)
        stages := args[2].([]Stage)
        
        return Pipeline{
            Name:   name,
            Stages: stages,
        }, nil
    })
    
    return ci
}

// DSL de ejemplo:
/*
pipeline mi-app
  etapa "build"
    trabajo construir en ubuntu-latest
    trabajo probar en ubuntu-latest
  etapa "deploy"
    trabajo desplegar en production
*/
```

### Caso 3: DSL para Reglas de Trading

**Desafío**: Sistema de trading algorítmico con reglas en español.

**Solución**:
```go
func createTradingDSL() *dslbuilder.DSL {
    trading := dslbuilder.New("Trading")
    
    // Tokens financieros
    trading.KeywordToken("SI", "si")
    trading.KeywordToken("PRECIO", "precio")
    trading.KeywordToken("VOLUMEN", "volumen")
    trading.KeywordToken("MAYOR", "mayor")
    trading.KeywordToken("MENOR", "menor")
    trading.KeywordToken("COMPRAR", "comprar")
    trading.KeywordToken("VENDER", "vender")
    trading.KeywordToken("ACCIONES", "acciones")
    trading.KeywordToken("ENTONCES", "entonces")
    trading.KeywordToken("Y", "y")
    trading.KeywordToken("RSI", "rsi")
    trading.KeywordToken("MEDIA", "media")
    
    trading.Token("NUMBER", "[0-9]+\\.?[0-9]*")
    trading.Token("SYMBOL", "[A-Z]{2,5}")
    
    // Reglas de trading
    trading.Rule("rule", []string{"SI", "condition", "ENTONCES", "action"}, "tradingRule")
    trading.Rule("condition", []string{"PRECIO", "SYMBOL", "MAYOR", "NUMBER"}, "priceAbove")
    trading.Rule("condition", []string{"RSI", "SYMBOL", "MENOR", "NUMBER"}, "rsiBelow")
    trading.Rule("condition", []string{"condition", "Y", "condition"}, "andCondition")
    trading.Rule("action", []string{"COMPRAR", "NUMBER", "ACCIONES", "SYMBOL"}, "buyAction")
    trading.Rule("action", []string{"VENDER", "NUMBER", "ACCIONES", "SYMBOL"}, "sellAction")
    
    // Acciones de trading
    trading.Action("tradingRule", func(args []interface{}) (interface{}, error) {
        condition := args[1].(TradingCondition)
        action := args[3].(TradingAction)
        
        return TradingRule{
            Condition: condition,
            Action:    action,
            Timestamp: time.Now(),
        }, nil
    })
    
    return trading
}

// Ejemplo de uso:
// "si precio AAPL mayor 150 y rsi AAPL menor 30 entonces comprar 100 acciones AAPL"
```

## 📚 Recursos Adicionales

### Documentación Relacionada
- [Guía de Instalación](instalacion.md) - Setup inicial
- [Guía Rápida](guia_rapida.md) - Conceptos básicos
- [Developer Onboarding](developer_onboarding.md) - Para contribuidores

### Ejemplos Completos
- `examples/contabilidad/` - Sistema contable empresarial
- `examples/accounting/` - Multi-país con contexto
- `examples/simple_context/` - Contexto básico
- `examples/query/` - Consultas LINQ

### Herramientas y Utilidades
```go
// Utilidad para crear DSL con logging
func NewDSLWithLogging(name string) *dslbuilder.DSL {
    dsl := dslbuilder.New(name)
    
    // Wrapper para logging de acciones
    originalAction := dsl.Action
    dsl.Action = func(name string, fn dslbuilder.ActionFunc) {
        wrappedFn := func(args []interface{}) (interface{}, error) {
            start := time.Now()
            result, err := fn(args)
            duration := time.Since(start)
            
            if err != nil {
                log.Printf("Action %s failed in %v: %v", name, duration, err)
            } else {
                log.Printf("Action %s completed in %v", name, duration)
            }
            
            return result, err
        }
        originalAction(name, wrappedFn)
    }
    
    return dsl
}
```

---

**¡Tienes todas las herramientas para crear DSLs empresariales poderosos!** 🚀

*¿Necesitas ayuda con un caso específico? Revisa los ejemplos o crea un issue en GitHub.*