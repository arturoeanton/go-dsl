// Journal Entries JavaScript - Motor Contable
// Version: 1.0
// Last Updated: 2024-01-15

// Global state
const state = {
    entries: [],
    filteredEntries: [],
    currentPage: 1,
    pageSize: 20,
    totalPages: 1,
    selectedEntries: new Set(),
    currentEntryLines: [],
    accounts: [] // For account selection
};

// Initialize on DOM ready
document.addEventListener('DOMContentLoaded', () => {
    initializeJournalEntries();
});

/**
 * Initialize journal entries
 */
function initializeJournalEntries() {
    setupEventListeners();
    loadJournalData();
    loadAccounts();
    setDefaultDates();
}

/**
 * Setup event listeners
 */
function setupEventListeners() {
    // Filter toggle
    document.getElementById('toggleFilters').addEventListener('click', toggleFilters);
    
    // Select all checkbox
    document.getElementById('selectAll').addEventListener('change', handleSelectAll);
    
    // Pagination
    document.getElementById('prevPage').addEventListener('click', () => changePage(-1));
    document.getElementById('nextPage').addEventListener('click', () => changePage(1));
    
    // Search
    document.getElementById('searchInput').addEventListener('keyup', debounce(handleSearch, 300));
}

/**
 * Load journal data
 */
async function loadJournalData() {
    showLoading();
    
    try {
        const result = await motorContableApi.journalEntries.getList({
            page: state.currentPage,
            per_page: state.pageSize
        });
        
        state.entries = result.data.entries;
        state.filteredEntries = [...state.entries];
        
        updateSummaryCards(result.data.summary);
        renderJournalTable();
        updatePagination();
        checkBalance();
        
    } catch (error) {
        console.error('Error loading journal entries:', error);
        showError('Error al cargar los asientos contables');
    } finally {
        hideLoading();
    }
}

/**
 * Load accounts for selection
 */
async function loadAccounts() {
    try {
        const data = await motorContableApi.accounts.getTree();
        state.accounts = flattenAccountTree(data.data.accounts);
    } catch (error) {
        console.error('Error loading accounts:', error);
    }
}

/**
 * Flatten account tree for easier selection
 */
function flattenAccountTree(accounts, level = 0) {
    let flat = [];
    accounts.forEach(account => {
        flat.push({
            ...account,
            level,
            displayName: '&nbsp;'.repeat(level * 4) + account.account_code + ' - ' + account.name
        });
        if (account.children && account.children.length > 0) {
            flat = flat.concat(flattenAccountTree(account.children, level + 1));
        }
    });
    return flat;
}

/**
 * Render journal table
 */
function renderJournalTable() {
    const tbody = document.getElementById('journalTableBody');
    tbody.innerHTML = '';
    
    // Calculate pagination
    const startIdx = (state.currentPage - 1) * state.pageSize;
    const endIdx = startIdx + state.pageSize;
    
    // Get all lines from filtered entries
    let allLines = [];
    state.filteredEntries.forEach(entry => {
        entry.lines.forEach((line, index) => {
            allLines.push({
                ...line,
                entryId: entry.id,
                entryNumber: entry.entry_number,
                entryDate: entry.entry_date,
                voucherNumber: entry.voucher_number,
                entryStatus: entry.status,
                isFirstLine: index === 0,
                totalLines: entry.lines.length
            });
        });
    });
    
    const pageLines = allLines.slice(startIdx, endIdx);
    let currentEntryId = null;
    let pageDebits = 0;
    let pageCredits = 0;
    
    // Render lines
    pageLines.forEach(line => {
        const row = createJournalRow(line, currentEntryId !== line.entryId);
        tbody.appendChild(row);
        currentEntryId = line.entryId;
        
        pageDebits += line.debit || 0;
        pageCredits += line.credit || 0;
    });
    
    // Update totals
    document.getElementById('pageDebits').textContent = '$' + formatCurrency(pageDebits);
    document.getElementById('pageCredits').textContent = '$' + formatCurrency(pageCredits);
    
    // Update showing info
    document.getElementById('showingFrom').textContent = startIdx + 1;
    document.getElementById('showingTo').textContent = Math.min(endIdx, allLines.length);
    document.getElementById('totalRecords').textContent = allLines.length;
}

/**
 * Create journal table row
 */
function createJournalRow(line, showEntryInfo) {
    const tr = document.createElement('tr');
    tr.className = `journal-row ${showEntryInfo ? 'entry-start' : ''}`;
    
    if (showEntryInfo && line.totalLines > 1) {
        // First line of multi-line entry
        tr.innerHTML = `
            <td rowspan="${line.totalLines}">
                <input type="checkbox" class="entry-checkbox" value="${line.entryId}"
                    ${state.selectedEntries.has(line.entryId) ? 'checked' : ''}>
            </td>
            <td rowspan="${line.totalLines}">${formatDate(line.entryDate)}</td>
            <td rowspan="${line.totalLines}">
                <a href="#" onclick="showEntryDetail('${line.entryId}')">${line.entryNumber}</a>
            </td>
            <td rowspan="${line.totalLines}">
                ${line.voucherNumber ? 
                    `<a href="vouchers_form.html?id=${line.voucherNumber}">${line.voucherNumber}</a>` : 
                    '<span class="text-muted">Manual</span>'}
            </td>
            <td>${line.account_code}</td>
            <td>${line.account_name}</td>
            <td>${line.description}</td>
            <td class="text-right">${line.debit ? '$' + formatCurrency(line.debit) : ''}</td>
            <td class="text-right">${line.credit ? '$' + formatCurrency(line.credit) : ''}</td>
            <td rowspan="${line.totalLines}">
                <span class="status-badge status-${line.entryStatus.toLowerCase()}">${getStatusLabel(line.entryStatus)}</span>
            </td>
            <td rowspan="${line.totalLines}" class="actions-cell">
                <button class="btn btn-sm btn-icon" title="Ver" onclick="showEntryDetail('${line.entryId}')">👁️</button>
                <button class="btn btn-sm btn-icon" title="Copiar" onclick="copyEntry('${line.entryId}')">📋</button>
                <button class="btn btn-sm btn-icon btn-warning" title="Reversar" 
                    onclick="reverseEntry('${line.entryId}')"
                    ${line.entryStatus !== 'POSTED' ? 'disabled' : ''}>↩️</button>
            </td>
        `;
    } else if (showEntryInfo) {
        // Single line entry
        tr.innerHTML = `
            <td>
                <input type="checkbox" class="entry-checkbox" value="${line.entryId}"
                    ${state.selectedEntries.has(line.entryId) ? 'checked' : ''}>
            </td>
            <td>${formatDate(line.entryDate)}</td>
            <td>
                <a href="#" onclick="showEntryDetail('${line.entryId}')">${line.entryNumber}</a>
            </td>
            <td>
                ${line.voucherNumber ? 
                    `<a href="vouchers_form.html?id=${line.voucherNumber}">${line.voucherNumber}</a>` : 
                    '<span class="text-muted">Manual</span>'}
            </td>
            <td>${line.account_code}</td>
            <td>${line.account_name}</td>
            <td>${line.description}</td>
            <td class="text-right">${line.debit ? '$' + formatCurrency(line.debit) : ''}</td>
            <td class="text-right">${line.credit ? '$' + formatCurrency(line.credit) : ''}</td>
            <td>
                <span class="status-badge status-${line.entryStatus.toLowerCase()}">${getStatusLabel(line.entryStatus)}</span>
            </td>
            <td class="actions-cell">
                <button class="btn btn-sm btn-icon" title="Ver" onclick="showEntryDetail('${line.entryId}')">👁️</button>
                <button class="btn btn-sm btn-icon" title="Copiar" onclick="copyEntry('${line.entryId}')">📋</button>
                <button class="btn btn-sm btn-icon btn-warning" title="Reversar" 
                    onclick="reverseEntry('${line.entryId}')"
                    ${line.entryStatus !== 'POSTED' ? 'disabled' : ''}>↩️</button>
            </td>
        `;
    } else {
        // Additional lines of multi-line entry
        tr.innerHTML = `
            <td>${line.account_code}</td>
            <td>${line.account_name}</td>
            <td>${line.description}</td>
            <td class="text-right">${line.debit ? '$' + formatCurrency(line.debit) : ''}</td>
            <td class="text-right">${line.credit ? '$' + formatCurrency(line.credit) : ''}</td>
        `;
    }
    
    // Add checkbox event listener if present
    const checkbox = tr.querySelector('.entry-checkbox');
    if (checkbox) {
        checkbox.addEventListener('change', () => handleEntrySelection(line.entryId, checkbox.checked));
    }
    
    return tr;
}

/**
 * Show manual entry modal
 */
function showManualEntry() {
    const modal = document.getElementById('manualEntryModal');
    
    // Reset form
    document.getElementById('manualEntryForm').reset();
    document.getElementById('entryDate').value = new Date().toISOString().split('T')[0];
    
    // Clear lines
    state.currentEntryLines = [];
    document.getElementById('entryLinesBody').innerHTML = '';
    
    // Add two initial lines
    addEntryLine();
    addEntryLine();
    
    modal.style.display = 'block';
}

/**
 * Close manual entry modal
 */
function closeManualEntry() {
    document.getElementById('manualEntryModal').style.display = 'none';
}

/**
 * Add entry line
 */
function addEntryLine() {
    const lineId = Date.now();
    state.currentEntryLines.push({
        id: lineId,
        account_code: '',
        description: '',
        debit: 0,
        credit: 0,
        cost_center: ''
    });
    
    const tbody = document.getElementById('entryLinesBody');
    const tr = document.createElement('tr');
    tr.id = `line-${lineId}`;
    
    tr.innerHTML = `
        <td>
            <select class="form-control account-select" onchange="updateLineAccount(${lineId}, this.value)">
                <option value="">Seleccionar cuenta...</option>
                ${state.accounts.filter(a => a.is_detail).map(account => 
                    `<option value="${account.account_code}">${account.displayName}</option>`
                ).join('')}
            </select>
        </td>
        <td>
            <input type="text" class="form-control" placeholder="Descripción"
                   onchange="updateLineDescription(${lineId}, this.value)">
        </td>
        <td>
            <input type="number" class="form-control text-right" placeholder="0"
                   onchange="updateLineDebit(${lineId}, this.value)">
        </td>
        <td>
            <input type="number" class="form-control text-right" placeholder="0"
                   onchange="updateLineCredit(${lineId}, this.value)">
        </td>
        <td>
            <input type="text" class="form-control" placeholder="Centro costo"
                   onchange="updateLineCostCenter(${lineId}, this.value)">
        </td>
        <td>
            <button class="btn btn-sm btn-danger" onclick="removeEntryLine(${lineId})">🗑️</button>
        </td>
    `;
    
    tbody.appendChild(tr);
}

/**
 * Remove entry line
 */
function removeEntryLine(lineId) {
    state.currentEntryLines = state.currentEntryLines.filter(l => l.id !== lineId);
    document.getElementById(`line-${lineId}`).remove();
    updateModalTotals();
}

/**
 * Update line account
 */
function updateLineAccount(lineId, accountCode) {
    const line = state.currentEntryLines.find(l => l.id === lineId);
    if (line) {
        line.account_code = accountCode;
        const account = state.accounts.find(a => a.account_code === accountCode);
        if (account) {
            line.account_name = account.name;
        }
    }
}

/**
 * Update line description
 */
function updateLineDescription(lineId, description) {
    const line = state.currentEntryLines.find(l => l.id === lineId);
    if (line) line.description = description;
}

/**
 * Update line debit
 */
function updateLineDebit(lineId, value) {
    const line = state.currentEntryLines.find(l => l.id === lineId);
    if (line) {
        line.debit = parseFloat(value) || 0;
        if (line.debit > 0) line.credit = 0;
        
        // Update credit field
        const tr = document.getElementById(`line-${lineId}`);
        tr.querySelector('td:nth-child(4) input').value = '';
    }
    updateModalTotals();
}

/**
 * Update line credit
 */
function updateLineCredit(lineId, value) {
    const line = state.currentEntryLines.find(l => l.id === lineId);
    if (line) {
        line.credit = parseFloat(value) || 0;
        if (line.credit > 0) line.debit = 0;
        
        // Update debit field
        const tr = document.getElementById(`line-${lineId}`);
        tr.querySelector('td:nth-child(3) input').value = '';
    }
    updateModalTotals();
}

/**
 * Update line cost center
 */
function updateLineCostCenter(lineId, costCenter) {
    const line = state.currentEntryLines.find(l => l.id === lineId);
    if (line) line.cost_center = costCenter;
}

/**
 * Update modal totals
 */
function updateModalTotals() {
    const totalDebits = state.currentEntryLines.reduce((sum, line) => sum + (line.debit || 0), 0);
    const totalCredits = state.currentEntryLines.reduce((sum, line) => sum + (line.credit || 0), 0);
    
    document.getElementById('totalDebitsModal').textContent = '$' + formatCurrency(totalDebits);
    document.getElementById('totalCreditsModal').textContent = '$' + formatCurrency(totalCredits);
    
    const balanceCheck = document.getElementById('balanceCheckModal');
    if (Math.abs(totalDebits - totalCredits) < 0.01) {
        balanceCheck.innerHTML = '✅ Balanceado';
        balanceCheck.className = 'balance-indicator balanced';
    } else {
        const diff = totalDebits - totalCredits;
        balanceCheck.innerHTML = `❌ Diferencia: $${formatCurrency(Math.abs(diff))}`;
        balanceCheck.className = 'balance-indicator unbalanced';
    }
}

/**
 * Save manual entry
 */
async function saveManualEntry() {
    // Validate form
    const entryDate = document.getElementById('entryDate').value;
    const entryDescription = document.getElementById('entryDescription').value;
    
    if (!entryDate || !entryDescription) {
        showError('Por favor complete todos los campos requeridos');
        return;
    }
    
    // Validate lines
    const validLines = state.currentEntryLines.filter(l => 
        l.account_code && (l.debit > 0 || l.credit > 0)
    );
    
    if (validLines.length < 2) {
        showError('El asiento debe tener al menos 2 líneas válidas');
        return;
    }
    
    // Check balance
    const totalDebits = validLines.reduce((sum, l) => sum + (l.debit || 0), 0);
    const totalCredits = validLines.reduce((sum, l) => sum + (l.credit || 0), 0);
    
    if (Math.abs(totalDebits - totalCredits) > 0.01) {
        showError('El asiento debe estar balanceado');
        return;
    }
    
    showLoading();
    
    try {
        // Simulate API call
        await new Promise(resolve => setTimeout(resolve, 1000));
        
        // Create new entry
        const newEntry = {
            id: 'je-' + Date.now(),
            entry_number: 'AS-2024-' + String(state.entries.length + 1).padStart(4, '0'),
            entry_date: entryDate,
            description: entryDescription,
            voucher_number: null,
            status: 'DRAFT',
            lines: validLines,
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString()
        };
        
        state.entries.unshift(newEntry);
        state.filteredEntries = [...state.entries];
        
        renderJournalTable();
        updateSummaryCards();
        closeManualEntry();
        showSuccess('Asiento creado correctamente');
        
    } catch (error) {
        showError('Error al crear el asiento');
    } finally {
        hideLoading();
    }
}

/**
 * Show entry detail
 */
function showEntryDetail(entryId) {
    const entry = state.entries.find(e => e.id === entryId);
    if (!entry) return;
    
    const modal = document.getElementById('entryDetailModal');
    const content = document.getElementById('entryDetailContent');
    
    content.innerHTML = `
        <div class="entry-detail">
            <div class="entry-header">
                <h4>${entry.entry_number}</h4>
                <span class="status-badge status-${entry.status.toLowerCase()}">${getStatusLabel(entry.status)}</span>
            </div>
            
            <div class="entry-info">
                <div class="info-group">
                    <label>Fecha:</label>
                    <span>${formatDate(entry.entry_date)}</span>
                </div>
                <div class="info-group">
                    <label>Comprobante:</label>
                    <span>${entry.voucher_number || 'Asiento Manual'}</span>
                </div>
                <div class="info-group">
                    <label>Descripción:</label>
                    <span>${entry.description}</span>
                </div>
            </div>
            
            <h5>Líneas del Asiento</h5>
            <table class="table">
                <thead>
                    <tr>
                        <th>Cuenta</th>
                        <th>Nombre</th>
                        <th>Descripción</th>
                        <th>Débito</th>
                        <th>Crédito</th>
                    </tr>
                </thead>
                <tbody>
                    ${entry.lines.map(line => `
                        <tr>
                            <td>${line.account_code}</td>
                            <td>${line.account_name}</td>
                            <td>${line.description}</td>
                            <td class="text-right">${line.debit ? '$' + formatCurrency(line.debit) : ''}</td>
                            <td class="text-right">${line.credit ? '$' + formatCurrency(line.credit) : ''}</td>
                        </tr>
                    `).join('')}
                </tbody>
                <tfoot>
                    <tr>
                        <td colspan="3" class="text-right"><strong>Totales:</strong></td>
                        <td class="text-right"><strong>$${formatCurrency(
                            entry.lines.reduce((sum, l) => sum + (l.debit || 0), 0)
                        )}</strong></td>
                        <td class="text-right"><strong>$${formatCurrency(
                            entry.lines.reduce((sum, l) => sum + (l.credit || 0), 0)
                        )}</strong></td>
                    </tr>
                </tfoot>
            </table>
            
            <div class="timestamps">
                <small>Creado: ${formatDateTime(entry.created_at)} | 
                Actualizado: ${formatDateTime(entry.updated_at)}</small>
            </div>
        </div>
    `;
    
    // Update reverse button
    document.getElementById('reverseBtn').disabled = entry.status !== 'POSTED';
    document.getElementById('reverseBtn').onclick = () => {
        closeEntryDetail();
        reverseEntry(entryId);
    };
    
    modal.style.display = 'block';
}

/**
 * Close entry detail
 */
function closeEntryDetail() {
    document.getElementById('entryDetailModal').style.display = 'none';
}

/**
 * Reverse entry
 */
async function reverseEntry(entryId) {
    const entry = state.entries.find(e => e.id === entryId);
    if (!entry || entry.status !== 'POSTED') return;
    
    if (!confirm('¿Está seguro de reversar este asiento?')) return;
    
    showLoading();
    
    try {
        // Simulate API call
        await new Promise(resolve => setTimeout(resolve, 1000));
        
        // Create reversal entry
        const reversalEntry = {
            id: 'je-rev-' + Date.now(),
            entry_number: 'REV-' + entry.entry_number,
            entry_date: new Date().toISOString().split('T')[0],
            description: `Reversión de ${entry.entry_number} - ${entry.description}`,
            voucher_number: entry.voucher_number,
            status: 'POSTED',
            original_entry_id: entryId,
            lines: entry.lines.map(line => ({
                ...line,
                debit: line.credit,
                credit: line.debit,
                description: `Reversión - ${line.description}`
            })),
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString()
        };
        
        // Update original entry status
        entry.status = 'REVERSED';
        
        // Add reversal entry
        state.entries.unshift(reversalEntry);
        state.filteredEntries = [...state.entries];
        
        renderJournalTable();
        updateSummaryCards();
        showSuccess('Asiento reversado correctamente');
        
    } catch (error) {
        showError('Error al reversar el asiento');
    } finally {
        hideLoading();
    }
}

/**
 * Copy entry
 */
function copyEntry(entryId) {
    const entry = state.entries.find(e => e.id === entryId);
    if (!entry) return;
    
    // Populate manual entry form with copied data
    showManualEntry();
    
    document.getElementById('entryDescription').value = `Copia de ${entry.description}`;
    
    // Clear default lines
    state.currentEntryLines = [];
    document.getElementById('entryLinesBody').innerHTML = '';
    
    // Copy lines
    entry.lines.forEach(line => {
        addEntryLine();
        const lineId = state.currentEntryLines[state.currentEntryLines.length - 1].id;
        
        updateLineAccount(lineId, line.account_code);
        updateLineDescription(lineId, line.description);
        if (line.debit > 0) updateLineDebit(lineId, line.debit);
        if (line.credit > 0) updateLineCredit(lineId, line.credit);
        if (line.cost_center) updateLineCostCenter(lineId, line.cost_center);
    });
    
    updateModalTotals();
}

/**
 * Apply filters
 */
function applyFilters() {
    const filters = {
        dateFrom: document.getElementById('dateFrom').value,
        dateTo: document.getElementById('dateTo').value,
        entryNumber: document.getElementById('entryNumber').value.toLowerCase(),
        accountCode: document.getElementById('accountCode').value.toLowerCase(),
        sourceVoucher: document.getElementById('sourceVoucher').value.toLowerCase(),
        amountFrom: parseFloat(document.getElementById('amountFrom').value) || null,
        amountTo: parseFloat(document.getElementById('amountTo').value) || null,
        status: document.getElementById('entryStatus').value
    };
    
    state.filteredEntries = state.entries.filter(entry => {
        // Date filter
        if (filters.dateFrom && entry.entry_date < filters.dateFrom) return false;
        if (filters.dateTo && entry.entry_date > filters.dateTo) return false;
        
        // Entry number filter
        if (filters.entryNumber && !entry.entry_number.toLowerCase().includes(filters.entryNumber)) return false;
        
        // Voucher filter
        if (filters.sourceVoucher && entry.voucher_number && 
            !entry.voucher_number.toLowerCase().includes(filters.sourceVoucher)) return false;
        
        // Status filter
        if (filters.status && entry.status !== filters.status) return false;
        
        // Account and amount filters (check in lines)
        if (filters.accountCode || filters.amountFrom !== null || filters.amountTo !== null) {
            const hasMatchingLine = entry.lines.some(line => {
                if (filters.accountCode && 
                    !line.account_code.toLowerCase().includes(filters.accountCode) &&
                    !line.account_name.toLowerCase().includes(filters.accountCode)) {
                    return false;
                }
                
                const amount = line.debit || line.credit || 0;
                if (filters.amountFrom !== null && amount < filters.amountFrom) return false;
                if (filters.amountTo !== null && amount > filters.amountTo) return false;
                
                return true;
            });
            
            if (!hasMatchingLine) return false;
        }
        
        return true;
    });
    
    state.currentPage = 1;
    renderJournalTable();
    updatePagination();
    updateSummaryCards();
}

/**
 * Update summary cards
 */
function updateSummaryCards(summary) {
    if (summary) {
        document.getElementById('totalEntries').textContent = formatNumber(summary.total_entries);
        document.getElementById('totalDebits').textContent = '$' + formatCurrency(summary.total_debits);
        document.getElementById('totalCredits').textContent = '$' + formatCurrency(summary.total_credits);
        document.getElementById('balance').textContent = '$' + formatCurrency(summary.balance);
    } else {
        // Calculate from filtered entries
        const totalEntries = state.filteredEntries.length;
        let totalDebits = 0;
        let totalCredits = 0;
        
        state.filteredEntries.forEach(entry => {
            entry.lines.forEach(line => {
                totalDebits += line.debit || 0;
                totalCredits += line.credit || 0;
            });
        });
        
        document.getElementById('totalEntries').textContent = formatNumber(totalEntries);
        document.getElementById('totalDebits').textContent = '$' + formatCurrency(totalDebits);
        document.getElementById('totalCredits').textContent = '$' + formatCurrency(totalCredits);
        document.getElementById('balance').textContent = '$' + formatCurrency(Math.abs(totalDebits - totalCredits));
    }
}

/**
 * Check balance
 */
function checkBalance() {
    let totalDebits = 0;
    let totalCredits = 0;
    
    state.filteredEntries.forEach(entry => {
        entry.lines.forEach(line => {
            totalDebits += line.debit || 0;
            totalCredits += line.credit || 0;
        });
    });
    
    const balanceStatus = document.getElementById('balanceStatus');
    const diff = Math.abs(totalDebits - totalCredits);
    
    if (diff < 0.01) {
        balanceStatus.innerHTML = `
            <span class="status-icon">✅</span>
            <span class="status-text">Asientos balanceados correctamente</span>
        `;
        balanceStatus.className = 'balance-status balanced';
    } else {
        balanceStatus.innerHTML = `
            <span class="status-icon">❌</span>
            <span class="status-text">Diferencia de balance: $${formatCurrency(diff)}</span>
        `;
        balanceStatus.className = 'balance-status unbalanced';
    }
}

/**
 * Export to Excel
 */
function exportToExcel() {
    // Create CSV content
    const headers = ['Fecha', 'Número', 'Comprobante', 'Cuenta', 'Nombre Cuenta', 'Descripción', 'Débito', 'Crédito', 'Estado'];
    const rows = [];
    
    state.filteredEntries.forEach(entry => {
        entry.lines.forEach(line => {
            rows.push([
                entry.entry_date,
                entry.entry_number,
                entry.voucher_number || 'Manual',
                line.account_code,
                line.account_name,
                line.description,
                line.debit || '',
                line.credit || '',
                entry.status
            ]);
        });
    });
    
    const csv = [headers, ...rows].map(row => row.join(',')).join('\n');
    
    // Download CSV
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `asientos_contables_${new Date().toISOString().split('T')[0]}.csv`;
    a.click();
    
    showSuccess('Asientos exportados correctamente');
}

/**
 * Print entries
 */
function printEntries() {
    window.print();
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
        state.filteredEntries = state.entries.filter(entry => {
            // Search in entry fields
            if (entry.entry_number.toLowerCase().includes(searchTerm) ||
                entry.description.toLowerCase().includes(searchTerm) ||
                (entry.voucher_number && entry.voucher_number.toLowerCase().includes(searchTerm))) {
                return true;
            }
            
            // Search in lines
            return entry.lines.some(line => 
                line.account_code.toLowerCase().includes(searchTerm) ||
                line.account_name.toLowerCase().includes(searchTerm) ||
                line.description.toLowerCase().includes(searchTerm)
            );
        });
    } else {
        state.filteredEntries = [...state.entries];
    }
    
    state.currentPage = 1;
    renderJournalTable();
    updatePagination();
    updateSummaryCards();
}

/**
 * Handle select all
 */
function handleSelectAll(event) {
    const isChecked = event.target.checked;
    const checkboxes = document.querySelectorAll('.entry-checkbox');
    
    checkboxes.forEach(cb => {
        cb.checked = isChecked;
        const entryId = cb.value;
        if (isChecked) {
            state.selectedEntries.add(entryId);
        } else {
            state.selectedEntries.delete(entryId);
        }
    });
}

/**
 * Handle entry selection
 */
function handleEntrySelection(entryId, isChecked) {
    if (isChecked) {
        state.selectedEntries.add(entryId);
    } else {
        state.selectedEntries.delete(entryId);
    }
    
    // Update select all checkbox
    const checkboxes = document.querySelectorAll('.entry-checkbox');
    const checkedCount = document.querySelectorAll('.entry-checkbox:checked').length;
    const selectAll = document.getElementById('selectAll');
    
    selectAll.checked = checkboxes.length > 0 && checkedCount === checkboxes.length;
    selectAll.indeterminate = checkedCount > 0 && checkedCount < checkboxes.length;
}

/**
 * Update pagination
 */
function updatePagination() {
    // Calculate total lines
    let totalLines = 0;
    state.filteredEntries.forEach(entry => {
        totalLines += entry.lines.length;
    });
    
    state.totalPages = Math.ceil(totalLines / state.pageSize);
    
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
    renderJournalTable();
    updatePagination();
}

/**
 * Set default dates
 */
function setDefaultDates() {
    const today = new Date();
    const firstDay = new Date(today.getFullYear(), today.getMonth(), 1);
    
    document.getElementById('dateFrom').value = firstDay.toISOString().split('T')[0];
    document.getElementById('dateTo').value = today.toISOString().split('T')[0];
    document.getElementById('entryDate').value = today.toISOString().split('T')[0];
}

/**
 * Show import dialog
 */
function showImportDialog() {
    alert('Funcionalidad de importación próximamente');
}

// Helper functions
function getStatusLabel(status) {
    const labels = {
        'DRAFT': 'Borrador',
        'POSTED': 'Mayorizado',
        'REVERSED': 'Reversado'
    };
    return labels[status] || status;
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

function formatNumber(num) {
    return new Intl.NumberFormat('es-CO').format(num);
}

function showLoading() {
    document.getElementById('loadingOverlay').style.display = 'flex';
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