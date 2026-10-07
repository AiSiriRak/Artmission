package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	pgmodel "github.com/AiSiriRak/Artmission/backend/internal/adapters/postgres/model"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/baserepo"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// orderModelToDomain converts a PostgreSQL order model into its domain value.
func orderModelToDomain(m *pgmodel.Order) order.Order {
	return order.Order{
		ID:                  m.ID,
		CustomerID:          m.CustomerID,
		ArtistID:            m.ArtistID,
		Name:                m.Name,
		ArtworkID:           m.ArtworkID,
		PriceSatangOrder:    m.PriceSatangOrder,
		CustomerDescription: m.CustomerDescription,
		DeadlineAt:          m.DeadlineAt,
		Status:              order.Status(m.Status),
		CompletedAt:         m.CompletedAt,
		CreatedAt:           m.CreatedAt,
		UpdatedAt:           m.UpdatedAt,
	}
}

type orderDetailModel struct {
	bun.BaseModel `bun:"table:orders,alias:o"`

	ID                  uuid.UUID               `bun:"id,pk"`
	CustomerID          uuid.UUID               `bun:"customer_id"`
	ArtistID            uuid.UUID               `bun:"artist_id"`
	Name                string                  `bun:"name"`
	ArtworkID           *uuid.UUID              `bun:"artwork_id"`
	ArtworkSnapshot     pgmodel.ArtworkSnapshot `bun:"artwork_snapshot"`
	PriceSatangOrder    int64                   `bun:"price_satang_order"`
	CustomerDescription string                  `bun:"customer_description"`
	DeadlineAt          time.Time               `bun:"deadline_at"`
	Status              string                  `bun:"status"`
	CompletedAt         *time.Time              `bun:"completed_at"`
	CreatedAt           time.Time               `bun:"created_at,nullzero"`
	UpdatedAt           time.Time               `bun:"updated_at,nullzero"`

	CustomerName  string `bun:"customer_name,scanonly"`
	CustomerEmail string `bun:"customer_email,scanonly"`

	ArtistName  string `bun:"artist_name,scanonly"`
	ArtistEmail string `bun:"artist_email,scanonly"`

	ArtistReviewScore *float64 `bun:"artist_review_score,scanonly"`
}

type orderRepository struct {
	exec baserepo.Executor
}

var _ order.OrderRepository = (*orderRepository)(nil)

// NewOrderRepository creates a PostgreSQL-backed order repository.
func NewOrderRepository(db *bun.DB) order.OrderRepository {
	return &orderRepository{exec: baserepo.NewExecutor(db)}
}

// orderSortColumns maps each ViewOrders sort field to the one trusted,
// allowlisted SQL column used for ORDER BY.
var orderSortColumns = map[order.SortField]string{
	order.SortFieldUpdatedAt: "o.updated_at",
	order.SortFieldPrice:     "o.price_satang_order",
	order.SortFieldDeadline:  "o.deadline_at",
}

// applyStatusFilter adds a status predicate when one or more statuses are supplied.
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

	page, err := baserepo.Paginate[pgmodel.Order](ctx, r.exec, func(q *bun.SelectQuery) *bun.SelectQuery {
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
		orders[i] = orderModelToDomain(&m)
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
		var deliverables []pgmodel.OrderDeliverable
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

// GetOrderByID returns an order's stored details and both participants when
// the requested participant ID belongs to that order.
func (r *orderRepository) GetOrderByID(
	ctx context.Context,
	participant order.Participant,
	participantID uuid.UUID,
	orderID uuid.UUID,
) (*order.OrderDetailData, error) {
	if participant != order.ParticipantCustomer &&
		participant != order.ParticipantArtist {
		return nil, apperror.Internal("unsupported participant", nil)
	}

	var model orderDetailModel

	err := r.exec.Run(ctx, func(idb bun.IDB) error {
		q := idb.NewSelect().
			Model(&model).
			Column(
				"o.id",
				"o.customer_id",
				"o.artist_id",
				"o.name",
				"o.artwork_id",
				"o.artwork_snapshot",
				"o.price_satang_order",
				"o.customer_description",
				"o.deadline_at",
				"o.status",
				"o.completed_at",
				"o.created_at",
				"o.updated_at",
			).
			ColumnExpr("customer.username AS customer_name").
			ColumnExpr("customer.email AS customer_email").
			ColumnExpr("artist.username AS artist_name").
			ColumnExpr("artist.email AS artist_email").
			ColumnExpr("rs.review_score AS artist_review_score").
			Join("JOIN users AS customer ON customer.id = o.customer_id").
			Join("JOIN users AS artist ON artist.id = o.artist_id").
			Join("LEFT JOIN ("+
				"SELECT artist_id, "+
				"ROUND(AVG(rating)::numeric, 1)::double precision AS review_score "+
				"FROM reviews "+
				"GROUP BY artist_id"+
				") AS rs ON rs.artist_id = o.artist_id").
			Where("o.id = ?", orderID)

		if participant == order.ParticipantCustomer {
			q = q.Where("o.customer_id = ?", participantID)
		} else {
			q = q.Where("o.artist_id = ?", participantID)
		}

		return q.Scan(ctx)
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, order.ErrOrderNotFound
		}

		if _, ok := errors.AsType[*apperror.Error](err); ok {
			return nil, err
		}

		return nil, apperror.Internal("failed to get order", err)
	}

	return &order.OrderDetailData{
		Order: order.Order{
			ID:                  model.ID,
			CustomerID:          model.CustomerID,
			ArtistID:            model.ArtistID,
			Name:                model.Name,
			ArtworkID:           model.ArtworkID,
			PriceSatangOrder:    model.PriceSatangOrder,
			CustomerDescription: model.CustomerDescription,
			DeadlineAt:          model.DeadlineAt,
			Status:              order.Status(model.Status),
			CompletedAt:         model.CompletedAt,
			CreatedAt:           model.CreatedAt,
			UpdatedAt:           model.UpdatedAt,
		},
		ArtworkSnapshot: order.ArtworkSnapshot{
			ArtworkName: model.ArtworkSnapshot.ArtworkName,
			CategoryID:  model.ArtworkSnapshot.CategoryID,
			StyleIDs:    model.ArtworkSnapshot.StyleIDs,
		},
		Customer: order.OrderParty{
			ID:    model.CustomerID,
			Name:  model.CustomerName,
			Email: model.CustomerEmail,
		},
		Artist: order.OrderParty{
			ID:                model.ArtistID,
			Name:              model.ArtistName,
			Email:             model.ArtistEmail,
			ArtistReviewScore: model.ArtistReviewScore,
		},
	}, nil
}
