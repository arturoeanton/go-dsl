import { test, expect } from '@playwright/test';

test.describe('Demo Simple - Motor Contable', () => {
  test('Demo básica de navegación', async ({ page }) => {
    console.log('\n🚀 Iniciando demo del Motor Contable...\n');
    
    // 1. Página principal
    await page.goto('http://localhost:3000');
    await page.waitForTimeout(2000);
    console.log('✅ Página principal cargada');
    
    // 2. Intentar acceder al POS
    await page.goto('http://localhost:3000/pos.html');
    await page.waitForTimeout(2000);
    console.log('✅ POS cargado');
    
    // Verificar si hay formulario de POS
    const hasForm = await page.$('#pos-form, form');
    if (hasForm) {
      console.log('✅ Formulario de POS encontrado');
    }
    
    // 3. Dashboard
    await page.goto('http://localhost:3000/dashboard.html');
    await page.waitForTimeout(2000);
    console.log('✅ Dashboard cargado');
    
    // 4. Comprobantes
    await page.goto('http://localhost:3000/vouchers.html');
    await page.waitForTimeout(2000);
    console.log('✅ Página de comprobantes cargada');
    
    // 5. Editor DSL
    await page.goto('http://localhost:3000/dsl_editor.html');
    await page.waitForTimeout(2000);
    console.log('✅ Editor DSL cargado');
    
    console.log('\n🎉 Demo completada exitosamente!\n');
  });

  test('Crear venta en POS', async ({ page }) => {
    console.log('\n🛒 Probando POS...\n');
    
    await page.goto('http://localhost:3000/pos.html');
    await page.waitForTimeout(1000);
    
    // Buscar campos del formulario
    const descField = await page.$('input[name="description"], #description, input[type="text"]').catch(() => null);
    const amountField = await page.$('input[name="amount"], #amount, input[type="number"]').catch(() => null);
    const submitButton = await page.$('button[type="submit"], #create-sale-btn, button').catch(() => null);
    
    if (descField && amountField && submitButton) {
      console.log('✅ Formulario POS encontrado, llenando datos...');
      
      await descField.fill('Venta de prueba desde Playwright');
      await amountField.fill('100000');
      
      console.log('💾 Enviando venta...');
      await submitButton.click();
      await page.waitForTimeout(2000);
      
      console.log('✅ Venta enviada');
    } else {
      console.log('⚠️  No se encontraron todos los campos del formulario');
    }
  });

  test('Verificar API', async ({ request }) => {
    console.log('\n🔌 Verificando endpoints API...\n');
    
    // Verificar organizaciones
    try {
      const orgs = await request.get('http://localhost:3000/api/v1/organizations');
      if (orgs.ok()) {
        console.log('✅ Endpoint /organizations funciona');
        const data = await orgs.json();
        if (data.data && data.data.length > 0) {
          console.log(`   Organización: ${data.data[0].name}`);
        }
      }
    } catch (e) {
      console.log('❌ Error en /organizations');
    }
    
    // Verificar dashboard KPIs
    try {
      const kpis = await request.get('http://localhost:3000/api/v1/dashboard/kpis');
      if (kpis.ok()) {
        console.log('✅ Endpoint /dashboard/kpis funciona');
      }
    } catch (e) {
      console.log('❌ Error en /dashboard/kpis');
    }
    
    // Verificar templates DSL
    try {
      const templates = await request.get('http://localhost:3000/api/v1/dsl/templates');
      if (templates.ok()) {
        console.log('✅ Endpoint /dsl/templates funciona');
        const data = await templates.json();
        if (data.data && data.data.length > 0) {
          console.log(`   ${data.data.length} templates encontrados`);
        }
      }
    } catch (e) {
      console.log('❌ Error en /dsl/templates');
    }
  });
});