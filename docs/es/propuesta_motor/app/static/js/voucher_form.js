// Voucher Form JavaScript - Motor Contable
// Version: 1.0
// Last Updated: 2024-01-15

// Global state for form
const formState = {
    voucherLines: [],
    accounts: [],
    thirdParties: [],
    currentLineId: 0
};

// Initialize on DOM ready
document.addEventListener('DOMContentLoaded', () => {
    initializeVoucherForm();
});

/**
 * Initialize voucher form
 */
function initializeVoucherForm() {
    setupFormEventListeners();
    loadFormData();
    addInitialLines();
}

/**
 * Setup form event listeners
 */
function setupFormEventListeners() {
    // Form submission
    const form = document.getElementById('voucherForm');
    if (form) {
        form.addEventListener('submit', handleFormSubmit);
    }

    // Add line button
    const addLineBtn = document.getElementById('addLineBtn');
    if (addLineBtn) {
        addLineBtn.addEventListener('click', addVoucherLine);
    }

    // Auto-calculate totals on change
    document.addEventListener('input', (e) => {
        if (e.target.classList.contains('debit-amount') || 
            e.target.classList.contains('credit-amount')) {
            calculateTotals();
        }
    });
}

/**
 * Load initial form data
 */
async function loadFormData() {
    try {
        // Load accounts
        const accountsResult = await motorContableApi.accounts.getAll();
        formState.accounts = accountsResult.data.accounts || [];
        
        // Load third parties - simulated for now
        formState.thirdParties = [
            { id: 'TP001', name: 'Cliente Ejemplo S.A.S', type: 'customer' },
            { id: 'TP002', name: 'Proveedor ABC Ltda', type: 'supplier' },
            { id: 'TP003', name: 'Empleado Juan Pérez', type: 'employee' }
        ];
        
        // Set today's date
        document.getElementById('voucherDate').value = new Date().toISOString().split('T')[0];
        
    } catch (error) {
        console.error('Error loading form data:', error);
        utils.toast.error('Error al cargar datos del formulario');
    }
}

/**
 * Add initial voucher lines
 */
function addInitialLines() {
    // Add at least 2 lines initially
    addVoucherLine();
    addVoucherLine();
}

/**
 * Add a new voucher line
 */
function addVoucherLine() {
    const lineId = formState.currentLineId++;
    const linesContainer = document.getElementById('voucherLines');
    
    const lineHtml = `
        <tr data-line-id="${lineId}">
            <td>
                <select class="account-select" name="account_${lineId}" required>
                    <option value="">Seleccionar cuenta...</option>
                    ${formState.accounts.map(acc => 
                        `<option value="${acc.id}">${acc.code} - ${acc.name}</option>`
                    ).join('')}
                </select>
            </td>
            <td>
                <input type="text" class="line-description" name="description_${lineId}" 
                       placeholder="Descripción de la línea" required>
            </td>
            <td>
                <input type="number" class="debit-amount" name="debit_${lineId}" 
                       step="0.01" min="0" value="0" placeholder="0.00">
            </td>
            <td>
                <input type="number" class="credit-amount" name="credit_${lineId}" 
                       step="0.01" min="0" value="0" placeholder="0.00">
            </td>
            <td>
                <select class="third-party-select" name="third_party_${lineId}">
                    <option value="">Sin tercero</option>
                    ${formState.thirdParties.map(tp => 
                        `<option value="${tp.id}">${tp.name}</option>`
                    ).join('')}
                </select>
            </td>
            <td>
                <button type="button" class="btn btn-danger btn-sm" onclick="removeLine(${lineId})">
                    ❌
                </button>
            </td>
        </tr>
    `;
    
    linesContainer.insertAdjacentHTML('beforeend', lineHtml);
    
    // Add line to state
    formState.voucherLines.push({
        id: lineId,
        accountId: '',
        description: '',
        debitAmount: 0,
        creditAmount: 0,
        thirdPartyId: ''
    });
}

/**
 * Remove a voucher line
 */
function removeLine(lineId) {
    // Check minimum lines
    const currentLines = document.querySelectorAll('#voucherLines tr');
    if (currentLines.length <= 2) {
        utils.toast.warning('Debe mantener al menos 2 líneas en el comprobante');
        return;
    }
    
    // Remove from DOM
    const lineRow = document.querySelector(`tr[data-line-id="${lineId}"]`);
    if (lineRow) {
        lineRow.remove();
    }
    
    // Remove from state
    formState.voucherLines = formState.voucherLines.filter(l => l.id !== lineId);
    
    // Recalculate totals
    calculateTotals();
}

/**
 * Calculate and update totals
 */
function calculateTotals() {
    let totalDebit = 0;
    let totalCredit = 0;
    
    // Sum all debit amounts
    document.querySelectorAll('.debit-amount').forEach(input => {
        totalDebit += parseFloat(input.value) || 0;
    });
    
    // Sum all credit amounts
    document.querySelectorAll('.credit-amount').forEach(input => {
        totalCredit += parseFloat(input.value) || 0;
    });
    
    // Update display
    document.getElementById('totalDebit').textContent = formatCurrency(totalDebit);
    document.getElementById('totalCredit').textContent = formatCurrency(totalCredit);
    document.getElementById('difference').textContent = formatCurrency(Math.abs(totalDebit - totalCredit));
    
    // Update balance status
    const balanceStatus = document.getElementById('balanceStatus');
    if (Math.abs(totalDebit - totalCredit) < 0.01) {
        balanceStatus.textContent = '✅ Balanceado';
        balanceStatus.className = 'balance-status balanced';
    } else {
        balanceStatus.textContent = '❌ Desbalanceado';
        balanceStatus.className = 'balance-status unbalanced';
    }
}

/**
 * Handle form submission
 */
async function handleFormSubmit(event) {
    event.preventDefault();
    
    // Validate balance
    const totalDebit = parseFloat(document.getElementById('totalDebit').textContent.replace(/[^0-9.-]+/g, ''));
    const totalCredit = parseFloat(document.getElementById('totalCredit').textContent.replace(/[^0-9.-]+/g, ''));
    
    if (Math.abs(totalDebit - totalCredit) >= 0.01) {
        utils.toast.error('El comprobante debe estar balanceado');
        return;
    }
    
    if (totalDebit === 0) {
        utils.toast.error('El comprobante debe tener al menos un movimiento');
        return;
    }
    
    // Collect form data
    const formData = new FormData(event.target);
    const voucherData = {
        voucher_type: formData.get('voucherType'),
        date: formData.get('voucherDate'),
        description: formData.get('description'),
        reference: formData.get('reference'),
        third_party_id: formData.get('thirdParty') || null,
        voucher_lines: []
    };
    
    // Collect lines
    document.querySelectorAll('#voucherLines tr').forEach(row => {
        const lineId = row.dataset.lineId;
        const accountId = formData.get(`account_${lineId}`);
        const description = formData.get(`description_${lineId}`);
        const debitAmount = parseFloat(formData.get(`debit_${lineId}`)) || 0;
        const creditAmount = parseFloat(formData.get(`credit_${lineId}`)) || 0;
        const thirdPartyId = formData.get(`third_party_${lineId}`) || null;
        
        if (accountId && (debitAmount > 0 || creditAmount > 0)) {
            voucherData.voucher_lines.push({
                account_id: accountId,
                description: description,
                debit_amount: debitAmount,
                credit_amount: creditAmount,
                third_party_id: thirdPartyId
            });
        }
    });
    
    // Submit
    try {
        utils.toast.loading('Creando comprobante...');
        
        const result = await motorContableApi.vouchers.create(voucherData);
        
        utils.toast.success('Comprobante creado exitosamente');
        
        // Redirect to voucher list
        setTimeout(() => {
            window.location.href = 'vouchers_list.html';
        }, 1500);
        
    } catch (error) {
        console.error('Error creating voucher:', error);
        utils.toast.error(error.message || 'Error al crear el comprobante');
    }
}

/**
 * Format currency
 */
function formatCurrency(amount) {
    return new Intl.NumberFormat('es-CO', {
        style: 'currency',
        currency: 'COP',
        minimumFractionDigits: 0,
        maximumFractionDigits: 2
    }).format(amount);
}

// Export functions to global scope
window.addVoucherLine = addVoucherLine;
window.removeLine = removeLine;