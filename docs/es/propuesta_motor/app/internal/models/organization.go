package models

import "encoding/json"

// Organization representa una organización (tenant) en el sistema multi-tenant
// Basado en el schema del Swagger y los mocks JSON
type Organization struct {
	BaseModel
	Code                    string          `json:"code" gorm:"uniqueIndex;not null"`
	Name                    string          `json:"name" gorm:"not null"`
	CommercialName          string          `json:"commercial_name"`
	DocumentType            string          `json:"document_type"`
	TaxID                   string          `json:"tax_id" gorm:"uniqueIndex"`
	CountryCode             string          `json:"country_code" gorm:"not null;default:'CO'"`
	CurrencyDefault         string          `json:"currency_default" gorm:"default:'COP'"`
	Language                string          `json:"language" gorm:"default:'es'"`
	Timezone                string          `json:"timezone" gorm:"default:'America/Bogota'"`
	ContactInfoJSON         string          `json:"-" gorm:"type:text;column:contact_info"`
	FiscalInfoJSON          string          `json:"-" gorm:"type:text;column:fiscal_info"`
	AccountingConfigJSON    string          `json:"-" gorm:"type:text;column:accounting_configuration"`
	DSLConfigJSON           string          `json:"-" gorm:"type:text;column:dsl_configuration"`
	UserPermissionsJSON     string          `json:"-" gorm:"type:text;column:user_permissions"`
	SubscriptionInfoJSON    string          `json:"-" gorm:"type:text;column:subscription_info"`
	SystemSettingsJSON      string          `json:"-" gorm:"type:text;column:system_settings"`
	StatisticsJSON          string          `json:"-" gorm:"type:text;column:statistics"`
	IsActive                bool            `json:"is_active" gorm:"default:true"`
}

// ContactInfo estructura para información de contacto
type ContactInfo struct {
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Address    string `json:"address"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Website    string `json:"website"`
}

// FiscalInfo estructura para información fiscal
type FiscalInfo struct {
	FiscalYearStart    string `json:"fiscal_year_start"`
	FiscalYearEnd      string `json:"fiscal_year_end"`
	TaxRegime          string `json:"tax_regime"`
	AccountingStandard string `json:"accounting_standard"`
	ReportingCurrency  string `json:"reporting_currency"`
	DecimalPlaces      int    `json:"decimal_places"`
	RoundingMethod     string `json:"rounding_method"`
}

// AccountingConfig configuración contable
type AccountingConfig struct {
	CurrentPeriod            string  `json:"current_period"`
	LastClosedPeriod         string  `json:"last_closed_period"`
	AutoNumbering            bool    `json:"auto_numbering"`
	RequireBalancedEntries   bool    `json:"require_balanced_entries"`
	AllowNegativeInventory   bool    `json:"allow_negative_inventory"`
	DefaultPaymentTerms      int     `json:"default_payment_terms"`
	DefaultTaxRate           float64 `json:"default_tax_rate"`
}

// DSLConfig configuración del motor DSL
type DSLConfig struct {
	AutoProcessVouchers bool   `json:"auto_process_vouchers"`
	ValidationLevel     string `json:"validation_level"`
	ErrorHandling       string `json:"error_handling"`
	TemplateCacheEnabled bool  `json:"template_cache_enabled"`
	ParallelProcessing  bool   `json:"parallel_processing"`
}

// UserPermissions permisos del usuario actual
type UserPermissions struct {
	CanEditOrganization   bool `json:"can_edit_organization"`
	CanManageUsers        bool `json:"can_manage_users"`
	CanClosePeriods       bool `json:"can_close_periods"`
	CanEditDSLTemplates   bool `json:"can_edit_dsl_templates"`
	CanViewAllData        bool `json:"can_view_all_data"`
}

// SubscriptionInfo información de suscripción
type SubscriptionInfo struct {
	Plan                   string   `json:"plan"`
	Status                 string   `json:"status"`
	ExpiresAt             string   `json:"expires_at"`
	MaxUsers              int      `json:"max_users"`
	CurrentUsers          int      `json:"current_users"`
	MaxMonthlyVouchers    int      `json:"max_monthly_vouchers"`
	CurrentMonthVouchers  int      `json:"current_month_vouchers"`
	Features              []string `json:"features"`
}

// SystemSettings configuración del sistema
type SystemSettings struct {
	BackupFrequency     string                 `json:"backup_frequency"`
	RetentionPeriodDays int                   `json:"retention_period_days"`
	AuditLevel         string                 `json:"audit_level"`
	NotificationPrefs  map[string]interface{} `json:"notification_preferences"`
}

// OrganizationStatistics estadísticas de la organización
type OrganizationStatistics struct {
	TotalVouchers      int `json:"total_vouchers"`
	TotalJournalEntries int `json:"total_journal_entries"`
	TotalAccounts      int `json:"total_accounts"`
	ActiveUsers        int `json:"active_users"`
	StorageUsedMB      int `json:"storage_used_mb"`
	APICallsMonth      int `json:"api_calls_month"`
}

// GetContactInfo deserializa la información de contacto
func (o *Organization) GetContactInfo() (*ContactInfo, error) {
	if o.ContactInfoJSON == "" {
		return &ContactInfo{}, nil
	}
	var info ContactInfo
	err := json.Unmarshal([]byte(o.ContactInfoJSON), &info)
	return &info, err
}

// SetContactInfo serializa la información de contacto
func (o *Organization) SetContactInfo(info ContactInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	o.ContactInfoJSON = string(data)
	return nil
}

// GetFiscalInfo deserializa la información fiscal
func (o *Organization) GetFiscalInfo() (*FiscalInfo, error) {
	if o.FiscalInfoJSON == "" {
		return &FiscalInfo{}, nil
	}
	var info FiscalInfo
	err := json.Unmarshal([]byte(o.FiscalInfoJSON), &info)
	return &info, err
}

// SetFiscalInfo serializa la información fiscal
func (o *Organization) SetFiscalInfo(info FiscalInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	o.FiscalInfoJSON = string(data)
	return nil
}

// GetAccountingConfig deserializa la configuración contable
func (o *Organization) GetAccountingConfig() (*AccountingConfig, error) {
	if o.AccountingConfigJSON == "" {
		return &AccountingConfig{}, nil
	}
	var config AccountingConfig
	err := json.Unmarshal([]byte(o.AccountingConfigJSON), &config)
	return &config, err
}

// SetAccountingConfig serializa la configuración contable
func (o *Organization) SetAccountingConfig(config AccountingConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	o.AccountingConfigJSON = string(data)
	return nil
}

// GetDSLConfig deserializa la configuración DSL
func (o *Organization) GetDSLConfig() (*DSLConfig, error) {
	if o.DSLConfigJSON == "" {
		return &DSLConfig{}, nil
	}
	var config DSLConfig
	err := json.Unmarshal([]byte(o.DSLConfigJSON), &config)
	return &config, err
}

// SetDSLConfig serializa la configuración DSL
func (o *Organization) SetDSLConfig(config DSLConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	o.DSLConfigJSON = string(data)
	return nil
}

// OrganizationDetail estructura completa para respuestas detalladas
// Combina Organization con todas sus configuraciones deserializadas
type OrganizationDetail struct {
	*Organization
	ContactInfo          *ContactInfo          `json:"contact_info"`
	FiscalInfo           *FiscalInfo           `json:"fiscal_info"`
	AccountingConfig     *AccountingConfig     `json:"accounting_configuration"`
	DSLConfig            *DSLConfig            `json:"dsl_configuration"`
	UserPermissions      *UserPermissions      `json:"user_permissions"`
	SubscriptionInfo     *SubscriptionInfo     `json:"subscription_info"`
	SystemSettings       *SystemSettings       `json:"system_settings"`
	Statistics           *OrganizationStatistics `json:"statistics"`
}

// ToDetail convierte una Organization a OrganizationDetail con todas las configuraciones
func (o *Organization) ToDetail() (*OrganizationDetail, error) {
	contactInfo, err := o.GetContactInfo()
	if err != nil {
		return nil, err
	}
	
	fiscalInfo, err := o.GetFiscalInfo()
	if err != nil {
		return nil, err
	}
	
	accountingConfig, err := o.GetAccountingConfig()
	if err != nil {
		return nil, err
	}
	
	dslConfig, err := o.GetDSLConfig()
	if err != nil {
		return nil, err
	}

	// TODO: Deserializar otros campos JSON cuando se implementen
	// Por ahora devolvemos valores por defecto
	
	return &OrganizationDetail{
		Organization:     o,
		ContactInfo:      contactInfo,
		FiscalInfo:       fiscalInfo,
		AccountingConfig: accountingConfig,
		DSLConfig:        dslConfig,
		UserPermissions: &UserPermissions{
			CanEditOrganization: true,
			CanManageUsers:      true,
			CanClosePeriods:     true,
			CanEditDSLTemplates: true,
			CanViewAllData:      true,
		},
		SubscriptionInfo: &SubscriptionInfo{
			Plan:                 "ENTERPRISE",
			Status:               "ACTIVE",
			ExpiresAt:           "2025-12-31T23:59:59Z",
			MaxUsers:            25,
			CurrentUsers:        8,
			MaxMonthlyVouchers:  100000,
			CurrentMonthVouchers: 3421,
			Features: []string{
				"MULTI_COMPANY",
				"ADVANCED_REPORTING",
				"DSL_EDITOR",
				"API_ACCESS",
				"AUDIT_TRAIL",
				"CUSTOM_TEMPLATES",
			},
		},
		SystemSettings: &SystemSettings{
			BackupFrequency:     "DAILY",
			RetentionPeriodDays: 2555,
			AuditLevel:         "FULL",
			NotificationPrefs: map[string]interface{}{
				"email_reports":    true,
				"email_errors":     true,
				"email_reminders":  true,
			},
		},
		Statistics: &OrganizationStatistics{
			TotalVouchers:       15420,
			TotalJournalEntries: 28945,
			TotalAccounts:       156,
			ActiveUsers:         8,
			StorageUsedMB:       2048,
			APICallsMonth:       15420,
		},
	}, nil
}