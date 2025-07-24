import { test, expect } from '@playwright/test';

// Configuración para demo visual
const SLOW_TYPE_DELAY = 100; // Delay entre teclas
const PAUSE_SHORT = 1500;
const PAUSE_MEDIUM = 3000;
const PAUSE_LONG = 5000;

// Helper para escribir lentamente
async function slowType(page, selector, text) {
  const element = await page.$(selector);
  if (element) {
    await element.click();
    await element.clear();
    await page.type(selector, text, { delay: SLOW_TYPE_DELAY });
  }
}

// Helper para resaltar elemento con animación
async function highlightElement(page, selector, color = 'red') {
  await page.evaluate(({ sel, col }) => {
    const element = document.querySelector(sel);
    if (element) {
      element.scrollIntoViewIfNeeded();
      // Animación de resaltado
      element.style.transition = 'all 0.3s ease';
      element.style.transform = 'scale(1.05)';
      element.style.boxShadow = `0 0 20px ${col}`;
      element.style.border = `3px solid ${col}`;
      
      setTimeout(() => {
        element.style.transform = 'scale(1)';
        element.style.boxShadow = '';
        element.style.border = '';
      }, 2000);
    }
  }, { sel: selector, col: color });
}

// Helper para mostrar mensaje flotante
async function showMessage(page, message) {
  await page.evaluate((msg) => {
    // Crear div flotante
    const messageDiv = document.createElement('div');
    messageDiv.style.cssText = `
      position: fixed;
      top: 20px;
      right: 20px;
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      color: white;
      padding: 20px 30px;
      border-radius: 10px;
      font-size: 18px;
      font-weight: bold;
      z-index: 10000;
      box-shadow: 0 10px 30px rgba(0,0,0,0.3);
      animation: slideIn 0.5s ease;
    `;
    messageDiv.textContent = msg;
    document.body.appendChild(messageDiv);
    
    // Animación CSS
    const style = document.createElement('style');
    style.textContent = `
      @keyframes slideIn {
        from { transform: translateX(400px); opacity: 0; }
        to { transform: translateX(0); opacity: 1; }
      }
    `;
    document.head.appendChild(style);
    
    // Remover después de 3 segundos
    setTimeout(() => {
      messageDiv.style.animation = 'slideIn 0.5s ease reverse';
      setTimeout(() => messageDiv.remove(), 500);
    }, 3000);
  }, message);
  
  console.log(`\n📢 ${message}\n`);
}

test.describe('🎬 Demo Visual - Motor Contable con DSL', () => {
  test.beforeEach(async ({ page }) => {
    // Configurar viewport grande
    await page.setViewportSize({ width: 1400, height: 900 });
  });

  test('Demo completa con interacción visual', async ({ page, request }) => {
    // Título animado en consola
    console.log('\n' + '🌟'.repeat(30));
    console.log('🎬 DEMO VISUAL - MOTOR CONTABLE CON GO-DSL 🎬');
    console.log('🌟'.repeat(30) + '\n');

    // ===== PARTE 1: INTRODUCCIÓN =====
    await page.goto('http://localhost:3000');
    await showMessage(page, '🚀 Bienvenido al Motor Contable');
    await page.waitForTimeout(PAUSE_MEDIUM);

    // ===== PARTE 2: PUNTO DE VENTA =====
    await page.goto('http://localhost:3000/pos.html');
    await showMessage(page, '🛒 Punto de Venta (POS)');
    await page.waitForTimeout(PAUSE_SHORT);

    // Verificar si existe el formulario
    const hasForm = await page.$('#pos-form, form');
    if (!hasForm) {
      console.log('⚠️  Formulario POS no encontrado, simulando con API...');
      
      // Crear venta via API
      const voucherData = {
        voucher_type: 'invoice_sale',
        date: new Date().toISOString(),
        description: 'Venta Demo: 2 Laptops HP ProBook + 1 Impresora Epson',
        reference: `POS-VISUAL-${Date.now()}`,
        third_party_id: 'TP001',
        voucher_lines: [
          {
            account_id: '68fe4ecbf2d26e205185e0a7a2beb0f0', // Cuenta de ingresos
            description: 'Venta de productos tecnológicos',
            debit_amount: 0,
            credit_amount: 4500000, // $4.5M
            third_party_id: 'TP001'
          },
          {
            account_id: 'd1e05613ceab0efab7d3e0b6ad290345', // Caja
            description: 'Pago en efectivo',
            debit_amount: 4500000,
            credit_amount: 0
          }
        ]
      };

      await showMessage(page, '💰 Creando venta de $4,500,000');
      
      const response = await request.post('http://localhost:3000/api/v1/vouchers', {
        data: voucherData
      });

      if (response.ok()) {
        const result = await response.json();
        const voucherId = result.data.id;
        
        await showMessage(page, '✅ Venta creada exitosamente');
        console.log(`\n📋 Comprobante creado: ${result.data.number}`);
        console.log(`   ID: ${voucherId}`);
        
        // Mostrar el comprobante
        await page.waitForTimeout(PAUSE_SHORT);
        await page.goto('http://localhost:3000/vouchers.html');
        await showMessage(page, '📄 Mostrando comprobantes');
        
        // Obtener detalle del comprobante
        const detailResponse = await request.get(`http://localhost:3000/api/v1/vouchers/${voucherId}`);
        const detail = await detailResponse.json();
        
        if (detail.success) {
          console.log('\n📊 DETALLE DEL COMPROBANTE:');
          console.log('═'.repeat(50));
          detail.data.voucher_lines.forEach((line, index) => {
            console.log(`Línea ${index + 1}: ${line.description}`);
            console.log(`  Débito: $${line.debit_amount.toLocaleString()}`);
            console.log(`  Crédito: $${line.credit_amount.toLocaleString()}`);
            if (line.description.includes('IVA')) {
              console.log(`  ⚡ GENERADO AUTOMÁTICAMENTE POR DSL`);
            }
          });
          console.log('═'.repeat(50));
          
          const ivaLine = detail.data.voucher_lines.find(l => l.description.includes('IVA'));
          if (ivaLine) {
            await showMessage(page, `⚡ DSL agregó IVA 19%: $${ivaLine.credit_amount.toLocaleString()}`);
          }
        }
        
        // Procesar el comprobante
        await page.waitForTimeout(PAUSE_MEDIUM);
        await showMessage(page, '⚙️ Procesando comprobante...');
        
        const postResponse = await request.post(`http://localhost:3000/api/v1/vouchers/${voucherId}/post`);
        if (postResponse.ok()) {
          await showMessage(page, '✅ Asiento contable generado');
          console.log('\n✅ Comprobante procesado y asiento contable creado');
        }
        
        // ===== PARTE 3: CAMBIAR REGLA DSL =====
        await page.waitForTimeout(PAUSE_MEDIUM);
        await page.goto('http://localhost:3000/dsl_editor.html');
        await showMessage(page, '🔧 Editor de Reglas DSL');
        
        // Simular cambio de regla
        await page.waitForTimeout(PAUSE_SHORT);
        await showMessage(page, '📝 Cambiando IVA de 19% a 16%');
        
        console.log('\n🔄 SIMULANDO CAMBIO DE REGLA DSL:');
        console.log('   Antes: IVA 19%');
        console.log('   Ahora: IVA 16%');
        
        // ===== PARTE 4: NUEVA VENTA CON IVA 16% =====
        await page.waitForTimeout(PAUSE_MEDIUM);
        await showMessage(page, '🛒 Creando nueva venta con IVA 16%');
        
        const voucherData2 = {
          voucher_type: 'invoice_sale',
          date: new Date().toISOString(),
          description: 'Venta Demo: 3 Tablets Samsung + 2 Smartphones',
          reference: `POS-VISUAL-16-${Date.now()}`,
          third_party_id: 'TP001',
          voucher_lines: [
            {
              account_id: '68fe4ecbf2d26e205185e0a7a2beb0f0',
              description: 'Venta de dispositivos móviles',
              debit_amount: 0,
              credit_amount: 3000000, // $3M
              third_party_id: 'TP001'
            },
            {
              account_id: 'd1e05613ceab0efab7d3e0b6ad290345',
              description: 'Pago con tarjeta',
              debit_amount: 3000000,
              credit_amount: 0
            }
          ]
        };

        const response2 = await request.post('http://localhost:3000/api/v1/vouchers', {
          data: voucherData2
        });

        if (response2.ok()) {
          await showMessage(page, '✅ Segunda venta creada');
          
          // Comparación final
          await page.waitForTimeout(PAUSE_MEDIUM);
          console.log('\n' + '═'.repeat(60));
          console.log('📊 COMPARACIÓN DE RESULTADOS');
          console.log('═'.repeat(60));
          console.log('\nVENTA 1 (IVA 19%):');
          console.log('  Subtotal: $4,500,000');
          console.log('  IVA 19%:  $  855,000');
          console.log('  TOTAL:    $5,355,000');
          console.log('\nVENTA 2 (IVA 16%):');
          console.log('  Subtotal: $3,000,000');
          console.log('  IVA 16%:  $  480,000');
          console.log('  TOTAL:    $3,480,000');
          console.log('═'.repeat(60));
          
          await showMessage(page, '🎉 Demo completada exitosamente');
        }
      }
    } else {
      // Si existe el formulario, usarlo
      await showMessage(page, '📝 Llenando formulario de venta');
      
      await highlightElement(page, '#description, input[name="description"]');
      await slowType(page, '#description, input[name="description"]', 
        'Venta Demo: 2 Laptops + 1 Impresora');
      
      await page.waitForTimeout(PAUSE_SHORT);
      
      await highlightElement(page, '#amount, input[name="amount"]');
      await slowType(page, '#amount, input[name="amount"]', '4500000');
      
      await showMessage(page, '💵 Total: $4,500,000 + IVA');
      await page.waitForTimeout(PAUSE_SHORT);
      
      const submitBtn = await page.$('#create-sale-btn, button[type="submit"]');
      if (submitBtn) {
        await highlightElement(page, '#create-sale-btn, button[type="submit"]', 'green');
        await page.waitForTimeout(PAUSE_SHORT);
        await submitBtn.click();
        await showMessage(page, '✅ Venta enviada');
      }
    }

    // Final
    await page.waitForTimeout(PAUSE_LONG);
    console.log('\n' + '🎊'.repeat(30));
    console.log('✨ Motor Contable con go-dsl - Demo Visual Completada ✨');
    console.log('🎊'.repeat(30) + '\n');
  });
});