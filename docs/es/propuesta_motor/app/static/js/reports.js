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
    // Simular datos
    const totalActivos = 125450000;
    const totalPasivos = 45000000;
    const totalPatrimonio = 80450000;
    const liquidez = 3.42;
    const endeudamiento = 0.36;
    const solvencia = 2.79;
    
    return `
        <div class="balance-general">
            <!-- KPI Cards -->
            <div class="kpi-cards">
                <div class="kpi-card">
                    <div class="kpi-icon">💰</div>
                    <div class="kpi-value">$125.5M</div>
                    <div class="kpi-label">Activos Totales</div>
                    <div class="kpi-change positive">+15.3% vs año anterior</div>
                </div>
                <div class="kpi-card">
                    <div class="kpi-icon">📊</div>
                    <div class="kpi-value">$80.5M</div>
                    <div class="kpi-label">Patrimonio Neto</div>
                    <div class="kpi-change positive">+22.1% vs año anterior</div>
                </div>
                <div class="kpi-card">
                    <div class="kpi-icon">⚖️</div>
                    <div class="kpi-value">3.42</div>
                    <div class="kpi-label">Razón de Liquidez</div>
                    <div class="kpi-change positive">Excelente</div>
                </div>
                <div class="kpi-card">
                    <div class="kpi-icon">📈</div>
                    <div class="kpi-value">36%</div>
                    <div class="kpi-label">Nivel de Endeudamiento</div>
                    <div class="kpi-change positive">Saludable</div>
                </div>
            </div>

            <!-- Chart Section -->
            <div class="chart-container">
                <div class="chart-header">
                    <h3 class="chart-title">Composición del Balance</h3>
                    <div class="chart-options">
                        <button class="chart-option active">Gráfico</button>
                        <button class="chart-option" onclick="toggleBalanceView('table')">Tabla</button>
                    </div>
                </div>
                <canvas id="balanceChart" width="400" height="200"></canvas>
            </div>

            <!-- Detailed Tables -->
            <div class="report-body" style="padding: 30px;">
                <h4 style="color: #2d3748; margin-bottom: 20px;">📊 ESTADO DE SITUACIÓN FINANCIERA</h4>
                
                <table class="report-table">
                    <thead>
                        <tr>
                            <th colspan="3">ACTIVOS</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr class="section-header">
                            <td colspan="2">ACTIVO CORRIENTE</td>
                            <td class="text-right">$85,450,000</td>
                        </tr>
                        <tr>
                            <td width="50"></td>
                            <td>
                                <span class="account-code">1105</span>
                                Efectivo y Equivalentes
                                <div class="progress-bar-container">
                                    <div class="progress-bar" style="width: 57%"></div>
                                </div>
                            </td>
                            <td class="text-right positive-value">$48,450,000</td>
                        </tr>
                        <tr>
                            <td></td>
                            <td>
                                <span class="account-code">1305</span>
                                Cuentas por Cobrar
                                <div class="progress-bar-container">
                                    <div class="progress-bar" style="width: 41%"></div>
                                </div>
                            </td>
                            <td class="text-right positive-value">$35,000,000</td>
                        </tr>
                        <tr>
                            <td></td>
                            <td>
                                <span class="account-code">1435</span>
                                Inventarios
                                <div class="progress-bar-container">
                                    <div class="progress-bar" style="width: 2%"></div>
                                </div>
                            </td>
                            <td class="text-right positive-value">$2,000,000</td>
                        </tr>
                        <tr class="section-header">
                            <td colspan="2">ACTIVO NO CORRIENTE</td>
                            <td class="text-right">$40,000,000</td>
                        </tr>
                        <tr>
                            <td></td>
                            <td>
                                <span class="account-code">1520</span>
                                Propiedad, Planta y Equipo
                                <div class="interactive-metric">
                                    <div class="progress-bar-container">
                                        <div class="progress-bar" style="width: 100%"></div>
                                    </div>
                                    <div class="metric-tooltip">Incluye edificios, maquinaria y vehículos</div>
                                </div>
                            </td>
                            <td class="text-right positive-value">$40,000,000</td>
                        </tr>
                        <tr class="total-row">
                            <td colspan="2">TOTAL ACTIVOS</td>
                            <td class="text-right">$125,450,000</td>
                        </tr>
                    </tbody>
                </table>
                
                <table class="report-table" style="margin-top: 40px;">
                    <thead>
                        <tr>
                            <th colspan="3">PASIVOS Y PATRIMONIO</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr class="section-header">
                            <td colspan="2">PASIVO CORRIENTE</td>
                            <td class="text-right">$25,000,000</td>
                        </tr>
                        <tr>
                            <td width="50"></td>
                            <td>
                                <span class="account-code">2205</span>
                                Proveedores
                                <div class="progress-bar-container">
                                    <div class="progress-bar" style="width: 48%; background: linear-gradient(90deg, #f56565 0%, #e53e3e 100%);"></div>
                                </div>
                            </td>
                            <td class="text-right negative-value">$12,000,000</td>
                        </tr>
                        <tr>
                            <td></td>
                            <td>
                                <span class="account-code">2505</span>
                                Obligaciones Laborales
                                <div class="progress-bar-container">
                                    <div class="progress-bar" style="width: 32%; background: linear-gradient(90deg, #f56565 0%, #e53e3e 100%);"></div>
                                </div>
                            </td>
                            <td class="text-right negative-value">$8,000,000</td>
                        </tr>
                        <tr>
                            <td></td>
                            <td>
                                <span class="account-code">2408</span>
                                Impuestos por Pagar
                                <div class="progress-bar-container">
                                    <div class="progress-bar" style="width: 20%; background: linear-gradient(90deg, #f56565 0%, #e53e3e 100%);"></div>
                                </div>
                            </td>
                            <td class="text-right negative-value">$5,000,000</td>
                        </tr>
                        <tr class="section-header">
                            <td colspan="2">PASIVO NO CORRIENTE</td>
                            <td class="text-right">$20,000,000</td>
                        </tr>
                        <tr>
                            <td></td>
                            <td>
                                <span class="account-code">2105</span>
                                Obligaciones Financieras
                                <div class="progress-bar-container">
                                    <div class="progress-bar" style="width: 100%; background: linear-gradient(90deg, #f56565 0%, #e53e3e 100%);"></div>
                                </div>
                            </td>
                            <td class="text-right negative-value">$20,000,000</td>
                        </tr>
                        <tr class="total-row" style="background: linear-gradient(90deg, #fed7d7 0%, #feb2b2 100%);">
                            <td colspan="2">TOTAL PASIVOS</td>
                            <td class="text-right">$45,000,000</td>
                        </tr>
                        <tr class="section-header">
                            <td colspan="2">PATRIMONIO</td>
                            <td class="text-right">$80,450,000</td>
                        </tr>
                        <tr>
                            <td></td>
                            <td>
                                <span class="account-code">3105</span>
                                Capital Social
                            </td>
                            <td class="text-right positive-value">$30,000,000</td>
                        </tr>
                        <tr>
                            <td></td>
                            <td>
                                <span class="account-code">3305</span>
                                Reservas
                            </td>
                            <td class="text-right positive-value">$10,000,000</td>
                        </tr>
                        <tr>
                            <td></td>
                            <td>
                                <span class="account-code">3605</span>
                                Utilidades Retenidas
                            </td>
                            <td class="text-right positive-value">$20,450,000</td>
                        </tr>
                        <tr>
                            <td></td>
                            <td>
                                <span class="account-code">3605</span>
                                Utilidad del Ejercicio
                                <span style="margin-left: 10px; color: #38a169; font-size: 0.875rem;">⬆ 22%</span>
                            </td>
                            <td class="text-right positive-value">$20,000,000</td>
                        </tr>
                        <tr class="total-row">
                            <td colspan="2">TOTAL PASIVOS Y PATRIMONIO</td>
                            <td class="text-right">$125,450,000</td>
                        </tr>
                    </tbody>
                </table>

                <!-- Financial Ratios -->
                <div class="chart-container" style="margin-top: 40px;">
                    <h3 class="chart-title">📊 Indicadores Financieros Clave</h3>
                    <div class="comparison-container" style="margin-top: 20px;">
                        <div class="comparison-column">
                            <div class="comparison-header">Liquidez y Solvencia</div>
                            <div class="comparison-metric">
                                <span class="metric-label">Razón Corriente</span>
                                <span class="metric-value positive-value">3.42</span>
                            </div>
                            <div class="comparison-metric">
                                <span class="metric-label">Prueba Ácida</span>
                                <span class="metric-value positive-value">3.34</span>
                            </div>
                            <div class="comparison-metric">
                                <span class="metric-label">Capital de Trabajo</span>
                                <span class="metric-value positive-value">$60.5M</span>
                            </div>
                        </div>
                        <div class="comparison-column">
                            <div class="comparison-header">Estructura Financiera</div>
                            <div class="comparison-metric">
                                <span class="metric-label">Endeudamiento Total</span>
                                <span class="metric-value">36%</span>
                            </div>
                            <div class="comparison-metric">
                                <span class="metric-label">Autonomía Financiera</span>
                                <span class="metric-value positive-value">64%</span>
                            </div>
                            <div class="comparison-metric">
                                <span class="metric-label">Leverage</span>
                                <span class="metric-value">0.56</span>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Report Actions -->
            <div class="report-actions">
                <button class="report-action" onclick="downloadBalanceDetail()">
                    📥 Descargar Detalle
                </button>
                <button class="report-action" onclick="showBalanceChart()">
                    📊 Ver Gráficos
                </button>
                <button class="report-action" onclick="compareWithPrevious()">
                    🔄 Comparar Períodos
                </button>
                <button class="report-action primary" onclick="generateAnalysis()">
                    🤖 Generar Análisis con IA
                </button>
            </div>
        </div>

        <script>
            // Generar gráfico de composición del balance
            setTimeout(() => {
                const ctx = document.getElementById('balanceChart');
                if (ctx && typeof Chart !== 'undefined') {
                    new Chart(ctx, {
                        type: 'doughnut',
                        data: {
                            labels: ['Activo Corriente', 'Activo No Corriente', 'Pasivo Corriente', 'Pasivo No Corriente', 'Patrimonio'],
                            datasets: [{
                                data: [85450000, 40000000, 25000000, 20000000, 80450000],
                                backgroundColor: [
                                    'rgba(102, 126, 234, 0.8)',
                                    'rgba(118, 75, 162, 0.8)',
                                    'rgba(245, 101, 101, 0.8)',
                                    'rgba(229, 62, 62, 0.8)',
                                    'rgba(72, 187, 120, 0.8)'
                                ],
                                borderWidth: 2,
                                borderColor: '#fff'
                            }]
                        },
                        options: {
                            responsive: true,
                            maintainAspectRatio: false,
                            plugins: {
                                legend: {
                                    position: 'right',
                                    labels: {
                                        padding: 20,
                                        font: {
                                            size: 14
                                        }
                                    }
                                },
                                tooltip: {
                                    callbacks: {
                                        label: function(context) {
                                            const label = context.label || '';
                                            const value = new Intl.NumberFormat('es-CO', {
                                                style: 'currency',
                                                currency: 'COP',
                                                minimumFractionDigits: 0
                                            }).format(context.parsed);
                                            const percentage = ((context.parsed / 251850000) * 100).toFixed(1);
                                            return label + ': ' + value + ' (' + percentage + '%)';
                                        }
                                    }
                                }
                            }
                        }
                    });
                }
            }, 100);
        </script>
    `;
}

/**
 * Generate Estado de Resultados content
 */
function generateEstadoResultados() {
    const showPercentages = document.getElementById('showPercentages')?.checked;
    
    return `
        <div class="estado-resultados">
            <!-- KPI Cards -->
            <div class="kpi-cards">
                <div class="kpi-card">
                    <div class="kpi-icon">💰</div>
                    <div class="kpi-value">$85M</div>
                    <div class="kpi-label">Ingresos Totales</div>
                    <div class="kpi-change positive">+18.5% vs período anterior</div>
                </div>
                <div class="kpi-card">
                    <div class="kpi-icon">📈</div>
                    <div class="kpi-value">$20M</div>
                    <div class="kpi-label">Utilidad Neta</div>
                    <div class="kpi-change positive">+25.0% vs período anterior</div>
                </div>
                <div class="kpi-card">
                    <div class="kpi-icon">📊</div>
                    <div class="kpi-value">88.2%</div>
                    <div class="kpi-label">Margen Bruto</div>
                    <div class="kpi-change positive">+2.3 puntos</div>
                </div>
                <div class="kpi-card">
                    <div class="kpi-icon">💹</div>
                    <div class="kpi-value">23.5%</div>
                    <div class="kpi-label">Margen Neto</div>
                    <div class="kpi-change positive">+1.5 puntos</div>
                </div>
            </div>

            <!-- Waterfall Chart -->
            <div class="chart-container">
                <div class="chart-header">
                    <h3 class="chart-title">Cascada de Resultados</h3>
                    <div class="chart-options">
                        <button class="chart-option active">Cascada</button>
                        <button class="chart-option" onclick="togglePnLView('trend')">Tendencia</button>
                        <button class="chart-option" onclick="togglePnLView('composition')">Composición</button>
                    </div>
                </div>
                <canvas id="waterfallChart" width="400" height="250"></canvas>
            </div>

            <!-- Detailed P&L -->
            <div class="report-body" style="padding: 30px;">
                <h4 style="color: #2d3748; margin-bottom: 20px;">💰 ESTADO DE RESULTADOS INTEGRAL</h4>
                
                <table class="report-table">
                    <thead>
                        <tr>
                            <th>Concepto</th>
                            <th class="text-right">Valor</th>
                            ${showPercentages ? '<th class="text-right">% Ventas</th>' : ''}
                            <th class="text-right">Variación</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr class="section-header">
                            <td>
                                <span class="account-code">4</span>
                                INGRESOS OPERACIONALES
                            </td>
                            <td class="text-right positive-value">$85,000,000</td>
                            ${showPercentages ? '<td class="text-right">100.0%</td>' : ''}
                            <td class="text-right">
                                <span style="color: #38a169;">⬆ 18.5%</span>
                            </td>
                        </tr>
                        <tr>
                            <td style="padding-left: 30px;">
                                <span class="account-code">4135</span>
                                Ventas de Servicios
                                <div class="progress-bar-container">
                                    <div class="progress-bar" style="width: 59%"></div>
                                </div>
                            </td>
                            <td class="text-right">$50,000,000</td>
                            ${showPercentages ? '<td class="text-right">58.8%</td>' : ''}
                            <td class="text-right">
                                <span style="color: #38a169; font-size: 0.875rem;">+22.0%</span>
                            </td>
                        </tr>
                        <tr>
                            <td style="padding-left: 30px;">
                                <span class="account-code">4175</span>
                                Ventas de Productos
                                <div class="progress-bar-container">
                                    <div class="progress-bar" style="width: 41%"></div>
                                </div>
                            </td>
                            <td class="text-right">$35,000,000</td>
                            ${showPercentages ? '<td class="text-right">41.2%</td>' : ''}
                            <td class="text-right">
                                <span style="color: #38a169; font-size: 0.875rem;">+13.5%</span>
                            </td>
                        </tr>
                        
                        <tr class="section-header">
                            <td>
                                <span class="account-code">6</span>
                                (-) COSTO DE VENTAS
                            </td>
                            <td class="text-right negative-value">$(10,000,000)</td>
                            ${showPercentages ? '<td class="text-right">11.8%</td>' : ''}
                            <td class="text-right">
                                <span style="color: #e53e3e;">⬆ 8.2%</span>
                            </td>
                        </tr>
                        
                        <tr class="total-row" style="background: linear-gradient(90deg, #c6f6d5 0%, #9ae6b4 100%);">
                            <td>UTILIDAD BRUTA</td>
                            <td class="text-right"><strong>$75,000,000</strong></td>
                            ${showPercentages ? '<td class="text-right"><strong>88.2%</strong></td>' : ''}
                            <td class="text-right">
                                <strong style="color: #276749;">+19.8%</strong>
                            </td>
                        </tr>
                        
                        <tr class="section-header">
                            <td>
                                <span class="account-code">5</span>
                                (-) GASTOS OPERACIONALES
                            </td>
                            <td class="text-right negative-value">$(55,000,000)</td>
                            ${showPercentages ? '<td class="text-right">64.7%</td>' : ''}
                            <td class="text-right">
                                <span style="color: #e53e3e;">⬆ 15.8%</span>
                            </td>
                        </tr>
                        <tr>
                            <td style="padding-left: 30px;">
                                <span class="account-code">51</span>
                                Gastos de Administración
                                <div class="progress-bar-container">
                                    <div class="progress-bar" style="width: 82%; background: linear-gradient(90deg, #fc8181 0%, #f56565 100%);"></div>
                                </div>
                            </td>
                            <td class="text-right negative-value">$(45,000,000)</td>
                            ${showPercentages ? '<td class="text-right">52.9%</td>' : ''}
                            <td class="text-right">
                                <span style="color: #e53e3e; font-size: 0.875rem;">+12.5%</span>
                            </td>
                        </tr>
                        <tr>
                            <td style="padding-left: 30px;">
                                <span class="account-code">52</span>
                                Gastos de Ventas
                                <div class="progress-bar-container">
                                    <div class="progress-bar" style="width: 18%; background: linear-gradient(90deg, #fc8181 0%, #f56565 100%);"></div>
                                </div>
                            </td>
                            <td class="text-right negative-value">$(10,000,000)</td>
                            ${showPercentages ? '<td class="text-right">11.8%</td>' : ''}
                            <td class="text-right">
                                <span style="color: #e53e3e; font-size: 0.875rem;">+30.2%</span>
                            </td>
                        </tr>
                        
                        <tr class="total-row" style="background: linear-gradient(90deg, #bee3f8 0%, #90cdf4 100%);">
                            <td>UTILIDAD OPERACIONAL</td>
                            <td class="text-right"><strong>$20,000,000</strong></td>
                            ${showPercentages ? '<td class="text-right"><strong>23.5%</strong></td>' : ''}
                            <td class="text-right">
                                <strong style="color: #2b6cb0;">+33.3%</strong>
                            </td>
                        </tr>
                        
                        <tr>
                            <td>
                                <span class="account-code">42</span>
                                (+) Otros Ingresos
                            </td>
                            <td class="text-right">$0</td>
                            ${showPercentages ? '<td class="text-right">0.0%</td>' : ''}
                            <td class="text-right">-</td>
                        </tr>
                        <tr>
                            <td>
                                <span class="account-code">53</span>
                                (-) Otros Gastos
                            </td>
                            <td class="text-right">$0</td>
                            ${showPercentages ? '<td class="text-right">0.0%</td>' : ''}
                            <td class="text-right">-</td>
                        </tr>
                        
                        <tr class="total-row">
                            <td>UTILIDAD ANTES DE IMPUESTOS</td>
                            <td class="text-right"><strong>$20,000,000</strong></td>
                            ${showPercentages ? '<td class="text-right"><strong>23.5%</strong></td>' : ''}
                            <td class="text-right">
                                <strong>+33.3%</strong>
                            </td>
                        </tr>
                        
                        <tr>
                            <td>
                                <span class="account-code">54</span>
                                (-) Impuesto de Renta
                            </td>
                            <td class="text-right">$0</td>
                            ${showPercentages ? '<td class="text-right">0.0%</td>' : ''}
                            <td class="text-right">-</td>
                        </tr>
                        
                        <tr class="total-row" style="background: linear-gradient(135deg, #68d391 0%, #38a169 100%); color: white;">
                            <td><strong>UTILIDAD NETA DEL EJERCICIO</strong></td>
                            <td class="text-right"><strong>$20,000,000</strong></td>
                            ${showPercentages ? '<td class="text-right"><strong>23.5%</strong></td>' : ''}
                            <td class="text-right">
                                <strong>+25.0%</strong>
                            </td>
                        </tr>
                    </tbody>
                </table>

                <!-- Margin Analysis -->
                <div class="comparison-container" style="margin-top: 40px;">
                    <div class="comparison-column">
                        <div class="comparison-header">Análisis de Márgenes</div>
                        <div class="comparison-metric">
                            <span class="metric-label">Margen Bruto</span>
                            <div class="interactive-metric">
                                <div class="progress-bar-container">
                                    <div class="progress-bar" style="width: 88.2%"></div>
                                </div>
                                <span class="metric-value positive-value">88.2%</span>
                            </div>
                        </div>
                        <div class="comparison-metric">
                            <span class="metric-label">Margen Operacional</span>
                            <div class="interactive-metric">
                                <div class="progress-bar-container">
                                    <div class="progress-bar" style="width: 23.5%"></div>
                                </div>
                                <span class="metric-value">23.5%</span>
                            </div>
                        </div>
                        <div class="comparison-metric">
                            <span class="metric-label">Margen Neto</span>
                            <div class="interactive-metric">
                                <div class="progress-bar-container">
                                    <div class="progress-bar" style="width: 23.5%"></div>
                                </div>
                                <span class="metric-value positive-value">23.5%</span>
                            </div>
                        </div>
                    </div>
                    <div class="comparison-column">
                        <div class="comparison-header">Eficiencia Operativa</div>
                        <div class="comparison-metric">
                            <span class="metric-label">Gastos Admin / Ventas</span>
                            <span class="metric-value">52.9%</span>
                        </div>
                        <div class="comparison-metric">
                            <span class="metric-label">Gastos Ventas / Ventas</span>
                            <span class="metric-value">11.8%</span>
                        </div>
                        <div class="comparison-metric">
                            <span class="metric-label">EBITDA Estimado</span>
                            <span class="metric-value positive-value">$22.5M</span>
                        </div>
                    </div>
                </div>

                <!-- Trend Chart -->
                <div class="chart-container" style="margin-top: 40px;">
                    <h3 class="chart-title">📈 Tendencia de Ingresos (Últimos 12 meses)</h3>
                    <canvas id="trendChart" width="400" height="150"></canvas>
                </div>
            </div>

            <!-- Report Actions -->
            <div class="report-actions">
                <button class="report-action" onclick="downloadPnLDetail()">
                    📥 Descargar Detalle
                </button>
                <button class="report-action" onclick="showPnLBreakdown()">
                    📊 Desglose por Centro
                </button>
                <button class="report-action" onclick="compareYearOverYear()">
                    📅 Comparar Años
                </button>
                <button class="report-action primary" onclick="generatePnLInsights()">
                    💡 Generar Insights
                </button>
            </div>
        </div>

        <script>
            // Generar gráfico de cascada
            setTimeout(() => {
                const ctx = document.getElementById('waterfallChart');
                if (ctx && typeof Chart !== 'undefined') {
                    new Chart(ctx, {
                        type: 'bar',
                        data: {
                            labels: ['Ingresos', 'Costo Ventas', 'Margen Bruto', 'Gastos Admin', 'Gastos Ventas', 'Utilidad Neta'],
                            datasets: [{
                                label: 'Flujo',
                                data: [85000000, -10000000, 75000000, -45000000, -10000000, 20000000],
                                backgroundColor: [
                                    'rgba(72, 187, 120, 0.8)',
                                    'rgba(245, 101, 101, 0.8)',
                                    'rgba(102, 126, 234, 0.8)',
                                    'rgba(245, 101, 101, 0.8)',
                                    'rgba(245, 101, 101, 0.8)',
                                    'rgba(72, 187, 120, 0.8)'
                                ],
                                borderColor: [
                                    'rgba(72, 187, 120, 1)',
                                    'rgba(245, 101, 101, 1)',
                                    'rgba(102, 126, 234, 1)',
                                    'rgba(245, 101, 101, 1)',
                                    'rgba(245, 101, 101, 1)',
                                    'rgba(72, 187, 120, 1)'
                                ],
                                borderWidth: 2
                            }]
                        },
                        options: {
                            responsive: true,
                            maintainAspectRatio: false,
                            plugins: {
                                legend: {
                                    display: false
                                },
                                tooltip: {
                                    callbacks: {
                                        label: function(context) {
                                            const value = new Intl.NumberFormat('es-CO', {
                                                style: 'currency',
                                                currency: 'COP',
                                                minimumFractionDigits: 0
                                            }).format(Math.abs(context.parsed.y));
                                            return context.parsed.y < 0 ? '-' + value : value;
                                        }
                                    }
                                }
                            },
                            scales: {
                                y: {
                                    beginAtZero: true,
                                    ticks: {
                                        callback: function(value) {
                                            return new Intl.NumberFormat('es-CO', {
                                                style: 'currency',
                                                currency: 'COP',
                                                minimumFractionDigits: 0,
                                                maximumFractionDigits: 0
                                            }).format(value);
                                        }
                                    }
                                }
                            }
                        }
                    });
                }

                // Gráfico de tendencia
                const trendCtx = document.getElementById('trendChart');
                if (trendCtx && typeof Chart !== 'undefined') {
                    new Chart(trendCtx, {
                        type: 'line',
                        data: {
                            labels: ['Ene', 'Feb', 'Mar', 'Abr', 'May', 'Jun', 'Jul', 'Ago', 'Sep', 'Oct', 'Nov', 'Dic'],
                            datasets: [{
                                label: 'Ingresos',
                                data: [65, 68, 70, 72, 75, 78, 80, 82, 83, 84, 85, 85],
                                borderColor: 'rgba(102, 126, 234, 1)',
                                backgroundColor: 'rgba(102, 126, 234, 0.1)',
                                tension: 0.3
                            }, {
                                label: 'Utilidad',
                                data: [12, 13, 14, 15, 16, 17, 18, 19, 19.5, 20, 20, 20],
                                borderColor: 'rgba(72, 187, 120, 1)',
                                backgroundColor: 'rgba(72, 187, 120, 0.1)',
                                tension: 0.3
                            }]
                        },
                        options: {
                            responsive: true,
                            maintainAspectRatio: false,
                            scales: {
                                y: {
                                    beginAtZero: true,
                                    ticks: {
                                        callback: function(value) {
                                            return '$' + value + 'M';
                                        }
                                    }
                                }
                            }
                        }
                    });
                }
            }, 100);
        </script>
    `;
}

/**
 * Generate Libro Diario content
 */
function generateLibroDiario() {
    return `
        <div class="libro-diario">
            <!-- KPI Cards -->
            <div class="kpi-cards">
                <div class="kpi-card">
                    <div class="kpi-icon">📚</div>
                    <div class="kpi-value">245</div>
                    <div class="kpi-label">Asientos del Período</div>
                    <div class="kpi-change positive">+15 vs período anterior</div>
                </div>
                <div class="kpi-card">
                    <div class="kpi-icon">✅</div>
                    <div class="kpi-value">100%</div>
                    <div class="kpi-label">Asientos Balanceados</div>
                    <div class="kpi-change positive">Todos correctos</div>
                </div>
                <div class="kpi-card">
                    <div class="kpi-icon">💰</div>
                    <div class="kpi-value">$450M</div>
                    <div class="kpi-label">Total Movimientos</div>
                    <div class="kpi-change positive">+22.5% vs período anterior</div>
                </div>
                <div class="kpi-card">
                    <div class="kpi-icon">📊</div>
                    <div class="kpi-value">52</div>
                    <div class="kpi-label">Cuentas Activas</div>
                    <div class="kpi-change">Sin cambios</div>
                </div>
            </div>

            <!-- Chart Section -->
            <div class="chart-container">
                <div class="chart-header">
                    <h3 class="chart-title">📈 Asientos por Día del Mes</h3>
                    <div class="chart-options">
                        <button class="chart-option active">Por Día</button>
                        <button class="chart-option">Por Tipo</button>
                        <button class="chart-option">Por Usuario</button>
                    </div>
                </div>
                <canvas id="diarioChart" width="400" height="150"></canvas>
            </div>

            <!-- Journal Entries Table -->
            <div class="report-body" style="padding: 30px;">
                <h4 style="color: #2d3748; margin-bottom: 20px;">📚 LIBRO DIARIO - ENERO 2024</h4>
                
                <table class="report-table">
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
                        <tr class="section-header">
                            <td colspan="6">📅 21 de Enero, 2024</td>
                        </tr>
                        <tr>
                            <td rowspan="3">21/01/2024</td>
                            <td rowspan="3">
                                <span class="account-code">AS-2024-0127</span>
                                <br>
                                <span style="font-size: 0.8rem; color: #718096;">
                                    <em>FV-2024-0127</em>
                                </span>
                            </td>
                            <td><span class="account-code">1305.05</span></td>
                            <td>Clientes Nacionales - ABC Tecnología S.A.S</td>
                            <td class="text-right positive-value">$2,450,000</td>
                            <td class="text-right">-</td>
                        </tr>
                        <tr>
                            <td><span class="account-code">4135.05</span></td>
                            <td>Ingresos por Servicios de Consultoría</td>
                            <td class="text-right">-</td>
                            <td class="text-right negative-value">$2,058,824</td>
                        </tr>
                        <tr>
                            <td><span class="account-code">2408.01</span></td>
                            <td>IVA por Pagar 19%</td>
                            <td class="text-right">-</td>
                            <td class="text-right negative-value">$391,176</td>
                        </tr>
                        <tr style="background: #f7fafc;">
                            <td colspan="4" class="text-right" style="font-style: italic;">Sumas iguales</td>
                            <td class="text-right" style="border-top: 2px solid #cbd5e0;"><strong>$2,450,000</strong></td>
                            <td class="text-right" style="border-top: 2px solid #cbd5e0;"><strong>$2,450,000</strong></td>
                        </tr>
                        
                        <tr>
                            <td rowspan="2">21/01/2024</td>
                            <td rowspan="2">
                                <span class="account-code">AS-2024-0128</span>
                                <br>
                                <span style="font-size: 0.8rem; color: #718096;">
                                    <em>RC-2024-0045</em>
                                </span>
                            </td>
                            <td><span class="account-code">1110.05</span></td>
                            <td>Bancos - Banco Nacional</td>
                            <td class="text-right positive-value">$560,000</td>
                            <td class="text-right">-</td>
                        </tr>
                        <tr>
                            <td><span class="account-code">1305.05</span></td>
                            <td>Clientes Nacionales - Pago Juan Pérez</td>
                            <td class="text-right">-</td>
                            <td class="text-right negative-value">$560,000</td>
                        </tr>
                        <tr style="background: #f7fafc;">
                            <td colspan="4" class="text-right" style="font-style: italic;">Sumas iguales</td>
                            <td class="text-right" style="border-top: 2px solid #cbd5e0;"><strong>$560,000</strong></td>
                            <td class="text-right" style="border-top: 2px solid #cbd5e0;"><strong>$560,000</strong></td>
                        </tr>
                        
                        <tr class="total-row">
                            <td colspan="4">TOTALES DEL DÍA</td>
                            <td class="text-right"><strong>$3,010,000</strong></td>
                            <td class="text-right"><strong>$3,010,000</strong></td>
                        </tr>
                    </tbody>
                </table>

                <!-- Summary Section -->
                <div class="comparison-container" style="margin-top: 40px;">
                    <div class="comparison-column">
                        <div class="comparison-header">Resumen por Tipo de Asiento</div>
                        <div class="comparison-metric">
                            <span class="metric-label">Facturas de Venta</span>
                            <span class="metric-value">125</span>
                        </div>
                        <div class="comparison-metric">
                            <span class="metric-label">Recibos de Caja</span>
                            <span class="metric-value">68</span>
                        </div>
                        <div class="comparison-metric">
                            <span class="metric-label">Compras</span>
                            <span class="metric-value">32</span>
                        </div>
                        <div class="comparison-metric">
                            <span class="metric-label">Ajustes</span>
                            <span class="metric-value">20</span>
                        </div>
                    </div>
                    <div class="comparison-column">
                        <div class="comparison-header">Estadísticas del Período</div>
                        <div class="comparison-metric">
                            <span class="metric-label">Promedio diario de asientos</span>
                            <span class="metric-value">8.2</span>
                        </div>
                        <div class="comparison-metric">
                            <span class="metric-label">Mayor asiento del período</span>
                            <span class="metric-value">$45,890,000</span>
                        </div>
                        <div class="comparison-metric">
                            <span class="metric-label">Cuentas más utilizadas</span>
                            <span class="metric-value">1305, 1110, 4135</span>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Report Actions -->
            <div class="report-actions">
                <button class="report-action" onclick="filterByVoucherType()">
                    🔍 Filtrar por Tipo
                </button>
                <button class="report-action" onclick="exportJournalEntries()">
                    📥 Exportar Asientos
                </button>
                <button class="report-action primary" onclick="validateJournalIntegrity()">
                    ✅ Validar Integridad
                </button>
            </div>
        </div>

        <script>
            // Generar gráfico de asientos por día
            setTimeout(() => {
                const ctx = document.getElementById('diarioChart');
                if (ctx && typeof Chart !== 'undefined') {
                    new Chart(ctx, {
                        type: 'bar',
                        data: {
                            labels: Array.from({length: 31}, (_, i) => i + 1),
                            datasets: [{
                                label: 'Asientos',
                                data: Array.from({length: 31}, () => Math.floor(Math.random() * 15) + 3),
                                backgroundColor: 'rgba(102, 126, 234, 0.8)',
                                borderColor: 'rgba(102, 126, 234, 1)',
                                borderWidth: 1
                            }]
                        },
                        options: {
                            responsive: true,
                            maintainAspectRatio: false,
                            plugins: {
                                legend: {
                                    display: false
                                }
                            },
                            scales: {
                                y: {
                                    beginAtZero: true,
                                    ticks: {
                                        stepSize: 5
                                    }
                                }
                            }
                        }
                    });
                }
            }, 100);
        </script>
    `;
}

/**
 * Generate Libro Mayor content
 */
function generateLibroMayor() {
    const showRunningBalance = document.getElementById('showRunningBalance')?.checked !== false;
    
    return `
        <div class="libro-mayor">
            <!-- Account Selector -->
            <div class="chart-container" style="margin-bottom: 20px;">
                <div class="chart-header">
                    <h3 class="chart-title">📖 Seleccione una Cuenta</h3>
                    <div class="chart-options">
                        <select class="form-control" style="width: 300px;" onchange="updateLibroMayor(this.value)">
                            <option value="1305.05">1305.05 - Clientes Nacionales</option>
                            <option value="1110.05">1110.05 - Bancos</option>
                            <option value="4135.05">4135.05 - Ingresos por Servicios</option>
                            <option value="5105.05">5105.05 - Gastos de Personal</option>
                        </select>
                    </div>
                </div>
            </div>

            <!-- KPI Cards for Selected Account -->
            <div class="kpi-cards">
                <div class="kpi-card">
                    <div class="kpi-icon">💰</div>
                    <div class="kpi-value">$31.9M</div>
                    <div class="kpi-label">Saldo Actual</div>
                    <div class="kpi-change positive">Deudor</div>
                </div>
                <div class="kpi-card">
                    <div class="kpi-icon">📈</div>
                    <div class="kpi-value">$2.45M</div>
                    <div class="kpi-label">Total Débitos</div>
                    <div class="kpi-change">Período actual</div>
                </div>
                <div class="kpi-card">
                    <div class="kpi-icon">📉</div>
                    <div class="kpi-value">$560K</div>
                    <div class="kpi-label">Total Créditos</div>
                    <div class="kpi-change">Período actual</div>
                </div>
                <div class="kpi-card">
                    <div class="kpi-icon">📊</div>
                    <div class="kpi-value">15</div>
                    <div class="kpi-label">Movimientos</div>
                    <div class="kpi-change positive">+3 vs mes anterior</div>
                </div>
            </div>

            <!-- Evolution Chart -->
            <div class="chart-container">
                <div class="chart-header">
                    <h3 class="chart-title">📈 Evolución del Saldo</h3>
                    <div class="chart-options">
                        <button class="chart-option active">Línea</button>
                        <button class="chart-option">Área</button>
                        <button class="chart-option">Barras</button>
                    </div>
                </div>
                <canvas id="mayorChart" width="400" height="200"></canvas>
            </div>

            <!-- Account Movements -->
            <div class="report-body" style="padding: 30px;">
                <h4 style="color: #2d3748; margin-bottom: 20px;">
                    <span class="account-code">1305.05</span> CLIENTES NACIONALES
                </h4>
                
                <table class="report-table">
                    <thead>
                        <tr>
                            <th>Fecha</th>
                            <th>Asiento</th>
                            <th>Descripción</th>
                            <th class="text-right">Débito</th>
                            <th class="text-right">Crédito</th>
                            ${showRunningBalance ? '<th class="text-right">Saldo</th>' : ''}
                        </tr>
                    </thead>
                    <tbody>
                        <tr class="section-header">
                            <td colspan="${showRunningBalance ? '6' : '5'}">
                                SALDO ANTERIOR AL 01/01/2024
                            </td>
                        </tr>
                        <tr>
                            <td>01/01/2024</td>
                            <td><span class="account-code">SALDO</span></td>
                            <td>Saldo inicial del período</td>
                            <td class="text-right">-</td>
                            <td class="text-right">-</td>
                            ${showRunningBalance ? '<td class="text-right positive-value">$30,000,000</td>' : ''}
                        </tr>
                        
                        <tr class="section-header">
                            <td colspan="${showRunningBalance ? '6' : '5'}">
                                MOVIMIENTOS DE ENERO 2024
                            </td>
                        </tr>
                        
                        <tr>
                            <td>05/01/2024</td>
                            <td>
                                <span class="account-code">AS-2024-0105</span>
                                <br>
                                <span style="font-size: 0.8rem; color: #718096;">FV-2024-0105</span>
                            </td>
                            <td>Factura venta - XYZ Corp</td>
                            <td class="text-right positive-value">$1,250,000</td>
                            <td class="text-right">-</td>
                            ${showRunningBalance ? '<td class="text-right">$31,250,000</td>' : ''}
                        </tr>
                        
                        <tr>
                            <td>08/01/2024</td>
                            <td>
                                <span class="account-code">AS-2024-0112</span>
                                <br>
                                <span style="font-size: 0.8rem; color: #718096;">RC-2024-0032</span>
                            </td>
                            <td>Recibo de caja - Pago XYZ Corp</td>
                            <td class="text-right">-</td>
                            <td class="text-right negative-value">$1,250,000</td>
                            ${showRunningBalance ? '<td class="text-right">$30,000,000</td>' : ''}
                        </tr>
                        
                        <tr>
                            <td>21/01/2024</td>
                            <td>
                                <span class="account-code">AS-2024-0127</span>
                                <br>
                                <span style="font-size: 0.8rem; color: #718096;">FV-2024-0127</span>
                            </td>
                            <td>Factura venta - ABC Tecnología S.A.S</td>
                            <td class="text-right positive-value">$2,450,000</td>
                            <td class="text-right">-</td>
                            ${showRunningBalance ? '<td class="text-right">$32,450,000</td>' : ''}
                        </tr>
                        
                        <tr>
                            <td>21/01/2024</td>
                            <td>
                                <span class="account-code">AS-2024-0128</span>
                                <br>
                                <span style="font-size: 0.8rem; color: #718096;">RC-2024-0045</span>
                            </td>
                            <td>Recibo de caja - Abono Juan Pérez</td>
                            <td class="text-right">-</td>
                            <td class="text-right negative-value">$560,000</td>
                            ${showRunningBalance ? '<td class="text-right">$31,890,000</td>' : ''}
                        </tr>
                        
                        <tr class="total-row">
                            <td colspan="3">TOTALES DEL PERÍODO</td>
                            <td class="text-right"><strong>$3,700,000</strong></td>
                            <td class="text-right"><strong>$1,810,000</strong></td>
                            ${showRunningBalance ? '<td class="text-right"><strong>$31,890,000</strong></td>' : ''}
                        </tr>
                        
                        <tr style="background: linear-gradient(90deg, #e6fffa 0%, #b2f5ea 100%);">
                            <td colspan="3"><strong>SALDO FINAL</strong></td>
                            <td colspan="${showRunningBalance ? '3' : '2'}" class="text-right">
                                <strong style="font-size: 1.2rem; color: #234e52;">$31,890,000 (DEUDOR)</strong>
                            </td>
                        </tr>
                    </tbody>
                </table>

                <!-- Account Analysis -->
                <div class="comparison-container" style="margin-top: 40px;">
                    <div class="comparison-column">
                        <div class="comparison-header">Análisis de Movimientos</div>
                        <div class="comparison-metric">
                            <span class="metric-label">Rotación de cartera</span>
                            <span class="metric-value">45 días</span>
                        </div>
                        <div class="comparison-metric">
                            <span class="metric-label">Mayor factura</span>
                            <span class="metric-value">$2,450,000</span>
                        </div>
                        <div class="comparison-metric">
                            <span class="metric-label">Recaudo promedio</span>
                            <span class="metric-value">$905,000</span>
                        </div>
                        <div class="comparison-metric">
                            <span class="metric-label">Antigüedad promedio</span>
                            <span class="metric-value">18 días</span>
                        </div>
                    </div>
                    <div class="comparison-column">
                        <div class="comparison-header">Composición del Saldo</div>
                        <div class="comparison-metric">
                            <span class="metric-label">0-30 días</span>
                            <div class="progress-bar-container">
                                <div class="progress-bar" style="width: 75%"></div>
                            </div>
                            <span class="metric-value">$23,917,500</span>
                        </div>
                        <div class="comparison-metric">
                            <span class="metric-label">31-60 días</span>
                            <div class="progress-bar-container">
                                <div class="progress-bar" style="width: 20%"></div>
                            </div>
                            <span class="metric-value">$6,378,000</span>
                        </div>
                        <div class="comparison-metric">
                            <span class="metric-label">61-90 días</span>
                            <div class="progress-bar-container">
                                <div class="progress-bar" style="width: 5%"></div>
                            </div>
                            <span class="metric-value">$1,594,500</span>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Report Actions -->
            <div class="report-actions">
                <button class="report-action" onclick="showAccountDetails()">
                    📋 Detalles de Cuenta
                </button>
                <button class="report-action" onclick="showAgingAnalysis()">
                    📊 Análisis de Vencimientos
                </button>
                <button class="report-action" onclick="exportAccountMovements()">
                    📥 Exportar Movimientos
                </button>
                <button class="report-action primary" onclick="reconcileAccount()">
                    ✅ Conciliar Cuenta
                </button>
            </div>
        </div>

        <script>
            // Generar gráfico de evolución
            setTimeout(() => {
                const ctx = document.getElementById('mayorChart');
                if (ctx && typeof Chart !== 'undefined') {
                    new Chart(ctx, {
                        type: 'line',
                        data: {
                            labels: ['01/01', '05/01', '08/01', '15/01', '21/01', '31/01'],
                            datasets: [{
                                label: 'Saldo',
                                data: [30000000, 31250000, 30000000, 30000000, 31890000, 31890000],
                                borderColor: 'rgba(102, 126, 234, 1)',
                                backgroundColor: 'rgba(102, 126, 234, 0.1)',
                                tension: 0.4,
                                fill: true
                            }]
                        },
                        options: {
                            responsive: true,
                            maintainAspectRatio: false,
                            plugins: {
                                legend: {
                                    display: false
                                },
                                tooltip: {
                                    callbacks: {
                                        label: function(context) {
                                            return new Intl.NumberFormat('es-CO', {
                                                style: 'currency',
                                                currency: 'COP',
                                                minimumFractionDigits: 0
                                            }).format(context.parsed.y);
                                        }
                                    }
                                }
                            },
                            scales: {
                                y: {
                                    beginAtZero: false,
                                    ticks: {
                                        callback: function(value) {
                                            return new Intl.NumberFormat('es-CO', {
                                                style: 'currency',
                                                currency: 'COP',
                                                minimumFractionDigits: 0,
                                                maximumFractionDigits: 0
                                            }).format(value);
                                        }
                                    }
                                }
                            }
                        }
                    });
                }
            }, 100);
        </script>
    `;
}

/**
 * Generate Balance de Comprobación content
 */
function generateBalanceComprobacion() {
    const showMovements = document.getElementById('showMovements')?.checked;
    const onlyWithBalance = document.getElementById('onlyWithBalance')?.checked;
    
    return `
        <div class="balance-comprobacion">
            <!-- KPI Cards -->
            <div class="kpi-cards">
                <div class="kpi-card">
                    <div class="kpi-icon">✅</div>
                    <div class="kpi-value">CUADRADO</div>
                    <div class="kpi-label">Estado del Balance</div>
                    <div class="kpi-change positive">Débitos = Créditos</div>
                </div>
                <div class="kpi-card">
                    <div class="kpi-icon">🏦</div>
                    <div class="kpi-value">87</div>
                    <div class="kpi-label">Cuentas Activas</div>
                    <div class="kpi-change">Con movimientos</div>
                </div>
                <div class="kpi-card">
                    <div class="kpi-icon">💰</div>
                    <div class="kpi-value">$125.5M</div>
                    <div class="kpi-label">Total Activos</div>
                    <div class="kpi-change positive">+15.3%</div>
                </div>
                <div class="kpi-card">
                    <div class="kpi-icon">📊</div>
                    <div class="kpi-value">$450M</div>
                    <div class="kpi-label">Total Movimientos</div>
                    <div class="kpi-change positive">+22.5%</div>
                </div>
            </div>

            <!-- Balance Verification Chart -->
            <div class="chart-container">
                <div class="chart-header">
                    <h3 class="chart-title">📊 Verificación del Balance</h3>
                    <div class="chart-options">
                        <button class="chart-option active">Por Clase</button>
                        <button class="chart-option">Por Naturaleza</button>
                        <button class="chart-option">Tendencia</button>
                    </div>
                </div>
                <canvas id="balanceVerChart" width="400" height="200"></canvas>
            </div>

            <!-- Balance Table -->
            <div class="report-body" style="padding: 30px;">
                <h4 style="color: #2d3748; margin-bottom: 20px;">✅ BALANCE DE COMPROBACIÓN - ENERO 2024</h4>
                
                <table class="report-table">
                    <thead>
                        <tr>
                            <th rowspan="2">Código</th>
                            <th rowspan="2">Cuenta</th>
                            <th colspan="2" class="text-center" style="background: #f7fafc;">Saldos Anteriores</th>
                            ${showMovements ? '<th colspan="2" class="text-center" style="background: #edf2f7;">Movimientos</th>' : ''}
                            <th colspan="2" class="text-center" style="background: #e6fffa;">Saldos Actuales</th>
                        </tr>
                        <tr>
                            <th class="text-right" style="background: #f7fafc;">Débito</th>
                            <th class="text-right" style="background: #f7fafc;">Crédito</th>
                            ${showMovements ? '<th class="text-right" style="background: #edf2f7;">Débito</th>' : ''}
                            ${showMovements ? '<th class="text-right" style="background: #edf2f7;">Crédito</th>' : ''}
                            <th class="text-right" style="background: #e6fffa;">Débito</th>
                            <th class="text-right" style="background: #e6fffa;">Crédito</th>
                        </tr>
                    </thead>
                    <tbody>
                        <!-- Clase 1: Activos -->
                        <tr class="section-header">
                            <td colspan="${showMovements ? '8' : '6'}">
                                <span class="account-code">1</span> ACTIVOS
                            </td>
                        </tr>
                        <tr>
                            <td><span class="account-code">1105</span></td>
                            <td>CAJA</td>
                            <td class="text-right">$2,000,000</td>
                            <td class="text-right">-</td>
                            ${showMovements ? '<td class="text-right positive-value">$560,000</td>' : ''}
                            ${showMovements ? '<td class="text-right">-</td>' : ''}
                            <td class="text-right positive-value">$2,560,000</td>
                            <td class="text-right">-</td>
                        </tr>
                        <tr>
                            <td><span class="account-code">1110</span></td>
                            <td>BANCOS</td>
                            <td class="text-right">$45,890,000</td>
                            <td class="text-right">-</td>
                            ${showMovements ? '<td class="text-right positive-value">$1,000,000</td>' : ''}
                            ${showMovements ? '<td class="text-right negative-value">$10,200,000</td>' : ''}
                            <td class="text-right positive-value">$36,690,000</td>
                            <td class="text-right">-</td>
                        </tr>
                        <tr>
                            <td><span class="account-code">1305</span></td>
                            <td>CLIENTES</td>
                            <td class="text-right">$30,000,000</td>
                            <td class="text-right">-</td>
                            ${showMovements ? '<td class="text-right positive-value">$5,450,000</td>' : ''}
                            ${showMovements ? '<td class="text-right negative-value">$1,560,000</td>' : ''}
                            <td class="text-right positive-value">$33,890,000</td>
                            <td class="text-right">-</td>
                        </tr>
                        <tr>
                            <td><span class="account-code">1540</span></td>
                            <td>FLOTA Y EQUIPO DE TRANSPORTE</td>
                            <td class="text-right">$45,000,000</td>
                            <td class="text-right">-</td>
                            ${showMovements ? '<td class="text-right">-</td>' : ''}
                            ${showMovements ? '<td class="text-right">-</td>' : ''}
                            <td class="text-right positive-value">$45,000,000</td>
                            <td class="text-right">-</td>
                        </tr>
                        <tr>
                            <td><span class="account-code">1592</span></td>
                            <td>DEPRECIACIÓN ACUMULADA</td>
                            <td class="text-right">-</td>
                            <td class="text-right">$8,550,000</td>
                            ${showMovements ? '<td class="text-right">-</td>' : ''}
                            ${showMovements ? '<td class="text-right negative-value">$1,140,000</td>' : ''}
                            <td class="text-right">-</td>
                            <td class="text-right negative-value">$9,690,000</td>
                        </tr>
                        
                        <!-- Subtotal Activos -->
                        <tr style="background: #f7fafc; font-weight: 600;">
                            <td colspan="2">Subtotal Activos</td>
                            <td class="text-right">$122,890,000</td>
                            <td class="text-right">$8,550,000</td>
                            ${showMovements ? '<td class="text-right">$7,010,000</td>' : ''}
                            ${showMovements ? '<td class="text-right">$12,900,000</td>' : ''}
                            <td class="text-right">$117,140,000</td>
                            <td class="text-right">$9,690,000</td>
                        </tr>
                        
                        <!-- Clase 2: Pasivos -->
                        <tr class="section-header">
                            <td colspan="${showMovements ? '8' : '6'}">
                                <span class="account-code">2</span> PASIVOS
                            </td>
                        </tr>
                        <tr>
                            <td><span class="account-code">2105</span></td>
                            <td>OBLIGACIONES FINANCIERAS</td>
                            <td class="text-right">-</td>
                            <td class="text-right">$15,000,000</td>
                            ${showMovements ? '<td class="text-right">-</td>' : ''}
                            ${showMovements ? '<td class="text-right">-</td>' : ''}
                            <td class="text-right">-</td>
                            <td class="text-right negative-value">$15,000,000</td>
                        </tr>
                        <tr>
                            <td><span class="account-code">2408</span></td>
                            <td>IVA POR PAGAR</td>
                            <td class="text-right">-</td>
                            <td class="text-right">$8,920,000</td>
                            ${showMovements ? '<td class="text-right">-</td>' : ''}
                            ${showMovements ? '<td class="text-right negative-value">$2,391,176</td>' : ''}
                            <td class="text-right">-</td>
                            <td class="text-right negative-value">$11,311,176</td>
                        </tr>
                        
                        <!-- Clase 3: Patrimonio -->
                        <tr class="section-header">
                            <td colspan="${showMovements ? '8' : '6'}">
                                <span class="account-code">3</span> PATRIMONIO
                            </td>
                        </tr>
                        <tr>
                            <td><span class="account-code">3115</span></td>
                            <td>CAPITAL SOCIAL</td>
                            <td class="text-right">-</td>
                            <td class="text-right">$50,000,000</td>
                            ${showMovements ? '<td class="text-right">-</td>' : ''}
                            ${showMovements ? '<td class="text-right">-</td>' : ''}
                            <td class="text-right">-</td>
                            <td class="text-right negative-value">$50,000,000</td>
                        </tr>
                        
                        <!-- Clase 4: Ingresos -->
                        <tr class="section-header">
                            <td colspan="${showMovements ? '8' : '6'}">
                                <span class="account-code">4</span> INGRESOS
                            </td>
                        </tr>
                        <tr>
                            <td><span class="account-code">4135</span></td>
                            <td>INGRESOS POR SERVICIOS</td>
                            <td class="text-right">-</td>
                            <td class="text-right">$38,000,000</td>
                            ${showMovements ? '<td class="text-right">-</td>' : ''}
                            ${showMovements ? '<td class="text-right negative-value">$10,058,824</td>' : ''}
                            <td class="text-right">-</td>
                            <td class="text-right negative-value">$48,058,824</td>
                        </tr>
                        
                        <!-- Clase 5: Gastos -->
                        <tr class="section-header">
                            <td colspan="${showMovements ? '8' : '6'}">
                                <span class="account-code">5</span> GASTOS
                            </td>
                        </tr>
                        <tr>
                            <td><span class="account-code">5105</span></td>
                            <td>GASTOS DE PERSONAL</td>
                            <td class="text-right">$2,420,000</td>
                            <td class="text-right">-</td>
                            ${showMovements ? '<td class="text-right positive-value">$3,580,000</td>' : ''}
                            ${showMovements ? '<td class="text-right">-</td>' : ''}
                            <td class="text-right positive-value">$6,000,000</td>
                            <td class="text-right">-</td>
                        </tr>
                        
                        <!-- Totales -->
                        <tr class="total-row">
                            <td colspan="2">SUMAS IGUALES</td>
                            <td class="text-right"><strong>$125,310,000</strong></td>
                            <td class="text-right"><strong>$125,310,000</strong></td>
                            ${showMovements ? '<td class="text-right"><strong>$10,590,000</strong></td>' : ''}
                            ${showMovements ? '<td class="text-right"><strong>$25,350,000</strong></td>' : ''}
                            <td class="text-right"><strong>$123,140,000</strong></td>
                            <td class="text-right"><strong>$123,140,000</strong></td>
                        </tr>
                        
                        <tr style="background: linear-gradient(90deg, #68d391 0%, #38a169 100%); color: white;">
                            <td colspan="${showMovements ? '8' : '6'}" class="text-center">
                                <strong>✅ BALANCE CUADRADO - DÉBITOS = CRÉDITOS</strong>
                            </td>
                        </tr>
                    </tbody>
                </table>

                <!-- Balance Analysis -->
                <div class="comparison-container" style="margin-top: 40px;">
                    <div class="comparison-column">
                        <div class="comparison-header">Composición por Clase</div>
                        <div class="comparison-metric">
                            <span class="metric-label">Activos</span>
                            <div class="progress-bar-container">
                                <div class="progress-bar" style="width: 100%; background: #48bb78;"></div>
                            </div>
                            <span class="metric-value">$107,450,000</span>
                        </div>
                        <div class="comparison-metric">
                            <span class="metric-label">Pasivos</span>
                            <div class="progress-bar-container">
                                <div class="progress-bar" style="width: 25%; background: #ed8936;"></div>
                            </div>
                            <span class="metric-value">$26,311,176</span>
                        </div>
                        <div class="comparison-metric">
                            <span class="metric-label">Patrimonio</span>
                            <div class="progress-bar-container">
                                <div class="progress-bar" style="width: 47%; background: #4299e1;"></div>
                            </div>
                            <span class="metric-value">$50,000,000</span>
                        </div>
                        <div class="comparison-metric">
                            <span class="metric-label">Resultado del Período</span>
                            <div class="progress-bar-container">
                                <div class="progress-bar" style="width: 30%; background: #805ad5;"></div>
                            </div>
                            <span class="metric-value positive-value">$31,138,824</span>
                        </div>
                    </div>
                    <div class="comparison-column">
                        <div class="comparison-header">Verificaciones</div>
                        <div class="comparison-metric">
                            <span class="metric-label">Ecuación Contable</span>
                            <span class="metric-value positive-value">✅ Cumple</span>
                        </div>
                        <div class="comparison-metric">
                            <span class="metric-label">Partida Doble</span>
                            <span class="metric-value positive-value">✅ Balanceado</span>
                        </div>
                        <div class="comparison-metric">
                            <span class="metric-label">Diferencia</span>
                            <span class="metric-value">$0.00</span>
                        </div>
                        <div class="comparison-metric">
                            <span class="metric-label">Cuentas sin Movimiento</span>
                            <span class="metric-value">23</span>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Report Actions -->
            <div class="report-actions">
                <button class="report-action" onclick="drillDownAccount()">
                    🔍 Explorar Cuenta
                </button>
                <button class="report-action" onclick="exportTrialBalance()">
                    📥 Exportar Balance
                </button>
                <button class="report-action" onclick="showAccountingEquation()">
                    ⚖️ Ver Ecuación Contable
                </button>
                <button class="report-action primary" onclick="auditBalance()">
                    🔎 Auditar Balance
                </button>
            </div>
        </div>

        <script>
            // Generar gráfico de verificación
            setTimeout(() => {
                const ctx = document.getElementById('balanceVerChart');
                if (ctx && typeof Chart !== 'undefined') {
                    new Chart(ctx, {
                        type: 'doughnut',
                        data: {
                            labels: ['Activos', 'Pasivos', 'Patrimonio', 'Ingresos', 'Gastos'],
                            datasets: [{
                                data: [107450000, 26311176, 50000000, 48058824, 6000000],
                                backgroundColor: [
                                    'rgba(72, 187, 120, 0.8)',
                                    'rgba(237, 137, 54, 0.8)',
                                    'rgba(66, 153, 225, 0.8)',
                                    'rgba(128, 90, 213, 0.8)',
                                    'rgba(245, 101, 101, 0.8)'
                                ],
                                borderColor: [
                                    'rgba(72, 187, 120, 1)',
                                    'rgba(237, 137, 54, 1)',
                                    'rgba(66, 153, 225, 1)',
                                    'rgba(128, 90, 213, 1)',
                                    'rgba(245, 101, 101, 1)'
                                ],
                                borderWidth: 2
                            }]
                        },
                        options: {
                            responsive: true,
                            maintainAspectRatio: false,
                            plugins: {
                                legend: {
                                    position: 'right'
                                },
                                tooltip: {
                                    callbacks: {
                                        label: function(context) {
                                            const value = new Intl.NumberFormat('es-CO', {
                                                style: 'currency',
                                                currency: 'COP',
                                                minimumFractionDigits: 0
                                            }).format(context.parsed);
                                            const total = context.dataset.data.reduce((a, b) => a + b, 0);
                                            const percentage = ((context.parsed / total) * 100).toFixed(1);
                                            return context.label + ': ' + value + ' (' + percentage + '%)';
                                        }
                                    }
                                }
                            }
                        }
                    });
                }
            }, 100);
        </script>
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