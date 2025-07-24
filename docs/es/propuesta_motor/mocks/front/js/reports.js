// Reports JavaScript - Motor Contable
// Version: 1.0
// Last Updated: 2024-01-15

// Global state
const state = {
    selectedReport: null,
    reportData: null,
    compareData: null
};

// Report configurations
const reportConfigs = {
    balance_general: {
        title: 'Balance General',
        description: 'Estado de Situación Financiera',
        icon: '⚖️',
        additionalOptions: [
            { id: 'showZeroBalances', label: 'Mostrar cuentas con saldo cero', type: 'checkbox' },
            { id: 'consolidateSubsidiaries', label: 'Consolidar filiales', type: 'checkbox' },
            { id: 'detailLevel', label: 'Nivel de detalle', type: 'select', 
              options: [
                { value: '3', label: 'Nivel 3 (Cuentas)' },
                { value: '4', label: 'Nivel 4 (Subcuentas)' },
                { value: 'all', label: 'Todos los niveles' }
              ]
            }
        ]
    },
    estado_resultados: {
        title: 'Estado de Resultados',
        description: 'Estado de Pérdidas y Ganancias',
        icon: '💰',
        additionalOptions: [
            { id: 'showPercentages', label: 'Mostrar porcentajes', type: 'checkbox', checked: true },
            { id: 'compareBudget', label: 'Comparar con presupuesto', type: 'checkbox' },
            { id: 'costCenter', label: 'Centro de costos', type: 'select',
              options: [
                { value: '', label: 'Todos' },
                { value: 'CC-001', label: 'Administración' },
                { value: 'CC-002', label: 'Ventas' },
                { value: 'CC-003', label: 'Producción' }
              ]
            }
        ]
    },
    libro_diario: {
        title: 'Libro Diario',
        description: 'Registro cronológico de asientos',
        icon: '📚',
        additionalOptions: [
            { id: 'includeReversed', label: 'Incluir asientos reversados', type: 'checkbox' },
            { id: 'groupByVoucher', label: 'Agrupar por comprobante', type: 'checkbox' }
        ]
    },
    libro_mayor: {
        title: 'Libro Mayor',
        description: 'Movimientos por cuenta',
        icon: '📖',
        additionalOptions: [
            { id: 'accountFilter', label: 'Filtrar cuentas', type: 'text', placeholder: 'Código o nombre' },
            { id: 'showRunningBalance', label: 'Mostrar saldo acumulado', type: 'checkbox', checked: true }
        ]
    },
    balance_comprobacion: {
        title: 'Balance de Comprobación',
        description: 'Saldos de todas las cuentas',
        icon: '✅',
        additionalOptions: [
            { id: 'showMovements', label: 'Mostrar movimientos del período', type: 'checkbox' },
            { id: 'onlyWithBalance', label: 'Solo cuentas con saldo', type: 'checkbox' }
        ]
    },
    auxiliar_terceros: {
        title: 'Auxiliar de Terceros',
        description: 'Estado de cuenta por tercero',
        icon: '👥',
        additionalOptions: [
            { id: 'thirdPartyType', label: 'Tipo de tercero', type: 'select',
              options: [
                { value: '', label: 'Todos' },
                { value: 'customer', label: 'Clientes' },
                { value: 'supplier', label: 'Proveedores' },
                { value: 'employee', label: 'Empleados' }
              ]
            },
            { id: 'thirdPartyId', label: 'Tercero específico', type: 'text', placeholder: 'NIT o nombre' },
            { id: 'showAging', label: 'Incluir análisis de vencimientos', type: 'checkbox' }
        ]
    },
    flujo_efectivo: {
        title: 'Flujo de Efectivo',
        description: 'Estado de flujos de efectivo',
        icon: '💸',
        additionalOptions: [
            { id: 'method', label: 'Método', type: 'select',
              options: [
                { value: 'direct', label: 'Método directo' },
                { value: 'indirect', label: 'Método indirecto' }
              ]
            }
        ]
    },
    indicadores: {
        title: 'Indicadores Financieros',
        description: 'Ratios y métricas',
        icon: '📊',
        additionalOptions: [
            { id: 'indicators', label: 'Indicadores a incluir', type: 'multiselect',
              options: [
                { value: 'liquidity', label: 'Liquidez' },
                { value: 'solvency', label: 'Solvencia' },
                { value: 'profitability', label: 'Rentabilidad' },
                { value: 'efficiency', label: 'Eficiencia' },
                { value: 'leverage', label: 'Apalancamiento' }
              ]
            }
        ]
    }
};

// Initialize on DOM ready
document.addEventListener('DOMContentLoaded', () => {
    initializeReports();
});

/**
 * Initialize reports
 */
function initializeReports() {
    setupEventListeners();
    setPeriodDefaults();
}

/**
 * Setup event listeners
 */
function setupEventListeners() {
    // Search
    document.getElementById('searchInput').addEventListener('keyup', debounce(filterReports, 300));
    
    // Period changes
    document.getElementById('dateFrom').addEventListener('change', validatePeriod);
    document.getElementById('dateTo').addEventListener('change', validatePeriod);
}

/**
 * Select report type
 */
function selectReport(reportType) {
    state.selectedReport = reportType;
    
    // Update UI
    document.querySelectorAll('.report-card').forEach(card => {
        card.classList.remove('selected');
    });
    document.querySelector(`[data-report="${reportType}"]`).classList.add('selected');
    
    // Show configuration
    showReportConfig(reportType);
}

/**
 * Show report configuration
 */
function showReportConfig(reportType) {
    const config = reportConfigs[reportType];
    if (!config) return;
    
    document.getElementById('reportConfig').style.display = 'block';
    
    // Add additional options
    const additionalOptions = document.getElementById('additionalOptions');
    additionalOptions.innerHTML = '';
    
    if (config.additionalOptions) {
        config.additionalOptions.forEach(option => {
            const formGroup = createFormGroup(option);
            additionalOptions.appendChild(formGroup);
        });
    }
    
    // Scroll to config
    document.getElementById('reportConfig').scrollIntoView({ behavior: 'smooth' });
}

/**
 * Create form group for additional option
 */
function createFormGroup(option) {
    const div = document.createElement('div');
    div.className = 'form-group';
    
    const label = document.createElement('label');
    label.textContent = option.label;
    div.appendChild(label);
    
    switch (option.type) {
        case 'checkbox':
            const checkbox = document.createElement('input');
            checkbox.type = 'checkbox';
            checkbox.id = option.id;
            checkbox.checked = option.checked || false;
            
            const checkboxLabel = document.createElement('label');
            checkboxLabel.className = 'checkbox-label';
            checkboxLabel.style.marginLeft = '10px';
            checkboxLabel.appendChild(checkbox);
            checkboxLabel.appendChild(document.createTextNode(' ' + option.label));
            
            div.innerHTML = '';
            div.appendChild(checkboxLabel);
            break;
            
        case 'select':
            const select = document.createElement('select');
            select.className = 'form-control';
            select.id = option.id;
            
            option.options.forEach(opt => {
                const optionEl = document.createElement('option');
                optionEl.value = opt.value;
                optionEl.textContent = opt.label;
                select.appendChild(optionEl);
            });
            
            div.appendChild(select);
            break;
            
        case 'text':
            const input = document.createElement('input');
            input.type = 'text';
            input.className = 'form-control';
            input.id = option.id;
            input.placeholder = option.placeholder || '';
            div.appendChild(input);
            break;
            
        case 'multiselect':
            const checkboxGroup = document.createElement('div');
            checkboxGroup.className = 'checkbox-group';
            
            option.options.forEach(opt => {
                const cbDiv = document.createElement('div');
                cbDiv.style.marginBottom = '5px';
                
                const cb = document.createElement('input');
                cb.type = 'checkbox';
                cb.id = `${option.id}_${opt.value}`;
                cb.value = opt.value;
                
                const cbLabel = document.createElement('label');
                cbLabel.style.marginLeft = '5px';
                cbLabel.setAttribute('for', cb.id);
                cbLabel.textContent = opt.label;
                
                cbDiv.appendChild(cb);
                cbDiv.appendChild(cbLabel);
                checkboxGroup.appendChild(cbDiv);
            });
            
            div.appendChild(checkboxGroup);
            break;
    }
    
    return div;
}

/**
 * Set period based on preset
 */
function setPeriod(preset) {
    const today = new Date();
    let dateFrom, dateTo;
    
    switch (preset) {
        case 'month':
            dateFrom = new Date(today.getFullYear(), today.getMonth(), 1);
            dateTo = new Date(today.getFullYear(), today.getMonth() + 1, 0);
            break;
            
        case 'quarter':
            const quarter = Math.floor(today.getMonth() / 3);
            dateFrom = new Date(today.getFullYear(), quarter * 3, 1);
            dateTo = new Date(today.getFullYear(), quarter * 3 + 3, 0);
            break;
            
        case 'year':
            dateFrom = new Date(today.getFullYear(), 0, 1);
            dateTo = new Date(today.getFullYear(), 11, 31);
            break;
            
        case 'custom':
            return; // Don't change dates for custom
    }
    
    document.getElementById('dateFrom').value = dateFrom.toISOString().split('T')[0];
    document.getElementById('dateTo').value = dateTo.toISOString().split('T')[0];
}

/**
 * Generate report
 */
async function generateReport() {
    if (!state.selectedReport) {
        showError('Por favor seleccione un tipo de reporte');
        return;
    }
    
    const format = document.querySelector('input[name="format"]:checked').value;
    
    if (format === 'preview') {
        showLoading('Generando vista previa del reporte...');
        
        try {
            // Simulate report generation
            await new Promise(resolve => setTimeout(resolve, 2000));
            
            // Load mock report data
            const reportData = await generateMockReport(state.selectedReport);
            state.reportData = reportData;
            
            // Show preview
            showReportPreview(reportData);
            
        } catch (error) {
            showError('Error al generar el reporte');
        } finally {
            hideLoading();
        }
    } else {
        // Export directly
        exportReport(format);
    }
}

/**
 * Generate mock report data
 */
async function generateMockReport(reportType) {
    const config = reportConfigs[reportType];
    const dateFrom = document.getElementById('dateFrom').value;
    const dateTo = document.getElementById('dateTo').value;
    const organization = document.getElementById('organization').options[
        document.getElementById('organization').selectedIndex
    ].text;
    
    return {
        type: reportType,
        title: config.title,
        organization,
        period: `${formatDate(dateFrom)} al ${formatDate(dateTo)}`,
        generatedAt: new Date().toISOString(),
        content: getMockReportContent(reportType)
    };
}

/**
 * Get mock report content based on type
 */
function getMockReportContent(reportType) {
    switch (reportType) {
        case 'balance_general':
            return generateBalanceGeneral();
        case 'estado_resultados':
            return generateEstadoResultados();
        case 'libro_diario':
            return generateLibroDiario();
        case 'libro_mayor':
            return generateLibroMayor();
        case 'balance_comprobacion':
            return generateBalanceComprobacion();
        default:
            return '<p>Contenido del reporte en construcción...</p>';
    }
}

/**
 * Generate Balance General content
 */
function generateBalanceGeneral() {
    return `
        <div class="balance-general">
            <table class="table">
                <thead>
                    <tr>
                        <th colspan="3" class="text-center">ACTIVOS</th>
                    </tr>
                </thead>
                <tbody>
                    <tr class="section-header">
                        <td colspan="2"><strong>ACTIVO CORRIENTE</strong></td>
                        <td class="text-right"><strong>85,450,000</strong></td>
                    </tr>
                    <tr>
                        <td width="50"></td>
                        <td>Efectivo y Equivalentes</td>
                        <td class="text-right">48,450,000</td>
                    </tr>
                    <tr>
                        <td></td>
                        <td>Cuentas por Cobrar</td>
                        <td class="text-right">35,000,000</td>
                    </tr>
                    <tr>
                        <td></td>
                        <td>Inventarios</td>
                        <td class="text-right">2,000,000</td>
                    </tr>
                    <tr class="section-header">
                        <td colspan="2"><strong>ACTIVO NO CORRIENTE</strong></td>
                        <td class="text-right"><strong>40,000,000</strong></td>
                    </tr>
                    <tr>
                        <td></td>
                        <td>Propiedad, Planta y Equipo</td>
                        <td class="text-right">40,000,000</td>
                    </tr>
                    <tr class="total-row">
                        <td colspan="2"><strong>TOTAL ACTIVOS</strong></td>
                        <td class="text-right"><strong>125,450,000</strong></td>
                    </tr>
                </tbody>
            </table>
            
            <table class="table mt-4">
                <thead>
                    <tr>
                        <th colspan="3" class="text-center">PASIVOS Y PATRIMONIO</th>
                    </tr>
                </thead>
                <tbody>
                    <tr class="section-header">
                        <td colspan="2"><strong>PASIVO CORRIENTE</strong></td>
                        <td class="text-right"><strong>25,000,000</strong></td>
                    </tr>
                    <tr>
                        <td width="50"></td>
                        <td>Proveedores</td>
                        <td class="text-right">12,000,000</td>
                    </tr>
                    <tr>
                        <td></td>
                        <td>Obligaciones Laborales</td>
                        <td class="text-right">8,000,000</td>
                    </tr>
                    <tr>
                        <td></td>
                        <td>Impuestos por Pagar</td>
                        <td class="text-right">5,000,000</td>
                    </tr>
                    <tr class="section-header">
                        <td colspan="2"><strong>PASIVO NO CORRIENTE</strong></td>
                        <td class="text-right"><strong>20,000,000</strong></td>
                    </tr>
                    <tr>
                        <td></td>
                        <td>Obligaciones Financieras</td>
                        <td class="text-right">20,000,000</td>
                    </tr>
                    <tr class="total-row">
                        <td colspan="2"><strong>TOTAL PASIVOS</strong></td>
                        <td class="text-right"><strong>45,000,000</strong></td>
                    </tr>
                    <tr class="section-header">
                        <td colspan="2"><strong>PATRIMONIO</strong></td>
                        <td class="text-right"><strong>80,450,000</strong></td>
                    </tr>
                    <tr>
                        <td></td>
                        <td>Capital Social</td>
                        <td class="text-right">30,000,000</td>
                    </tr>
                    <tr>
                        <td></td>
                        <td>Reservas</td>
                        <td class="text-right">10,000,000</td>
                    </tr>
                    <tr>
                        <td></td>
                        <td>Utilidades Retenidas</td>
                        <td class="text-right">20,450,000</td>
                    </tr>
                    <tr>
                        <td></td>
                        <td>Utilidad del Ejercicio</td>
                        <td class="text-right">20,000,000</td>
                    </tr>
                    <tr class="total-row">
                        <td colspan="2"><strong>TOTAL PASIVOS Y PATRIMONIO</strong></td>
                        <td class="text-right"><strong>125,450,000</strong></td>
                    </tr>
                </tbody>
            </table>
        </div>
    `;
}

/**
 * Generate Estado de Resultados content
 */
function generateEstadoResultados() {
    const showPercentages = document.getElementById('showPercentages')?.checked;
    
    return `
        <div class="estado-resultados">
            <table class="table">
                <thead>
                    <tr>
                        <th>Concepto</th>
                        <th class="text-right">Valor</th>
                        ${showPercentages ? '<th class="text-right">%</th>' : ''}
                    </tr>
                </thead>
                <tbody>
                    <tr class="section-header">
                        <td><strong>INGRESOS OPERACIONALES</strong></td>
                        <td class="text-right"><strong>85,000,000</strong></td>
                        ${showPercentages ? '<td class="text-right"><strong>100.0%</strong></td>' : ''}
                    </tr>
                    <tr>
                        <td style="padding-left: 20px;">Ventas de Servicios</td>
                        <td class="text-right">50,000,000</td>
                        ${showPercentages ? '<td class="text-right">58.8%</td>' : ''}
                    </tr>
                    <tr>
                        <td style="padding-left: 20px;">Ventas de Productos</td>
                        <td class="text-right">35,000,000</td>
                        ${showPercentages ? '<td class="text-right">41.2%</td>' : ''}
                    </tr>
                    <tr class="section-header">
                        <td><strong>(-) COSTO DE VENTAS</strong></td>
                        <td class="text-right"><strong>(10,000,000)</strong></td>
                        ${showPercentages ? '<td class="text-right"><strong>11.8%</strong></td>' : ''}
                    </tr>
                    <tr class="total-row">
                        <td><strong>UTILIDAD BRUTA</strong></td>
                        <td class="text-right"><strong>75,000,000</strong></td>
                        ${showPercentages ? '<td class="text-right"><strong>88.2%</strong></td>' : ''}
                    </tr>
                    <tr class="section-header">
                        <td><strong>(-) GASTOS OPERACIONALES</strong></td>
                        <td class="text-right"><strong>(55,000,000)</strong></td>
                        ${showPercentages ? '<td class="text-right"><strong>64.7%</strong></td>' : ''}
                    </tr>
                    <tr>
                        <td style="padding-left: 20px;">Gastos de Administración</td>
                        <td class="text-right">(45,000,000)</td>
                        ${showPercentages ? '<td class="text-right">52.9%</td>' : ''}
                    </tr>
                    <tr>
                        <td style="padding-left: 20px;">Gastos de Ventas</td>
                        <td class="text-right">(10,000,000)</td>
                        ${showPercentages ? '<td class="text-right">11.8%</td>' : ''}
                    </tr>
                    <tr class="total-row">
                        <td><strong>UTILIDAD OPERACIONAL</strong></td>
                        <td class="text-right"><strong>20,000,000</strong></td>
                        ${showPercentages ? '<td class="text-right"><strong>23.5%</strong></td>' : ''}
                    </tr>
                    <tr>
                        <td>(+) Otros Ingresos</td>
                        <td class="text-right">0</td>
                        ${showPercentages ? '<td class="text-right">0.0%</td>' : ''}
                    </tr>
                    <tr>
                        <td>(-) Otros Gastos</td>
                        <td class="text-right">0</td>
                        ${showPercentages ? '<td class="text-right">0.0%</td>' : ''}
                    </tr>
                    <tr class="total-row">
                        <td><strong>UTILIDAD ANTES DE IMPUESTOS</strong></td>
                        <td class="text-right"><strong>20,000,000</strong></td>
                        ${showPercentages ? '<td class="text-right"><strong>23.5%</strong></td>' : ''}
                    </tr>
                    <tr>
                        <td>(-) Impuesto de Renta</td>
                        <td class="text-right">0</td>
                        ${showPercentages ? '<td class="text-right">0.0%</td>' : ''}
                    </tr>
                    <tr class="total-row" style="background-color: #e8f5e9;">
                        <td><strong>UTILIDAD NETA</strong></td>
                        <td class="text-right"><strong>20,000,000</strong></td>
                        ${showPercentages ? '<td class="text-right"><strong>23.5%</strong></td>' : ''}
                    </tr>
                </tbody>
            </table>
        </div>
    `;
}

/**
 * Generate Libro Diario content
 */
function generateLibroDiario() {
    return `
        <div class="libro-diario">
            <table class="table table-sm">
                <thead>
                    <tr>
                        <th>Fecha</th>
                        <th>Asiento</th>
                        <th>Cuenta</th>
                        <th>Descripción</th>
                        <th class="text-right">Débito</th>
                        <th class="text-right">Crédito</th>
                    </tr>
                </thead>
                <tbody>
                    <tr>
                        <td rowspan="3">2024-01-21</td>
                        <td rowspan="3">AS-2024-0127</td>
                        <td>1305.05</td>
                        <td>Clientes Nacionales - FV-2024-0127</td>
                        <td class="text-right">2,450,000</td>
                        <td class="text-right"></td>
                    </tr>
                    <tr>
                        <td>4135.05</td>
                        <td>Ingresos por Servicios de Consultoría</td>
                        <td class="text-right"></td>
                        <td class="text-right">2,058,824</td>
                    </tr>
                    <tr>
                        <td>2408.01</td>
                        <td>IVA por Pagar 19%</td>
                        <td class="text-right"></td>
                        <td class="text-right">391,176</td>
                    </tr>
                    <tr class="total-row">
                        <td colspan="4" class="text-right">Totales del día</td>
                        <td class="text-right"><strong>2,450,000</strong></td>
                        <td class="text-right"><strong>2,450,000</strong></td>
                    </tr>
                </tbody>
            </table>
        </div>
    `;
}

/**
 * Generate Libro Mayor content
 */
function generateLibroMayor() {
    return `
        <div class="libro-mayor">
            <h4>Cuenta: 1305.05 - Clientes Nacionales</h4>
            <table class="table table-sm">
                <thead>
                    <tr>
                        <th>Fecha</th>
                        <th>Asiento</th>
                        <th>Descripción</th>
                        <th class="text-right">Débito</th>
                        <th class="text-right">Crédito</th>
                        <th class="text-right">Saldo</th>
                    </tr>
                </thead>
                <tbody>
                    <tr>
                        <td>2024-01-01</td>
                        <td>SALDO</td>
                        <td>Saldo inicial</td>
                        <td class="text-right"></td>
                        <td class="text-right"></td>
                        <td class="text-right">30,000,000</td>
                    </tr>
                    <tr>
                        <td>2024-01-21</td>
                        <td>AS-2024-0127</td>
                        <td>FV-2024-0127 ABC Tecnología</td>
                        <td class="text-right">2,450,000</td>
                        <td class="text-right"></td>
                        <td class="text-right">32,450,000</td>
                    </tr>
                    <tr>
                        <td>2024-01-21</td>
                        <td>AS-2024-0123</td>
                        <td>RC-2024-0123 Abono Juan Pérez</td>
                        <td class="text-right"></td>
                        <td class="text-right">560,000</td>
                        <td class="text-right">31,890,000</td>
                    </tr>
                    <tr class="total-row">
                        <td colspan="3">Movimientos del período</td>
                        <td class="text-right"><strong>2,450,000</strong></td>
                        <td class="text-right"><strong>560,000</strong></td>
                        <td class="text-right"><strong>31,890,000</strong></td>
                    </tr>
                </tbody>
            </table>
        </div>
    `;
}

/**
 * Generate Balance de Comprobación content
 */
function generateBalanceComprobacion() {
    return `
        <div class="balance-comprobacion">
            <table class="table table-sm">
                <thead>
                    <tr>
                        <th rowspan="2">Código</th>
                        <th rowspan="2">Cuenta</th>
                        <th colspan="2" class="text-center">Saldos Anteriores</th>
                        <th colspan="2" class="text-center">Movimientos</th>
                        <th colspan="2" class="text-center">Saldos Actuales</th>
                    </tr>
                    <tr>
                        <th class="text-right">Débito</th>
                        <th class="text-right">Crédito</th>
                        <th class="text-right">Débito</th>
                        <th class="text-right">Crédito</th>
                        <th class="text-right">Débito</th>
                        <th class="text-right">Crédito</th>
                    </tr>
                </thead>
                <tbody>
                    <tr>
                        <td>1105</td>
                        <td>CAJA</td>
                        <td class="text-right">2,000,000</td>
                        <td class="text-right"></td>
                        <td class="text-right">560,000</td>
                        <td class="text-right"></td>
                        <td class="text-right">2,560,000</td>
                        <td class="text-right"></td>
                    </tr>
                    <tr>
                        <td>1110</td>
                        <td>BANCOS</td>
                        <td class="text-right">45,890,000</td>
                        <td class="text-right"></td>
                        <td class="text-right">1,000,000</td>
                        <td class="text-right">10,200,000</td>
                        <td class="text-right">36,690,000</td>
                        <td class="text-right"></td>
                    </tr>
                    <tr>
                        <td>1305</td>
                        <td>CLIENTES</td>
                        <td class="text-right">30,000,000</td>
                        <td class="text-right"></td>
                        <td class="text-right">5,450,000</td>
                        <td class="text-right">1,560,000</td>
                        <td class="text-right">33,890,000</td>
                        <td class="text-right"></td>
                    </tr>
                    <tr class="total-row">
                        <td colspan="2"><strong>TOTALES</strong></td>
                        <td class="text-right"><strong>77,890,000</strong></td>
                        <td class="text-right"><strong>0</strong></td>
                        <td class="text-right"><strong>7,010,000</strong></td>
                        <td class="text-right"><strong>11,760,000</strong></td>
                        <td class="text-right"><strong>73,140,000</strong></td>
                        <td class="text-right"><strong>0</strong></td>
                    </tr>
                </tbody>
            </table>
        </div>
    `;
}

/**
 * Show report preview
 */
function showReportPreview(reportData) {
    document.getElementById('reportPreview').style.display = 'block';
    document.getElementById('reportTitle').textContent = reportData.title;
    
    const content = document.getElementById('reportContent');
    content.innerHTML = `
        <div class="report-header">
            <h2 class="text-center">${reportData.organization}</h2>
            <h3 class="text-center">${reportData.title}</h3>
            <p class="text-center">Período: ${reportData.period}</p>
            <hr>
        </div>
        
        <div class="report-body">
            ${reportData.content}
        </div>
        
        <div class="report-footer">
            <hr>
            <p class="text-center text-muted">
                Generado el ${formatDateTime(reportData.generatedAt)} por ${getCurrentUser()}
            </p>
        </div>
    `;
    
    // Scroll to preview
    document.getElementById('reportPreview').scrollIntoView({ behavior: 'smooth' });
}

/**
 * Print report
 */
function printReport() {
    window.print();
}

/**
 * Export report
 */
async function exportReport(format) {
    if (!format) {
        format = document.querySelector('input[name="format"]:checked').value;
    }
    
    showLoading(`Exportando reporte en formato ${format.toUpperCase()}...`);
    
    try {
        // Simulate export
        await new Promise(resolve => setTimeout(resolve, 1500));
        
        // In real implementation, generate and download file
        const filename = `${state.selectedReport}_${new Date().toISOString().split('T')[0]}.${format}`;
        
        showSuccess(`Reporte exportado: ${filename}`);
        
        // Simulate download
        const a = document.createElement('a');
        a.href = '#';
        a.download = filename;
        a.click();
        
    } catch (error) {
        showError('Error al exportar el reporte');
    } finally {
        hideLoading();
    }
}

/**
 * Save report
 */
async function saveReport() {
    const reportName = prompt('Nombre del reporte:');
    if (!reportName) return;
    
    showLoading('Guardando reporte...');
    
    try {
        await new Promise(resolve => setTimeout(resolve, 1000));
        showSuccess('Reporte guardado correctamente');
    } catch (error) {
        showError('Error al guardar el reporte');
    } finally {
        hideLoading();
    }
}

/**
 * Email report
 */
async function emailReport() {
    const email = prompt('Correo electrónico del destinatario:');
    if (!email) return;
    
    showLoading('Enviando reporte por correo...');
    
    try {
        await new Promise(resolve => setTimeout(resolve, 1500));
        showSuccess(`Reporte enviado a ${email}`);
    } catch (error) {
        showError('Error al enviar el reporte');
    } finally {
        hideLoading();
    }
}

/**
 * Reset report configuration
 */
function resetReport() {
    state.selectedReport = null;
    document.querySelectorAll('.report-card').forEach(card => {
        card.classList.remove('selected');
    });
    document.getElementById('reportConfig').style.display = 'none';
    document.getElementById('reportPreview').style.display = 'none';
}

/**
 * Filter reports based on search
 */
function filterReports(event) {
    const searchTerm = event.target.value.toLowerCase();
    
    document.querySelectorAll('.report-card').forEach(card => {
        const title = card.querySelector('.report-title').textContent.toLowerCase();
        const description = card.querySelector('.report-description').textContent.toLowerCase();
        
        if (title.includes(searchTerm) || description.includes(searchTerm)) {
            card.style.display = '';
        } else {
            card.style.display = 'none';
        }
    });
}

/**
 * Validate period
 */
function validatePeriod() {
    const dateFrom = new Date(document.getElementById('dateFrom').value);
    const dateTo = new Date(document.getElementById('dateTo').value);
    
    if (dateFrom > dateTo) {
        showError('La fecha inicial no puede ser mayor que la fecha final');
        document.getElementById('dateFrom').value = document.getElementById('dateTo').value;
    }
}

/**
 * Set period defaults
 */
function setPeriodDefaults() {
    const today = new Date();
    const firstDay = new Date(today.getFullYear(), today.getMonth(), 1);
    const lastDay = new Date(today.getFullYear(), today.getMonth() + 1, 0);
    
    document.getElementById('dateFrom').value = firstDay.toISOString().split('T')[0];
    document.getElementById('dateTo').value = lastDay.toISOString().split('T')[0];
}

/**
 * Show schedule modal
 */
function showScheduleModal() {
    document.getElementById('scheduleModal').style.display = 'block';
}

/**
 * Close schedule modal
 */
function closeScheduleModal() {
    document.getElementById('scheduleModal').style.display = 'none';
}

/**
 * Save schedule
 */
async function saveSchedule() {
    const reportType = document.getElementById('scheduleReportType').value;
    if (!reportType) {
        showError('Por favor seleccione un tipo de reporte');
        return;
    }
    
    showLoading('Guardando programación...');
    
    try {
        await new Promise(resolve => setTimeout(resolve, 1000));
        closeScheduleModal();
        showSuccess('Programación guardada correctamente');
    } catch (error) {
        showError('Error al guardar la programación');
    } finally {
        hideLoading();
    }
}

/**
 * Show saved reports
 */
function showSavedReports() {
    alert('Funcionalidad de reportes guardados próximamente');
}

// Helper functions
function getCurrentUser() {
    return 'Juan Pérez';
}

function formatDate(dateStr) {
    const date = new Date(dateStr);
    return date.toLocaleDateString('es-CO', { 
        year: 'numeric', 
        month: 'long', 
        day: 'numeric' 
    });
}

function formatDateTime(dateStr) {
    const date = new Date(dateStr);
    return date.toLocaleString('es-CO');
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

// Add print styles
const style = document.createElement('style');
style.textContent = `
    @media print {
        .report-header h2,
        .report-header h3 {
            margin: 10px 0;
        }
        
        .section-header td {
            background-color: #f5f5f5 !important;
            font-weight: bold;
        }
        
        .total-row td {
            border-top: 2px solid #000;
            border-bottom: 3px double #000;
            font-weight: bold;
        }
        
        table {
            width: 100%;
            border-collapse: collapse;
        }
        
        table td,
        table th {
            border: 1px solid #ddd;
            padding: 8px;
        }
        
        .report-preview {
            border: none;
            padding: 0;
        }
    }
`;
document.head.appendChild(style);