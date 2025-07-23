# Sistema de Consultas DSL - Query en Español

**Sistema de consultas tipo SQL en español que demuestra gramáticas complejas, filtros avanzados y procesamiento de datos empresariales.**

## 🎯 Objetivo

Este ejemplo demuestra cómo crear un **sistema de consultas empresarial** en español que incluye:

- 🔍 Sintaxis de consulta natural en español
- 📊 Filtros complejos con múltiples condiciones
- 🗂️ Procesamiento de estructuras de datos complejas
- 🎯 Operaciones de agregación (count, list, search)
- 🔄 Contexto dinámico para datasets variables
- 📝 Queries tipo LINQ/SQL en español

## 🚀 Ejecución Rápida

```bash
cd examples/query
go run main.go
```

## 📚 Características del DSL

### Tokens Definidos

```go
// Comandos principales
dsl.KeywordToken("LISTAR", "listar")         // list
dsl.KeywordToken("BUSCAR", "buscar")         // search
dsl.KeywordToken("CONTAR", "contar")         // count

// Entidades
dsl.KeywordToken("PRODUCTOS", "productos")   // products

// Filtros y condiciones  
dsl.KeywordToken("DONDE", "donde")           // where
dsl.KeywordToken("CATEGORIA", "categoria")   // category
dsl.KeywordToken("PRECIO", "precio")         // price
dsl.KeywordToken("STOCK", "stock")           // stock
dsl.KeywordToken("NOMBRE", "nombre")         // name

// Operadores
dsl.KeywordToken("ES", "es")                 // is
dsl.KeywordToken("MAYOR", "mayor")           // greater
dsl.KeywordToken("MENOR", "menor")           // less
dsl.KeywordToken("CONTIENE", "contiene")     // contains

// Valores
dsl.Token("STRING", "\"[^\"]*\"")            // strings
dsl.Token("NUMBER", "[0-9]+\\.?[0-9]*")      // numbers
dsl.Token("WORD", "[a-zA-Z]+")               // category names
```

### Comandos Soportados

#### 1. Consultas Básicas
```
listar productos                    # Lista todos los productos
contar productos                    # Cuenta total de productos
```

#### 2. Filtros por Categoría
```
buscar productos donde categoria es Electronics
listar productos donde categoria es Furniture
```

#### 3. Filtros por Precio
```
listar productos donde precio mayor 100
buscar productos donde precio menor 50
```

#### 4. Filtros por Stock
```
contar productos donde stock menor 10
listar productos donde stock mayor 20
```

#### 5. Búsqueda por Nombre
```
buscar productos donde nombre contiene "Desk"
listar productos donde nombre contiene "USB"
```

#### 6. Combinaciones Complejas
Las reglas están organizadas de más específica a menos específica para máxima flexibilidad.

## 🏗️ Arquitectura del Sistema

### Estructura de Datos

```go
type Product struct {
    Name     string
    Category string
    Price    float64
    Stock    int
}

// Dataset de ejemplo
products := []Product{
    {"Laptop Dell", "Electronics", 1200.00, 5},
    {"Mouse Logitech", "Electronics", 25.00, 50},
    {"Desk Chair", "Furniture", 350.00, 10},
    {"Standing Desk", "Furniture", 600.00, 3},
    {"USB Cable", "Electronics", 10.00, 100},
    {"Monitor 27\"", "Electronics", 400.00, 8},
    {"Office Lamp", "Furniture", 45.00, 20},
}
```

### Tipos de Filtros Implementados

```go
type Filter struct {
    Field     string      // "categoria", "precio", "stock", "nombre"
    Operator  string      // "es", "mayor", "menor", "contiene"
    Value     interface{} // Valor a comparar
}

// Función de filtrado universal
func applyFilter(products []Product, filter Filter) []Product {
    var result []Product
    
    for _, product := range products {
        matches := false
        
        switch filter.Field {
        case "categoria":
            matches = (filter.Operator == "es" && product.Category == filter.Value.(string))
            
        case "precio":
            switch filter.Operator {
            case "mayor":
                matches = product.Price > filter.Value.(float64)
            case "menor":
                matches = product.Price < filter.Value.(float64)
            }
            
        case "stock":
            switch filter.Operator {
            case "mayor":
                matches = float64(product.Stock) > filter.Value.(float64)
            case "menor":
                matches = float64(product.Stock) < filter.Value.(float64)
            }
            
        case "nombre":
            if filter.Operator == "contiene" {
                matches = strings.Contains(product.Name, filter.Value.(string))
            }
        }
        
        if matches {
            result = append(result, product)
        }
    }
    
    return result
}
```

## 🔧 Características Técnicas Avanzadas

### 1. Reglas Organizadas por Especificidad

```go
// MÁS específicas primero (patrones más largos)
query.Rule("query", []string{"BUSCAR", "PRODUCTOS", "DONDE", "CATEGORIA", "ES", "WORD"}, "filterByCategory")
query.Rule("query", []string{"LISTAR", "PRODUCTOS", "DONDE", "PRECIO", "MAYOR", "NUMBER"}, "filterByPriceGreater")
query.Rule("query", []string{"CONTAR", "PRODUCTOS", "DONDE", "STOCK", "MENOR", "NUMBER"}, "countByStockLess")
query.Rule("query", []string{"BUSCAR", "PRODUCTOS", "DONDE", "NOMBRE", "CONTIENE", "STRING"}, "filterByNameContains")

// Menos específicas después (patrones más cortos)
query.Rule("query", []string{"LISTAR", "PRODUCTOS"}, "listAll")
query.Rule("query", []string{"CONTAR", "PRODUCTOS"}, "countAll")
```

**¿Por qué este orden?**
- go-dsl intenta reglas en orden de definición
- Reglas más específicas capturan casos especiales
- Reglas generales capturan casos básicos
- Evita que reglas simples "capturen" comandos complejos

### 2. Contexto Dinámico para Datasets

```go
// El contexto puede cambiar por consulta
contextDemo := map[string]interface{}{
    "expensiveProducts": getExpensiveProducts(),
    "lowStockProducts":  getLowStockProducts(),
}

// Misma sintaxis, diferentes datos
result1, _ := query.Use("listar productos", context1)
result2, _ := query.Use("listar productos", context2)
```

### 3. Acciones Reutilizables

```go
// Acción genérica de filtrado
query.Action("applyFilter", func(args []interface{}) (interface{}, error) {
    // Construir filtro desde argumentos
    filter := Filter{
        Field:    extractField(args),
        Operator: extractOperator(args),
        Value:    extractValue(args),
    }
    
    // Obtener productos del contexto
    products := query.GetContext("products").([]Product)
    
    // Aplicar filtro
    return applyFilter(products, filter), nil
})
```

## 📊 Ejemplo de Salida

```
Query DSL Demo
==============

Query: listar productos
Resultado: 7 productos encontrados
  - Laptop Dell (Electronics) $1200.00 [Stock: 5]
  - Mouse Logitech (Electronics) $25.00 [Stock: 50]
  - Desk Chair (Furniture) $350.00 [Stock: 10]
  - Standing Desk (Furniture) $600.00 [Stock: 3]
  - USB Cable (Electronics) $10.00 [Stock: 100]
  - Monitor 27" (Electronics) $400.00 [Stock: 8]
  - Office Lamp (Furniture) $45.00 [Stock: 20]

Query: buscar productos donde categoria es Electronics
Resultado: 4 productos encontrados
  - Laptop Dell (Electronics) $1200.00 [Stock: 5]
  - Mouse Logitech (Electronics) $25.00 [Stock: 50]
  - USB Cable (Electronics) $10.00 [Stock: 100]
  - Monitor 27" (Electronics) $400.00 [Stock: 8]

Query: listar productos donde precio mayor 100
Resultado: 4 productos encontrados
  - Laptop Dell (Electronics) $1200.00 [Stock: 5]
  - Desk Chair (Furniture) $350.00 [Stock: 10]
  - Standing Desk (Furniture) $600.00 [Stock: 3]
  - Monitor 27" (Electronics) $400.00 [Stock: 8]

Query: buscar productos donde nombre contiene "Desk"
Resultado: 2 productos encontrados
  - Desk Chair (Furniture) $350.00 [Stock: 10]
  - Standing Desk (Furniture) $600.00 [Stock: 3]
```

## 🎯 Casos de Uso Empresariales

### 1. **Sistemas de Inventario**
```
listar productos donde stock menor 5          # Productos con bajo stock
contar productos donde categoria es Críticos   # Productos críticos
buscar productos donde precio mayor 1000      # Productos premium
```

### 2. **Análisis de Ventas**
```
listar ventas donde fecha mayor "2025-01-01"  # Ventas recientes
contar clientes donde ciudad es "Madrid"       # Clientes por ciudad
buscar pedidos donde estado es Pendiente       # Pedidos pendientes
```

### 3. **Gestión de Recursos Humanos**
```
listar empleados donde departamento es IT      # Empleados de IT
contar usuarios donde activo es true          # Usuarios activos
buscar candidatos donde experiencia mayor 5   # Candidatos senior
```

### 4. **Análisis Financiero**
```
listar gastos donde categoria es Marketing     # Gastos de marketing
contar facturas donde estado es Vencidas      # Facturas vencidas
buscar transacciones donde monto mayor 10000  # Transacciones grandes
```

## 🔧 Extensiones Posibles

### 1. **Operadores Adicionales**
```go
// Rangos
dsl.KeywordToken("ENTRE", "entre")
// "precio entre 100 y 500"

// Múltiples valores  
dsl.KeywordToken("EN", "en")
// "categoria en Electronics,Furniture"

// Fechas
dsl.KeywordToken("DESDE", "desde")
dsl.KeywordToken("HASTA", "hasta")
// "fecha desde 2025-01-01 hasta 2025-01-31"
```

### 2. **Agregaciones Avanzadas**
```go
// Estadísticas
query.Rule("query", []string{"PROMEDIO", "PRECIO", "PRODUCTOS"}, "averagePrice")
query.Rule("query", []string{"SUMA", "STOCK", "PRODUCTOS"}, "totalStock")
query.Rule("query", []string{"MAXIMO", "PRECIO", "PRODUCTOS"}, "maxPrice")
```

### 3. **Ordenamiento**
```go
// Orden
dsl.KeywordToken("ORDENAR", "ordenar")
dsl.KeywordToken("POR", "por")
dsl.KeywordToken("ASC", "asc")
dsl.KeywordToken("DESC", "desc")
// "listar productos ordenar por precio desc"
```

### 4. **Agrupación**
```go
// Grupos
dsl.KeywordToken("AGRUPAR", "agrupar")
// "contar productos agrupar por categoria"
```

## 🎓 Lecciones Técnicas

### 1. **KeywordToken Evita Conflictos**
Sin KeywordToken, palabras como "productos" podrían ser capturadas por patrones genéricos.

### 2. **Orden de Reglas es Crítico**
Reglas más específicas deben definirse primero para evitar que las generales las capturen.

### 3. **Contexto Permite Flexibilidad**
El mismo DSL puede trabajar con diferentes datasets según el contexto.

### 4. **Filtros Reutilizables**
Lógica de filtrado centralizada permite fácil extensión y mantenimiento.

## 🔗 Patrones Similares

- **Motores de Búsqueda**: Queries en lenguaje natural
- **Sistemas BI**: Consultas ad-hoc para análisis
- **APIs de Filtrado**: Queries complejas en APIs REST
- **Reporting**: Generación dinámica de reportes
- **Data Discovery**: Exploración de datasets grandes

## 🚀 Próximos Pasos

1. **Ejecuta el ejemplo**: `go run main.go`
2. **Modifica las consultas** en el código
3. **Agrega nuevos operadores** y filtros
4. **Experimenta con diferentes datasets**
5. **Combina con contexto dinámico** para casos reales

## 📞 Referencias y Documentación

- **Código fuente**: [`main.go`](main.go)
- **Contexto dinámico**: [Simple Context Example](../simple_context/)
- **Manual completo**: [Manual de Uso](../../docs/es/manual_de_uso.md)
- **Guía para contribuidores**: [Developer Onboarding](../../docs/es/developer_onboarding.md)

---

**¡Demuestra que go-dsl puede crear interfaces de consulta naturales en español para sistemas empresariales!** 🔍🎉