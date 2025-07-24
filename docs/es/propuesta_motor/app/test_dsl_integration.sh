#!/bin/bash

# Script de prueba completo para la integración DSL
# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=====================================${NC}"
echo -e "${BLUE}🧪 Test de Integración DSL Completa${NC}"
echo -e "${BLUE}=====================================${NC}"

# Función para verificar respuesta
check_response() {
    local response=$1
    local test_name=$2
    
    if echo "$response" | jq -e '.success == true' > /dev/null 2>&1; then
        echo -e "${GREEN}✅ $test_name: PASÓ${NC}"
        return 0
    else
        echo -e "${RED}❌ $test_name: FALLÓ${NC}"
        echo -e "${YELLOW}Error: $(echo "$response" | jq -r '.message // "Error desconocido"')${NC}"
        return 1
    fi
}

# Verificar que el servidor esté corriendo
echo -e "\n${YELLOW}1. Verificando servidor...${NC}"
if curl -s http://localhost:3000/health | grep -q "ok"; then
    echo -e "${GREEN}✅ Servidor funcionando${NC}"
else
    echo -e "${RED}❌ Servidor no responde. Iniciando...${NC}"
    cd /Users/arturoeliasanton/github.com/arturoeanton/go-dsl/docs/es/propuesta_motor/app
    go run main.go &
    sleep 5
fi

# Test 1: Crear factura de venta con IVA automático
echo -e "\n${YELLOW}2. Test: Factura de venta con IVA automático${NC}"
VENTA_RESPONSE=$(curl -s -X POST http://localhost:3000/api/v1/vouchers \
  -H "Content-Type: application/json" \
  -d '{
    "voucher_type": "invoice_sale",
    "date": "2025-07-24T00:00:00Z",
    "description": "Test DSL - Venta con IVA",
    "reference": "TEST-DSL-001",
    "third_party_id": "TP001",
    "voucher_lines": [
      {
        "account_id": "02d0cc5b7214aa0a543fe2c5224c86d7",
        "description": "Venta de servicios",
        "debit_amount": 0,
        "credit_amount": 100000,
        "third_party_id": "TP001"
      },
      {
        "account_id": "a757c937d68d833683d72c91c679a962",
        "description": "Por cobrar cliente",
        "debit_amount": 119000,
        "credit_amount": 0,
        "third_party_id": "TP001"
      }
    ]
  }')

if check_response "$VENTA_RESPONSE" "Crear factura de venta"; then
    VENTA_ID=$(echo "$VENTA_RESPONSE" | jq -r '.data.id')
    
    # Verificar que se generó el IVA
    LINES=$(curl -s http://localhost:3000/api/v1/vouchers/$VENTA_ID | jq '.data.voucher_lines | length')
    if [ "$LINES" -eq "3" ]; then
        echo -e "${GREEN}✅ IVA generado automáticamente (3 líneas)${NC}"
    else
        echo -e "${RED}❌ IVA no generado (esperado 3 líneas, encontrado $LINES)${NC}"
    fi
fi

# Test 2: Factura de compra con retención
echo -e "\n${YELLOW}3. Test: Factura de compra con retención automática${NC}"
COMPRA_RESPONSE=$(curl -s -X POST http://localhost:3000/api/v1/vouchers \
  -H "Content-Type: application/json" \
  -d '{
    "voucher_type": "invoice_purchase",
    "date": "2025-07-24T00:00:00Z",
    "description": "Test DSL - Compra con retención",
    "reference": "TEST-DSL-002",
    "third_party_id": "TP002",
    "voucher_lines": [
      {
        "account_id": "d1e05613ceab0efab7d3e0b6ad290345",
        "description": "Compra de suministros",
        "debit_amount": 2000000,
        "credit_amount": 0,
        "third_party_id": "TP002"
      },
      {
        "account_id": "a757c937d68d833683d72c91c679a962",
        "description": "Por pagar proveedor",
        "debit_amount": 0,
        "credit_amount": 2380000,
        "third_party_id": "TP002"
      }
    ]
  }')

if check_response "$COMPRA_RESPONSE" "Crear factura de compra"; then
    COMPRA_ID=$(echo "$COMPRA_RESPONSE" | jq -r '.data.id')
    
    # Verificar líneas generadas
    LINES_INFO=$(curl -s http://localhost:3000/api/v1/vouchers/$COMPRA_ID | \
        jq -r '.data.voucher_lines[] | "\(.description)"')
    
    if echo "$LINES_INFO" | grep -q "IVA" && echo "$LINES_INFO" | grep -q "Retención"; then
        echo -e "${GREEN}✅ IVA y Retención generados automáticamente${NC}"
    else
        echo -e "${RED}❌ Faltan líneas automáticas${NC}"
    fi
fi

# Test 3: Pago con workflow de aprobación
echo -e "\n${YELLOW}4. Test: Pago con workflow de aprobación${NC}"
PAGO_RESPONSE=$(curl -s -X POST http://localhost:3000/api/v1/vouchers \
  -H "Content-Type: application/json" \
  -d '{
    "voucher_type": "payment",
    "date": "2025-07-24T00:00:00Z",
    "description": "Test DSL - Pago grande",
    "reference": "TEST-DSL-003",
    "third_party_id": "TP002",
    "voucher_lines": [
      {
        "account_id": "a757c937d68d833683d72c91c679a962",
        "description": "Pago a proveedor",
        "debit_amount": 6000000,
        "credit_amount": 0,
        "third_party_id": "TP002"
      },
      {
        "account_id": "7d3c841e89ca0d1aca70e06688a6028a",
        "description": "Salida banco",
        "debit_amount": 0,
        "credit_amount": 6000000
      }
    ]
  }')

if check_response "$PAGO_RESPONSE" "Crear pago"; then
    PAGO_ID=$(echo "$PAGO_RESPONSE" | jq -r '.data.id')
    
    # Intentar procesar (debe fallar por workflow)
    POST_RESPONSE=$(curl -s -X POST http://localhost:3000/api/v1/vouchers/$PAGO_ID/post)
    
    if echo "$POST_RESPONSE" | grep -q "workflow"; then
        echo -e "${GREEN}✅ Workflow de aprobación activado correctamente${NC}"
    else
        echo -e "${RED}❌ Workflow no se activó${NC}"
    fi
fi

# Test 4: Verificar clasificaciones automáticas
echo -e "\n${YELLOW}5. Test: Clasificaciones automáticas DSL${NC}"
if [ ! -z "$VENTA_ID" ]; then
    METADATA=$(curl -s http://localhost:3000/api/v1/vouchers/$VENTA_ID | \
        jq '.data.additional_data.custom_fields')
    
    if echo "$METADATA" | grep -q "dsl_processed" && \
       echo "$METADATA" | grep -q "revenue_type"; then
        echo -e "${GREEN}✅ Clasificaciones DSL aplicadas${NC}"
        echo -e "${BLUE}Metadata: $(echo $METADATA | jq -c .)${NC}"
    else
        echo -e "${RED}❌ Clasificaciones DSL no encontradas${NC}"
    fi
fi

# Test 5: POS - Punto de venta
echo -e "\n${YELLOW}6. Test: POS - Venta rápida${NC}"
POS_RESPONSE=$(curl -s -X POST http://localhost:3000/api/v1/vouchers \
  -H "Content-Type: application/json" \
  -d '{
    "voucher_type": "invoice_sale",
    "date": "2025-07-24T00:00:00Z",
    "description": "Venta POS: Café x2, Sandwich x1",
    "reference": "POS-'$(date +%s)'",
    "third_party_id": "TP001",
    "voucher_lines": [
      {
        "account_id": "02d0cc5b7214aa0a543fe2c5224c86d7",
        "description": "Venta productos",
        "debit_amount": 0,
        "credit_amount": 20000,
        "third_party_id": "TP001"
      },
      {
        "account_id": "d1e05613ceab0efab7d3e0b6ad290345",
        "description": "Pago efectivo",
        "debit_amount": 23800,
        "credit_amount": 0
      }
    ]
  }')

check_response "$POS_RESPONSE" "Venta POS"

# Test 6: Verificar notificaciones para montos altos
echo -e "\n${YELLOW}7. Test: Notificaciones DSL para montos altos${NC}"
ALTO_VALOR_RESPONSE=$(curl -s -X POST http://localhost:3000/api/v1/vouchers \
  -H "Content-Type: application/json" \
  -d '{
    "voucher_type": "invoice_sale",
    "date": "2025-07-24T00:00:00Z",
    "description": "Test DSL - Venta de alto valor",
    "reference": "TEST-DSL-HIGH",
    "third_party_id": "TP001",
    "voucher_lines": [
      {
        "account_id": "02d0cc5b7214aa0a543fe2c5224c86d7",
        "description": "Mega proyecto",
        "debit_amount": 0,
        "credit_amount": 60000000,
        "third_party_id": "TP001"
      },
      {
        "account_id": "a757c937d68d833683d72c91c679a962",
        "description": "Por cobrar",
        "debit_amount": 71400000,
        "credit_amount": 0,
        "third_party_id": "TP001"
      }
    ]
  }')

if check_response "$ALTO_VALOR_RESPONSE" "Crear venta alto valor"; then
    echo -e "${BLUE}Revisar logs del servidor para notificaciones DSL${NC}"
fi

# Resumen
echo -e "\n${BLUE}=====================================${NC}"
echo -e "${BLUE}📊 Resumen de Pruebas DSL${NC}"
echo -e "${BLUE}=====================================${NC}"

# Estadísticas
TOTAL_VOUCHERS=$(curl -s http://localhost:3000/api/v1/vouchers | jq '.data.vouchers | length')
echo -e "${GREEN}Total comprobantes creados: $TOTAL_VOUCHERS${NC}"

# KPIs
KPIS=$(curl -s http://localhost:3000/api/v1/dashboard/kpis)
echo -e "${GREEN}KPIs actualizados:${NC}"
echo "$KPIS" | jq -r '
  "- Comprobantes hoy: \(.data.vouchers_today)",
  "- Comprobantes mes: \(.data.vouchers_month)",
  "- Pendientes: \(.data.pending_vouchers)",
  "- Tasa procesamiento: \(.data.processing_rate)%"
'

echo -e "\n${GREEN}✅ Pruebas de integración DSL completadas${NC}"
echo -e "${BLUE}Acceder a http://localhost:3000 para ver el sistema completo${NC}"
echo -e "${BLUE}POS disponible en http://localhost:3000/pos.html${NC}"