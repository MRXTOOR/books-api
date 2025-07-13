package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"books-api/internal/books"
	"books-api/internal/collections"
	"books-api/internal/db"
	"books-api/internal/kafka"
	custommw "books-api/internal/middleware"
)

func main() {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		log.Fatal("DATABASE_DSN is not set")
	}
	pool, err := db.NewDB(dsn)
	if err != nil {
		log.Fatalf("DB error: %v", err)
	}
	defer pool.Close()

	brokers := []string{"localhost:9092"}
	writer := kafka.NewProducer(brokers, "books-events")
	defer writer.Close()

	dbAdapter := &db.PgxPoolTxDB{Pool: pool}
	books.SetBookDB(dbAdapter)
	books.SetProducer(writer)
	collections.SetCollectionDB(dbAdapter)
	collections.SetProducer(writer)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(custommw.Logger)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	r.Route("/api/v1", func(r chi.Router) {
		books.RegisterRoutes(r)
		collections.RegisterRoutes(r)
	})

	log.Println("Server started on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
