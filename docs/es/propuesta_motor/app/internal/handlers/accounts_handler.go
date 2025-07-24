package handlers

import (
	"motor-contable-poc/internal/models"
	"motor-contable-poc/internal/services"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// AccountsHandler maneja las peticiones HTTP para cuentas contables
type AccountsHandler struct {
	accountService *services.AccountService
	orgService     *services.OrganizationService
}

// NewAccountsHandler crea una nueva instancia del handler
func NewAccountsHandler(db *gorm.DB) *AccountsHandler {
	return &AccountsHandler{
		accountService: services.NewAccountService(db),
		orgService:     services.NewOrganizationService(db),
	}
}

// GetTree obtiene el árbol de cuentas contables
// @Summary Árbol de cuentas
// @Description Retorna la estructura jerárquica del plan de cuentas
// @Tags Cuentas
// @Accept json
// @Produce json
// @Success 200 {object} models.StandardResponse{data=models.AccountTreeResponse} "Árbol de cuentas"
// @Failure 500 {object} models.StandardResponse "Error interno del servidor"
// @Router /api/v1/accounts/tree [get]
func (h *AccountsHandler) GetTree(c *fiber.Ctx) error {
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

	// Obtener el árbol de cuentas
	tree, err := h.accountService.GetTree(org.ID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(
			models.NewErrorResponse("INTERNAL_ERROR", "Error obteniendo árbol de cuentas"))
	}

	response := models.AccountTreeResponse{
		Accounts: tree,
	}

	return c.JSON(models.NewSuccessResponse(response))
}

// GetList obtiene una lista paginada de cuentas
// @Summary Lista de cuentas
// @Description Retorna una lista paginada de cuentas contables
// @Tags Cuentas
// @Accept json
// @Produce json
// @Param page query int false "Número de página" default(1)
// @Param limit query int false "Elementos por página" default(20)
// @Param type query string false "Filtrar por tipo de cuenta"
// @Param active query bool false "Filtrar por estado activo/inactivo"
// @Success 200 {object} models.StandardResponse{data=models.AccountsListResponse} "Lista de cuentas"
// @Failure 400 {object} models.StandardResponse "Parámetros inválidos"
// @Failure 500 {object} models.StandardResponse "Error interno del servidor"
// @Router /api/v1/accounts [get]
func (h *AccountsHandler) GetList(c *fiber.Ctx) error {
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
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Filtros opcionales
	filters := models.AccountFilters{
		AccountType: c.Query("type"),
	}

	if activeStr := c.Query("active"); activeStr != "" {
		if active, err := strconv.ParseBool(activeStr); err == nil {
			filters.IsActive = &active
		}
	}

	// Obtener lista de cuentas
	response, err := h.accountService.GetList(org.ID, page, limit, filters)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(
			models.NewErrorResponse("INTERNAL_ERROR", "Error obteniendo lista de cuentas"))
	}

	return c.JSON(models.NewSuccessResponse(response))
}

// GetByCode obtiene una cuenta específica por código
// @Summary Detalle de cuenta por código
// @Description Retorna el detalle completo de una cuenta específica por su código
// @Tags Cuentas
// @Accept json
// @Produce json
// @Param code path string true "Código de la cuenta"
// @Success 200 {object} models.StandardResponse{data=models.AccountDetail} "Detalle de la cuenta"
// @Failure 404 {object} models.StandardResponse "Cuenta no encontrada"
// @Failure 500 {object} models.StandardResponse "Error interno del servidor"
// @Router /api/v1/accounts/{code} [get]
func (h *AccountsHandler) GetByCode(c *fiber.Ctx) error {
	code := c.Params("code")
	if code == "" {
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("MISSING_CODE", "Código de cuenta requerido"))
	}

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

	// Obtener la cuenta por código
	account, err := h.accountService.GetByCode(org.ID, code)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(
				models.NewErrorResponse("ACCOUNT_NOT_FOUND", "Cuenta no encontrada"))
		}
		return c.Status(http.StatusInternalServerError).JSON(
			models.NewErrorResponse("INTERNAL_ERROR", "Error obteniendo cuenta"))
	}

	// Convertir a detalle
	detail, err := account.ToDetail()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(
			models.NewErrorResponse("INTERNAL_ERROR", "Error convirtiendo detalle de cuenta"))
	}

	return c.JSON(models.NewSuccessResponse(detail))
}

// GetTypes obtiene los tipos de cuenta disponibles
// @Summary Tipos de cuenta
// @Description Retorna los tipos de cuenta contable disponibles en el sistema
// @Tags Cuentas
// @Accept json
// @Produce json
// @Success 200 {object} models.StandardResponse{data=[]models.AccountType} "Tipos de cuenta"
// @Failure 500 {object} models.StandardResponse "Error interno del servidor"
// @Router /api/v1/accounts/types [get]
func (h *AccountsHandler) GetTypes(c *fiber.Ctx) error {
	types := []models.AccountType{
		{Code: "ASSET", Name: "Activo", Nature: "D", Description: "Bienes y derechos de la empresa"},
		{Code: "LIABILITY", Name: "Pasivo", Nature: "C", Description: "Obligaciones y deudas de la empresa"},
		{Code: "EQUITY", Name: "Patrimonio", Nature: "C", Description: "Capital y reservas de la empresa"},
		{Code: "INCOME", Name: "Ingreso", Nature: "C", Description: "Ingresos y ventas de la empresa"},
		{Code: "EXPENSE", Name: "Gasto", Nature: "D", Description: "Gastos y costos de la empresa"},
	}

	return c.JSON(models.NewSuccessResponse(types))
}

// Create crea una nueva cuenta contable
// @Summary Crear cuenta
// @Description Crea una nueva cuenta contable en el plan de cuentas
// @Tags Cuentas
// @Accept json
// @Produce json
// @Param account body models.AccountCreateRequest true "Datos de la cuenta"
// @Success 201 {object} models.StandardResponse{data=models.Account} "Cuenta creada"
// @Failure 400 {object} models.StandardResponse "Datos inválidos"
// @Failure 500 {object} models.StandardResponse "Error interno del servidor"
// @Router /api/v1/accounts [post]
func (h *AccountsHandler) Create(c *fiber.Ctx) error {
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

	var request models.AccountCreateRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("INVALID_JSON", "Formato JSON inválido"))
	}

	// Crear la cuenta
	account, err := h.accountService.Create(org.ID, request)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("CREATION_ERROR", err.Error()))
	}

	return c.Status(http.StatusCreated).JSON(models.NewSuccessResponse(account))
}

// Update actualiza una cuenta contable
// @Summary Actualizar cuenta
// @Description Actualiza los datos de una cuenta contable existente
// @Tags Cuentas
// @Accept json
// @Produce json
// @Param id path string true "ID de la cuenta"
// @Param account body models.AccountUpdateRequest true "Datos a actualizar"
// @Success 200 {object} models.StandardResponse{data=models.Account} "Cuenta actualizada"
// @Failure 400 {object} models.StandardResponse "Datos inválidos"
// @Failure 404 {object} models.StandardResponse "Cuenta no encontrada"
// @Failure 500 {object} models.StandardResponse "Error interno del servidor"
// @Router /api/v1/accounts/{id} [put]
func (h *AccountsHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("MISSING_ID", "ID de cuenta requerido"))
	}

	var request models.AccountUpdateRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("INVALID_JSON", "Formato JSON inválido"))
	}

	// Actualizar la cuenta
	account, err := h.accountService.Update(id, request)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(
				models.NewErrorResponse("ACCOUNT_NOT_FOUND", "Cuenta no encontrada"))
		}
		return c.Status(http.StatusBadRequest).JSON(
			models.NewErrorResponse("UPDATE_ERROR", err.Error()))
	}

	return c.JSON(models.NewSuccessResponse(account))
}

// RegisterRoutes registra las rutas del handler de cuentas
func (h *AccountsHandler) RegisterRoutes(router fiber.Router) {
	accounts := router.Group("/accounts")
	{
		accounts.Get("/tree", h.GetTree)
		accounts.Get("/types", h.GetTypes)
		accounts.Get("/", h.GetList)
		accounts.Post("/", h.Create)
		accounts.Get("/:code", h.GetByCode)
		accounts.Put("/:id", h.Update)
	}
}