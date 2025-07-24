# 📚 Documentación API - Motor Contable

## 🌐 Información General

**Base URL**: `http://localhost:3000/api/v1`  
**Formato**: JSON  
**Autenticación**: No implementada en POC

## 📋 Endpoints Disponibles

### 🏢 Organizaciones

#### GET /organizations
Obtiene la lista de organizaciones.

**Respuesta exitosa (200)**:
```json
{
  "success": true,
  "data": [
    {
      "id": "b1dc2897-beee-4761-8abe-8234d3e8fd23",
      "name": "Empresa Demo S.A.S",
      "tax_id": "900123456-7",
      "address": "Calle 100 #15-20, Bogotá",
      "phone": "+57 1 234 5678",
      "email": "contacto@empresademo.com",
      "created_at": "2025-01-01T00:00:00Z"
    }
  ]
}
```

### 📊 Dashboard

#### GET /dashboard/kpis
Obtiene los KPIs principales del sistema.

**Respuesta exitosa (200)**:
```json
{
  "success": true,
  "data": {
    "total_vouchers": 145,
    "pending_vouchers": 12,
    "total_amount": 125500000,
    "accounts_count": 257,
    "active_templates": 8
  }
}
```

#### GET /dashboard/stats
Obtiene estadísticas detalladas.

**Respuesta exitosa (200)**:
```json
{
  "success": true,
  "data": {
    "vouchers_by_type": {
      "invoice_sale": 45,
      "invoice_purchase": 38,
      "payment": 22,
      "receipt": 18,
      "general": 22
    },
    "vouchers_by_status": {
      "DRAFT": 12,
      "POSTED": 120,
      "CANCELLED": 13
    },
    "monthly_totals": [
      {"month": "2025-01", "total": 45000000},
      {"month": "2024-12", "total": 38000000}
    ]
  }
}
```

### 📝 Comprobantes

#### GET /vouchers
Obtiene lista paginada de comprobantes.

**Parámetros Query**:
- `page` (int): Número de página (default: 1)
- `limit` (int): Items por página (default: 20)
- `status` (string): Filtrar por estado
- `type` (string): Filtrar por tipo

**Respuesta exitosa (200)**:
```json
{
  "success": true,
  "data": {
    "vouchers": [
      {
        "id": "v001",
        "number": "FV-2025-0001",
        "voucher_type": "invoice_sale",
        "date": "2025-01-24T10:30:00Z",
        "description": "Venta de servicios profesionales",
        "total_debit": 1190000,
        "total_credit": 1190000,
        "status": "POSTED",
        "is_balanced": true
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 145,
      "pages": 8
    },
    "stats": {
      "total_vouchers": 145,
      "total_amount": 125500000,
      "pending_count": 12,
      "error_count": 0
    }
  }
}
```

#### POST /vouchers
Crea un nuevo comprobante.

**Body**:
```json
{
  "voucher_type": "invoice_sale",
  "date": "2025-01-24T10:30:00Z",
  "description": "Venta de productos",
  "reference": "REF-001",
  "third_party_id": "TP001",
  "voucher_lines": [
    {
      "account_id": "68fe4ecbf2d26e205185e0a7a2beb0f0",
      "description": "Venta de servicios",
      "debit_amount": 0,
      "credit_amount": 1000000,
      "third_party_id": "TP001"
    },
    {
      "account_id": "d1e05613ceab0efab7d3e0b6ad290345",
      "description": "Pago en efectivo",
      "debit_amount": 1000000,
      "credit_amount": 0
    }
  ]
}
```

**Respuesta exitosa (201)**:
```json
{
  "success": true,
  "data": {
    "id": "v002",
    "number": "FV-2025-0002",
    "voucher_type": "invoice_sale",
    "status": "DRAFT",
    "total_debit": 1190000,
    "total_credit": 1190000,
    "is_balanced": true,
    "voucher_lines": [
      {
        "line_number": 1,
        "account_id": "68fe4ecbf2d26e205185e0a7a2beb0f0",
        "description": "Venta de servicios",
        "debit_amount": 0,
        "credit_amount": 1000000
      },
      {
        "line_number": 2,
        "account_id": "d1e05613ceab0efab7d3e0b6ad290345",
        "description": "Pago en efectivo",
        "debit_amount": 1190000,
        "credit_amount": 0
      },
      {
        "line_number": 3,
        "account_id": "33dc34fa7cf3e08f2e277e067b96f22a",
        "description": "IVA 19% generado por DSL",
        "debit_amount": 0,
        "credit_amount": 190000,
        "metadata": {
          "generated_by": "dsl_rules_engine",
          "rule": "automatic_tax_generation"
        }
      }
    ]
  }
}
```

#### GET /vouchers/:id
Obtiene un comprobante específico.

**Respuesta exitosa (200)**:
```json
{
  "success": true,
  "data": {
    "id": "v001",
    "number": "FV-2025-0001",
    "voucher_type": "invoice_sale",
    "date": "2025-01-24T10:30:00Z",
    "description": "Venta de servicios profesionales",
    "reference": "REF-001",
    "status": "POSTED",
    "total_debit": 1190000,
    "total_credit": 1190000,
    "is_balanced": true,
    "third_party": {
      "id": "TP001",
      "name": "Cliente Demo S.A.S",
      "tax_id": "900111222-3"
    },
    "voucher_lines": [
      {
        "line_number": 1,
        "account": {
          "code": "413510",
          "name": "VENTA DE SERVICIOS"
        },
        "description": "Servicios profesionales",
        "debit_amount": 0,
        "credit_amount": 1000000
      }
    ],
    "journal_entry": {
      "id": "je001",
      "number": "AS-2025-0001",
      "posted_at": "2025-01-24T10:35:00Z"
    }
  }
}
```

#### POST /vouchers/:id/post
Procesa y contabiliza un comprobante.

**Respuesta exitosa (200)**:
```json
{
  "success": true,
  "message": "Comprobante procesado exitosamente",
  "data": {
    "voucher_id": "v001",
    "journal_entry_id": "je001",
    "posted_at": "2025-01-24T10:35:00Z"
  }
}
```

**Error por workflow (400)**:
```json
{
  "success": false,
  "error": {
    "code": "WORKFLOW_REQUIRED",
    "message": "Este comprobante requiere aprobación mediante el workflow: CFO_APPROVAL",
    "details": {
      "amount": 25000000,
      "threshold": 20000000,
      "approver_role": "CFO"
    }
  }
}
```

#### POST /vouchers/from-template
Crea un comprobante desde un template DSL.

**Body**:
```json
{
  "template_id": "tpl-invoice-sale-001",
  "parameters": {
    "customer_name": "Cliente ABC",
    "amount": 500000,
    "description": "Servicios de consultoría"
  }
}
```

**Respuesta exitosa (201)**:
```json
{
  "success": true,
  "data": {
    "id": "v003",
    "number": "FV-2025-0003",
    "description": "Generado desde template: Factura de Venta Estándar",
    "template_applied": "tpl-invoice-sale-001"
  }
}
```

### 📚 Plan de Cuentas

#### GET /accounts
Obtiene el plan de cuentas.

**Parámetros Query**:
- `search` (string): Buscar por código o nombre
- `level` (int): Filtrar por nivel (1-6)
- `accepts_movement` (bool): Solo cuentas que aceptan movimiento

**Respuesta exitosa (200)**:
```json
{
  "success": true,
  "data": {
    "accounts": [
      {
        "id": "acc001",
        "code": "1",
        "name": "ACTIVO",
        "level": 1,
        "nature": "DEBIT",
        "account_type": "ASSET",
        "accepts_movement": false,
        "is_active": true,
        "children_count": 15
      },
      {
        "id": "acc002",
        "code": "110505",
        "name": "CAJA GENERAL",
        "level": 3,
        "nature": "DEBIT",
        "account_type": "ASSET",
        "accepts_movement": true,
        "is_active": true,
        "parent_code": "1105"
      }
    ],
    "total": 257
  }
}
```

### 🔧 Templates DSL

#### GET /dsl/templates
Obtiene los templates DSL disponibles.

**Respuesta exitosa (200)**:
```json
{
  "success": true,
  "data": [
    {
      "id": "tpl-tax-001",
      "name": "Cálculo IVA 19%",
      "category": "tax_rules",
      "description": "Genera automáticamente el IVA del 19% para ventas",
      "status": "ACTIVE",
      "version": 1
    },
    {
      "id": "tpl-retention-001",
      "name": "Retención en Compras",
      "category": "retention_rules",
      "description": "Aplica retención del 2.5% en compras > $1M",
      "status": "ACTIVE",
      "version": 1
    }
  ]
}
```

#### POST /dsl/validate
Valida una regla DSL sin ejecutarla.

**Body**:
```json
{
  "dsl_code": "rule test_rule { when { voucher.amount > 1000000 } then { apply_retention(0.025) } }",
  "context": {
    "voucher_type": "invoice_purchase",
    "amount": 2000000
  }
}
```

**Respuesta exitosa (200)**:
```json
{
  "success": true,
  "data": {
    "valid": true,
    "would_trigger": true,
    "actions": ["apply_retention with rate 0.025"]
  }
}
```

### 👥 Terceros

#### GET /third-parties
Obtiene lista de terceros.

**Respuesta exitosa (200)**:
```json
{
  "success": true,
  "data": [
    {
      "id": "TP001",
      "name": "Cliente Demo S.A.S",
      "tax_id": "900111222-3",
      "third_party_type": "CUSTOMER",
      "email": "cliente@demo.com",
      "phone": "+57 300 123 4567",
      "is_active": true
    }
  ]
}
```

## 🔴 Manejo de Errores

Todos los endpoints siguen el mismo formato de error:

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Descripción del error",
    "details": {}
  }
}
```

### Códigos de Error Comunes

| Código | Descripción | HTTP Status |
|--------|-------------|-------------|
| `VALIDATION_ERROR` | Datos de entrada inválidos | 400 |
| `NOT_FOUND` | Recurso no encontrado | 404 |
| `UNBALANCED_VOUCHER` | Comprobante no balanceado | 400 |
| `WORKFLOW_REQUIRED` | Requiere aprobación | 400 |
| `INVALID_ACCOUNT` | Cuenta inválida o inactiva | 400 |
| `SERVER_ERROR` | Error interno del servidor | 500 |

## 🚀 Casos de Uso

### 1. Crear Venta con IVA Automático

```bash
curl -X POST http://localhost:3000/api/v1/vouchers \
  -H "Content-Type: application/json" \
  -d '{
    "voucher_type": "invoice_sale",
    "date": "2025-01-24T10:00:00Z",
    "description": "Venta de productos",
    "voucher_lines": [
      {
        "account_id": "68fe4ecbf2d26e205185e0a7a2beb0f0",
        "credit_amount": 100000,
        "debit_amount": 0
      },
      {
        "account_id": "d1e05613ceab0efab7d3e0b6ad290345",
        "debit_amount": 100000,
        "credit_amount": 0
      }
    ]
  }'
```

El DSL agregará automáticamente una línea de IVA del 19%.

### 2. Crear Compra con Retención

```bash
curl -X POST http://localhost:3000/api/v1/vouchers \
  -H "Content-Type: application/json" \
  -d '{
    "voucher_type": "invoice_purchase",
    "date": "2025-01-24T10:00:00Z",
    "description": "Compra de equipos",
    "voucher_lines": [
      {
        "account_id": "2938717a9252a428b0f1963a49cf087f",
        "debit_amount": 2000000,
        "credit_amount": 0
      },
      {
        "account_id": "a757c937d68d833683d72c91c679a962",
        "credit_amount": 2000000,
        "debit_amount": 0
      }
    ]
  }'
```

Para compras > $1M, el DSL aplicará retención del 2.5%.

### 3. Verificar Workflow de Aprobación

```bash
# 1. Crear pago grande
VOUCHER_ID=$(curl -X POST http://localhost:3000/api/v1/vouchers \
  -H "Content-Type: application/json" \
  -d '{
    "voucher_type": "payment",
    "date": "2025-01-24T10:00:00Z",
    "description": "Pago a proveedor",
    "voucher_lines": [
      {
        "account_id": "a757c937d68d833683d72c91c679a962",
        "debit_amount": 25000000,
        "credit_amount": 0
      },
      {
        "account_id": "7d3c841e89ca0d1aca70e06688a6028a",
        "credit_amount": 25000000,
        "debit_amount": 0
      }
    ]
  }' | jq -r '.data.id')

# 2. Intentar procesar (fallará por workflow)
curl -X POST http://localhost:3000/api/v1/vouchers/$VOUCHER_ID/post
```

## 📦 Colecciones Postman

Importa el archivo `motor-contable.postman_collection.json` incluido en el repositorio para acceder a todos los endpoints preconfigurados.

## 🔗 Enlaces Relacionados

- [README Principal](README.md)
- [Documentación DSL](https://github.com/arturoeanton/go-dsl)
- [Swagger UI](http://localhost:3000/swagger)