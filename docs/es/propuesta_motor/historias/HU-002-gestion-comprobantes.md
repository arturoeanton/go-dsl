# HU-002: Gestión de Comprobantes

## Historia de Usuario

**Como** auxiliar contable  
**Quiero** poder crear, editar y gestionar comprobantes  
**Para** registrar todas las transacciones de la empresa de forma organizada

## Criterios de Aceptación

1. ✅ Crear comprobantes con validación en tiempo real
2. ✅ Soporte para múltiples tipos: Factura, Recibo, Pago, Nota Débito/Crédito
3. ✅ Numeración automática según configuración
4. ✅ Búsqueda por número, fecha, tercero o monto
5. ✅ Adjuntar archivos PDF/XML (máx 10MB)
6. ✅ Validación de totales y cálculos automáticos
7. ✅ Estados: Borrador, Pendiente, Procesado, Anulado
8. ✅ Historial de cambios completo

## Especificaciones Técnicas

- **Tabla Principal**: `vouchers` (particionada por fecha)
- **API Base**: `/api/v1/vouchers`
- **Paginación**: 50 registros por página
- **Archivos**: Almacenamiento en Object Storage
- **Validaciones**: Según tipo y país

## Tareas de Desarrollo

### 1. Backend - Modelo de Datos (3h)
- [ ] Extender tabla `vouchers` con campos adicionales
- [ ] Crear tabla `voucher_lines` para items
- [ ] Crear tabla `voucher_attachments`
- [ ] Implementar particionamiento mensual
- [ ] Índices para búsquedas frecuentes

### 2. Backend - Voucher Service (6h)
- [ ] CRUD básico de comprobantes
- [ ] Validaciones por tipo de documento
- [ ] Cálculo automático de totales
- [ ] Generación de numeración secuencial
- [ ] Manejo de estados y transiciones
- [ ] Soft delete con auditoría

### 3. Backend - API Endpoints (4h)
- [ ] `GET /vouchers` con filtros y paginación
- [ ] `GET /vouchers/:id` con relaciones
- [ ] `POST /vouchers` con validación
- [ ] `PUT /vouchers/:id` con control de versión
- [ ] `DELETE /vouchers/:id` (soft delete)
- [ ] `POST /vouchers/bulk` para carga masiva
- [ ] `GET /vouchers/export` en CSV/Excel

### 4. Backend - Integración de Archivos (3h)
- [ ] Upload de archivos a Object Storage
- [ ] Validación de tipos y tamaños
- [ ] Generación de URLs firmadas
- [ ] Limpieza de archivos huérfanos

### 5. Frontend - Lista de Comprobantes (4h)
- [ ] Componente `VoucherList` con DataGrid
- [ ] Filtros avanzados (fecha, tipo, estado)
- [ ] Ordenamiento por columnas
- [ ] Acciones bulk (eliminar, exportar)
- [ ] Vista previa rápida

### 6. Frontend - Formulario de Comprobante (6h)
- [ ] Componente `VoucherForm` con validación
- [ ] Sección de información básica
- [ ] Gestión dinámica de líneas/items
- [ ] Cálculo automático de totales
- [ ] Drag & drop para archivos
- [ ] Autocompletado de terceros

### 7. Frontend - Flujo de Estados (2h)
- [ ] Visualización de estado actual
- [ ] Botones de acción según estado
- [ ] Confirmación para cambios críticos
- [ ] Timeline de cambios

### 8. Validaciones y Reglas de Negocio (4h)
- [ ] Validación de formato de números según país
- [ ] Verificación de fechas válidas
- [ ] Control de duplicados
- [ ] Límites por tipo de usuario
- [ ] Validaciones fiscales básicas

### 9. Testing (4h)
- [ ] Tests unitarios para VoucherService
- [ ] Tests de integración API
- [ ] Tests de componentes React
- [ ] Tests E2E flujo completo
- [ ] Tests de carga (1000 comprobantes)

### 10. Optimización (2h)
- [ ] Caché de listados frecuentes
- [ ] Lazy loading de relaciones
- [ ] Compresión de respuestas
- [ ] Índices para reportes

### 11. Documentación (2h)
- [ ] API docs con ejemplos
- [ ] Guía de usuario con screenshots
- [ ] Diagrama de estados
- [ ] Troubleshooting común

## Estimación Total: 40 horas

## Dependencias

- HU-001: Autenticación (para permisos)
- HU-003: Catálogo de Cuentas (para validaciones)

## Riesgos

1. **Performance con volúmenes altos**: Implementar paginación del lado del servidor
2. **Concurrencia en numeración**: Usar secuencias atómicas de PostgreSQL
3. **Archivos grandes**: Implementar upload por chunks
4. **Validaciones complejas**: Cachear reglas por tipo

## Notas de Implementación

```typescript
// Estructura de un comprobante
interface Voucher {
  id: string;
  organization_id: string;
  voucher_number: string;
  voucher_type: VoucherType;
  voucher_date: Date;
  description: string;
  total_amount: number;
  currency_code: string;
  status: VoucherStatus;
  metadata: {
    customer?: Customer;
    items?: LineItem[];
    taxes?: Record<string, number>;
    attachments?: Attachment[];
  };
  created_at: Date;
  updated_at: Date;
}

// Estados del comprobante
enum VoucherStatus {
  DRAFT = 'DRAFT',
  PENDING = 'PENDING',
  PROCESSING = 'PROCESSING',
  PROCESSED = 'PROCESSED',
  ERROR = 'ERROR',
  CANCELLED = 'CANCELLED'
}
```

## Mockups Relacionados

- [Lista de Comprobantes](../mocks/front/html/vouchers_list.html)
- [Formulario de Comprobante](../mocks/front/html/vouchers_form.html)
- [API Voucher Response](../mocks/api/vouchers_create.json)

## Métricas de Éxito

- Tiempo promedio de creación: < 2 minutos
- Tasa de error en validaciones: < 5%
- Comprobantes procesados por minuto: > 100
- Satisfacción del usuario: > 90%