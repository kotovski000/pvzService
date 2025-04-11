package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"pvzService/internal/db"
	"pvzService/internal/handlers"
	"pvzService/internal/models"
	"pvzService/internal/processors"
	"pvzService/internal/repository"
)

func TestFullPVZWorkflowWithRoles(t *testing.T) {
	t.Log("=== Начало комплексного теста рабочего процесса ПВЗ ===")

	// Настройка тестовой БД
	ctx := context.Background()
	t.Log("Инициализация тестовой БД PostgreSQL с использованием Testcontainers...")
	postgresContainer, dsn := setupTestDB(ctx, t)
	defer func() {
		t.Log("Остановка контейнера с тестовой БД...")
		postgresContainer.Terminate(ctx)
	}()

	t.Log("Подключение к тестовой БД...")
	testDB := connectToTestDB(t, dsn)
	defer func() {
		t.Log("Закрытие соединения с тестовой БД...")
		testDB.Close()
	}()

	t.Log("Применение миграций...")
	applyMigrations(t, testDB)

	// Инициализация слоев приложения
	t.Log("Инициализация репозиториев и процессоров...")
	pvzRepo := repository.NewPVZRepository(testDB)
	receptionRepo := repository.NewReceptionRepository(testDB)
	productRepo := repository.NewProductRepository(testDB)

	pvzProcessor := processors.NewPVZProcessor(pvzRepo)
	receptionProcessor := processors.NewReceptionProcessor(receptionRepo)
	productProcessor := processors.NewProductProcessor(productRepo, receptionRepo)

	pvzHandlers := handlers.NewPVZHandlers(pvzProcessor)
	receptionHandlers := handlers.NewReceptionHandlers(receptionProcessor)
	productHandlers := handlers.NewProductHandlers(productProcessor)

	// Создаем Fiber приложение с middleware
	t.Log("Создание Fiber приложения с middleware проверки ролей...")
	app := fiber.New()
	setupRoutesWithRoles(app, pvzHandlers, receptionHandlers, productHandlers)

	// 1. Создание нового ПВЗ
	t.Log("Тест 1: Создание ПВЗ с ролью moderator...")
	pvzID := createPVZAsModerator(t, app)
	assert.NotEmpty(t, pvzID)
	t.Logf("Успешно создан ПВЗ с ID: %s", pvzID)

	// 2. Добавление приёмки
	t.Log("Тест 2: Создание приёмки с ролью employee...")
	receptionID := createReceptionAsEmployee(t, app, pvzID)
	assert.NotEmpty(t, receptionID)
	t.Logf("Успешно создана приёмка с ID: %s", receptionID)

	// 3. Добавление товаров
	t.Logf("Тест 3: Добавление %d товаров с ролью employee...", 50)
	productIDs := addProductsAsEmployee(t, app, pvzID, 50)
	assert.Len(t, productIDs, 50)
	t.Logf("Успешно добавлено %d товаров", len(productIDs))

	// 4. Закрытие приёмки
	t.Log("Тест 4: Закрытие приёмки с ролью employee...")
	closedReception := closeReceptionAsEmployee(t, app, pvzID)
	assert.Equal(t, "close", closedReception.Status)
	assert.NotNil(t, closedReception.ClosedAt)
	t.Logf("Приёмка успешно закрыта: %s", closedReception.ClosedAt.Format(time.RFC3339))

	t.Log("=== Комплексный тест завершен успешно ===")
}

func setupTestDB(ctx context.Context, t *testing.T) (testcontainers.Container, string) {
	t.Log("Настройка контейнера PostgreSQL...")
	req := testcontainers.ContainerRequest{
		Image:        "postgres:13",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "test",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections"),
	}

	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	assert.NoError(t, err)

	host, err := postgresContainer.Host(ctx)
	assert.NoError(t, err)

	port, err := postgresContainer.MappedPort(ctx, "5432")
	assert.NoError(t, err)

	dsn := fmt.Sprintf("postgres://test:test@%s:%s/test?sslmode=disable", host, port.Port())
	t.Logf("БД доступна по DSN: %s", dsn)
	return postgresContainer, dsn
}

func connectToTestDB(t *testing.T, dsn string) *sql.DB {
	t.Logf("Попытка подключения к БД (до 5 попыток)...")
	var testDB *sql.DB
	var err error

	for i := 0; i < 5; i++ {
		testDB, err = db.InitializeTestDB(dsn)
		if err == nil {
			t.Log("Успешное подключение к БД")
			return testDB
		}
		t.Logf("Попытка %d: ошибка подключения: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	t.Fatalf("Не удалось подключиться к БД: %v", err)
	return nil
}

func applyMigrations(t *testing.T, db *sql.DB) {
	t.Log("Применение SQL миграций...")
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS pvz (
			id TEXT PRIMARY KEY,
			registration_date TIMESTAMP NOT NULL DEFAULT NOW(),
			city TEXT NOT NULL
		);
		
		CREATE TABLE IF NOT EXISTS receptions (
			id TEXT PRIMARY KEY,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			pvz_id TEXT REFERENCES pvz(id),
			status TEXT NOT NULL,
			closed_at TIMESTAMP
		);
		
		CREATE TABLE IF NOT EXISTS products (
			id TEXT PRIMARY KEY,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			type TEXT NOT NULL,
			reception_id TEXT REFERENCES receptions(id)
		);
	`)
	assert.NoError(t, err)
	t.Log("Миграции успешно применены")
}

func setupRoutesWithRoles(app *fiber.App,
	pvzHandlers *handlers.PVZHandlers,
	receptionHandlers *handlers.ReceptionHandlers,
	productHandlers *handlers.ProductHandlers) {

	// Middleware для разных ролей
	moderatorOnly := func(c *fiber.Ctx) error {
		user := c.Locals("user").(jwt.MapClaims)
		if user["role"] != "moderator" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "moderator role required",
			})
		}
		return c.Next()
	}

	employeeOnly := func(c *fiber.Ctx) error {
		user := c.Locals("user").(jwt.MapClaims)
		if user["role"] != "employee" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "employee role required",
			})
		}
		return c.Next()
	}

	// Роуты с проверкой ролей
	app.Post("/pvz", setTestRole("moderator"), moderatorOnly, pvzHandlers.CreatePVZHandler())
	app.Post("/receptions", setTestRole("employee"), employeeOnly, receptionHandlers.CreateReceptionHandler())
	app.Post("/products", setTestRole("employee"), employeeOnly, productHandlers.AddProductHandler())
	app.Post("/pvz/:pvzId/close_last_reception", setTestRole("employee"), employeeOnly, receptionHandlers.CloseLastReceptionHandler())
}

func setTestRole(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := jwt.MapClaims{
			"userId": "test-user",
			"role":   role,
		}
		c.Locals("user", claims)
		return c.Next()
	}
}

func createPVZAsModerator(t *testing.T, app *fiber.App) string {
	t.Log("Создание тестового запроса для ПВЗ...")
	pvzReq := models.PVZ{
		City: "Москва",
	}

	reqBody, _ := json.Marshal(pvzReq)
	req := httptest.NewRequest("POST", "/pvz", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	t.Log("Выполнение HTTP-запроса на создание ПВЗ...")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var pvz models.PVZ
	err = json.NewDecoder(resp.Body).Decode(&pvz)
	assert.NoError(t, err)

	return pvz.ID
}

func createReceptionAsEmployee(t *testing.T, app *fiber.App, pvzID string) string {
	t.Log("Формирование запроса на создание приёмки...")
	reqBody := fmt.Sprintf(`{"pvzId": "%s"}`, pvzID)
	req := httptest.NewRequest("POST", "/receptions", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	t.Log("Отправка запроса на создание приёмки...")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var reception models.Reception
	err = json.NewDecoder(resp.Body).Decode(&reception)
	assert.NoError(t, err)

	return reception.ID
}

func addProductsAsEmployee(t *testing.T, app *fiber.App, pvzID string, count int) []string {
	t.Logf("Начало добавления %d товаров...", count)
	productIDs := make([]string, 0, count)
	productTypes := []string{"электроника", "одежда", "обувь"}

	for i := 0; i < count; i++ {
		productType := productTypes[i%len(productTypes)]
		reqBody := fmt.Sprintf(`{"type": "%s", "pvzId": "%s"}`, productType, pvzID)

		req := httptest.NewRequest("POST", "/products", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		t.Logf("Добавление товара %d/%d типа '%s'...", i+1, count, productType)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var product models.Product
		err = json.NewDecoder(resp.Body).Decode(&product)
		assert.NoError(t, err)

		productIDs = append(productIDs, product.ID)
	}

	return productIDs
}

func closeReceptionAsEmployee(t *testing.T, app *fiber.App, pvzID string) models.Reception {
	t.Log("Формирование запроса на закрытие приёмки...")
	req := httptest.NewRequest("POST", fmt.Sprintf("/pvz/%s/close_last_reception", pvzID), nil)

	t.Log("Отправка запроса на закрытие приёмки...")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var reception models.Reception
	err = json.NewDecoder(resp.Body).Decode(&reception)
	assert.NoError(t, err)

	return reception
}
