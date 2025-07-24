package services

import (
	"fmt"
	"motor-contable-poc/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Auto-migrate all models
	err = db.AutoMigrate(
		&models.Organization{},
		&models.Account{},
		&models.ThirdParty{},
		&models.Period{},
		&models.Voucher{},
		&models.VoucherLine{},
		&models.JournalEntry{},
		&models.JournalLine{},
		&models.DSLTemplate{},
	)
	require.NoError(t, err)

	return db
}

// createTestOrganization creates a test organization
func createTestOrganization(t *testing.T, db *gorm.DB) *models.Organization {
	org := &models.Organization{
		Code:            "TEST001",
		Name:            "Test Organization",
		TaxID:           "900123456-7",
		CountryCode:     "CO",
		CurrencyDefault: "COP",
		IsActive:        true,
	}

	err := db.Create(org).Error
	require.NoError(t, err)
	return org
}

// createTestAccounts creates basic test accounts
func createTestAccounts(t *testing.T, db *gorm.DB, orgID string) []*models.Account {
	accounts := []*models.Account{
		{
			OrganizationID:  orgID,
			Code:            "1105",
			Name:            "Caja",
			AccountType:     "ASSET",
			Level:           4,
			NaturalBalance:  "DEBIT",
			AcceptsMovement: true,
			IsActive:        true,
		},
		{
			OrganizationID:  orgID,
			Code:            "2408",
			Name:            "IVA por Pagar",
			AccountType:     "LIABILITY",
			Level:           4,
			NaturalBalance:  "CREDIT",
			AcceptsMovement: true,
			IsActive:        true,
		},
	}

	for _, account := range accounts {
		err := db.Create(account).Error
		require.NoError(t, err)
	}

	return accounts
}

func TestOrganizationService_GetCurrent(t *testing.T) {
	db := setupTestDB(t)
	service := NewOrganizationService(db)

	// Test when no organization exists
	_, err := service.GetCurrent()
	assert.Error(t, err)

	// Create test organization
	org := createTestOrganization(t, db)

	// Test successful retrieval
	detail, err := service.GetCurrent()
	require.NoError(t, err)
	assert.NotNil(t, detail)
	assert.Equal(t, org.Code, detail.Code)
	assert.Equal(t, org.Name, detail.Name)
}

func TestOrganizationService_UpdateConfiguration(t *testing.T) {
	db := setupTestDB(t)
	service := NewOrganizationService(db)
	org := createTestOrganization(t, db)

	// Test updating contact info
	contactInfo := models.ContactInfo{
		Email: "updated@test.com",
		Phone: "+57 1 999 8888",
		City:  "Medellín",
	}

	updates := map[string]interface{}{
		"contact_info": contactInfo,
	}

	err := service.UpdateConfiguration(org.ID, updates)
	assert.NoError(t, err)

	// Verify update
	detail, err := service.GetByID(org.ID)
	require.NoError(t, err)
	assert.Equal(t, contactInfo.Email, detail.ContactInfo.Email)
	assert.Equal(t, contactInfo.City, detail.ContactInfo.City)
}

func TestOrganizationService_ValidateBusinessRules(t *testing.T) {
	db := setupTestDB(t)
	service := NewOrganizationService(db)

	// Create organization without required fields
	org := &models.Organization{
		Code:     "INVALID001",
		Name:     "Invalid Org",
		IsActive: true,
	}
	err := db.Create(org).Error
	require.NoError(t, err)

	violations, err := service.ValidateBusinessRules(org.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, violations)
	assert.Contains(t, violations[0], "NIT requerido")
}

func TestVoucherService_Create(t *testing.T) {
	db := setupTestDB(t)
	service := NewVoucherService(db)
	org := createTestOrganization(t, db)
	accounts := createTestAccounts(t, db, org.ID)

	// Create valid voucher request
	request := models.VoucherCreateRequest{
		VoucherType: "JOURNAL",
		Date:        time.Now(),
		Description: "Test voucher",
		VoucherLines: []models.VoucherLineRequest{
			{
				AccountID:    accounts[0].ID, // Caja (debit)
				Description:  "Ingreso de efectivo",
				DebitAmount:  1000.0,
				CreditAmount: 0.0,
			},
			{
				AccountID:    accounts[1].ID, // IVA (credit)
				Description:  "IVA por pagar",
				DebitAmount:  0.0,
				CreditAmount: 1000.0,
			},
		},
	}

	voucher, err := service.Create(org.ID, request)
	require.NoError(t, err)
	assert.NotNil(t, voucher)
	assert.Equal(t, request.Description, voucher.Description)
	assert.Equal(t, request.VoucherType, voucher.VoucherType)
	assert.Equal(t, "DRAFT", voucher.Status)
	assert.True(t, voucher.IsBalanced)
	assert.Equal(t, 1000.0, voucher.TotalDebit)
	assert.Equal(t, 1000.0, voucher.TotalCredit)
	assert.Len(t, voucher.VoucherLines, 2)
}

func TestVoucherService_CreateUnbalanced(t *testing.T) {
	db := setupTestDB(t)
	service := NewVoucherService(db)
	org := createTestOrganization(t, db)
	accounts := createTestAccounts(t, db, org.ID)

	// Create unbalanced voucher request
	request := models.VoucherCreateRequest{
		VoucherType: "JOURNAL",
		Date:        time.Now(),
		Description: "Unbalanced voucher",
		VoucherLines: []models.VoucherLineRequest{
			{
				AccountID:    accounts[0].ID,
				Description:  "Debit line",
				DebitAmount:  1000.0,
				CreditAmount: 0.0,
			},
			{
				AccountID:    accounts[1].ID,
				Description:  "Credit line",
				DebitAmount:  0.0,
				CreditAmount: 500.0, // Unbalanced!
			},
		},
	}

	voucher, err := service.Create(org.ID, request)
	assert.Error(t, err)
	assert.Nil(t, voucher)
	assert.Contains(t, err.Error(), "no está balanceado")
}

func TestVoucherService_CreateInsufficientLines(t *testing.T) {
	db := setupTestDB(t)
	service := NewVoucherService(db)
	org := createTestOrganization(t, db)
	accounts := createTestAccounts(t, db, org.ID)

	// Create request with only one line
	request := models.VoucherCreateRequest{
		VoucherType: "JOURNAL",
		Date:        time.Now(),
		Description: "Single line voucher",
		VoucherLines: []models.VoucherLineRequest{
			{
				AccountID:    accounts[0].ID,
				Description:  "Only line",
				DebitAmount:  1000.0,
				CreditAmount: 0.0,
			},
		},
	}

	voucher, err := service.Create(org.ID, request)
	assert.Error(t, err)
	assert.Nil(t, voucher)
	assert.Contains(t, err.Error(), "al menos 2 líneas")
}

func TestVoucherService_CreateInvalidAccount(t *testing.T) {
	db := setupTestDB(t)
	service := NewVoucherService(db)
	org := createTestOrganization(t, db)

	// Create request with non-existent account
	request := models.VoucherCreateRequest{
		VoucherType: "JOURNAL",
		Date:        time.Now(),
		Description: "Invalid account voucher",
		VoucherLines: []models.VoucherLineRequest{
			{
				AccountID:    "non-existent-id",
				Description:  "Invalid account",
				DebitAmount:  1000.0,
				CreditAmount: 0.0,
			},
			{
				AccountID:    "another-non-existent-id",
				Description:  "Another invalid account",
				DebitAmount:  0.0,
				CreditAmount: 1000.0,
			},
		},
	}

	voucher, err := service.Create(org.ID, request)
	assert.Error(t, err)
	assert.Nil(t, voucher)
	assert.Contains(t, err.Error(), "no encontrada")
}

func TestVoucherService_Post(t *testing.T) {
	db := setupTestDB(t)
	service := NewVoucherService(db)
	org := createTestOrganization(t, db)
	accounts := createTestAccounts(t, db, org.ID)

	// Create a voucher first
	request := models.VoucherCreateRequest{
		VoucherType: "JOURNAL",
		Date:        time.Now(),
		Description: "Test voucher for posting",
		VoucherLines: []models.VoucherLineRequest{
			{
				AccountID:    accounts[0].ID,
				Description:  "Debit line",
				DebitAmount:  1000.0,
				CreditAmount: 0.0,
			},
			{
				AccountID:    accounts[1].ID,
				Description:  "Credit line",
				DebitAmount:  0.0,
				CreditAmount: 1000.0,
			},
		},
	}

	voucher, err := service.Create(org.ID, request)
	require.NoError(t, err)
	assert.Equal(t, "DRAFT", voucher.Status)

	// Test posting the voucher
	err = service.Post(voucher.ID, "test-user-id")
	assert.NoError(t, err)

	// Verify the voucher was posted
	postedVoucher, err := service.GetByID(voucher.ID)
	require.NoError(t, err)
	assert.Equal(t, "POSTED", postedVoucher.Status)
}

func TestVoucherService_PostInvalidVoucher(t *testing.T) {
	db := setupTestDB(t)
	service := NewVoucherService(db)

	// Test posting non-existent voucher
	err := service.Post("non-existent-id", "test-user-id")
	assert.Error(t, err)
}

func TestVoucherService_Cancel(t *testing.T) {
	db := setupTestDB(t)
	service := NewVoucherService(db)
	org := createTestOrganization(t, db)
	accounts := createTestAccounts(t, db, org.ID)

	// Create a voucher
	request := models.VoucherCreateRequest{
		VoucherType: "JOURNAL",
		Date:        time.Now(),
		Description: "Test voucher for cancellation",
		VoucherLines: []models.VoucherLineRequest{
			{
				AccountID:    accounts[0].ID,
				Description:  "Debit line",
				DebitAmount:  1000.0,
				CreditAmount: 0.0,
			},
			{
				AccountID:    accounts[1].ID,
				Description:  "Credit line",
				DebitAmount:  0.0,
				CreditAmount: 1000.0,
			},
		},
	}

	voucher, err := service.Create(org.ID, request)
	require.NoError(t, err)

	// Test canceling the voucher
	err = service.Cancel(voucher.ID)
	assert.NoError(t, err)

	// Verify voucher was cancelled
	cancelledVoucher, err := service.GetByID(voucher.ID)
	require.NoError(t, err)
	assert.Equal(t, "CANCELLED", cancelledVoucher.Status)
}

func TestVoucherService_GetList(t *testing.T) {
	db := setupTestDB(t)
	service := NewVoucherService(db)
	org := createTestOrganization(t, db)
	accounts := createTestAccounts(t, db, org.ID)

	// Create multiple vouchers
	for i := 0; i < 5; i++ {
		request := models.VoucherCreateRequest{
			VoucherType: "JOURNAL",
			Date:        time.Now().Add(time.Duration(i) * time.Minute), // Different times to avoid conflicts
			Description: fmt.Sprintf("Test voucher %d", i),
			VoucherLines: []models.VoucherLineRequest{
				{
					AccountID:    accounts[0].ID,
					Description:  "Debit line",
					DebitAmount:  1000.0,
					CreditAmount: 0.0,
				},
				{
					AccountID:    accounts[1].ID,
					Description:  "Credit line",
					DebitAmount:  0.0,
					CreditAmount: 1000.0,
				},
			},
		}

		_, err := service.Create(org.ID, request)
		require.NoError(t, err)
	}

	// Test getting list
	response, err := service.GetList(org.ID, 1, 10)
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Vouchers, 5)
	assert.NotNil(t, response.Pagination)
	assert.Equal(t, 1, response.Pagination.Page)
	assert.Equal(t, 10, response.Pagination.Limit)
	assert.Equal(t, 5, response.Pagination.Total)
}

func TestVoucherService_GetByDateRange(t *testing.T) {
	db := setupTestDB(t)
	service := NewVoucherService(db)
	org := createTestOrganization(t, db)
	accounts := createTestAccounts(t, db, org.ID)

	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	tomorrow := now.AddDate(0, 0, 1)

	// Create voucher for yesterday
	request := models.VoucherCreateRequest{
		VoucherType: "JOURNAL",
		Date:        yesterday,
		Description: "Yesterday voucher",
		VoucherLines: []models.VoucherLineRequest{
			{
				AccountID:    accounts[0].ID,
				Description:  "Debit line",
				DebitAmount:  1000.0,
				CreditAmount: 0.0,
			},
			{
				AccountID:    accounts[1].ID,
				Description:  "Credit line",
				DebitAmount:  0.0,
				CreditAmount: 1000.0,
			},
		},
	}

	_, err := service.Create(org.ID, request)
	require.NoError(t, err)

	// Create voucher for today
	request.Date = now
	request.Description = "Today voucher"
	_, err = service.Create(org.ID, request)
	require.NoError(t, err)

	// Test getting vouchers by date range (only yesterday)
	response, err := service.GetByDateRange(org.ID, yesterday, yesterday, 1, 10)
	require.NoError(t, err)
	assert.Len(t, response.Vouchers, 1)
	assert.Contains(t, response.Vouchers[0].Description, "Yesterday")

	// Test getting vouchers by date range (yesterday to tomorrow)
	response, err = service.GetByDateRange(org.ID, yesterday, tomorrow, 1, 10)
	require.NoError(t, err)
	assert.Len(t, response.Vouchers, 2)
}

// Benchmark tests for performance validation
func BenchmarkVoucherService_Create(b *testing.B) {
	db := setupTestDB(&testing.T{})
	service := NewVoucherService(db)
	org := createTestOrganization(&testing.T{}, db)
	accounts := createTestAccounts(&testing.T{}, db, org.ID)

	request := models.VoucherCreateRequest{
		VoucherType: "JOURNAL",
		Date:        time.Now(),
		Description: "Benchmark voucher",
		VoucherLines: []models.VoucherLineRequest{
			{
				AccountID:    accounts[0].ID,
				Description:  "Debit line",
				DebitAmount:  1000.0,
				CreditAmount: 0.0,
			},
			{
				AccountID:    accounts[1].ID,
				Description:  "Credit line",
				DebitAmount:  0.0,
				CreditAmount: 1000.0,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request.Description = "Benchmark voucher " + string(rune(i))
		_, err := service.Create(org.ID, request)
		if err != nil {
			b.Fatalf("Error creating voucher: %v", err)
		}
	}
}

func BenchmarkVoucherService_GetList(b *testing.B) {
	db := setupTestDB(&testing.T{})
	service := NewVoucherService(db)
	org := createTestOrganization(&testing.T{}, db)
	accounts := createTestAccounts(&testing.T{}, db, org.ID)

	// Create some test data
	request := models.VoucherCreateRequest{
		VoucherType: "JOURNAL",
		Date:        time.Now(),
		Description: "Benchmark voucher",
		VoucherLines: []models.VoucherLineRequest{
			{
				AccountID:    accounts[0].ID,
				Description:  "Debit line",
				DebitAmount:  1000.0,
				CreditAmount: 0.0,
			},
			{
				AccountID:    accounts[1].ID,
				Description:  "Credit line",
				DebitAmount:  0.0,
				CreditAmount: 1000.0,
			},
		},
	}

	for i := 0; i < 100; i++ {
		request.Description = "Benchmark voucher " + string(rune(i))
		_, err := service.Create(org.ID, request)
		if err != nil {
			b.Fatalf("Error creating test voucher: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetList(org.ID, 1, 20)
		if err != nil {
			b.Fatalf("Error getting voucher list: %v", err)
		}
	}
}