package handlers

import (
	"log"
	"motor-contable-poc/internal/models"
	"motor-contable-poc/internal/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// VoucherHandler maneja las peticiones HTTP para comprobantes
type VoucherHandler struct {
	voucherService *services.VoucherService
	orgService     *services.OrganizationService
}

// NewVoucherHandler crea una nueva instancia del handler
func NewVoucherHandler(db *gorm.DB) *VoucherHandler {
	return &VoucherHandler{
		voucherService: services.NewVoucherService(db),
		orgService:     services.NewOrganizationService(db),
	}
}

// CalculateIVA calcula el IVA dinámicamente para el POS
// @Summary Calcular IVA
// @Description Calcula el IVA dinámicamente basado en el subtotal usando las reglas DSL
// @Tags Comprobantes
// @Accept json
// @Produce json
// @Param request body object{subtotal=float64,voucher_type=string} true "Datos para calcular IVA"
// @Success 200 {object} models.StandardResponse{data=object{subtotal=float64,iva_rate=float64,iva_amount=float64,total=float64}} "Cálculo de IVA"
// @Failure 400 {object} models.StandardResponse "Parámetros inválidos"
// @Failure 500 {object} models.StandardResponse "Error interno del servidor"
// @Router /api/v1/vouchers/calculate-iva [post]
func (h *VoucherHandler) CalculateIVA(c *fiber.Ctx) error {
	log.Printf("[INFO] VoucherHandler.CalculateIVA: Iniciando cálculo de IVA")
	
	// Verificar que el DSL engine existe
	if h.voucherService == nil {
		log.Printf("[ERROR] VoucherHandler.CalculateIVA: voucherService es nil")
		return c.Status(http.StatusInternalServerError).JSON(
			models.NewErrorResponse("INTERNAL_ERROR", "Servicio no disponible"))
	}
	
	var request struct {
		Subtotal    float64 `json:"subtotal" validate:"required,min=0"`
		VoucherType string  `json:"voucher_type" validate:"required"`
	}

	if err := c.BodyParser(&request); err != nil {
		log.Printf("[ERROR] VoucherHandler.CalculateIVA: Error parseando request - %v", err)
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("INVALID_REQUEST", "Datos de entrada inválidos"))
	}

	if request.Subtotal <= 0 {
		log.Printf("[ERROR] VoucherHandler.CalculateIVA: Subtotal inválido - %f", request.Subtotal)
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("INVALID_SUBTOTAL", "El subtotal debe ser mayor a 0"))
	}

	// Obtener la tasa de IVA actual del DSL
	dslEngine := h.voucherService.GetDSLEngine()
	ivaRate := dslEngine.GetIVARate()
	
	// Calcular IVA
	ivaAmount := request.Subtotal * ivaRate
	total := request.Subtotal + ivaAmount

	log.Printf("[INFO] VoucherHandler.CalculateIVA: Subtotal=%.2f, Tasa=%.2f%%, IVA=%.2f, Total=%.2f", 
		request.Subtotal, ivaRate*100, ivaAmount, total)

	response := map[string]interface{}{
		"subtotal":   request.Subtotal,
		"iva_rate":   ivaRate,
		"iva_amount": ivaAmount,
		"total":      total,
		"currency":   "COP",
		"calculated_at": time.Now().Format(time.RFC3339),
	}

	return c.JSON(models.NewSuccessResponse(response))
}

// RecalculateWithDSL recalcula un comprobante aplicando reglas DSL
// @Summary Recalcular comprobante con DSL
// @Description Recalcula un comprobante aplicando las reglas DSL actuales
// @Tags Comprobantes
// @Accept json
// @Produce json
// @Param id path string true "ID del comprobante"
// @Success 200 {object} models.StandardResponse{data=models.Voucher} "Comprobante recalculado"
// @Failure 400 {object} models.StandardResponse "Parámetros inválidos"
// @Failure 404 {object} models.StandardResponse "Comprobante no encontrado"
// @Failure 500 {object} models.StandardResponse "Error interno del servidor"
// @Router /api/v1/vouchers/{id}/recalculate [post]
func (h *VoucherHandler) RecalculateWithDSL(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		log.Printf("[ERROR] VoucherHandler.RecalculateWithDSL: ID faltante")
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("MISSING_ID", "ID del comprobante requerido"))
	}

	log.Printf("[INFO] VoucherHandler.RecalculateWithDSL: Recalculando comprobante %s", id)

	voucher, err := h.voucherService.RecalculateWithDSL(id)
	if err != nil {
		log.Printf("[ERROR] VoucherHandler.RecalculateWithDSL: Error recalculando comprobante %s - %v", id, err)
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("RECALCULATE_ERROR", err.Error()))
	}

	log.Printf("[INFO] VoucherHandler.RecalculateWithDSL: Comprobante %s recalculado exitosamente", id)
	return c.JSON(models.NewSuccessResponse(voucher))
}

// GetList obtiene una lista paginada de comprobantes
// @Summary Lista de comprobantes
// @Description Retorna una lista paginada de comprobantes de la organización
// @Tags Comprobantes
// @Accept json
// @Produce json
// @Param page query int false "Número de página" default(1)
// @Param limit query int false "Elementos por página" default(20)
// @Success 200 {object} models.StandardResponse{data=models.VouchersListResponse} "Lista de comprobantes"
// @Failure 400 {object} models.StandardResponse "Parámetros inválidos"
// @Failure 500 {object} models.StandardResponse "Error interno del servidor"
// @Router /api/v1/vouchers [get]
func (h *VoucherHandler) GetList(c *fiber.Ctx) error {
	// Obtener la organización actual desde la base de datos
	org, err := h.orgService.GetCurrent()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("[ERROR] VoucherHandler.GetList: Organización no encontrada - %v", err)
			return c.Status(http.StatusNotFound).JSON(
				models.NewErrorResponse("ORG_NOT_FOUND", "Organización no encontrada"))
		}
		log.Printf("[ERROR] VoucherHandler.GetList: Error obteniendo organización - %v", err)
		return c.Status(http.StatusInternalServerError).JSON(
			models.NewErrorResponse("INTERNAL_ERROR", "Error obteniendo organización"))
	}
	
	// Obtener parámetros de paginación
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	
	// TODO: En el futuro, se usaría go-dsl para filtros dinámicos
	// basados en permisos de usuario y configuraciones de vista
	
	response, err := h.voucherService.GetList(org.ID, page, limit)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(
			models.NewErrorResponse("INTERNAL_ERROR", "Error obteniendo lista de comprobantes"))
	}
	
	return c.JSON(models.NewSuccessResponse(response))
}

// GetByID obtiene un comprobante específico
// @Summary Detalle de comprobante
// @Description Retorna el detalle completo de un comprobante específico
// @Tags Comprobantes
// @Accept json
// @Produce json
// @Param id path string true "ID del comprobante"
// @Success 200 {object} models.StandardResponse{data=models.VoucherDetail} "Detalle del comprobante"
// @Failure 404 {object} models.StandardResponse "Comprobante no encontrado"
// @Failure 500 {object} models.StandardResponse "Error interno del servidor"
// @Router /api/v1/vouchers/{id} [get]
func (h *VoucherHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("MISSING_ID", "ID del comprobante requerido"))
	}
	
	// TODO: Validar permisos de acceso usando reglas DSL
	
	detail, err := h.voucherService.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(
				models.NewErrorResponse("VOUCHER_NOT_FOUND", "Comprobante no encontrado"))
		}
		return c.Status(http.StatusInternalServerError).JSON(
			models.NewErrorResponse("INTERNAL_ERROR", "Error obteniendo comprobante"))
	}
	
	return c.JSON(models.NewSuccessResponse(detail))
}

// Create crea un nuevo comprobante
// @Summary Crear comprobante
// @Description Crea un nuevo comprobante contable con sus líneas
// @Tags Comprobantes
// @Accept json
// @Produce json
// @Param voucher body models.VoucherCreateRequest true "Datos del comprobante"
// @Success 201 {object} models.StandardResponse{data=models.Voucher} "Comprobante creado"
// @Failure 400 {object} models.StandardResponse "Datos inválidos"
// @Failure 500 {object} models.StandardResponse "Error interno del servidor"
// @Router /api/v1/vouchers [post]
func (h *VoucherHandler) Create(c *fiber.Ctx) error {
	// Obtener la organización actual desde la base de datos
	org, err := h.orgService.GetCurrent()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(
				models.NewErrorResponse("ORG_NOT_FOUND", "Organización no encontrada"))
		}
		return c.Status(http.StatusInternalServerError).JSON(
			models.NewErrorResponse("INTERNAL_ERROR", "Error obteniendo organización"))
	}
	
	var request models.VoucherCreateRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("INVALID_JSON", "Formato JSON inválido"))
	}
	
	// TODO: Validar permisos de creación usando reglas DSL
	
	// TODO: En el futuro, aquí se usaría go-dsl para:
	// 1. Validar reglas de negocio específicas del tipo de comprobante
	// 2. Aplicar plantillas de automatización
	// 3. Generar líneas adicionales automáticamente (impuestos, etc.)
	// 4. Ejecutar workflows de aprobación si son requeridos
	
	voucher, err := h.voucherService.Create(org.ID, request)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("CREATION_ERROR", err.Error()))
	}
	
	return c.Status(http.StatusCreated).JSON(models.NewSuccessResponse(voucher))
}

// Post contabiliza un comprobante
// @Summary Contabilizar comprobante
// @Description Cambia el estado del comprobante a contabilizado y genera asiento
// @Tags Comprobantes
// @Accept json
// @Produce json
// @Param id path string true "ID del comprobante"
// @Success 200 {object} models.StandardResponse "Comprobante contabilizado exitosamente"
// @Failure 400 {object} models.StandardResponse "No se puede contabilizar"
// @Failure 404 {object} models.StandardResponse "Comprobante no encontrado"
// @Failure 500 {object} models.StandardResponse "Error interno del servidor"
// @Router /api/v1/vouchers/{id}/post [post]
func (h *VoucherHandler) Post(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("MISSING_ID", "ID del comprobante requerido"))
	}
	
	// TODO: Obtener userID del contexto de usuario autenticado
	userID := "system-user" // Placeholder para el POC
	
	// TODO: Validar permisos de contabilización usando reglas DSL
	
	// TODO: En el futuro, este proceso usaría go-dsl para:
	// 1. Validar reglas contables complejas antes de contabilizar
	// 2. Generar automáticamente el asiento contable correspondiente
	// 3. Aplicar clasificaciones y distribuciones automáticas
	// 4. Ejecutar procesos post-contabilización (notificaciones, etc.)
	
	err := h.voucherService.Post(id, userID)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("POST_ERROR", err.Error()))
	}
	
	return c.JSON(models.NewSuccessResponse(map[string]interface{}{
		"message": "Comprobante contabilizado exitosamente",
		"voucher_id": id,
		"posted_at": time.Now(),
	}))
}

// Cancel cancela un comprobante
// @Summary Cancelar comprobante
// @Description Cambia el estado del comprobante a cancelado
// @Tags Comprobantes
// @Accept json
// @Produce json
// @Param id path string true "ID del comprobante"
// @Success 200 {object} models.StandardResponse "Comprobante cancelado exitosamente"
// @Failure 400 {object} models.StandardResponse "No se puede cancelar"
// @Failure 404 {object} models.StandardResponse "Comprobante no encontrado"
// @Failure 500 {object} models.StandardResponse "Error interno del servidor"
// @Router /api/v1/vouchers/{id}/cancel [post]
func (h *VoucherHandler) Cancel(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("MISSING_ID", "ID del comprobante requerido"))
	}
	
	// TODO: Validar permisos de cancelación usando reglas DSL
	
	err := h.voucherService.Cancel(id)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("CANCEL_ERROR", err.Error()))
	}
	
	return c.JSON(models.NewSuccessResponse(map[string]interface{}{
		"message": "Comprobante cancelado exitosamente",
		"voucher_id": id,
		"cancelled_at": time.Now(),
	}))
}

// ApplyTemplate aplica un template DSL a un comprobante existente
// @Summary Aplicar template DSL a comprobante
// @Description Aplica un template DSL para generar líneas automáticas o asientos contables
// @Tags Comprobantes
// @Accept json
// @Produce json
// @Param id path string true "ID del comprobante"
// @Param body body models.ApplyTemplateRequest true "Template y parámetros"
// @Success 200 {object} models.StandardResponse "Template aplicado exitosamente"
// @Failure 400 {object} models.StandardResponse "Parámetros inválidos"
// @Failure 404 {object} models.StandardResponse "Comprobante o template no encontrado"
// @Failure 500 {object} models.StandardResponse "Error interno del servidor"
// @Router /api/v1/vouchers/{id}/apply-template [post]
func (h *VoucherHandler) ApplyTemplate(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("MISSING_ID", "ID del comprobante requerido"))
	}
	
	var request struct {
		TemplateID string                 `json:"template_id" validate:"required"`
		Parameters map[string]interface{} `json:"parameters"`
	}
	
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("INVALID_BODY", "Error al procesar la solicitud"))
	}
	
	if request.TemplateID == "" {
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("MISSING_TEMPLATE", "ID del template requerido"))
	}
	
	err := h.voucherService.ApplyTemplate(id, request.TemplateID, request.Parameters)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("APPLY_ERROR", err.Error()))
	}
	
	return c.JSON(models.NewSuccessResponse(map[string]interface{}{
		"message": "Template aplicado exitosamente",
		"voucher_id": id,
		"template_id": request.TemplateID,
		"applied_at": time.Now(),
	}))
}

// CreateFromTemplate crea un nuevo comprobante usando un template DSL
// @Summary Crear comprobante desde template
// @Description Crea un nuevo comprobante usando un template DSL predefinido
// @Tags Comprobantes
// @Accept json
// @Produce json
// @Param body body models.CreateFromTemplateRequest true "Template y parámetros"
// @Success 201 {object} models.StandardResponse{data=models.Voucher} "Comprobante creado"
// @Failure 400 {object} models.StandardResponse "Parámetros inválidos"
// @Failure 404 {object} models.StandardResponse "Template no encontrado"
// @Failure 500 {object} models.StandardResponse "Error interno del servidor"
// @Router /api/v1/vouchers/from-template [post]
func (h *VoucherHandler) CreateFromTemplate(c *fiber.Ctx) error {
	var request struct {
		TemplateID string                 `json:"template_id" validate:"required"`
		Parameters map[string]interface{} `json:"parameters"`
	}
	
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("INVALID_BODY", "Error al procesar la solicitud"))
	}
	
	// Obtener la organización actual
	org, err := h.orgService.GetCurrent()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(
			models.NewErrorResponse("ORG_ERROR", "Error obteniendo organización"))
	}
	
	voucher, err := h.voucherService.CreateFromTemplate(org.ID, request.TemplateID, request.Parameters)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("CREATE_ERROR", err.Error()))
	}
	
	return c.Status(http.StatusCreated).JSON(models.NewSuccessResponse(voucher))
}

// GetByDateRange obtiene comprobantes por rango de fechas
// @Summary Comprobantes por fechas
// @Description Retorna comprobantes filtrados por rango de fechas
// @Tags Comprobantes
// @Accept json
// @Produce json
// @Param start_date query string true "Fecha inicial (YYYY-MM-DD)"
// @Param end_date query string true "Fecha final (YYYY-MM-DD)"
// @Param page query int false "Número de página" default(1)
// @Param limit query int false "Elementos por página" default(20)
// @Success 200 {object} models.StandardResponse{data=models.VouchersListResponse} "Lista de comprobantes"
// @Failure 400 {object} models.StandardResponse "Parámetros inválidos"
// @Failure 500 {object} models.StandardResponse "Error interno del servidor"
// @Router /api/v1/vouchers/by-date-range [get]
func (h *VoucherHandler) GetByDateRange(c *fiber.Ctx) error {
	// Obtener la organización actual desde la base de datos
	org, err := h.orgService.GetCurrent()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(
				models.NewErrorResponse("ORG_NOT_FOUND", "Organización no encontrada"))
		}
		return c.Status(http.StatusInternalServerError).JSON(
			models.NewErrorResponse("INTERNAL_ERROR", "Error obteniendo organización"))
	}
	
	// Obtener y validar parámetros de fecha
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	
	if startDateStr == "" || endDateStr == "" {
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("MISSING_DATES", "Fechas inicial y final requeridas"))
	}
	
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("INVALID_START_DATE", "Formato de fecha inicial inválido (usar YYYY-MM-DD)"))
	}
	
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("INVALID_END_DATE", "Formato de fecha final inválido (usar YYYY-MM-DD)"))
	}
	
	// Obtener parámetros de paginación
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	
	response, err := h.voucherService.GetByDateRange(org.ID, startDate, endDate, page, limit)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(
			models.NewErrorResponse("INTERNAL_ERROR", "Error obteniendo comprobantes por fecha"))
	}
	
	return c.JSON(models.NewSuccessResponse(response))
}

// GenerateFromTemplate genera un comprobante usando una plantilla DSL
// @Summary Generar desde plantilla
// @Description Genera un comprobante automáticamente usando una plantilla DSL
// @Tags Comprobantes
// @Accept json
// @Produce json
// @Param generation body map[string]interface{} true "Datos para generación"
// @Success 201 {object} models.StandardResponse{data=models.Voucher} "Comprobante generado"
// @Failure 400 {object} models.StandardResponse "Datos inválidos"
// @Failure 500 {object} models.StandardResponse "Error interno del servidor"
// @Router /api/v1/vouchers/generate [post]
func (h *VoucherHandler) GenerateFromTemplate(c *fiber.Ctx) error {
	// Obtener la organización actual desde la base de datos
	org, err := h.orgService.GetCurrent()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(
				models.NewErrorResponse("ORG_NOT_FOUND", "Organización no encontrada"))
		}
		return c.Status(http.StatusInternalServerError).JSON(
			models.NewErrorResponse("INTERNAL_ERROR", "Error obteniendo organización"))
	}
	
	var request struct {
		TemplateID string                 `json:"template_id"`
		Variables  map[string]interface{} `json:"variables"`
	}
	
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("INVALID_JSON", "Formato JSON inválido"))
	}
	
	if request.TemplateID == "" {
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("MISSING_TEMPLATE", "ID de plantilla requerido"))
	}
	
	// TODO: En el futuro, este endpoint usaría go-dsl para:
	// 1. Cargar y validar la plantilla DSL especificada
	// 2. Validar las variables de entrada contra el esquema de la plantilla
	// 3. Ejecutar el código DSL para generar el comprobante
	// 4. Aplicar todas las validaciones y reglas de negocio
	// 5. Retornar el comprobante generado listo para revisar/contabilizar
	
	voucher, err := h.voucherService.GenerateFromTemplate(org.ID, request.TemplateID, request.Variables)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("GENERATION_ERROR", err.Error()))
	}
	
	return c.Status(http.StatusCreated).JSON(models.NewSuccessResponse(voucher))
}

// RegisterRoutes registra las rutas del handler de comprobantes
func (h *VoucherHandler) RegisterRoutes(router fiber.Router) {
	vouchers := router.Group("/vouchers")
	{
		vouchers.Get("/", h.GetList)
		vouchers.Post("/", h.Create)
		vouchers.Post("/calculate-iva", h.CalculateIVA)
		vouchers.Get("/by-date-range", h.GetByDateRange)
		vouchers.Post("/generate", h.GenerateFromTemplate)
		vouchers.Post("/from-template", h.CreateFromTemplate)
		vouchers.Get("/:id", h.GetByID)
		vouchers.Post("/:id/post", h.Post)
		vouchers.Post("/:id/cancel", h.Cancel)
		vouchers.Post("/:id/recalculate", h.RecalculateWithDSL)
		vouchers.Post("/:id/apply-template", h.ApplyTemplate)
	}
}