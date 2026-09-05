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

	ID                          uuid.UUID  `bun:"id,pk"`
	CustomerID                  uuid.UUID  `bun:"customer_id"`
	ArtistID                    uuid.UUID  `bun:"artist_id"`
	ArtworkID                   *uuid.UUID `bun:"artwork_id"`
	ArtworkNameSnapshot         string     `bun:"artwork_name_snapshot"`
	ArtworkDescriptionSnapshot  string     `bun:"artwork_description_snapshot"`
	PriceSatangSnapshot         int64      `bun:"price_satang_snapshot"`
	MinimumDeadlineDaysSnapshot int        `bun:"minimum_deadline_days_snapshot"`
	PreviewImageURLSnapshot     string     `bun:"preview_image_url_snapshot"`
	CustomerDescription         string     `bun:"customer_description"`
	SelectedDeadlineDays        int        `bun:"selected_deadline_days"`
	DeadlineAt                  *time.Time `bun:"deadline_at"`
	Status                      string     `bun:"status"`
	CompletedAt                 *time.Time `bun:"completed_at"`
	CreatedAt                   time.Time  `bun:"created_at,nullzero"`
	UpdatedAt                   time.Time  `bun:"updated_at,nullzero"`
}

type orderDeliverableModel struct {
	bun.BaseModel `bun:"table:order_deliverables,alias:od"`

	ID               uuid.UUID `bun:"id,pk"`
	OrderID          uuid.UUID `bun:"order_id"`
	OriginalImageURL string    `bun:"original_image_url"`
	PreviewImageURL  string    `bun:"preview_image_url"`
	SortOrder        int       `bun:"sort_order"`
	CreatedAt        time.Time `bun:"created_at"`
}

func (m *orderModel) toDomain() order.Order {
	return order.Order{
		ID:                          m.ID,
		CustomerID:                  m.CustomerID,
		ArtistID:                    m.ArtistID,
		ArtworkID:                   m.ArtworkID,
		ArtworkNameSnapshot:         m.ArtworkNameSnapshot,
		ArtworkDescriptionSnapshot:  m.ArtworkDescriptionSnapshot,
		PriceSatangSnapshot:         m.PriceSatangSnapshot,
		MinimumDeadlineDaysSnapshot: m.MinimumDeadlineDaysSnapshot,
		PreviewImageURLSnapshot:     m.PreviewImageURLSnapshot,
		CustomerDescription:         m.CustomerDescription,
		SelectedDeadlineDays:        m.SelectedDeadlineDays,
		DeadlineAt:                  m.DeadlineAt,
		Status:                      order.Status(m.Status),
		CompletedAt:                 m.CompletedAt,
		CreatedAt:                   m.CreatedAt,
		UpdatedAt:                   m.UpdatedAt,
	}
}

func (m *orderDeliverableModel) toDomain() order.Deliverable {
	return order.Deliverable{
		ID:               m.ID,
		OriginalImageURL: m.OriginalImageURL,
		PreviewImageURL:  m.PreviewImageURL,
		SortOrder:        m.SortOrder,
		CreatedAt:        m.CreatedAt,
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
			ColumnExpr("o.artwork_id").
			ColumnExpr("o.artwork_name_snapshot").
			ColumnExpr("o.artwork_description_snapshot").
			ColumnExpr("o.price_satang_snapshot").
			ColumnExpr("o.minimum_deadline_days_snapshot").
			ColumnExpr("o.preview_image_url_snapshot").
			ColumnExpr("o.customer_description").
			ColumnExpr("o.selected_deadline_days").
			ColumnExpr("o.deadline_at").
			ColumnExpr("o.status").
			ColumnExpr("o.completed_at").
			ColumnExpr("o.created_at").
			ColumnExpr("o.updated_at").
			Where("o.customer_id = ?", customerID).
			OrderExpr("o.created_at DESC, o.id DESC").
			Scan(ctx)
	})
	if err != nil {
		return nil, apperror.Internal("failed to list orders for customer", err)
	}

	orders := make([]order.Order, len(models))
	for i, m := range models {
		orders[i] = m.toDomain()
	}

	if len(orders) > 0 {
		successfulOrderIDs := make([]uuid.UUID, 0, len(orders))
		byOrderID := make(map[uuid.UUID]*order.Order, len(orders))
		for i := range orders {
			byOrderID[orders[i].ID] = &orders[i]
			if orders[i].Status == order.StatusSuccess {
				successfulOrderIDs = append(successfulOrderIDs, orders[i].ID)
			}
		}
		if len(successfulOrderIDs) > 0 {
			err := r.exec.Run(ctx, func(idb bun.IDB) error {
				var deliverables []orderDeliverableModel
				if err := idb.NewSelect().
					Model(&deliverables).
					Where("od.order_id IN (?)", bun.In(successfulOrderIDs)).
					OrderExpr("od.order_id ASC, od.sort_order ASC").
					Scan(ctx); err != nil {
					return err
				}
				for i := range deliverables {
					parent := byOrderID[deliverables[i].OrderID]
					parent.Deliverables = append(parent.Deliverables, deliverables[i].toDomain())
				}
				return nil
			})
			if err != nil {
				return nil, apperror.Internal("failed to list order deliverables for customer", err)
			}
		}
	}
	return orders, nil
}
