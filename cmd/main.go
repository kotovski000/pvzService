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

	app := fiber.New()

	app.Use(cors.New())
	app.Use(logger.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Public Routes
	app.Post("/dummyLogin", handlers.DummyLoginHandler(database, cfg.JWTSecret))
	app.Post("/register", handlers.RegisterHandler(database, cfg.JWTSecret))
	app.Post("/login", handlers.LoginHandler(database, cfg.JWTSecret))

	// Protected Routes
	//api := app.Group("/", middleware.AuthMiddleware(cfg.JWTSecret))
	//
	//api.Post("/pvz", middleware.CheckRole("moderator"), handlers.CreatePVZHandler(database))
	//api.Get("/pvz", middleware.CheckRole("employee", "moderator"), handlers.GetPVZListHandler(database))
	//api.Post("/receptions", middleware.CheckRole("employee"), handlers.CreateReceptionHandler(database))
	//api.Post("/products", middleware.CheckRole("employee"), handlers.AddProductHandler(database))
	//api.Post("/pvz/:pvzId/close_last_reception", middleware.CheckRole("employee"), handlers.CloseLastReceptionHandler(database))
	//api.Post("/pvz/:pvzId/delete_last_product", middleware.CheckRole("employee"), handlers.DeleteLastProductHandler(database))

	log.Printf("Server listening on port %s", cfg.Port)
	log.Fatal(app.Listen(fmt.Sprintf("0.0.0.0:%s", cfg.Port)))
}
