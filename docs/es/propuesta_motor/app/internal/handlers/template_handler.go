package handlers

import (
	"motor-contable-poc/internal/models"
	"motor-contable-poc/internal/repository"
	"motor-contable-poc/internal/services"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type TemplateHandler struct {
	templateService *services.TemplateService
}

func NewTemplateHandler(db *gorm.DB) *TemplateHandler {
	// Create repositories
	templateRepo := repository.NewTemplateRepository(db)
	
	// Create services - using simplified versions for POC
	journalService := services.NewJournalEntryService(db)
	accountService := services.NewAccountService(db)
	templateService := services.NewTemplateService(templateRepo, journalService, accountService)
	
	return &TemplateHandler{
		templateService: templateService,
	}
}

// RegisterRoutes registers all template routes
func (h *TemplateHandler) RegisterRoutes(api fiber.Router) {
	templates := api.Group("/templates")
	
	templates.Get("/", h.GetTemplates)
	templates.Get("/:id", h.GetTemplate)
	templates.Post("/", h.CreateTemplate)
	templates.Put("/:id", h.UpdateTemplate)
	templates.Delete("/:id", h.DeleteTemplate)
	templates.Post("/:id/preview", h.PreviewTemplate)
	templates.Post("/:id/execute", h.ExecuteTemplate)
	templates.Get("/:id/history", h.GetTemplateHistory)
}

// GetTemplates returns all templates
func (h *TemplateHandler) GetTemplates(c *fiber.Ctx) error {
	// For POC, return mock templates
	templates := []models.JournalTemplate{
		{
			ID:          "tpl-001",
			Name:        "Nómina Mensual",
			Description: "Template para registro de nómina mensual con prestaciones básicas",
			DSLCode: `template payroll_monthly
  params ($total_salaries, $period)
  
  entry
    description: "Nómina mensual - " + $period
    date: last_day($period)
    
    line debit account("510506") amount($total_salaries) 
         description("Sueldos y salarios")
    
    line credit account("250505") amount($total_salaries * 0.8)
         description("Salarios por pagar")`,
			Parameters: []models.TemplateParameter{
				{Name: "total_salaries", Type: "number", Required: true, Description: "Total de salarios base"},
				{Name: "period", Type: "string", Required: true, Description: "Período (YYYY-MM)"},
			},
			IsActive: true,
		},
		{
			ID:          "tpl-002",
			Name:        "Depreciación Mensual",
			Description: "Cálculo y registro de depreciación mensual de activos fijos",
			DSLCode: `template depreciation_monthly
  params ($asset_value, $monthly_rate, $asset_description, $period)
  
  entry
    description: "Depreciación " + $asset_description + " - " + $period
    date: last_day($period)
    
    line debit account("516005") amount($asset_value * $monthly_rate)
         description("Gasto depreciación")
    
    line credit account("159205") amount($asset_value * $monthly_rate)
         description("Depreciación acumulada")`,
			Parameters: []models.TemplateParameter{
				{Name: "asset_value", Type: "number", Required: true, Description: "Valor del activo"},
				{Name: "monthly_rate", Type: "number", Required: true, Description: "Tasa mensual"},
				{Name: "asset_description", Type: "string", Required: true, Description: "Descripción del activo"},
				{Name: "period", Type: "string", Required: true, Description: "Período (YYYY-MM)"},
			},
			IsActive: true,
		},
		{
			ID:          "tpl-003",
			Name:        "Factura de Venta Recurrente",
			Description: "Template para facturas de venta que se repiten mensualmente",
			DSLCode: `template recurring_invoice
  params ($customer_name, $service_amount, $invoice_date)
  
  entry
    description: "Factura venta - " + $customer_name
    date: $invoice_date
    
    line debit account("130505") amount($service_amount * 1.19)
         description("CxC " + $customer_name)
    
    line credit account("413535") amount($service_amount)
         description("Ingreso servicios")
    
    line credit account("240801") amount($service_amount * 0.19)
         description("IVA generado 19%")`,
			Parameters: []models.TemplateParameter{
				{Name: "customer_name", Type: "string", Required: true, Description: "Nombre del cliente"},
				{Name: "service_amount", Type: "number", Required: true, Description: "Valor del servicio (sin IVA)"},
				{Name: "invoice_date", Type: "date", Required: true, Description: "Fecha de la factura"},
			},
			IsActive: true,
		},
	}
	
	return c.JSON(models.TemplateListResponse{
		Templates: templates,
		Total:     len(templates),
	})
}

// GetTemplate returns a specific template
func (h *TemplateHandler) GetTemplate(c *fiber.Ctx) error {
	id := c.Params("id")
	
	// Mock data for POC
	templates := map[string]models.JournalTemplate{
		"tpl-001": {
			ID:          "tpl-001",
			Name:        "Nómina Mensual",
			Description: "Template para registro de nómina mensual con prestaciones básicas",
			DSLCode: `template payroll_monthly
  params ($total_salaries, $period)
  
  entry
    description: "Nómina mensual - " + $period
    date: last_day($period)
    
    line debit account("510506") amount($total_salaries) 
         description("Sueldos y salarios")
    
    line credit account("250505") amount($total_salaries * 0.8)
         description("Salarios por pagar")`,
			Parameters: []models.TemplateParameter{
				{Name: "total_salaries", Type: "number", Required: true, Description: "Total de salarios base"},
				{Name: "period", Type: "string", Required: true, Description: "Período (YYYY-MM)"},
			},
			IsActive: true,
		},
	}
	
	template, exists := templates[id]
	if !exists {
		return c.Status(404).JSON(fiber.Map{
			"error": "Template not found",
		})
	}
	
	return c.JSON(template)
}

// CreateTemplate creates a new template
func (h *TemplateHandler) CreateTemplate(c *fiber.Ctx) error {
	var template models.JournalTemplate
	if err := c.BodyParser(&template); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	
	// Validate required fields
	if template.Name == "" || template.DSLCode == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Name and DSL code are required",
		})
	}
	
	// Set organization ID (in real app from auth context)
	template.OrganizationID = "org123"
	
	// Create template
	userID := "user123" // In real app from auth context
	if err := h.templateService.Create(&template, userID); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Error creating template",
		})
	}
	
	return c.Status(201).JSON(template)
}

// UpdateTemplate updates an existing template
func (h *TemplateHandler) UpdateTemplate(c *fiber.Ctx) error {
	id := c.Params("id")
	
	// Get existing template
	existing, err := h.templateService.GetByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Template not found",
		})
	}
	
	// Parse update data
	var update models.JournalTemplate
	if err := c.BodyParser(&update); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	
	// Update fields
	existing.Name = update.Name
	existing.Description = update.Description
	existing.DSLCode = update.DSLCode
	existing.Parameters = update.Parameters
	existing.IsActive = update.IsActive
	
	// Save changes
	if err := h.templateService.Update(existing); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Error updating template",
		})
	}
	
	return c.JSON(existing)
}

// DeleteTemplate deletes a template
func (h *TemplateHandler) DeleteTemplate(c *fiber.Ctx) error {
	id := c.Params("id")
	
	if err := h.templateService.Delete(id); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Error deleting template",
		})
	}
	
	return c.JSON(fiber.Map{
		"message": "Template deleted successfully",
	})
}

// PreviewTemplate executes a template in preview mode
func (h *TemplateHandler) PreviewTemplate(c *fiber.Ctx) error {
	id := c.Params("id")
	
	var request models.TemplatePreviewRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	
	// Execute preview
	entry, err := h.templateService.PreviewTemplate(id, request.Parameters)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	// Build response
	response := models.TemplatePreviewResponse{
		Preview: entry,
	}
	response.Summary.TotalDebit = entry.TotalDebit
	response.Summary.TotalCredit = entry.TotalCredit
	response.Summary.IsBalanced = entry.IsBalanced()
	response.Summary.LinesCount = len(entry.JournalLines)
	
	return c.JSON(response)
}

// ExecuteTemplate executes a template and creates a journal entry
func (h *TemplateHandler) ExecuteTemplate(c *fiber.Ctx) error {
	id := c.Params("id")
	
	var request models.TemplateExecuteRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	
	// If dry run, just preview
	if request.DryRun {
		entry, err := h.templateService.PreviewTemplate(id, request.Parameters)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.JSON(fiber.Map{
			"dry_run": true,
			"entry":   entry,
		})
	}
	
	// Execute template
	userID := "user123" // In real app from auth context
	entry, err := h.templateService.ExecuteTemplate(id, request.Parameters, userID)
	if err != nil {
		return c.Status(400).JSON(models.TemplateExecuteResponse{
			Success: false,
			Error:   err.Error(),
		})
	}
	
	return c.JSON(models.TemplateExecuteResponse{
		Success:     true,
		EntryID:     entry.ID,
		EntryNumber: entry.EntryNumber,
	})
}

// GetTemplateHistory returns execution history for a template
func (h *TemplateHandler) GetTemplateHistory(c *fiber.Ctx) error {
	id := c.Params("id")
	limit := c.QueryInt("limit", 10)
	
	history, err := h.templateService.GetExecutionHistory(id, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Error fetching history",
		})
	}
	
	return c.JSON(fiber.Map{
		"history": history,
		"total":   len(history),
	})
}