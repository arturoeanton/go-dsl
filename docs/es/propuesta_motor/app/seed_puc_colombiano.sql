-- Script para agregar las cuentas del PUC Colombiano
-- Primero obtenemos el organization_id
-- Asumiendo que hay una sola organización en el POC

-- Eliminar cuentas existentes (opcional, comentar si no se desea)
-- DELETE FROM accounts;

-- Insertar cuentas del PUC Colombiano
-- NOTA: Reemplazar 'ORG_ID' con el ID real de la organización

-- CLASE 1: ACTIVO
INSERT INTO accounts (id, created_at, updated_at, organization_id, code, name, account_type, level, natural_balance, accepts_movement, puc_code, is_active) VALUES
-- Nivel 1
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '1', 'ACTIVO', 'ASSET', 1, 'D', 0, '1', 1),

-- Nivel 2 - DISPONIBLE
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '11', 'DISPONIBLE', 'ASSET', 2, 'D', 0, '11', 1),
-- Nivel 3
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '1105', 'CAJA', 'ASSET', 3, 'D', 0, '1105', 1),
-- Nivel 4
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '110505', 'CAJA GENERAL', 'ASSET', 4, 'D', 1, '110505', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '110510', 'CAJAS MENORES', 'ASSET', 4, 'D', 1, '110510', 1),
-- Nivel 3
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '1110', 'BANCOS', 'ASSET', 3, 'D', 0, '1110', 1),
-- Nivel 4
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '111005', 'BANCOS NACIONALES', 'ASSET', 4, 'D', 1, '111005', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '111010', 'BANCOS DEL EXTERIOR', 'ASSET', 4, 'D', 1, '111010', 1),

-- Nivel 2 - INVERSIONES
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '12', 'INVERSIONES', 'ASSET', 2, 'D', 0, '12', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '1205', 'ACCIONES', 'ASSET', 3, 'D', 0, '1205', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '120505', 'AGRICULTURA, GANADERÍA, CAZA Y SILVICULTURA', 'ASSET', 4, 'D', 1, '120505', 1),

-- Nivel 2 - DEUDORES
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '13', 'DEUDORES', 'ASSET', 2, 'D', 0, '13', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '1305', 'CLIENTES', 'ASSET', 3, 'D', 0, '1305', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '130505', 'CLIENTES NACIONALES', 'ASSET', 4, 'D', 1, '130505', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '130510', 'CLIENTES DEL EXTERIOR', 'ASSET', 4, 'D', 1, '130510', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '1325', 'CUENTAS POR COBRAR A SOCIOS Y ACCIONISTAS', 'ASSET', 3, 'D', 0, '1325', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '132505', 'A SOCIOS', 'ASSET', 4, 'D', 1, '132505', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '1330', 'ANTICIPOS Y AVANCES', 'ASSET', 3, 'D', 0, '1330', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '133005', 'A PROVEEDORES', 'ASSET', 4, 'D', 1, '133005', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '1355', 'ANTICIPO DE IMPUESTOS Y CONTRIBUCIONES', 'ASSET', 3, 'D', 0, '1355', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '135515', 'RETENCIÓN EN LA FUENTE', 'ASSET', 4, 'D', 1, '135515', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '135517', 'IMPUESTO DE INDUSTRIA Y COMERCIO RETENIDO', 'ASSET', 4, 'D', 1, '135517', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '1365', 'CUENTAS POR COBRAR A TRABAJADORES', 'ASSET', 3, 'D', 0, '1365', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '136505', 'PRÉSTAMOS', 'ASSET', 4, 'D', 1, '136505', 1),

-- Nivel 2 - INVENTARIOS
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '14', 'INVENTARIOS', 'ASSET', 2, 'D', 0, '14', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '1435', 'MERCANCÍAS NO FABRICADAS POR LA EMPRESA', 'ASSET', 3, 'D', 0, '1435', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '143505', 'MERCANCÍAS EN ALMACÉN', 'ASSET', 4, 'D', 1, '143505', 1),

-- Nivel 2 - PROPIEDAD PLANTA Y EQUIPO
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '15', 'PROPIEDADES PLANTA Y EQUIPO', 'ASSET', 2, 'D', 0, '15', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '1504', 'TERRENOS', 'ASSET', 3, 'D', 0, '1504', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '150405', 'URBANOS', 'ASSET', 4, 'D', 1, '150405', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '1516', 'CONSTRUCCIONES Y EDIFICACIONES', 'ASSET', 3, 'D', 0, '1516', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '151605', 'EDIFICIOS', 'ASSET', 4, 'D', 1, '151605', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '1520', 'MAQUINARIA Y EQUIPO', 'ASSET', 3, 'D', 0, '1520', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '152005', 'MAQUINARIA', 'ASSET', 4, 'D', 1, '152005', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '1524', 'EQUIPO DE OFICINA', 'ASSET', 3, 'D', 0, '1524', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '152405', 'MUEBLES Y ENSERES', 'ASSET', 4, 'D', 1, '152405', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '1528', 'EQUIPO DE COMPUTACIÓN Y COMUNICACIÓN', 'ASSET', 3, 'D', 0, '1528', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '152805', 'EQUIPOS DE PROCESAMIENTO DE DATOS', 'ASSET', 4, 'D', 1, '152805', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '1540', 'FLOTA Y EQUIPO DE TRANSPORTE', 'ASSET', 3, 'D', 0, '1540', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '154005', 'AUTOS, CAMIONETAS Y CAMPEROS', 'ASSET', 4, 'D', 1, '154005', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '1592', 'DEPRECIACIÓN ACUMULADA', 'ASSET', 3, 'C', 0, '1592', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '159205', 'CONSTRUCCIONES Y EDIFICACIONES', 'ASSET', 4, 'C', 1, '159205', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '159210', 'MAQUINARIA Y EQUIPO', 'ASSET', 4, 'C', 1, '159210', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '159215', 'EQUIPO DE OFICINA', 'ASSET', 4, 'C', 1, '159215', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '159220', 'EQUIPO DE COMPUTACIÓN Y COMUNICACIÓN', 'ASSET', 4, 'C', 1, '159220', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '159235', 'FLOTA Y EQUIPO DE TRANSPORTE', 'ASSET', 4, 'C', 1, '159235', 1),

-- CLASE 2: PASIVO
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '2', 'PASIVO', 'LIABILITY', 1, 'C', 0, '2', 1),

-- Nivel 2 - OBLIGACIONES FINANCIERAS
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '21', 'OBLIGACIONES FINANCIERAS', 'LIABILITY', 2, 'C', 0, '21', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '2105', 'BANCOS NACIONALES', 'LIABILITY', 3, 'C', 0, '2105', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '210505', 'SOBREGIROS', 'LIABILITY', 4, 'C', 1, '210505', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '210510', 'PAGARÉS', 'LIABILITY', 4, 'C', 1, '210510', 1),

-- Nivel 2 - PROVEEDORES
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '22', 'PROVEEDORES', 'LIABILITY', 2, 'C', 0, '22', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '2205', 'PROVEEDORES NACIONALES', 'LIABILITY', 3, 'C', 0, '2205', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '220501', 'PROVEEDORES NACIONALES', 'LIABILITY', 4, 'C', 1, '220501', 1),

-- Nivel 2 - CUENTAS POR PAGAR
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '23', 'CUENTAS POR PAGAR', 'LIABILITY', 2, 'C', 0, '23', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '2335', 'COSTOS Y GASTOS POR PAGAR', 'LIABILITY', 3, 'C', 0, '2335', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '233525', 'HONORARIOS', 'LIABILITY', 4, 'C', 1, '233525', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '233530', 'SERVICIOS TÉCNICOS', 'LIABILITY', 4, 'C', 1, '233530', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '233535', 'SERVICIOS DE MANTENIMIENTO', 'LIABILITY', 4, 'C', 1, '233535', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '233540', 'ARRENDAMIENTOS', 'LIABILITY', 4, 'C', 1, '233540', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '233550', 'SERVICIOS PÚBLICOS', 'LIABILITY', 4, 'C', 1, '233550', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '233595', 'OTROS', 'LIABILITY', 4, 'C', 1, '233595', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '2365', 'RETENCIÓN EN LA FUENTE', 'LIABILITY', 3, 'C', 0, '2365', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '236515', 'HONORARIOS', 'LIABILITY', 4, 'C', 1, '236515', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '236520', 'COMISIONES', 'LIABILITY', 4, 'C', 1, '236520', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '236525', 'SERVICIOS', 'LIABILITY', 4, 'C', 1, '236525', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '236530', 'ARRENDAMIENTOS', 'LIABILITY', 4, 'C', 1, '236530', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '236540', 'COMPRAS', 'LIABILITY', 4, 'C', 1, '236540', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '2367', 'IMPUESTO A LAS VENTAS RETENIDO', 'LIABILITY', 3, 'C', 0, '2367', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '236705', 'RETENCIÓN DE IVA RÉGIMEN COMÚN', 'LIABILITY', 4, 'C', 1, '236705', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '2368', 'IMPUESTO DE INDUSTRIA Y COMERCIO RETENIDO', 'LIABILITY', 3, 'C', 0, '2368', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '236801', 'RETENCIÓN ICA', 'LIABILITY', 4, 'C', 1, '236801', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '2370', 'RETENCIONES Y APORTES DE NÓMINA', 'LIABILITY', 3, 'C', 0, '2370', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '237005', 'APORTES A ENTIDADES PROMOTORAS DE SALUD EPS', 'LIABILITY', 4, 'C', 1, '237005', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '237006', 'APORTES A ADMINISTRADORAS DE RIESGOS PROFESIONALES ARP', 'LIABILITY', 4, 'C', 1, '237006', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '237010', 'APORTES AL ICBF, SENA Y CAJAS', 'LIABILITY', 4, 'C', 1, '237010', 1),

-- Nivel 2 - IMPUESTOS
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '24', 'IMPUESTOS, GRAVÁMENES Y TASAS', 'LIABILITY', 2, 'C', 0, '24', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '2404', 'DE RENTA Y COMPLEMENTARIOS', 'LIABILITY', 3, 'C', 0, '2404', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '240405', 'VIGENCIA FISCAL CORRIENTE', 'LIABILITY', 4, 'C', 1, '240405', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '2408', 'IMPUESTO SOBRE LAS VENTAS POR PAGAR', 'LIABILITY', 3, 'C', 0, '2408', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '240801', 'IVA GENERADO', 'LIABILITY', 4, 'C', 1, '240801', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '240802', 'IVA DESCONTABLE', 'LIABILITY', 4, 'D', 1, '240802', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '2412', 'DE INDUSTRIA Y COMERCIO', 'LIABILITY', 3, 'C', 0, '2412', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '241205', 'VIGENCIA FISCAL CORRIENTE', 'LIABILITY', 4, 'C', 1, '241205', 1),

-- Nivel 2 - OBLIGACIONES LABORALES
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '25', 'OBLIGACIONES LABORALES', 'LIABILITY', 2, 'C', 0, '25', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '2505', 'SALARIOS POR PAGAR', 'LIABILITY', 3, 'C', 0, '2505', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '250501', 'SALARIOS POR PAGAR', 'LIABILITY', 4, 'C', 1, '250501', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '2510', 'CESANTÍAS CONSOLIDADAS', 'LIABILITY', 3, 'C', 0, '2510', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '251005', 'CESANTÍAS', 'LIABILITY', 4, 'C', 1, '251005', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '2515', 'INTERESES SOBRE CESANTÍAS', 'LIABILITY', 3, 'C', 0, '2515', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '251505', 'INTERESES SOBRE CESANTÍAS', 'LIABILITY', 4, 'C', 1, '251505', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '2520', 'PRIMA DE SERVICIOS', 'LIABILITY', 3, 'C', 0, '2520', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '252005', 'PRIMA DE SERVICIOS', 'LIABILITY', 4, 'C', 1, '252005', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '2525', 'VACACIONES CONSOLIDADAS', 'LIABILITY', 3, 'C', 0, '2525', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '252505', 'VACACIONES', 'LIABILITY', 4, 'C', 1, '252505', 1),

-- Nivel 2 - PASIVOS ESTIMADOS Y PROVISIONES
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '26', 'PASIVOS ESTIMADOS Y PROVISIONES', 'LIABILITY', 2, 'C', 0, '26', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '2610', 'PARA OBLIGACIONES LABORALES', 'LIABILITY', 3, 'C', 0, '2610', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '261005', 'CESANTÍAS', 'LIABILITY', 4, 'C', 1, '261005', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '261010', 'INTERESES SOBRE CESANTÍAS', 'LIABILITY', 4, 'C', 1, '261010', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '261015', 'PRIMA DE SERVICIOS', 'LIABILITY', 4, 'C', 1, '261015', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '261020', 'VACACIONES', 'LIABILITY', 4, 'C', 1, '261020', 1),

-- CLASE 3: PATRIMONIO
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '3', 'PATRIMONIO', 'EQUITY', 1, 'C', 0, '3', 1),

-- Nivel 2 - CAPITAL SOCIAL
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '31', 'CAPITAL SOCIAL', 'EQUITY', 2, 'C', 0, '31', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '3105', 'CAPITAL SUSCRITO Y PAGADO', 'EQUITY', 3, 'C', 0, '3105', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '310505', 'CAPITAL AUTORIZADO', 'EQUITY', 4, 'C', 1, '310505', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '310510', 'CAPITAL POR SUSCRIBIR', 'EQUITY', 4, 'D', 1, '310510', 1),

-- Nivel 2 - RESERVAS
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '33', 'RESERVAS', 'EQUITY', 2, 'C', 0, '33', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '3305', 'RESERVA LEGAL', 'EQUITY', 3, 'C', 0, '3305', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '330505', 'RESERVA LEGAL', 'EQUITY', 4, 'C', 1, '330505', 1),

-- Nivel 2 - RESULTADOS DEL EJERCICIO
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '36', 'RESULTADOS DEL EJERCICIO', 'EQUITY', 2, 'C', 0, '36', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '3605', 'UTILIDAD DEL EJERCICIO', 'EQUITY', 3, 'C', 0, '3605', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '360505', 'UTILIDAD DEL EJERCICIO', 'EQUITY', 4, 'C', 1, '360505', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '3610', 'PÉRDIDA DEL EJERCICIO', 'EQUITY', 3, 'D', 0, '3610', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '361005', 'PÉRDIDA DEL EJERCICIO', 'EQUITY', 4, 'D', 1, '361005', 1),

-- Nivel 2 - RESULTADOS DE EJERCICIOS ANTERIORES
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '37', 'RESULTADOS DE EJERCICIOS ANTERIORES', 'EQUITY', 2, 'C', 0, '37', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '3705', 'UTILIDADES ACUMULADAS', 'EQUITY', 3, 'C', 0, '3705', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '370505', 'UTILIDADES ACUMULADAS', 'EQUITY', 4, 'C', 1, '370505', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '3710', 'PÉRDIDAS ACUMULADAS', 'EQUITY', 3, 'D', 0, '3710', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '371005', 'PÉRDIDAS ACUMULADAS', 'EQUITY', 4, 'D', 1, '371005', 1),

-- CLASE 4: INGRESOS
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '4', 'INGRESOS', 'INCOME', 1, 'C', 0, '4', 1),

-- Nivel 2 - OPERACIONALES
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '41', 'OPERACIONALES', 'INCOME', 2, 'C', 0, '41', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '4135', 'COMERCIO AL POR MAYOR Y AL POR MENOR', 'INCOME', 3, 'C', 0, '4135', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '413505', 'VENTA DE PRODUCTOS', 'INCOME', 4, 'C', 1, '413505', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '413510', 'VENTA DE SERVICIOS', 'INCOME', 4, 'C', 1, '413510', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '4140', 'HOTELES Y RESTAURANTES', 'INCOME', 3, 'C', 0, '4140', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '414005', 'HOTELERÍA', 'INCOME', 4, 'C', 1, '414005', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '414010', 'RESTAURANTES', 'INCOME', 4, 'C', 1, '414010', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '4145', 'TRANSPORTE, ALMACENAMIENTO Y COMUNICACIONES', 'INCOME', 3, 'C', 0, '4145', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '414505', 'SERVICIO DE TRANSPORTE', 'INCOME', 4, 'C', 1, '414505', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '4155', 'ACTIVIDADES INMOBILIARIAS, EMPRESARIALES Y DE ALQUILER', 'INCOME', 3, 'C', 0, '4155', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '415505', 'ARRENDAMIENTOS DE BIENES INMUEBLES', 'INCOME', 4, 'C', 1, '415505', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '415510', 'ARRENDAMIENTOS DE BIENES MUEBLES', 'INCOME', 4, 'C', 1, '415510', 1),

-- Nivel 2 - NO OPERACIONALES
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '42', 'NO OPERACIONALES', 'INCOME', 2, 'C', 0, '42', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '4210', 'FINANCIEROS', 'INCOME', 3, 'C', 0, '4210', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '421005', 'INTERESES', 'INCOME', 4, 'C', 1, '421005', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '421040', 'DIFERENCIA EN CAMBIO', 'INCOME', 4, 'C', 1, '421040', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '4250', 'RECUPERACIONES', 'INCOME', 3, 'C', 0, '4250', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '425050', 'REINTEGRO DE OTROS COSTOS Y GASTOS', 'INCOME', 4, 'C', 1, '425050', 1),

-- CLASE 5: GASTOS
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '5', 'GASTOS', 'EXPENSE', 1, 'D', 0, '5', 1),

-- Nivel 2 - OPERACIONALES DE ADMINISTRACIÓN
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '51', 'OPERACIONALES DE ADMINISTRACIÓN', 'EXPENSE', 2, 'D', 0, '51', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '5105', 'GASTOS DE PERSONAL', 'EXPENSE', 3, 'D', 0, '5105', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '510506', 'SUELDOS', 'EXPENSE', 4, 'D', 1, '510506', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '510515', 'HORAS EXTRAS Y RECARGOS', 'EXPENSE', 4, 'D', 1, '510515', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '510518', 'COMISIONES', 'EXPENSE', 4, 'D', 1, '510518', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '510520', 'AUXILIO DE TRANSPORTE', 'EXPENSE', 4, 'D', 1, '510520', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '510527', 'CESANTÍAS', 'EXPENSE', 4, 'D', 1, '510527', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '510530', 'INTERESES SOBRE CESANTÍAS', 'EXPENSE', 4, 'D', 1, '510530', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '510533', 'PRIMA DE SERVICIOS', 'EXPENSE', 4, 'D', 1, '510533', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '510536', 'VACACIONES', 'EXPENSE', 4, 'D', 1, '510536', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '510568', 'APORTES A ADMINISTRADORAS DE RIESGOS PROFESIONALES', 'EXPENSE', 4, 'D', 1, '510568', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '510569', 'APORTES A ENTIDADES PROMOTORAS DE SALUD', 'EXPENSE', 4, 'D', 1, '510569', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '510570', 'APORTES A FONDOS DE PENSIONES Y/O CESANTÍAS', 'EXPENSE', 4, 'D', 1, '510570', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '510572', 'APORTES CAJAS DE COMPENSACIÓN FAMILIAR', 'EXPENSE', 4, 'D', 1, '510572', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '510575', 'APORTES ICBF', 'EXPENSE', 4, 'D', 1, '510575', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '510578', 'APORTES SENA', 'EXPENSE', 4, 'D', 1, '510578', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '5110', 'HONORARIOS', 'EXPENSE', 3, 'D', 0, '5110', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '511005', 'JUNTA DIRECTIVA', 'EXPENSE', 4, 'D', 1, '511005', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '511010', 'REVISORÍA FISCAL', 'EXPENSE', 4, 'D', 1, '511010', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '511015', 'AUDITORÍA EXTERNA', 'EXPENSE', 4, 'D', 1, '511015', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '511025', 'ASESORÍA JURÍDICA', 'EXPENSE', 4, 'D', 1, '511025', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '511030', 'ASESORÍA FINANCIERA', 'EXPENSE', 4, 'D', 1, '511030', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '511035', 'ASESORÍA TÉCNICA', 'EXPENSE', 4, 'D', 1, '511035', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '511095', 'OTROS', 'EXPENSE', 4, 'D', 1, '511095', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '5115', 'IMPUESTOS', 'EXPENSE', 3, 'D', 0, '5115', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '511505', 'INDUSTRIA Y COMERCIO', 'EXPENSE', 4, 'D', 1, '511505', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '511515', 'A LA PROPIEDAD RAÍZ', 'EXPENSE', 4, 'D', 1, '511515', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '511540', 'DE VEHÍCULOS', 'EXPENSE', 4, 'D', 1, '511540', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '511595', 'OTROS', 'EXPENSE', 4, 'D', 1, '511595', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '5120', 'ARRENDAMIENTOS', 'EXPENSE', 3, 'D', 0, '5120', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '512010', 'CONSTRUCCIONES Y EDIFICACIONES', 'EXPENSE', 4, 'D', 1, '512010', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '512020', 'EQUIPO DE OFICINA', 'EXPENSE', 4, 'D', 1, '512020', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '512025', 'EQUIPO DE COMPUTACIÓN Y COMUNICACIÓN', 'EXPENSE', 4, 'D', 1, '512025', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '5125', 'CONTRIBUCIONES Y AFILIACIONES', 'EXPENSE', 3, 'D', 0, '5125', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '512505', 'CONTRIBUCIONES', 'EXPENSE', 4, 'D', 1, '512505', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '512510', 'AFILIACIONES Y SOSTENIMIENTO', 'EXPENSE', 4, 'D', 1, '512510', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '5130', 'SEGUROS', 'EXPENSE', 3, 'D', 0, '5130', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '513005', 'MANEJO', 'EXPENSE', 4, 'D', 1, '513005', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '513010', 'CUMPLIMIENTO', 'EXPENSE', 4, 'D', 1, '513010', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '513025', 'INCENDIO', 'EXPENSE', 4, 'D', 1, '513025', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '513095', 'OTROS', 'EXPENSE', 4, 'D', 1, '513095', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '5135', 'SERVICIOS', 'EXPENSE', 3, 'D', 0, '5135', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '513505', 'ASEO Y VIGILANCIA', 'EXPENSE', 4, 'D', 1, '513505', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '513510', 'TEMPORALES', 'EXPENSE', 4, 'D', 1, '513510', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '513515', 'ASISTENCIA TÉCNICA', 'EXPENSE', 4, 'D', 1, '513515', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '513520', 'PROCESAMIENTO ELECTRÓNICO DE DATOS', 'EXPENSE', 4, 'D', 1, '513520', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '513525', 'ACUEDUCTO Y ALCANTARILLADO', 'EXPENSE', 4, 'D', 1, '513525', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '513530', 'ENERGÍA ELÉCTRICA', 'EXPENSE', 4, 'D', 1, '513530', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '513535', 'TELÉFONO', 'EXPENSE', 4, 'D', 1, '513535', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '513540', 'CORREO, PORTES Y TELEGRAMAS', 'EXPENSE', 4, 'D', 1, '513540', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '513550', 'TRANSPORTE, FLETES Y ACARREOS', 'EXPENSE', 4, 'D', 1, '513550', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '513595', 'OTROS', 'EXPENSE', 4, 'D', 1, '513595', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '5140', 'GASTOS LEGALES', 'EXPENSE', 3, 'D', 0, '5140', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '514005', 'NOTARIALES', 'EXPENSE', 4, 'D', 1, '514005', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '514010', 'REGISTRO MERCANTIL', 'EXPENSE', 4, 'D', 1, '514010', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '514015', 'TRÁMITES Y LICENCIAS', 'EXPENSE', 4, 'D', 1, '514015', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '514095', 'OTROS', 'EXPENSE', 4, 'D', 1, '514095', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '5145', 'MANTENIMIENTO Y REPARACIONES', 'EXPENSE', 3, 'D', 0, '5145', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '514510', 'CONSTRUCCIONES Y EDIFICACIONES', 'EXPENSE', 4, 'D', 1, '514510', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '514515', 'MAQUINARIA Y EQUIPO', 'EXPENSE', 4, 'D', 1, '514515', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '514520', 'EQUIPO DE OFICINA', 'EXPENSE', 4, 'D', 1, '514520', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '514525', 'EQUIPO DE COMPUTACIÓN Y COMUNICACIÓN', 'EXPENSE', 4, 'D', 1, '514525', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '514535', 'FLOTA Y EQUIPO DE TRANSPORTE', 'EXPENSE', 4, 'D', 1, '514535', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '5150', 'ADECUACIÓN E INSTALACIÓN', 'EXPENSE', 3, 'D', 0, '5150', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '515005', 'INSTALACIONES ELÉCTRICAS', 'EXPENSE', 4, 'D', 1, '515005', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '515015', 'REPARACIONES LOCATIVAS', 'EXPENSE', 4, 'D', 1, '515015', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '515095', 'OTROS', 'EXPENSE', 4, 'D', 1, '515095', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '5155', 'GASTOS DE VIAJE', 'EXPENSE', 3, 'D', 0, '5155', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '515505', 'ALOJAMIENTO Y MANUTENCIÓN', 'EXPENSE', 4, 'D', 1, '515505', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '515515', 'PASAJES AÉREOS', 'EXPENSE', 4, 'D', 1, '515515', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '515520', 'PASAJES TERRESTRES', 'EXPENSE', 4, 'D', 1, '515520', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '515595', 'OTROS', 'EXPENSE', 4, 'D', 1, '515595', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '5160', 'DEPRECIACIONES', 'EXPENSE', 3, 'D', 0, '5160', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '516005', 'CONSTRUCCIONES Y EDIFICACIONES', 'EXPENSE', 4, 'D', 1, '516005', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '516010', 'MAQUINARIA Y EQUIPO', 'EXPENSE', 4, 'D', 1, '516010', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '516015', 'EQUIPO DE OFICINA', 'EXPENSE', 4, 'D', 1, '516015', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '516020', 'EQUIPO DE COMPUTACIÓN Y COMUNICACIÓN', 'EXPENSE', 4, 'D', 1, '516020', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '516035', 'FLOTA Y EQUIPO DE TRANSPORTE', 'EXPENSE', 4, 'D', 1, '516035', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '5195', 'DIVERSOS', 'EXPENSE', 3, 'D', 0, '5195', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '519505', 'COMISIONES', 'EXPENSE', 4, 'D', 1, '519505', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '519510', 'LIBROS, SUSCRIPCIONES, PERIÓDICOS Y REVISTAS', 'EXPENSE', 4, 'D', 1, '519510', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '519520', 'GASTOS DE REPRESENTACIÓN Y RELACIONES PÚBLICAS', 'EXPENSE', 4, 'D', 1, '519520', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '519525', 'ELEMENTOS DE ASEO Y CAFETERÍA', 'EXPENSE', 4, 'D', 1, '519525', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '519530', 'ÚTILES, PAPELERÍA Y FOTOCOPIAS', 'EXPENSE', 4, 'D', 1, '519530', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '519535', 'COMBUSTIBLES Y LUBRICANTES', 'EXPENSE', 4, 'D', 1, '519535', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '519540', 'ENVASES Y EMPAQUES', 'EXPENSE', 4, 'D', 1, '519540', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '519545', 'TAXIS Y BUSES', 'EXPENSE', 4, 'D', 1, '519545', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '519565', 'PARQUEADEROS', 'EXPENSE', 4, 'D', 1, '519565', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '519595', 'OTROS', 'EXPENSE', 4, 'D', 1, '519595', 1),

-- Nivel 2 - OPERACIONALES DE VENTAS
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '52', 'OPERACIONALES DE VENTAS', 'EXPENSE', 2, 'D', 0, '52', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '5205', 'GASTOS DE PERSONAL', 'EXPENSE', 3, 'D', 0, '5205', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '520506', 'SUELDOS', 'EXPENSE', 4, 'D', 1, '520506', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '520518', 'COMISIONES', 'EXPENSE', 4, 'D', 1, '520518', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '5295', 'DIVERSOS', 'EXPENSE', 3, 'D', 0, '5295', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '529505', 'COMISIONES', 'EXPENSE', 4, 'D', 1, '529505', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '529515', 'PUBLICIDAD, PROPAGANDA Y PROMOCIÓN', 'EXPENSE', 4, 'D', 1, '529515', 1),

-- Nivel 2 - NO OPERACIONALES
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '53', 'NO OPERACIONALES', 'EXPENSE', 2, 'D', 0, '53', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '5305', 'FINANCIEROS', 'EXPENSE', 3, 'D', 0, '5305', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '530505', 'GASTOS BANCARIOS', 'EXPENSE', 4, 'D', 1, '530505', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '530515', 'COMISIONES', 'EXPENSE', 4, 'D', 1, '530515', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '530520', 'INTERESES', 'EXPENSE', 4, 'D', 1, '530520', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '530535', 'DIFERENCIA EN CAMBIO', 'EXPENSE', 4, 'D', 1, '530535', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '5395', 'GASTOS DIVERSOS', 'EXPENSE', 3, 'D', 0, '5395', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '539520', 'MULTAS, SANCIONES Y LITIGIOS', 'EXPENSE', 4, 'D', 1, '539520', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '539595', 'OTROS', 'EXPENSE', 4, 'D', 1, '539595', 1),

-- Nivel 2 - IMPUESTO DE RENTA
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '54', 'IMPUESTO DE RENTA Y COMPLEMENTARIOS', 'EXPENSE', 2, 'D', 0, '54', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '5405', 'IMPUESTO DE RENTA Y COMPLEMENTARIOS', 'EXPENSE', 3, 'D', 0, '5405', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '540505', 'IMPUESTO DE RENTA Y COMPLEMENTARIOS', 'EXPENSE', 4, 'D', 1, '540505', 1),

-- CLASE 6: COSTOS DE VENTAS
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '6', 'COSTOS DE VENTAS', 'COST', 1, 'D', 0, '6', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '61', 'COSTO DE VENTAS Y DE PRESTACIÓN DE SERVICIOS', 'COST', 2, 'D', 0, '61', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '6135', 'COMERCIO AL POR MAYOR Y AL POR MENOR', 'COST', 3, 'D', 0, '6135', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '613505', 'COSTO DE VENTAS', 'COST', 4, 'D', 1, '613505', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '6155', 'ACTIVIDADES INMOBILIARIAS, EMPRESARIALES Y DE ALQUILER', 'COST', 3, 'D', 0, '6155', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '615505', 'COSTO DE SERVICIOS', 'COST', 4, 'D', 1, '615505', 1),

-- CLASE 7: COSTOS DE PRODUCCIÓN
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '7', 'COSTOS DE PRODUCCIÓN O DE OPERACIÓN', 'COST', 1, 'D', 0, '7', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '71', 'MATERIA PRIMA', 'COST', 2, 'D', 0, '71', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '7101', 'MATERIA PRIMA', 'COST', 3, 'D', 0, '7101', 1),
(lower(hex(randomblob(16))), datetime('now'), datetime('now'), (SELECT id FROM organizations LIMIT 1), '710101', 'MATERIA PRIMA DIRECTA', 'COST', 4, 'D', 1, '710101', 1);

-- Verificar el total de cuentas insertadas
SELECT COUNT(*) as total_cuentas FROM accounts;