-- Seed data for Journal Templates
-- These are example templates using a simplified DSL syntax for the POC

-- Template 1: Nómina Mensual Simplificada
INSERT INTO journal_templates (id, name, description, dsl_code, parameters, is_active, organization_id, created_at, updated_at, created_by)
VALUES (
    'tpl-001',
    'Nómina Mensual',
    'Template para registro de nómina mensual con prestaciones básicas',
    'template payroll_monthly
  params ($total_salaries, $period)
  
  entry
    description: "Nómina mensual - " + $period
    date: last_day($period)
    reference: "NOM-" + $period
    
    # Salarios
    line debit account("510506") amount($total_salaries) 
         description("Sueldos y salarios")
    
    # Prestaciones sociales (30% aprox)
    line debit account("510527") amount($total_salaries * 0.0833)
         description("Cesantías")
    
    line debit account("510530") amount($total_salaries * 0.01)
         description("Intereses cesantías")
    
    # Contrapartidas
    line credit account("250505") amount($total_salaries * 0.8)
         description("Salarios por pagar")
    
    line credit account("237005") amount($total_salaries * 0.2)
         description("Retenciones por pagar")
    
    line credit account("261005") amount($total_salaries * 0.0933)
         description("Cesantías consolidadas")',
    '[{"name":"total_salaries","type":"number","required":true,"description":"Total de salarios base"},{"name":"period","type":"string","required":true,"description":"Período (YYYY-MM)"}]',
    1,
    'org123',
    datetime('now'),
    datetime('now'),
    'system'
);

-- Template 2: Depreciación Mensual
INSERT INTO journal_templates (id, name, description, dsl_code, parameters, is_active, organization_id, created_at, updated_at, created_by)
VALUES (
    'tpl-002',
    'Depreciación Mensual',
    'Cálculo y registro de depreciación mensual de activos fijos',
    'template depreciation_monthly
  params ($asset_value, $monthly_rate, $asset_description, $period)
  
  entry
    description: "Depreciación " + $asset_description + " - " + $period
    date: last_day($period)
    
    line debit account("516005") amount($asset_value * $monthly_rate)
         description("Gasto depreciación " + $asset_description)
    
    line credit account("159205") amount($asset_value * $monthly_rate)
         description("Depreciación acumulada " + $asset_description)',
    '[{"name":"asset_value","type":"number","required":true,"description":"Valor del activo"},{"name":"monthly_rate","type":"number","required":true,"description":"Tasa mensual (ej: 0.0083 para 10 años)"},{"name":"asset_description","type":"string","required":true,"description":"Descripción del activo"},{"name":"period","type":"string","required":true,"description":"Período (YYYY-MM)"}]',
    1,
    'org123',
    datetime('now'),
    datetime('now'),
    'system'
);

-- Template 3: Factura de Venta Recurrente
INSERT INTO journal_templates (id, name, description, dsl_code, parameters, is_active, organization_id, created_at, updated_at, created_by)
VALUES (
    'tpl-003',
    'Factura de Venta Recurrente',
    'Template para facturas de venta que se repiten mensualmente',
    'template recurring_invoice
  params ($customer_name, $service_amount, $invoice_date)
  
  entry
    description: "Factura venta - " + $customer_name
    date: $invoice_date
    
    # Cuenta por cobrar (total con IVA)
    line debit account("130505") amount($service_amount * 1.19)
         description("CxC " + $customer_name)
    
    # Ingreso por servicios
    line credit account("413535") amount($service_amount)
         description("Ingreso servicios profesionales")
    
    # IVA generado (19%)
    line credit account("240801") amount($service_amount * 0.19)
         description("IVA generado 19%")',
    '[{"name":"customer_name","type":"string","required":true,"description":"Nombre del cliente"},{"name":"service_amount","type":"number","required":true,"description":"Valor del servicio (sin IVA)"},{"name":"invoice_date","type":"date","required":true,"description":"Fecha de la factura"}]',
    1,
    'org123',
    datetime('now'),
    datetime('now'),
    'system'
);

-- Template 4: Cierre de Caja Diario
INSERT INTO journal_templates (id, name, description, dsl_code, parameters, is_active, organization_id, created_at, updated_at, created_by)
VALUES (
    'tpl-004',
    'Cierre de Caja Diario',
    'Registro del cierre de caja con ventas del día',
    'template daily_cash_close
  params ($cash_sales, $card_sales, $expenses, $date)
  
  entry
    description: "Cierre de caja del día"
    date: $date
    
    # Efectivo recibido
    line debit account("110505") amount($cash_sales - $expenses)
         description("Efectivo en caja")
    
    # Gastos menores pagados
    line debit account("519595") amount($expenses)
         description("Gastos menores de caja")
    
    # Ventas en tarjeta (pendiente de consignar)
    line debit account("110510") amount($card_sales)
         description("Ventas con tarjeta por consignar")
    
    # Total ventas del día
    line credit account("413501") amount($cash_sales + $card_sales)
         description("Ventas del día")',
    '[{"name":"cash_sales","type":"number","required":true,"description":"Ventas en efectivo"},{"name":"card_sales","type":"number","required":true,"description":"Ventas con tarjeta"},{"name":"expenses","type":"number","required":true,"description":"Gastos pagados de caja"},{"name":"date","type":"date","required":true,"description":"Fecha del cierre"}]',
    1,
    'org123',
    datetime('now'),
    datetime('now'),
    'system'
);

-- Template 5: Provisión de Servicios Públicos
INSERT INTO journal_templates (id, name, description, dsl_code, parameters, is_active, organization_id, created_at, updated_at, created_by)
VALUES (
    'tpl-005',
    'Provisión Servicios Públicos',
    'Provisión mensual de servicios públicos pendientes de facturar',
    'template utilities_provision
  params ($electricity, $water, $internet, $phone, $period)
  
  entry
    description: "Provisión servicios públicos - " + $period
    date: last_day($period)
    
    # Gastos de servicios
    line debit account("513530") amount($electricity)
         description("Energía eléctrica")
    
    line debit account("513535") amount($water)
         description("Acueducto y alcantarillado")
    
    line debit account("513540") amount($internet)
         description("Internet")
    
    line debit account("513525") amount($phone)
         description("Teléfono")
    
    # Provisión por pagar
    line credit account("233550") amount($electricity + $water + $internet + $phone)
         description("Servicios públicos por pagar")',
    '[{"name":"electricity","type":"number","required":true,"description":"Consumo estimado electricidad"},{"name":"water","type":"number","required":true,"description":"Consumo estimado agua"},{"name":"internet","type":"number","required":true,"description":"Costo mensual internet"},{"name":"phone","type":"number","required":true,"description":"Costo mensual teléfono"},{"name":"period","type":"string","required":true,"description":"Período (YYYY-MM)"}]',
    1,
    'org123',
    datetime('now'),
    datetime('now'),
    'system'
);

-- Template 6: Pago a Proveedores
INSERT INTO journal_templates (id, name, description, dsl_code, parameters, is_active, organization_id, created_at, updated_at, created_by)
VALUES (
    'tpl-006',
    'Pago a Proveedores',
    'Template para registro de pagos a proveedores',
    'template supplier_payment
  params ($supplier_name, $amount, $retention, $payment_date, $bank_account)
  
  entry
    description: "Pago a " + $supplier_name
    date: $payment_date
    
    # Cuenta por pagar
    line debit account("220505") amount($amount)
         description("Pago factura " + $supplier_name)
    
    # Retención en la fuente
    line debit account("236540") amount($retention)
         description("Retención aplicada")
    
    # Salida de banco
    line credit account($bank_account) amount($amount - $retention)
         description("Pago por transferencia")',
    '[{"name":"supplier_name","type":"string","required":true,"description":"Nombre del proveedor"},{"name":"amount","type":"number","required":true,"description":"Valor de la factura"},{"name":"retention","type":"number","required":true,"description":"Retención en la fuente"},{"name":"payment_date","type":"date","required":true,"description":"Fecha de pago"},{"name":"bank_account","type":"string","required":true,"description":"Cuenta bancaria (código)"}]',
    1,
    'org123',
    datetime('now'),
    datetime('now'),
    'system'
);

-- Execute the inserts
-- Note: SQLite will handle the JSON fields properly