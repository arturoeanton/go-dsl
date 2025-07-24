# 🎭 Automatización con Playwright - Motor Contable

Este proyecto contiene la automatización de la demostración del Motor Contable con go-dsl usando Playwright.

## 📋 Requisitos

- Node.js 16+ instalado
- El servidor del Motor Contable debe estar corriendo en `http://localhost:3000`

## 🚀 Instalación

```bash
# Instalar dependencias
npm install

# Instalar navegadores de Playwright
npm run install:browsers
```

## 🎮 Ejecutar la Demo

### Demo Interactiva (Recomendado)
```bash
# Ejecuta la demo con navegador visible
npm run demo
```

### Demo Rápida (Solo Chrome)
```bash
# Ejecuta solo en Chrome para mayor velocidad
npm run demo:fast
```

### Demo con Video
```bash
# Genera un video de la demo completa
npm run demo:video
```

## 📁 Estructura del Proyecto

```
auto/
├── package.json           # Configuración de npm
├── playwright.config.js   # Configuración de Playwright
├── README.md             # Este archivo
└── tests/
    └── demo.spec.js      # Script principal de la demo
```

## 🎯 Características de la Demo

La automatización demuestra:

1. **Dashboard y Navegación** - KPIs en tiempo real
2. **POS (Punto de Venta)** - Generación automática de IVA con DSL
3. **Comprobantes** - Gestión con reglas DSL aplicadas
4. **Editor DSL Visual** - Creación y edición de reglas
5. **Plan de Cuentas** - Navegación por el PUC
6. **Workflows** - Aprobaciones automáticas según montos
7. **Resumen Final** - Conclusiones de la demo

## 🎨 Personalización

### Ajustar Tiempos
Edita `DEMO_CONFIG` en `demo.spec.js`:
```javascript
const DEMO_CONFIG = {
  pauseTime: 2000,      // Tiempo entre acciones
  animationTime: 1000,  // Duración de animaciones
};
```

### Agregar Nuevos Tests
Crea nuevos archivos en la carpeta `tests/` siguiendo el patrón de `demo.spec.js`.

## 🐛 Solución de Problemas

### El servidor no está corriendo
```bash
cd ../app
go run main.go
```

### Error de timeout
Aumenta los timeouts en `playwright.config.js`:
```javascript
use: {
  actionTimeout: 20000,
  navigationTimeout: 60000,
}
```

### Navegador no se abre
Reinstala los navegadores:
```bash
npx playwright install --with-deps
```

## 📊 Reportes

Los resultados de las pruebas se guardan en:
- `test-results/` - Capturas y videos de fallos
- `playwright-report/` - Reporte HTML (cuando se usa --reporter=html)

## 🎬 Generar Video Demo

Para crear un video profesional de la demo:
```bash
npm run demo:video
```

El video se guardará en `test-results/` con el nombre del test.

## 💡 Tips

- La demo está diseñada para ser visual y educativa
- Los elementos se resaltan antes de interactuar
- Los mensajes en consola explican cada paso
- Ajusta `slowMo` en la configuración para cambiar velocidad

## 📚 Más Información

- [Documentación de Playwright](https://playwright.dev)
- [Motor Contable README](../README.md)
- [Documentación go-dsl](https://github.com/arturoeanton/go-dsl)