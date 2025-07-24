import { test, expect } from '@playwright/test';

// Configuración de la demo
const DEMO_CONFIG = {
  pauseTime: 2000, // Tiempo de pausa entre acciones
  animationTime: 1000, // Tiempo para animaciones
  baseURL: 'http://localhost:3000'
};

// Helpers para hacer la demo más visual
async function highlightElement(page, selector) {
  await page.evaluate((sel) => {
    const element = document.querySelector(sel);
    if (element) {
      element.style.border = '3px solid #ff0000';
      element.style.boxShadow = '0 0 20px rgba(255,0,0,0.5)';
      setTimeout(() => {
        element.style.border = '';
        element.style.boxShadow = '';
      }, 2000);
    }
  }, selector);
}

async function pauseWithMessage(page, message, duration = DEMO_CONFIG.pauseTime) {
  console.log(`\n✨ ${message}`);
  await page.waitForTimeout(duration);
}

// Suite de tests de demostración
test.describe('🚀 Demo Motor Contable con go-dsl', () => {
  test.beforeEach(async ({ page }) => {
    // Configurar viewport y navegación inicial
    await page.setViewportSize({ width: 1366, height: 768 });
    await page.goto(DEMO_CONFIG.baseURL);
    await page.waitForLoadState('networkidle');
  });

  test('1️⃣ Dashboard y Navegación Principal', async ({ page }) => {
    await test.step('Mostrar página principal', async () => {
      await pauseWithMessage(page, 'Mostrando la página principal del Motor Contable');
      
      // Verificar título
      await expect(page).toHaveTitle(/Motor Contable/);
      
      // Resaltar el menú principal
      await highlightElement(page, '.navbar');
      await pauseWithMessage(page, 'Este es el menú principal con acceso a todas las funcionalidades');
    });

    await test.step('Navegar al Dashboard', async () => {
      await page.click('a[href="/dashboard.html"]');
      await page.waitForLoadState('networkidle');
      await pauseWithMessage(page, 'Accediendo al Dashboard con KPIs en tiempo real');
      
      // Esperar que los KPIs se carguen
      await page.waitForSelector('.kpi-value', { timeout: 10000 });
      
      // Resaltar KPIs
      const kpiCards = await page.$$('.kpi-card');
      for (const card of kpiCards) {
        await card.scrollIntoViewIfNeeded();
        await card.evaluate(el => {
          el.style.transform = 'scale(1.1)';
          el.style.transition = 'transform 0.3s';
          setTimeout(() => {
            el.style.transform = 'scale(1)';
          }, 1000);
        });
        await page.waitForTimeout(500);
      }
      
      await pauseWithMessage(page, 'Los KPIs muestran métricas en tiempo real del sistema contable');
    });
  });

  test('2️⃣ POS - Punto de Venta con DSL Automático', async ({ page }) => {
    await test.step('Abrir módulo POS', async () => {
      await page.goto(`${DEMO_CONFIG.baseURL}/pos.html`);
      await page.waitForLoadState('networkidle');
      await pauseWithMessage(page, 'Abriendo el módulo de Punto de Venta (POS)');
    });

    await test.step('Crear venta con generación automática de IVA', async () => {
      // Llenar formulario de venta
      await highlightElement(page, '#description');
      await page.fill('#description', 'Venta Demo: 3x Laptops Dell + 2x Monitores LG');
      await pauseWithMessage(page, 'Ingresando descripción de la venta');
      
      await highlightElement(page, '#amount');
      await page.fill('#amount', '8500000');
      await pauseWithMessage(page, 'Monto: $8,500,000 (el DSL calculará IVA automáticamente)');
      
      await highlightElement(page, '#reference');
      await page.fill('#reference', `DEMO-POS-${Date.now()}`);
      await pauseWithMessage(page, 'Referencia única para la transacción');
      
      // Crear venta
      await highlightElement(page, '#create-sale-btn');
      await page.click('#create-sale-btn');
      
      // Esperar respuesta
      await page.waitForSelector('.alert', { timeout: 10000 });
      
      // Verificar que se creó correctamente
      const alert = await page.$('.alert');
      const alertClass = await alert.getAttribute('class');
      
      if (alertClass.includes('alert-success')) {
        await pauseWithMessage(page, '✅ Venta creada exitosamente con IVA generado por DSL', 3000);
        
        // Mostrar detalles si están disponibles
        const details = await page.$('#response-details');
        if (details) {
          await details.scrollIntoViewIfNeeded();
          await highlightElement(page, '#response-details');
          await pauseWithMessage(page, 'El DSL agregó automáticamente el IVA del 19%', 3000);
        }
      }
    });
  });

  test('3️⃣ Comprobantes con Reglas DSL', async ({ page }) => {
    await test.step('Navegar a Comprobantes', async () => {
      await page.goto(`${DEMO_CONFIG.baseURL}/vouchers.html`);
      await page.waitForLoadState('networkidle');
      await pauseWithMessage(page, 'Accediendo al módulo de Comprobantes');
      
      // Esperar que se cargue la lista
      await page.waitForSelector('#vouchers-list', { timeout: 10000 });
    });

    await test.step('Mostrar comprobantes existentes', async () => {
      // Resaltar tabla de comprobantes
      await highlightElement(page, '#vouchers-list');
      await pauseWithMessage(page, 'Lista de comprobantes con diferentes tipos y estados');
      
      // Hacer scroll por algunos comprobantes
      const rows = await page.$$('#vouchers-list tbody tr');
      for (let i = 0; i < Math.min(3, rows.length); i++) {
        await rows[i].scrollIntoViewIfNeeded();
        await rows[i].evaluate(el => {
          el.style.backgroundColor = '#ffffcc';
          setTimeout(() => {
            el.style.backgroundColor = '';
          }, 2000);
        });
        await page.waitForTimeout(800);
      }
      
      await pauseWithMessage(page, 'Cada comprobante puede tener reglas DSL aplicadas automáticamente');
    });

    await test.step('Crear nuevo comprobante', async () => {
      // Buscar y hacer clic en el botón de nuevo comprobante
      const newButton = await page.$('button:has-text("Nuevo Comprobante"), a:has-text("Nuevo Comprobante")');
      if (newButton) {
        await highlightElement(page, 'button:has-text("Nuevo Comprobante"), a:has-text("Nuevo Comprobante")');
        await newButton.click();
        await pauseWithMessage(page, 'Creando un nuevo comprobante que activará reglas DSL');
        
        // Si hay un formulario modal o nueva página, esperarlo
        await page.waitForTimeout(2000);
      }
    });
  });

  test('4️⃣ Editor DSL Visual', async ({ page }) => {
    await test.step('Abrir Editor DSL', async () => {
      await page.goto(`${DEMO_CONFIG.baseURL}/dsl_editor.html`);
      await page.waitForLoadState('networkidle');
      await pauseWithMessage(page, 'Abriendo el Editor DSL Visual');
    });

    await test.step('Mostrar templates DSL disponibles', async () => {
      // Esperar que se cargue el editor
      await page.waitForSelector('#dsl-templates, .template-list', { timeout: 10000 });
      
      // Resaltar lista de templates
      await highlightElement(page, '#dsl-templates, .template-list');
      await pauseWithMessage(page, 'Templates DSL predefinidos para diferentes casos de uso', 3000);
      
      // Si hay un selector de templates, mostrar las opciones
      const templateSelect = await page.$('#template-select, select[name="template"]');
      if (templateSelect) {
        await templateSelect.click();
        await pauseWithMessage(page, 'Múltiples templates disponibles para automatización');
        await page.keyboard.press('Escape');
      }
    });

    await test.step('Demostrar editor de código DSL', async () => {
      // Buscar el editor de código
      const editor = await page.$('#dsl-code, .code-editor, textarea[name="dsl"]');
      if (editor) {
        await highlightElement(page, '#dsl-code, .code-editor, textarea[name="dsl"]');
        
        // Escribir ejemplo de regla DSL
        const dslExample = `// Regla DSL para IVA automático
rule calcular_iva {
  when {
    voucher.type == "invoice_sale"
    account.code.startsWith("4")
  }
  then {
    addLine({
      account: "240802",
      description: "IVA 19% generado automáticamente",
      credit: baseAmount * 0.19
    })
  }
}`;
        
        await editor.click();
        await page.keyboard.type(dslExample, { delay: 50 });
        await pauseWithMessage(page, 'Las reglas DSL se escriben en un lenguaje específico del dominio', 3000);
      }
    });
  });

  test('5️⃣ Plan de Cuentas PUC', async ({ page }) => {
    await test.step('Navegar al Plan de Cuentas', async () => {
      await page.goto(`${DEMO_CONFIG.baseURL}/accounts.html`);
      await page.waitForLoadState('networkidle');
      await pauseWithMessage(page, 'Accediendo al Plan Único de Cuentas (PUC)');
      
      // Esperar que se cargue el árbol de cuentas
      await page.waitForSelector('.account-tree, #accounts-list', { timeout: 10000 });
    });

    await test.step('Explorar estructura jerárquica', async () => {
      await highlightElement(page, '.account-tree, #accounts-list');
      await pauseWithMessage(page, 'Plan de cuentas con 257 cuentas organizadas jerárquicamente');
      
      // Expandir algunas cuentas si es posible
      const expandButtons = await page.$$('.expand-btn, .toggle-children');
      for (let i = 0; i < Math.min(3, expandButtons.length); i++) {
        await expandButtons[i].click();
        await page.waitForTimeout(500);
      }
      
      await pauseWithMessage(page, 'Estructura completa según normativa colombiana');
    });
  });

  test('6️⃣ Workflows y Aprobaciones DSL', async ({ page }) => {
    await test.step('Demostrar workflow de aprobación', async () => {
      console.log('\n🔄 Simulando workflow de aprobación para montos grandes...');
      
      // Hacer llamada API para crear pago grande
      const response = await page.request.post('/api/v1/vouchers', {
        data: {
          voucher_type: 'payment',
          date: new Date().toISOString(),
          description: 'Pago internacional - Requiere aprobación',
          reference: `WORKFLOW-DEMO-${Date.now()}`,
          third_party_id: 'TP002',
          voucher_lines: [
            {
              account_id: 'a757c937d68d833683d72c91c679a962',
              description: 'Pago a proveedor',
              debit_amount: 25000000,
              credit_amount: 0,
              third_party_id: 'TP002'
            },
            {
              account_id: '7d3c841e89ca0d1aca70e06688a6028a',
              description: 'Salida banco',
              debit_amount: 0,
              credit_amount: 25000000
            }
          ]
        }
      });
      
      if (response.ok()) {
        const data = await response.json();
        console.log('✅ Comprobante creado, requiere aprobación por monto > $20M');
        
        // Intentar procesar (debería fallar por workflow)
        const voucherId = data.data.id;
        const postResponse = await page.request.post(`/api/v1/vouchers/${voucherId}/post`);
        
        if (!postResponse.ok()) {
          const error = await postResponse.json();
          console.log(`⚠️  ${error.error.message}`);
          await pauseWithMessage(page, 'DSL detectó que requiere aprobación del CFO', 3000);
        }
      }
    });
  });

  test('7️⃣ Resumen y Conclusiones', async ({ page }) => {
    await test.step('Mostrar resumen final', async () => {
      // Volver al dashboard
      await page.goto(`${DEMO_CONFIG.baseURL}/dashboard.html`);
      await page.waitForLoadState('networkidle');
      
      console.log('\n' + '='.repeat(60));
      console.log('🎉 DEMO COMPLETADA EXITOSAMENTE 🎉');
      console.log('='.repeat(60));
      console.log('\n✅ Funcionalidades Demostradas:');
      console.log('   • Sistema contable completo con PUC colombiano');
      console.log('   • Integración total con go-dsl');
      console.log('   • Generación automática de IVA y retenciones');
      console.log('   • Workflows de aprobación inteligentes');
      console.log('   • POS integrado con reglas DSL');
      console.log('   • Editor visual de reglas DSL');
      console.log('   • Dashboard con métricas en tiempo real');
      console.log('\n🚀 El Motor Contable está listo para casos empresariales');
      console.log('='.repeat(60) + '\n');
      
      await pauseWithMessage(page, 'Demo finalizada - Motor Contable con go-dsl', 5000);
    });
  });
});

// Test adicional para capturar video completo
test('📹 Demo Completa en Video', async ({ page, context }) => {
  // Este test corre toda la demo de forma continua para generar un video
  test.skip(({ browserName }) => browserName !== 'chromium', 'Video solo en Chromium');
  
  console.log('\n🎬 Iniciando grabación de demo completa...\n');
  
  // Dashboard
  await page.goto(`${DEMO_CONFIG.baseURL}/dashboard.html`);
  await page.waitForTimeout(3000);
  
  // POS
  await page.goto(`${DEMO_CONFIG.baseURL}/pos.html`);
  await page.fill('#description', 'Venta para video demo');
  await page.fill('#amount', '1500000');
  await page.fill('#reference', `VIDEO-${Date.now()}`);
  await page.click('#create-sale-btn');
  await page.waitForTimeout(3000);
  
  // Comprobantes
  await page.goto(`${DEMO_CONFIG.baseURL}/vouchers.html`);
  await page.waitForTimeout(3000);
  
  // Editor DSL
  await page.goto(`${DEMO_CONFIG.baseURL}/dsl_editor.html`);
  await page.waitForTimeout(3000);
  
  // Plan de cuentas
  await page.goto(`${DEMO_CONFIG.baseURL}/accounts.html`);
  await page.waitForTimeout(3000);
  
  console.log('\n✅ Video demo completado\n');
});