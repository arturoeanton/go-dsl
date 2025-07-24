package models

import (
	"encoding/json"
	"time"
)

// DSLTemplate representa una plantilla de reglas DSL para automatización contable
// Basado en el sistema go-dsl y requerimientos de automatización
type DSLTemplate struct {
	BaseModel
	OrganizationID    string    `json:"organization_id" gorm:"index;not null"`
	Name              string    `json:"name" gorm:"not null"`
	Description       string    `json:"description"`
	Category          string    `json:"category" gorm:"not null;index"` // VOUCHER_GENERATION, VALIDATION, CALCULATION, etc.
	Version           string    `json:"version" gorm:"not null;default:'1.0'"`
	Status            string    `json:"status" gorm:"default:'DRAFT'"` // DRAFT, ACTIVE, INACTIVE, DEPRECATED
	DSLCode           string    `json:"dsl_code" gorm:"type:text;not null"`
	CompiledCode      string    `json:"compiled_code" gorm:"type:text"`
	CompilationStatus string    `json:"compilation_status" gorm:"default:'PENDING'"` // PENDING, SUCCESS, ERROR
	CompilationError  string    `json:"compilation_error" gorm:"type:text"`
	LastCompiledAt    *time.Time `json:"last_compiled_at"`
	CreatedByUserID   string    `json:"created_by_user_id" gorm:"index;not null"`
	UpdatedByUserID   string    `json:"updated_by_user_id" gorm:"index"`
	IsPublic          bool      `json:"is_public" gorm:"default:false"`
	UsageCount        int       `json:"usage_count" gorm:"default:0"`
	LastUsedAt        *time.Time `json:"last_used_at"`
	MetadataJSON      string    `json:"-" gorm:"type:text;column:metadata"`
	VariablesJSON     string    `json:"-" gorm:"type:text;column:variables"`
}

// DSLTemplateMetadata metadatos de la plantilla DSL
type DSLTemplateMetadata struct {
	Author           string                 `json:"author"`
	AuthorEmail      string                 `json:"author_email"`
	Tags             []string               `json:"tags"`
	Dependencies     []string               `json:"dependencies"`
	RequiredAccounts []string               `json:"required_accounts"`
	OutputFormat     string                 `json:"output_format"`
	Documentation    string                 `json:"documentation"`
	Examples         []DSLTemplateExample   `json:"examples"`
	CustomFields     map[string]interface{} `json:"custom_fields"`
}

// DSLTemplateExample ejemplo de uso de la plantilla
type DSLTemplateExample struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Input       map[string]interface{} `json:"input"`
	ExpectedOutput map[string]interface{} `json:"expected_output"`
}

// DSLTemplateVariable variable configurable de la plantilla
type DSLTemplateVariable struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"` // STRING, NUMBER, BOOLEAN, DATE, ACCOUNT, THIRD_PARTY
	Description  string      `json:"description"`
	DefaultValue interface{} `json:"default_value"`
	Required     bool        `json:"required"`
	Validation   string      `json:"validation"` // Expresión regular o regla de validación
	Options      []string    `json:"options,omitempty"` // Para variables tipo SELECT
}

// GetMetadata deserializa los metadatos de la plantilla
func (dt *DSLTemplate) GetMetadata() (*DSLTemplateMetadata, error) {
	if dt.MetadataJSON == "" {
		return &DSLTemplateMetadata{}, nil
	}
	var metadata DSLTemplateMetadata
	err := json.Unmarshal([]byte(dt.MetadataJSON), &metadata)
	return &metadata, err
}

// SetMetadata serializa los metadatos de la plantilla
func (dt *DSLTemplate) SetMetadata(metadata DSLTemplateMetadata) error {
	data, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	dt.MetadataJSON = string(data)
	return nil
}

// GetVariables deserializa las variables de la plantilla
func (dt *DSLTemplate) GetVariables() ([]DSLTemplateVariable, error) {
	if dt.VariablesJSON == "" {
		return []DSLTemplateVariable{}, nil
	}
	var variables []DSLTemplateVariable
	err := json.Unmarshal([]byte(dt.VariablesJSON), &variables)
	return variables, err
}

// SetVariables serializa las variables de la plantilla
func (dt *DSLTemplate) SetVariables(variables []DSLTemplateVariable) error {
	data, err := json.Marshal(variables)
	if err != nil {
		return err
	}
	dt.VariablesJSON = string(data)
	return nil
}

// IsCompiled verifica si la plantilla está compilada
func (dt *DSLTemplate) IsCompiled() bool {
	return dt.CompilationStatus == "SUCCESS" && dt.CompiledCode != ""
}

// CanExecute verifica si la plantilla puede ejecutarse
func (dt *DSLTemplate) CanExecute() bool {
	return dt.Status == "ACTIVE" && dt.IsCompiled()
}

// DSLTemplateDetail estructura completa para respuestas detalladas
type DSLTemplateDetail struct {
	*DSLTemplate
	Metadata  *DSLTemplateMetadata   `json:"metadata"`
	Variables []DSLTemplateVariable  `json:"variables"`
	CreatedBy *UserInfo              `json:"created_by,omitempty"`
	UpdatedBy *UserInfo              `json:"updated_by,omitempty"`
}

// UserInfo información básica del usuario
type UserInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ToDetail convierte un DSLTemplate a DSLTemplateDetail con metadatos
func (dt *DSLTemplate) ToDetail() (*DSLTemplateDetail, error) {
	metadata, err := dt.GetMetadata()
	if err != nil {
		return nil, err
	}
	
	variables, err := dt.GetVariables()
	if err != nil {
		return nil, err
	}
	
	return &DSLTemplateDetail{
		DSLTemplate: dt,
		Metadata:    metadata,
		Variables:   variables,
	}, nil
}

// DSLTemplateCreateRequest estructura para crear plantillas DSL
type DSLTemplateCreateRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	Category    string                 `json:"category" binding:"required"`
	DSLCode     string                 `json:"dsl_code" binding:"required"`
	IsPublic    bool                   `json:"is_public"`
	Metadata    *DSLTemplateMetadata   `json:"metadata"`
	Variables   []DSLTemplateVariable  `json:"variables"`
}

// DSLTemplateUpdateRequest estructura para actualizar plantillas DSL
type DSLTemplateUpdateRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	DSLCode     string                 `json:"dsl_code"`
	Status      string                 `json:"status"`
	IsPublic    bool                   `json:"is_public"`
	Metadata    *DSLTemplateMetadata   `json:"metadata"`
	Variables   []DSLTemplateVariable  `json:"variables"`
}

// DSLTemplatesListResponse respuesta para listado de plantillas DSL
type DSLTemplatesListResponse struct {
	Templates  []DSLTemplate   `json:"templates"`
	Pagination *PaginationInfo `json:"pagination"`
}

// DSLValidationRequest estructura para validar código DSL
type DSLValidationRequest struct {
	DSLCode   string                 `json:"dsl_code" binding:"required"`
	Variables map[string]interface{} `json:"variables"`
}

// DSLValidationResponse respuesta de validación DSL
type DSLValidationResponse struct {
	IsValid     bool                   `json:"is_valid"`
	Errors      []DSLValidationError   `json:"errors,omitempty"`
	Warnings    []DSLValidationWarning `json:"warnings,omitempty"`
	ParsedAST   map[string]interface{} `json:"parsed_ast,omitempty"`
	CompiledCode string                `json:"compiled_code,omitempty"`
}

// DSLValidationError error de validación DSL
type DSLValidationError struct {
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// DSLValidationWarning advertencia de validación DSL
type DSLValidationWarning struct {
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// DSLTestRequest estructura para probar plantillas DSL
type DSLTestRequest struct {
	TemplateID string                 `json:"template_id"`
	TestData   map[string]interface{} `json:"test_data" binding:"required"`
	Variables  map[string]interface{} `json:"variables"`
}

// DSLTestResponse respuesta de prueba DSL
type DSLTestResponse struct {
	Success         bool                   `json:"success"`
	ExecutionTimeMs int64                  `json:"execution_time_ms"`
	Result          map[string]interface{} `json:"result,omitempty"`
	GeneratedVoucher *Voucher              `json:"generated_voucher,omitempty"`
	Errors          []string               `json:"errors,omitempty"`
	Logs            []string               `json:"logs,omitempty"`
}

// DSLExecution representa una ejecución de plantilla DSL
type DSLExecution struct {
	BaseModel
	OrganizationID   string    `json:"organization_id" gorm:"index;not null"`
	TemplateID       string    `json:"template_id" gorm:"index;not null"`
	ExecutedByUserID string    `json:"executed_by_user_id" gorm:"index;not null"`
	Status           string    `json:"status" gorm:"not null"` // SUCCESS, ERROR, TIMEOUT
	ExecutionTimeMs  int64     `json:"execution_time_ms"`
	InputData        string    `json:"input_data" gorm:"type:text"`
	OutputData       string    `json:"output_data" gorm:"type:text"`
	ErrorMessage     string    `json:"error_message" gorm:"type:text"`
	GeneratedVoucherID *string `json:"generated_voucher_id" gorm:"index"`
	LogMessages      string    `json:"log_messages" gorm:"type:text"`
	ExecutedAt       time.Time `json:"executed_at" gorm:"not null;index"`
}