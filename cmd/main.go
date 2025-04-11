package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"log"
	"net/http"

	"pvzService/internal/config"
	"pvzService/internal/db"
	"pvzService/internal/handlers"
	"pvzService/internal/middleware"
	"pvzService/internal/processors"
	"pvzService/internal/repository"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func startMetricsServer() {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Println("Starting metrics server on :9000")
		if err := http.ListenAndServe(":9000", nil); err != nil {
			log.Printf("Metrics server error: %v", err)
		}
	}()
}

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	cfg := config.LoadConfig()

	database, err := db.InitializeDB(cfg.DBDSN)
	if err != nil {
		log.Fatal("Failed to initialize DB:", err)
	}
	defer database.Close()

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

	// Запускаем сервер метрик
	startMetricsServer()

	app.Use(cors.New())
	app.Use(logger.New())
	app.Use(middleware.PrometheusMiddleware()) // Добавляем Prometheus middleware

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Public Routes
	app.Post("/dummyLogin", authHandlers.DummyLoginHandler())
	app.Post("/register", authHandlers.RegisterHandler())
	app.Post("/login", authHandlers.LoginHandler())

	// Protected Routes
	api := app.Group("/", middleware.AuthMiddleware(cfg.JWTSecret))

	api.Post("/pvz", middleware.CheckRole("moderator"), pvzHandlers.CreatePVZHandler())
	api.Get("/pvz", middleware.CheckRole("employee", "moderator"), pvzHandlers.GetPVZListHandler())
	api.Post("/receptions", middleware.CheckRole("employee"), receptionHandlers.CreateReceptionHandler())
	api.Post("/products", middleware.CheckRole("employee"), productHandlers.AddProductHandler())
	api.Post("/pvz/:pvzId/close_last_reception", middleware.CheckRole("employee"), receptionHandlers.CloseLastReceptionHandler())
	api.Post("/pvz/:pvzId/delete_last_product", middleware.CheckRole("employee"), productHandlers.DeleteLastProductHandler())

	log.Printf("Server listening on port %s", cfg.Port)
	log.Fatal(app.Listen(fmt.Sprintf("0.0.0.0:%s", cfg.Port)))
}
