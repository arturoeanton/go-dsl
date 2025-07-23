# Developer Onboarding - go-dsl

**Guía para contribuidores: Entiende la arquitectura interna de go-dsl y contribuye valor desde el primer día.**

Esta documentación está diseñada para desarrolladores que quieren contribuir al proyecto go-dsl. Te ayudará a entender la arquitectura interna, patrones de diseño y cómo agregar nuevas características de manera efectiva.

## 🎯 Objetivo de Esta Guía

- Reducir la curva de aprendizaje de nuevos contribuidores
- Explicar decisiones de arquitectura y patrones usados
- Mostrar cómo el framework resuelve problemas técnicos complejos
- Guiar la implementación de nuevas características

## 🏗️ Arquitectura General

### Visión de Alto Nivel

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   User API      │───▶│   DSL Builder    │───▶│   Parser        │
│                 │    │                  │    │                 │
│ dsl.KeywordToken│    │ - Tokens         │    │ - Grammar       │
│ dsl.Rule        │    │ - Rules          │    │ - Tokenization  │
│ dsl.Action      │    │ - Actions        │    │ - Parsing       │
│ dsl.Parse       │    │ - Context        │    │ - AST           │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │
                                ▼
                    ┌──────────────────┐
                    │   Action Engine  │
                    │                  │
                    │ - Execute        │
                    │ - Context Access │
                    │ - Error Handling │
                    └──────────────────┘
```

### Estructura de Directorios

```
go-dsl/
├── pkg/dslbuilder/           # 🏗️ Core del framework
│   ├── dsl.go               # DSL principal + API pública
│   ├── parser.go            # Parser básico
│   ├── improved_parser.go   # Parser con recursión izquierda
│   ├── tokenizer.go         # Tokenización y regex
│   ├── grammar.go           # Reglas gramaticales
│   ├── action.go            # Sistema de acciones
│   └── dsl_test.go          # Tests unitarios
├── examples/                # 📚 Casos de uso reales
│   ├── contabilidad/        # Sistema contable complejo
│   ├── accounting/          # Multi-país empresarial
│   ├── simple_context/      # Contexto básico
│   └── query/               # Consultas LINQ
└── docs/es/                 # 📖 Documentación
    ├── guia_rapida.md
    ├── instalacion.md
    ├── manual_de_uso.md
    └── developer_onboarding.md  # 👈 Este documento
```

## 🔍 Análisis de Componentes Clave

### 1. DSL Core (`dsl.go`)

**Responsabilidad**: API principal y orquestación de componentes.

```go
type DSL struct {
    name     string
    tokens   []TokenDefinition      // Definiciones de tokens
    rules    []Rule                 // Reglas gramaticales
    actions  map[string]ActionFunc  // Funciones de acción
    context  map[string]interface{} // Contexto dinámico
    grammar  *Grammar               // Gramática compilada
    funcs    map[string]interface{} // Funciones Go registradas
}
```

**Decisiones de Diseño:**
- **¿Por qué struct en lugar de interface?** Simplicidad para usuarios, flexibilidad interna
- **¿Por qué map para actions?** Lookup O(1) y registro dinámico
- **¿Por qué grammar separada?** Permite optimizaciones y caching

**Patrones Implementados:**
- **Builder Pattern**: `dsl.KeywordToken().Rule().Action()`
- **Strategy Pattern**: Diferentes parsers según complejidad
- **Context Pattern**: Datos dinámicos como r2lang

### 2. Sistema de Tokenización (`tokenizer.go`)

**Problema Resuelto**: Conflictos de tokens (ej: "venta" vs patrón `[a-zA-Z]+`)

```go
type TokenDefinition struct {
    Name     string
    Pattern  string
    Priority int    // 🎯 CLAVE: KeywordToken=90, Token=0
    Regex    *regexp.Regexp
}

type TokenMatch struct {
    TokenType string
    Value     string
    Position  int
}
```

**Solución Elegante: Sistema de Prioridades**
```go
// ✅ KeywordToken tiene prioridad automática
func (d *DSL) KeywordToken(name, pattern string) {
    d.tokens = append(d.tokens, TokenDefinition{
        Name:     name,
        Pattern:  pattern,
        Priority: 90,  // 🔥 Alta prioridad
        Regex:    regexp.MustCompile("^(" + pattern + ")"),
    })
}

// Token normal tiene prioridad baja
func (d *DSL) Token(name, pattern string) {
    // Priority: 0 (default)
}
```

**¿Por qué funciona?**
1. Tokens se ordenan por prioridad antes de matching
2. KeywordToken siempre gana vs Token genérico
3. Elimina errores intermitentes sin depender del orden de definición

### 3. Parsers Duales (`parser.go` + `improved_parser.go`)

**Problema**: Gramáticas recursivas por la izquierda causan stack overflow en parsers descendentes.

**Solución**: Dos parsers especializados con selección automática.

#### Parser Básico (Descendente Recursivo)
```go
type Parser struct {
    grammar *Grammar
    tokens  []TokenMatch
    pos     int
}

func (p *Parser) parseRule(ruleName string) (interface{}, error) {
    // Descendente recursivo simple
    // ❌ Falla con: movements -> movements movement
}
```

#### ImprovedParser (Con Memoización)
```go
type ImprovedParser struct {
    grammar *Grammar
    tokens  []TokenMatch
    pos     int
    memo    map[string]map[int]ParseResult  // 🎯 CLAVE: Memoización
}

func (p *ImprovedParser) parseRule(ruleName string) (interface{}, error) {
    // ✅ Maneja recursión izquierda con packrat parsing
    key := fmt.Sprintf("%s_%d", ruleName, p.pos)
    if result, exists := p.memo[ruleName][p.pos]; exists {
        return result.Value, result.Error
    }
    // ... parsing con memoización
}
```

**¿Cuándo usar cada uno?**
- **Parser básico**: Gramáticas simples, mejor rendimiento
- **ImprovedParser**: Gramáticas recursivas, casos complejos
- **Selección automática**: El DSL decide internamente

### 4. Sistema de Gramática (`grammar.go`)

**Responsabilidad**: Compilar reglas de usuario en estructura optimizada.

```go
type Grammar struct {
    Rules map[string][]Alternative  // Regla -> Lista alternativas
}

type Alternative struct {
    Sequence []string  // Secuencia de tokens
    Action   string    // Nombre de acción
}

type Rule struct {
    Name        string
    Sequence    []string
    Action      string
    Priority    int      // Para ordenar alternativas
}
```

**Optimizaciones Implementadas:**
1. **Indexación por nombre**: Rules map para O(1) lookup
2. **Ordenación por especificidad**: Alternativas más largas primero
3. **Validación en compile-time**: Detectar referencias faltantes

### 5. Sistema de Acciones (`action.go`)

**Problema**: Conectar parsing con lógica de negocio de manera flexible.

```go
type ActionFunc func(args []interface{}) (interface{}, error)

// Registro dinámico
func (d *DSL) Action(name string, fn ActionFunc) {
    d.actions[name] = fn
}

// Ejecución con contexto
func (d *DSL) executeAction(actionName string, args []interface{}) (interface{}, error) {
    if action, exists := d.actions[actionName]; exists {
        // 🎯 Aquí el contexto está disponible via d.context
        return action(args)
    }
    return nil, fmt.Errorf("action '%s' not found", actionName)
}
```

**Patrones de Uso Común:**
```go
// Patrón 1: Transformación simple
dsl.Action("add", func(args []interface{}) (interface{}, error) {
    left := args[0].(int)
    right := args[2].(int)  // args[1] es el operador "+"
    return left + right, nil
})

// Patrón 2: Acceso a contexto
dsl.Action("saleWithTax", func(args []interface{}) (interface{}, error) {
    amount := parseFloat(args[2])
    country := dsl.GetContext("country").(string)
    taxRate := getTaxRate(country)
    return Transaction{Amount: amount, Tax: amount * taxRate}, nil
})

// Patrón 3: Validación y estado
dsl.Action("balancedEntry", func(args []interface{}) (interface{}, error) {
    movements := args[1].([]Movement)
    totalDebit := sumDebits(movements)
    totalCredit := sumCredits(movements)
    if totalDebit != totalCredit {
        return nil, fmt.Errorf("unbalanced entry: %.2f != %.2f", totalDebit, totalCredit)
    }
    return createEntry(movements), nil
})
```

## 🔧 Patrones de Contribución

### 1. Agregar Nueva Funcionalidad de Token

**Ejemplo**: Implementar `RegexToken` con validación automática.

```go
// En dsl.go
func (d *DSL) RegexToken(name, pattern string, validator func(string) bool) {
    compiled, err := regexp.Compile("^(" + pattern + ")")
    if err != nil {
        panic(fmt.Sprintf("Invalid regex pattern for token %s: %v", name, err))
    }
    
    d.tokens = append(d.tokens, TokenDefinition{
        Name:      name,
        Pattern:   pattern,
        Priority:  50,  // Prioridad media
        Regex:     compiled,
        Validator: validator,  // 🆕 Nueva característica
    })
}

// En tokenizer.go - modificar tokenización
func (t *Tokenizer) tokenize(input string) ([]TokenMatch, error) {
    // ... código existente ...
    
    if token.Validator != nil && !token.Validator(value) {
        continue  // Skip si no pasa validación
    }
    
    // ... resto del código ...
}
```

### 2. Agregar Nuevo Tipo de Parser

**Ejemplo**: Parser para gramáticas con precedencia de operadores.

```go
// Crear nuevo archivo: precedence_parser.go
type PrecedenceParser struct {
    grammar    *Grammar
    tokens     []TokenMatch
    pos        int
    precedence map[string]int  // Token -> precedencia
}

func NewPrecedenceParser(grammar *Grammar) *PrecedenceParser {
    return &PrecedenceParser{
        grammar:    grammar,
        precedence: make(map[string]int),
    }
}

func (p *PrecedenceParser) SetPrecedence(token string, level int) {
    p.precedence[token] = level
}

// En dsl.go - agregar soporte
func (d *DSL) SetOperatorPrecedence(token string, level int) {
    // Inicializar precedence parser si no existe
    if d.precedenceParser == nil {
        d.precedenceParser = NewPrecedenceParser(d.grammar)
    }
    d.precedenceParser.SetPrecedence(token, level)
}
```

### 3. Extender Sistema de Contexto

**Ejemplo**: Contexto con scoping y herencia.

```go
// Nuevo tipo de contexto
type ScopedContext struct {
    scopes []map[string]interface{}
    current int
}

func (sc *ScopedContext) PushScope() {
    sc.scopes = append(sc.scopes, make(map[string]interface{}))
    sc.current++
}

func (sc *ScopedContext) PopScope() {
    if sc.current > 0 {
        sc.scopes = sc.scopes[:sc.current]
        sc.current--
    }
}

func (sc *ScopedContext) Get(key string) interface{} {
    // Buscar en orden inverso (scope actual primero)
    for i := sc.current; i >= 0; i-- {
        if value, exists := sc.scopes[i][key]; exists {
            return value
        }
    }
    return nil
}

// En dsl.go
func (d *DSL) PushScope() {
    if d.scopedContext == nil {
        d.scopedContext = &ScopedContext{scopes: []map[string]interface{}{{}}}
    }
    d.scopedContext.PushScope()
}
```

### 4. Agregar Herramientas de Debug

**Ejemplo**: Profiler de parsing con métricas.

```go
// Nuevo archivo: profiler.go
type ParseProfiler struct {
    ruleStats    map[string]*RuleStats
    tokenStats   map[string]*TokenStats
    startTime    time.Time
    enabled      bool
}

type RuleStats struct {
    Name        string
    CallCount   int
    TotalTime   time.Duration
    AverageTime time.Duration
    Errors      int
}

func (d *DSL) EnableProfiling() {
    d.profiler = &ParseProfiler{
        ruleStats:  make(map[string]*RuleStats),
        tokenStats: make(map[string]*TokenStats),
        enabled:    true,
    }
}

func (d *DSL) GetProfilingReport() *ProfilingReport {
    if d.profiler == nil {
        return nil
    }
    return d.profiler.GenerateReport()
}

// Instrumentar parsers existentes
func (p *Parser) parseRule(ruleName string) (interface{}, error) {
    if p.dsl.profiler != nil && p.dsl.profiler.enabled {
        start := time.Now()
        defer func() {
            p.dsl.profiler.RecordRule(ruleName, time.Since(start))
        }()
    }
    
    // ... código de parsing existente ...
}
```

## 🧪 Testing y Calidad

### Estructura de Tests

```go
// dsl_test.go - Patrones de testing
func TestKeywordTokenPriority(t *testing.T) {
    dsl := New("TestPriority")
    
    // Setup: El orden no debería importar
    dsl.Token("ID", "[a-zA-Z]+")      // Prioridad 0
    dsl.KeywordToken("VENTA", "venta") // Prioridad 90
    
    // Test: KeywordToken siempre gana
    tokens, err := dsl.DebugTokens("venta")
    assert.NoError(t, err)
    assert.Equal(t, "VENTA", tokens[0].TokenType)
    assert.NotEqual(t, "ID", tokens[0].TokenType)
}

func TestLeftRecursiveGrammar(t *testing.T) {
    dsl := New("TestRecursion")
    
    // Setup: Gramática recursiva izquierda
    dsl.KeywordToken("DEBE", "debe")
    dsl.Token("AMOUNT", "[0-9]+")
    dsl.Rule("movements", []string{"movement"}, "single")
    dsl.Rule("movements", []string{"movements", "movement"}, "multiple")
    dsl.Rule("movement", []string{"DEBE", "AMOUNT"}, "debit")
    
    // Test: Debe usar ImprovedParser automáticamente
    result, err := dsl.Parse("debe 1000 debe 2000 debe 3000")
    assert.NoError(t, err)
    movements := result.GetOutput().([]Movement)
    assert.Len(t, movements, 3)
}
```

### Benchmarking

```go
func BenchmarkSimpleParsing(b *testing.B) {
    dsl := setupSimpleDSL()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := dsl.Parse("simple expression")
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkComplexParsing(b *testing.B) {
    dsl := setupAccountingDSL()
    command := "asiento debe 1101 10000 debe 1401 1600 haber 2101 11600"
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := dsl.Parse(command)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## 🚀 Roadmap Técnico para Contribuidores

### Prioridad Alta (Ready to Implement)

1. **Mejor Manejo de Errores**
   - Ubicación específica de errores (línea/columna)
   - Error recovery parcial
   - Sugerencias de corrección

2. **Optimizaciones de Performance**
   - Token caching
   - Grammar compilation optimizations
   - Memory pooling para parsers

### Prioridad Media (Research Needed)

1. **Herramientas de Desarrollo**
   - Grammar visualizer
   - Interactive debugger
   - Performance profiler

2. **Extensiones de Gramática**
   - Operator precedence
   - Associativity rules
   - Kleene star operators (`*`, `+`, `?`)

### Prioridad Baja (Future Vision)

1. **Code Generation**
   - Generate optimized parsers
   - Go struct generation from DSL
   - Language server protocol

2. **Advanced Features**
   - Parallel parsing
   - Streaming for large inputs
   - Plugin system

## 📋 Checklist para Nuevos Contribuidores

### Antes de Empezar
- [ ] Leer esta guía completa
- [ ] Ejecutar todos los ejemplos
- [ ] Correr tests: `go test ./pkg/dslbuilder/...`
- [ ] Entender KeywordToken vs Token
- [ ] Probar gramáticas recursivas

### Para Cada Contribución
- [ ] Escribir tests que cubran el caso
- [ ] Mantener backward compatibility
- [ ] Actualizar documentación si es necesario
- [ ] Seguir patrones existentes
- [ ] Benchmark si afecta performance

### Code Review Checklist
- [ ] ¿La API es consistente con el resto?
- [ ] ¿Los tests cubren edge cases?
- [ ] ¿Hay impacto en performance?
- [ ] ¿La documentación está actualizada?
- [ ] ¿Funciona con ambos parsers?

## 🎓 Conceptos Clave para Entender

### 1. ¿Por qué KeywordToken funciona?
No es magia - es **ordenación por prioridad** antes de matching.

### 2. ¿Cómo maneja recursión izquierda?
**Memoización (packrat parsing)** - cachea resultados parciales para evitar loops infinitos.

### 3. ¿Por qué instancias frescas?
**Aislamiento de estado** - elimina condiciones de carrera en sistemas concurrentes.

### 4. ¿Cómo funciona el contexto dinámico?
**Inyección de dependencias** - las acciones acceden al contexto via closure.

## 📞 Soporte para Contribuidores

- **Código**: Revisar ejemplos en `examples/`
- **Tests**: Patrones en `pkg/dslbuilder/dsl_test.go`
- **Issues**: Reportar en GitHub con label "contributor-help"
- **Arquitectura**: Este documento + código comentado

---

**¡Bienvenido al equipo! Tu contribución ayudará a que go-dsl sea aún mejor.** 🚀

*¿Tienes dudas sobre algún patrón o decisión de diseño? Abre un issue con tag `architecture-question`.*