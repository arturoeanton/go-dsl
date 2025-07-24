# Documentación de Mocks - Motor Contable

## Estado Actual: ✅ Mock Funcional Completo

Este documento describe la implementación completa del sistema de mocks funcionales para el Motor Contable, incluyendo interfaz web responsive, API centralizada y datos de ejemplo.

---

## 🏗️ Arquitectura del Mock

### Sistema de API Centralizado

```javascript
// api_service.js - Configuración
const API_CONFIG = {
    USE_MOCK: true,      // Cambiar a false para usar API real
    BASE_URL: '/api/v1', // URL del backend real
    MOCK_BASE: '../../api',
    TIMEOUT: 30000
};
```

### Catálogo de APIs (catalog_api.json)

Define 40+ endpoints organizados en módulos:
- Dashboard & Analytics
- Vouchers Management 
- Journal Entries
- Accounts & Chart
- Reports Generation
- DSL Templates
- Organizations
- Users & Auth
- Lookups & Catalogs

---

## 📊 Páginas Implementadas

### 1. Dashboard Principal
**Archivo**: `dashboard.html`

**Características**:
- 4 KPIs principales con cambios porcentuales
- 3 gráficos interactivos (Chart.js):
  - Líneas: Procesamiento diario (7 días)
  - Dona: Distribución por tipo de comprobante
  - Barras: Estados de procesamiento
- Feed de actividad en tiempo real
- Métricas de rendimiento del sistema

**APIs utilizadas**:
- `motorContableApi.dashboard.getStats()`

### 2. Gestión de Comprobantes

#### Lista de Comprobantes
**Archivo**: `vouchers_list.html`

**Funcionalidades**:
- Tabla con 8 columnas de información
- Filtros por: tipo, estado, fecha, organización
- Búsqueda en tiempo real con debounce
- Paginación configurable (10/25/50/100)
- Acciones en lote (aprobar, rechazar, eliminar)
- Vista rápida en modal
- Exportación a Excel/CSV
- Estados con badges de colores

#### Formulario de Comprobantes
**Archivo**: `vouchers_form.html`

**Campos**:
- Información básica (tipo, número, fecha, moneda)
- Búsqueda de terceros con autocompletado
- Líneas de items dinámicas
- Cálculo automático de:
  - Subtotales
  - Descuentos
  - Impuestos (IVA 5%, 19%)
  - Retenciones
  - Total
- Guardar como borrador
- Procesar y contabilizar

### 3. Asientos Contables
**Archivo**: `journal_entries.html`

**Características**:
- Lista con filtros avanzados
- Creación manual de asientos
- Validación de balance (débito = crédito)
- Líneas múltiples con cuentas
- Reversión de asientos
- Vista previa del impacto
- Adjuntar documentos soporte

### 4. Plan de Cuentas
**Archivo**: `accounts_chart.html`

**Visualización**:
- Árbol expandible hasta 5 niveles
- 50+ cuentas de ejemplo
- Indicadores visuales:
  - Tipo de cuenta (color)
  - Naturaleza (D/C)
  - Estado (activa/inactiva)
- Búsqueda por código o nombre
- Filtros por tipo de cuenta
- Acciones CRUD completas

### 5. Generador de Reportes
**Archivo**: `reports.html`

**Tipos de reportes**:
1. Balance General
2. Estado de Resultados
3. Libro Diario
4. Libro Mayor
5. Balance de Comprobación
6. Auxiliar de Terceros
7. Flujo de Efectivo
8. Estado de Cambios en el Patrimonio
9. Notas a los Estados Financieros
10. Reportes Fiscales
11. Indicadores Financieros
12. Reporte Personalizado (DSL)

**Funcionalidades**:
- Configuración de parámetros
- Comparación entre períodos
- Vista previa en línea
- Exportación múltiple (PDF, Excel, CSV)
- Programación de reportes
- Envío por email
- Historial de generados

### 6. Editor DSL
**Archivo**: `dsl_editor.html`

**Editor de código**:
- Sintaxis sin resaltado (estabilidad)
- Números de línea
- Auto-indentación
- 5 plantillas de ejemplo:
  - invoice_sale_co
  - invoice_purchase_co
  - payment_bank_transfer
  - receipt_cash
  - payroll_monthly

**Herramientas**:
- Validación de sintaxis
- Formateo de código
- Inserción de snippets
- Variables disponibles (modal)
- Funciones del DSL (referencia)
- Test con datos de prueba
- Historial de versiones

---

## 🔌 Mocks de API Implementados

### 1. Dashboard
**Archivo**: `dashboard.json`

```json
{
  "kpis": {
    "vouchers_today": 127,
    "vouchers_month": 3421,
    "total_amount_month": 458750000.50,
    "pending_vouchers": 23,
    "processing_rate": 98.5,
    "average_processing_time": 1.2
  },
  "charts": {
    "vouchers_by_day": { "labels": [...], "values": [...] },
    "vouchers_by_type": { "labels": [...], "values": [...] },
    "vouchers_by_status": { "labels": [...], "values": [...] }
  }
}
```

### 2. Gestión de Datos

#### vouchers_list.json
- 20 comprobantes de ejemplo
- Diferentes tipos y estados
- Paginación simulada
- Estadísticas de resumen

#### voucher_types.json
- 8 tipos de comprobantes
- Configuración específica por tipo
- Iconos y colores distintivos

#### journal_entries.json
- 15 asientos de ejemplo
- Balance cuadrado
- Múltiples líneas por asiento

#### accounts_tree.json
- Plan de cuentas PUC Colombia
- Estructura jerárquica
- 9 tipos de cuenta principales

### 3. Catálogos

#### countries.json
- 6 países: CO, MX, CL, PE, EC, UY
- Formato de documento fiscal
- Moneda por defecto

#### currencies.json
- 7 monedas latinoamericanas
- Formato de números
- Símbolo y decimales

#### organizations.json
- 3 empresas multi-país
- Configuración fiscal
- Estándares contables

### 4. Configuración

#### user_profile.json
- Perfil completo
- Permisos granulares
- Preferencias de UI

#### dsl_templates.json
- 5 plantillas funcionales
- Versionado
- Tags y categorización

#### report_types.json
- 10 tipos de reportes
- Parámetros requeridos
- Categorías

---

## 🎨 Características de UI/UX

### Sistema de Diseño

**Variables CSS**:
```css
:root {
  --primary-color: #1890ff;
  --success-color: #52c41a;
  --warning-color: #faad14;
  --danger-color: #f5222d;
  --info-color: #1890ff;
  --spacing-sm: 8px;
  --spacing-md: 16px;
  --spacing-lg: 24px;
  --radius-md: 4px;
  --shadow-sm: 0 1px 2px rgba(0,0,0,0.05);
}
```

### Componentes Reutilizables

**Cards**:
- Header con título y acciones
- Body con padding consistente
- Footer opcional

**Tablas**:
- Headers sticky
- Ordenamiento por columnas
- Selección múltiple
- Acciones por fila

**Formularios**:
- Labels descriptivos
- Placeholders de ayuda
- Validación en tiempo real
- Mensajes de error

**Modales**:
- Overlay oscuro
- Animación de entrada
- Botón de cierre
- Footer con acciones

### Navegación

**Sidebar**:
- Colapsable (solo iconos)
- Estado persistente
- Indicador de página activa
- Iconos descriptivos

**Navbar**:
- Logo y nombre
- Búsqueda global
- Selector de organización
- Menú de usuario

### Responsive Design

**Breakpoints**:
- Mobile: < 768px
- Tablet: 768px - 1024px
- Desktop: > 1024px

**Adaptaciones**:
- Sidebar se oculta en mobile
- Tablas con scroll horizontal
- Grids adaptativos
- Modales fullscreen en mobile

---

## 🚀 Migración a Backend Real

### Paso 1: Configuración
```javascript
// En api_service.js
API_CONFIG.USE_MOCK = false;
API_CONFIG.BASE_URL = 'https://api.motorcontable.com/v1';
```

### Paso 2: Autenticación
```javascript
getAuthToken() {
    return localStorage.getItem('auth_token');
}
```

### Paso 3: Endpoints
Todos los endpoints ya están mapeados en `catalog_api.json`

### Paso 4: Manejo de Errores
El servicio ya maneja errores y muestra notificaciones

---

## 📋 Checklist de Funcionalidades

### ✅ Completadas
- [x] Sistema de navegación completo
- [x] Dashboard con gráficos interactivos
- [x] CRUD de comprobantes
- [x] Gestión de asientos contables
- [x] Plan de cuentas jerárquico
- [x] Generador de reportes
- [x] Editor DSL funcional
- [x] API service centralizado
- [x] Sistema de mocks completo
- [x] Diseño responsive
- [x] Componentes reutilizables

### 🚧 Próximas Fases
- [ ] Autenticación y login
- [ ] WebSockets para tiempo real
- [ ] Notificaciones push
- [ ] Modo offline (PWA)
- [ ] Tests E2E automatizados

---

## 🛠️ Herramientas Utilizadas

- **HTML5**: Estructura semántica
- **CSS3**: Variables CSS, Flexbox, Grid
- **JavaScript ES6+**: Async/await, clases, módulos
- **Chart.js**: Gráficos interactivos
- **LocalStorage**: Persistencia de preferencias

---

## 📂 Estructura de Archivos

```
mocks/
├── api/
│   ├── catalog_api.json         # Catálogo de endpoints
│   ├── dashboard.json           # Datos del dashboard
│   ├── vouchers_list.json       # Lista de comprobantes
│   ├── voucher_types.json       # Tipos de comprobantes
│   ├── journal_entries.json     # Asientos contables
│   ├── accounts_tree.json       # Plan de cuentas
│   ├── account_types.json       # Tipos de cuenta
│   ├── dsl_templates.json       # Plantillas DSL
│   ├── organizations.json       # Empresas
│   ├── countries.json           # Países
│   ├── currencies.json          # Monedas
│   ├── user_profile.json        # Perfil usuario
│   ├── report_types.json        # Tipos de reportes
│   └── reports_recent.json      # Reportes generados
├── front/
│   ├── html/
│   │   ├── dashboard.html       # Dashboard
│   │   ├── vouchers_list.html   # Lista comprobantes
│   │   ├── vouchers_form.html   # Form comprobantes
│   │   ├── journal_entries.html # Asientos
│   │   ├── accounts_chart.html  # Plan de cuentas
│   │   ├── reports.html         # Reportes
│   │   └── dsl_editor.html      # Editor DSL
│   ├── js/
│   │   ├── api_service.js       # Servicio API
│   │   ├── utils.js             # Utilidades
│   │   ├── dashboard.js         # Lógica dashboard
│   │   ├── vouchers_list.js     # Lista comprobantes
│   │   ├── journal_entries.js   # Asientos
│   │   ├── accounts_chart.js    # Plan cuentas
│   │   ├── reports.js           # Reportes
│   │   └── dsl_editor.js        # Editor DSL
│   ├── css/
│   │   ├── style.css            # Estilos base
│   │   └── components.css       # Componentes
│   └── img/
│       └── avatar.png           # Placeholder
```

---

*Última actualización: Enero 2025*  
*Versión: 2.0 - Mock Funcional Completo*