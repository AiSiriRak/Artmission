package postgres

import (
	"context"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/baserepo"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type orderModel struct {
	bun.BaseModel `bun:"table:orders,alias:o"`

	ID          uuid.UUID  `bun:"id,pk"`
	CustomerID  uuid.UUID  `bun:"customer_id"`
	ArtistID    uuid.UUID  `bun:"artist_id"`
	Description string     `bun:"description"`
	Category    string     `bun:"category,nullzero"`
	Style       string     `bun:"style,nullzero"`
	Price       *float64   `bun:"price"`
	Status      string     `bun:"status"`
	Deadline    *time.Time `bun:"deadline"`
	CompletedAt *time.Time `bun:"completed_at"`
	CreatedAt   time.Time  `bun:"created_at,nullzero"`
	UpdatedAt   time.Time  `bun:"updated_at,nullzero"`
}

func (m *orderModel) toDomain() order.Order {
	return order.Order{
		ID:          m.ID,
		CustomerID:  m.CustomerID,
		ArtistID:    m.ArtistID,
		Description: m.Description,
		Category:    m.Category,
		Style:       m.Style,
		Price:       m.Price,
		Status:      order.Status(m.Status),
		Deadline:    m.Deadline,
		CompletedAt: m.CompletedAt,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

type orderRepository struct {
	exec baserepo.Executor
}

var _ order.OrderRepository = (*orderRepository)(nil)

func NewOrderRepository(db *bun.DB) order.OrderRepository {
	return &orderRepository{exec: baserepo.NewExecutor(db)}
}

func (r *orderRepository) ListByCustomerID(ctx context.Context, customerID uuid.UUID) ([]order.Order, error) {
	var models []orderModel
	err := r.exec.Run(ctx, func(idb bun.IDB) error {
		return idb.NewSelect().
			Model(&models).
			Where("customer_id = ?", customerID).
			OrderExpr("created_at DESC").
			Scan(ctx)
	})
	if err != nil {
		return nil, apperror.Internal("failed to list orders for customer", err)
	}

	orders := make([]order.Order, len(models))
	for i, m := range models {
		orders[i] = m.toDomain()
	}
	return orders, nil
}
