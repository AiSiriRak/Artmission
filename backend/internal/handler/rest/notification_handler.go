package rest

import (
	"context"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/auth"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/notification"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type NotificationHandler struct {
	usecase     notification.NotificationUsecase
	authUsecase auth.AuthUsecase
}

func NewNotificationHandler(uc notification.NotificationUsecase, a auth.AuthUsecase) *NotificationHandler {
	return &NotificationHandler{usecase: uc, authUsecase: a}
}
func (h *NotificationHandler) Register(api huma.API) {
	operation := func(id, summary, description string) func(*huma.Operation) {
		return func(o *huma.Operation) {
			o.OperationID = id
			o.Summary = summary
			o.Description = description
			o.Middlewares = append(o.Middlewares, requireAuth(api, h.authUsecase))
		}
	}
	huma.Get(api, "/notifications", h.listMine, huma.OperationTags("notifications"), operation("list-notifications", "ListNotifications", "List your notifications, newest first"))
	huma.Get(api, "/notifications/unread-count", h.countUnread, huma.OperationTags("notifications"), operation("count-unread-notifications", "CountUnreadNotifications", "Count your unread notifications"))
	huma.Patch(api, "/notifications/{id}/read", h.markRead, huma.OperationTags("notifications"), operation("mark-notification-read", "MarkNotificationRead", "Mark your notification read"), func(o *huma.Operation) { o.DefaultStatus = http.StatusNoContent })
}

type notificationView struct {
	ID          uuid.UUID              `json:"id"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Type        notification.Type      `json:"type"`
	IsRead      bool                   `json:"is_read"`
	CreatedAt   time.Time              `json:"created_at"`
	EventType   notification.EventType `json:"event_type"`
	OrderID     *uuid.UUID             `json:"order_id"`
	OrderStatus *order.Status          `json:"order_status"`
	DeadlineAt  *time.Time             `json:"deadline_at"`
}
type listNotificationsInput struct {
	Limit  int `query:"limit" minimum:"1" maximum:"100" default:"20"`
	Offset int `query:"offset" minimum:"0" default:"0"`
}
type listNotificationsOutput struct {
	Body struct {
		Items []notificationView `json:"items"`
		Total int                `json:"total"`
	}
}

func (h *NotificationHandler) listMine(ctx context.Context, in *listNotificationsInput) (*listNotificationsOutput, error) {
	info, ok := authInfoFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("missing authentication")
	}
	page, err := h.usecase.ListMine(ctx, info.UserID, notification.ListQuery{Limit: in.Limit, Offset: in.Offset})
	if err != nil {
		return nil, mapAppError(err)
	}
	out := &listNotificationsOutput{}
	out.Body.Items = make([]notificationView, 0, len(page.Items))
	out.Body.Total = page.Total
	for _, n := range page.Items {
		out.Body.Items = append(out.Body.Items, notificationView{
			ID: n.ID, Title: n.Title, Message: n.Message, Type: n.Type, IsRead: n.IsRead, CreatedAt: n.CreatedAt,
			EventType: n.EventType, OrderID: n.OrderID, OrderStatus: n.OrderStatus, DeadlineAt: n.DeadlineAt,
		})
	}
	return out, nil
}

type markNotificationReadInput struct {
	ID uuid.UUID `path:"id"`
}

func (h *NotificationHandler) markRead(ctx context.Context, in *markNotificationReadInput) (*struct{}, error) {
	info, ok := authInfoFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("missing authentication")
	}
	if err := h.usecase.MarkRead(ctx, in.ID, info.UserID); err != nil {
		return nil, mapAppError(err)
	}
	return nil, nil
}

type unreadNotificationsOutput struct {
	Body struct {
		Count int `json:"count"`
	}
}

func (h *NotificationHandler) countUnread(ctx context.Context, _ *struct{}) (*unreadNotificationsOutput, error) {
	info, ok := authInfoFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("missing authentication")
	}
	count, err := h.usecase.CountUnread(ctx, info.UserID)
	if err != nil {
		return nil, mapAppError(err)
	}
	out := &unreadNotificationsOutput{}
	out.Body.Count = count
	return out, nil
}
