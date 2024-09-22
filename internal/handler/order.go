package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/kbannyi/gophermart/internal/domain"
	"github.com/kbannyi/gophermart/internal/dto"
	"github.com/kbannyi/gophermart/internal/logger"
	"github.com/kbannyi/gophermart/internal/repository"
	"github.com/kbannyi/gophermart/internal/service"
	"github.com/kbannyi/gophermart/internal/service/luhn"
)

type OrderService interface {
	SaveNewOrder(ctx context.Context, id string) error
	GetOrders(ctx context.Context) ([]domain.Order, error)
}

type OrderHandler struct {
	service OrderService
}

func NewOrderHandler(s OrderService) OrderHandler {
	return OrderHandler{service: s}
}

func (h OrderHandler) SaveOrder(w http.ResponseWriter, r *http.Request) {
	linkBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	orderid := string(linkBytes)
	if !luhn.Valid(orderid) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	ctx := r.Context()
	err = h.service.SaveNewOrder(ctx, orderid)
	if errors.Is(err, service.ErrBelongToAnother) {
		w.WriteHeader(http.StatusConflict)
		return
	}
	if errors.Is(err, repository.ErrAlreadyExists) {
		w.WriteHeader(http.StatusOK)
		return
	}
	if err != nil {
		logger.Log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orders, err := h.service.GetOrders(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	response := make([]dto.OrderResponse, 0, len(orders))
	for _, o := range orders {
		response = append(response, dto.OrderResponse{
			Number:     o.ID,
			Status:     o.Status.String(),
			Accrual:    o.Accrual,
			UploadedAt: o.CreatedUTC.Format(time.RFC3339),
		})
	}

	encoder := json.NewEncoder(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := encoder.Encode(&response); err != nil {
		logger.Log.Error(err.Error())
		return
	}
}
