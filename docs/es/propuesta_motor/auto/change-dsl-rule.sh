#!/bin/bash

# Script para cambiar la regla DSL de IVA en la base de datos

DB_PATH="../app/db_contable.db"

echo "🔧 Cambiando regla DSL de IVA..."

# Actualizar la regla de IVA al 15%
sqlite3 "$DB_PATH" <<EOF
UPDATE dsl_templates 
SET content = '// Regla DSL para cálculo de IVA
rule automatic_tax_generation {
    when {
        voucher.type == "invoice_sale"
        account.code.startsWith("4")
    }
    then {
        taxRate = 0.15  // IVA reducido al 15%
        taxAmount = baseAmount * taxRate
        addLine({
            account: "240802",
            description: "IVA 15% generado por DSL",
            credit: taxAmount,
            metadata: {
                generated_by: "dsl_rules_engine",
                rule: "automatic_tax_generation",
                rate: "15%"
            }
        })
    }
}'
WHERE id = 'tpl-tax-001';
EOF

echo "✅ Regla DSL actualizada: IVA ahora es 15%"
echo ""
echo "Para restaurar al 19%, ejecuta:"
echo "  ./restore-dsl-rule.sh"