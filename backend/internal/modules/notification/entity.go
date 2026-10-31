// Package notification owns personal notification feeds and read status.
package notification

import (
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/google/uuid"
)

// EventType lets the frontend select a card icon without interpreting its title or message.
type EventType string

const (
	EventGeneral       EventType = "GENERAL"
	EventOrderCreated  EventType = "ORDER_CREATED"
	EventOrderAccepted EventType = "ORDER_ACCEPTED"
)

func (e EventType) IsValid() bool {
	return e == EventGeneral || e == EventOrderCreated || e == EventOrderAccepted
}

// Type identifies the category of activity that produced a notification.
type Type string

const (
	TypeOrderUpdate Type = "ORDER_UPDATE"
	TypeReview      Type = "REVIEW"
	TypeSystem      Type = "SYSTEM"
)

// Notification is a message addressed to one recipient.
type Notification struct {
	ID uuid.UUID
	// UserID identifies the recipient, rather than the actor who triggered the activity.
	UserID    uuid.UUID
	Title     string
	Message   string
	Type      Type
	IsRead    bool
	CreatedAt time.Time
	EventType EventType
	OrderID   *uuid.UUID
	// OrderStatus records the status when the event occurred, even if the order changes later.
	OrderStatus *order.Status
	DeadlineAt  *time.Time
}
