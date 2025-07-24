package models

import "encoding/json"

// Account representa una cuenta contable en el Plan Único de Cuentas (PUC)
// Basado en el schema del Swagger y estándares contables colombianos
type Account struct {
	BaseModel
	OrganizationID    string  `json:"organization_id" gorm:"index;not null"`
	Code              string  `json:"code" gorm:"uniqueIndex:idx_org_account;not null"`
	Name              string  `json:"name" gorm:"not null"`
	Description       string  `json:"description"`
	AccountType       string  `json:"account_type" gorm:"not null"` // ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE
	Level             int     `json:"level" gorm:"not null"`        // 1,2,3,4,5 según PUC
	ParentAccountID   *string `json:"parent_account_id" gorm:"index"`
	PUCCode           string  `json:"puc_code"` // Código oficial PUC
	NaturalBalance    string  `json:"natural_balance" gorm:"not null"` // DEBIT, CREDIT
	AcceptsMovement   bool    `json:"accepts_movement" gorm:"default:true"`
	RequiresThirdParty bool   `json:"requires_third_party" gorm:"default:false"`
	RequiresCostCenter bool   `json:"requires_cost_center" gorm:"default:false"`
	IsActive          bool    `json:"is_active" gorm:"default:true"`
	BalanceDebit      float64 `json:"balance_debit" gorm:"type:decimal(15,2);default:0"`
	BalanceCredit     float64 `json:"balance_credit" gorm:"type:decimal(15,2);default:0"`
	CurrentBalance    float64 `json:"current_balance" gorm:"type:decimal(15,2);default:0"`
	MetadataJSON      string  `json:"-" gorm:"type:text;column:metadata"`
}

// AccountMetadata metadatos adicionales de la cuenta
type AccountMetadata struct {
	TaxCategory        string                 `json:"tax_category"`
	TaxRate           float64                `json:"tax_rate"`
	BankInfo          *BankInfo              `json:"bank_info,omitempty"`
	CostCenterRequired bool                   `json:"cost_center_required"`
	CustomFields      map[string]interface{} `json:"custom_fields,omitempty"`
}

// BankInfo información bancaria para cuentas de bancos
type BankInfo struct {
	BankName      string `json:"bank_name"`
	AccountNumber string `json:"account_number"`
	AccountType   string `json:"account_type"`
	Currency      string `json:"currency"`
}

// GetMetadata deserializa los metadatos de la cuenta
func (a *Account) GetMetadata() (*AccountMetadata, error) {
	if a.MetadataJSON == "" {
		return &AccountMetadata{}, nil
	}
	var metadata AccountMetadata
	err := json.Unmarshal([]byte(a.MetadataJSON), &metadata)
	return &metadata, err
}

// SetMetadata serializa los metadatos de la cuenta
func (a *Account) SetMetadata(metadata AccountMetadata) error {
	data, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	a.MetadataJSON = string(data)
	return nil
}

// AccountDetail estructura completa para respuestas detalladas de cuenta
type AccountDetail struct {
	*Account
	Metadata     *AccountMetadata `json:"metadata"`
	ParentAccount *Account        `json:"parent_account,omitempty"`
	SubAccounts  []Account       `json:"sub_accounts,omitempty"`
}

// ToDetail convierte una Account a AccountDetail con metadatos
func (a *Account) ToDetail() (*AccountDetail, error) {
	metadata, err := a.GetMetadata()
	if err != nil {
		return nil, err
	}
	
	return &AccountDetail{
		Account:  a,
		Metadata: metadata,
	}, nil
}

// IsBalanced verifica si la cuenta está balanceada según su naturaleza
func (a *Account) IsBalanced() bool {
	if a.NaturalBalance == "DEBIT" {
		return a.BalanceDebit >= a.BalanceCredit
	}
	return a.BalanceCredit >= a.BalanceDebit
}

// UpdateBalance actualiza el balance actual de la cuenta
func (a *Account) UpdateBalance() {
	if a.NaturalBalance == "DEBIT" {
		a.CurrentBalance = a.BalanceDebit - a.BalanceCredit
	} else {
		a.CurrentBalance = a.BalanceCredit - a.BalanceDebit
	}
}

// AccountsListResponse respuesta para listado de cuentas con paginación
type AccountsListResponse struct {
	Accounts   []Account       `json:"accounts"`
	Pagination *PaginationInfo `json:"pagination"`
}

// AccountTree estructura para árbol jerárquico de cuentas
type AccountTree struct {
	Account      *Account       `json:"account"`
	SubAccounts  []*AccountTree `json:"sub_accounts,omitempty"`
	HasChildren  bool           `json:"has_children"`
	TotalBalance float64        `json:"total_balance"`
}

// AccountTreeNode estructura para nodo del árbol compatible con frontend
type AccountTreeNode struct {
	ID           string             `json:"id"`
	AccountCode  string             `json:"account_code"`
	Name         string             `json:"name"`
	Type         string             `json:"type"`
	Nature       string             `json:"nature"`
	IsDetail     bool               `json:"is_detail"`
	IsActive     bool               `json:"is_active"`
	Balance      float64            `json:"balance"`
	Level        int                `json:"level"`
	ParentID     *string            `json:"parent_id"`
	Children     []AccountTreeNode  `json:"children,omitempty"`
}

// AccountTreeResponse respuesta para el árbol de cuentas
type AccountTreeResponse struct {
	Accounts []AccountTreeNode `json:"accounts"`
}

// AccountFilters filtros para la lista de cuentas
type AccountFilters struct {
	AccountType string `json:"account_type"`
	IsActive    *bool  `json:"is_active"`
	Level       *int   `json:"level"`
	ParentID    string `json:"parent_id"`
}

// AccountCreateRequest request para crear cuenta
type AccountCreateRequest struct {
	Code               string   `json:"code" validate:"required"`
	Name               string   `json:"name" validate:"required"`
	Description        string   `json:"description"`
	AccountType        string   `json:"account_type" validate:"required"`
	ParentAccountID    *string  `json:"parent_account_id"`
	NaturalBalance     string   `json:"natural_balance" validate:"required"`
	AcceptsMovement    bool     `json:"accepts_movement"`
	RequiresThirdParty bool     `json:"requires_third_party"`
	RequiresCostCenter bool     `json:"requires_cost_center"`
	IsActive           bool     `json:"is_active"`
}

// AccountUpdateRequest request para actualizar cuenta
type AccountUpdateRequest struct {
	Name               string  `json:"name"`
	Description        string  `json:"description"`
	AcceptsMovement    *bool   `json:"accepts_movement"`
	RequiresThirdParty *bool   `json:"requires_third_party"`
	RequiresCostCenter *bool   `json:"requires_cost_center"`
	IsActive           *bool   `json:"is_active"`
}

// AccountType tipo de cuenta disponible
type AccountType struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Nature      string `json:"nature"`
	Description string `json:"description"`
}

// ToTreeNode convierte Account a AccountTreeNode para el frontend
func (a *Account) ToTreeNode() AccountTreeNode {
	return AccountTreeNode{
		ID:          a.ID,
		AccountCode: a.Code,
		Name:        a.Name,
		Type:        a.AccountType,
		Nature:      a.NaturalBalance,
		IsDetail:    a.AcceptsMovement,
		IsActive:    a.IsActive,
		Balance:     a.CurrentBalance,
		Level:       a.Level,
		ParentID:    a.ParentAccountID,
		Children:    make([]AccountTreeNode, 0),
	}
}