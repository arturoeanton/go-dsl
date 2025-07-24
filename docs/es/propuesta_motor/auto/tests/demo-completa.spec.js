import { test, expect } from '@playwright/test';

// Configuración de tiempos para que sea visible
const DELAY = {
  TYPING: 150,      // Velocidad de escritura
  CLICK: 500,       // Pausa después de click
  SHORT: 1000,      // Pausa corta
  MEDIUM: 2000,     // Pausa media
  LONG: 3000,       // Pausa larga
  EXTRA: 4000       // Pausa extra larga
};

// Helper para crear banner informativo
async function showBanner(page, title, subtitle = '', type = 'info') {
  const colors = {
    info: '#667eea',
    success: '#48bb78',
    warning: '#ed8936',
    error: '#f56565'
  };
  
  await page.evaluate(({ t, s, c }) => {
    // Remover banner anterior si existe
    const existing = document.getElementById('demo-banner');
    if (existing) existing.remove();
    
    const banner = document.createElement('div');
    banner.id = 'demo-banner';
    banner.style.cssText = `
      position: fixed;
      top: 0;
      left: 0;
      right: 0;
      background: ${c};
      color: white;
      padding: 20px;
      text-align: center;
      z-index: 10000;
      box-shadow: 0 4px 20px rgba(0,0,0,0.3);
      animation: slideDown 0.5s ease;
    `;
    banner.innerHTML = `
      <h2 style="margin: 0; font-size: 24px;">${t}</h2>
      ${s ? `<p style="margin: 5px 0 0 0; opacity: 0.9;">${s}</p>` : ''}
    `;
    
    const style = document.createElement('style');
    style.textContent = `
      @keyframes slideDown {
        from { transform: translateY(-100%); }
        to { transform: translateY(0); }
      }
      @keyframes pulse {
        0%, 100% { transform: scale(1); }
        50% { transform: scale(1.05); }
      }
    `;
    document.head.appendChild(style);
    document.body.appendChild(banner);
    
    setTimeout(() => {
      banner.style.animation = 'slideDown 0.5s ease reverse';
      setTimeout(() => banner.remove(), 500);
    }, 3000);
  }, { t: title, s: subtitle, c: colors[type] });
  
  console.log(`\n${type.toUpperCase()}: ${title} ${subtitle ? '- ' + subtitle : ''}\n`);
}

// Helper para resaltar elemento con efecto visual
async function focusElement(page, selector, color = '#ff0066') {
  await page.evaluate(({ sel, col }) => {
    const element = document.querySelector(sel);
    if (element) {
      element.scrollIntoViewIfNeeded();
      
      // Crear overlay de foco
      const overlay = document.createElement('div');
      overlay.style.cssText = `
        position: fixed;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        pointer-events: none;
        z-index: 9999;
        background: radial-gradient(circle at center, transparent 200px, rgba(0,0,0,0.5) 400px);
        animation: focusIn 0.5s ease;
      `;
      
      // Posicionar el gradiente en el elemento
      const rect = element.getBoundingClientRect();
      const centerX = rect.left + rect.width / 2;
      const centerY = rect.top + rect.height / 2;
      overlay.style.background = `radial-gradient(circle at ${centerX}px ${centerY}px, transparent 100px, rgba(0,0,0,0.7) 300px)`;
      
      document.body.appendChild(overlay);
      
      // Resaltar elemento
      element.style.position = 'relative';
      element.style.zIndex = '10000';
      element.style.transition = 'all 0.3s ease';
      element.style.transform = 'scale(1.1)';
      element.style.boxShadow = `0 0 30px ${col}`;
      element.style.border = `3px solid ${col}`;
      
      // Limpiar después de 2 segundos
      setTimeout(() => {
        overlay.remove();
        element.style.transform = '';
        element.style.boxShadow = '';
        element.style.border = '';
        element.style.zIndex = '';
      }, 2000);
    }
  }, { sel: selector, col: color });
}

// Helper para simular movimiento del mouse
async function moveMouseToElement(page, selector) {
  const element = await page.$(selector);
  if (element) {
    const box = await element.boundingBox();
    if (box) {
      // Mover el mouse lentamente al elemento
      await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2, { steps: 10 });
    }
  }
}

test.describe('🎯 Demo Completa - Motor Contable con DSL', () => {
  test('Flujo completo con todas las interacciones', async ({ page }) => {
    // Configuración inicial
    await page.setViewportSize({ width: 1400, height: 900 });
    page.setDefaultTimeout(30000);
    
    console.log('\n' + '='.repeat(80));
    console.log('🎬 DEMO COMPLETA - MOTOR CONTABLE CON GO-DSL');
    console.log('='.repeat(80) + '\n');

    // ========== PARTE 1: INTRODUCCIÓN ==========
    await page.goto('http://localhost:3000');
    await showBanner(page, '🎉 Bienvenido a la Demo', 'Motor Contable con go-dsl integrado');
    await page.waitForTimeout(DELAY.LONG);

    // ========== PARTE 2: NAVEGACIÓN AL POS ==========
    await page.goto('http://localhost:3000/pos.html');
    await showBanner(page, '🛒 Punto de Venta (POS)', 'Sistema de facturación integrado');
    await page.waitForTimeout(DELAY.MEDIUM);

    // ========== PARTE 3: SELECCIÓN DE PRODUCTOS ==========
    await showBanner(page, '📦 Seleccionando Productos', 'Agregando items al carrito');
    await page.waitForTimeout(DELAY.SHORT);

    // 3.1 Seleccionar categoría Bebidas
    await focusElement(page, 'button:has-text("☕ Bebidas")', '#2196F3');
    await moveMouseToElement(page, 'button:has-text("☕ Bebidas")');
    await page.waitForTimeout(DELAY.CLICK);
    await page.click('button:has-text("☕ Bebidas")');
    await page.waitForTimeout(DELAY.SHORT);

    // 3.2 Agregar Café Americano
    await focusElement(page, '.product-card:has-text("Café Americano")');
    await moveMouseToElement(page, '.product-card:has-text("Café Americano")');
    await page.waitForTimeout(DELAY.CLICK);
    await page.click('.product-card:has-text("Café Americano")');
    await page.waitForTimeout(DELAY.SHORT);

    // 3.3 Agregar otro Café (para mostrar incremento de cantidad)
    await page.click('.product-card:has-text("Café Americano")');
    await showBanner(page, '➕ Cantidad Incrementada', 'Café Americano x2', 'info');
    await page.waitForTimeout(DELAY.SHORT);

    // 3.4 Agregar Cappuccino
    await focusElement(page, '.product-card:has-text("Cappuccino")');
    await page.click('.product-card:has-text("Cappuccino")');
    await page.waitForTimeout(DELAY.SHORT);

    // ========== PARTE 4: CAMBIAR CATEGORÍA Y AGREGAR MÁS ==========
    await showBanner(page, '🍔 Cambiando Categoría', 'Seleccionando comidas');
    
    // 4.1 Seleccionar categoría Comidas
    await focusElement(page, 'button:has-text("🍔 Comidas")', '#FF9800');
    await page.click('button:has-text("🍔 Comidas")');
    await page.waitForTimeout(DELAY.SHORT);

    // 4.2 Agregar Hamburguesa
    await focusElement(page, '.product-card:has-text("Hamburguesa")');
    await page.click('.product-card:has-text("Hamburguesa")');
    await page.waitForTimeout(DELAY.SHORT);

    // 4.3 Agregar Pizza
    await focusElement(page, '.product-card:has-text("Pizza Personal")');
    await page.click('.product-card:has-text("Pizza Personal")');
    await page.waitForTimeout(DELAY.MEDIUM);

    // ========== PARTE 5: INTERACCIONES CON EL CARRITO ==========
    await showBanner(page, '🛒 Modificando Carrito', 'Ajustando cantidades');
    await page.waitForTimeout(DELAY.SHORT);

    // 5.1 Incrementar cantidad de Pizza
    await focusElement(page, '.cart-item:has-text("Pizza") button:has-text("+")', '#4CAF50');
    await page.click('.cart-item:has-text("Pizza") button:has-text("+")');
    await page.waitForTimeout(DELAY.SHORT);

    // 5.2 Decrementar cantidad de Café
    await focusElement(page, '.cart-item:has-text("Café Americano") button:has-text("-")', '#FFC107');
    await page.click('.cart-item:has-text("Café Americano") button:has-text("-")');
    await page.waitForTimeout(DELAY.SHORT);

    // 5.3 Eliminar un producto (Cappuccino)
    await showBanner(page, '🗑️ Eliminando Producto', 'Removiendo Cappuccino del carrito', 'warning');
    await focusElement(page, '.cart-item:has-text("Cappuccino") button:has-text("🗑️")', '#F44336');
    await page.click('.cart-item:has-text("Cappuccino") button:has-text("🗑️")');
    await page.waitForTimeout(DELAY.MEDIUM);

    // ========== PARTE 6: AGREGAR SERVICIOS ==========
    await showBanner(page, '💼 Agregando Servicios', 'Productos de alto valor');
    
    // 6.1 Cambiar a categoría Servicios
    await focusElement(page, 'button:has-text("💼 Servicios")', '#9C27B0');
    await page.click('button:has-text("💼 Servicios")');
    await page.waitForTimeout(DELAY.SHORT);

    // 6.2 Agregar Desarrollo Web
    await focusElement(page, '.product-card:has-text("Desarrollo Web")');
    await page.click('.product-card:has-text("Desarrollo Web")');
    await page.waitForTimeout(DELAY.SHORT);

    // ========== PARTE 7: MOSTRAR TOTALES ==========
    await showBanner(page, '💰 Calculando Totales', 'DSL aplicará IVA automáticamente', 'info');
    await focusElement(page, '.cart-totals', '#4CAF50');
    await page.waitForTimeout(DELAY.LONG);

    // Obtener y mostrar totales
    const subtotal = await page.textContent('#subtotal');
    const iva = await page.textContent('#iva');
    const total = await page.textContent('#total');
    
    console.log('\n💵 TOTALES CALCULADOS:');
    console.log(`   Subtotal: ${subtotal}`);
    console.log(`   IVA (19%): ${iva}`);
    console.log(`   TOTAL: ${total}`);
    console.log('   ⚡ El IVA será recalculado por DSL al procesar\n');

    // ========== PARTE 8: SELECCIÓN DE CLIENTE ==========
    await showBanner(page, '👤 Seleccionando Cliente', 'Asignando cliente a la venta');
    await focusElement(page, '#customerSelect', '#3F51B5');
    await page.selectOption('#customerSelect', 'TP001');
    await page.waitForTimeout(DELAY.MEDIUM);

    // ========== PARTE 9: PROCESAR VENTA ==========
    await showBanner(page, '💳 Procesando Venta', 'Generando comprobante con DSL', 'success');
    await focusElement(page, '#checkoutBtn', '#4CAF50');
    await page.waitForTimeout(DELAY.SHORT);
    
    // Click en procesar
    await page.click('#checkoutBtn');
    
    // Esperar modal de éxito
    await page.waitForSelector('.success-animation', { state: 'visible', timeout: 10000 });
    await showBanner(page, '✅ Venta Exitosa', 'Comprobante generado con IVA por DSL', 'success');
    await page.waitForTimeout(DELAY.LONG);
    
    // Capturar información del modal
    const successText = await page.textContent('#successMessage');
    console.log('\n✅ VENTA PROCESADA:');
    console.log(successText);
    
    // Cerrar modal
    await page.click('button:has-text("Nueva Venta")');
    await page.waitForTimeout(DELAY.MEDIUM);

    // ========== PARTE 10: VER COMPROBANTES ==========
    await showBanner(page, '📄 Navegando a Comprobantes', 'Verificando el comprobante generado');
    await page.goto('http://localhost:3000/vouchers.html');
    await page.waitForTimeout(DELAY.MEDIUM);
    
    // Aquí podrías agregar interacción con la lista de comprobantes
    await focusElement(page, 'tr:first-child', '#2196F3');
    await page.waitForTimeout(DELAY.MEDIUM);

    // ========== PARTE 11: EDITOR DSL ==========
    await showBanner(page, '🔧 Editor de Reglas DSL', 'Modificando reglas de negocio');
    await page.goto('http://localhost:3000/dsl_editor.html');
    await page.waitForTimeout(DELAY.MEDIUM);
    
    // Simular edición de regla
    const editorExists = await page.$('#dsl-code, textarea');
    if (editorExists) {
      await focusElement(page, '#dsl-code, textarea', '#9C27B0');
      await page.waitForTimeout(DELAY.MEDIUM);
    }

    // ========== PARTE 12: DASHBOARD ==========
    await showBanner(page, '📊 Dashboard', 'Métricas y KPIs en tiempo real');
    await page.goto('http://localhost:3000/dashboard.html');
    await page.waitForTimeout(DELAY.MEDIUM);
    
    // Resaltar KPIs
    const kpiCards = await page.$$('.kpi-card, .metric-card');
    for (const card of kpiCards.slice(0, 3)) {
      await card.evaluate(el => {
        el.style.transition = 'all 0.3s ease';
        el.style.transform = 'scale(1.1)';
        setTimeout(() => el.style.transform = '', 300);
      });
      await page.waitForTimeout(300);
    }

    // ========== RESUMEN FINAL ==========
    await page.waitForTimeout(DELAY.LONG);
    await showBanner(page, '🎉 Demo Completada', 'Motor Contable con go-dsl', 'success');
    
    console.log('\n' + '='.repeat(80));
    console.log('📊 RESUMEN DE LA DEMOSTRACIÓN');
    console.log('='.repeat(80));
    console.log('\n✅ FUNCIONALIDADES DEMOSTRADAS:');
    console.log('   1. Navegación completa por el sistema');
    console.log('   2. Selección de productos por categorías');
    console.log('   3. Gestión del carrito (agregar, modificar, eliminar)');
    console.log('   4. Cálculo automático de totales');
    console.log('   5. Selección de cliente');
    console.log('   6. Procesamiento de venta');
    console.log('   7. Generación de comprobante con IVA por DSL');
    console.log('   8. Visualización de comprobantes');
    console.log('   9. Editor de reglas DSL');
    console.log('   10. Dashboard con métricas');
    console.log('\n🚀 El Motor Contable está listo para producción');
    console.log('⚡ Powered by go-dsl\n');
    console.log('='.repeat(80) + '\n');
    
    await page.waitForTimeout(DELAY.EXTRA);
  });
});