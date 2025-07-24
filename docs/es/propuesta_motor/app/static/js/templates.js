// Templates management
let templates = [];
let currentTemplate = null;
let dslEditor = null;

// Initialize page
document.addEventListener('DOMContentLoaded', function() {
    initializeDSLEditor();
    loadTemplates();
    setupEventListeners();
});

// Initialize CodeMirror editor
function initializeDSLEditor() {
    const editorElement = document.getElementById('dslEditor');
    if (editorElement && typeof CodeMirror !== 'undefined') {
        // Clear any existing content
        editorElement.innerHTML = '';
        
        dslEditor = CodeMirror(editorElement, {
            mode: 'text/plain',
            theme: 'monokai',
            lineNumbers: true,
            autoCloseBrackets: true,
            matchBrackets: true,
            indentUnit: 2,
            value: '# Template de asiento\ntemplate example\n  params ($param1, $param2)\n  \n  entry\n    description: "Descripción del asiento"\n    date: last_day($period)\n    \n    line debit account("1105") amount($param1)\n         description("Débito ejemplo")\n    \n    line credit account("4105") amount($param1)\n         description("Crédito ejemplo")'
        });
        
        console.log('CodeMirror initialized successfully');
    } else {
        console.log('CodeMirror not available or element not found', {
            editorElement: !!editorElement,
            CodeMirror: typeof CodeMirror
        });
        
        // Fallback to textarea if CodeMirror is not loaded
        const textarea = document.createElement('textarea');
        textarea.id = 'dslTextarea';
        textarea.style.width = '100%';
        textarea.style.height = '100%';
        textarea.style.fontFamily = 'monospace';
        textarea.style.backgroundColor = '#272822';
        textarea.style.color = '#f8f8f2';
        textarea.style.padding = '10px';
        textarea.style.border = 'none';
        textarea.style.outline = 'none';
        textarea.value = '# Template de asiento\ntemplate example\n  params ($param1, $param2)';
        if (editorElement) {
            editorElement.innerHTML = '';
            editorElement.appendChild(textarea);
        }
    }
}

// Setup event listeners
function setupEventListeners() {
    // Search functionality
    const searchInput = document.getElementById('templateSearch');
    if (searchInput) {
        searchInput.addEventListener('input', utils.debounce(filterTemplates, 300));
    }

    // Filter buttons
    document.querySelectorAll('.filter-btn').forEach(btn => {
        btn.addEventListener('click', function() {
            document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
            this.classList.add('active');
            filterTemplates();
        });
    });
}

// Filter templates
function filterTemplates() {
    const searchTerm = document.getElementById('templateSearch')?.value.toLowerCase() || '';
    const activeFilter = document.querySelector('.filter-btn.active')?.dataset.filter || 'all';
    
    let filtered = templates;
    
    // Apply search filter
    if (searchTerm) {
        filtered = filtered.filter(t => 
            t.name.toLowerCase().includes(searchTerm) ||
            (t.description && t.description.toLowerCase().includes(searchTerm))
        );
    }
    
    // Apply status filter
    if (activeFilter !== 'all') {
        filtered = filtered.filter(t => {
            // Handle both is_active and status fields
            const isActive = t.is_active !== undefined ? t.is_active : (t.status === 'active');
            return (activeFilter === 'active' && isActive) ||
                   (activeFilter === 'inactive' && !isActive);
        });
    }
    
    renderFilteredTemplates(filtered);
}

// Render filtered templates
function renderFilteredTemplates(filteredTemplates) {
    const grid = document.getElementById('templatesContainer');
    
    if (filteredTemplates.length === 0) {
        grid.innerHTML = `
            <div class="empty-state">
                <i class="fas fa-search fa-3x"></i>
                <p>No se encontraron templates</p>
            </div>
        `;
        return;
    }
    
    grid.innerHTML = filteredTemplates.map(template => `
        <div class="template-card">
            <div class="template-header">
                <h4>${template.name}</h4>
                <div class="template-actions">
                    <button class="btn btn-sm btn-info" onclick="showExecuteModal('${template.id}')" title="Ejecutar">
                        ▶️
                    </button>
                    <button class="btn btn-sm btn-secondary" onclick="editTemplate('${template.id}')" title="Editar">
                        ✏️
                    </button>
                    <button class="btn btn-sm btn-danger" onclick="deleteTemplate('${template.id}')" title="Eliminar">
                        🗑️
                    </button>
                </div>
            </div>
            <p class="template-description">${template.description || 'Sin descripción'}</p>
            <div class="template-meta">
                <span>📋 ${getTemplateTypeLabel(template.type)}</span>
                <span class="${template.status === 'active' ? 'status-active' : 'status-inactive'}">
                    ${template.status === 'active' ? 'Activo' : 'Inactivo'}
                </span>
            </div>
        </div>
    `).join('');
}

// Load templates
async function loadTemplates() {
    try {
        console.log('Loading templates...');
        
        // Try regular endpoint first
        console.log('Trying regular templates endpoint...');
        const response = await fetch('/api/v1/templates');
        console.log('Response status:', response.status);
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        const data = await response.json();
        console.log('API response data:', data);
        
        // Handle different response formats
        if (Array.isArray(data)) {
            templates = data;
        } else if (data.templates && Array.isArray(data.templates)) {
            templates = data.templates;
        } else if (data.data && Array.isArray(data.data)) {
            templates = data.data;
        } else {
            console.error('Unexpected response format:', data);
            throw new Error('Invalid response format');
        }
        
        console.log('Templates loaded:', templates.length, 'templates');
        renderTemplates();
    } catch (error) {
        console.error('Error loading templates:', error);
        
        // Try apiService as fallback
        try {
            if (window.apiService && window.apiService.getTemplates) {
                console.log('Trying apiService fallback...');
                const response = await window.apiService.getTemplates();
                
                if (response.success && response.data) {
                    templates = Array.isArray(response.data) ? response.data : [];
                    console.log('Loaded via apiService:', templates.length, 'templates');
                    renderTemplates();
                    return;
                }
            }
        } catch (apiError) {
            console.error('ApiService fallback also failed:', apiError);
        }
        
        // Last resort: mock data
        console.log('All API calls failed. Using mock data...');
        templates = getMockTemplates();
        renderTemplates();
    }
}

// Get mock templates for development
function getMockTemplates() {
    return [
        {
            id: 'tpl_001',
            name: 'Factura de Venta Estándar',
            description: 'Genera asiento contable para facturas de venta con IVA 19%',
            type: 'invoice_sale',
            status: 'active'
        },
        {
            id: 'tpl_002',
            name: 'Factura de Venta con Retención',
            description: 'Genera asiento para facturas con retención en la fuente',
            type: 'invoice_sale',
            status: 'active'
        },
        {
            id: 'tpl_003',
            name: 'Nómina Mensual Básica',
            description: 'Genera asiento contable para pago de nómina mensual',
            type: 'payroll',
            status: 'active'
        },
        {
            id: 'tpl_004',
            name: 'Compra con IVA',
            description: 'Registra compras de inventario o gastos con IVA',
            type: 'invoice_purchase',
            status: 'active'
        },
        {
            id: 'tpl_005',
            name: 'Pago a Proveedor',
            description: 'Registra el pago a proveedores desde banco',
            type: 'payment',
            status: 'active'
        }
    ];
}

// Render templates grid
function renderTemplates() {
    const grid = document.getElementById('templatesContainer');
    
    if (!grid) {
        console.error('Templates container not found');
        return;
    }
    
    // Ensure templates is an array
    if (!Array.isArray(templates)) {
        console.error('Templates is not an array:', templates);
        templates = [];
    }
    
    if (templates.length === 0) {
        grid.innerHTML = `
            <div class="empty-state">
                <i style="font-size: 3rem; opacity: 0.5;">📋</i>
                <p>No hay templates creados</p>
                <button class="btn btn-primary" onclick="showCreateModal()">
                    ➕ Crear primer template
                </button>
            </div>
        `;
        updateStats();
        return;
    }
    
    grid.innerHTML = templates.map(template => `
        <div class="template-card">
            <div class="template-header">
                <h4>${template.name}</h4>
                <div class="template-actions">
                    <button class="btn btn-sm btn-info" onclick="showExecuteModal('${template.id}')" title="Ejecutar">
                        ▶️
                    </button>
                    <button class="btn btn-sm btn-secondary" onclick="editTemplate('${template.id}')" title="Editar">
                        ✏️
                    </button>
                    <button class="btn btn-sm btn-danger" onclick="deleteTemplate('${template.id}')" title="Eliminar">
                        🗑️
                    </button>
                </div>
            </div>
            <p class="template-description">${template.description || 'Sin descripción'}</p>
            <div class="template-meta">
                <span>📋 ${getTemplateTypeLabel(template.type)}</span>
                <span class="${template.status === 'active' ? 'status-active' : 'status-inactive'}">
                    ${template.status === 'active' ? 'Activo' : 'Inactivo'}
                </span>
            </div>
        </div>
    `).join('');
    
    // Update stats
    updateStats();
}

// Show create modal
function showCreateModal() {
    currentTemplate = null;
    document.getElementById('modalTitle').textContent = 'Nuevo Template';
    document.getElementById('templateName').value = '';
    document.getElementById('templateDescription').value = '';
    document.getElementById('parametersContainer').innerHTML = `
        <div class="parameter-row">
            <input type="text" placeholder="Nombre" class="param-name">
            <select class="param-type">
                <option value="number">Número</option>
                <option value="string">Texto</option>
                <option value="date">Fecha</option>
            </select>
            <input type="text" placeholder="Descripción" class="param-description">
            <button class="btn btn-sm btn-danger" onclick="removeParameter(this)">
                <i class="fas fa-trash"></i>
            </button>
        </div>
    `;
    
    // Set default DSL content
    const defaultDSL = '# Template de asiento\ntemplate example\n  params ($param1, $param2)\n  \n  entry\n    description: "Descripción del asiento"\n    date: last_day($period)';
    
    if (dslEditor) {
        dslEditor.setValue(defaultDSL);
        setTimeout(() => dslEditor.refresh(), 100);
    } else if (document.getElementById('dslTextarea')) {
        document.getElementById('dslTextarea').value = defaultDSL;
    }
    
    document.getElementById('templateModal').style.display = 'block';
}

// Edit template
async function editTemplate(id) {
    try {
        console.log('Editing template with ID:', id);
        
        // Find template in local array first
        currentTemplate = templates.find(t => t.id === id);
        console.log('Found template in local array:', currentTemplate);
        
        if (!currentTemplate) {
            // Try to fetch from API
            console.log('Template not found locally, fetching from API...');
            if (window.apiService && window.apiService.getTemplate) {
                currentTemplate = await window.apiService.getTemplate(id);
            } else {
                // Direct API call
                const response = await fetch(`/api/v1/templates/${id}`);
                if (response.ok) {
                    currentTemplate = await response.json();
                } else {
                    throw new Error('Template not found');
                }
            }
            console.log('Fetched template from API:', currentTemplate);
        }
        
        // Log the DSL content
        console.log('Template DSL content:', currentTemplate.dsl_content);
        console.log('Full template object:', JSON.stringify(currentTemplate, null, 2));
        
        document.getElementById('modalTitle').textContent = 'Editar Template';
        document.getElementById('templateName').value = currentTemplate.name;
        document.getElementById('templateDescription').value = currentTemplate.description || '';
        
        // Load parameters
        const container = document.getElementById('parametersContainer');
        container.innerHTML = '';
        if (currentTemplate.parameters) {
            // Handle parameters as JSON string or array
            let params = currentTemplate.parameters;
            if (typeof params === 'string') {
                try {
                    params = JSON.parse(params);
                } catch (e) {
                    params = [];
                }
            }
            if (Array.isArray(params) && params.length > 0) {
                params.forEach(param => {
                    addParameterRow(param.name, param.type, param.description);
                });
            } else {
                addParameter();
            }
        } else {
            addParameter();
        }
        
        // Load DSL code
        console.log('dslEditor exists:', !!dslEditor);
        console.log('dslTextarea exists:', !!document.getElementById('dslTextarea'));
        
        if (dslEditor) {
            console.log('Setting DSL content in CodeMirror:', currentTemplate.dsl_content);
            dslEditor.setValue(currentTemplate.dsl_content || '');
            // Force refresh
            setTimeout(() => {
                dslEditor.refresh();
            }, 100);
        } else if (document.getElementById('dslTextarea')) {
            console.log('Setting DSL content in textarea:', currentTemplate.dsl_content);
            document.getElementById('dslTextarea').value = currentTemplate.dsl_content || '';
        }
        
        document.getElementById('templateModal').style.display = 'block';
    } catch (error) {
        console.error('Error loading template:', error);
        showError('Error cargando template');
    }
}

// Add parameter
function addParameter() {
    addParameterRow('', 'number', '');
}

function addParameterRow(name, type, description) {
    const container = document.getElementById('parametersContainer');
    const row = document.createElement('div');
    row.className = 'parameter-row';
    row.innerHTML = `
        <input type="text" placeholder="Nombre" class="param-name" value="${name}">
        <select class="param-type">
            <option value="number" ${type === 'number' ? 'selected' : ''}>Número</option>
            <option value="string" ${type === 'string' ? 'selected' : ''}>Texto</option>
            <option value="date" ${type === 'date' ? 'selected' : ''}>Fecha</option>
        </select>
        <input type="text" placeholder="Descripción" class="param-description" value="${description}">
        <button class="btn btn-sm btn-danger" onclick="removeParameter(this)">
            <i class="fas fa-trash"></i>
        </button>
    `;
    container.appendChild(row);
}

// Remove parameter
function removeParameter(button) {
    const row = button.closest('.parameter-row');
    if (document.querySelectorAll('.parameter-row').length > 1) {
        row.remove();
    }
}

// Delete template
async function deleteTemplate(id) {
    if (!confirm('¿Está seguro de eliminar este template?')) {
        return;
    }
    
    try {
        if (window.apiService && window.apiService.deleteTemplate) {
            const response = await window.apiService.deleteTemplate(id);
            if (response.success) {
                showSuccess('Template eliminado correctamente');
                loadTemplates();
            } else {
                throw new Error('Error al eliminar');
            }
        } else {
            // Direct API call
            const response = await fetch(`/api/v1/templates/${id}`, {
                method: 'DELETE'
            });
            if (response.ok) {
                showSuccess('Template eliminado correctamente');
                loadTemplates();
            } else {
                throw new Error('Error al eliminar');
            }
        }
    } catch (error) {
        console.error('Error deleting template:', error);
        showError('Error eliminando template');
    }
}

// Show execute modal
function showExecuteModal(templateId) {
    const template = templates.find(t => t.id === templateId);
    if (!template) return;
    
    currentTemplate = template;
    alert('Función de ejecución en desarrollo');
    // TODO: Implement execute modal
}

// Save template
async function saveTemplate() {
    const name = document.getElementById('templateName').value.trim();
    const description = document.getElementById('templateDescription').value.trim();
    const dslCode = dslEditor ? dslEditor.getValue() : document.getElementById('dslTextarea')?.value || '';
    
    if (!name) {
        showError('El nombre del template es requerido');
        return;
    }
    
    if (!dslCode) {
        showError('El código DSL es requerido');
        return;
    }
    
    // Collect parameters
    const parameters = [];
    document.querySelectorAll('.parameter-row').forEach(row => {
        const paramName = row.querySelector('.param-name').value.trim();
        const paramType = row.querySelector('.param-type').value;
        const paramDesc = row.querySelector('.param-description').value.trim();
        
        if (paramName) {
            parameters.push({
                name: paramName,
                type: paramType,
                description: paramDesc
            });
        }
    });
    
    const templateData = {
        name: name,
        description: description,
        dsl_content: dslCode,
        parameters: parameters,
        is_active: true
    };
    
    try {
        if (currentTemplate) {
            const response = await window.apiService.updateTemplate(currentTemplate.id, templateData);
            if (response.success) {
                showSuccess('Template actualizado correctamente');
            } else {
                throw new Error(response.error || 'Error updating template');
            }
        } else {
            const response = await window.apiService.createTemplate(templateData);
            if (response.success) {
                showSuccess('Template creado correctamente');
            } else {
                throw new Error(response.error || 'Error creating template');
            }
        }
        closeModal();
        loadTemplates();
    } catch (error) {
        console.error('Error saving template:', error);
        showError('Error guardando template');
    }
}


// Show execute modal
async function showExecuteModal(templateId) {
    const template = templates.find(t => t.id === templateId);
    if (!template) return;
    
    currentTemplate = template;
    document.getElementById('executeTemplateName').textContent = template.name;
    
    // Build parameters form
    const form = document.getElementById('parametersForm');
    form.innerHTML = '';
    
    if (template.parameters && template.parameters.length > 0) {
        template.parameters.forEach(param => {
            const group = document.createElement('div');
            group.className = 'form-group';
            
            let inputHtml = '';
            switch (param.type) {
                case 'number':
                    inputHtml = `<input type="number" id="param_${param.name}" class="form-control" step="0.01">`;
                    break;
                case 'date':
                    inputHtml = `<input type="date" id="param_${param.name}" class="form-control">`;
                    break;
                default:
                    inputHtml = `<input type="text" id="param_${param.name}" class="form-control">`;
            }
            
            group.innerHTML = `
                <label>${param.name} ${param.description ? `<small>(${param.description})</small>` : ''}</label>
                ${inputHtml}
            `;
            form.appendChild(group);
        });
    } else {
        form.innerHTML = '<p>Este template no requiere parámetros</p>';
    }
    
    document.getElementById('previewContainer').style.display = 'none';
    document.getElementById('executeModal').style.display = 'block';
}

// Preview template
async function previewTemplate() {
    const params = collectParameters();
    
    try {
        const result = await apiService.post(`templates/${currentTemplate.id}/preview`, {
            parameters: params
        });
        
        renderPreview(result.preview);
        document.getElementById('previewContainer').style.display = 'block';
    } catch (error) {
        console.error('Error previewing template:', error);
        showError('Error generando vista previa');
    }
}

// Execute template
async function executeTemplate() {
    const params = collectParameters();
    
    if (!confirm('¿Está seguro de ejecutar este template?')) {
        return;
    }
    
    try {
        const result = await apiService.post(`templates/${currentTemplate.id}/execute`, {
            parameters: params,
            dry_run: false
        });
        
        showSuccess(`Asiento ${result.entry_number} creado correctamente`);
        closeExecuteModal();
        
        // Redirect to journal entries
        setTimeout(() => {
            window.location.href = '/journal-entries';
        }, 1500);
    } catch (error) {
        console.error('Error executing template:', error);
        showError('Error ejecutando template');
    }
}

// Collect parameters from form
function collectParameters() {
    const params = {};
    
    if (currentTemplate.parameters) {
        currentTemplate.parameters.forEach(param => {
            const input = document.getElementById(`param_${param.name}`);
            if (input) {
                let value = input.value;
                if (param.type === 'number') {
                    value = parseFloat(value) || 0;
                }
                params[param.name] = value;
            }
        });
    }
    
    return params;
}

// Render preview
function renderPreview(entry) {
    const preview = document.getElementById('entryPreview');
    
    let html = `
        <div class="entry-preview">
            <div class="entry-header">
                <p><strong>Fecha:</strong> ${formatDate(entry.date)}</p>
                <p><strong>Descripción:</strong> ${entry.description}</p>
                ${entry.reference ? `<p><strong>Referencia:</strong> ${entry.reference}</p>` : ''}
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
    
    entry.lines.forEach(line => {
        html += `
            <tr>
                <td>${line.account_code}</td>
                <td>${line.description}</td>
                <td class="amount">${line.debit_amount ? formatCurrency(line.debit_amount) : ''}</td>
                <td class="amount">${line.credit_amount ? formatCurrency(line.credit_amount) : ''}</td>
            </tr>
        `;
    });
    
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
        </div>
    `;
    
    preview.innerHTML = html;
}

// DSL functions
function formatDSL() {
    // Simple formatting - in real implementation would use proper DSL parser
    const code = dslEditor.getValue();
    const formatted = code
        .split('\n')
        .map(line => {
            const trimmed = line.trim();
            if (trimmed.startsWith('line') || trimmed.startsWith('if')) {
                return '    ' + trimmed;
            }
            if (trimmed.startsWith('description') || trimmed.startsWith('amount')) {
                return '         ' + trimmed;
            }
            return trimmed;
        })
        .join('\n');
    dslEditor.setValue(formatted);
}

function validateDSL() {
    showInfo('Validación DSL en desarrollo');
}

function showDSLHelp() {
    alert(`Sintaxis DSL para Templates:

template nombre
  params ($param1, $param2)
  
  entry
    description: "Texto"
    date: función_fecha()
    
    line debit account("codigo") amount(valor)
         description("texto")
    
    line credit account("codigo") amount(valor)
         description("texto")

Funciones disponibles:
- last_day($period): Último día del período
- get_balance($account): Balance de cuenta
- sum(): Suma de valores
- if/then/else: Condicionales
- foreach: Iteración`);
}

// Modal functions
function closeModal() {
    document.getElementById('templateModal').style.display = 'none';
    currentTemplate = null;
}

function closeExecuteModal() {
    document.getElementById('executeModal').style.display = 'none';
    document.getElementById('previewContainer').style.display = 'none';
    currentTemplate = null;
}

// Utility functions
function formatDate(dateStr) {
    return new Date(dateStr).toLocaleDateString('es-CO');
}

function formatCurrency(amount) {
    return new Intl.NumberFormat('es-CO', {
        style: 'currency',
        currency: 'COP'
    }).format(amount);
}

function showSuccess(message) {
    alert(message); // Replace with proper notification
}

function showError(message) {
    alert('Error: ' + message); // Replace with proper notification
}

function showInfo(message) {
    alert(message); // Replace with proper notification
}

// Get template type label
function getTemplateTypeLabel(type) {
    const labels = {
        'invoice_sale': 'Factura Venta',
        'invoice_purchase': 'Factura Compra',
        'payroll': 'Nómina',
        'payment': 'Pago',
        'receipt': 'Recibo',
        'adjustment': 'Ajuste'
    };
    return labels[type] || type;
}

// Update statistics
function updateStats() {
    const totalElement = document.getElementById('totalTemplates');
    const activeElement = document.getElementById('activeTemplates');
    
    if (totalElement) totalElement.textContent = templates.length;
    if (activeElement) {
        const activeCount = templates.filter(t => t.status === 'active').length;
        activeElement.textContent = activeCount;
    }
}

// Add global functions
window.showNewTemplateModal = showCreateModal;
window.closeTemplateModal = closeModal;
window.closeExecuteModal = closeExecuteModal;
window.closeTestModal = () => {
    document.getElementById('testModal').style.display = 'none';
};
window.editTemplate = editTemplate;
window.deleteTemplate = deleteTemplate;
window.showExecuteModal = showExecuteModal;
window.saveTemplate = saveTemplate;
window.addParameter = addParameter;
window.removeParameter = removeParameter;
window.formatDSL = formatDSL;
window.validateDSL = validateDSL;
window.showDSLHelp = showDSLHelp;

// Add styles for templates
const style = document.createElement('style');
style.textContent = `
.templates-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
    gap: 20px;
    margin-top: 20px;
}

.template-card {
    background: var(--surface);
    border-radius: 8px;
    padding: 20px;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.template-header {
    display: flex;
    justify-content: space-between;
    align-items: start;
    margin-bottom: 10px;
}

.template-header h4 {
    margin: 0;
    color: var(--primary);
}

.template-actions {
    display: flex;
    gap: 5px;
}

.template-description {
    color: var(--text-secondary);
    margin: 10px 0;
    font-size: 14px;
}

.template-meta {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: 15px;
    font-size: 12px;
    color: var(--text-secondary);
}

.status-active {
    color: var(--success);
    font-weight: 500;
}

.status-inactive {
    color: var(--danger);
    font-weight: 500;
}

.parameter-row {
    display: grid;
    grid-template-columns: 1fr 120px 2fr auto;
    gap: 10px;
    margin-bottom: 10px;
    align-items: center;
}

.param-name, .param-type, .param-description {
    padding: 8px;
    border: 1px solid var(--border);
    border-radius: 4px;
}

#dslEditor {
    border: 1px solid var(--border);
    border-radius: 4px;
    height: 400px;
}

.editor-toolbar {
    display: flex;
    gap: 10px;
    margin-top: 10px;
}

.entry-preview {
    background: var(--surface);
    padding: 15px;
    border-radius: 8px;
    margin-top: 15px;
}

.entry-header p {
    margin: 5px 0;
}

.entry-lines {
    width: 100%;
    margin-top: 15px;
}

.balance-check {
    text-align: center;
    padding: 10px;
    margin-top: 15px;
    border-radius: 4px;
}

.balance-check.balanced {
    background: var(--success-bg);
    color: var(--success);
}

.balance-check.unbalanced {
    background: var(--danger-bg);
    color: var(--danger);
}

.modal-large {
    max-width: 900px;
    width: 90%;
}
`;
document.head.appendChild(style);