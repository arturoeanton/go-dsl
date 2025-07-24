# 🚀 Guía Completa de Demo - Motor Contable con DSL

## 📋 Tabla de Contenidos

1. [Preparación del Ambiente](#preparación-del-ambiente)
2. [Demo 1: POS - Punto de Venta Simplificado](#demo-1-pos---punto-de-venta-simplificado)
3. [Demo 2: Cambio Dinámico de Reglas DSL](#demo-2-cambio-dinámico-de-reglas-dsl)
4. [Demo 3: Flujo Completo con Workflows](#demo-3-flujo-completo-con-workflows)
5. [Demo 4: Dashboard y Reportes](#demo-4-dashboard-y-reportes)
6. [Troubleshooting](#troubleshooting)

---

## 🔧 Preparación del Ambiente

### Paso 1: Iniciar el servidor
```bash
# Navegar al directorio
cd /Users/arturoeliasanton/github.com/arturoeanton/go-dsl/docs/es/propuesta_motor/app

# Detener cualquier instancia previa
pkill -f "go run main.go" 2>/dev/null

# Iniciar el servidor
go run main.go &

# Esperar que inicie (verificar logs)
sleep 3

# Verificar que esté funcionando
curl http://localhost:3000/health
```

### Paso 2: Verificar datos iniciales
```bash
# Ver organizaciones
curl http://localhost:3000/api/v1/organizations | jq '.data[0]'

# Ver algunas cuentas contables
curl http://localhost:3000/api/v1/accounts?limit=5 | jq '.data.accounts[].name'

# Ver plantillas DSL cargadas
sqlite3 db_contable.db "SELECT name FROM dsl_templates WHERE status='ACTIVE';"
```

---

## 🛒 Demo 1: POS - Punto de Venta Simplificado

### Contexto
Simularemos un punto de venta donde:
- Las ventas se registran rápidamente
- El IVA se calcula automáticamente
- Se generan comprobantes contables completos

### Paso 1: Crear función helper para POS
```bash
# Crear archivo de funciones POS
cat > pos_demo.sh << 'EOF'
#!/bin/bash

# Función para registrar venta POS
registrar_venta_pos() {
    local descripcion="$1"
    local monto_base="$2"
    local cliente="${3:-Cliente Mostrador}"
    
    # Calcular total con IVA
    local monto_total=$(echo "$monto_base * 1.19" | bc)
    
    echo "🛒 Registrando venta: $descripcion"
    echo "💰 Base: \$$(printf "%'.0f" $monto_base)"
    echo "💸 Total con IVA: \$$(printf "%'.0f" $monto_total)"
    
    # Generar referencia única
    local ref="POS-$(date +%Y%m%d%H%M%S)"
    
    # Crear comprobante
    local response=$(curl -s -X POST http://localhost:3000/api/v1/vouchers \
      -H "Content-Type: application/json" \
      -d "{
        \"voucher_type\": \"invoice_sale\",
        \"date\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",
        \"description\": \"Venta POS: $descripcion\",
        \"reference\": \"$ref\",
        \"third_party_id\": \"TP001\",
        \"voucher_lines\": [
          {
            \"account_id\": \"02d0cc5b7214aa0a543fe2c5224c86d7\",
            \"description\": \"$descripcion\",
            \"debit_amount\": 0,
            \"credit_amount\": $monto_base,
            \"third_party_id\": \"TP001\"
          },
          {
            \"account_id\": \"d1e05613ceab0efab7d3e0b6ad290345\",
            \"description\": \"Pago en efectivo - $cliente\",
            \"debit_amount\": $monto_total,
            \"credit_amount\": 0,
            \"third_party_id\": null
          }
        ]
      }")
    
    local voucher_id=$(echo "$response" | jq -r '.data.id')
    local success=$(echo "$response" | jq -r '.success')
    
    if [ "$success" = "true" ]; then
        echo "✅ Venta registrada: $ref (ID: $voucher_id)"
        
        # Ver las líneas generadas (incluyendo IVA automático)
        echo "📋 Detalle del comprobante:"
        curl -s http://localhost:3000/api/v1/vouchers/$voucher_id | \
          jq -r '.data.voucher_lines[] | "   - \(.description): D: \(.debit_amount) C: \(.credit_amount)"'
    else
        echo "❌ Error al registrar venta:"
        echo "$response" | jq '.message'
    fi
    echo "---"
}

# Función para ver ventas del día
ventas_del_dia() {
    local fecha=$(date +%Y-%m-%d)
    echo "📊 Ventas del día $fecha:"
    
    curl -s "http://localhost:3000/api/v1/vouchers?start_date=$fecha&end_date=$fecha" | \
      jq -r '.data.vouchers[] | select(.voucher_type == "invoice_sale") | 
        "[\(.date | split("T")[0])] \(.reference): \(.description) - Total: $\(.total_debit)"'
}

# Exportar funciones
export -f registrar_venta_pos
export -f ventas_del_dia
EOF

# Hacer ejecutable
chmod +x pos_demo.sh

# Cargar funciones
source pos_demo.sh
```

### Paso 2: Simular ventas POS
```bash
# Venta 1: Producto simple
registrar_venta_pos "Café Americano Grande" 5000

# Venta 2: Múltiples productos
registrar_venta_pos "Combo almuerzo ejecutivo" 25000

# Venta 3: Venta mayor
registrar_venta_pos "Catering evento corporativo" 500000

# Ver resumen de ventas
ventas_del_dia
```

### Observaciones Esperadas:
- ✅ Cada venta genera automáticamente la línea de IVA
- ✅ Los comprobantes quedan balanceados
- ✅ Se aplican las reglas DSL de clasificación

---

## 🔄 Demo 2: Cambio Dinámico de Reglas DSL

### Contexto
Demostraremos cómo cambiar las reglas DSL en tiempo real:
1. Cambiar tasa de IVA de 19% a 16%
2. Agregar descuento automático para ventas > $100,000
3. Cambiar límites de aprobación

### Paso 1: Ver regla actual de IVA
```bash
# Ver la regla actual en el código
grep -n "0.19" internal/services/dsl_rules_engine.go
```

### Paso 2: Crear versión actualizada del motor DSL
```bash
# Hacer backup del actual
cp internal/services/dsl_rules_engine.go internal/services/dsl_rules_engine.go.bak

# Crear script para actualizar reglas
cat > update_dsl_rules.sh << 'EOF'
#!/bin/bash

# Función para cambiar tasa de IVA
cambiar_tasa_iva() {
    local nueva_tasa=$1
    local tasa_decimal=$(echo "scale=2; $nueva_tasa / 100" | bc)
    
    echo "🔄 Cambiando tasa de IVA a $nueva_tasa%..."
    
    # Actualizar en el motor DSL
    sed -i.tmp "s/0\.19/$tasa_decimal/g" internal/services/dsl_rules_engine.go
    sed -i.tmp "s/19%/$nueva_tasa%/g" internal/services/dsl_rules_engine.go
    
    echo "✅ Tasa actualizada. Reiniciando servidor..."
    
    # Reiniciar servidor
    pkill -f "go run main.go"
    sleep 2
    go run main.go &
    sleep 3
    
    echo "✅ Servidor reiniciado con nueva tasa de IVA: $nueva_tasa%"
}

# Función para cambiar límites de aprobación
cambiar_limite_aprobacion() {
    local tipo=$1
    local nuevo_limite=$2
    
    echo "🔄 Cambiando límite de $tipo a \$$(printf "%'.0f" $nuevo_limite)..."
    
    case $tipo in
        "CEO")
            sed -i.tmp "s/voucher\.TotalDebit > 50000000/voucher.TotalDebit > $nuevo_limite/g" \
              internal/services/dsl_rules_engine.go
            ;;
        "CFO")
            sed -i.tmp "s/voucher\.TotalDebit > 20000000/voucher.TotalDebit > $nuevo_limite/g" \
              internal/services/dsl_rules_engine.go
            ;;
        "PAYMENT")
            sed -i.tmp "s/voucher\.TotalDebit > 5000000/voucher.TotalDebit > $nuevo_limite/g" \
              internal/services/dsl_rules_engine.go
            ;;
    esac
    
    echo "✅ Límite actualizado"
}

export -f cambiar_tasa_iva
export -f cambiar_limite_aprobacion
EOF

chmod +x update_dsl_rules.sh
source update_dsl_rules.sh
```

### Paso 3: Probar cambio de IVA
```bash
# Crear venta con IVA 19%
registrar_venta_pos "Producto antes del cambio" 100000

# Cambiar IVA a 16%
cambiar_tasa_iva 16

# Crear venta con IVA 16%
registrar_venta_pos "Producto después del cambio" 100000

# Comparar los dos comprobantes
echo "📊 Comparación de IVA:"
curl -s http://localhost:3000/api/v1/vouchers | \
  jq -r '.data.vouchers[0:2] | .[] | 
    "[\(.reference)] IVA: \(.voucher_lines[] | select(.description | contains("IVA")) | .credit_amount)"'
```

### Paso 4: Agregar regla de descuento automático
```bash
# Agregar función de descuento al motor DSL
cat >> internal/services/dsl_rules_engine.go << 'EOF'

// GenerateAutomaticDiscount genera descuentos basados en montos
func (e *DSLRulesEngine) GenerateAutomaticDiscount(voucher *models.Voucher) ([]models.VoucherLine, error) {
	var discountLines []models.VoucherLine
	
	// Descuento del 5% para ventas > $100,000
	if voucher.VoucherType == "invoice_sale" && voucher.TotalCredit > 100000 {
		discountAmount := voucher.TotalCredit * 0.05
		
		discountLine := models.VoucherLine{
			AccountID:    "419595000000000000000000000000000", // Cuenta descuentos
			Description:  "Descuento 5% por volumen",
			DebitAmount:  discountAmount,
			CreditAmount: 0,
			LineNumber:   len(voucher.VoucherLines) + 1,
		}
		
		discountLines = append(discountLines, discountLine)
		
		fmt.Printf("[DSL] Descuento automático aplicado: $%.2f\n", discountAmount)
	}
	
	return discountLines, nil
}
EOF

# Reiniciar servidor
pkill -f "go run main.go"
go run main.go &
sleep 3
```

### Paso 5: Probar límites de aprobación
```bash
# Cambiar límite de aprobación de pagos a $3,000,000
cambiar_limite_aprobacion "PAYMENT" 3000000

# Crear pago de $4,000,000 (debería requerir aprobación)
curl -X POST http://localhost:3000/api/v1/vouchers \
  -H "Content-Type: application/json" \
  -d '{
    "voucher_type": "payment",
    "date": "2025-07-24T00:00:00Z",
    "description": "Pago prueba nuevo límite",
    "reference": "PAGO-TEST-001",
    "third_party_id": "TP002",
    "voucher_lines": [
      {
        "account_id": "a757c937d68d833683d72c91c679a962",
        "description": "Pago a proveedor",
        "debit_amount": 4000000,
        "credit_amount": 0,
        "third_party_id": "TP002"
      },
      {
        "account_id": "7d3c841e89ca0d1aca70e06688a6028a",
        "description": "Salida banco",
        "debit_amount": 0,
        "credit_amount": 4000000
      }
    ]
  }' | jq -r '.data.id' > pago_id.txt

# Intentar procesar (debería fallar por workflow)
curl -X POST http://localhost:3000/api/v1/vouchers/$(cat pago_id.txt)/post | jq '.'
```

---

## 🔄 Demo 3: Flujo Completo con Workflows

### Contexto
Mostrar el flujo completo desde creación hasta procesamiento con todos los pasos DSL.

### Paso 1: Crear comprobante complejo
```bash
# Factura de compra grande con múltiples líneas
COMPRA_ID=$(curl -s -X POST http://localhost:3000/api/v1/vouchers \
  -H "Content-Type: application/json" \
  -d '{
    "voucher_type": "invoice_purchase",
    "date": "2025-07-24T00:00:00Z",
    "description": "Compra de mobiliario y equipos",
    "reference": "FC-2025-1234",
    "third_party_id": "TP002",
    "voucher_lines": [
      {
        "account_id": "152405000000000000000000000000001",
        "description": "Muebles y enseres",
        "debit_amount": 8000000,
        "credit_amount": 0,
        "third_party_id": "TP002"
      },
      {
        "account_id": "152805000000000000000000000000001", 
        "description": "Equipos de cómputo",
        "debit_amount": 12000000,
        "credit_amount": 0,
        "third_party_id": "TP002"
      },
      {
        "account_id": "a757c937d68d833683d72c91c679a962",
        "description": "Por pagar a proveedor",
        "debit_amount": 0,
        "credit_amount": 20000000,
        "third_party_id": "TP002"
      }
    ]
  }' | jq -r '.data.id')

echo "📋 Comprobante creado: $COMPRA_ID"
```

### Paso 2: Ver las líneas automáticas generadas
```bash
echo "🔍 Analizando comprobante con líneas DSL automáticas:"

curl -s http://localhost:3000/api/v1/vouchers/$COMPRA_ID | jq -r '
  .data | 
  "📌 Comprobante: \(.number) - \(.description)",
  "💰 Total: Débitos: \(.total_debit) | Créditos: \(.total_credit)",
  "📊 Líneas (\(.voucher_lines | length) total):",
  (.voucher_lines[] | "   [\(.line_number)] \(.description): D:\(.debit_amount) C:\(.credit_amount)"),
  "",
  "🤖 Metadata DSL:",
  (.additional_data.custom_fields | to_entries[] | "   - \(.key): \(.value)")
'
```

### Paso 3: Intentar procesar (verificar workflows)
```bash
echo "⚡ Intentando procesar comprobante..."

RESULT=$(curl -s -X POST http://localhost:3000/api/v1/vouchers/$COMPRA_ID/post)
SUCCESS=$(echo $RESULT | jq -r '.success')

if [ "$SUCCESS" = "true" ]; then
    echo "✅ Comprobante procesado exitosamente"
    echo "📚 Asiento contable generado"
else
    echo "❌ Procesamiento bloqueado:"
    echo $RESULT | jq -r '.message'
    echo "🔄 Workflow requerido: $(echo $RESULT | jq -r '.message' | grep -o '[A-Z_]*APPROVAL')"
fi
```

### Paso 4: Simular aprobación y procesar
```bash
# En un caso real, aquí iría el proceso de aprobación
# Para la demo, modificaremos temporalmente el límite

echo "👤 Simulando aprobación del CFO..."
sleep 2

# Si el monto es menor al límite CEO, debería procesarse
# Crear un comprobante que sí se pueda procesar
VENTA_ID=$(curl -s -X POST http://localhost:3000/api/v1/vouchers \
  -H "Content-Type: application/json" \
  -d '{
    "voucher_type": "invoice_sale",
    "date": "2025-07-24T00:00:00Z",
    "description": "Venta procesable - Demo completa",
    "reference": "FV-PROCESS-001",
    "third_party_id": "TP001",
    "voucher_lines": [
      {
        "account_id": "02d0cc5b7214aa0a543fe2c5224c86d7",
        "description": "Servicios profesionales",
        "debit_amount": 0,
        "credit_amount": 5000000,
        "third_party_id": "TP001"
      },
      {
        "account_id": "a757c937d68d833683d72c91c679a962",
        "description": "Por cobrar cliente",
        "debit_amount": 5950000,
        "credit_amount": 0,
        "third_party_id": "TP001"
      }
    ]
  }' | jq -r '.data.id')

# Procesar
echo "⚡ Procesando venta..."
curl -s -X POST http://localhost:3000/api/v1/vouchers/$VENTA_ID/post | jq '.'

# Ver el asiento contable generado
echo "📚 Verificando asiento contable generado:"
curl -s http://localhost:3000/api/v1/journal-entries | \
  jq -r '.data.entries[0] | "Asiento: \(.number) - Estado: \(.status) - Líneas: \(.journal_lines | length)"'
```

---

## 📊 Demo 4: Dashboard y Reportes

### Paso 1: Abrir dashboard en navegador
```bash
echo "🌐 Abriendo dashboard..."
echo "URL: http://localhost:3000/dashboard.html"

# En macOS
open http://localhost:3000/dashboard.html

# O manualmente abrir en el navegador
```

### Paso 2: Generar datos para el dashboard
```bash
# Script para generar múltiples transacciones
cat > generar_datos_demo.sh << 'EOF'
#!/bin/bash

echo "🏭 Generando datos de demo..."

# Generar 10 ventas aleatorias
for i in {1..10}; do
    monto=$((RANDOM % 50000 + 10000))
    registrar_venta_pos "Venta demo $i" $monto
    sleep 0.5
done

# Generar algunos pagos
for i in {1..3}; do
    monto=$((RANDOM % 1000000 + 100000))
    curl -s -X POST http://localhost:3000/api/v1/vouchers \
      -H "Content-Type: application/json" \
      -d "{
        \"voucher_type\": \"payment\",
        \"date\": \"2025-07-24T00:00:00Z\",
        \"description\": \"Pago demo $i\",
        \"reference\": \"PAY-DEMO-$i\",
        \"voucher_lines\": [
          {
            \"account_id\": \"a757c937d68d833683d72c91c679a962\",
            \"description\": \"Pago cuenta\",
            \"debit_amount\": $monto,
            \"credit_amount\": 0
          },
          {
            \"account_id\": \"d1e05613ceab0efab7d3e0b6ad290345\",
            \"description\": \"Salida caja\",
            \"debit_amount\": 0,
            \"credit_amount\": $monto
          }
        ]
      }" > /dev/null
    echo "💰 Pago $i: \$$(printf "%'.0f" $monto)"
done

echo "✅ Datos generados"
EOF

chmod +x generar_datos_demo.sh
./generar_datos_demo.sh
```

### Paso 3: Ver estadísticas
```bash
# Resumen de comprobantes
echo "📊 Resumen de comprobantes:"
curl -s http://localhost:3000/api/v1/vouchers/stats | jq '.'

# KPIs del dashboard
echo "📈 KPIs actuales:"
curl -s http://localhost:3000/api/v1/dashboard/kpis | jq '.'
```

### Paso 4: Explorar funcionalidades web
1. **Dashboard**: Ver KPIs en tiempo real
2. **Lista de Comprobantes**: Filtrar, buscar, ordenar
3. **Crear Comprobante**: Usar el formulario web
4. **Plan de Cuentas**: Explorar el PUC colombiano
5. **Editor DSL**: Ver y editar plantillas

---

## 🔧 Troubleshooting

### Problema: El servidor no inicia
```bash
# Verificar si hay otro proceso usando el puerto
lsof -i :3000

# Matar procesos anteriores
pkill -f "go run main.go"

# Verificar logs
go run main.go 2>&1 | tee server.log
```

### Problema: Las reglas DSL no se aplican
```bash
# Verificar que el archivo DSL esté correcto
go build ./...

# Ver logs del servidor para errores DSL
grep "DSL" server.log

# Restaurar backup si es necesario
cp internal/services/dsl_rules_engine.go.bak internal/services/dsl_rules_engine.go
```

### Problema: Los comprobantes no se balancean
```bash
# Función helper para verificar balance
verificar_balance() {
    local voucher_id=$1
    curl -s http://localhost:3000/api/v1/vouchers/$voucher_id | \
      jq -r '.data | 
        "Débitos: \(.total_debit)",
        "Créditos: \(.total_credit)",
        "Diferencia: \(.total_debit - .total_credit)",
        if (.total_debit == .total_credit) then "✅ Balanceado" else "❌ Desbalanceado" end'
}

# Usar así:
verificar_balance "id-del-comprobante"
```

---

## 🎯 Conclusión de la Demo

### Lo que hemos demostrado:

1. **POS Simplificado**: 
   - Registro rápido de ventas
   - Cálculo automático de impuestos
   - Generación de comprobantes completos

2. **Reglas DSL Dinámicas**:
   - Cambio de tasas de impuesto en caliente
   - Modificación de límites de aprobación
   - Adición de nuevas reglas sin recompilar

3. **Workflows Inteligentes**:
   - Aprobaciones automáticas según montos
   - Validaciones específicas por tipo
   - Notificaciones y alertas

4. **Integración Completa**:
   - Dashboard en tiempo real
   - API REST completa
   - Interfaz web funcional

### Próximos pasos sugeridos:
1. Integrar con sistemas externos (ERP, CRM)
2. Agregar más tipos de comprobantes
3. Implementar firma digital
4. Añadir reportes fiscales automáticos

¡El Motor Contable con go-dsl está listo para casos de uso empresariales reales!