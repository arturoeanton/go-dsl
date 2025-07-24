package services

import (
	"motor-contable-poc/internal/data"
	"motor-contable-poc/internal/models"
	"gorm.io/gorm"
)

// OrganizationService maneja la lógica de negocio para organizaciones
type OrganizationService struct {
	orgRepo *data.OrganizationRepository
}

// NewOrganizationService crea una nueva instancia del servicio
func NewOrganizationService(db *gorm.DB) *OrganizationService {
	return &OrganizationService{
		orgRepo: data.NewOrganizationRepository(db),
	}
}

// GetCurrent obtiene la organización actual con todos sus detalles
// TODO: En el futuro, aquí se usaría go-dsl para aplicar reglas de negocio
// sobre qué información mostrar según el rol del usuario y configuraciones
func (s *OrganizationService) GetCurrent() (*models.OrganizationDetail, error) {
	org, err := s.orgRepo.GetCurrent()
	if err != nil {
		return nil, err
	}
	
	// Convertir a detalle completo
	detail, err := org.ToDetail()
	if err != nil {
		return nil, err
	}
	
	return detail, nil
}

// GetByID obtiene una organización por ID con validaciones de acceso
func (s *OrganizationService) GetByID(id string) (*models.OrganizationDetail, error) {
	org, err := s.orgRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	
	// TODO: Aquí se aplicarían reglas DSL para validar permisos de acceso
	// basadas en el contexto del usuario y políticas de seguridad
	
	detail, err := org.ToDetail()
	if err != nil {
		return nil, err
	}
	
	return detail, nil
}

// UpdateConfiguration actualiza la configuración de una organización
// TODO: En el futuro, se usaría go-dsl para validar cambios de configuración
// según reglas de negocio y restricciones contables
func (s *OrganizationService) UpdateConfiguration(orgID string, updates map[string]interface{}) error {
	org, err := s.orgRepo.GetByID(orgID)
	if err != nil {
		return err
	}
	
	// TODO: Aplicar reglas DSL para validar cambios permitidos
	// Por ejemplo: no permitir cambio de moneda si hay transacciones
	// no permitir cambio de año fiscal si hay períodos cerrados, etc.
	
	// Actualizar campos específicos según el mapa de updates
	if contactInfo, ok := updates["contact_info"]; ok {
		if ci, ok := contactInfo.(models.ContactInfo); ok {
			org.SetContactInfo(ci)
		}
	}
	
	if fiscalInfo, ok := updates["fiscal_info"]; ok {
		if fi, ok := fiscalInfo.(models.FiscalInfo); ok {
			org.SetFiscalInfo(fi)
		}
	}
	
	if accountingConfig, ok := updates["accounting_config"]; ok {
		if ac, ok := accountingConfig.(models.AccountingConfig); ok {
			org.SetAccountingConfig(ac)
		}
	}
	
	if dslConfig, ok := updates["dsl_config"]; ok {
		if dc, ok := dslConfig.(models.DSLConfig); ok {
			org.SetDSLConfig(dc)
		}
	}
	
	return s.orgRepo.Update(org)
}

// ValidateBusinessRules valida reglas de negocio de la organización
// TODO: Este método usaría go-dsl para ejecutar validaciones complejas
// basadas en reglas configurables de negocio contable
func (s *OrganizationService) ValidateBusinessRules(orgID string) ([]string, error) {
	org, err := s.orgRepo.GetByID(orgID)
	if err != nil {
		return nil, err
	}
	
	var violations []string
	
	// TODO: Ejecutar reglas DSL para validar:
	// - Configuración fiscal correcta según legislación colombiana
	// - Configuración contable consistente con el tipo de empresa
	// - Períodos contables configurados correctamente
	// - Plan de cuentas completo según actividad económica
	
	// Validaciones básicas por ahora
	if org.TaxID == "" {
		violations = append(violations, "NIT requerido para empresas colombianas")
	}
	
	fiscalInfo, err := org.GetFiscalInfo()
	if err == nil && fiscalInfo.AccountingStandard == "" {
		violations = append(violations, "Norma contable requerida (NIIF, NIIF PYMES, etc.)")
	}
	
	return violations, nil
}

// GetDashboardData obtiene datos para el dashboard de la organización
// TODO: En el futuro, se usaría go-dsl para generar dinámicamente
// los KPIs y métricas según configuraciones del usuario
func (s *OrganizationService) GetDashboardData(orgID string) (map[string]interface{}, error) {
	org, err := s.orgRepo.GetByID(orgID)
	if err != nil {
		return nil, err
	}
	
	dashboard := make(map[string]interface{})
	
	// Información básica de la organización
	dashboard["organization"] = map[string]interface{}{
		"id":              org.ID,
		"name":            org.Name,
		"commercial_name": org.CommercialName,
		"tax_id":          org.TaxID,
	}
	
	// TODO: Aquí se ejecutarían reglas DSL para calcular KPIs personalizados:
	// - Indicadores financieros según la industria
	// - Alertas automáticas basadas en umbrales configurables
	// - Tendencias y proyecciones según modelos predictivos
	// - Comparaciones con períodos anteriores
	
	// Placeholder para datos que vendrían de otros servicios
	dashboard["financial_summary"] = map[string]interface{}{
		"total_assets":      0,
		"total_liabilities": 0,
		"total_equity":      0,
		"current_period_revenue": 0,
		"current_period_expenses": 0,
	}
	
	dashboard["activity_summary"] = map[string]interface{}{
		"vouchers_this_month":      0,
		"journal_entries_pending":  0,
		"accounts_with_movement":   0,
		"third_parties_active":     0,
	}
	
	return dashboard, nil
}