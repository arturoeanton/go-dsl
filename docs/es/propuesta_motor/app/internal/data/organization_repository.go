package data

import (
	"motor-contable-poc/internal/models"
	"gorm.io/gorm"
)

// OrganizationRepository maneja las operaciones de datos para organizaciones
type OrganizationRepository struct {
	db *gorm.DB
}

// NewOrganizationRepository crea una nueva instancia del repositorio
func NewOrganizationRepository(db *gorm.DB) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

// GetCurrent obtiene la organización actual (para el POC usamos la primera)
// TODO: En el futuro, aquí se usaría go-dsl para determinar la organización
// basada en reglas de tenant/contexto de usuario
func (r *OrganizationRepository) GetCurrent() (*models.Organization, error) {
	var org models.Organization
	err := r.db.Where("is_active = ?", true).First(&org).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// GetByID obtiene una organización por ID
func (r *OrganizationRepository) GetByID(id string) (*models.Organization, error) {
	var org models.Organization
	err := r.db.Where("id = ?", id).First(&org).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// GetByCode obtiene una organización por código
func (r *OrganizationRepository) GetByCode(code string) (*models.Organization, error) {
	var org models.Organization
	err := r.db.Where("code = ?", code).First(&org).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// Update actualiza una organización
func (r *OrganizationRepository) Update(org *models.Organization) error {
	return r.db.Save(org).Error
}

// Create crea una nueva organización
func (r *OrganizationRepository) Create(org *models.Organization) error {
	return r.db.Create(org).Error
}