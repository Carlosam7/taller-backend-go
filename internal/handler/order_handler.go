package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/Carlosam7/taller-backend-go/internal/models"
	"github.com/Carlosam7/taller-backend-go/internal/service"
)

type OrderCreator interface {
	CreateOrder(ctx context.Context, request models.CreateOrderRequest) (models.Order, error)
}

type OrderHandler struct {
	service OrderCreator
}

func NewOrderHandler(service OrderCreator) *OrderHandler {
	return &OrderHandler{
		service: service,
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *OrderHandler) CreateOrder(
	w http.ResponseWriter,
	r *http.Request,
) {
	// 1. Validate Content-Type.
	contentType, _, err := mime.ParseMediaType(
		r.Header.Get("Content-Type"),
	)

	if err != nil || contentType != "application/json" {
		writeJSON(w, http.StatusUnsupportedMediaType, ErrorResponse{
			Error: "Content-Type must be application/json",
		})
		return
	}

	// 2. Limit request body size to 1 MB.
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	// 3. Decode JSON request.
	var request models.CreateOrderRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "Invalid JSON request body",
		})
		return
	}

	// 4. Reject additional JSON values.
	var extra any

	if err := decoder.Decode(&extra); err != io.EOF {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "Request body must contain a single JSON object",
		})
		return
	}

	// 5. Delegate business logic to OrderService.
	order, err := h.service.CreateOrder(r.Context(), request)

	if err != nil {
		handleServiceError(w, err)
		return
	}

	// 6. Build response.
	response := models.CreateOrderResponse{
		OrderId:     order.OrderId,
		CustomerId:  order.CustomerId,
		TotalAmount: order.TotalAmount,
		ItemsCount:  order.ItemsCount,
		ProcessedAt: order.CreatedAt,
	}

	// 7. Return HTTP 201 Created.
	writeJSON(w, http.StatusCreated, response)
}

func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrCustomerIdRequired),
		errors.Is(err, service.ErrItemsRequired),
		errors.Is(err, service.ErrProductIdRequired),
		errors.Is(err, service.ErrInvalidQuantity),
		errors.Is(err, service.ErrInvalidPrice),
		errors.Is(err, service.ErrInvalidPriceFormat),
		errors.Is(err, service.ErrAmountExceeded):

		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		})

	default:
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error: "Internal server error",
		})
	}
}

func writeJSON(
	w http.ResponseWriter,
	status int,
	data any,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(data)
}
