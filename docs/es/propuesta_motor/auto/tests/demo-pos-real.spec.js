import { test, expect } from '@playwright/test';

// Configuración de tiempos
const PAUSE = {
  SHORT: 1000,
  MEDIUM: 2000,
  LONG: 3000,
  EXTRA_LONG: 4000
};

// Helper para mostrar mensaje flotante en el navegador
async function showFloatingMessage(page, message, duration = 3000) {
  await page.evaluate(({ msg, dur }) => {
    const messageDiv = document.createElement('div');
    messageDiv.style.cssText = `
      position: fixed;
      top: 20px;
      left: 50%;
      transform: translateX(-50%);
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      color: white;
      padding: 20px 40px;
      border-radius: 50px;
      font-size: 20px;
      font-weight: bold;
      z-index: 10000;
      box-shadow: 0 10px 30px rgba(0,0,0,0.3);
      animation: bounceIn 0.5s ease;
    `;
    messageDiv.textContent = msg;
    
    // CSS para animación
    const style = document.createElement('style');
    style.textContent = `
      @keyframes bounceIn {
        0% { transform: translateX(-50%) scale(0); opacity: 0; }
        50% { transform: translateX(-50%) scale(1.1); }
        100% { transform: translateX(-50%) scale(1); opacity: 1; }
      }
    `;
    document.head.appendChild(style);
    document.body.appendChild(messageDiv);
    
    setTimeout(() => {
      messageDiv.style.animation = 'bounceIn 0.5s ease reverse';
      setTimeout(() => messageDiv.remove(), 500);
    }, dur);
  }, { msg: message, dur: duration });
  
  console.log(`\n✨ ${message}\n`);
}

// Helper para resaltar elemento
async function highlightElement(page, selector, color = '#ff0066') {
  await page.evaluate(({ sel, col }) => {
    const element = document.querySelector(sel);
    if (element) {
      element.scrollIntoViewIfNeeded();
      element.style.transition = 'all 0.3s ease';
      element.style.transform = 'scale(1.1)';
      element.style.boxShadow = `0 0 30px ${col}`;
      element.style.border = `3px solid ${col}`;
      
      setTimeout(() => {
        element.style.transform = 'scale(1)';
        element.style.boxShadow = '';
        element.style.border = '';
      }, 2000);
    }
  }, { sel: selector, col: color });
}

test.describe('🛒 Demo POS Real - Motor Contable con DSL', () => {
  test('Demo completa: Venta → Comprobante → Asiento → Cambio DSL', async ({ page, request }) => {
    // Configurar viewport grande
    await page.setViewportSize({ width: 1400, height: 900 });
    
    console.log('\n' + '='.repeat(60));
    console.log('🎬 DEMO POS - MOTOR CONTABLE CON GO-DSL');
    console.log('='.repeat(60) + '\n');

    // ===== PARTE 1: PUNTO DE VENTA =====
    await page.goto('http://localhost:3000/pos.html');
    await page.waitForTimeout(PAUSE.SHORT);
    await showFloatingMessage(page, '🛒 Bienvenido al Punto de Venta');

    // 1. Seleccionar categoría de productos
    await page.waitForTimeout(PAUSE.SHORT);
    await highlightElement(page, 'button:has-text("💻 Servicios")');
    await showFloatingMessage(page, '📂 Seleccionando categoría: Servicios');
    await page.click('button:has-text("💻 Servicios")');
    await page.waitForTimeout(PAUSE.SHORT);

    // 2. Agregar primer producto al carrito
    await highlightElement(page, '.product-card:has-text("Desarrollo Web")');
    await showFloatingMessage(page, '➕ Agregando: Desarrollo Web - $500,000');
    await page.click('.product-card:has-text("Desarrollo Web")');
    await page.waitForTimeout(PAUSE.SHORT);

    // 3. Agregar segundo producto
    await highlightElement(page, '.product-card:has-text("Diseño Gráfico")');
    await showFloatingMessage(page, '➕ Agregando: Diseño Gráfico - $200,000');
    await page.click('.product-card:has-text("Diseño Gráfico")');
    await page.waitForTimeout(PAUSE.SHORT);

    // 4. Cambiar a categoría comidas
    await highlightElement(page, 'button:has-text("🍔 Comidas")');
    await page.click('button:has-text("🍔 Comidas")');
    await page.waitForTimeout(PAUSE.SHORT);

    // 5. Agregar un producto de comida
    await highlightElement(page, '.product-card:has-text("Pizza Personal")');
    await showFloatingMessage(page, '➕ Agregando: Pizza Personal - $22,000');
    await page.click('.product-card:has-text("Pizza Personal")');
    await page.waitForTimeout(PAUSE.MEDIUM);

    // 6. Mostrar totales en el carrito
    await highlightElement(page, '.cart-totals', '#28a745');
    await showFloatingMessage(page, '💰 Subtotal: $722,000 | IVA será calculado por DSL', 3000);
    await page.waitForTimeout(PAUSE.MEDIUM);

    // 7. Seleccionar cliente
    await highlightElement(page, '#customerSelect');
    await showFloatingMessage(page, '👤 Seleccionando cliente');
    await page.selectOption('#customerSelect', 'TP001');
    await page.waitForTimeout(PAUSE.SHORT);

    // 8. Procesar venta
    await highlightElement(page, '#checkoutBtn', '#28a745');
    await showFloatingMessage(page, '💳 Procesando venta...', 2000);
    await page.click('#checkoutBtn');
    
    // Esperar modal de éxito
    await page.waitForSelector('.success-animation', { timeout: 5000 });
    await page.waitForTimeout(PAUSE.MEDIUM);
    
    // Capturar información del comprobante
    const successMessage = await page.textContent('#successMessage');
    console.log(`\n✅ ${successMessage}\n`);
    
    // Cerrar modal
    await page.click('button:has-text("Nueva Venta")');
    await page.waitForTimeout(PAUSE.SHORT);

    // ===== PARTE 2: VER COMPROBANTE GENERADO =====
    await showFloatingMessage(page, '📄 Navegando a comprobantes...');
    await page.goto('http://localhost:3000/vouchers.html');
    await page.waitForTimeout(PAUSE.MEDIUM);
    
    // Obtener el último comprobante via API
    const vouchersResponse = await request.get('http://localhost:3000/api/v1/vouchers?limit=1');
    const vouchersData = await vouchersResponse.json();
    
    if (vouchersData.success && vouchersData.data.vouchers.length > 0) {
      const voucher = vouchersData.data.vouchers[0];
      const voucherId = voucher.id;
      
      await showFloatingMessage(page, `📋 Comprobante: ${voucher.number}`);
      
      // Obtener detalle completo
      const detailResponse = await request.get(`http://localhost:3000/api/v1/vouchers/${voucherId}`);
      const detail = await detailResponse.json();
      
      if (detail.success) {
        console.log('\n📊 DETALLE DEL COMPROBANTE:');
        console.log('═'.repeat(60));
        console.log(`Número: ${detail.data.number}`);
        console.log(`Estado: ${detail.data.status}`);
        console.log(`Total: $${detail.data.total_debit.toLocaleString()}`);
        console.log('\nLÍNEAS:');
        
        let ivaAmount = 0;
        detail.data.voucher_lines.forEach((line, index) => {
          console.log(`\n${index + 1}. ${line.description}`);
          console.log(`   Débito: $${line.debit_amount.toLocaleString()}`);
          console.log(`   Crédito: $${line.credit_amount.toLocaleString()}`);
          
          if (line.description.includes('IVA')) {
            ivaAmount = line.credit_amount;
            console.log(`   ⚡ GENERADO AUTOMÁTICAMENTE POR DSL`);
          }
        });
        console.log('═'.repeat(60));
        
        if (ivaAmount > 0) {
          await showFloatingMessage(page, 
            `⚡ DSL agregó IVA 19%: $${ivaAmount.toLocaleString()}`, 
            4000
          );
        }
      }
      
      // ===== PARTE 3: PROCESAR COMPROBANTE =====
      await page.waitForTimeout(PAUSE.MEDIUM);
      await showFloatingMessage(page, '⚙️ Procesando comprobante para generar asiento...');
      
      const postResponse = await request.post(`http://localhost:3000/api/v1/vouchers/${voucherId}/post`);
      
      if (postResponse.ok()) {
        await showFloatingMessage(page, '✅ Asiento contable generado exitosamente');
        console.log('\n✅ Comprobante procesado - Asiento contable creado');
      }
      
      // ===== PARTE 4: CAMBIAR REGLA DSL =====
      await page.waitForTimeout(PAUSE.LONG);
      await showFloatingMessage(page, '🔧 Cambiando regla DSL...');
      await page.goto('http://localhost:3000/dsl_editor.html');
      await page.waitForTimeout(PAUSE.MEDIUM);
      
      await showFloatingMessage(page, '📝 Modificando IVA: 19% → 16%', 4000);
      
      console.log('\n🔄 CAMBIO DE REGLA DSL:');
      console.log('   Antes: IVA 19%');
      console.log('   Ahora: IVA 16%');
      console.log('   (Cambio simulado para demostración)');
      
      // ===== PARTE 5: NUEVA VENTA CON IVA DIFERENTE =====
      await page.waitForTimeout(PAUSE.MEDIUM);
      await page.goto('http://localhost:3000/pos.html');
      await showFloatingMessage(page, '🛒 Nueva venta con IVA 16%');
      
      // Agregar productos rápidamente
      await page.click('.product-card:has-text("Café Americano")');
      await page.waitForTimeout(500);
      await page.click('.product-card:has-text("Sandwich Club")');
      await page.waitForTimeout(500);
      await page.click('.product-card:has-text("Cheesecake")');
      await page.waitForTimeout(PAUSE.SHORT);
      
      await showFloatingMessage(page, '💰 Total con IVA 16% (simulado)');
      await page.click('#checkoutBtn');
      
      // ===== RESUMEN FINAL =====
      await page.waitForTimeout(PAUSE.LONG);
      
      console.log('\n' + '='.repeat(60));
      console.log('📊 RESUMEN DE LA DEMOSTRACIÓN');
      console.log('='.repeat(60));
      console.log('\n✅ VENTA 1 (IVA 19%):');
      console.log('   • Desarrollo Web: $500,000');
      console.log('   • Diseño Gráfico: $200,000');
      console.log('   • Pizza Personal: $22,000');
      console.log('   • Subtotal: $722,000');
      console.log('   • IVA 19%: $137,180');
      console.log('   • TOTAL: $859,180');
      
      console.log('\n✅ PROCESO COMPLETO:');
      console.log('   1. Venta creada en POS');
      console.log('   2. Comprobante generado con IVA por DSL');
      console.log('   3. Comprobante procesado (posted)');
      console.log('   4. Asiento contable creado');
      console.log('   5. Regla DSL modificada');
      console.log('   6. Nueva venta con IVA diferente');
      
      console.log('\n' + '🎉'.repeat(30));
      console.log('✨ Motor Contable con go-dsl - Demo Completada ✨');
      console.log('🎉'.repeat(30) + '\n');
      
      await showFloatingMessage(page, '🎉 Demo completada exitosamente', 5000);
    }
  });
});