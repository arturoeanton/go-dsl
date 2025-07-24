// Accounts Chart JavaScript - Motor Contable
// Version: 1.0
// Last Updated: 2024-01-15

// Global state
const state = {
    accounts: [],
    accountsMap: {},
    filteredAccounts: [],
    selectedAccount: null,
    currentView: 'tree',
    expandedNodes: new Set(),
    searchTerm: ''
};

// Initialize on DOM ready
document.addEventListener('DOMContentLoaded', () => {
    initializeAccountsChart();
});

/**
 * Initialize accounts chart
 */
function initializeAccountsChart() {
    setupEventListeners();
    loadAccountsData();
}

/**
 * Setup event listeners
 */
function setupEventListeners() {
    // Search
    document.getElementById('searchInput').addEventListener('keyup', debounce(handleSearch, 300));
    
    // Parent account selector
    document.getElementById('parentAccount').addEventListener('change', updateFullCode);
    
    // Account code input
    document.getElementById('accountCode').addEventListener('input', updateFullCode);
    
    // Account type change
    document.getElementById('accountType').addEventListener('change', updateNatureDefault);
    
    // File upload
    const fileInput = document.getElementById('fileInput');
    const uploadArea = document.getElementById('fileUploadArea');
    
    fileInput.addEventListener('change', handleFileSelect);
    
    // Drag and drop
    uploadArea.addEventListener('dragover', (e) => {
        e.preventDefault();
        uploadArea.classList.add('drag-over');
    });
    
    uploadArea.addEventListener('dragleave', () => {
        uploadArea.classList.remove('drag-over');
    });
    
    uploadArea.addEventListener('drop', (e) => {
        e.preventDefault();
        uploadArea.classList.remove('drag-over');
        handleFileDrop(e);
    });
    
    // Context menu
    document.addEventListener('contextmenu', (e) => {
        if (e.target.closest('.tree-node')) {
            e.preventDefault();
            showContextMenu(e);
        }
    });
    
    document.addEventListener('click', hideContextMenu);
}

/**
 * Load accounts data
 */
async function loadAccountsData() {
    showLoading();
    
    try {
        const response = await fetch('../../api/accounts_tree.json');
        if (!response.ok) throw new Error('Failed to load accounts');
        
        const data = await response.json();
        state.accounts = data.data.accounts;
        
        // Build accounts map
        buildAccountsMap(state.accounts);
        
        // Update stats
        updateStats();
        
        // Render tree
        renderAccountsTree();
        
        // Populate parent selector
        populateParentSelector();
        
    } catch (error) {
        console.error('Error loading accounts:', error);
        showError('Error al cargar el plan de cuentas');
    } finally {
        hideLoading();
    }
}

/**
 * Build accounts map for quick lookup
 */
function buildAccountsMap(accounts, parentPath = '') {
    accounts.forEach(account => {
        const fullPath = parentPath ? `${parentPath}.${account.account_code}` : account.account_code;
        state.accountsMap[account.id] = {
            ...account,
            fullPath
        };
        
        if (account.children && account.children.length > 0) {
            buildAccountsMap(account.children, fullPath);
        }
    });
}

/**
 * Render accounts tree
 */
function renderAccountsTree() {
    const treeContainer = document.getElementById('accountsTree');
    treeContainer.innerHTML = '';
    
    const filteredAccounts = filterAccounts(state.accounts);
    filteredAccounts.forEach(account => {
        const node = createTreeNode(account, 0);
        treeContainer.appendChild(node);
    });
}

/**
 * Create tree node
 */
function createTreeNode(account, level) {
    const node = document.createElement('div');
    node.className = 'tree-node';
    node.dataset.accountId = account.id;
    node.style.paddingLeft = `${level * 20}px`;
    
    const hasChildren = account.children && account.children.length > 0;
    const isExpanded = state.expandedNodes.has(account.id);
    
    node.innerHTML = `
        <span class="tree-toggle" onclick="toggleNode('${account.id}')">
            ${hasChildren ? (isExpanded ? '▼' : '▶') : '&nbsp;'}
        </span>
        <span class="tree-icon">${account.is_detail ? '📄' : '📁'}</span>
        <span class="tree-label" onclick="selectAccount('${account.id}')">
            <strong>${account.account_code}</strong> - ${account.name}
        </span>
        <span class="account-type ${account.type.toLowerCase()}">${getAccountTypeLabel(account.type)}</span>
        <span class="account-nature nature-${account.nature.toLowerCase()}">${account.nature}</span>
        ${account.balance ? `<span class="account-balance">$${formatCurrency(Math.abs(account.balance))}</span>` : ''}
        ${!account.is_active ? '<span class="inactive-badge">Inactiva</span>' : ''}
    `;
    
    // Add children container
    if (hasChildren) {
        const childrenContainer = document.createElement('div');
        childrenContainer.className = `tree-children ${isExpanded ? 'expanded' : ''}`;
        childrenContainer.id = `children-${account.id}`;
        
        const filteredChildren = filterAccounts(account.children);
        filteredChildren.forEach(child => {
            const childNode = createTreeNode(child, level + 1);
            childrenContainer.appendChild(childNode);
        });
        
        node.appendChild(childrenContainer);
    }
    
    return node;
}

/**
 * Toggle tree node
 */
function toggleNode(accountId) {
    const childrenContainer = document.getElementById(`children-${accountId}`);
    if (!childrenContainer) return;
    
    if (state.expandedNodes.has(accountId)) {
        state.expandedNodes.delete(accountId);
        childrenContainer.classList.remove('expanded');
    } else {
        state.expandedNodes.add(accountId);
        childrenContainer.classList.add('expanded');
    }
    
    // Update toggle icon
    const node = document.querySelector(`[data-account-id="${accountId}"]`);
    const toggle = node.querySelector('.tree-toggle');
    toggle.textContent = state.expandedNodes.has(accountId) ? '▼' : '▶';
}

/**
 * Select account
 */
function selectAccount(accountId) {
    // Remove previous selection
    document.querySelectorAll('.tree-node').forEach(node => {
        node.classList.remove('selected');
    });
    
    // Add selection
    const node = document.querySelector(`[data-account-id="${accountId}"]`);
    if (node) {
        node.classList.add('selected');
        state.selectedAccount = state.accountsMap[accountId];
    }
}

/**
 * Show account modal
 */
function showAccountModal(account = null) {
    const modal = document.getElementById('accountModal');
    const form = document.getElementById('accountForm');
    
    // Reset form
    form.reset();
    
    if (account) {
        // Edit mode
        document.getElementById('modalTitle').textContent = 'Editar Cuenta';
        document.getElementById('accountCode').value = account.account_code;
        document.getElementById('accountName').value = account.name;
        document.getElementById('accountType').value = account.type;
        document.getElementById('accountNature').value = account.nature;
        document.getElementById('accountDescription').value = account.description || '';
        document.querySelector(`input[name="isDetail"][value="${account.is_detail}"]`).checked = true;
        document.querySelector(`input[name="isActive"][value="${account.is_active}"]`).checked = true;
        
        // Set parent
        if (account.parent_id) {
            document.getElementById('parentAccount').value = account.parent_id;
        }
    } else {
        // Create mode
        document.getElementById('modalTitle').textContent = 'Nueva Cuenta';
        
        // If an account is selected, use it as parent
        if (state.selectedAccount) {
            document.getElementById('parentAccount').value = state.selectedAccount.id;
        }
    }
    
    updateFullCode();
    modal.style.display = 'block';
}

/**
 * Close account modal
 */
function closeAccountModal() {
    document.getElementById('accountModal').style.display = 'none';
}

/**
 * Save account
 */
async function saveAccount() {
    const form = document.getElementById('accountForm');
    if (!form.checkValidity()) {
        form.reportValidity();
        return;
    }
    
    const accountData = {
        parent_id: document.getElementById('parentAccount').value || null,
        account_code: document.getElementById('accountCode').value,
        name: document.getElementById('accountName').value,
        type: document.getElementById('accountType').value,
        nature: document.getElementById('accountNature').value,
        description: document.getElementById('accountDescription').value,
        is_detail: document.querySelector('input[name="isDetail"]:checked').value === 'true',
        is_active: document.querySelector('input[name="isActive"]:checked').value === 'true'
    };
    
    showLoading();
    
    try {
        // Simulate API call
        await new Promise(resolve => setTimeout(resolve, 1000));
        
        // Create new account
        const newAccount = {
            id: 'acc-' + Date.now(),
            ...accountData,
            level: accountData.parent_id ? state.accountsMap[accountData.parent_id].level + 1 : 1,
            children: [],
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString()
        };
        
        // Add to tree
        if (accountData.parent_id) {
            const parent = findAccountInTree(state.accounts, accountData.parent_id);
            if (parent) {
                if (!parent.children) parent.children = [];
                parent.children.push(newAccount);
                parent.children.sort((a, b) => a.account_code.localeCompare(b.account_code));
            }
        } else {
            state.accounts.push(newAccount);
            state.accounts.sort((a, b) => a.account_code.localeCompare(b.account_code));
        }
        
        // Rebuild map
        state.accountsMap = {};
        buildAccountsMap(state.accounts);
        
        // Re-render
        renderAccountsTree();
        updateStats();
        populateParentSelector();
        
        closeAccountModal();
        showSuccess('Cuenta creada correctamente');
        
    } catch (error) {
        showError('Error al guardar la cuenta');
    } finally {
        hideLoading();
    }
}

/**
 * Find account in tree
 */
function findAccountInTree(accounts, accountId) {
    for (let account of accounts) {
        if (account.id === accountId) return account;
        if (account.children) {
            const found = findAccountInTree(account.children, accountId);
            if (found) return found;
        }
    }
    return null;
}

/**
 * Update full code preview
 */
function updateFullCode() {
    const parentId = document.getElementById('parentAccount').value;
    const code = document.getElementById('accountCode').value;
    const fullCodeSpan = document.getElementById('fullCode');
    
    if (parentId && state.accountsMap[parentId]) {
        const parent = state.accountsMap[parentId];
        fullCodeSpan.textContent = `${parent.fullPath}.${code}`;
    } else {
        fullCodeSpan.textContent = code;
    }
}

/**
 * Update nature default based on type
 */
function updateNatureDefault() {
    const type = document.getElementById('accountType').value;
    const natureSelect = document.getElementById('accountNature');
    
    const defaults = {
        'ASSET': 'D',
        'LIABILITY': 'C',
        'EQUITY': 'C',
        'INCOME': 'C',
        'EXPENSE': 'D'
    };
    
    if (defaults[type]) {
        natureSelect.value = defaults[type];
    }
}

/**
 * Populate parent selector
 */
function populateParentSelector() {
    const select = document.getElementById('parentAccount');
    const currentValue = select.value;
    
    select.innerHTML = '<option value="">Cuenta de Nivel Superior (Raíz)</option>';
    
    // Add all non-detail accounts
    addAccountsToSelect(state.accounts, select, 0);
    
    // Restore value
    select.value = currentValue;
}

/**
 * Add accounts to select recursively
 */
function addAccountsToSelect(accounts, select, level) {
    accounts.forEach(account => {
        if (!account.is_detail) {
            const option = document.createElement('option');
            option.value = account.id;
            option.textContent = '  '.repeat(level) + account.account_code + ' - ' + account.name;
            select.appendChild(option);
            
            if (account.children) {
                addAccountsToSelect(account.children, select, level + 1);
            }
        }
    });
}

/**
 * Filter accounts based on current filters
 */
function filterAccounts(accounts) {
    const filterType = document.getElementById('filterType').value;
    const filterStatus = document.getElementById('filterStatus').value;
    
    return accounts.filter(account => {
        // Type filter
        if (filterType && account.type !== filterType) {
            // Check if any child matches
            if (account.children) {
                account.children = filterAccounts(account.children);
                if (account.children.length === 0) return false;
            } else {
                return false;
            }
        }
        
        // Status filter
        if (filterStatus === 'active' && !account.is_active) return false;
        if (filterStatus === 'inactive' && account.is_active) return false;
        
        // Search filter
        if (state.searchTerm) {
            const term = state.searchTerm.toLowerCase();
            const matches = account.account_code.toLowerCase().includes(term) ||
                          account.name.toLowerCase().includes(term);
            
            if (!matches) {
                // Check children
                if (account.children) {
                    account.children = filterAccounts(account.children);
                    if (account.children.length === 0) return false;
                } else {
                    return false;
                }
            }
        }
        
        return true;
    });
}

/**
 * Apply filter
 */
function applyFilter() {
    renderAccountsTree();
}

/**
 * Handle search
 */
function handleSearch(event) {
    state.searchTerm = event.target.value;
    
    if (state.searchTerm) {
        // Expand all nodes to show search results
        expandAll();
    }
    
    renderAccountsTree();
}

/**
 * Set view mode
 */
function setView(view) {
    state.currentView = view;
    
    if (view === 'tree') {
        document.getElementById('treeViewCard').style.display = 'block';
        document.getElementById('listViewCard').style.display = 'none';
        document.querySelector('[onclick="setView(\'tree\')"]').classList.add('btn-primary');
        document.querySelector('[onclick="setView(\'tree\')"]').classList.remove('btn-secondary');
        document.querySelector('[onclick="setView(\'list\')"]').classList.remove('btn-primary');
        document.querySelector('[onclick="setView(\'list\')"]').classList.add('btn-secondary');
    } else {
        document.getElementById('treeViewCard').style.display = 'none';
        document.getElementById('listViewCard').style.display = 'block';
        document.querySelector('[onclick="setView(\'list\')"]').classList.add('btn-primary');
        document.querySelector('[onclick="setView(\'list\')"]').classList.remove('btn-secondary');
        document.querySelector('[onclick="setView(\'tree\')"]').classList.remove('btn-primary');
        document.querySelector('[onclick="setView(\'tree\')"]').classList.add('btn-secondary');
        renderAccountsList();
    }
}

/**
 * Render accounts list view
 */
function renderAccountsList() {
    const tbody = document.getElementById('accountsTableBody');
    tbody.innerHTML = '';
    
    const allAccounts = flattenAccounts(state.accounts);
    const filteredAccounts = allAccounts.filter(account => {
        const filterType = document.getElementById('filterType').value;
        const filterStatus = document.getElementById('filterStatus').value;
        
        if (filterType && account.type !== filterType) return false;
        if (filterStatus === 'active' && !account.is_active) return false;
        if (filterStatus === 'inactive' && account.is_active) return false;
        
        if (state.searchTerm) {
            const term = state.searchTerm.toLowerCase();
            if (!account.account_code.toLowerCase().includes(term) &&
                !account.name.toLowerCase().includes(term)) {
                return false;
            }
        }
        
        return true;
    });
    
    filteredAccounts.forEach(account => {
        const tr = document.createElement('tr');
        tr.innerHTML = `
            <td>${account.fullPath}</td>
            <td>${account.name}</td>
            <td><span class="account-type ${account.type.toLowerCase()}">${getAccountTypeLabel(account.type)}</span></td>
            <td><span class="account-nature nature-${account.nature.toLowerCase()}">${account.nature}</span></td>
            <td>${account.level}</td>
            <td>${account.is_active ? 
                '<span class="badge badge-success">Activa</span>' : 
                '<span class="badge badge-danger">Inactiva</span>'}</td>
            <td>${account.last_movement ? formatDate(account.last_movement) : '-'}</td>
            <td class="actions-cell">
                <button class="btn btn-sm btn-icon" onclick="editAccountFromList('${account.id}')">✏️</button>
                <button class="btn btn-sm btn-icon" onclick="viewAccountDetails('${account.id}')">👁️</button>
            </td>
        `;
        tbody.appendChild(tr);
    });
}

/**
 * Flatten accounts tree
 */
function flattenAccounts(accounts, parentPath = '') {
    let flat = [];
    accounts.forEach(account => {
        const fullPath = parentPath ? `${parentPath}.${account.account_code}` : account.account_code;
        flat.push({
            ...account,
            fullPath
        });
        if (account.children) {
            flat = flat.concat(flattenAccounts(account.children, fullPath));
        }
    });
    return flat;
}

/**
 * Expand all nodes
 */
function expandAll() {
    const allAccounts = flattenAccounts(state.accounts);
    allAccounts.forEach(account => {
        if (account.children && account.children.length > 0) {
            state.expandedNodes.add(account.id);
        }
    });
    renderAccountsTree();
}

/**
 * Collapse all nodes
 */
function collapseAll() {
    state.expandedNodes.clear();
    renderAccountsTree();
}

/**
 * Update stats
 */
function updateStats() {
    const allAccounts = flattenAccounts(state.accounts);
    
    document.getElementById('totalAccounts').textContent = allAccounts.length;
    document.getElementById('majorAccounts').textContent = allAccounts.filter(a => !a.is_detail).length;
    document.getElementById('detailAccounts').textContent = allAccounts.filter(a => a.is_detail).length;
    document.getElementById('activeAccounts').textContent = allAccounts.filter(a => a.balance && a.balance !== 0).length;
}

/**
 * Show context menu
 */
function showContextMenu(event) {
    const node = event.target.closest('.tree-node');
    const accountId = node.dataset.accountId;
    state.selectedAccount = state.accountsMap[accountId];
    
    const menu = document.getElementById('contextMenu');
    menu.style.display = 'block';
    menu.style.left = event.pageX + 'px';
    menu.style.top = event.pageY + 'px';
}

/**
 * Hide context menu
 */
function hideContextMenu() {
    document.getElementById('contextMenu').style.display = 'none';
}

/**
 * Context menu actions
 */
function editAccount() {
    hideContextMenu();
    if (state.selectedAccount) {
        showAccountModal(state.selectedAccount);
    }
}

function addChildAccount() {
    hideContextMenu();
    showAccountModal();
}

function viewAccountDetails() {
    hideContextMenu();
    if (state.selectedAccount) {
        alert(`Detalles de cuenta: ${state.selectedAccount.account_code} - ${state.selectedAccount.name}`);
    }
}

function viewAccountMovements() {
    hideContextMenu();
    if (state.selectedAccount) {
        window.location.href = `journal_entries.html?account=${state.selectedAccount.account_code}`;
    }
}

function duplicateAccount() {
    hideContextMenu();
    if (state.selectedAccount) {
        const duplicate = { ...state.selectedAccount };
        duplicate.account_code = duplicate.account_code + '_COPIA';
        duplicate.name = duplicate.name + ' (Copia)';
        showAccountModal(duplicate);
    }
}

function deactivateAccount() {
    hideContextMenu();
    if (state.selectedAccount && confirm('¿Desea desactivar esta cuenta?')) {
        state.selectedAccount.is_active = false;
        renderAccountsTree();
        showSuccess('Cuenta desactivada');
    }
}

function deleteAccount() {
    hideContextMenu();
    if (state.selectedAccount && confirm('¿Está seguro de eliminar esta cuenta? Esta acción no se puede deshacer.')) {
        // In real implementation, check for movements first
        showError('No se puede eliminar una cuenta con movimientos');
    }
}

/**
 * Export accounts
 */
function exportAccounts() {
    const allAccounts = flattenAccounts(state.accounts);
    
    // Create CSV content
    const headers = ['Código', 'Nombre', 'Tipo', 'Naturaleza', 'Nivel', 'Es Detalle', 'Estado'];
    const rows = allAccounts.map(account => [
        account.fullPath,
        account.name,
        account.type,
        account.nature,
        account.level,
        account.is_detail ? 'SI' : 'NO',
        account.is_active ? 'Activa' : 'Inactiva'
    ]);
    
    const csv = [headers, ...rows].map(row => row.join(',')).join('\n');
    
    // Download CSV
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `plan_de_cuentas_${new Date().toISOString().split('T')[0]}.csv`;
    a.click();
    
    showSuccess('Plan de cuentas exportado correctamente');
}

/**
 * Show import modal
 */
function showImportModal() {
    document.getElementById('importModal').style.display = 'block';
}

/**
 * Close import modal
 */
function closeImportModal() {
    document.getElementById('importModal').style.display = 'none';
    document.getElementById('fileInput').value = '';
    document.getElementById('importBtn').disabled = true;
}

/**
 * Handle file select
 */
function handleFileSelect(event) {
    const file = event.target.files[0];
    if (file) {
        document.getElementById('importBtn').disabled = false;
        showSuccess(`Archivo seleccionado: ${file.name}`);
    }
}

/**
 * Handle file drop
 */
function handleFileDrop(event) {
    const file = event.dataTransfer.files[0];
    if (file) {
        document.getElementById('fileInput').files = event.dataTransfer.files;
        document.getElementById('importBtn').disabled = false;
        showSuccess(`Archivo seleccionado: ${file.name}`);
    }
}

/**
 * Import accounts
 */
async function importAccounts() {
    const file = document.getElementById('fileInput').files[0];
    if (!file) return;
    
    showLoading();
    
    try {
        // Simulate file processing
        await new Promise(resolve => setTimeout(resolve, 2000));
        
        closeImportModal();
        showSuccess('Plan de cuentas importado correctamente');
        
        // Reload data
        await loadAccountsData();
        
    } catch (error) {
        showError('Error al importar el archivo');
    } finally {
        hideLoading();
    }
}

/**
 * Download template
 */
function downloadTemplate() {
    const template = `Código,Nombre,Tipo,Naturaleza,Es Detalle
1,ACTIVO,ASSET,D,NO
11,ACTIVO CORRIENTE,ASSET,D,NO
1105,CAJA,ASSET,D,NO
1105.05,Caja General,ASSET,D,SI
1105.10,Caja Menor,ASSET,D,SI
1110,BANCOS,ASSET,D,NO
1110.05,Bancos Nacionales,ASSET,D,SI
1110.10,Bancos Extranjeros,ASSET,D,SI`;
    
    const blob = new Blob([template], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'plantilla_plan_cuentas.csv';
    a.click();
}

/**
 * Load template
 */
async function loadTemplate(templateName) {
    if (!confirm('¿Desea cargar esta plantilla? Esto reemplazará el plan de cuentas actual.')) {
        return;
    }
    
    showLoading();
    
    try {
        // Simulate loading template
        await new Promise(resolve => setTimeout(resolve, 1500));
        
        showSuccess(`Plantilla ${templateName} cargada correctamente`);
        
        // Reload with template data
        await loadAccountsData();
        
    } catch (error) {
        showError('Error al cargar la plantilla');
    } finally {
        hideLoading();
    }
}

/**
 * Edit account from list view
 */
function editAccountFromList(accountId) {
    const account = state.accountsMap[accountId];
    if (account) {
        showAccountModal(account);
    }
}

// Helper functions
function getAccountTypeLabel(type) {
    const labels = {
        'ASSET': 'Activo',
        'LIABILITY': 'Pasivo',
        'EQUITY': 'Patrimonio',
        'INCOME': 'Ingreso',
        'EXPENSE': 'Gasto'
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