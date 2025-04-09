package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"log"

	"pvzService/internal/config"
	"pvzService/internal/db"
	"pvzService/internal/handlers"
	"pvzService/internal/middleware"
)

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

	handler := handlers.NewHandler(database)

	app := fiber.New()

	app.Use(cors.New())
	app.Use(logger.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Public Routes
	app.Post("/dummyLogin", handler.DummyLoginHandler(cfg.JWTSecret))
	app.Post("/register", handler.RegisterHandler(cfg.JWTSecret))
	app.Post("/login", handler.LoginHandler(cfg.JWTSecret))

	// Protected Routes
	api := app.Group("/", middleware.AuthMiddleware(cfg.JWTSecret))

	api.Post("/pvz", middleware.CheckRole("moderator"), handler.CreatePVZHandler())
	api.Get("/pvz", middleware.CheckRole("employee", "moderator"), handler.GetPVZListHandler())
	api.Post("/receptions", middleware.CheckRole("employee"), handler.CreateReceptionHandler())
	api.Post("/products", middleware.CheckRole("employee"), handler.AddProductHandler())
	api.Post("/pvz/:pvzId/close_last_reception", middleware.CheckRole("employee"), handler.CloseLastReceptionHandler())
	api.Post("/pvz/:pvzId/delete_last_product", middleware.CheckRole("employee"), handler.DeleteLastProductHandler())

	log.Printf("Server listening on port %s", cfg.Port)
	log.Fatal(app.Listen(fmt.Sprintf(":%s", cfg.Port)))
}
