package handlers

import (
	"motor-contable-poc/internal/models"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// DSLHandler maneja las peticiones HTTP para plantillas DSL
type DSLHandler struct {
	db *gorm.DB
}

// NewDSLHandler crea una nueva instancia del handler
func NewDSLHandler(db *gorm.DB) *DSLHandler {
	return &DSLHandler{
		db: db,
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
	if id == "" {
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
	var request struct {
		Code string `json:"code" binding:"required"`
	}

	if err := c.BodyParser(&request); err != nil {
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
	var request struct {
		TemplateID string                 `json:"templateId" binding:"required"`
		TestData   map[string]interface{} `json:"testData" binding:"required"`
	}

	if err := c.BodyParser(&request); err != nil {
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

	return c.JSON(models.NewSuccessResponse(response))
}

// RegisterRoutes registra las rutas del handler de DSL
func (h *DSLHandler) RegisterRoutes(router fiber.Router) {
	dsl := router.Group("/dsl")
	{
		dsl.Get("/templates", h.GetTemplates)
		dsl.Get("/templates/:id", h.GetTemplateByID)
		dsl.Post("/validate", h.ValidateDSL)
		dsl.Post("/test", h.TestDSL)
	}
}