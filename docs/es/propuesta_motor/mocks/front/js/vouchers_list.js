// Vouchers List JavaScript - Motor Contable
// Version: 1.0
// Last Updated: 2024-01-15

// Global state
const state = {
    vouchers: [],
    filteredVouchers: [],
    currentPage: 1,
    pageSize: 50,
    totalPages: 1,
    selectedVouchers: new Set(),
    sortColumn: 'voucher_date',
    sortDirection: 'desc',
    filters: {
        dateFrom: null,
        dateTo: null,
        type: '',
        status: '',
        thirdParty: '',
        amountFrom: null,
        amountTo: null
    }
};

// Initialize on DOM ready
document.addEventListener('DOMContentLoaded', () => {
    initializeVouchersList();
});

/**
 * Initialize vouchers list
 */
function initializeVouchersList() {
    setupEventListeners();
    loadVouchersData();
    setupFilters();
}

/**
 * Setup event listeners
 */
function setupEventListeners() {
    // Select all checkbox
    document.getElementById('selectAll').addEventListener('change', handleSelectAll);
    
    // Bulk action buttons
    document.getElementById('bulkProcess').addEventListener('click', handleBulkProcess);
    document.getElementById('bulkExport').addEventListener('click', handleBulkExport);
    document.getElementById('bulkDelete').addEventListener('click', handleBulkDelete);
    
    // Page size selector
    document.getElementById('pageSize').addEventListener('change', handlePageSizeChange);
    
    // Pagination buttons
    document.getElementById('prevPage').addEventListener('click', () => changePage(-1));
    document.getElementById('nextPage').addEventListener('click', () => changePage(1));
    
    // Filter toggle
    document.getElementById('toggleFilters').addEventListener('click', toggleFilters);
    
    // Search input
    document.getElementById('searchInput').addEventListener('keyup', debounce(handleSearch, 300));
    
    // Sort headers
    setupSortHeaders();
}

/**
 * Load vouchers data
 */
async function loadVouchersData() {
    showLoading();
    
    try {
        // Fetch data using API service
        const result = await motorContableApi.vouchers.getList({
            page: state.currentPage,
            per_page: state.pageSize,
            sort: state.sortBy,
            order: state.sortOrder
        });
        
        state.vouchers = result.data.vouchers;
        state.filteredVouchers = [...state.vouchers];
        
        updateStats(result.data.stats);
        applyFiltersAndSort();
        renderVouchersTable();
        updatePagination();
        
    } catch (error) {
        console.error('Error loading vouchers:', error);
        showError('Error al cargar los comprobantes');
    } finally {
        hideLoading();
    }
}

/**
 * Render vouchers table
 */
function renderVouchersTable() {
    const tbody = document.getElementById('vouchersTableBody');
    tbody.innerHTML = '';
    
    // Calculate pagination
    const startIdx = (state.currentPage - 1) * state.pageSize;
    const endIdx = startIdx + state.pageSize;
    const pageVouchers = state.filteredVouchers.slice(startIdx, endIdx);
    
    // Render rows
    pageVouchers.forEach(voucher => {
        const row = createVoucherRow(voucher);
        tbody.appendChild(row);
    });
    
    // Update showing info
    document.getElementById('showingFrom').textContent = startIdx + 1;
    document.getElementById('showingTo').textContent = Math.min(endIdx, state.filteredVouchers.length);
    document.getElementById('totalRecords').textContent = state.filteredVouchers.length;
}

/**
 * Create voucher table row
 */
function createVoucherRow(voucher) {
    const tr = document.createElement('tr');
    tr.className = `voucher-row status-${voucher.status.toLowerCase()}`;
    
    tr.innerHTML = `
        <td>
            <input type="checkbox" class="voucher-checkbox" value="${voucher.id}"
                ${state.selectedVouchers.has(voucher.id) ? 'checked' : ''}>
        </td>
        <td class="voucher-number">
            <a href="#" onclick="showQuickView('${voucher.id}')">${voucher.voucher_number}</a>
        </td>
        <td>${formatDate(voucher.voucher_date)}</td>
        <td><span class="badge badge-${getTypeClass(voucher.type)}">${getVoucherTypeLabel(voucher.type)}</span></td>
        <td>${voucher.third_party.name}<br><small class="text-muted">${voucher.third_party.id_number}</small></td>
        <td class="description-cell">${voucher.description}</td>
        <td class="text-right">$${formatCurrency(voucher.total_amount)}</td>
        <td><span class="status-badge status-${voucher.status.toLowerCase()}">${getStatusLabel(voucher.status)}</span></td>
        <td class="actions-cell">
            <button class="btn btn-sm btn-icon" title="Ver" onclick="showQuickView('${voucher.id}')">👁️</button>
            <button class="btn btn-sm btn-icon" title="Editar" onclick="editVoucher('${voucher.id}')">✏️</button>
            <button class="btn btn-sm btn-icon" title="Procesar" onclick="processVoucher('${voucher.id}')"
                ${voucher.status !== 'PENDING' ? 'disabled' : ''}>⚡</button>
            <button class="btn btn-sm btn-icon btn-danger" title="Eliminar" onclick="deleteVoucher('${voucher.id}')">🗑️</button>
        </td>
    `;
    
    // Add checkbox event listener
    const checkbox = tr.querySelector('.voucher-checkbox');
    checkbox.addEventListener('change', () => handleVoucherSelection(voucher.id, checkbox.checked));
    
    return tr;
}

/**
 * Handle voucher selection
 */
function handleVoucherSelection(voucherId, isChecked) {
    if (isChecked) {
        state.selectedVouchers.add(voucherId);
    } else {
        state.selectedVouchers.delete(voucherId);
    }
    
    updateBulkActionButtons();
    updateSelectAllCheckbox();
}

/**
 * Handle select all
 */
function handleSelectAll(event) {
    const isChecked = event.target.checked;
    const checkboxes = document.querySelectorAll('.voucher-checkbox');
    
    checkboxes.forEach(cb => {
        cb.checked = isChecked;
        const voucherId = cb.value;
        if (isChecked) {
            state.selectedVouchers.add(voucherId);
        } else {
            state.selectedVouchers.delete(voucherId);
        }
    });
    
    updateBulkActionButtons();
}

/**
 * Update bulk action buttons
 */
function updateBulkActionButtons() {
    const hasSelection = state.selectedVouchers.size > 0;
    document.getElementById('bulkProcess').disabled = !hasSelection;
    document.getElementById('bulkExport').disabled = !hasSelection;
    document.getElementById('bulkDelete').disabled = !hasSelection;
}

/**
 * Update select all checkbox
 */
function updateSelectAllCheckbox() {
    const checkboxes = document.querySelectorAll('.voucher-checkbox');
    const checkedCount = document.querySelectorAll('.voucher-checkbox:checked').length;
    const selectAll = document.getElementById('selectAll');
    
    selectAll.checked = checkboxes.length > 0 && checkedCount === checkboxes.length;
    selectAll.indeterminate = checkedCount > 0 && checkedCount < checkboxes.length;
}

/**
 * Apply filters and sort
 */
function applyFiltersAndSort() {
    // Apply filters
    state.filteredVouchers = state.vouchers.filter(voucher => {
        // Date filter
        if (state.filters.dateFrom && new Date(voucher.voucher_date) < new Date(state.filters.dateFrom)) {
            return false;
        }
        if (state.filters.dateTo && new Date(voucher.voucher_date) > new Date(state.filters.dateTo)) {
            return false;
        }
        
        // Type filter
        if (state.filters.type && voucher.type !== state.filters.type) {
            return false;
        }
        
        // Status filter
        if (state.filters.status && voucher.status !== state.filters.status) {
            return false;
        }
        
        // Third party filter
        if (state.filters.thirdParty) {
            const searchTerm = state.filters.thirdParty.toLowerCase();
            if (!voucher.third_party.name.toLowerCase().includes(searchTerm) &&
                !voucher.third_party.id_number.includes(searchTerm)) {
                return false;
            }
        }
        
        // Amount filter
        if (state.filters.amountFrom && voucher.total_amount < state.filters.amountFrom) {
            return false;
        }
        if (state.filters.amountTo && voucher.total_amount > state.filters.amountTo) {
            return false;
        }
        
        return true;
    });
    
    // Apply sort
    state.filteredVouchers.sort((a, b) => {
        let aVal = a[state.sortColumn];
        let bVal = b[state.sortColumn];
        
        // Handle nested properties
        if (state.sortColumn === 'third_party') {
            aVal = a.third_party.name;
            bVal = b.third_party.name;
        }
        
        // Compare values
        if (aVal < bVal) return state.sortDirection === 'asc' ? -1 : 1;
        if (aVal > bVal) return state.sortDirection === 'asc' ? 1 : -1;
        return 0;
    });
    
    // Reset to first page
    state.currentPage = 1;
}

/**
 * Apply filters (called from UI)
 */
function applyFilters() {
    // Get filter values
    state.filters.dateFrom = document.getElementById('dateFrom').value;
    state.filters.dateTo = document.getElementById('dateTo').value;
    state.filters.type = document.getElementById('voucherType').value;
    state.filters.status = document.getElementById('voucherStatus').value;
    state.filters.thirdParty = document.getElementById('thirdParty').value;
    state.filters.amountFrom = parseFloat(document.getElementById('amountFrom').value) || null;
    state.filters.amountTo = parseFloat(document.getElementById('amountTo').value) || null;
    
    applyFiltersAndSort();
    renderVouchersTable();
    updatePagination();
}

/**
 * Show quick view modal
 */
function showQuickView(voucherId) {
    const voucher = state.vouchers.find(v => v.id === voucherId);
    if (!voucher) return;
    
    const modal = document.getElementById('quickViewModal');
    const content = document.getElementById('quickViewContent');
    
    content.innerHTML = `
        <div class="quick-view-content">
            <div class="quick-view-header">
                <h4>${voucher.voucher_number}</h4>
                <span class="status-badge status-${voucher.status.toLowerCase()}">${getStatusLabel(voucher.status)}</span>
            </div>
            
            <div class="quick-view-grid">
                <div class="info-group">
                    <label>Tipo:</label>
                    <span>${getVoucherTypeLabel(voucher.type)}</span>
                </div>
                <div class="info-group">
                    <label>Fecha:</label>
                    <span>${formatDate(voucher.voucher_date)}</span>
                </div>
                <div class="info-group">
                    <label>Tercero:</label>
                    <span>${voucher.third_party.name}</span>
                </div>
                <div class="info-group">
                    <label>NIT/CC:</label>
                    <span>${voucher.third_party.id_number}</span>
                </div>
                <div class="info-group">
                    <label>Monto Total:</label>
                    <span class="amount">$${formatCurrency(voucher.total_amount)}</span>
                </div>
                <div class="info-group">
                    <label>Moneda:</label>
                    <span>${voucher.currency_code}</span>
                </div>
            </div>
            
            <div class="description-section">
                <label>Descripción:</label>
                <p>${voucher.description}</p>
            </div>
            
            ${voucher.items ? renderQuickViewItems(voucher.items) : ''}
            
            ${voucher.journal_entry ? renderQuickViewJournalEntry(voucher.journal_entry) : ''}
            
            <div class="timestamps">
                <small>Creado: ${formatDateTime(voucher.created_at)} | 
                Actualizado: ${formatDateTime(voucher.updated_at)}</small>
            </div>
        </div>
    `;
    
    // Setup action buttons
    document.getElementById('editFromQuickView').onclick = () => {
        closeQuickView();
        editVoucher(voucherId);
    };
    
    document.getElementById('processFromQuickView').onclick = () => {
        closeQuickView();
        processVoucher(voucherId);
    };
    
    document.getElementById('processFromQuickView').disabled = voucher.status !== 'PENDING';
    
    modal.style.display = 'block';
}

/**
 * Close quick view modal
 */
function closeQuickView() {
    document.getElementById('quickViewModal').style.display = 'none';
}

/**
 * Process voucher
 */
async function processVoucher(voucherId) {
    if (!confirm('¿Desea procesar este comprobante?')) return;
    
    showLoading();
    
    try {
        // Simulate API call
        await new Promise(resolve => setTimeout(resolve, 1000));
        
        // Update voucher status
        const voucher = state.vouchers.find(v => v.id === voucherId);
        if (voucher) {
            voucher.status = 'PROCESSING';
            renderVouchersTable();
            showSuccess('Comprobante enviado a procesar');
        }
        
    } catch (error) {
        showError('Error al procesar el comprobante');
    } finally {
        hideLoading();
    }
}

/**
 * Edit voucher
 */
function editVoucher(voucherId) {
    window.location.href = `vouchers_form.html?id=${voucherId}`;
}

/**
 * Delete voucher
 */
async function deleteVoucher(voucherId) {
    if (!confirm('¿Está seguro de eliminar este comprobante?')) return;
    
    showLoading();
    
    try {
        // Simulate API call
        await new Promise(resolve => setTimeout(resolve, 500));
        
        // Remove from state
        const index = state.vouchers.findIndex(v => v.id === voucherId);
        if (index !== -1) {
            state.vouchers.splice(index, 1);
            applyFiltersAndSort();
            renderVouchersTable();
            updatePagination();
            showSuccess('Comprobante eliminado correctamente');
        }
        
    } catch (error) {
        showError('Error al eliminar el comprobante');
    } finally {
        hideLoading();
    }
}

/**
 * Handle bulk process
 */
async function handleBulkProcess() {
    const count = state.selectedVouchers.size;
    if (!confirm(`¿Desea procesar ${count} comprobantes?`)) return;
    
    showLoading();
    
    try {
        // Simulate API call
        await new Promise(resolve => setTimeout(resolve, 1500));
        
        // Update status for selected vouchers
        state.vouchers.forEach(voucher => {
            if (state.selectedVouchers.has(voucher.id) && voucher.status === 'PENDING') {
                voucher.status = 'PROCESSING';
            }
        });
        
        state.selectedVouchers.clear();
        renderVouchersTable();
        updateBulkActionButtons();
        showSuccess(`${count} comprobantes enviados a procesar`);
        
    } catch (error) {
        showError('Error al procesar los comprobantes');
    } finally {
        hideLoading();
    }
}

/**
 * Handle bulk export
 */
function handleBulkExport() {
    const selectedData = state.vouchers.filter(v => state.selectedVouchers.has(v.id));
    
    // Create CSV content
    const headers = ['Número', 'Fecha', 'Tipo', 'Tercero', 'NIT', 'Descripción', 'Monto', 'Estado'];
    const rows = selectedData.map(v => [
        v.voucher_number,
        v.voucher_date,
        v.type,
        v.third_party.name,
        v.third_party.id_number,
        v.description,
        v.total_amount,
        v.status
    ]);
    
    const csv = [headers, ...rows].map(row => row.join(',')).join('\n');
    
    // Download CSV
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `comprobantes_${new Date().toISOString().split('T')[0]}.csv`;
    a.click();
    
    showSuccess('Comprobantes exportados correctamente');
}

/**
 * Handle bulk delete
 */
async function handleBulkDelete() {
    const count = state.selectedVouchers.size;
    if (!confirm(`¿Está seguro de eliminar ${count} comprobantes?`)) return;
    
    showLoading();
    
    try {
        // Simulate API call
        await new Promise(resolve => setTimeout(resolve, 1000));
        
        // Remove selected vouchers
        state.vouchers = state.vouchers.filter(v => !state.selectedVouchers.has(v.id));
        state.selectedVouchers.clear();
        
        applyFiltersAndSort();
        renderVouchersTable();
        updatePagination();
        updateBulkActionButtons();
        showSuccess(`${count} comprobantes eliminados`);
        
    } catch (error) {
        showError('Error al eliminar los comprobantes');
    } finally {
        hideLoading();
    }
}

/**
 * Update pagination
 */
function updatePagination() {
    state.totalPages = Math.ceil(state.filteredVouchers.length / state.pageSize);
    
    // Update buttons
    document.getElementById('prevPage').disabled = state.currentPage === 1;
    document.getElementById('nextPage').disabled = state.currentPage === state.totalPages;
    
    // Update page numbers
    const pageNumbers = document.getElementById('pageNumbers');
    pageNumbers.innerHTML = '';
    
    // Show max 5 page numbers
    let startPage = Math.max(1, state.currentPage - 2);
    let endPage = Math.min(state.totalPages, startPage + 4);
    
    if (endPage - startPage < 4) {
        startPage = Math.max(1, endPage - 4);
    }
    
    for (let i = startPage; i <= endPage; i++) {
        const btn = document.createElement('button');
        btn.className = `btn btn-sm ${i === state.currentPage ? 'btn-primary' : ''}`;
        btn.textContent = i;
        btn.onclick = () => goToPage(i);
        pageNumbers.appendChild(btn);
    }
}

/**
 * Change page
 */
function changePage(direction) {
    const newPage = state.currentPage + direction;
    if (newPage >= 1 && newPage <= state.totalPages) {
        goToPage(newPage);
    }
}

/**
 * Go to page
 */
function goToPage(page) {
    state.currentPage = page;
    renderVouchersTable();
    updatePagination();
}

/**
 * Handle page size change
 */
function handlePageSizeChange(event) {
    state.pageSize = parseInt(event.target.value);
    state.currentPage = 1;
    renderVouchersTable();
    updatePagination();
}

/**
 * Toggle filters
 */
function toggleFilters() {
    const filtersBody = document.getElementById('filtersBody');
    filtersBody.classList.toggle('collapsed');
}

/**
 * Handle search
 */
function handleSearch(event) {
    const searchTerm = event.target.value.toLowerCase();
    
    if (searchTerm) {
        state.filteredVouchers = state.vouchers.filter(voucher => {
            return voucher.voucher_number.toLowerCase().includes(searchTerm) ||
                   voucher.description.toLowerCase().includes(searchTerm) ||
                   voucher.third_party.name.toLowerCase().includes(searchTerm) ||
                   voucher.third_party.id_number.includes(searchTerm);
        });
    } else {
        applyFiltersAndSort();
    }
    
    renderVouchersTable();
    updatePagination();
}

/**
 * Setup sort headers
 */
function setupSortHeaders() {
    const sortableHeaders = {
        1: 'voucher_number',
        2: 'voucher_date',
        3: 'type',
        4: 'third_party',
        6: 'total_amount',
        7: 'status'
    };
    
    Object.entries(sortableHeaders).forEach(([index, column]) => {
        const th = document.querySelector(`#vouchersTable thead th:nth-child(${parseInt(index) + 1})`);
        th.style.cursor = 'pointer';
        th.addEventListener('click', () => handleSort(column));
    });
}

/**
 * Handle sort
 */
function handleSort(column) {
    if (state.sortColumn === column) {
        state.sortDirection = state.sortDirection === 'asc' ? 'desc' : 'asc';
    } else {
        state.sortColumn = column;
        state.sortDirection = 'asc';
    }
    
    applyFiltersAndSort();
    renderVouchersTable();
}

/**
 * Update stats
 */
function updateStats(stats) {
    document.getElementById('totalVouchers').textContent = stats.total_vouchers;
    document.getElementById('totalAmount').textContent = `$${formatCurrency(stats.total_amount)}`;
    document.getElementById('pendingCount').textContent = stats.pending_count;
    document.getElementById('errorCount').textContent = stats.error_count;
}

/**
 * Setup filters
 */
function setupFilters() {
    // Set default dates
    const today = new Date();
    const firstDay = new Date(today.getFullYear(), today.getMonth(), 1);
    
    document.getElementById('dateFrom').value = firstDay.toISOString().split('T')[0];
    document.getElementById('dateTo').value = today.toISOString().split('T')[0];
}

// Helper functions
function getVoucherTypeLabel(type) {
    const labels = {
        'invoice_sale': 'Factura Venta',
        'invoice_purchase': 'Factura Compra',
        'payment': 'Pago',
        'receipt': 'Recibo',
        'credit_note': 'Nota Crédito',
        'debit_note': 'Nota Débito'
    };
    return labels[type] || type;
}

function getStatusLabel(status) {
    const labels = {
        'DRAFT': 'Borrador',
        'PENDING': 'Pendiente',
        'PROCESSING': 'Procesando',
        'PROCESSED': 'Procesado',
        'ERROR': 'Error',
        'CANCELLED': 'Anulado'
    };
    return labels[status] || status;
}

function getTypeClass(type) {
    const classes = {
        'invoice_sale': 'success',
        'invoice_purchase': 'info',
        'payment': 'primary',
        'receipt': 'secondary',
        'credit_note': 'warning',
        'debit_note': 'danger'
    };
    return classes[type] || 'secondary';
}

function formatDate(dateStr) {
    const date = new Date(dateStr);
    return date.toLocaleDateString('es-CO');
}

function formatDateTime(dateStr) {
    const date = new Date(dateStr);
    return date.toLocaleString('es-CO');
}

function formatCurrency(amount) {
    return new Intl.NumberFormat('es-CO', {
        minimumFractionDigits: 0,
        maximumFractionDigits: 2
    }).format(amount);
}

function renderQuickViewItems(items) {
    if (!items || items.length === 0) return '';
    
    return `
        <div class="items-section">
            <h5>Detalle de Items</h5>
            <table class="table table-sm">
                <thead>
                    <tr>
                        <th>Descripción</th>
                        <th>Cantidad</th>
                        <th>Precio</th>
                        <th>Total</th>
                    </tr>
                </thead>
                <tbody>
                    ${items.map(item => `
                        <tr>
                            <td>${item.description}</td>
                            <td>${item.quantity}</td>
                            <td>$${formatCurrency(item.unit_price)}</td>
                            <td>$${formatCurrency(item.total)}</td>
                        </tr>
                    `).join('')}
                </tbody>
            </table>
        </div>
    `;
}

function renderQuickViewJournalEntry(entry) {
    if (!entry) return '';
    
    return `
        <div class="journal-entry-section">
            <h5>Asiento Contable</h5>
            <table class="table table-sm">
                <thead>
                    <tr>
                        <th>Cuenta</th>
                        <th>Descripción</th>
                        <th>Débito</th>
                        <th>Crédito</th>
                    </tr>
                </thead>
                <tbody>
                    ${entry.lines.map(line => `
                        <tr>
                            <td>${line.account_code}</td>
                            <td>${line.description}</td>
                            <td>${line.debit ? '$' + formatCurrency(line.debit) : ''}</td>
                            <td>${line.credit ? '$' + formatCurrency(line.credit) : ''}</td>
                        </tr>
                    `).join('')}
                </tbody>
            </table>
        </div>
    `;
}

function showLoading() {
    document.getElementById('loadingOverlay').style.display = 'flex';
}

function hideLoading() {
    document.getElementById('loadingOverlay').style.display = 'none';
}

function showSuccess(message) {
    // In real implementation, show toast notification
    console.log('Success:', message);
}

function showError(message) {
    // In real implementation, show toast notification
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