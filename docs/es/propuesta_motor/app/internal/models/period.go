package models

import "time"

// Period representa un período contable
// Basado en las prácticas contables colombianas y requerimientos de cierre
type Period struct {
	BaseModel
	OrganizationID string    `json:"organization_id" gorm:"index;not null"`
	Code           string    `json:"code" gorm:"uniqueIndex:idx_org_period;not null"`
	Name           string    `json:"name" gorm:"not null"`
	PeriodType     string    `json:"period_type" gorm:"not null"` // MONTHLY, QUARTERLY, YEARLY
	Year           int       `json:"year" gorm:"not null;index"`
	Month          *int      `json:"month" gorm:"index"` // NULL para períodos anuales
	Quarter        *int      `json:"quarter" gorm:"index"` // NULL para períodos mensuales/anuales
	StartDate      time.Time `json:"start_date" gorm:"not null;index"`
	EndDate        time.Time `json:"end_date" gorm:"not null;index"`
	Status         string    `json:"status" gorm:"default:'OPEN'"` // OPEN, CLOSED, LOCKED
	IsCurrent      bool      `json:"is_current" gorm:"default:false;index"`
	ClosedAt       *time.Time `json:"closed_at"`
	ClosedByUserID *string   `json:"closed_by_user_id" gorm:"index"`
	IsActive       bool      `json:"is_active" gorm:"default:true"`
}

// PeriodDetail estructura completa para respuestas detalladas
type PeriodDetail struct {
	*Period
	VouchersCount      int     `json:"vouchers_count"`
	JournalEntriesCount int    `json:"journal_entries_count"`
	TotalMovements     float64 `json:"total_movements"`
	CanBeClosed        bool    `json:"can_be_closed"`
	ClosingNotes       string  `json:"closing_notes"`
}

// CanClose verifica si el período puede ser cerrado
func (p *Period) CanClose() bool {
	return p.Status == "OPEN" && time.Now().After(p.EndDate)
}

// IsCurrentPeriod verifica si es el período actual
func (p *Period) IsCurrentPeriod() bool {
	now := time.Now()
	return now.After(p.StartDate) && now.Before(p.EndDate.Add(24*time.Hour))
}

// PeriodCreateRequest estructura para crear períodos
type PeriodCreateRequest struct {
	Code       string    `json:"code" binding:"required"`
	Name       string    `json:"name" binding:"required"`
	PeriodType string    `json:"period_type" binding:"required"`
	Year       int       `json:"year" binding:"required"`
	Month      *int      `json:"month"`
	Quarter    *int      `json:"quarter"`
	StartDate  time.Time `json:"start_date" binding:"required"`
	EndDate    time.Time `json:"end_date" binding:"required"`
}

// PeriodCloseRequest estructura para cerrar períodos
type PeriodCloseRequest struct {
	ClosingNotes string `json:"closing_notes"`
	ForceClose   bool   `json:"force_close"` // Permite cerrar aunque haya inconsistencias
}

// PeriodsListResponse respuesta para listado de períodos
type PeriodsListResponse struct {
	Periods    []Period        `json:"periods"`
	Pagination *PaginationInfo `json:"pagination"`
}

// TaxType representa un tipo de impuesto
type TaxType struct {
	BaseModel
	OrganizationID string  `json:"organization_id" gorm:"index;not null"`
	Code           string  `json:"code" gorm:"uniqueIndex:idx_org_tax_type;not null"`
	Name           string  `json:"name" gorm:"not null"`
	Description    string  `json:"description"`
	TaxCategory    string  `json:"tax_category" gorm:"not null"` // IVA, RETENCION, RETEICA, etc.
	Rate           float64 `json:"rate" gorm:"type:decimal(5,2);not null"`
	IsActive       bool    `json:"is_active" gorm:"default:true"`
	AccountID      *string `json:"account_id" gorm:"index"` // Cuenta contable asociada
}

// DocumentType representa un tipo de documento
type DocumentType struct {
	BaseModel
	OrganizationID string `json:"organization_id" gorm:"index;not null"`
	Code           string `json:"code" gorm:"uniqueIndex:idx_org_doc_type;not null"`
	Name           string `json:"name" gorm:"not null"`
	Description    string `json:"description"`
	Category       string `json:"category" gorm:"not null"` // IDENTITY, VOUCHER, INVOICE, etc.
	IsForPersons   bool   `json:"is_for_persons" gorm:"default:false"`
	IsForCompanies bool   `json:"is_for_companies" gorm:"default:false"`
	RequiresVerificationDigit bool `json:"requires_verification_digit" gorm:"default:false"`
	IsActive       bool   `json:"is_active" gorm:"default:true"`
}

// CostCenter representa un centro de costos
type CostCenter struct {
	BaseModel
	OrganizationID  string  `json:"organization_id" gorm:"index;not null"`
	Code            string  `json:"code" gorm:"uniqueIndex:idx_org_cost_center;not null"`
	Name            string  `json:"name" gorm:"not null"`
	Description     string  `json:"description"`
	ParentID        *string `json:"parent_id" gorm:"index"`
	Level           int     `json:"level" gorm:"not null"`
	IsActive        bool    `json:"is_active" gorm:"default:true"`
	BudgetAmount    float64 `json:"budget_amount" gorm:"type:decimal(15,2);default:0"`
	ActualAmount    float64 `json:"actual_amount" gorm:"type:decimal(15,2);default:0"`
	ResponsibleUserID *string `json:"responsible_user_id" gorm:"index"`
}

// AuditLog representa el log de auditoría
type AuditLog struct {
	BaseModel
	OrganizationID string    `json:"organization_id" gorm:"index;not null"`
	UserID         string    `json:"user_id" gorm:"index;not null"`
	Action         string    `json:"action" gorm:"not null;index"` // CREATE, UPDATE, DELETE, VIEW
	EntityType     string    `json:"entity_type" gorm:"not null;index"` // VOUCHER, ACCOUNT, etc.
	EntityID       string    `json:"entity_id" gorm:"not null;index"`
	Description    string    `json:"description" gorm:"not null"`
	IPAddress      string    `json:"ip_address"`
	UserAgent      string    `json:"user_agent"`
	OldValues      string    `json:"old_values" gorm:"type:text"`
	NewValues      string    `json:"new_values" gorm:"type:text"`
	Timestamp      time.Time `json:"timestamp" gorm:"not null;index"`
}

// UserPreferences movido a user_preferences.go para evitar duplicación