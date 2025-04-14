package repository

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestAuthRepository_CreateUser_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewAuthRepository(db)

	mock.ExpectExec("INSERT INTO users").
		WithArgs(sqlmock.AnyArg(), "test@example.com", sqlmock.AnyArg(), "employee").
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err = repo.CreateUser("test@example.com", "hashedpassword", "employee")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_CreateUser_EmailExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewAuthRepository(db)

	mock.ExpectExec("INSERT INTO users").
		WithArgs(sqlmock.AnyArg(), "exists@example.com", sqlmock.AnyArg(), "employee").
		WillReturnError(errors.New("email already exists"))

	_, err = repo.CreateUser("exists@example.com", "hashedpassword", "employee")
	assert.Error(t, err)
	assert.Equal(t, "email already exists", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_FindUserByEmail_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewAuthRepository(db)

	expectedID := "user123"
	expectedPassword := "hashedpassword"
	expectedRole := "employee"

	mock.ExpectQuery("SELECT id, password, role FROM users WHERE email =").
		WithArgs("test@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "password", "role"}).
			AddRow(expectedID, expectedPassword, expectedRole))

	id, password, role, err := repo.FindUserByEmail("test@example.com")
	assert.NoError(t, err)
	assert.Equal(t, expectedID, id)
	assert.Equal(t, expectedPassword, password)
	assert.Equal(t, expectedRole, role)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_FindUserByEmail_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewAuthRepository(db)

	mock.ExpectQuery("SELECT id, password, role FROM users WHERE email =").
		WithArgs("nonexistent@example.com").
		WillReturnError(sql.ErrNoRows)

	_, _, _, err = repo.FindUserByEmail("nonexistent@example.com")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, sql.ErrNoRows))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_FindUserByRole_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewAuthRepository(db)

	expectedID := "user123"

	mock.ExpectQuery("SELECT id FROM users WHERE role =").
		WithArgs("employee").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(expectedID))

	id, err := repo.FindUserByRole("employee")
	assert.NoError(t, err)
	assert.Equal(t, expectedID, id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_FindUserByRole_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewAuthRepository(db)

	mock.ExpectQuery("SELECT id FROM users WHERE role =").
		WithArgs("moderator").
		WillReturnError(sql.ErrNoRows)

	_, err = repo.FindUserByRole("moderator")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, sql.ErrNoRows))
	assert.NoError(t, mock.ExpectationsWereMet())
}
