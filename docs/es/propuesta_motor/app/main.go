package main

import (
	"log"
	"motor-contable-poc/internal/database"
	"motor-contable-poc/internal/handlers"
	"motor-contable-poc/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	
	_ "motor-contable-poc/docs" // Import docs package for Swagger
)

// @title Motor Contable POC API
// @version 1.0
// @description API para el Proof of Concept del Motor Contable Colombiano
// @termsOfService http://swagger.io/terms/
// @contact.name Soporte API
// @contact.email soporte@motorcontable.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:3000
// @BasePath /api/v1
// @schemes http https
func main() {
	// Inicializar base de datos SQLite
	err := database.InitDatabase()
	if err != nil {
		log.Fatal("Error inicializando base de datos:", err)
	}
	defer database.CloseDatabase()

	// Cargar datos de demostración
	// TODO: En el futuro, este proceso usaría go-dsl para generar
	// datos de prueba más realistas basados en plantillas contables
	err = database.SeedData()
	if err != nil {
		log.Printf("Advertencia: Error cargando datos de demo: %v", err)
	}

	// Inicializar aplicación Fiber
	app := fiber.New(fiber.Config{
		AppName:      "Motor Contable POC v1.0",
		ServerHeader: "Motor-Contable-POC",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success":   false,
				"error":     err.Error(),
				"timestamp": fiber.Config{}.JSONEncoder,
			})
		},
	})

	// Middleware
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} (${latency})\n",
	}))
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Servir archivos estáticos (frontend)
	app.Static("/", "./static")

	// Swagger UI
	app.Get("/swagger/*", swagger.HandlerDefault)
	
	// Serve Swagger JSON file directly for API tools
	app.Get("/doc.json", func(c *fiber.Ctx) error {
		return c.SendFile("./docs/swagger.json")
	})

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		// Verificar estado de la base de datos
		err := database.DatabaseHealthCheck()
		if err != nil {
			return c.Status(503).JSON(fiber.Map{
				"status": "unhealthy",
				"database": "disconnected",
				"error": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"status":   "healthy",
			"database": "connected",
			"version":  "1.0.0",
			"service":  "motor-contable-poc",
		})
	})

	// Rutas API v1
	api := app.Group("/api/v1")

	// Inicializar handlers con base de datos
	db := database.GetDB()
	
	// Registrar rutas de handlers
	orgHandler := handlers.NewOrganizationHandler(db)
	orgHandler.RegisterRoutes(api)
	
	// Crear el servicio de voucher primero para poder obtener el dslEngine
	voucherService := services.NewVoucherService(db)
	dslEngine := voucherService.GetDSLEngine()
	
	// Crear voucher handler usando el servicio compartido
	voucherHandler := handlers.NewVoucherHandlerWithService(voucherService)
	voucherHandler.RegisterRoutes(api)

	// Dashboard handler para estadísticas y actividad
	dashboardHandler := handlers.NewDashboardHandler(db)
	dashboardHandler.RegisterRoutes(api)

	// Accounts handler para plan de cuentas
	accountsHandler := handlers.NewAccountsHandler(db)
	accountsHandler.RegisterRoutes(api)

	// Journal entries handler para asientos contables
	journalEntriesHandler := handlers.NewJournalEntriesHandler(db)
	journalEntriesHandler.RegisterRoutes(api)

	// DSL handler para plantillas DSL
	dslHandler := handlers.NewDSLHandler(db, dslEngine)
	dslHandler.RegisterRoutes(api)

	// Template handler para templates de asientos
	// Use simple handler for SQLite
	templateHandler := handlers.NewTemplateSimpleHandler(db)
	templateHandler.RegisterRoutes(api)
	
	// Debug handler for templates
	debugHandler := handlers.NewTemplateDebugHandler()
	debugHandler.RegisterRoutes(api)

	// TODO: Registrar handlers adicionales para:
	// - Terceros (third-parties)
	// - Períodos contables (periods)
	// - Reportes (reports)
	// - Auditoría (audit)

	// Ruta raíz
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Motor Contable POC API",
			"version": "1.0.0",
			"docs":    "/swagger/",
			"health":  "/health",
			"apis":    "/api/v1/",
		})
	})

	// Iniciar servidor
	log.Println("🚀 Motor Contable POC iniciando en puerto 3000")
	log.Println("📚 Documentación Swagger: http://localhost:3000/swagger/")
	log.Println("💾 Base de datos SQLite: db_contable.db")
	log.Println("🏥 Health Check: http://localhost:3000/health")
	
	log.Fatal(app.Listen(":3000"))
}
