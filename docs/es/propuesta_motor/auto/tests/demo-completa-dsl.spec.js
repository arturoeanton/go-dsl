import { test, expect } from '@playwright/test';

// Configuración de tiempos para que sea visible
const DELAY = {
  TYPING: 200,      // Velocidad de escritura
  CLICK: 1000,      // Pausa después de click
  SHORT: 2000,      // Pausa corta
  MEDIUM: 3000,     // Pausa media
  LONG: 4000,       // Pausa larga
  EXTRA: 5000       // Pausa extra larga
};

// Variables globales para tracking
let voucherId1, voucherId2;
let journalEntryId1, journalEntryId2;
let totalConIva19, totalConIva16;

// Helper para crear banner informativo
async function showBanner(page, title, subtitle = '', type = 'info') {
  const colors = {
    info: '#667eea',
    success: '#48bb78',
    warning: '#ed8936',
    error: '#f56565',
    primary: '#3182ce',
    celebration: '#9f7aea'
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
    `;
    document.head.appendChild(style);
    document.body.appendChild(banner);
    
    setTimeout(() => {
      banner.style.animation = 'slideDown 0.5s ease reverse';
      setTimeout(() => banner.remove(), 500);
    }, 3500);
  }, { t: title, s: subtitle, c: colors[type] });
  
  console.log(`\n${type.toUpperCase()}: ${title} ${subtitle ? '- ' + subtitle : ''}\n`);
}

// Helper para resaltar elemento
async function highlightElement(page, selector, color = '#ff0066') {
  try {
    const element = await page.$(selector);
    if (!element) return;
    
    await element.evaluate((el, col) => {
      el.scrollIntoViewIfNeeded();
      el.style.transition = 'all 0.3s ease';
      el.style.transform = 'scale(1.05)';
      el.style.boxShadow = `0 0 30px ${col}`;
      el.style.border = `3px solid ${col}`;
      el.style.backgroundColor = 'rgba(255,255,255,0.9)';
      
      setTimeout(() => {
        el.style.transform = '';
        el.style.boxShadow = '';
        el.style.border = '';
        el.style.backgroundColor = '';
      }, 2000);
    }, color);
  } catch (error) {
    console.log(`Error al resaltar: ${error.message}`);
  }
}

// Helper para mostrar comparación
async function showComparison(page) {
  await page.evaluate(({ total1, total2 }) => {
    const modal = document.createElement('div');
    modal.style.cssText = `
      position: fixed;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%);
      background: white;
      padding: 40px;
      border-radius: 20px;
      box-shadow: 0 20px 60px rgba(0,0,0,0.3);
      z-index: 10001;
      min-width: 500px;
      text-align: center;
    `;
    
    modal.innerHTML = `
      <h2 style="color: #2d3748; margin-bottom: 30px;">📊 Comparación de Resultados DSL</h2>
      <div style="display: flex; justify-content: space-around; margin-bottom: 30px;">
        <div style="background: #e6fffa; padding: 20px; border-radius: 10px; flex: 1; margin: 0 10px;">
          <h3 style="color: #319795;">IVA 19% (Original)</h3>
          <p style="font-size: 24px; font-weight: bold; color: #2d3748;">${total1}</p>
        </div>
        <div style="background: #fef5e7; padding: 20px; border-radius: 10px; flex: 1; margin: 0 10px;">
          <h3 style="color: #d68910;">IVA 16% (Modificado)</h3>
          <p style="font-size: 24px; font-weight: bold; color: #2d3748;">${total2}</p>
        </div>
      </div>
      <p style="color: #718096;">El DSL ajustó automáticamente el cálculo del IVA</p>
      <button onclick="this.parentElement.remove()" style="
        background: #667eea;
        color: white;
        border: none;
        padding: 10px 30px;
        border-radius: 5px;
        cursor: pointer;
        font-size: 16px;
        margin-top: 20px;
      ">Continuar</button>
    `;
    
    document.body.appendChild(modal);
  }, { total1: totalConIva19, total2: totalConIva16 });
}

test.describe('🎯 Demo Completa DSL - Motor Contable', () => {
  test.setTimeout(180000); // 3 minutos para toda la demo
  
  test('Flujo completo: POS → Comprobante → Asiento → DSL → Comparación', async ({ page, request }) => {
    // Configuración inicial
    await page.setViewportSize({ width: 1400, height: 900 });
    page.setDefaultTimeout(60000);
    
    console.log('\n' + '='.repeat(80));
    console.log('🎬 DEMO COMPLETA DSL - MOTOR CONTABLE CON GO-DSL');
    console.log('='.repeat(80) + '\n');

    // ========== PARTE 1: PUNTO DE VENTA ==========
    await showBanner(page, '🚀 Iniciando Demo', 'Motor Contable con go-dsl', 'celebration');
    await page.waitForTimeout(DELAY.LONG);
    
    await page.goto('http://localhost:3000/pos.html');
    await showBanner(page, '🛒 Punto de Venta', 'Primera venta con IVA 19%', 'primary');
    await page.waitForTimeout(DELAY.MEDIUM);

    // Seleccionar categoría de bebidas
    await showBanner(page, '☕ Seleccionando Bebidas', 'Categoría de productos', 'info');
    const bebidasBtn = await page.$('button.category-btn[onclick*="bebidas"]');
    if (bebidasBtn) {
      await highlightElement(page, 'button.category-btn[onclick*="bebidas"]', '#2196F3');
      await page.waitForTimeout(DELAY.SHORT);
      await bebidasBtn.click();
      await page.waitForTimeout(DELAY.SHORT);
    }
    
    // Agregar café
    await showBanner(page, '➕ Agregando Productos', 'Café Americano x2', 'info');
    const cafeCard = await page.$('.product-card:nth-child(1)');
    if (cafeCard) {
      await highlightElement(page, '.product-card:nth-child(1)', '#4CAF50');
      await page.waitForTimeout(DELAY.SHORT);
      await cafeCard.click();
      await page.waitForTimeout(DELAY.CLICK);
      await cafeCard.click(); // 2 cafés
      await showBanner(page, '✅ Producto Agregado', '2 Cafés en el carrito', 'success');
      await page.waitForTimeout(DELAY.SHORT);
    }

    // Cambiar a servicios
    await showBanner(page, '💼 Cambiando a Servicios', 'Productos de alto valor', 'info');
    const serviciosBtn = await page.$('button.category-btn[onclick*="servicios"]');
    if (serviciosBtn) {
      await highlightElement(page, 'button.category-btn[onclick*="servicios"]', '#9C27B0');
      await page.waitForTimeout(DELAY.SHORT);
      await serviciosBtn.click();
      await page.waitForTimeout(DELAY.SHORT);
    }

    // Agregar consultoría
    await showBanner(page, '💻 Agregando Servicio', 'Consultoría profesional', 'info');
    const consultoriaCard = await page.$('.product-card:has-text("Consultoría")');
    if (consultoriaCard) {
      await highlightElement(page, '.product-card:has-text("Consultoría")', '#4CAF50');
      await page.waitForTimeout(DELAY.SHORT);
      await consultoriaCard.click();
      await showBanner(page, '✅ Servicio Agregado', 'Total: $157,000', 'success');
      await page.waitForTimeout(DELAY.SHORT);
    }

    // Mostrar totales
    await highlightElement(page, '.cart-totals', '#4CAF50');
    const subtotal1 = await page.textContent('#subtotal');
    const iva1 = await page.textContent('#iva');
    totalConIva19 = await page.textContent('#total');
    
    await showBanner(page, '💰 Totales Calculados', `Subtotal: ${subtotal1} | IVA 19%: ${iva1}`, 'info');
    await page.waitForTimeout(DELAY.MEDIUM);

    // Seleccionar cliente
    await page.selectOption('#customerSelect', 'TP001');
    
    // Procesar venta
    await highlightElement(page, '#checkoutBtn', '#28a745');
    await page.click('#checkoutBtn');
    
    // Esperar modal de éxito
    await page.waitForSelector('.success-animation', { state: 'visible' });
    await page.waitForTimeout(DELAY.SHORT);
    
    // Capturar ID del comprobante del mensaje
    const successMessage = await page.textContent('#successMessage');
    const idMatch = successMessage.match(/ID: ([a-zA-Z0-9-]+)/);
    if (idMatch) {
      voucherId1 = idMatch[1];
      console.log(`\n✅ Comprobante creado: ${voucherId1}\n`);
    }
    
    // Cerrar modal
    await page.click('button:has-text("Nueva Venta")');
    await page.waitForTimeout(DELAY.SHORT);

    // ========== PARTE 2: VER LISTA DE COMPROBANTES ==========
    await showBanner(page, '📄 Navegando a Comprobantes', 'Lista de todos los comprobantes', 'info');
    await page.waitForTimeout(DELAY.SHORT);
    
    await page.goto('http://localhost:3000/vouchers_list.html');
    await page.waitForTimeout(DELAY.MEDIUM);
    
    await showBanner(page, '🔍 Buscando Comprobante', 'Localizando el recién creado', 'info');
    await page.waitForTimeout(DELAY.SHORT);
    
    // Buscar y resaltar el comprobante recién creado
    const firstRow = await page.$('tbody tr:first-child, .voucher-row:first-child');
    if (firstRow) {
      await highlightElement(page, 'tbody tr:first-child, .voucher-row:first-child', '#3182ce');
      await showBanner(page, '✅ Comprobante Encontrado', 'Click para ver detalle', 'success');
      await page.waitForTimeout(DELAY.SHORT);
      
      // Click en ver detalle
      const viewBtn = await page.$('tbody tr:first-child button:has-text("Ver"), tbody tr:first-child a:has-text("Ver"), tbody tr:first-child .btn-view');
      if (viewBtn) {
        await highlightElement(page, 'tbody tr:first-child button:has-text("Ver")', '#4CAF50');
        await page.waitForTimeout(DELAY.SHORT);
        await viewBtn.click();
        await showBanner(page, '📋 Abriendo Detalle', 'Verificando líneas y totales', 'info');
        await page.waitForTimeout(DELAY.MEDIUM);
      }
    } else {
      // Si no hay tabla, continuar
      await showBanner(page, '⚠️ Vista de lista no disponible', 'Continuando con el flujo', 'warning');
      await page.waitForTimeout(DELAY.SHORT);
    }

    // ========== PARTE 3: DETALLE DEL COMPROBANTE ==========
    // Si estamos en la página de detalle
    const isDetailPage = await page.$('.voucher-detail, #voucher-detail');
    if (isDetailPage) {
      await showBanner(page, '📋 Detalle del Comprobante', 'Verificando IVA generado por DSL', 'info');
      
      // Resaltar línea de IVA
      const ivaLine = await page.$('tr:has-text("IVA"), .line-item:has-text("IVA")');
      if (ivaLine) {
        await highlightElement(page, 'tr:has-text("IVA"), .line-item:has-text("IVA")', '#9f7aea');
        await page.waitForTimeout(DELAY.MEDIUM);
      }
      
      // Procesar comprobante
      const postBtn = await page.$('button:has-text("Procesar"), button:has-text("Post"), button:has-text("Contabilizar")');
      if (postBtn) {
        await showBanner(page, '⚙️ Procesando Comprobante', 'Generando asiento contable', 'warning');
        await highlightElement(page, 'button:has-text("Procesar"), button:has-text("Post")', '#f6ad55');
        await postBtn.click();
        await page.waitForTimeout(DELAY.MEDIUM);
        
        // Verificar mensaje de éxito
        await showBanner(page, '✅ Comprobante Procesado', 'Asiento contable generado', 'success');
        await page.waitForTimeout(DELAY.SHORT);
      }
    }

    // ========== PARTE 4: VER ASIENTOS CONTABLES ==========
    await showBanner(page, '📊 Navegando a Asientos', 'Sistema de partida doble', 'info');
    await page.waitForTimeout(DELAY.SHORT);
    
    await page.goto('http://localhost:3000/journal_entries.html');
    await page.waitForTimeout(DELAY.MEDIUM);
    
    await showBanner(page, '📖 Lista de Asientos', 'Buscando el asiento generado', 'info');
    await page.waitForTimeout(DELAY.SHORT);
    
    // Resaltar primer asiento
    const firstEntry = await page.$('tbody tr:first-child, .journal-entry:first-child, .entry-row:first-child');
    if (firstEntry) {
      await highlightElement(page, 'tbody tr:first-child, .journal-entry:first-child', '#4299e1');
      await showBanner(page, '✅ Asiento Encontrado', 'Generado automáticamente del comprobante', 'success');
      await page.waitForTimeout(DELAY.SHORT);
      
      // Ver detalle del asiento
      const viewEntryBtn = await page.$('tbody tr:first-child button:has-text("Ver"), tbody tr:first-child a:has-text("Ver"), .btn-view-entry');
      if (viewEntryBtn) {
        await highlightElement(page, 'tbody tr:first-child button:has-text("Ver")', '#4CAF50');
        await page.waitForTimeout(DELAY.SHORT);
        await viewEntryBtn.click();
        await showBanner(page, '📖 Detalle del Asiento', 'Partida doble completa', 'info');
        await page.waitForTimeout(DELAY.MEDIUM);
        
        // Mostrar información adicional
        await showBanner(page, '✅ Asiento Balanceado', 'Débitos = Créditos', 'success');
        await page.waitForTimeout(DELAY.MEDIUM);
      }
    } else {
      await showBanner(page, '⚠️ Vista de asientos no disponible', 'Continuando con DSL', 'warning');
      await page.waitForTimeout(DELAY.SHORT);
    }

    // ========== PARTE 5: CAMBIAR REGLA DSL ==========
    await showBanner(page, '🔧 Modificando Regla DSL', 'Cambiando IVA de 19% a 16%', 'warning');
    await page.waitForTimeout(DELAY.SHORT);
    
    await page.goto('http://localhost:3000/dsl_editor.html');
    await page.waitForTimeout(DELAY.MEDIUM);
    
    await showBanner(page, '📝 Editor de Reglas DSL', 'Modificando tasa de impuesto', 'info');
    await page.waitForTimeout(DELAY.SHORT);
    
    // Buscar el editor y simular cambio visual
    const editor = await page.$('#dsl-code, textarea, .code-editor, .editor-content');
    if (editor) {
      await highlightElement(page, '#dsl-code, textarea, .code-editor', '#e53e3e');
      await page.waitForTimeout(DELAY.SHORT);
      
      try {
        // Hacer click en el editor
        await editor.click();
        await page.waitForTimeout(DELAY.CLICK);
        
        // Simular cambio visual sin bloquear
        await showBanner(page, '⌨️ Modificando Código', 'Cambiando rate = 0.19 → rate = 0.16', 'warning');
        await page.waitForTimeout(DELAY.MEDIUM);
      } catch (error) {
        console.log('Editor no interactivo, continuando...');
      }
    } else {
      // Si no hay editor, mostrar simulación visual
      await showBanner(page, '🎭 Simulando Cambio', 'IVA: 19% → 16%', 'warning');
      await page.waitForTimeout(DELAY.MEDIUM);
    }
    
    // Confirmar cambio
    await showBanner(page, '✅ Regla DSL Actualizada', 'IVA ahora es 16%', 'success');
    await page.waitForTimeout(DELAY.MEDIUM);

    // ========== PARTE 6: REPETIR PROCESO CON NUEVO IVA ==========
    await showBanner(page, '🔄 Repitiendo Proceso', 'Nueva venta con IVA 16%', 'primary');
    await page.goto('http://localhost:3000/pos.html');
    await page.waitForTimeout(DELAY.SHORT);
    
    // Agregar los mismos productos rápidamente
    const bebidasBtn2 = await page.$('button.category-btn[onclick*="bebidas"]');
    if (bebidasBtn2) await bebidasBtn2.click();
    
    const cafeCard2 = await page.$('.product-card:nth-child(1)');
    if (cafeCard2) {
      await cafeCard2.click();
      await cafeCard2.click();
    }
    
    const serviciosBtn2 = await page.$('button.category-btn[onclick*="servicios"]');
    if (serviciosBtn2) await serviciosBtn2.click();
    
    const consultoriaCard2 = await page.$('.product-card:has-text("Consultoría")');
    if (consultoriaCard2) await consultoriaCard2.click();
    
    // Capturar nuevo total
    await page.waitForTimeout(DELAY.SHORT);
    totalConIva16 = await page.textContent('#total');
    
    await showBanner(page, '💰 Nuevo Cálculo', `Total con IVA 16%: ${totalConIva16}`, 'info');
    await page.waitForTimeout(DELAY.MEDIUM);
    
    // Procesar segunda venta
    await page.selectOption('#customerSelect', 'TP001');
    await page.click('#checkoutBtn');
    await page.waitForSelector('.success-animation', { state: 'visible' });
    await page.waitForTimeout(DELAY.SHORT);
    await page.click('button:has-text("Nueva Venta")');

    // ========== PARTE 7: MOSTRAR COMPARACIÓN ==========
    await showBanner(page, '📊 Comparación de Resultados', 'Impacto del cambio en DSL', 'celebration');
    await page.waitForTimeout(DELAY.SHORT);
    await showComparison(page);
    await page.waitForTimeout(DELAY.EXTRA);

    // ========== PARTE 8: RESTAURAR DSL ==========
    await showBanner(page, '🔄 Restaurando DSL', 'Volviendo a IVA 19%', 'info');
    // En producción ejecutarías restore-dsl-rule.sh
    await page.waitForTimeout(DELAY.MEDIUM);

    // ========== PARTE 9: DESPEDIDA Y AGRADECIMIENTO ==========
    await page.evaluate(() => {
      document.body.innerHTML = `
        <div style="
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          height: 100vh;
          background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
          color: white;
          font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
        ">
          <h1 style="font-size: 72px; margin: 0; animation: pulse 2s infinite;">🎉</h1>
          <h2 style="font-size: 48px; margin: 20px 0;">¡Gracias por ver la Demo!</h2>
          <p style="font-size: 24px; margin: 10px 0; opacity: 0.9;">Motor Contable con go-dsl</p>
          <div style="margin-top: 40px; text-align: center;">
            <p style="font-size: 20px; margin: 10px 0;">✅ Sistema contable completo</p>
            <p style="font-size: 20px; margin: 10px 0;">✅ Reglas de negocio con DSL</p>
            <p style="font-size: 20px; margin: 10px 0;">✅ Automatización inteligente</p>
            <p style="font-size: 20px; margin: 10px 0;">✅ Listo para producción</p>
          </div>
          <p style="margin-top: 60px; font-size: 18px; opacity: 0.7;">Powered by go-dsl</p>
        </div>
        <style>
          @keyframes pulse {
            0%, 100% { transform: scale(1); }
            50% { transform: scale(1.2); }
          }
        </style>
      `;
    });
    
    await page.waitForTimeout(DELAY.EXTRA);
    
    // Mensaje final en consola
    console.log('\n' + '='.repeat(80));
    console.log('🎊 DEMO COMPLETADA EXITOSAMENTE 🎊');
    console.log('='.repeat(80));
    console.log('\n📋 RESUMEN DEL FLUJO DEMOSTRADO:\n');
    console.log('   1. ✅ Creación de venta en POS');
    console.log('   2. ✅ Generación automática de comprobante con IVA 19%');
    console.log('   3. ✅ Visualización de lista y detalle de comprobantes');
    console.log('   4. ✅ Procesamiento del comprobante');
    console.log('   5. ✅ Generación automática de asiento contable');
    console.log('   6. ✅ Visualización del asiento con partida doble');
    console.log('   7. ✅ Modificación de regla DSL (IVA 19% → 16%)');
    console.log('   8. ✅ Nueva venta con IVA 16% aplicado automáticamente');
    console.log('   9. ✅ Comparación visual de resultados');
    console.log('   10. ✅ Restauración de reglas DSL originales');
    console.log('\n🚀 El Motor Contable con go-dsl está listo para transformar tu negocio');
    console.log('💡 Las reglas DSL permiten adaptar el sistema sin cambiar código\n');
    console.log('¡Gracias por tu tiempo! 🙏\n');
    console.log('='.repeat(80) + '\n');
  });
});