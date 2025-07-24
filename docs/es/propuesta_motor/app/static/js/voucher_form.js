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
    console.log('=== VOUCHER FORM INITIALIZING ===');
    initializeVoucherForm();
});

/**
 * Initialize voucher form
 */
async function initializeVoucherForm() {
    setupFormEventListeners();
    await loadFormData();
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
        const accountsResult = await motorContableApi.accounts.getList();
        formState.accounts = accountsResult.data?.accounts || [];
        console.log('Loaded accounts:', formState.accounts);
        
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
        utils.toast.show('Error al cargar datos del formulario', 'error');
    }
}

/**
 * Add initial voucher lines
 */
function addInitialLines() {
    console.log('Adding initial lines, accounts available:', formState.accounts.length);
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
        date: formData.get('voucherDate') + 'T00:00:00Z', // Convert to ISO format
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
                debit_amount: Math.round(debitAmount * 100) / 100, // Redondear a 2 decimales
                credit_amount: Math.round(creditAmount * 100) / 100, // Redondear a 2 decimales
                third_party_id: thirdPartyId
            });
        }
    });
    
    // Submit
    try {
        // Validate minimum lines
        if (voucherData.voucher_lines.length < 2) {
            utils.toast.show('Se requieren al menos 2 líneas contables', 'error');
            return;
        }
        
        // Validate balance
        const totalDebit = voucherData.voucher_lines.reduce((sum, line) => sum + line.debit_amount, 0);
        const totalCredit = voucherData.voucher_lines.reduce((sum, line) => sum + line.credit_amount, 0);
        
        console.log('Balance check (main form):', { totalDebit, totalCredit, difference: Math.abs(totalDebit - totalCredit) });
        
        if (Math.abs(totalDebit - totalCredit) > 0.01) { // Tolerancia de 1 centavo
            utils.toast.show(`El comprobante no está balanceado. Débitos: ${totalDebit}, Créditos: ${totalCredit}`, 'error');
            return;
        }
        
        // Debug: log data before sending
        console.log('Sending voucher data:', voucherData);
        
        utils.toast.show('Creando comprobante...', 'info');
        
        const result = await motorContableApi.vouchers.create(voucherData);
        
        utils.toast.show('Comprobante creado exitosamente', 'success');
        
        // Redirect to voucher list
        setTimeout(() => {
            window.location.href = 'vouchers_list.html';
        }, 1500);
        
    } catch (error) {
        console.error('Error creating voucher:', error);
        utils.toast.show(error.message || 'Error al crear el comprobante', 'error');
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

/**
 * Save voucher as draft
 */
async function saveDraft() {
    console.log('=== SAVE DRAFT CALLED ===');
    try {
        // Get form data
        const formData = new FormData(document.getElementById('voucherForm'));
        console.log('Form data keys:', Array.from(formData.keys()));
        console.log('Form values:', Object.fromEntries(formData));
        
        // Build voucher object with DRAFT status
        const voucherData = {
            voucher_type: formData.get('voucherType'),
            reference: formData.get('reference'),
            date: formData.get('voucherDate') + 'T00:00:00Z', // Convert to ISO format
            description: formData.get('description'),
            third_party_id: formData.get('thirdParty') || null,
            voucher_lines: []
        };
        
        // Collect lines from DOM (same as main form)
        console.log('Looking for voucher lines...');
        const rows = document.querySelectorAll('#voucherLines tr');
        console.log('Found rows:', rows.length);
        
        rows.forEach(row => {
            const lineId = row.dataset.lineId;
            const accountId = formData.get(`account_${lineId}`);
            const description = formData.get(`description_${lineId}`);
            const debitAmount = parseFloat(formData.get(`debit_${lineId}`)) || 0;
            const creditAmount = parseFloat(formData.get(`credit_${lineId}`)) || 0;
            const thirdPartyId = formData.get(`third_party_${lineId}`) || null;
            
            console.log(`Line ${lineId}:`, {
                accountId, description, debitAmount, creditAmount, thirdPartyId
            });
            
            if (accountId && description && (debitAmount > 0 || creditAmount > 0)) {
                const lineData = {
                    account_id: accountId,
                    description: description || 'Descripción de línea',
                    debit_amount: Math.round(debitAmount * 100) / 100, // Redondear a 2 decimales
                    credit_amount: Math.round(creditAmount * 100) / 100, // Redondear a 2 decimales
                    third_party_id: thirdPartyId
                };
                console.log('Adding line to voucher:', lineData);
                voucherData.voucher_lines.push(lineData);
            } else {
                console.log(`Line ${lineId} REJECTED:`, 'accountId:', !!accountId, 'description:', !!description, 'amounts:', debitAmount, creditAmount);
            }
        });
        
        // Validate minimum lines
        if (voucherData.voucher_lines.length < 2) {
            utils.toast.show('Se requieren al menos 2 líneas contables', 'error');
            return;
        }
        
        // Validate balance
        const totalDebit = voucherData.voucher_lines.reduce((sum, line) => sum + line.debit_amount, 0);
        const totalCredit = voucherData.voucher_lines.reduce((sum, line) => sum + line.credit_amount, 0);
        
        console.log('Balance check:', { totalDebit, totalCredit, difference: Math.abs(totalDebit - totalCredit) });
        
        if (Math.abs(totalDebit - totalCredit) > 0.01) { // Tolerancia de 1 centavo
            utils.toast.show(`El comprobante no está balanceado. Débitos: ${totalDebit}, Créditos: ${totalCredit}`, 'error');
            return;
        }
        
        // Debug: log data before sending
        console.log('Final voucher data (saveDraft):');
        console.log('- voucher_lines:', voucherData.voucher_lines);
        console.log('- voucher_lines length:', voucherData.voucher_lines.length);
        console.log('Complete object:', JSON.stringify(voucherData, null, 2));
        
        // Save draft
        utils.toast.show('Guardando borrador...', 'info');
        
        const result = await motorContableApi.vouchers.create(voucherData);
        
        utils.toast.show('Borrador guardado exitosamente', 'success');
        
        // Redirect to voucher list
        setTimeout(() => {
            window.location.href = 'vouchers_list.html';
        }, 1500);
        
    } catch (error) {
        console.error('Error saving draft:', error);
        utils.toast.show(error.message || 'Error al guardar el borrador', 'error');
    }
}

// Export functions to global scope
window.addVoucherLine = addVoucherLine;
window.removeLine = removeLine;
window.saveDraft = saveDraft;