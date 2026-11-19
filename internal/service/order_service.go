package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"uuid"

	"github.com/Carlosam7/taller-backend-go/internal/models"
)

const maxAmountCents int64 = 99_999_999_99

var (
	ErrCustomerIdRequired = errors.New("Customer id is required")
	ErrItemsRequired      = errors.New("Items are required")
	ErrProductIdRequired  = errors.New("Product id is required")
	ErrInvalidQuantity    = errors.New("Quantity must be greater than 0")
	ErrInvalidPrice       = errors.New("Price cannot be negative")
	ErrInvalidPriceFormat = errors.New("Invalid price format")
	ErrAmountExceeded     = errors.New("Order total exceeds the database limit")
)

type OrderRepository interface {
	Create(ctx context.Context, order *models.Order) error
}

type OrderService struct {
	repository OrderRepository
}

func NewOrderService(repository OrderRepository) *OrderService {
	return &OrderService{
		repository: repository,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, request models.CreateOrderRequest) (models.Order, error) {
	if strings.TrimSpace(request.CustomerId) == "" {
		return models.Order{}, ErrCustomerIdRequired
	}

	if len(request.Items) == 0 {
		return models.Order{}, ErrItemsRequired
	}

	orderId := uuid.New().String()

	var totalCents int64
	var itemsCount int

	orderItems := make([]models.OrderItem, 0, len(request.Items))

	for i, item := range request.Items {
		if strings.TrimSpace(item.ProductId) == "" {
			return models.Order{}, fmt.Errorf("items[%d]: %w", i, ErrProductIdRequired)
		}

		if item.Quantity <= 0 {
			return models.Order{}, fmt.Errorf("items[%d]: %w", i, ErrInvalidQuantity)
		}

		if math.IsNaN(item.Price) || math.IsInf(item.Price, 0) || item.Price < 0 {
			return models.Order{}, fmt.Errorf("items[%d]: %w", i, ErrInvalidPrice)
		}

		scaledPrice := item.Price * 100
		roundedPrice := math.Round(scaledPrice)

		if math.Abs(scaledPrice-roundedPrice) > 0.000001 {
			return models.Order{}, fmt.Errorf(
				"items[%d]: %w", i, ErrInvalidPriceFormat,
			)
		}

		if roundedPrice > float64(maxAmountCents) {
			return models.Order{}, fmt.Errorf(
				"items[%d]: %w", i, ErrAmountExceeded,
			)
		}

		priceCents := int64(roundedPrice)
		quantity := int64(item.Quantity)

		// Check multiplication and total limits before calculating.
		if priceCents > 0 &&
			quantity > maxAmountCents/priceCents {
			return models.Order{}, ErrAmountExceeded
		}

		lineTotal := priceCents * quantity

		if totalCents > maxAmountCents-lineTotal {
			return models.Order{}, ErrAmountExceeded
		}

		totalCents += lineTotal
		itemsCount += item.Quantity

		orderItems = append(orderItems, models.OrderItem{
			OrderId:   orderId,
			ProductId: item.ProductId,
			Quantity:  item.Quantity,
			Price:     float64(priceCents) / 100,
		})
	}
	// 6. Calculate final total.
	totalAmount := float64(totalCents) / 100

	// 7. Set processing timestamp.
	now := time.Now().UTC()

	// 8. Build order entity.
	order := models.Order{
		OrderId:     orderId,
		CustomerId:  strings.TrimSpace(request.CustomerId),
		TotalAmount: totalAmount,
		ItemsCount:  itemsCount,
		CreatedAt:   now,
		Items:       orderItems,
	}

	if err := s.repository.Create(ctx, &order); err != nil {
		return models.Order{}, fmt.Errorf("create order: %w", err)
	}

	return order, nil

}
