package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseModel_BeforeCreate(t *testing.T) {
	base := &BaseModel{}
	
	// Test ID generation
	err := base.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, base.ID)
	assert.Len(t, base.ID, 36) // UUID length
}

func TestOrganization_JSONSerialization(t *testing.T) {
	org := &Organization{
		BaseModel: BaseModel{ID: "test-id"},
		Code:      "TEST001",
		Name:      "Test Organization",
	}

	// Test contact info serialization
	contactInfo := ContactInfo{
		Email:   "test@example.com",
		Phone:   "+57 1 234 5678",
		Address: "Test Address",
		City:    "Bogotá",
	}
	
	err := org.SetContactInfo(contactInfo)
	require.NoError(t, err)
	
	retrievedInfo, err := org.GetContactInfo()
	require.NoError(t, err)
	assert.Equal(t, contactInfo.Email, retrievedInfo.Email)
	assert.Equal(t, contactInfo.Phone, retrievedInfo.Phone)

	// Test fiscal info serialization
	fiscalInfo := FiscalInfo{
		FiscalYearStart:    "2024-01-01",
		FiscalYearEnd:      "2024-12-31",
		TaxRegime:          "RESPONSABLE_IVA",
		AccountingStandard: "NIIF_PYMES",
		ReportingCurrency:  "COP",
		DecimalPlaces:      2,
	}
	
	err = org.SetFiscalInfo(fiscalInfo)
	require.NoError(t, err)
	
	retrievedFiscal, err := org.GetFiscalInfo()
	require.NoError(t, err)
	assert.Equal(t, fiscalInfo.TaxRegime, retrievedFiscal.TaxRegime)
	assert.Equal(t, fiscalInfo.AccountingStandard, retrievedFiscal.AccountingStandard)
}

func TestOrganization_ToDetail(t *testing.T) {
	org := &Organization{
		BaseModel: BaseModel{ID: "test-id"},
		Code:      "TEST001",
		Name:      "Test Organization",
	}

	// Set some test data
	contactInfo := ContactInfo{Email: "test@example.com"}
	err := org.SetContactInfo(contactInfo)
	require.NoError(t, err)

	detail, err := org.ToDetail()
	require.NoError(t, err)
	assert.NotNil(t, detail)
	assert.Equal(t, org.Code, detail.Code)
	assert.NotNil(t, detail.ContactInfo)
	assert.Equal(t, contactInfo.Email, detail.ContactInfo.Email)
}

func TestAccount_UpdateBalance(t *testing.T) {
	account := &Account{
		Code:           "1105",
		Name:           "Caja",
		NaturalBalance: "DEBIT",
		BalanceDebit:   1000.0,
		BalanceCredit:  200.0,
	}

	// Test debit account
	account.UpdateBalance()
	assert.Equal(t, 800.0, account.CurrentBalance)

	// Test credit account
	account.NaturalBalance = "CREDIT"
	account.BalanceDebit = 200.0
	account.BalanceCredit = 1000.0
	account.UpdateBalance()
	assert.Equal(t, 800.0, account.CurrentBalance)
}

func TestAccount_IsBalanced(t *testing.T) {
	account := &Account{
		NaturalBalance: "DEBIT",
		BalanceDebit:   1000.0,
		BalanceCredit:  200.0,
	}

	// Debit account should be balanced when debit >= credit
	assert.True(t, account.IsBalanced())

	// Should not be balanced when credit > debit for debit account
	account.BalanceCredit = 1500.0
	assert.False(t, account.IsBalanced())

	// Credit account should be balanced when credit >= debit
	account.NaturalBalance = "CREDIT"
	assert.True(t, account.IsBalanced())
}

func TestVoucher_CalculateTotals(t *testing.T) {
	voucher := &Voucher{
		BaseModel:   BaseModel{ID: "test-voucher"},
		Description: "Test Voucher",
		VoucherLines: []VoucherLine{
			{
				Description:  "Line 1",
				DebitAmount:  1000.0,
				CreditAmount: 0.0,
				LineNumber:   1,
			},
			{
				Description:  "Line 2",
				DebitAmount:  0.0,
				CreditAmount: 500.0,
				LineNumber:   2,
			},
			{
				Description:  "Line 3",
				DebitAmount:  0.0,
				CreditAmount: 500.0,
				LineNumber:   3,
			},
		},
	}

	voucher.CalculateTotals()
	assert.Equal(t, 1000.0, voucher.TotalDebit)
	assert.Equal(t, 1000.0, voucher.TotalCredit)
	assert.True(t, voucher.IsBalanced)
}

func TestVoucher_CalculateTotalsUnbalanced(t *testing.T) {
	voucher := &Voucher{
		BaseModel:   BaseModel{ID: "test-voucher"},
		Description: "Unbalanced Voucher",
		VoucherLines: []VoucherLine{
			{
				Description:  "Line 1",
				DebitAmount:  1000.0,
				CreditAmount: 0.0,
				LineNumber:   1,
			},
			{
				Description:  "Line 2",
				DebitAmount:  0.0,
				CreditAmount: 300.0,
				LineNumber:   2,
			},
		},
	}

	voucher.CalculateTotals()
	assert.Equal(t, 1000.0, voucher.TotalDebit)
	assert.Equal(t, 300.0, voucher.TotalCredit)
	assert.False(t, voucher.IsBalanced)
}

func TestJournalEntry_CalculateTotals(t *testing.T) {
	entry := &JournalEntry{
		BaseModel:   BaseModel{ID: "test-entry"},
		Description: "Test Entry",
		JournalLines: []JournalLine{
			{
				Description:  "Line 1",
				DebitAmount:  2000.0,
				CreditAmount: 0.0,
				LineNumber:   1,
			},
			{
				Description:  "Line 2",
				DebitAmount:  0.0,
				CreditAmount: 2000.0,
				LineNumber:   2,
			},
		},
	}

	entry.CalculateTotals()
	assert.Equal(t, 2000.0, entry.TotalDebit)
	assert.Equal(t, 2000.0, entry.TotalCredit)
	assert.True(t, entry.IsBalanced())
}

func TestJournalEntry_IsBalanced(t *testing.T) {
	entry := &JournalEntry{
		TotalDebit:  1000.0,
		TotalCredit: 1000.0,
	}

	assert.True(t, entry.IsBalanced())

	entry.TotalCredit = 800.0
	assert.False(t, entry.IsBalanced())

	// Zero amounts should not be balanced
	entry.TotalDebit = 0.0
	entry.TotalCredit = 0.0
	assert.False(t, entry.IsBalanced())
}

func TestThirdParty_GetDisplayName(t *testing.T) {
	// Test natural person
	tp := &ThirdParty{
		PersonType: "NATURAL",
		FirstName:  "Juan",
		LastName:   "Pérez",
		Code:       "CLI001",
	}
	assert.Equal(t, "Juan Pérez", tp.GetDisplayName())

	// Test juridical person with company name
	tp = &ThirdParty{
		PersonType:  "JURIDICA",
		CompanyName: "Empresa S.A.S",
		Code:        "PROV001",
	}
	assert.Equal(t, "Empresa S.A.S", tp.GetDisplayName())

	// Test with commercial name
	tp = &ThirdParty{
		CommercialName: "Comercial XYZ",
		Code:           "CLI002",
	}
	assert.Equal(t, "Comercial XYZ", tp.GetDisplayName())

	// Test fallback to code
	tp = &ThirdParty{
		Code: "CLI003",
	}
	assert.Equal(t, "CLI003", tp.GetDisplayName())
}

func TestPeriod_CanClose(t *testing.T) {
	now := time.Now()
	
	// Period that ended yesterday should be closable
	period := &Period{
		Status:    "OPEN",
		StartDate: now.AddDate(0, -1, 0),
		EndDate:   now.AddDate(0, 0, -1),
	}
	assert.True(t, period.CanClose())

	// Period that ends tomorrow should not be closable
	period.EndDate = now.AddDate(0, 0, 1)
	assert.False(t, period.CanClose())

	// Closed period should not be closable
	period.Status = "CLOSED"
	period.EndDate = now.AddDate(0, 0, -1)
	assert.False(t, period.CanClose())
}

func TestPeriod_IsCurrentPeriod(t *testing.T) {
	now := time.Now()
	
	// Period that contains today
	period := &Period{
		StartDate: now.AddDate(0, 0, -15),
		EndDate:   now.AddDate(0, 0, 15),
	}
	assert.True(t, period.IsCurrentPeriod())

	// Period that ended yesterday
	period.StartDate = now.AddDate(0, -1, 0)
	period.EndDate = now.AddDate(0, 0, -1)
	assert.False(t, period.IsCurrentPeriod())

	// Period that starts tomorrow
	period.StartDate = now.AddDate(0, 0, 1)
	period.EndDate = now.AddDate(0, 0, 30)
	assert.False(t, period.IsCurrentPeriod())
}

func TestDSLTemplate_IsCompiled(t *testing.T) {
	template := &DSLTemplate{
		CompilationStatus: "SUCCESS",
		CompiledCode:      "compiled code here",
	}
	assert.True(t, template.IsCompiled())

	template.CompilationStatus = "ERROR"
	assert.False(t, template.IsCompiled())

	template.CompilationStatus = "SUCCESS"
	template.CompiledCode = ""
	assert.False(t, template.IsCompiled())
}

func TestDSLTemplate_CanExecute(t *testing.T) {
	template := &DSLTemplate{
		Status:            "ACTIVE",
		CompilationStatus: "SUCCESS",
		CompiledCode:      "compiled code here",
	}
	assert.True(t, template.CanExecute())

	template.Status = "INACTIVE"
	assert.False(t, template.CanExecute())

	template.Status = "ACTIVE"
	template.CompilationStatus = "ERROR"
	assert.False(t, template.CanExecute())
}

func TestResponseHelpers(t *testing.T) {
	// Test success response
	data := map[string]interface{}{"test": "value"}
	response := NewSuccessResponse(data)
	
	assert.True(t, response.Success)
	assert.Equal(t, data, response.Data)
	assert.Nil(t, response.Error)
	assert.NotZero(t, response.Timestamp)

	// Test error response
	errorResponse := NewErrorResponse("TEST_ERROR", "Test error message", "detail1", "detail2")
	
	assert.False(t, errorResponse.Success)
	assert.Nil(t, errorResponse.Data)
	assert.NotNil(t, errorResponse.Error)
	assert.Equal(t, "TEST_ERROR", errorResponse.Error.Code)
	assert.Equal(t, "Test error message", errorResponse.Error.Message)
	assert.Len(t, errorResponse.Error.Details, 2)
	assert.NotZero(t, errorResponse.Timestamp)
}