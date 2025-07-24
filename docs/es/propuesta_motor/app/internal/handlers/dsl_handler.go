package handlers

import (
	"log"
	"motor-contable-poc/internal/models"
	"motor-contable-poc/internal/services"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// DSLHandler maneja las peticiones HTTP para plantillas DSL
type DSLHandler struct {
	db *gorm.DB
	dslEngine *services.DSLRulesEngine
}

// NewDSLHandler crea una nueva instancia del handler
func NewDSLHandler(db *gorm.DB, dslEngine *services.DSLRulesEngine) *DSLHandler {
	return &DSLHandler{
		db: db,
		dslEngine: dslEngine,
	}
}

// GetTemplates obtiene la lista de plantillas DSL
// @Summary Lista de plantillas DSL
// @Description Retorna todas las plantillas DSL disponibles
// @Tags DSL
// @Accept json
// @Produce json
// @Success 200 {object} models.StandardResponse{data=[]models.DSLTemplate} "Lista de plantillas"
// @Failure 500 {object} models.StandardResponse "Error interno del servidor"
// @Router /api/v1/dsl/templates [get]
func (h *DSLHandler) GetTemplates(c *fiber.Ctx) error {
	log.Printf("[INFO] DSLHandler.GetTemplates: Iniciando obtención de plantillas DSL")
	// TODO: En el futuro, aquí se integraría go-dsl para gestionar plantillas
	// Por ahora retornamos plantillas de ejemplo
	templates := []models.DSLTemplate{
		{
			BaseModel: models.BaseModel{ID: "tpl-001"},
			Name:      "Factura de Venta",
			Category:  "VOUCHER",
			Description: "Plantilla para registro de facturas de venta",
			Status:    "ACTIVE",
			Version:   "1.0.0",
			IsPublic:  true,
		},
		{
			BaseModel: models.BaseModel{ID: "tpl-002"},
			Name:      "Nómina Mensual",
			Category:  "PAYROLL",
			Description: "Plantilla para cálculo y registro de nómina",
			Status:    "ACTIVE",
			Version:   "1.0.0",
			IsPublic:  true,
		},
		{
			BaseModel: models.BaseModel{ID: "tpl-003"},
			Name:      "Depreciación de Activos",
			Category:  "ASSET",
			Description: "Plantilla para cálculo automático de depreciación",
			Status:    "ACTIVE",
			Version:   "1.0.0",
			IsPublic:  true,
		},
	}

	log.Printf("[INFO] DSLHandler.GetTemplates: Retornando %d plantillas", len(templates))
	return c.JSON(models.NewSuccessResponse(templates))
}

// GetTemplateByID obtiene una plantilla específica
// @Summary Detalle de plantilla DSL
// @Description Retorna el detalle de una plantilla DSL específica
// @Tags DSL
// @Accept json
// @Produce json
// @Param id path string true "ID de la plantilla"
// @Success 200 {object} models.StandardResponse{data=models.DSLTemplate} "Detalle de la plantilla"
// @Failure 404 {object} models.StandardResponse "Plantilla no encontrada"
// @Failure 500 {object} models.StandardResponse "Error interno del servidor"
// @Router /api/v1/dsl/templates/{id} [get]
func (h *DSLHandler) GetTemplateByID(c *fiber.Ctx) error {
	id := c.Params("id")
	log.Printf("[INFO] DSLHandler.GetTemplateByID: Obteniendo plantilla %s", id)
	if id == "" {
		log.Printf("[ERROR] DSLHandler.GetTemplateByID: ID faltante")
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("MISSING_ID", "ID de plantilla requerido"))
	}

	// TODO: Buscar en base de datos
	// Por ahora retornamos un ejemplo
	template := models.DSLTemplate{
		BaseModel:   models.BaseModel{ID: id},
		Name:        "Factura de Venta",
		Category:    "VOUCHER",
		Description: "Plantilla para registro de facturas de venta",
		Status:      "ACTIVE",
		Version:     "1.0.0",
		IsPublic:    true,
	}

	// Agregar código DSL de ejemplo directamente
	template.DSLCode = `// Plantilla de Factura de Venta
voucher {
    type: "SALE_INVOICE"
    
    // Línea de ingreso por venta
    line {
        account: "4135"  // Ingresos por ventas
        amount: voucher.subtotal
        side: "credit"
    }
    
    // Línea de IVA
    if voucher.tax > 0 {
        line {
            account: "2408"  // IVA por pagar
            amount: voucher.tax
            side: "credit"
        }
    }
    
    // Línea de cliente
    line {
        account: "130505"  // Clientes nacionales
        amount: voucher.total
        side: "debit"
        third_party: voucher.customer_id
    }
}`

	log.Printf("[INFO] DSLHandler.GetTemplateByID: Plantilla %s encontrada", id)
	return c.JSON(models.NewSuccessResponse(template))
}

// ValidateDSL valida código DSL
// @Summary Validar código DSL
// @Description Valida que el código DSL sea sintácticamente correcto
// @Tags DSL
// @Accept json
// @Produce json
// @Param code body models.DSLValidateRequest true "Código DSL a validar"
// @Success 200 {object} models.StandardResponse{data=models.DSLValidateResponse} "Resultado de validación"
// @Failure 400 {object} models.StandardResponse "Código inválido"
// @Failure 500 {object} models.StandardResponse "Error interno del servidor"
// @Router /api/v1/dsl/validate [post]
func (h *DSLHandler) ValidateDSL(c *fiber.Ctx) error {
	log.Printf("[INFO] DSLHandler.ValidateDSL: Iniciando validación de código DSL")
	var request struct {
		Code string `json:"code" binding:"required"`
	}

	if err := c.BodyParser(&request); err != nil {
		log.Printf("[ERROR] DSLHandler.ValidateDSL: Error parseando request - %v", err)
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("INVALID_JSON", "Formato JSON inválido"))
	}

	// TODO: Aquí se integraría go-dsl para validación real
	// Por ahora simulamos validación
	response := map[string]interface{}{
		"valid":   true,
		"message": "Código DSL válido",
		"errors":  []string{},
	}

	log.Printf("[INFO] DSLHandler.ValidateDSL: Validación completada exitosamente")
	return c.JSON(models.NewSuccessResponse(response))
}

// TestDSL prueba una plantilla DSL con datos de ejemplo
// @Summary Probar plantilla DSL
// @Description Ejecuta una plantilla DSL con datos de prueba y retorna el resultado
// @Tags DSL
// @Accept json
// @Produce json
// @Param test body models.DSLTestRequest true "Datos de prueba"
// @Success 200 {object} models.StandardResponse{data=models.DSLTestResponse} "Resultado de prueba"
// @Failure 400 {object} models.StandardResponse "Datos inválidos"
// @Failure 500 {object} models.StandardResponse "Error interno del servidor"
// @Router /api/v1/dsl/test [post]
func (h *DSLHandler) TestDSL(c *fiber.Ctx) error {
	log.Printf("[INFO] DSLHandler.TestDSL: Iniciando prueba de plantilla DSL")
	var request struct {
		TemplateID string                 `json:"templateId" binding:"required"`
		TestData   map[string]interface{} `json:"testData" binding:"required"`
	}

	if err := c.BodyParser(&request); err != nil {
		log.Printf("[ERROR] DSLHandler.TestDSL: Error parseando request - %v", err)
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("INVALID_JSON", "Formato JSON inválido"))
	}

	// TODO: Aquí se integraría go-dsl para ejecución real
	// Por ahora simulamos resultado
	response := map[string]interface{}{
		"success": true,
		"result": map[string]interface{}{
			"voucher_type": "SALE_INVOICE",
			"lines": []map[string]interface{}{
				{
					"account":     "4135",
					"description": "Venta de productos",
					"debit":       0,
					"credit":      100000,
				},
				{
					"account":     "2408",
					"description": "IVA 19%",
					"debit":       0,
					"credit":      19000,
				},
				{
					"account":     "130505",
					"description": "Cliente ABC S.A.S",
					"debit":       119000,
					"credit":      0,
				},
			},
		},
		"logs": []string{
			"Iniciando procesamiento de plantilla",
			"Evaluando reglas de negocio",
			"Generando líneas contables",
			"Validando balance",
			"Procesamiento completado exitosamente",
		},
	}

	log.Printf("[INFO] DSLHandler.TestDSL: Prueba de plantilla %s completada exitosamente", request.TemplateID)
	return c.JSON(models.NewSuccessResponse(response))
}

// GetIVARate obtiene la tasa de IVA actual
// @Summary Obtener tasa de IVA
// @Description Retorna la tasa de IVA configurada en el DSL
// @Tags DSL
// @Accept json
// @Produce json
// @Success 200 {object} models.StandardResponse
// @Router /api/v1/dsl/iva-rate [get]
func (h *DSLHandler) GetIVARate(c *fiber.Ctx) error {
	log.Printf("[INFO] DSLHandler.GetIVARate: Obteniendo tasa de IVA actual")
	if h.dslEngine == nil {
		log.Printf("[ERROR] DSLHandler.GetIVARate: Motor DSL no inicializado")
		return c.Status(http.StatusInternalServerError).JSON(
			models.NewErrorResponse("DSL_ENGINE_ERROR", "Motor DSL no inicializado"))
	}
	
	rate := h.dslEngine.GetIVARate()
	log.Printf("[INFO] DSLHandler.GetIVARate: Tasa actual: %.2f%%", rate*100)
	
	return c.JSON(models.NewSuccessResponse(map[string]interface{}{
		"rate": rate,
		"percentage": rate * 100,
	}))
}

// SetIVARate establece una nueva tasa de IVA
// @Summary Cambiar tasa de IVA
// @Description Actualiza la tasa de IVA en el motor DSL
// @Tags DSL
// @Accept json
// @Produce json
// @Param body body models.IVARateRequest true "Nueva tasa de IVA"
// @Success 200 {object} models.StandardResponse
// @Failure 400 {object} models.StandardResponse
// @Router /api/v1/dsl/iva-rate [post]
func (h *DSLHandler) SetIVARate(c *fiber.Ctx) error {
	log.Printf("[INFO] DSLHandler.SetIVARate: Iniciando actualización de tasa de IVA")
	if h.dslEngine == nil {
		log.Printf("[ERROR] DSLHandler.SetIVARate: Motor DSL no inicializado")
		return c.Status(http.StatusInternalServerError).JSON(
			models.NewErrorResponse("DSL_ENGINE_ERROR", "Motor DSL no inicializado"))
	}
	
	type Request struct {
		Rate float64 `json:"rate" validate:"required,min=0,max=1"`
	}
	
	var req Request
	if err := c.BodyParser(&req); err != nil {
		log.Printf("[ERROR] DSLHandler.SetIVARate: Error parseando request - %v", err)
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("INVALID_BODY", "Error al procesar la solicitud"))
	}
	
	// Actualizar la tasa en el motor DSL
	old_rate := h.dslEngine.GetIVARate()
	h.dslEngine.SetIVARate(req.Rate)
	log.Printf("[INFO] DSLHandler.SetIVARate: Tasa actualizada de %.2f%% a %.2f%%", old_rate*100, req.Rate*100)
	
	return c.JSON(models.NewSuccessResponse(map[string]interface{}{
		"message": "Tasa de IVA actualizada exitosamente",
		"rate": req.Rate,
		"percentage": req.Rate * 100,
	}))
}

// RegisterRoutes registra las rutas del handler de DSL
func (h *DSLHandler) RegisterRoutes(router fiber.Router) {
	dsl := router.Group("/dsl")
	{
		dsl.Get("/templates", h.GetTemplates)
		dsl.Get("/templates/:id", h.GetTemplateByID)
		dsl.Post("/validate", h.ValidateDSL)
		dsl.Post("/test", h.TestDSL)
		
		// Rutas de configuración
		dsl.Get("/iva-rate", h.GetIVARate)
		dsl.Post("/iva-rate", h.SetIVARate)
	}
}