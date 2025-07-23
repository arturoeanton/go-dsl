# Demostración de Errores Mejorados - Error Demo en Español

**Ejemplo que demuestra el sistema de errores mejorado de go-dsl con información de línea y columna, manteniendo 100% compatibilidad hacia atrás.**

## 🎯 Objetivo

Este ejemplo demuestra el **sistema de errores mejorado** de go-dsl, mostrando:

- 🎯 Errores con información de línea y columna específicas
- 📊 Compatibilidad completa con código existente
- 🔧 Tipo `ParseError` con contexto detallado
- 🔄 Funciones helper para manejo de errores
- 📝 Ejemplos de diferentes tipos de errores de parsing

## 🚀 Ejecución Rápida

```bash
cd examples/error_demo
go run main.go
```

## ✨ Nuevas Características de Error

### Tipo ParseError

```go
type ParseError struct {
    Message  string  // Mensaje de error descriptivo
    Line     int     // Número de línea (1-based)
    Column   int     // Número de columna (1-based) 
    Position int     // Posición absoluta en el input
    Token    string  // Token que causó el error
    Input    string  // Input completo para contexto
}
```

### Funciones Helper

```go
// Verificar si un error es ParseError
func IsParseError(err error) bool

// Obtener información detallada del error
func GetDetailedError(err error) string
```

## 📚 Compatibilidad hacia Atrás

### Código Existente Sigue Funcionando

```go
// ✅ Código existente - SIN CAMBIOS
result, err := dsl.Parse("comando inválido")
if err != nil {
    fmt.Println("Error:", err.Error())
    // Funciona exactamente igual que antes
}
```

### Código Nuevo Puede Usar Características Mejoradas

```go
// ✅ Código nuevo - CON características mejoradas
result, err := dsl.Parse("comando inválido")
if err != nil {
    if IsParseError(err) {
        parseErr := err.(*ParseError)
        fmt.Printf("Error en línea %d, columna %d: %s\n", 
                   parseErr.Line, parseErr.Column, parseErr.Message)
    } else {
        // Error regular (no de parsing)
        fmt.Println("Error:", err.Error())
    }
}
```

## 🔧 Tipos de Errores Demostrados

### 1. Token No Reconocido

```
Input: "xyz abc"
Error: Token no reconocido 'xyz' en línea 1, columna 1
Posición: 0
```

### 2. Regla No Encontrada

```
Input: "HOLA mundo"
Error: No se encontró regla que coincida con los tokens: [HOLA WORLD] en línea 1, columna 1
Posición: 0
```

### 3. Input Vacío

```
Input: ""
Error: Input vacío en línea 1, columna 1
Posición: 0
```

### 4. Solo Espacios

```
Input: "   "
Error: Input vacío (solo espacios) en línea 1, columna 1
Posición: 0
```

## 🏗️ Implementación Técnica

### DSL de Demostración

```go
// Crear DSL simple para demostración
demo := dslbuilder.NewDSL("ErrorDemo")

// Tokens básicos
demo.KeywordToken("HELLO", "hello")
demo.Token("WORLD", "world")
demo.Token("NUMBER", "[0-9]+")

// Regla simple
demo.Rule("greeting", []string{"HELLO", "WORLD"}, "sayHello")

// Acción
demo.Action("sayHello", func(args []interface{}) (interface{}, error) {
    return "Hello World!", nil
})
```

### Casos de Prueba

```go
testCases := []struct {
    input       string
    description string
}{
    {"hello world", "✅ Comando válido"},
    {"xyz abc", "❌ Token no reconocido"},
    {"hello", "❌ Regla incompleta"},
    {"", "❌ Input vacío"},
    {"   ", "❌ Solo espacios"},
    {"hello 123", "❌ Token inesperado"},
}
```

## 📊 Ejemplo de Salida

```
=== go-dsl Error Demo ===
Demostración del sistema de errores mejorado con línea y columna

1. Comando válido: 'hello world'
   ✅ Resultado: Hello World!

2. Token no reconocido: 'xyz abc'
   ❌ Error en línea 1, columna 1: Token no reconocido 'xyz'
   Contexto: xyz abc
            ^

3. Comando incompleto: 'hello'
   ❌ Error en línea 1, columna 1: No se encontró regla que coincida con los tokens: [HELLO]
   Contexto: hello
            ^

4. Input vacío: ''
   ❌ Error en línea 1, columna 1: Input vacío
   Contexto: 
            ^

5. Solo espacios: '   '
   ❌ Error en línea 1, columna 1: Input vacío (solo espacios)
   Contexto:    
            ^

6. Token inesperado: 'hello 123'
   ❌ Error en línea 1, columna 7: No se encontró regla que coincida con los tokens: [HELLO NUMBER]
   Contexto: hello 123
                  ^

=== Comparación: Error estándar vs ParseError ===

Error estándar (compatible):
  Token no reconocido 'xyz'

ParseError (mejorado):
  Error en línea 1, columna 1: Token no reconocido 'xyz'
  Posición: 0
  Token: 'xyz'
  Input: 'xyz abc'
  Contexto visual:
    xyz abc
    ^
```

## 🔄 Migración Gradual

### Estrategia Recomendada

```go
// Paso 1: Mantener código existente funcionando
result, err := dsl.Parse(input)
if err != nil {
    // Código existente sigue funcionando
    log.Printf("Error: %s", err.Error())
    
    // Paso 2: Agregar información mejorada gradualmente
    if IsParseError(err) {
        parseErr := err.(*ParseError)
        log.Printf("Detalles: línea %d, columna %d", 
                   parseErr.Line, parseErr.Column)
    }
}
```

### Ventajas de la Migración

1. **Sin Breaking Changes**: Todo el código existente funciona
2. **Información Opcional**: Puedes usar características nuevas cuando necesites
3. **Mejor Debugging**: Errores más informativos para desarrollo
4. **Mejor UX**: Usuarios finales obtienen mejores mensajes de error

## 🎯 Casos de Uso Prácticos

### 1. **IDEs y Editores**
```go
if IsParseError(err) {
    parseErr := err.(*ParseError)
    // Resaltar error en línea y columna específicas
    highlightError(parseErr.Line, parseErr.Column)
}
```

### 2. **APIs Web**
```go
// Respuesta JSON con información detallada
if IsParseError(err) {
    parseErr := err.(*ParseError)
    return ErrorResponse{
        Message:  parseErr.Message,
        Line:     parseErr.Line,
        Column:   parseErr.Column,
        Context:  parseErr.Input,
    }
}
```

### 3. **Herramientas CLI**
```go
// Mostrar error con contexto visual
if IsParseError(err) {
    parseErr := err.(*ParseError)
    showVisualError(parseErr.Input, parseErr.Line, parseErr.Column)
}
```

### 4. **Sistemas de Logging**
```go
// Log estructurado con información completa
if IsParseError(err) {
    parseErr := err.(*ParseError)
    logger.WithFields(map[string]interface{}{
        "line":     parseErr.Line,
        "column":   parseErr.Column,
        "position": parseErr.Position,
        "token":    parseErr.Token,
    }).Error(parseErr.Message)
}
```

## 🔧 Características Técnicas

### 1. **Preservación de Tipos**
```go
// ParseError se preserva a través de la cadena de llamadas
func (dsl *DSL) Parse(input string) (*Result, error) {
    // ...
    if IsParseError(err) {
        return nil, err  // Preserva ParseError, no lo envuelve
    }
    return nil, fmt.Errorf("otro tipo de error: %w", err)
}
```

### 2. **Detección Inteligente**
```go
// Función helper para detección robusta
func IsParseError(err error) bool {
    if err == nil {
        return false
    }
    _, ok := err.(*ParseError)
    return ok
}
```

### 3. **Contexto Visual**
```go
// Función para mostrar contexto visual
func getContextLine(input string, position int) string {
    return input  // Input completo disponible
}
```

## 🎓 Lecciones Técnicas

### 1. **Compatibilidad es Clave**
Nuevas características no deben romper código existente.

### 2. **Información Gradual**
Usuarios pueden adoptar características nuevas a su ritmo.

### 3. **Tipos Explícitos**
`ParseError` es un tipo específico, no solo un `error` genérico.

### 4. **Helper Functions**
Funciones como `IsParseError()` simplifican el uso.

## 🔗 Casos Similares

- **Compiladores**: Errores de sintaxis con línea/columna
- **Linters**: Reporte de problemas con ubicación
- **IDEs**: Subrayado de errores en código
- **APIs**: Respuestas de error estructuradas
- **CLIs**: Mensajes de error informativos

## 🚀 Próximos Pasos

1. **Ejecuta el ejemplo**: `go run main.go`
2. **Modifica los inputs de prueba** en el código
3. **Experimenta con diferentes tipos de errores**
4. **Integra en tu propio código** gradualmente
5. **Compara con sistemas de error existentes**

## 📞 Referencias y Documentación

- **Código fuente**: [`main.go`](main.go)
- **Tests de ParseError**: [../../pkg/dslbuilder/parse_error_test.go](../../pkg/dslbuilder/parse_error_test.go)
- **Manual completo**: [Manual de Uso](../../docs/es/manual_de_uso.md)
- **Documentación técnica**: [Developer Onboarding](../../docs/es/developer_onboarding.md)

---

**¡Demuestra que go-dsl tiene errores informativos sin romper compatibilidad!** 🔧🎉