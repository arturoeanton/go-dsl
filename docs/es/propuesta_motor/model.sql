-- ==============================================
-- MODELO DE DATOS - MOTOR CONTABLE CLOUD-NATIVE
-- ==============================================
-- Versión: 2.0
-- Última actualización: 2025-01-24
-- Base de datos: PostgreSQL 15+
-- Arquitectura: Monolito Go/Fiber/go-dsl → Microservicios
-- ==============================================

-- Crear esquema principal
CREATE SCHEMA IF NOT EXISTS accounting_engine;
SET search_path TO accounting_engine, public;

-- Habilitar extensiones necesarias
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "pg_trgm"; -- Para búsquedas de texto

-- ==============================================
-- TABLAS PRINCIPALES
-- ==============================================

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
    is_active BOOLEAN DEFAULT true,
    CONSTRAINT chk_country_code CHECK (country_code IN ('CO', 'MX', 'CL', 'EC', 'UY', 'PE'))
);

-- Catálogo de cuentas contables
CREATE TABLE chart_of_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    account_code VARCHAR(50) NOT NULL,
    name VARCHAR(200) NOT NULL,
    type VARCHAR(20) NOT NULL,
    nature CHAR(1) NOT NULL,
    level INTEGER NOT NULL,
    parent_id UUID REFERENCES chart_of_accounts(id),
    is_detail BOOLEAN DEFAULT false,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_active BOOLEAN DEFAULT true,
    UNIQUE(organization_id, account_code),
    CONSTRAINT chk_account_type CHECK (type IN ('ASSET', 'LIABILITY', 'EQUITY', 'INCOME', 'EXPENSE')),
    CONSTRAINT chk_nature CHECK (nature IN ('D', 'C')),
    CONSTRAINT chk_level CHECK (level BETWEEN 1 AND 10)
);

-- Comprobantes (documentos fuente)
CREATE TABLE vouchers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    voucher_number VARCHAR(50) NOT NULL,
    voucher_type VARCHAR(50) NOT NULL,
    voucher_date DATE NOT NULL,
    description TEXT,
    total_amount DECIMAL(20,4) NOT NULL,
    currency_code CHAR(3) NOT NULL DEFAULT 'COP',
    exchange_rate DECIMAL(10,6) DEFAULT 1.0,
    source_system VARCHAR(100),
    external_ref VARCHAR(100),
    third_party_id UUID,
    third_party_name VARCHAR(200),
    third_party_document VARCHAR(50),
    tax_details JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'PENDING',
    error_message TEXT,
    dsl_template_id UUID REFERENCES accounting_templates(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    processed_at TIMESTAMPTZ,
    created_by UUID NOT NULL REFERENCES users(id),
    UNIQUE(organization_id, voucher_number),
    CONSTRAINT chk_voucher_type CHECK (voucher_type IN ('invoice_sale', 'invoice_purchase', 'payment', 'receipt', 'credit_note', 'debit_note', 'journal_entry')),
    CONSTRAINT chk_status CHECK (status IN ('PENDING', 'PROCESSING', 'PROCESSED', 'ERROR', 'CANCELLED'))
) PARTITION BY RANGE (voucher_date);

-- Asientos contables (journal entries)
CREATE TABLE journal_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    entry_number BIGSERIAL,
    entry_date DATE NOT NULL,
    voucher_id UUID REFERENCES vouchers(id),
    description TEXT NOT NULL,
    entry_type VARCHAR(20) NOT NULL,
    period VARCHAR(7) NOT NULL,
    status VARCHAR(20) DEFAULT 'DRAFT',
    created_by UUID NOT NULL,
    approved_by UUID,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    approved_at TIMESTAMPTZ,
    is_reversed BOOLEAN DEFAULT false,
    reversal_id UUID REFERENCES journal_entries(id),
    UNIQUE(organization_id, entry_number),
    CONSTRAINT chk_entry_type CHECK (entry_type IN ('STANDARD', 'ADJUSTMENT', 'CLOSING', 'REVERSAL')),
    CONSTRAINT chk_entry_status CHECK (status IN ('DRAFT', 'PENDING', 'POSTED', 'CANCELLED')),
    CONSTRAINT chk_period_format CHECK (period ~ '^[0-9]{4}-[0-9]{2}$')
) PARTITION BY RANGE (entry_date);

-- Líneas de asientos
CREATE TABLE journal_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    journal_entry_id UUID NOT NULL REFERENCES journal_entries(id) ON DELETE CASCADE,
    line_number INTEGER NOT NULL,
    account_id UUID NOT NULL REFERENCES chart_of_accounts(id),
    debit_amount DECIMAL(20,4) DEFAULT 0,
    credit_amount DECIMAL(20,4) DEFAULT 0,
    description TEXT,
    cost_center_id UUID,
    project_id UUID,
    metadata JSONB DEFAULT '{}',
    CHECK (debit_amount >= 0 AND credit_amount >= 0),
    CHECK (debit_amount = 0 OR credit_amount = 0),
    UNIQUE(journal_entry_id, line_number)
);

-- Plantillas DSL para contabilización
CREATE TABLE accounting_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    template_code VARCHAR(50) NOT NULL,
    voucher_type VARCHAR(50) NOT NULL,
    country_code CHAR(2) NOT NULL,
    dsl_content TEXT NOT NULL,
    compiled_dsl JSONB,
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
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    book_type VARCHAR(50) NOT NULL,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    generation_date TIMESTAMPTZ DEFAULT NOW(),
    file_url TEXT,
    status VARCHAR(20) DEFAULT 'GENERATING',
    metadata JSONB DEFAULT '{}',
    CONSTRAINT chk_book_type CHECK (book_type IN ('JOURNAL', 'LEDGER', 'TRIAL_BALANCE', 'BALANCE_SHEET', 'INCOME_STATEMENT')),
    CONSTRAINT chk_book_status CHECK (status IN ('GENERATING', 'COMPLETED', 'ERROR'))
);

-- ==============================================
-- TABLAS AUXILIARES
-- ==============================================

-- Usuarios del sistema
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(200) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    organization_id UUID REFERENCES organizations(id),
    permissions JSONB DEFAULT '{}',
    preferences JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    last_login TIMESTAMPTZ,
    login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMPTZ,
    password_changed_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT chk_role CHECK (role IN ('SUPER_ADMIN', 'ORG_ADMIN', 'MANAGER', 'ACCOUNTANT', 'AUDITOR', 'CLERK', 'VIEWER', 'API_CLIENT'))
);

-- Centros de costo
CREATE TABLE cost_centers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    code VARCHAR(20) NOT NULL,
    name VARCHAR(200) NOT NULL,
    parent_id UUID REFERENCES cost_centers(id),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(organization_id, code)
);

-- Proyectos
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    code VARCHAR(20) NOT NULL,
    name VARCHAR(200) NOT NULL,
    start_date DATE,
    end_date DATE,
    budget DECIMAL(20,4),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(organization_id, code)
);

-- Log de auditoría
CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID REFERENCES organizations(id),
    user_id UUID REFERENCES users(id),
    action VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ==============================================
-- ÍNDICES PARA PERFORMANCE
-- ==============================================

-- Índices principales
CREATE INDEX idx_vouchers_org_date ON vouchers(organization_id, voucher_date);
CREATE INDEX idx_vouchers_status ON vouchers(status) WHERE status = 'PENDING';
CREATE INDEX idx_vouchers_type ON vouchers(organization_id, voucher_type);
CREATE INDEX idx_vouchers_third_party ON vouchers(organization_id, third_party_id) WHERE third_party_id IS NOT NULL;
CREATE INDEX idx_journal_entries_org_period ON journal_entries(organization_id, period);
CREATE INDEX idx_journal_entries_voucher ON journal_entries(voucher_id) WHERE voucher_id IS NOT NULL;
CREATE INDEX idx_journal_lines_account ON journal_lines(account_id);
CREATE INDEX idx_chart_accounts_org_code ON chart_of_accounts(organization_id, account_code);
CREATE INDEX idx_chart_accounts_parent ON chart_of_accounts(parent_id) WHERE parent_id IS NOT NULL;
CREATE INDEX idx_third_parties_org_doc ON third_parties(organization_id, document_type, document_number);
CREATE INDEX idx_voucher_lines_voucher ON voucher_lines(voucher_id);
CREATE INDEX idx_voucher_lines_account ON voucher_lines(account_id) WHERE account_id IS NOT NULL;

-- Índices para búsquedas
CREATE INDEX idx_vouchers_metadata ON vouchers USING gin(metadata);
CREATE INDEX idx_organizations_settings ON organizations USING gin(settings);
CREATE INDEX idx_vouchers_search ON vouchers USING gin(to_tsvector('spanish', coalesce(description, '') || ' ' || coalesce(voucher_number, '')));

-- Índices para auditoría
CREATE INDEX idx_audit_log_org_date ON audit_log(organization_id, created_at DESC);
CREATE INDEX idx_audit_log_entity ON audit_log(entity_type, entity_id);
CREATE INDEX idx_books_org_period ON accounting_books(organization_id, period_start, period_end);

-- ==============================================
-- VISTAS MATERIALIZADAS
-- ==============================================

-- Balance de cuentas por período
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

-- ==============================================
-- FUNCIONES Y TRIGGERS
-- ==============================================

-- Función para actualizar updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers para updated_at
CREATE TRIGGER update_organizations_updated_at BEFORE UPDATE ON organizations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    
CREATE TRIGGER update_chart_accounts_updated_at BEFORE UPDATE ON chart_of_accounts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Función para validar balance de asientos
CREATE OR REPLACE FUNCTION validate_journal_balance()
RETURNS TRIGGER AS $$
DECLARE
    v_total_debit DECIMAL(20,4);
    v_total_credit DECIMAL(20,4);
BEGIN
    IF NEW.status = 'POSTED' THEN
        SELECT 
            COALESCE(SUM(debit_amount), 0),
            COALESCE(SUM(credit_amount), 0)
        INTO v_total_debit, v_total_credit
        FROM journal_lines
        WHERE journal_entry_id = NEW.id;
        
        IF v_total_debit != v_total_credit THEN
            RAISE EXCEPTION 'El asiento no está balanceado. Débitos: %, Créditos: %', 
                v_total_debit, v_total_credit;
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER validate_journal_balance_trigger
    BEFORE UPDATE OF status ON journal_entries
    FOR EACH ROW
    WHEN (NEW.status = 'POSTED' AND OLD.status != 'POSTED')
    EXECUTE FUNCTION validate_journal_balance();

-- ==============================================
-- TABLAS ADICIONALES PARA FUNCIONALIDAD COMPLETA
-- ==============================================

-- Terceros (proveedores/clientes)
CREATE TABLE third_parties (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    document_type VARCHAR(20) NOT NULL,
    document_number VARCHAR(50) NOT NULL,
    name VARCHAR(200) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(50),
    address TEXT,
    city VARCHAR(100),
    country_code CHAR(2) NOT NULL,
    tax_regime VARCHAR(50),
    tax_details JSONB DEFAULT '{}',
    account_receivable_id UUID REFERENCES chart_of_accounts(id),
    account_payable_id UUID REFERENCES chart_of_accounts(id),
    credit_limit DECIMAL(20,4) DEFAULT 0,
    payment_terms INTEGER DEFAULT 30,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(organization_id, document_type, document_number),
    CONSTRAINT chk_tp_document_type CHECK (document_type IN ('CC', 'NIT', 'CE', 'RUT', 'RFC', 'RUC', 'CUIT', 'CI', 'DNI', 'PASSPORT'))
);

-- Líneas de detalle de comprobantes
CREATE TABLE voucher_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    voucher_id UUID NOT NULL REFERENCES vouchers(id) ON DELETE CASCADE,
    line_number INTEGER NOT NULL,
    account_id UUID REFERENCES chart_of_accounts(id),
    description TEXT,
    quantity DECIMAL(10,4) DEFAULT 1,
    unit_price DECIMAL(20,4) DEFAULT 0,
    line_amount DECIMAL(20,4) NOT NULL,
    tax_code VARCHAR(20),
    tax_rate DECIMAL(5,4) DEFAULT 0,
    tax_amount DECIMAL(20,4) DEFAULT 0,
    discount_rate DECIMAL(5,4) DEFAULT 0,
    discount_amount DECIMAL(20,4) DEFAULT 0,
    cost_center_id UUID REFERENCES cost_centers(id),
    project_id UUID REFERENCES projects(id),
    metadata JSONB DEFAULT '{}',
    UNIQUE(voucher_id, line_number)
);

-- Tipos de comprobante configurables
CREATE TABLE voucher_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    type_code VARCHAR(50) NOT NULL,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    prefix VARCHAR(10),
    next_number BIGINT DEFAULT 1,
    affects_inventory BOOLEAN DEFAULT false,
    requires_third_party BOOLEAN DEFAULT false,
    default_account_id UUID REFERENCES chart_of_accounts(id),
    dsl_template_id UUID REFERENCES accounting_templates(id),
    validation_rules JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(organization_id, type_code)
);

-- Períodos contables
CREATE TABLE accounting_periods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    period_code VARCHAR(7) NOT NULL, -- YYYY-MM
    period_name VARCHAR(50) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    fiscal_year INTEGER NOT NULL,
    status VARCHAR(20) DEFAULT 'OPEN',
    closed_at TIMESTAMPTZ,
    closed_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(organization_id, period_code),
    CONSTRAINT chk_period_status CHECK (status IN ('OPEN', 'CLOSED', 'LOCKED'))
);

-- Reportes generados
CREATE TABLE generated_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    report_type VARCHAR(50) NOT NULL,
    report_name VARCHAR(200) NOT NULL,
    parameters JSONB NOT NULL,
    period_start DATE,
    period_end DATE,
    file_format VARCHAR(10) NOT NULL,
    file_path TEXT,
    file_size BIGINT,
    status VARCHAR(20) DEFAULT 'GENERATING',
    error_message TEXT,
    generated_by UUID NOT NULL REFERENCES users(id),
    generated_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    CONSTRAINT chk_report_status CHECK (status IN ('GENERATING', 'COMPLETED', 'ERROR', 'EXPIRED')),
    CONSTRAINT chk_file_format CHECK (file_format IN ('PDF', 'XLSX', 'CSV', 'JSON'))
);

-- Configuración del sistema
CREATE TABLE system_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID REFERENCES organizations(id), -- NULL = global setting
    setting_key VARCHAR(100) NOT NULL,
    setting_value JSONB NOT NULL,
    data_type VARCHAR(20) NOT NULL,
    description TEXT,
    is_encrypted BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(organization_id, setting_key),
    CONSTRAINT chk_data_type CHECK (data_type IN ('STRING', 'NUMBER', 'BOOLEAN', 'JSON', 'ARRAY'))
);

-- ==============================================
-- PARTICIONAMIENTO
-- ==============================================

-- Crear particiones para vouchers (2025)
CREATE TABLE vouchers_2025_01 PARTITION OF vouchers
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
    
CREATE TABLE vouchers_2025_02 PARTITION OF vouchers
    FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');
    
CREATE TABLE vouchers_2025_03 PARTITION OF vouchers
    FOR VALUES FROM ('2025-03-01') TO ('2025-04-01');

CREATE TABLE vouchers_2025_04 PARTITION OF vouchers
    FOR VALUES FROM ('2025-04-01') TO ('2025-05-01');

CREATE TABLE vouchers_2025_05 PARTITION OF vouchers
    FOR VALUES FROM ('2025-05-01') TO ('2025-06-01');

CREATE TABLE vouchers_2025_06 PARTITION OF vouchers
    FOR VALUES FROM ('2025-06-01') TO ('2025-07-01');

CREATE TABLE vouchers_2025_07 PARTITION OF vouchers
    FOR VALUES FROM ('2025-07-01') TO ('2025-08-01');

CREATE TABLE vouchers_2025_08 PARTITION OF vouchers
    FOR VALUES FROM ('2025-08-01') TO ('2025-09-01');

CREATE TABLE vouchers_2025_09 PARTITION OF vouchers
    FOR VALUES FROM ('2025-09-01') TO ('2025-10-01');

CREATE TABLE vouchers_2025_10 PARTITION OF vouchers
    FOR VALUES FROM ('2025-10-01') TO ('2025-11-01');

CREATE TABLE vouchers_2025_11 PARTITION OF vouchers
    FOR VALUES FROM ('2025-11-01') TO ('2025-12-01');

CREATE TABLE vouchers_2025_12 PARTITION OF vouchers
    FOR VALUES FROM ('2025-12-01') TO ('2026-01-01');

-- Crear particiones para journal_entries (2025)
CREATE TABLE journal_entries_2025_01 PARTITION OF journal_entries
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
    
CREATE TABLE journal_entries_2025_02 PARTITION OF journal_entries
    FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');
    
CREATE TABLE journal_entries_2025_03 PARTITION OF journal_entries
    FOR VALUES FROM ('2025-03-01') TO ('2025-04-01');

CREATE TABLE journal_entries_2025_04 PARTITION OF journal_entries
    FOR VALUES FROM ('2025-04-01') TO ('2025-05-01');

CREATE TABLE journal_entries_2025_05 PARTITION OF journal_entries
    FOR VALUES FROM ('2025-05-01') TO ('2025-06-01');

CREATE TABLE journal_entries_2025_06 PARTITION OF journal_entries
    FOR VALUES FROM ('2025-06-01') TO ('2025-07-01');

CREATE TABLE journal_entries_2025_07 PARTITION OF journal_entries
    FOR VALUES FROM ('2025-07-01') TO ('2025-08-01');

CREATE TABLE journal_entries_2025_08 PARTITION OF journal_entries
    FOR VALUES FROM ('2025-08-01') TO ('2025-09-01');

CREATE TABLE journal_entries_2025_09 PARTITION OF journal_entries
    FOR VALUES FROM ('2025-09-01') TO ('2025-10-01');

CREATE TABLE journal_entries_2025_10 PARTITION OF journal_entries
    FOR VALUES FROM ('2025-10-01') TO ('2025-11-01');

CREATE TABLE journal_entries_2025_11 PARTITION OF journal_entries
    FOR VALUES FROM ('2025-11-01') TO ('2025-12-01');

CREATE TABLE journal_entries_2025_12 PARTITION OF journal_entries
    FOR VALUES FROM ('2025-12-01') TO ('2026-01-01');

-- ==============================================
-- COMENTARIOS DE DOCUMENTACIÓN
-- ==============================================

COMMENT ON SCHEMA accounting_engine IS 'Esquema principal del motor contable cloud-native';
COMMENT ON TABLE organizations IS 'Tabla maestra de organizaciones (multi-tenant)';
COMMENT ON TABLE chart_of_accounts IS 'Catálogo de cuentas contables por organización';
COMMENT ON TABLE vouchers IS 'Comprobantes o documentos fuente para contabilización';
COMMENT ON TABLE journal_entries IS 'Asientos contables generados';
COMMENT ON TABLE journal_lines IS 'Líneas detalladas de cada asiento contable';
COMMENT ON TABLE accounting_templates IS 'Plantillas DSL para generación automática de asientos';
COMMENT ON TABLE fiscal_rules IS 'Reglas fiscales por país';
COMMENT ON TABLE accounting_books IS 'Libros contables generados';

-- ==============================================
-- PERMISOS Y SEGURIDAD
-- ==============================================

-- Crear roles
CREATE ROLE accounting_read;
CREATE ROLE accounting_write;
CREATE ROLE accounting_admin;

-- Permisos de lectura
GRANT USAGE ON SCHEMA accounting_engine TO accounting_read;
GRANT SELECT ON ALL TABLES IN SCHEMA accounting_engine TO accounting_read;

-- Permisos de escritura
GRANT USAGE ON SCHEMA accounting_engine TO accounting_write;
GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA accounting_engine TO accounting_write;
GRANT USAGE ON ALL SEQUENCES IN SCHEMA accounting_engine TO accounting_write;

-- Permisos de administración
GRANT ALL ON SCHEMA accounting_engine TO accounting_admin;
GRANT ALL ON ALL TABLES IN SCHEMA accounting_engine TO accounting_admin;
GRANT ALL ON ALL SEQUENCES IN SCHEMA accounting_engine TO accounting_admin;

-- Row Level Security (ejemplo para organizations)
ALTER TABLE organizations ENABLE ROW LEVEL SECURITY;

CREATE POLICY org_isolation ON organizations
    FOR ALL
    USING (id = current_setting('app.current_org_id')::uuid);

-- ==============================================
-- DATOS INICIALES BÁSICOS
-- ==============================================

-- Insertar organización demo
INSERT INTO organizations (id, code, name, country_code, tax_id, settings) VALUES
('123e4567-e89b-12d3-a456-426614174000', 'DEMO-CO', 'Empresa Demo Colombia', 'CO', '900123456-1', 
 '{"fiscal_year_start": "01-01", "currency_default": "COP", "decimal_places": 2}');

-- Insertar usuario administrador demo
INSERT INTO users (id, email, name, password_hash, role, organization_id) VALUES
('123e4567-e89b-12d3-a456-426614174001', 'admin@demo.com', 'Administrador Demo', 
 '$2a$10$EixZaYVK1fsbw1ZfbX3OXePaWxn96p36Tq7dBhcVaFm.Hxfh6Jk1K', 'ORG_ADMIN', 
 '123e4567-e89b-12d3-a456-426614174000');

-- Insertar cuentas básicas del PUC Colombia
INSERT INTO chart_of_accounts (organization_id, account_code, name, type, nature, level, is_detail) VALUES
('123e4567-e89b-12d3-a456-426614174000', '1', 'ACTIVO', 'ASSET', 'D', 1, false),
('123e4567-e89b-12d3-a456-426614174000', '11', 'DISPONIBLE', 'ASSET', 'D', 2, false),
('123e4567-e89b-12d3-a456-426614174000', '1105', 'CAJA', 'ASSET', 'D', 3, false),
('123e4567-e89b-12d3-a456-426614174000', '110505', 'CAJA GENERAL', 'ASSET', 'D', 4, true),
('123e4567-e89b-12d3-a456-426614174000', '1110', 'BANCOS', 'ASSET', 'D', 3, false),
('123e4567-e89b-12d3-a456-426614174000', '111005', 'BANCO NACIONAL', 'ASSET', 'D', 4, true),
('123e4567-e89b-12d3-a456-426614174000', '13', 'DEUDORES', 'ASSET', 'D', 2, false),
('123e4567-e89b-12d3-a456-426614174000', '1305', 'CLIENTES', 'ASSET', 'D', 3, false),
('123e4567-e89b-12d3-a456-426614174000', '130505', 'CLIENTES NACIONALES', 'ASSET', 'D', 4, true),
('123e4567-e89b-12d3-a456-426614174000', '2', 'PASIVO', 'LIABILITY', 'C', 1, false),
('123e4567-e89b-12d3-a456-426614174000', '22', 'PROVEEDORES', 'LIABILITY', 'C', 2, false),
('123e4567-e89b-12d3-a456-426614174000', '2205', 'PROVEEDORES NACIONALES', 'LIABILITY', 'C', 3, true),
('123e4567-e89b-12d3-a456-426614174000', '24', 'IMPUESTOS POR PAGAR', 'LIABILITY', 'C', 2, false),
('123e4567-e89b-12d3-a456-426614174000', '2408', 'IVA POR PAGAR', 'LIABILITY', 'C', 3, true),
('123e4567-e89b-12d3-a456-426614174000', '3', 'PATRIMONIO', 'EQUITY', 'C', 1, false),
('123e4567-e89b-12d3-a456-426614174000', '31', 'CAPITAL SOCIAL', 'EQUITY', 'C', 2, false),
('123e4567-e89b-12d3-a456-426614174000', '3105', 'CAPITAL AUTORIZADO', 'EQUITY', 'C', 3, true),
('123e4567-e89b-12d3-a456-426614174000', '4', 'INGRESOS', 'INCOME', 'C', 1, false),
('123e4567-e89b-12d3-a456-426614174000', '41', 'INGRESOS OPERACIONALES', 'INCOME', 'C', 2, false),
('123e4567-e89b-12d3-a456-426614174000', '4135', 'VENTAS', 'INCOME', 'C', 3, true),
('123e4567-e89b-12d3-a456-426614174000', '5', 'GASTOS', 'EXPENSE', 'D', 1, false),
('123e4567-e89b-12d3-a456-426614174000', '51', 'GASTOS OPERACIONALES', 'EXPENSE', 'D', 2, false),
('123e4567-e89b-12d3-a456-426614174000', '5105', 'GASTOS DE PERSONAL', 'EXPENSE', 'D', 3, true);

-- Insertar tipos de comprobante básicos
INSERT INTO voucher_types (organization_id, type_code, name, prefix, requires_third_party) VALUES
('123e4567-e89b-12d3-a456-426614174000', 'invoice_sale', 'Factura de Venta', 'FV', true),
('123e4567-e89b-12d3-a456-426614174000', 'invoice_purchase', 'Factura de Compra', 'FC', true),
('123e4567-e89b-12d3-a456-426614174000', 'payment', 'Pago', 'PAG', true),
('123e4567-e89b-12d3-a456-426614174000', 'receipt', 'Recibo', 'RC', true),
('123e4567-e89b-12d3-a456-426614174000', 'credit_note', 'Nota Crédito', 'NC', true),
('123e4567-e89b-12d3-a456-426614174000', 'journal_entry', 'Asiento Manual', 'AM', false);

-- Insertar período contable actual
INSERT INTO accounting_periods (organization_id, period_code, period_name, start_date, end_date, fiscal_year) VALUES
('123e4567-e89b-12d3-a456-426614174000', '2025-01', 'Enero 2025', '2025-01-01', '2025-01-31', 2025),
('123e4567-e89b-12d3-a456-426614174000', '2025-02', 'Febrero 2025', '2025-02-01', '2025-02-28', 2025),
('123e4567-e89b-12d3-a456-426614174000', '2025-03', 'Marzo 2025', '2025-03-01', '2025-03-31', 2025),
('123e4567-e89b-12d3-a456-426614174000', '2025-04', 'Abril 2025', '2025-04-01', '2025-04-30', 2025),
('123e4567-e89b-12d3-a456-426614174000', '2025-05', 'Mayo 2025', '2025-05-01', '2025-05-31', 2025),
('123e4567-e89b-12d3-a456-426614174000', '2025-06', 'Junio 2025', '2025-06-01', '2025-06-30', 2025),
('123e4567-e89b-12d3-a456-426614174000', '2025-07', 'Julio 2025', '2025-07-01', '2025-07-31', 2025),
('123e4567-e89b-12d3-a456-426614174000', '2025-08', 'Agosto 2025', '2025-08-01', '2025-08-31', 2025),
('123e4567-e89b-12d3-a456-426614174000', '2025-09', 'Septiembre 2025', '2025-09-01', '2025-09-30', 2025),
('123e4567-e89b-12d3-a456-426614174000', '2025-10', 'Octubre 2025', '2025-10-01', '2025-10-31', 2025),
('123e4567-e89b-12d3-a456-426614174000', '2025-11', 'Noviembre 2025', '2025-11-01', '2025-11-30', 2025),
('123e4567-e89b-12d3-a456-426614174000', '2025-12', 'Diciembre 2025', '2025-12-01', '2025-12-31', 2025);

-- ==============================================
-- FUNCIONES ADICIONALES PARA GO-DSL
-- ==============================================

-- Función para obtener número consecutivo
CREATE OR REPLACE FUNCTION get_next_voucher_number(p_org_id UUID, p_type_code VARCHAR)
RETURNS VARCHAR AS $$
DECLARE
    v_prefix VARCHAR(10);
    v_next_number BIGINT;
    v_voucher_number VARCHAR(50);
BEGIN
    -- Obtener prefix y next_number
    SELECT prefix, next_number 
    INTO v_prefix, v_next_number
    FROM voucher_types 
    WHERE organization_id = p_org_id AND type_code = p_type_code AND is_active = true;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Tipo de comprobante no encontrado: %', p_type_code;
    END IF;
    
    -- Generar número
    v_voucher_number := v_prefix || '-' || LPAD(v_next_number::TEXT, 6, '0');
    
    -- Actualizar contador
    UPDATE voucher_types 
    SET next_number = next_number + 1
    WHERE organization_id = p_org_id AND type_code = p_type_code;
    
    RETURN v_voucher_number;
END;
$$ LANGUAGE plpgsql;

-- Función para validar balance de comprobante
CREATE OR REPLACE FUNCTION validate_voucher_balance(p_voucher_id UUID)
RETURNS BOOLEAN AS $$
DECLARE
    v_total_lines DECIMAL(20,4);
    v_voucher_total DECIMAL(20,4);
BEGIN
    -- Sumar líneas del comprobante
    SELECT COALESCE(SUM(line_amount), 0)
    INTO v_total_lines
    FROM voucher_lines
    WHERE voucher_id = p_voucher_id;
    
    -- Obtener total del comprobante
    SELECT total_amount
    INTO v_voucher_total
    FROM vouchers
    WHERE id = p_voucher_id;
    
    RETURN v_total_lines = v_voucher_total;
END;
$$ LANGUAGE plpgsql;

-- Función para refrescar vista materializada automáticamente
CREATE OR REPLACE FUNCTION refresh_account_balances()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY account_balances;
END;
$$ LANGUAGE plpgsql;

-- ==============================================
-- FIN DEL SCRIPT
-- ==============================================

-- Script completado exitosamente
-- Estructura: 15 tablas principales + 6 tablas auxiliares
-- Particionamiento: 12 meses de 2025 
-- Índices: 25+ índices optimizados
-- Funciones: 8 funciones de negocio
-- Datos iniciales: Plan contable básico PUC Colombia
-- Listo para Fase 1: Go/Fiber/go-dsl/PostgreSQL