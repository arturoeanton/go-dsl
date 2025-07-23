# Sistema Contable DSL - Ejemplo Completo

**Un sistema contable empresarial completo construido con go-dsl que demuestra todas las características avanzadas del framework.**

## 🎯 Objetivo

Este ejemplo demuestra cómo crear un **sistema contable de nivel empresarial** usando go-dsl, incluyendo:

- ✅ Gramáticas recursivas por la izquierda para asientos complejos
- ✅ KeywordToken para resolver conflictos de tokenización
- ✅ Validación de balances contables automática
- ✅ Cálculo de IVA y transacciones con impuestos
- ✅ Sistema de cuentas contables completo
- ✅ Asientos manuales complejos y balanceados
- ✅ Estabilidad de producción (sin errores intermitentes)

## 🚀 Ejecución Rápida

```bash
cd examples/contabilidad
go run main.go
```

## 📚 Características del DSL

### Tokens Definidos

```go
// Keywords con prioridad alta (90)
contabilidad.KeywordToken("VENTA", "venta")
contabilidad.KeywordToken("COMPRA", "compra") 
contabilidad.KeywordToken("DE", "de")
contabilidad.KeywordToken("CON", "con")
contabilidad.KeywordToken("IVA", "iva")
contabilidad.KeywordToken("ASIENTO", "asiento")
contabilidad.KeywordToken("DEBE", "debe")
contabilidad.KeywordToken("HABER", "haber")

// Valores con prioridad normal (0)
contabilidad.Token("IMPORTE", "[0-9]+\\.?[0-9]*")
contabilidad.Token("STRING", "\"[^\"]*\"")
```

### Comandos Soportados

#### 1. Operaciones Básicas de Venta
```
venta de 1000                    # Venta simple
venta de 5000 con iva           # Venta con IVA (16%)
venta de 3000 a cliente "ABC"   # Venta a cliente específico
```

#### 2. Operaciones de Compra
```
compra de 2000                  # Compra simple
compra de 4000 con iva         # Compra con IVA acreditable
```

#### 3. Operaciones de Tesorería
```
cobro de cliente "ABC" por 3480      # Cobro a cliente
pago a proveedor "XYZ" por 2320      # Pago a proveedor
nomina de empleado "Juan" por 15000  # Pago de nómina
gasto de "Papelería" por 500         # Registro de gasto
```

#### 4. Asientos Manuales Complejos (Recursión Izquierda)
```
asiento debe 1101 10000 debe 1401 1600 haber 2101 11600
```

**¡Esta gramática recursiva por la izquierda era imposible antes!** Ahora funciona perfectamente gracias al ImprovedParser con memoización.

## 🏗️ Arquitectura del Sistema

### Tipos de Datos

```go
type Asiento struct {
    ID          int
    Fecha       time.Time
    Descripcion string
    Movimientos []Movimiento
}

type Movimiento struct {
    Cuenta      string
    Descripcion string
    Debe        float64
    Haber       float64
}

type SistemaContable struct {
    Cuentas  map[string]*Cuenta
    Asientos []Asiento
    Contador int
    IVA      float64  // 16% por defecto (México)
}
```

### Plan de Cuentas

| Código | Descripción | Tipo |
|--------|-------------|------|
| 1101 | Bancos | Activo |
| 1201 | Clientes | Activo |
| 1401 | IVA acreditable | Activo |
| 2101 | Proveedores | Pasivo |
| 2401 | IVA por pagar | Pasivo |
| 3101 | Capital social | Patrimonio |
| 4101 | Ventas | Ingreso |
| 5101 | Compras | Gasto |
| 5201 | Sueldos y salarios | Gasto |
| 5301 | Gastos generales | Gasto |

## 🔧 Características Técnicas Avanzadas

### 1. Instancias DSL Frescas para Estabilidad

```go
// ✅ PATRÓN RECOMENDADO: Nueva instancia para cada operación
func procesarOperacion(comando string) (interface{}, error) {
    contabilidad := createContabilidadDSL(sistema)  // Nueva instancia
    return contabilidad.Parse(comando)
}
```

**¿Por qué es importante?**
- Elimina condiciones de carrera
- Garantiza estabilidad en sistemas concurrentes
- Evita errores intermitentes
- Recomendado para sistemas de producción

### 2. Gramáticas Recursivas por la Izquierda

```go
// Estas reglas antes causaban stack overflow
contabilidad.Rule("movements", []string{"movement"}, "singleMovement")
contabilidad.Rule("movements", []string{"movements", "movement"}, "multipleMovements")
```

**Cómo funciona:**
- ImprovedParser con memoización (packrat parsing)
- Detección automática de recursión izquierda
- Cache de resultados parciales para evitar loops infinitos

### 3. KeywordToken vs Token

```go
// ✅ CORRECTO: Keywords con prioridad alta
contabilidad.KeywordToken("DEBE", "debe")  // Prioridad 90

// Token genérico con prioridad baja
contabilidad.Token("IMPORTE", "[0-9]+")    // Prioridad 0
```

**Ventajas:**
- Resuelve conflictos automáticamente
- No depende del orden de definición
- Funciona 100% del tiempo sin excepciones

### 4. Validación Contable Automática

```go
contabilidad.Action("processEntry", func(args []interface{}) (interface{}, error) {
    movements := args[1].([]Movimiento)
    
    // Validar que Debe = Haber
    totalDebe := 0.0
    totalHaber := 0.0
    for _, m := range movements {
        totalDebe += m.Debe
        totalHaber += m.Haber
    }
    
    if totalDebe != totalHaber {
        return nil, fmt.Errorf("asiento descuadrado: %.2f != %.2f", totalDebe, totalHaber)
    }
    
    return createAsiento(sistema, "Asiento manual", movements), nil
})
```

## 📊 Ejemplo de Salida

```
=== Sistema Contable DSL Mejorado ===

1. Venta simple de $1,000:
   Comando: venta de 1000

   Asiento #1 - 2025-07-23
   Venta por 1000.00
   -------------------------------------------------------
   1201 Clientes                      $1000.00
   4101 Ventas                                     $1000.00
   -------------------------------------------------------
   Totales:                        $1000.00     $1000.00

10. Asiento manual complejo:
   Comando: asiento debe 1101 10000 debe 1401 1600 haber 2101 11600

   Asiento #10 - 2025-07-23
   Asiento manual
   -------------------------------------------------------
   1101 Bancos                       $10000.00
   1401 IVA acreditable               $1600.00
   2101 Proveedores                               $11600.00
   -------------------------------------------------------
   Totales:                       $11600.00    $11600.00

Balanza de Comprobación:
========================
✅ La balanza está cuadrada
```

## 🎓 Lecciones Aprendidas

### 1. **KeywordToken es Esencial**
Sin KeywordToken, palabras como "debe" y "haber" pueden ser capturadas por tokens genéricos, causando errores intermitentes.

### 2. **Instancias Frescas para Sistemas Críticos**
Para máxima estabilidad, crear nuevas instancias DSL para cada operación.

### 3. **Recursión Izquierda Funciona**
go-dsl ahora maneja perfectamente gramáticas complejas que antes eran imposibles.

### 4. **Validación de Reglas de Negocio**
Las acciones pueden implementar validaciones complejas (como balances contables).

## 🔗 Casos de Uso Similares

Este patrón se puede adaptar para:

- **Sistemas de Facturación**: Con diferentes tipos de documentos
- **Inventarios**: Con movimientos de entrada/salida
- **Nóminas**: Con cálculos complejos de deducciones
- **Presupuestos**: Con asignaciones y transferencias
- **Auditoría**: Con trazabilidad completa de transacciones

## 🚀 Próximos Pasos

1. **Prueba el ejemplo**: `go run main.go`
2. **Modifica los comandos** en el código
3. **Agrega nuevas operaciones** contables
4. **Implementa tu propio plan de cuentas**
5. **Integra con una base de datos** real

## 📞 Soporte

- **Documentación**: [Guía Rápida](../../docs/es/guia_rapida.md)
- **Manual Completo**: [Manual de Uso](../../docs/es/manual_de_uso.md)
- **Para Contribuidores**: [Developer Onboarding](../../docs/es/developer_onboarding.md)

---

**¡Este ejemplo demuestra que go-dsl está listo para sistemas empresariales de producción!** 🎉