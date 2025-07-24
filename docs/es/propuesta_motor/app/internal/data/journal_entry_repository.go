package data

import (
	"fmt"
	"motor-contable-poc/internal/models"
	"time"

	"gorm.io/gorm"
)

// JournalEntryRepository maneja las operaciones de datos para asientos contables
type JournalEntryRepository struct {
	db *gorm.DB
}

// NewJournalEntryRepository crea una nueva instancia del repositorio
func NewJournalEntryRepository(db *gorm.DB) *JournalEntryRepository {
	return &JournalEntryRepository{db: db}
}

// GetByOrganization obtiene todos los asientos de una organización
func (r *JournalEntryRepository) GetByOrganization(orgID string, page, limit int) ([]models.JournalEntry, int64, error) {
	var entries []models.JournalEntry
	var total int64

	query := r.db.Where("organization_id = ?", orgID).
		Preload("JournalLines").
		Preload("JournalLines.Account")

	// Contar total
	query.Model(&models.JournalEntry{}).Count(&total)

	// Obtener datos paginados
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("date DESC, entry_number DESC").Find(&entries).Error

	return entries, total, err
}

// GetByID obtiene un asiento por ID con todas sus líneas
func (r *JournalEntryRepository) GetByID(id string) (*models.JournalEntry, error) {
	var entry models.JournalEntry
	err := r.db.Where("id = ?", id).
		Preload("JournalLines").
		Preload("JournalLines.Account").
		Preload("JournalLines.ThirdParty").
		First(&entry).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

// GetByNumber obtiene un asiento por número dentro de una organización
func (r *JournalEntryRepository) GetByNumber(orgID, number string) (*models.JournalEntry, error) {
	var entry models.JournalEntry
	err := r.db.Where("organization_id = ? AND entry_number = ?", orgID, number).
		Preload("JournalLines").
		Preload("JournalLines.Account").
		First(&entry).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

// GetByVoucherID obtiene el asiento asociado a un comprobante
func (r *JournalEntryRepository) GetByVoucherID(voucherID string) (*models.JournalEntry, error) {
	var entry models.JournalEntry
	err := r.db.Where("voucher_id = ?", voucherID).
		Preload("JournalLines").
		Preload("JournalLines.Account").
		First(&entry).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

// GetByDateRange obtiene asientos por rango de fechas
func (r *JournalEntryRepository) GetByDateRange(orgID string, startDate, endDate time.Time, page, limit int) ([]models.JournalEntry, int64, error) {
	var entries []models.JournalEntry
	var total int64

	query := r.db.Where("organization_id = ? AND date BETWEEN ? AND ?", orgID, startDate, endDate).
		Preload("JournalLines").
		Preload("JournalLines.Account")

	// Contar total
	query.Model(&models.JournalEntry{}).Count(&total)

	// Obtener datos paginados
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("date DESC, entry_number DESC").Find(&entries).Error

	return entries, total, err
}

// GetByStatus obtiene asientos por estado
func (r *JournalEntryRepository) GetByStatus(orgID, status string, page, limit int) ([]models.JournalEntry, int64, error) {
	var entries []models.JournalEntry
	var total int64

	query := r.db.Where("organization_id = ? AND status = ?", orgID, status).
		Preload("JournalLines").
		Preload("JournalLines.Account")

	// Contar total
	query.Model(&models.JournalEntry{}).Count(&total)

	// Obtener datos paginados
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("date DESC, entry_number DESC").Find(&entries).Error

	return entries, total, err
}

// GetByAccount obtiene asientos que involucran una cuenta específica
func (r *JournalEntryRepository) GetByAccount(orgID, accountID string, page, limit int) ([]models.JournalEntry, int64, error) {
	var entries []models.JournalEntry
	var total int64

	// Subconsulta para obtener asientos que tienen líneas con la cuenta especificada
	subQuery := r.db.Table("journal_lines").
		Select("journal_entry_id").
		Where("account_id = ?", accountID)

	query := r.db.Where("organization_id = ? AND id IN (?)", orgID, subQuery).
		Preload("JournalLines").
		Preload("JournalLines.Account")

	// Contar total
	query.Model(&models.JournalEntry{}).Count(&total)

	// Obtener datos paginados
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("date DESC, entry_number DESC").Find(&entries).Error

	return entries, total, err
}

// Create crea un nuevo asiento con sus líneas
func (r *JournalEntryRepository) Create(entry *models.JournalEntry) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Crear el asiento
		if err := tx.Create(entry).Error; err != nil {
			return err
		}

		// Actualizar balances de las cuentas
		for _, line := range entry.JournalLines {
			if line.DebitAmount > 0 {
				if err := tx.Model(&models.Account{}).Where("id = ?", line.AccountID).
					Update("balance_debit", gorm.Expr("balance_debit + ?", line.DebitAmount)).Error; err != nil {
					return err
				}
			}
			if line.CreditAmount > 0 {
				if err := tx.Model(&models.Account{}).Where("id = ?", line.AccountID).
					Update("balance_credit", gorm.Expr("balance_credit + ?", line.CreditAmount)).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// Update actualiza un asiento
func (r *JournalEntryRepository) Update(entry *models.JournalEntry) error {
	return r.db.Save(entry).Error
}

// Post cambia el estado de un asiento a POSTED
func (r *JournalEntryRepository) Post(entryID, userID string) error {
	now := time.Now()
	return r.db.Model(&models.JournalEntry{}).Where("id = ?", entryID).Updates(map[string]interface{}{
		"status":    "POSTED",
		"posted_at": &now,
	}).Error
}

// MarkAsReversed marca un asiento como reversado
func (r *JournalEntryRepository) MarkAsReversed(entryID, reversalEntryID string) error {
	return r.db.Model(&models.JournalEntry{}).Where("id = ?", entryID).Updates(map[string]interface{}{
		"is_reversed":           true,
		"reversed_by_entry_id":  reversalEntryID,
		"reversal_reason":      "Asiento reversado automáticamente",
		"status":               "REVERSED",
	}).Error
}

// Delete elimina un asiento (solo si no está contabilizado)
func (r *JournalEntryRepository) Delete(id string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Verificar que no esté contabilizado
		var entry models.JournalEntry
		if err := tx.Where("id = ?", id).First(&entry).Error; err != nil {
			return err
		}

		if entry.Status == "POSTED" {
			return fmt.Errorf("no se puede eliminar un asiento contabilizado")
		}

		// Eliminar líneas primero
		if err := tx.Where("journal_entry_id = ?", id).Delete(&models.JournalLine{}).Error; err != nil {
			return err
		}

		// Eliminar asiento
		return tx.Delete(&models.JournalEntry{}, "id = ?", id).Error
	})
}

// GetNextNumber obtiene el siguiente número disponible para un asiento
func (r *JournalEntryRepository) GetNextNumber(orgID string) (string, error) {
	var count int64
	err := r.db.Model(&models.JournalEntry{}).
		Where("organization_id = ?", orgID).
		Count(&count).Error

	if err != nil {
		return "", err
	}

	// Generar número secuencial basado en el conteo
	nextNumber := count + 1
	return fmt.Sprintf("AS-%04d", nextNumber), nil
}

// GetStatistics obtiene estadísticas de asientos contables
func (r *JournalEntryRepository) GetStatistics(orgID string, periodID string) (map[string]interface{}, error) {
	var stats map[string]interface{} = make(map[string]interface{})

	// Total de asientos
	var totalEntries int64
	r.db.Model(&models.JournalEntry{}).Where("organization_id = ? AND period_id = ?", orgID, periodID).Count(&totalEntries)
	stats["total_entries"] = totalEntries

	// Asientos por estado
	var statusCounts []struct {
		Status string
		Count  int64
	}
	r.db.Model(&models.JournalEntry{}).
		Where("organization_id = ? AND period_id = ?", orgID, periodID).
		Group("status").
		Select("status, count(*) as count").
		Scan(&statusCounts)
	stats["by_status"] = statusCounts

	// Totales de débito y crédito
	var totals struct {
		TotalDebit  float64
		TotalCredit float64
	}
	r.db.Model(&models.JournalEntry{}).
		Where("organization_id = ? AND period_id = ?", orgID, periodID).
		Select("SUM(total_debit) as total_debit, SUM(total_credit) as total_credit").
		Scan(&totals)
	stats["total_debit"] = totals.TotalDebit
	stats["total_credit"] = totals.TotalCredit

	return stats, nil
}

// GetTrialBalance genera el balance de comprobación
func (r *JournalEntryRepository) GetTrialBalance(orgID string, periodID string) (*models.TrialBalanceResponse, error) {
	// Esta es una implementación simplificada para el POC
	// En una implementación real, esto sería mucho más complejo

	var entries []models.TrialBalanceEntry

	// Obtener movimientos por cuenta
	rows, err := r.db.Raw(`
		SELECT 
			a.id as account_id,
			a.code as account_code,
			a.name as account_name,
			a.account_type,
			a.level,
			COALESCE(SUM(jl.debit_amount), 0) as movement_debit,
			COALESCE(SUM(jl.credit_amount), 0) as movement_credit
		FROM accounts a
		LEFT JOIN journal_lines jl ON a.id = jl.account_id
		LEFT JOIN journal_entries je ON jl.journal_entry_id = je.id
		WHERE a.organization_id = ? AND (je.period_id = ? OR je.period_id IS NULL)
		GROUP BY a.id, a.code, a.name, a.account_type, a.level
		ORDER BY a.code
	`, orgID, periodID).Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var totalDebit, totalCredit float64

	for rows.Next() {
		var entry models.TrialBalanceEntry
		err := rows.Scan(
			&entry.AccountID,
			&entry.AccountCode,
			&entry.AccountName,
			&entry.AccountType,
			&entry.Level,
			&entry.MovementDebit,
			&entry.MovementCredit,
		)
		if err != nil {
			return nil, err
		}

		// Para el POC, no tenemos saldos iniciales
		entry.InitialDebit = 0
		entry.InitialCredit = 0
		entry.FinalDebit = entry.InitialDebit + entry.MovementDebit
		entry.FinalCredit = entry.InitialCredit + entry.MovementCredit

		entries = append(entries, entry)
		totalDebit += entry.FinalDebit
		totalCredit += entry.FinalCredit
	}

	response := &models.TrialBalanceResponse{
		PeriodID:    periodID,
		PeriodName:  "Período Actual", // Placeholder
		GeneratedAt: time.Now(),
		Entries:     entries,
		TotalDebit:  totalDebit,
		TotalCredit: totalCredit,
		IsBalanced:  totalDebit == totalCredit,
	}

	return response, nil
}