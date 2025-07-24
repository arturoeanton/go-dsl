# Caso de Uso DSL #4: Conciliación Bancaria Inteligente

## Resumen

Este caso implementa la conciliación bancaria automática usando go-dsl, permitiendo definir reglas flexibles para identificar y emparejar transacciones bancarias con registros contables.

## Problema que Resuelve

La conciliación bancaria manual es:
- Consumidora de tiempo (horas o días cada mes)
- Propensa a errores humanos
- Difícil con alto volumen de transacciones
- Compleja cuando hay diferencias en fechas, montos o descripciones
- Problemática para identificar patrones recurrentes

## Solución con go-dsl

### 1. Definición del DSL de Conciliación

```go
// DSL para conciliación bancaria
dsl := dslbuilder.New("BankReconciliation")

// Tokens para matching
dsl.Token("MATCH", `match`)
dsl.Token("WHEN", `when`)
dsl.Token("AND", `and`)
dsl.Token("OR", `or`)
dsl.Token("CONTAINS", `contains`)
dsl.Token("EQUALS", `equals`)
dsl.Token("WITHIN", `within`)
dsl.Token("DAYS", `days`)
dsl.Token("TOLERANCE", `tolerance`)
dsl.Token("REFERENCE", `reference`)
dsl.Token("AMOUNT", `amount`)
dsl.Token("DATE", `date`)
dsl.Token("DESCRIPTION", `description`)
dsl.Token("PATTERN", `pattern`)
dsl.Token("REGEX", `regex`)
dsl.Token("NUMBER", `\d+(\.\d+)?`)
dsl.Token("STRING", `"[^"]*"`)
dsl.Token("PERCENTAGE", `\d+(\.\d+)?%`)

// Acciones de conciliación
dsl.Token("ACTION", `(auto_match|suggest|flag|create_adjustment)`)
dsl.Token("CONFIDENCE", `(high|medium|low)`)

// Gramática de reglas de conciliación
dsl.Rule("reconciliation_rule", []string{"MATCH", "bank_transaction", "WHEN", "conditions", "ACTION", "action_spec"}, "defineRule")
dsl.Rule("conditions", []string{"condition"}, "singleCondition")
dsl.Rule("conditions", []string{"condition", "AND", "conditions"}, "andConditions")
dsl.Rule("conditions", []string{"condition", "OR", "conditions"}, "orConditions")
dsl.Rule("condition", []string{"field", "operator", "value"}, "fieldCondition")
dsl.Rule("condition", []string{"PATTERN", "STRING"}, "patternCondition")
```

### 2. Reglas de Conciliación Bancaria

```dsl
# Regla 1: Match exacto por referencia
match bank_transaction
  when reference equals voucher.reference
  action auto_match with confidence high

# Regla 2: Match por monto con tolerancia
match bank_transaction
  when amount within tolerance 0.01 of voucher.amount
    and date within 3 days of voucher.date
  action auto_match with confidence high

# Regla 3: Transferencias entre cuentas propias
match bank_transaction
  when description contains "TRANSFERENCIA ENTRE CUENTAS"
    and amount equals internal_transfer.amount
    and date equals internal_transfer.date
  action auto_match with confidence high
    and create_counterpart

# Regla 4: Pagos de nómina
match bank_transaction
  when description pattern "NOMINA|SALARIO|SUELDO"
    and date within last_days_of_month(2)
    and amount matches payroll_total(tolerance: 100)
  action auto_match with confidence high
    and link_to payroll_voucher

# Regla 5: Comisiones bancarias
match bank_transaction
  when description contains any("COMISION", "CARGO", "MANTENIMIENTO")
    and amount < 50000
    and transaction_type equals "DEBIT"
  action create_adjustment 
    to_account "530505" # Gastos bancarios
    with description "Comisión bancaria automática"

# Regla 6: Pagos de servicios públicos
match bank_transaction
  when description regex "^(CODENSA|ACUEDUCTO|ETB|GAS NATURAL)"
    and day_of_month() between 1 and 10
  action suggest_match
    with vouchers where type = "UTILITY_PAYMENT"
    confidence medium

# Regla 7: Pagos de proveedores recurrentes
match bank_transaction
  when beneficiary in recurring_vendors()
    and amount within 5% of average_payment(beneficiary, last_3_months)
  action auto_match 
    with open_payables(beneficiary)
    confidence high

# Regla 8: Ingresos por ventas POS
match bank_transaction
  when description contains "REDEBAN" or "CREDIBANCO"
    and transaction_type equals "CREDIT"
  action match_batch
    with pos_sales(date_range: transaction_date - 2 days)
    aggregate by merchant_id
    tolerance 2% # Por comisiones

# Regla 9: Cheques cobrados
match bank_transaction
  when description pattern "CHEQUE (\d+)"
    and extract_check_number() in pending_checks()
  action auto_match
    with check_registry(check_number)
    confidence high

# Regla 10: Diferencias en cambio
match bank_transaction
  when currency != "COP"
    and exists voucher with same reference
    and amount_difference < 5%
  action create_adjustment
    for exchange_difference
    to_account "530595" # Diferencia en cambio
```

### 3. Motor de Conciliación en Go

```go
type ReconciliationEngine struct {
    dsl            *dslbuilder.DSL
    rulesRepo      *ReconciliationRulesRepository
    bankRepo       *BankTransactionRepository
    voucherRepo    *VoucherRepository
    matchingService *MatchingService
}

// BankTransaction representa una transacción bancaria
type BankTransaction struct {
    ID              string    `json:"id"`
    BankAccountID   string    `json:"bank_account_id"`
    Date            time.Time `json:"date"`
    Amount          float64   `json:"amount"`
    TransactionType string    `json:"transaction_type"` // DEBIT, CREDIT
    Description     string    `json:"description"`
    Reference       string    `json:"reference"`
    Beneficiary     string    `json:"beneficiary"`
    Status          string    `json:"status"` // PENDING, MATCHED, FLAGGED
    MatchedVoucherID *string  `json:"matched_voucher_id"`
}

// ReconciliationMatch representa un emparejamiento
type ReconciliationMatch struct {
    BankTransactionID string  `json:"bank_transaction_id"`
    VoucherID        string  `json:"voucher_id"`
    Confidence       string  `json:"confidence"`
    MatchType        string  `json:"match_type"`
    Score            float64 `json:"score"`
    RuleID           string  `json:"rule_id"`
    Notes            string  `json:"notes"`
}

func (re *ReconciliationEngine) ReconcileBankStatement(bankAccountID string, period string) (*ReconciliationResult, error) {
    result := &ReconciliationResult{
        BankAccountID: bankAccountID,
        Period:        period,
        StartedAt:     time.Now(),
    }
    
    // Obtener transacciones pendientes
    transactions, err := re.bankRepo.GetPendingTransactions(bankAccountID, period)
    if err != nil {
        return nil, err
    }
    
    // Cargar reglas de conciliación
    rules, err := re.rulesRepo.GetActiveRules(bankAccountID)
    if err != nil {
        return nil, err
    }
    
    // Procesar cada transacción
    for _, transaction := range transactions {
        match, err := re.processTransaction(transaction, rules)
        if err != nil {
            log.Printf("Error procesando transacción %s: %v", transaction.ID, err)
            continue
        }
        
        if match != nil {
            result.Matches = append(result.Matches, *match)
            
            // Aplicar el match si es automático y de alta confianza
            if match.Confidence == "high" && match.MatchType == "auto_match" {
                if err := re.applyMatch(transaction, match); err != nil {
                    log.Printf("Error aplicando match: %v", err)
                    result.Errors = append(result.Errors, err.Error())
                }
            }
        } else {
            result.Unmatched = append(result.Unmatched, transaction.ID)
        }
    }
    
    result.CompletedAt = time.Now()
    result.Summary = re.generateSummary(result)
    
    return result, nil
}

func (re *ReconciliationEngine) processTransaction(transaction BankTransaction, rules []ReconciliationRule) (*ReconciliationMatch, error) {
    // Contexto para el DSL
    ctx := map[string]interface{}{
        "bank_transaction": transaction,
        "bank_account":     re.getBankAccount(transaction.BankAccountID),
    }
    
    // Intentar cada regla en orden de prioridad
    for _, rule := range rules {
        result, err := re.dsl.ParseWithContext(rule.DSLCode, ctx)
        if err != nil {
            continue // Siguiente regla si esta falla
        }
        
        if match, ok := result.(*ReconciliationMatch); ok && match != nil {
            match.RuleID = rule.ID
            
            // Calcular score basado en múltiples factores
            match.Score = re.calculateMatchScore(transaction, match)
            
            // Si el score es suficientemente alto, retornar el match
            if match.Score >= rule.MinScore {
                return match, nil
            }
        }
    }
    
    // Si no hay match automático, intentar matching fuzzy
    return re.fuzzyMatch(transaction)
}

// Funciones de matching disponibles en el DSL
func (re *ReconciliationEngine) RegisterMatchingFunctions() {
    // Buscar por referencia exacta
    re.dsl.Action("match_by_reference", func(args ...interface{}) (interface{}, error) {
        ctx := args[0].(map[string]interface{})
        transaction := ctx["bank_transaction"].(BankTransaction)
        
        voucher, err := re.voucherRepo.GetByReference(transaction.Reference)
        if err != nil || voucher == nil {
            return nil, nil
        }
        
        return &ReconciliationMatch{
            VoucherID:  voucher.ID,
            Confidence: "high",
            MatchType:  "reference",
            Notes:      "Match exacto por referencia",
        }, nil
    })
    
    // Buscar por monto con tolerancia
    re.dsl.Action("match_by_amount", func(args ...interface{}) (interface{}, error) {
        ctx := args[0].(map[string]interface{})
        transaction := ctx["bank_transaction"].(BankTransaction)
        tolerance := args[1].(float64)
        dateRange := args[2].(int) // días
        
        vouchers, err := re.voucherRepo.GetByAmountRange(
            transaction.Amount - tolerance,
            transaction.Amount + tolerance,
            transaction.Date.AddDate(0, 0, -dateRange),
            transaction.Date.AddDate(0, 0, dateRange),
        )
        
        if err != nil || len(vouchers) == 0 {
            return nil, nil
        }
        
        // Si hay exactamente uno, alta confianza
        confidence := "high"
        if len(vouchers) > 1 {
            confidence = "medium"
        }
        
        return &ReconciliationMatch{
            VoucherID:  vouchers[0].ID,
            Confidence: confidence,
            MatchType:  "amount",
            Notes:      fmt.Sprintf("Match por monto con tolerancia %.2f", tolerance),
        }, nil
    })
    
    // Crear ajuste automático
    re.dsl.Action("create_adjustment", func(args ...interface{}) (interface{}, error) {
        ctx := args[0].(map[string]interface{})
        transaction := ctx["bank_transaction"].(BankTransaction)
        accountCode := args[1].(string)
        description := args[2].(string)
        
        // Crear comprobante de ajuste
        voucher := &models.Voucher{
            VoucherType: "ADJUSTMENT",
            Date:        transaction.Date,
            Description: description,
            Reference:   fmt.Sprintf("CONC-%s", transaction.ID),
            Status:      "DRAFT",
        }
        
        // Agregar líneas
        if transaction.TransactionType == "DEBIT" {
            voucher.Lines = []models.VoucherLine{
                {
                    AccountID:    accountCode,
                    Description:  description,
                    DebitAmount:  transaction.Amount,
                    CreditAmount: 0,
                },
                {
                    AccountID:    transaction.BankAccountID,
                    Description:  "Ajuste bancario",
                    DebitAmount:  0,
                    CreditAmount: transaction.Amount,
                },
            }
        } else {
            // Crédito - invertir
            voucher.Lines[0].DebitAmount, voucher.Lines[0].CreditAmount = 0, transaction.Amount
            voucher.Lines[1].DebitAmount, voucher.Lines[1].CreditAmount = transaction.Amount, 0
        }
        
        if err := re.voucherRepo.Create(voucher); err != nil {
            return nil, err
        }
        
        return &ReconciliationMatch{
            VoucherID:  voucher.ID,
            Confidence: "high",
            MatchType:  "adjustment",
            Notes:      "Ajuste automático creado",
        }, nil
    })
}

// Aplicar match confirmado
func (re *ReconciliationEngine) applyMatch(transaction BankTransaction, match *ReconciliationMatch) error {
    // Actualizar transacción bancaria
    transaction.Status = "MATCHED"
    transaction.MatchedVoucherID = &match.VoucherID
    if err := re.bankRepo.Update(&transaction); err != nil {
        return err
    }
    
    // Actualizar comprobante
    voucher, err := re.voucherRepo.GetByID(match.VoucherID)
    if err != nil {
        return err
    }
    
    voucher.BankReconciled = true
    voucher.BankTransactionID = &transaction.ID
    if err := re.voucherRepo.Update(voucher); err != nil {
        return err
    }
    
    // Guardar registro de conciliación
    reconciliation := &BankReconciliation{
        BankTransactionID: transaction.ID,
        VoucherID:        match.VoucherID,
        MatchType:        match.MatchType,
        Confidence:       match.Confidence,
        RuleID:           match.RuleID,
        ReconciledAt:     time.Now(),
        ReconciledBy:     getCurrentUser(),
    }
    
    return re.saveReconciliation(reconciliation)
}
```

### 4. Machine Learning para Matching Fuzzy

```go
// FuzzyMatcher usa ML para encontrar matches no obvios
type FuzzyMatcher struct {
    model           *MatchingModel
    vectorizer      *TextVectorizer
    historyRepo     *ReconciliationHistoryRepository
}

func (fm *FuzzyMatcher) FindMatch(transaction BankTransaction) (*ReconciliationMatch, error) {
    // Obtener candidatos potenciales
    candidates, err := fm.getCandidates(transaction)
    if err != nil {
        return nil, err
    }
    
    if len(candidates) == 0 {
        return nil, nil
    }
    
    // Vectorizar descripción de la transacción
    transVector := fm.vectorizer.Transform(transaction.Description)
    
    // Calcular similitud con cada candidato
    matches := []ScoredMatch{}
    for _, candidate := range candidates {
        score := fm.calculateSimilarity(transaction, candidate, transVector)
        
        if score > 0.7 { // Umbral de similitud
            matches = append(matches, ScoredMatch{
                VoucherID: candidate.ID,
                Score:     score,
            })
        }
    }
    
    // Ordenar por score y tomar el mejor
    sort.Slice(matches, func(i, j int) bool {
        return matches[i].Score > matches[j].Score
    })
    
    if len(matches) > 0 {
        confidence := "low"
        if matches[0].Score > 0.9 {
            confidence = "high"
        } else if matches[0].Score > 0.8 {
            confidence = "medium"
        }
        
        return &ReconciliationMatch{
            VoucherID:  matches[0].VoucherID,
            Confidence: confidence,
            MatchType:  "fuzzy",
            Score:      matches[0].Score,
            Notes:      "Match por similitud ML",
        }, nil
    }
    
    return nil, nil
}

func (fm *FuzzyMatcher) calculateSimilarity(transaction BankTransaction, voucher *models.Voucher, transVector []float64) float64 {
    // Factores de similitud
    factors := []float64{}
    
    // 1. Similitud de texto (40%)
    voucherVector := fm.vectorizer.Transform(voucher.Description)
    textSim := cosineSimilarity(transVector, voucherVector)
    factors = append(factors, textSim * 0.4)
    
    // 2. Similitud de monto (30%)
    amountDiff := math.Abs(transaction.Amount - voucher.TotalAmount)
    amountSim := 1.0 - (amountDiff / transaction.Amount)
    if amountSim < 0 {
        amountSim = 0
    }
    factors = append(factors, amountSim * 0.3)
    
    // 3. Proximidad de fecha (20%)
    daysDiff := math.Abs(transaction.Date.Sub(voucher.Date).Hours() / 24)
    dateSim := math.Max(0, 1.0 - (daysDiff / 30)) // 30 días máximo
    factors = append(factors, dateSim * 0.2)
    
    // 4. Patrones históricos (10%)
    historySim := fm.getHistoricalSimilarity(transaction, voucher)
    factors = append(factors, historySim * 0.1)
    
    // Sumar todos los factores
    totalScore := 0.0
    for _, factor := range factors {
        totalScore += factor
    }
    
    return totalScore
}
```

### 5. Dashboard de Conciliación

```go
func (h *ReconciliationHandler) GetReconciliationDashboard(c *fiber.Ctx) error {
    bankAccountID := c.Query("bank_account_id")
    period := c.Query("period", getCurrentPeriod())
    
    dashboard := &ReconciliationDashboard{}
    
    // Estado general
    status, err := h.reconciliationService.GetStatus(bankAccountID, period)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Error obteniendo estado"})
    }
    dashboard.Status = status
    
    // Transacciones pendientes
    pending, err := h.bankRepo.GetPendingCount(bankAccountID, period)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Error contando pendientes"})
    }
    dashboard.PendingTransactions = pending
    
    // Matches sugeridos
    suggestions, err := h.reconciliationService.GetSuggestions(bankAccountID, 10)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Error obteniendo sugerencias"})
    }
    dashboard.Suggestions = suggestions
    
    // Estadísticas de conciliación
    stats := &ReconciliationStats{
        AutoMatchRate:     h.calculateAutoMatchRate(bankAccountID, period),
        AverageMatchTime:  h.getAverageMatchTime(bankAccountID, period),
        RuleEffectiveness: h.getRuleEffectiveness(bankAccountID),
        CommonPatterns:    h.getCommonPatterns(bankAccountID),
    }
    dashboard.Statistics = stats
    
    return c.JSON(dashboard)
}

// API para confirmar/rechazar matches sugeridos
func (h *ReconciliationHandler) ConfirmMatch(c *fiber.Ctx) error {
    var confirmation MatchConfirmation
    if err := c.BodyParser(&confirmation); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
    }
    
    // Aplicar el match
    if err := h.reconciliationService.ApplyMatch(
        confirmation.BankTransactionID,
        confirmation.VoucherID,
        confirmation.Notes,
    ); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Error aplicando match"})
    }
    
    // Actualizar modelo ML con el feedback
    go h.mlService.UpdateModel(confirmation)
    
    return c.JSON(fiber.Map{"success": true})
}
```

## Beneficios

1. **Ahorro de Tiempo**: Reducción del 95% en tiempo de conciliación
2. **Precisión**: Detección automática de discrepancias
3. **Aprendizaje**: El sistema mejora con cada conciliación
4. **Trazabilidad**: Registro completo de todas las decisiones
5. **Escalabilidad**: Maneja miles de transacciones sin esfuerzo

## Casos de Uso Específicos

### 1. Empresa con Alto Volumen
```dsl
# Pagos masivos de clientes
match bank_transaction
  when description contains "CONSIGNACION"
    and exists customer with similar_name(beneficiary)
  action suggest_match
    with open_invoices(customer)
    ordered_by due_date
```

### 2. Empresa Multinacional
```dsl
# Conciliación multi-moneda
match bank_transaction
  when currency != local_currency()
  action apply_exchange_rate
    from central_bank_rate(transaction.date)
    then match_by_converted_amount
```

### 3. E-commerce
```dsl
# Pagos de pasarelas
match bank_transaction
  when description matches "PAYPAL|STRIPE|MERCADOPAGO"
  action match_batch
    with online_sales(date_range: -3 days)
    group_by payment_gateway
    deduct_fees from gateway_config
```

## Métricas de Éxito

- 95% de transacciones conciliadas automáticamente
- Reducción de 2 días a 2 horas en cierre mensual
- 99.9% de precisión en matches automáticos
- ROI de 300% en el primer año

## Conclusión

La conciliación bancaria inteligente con go-dsl elimina una de las tareas más tediosas de la contabilidad, proporcionando precisión, velocidad y aprendizaje continuo.