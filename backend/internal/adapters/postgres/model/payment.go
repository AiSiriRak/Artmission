package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Payment struct {
	bun.BaseModel `bun:"table:payments,alias:p"`

	ID                     uuid.UUID  `bun:"id,pk"`
	OrderID                uuid.UUID  `bun:"order_id"`
	AmountSatang           int64      `bun:"amount_satang"`
	Status                 string     `bun:"status"`
	QRPayload              *string    `bun:"qr_payload"`
	QRExpire               *time.Time `bun:"qr_expire"`
	ProviderTransactionRef *string    `bun:"provider_transaction_ref"`
	CreatedAt              time.Time  `bun:"created_at,nullzero"`
	UpdatedAt              time.Time  `bun:"updated_at,nullzero"`
}

type PaymentTransaction struct {
	bun.BaseModel `bun:"table:payment_transactions,alias:pt"`

	ID           uuid.UUID `bun:"id,pk"`
	PaymentID    uuid.UUID `bun:"payment_id"`
	Type         string    `bun:"type"`
	AmountSatang int64     `bun:"amount_satang"`
	CreatedAt    time.Time `bun:"created_at,nullzero"`
}
