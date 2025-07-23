# Ejemplo de DSL Declarativo

Este ejemplo demuestra las nuevas características declarativas agregadas a go-dsl:
- **Builder Pattern** para API fluida
- **Configuración YAML/JSON** para definición de DSL
- **100% compatible** con código existente

## Características Demostradas

### 1. Builder Pattern (API Fluida)

```go
dsl := dslbuilder.New("Calculator").
    WithToken("NUMBER", "[0-9]+").
    WithToken("PLUS", "\\+").
    WithRule("expr", []string{"NUMBER", "PLUS", "NUMBER"}, "add").
    WithAction("add", func(args []interface{}) (interface{}, error) {
        // Implementación
    })
```

### 2. Configuración YAML

```yaml
name: "Calculator"
tokens:
  NUMBER: "[0-9]+"
  PLUS: "+"
  MINUS: "-"
rules:
  - name: "expr"
    pattern: ["NUMBER", "PLUS", "NUMBER"]
    action: "add"
```

### 3. Exportar/Importar JSON

```go
// Guardar DSL a JSON
jsonData, _ := dsl.SaveToJSON()
dsl.SaveToJSONFile("calculator.json")

// Cargar DSL desde JSON
loadedDSL, _ := dslbuilder.LoadFromJSONFile("calculator.json")
```

## Ejecutar el Ejemplo

```bash
cd examples/declarative
go run main.go
```

## Salida

El ejemplo:
1. Carga un DSL de calculadora desde `calculator.yaml`
2. Crea el mismo DSL usando el Builder Pattern
3. Exporta el DSL a formato JSON
4. Prueba compatibilidad con API tradicional

## Archivos

- `main.go` - Implementación del ejemplo
- `calculator.yaml` - Definición DSL en YAML
- `calculator.json` - Configuración JSON generada (después de ejecutar)

## Compatibilidad Hacia Atrás

Todo el código existente sigue funcionando:

```go
// API tradicional aún funciona
dsl := dslbuilder.New("Calculator")
dsl.Token("NUMBER", "[0-9]+")
dsl.Rule("expr", []string{"NUMBER", "PLUS", "NUMBER"}, "add")
dsl.Action("add", addFunction)
```