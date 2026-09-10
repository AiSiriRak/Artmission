package postgres

import (
	"context"
	"errors"
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
	Name                        string     `bun:"name"`
	ArtworkID                   *uuid.UUID `bun:"artwork_id"`
	ArtworkNameSnapshot         string     `bun:"artwork_name_snapshot"`
	ArtworkDescriptionSnapshot  string     `bun:"artwork_description_snapshot"`
	PriceSatangSnapshot         int64      `bun:"price_satang_snapshot"`
	MinimumDeadlineDaysSnapshot int        `bun:"minimum_deadline_days_snapshot"`
	CustomerDescription         string     `bun:"customer_description"`
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
	Version          int       `bun:"version"`
	Decision         *string   `bun:"decision"`
	OriginalImageKey string    `bun:"original_image_key"`
	PreviewImageKey  string    `bun:"preview_image_key"`
	CreatedAt        time.Time `bun:"created_at"`
}

func (m *orderModel) toDomain() order.Order {
	return order.Order{
		ID:                          m.ID,
		CustomerID:                  m.CustomerID,
		ArtistID:                    m.ArtistID,
		Name:                        m.Name,
		ArtworkID:                   m.ArtworkID,
		ArtworkNameSnapshot:         m.ArtworkNameSnapshot,
		ArtworkDescriptionSnapshot:  m.ArtworkDescriptionSnapshot,
		PriceSatangSnapshot:         m.PriceSatangSnapshot,
		MinimumDeadlineDaysSnapshot: m.MinimumDeadlineDaysSnapshot,
		CustomerDescription:         m.CustomerDescription,
		DeadlineAt:                  m.DeadlineAt,
		Status:                      order.Status(m.Status),
		CompletedAt:                 m.CompletedAt,
		CreatedAt:                   m.CreatedAt,
		UpdatedAt:                   m.UpdatedAt,
	}
}

type orderRepository struct {
	exec baserepo.Executor
}

var _ order.OrderRepository = (*orderRepository)(nil)

func NewOrderRepository(db *bun.DB) order.OrderRepository {
	return &orderRepository{exec: baserepo.NewExecutor(db)}
}

// orderSortColumns maps each ViewOrders sort field to the one trusted,
// allowlisted SQL column used for ORDER BY.
var orderSortColumns = map[order.SortField]string{
	order.SortFieldUpdatedAt: "o.updated_at",
	order.SortFieldPrice:     "o.price_satang_snapshot",
	order.SortFieldDeadline:  "o.deadline_at",
}

func applyStatusFilter(q *bun.SelectQuery, statuses []order.Status) *bun.SelectQuery {
	if len(statuses) == 0 {
		return q
	}
	values := make([]string, len(statuses))
	for i, s := range statuses {
		values[i] = string(s)
	}
	return q.Where("o.status IN (?)", bun.List(values))
}

// ListOrders scopes strictly to query.Participant/ParticipantID, applies
// query.Statuses, and paginates by query.Sort/Order/Limit/Offset via the
// shared baserepo.Paginate helper. query is assumed already validated and
// defaulted by orderUsecase.ViewOrders. Every order is enriched with its
// latest submitted deliverable's preview key, regardless of status.
func (r *orderRepository) ListOrders(ctx context.Context, query order.ListQuery) (order.Page, error) {
	column, ok := orderSortColumns[query.Sort]
	if !ok {
		return order.Page{}, apperror.Internal("unsupported sort field", nil)
	}
	if query.Participant != order.ParticipantArtist && query.Participant != order.ParticipantCustomer {
		// orderUsecase.ViewOrders already rejects any other participant;
		// this is a defensive guard against ever running a query with no
		// participant scope predicate at all.
		return order.Page{}, apperror.Internal("unsupported participant", nil)
	}

	dir := "ASC"
	if query.Order == order.SortOrderDesc {
		dir = "DESC"
	}

	page, err := baserepo.Paginate[orderModel](ctx, r.exec, func(q *bun.SelectQuery) *bun.SelectQuery {
		if query.Participant == order.ParticipantArtist {
			q = q.Where("o.artist_id = ?", query.ParticipantID)
		} else {
			q = q.Where("o.customer_id = ?", query.ParticipantID)
		}

		q = applyStatusFilter(q, query.Statuses)

		// o.id is a deterministic tie-breaker: without it, Postgres does
		// not guarantee a stable order across requests when column has
		// duplicate values, which could duplicate or skip rows across
		// pages.
		return q.Order(column+" "+dir, "o.id "+dir)
	}, baserepo.PaginationInput{Limit: query.Limit, Offset: query.Offset})
	if err != nil {
		if _, ok := errors.AsType[*apperror.Error](err); ok {
			return order.Page{}, err
		}
		return order.Page{}, apperror.Internal("failed to list orders", err)
	}

	orders := make([]order.Order, len(page.Items))
	for i, m := range page.Items {
		orders[i] = m.toDomain()
	}
	if err := r.attachLatestDeliverablePreviewKeys(ctx, orders); err != nil {
		return order.Page{}, apperror.Internal("failed to attach deliverable preview keys", err)
	}

	return order.Page{Orders: orders, Total: page.Total}, nil
}

// attachLatestDeliverablePreviewKeys sets DeliverablePreviewKey on every
// order in orders to its most recently submitted deliverable version's
// preview_image_key, regardless of order status — nil if none has been
// submitted yet.
func (r *orderRepository) attachLatestDeliverablePreviewKeys(ctx context.Context, orders []order.Order) error {
	if len(orders) == 0 {
		return nil
	}

	orderIDs := make([]uuid.UUID, len(orders))
	byOrderID := make(map[uuid.UUID]*order.Order, len(orders))
	for i := range orders {
		orderIDs[i] = orders[i].ID
		byOrderID[orders[i].ID] = &orders[i]
	}

	return r.exec.Run(ctx, func(idb bun.IDB) error {
		var deliverables []orderDeliverableModel
		if err := idb.NewSelect().
			Model(&deliverables).
			Column("order_id", "preview_image_key").
			DistinctOn("od.order_id").
			Where("od.order_id IN (?)", bun.List(orderIDs)).
			OrderExpr("od.order_id ASC, od.version DESC").
			Scan(ctx); err != nil {
			return err
		}
		for i := range deliverables {
			orderID := deliverables[i].OrderID
			key := deliverables[i].PreviewImageKey
			byOrderID[orderID].DeliverablePreviewKey = &key
		}
		return nil
	})
}
