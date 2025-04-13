package repository

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"pvzService/internal/models"
)

func TestCreateReception(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewReceptionRepository(db)

	t.Run("successful creation", func(t *testing.T) {
		pvzID := uuid.New().String()
		expectedID := uuid.New()

		mock.ExpectExec("INSERT INTO receptions").
			WithArgs(expectedID.String(), pvzID, "in_progress", sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		id, err := repo.CreateReception(pvzID, func() uuid.UUID { return expectedID })

		assert.NoError(t, err)
		assert.Equal(t, expectedID.String(), id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		pvzID := uuid.New().String()
		expectedError := errors.New("database error")

		mock.ExpectExec("INSERT INTO receptions").
			WithArgs(sqlmock.AnyArg(), pvzID, "in_progress", sqlmock.AnyArg()).
			WillReturnError(expectedError)

		_, err := repo.CreateReception(pvzID, uuid.New)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetReceptionByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewReceptionRepository(db)

	t.Run("successful retrieval", func(t *testing.T) {
		receptionID := uuid.New().String()
		pvzID := uuid.New().String()
		createdAt := time.Now()
		closedAt := createdAt.Add(time.Hour)

		expectedReception := models.Reception{
			ID:       receptionID,
			DateTime: createdAt,
			PvzId:    pvzID,
			Status:   "close",
			ClosedAt: &closedAt,
		}

		rows := sqlmock.NewRows([]string{"id", "created_at", "pvz_id", "status", "closed_at"}).
			AddRow(expectedReception.ID, expectedReception.DateTime, expectedReception.PvzId, expectedReception.Status, expectedReception.ClosedAt)

		mock.ExpectQuery("SELECT id, created_at, pvz_id, status, closed_at FROM receptions WHERE id = \\$1").
			WithArgs(receptionID).
			WillReturnRows(rows)

		reception, err := repo.GetReceptionByID(receptionID)

		assert.NoError(t, err)
		assert.Equal(t, expectedReception, reception)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		receptionID := uuid.New().String()

		mock.ExpectQuery("SELECT id, created_at, pvz_id, status, closed_at FROM receptions WHERE id = \\$1").
			WithArgs(receptionID).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.GetReceptionByID(receptionID)

		assert.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		receptionID := uuid.New().String()
		expectedError := errors.New("database error")

		mock.ExpectQuery("SELECT id, created_at, pvz_id, status, closed_at FROM receptions WHERE id = \\$1").
			WithArgs(receptionID).
			WillReturnError(expectedError)

		_, err := repo.GetReceptionByID(receptionID)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetOpenReception(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewReceptionRepository(db)

	t.Run("successful retrieval", func(t *testing.T) {
		pvzID := uuid.New().String()
		receptionID := uuid.New().String()
		createdAt := time.Now()

		expectedReception := models.Reception{
			ID:       receptionID,
			DateTime: createdAt,
			PvzId:    pvzID,
			Status:   "in_progress",
			ClosedAt: nil,
		}

		rows := sqlmock.NewRows([]string{"id", "created_at", "pvz_id", "status", "closed_at"}).
			AddRow(expectedReception.ID, expectedReception.DateTime, expectedReception.PvzId, expectedReception.Status, nil)

		mock.ExpectQuery("SELECT id, created_at, pvz_id, status, closed_at FROM receptions WHERE pvz_id = \\$1 AND status = 'in_progress'").
			WithArgs(pvzID).
			WillReturnRows(rows)

		reception, err := repo.GetOpenReception(pvzID)

		assert.NoError(t, err)
		assert.Equal(t, expectedReception, reception)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		pvzID := uuid.New().String()

		mock.ExpectQuery("SELECT id, created_at, pvz_id, status, closed_at FROM receptions WHERE pvz_id = \\$1 AND status = 'in_progress'").
			WithArgs(pvzID).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.GetOpenReception(pvzID)

		assert.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCloseReception(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewReceptionRepository(db)

	t.Run("successful close", func(t *testing.T) {
		receptionID := uuid.New().String()
		closeTime := time.Now()

		mock.ExpectExec("UPDATE receptions SET status = 'close', closed_at = \\$1 WHERE id = \\$2").
			WithArgs(closeTime, receptionID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.CloseReception(receptionID, closeTime)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no rows affected", func(t *testing.T) {
		receptionID := uuid.New().String()
		closeTime := time.Now()

		mock.ExpectExec("UPDATE receptions SET status = 'close', closed_at = \\$1 WHERE id = \\$2").
			WithArgs(closeTime, receptionID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.CloseReception(receptionID, closeTime)

		assert.NoError(t, err) // Or assert.Error if you want to handle this case as an error
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		receptionID := uuid.New().String()
		closeTime := time.Now()
		expectedError := errors.New("database error")

		mock.ExpectExec("UPDATE receptions SET status = 'close', closed_at = \\$1 WHERE id = \\$2").
			WithArgs(closeTime, receptionID).
			WillReturnError(expectedError)

		err := repo.CloseReception(receptionID, closeTime)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestHasOpenReception(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewReceptionRepository(db)

	t.Run("has open reception", func(t *testing.T) {
		pvzID := uuid.New().String()

		rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)

		mock.ExpectQuery("SELECT EXISTS \\(SELECT 1 FROM receptions WHERE pvz_id = \\$1 AND status = 'in_progress'\\)").
			WithArgs(pvzID).
			WillReturnRows(rows)

		exists, err := repo.HasOpenReception(pvzID)

		assert.NoError(t, err)
		assert.True(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no open reception", func(t *testing.T) {
		pvzID := uuid.New().String()

		rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)

		mock.ExpectQuery("SELECT EXISTS \\(SELECT 1 FROM receptions WHERE pvz_id = \\$1 AND status = 'in_progress'\\)").
			WithArgs(pvzID).
			WillReturnRows(rows)

		exists, err := repo.HasOpenReception(pvzID)

		assert.NoError(t, err)
		assert.False(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		pvzID := uuid.New().String()
		expectedError := errors.New("database error")

		mock.ExpectQuery("SELECT EXISTS \\(SELECT 1 FROM receptions WHERE pvz_id = \\$1 AND status = 'in_progress'\\)").
			WithArgs(pvzID).
			WillReturnError(expectedError)

		_, err := repo.HasOpenReception(pvzID)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
