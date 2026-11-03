package payment

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Usecase interface {
	GetQR(ctx context.Context, customerID, orderID uuid.UUID) (*QRPayment, error)
	GenerateQR(ctx context.Context, customerID, orderID uuid.UUID) (*QRPayment, error)
	ExpirePayments(ctx context.Context) ([]uuid.UUID, error)
	VerifySlip(
		ctx context.Context,
		customerID, orderID uuid.UUID,
		image []byte,
		contentType string,
	) (*VerifiedPayment, error)
}

type Repository interface {
	GetOrderForPayment(ctx context.Context, customerID, orderID uuid.UUID) (*OrderForPayment, error)
	GetQRForOrder(ctx context.Context, orderID uuid.UUID) (*Payment, error)
	GetQRGenerationCandidate(ctx context.Context, paymentID uuid.UUID) (*QRGenerationCandidate, error)
	TryLockQRGeneration(ctx context.Context, paymentID uuid.UUID) (QRGenerationLock, bool, error)
	SaveQRPayload(ctx context.Context, paymentID uuid.UUID, payload string) (bool, error)
	ListExpiredPaymentIDs(ctx context.Context, limit int) ([]uuid.UUID, error)
	ExpirePayment(ctx context.Context, paymentID uuid.UUID) (orderID uuid.UUID, cancelled bool, expired bool, err error)
	GetPaymentForVerification(
		ctx context.Context,
		customerID, orderID uuid.UUID,
	) (*PaymentForVerification, error)
	ConfirmSlipPayment(
		ctx context.Context,
		customerID, orderID uuid.UUID,
		transRef string,
	) (*VerifiedPayment, error)
}

type QRGenerationLock interface {
	Release() error
}

type QRProvider interface {
	GeneratePromptPayQR(ctx context.Context, amountSatang int64) (QRImage, error)
	VerifySlipImage(
		ctx context.Context,
		image []byte,
		contentType string,
		amountSatang int64,
		notBefore time.Time,
	) (SlipVerification, error)
}
