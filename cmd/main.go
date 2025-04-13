package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"net"
	"net/http"
	"pvzService/internal/config"
	"pvzService/internal/db"
	"pvzService/internal/handlers"
	"pvzService/internal/middleware"
	"pvzService/internal/processors"
	"pvzService/internal/prometheus"
	pb "pvzService/internal/proto"
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

type server struct {
	pb.UnimplementedPVZServiceServer
	db *sql.DB
}

func (s *server) GetPVZList(context.Context, *pb.GetPVZListRequest) (*pb.GetPVZListResponse, error) {

	rows, err := s.db.Query(`
        SELECT 
          p.id, p.registration_date, p.city
        FROM pvz p
        ORDER BY p.registration_date
      `)
	if err != nil {
		return nil, err // ???
	}
	defer rows.Close()

	var pvzList []*pb.PVZ

	for rows.Next() {
		var pvzID sql.NullString
		var pvzRegDate sql.NullTime
		var pvzCity sql.NullString

		err = rows.Scan(
			&pvzID, &pvzRegDate, &pvzCity,
		)

		if err != nil {
			log.Println("Error scanning row:", err)
			continue
		}
		if pvzID.Valid {
			pvz := &pb.PVZ{
				Id:               pvzID.String,
				RegistrationDate: timestamppb.New(pvzRegDate.Time),
				City:             pvzCity.String,
			}

			pvzList = append(pvzList, pvz)
		}
	}

	if err := rows.Err(); err != nil {
		log.Println("Error iterating rows:", err)
	}

	return &pb.GetPVZListResponse{Pvzs: pvzList}, nil
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

	lis, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterPVZServiceServer(s, &server{db: database})
	go func() {
		log.Printf("server listening at %v", lis.Addr())
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	app := fiber.New()

	startMetricsServer()

	app.Use(cors.New())
	app.Use(logger.New(logger.Config{
		Format:     "${time} | ${status} | ${latency} | ${method} ${path}\n",
		TimeFormat: "2006-01-02 15:04:05",
	}))
	app.Use(prometheus.PrometheusMiddleware())

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
