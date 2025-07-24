-- DSL Rules Templates for Motor Contable POC
-- These templates demonstrate the power of go-dsl integration

-- Clear existing templates
DELETE FROM dsl_templates;

-- Template for Invoice Sales with automatic tax calculation
INSERT INTO dsl_templates (
    id, organization_id, name, description, category, dsl_code, 
    status, created_by_user_id, version, metadata, variables, created_at, updated_at
) VALUES 
(
    'tpl-invoice-sale-001',
    'org-001',
    'Factura de Venta con IVA Automático',
    'Genera automáticamente líneas de IVA 19% para facturas de venta',
    'voucher_rules',
    '# Reglas para Factura de Venta
rule validate_customer 
  when voucher.type == "invoice_sale" 
  then {
    validate "has_customer"
    validate "has_valid_date"
  }

rule calculate_tax_19
  when voucher.type == "invoice_sale" AND voucher.total > 0
  then {
    calculate tax(19%)
    add_line("IVA 19% sobre ventas", "240802", credit)
  }

rule notification_high_value
  when voucher.total > 50000000
  then {
    notify "Factura de alto valor generada: ${voucher.number}"
  }',
    'ACTIVE',
    'system',
    '1.0',
    '{"automated": true, "tax_rate": 19}',
    '{"customer_id": "string", "base_amount": "number"}',
    datetime('now'),
    datetime('now')
),

-- Template for Purchase Invoices with retention
(
    'tpl-invoice-purchase-001',
    'org-001',
    'Factura de Compra con Retención',
    'Aplica retención en la fuente automática para compras',
    'voucher_rules',
    '# Reglas para Factura de Compra
rule validate_supplier
  when voucher.type == "invoice_purchase"
  then {
    validate "has_supplier"
    validate "has_invoice_number"
  }

rule calculate_tax_deductible
  when voucher.type == "invoice_purchase"
  then {
    calculate tax(19%)
    add_line("IVA descontable en compras", "240805", debit)
  }

rule apply_retention
  when voucher.type == "invoice_purchase" AND voucher.total > 1000000
  then {
    calculate retention(2.5%)
    add_line("Retención en la fuente 2.5%", "236540", credit)
  }

rule cost_center_assignment
  when account.code STARTS_WITH "5"
  then {
    set cost_center "CC-ADMIN-001"
  }',
    'ACTIVE',
    'system',
    '1.0',
    '{"automated": true, "retention_rate": 2.5}',
    '{"supplier_id": "string", "invoice_number": "string"}',
    datetime('now'),
    datetime('now')
),

-- Template for Payments with approval workflow
(
    'tpl-payment-001',
    'org-001',
    'Comprobante de Egreso con Workflow',
    'Aplica reglas de aprobación para pagos según monto',
    'voucher_rules',
    '# Reglas para Comprobantes de Egreso
rule validate_payment
  when voucher.type == "payment"
  then {
    validate "has_bank_account"
    validate "has_beneficiary"
  }

rule small_payment_auto_approve
  when voucher.type == "payment" AND voucher.total <= 1000000
  then {
    set approval_status "auto_approved"
  }

rule medium_payment_approval
  when voucher.type == "payment" AND voucher.total > 1000000 AND voucher.total <= 5000000
  then {
    workflow "single_approval"
    notify "Pago requiere aprobación del supervisor"
  }

rule high_payment_approval
  when voucher.type == "payment" AND voucher.total > 5000000
  then {
    workflow "dual_approval"
    notify "Pago de alto valor requiere doble aprobación"
  }

rule cash_movement_alert
  when account.code == "110505" AND line.credit > 0
  then {
    notify "Movimiento de caja detectado: ${line.amount}"
    set requires_cash_count true
  }',
    'ACTIVE',
    'system',
    '1.0',
    '{"workflows": ["single_approval", "dual_approval"]}',
    '{"beneficiary_id": "string", "payment_method": "string"}',
    datetime('now'),
    datetime('now')
),

-- Template for Payroll Entry
(
    'tpl-payroll-001',
    'org-001',
    'Nómina Mensual Automatizada',
    'Genera asientos de nómina con cálculos automáticos',
    'journal_template',
    '# Plantilla de Nómina
template payroll_entry
  params (period, employees[])
  
  for each employee in employees {
    # Salario base
    add_line(
      account: "510506",
      description: "Salario ${employee.name}",
      debit: employee.base_salary,
      cost_center: employee.cost_center
    )
    
    # Cálculo automático de prestaciones
    calculate benefits(employee.base_salary) {
      cesantias: base_salary * 0.0833
      intereses: cesantias * 0.12
      prima: base_salary * 0.0833
      vacaciones: base_salary * 0.0417
    }
    
    # Líneas de prestaciones
    add_line("510527", "Cesantías", debit: benefits.cesantias)
    add_line("510530", "Intereses cesantías", debit: benefits.intereses)
    add_line("510533", "Prima", debit: benefits.prima)
    add_line("510536", "Vacaciones", debit: benefits.vacaciones)
    
    # Seguridad social empresa
    calculate social_security(employee.base_salary) {
      salud_empresa: base_salary * 0.085
      pension_empresa: base_salary * 0.12
      arl: base_salary * 0.00522
    }
    
    add_line("510569", "Aporte salud empresa", debit: social_security.salud_empresa)
    add_line("510570", "Aporte pensión empresa", debit: social_security.pension_empresa)
    add_line("510568", "ARL", debit: social_security.arl)
    
    # Deducciones empleado
    calculate deductions(employee.base_salary) {
      salud_empleado: base_salary * 0.04
      pension_empleado: base_salary * 0.04
    }
    
    # Cuenta por pagar al empleado
    net_pay = employee.base_salary - deductions.salud_empleado - deductions.pension_empleado
    add_line("250505", "Por pagar ${employee.name}", credit: net_pay)
    
    # Cuentas por pagar aportes
    add_line("237005", "Salud por pagar", credit: social_security.salud_empresa + deductions.salud_empleado)
    add_line("237006", "Pensión por pagar", credit: social_security.pension_empresa + deductions.pension_empleado)
    add_line("237010", "ARL por pagar", credit: social_security.arl)
  }
  
  # Validación final
  validate balance
  
  # Notificaciones
  notify "Nómina procesada: ${employees.count} empleados"
  
  if total > 100000000 {
    workflow "payroll_approval"
  }',
    'ACTIVE',
    'system',
    '1.0',
    '{"category": "payroll", "frequency": "monthly"}',
    '{"period": "string", "employees": "array"}',
    datetime('now'),
    datetime('now')
),

-- Template for Automatic Depreciation
(
    'tpl-depreciation-001',
    'org-001',
    'Depreciación Mensual Automática',
    'Calcula y registra depreciación de activos fijos',
    'journal_template',
    '# Plantilla de Depreciación
template monthly_depreciation
  params (period, assets[])
  
  for each asset in assets {
    if asset.depreciation_method == "straight_line" {
      monthly_amount = asset.value / asset.useful_life_months
      
      add_line(
        account: get_expense_account(asset.category),
        description: "Depreciación ${asset.name}",
        debit: monthly_amount,
        cost_center: asset.cost_center
      )
      
      add_line(
        account: get_accumulated_account(asset.category),
        description: "Depreciación acumulada ${asset.name}",
        credit: monthly_amount
      )
    }
  }
  
  # Funciones auxiliares
  function get_expense_account(category) {
    switch category {
      case "building": return "516005"
      case "machinery": return "516010"
      case "vehicle": return "516015"
      case "computer": return "516020"
      default: return "516095"
    }
  }
  
  function get_accumulated_account(category) {
    switch category {
      case "building": return "159205"
      case "machinery": return "159210"
      case "vehicle": return "159215"
      case "computer": return "159220"
      default: return "159295"
    }
  }
  
  validate balance
  set automated true
  set reversible false',
    'ACTIVE',
    'system',
    '1.0',
    '{"category": "depreciation", "automated": true}',
    '{"period": "string", "assets": "array"}',
    datetime('now'),
    datetime('now')
),

-- Template for Tax Provision
(
    'tpl-tax-provision-001',
    'org-001',
    'Provisión de Impuestos',
    'Calcula provisiones de impuestos basado en ventas',
    'journal_template',
    '# Provisión mensual de impuestos
template tax_provision
  params (period, sales_total, purchases_total)
  
  # Calcular IVA neto
  calculate iva_computation {
    iva_generated: sales_total * 0.19
    iva_deductible: purchases_total * 0.19
    iva_to_pay: iva_generated - iva_deductible
  }
  
  if iva_computation.iva_to_pay > 0 {
    add_line("240802", "IVA generado del período", debit: iva_computation.iva_generated)
    add_line("240805", "IVA descontable del período", credit: iva_computation.iva_deductible)
    add_line("240810", "IVA por pagar", credit: iva_computation.iva_to_pay)
  }
  
  # Calcular retención en la fuente
  calculate retention {
    sales_retention: sales_total * 0.025
    purchases_retention: purchases_total * 0.025
  }
  
  add_line("135515", "Retención practicada", debit: retention.purchases_retention)
  add_line("236540", "Retención por pagar", credit: retention.sales_retention)
  
  # Provisión renta
  calculate income_tax {
    net_income: (sales_total - purchases_total) * 0.8  # Simplified
    tax_rate: 0.35
    provision: net_income * tax_rate / 12  # Monthly provision
  }
  
  add_line("540505", "Provisión impuesto de renta", debit: income_tax.provision)
  add_line("261505", "Provisión impuesto de renta", credit: income_tax.provision)
  
  validate balance
  notify "Provisión de impuestos calculada para ${period}"',
    'ACTIVE',
    'system',
    '1.0',
    '{"category": "tax_provision", "automated": true}',
    '{"period": "string", "sales_total": "number", "purchases_total": "number"}',
    datetime('now'),
    datetime('now')
);

-- Create DSL validation rules for different scenarios
INSERT INTO dsl_templates (
    id, organization_id, name, description, category, dsl_code, 
    status, created_by_user_id, version, metadata, variables, created_at, updated_at
) VALUES 
(
    'val-general-001',
    'org-001',
    'Validaciones Generales Contables',
    'Reglas de validación aplicables a todos los comprobantes',
    'voucher_rules',
    '# Validaciones generales
rule period_open
  when voucher.date < period.start_date OR voucher.date > period.end_date
  then {
    error "El período contable está cerrado para esta fecha"
  }

rule duplicate_reference
  when exists(voucher.reference) AND count(vouchers.reference) > 1
  then {
    error "Ya existe un comprobante con esta referencia"
  }

rule inactive_account
  when line.account.is_active == false
  then {
    error "La cuenta ${line.account.code} no está activa"
  }

rule third_party_required
  when line.account.requires_third_party == true AND line.third_party == null
  then {
    error "La cuenta ${line.account.code} requiere tercero"
  }

rule cost_center_required
  when line.account.requires_cost_center == true AND line.cost_center == null
  then {
    error "La cuenta ${line.account.code} requiere centro de costo"
  }

rule minimum_amount
  when line.debit < 0 OR line.credit < 0
  then {
    error "Los montos no pueden ser negativos"
  }

rule balanced_entry
  when sum(lines.debit) != sum(lines.credit)
  then {
    error "El comprobante no está balanceado"
  }',
    'ACTIVE',
    'system',
    '1.0',
    '{"category": "general_validations"}',
    '{}',
    datetime('now'),
    datetime('now')
);

-- Sample DSL template for complex business rules
INSERT INTO dsl_templates (
    id, organization_id, name, description, category, dsl_code, 
    status, created_by_user_id, version, metadata, variables, created_at, updated_at
) VALUES 
(
    'rule-inventory-001',
    'org-001',
    'Control de Inventario',
    'Reglas para manejo automático de inventario y costos',
    'voucher_rules',
    '# Control de Inventario y Costos
rule inventory_purchase
  when voucher.type == "purchase" AND line.account.code STARTS_WITH "14"
  then {
    # Calcular costo promedio
    calculate average_cost {
      current_qty: inventory.quantity
      current_cost: inventory.total_cost
      new_qty: line.quantity
      new_cost: line.debit
      avg_cost: (current_cost + new_cost) / (current_qty + new_qty)
    }
    
    set inventory.average_cost avg_cost
    set inventory.quantity current_qty + new_qty
    
    # Generar movimiento de inventario
    create inventory_movement {
      type: "IN"
      quantity: new_qty
      cost: avg_cost
      reference: voucher.number
    }
  }

rule inventory_sale
  when voucher.type == "sale" AND line.account.code STARTS_WITH "41"
  then {
    # Buscar producto relacionado
    product = find_product(line.product_id)
    
    if inventory.quantity < line.quantity {
      error "Stock insuficiente para ${product.name}"
    }
    
    # Calcular costo de venta
    cost_amount = line.quantity * inventory.average_cost
    
    # Generar línea de costo automáticamente
    add_line("613505", "Costo de venta ${product.name}", debit: cost_amount)
    add_line("143505", "Inventario ${product.name}", credit: cost_amount)
    
    # Actualizar inventario
    set inventory.quantity inventory.quantity - line.quantity
    
    # Generar movimiento
    create inventory_movement {
      type: "OUT"
      quantity: line.quantity
      cost: inventory.average_cost
      reference: voucher.number
    }
  }',
    'ACTIVE',
    'system',
    '1.0',
    '{"category": "inventory_control"}',
    '{"product_id": "string", "quantity": "number"}',
    datetime('now'),
    datetime('now')
);

-- Indices ya existen en el esquema, no es necesario crearlos nuevamente