-- ==============================================
-- DATOS INICIALES - MOTOR CONTABLE CLOUD-NATIVE
-- ==============================================
-- Versión: 1.0
-- Última actualización: 2024-01-15
-- Propósito: Datos de prueba para desarrollo
-- ==============================================

-- Usar esquema
SET search_path TO accounting_engine, public;

-- ==============================================
-- LIMPIAR DATOS EXISTENTES (SOLO DESARROLLO)
-- ==============================================

TRUNCATE TABLE audit_log CASCADE;
TRUNCATE TABLE journal_lines CASCADE;
TRUNCATE TABLE journal_entries CASCADE;
TRUNCATE TABLE vouchers CASCADE;
TRUNCATE TABLE accounting_templates CASCADE;
TRUNCATE TABLE fiscal_rules CASCADE;
TRUNCATE TABLE accounting_books CASCADE;
TRUNCATE TABLE projects CASCADE;
TRUNCATE TABLE cost_centers CASCADE;
TRUNCATE TABLE chart_of_accounts CASCADE;
TRUNCATE TABLE users CASCADE;
TRUNCATE TABLE organizations CASCADE;

-- ==============================================
-- ORGANIZACIONES DE PRUEBA
-- ==============================================

INSERT INTO organizations (id, code, name, country_code, tax_id, settings) VALUES
('11111111-1111-1111-1111-111111111111', 'DEMO-CO', 'Empresa Demo Colombia S.A.S.', 'CO', '900123456-7', 
 '{
   "currency": "COP",
   "fiscal_year": "calendar",
   "decimal_places": 2,
   "timezone": "America/Bogota",
   "tax_regime": "common",
   "industry": "technology"
 }'),
('22222222-2222-2222-2222-222222222222', 'DEMO-MX', 'Empresa Demo México S.A. de C.V.', 'MX', 'EDM850101ABC', 
 '{
   "currency": "MXN",
   "fiscal_year": "calendar",
   "decimal_places": 2,
   "timezone": "America/Mexico_City",
   "tax_regime": "general",
   "industry": "retail"
 }'),
('33333333-3333-3333-3333-333333333333', 'DEMO-CL', 'Empresa Demo Chile SpA', 'CL', '76.123.456-7', 
 '{
   "currency": "CLP",
   "fiscal_year": "calendar",
   "decimal_places": 0,
   "timezone": "America/Santiago",
   "tax_regime": "simplified",
   "industry": "services"
 }');

-- ==============================================
-- USUARIOS DE PRUEBA
-- ==============================================

INSERT INTO users (id, email, name, password_hash, role, organization_id) VALUES
-- Super Admin (sin organización específica)
('550e8400-e29b-41d4-a716-446655440001', 'admin@motorcontable.com', 'Administrador Sistema', 
 '$2a$10$YourHashedPasswordHere', 'SUPER_ADMIN', NULL),

-- Admins por organización
('550e8400-e29b-41d4-a716-446655440002', 'admin.co@demo.com', 'Admin Colombia', 
 '$2a$10$YourHashedPasswordHere', 'ORG_ADMIN', '11111111-1111-1111-1111-111111111111'),
('550e8400-e29b-41d4-a716-446655440003', 'admin.mx@demo.com', 'Admin México', 
 '$2a$10$YourHashedPasswordHere', 'ORG_ADMIN', '22222222-2222-2222-2222-222222222222'),

-- Contadores
('550e8400-e29b-41d4-a716-446655440004', 'contador@demo.com', 'Juan Pérez', 
 '$2a$10$YourHashedPasswordHere', 'ACCOUNTANT', '11111111-1111-1111-1111-111111111111'),

-- Auxiliares
('550e8400-e29b-41d4-a716-446655440005', 'auxiliar@demo.com', 'María García', 
 '$2a$10$YourHashedPasswordHere', 'CLERK', '11111111-1111-1111-1111-111111111111');

-- ==============================================
-- CENTROS DE COSTO Y PROYECTOS
-- ==============================================

INSERT INTO cost_centers (id, organization_id, code, name, parent_id) VALUES
('660e8400-e29b-41d4-a716-446655440001', '11111111-1111-1111-1111-111111111111', 'ADM', 'Administración', NULL),
('660e8400-e29b-41d4-a716-446655440002', '11111111-1111-1111-1111-111111111111', 'VEN', 'Ventas', NULL),
('660e8400-e29b-41d4-a716-446655440003', '11111111-1111-1111-1111-111111111111', 'OPE', 'Operaciones', NULL);

INSERT INTO projects (id, organization_id, code, name, start_date, end_date, budget) VALUES
('770e8400-e29b-41d4-a716-446655440001', '11111111-1111-1111-1111-111111111111', 
 'PROY-001', 'Implementación ERP', '2024-01-01', '2024-12-31', 50000000),
('770e8400-e29b-41d4-a716-446655440002', '11111111-1111-1111-1111-111111111111', 
 'PROY-002', 'Desarrollo Web', '2024-02-01', '2024-06-30', 25000000);

-- ==============================================
-- FUNCIÓN AUXILIAR PARA CREAR CUENTAS
-- ==============================================

CREATE OR REPLACE FUNCTION create_account(
    p_org_id UUID,
    p_code VARCHAR,
    p_name VARCHAR,
    p_type VARCHAR,
    p_nature CHAR,
    p_parent_code VARCHAR DEFAULT NULL
) RETURNS UUID AS $$
DECLARE
    v_parent_id UUID;
    v_level INTEGER;
    v_account_id UUID;
BEGIN
    -- Determinar el nivel basado en el código
    v_level := LENGTH(REGEXP_REPLACE(p_code, '[^.]', '', 'g')) + 1;
    
    -- Buscar cuenta padre si existe
    IF p_parent_code IS NOT NULL THEN
        SELECT id INTO v_parent_id 
        FROM chart_of_accounts 
        WHERE organization_id = p_org_id AND account_code = p_parent_code;
    END IF;
    
    -- Insertar la cuenta
    INSERT INTO chart_of_accounts (
        organization_id, account_code, name, type, nature, 
        level, parent_id, is_detail, metadata
    ) VALUES (
        p_org_id, p_code, p_name, p_type, p_nature,
        v_level, v_parent_id, 
        v_level >= 3, -- Cuentas de nivel 3+ son de detalle
        '{"created_by": "initial_load"}'
    ) RETURNING id INTO v_account_id;
    
    RETURN v_account_id;
END;
$$ LANGUAGE plpgsql;

-- ==============================================
-- CATÁLOGO DE CUENTAS - COLOMBIA (NIIF PYMES)
-- ==============================================

DO $$
DECLARE
    v_org_id UUID := '11111111-1111-1111-1111-111111111111';
BEGIN
    -- CLASE 1: ACTIVOS
    PERFORM create_account(v_org_id, '1', 'ACTIVOS', 'ASSET', 'D');
    
    -- Grupo 11: Disponible
    PERFORM create_account(v_org_id, '11', 'DISPONIBLE', 'ASSET', 'D', '1');
    PERFORM create_account(v_org_id, '1105', 'CAJA', 'ASSET', 'D', '11');
    PERFORM create_account(v_org_id, '1105.01', 'Caja General', 'ASSET', 'D', '1105');
    PERFORM create_account(v_org_id, '1105.02', 'Caja Menor', 'ASSET', 'D', '1105');
    PERFORM create_account(v_org_id, '1110', 'BANCOS', 'ASSET', 'D', '11');
    PERFORM create_account(v_org_id, '1110.01', 'Banco Nacional - Cta Corriente', 'ASSET', 'D', '1110');
    PERFORM create_account(v_org_id, '1110.02', 'Banco Nacional - Cta Ahorros', 'ASSET', 'D', '1110');
    
    -- Grupo 13: Deudores
    PERFORM create_account(v_org_id, '13', 'DEUDORES', 'ASSET', 'D', '1');
    PERFORM create_account(v_org_id, '1305', 'CLIENTES', 'ASSET', 'D', '13');
    PERFORM create_account(v_org_id, '1305.01', 'Clientes Nacionales', 'ASSET', 'D', '1305');
    PERFORM create_account(v_org_id, '1305.05', 'Clientes del Exterior', 'ASSET', 'D', '1305');
    PERFORM create_account(v_org_id, '1355', 'ANTICIPO DE IMPUESTOS', 'ASSET', 'D', '13');
    PERFORM create_account(v_org_id, '1355.25', 'Retención en la Fuente', 'ASSET', 'D', '1355');
    
    -- Grupo 14: Inventarios
    PERFORM create_account(v_org_id, '14', 'INVENTARIOS', 'ASSET', 'D', '1');
    PERFORM create_account(v_org_id, '1435', 'MERCANCÍAS NO FABRICADAS', 'ASSET', 'D', '14');
    PERFORM create_account(v_org_id, '1435.01', 'Mercancías en Almacén', 'ASSET', 'D', '1435');
    
    -- CLASE 2: PASIVOS
    PERFORM create_account(v_org_id, '2', 'PASIVOS', 'LIABILITY', 'C');
    
    -- Grupo 22: Proveedores
    PERFORM create_account(v_org_id, '22', 'PROVEEDORES', 'LIABILITY', 'C', '2');
    PERFORM create_account(v_org_id, '2205', 'PROVEEDORES NACIONALES', 'LIABILITY', 'C', '22');
    PERFORM create_account(v_org_id, '2205.01', 'Proveedores Varios', 'LIABILITY', 'C', '2205');
    
    -- Grupo 24: Impuestos
    PERFORM create_account(v_org_id, '24', 'IMPUESTOS GRAVÁMENES Y TASAS', 'LIABILITY', 'C', '2');
    PERFORM create_account(v_org_id, '2408', 'IVA POR PAGAR', 'LIABILITY', 'C', '24');
    PERFORM create_account(v_org_id, '2408.01', 'IVA 19% por Pagar', 'LIABILITY', 'C', '2408');
    
    -- CLASE 3: PATRIMONIO
    PERFORM create_account(v_org_id, '3', 'PATRIMONIO', 'EQUITY', 'C');
    PERFORM create_account(v_org_id, '31', 'CAPITAL SOCIAL', 'EQUITY', 'C', '3');
    PERFORM create_account(v_org_id, '3105', 'CAPITAL SUSCRITO Y PAGADO', 'EQUITY', 'C', '31');
    PERFORM create_account(v_org_id, '3105.01', 'Capital Social', 'EQUITY', 'C', '3105');
    
    -- CLASE 4: INGRESOS
    PERFORM create_account(v_org_id, '4', 'INGRESOS', 'INCOME', 'C');
    PERFORM create_account(v_org_id, '41', 'OPERACIONALES', 'INCOME', 'C', '4');
    PERFORM create_account(v_org_id, '4135', 'COMERCIO AL POR MAYOR Y MENOR', 'INCOME', 'C', '41');
    PERFORM create_account(v_org_id, '4135.01', 'Venta de Mercancías', 'INCOME', 'C', '4135');
    
    -- CLASE 5: GASTOS
    PERFORM create_account(v_org_id, '5', 'GASTOS', 'EXPENSE', 'D');
    PERFORM create_account(v_org_id, '51', 'OPERACIONALES DE ADMINISTRACIÓN', 'EXPENSE', 'D', '5');
    PERFORM create_account(v_org_id, '5135', 'SERVICIOS', 'EXPENSE', 'D', '51');
    PERFORM create_account(v_org_id, '5135.01', 'Servicios Públicos', 'EXPENSE', 'D', '5135');
    
    -- CLASE 6: COSTOS
    PERFORM create_account(v_org_id, '6', 'COSTOS DE VENTAS', 'EXPENSE', 'D');
    PERFORM create_account(v_org_id, '61', 'COSTO DE VENTAS Y PRESTACIÓN DE SERVICIOS', 'EXPENSE', 'D', '6');
    PERFORM create_account(v_org_id, '6135', 'COMERCIO AL POR MAYOR Y MENOR', 'EXPENSE', 'D', '61');
    PERFORM create_account(v_org_id, '6135.01', 'Costo de Mercancías Vendidas', 'EXPENSE', 'D', '6135');
END $$;

-- ==============================================
-- COMPROBANTES DE PRUEBA
-- ==============================================

INSERT INTO vouchers (id, organization_id, voucher_number, voucher_type, voucher_date, 
                     description, total_amount, currency_code, status, metadata) VALUES
('880e8400-e29b-41d4-a716-446655440001', '11111111-1111-1111-1111-111111111111', 
 'FV-001-2024', 'INVOICE', '2024-01-15', 
 'Factura venta productos tecnológicos', 11900000, 'COP', 'PENDING',
 '{
   "customer": {"tax_id": "900789123-4", "name": "Cliente Ejemplo S.A.S."},
   "items": [
     {"description": "Laptop Dell", "quantity": 2, "unit_price": 5000000, "tax_rate": 19}
   ],
   "subtotal": 10000000,
   "taxes": {"iva_19": 1900000}
 }'),
 
('880e8400-e29b-41d4-a716-446655440002', '11111111-1111-1111-1111-111111111111', 
 'RC-001-2024', 'RECEIPT', '2024-01-20', 
 'Recibo de caja cliente FV-001', 11900000, 'COP', 'PENDING',
 '{
   "payment_method": "transfer",
   "bank_reference": "REF123456",
   "related_invoice": "FV-001-2024"
 }'),
 
('880e8400-e29b-41d4-a716-446655440003', '11111111-1111-1111-1111-111111111111', 
 'CE-001-2024', 'PAYMENT', '2024-01-25', 
 'Pago servicios públicos enero', 850000, 'COP', 'PENDING',
 '{
   "expense_type": "utilities",
   "period": "2024-01",
   "provider": "Empresa de Energía"
 }');

-- ==============================================
-- PLANTILLAS DSL
-- ==============================================

INSERT INTO accounting_templates (organization_id, template_code, voucher_type, 
                                 country_code, dsl_content, version) VALUES
-- Plantilla para facturas de venta Colombia
('11111111-1111-1111-1111-111111111111', 'INVOICE_SALE_CO', 'INVOICE', 'CO', 
'template invoice_sale_co {
  // Variables del comprobante
  let subtotal = voucher.metadata.subtotal
  let iva_19 = voucher.metadata.taxes.iva_19
  let total = voucher.total_amount
  
  // Validaciones
  require subtotal > 0 : "El subtotal debe ser mayor a cero"
  require total == subtotal + iva_19 : "El total no coincide"
  
  // Generar asiento contable
  entry {
    // Debitar cuentas por cobrar
    debit account("1305.01") amount(total) {
      description = "Factura " + voucher.voucher_number
      metadata.customer = voucher.metadata.customer
    }
    
    // Acreditar ingresos
    credit account("4135.01") amount(subtotal) {
      description = "Venta de mercancías"
    }
    
    // Acreditar IVA
    credit account("2408.01") amount(iva_19) {
      description = "IVA 19% por pagar"
    }
  }
}', 1),

-- Plantilla para recibos de caja
('11111111-1111-1111-1111-111111111111', 'RECEIPT_CASH_CO', 'RECEIPT', 'CO',
'template receipt_cash_co {
  let amount = voucher.total_amount
  let payment_method = voucher.metadata.payment_method
  
  entry {
    // Debitar banco o caja según método de pago
    if payment_method == "cash" {
      debit account("1105.01") amount(amount) {
        description = "Recibo " + voucher.voucher_number
      }
    } else {
      debit account("1110.01") amount(amount) {
        description = "Recibo " + voucher.voucher_number
        metadata.reference = voucher.metadata.bank_reference
      }
    }
    
    // Acreditar cuentas por cobrar
    credit account("1305.01") amount(amount) {
      description = "Abono factura " + voucher.metadata.related_invoice
    }
  }
}', 1),

-- Plantilla para pagos
('11111111-1111-1111-1111-111111111111', 'PAYMENT_EXPENSE_CO', 'PAYMENT', 'CO',
'template payment_expense_co {
  let amount = voucher.total_amount
  let expense_type = voucher.metadata.expense_type
  
  entry {
    // Debitar gasto según tipo
    if expense_type == "utilities" {
      debit account("5135.01") amount(amount) {
        description = "Servicios públicos " + voucher.metadata.period
      }
    }
    
    // Acreditar banco
    credit account("1110.01") amount(amount) {
      description = "Pago " + voucher.voucher_number
    }
  }
}', 1);

-- ==============================================
-- REGLAS FISCALES
-- ==============================================

INSERT INTO fiscal_rules (country_code, rule_code, rule_type, dsl_content, 
                         effective_date, metadata) VALUES
('CO', 'IVA_RATES', 'TAX_VALIDATION', 
'rule iva_rates_co {
  define valid_rates = [0, 5, 19]
  
  validate invoice {
    for item in invoice.items {
      require item.tax_rate in valid_rates : 
        "Tasa de IVA inválida: " + item.tax_rate
    }
  }
}', '2024-01-01',
'{"description": "Tasas de IVA válidas en Colombia"}'),

('CO', 'INVOICE_NUMBERING', 'DOCUMENT_VALIDATION',
'rule invoice_numbering_co {
  pattern = "^[A-Z]{2,4}-[0-9]{3,6}-[0-9]{4}$"
  
  validate voucher {
    if voucher.type == "INVOICE" {
      require voucher.number matches pattern :
        "Formato de factura inválido"
    }
  }
}', '2024-01-01',
'{"description": "Formato de numeración de facturas"}');

-- ==============================================
-- ASIENTOS CONTABLES DE EJEMPLO
-- ==============================================

-- Asiento para la factura FV-001-2024
INSERT INTO journal_entries (id, organization_id, entry_number, entry_date, 
                           voucher_id, description, entry_type, period, 
                           status, created_by, approved_by, approved_at) VALUES
('990e8400-e29b-41d4-a716-446655440001', '11111111-1111-1111-1111-111111111111',
 1, '2024-01-15', '880e8400-e29b-41d4-a716-446655440001',
 'Venta según factura FV-001-2024', 'STANDARD', '2024-01',
 'POSTED', '550e8400-e29b-41d4-a716-446655440004',
 '550e8400-e29b-41d4-a716-446655440002', NOW());

-- Líneas del asiento
INSERT INTO journal_lines (journal_entry_id, line_number, account_id, 
                          debit_amount, credit_amount, description) VALUES
-- Debito a cuentas por cobrar
('990e8400-e29b-41d4-a716-446655440001', 1,
 (SELECT id FROM chart_of_accounts WHERE organization_id = '11111111-1111-1111-1111-111111111111' 
  AND account_code = '1305.01'),
 11900000, 0, 'Factura FV-001-2024'),
 
-- Crédito a ventas
('990e8400-e29b-41d4-a716-446655440001', 2,
 (SELECT id FROM chart_of_accounts WHERE organization_id = '11111111-1111-1111-1111-111111111111' 
  AND account_code = '4135.01'),
 0, 10000000, 'Venta de mercancías'),
 
-- Crédito a IVA por pagar
('990e8400-e29b-41d4-a716-446655440001', 3,
 (SELECT id FROM chart_of_accounts WHERE organization_id = '11111111-1111-1111-1111-111111111111' 
  AND account_code = '2408.01'),
 0, 1900000, 'IVA 19% por pagar');

-- ==============================================
-- LIBROS CONTABLES DE EJEMPLO
-- ==============================================

INSERT INTO accounting_books (organization_id, book_type, period_start, 
                            period_end, status, file_url) VALUES
('11111111-1111-1111-1111-111111111111', 'JOURNAL', 
 '2024-01-01', '2024-01-31', 'COMPLETED',
 '/reports/2024/01/journal_202401.pdf'),
 
('11111111-1111-1111-1111-111111111111', 'TRIAL_BALANCE', 
 '2024-01-01', '2024-01-31', 'GENERATING', NULL);

-- ==============================================
-- DATOS DE AUDITORÍA
-- ==============================================

INSERT INTO audit_log (organization_id, user_id, action, entity_type, 
                      entity_id, new_values, ip_address) VALUES
('11111111-1111-1111-1111-111111111111', '550e8400-e29b-41d4-a716-446655440004',
 'CREATE', 'voucher', '880e8400-e29b-41d4-a716-446655440001',
 '{"voucher_number": "FV-001-2024", "total": 11900000}',
 '192.168.1.100'::inet),
 
('11111111-1111-1111-1111-111111111111', '550e8400-e29b-41d4-a716-446655440002',
 'APPROVE', 'journal_entry', '990e8400-e29b-41d4-a716-446655440001',
 '{"status": "POSTED", "approved_at": "2024-01-15T10:30:00Z"}',
 '192.168.1.101'::inet);

-- ==============================================
-- ACTUALIZAR VISTA MATERIALIZADA
-- ==============================================

REFRESH MATERIALIZED VIEW account_balances;

-- ==============================================
-- LIMPIAR FUNCIONES TEMPORALES
-- ==============================================

DROP FUNCTION IF EXISTS create_account(UUID, VARCHAR, VARCHAR, VARCHAR, CHAR, VARCHAR);

-- ==============================================
-- VERIFICACIÓN FINAL
-- ==============================================

DO $$
DECLARE
    v_org_count INTEGER;
    v_user_count INTEGER;
    v_account_count INTEGER;
    v_voucher_count INTEGER;
    v_entry_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO v_org_count FROM organizations;
    SELECT COUNT(*) INTO v_user_count FROM users;
    SELECT COUNT(*) INTO v_account_count FROM chart_of_accounts;
    SELECT COUNT(*) INTO v_voucher_count FROM vouchers;
    SELECT COUNT(*) INTO v_entry_count FROM journal_entries;
    
    RAISE NOTICE '======================================';
    RAISE NOTICE 'Datos iniciales cargados exitosamente:';
    RAISE NOTICE '======================================';
    RAISE NOTICE 'Organizaciones: %', v_org_count;
    RAISE NOTICE 'Usuarios: %', v_user_count;
    RAISE NOTICE 'Cuentas contables: %', v_account_count;
    RAISE NOTICE 'Comprobantes: %', v_voucher_count;
    RAISE NOTICE 'Asientos contables: %', v_entry_count;
    RAISE NOTICE '======================================';
END $$;

-- ==============================================
-- FIN DEL SCRIPT DE DATOS INICIALES
-- ==============================================