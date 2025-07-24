# HU-003: Catálogo de Cuentas Contables

## Historia de Usuario

**Como** contador  
**Quiero** gestionar el plan de cuentas contables  
**Para** organizar correctamente la estructura contable de la empresa según normativas

## Criterios de Aceptación

1. ✅ Estructura jerárquica hasta 6 niveles
2. ✅ Códigos de cuenta únicos por organización
3. ✅ Tipos: Activo, Pasivo, Patrimonio, Ingreso, Gasto
4. ✅ Naturaleza: Débito o Crédito
5. ✅ Importación desde Excel/CSV
6. ✅ Plantillas por país (NIIF, SAT, SII)
7. ✅ Búsqueda por código, nombre o tipo
8. ✅ Validación de integridad referencial
9. ✅ Historial de cambios
10. ✅ Activar/desactivar cuentas

## Especificaciones Técnicas

- **Tabla Principal**: `chart_of_accounts`
- **Estructura**: Árbol con `parent_id`
- **API Base**: `/api/v1/accounts`
- **Caché**: Redis para árbol completo
- **Importación**: Procesamiento asíncrono

## Tareas de Desarrollo

### 1. Backend - Modelo de Datos (2h)
- [ ] Optimizar tabla `chart_of_accounts`
- [ ] Crear vista materializada para árbol
- [ ] Índices para búsquedas recursivas
- [ ] Triggers para validar jerarquía

### 2. Backend - Account Service (5h)
- [ ] CRUD de cuentas con validaciones
- [ ] Método para obtener árbol completo
- [ ] Validación de eliminación (sin movimientos)
- [ ] Clonación de estructura entre orgs
- [ ] Verificación de integridad

### 3. Backend - API Endpoints (3h)
- [ ] `GET /accounts/tree` estructura completa
- [ ] `GET /accounts` listado plano con filtros
- [ ] `POST /accounts` crear con validación padre
- [ ] `PUT /accounts/:id` actualizar
- [ ] `DELETE /accounts/:id` soft delete
- [ ] `POST /accounts/import` carga masiva
- [ ] `GET /accounts/templates/:country` plantillas

### 4. Backend - Importador (4h)
- [ ] Parser para Excel (xlsx)
- [ ] Parser para CSV
- [ ] Validación de estructura
- [ ] Detección de duplicados
- [ ] Reporte de errores detallado
- [ ] Rollback en caso de fallo

### 5. Frontend - Vista de Árbol (5h)
- [ ] Componente `AccountTree` con react-tree
- [ ] Expansión/colapso de nodos
- [ ] Drag & drop para reorganizar
- [ ] Menú contextual (editar, eliminar)
- [ ] Indicadores visuales por tipo
- [ ] Búsqueda con highlighting

### 6. Frontend - Formulario de Cuenta (3h)
- [ ] Modal para crear/editar cuenta
- [ ] Selección de cuenta padre
- [ ] Validación de código único
- [ ] Preview de código completo
- [ ] Campos condicionales por tipo

### 7. Frontend - Importador (4h)
- [ ] Wizard de importación
- [ ] Drag & drop para archivos
- [ ] Mapeo de columnas
- [ ] Preview de datos
- [ ] Progreso en tiempo real
- [ ] Descarga de reporte de errores

### 8. Plantillas por País (6h)
- [ ] Plantilla Colombia (NIIF PYMES)
- [ ] Plantilla México (SAT)
- [ ] Plantilla Chile (SII)
- [ ] Sistema de versionado
- [ ] Actualización automática

### 9. Validaciones Especiales (3h)
- [ ] Cuentas de mayor no pueden tener movimientos
- [ ] Cambio de tipo requiere sin saldos
- [ ] Códigos según formato del país
- [ ] Niveles máximos configurables

### 10. Testing (4h)
- [ ] Tests de jerarquía y ciclos
- [ ] Tests de importación con casos límite
- [ ] Tests de rendimiento (árbol 1000+ nodos)
- [ ] Tests de concurrencia

### 11. Documentación (2h)
- [ ] Guía de estructura contable
- [ ] Formato de importación
- [ ] Best practices por país
- [ ] Video tutorial

## Estimación Total: 41 horas

## Dependencias

- HU-001: Autenticación (permisos por rol)

## Riesgos

1. **Performance con árboles grandes**: Implementar lazy loading
2. **Ciclos en jerarquía**: Validación en DB con CTE recursivo
3. **Importaciones grandes**: Procesar en chunks de 1000
4. **Cambios que rompen integridad**: Validación exhaustiva

## Notas de Implementación

```sql
-- Query recursivo para obtener árbol completo
WITH RECURSIVE account_tree AS (
    -- Cuentas raíz
    SELECT 
        id, account_code, name, type, nature, 
        level, parent_id, is_detail,
        account_code as path,
        ARRAY[id] as path_ids
    FROM chart_of_accounts
    WHERE parent_id IS NULL 
        AND organization_id = $1
        AND is_active = true
    
    UNION ALL
    
    -- Cuentas hijas
    SELECT 
        c.id, c.account_code, c.name, c.type, c.nature,
        c.level, c.parent_id, c.is_detail,
        t.path || '.' || c.account_code,
        t.path_ids || c.id
    FROM chart_of_accounts c
    INNER JOIN account_tree t ON c.parent_id = t.id
    WHERE c.is_active = true
)
SELECT * FROM account_tree
ORDER BY path;
```

```typescript
// Estructura del árbol de cuentas
interface AccountNode {
  id: string;
  code: string;
  name: string;
  type: AccountType;
  nature: 'D' | 'C';
  level: number;
  isDetail: boolean;
  children?: AccountNode[];
  metadata?: {
    lastMovement?: Date;
    balance?: number;
    isActive: boolean;
  };
}

// Tipos de cuenta
enum AccountType {
  ASSET = 'ASSET',
  LIABILITY = 'LIABILITY', 
  EQUITY = 'EQUITY',
  INCOME = 'INCOME',
  EXPENSE = 'EXPENSE'
}
```

## Mockups Relacionados

- [Vista de Árbol de Cuentas](../mocks/front/html/accounts_chart.html)
- [Importador de Cuentas](../mocks/front/html/accounts_import.html)
- [API Accounts Tree](../mocks/api/accounts_tree.json)

## Métricas de Éxito

- Tiempo de carga del árbol: < 500ms
- Éxito en importaciones: > 95%
- Errores de validación: < 2%
- Adopción de plantillas: > 80%