/*
GUÍA COMPLETA DE IMPLEMENTACIÓN API EN FIBER
Motor Contable Cloud-Native - Fase 1

Esta guía proporciona una implementación completa de todas las APIs definidas
en el Swagger y los mocks JSON para el Motor Contable usando Go + Fiber.

INSTRUCCIONES DE USO:
1. Revisar la estructura del proyecto recomendada
2. Implementar cada endpoint siguiendo los ejemplos proporcionados
3. Usar los mocks JSON como referencia para las respuestas
4. Integrar con go-dsl para el procesamiento de comprobantes
5. Configurar PostgreSQL según el esquema proporcionado

TABLA DE CONTENIDOS:
- Estructura del Proyecto
- Configuración Inicial
- Modelos de Datos
- Middlewares
- Controladores por Módulo
- Integración con go-dsl
- Configuración de Base de Datos
- Ejemplos de Testing
*/

package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

/*
===========================
ESTRUCTURA DEL PROYECTO RECOMENDADA
===========================

motor-contable/
├── cmd/
│   └── server/
│       └── main.go                 # Punto de entrada de la aplicación
├── internal/
│   ├── config/
│   │   └── config.go              # Configuración de la aplicación
│   ├── database/
│   │   ├── connection.go          # Conexión a PostgreSQL
│   │   └── migrations/            # Migraciones de la base de datos
│   ├── models/                    # Modelos de datos (structs)
│   │   ├── organization.go
│   │   ├── voucher.go
│   │   ├── account.go
│   │   ├── journal_entry.go
│   │   ├── third_party.go
│   │   ├── dsl_template.go
│   │   └── common.go
│   ├── handlers/                  # Controladores HTTP
│   │   ├── health.go
│   │   ├── dashboard.go
│   │   ├── vouchers.go
│   │   ├── journal_entries.go
│   │   ├── accounts.go
│   │   ├── reports.go
│   │   ├── dsl.go
│   │   ├── third_parties.go
│   │   ├── catalogs.go
│   │   └── audit.go
│   ├── services/                  # Lógica de negocio
│   │   ├── voucher_service.go
│   │   ├── dsl_service.go
│   │   ├── account_service.go
│   │   └── report_service.go
│   ├── middleware/               # Middlewares personalizados
│   │   ├── auth.go
│   │   ├── tenant.go
│   │   └── validation.go
│   ├── utils/                   # Utilidades
│   │   ├── response.go
│   │   ├── validation.go
│   │   └── pagination.go
│   └── routes/                 # Definición de rutas
│       └── routes.go
├── pkg/                       # Paquetes reutilizables
│   ├── dslengine/            # Integración con go-dsl
│   │   ├── engine.go
│   │   ├── templates.go
│   │   └── processor.go
│   └── errors/               # Manejo de errores
│       └── errors.go
├── docs/                     # Documentación
├── scripts/                  # Scripts de deployment
├── docker-compose.yml        # Para desarrollo local
├── Dockerfile
├── go.mod
└── go.sum
*/

/*
===========================
CONFIGURACIÓN INICIAL
===========================
*/

// Config representa la configuración de la aplicación
type Config struct {
	Port        string
	DatabaseURL string
	Environment string
	DSLEnabled  bool
	LogLevel    string
	CORSOrigins []string
}

// LoadConfig carga la configuración desde variables de entorno
func LoadConfig() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/motor_contable?sslmode=disable"),
		Environment: getEnv("ENVIRONMENT", "development"),
		DSLEnabled:  getEnvBool("DSL_ENABLED", true),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		CORSOrigins: []string{getEnv("CORS_ORIGINS", "*")},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

/*
===========================
MODELOS DE DATOS
===========================
*/

// BaseModel contiene campos comunes para todos los modelos
type BaseModel struct {
	ID        string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// Organization representa una organización (tenant)
type Organization struct {
	BaseModel
	Code                    string                 `json:"code" gorm:"uniqueIndex;not null"`
	Name                    string                 `json:"name" gorm:"not null"`
	CommercialName          string                 `json:"commercial_name"`
	DocumentType            string                 `json:"document_type"`
	TaxID                   string                 `json:"tax_id" gorm:"uniqueIndex"`
	CountryCode             string                 `json:"country_code" gorm:"not null"`
	CurrencyDefault         string                 `json:"currency_default" gorm:"default:'COP'"`
	Language                string                 `json:"language" gorm:"default:'es'"`
	Timezone                string                 `json:"timezone" gorm:"default:'America/Bogota'"`
	ContactInfo             map[string]interface{} `json:"contact_info" gorm:"type:jsonb"`
	FiscalInfo              map[string]interface{} `json:"fiscal_info" gorm:"type:jsonb"`
	AccountingConfiguration map[string]interface{} `json:"accounting_configuration" gorm:"type:jsonb"`
	DSLConfiguration        map[string]interface{} `json:"dsl_configuration" gorm:"type:jsonb"`
	IsActive                bool                   `json:"is_active" gorm:"default:true"`
}

// Account representa una cuenta contable
type Account struct {
	BaseModel
	OrganizationID string                 `json:"organization_id" gorm:"type:uuid;not null;index"`
	Organization   Organization           `json:"-" gorm:"foreignKey:OrganizationID"`
	AccountCode    string                 `json:"account_code" gorm:"not null;index"`
	Name           string                 `json:"name" gorm:"not null"`
	Type           string                 `json:"type" gorm:"not null"`   // ASSET, LIABILITY, EQUITY, INCOME, EXPENSE
	Nature         string                 `json:"nature" gorm:"not null"` // D, C
	Level          int                    `json:"level" gorm:"not null"`
	ParentID       *string                `json:"parent_id" gorm:"type:uuid"`
	Parent         *Account               `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children       []Account              `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	IsDetail       bool                   `json:"is_detail" gorm:"default:false"`
	IsActive       bool                   `json:"is_active" gorm:"default:true"`
	Metadata       map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
}

// Voucher representa un comprobante
type Voucher struct {
	BaseModel
	OrganizationID  string                 `json:"organization_id" gorm:"type:uuid;not null;index"`
	Organization    Organization           `json:"-" gorm:"foreignKey:OrganizationID"`
	VoucherNumber   string                 `json:"voucher_number" gorm:"not null;index"`
	VoucherType     string                 `json:"voucher_type" gorm:"not null;index"`
	VoucherDate     time.Time              `json:"voucher_date" gorm:"not null;index"`
	Description     string                 `json:"description" gorm:"not null"`
	TotalAmount     float64                `json:"total_amount" gorm:"not null"`
	Subtotal        float64                `json:"subtotal"`
	TaxAmount       float64                `json:"tax_amount"`
	RetentionAmount float64                `json:"retention_amount"`
	CurrencyCode    string                 `json:"currency_code" gorm:"default:'COP'"`
	ExchangeRate    float64                `json:"exchange_rate" gorm:"default:1.0"`
	ThirdPartyID    *string                `json:"third_party_id" gorm:"type:uuid"`
	ThirdParty      *ThirdParty            `json:"third_party,omitempty" gorm:"foreignKey:ThirdPartyID"`
	Status          string                 `json:"status" gorm:"default:'PENDING';index"` // PENDING, PROCESSING, PROCESSED, ERROR, CANCELLED
	ErrorMessage    *string                `json:"error_message"`
	SourceSystem    string                 `json:"source_system" gorm:"default:'WEB_APP'"`
	ProcessedAt     *time.Time             `json:"processed_at"`
	Lines           []VoucherLine          `json:"lines" gorm:"foreignKey:VoucherID"`
	JournalEntries  []JournalEntry         `json:"journal_entries,omitempty" gorm:"foreignKey:VoucherID"`
	Metadata        map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
}

// VoucherLine representa una línea de un comprobante
type VoucherLine struct {
	BaseModel
	VoucherID      string   `json:"voucher_id" gorm:"type:uuid;not null;index"`
	Voucher        Voucher  `json:"-" gorm:"foreignKey:VoucherID"`
	LineNumber     int      `json:"line_number" gorm:"not null"`
	AccountID      *string  `json:"account_id" gorm:"type:uuid"`
	Account        *Account `json:"account,omitempty" gorm:"foreignKey:AccountID"`
	Description    string   `json:"description" gorm:"not null"`
	Quantity       float64  `json:"quantity" gorm:"default:1"`
	UnitPrice      float64  `json:"unit_price" gorm:"not null"`
	LineAmount     float64  `json:"line_amount" gorm:"not null"`
	TaxCode        *string  `json:"tax_code"`
	TaxRate        float64  `json:"tax_rate" gorm:"default:0"`
	TaxAmount      float64  `json:"tax_amount" gorm:"default:0"`
	DiscountRate   float64  `json:"discount_rate" gorm:"default:0"`
	DiscountAmount float64  `json:"discount_amount" gorm:"default:0"`
}

// JournalEntry representa un asiento contable
type JournalEntry struct {
	BaseModel
	OrganizationID string                 `json:"organization_id" gorm:"type:uuid;not null;index"`
	Organization   Organization           `json:"-" gorm:"foreignKey:OrganizationID"`
	EntryNumber    int                    `json:"entry_number" gorm:"not null;index"`
	EntryDate      time.Time              `json:"entry_date" gorm:"not null;index"`
	VoucherID      *string                `json:"voucher_id" gorm:"type:uuid"`
	Voucher        *Voucher               `json:"voucher,omitempty" gorm:"foreignKey:VoucherID"`
	Description    string                 `json:"description" gorm:"not null"`
	EntryType      string                 `json:"entry_type" gorm:"default:'STANDARD'"` // STANDARD, ADJUSTMENT, CLOSING, REVERSAL
	Period         string                 `json:"period" gorm:"not null;index"`
	Status         string                 `json:"status" gorm:"default:'DRAFT'"` // DRAFT, PENDING, POSTED, CANCELLED
	IsReversed     bool                   `json:"is_reversed" gorm:"default:false"`
	ReversalID     *string                `json:"reversal_id" gorm:"type:uuid"`
	Reversal       *JournalEntry          `json:"reversal,omitempty" gorm:"foreignKey:ReversalID"`
	ApprovedAt     *time.Time             `json:"approved_at"`
	Lines          []JournalLine          `json:"lines" gorm:"foreignKey:JournalEntryID"`
	Metadata       map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
}

// JournalLine representa una línea de un asiento contable
type JournalLine struct {
	BaseModel
	JournalEntryID string       `json:"journal_entry_id" gorm:"type:uuid;not null;index"`
	JournalEntry   JournalEntry `json:"-" gorm:"foreignKey:JournalEntryID"`
	LineNumber     int          `json:"line_number" gorm:"not null"`
	AccountID      string       `json:"account_id" gorm:"type:uuid;not null"`
	Account        Account      `json:"account" gorm:"foreignKey:AccountID"`
	DebitAmount    float64      `json:"debit_amount" gorm:"default:0"`
	CreditAmount   float64      `json:"credit_amount" gorm:"default:0"`
	Description    string       `json:"description" gorm:"not null"`
}

// ThirdParty representa un tercero (cliente/proveedor)
type ThirdParty struct {
	BaseModel
	OrganizationID       string                   `json:"organization_id" gorm:"type:uuid;not null;index"`
	Organization         Organization             `json:"-" gorm:"foreignKey:OrganizationID"`
	DocumentType         string                   `json:"document_type" gorm:"not null"`
	DocumentNumber       string                   `json:"document_number" gorm:"not null;index"`
	Name                 string                   `json:"name" gorm:"not null"`
	CommercialName       string                   `json:"commercial_name"`
	Email                *string                  `json:"email"`
	Phone                *string                  `json:"phone"`
	Mobile               *string                  `json:"mobile"`
	Fax                  *string                  `json:"fax"`
	Website              *string                  `json:"website"`
	Address              *string                  `json:"address"`
	City                 *string                  `json:"city"`
	State                *string                  `json:"state"`
	PostalCode           *string                  `json:"postal_code"`
	CountryCode          string                   `json:"country_code" gorm:"not null"`
	TaxRegime            *string                  `json:"tax_regime"`
	TaxDetails           map[string]interface{}   `json:"tax_details" gorm:"type:jsonb"`
	FinancialInfo        map[string]interface{}   `json:"financial_info" gorm:"type:jsonb"`
	AccountConfiguration map[string]interface{}   `json:"account_configuration" gorm:"type:jsonb"`
	ContactPersons       []map[string]interface{} `json:"contact_persons" gorm:"type:jsonb"`
	RelationshipType     string                   `json:"relationship_type" gorm:"not null"` // CUSTOMER, VENDOR, BOTH
	CreditLimit          float64                  `json:"credit_limit" gorm:"default:0"`
	PaymentTerms         int                      `json:"payment_terms" gorm:"default:30"`
	IsActive             bool                     `json:"is_active" gorm:"default:true"`
	Metadata             map[string]interface{}   `json:"metadata" gorm:"type:jsonb"`
}

// DSLTemplate representa una plantilla DSL
type DSLTemplate struct {
	BaseModel
	OrganizationID string       `json:"organization_id" gorm:"type:uuid;not null;index"`
	Organization   Organization `json:"-" gorm:"foreignKey:OrganizationID"`
	TemplateCode   string       `json:"template_code" gorm:"not null;index"`
	Name           string       `json:"name" gorm:"not null"`
	Description    string       `json:"description"`
	VoucherType    string       `json:"voucher_type" gorm:"not null"`
	CountryCode    string       `json:"country_code" gorm:"not null"`
	DSLContent     string       `json:"dsl_content" gorm:"type:text;not null"`
	Version        int          `json:"version" gorm:"default:1"`
	IsActive       bool         `json:"is_active" gorm:"default:true"`
	CompiledAt     *time.Time   `json:"compiled_at"`
	CompileErrors  *string      `json:"compile_errors"`
}

// AuditLog representa un log de auditoría
type AuditLog struct {
	BaseModel
	OrganizationID string                 `json:"organization_id" gorm:"type:uuid;not null;index"`
	Organization   Organization           `json:"-" gorm:"foreignKey:OrganizationID"`
	Timestamp      time.Time              `json:"timestamp" gorm:"not null;index"`
	Level          string                 `json:"level" gorm:"not null;index"` // INFO, WARNING, ERROR
	EventType      string                 `json:"event_type" gorm:"not null;index"`
	Action         string                 `json:"action" gorm:"not null"`
	ResourceType   string                 `json:"resource_type" gorm:"not null;index"`
	ResourceID     string                 `json:"resource_id" gorm:"not null"`
	UserID         string                 `json:"user_id" gorm:"type:uuid;not null;index"`
	UserEmail      string                 `json:"user_email" gorm:"not null"`
	IPAddress      string                 `json:"ip_address" gorm:"not null"`
	UserAgent      string                 `json:"user_agent"`
	SessionID      string                 `json:"session_id"`
	Description    string                 `json:"description" gorm:"not null"`
	Details        map[string]interface{} `json:"details" gorm:"type:jsonb"`
	Changes        map[string]interface{} `json:"changes" gorm:"type:jsonb"`
	Result         string                 `json:"result" gorm:"not null"` // SUCCESS, ERROR, WARNING
	DurationMS     int                    `json:"duration_ms"`
}

/*
===========================
RESPUESTAS ESTÁNDAR
===========================
*/

// StandardResponse estructura estándar para respuestas API
type StandardResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// ErrorInfo información de error
type ErrorInfo struct {
	Code    string   `json:"code"`
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"`
}

// PaginationInfo información de paginación
type PaginationInfo struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
	Pages int `json:"pages"`
}

/*
===========================
MIDDLEWARE DE TENANT
===========================
*/

// TenantMiddleware middleware para manejo multi-tenant
func TenantMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		orgID := c.Query("organization_id")
		if orgID == "" {
			orgID = c.Get("X-Organization-ID")
		}

		if orgID == "" {
			return c.Status(400).JSON(StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "MISSING_ORGANIZATION_ID",
					Message: "organization_id es requerido",
				},
				Timestamp: time.Now(),
			})
		}

		// Validar que la organización existe
		// TODO: Implementar validación en base de datos

		c.Locals("organization_id", orgID)
		return c.Next()
	}
}

/*
===========================
CONTROLADORES - HEALTH & METRICS
===========================
*/

// HealthHandler maneja el endpoint de health check
func HealthHandler(c *fiber.Ctx) error {
	// TODO: Implementar validaciones reales de salud
	// - Verificar conexión a base de datos
	// - Verificar estado del motor DSL
	// - Verificar otros servicios críticos

	healthData := map[string]interface{}{
		"status":     "healthy",
		"version":    "2.0.0",
		"database":   "connected",
		"dsl_engine": "ready",
		"uptime":     3600.5,
	}

	return c.JSON(StandardResponse{
		Success:   true,
		Data:      healthData,
		Timestamp: time.Now(),
	})
}

// MetricsHandler maneja el endpoint de métricas
func MetricsHandler(c *fiber.Ctx) error {
	// TODO: Implementar métricas reales
	// - Usar Prometheus o métricas personalizadas
	// - Estadísticas de performance
	// - Contadores de uso

	metricsData := map[string]interface{}{
		"requests_total":       15420,
		"requests_per_second":  45.2,
		"response_time_avg":    120.5,
		"database_connections": 12,
		"memory_usage":         245.6,
		"cpu_usage":            15.3,
	}

	return c.JSON(StandardResponse{
		Success:   true,
		Data:      metricsData,
		Timestamp: time.Now(),
	})
}

/*
===========================
CONTROLADORES - DASHBOARD
===========================
*/

// DashboardHandler maneja el endpoint principal del dashboard
func DashboardHandler(c *fiber.Ctx) error {
	orgID := c.Locals("organization_id").(string)

	// TODO: Implementar consultas reales a la base de datos
	// - Calcular KPIs del período actual
	// - Obtener datos para gráficos
	// - Actividad reciente del usuario
	// - Estado de salud del sistema

	dashboardData := map[string]interface{}{
		"kpis": map[string]interface{}{
			"vouchers_today":          127,
			"vouchers_month":          3421,
			"total_amount_month":      458750000.50,
			"pending_vouchers":        23,
			"processing_rate":         98.5,
			"average_processing_time": 1.2,
		},
		"charts": map[string]interface{}{
			"vouchers_by_day": map[string]interface{}{
				"labels": []string{"Lun", "Mar", "Mié", "Jue", "Vie", "Sáb", "Dom"},
				"values": []int{156, 143, 167, 134, 189, 98, 127},
			},
			"vouchers_by_type": map[string]interface{}{
				"labels": []string{"Facturas", "Pagos", "Recibos", "Notas"},
				"values": []int{2456, 1234, 890, 341},
			},
		},
		"recent_activity": []map[string]interface{}{
			{
				"id":          "v-2024-0127",
				"type":        "invoice_sale",
				"amount":      2450000,
				"status":      "PROCESSED",
				"description": "Factura Cliente ABC S.A.",
				"created_at":  time.Now().Add(-2 * time.Hour),
			},
		},
		"system_health": map[string]interface{}{
			"status":               "healthy",
			"uptime":               99.98,
			"api_response_time":    45,
			"database_connections": 12,
			"queue_size":           3,
			"workers_active":       5,
			"cache_hit_rate":       87.5,
		},
	}

	return c.JSON(StandardResponse{
		Success:   true,
		Data:      dashboardData,
		Timestamp: time.Now(),
	})
}

/*
===========================
CONTROLADORES - VOUCHERS
===========================
*/

// VouchersListHandler lista comprobantes con filtros y paginación
func VouchersListHandler(c *fiber.Ctx) error {
	orgID := c.Locals("organization_id").(string)

	// Parámetros de consulta
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 25)
	voucherType := c.Query("type")
	status := c.Query("status")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	search := c.Query("search")

	// TODO: Implementar consulta real a la base de datos
	// - Filtrar por organization_id
	// - Aplicar filtros opcionales
	// - Implementar paginación
	// - Buscar por número, descripción o tercero

	// Simulación de datos basada en vouchers_list.json
	vouchers := []map[string]interface{}{
		{
			"id":               "voucher-001",
			"voucher_number":   "FV-000127",
			"voucher_type":     "invoice_sale",
			"voucher_date":     "2025-01-24",
			"description":      "Venta productos terminados - Cliente ABC",
			"total_amount":     2380000.00,
			"currency_code":    "COP",
			"third_party_name": "ABC Comercializadora S.A.S.",
			"status":           "PROCESSED",
			"created_at":       time.Now().Add(-1 * time.Hour),
			"processed_at":     time.Now().Add(-30 * time.Minute),
		},
	}

	pagination := PaginationInfo{
		Page:  page,
		Limit: limit,
		Total: 125,
		Pages: 5,
	}

	responseData := map[string]interface{}{
		"vouchers":   vouchers,
		"pagination": pagination,
		"summary": map[string]interface{}{
			"total_amount": 285750000.00,
			"by_status": map[string]int{
				"PENDING":    2,
				"PROCESSING": 5,
				"PROCESSED":  115,
				"ERROR":      1,
				"CANCELLED":  2,
			},
		},
	}

	return c.JSON(StandardResponse{
		Success:   true,
		Data:      responseData,
		Timestamp: time.Now(),
	})
}

// VoucherDetailHandler obtiene el detalle de un comprobante
func VoucherDetailHandler(c *fiber.Ctx) error {
	voucherID := c.Params("id")
	orgID := c.Locals("organization_id").(string)

	// TODO: Implementar consulta real a la base de datos
	// - Buscar por ID y organization_id
	// - Incluir líneas del comprobante
	// - Incluir información del tercero
	// - Incluir asientos contables si están procesados

	// Simulación basada en voucher_detail.json
	voucherDetail := map[string]interface{}{
		"id":               voucherID,
		"organization_id":  orgID,
		"voucher_number":   "FV-000127",
		"voucher_type":     "invoice_sale",
		"voucher_date":     "2025-01-24",
		"description":      "Venta productos terminados - Cliente ABC",
		"total_amount":     2380000.00,
		"subtotal":         2000000.00,
		"tax_amount":       380000.00,
		"retention_amount": 0.00,
		"currency_code":    "COP",
		"exchange_rate":    1.0,
		"status":           "PROCESSED",
		"source_system":    "WEB_APP",
		"created_at":       time.Now().Add(-1 * time.Hour),
		"processed_at":     time.Now().Add(-30 * time.Minute),
		"third_party": map[string]interface{}{
			"id":              "third-001",
			"document_type":   "NIT",
			"document_number": "900123456-1",
			"name":            "ABC Comercializadora S.A.S.",
			"email":           "facturacion@abc.com.co",
		},
		"lines": []map[string]interface{}{
			{
				"id":              "line-001",
				"line_number":     1,
				"description":     "Producto A - Unidad",
				"quantity":        10.00,
				"unit_price":      120000.00,
				"line_amount":     1200000.00,
				"tax_code":        "IVA_19",
				"tax_rate":        0.19,
				"tax_amount":      228000.00,
				"discount_rate":   0.00,
				"discount_amount": 0.00,
			},
		},
		"processing_info": map[string]interface{}{
			"dsl_template_used":     "invoice_sale_co",
			"processing_time_ms":    185,
			"journal_entries_count": 3,
			"validation_passed":     true,
		},
	}

	return c.JSON(StandardResponse{
		Success:   true,
		Data:      voucherDetail,
		Timestamp: time.Now(),
	})
}

// CreateVoucherHandler crea un nuevo comprobante
func CreateVoucherHandler(c *fiber.Ctx) error {
	orgID := c.Locals("organization_id").(string)

	// TODO: Implementar validación de entrada
	var request map[string]interface{}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_JSON",
				Message: "El formato JSON no es válido",
			},
			Timestamp: time.Now(),
		})
	}

	// TODO: Implementar validaciones de negocio
	// - Validar que el voucher_type sea válido
	// - Validar que las líneas estén balanceadas
	// - Validar que el tercero exista (si aplica)
	// - Validar que las cuentas existan
	// - Generar número de comprobante automático

	// TODO: Guardar en base de datos
	// - Crear registro de voucher
	// - Crear líneas asociadas
	// - Programar procesamiento automático si está habilitado

	// Simulación basada en voucher_create.json
	newVoucher := map[string]interface{}{
		"id":              "550e8400-e29b-41d4-a716-446655440005",
		"organization_id": orgID,
		"voucher_number":  "FV-000128",
		"voucher_type":    request["voucher_type"],
		"voucher_date":    request["voucher_date"],
		"description":     request["description"],
		"total_amount":    2380000.00,
		"status":          "PENDING",
		"source_system":   "WEB_APP",
		"created_at":      time.Now(),
		"processing_info": map[string]interface{}{
			"will_auto_process":      true,
			"estimated_process_time": "15 segundos",
			"dsl_template":           "invoice_sale_co",
			"validation_passed":      true,
		},
	}

	return c.Status(201).JSON(StandardResponse{
		Success:   true,
		Data:      newVoucher,
		Timestamp: time.Now(),
	})
}

// ProcessVoucherHandler procesa un comprobante con DSL
func ProcessVoucherHandler(c *fiber.Ctx) error {
	voucherID := c.Params("id")
	orgID := c.Locals("organization_id").(string)

	// TODO: Implementar procesamiento real con go-dsl
	// - Buscar el comprobante en la base de datos
	// - Determinar la plantilla DSL a usar
	// - Ejecutar la plantilla con los datos del comprobante
	// - Generar asientos contables
	// - Validar que los asientos estén balanceados
	// - Guardar asientos en la base de datos
	// - Actualizar estado del comprobante
	// - Registrar en logs de auditoría

	// Simulación del procesamiento
	processingResult := map[string]interface{}{
		"voucher_id":         voucherID,
		"journal_entry_id":   "entry-001",
		"lines_generated":    3,
		"total_debit":        2380000.00,
		"total_credit":       2380000.00,
		"processing_time_ms": 185,
		"template_used":      "invoice_sale_co",
		"validation": map[string]interface{}{
			"is_balanced":    true,
			"balance_check":  "PASS",
			"business_rules": "PASS",
			"data_integrity": "PASS",
		},
	}

	return c.JSON(StandardResponse{
		Success:   true,
		Data:      processingResult,
		Timestamp: time.Now(),
	})
}

/*
===========================
CONTROLADORES - JOURNAL ENTRIES
===========================
*/

// JournalEntriesListHandler lista asientos contables
func JournalEntriesListHandler(c *fiber.Ctx) error {
	orgID := c.Locals("organization_id").(string)
	period := c.Query("period")
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 25)

	// TODO: Implementar consulta real
	// - Filtrar por organization_id y período
	// - Incluir información resumida de las líneas
	// - Calcular totales de débito y crédito

	journalEntries := []map[string]interface{}{
		{
			"id":           "entry-001",
			"entry_number": 1,
			"entry_date":   "2025-01-24",
			"description":  "Asiento por Factura FV-000127",
			"entry_type":   "STANDARD",
			"total_debit":  2380000.00,
			"total_credit": 2380000.00,
			"lines_count":  3,
			"status":       "POSTED",
			"created_at":   time.Now().Add(-30 * time.Minute),
		},
	}

	pagination := PaginationInfo{
		Page:  page,
		Limit: limit,
		Total: 45,
		Pages: 2,
	}

	responseData := map[string]interface{}{
		"journal_entries": journalEntries,
		"pagination":      pagination,
	}

	return c.JSON(StandardResponse{
		Success:   true,
		Data:      responseData,
		Timestamp: time.Now(),
	})
}

// CreateJournalEntryHandler crea un asiento manual
func CreateJournalEntryHandler(c *fiber.Ctx) error {
	orgID := c.Locals("organization_id").(string)

	var request map[string]interface{}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_JSON",
				Message: "El formato JSON no es válido",
			},
			Timestamp: time.Now(),
		})
	}

	// TODO: Implementar validaciones
	// - Validar que las líneas estén balanceadas
	// - Validar que las cuentas existan
	// - Validar que la fecha esté en un período abierto
	// - Generar número de asiento automático

	// Simulación basada en journal_entry_create.json
	newEntry := map[string]interface{}{
		"id":              "entry-002",
		"organization_id": orgID,
		"entry_number":    2,
		"entry_date":      request["entry_date"],
		"description":     request["description"],
		"entry_type":      "STANDARD",
		"status":          "POSTED",
		"validation": map[string]interface{}{
			"is_balanced":  true,
			"total_debit":  500000.00,
			"total_credit": 500000.00,
		},
		"created_at": time.Now(),
	}

	return c.Status(201).JSON(StandardResponse{
		Success:   true,
		Data:      newEntry,
		Timestamp: time.Now(),
	})
}

/*
===========================
CONTROLADORES - ACCOUNTS
===========================
*/

// AccountsTreeHandler obtiene el árbol de cuentas
func AccountsTreeHandler(c *fiber.Ctx) error {
	orgID := c.Locals("organization_id").(string)
	expanded := c.QueryBool("expanded", false)

	// TODO: Implementar consulta real
	// - Construir árbol jerárquico de cuentas
	// - Incluir balances si se solicita
	// - Aplicar filtros de usuario

	// Simulación basada en accounts_tree.json
	accountsTree := []map[string]interface{}{
		{
			"id":           "account-001",
			"account_code": "1",
			"name":         "ACTIVOS",
			"type":         "ASSET",
			"nature":       "D",
			"level":        1,
			"is_detail":    false,
			"is_active":    true,
			"balance":      15420000.00,
			"children": []map[string]interface{}{
				{
					"id":           "account-002",
					"account_code": "11",
					"name":         "ACTIVOS CORRIENTES",
					"type":         "ASSET",
					"nature":       "D",
					"level":        2,
					"is_detail":    false,
					"is_active":    true,
					"balance":      8750000.00,
					"children": []map[string]interface{}{
						{
							"id":           "account-003",
							"account_code": "1105",
							"name":         "CAJA",
							"type":         "ASSET",
							"nature":       "D",
							"level":        3,
							"is_detail":    true,
							"is_active":    true,
							"balance":      2450000.00,
							"children":     []map[string]interface{}{},
						},
					},
				},
			},
		},
	}

	return c.JSON(StandardResponse{
		Success:   true,
		Data:      accountsTree,
		Timestamp: time.Now(),
	})
}

// AccountsListHandler lista cuentas con filtros
func AccountsListHandler(c *fiber.Ctx) error {
	orgID := c.Locals("organization_id").(string)
	accountType := c.Query("type")
	level := c.QueryInt("level", 0)
	detailOnly := c.QueryBool("detail_only", false)
	search := c.Query("search")

	// TODO: Implementar consulta real con filtros

	// Simulación basada en accounts_list.json
	accounts := []map[string]interface{}{
		{
			"id":           "account-003",
			"account_code": "1105",
			"name":         "CAJA",
			"type":         "ASSET",
			"nature":       "D",
			"level":        3,
			"is_detail":    true,
			"is_active":    true,
			"balance":      2450000.00,
		},
	}

	return c.JSON(StandardResponse{
		Success:   true,
		Data:      accounts,
		Timestamp: time.Now(),
	})
}

// AccountDetailHandler obtiene detalle de una cuenta
func AccountDetailHandler(c *fiber.Ctx) error {
	accountID := c.Params("id")
	orgID := c.Locals("organization_id").(string)

	// TODO: Implementar consulta real
	// - Incluir información de jerarquía
	// - Calcular balances del período
	// - Incluir transacciones recientes

	// Simulación basada en account_detail.json
	accountDetail := map[string]interface{}{
		"id":              accountID,
		"organization_id": orgID,
		"account_code":    "130505",
		"name":            "CLIENTES NACIONALES",
		"type":            "ASSET",
		"nature":          "D",
		"level":           4,
		"is_detail":       true,
		"is_active":       true,
		"parent": map[string]interface{}{
			"id":           "account-parent",
			"account_code": "1305",
			"name":         "CLIENTES",
		},
		"balance_info": map[string]interface{}{
			"current_balance": 12450000.00,
			"debit_balance":   15230000.00,
			"credit_balance":  2780000.00,
			"period_movement": 8450000.00,
		},
		"recent_movements": []map[string]interface{}{
			{
				"date":         "2025-01-24",
				"entry_number": 1,
				"description":  "Factura FV-000127",
				"debit":        2380000.00,
				"credit":       0.00,
				"balance":      12450000.00,
			},
		},
	}

	return c.JSON(StandardResponse{
		Success:   true,
		Data:      accountDetail,
		Timestamp: time.Now(),
	})
}

// AccountTypesHandler obtiene tipos de cuentas
func AccountTypesHandler(c *fiber.Ctx) error {
	accountTypes := []map[string]interface{}{
		{
			"code":        "ASSET",
			"name":        "Activos",
			"description": "Bienes y derechos de la empresa",
			"nature":      "D",
			"is_active":   true,
		},
		{
			"code":        "LIABILITY",
			"name":        "Pasivos",
			"description": "Obligaciones y deudas de la empresa",
			"nature":      "C",
			"is_active":   true,
		},
		{
			"code":        "EQUITY",
			"name":        "Patrimonio",
			"description": "Capital y reservas de la empresa",
			"nature":      "C",
			"is_active":   true,
		},
		{
			"code":        "INCOME",
			"name":        "Ingresos",
			"description": "Ingresos operacionales y no operacionales",
			"nature":      "C",
			"is_active":   true,
		},
		{
			"code":        "EXPENSE",
			"name":        "Gastos",
			"description": "Gastos operacionales y no operacionales",
			"nature":      "D",
			"is_active":   true,
		},
	}

	return c.JSON(StandardResponse{
		Success:   true,
		Data:      accountTypes,
		Timestamp: time.Now(),
	})
}

/*
===========================
CONTROLADORES - DSL
===========================
*/

// DSLTemplatesHandler lista plantillas DSL
func DSLTemplatesHandler(c *fiber.Ctx) error {
	orgID := c.Locals("organization_id").(string)
	voucherType := c.Query("voucher_type")
	activeOnly := c.QueryBool("active_only", true)

	// TODO: Implementar consulta real
	// - Filtrar por organization_id
	// - Aplicar filtros opcionales
	// - Incluir información de compilación

	// Simulación basada en dsl_templates.json
	templates := []map[string]interface{}{
		{
			"id":            "template-001",
			"template_code": "invoice_sale_co",
			"name":          "Factura de Venta Colombia",
			"description":   "Plantilla estándar para facturas de venta en Colombia",
			"voucher_type":  "invoice_sale",
			"country_code":  "CO",
			"version":       3,
			"is_active":     true,
			"compiled_at":   time.Now().Add(-24 * time.Hour),
			"created_at":    time.Now().Add(-30 * 24 * time.Hour),
		},
	}

	return c.JSON(StandardResponse{
		Success:   true,
		Data:      templates,
		Timestamp: time.Now(),
	})
}

// DSLTemplateDetailHandler obtiene detalle de una plantilla DSL
func DSLTemplateDetailHandler(c *fiber.Ctx) error {
	templateID := c.Params("id")
	orgID := c.Locals("organization_id").(string)

	// TODO: Implementar consulta real
	// - Incluir código DSL completo
	// - Información de compilación
	// - Estadísticas de uso

	// Simulación basada en dsl_template_detail.json
	templateDetail := map[string]interface{}{
		"id":              templateID,
		"organization_id": orgID,
		"template_code":   "invoice_sale_co",
		"name":            "Factura de Venta Colombia",
		"description":     "Plantilla para procesar facturas de venta según normativas colombianas",
		"voucher_type":    "invoice_sale",
		"country_code":    "CO",
		"version":         3,
		"is_active":       true,
		"dsl_content": `rule "process_invoice_sale" {
    when {
        voucher.type == "invoice_sale"
        voucher.country == "CO"
    }
    then {
        debit("130505", voucher.total_amount, voucher.description)
        credit("4135", voucher.subtotal, "Venta " + voucher.number)
        if (voucher.iva_amount > 0) {
            credit("2408", voucher.iva_amount, "IVA Factura " + voucher.number)
        }
    }
}`,
		"compilation_info": map[string]interface{}{
			"status":          "SUCCESS",
			"compiled_at":     time.Now().Add(-24 * time.Hour),
			"compile_time_ms": 150,
			"warnings":        []string{},
			"errors":          []string{},
		},
		"usage_statistics": map[string]interface{}{
			"times_used":        1250,
			"last_used":         time.Now().Add(-1 * time.Hour),
			"average_exec_time": 85.5,
			"success_rate":      99.2,
		},
	}

	return c.JSON(StandardResponse{
		Success:   true,
		Data:      templateDetail,
		Timestamp: time.Now(),
	})
}

// ValidateDSLHandler valida código DSL
func ValidateDSLHandler(c *fiber.Ctx) error {
	var request map[string]interface{}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_JSON",
				Message: "El formato JSON no es válido",
			},
			Timestamp: time.Now(),
		})
	}

	dslContent := request["dsl_content"].(string)

	// TODO: Implementar validación real con go-dsl
	// - Parsear el código DSL
	// - Verificar sintaxis
	// - Validar reglas de negocio
	// - Verificar que las cuentas referenciadas existan

	// Simulación basada en dsl_validate.json
	validationResult := map[string]interface{}{
		"valid": true,
		"syntax_check": map[string]interface{}{
			"status":  "PASS",
			"message": "Sintaxis correcta",
		},
		"semantic_check": map[string]interface{}{
			"status":  "PASS",
			"message": "Semántica correcta",
		},
		"business_rules": map[string]interface{}{
			"status":   "WARNING",
			"message":  "Se encontraron advertencias menores",
			"warnings": []string{"La función 'calculateTax()' está marcada como obsoleta"},
		},
		"performance_analysis": map[string]interface{}{
			"estimated_execution_time": 15.5,
			"complexity_score":         "LOW",
			"optimization_suggestions": []string{},
		},
	}

	return c.JSON(StandardResponse{
		Success:   true,
		Data:      validationResult,
		Timestamp: time.Now(),
	})
}

// TestDSLHandler prueba código DSL con datos
func TestDSLHandler(c *fiber.Ctx) error {
	var request map[string]interface{}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_JSON",
				Message: "El formato JSON no es válido",
			},
			Timestamp: time.Now(),
		})
	}

	dslContent := request["dsl_content"].(string)
	testData := request["test_data"].(map[string]interface{})

	// TODO: Implementar ejecución real con go-dsl
	// - Compilar el código DSL
	// - Ejecutar con los datos de prueba
	// - Generar asientos contables de prueba
	// - Validar que estén balanceados
	// - Calcular métricas de performance

	// Simulación basada en dsl_test.json
	testResult := map[string]interface{}{
		"test_result": map[string]interface{}{
			"execution_successful": true,
			"execution_time":       18.5,
			"generated_entries":    3,
			"is_balanced":          true,
			"total_debit":          2380000.00,
			"total_credit":         2380000.00,
		},
		"generated_journal_lines": []map[string]interface{}{
			{
				"line_number":   1,
				"account_code":  "130505",
				"account_name":  "CLIENTES NACIONALES",
				"debit_amount":  2380000.00,
				"credit_amount": 0.00,
				"description":   "Venta productos terminados - Cliente ABC",
			},
			{
				"line_number":   2,
				"account_code":  "4135",
				"account_name":  "VENTAS",
				"debit_amount":  0.00,
				"credit_amount": 2000000.00,
				"description":   "Venta FV-TEST-001",
			},
			{
				"line_number":   3,
				"account_code":  "2408",
				"account_name":  "IVA POR PAGAR",
				"debit_amount":  0.00,
				"credit_amount": 380000.00,
				"description":   "IVA Factura FV-TEST-001",
			},
		},
		"validation_checks": map[string]interface{}{
			"balance_check": map[string]interface{}{
				"status":  "PASS",
				"message": "Débitos = Créditos (2,380,000.00)",
			},
			"account_existence": map[string]interface{}{
				"status":  "PASS",
				"message": "Todas las cuentas existen en el plan",
			},
			"business_rules": map[string]interface{}{
				"status":  "PASS",
				"message": "Todas las reglas de negocio se cumplen",
			},
		},
		"performance_metrics": map[string]interface{}{
			"parse_time":      2.1,
			"execution_time":  14.8,
			"validation_time": 1.6,
			"total_time":      18.5,
			"memory_used":     1024,
			"cpu_cycles":      15420,
		},
	}

	return c.JSON(StandardResponse{
		Success:   true,
		Data:      testResult,
		Timestamp: time.Now(),
	})
}

/*
===========================
CONTROLADORES - THIRD PARTIES
===========================
*/

// ThirdPartiesSearchHandler busca terceros
func ThirdPartiesSearchHandler(c *fiber.Ctx) error {
	orgID := c.Locals("organization_id").(string)
	query := c.Query("query")
	relationshipType := c.Query("relationship_type")
	activeOnly := c.QueryBool("active_only", true)
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 25)

	// TODO: Implementar búsqueda real
	// - Búsqueda por nombre, documento o email
	// - Filtros por tipo de relación
	// - Paginación
	// - Ordenamiento

	// Simulación basada en third_parties_search.json
	thirdParties := []map[string]interface{}{
		{
			"id":                "third-001",
			"document_type":     "NIT",
			"document_number":   "900123456-1",
			"name":              "ABC Comercializadora S.A.S.",
			"email":             "facturacion@abc.com.co",
			"phone":             "+57 1 234-5678",
			"city":              "Bogotá",
			"country_code":      "CO",
			"credit_limit":      50000000.00,
			"current_balance":   12450000.00,
			"payment_terms":     30,
			"is_active":         true,
			"relationship_type": "CUSTOMER",
		},
	}

	responseData := map[string]interface{}{
		"third_parties": thirdParties,
		"search_metadata": map[string]interface{}{
			"query":         query,
			"total_found":   5,
			"search_fields": []string{"name", "document_number", "email"},
			"filters_applied": map[string]interface{}{
				"organization_id":   orgID,
				"active_only":       activeOnly,
				"relationship_type": relationshipType,
			},
		},
		"summary": map[string]interface{}{
			"total_customers": 3,
			"total_vendors":   1,
			"total_both":      1,
		},
	}

	return c.JSON(StandardResponse{
		Success:   true,
		Data:      responseData,
		Timestamp: time.Now(),
	})
}

// ThirdPartyDetailHandler obtiene detalle de un tercero
func ThirdPartyDetailHandler(c *fiber.Ctx) error {
	thirdPartyID := c.Params("id")
	orgID := c.Locals("organization_id").(string)

	// TODO: Implementar consulta real
	// - Incluir información completa del tercero
	// - Transacciones recientes
	// - Estadísticas de la relación
	// - Alertas de crédito

	// Simulación basada en third_party_detail.json
	thirdPartyDetail := map[string]interface{}{
		"id":                thirdPartyID,
		"organization_id":   orgID,
		"document_type":     "NIT",
		"document_number":   "900123456-1",
		"name":              "ABC Comercializadora S.A.S.",
		"commercial_name":   "ABC Comercializadora",
		"email":             "facturacion@abc.com.co",
		"phone":             "+57 1 234-5678",
		"address":           "Calle 100 # 50-25 Oficina 801",
		"city":              "Bogotá",
		"country_code":      "CO",
		"relationship_type": "CUSTOMER",
		"financial_info": map[string]interface{}{
			"credit_limit":     50000000.00,
			"current_balance":  12450000.00,
			"available_credit": 37550000.00,
			"payment_terms":    30,
			"credit_score":     85,
			"risk_level":       "BAJO",
		},
		"recent_transactions": []map[string]interface{}{
			{
				"date":     "2025-01-22",
				"type":     "INVOICE",
				"number":   "FV-000125",
				"amount":   2800000.00,
				"status":   "PAID",
				"due_date": "2025-02-21",
			},
		},
		"statistics": map[string]interface{}{
			"total_invoices_ytd":     145,
			"total_amount_ytd":       285750000.00,
			"average_invoice_amount": 1971034.48,
			"payment_history": map[string]interface{}{
				"on_time_payments":      92,
				"late_payments":         3,
				"average_payment_delay": 2.5,
			},
		},
		"alerts": []map[string]interface{}{
			{
				"type":       "CREDIT_LIMIT",
				"severity":   "WARNING",
				"message":    "Cliente cerca del límite de crédito (75% utilizado)",
				"percentage": 75.1,
			},
		},
	}

	return c.JSON(StandardResponse{
		Success:   true,
		Data:      thirdPartyDetail,
		Timestamp: time.Now(),
	})
}

/*
===========================
CONTROLADORES - CATALOGS
===========================
*/

// TaxTypesHandler obtiene tipos de impuestos
func TaxTypesHandler(c *fiber.Ctx) error {
	// Simulación basada en tax_types.json
	taxTypes := map[string]interface{}{
		"tax_types": []map[string]interface{}{
			{
				"id":           "tax-001",
				"code":         "IVA_19",
				"name":         "IVA 19%",
				"description":  "Impuesto al Valor Agregado - Tarifa General",
				"rate":         0.19,
				"type":         "PERCENTAGE",
				"category":     "IVA",
				"is_retention": false,
				"applies_to":   []string{"SALES", "PURCHASES"},
				"country_code": "CO",
				"is_active":    true,
			},
		},
		"categories": []map[string]interface{}{
			{
				"code":        "IVA",
				"name":        "Impuesto al Valor Agregado",
				"description": "Impuesto nacional sobre las ventas",
			},
		},
		"summary": map[string]interface{}{
			"total_tax_types":     8,
			"active_tax_types":    8,
			"retention_types":     4,
			"sales_applicable":    4,
			"purchase_applicable": 6,
		},
	}

	return c.JSON(StandardResponse{
		Success:   true,
		Data:      taxTypes,
		Timestamp: time.Now(),
	})
}

// DocumentTypesHandler obtiene tipos de documentos
func DocumentTypesHandler(c *fiber.Ctx) error {
	// Simulación basada en document_types.json
	documentTypes := map[string]interface{}{
		"document_types": []map[string]interface{}{
			{
				"id":             "doc-001",
				"code":           "CC",
				"name":           "Cédula de Ciudadanía",
				"description":    "Documento de identidad para ciudadanos colombianos",
				"category":       "PERSONAL",
				"country_code":   "CO",
				"format_pattern": "^[0-9]{8,10}$",
				"format_example": "12345678",
				"min_length":     8,
				"max_length":     10,
				"is_active":      true,
			},
		},
		"categories": []map[string]interface{}{
			{
				"code":        "PERSONAL",
				"name":        "Documentos Personales",
				"description": "Documentos de identificación personal",
			},
		},
		"summary": map[string]interface{}{
			"total_document_types":    10,
			"active_document_types":   9,
			"personal_documents":      6,
			"business_documents":      2,
			"with_verification_digit": 2,
		},
	}

	return c.JSON(StandardResponse{
		Success:   true,
		Data:      documentTypes,
		Timestamp: time.Now(),
	})
}

// PeriodCurrentHandler obtiene el período contable actual
func PeriodCurrentHandler(c *fiber.Ctx) error {
	orgID := c.Locals("organization_id").(string)

	// TODO: Implementar consulta real
	// - Buscar período actual de la organización
	// - Incluir estadísticas del período
	// - Información de cierre
	// - Validaciones pendientes

	// Simulación basada en period_current.json
	currentPeriod := map[string]interface{}{
		"id":              "period-202501",
		"organization_id": orgID,
		"year":            2025,
		"month":           1,
		"period_code":     "2025-01",
		"period_name":     "Enero 2025",
		"start_date":      "2025-01-01",
		"end_date":        "2025-01-31",
		"status":          "OPEN",
		"is_current":      true,
		"is_adjustable":   true,
		"fiscal_year":     2025,
		"period_statistics": map[string]interface{}{
			"total_vouchers":        127,
			"total_journal_entries": 245,
			"processed_vouchers":    125,
			"pending_vouchers":      2,
			"error_vouchers":        0,
			"total_amount":          1580450000.00,
			"balance_status":        "BALANCED",
		},
		"closure_info": map[string]interface{}{
			"can_close":           false,
			"close_deadline":      "2025-02-15",
			"days_until_deadline": 22,
			"required_tasks": []map[string]interface{}{
				{
					"task":        "PROCESS_ALL_VOUCHERS",
					"status":      "PENDING",
					"description": "Procesar todos los comprobantes pendientes",
				},
			},
		},
	}

	return c.JSON(StandardResponse{
		Success:   true,
		Data:      currentPeriod,
		Timestamp: time.Now(),
	})
}

/*
===========================
CONTROLADORES - AUDIT
===========================
*/

// AuditLogsHandler obtiene logs de auditoría
func AuditLogsHandler(c *fiber.Ctx) error {
	orgID := c.Locals("organization_id").(string)
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	level := c.Query("level")
	eventType := c.Query("event_type")
	userID := c.Query("user_id")
	resourceType := c.Query("resource_type")
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 25)

	// TODO: Implementar consulta real
	// - Filtrar por organización y parámetros
	// - Paginación eficiente
	// - Agregaciones para el resumen

	// Simulación basada en audit_logs.json
	auditLogs := map[string]interface{}{
		"logs": []map[string]interface{}{
			{
				"id":              "audit-001",
				"organization_id": orgID,
				"timestamp":       time.Now().Add(-2 * time.Hour),
				"level":           "INFO",
				"event_type":      "VOUCHER_CREATED",
				"action":          "CREATE",
				"resource_type":   "VOUCHER",
				"resource_id":     "FV-000127",
				"user_id":         "user-001",
				"user_email":      "juan.perez@empresa.com",
				"ip_address":      "192.168.1.100",
				"description":     "Creación de factura de venta FV-000127",
				"result":          "SUCCESS",
				"duration_ms":     250,
			},
		},
		"pagination": PaginationInfo{
			Page:  page,
			Limit: limit,
			Total: 156,
			Pages: 7,
		},
		"summary": map[string]interface{}{
			"total_logs_today":  8,
			"info_logs":         6,
			"warning_logs":      1,
			"error_logs":        1,
			"unique_users":      4,
			"most_common_event": "VOUCHER_CREATED",
		},
	}

	return c.JSON(StandardResponse{
		Success:   true,
		Data:      auditLogs,
		Timestamp: time.Now(),
	})
}

/*
===========================
CONFIGURACIÓN DE RUTAS
===========================
*/

func setupRoutes(app *fiber.App) {
	// Grupo principal de la API
	api := app.Group("/api/v1")

	// Health y métricas (sin middleware de tenant)
	api.Get("/health", HealthHandler)
	api.Get("/metrics", MetricsHandler)

	// Aplicar middleware de tenant a todas las rutas que lo requieren
	tenantRoutes := api.Group("", TenantMiddleware())

	// Dashboard
	tenantRoutes.Get("/dashboard", DashboardHandler)

	// Vouchers
	voucherRoutes := tenantRoutes.Group("/vouchers")
	voucherRoutes.Get("/", VouchersListHandler)
	voucherRoutes.Post("/", CreateVoucherHandler)
	voucherRoutes.Get("/:id", VoucherDetailHandler)
	voucherRoutes.Post("/:id/process", ProcessVoucherHandler)

	// Journal Entries
	journalRoutes := tenantRoutes.Group("/journal-entries")
	journalRoutes.Get("/", JournalEntriesListHandler)
	journalRoutes.Post("/", CreateJournalEntryHandler)

	// Accounts
	accountRoutes := tenantRoutes.Group("/accounts")
	accountRoutes.Get("/tree", AccountsTreeHandler)
	accountRoutes.Get("/", AccountsListHandler)
	accountRoutes.Get("/types", AccountTypesHandler)
	accountRoutes.Get("/:id", AccountDetailHandler)

	// DSL
	dslRoutes := tenantRoutes.Group("/dsl")
	dslRoutes.Get("/templates", DSLTemplatesHandler)
	dslRoutes.Get("/templates/:id", DSLTemplateDetailHandler)
	dslRoutes.Post("/validate", ValidateDSLHandler)
	dslRoutes.Post("/test", TestDSLHandler)

	// Third Parties
	thirdPartyRoutes := tenantRoutes.Group("/third-parties")
	thirdPartyRoutes.Get("/search", ThirdPartiesSearchHandler)
	thirdPartyRoutes.Get("/:id", ThirdPartyDetailHandler)

	// Catalogs y Lookups
	catalogRoutes := tenantRoutes.Group("/lookups")
	catalogRoutes.Get("/tax-types", TaxTypesHandler)
	catalogRoutes.Get("/document-types", DocumentTypesHandler)

	// Periods
	periodRoutes := tenantRoutes.Group("/periods")
	periodRoutes.Get("/current", PeriodCurrentHandler)

	// Audit
	auditRoutes := tenantRoutes.Group("/audit")
	auditRoutes.Get("/logs", AuditLogsHandler)
}

/*
===========================
CONFIGURACIÓN DE BASE DE DATOS
===========================
*/

func setupDatabase(config *Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(config.DatabaseURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate modelos (solo en desarrollo)
	if config.Environment == "development" {
		err = db.AutoMigrate(
			&Organization{},
			&Account{},
			&Voucher{},
			&VoucherLine{},
			&JournalEntry{},
			&JournalLine{},
			&ThirdParty{},
			&DSLTemplate{},
			&AuditLog{},
		)
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}

/*
===========================
FUNCIÓN PRINCIPAL
===========================
*/

func main() {
	// Cargar configuración
	config := LoadConfig()

	// Configurar Fiber
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			message := "Error interno del servidor"

			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
				message = e.Message
			}

			return c.Status(code).JSON(StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "HTTP_ERROR",
					Message: message,
				},
				Timestamp: time.Now(),
			})
		},
	})

	// Middlewares globales
	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} - ${latency}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: config.CORSOrigins[0],
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Organization-ID",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	// Configurar base de datos
	db, err := setupDatabase(config)
	if err != nil {
		log.Fatalf("Error conectando a la base de datos: %v", err)
	}

	// Hacer la conexión disponible en el contexto
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("db", db)
		return c.Next()
	})

	// Configurar rutas
	setupRoutes(app)

	// Iniciar servidor
	log.Printf("Servidor iniciando en puerto %s", config.Port)
	log.Fatal(app.Listen(":" + config.Port))
}

/*
===========================
INTEGRACIÓN CON GO-DSL
===========================

Para integrar con go-dsl, crear el siguiente paquete en pkg/dslengine/:

package dslengine

import (
	"fmt"
	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

type DSLEngine struct {
	dsl *dslbuilder.DSL
}

func NewDSLEngine() *DSLEngine {
	dsl := dslbuilder.New("accounting-engine")

	// Configurar tokens
	dsl.Token("ACCOUNT", `"[0-9]+"`).
		Token("AMOUNT", `[0-9]+(\.[0-9]+)?`).
		Token("STRING", `"[^"]*"`).
		Token("WHEN", `when`).
		Token("THEN", `then`).
		Token("DEBIT", `debit`).
		Token("CREDIT", `credit`)

	// Configurar gramática
	dsl.Rule("rule", []string{"WHEN", "condition", "THEN", "actions"}, "processRule").
		Rule("condition", []string{"voucher_check"}, "validateCondition").
		Rule("actions", []string{"accounting_entry"}, "executeActions").
		Rule("accounting_entry", []string{"DEBIT", "ACCOUNT", "AMOUNT", "STRING"}, "createDebitEntry").
		Rule("accounting_entry", []string{"CREDIT", "ACCOUNT", "AMOUNT", "STRING"}, "createCreditEntry")

	// Registrar acciones
	dsl.Action("processRule", func(ctx map[string]interface{}) interface{} {
		// Lógica para procesar una regla completa
		return ctx
	}).
	Action("createDebitEntry", func(ctx map[string]interface{}) interface{} {
		// Crear entrada de débito
		return map[string]interface{}{
			"type": "debit",
			"account": ctx["ACCOUNT"],
			"amount": ctx["AMOUNT"],
			"description": ctx["STRING"],
		}
	}).
	Action("createCreditEntry", func(ctx map[string]interface{}) interface{} {
		// Crear entrada de crédito
		return map[string]interface{}{
			"type": "credit",
			"account": ctx["ACCOUNT"],
			"amount": ctx["AMOUNT"],
			"description": ctx["STRING"],
		}
	})

	return &DSLEngine{dsl: dsl}
}

func (e *DSLEngine) ProcessVoucher(dslContent string, voucherData map[string]interface{}) ([]JournalLine, error) {
	result, err := e.dsl.ParseWithContext(dslContent, voucherData)
	if err != nil {
		return nil, fmt.Errorf("error procesando DSL: %v", err)
	}

	// Convertir resultado a líneas de asiento contable
	lines := make([]JournalLine, 0)
	// ... lógica de conversión ...

	return lines, nil
}

===========================
ESQUEMA DE BASE DE DATOS POSTGRESQL
===========================

-- Crear extensiones necesarias
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Tabla de organizaciones (tenant principal)
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    commercial_name VARCHAR(255),
    document_type VARCHAR(10),
    tax_id VARCHAR(50) UNIQUE,
    country_code CHAR(2) NOT NULL,
    currency_default CHAR(3) DEFAULT 'COP',
    language CHAR(2) DEFAULT 'es',
    timezone VARCHAR(50) DEFAULT 'America/Bogota',
    contact_info JSONB,
    fiscal_info JSONB,
    accounting_configuration JSONB,
    dsl_configuration JSONB,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de cuentas contables
CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    account_code VARCHAR(20) NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('ASSET', 'LIABILITY', 'EQUITY', 'INCOME', 'EXPENSE')),
    nature CHAR(1) NOT NULL CHECK (nature IN ('D', 'C')),
    level INTEGER NOT NULL,
    parent_id UUID REFERENCES accounts(id),
    is_detail BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(organization_id, account_code)
);

-- Tabla de comprobantes
CREATE TABLE vouchers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    voucher_number VARCHAR(50) NOT NULL,
    voucher_type VARCHAR(30) NOT NULL,
    voucher_date DATE NOT NULL,
    description TEXT NOT NULL,
    total_amount DECIMAL(15,2) NOT NULL,
    subtotal DECIMAL(15,2),
    tax_amount DECIMAL(15,2),
    retention_amount DECIMAL(15,2),
    currency_code CHAR(3) DEFAULT 'COP',
    exchange_rate DECIMAL(10,4) DEFAULT 1.0,
    third_party_id UUID REFERENCES third_parties(id),
    status VARCHAR(20) DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'PROCESSING', 'PROCESSED', 'ERROR', 'CANCELLED')),
    error_message TEXT,
    source_system VARCHAR(50) DEFAULT 'WEB_APP',
    processed_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(organization_id, voucher_number)
);

-- Particionamiento por fecha para vouchers (mejora performance)
CREATE TABLE vouchers_y2025m01 PARTITION OF vouchers
FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

-- Crear índices optimizados
CREATE INDEX idx_vouchers_org_date ON vouchers (organization_id, voucher_date);
CREATE INDEX idx_vouchers_org_type ON vouchers (organization_id, voucher_type);
CREATE INDEX idx_vouchers_org_status ON vouchers (organization_id, status);
CREATE INDEX idx_vouchers_number_trgm ON vouchers USING gin (voucher_number gin_trgm_ops);

-- Tabla de líneas de comprobantes
CREATE TABLE voucher_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    voucher_id UUID NOT NULL REFERENCES vouchers(id) ON DELETE CASCADE,
    line_number INTEGER NOT NULL,
    account_id UUID REFERENCES accounts(id),
    description TEXT NOT NULL,
    quantity DECIMAL(10,3) DEFAULT 1,
    unit_price DECIMAL(15,2) NOT NULL,
    line_amount DECIMAL(15,2) NOT NULL,
    tax_code VARCHAR(20),
    tax_rate DECIMAL(5,4) DEFAULT 0,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    discount_rate DECIMAL(5,4) DEFAULT 0,
    discount_amount DECIMAL(15,2) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(voucher_id, line_number)
);

-- Tabla de asientos contables
CREATE TABLE journal_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    entry_number SERIAL,
    entry_date DATE NOT NULL,
    voucher_id UUID REFERENCES vouchers(id),
    description TEXT NOT NULL,
    entry_type VARCHAR(20) DEFAULT 'STANDARD' CHECK (entry_type IN ('STANDARD', 'ADJUSTMENT', 'CLOSING', 'REVERSAL')),
    period VARCHAR(7) NOT NULL, -- formato YYYY-MM
    status VARCHAR(20) DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'PENDING', 'POSTED', 'CANCELLED')),
    is_reversed BOOLEAN DEFAULT false,
    reversal_id UUID REFERENCES journal_entries(id),
    approved_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(organization_id, entry_number)
);

-- Particionamiento por período para journal_entries
CREATE TABLE journal_entries_y2025m01 PARTITION OF journal_entries
FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

-- Tabla de líneas de asientos contables
CREATE TABLE journal_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    journal_entry_id UUID NOT NULL REFERENCES journal_entries(id) ON DELETE CASCADE,
    line_number INTEGER NOT NULL,
    account_id UUID NOT NULL REFERENCES accounts(id),
    debit_amount DECIMAL(15,2) DEFAULT 0,
    credit_amount DECIMAL(15,2) DEFAULT 0,
    description TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(journal_entry_id, line_number),
    CONSTRAINT chk_debit_or_credit CHECK (
        (debit_amount > 0 AND credit_amount = 0) OR
        (credit_amount > 0 AND debit_amount = 0)
    )
);

-- Tabla de terceros
CREATE TABLE third_parties (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    document_type VARCHAR(10) NOT NULL,
    document_number VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    commercial_name VARCHAR(255),
    email VARCHAR(255),
    phone VARCHAR(20),
    mobile VARCHAR(20),
    fax VARCHAR(20),
    website VARCHAR(255),
    address TEXT,
    city VARCHAR(100),
    state VARCHAR(100),
    postal_code VARCHAR(20),
    country_code CHAR(2) NOT NULL,
    tax_regime VARCHAR(50),
    tax_details JSONB,
    financial_info JSONB,
    account_configuration JSONB,
    contact_persons JSONB,
    relationship_type VARCHAR(20) NOT NULL CHECK (relationship_type IN ('CUSTOMER', 'VENDOR', 'BOTH')),
    credit_limit DECIMAL(15,2) DEFAULT 0,
    payment_terms INTEGER DEFAULT 30,
    is_active BOOLEAN DEFAULT true,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(organization_id, document_number)
);

-- Tabla de plantillas DSL
CREATE TABLE dsl_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    template_code VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    voucher_type VARCHAR(30) NOT NULL,
    country_code CHAR(2) NOT NULL,
    dsl_content TEXT NOT NULL,
    version INTEGER DEFAULT 1,
    is_active BOOLEAN DEFAULT true,
    compiled_at TIMESTAMP WITH TIME ZONE,
    compile_errors TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(organization_id, template_code)
);

-- Tabla de logs de auditoría
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    level VARCHAR(10) NOT NULL CHECK (level IN ('INFO', 'WARNING', 'ERROR')),
    event_type VARCHAR(50) NOT NULL,
    action VARCHAR(20) NOT NULL,
    resource_type VARCHAR(30) NOT NULL,
    resource_id VARCHAR(255) NOT NULL,
    user_id UUID NOT NULL,
    user_email VARCHAR(255) NOT NULL,
    ip_address INET NOT NULL,
    user_agent TEXT,
    session_id VARCHAR(100),
    description TEXT NOT NULL,
    details JSONB,
    changes JSONB,
    result VARCHAR(10) NOT NULL CHECK (result IN ('SUCCESS', 'ERROR', 'WARNING')),
    duration_ms INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Particionamiento por mes para audit_logs (optimización para consultas)
CREATE TABLE audit_logs_y2025m01 PARTITION OF audit_logs
FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

-- Índices para optimización de consultas
CREATE INDEX idx_audit_logs_org_timestamp ON audit_logs (organization_id, timestamp DESC);
CREATE INDEX idx_audit_logs_org_level ON audit_logs (organization_id, level);
CREATE INDEX idx_audit_logs_org_event_type ON audit_logs (organization_id, event_type);
CREATE INDEX idx_audit_logs_org_resource ON audit_logs (organization_id, resource_type, resource_id);
CREATE INDEX idx_audit_logs_user ON audit_logs (user_id, timestamp DESC);

-- Función para auto-particionamiento mensual
CREATE OR REPLACE FUNCTION create_monthly_partitions()
RETURNS void AS $$
DECLARE
    start_date date;
    end_date date;
    table_name text;
BEGIN
    start_date := date_trunc('month', CURRENT_DATE);
    end_date := start_date + interval '1 month';
    table_name := 'vouchers_y' || extract(year from start_date) || 'm' ||
                  lpad(extract(month from start_date)::text, 2, '0');

    EXECUTE format('CREATE TABLE IF NOT EXISTS %I PARTITION OF vouchers
                    FOR VALUES FROM (%L) TO (%L)',
                   table_name, start_date, end_date);

    -- Repetir para journal_entries y audit_logs
    table_name := 'journal_entries_y' || extract(year from start_date) || 'm' ||
                  lpad(extract(month from start_date)::text, 2, '0');

    EXECUTE format('CREATE TABLE IF NOT EXISTS %I PARTITION OF journal_entries
                    FOR VALUES FROM (%L) TO (%L)',
                   table_name, start_date, end_date);

    table_name := 'audit_logs_y' || extract(year from start_date) || 'm' ||
                  lpad(extract(month from start_date)::text, 2, '0');

    EXECUTE format('CREATE TABLE IF NOT EXISTS %I PARTITION OF audit_logs
                    FOR VALUES FROM (%L) TO (%L)',
                   table_name, start_date, end_date);
END;
$$ LANGUAGE plpgsql;

-- Programar la función para ejecutarse mensualmente
-- (Requiere pg_cron extension en producción)
-- SELECT cron.schedule('create-partitions', '0 0 1 * *', 'SELECT create_monthly_partitions();');

===========================
EJEMPLOS DE TESTING
===========================

Crear archivo internal/handlers/vouchers_test.go:

package handlers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestVouchersListHandler(t *testing.T) {
	app := fiber.New()
	app.Get("/vouchers", VouchersListHandler)

	req := httptest.NewRequest("GET", "/vouchers?organization_id=123e4567-e89b-12d3-a456-426614174000", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response StandardResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)
}

func TestCreateVoucherHandler(t *testing.T) {
	app := fiber.New()
	app.Post("/vouchers", CreateVoucherHandler)

	requestBody := map[string]interface{}{
		"organization_id": "123e4567-e89b-12d3-a456-426614174000",
		"voucher_type":    "invoice_sale",
		"voucher_date":    "2025-01-24",
		"description":     "Test voucher",
		"lines": []map[string]interface{}{
			{
				"description": "Test product",
				"quantity":    1,
				"unit_price":  100000,
			},
		},
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/vouchers", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)

	var response StandardResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
}

===========================
DOCKER Y DEPLOYMENT
===========================

Dockerfile:

FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o motor-contable ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/motor-contable .

EXPOSE 8080
CMD ["./motor-contable"]

docker-compose.yml:

version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: motor_contable
      POSTGRES_USER: motor_user
      POSTGRES_PASSWORD: motor_pass
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  motor-contable:
    build: .
    ports:
      - "8080:8080"
    environment:
      DATABASE_URL: postgres://motor_user:motor_pass@postgres:5432/motor_contable?sslmode=disable
      REDIS_URL: redis://redis:6379
      ENVIRONMENT: development
    depends_on:
      - postgres
      - redis

volumes:
  postgres_data:

===========================
SCRIPTS DE MIGRACIÓN
===========================

Crear archivo scripts/migrate.go:

package main

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL no está configurada")
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Error conectando a la base de datos:", err)
	}

	// Ejecutar migraciones
	err = db.AutoMigrate(
		&Organization{},
		&Account{},
		&Voucher{},
		&VoucherLine{},
		&JournalEntry{},
		&JournalLine{},
		&ThirdParty{},
		&DSLTemplate{},
		&AuditLog{},
	)
	if err != nil {
		log.Fatal("Error ejecutando migraciones:", err)
	}

	log.Println("Migraciones ejecutadas exitosamente")
}

INSTRUCCIONES FINALES:
1. Copiar este código como base del proyecto
2. Instalar dependencias: go mod init && go mod tidy
3. Configurar variables de entorno
4. Ejecutar migraciones de base de datos
5. Implementar la lógica real reemplazando las simulaciones
6. Integrar con go-dsl para el procesamiento real
7. Añadir tests completos
8. Configurar CI/CD y deployment

Esta implementación proporciona una base sólida y completa para desarrollar
el Motor Contable en Go/Fiber con todas las APIs documentadas en Swagger.
*/
