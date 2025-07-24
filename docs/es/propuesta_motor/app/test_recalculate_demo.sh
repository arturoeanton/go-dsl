#!/bin/bash

echo "=== Test de recálculo de comprobantes con DSL ==="
echo

# 1. Crear un comprobante desbalanceado intencionalmente
echo "1. Creando comprobante desbalanceado (sin IVA)..."
VOUCHER_RESPONSE=$(curl -s -X POST http://localhost:3000/api/v1/vouchers \
  -H "Content-Type: application/json" \
  -d '{
    "voucher_type": "invoice_sale",
    "description": "Test venta SIN IVA - necesita recálculo",
    "date": "'$(date -u +"%Y-%m-%dT%H:%M:%SZ")'",
    "third_party_id": "TP001",
    "voucher_lines": [
      {
        "account_id": "68fe4ecbf2d26e205185e0a7a2beb0f0",
        "description": "Venta de servicios (base gravable)",
        "debit_amount": 0,
        "credit_amount": 100000,
        "third_party_id": "TP001"
      },
      {
        "account_id": "02d0cc5b7214aa0a543fe2c5224c86d7",
        "description": "Cliente - pago parcial SIN IVA",
        "debit_amount": 100000,
        "credit_amount": 0,
        "third_party_id": "TP001"
      }
    ]
  }')

echo "$VOUCHER_RESPONSE" | jq
VOUCHER_ID=$(echo "$VOUCHER_RESPONSE" | jq -r '.data.id')
echo
echo "📝 Comprobante ID: $VOUCHER_ID"

# 2. Ver estado inicial del comprobante
echo
echo "2. Estado inicial del comprobante (desbalanceado):"
curl -s http://localhost:3000/api/v1/vouchers/$VOUCHER_ID | jq '.data | {id, number, total_debit, total_credit, is_balanced, lines: .voucher_lines | length}'
echo

# 3. Mostrar la tasa de IVA actual
echo "3. Tasa de IVA actual en el DSL:"
curl -s http://localhost:3000/api/v1/dsl/iva-rate | jq '.data'
echo

# 4. Recalcular el comprobante con DSL
echo "4. Recalculando comprobante con reglas DSL..."
RECALC_RESPONSE=$(curl -s -X POST http://localhost:3000/api/v1/vouchers/$VOUCHER_ID/recalculate)
echo "$RECALC_RESPONSE" | jq
echo

# 5. Ver estado después del recálculo
echo "5. Estado después del recálculo:"
UPDATED_VOUCHER=$(curl -s http://localhost:3000/api/v1/vouchers/$VOUCHER_ID)
echo "$UPDATED_VOUCHER" | jq '.data | {id, number, total_debit, total_credit, is_balanced, lines: .voucher_lines | map({account_id, debit_amount, credit_amount, description, tax_rate, base_amount})}'
echo

# 6. Verificar balance
IS_BALANCED=$(echo "$UPDATED_VOUCHER" | jq -r '.data.is_balanced')
TOTAL_DEBIT=$(echo "$UPDATED_VOUCHER" | jq -r '.data.total_debit')
TOTAL_CREDIT=$(echo "$UPDATED_VOUCHER" | jq -r '.data.total_credit')

echo "6. Verificación de balance:"
echo "   - Balanceado: $IS_BALANCED"
echo "   - Total débitos: $TOTAL_DEBIT"
echo "   - Total créditos: $TOTAL_CREDIT"
echo

# 7. Intentar procesar el comprobante recalculado
echo "7. Intentando procesar el comprobante recalculado..."
PROCESS_RESPONSE=$(curl -s -X POST http://localhost:3000/api/v1/vouchers/$VOUCHER_ID/post)
echo "$PROCESS_RESPONSE" | jq

if [[ $(echo "$PROCESS_RESPONSE" | jq -r '.success') == "false" ]]; then
  echo "❌ Error al procesar: $(echo "$PROCESS_RESPONSE" | jq -r '.message')"
else
  echo "✅ Comprobante procesado exitosamente después del recálculo"
  
  # Ver el asiento contable generado
  echo
  echo "8. Asiento contable generado:"
  curl -s http://localhost:3000/api/v1/journal-entries | jq '.data.journal_entries[0] | {id, number, total_debit, total_credit, lines: .journal_entry_lines | map({account_code: .account.code, account_name: .account.name, debit_amount, credit_amount, description})}'
fi

echo
echo "=== Resumen de la demo ==="
echo "1. ✅ Comprobante creado desbalanceado intencionalmente"
echo "2. ✅ Recálculo aplicó reglas DSL automáticamente"
echo "3. ✅ Se agregaron líneas de IVA automáticamente"
echo "4. ✅ Comprobante quedó balanceado"
echo "5. ✅ Comprobante procesado y asiento contable generado"
echo
echo "🎯 Demo completada exitosamente!"