-- Create templates table
CREATE TABLE IF NOT EXISTS templates (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    name TEXT NOT NULL,
    description TEXT,
    type TEXT NOT NULL CHECK(type IN ('invoice_sale', 'invoice_purchase', 'payroll', 'payment', 'receipt', 'adjustment')),
    dsl_content TEXT NOT NULL,
    parameters JSON,
    status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active', 'inactive')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT,
    company_id TEXT,
    UNIQUE(name, company_id)
);

-- Create index for performance
CREATE INDEX idx_templates_company_status ON templates(company_id, status);
CREATE INDEX idx_templates_type ON templates(type);

-- Create trigger for updated_at
CREATE TRIGGER update_templates_timestamp 
AFTER UPDATE ON templates
BEGIN
    UPDATE templates SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Insert sample templates
INSERT INTO templates (id, name, description, type, dsl_content, parameters, status, company_id) VALUES
-- Template 1: Factura de Venta Estándar
('tpl_001', 'Factura de Venta Estándar', 'Genera asiento contable para facturas de venta con IVA 19%', 'invoice_sale', 
'# Template: Factura de Venta Estándar
# Genera asientos contables para facturas de venta con IVA 19%

template factura_venta_standard
  params (invoice_number, customer_name, base_amount, invoice_date)
  
  # Calcular valores
  set tax_rate = 0.19
  set tax_amount = base_amount * tax_rate
  set total_amount = base_amount + tax_amount
  
  entry
    description: "Factura Venta " + invoice_number + " - " + customer_name
    date: invoice_date
    reference: invoice_number
    
    # Débito a cuentas por cobrar
    line debit account("130505") amount(total_amount)
         description("CxC Cliente " + customer_name)
    
    # Crédito a ventas
    line credit account("413595") amount(base_amount)
         description("Venta de productos")
    
    # Crédito a IVA por pagar
    line credit account("240802") amount(tax_amount)
         description("IVA generado 19%")',
'[{"name":"invoice_number","type":"string","description":"Número de factura"},{"name":"customer_name","type":"string","description":"Nombre del cliente"},{"name":"base_amount","type":"number","description":"Monto base sin IVA"},{"name":"invoice_date","type":"date","description":"Fecha de la factura"}]',
'active', 'DEMO-CO'),

-- Template 2: Factura de Venta con Retención
('tpl_002', 'Factura de Venta con Retención', 'Genera asiento para facturas con retención en la fuente', 'invoice_sale',
'# Template: Factura de Venta con Retención
# Aplica retención en la fuente del 2.5%

template factura_venta_retencion
  params (invoice_number, customer_name, base_amount, invoice_date)
  
  # Calcular valores
  set tax_rate = 0.19
  set retention_rate = 0.025
  set tax_amount = base_amount * tax_rate
  set retention_amount = base_amount * retention_rate
  set total_amount = base_amount + tax_amount
  set net_receivable = total_amount - retention_amount
  
  entry
    description: "Factura Venta " + invoice_number + " - " + customer_name + " (con retención)"
    date: invoice_date
    reference: invoice_number
    
    # Débito a cuentas por cobrar (neto)
    line debit account("130505") amount(net_receivable)
         description("CxC Cliente " + customer_name + " (neto)")
    
    # Débito a retención por cobrar
    line debit account("135515") amount(retention_amount)
         description("Retención en la fuente por cobrar")
    
    # Crédito a ventas
    line credit account("413595") amount(base_amount)
         description("Venta de productos")
    
    # Crédito a IVA por pagar
    line credit account("240802") amount(tax_amount)
         description("IVA generado 19%")',
'[{"name":"invoice_number","type":"string","description":"Número de factura"},{"name":"customer_name","type":"string","description":"Nombre del cliente"},{"name":"base_amount","type":"number","description":"Monto base sin IVA"},{"name":"invoice_date","type":"date","description":"Fecha de la factura"}]',
'active', 'DEMO-CO'),

-- Template 3: Nómina Mensual Básica
('tpl_003', 'Nómina Mensual Básica', 'Genera asiento contable para pago de nómina mensual', 'payroll',
'# Template: Nómina Mensual Básica
# Procesa el pago de nómina con deducciones estándar

template nomina_mensual
  params (employee_name, basic_salary, period)
  
  # Calcular deducciones
  set health_deduction = basic_salary * 0.04      # 4% salud
  set pension_deduction = basic_salary * 0.04     # 4% pensión
  set total_deductions = health_deduction + pension_deduction
  set net_payment = basic_salary - total_deductions
  
  # Aportes patronales
  set employer_health = basic_salary * 0.085      # 8.5% salud patronal
  set employer_pension = basic_salary * 0.12      # 12% pensión patronal
  set arl = basic_salary * 0.00522                # 0.522% ARL
  
  entry
    description: "Nómina " + period + " - " + employee_name
    date: last_day(period)
    reference: "NOM-" + period
    
    # Gastos de nómina
    line debit account("510506") amount(basic_salary)
         description("Salario básico")
    
    line debit account("510568") amount(employer_health)
         description("Aporte patronal salud")
    
    line debit account("510569") amount(employer_pension)
         description("Aporte patronal pensión")
    
    line debit account("510570") amount(arl)
         description("ARL")
    
    # Deducciones empleado
    line credit account("237005") amount(health_deduction)
         description("Salud empleado por pagar")
    
    line credit account("238030") amount(pension_deduction)
         description("Pensión empleado por pagar")
    
    # Aportes patronales por pagar
    line credit account("237006") amount(employer_health)
         description("Salud patronal por pagar")
    
    line credit account("238031") amount(employer_pension)
         description("Pensión patronal por pagar")
    
    line credit account("237010") amount(arl)
         description("ARL por pagar")
    
    # Neto a pagar
    line credit account("250505") amount(net_payment)
         description("Salario neto por pagar")',
'[{"name":"employee_name","type":"string","description":"Nombre del empleado"},{"name":"basic_salary","type":"number","description":"Salario básico mensual"},{"name":"period","type":"string","description":"Período (YYYY-MM)"}]',
'active', 'DEMO-CO'),

-- Template 4: Compra con IVA
('tpl_004', 'Compra con IVA', 'Registra compras de inventario o gastos con IVA', 'invoice_purchase',
'# Template: Compra con IVA
# Registra compras aplicando IVA descontable

template compra_con_iva
  params (invoice_number, supplier_name, base_amount, purchase_type, purchase_date)
  
  # Calcular IVA
  set tax_rate = 0.19
  set tax_amount = base_amount * tax_rate
  set total_amount = base_amount + tax_amount
  
  # Determinar cuenta según tipo
  if purchase_type == "inventory"
    set expense_account = "143505"  # Inventario
    set expense_desc = "Compra de inventario"
  else
    set expense_account = "519595"  # Gastos diversos
    set expense_desc = "Gastos generales"
  
  entry
    description: "Compra FC " + invoice_number + " - " + supplier_name
    date: purchase_date
    reference: invoice_number
    
    # Débito al gasto o inventario
    line debit account(expense_account) amount(base_amount)
         description(expense_desc)
    
    # Débito IVA descontable
    line debit account("240801") amount(tax_amount)
         description("IVA descontable 19%")
    
    # Crédito a proveedores
    line credit account("220505") amount(total_amount)
         description("CxP " + supplier_name)',
'[{"name":"invoice_number","type":"string","description":"Número de factura"},{"name":"supplier_name","type":"string","description":"Nombre del proveedor"},{"name":"base_amount","type":"number","description":"Monto base sin IVA"},{"name":"purchase_type","type":"string","description":"Tipo: inventory o expense"},{"name":"purchase_date","type":"date","description":"Fecha de compra"}]',
'active', 'DEMO-CO'),

-- Template 5: Pago a Proveedor
('tpl_005', 'Pago a Proveedor', 'Registra el pago a proveedores desde banco', 'payment',
'# Template: Pago a Proveedor
# Registra pagos realizados a proveedores

template pago_proveedor
  params (payment_number, supplier_name, payment_amount, payment_date, bank_account)
  
  entry
    description: "Pago a " + supplier_name + " - Comp. " + payment_number
    date: payment_date
    reference: payment_number
    
    # Débito a proveedores (reduce la deuda)
    line debit account("220505") amount(payment_amount)
         description("Pago CxP " + supplier_name)
    
    # Crédito al banco
    line credit account(bank_account) amount(payment_amount)
         description("Egreso bancario")',
'[{"name":"payment_number","type":"string","description":"Número de comprobante de pago"},{"name":"supplier_name","type":"string","description":"Nombre del proveedor"},{"name":"payment_amount","type":"number","description":"Monto del pago"},{"name":"payment_date","type":"date","description":"Fecha del pago"},{"name":"bank_account","type":"string","description":"Cuenta bancaria (ej: 111005)"}]',
'active', 'DEMO-CO');

-- Create template execution history table
CREATE TABLE IF NOT EXISTS template_executions (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    template_id TEXT NOT NULL,
    parameters JSON NOT NULL,
    result_entry_id TEXT,
    status TEXT NOT NULL CHECK(status IN ('success', 'error')),
    error_message TEXT,
    executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    executed_by TEXT,
    FOREIGN KEY (template_id) REFERENCES templates(id)
);

CREATE INDEX idx_template_executions_template ON template_executions(template_id);
CREATE INDEX idx_template_executions_date ON template_executions(executed_at);