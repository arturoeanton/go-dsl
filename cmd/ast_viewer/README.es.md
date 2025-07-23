# Visualizador de AST

Herramienta para visualizar el Árbol de Sintaxis Abstracta (AST) de los resultados de parsing de tu DSL.

## Descripción General

El Visualizador de AST te ayuda a entender cómo tu DSL parsea la entrada mostrando la estructura resultante en varios formatos. Esto es esencial para depurar reglas gramaticales y entender el proceso de parsing.

## Instalación

```bash
go install github.com/arturoeanton/go-dsl/cmd/ast_viewer@latest
```

O compilar desde el código fuente:

```bash
cd cmd/ast_viewer
go build -o ast_viewer
```

## Uso

```bash
ast_viewer -dsl <archivo-dsl> -input <cadena-entrada> [opciones]
```

### Opciones

- `-dsl` - Archivo de configuración DSL (YAML o JSON) **[requerido]**
- `-input` - Cadena de entrada para parsear
- `-file` - Archivo de entrada para parsear (alternativa a -input)
- `-format` - Formato de salida: `json`, `yaml`, o `tree` (por defecto: json)
- `-indent` - Indentar salida para json/yaml (por defecto: true)
- `-verbose` - Mostrar información detallada de tokens y reglas

### Ejemplos

**Uso básico con salida JSON:**
```bash
ast_viewer -dsl calculadora.yaml -input "10 + 20"
```

**Visualización en formato árbol:**
```bash
ast_viewer -dsl calculadora.yaml -input "10 + 20 * 30" -format tree
```

**Parsear desde archivo con salida YAML:**
```bash
ast_viewer -dsl consultas.json -file consultas.txt -format yaml
```

**Modo detallado para depuración:**
```bash
ast_viewer -dsl contabilidad.yaml -input "venta de 1000 con iva" -format tree -verbose
```

## Formatos de Salida

### Formato JSON
```json
{
  "type": "root",
  "value": "30",
  "children": [
    {
      "type": "arg_0",
      "value": "10"
    },
    {
      "type": "arg_1",
      "value": "+"
    },
    {
      "type": "arg_2",
      "value": "20"
    }
  ]
}
```

### Formato Árbol
```
root: 30
  ├─ arg_0: 10
  ├─ arg_1: +
  └─ arg_2: 20
```

### Formato YAML
```yaml
type: root
value: "30"
children:
  - type: arg_0
    value: "10"
  - type: arg_1
    value: "+"
  - type: arg_2
    value: "20"
```

## Casos de Uso

1. **Depuración de Gramática**: Entender cómo se emparejan tus reglas
2. **Análisis de Tokens**: Ver qué tokens están siendo reconocidos
3. **Optimización de Reglas**: Identificar patrones de parsing ineficientes
4. **Documentación**: Generar representaciones visuales de resultados de parsing

## Limitaciones

- La implementación actual muestra una estructura AST simplificada
- La información completa del árbol de parsing requiere integración más profunda con el parser
- Las acciones se ejecutan pero su lógica interna no se muestra

## Mejoras Futuras

- [ ] Árbol de parsing completo con todos los nodos intermedios
- [ ] Seguimiento de posición de tokens (línea, columna)
- [ ] Visualización del emparejamiento de reglas
- [ ] Visor interactivo basado en web
- [ ] Exportar a formato GraphViz