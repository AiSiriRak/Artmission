// Package payment owns payment creation and settlement lifecycle.
package payment

import (
	"time"

	"github.com/google/uuid"
)

const EligibleOrderStatus = "NOT_PAID"

const QRValidityPeriod = 24 * time.Hour

type Status string

const (
	StatusPending Status = "PENDING"
	StatusHeld    Status = "HELD"
	StatusSettled Status = "SETTLED"
	StatusFailed  Status = "FAILED"
	StatusExpired Status = "EXPIRED"
)

type Payment struct {
	ID                     uuid.UUID
	OrderID                uuid.UUID
	AmountSatang           int64
	Status                 Status
	QRPayload              *string
	QRExpire               *time.Time
	ProviderTransactionRef *string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type OrderForPayment struct {
	ID     uuid.UUID
	Status string
}

type QRImage struct {
	Payload      string
	AccountName  string
	AmountSatang int64
}

type QRPayment struct {
	PaymentID uuid.UUID
	OrderID   uuid.UUID
	Payload   string
	ExpiresAt time.Time
	// AlreadyExisted is set by GenerateQR when a stored QR was returned instead of a new one.
	AlreadyExisted bool
}

type QRGenerationCandidate struct {
	PaymentID    uuid.UUID
	OrderID      uuid.UUID
	AmountSatang int64
}

type PaymentForVerification struct {
	PaymentID    uuid.UUID
	OrderID      uuid.UUID
	AmountSatang int64
	Status       Status
	OrderStatus  string
	QRPayload    *string
	QRExpire     *time.Time
}

type SlipVerification struct {
	Valid        bool
	ResultCode   string
	TransRef     string
	AmountSatang int64
}

type VerifiedPayment struct {
	PaymentID uuid.UUID
	OrderID   uuid.UUID
	Status    Status
}
