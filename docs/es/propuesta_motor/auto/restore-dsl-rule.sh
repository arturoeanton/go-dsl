#!/bin/bash

# Script para restaurar la regla DSL de IVA al 19%

DB_PATH="../app/db_contable.db"

echo "🔧 Restaurando regla DSL de IVA al 19%..."

# Restaurar la regla de IVA al 19%
sqlite3 "$DB_PATH" <<EOF
UPDATE dsl_templates 
SET content = '// Regla DSL para cálculo de IVA
rule automatic_tax_generation {
    when {
        voucher.type == "invoice_sale"
        account.code.startsWith("4")
    }
    then {
        taxRate = 0.19  // IVA estándar 19%
        taxAmount = baseAmount * taxRate
        addLine({
            account: "240802",
            description: "IVA 19% generado por DSL",
            credit: taxAmount,
            metadata: {
                generated_by: "dsl_rules_engine",
                rule: "automatic_tax_generation",
                rate: "19%"
            }
        })
    }
}'
WHERE id = 'tpl-tax-001';
EOF

echo "✅ Regla DSL restaurada: IVA es 19%"