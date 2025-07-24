package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel contiene campos comunes para todos los modelos
// Incluye ID con UUID, timestamps automáticos
type BaseModel struct {
	ID        string    `json:"id" gorm:"type:text;primaryKey"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate se ejecuta antes de crear un registro para generar UUID
func (base *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if base.ID == "" {
		base.ID = uuid.New().String()
	}
	return nil
}

// StandardResponse estructura estándar para respuestas API siguiendo el patrón del Swagger
type StandardResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// ErrorInfo información de error según especificación Swagger
type ErrorInfo struct {
	Code    string   `json:"code"`
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"`
}

// PaginationInfo información de paginación para listados
type PaginationInfo struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
	Pages int `json:"pages"`
}

// NewSuccessResponse crea una respuesta exitosa estándar
func NewSuccessResponse(data interface{}) StandardResponse {
	return StandardResponse{
		Success:   true,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// NewErrorResponse crea una respuesta de error estándar
func NewErrorResponse(code, message string, details ...string) StandardResponse {
	return StandardResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now(),
	}
}

// Dashboard related models

// DashboardStats representa las estadísticas del dashboard
type DashboardStats struct {
	KPIs         DashboardKPIs    `json:"kpis"`
	Charts       DashboardCharts  `json:"charts"`
	SystemHealth SystemHealth     `json:"system_health"`
}

// DashboardKPIs contiene los indicadores clave de rendimiento
type DashboardKPIs struct {
	VouchersToday   int     `json:"vouchers_today"`
	VouchersMonth   int     `json:"vouchers_month"`
	PendingVouchers int     `json:"pending_vouchers"`
	ProcessingRate  float64 `json:"processing_rate"`
}

// DashboardCharts contiene los datos para los gráficos del dashboard
type DashboardCharts struct {
	VouchersByDay    ChartData `json:"vouchers_by_day"`
	VouchersByType   ChartData `json:"vouchers_by_type"`
	VouchersByStatus ChartData `json:"vouchers_by_status"`
}

// ChartData representa datos para gráficos
type ChartData struct {
	Labels []string `json:"labels"`
	Values []int    `json:"values"`
}

// SystemHealth representa el estado del sistema
type SystemHealth struct {
	Status          string  `json:"status"`
	Uptime          float64 `json:"uptime"`
	APIResponseTime int     `json:"api_response_time"`
	CacheHitRate    int     `json:"cache_hit_rate"`
}

// ActivityItem representa una actividad reciente
type ActivityItem struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Error       string    `json:"error,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}