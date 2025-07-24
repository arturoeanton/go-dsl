# Guía para Crear Comprobantes - Motor Contable POC

## 📋 Instrucciones paso a paso para la interfaz web

### 1. Acceder al formulario
- URL: `http://localhost:3000/vouchers_form.html`
- Recargar con `Ctrl+Shift+R` para asegurar la última versión

### 2. Llenar información básica
- **Tipo de Comprobante**: "Factura de Compra"
- **Referencia**: "TEST-001" (o cualquier referencia única)
- **Fecha**: Dejar la fecha de hoy
- **Descripción**: "Prueba de comprobante"
- **Tercero**: "Proveedor ABC Ltda"

### 3. Agregar líneas contables (MUY IMPORTANTE: debe estar balanceado)

#### Primera línea:
- **Cuenta**: "110505 - CAJA GENERAL"
- **Descripción**: "Gasto"
- **Débito**: **100**
- **Crédito**: **0**
- **Tercero**: Cualquiera

#### Segunda línea:
- **Cuenta**: "130505 - CLIENTES NACIONALES"
- **Descripción**: "Por pagar"
- **Débito**: **0**
- **Crédito**: **100**
- **Tercero**: Cualquiera

### 4. Validaciones importantes
- ✅ **Balance**: Suma de débitos = Suma de créditos
- ✅ **Mínimo 2 líneas**: Requerido por el sistema
- ✅ **Cuentas válidas**: Deben existir y aceptar movimiento

### 5. Guardar
- Hacer clic en "💾 Guardar Borrador" o "✅ Crear y Procesar"

---

## 🔧 Usando cURL para crear comprobantes

### Ejemplo básico de comprobante balanceado:

```bash
curl -X POST http://localhost:3000/api/v1/vouchers \
  -H "Content-Type: application/json" \
  -d '{
    "voucher_type": "PURCHASE",
    "date": "2025-07-24T00:00:00Z",
    "description": "Compra de suministros de oficina",
    "reference": "COMP-001",
    "third_party_id": null,
    "voucher_lines": [
      {
        "account_id": "d1e05613ceab0efab7d3e0b6ad290345",
        "description": "Gastos de oficina",
        "debit_amount": 100,
        "credit_amount": 0,
        "third_party_id": null
      },
      {
        "account_id": "a757c937d68d833683d72c91c679a962",
        "description": "Cuenta por pagar proveedor",
        "debit_amount": 0,
        "credit_amount": 100,
        "third_party_id": null
      }
    ]
  }'
```

### Ejemplo con múltiples líneas y balanceado:

```bash
curl -X POST http://localhost:3000/api/v1/vouchers \
  -H "Content-Type: application/json" \
  -d '{
    "voucher_type": "SALE",
    "date": "2025-07-24T00:00:00Z",
    "description": "Venta de productos",
    "reference": "FV-001",
    "third_party_id": null,
    "voucher_lines": [
      {
        "account_id": "7d3c841e89ca0d1aca70e06688a6028a",
        "description": "Ingreso a bancos",
        "debit_amount": 119,
        "credit_amount": 0,
        "third_party_id": null
      },
      {
        "account_id": "02d0cc5b7214aa0a543fe2c5224c86d7",
        "description": "Venta de productos",
        "debit_amount": 0,
        "credit_amount": 100,
        "third_party_id": null
      },
      {
        "account_id": "d34b750ba305132c7196b47c4c528d6f",
        "description": "IVA cobrado",
        "debit_amount": 0,
        "credit_amount": 19,
        "third_party_id": null
      }
    ]
  }'
```

---

## 🔍 Verificar el comprobante creado

### 1. Listar comprobantes:
```bash
curl http://localhost:3000/api/v1/vouchers
```

### 2. Ver detalle de un comprobante:
```bash
curl http://localhost:3000/api/v1/vouchers/{voucher_id}
```

---

## ⚠️ Errores comunes y soluciones

### Error 400: "Comprobante no está balanceado"
- **Problema**: Suma de débitos ≠ Suma de créditos
- **Solución**: Verificar que la suma de todos los débitos sea exactamente igual a la suma de todos los créditos

### Error 400: "Un comprobante debe tener al menos 2 líneas"
- **Problema**: Menos de 2 líneas contables
- **Solución**: Agregar al menos 2 líneas con movimientos válidos

### Error 400: "Cuenta no encontrada"
- **Problema**: account_id no existe
- **Solución**: Usar un ID de cuenta válido de la lista de cuentas disponibles

---

## 📊 IDs de cuentas comunes (para cURL)

| Código | Cuenta | ID |
|--------|--------|-----|
| 110505 | CAJA GENERAL | d1e05613ceab0efab7d3e0b6ad290345 |
| 111005 | BANCOS NACIONALES | 7d3c841e89ca0d1aca70e06688a6028a |
| 130505 | CLIENTES NACIONALES | 02d0cc5b7214aa0a543fe2c5224c86d7 |
| 220501 | PROVEEDORES NACIONALES | a757c937d68d833683d72c91c679a962 |
| 240802 | IVA DESCONTABLE | d34b750ba305132c7196b47c4c528d6f |

### Para obtener la lista completa de cuentas:
```bash
curl http://localhost:3000/api/v1/accounts
```

---

## 🎯 Tips para la demo

1. **Siempre usar valores enteros** para evitar problemas de precisión
2. **Verificar balance antes de enviar** (suma débitos = suma créditos)
3. **Usar referencias únicas** para evitar conflictos
4. **Probar primero con montos pequeños** (ej: 100, 200)
5. **Recargar página con Ctrl+Shift+R** si hay problemas de caché

---

## 🚀 Estados de comprobantes

- **DRAFT**: Borrador, se puede editar
- **POSTED**: Procesado, genera asientos contables
- **CANCELLED**: Cancelado

**Nota**: Por defecto se crean en estado DRAFT. Para procesarlos y generar asientos automáticamente, usar el botón "✅ Crear y Procesar" en la interfaz web.