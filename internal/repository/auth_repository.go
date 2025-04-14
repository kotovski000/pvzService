package repository

import (
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type AuthRepository interface {
	CreateUser(email, hashedPassword, role string) (string, error)
	FindUserByEmail(email string) (string, string, string, error)
	FindUserByRole(role string) (string, error)
}

type AuthRepositoryImpl struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) AuthRepository {
	return &AuthRepositoryImpl{db: db}
}

func (r *AuthRepositoryImpl) CreateUser(email, hashedPassword, role string) (string, error) {
	userID := uuid.New().String()
	_, err := r.db.Exec(
		"INSERT INTO users (id, email, password, role) VALUES ($1, $2, $3, $4)",
		userID, email, hashedPassword, role,
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return "", errors.New("email already exists")
		}
		return "", err
	}
	return userID, nil
}

func (r *AuthRepositoryImpl) FindUserByEmail(email string) (string, string, string, error) {
	var userID, hashedPassword, role string
	err := r.db.QueryRow(
		"SELECT id, password, role FROM users WHERE email = $1",
		email,
	).Scan(&userID, &hashedPassword, &role)
	return userID, hashedPassword, role, err
}

func (r *AuthRepositoryImpl) FindUserByRole(role string) (string, error) {
	var userID string
	err := r.db.QueryRow("SELECT id FROM users WHERE role = $1 LIMIT 1", role).Scan(&userID)
	return userID, err
}
