package processors

import (
	"database/sql"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"pvzService/internal/repository"
	"time"
)

type Processor struct {
	Repo *repository.Repository
}

func NewProcessor(db *sql.DB) *Processor {
	return &Processor{
		Repo: repository.NewRepository(db),
	}
}

// --- Authentication ---

func (p *Processor) GenerateToken(userID, role string, secret string) (string, error) {
	claims := jwt.MapClaims{
		"userId": userID,
		"role":   role,
		"exp":    time.Now().Add(time.Hour * 24).Unix(),
		"iat":    time.Now().Unix(),
		"nbf":    time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func (p *Processor) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func (p *Processor) ComparePassword(hashedPassword string, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err
}

// --- Processors ---

func (p *Processor) DummyLogin(role string, secret string) (string, error) {
	if role != "employee" && role != "moderator" {
		return "", errors.New("Invalid role")
	}

	userID, err := p.Repo.GetUserIDByRole(role)

	if errors.Is(err, sql.ErrNoRows) {
		// Hash the password for the dummy user
		hashedPassword, err := p.HashPassword("password")
		if err != nil {
			return "", errors.New("Failed to create dummy user")
		}

		userID = uuid.New().String()
		err = p.Repo.CreateUser(userID, "dummy@example.com", hashedPassword, role)

		if err != nil {
			return "", errors.New("Failed to create dummy user")
		}
	} else if err != nil {
		return "", errors.New("Database error")
	}

	token, err := p.GenerateToken(userID, role, secret)
	if err != nil {
		return "", errors.New("Failed to generate token")
	}

	return token, nil
}

func (p *Processor) Register(email string, password string, role string, secret string) (string, error) {
	if role != "employee" && role != "moderator" {
		return "", errors.New("Invalid role")
	}

	hashedPassword, err := p.HashPassword(password)
	if err != nil {
		return "", errors.New("Failed to process password")
	}

	userID := uuid.New().String()
	err = p.Repo.CreateUser(userID, email, hashedPassword, role)

	if err != nil {
		return "", errors.New("Failed to create user")
	}

	token, err := p.GenerateToken(userID, role, secret)
	if err != nil {
		return "", errors.New("Failed to generate token")
	}

	return token, nil
}

func (p *Processor) Login(email string, password string, secret string) (string, error) {
	userID, hashedPassword, role, err := p.Repo.GetUserByEmail(email)

	if err != nil {
		return "", errors.New("Invalid credentials")
	}

	err = p.ComparePassword(hashedPassword, password)

	if err != nil {
		return "", errors.New("Invalid credentials")
	}

	token, err := p.GenerateToken(userID, role, secret)
	if err != nil {
		return "", errors.New("Failed to generate token")
	}

	return token, nil
}
