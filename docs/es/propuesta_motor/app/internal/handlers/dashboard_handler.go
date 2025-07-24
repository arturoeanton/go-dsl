package handlers

import (
	"motor-contable-poc/internal/models"
	"motor-contable-poc/internal/services"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// DashboardHandler maneja las peticiones HTTP para el dashboard
type DashboardHandler struct {
	voucherService *services.VoucherService
	orgService     *services.OrganizationService
}

// NewDashboardHandler crea una nueva instancia del handler
func NewDashboardHandler(db *gorm.DB) *DashboardHandler {
	return &DashboardHandler{
		voucherService: services.NewVoucherService(db),
		orgService:     services.NewOrganizationService(db),
	}
}

// GetStats obtiene las estadísticas del dashboard
// @Summary Estadísticas del dashboard
// @Description Retorna KPIs y estadísticas principales para el dashboard
// @Tags Dashboard
// @Accept json
// @Produce json
// @Success 200 {object} models.StandardResponse{data=models.DashboardStats} "Estadísticas del dashboard"
// @Failure 500 {object} models.StandardResponse "Error interno del servidor"
// @Router /api/v1/dashboard/stats [get]
func (h *DashboardHandler) GetStats(c *fiber.Ctx) error {
	// Obtener la organización actual
	org, err := h.orgService.GetCurrent()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(
				models.NewErrorResponse("ORG_NOT_FOUND", "Organización no encontrada"))
		}
		return c.Status(http.StatusInternalServerError).JSON(
			models.NewErrorResponse("INTERNAL_ERROR", "Error obteniendo organización"))
	}

	// Obtener estadísticas de comprobantes
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// Contar comprobantes del día
	vouchersToday, err := h.voucherService.CountByDateRange(org.ID, today, now)
	if err != nil {
		vouchersToday = 0 // No fallar por estadísticas
	}

	// Contar comprobantes del mes
	vouchersMonth, err := h.voucherService.CountByDateRange(org.ID, startOfMonth, now)
	if err != nil {
		vouchersMonth = 0
	}

	// Contar comprobantes pendientes
	pendingVouchers, err := h.voucherService.CountByStatus(org.ID, "DRAFT")
	if err != nil {
		pendingVouchers = 0
	}

	// Calcular tasa de procesamiento (simulada para el POC)
	processingRate := float64(85.5)
	if vouchersMonth > 0 {
		posted, _ := h.voucherService.CountByStatus(org.ID, "POSTED")
		if posted > 0 {
			processingRate = float64(posted) / float64(vouchersMonth) * 100
		}
	}

	// Generar datos de gráfico de los últimos 7 días
	chartLabels := make([]string, 7)
	chartValues := make([]int, 7)
	
	for i := 6; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		chartLabels[6-i] = date.Format("02/01")
		
		dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		dayEnd := dayStart.Add(24 * time.Hour)
		
		count, _ := h.voucherService.CountByDateRange(org.ID, dayStart, dayEnd)
		chartValues[6-i] = count
	}

	// Estadísticas por tipo (datos reales)
	countsByType, err := h.voucherService.CountByType(org.ID)
	if err != nil {
		countsByType = make(map[string]int) // No fallar por estadísticas
	}
	
	// Traducir tipos y preparar datos para el gráfico
	typeLabels := []string{}
	typeValues := []int{}
	
	typeTranslations := map[string]string{
		"PURCHASE": "Compras",
		"SALE":     "Ventas",
		"PAYMENT":  "Pagos",
		"RECEIPT":  "Recibos",
		"PAYROLL":  "Nómina",
		"EXPENSE":  "Gastos",
		"ADJUSTMENT": "Ajustes",
		"OTHER":    "Otros",
	}
	
	// Ordenar para consistencia
	typeOrder := []string{"SALE", "PURCHASE", "PAYMENT", "RECEIPT", "PAYROLL", "EXPENSE", "ADJUSTMENT", "OTHER"}
	
	for _, vType := range typeOrder {
		if count, exists := countsByType[vType]; exists && count > 0 {
			typeLabels = append(typeLabels, typeTranslations[vType])
			typeValues = append(typeValues, count)
		}
	}
	
	// Si no hay datos, usar valores por defecto
	if len(typeLabels) == 0 {
		typeLabels = []string{"Sin datos"}
		typeValues = []int{0}
	}
	
	vouchersByType := models.ChartData{
		Labels: typeLabels,
		Values: typeValues,
	}

	// Estadísticas por estado
	totalVouchers, _ := h.voucherService.Count(org.ID)
	postedVouchers, _ := h.voucherService.CountByStatus(org.ID, "POSTED")
	draftVouchers, _ := h.voucherService.CountByStatus(org.ID, "DRAFT")
	cancelledVouchers, _ := h.voucherService.CountByStatus(org.ID, "CANCELLED")
	errorVouchers, _ := h.voucherService.CountByStatus(org.ID, "ERROR")
	
	// Calcular otros estados no manejados
	otherVouchers := totalVouchers - postedVouchers - draftVouchers - cancelledVouchers - errorVouchers
	if otherVouchers < 0 {
		otherVouchers = 0
	}

	// Preparar datos para el gráfico de estado
	statusLabels := []string{}
	statusValues := []int{}
	
	if postedVouchers > 0 {
		statusLabels = append(statusLabels, "Contabilizados")
		statusValues = append(statusValues, postedVouchers)
	}
	if draftVouchers > 0 {
		statusLabels = append(statusLabels, "Borradores")
		statusValues = append(statusValues, draftVouchers)
	}
	if cancelledVouchers > 0 {
		statusLabels = append(statusLabels, "Cancelados")
		statusValues = append(statusValues, cancelledVouchers)
	}
	if errorVouchers > 0 {
		statusLabels = append(statusLabels, "Con Error")
		statusValues = append(statusValues, errorVouchers)
	}
	if otherVouchers > 0 {
		statusLabels = append(statusLabels, "Otros")
		statusValues = append(statusValues, otherVouchers)
	}
	
	// Si no hay datos
	if len(statusLabels) == 0 {
		statusLabels = []string{"Sin datos"}
		statusValues = []int{0}
	}

	vouchersByStatus := models.ChartData{
		Labels: statusLabels,
		Values: statusValues,
	}

	stats := models.DashboardStats{
		KPIs: models.DashboardKPIs{
			VouchersToday:    vouchersToday,
			VouchersMonth:    vouchersMonth,
			PendingVouchers:  pendingVouchers,
			ProcessingRate:   processingRate,
		},
		Charts: models.DashboardCharts{
			VouchersByDay: models.ChartData{
				Labels: chartLabels,
				Values: chartValues,
			},
			VouchersByType:   vouchersByType,
			VouchersByStatus: vouchersByStatus,
		},
		SystemHealth: models.SystemHealth{
			Status:           "healthy",
			Uptime:           99.9,
			APIResponseTime:  45,
			CacheHitRate:     92,
		},
	}

	return c.JSON(models.NewSuccessResponse(stats))
}

// GetActivity obtiene la actividad reciente del sistema
// @Summary Actividad reciente
// @Description Retorna las actividades más recientes del sistema
// @Tags Dashboard
// @Accept json
// @Produce json
// @Param limit query int false "Número de actividades" default(10)
// @Success 200 {object} models.StandardResponse{data=[]models.ActivityItem} "Lista de actividades"
// @Failure 500 {object} models.StandardResponse "Error interno del servidor"
// @Router /api/v1/dashboard/activity [get]
func (h *DashboardHandler) GetActivity(c *fiber.Ctx) error {
	// Obtener la organización actual
	org, err := h.orgService.GetCurrent()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(
				models.NewErrorResponse("ORG_NOT_FOUND", "Organización no encontrada"))
		}
		return c.Status(http.StatusInternalServerError).JSON(
			models.NewErrorResponse("INTERNAL_ERROR", "Error obteniendo organización"))
	}

	// Obtener últimos comprobantes para simular actividad
	// En una implementación real, tendríamos una tabla de actividades/logs
	vouchersList, err := h.voucherService.GetList(org.ID, 1, 10)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(
			models.NewErrorResponse("INTERNAL_ERROR", "Error obteniendo actividad"))
	}

	activities := make([]models.ActivityItem, 0)
	for _, voucher := range vouchersList.Vouchers {
		status := "info"
		if voucher.Status == "POSTED" {
			status = "success"
		} else if voucher.Status == "CANCELLED" {
			status = "error"
		}

		activity := models.ActivityItem{
			ID:          voucher.Number,
			Type:        getVoucherTypeForActivity(voucher.VoucherType),
			Status:      status,
			Description: voucher.Description,
			Amount:      voucher.TotalDebit, // Usar el mayor de débito o crédito
			CreatedAt:   voucher.CreatedAt,
		}

		if voucher.TotalCredit > voucher.TotalDebit {
			activity.Amount = voucher.TotalCredit
		}

		activities = append(activities, activity)
	}

	response := map[string]interface{}{
		"activities": activities,
	}

	return c.JSON(models.NewSuccessResponse(response))
}

// getVoucherTypeForActivity convierte el tipo de comprobante a formato de actividad
func getVoucherTypeForActivity(voucherType string) string {
	switch voucherType {
	case "SALE":
		return "invoice_sale"
	case "PURCHASE":
		return "invoice_purchase"
	case "PAYMENT":
		return "payment"
	case "RECEIPT":
		return "receipt"
	default:
		return "other"
	}
}

// GetDetailedStats obtiene estadísticas detalladas del sistema
// @Summary Estadísticas detalladas
// @Description Retorna estadísticas detalladas de comprobantes, asientos y cuentas
// @Tags Dashboard
// @Accept json
// @Produce json
// @Success 200 {object} models.StandardResponse{data=map[string]interface{}} "Estadísticas detalladas"
// @Failure 500 {object} models.StandardResponse "Error interno del servidor"
// @Router /api/v1/dashboard/stats/detailed [get]
func (h *DashboardHandler) GetDetailedStats(c *fiber.Ctx) error {
	// Obtener la organización actual
	org, err := h.orgService.GetCurrent()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(
				models.NewErrorResponse("ORG_NOT_FOUND", "Organización no encontrada"))
		}
		return c.Status(http.StatusInternalServerError).JSON(
			models.NewErrorResponse("INTERNAL_ERROR", "Error obteniendo organización"))
	}

	// Estadísticas de comprobantes por tipo
	vouchersByType, _ := h.voucherService.CountByType(org.ID)
	
	// Estadísticas de comprobantes por estado
	voucherStats := map[string]int{
		"draft":     0,
		"posted":    0,
		"cancelled": 0,
		"error":     0,
		"total":     0,
	}
	
	voucherStats["draft"], _ = h.voucherService.CountByStatus(org.ID, "DRAFT")
	voucherStats["posted"], _ = h.voucherService.CountByStatus(org.ID, "POSTED")
	voucherStats["cancelled"], _ = h.voucherService.CountByStatus(org.ID, "CANCELLED")
	voucherStats["error"], _ = h.voucherService.CountByStatus(org.ID, "ERROR")
	voucherStats["total"], _ = h.voucherService.Count(org.ID)
	
	// Estadísticas por período (últimos 30 días)
	monthlyStats := make([]map[string]interface{}, 30)
	now := time.Now()
	
	for i := 29; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		dayEnd := dayStart.Add(24 * time.Hour)
		
		count, _ := h.voucherService.CountByDateRange(org.ID, dayStart, dayEnd)
		
		monthlyStats[29-i] = map[string]interface{}{
			"date":  date.Format("2006-01-02"),
			"count": count,
		}
	}
	
	// Obtener estadísticas adicionales (simuladas para el POC)
	// En una implementación real, estas vendrían de servicios específicos
	accountStats := map[string]int{
		"total_accounts":    257, // Cuentas del PUC colombiano
		"active_accounts":   183,
		"inactive_accounts": 74,
		"detail_accounts":   198,
		"major_accounts":    59,
	}
	
	journalStats := map[string]int{
		"total_entries":     5, // Basado en los datos de demo
		"posted_entries":    5,
		"draft_entries":     0,
		"reversed_entries":  0,
	}
	
	// Resumen de salud del sistema
	systemMetrics := map[string]interface{}{
		"database_size_mb":    12.5,
		"active_users":        3,
		"last_backup":         "2024-01-15T03:00:00Z",
		"api_version":         "1.0.0",
		"uptime_hours":        168,
	}
	
	response := map[string]interface{}{
		"vouchers": map[string]interface{}{
			"by_type":       vouchersByType,
			"by_status":     voucherStats,
			"monthly_trend": monthlyStats,
		},
		"accounts": accountStats,
		"journal_entries": journalStats,
		"system": systemMetrics,
		"generated_at": time.Now(),
	}
	
	return c.JSON(models.NewSuccessResponse(response))
}

// RegisterRoutes registra las rutas del handler de dashboard
func (h *DashboardHandler) RegisterRoutes(router fiber.Router) {
	dashboard := router.Group("/dashboard")
	{
		dashboard.Get("/stats", h.GetStats)
		dashboard.Get("/stats/detailed", h.GetDetailedStats)
		dashboard.Get("/activity", h.GetActivity)
	}
}