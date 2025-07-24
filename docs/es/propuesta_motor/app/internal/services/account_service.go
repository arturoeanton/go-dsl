package services

import (
	"errors"
	"fmt"
	"motor-contable-poc/internal/data"
	"motor-contable-poc/internal/models"
	"strings"

	"gorm.io/gorm"
)

// AccountService maneja la lógica de negocio para cuentas contables
type AccountService struct {
	accountRepo *data.AccountRepository
}

// NewAccountService crea una nueva instancia del servicio
func NewAccountService(db *gorm.DB) *AccountService {
	return &AccountService{
		accountRepo: data.NewAccountRepository(db),
	}
}

// GetTree obtiene el árbol jerárquico de cuentas
func (s *AccountService) GetTree(orgID string) ([]models.AccountTreeNode, error) {
	// Obtener todas las cuentas de la organización
	accounts, err := s.accountRepo.GetAllByOrganization(orgID)
	if err != nil {
		return nil, err
	}

	// Construir el árbol jerárquico
	tree := s.buildAccountTree(accounts)
	return tree, nil
}

// GetList obtiene una lista paginada de cuentas
func (s *AccountService) GetList(orgID string, page, limit int, filters models.AccountFilters) (*models.AccountsListResponse, error) {
	accounts, total, err := s.accountRepo.GetByOrganization(orgID, page, limit, filters)
	if err != nil {
		return nil, err
	}

	pages := int((total + int64(limit) - 1) / int64(limit))

	return &models.AccountsListResponse{
		Accounts: accounts,
		Pagination: &models.PaginationInfo{
			Page:  page,
			Limit: limit,
			Total: int(total),
			Pages: pages,
		},
	}, nil
}

// GetByCode obtiene una cuenta por su código
func (s *AccountService) GetByCode(orgID, code string) (*models.Account, error) {
	return s.accountRepo.GetByCode(orgID, code)
}

// GetByID obtiene una cuenta por ID
func (s *AccountService) GetByID(id string) (*models.Account, error) {
	return s.accountRepo.GetByID(id)
}

// Create crea una nueva cuenta contable
func (s *AccountService) Create(orgID string, request models.AccountCreateRequest) (*models.Account, error) {
	// Validar que el código no exista
	existingAccount, err := s.accountRepo.GetByCode(orgID, request.Code)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if existingAccount != nil {
		return nil, errors.New("ya existe una cuenta con ese código")
	}

	// Validar cuenta padre si se especifica
	var parentAccount *models.Account
	var level int = 1
	if request.ParentAccountID != nil {
		parentAccount, err = s.accountRepo.GetByID(*request.ParentAccountID)
		if err != nil {
			return nil, fmt.Errorf("cuenta padre no encontrada: %v", err)
		}
		if parentAccount.OrganizationID != orgID {
			return nil, errors.New("cuenta padre no pertenece a la organización")
		}
		level = parentAccount.Level + 1
	}

	// Validar código jerárquico
	if err := s.validateAccountCode(request.Code, parentAccount); err != nil {
		return nil, err
	}

	// Crear la cuenta
	account := &models.Account{
		OrganizationID:     orgID,
		Code:               request.Code,
		Name:               request.Name,
		Description:        request.Description,
		AccountType:        request.AccountType,
		Level:              level,
		ParentAccountID:    request.ParentAccountID,
		NaturalBalance:     request.NaturalBalance,
		AcceptsMovement:    request.AcceptsMovement,
		RequiresThirdParty: request.RequiresThirdParty,
		RequiresCostCenter: request.RequiresCostCenter,
		IsActive:           request.IsActive,
	}

	// Establecer defaults
	if account.AcceptsMovement && level <= 2 {
		account.AcceptsMovement = false // Cuentas de nivel alto no aceptan movimiento
	}

	err = s.accountRepo.Create(account)
	if err != nil {
		return nil, err
	}

	return account, nil
}

// Update actualiza una cuenta contable existente
func (s *AccountService) Update(id string, request models.AccountUpdateRequest) (*models.Account, error) {
	account, err := s.accountRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Aplicar cambios
	if request.Name != "" {
		account.Name = request.Name
	}
	if request.Description != "" {
		account.Description = request.Description
	}
	if request.AcceptsMovement != nil {
		account.AcceptsMovement = *request.AcceptsMovement
	}
	if request.RequiresThirdParty != nil {
		account.RequiresThirdParty = *request.RequiresThirdParty
	}
	if request.RequiresCostCenter != nil {
		account.RequiresCostCenter = *request.RequiresCostCenter
	}
	if request.IsActive != nil {
		account.IsActive = *request.IsActive
	}

	err = s.accountRepo.Update(account)
	if err != nil {
		return nil, err
	}

	return account, nil
}

// Delete elimina una cuenta (soft delete)
func (s *AccountService) Delete(id string) error {
	// Verificar que no tenga movimientos
	hasMovements, err := s.accountRepo.HasMovements(id)
	if err != nil {
		return err
	}
	if hasMovements {
		return errors.New("no se puede eliminar una cuenta con movimientos")
	}

	// Verificar que no tenga subcuentas
	hasSubAccounts, err := s.accountRepo.HasSubAccounts(id)
	if err != nil {
		return err
	}
	if hasSubAccounts {
		return errors.New("no se puede eliminar una cuenta con subcuentas")
	}

	return s.accountRepo.Delete(id)
}

// buildAccountTree construye el árbol jerárquico de cuentas
func (s *AccountService) buildAccountTree(accounts []models.Account) []models.AccountTreeNode {
	// Crear mapa para acceso rápido
	accountMap := make(map[string]*models.AccountTreeNode)
	var rootNodes []models.AccountTreeNode

	// Crear todos los nodos
	for _, account := range accounts {
		node := account.ToTreeNode()
		accountMap[account.ID] = &node
	}

	// Construir la jerarquía
	for _, account := range accounts {
		node := accountMap[account.ID]
		if account.ParentAccountID != nil {
			// Es una subcuenta
			if parent, exists := accountMap[*account.ParentAccountID]; exists {
				parent.Children = append(parent.Children, *node)
			}
		} else {
			// Es una cuenta raíz
			rootNodes = append(rootNodes, *node)
		}
	}

	return rootNodes
}

// validateAccountCode valida que el código de cuenta sea consistente con la jerarquía
func (s *AccountService) validateAccountCode(code string, parentAccount *models.Account) error {
	if parentAccount == nil {
		// Es una cuenta raíz, debe ser de 1 dígito
		if len(code) != 1 {
			return errors.New("cuentas raíz deben tener 1 dígito")
		}
		return nil
	}

	// Debe empezar con el código del padre
	if !strings.HasPrefix(code, parentAccount.Code) {
		return fmt.Errorf("código debe empezar con el código del padre: %s", parentAccount.Code)
	}

	// Validar longitud según nivel
	expectedLength := len(parentAccount.Code) + 2 // Generalmente se incrementa de 2 en 2
	if len(code) != expectedLength {
		return fmt.Errorf("código debe tener %d dígitos para nivel %d", expectedLength, parentAccount.Level+1)
	}

	return nil
}