#!/bin/bash

echo "=== Test de creación y procesamiento de comprobante ==="
echo

# Crear un comprobante de venta simple
echo "1. Creando comprobante de venta..."
VOUCHER_RESPONSE=$(curl -s -X POST http://localhost:3000/api/v1/vouchers \
  -H "Content-Type: application/json" \
  -d '{
    "voucher_type": "invoice_sale",
    "description": "Test venta con IVA",
    "date": "'$(date -u +"%Y-%m-%dT%H:%M:%SZ")'",
    "third_party_id": "TP001",
    "voucher_lines": [
      {
        "account_id": "68fe4ecbf2d26e205185e0a7a2beb0f0",
        "description": "Venta de servicios",
        "debit_amount": 0,
        "credit_amount": 100000,
        "third_party_id": "TP001"
      },
      {
        "account_id": "d1e05613ceab0efab7d3e0b6ad290345",
        "description": "Pago en efectivo",
        "debit_amount": 100000,
        "credit_amount": 0
      }
    ]
  }')

echo "$VOUCHER_RESPONSE" | jq
VOUCHER_ID=$(echo "$VOUCHER_RESPONSE" | jq -r '.data.id')
echo

# Ver detalle del comprobante
echo "2. Detalle del comprobante creado:"
curl -s http://localhost:3000/api/v1/vouchers/$VOUCHER_ID | jq '.data | {id, total_debit, total_credit, is_balanced, lines: .voucher_lines | map({account_id, debit_amount, credit_amount, description})}'
echo

# Intentar procesar el comprobante
echo "3. Procesando comprobante..."
PROCESS_RESPONSE=$(curl -s -X POST http://localhost:3000/api/v1/vouchers/$VOUCHER_ID/post)
echo "$PROCESS_RESPONSE" | jq
echo

# Si falló, ver el error
if [[ $(echo "$PROCESS_RESPONSE" | jq -r '.success') == "false" ]]; then
  echo "❌ Error al procesar: $(echo "$PROCESS_RESPONSE" | jq -r '.message')"
else
  echo "✅ Comprobante procesado exitosamente"
  
  # Ver el asiento contable generado
  echo
  echo "4. Asiento contable generado:"
  curl -s http://localhost:3000/api/v1/journal-entries | jq '.data.journal_entries[0]'
fi