package handlers

import (
	"motor-contable-poc/internal/models"
	"motor-contable-poc/internal/services"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// OrganizationHandler maneja las peticiones HTTP para organizaciones
type OrganizationHandler struct {
	orgService *services.OrganizationService
}

// NewOrganizationHandler crea una nueva instancia del handler
func NewOrganizationHandler(db *gorm.DB) *OrganizationHandler {
	return &OrganizationHandler{
		orgService: services.NewOrganizationService(db),
	}
}

// GetCurrent obtiene la información de la organización actual
// @Summary Obtiene organización actual
// @Description Retorna la información completa de la organización actual del sistema
// @Tags Organización
// @Accept json
// @Produce json
// @Success 200 {object} models.StandardResponse{data=models.OrganizationDetail} "Información de la organización"
// @Failure 404 {object} models.StandardResponse "Organización no encontrada"
// @Failure 500 {object} models.StandardResponse "Error interno del servidor"
// @Router /api/v1/organization/current [get]
func (h *OrganizationHandler) GetCurrent(c *fiber.Ctx) error {
	// TODO: En el futuro, aquí se usaría go-dsl para determinar la organización
	// actual basada en el contexto del usuario autenticado y reglas de tenant
	
	org, err := h.orgService.GetCurrent()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(
				models.NewErrorResponse("ORG_NOT_FOUND", "Organización no encontrada"))
		}
		return c.Status(http.StatusInternalServerError).JSON(
			models.NewErrorResponse("INTERNAL_ERROR", "Error interno del servidor"))
	}
	
	return c.JSON(models.NewSuccessResponse(org))
}

// GetDashboard obtiene datos del dashboard de la organización
// @Summary Dashboard de organización
// @Description Retorna métricas y KPIs principales de la organización
// @Tags Organización
// @Accept json
// @Produce json
// @Success 200 {object} models.StandardResponse{data=map[string]interface{}} "Datos del dashboard"
// @Failure 404 {object} models.StandardResponse "Organización no encontrada"
// @Failure 500 {object} models.StandardResponse "Error interno del servidor"
// @Router /api/v1/organization/dashboard [get]
func (h *OrganizationHandler) GetDashboard(c *fiber.Ctx) error {
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
	
	// TODO: En el futuro, se usaría go-dsl para generar dinámicamente
	// el dashboard basado en configuraciones personalizadas del usuario
	
	dashboard, err := h.orgService.GetDashboardData(org.ID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(
			models.NewErrorResponse("INTERNAL_ERROR", "Error obteniendo datos del dashboard"))
	}
	
	return c.JSON(models.NewSuccessResponse(dashboard))
}

// Update actualiza la configuración de la organización
// @Summary Actualiza organización
// @Description Actualiza la configuración de la organización actual
// @Tags Organización
// @Accept json
// @Produce json
// @Param organization body map[string]interface{} true "Datos de la organización a actualizar"
// @Success 200 {object} models.StandardResponse{data=models.OrganizationDetail} "Organización actualizada"
// @Failure 400 {object} models.StandardResponse "Datos inválidos"
// @Failure 404 {object} models.StandardResponse "Organización no encontrada"
// @Failure 500 {object} models.StandardResponse "Error interno del servidor"
// @Router /api/v1/organization/current [put]
func (h *OrganizationHandler) Update(c *fiber.Ctx) error {
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
	
	var updates map[string]interface{}
	if err := c.BodyParser(&updates); err != nil {
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("INVALID_JSON", "Formato JSON inválido"))
	}
	
	// TODO: Validar permisos del usuario para actualizar la organización
	// usando reglas DSL de autorización
	
	err = h.orgService.UpdateConfiguration(org.ID, updates)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(
				models.NewErrorResponse("ORG_NOT_FOUND", "Organización no encontrada"))
		}
		return c.Status(http.StatusInternalServerError).JSON(
			models.NewErrorResponse("UPDATE_ERROR", "Error actualizando la organización"))
	}
	
	// Obtener la organización actualizada
	updatedOrg, err := h.orgService.GetByID(org.ID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(
			models.NewErrorResponse("INTERNAL_ERROR", "Error obteniendo organización actualizada"))
	}
	
	return c.JSON(models.NewSuccessResponse(updatedOrg))
}

// ValidateBusinessRules valida las reglas de negocio de la organización
// @Summary Valida reglas de negocio
// @Description Ejecuta validaciones de reglas contables y de negocio sobre la organización
// @Tags Organización
// @Accept json
// @Produce json
// @Success 200 {object} models.StandardResponse{data=map[string]interface{}} "Resultado de validaciones"
// @Failure 404 {object} models.StandardResponse "Organización no encontrada"
// @Failure 500 {object} models.StandardResponse "Error interno del servidor"
// @Router /api/v1/organization/validate [post]
func (h *OrganizationHandler) ValidateBusinessRules(c *fiber.Ctx) error {
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
	
	// TODO: En el futuro, este endpoint usaría go-dsl para ejecutar
	// validaciones complejas configurables según las reglas de negocio
	// específicas de cada organización y tipo de empresa
	
	violations, err := h.orgService.ValidateBusinessRules(org.ID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(
			models.NewErrorResponse("VALIDATION_ERROR", "Error ejecutando validaciones"))
	}
	
	result := map[string]interface{}{
		"is_valid":    len(violations) == 0,
		"violations":  violations,
		"total_errors": len(violations),
		"validated_at": "2024-07-24T10:00:00Z",
	}
	
	return c.JSON(models.NewSuccessResponse(result))
}

// RegisterRoutes registra las rutas del handler de organizaciones
func (h *OrganizationHandler) RegisterRoutes(router fiber.Router) {
	org := router.Group("/organization")
	{
		org.Get("/current", h.GetCurrent)
		org.Put("/current", h.Update)
		org.Get("/dashboard", h.GetDashboard)
		org.Post("/validate", h.ValidateBusinessRules)
	}
}