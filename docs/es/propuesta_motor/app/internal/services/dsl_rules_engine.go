package services

import (
	"fmt"
	"motor-contable-poc/internal/models"
	"gorm.io/gorm"
	"time"
)

// DSLRulesEngine maneja la ejecución de reglas DSL simplificado para el POC
type DSLRulesEngine struct {
	db              *gorm.DB
	accountService  *AccountService
}

// NewDSLRulesEngine crea una nueva instancia del motor de reglas
func NewDSLRulesEngine(db *gorm.DB) *DSLRulesEngine {
	return &DSLRulesEngine{
		db:             db,
		accountService: NewAccountService(db),
	}
}

// ValidateVoucherPrePost valida un comprobante antes de ser procesado
func (e *DSLRulesEngine) ValidateVoucherPrePost(voucher *models.Voucher) error {
	// Validaciones específicas por tipo de comprobante
	switch voucher.VoucherType {
	case "invoice_sale":
		// Validar que tenga cliente
		if voucher.ThirdPartyID == nil {
			return fmt.Errorf("las facturas de venta requieren un cliente")
		}
		
	case "invoice_purchase":
		// Validar que tenga proveedor
		if voucher.ThirdPartyID == nil {
			return fmt.Errorf("las facturas de compra requieren un proveedor")
		}
		
	case "payment":
		// Validar que tenga cuenta bancaria
		hasBankAccount := false
		for _, line := range voucher.VoucherLines {
			account, _ := e.accountService.GetByID(line.AccountID)
			if account != nil && (account.Code[:4] == "1110" || account.Code[:4] == "1105") {
				hasBankAccount = true
				break
			}
		}
		if !hasBankAccount {
			return fmt.Errorf("los pagos requieren una cuenta bancaria o de caja")
		}
	}
	
	// Validaciones generales
	if voucher.TotalDebit != voucher.TotalCredit {
		return fmt.Errorf("el comprobante no está balanceado")
	}
	
	if len(voucher.VoucherLines) < 2 {
		return fmt.Errorf("el comprobante debe tener al menos 2 líneas")
	}
	
	return nil
}

// GenerateAutomaticLines genera líneas automáticas basadas en reglas DSL
func (e *DSLRulesEngine) GenerateAutomaticLines(voucher *models.Voucher) ([]models.VoucherLine, error) {
	additionalLines := []models.VoucherLine{}
	
	// Reglas para generar líneas de impuestos
	if voucher.VoucherType == "invoice_sale" || voucher.VoucherType == "invoice_purchase" {
		// Buscar líneas que representen base gravable (cuentas de ingreso o gasto)
		var baseAmount float64
		
		for _, line := range voucher.VoucherLines {
			account, err := e.accountService.GetByID(line.AccountID)
			if err != nil {
				continue
			}
			
			// Para ventas, buscar cuentas de ingreso (4xxx)
			if voucher.VoucherType == "invoice_sale" && account.Code[:1] == "4" {
				baseAmount += line.CreditAmount
			}
			
			// Para compras, buscar cuentas de gasto (5xxx) o inventario (14xx)
			if voucher.VoucherType == "invoice_purchase" && 
				(account.Code[:1] == "5" || account.Code[:2] == "14") {
				baseAmount += line.DebitAmount
			}
		}
		
		// Si hay base gravable, calcular IVA 19%
		if baseAmount > 0 {
			taxAmount := baseAmount * 0.19
			
			var taxLine models.VoucherLine
			
			if voucher.VoucherType == "invoice_sale" {
				// IVA por pagar (cuenta 240802)
				taxLine = models.VoucherLine{
					AccountID:    "d34b750ba305132c7196b47c4c528d6f", // 240802 - IVA
					Description:  fmt.Sprintf("IVA 19%% sobre ventas"),
					DebitAmount:  0,
					CreditAmount: taxAmount,
					TaxRate:      19,
					BaseAmount:   baseAmount,
					LineNumber:   len(voucher.VoucherLines) + 1,
				}
			} else {
				// IVA descontable (cuenta 240805)
				taxLine = models.VoucherLine{
					AccountID:    "a8f5c3d2e1b94f7a9d6e2c8b5a4f1e3d", // 240805 - IVA descontable
					Description:  fmt.Sprintf("IVA descontable 19%%"),
					DebitAmount:  taxAmount,
					CreditAmount: 0,
					TaxRate:      19,
					BaseAmount:   baseAmount,
					LineNumber:   len(voucher.VoucherLines) + 1,
				}
			}
			
			additionalLines = append(additionalLines, taxLine)
			
			// Para facturas de compra con monto alto, agregar retención
			if voucher.VoucherType == "invoice_purchase" && baseAmount > 1000000 {
				retentionAmount := baseAmount * 0.025 // 2.5% retención
				
				retentionLine := models.VoucherLine{
					AccountID:    "236540d8c89e5810e576249db7c95e7f", // Retención cuenta
					Description:  "Retención en la fuente 2.5%",
					DebitAmount:  0,
					CreditAmount: retentionAmount,
					LineNumber:   len(voucher.VoucherLines) + len(additionalLines) + 1,
				}
				
				additionalLines = append(additionalLines, retentionLine)
			}
		}
	}
	
	// Reglas para pagos con retención
	if voucher.VoucherType == "payment" && voucher.TotalDebit > 5000000 {
		// Buscar si ya tiene retención aplicada
		hasRetention := false
		for _, line := range voucher.VoucherLines {
			if line.AccountID == "236540d8c89e5810e576249db7c95e7f" {
				hasRetention = true
				break
			}
		}
		
		if !hasRetention {
			// Aplicar retención automática
			retentionAmount := voucher.TotalDebit * 0.035 // 3.5% para pagos grandes
			
			retentionLine := models.VoucherLine{
				AccountID:    "236540d8c89e5810e576249db7c95e7f",
				Description:  "Retención en la fuente 3.5% - Pago mayor",
				DebitAmount:  0,
				CreditAmount: retentionAmount,
				LineNumber:   len(voucher.VoucherLines) + 1,
			}
			
			additionalLines = append(additionalLines, retentionLine)
		}
	}
	
	return additionalLines, nil
}

// ApplyAutomaticClassifications aplica clasificaciones automáticas
func (e *DSLRulesEngine) ApplyAutomaticClassifications(voucher *models.Voucher) error {
	// Clasificaciones por tipo de comprobante
	classifications := map[string]interface{}{
		"dsl_processed": true,
		"processed_at": time.Now().Format(time.RFC3339),
	}
	
	switch voucher.VoucherType {
	case "invoice_sale":
		classifications["revenue_type"] = "operational"
		classifications["tax_regime"] = "common"
		classifications["requires_electronic_invoice"] = true
		
	case "invoice_purchase":
		classifications["expense_type"] = "operational"
		classifications["deductible"] = true
		classifications["requires_support_doc"] = true
		
	case "payment":
		classifications["payment_method"] = "bank_transfer"
		if voucher.TotalDebit > 10000000 {
			classifications["requires_dual_approval"] = true
		} else if voucher.TotalDebit > 5000000 {
			classifications["requires_approval"] = true
		}
		
	case "receipt":
		classifications["receipt_type"] = "cash"
		classifications["requires_deposit"] = voucher.TotalDebit > 2000000
	}
	
	// Aplicar centro de costo automático basado en cuentas
	for i, line := range voucher.VoucherLines {
		account, err := e.accountService.GetByID(line.AccountID)
		if err != nil {
			continue
		}
		
		// Cuentas de gastos (5xxx) requieren centro de costo
		if account.Code[:1] == "5" && line.CostCenterID == nil {
			// Asignar centro de costo por defecto según el tipo de gasto
			var costCenter string
			switch account.Code[:3] {
			case "510": // Gastos de personal
				costCenter = "cc-rrhh-001"
			case "511": // Honorarios
				costCenter = "cc-admin-001"
			case "512": // Impuestos
				costCenter = "cc-financiero-001"
			case "513": // Arrendamientos
				costCenter = "cc-admin-001"
			case "514": // Seguros
				costCenter = "cc-admin-001"
			case "515": // Servicios
				costCenter = "cc-operaciones-001"
			default:
				costCenter = "cc-general-001"
			}
			voucher.VoucherLines[i].CostCenterID = &costCenter
		}
		
		// Cuentas de inventario requieren referencia de producto
		if account.Code[:2] == "14" && line.CostCenterID == nil {
			costCenter := "cc-inventario-001"
			voucher.VoucherLines[i].CostCenterID = &costCenter
		}
	}
	
	// Aplicar las clasificaciones al voucher
	existingData, _ := voucher.GetAdditionalData()
	if existingData == nil {
		existingData = &models.VoucherAdditionalData{}
	}
	if existingData.CustomFields == nil {
		existingData.CustomFields = make(map[string]interface{})
	}
	for k, v := range classifications {
		existingData.CustomFields[k] = v
	}
	// Actualizar algunos campos específicos
	existingData.AutoGenerated = true
	voucher.SetAdditionalData(*existingData)
	
	return nil
}

// CheckWorkflowRequirements verifica si se requieren workflows de aprobación
func (e *DSLRulesEngine) CheckWorkflowRequirements(voucher *models.Voucher) (bool, string, error) {
	// Reglas de workflow basadas en montos
	if voucher.TotalDebit > 50000000 {
		return true, "CEO_APPROVAL", nil
	}
	
	if voucher.TotalDebit > 20000000 {
		return true, "CFO_APPROVAL", nil
	}
	
	if voucher.TotalDebit > 10000000 {
		return true, "DUAL_APPROVAL", nil
	}
	
	// Reglas específicas por tipo
	switch voucher.VoucherType {
	case "payment":
		if voucher.TotalDebit > 5000000 {
			return true, "PAYMENT_APPROVAL", nil
		}
		
		// Pagos a cuentas sensibles
		for _, line := range voucher.VoucherLines {
			account, _ := e.accountService.GetByID(line.AccountID)
			if account != nil {
				// Salidas de caja o bancos
				if (account.Code[:4] == "1105" || account.Code[:4] == "1110") && line.CreditAmount > 0 {
					if line.CreditAmount > 1000000 {
						return true, "CASH_MOVEMENT_APPROVAL", nil
					}
				}
			}
		}
		
	case "invoice_purchase":
		// Compras grandes requieren aprobación
		if voucher.TotalDebit > 15000000 {
			return true, "PURCHASE_APPROVAL", nil
		}
		
	case "journal_entry":
		// Asientos manuales siempre requieren aprobación
		return true, "MANUAL_ENTRY_APPROVAL", nil
	}
	
	// Verificar si hay ajustes a cuentas críticas
	criticalAccounts := []string{"1105", "1110", "3105", "3605"} // Caja, Bancos, Capital, Utilidades
	for _, line := range voucher.VoucherLines {
		account, _ := e.accountService.GetByID(line.AccountID)
		if account != nil {
			for _, critical := range criticalAccounts {
				if account.Code[:4] == critical {
					return true, "CRITICAL_ACCOUNT_APPROVAL", nil
				}
			}
		}
	}
	
	return false, "", nil
}

// ExecutePostProcessing ejecuta acciones post-procesamiento
func (e *DSLRulesEngine) ExecutePostProcessing(voucher *models.Voucher) error {
	notifications := []string{}
	
	// Notificaciones basadas en montos
	if voucher.TotalDebit > 100000000 {
		notifications = append(notifications, 
			fmt.Sprintf("🚨 ALERTA CRÍTICA: Comprobante de muy alto valor procesado: %s por $%.2f", 
				voucher.Number, voucher.TotalDebit))
	} else if voucher.TotalDebit > 50000000 {
		notifications = append(notifications, 
			fmt.Sprintf("⚠️ ALERTA: Comprobante de alto valor procesado: %s por $%.2f", 
				voucher.Number, voucher.TotalDebit))
	}
	
	// Notificaciones por tipo
	switch voucher.VoucherType {
	case "payment":
		if voucher.TotalDebit > 10000000 {
			notifications = append(notifications, 
				fmt.Sprintf("💰 Pago importante procesado: %s - %s por $%.2f", 
					voucher.Number, voucher.Description, voucher.TotalDebit))
		}
		
	case "invoice_sale":
		if voucher.TotalCredit > 50000000 {
			notifications = append(notifications, 
				fmt.Sprintf("💵 Venta importante registrada: %s por $%.2f", 
					voucher.Number, voucher.TotalCredit))
		}
		
	case "invoice_purchase":
		// Notificar compras con retención
		for _, line := range voucher.VoucherLines {
			if line.AccountID == "236540d8c89e5810e576249db7c95e7f" && line.CreditAmount > 0 {
				notifications = append(notifications, 
					fmt.Sprintf("📋 Compra con retención: %s - Retención: $%.2f", 
						voucher.Number, line.CreditAmount))
				break
			}
		}
	}
	
	// Notificaciones de cuentas críticas
	for _, line := range voucher.VoucherLines {
		account, _ := e.accountService.GetByID(line.AccountID)
		if account != nil {
			// Movimientos en caja
			if account.Code[:4] == "1105" && line.CreditAmount > 5000000 {
				notifications = append(notifications, 
					fmt.Sprintf("💸 Salida de caja significativa: $%.2f", line.CreditAmount))
			}
			
			// Movimientos en bancos
			if account.Code[:4] == "1110" && line.CreditAmount > 10000000 {
				notifications = append(notifications, 
					fmt.Sprintf("🏦 Transferencia bancaria importante: $%.2f", line.CreditAmount))
			}
		}
	}
	
	// Simular envío de notificaciones (en producción se integraría con sistema real)
	for _, notification := range notifications {
		fmt.Printf("[DSL NOTIFICACIÓN] %s\n", notification)
		// Aquí se integraría con sistema de notificaciones real:
		// - Email
		// - SMS
		// - Slack/Teams
		// - Dashboard alerts
	}
	
	// Registrar post-procesamiento en metadata
	postProcessData := map[string]interface{}{
		"post_process_completed": true,
		"post_process_timestamp": time.Now().Format(time.RFC3339),
		"notifications_sent": len(notifications),
		"dsl_rules_applied": []string{
			"tax_calculation",
			"retention_rules", 
			"workflow_validation",
			"cost_center_assignment",
			"notification_rules",
		},
	}
	
	// Log especial para demo
	if len(notifications) > 0 {
		fmt.Println("\n========== DSL POST-PROCESAMIENTO ==========")
		fmt.Printf("Comprobante: %s procesado exitosamente\n", voucher.Number)
		fmt.Printf("Reglas DSL aplicadas: %d\n", len(postProcessData["dsl_rules_applied"].([]string)))
		fmt.Printf("Notificaciones enviadas: %d\n", len(notifications))
		fmt.Println("===========================================\n")
	}
	
	existingData, _ := voucher.GetAdditionalData()
	if existingData == nil {
		existingData = &models.VoucherAdditionalData{}
	}
	if existingData.CustomFields == nil {
		existingData.CustomFields = make(map[string]interface{})
	}
	for k, v := range postProcessData {
		existingData.CustomFields[k] = v
	}
	voucher.SetAdditionalData(*existingData)
	
	return nil
}