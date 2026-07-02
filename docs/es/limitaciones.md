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

#### Limitación: La firma base de acciones usa `interface{}`
Por compatibilidad, la firma clásica sigue siendo:

```go
func (d *DSL) Action(name string, fn func(args []interface{}) (interface{}, error))
```

**✅ RESUELTO con helpers genéricos integrados** (`Action1/Action2/Action3` y `Args`):

```go
// Adaptadores tipados: convierten y validan los argumentos por vos
dsl.Action("number", dslbuilder.Action1(strconv.Atoi))
dsl.Action("add", dslbuilder.Action3(func(l int, _ string, r int) (int, error) {
    return l + r, nil
}))

// O el helper Args para acceso tipado posicional
dsl.Action("add", func(raw []interface{}) (interface{}, error) {
    args := dslbuilder.Args(raw)
    return args.Int(0) + args.Int(2), nil
})
```

La firma cruda sigue disponible si preferís manejar los tipos manualmente.

## Limitaciones de Diseño

### 1. Gramáticas Ambiguas

#### Limitación: Detección de ambigüedades parcial
Las alternativas usan elección ordenada (estilo PEG), así que una gramática
"ambigua" se resuelve determinísticamente por orden de declaración — pero eso
puede sorprender si declarás la alternativa corta primero.

**✅ Mitigado con el validador integrado en el core**:

```go
warnings, err := dsl.Validate()
// err      -> símbolos desconocidos, tokens inválidos, reglas vacías
// warnings -> reglas inalcanzables, acciones sin registrar,
//             ciclos no productivos, recursión izquierda indirecta
```

```bash
validator -dsl grammar.yaml -strict   # mismo análisis desde la CLI
```

**Regla práctica**: declarar siempre las alternativas más específicas primero,
y usar `Expression()` (Pratt) para precedencia de operadores en vez de reglas
ambiguas.

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

#### ✅ RESUELTO: Farthest-failure tracking
Los errores de sintaxis ya no son genéricos: el parser registra el punto más
lejano alcanzado, qué tokens esperaba ahí y el stack de reglas:

```
no alternative matched for rule condition: expected IDENT or NUMBER, got THEN "then"
rule stack: if_stmt > condition > value at line 1, column 14:
if status == then
             ^
```

```go
if dslbuilder.IsParseError(err) {
    fmt.Println(dslbuilder.GetDetailedError(err)) // posición + expected + puntero
}
```

**También resuelto**:
- `RuleWithError(name, pattern, action, msg)` agrega mensajes personalizados
  por regla (aparecen como `hint:` en el error).
- `Diagnostics(code)` recupera tras cada fallo y reporta TODOS los errores
  de una pasada (es lo que usa `cmd/lsp`).

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

### 2. Características Avanzadas

#### ✅ Ahora soportadas:
- **Recuperación de errores**: `Diagnostics(code)` reporta todos los errores de una pasada (con resincronización por FIRST-set)
- **Streaming**: `ParseStream(io.Reader, handler)` procesa scripts línea a línea sin cargar todo en memoria
- **Generación de código**: `cmd/dslgen` convierte la gramática declarativa en Go versionable
- **LSP**: `cmd/lsp` publica diagnósticos en vivo en el editor

#### Sin soporte:
- **Gramáticas de atributos**: No hay síntesis/herencia automática de atributos
- **Parsing incremental real**: no se reusan árboles entre ediciones (el LSP re-parsea el documento completo, que para DSLs es barato)
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
| Ambigüedad | Elección ordenada (determinista) | ❌ No ambiguo por diseño |
| Recursión izquierda | ✅ Directa (indirecta NO — usar `Expression()`) | ❌ No directamente |
| Backtracking | ✅ Con memoización (packrat) | ✅ Completo |

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

Resueltas en la versión actual:
- [x] ✅ Mensajes de error específicos (farthest-failure con expected + rule stack)
- [x] ✅ Acciones tipadas (`Action1/2/3`, `Args`)
- [x] ✅ Validador integrado en el core (`dsl.Validate()` / `Build()`)
- [x] ✅ AST real inspeccionable (`ParseAST` + `node.Pretty()`)
- [x] ✅ Precedencia confiable de operadores (parser Pratt vía `Expression()`)
- [x] ✅ Tokenizer determinista y lineal (patrones anclados, orden de declaración)
- [x] ✅ Límites de profundidad (sin stack overflow con input hostil)
- [x] ✅ Fuzzing y benchmarks en la suite

Resueltas también en esta versión:
- [x] ✅ Recuperación de errores: `Diagnostics()` reporta múltiples errores por pasada
- [x] ✅ Mensajes de error personalizados por regla: `RuleWithError()`
- [x] ✅ Recursión izquierda indirecta: growing generalizado (el validador aún la señala como candidata a refactor)
- [x] ✅ Streaming orientado a líneas: `ParseStream(io.Reader, handler)`
- [x] ✅ Generación de código: `cmd/dslgen`
- [x] ✅ Servidor LSP mínimo: `cmd/lsp` (diagnósticos en vivo)
- [x] ✅ Detección de alternativas ensombrecidas (elección ordenada PEG) en `Validate()`

Pendientes para futuras versiones:
- [ ] Parsing incremental real (reuso de árboles entre ediciones)
- [ ] Autocompletado/hover en el LSP (hoy: solo diagnósticos)