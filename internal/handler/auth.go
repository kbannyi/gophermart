package handler

import (
	"io"
	"net/http"

	"github.com/kbannyi/gophermart/internal/logger"
)

type AuthHandler struct{}

func NewAuthHandler() AuthHandler {
	return AuthHandler{}
}

func (h AuthHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	if _, err := io.WriteString(w, "Register"); err != nil {
		logger.Log.Error(err.Error())
	}

}

func (h AuthHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	if _, err := io.WriteString(w, "Login"); err != nil {
		logger.Log.Error(err.Error())
	}
}
