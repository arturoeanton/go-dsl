// Dashboard JavaScript - Motor Contable
// Version: 1.0
// Last Updated: 2024-01-15

// Global configuration
const CONFIG = {
    API_BASE_URL: '/api/v1',
    REFRESH_INTERVAL: 30000, // 30 seconds
    CHART_UPDATE_INTERVAL: 60000, // 1 minute
    COLORS: {
        primary: '#1890ff',
        success: '#52c41a',
        warning: '#faad14',
        error: '#f5222d',
        dark: '#001529'
    }
};

// State management
const state = {
    currentOrg: null,
    refreshTimer: null,
    chartInstance: null,
    lastUpdate: new Date()
};

// Initialize dashboard on DOM ready
document.addEventListener('DOMContentLoaded', () => {
    initializeDashboard();
});

/**
 * Initialize all dashboard components
 */
function initializeDashboard() {
    // Setup event listeners
    setupEventListeners();
    
    // Load initial data
    loadDashboardData();
    
    // Initialize charts
    initializeCharts();
    createAdditionalCharts();
    
    // Setup auto-refresh
    setupAutoRefresh();
    
    // Initialize organization selector
    initOrgSelector();
}

/**
 * Setup all event listeners
 */
function setupEventListeners() {
    // Menu toggle - handled by utils.js
    // const menuToggle = document.getElementById('menuToggle');
    // if (menuToggle) {
    //     menuToggle.addEventListener('click', toggleSidebar);
    // }
    
    // Organization selector
    const orgSelector = document.querySelector('.org-selector');
    if (orgSelector) {
        orgSelector.addEventListener('change', handleOrgChange);
    }
    
    // Refresh button
    const refreshBtn = document.querySelector('.btn-refresh');
    if (refreshBtn) {
        refreshBtn.addEventListener('click', refreshActivityFeed);
    }
    
    // Period selector for charts
    const periodSelector = document.querySelector('.period-selector');
    if (periodSelector) {
        periodSelector.addEventListener('change', updateChartPeriod);
    }
}

/**
 * Toggle sidebar visibility - MOVED TO utils.js for consistency
 */
// function toggleSidebar() {
//     const sidebar = document.getElementById('sidebar');
//     sidebar.classList.toggle('collapsed');
//     
//     // Save preference
//     localStorage.setItem('sidebarCollapsed', sidebar.classList.contains('collapsed'));
// }

/**
 * Load dashboard data from API
 */
async function loadDashboardData() {
    try {
        // Show loading state
        showLoadingState();
        
        // Fetch dashboard data using API service
        const result = await motorContableApi.dashboard.getStats();
        const data = result.data;
        
        // Update UI with data
        updateKPICards(data.kpis);
        updateActivityFeed(data.recent_activity);
        updateCharts(data.charts);
        updateSystemHealth(data.system_health);
        
        // Hide loading state
        hideLoadingState();
        
    } catch (error) {
        console.error('Error loading dashboard:', error);
        showError('Error al cargar el dashboard. Por favor, intente nuevamente.');
    }
}

/**
 * Update KPI cards with data
 */
function updateKPICards(kpis) {
    // Update each KPI card
    const kpiData = [
        { selector: '.stat-card:nth-child(1) .stat-value', value: kpis.vouchers_today },
        { selector: '.stat-card:nth-child(2) .stat-value', value: kpis.vouchers_month },
        { selector: '.stat-card:nth-child(3) .stat-value', value: kpis.pending_vouchers },
        { selector: '.stat-card:nth-child(4) .stat-value', value: Math.round(kpis.processing_rate) + '%' }
    ];
    
    kpiData.forEach(kpi => {
        const element = document.querySelector(kpi.selector);
        if (element) {
            if (typeof kpi.value === 'string') {
                element.textContent = kpi.value;
            } else {
                animateValue(element, kpi.value);
            }
        }
    });
}

/**
 * Animate numeric value change
 */
function animateValue(element, endValue) {
    const startValue = parseInt(element.textContent.replace(/[^0-9]/g, '')) || 0;
    const duration = 1000; // 1 second
    const startTime = performance.now();
    
    function update(currentTime) {
        const elapsed = currentTime - startTime;
        const progress = Math.min(elapsed / duration, 1);
        
        const currentValue = Math.floor(startValue + (endValue - startValue) * progress);
        element.textContent = formatNumber(currentValue);
        
        if (progress < 1) {
            requestAnimationFrame(update);
        }
    }
    
    requestAnimationFrame(update);
}

/**
 * Format number with thousands separator
 */
function formatNumber(num) {
    return num.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',');
}

/**
 * Initialize charts
 */
function initializeCharts() {
    const ctx = document.getElementById('processingChart');
    if (!ctx) return;
    
    // Initialize with empty chart
    state.chartInstance = new Chart(ctx, {
        type: 'line',
        data: {
            labels: [],
            datasets: [{
                label: 'Comprobantes Procesados',
                data: [],
                borderColor: CONFIG.COLORS.primary,
                backgroundColor: CONFIG.COLORS.primary + '20',
                borderWidth: 2,
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
                    mode: 'index',
                    intersect: false,
                    callbacks: {
                        label: function(context) {
                            return 'Procesados: ' + context.parsed.y.toLocaleString('es-CO');
                        }
                    }
                }
            },
            scales: {
                x: {
                    grid: {
                        display: false
                    }
                },
                y: {
                    beginAtZero: true,
                    ticks: {
                        callback: function(value) {
                            return value.toLocaleString('es-CO');
                        }
                    }
                }
            }
        }
    });
}

/**
 * Update vouchers by day chart
 */
function updateVouchersByDayChart(data) {
    if (!state.chartInstance) return;
    
    // Update chart data
    state.chartInstance.data.labels = data.labels;
    state.chartInstance.data.datasets[0].data = data.values;
    state.chartInstance.update();
}

/**
 * Update activity feed
 */
function updateActivityFeed(activities) {
    const activityList = document.querySelector('.activity-list');
    if (!activityList) return;
    
    // Clear existing activities
    activityList.innerHTML = '';
    
    // Add new activities
    activities.forEach(activity => {
        const item = createActivityItem(activity);
        activityList.appendChild(item);
    });
}

/**
 * Update charts with data
 */
function updateCharts(chartsData) {
    // Update main processing chart
    if (chartsData.vouchers_by_day) {
        updateVouchersByDayChart(chartsData.vouchers_by_day);
    }
    
    // Update vouchers by type chart
    if (chartsData.vouchers_by_type && state.typeChartInstance) {
        state.typeChartInstance.data.labels = chartsData.vouchers_by_type.labels;
        state.typeChartInstance.data.datasets[0].data = chartsData.vouchers_by_type.values;
        state.typeChartInstance.update();
    }
    
    // Update vouchers by status chart
    if (chartsData.vouchers_by_status && state.statusChartInstance) {
        state.statusChartInstance.data.datasets[0].data = chartsData.vouchers_by_status.values;
        state.statusChartInstance.update();
    }
}

/**
 * Update system health indicators
 */
function updateSystemHealth(health) {
    const healthIndicator = document.querySelector('.system-health');
    if (!healthIndicator) return;
    
    // Update status
    const statusElement = healthIndicator.querySelector('.health-status');
    if (statusElement) {
        statusElement.textContent = health.status === 'healthy' ? '✅ Sistema Operativo' : '⚠️ Sistema con Problemas';
        statusElement.className = `health-status ${health.status}`;
    }
    
    // Update metrics
    const metrics = [
        { label: 'Uptime', value: `${health.uptime}%` },
        { label: 'Response Time', value: `${health.api_response_time}ms` },
        { label: 'Cache Hit Rate', value: `${health.cache_hit_rate}%` }
    ];
    
    const metricsContainer = healthIndicator.querySelector('.health-metrics');
    if (metricsContainer) {
        metricsContainer.innerHTML = metrics.map(m => 
            `<div class="health-metric"><span>${m.label}:</span> <strong>${m.value}</strong></div>`
        ).join('');
    }
}

/**
 * Create additional charts
 */
function createAdditionalCharts() {
    // Vouchers by Type Chart (Doughnut)
    const typeCtx = document.getElementById('voucherTypeChart');
    if (typeCtx) {
        state.typeChartInstance = new Chart(typeCtx, {
            type: 'doughnut',
            data: {
                labels: [],
                datasets: [{
                    data: [],
                    backgroundColor: [
                        '#52c41a',
                        '#1890ff',
                        '#faad14',
                        '#f5222d',
                        '#722ed1',
                        '#13c2c2'
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
                        position: 'bottom',
                        labels: {
                            padding: 15,
                            font: {
                                size: 12
                            }
                        }
                    },
                    tooltip: {
                        callbacks: {
                            label: function(context) {
                                const label = context.label || '';
                                const value = context.parsed || 0;
                                const total = context.dataset.data.reduce((a, b) => a + b, 0);
                                const percentage = ((value / total) * 100).toFixed(1);
                                return label + ': ' + value.toLocaleString('es-CO') + ' (' + percentage + '%)';
                            }
                        }
                    }
                }
            }
        });
    }

    // Voucher Status Chart (Bar)
    const statusCtx = document.getElementById('voucherStatusChart');
    if (statusCtx) {
        state.statusChartInstance = new Chart(statusCtx, {
            type: 'bar',
            data: {
                labels: ['Procesados', 'Pendientes', 'Con Error', 'En Proceso'],
                datasets: [{
                    label: 'Comprobantes',
                    data: [],
                    backgroundColor: [
                        '#52c41a',
                        '#faad14',
                        '#f5222d',
                        '#1890ff'
                    ],
                    borderWidth: 0
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
                                return context.parsed.y.toLocaleString('es-CO') + ' comprobantes';
                            }
                        }
                    }
                },
                scales: {
                    x: {
                        grid: {
                            display: false
                        }
                    },
                    y: {
                        beginAtZero: true,
                        ticks: {
                            callback: function(value) {
                                return value.toLocaleString('es-CO');
                            }
                        }
                    }
                }
            }
        });
    }
}

/**
 * Create activity item element
 */
function createActivityItem(activity) {
    const div = document.createElement('div');
    div.className = 'activity-item';
    
    const statusClass = activity.status.toLowerCase();
    const statusIcon = getStatusIcon(activity.status);
    
    div.innerHTML = `
        <span class="activity-icon ${statusClass}">${statusIcon}</span>
        <div class="activity-content">
            <p><strong>${activity.id}</strong> - ${activity.description}</p>
            <p class="activity-meta">
                <span class="activity-type">${getVoucherTypeLabel(activity.type)}</span>
                <span class="activity-amount">$${formatCurrency(activity.amount)}</span>
            </p>
            ${activity.error ? `<p class="activity-error">❌ ${activity.error}</p>` : ''}
            <small>${formatTimeAgo(activity.created_at)}</small>
        </div>
    `;
    
    return div;
}

/**
 * Get activity icon based on type
 */
function getActivityIcon(type) {
    const icons = {
        success: '✅',
        info: '📃',
        warning: '⚠️',
        error: '❌'
    };
    return icons[type] || '📄';
}

/**
 * Get status icon
 */
function getStatusIcon(status) {
    const icons = {
        'PROCESSED': '✅',
        'PENDING': '⏳',
        'PROCESSING': '🔄',
        'ERROR': '❌',
        'CANCELLED': '🚫'
    };
    return icons[status] || '📄';
}

/**
 * Get voucher type label
 */
function getVoucherTypeLabel(type) {
    const labels = {
        'invoice_sale': 'Factura de Venta',
        'invoice_purchase': 'Factura de Compra',
        'payment': 'Pago',
        'receipt': 'Recibo',
        'credit_note': 'Nota Crédito',
        'debit_note': 'Nota Débito'
    };
    return labels[type] || type;
}

/**
 * Format currency amount
 */
function formatCurrency(amount) {
    return new Intl.NumberFormat('es-CO', {
        minimumFractionDigits: 0,
        maximumFractionDigits: 2
    }).format(amount);
}

/**
 * Format time ago
 */
function formatTimeAgo(timestamp) {
    const date = new Date(timestamp);
    const now = new Date();
    const diff = now - date;
    
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(diff / 3600000);
    const days = Math.floor(diff / 86400000);
    
    if (minutes < 1) return 'hace un momento';
    if (minutes < 60) return `hace ${minutes} minutos`;
    if (hours < 24) return `hace ${hours} horas`;
    return `hace ${days} días`;
}

/**
 * Setup auto-refresh
 */
function setupAutoRefresh() {
    // Clear existing timer
    if (state.refreshTimer) {
        clearInterval(state.refreshTimer);
    }
    
    // Setup new timer
    state.refreshTimer = setInterval(() => {
        refreshActivityFeed();
    }, CONFIG.REFRESH_INTERVAL);
}

/**
 * Refresh activity feed
 */
async function refreshActivityFeed() {
    const refreshBtn = document.querySelector('.btn-refresh');
    if (refreshBtn) {
        refreshBtn.classList.add('spinning');
    }
    
    try {
        const data = await motorContableApi.dashboard.getActivity();
        updateActivityFeed(data.activities);
    } catch (error) {
        console.error('Error refreshing activities:', error);
    } finally {
        if (refreshBtn) {
            refreshBtn.classList.remove('spinning');
        }
    }
}

/**
 * Handle organization change
 */
function handleOrgChange(event) {
    const orgId = event.target.value;
    state.currentOrg = orgId;
    
    // Reload dashboard data for new organization
    loadDashboardData();
}

/**
 * Initialize organization selector
 */
function initOrgSelector() {
    const savedOrg = localStorage.getItem('selectedOrg');
    if (savedOrg) {
        const selector = document.querySelector('.org-selector');
        if (selector) {
            selector.value = savedOrg;
            state.currentOrg = savedOrg;
        }
    }
}

/**
 * Update chart period
 */
function updateChartPeriod(event) {
    const period = event.target.value;
    
    // In real implementation, fetch new data based on period
    console.log('Updating chart for period:', period);
    
    // Redraw chart with new data
    initializeCharts();
}

/**
 * Show loading state
 */
function showLoadingState() {
    // Add loading class to KPI cards
    document.querySelectorAll('.stat-card').forEach(card => {
        card.classList.add('loading');
    });
}

/**
 * Hide loading state
 */
function hideLoadingState() {
    // Remove loading class from KPI cards
    document.querySelectorAll('.stat-card').forEach(card => {
        card.classList.remove('loading');
    });
}

/**
 * Show error message
 */
function showError(message) {
    // In real implementation, show toast or modal
    console.error(message);
    alert(message);
}

/**
 * Get auth token
 */
function getAuthToken() {
    // In real implementation, get from secure storage
    return localStorage.getItem('authToken') || 'mock-token';
}


// Export functions for testing
if (typeof module !== 'undefined' && module.exports) {
    module.exports = {
        formatNumber,
        formatTimeAgo,
        getActivityIcon
    };
}