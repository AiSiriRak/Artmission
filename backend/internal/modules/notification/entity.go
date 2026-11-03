// Package notification owns personal notification feeds and read status.
package notification

import (
	"time"

	"github.com/google/uuid"
)

// Type lets the frontend select a card icon without interpreting its title or message.
type Type string

const (
	TypeGeneral           Type = "GENERAL"
	TypeOrderCreated      Type = "ORDER_CREATED"
	TypeOrderAccepted     Type = "ORDER_ACCEPTED"
	TypePaymentSecured    Type = "PAYMENT_SECURED"
	TypeOrderDeclined     Type = "ORDER_DECLINED"
	TypeOrderCanceled     Type = "ORDER_CANCELED"
	TypeDraftSubmitted    Type = "DRAFT_SUBMITTED"
	TypeRevisionRequested Type = "REVISION_REQUESTED"
	TypeOrderCompleted    Type = "ORDER_COMPLETED"
	TypeFinalArtworkReady Type = "FINAL_ARTWORK_READY"
	TypeDeadlineRequested Type = "DEADLINE_EXTENSION_REQUESTED"
	TypeDeadlineApproved  Type = "DEADLINE_EXTENSION_APPROVED"
	TypeDeadlineRejected  Type = "DEADLINE_EXTENSION_REJECTED"
)

func (e Type) IsValid() bool {
	switch e {
	case TypeGeneral, TypeOrderCreated, TypeOrderAccepted, TypePaymentSecured, TypeOrderDeclined, TypeOrderCanceled,
		TypeDraftSubmitted, TypeRevisionRequested, TypeOrderCompleted, TypeFinalArtworkReady,
		TypeDeadlineRequested, TypeDeadlineApproved, TypeDeadlineRejected:
		return true
	}
	return false
}

// Notification is a message addressed to one recipient.
type Notification struct {
	ID uuid.UUID
	// UserID identifies the recipient, rather than the actor who triggered the activity.
	UserID         uuid.UUID
	Title          string
	Message        string
	Type           Type
	IsRead         bool
	CreatedAt      time.Time
	DeliverableID  *uuid.UUID
	OrderID        *uuid.UUID
	DelayRequestID *uuid.UUID
	// These fields are derived from linked records rather than stored on notifications.
	DeadlineAt         *time.Time
	DeliverableVersion *int
}
