package repository

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPVZRepository_CreatePVZ(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPVZRepository(db)
	pvzID := uuid.NewString()
	now := time.Now()

	// Тест успешного создания
	t.Run("success", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO pvz").
			WithArgs(pvzID, "Москва").
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectQuery("SELECT id, registration_date, city FROM pvz WHERE id =").
			WithArgs(pvzID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "registration_date", "city"}).
				AddRow(pvzID, now, "Москва"))

		pvz, err := repo.CreatePVZ("Москва", func() uuid.UUID {
			return uuid.MustParse(pvzID)
		})

		assert.NoError(t, err)
		assert.Equal(t, pvzID, pvz.ID)
		assert.Equal(t, "Москва", pvz.City)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	// Тест ошибки при вставке
	t.Run("insert error", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO pvz").
			WithArgs(pvzID, "Москва").
			WillReturnError(sql.ErrConnDone)

		_, err := repo.CreatePVZ("Москва", func() uuid.UUID {
			return uuid.MustParse(pvzID)
		})

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPVZRepository_GetPVZByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPVZRepository(db)
	pvzID := uuid.NewString()
	now := time.Now()

	// Тест успешного получения
	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, registration_date, city FROM pvz WHERE id =").
			WithArgs(pvzID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "registration_date", "city"}).
				AddRow(pvzID, now, "Москва"))

		pvz, err := repo.GetPVZByID(pvzID)

		assert.NoError(t, err)
		assert.Equal(t, pvzID, pvz.ID)
		assert.Equal(t, "Москва", pvz.City)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	// Тест не найден
	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, registration_date, city FROM pvz WHERE id =").
			WithArgs(pvzID).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.GetPVZByID(pvzID)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPVZRepository_ListPVZsWithRelations(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPVZRepository(db)
	now := time.Now()

	t.Run("success without date filter", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "registration_date", "city",
			"r.id", "r.created_at", "r.pvz_id", "r.status", "r.closed_at",
			"pr.id", "pr.created_at", "pr.type", "pr.reception_id",
		}).
			AddRow(
				"pvz1", now, "Москва",
				"rec1", now, "pvz1", "in_progress", nil,
				"prod1", now, "электроника", "rec1",
			).
			AddRow(
				"pvz1", now, "Москва",
				"rec1", now, "pvz1", "in_progress", nil,
				"prod2", now, "одежда", "rec1",
			).
			AddRow(
				"pvz2", now, "Санкт-Петербург",
				"rec2", now, "pvz2", "closed", now,
				nil, nil, nil, nil,
			)

		mock.ExpectQuery(`SELECT .* FROM pvz p`).
			WillReturnRows(rows)

		result, err := repo.ListPVZsWithRelations(time.Time{}, time.Time{}, 10, 0)

		assert.NoError(t, err)
		assert.Len(t, result, 2)

		assert.Equal(t, "Москва", result[0].PVZ.City)
		assert.Len(t, result[0].Receptions, 1)
		assert.Len(t, result[0].Receptions[0].Products, 2)

		assert.Equal(t, "Санкт-Петербург", result[1].PVZ.City)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
