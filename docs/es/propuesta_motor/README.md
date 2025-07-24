# Motor Contable Cloud-Native con go-dsl

## Visión General

Este proyecto implementa un motor contable cloud-native en Go, diseñado para manejar operaciones contables de alto volumen con capacidades de procesamiento en tiempo real y automatización mediante el DSL builder [go-dsl](https://github.com/arturoeanton/go-dsl).

## Estado Actual: POC Funcional ✅

La Prueba de Concepto (POC) está completamente implementada y funcional con:

### Características Implementadas

1. **Sistema Contable Completo**
   - ✅ Plan de cuentas PUC colombiano (257 cuentas)
   - ✅ Comprobantes contables con múltiples tipos
   - ✅ Asientos contables con partida doble
   - ✅ Gestión de terceros (clientes, proveedores, empleados)
   - ✅ Transformación automática de comprobantes a asientos

2. **API RESTful**
   - ✅ Endpoints completos para todas las entidades
   - ✅ Documentación Swagger integrada
   - ✅ Validaciones y manejo de errores
   - ✅ Paginación y filtros

3. **Frontend Web**
   - ✅ Dashboard con estadísticas en tiempo real
   - ✅ Gestión de comprobantes
   - ✅ Visualización de asientos contables
   - ✅ Plan de cuentas interactivo con árbol expandible
   - ✅ Editor DSL con sintaxis highlighting

4. **Integración DSL**
   - ✅ Editor de reglas DSL
   - ✅ Templates predefinidos
   - ✅ Validación de sintaxis
   - ✅ Preparado para ejecución de reglas

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

## Casos de Uso del DSL

El motor está preparado para implementar 5 casos de uso principales con go-dsl:

1. **[Validación Inteligente de Comprobantes](caso_dsl_1.md)**
   - Validaciones contables complejas
   - Reglas de negocio personalizables
   - Detección de errores y anomalías

2. **[Cálculo Automático de Impuestos](caso_dsl_2.md)**
   - IVA, retenciones, ICA
   - Reglas según tipo de tercero
   - Cumplimiento normativo colombiano

3. **[Distribución Automática de Costos](caso_dsl_3.md)**
   - Prorrateo por centros de costo
   - Distribución por drivers
   - Asignación multinivel

4. **[Conciliación Bancaria Inteligente](caso_dsl_4.md)**
   - Matching automático
   - Reglas de conciliación
   - Detección de diferencias

5. **[Templates de Asientos Recurrentes](caso_dsl_5.md)**
   - Plantillas parametrizables
   - Generación automática
   - Variables y cálculos dinámicos

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
- Plan de cuentas PUC colombiano completo
- Comprobantes de ejemplo (ventas, compras, nómina, etc.)
- Terceros de prueba
- Asientos contables generados

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