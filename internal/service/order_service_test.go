package service

import (
	"context"
	"errors"
	"testing"

	"uuid"

	"github.com/Carlosam7/taller-backend-go/internal/models"
)

type MockOrderRepository struct {
	CreatedOrder *models.Order
	Err          error
}

func (m *MockOrderRepository) Create(
	ctx context.Context,
	order *models.Order,
) error {
	m.CreatedOrder = order

	return m.Err
}

func TestCreateOrder(t *testing.T) {
	repository := &MockOrderRepository{}
	service := NewOrderService(repository)

	request := models.CreateOrderRequest{
		CustomerId: "C123",
		Items: []models.CreateOrderItemRequest{
			{
				ProductId: "P1",
				Quantity:  2,
				Price:     10.50,
			},
			{
				ProductId: "P2",
				Quantity:  1,
				Price:     5.00,
			},
		},
	}

	order, err := service.CreateOrder(context.Background(), request)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if order.TotalAmount != 26.00 {
		t.Errorf("expected total 26.00, got %.2f", order.TotalAmount)
	}

	if order.ItemsCount != 3 {
		t.Errorf("expected 3 items, got %d", order.ItemsCount)
	}

	if len(order.Items) != 2 {
		t.Errorf("expected 2 order items, got %d", len(order.Items))
	}

	parsedUUID, err := uuid.Parse(order.OrderId)

	if err != nil || (parsedUUID[6]>>4) != 4 {
		t.Errorf("expected valid UUID v4, got %s", order.OrderId)
	}

	for _, item := range order.Items {
		if item.OrderId != order.OrderId {
			t.Errorf("item order ID does not match parent order")
		}
	}

	if order.CreatedAt.IsZero() {
		t.Error("expected processing timestamp")
	}
}

func TestCreateOrderValidation(t *testing.T) {
	repository := &MockOrderRepository{}
	service := NewOrderService(repository)

	tests := []struct {
		name    string
		request models.CreateOrderRequest
		wantErr error
	}{
		{
			name: "missing customer",
			request: models.CreateOrderRequest{
				Items: []models.CreateOrderItemRequest{
					{ProductId: "P1", Quantity: 1, Price: 10},
				},
			},
			wantErr: ErrCustomerIdRequired,
		},
		{
			name: "empty items",
			request: models.CreateOrderRequest{
				CustomerId: "C123",
			},
			wantErr: ErrItemsRequired,
		},
		{
			name: "missing product ID",
			request: models.CreateOrderRequest{
				CustomerId: "C123",
				Items: []models.CreateOrderItemRequest{
					{Quantity: 1, Price: 10},
				},
			},
			wantErr: ErrProductIdRequired,
		},
		{
			name: "invalid quantity",
			request: models.CreateOrderRequest{
				CustomerId: "C123",
				Items: []models.CreateOrderItemRequest{
					{ProductId: "P1", Quantity: 0, Price: 10},
				},
			},
			wantErr: ErrInvalidQuantity,
		},
		{
			name: "negative price",
			request: models.CreateOrderRequest{
				CustomerId: "C123",
				Items: []models.CreateOrderItemRequest{
					{ProductId: "P1", Quantity: 1, Price: -10},
				},
			},
			wantErr: ErrInvalidPrice,
		},
		{
			name: "more than two decimals",
			request: models.CreateOrderRequest{
				CustomerId: "C123",
				Items: []models.CreateOrderItemRequest{
					{ProductId: "P1", Quantity: 1, Price: 10.555},
				},
			},
			wantErr: ErrInvalidPriceFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.CreateOrder(context.Background(), tt.request)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}
