# Caso de Uso DSL #1: Validación Inteligente de Comprobantes

## Resumen

Este caso implementa validaciones complejas y personalizables para comprobantes contables usando go-dsl, permitiendo que cada organización defina sus propias reglas de negocio sin modificar el código del sistema.

## Problema que Resuelve

En contabilidad, las validaciones van más allá de verificar que los débitos igualen los créditos. Cada empresa tiene reglas específicas:
- Límites de montos por tipo de usuario
- Cuentas restringidas por período
- Validaciones cruzadas entre terceros y cuentas
- Reglas específicas por tipo de comprobante
- Restricciones por centro de costo

## Solución con go-dsl

### 1. Definición del DSL de Validación

```go
// DSL para definir reglas de validación de comprobantes
dsl := dslbuilder.New("VoucherValidation")

// Tokens básicos
dsl.Token("NUMBER", `\d+(\.\d+)?`)
dsl.Token("STRING", `"[^"]*"`)
dsl.Token("ACCOUNT", `\d{4,8}`)
dsl.Token("COMPARISON", `(>|<|>=|<=|==|!=)`)
dsl.Token("LOGICAL", `(AND|OR)`)
dsl.Token("IF", `if`)
dsl.Token("THEN", `then`)
dsl.Token("ERROR", `error`)
dsl.Token("WARNING", `warning`)

// Funciones disponibles
dsl.Token("FUNCTION", `(account_type|account_balance|voucher_total|third_party_type|user_role|period_status)`)
dsl.Token("LPAREN", `\(`)
dsl.Token("RPAREN", `\)`)

// Gramática de validación
dsl.Rule("validation", []string{"IF", "condition", "THEN", "action"}, "executeValidation")
dsl.Rule("condition", []string{"expression"}, "evaluateCondition")
dsl.Rule("condition", []string{"expression", "LOGICAL", "condition"}, "combineConditions")
dsl.Rule("expression", []string{"FUNCTION", "LPAREN", "STRING", "RPAREN", "COMPARISON", "value"}, "checkFunction")
dsl.Rule("expression", []string{"voucher_field", "COMPARISON", "value"}, "checkField")
dsl.Rule("value", []string{"NUMBER"}, "numericValue")
dsl.Rule("value", []string{"STRING"}, "stringValue")
dsl.Rule("action", []string{"ERROR", "STRING"}, "raiseError")
dsl.Rule("action", []string{"WARNING", "STRING"}, "raiseWarning")
```

### 2. Ejemplos de Reglas de Validación

```dsl
# Validación 1: Cuentas de bancos requieren referencia
if account_type("1110") == "BANK" AND voucher.reference == "" then error "Las cuentas bancarias requieren número de referencia"

# Validación 2: Límite de monto por rol de usuario
if user_role() == "JUNIOR" AND voucher_total() > 1000000 then error "Monto excede límite autorizado para su perfil"

# Validación 3: Restricción de cuentas por período
if period_status() == "CLOSING" AND account_type() != "TRANSIT" then error "Solo se permiten cuentas transitorias en período de cierre"

# Validación 4: Tercero requerido para cuentas por cobrar/pagar
if (account_type() == "RECEIVABLE" OR account_type() == "PAYABLE") AND third_party() == null then error "Esta cuenta requiere un tercero asociado"

# Validación 5: Validación cruzada de tipo de tercero
if account("130505") > 0 AND third_party_type() != "CUSTOMER" then error "La cuenta 130505 solo acepta clientes"

# Validación 6: Advertencia por montos inusuales
if voucher_type() == "EXPENSE" AND line_amount() > average_amount() * 3 then warning "Monto significativamente superior al promedio"

# Validación 7: Cuentas de impuestos con porcentaje válido
if account_type() == "TAX" AND tax_rate() NOT IN (0, 5, 19) then error "Tasa de impuesto no válida para Colombia"

# Validación 8: Centros de costo obligatorios
if account_requires_cost_center() AND cost_center() == null then error "Esta cuenta requiere centro de costo"

# Validación 9: Fecha dentro del período
if voucher_date() < period_start() OR voucher_date() > period_end() then error "Fecha fuera del período contable activo"

# Validación 10: Documentos soporte para gastos
if voucher_type() == "EXPENSE" AND amount() > 500000 AND attachments_count() == 0 then warning "Se recomienda adjuntar documentos soporte"
```

### 3. Implementación en Go

```go
// ValidacionService ejecuta las reglas DSL sobre un comprobante
type ValidationService struct {
    dsl        *dslbuilder.DSL
    repository *RulesRepository
}

func (v *ValidationService) ValidateVoucher(voucher *models.Voucher, ctx map[string]interface{}) []ValidationResult {
    // Cargar reglas activas para la organización
    rules, err := v.repository.GetActiveRules(voucher.OrganizationID, "VOUCHER_VALIDATION")
    if err != nil {
        return []ValidationResult{{Type: "ERROR", Message: "Error cargando reglas"}}
    }
    
    results := []ValidationResult{}
    
    // Preparar contexto con datos del comprobante
    ctx["voucher"] = voucher
    ctx["user"] = getCurrentUser()
    ctx["period"] = getCurrentPeriod()
    ctx["organization"] = getOrganization(voucher.OrganizationID)
    
    // Ejecutar cada regla
    for _, rule := range rules {
        result, err := v.dsl.ParseWithContext(rule.DSLCode, ctx)
        if err != nil {
            results = append(results, ValidationResult{
                Type: "ERROR",
                Message: fmt.Sprintf("Error en regla %s: %v", rule.Name, err),
            })
            continue
        }
        
        // Procesar resultado de la regla
        if validationResult, ok := result.(ValidationResult); ok {
            validationResult.RuleID = rule.ID
            validationResult.RuleName = rule.Name
            results = append(results, validationResult)
        }
    }
    
    return results
}

// Funciones disponibles en el DSL
func (v *ValidationService) RegisterFunctions() {
    // Función para verificar tipo de cuenta
    v.dsl.Action("account_type", func(args ...interface{}) (interface{}, error) {
        accountCode := args[0].(string)
        account := v.getAccount(accountCode)
        return account.Type, nil
    })
    
    // Función para obtener balance de cuenta
    v.dsl.Action("account_balance", func(args ...interface{}) (interface{}, error) {
        accountCode := args[0].(string)
        balance := v.getAccountBalance(accountCode)
        return balance, nil
    })
    
    // Función para total del comprobante
    v.dsl.Action("voucher_total", func(ctx map[string]interface{}) (interface{}, error) {
        voucher := ctx["voucher"].(*models.Voucher)
        return voucher.TotalDebit, nil
    })
    
    // Función para tipo de tercero
    v.dsl.Action("third_party_type", func(ctx map[string]interface{}) (interface{}, error) {
        voucher := ctx["voucher"].(*models.Voucher)
        if voucher.ThirdPartyID == nil {
            return nil, nil
        }
        thirdParty := v.getThirdParty(*voucher.ThirdPartyID)
        return thirdParty.Type, nil
    })
}
```

### 4. Integración con el Motor Contable

```go
// En el VoucherHandler
func (h *VoucherHandler) CreateVoucher(c *fiber.Ctx) error {
    var voucher models.Voucher
    if err := c.BodyParser(&voucher); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
    }
    
    // Validaciones básicas
    if err := h.basicValidations(&voucher); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": err.Error()})
    }
    
    // Ejecutar validaciones DSL
    validationResults := h.validationService.ValidateVoucher(&voucher, nil)
    
    // Procesar resultados
    errors := []string{}
    warnings := []string{}
    
    for _, result := range validationResults {
        switch result.Type {
        case "ERROR":
            errors = append(errors, result.Message)
        case "WARNING":
            warnings = append(warnings, result.Message)
        }
    }
    
    // Si hay errores, rechazar
    if len(errors) > 0 {
        return c.Status(400).JSON(fiber.Map{
            "error": "Validación fallida",
            "errors": errors,
            "warnings": warnings,
        })
    }
    
    // Guardar comprobante
    if err := h.voucherService.Create(&voucher); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Error guardando comprobante"})
    }
    
    return c.JSON(fiber.Map{
        "voucher": voucher,
        "warnings": warnings,
    })
}
```

### 5. Configuración de Reglas por Organización

```go
// Estructura para almacenar reglas
type ValidationRule struct {
    ID             string    `json:"id"`
    OrganizationID string    `json:"organization_id"`
    Name           string    `json:"name"`
    Description    string    `json:"description"`
    DSLCode        string    `json:"dsl_code"`
    Type           string    `json:"type"` // VOUCHER_VALIDATION
    Priority       int       `json:"priority"`
    IsActive       bool      `json:"is_active"`
    CreatedAt      time.Time `json:"created_at"`
    UpdatedAt      time.Time `json:"updated_at"`
}

// API para gestionar reglas
func (h *RulesHandler) CreateRule(c *fiber.Ctx) error {
    var rule ValidationRule
    if err := c.BodyParser(&rule); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
    }
    
    // Validar sintaxis DSL
    if err := h.dslService.ValidateSyntax(rule.DSLCode); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Sintaxis DSL inválida",
            "details": err.Error(),
        })
    }
    
    // Guardar regla
    if err := h.rulesService.Create(&rule); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Error guardando regla"})
    }
    
    return c.JSON(rule)
}
```

## Beneficios

1. **Flexibilidad**: Cada organización define sus propias reglas sin cambiar código
2. **Mantenibilidad**: Las reglas se gestionan desde la interfaz, no requieren deployment
3. **Auditoría**: Todas las validaciones quedan registradas con su regla origen
4. **Performance**: Las reglas se compilan una vez y se reutilizan
5. **Escalabilidad**: Fácil agregar nuevas funciones y operadores

## Casos de Uso Reales

1. **Empresa de Retail**: Validar que ventas al contado tengan forma de pago
2. **Empresa de Servicios**: Requerir orden de trabajo en gastos de mantenimiento
3. **Empresa Manufacturera**: Validar centro de costo en materias primas
4. **Entidad Gubernamental**: Restricciones por tipo de gasto y período
5. **Empresa Multinacional**: Validaciones por país y moneda

## Métricas de Éxito

- Reducción del 90% en tiempo de implementación de nuevas validaciones
- Cero deployments para cambios en reglas de negocio
- Detección temprana del 95% de errores contables
- Satisfacción del usuario por personalización

## Implementación en el Código Actual

### Dónde Implementar

1. **Crear nuevo paquete**: `/internal/dsl/validation/`
   ```go
   // /internal/dsl/validation/engine.go
   type ValidationEngine struct {
       dsl *dslbuilder.DSL
       rulesRepo *repository.RulesRepository
   }
   ```

2. **Integrar en VoucherService**: `/internal/services/voucher_service.go`
   ```go
   type VoucherService struct {
       repository    repository.VoucherRepository
       journalService *JournalEntryService
       validationEngine *validation.ValidationEngine // NUEVO
   }
   ```

3. **Modificar método Post**: Línea ~85 en `voucher_service.go`
   ```go
   func (s *VoucherService) Post(voucher *models.Voucher, userID string) error {
       // NUEVO: Ejecutar validaciones DSL antes de procesar
       validationResults := s.validationEngine.ValidateVoucher(voucher, nil)
       if hasErrors(validationResults) {
           return NewValidationError(validationResults)
       }
       
       // Código existente continúa...
       return s.repository.WithTransaction(func(repo repository.VoucherRepository) error {
   ```

### Dónde se Llamaría

1. **En el Handler**: `/internal/handlers/voucher_handler.go`
   - Método `CreateVoucher` (línea ~45)
   - Método `UpdateVoucher` (línea ~85)
   - Método `PostVoucher` (línea ~125)

2. **En Procesos Batch**: Para validar importaciones masivas

### Ventajas Específicas

1. **Reducción de Código**: Elimina ~500 líneas de validaciones hardcodeadas
2. **Tiempo de Respuesta**: Las reglas cambian en caliente sin reiniciar
3. **Multi-tenant**: Cada organización con sus propias reglas
4. **Auditoría**: Log automático de qué regla falló y por qué
5. **Testing**: Validaciones probables sin compilar

### Ejemplo de Migración

**Antes** (código actual hardcodeado):
```go
// En voucher_service.go
if voucher.Amount > 1000000 && user.Role == "JUNIOR" {
    return errors.New("Monto excede límite")
}
```

**Después** (con DSL):
```dsl
if user_role() == "JUNIOR" AND voucher_total() > 1000000 
then error "Monto excede límite autorizado para su perfil"
```

## Conclusión

La validación inteligente con go-dsl transforma las reglas de negocio de código duro a configuración dinámica, permitiendo que el sistema se adapte a cualquier organización sin modificar el código fuente.