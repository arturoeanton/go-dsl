# Introducción a los Lenguajes Específicos de Dominio (DSL)

## 📚 Tabla de Contenidos

1. [¿Qué es un DSL?](#qué-es-un-dsl)
2. [¿Por qué usar DSLs?](#por-qué-usar-dsls)
3. [Tipos de DSLs](#tipos-de-dsls)
4. [DSLs vs Lenguajes de Propósito General](#dsls-vs-lenguajes-de-propósito-general)
5. [Ejemplos del Mundo Real](#ejemplos-del-mundo-real)
6. [Anatomía de un DSL](#anatomía-de-un-dsl)
7. [Cuándo Crear un DSL](#cuándo-crear-un-dsl)
8. [Ventajas y Desventajas](#ventajas-y-desventajas)
9. [DSLs en la Práctica](#dsls-en-la-práctica)
10. [Cómo go-dsl Facilita la Creación de DSLs](#cómo-go-dsl-facilita-la-creación-de-dsls)

## ¿Qué es un DSL?

Un **Lenguaje Específico de Dominio** (Domain Specific Language o DSL) es un lenguaje de programación o especificación diseñado para resolver problemas en un dominio particular. A diferencia de los lenguajes de propósito general como Go, Python o Java, los DSLs están optimizados para expresar soluciones en un contexto específico.

### Definición Simple

> Un DSL es un mini-lenguaje creado para hacer una tarea específica de manera más fácil y expresiva.

### Ejemplo Cotidiano

Piensa en las expresiones regulares (regex). Son un DSL para buscar patrones en texto:

```regex
^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$
```

Esta línea valida un email. Imagina escribir la misma lógica en Go puro - necesitarías decenas de líneas de código.

## ¿Por qué usar DSLs?

### 1. **Expresividad**
Los DSLs permiten expresar conceptos complejos de forma simple y natural para el dominio.

**Sin DSL (Go puro):**
```go
transaction := Transaction{
    Type: "sale",
    Amount: 5000,
    Tax: 5000 * 0.16,
    Total: 5000 * 1.16,
}
```

**Con DSL:**
```
venta de 5000 con iva
```

### 2. **Reducción de Errores**
Al limitar las opciones a lo relevante del dominio, se reducen los errores posibles.

### 3. **Accesibilidad**
Personas no programadoras pueden entender y modificar reglas de negocio.

### 4. **Mantenibilidad**
Los cambios en las reglas de negocio no requieren modificar el código fuente.

## Tipos de DSLs

### DSLs Internos (Embedded DSLs)

Se construyen dentro de un lenguaje anfitrión existente, aprovechando su sintaxis.

**Ejemplo en Go:**
```go
// DSL interno para construir consultas
query := Select("name", "age").
         From("users").
         Where("age > ?", 18).
         OrderBy("name")
```

### DSLs Externos

Tienen su propia sintaxis independiente y requieren un parser.

**Ejemplo:**
```sql
SELECT name, age 
FROM users 
WHERE age > 18 
ORDER BY name
```

### DSLs Visuales

Usan elementos gráficos en lugar de texto.

**Ejemplo:** Scratch, LabVIEW, diagramas de flujo

## DSLs vs Lenguajes de Propósito General

| Aspecto | DSL | Lenguaje General |
|---------|-----|------------------|
| **Alcance** | Dominio específico | Cualquier problema |
| **Curva de aprendizaje** | Baja (en su dominio) | Alta |
| **Expresividad** | Alta (en su dominio) | Media |
| **Flexibilidad** | Limitada | Total |
| **Usuarios** | Expertos del dominio | Programadores |

## Ejemplos del Mundo Real

### 1. SQL - Consultas de Base de Datos
```sql
SELECT cliente, SUM(monto) as total
FROM ventas
WHERE fecha >= '2024-01-01'
GROUP BY cliente
HAVING total > 10000
```

### 2. HTML - Estructura de Páginas Web
```html
<article>
  <h1>Título del Artículo</h1>
  <p>Contenido del artículo...</p>
</article>
```

### 3. CSS - Estilos Visuales
```css
.boton-principal {
  background-color: #007bff;
  padding: 10px 20px;
  border-radius: 5px;
}
```

### 4. Make - Automatización de Compilación
```makefile
programa: main.o utils.o
    gcc -o programa main.o utils.o

main.o: main.c
    gcc -c main.c
```

### 5. Gherkin - Pruebas BDD
```gherkin
Escenario: Compra exitosa
  Dado que tengo 100 euros en mi cuenta
  Cuando compro un producto de 50 euros
  Entonces mi saldo debe ser 50 euros
```

### 6. Docker - Configuración de Contenedores
```dockerfile
FROM golang:1.19
WORKDIR /app
COPY . .
RUN go build -o main .
CMD ["./main"]
```

## Anatomía de un DSL

Todo DSL tiene componentes fundamentales:

### 1. **Léxico (Tokens)**
Las palabras o símbolos básicos del lenguaje.

```
Tokens: VENTA, DE, CON, IVA, números, identificadores
```

### 2. **Sintaxis (Gramática)**
Las reglas de cómo se combinan los tokens.

```
transacción := VENTA DE cantidad CON IVA
```

### 3. **Semántica (Significado)**
Qué hace cada construcción del lenguaje.

```
"venta de 1000 con iva" → Crear transacción con monto=1000, iva=160, total=1160
```

### 4. **Runtime/Intérprete**
El motor que ejecuta el DSL.

## Cuándo Crear un DSL

### ✅ Crear un DSL cuando:

1. **Dominio complejo y repetitivo**
   - Muchas reglas de negocio similares
   - Patrones que se repiten constantemente

2. **Usuarios no técnicos**
   - Contadores que necesitan definir reglas fiscales
   - Analistas que crean consultas de datos

3. **Cambios frecuentes**
   - Reglas que cambian con regulaciones
   - Configuraciones que varían por cliente

4. **Expresividad importante**
   - El código actual es difícil de leer
   - La lógica del dominio se pierde en detalles técnicos

### ❌ NO crear un DSL cuando:

1. **Problema simple**
   - Una función o biblioteca es suficiente

2. **Uso único**
   - No hay reutilización prevista

3. **Sin expertos del dominio**
   - Nadie puede definir el lenguaje adecuado

4. **Recursos limitados**
   - Crear y mantener un DSL requiere inversión

## Ventajas y Desventajas

### ✅ Ventajas

1. **Productividad**
   - Desarrollo más rápido en el dominio
   - Menos código para mantener

2. **Calidad**
   - Menos errores por abstracción adecuada
   - Validación en tiempo de parsing

3. **Comunicación**
   - Lenguaje común entre técnicos y negocio
   - Documentación ejecutable

4. **Evolución**
   - Cambios centralizados
   - Fácil agregar nuevas características

### ❌ Desventajas

1. **Costo inicial**
   - Tiempo de diseño e implementación
   - Curva de aprendizaje inicial

2. **Mantenimiento**
   - Otro lenguaje que mantener
   - Documentación adicional

3. **Limitaciones**
   - No puede hacer todo
   - Puede necesitar "escape hatches"

4. **Debugging**
   - Herramientas de debug limitadas
   - Stack traces menos claros

## DSLs en la Práctica

### Proceso de Diseño

1. **Análisis del Dominio**
   ```
   - ¿Qué problemas resuelve?
   - ¿Quiénes son los usuarios?
   - ¿Qué operaciones son comunes?
   ```

2. **Diseño del Lenguaje**
   ```
   - Vocabulario natural del dominio
   - Sintaxis intuitiva
   - Casos de uso principales
   ```

3. **Prototipo Rápido**
   ```
   - Parser básico
   - Casos de prueba
   - Feedback de usuarios
   ```

4. **Implementación**
   ```
   - Parser robusto
   - Manejo de errores
   - Herramientas de soporte
   ```

### Ejemplo: DSL Contable

**Análisis:**
- Usuarios: Contadores
- Dominio: Transacciones y asientos contables
- Operaciones: Ventas, compras, asientos

**Diseño:**
```
venta de 5000 con iva
compra de materiales por 3000 sin iva
asiento debe 1101 1000 haber 2101 1000
```

**Beneficios:**
- Natural para contadores
- Validación automática de partida doble
- Cálculos de impuestos integrados

## Cómo go-dsl Facilita la Creación de DSLs

go-dsl elimina la complejidad de crear DSLs proporcionando:

### 1. **Parser Automático**
No necesitas escribir un parser desde cero.

```go
dsl := dslbuilder.New("MiDSL")
// go-dsl genera el parser automáticamente
```

### 2. **Sistema de Tokens Flexible**
Define tu vocabulario fácilmente.

```go
dsl.KeywordToken("VENTA", "venta")
dsl.Token("NUMERO", "[0-9]+")
```

### 3. **Gramáticas Declarativas**
Especifica reglas de forma simple.

```go
dsl.Rule("transaccion", []string{"VENTA", "DE", "NUMERO"}, "procesarVenta")
```

### 4. **Acciones Programables**
Conecta tu DSL con lógica Go.

```go
dsl.Action("procesarVenta", func(args []interface{}) (interface{}, error) {
    monto := args[2].(string)
    // Tu lógica aquí
})
```

### 5. **Soporte YAML/JSON**
Define DSLs sin escribir código.

```yaml
name: "Contabilidad"
tokens:
  VENTA: "venta"
  NUMERO: "[0-9]+"
rules:
  - name: "transaccion"
    pattern: ["VENTA", "DE", "NUMERO"]
    action: "procesarVenta"
```

## Conclusión

Los DSLs son herramientas poderosas para resolver problemas específicos de forma elegante. Con go-dsl, crear tu propio DSL es accesible y práctico, permitiéndote:

- Expresar soluciones en el lenguaje del dominio
- Empoderar a usuarios no técnicos
- Mantener reglas de negocio separadas del código
- Evolucionar rápidamente con los requisitos

### Próximos Pasos

1. Lee la [Guía Rápida](guia_rapida.md) para empezar con go-dsl
2. Explora los [ejemplos](../../examples/) para ver DSLs reales
3. Crea tu primer DSL siguiendo los patrones aprendidos

---

*"El mejor código es el que no tienes que escribir. El segundo mejor es el que se lee como prosa en el lenguaje del dominio."*