package handlers

import (
	"motor-contable-poc/internal/models"
	"motor-contable-poc/internal/services"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// JournalEntriesHandler maneja las peticiones HTTP para asientos contables
type JournalEntriesHandler struct {
	journalService *services.JournalEntryService
	orgService     *services.OrganizationService
}

// NewJournalEntriesHandler crea una nueva instancia del handler
func NewJournalEntriesHandler(db *gorm.DB) *JournalEntriesHandler {
	return &JournalEntriesHandler{
		journalService: services.NewJournalEntryService(db),
		orgService:     services.NewOrganizationService(db),
	}
}

// GetList obtiene una lista paginada de asientos contables
// @Summary Lista de asientos contables
// @Description Retorna una lista paginada de asientos contables con sus líneas
// @Tags Asientos Contables
// @Accept json
// @Produce json
// @Param page query int false "Número de página" default(1)
// @Param per_page query int false "Elementos por página" default(20)
// @Success 200 {object} models.StandardResponse{data=models.JournalEntriesListResponse} "Lista de asientos"
// @Failure 400 {object} models.StandardResponse "Parámetros inválidos"
// @Failure 500 {object} models.StandardResponse "Error interno del servidor"
// @Router /api/v1/journal-entries [get]
func (h *JournalEntriesHandler) GetList(c *fiber.Ctx) error {
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

	// Obtener parámetros de paginación
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	// Obtener lista de asientos
	response, err := h.journalService.GetList(org.ID, page, perPage)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(
			models.NewErrorResponse("INTERNAL_ERROR", "Error obteniendo lista de asientos"))
	}

	return c.JSON(models.NewSuccessResponse(response))
}

// Create crea un nuevo asiento contable
// @Summary Crear asiento contable
// @Description Crea un nuevo asiento contable manual con sus líneas
// @Tags Asientos Contables
// @Accept json
// @Produce json
// @Param entry body models.JournalEntryCreateRequest true "Datos del asiento"
// @Success 201 {object} models.StandardResponse{data=models.JournalEntry} "Asiento creado"
// @Failure 400 {object} models.StandardResponse "Datos inválidos"
// @Failure 500 {object} models.StandardResponse "Error interno del servidor"
// @Router /api/v1/journal-entries [post]
func (h *JournalEntriesHandler) Create(c *fiber.Ctx) error {
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

	var request models.JournalEntryCreateRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("INVALID_JSON", "Formato JSON inválido"))
	}

	// Crear el asiento
	entry, err := h.journalService.Create(org.ID, request)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("CREATION_ERROR", err.Error()))
	}

	return c.Status(http.StatusCreated).JSON(models.NewSuccessResponse(entry))
}

// GetByID obtiene un asiento contable específico
// @Summary Detalle de asiento contable
// @Description Retorna el detalle completo de un asiento contable específico
// @Tags Asientos Contables
// @Accept json
// @Produce json
// @Param id path string true "ID del asiento"
// @Success 200 {object} models.StandardResponse{data=models.JournalEntryDetail} "Detalle del asiento"
// @Failure 404 {object} models.StandardResponse "Asiento no encontrado"
// @Failure 500 {object} models.StandardResponse "Error interno del servidor"
// @Router /api/v1/journal-entries/{id} [get]
func (h *JournalEntriesHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("MISSING_ID", "ID del asiento requerido"))
	}

	detail, err := h.journalService.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(
				models.NewErrorResponse("ENTRY_NOT_FOUND", "Asiento no encontrado"))
		}
		return c.Status(http.StatusInternalServerError).JSON(
			models.NewErrorResponse("INTERNAL_ERROR", "Error obteniendo asiento"))
	}

	return c.JSON(models.NewSuccessResponse(detail))
}

// Post contabiliza un asiento contable
// @Summary Contabilizar asiento
// @Description Cambia el estado del asiento a contabilizado
// @Tags Asientos Contables
// @Accept json
// @Produce json
// @Param id path string true "ID del asiento"
// @Success 200 {object} models.StandardResponse "Asiento contabilizado exitosamente"
// @Failure 400 {object} models.StandardResponse "No se puede contabilizar"
// @Failure 404 {object} models.StandardResponse "Asiento no encontrado"
// @Failure 500 {object} models.StandardResponse "Error interno del servidor"
// @Router /api/v1/journal-entries/{id}/post [post]
func (h *JournalEntriesHandler) Post(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("MISSING_ID", "ID del asiento requerido"))
	}

	// TODO: Obtener userID del contexto de usuario autenticado
	userID := "system-user" // Placeholder para el POC

	err := h.journalService.Post(id, userID)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("POST_ERROR", err.Error()))
	}

	return c.JSON(models.NewSuccessResponse(map[string]interface{}{
		"message":   "Asiento contabilizado exitosamente",
		"entry_id":  id,
		"posted_at": models.JSONTime{},
	}))
}

// Reverse reversa un asiento contable
// @Summary Reversar asiento
// @Description Crea un asiento de reversión para un asiento ya contabilizado
// @Tags Asientos Contables
// @Accept json
// @Produce json
// @Param id path string true "ID del asiento"
// @Success 200 {object} models.StandardResponse{data=models.JournalEntry} "Asiento de reversión creado"
// @Failure 400 {object} models.StandardResponse "No se puede reversar"
// @Failure 404 {object} models.StandardResponse "Asiento no encontrado"
// @Failure 500 {object} models.StandardResponse "Error interno del servidor"
// @Router /api/v1/journal-entries/{id}/reverse [post]
func (h *JournalEntriesHandler) Reverse(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("MISSING_ID", "ID del asiento requerido"))
	}

	// TODO: Obtener userID del contexto de usuario autenticado
	userID := "system-user" // Placeholder para el POC

	reverseEntry, err := h.journalService.Reverse(id, userID)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("REVERSE_ERROR", err.Error()))
	}

	return c.JSON(models.NewSuccessResponse(reverseEntry))
}

// RegisterRoutes registra las rutas del handler de asientos contables
func (h *JournalEntriesHandler) RegisterRoutes(router fiber.Router) {
	entries := router.Group("/journal-entries")
	{
		entries.Get("/", h.GetList)
		entries.Post("/", h.Create)
		entries.Get("/:id", h.GetByID)
		entries.Post("/:id/post", h.Post)
		entries.Post("/:id/reverse", h.Reverse)
	}
}