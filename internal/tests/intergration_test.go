package tests

//
//import (
//	"bytes"
//	"context"
//	"database/sql"
//	"encoding/json"
//	"fmt"
//	"github.com/golang-jwt/jwt/v5"
//	"net/http"
//	"net/http/httptest"
//	"strings"
//	"testing"
//	"time"
//
//	"github.com/gofiber/fiber/v2"
//	"github.com/stretchr/testify/assert"
//	"github.com/testcontainers/testcontainers-go"
//	"github.com/testcontainers/testcontainers-go/wait"
//
//	"pvzService/cmd/app"
//	"pvzService/internal/config"
//	"pvzService/internal/db"
//	"pvzService/internal/models"
//)
//
//func TestFullPVZWorkflowWithRoles(t *testing.T) {
//	t.Log("=== Начало комплексного теста рабочего процесса ПВЗ ===")
//
//	ctx := context.Background()
//	t.Log("Инициализация тестовой БД PostgreSQL с использованием Testcontainers...")
//	postgresContainer, dsn := setupTestDB(ctx, t)
//	defer func() {
//		t.Log("Остановка контейнера с тестовой БД...")
//		postgresContainer.Terminate(ctx)
//	}()
//
//	t.Log("Подключение к тестовой БД...")
//	testDB := connectToTestDB(t, dsn)
//	defer func() {
//		t.Log("Закрытие соединения с тестовой БД...")
//		testDB.Close()
//	}()
//
//	t.Log("Применение миграций...")
//	applyMigrations(t, testDB)
//
//	// Создаем тестовую конфигурацию
//	testCfg := config.Config{
//		JWTSecret: "test-secret",
//	}
//
//	// Создаем приложение в тестовом режиме
//	testApp := app.MakeApp(testDB, testCfg, true)
//	setupTestRoles(testApp)
//
//	// 1. Создание нового ПВЗ
//	pvzID := createPVZAsModerator(t, testApp)
//	assert.NotEmpty(t, pvzID)
//
//	// 2. Добавление приёмки
//	receptionID := createReceptionAsEmployee(t, testApp, pvzID)
//	assert.NotEmpty(t, receptionID)
//
//	// 3. Добавление товаров
//	productIDs := addProductsAsEmployee(t, testApp, pvzID, 50)
//	assert.Len(t, productIDs, 50)
//
//	// 4. Закрытие приёмки
//	closedReception := closeReceptionAsEmployee(t, testApp, pvzID)
//	assert.Equal(t, "close", closedReception.Status)
//	assert.NotNil(t, closedReception.ClosedAt)
//}
//
//func setupTestRoles(app *fiber.App) {
//	app.Use(func(c *fiber.Ctx) error {
//		claims := jwt.MapClaims{
//			"userId": "test-user",
//			"role":   "moderator",
//		}
//		c.Locals("user", claims)
//		return c.Next()
//	})
//}
//
//func setupTestDB(ctx context.Context, t *testing.T) (testcontainers.Container, string) {
//	t.Log("Настройка контейнера PostgreSQL...")
//	req := testcontainers.ContainerRequest{
//		Image:        "postgres:13",
//		ExposedPorts: []string{"5432/tcp"},
//		Env: map[string]string{
//			"POSTGRES_USER":     "test",
//			"POSTGRES_PASSWORD": "test",
//			"POSTGRES_DB":       "test",
//		},
//		WaitingFor: wait.ForLog("database system is ready to accept connections"),
//	}
//
//	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
//		ContainerRequest: req,
//		Started:          true,
//	})
//	assert.NoError(t, err)
//
//	host, err := postgresContainer.Host(ctx)
//	assert.NoError(t, err)
//
//	port, err := postgresContainer.MappedPort(ctx, "5432")
//	assert.NoError(t, err)
//
//	dsn := fmt.Sprintf("postgres://test:test@%s:%s/test?sslmode=disable", host, port.Port())
//	t.Logf("БД доступна по DSN: %s", dsn)
//	return postgresContainer, dsn
//}
//
//func connectToTestDB(t *testing.T, dsn string) *sql.DB {
//	t.Logf("Попытка подключения к БД (до 5 попыток)...")
//	var testDB *sql.DB
//	var err error
//
//	for i := 0; i < 5; i++ {
//		testDB, err = db.InitializeTestDB(dsn)
//		if err == nil {
//			t.Log("Успешное подключение к БД")
//			return testDB
//		}
//		t.Logf("Попытка %d: ошибка подключения: %v", i+1, err)
//		time.Sleep(2 * time.Second)
//	}
//	t.Fatalf("Не удалось подключиться к БД: %v", err)
//	return nil
//}
//
//func applyMigrations(t *testing.T, db *sql.DB) {
//	t.Log("Применение SQL миграций...")
//	_, err := db.Exec(`
//		CREATE TABLE IF NOT EXISTS pvz (
//			id TEXT PRIMARY KEY,
//			registration_date TIMESTAMP NOT NULL DEFAULT NOW(),
//			city TEXT NOT NULL
//		);
//
//		CREATE TABLE IF NOT EXISTS receptions (
//			id TEXT PRIMARY KEY,
//			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
//			pvz_id TEXT REFERENCES pvz(id),
//			status TEXT NOT NULL,
//			closed_at TIMESTAMP
//		);
//
//		CREATE TABLE IF NOT EXISTS products (
//			id TEXT PRIMARY KEY,
//			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
//			type TEXT NOT NULL,
//			reception_id TEXT REFERENCES receptions(id)
//		);
//	`)
//	assert.NoError(t, err)
//	t.Log("Миграции успешно применены")
//}
//
//func createPVZAsModerator(t *testing.T, app *fiber.App) string {
//	t.Log("Создание тестового запроса для ПВЗ...")
//	pvzReq := models.PVZ{
//		City: "Москва",
//	}
//
//	reqBody, _ := json.Marshal(pvzReq)
//	req := httptest.NewRequest("POST", "/pvz", bytes.NewReader(reqBody))
//	req.Header.Set("Content-Type", "application/json")
//
//	t.Log("Выполнение HTTP-запроса на создание ПВЗ...")
//	resp, err := app.Test(req)
//	assert.NoError(t, err)
//	assert.Equal(t, http.StatusCreated, resp.StatusCode)
//
//	var pvz models.PVZ
//	err = json.NewDecoder(resp.Body).Decode(&pvz)
//	assert.NoError(t, err)
//
//	return pvz.ID
//}
//
//func createReceptionAsEmployee(t *testing.T, app *fiber.App, pvzID string) string {
//	t.Log("Формирование запроса на создание приёмки...")
//	reqBody := fmt.Sprintf(`{"pvzId": "%s"}`, pvzID)
//	req := httptest.NewRequest("POST", "/receptions", strings.NewReader(reqBody))
//	req.Header.Set("Content-Type", "application/json")
//
//	t.Log("Отправка запроса на создание приёмки...")
//	resp, err := app.Test(req)
//	assert.NoError(t, err)
//	assert.Equal(t, http.StatusCreated, resp.StatusCode)
//
//	var reception models.Reception
//	err = json.NewDecoder(resp.Body).Decode(&reception)
//	assert.NoError(t, err)
//
//	return reception.ID
//}
//
//func addProductsAsEmployee(t *testing.T, app *fiber.App, pvzID string, count int) []string {
//	t.Logf("Начало добавления %d товаров...", count)
//	productIDs := make([]string, 0, count)
//	productTypes := []string{"электроника", "одежда", "обувь"}
//
//	for i := 0; i < count; i++ {
//		productType := productTypes[i%len(productTypes)]
//		reqBody := fmt.Sprintf(`{"type": "%s", "pvzId": "%s"}`, productType, pvzID)
//
//		req := httptest.NewRequest("POST", "/products", strings.NewReader(reqBody))
//		req.Header.Set("Content-Type", "application/json")
//
//		t.Logf("Добавление товара %d/%d типа '%s'...", i+1, count, productType)
//		resp, err := app.Test(req)
//		assert.NoError(t, err)
//		assert.Equal(t, http.StatusCreated, resp.StatusCode)
//
//		var product models.Product
//		err = json.NewDecoder(resp.Body).Decode(&product)
//		assert.NoError(t, err)
//
//		productIDs = append(productIDs, product.ID)
//	}
//
//	return productIDs
//}
//
//func closeReceptionAsEmployee(t *testing.T, app *fiber.App, pvzID string) models.Reception {
//	t.Log("Формирование запроса на закрытие приёмки...")
//	req := httptest.NewRequest("POST", fmt.Sprintf("/pvz/%s/close_last_reception", pvzID), nil)
//
//	t.Log("Отправка запроса на закрытие приёмки...")
//	resp, err := app.Test(req)
//	assert.NoError(t, err)
//	assert.Equal(t, http.StatusOK, resp.StatusCode)
//
//	var reception models.Reception
//	err = json.NewDecoder(resp.Body).Decode(&reception)
//	assert.NoError(t, err)
//
//	return reception
//}
