#!/bin/bash

echo "=== Test de cambio de IVA en DSL ==="
echo

# Ver IVA actual
echo "1. IVA actual:"
curl -s http://localhost:3000/api/v1/dsl/iva-rate | jq
echo

# Cambiar IVA a 16%
echo "2. Cambiando IVA a 16%:"
curl -s -X POST http://localhost:3000/api/v1/dsl/iva-rate \
  -H "Content-Type: application/json" \
  -d '{"rate": 0.16}' | jq
echo

# Verificar cambio
echo "3. Verificando cambio:"
curl -s http://localhost:3000/api/v1/dsl/iva-rate | jq
echo

# Restaurar IVA a 19%
echo "4. Restaurando IVA a 19%:"
curl -s -X POST http://localhost:3000/api/v1/dsl/iva-rate \
  -H "Content-Type: application/json" \
  -d '{"rate": 0.19}' | jq
echo

echo "✅ Test completado"