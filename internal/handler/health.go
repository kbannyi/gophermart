package handler

import (
	"io"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/kbannyi/gophermart/internal/logger"
)

func NewHealthHandler() http.Handler {
	r := chi.NewRouter()

	// Public
	r.Group(func(r chi.Router) {
		r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			if _, err := io.WriteString(w, "Ok."); err != nil {
				logger.Log.Error("", "err", err)
			}
		})
	})

	return r
}
