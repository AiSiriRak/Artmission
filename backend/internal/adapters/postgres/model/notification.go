package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Notification represents a user notification for order updates, system messages, or activities.
type Notification struct {
	bun.BaseModel `bun:"table:notifications,alias:n"`

	ID          uuid.UUID  `bun:"id,pk"`
	UserID      uuid.UUID  `bun:"user_id"`
	Title       string     `bun:"title"`
	Message     string     `bun:"message"`
	Type        string     `bun:"type"`
	IsRead      bool       `bun:"is_read"`
	CreatedAt   time.Time  `bun:"created_at,nullzero"`
	EventType   string     `bun:"event_type,notnull,default:'GENERAL'"`
	OrderID     *uuid.UUID `bun:"order_id"`
	OrderStatus *string    `bun:"order_status"`
	DeadlineAt  *time.Time `bun:"deadline_at"`
	// CurrentOrderDeadlineAt is populated by feed reads and is not a table column.
	CurrentOrderDeadlineAt *time.Time `bun:"current_order_deadline_at,scanonly"`
}
