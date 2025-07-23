# Limitaciones de go-dsl

Este documento describe las limitaciones conocidas de go-dsl y las alternativas o soluciones recomendadas para cada caso.

## Tabla de Contenidos
1. [Limitaciones del Lenguaje Go](#limitaciones-del-lenguaje-go)
2. [Limitaciones de Diseño](#limitaciones-de-diseño)
3. [Limitaciones de Performance](#limitaciones-de-performance)
4. [Limitaciones de Funcionalidad](#limitaciones-de-funcionalidad)
5. [Comparación con Otras Herramientas](#comparación-con-otras-herramientas)
6. [Soluciones y Alternativas](#soluciones-y-alternativas)

## Limitaciones del Lenguaje Go

### 1. Expresiones Regulares

#### Limitación: No soporta lookbehind
Go usa el paquete `regexp` que implementa RE2, el cual no soporta:
- Lookbehind positivo `(?<=...)`
- Lookbehind negativo `(?<!...)`
- Backreferences `\1`, `\2`, etc.

**Impacto en go-dsl**:
```go
// Esto NO funcionará:
dsl.Token("UNIT_AFTER_NUMBER", "(?<=[0-9])px")  // Error: invalid regex
```

**Solución actual**:
```go
// Usar prioridad de tokens
dsl.KeywordToken("PX", "px")     // Alta prioridad
dsl.Token("NUMBER", "[0-9]+")    // Baja prioridad

// O procesar en el parser
dsl.Rule("measurement", []string{"NUMBER", "PX"}, "numberWithUnit")
```

#### Limitación: No soporta modo PCRE completo
- No hay grupos con nombre `(?P<name>...)`
- No hay condicionales `(?(condition)yes|no)`
- No hay recursión de patrones

### 2. Sistema de Tipos

#### Limitación: No hay genéricos para versiones < Go 1.18
Para mantener compatibilidad con Go 1.18+, go-dsl usa `interface{}`:

```go
// API actual
func (d *DSL) Action(name string, fn func(args []interface{}) (interface{}, error))

// Con genéricos sería más seguro:
// func (d *DSL) Action[T any](name string, fn func(args []T) (T, error))
```

**Impacto**: 
- Necesidad de type assertions
- Posibles errores en runtime
- Menos ayuda del compilador

**Mitigación**:
```go
// Funciones helper para conversión segura
func toInt(v interface{}) int {
    switch n := v.(type) {
    case int:
        return n
    case string:
        i, _ := strconv.Atoi(n)
        return i
    default:
        return 0
    }
}
```

## Limitaciones de Diseño

### 1. Gramáticas Ambiguas

#### Limitación: No detecta ambigüedades automáticamente
go-dsl no analiza estáticamente la gramática para detectar ambigüedades:

```go
// Gramática ambigua - go-dsl la acepta sin advertencias
dsl.Rule("expr", []string{"expr", "PLUS", "expr"}, "add")
dsl.Rule("expr", []string{"expr", "MINUS", "expr"}, "subtract")
// Sin precedencia, "1 + 2 - 3" es ambiguo
```

**Solución**: Usar el validador de gramática:
```bash
validator -dsl grammar.yaml -strict
```

### 2. Análisis Sintáctico

#### Limitación: Parser LL con backtracking limitado
- No es un parser LR completo
- Puede ser menos eficiente para gramáticas muy complejas
- El backtracking puede causar comportamiento exponencial en casos patológicos

**Ejemplo problemático**:
```go
// Gramática con mucho backtracking
dsl.Rule("A", []string{"B", "C", "D", "E"}, "a1")
dsl.Rule("A", []string{"B", "C", "D", "F"}, "a2")
dsl.Rule("A", []string{"B", "C", "G", "H"}, "a3")
// El parser probará todas las alternativas
```

### 3. Manejo de Errores

#### Limitación: Mensajes de error genéricos en gramáticas complejas
Para gramáticas muy anidadas, los mensajes pueden ser poco específicos:

```
Error: no alternative matched for rule expr at line 1, column 15
```

**Mejora propuesta** (no implementada):
```go
// Errores personalizados por regla
dsl.RuleWithError("expr", pattern, action, "Expected expression after operator")
```

## Limitaciones de Performance

### 1. Memoización

#### Limitación: Uso de memoria en textos largos
La memoización (packrat parsing) intercambia memoria por velocidad:

- **Complejidad temporal**: O(n) para gramáticas sin ambigüedad
- **Complejidad espacial**: O(n × r) donde r = número de reglas

**Para textos muy largos** (>1MB):
```go
// Considerar streaming o chunking
chunks := splitIntoChunks(largeText, 1024*100) // 100KB chunks
for _, chunk := range chunks {
    result, _ := dsl.Parse(chunk)
}
```

### 2. Compilación de Gramáticas

#### Limitación: No hay compilación a código nativo
Otras herramientas como ANTLR generan código compilado. go-dsl interpreta en runtime.

**Impacto**:
- Mayor overhead por interpretación
- Menor velocidad en parsing intensivo

**Comparación de velocidad** (aproximada):
```
ANTLR (Java generado): 100%
Yacc/Bison (C generado): 95%
go-dsl (interpretado): 40-60%
```

## Limitaciones de Funcionalidad

### 1. Análisis Semántico

#### Limitación: No hay sistema de tipos integrado
go-dsl no valida tipos durante el parsing:

```go
// go-dsl acepta esto sintácticamente
// La validación semántica debe ser manual
"variable = 'string' + 123"
```

**Solución**: Implementar validación en las acciones:
```go
dsl.Action("add", func(args []interface{}) (interface{}, error) {
    // Validación manual de tipos
    if !isNumber(args[0]) || !isNumber(args[2]) {
        return nil, fmt.Errorf("type error: cannot add non-numbers")
    }
    // ...
})
```

### 2. Características Avanzadas No Soportadas

#### No hay soporte para:
- **Gramáticas de atributos**: No hay síntesis/herencia automática de atributos
- **Parsing incremental**: Todo el texto debe parsearse completamente
- **Recuperación de errores**: El parsing se detiene en el primer error
- **Múltiples archivos**: No hay sistema de módulos/imports integrado

### 3. Debugging Limitado

#### Limitación: No hay debugger paso a paso
No existe una forma integrada de debuggear el proceso de parsing paso a paso.

**Alternativas actuales**:
```go
// 1. Usar el modo debug del AST viewer
ast_viewer -dsl grammar.yaml -input "test" -format debug

// 2. Agregar logs en las acciones
dsl.Action("rule", func(args []interface{}) (interface{}, error) {
    fmt.Printf("DEBUG: rule matched with args: %v\n", args)
    return result, nil
})
```

## Comparación con Otras Herramientas

### vs ANTLR
| Característica | go-dsl | ANTLR |
|----------------|---------|--------|
| Generación de código | ❌ Interpretado | ✅ Genera código |
| Velocidad | Moderada | Alta |
| Curva de aprendizaje | ✅ Fácil | ❌ Compleja |
| Gramáticas soportadas | CFG + left-rec | LL(*) |
| IDE support | Básico | ✅ Extensivo |
| Múltiples lenguajes | ❌ Solo Go | ✅ Multi-target |

### vs Yacc/Bison
| Característica | go-dsl | Yacc/Bison |
|----------------|---------|------------|
| Tipo de parser | LL + memoización | LALR(1) |
| Conflictos S/R | Resuelve con precedencia | Reporta warnings |
| API | ✅ Go idiomático | C-style |
| Debugging | Moderado | ✅ Extensivo |

### vs PEG (Parsing Expression Grammars)
| Característica | go-dsl | PEG |
|----------------|---------|-----|
| Ambigüedad | Posible | ❌ No ambiguo por diseño |
| Recursión izquierda | ✅ Soportada | ❌ No directamente |
| Backtracking | Limitado | ✅ Completo |

## Soluciones y Alternativas

### Para Alto Rendimiento
Si necesitas máximo rendimiento:
1. **Genera código con go-dsl**: Usa go-dsl para prototipar, luego genera un parser manual
2. **Usa goyacc**: Para gramáticas LALR complejas
3. **Combina con regex**: Pre-tokeniza con regex optimizadas

### Para Gramáticas Complejas
Si tu gramática es muy compleja:
1. **Simplifica la gramática**: Refactoriza para reducir ambigüedad
2. **Usa múltiples pasadas**: Tokenización → Parsing → Validación semántica
3. **Considera ANTLR**: Para gramáticas que requieren LL(*)

### Para Debugging Avanzado
Para mejor debugging:
1. **Instrumenta las acciones**: Agrega logging detallado
2. **Usa el AST viewer**: Visualiza la estructura parseada
3. **Tests unitarios**: Prueba reglas individuales

### Para Validación de Tipos
Implementa un sistema de tipos sobre go-dsl:
```go
type TypeChecker struct {
    symbols map[string]Type
}

func (tc *TypeChecker) checkExpression(ast interface{}) (Type, error) {
    // Implementación del type checker
}
```

## Conclusión

go-dsl está diseñado para ser una herramienta práctica y fácil de usar para crear DSLs en Go. Sus limitaciones son principalmente:

1. **Trade-offs de diseño**: Simplicidad sobre características avanzadas
2. **Limitaciones de Go**: Especialmente en regex
3. **Enfoque interpretado**: Flexibilidad sobre rendimiento máximo

Para la mayoría de casos de uso (DSLs de configuración, lenguajes de dominio específico, calculadoras, procesadores de reglas), estas limitaciones no son significativas. Para casos que requieren máximo rendimiento o características muy avanzadas, considera combinar go-dsl con otras herramientas o usarlo como prototipo antes de una implementación más específica.

## Roadmap de Mejoras

Algunas limitaciones podrían abordarse en futuras versiones:
- [ ] Soporte para streaming/parsing incremental
- [ ] Mejor recuperación de errores
- [ ] Optimizaciones de rendimiento
- [ ] Debugger integrado
- [ ] Generación de código opcional