package collections

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/segmentio/kafka-go"

	"books-api/internal/db"
)

type mockRows struct{ idx int }

func (r *mockRows) Next() bool { r.idx++; return r.idx == 1 }
func (r *mockRows) Scan(dest ...any) error {
	*dest[0].(*int) = 1
	return nil
}
func (r *mockRows) Close() {}

type mockRow struct{}

func (r *mockRow) Scan(dest ...any) error {
	*dest[0].(*int) = 1
	return nil
}

type mockDB struct{}

func (m *mockDB) Query(ctx context.Context, sql string, args ...any) (db.Rows, error) {
	return &mockRows{}, nil
}
func (m *mockDB) QueryRow(ctx context.Context, sql string, args ...any) db.Row {
	return &mockRow{}
}
func (m *mockDB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return pgconn.NewCommandTag("MOCK"), nil
}
func (m *mockDB) BeginTx(ctx context.Context, opts pgx.TxOptions) (db.TxDB, error) {
	return m, nil
}
func (m *mockDB) Rollback(ctx context.Context) error { return nil }
func (m *mockDB) Commit(ctx context.Context) error   { return nil }

type mockProducer struct{}

func (m *mockProducer) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	return nil
}
func (m *mockProducer) Close() error { return nil }

func TestCreateCollection(t *testing.T) {
	SetCollectionDB(&mockDB{})
	SetProducer(&mockProducer{})
	c := Collection{Name: "Test", Description: "Desc"}
	body, _ := json.Marshal(c)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/collections", bytes.NewReader(body))
	w := httptest.NewRecorder()
	CreateCollection(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestListCollections(t *testing.T) {
	SetCollectionDB(&mockDB{})
	SetProducer(&mockProducer{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/collections", nil)
	w := httptest.NewRecorder()
	ListCollections(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetCollection(t *testing.T) {
	SetCollectionDB(&mockDB{})
	SetProducer(&mockProducer{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/collections/1", nil)
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
	w := httptest.NewRecorder()
	GetCollection(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAddBookToCollection(t *testing.T) {
	SetCollectionDB(&mockDB{})
	SetProducer(&mockProducer{})
	body := []byte(`{"book_id":1}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/collections/1/books", bytes.NewReader(body))
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
	w := httptest.NewRecorder()
	AddBookToCollection(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestRemoveBookFromCollection(t *testing.T) {
	SetCollectionDB(&mockDB{})
	SetProducer(&mockProducer{})
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/collections/1/books/1", nil)
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", "1")
	ctx.URLParams.Add("book_id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
	w := httptest.NewRecorder()
	RemoveBookFromCollection(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}
