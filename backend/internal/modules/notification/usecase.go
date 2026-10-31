package notification

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/google/uuid"
)

type notificationUsecase struct{ repo NotificationRepository }

// NewNotificationUsecase connects notification operations to their persistence dependency.
func NewNotificationUsecase(repo NotificationRepository) NotificationUsecase {
	return &notificationUsecase{repo: repo}
}

// Create validates the recipient and category, trims message text, and saves an unread notification.
func (u *notificationUsecase) Create(ctx context.Context, in CreateInput) (*Notification, error) {
	title, message := strings.TrimSpace(in.Title), strings.TrimSpace(in.Message)
	if in.UserID == uuid.Nil {
		return nil, ErrMissingUserID
	}
	if title == "" {
		return nil, ErrMissingTitle
	}
	if message == "" {
		return nil, ErrMissingMessage
	}
	switch in.Type {
	case TypeOrderUpdate, TypeReview, TypeSystem:
	default:
		return nil, ErrInvalidType
	}
	if in.EventType == "" {
		in.EventType = EventGeneral
	}
	if !in.EventType.IsValid() {
		return nil, ErrInvalidEventType
	}
	if in.OrderStatus != nil && *in.OrderStatus != order.StatusPending && *in.OrderStatus != order.StatusNotPaid {
		return nil, ErrInvalidOrderStatus
	}
	n := &Notification{
		ID: uuid.New(), UserID: in.UserID, Title: title, Message: message, Type: in.Type, CreatedAt: time.Now(),
		EventType: in.EventType, OrderID: in.OrderID, OrderStatus: in.OrderStatus, DeadlineAt: in.DeadlineAt,
	}
	if err := u.repo.Create(ctx, n); err != nil {
		return nil, err
	}
	return n, nil
}

// ListMine validates pagination and returns an empty slice when the recipient has no results.
func (u *notificationUsecase) ListMine(ctx context.Context, id uuid.UUID, q ListQuery) (Page, error) {
	if id == uuid.Nil {
		return Page{}, ErrMissingUserID
	}
	if q.Limit < 1 || q.Limit > 100 {
		return Page{}, ErrInvalidLimit
	}
	if q.Offset < 0 {
		return Page{}, ErrInvalidOffset
	}
	page, err := u.repo.ListByUserID(ctx, id, q)
	if page.Items == nil {
		page.Items = []Notification{}
	}
	return page, err
}

// MarkRead delegates the ownership check and read update to one repository operation.
func (u *notificationUsecase) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return ErrMissingUserID
	}
	if id == uuid.Nil {
		return ErrMissingNotificationID
	}
	return u.repo.MarkRead(ctx, id, userID)
}

// CountUnread returns the recipient's unread total independently of feed pagination.
func (u *notificationUsecase) CountUnread(ctx context.Context, id uuid.UUID) (int, error) {
	if id == uuid.Nil {
		return 0, ErrMissingUserID
	}
	return u.repo.CountUnread(ctx, id)
}

// NotificationCreator is satisfied by the notification module without exposing its feed operations.
type NotificationCreator interface {
	Create(context.Context, CreateInput) (*Notification, error)
}

type NotificationTransactioner interface {
	Transaction(context.Context, func(context.Context) error) error
}

type notifyingRepository struct {
	order.OrderRepository
	notifications NotificationCreator
	tx            NotificationTransactioner
}

// NewNotifyingRepository adds order notifications to successful writes.
// The write and notification share a transaction, including repository-owned nested operations.
func NewNotifyingRepository(repo order.OrderRepository, notifications NotificationCreator, tx NotificationTransactioner) order.OrderRepository {
	return &notifyingRepository{OrderRepository: repo, notifications: notifications, tx: tx}
}

func (r *notifyingRepository) notify(ctx context.Context, item order.Order, recipient uuid.UUID, event EventType, title, message string) error {
	var deadline *time.Time
	if !item.DeadlineAt.IsZero() {
		deadline = &item.DeadlineAt
	}
	_, err := r.notifications.Create(ctx, CreateInput{
		UserID: recipient, Type: TypeOrderUpdate, EventType: event,
		Title: fmt.Sprintf("%s: %s", title, item.Name), Message: message,
		OrderID: &item.ID, OrderStatus: &item.Status, DeadlineAt: deadline,
	})
	return err
}

// Create notifies the artist only for a newly pending order.
func (r *notifyingRepository) Create(ctx context.Context, item *order.Order) error {
	return r.tx.Transaction(ctx, func(ctx context.Context) error {
		if err := r.OrderRepository.Create(ctx, item); err != nil {
			return err
		}
		if item.Status != order.StatusPending {
			return nil
		}
		return r.notify(ctx, *item, item.ArtistID, EventOrderCreated, "New Order",
			"The customer has placed a new order. Please check the order details to review the brief.")
	})
}

// ConfirmOrder notifies the customer when artist acceptance makes payment due.
// Rejection and later workflow actions keep their original order behavior without generating notifications.
func (r *notifyingRepository) ConfirmOrder(ctx context.Context, artistID, orderID uuid.UUID, status order.Status) error {
	return r.tx.Transaction(ctx, func(ctx context.Context) error {
		if err := r.OrderRepository.ConfirmOrder(ctx, artistID, orderID, status); err != nil {
			return err
		}
		if status != order.StatusNotPaid {
			return nil
		}
		detail, err := r.OrderRepository.GetOrderByID(ctx, order.ParticipantArtist, artistID, orderID)
		if err != nil {
			return err
		}
		return r.notify(ctx, detail.Order, detail.Order.CustomerID, EventOrderAccepted, "Order Accepted",
			"The artist has accepted your order! Please complete the payment to secure your commission.")
	})
}
