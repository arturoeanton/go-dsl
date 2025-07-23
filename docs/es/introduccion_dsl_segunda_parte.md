# Introducción a DSL - Segunda Parte: Conceptos Avanzados de Gramáticas

## Tabla de Contenidos
1. [Tipos de Gramáticas](#tipos-de-gramáticas)
2. [Jerarquía de Chomsky](#jerarquía-de-chomsky)
3. [Gramáticas Libres de Contexto](#gramáticas-libres-de-contexto)
4. [Recursión en Gramáticas](#recursión-en-gramáticas)
5. [Precedencia y Asociatividad](#precedencia-y-asociatividad)
6. [Operadores de Repetición](#operadores-de-repetición)
7. [Análisis Sintáctico](#análisis-sintáctico)
8. [Aplicación en go-dsl](#aplicación-en-go-dsl)

## Tipos de Gramáticas

Una gramática formal es un conjunto de reglas que describe la estructura de un lenguaje. En teoría de lenguajes formales, las gramáticas se clasifican según su poder expresivo y las restricciones en sus reglas de producción.

### Componentes de una Gramática

Toda gramática formal G se define como una tupla (V, Σ, R, S) donde:
- **V**: Conjunto de símbolos no terminales (variables)
- **Σ**: Conjunto de símbolos terminales (alfabeto)
- **R**: Conjunto de reglas de producción
- **S**: Símbolo inicial (S ∈ V)

## Jerarquía de Chomsky

Noam Chomsky clasificó las gramáticas en cuatro tipos, ordenados por su poder expresivo:

### Tipo 0: Gramáticas sin Restricciones
- **Forma**: α → β (donde α contiene al menos un no terminal)
- **Lenguaje**: Recursivamente enumerable
- **Autómata**: Máquina de Turing

### Tipo 1: Gramáticas Sensibles al Contexto
- **Forma**: αAβ → αγβ (donde A es no terminal, α,β,γ son cadenas)
- **Restricción**: |α| ≤ |β| (no se puede reducir la longitud)
- **Lenguaje**: Sensible al contexto
- **Autómata**: Autómata linealmente acotado

### Tipo 2: Gramáticas Libres de Contexto (CFG)
- **Forma**: A → γ (donde A es no terminal, γ es cualquier cadena)
- **Lenguaje**: Libre de contexto
- **Autómata**: Autómata de pila
- **Ejemplo en go-dsl**:
```go
// Una regla CFG típica
dsl.Rule("expr", []string{"expr", "PLUS", "term"}, "add")
// expr → expr PLUS term
```

### Tipo 3: Gramáticas Regulares
- **Forma**: A → aB o A → a (gramática lineal derecha)
- **Lenguaje**: Regular
- **Autómata**: Autómata finito
- **Ejemplo en go-dsl**:
```go
// Los tokens son expresiones regulares
dsl.Token("NUMBER", "[0-9]+")
dsl.Token("ID", "[a-zA-Z][a-zA-Z0-9]*")
```

## Gramáticas Libres de Contexto

go-dsl se enfoca en gramáticas libres de contexto (CFG), que son ideales para lenguajes de programación y DSLs porque:

1. **Suficiente poder expresivo**: Pueden describir estructuras anidadas y recursivas
2. **Análisis eficiente**: Existen algoritmos de parsing en tiempo polinomial
3. **Balance entre expresividad y complejidad**: No tan restrictivas como las regulares, no tan complejas como las sensibles al contexto

### Ejemplo de CFG para Expresiones Aritméticas

```
E → E + T | E - T | T
T → T * F | T / F | F
F → ( E ) | número
```

En go-dsl:
```go
calc.Rule("E", []string{"E", "PLUS", "T"}, "add")
calc.Rule("E", []string{"E", "MINUS", "T"}, "subtract")
calc.Rule("E", []string{"T"}, "passthrough")

calc.Rule("T", []string{"T", "MULTIPLY", "F"}, "multiply")
calc.Rule("T", []string{"T", "DIVIDE", "F"}, "divide")
calc.Rule("T", []string{"F"}, "passthrough")

calc.Rule("F", []string{"LPAREN", "E", "RPAREN"}, "paren")
calc.Rule("F", []string{"NUMBER"}, "number")
```

## Recursión en Gramáticas

### Recursión Izquierda
Una gramática tiene recursión izquierda cuando existe una regla A → Aα:

```
A → Aα | β
```

**Problema**: Muchos parsers tradicionales entran en bucle infinito.

**Solución en go-dsl**: Parser mejorado con memoización y análisis iterativo.

```go
// Recursión izquierda directa
list.Rule("items", []string{"items", "item"}, "append")
list.Rule("items", []string{"item"}, "single")
```

### Recursión Derecha
Una gramática tiene recursión derecha cuando existe una regla A → αA:

```
A → αA | β
```

**Ventaja**: Más fácil de parsear con parsers descendentes recursivos.

```go
// Recursión derecha
list.Rule("items", []string{"item", "items"}, "prepend")
list.Rule("items", []string{"item"}, "single")
```

### Recursión Mutua
Dos o más reglas se llaman mutuamente:

```
A → B α
B → A β | γ
```

## Precedencia y Asociatividad

### Precedencia de Operadores
Define qué operador se evalúa primero cuando hay ambigüedad:

```
2 + 3 * 4 = ?
```

Sin precedencia: podría ser (2 + 3) * 4 = 20 o 2 + (3 * 4) = 14

Con precedencia (multiplicación > suma): 2 + (3 * 4) = 14

### Implementación en go-dsl

```go
// Menor precedencia = menor número
calc.RuleWithPrecedence("expr", []string{"expr", "PLUS", "term"}, "add", 1, "left")
calc.RuleWithPrecedence("term", []string{"term", "MULTIPLY", "factor"}, "multiply", 2, "left")
calc.RuleWithPrecedence("factor", []string{"base", "POWER", "factor"}, "power", 3, "right")
```

### Asociatividad
Define cómo se agrupan operadores del mismo nivel:

- **Asociatividad izquierda**: a - b - c = (a - b) - c
- **Asociatividad derecha**: a ^ b ^ c = a ^ (b ^ c)
- **No asociativo**: a < b < c = error

## Operadores de Repetición

### Kleene Star (*)
Cero o más repeticiones de un elemento:

```
A* = ε | A | AA | AAA | ...
```

En notación EBNF:
```
lista ::= elemento*
```

En go-dsl:
```go
dsl.RuleWithRepetition("items", "item", "items")
// Genera automáticamente:
// items → ε
// items → items item
```

### Kleene Plus (+)
Una o más repeticiones de un elemento:

```
A+ = A | AA | AAA | ...
```

En notación EBNF:
```
lista ::= elemento+
```

En go-dsl:
```go
dsl.RuleWithPlusRepetition("items", "item", "items")
// Genera automáticamente:
// items → item
// items → items item
```

### Opcional (?)
Cero o una ocurrencia:

```
A? = ε | A
```

## Análisis Sintáctico

### Tipos de Parsers

#### Top-Down (Descendente)
- Comienza desde el símbolo inicial
- Intenta derivar la entrada
- Ejemplos: LL(k), Recursive Descent

#### Bottom-Up (Ascendente)
- Comienza desde la entrada
- Intenta reducir al símbolo inicial
- Ejemplos: LR(k), LALR, SLR

### Parser de go-dsl
go-dsl utiliza un parser descendente recursivo mejorado con:

1. **Memoización**: Evita recalcular subárboles (Packrat parsing)
2. **Manejo de recursión izquierda**: Análisis iterativo especial
3. **Backtracking**: Prueba alternativas cuando falla

```go
// Internamente, el parser maneja la recursión izquierda
func (p *ImprovedParser) parseLeftRecursive(ruleName string) (interface{}, error) {
    // 1. Encuentra casos base (no recursivos)
    // 2. Aplica iterativamente las reglas recursivas
    // 3. Construye el AST de izquierda a derecha
}
```

## Aplicación en go-dsl

### Ejemplo Completo: Calculadora con Precedencia

```go
package main

import (
    "github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

func main() {
    calc := dslbuilder.New("AdvancedCalc")
    
    // Tokens (expresiones regulares - Tipo 3)
    calc.Token("NUMBER", "[0-9]+")
    calc.Token("PLUS", "\\+")
    calc.Token("MINUS", "-")
    calc.Token("MULTIPLY", "\\*")
    calc.Token("DIVIDE", "/")
    calc.Token("POWER", "\\^")
    
    // Reglas CFG con precedencia (Tipo 2)
    // Nivel 1: Suma/Resta (menor precedencia)
    calc.RuleWithPrecedence("expr", []string{"expr", "PLUS", "term"}, "add", 1, "left")
    calc.RuleWithPrecedence("expr", []string{"expr", "MINUS", "term"}, "subtract", 1, "left")
    calc.Rule("expr", []string{"term"}, "passthrough")
    
    // Nivel 2: Multiplicación/División
    calc.RuleWithPrecedence("term", []string{"term", "MULTIPLY", "factor"}, "multiply", 2, "left")
    calc.RuleWithPrecedence("term", []string{"term", "DIVIDE", "factor"}, "divide", 2, "left")
    calc.Rule("term", []string{"factor"}, "passthrough")
    
    // Nivel 3: Potenciación (mayor precedencia, asociativa derecha)
    calc.RuleWithPrecedence("factor", []string{"primary", "POWER", "factor"}, "power", 3, "right")
    calc.Rule("factor", []string{"primary"}, "passthrough")
    
    // Expresiones primarias
    calc.Rule("primary", []string{"NUMBER"}, "number")
    calc.Rule("primary", []string{"LPAREN", "expr", "RPAREN"}, "paren")
}
```

### Ventajas del Enfoque de go-dsl

1. **Declarativo**: Define qué, no cómo
2. **Flexible**: Soporta múltiples estilos de gramática
3. **Potente**: Maneja recursión izquierda automáticamente
4. **Eficiente**: Memoización evita recálculos
5. **Práctico**: API intuitiva para casos comunes

## Conclusión

go-dsl implementa un sistema de parsing basado en gramáticas libres de contexto con extensiones prácticas:

- **Precedencia y asociatividad** configurables
- **Operadores de repetición** (Kleene star/plus)
- **Manejo automático de recursión izquierda**
- **Prioridad de tokens** para resolver ambigüedades

Estas características permiten crear DSLs sofisticados manteniendo una API simple y declarativa, haciendo que conceptos teóricos complejos sean accesibles para desarrolladores prácticos.

## Referencias

- Aho, A. V., Lam, M. S., Sethi, R., & Ullman, J. D. (2006). *Compilers: Principles, Techniques, and Tools* (2nd ed.)
- Chomsky, N. (1956). "Three models for the description of language"
- Ford, B. (2004). "Parsing expression grammars: a recognition-based syntactic foundation"