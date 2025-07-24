# Historias de Usuario - Motor Contable Cloud-Native

## 📋 Información General

**Proyecto**: Motor Contable Cloud-Native  
**Versión**: 2.0  
**Fecha**: Enero 2025  
**Stack**: Go/Fiber/go-dsl/PostgreSQL  

## 🎯 Metodología

- **Framework**: Scrum/Agile
- **Estimación**: Story Points (Fibonacci)
- **Criterios de Aceptación**: Given/When/Then (Gherkin)
- **Priorización**: MoSCoW (Must/Should/Could/Won't)

---

## 🚀 **ÉPICA 1: FUNDACIÓN DEL SISTEMA**
### **Fase 1 - Monolito Base**

### **US-001: Setup Inicial del Proyecto**
**Como** desarrollador  
**Quiero** configurar la estructura base del proyecto Go/Fiber  
**Para** poder comenzar el desarrollo del motor contable  

**Criterios de Aceptación**:
- **Dado** que necesito iniciar el proyecto
- **Cuando** ejecuto el setup inicial
- **Entonces** debo tener:
  - Estructura de directorios Go estándar
  - Configuración Fiber v2
  - Conexión PostgreSQL 15+
  - Variables de entorno configuradas
  - Logging estructurado funcionando

**Estimación**: 3 SP  
**Prioridad**: Must Have  
**Dependencias**: Ninguna  

---

### **US-002: Base de Datos y Migraciones**
**Como** administrador del sistema  
**Quiero** tener el esquema de base de datos inicial  
**Para** poder almacenar la información contable  

**Criterios de Aceptación**:
- **Dado** que tengo PostgreSQL configurado
- **Cuando** ejecuto las migraciones
- **Entonces** debo tener:
  - 21 tablas creadas correctamente
  - Índices de performance aplicados
  - Particionamiento por fecha funcionando
  - Datos iniciales (org demo, cuentas PUC)
  - Row Level Security configurado

**Estimación**: 5 SP  
**Prioridad**: Must Have  
**Dependencias**: US-001  

---

### **US-003: Integración Motor go-dsl**
**Como** desarrollador  
**Quiero** integrar el motor go-dsl existente  
**Para** procesar plantillas de contabilización  

**Criterios de Aceptación**:
- **Dado** que tengo el motor go-dsl disponible
- **Cuando** integro el motor en la aplicación
- **Entonces** debo poder:
  - Compilar plantillas DSL desde base de datos
  - Ejecutar plantillas con contexto
  - Cachear plantillas compiladas
  - Manejar errores de sintaxis/ejecución
  - Validar resultado de plantillas

**Estimación**: 8 SP  
**Prioridad**: Must Have  
**Dependencias**: US-001, US-002  

---

## 📊 **ÉPICA 2: APIs CORE**
### **Dashboard y Analytics**

### **US-004: API Dashboard - KPIs**
**Como** usuario contable  
**Quiero** ver indicadores clave en el dashboard  
**Para** monitorear el estado del sistema  

**Criterios de Aceptación**:
- **Dado** que accedo al dashboard
- **Cuando** solicito los KPIs
- **Entonces** debo ver:
  - Comprobantes procesados hoy/mes
  - Comprobantes pendientes
  - Tasa de procesamiento
  - Tiempo promedio de procesamiento
  - Estados actualizados en tiempo real

**Endpoint**: `GET /api/dashboard`  
**Estimación**: 5 SP  
**Prioridad**: Must Have  
**Dependencias**: US-003  

---

### **US-005: API Dashboard - Gráficos**
**Como** usuario contable  
**Quiero** visualizar datos en gráficos interactivos  
**Para** analizar tendencias de procesamiento  

**Criterios de Aceptación**:
- **Dado** que estoy en el dashboard
- **Cuando** solicito datos de gráficos
- **Entonces** debo obtener:
  - Datos para gráfico de líneas (7 días)
  - Datos para gráfico de dona (tipos)
  - Datos para gráfico de barras (estados)
  - Formato compatible con Chart.js
  - Datos agregados correctamente

**Endpoint**: `GET /api/dashboard/charts`  
**Estimación**: 3 SP  
**Prioridad**: Should Have  
**Dependencias**: US-004  

---

## 📄 **ÉPICA 3: GESTIÓN DE COMPROBANTES**

### **US-006: CRUD Comprobantes - Listar**
**Como** usuario contable  
**Quiero** listar comprobantes con filtros  
**Para** encontrar documentos específicos  

**Criterios de Aceptación**:
- **Dado** que accedo a la lista de comprobantes
- **Cuando** solicito la lista con filtros
- **Entonces** debo poder:
  - Filtrar por tipo, estado, fecha, organización
  - Buscar por número, descripción, tercero
  - Paginar resultados (10/25/50/100)
  - Ordenar por cualquier columna
  - Ver estado visual con badges

**Endpoint**: `GET /api/vouchers`  
**Estimación**: 8 SP  
**Prioridad**: Must Have  
**Dependencias**: US-002  

---

### **US-007: CRUD Comprobantes - Crear**
**Como** usuario contable  
**Quiero** crear nuevos comprobantes  
**Para** registrar transacciones comerciales  

**Criterios de Aceptación**:
- **Dado** que estoy creando un comprobante
- **Cuando** ingreso los datos requeridos
- **Entonces** debo poder:
  - Seleccionar tipo de comprobante
  - Ingresar datos de cabecera
  - Agregar múltiples líneas de detalle
  - Calcular totales automáticamente
  - Validar balance antes de guardar
  - Generar número consecutivo automático

**Endpoint**: `POST /api/vouchers`  
**Estimación**: 13 SP  
**Prioridad**: Must Have  
**Dependencias**: US-006, US-003  

---

### **US-008: CRUD Comprobantes - Editar/Eliminar**
**Como** usuario contable  
**Quiero** modificar comprobantes existentes  
**Para** corregir errores o actualizar información  

**Criterios de Aceptación**:
- **Dado** que tengo un comprobante existente
- **Cuando** realizo modificaciones
- **Entonces** debo poder:
  - Editar solo comprobantes en estado borrador
  - Modificar líneas de detalle
  - Recalcular totales automáticamente
  - Eliminar comprobantes no procesados
  - Mantener audit trail de cambios

**Endpoints**: `PUT /api/vouchers/{id}`, `DELETE /api/vouchers/{id}`  
**Estimación**: 8 SP  
**Prioridad**: Must Have  
**Dependencias**: US-007  

---

## 📚 **ÉPICA 4: CONTABILIZACIÓN AUTOMÁTICA**

### **US-009: Procesamiento DSL - Motor**
**Como** sistema  
**Quiero** procesar comprobantes con plantillas DSL  
**Para** generar asientos contables automáticamente  

**Criterios de Aceptación**:
- **Dado** que tengo un comprobante pendiente
- **Cuando** ejecuto el procesamiento
- **Entonces** debo:
  - Seleccionar plantilla DSL correcta
  - Ejecutar plantilla con datos del comprobante
  - Generar líneas de asiento balanceadas
  - Validar que débitos = créditos
  - Cambiar estado a "PROCESADO"
  - Registrar errores si los hay

**Endpoint**: `POST /api/vouchers/{id}/process`  
**Estimación**: 13 SP  
**Prioridad**: Must Have  
**Dependencias**: US-007, US-003  

---

### **US-010: Asientos Contables - CRUD**
**Como** usuario contable  
**Quiero** gestionar asientos contables  
**Para** revisar y crear asientos manuales  

**Criterios de Aceptación**:
- **Dado** que accedo a asientos contables
- **Cuando** realizo operaciones CRUD
- **Entonces** debo poder:
  - Listar asientos con filtros por período
  - Ver detalle completo del asiento
  - Crear asientos manuales balanceados
  - Reversar asientos existentes
  - Validar balance automáticamente

**Endpoints**: `GET/POST/PUT/DELETE /api/journal-entries`  
**Estimación**: 10 SP  
**Prioridad**: Must Have  
**Dependencias**: US-009  

---

## 🏦 **ÉPICA 5: PLAN DE CUENTAS**

### **US-011: Plan Cuentas - Visualización**
**Como** usuario contable  
**Quiero** navegar el plan de cuentas jerárquico  
**Para** entender la estructura contable  

**Criterios de Aceptación**:
- **Dado** que accedo al plan de cuentas
- **Cuando** navego la estructura
- **Entonces** debo ver:
  - Árbol expandible hasta 5 niveles
  - Indicadores visuales por tipo
  - Naturaleza (Débito/Crédito)
  - Estado activo/inactivo
  - Búsqueda por código/nombre
  - Cuentas de detalle marcadas

**Endpoint**: `GET /api/accounts/tree`  
**Estimación**: 8 SP  
**Prioridad**: Must Have  
**Dependencias**: US-002  

---

### **US-012: Plan Cuentas - Gestión**
**Como** administrador contable  
**Quiero** gestionar cuentas contables  
**Para** mantener el catálogo actualizado  

**Criterios de Aceptación**:
- **Dado** que soy administrador
- **Cuando** gestiono cuentas
- **Entonces** debo poder:
  - Crear nuevas cuentas respetando jerarquía
  - Modificar propiedades de cuentas
  - Activar/desactivar cuentas
  - Validar códigos únicos
  - Mantener integridad referencial

**Endpoints**: `POST/PUT /api/accounts`  
**Estimación**: 10 SP  
**Prioridad**: Must Have  
**Dependencias**: US-011  

---

## 📈 **ÉPICA 6: REPORTES Y LIBROS**

### **US-013: Generador de Reportes**
**Como** usuario contable  
**Quiero** generar reportes estándar  
**Para** cumplir requisitos contables y fiscales  

**Criterios de Aceptación**:
- **Dado** que solicito un reporte
- **Cuando** configuro parámetros
- **Entonces** debo poder:
  - Seleccionar entre 12 tipos de reporte
  - Configurar período de fechas
  - Filtrar por cuentas/centros de costo
  - Exportar en PDF/Excel/CSV
  - Ver vista previa antes de generar
  - Acceder a historial de reportes

**Endpoint**: `POST /api/reports/generate`  
**Estimación**: 13 SP  
**Prioridad**: Must Have  
**Dependencias**: US-010  

---

### **US-014: Libros Contables**
**Como** auditor  
**Quiero** acceder a libros contables oficiales  
**Para** realizar auditorías y cumplir normativa  

**Criterios de Aceptación**:
- **Dado** que requiero libros oficiales
- **Cuando** solicito generación
- **Entonces** debo obtener:
  - Libro Diario completo
  - Libro Mayor por cuenta
  - Balance de Comprobación
  - Formato PDF oficial
  - Numeración consecutiva
  - Firmas digitales opcionales

**Endpoint**: `POST /api/books/generate`  
**Estimación**: 10 SP  
**Prioridad**: Should Have  
**Dependencias**: US-013  

---

## ⚙️ **ÉPICA 7: EDITOR DSL**

### **US-015: Editor DSL - Plantillas**
**Como** desarrollador contable  
**Quiero** gestionar plantillas DSL  
**Para** personalizar reglas de contabilización  

**Criterios de Aceptación**:
- **Dado** que accedo al editor DSL
- **Cuando** trabajo con plantillas
- **Entonces** debo poder:
  - Crear/editar plantillas DSL
  - Validar sintaxis en tiempo real
  - Probar con datos de ejemplo
  - Ver variables disponibles
  - Versionar plantillas
  - Activar/desactivar plantillas

**Endpoints**: `GET/POST/PUT /api/dsl/templates`  
**Estimación**: 13 SP  
**Prioridad**: Should Have  
**Dependencias**: US-003, US-009  

---

### **US-016: DSL - Ejemplos y Documentación**
**Como** usuario del sistema  
**Quiero** acceder a ejemplos DSL  
**Para** aprender y crear mis propias plantillas  

**Criterios de Aceptación**:
- **Dado** que estoy en el editor DSL
- **Cuando** busco ayuda
- **Entonces** debo encontrar:
  - 5 plantillas de ejemplo funcionales
  - Documentación de sintaxis
  - Lista de funciones disponibles
  - Snippets de código común
  - Casos de uso típicos

**Endpoint**: `GET /api/dsl/examples`  
**Estimación**: 5 SP  
**Prioridad**: Should Have  
**Dependencias**: US-015  

---

## 📋 **ÉPICA 8: CATÁLOGOS Y CONFIGURACIÓN**

### **US-017: Gestión de Terceros**
**Como** usuario contable  
**Quiero** gestionar terceros (clientes/proveedores)  
**Para** asociarlos a comprobantes  

**Criterios de Aceptación**:
- **Dado** que gestiono terceros
- **Cuando** realizo operaciones CRUD
- **Entonces** debo poder:
  - Crear terceros con datos fiscales
  - Configurar cuentas por cobrar/pagar
  - Establecer límites de crédito
  - Validar documentos de identidad
  - Buscar por nombre/documento
  - Marcar como activo/inactivo

**Endpoints**: `GET/POST/PUT/DELETE /api/third-parties`  
**Estimación**: 10 SP  
**Prioridad**: Must Have  
**Dependencias**: US-007  

---

### **US-018: Configuración del Sistema**
**Como** administrador  
**Quiero** configurar parámetros del sistema  
**Para** personalizar comportamiento por organización  

**Criterios de Aceptación**:
- **Dado** que soy administrador
- **Cuando** configuro el sistema
- **Entonces** debo poder:
  - Configurar moneda por defecto
  - Establecer año fiscal
  - Configurar numeración de comprobantes
  - Definir tipos de comprobante
  - Configurar impuestos por país
  - Establecer períodos contables

**Endpoints**: `GET/PUT /api/settings`  
**Estimación**: 8 SP  
**Prioridad**: Should Have  
**Dependencias**: US-002  

---

## 🔐 **ÉPICA 9: AUTENTICACIÓN (FASE 2)**

### **US-019: Sistema de Login**
**Como** usuario del sistema  
**Quiero** autenticarme de forma segura  
**Para** acceder a funcionalidades según mi rol  

**Criterios de Aceptación**:
- **Dado** que accedo al sistema
- **Cuando** me autentico
- **Entonces** debo poder:
  - Login con email/contraseña
  - Recibir JWT token válido
  - Refrescar token automáticamente
  - Logout y invalidar sesión
  - Bloqueo tras intentos fallidos
  - Recuperación de contraseña

**Endpoints**: `POST /api/auth/login`, `POST /api/auth/refresh`, `POST /api/auth/logout`  
**Estimación**: 10 SP  
**Prioridad**: Must Have (Fase 2)  
**Dependencias**: Todas las APIs de Fase 1  

---

### **US-020: Control de Acceso (RBAC)**
**Como** administrador  
**Quiero** controlar acceso granular  
**Para** garantizar seguridad por rol  

**Criterios de Aceptación**:
- **Dado** que un usuario accede al sistema
- **Cuando** intenta realizar una acción
- **Entonces** el sistema debe:
  - Validar JWT token
  - Verificar permisos por endpoint
  - Aplicar filtros por organización
  - Registrar en audit log
  - Denegar acceso no autorizado
  - Mostrar solo opciones permitidas

**Roles**: SUPER_ADMIN, ORG_ADMIN, MANAGER, ACCOUNTANT, AUDITOR, CLERK, VIEWER  
**Estimación**: 13 SP  
**Prioridad**: Must Have (Fase 2)  
**Dependencias**: US-019  

---

## 🐳 **ÉPICA 10: DOCKERIZACIÓN (FASE 3)**

### **US-021: Containerización**
**Como** DevOps engineer  
**Quiero** containerizar la aplicación  
**Para** deployar en cualquier entorno  

**Criterios de Aceptación**:
- **Dado** que tengo la aplicación lista
- **Cuando** la containerizo
- **Entonces** debo tener:
  - Dockerfile multi-stage optimizado
  - Imagen < 50MB
  - Docker Compose completo
  - Health checks configurados
  - Variables de entorno
  - Volúmenes persistentes

**Estimación**: 8 SP  
**Prioridad**: Must Have (Fase 3)  
**Dependencias**: Todas las funcionalidades  

---

### **US-022: CI/CD Pipeline**
**Como** desarrollador  
**Quiero** pipeline automatizado  
**Para** deployar de forma confiable  

**Criterios de Aceptación**:
- **Dado** que hago push al repositorio
- **Cuando** se ejecuta el pipeline
- **Entonces** debe:
  - Ejecutar tests automáticamente
  - Crear imagen Docker
  - Realizar security scan
  - Deploy a staging/producción
  - Rollback automático en fallos
  - Notificar resultados

**Estimación**: 10 SP  
**Prioridad**: Should Have (Fase 3)  
**Dependencias**: US-021  

---

## 📊 **RESUMEN DE ÉPICAS**

| **Épica** | **Historias** | **Story Points** | **Prioridad** | **Fase** |
|-----------|---------------|------------------|---------------|----------|
| 1. Fundación | US-001 a US-003 | 16 SP | Must Have | 1 |
| 2. APIs Core | US-004 a US-005 | 8 SP | Must Have | 1 |
| 3. Comprobantes | US-006 a US-008 | 29 SP | Must Have | 1 |
| 4. Contabilización | US-009 a US-010 | 23 SP | Must Have | 1 |
| 5. Plan Cuentas | US-011 a US-012 | 18 SP | Must Have | 1 |
| 6. Reportes | US-013 a US-014 | 23 SP | Must Have | 1 |
| 7. Editor DSL | US-015 a US-016 | 18 SP | Should Have | 1 |
| 8. Catálogos | US-017 a US-018 | 18 SP | Must Have | 1 |
| 9. Autenticación | US-019 a US-020 | 23 SP | Must Have | 2 |
| 10. Docker/CI | US-021 a US-022 | 18 SP | Must Have | 3 |

**Total**: 22 Historias de Usuario, 172 Story Points

---

## 🚧 **CRITERIOS DE DEFINICIÓN DE HECHO (DoD)**

Para cada historia de usuario:

### **Desarrollo**
- [ ] Código implementado según criterios de aceptación
- [ ] Tests unitarios con coverage > 80%
- [ ] Tests de integración funcionando
- [ ] Documentación API actualizada
- [ ] Code review aprobado

### **Testing**
- [ ] Tests manuales pasados
- [ ] Tests automatizados en pipeline
- [ ] Validación de performance (< 200ms)
- [ ] Pruebas de seguridad básicas
- [ ] Testing en diferentes navegadores

### **Deployment**
- [ ] Desplegado en environment de testing
- [ ] Migraciones de DB ejecutadas
- [ ] Configuración de producción validada
- [ ] Monitoring y logs funcionando
- [ ] Rollback plan documentado

---

## 📅 **PLANIFICACIÓN DE SPRINTS**

### **Sprint 1-2: Fundación (4 semanas)**
- US-001, US-002, US-003
- Setup completo del proyecto
- Base de datos operativa
- Motor DSL integrado

### **Sprint 3-4: APIs Core (4 semanas)**
- US-004, US-005, US-006, US-007
- Dashboard funcional
- CRUD comprobantes básico

### **Sprint 5-6: Contabilización (4 semanas)**
- US-008, US-009, US-010
- Procesamiento automático
- Asientos contables

### **Sprint 7-8: Plan y Reportes (4 semanas)**
- US-011, US-012, US-013, US-014
- Plan de cuentas completo
- Generación de reportes

### **Sprint 9-10: DSL y Catálogos (4 semanas)**
- US-015, US-016, US-017, US-018
- Editor DSL funcional
- Configuración completa

### **Sprint 11-12: Autenticación (4 semanas)**
- US-019, US-020
- Login y RBAC completo

### **Sprint 13-14: Dockerización (4 semanas)**
- US-021, US-022
- Containerización y CI/CD

---

*Última actualización: Enero 2025*  
*Versión: 2.0 - Historias Completas*