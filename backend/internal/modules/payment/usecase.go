package payment

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/google/uuid"
)

const (
	expirationBatchSize = 100
)

type paymentUsecase struct {
	repo     Repository
	provider QRProvider
	now      func() time.Time
	notifier Notifier
	tx       Transactioner
}

var _ Usecase = (*paymentUsecase)(nil)

func NewPaymentUsecase(repo Repository, provider QRProvider, options ...UsecaseOption) Usecase {
	u := &paymentUsecase{repo: repo, provider: provider, now: time.Now}
	for _, option := range options {
		option(u)
	}
	return u
}

type UsecaseOption func(*paymentUsecase)

func WithNotifications(notifier Notifier, tx Transactioner) UsecaseOption {
	return func(u *paymentUsecase) { u.notifier = notifier; u.tx = tx }
}

func (u *paymentUsecase) transaction(ctx context.Context, fn func(context.Context) error) error {
	if u.tx == nil {
		return fn(ctx)
	}
	return u.tx.Transaction(ctx, fn)
}

func (u *paymentUsecase) GetQR(
	ctx context.Context,
	customerID, orderID uuid.UUID,
) (*QRPayment, error) {
	if err := validateIDs(customerID, orderID); err != nil {
		return nil, err
	}

	order, err := u.repo.GetOrderForPayment(ctx, customerID, orderID)
	if err != nil {
		return nil, err
	}
	if order.Status != EligibleOrderStatus {
		return nil, ErrOrderNotPayable
	}

	item, err := u.repo.GetQRForOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if item.Status != StatusPending {
		return nil, ErrOrderNotPayable
	}
	if item.QRExpire == nil || !item.QRExpire.After(u.now()) {
		return nil, ErrQRExpired
	}
	if item.QRPayload == nil || strings.TrimSpace(*item.QRPayload) == "" {
		return nil, ErrQRNotReady
	}

	return &QRPayment{
		PaymentID: item.ID,
		OrderID:   item.OrderID,
		Payload:   *item.QRPayload,
		ExpiresAt: *item.QRExpire,
	}, nil
}

func (u *paymentUsecase) GenerateQR(
	ctx context.Context,
	customerID, orderID uuid.UUID,
) (resultQR *QRPayment, resultErr error) {
	if err := validateIDs(customerID, orderID); err != nil {
		return nil, err
	}

	// Fast path: an already stored QR is returned without touching Slip2Go.
	if qr, err := u.GetQR(ctx, customerID, orderID); err == nil || !errors.Is(err, ErrQRNotReady) {
		if qr != nil {
			qr.AlreadyExisted = true
		}
		return qr, err
	}

	item, err := u.repo.GetQRForOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	lock, acquired, err := u.repo.TryLockQRGeneration(ctx, item.ID)
	if err != nil {
		return nil, err
	}
	if !acquired {
		return nil, ErrQRInProgress
	}
	defer func() {
		if err := lock.Release(); err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("release QR generation lock: %w", err))
		}
	}()

	// Re-check under the lock: a concurrent request may have just stored the QR.
	candidate, err := u.repo.GetQRGenerationCandidate(ctx, item.ID)
	if err != nil {
		return nil, err
	}
	if candidate == nil {
		existing, err := u.GetQR(ctx, customerID, orderID)
		if existing != nil {
			existing.AlreadyExisted = true
		}
		return existing, err
	}
	if candidate.AmountSatang < 0 {
		return nil, errors.New("payment has a negative amount")
	}

	qr, err := u.provider.GeneratePromptPayQR(ctx, candidate.AmountSatang)
	if err != nil {
		return nil, apperror.Internal(
			fmt.Sprintf("generate payment QR for order %s", candidate.OrderID),
			err,
		)
	}
	if strings.TrimSpace(qr.Payload) == "" || qr.AmountSatang != candidate.AmountSatang {
		return nil, apperror.Internal(
			fmt.Sprintf("Slip2Go returned invalid QR data for order %s", candidate.OrderID),
			nil,
		)
	}

	// If the conditional save lost a race, GetQR reports the stored or ineligible state.
	saved, err := u.repo.SaveQRPayload(ctx, item.ID, qr.Payload)
	if err != nil {
		return nil, err
	}
	result, err := u.GetQR(ctx, customerID, orderID)
	if result != nil {
		result.AlreadyExisted = !saved
	}
	return result, err
}

func (u *paymentUsecase) ExpirePayments(ctx context.Context) ([]uuid.UUID, error) {
	paymentIDs, err := u.repo.ListExpiredPaymentIDs(ctx, expirationBatchSize)
	if err != nil {
		return nil, err
	}

	cancelledOrders := make([]uuid.UUID, 0, len(paymentIDs))
	var sweepErr error
	for _, paymentID := range paymentIDs {
		var orderID uuid.UUID
		var cancelled, expired bool
		err := u.transaction(ctx, func(ctx context.Context) error {
			var err error
			orderID, cancelled, expired, err = u.repo.ExpirePayment(ctx, paymentID)
			if err != nil {
				return err
			}
			if expired && cancelled && u.notifier != nil {
				return u.notifier.PaymentExpired(ctx, orderID)
			}
			return nil
		})
		if err != nil {
			sweepErr = errors.Join(sweepErr, err)
			continue
		}
		if expired && cancelled {
			cancelledOrders = append(cancelledOrders, orderID)
		}
	}
	return cancelledOrders, sweepErr
}

func (u *paymentUsecase) VerifySlip(
	ctx context.Context,
	customerID, orderID uuid.UUID,
	image []byte,
	contentType string,
) (*VerifiedPayment, error) {
	if err := validateIDs(customerID, orderID); err != nil {
		return nil, err
	}
	if len(image) == 0 {
		return nil, ErrInvalidSlipImage
	}

	item, err := u.repo.GetPaymentForVerification(ctx, customerID, orderID)
	if err != nil {
		return nil, err
	}
	if item.OrderStatus != EligibleOrderStatus {
		return nil, ErrOrderNotPayable
	}
	if item.Status != StatusPending {
		return nil, ErrPaymentNotVerifiable
	}
	if item.QRExpire == nil || !item.QRExpire.After(u.now()) {
		return nil, ErrQRExpired
	}
	if item.QRPayload == nil || strings.TrimSpace(*item.QRPayload) == "" {
		return nil, ErrQRNotReady
	}

	verification, err := u.provider.VerifySlipImage(ctx, image, contentType, item.AmountSatang, item.CreatedAt)
	if err != nil {
		return nil, apperror.Internal("slip verification provider failed", err)
	}
	if !verification.Valid {
		return nil, rejectionFromVerification(verification, item.AmountSatang)
	}
	if strings.TrimSpace(verification.TransRef) == "" ||
		verification.AmountSatang != item.AmountSatang {
		return nil, apperror.Internal("Slip2Go returned incomplete or mismatched verification data", nil)
	}

	// Provider verification stays outside the transaction; only database writes and notifications commit together.
	var result *VerifiedPayment
	err = u.transaction(ctx, func(ctx context.Context) error {
		var err error
		result, err = u.repo.ConfirmSlipPayment(ctx, customerID, orderID, verification.TransRef)
		if err != nil {
			return err
		}
		if u.notifier != nil {
			return u.notifier.PaymentSecured(ctx, orderID)
		}
		return nil
	})
	if errors.Is(err, ErrDuplicateSlip) {
		return nil, newSlipRejection("SLIP_ALREADY_USED", "This slip has already been used.")
	}
	if err != nil {
		// The provider already consumed this slip (checkDuplicate), so a retry would
		// be rejected as used. Log enough for support to recover manually.
		slog.ErrorContext(ctx, "slip verified by provider but payment was not recorded",
			"order_id", orderID, "payment_id", item.PaymentID,
			"trans_ref", verification.TransRef, "error", err)
		return nil, err
	}
	return result, nil
}

func validateIDs(customerID, orderID uuid.UUID) error {
	if customerID == uuid.Nil {
		return ErrMissingCustomerID
	}
	if orderID == uuid.Nil {
		return ErrMissingOrderID
	}
	return nil
}
