package data

import (
	"motor-contable-poc/internal/models"
	"gorm.io/gorm"
)

// AccountRepository maneja las operaciones de datos para cuentas contables
type AccountRepository struct {
	db *gorm.DB
}

// NewAccountRepository crea una nueva instancia del repositorio
func NewAccountRepository(db *gorm.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

// GetByOrganization obtiene cuentas de una organización con filtros
func (r *AccountRepository) GetByOrganization(orgID string, page, limit int, filters models.AccountFilters) ([]models.Account, int64, error) {
	var accounts []models.Account
	var total int64
	
	query := r.db.Where("organization_id = ?", orgID)
	
	// Aplicar filtros
	if filters.AccountType != "" {
		query = query.Where("account_type = ?", filters.AccountType)
	}
	if filters.IsActive != nil {
		query = query.Where("is_active = ?", *filters.IsActive)
	}
	if filters.Level != nil {
		query = query.Where("level = ?", *filters.Level)
	}
	if filters.ParentID != "" {
		query = query.Where("parent_account_id = ?", filters.ParentID)
	}
	
	// Contar total
	query.Model(&models.Account{}).Count(&total)
	
	// Obtener datos paginados
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("code").Find(&accounts).Error
	
	return accounts, total, err
}

// GetAllByOrganization obtiene todas las cuentas de una organización sin paginación
func (r *AccountRepository) GetAllByOrganization(orgID string) ([]models.Account, error) {
	var accounts []models.Account
	err := r.db.Where("organization_id = ?", orgID).Order("code").Find(&accounts).Error
	return accounts, err
}

// GetByID obtiene una cuenta por ID
func (r *AccountRepository) GetByID(id string) (*models.Account, error) {
	var account models.Account
	err := r.db.Where("id = ?", id).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// GetByCode obtiene una cuenta por código dentro de una organización
func (r *AccountRepository) GetByCode(orgID, code string) (*models.Account, error) {
	var account models.Account
	err := r.db.Where("organization_id = ? AND code = ?", orgID, code).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// GetHierarchy obtiene la jerarquía de cuentas (árbol)
// TODO: En el futuro, se usaría go-dsl para generar estructuras jerárquicas
// basadas en reglas contables y configuraciones del PUC
func (r *AccountRepository) GetHierarchy(orgID string) ([]*models.AccountTree, error) {
	var accounts []models.Account
	err := r.db.Where("organization_id = ? AND is_active = ?", orgID, true).
		Order("code").Find(&accounts).Error
	if err != nil {
		return nil, err
	}
	
	// Construir árbol jerárquico
	accountMap := make(map[string]*models.AccountTree)
	var roots []*models.AccountTree
	
	// Crear nodos
	for _, account := range accounts {
		node := &models.AccountTree{
			Account:      &account,
			SubAccounts:  []*models.AccountTree{},
			HasChildren:  false,
			TotalBalance: account.CurrentBalance,
		}
		accountMap[account.ID] = node
		
		if account.ParentAccountID == nil {
			roots = append(roots, node)
		}
	}
	
	// Establecer relaciones padre-hijo
	for _, account := range accounts {
		if account.ParentAccountID != nil {
			if parent, exists := accountMap[*account.ParentAccountID]; exists {
				if child, exists := accountMap[account.ID]; exists {
					parent.SubAccounts = append(parent.SubAccounts, child)
					parent.HasChildren = true
				}
			}
		}
	}
	
	return roots, nil
}

// GetByType obtiene cuentas por tipo
func (r *AccountRepository) GetByType(orgID, accountType string) ([]models.Account, error) {
	var accounts []models.Account
	err := r.db.Where("organization_id = ? AND account_type = ? AND is_active = ?", 
		orgID, accountType, true).Order("code").Find(&accounts).Error
	return accounts, err
}

// GetMovementAccounts obtiene solo las cuentas que aceptan movimiento
func (r *AccountRepository) GetMovementAccounts(orgID string) ([]models.Account, error) {
	var accounts []models.Account
	err := r.db.Where("organization_id = ? AND accepts_movement = ? AND is_active = ?", 
		orgID, true, true).Order("code").Find(&accounts).Error
	return accounts, err
}

// Search busca cuentas por código o nombre
func (r *AccountRepository) Search(orgID, query string, limit int) ([]models.Account, error) {
	var accounts []models.Account
	searchPattern := "%" + query + "%"
	
	err := r.db.Where("organization_id = ? AND is_active = ? AND (code LIKE ? OR name LIKE ?)",
		orgID, true, searchPattern, searchPattern).
		Limit(limit).Order("code").Find(&accounts).Error
	
	return accounts, err
}

// UpdateBalance actualiza el balance de una cuenta
func (r *AccountRepository) UpdateBalance(accountID string, debitAmount, creditAmount float64) error {
	return r.db.Model(&models.Account{}).Where("id = ?", accountID).Updates(map[string]interface{}{
		"balance_debit":  gorm.Expr("balance_debit + ?", debitAmount),
		"balance_credit": gorm.Expr("balance_credit + ?", creditAmount),
	}).Error
}

// Create crea una nueva cuenta
func (r *AccountRepository) Create(account *models.Account) error {
	return r.db.Create(account).Error
}

// Update actualiza una cuenta
func (r *AccountRepository) Update(account *models.Account) error {
	return r.db.Save(account).Error
}

// Delete elimina una cuenta (soft delete)
func (r *AccountRepository) Delete(id string) error {
	return r.db.Model(&models.Account{}).Where("id = ?", id).Update("is_active", false).Error
}

// HasMovements verifica si una cuenta tiene movimientos
func (r *AccountRepository) HasMovements(accountID string) (bool, error) {
	var count int64
	err := r.db.Table("voucher_lines").Where("account_id = ?", accountID).Count(&count).Error
	return count > 0, err
}

// HasSubAccounts verifica si una cuenta tiene subcuentas
func (r *AccountRepository) HasSubAccounts(accountID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Account{}).Where("parent_account_id = ?", accountID).Count(&count).Error
	return count > 0, err
}