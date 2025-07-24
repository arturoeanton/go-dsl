# HU-001: Autenticación de Usuarios

## Historia de Usuario

**Como** usuario del sistema contable  
**Quiero** poder autenticarme de forma segura  
**Para** acceder a las funcionalidades según mi rol y organización

## Criterios de Aceptación

1. ✅ El usuario puede iniciar sesión con email y contraseña
2. ✅ El sistema valida credenciales contra la base de datos
3. ✅ Se genera un JWT con expiración de 1 hora
4. ✅ Se incluye refresh token con expiración de 7 días
5. ✅ El token incluye: user_id, org_id, role, permissions
6. ✅ Se registra el último acceso del usuario
7. ✅ Se bloquea la cuenta después de 5 intentos fallidos
8. ✅ Se envía email de notificación en accesos sospechosos

## Especificaciones Técnicas

- **Endpoint**: `POST /api/v1/auth/login`
- **Método**: JWT con RS256
- **Duración Access Token**: 1 hora
- **Duración Refresh Token**: 7 días
- **Encriptación Password**: bcrypt (cost 10)

## Tareas de Desarrollo

### 1. Backend - Modelo de Datos (2h)
- [ ] Crear tabla `users` con campos requeridos
- [ ] Crear tabla `user_sessions` para tracking
- [ ] Crear tabla `login_attempts` para seguridad
- [ ] Añadir índices para email y org_id

### 2. Backend - Servicio de Autenticación (4h)
- [ ] Implementar `AuthService` en Go
- [ ] Método `Login(email, password) (tokens, error)`
- [ ] Método `RefreshToken(refreshToken) (newTokens, error)`
- [ ] Método `Logout(token) error`
- [ ] Validación de intentos fallidos

### 3. Backend - JWT Manager (3h)
- [ ] Implementar generación de tokens JWT
- [ ] Configurar claves RSA para firma
- [ ] Implementar validación de tokens
- [ ] Crear middleware de autenticación

### 4. Backend - API Endpoints (2h)
- [ ] Implementar `POST /api/v1/auth/login`
- [ ] Implementar `POST /api/v1/auth/refresh`
- [ ] Implementar `POST /api/v1/auth/logout`
- [ ] Agregar rate limiting

### 5. Frontend - Pantalla de Login (3h)
- [ ] Crear componente `LoginForm` en React
- [ ] Validación de formulario con react-hook-form
- [ ] Integración con API usando TanStack Query
- [ ] Manejo de errores y mensajes
- [ ] Remember me functionality

### 6. Frontend - Manejo de Sesión (2h)
- [ ] Implementar `AuthContext` con Zustand
- [ ] Almacenamiento seguro de tokens
- [ ] Auto-refresh de tokens
- [ ] Redirección en rutas protegidas

### 7. Testing (3h)
- [ ] Tests unitarios para AuthService
- [ ] Tests de integración para endpoints
- [ ] Tests E2E para flujo de login
- [ ] Tests de seguridad (SQL injection, XSS)

### 8. Seguridad (2h)
- [ ] Implementar CORS correctamente
- [ ] Headers de seguridad (HSTS, CSP)
- [ ] Protección contra CSRF
- [ ] Auditoría de accesos

### 9. Documentación (1h)
- [ ] Documentar API en OpenAPI/Swagger
- [ ] Guía de integración para frontend
- [ ] Documentar flujo de autenticación

## Estimación Total: 22 horas

## Dependencias

- Ninguna (es la primera historia)

## Riesgos

1. **Seguridad de tokens**: Mitigar con rotación automática
2. **Ataques de fuerza bruta**: Implementar captcha después de 3 intentos
3. **Sesión hijacking**: Validar IP y user agent

## Notas de Implementación

```go
// Estructura del JWT payload
type JWTClaims struct {
    UserID       string   `json:"user_id"`
    Email        string   `json:"email"`
    OrgID        string   `json:"org_id"`
    Role         string   `json:"role"`
    Permissions  []string `json:"permissions"`
    jwt.StandardClaims
}

// Respuesta de login
type LoginResponse struct {
    User         UserInfo `json:"user"`
    AccessToken  string   `json:"access_token"`
    RefreshToken string   `json:"refresh_token"`
    ExpiresIn    int      `json:"expires_in"`
    TokenType    string   `json:"token_type"`
}
```

## Mockups Relacionados

- [Login Page](../mocks/front/html/login.html)
- [API Login Response](../mocks/api/auth_login.json)