package handler

import (
	"io"
	"net/http"

	"github.com/go-chi/chi"
)

func NewHealthHandler() http.Handler {
	r := chi.NewRouter()

	// Public
	r.Group(func(r chi.Router) {
		r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			io.WriteString(w, "Ok.")
		})
	})

	return r
}
