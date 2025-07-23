# DSL Universal de Reglas de Negocio (Inspirado en Drools)

Un motor de reglas de negocio universal y genérico inspirado en Drools, que funciona con cualquier tipo de estructura usando reflexión de Go.

## 🎯 Características

- **100% Universal**: Funciona con cualquier estructura Go usando reflexión
- **Sintaxis tipo Drools**: Sintaxis familiar de reglas de negocio
- **Bilingüe**: Soporta reglas en Español e Inglés
- **Salience/Prioridad**: Control de orden de ejecución de reglas
- **Contexto Dinámico**: Soporte para contextos dinámicos
- **Forward-chaining**: Ejecución de reglas en cadena
- **Working Memory**: Patrón de memoria de trabajo como Drools
- **Tags de Estructura**: Soporte para tags `drools:"fieldname"`

## 🚀 Uso

```bash
cd examples/drools
go run main.go
```

## 📖 Sintaxis de Reglas

### Estructura Básica (Español)
```
rule "Nombre de la Regla" 
salience PRIORIDAD 
when ENTIDAD CAMPO OPERADOR VALOR 
then ACCION CAMPO a VALOR 
end
```

### Estructura Básica (English)
```
rule "Rule Name" 
salience PRIORITY 
when ENTITY FIELD OPERATOR VALUE 
then ACTION FIELD to VALUE 
end
```

## 🔧 Ejemplos de Reglas

### Reglas en Español
```go
// Categorización de clientes
rule "Categorizar Cliente Premium" salience 100 
when customer categoria es "regular" and customer balance mayor 10000 
then establecer categoria a "premium" 
end

// Órdenes de alta prioridad
rule "Orden Alta Prioridad" salience 70 
when order monto mayor 3000 
then establecer prioridad a "high" 
end

// Control de inventario
rule "Stock Bajo" salience 50 
when product stock menor 10 
then establecer estado a "low_stock" 
end
```

### Reglas en English
```go
// Customer categorization
rule "Premium Customer Discount" salience 30 
when customer categoria is "premium" 
then set estado to "discount_eligible" 
end

// VIP treatment
rule "VIP Customer Special" salience 20 
when customer categoria is "vip" 
then set estado to "vip_treatment" 
end
```

## ⚙️ Operadores Soportados

### Español
- `es` - igualdad
- `mayor` - mayor que
- `menor` - menor que
- `contiene` - contiene texto
- `coincide` - coincide patrón

### English
- `is` / `equals` - equality
- `greater` - greater than
- `less` - less than
- `contains` - contains text
- `matches` - matches pattern

## 🎭 Acciones Soportadas

### Español
- `establecer` - establecer campo
- `modificar` - modificar entidad
- `insertar` - insertar nueva entidad
- `retirar` - remover entidad
- `ejecutar` - ejecutar función

### English
- `set` - set field
- `modify` - modify entity
- `insert` - insert new entity
- `retract` - retract entity
- `execute` - execute function

## 📊 Soporte de Tipos

El motor funciona con **cualquier estructura Go**:

```go
type Customer struct {
    ID       int     `drools:"id"`
    Name     string  `drools:"nombre"`
    Category string  `drools:"categoria"`
    Balance  float64 `drools:"balance"`
    Status   string  `drools:"estado"`
}

type Product struct {
    ID       int     `drools:"id"`
    Name     string  `drools:"nombre"`
    Price    float64 `drools:"precio"`
    Stock    int     `drools:"stock"`
    Status   string  `drools:"estado"`
}

type Order struct {
    ID         int     `drools:"id"`
    CustomerID int     `drools:"customer_id"`
    Amount     float64 `drools:"monto"`
    Status     string  `drools:"estado"`
    Priority   string  `drools:"prioridad"`
}
```

## 🔄 Ciclo de Ejecución

1. **Insertar Facts**: Agregar entidades a la memoria de trabajo
2. **Definir Reglas**: Crear reglas de negocio con salience
3. **Evaluar Condiciones**: El motor evalúa todas las condiciones
4. **Ejecutar Acciones**: Se ejecutan las acciones de reglas que coinciden
5. **Forward-chaining**: Las modificaciones pueden disparar nuevas reglas

## 🌟 Salience (Prioridad)

```go
// Reglas ejecutadas en orden de prioridad (mayor salience primero)
rule "Critical Rule" salience 100 when ... then ... end     // Ejecuta primero
rule "Important Rule" salience 50 when ... then ... end     // Ejecuta segundo  
rule "Normal Rule" salience 10 when ... then ... end        // Ejecuta tercero
```

## 🔧 API del Motor

### Crear Motor
```go
drools := universal.NewUniversalDroolsDSL()
engine := drools.GetEngine()
```

### Insertar Facts
```go
customer := &Customer{1, "Juan García", "regular", 5000.0, "active"}
drools.InsertFact(customer)
```

### Definir Reglas
```go
rule := `rule "Mi Regla" when customer balance mayor 1000 then establecer categoria a "vip" end`
result, err := drools.Parse(rule)
```

### Ejecutar Reglas
```go
err := drools.FireAllRules()
```

### Contexto Dinámico
```go
context := map[string]interface{}{
    "season": "holiday",
    "promotion": "black_friday",
}
result, err := drools.Use(rule, context)
```

## 🎯 Casos de Uso

- **E-commerce**: Categorización de clientes, descuentos dinámicos
- **Finanzas**: Aprobación de créditos, detección de fraude  
- **Inventario**: Control de stock, reordenamiento automático
- **Marketing**: Segmentación de clientes, campañas personalizadas
- **Logística**: Priorización de órdenes, optimización de rutas

## 🏗️ Arquitectura

- **UniversalRuleEngine**: Motor principal con reflexión
- **UniversalDroolsDSL**: Parser de sintaxis Drools
- **Rule**: Estructura de regla con condiciones y acciones
- **Working Memory**: Almacenamiento de facts
- **Forward-chaining**: Ejecución reactiva de reglas

## ✅ Características Empresariales

- ✅ **100% Genérico** - Funciona con cualquier estructura
- ✅ **Alto Rendimiento** - Optimizado con reflexión eficiente  
- ✅ **Sintaxis Natural** - Similar a lenguaje natural
- ✅ **Multiidioma** - Español e Inglés
- ✅ **Extensible** - Fácil agregar nuevos operadores/acciones
- ✅ **Production Ready** - Listo para producción
- ✅ **Zero Dependencies** - Solo Go estándar + go-dsl

¡Un motor de reglas de negocio completo y universal para Go! 🚀