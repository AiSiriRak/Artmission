package notification

import (
	"context"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/google/uuid"
)

// ListQuery selects an offset page; Limit must be 1–100 and Offset nonnegative.
type ListQuery struct{ Limit, Offset int }

// Page contains one page of notifications and the total matching count before pagination.
type Page struct {
	Items []Notification
	Total int
}

// CreateInput supplies the recipient and message for an internally generated notification.
type CreateInput struct {
	UserID         uuid.UUID
	Title, Message string
	Type           Type
	EventType      EventType
	OrderID        *uuid.UUID
	OrderStatus    *order.Status
	DeadlineAt     *time.Time
}
type NotificationUsecase interface {
	// Create validates and trims the message, then saves a new unread notification.
	// This is called by backend features; it is not exposed as a public creation endpoint.
	Create(context.Context, CreateInput) (*Notification, error)
	// ListMine returns the recipient's notifications, newest first.
	ListMine(context.Context, uuid.UUID, ListQuery) (Page, error)
	// MarkRead takes the notification ID followed by the authenticated recipient ID.
	// Missing notifications and notifications owned by another user both return ErrNotFound.
	// Marking an already-read notification succeeds.
	MarkRead(context.Context, uuid.UUID, uuid.UUID) error
	// CountUnread counts unread notifications across all pages for the recipient.
	CountUnread(context.Context, uuid.UUID) (int, error)
}
type NotificationRepository interface {
	Create(context.Context, *Notification) error
	// ListByUserID filters by recipient and sorts by created_at DESC, then id DESC.
	ListByUserID(context.Context, uuid.UUID, ListQuery) (Page, error)
	// MarkRead must match both notification ID and recipient ID in the update.
	// Return ErrNotFound when no owned notification matches, including a missing ID.
	MarkRead(context.Context, uuid.UUID, uuid.UUID) error
	CountUnread(context.Context, uuid.UUID) (int, error)
}
