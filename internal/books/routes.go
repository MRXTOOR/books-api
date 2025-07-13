package books

import "github.com/go-chi/chi/v5"

// RegisterRoutes регистрирует роуты для книг
func RegisterRoutes(r chi.Router) {
	r.Route("/books", func(r chi.Router) {
		// @Summary Список книг
		// @Tags books
		// @Produce json
		// @Success 200 {array} Book
		// @Router /books [get]
		r.Get("/", ListBooks)

		// @Summary Создать книгу
		// @Tags books
		// @Accept json
		// @Produce json
		// @Param book body Book true "Книга"
		// @Success 201 {object} Book
		// @Router /books [post]
		r.Post("/", CreateBook)

		// @Summary Получить книгу
		// @Tags books
		// @Produce json
		// @Param id path int true "ID книги"
		// @Success 200 {object} Book
		// @Router /books/{id} [get]
		r.Get("/{id}", GetBook)

		// @Summary Обновить книгу
		// @Tags books
		// @Accept json
		// @Produce json
		// @Param id path int true "ID книги"
		// @Param book body Book true "Книга"
		// @Success 200 {object} Book
		// @Router /books/{id} [put]
		r.Put("/{id}", UpdateBook)

		// @Summary Удалить книгу
		// @Tags books
		// @Param id path int true "ID книги"
		// @Success 204
		// @Router /books/{id} [delete]
		r.Delete("/{id}", DeleteBook)
	})
}
