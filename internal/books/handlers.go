package books

import (
	"context"
	"encoding/json"
	"log"
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

func SetBookDB(database db.TxDB) {
	dbi = database
}

func SetProducer(w Producer) {
	producer = w
}

type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

// @Summary Получить список книг
// @Tags books
// @Produce json
// @Success 200 {array} books.Book
// @Router /api/v1/books [get]
func ListBooks(w http.ResponseWriter, r *http.Request) {
	rows, err := dbi.Query(r.Context(), "SELECT id, title, author FROM books")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	var books []Book
	for rows.Next() {
		var b Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author); err != nil {
			continue
		}
		books = append(books, b)
	}
	if err := json.NewEncoder(w).Encode(books); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

// @Summary Создать книгу
// @Tags books
// @Accept json
// @Produce json
// @Param book body Book true "Книга"
// @Success 201 {object} Book
// @Router /api/v1/books [post]
func CreateBook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tx, err := dbi.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer func() {
		err := tx.Rollback(ctx)
		if err != nil && err != pgx.ErrTxClosed {
			log.Printf("ошибка Rollback: %v", err)
		}
	}()
	var b Book
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	row := tx.QueryRow(ctx, "INSERT INTO books (title, author) VALUES ($1, $2) RETURNING id", b.Title, b.Author)
	if err := row.Scan(&b.ID); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if producer != nil {
		if err := producer.WriteMessages(ctx, kafka.Message{Value: []byte("created book: " + b.Title)}); err != nil {
			log.Printf("ошибка отправки в Kafka: %v", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(b); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

// @Summary Получить книгу по id
// @Tags books
// @Produce json
// @Param id path int true "ID книги"
// @Success 200 {object} Book
// @Failure 404 {string} string "не найдено"
// @Router /api/v1/books/{id} [get]
func GetBook(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var b Book
	row := dbi.QueryRow(r.Context(), "SELECT id, title, author FROM books WHERE id=$1", id)
	if err := row.Scan(&b.ID, &b.Title, &b.Author); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err := json.NewEncoder(w).Encode(b); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

// @Summary Обновить книгу
// @Tags books
// @Accept json
// @Produce json
// @Param id path int true "ID книги"
// @Param book body Book true "Книга"
// @Success 200 {object} Book
// @Router /api/v1/books/{id} [put]
func UpdateBook(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var b Book
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	row := dbi.QueryRow(r.Context(), "UPDATE books SET title=$1, author=$2 WHERE id=$3 RETURNING id, title, author", b.Title, b.Author, id)
	if err := row.Scan(&b.ID, &b.Title, &b.Author); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if producer != nil {
		if err := producer.WriteMessages(r.Context(), kafka.Message{Value: []byte("updated book: " + id)}); err != nil {
			log.Printf("ошибка отправки в Kafka: %v", err)
		}
	}
	if err := json.NewEncoder(w).Encode(b); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

// @Summary Удалить книгу
// @Tags books
// @Param id path int true "ID книги"
// @Success 204 {string} string "Книга удалена"
// @Router /api/v1/books/{id} [delete]
func DeleteBook(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	row := dbi.QueryRow(r.Context(), "DELETE FROM books WHERE id=$1 RETURNING id", id)
	var deletedID int
	if err := row.Scan(&deletedID); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if producer != nil {
		if err := producer.WriteMessages(r.Context(), kafka.Message{Value: []byte("deleted book: " + id)}); err != nil {
			log.Printf("ошибка отправки в Kafka: %v", err)
		}
	}
	w.WriteHeader(http.StatusNoContent)
}
