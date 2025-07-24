package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// generateID generates a new UUID
func generateID() string {
	return uuid.New().String()
}

// JournalTemplate represents a reusable journal entry template
type JournalTemplate struct {
	ID             string              `json:"id" gorm:"primaryKey"`
	Name           string              `json:"name" gorm:"not null"`
	Description    string              `json:"description"`
	DSLCode        string              `json:"dsl_code" gorm:"type:text;not null"`
	Parameters     TemplateParameters  `json:"parameters" gorm:"type:json"`
	IsActive       bool                `json:"is_active" gorm:"default:true"`
	OrganizationID string              `json:"organization_id"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
	CreatedBy      string              `json:"created_by"`
	
	// Relations
	Organization Organization `json:"-" gorm:"foreignKey:OrganizationID"`
}

// TemplateParameter defines a parameter for the template
type TemplateParameter struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"` // number, string, date
	Required     bool        `json:"required"`
	DefaultValue interface{} `json:"default_value,omitempty"`
	Description  string      `json:"description"`
}

// TemplateParameters is a slice of parameters
type TemplateParameters []TemplateParameter

// Value implements the driver.Valuer interface
func (tp TemplateParameters) Value() (driver.Value, error) {
	return json.Marshal(tp)
}

// Scan implements the sql.Scanner interface
func (tp *TemplateParameters) Scan(value interface{}) error {
	if value == nil {
		*tp = nil
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	
	return json.Unmarshal(bytes, tp)
}

// TemplateExecution represents an execution of a template
type TemplateExecution struct {
	ID               string                 `json:"id" gorm:"primaryKey"`
	TemplateID       string                 `json:"template_id" gorm:"not null"`
	JournalEntryID   string                 `json:"journal_entry_id"`
	Parameters       map[string]interface{} `json:"parameters" gorm:"type:json"`
	ExecutedAt       time.Time              `json:"executed_at"`
	ExecutedBy       string                 `json:"executed_by"`
	Status           string                 `json:"status"` // SUCCESS, FAILED
	ErrorMessage     string                 `json:"error_message,omitempty"`
	
	// Relations
	Template     JournalTemplate `json:"template,omitempty" gorm:"foreignKey:TemplateID"`
	JournalEntry *JournalEntry   `json:"journal_entry,omitempty" gorm:"foreignKey:JournalEntryID"`
}

// BeforeCreate sets the ID and timestamps
func (t *JournalTemplate) BeforeCreate(tx *gorm.DB) error {
	t.ID = generateID()
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate sets the updated timestamp
func (t *JournalTemplate) BeforeUpdate(tx *gorm.DB) error {
	t.UpdatedAt = time.Now()
	return nil
}

// BeforeCreate sets the ID for template execution
func (te *TemplateExecution) BeforeCreate(tx *gorm.DB) error {
	te.ID = generateID()
	te.ExecutedAt = time.Now()
	return nil
}

// TemplateListResponse for API responses
type TemplateListResponse struct {
	Templates []JournalTemplate `json:"templates"`
	Total     int               `json:"total"`
}

// TemplatePreviewRequest for preview API
type TemplatePreviewRequest struct {
	Parameters map[string]interface{} `json:"parameters"`
}

// TemplateExecuteRequest for execute API
type TemplateExecuteRequest struct {
	Parameters map[string]interface{} `json:"parameters"`
	DryRun     bool                   `json:"dry_run"`
}

// TemplatePreviewResponse for preview results
type TemplatePreviewResponse struct {
	Preview *JournalEntry `json:"preview"`
	Summary struct {
		TotalDebit  float64 `json:"total_debit"`
		TotalCredit float64 `json:"total_credit"`
		IsBalanced  bool    `json:"is_balanced"`
		LinesCount  int     `json:"lines_count"`
	} `json:"summary"`
}

// TemplateExecuteResponse for execution results
type TemplateExecuteResponse struct {
	Success     bool   `json:"success"`
	EntryID     string `json:"entry_id,omitempty"`
	EntryNumber string `json:"entry_number,omitempty"`
	Error       string `json:"error,omitempty"`
}