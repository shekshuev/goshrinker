package store

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shekshuev/shortener/internal/app/config"
	"github.com/shekshuev/shortener/internal/app/models"
)

func BenchmarkPostgresURLStore_SetURL(b *testing.B) {
	cfg := config.GetConfig()
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("Error creating db mock: %v", err)
	}
	defer db.Close()
	store := &PostgresURLStore{cfg: &cfg, db: db}

	mock.ExpectQuery(`INSERT INTO urls`).WillReturnRows(sqlmock.NewRows([]string{"is_new", "shorted_url"}).AddRow(true, "short_test"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = store.SetURL("short_test", "https://example.com", "1")
	}
}

func BenchmarkPostgresURLStore_SetBatchURL(b *testing.B) {
	cfg := config.GetConfig()
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("Error creating db mock: %v", err)
	}
	defer db.Close()
	store := &PostgresURLStore{cfg: &cfg, db: db}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO urls`).WillReturnRows(sqlmock.NewRows([]string{"is_new", "shorted_url"}).AddRow(true, "short_test"))
	mock.ExpectCommit()

	createDTO := []models.BatchShortURLCreateDTO{
		{CorrelationID: "test1", OriginalURL: "https://example1.com", ShortURL: "short1"},
		{CorrelationID: "test2", OriginalURL: "https://example2.com", ShortURL: "short2"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = store.SetBatchURL(createDTO, "1")
	}
}

func BenchmarkPostgresURLStore_GetURL(b *testing.B) {
	cfg := config.GetConfig()
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("Error creating db mock: %v", err)
	}
	defer db.Close()
	store := &PostgresURLStore{cfg: &cfg, db: db}

	mock.ExpectQuery(`SELECT original_url, deleted_at`).WillReturnRows(sqlmock.NewRows([]string{"original_url", "is_deleted"}).AddRow("https://example.com", false))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = store.GetURL("short_test")
	}
}

func BenchmarkPostgresURLStore_GetUserURLs(b *testing.B) {
	cfg := config.GetConfig()
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("Error creating db mock: %v", err)
	}
	defer db.Close()
	store := &PostgresURLStore{cfg: &cfg, db: db}

	mock.ExpectQuery(`SELECT original_url, shorted_url`).WillReturnRows(
		sqlmock.NewRows([]string{"original_url", "shorted_url"}).
			AddRow("https://example1.com", "short1").
			AddRow("https://example2.com", "short2"),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = store.GetUserURLs("1")
	}
}

func BenchmarkPostgresURLStore_DeleteURLs(b *testing.B) {
	cfg := config.GetConfig()
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("Error creating db mock: %v", err)
	}
	defer db.Close()
	store := &PostgresURLStore{cfg: &cfg, db: db}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE urls SET deleted_at`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	urls := []string{"short1", "short2", "short3"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = store.DeleteURLs("1", urls)
	}
}
