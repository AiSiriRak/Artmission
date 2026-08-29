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
	CategoryID  *uuid.UUID `bun:"category_id"`
	Category    string     `bun:"category_label,nullzero"`
	StyleID     *uuid.UUID `bun:"style_id"`
	Style       string     `bun:"style_label,nullzero"`
	Price       *float64   `bun:"price"`
	Status      string     `bun:"status"`
	Deadline    *time.Time `bun:"deadline"`
	CompletedAt *time.Time `bun:"completed_at"`
	CreatedAt   time.Time  `bun:"created_at,nullzero"`
	UpdatedAt   time.Time  `bun:"updated_at,nullzero"`
}

func (m *orderModel) toDomain() order.Order {
	var category *order.Category
	if m.CategoryID != nil {
		category = &order.Category{ID: *m.CategoryID, Label: m.Category}
	}

	var style *order.Style
	if m.StyleID != nil {
		style = &order.Style{ID: *m.StyleID, Label: m.Style}
	}

	return order.Order{
		ID:          m.ID,
		CustomerID:  m.CustomerID,
		ArtistID:    m.ArtistID,
		Description: m.Description,
		Category:    category,
		Style:       style,
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
			ColumnExpr("o.id").
			ColumnExpr("o.customer_id").
			ColumnExpr("o.artist_id").
			ColumnExpr("o.description").
			ColumnExpr("o.category_id").
			ColumnExpr("c.label AS category_label").
			ColumnExpr("o.style_id").
			ColumnExpr("s.label AS style_label").
			ColumnExpr("o.price").
			ColumnExpr("o.status").
			ColumnExpr("o.deadline").
			ColumnExpr("o.completed_at").
			ColumnExpr("o.created_at").
			ColumnExpr("o.updated_at").
			Join("LEFT JOIN categories AS c ON c.id = o.category_id").
			Join("LEFT JOIN styles AS s ON s.id = o.style_id").
			Where("o.customer_id = ?", customerID).
			OrderExpr("o.created_at DESC").
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
