package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Notification represents a user notification for order updates, system messages, or activities.
type Notification struct {
	bun.BaseModel      `bun:"table:notifications,alias:n"`
	ID                 uuid.UUID  `bun:"id,pk"`
	UserID             uuid.UUID  `bun:"user_id"`
	Title              string     `bun:"title"`
	Message            string     `bun:"message"`
	Type               string     `bun:"type"`
	IsRead             bool       `bun:"is_read"`
	CreatedAt          time.Time  `bun:"created_at,nullzero"`
	OrderID            *uuid.UUID `bun:"order_id"`
	DeliverableID      *uuid.UUID `bun:"deliverable_id"`
	DelayRequestID     *uuid.UUID `bun:"delay_request_id"`
	DeliverableVersion *int       `bun:"deliverable_version,scanonly"`
	DeadlineAt         *time.Time `bun:"deadline_at,scanonly"`
}
