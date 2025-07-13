package books

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
	*dest[1].(*string) = "Test Book"
	*dest[2].(*string) = "Author"
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

func TestListBooks(t *testing.T) {
	SetBookDB(&mockDB{})
	SetProducer(&mockProducer{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/books", nil)
	w := httptest.NewRecorder()
	ListBooks(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var books []Book
	if err := json.NewDecoder(w.Body).Decode(&books); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(books) != 1 || books[0].Title != "Test Book" {
		t.Fatalf("unexpected books: %+v", books)
	}
}

func TestCreateBook(t *testing.T) {
	SetBookDB(&mockDB{})
	SetProducer(&mockProducer{})
	b := Book{Title: "Test", Author: "A"}
	body, _ := json.Marshal(b)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/books", bytes.NewReader(body))
	w := httptest.NewRecorder()
	CreateBook(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	var resp Book
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.ID != 1 {
		t.Fatalf("unexpected id: %d", resp.ID)
	}
}

func TestGetBook(t *testing.T) {
	SetBookDB(&mockDB{})
	SetProducer(&mockProducer{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/books/1", nil)
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
	w := httptest.NewRecorder()
	GetBook(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var b Book
	if err := json.NewDecoder(w.Body).Decode(&b); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if b.ID != 1 {
		t.Fatalf("unexpected id: %d", b.ID)
	}
}

func TestUpdateBook(t *testing.T) {
	SetBookDB(&mockDB{})
	SetProducer(&mockProducer{})
	b := Book{Title: "Updated", Author: "B"}
	body, _ := json.Marshal(b)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/books/1", bytes.NewReader(body))
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
	w := httptest.NewRecorder()
	UpdateBook(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp Book
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.ID != 1 {
		t.Fatalf("unexpected id: %d", resp.ID)
	}
}

func TestDeleteBook(t *testing.T) {
	SetBookDB(&mockDB{})
	SetProducer(&mockProducer{})
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/books/1", nil)
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
	w := httptest.NewRecorder()
	DeleteBook(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}
