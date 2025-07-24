package models

import "encoding/json"

// ThirdParty representa un tercero (cliente, proveedor, empleado, etc.)
// Basado en el schema del Swagger y requerimientos contables colombianos
type ThirdParty struct {
	BaseModel
	OrganizationID     string  `json:"organization_id" gorm:"index;not null"`
	Code               string  `json:"code" gorm:"uniqueIndex:idx_org_third_party;not null"`
	DocumentType       string  `json:"document_type" gorm:"not null"` // CC, NIT, CE, PP, etc.
	DocumentNumber     string  `json:"document_number" gorm:"not null;index"`
	VerificationDigit  *string `json:"verification_digit"`
	FirstName          string  `json:"first_name"`
	LastName           string  `json:"last_name"`
	CompanyName        string  `json:"company_name"`
	CommercialName     string  `json:"commercial_name"`
	PersonType         string  `json:"person_type" gorm:"not null"` // NATURAL, JURIDICA
	ThirdPartyType     string  `json:"third_party_type" gorm:"not null"` // CUSTOMER, SUPPLIER, EMPLOYEE, OTHER
	TaxpayerType       string  `json:"taxpayer_type"` // RESPONSABLE_IVA, NO_RESPONSABLE_IVA, REGIMEN_SIMPLE
	IsActive           bool    `json:"is_active" gorm:"default:true"`
	ContactInfoJSON    string  `json:"-" gorm:"type:text;column:contact_info"`
	TaxInfoJSON        string  `json:"-" gorm:"type:text;column:tax_info"`
	PaymentInfoJSON    string  `json:"-" gorm:"type:text;column:payment_info"`
	AdditionalInfoJSON string  `json:"-" gorm:"type:text;column:additional_info"`
}

// ThirdPartyContactInfo información de contacto del tercero
type ThirdPartyContactInfo struct {
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	Mobile        string `json:"mobile"`
	Address       string `json:"address"`
	City          string `json:"city"`
	State         string `json:"state"`
	Country       string `json:"country"`
	PostalCode    string `json:"postal_code"`
	Website       string `json:"website"`
	ContactPerson string `json:"contact_person"`
}

// ThirdPartyTaxInfo información fiscal del tercero
type ThirdPartyTaxInfo struct {
	TaxRegime              string   `json:"tax_regime"`
	ResponsibleForIVA      bool     `json:"responsible_for_iva"`
	WithholdingAgent       bool     `json:"withholding_agent"`
	SelfRetaining          bool     `json:"self_retaining"`
	TaxResponsibilities    []string `json:"tax_responsibilities"`
	EconomicActivity       string   `json:"economic_activity"`
	CIIUCode               string   `json:"ciiu_code"`
	TributaryObligations   []string `json:"tributary_obligations"`
	WithholdingPercentage  float64  `json:"withholding_percentage"`
}

// ThirdPartyPaymentInfo información de pagos del tercero
type ThirdPartyPaymentInfo struct {
	PreferredPaymentMethod string                 `json:"preferred_payment_method"` // CASH, CHECK, TRANSFER, CARD
	PaymentTerms           int                    `json:"payment_terms"` // Días de crédito
	CreditLimit            float64                `json:"credit_limit"`
	CurrentBalance         float64                `json:"current_balance"`
	BankAccounts           []ThirdPartyBankAccount `json:"bank_accounts"`
	DiscountPercentage     float64                `json:"discount_percentage"`
}

// ThirdPartyBankAccount cuenta bancaria del tercero
type ThirdPartyBankAccount struct {
	BankName      string `json:"bank_name"`
	AccountNumber string `json:"account_number"`
	AccountType   string `json:"account_type"`
	IsDefault     bool   `json:"is_default"`
}

// ThirdPartyAdditionalInfo información adicional del tercero
type ThirdPartyAdditionalInfo struct {
	Notes                string                 `json:"notes"`
	InternalCode         string                 `json:"internal_code"`
	ExternalCode         string                 `json:"external_code"`
	SalesRepresentative  string                 `json:"sales_representative"`
	CustomerSegment      string                 `json:"customer_segment"`
	SupplierCategory     string                 `json:"supplier_category"`
	CustomFields         map[string]interface{} `json:"custom_fields"`
}

// GetContactInfo deserializa la información de contacto
func (tp *ThirdParty) GetContactInfo() (*ThirdPartyContactInfo, error) {
	if tp.ContactInfoJSON == "" {
		return &ThirdPartyContactInfo{}, nil
	}
	var info ThirdPartyContactInfo
	err := json.Unmarshal([]byte(tp.ContactInfoJSON), &info)
	return &info, err
}

// SetContactInfo serializa la información de contacto
func (tp *ThirdParty) SetContactInfo(info ThirdPartyContactInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	tp.ContactInfoJSON = string(data)
	return nil
}

// GetTaxInfo deserializa la información fiscal
func (tp *ThirdParty) GetTaxInfo() (*ThirdPartyTaxInfo, error) {
	if tp.TaxInfoJSON == "" {
		return &ThirdPartyTaxInfo{}, nil
	}
	var info ThirdPartyTaxInfo
	err := json.Unmarshal([]byte(tp.TaxInfoJSON), &info)
	return &info, err
}

// SetTaxInfo serializa la información fiscal
func (tp *ThirdParty) SetTaxInfo(info ThirdPartyTaxInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	tp.TaxInfoJSON = string(data)
	return nil
}

// GetPaymentInfo deserializa la información de pagos
func (tp *ThirdParty) GetPaymentInfo() (*ThirdPartyPaymentInfo, error) {
	if tp.PaymentInfoJSON == "" {
		return &ThirdPartyPaymentInfo{}, nil
	}
	var info ThirdPartyPaymentInfo
	err := json.Unmarshal([]byte(tp.PaymentInfoJSON), &info)
	return &info, err
}

// SetPaymentInfo serializa la información de pagos
func (tp *ThirdParty) SetPaymentInfo(info ThirdPartyPaymentInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	tp.PaymentInfoJSON = string(data)
	return nil
}

// GetAdditionalInfo deserializa la información adicional
func (tp *ThirdParty) GetAdditionalInfo() (*ThirdPartyAdditionalInfo, error) {
	if tp.AdditionalInfoJSON == "" {
		return &ThirdPartyAdditionalInfo{}, nil
	}
	var info ThirdPartyAdditionalInfo
	err := json.Unmarshal([]byte(tp.AdditionalInfoJSON), &info)
	return &info, err
}

// SetAdditionalInfo serializa la información adicional
func (tp *ThirdParty) SetAdditionalInfo(info ThirdPartyAdditionalInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	tp.AdditionalInfoJSON = string(data)
	return nil
}

// GetDisplayName obtiene el nombre a mostrar del tercero
func (tp *ThirdParty) GetDisplayName() string {
	if tp.PersonType == "NATURAL" {
		if tp.FirstName != "" && tp.LastName != "" {
			return tp.FirstName + " " + tp.LastName
		}
	}
	if tp.CompanyName != "" {
		return tp.CompanyName
	}
	if tp.CommercialName != "" {
		return tp.CommercialName
	}
	return tp.Code
}

// ThirdPartyDetail estructura completa para respuestas detalladas
type ThirdPartyDetail struct {
	*ThirdParty
	ContactInfo    *ThirdPartyContactInfo    `json:"contact_info"`
	TaxInfo        *ThirdPartyTaxInfo        `json:"tax_info"`
	PaymentInfo    *ThirdPartyPaymentInfo    `json:"payment_info"`
	AdditionalInfo *ThirdPartyAdditionalInfo `json:"additional_info"`
	DisplayName    string                    `json:"display_name"`
}

// ToDetail convierte un ThirdParty a ThirdPartyDetail con todas las configuraciones
func (tp *ThirdParty) ToDetail() (*ThirdPartyDetail, error) {
	contactInfo, err := tp.GetContactInfo()
	if err != nil {
		return nil, err
	}
	
	taxInfo, err := tp.GetTaxInfo()
	if err != nil {
		return nil, err
	}
	
	paymentInfo, err := tp.GetPaymentInfo()
	if err != nil {
		return nil, err
	}
	
	additionalInfo, err := tp.GetAdditionalInfo()
	if err != nil {
		return nil, err
	}
	
	return &ThirdPartyDetail{
		ThirdParty:     tp,
		ContactInfo:    contactInfo,
		TaxInfo:        taxInfo,
		PaymentInfo:    paymentInfo,
		AdditionalInfo: additionalInfo,
		DisplayName:    tp.GetDisplayName(),
	}, nil
}

// ThirdPartyCreateRequest estructura para crear terceros
type ThirdPartyCreateRequest struct {
	Code               string                    `json:"code" binding:"required"`
	DocumentType       string                    `json:"document_type" binding:"required"`
	DocumentNumber     string                    `json:"document_number" binding:"required"`
	VerificationDigit  *string                   `json:"verification_digit"`
	FirstName          string                    `json:"first_name"`
	LastName           string                    `json:"last_name"`
	CompanyName        string                    `json:"company_name"`
	CommercialName     string                    `json:"commercial_name"`
	PersonType         string                    `json:"person_type" binding:"required"`
	ThirdPartyType     string                    `json:"third_party_type" binding:"required"`
	TaxpayerType       string                    `json:"taxpayer_type"`
	ContactInfo        *ThirdPartyContactInfo    `json:"contact_info"`
	TaxInfo            *ThirdPartyTaxInfo        `json:"tax_info"`
	PaymentInfo        *ThirdPartyPaymentInfo    `json:"payment_info"`
	AdditionalInfo     *ThirdPartyAdditionalInfo `json:"additional_info"`
}

// ThirdPartiesListResponse respuesta para listado de terceros
type ThirdPartiesListResponse struct {
	ThirdParties []ThirdParty    `json:"third_parties"`
	Pagination   *PaginationInfo `json:"pagination"`
}

// ThirdPartySearchRequest estructura para búsqueda de terceros
type ThirdPartySearchRequest struct {
	Query            string `json:"query"`
	DocumentType     string `json:"document_type"`
	DocumentNumber   string `json:"document_number"`
	ThirdPartyType   string `json:"third_party_type"`
	PersonType       string `json:"person_type"`
	IsActive         *bool  `json:"is_active"`
	Page             int    `json:"page"`
	Limit            int    `json:"limit"`
}