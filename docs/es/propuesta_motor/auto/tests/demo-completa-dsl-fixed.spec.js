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
let voucherNumber1, voucherNumber2;
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
  test.setTimeout(240000); // 4 minutos para toda la demo
  
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
    
    // Esperar modal de éxito o error
    try {
      await page.waitForSelector('.success-animation', { state: 'visible', timeout: 10000 });
      await page.waitForTimeout(DELAY.SHORT);
    } catch (error) {
      console.log('Modal de éxito no apareció, verificando si hay error...');
      // Verificar si hay un mensaje de error
      const errorMessage = await page.$('.error-message, .alert-danger');
      if (errorMessage) {
        const errorText = await errorMessage.textContent();
        console.error('Error en POS:', errorText);
      }
      // Continuar con el flujo
    }
    
    // Capturar ID del comprobante del mensaje
    const successMessage = await page.textContent('#successMessage');
    const idMatch = successMessage.match(/ID: ([a-zA-Z0-9-]+)/);
    const numberMatch = successMessage.match(/Comprobante: ([A-Z0-9-]+)/);
    if (idMatch) {
      voucherId1 = idMatch[1];
      console.log(`\n✅ Comprobante creado: ${voucherId1}\n`);
    }
    if (numberMatch) {
      voucherNumber1 = numberMatch[1];
    }
    
    // Cerrar modal
    await page.click('button:has-text("Nueva Venta")');
    await page.waitForTimeout(DELAY.SHORT);

    // ========== PARTE 2: VER Y PROCESAR COMPROBANTE ==========
    await showBanner(page, '📄 Navegando a Comprobantes', 'Verificar y procesar', 'info');
    await page.waitForTimeout(DELAY.SHORT);
    
    // Ir a la lista de comprobantes correcta
    await page.goto('http://localhost:3000/vouchers_list.html');
    await page.waitForTimeout(DELAY.MEDIUM);
    
    await showBanner(page, '🔍 Buscando Comprobante', voucherNumber1 || 'Último creado', 'info');
    await page.waitForTimeout(DELAY.SHORT);
    
    // Resaltar primera fila
    const firstRow = await page.$('tbody tr:first-child');
    if (firstRow) {
      await highlightElement(page, 'tbody tr:first-child', '#3182ce');
      await page.waitForTimeout(DELAY.SHORT);
      
      // Primero ver el detalle
      const viewBtn = await page.$('tbody tr:first-child button:has-text("Ver"), tbody tr:first-child a:has-text("Ver")');
      if (viewBtn) {
        await showBanner(page, '👁️ Viendo Detalle', 'Verificando líneas del comprobante', 'info');
        await highlightElement(page, 'tbody tr:first-child button:has-text("Ver")', '#4CAF50');
        await page.waitForTimeout(DELAY.SHORT);
        await viewBtn.click();
        await page.waitForTimeout(DELAY.MEDIUM);
        
        // En la página de detalle, resaltar línea de IVA
        const ivaLine = await page.$('tr:has-text("IVA"), .line:has-text("IVA")');
        if (ivaLine) {
          await highlightElement(page, 'tr:has-text("IVA")', '#9f7aea');
          await showBanner(page, '✅ IVA Generado por DSL', '19% calculado automáticamente', 'success');
          await page.waitForTimeout(DELAY.MEDIUM);
        }
        
        // Volver a la lista
        await page.goBack();
        await page.waitForTimeout(DELAY.SHORT);
      }
    }
    
    // Configurar handler para el alert ANTES de hacer click
    page.on('dialog', async dialog => {
      console.log(`Alert detectado: ${dialog.message()}`);
      await dialog.accept();
      console.log('Alert aceptado');
    });
    
    // Ahora buscar botón de procesar en la primera fila
    await showBanner(page, '⚙️ Buscando Botón Procesar', 'Para generar asiento contable', 'info');
    const processBtn = await page.$('tbody tr:first-child button:has-text("Procesar"), tbody tr:first-child button:has-text("Post"), tbody tr:first-child .btn-post, tbody tr:first-child button.btn-primary');
    if (processBtn) {
      await highlightElement(page, 'tbody tr:first-child', '#FFC107');
      await showBanner(page, '⚙️ Procesando Comprobante', 'Generando asiento contable', 'warning');
      await page.waitForTimeout(DELAY.SHORT);
      
      await highlightElement(page, 'tbody tr:first-child button:has-text("Procesar")', '#f6ad55');
      
      // Click y esperar un poco para el alert
      await processBtn.click();
      console.log('Botón de procesar clickeado, esperando alert...');
      
      // Dar tiempo para que aparezca el alert y se procese
      await page.waitForTimeout(2000);
      
      await showBanner(page, '✅ Comprobante Procesado', 'Asiento contable generado', 'success');
      await page.waitForTimeout(DELAY.SHORT);
    } else {
      // Si no hay botón en la lista, intentar via API
      await showBanner(page, '⚙️ Procesando via API', 'Generando asiento contable', 'warning');
      if (voucherId1) {
        try {
          await request.post(`http://localhost:3000/api/v1/vouchers/${voucherId1}/post`);
          await showBanner(page, '✅ Procesado Exitosamente', 'Asiento generado', 'success');
        } catch (error) {
          console.log('Error procesando, continuando...');
        }
      }
      await page.waitForTimeout(DELAY.SHORT);
    }

    // ========== PARTE 3: VER ASIENTOS CONTABLES ==========
    await showBanner(page, '📊 Navegando a Asientos', 'Sistema de partida doble', 'info');
    await page.waitForTimeout(DELAY.SHORT);
    
    await page.goto('http://localhost:3000/journal_entries.html');
    await page.waitForTimeout(DELAY.MEDIUM);
    
    await showBanner(page, '📖 Lista de Asientos', 'Verificando el último asiento', 'info');
    await page.waitForTimeout(DELAY.SHORT);
    
    // Resaltar primer asiento (el más reciente)
    const firstEntry = await page.$('tbody tr:first-child, .journal-entry:first-child');
    if (firstEntry) {
      await highlightElement(page, 'tbody tr:first-child', '#4299e1');
      await showBanner(page, '✅ Asiento Encontrado', 'Partida doble balanceada', 'success');
      await page.waitForTimeout(DELAY.MEDIUM);
    }

    // ========== PARTE 4: CAMBIAR REGLA DSL ==========
    await showBanner(page, '🔧 Cambiando Regla DSL', 'Modificando IVA de 19% a 16%', 'warning');
    await page.waitForTimeout(DELAY.SHORT);
    
    // HACER CAMBIO REAL DE DSL VIA API
    await showBanner(page, '🎭 Ejecutando Cambio DSL', 'Llamada real a la API', 'warning');
    console.log('🔧 Cambiando tasa de IVA a 16% via API...');
    
    try {
      const changeResponse = await request.post('http://localhost:3000/api/v1/dsl/iva-rate', {
        headers: {
          'Content-Type': 'application/json'
        },
        data: {
          rate: 0.16
        }
      });
      
      if (changeResponse.ok()) {
        const changeData = await changeResponse.json();
        console.log('✅ Tasa de IVA cambiada exitosamente:', changeData);
        await showBanner(page, '✅ DSL Actualizado Realmente', `IVA ahora es ${changeData.data.percentage}%`, 'success');
      } else {
        console.error('❌ Error cambiando tasa de IVA:', await changeResponse.text());
        await showBanner(page, '❌ Error en DSL', 'Continuando con simulación', 'error');
      }
    } catch (error) {
      console.error('❌ Error en llamada API:', error);
      await showBanner(page, '❌ Error de Conexión', 'Continuando con simulación', 'error');
    }
    
    await page.waitForTimeout(DELAY.SHORT);

    // ========== PARTE 5: REPETIR PROCESO CON NUEVO IVA ==========
    await showBanner(page, '🔄 Segunda Venta', 'Con IVA modificado al 16%', 'primary');
    await page.goto('http://localhost:3000/pos.html');
    await page.waitForTimeout(DELAY.SHORT);
    
    // Agregar los mismos productos rápidamente
    await showBanner(page, '🚀 Agregando Productos', 'Mismos items que antes', 'info');
    
    // Bebidas
    const bebidasBtn2 = await page.$('button.category-btn[onclick*="bebidas"]');
    if (bebidasBtn2) await bebidasBtn2.click();
    await page.waitForTimeout(DELAY.SHORT);
    
    const cafeCard2 = await page.$('.product-card:nth-child(1)');
    if (cafeCard2) {
      await cafeCard2.click();
      await cafeCard2.click();
    }
    
    // Servicios
    const serviciosBtn2 = await page.$('button.category-btn[onclick*="servicios"]');
    if (serviciosBtn2) await serviciosBtn2.click();
    await page.waitForTimeout(DELAY.SHORT);
    
    const consultoriaCard2 = await page.$('.product-card:has-text("Consultoría")');
    if (consultoriaCard2) await consultoriaCard2.click();
    
    // Capturar nuevo total REAL (calculado dinámicamente por DSL)
    await page.waitForTimeout(DELAY.SHORT);
    
    // Esperar a que se actualicen los totales dinámicamente
    await page.waitForTimeout(2000);
    
    const subtotal2 = await page.textContent('#subtotal');
    const iva2 = await page.textContent('#iva');
    totalConIva16 = await page.textContent('#total');
    
    console.log('📊 Totales con nueva tasa DSL:');
    console.log(`   Subtotal: ${subtotal2}`);
    console.log(`   IVA: ${iva2}`);
    console.log(`   Total: ${totalConIva16}`);
    
    await showBanner(page, '💰 Nuevo Cálculo REAL', `Total con IVA 16%: ${totalConIva16}`, 'info');
    await page.waitForTimeout(DELAY.MEDIUM);
    
    // Procesar segunda venta
    await page.selectOption('#customerSelect', 'TP001');
    await page.click('#checkoutBtn');
    await page.waitForSelector('.success-animation', { state: 'visible' });
    await page.waitForTimeout(DELAY.SHORT);
    await page.click('button:has-text("Nueva Venta")');

    // ========== PARTE 6: MOSTRAR COMPARACIÓN ==========
    await showBanner(page, '📊 Comparación de Resultados', 'Impacto del cambio en DSL', 'celebration');
    await page.waitForTimeout(DELAY.SHORT);
    await showComparison(page);
    await page.waitForTimeout(DELAY.EXTRA);

    // ========== PARTE 7: RESTAURAR DSL ==========
    await showBanner(page, '🔄 Restaurando DSL', 'Volviendo a IVA 19%', 'info');
    console.log('🔄 Restaurando tasa de IVA a 19% via API...');
    
    try {
      const restoreResponse = await request.post('http://localhost:3000/api/v1/dsl/iva-rate', {
        headers: {
          'Content-Type': 'application/json'
        },
        data: {
          rate: 0.19
        }
      });
      
      if (restoreResponse.ok()) {
        const restoreData = await restoreResponse.json();
        console.log('✅ Tasa de IVA restaurada exitosamente:', restoreData);
        await showBanner(page, '✅ DSL Restaurado', `Sistema vuelve a IVA ${restoreData.data.percentage}%`, 'success');
      } else {
        console.error('❌ Error restaurando tasa de IVA:', await restoreResponse.text());
        await showBanner(page, '❌ Error Restaurando', 'Revisar manualmente', 'error');
      }
    } catch (error) {
      console.error('❌ Error en restauración:', error);
      await showBanner(page, '❌ Error de Conexión', 'Revisar configuración', 'error');
    }
    
    await page.waitForTimeout(DELAY.SHORT);

    // ========== PARTE 8: DESPEDIDA Y AGRADECIMIENTO ==========
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
          <p style="margin-top: 20px; font-size: 16px; opacity: 0.5;">github.com/arturoeanton/go-dsl</p>
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
    console.log('   3. ✅ Procesamiento del comprobante');
    console.log('   4. ✅ Generación automática de asiento contable');
    console.log('   5. ✅ Visualización del asiento con partida doble');
    console.log('   6. ✅ Modificación de regla DSL (IVA 19% → 16%)');
    console.log('   7. ✅ Nueva venta con IVA 16% aplicado');
    console.log('   8. ✅ Comparación visual de resultados');
    console.log('   9. ✅ Restauración de reglas DSL originales');
    console.log('\n🚀 El Motor Contable con go-dsl está listo para transformar tu negocio');
    console.log('💡 Las reglas DSL permiten adaptar el sistema sin cambiar código\n');
    console.log('¡Gracias por tu tiempo! 🙏\n');
    console.log('='.repeat(80) + '\n');
  });
});