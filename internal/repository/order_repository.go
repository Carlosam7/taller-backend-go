package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Carlosam7/taller-backend-go/internal/models"
)

type PostgresOrderRepository struct {
	db *pgxpool.Pool
}

func NewPostgresOrderRepository(db *pgxpool.Pool) *PostgresOrderRepository {
	return &PostgresOrderRepository{
		db: db,
	}
}

func (r *PostgresOrderRepository) Create(
	ctx context.Context,
	order *models.Order,
) error {

	// 1. Begin transaction.
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	// Ensure rollback if the transaction has not been committed.
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// 2. Insert order.
	err = tx.QueryRow(ctx, `
		INSERT INTO orders (
			order_id,
			customer_id,
			total_amount,
			items_count
		)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at
	`,
		order.OrderId,
		order.CustomerId,
		order.TotalAmount,
		order.ItemsCount,
	).Scan(&order.CreatedAt)

	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}

	// 3. Insert order items.
	for _, item := range order.Items {
		_, err := tx.Exec(ctx, `
			INSERT INTO order_items (
				order_id,
				product_id,
				quantity,
				price
			)
			VALUES ($1, $2, $3, $4)
		`,
			order.OrderId,
			item.ProductId,
			item.Quantity,
			item.Price,
		)

		if err != nil {
			return fmt.Errorf("insert order item: %w", err)
		}
	}

	// 4. Commit transaction.
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
