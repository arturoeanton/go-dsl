# Ejemplos de Templates - Motor Contable

Este documento muestra cómo diferentes templates generan asientos contables distintos para la misma transacción.

## Caso de Uso: Factura de Venta por $1,000,000

### Datos de Entrada
```json
{
  "invoice_number": "FV-001-2024",
  "customer_name": "Cliente ABC S.A.S",
  "base_amount": 1000000,
  "invoice_date": "2024-01-31"
}
```

### Template 1: Factura de Venta Estándar
**Descripción**: Aplica IVA del 19% sin retenciones

**Asiento Generado**:
```
Fecha: 2024-01-31
Descripción: Factura Venta FV-001-2024 - Cliente ABC S.A.S
Referencia: FV-001-2024

Cuenta    | Descripción                    | Débito      | Crédito
----------|--------------------------------|-------------|------------
130505    | CxC Cliente ABC S.A.S          | $1,190,000  |
413595    | Venta de productos             |             | $1,000,000
240802    | IVA generado 19%               |             | $190,000
----------|--------------------------------|-------------|------------
TOTALES   |                                | $1,190,000  | $1,190,000
```

### Template 2: Factura de Venta con Retención
**Descripción**: Aplica IVA del 19% y retención en la fuente del 2.5%

**Asiento Generado**:
```
Fecha: 2024-01-31
Descripción: Factura Venta FV-001-2024 - Cliente ABC S.A.S (con retención)
Referencia: FV-001-2024

Cuenta    | Descripción                           | Débito      | Crédito
----------|---------------------------------------|-------------|------------
130505    | CxC Cliente ABC S.A.S (neto)          | $1,165,000  |
135515    | Retención en la fuente por cobrar     | $25,000     |
413595    | Venta de productos                    |             | $1,000,000
240802    | IVA generado 19%                      |             | $190,000
----------|---------------------------------------|-------------|------------
TOTALES   |                                       | $1,190,000  | $1,190,000
```

## Comparación de Resultados

| Concepto                    | Template Estándar | Template con Retención |
|-----------------------------|-------------------|------------------------|
| Total Factura               | $1,190,000        | $1,190,000            |
| CxC Directo al Cliente      | $1,190,000        | $1,165,000            |
| Retención por Cobrar        | $0                | $25,000               |
| Valor a Recibir del Cliente | $1,190,000        | $1,165,000            |

## Caso de Uso 2: Nómina del Empleado

### Datos de Entrada
```json
{
  "employee_name": "Juan Pérez",
  "basic_salary": 2000000,
  "period": "2024-01"
}
```

### Template: Nómina Mensual Básica
**Asiento Generado**:
```
Fecha: 2024-01-31
Descripción: Nómina 2024-01 - Juan Pérez
Referencia: NOM-2024-01

Cuenta    | Descripción                    | Débito      | Crédito
----------|--------------------------------|-------------|------------
510506    | Salario básico                 | $2,000,000  |
510568    | Aporte patronal salud          | $170,000    |
510569    | Aporte patronal pensión        | $240,000    |
510570    | ARL                            | $10,440     |
237005    | Salud empleado por pagar       |             | $80,000
238030    | Pensión empleado por pagar     |             | $80,000
237006    | Salud patronal por pagar       |             | $170,000
238031    | Pensión patronal por pagar     |             | $240,000
237010    | ARL por pagar                  |             | $10,440
250505    | Salario neto por pagar         |             | $1,840,000
----------|--------------------------------|-------------|------------
TOTALES   |                                | $2,420,440  | $2,420,440
```

**Resumen**:
- Costo total para la empresa: $2,420,440
- Salario neto del empleado: $1,840,000
- Deducciones del empleado: $160,000 (8%)
- Aportes patronales: $420,440

## Caso de Uso 3: Compra con Diferentes Tipos

### Datos de Entrada Base
```json
{
  "invoice_number": "FC-789",
  "supplier_name": "Proveedor XYZ",
  "base_amount": 500000,
  "purchase_date": "2024-01-15"
}
```

### Variación 1: Compra de Inventario
```json
{
  ...datos_base,
  "purchase_type": "inventory"
}
```

**Asiento**:
```
Cuenta    | Descripción                    | Débito      | Crédito
----------|--------------------------------|-------------|------------
143505    | Compra de inventario           | $500,000    |
240801    | IVA descontable 19%            | $95,000     |
220505    | CxP Proveedor XYZ              |             | $595,000
```

### Variación 2: Compra de Gastos
```json
{
  ...datos_base,
  "purchase_type": "expense"
}
```

**Asiento**:
```
Cuenta    | Descripción                    | Débito      | Crédito
----------|--------------------------------|-------------|------------
519595    | Gastos generales               | $500,000    |
240801    | IVA descontable 19%            | $95,000     |
220505    | CxP Proveedor XYZ              |             | $595,000
```

## Ventajas del Sistema de Templates

1. **Flexibilidad**: Un mismo tipo de transacción puede generar diferentes asientos según las necesidades del negocio.

2. **Configurabilidad**: Los usuarios pueden crear sus propios templates sin modificar código.

3. **Consistencia**: Garantiza que las transacciones similares se registren de manera uniforme.

4. **Adaptabilidad**: Fácil adaptación a diferentes regulaciones fiscales o políticas contables.

5. **Trazabilidad**: Cada asiento generado mantiene referencia al template utilizado.

## Funciones DSL Disponibles

- `last_day(period)`: Obtiene el último día del período
- `if/else`: Condicionales para lógica compleja
- Operadores aritméticos: `+`, `-`, `*`, `/`
- Concatenación de strings: `+`
- Variables calculadas: `set variable = expresión`

## Próximos Pasos

1. Implementar validaciones de cuentas contables
2. Agregar soporte para múltiples monedas
3. Incluir templates para consolidación
4. Desarrollar templates para cierre de período