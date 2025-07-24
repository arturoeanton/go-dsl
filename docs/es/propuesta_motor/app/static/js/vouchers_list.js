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
    sortBy: 'date',
    sortOrder: 'desc',
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
    console.log('=== VOUCHERS LIST INITIALIZING ===');
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
    console.log('=== LOADING VOUCHERS DATA ===');
    showLoading();
    
    try {
        console.log('Calling API with params:', {
            page: state.currentPage,
            per_page: state.pageSize,
            sort: state.sortBy,
            order: state.sortOrder
        });
        
        // Fetch data using API service
        const result = await motorContableApi.vouchers.getList({
            page: state.currentPage,
            per_page: state.pageSize,
            sort: state.sortBy,
            order: state.sortOrder
        });
        
        console.log('API Response:', result);
        console.log('Vouchers data:', result.data);
        
        state.vouchers = result.data.vouchers;
        state.filteredVouchers = [...state.vouchers];
        
        console.log('Loaded vouchers:', state.vouchers.length);
        
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
    console.log('=== RENDERING VOUCHERS TABLE ===');
    console.log('Filtered vouchers:', state.filteredVouchers.length);
    
    const tbody = document.getElementById('vouchersTableBody');
    if (!tbody) {
        console.error('Table body not found!');
        return;
    }
    
    tbody.innerHTML = '';
    
    // Calculate pagination
    const startIdx = (state.currentPage - 1) * state.pageSize;
    const endIdx = startIdx + state.pageSize;
    const pageVouchers = state.filteredVouchers.slice(startIdx, endIdx);
    
    console.log('Rendering vouchers for page:', pageVouchers.length);
    
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
            <a href="#" onclick="showQuickView('${voucher.id}')">${voucher.number || voucher.voucher_number}</a>
        </td>
        <td>${formatDate(voucher.date || voucher.voucher_date)}</td>
        <td><span class="badge badge-${getTypeClass(voucher.voucher_type || voucher.type)}">${getVoucherTypeLabel(voucher.voucher_type || voucher.type)}</span></td>
        <td>${voucher.third_party ? 
            `${voucher.third_party.company_name || (voucher.third_party.first_name + ' ' + voucher.third_party.last_name)}<br><small class="text-muted">${voucher.third_party.document_number || ''}</small>` : 
            '<span class="text-muted">Sin tercero</span>'}</td>
        <td class="description-cell">${voucher.description}</td>
        <td class="text-right">
            $${formatCurrency(voucher.total_debit || voucher.total_amount || 0)}
            ${voucher.status === 'DRAFT' && voucher.is_balanced === false ? 
                '<br><small class="text-danger">⚠️ Desbalanceado</small>' : 
                voucher.status === 'DRAFT' ? '<br><small class="text-success">✅ Balanceado</small>' : ''}
        </td>
        <td><span class="status-badge status-${voucher.status.toLowerCase()}">${getStatusLabel(voucher.status)}</span></td>
        <td class="actions-cell">
            <button class="btn btn-sm btn-icon" title="Ver" onclick="showQuickView('${voucher.id}')">👁️</button>
            <button class="btn btn-sm btn-icon" title="Editar" onclick="editVoucher('${voucher.id}')">✏️</button>
            ${voucher.status === 'DRAFT' && (!voucher.is_balanced || voucher.is_balanced === false) ? 
                `<button class="btn btn-sm btn-warning" title="Recalcular con reglas DSL" onclick="recalculateVoucher('${voucher.id}')">🔄 Recalcular</button>` : ''}
            <button class="btn btn-sm ${voucher.status === 'DRAFT' ? 'btn-primary' : 'btn-secondary'}" 
                    title="${voucher.status === 'DRAFT' ? 'Procesar y generar asiento contable' : 'Ya procesado'}" 
                    onclick="processVoucher('${voucher.id}')"
                    ${voucher.status !== 'DRAFT' ? 'disabled' : ''}>
                ${voucher.status === 'DRAFT' ? '⚡ Procesar' : '✅ Procesado'}
            </button>
            ${voucher.status === 'POSTED' ? 
                `<button class="btn btn-sm btn-info" title="Ver asiento contable" onclick="showJournalEntry('${voucher.id}')">📊 Asiento</button>` : ''}
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
                    <span>${getVoucherTypeLabel(voucher.voucher_type || voucher.type)}</span>
                </div>
                <div class="info-group">
                    <label>Fecha:</label>
                    <span>${formatDate(voucher.date || voucher.voucher_date)}</span>
                </div>
                <div class="info-group">
                    <label>Tercero:</label>
                    <span>${voucher.third_party ? 
                        (voucher.third_party.company_name || (voucher.third_party.first_name + ' ' + voucher.third_party.last_name)) : 
                        'Sin tercero'}</span>
                </div>
                <div class="info-group">
                    <label>NIT/CC:</label>
                    <span>${voucher.third_party ? voucher.third_party.document_number : '-'}</span>
                </div>
                <div class="info-group">
                    <label>Monto Total:</label>
                    <span class="amount">$${formatCurrency(voucher.total_debit || voucher.total_amount || 0)}</span>
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
        // Llamada real a la API para contabilizar
        const response = await fetch(`/api/v1/vouchers/${voucherId}/post`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        
        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.message || 'Error al procesar el comprobante');
        }
        
        const result = await response.json();
        console.log('Comprobante procesado exitosamente:', result);
        
        // Mostrar mensaje de éxito primero
        showSuccess('Comprobante contabilizado exitosamente. Asiento contable generado.');
        
        // Pequeño delay para asegurar que el backend haya actualizado
        await new Promise(resolve => setTimeout(resolve, 500));
        
        // Limpiar cache para forzar nueva consulta
        console.log('Limpiando cache de API...');
        if (window.motorContableApi && window.motorContableApi.clearCache) {
            await window.motorContableApi.clearCache();
        }
        
        // Recargar la lista de comprobantes para reflejar el cambio
        console.log('Recargando lista de comprobantes...');
        await loadVouchersData();
        console.log('Lista de comprobantes recargada');
        
        // Verificar el estado del voucher específico que procesamos
        const processedVoucher = state.vouchers.find(v => v.id === voucherId);
        if (processedVoucher) {
            console.log(`Estado del voucher ${voucherId} después de recargar:`, processedVoucher.status);
        } else {
            console.log(`Voucher ${voucherId} no encontrado en la lista recargada`);
        }
        
    } catch (error) {
        console.error('Error procesando comprobante:', error);
        showError('Error al procesar el comprobante: ' + error.message);
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
 * Recalculate voucher with DSL
 */
async function recalculateVoucher(voucherId) {
    if (!confirm('¿Desea recalcular este comprobante aplicando las reglas DSL actuales?\n\nEsto actualizará automáticamente los impuestos y líneas contables.')) return;
    
    showLoading();
    
    try {
        console.log(`Recalculando comprobante ${voucherId} con DSL...`);
        
        // Llamada real a la API para recalcular
        const response = await fetch(`/api/v1/vouchers/${voucherId}/recalculate`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        
        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.message || 'Error al recalcular el comprobante');
        }
        
        const result = await response.json();
        console.log('Comprobante recalculado:', result);
        
        // Mostrar mensaje de éxito primero
        showSuccess('Comprobante recalculado exitosamente con reglas DSL.');
        
        // Limpiar cache para forzar nueva consulta
        console.log('Limpiando cache de API...');
        if (window.motorContableApi && window.motorContableApi.clearCache) {
            await window.motorContableApi.clearCache();
        }
        
        // Recargar la lista de comprobantes para reflejar los cambios
        console.log('Recargando lista después del recálculo...');
        await loadVouchersData();
        console.log('Lista recargada después del recálculo');
        
    } catch (error) {
        console.error('Error recalculando comprobante:', error);
        showError('Error al recalcular el comprobante: ' + error.message);
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
        'POSTED': 'Contabilizado',
        'CANCELLED': 'Cancelado',
        'ERROR': 'Error',
        'PENDING': 'Pendiente',
        'PROCESSING': 'Procesando'
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
    console.log('Success:', message);
    showToast(message, 'success');
}

function showError(message) {
    console.error('Error:', message);
    showToast(message, 'error');
}

function showToast(message, type = 'info') {
    // Remove any existing toast
    const existingToast = document.getElementById('toast-notification');
    if (existingToast) {
        existingToast.remove();
    }
    
    // Create toast element
    const toast = document.createElement('div');
    toast.id = 'toast-notification';
    toast.className = `toast toast-${type}`;
    
    const colors = {
        success: '#48bb78',
        error: '#f56565',
        info: '#667eea',
        warning: '#ed8936'
    };
    
    const icons = {
        success: '✅',
        error: '❌',
        info: 'ℹ️',
        warning: '⚠️'
    };
    
    toast.innerHTML = `
        <div class="toast-content">
            <span class="toast-icon">${icons[type]}</span>
            <span class="toast-message">${message}</span>
            <button class="toast-close" onclick="this.parentElement.parentElement.remove()">×</button>
        </div>
    `;
    
    // Add styles
    toast.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        background: ${colors[type]};
        color: white;
        padding: 16px 20px;
        border-radius: 8px;
        box-shadow: 0 4px 12px rgba(0,0,0,0.2);
        z-index: 10000;
        max-width: 400px;
        animation: slideInRight 0.3s ease;
        font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
    `;
    
    // Add CSS animation if not already added
    if (!document.getElementById('toast-styles')) {
        const style = document.createElement('style');
        style.id = 'toast-styles';
        style.textContent = `
            @keyframes slideInRight {
                from { transform: translateX(100%); opacity: 0; }
                to { transform: translateX(0); opacity: 1; }
            }
            @keyframes slideOutRight {
                from { transform: translateX(0); opacity: 1; }
                to { transform: translateX(100%); opacity: 0; }
            }
            .toast-content {
                display: flex;
                align-items: center;
                gap: 12px;
            }
            .toast-icon {
                font-size: 18px;
                flex-shrink: 0;
            }
            .toast-message {
                flex: 1;
                font-size: 14px;
                font-weight: 500;
            }
            .toast-close {
                background: none;
                border: none;
                color: white;
                font-size: 20px;
                cursor: pointer;
                padding: 0;
                margin-left: 8px;
                opacity: 0.8;
                flex-shrink: 0;
            }
            .toast-close:hover {
                opacity: 1;
            }
        `;
        document.head.appendChild(style);
    }
    
    // Add to page
    document.body.appendChild(toast);
    
    // Auto remove after 4 seconds
    setTimeout(() => {
        if (toast.parentElement) {
            toast.style.animation = 'slideOutRight 0.3s ease';
            setTimeout(() => {
                if (toast.parentElement) {
                    toast.remove();
                }
            }, 300);
        }
    }, 4000);
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

/**
 * Show journal entry modal
 */
async function showJournalEntry(voucherId) {
    console.log('=== SHOWING JOURNAL ENTRY ===');
    console.log('Voucher ID:', voucherId);
    
    showLoading();
    
    try {
        console.log('Calling journal entry API...');
        const response = await fetch(`/api/v1/vouchers/${voucherId}/journal-entry`);
        
        if (!response.ok) {
            const errorData = await response.json();
            if (response.status === 404) {
                showError('Este comprobante no tiene un asiento contable asociado.');
                return;
            }
            throw new Error(errorData.message || 'Error obteniendo asiento contable');
        }
        
        const result = await response.json();
        console.log('Journal entry data:', result.data);
        
        const { voucher, journal_entry } = result.data;
        
        // Create and show modal
        showJournalEntryModal(voucher, journal_entry);
        
    } catch (error) {
        console.error('Error loading journal entry:', error);
        showError('Error al cargar el asiento contable: ' + error.message);
    } finally {
        hideLoading();
    }
}

/**
 * Show journal entry modal with data
 */
function showJournalEntryModal(voucher, journalEntry) {
    // Remove existing modal if any
    const existingModal = document.getElementById('journalEntryModal');
    if (existingModal) {
        existingModal.remove();
    }
    
    // Calculate totals
    const totalDebit = journalEntry.lines.reduce((sum, line) => sum + line.debit, 0);
    const totalCredit = journalEntry.lines.reduce((sum, line) => sum + line.credit, 0);
    
    // Create modal HTML
    const modalHTML = `
        <div id="journalEntryModal" class="modal-overlay" onclick="closeJournalEntryModal()">
            <div class="modal-content journal-entry-modal" onclick="event.stopPropagation()">
                <div class="modal-header">
                    <h3>📊 Asiento Contable</h3>
                    <button class="modal-close" onclick="closeJournalEntryModal()">✕</button>
                </div>
                
                <div class="modal-body">
                    <!-- Voucher Info -->
                    <div class="voucher-info-section">
                        <h4>Información del Comprobante</h4>
                        <div class="info-grid">
                            <div class="info-item">
                                <label>Número:</label>
                                <span>${voucher.number || voucher.id}</span>
                            </div>
                            <div class="info-item">
                                <label>Tipo:</label>
                                <span>${getVoucherTypeLabel(voucher.voucher_type)}</span>
                            </div>
                            <div class="info-item">
                                <label>Fecha:</label>
                                <span>${formatDate(voucher.date)}</span>
                            </div>
                            <div class="info-item">
                                <label>Estado:</label>
                                <span class="status-badge status-${voucher.status.toLowerCase()}">${getStatusLabel(voucher.status)}</span>
                            </div>
                        </div>
                        <div class="description">
                            <label>Descripción:</label>
                            <p>${voucher.description}</p>
                        </div>
                    </div>
                    
                    <!-- Journal Entry Info -->
                    <div class="journal-entry-section">
                        <h4>Información del Asiento</h4>
                        <div class="info-grid">
                            <div class="info-item">
                                <label>Número de Asiento:</label>
                                <span>${journalEntry.entry_number}</span>
                            </div>
                            <div class="info-item">
                                <label>Fecha:</label>
                                <span>${formatDate(journalEntry.entry_date)}</span>
                            </div>
                            <div class="info-item">
                                <label>Estado:</label>
                                <span class="status-badge status-posted">Contabilizado</span>
                            </div>
                        </div>
                        <div class="description">
                            <label>Descripción:</label>
                            <p>${journalEntry.description}</p>
                        </div>
                    </div>
                    
                    <!-- Journal Lines Table -->
                    <div class="journal-lines-section">
                        <h4>Partidas del Asiento</h4>
                        <div class="table-responsive">
                            <table class="journal-lines-table">
                                <thead>
                                    <tr>
                                        <th>Cuenta</th>
                                        <th>Nombre de la Cuenta</th>
                                        <th>Descripción</th>
                                        <th class="text-right">Débito</th>
                                        <th class="text-right">Crédito</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    ${journalEntry.lines.map(line => `
                                        <tr>
                                            <td class="account-code">${line.account_code}</td>
                                            <td class="account-name">${line.account_name}</td>
                                            <td class="line-description">${line.description}</td>
                                            <td class="text-right debit-amount">
                                                ${line.debit ? '$' + formatCurrency(line.debit) : ''}
                                            </td>
                                            <td class="text-right credit-amount">
                                                ${line.credit ? '$' + formatCurrency(line.credit) : ''}
                                            </td>
                                        </tr>
                                    `).join('')}
                                </tbody>
                                <tfoot>
                                    <tr class="totals-row">
                                        <td colspan="3"><strong>TOTALES:</strong></td>
                                        <td class="text-right"><strong>$${formatCurrency(totalDebit)}</strong></td>
                                        <td class="text-right"><strong>$${formatCurrency(totalCredit)}</strong></td>
                                    </tr>
                                    <tr class="balance-check">
                                        <td colspan="3"><strong>Balance:</strong></td>
                                        <td colspan="2" class="text-center ${totalDebit === totalCredit ? 'balanced' : 'unbalanced'}">
                                            <strong>${totalDebit === totalCredit ? '✅ Balanceado' : '❌ Desbalanceado'}</strong>
                                        </td>
                                    </tr>
                                </tfoot>
                            </table>
                        </div>
                    </div>
                    
                    <!-- Timestamps -->
                    <div class="timestamps-section">
                        <small class="text-muted">
                            Creado: ${formatDateTime(journalEntry.created_at)} | 
                            Actualizado: ${formatDateTime(journalEntry.updated_at)}
                        </small>
                    </div>
                </div>
                
                <div class="modal-footer">
                    <button class="btn btn-secondary" onclick="closeJournalEntryModal()">Cerrar</button>
                    <button class="btn btn-primary" onclick="printJournalEntry()">🖨️ Imprimir</button>
                </div>
            </div>
        </div>
    `;
    
    // Add modal to page
    document.body.insertAdjacentHTML('beforeend', modalHTML);
    
    // Add CSS if not already added
    if (!document.getElementById('journal-entry-modal-styles')) {
        const style = document.createElement('style');
        style.id = 'journal-entry-modal-styles';
        style.textContent = `
            .modal-overlay {
                position: fixed;
                top: 0;
                left: 0;
                width: 100%;
                height: 100%;
                background: rgba(0, 0, 0, 0.7);
                display: flex;
                align-items: center;
                justify-content: center;
                z-index: 10000;
                animation: fadeIn 0.3s ease;
            }
            
            .journal-entry-modal {
                background: white;
                border-radius: 12px;
                box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
                max-width: 900px;
                width: 90%;
                max-height: 90vh;
                overflow-y: auto;
                animation: slideIn 0.3s ease;
            }
            
            .modal-header {
                padding: 20px 30px;
                border-bottom: 1px solid #e0e0e0;
                display: flex;
                justify-content: space-between;
                align-items: center;
                background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                color: white;
                border-radius: 12px 12px 0 0;
            }
            
            .modal-header h3 {
                margin: 0;
                font-size: 24px;
            }
            
            .modal-close {
                background: rgba(255, 255, 255, 0.2);
                border: none;
                color: white;
                font-size: 24px;
                cursor: pointer;
                padding: 5px 10px;
                border-radius: 50%;
                transition: background 0.3s;
            }
            
            .modal-close:hover {
                background: rgba(255, 255, 255, 0.3);
            }
            
            .modal-body {
                padding: 30px;
            }
            
            .voucher-info-section, .journal-entry-section {
                margin-bottom: 30px;
                padding: 20px;
                background: #f8f9fa;
                border-radius: 8px;
            }
            
            .journal-lines-section {
                margin-bottom: 20px;
            }
            
            .voucher-info-section h4, .journal-entry-section h4, .journal-lines-section h4 {
                margin: 0 0 15px 0;
                color: #333;
                border-bottom: 2px solid #667eea;
                padding-bottom: 5px;
            }
            
            .info-grid {
                display: grid;
                grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
                gap: 15px;
                margin-bottom: 15px;
            }
            
            .info-item {
                display: flex;
                flex-direction: column;
            }
            
            .info-item label {
                font-weight: 600;
                color: #555;
                margin-bottom: 5px;
                font-size: 14px;
            }
            
            .info-item span, .description p {
                font-size: 16px;
                color: #333;
            }
            
            .description {
                margin-top: 15px;
            }
            
            .description label {
                font-weight: 600;
                color: #555;
                margin-bottom: 5px;
                display: block;
            }
            
            .journal-lines-table {
                width: 100%;
                border-collapse: collapse;
                margin-top: 10px;
            }
            
            .journal-lines-table th,
            .journal-lines-table td {
                padding: 12px;
                text-align: left;
                border-bottom: 1px solid #ddd;
            }
            
            .journal-lines-table th {
                background: #f8f9fa;
                font-weight: 600;
                color: #333;
                border-bottom: 2px solid #667eea;
            }
            
            .journal-lines-table tbody tr:hover {
                background: #f8f9fa;
            }
            
            .account-code {
                font-family: monospace;
                font-weight: 600;
                color: #2c5aa0;
            }
            
            .account-name {
                font-weight: 500;
                color: #333;
            }
            
            .debit-amount, .credit-amount {
                font-family: monospace;
                font-weight: 600;
            }
            
            .debit-amount {
                color: #d32f2f;
            }
            
            .credit-amount {
                color: #388e3c;
            }
            
            .totals-row {
                background: #e3f2fd !important;
                border-top: 2px solid #667eea;
            }
            
            .balance-check {
                background: #f1f8e9 !important;
            }
            
            .balanced {
                color: #388e3c;
            }
            
            .unbalanced {
                color: #d32f2f;
            }
            
            .timestamps-section {
                text-align: center;
                padding-top: 20px;
                border-top: 1px solid #e0e0e0;
            }
            
            .modal-footer {
                padding: 20px 30px;
                border-top: 1px solid #e0e0e0;
                display: flex;
                justify-content: flex-end;
                gap: 10px;
                background: #f8f9fa;
                border-radius: 0 0 12px 12px;
            }
            
            @keyframes fadeIn {
                from { opacity: 0; }
                to { opacity: 1; }
            }
            
            @keyframes slideIn {
                from { transform: scale(0.8); opacity: 0; }
                to { transform: scale(1); opacity: 1; }
            }
            
            .btn-info {
                background: #17a2b8;
                color: white;
                border: 1px solid #17a2b8;
            }
            
            .btn-info:hover {
                background: #138496;
                border-color: #117a8b;
            }
        `;
        document.head.appendChild(style);
    }
}

/**
 * Close journal entry modal
 */
function closeJournalEntryModal() {
    const modal = document.getElementById('journalEntryModal');
    if (modal) {
        modal.style.animation = 'fadeIn 0.3s ease reverse';
        setTimeout(() => {
            modal.remove();
        }, 300);
    }
}

/**
 * Print journal entry (placeholder)
 */
function printJournalEntry() {
    // For now, just show a message
    showToast('Funcionalidad de impresión en desarrollo', 'info');
}