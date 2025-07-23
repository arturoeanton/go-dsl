# Guía Rápida - go-dsl

**Un potente constructor de lenguajes específicos de dominio (DSL) para Go con soporte completo para gramáticas recursivas por la izquierda y características de nivel empresarial.**

go-dsl es un constructor de lenguajes específicos de dominio (DSL) dinámico para Go que permite crear lenguajes personalizados con sintaxis y reglas de gramática propias. **Ahora con soporte completo para gramáticas recursivas por la izquierda y estabilidad lista para producción.**

## ✨ Características Principales

- 🚀 **Creación Dinámica de DSL**: Construye lenguajes personalizados en tiempo de ejecución
- 🎯 **Sistema de Gramática Avanzado**: Soporte completo para gramáticas recursivas por la izquierda con memoización
- 🔄 **Soporte de Contexto**: Pasa datos dinámicos como el método `q.use()` de r2lang
- 🧠 **Parser Listo para Producción**: Maneja escenarios empresariales complejos con estabilidad
- 🎨 **Prioridad de KeywordToken**: Resuelve conflictos de tokens con coincidencia basada en prioridad
- ⚡ **Alto Rendimiento**: Parsing eficiente con tokenización inteligente

## Instalación

```bash
go get github.com/arturoeanton/go-dsl/pkg/dslbuilder
```

## Ejemplo Básico (Actualizado)

```go
package main

import (
    "fmt"
    "log"
    "github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

func main() {
    // 1. Crear una instancia DSL
    dsl := dslbuilder.New("MiDSL")
    
    // 2. Definir tokens - ¡USA KeywordToken para palabras clave!
    dsl.KeywordToken("HELLO", "hola")  // Prioridad alta (90)
    dsl.KeywordToken("WORLD", "mundo") // Prioridad alta (90)
    
    // 3. Definir reglas de gramática
    dsl.Rule("saludo", []string{"HELLO", "WORLD"}, "procesarSaludo")
    
    // 4. Definir acciones para las reglas
    dsl.Action("procesarSaludo", func(args []interface{}) (interface{}, error) {
        return "¡Hola, mundo!", nil
    })
    
    // 5. Parsear y ejecutar
    result, err := dsl.Parse("hola mundo")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(result.GetOutput()) // ¡Hola, mundo!
}
```

## Paso a Paso para Crear un DSL

### 1. Crear la Instancia DSL

```go
dsl := dslbuilder.New("NombreDeMiDSL")
```

### 2. Definir Tokens

**IMPORTANTE**: Usa `KeywordToken` para palabras clave específicas y `Token` para patrones generales. Los KeywordToken tienen prioridad alta (90) vs prioridad normal (0).

```go
// ✅ CORRECTO: KeywordTokens PRIMERO (prioridad alta)
dsl.KeywordToken("IF", "if")
dsl.KeywordToken("THEN", "entonces")
dsl.KeywordToken("VENTA", "venta")
dsl.KeywordToken("COMPRA", "compra")

// Después, tokens de patrones generales (prioridad normal)
dsl.Token("NUMBER", "[0-9]+")
dsl.Token("IMPORTE", "[0-9]+\\.?[0-9]*")
dsl.Token("STRING", "\"[^\"]*\"")
dsl.Token("ID", "[a-zA-Z][a-zA-Z0-9]*")

// ❌ INCORRECTO: Token genérico capturaría "venta" antes que KeywordToken
// dsl.Token("ID", "[a-zA-Z]+")    // Esto capturaría "venta"
// dsl.KeywordToken("VENTA", "venta")  // Nunca se ejecutaría
```

**¿Por qué KeywordToken?** 
- Resuelve conflictos de tokenización automáticamente
- Garantiza que palabras clave específicas tengan prioridad
- Elimina errores intermitentes de parsing

### 3. Definir Reglas de Gramática

Las reglas definen cómo se combinan los tokens. **Ahora con soporte completo para recursión por la izquierda:**

```go
// Regla simple: una expresión es un número
dsl.Rule("expresion", []string{"NUMBER"}, "numero")

// Regla con operadores: suma de dos números  
dsl.Rule("expresion", []string{"NUMBER", "PLUS", "NUMBER"}, "suma")

// ✨ NUEVO: Reglas recursivas por la izquierda (¡ahora funciona!)
dsl.Rule("movements", []string{"movement"}, "singleMovement")
dsl.Rule("movements", []string{"movements", "movement"}, "multipleMovements")

// Reglas complejas para contabilidad
dsl.Rule("command", []string{"VENTA", "DE", "IMPORTE", "CON", "IVA"}, "saleWithTax")
dsl.Rule("command", []string{"VENTA", "DE", "IMPORTE"}, "simpleSale")

// Condicionales
dsl.Rule("condicional", []string{"IF", "expresion", "THEN", "expresion"}, "si_entonces")
```

**💡 Mejores Prácticas para Reglas:**
- **Reglas más específicas PRIMERO** (patrones más largos)
- Las reglas recursivas por la izquierda ahora funcionan perfectamente
- Usa el ImprovedParser automáticamente para manejar la recursión

### 4. Definir Acciones

Las acciones procesan los tokens capturados por las reglas:

```go
dsl.Action("numero", func(args []interface{}) (interface{}, error) {
    // args[0] contiene el token NUMBER como string
    num, err := strconv.Atoi(args[0].(string))
    return num, err
})

dsl.Action("suma", func(args []interface{}) (interface{}, error) {
    // Para "NUMBER PLUS NUMBER": args[0]=num1, args[1]="+", args[2]=num2
    left := args[0].(int)
    right := args[2].(int)
    return left + right, nil
})

dsl.Action("si_entonces", func(args []interface{}) (interface{}, error) {
    condicion := args[1].(int)
    valor := args[3].(int)
    
    if condicion > 0 {
        return valor, nil
    }
    return 0, nil
})
```

### 5. Usar el DSL

```go
// Parsear una expresión
result, err := dsl.Parse("5 + 3")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Resultado: %v\n", result.GetOutput()) // Resultado: 8
```

## Ejemplos por Dominio

### DSL Contable Empresarial (Listo para Producción)

Este es un ejemplo completo de un sistema contable que demuestra todas las características avanzadas:

```go
func crearDSLContable() *dslbuilder.DSL {
    contabilidad := dslbuilder.New("Contabilidad")
    
    // Tokens con KeywordToken para prioridad
    contabilidad.KeywordToken("VENTA", "venta")
    contabilidad.KeywordToken("COMPRA", "compra")
    contabilidad.KeywordToken("DE", "de")
    contabilidad.KeywordToken("CON", "con")
    contabilidad.KeywordToken("IVA", "iva")
    contabilidad.KeywordToken("ASIENTO", "asiento")
    contabilidad.KeywordToken("DEBE", "debe")
    contabilidad.KeywordToken("HABER", "haber")
    
    // Valores con prioridad normal
    contabilidad.Token("IMPORTE", "[0-9]+\\.?[0-9]*")
    contabilidad.Token("STRING", "\"[^\"]*\"")
    
    // Reglas más específicas primero
    contabilidad.Rule("command", []string{"VENTA", "DE", "IMPORTE", "CON", "IVA"}, "saleWithTax")
    contabilidad.Rule("command", []string{"VENTA", "DE", "IMPORTE"}, "simpleSale")
    
    // ✨ Reglas recursivas por la izquierda para asientos complejos
    contabilidad.Rule("command", []string{"ASIENTO", "movements"}, "processEntry")
    contabilidad.Rule("movements", []string{"movement"}, "singleMovement")
    contabilidad.Rule("movements", []string{"movements", "movement"}, "multipleMovements")
    contabilidad.Rule("movement", []string{"DEBE", "IMPORTE", "IMPORTE"}, "debitMovement")
    contabilidad.Rule("movement", []string{"HABER", "IMPORTE", "IMPORTE"}, "creditMovement")
    
    // Acciones con lógica de negocio
    contabilidad.Action("saleWithTax", func(args []interface{}) (interface{}, error) {
        amount, _ := strconv.ParseFloat(args[2].(string), 64)
        tax := amount * 0.16 // 16% IVA México
        return Transaction{Amount: amount, Tax: tax, Total: amount + tax}, nil
    })
    
    // Procesamiento de asientos complejos
    contabilidad.Action("multipleMovements", func(args []interface{}) (interface{}, error) {
        movements := args[0].([]Movement)
        newMovement := args[1].(Movement)
        return append(movements, newMovement), nil
    })
    
    return contabilidad
}

// Uso:
// "venta de 5000 con iva" → Transaction{Amount: 5000, Tax: 800, Total: 5800}
// "asiento debe 1101 10000 debe 1401 1600 haber 2101 11600" → Asiento balanceado
```

### Sistema Multi-País con Contexto

```go
contabilidad := crearDSLContable()

// Contexto para México
ctxMX := map[string]interface{}{"country": "MX", "iva": 0.16}
result, _ := contabilidad.Use("venta de 1000 con iva", ctxMX)

// Contexto para Colombia  
ctxCOL := map[string]interface{}{"country": "COL", "iva": 0.19}
result, _ := contabilidad.Use("venta de 1000 con iva", ctxCOL)

// El mismo código DSL, diferentes resultados según el contexto
```

### Calculadora Simple

```go
func crearCalculadora() *dslbuilder.DSL {
    calc := dslbuilder.New("Calculadora")
    
    // Tokens
    calc.Token("NUMBER", "[0-9]+")
    calc.Token("PLUS", "\\+")
    calc.Token("MINUS", "-")
    calc.Token("MULTIPLY", "\\*")
    calc.Token("DIVIDE", "/")
    calc.Token("LPAREN", "\\(")
    calc.Token("RPAREN", "\\)")
    
    // Gramática con precedencia
    calc.Rule("expression", []string{"term"}, "passthrough")
    calc.Rule("expression", []string{"expression", "PLUS", "term"}, "add")
    calc.Rule("expression", []string{"expression", "MINUS", "term"}, "subtract")
    
    calc.Rule("term", []string{"factor"}, "passthrough")
    calc.Rule("term", []string{"term", "MULTIPLY", "factor"}, "multiply")
    calc.Rule("term", []string{"term", "DIVIDE", "factor"}, "divide")
    
    calc.Rule("factor", []string{"NUMBER"}, "number")
    calc.Rule("factor", []string{"LPAREN", "expression", "RPAREN"}, "parentheses")
    
    // Acciones
    calc.Action("number", func(args []interface{}) (interface{}, error) {
        return strconv.ParseFloat(args[0].(string), 64)
    })
    
    calc.Action("add", func(args []interface{}) (interface{}, error) {
        left := args[0].(float64)
        right := args[2].(float64)
        return left + right, nil
    })
    
    // ... más acciones
    
    return calc
}
```

### Sistema de Consultas

```go
func crearDSLConsultas() *dslbuilder.DSL {
    query := dslbuilder.New("Consultas")
    
    // Tokens
    query.Token("SELECT", "select|SELECCIONAR")
    query.Token("FROM", "from|DE")
    query.Token("WHERE", "where|DONDE")
    query.Token("FIELD", "[a-zA-Z_][a-zA-Z0-9_]*")
    query.Token("STRING", "\"[^\"]*\"")
    query.Token("NUMBER", "[0-9]+")
    query.Token("EQUALS", "=|==")
    query.Token("GREATER", ">")
    
    // Reglas
    query.Rule("consulta", []string{"SELECT", "FIELD", "FROM", "FIELD"}, "seleccionar")
    query.Rule("consulta", []string{"SELECT", "FIELD", "FROM", "FIELD", "WHERE", "condicion"}, "seleccionarConWhere")
    query.Rule("condicion", []string{"FIELD", "EQUALS", "STRING"}, "igualString")
    query.Rule("condicion", []string{"FIELD", "GREATER", "NUMBER"}, "mayorQue")
    
    return query
}
```

### Lenguaje Comercial (Contabilidad)

```go
func crearDSLContable() *dslbuilder.DSL {
    contable := dslbuilder.New("Contabilidad")
    
    // Tokens específicos del dominio
    contable.Token("VENTA", "venta|VENTA")
    contable.Token("COMPRA", "compra|COMPRA")
    contable.Token("DEBE", "debe|DEBE")
    contable.Token("HABER", "haber|HABER")
    contable.Token("CUENTA", "[0-9]+")
    contable.Token("IMPORTE", "[0-9]+\\.?[0-9]*")
    contable.Token("CON", "con|CON")
    contable.Token("IVA", "iva|IVA")
    
    // Reglas comerciales
    contable.Rule("operacion", []string{"VENTA", "DE", "IMPORTE"}, "ventaSimple")
    contable.Rule("operacion", []string{"VENTA", "DE", "IMPORTE", "CON", "IVA"}, "ventaConIVA")
    contable.Rule("asiento", []string{"DEBE", "CUENTA", "IMPORTE"}, "movimientoDebe")
    contable.Rule("asiento", []string{"HABER", "CUENTA", "IMPORTE"}, "movimientoHaber")
    
    return contable
}
```

## 🔄 Manejo de Contexto Avanzado

go-dsl permite pasar datos dinámicos usando contexto, **exactamente como el método `q.use()` de r2lang**. Esta es una de las características más poderosas para crear DSLs dinámicos empresariales.

### Método 1: Use() - Equivalente a r2lang

```go
// r2lang: q.use("query", {data: myData})
// go-dsl: dsl.Use("query", map[string]interface{}{"data": myData})

dsl := dslbuilder.New("ContextDSL")
dsl.Token("GET", "get")
dsl.Token("VAR", "[a-zA-Z_]+")
dsl.Rule("command", []string{"GET", "VAR"}, "getValue")

dsl.Action("getValue", func(args []interface{}) (interface{}, error) {
    varName := args[1].(string)
    value := dsl.GetContext(varName)
    return value, nil
})

// Pasar contexto con Use() - como r2lang
context := map[string]interface{}{
    "name":  "Juan García",
    "age":   30,
    "city":  "Madrid",
}

result, err := dsl.Use("get name", context)
// result.GetOutput() -> "Juan García"
```

### Método 2: SetContext() - Para valores persistentes

```go
// Establecer contexto que persiste entre llamadas
dsl.SetContext("usuario", "Juan")
dsl.SetContext("moneda", "EUR")

// Usar en acciones
dsl.Action("procesarVenta", func(args []interface{}) (interface{}, error) {
    usuario := dsl.GetContext("usuario")
    moneda := dsl.GetContext("moneda")
    
    return fmt.Sprintf("Venta registrada por %s en %s", usuario, moneda), nil
})

result, err := dsl.Parse("procesar venta")
```

### Ejemplo Complejo: Datos Dinámicos

```go
type Person struct {
    Name string
    Age  int
    City string
}

dsl := dslbuilder.New("DataQuery")
dsl.Token("FIND", "find")
dsl.Token("FIELD", "name|age|city")
dsl.Token("IN", "in")
dsl.Token("DATASET", "[a-zA-Z_]+")
dsl.Rule("query", []string{"FIND", "FIELD", "IN", "DATASET"}, "findField")

dsl.Action("findField", func(args []interface{}) (interface{}, error) {
    field := args[1].(string)
    dataset := args[3].(string)
    
    data := dsl.GetContext(dataset)
    people := data.([]Person)
    
    var results []string
    for _, person := range people {
        switch field {
        case "name": results = append(results, person.Name)
        case "age": results = append(results, strconv.Itoa(person.Age))
        case "city": results = append(results, person.City)
        }
    }
    return results, nil
})

// Datos dinámicos
people := []Person{
    {"Juan García", 28, "Madrid"},
    {"María López", 35, "Barcelona"},
}

// Consulta con contexto dinámico
context := map[string]interface{}{"people": people}
result, err := dsl.Use("find name in people", context)
// result.GetOutput() -> ["Juan García", "María López"]
```

### Comparación r2lang vs go-dsl

| r2lang | go-dsl |
|--------|--------|
| `q.use("query", {data: myData})` | `dsl.Use("query", map[string]interface{}{"data": myData})` |
| `context.data` | `dsl.GetContext("data")` |
| Automático | Requiere type assertion: `data.([]MyType)` |

## 💡 Mejores Prácticas (Actualizadas)

### 1. ¡USA SIEMPRE KeywordToken para Palabras Clave!

**Esta es la regla #1 más importante para evitar errores de parsing:**

```go
// ✅ CORRECTO - KeywordToken tiene prioridad automática
dsl.KeywordToken("VENTA", "venta")     // Prioridad 90
dsl.KeywordToken("COMPRA", "compra")   // Prioridad 90  
dsl.KeywordToken("CON", "con")         // Prioridad 90
dsl.Token("ID", "[a-zA-Z]+")           // Prioridad 0

// ❌ INCORRECTO - Token genérico captura palabras clave
dsl.Token("ID", "[a-zA-Z]+")     // Capturaría "venta", "compra", "con"
dsl.Token("VENTA", "venta")      // Nunca se ejecutaría
```

**¿Por qué es tan importante?**
- Elimina errores intermitentes de parsing
- No depende del orden de definición
- Funciona 100% del tiempo sin excepciones

### 2. Instancias DSL Frescas para Estabilidad

**Para máxima estabilidad, especialmente en sistemas contables, crea instancias DSL frescas:**

```go
// ✅ RECOMENDADO para sistemas críticos
func procesarComando(comando string) (interface{}, error) {
    // Nueva instancia para cada operación
    contabilidad := createContabilidadDSL(sistema)
    return contabilidad.Parse(comando)
}

// ❌ Puede causar problemas de estado en sistemas complejos
var globalDSL = createContabilidadDSL(sistema)
func procesarComando(comando string) (interface{}, error) {
    return globalDSL.Parse(comando)  // Reutiliza la misma instancia
}
```

### 3. Manejo de Errores
Siempre maneja errores en las acciones:

```go
dsl.Action("dividir", func(args []interface{}) (interface{}, error) {
    left := args[0].(float64)
    right := args[2].(float64)
    
    if right == 0 {
        return nil, fmt.Errorf("división por cero")
    }
    
    return left / right, nil
})
```

### 3. Validación de Tipos
Usa aserciones de tipo seguras:

```go
dsl.Action("procesar", func(args []interface{}) (interface{}, error) {
    if len(args) < 1 {
        return nil, fmt.Errorf("argumentos insuficientes")
    }
    
    valor, ok := args[0].(string)
    if !ok {
        return nil, fmt.Errorf("se esperaba string, se recibió %T", args[0])
    }
    
    return strings.ToUpper(valor), nil
})
```

### 4. Testing
Escribe tests para tu DSL:

```go
func TestMiDSL(t *testing.T) {
    dsl := crearMiDSL()
    
    tests := []struct {
        input    string
        expected interface{}
        hasError bool
    }{
        {"hola mundo", "¡Hola, mundo!", false},
        {"hello world", "¡Hola, mundo!", false},
        {"syntax error", nil, true},
    }
    
    for _, tt := range tests {
        result, err := dsl.Parse(tt.input)
        
        if tt.hasError {
            assert.Error(t, err)
        } else {
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, result.GetOutput())
        }
    }
}
```

## 🎯 Casos de Uso Empresariales

### Casos de Éxito Comprobados

- **✅ Sistemas Contables Empresariales**: Cálculo de IVA multi-país, asientos contables complejos
- **✅ DSL para Reglas de Negocio**: Pricing dinámico, descuentos, comisiones
- **✅ Sistemas de Consulta LINQ**: Consultas tipo SQL en español con contexto dinámico
- **✅ Calculadoras Especializadas**: Financieras, científicas, actuariales
- **✅ Lenguajes de Configuración**: Para aplicaciones empresariales complejas
- **✅ Procesadores de Comandos**: CLI empresariales con gramática natural
- **✅ Sistemas de Workflow**: Automatización con reglas recursivas complejas

### Nuevas Capacidades Empresariales

- **🔥 Gramáticas Recursivas por la Izquierda**: Para estructuras complejas como asientos contables
- **🔥 Contexto Dinámico**: Como r2lang, para datos que cambian en tiempo real
- **🔥 Multi-País/Multi-Moneda**: Misma gramática, diferentes contextos fiscales
- **🔥 Estabilidad de Producción**: Sin errores intermitentes, listo para sistemas críticos

## 📚 Recursos Adicionales

- **Ejemplos Empresariales Completos**: `/examples/contabilidad/`, `/examples/accounting/`
- **Tests Unitarios**: `/pkg/dslbuilder/dsl_test.go` 
- **Documentación de Mejoras**: `docs/es/propuesta_de_mejoras.md`
- **README en Inglés**: Documentación completa con ejemplos multi-país

## ⚡ Empezar Ahora

```bash
# Clona y prueba los ejemplos
git clone https://github.com/arturoeanton/go-dsl
cd go-dsl

# Prueba el sistema contable empresarial
go run examples/contabilidad/main.go

# Prueba el sistema multi-país
go run examples/accounting/main.go
```

**¡Tu DSL empresarial está a solo unos minutos de distancia!**

---

**Última actualización**: 2025-07-23 - Con soporte completo para gramáticas recursivas por la izquierda y estabilidad de producción.