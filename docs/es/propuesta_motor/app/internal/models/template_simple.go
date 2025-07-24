package models

import (
	"time"
)

// Template represents a simple template for SQLite
type Template struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	DSLContent  string    `json:"dsl_content,omitempty" gorm:"column:dsl_content"`
	Parameters  string    `json:"parameters"` // JSON string
	Status      string    `json:"status"`
	IsActive    bool      `json:"is_active" gorm:"column:is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedBy   string    `json:"created_by"`
	CompanyID   string    `json:"company_id"`
}

// TableName specifies the table name
func (Template) TableName() string {
	return "templates"
}