package models

import (
	"time"
	"uuid"
)

type UUID uuid.UUID

type Order struct {
	Id          UUID        `json:"id"`
	OrderId     string      `json:"order_id"`
	CustomerId  string      `json:"customer_id"`
	TotalAmount float64     `json:"total_amount"`
	ItemsCount  int         `json:"items_count"`
	CreatedAt   time.Time   `json:"created_at"`
	Items       []OrderItem `json:"items"`
}

type OrderItem struct {
	Id        UUID    `json:"id"`
	OrderId   string  `json:"order_id"`
	ProductId string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type CreateOrderRequest struct {
	CustomerId string                   `json:"customer_id"`
	Items      []CreateOrderItemRequest `json:"items"`
}

type CreateOrderItemRequest struct {
	ProductId string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type CreateOrderResponse struct {
	OrderId     string    `json:"order_id"`
	CustomerId  string    `json:"customer_id"`
	TotalAmount float64   `json:"total_amount"`
	ItemsCount  int       `json:"items_count"`
	ProcessedAt time.Time `json:"processed_at"`
}
