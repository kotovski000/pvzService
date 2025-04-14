package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/wait"

	"pvzService/cmd/app"
	"pvzService/internal/config"
	"pvzService/internal/db"
	"pvzService/internal/models"
)

func TestFullPVZWorkflowWithRoles(t *testing.T) {
	t.Log("=== Начало комплексного теста рабочего процесса ПВЗ с проверкой ролей ===")

	ctx := context.Background()
	t.Log("Инициализация тестовой БД PostgreSQL с использованием Testcontainers...")
	postgresContainer, dsn := setupTestDB(ctx, t)
	defer func() {
		t.Log("Остановка контейнера с тестовой БД...")
		postgresContainer.Terminate(ctx)
	}()

	t.Log("Подключение к тестовой БД...")
	testDB := connectToTestDB(t, dsn)
	defer testDB.Close()

	t.Log("Применение миграций...")
	applyMigrations(t, testDB)

	testCfg := config.Config{
		JWTSecret: "test-secret",
	}

	testApp := app.MakeApp(testDB, testCfg)

	// 1. Создание нового ПВЗ (требуется роль moderator)
	pvzID := createPVZAsModerator(t, testApp, testCfg)
	assert.NotEmpty(t, pvzID)

	// 2. Добавление приёмки (требуется роль employee)
	receptionID := createReceptionAsEmployee(t, testApp, testCfg, pvzID)
	assert.NotEmpty(t, receptionID)

	// 3. Добавление товаров (требуется роль employee)
	productIDs := addProductsAsEmployee(t, testApp, testCfg, pvzID, 50)
	assert.Len(t, productIDs, 50)

	// 4. Закрытие приёмки (требуется роль employee)
	closedReception := closeReceptionAsEmployee(t, testApp, testCfg, pvzID)
	assert.Equal(t, "close", closedReception.Status)
	assert.NotNil(t, closedReception.ClosedAt)

	// 5. Попытка создания ПВЗ с ролью employee (должна завершиться ошибкой)
	tryCreatePVZAsEmployee(t, testApp, testCfg)
}

func generateTokenWithRole(role string, secret string) (string, error) {
	claims := jwt.MapClaims{
		"userId": "test-user",
		"role":   role,
		"exp":    time.Now().Add(time.Hour * 1).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func setupTestDB(ctx context.Context, t *testing.T) (testcontainers.Container, string) {
	t.Log("Настройка контейнера PostgreSQL...")
	req := testcontainers.ContainerRequest{
		Image:        "postgres:13-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "test",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections"),
			wait.ForListeningPort("5432/tcp"),
		),
		LifecycleHooks: []testcontainers.ContainerLifecycleHooks{
			{
				PostStarts: []testcontainers.ContainerHook{
					func(ctx context.Context, container testcontainers.Container) error {
						// Ждем полной инициализации БД
						time.Sleep(2 * time.Second)
						return nil
					},
				},
			},
		},
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
			_, err = testDB.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`)
			if err != nil {
				t.Logf("Ошибка при проверке расширений: %v", err)
				continue
			}
			t.Log("Успешное подключение к БД и проверка расширений")
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

	_, err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`)
	assert.NoError(t, err, "Не удалось подключить расширение pgcrypto")

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL CHECK (role IN ('employee', 'moderator')),
			created_at TIMESTAMP DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS pvz (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			city TEXT NOT NULL CHECK (city IN ('Москва', 'Санкт-Петербург', 'Казань')),
			registration_date TIMESTAMP DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS receptions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			pvz_id UUID REFERENCES pvz(id),
			status TEXT NOT NULL CHECK (status IN ('in_progress', 'close')),
			created_at TIMESTAMP DEFAULT NOW(),
			closed_at TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS products (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			reception_id UUID REFERENCES receptions(id),
			type TEXT NOT NULL CHECK (type IN ('электроника', 'одежда', 'обувь')),
			created_at TIMESTAMP DEFAULT NOW()
		);

		INSERT INTO users (email, password, role) VALUES (
			'moderator@test.com',
			crypt('moderator123', gen_salt('bf')),
			'moderator'
		) ON CONFLICT DO NOTHING;

		INSERT INTO users (email, password, role) VALUES (
			'employee@test.com',
			crypt('employee123', gen_salt('bf')),
			'employee'
		) ON CONFLICT DO NOTHING;
	`)
	assert.NoError(t, err, "Не удалось применить миграции")
	t.Log("Миграции успешно применены")
}

func createPVZAsModerator(t *testing.T, app *fiber.App, cfg config.Config) string {
	token, err := generateTokenWithRole("moderator", cfg.JWTSecret)
	assert.NoError(t, err)

	t.Log("Создание тестового запроса для ПВЗ...")
	pvzReq := models.PVZ{
		City: "Москва",
	}

	reqBody, _ := json.Marshal(pvzReq)
	req := httptest.NewRequest("POST", "/pvz", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	t.Log("Выполнение HTTP-запроса на создание ПВЗ...")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var pvz models.PVZ
	err = json.NewDecoder(resp.Body).Decode(&pvz)
	assert.NoError(t, err)

	return pvz.ID
}

func createReceptionAsEmployee(t *testing.T, app *fiber.App, cfg config.Config, pvzID string) string {
	token, err := generateTokenWithRole("employee", cfg.JWTSecret)
	assert.NoError(t, err)

	t.Log("Формирование запроса на создание приёмки...")
	reqBody := fmt.Sprintf(`{"pvzId": "%s"}`, pvzID)
	req := httptest.NewRequest("POST", "/receptions", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	t.Log("Отправка запроса на создание приёмки...")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var reception models.Reception
	err = json.NewDecoder(resp.Body).Decode(&reception)
	assert.NoError(t, err)

	return reception.ID
}

func addProductsAsEmployee(t *testing.T, app *fiber.App, cfg config.Config, pvzID string, count int) []string {
	token, err := generateTokenWithRole("employee", cfg.JWTSecret)
	assert.NoError(t, err)

	t.Logf("Начало добавления %d товаров...", count)
	productIDs := make([]string, 0, count)
	productTypes := []string{"электроника", "одежда", "обувь"}

	for i := 0; i < count; i++ {
		productType := productTypes[i%len(productTypes)]
		reqBody := fmt.Sprintf(`{"type": "%s", "pvzId": "%s"}`, productType, pvzID)

		req := httptest.NewRequest("POST", "/products", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

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

func closeReceptionAsEmployee(t *testing.T, app *fiber.App, cfg config.Config, pvzID string) models.Reception {
	token, err := generateTokenWithRole("employee", cfg.JWTSecret)
	assert.NoError(t, err)

	t.Log("Формирование запроса на закрытие приёмки...")
	req := httptest.NewRequest("POST", fmt.Sprintf("/pvz/%s/close_last_reception", pvzID), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	t.Log("Отправка запроса на закрытие приёмки...")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var reception models.Reception
	err = json.NewDecoder(resp.Body).Decode(&reception)
	assert.NoError(t, err)

	return reception
}

func tryCreatePVZAsEmployee(t *testing.T, app *fiber.App, cfg config.Config) {
	token, err := generateTokenWithRole("employee", cfg.JWTSecret)
	assert.NoError(t, err)

	t.Log("Попытка создания ПВЗ с ролью employee (должна завершиться ошибкой)...")
	pvzReq := models.PVZ{
		City: "Санкт-Петербург",
	}

	reqBody, _ := json.Marshal(pvzReq)
	req := httptest.NewRequest("POST", "/pvz", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)

	var errorResp models.Error
	err = json.NewDecoder(resp.Body).Decode(&errorResp)
	assert.NoError(t, err)
	assert.Contains(t, errorResp.Message, "Insufficient role")
}
