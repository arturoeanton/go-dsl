# DSL Definitions

Este directorio centraliza todas las definiciones de DSL (Domain Specific Languages) para el motor contable.

## Estructura

- **validation_dsl.go**: DSL para validación de comprobantes
  - Reglas de validación personalizables por organización
  - Validaciones de montos, cuentas, terceros, etc.

- **template_dsl.go**: DSL para templates de asientos recurrentes
  - Plantillas parametrizables para nómina, depreciación, etc.
  - Funciones de fecha y cálculo

- **tax_dsl.go**: DSL para cálculo automático de impuestos
  - IVA, retenciones, ICA
  - Tasas condicionales según tipo de tercero

- **distribution_dsl.go**: DSL para distribución de costos (futuro)
  - Prorrateo por drivers (área, headcount, etc.)
  - Distribuciones en cascada

- **reconciliation_dsl.go**: DSL para conciliación bancaria (futuro)
  - Reglas de matching automático
  - Tolerancias y patrones

## Uso

```go
// Ejemplo de uso del ValidationDSL
validationDSL := dsls.NewValidationDSL()
result, err := validationDSL.Parse(
    `if account_type("1110") == "BANK" AND voucher.reference == "" 
     then error "Las cuentas bancarias requieren referencia"`,
    context,
)
```

## Ventajas de Centralización

1. **Mantenibilidad**: Todas las definiciones DSL en un solo lugar
2. **Reutilización**: DSLs compartidos entre servicios
3. **Testabilidad**: Tests unitarios por DSL
4. **Documentación**: Cada DSL con su documentación específica
5. **Evolución**: Fácil agregar nuevos DSLs o modificar existentes