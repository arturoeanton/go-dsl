package handlers

import (
	"motor-contable-poc/internal/models"
	"time"
	
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type TemplateSimpleHandler struct {
	db *gorm.DB
}

func NewTemplateSimpleHandler(db *gorm.DB) *TemplateSimpleHandler {
	return &TemplateSimpleHandler{
		db: db,
	}
}

// RegisterRoutes registers all template routes
func (h *TemplateSimpleHandler) RegisterRoutes(api fiber.Router) {
	templates := api.Group("/templates")
	
	templates.Get("/", h.GetTemplates)
	templates.Get("/test", h.TestEndpoint)
	templates.Get("/:id", h.GetTemplate)
	templates.Post("/", h.CreateTemplate)
	templates.Put("/:id", h.UpdateTemplate)
	templates.Delete("/:id", h.DeleteTemplate)
}

// TestEndpoint for debugging
func (h *TemplateSimpleHandler) TestEndpoint(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Templates endpoint is working",
		"time": time.Now().Format(time.RFC3339),
	})
}

// GetTemplates returns all templates from SQLite
func (h *TemplateSimpleHandler) GetTemplates(c *fiber.Ctx) error {
	var templates []models.Template
	
	// Get company ID from context (mock for now)
	companyID := "DEMO-CO"
	
	// Log the query
	println("GetTemplates called - Looking for templates with company_id:", companyID)
	
	// Query templates from database
	result := h.db.Where("company_id = ? AND is_active = ?", companyID, true).
		Order("created_at DESC").
		Find(&templates)
	
	if result.Error != nil {
		println("Database error:", result.Error.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error fetching templates",
			"details": result.Error.Error(),
		})
	}
	
	println("Found", len(templates), "templates")
	
	// If no templates found, create a simple response
	if len(templates) == 0 {
		println("No templates found, returning empty array")
		return c.JSON(fiber.Map{
			"success":   true,
			"templates": []models.Template{},
			"total":     0,
		})
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"templates": templates,
		"total":     len(templates),
	})
}

// GetTemplate returns a specific template
func (h *TemplateSimpleHandler) GetTemplate(c *fiber.Ctx) error {
	id := c.Params("id")
	
	var template models.Template
	result := h.db.First(&template, "id = ?", id)
	
	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Template not found",
		})
	}
	
	return c.JSON(template)
}

// CreateTemplate creates a new template
func (h *TemplateSimpleHandler) CreateTemplate(c *fiber.Ctx) error {
	var template models.Template
	
	if err := c.BodyParser(&template); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	
	// Validate required fields
	if template.Name == "" || template.DSLContent == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Name and DSL content are required",
		})
	}
	
	// Set defaults
	template.CompanyID = "DEMO-CO"
	template.Status = "active"
	template.IsActive = true
	template.CreatedBy = "user123"
	
	// Create in database
	result := h.db.Create(&template)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Error creating template",
		})
	}
	
	return c.Status(201).JSON(template)
}

// UpdateTemplate updates an existing template
func (h *TemplateSimpleHandler) UpdateTemplate(c *fiber.Ctx) error {
	id := c.Params("id")
	
	// Get existing template
	var existing models.Template
	result := h.db.First(&existing, "id = ?", id)
	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Template not found",
		})
	}
	
	// Parse update data
	var update models.Template
	if err := c.BodyParser(&update); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	
	// Update fields
	existing.Name = update.Name
	existing.Description = update.Description
	existing.DSLContent = update.DSLContent
	existing.Parameters = update.Parameters
	existing.Status = update.Status
	existing.Type = update.Type
	existing.IsActive = update.IsActive
	
	// Save changes
	result = h.db.Save(&existing)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Error updating template",
		})
	}
	
	return c.JSON(existing)
}

// DeleteTemplate deletes a template (soft delete by setting status to inactive)
func (h *TemplateSimpleHandler) DeleteTemplate(c *fiber.Ctx) error {
	id := c.Params("id")
	
	// Update status to inactive
	result := h.db.Model(&models.Template{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status": "inactive",
			"is_active": false,
		})
	
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Error deleting template",
		})
	}
	
	if result.RowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{
			"error": "Template not found",
		})
	}
	
	return c.JSON(fiber.Map{
		"message": "Template deleted successfully",
	})
}