package collections

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router) {
	r.Route("/collections", func(r chi.Router) {
		r.Post("/", CreateCollection)
		r.Get("/", ListCollections)
		r.Get("/{id}", GetCollection)
		r.Post("/{id}/books", AddBookToCollection)
		r.Delete("/{id}/books/{book_id}", RemoveBookFromCollection)
	})
}
