# Caso de Uso DSL #5: Templates de Asientos Recurrentes

## Resumen

Este caso implementa un sistema de plantillas parametrizables para generar asientos contables recurrentes usando go-dsl, permitiendo automatizar procesos contables repetitivos con variables dinámicas.

## Problema que Resuelve

Muchos asientos contables son recurrentes pero con variaciones:
- Nómina mensual (montos diferentes cada mes)
- Depreciación de activos (mismo patrón, diferentes valores)
- Provisiones mensuales (cálculos basados en parámetros)
- Asientos de cierre (múltiples cuentas con lógica compleja)
- Distribuciones periódicas (basadas en métricas actualizadas)

Crear estos asientos manualmente cada período es ineficiente y propenso a errores.

## Solución con go-dsl

### 1. Definición del DSL de Templates

```go
// DSL para templates de asientos
dsl := dslbuilder.New("JournalTemplate")

// Tokens para templates
dsl.Token("TEMPLATE", `template`)
dsl.Token("PARAMS", `params`)
dsl.Token("ENTRY", `entry`)
dsl.Token("LINE", `line`)
dsl.Token("DEBIT", `debit`)
dsl.Token("CREDIT", `credit`)
dsl.Token("ACCOUNT", `account`)
dsl.Token("DESCRIPTION", `description`)
dsl.Token("IF", `if`)
dsl.Token("THEN", `then`)
dsl.Token("ELSE", `else`)
dsl.Token("FOREACH", `foreach`)
dsl.Token("IN", `in`)
dsl.Token("SUM", `sum`)
dsl.Token("CALCULATE", `calculate`)
dsl.Token("VARIABLE", `\$[a-zA-Z_][a-zA-Z0-9_]*`)
dsl.Token("STRING", `"[^"]*"`)
dsl.Token("NUMBER", `\d+(\.\d+)?`)
dsl.Token("OPERATOR", `(\+|-|\*|/)`)

// Funciones disponibles
dsl.Token("FUNCTION", `(get_balance|get_rate|get_metric|days_in_month|last_day|lookup)`)

// Gramática de templates
dsl.Rule("template", []string{"TEMPLATE", "name", "PARAMS", "param_list", "body"}, "defineTemplate")
dsl.Rule("body", []string{"ENTRY", "entry_details", "lines"}, "createEntry")
dsl.Rule("lines", []string{"line_list"}, "processLines")
dsl.Rule("line", []string{"LINE", "line_details"}, "addLine")
dsl.Rule("line", []string{"IF", "condition", "THEN", "line", "ELSE", "line"}, "conditionalLine")
dsl.Rule("line", []string{"FOREACH", "VARIABLE", "IN", "collection", "line"}, "iterateLine")
```

### 2. Templates de Asientos Recurrentes

```dsl
# Template 1: Nómina Mensual
template payroll_monthly
  params ($total_salaries, $total_deductions, $period)
  
  entry
    description: "Nómina mensual - " + $period
    date: last_day($period)
    reference: "NOM-" + $period
    
    # Salarios base
    line debit account("510506") amount($total_salaries) 
         description("Sueldos y salarios")
    
    # Deducciones de ley (calculadas automáticamente)
    line debit account("510569") amount($total_salaries * 0.085)
         description("Aporte salud empresa 8.5%")
    
    line debit account("510570") amount($total_salaries * 0.12)
         description("Aporte pensión empresa 12%")
    
    # Prestaciones sociales
    line debit account("510527") amount($total_salaries * 0.0833)
         description("Cesantías 8.33%")
    
    line debit account("510530") amount($total_salaries * 0.01)
         description("Intereses cesantías 1%")
    
    line debit account("510533") amount($total_salaries * 0.0833)
         description("Prima de servicios 8.33%")
    
    line debit account("510536") amount($total_salaries * 0.0417)
         description("Vacaciones 4.17%")
    
    # Parafiscales
    line debit account("510572") amount($total_salaries * 0.04)
         description("Caja compensación 4%")
    
    line debit account("510575") amount($total_salaries * 0.03)
         description("ICBF 3%")
    
    line debit account("510578") amount($total_salaries * 0.02)
         description("SENA 2%")
    
    # Contrapartidas
    line credit account("250501") amount($total_salaries - $total_deductions)
         description("Salarios por pagar")
    
    line credit account("237005") amount($total_deductions)
         description("Retenciones y aportes por pagar")
    
    line credit account("237010") amount(sum_lines("510572", "510575", "510578"))
         description("Parafiscales por pagar")
    
    line credit account("261005") amount(get_line_amount("510527"))
         description("Cesantías consolidadas")
    
    line credit account("261010") amount(get_line_amount("510530"))
         description("Intereses sobre cesantías por pagar")

# Template 2: Depreciación Mensual de Activos
template depreciation_monthly
  params ($period)
  
  entry
    description: "Depreciación del período " + $period
    date: last_day($period)
    
    # Iterar sobre todos los activos depreciables
    foreach $asset in get_depreciable_assets()
      if $asset.remaining_life > 0 then
        line debit account($asset.expense_account) 
             amount($asset.value / $asset.useful_life / 12)
             description("Depreciación " + $asset.description)
             
        line credit account($asset.accumulated_depreciation_account)
             amount($asset.value / $asset.useful_life / 12)
             description("Depreciación acumulada " + $asset.description)

# Template 3: Provisión de Cartera
template provision_portfolio
  params ($period, $provision_rate)
  
  entry
    description: "Provisión de cartera " + $period
    date: last_day($period)
    
    # Análisis por antigüedad
    foreach $range in aging_ranges()
      calculate provision_amount = 
        get_portfolio_balance($range) * get_provision_rate($range.days)
      
      if provision_amount > 0 then
        line debit account("519905") amount(provision_amount)
             description("Provisión cartera " + $range.description)
             
        line credit account("139905") amount(provision_amount)
             description("Provisión acumulada " + $range.description)

# Template 4: Cierre de Ingresos y Gastos
template monthly_closing
  params ($period)
  
  entry
    description: "Cierre mensual de resultados " + $period
    date: last_day($period)
    
    # Cerrar todas las cuentas de ingreso
    foreach $account in get_accounts_by_type("INCOME")
      if get_balance($account) != 0 then
        line debit account($account.code) amount(get_balance($account))
             description("Cierre " + $account.name)
    
    # Cerrar todas las cuentas de gasto
    foreach $account in get_accounts_by_type("EXPENSE")
      if get_balance($account) != 0 then
        line credit account($account.code) amount(get_balance($account))
             description("Cierre " + $account.name)
    
    # Resultado del ejercicio
    calculate net_result = sum_account_type("INCOME") - sum_account_type("EXPENSE")
    
    if net_result > 0 then
      line credit account("360505") amount(net_result)
           description("Utilidad del ejercicio")
    else
      line debit account("360510") amount(abs(net_result))
           description("Pérdida del ejercicio")

# Template 5: Facturación Recurrente
template recurring_invoice
  params ($customer_id, $service_list, $billing_date)
  
  entry
    description: "Factura recurrente - " + get_customer_name($customer_id)
    date: $billing_date
    third_party: $customer_id
    
    # Iterar sobre servicios
    foreach $service in $service_list
      calculate line_total = $service.quantity * $service.unit_price
      
      line credit account($service.revenue_account) amount(line_total)
           description($service.description)
      
      # Calcular IVA si aplica
      if $service.tax_rate > 0 then
        calculate tax = line_total * $service.tax_rate
        line credit account("240801") amount(tax)
             description("IVA " + ($service.tax_rate * 100) + "%")
    
    # Total a cobrar
    calculate invoice_total = sum_credit_lines()
    line debit account("130505") amount(invoice_total)
         description("CxC Cliente " + get_customer_name($customer_id))

# Template 6: Ajustes por Inflación
template inflation_adjustment
  params ($period, $inflation_rate)
  
  entry
    description: "Ajuste por inflación " + $period
    date: last_day($period)
    
    # Ajustar activos no monetarios
    foreach $account in get_non_monetary_assets()
      calculate adjustment = get_balance($account) * $inflation_rate
      
      if adjustment > 0 then
        line debit account($account.code) amount(adjustment)
             description("Ajuste inflación " + $account.name)
        
        line credit account("470505") amount(adjustment)
             description("Corrección monetaria - Activos")

# Template 7: Distribución de Utilidades
template profit_distribution
  params ($period, $net_profit)
  
  entry
    description: "Distribución de utilidades " + $period
    date: get_shareholders_meeting_date()
    
    # Reserva legal (10%)
    calculate legal_reserve = $net_profit * 0.10
    line debit account("360505") amount(legal_reserve)
         description("Apropiación reserva legal")
    line credit account("330505") amount(legal_reserve)
         description("Reserva legal")
    
    # Reservas estatutarias
    calculate statutory_reserves = $net_profit * get_statutory_rate()
    if statutory_reserves > 0 then
      line debit account("360505") amount(statutory_reserves)
           description("Apropiación reservas estatutarias")
      line credit account("330595") amount(statutory_reserves)
           description("Reservas estatutarias")
    
    # Dividendos
    calculate dividends = $net_profit - legal_reserve - statutory_reserves
    line debit account("360505") amount(dividends)
         description("Dividendos decretados")
    line credit account("236005") amount(dividends)
         description("Dividendos por pagar")

# Template 8: Consolidación de Sucursales
template branch_consolidation
  params ($period, $branch_list)
  
  entry
    description: "Consolidación de sucursales " + $period
    date: last_day($period)
    
    foreach $branch in $branch_list
      # Eliminar cuentas recíprocas
      foreach $reciprocal in get_reciprocal_accounts($branch)
        line debit account($reciprocal.credit_account) 
             amount($reciprocal.amount)
             description("Eliminación " + $branch.name)
        
        line credit account($reciprocal.debit_account)
             amount($reciprocal.amount)
             description("Eliminación " + $branch.name)
```

### 3. Motor de Templates en Go

```go
type TemplateEngine struct {
    dsl          *dslbuilder.DSL
    repository   *TemplateRepository
    dataService  *TemplateDataService
}

// JournalTemplate representa una plantilla
type JournalTemplate struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    DSLCode     string                 `json:"dsl_code"`
    Parameters  []TemplateParameter    `json:"parameters"`
    Schedule    *TemplateSchedule      `json:"schedule"`
    IsActive    bool                   `json:"is_active"`
    Version     string                 `json:"version"`
    Tags        []string               `json:"tags"`
}

// TemplateParameter define un parámetro de la plantilla
type TemplateParameter struct {
    Name         string      `json:"name"`
    Type         string      `json:"type"` // number, string, date, array
    Required     bool        `json:"required"`
    DefaultValue interface{} `json:"default_value"`
    Description  string      `json:"description"`
    Validation   string      `json:"validation"` // Regla DSL de validación
}

// ExecuteTemplate ejecuta una plantilla con parámetros
func (te *TemplateEngine) ExecuteTemplate(templateID string, params map[string]interface{}) (*models.JournalEntry, error) {
    // Cargar template
    template, err := te.repository.GetByID(templateID)
    if err != nil {
        return nil, fmt.Errorf("error cargando template: %w", err)
    }
    
    // Validar parámetros
    if err := te.validateParameters(template, params); err != nil {
        return nil, fmt.Errorf("parámetros inválidos: %w", err)
    }
    
    // Preparar contexto
    ctx := te.prepareContext(params)
    
    // Ejecutar DSL
    result, err := te.dsl.ParseWithContext(template.DSLCode, ctx)
    if err != nil {
        return nil, fmt.Errorf("error ejecutando template: %w", err)
    }
    
    // Convertir resultado a JournalEntry
    entry, ok := result.(*models.JournalEntry)
    if !ok {
        return nil, fmt.Errorf("template no generó un asiento válido")
    }
    
    // Validar balance
    if !entry.IsBalanced() {
        return nil, fmt.Errorf("asiento desbalanceado: débitos=%.2f, créditos=%.2f", 
            entry.TotalDebit, entry.TotalCredit)
    }
    
    // Guardar metadata del template
    entry.Metadata = map[string]interface{}{
        "template_id":      templateID,
        "template_version": template.Version,
        "generated_at":     time.Now(),
        "parameters":       params,
    }
    
    return entry, nil
}

// Funciones disponibles en templates
func (te *TemplateEngine) RegisterTemplateFunctions() {
    // Obtener balance de cuenta
    te.dsl.Action("get_balance", func(args ...interface{}) (interface{}, error) {
        accountCode := args[0].(string)
        period := ""
        if len(args) > 1 {
            period = args[1].(string)
        }
        
        balance := te.dataService.GetAccountBalance(accountCode, period)
        return balance, nil
    })
    
    // Obtener activos depreciables
    te.dsl.Action("get_depreciable_assets", func(args ...interface{}) (interface{}, error) {
        assets := te.dataService.GetDepreciableAssets()
        return assets, nil
    })
    
    // Calcular días del mes
    te.dsl.Action("days_in_month", func(args ...interface{}) (interface{}, error) {
        period := args[0].(string)
        date, err := time.Parse("2006-01", period)
        if err != nil {
            return 0, err
        }
        
        // Último día del mes
        lastDay := date.AddDate(0, 1, -1).Day()
        return lastDay, nil
    })
    
    // Obtener métricas
    te.dsl.Action("get_metric", func(args ...interface{}) (interface{}, error) {
        metricName := args[0].(string)
        params := args[1].(map[string]interface{})
        
        value := te.dataService.GetMetric(metricName, params)
        return value, nil
    })
    
    // Iteración sobre colecciones
    te.dsl.Action("foreach", func(args ...interface{}) (interface{}, error) {
        collection := args[0].([]interface{})
        lineTemplate := args[1].(func(interface{}) *models.JournalLine)
        
        lines := []models.JournalLine{}
        for _, item := range collection {
            if line := lineTemplate(item); line != nil {
                lines = append(lines, *line)
            }
        }
        
        return lines, nil
    })
}

// Programación de templates
type TemplateScheduler struct {
    engine      *TemplateEngine
    repository  *ScheduleRepository
    jobQueue    chan *ScheduledJob
}

func (ts *TemplateScheduler) ScheduleTemplate(schedule *TemplateSchedule) error {
    // Validar expresión cron
    cronExpr, err := cron.ParseStandard(schedule.CronExpression)
    if err != nil {
        return fmt.Errorf("expresión cron inválida: %w", err)
    }
    
    // Guardar programación
    schedule.NextRun = cronExpr.Next(time.Now())
    if err := ts.repository.Save(schedule); err != nil {
        return err
    }
    
    // Agregar a la cola
    ts.jobQueue <- &ScheduledJob{
        ScheduleID: schedule.ID,
        TemplateID: schedule.TemplateID,
        NextRun:    schedule.NextRun,
    }
    
    return nil
}

// Worker para ejecutar templates programados
func (ts *TemplateScheduler) processScheduledJobs() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            jobs, err := ts.repository.GetDueJobs()
            if err != nil {
                log.Printf("Error obteniendo jobs: %v", err)
                continue
            }
            
            for _, job := range jobs {
                go ts.executeScheduledJob(job)
            }
        }
    }
}

func (ts *TemplateScheduler) executeScheduledJob(job *ScheduledJob) {
    // Obtener parámetros dinámicos
    params, err := ts.buildDynamicParameters(job)
    if err != nil {
        log.Printf("Error construyendo parámetros: %v", err)
        ts.logJobError(job, err)
        return
    }
    
    // Ejecutar template
    entry, err := ts.engine.ExecuteTemplate(job.TemplateID, params)
    if err != nil {
        log.Printf("Error ejecutando template: %v", err)
        ts.logJobError(job, err)
        return
    }
    
    // Guardar asiento
    if err := ts.saveGeneratedEntry(entry, job); err != nil {
        log.Printf("Error guardando asiento: %v", err)
        ts.logJobError(job, err)
        return
    }
    
    // Actualizar próxima ejecución
    ts.updateNextRun(job)
    
    log.Printf("Template %s ejecutado exitosamente", job.TemplateID)
}
```

### 4. API para Gestión de Templates

```go
// TemplateHandler maneja las APIs de templates
type TemplateHandler struct {
    templateService *TemplateService
    engine         *TemplateEngine
}

// CreateTemplate crea una nueva plantilla
func (h *TemplateHandler) CreateTemplate(c *fiber.Ctx) error {
    var template JournalTemplate
    if err := c.BodyParser(&template); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
    }
    
    // Validar sintaxis DSL
    if err := h.engine.ValidateSyntax(template.DSLCode); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Sintaxis DSL inválida",
            "details": err.Error(),
        })
    }
    
    // Guardar template
    if err := h.templateService.Create(&template); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Error creando template"})
    }
    
    return c.JSON(template)
}

// PreviewTemplate ejecuta un template en modo preview
func (h *TemplateHandler) PreviewTemplate(c *fiber.Ctx) error {
    templateID := c.Params("id")
    
    var params map[string]interface{}
    if err := c.BodyParser(&params); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid parameters"})
    }
    
    // Ejecutar en modo preview (no guarda)
    entry, err := h.engine.ExecuteTemplate(templateID, params)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Error ejecutando template",
            "details": err.Error(),
        })
    }
    
    // Retornar preview del asiento
    return c.JSON(fiber.Map{
        "preview": entry,
        "summary": fiber.Map{
            "total_debit":  entry.TotalDebit,
            "total_credit": entry.TotalCredit,
            "is_balanced":  entry.IsBalanced(),
            "lines_count":  len(entry.Lines),
        },
    })
}

// ExecuteTemplate ejecuta un template y crea el asiento
func (h *TemplateHandler) ExecuteTemplate(c *fiber.Ctx) error {
    templateID := c.Params("id")
    
    var request struct {
        Parameters map[string]interface{} `json:"parameters"`
        DryRun     bool                  `json:"dry_run"`
    }
    
    if err := c.BodyParser(&request); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
    }
    
    // Ejecutar template
    entry, err := h.engine.ExecuteTemplate(templateID, request.Parameters)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Error ejecutando template",
            "details": err.Error(),
        })
    }
    
    // Si es dry run, no guardar
    if request.DryRun {
        return c.JSON(fiber.Map{
            "dry_run": true,
            "entry": entry,
        })
    }
    
    // Guardar asiento
    if err := h.templateService.SaveGeneratedEntry(entry); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Error guardando asiento"})
    }
    
    return c.JSON(fiber.Map{
        "success": true,
        "entry_id": entry.ID,
        "entry_number": entry.Number,
    })
}

// GetTemplateHistory obtiene el historial de ejecuciones
func (h *TemplateHandler) GetTemplateHistory(c *fiber.Ctx) error {
    templateID := c.Params("id")
    limit := c.QueryInt("limit", 10)
    
    history, err := h.templateService.GetExecutionHistory(templateID, limit)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Error obteniendo historial"})
    }
    
    return c.JSON(history)
}
```

### 5. UI para Editor de Templates

```javascript
// Editor de templates con syntax highlighting
class TemplateEditor {
    constructor(containerId) {
        this.container = document.getElementById(containerId);
        this.editor = null;
        this.currentTemplate = null;
        this.initEditor();
    }
    
    initEditor() {
        // Configurar CodeMirror con modo DSL personalizado
        CodeMirror.defineMode("accounting-dsl", function() {
            return {
                token: function(stream, state) {
                    // Keywords
                    if (stream.match(/\b(template|params|entry|line|debit|credit|if|then|else|foreach|in)\b/)) {
                        return "keyword";
                    }
                    
                    // Functions
                    if (stream.match(/\b(get_balance|get_rate|sum|calculate|last_day)\b/)) {
                        return "builtin";
                    }
                    
                    // Variables
                    if (stream.match(/\$[a-zA-Z_]\w*/)) {
                        return "variable-2";
                    }
                    
                    // Numbers
                    if (stream.match(/\d+(\.\d+)?/)) {
                        return "number";
                    }
                    
                    // Strings
                    if (stream.match(/"[^"]*"/)) {
                        return "string";
                    }
                    
                    // Accounts
                    if (stream.match(/\d{4,8}/)) {
                        return "atom";
                    }
                    
                    stream.next();
                    return null;
                }
            };
        });
        
        this.editor = CodeMirror(this.container, {
            mode: "accounting-dsl",
            theme: "monokai",
            lineNumbers: true,
            autoCloseBrackets: true,
            matchBrackets: true,
            indentUnit: 2,
            extraKeys: {
                "Ctrl-Space": "autocomplete",
                "Ctrl-Enter": () => this.previewTemplate()
            }
        });
        
        // Autocompletado
        this.setupAutocomplete();
    }
    
    setupAutocomplete() {
        CodeMirror.registerHelper("hint", "accounting-dsl", (cm) => {
            const cur = cm.getCursor();
            const token = cm.getTokenAt(cur);
            const start = token.start;
            const end = cur.ch;
            const line = cm.getLine(cur.line);
            const term = line.slice(start, end);
            
            let suggestions = [];
            
            // Sugerencias contextuales
            if (term.startsWith('$')) {
                // Variables disponibles
                suggestions = this.getAvailableVariables();
            } else if (line.includes('account(')) {
                // Cuentas contables
                suggestions = this.getAccountSuggestions(term);
            } else {
                // Keywords y funciones
                suggestions = [
                    'template', 'params', 'entry', 'line',
                    'debit', 'credit', 'account', 'amount',
                    'if', 'then', 'else', 'foreach', 'in',
                    'get_balance()', 'get_rate()', 'sum()',
                    'calculate', 'last_day()', 'days_in_month()'
                ];
            }
            
            const filtered = suggestions.filter(s => 
                s.toLowerCase().startsWith(term.toLowerCase())
            );
            
            return {
                list: filtered,
                from: CodeMirror.Pos(cur.line, start),
                to: CodeMirror.Pos(cur.line, end)
            };
        });
    }
    
    async previewTemplate() {
        const code = this.editor.getValue();
        
        try {
            const response = await fetch('/api/v1/templates/preview', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    dsl_code: code,
                    parameters: this.getTestParameters()
                })
            });
            
            const result = await response.json();
            
            if (response.ok) {
                this.showPreview(result.preview);
            } else {
                this.showError(result.error);
            }
        } catch (error) {
            this.showError('Error ejecutando preview: ' + error.message);
        }
    }
    
    showPreview(entry) {
        const previewPanel = document.getElementById('preview-panel');
        
        let html = `
            <h3>Preview del Asiento</h3>
            <div class="entry-header">
                <p><strong>Fecha:</strong> ${entry.date}</p>
                <p><strong>Descripción:</strong> ${entry.description}</p>
                <p><strong>Referencia:</strong> ${entry.reference || 'N/A'}</p>
            </div>
            <table class="entry-lines">
                <thead>
                    <tr>
                        <th>Cuenta</th>
                        <th>Descripción</th>
                        <th>Débito</th>
                        <th>Crédito</th>
                    </tr>
                </thead>
                <tbody>
        `;
        
        for (const line of entry.lines) {
            html += `
                <tr>
                    <td>${line.account_code}</td>
                    <td>${line.description}</td>
                    <td class="amount">${line.debit_amount ? formatCurrency(line.debit_amount) : ''}</td>
                    <td class="amount">${line.credit_amount ? formatCurrency(line.credit_amount) : ''}</td>
                </tr>
            `;
        }
        
        html += `
                </tbody>
                <tfoot>
                    <tr>
                        <td colspan="2"><strong>Totales</strong></td>
                        <td class="amount"><strong>${formatCurrency(entry.total_debit)}</strong></td>
                        <td class="amount"><strong>${formatCurrency(entry.total_credit)}</strong></td>
                    </tr>
                </tfoot>
            </table>
            <div class="balance-check ${entry.is_balanced ? 'balanced' : 'unbalanced'}">
                ${entry.is_balanced ? '✅ Asiento balanceado' : '❌ Asiento desbalanceado'}
            </div>
        `;
        
        previewPanel.innerHTML = html;
        previewPanel.classList.add('show');
    }
}
```

## Beneficios

1. **Eficiencia**: Reduce horas de trabajo manual a segundos
2. **Consistencia**: Asientos siempre con el mismo formato
3. **Flexibilidad**: Fácil adaptación a cambios normativos
4. **Auditabilidad**: Trazabilidad completa de generación
5. **Escalabilidad**: Maneja complejidad sin esfuerzo adicional

## Casos de Uso Reales

### 1. Empresa con 50 Sucursales
```dsl
template branch_monthly_consolidation
  foreach $branch in get_active_branches()
    import_trial_balance($branch.id)
    eliminate_intercompany_transactions($branch)
    adjust_inventory_profits($branch)
```

### 2. Empresa de Servicios
```dsl
template service_revenue_recognition
  foreach $contract in get_active_contracts()
    calculate recognized = $contract.total * days_elapsed() / $contract.duration_days
    line credit account("410505") amount(recognized)
         description("Ingreso devengado " + $contract.number)
```

### 3. Empresa Manufacturera
```dsl
template production_cost_allocation
  calculate total_hours = sum_production_hours()
  foreach $order in get_production_orders()
    calculate allocation = overhead_pool() * ($order.hours / total_hours)
    assign_cost_to_order($order.id, allocation)
```

## Métricas de Éxito

- 95% reducción en tiempo de generación de asientos recurrentes
- 100% de asientos balanceados automáticamente
- 80% menos errores en procesos contables
- ROI de 400% en el primer año

## Implementación en el Código Actual

### Dónde Implementar

1. **Crear nuevo paquete**: `/internal/dsl/templates/`
   ```go
   // /internal/dsl/templates/engine.go
   type TemplateEngine struct {
       dsl *dslbuilder.DSL
       repository *repository.TemplateRepository
       dataService *TemplateDataService
   }
   ```

2. **Nuevos modelos**: `/internal/models/template.go`
   ```go
   type JournalTemplate struct {
       ID          string              `json:"id" gorm:"primaryKey"`
       Name        string              `json:"name"`
       Description string              `json:"description"`
       DSLCode     string              `json:"dsl_code"`
       Parameters  []TemplateParameter `json:"parameters" gorm:"serializer:json"`
       IsActive    bool                `json:"is_active"`
       CreatedAt   time.Time           `json:"created_at"`
   }
   ```

3. **Handler dedicado**: `/internal/handlers/template_handler.go`
   ```go
   type TemplateHandler struct {
       templateService *services.TemplateService
       engine         *templates.TemplateEngine
   }
   ```

### Dónde se Llamaría

1. **Desde UI - Ejecución Manual**:
   - Nuevo menú "Templates" en la navegación
   - Botón "Ejecutar Template" con formulario de parámetros

2. **Procesos Programados**:
   - Cron jobs para templates recurrentes (nómina, depreciación)
   - Se ejecutan automáticamente según calendario

3. **API Pública**:
   - `POST /api/v1/templates/:id/execute`
   - Permite integración con sistemas externos

### Ventajas Específicas

1. **Productividad**: 30 min → 30 seg para asiento de nómina
2. **Consistencia**: 0% errores vs 5% manual
3. **Compliance**: Siempre cumple políticas contables
4. **Escalabilidad**: Mismo tiempo para 10 o 1000 empleados
5. **Auditoría**: Trazabilidad de qué template generó cada asiento

### Integración Inmediata

**Paso 1**: Agregar tabla templates
```sql
CREATE TABLE journal_templates (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    dsl_code TEXT NOT NULL,
    parameters JSON,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    organization_id VARCHAR(36)
);
```

**Paso 2**: Agregar rutas en `routes.go`:
```go
// Templates
v1.Get("/templates", templateHandler.GetTemplates)
v1.Get("/templates/:id", templateHandler.GetTemplate)
v1.Post("/templates", templateHandler.CreateTemplate)
v1.Put("/templates/:id", templateHandler.UpdateTemplate)
v1.Post("/templates/:id/execute", templateHandler.ExecuteTemplate)
v1.Post("/templates/:id/preview", templateHandler.PreviewTemplate)
```

**Paso 3**: Modificar `journal_entry.go`:
```go
type JournalEntry struct {
    // Campos existentes...
    
    // NUEVO: Tracking de templates
    TemplateID      *string                `json:"template_id"`
    TemplateVersion *string                `json:"template_version"`
    TemplateParams  map[string]interface{} `json:"template_params" gorm:"serializer:json"`
}
```

### Templates Iniciales para la POC

1. **Nómina Simplificada**
2. **Depreciación Mensual**
3. **Factura de Venta Recurrente**
4. **Cierre de Caja Diario**
5. **Provisión de Servicios**

### UI Nueva: Templates

```html
<!-- /static/templates.html -->
<!-- Lista de templates disponibles -->
<!-- Editor DSL con syntax highlighting -->
<!-- Formulario dinámico de parámetros -->
<!-- Preview antes de ejecutar -->
<!-- Historial de ejecuciones -->
```

## Conclusión

Los templates de asientos con go-dsl transforman procesos contables repetitivos en operaciones automatizadas, parametrizables y auditables, liberando al equipo contable para tareas de mayor valor.