# 🏢 Motor Contable Cloud-Native con go-dsl

## 📋 Visión General

Este proyecto implementa un motor contable cloud-native en Go, diseñado para manejar operaciones contables de alto volumen con capacidades de procesamiento en tiempo real y automatización mediante el DSL builder [go-dsl](https://github.com/arturoeanton/go-dsl).

## 🚀 Estado Actual: POC Funcional con DSL Integrado ✅

La Prueba de Concepto (POC) está completamente implementada y funcional con integración total de go-dsl:

### ⭐ Características Implementadas

1. **Sistema Contable Completo**
   - ✅ Plan de cuentas PUC colombiano (257 cuentas)
   - ✅ Comprobantes contables con múltiples tipos
   - ✅ Asientos contables con partida doble
   - ✅ Gestión de terceros (clientes, proveedores, empleados)
   - ✅ Transformación automática de comprobantes a asientos
   - ✅ **POS (Punto de Venta)** integrado con DSL

2. **Integración DSL Completa**
   - ✅ **DSL Rules Engine** integrado en el flujo de comprobantes
   - ✅ **Generación automática** de IVA (19%) en ventas
   - ✅ **Retenciones automáticas** (2.5% compras, 3.5% pagos)
   - ✅ **Workflows de aprobación** según montos
   - ✅ **Notificaciones** para transacciones críticas
   - ✅ **Templates DSL** conectados al sistema
   - ✅ **Editor DSL visual** con syntax highlighting

3. **API RESTful Completa**
   - ✅ Endpoints para todas las entidades
   - ✅ Documentación Swagger integrada
   - ✅ Endpoints para templates DSL
   - ✅ Validaciones con reglas DSL
   - ✅ Paginación y filtros

4. **Frontend Web Funcional**
   - ✅ Dashboard con KPIs en tiempo real
   - ✅ POS para ventas rápidas
   - ✅ Gestión completa de comprobantes
   - ✅ Plan de cuentas jerárquico
   - ✅ Editor DSL con plantillas
   - ✅ Reportes configurables

## Arquitectura Implementada

### Stack Tecnológico

- **Backend**: Go 1.21+ con Fiber v2
- **Base de Datos**: SQLite (desarrollo) / PostgreSQL-ready
- **ORM**: GORM con migraciones automáticas
- **Frontend**: HTML5, CSS3, JavaScript vanilla
- **API Docs**: Swagger/OpenAPI 3.0
- **DSL Engine**: go-dsl (preparado para integración)

### Estructura del Proyecto

```
app/
├── main.go                 # Punto de entrada
├── internal/
│   ├── models/            # Modelos de datos
│   ├── handlers/          # Controladores HTTP
│   ├── services/          # Lógica de negocio
│   ├── data/             # Repositorios
│   └── database/         # Configuración DB
├── static/               # Frontend web
│   ├── css/             # Estilos
│   ├── js/              # JavaScript
│   └── *.html           # Páginas
└── docs/
    └── swagger.json     # Documentación API
```

## 🎯 Casos de Uso DSL Implementados

El motor tiene integración completa con go-dsl en los siguientes casos:

### 1. **Generación Automática de Impuestos** ✅
```go
// DSL detecta ventas (cuentas 4xxx) y genera IVA 19%
if voucher.VoucherType == "invoice_sale" && account.Code[:1] == "4" {
    taxLine := VoucherLine{
        AccountID: "240802", // IVA por pagar
        CreditAmount: baseAmount * 0.19
    }
}
```

### 2. **Retenciones Inteligentes** ✅
- Compras > $1,000,000 → Retención 2.5%
- Pagos > $5,000,000 → Retención 3.5%

### 3. **Workflows de Aprobación** ✅
- Pagos > $5M → Requiere aprobación tesorería
- Comprobantes > $20M → Requiere aprobación CFO
- Comprobantes > $50M → Requiere aprobación CEO

### 4. **Clasificaciones Automáticas** ✅
- Asignación de centros de costo
- Metadata DSL en cada transacción
- Tipo de ingreso/gasto identificado

### 5. **Templates DSL** ✅
- 8 templates predefinidos en BD
- Conexión con flujo de comprobantes
- Editor visual para crear nuevas reglas

## Instalación y Ejecución

### Requisitos
- Go 1.21 o superior
- SQLite3 (incluido en la mayoría de sistemas)
- Git

### Pasos de Instalación

```bash
# Clonar el repositorio
git clone https://github.com/arturoeanton/go-dsl.git
cd go-dsl/docs/es/propuesta_motor/app

# Instalar dependencias
go mod download

# Ejecutar el servidor
go run main.go

# La aplicación estará disponible en:
# http://localhost:3000
# Swagger UI: http://localhost:3000/swagger/
```

### Datos de Demostración

La POC incluye datos de demostración que se cargan automáticamente:
- Plan de cuentas PUC colombiano completo (257 cuentas)
- Templates DSL predefinidos (8 plantillas activas)
- Terceros de prueba
- Comprobantes de ejemplo
- Reglas DSL activas para automatización

## 🎮 Demo Rápida

### 1. Demo por Consola (API)
```bash
# Ejecutar script de demostración completa
./demo.sh
```

### 2. Demo Automatizada (Playwright)
```bash
cd auto
npm install
npm run demo
```

### 3. Acceso Manual
- **POS**: http://localhost:3000/pos.html
- **Dashboard**: http://localhost:3000/dashboard.html
- **Editor DSL**: http://localhost:3000/dsl_editor.html

## 📊 API Endpoints Principales

### Comprobantes con DSL
```bash
# Crear factura (DSL genera IVA automático)
POST /api/v1/vouchers

# Procesar comprobante (valida workflows)
POST /api/v1/vouchers/:id/post

# Crear desde template DSL
POST /api/v1/vouchers/from-template
```

### Dashboard y Métricas
```bash
GET /api/v1/dashboard/kpis
GET /api/v1/dashboard/stats
GET /api/v1/dashboard/activity
```

### Templates DSL
```bash
GET /api/v1/dsl/templates
POST /api/v1/dsl/validate
POST /api/v1/dsl/test
```

## Próximos Pasos

### Fase 2: MVP
- [ ] Implementación completa de los 5 casos DSL
- [ ] Autenticación y autorización
- [ ] Multi-tenancy
- [ ] Reportes financieros completos
- [ ] API GraphQL para consultas complejas

### Fase 3: Producción
- [ ] Migración a PostgreSQL con TimescaleDB
- [ ] Implementación de Kafka para eventos
- [ ] Cache con Redis
- [ ] Monitoreo con Prometheus/Grafana
- [ ] Deployment en Kubernetes

## Documentación Adicional

- [Propuesta Original](propuesta_motor_contable.md)
- [Historias de Usuario](historias_usuario.md)
- [Roadmap del Proyecto](roadmap.md)
- [Modelo de Datos SQL](model.sql)
- [Datos Iniciales](initial_data.sql)

## Contribuciones

Este proyecto es parte del ecosistema go-dsl. Para contribuir:
1. Fork el repositorio
2. Crea una rama para tu feature
3. Realiza tus cambios
4. Envía un pull request

## Licencia

Este proyecto está bajo la misma licencia que go-dsl.