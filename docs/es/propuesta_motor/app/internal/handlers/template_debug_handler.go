package handlers

import (
	"github.com/gofiber/fiber/v2"
)

type TemplateDebugHandler struct{}

func NewTemplateDebugHandler() *TemplateDebugHandler {
	return &TemplateDebugHandler{}
}

// RegisterRoutes registers debug routes
func (h *TemplateDebugHandler) RegisterRoutes(api fiber.Router) {
	api.Get("/templates-debug", h.GetTemplatesDebug)
}

// GetTemplatesDebug returns hardcoded templates for testing
func (h *TemplateDebugHandler) GetTemplatesDebug(c *fiber.Ctx) error {
	templates := []map[string]interface{}{
		{
			"id":          "tpl_001",
			"name":        "Factura de Venta Estándar",
			"description": "Genera asiento contable para facturas de venta con IVA 19%",
			"type":        "invoice_sale",
			"status":      "active",
			"is_active":   true,
			"dsl_content": `# Template: Factura de Venta Estándar
template factura_venta_standard
  params (invoice_number, customer_name, base_amount, invoice_date)
  
  set tax_rate = 0.19
  set tax_amount = base_amount * tax_rate
  set total_amount = base_amount + tax_amount
  
  entry
    description: "Factura Venta " + invoice_number + " - " + customer_name
    date: invoice_date
    
    line debit account("130505") amount(total_amount)
         description("CxC Cliente " + customer_name)
    
    line credit account("413595") amount(base_amount)
         description("Venta de productos")
    
    line credit account("240802") amount(tax_amount)
         description("IVA generado 19%")`,
		},
		{
			"id":          "tpl_002",
			"name":        "Factura de Venta con Retención",
			"description": "Genera asiento para facturas con retención en la fuente",
			"type":        "invoice_sale",
			"status":      "active",
			"is_active":   true,
			"dsl_content": `# Template: Factura con Retención
template factura_retencion
  params (invoice_number, customer_name, base_amount)
  
  set tax_rate = 0.19
  set retention_rate = 0.025
  set tax_amount = base_amount * tax_rate
  set retention_amount = base_amount * retention_rate
  
  entry
    description: "Factura " + invoice_number + " con retención"
    
    line debit account("130505") amount(base_amount + tax_amount - retention_amount)
    line debit account("135515") amount(retention_amount)
    line credit account("413595") amount(base_amount)
    line credit account("240802") amount(tax_amount)`,
		},
		{
			"id":          "tpl_003",
			"name":        "Nómina Mensual Básica",
			"description": "Genera asiento contable para pago de nómina mensual",
			"type":        "payroll",
			"status":      "active",
			"is_active":   true,
			"dsl_content": `# Template: Nómina
template nomina_mensual
  params (employee_name, basic_salary, period)
  
  set health = basic_salary * 0.04
  set pension = basic_salary * 0.04
  set net_payment = basic_salary - health - pension
  
  entry
    description: "Nómina " + period + " - " + employee_name
    
    line debit account("510506") amount(basic_salary)
    line credit account("237005") amount(health)
    line credit account("238030") amount(pension)
    line credit account("250505") amount(net_payment)`,
		},
		{
			"id":          "tpl_004",
			"name":        "Compra con IVA",
			"description": "Registra compras de inventario o gastos con IVA",
			"type":        "invoice_purchase",
			"status":      "active",
			"is_active":   true,
			"dsl_content": `# Template: Compra con IVA
template compra_iva
  params (invoice_number, supplier_name, base_amount)
  
  set tax_amount = base_amount * 0.19
  set total = base_amount + tax_amount
  
  entry
    description: "Compra FC " + invoice_number
    
    line debit account("143505") amount(base_amount)
    line debit account("240801") amount(tax_amount)
    line credit account("220505") amount(total)`,
		},
		{
			"id":          "tpl_005",
			"name":        "Pago a Proveedor",
			"description": "Registra el pago a proveedores desde banco",
			"type":        "payment",
			"status":      "inactive",
			"is_active":   false,
			"dsl_content": `# Template: Pago a Proveedor
template pago_proveedor
  params (payment_number, supplier_name, amount)
  
  entry
    description: "Pago a " + supplier_name
    
    line debit account("220505") amount(amount)
    line credit account("111005") amount(amount)`,
		},
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"templates": templates,
		"total":     len(templates),
	})
}