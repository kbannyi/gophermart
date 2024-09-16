package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/kbannyi/gophermart/internal/auth"
	"github.com/kbannyi/gophermart/internal/domain"
	"github.com/kbannyi/gophermart/internal/logger"
	"github.com/kbannyi/gophermart/internal/repository"
)

type AuthService interface {
	Register(ctx context.Context, login string, pass string) (*domain.User, error)
	Login(ctx context.Context, login string, pass string) (*domain.User, error)
}

type AuthHandler struct{ service AuthService }

func NewAuthHandler(s AuthService) AuthHandler {
	return AuthHandler{s}
}

func (h AuthHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var reqmodel RegisterRequest
	if err := decoder.Decode(&reqmodel); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(reqmodel.Login) == 0 {
		http.Error(w, "login can't be empty", http.StatusBadRequest)
		return
	}
	if len(reqmodel.Password) == 0 {
		http.Error(w, "password can't be empty", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	user, err := h.service.Register(ctx, reqmodel.Login, reqmodel.Password)
	if errors.Is(err, repository.ErrAlreadyExists) {
		http.Error(w, "login already exists", http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	setAuthToken(w, auth.AuthUser{
		UserID: user.ID,
	})
	w.WriteHeader(http.StatusOK)
}

func (h AuthHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var reqmodel LoginRequest
	if err := decoder.Decode(&reqmodel); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(reqmodel.Login) == 0 {
		http.Error(w, "login can't be empty", http.StatusBadRequest)
		return
	}
	if len(reqmodel.Password) == 0 {
		http.Error(w, "password can't be empty", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	user, err := h.service.Login(ctx, reqmodel.Login, reqmodel.Password)
	if errors.Is(err, repository.ErrNotFound) {
		http.Error(w, "invalid login and/or password", http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	setAuthToken(w, auth.AuthUser{
		UserID: user.ID,
	})
	w.WriteHeader(http.StatusOK)
}

func setAuthToken(w http.ResponseWriter, u auth.AuthUser) {
	token, err := auth.BuildJWTString(u)
	if err != nil {
		logger.Log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set(auth.HeaderName, token)
	http.SetCookie(w, &http.Cookie{Name: auth.CookieName, Value: token})
}
