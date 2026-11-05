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
	// StatusPending waits for successful payment verification; a failed slip check leaves it retryable.
	StatusPending Status = "PENDING"
	// StatusHeld has a verified charge waiting for the order outcome.
	StatusHeld Status = "HELD"
	// StatusSettled records completed payout/refund credits in the internal wallets.
	StatusSettled Status = "SETTLED"
	// StatusFailed is reserved for a failed payment; it is distinct from a rejected slip attempt.
	StatusFailed Status = "FAILED"
	// StatusExpired closes an unpaid payment when its payment window or order deadline expires.
	StatusExpired Status = "EXPIRED"
	// StatusCancelled closes only an unpaid PENDING payment when its order is explicitly cancelled.
	StatusCancelled Status = "CANCELLED"
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
	CreatedAt    time.Time
}

type SlipVerification struct {
	Valid        bool
	ResultCode   string
	TransRef     string
	AmountSatang int64
	DateTime     time.Time
}

type VerifiedPayment struct {
	PaymentID uuid.UUID
	OrderID   uuid.UUID
	Status    Status
}
