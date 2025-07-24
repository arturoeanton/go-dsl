package models

import (
	"encoding/json"
	"time"
)

// UserPreferences representa las preferencias de usuario del sistema contable
type UserPreferences struct {
	BaseModel
	OrganizationID string    `json:"organization_id" gorm:"index;not null"`
	UserID         string    `json:"user_id" gorm:"index;not null"`
	Category       string    `json:"category" gorm:"not null;index"` // UI, ACCOUNTING, REPORTING, NOTIFICATIONS, etc.
	PreferencesJSON string   `json:"-" gorm:"type:text;column:preferences"`
	LastUpdated    time.Time `json:"last_updated" gorm:"not null"`
}

// UIPreferences preferencias de interfaz de usuario
type UIPreferences struct {
	Theme           string `json:"theme"`            // light, dark, auto
	Language        string `json:"language"`         // es, en
	DateFormat      string `json:"date_format"`      // DD/MM/YYYY, MM/DD/YYYY, YYYY-MM-DD
	NumberFormat    string `json:"number_format"`    // 1,234.56, 1.234,56, 1 234,56
	TimeZone        string `json:"timezone"`         // America/Bogota, etc.
	SidebarCollapsed bool  `json:"sidebar_collapsed"`
	TablePageSize   int    `json:"table_page_size"`  // 10, 20, 50, 100
	DefaultView     string `json:"default_view"`     // dashboard, accounts, vouchers
}

// AccountingPreferences preferencias contables
type AccountingPreferences struct {
	DefaultCurrency        string  `json:"default_currency"`         // COP, USD, EUR
	DecimalPlaces          int     `json:"decimal_places"`           // 2, 4
	RoundingMethod         string  `json:"rounding_method"`          // ROUND_HALF_UP, ROUND_HALF_DOWN
	ShowAccountCodes       bool    `json:"show_account_codes"`       // true, false
	RequireDescriptions    bool    `json:"require_descriptions"`     // true, false
	AutoGenerateNumbers    bool    `json:"auto_generate_numbers"`    // true, false
	DefaultTaxRate         float64 `json:"default_tax_rate"`         // 19.0
	AllowNegativeBalances  bool    `json:"allow_negative_balances"`  // true, false
	DefaultCostCenter      string  `json:"default_cost_center"`
	DefaultPaymentTerms    int     `json:"default_payment_terms"`    // 30 días
}

// ReportingPreferences preferencias de reportes
type ReportingPreferences struct {
	DefaultPeriod        string   `json:"default_period"`         // current, last_month, last_quarter
	IncludeInactiveAccounts bool  `json:"include_inactive_accounts"`
	ShowZeroBalances     bool     `json:"show_zero_balances"`
	DefaultExportFormat  string   `json:"default_export_format"`  // PDF, EXCEL, CSV
	EmailReports         bool     `json:"email_reports"`
	ReportSchedule       string   `json:"report_schedule"`        // daily, weekly, monthly
	FavoriteReports      []string `json:"favorite_reports"`
}

// NotificationPreferences preferencias de notificaciones
type NotificationPreferences struct {
	EmailNotifications    bool     `json:"email_notifications"`
	PushNotifications     bool     `json:"push_notifications"`
	EmailAddress          string   `json:"email_address"`
	NotifyOnVoucherPost   bool     `json:"notify_on_voucher_post"`
	NotifyOnPeriodClose   bool     `json:"notify_on_period_close"`
	NotifyOnErrors        bool     `json:"notify_on_errors"`
	NotifyOnBackups       bool     `json:"notify_on_backups"`
	DigestFrequency       string   `json:"digest_frequency"`       // daily, weekly, monthly, never
	AllowedNotificationTypes []string `json:"allowed_notification_types"`
}

// DashboardPreferences preferencias del dashboard
type DashboardPreferences struct {
	DefaultWidgets    []string               `json:"default_widgets"`
	WidgetLayout      map[string]interface{} `json:"widget_layout"`
	RefreshInterval   int                    `json:"refresh_interval"`   // segundos
	ShowKPIs          bool                   `json:"show_kpis"`
	ShowCharts        bool                   `json:"show_charts"`
	ShowRecentActivity bool                  `json:"show_recent_activity"`
	FavoriteCharts    []string               `json:"favorite_charts"`
}

// GetPreferences deserializa las preferencias según la categoría
func (up *UserPreferences) GetPreferences() (interface{}, error) {
	if up.PreferencesJSON == "" {
		return up.getDefaultPreferences(), nil
	}

	switch up.Category {
	case "UI":
		var prefs UIPreferences
		err := json.Unmarshal([]byte(up.PreferencesJSON), &prefs)
		return &prefs, err
	case "ACCOUNTING":
		var prefs AccountingPreferences
		err := json.Unmarshal([]byte(up.PreferencesJSON), &prefs)
		return &prefs, err
	case "REPORTING":
		var prefs ReportingPreferences
		err := json.Unmarshal([]byte(up.PreferencesJSON), &prefs)
		return &prefs, err
	case "NOTIFICATIONS":
		var prefs NotificationPreferences
		err := json.Unmarshal([]byte(up.PreferencesJSON), &prefs)
		return &prefs, err
	case "DASHBOARD":
		var prefs DashboardPreferences
		err := json.Unmarshal([]byte(up.PreferencesJSON), &prefs)
		return &prefs, err
	default:
		var prefs map[string]interface{}
		err := json.Unmarshal([]byte(up.PreferencesJSON), &prefs)
		return prefs, err
	}
}

// SetPreferences serializa las preferencias
func (up *UserPreferences) SetPreferences(preferences interface{}) error {
	data, err := json.Marshal(preferences)
	if err != nil {
		return err
	}
	up.PreferencesJSON = string(data)
	up.LastUpdated = time.Now()
	return nil
}

// getDefaultPreferences retorna las preferencias por defecto según la categoría
func (up *UserPreferences) getDefaultPreferences() interface{} {
	switch up.Category {
	case "UI":
		return &UIPreferences{
			Theme:           "light",
			Language:        "es",
			DateFormat:      "DD/MM/YYYY",
			NumberFormat:    "1.234,56",
			TimeZone:        "America/Bogota",
			SidebarCollapsed: false,
			TablePageSize:   20,
			DefaultView:     "dashboard",
		}
	case "ACCOUNTING":
		return &AccountingPreferences{
			DefaultCurrency:       "COP",
			DecimalPlaces:         2,
			RoundingMethod:        "ROUND_HALF_UP",
			ShowAccountCodes:      true,
			RequireDescriptions:   true,
			AutoGenerateNumbers:   true,
			DefaultTaxRate:        19.0,
			AllowNegativeBalances: false,
			DefaultPaymentTerms:   30,
		}
	case "REPORTING":
		return &ReportingPreferences{
			DefaultPeriod:           "current",
			IncludeInactiveAccounts: false,
			ShowZeroBalances:        false,
			DefaultExportFormat:     "PDF",
			EmailReports:            false,
			ReportSchedule:          "monthly",
			FavoriteReports:         []string{},
		}
	case "NOTIFICATIONS":
		return &NotificationPreferences{
			EmailNotifications:       true,
			PushNotifications:        true,
			NotifyOnVoucherPost:      true,
			NotifyOnPeriodClose:      true,
			NotifyOnErrors:           true,
			NotifyOnBackups:          false,
			DigestFrequency:          "weekly",
			AllowedNotificationTypes: []string{"voucher", "period", "error"},
		}
	case "DASHBOARD":
		return &DashboardPreferences{
			DefaultWidgets:     []string{"kpis", "recent_activity", "charts"},
			WidgetLayout:       map[string]interface{}{},
			RefreshInterval:    300, // 5 minutos
			ShowKPIs:           true,
			ShowCharts:         true,
			ShowRecentActivity: true,
			FavoriteCharts:     []string{"vouchers_by_month", "accounts_summary"},
		}
	default:
		return map[string]interface{}{}
	}
}

// UserPreferencesDetail estructura completa para respuestas detalladas
type UserPreferencesDetail struct {
	*UserPreferences
	Preferences interface{} `json:"preferences"`
}

// ToDetail convierte UserPreferences a UserPreferencesDetail con preferencias deserializadas
func (up *UserPreferences) ToDetail() (*UserPreferencesDetail, error) {
	preferences, err := up.GetPreferences()
	if err != nil {
		return nil, err
	}

	return &UserPreferencesDetail{
		UserPreferences: up,
		Preferences:     preferences,
	}, nil
}

// UserPreferencesUpdateRequest estructura para actualizar preferencias
type UserPreferencesUpdateRequest struct {
	Category    string      `json:"category" binding:"required"`
	Preferences interface{} `json:"preferences" binding:"required"`
}

// UserPreferencesListResponse respuesta para listado de preferencias
type UserPreferencesListResponse struct {
	Preferences []UserPreferencesDetail `json:"preferences"`
	Categories  []string                `json:"categories"`
}

// GetAllUserPreferences estructura para obtener todas las preferencias de un usuario
type AllUserPreferences struct {
	UI            *UIPreferences            `json:"ui,omitempty"`
	Accounting    *AccountingPreferences    `json:"accounting,omitempty"`
	Reporting     *ReportingPreferences     `json:"reporting,omitempty"`
	Notifications *NotificationPreferences  `json:"notifications,omitempty"`
	Dashboard     *DashboardPreferences     `json:"dashboard,omitempty"`
	Custom        map[string]interface{}    `json:"custom,omitempty"`
}