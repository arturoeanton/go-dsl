package database

import (
	"motor-contable-poc/internal/models"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB instancia global de la base de datos
var DB *gorm.DB

// InitDatabase inicializa la conexión a SQLite y ejecuta migraciones
func InitDatabase() error {
	var err error
	
	// Configurar SQLite con opciones optimizadas
	DB, err = gorm.Open(sqlite.Open("db_contable.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	
	if err != nil {
		return err
	}
	
	// Configuraciones específicas de SQLite para mejor rendimiento
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	
	// Configurar pool de conexiones
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	
	// Ejecutar migraciones automáticas
	err = runMigrations()
	if err != nil {
		return err
	}
	
	return nil
}

// runMigrations ejecuta todas las migraciones automáticas de GORM
func runMigrations() error {
	// Migración automática de todos los modelos
	err := DB.AutoMigrate(
		// Modelos principales
		&models.Organization{},
		&models.Account{},
		&models.ThirdParty{},
		&models.Period{},
		&models.TaxType{},
		&models.DocumentType{},
		&models.CostCenter{},
		
		// Modelos de comprobantes y asientos
		&models.Voucher{},
		&models.VoucherLine{},
		&models.JournalEntry{},
		&models.JournalLine{},
		
		// Modelos DSL
		&models.DSLTemplate{},
		&models.DSLExecution{},
		
		// Modelos de Templates
		&models.JournalTemplate{},
		&models.TemplateExecution{},
		
		// Modelos de auditoría y preferencias
		&models.AuditLog{},
		&models.UserPreferences{},
	)
	
	if err != nil {
		return err
	}
	
	// Ejecutar índices personalizados y configuraciones adicionales
	err = createCustomIndexes()
	if err != nil {
		return err
	}
	
	return nil
}

// createCustomIndexes crea índices personalizados para optimizar consultas
func createCustomIndexes() error {
	// Índices compuestos para optimizar consultas frecuentes
	
	// Índice para búsquedas de cuentas por organización y código
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_accounts_org_code ON accounts(organization_id, code)")
	
	// Índice para búsquedas de terceros por organización y documento
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_third_parties_org_document ON third_parties(organization_id, document_type, document_number)")
	
	// Índice para búsquedas de comprobantes por organización y fecha
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_vouchers_org_date ON vouchers(organization_id, date DESC)")
	
	// Índice para búsquedas de asientos por organización y fecha
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_journal_entries_org_date ON journal_entries(organization_id, date DESC)")
	
	// Índice para líneas de comprobante por cuenta
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_voucher_lines_account ON voucher_lines(account_id)")
	
	// Índice para líneas de asiento por cuenta
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_journal_lines_account ON journal_lines(account_id)")
	
	// Índice para auditoría por entidad
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_audit_logs_entity ON audit_logs(entity_type, entity_id, timestamp DESC)")
	
	// Índice para plantillas DSL por categoría y estado
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_dsl_templates_category_status ON dsl_templates(category, status)")
	
	return nil
}

// GetDB retorna la instancia de la base de datos
func GetDB() *gorm.DB {
	return DB
}

// CloseDatabase cierra la conexión a la base de datos
func CloseDatabase() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// DatabaseHealthCheck verifica el estado de la base de datos
func DatabaseHealthCheck() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// SeedData carga datos de demostración en la base de datos
// TODO: En el futuro, aquí se usaría go-dsl para generar datos de prueba
// basados en reglas DSL y plantillas contables colombianas
func SeedData() error {
	// Verificar si ya existen datos
	var count int64
	DB.Model(&models.Organization{}).Count(&count)
	if count > 0 {
		// Ya hay datos, no hacer seed
		return nil
	}
	
	// Crear organización de demo
	org := &models.Organization{
		Code:            "DEMO001",
		Name:            "Empresa Demo S.A.S",
		CommercialName:  "Demo Contabilidad",
		DocumentType:    "NIT",
		TaxID:           "900123456-7",
		CountryCode:     "CO",
		CurrencyDefault: "COP",
		Language:        "es",
		Timezone:        "America/Bogota",
		IsActive:        true,
	}
	
	// Configurar información de contacto
	contactInfo := models.ContactInfo{
		Email:      "demo@empresa.com",
		Phone:      "+57 1 234 5678",
		Address:    "Carrera 7 # 12-34",
		City:       "Bogotá",
		State:      "Cundinamarca",
		PostalCode: "110111",
		Website:    "https://www.empresa.com",
	}
	org.SetContactInfo(contactInfo)
	
	// Configurar información fiscal
	fiscalInfo := models.FiscalInfo{
		FiscalYearStart:    "2024-01-01",
		FiscalYearEnd:      "2024-12-31",
		TaxRegime:          "RESPONSABLE_IVA",
		AccountingStandard: "NIIF_PYMES",
		ReportingCurrency:  "COP",
		DecimalPlaces:      2,
		RoundingMethod:     "ROUND_HALF_UP",
	}
	org.SetFiscalInfo(fiscalInfo)
	
	// Configurar configuración contable
	accountingConfig := models.AccountingConfig{
		CurrentPeriod:            "2024-07",
		LastClosedPeriod:         "2024-06",
		AutoNumbering:            true,
		RequireBalancedEntries:   true,
		AllowNegativeInventory:   false,
		DefaultPaymentTerms:      30,
		DefaultTaxRate:           19.0,
	}
	org.SetAccountingConfig(accountingConfig)
	
	// Crear la organización
	if err := DB.Create(org).Error; err != nil {
		return err
	}
	
	// Crear período actual
	period := &models.Period{
		OrganizationID: org.ID,
		Code:           "2024-07",
		Name:           "Julio 2024",
		PeriodType:     "MONTHLY",
		Year:           2024,
		Month:          func(i int) *int { return &i }(7),
		StartDate:      mustParseDate("2024-07-01"),
		EndDate:        mustParseDate("2024-07-31"),
		Status:         "OPEN",
		IsCurrent:      true,
		IsActive:       true,
	}
	
	if err := DB.Create(period).Error; err != nil {
		return err
	}
	
	// Crear cuentas básicas del PUC
	accounts := []models.Account{
		{OrganizationID: org.ID, Code: "1", Name: "ACTIVO", AccountType: "ASSET", Level: 1, NaturalBalance: "D", AcceptsMovement: false, PUCCode: "1", IsActive: true},
		{OrganizationID: org.ID, Code: "11", Name: "DISPONIBLE", AccountType: "ASSET", Level: 2, NaturalBalance: "D", AcceptsMovement: false, PUCCode: "11", IsActive: true},
		{OrganizationID: org.ID, Code: "1105", Name: "CAJA", AccountType: "ASSET", Level: 3, NaturalBalance: "D", AcceptsMovement: false, PUCCode: "1105", IsActive: true},
		{OrganizationID: org.ID, Code: "110505", Name: "CAJA GENERAL", AccountType: "ASSET", Level: 4, NaturalBalance: "D", AcceptsMovement: true, PUCCode: "110505", IsActive: true},
		{OrganizationID: org.ID, Code: "1110", Name: "BANCOS", AccountType: "ASSET", Level: 3, NaturalBalance: "D", AcceptsMovement: false, PUCCode: "1110", IsActive: true},
		{OrganizationID: org.ID, Code: "111005", Name: "BANCO BOGOTÁ", AccountType: "ASSET", Level: 4, NaturalBalance: "D", AcceptsMovement: true, PUCCode: "111005", IsActive: true},
		
		{OrganizationID: org.ID, Code: "2", Name: "PASIVO", AccountType: "LIABILITY", Level: 1, NaturalBalance: "C", AcceptsMovement: false, PUCCode: "2", IsActive: true},
		{OrganizationID: org.ID, Code: "24", Name: "IMPUESTOS GRAVÁMENES Y TASAS", AccountType: "LIABILITY", Level: 2, NaturalBalance: "C", AcceptsMovement: false, PUCCode: "24", IsActive: true},
		{OrganizationID: org.ID, Code: "2408", Name: "IMPUESTO A LAS VENTAS POR PAGAR", AccountType: "LIABILITY", Level: 3, NaturalBalance: "C", AcceptsMovement: true, PUCCode: "2408", IsActive: true},
		
		{OrganizationID: org.ID, Code: "3", Name: "PATRIMONIO", AccountType: "EQUITY", Level: 1, NaturalBalance: "C", AcceptsMovement: false, PUCCode: "3", IsActive: true},
		{OrganizationID: org.ID, Code: "31", Name: "CAPITAL SOCIAL", AccountType: "EQUITY", Level: 2, NaturalBalance: "C", AcceptsMovement: false, PUCCode: "31", IsActive: true},
		{OrganizationID: org.ID, Code: "3105", Name: "CAPITAL SUSCRITO Y PAGADO", AccountType: "EQUITY", Level: 3, NaturalBalance: "C", AcceptsMovement: true, PUCCode: "3105", IsActive: true},
		
		{OrganizationID: org.ID, Code: "4", Name: "INGRESOS", AccountType: "INCOME", Level: 1, NaturalBalance: "C", AcceptsMovement: false, PUCCode: "4", IsActive: true},
		{OrganizationID: org.ID, Code: "41", Name: "OPERACIONALES", AccountType: "INCOME", Level: 2, NaturalBalance: "C", AcceptsMovement: false, PUCCode: "41", IsActive: true},
		{OrganizationID: org.ID, Code: "4135", Name: "COMERCIO AL POR MAYOR Y AL POR MENOR", AccountType: "INCOME", Level: 3, NaturalBalance: "C", AcceptsMovement: true, PUCCode: "4135", IsActive: true},
		
		{OrganizationID: org.ID, Code: "5", Name: "GASTOS", AccountType: "EXPENSE", Level: 1, NaturalBalance: "D", AcceptsMovement: false, PUCCode: "5", IsActive: true},
		{OrganizationID: org.ID, Code: "51", Name: "OPERACIONALES DE ADMINISTRACIÓN", AccountType: "EXPENSE", Level: 2, NaturalBalance: "D", AcceptsMovement: false, PUCCode: "51", IsActive: true},
		{OrganizationID: org.ID, Code: "5105", Name: "GASTOS DE PERSONAL", AccountType: "EXPENSE", Level: 3, NaturalBalance: "D", AcceptsMovement: false, PUCCode: "5105", IsActive: true},
		{OrganizationID: org.ID, Code: "510506", Name: "SUELDOS", AccountType: "EXPENSE", Level: 4, NaturalBalance: "D", AcceptsMovement: true, PUCCode: "510506", IsActive: true},
	}
	
	for _, account := range accounts {
		if err := DB.Create(&account).Error; err != nil {
			return err
		}
	}
	
	// Crear algunos terceros de demo
	thirdParties := []models.ThirdParty{
		{
			OrganizationID: org.ID,
			Code:           "CLI001",
			DocumentType:   "CC",
			DocumentNumber: "12345678",
			FirstName:      "Juan",
			LastName:       "Pérez",
			PersonType:     "NATURAL",
			ThirdPartyType: "CUSTOMER",
			TaxpayerType:   "NO_RESPONSABLE_IVA",
			IsActive:       true,
		},
		{
			OrganizationID: org.ID,
			Code:           "PROV001",
			DocumentType:   "NIT",
			DocumentNumber: "800123456",
			VerificationDigit: func(s string) *string { return &s }("7"),
			CompanyName:    "Proveedor Demo S.A.S",
			PersonType:     "JURIDICA",
			ThirdPartyType: "SUPPLIER",
			TaxpayerType:   "RESPONSABLE_IVA",
			IsActive:       true,
		},
	}
	
	for _, thirdParty := range thirdParties {
		if err := DB.Create(&thirdParty).Error; err != nil {
			return err
		}
	}
	
	// Crear tipos de documento
	documentTypes := []models.DocumentType{
		{OrganizationID: org.ID, Code: "CC", Name: "Cédula de Ciudadanía", Category: "IDENTITY", IsForPersons: true, IsActive: true},
		{OrganizationID: org.ID, Code: "NIT", Name: "Número de Identificación Tributaria", Category: "IDENTITY", IsForCompanies: true, RequiresVerificationDigit: true, IsActive: true},
		{OrganizationID: org.ID, Code: "CE", Name: "Cédula de Extranjería", Category: "IDENTITY", IsForPersons: true, IsActive: true},
		{OrganizationID: org.ID, Code: "PP", Name: "Pasaporte", Category: "IDENTITY", IsForPersons: true, IsActive: true},
		{OrganizationID: org.ID, Code: "FV", Name: "Factura de Venta", Category: "INVOICE", IsActive: true},
		{OrganizationID: org.ID, Code: "RC", Name: "Recibo de Caja", Category: "VOUCHER", IsActive: true},
	}
	
	for _, docType := range documentTypes {
		if err := DB.Create(&docType).Error; err != nil {
			return err
		}
	}
	
	// Crear tipos de impuesto
	taxTypes := []models.TaxType{
		{OrganizationID: org.ID, Code: "IVA19", Name: "IVA 19%", TaxCategory: "IVA", Rate: 19.0, IsActive: true},
		{OrganizationID: org.ID, Code: "IVA5", Name: "IVA 5%", TaxCategory: "IVA", Rate: 5.0, IsActive: true},
		{OrganizationID: org.ID, Code: "IVA0", Name: "IVA 0%", TaxCategory: "IVA", Rate: 0.0, IsActive: true},
		{OrganizationID: org.ID, Code: "RETEIVA", Name: "Retención en la Fuente IVA", TaxCategory: "RETENCION", Rate: 15.0, IsActive: true},
		{OrganizationID: org.ID, Code: "RETEFTE", Name: "Retención en la Fuente", TaxCategory: "RETENCION", Rate: 3.5, IsActive: true},
	}
	
	for _, taxType := range taxTypes {
		if err := DB.Create(&taxType).Error; err != nil {
			return err
		}
	}
	
	return nil
}

// mustParseDate helper para parsear fechas en el seed
func mustParseDate(dateStr string) time.Time {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		panic(err)
	}
	return date
}