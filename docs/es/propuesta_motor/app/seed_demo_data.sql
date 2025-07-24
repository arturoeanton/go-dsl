-- Script para agregar datos de demostración
-- Comprobantes, asientos contables y terceros adicionales

-- Agregar más terceros
INSERT INTO third_parties (id, created_at, updated_at, organization_id, code, document_type, document_number, verification_digit, first_name, last_name, company_name, person_type, third_party_type, taxpayer_type, is_active) VALUES
-- Clientes
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), 'CLI002', 'NIT', '900234567', '8', NULL, NULL, 'Tecnología Avanzada S.A.S', 'JURIDICA', 'CUSTOMER', 'RESPONSABLE_IVA', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), 'CLI003', 'NIT', '900345678', '9', NULL, NULL, 'Comercializadora del Pacífico Ltda', 'JURIDICA', 'CUSTOMER', 'RESPONSABLE_IVA', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), 'CLI004', 'CC', '98765432', NULL, 'María', 'González', NULL, 'NATURAL', 'CUSTOMER', 'NO_RESPONSABLE_IVA', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), 'CLI005', 'NIT', '900456789', '0', NULL, NULL, 'Distribuidora Nacional S.A', 'JURIDICA', 'CUSTOMER', 'RESPONSABLE_IVA', 1),

-- Proveedores
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), 'PROV002', 'NIT', '800234567', '8', NULL, NULL, 'Suministros Industriales S.A', 'JURIDICA', 'SUPPLIER', 'RESPONSABLE_IVA', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), 'PROV003', 'NIT', '800345678', '9', NULL, NULL, 'Papelería y Oficina Ltda', 'JURIDICA', 'SUPPLIER', 'RESPONSABLE_IVA', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), 'PROV004', 'CC', '87654321', NULL, 'Carlos', 'Rodríguez', NULL, 'NATURAL', 'SUPPLIER', 'NO_RESPONSABLE_IVA', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), 'PROV005', 'NIT', '800456789', '0', NULL, NULL, 'Servicios Profesionales S.A.S', 'JURIDICA', 'SUPPLIER', 'RESPONSABLE_IVA', 1),

-- Empleados
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), 'EMP001', 'CC', '11223344', NULL, 'Ana', 'Martínez', NULL, 'NATURAL', 'EMPLOYEE', 'NO_RESPONSABLE_IVA', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), 'EMP002', 'CC', '22334455', NULL, 'Luis', 'Hernández', NULL, 'NATURAL', 'EMPLOYEE', 'NO_RESPONSABLE_IVA', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), 'EMP003', 'CC', '33445566', NULL, 'Laura', 'García', NULL, 'NATURAL', 'EMPLOYEE', 'NO_RESPONSABLE_IVA', 1);

-- Crear comprobantes de muestra
-- Variables para IDs únicos
-- Nota: En SQLite, usaremos randomblob para generar IDs únicos

-- Comprobante 1: Factura de Venta
INSERT INTO vouchers (id, created_at, updated_at, organization_id, number, voucher_type, date, description, reference, period_id, status, total_debit, total_credit, is_balanced, created_by_user_id) 
VALUES (
    'voucher-001', 
    datetime('now'), 
    datetime('now'), 
    (SELECT id FROM organizations LIMIT 1), 
    'FV-2024-001', 
    'SALE', 
    date('now', '-10 days'),
    'Venta de servicios de consultoría', 
    'CONT-2024-001',
    'current-period-id',
    'POSTED',
    1190000,
    1190000,
    1,
    'system'
);

-- Líneas del comprobante 1
INSERT INTO voucher_lines (id, created_at, updated_at, voucher_id, account_id, description, debit_amount, credit_amount, line_number) VALUES
-- Débito a Clientes
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-001', 
 (SELECT id FROM accounts WHERE code = '130505' LIMIT 1), 
 'Factura venta servicios consultoría', 1190000, 0, 1),
-- Crédito a Ingresos
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-001', 
 (SELECT id FROM accounts WHERE code = '413510' LIMIT 1), 
 'Ingresos por servicios', 0, 1000000, 2),
-- Crédito a IVA
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-001', 
 (SELECT id FROM accounts WHERE code = '240801' LIMIT 1), 
 'IVA 19% generado', 0, 190000, 3);

-- Comprobante 2: Factura de Compra
INSERT INTO vouchers (id, created_at, updated_at, organization_id, number, voucher_type, date, description, reference, period_id, status, total_debit, total_credit, is_balanced, created_by_user_id) 
VALUES (
    'voucher-002', 
    datetime('now'), 
    datetime('now'), 
    (SELECT id FROM organizations LIMIT 1), 
    'FC-2024-001', 
    'PURCHASE', 
    date('now', '-8 days'),
    'Compra de suministros de oficina', 
    'FAC-PROV-123',
    'current-period-id',
    'POSTED',
    357000,
    357000,
    1,
    'system'
);

-- Líneas del comprobante 2
INSERT INTO voucher_lines (id, created_at, updated_at, voucher_id, account_id, description, debit_amount, credit_amount, line_number) VALUES
-- Débito a Gastos
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-002', 
 (SELECT id FROM accounts WHERE code = '519530' LIMIT 1), 
 'Útiles y papelería', 300000, 0, 1),
-- Débito a IVA descontable
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-002', 
 (SELECT id FROM accounts WHERE code = '240802' LIMIT 1), 
 'IVA descontable 19%', 57000, 0, 2),
-- Crédito a Proveedores
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-002', 
 (SELECT id FROM accounts WHERE code = '220501' LIMIT 1), 
 'Por pagar a proveedor', 0, 357000, 3);

-- Comprobante 3: Nómina mensual
INSERT INTO vouchers (id, created_at, updated_at, organization_id, number, voucher_type, date, description, reference, period_id, status, total_debit, total_credit, is_balanced, created_by_user_id) 
VALUES (
    'voucher-003', 
    datetime('now'), 
    datetime('now'), 
    (SELECT id FROM organizations LIMIT 1), 
    'NOM-2024-001', 
    'PAYROLL', 
    date('now', '-5 days'),
    'Nómina mes de enero 2024', 
    'NOM-ENE-2024',
    'current-period-id',
    'POSTED',
    15678900,
    15678900,
    1,
    'system'
);

-- Líneas del comprobante 3 (simplificado)
INSERT INTO voucher_lines (id, created_at, updated_at, voucher_id, account_id, description, debit_amount, credit_amount, line_number) VALUES
-- Débito a Sueldos
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-003', 
 (SELECT id FROM accounts WHERE code = '510506' LIMIT 1), 
 'Sueldos y salarios', 10000000, 0, 1),
-- Débito a Auxilio transporte
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-003', 
 (SELECT id FROM accounts WHERE code = '510520' LIMIT 1), 
 'Auxilio de transporte', 1200000, 0, 2),
-- Débito a Aportes salud empresa
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-003', 
 (SELECT id FROM accounts WHERE code = '510569' LIMIT 1), 
 'Aportes EPS empresa', 850000, 0, 3),
-- Débito a Aportes pensión empresa
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-003', 
 (SELECT id FROM accounts WHERE code = '510570' LIMIT 1), 
 'Aportes pensión empresa', 1200000, 0, 4),
-- Débito a Cesantías
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-003', 
 (SELECT id FROM accounts WHERE code = '510527' LIMIT 1), 
 'Cesantías', 933300, 0, 5),
-- Débito a Prima
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-003', 
 (SELECT id FROM accounts WHERE code = '510533' LIMIT 1), 
 'Prima de servicios', 933300, 0, 6),
-- Débito a Vacaciones
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-003', 
 (SELECT id FROM accounts WHERE code = '510536' LIMIT 1), 
 'Vacaciones', 416700, 0, 7),
-- Débito a Intereses cesantías
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-003', 
 (SELECT id FROM accounts WHERE code = '510530' LIMIT 1), 
 'Intereses sobre cesantías', 112000, 0, 8),
-- Débito a ARL
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-003', 
 (SELECT id FROM accounts WHERE code = '510568' LIMIT 1), 
 'ARL', 52200, 0, 9),
-- Débito a Caja compensación
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-003', 
 (SELECT id FROM accounts WHERE code = '510572' LIMIT 1), 
 'Caja de compensación', 400000, 0, 10),
-- Débito a ICBF
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-003', 
 (SELECT id FROM accounts WHERE code = '510575' LIMIT 1), 
 'ICBF', 300000, 0, 11),
-- Débito a SENA
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-003', 
 (SELECT id FROM accounts WHERE code = '510578' LIMIT 1), 
 'SENA', 200000, 0, 12),
-- Crédito a Salarios por pagar
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-003', 
 (SELECT id FROM accounts WHERE code = '250501' LIMIT 1), 
 'Salarios netos por pagar', 0, 8400000, 13),
-- Crédito a Retenciones empleados
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-003', 
 (SELECT id FROM accounts WHERE code = '237005' LIMIT 1), 
 'Aportes salud y pensión empleados', 0, 800000, 14),
-- Crédito a Aportes por pagar
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-003', 
 (SELECT id FROM accounts WHERE code = '237005' LIMIT 1), 
 'Seguridad social por pagar', 0, 2302200, 15),
-- Crédito a Parafiscales por pagar
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-003', 
 (SELECT id FROM accounts WHERE code = '237010' LIMIT 1), 
 'Parafiscales por pagar', 0, 900000, 16),
-- Crédito a Prestaciones por pagar
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-003', 
 (SELECT id FROM accounts WHERE code = '261005' LIMIT 1), 
 'Cesantías por pagar', 0, 933300, 17),
-- Crédito a Prima por pagar
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-003', 
 (SELECT id FROM accounts WHERE code = '261015' LIMIT 1), 
 'Prima por pagar', 0, 933300, 18),
-- Crédito a Vacaciones por pagar
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-003', 
 (SELECT id FROM accounts WHERE code = '261020' LIMIT 1), 
 'Vacaciones por pagar', 0, 416700, 19),
-- Crédito a Intereses cesantías por pagar
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-003', 
 (SELECT id FROM accounts WHERE code = '261010' LIMIT 1), 
 'Intereses cesantías por pagar', 0, 112000, 20);

-- Comprobante 4: Pago a proveedor
INSERT INTO vouchers (id, created_at, updated_at, organization_id, number, voucher_type, date, description, reference, period_id, status, total_debit, total_credit, is_balanced, created_by_user_id) 
VALUES (
    'voucher-004', 
    datetime('now'), 
    datetime('now'), 
    (SELECT id FROM organizations LIMIT 1), 
    'CE-2024-001', 
    'PAYMENT', 
    date('now', '-3 days'),
    'Pago factura proveedor suministros', 
    'TRANSF-12345',
    'current-period-id',
    'POSTED',
    357000,
    357000,
    1,
    'system'
);

-- Líneas del comprobante 4
INSERT INTO voucher_lines (id, created_at, updated_at, voucher_id, account_id, description, debit_amount, credit_amount, line_number) VALUES
-- Débito a Proveedores
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-004', 
 (SELECT id FROM accounts WHERE code = '220501' LIMIT 1), 
 'Pago factura FAC-PROV-123', 357000, 0, 1),
-- Crédito a Bancos
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-004', 
 (SELECT id FROM accounts WHERE code = '111005' LIMIT 1), 
 'Transferencia bancaria', 0, 357000, 2);

-- Comprobante 5: Recibo de caja
INSERT INTO vouchers (id, created_at, updated_at, organization_id, number, voucher_type, date, description, reference, period_id, status, total_debit, total_credit, is_balanced, created_by_user_id) 
VALUES (
    'voucher-005', 
    datetime('now'), 
    datetime('now'), 
    (SELECT id FROM organizations LIMIT 1), 
    'RC-2024-001', 
    'RECEIPT', 
    date('now', '-2 days'),
    'Recibo pago cliente factura FV-2024-001', 
    'CONSIG-67890',
    'current-period-id',
    'POSTED',
    1190000,
    1190000,
    1,
    'system'
);

-- Líneas del comprobante 5
INSERT INTO voucher_lines (id, created_at, updated_at, voucher_id, account_id, description, debit_amount, credit_amount, line_number) VALUES
-- Débito a Bancos
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-005', 
 (SELECT id FROM accounts WHERE code = '111005' LIMIT 1), 
 'Consignación bancaria', 1190000, 0, 1),
-- Crédito a Clientes
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-005', 
 (SELECT id FROM accounts WHERE code = '130505' LIMIT 1), 
 'Abono factura FV-2024-001', 0, 1190000, 2);

-- Comprobante 6: Servicios públicos
INSERT INTO vouchers (id, created_at, updated_at, organization_id, number, voucher_type, date, description, reference, period_id, status, total_debit, total_credit, is_balanced, created_by_user_id) 
VALUES (
    'voucher-006', 
    datetime('now'), 
    datetime('now'), 
    (SELECT id FROM organizations LIMIT 1), 
    'GV-2024-001', 
    'EXPENSE', 
    date('now', '-1 days'),
    'Servicios públicos enero 2024', 
    'SERV-ENE-2024',
    'current-period-id',
    'DRAFT',
    892500,
    892500,
    1,
    'system'
);

-- Líneas del comprobante 6
INSERT INTO voucher_lines (id, created_at, updated_at, voucher_id, account_id, description, debit_amount, credit_amount, line_number) VALUES
-- Débito a Energía
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-006', 
 (SELECT id FROM accounts WHERE code = '513530' LIMIT 1), 
 'Energía eléctrica', 450000, 0, 1),
-- Débito a Acueducto
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-006', 
 (SELECT id FROM accounts WHERE code = '513525' LIMIT 1), 
 'Acueducto y alcantarillado', 250000, 0, 2),
-- Débito a Teléfono e Internet
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-006', 
 (SELECT id FROM accounts WHERE code = '513535' LIMIT 1), 
 'Teléfono e internet', 150000, 0, 3),
-- Débito a IVA descontable
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-006', 
 (SELECT id FROM accounts WHERE code = '240802' LIMIT 1), 
 'IVA servicios públicos', 42500, 0, 4),
-- Crédito a Cuentas por pagar
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-006', 
 (SELECT id FROM accounts WHERE code = '233550' LIMIT 1), 
 'Servicios públicos por pagar', 0, 892500, 5);

-- Comprobante 7: Depreciación mensual
INSERT INTO vouchers (id, created_at, updated_at, organization_id, number, voucher_type, date, description, reference, period_id, status, total_debit, total_credit, is_balanced, created_by_user_id) 
VALUES (
    'voucher-007', 
    datetime('now'), 
    datetime('now'), 
    (SELECT id FROM organizations LIMIT 1), 
    'AJ-2024-001', 
    'ADJUSTMENT', 
    date('now'),
    'Depreciación mensual enero 2024', 
    'DEP-ENE-2024',
    'current-period-id',
    'DRAFT',
    2500000,
    2500000,
    1,
    'system'
);

-- Líneas del comprobante 7
INSERT INTO voucher_lines (id, created_at, updated_at, voucher_id, account_id, description, debit_amount, credit_amount, line_number) VALUES
-- Débito a Gasto depreciación edificios
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-007', 
 (SELECT id FROM accounts WHERE code = '516005' LIMIT 1), 
 'Depreciación edificios', 1000000, 0, 1),
-- Débito a Gasto depreciación maquinaria
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-007', 
 (SELECT id FROM accounts WHERE code = '516010' LIMIT 1), 
 'Depreciación maquinaria', 500000, 0, 2),
-- Débito a Gasto depreciación equipos de cómputo
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-007', 
 (SELECT id FROM accounts WHERE code = '516020' LIMIT 1), 
 'Depreciación equipos cómputo', 750000, 0, 3),
-- Débito a Gasto depreciación vehículos
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-007', 
 (SELECT id FROM accounts WHERE code = '516035' LIMIT 1), 
 'Depreciación vehículos', 250000, 0, 4),
-- Crédito a Depreciación acumulada edificios
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-007', 
 (SELECT id FROM accounts WHERE code = '159205' LIMIT 1), 
 'Depreciación acumulada edificios', 0, 1000000, 5),
-- Crédito a Depreciación acumulada maquinaria
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-007', 
 (SELECT id FROM accounts WHERE code = '159210' LIMIT 1), 
 'Depreciación acumulada maquinaria', 0, 500000, 6),
-- Crédito a Depreciación acumulada equipos
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-007', 
 (SELECT id FROM accounts WHERE code = '159220' LIMIT 1), 
 'Depreciación acumulada equipos', 0, 750000, 7),
-- Crédito a Depreciación acumulada vehículos
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'voucher-007', 
 (SELECT id FROM accounts WHERE code = '159235' LIMIT 1), 
 'Depreciación acumulada vehículos', 0, 250000, 8);

-- Crear asientos contables para los comprobantes POSTED
-- Estos normalmente se crearían automáticamente, pero los agregamos para el demo

-- Asiento 1: De la factura de venta
INSERT INTO journal_entries (id, created_at, updated_at, organization_id, entry_number, date, description, reference, voucher_id, period_id, status, total_debit, total_credit, is_reversed, created_by_user_id, posted_at) 
VALUES (
    'entry-001',
    datetime('now'),
    datetime('now'),
    (SELECT id FROM organizations LIMIT 1),
    'AS-0001',
    date('now', '-10 days'),
    'Asiento de SALE - Venta de servicios de consultoría',
    'FV-2024-001',
    'voucher-001',
    'current-period-id',
    'POSTED',
    1190000,
    1190000,
    0,
    'system',
    datetime('now', '-10 days')
);

-- Líneas del asiento 1
INSERT INTO journal_lines (id, created_at, updated_at, journal_entry_id, account_id, description, debit_amount, credit_amount, line_number) VALUES
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'entry-001', 
 (SELECT id FROM accounts WHERE code = '130505' LIMIT 1), 
 'Factura venta servicios consultoría', 1190000, 0, 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'entry-001', 
 (SELECT id FROM accounts WHERE code = '413510' LIMIT 1), 
 'Ingresos por servicios', 0, 1000000, 2),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'entry-001', 
 (SELECT id FROM accounts WHERE code = '240801' LIMIT 1), 
 'IVA 19% generado', 0, 190000, 3);

-- Asiento 2: De la factura de compra
INSERT INTO journal_entries (id, created_at, updated_at, organization_id, entry_number, date, description, reference, voucher_id, period_id, status, total_debit, total_credit, is_reversed, created_by_user_id, posted_at) 
VALUES (
    'entry-002',
    datetime('now'),
    datetime('now'),
    (SELECT id FROM organizations LIMIT 1),
    'AS-0002',
    date('now', '-8 days'),
    'Asiento de PURCHASE - Compra de suministros de oficina',
    'FC-2024-001',
    'voucher-002',
    'current-period-id',
    'POSTED',
    357000,
    357000,
    0,
    'system',
    datetime('now', '-8 days')
);

-- Líneas del asiento 2
INSERT INTO journal_lines (id, created_at, updated_at, journal_entry_id, account_id, description, debit_amount, credit_amount, line_number) VALUES
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'entry-002', 
 (SELECT id FROM accounts WHERE code = '519530' LIMIT 1), 
 'Útiles y papelería', 300000, 0, 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'entry-002', 
 (SELECT id FROM accounts WHERE code = '240802' LIMIT 1), 
 'IVA descontable 19%', 57000, 0, 2),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'entry-002', 
 (SELECT id FROM accounts WHERE code = '220501' LIMIT 1), 
 'Por pagar a proveedor', 0, 357000, 3);

-- Asiento 3: De la nómina
INSERT INTO journal_entries (id, created_at, updated_at, organization_id, entry_number, date, description, reference, voucher_id, period_id, status, total_debit, total_credit, is_reversed, created_by_user_id, posted_at) 
VALUES (
    'entry-003',
    datetime('now'),
    datetime('now'),
    (SELECT id FROM organizations LIMIT 1),
    'AS-0003',
    date('now', '-5 days'),
    'Asiento de PAYROLL - Nómina mes de enero 2024',
    'NOM-2024-001',
    'voucher-003',
    'current-period-id',
    'POSTED',
    15678900,
    15678900,
    0,
    'system',
    datetime('now', '-5 days')
);

-- Asiento 4: Del pago a proveedor
INSERT INTO journal_entries (id, created_at, updated_at, organization_id, entry_number, date, description, reference, voucher_id, period_id, status, total_debit, total_credit, is_reversed, created_by_user_id, posted_at) 
VALUES (
    'entry-004',
    datetime('now'),
    datetime('now'),
    (SELECT id FROM organizations LIMIT 1),
    'AS-0004',
    date('now', '-3 days'),
    'Asiento de PAYMENT - Pago factura proveedor suministros',
    'CE-2024-001',
    'voucher-004',
    'current-period-id',
    'POSTED',
    357000,
    357000,
    0,
    'system',
    datetime('now', '-3 days')
);

-- Líneas del asiento 4
INSERT INTO journal_lines (id, created_at, updated_at, journal_entry_id, account_id, description, debit_amount, credit_amount, line_number) VALUES
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'entry-004', 
 (SELECT id FROM accounts WHERE code = '220501' LIMIT 1), 
 'Pago factura FAC-PROV-123', 357000, 0, 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'entry-004', 
 (SELECT id FROM accounts WHERE code = '111005' LIMIT 1), 
 'Transferencia bancaria', 0, 357000, 2);

-- Asiento 5: Del recibo de caja
INSERT INTO journal_entries (id, created_at, updated_at, organization_id, entry_number, date, description, reference, voucher_id, period_id, status, total_debit, total_credit, is_reversed, created_by_user_id, posted_at) 
VALUES (
    'entry-005',
    datetime('now'),
    datetime('now'),
    (SELECT id FROM organizations LIMIT 1),
    'AS-0005',
    date('now', '-2 days'),
    'Asiento de RECEIPT - Recibo pago cliente factura FV-2024-001',
    'RC-2024-001',
    'voucher-005',
    'current-period-id',
    'POSTED',
    1190000,
    1190000,
    0,
    'system',
    datetime('now', '-2 days')
);

-- Líneas del asiento 5
INSERT INTO journal_lines (id, created_at, updated_at, journal_entry_id, account_id, description, debit_amount, credit_amount, line_number) VALUES
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'entry-005', 
 (SELECT id FROM accounts WHERE code = '111005' LIMIT 1), 
 'Consignación bancaria', 1190000, 0, 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), 'entry-005', 
 (SELECT id FROM accounts WHERE code = '130505' LIMIT 1), 
 'Abono factura FV-2024-001', 0, 1190000, 2);

-- Verificar datos insertados
SELECT 'Terceros adicionales:' as info, COUNT(*) as total FROM third_parties;
SELECT 'Comprobantes creados:' as info, COUNT(*) as total FROM vouchers;
SELECT 'Asientos contables:' as info, COUNT(*) as total FROM journal_entries;
SELECT 'Comprobantes por tipo:' as info, voucher_type, COUNT(*) as cantidad FROM vouchers GROUP BY voucher_type;
SELECT 'Comprobantes por estado:' as info, status, COUNT(*) as cantidad FROM vouchers GROUP BY status;