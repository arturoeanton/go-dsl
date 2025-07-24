package data

import (
	"fmt"
	"motor-contable-poc/internal/models"
	"time"
	"gorm.io/gorm"
)

// VoucherRepository maneja las operaciones de datos para comprobantes
type VoucherRepository struct {
	db *gorm.DB
}

// NewVoucherRepository crea una nueva instancia del repositorio
func NewVoucherRepository(db *gorm.DB) *VoucherRepository {
	return &VoucherRepository{db: db}
}

// GetByOrganization obtiene todos los comprobantes de una organización
func (r *VoucherRepository) GetByOrganization(orgID string, page, limit int) ([]models.Voucher, int64, error) {
	var vouchers []models.Voucher
	var total int64
	
	query := r.db.Where("organization_id = ?", orgID).
		Preload("VoucherLines").
		Preload("VoucherLines.Account").
		Preload("ThirdParty")
	
	// Contar total
	query.Model(&models.Voucher{}).Count(&total)
	
	// Obtener datos paginados
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("date DESC, number DESC").Find(&vouchers).Error
	
	return vouchers, total, err
}

// GetByID obtiene un comprobante por ID con todas sus líneas
func (r *VoucherRepository) GetByID(id string) (*models.Voucher, error) {
	var voucher models.Voucher
	err := r.db.Where("id = ?", id).
		Preload("VoucherLines").
		Preload("VoucherLines.Account").
		Preload("VoucherLines.ThirdParty").
		Preload("ThirdParty").
		First(&voucher).Error
	if err != nil {
		return nil, err
	}
	return &voucher, nil
}

// GetByNumber obtiene un comprobante por número dentro de una organización
func (r *VoucherRepository) GetByNumber(orgID, number string) (*models.Voucher, error) {
	var voucher models.Voucher
	err := r.db.Where("organization_id = ? AND number = ?", orgID, number).
		Preload("VoucherLines").
		Preload("VoucherLines.Account").
		Preload("ThirdParty").
		First(&voucher).Error
	if err != nil {
		return nil, err
	}
	return &voucher, nil
}

// GetByDateRange obtiene comprobantes por rango de fechas
func (r *VoucherRepository) GetByDateRange(orgID string, startDate, endDate time.Time, page, limit int) ([]models.Voucher, int64, error) {
	var vouchers []models.Voucher
	var total int64
	
	query := r.db.Where("organization_id = ? AND date BETWEEN ? AND ?", orgID, startDate, endDate).
		Preload("VoucherLines")
	
	// Contar total
	query.Model(&models.Voucher{}).Count(&total)
	
	// Obtener datos paginados
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("date DESC, number DESC").Find(&vouchers).Error
	
	return vouchers, total, err
}

// GetByStatus obtiene comprobantes por estado
func (r *VoucherRepository) GetByStatus(orgID, status string, page, limit int) ([]models.Voucher, int64, error) {
	var vouchers []models.Voucher
	var total int64
	
	query := r.db.Where("organization_id = ? AND status = ?", orgID, status).
		Preload("VoucherLines")
	
	// Contar total
	query.Model(&models.Voucher{}).Count(&total)
	
	// Obtener datos paginados
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("date DESC, number DESC").Find(&vouchers).Error
	
	return vouchers, total, err
}

// GetByType obtiene comprobantes por tipo
func (r *VoucherRepository) GetByType(orgID, voucherType string, page, limit int) ([]models.Voucher, int64, error) {
	var vouchers []models.Voucher
	var total int64
	
	query := r.db.Where("organization_id = ? AND voucher_type = ?", orgID, voucherType).
		Preload("VoucherLines")
	
	// Contar total
	query.Model(&models.Voucher{}).Count(&total)
	
	// Obtener datos paginados
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("date DESC, number DESC").Find(&vouchers).Error
	
	return vouchers, total, err
}

// Create crea un nuevo comprobante con sus líneas
// TODO: En el futuro, aquí se usaría go-dsl para validar reglas contables,
// generar automáticamente líneas adicionales (impuestos, retenciones),
// y aplicar plantillas de automatización
func (r *VoucherRepository) Create(voucher *models.Voucher) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Crear el comprobante
		if err := tx.Create(voucher).Error; err != nil {
			return err
		}
		
		// Actualizar balances de las cuentas
		for _, line := range voucher.VoucherLines {
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

// Update actualiza un comprobante
func (r *VoucherRepository) Update(voucher *models.Voucher) error {
	return r.db.Save(voucher).Error
}

// Post cambia el estado de un comprobante a POSTED y genera el asiento contable
// TODO: Aquí se usaría go-dsl para generar automáticamente el asiento contable
// basado en las reglas DSL configuradas para cada tipo de comprobante
func (r *VoucherRepository) Post(voucherID, userID string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Actualizar estado del comprobante
		now := time.Now()
		err := tx.Model(&models.Voucher{}).Where("id = ?", voucherID).Updates(map[string]interface{}{
			"status":           "POSTED",
			"posted_by_user_id": userID,
			"posted_at":        &now,
		}).Error
		if err != nil {
			return err
		}
		
		// TODO: Generar asiento contable automáticamente usando go-dsl
		// Aquí se aplicarían las reglas DSL para:
		// 1. Validar que el comprobante cumple las reglas contables
		// 2. Generar las líneas del asiento contable
		// 3. Aplicar clasificaciones automáticas
		// 4. Calcular impuestos y retenciones
		
		return nil
	})
}

// Cancel cancela un comprobante
func (r *VoucherRepository) Cancel(voucherID string) error {
	return r.db.Model(&models.Voucher{}).Where("id = ?", voucherID).
		Update("status", "CANCELLED").Error
}

// CountByType cuenta los comprobantes por tipo
func (r *VoucherRepository) CountByType(orgID string) (map[string]int, error) {
	var results []struct {
		VoucherType string
		Count       int
	}
	
	err := r.db.Model(&models.Voucher{}).
		Where("organization_id = ?", orgID).
		Group("voucher_type").
		Select("voucher_type, COUNT(*) as count").
		Scan(&results).Error
	
	if err != nil {
		return nil, err
	}
	
	// Convertir a map
	countByType := make(map[string]int)
	for _, result := range results {
		countByType[result.VoucherType] = result.Count
	}
	
	return countByType, nil
}

// GetNextNumber obtiene el siguiente número disponible para un tipo de comprobante
// TODO: En el futuro, se usaría go-dsl para generar números según reglas
// configurables de numeración automática
func (r *VoucherRepository) GetNextNumber(orgID, voucherType string) (string, error) {
	var count int64
	err := r.db.Model(&models.Voucher{}).
		Where("organization_id = ? AND voucher_type = ?", orgID, voucherType).
		Count(&count).Error
	
	if err != nil {
		return "", err
	}
	
	// Generar número secuencial basado en el conteo
	nextNumber := count + 1
	return fmt.Sprintf("%s-%03d", voucherType, nextNumber), nil
}

// GetStatistics obtiene estadísticas de comprobantes
func (r *VoucherRepository) GetStatistics(orgID string, periodID string) (map[string]interface{}, error) {
	var stats map[string]interface{} = make(map[string]interface{})
	
	// Total de comprobantes
	var totalVouchers int64
	r.db.Model(&models.Voucher{}).Where("organization_id = ? AND period_id = ?", orgID, periodID).Count(&totalVouchers)
	stats["total_vouchers"] = totalVouchers
	
	// Comprobantes por estado
	var statusCounts []struct {
		Status string
		Count  int64
	}
	r.db.Model(&models.Voucher{}).
		Where("organization_id = ? AND period_id = ?", orgID, periodID).
		Group("status").
		Select("status, count(*) as count").
		Scan(&statusCounts)
	stats["by_status"] = statusCounts
	
	// Comprobantes por tipo
	var typeCounts []struct {
		VoucherType string
		Count       int64
	}
	r.db.Model(&models.Voucher{}).
		Where("organization_id = ? AND period_id = ?", orgID, periodID).
		Group("voucher_type").
		Select("voucher_type, count(*) as count").
		Scan(&typeCounts)
	stats["by_type"] = typeCounts
	
	return stats, nil
}

// CountByDateRange cuenta comprobantes en un rango de fechas
func (r *VoucherRepository) CountByDateRange(orgID string, startDate, endDate time.Time) (int, error) {
	var count int64
	err := r.db.Model(&models.Voucher{}).
		Where("organization_id = ? AND date BETWEEN ? AND ?", orgID, startDate, endDate).
		Count(&count).Error
	return int(count), err
}

// CountByStatus cuenta comprobantes por estado
func (r *VoucherRepository) CountByStatus(orgID string, status string) (int, error) {
	var count int64
	err := r.db.Model(&models.Voucher{}).
		Where("organization_id = ? AND status = ?", orgID, status).
		Count(&count).Error
	return int(count), err
}

// Count cuenta el total de comprobantes
func (r *VoucherRepository) Count(orgID string) (int, error) {
	var count int64
	err := r.db.Model(&models.Voucher{}).
		Where("organization_id = ?", orgID).
		Count(&count).Error
	return int(count), err
}