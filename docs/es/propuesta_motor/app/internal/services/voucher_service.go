package services

import (
	"errors"
	"fmt"
	"motor-contable-poc/internal/data"
	"motor-contable-poc/internal/models"
	"time"
	"gorm.io/gorm"
)

// VoucherService maneja la lógica de negocio para comprobantes
type VoucherService struct {
	voucherRepo         *data.VoucherRepository
	accountRepo         *data.AccountRepository
	journalEntryService *JournalEntryService
	db                  *gorm.DB
}

// NewVoucherService crea una nueva instancia del servicio
func NewVoucherService(db *gorm.DB) *VoucherService {
	return &VoucherService{
		voucherRepo:         data.NewVoucherRepository(db),
		accountRepo:         data.NewAccountRepository(db),
		journalEntryService: NewJournalEntryService(db),
		db:                  db,
	}
}

// GetList obtiene una lista paginada de comprobantes
func (s *VoucherService) GetList(orgID string, page, limit int) (*models.VouchersListResponse, error) {
	vouchers, total, err := s.voucherRepo.GetByOrganization(orgID, page, limit)
	if err != nil {
		return nil, err
	}
	
	pages := int((total + int64(limit) - 1) / int64(limit))
	
	// Calcular estadísticas
	stats := &models.VoucherStats{
		TotalVouchers: int(total),
		TotalAmount:   0,
		PendingCount:  0,
		ErrorCount:    0,
	}
	
	// Calcular totales y contadores
	for _, voucher := range vouchers {
		stats.TotalAmount += voucher.TotalDebit
		if voucher.Status == "DRAFT" || voucher.Status == "PENDING" {
			stats.PendingCount++
		} else if voucher.Status == "ERROR" {
			stats.ErrorCount++
		}
	}
	
	return &models.VouchersListResponse{
		Vouchers: vouchers,
		Pagination: &models.PaginationInfo{
			Page:  page,
			Limit: limit,
			Total: int(total),
			Pages: pages,
		},
		Stats: stats,
	}, nil
}

// GetByID obtiene un comprobante por ID con validaciones
func (s *VoucherService) GetByID(id string) (*models.VoucherDetail, error) {
	voucher, err := s.voucherRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	
	// TODO: Aquí se aplicarían reglas DSL para validar permisos de acceso
	// según el rol del usuario y configuraciones de seguridad
	
	detail, err := voucher.ToDetail()
	if err != nil {
		return nil, err
	}
	
	return detail, nil
}

// Create crea un nuevo comprobante con validaciones de negocio
// TODO: En el futuro, se usaría go-dsl para aplicar reglas de validación
// automática, generar líneas adicionales y aplicar plantillas contables
func (s *VoucherService) Create(orgID string, request models.VoucherCreateRequest) (*models.Voucher, error) {
	// Validar que las líneas están balanceadas
	err := s.validateVoucherLines(request.VoucherLines)
	if err != nil {
		return nil, err
	}
	
	// Validar que las cuentas existen y aceptan movimiento
	err = s.validateAccounts(orgID, request.VoucherLines)
	if err != nil {
		return nil, err
	}
	
	// Generar número automático
	number, err := s.voucherRepo.GetNextNumber(orgID, request.VoucherType)
	if err != nil {
		return nil, err
	}
	
	// TODO: Aquí se ejecutarían reglas DSL para:
	// 1. Aplicar plantillas de comprobantes según el tipo
	// 2. Generar automáticamente líneas de impuestos y retenciones
	// 3. Validar reglas contables específicas (ej: cuentas de terceros)
	// 4. Aplicar clasificaciones automáticas según configuraciones
	
	// Crear el comprobante
	voucher := &models.Voucher{
		OrganizationID: orgID,
		Number:         number,
		VoucherType:    request.VoucherType,
		Date:           request.Date,
		Description:    request.Description,
		Reference:      request.Reference,
		PeriodID:       s.getCurrentPeriodID(orgID), // TODO: Implementar
		Status:         "DRAFT",
		ThirdPartyID:   request.ThirdPartyID,
		CostCenterID:   request.CostCenterID,
		CreatedByUserID: "system", // TODO: Obtener del contexto
	}
	
	// Convertir líneas del request
	for i, lineReq := range request.VoucherLines {
		line := models.VoucherLine{
			AccountID:    lineReq.AccountID,
			Description:  lineReq.Description,
			DebitAmount:  lineReq.DebitAmount,
			CreditAmount: lineReq.CreditAmount,
			ThirdPartyID: lineReq.ThirdPartyID,
			CostCenterID: lineReq.CostCenterID,
			TaxAmount:    lineReq.TaxAmount,
			TaxRate:      lineReq.TaxRate,
			BaseAmount:   lineReq.BaseAmount,
			LineNumber:   i + 1,
		}
		voucher.VoucherLines = append(voucher.VoucherLines, line)
	}
	
	// Calcular totales
	voucher.CalculateTotals()
	
	// Establecer datos adicionales si se proporcionan
	if request.AdditionalData != nil {
		voucher.SetAdditionalData(*request.AdditionalData)
	}
	
	// Crear en la base de datos
	err = s.voucherRepo.Create(voucher)
	if err != nil {
		return nil, err
	}
	
	return voucher, nil
}

// Post procesa y contabiliza un comprobante
// Crea automáticamente el asiento contable correspondiente
func (s *VoucherService) Post(voucherID, userID string) error {
	// Iniciar transacción
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	// Obtener el comprobante con sus líneas
	voucher, err := s.voucherRepo.GetByID(voucherID)
	if err != nil {
		tx.Rollback()
		return err
	}
	
	// Validar que el comprobante puede ser contabilizado
	if voucher.Status != "DRAFT" {
		tx.Rollback()
		return errors.New("solo se pueden contabilizar comprobantes en estado BORRADOR")
	}
	
	if !voucher.IsBalanced {
		tx.Rollback()
		return errors.New("el comprobante debe estar balanceado para ser contabilizado")
	}
	
	// TODO: En el futuro, aquí se ejecutarían reglas DSL para:
	// 1. Validar restricciones contables específicas del tipo de comprobante
	// 2. Aplicar clasificaciones y distribuciones automáticas
	// 3. Calcular y generar impuestos y retenciones automáticamente
	// 4. Ejecutar workflows de aprobación si son requeridos
	
	// Crear el asiento contable desde el comprobante
	_, err = s.journalEntryService.CreateFromVoucher(voucher, userID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error creando asiento contable: %v", err)
	}
	
	// Actualizar el estado del comprobante
	voucher.Status = "POSTED"
	voucher.PostedByUserID = &userID
	now := time.Now()
	voucher.PostedAt = &now
	
	// Guardar cambios del comprobante
	if err := tx.Save(voucher).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("error actualizando comprobante: %v", err)
	}
	
	// Commit de la transacción
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("error confirmando transacción: %v", err)
	}
	
	return nil
}

// Cancel cancela un comprobante con validaciones
func (s *VoucherService) Cancel(voucherID string) error {
	voucher, err := s.voucherRepo.GetByID(voucherID)
	if err != nil {
		return err
	}
	
	// Validar que el comprobante puede ser cancelado
	if voucher.Status == "POSTED" {
		return errors.New("no se pueden cancelar comprobantes ya contabilizados")
	}
	
	// TODO: Ejecutar reglas DSL para validar cancelación según
	// políticas específicas de la organización
	
	return s.voucherRepo.Cancel(voucherID)
}

// GetByDateRange obtiene comprobantes en un rango de fechas
func (s *VoucherService) GetByDateRange(orgID string, startDate, endDate time.Time, page, limit int) (*models.VouchersListResponse, error) {
	vouchers, total, err := s.voucherRepo.GetByDateRange(orgID, startDate, endDate, page, limit)
	if err != nil {
		return nil, err
	}
	
	pages := int((total + int64(limit) - 1) / int64(limit))
	
	return &models.VouchersListResponse{
		Vouchers: vouchers,
		Pagination: &models.PaginationInfo{
			Page:  page,
			Limit: limit,
			Total: int(total),
			Pages: pages,
		},
	}, nil
}

// GenerateFromTemplate genera un comprobante usando una plantilla DSL
// TODO: Este método usaría go-dsl para ejecutar plantillas de automatización
// y generar comprobantes complejos basados en reglas de negocio
func (s *VoucherService) GenerateFromTemplate(orgID, templateID string, variables map[string]interface{}) (*models.Voucher, error) {
	// TODO: Implementar la lógica de generación automática usando go-dsl:
	// 1. Cargar la plantilla DSL especificada
	// 2. Validar las variables de entrada
	// 3. Ejecutar el código DSL con las variables
	// 4. Generar el comprobante resultante
	// 5. Aplicar validaciones automáticas
	// 6. Retornar el comprobante generado
	
	return nil, errors.New("generación por plantilla DSL no implementada aún")
}

// validateVoucherLines valida que las líneas del comprobante estén balanceadas
func (s *VoucherService) validateVoucherLines(lines []models.VoucherLineRequest) error {
	var totalDebit, totalCredit float64
	
	if len(lines) < 2 {
		return errors.New("un comprobante debe tener al menos 2 líneas")
	}
	
	for _, line := range lines {
		totalDebit += line.DebitAmount
		totalCredit += line.CreditAmount
		
		// Validar que cada línea tenga movimiento en un solo lado
		if line.DebitAmount > 0 && line.CreditAmount > 0 {
			return errors.New("una línea no puede tener valores tanto en débito como en crédito")
		}
		
		if line.DebitAmount == 0 && line.CreditAmount == 0 {
			return errors.New("una línea debe tener valor en débito o crédito")
		}
	}
	
	// Validar que esté balanceado
	if totalDebit != totalCredit {
		return fmt.Errorf("el comprobante no está balanceado: débitos=%.2f, créditos=%.2f", totalDebit, totalCredit)
	}
	
	return nil
}

// validateAccounts valida que las cuentas existan y acepten movimiento
func (s *VoucherService) validateAccounts(orgID string, lines []models.VoucherLineRequest) error {
	for _, line := range lines {
		account, err := s.accountRepo.GetByID(line.AccountID)
		if err != nil {
			return fmt.Errorf("cuenta %s no encontrada", line.AccountID)
		}
		
		if account.OrganizationID != orgID {
			return fmt.Errorf("cuenta %s no pertenece a la organización", line.AccountID)
		}
		
		if !account.AcceptsMovement {
			return fmt.Errorf("la cuenta %s (%s) no acepta movimiento directo", account.Code, account.Name)
		}
		
		if !account.IsActive {
			return fmt.Errorf("la cuenta %s (%s) no está activa", account.Code, account.Name)
		}
	}
	
	return nil
}

// getCurrentPeriodID obtiene el ID del período actual
// TODO: Implementar lógica para obtener el período actual
func (s *VoucherService) getCurrentPeriodID(orgID string) string {
	// Placeholder - en la implementación real esto vendría de un servicio de períodos
	return "current-period-id"
}

// CountByDateRange cuenta comprobantes en un rango de fechas
func (s *VoucherService) CountByDateRange(orgID string, startDate, endDate time.Time) (int, error) {
	return s.voucherRepo.CountByDateRange(orgID, startDate, endDate)
}

// CountByStatus cuenta comprobantes por estado
func (s *VoucherService) CountByStatus(orgID string, status string) (int, error) {
	return s.voucherRepo.CountByStatus(orgID, status)
}

// Count cuenta el total de comprobantes
func (s *VoucherService) Count(orgID string) (int, error) {
	return s.voucherRepo.Count(orgID)
}

// CountByType obtiene el conteo de comprobantes por tipo
func (s *VoucherService) CountByType(orgID string) (map[string]int, error) {
	return s.voucherRepo.CountByType(orgID)
}