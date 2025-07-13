package collections

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/segmentio/kafka-go"

	"books-api/internal/db"
)

type Producer interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

var dbi db.TxDB
var producer Producer

func SetCollectionDB(database db.TxDB) {
	dbi = database
}

func SetProducer(w Producer) {
	producer = w
}

type Collection struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Books       []int  `json:"books,omitempty"`
}

// @Summary Создать подборку
// @Tags collections
// @Accept json
// @Produce json
// @Param collection body Collection true "Подборка"
// @Success 201 {object} Collection
// @Router /api/v1/collections [post]
func CreateCollection(w http.ResponseWriter, r *http.Request) {
	var c Collection
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	row := dbi.QueryRow(r.Context(), "INSERT INTO collections (name, description) VALUES ($1, $2) RETURNING id, name, description", c.Name, c.Description)
	if err := row.Scan(&c.ID, &c.Name, &c.Description); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if producer != nil {
		if err := producer.WriteMessages(r.Context(), kafka.Message{Value: []byte("created collection: " + c.Name)}); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}
	if err := json.NewEncoder(w).Encode(c); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(201)
}

// @Summary Получить список подборок
// @Tags collections
// @Produce json
// @Success 200 {array} Collection
// @Router /api/v1/collections [get]
func ListCollections(w http.ResponseWriter, r *http.Request) {
	rows, err := dbi.Query(r.Context(), "SELECT id, name, description FROM collections")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	var collections []Collection
	for rows.Next() {
		var c Collection
		if err := rows.Scan(&c.ID, &c.Name, &c.Description); err != nil {
			continue
		}
		collections = append(collections, c)
	}
	if err := json.NewEncoder(w).Encode(collections); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

// @Summary Получить подборку по id
// @Tags collections
// @Produce json
// @Param id path int true "ID подборки"
// @Success 200 {object} Collection
// @Failure 404 {string} string "не найдено"
// @Router /api/v1/collections/{id} [get]
func GetCollection(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "не указан id коллекции", 400)
		return
	}
	var c Collection
	row := dbi.QueryRow(r.Context(), "SELECT id, name, description FROM collections WHERE id=$1", id)
	if err := row.Scan(&c.ID, &c.Name, &c.Description); err != nil {
		http.Error(w, "не найдено", 404)
		return
	}
	// Получаем книги в подборке
	booksRows, err := dbi.Query(r.Context(), "SELECT book_id FROM collection_books WHERE collection_id=$1", id)
	if err == nil {
		defer booksRows.Close()
		for booksRows.Next() {
			var bookID int
			if err := booksRows.Scan(&bookID); err == nil {
				c.Books = append(c.Books, bookID)
			}
		}
	}
	if err := json.NewEncoder(w).Encode(c); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

// @Summary Добавить книгу в подборку
// @Tags collections
// @Accept json
// @Param id path int true "ID подборки"
// @Param book_id body object true "ID книги"
// @Success 204 {string} string "Книга добавлена"
// @Router /api/v1/collections/{id}/books [post]
func AddBookToCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tx, err := dbi.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			// Можно логировать ошибку
		}
	}()
	id := chi.URLParam(r, "id")
	var req struct {
		BookID int `json:"book_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	_, err = tx.Exec(ctx, "INSERT INTO collection_books (collection_id, book_id) VALUES ($1, $2)", id, req.BookID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if producer != nil {
		if err := producer.WriteMessages(ctx, kafka.Message{Value: []byte("added book to collection: " + id)}); err != nil {
			// Можно логировать ошибку
		}
	}
	if err := tx.Commit(ctx); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(204)
}

// @Summary Удалить книгу из подборки
// @Tags collections
// @Param id path int true "ID подборки"
// @Param book_id path int true "ID книги"
// @Success 204 {string} string "Книга удалена"
// @Router /api/v1/collections/{id}/books/{book_id} [delete]
func RemoveBookFromCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tx, err := dbi.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			// Можно логировать ошибку
		}
	}()
	id := chi.URLParam(r, "id")
	bookID := chi.URLParam(r, "book_id")
	row := tx.QueryRow(ctx, "DELETE FROM collection_books WHERE collection_id=$1 AND book_id=$2 RETURNING book_id", id, bookID)
	var deletedID int
	if err := row.Scan(&deletedID); err != nil {
		http.Error(w, "not found", 404)
		return
	}
	if producer != nil {
		if err := producer.WriteMessages(ctx, kafka.Message{Value: []byte("removed book from collection: " + id)}); err != nil {
			// Можно логировать ошибку
		}
	}
	if err := tx.Commit(ctx); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(204)
}
