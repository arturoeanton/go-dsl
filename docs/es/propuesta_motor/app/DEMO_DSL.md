# 🚀 DEMO: Motor Contable con DSL Integrado

## 🎯 Casos de Uso Demostrados

### 1. Factura de Venta con IVA Automático

**Scenario**: Crear una factura de venta que automáticamente calcula IVA 19%

```bash
# Crear factura de venta base
curl -X POST http://localhost:3000/api/v1/vouchers \
  -H "Content-Type: application/json" \
  -d '{
    "voucher_type": "invoice_sale",
    "date": "2025-07-24T00:00:00Z",
    "description": "Venta de servicios de consultoría",
    "reference": "FV-2025-001",
    "third_party_id": "TP001",
    "voucher_lines": [
      {
        "account_id": "02d0cc5b7214aa0a543fe2c5224c86d7",
        "description": "Servicios de consultoría técnica",
        "debit_amount": 0,
        "credit_amount": 10000000,
        "third_party_id": "TP001"
      },
      {
        "account_id": "a757c937d68d833683d72c91c679a962", 
        "description": "Cuenta por cobrar cliente",
        "debit_amount": 11900000,
        "credit_amount": 0,
        "third_party_id": "TP001"
      }
    ]
  }'
```

**Resultado Esperado**:
- ✅ DSL detecta que es una factura de venta
- ✅ Calcula automáticamente IVA 19% sobre la base gravable
- ✅ Agrega línea adicional: IVA por pagar $1,900,000
- ✅ Aplica clasificaciones automáticas: revenue_type, tax_regime
- ✅ Valida que tenga cliente (third_party_id)

### 2. Factura de Compra con Retención Automática

**Scenario**: Compra mayor a $1,000,000 que genera retención automática

```bash
# Crear factura de compra con monto alto
curl -X POST http://localhost:3000/api/v1/vouchers \
  -H "Content-Type: application/json" \
  -d '{
    "voucher_type": "invoice_purchase",
    "date": "2025-07-24T00:00:00Z", 
    "description": "Compra de equipos de cómputo",
    "reference": "FC-2025-789",
    "third_party_id": "TP002",
    "voucher_lines": [
      {
        "account_id": "d1e05613ceab0efab7d3e0b6ad290345",
        "description": "Equipos de cómputo",
        "debit_amount": 5000000,
        "credit_amount": 0,
        "third_party_id": "TP002"
      },
      {
        "account_id": "a757c937d68d833683d72c91c679a962",
        "description": "Cuenta por pagar proveedor", 
        "debit_amount": 0,
        "credit_amount": 5950000,
        "third_party_id": "TP002"
      }
    ]
  }'
```

**Resultado Esperado**:
- ✅ Genera automáticamente IVA descontable 19%: $950,000
- ✅ Genera retención en la fuente 2.5%: $125,000
- ✅ Asigna centro de costo automático a gastos
- ✅ Marca como deductible y requiere documento soporte

### 3. Pago con Workflow de Aprobación

**Scenario**: Pago mayor a $5,000,000 que requiere aprobación

```bash
# Crear pago de alto valor
curl -X POST http://localhost:3000/api/v1/vouchers \
  -H "Content-Type: application/json" \
  -d '{
    "voucher_type": "payment",
    "date": "2025-07-24T00:00:00Z",
    "description": "Pago a proveedor servicios",
    "reference": "CE-2025-456", 
    "third_party_id": "TP002",
    "voucher_lines": [
      {
        "account_id": "a757c937d68d833683d72c91c679a962",
        "description": "Cancelación factura proveedor",
        "debit_amount": 8000000,
        "credit_amount": 0,
        "third_party_id": "TP002"
      },
      {
        "account_id": "7d3c841e89ca0d1aca70e06688a6028a",
        "description": "Salida banco",
        "debit_amount": 0,
        "credit_amount": 8000000,
        "third_party_id": null
      }
    ]
  }'
```

**Resultado Esperado**:
- ✅ Aplica retención automática 3.5% para pagos grandes
- ✅ Detecta salida de banco y requiere aprobación
- ✅ Clasifica como "requires_approval": true
- ✅ Al intentar procesar, bloquea con workflow "PAYMENT_APPROVAL"

### 4. Procesamiento (Post) con Notificaciones

**Scenario**: Procesar un comprobante y ver todas las reglas DSL en acción

```bash
# Primero crear un comprobante de alto valor
VOUCHER_ID=$(curl -s -X POST http://localhost:3000/api/v1/vouchers \
  -H "Content-Type: application/json" \
  -d '{
    "voucher_type": "invoice_sale",
    "date": "2025-07-24T00:00:00Z",
    "description": "Mega venta del año",
    "reference": "FV-MEGA-001",
    "third_party_id": "TP001",
    "voucher_lines": [
      {
        "account_id": "02d0cc5b7214aa0a543fe2c5224c86d7",
        "description": "Venta de proyecto grande",
        "debit_amount": 0,
        "credit_amount": 60000000,
        "third_party_id": "TP001"
      },
      {
        "account_id": "a757c937d68d833683d72c91c679a962",
        "description": "Por cobrar mega cliente",
        "debit_amount": 71400000,
        "credit_amount": 0,
        "third_party_id": "TP001"
      }
    ]
  }' | jq -r '.data.id')

# Procesar el comprobante
curl -X POST http://localhost:3000/api/v1/vouchers/$VOUCHER_ID/post
```

**Resultado Esperado**:
- ✅ Validaciones DSL pre-procesamiento ejecutadas
- ✅ Si es > $50M, requiere workflow "CEO_APPROVAL"
- ✅ Al procesar exitosamente:
  - Genera asiento contable automático
  - Ejecuta post-procesamiento DSL
  - Envía notificaciones: "🚨 ALERTA CRÍTICA: Comprobante de muy alto valor"
  - Registra metadata con todas las reglas aplicadas

### 5. Ver Comprobante con Todas las Reglas DSL Aplicadas

```bash
# Ver detalle del comprobante procesado
curl http://localhost:3000/api/v1/vouchers/$VOUCHER_ID | jq '.'
```

**Verás en additional_data**:
```json
{
  "auto_generated": true,
  "custom_fields": {
    "dsl_processed": true,
    "revenue_type": "operational",
    "tax_regime": "common", 
    "requires_electronic_invoice": true,
    "post_process_completed": true,
    "notifications_sent": 2,
    "dsl_rules_applied": [
      "tax_calculation",
      "retention_rules",
      "workflow_validation", 
      "cost_center_assignment",
      "notification_rules"
    ]
  }
}
```

## 🔥 Demo en Vivo

### Paso 1: Limpiar y preparar
```bash
# Detener servidor si está corriendo
pkill -f "go run main.go"

# Iniciar servidor fresco
cd /Users/arturoeliasanton/github.com/arturoeanton/go-dsl/docs/es/propuesta_motor/app
go run main.go &

# Esperar que inicie
sleep 3
```

### Paso 2: Crear factura con IVA automático
```bash
curl -X POST http://localhost:3000/api/v1/vouchers \
  -H "Content-Type: application/json" \
  -d '{
    "voucher_type": "invoice_sale",
    "date": "2025-07-24T00:00:00Z",
    "description": "Demo DSL - Venta con IVA automático",
    "reference": "DEMO-001",
    "third_party_id": "TP001",
    "voucher_lines": [
      {
        "account_id": "02d0cc5b7214aa0a543fe2c5224c86d7",
        "description": "Venta base gravable",
        "debit_amount": 0,
        "credit_amount": 1000000,
        "third_party_id": "TP001"
      },
      {
        "account_id": "a757c937d68d833683d72c91c679a962",
        "description": "Por cobrar cliente",
        "debit_amount": 1000000,
        "credit_amount": 0,
        "third_party_id": "TP001"
      }
    ]
  }' | jq '.'
```

**Observa**: El sistema rechazará porque no está balanceado. El DSL detectó que necesita IVA pero el usuario no lo incluyó en el total.

### Paso 3: Crear factura correcta con total incluyendo IVA
```bash
curl -X POST http://localhost:3000/api/v1/vouchers \
  -H "Content-Type: application/json" \
  -d '{
    "voucher_type": "invoice_sale", 
    "date": "2025-07-24T00:00:00Z",
    "description": "Demo DSL - Venta con IVA automático",
    "reference": "DEMO-002",
    "third_party_id": "TP001",
    "voucher_lines": [
      {
        "account_id": "02d0cc5b7214aa0a543fe2c5224c86d7",
        "description": "Venta base gravable",
        "debit_amount": 0,
        "credit_amount": 1000000,
        "third_party_id": "TP001"
      },
      {
        "account_id": "a757c937d68d833683d72c91c679a962",
        "description": "Por cobrar cliente (incluye IVA)",
        "debit_amount": 1190000,
        "credit_amount": 0, 
        "third_party_id": "TP001"
      }
    ]
  }' | jq '.'
```

**Observa**: 
- ✅ Comprobante creado exitosamente
- ✅ DSL agregó automáticamente línea de IVA
- ✅ Verificar que ahora tiene 3 líneas en lugar de 2

### Paso 4: Ver el comprobante con líneas automáticas
```bash
# Obtener el ID del último comprobante
VOUCHER_ID=$(curl -s http://localhost:3000/api/v1/vouchers | jq -r '.data.vouchers[0].id')

# Ver detalle
curl http://localhost:3000/api/v1/vouchers/$VOUCHER_ID | jq '.data.voucher_lines'
```

**Verás**: 3 líneas, incluyendo la línea de IVA generada automáticamente por DSL

### Paso 5: Intentar procesar con workflow requerido
```bash
# Crear pago grande que requiere aprobación
PAYMENT_ID=$(curl -s -X POST http://localhost:3000/api/v1/vouchers \
  -H "Content-Type: application/json" \
  -d '{
    "voucher_type": "payment",
    "date": "2025-07-24T00:00:00Z",
    "description": "Pago importante - Demo workflow DSL",
    "reference": "PAGO-DEMO-001",
    "third_party_id": "TP002",
    "voucher_lines": [
      {
        "account_id": "a757c937d68d833683d72c91c679a962",
        "description": "Pago a proveedor",
        "debit_amount": 15000000,
        "credit_amount": 0,
        "third_party_id": "TP002"
      },
      {
        "account_id": "7d3c841e89ca0d1aca70e06688a6028a",
        "description": "Salida banco",
        "debit_amount": 0,
        "credit_amount": 15000000
      }
    ]
  }' | jq -r '.data.id')

# Intentar procesar
curl -X POST http://localhost:3000/api/v1/vouchers/$PAYMENT_ID/post | jq '.'
```

**Observa**: Error con mensaje del workflow requerido por DSL

## 📊 Verificación en Dashboard

1. Abrir http://localhost:3000/dashboard.html
2. Ver los KPIs actualizados con comprobantes creados
3. Ir a Lista de Comprobantes
4. Verificar estados y montos
5. Abrir detalle de un comprobante para ver líneas automáticas

## 🎉 Conclusión

El POC demuestra exitosamente:

1. **Validaciones DSL**: Pre y post procesamiento
2. **Generación Automática**: IVA, retenciones según reglas
3. **Clasificaciones**: Centros de costo, tipos automáticos  
4. **Workflows**: Aprobaciones según montos y tipos
5. **Notificaciones**: Alertas para valores críticos
6. **Integración Completa**: DSL embebido en todo el flujo

¡El motor contable con go-dsl está completamente funcional y listo para casos de uso complejos!