package postgres

import (
	"context"
	pgmodel "github.com/AiSiriRak/Artmission/backend/internal/adapters/postgres/model"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/notification"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/baserepo"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type notificationRepository struct {
	base baserepo.BaseRepo[pgmodel.Notification]
	exec baserepo.Executor
}

var _ notification.NotificationRepository = (*notificationRepository)(nil)

func NewNotificationRepository(db *bun.DB) notification.NotificationRepository {
	return &notificationRepository{base: baserepo.NewBaseRepo[pgmodel.Notification](db, "notification"), exec: baserepo.NewExecutor(db)}
}
func (r *notificationRepository) Create(ctx context.Context, n *notification.Notification) error {
	m := &pgmodel.Notification{
		ID: n.ID, UserID: n.UserID, Title: n.Title, Message: n.Message, Type: string(n.Type), IsRead: n.IsRead, CreatedAt: n.CreatedAt,
		EventType: string(n.EventType), OrderID: n.OrderID, OrderStatus: (*string)(n.OrderStatus), DeadlineAt: n.DeadlineAt,
	}
	if err := r.base.Create(ctx, m); err != nil {
		return apperror.Internal("failed to create notification", err)
	}
	return nil
}
func (r *notificationRepository) ListByUserID(ctx context.Context, id uuid.UUID, q notification.ListQuery) (notification.Page, error) {
	page, err := baserepo.Paginate[pgmodel.Notification](ctx, r.exec, func(s *bun.SelectQuery) *bun.SelectQuery {
		return s.Column("n.id", "n.user_id", "n.title", "n.message", "n.type", "n.is_read",
			"n.created_at", "n.event_type", "n.order_id", "n.order_status", "n.deadline_at").
			ColumnExpr("o.deadline_at AS current_order_deadline_at").
			Join("LEFT JOIN orders AS o ON o.id = n.order_id AND n.user_id IN (o.customer_id, o.artist_id)").
			Where("n.user_id = ?", id).OrderExpr("n.created_at DESC, n.id DESC")
	}, baserepo.PaginationInput{Limit: q.Limit, Offset: q.Offset})
	if err != nil {
		return notification.Page{}, apperror.Internal("failed to list notifications", err)
	}
	items := make([]notification.Notification, 0, len(page.Items))
	for _, m := range page.Items {
		deadline := m.DeadlineAt
		if m.CurrentOrderDeadlineAt != nil {
			deadline = m.CurrentOrderDeadlineAt
		}
		items = append(items, notification.Notification{
			ID: m.ID, UserID: m.UserID, Title: m.Title, Message: m.Message, Type: notification.Type(m.Type), IsRead: m.IsRead, CreatedAt: m.CreatedAt,
			EventType: notification.EventType(m.EventType), OrderID: m.OrderID, OrderStatus: (*order.Status)(m.OrderStatus), DeadlineAt: deadline,
		})
	}
	return notification.Page{Items: items, Total: page.Total}, nil
}
func (r *notificationRepository) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	return r.exec.Run(ctx, func(db bun.IDB) error {
		result, err := db.NewUpdate().Model((*pgmodel.Notification)(nil)).Set("is_read = ?", true).Where("id = ? AND user_id = ?", id, userID).Exec(ctx)
		if err != nil {
			return apperror.Internal("failed to mark notification read", err)
		}
		count, err := result.RowsAffected()
		if err != nil {
			return apperror.Internal("failed to check notification update", err)
		}
		if count == 0 {
			return notification.ErrNotFound
		}
		return nil
	})
}
func (r *notificationRepository) CountUnread(ctx context.Context, id uuid.UUID) (int, error) {
	var count int
	err := r.exec.Run(ctx, func(db bun.IDB) error {
		var err error
		count, err = db.NewSelect().Model((*pgmodel.Notification)(nil)).Where("user_id = ? AND is_read = ?", id, false).Count(ctx)
		return err
	})
	if err != nil {
		return 0, apperror.Internal("failed to count unread notifications", err)
	}
	return count, nil
}
