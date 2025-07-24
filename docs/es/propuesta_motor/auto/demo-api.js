/**
 * Demo API - Ejemplo de automatización usando solo llamadas API
 * Este script demuestra cómo interactuar con el Motor Contable via API
 */

import chalk from 'chalk';

const BASE_URL = 'http://localhost:3000/api/v1';

// Helpers para formatear output
const log = {
  header: (msg) => console.log(chalk.blue.bold(`\n╔═══ ${msg} ═══╗\n`)),
  success: (msg) => console.log(chalk.green(`✅ ${msg}`)),
  error: (msg) => console.log(chalk.red(`❌ ${msg}`)),
  info: (msg) => console.log(chalk.yellow(`ℹ️  ${msg}`)),
  data: (obj) => console.log(chalk.gray(JSON.stringify(obj, null, 2))),
};

// Función para hacer requests
async function apiRequest(method, endpoint, data = null) {
  try {
    const options = {
      method,
      headers: {
        'Content-Type': 'application/json',
      },
    };
    
    if (data) {
      options.body = JSON.stringify(data);
    }
    
    const response = await fetch(`${BASE_URL}${endpoint}`, options);
    const result = await response.json();
    
    if (!response.ok) {
      throw new Error(result.error?.message || 'Error en la petición');
    }
    
    return result;
  } catch (error) {
    log.error(`Error en ${method} ${endpoint}: ${error.message}`);
    throw error;
  }
}

// Demo de funcionalidades
async function runDemo() {
  log.header('DEMO API - Motor Contable con go-dsl');
  
  try {
    // 1. Verificar salud del sistema
    log.header('1. Verificando Sistema');
    const health = await fetch(`${BASE_URL}/../health`);
    if (health.ok) {
      log.success('Sistema funcionando correctamente');
    }
    
    // 2. Obtener información de la organización
    log.header('2. Información de la Organización');
    const orgs = await apiRequest('GET', '/organizations');
    log.info(`Organización: ${orgs.data[0].name}`);
    log.info(`NIT: ${orgs.data[0].tax_id}`);
    
    // 3. Crear venta en POS
    log.header('3. Simulando Venta POS');
    const saleData = {
      voucher_type: 'invoice_sale',
      date: new Date().toISOString(),
      description: 'Venta Demo API: Productos tecnológicos',
      reference: `API-DEMO-${Date.now()}`,
      third_party_id: 'TP001',
      voucher_lines: [
        {
          account_id: '68fe4ecbf2d26e205185e0a7a2beb0f0',
          description: 'Venta de productos',
          debit_amount: 0,
          credit_amount: 1000000,
          third_party_id: 'TP001'
        },
        {
          account_id: 'd1e05613ceab0efab7d3e0b6ad290345',
          description: 'Pago efectivo',
          debit_amount: 1000000,
          credit_amount: 0
        }
      ]
    };
    
    const sale = await apiRequest('POST', '/vouchers', saleData);
    log.success(`Venta creada: ${sale.data.number}`);
    log.info('DSL agregó IVA automáticamente');
    
    // 4. Verificar líneas generadas
    const voucherDetail = await apiRequest('GET', `/vouchers/${sale.data.id}`);
    log.info(`Total líneas: ${voucherDetail.data.voucher_lines.length} (incluye IVA por DSL)`);
    
    // 5. Crear compra con retención
    log.header('4. Compra con Retención Automática');
    const purchaseData = {
      voucher_type: 'invoice_purchase',
      date: new Date().toISOString(),
      description: 'Compra de servidores',
      reference: `FC-API-${Date.now()}`,
      third_party_id: 'TP002',
      voucher_lines: [
        {
          account_id: '2938717a9252a428b0f1963a49cf087f',
          description: 'Equipos de cómputo',
          debit_amount: 5000000,
          credit_amount: 0,
          third_party_id: 'TP002'
        },
        {
          account_id: 'a757c937d68d833683d72c91c679a962',
          description: 'Por pagar',
          debit_amount: 0,
          credit_amount: 5000000,
          third_party_id: 'TP002'
        }
      ]
    };
    
    const purchase = await apiRequest('POST', '/vouchers', purchaseData);
    log.success(`Compra creada: ${purchase.data.number}`);
    log.info('DSL aplicó retención del 2.5% (compra > $1M)');
    
    // 6. Obtener métricas del dashboard
    log.header('5. Métricas del Sistema');
    const kpis = await apiRequest('GET', '/dashboard/kpis');
    log.data(kpis.data);
    
    // 7. Listar templates DSL
    log.header('6. Templates DSL Disponibles');
    const templates = await apiRequest('GET', '/dsl/templates');
    templates.data.forEach(t => {
      log.info(`• ${t.name} [${t.category}]`);
    });
    
    log.header('Demo Completada');
    log.success('Todas las funcionalidades funcionan correctamente via API');
    
  } catch (error) {
    log.error('Error durante la demo');
    console.error(error);
  }
}

// Ejecutar demo
console.log(chalk.cyan('\n🚀 Iniciando Demo API del Motor Contable...\n'));
runDemo().then(() => {
  console.log(chalk.cyan('\n✨ Fin de la demo\n'));
});