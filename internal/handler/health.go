package handler

import (
	"io"
	"net/http"

	"github.com/kbannyi/gophermart/internal/logger"
)

type HealthHandler struct{}

func (h HealthHandler) Ping(w http.ResponseWriter, r *http.Request) {
	if _, err := io.WriteString(w, "Ok."); err != nil {
		logger.Log.Error("", "err", err)
	}
}

func NewHealthHandler() HealthHandler {
	return HealthHandler{}
}
