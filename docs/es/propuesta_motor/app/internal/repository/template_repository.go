package repository

import (
	"motor-contable-poc/internal/models"
	"gorm.io/gorm"
)

type TemplateRepository interface {
	Create(template *models.JournalTemplate) error
	Update(template *models.JournalTemplate) error
	Delete(id string) error
	GetByID(id string) (*models.JournalTemplate, error)
	GetAll(organizationID string) ([]models.JournalTemplate, error)
	GetActive(organizationID string) ([]models.JournalTemplate, error)
	
	// Execution history
	CreateExecution(execution *models.TemplateExecution) error
	GetExecutionHistory(templateID string, limit int) ([]models.TemplateExecution, error)
}

type templateRepository struct {
	db *gorm.DB
}

func NewTemplateRepository(db *gorm.DB) TemplateRepository {
	return &templateRepository{db: db}
}

func (r *templateRepository) Create(template *models.JournalTemplate) error {
	return r.db.Create(template).Error
}

func (r *templateRepository) Update(template *models.JournalTemplate) error {
	return r.db.Save(template).Error
}

func (r *templateRepository) Delete(id string) error {
	return r.db.Delete(&models.JournalTemplate{}, "id = ?", id).Error
}

func (r *templateRepository) GetByID(id string) (*models.JournalTemplate, error) {
	var template models.JournalTemplate
	err := r.db.First(&template, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *templateRepository) GetAll(organizationID string) ([]models.JournalTemplate, error) {
	var templates []models.JournalTemplate
	err := r.db.Where("organization_id = ?", organizationID).
		Order("name").
		Find(&templates).Error
	return templates, err
}

func (r *templateRepository) GetActive(organizationID string) ([]models.JournalTemplate, error) {
	var templates []models.JournalTemplate
	err := r.db.Where("organization_id = ? AND is_active = ?", organizationID, true).
		Order("name").
		Find(&templates).Error
	return templates, err
}

func (r *templateRepository) CreateExecution(execution *models.TemplateExecution) error {
	return r.db.Create(execution).Error
}

func (r *templateRepository) GetExecutionHistory(templateID string, limit int) ([]models.TemplateExecution, error) {
	var executions []models.TemplateExecution
	query := r.db.Where("template_id = ?", templateID).
		Preload("JournalEntry").
		Order("executed_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Find(&executions).Error
	return executions, err
}