# Propuesta Técnica: Motor Contable Cloud-Native

## Resumen Ejecutivo

Este documento presenta la propuesta técnica para el desarrollo de un motor contable cloud-native diseñado para procesar hasta 5 millones de comprobantes diarios, con soporte multi-país y multi-tenant. La solución se basa en Go, PostgreSQL y tecnologías modernas de contenedores, con un enfoque especial en el uso de DSLs (Domain Specific Languages) para la configuración dinámica de reglas contables.

### Estado Actual del Proyecto

✅ **Fase 0 - Prototipado (COMPLETADA)**
- Mock funcional completo con interfaz web responsive
- Sistema de API centralizado con catálogo de endpoints
- 13 mocks de API implementados
- Dashboard interactivo con gráficos Chart.js
- Editor DSL con sintaxis y ejemplos
- Documentación técnica completa

🚧 **Próxima Fase - MVP Backend**
- Implementación del backend en Go/Fiber
- Integración con go-dsl
- APIs REST funcionales
- Base de datos PostgreSQL

---

## 🧩 1. Objetivo General y Alcance

### 1.1 Objetivo Principal

Desarrollar un motor contable empresarial que permita:

- **Procesamiento masivo**: Capacidad de procesar y contabilizar hasta 5 millones de comprobantes diarios con alta disponibilidad.
- **Generación automática**: Crear asientos contables según plantillas dinámicas configurables por país y tipo de comprobante.
- **Libros contables completos**: Generar automáticamente Libro Diario, Libro Mayor, Balance General, Estado de Resultados, Balance de Comprobación y reportes auxiliares.
- **Multi-tenant con aislamiento**: Soporte para múltiples empresas con aislamiento lógico completo de datos.
- **Multi-país**: Cumplimiento normativo para Colombia, México, Chile, Ecuador, Uruguay y Perú.

### 1.2 Alcance Funcional

#### Módulos principales:
1. **Motor de Ingesta**: Recepción y validación de comprobantes en múltiples formatos (JSON, XML, CSV).
2. **Motor de Contabilización**: Generación automática de asientos según plantillas DSL.
3. **Motor de Consolidación**: Generación de libros contables y estados financieros.
4. **Sistema de Auditoría**: Trazabilidad completa de todas las operaciones.
5. **API Gateway**: Exposición de servicios REST/GraphQL para integración.
6. **Dashboard Administrativo**: Interfaz web para configuración y monitoreo.

### 1.3 Requerimientos No Funcionales

- **Performance**: Procesamiento de 5M comprobantes/día (≈58 comprobantes/segundo).
- **Disponibilidad**: 99.9% uptime (máximo 8.76 horas de downtime anual).
- **Escalabilidad**: Horizontal mediante contenedores Docker en DigitalOcean.
- **Seguridad**: Encriptación en tránsito (TLS 1.3) y en reposo (AES-256).
- **Cumplimiento**: NIIF, NIC, normativas locales por país.

---

## 🛠️ 2. Stack Tecnológico

### 2.1 Core Technologies

#### Backend
- **Lenguaje**: Go 1.21+ 
  - Alto rendimiento para procesamiento concurrente
  - Gestión eficiente de memoria
  - Excelente soporte para microservicios

- **Framework Web**: Fiber v2
  - Framework inspirado en Express.js
  - Rendimiento excepcional (basado en Fasthttp)
  - Middleware rico y extensible

- **Motor DSL**: go-dsl (github.com/arturoeanton/go-dsl)
  - Procesamiento de reglas contables dinámicas
  - Parser con soporte para left recursion
  - Sistema de acciones extensible

- **Framework Web**: Fiber v2 (https://gofiber.io/)
  - Basado en Fasthttp, hasta 10x más rápido que net/http
  - API similar a Express.js, fácil adopción
  - Middleware robusto para autenticación, CORS, rate limiting

- **Base de Datos**: PostgreSQL 15+
  - ACID compliance crítico para datos contables
  - Particionamiento nativo para escalabilidad
  - JSONB para metadatos flexibles
  - Extensiones: pg_partman, timescaledb para series temporales

#### Infraestructura
- **Contenedores**: Docker + Docker Compose
  - Imágenes multi-stage para optimización
  - Health checks integrados
  - Secretos manejados via Docker Secrets

- **Orquestación**: Kubernetes (DigitalOcean Managed K8s)
  - Auto-scaling horizontal basado en CPU/memoria
  - Rolling updates sin downtime
  - Service mesh con Istio para observabilidad

#### DSL Engine
- **go-dsl** (https://github.com/arturoeanton/go-dsl)
  - Motor de DSL para reglas contables dinámicas
  - Parser con soporte para gramáticas complejas
  - Integración nativa con Go

### 2.2 Herramientas de Desarrollo

- **Testing**: Go testing + testify + gomock
- **CI/CD**: GitHub Actions + ArgoCD
- **Monitoreo**: Prometheus + Grafana + Jaeger
- **Logging**: Zap + Loki
- **Message Queue**: NATS JetStream (para procesamiento asíncrono)
- **Cache**: Redis Cluster
- **Object Storage**: DigitalOcean Spaces (para archivos adjuntos)

---

## 🧮 3. Modelo de Datos PostgreSQL

### 3.1 Diseño de Base de Datos

#### Esquema Multi-Tenant
```sql
-- Esquema principal: accounting_engine

-- Tabla de organizaciones (tenants)
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(200) NOT NULL,
    country_code CHAR(2) NOT NULL,
    tax_id VARCHAR(50) NOT NULL,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_active BOOLEAN DEFAULT true
);

-- Catálogo de cuentas contables
CREATE TABLE chart_of_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    account_code VARCHAR(50) NOT NULL,
    name VARCHAR(200) NOT NULL,
    type VARCHAR(20) NOT NULL, -- ASSET, LIABILITY, EQUITY, INCOME, EXPENSE
    nature CHAR(1) NOT NULL, -- D: Debit, C: Credit
    level INTEGER NOT NULL,
    parent_id UUID REFERENCES chart_of_accounts(id),
    is_detail BOOLEAN DEFAULT false,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_active BOOLEAN DEFAULT true,
    UNIQUE(organization_id, account_code)
);

-- Comprobantes (documentos fuente)
CREATE TABLE vouchers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    voucher_number VARCHAR(50) NOT NULL,
    voucher_type VARCHAR(50) NOT NULL, -- INVOICE, RECEIPT, PAYMENT, etc
    voucher_date DATE NOT NULL,
    description TEXT,
    total_amount DECIMAL(20,4) NOT NULL,
    currency_code CHAR(3) NOT NULL DEFAULT 'USD',
    exchange_rate DECIMAL(10,6) DEFAULT 1.0,
    source_system VARCHAR(100),
    external_ref VARCHAR(100),
    metadata JSONB DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'PENDING',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    processed_at TIMESTAMPTZ,
    UNIQUE(organization_id, voucher_number)
) PARTITION BY RANGE (voucher_date);

-- Particiones mensuales para vouchers
CREATE TABLE vouchers_2024_01 PARTITION OF vouchers
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
-- ... más particiones generadas dinámicamente

-- Asientos contables (journal entries)
CREATE TABLE journal_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    entry_number BIGSERIAL,
    entry_date DATE NOT NULL,
    voucher_id UUID REFERENCES vouchers(id),
    description TEXT NOT NULL,
    entry_type VARCHAR(20) NOT NULL, -- STANDARD, ADJUSTMENT, CLOSING
    period VARCHAR(7) NOT NULL, -- YYYY-MM
    status VARCHAR(20) DEFAULT 'DRAFT',
    created_by UUID NOT NULL,
    approved_by UUID,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    approved_at TIMESTAMPTZ,
    is_reversed BOOLEAN DEFAULT false,
    reversal_id UUID REFERENCES journal_entries(id),
    UNIQUE(organization_id, entry_number)
) PARTITION BY RANGE (entry_date);

-- Líneas de asientos
CREATE TABLE journal_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    journal_entry_id UUID NOT NULL REFERENCES journal_entries(id),
    line_number INTEGER NOT NULL,
    account_id UUID NOT NULL REFERENCES chart_of_accounts(id),
    debit_amount DECIMAL(20,4) DEFAULT 0,
    credit_amount DECIMAL(20,4) DEFAULT 0,
    description TEXT,
    cost_center_id UUID,
    project_id UUID,
    metadata JSONB DEFAULT '{}',
    CHECK (debit_amount >= 0 AND credit_amount >= 0),
    CHECK (debit_amount = 0 OR credit_amount = 0)
);

-- Plantillas DSL para contabilización
CREATE TABLE accounting_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    template_code VARCHAR(50) NOT NULL,
    voucher_type VARCHAR(50) NOT NULL,
    country_code CHAR(2) NOT NULL,
    dsl_content TEXT NOT NULL, -- DSL definition
    compiled_dsl JSONB, -- Compiled AST for performance
    version INTEGER DEFAULT 1,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID NOT NULL,
    UNIQUE(organization_id, template_code, version)
);

-- Reglas fiscales por país
CREATE TABLE fiscal_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    country_code CHAR(2) NOT NULL,
    rule_code VARCHAR(50) NOT NULL,
    rule_type VARCHAR(50) NOT NULL,
    dsl_content TEXT NOT NULL,
    effective_date DATE NOT NULL,
    expiry_date DATE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(country_code, rule_code, effective_date)
);

-- Libros contables generados
CREATE TABLE accounting_books (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    book_type VARCHAR(50) NOT NULL, -- JOURNAL, LEDGER, TRIAL_BALANCE, etc
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    generation_date TIMESTAMPTZ DEFAULT NOW(),
    file_url TEXT,
    status VARCHAR(20) DEFAULT 'GENERATING',
    metadata JSONB DEFAULT '{}',
    INDEX idx_books_org_period (organization_id, period_start, period_end)
);

-- Índices para performance
CREATE INDEX idx_vouchers_org_date ON vouchers(organization_id, voucher_date);
CREATE INDEX idx_vouchers_status ON vouchers(status) WHERE status = 'PENDING';
CREATE INDEX idx_journal_entries_org_period ON journal_entries(organization_id, period);
CREATE INDEX idx_journal_lines_account ON journal_lines(account_id);
CREATE INDEX idx_chart_accounts_org_code ON chart_of_accounts(organization_id, account_code);

-- Vistas materializadas para reportes
CREATE MATERIALIZED VIEW account_balances AS
SELECT 
    jl.account_id,
    je.organization_id,
    je.period,
    SUM(jl.debit_amount) as total_debit,
    SUM(jl.credit_amount) as total_credit,
    SUM(jl.debit_amount) - SUM(jl.credit_amount) as balance
FROM journal_lines jl
JOIN journal_entries je ON jl.journal_entry_id = je.id
WHERE je.status = 'POSTED'
GROUP BY jl.account_id, je.organization_id, je.period;

CREATE INDEX idx_account_balances ON account_balances(organization_id, account_id, period);
```

### 3.2 Estrategias de Escalabilidad

#### Particionamiento
- **Vouchers**: Particionamiento mensual por fecha
- **Journal Entries**: Particionamiento mensual por fecha
- **Mantenimiento automático**: pg_partman para creación/eliminación de particiones

#### Índices Especializados
- **B-tree**: Para búsquedas exactas y rangos
- **Hash**: Para joins frecuentes en organization_id
- **GIN**: Para búsquedas en campos JSONB
- **BRIN**: Para datos temporales en particiones grandes

#### Optimizaciones
- **Connection Pooling**: PgBouncer con pool de 100-200 conexiones
- **Read Replicas**: Para reportes y consultas pesadas
- **Prepared Statements**: Para queries frecuentes
- **Batch Processing**: Inserción bulk de hasta 1000 registros

### 3.3 Estimación de Volumen

Para 5M comprobantes/día:
- **Vouchers**: 5M registros/día = 150M/mes
- **Journal Entries**: ~10M registros/día (2 asientos promedio por comprobante)
- **Journal Lines**: ~40M registros/día (4 líneas promedio por asiento)
- **Almacenamiento estimado**: ~500GB/mes con índices y metadata

### 3.4 Carga Inicial del Catálogo de Cuentas

#### Script para Desarrollo (con datos de prueba)

```sql
-- ==============================================
-- SCRIPT DE DESARROLLO: Datos de prueba completos
-- ==============================================

-- Limpiar datos existentes (SOLO DESARROLLO)
TRUNCATE TABLE journal_lines CASCADE;
TRUNCATE TABLE journal_entries CASCADE;
TRUNCATE TABLE vouchers CASCADE;
TRUNCATE TABLE chart_of_accounts CASCADE;
TRUNCATE TABLE organizations CASCADE;

-- Crear organizaciones de prueba
INSERT INTO organizations (id, code, name, country_code, tax_id, settings) VALUES
('11111111-1111-1111-1111-111111111111', 'DEMO-CO', 'Empresa Demo Colombia', 'CO', '900123456-7', 
 '{"currency": "COP", "fiscal_year": "calendar", "decimal_places": 2}'),
('22222222-2222-2222-2222-222222222222', 'DEMO-MX', 'Empresa Demo México', 'MX', 'EDM850101ABC', 
 '{"currency": "MXN", "fiscal_year": "calendar", "decimal_places": 2}'),
('33333333-3333-3333-3333-333333333333', 'DEMO-CL', 'Empresa Demo Chile', 'CL', '76.123.456-7', 
 '{"currency": "CLP", "fiscal_year": "calendar", "decimal_places": 0}');

-- Función auxiliar para crear cuentas jerárquicas
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
    PERFORM create_account(v_org_id, '1330', 'ANTICIPOS Y AVANCES', 'ASSET', 'D', '13');
    PERFORM create_account(v_org_id, '1330.01', 'Anticipos a Proveedores', 'ASSET', 'D', '1330');
    PERFORM create_account(v_org_id, '1355', 'ANTICIPO DE IMPUESTOS', 'ASSET', 'D', '13');
    PERFORM create_account(v_org_id, '1355.25', 'Retención en la Fuente', 'ASSET', 'D', '1355');
    PERFORM create_account(v_org_id, '1365', 'CUENTAS POR COBRAR A EMPLEADOS', 'ASSET', 'D', '13');
    PERFORM create_account(v_org_id, '1365.01', 'Préstamos a Empleados', 'ASSET', 'D', '1365');
    
    -- Grupo 14: Inventarios
    PERFORM create_account(v_org_id, '14', 'INVENTARIOS', 'ASSET', 'D', '1');
    PERFORM create_account(v_org_id, '1435', 'MERCANCÍAS NO FABRICADAS', 'ASSET', 'D', '14');
    PERFORM create_account(v_org_id, '1435.01', 'Mercancías en Almacén', 'ASSET', 'D', '1435');
    
    -- Grupo 15: Propiedad, Planta y Equipo
    PERFORM create_account(v_org_id, '15', 'PROPIEDAD PLANTA Y EQUIPO', 'ASSET', 'D', '1');
    PERFORM create_account(v_org_id, '1520', 'MAQUINARIA Y EQUIPO', 'ASSET', 'D', '15');
    PERFORM create_account(v_org_id, '1520.01', 'Maquinaria', 'ASSET', 'D', '1520');
    PERFORM create_account(v_org_id, '1524', 'EQUIPO DE OFICINA', 'ASSET', 'D', '15');
    PERFORM create_account(v_org_id, '1524.01', 'Muebles y Enseres', 'ASSET', 'D', '1524');
    PERFORM create_account(v_org_id, '1528', 'EQUIPO DE COMPUTACIÓN', 'ASSET', 'D', '15');
    PERFORM create_account(v_org_id, '1528.01', 'Equipos de Cómputo', 'ASSET', 'D', '1528');
    PERFORM create_account(v_org_id, '1592', 'DEPRECIACIÓN ACUMULADA', 'ASSET', 'C', '15');
    PERFORM create_account(v_org_id, '1592.01', 'Dep. Maquinaria y Equipo', 'ASSET', 'C', '1592');
    PERFORM create_account(v_org_id, '1592.02', 'Dep. Equipo de Oficina', 'ASSET', 'C', '1592');
    PERFORM create_account(v_org_id, '1592.03', 'Dep. Equipo de Computación', 'ASSET', 'C', '1592');
    
    -- CLASE 2: PASIVOS
    PERFORM create_account(v_org_id, '2', 'PASIVOS', 'LIABILITY', 'C');
    
    -- Grupo 22: Proveedores
    PERFORM create_account(v_org_id, '22', 'PROVEEDORES', 'LIABILITY', 'C', '2');
    PERFORM create_account(v_org_id, '2205', 'PROVEEDORES NACIONALES', 'LIABILITY', 'C', '22');
    PERFORM create_account(v_org_id, '2205.01', 'Proveedores Varios', 'LIABILITY', 'C', '2205');
    
    -- Grupo 23: Cuentas por Pagar
    PERFORM create_account(v_org_id, '23', 'CUENTAS POR PAGAR', 'LIABILITY', 'C', '2');
    PERFORM create_account(v_org_id, '2335', 'COSTOS Y GASTOS POR PAGAR', 'LIABILITY', 'C', '23');
    PERFORM create_account(v_org_id, '2335.01', 'Servicios Públicos', 'LIABILITY', 'C', '2335');
    PERFORM create_account(v_org_id, '2335.02', 'Honorarios', 'LIABILITY', 'C', '2335');
    PERFORM create_account(v_org_id, '2365', 'RETENCIÓN EN LA FUENTE', 'LIABILITY', 'C', '23');
    PERFORM create_account(v_org_id, '2365.01', 'Ret. Honorarios 11%', 'LIABILITY', 'C', '2365');
    PERFORM create_account(v_org_id, '2365.02', 'Ret. Servicios 4%', 'LIABILITY', 'C', '2365');
    PERFORM create_account(v_org_id, '2365.03', 'Ret. Compras 2.5%', 'LIABILITY', 'C', '2365');
    
    -- Grupo 24: Impuestos por Pagar
    PERFORM create_account(v_org_id, '24', 'IMPUESTOS GRAVÁMENES Y TASAS', 'LIABILITY', 'C', '2');
    PERFORM create_account(v_org_id, '2408', 'IVA POR PAGAR', 'LIABILITY', 'C', '24');
    PERFORM create_account(v_org_id, '2408.01', 'IVA 19% por Pagar', 'LIABILITY', 'C', '2408');
    PERFORM create_account(v_org_id, '2408.02', 'IVA 5% por Pagar', 'LIABILITY', 'C', '2408');
    PERFORM create_account(v_org_id, '2412', 'ICA POR PAGAR', 'LIABILITY', 'C', '24');
    PERFORM create_account(v_org_id, '2412.01', 'Industria y Comercio', 'LIABILITY', 'C', '2412');
    
    -- Grupo 25: Obligaciones Laborales
    PERFORM create_account(v_org_id, '25', 'OBLIGACIONES LABORALES', 'LIABILITY', 'C', '2');
    PERFORM create_account(v_org_id, '2505', 'SALARIOS POR PAGAR', 'LIABILITY', 'C', '25');
    PERFORM create_account(v_org_id, '2505.01', 'Sueldos por Pagar', 'LIABILITY', 'C', '2505');
    PERFORM create_account(v_org_id, '2510', 'CESANTÍAS CONSOLIDADAS', 'LIABILITY', 'C', '25');
    PERFORM create_account(v_org_id, '2510.01', 'Cesantías por Pagar', 'LIABILITY', 'C', '2510');
    
    -- CLASE 3: PATRIMONIO
    PERFORM create_account(v_org_id, '3', 'PATRIMONIO', 'EQUITY', 'C');
    
    -- Grupo 31: Capital Social
    PERFORM create_account(v_org_id, '31', 'CAPITAL SOCIAL', 'EQUITY', 'C', '3');
    PERFORM create_account(v_org_id, '3105', 'CAPITAL SUSCRITO Y PAGADO', 'EQUITY', 'C', '31');
    PERFORM create_account(v_org_id, '3105.01', 'Capital Social', 'EQUITY', 'C', '3105');
    
    -- Grupo 36: Resultados del Ejercicio
    PERFORM create_account(v_org_id, '36', 'RESULTADOS DEL EJERCICIO', 'EQUITY', 'C', '3');
    PERFORM create_account(v_org_id, '3605', 'UTILIDAD DEL EJERCICIO', 'EQUITY', 'C', '36');
    PERFORM create_account(v_org_id, '3605.01', 'Utilidad del Período', 'EQUITY', 'C', '3605');
    PERFORM create_account(v_org_id, '3610', 'PÉRDIDA DEL EJERCICIO', 'EQUITY', 'D', '36');
    PERFORM create_account(v_org_id, '3610.01', 'Pérdida del Período', 'EQUITY', 'D', '3610');
    
    -- CLASE 4: INGRESOS
    PERFORM create_account(v_org_id, '4', 'INGRESOS', 'INCOME', 'C');
    
    -- Grupo 41: Operacionales
    PERFORM create_account(v_org_id, '41', 'OPERACIONALES', 'INCOME', 'C', '4');
    PERFORM create_account(v_org_id, '4135', 'COMERCIO AL POR MAYOR Y MENOR', 'INCOME', 'C', '41');
    PERFORM create_account(v_org_id, '4135.01', 'Venta de Mercancías', 'INCOME', 'C', '4135');
    PERFORM create_account(v_org_id, '4135.02', 'Venta de Servicios', 'INCOME', 'C', '4135');
    PERFORM create_account(v_org_id, '4175', 'DEVOLUCIONES EN VENTAS', 'INCOME', 'D', '41');
    PERFORM create_account(v_org_id, '4175.01', 'Devoluciones en Ventas', 'INCOME', 'D', '4175');
    
    -- Grupo 42: No Operacionales
    PERFORM create_account(v_org_id, '42', 'NO OPERACIONALES', 'INCOME', 'C', '4');
    PERFORM create_account(v_org_id, '4210', 'FINANCIEROS', 'INCOME', 'C', '42');
    PERFORM create_account(v_org_id, '4210.01', 'Intereses Bancarios', 'INCOME', 'C', '4210');
    
    -- CLASE 5: GASTOS
    PERFORM create_account(v_org_id, '5', 'GASTOS', 'EXPENSE', 'D');
    
    -- Grupo 51: Operacionales de Administración
    PERFORM create_account(v_org_id, '51', 'OPERACIONALES DE ADMINISTRACIÓN', 'EXPENSE', 'D', '5');
    PERFORM create_account(v_org_id, '5105', 'GASTOS DE PERSONAL', 'EXPENSE', 'D', '51');
    PERFORM create_account(v_org_id, '5105.01', 'Sueldos', 'EXPENSE', 'D', '5105');
    PERFORM create_account(v_org_id, '5105.02', 'Prestaciones Sociales', 'EXPENSE', 'D', '5105');
    PERFORM create_account(v_org_id, '5105.03', 'Aportes Seguridad Social', 'EXPENSE', 'D', '5105');
    PERFORM create_account(v_org_id, '5111', 'HONORARIOS', 'EXPENSE', 'D', '51');
    PERFORM create_account(v_org_id, '5111.01', 'Asesoría Contable', 'EXPENSE', 'D', '5111');
    PERFORM create_account(v_org_id, '5111.02', 'Asesoría Legal', 'EXPENSE', 'D', '5111');
    PERFORM create_account(v_org_id, '5115', 'IMPUESTOS', 'EXPENSE', 'D', '51');
    PERFORM create_account(v_org_id, '5115.01', 'Industria y Comercio', 'EXPENSE', 'D', '5115');
    PERFORM create_account(v_org_id, '5135', 'SERVICIOS', 'EXPENSE', 'D', '51');
    PERFORM create_account(v_org_id, '5135.01', 'Servicios Públicos', 'EXPENSE', 'D', '5135');
    PERFORM create_account(v_org_id, '5135.02', 'Telecomunicaciones', 'EXPENSE', 'D', '5135');
    PERFORM create_account(v_org_id, '5135.03', 'Internet', 'EXPENSE', 'D', '5135');
    
    -- Grupo 52: Operacionales de Ventas
    PERFORM create_account(v_org_id, '52', 'OPERACIONALES DE VENTAS', 'EXPENSE', 'D', '5');
    PERFORM create_account(v_org_id, '5205', 'GASTOS DE PERSONAL', 'EXPENSE', 'D', '52');
    PERFORM create_account(v_org_id, '5205.01', 'Sueldos Vendedores', 'EXPENSE', 'D', '5205');
    PERFORM create_account(v_org_id, '5205.02', 'Comisiones', 'EXPENSE', 'D', '5205');
    
    -- Grupo 53: No Operacionales
    PERFORM create_account(v_org_id, '53', 'NO OPERACIONALES', 'EXPENSE', 'D', '5');
    PERFORM create_account(v_org_id, '5305', 'FINANCIEROS', 'EXPENSE', 'D', '53');
    PERFORM create_account(v_org_id, '5305.01', 'Gastos Bancarios', 'EXPENSE', 'D', '5305');
    PERFORM create_account(v_org_id, '5305.02', 'Comisiones Bancarias', 'EXPENSE', 'D', '5305');
    
    -- CLASE 6: COSTOS DE VENTAS
    PERFORM create_account(v_org_id, '6', 'COSTOS DE VENTAS', 'EXPENSE', 'D');
    PERFORM create_account(v_org_id, '61', 'COSTO DE VENTAS Y PRESTACIÓN DE SERVICIOS', 'EXPENSE', 'D', '6');
    PERFORM create_account(v_org_id, '6135', 'COMERCIO AL POR MAYOR Y MENOR', 'EXPENSE', 'D', '61');
    PERFORM create_account(v_org_id, '6135.01', 'Costo de Mercancías Vendidas', 'EXPENSE', 'D', '6135');
END $$;

-- ==============================================
-- DATOS DE PRUEBA ADICIONALES
-- ==============================================

-- Crear algunos comprobantes de prueba
INSERT INTO vouchers (organization_id, voucher_number, voucher_type, voucher_date, 
                     description, total_amount, currency_code, status) VALUES
('11111111-1111-1111-1111-111111111111', 'FV-001-2024', 'INVOICE', '2024-01-15', 
 'Factura venta productos varios', 1190000, 'COP', 'PENDING'),
('11111111-1111-1111-1111-111111111111', 'FV-002-2024', 'INVOICE', '2024-01-16', 
 'Factura venta servicios consultoría', 5950000, 'COP', 'PENDING'),
('11111111-1111-1111-1111-111111111111', 'RC-001-2024', 'RECEIPT', '2024-01-20', 
 'Recibo de caja cliente FV-001', 1190000, 'COP', 'PENDING');

-- Crear plantillas DSL de ejemplo
INSERT INTO accounting_templates (organization_id, template_code, voucher_type, 
                                 country_code, dsl_content, version) VALUES
('11111111-1111-1111-1111-111111111111', 'INVOICE_SALE_CO', 'INVOICE', 'CO', 
'template invoice_sale_co {
  let subtotal = voucher.subtotal
  let iva_19 = voucher.taxes["iva_19"]
  let total = voucher.total
  
  entry {
    debit account("1305.01") amount(total) {
      description = "Factura " + voucher.number
    }
    credit account("4135.01") amount(subtotal) {
      description = "Venta de mercancías"
    }
    credit account("2408.01") amount(iva_19) {
      description = "IVA 19% por pagar"
    }
  }
}', 1);

-- Función para limpiar
DROP FUNCTION IF EXISTS create_account(UUID, VARCHAR, VARCHAR, VARCHAR, CHAR, VARCHAR);

-- Mensaje de confirmación
DO $$
BEGIN
    RAISE NOTICE 'Script de desarrollo completado. Creadas:';
    RAISE NOTICE '- 3 organizaciones de prueba';
    RAISE NOTICE '- % cuentas contables', (SELECT COUNT(*) FROM chart_of_accounts);
    RAISE NOTICE '- 3 comprobantes de ejemplo';
    RAISE NOTICE '- 1 plantilla DSL';
END $$;
```

#### Script para Producción (base limpia)

```sql
-- ==============================================
-- SCRIPT DE PRODUCCIÓN: Base limpia y estructura mínima
-- ==============================================

-- Verificar que no existan datos (seguridad para producción)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM organizations LIMIT 1) THEN
        RAISE EXCEPTION 'La base de datos ya contiene datos. Este script es solo para bases limpias.';
    END IF;
END $$;

-- Crear funciones auxiliares para carga masiva
CREATE OR REPLACE FUNCTION load_chart_of_accounts(
    p_org_id UUID,
    p_country_code CHAR(2),
    p_chart_data JSONB
) RETURNS INTEGER AS $$
DECLARE
    v_account JSONB;
    v_count INTEGER := 0;
    v_parent_id UUID;
BEGIN
    -- Procesar cada cuenta del JSON
    FOR v_account IN SELECT * FROM jsonb_array_elements(p_chart_data)
    LOOP
        -- Buscar cuenta padre si existe
        IF v_account->>'parent_code' IS NOT NULL THEN
            SELECT id INTO v_parent_id 
            FROM chart_of_accounts 
            WHERE organization_id = p_org_id 
            AND account_code = v_account->>'parent_code';
        ELSE
            v_parent_id := NULL;
        END IF;
        
        -- Insertar cuenta
        INSERT INTO chart_of_accounts (
            organization_id,
            account_code,
            name,
            type,
            nature,
            level,
            parent_id,
            is_detail,
            metadata
        ) VALUES (
            p_org_id,
            v_account->>'code',
            v_account->>'name',
            v_account->>'type',
            (v_account->>'nature')::CHAR(1),
            (v_account->>'level')::INTEGER,
            v_parent_id,
            COALESCE((v_account->>'is_detail')::BOOLEAN, FALSE),
            COALESCE(v_account->'metadata', '{}')
        );
        
        v_count := v_count + 1;
    END LOOP;
    
    RETURN v_count;
END;
$$ LANGUAGE plpgsql;

-- Crear plantillas base por país
CREATE OR REPLACE FUNCTION create_base_templates(
    p_org_id UUID,
    p_country_code CHAR(2)
) RETURNS INTEGER AS $$
DECLARE
    v_count INTEGER := 0;
BEGIN
    -- Plantilla base para facturas de venta
    INSERT INTO accounting_templates (
        organization_id, template_code, voucher_type, 
        country_code, dsl_content, version
    ) VALUES (
        p_org_id, 
        'INVOICE_SALE_' || p_country_code, 
        'INVOICE', 
        p_country_code,
        'template invoice_sale {
  // Template básico - personalizar según normativa local
  let total = voucher.total
  
  entry {
    debit account("DEUDORES") amount(total) {
      description = "Factura " + voucher.number
    }
    credit account("INGRESOS") amount(total) {
      description = "Venta"
    }
  }
}',
        1
    );
    v_count := v_count + 1;
    
    -- Plantilla base para recibos de caja
    INSERT INTO accounting_templates (
        organization_id, template_code, voucher_type, 
        country_code, dsl_content, version
    ) VALUES (
        p_org_id, 
        'RECEIPT_CASH_' || p_country_code, 
        'RECEIPT', 
        p_country_code,
        'template receipt_cash {
  let amount = voucher.total
  
  entry {
    debit account("CAJA") amount(amount) {
      description = "Recibo " + voucher.number
    }
    credit account("DEUDORES") amount(amount) {
      description = "Abono cliente"
    }
  }
}',
        1
    );
    v_count := v_count + 1;
    
    RETURN v_count;
END;
$$ LANGUAGE plpgsql;

-- Crear índices adicionales para performance
CREATE INDEX IF NOT EXISTS idx_accounts_tree 
ON chart_of_accounts(organization_id, parent_id, account_code);

CREATE INDEX IF NOT EXISTS idx_templates_lookup 
ON accounting_templates(organization_id, voucher_type, country_code, is_active);

-- Crear vista para árbol de cuentas
CREATE OR REPLACE VIEW v_account_tree AS
WITH RECURSIVE account_hierarchy AS (
    -- Cuentas raíz
    SELECT 
        id,
        organization_id,
        account_code,
        name,
        type,
        nature,
        level,
        parent_id,
        is_detail,
        account_code AS path,
        name AS full_name
    FROM chart_of_accounts
    WHERE parent_id IS NULL
    
    UNION ALL
    
    -- Cuentas hijas
    SELECT 
        c.id,
        c.organization_id,
        c.account_code,
        c.name,
        c.type,
        c.nature,
        c.level,
        c.parent_id,
        c.is_detail,
        h.path || '.' || c.account_code AS path,
        h.full_name || ' > ' || c.name AS full_name
    FROM chart_of_accounts c
    INNER JOIN account_hierarchy h ON c.parent_id = h.id
)
SELECT * FROM account_hierarchy
ORDER BY path;

-- Procedimiento para inicializar una nueva organización
CREATE OR REPLACE PROCEDURE initialize_organization(
    p_code VARCHAR,
    p_name VARCHAR,
    p_country_code CHAR(2),
    p_tax_id VARCHAR,
    p_settings JSONB DEFAULT '{}'
) AS $$
DECLARE
    v_org_id UUID;
BEGIN
    -- Crear organización
    INSERT INTO organizations (code, name, country_code, tax_id, settings)
    VALUES (p_code, p_name, p_country_code, p_tax_id, p_settings)
    RETURNING id INTO v_org_id;
    
    -- Crear plantillas base
    PERFORM create_base_templates(v_org_id, p_country_code);
    
    -- Log
    RAISE NOTICE 'Organización % creada con ID: %', p_code, v_org_id;
    RAISE NOTICE 'Use load_chart_of_accounts() para cargar el catálogo de cuentas';
END;
$$ LANGUAGE plpgsql;

-- Configuración de parámetros recomendados
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET effective_cache_size = '1GB';
ALTER SYSTEM SET maintenance_work_mem = '128MB';
ALTER SYSTEM SET checkpoint_completion_target = '0.9';
ALTER SYSTEM SET wal_buffers = '16MB';
ALTER SYSTEM SET default_statistics_target = '100';
ALTER SYSTEM SET random_page_cost = '1.1';

-- Mensaje final
DO $$
BEGIN
    RAISE NOTICE '==============================================';
    RAISE NOTICE 'Base de datos inicializada correctamente';
    RAISE NOTICE '';
    RAISE NOTICE 'Próximos pasos:';
    RAISE NOTICE '1. Ejecutar: CALL initialize_organization(...)';
    RAISE NOTICE '2. Cargar catálogo con: SELECT load_chart_of_accounts(...)';
    RAISE NOTICE '3. Verificar con: SELECT * FROM v_account_tree';
    RAISE NOTICE '==============================================';
END $$;

-- Script de ejemplo para cargar un catálogo mínimo
/*
-- Ejemplo de uso:
CALL initialize_organization(
    'EMPRESA01',
    'Mi Empresa S.A.S.',
    'CO',
    '900123456-7',
    '{"currency": "COP", "fiscal_year": "calendar"}'
);

-- Cargar catálogo desde JSON (ejemplo mínimo)
SELECT load_chart_of_accounts(
    (SELECT id FROM organizations WHERE code = 'EMPRESA01'),
    'CO',
    '[
        {"code": "1", "name": "ACTIVOS", "type": "ASSET", "nature": "D", "level": 1},
        {"code": "11", "name": "DISPONIBLE", "type": "ASSET", "nature": "D", "level": 2, "parent_code": "1"},
        {"code": "1105", "name": "CAJA", "type": "ASSET", "nature": "D", "level": 3, "parent_code": "11", "is_detail": true},
        {"code": "2", "name": "PASIVOS", "type": "LIABILITY", "nature": "C", "level": 1},
        {"code": "3", "name": "PATRIMONIO", "type": "EQUITY", "nature": "C", "level": 1},
        {"code": "4", "name": "INGRESOS", "type": "INCOME", "nature": "C", "level": 1},
        {"code": "5", "name": "GASTOS", "type": "EXPENSE", "nature": "D", "level": 1}
    ]'::jsonb
);
*/
```

#### Documentación de Carga Inicial

##### Para Desarrollo:
1. Ejecutar el script completo de desarrollo
2. Se crean automáticamente 3 organizaciones demo
3. Catálogo completo NIIF para Colombia
4. Datos de prueba incluidos

##### Para Producción:
1. Ejecutar script de producción en base limpia
2. Usar `initialize_organization()` para cada empresa
3. Cargar catálogo específico con `load_chart_of_accounts()`
4. Personalizar plantillas DSL según normativa local

##### Consideraciones:
- Los scripts incluyen validaciones de seguridad
- Índices optimizados para consultas frecuentes
- Funciones auxiliares reutilizables
- Vista jerárquica para navegación del catálogo
- Configuración de PostgreSQL optimizada

---

## ⚙️ 4. Arquitectura de APIs y Servicios

### 4.1 Arquitectura de Microservicios

```
┌─────────────────────────────────────────────────────────────┐
│                      API Gateway (Fiber)                     │
│                    Rate Limiting | Auth | CORS               │
└─────────────┬──────────────────┬──────────────────┬─────────┘
              │                  │                  │
      ┌───────▼────────┐ ┌──────▼───────┐ ┌───────▼────────┐
      │ Voucher Service │ │ Accounting   │ │ Reporting      │
      │                │ │   Service    │ │   Service      │
      └───────┬────────┘ └──────┬───────┘ └───────┬────────┘
              │                  │                  │
      ┌───────▼──────────────────▼──────────────────▼────────┐
      │              PostgreSQL Cluster (Primary + Replicas)  │
      └───────────────────────────────────────────────────────┘
              │                                      │
      ┌───────▼────────┐                    ┌───────▼────────┐
      │  Redis Cache   │                    │ NATS JetStream │
      └────────────────┘                    └────────────────┘
```

### 4.2 API REST Endpoints

#### Voucher Service
```go
// Endpoints principales
POST   /api/v1/vouchers                    // Crear comprobante
GET    /api/v1/vouchers/:id               // Obtener comprobante
GET    /api/v1/vouchers                   // Listar con paginación
PUT    /api/v1/vouchers/:id              // Actualizar comprobante
DELETE /api/v1/vouchers/:id              // Anular comprobante
POST   /api/v1/vouchers/bulk             // Carga masiva
POST   /api/v1/vouchers/:id/process      // Procesar y contabilizar
```

#### Accounting Service
```go
// Gestión de cuentas
GET    /api/v1/accounts                   // Catálogo de cuentas
POST   /api/v1/accounts                   // Crear cuenta
PUT    /api/v1/accounts/:id              // Actualizar cuenta

// Asientos contables
GET    /api/v1/journal-entries           // Listar asientos
POST   /api/v1/journal-entries           // Crear asiento manual
GET    /api/v1/journal-entries/:id       // Detalle de asiento
POST   /api/v1/journal-entries/:id/post  // Mayorizar asiento
POST   /api/v1/journal-entries/:id/reverse // Reversar asiento

// Plantillas DSL
GET    /api/v1/templates                  // Listar plantillas
POST   /api/v1/templates                  // Crear plantilla
PUT    /api/v1/templates/:id             // Actualizar plantilla
POST   /api/v1/templates/validate        // Validar DSL
```

#### Reporting Service
```go
// Libros contables
POST   /api/v1/reports/journal           // Generar libro diario
POST   /api/v1/reports/ledger            // Generar libro mayor
POST   /api/v1/reports/trial-balance     // Balance de comprobación
POST   /api/v1/reports/balance-sheet     // Balance general
POST   /api/v1/reports/income-statement  // Estado de resultados

// Reportes auxiliares
GET    /api/v1/reports/account-balance/:accountId
GET    /api/v1/reports/voucher-summary
POST   /api/v1/reports/custom            // Reportes via DSL
```

### 4.3 GraphQL Schema

```graphql
type Query {
  # Organizaciones
  organization(id: ID!): Organization
  organizations(filter: OrganizationFilter, page: PageInput): OrganizationConnection
  
  # Comprobantes
  voucher(id: ID!): Voucher
  vouchers(filter: VoucherFilter, page: PageInput): VoucherConnection
  
  # Cuentas contables
  account(id: ID!): Account
  accounts(organizationId: ID!, filter: AccountFilter): [Account]
  
  # Asientos
  journalEntry(id: ID!): JournalEntry
  journalEntries(filter: JournalEntryFilter, page: PageInput): JournalEntryConnection
  
  # Reportes
  accountBalance(accountId: ID!, period: String!): AccountBalance
  trialBalance(organizationId: ID!, date: Date!): TrialBalance
}

type Mutation {
  # Comprobantes
  createVoucher(input: VoucherInput!): Voucher
  processVoucher(id: ID!): ProcessResult
  
  # Asientos
  createJournalEntry(input: JournalEntryInput!): JournalEntry
  postJournalEntry(id: ID!): JournalEntry
  
  # Plantillas
  createTemplate(input: TemplateInput!): Template
  validateTemplate(dsl: String!): ValidationResult
}

type Subscription {
  voucherProcessed(organizationId: ID!): VoucherProcessedEvent
  reportGenerated(organizationId: ID!): ReportGeneratedEvent
}
```

### 4.4 Servicios Internos

#### Message Queue (NATS JetStream)
```go
// Streams principales
accounting.voucher.created      // Nuevo comprobante
accounting.voucher.process      // Solicitud de procesamiento
accounting.entry.created        // Asiento generado
accounting.report.generate      // Solicitud de reporte
accounting.audit.log           // Eventos de auditoría

// Consumer Groups
voucher-processor              // Procesa comprobantes
report-generator              // Genera reportes
audit-logger                  // Registra auditoría
notification-sender           // Envía notificaciones
```

#### Cache Strategy (Redis)
```go
// Patrones de cache
organizations:{id}            // TTL: 1 hora
accounts:{org_id}            // TTL: 30 minutos
templates:{org_id}:{type}    // TTL: 1 hora
user_permissions:{user_id}   // TTL: 15 minutos
report:{id}                  // TTL: 24 horas
```

### 4.5 Autenticación y Autorización

#### JWT + RBAC
```go
// Roles principales
SUPER_ADMIN       // Acceso total sistema
ORG_ADMIN        // Admin de organización
ACCOUNTANT       // Contador (CRUD completo)
AUDITOR          // Solo lectura + reportes
CLERK            // Crear comprobantes
API_CLIENT       // Acceso programático

// Middleware de autorización
func RequirePermission(permission string) fiber.Handler
func RequireRole(roles ...string) fiber.Handler
func RequireOrgAccess() fiber.Handler
```

---

## 🖥️ 5. Diseño de Interfaces de Usuario

### 5.1 Arquitectura Frontend

- **Framework**: React 18 + TypeScript
- **UI Library**: Ant Design Pro
- **State Management**: Zustand
- **Data Fetching**: TanStack Query
- **Routing**: React Router v6
- **Build Tool**: Vite
- **CSS**: Tailwind CSS + CSS Modules

### 5.2 Módulos de UI

#### Dashboard Principal
```
┌─────────────────────────────────────────────────────────┐
│ 🏢 Motor Contable            👤 Usuario | 🏢 Empresa   │
├─────────────────────────────────────────────────────────┤
│ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐       │
│ │ Comprobantes│ │  Asientos   │ │  Pendientes │       │
│ │   150,234   │ │   300,468   │ │    1,234    │       │
│ │   📈 +12%   │ │   📈 +15%   │ │   📉 -5%    │       │
│ └─────────────┘ └─────────────┘ └─────────────┘       │
│                                                         │
│ 📊 Procesamiento Diario          ⚡ Rendimiento        │
│ [████████████████░░] 85%        58 comp/seg           │
│                                                         │
│ 📋 Actividad Reciente                                  │
│ ┌─────────────────────────────────────────────────┐   │
│ │ • Factura #12345 procesada - hace 2 min        │   │
│ │ • Libro Diario generado - hace 15 min          │   │
│ │ • 500 comprobantes importados - hace 1 hora    │   │
│ └─────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
```

#### Gestión de Comprobantes
- **Lista de comprobantes**: DataGrid con filtros avanzados, paginación server-side
- **Detalle de comprobante**: Vista/edición con validaciones en tiempo real
- **Carga masiva**: Drag & drop para archivos CSV/XML con preview
- **Monitor de procesamiento**: Estado en tiempo real con WebSockets

#### Configuración de Cuentas
- **Plan de cuentas**: TreeView jerárquico con drag & drop
- **Importación**: Asistente para importar desde Excel
- **Mapeo**: Interface visual para mapear cuentas entre sistemas
- **Validación**: Verificación de integridad del catálogo

#### Plantillas DSL
- **Editor de plantillas**: Monaco Editor con syntax highlighting para DSL
- **Testing sandbox**: Área para probar plantillas con datos de ejemplo
- **Versionado**: Control de versiones con diff viewer
- **Documentación inline**: Tooltips y autocompletado

#### Generación de Reportes
- **Selector de reportes**: Cards con descripción y requisitos
- **Parámetros**: Formulario dinámico según tipo de reporte
- **Vista previa**: Renderizado HTML antes de exportar
- **Exportación**: PDF, Excel, CSV con formatos personalizables
- **Programación**: Scheduler para reportes automáticos

#### Panel de Auditoría
- **Log de eventos**: Timeline con filtros por usuario/acción
- **Trazabilidad**: Seguimiento completo de cambios
- **Alertas**: Notificaciones de actividades sospechosas
- **Exportación**: Logs en formato estándar para auditoría externa

#### Administración Multi-tenant
- **Gestión de organizaciones**: CRUD con configuración por país
- **Usuarios y permisos**: Asignación granular de roles
- **Configuración fiscal**: Parámetros por país/empresa
- **Monitoreo de uso**: Métricas por tenant

### 5.3 Características de UX

#### Responsive Design
- Desktop: Layout completo con sidebars
- Tablet: Menú colapsable, grids adaptables
- Mobile: Navegación bottom tab, vistas simplificadas

#### Performance
- Code splitting por ruta
- Lazy loading de componentes pesados
- Virtual scrolling para listas grandes
- Optimistic updates para mejor percepción

#### Accesibilidad
- WCAG 2.1 Level AA compliance
- Navegación por teclado completa
- Screen reader friendly
- Alto contraste disponible

---

## 🧑‍💻 6. Domain Specific Languages (DSLs)

### 6.1 DSL para Asientos Contables

#### Justificación
La contabilización de comprobantes varía significativamente entre países, tipos de documento y políticas empresariales. Un DSL permite:

1. **Flexibilidad**: Cambiar reglas sin modificar código
2. **Claridad**: Reglas legibles por contadores
3. **Versionado**: Control de cambios en normativas
4. **Performance**: Compilación a AST para ejecución rápida

#### Sintaxis del DSL
```dsl
// Plantilla para Factura de Venta - Colombia
template invoice_sale_co {
  // Definir variables desde el comprobante
  let subtotal = voucher.subtotal
  let tax_19 = voucher.taxes["iva_19"]
  let tax_5 = voucher.taxes["iva_5"]
  let retention = voucher.retentions["rte_fuente"]
  let total = voucher.total
  
  // Reglas de validación
  require subtotal > 0 : "El subtotal debe ser positivo"
  require total == subtotal + tax_19 + tax_5 - retention : "Total no cuadra"
  
  // Generar asientos
  entry {
    // Cargo a cuentas por cobrar
    debit account("1305.01") amount(total) {
      description = "Venta según factura " + voucher.number
      cost_center = voucher.cost_center
    }
    
    // Abono a ingresos
    credit account("4135.01") amount(subtotal) {
      description = "Ingresos por ventas"
      split_by = voucher.line_items using item.account
    }
    
    // IVA 19%
    if tax_19 > 0 {
      credit account("2408.01") amount(tax_19) {
        description = "IVA 19% por pagar"
      }
    }
    
    // IVA 5%
    if tax_5 > 0 {
      credit account("2408.02") amount(tax_5) {
        description = "IVA 5% por pagar"
      }
    }
    
    // Retención en la fuente
    if retention > 0 {
      debit account("1355.25") amount(retention) {
        description = "Retención en la fuente practicada"
      }
    }
  }
  
  // Post-procesamiento
  after {
    notify_if total > 10000000 : "Factura de alto valor"
    update_customer_balance(voucher.customer_id, total)
  }
}
```

#### Parser y Ejecución
```go
// Estructura AST generada
type Template struct {
    Name        string
    Variables   []Variable
    Validations []Validation
    Entry       EntryBlock
    AfterBlock  []Action
}

// Ejecución del template
func (t *Template) Execute(voucher Voucher, ctx Context) (*JournalEntry, error) {
    // 1. Evaluar variables
    vars := t.evaluateVariables(voucher)
    
    // 2. Ejecutar validaciones
    if err := t.runValidations(vars); err != nil {
        return nil, err
    }
    
    // 3. Generar asiento
    entry := t.generateEntry(vars, ctx)
    
    // 4. Ejecutar post-procesamiento
    t.runAfterActions(entry, ctx)
    
    return entry, nil
}
```

### 6.2 DSL para Reportes Contables

#### Justificación
Los reportes contables requieren:
1. **Personalización**: Cada empresa tiene formatos específicos
2. **Cálculos complejos**: Fórmulas entre cuentas y períodos
3. **Agrupaciones dinámicas**: Por cuenta, centro de costo, proyecto
4. **Formatos múltiples**: PDF, Excel, XML para entidades regulatorias

#### Sintaxis del DSL de Reportes
```dsl
report balance_sheet {
  // Metadatos
  title = "Balance General"
  period = params.start_date to params.end_date
  currency = params.currency default "COP"
  
  // Definir secciones
  section assets {
    title = "ACTIVOS"
    
    group current_assets {
      title = "Activos Corrientes"
      
      line cash {
        label = "Efectivo y equivalentes"
        accounts = ["1105.*", "1110.*"]
        formula = sum(balance)
      }
      
      line accounts_receivable {
        label = "Cuentas por cobrar"
        accounts = ["1305.*", "1330.*", "1355.*"]
        formula = sum(debit) - sum(credit)
      }
      
      line inventory {
        label = "Inventarios"
        accounts = ["14*"]
        formula = sum(balance)
      }
      
      subtotal {
        label = "Total Activos Corrientes"
        bold = true
      }
    }
    
    group fixed_assets {
      title = "Activos Fijos"
      
      line property {
        label = "Propiedad, planta y equipo"
        accounts = ["15*"]
        formula = sum(balance)
      }
      
      line depreciation {
        label = "Depreciación acumulada"
        accounts = ["1592*", "1598*"]
        formula = -sum(balance)
        format = "(#,##0.00)"
      }
      
      subtotal {
        label = "Total Activos Fijos Netos"
        bold = true
      }
    }
    
    total {
      label = "TOTAL ACTIVOS"
      bold = true
      style = "double-underline"
    }
  }
  
  // Validaciones del reporte
  validate {
    assets.total == liabilities.total + equity.total : "Balance no cuadra"
  }
  
  // Formato de salida
  output {
    format = params.format in ["pdf", "excel", "xml"]
    template = "templates/balance_sheet.tmpl"
    
    // Configuración específica por formato
    if format == "pdf" {
      orientation = "portrait"
      margins = [2cm, 2cm, 2cm, 2cm]
      font = "Arial"
      font_size = 10
    }
    
    if format == "excel" {
      freeze_panes = "B4"
      auto_filter = true
      column_widths = [20, 40, 15, 15]
    }
  }
}
```

#### Motor de Ejecución de Reportes
```go
// Pipeline de generación
type ReportEngine struct {
    parser   *dsl.Parser
    data     DataProvider
    renderer Renderer
}

func (e *ReportEngine) Generate(dslContent string, params map[string]any) (Report, error) {
    // 1. Parsear DSL
    ast, err := e.parser.Parse(dslContent)
    if err != nil {
        return nil, fmt.Errorf("parse error: %w", err)
    }
    
    // 2. Obtener datos
    data, err := e.data.Fetch(ast, params)
    if err != nil {
        return nil, fmt.Errorf("data fetch error: %w", err)
    }
    
    // 3. Ejecutar cálculos
    results := e.executeCalculations(ast, data)
    
    // 4. Validar
    if err := e.validate(ast, results); err != nil {
        return nil, fmt.Errorf("validation error: %w", err)
    }
    
    // 5. Renderizar
    report, err := e.renderer.Render(ast, results, params)
    if err != nil {
        return nil, fmt.Errorf("render error: %w", err)
    }
    
    return report, nil
}
```

### 6.3 DSL para Reglas Fiscales (Evaluación)

#### Consideraciones
Un tercer DSL para reglas fiscales podría ser beneficioso pero añade complejidad. 

**Ventajas**:
- Actualización rápida de cambios regulatorios
- Reglas específicas por país sin tocar código
- Auditoría clara de qué reglas se aplicaron

**Desventajas**:
- Complejidad adicional en mantenimiento
- Curva de aprendizaje para tres DSLs
- Posible duplicación con DSL de asientos

**Recomendación**: Iniciar con reglas fiscales como parte del DSL de asientos contables. Si la complejidad crece, extraer a DSL separado en fase posterior.

---

## 🪜 7. Plan de Implementación por Fases

### 7.1 Fase 1: MVP Core (3 meses)
**Impacto**: Alto | **Complejidad**: Media

#### Entregables:
1. **Modelo de datos base**
   - Esquema PostgreSQL core
   - Migración y seeders
   - Índices básicos

2. **API REST básica**
   - CRUD de comprobantes
   - CRUD de cuentas contables
   - Autenticación JWT

3. **Motor de contabilización simple**
   - Procesamiento síncrono
   - Plantillas hardcodeadas para Colombia
   - Generación de asientos básicos

4. **UI mínima**
   - Login y navegación
   - Lista y creación de comprobantes
   - Vista de asientos generados

#### Métricas de éxito:
- Procesar 1000 comprobantes/día
- Generar asientos correctos para 3 tipos de documento
- Tiempo de respuesta < 500ms

### 7.2 Fase 2: DSL de Plantillas (2 meses)
**Impacto**: Alto | **Complejidad**: Alta

#### Entregables:
1. **Parser DSL con go-dsl**
   - Gramática completa
   - Validador sintáctico
   - Compilador a AST

2. **Motor de ejecución**
   - Runtime para plantillas
   - Cache de templates compilados
   - Hot reload de cambios

3. **Editor de plantillas**
   - Syntax highlighting
   - Validación en tiempo real
   - Pruebas con datos mock

4. **Migración de reglas**
   - Convertir reglas hardcodeadas a DSL
   - Templates para 5 tipos de documento
   - Documentación de sintaxis

#### Métricas de éxito:
- 90% de asientos generados vía DSL
- Tiempo de compilación < 100ms
- 0 errores en producción por DSL

### 7.3 Fase 3: Multi-tenant y Seguridad (2 meses)
**Impacto**: Alto | **Complejidad**: Media

#### Entregables:
1. **Aislamiento de datos**
   - Row Level Security en PostgreSQL
   - Middleware de tenant detection
   - Validación en todas las queries

2. **Gestión de organizaciones**
   - CRUD de empresas
   - Configuración por tenant
   - Límites y quotas

3. **RBAC completo**
   - Roles y permisos granulares
   - Auditoría de accesos
   - API keys para integraciones

4. **Seguridad reforzada**
   - Encriptación de datos sensibles
   - WAF rules
   - Rate limiting por tenant

#### Métricas de éxito:
- 100% de queries con tenant isolation
- 0 fugas de datos entre tenants
- Cumplimiento SOC 2 Type I

### 7.4 Fase 4: Libros Contables y Reportes (2 meses)
**Impacto**: Alto | **Complejidad**: Media

#### Entregables:
1. **Generador de libros**
   - Libro Diario
   - Libro Mayor
   - Balance de Comprobación

2. **DSL de reportes**
   - Parser y runtime
   - Templates base
   - Exportación PDF/Excel

3. **UI de reportes**
   - Catálogo de reportes
   - Parametrización
   - Programación automática

4. **Optimizaciones**
   - Vistas materializadas
   - Pre-cálculo nocturno
   - Cache de reportes

#### Métricas de éxito:
- Generación de reportes < 10 segundos
- 100% precisión en cálculos
- Soporte 10 formatos de reporte

### 7.5 Fase 5: Escalabilidad Masiva (3 meses)
**Impacto**: Alto | **Complejidad**: Alta

#### Entregables:
1. **Procesamiento asíncrono**
   - Workers con NATS JetStream
   - Bulk processing
   - Retry mechanism

2. **Optimización de base de datos**
   - Particionamiento automático
   - Read replicas
   - Connection pooling

3. **Infraestructura elástica**
   - Kubernetes en DigitalOcean
   - Auto-scaling
   - Monitoring completo

4. **Performance tuning**
   - Profiling y optimización
   - Índices especializados
   - Query optimization

#### Métricas de éxito:
- Procesar 5M comprobantes/día
- Latencia p99 < 200ms
- Disponibilidad 99.9%

### 7.6 Fase 6: Multi-país Completo (3 meses)
**Impacto**: Medio | **Complejidad**: Alta

#### Entregables:
1. **Adaptación por país**
   - Templates para MX, CL, EC, UY, PE
   - Catálogos de cuentas locales
   - Validaciones fiscales

2. **Reportes regulatorios**
   - Formatos XML/JSON oficiales
   - Validaciones DIAN, SAT, SII
   - Firma electrónica

3. **Localización UI**
   - Traducción completa
   - Formatos de fecha/moneda
   - Ayuda contextual

4. **Cumplimiento normativo**
   - Certificaciones locales
   - Auditoría de cumplimiento
   - Documentación legal

#### Métricas de éxito:
- 100% cumplimiento normativo
- Templates para 20 tipos documento/país
- Certificación en 3 países

### 7.7 Fase 7: Inteligencia y Automatización (2 meses)
**Impacto**: Medio | **Complejidad**: Media

#### Entregables:
1. **Dashboard analítico**
   - KPIs en tiempo real
   - Tendencias y proyecciones
   - Alertas inteligentes

2. **Automatización avanzada**
   - Detección de anomalías
   - Sugerencias de mejora
   - Auto-categorización

3. **Integraciones**
   - APIs de bancos
   - ERPs principales
   - Plataformas de facturación

4. **Machine Learning** (opcional)
   - Clasificación automática
   - Detección de fraude
   - Predicción de flujo de caja

#### Métricas de éxito:
- 80% de comprobantes auto-clasificados
- Reducción 50% en tiempo de cierre
- ROI demostrable para clientes

---

## 🔁 8. Estrategias Técnicas Adicionales

### 8.1 Estrategia de Cache

#### Redis Cluster Architecture
```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  Redis Master 1 │────▶│  Redis Master 2 │────▶│  Redis Master 3 │
│   Slots 0-5460  │     │ Slots 5461-10922│     │Slots 10923-16383│
└────────┬────────┘     └────────┬────────┘     └────────┬────────┘
         │                       │                        │
    ┌────▼────┐            ┌────▼────┐             ┌────▼────┐
    │Replica 1│            │Replica 2│             │Replica 3│
    └─────────┘            └─────────┘             └─────────┘
```

#### Políticas de Cache
```go
// Cache layers
type CacheStrategy struct {
    L1 *MemoryCache  // In-process cache (10MB)
    L2 *RedisCache   // Distributed cache
}

// Patrones de invalidación
func (c *CacheStrategy) InvalidatePattern(pattern string) {
    // Invalida L1
    c.L1.InvalidatePattern(pattern)
    
    // Invalida L2 con Lua script
    script := `
        local keys = redis.call('keys', ARGV[1])
        for i=1,#keys,5000 do
            redis.call('del', unpack(keys, i, math.min(i+4999, #keys)))
        end
        return #keys
    `
    c.L2.Eval(script, pattern)
}
```

### 8.2 Database Sharding Strategy

#### Sharding por Organization ID
```sql
-- Función de sharding
CREATE OR REPLACE FUNCTION get_shard_id(org_id UUID) 
RETURNS INTEGER AS $$
BEGIN
    RETURN abs(hashtext(org_id::text)) % 4;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Tablas distribuidas
CREATE TABLE vouchers_shard_0 () INHERITS (vouchers);
CREATE TABLE vouchers_shard_1 () INHERITS (vouchers);
CREATE TABLE vouchers_shard_2 () INHERITS (vouchers);
CREATE TABLE vouchers_shard_3 () INHERITS (vouchers);

-- Trigger de routing
CREATE OR REPLACE FUNCTION voucher_insert_trigger()
RETURNS TRIGGER AS $$
BEGIN
    EXECUTE format('INSERT INTO vouchers_shard_%s VALUES ($1.*)', 
                   get_shard_id(NEW.organization_id))
    USING NEW;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
```

### 8.3 Async Workers con NATS

#### Worker Pool Architecture
```go
type WorkerPool struct {
    workers    int
    queue      chan Job
    wg         sync.WaitGroup
    metrics    *prometheus.Registry
}

func (p *WorkerPool) Start(ctx context.Context) {
    for i := 0; i < p.workers; i++ {
        p.wg.Add(1)
        go p.worker(ctx, i)
    }
}

func (p *WorkerPool) worker(ctx context.Context, id int) {
    defer p.wg.Done()
    
    for {
        select {
        case <-ctx.Done():
            return
        case job := <-p.queue:
            p.processJob(job, id)
        }
    }
}

// Job processing con circuit breaker
func (p *WorkerPool) processJob(job Job, workerID int) {
    start := time.Now()
    
    err := hystrix.Do(job.Type, func() error {
        return job.Execute()
    }, func(err error) error {
        // Fallback logic
        return p.handleFailure(job, err)
    })
    
    p.recordMetrics(job.Type, time.Since(start), err)
}
```

### 8.4 Indexing Strategy

#### Índices Especializados
```sql
-- Índice parcial para comprobantes pendientes
CREATE INDEX idx_vouchers_pending 
ON vouchers(organization_id, created_at) 
WHERE status = 'PENDING';

-- Índice GIN para búsquedas en metadata
CREATE INDEX idx_vouchers_metadata 
ON vouchers USING gin(metadata);

-- Índice BRIN para datos temporales
CREATE INDEX idx_journal_entries_date 
ON journal_entries USING brin(entry_date);

-- Índice compuesto para queries frecuentes
CREATE INDEX idx_journal_lines_lookup 
ON journal_lines(journal_entry_id, account_id) 
INCLUDE (debit_amount, credit_amount);

-- Índice para full-text search
CREATE INDEX idx_vouchers_search 
ON vouchers USING gin(
    to_tsvector('spanish', 
        coalesce(description, '') || ' ' || 
        coalesce(voucher_number, '')
    )
);
```

### 8.5 Monitoring y Observability

#### Stack de Monitoreo
```yaml
# Prometheus scrape config
scrape_configs:
  - job_name: 'accounting-engine'
    static_configs:
      - targets: ['api:8080', 'workers:8081']
    metrics_path: '/metrics'
    scrape_interval: 15s

# Grafana dashboards
dashboards:
  - name: "Accounting Engine Overview"
    panels:
      - vouchers_processed_rate
      - journal_entries_created_rate
      - api_latency_percentiles
      - error_rate_by_service
      - database_connections
      - cache_hit_ratio

# Alerting rules
alerts:
  - name: HighErrorRate
    expr: rate(errors_total[5m]) > 0.05
    severity: warning
    
  - name: DatabaseConnectionExhaustion
    expr: pg_connections_active / pg_connections_max > 0.8
    severity: critical
```

#### Distributed Tracing
```go
// OpenTelemetry setup
func InitTracing() {
    exporter, _ := jaeger.New(
        jaeger.WithCollectorEndpoint(
            jaeger.WithEndpoint("http://jaeger:14268/api/traces"),
        ),
    )
    
    tp := tracesdk.NewTracerProvider(
        tracesdk.WithBatcher(exporter),
        tracesdk.WithResource(resource.NewWithAttributes(
            semconv.ServiceNameKey.String("accounting-engine"),
            semconv.ServiceVersionKey.String("1.0.0"),
        )),
    )
    
    otel.SetTracerProvider(tp)
}

// Instrumentación
func (s *VoucherService) ProcessVoucher(ctx context.Context, id string) error {
    ctx, span := otel.Tracer("voucher-service").Start(ctx, "ProcessVoucher")
    defer span.End()
    
    span.SetAttributes(
        attribute.String("voucher.id", id),
        attribute.String("tenant.id", GetTenantID(ctx)),
    )
    
    // Business logic with sub-spans
    if err := s.validate(ctx, id); err != nil {
        span.RecordError(err)
        return err
    }
    
    return s.generateEntries(ctx, id)
}
```

---

## 📋 Conclusiones y Próximos Pasos

### Resumen de la Propuesta

Esta arquitectura proporciona:

1. **Escalabilidad**: Diseñada para manejar 5M+ comprobantes/día
2. **Flexibilidad**: DSLs permiten cambios sin modificar código
3. **Multi-tenancy**: Aislamiento completo entre organizaciones
4. **Multi-país**: Soporte para normativas de 6 países
5. **Observabilidad**: Monitoreo completo de la plataforma

### Inversión Estimada

#### Recursos Humanos (12 meses)
- 1 Tech Lead / Arquitecto
- 3 Backend Engineers (Go)
- 2 Frontend Engineers (React)
- 1 DevOps Engineer
- 1 QA Engineer
- 1 Product Manager

#### Infraestructura (DigitalOcean)
- **Desarrollo**: $500/mes
- **Staging**: $1,000/mes
- **Producción**: $3,000-5,000/mes (según carga)

### Riesgos y Mitigaciones

1. **Complejidad de DSLs**
   - Mitigación: Documentación exhaustiva y tooling
   
2. **Performance con 5M registros/día**
   - Mitigación: Pruebas de carga tempranas y continuas
   
3. **Cambios regulatorios**
   - Mitigación: Arquitectura flexible y DSLs actualizables
   
4. **Adopción por usuarios**
   - Mitigación: UI intuitiva y migración asistida

### Siguientes Pasos Inmediatos

1. **Validación técnica**: POC del DSL engine con casos reales
2. **Validación de negocio**: Feedback de contadores y empresas piloto
3. **Definición de MVP**: Priorizar features para primera versión
4. **Formación del equipo**: Reclutar talento especializado
5. **Setup inicial**: Infraestructura de desarrollo y CI/CD

---

*Documento preparado para el desarrollo del Motor Contable Cloud-Native*
*Versión 1.0 - Diciembre 2024*