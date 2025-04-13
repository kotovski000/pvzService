package app

import (
	"database/sql"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"pvzService/internal/config"
	"pvzService/internal/handlers"
	"pvzService/internal/middleware"
	"pvzService/internal/processors"
	"pvzService/internal/prometheus"
	"pvzService/internal/repository"
)

func MakeApp(database *sql.DB, cfg config.Config) *fiber.App {
	// Initialize repositories
	authRepo := repository.NewAuthRepository(database)
	pvzRepo := repository.NewPVZRepository(database)
	receptionRepo := repository.NewReceptionRepository(database)
	productRepo := repository.NewProductRepository(database)

	// Initialize processors
	authProcessor := processors.NewAuthProcessor(authRepo)
	pvzProcessor := processors.NewPVZProcessor(pvzRepo)
	receptionProcessor := processors.NewReceptionProcessor(receptionRepo)
	productProcessor := processors.NewProductProcessor(productRepo, receptionRepo)

	// Initialize handlers
	authHandlers := handlers.NewAuthHandlers(authProcessor, cfg.JWTSecret)
	pvzHandlers := handlers.NewPVZHandlers(pvzProcessor)
	receptionHandlers := handlers.NewReceptionHandlers(receptionProcessor)
	productHandlers := handlers.NewProductHandlers(productProcessor)

	app := fiber.New()

	// Common middleware
	app.Use(cors.New())
	app.Use(logger.New(logger.Config{
		Format:     "${time} | ${status} | ${latency} | ${method} ${path}\n",
		TimeFormat: "2006-01-02 15:04:05",
	}))
	app.Use(prometheus.PrometheusMiddleware())

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Public Routes
	app.Post("/dummyLogin", authHandlers.DummyLoginHandler())
	app.Post("/register", authHandlers.RegisterHandler())
	app.Post("/login", authHandlers.LoginHandler())

	// Protected Routes
	api := app.Group("/")
	api.Use(middleware.AuthMiddleware(cfg.JWTSecret))

	// Routes configuration with role checks
	api.Post("/pvz", middleware.CheckRole("moderator"), pvzHandlers.CreatePVZHandler())
	api.Get("/pvz", middleware.CheckRole("employee", "moderator"), pvzHandlers.GetPVZListHandler())
	api.Post("/receptions", middleware.CheckRole("employee"), receptionHandlers.CreateReceptionHandler())
	api.Post("/products", middleware.CheckRole("employee"), productHandlers.AddProductHandler())
	api.Post("/pvz/:pvzId/close_last_reception", middleware.CheckRole("employee"), receptionHandlers.CloseLastReceptionHandler())
	api.Post("/pvz/:pvzId/delete_last_product", middleware.CheckRole("employee"), productHandlers.DeleteLastProductHandler())

	return app
}
