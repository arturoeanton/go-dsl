# HU-004: Motor de Contabilización con DSL

## Historia de Usuario

**Como** administrador contable  
**Quiero** configurar reglas de contabilización dinámicas  
**Para** automatizar la generación de asientos según el tipo de comprobante y país

## Criterios de Aceptación

1. ✅ Procesar comprobantes y generar asientos automáticamente
2. ✅ Plantillas DSL configurables por tipo y país
3. ✅ Validación de sintaxis DSL en tiempo real
4. ✅ Preview de asientos antes de confirmar
5. ✅ Manejo de errores con mensajes claros
6. ✅ Versionado de plantillas
7. ✅ Rollback en caso de fallo
8. ✅ Logs detallados de procesamiento
9. ✅ Procesamiento en lote (hasta 1000)
10. ✅ Notificaciones de resultado

## Especificaciones Técnicas

- **Motor DSL**: go-dsl integrado
- **Procesamiento**: Asíncrono con NATS
- **Cache**: Plantillas compiladas en Redis
- **Transacciones**: ACID garantizado
- **Performance**: 100+ comprobantes/segundo

## Tareas de Desarrollo

### 1. Backend - Integración go-dsl (4h)
- [ ] Integrar librería go-dsl al proyecto
- [ ] Definir gramática para contabilidad
- [ ] Parser para plantillas contables
- [ ] Compilador a AST optimizado
- [ ] Runtime de ejecución

### 2. Backend - Template Service (5h)
- [ ] CRUD de plantillas DSL
- [ ] Validación de sintaxis
- [ ] Compilación y cache
- [ ] Versionado automático
- [ ] Selección por tipo/país
- [ ] Plantillas por defecto

### 3. Backend - Motor de Contabilización (8h)
- [ ] `AccountingEngine` principal
- [ ] Procesador de comprobantes
- [ ] Generador de asientos
- [ ] Validador de balances
- [ ] Manejo de transacciones
- [ ] Sistema de rollback

### 4. Backend - Procesamiento Asíncrono (4h)
- [ ] Worker pool con NATS
- [ ] Colas por prioridad
- [ ] Reintentos automáticos
- [ ] Dead letter queue
- [ ] Monitoreo de workers

### 5. Backend - API Endpoints (3h)
- [ ] `POST /vouchers/:id/process` individual
- [ ] `POST /vouchers/bulk-process` masivo
- [ ] `GET /processing/status/:jobId`
- [ ] `POST /templates/validate` validar DSL
- [ ] `POST /templates/preview` preview asientos

### 6. DSL - Definición del Lenguaje (6h)
- [ ] Sintaxis para variables
- [ ] Operadores matemáticos
- [ ] Condicionales (if/else)
- [ ] Bucles para items
- [ ] Funciones built-in
- [ ] Acceso a metadatos

### 7. Frontend - Editor DSL (6h)
- [ ] Monaco Editor integrado
- [ ] Syntax highlighting para DSL
- [ ] Autocompletado inteligente
- [ ] Validación en tiempo real
- [ ] Snippets predefinidos
- [ ] Diff viewer para versiones

### 8. Frontend - Testing de Plantillas (4h)
- [ ] Sandbox para pruebas
- [ ] Datos de prueba mock
- [ ] Visualización de AST
- [ ] Debug step-by-step
- [ ] Comparación de resultados

### 9. Frontend - Monitor de Procesamiento (3h)
- [ ] Dashboard de jobs
- [ ] Progreso en tiempo real
- [ ] Logs de procesamiento
- [ ] Estadísticas de rendimiento
- [ ] Alertas de errores

### 10. Optimizaciones (4h)
- [ ] Pool de plantillas compiladas
- [ ] Batch processing eficiente
- [ ] Índices para consultas frecuentes
- [ ] Compresión de logs
- [ ] Cleanup de jobs antiguos

### 11. Testing (5h)
- [ ] Tests del parser DSL
- [ ] Tests de plantillas complejas
- [ ] Tests de concurrencia
- [ ] Tests de rollback
- [ ] Benchmarks de performance

### 12. Documentación (3h)
- [ ] Manual del lenguaje DSL
- [ ] Ejemplos por tipo de documento
- [ ] Guía de troubleshooting
- [ ] Best practices

## Estimación Total: 55 horas

## Dependencias

- HU-002: Gestión de Comprobantes
- HU-003: Catálogo de Cuentas
- Librería go-dsl

## Riesgos

1. **Complejidad del DSL**: Mantener sintaxis simple pero poderosa
2. **Performance con plantillas complejas**: Cache agresivo + optimización AST
3. **Errores en producción**: Validación exhaustiva + rollback automático
4. **Deadlocks en procesamiento**: Timeouts + circuit breakers

## Notas de Implementación

### Ejemplo de Plantilla DSL

```dsl
template invoice_sale_co {
  // Definir variables desde el comprobante
  let subtotal = voucher.metadata.subtotal
  let tax_rate = 0.19
  let tax_amount = voucher.metadata.taxes.iva_19
  let total = voucher.total_amount
  
  // Validaciones
  require subtotal > 0 : "Subtotal debe ser positivo"
  require tax_amount == subtotal * tax_rate : "IVA calculado incorrectamente"
  require total == subtotal + tax_amount : "Total no cuadra"
  
  // Generar asiento contable
  entry {
    // Debitar cuentas por cobrar
    debit account("1305.01") amount(total) {
      description = "Factura " + voucher.voucher_number
      metadata.customer_id = voucher.metadata.customer.id
      metadata.due_date = date_add(voucher.voucher_date, 30, "days")
    }
    
    // Acreditar ventas por líneas
    for item in voucher.metadata.items {
      credit account(item.income_account) amount(item.subtotal) {
        description = item.description
        cost_center = item.cost_center
        project = item.project_id
      }
    }
    
    // Acreditar IVA si aplica
    if tax_amount > 0 {
      credit account("2408.01") amount(tax_amount) {
        description = "IVA 19% por pagar"
        metadata.tax_period = format_date(voucher.voucher_date, "YYYY-MM")
      }
    }
  }
  
  // Post-procesamiento
  after {
    // Actualizar saldo del cliente
    update_customer_balance(voucher.metadata.customer.id, total)
    
    // Notificar si es monto alto
    if total > 10000000 {
      notify("high_value_invoice", {
        invoice_number: voucher.voucher_number,
        amount: total,
        customer: voucher.metadata.customer.name
      })
    }
  }
}
```

### Estructura del Motor

```go
type AccountingEngine struct {
    dslParser    *dsl.Parser
    templateSvc  *TemplateService
    accountSvc   *AccountService
    jobQueue     *JobQueue
    workers      []*Worker
}

type ProcessingJob struct {
    ID           string
    VoucherID    string
    TemplateID   string
    Status       JobStatus
    Result       *JournalEntry
    Error        error
    StartedAt    time.Time
    CompletedAt  time.Time
}

func (e *AccountingEngine) ProcessVoucher(ctx context.Context, voucherID string) (*ProcessingJob, error) {
    // 1. Cargar comprobante
    voucher, err := e.voucherSvc.GetByID(ctx, voucherID)
    
    // 2. Seleccionar plantilla
    template, err := e.templateSvc.SelectTemplate(voucher.Type, voucher.Country)
    
    // 3. Compilar si no está en cache
    compiled, err := e.compileTemplate(template)
    
    // 4. Ejecutar plantilla
    entry, err := e.executeTemplate(compiled, voucher)
    
    // 5. Validar asiento
    if err := e.validateEntry(entry); err != nil {
        return nil, err
    }
    
    // 6. Guardar asiento
    return e.saveEntry(ctx, entry)
}
```

## Mockups Relacionados

- [Editor DSL](../mocks/front/html/dsl_editor.html)
- [Monitor de Procesamiento](../mocks/front/html/processing_monitor.html)
- [Preview de Asientos](../mocks/front/html/entry_preview.html)

## Métricas de Éxito

- Comprobantes procesados por minuto: > 6,000
- Tasa de error: < 0.1%
- Tiempo de compilación DSL: < 50ms
- Disponibilidad del motor: 99.9%
- Adopción de plantillas custom: > 60%