# Caso de Uso DSL #3: Distribución Automática de Costos

## Resumen

Este caso implementa la distribución inteligente de costos entre centros de costo, proyectos y departamentos usando go-dsl, permitiendo definir reglas complejas de prorrateo y asignación.

## Problema que Resuelve

La distribución de costos indirectos es un desafío contable:
- Costos compartidos entre múltiples centros (electricidad, alquiler)
- Diferentes bases de distribución (m², empleados, horas)
- Distribuciones en cascada (costos de soporte a producción)
- Cambios frecuentes en la estructura organizacional
- Necesidad de trazabilidad en las asignaciones

## Solución con go-dsl

### 1. Definición del DSL de Distribución

```go
// DSL para distribución de costos
dsl := dslbuilder.New("CostDistribution")

// Tokens para expresiones de distribución
dsl.Token("NUMBER", `\d+(\.\d+)?`)
dsl.Token("PERCENTAGE", `\d+(\.\d+)?%`)
dsl.Token("IDENTIFIER", `[a-zA-Z_][a-zA-Z0-9_]*`)
dsl.Token("STRING", `"[^"]*"`)
dsl.Token("DISTRIBUTE", `distribute`)
dsl.Token("FROM", `from`)
dsl.Token("TO", `to`)
dsl.Token("USING", `using`)
dsl.Token("WHERE", `where`)
dsl.Token("PROPORTIONAL", `proportional`)
dsl.Token("EQUAL", `equal`)
dsl.Token("FIXED", `fixed`)
dsl.Token("CASCADE", `cascade`)

// Drivers de distribución
dsl.Token("DRIVER", `(headcount|area|revenue|hours|units|custom)`)

// Gramática de distribución
dsl.Rule("distribution", []string{"DISTRIBUTE", "source", "TO", "targets", "USING", "method"}, "executeDistribution")
dsl.Rule("source", []string{"FROM", "cost_source"}, "defineSource")
dsl.Rule("targets", []string{"target_list"}, "defineTargets")
dsl.Rule("method", []string{"DRIVER", "driver_params"}, "distributionMethod")
dsl.Rule("method", []string{"PERCENTAGE", "allocation_table"}, "percentageMethod")
dsl.Rule("method", []string{"CASCADE", "cascade_rules"}, "cascadeMethod")
```

### 2. Reglas de Distribución de Costos

```dsl
# Distribución 1: Alquiler por metros cuadrados
distribute cost_pool("ALQUILER")
  from cost_center("ADMINISTRACION")
  to all_cost_centers()
  using driver("area") proportional

# Distribución 2: Servicios públicos por consumo estimado
distribute account("513530") # Energía eléctrica
  from cost_center("SERVICIOS_GENERALES")
  to production_centers()
  using driver("machine_hours") * factor(1.2)
  where amount > 1000000

# Distribución 3: Costos de TI por usuarios
distribute cost_center("IT")
  to departments(exclude: "IT")
  using driver("active_users") 
    weighted by (
      "ADMIN": 0.5,
      "SALES": 1.0,
      "PRODUCTION": 1.5
    )

# Distribución 4: Costos de RRHH por headcount y salarios
distribute cost_center("RRHH")
  to all_departments()
  using composite_driver(
    headcount: 60%,
    payroll_amount: 40%
  )

# Distribución 5: Marketing por ingresos
distribute account("52*") where description contains "MARKETING"
  from cost_center("MARKETING")
  to business_units()
  using driver("revenue") 
    from period(current_month() - 1)

# Distribución 6: Cascada de costos de soporte
distribute cascade
  level 1: cost_center("MANTENIMIENTO") to production_lines() using "machine_value"
  level 2: cost_center("CALIDAD") to production_lines() using "production_units"
  level 3: production_lines() to products() using "standard_hours"

# Distribución 7: Costos compartidos con porcentajes fijos
distribute account("5260*") # Seguros
  to centers(
    "PLANTA_1": 40%,
    "PLANTA_2": 35%,
    "OFICINAS": 25%
  )

# Distribución 8: Costos de proyecto
distribute project_costs("PROYECTO_X")
  to project_phases()
  using driver("planned_hours") 
    adjusted_by actual_progress()

# Distribución 9: Costos ambientales
distribute environmental_costs()
  to production_centers()
  using driver("carbon_footprint")
    penalty_factor: high_emissions(2.0)

# Distribución 10: Pool de costos indirectos
distribute overhead_pool()
  to products()
  using activity_based_costing(
    activities: ["setup", "inspection", "movement"],
    cost_drivers: ["setup_hours", "inspection_count", "distance_moved"]
  )
```

### 3. Motor de Distribución en Go

```go
type DistributionEngine struct {
    dsl          *dslbuilder.DSL
    repository   *DistributionRulesRepository
    metricsRepo  *MetricsRepository
    costRepo     *CostRepository
}

// DistributionResult representa el resultado de una distribución
type DistributionResult struct {
    SourceID         string
    SourceAmount     float64
    Distributions    []Distribution
    Method          string
    Driver          string
    ExecutedAt      time.Time
    RuleID          string
}

type Distribution struct {
    TargetID         string
    TargetType       string // cost_center, project, product
    AllocatedAmount  float64
    Percentage       float64
    DriverValue      float64
    DriverUnit       string
}

func (de *DistributionEngine) ExecuteDistributions(period string) ([]DistributionResult, error) {
    results := []DistributionResult{}
    
    // Cargar reglas activas
    rules, err := de.repository.GetActiveRules(period)
    if err != nil {
        return nil, err
    }
    
    // Ordenar por prioridad (para distribuciones en cascada)
    sort.Slice(rules, func(i, j int) bool {
        return rules[i].Priority < rules[j].Priority
    })
    
    // Contexto para el DSL
    ctx := map[string]interface{}{
        "period": period,
        "metrics": de.metricsRepo,
        "costs": de.costRepo,
    }
    
    // Ejecutar cada regla
    for _, rule := range rules {
        result, err := de.dsl.ParseWithContext(rule.DSLCode, ctx)
        if err != nil {
            log.Printf("Error en regla %s: %v", rule.Name, err)
            continue
        }
        
        if distResult, ok := result.(*DistributionResult); ok {
            distResult.RuleID = rule.ID
            results = append(results, *distResult)
            
            // Actualizar contexto para cascadas
            de.applyDistributionToContext(ctx, distResult)
        }
    }
    
    return results, nil
}

// Implementación de drivers de distribución
func (de *DistributionEngine) RegisterDrivers() {
    // Driver por área (metros cuadrados)
    de.dsl.Action("area_driver", func(args ...interface{}) (interface{}, error) {
        targets := args[0].([]string)
        
        areas := make(map[string]float64)
        totalArea := 0.0
        
        for _, target := range targets {
            area := de.metricsRepo.GetArea(target)
            areas[target] = area
            totalArea += area
        }
        
        // Calcular porcentajes
        distributions := []Distribution{}
        for target, area := range areas {
            dist := Distribution{
                TargetID:    target,
                TargetType:  "cost_center",
                Percentage:  (area / totalArea) * 100,
                DriverValue: area,
                DriverUnit:  "m2",
            }
            distributions = append(distributions, dist)
        }
        
        return distributions, nil
    })
    
    // Driver por headcount
    de.dsl.Action("headcount_driver", func(args ...interface{}) (interface{}, error) {
        targets := args[0].([]string)
        weights := args[1].(map[string]float64)
        
        weightedHeadcount := 0.0
        headcounts := make(map[string]float64)
        
        for _, target := range targets {
            hc := float64(de.metricsRepo.GetHeadcount(target))
            weight := weights[target]
            if weight == 0 {
                weight = 1.0
            }
            
            weighted := hc * weight
            headcounts[target] = weighted
            weightedHeadcount += weighted
        }
        
        // Calcular distribuciones
        distributions := []Distribution{}
        for target, weighted := range headcounts {
            dist := Distribution{
                TargetID:    target,
                TargetType:  "department",
                Percentage:  (weighted / weightedHeadcount) * 100,
                DriverValue: weighted,
                DriverUnit:  "weighted_headcount",
            }
            distributions = append(distributions, dist)
        }
        
        return distributions, nil
    })
    
    // Driver ABC (Activity Based Costing)
    de.dsl.Action("abc_driver", func(args ...interface{}) (interface{}, error) {
        activities := args[0].([]string)
        costDrivers := args[1].([]string)
        products := args[2].([]string)
        
        // Calcular costo por actividad
        activityCosts := make(map[string]float64)
        for i, activity := range activities {
            driver := costDrivers[i]
            activityCosts[activity] = de.costRepo.GetActivityCost(activity, driver)
        }
        
        // Asignar a productos
        distributions := []Distribution{}
        for _, product := range products {
            totalCost := 0.0
            
            for activity, costPerDriver := range activityCosts {
                consumption := de.metricsRepo.GetActivityConsumption(product, activity)
                totalCost += costPerDriver * consumption
            }
            
            dist := Distribution{
                TargetID:        product,
                TargetType:      "product",
                AllocatedAmount: totalCost,
                DriverValue:     totalCost,
                DriverUnit:      "abc_cost",
            }
            distributions = append(distributions, dist)
        }
        
        return distributions, nil
    })
}

// Aplicar distribuciones a la contabilidad
func (de *DistributionEngine) ApplyDistributions(results []DistributionResult) error {
    for _, result := range results {
        // Crear asiento de distribución
        entry := models.JournalEntry{
            Date:        time.Now(),
            Description: fmt.Sprintf("Distribución de costos - %s", result.RuleID),
            Reference:   fmt.Sprintf("DIST-%s", result.RuleID),
            Type:        "DISTRIBUTION",
        }
        
        // Línea de crédito (origen)
        entry.Lines = append(entry.Lines, models.JournalLine{
            AccountID:    de.getDistributionAccount(result.SourceID),
            Description:  "Origen distribución",
            CreditAmount: result.SourceAmount,
        })
        
        // Líneas de débito (destinos)
        for _, dist := range result.Distributions {
            entry.Lines = append(entry.Lines, models.JournalLine{
                AccountID:    de.getDistributionAccount(dist.TargetID),
                CostCenterID: &dist.TargetID,
                Description:  fmt.Sprintf("Distribución %.2f%% - %s", dist.Percentage, dist.DriverUnit),
                DebitAmount:  dist.AllocatedAmount,
            })
        }
        
        // Guardar asiento
        if err := de.costRepo.CreateJournalEntry(&entry); err != nil {
            return err
        }
        
        // Guardar detalle de distribución para auditoría
        if err := de.saveDistributionAudit(result); err != nil {
            log.Printf("Error guardando auditoría: %v", err)
        }
    }
    
    return nil
}
```

### 4. Configuración de Métricas y Drivers

```go
// Estructura para métricas de distribución
type DistributionMetric struct {
    ID           string    `json:"id"`
    EntityID     string    `json:"entity_id"`
    EntityType   string    `json:"entity_type"`
    MetricType   string    `json:"metric_type"`
    Value        float64   `json:"value"`
    Unit         string    `json:"unit"`
    Period       string    `json:"period"`
    UpdatedAt    time.Time `json:"updated_at"`
}

// API para gestionar métricas
func (h *MetricsHandler) UpdateMetric(c *fiber.Ctx) error {
    var metric DistributionMetric
    if err := c.BodyParser(&metric); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
    }
    
    // Validar tipo de métrica
    if !isValidMetricType(metric.MetricType) {
        return c.Status(400).JSON(fiber.Map{
            "error": "Tipo de métrica inválido",
            "valid_types": getValidMetricTypes(),
        })
    }
    
    if err := h.metricsService.Update(&metric); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Error actualizando métrica"})
    }
    
    // Recalcular distribuciones afectadas
    go h.distributionEngine.RecalculateAffected(metric.EntityID, metric.Period)
    
    return c.JSON(metric)
}
```

### 5. Dashboard de Análisis de Costos

```go
func (h *CostAnalysisHandler) GetDistributionAnalysis(c *fiber.Ctx) error {
    period := c.Query("period", getCurrentPeriod())
    entityID := c.Query("entity_id")
    
    analysis := map[string]interface{}{}
    
    // Costos recibidos
    received, err := h.costService.GetReceivedCosts(entityID, period)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Error obteniendo costos recibidos"})
    }
    analysis["received_costs"] = received
    
    // Costos distribuidos
    distributed, err := h.costService.GetDistributedCosts(entityID, period)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Error obteniendo costos distribuidos"})
    }
    analysis["distributed_costs"] = distributed
    
    // Drivers utilizados
    drivers, err := h.metricsService.GetDriverAnalysis(entityID, period)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Error obteniendo análisis de drivers"})
    }
    analysis["driver_analysis"] = drivers
    
    // Tendencias
    trends, err := h.costService.GetCostTrends(entityID, 6) // últimos 6 períodos
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Error obteniendo tendencias"})
    }
    analysis["trends"] = trends
    
    return c.JSON(analysis)
}
```

## Beneficios

1. **Precisión**: Distribuciones exactas basadas en drivers reales
2. **Flexibilidad**: Fácil cambio de métodos de distribución
3. **Trazabilidad**: Registro completo del origen de cada costo
4. **Automatización**: Distribuciones ejecutadas automáticamente
5. **Análisis**: Visibilidad total de la estructura de costos

## Casos de Uso Específicos

### 1. Manufactura
```dsl
# Distribución de costos de mantenimiento a líneas de producción
distribute maintenance_costs()
  to production_lines()
  using composite_driver(
    machine_hours: 70%,
    breakdown_incidents: 30%
  )
```

### 2. Servicios Profesionales
```dsl
# Distribución de costos administrativos a proyectos
distribute admin_overhead()
  to active_projects()
  using driver("billable_hours")
  where project.status = "ACTIVE"
```

### 3. Retail
```dsl
# Distribución de costos logísticos a tiendas
distribute logistics_costs()
  to stores()
  using driver("shipment_volume")
    adjusted_by distance_factor()
```

## Métricas de Éxito

- Reducción del 80% en tiempo de cierre mensual
- 100% de trazabilidad en asignación de costos
- Mejora del 40% en precisión de costeo
- Identificación de oportunidades de ahorro por $2M/año

## Implementación en el Código Actual

### Dónde Implementar

1. **Crear nuevo paquete**: `/internal/dsl/distribution/`
   ```go
   // /internal/dsl/distribution/engine.go
   type DistributionEngine struct {
       dsl *dslbuilder.DSL
       metricsRepo *repository.MetricsRepository
       costRepo *repository.CostRepository
   }
   ```

2. **Nuevo servicio**: `/internal/services/cost_distribution_service.go`
   ```go
   type CostDistributionService struct {
       repository repository.DistributionRepository
       engine     *distribution.DistributionEngine
       journalService *JournalEntryService
   }
   ```

3. **Proceso programado**: En `main.go` o scheduler
   ```go
   // Ejecutar distribuciones al cierre de cada período
   scheduler.AddJob("0 0 L * *", func() {
       costDistService.ExecuteMonthlyDistributions()
   })
   ```

### Dónde se Llamaría

1. **Proceso de Cierre Mensual**:
   - Automáticamente después de cerrar costos directos
   - Antes de generar estados financieros

2. **API Manual**:
   - `POST /api/v1/cost-distributions/execute`
   - Para redistribuciones o ajustes

3. **Dashboard de Costos**:
   - Vista previa de distribuciones antes de aplicar

### Ventajas Específicas

1. **Precisión ABC**: Costeo real por actividad vs prorrateo simple
2. **Tiempo de Cierre**: 2 horas → 10 minutos
3. **Trazabilidad**: Saber origen de cada $ en cada producto
4. **Flexibilidad**: Cambiar drivers sin reprogramar
5. **Análisis**: Identificar productos/centros no rentables

### Integración con Modelos Existentes

**Nuevo modelo** `models/cost_center.go`:
```go
type CostCenter struct {
    ID          string  `json:"id"`
    Code        string  `json:"code"`
    Name        string  `json:"name"`
    Type        string  `json:"type"` // PRODUCTION, SUPPORT, ADMIN
    ParentID    *string `json:"parent_id"`
    Metrics     []CostCenterMetric `json:"metrics"`
}

type CostCenterMetric struct {
    CenterID    string  `json:"center_id"`
    MetricType  string  `json:"metric_type"` // area, headcount, etc
    Value       float64 `json:"value"`
    Unit        string  `json:"unit"`
    Period      string  `json:"period"`
}
```

**Modificar** `journal_line.go`:
```go
type JournalLine struct {
    // Campos existentes...
    
    // NUEVO: Para tracking de distribuciones
    CostCenterID     *string `json:"cost_center_id"`
    DistributionID   *string `json:"distribution_id"`
    AllocationFactor float64 `json:"allocation_factor"`
}
```

### Ejemplo Real: Distribución de IT

**Métrica de usuarios**:
```json
{
  "VENTAS": 50,
  "PRODUCCION": 30,
  "ADMIN": 20
}
```

**DSL Ejecuta**:
```dsl
distribute cost_center("IT")
  to departments(exclude: "IT")
  using driver("active_users") weighted
```

**Resultado**:
- VENTAS: 50% del costo IT
- PRODUCCION: 30% del costo IT
- ADMIN: 20% del costo IT

### Dashboard Propuesto

```javascript
// Nueva página: cost_distribution.html
// Mostrar:
// - Árbol de centros de costo
// - Drivers y métricas actuales
// - Preview de distribuciones
// - Histórico de distribuciones aplicadas
```

## Conclusión

La distribución automática de costos con go-dsl transforma un proceso manual y propenso a errores en un sistema preciso, auditable y flexible que se adapta a los cambios organizacionales.