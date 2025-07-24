// DSL Editor JavaScript - Motor Contable
// Version: 1.0
// Last Updated: 2024-01-15

// Global state
const state = {
    templates: [],
    selectedTemplate: null,
    isDirty: false,
    currentCode: '',
    syntaxErrors: [],
    testData: {}
};

// DSL Templates Examples
const dslTemplates = [
    {
        id: 'tpl-001',
        name: 'invoice_sale_co',
        description: 'Factura de venta para Colombia con IVA 19%',
        type: 'invoice_sale',
        country: 'CO',
        version: '1.0',
        is_active: true,
        created_at: '2024-01-01T10:00:00Z',
        updated_at: '2024-01-20T15:30:00Z',
        code: `template invoice_sale_co {
  // Definir variables desde el comprobante
  let subtotal = voucher.metadata.subtotal
  let tax_rate = 0.19
  let tax_amount = voucher.metadata.taxes.iva_19
  let total = voucher.total_amount
  
  // Validaciones
  require subtotal > 0 : "Subtotal debe ser positivo"
  require tax_amount == subtotal * tax_rate : "IVA calculado incorrectamente"
  require total == subtotal + tax_amount : "Total no cuadra"
  
  // Generar asiento contable
  entry {
    // Debitar cuentas por cobrar
    debit account("1305.05") amount(total) {
      description = "Factura " + voucher.voucher_number
      metadata.customer_id = voucher.metadata.customer.id
      metadata.due_date = date_add(voucher.voucher_date, 30, "days")
    }
    
    // Acreditar ventas
    credit account("4135.05") amount(subtotal) {
      description = "Venta de servicios"
      cost_center = "CC-001"
    }
    
    // Acreditar IVA
    credit account("2408.01") amount(tax_amount) {
      description = "IVA 19% por pagar"
      metadata.tax_period = format_date(voucher.voucher_date, "YYYY-MM")
    }
  }
}`
    },
    {
        id: 'tpl-002',
        name: 'invoice_purchase_co',
        description: 'Factura de compra con retenciones',
        type: 'invoice_purchase',
        country: 'CO',
        version: '1.0',
        is_active: true,
        created_at: '2024-01-05T09:00:00Z',
        updated_at: '2024-01-15T14:20:00Z',
        code: `template invoice_purchase_co {
  // Variables
  let subtotal = voucher.metadata.subtotal
  let iva = voucher.metadata.taxes.iva_19
  let retefuente = subtotal * 0.025  // 2.5% retención
  let reteiva = iva * 0.15  // 15% de retención sobre IVA
  let total = subtotal + iva
  let total_pagar = total - retefuente - reteiva
  
  // Validaciones
  require subtotal > 0 : "Subtotal inválido"
  require iva >= 0 : "IVA no puede ser negativo"
  
  // Asiento contable
  entry {
    // Debitar gastos o inventario
    if voucher.metadata.expense_type == "inventory" {
      debit account("1435.05") amount(subtotal) {
        description = "Compra de inventario"
      }
    } else {
      debit account("5135.30") amount(subtotal) {
        description = "Gasto " + voucher.description
        cost_center = voucher.metadata.cost_center
      }
    }
    
    // Debitar IVA descontable
    debit account("2408.05") amount(iva) {
      description = "IVA descontable 19%"
    }
    
    // Acreditar proveedor
    credit account("2205.01") amount(total_pagar) {
      description = "Por pagar a " + voucher.metadata.supplier.name
      metadata.supplier_id = voucher.metadata.supplier.id
      metadata.due_date = date_add(voucher.voucher_date, 30, "days")
    }
    
    // Acreditar retenciones
    credit account("2365.25") amount(retefuente) {
      description = "Retención en la fuente 2.5%"
    }
    
    credit account("2365.30") amount(reteiva) {
      description = "Retención de IVA 15%"
    }
  }
}`
    },
    {
        id: 'tpl-003',
        name: 'payment_bank_transfer',
        description: 'Pago por transferencia bancaria',
        type: 'payment',
        country: 'ALL',
        version: '1.0',
        is_active: true,
        created_at: '2024-01-10T11:00:00Z',
        updated_at: '2024-01-18T09:15:00Z',
        code: `template payment_bank_transfer {
  let amount = voucher.total_amount
  let payment_method = voucher.metadata.payment_method
  let reference = voucher.metadata.reference_number
  
  // Validaciones
  require amount > 0 : "Monto debe ser positivo"
  require payment_method == "bank_transfer" : "Método de pago incorrecto"
  require reference != "" : "Número de referencia requerido"
  
  // Generar asiento
  entry {
    // Debitar cuenta por pagar
    debit account("2205.01") amount(amount) {
      description = "Pago a proveedor ref: " + reference
      metadata.supplier_id = voucher.metadata.supplier.id
      metadata.payment_ref = reference
    }
    
    // Acreditar banco
    credit account("1110.05") amount(amount) {
      description = "Transferencia " + reference
      metadata.bank_account = voucher.metadata.bank_account
    }
  }
  
  // Post-procesamiento
  after {
    // Actualizar saldo del proveedor
    update_supplier_balance(voucher.metadata.supplier.id, -amount)
    
    // Notificar pago realizado
    notify("payment_completed", {
      supplier: voucher.metadata.supplier.name,
      amount: amount,
      reference: reference
    })
  }
}`
    },
    {
        id: 'tpl-004',
        name: 'receipt_cash',
        description: 'Recibo de caja por cobro a clientes',
        type: 'receipt',
        country: 'ALL',
        version: '1.0',
        is_active: true,
        created_at: '2024-01-12T08:30:00Z',
        updated_at: '2024-01-12T08:30:00Z',
        code: `template receipt_cash {
  let amount = voucher.total_amount
  let customer = voucher.metadata.customer
  
  // Asiento simple
  entry {
    // Debitar caja
    debit account("1105.05") amount(amount) {
      description = "Recibo de caja " + voucher.voucher_number
    }
    
    // Acreditar cliente
    credit account("1305.05") amount(amount) {
      description = "Abono cliente " + customer.name
      metadata.customer_id = customer.id
      metadata.invoice_refs = voucher.metadata.invoices_paid
    }
  }
}`
    },
    {
        id: 'tpl-005',
        name: 'payroll_monthly',
        description: 'Nómina mensual con prestaciones',
        type: 'custom',
        country: 'CO',
        version: '2.0',
        is_active: true,
        created_at: '2024-01-03T12:00:00Z',
        updated_at: '2024-01-19T16:45:00Z',
        code: `template payroll_monthly {
  // Calcular componentes de nómina
  let basic_salary = voucher.metadata.basic_salary
  let transport_allowance = voucher.metadata.transport_allowance
  let total_earnings = basic_salary + transport_allowance
  
  // Deducciones empleado
  let health_employee = basic_salary * 0.04
  let pension_employee = basic_salary * 0.04
  let total_deductions = health_employee + pension_employee
  
  // Aportes empresa
  let health_company = basic_salary * 0.085
  let pension_company = basic_salary * 0.12
  let arl = basic_salary * 0.00522
  let sena = basic_salary * 0.02
  let icbf = basic_salary * 0.03
  let ccf = basic_salary * 0.04
  
  // Prestaciones sociales
  let cesantias = total_earnings * 0.0833
  let intereses_cesantias = cesantias * 0.01
  let prima = total_earnings * 0.0833
  let vacaciones = basic_salary * 0.0417
  
  let net_pay = total_earnings - total_deductions
  
  // Validaciones
  require basic_salary > 0 : "Salario básico inválido"
  require net_pay > 0 : "Pago neto debe ser positivo"
  
  // Asientos contables
  entry {
    // Gastos de nómina
    debit account("5105.06") amount(basic_salary) {
      description = "Sueldos y salarios"
      cost_center = voucher.metadata.cost_center
    }
    
    debit account("5105.15") amount(transport_allowance) {
      description = "Auxilio de transporte"
      cost_center = voucher.metadata.cost_center
    }
    
    // Aportes patronales
    debit account("5105.30") amount(health_company + pension_company + arl) {
      description = "Aportes seguridad social empresa"
    }
    
    debit account("5105.33") amount(sena + icbf + ccf) {
      description = "Aportes parafiscales"
    }
    
    // Prestaciones sociales
    debit account("5105.36") amount(cesantias + intereses_cesantias + prima + vacaciones) {
      description = "Prestaciones sociales"
    }
    
    // Pasivos
    credit account("2505.05") amount(net_pay) {
      description = "Salarios por pagar"
      metadata.employee_id = voucher.metadata.employee.id
    }
    
    credit account("2370.05") amount(health_employee + pension_employee) {
      description = "Retenciones empleados"
    }
    
    credit account("2370.25") amount(health_company + pension_company + arl) {
      description = "Seguridad social por pagar"
    }
    
    credit account("2370.30") amount(sena + icbf + ccf) {
      description = "Parafiscales por pagar"
    }
    
    credit account("2610.05") amount(cesantias) {
      description = "Cesantías consolidadas"
    }
    
    credit account("2610.10") amount(intereses_cesantias) {
      description = "Intereses sobre cesantías"
    }
    
    credit account("2610.15") amount(prima) {
      description = "Prima de servicios"
    }
    
    credit account("2610.20") amount(vacaciones) {
      description = "Vacaciones consolidadas"
    }
  }
}`
    }
];

// Test voucher data
const testVouchers = {
    invoice_sale_1: {
        voucher_number: "FV-2024-0150",
        voucher_date: "2024-01-21",
        total_amount: 1190000,
        metadata: {
            customer: {
                id: "cust-001",
                name: "Cliente Ejemplo S.A.S",
                tax_id: "900123456-7"
            },
            subtotal: 1000000,
            taxes: {
                iva_19: 190000
            },
            items: [
                {
                    description: "Servicio de consultoría",
                    quantity: 1,
                    unit_price: 1000000,
                    total: 1000000
                }
            ]
        }
    },
    invoice_sale_2: {
        voucher_number: "FV-2024-0151",
        voucher_date: "2024-01-21",
        total_amount: 4284000,
        metadata: {
            customer: {
                id: "cust-002",
                name: "Distribuidora Nacional",
                tax_id: "860456789-1"
            },
            subtotal: 3600000,
            taxes: {
                iva_19: 684000
            },
            items: [
                {
                    description: "Producto A",
                    quantity: 10,
                    unit_price: 200000,
                    total: 2000000
                },
                {
                    description: "Producto B",
                    quantity: 8,
                    unit_price: 200000,
                    total: 1600000
                }
            ]
        }
    },
    invoice_purchase_1: {
        voucher_number: "FC-2024-0089",
        voucher_date: "2024-01-21",
        total_amount: 1190000,
        metadata: {
            supplier: {
                id: "supp-001",
                name: "Proveedor XYZ Ltda",
                tax_id: "800789456-2"
            },
            subtotal: 1000000,
            taxes: {
                iva_19: 190000
            },
            expense_type: "services",
            cost_center: "CC-002"
        }
    },
    payment_1: {
        voucher_number: "CE-2024-0045",
        voucher_date: "2024-01-21",
        total_amount: 500000,
        metadata: {
            payment_method: "cash",
            supplier: {
                id: "supp-001",
                name: "Proveedor XYZ Ltda"
            }
        }
    },
    payment_2: {
        voucher_number: "CE-2024-0046",
        voucher_date: "2024-01-21",
        total_amount: 1500000,
        metadata: {
            payment_method: "bank_transfer",
            reference_number: "REF-123456",
            bank_account: "001-123456-78",
            supplier: {
                id: "supp-002",
                name: "Servicios ABC S.A."
            }
        }
    }
};

// Helper functions
function showLoading(message = 'Cargando...') {
    const overlay = document.getElementById('loadingOverlay');
    if (overlay) {
        overlay.querySelector('p').textContent = message;
        overlay.style.display = 'flex';
    }
}

function hideLoading() {
    const overlay = document.getElementById('loadingOverlay');
    if (overlay) {
        overlay.style.display = 'none';
    }
}

function showError(message) {
    if (window.utils && window.utils.toast) {
        window.utils.toast.error(message);
    } else {
        alert(message);
    }
}

// Initialize on DOM ready
document.addEventListener('DOMContentLoaded', () => {
    initializeDSLEditor();
});

/**
 * Initialize DSL editor
 */
function initializeDSLEditor() {
    setupEventListeners();
    loadTemplates();
    setupCodeEditor();
}

/**
 * Setup event listeners
 */
function setupEventListeners() {
    // Code editor
    const codeEditor = document.getElementById('codeEditor');
    if (codeEditor) {
        codeEditor.addEventListener('input', handleCodeChange);
        codeEditor.addEventListener('keydown', handleKeyDown);
        codeEditor.addEventListener('scroll', syncLineNumbers);
    }
    
    // Search
    document.getElementById('searchInput').addEventListener('keyup', debounce(filterTemplates, 300));
    
    // Hide snippet menu on click outside
    document.addEventListener('click', (e) => {
        if (!e.target.closest('#snippetMenu')) {
            document.getElementById('snippetMenu').style.display = 'none';
        }
    });
}

/**
 * Load templates
 */
async function loadTemplates() {
    try {
        showLoading('Cargando plantillas...');
        
        // Fetch templates from API
        const result = await motorContableApi.dsl.getTemplates();
        
        // Temporarily keep the hardcoded template code for display
        const templates = Array.isArray(result.data) ? result.data : (result.data.templates || []);
        const templatesWithCode = templates.map(t => {
            const hardcoded = dslTemplates.find(dt => dt.id === t.id);
            return {
                ...t,
                code: hardcoded ? hardcoded.code : '// Template code not available'
            };
        });
        
        state.templates = templatesWithCode;
        renderTemplateList();
        
    } catch (error) {
        console.error('Error loading templates:', error);
        showError('Error al cargar las plantillas DSL');
        
        // Fallback to hardcoded templates
        state.templates = dslTemplates;
        renderTemplateList();
    } finally {
        hideLoading();
    }
}

/**
 * Render template list
 */
function renderTemplateList() {
    const templateList = document.getElementById('templateList');
    templateList.innerHTML = '';
    
    const filteredTemplates = filterTemplatesByUI();
    
    filteredTemplates.forEach(template => {
        const item = document.createElement('div');
        item.className = `template-item ${state.selectedTemplate?.id === template.id ? 'active' : ''}`;
        item.onclick = () => selectTemplate(template);
        
        item.innerHTML = `
            <div style="display: flex; justify-content: space-between; align-items: start;">
                <div>
                    <h4 style="margin: 0 0 5px 0;">${template.name}</h4>
                    <p style="margin: 0; color: var(--text-muted); font-size: 0.9rem;">${template.description}</p>
                </div>
                <span class="badge badge-${template.is_active ? 'success' : 'secondary'}">
                    ${template.is_active ? 'Activa' : 'Inactiva'}
                </span>
            </div>
            <div class="template-meta">
                <span>📄 ${getVoucherTypeLabel(template.type)}</span>
                <span>🌍 ${template.country}</span>
                <span>📌 v${template.version}</span>
                <span>📅 ${formatDate(template.updated_at)}</span>
            </div>
        `;
        
        templateList.appendChild(item);
    });
}

/**
 * Filter templates by UI filters
 */
function filterTemplatesByUI() {
    const typeFilter = document.getElementById('filterType')?.value || '';
    const countryFilter = document.getElementById('filterCountry')?.value || '';
    const searchTerm = document.getElementById('searchInput')?.value.toLowerCase() || '';
    
    return state.templates.filter(template => {
        if (typeFilter && template.type !== typeFilter) return false;
        if (countryFilter && template.country !== countryFilter && template.country !== 'ALL') return false;
        if (searchTerm && !template.name.toLowerCase().includes(searchTerm) && 
            !template.description.toLowerCase().includes(searchTerm)) return false;
        return true;
    });
}

/**
 * Filter templates
 */
function filterTemplates() {
    renderTemplateList();
}

/**
 * Select template
 */
function selectTemplate(template) {
    if (state.isDirty && !confirm('¿Descartar cambios no guardados?')) {
        return;
    }
    
    state.selectedTemplate = template;
    state.currentCode = template.code;
    state.isDirty = false;
    
    // Update UI
    document.getElementById('editorCard').style.display = 'block';
    document.getElementById('editorTitle').textContent = `Editor DSL - ${template.name}`;
    document.getElementById('templateVersion').textContent = `v${template.version}`;
    
    // Load code
    const codeEditor = document.getElementById('codeEditor');
    codeEditor.textContent = template.code;
    
    // Update line numbers only (syntax highlighting disabled)
    updateLineNumbers();
    
    // Update template list
    renderTemplateList();
    
    // Clear test data
    document.getElementById('testVoucher').value = '';
    document.getElementById('testDataJson').value = '';
}

/**
 * Setup code editor
 */
function setupCodeEditor() {
    const codeEditor = document.getElementById('codeEditor');
    
    // Prevent formatting on paste
    codeEditor.addEventListener('paste', (e) => {
        e.preventDefault();
        const text = e.clipboardData.getData('text/plain');
        document.execCommand('insertText', false, text);
    });
}

/**
 * Handle code change
 */
function handleCodeChange(event) {
    state.isDirty = true;
    state.currentCode = event.target.textContent;
    
    updateLineNumbers();
    updateCursorPosition();
    
    // Syntax highlighting disabled
}

/**
 * Handle key down
 */
function handleKeyDown(event) {
    // Tab handling
    if (event.key === 'Tab') {
        event.preventDefault();
        document.execCommand('insertText', false, '  ');
    }
    
    // Auto-indent on Enter
    if (event.key === 'Enter') {
        event.preventDefault();
        const selection = window.getSelection();
        const range = selection.getRangeAt(0);
        const currentLine = getCurrentLine(range);
        const indent = getLineIndent(currentLine);
        
        let newIndent = indent;
        if (currentLine.trim().endsWith('{')) {
            newIndent += '  ';
        }
        
        document.execCommand('insertText', false, '\n' + newIndent);
    }
}

/**
 * Update syntax highlighting - DISABLED
 * Keeping plain text for better stability
 */
function updateSyntaxHighlighting() {
    // Syntax highlighting disabled - showing plain text
    return;
}

/**
 * Update line numbers
 */
function updateLineNumbers() {
    const codeEditor = document.getElementById('codeEditor');
    const lineNumbers = document.getElementById('lineNumbers');
    const lines = codeEditor.textContent.split('\n');
    
    lineNumbers.textContent = lines.map((_, i) => i + 1).join('\n');
}

/**
 * Sync line numbers scroll
 */
function syncLineNumbers() {
    const codeEditor = document.getElementById('codeEditor');
    const lineNumbers = document.getElementById('lineNumbers');
    lineNumbers.scrollTop = codeEditor.scrollTop;
}

/**
 * Update cursor position
 */
function updateCursorPosition() {
    const selection = window.getSelection();
    if (selection.rangeCount === 0) return;
    
    const range = selection.getRangeAt(0);
    const text = document.getElementById('codeEditor').textContent;
    const beforeCursor = text.substring(0, range.startOffset);
    const lines = beforeCursor.split('\n');
    const line = lines.length;
    const col = lines[lines.length - 1].length + 1;
    
    document.getElementById('cursorPosition').textContent = `Línea ${line}, Col ${col}`;
}

/**
 * Validate syntax
 */
async function validateSyntax() {
    const code = state.currentCode || document.getElementById('codeEditor').textContent;
    
    showLoading('Validando sintaxis...');
    
    try {
        // Simulate syntax validation
        await new Promise(resolve => setTimeout(resolve, 500));
        
        // Simple validation rules
        const errors = [];
        
        // Check for template declaration
        if (!code.includes('template')) {
            errors.push({
                line: 1,
                message: 'Falta declaración de template'
            });
        }
        
        // Check for matching braces
        const openBraces = (code.match(/{/g) || []).length;
        const closeBraces = (code.match(/}/g) || []).length;
        if (openBraces !== closeBraces) {
            errors.push({
                line: 0,
                message: 'Llaves desbalanceadas'
            });
        }
        
        // Check for entry block
        if (!code.includes('entry {')) {
            errors.push({
                line: 0,
                message: 'Falta bloque entry para generar asiento'
            });
        }
        
        // Update feedback
        const feedback = document.getElementById('syntaxFeedback');
        if (errors.length === 0) {
            feedback.innerHTML = '<div class="syntax-valid">✅ Sintaxis válida</div>';
        } else {
            feedback.innerHTML = errors.map(err => 
                `<div class="syntax-error">❌ ${err.line ? `Línea ${err.line}: ` : ''}${err.message}</div>`
            ).join('');
        }
        
        state.syntaxErrors = errors;
        
    } catch (error) {
        showError('Error al validar sintaxis');
    } finally {
        hideLoading();
    }
}

/**
 * Format code
 */
function formatCode() {
    const codeEditor = document.getElementById('codeEditor');
    let code = codeEditor.textContent;
    
    // Simple formatting
    const lines = code.split('\n');
    let indentLevel = 0;
    const formatted = [];
    
    lines.forEach(line => {
        const trimmed = line.trim();
        
        // Decrease indent for closing braces
        if (trimmed === '}' || trimmed.startsWith('}')) {
            indentLevel = Math.max(0, indentLevel - 1);
        }
        
        // Add formatted line
        if (trimmed) {
            formatted.push('  '.repeat(indentLevel) + trimmed);
        } else {
            formatted.push('');
        }
        
        // Increase indent for opening braces
        if (trimmed.endsWith('{')) {
            indentLevel++;
        }
    });
    
    codeEditor.textContent = formatted.join('\n');
    // Syntax highlighting disabled
    updateLineNumbers();
}

/**
 * Show snippets menu
 */
function showSnippets() {
    const menu = document.getElementById('snippetMenu');
    menu.style.display = 'block';
    
    // Position near button
    const btn = event.target;
    const rect = btn.getBoundingClientRect();
    menu.style.left = rect.left + 'px';
    menu.style.top = (rect.bottom + 5) + 'px';
}

/**
 * Insert snippet
 */
function insertSnippet(type) {
    const snippets = {
        validation: `// Validación
require variable > 0 : "Mensaje de error"`,
        entry: `// Asiento contable
entry {
  debit account("1105.05") amount(100) {
    description = "Descripción"
  }
  
  credit account("2105.05") amount(100) {
    description = "Descripción"
  }
}`,
        loop: `// Bucle de items
for item in voucher.metadata.items {
  debit account(item.account) amount(item.total) {
    description = item.description
  }
}`,
        conditional: `// Condicional
if condition {
  // código si verdadero
} else {
  // código si falso
}`,
        tax: `// Cálculo de impuesto
let tax_base = subtotal
let tax_rate = 0.19
let tax_amount = tax_calculate(tax_base, tax_rate)`,
        notification: `// Notificación
after {
  notify("event_type", {
    key: value
  })
}`
    };
    
    const snippet = snippets[type];
    if (snippet) {
        document.execCommand('insertText', false, '\n' + snippet + '\n');
        document.getElementById('snippetMenu').style.display = 'none';
    }
}

/**
 * Show variables modal
 */
function showVariables() {
    document.getElementById('variablesModal').style.display = 'block';
}

/**
 * Close variables modal
 */
function closeVariablesModal() {
    document.getElementById('variablesModal').style.display = 'none';
}

/**
 * Show functions reference
 */
function showFunctions() {
    alert('Referencia de funciones próximamente');
}

/**
 * Save template
 */
async function saveTemplate() {
    if (!state.selectedTemplate) return;
    
    // Validate first
    await validateSyntax();
    if (state.syntaxErrors.length > 0) {
        if (!confirm('Hay errores de sintaxis. ¿Guardar de todos modos?')) {
            return;
        }
    }
    
    showLoading('Guardando plantilla...');
    
    try {
        await new Promise(resolve => setTimeout(resolve, 1000));
        
        // Update template
        state.selectedTemplate.code = state.currentCode;
        state.selectedTemplate.updated_at = new Date().toISOString();
        state.selectedTemplate.version = incrementVersion(state.selectedTemplate.version);
        
        state.isDirty = false;
        
        // Update UI
        document.getElementById('templateVersion').textContent = `v${state.selectedTemplate.version}`;
        
        showSuccess('Plantilla guardada correctamente');
        
    } catch (error) {
        showError('Error al guardar la plantilla');
    } finally {
        hideLoading();
    }
}

/**
 * Test template
 */
async function testTemplate() {
    const testDataSelect = document.getElementById('testVoucher');
    const testDataJson = document.getElementById('testDataJson');
    
    let testData;
    if (testDataSelect.value) {
        testData = testVouchers[testDataSelect.value];
    } else if (testDataJson.value) {
        try {
            testData = JSON.parse(testDataJson.value);
        } catch (e) {
            showError('JSON de prueba inválido');
            return;
        }
    } else {
        showError('Seleccione datos de prueba');
        return;
    }
    
    showLoading('Ejecutando plantilla...');
    
    try {
        await new Promise(resolve => setTimeout(resolve, 1500));
        
        // Simulate template execution
        const result = {
            success: true,
            entry: {
                id: 'je-test-001',
                lines: [
                    {
                        account_code: '1305.05',
                        account_name: 'Clientes Nacionales',
                        debit: testData.total_amount,
                        credit: 0,
                        description: `Factura ${testData.voucher_number}`
                    },
                    {
                        account_code: '4135.05',
                        account_name: 'Ingresos por Servicios',
                        debit: 0,
                        credit: testData.metadata.subtotal,
                        description: 'Venta de servicios'
                    },
                    {
                        account_code: '2408.01',
                        account_name: 'IVA por Pagar',
                        debit: 0,
                        credit: testData.metadata.taxes?.iva_19 || 0,
                        description: 'IVA 19% por pagar'
                    }
                ]
            },
            execution_time: 23,
            memory_used: '1.2 MB'
        };
        
        // Show result
        showTestResult(result);
        
    } catch (error) {
        showError('Error al ejecutar la plantilla');
    } finally {
        hideLoading();
    }
}

/**
 * Show test result
 */
function showTestResult(result) {
    const previewContent = document.getElementById('previewContent');
    const previewMode = document.getElementById('previewMode').value;
    
    if (previewMode === 'result') {
        if (result.success) {
            let html = '<h4>✅ Ejecución Exitosa</h4>';
            html += `<p>Tiempo: ${result.execution_time}ms | Memoria: ${result.memory_used}</p>`;
            html += '<h5>Asiento Generado:</h5>';
            html += '<table class="table table-sm">';
            html += '<thead><tr><th>Cuenta</th><th>Nombre</th><th>Débito</th><th>Crédito</th></tr></thead>';
            html += '<tbody>';
            
            result.entry.lines.forEach(line => {
                html += `<tr>
                    <td>${line.account_code}</td>
                    <td>${line.account_name}</td>
                    <td class="text-right">${line.debit ? '$' + formatCurrency(line.debit) : ''}</td>
                    <td class="text-right">${line.credit ? '$' + formatCurrency(line.credit) : ''}</td>
                </tr>`;
            });
            
            const totalDebit = result.entry.lines.reduce((sum, l) => sum + (l.debit || 0), 0);
            const totalCredit = result.entry.lines.reduce((sum, l) => sum + (l.credit || 0), 0);
            
            html += `<tr class="font-weight-bold">
                <td colspan="2">Totales</td>
                <td class="text-right">$${formatCurrency(totalDebit)}</td>
                <td class="text-right">$${formatCurrency(totalCredit)}</td>
            </tr>`;
            
            html += '</tbody></table>';
            
            if (Math.abs(totalDebit - totalCredit) < 0.01) {
                html += '<p class="text-success">✅ Asiento balanceado</p>';
            } else {
                html += '<p class="text-danger">❌ Asiento desbalanceado</p>';
            }
            
            previewContent.innerHTML = html;
        } else {
            previewContent.innerHTML = '<div class="syntax-error">❌ Error en la ejecución</div>';
        }
    } else if (previewMode === 'ast') {
        // Show AST
        previewContent.innerHTML = `<div class="ast-tree">
{
  "type": "template",
  "name": "${state.selectedTemplate?.name}",
  "body": {
    "variables": [
      { "type": "let", "name": "subtotal", "value": "voucher.metadata.subtotal" },
      { "type": "let", "name": "tax_rate", "value": 0.19 },
      { "type": "let", "name": "tax_amount", "value": "voucher.metadata.taxes.iva_19" }
    ],
    "validations": [
      { "type": "require", "condition": "subtotal > 0", "message": "Subtotal debe ser positivo" }
    ],
    "entry": {
      "type": "entry",
      "lines": [
        {
          "type": "debit",
          "account": "1305.05",
          "amount": "total",
          "metadata": {...}
        },
        {
          "type": "credit",
          "account": "4135.05",
          "amount": "subtotal",
          "metadata": {...}
        }
      ]
    }
  }
}</div>`;
    } else if (previewMode === 'debug') {
        // Show debug info
        previewContent.innerHTML = `<div style="font-family: monospace; font-size: 12px;">
<h5>Debug Information</h5>
<p><strong>Template:</strong> ${state.selectedTemplate?.name}</p>
<p><strong>Version:</strong> ${state.selectedTemplate?.version}</p>
<p><strong>Code Length:</strong> ${state.currentCode.length} characters</p>
<p><strong>Lines:</strong> ${state.currentCode.split('\n').length}</p>
<p><strong>Variables Found:</strong> subtotal, tax_rate, tax_amount, total</p>
<p><strong>Accounts Used:</strong> 1305.05, 4135.05, 2408.01</p>
<p><strong>Test Data:</strong></p>
<pre>${JSON.stringify(testVouchers[document.getElementById('testVoucher').value] || {}, null, 2)}</pre>
</div>`;
    }
}

/**
 * Load test data
 */
function loadTestData() {
    const testVoucher = document.getElementById('testVoucher').value;
    const testDataJson = document.getElementById('testDataJson');
    
    if (testVoucher && testVouchers[testVoucher]) {
        testDataJson.value = JSON.stringify(testVouchers[testVoucher], null, 2);
    }
}

/**
 * Update preview
 */
function updatePreview() {
    // Re-show last result if any
    const previewContent = document.getElementById('previewContent');
    if (!previewContent.innerHTML.trim()) {
        previewContent.innerHTML = '<p class="text-muted text-center" style="margin-top: 50px;">Ejecute una prueba para ver resultados</p>';
    }
}

/**
 * Create new template
 */
function createNewTemplate() {
    document.getElementById('templateModal').style.display = 'block';
}

/**
 * Close template modal
 */
function closeTemplateModal() {
    document.getElementById('templateModal').style.display = 'none';
    document.getElementById('templateForm').reset();
}

/**
 * Create template
 */
async function createTemplate() {
    const form = document.getElementById('templateForm');
    if (!form.checkValidity()) {
        form.reportValidity();
        return;
    }
    
    const newTemplate = {
        id: 'tpl-' + Date.now(),
        name: document.getElementById('templateName').value,
        description: document.getElementById('templateDescription').value,
        type: document.getElementById('templateVoucherType').value,
        country: document.getElementById('templateCountry').value,
        version: '1.0',
        is_active: true,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        code: getTemplateBase(document.getElementById('templateBase').value)
    };
    
    state.templates.unshift(newTemplate);
    renderTemplateList();
    
    closeTemplateModal();
    selectTemplate(newTemplate);
    
    showSuccess('Plantilla creada correctamente');
}

/**
 * Get template base code
 */
function getTemplateBase(baseType) {
    const bases = {
        basic_sale: `template ${document.getElementById('templateName').value} {
  // Variables
  let subtotal = voucher.metadata.subtotal
  let tax = voucher.metadata.taxes.iva_19
  let total = voucher.total_amount
  
  // Validaciones
  require total > 0 : "Total debe ser positivo"
  
  // Asiento contable
  entry {
    debit account("1305.05") amount(total) {
      description = "Venta " + voucher.voucher_number
    }
    
    credit account("4135.05") amount(subtotal) {
      description = "Ingreso por ventas"
    }
    
    credit account("2408.01") amount(tax) {
      description = "IVA por pagar"
    }
  }
}`,
        basic_purchase: `template ${document.getElementById('templateName').value} {
  // Variables
  let amount = voucher.total_amount
  
  // Asiento contable
  entry {
    debit account("5135.30") amount(amount) {
      description = voucher.description
    }
    
    credit account("2205.01") amount(amount) {
      description = "Por pagar"
    }
  }
}`,
        basic_payment: `template ${document.getElementById('templateName').value} {
  let amount = voucher.total_amount
  
  entry {
    debit account("2205.01") amount(amount) {
      description = "Pago " + voucher.voucher_number
    }
    
    credit account("1110.05") amount(amount) {
      description = "Salida de banco"
    }
  }
}`
    };
    
    return bases[baseType] || `template ${document.getElementById('templateName').value} {\n  // Código de la plantilla\n  \n}`;
}

/**
 * Show version history
 */
function showVersionHistory() {
    alert('Historial de versiones próximamente');
}

/**
 * Show DSL docs
 */
function showDSLDocs() {
    window.open('https://github.com/arturoeanton/go-dsl/blob/main/README.md', '_blank');
}

// Helper functions
function getCurrentLine(range) {
    const text = document.getElementById('codeEditor').textContent;
    const beforeCursor = text.substring(0, range.startOffset);
    const lines = beforeCursor.split('\n');
    return lines[lines.length - 1];
}

function getLineIndent(line) {
    const match = line.match(/^(\s*)/);
    return match ? match[1] : '';
}

function incrementVersion(version) {
    const parts = version.split('.');
    const minor = parseInt(parts[1]) + 1;
    return `${parts[0]}.${minor}`;
}

function getVoucherTypeLabel(type) {
    const labels = {
        'invoice_sale': 'Factura Venta',
        'invoice_purchase': 'Factura Compra',
        'payment': 'Pago',
        'receipt': 'Recibo',
        'credit_note': 'Nota Crédito',
        'debit_note': 'Nota Débito',
        'custom': 'Personalizado'
    };
    return labels[type] || type;
}

function formatDate(dateStr) {
    const date = new Date(dateStr);
    return date.toLocaleDateString('es-CO');
}

function formatCurrency(amount) {
    return new Intl.NumberFormat('es-CO', {
        minimumFractionDigits: 0,
        maximumFractionDigits: 2
    }).format(amount);
}

function showLoading(message) {
    const overlay = document.getElementById('loadingOverlay');
    overlay.querySelector('p').textContent = message || 'Cargando...';
    overlay.style.display = 'flex';
}

function hideLoading() {
    document.getElementById('loadingOverlay').style.display = 'none';
}

function showSuccess(message) {
    console.log('Success:', message);
}

function showError(message) {
    console.error('Error:', message);
    alert(message);
}

function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}