package db

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
)

func execSQLFile(t *testing.T, conn *pgx.Conn, path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read migration file %s: %v", path, err)
	}
	if _, err := conn.Exec(context.Background(), string(data)); err != nil {
		t.Fatalf("failed to execute migration %s: %v", path, err)
	}
}

func TestMigrationsApplied(t *testing.T) {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "postgres://books:books@localhost:5432/books?sslmode=disable"
	}
	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		t.Fatalf("cannot connect to db: %v", err)
	}
	defer conn.Close(context.Background())

	// Clean up tables before applying migrations
	if _, err := conn.Exec(context.Background(), "DROP TABLE IF EXISTS collection_books"); err != nil {
		t.Fatalf("failed to drop collection_books: %v", err)
	}
	if _, err := conn.Exec(context.Background(), "DROP TABLE IF EXISTS collections"); err != nil {
		t.Fatalf("failed to drop collections: %v", err)
	}
	if _, err := conn.Exec(context.Background(), "DROP TABLE IF EXISTS books"); err != nil {
		t.Fatalf("failed to drop books: %v", err)
	}

	execSQLFile(t, conn, "../../migrations/001_create_books.sql")
	execSQLFile(t, conn, "../../migrations/002_create_collections.sql")

	tables := []string{"books", "collections", "collection_books"}
	for _, table := range tables {
		var exists bool
		err := conn.QueryRow(context.Background(),
			`SELECT EXISTS (
				SELECT 1 FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = $1
			)`, table).Scan(&exists)
		if err != nil {
			t.Errorf("error checking table %s: %v", table, err)
		}
		if !exists {
			t.Errorf("table %s does not exist", table)
		}
	}
}
