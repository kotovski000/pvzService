package repository

import (
	"database/sql"
)

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) GetUserIDByRole(role string) (string, error) {
	var userID string
	err := r.DB.QueryRow("SELECT id FROM users WHERE role = $1 LIMIT 1", role).Scan(&userID)
	return userID, err
}

func (r *Repository) CreateUser(userID string, email string, hashedPassword string, role string) error {
	_, err := r.DB.Exec(
		"INSERT INTO users (id, email, password, role) VALUES ($1, $2, $3, $4)",
		userID, email, hashedPassword, role,
	)
	return err
}

func (r *Repository) GetUserByEmail(email string) (string, string, string, error) {
	var userID, hashedPassword, role string
	err := r.DB.QueryRow(
		"SELECT id, password, role FROM users WHERE email = $1",
		email,
	).Scan(&userID, &hashedPassword, &role)
	return userID, hashedPassword, role, err
}
