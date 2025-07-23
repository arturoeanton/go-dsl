# Documentación en Español - go-dsl

Bienvenido a la documentación completa de go-dsl en español. Aquí encontrarás toda la información necesaria para crear tus propios lenguajes específicos de dominio (DSL) con Go.

## 📚 Índice de Documentación

### 🚀 Comenzar
- **[Guía Rápida](guia_rapida.md)** - Introducción completa, instalación y ejemplos prácticos
  - Instalación y configuración
  - Conceptos básicos: tokens, reglas y acciones
  - Características avanzadas: precedencia, asociatividad y repetición
  - Ejemplos empresariales completos

### 🎓 Conceptos Avanzados
- **[Introducción a DSL - Segunda Parte](introduccion_dsl_segunda_parte.md)** - Teoría de lenguajes formales
  - Tipos de gramáticas y jerarquía de Chomsky
  - Gramáticas libres de contexto (CFG)
  - Recursión izquierda y derecha
  - Precedencia y asociatividad de operadores
  - Operadores de repetición (Kleene star/plus)
  - Análisis sintáctico y tipos de parsers

### ⚠️ Consideraciones Técnicas
- **[Limitaciones](limitaciones.md)** - Limitaciones conocidas y soluciones
  - Limitaciones del lenguaje Go (regex, tipos)
  - Limitaciones de diseño y performance
  - Comparación con otras herramientas (ANTLR, Yacc, PEG)
  - Soluciones y alternativas recomendadas

### 📈 Desarrollo y Mejoras
- **[Propuesta de Mejoras](propuesta_de_mejoras.md)** - Roadmap y estado actual
  - Mejoras implementadas ✅
  - Mejoras en desarrollo 🚧
  - Propuestas futuras 📋

## 🔧 Herramientas de Desarrollo

### Visualizador de AST
```bash
ast_viewer -dsl grammar.yaml -input "expresión" -format tree
```
- Visualiza el árbol sintáctico abstracto
- Soporta múltiples formatos de salida
- Ideal para debugging

### Validador de Gramática
```bash
validator -dsl grammar.yaml -strict -verbose
```
- Detecta conflictos de tokens
- Analiza dependencias circulares
- Valida la estructura de la gramática

### REPL Interactivo
```bash
repl -dsl grammar.yaml -context data.json
```
- Prueba expresiones interactivamente
- Historial de comandos con flechas
- Comandos especiales (.help, .tokens, .rules)

## 💡 Guía de Aprendizaje

### Para Principiantes
1. Lee la **[Guía Rápida](guia_rapida.md)** completa
2. Prueba los ejemplos en `/examples/simple/`
3. Experimenta con el REPL interactivo
4. Crea tu primer DSL siguiendo los ejemplos

### Para Usuarios Intermedios
1. Estudia **[Conceptos Avanzados](introduccion_dsl_segunda_parte.md)**
2. Implementa gramáticas con precedencia de operadores
3. Usa reglas de repetición (Kleene star/plus)
4. Revisa `/examples/advanced_grammar/`

### Para Usuarios Avanzados
1. Lee **[Limitaciones](limitaciones.md)** para entender los trade-offs
2. Estudia la implementación del parser mejorado
3. Contribuye con mejoras al proyecto
4. Comparte tus DSLs con la comunidad

## 🌟 Ejemplos Destacados

### Sistema Contable Multi-país
```go
// Soporta México, Colombia, Argentina, Perú
"venta de 1000"        // IVA según país
"asiento debe 1101 5000 haber 4001 5000"
```
Ver: `/examples/accounting/`

### Calculadora con Precedencia
```go
// Precedencia correcta de operadores
"2 + 3 * 4"    // = 14 (no 20)
"2 ^ 3 ^ 2"    // = 512 (asociativa derecha)
```
Ver: `/examples/advanced_grammar/`

### DSL Declarativo YAML
```yaml
name: "MiDSL"
tokens:
  NUMBER: "[0-9]+"
  PLUS: "\\+"
rules:
  - name: "expr"
    pattern: ["NUMBER", "PLUS", "NUMBER"]
    action: "add"
```
Ver: `/examples/declarative/`

## 🤝 Contribuir

¿Encontraste un error? ¿Tienes una sugerencia? ¡Tu ayuda es bienvenida!

1. Lee las guías de contribución
2. Abre un issue describiendo tu propuesta
3. Envía un pull request con tus cambios
4. Únete a las discusiones en GitHub

## 📬 Contacto y Soporte

- **GitHub Issues**: [Reportar problemas](https://github.com/arturoeanton/go-dsl/issues)
- **Discusiones**: [Preguntas y respuestas](https://github.com/arturoeanton/go-dsl/discussions)
- **Autor**: Arturo Elias Anton ([@arturoeanton](https://github.com/arturoeanton))

---

⭐ Si este proyecto te resulta útil, ¡no olvides darle una estrella en GitHub!