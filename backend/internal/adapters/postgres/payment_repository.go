package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	pgmodel "github.com/AiSiriRak/Artmission/backend/internal/adapters/postgres/model"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/payment"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/baserepo"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/uptrace/bun"
)

type paymentRepository struct {
	db   *bun.DB
	exec baserepo.Executor
	tx   baserepo.Transactioner
}

var _ payment.Repository = (*paymentRepository)(nil)

func NewPaymentRepository(db *bun.DB) payment.Repository {
	return &paymentRepository{
		db:   db,
		exec: baserepo.NewExecutor(db),
		tx:   baserepo.NewTransactioner(db),
	}
}

func (r *paymentRepository) GetOrderForPayment(
	ctx context.Context,
	customerID, orderID uuid.UUID,
) (*payment.OrderForPayment, error) {
	model := new(pgmodel.Order)
	err := r.exec.Run(ctx, func(idb bun.IDB) error {
		return idb.NewSelect().
			Model(model).
			Column("o.id", "o.status").
			Where("o.id = ?", orderID).
			Where("o.customer_id = ?", customerID).
			Scan(ctx)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, payment.ErrOrderNotFound
		}
		return nil, apperror.Internal("failed to load order for payment", err)
	}
	return &payment.OrderForPayment{ID: model.ID, Status: model.Status}, nil
}

func (r *paymentRepository) GetQRForOrder(ctx context.Context, orderID uuid.UUID) (*payment.Payment, error) {
	model := new(pgmodel.Payment)
	err := r.exec.Run(ctx, func(idb bun.IDB) error {
		return idb.NewSelect().
			Model(model).
			Join("JOIN orders AS o ON o.id = p.order_id").
			Where("p.order_id = ?", orderID).
			Where("o.status = ?", payment.EligibleOrderStatus).
			Scan(ctx)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, payment.ErrQRNotReady
		}
		return nil, apperror.Internal("failed to load payment QR", err)
	}
	return paymentFromModel(model), nil
}

func (r *paymentRepository) GetQRGenerationCandidate(
	ctx context.Context,
	paymentID uuid.UUID,
) (*payment.QRGenerationCandidate, error) {
	model := new(pgmodel.Payment)
	err := r.exec.Run(ctx, func(idb bun.IDB) error {
		return idb.NewSelect().
			Model(model).
			Column("p.id", "p.order_id", "p.amount_satang").
			Join("JOIN orders AS o ON o.id = p.order_id").
			Where("p.id = ?", paymentID).
			Where("p.status = ?", string(payment.StatusPending)).
			Where("p.qr_payload IS NULL").
			Where("p.qr_expire > clock_timestamp()").
			Where("o.status = ?", payment.EligibleOrderStatus).
			Scan(ctx)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, apperror.Internal("failed to reload payment for QR generation", err)
	}
	return &payment.QRGenerationCandidate{
		PaymentID:    model.ID,
		OrderID:      model.OrderID,
		AmountSatang: model.AmountSatang,
	}, nil
}

func (r *paymentRepository) TryLockQRGeneration(
	ctx context.Context,
	paymentID uuid.UUID,
) (payment.QRGenerationLock, bool, error) {
	// A single short statement claims the payment; no pooled connection is held
	// while the provider is called. A crashed holder is superseded after the TTL.
	result, err := r.db.NewUpdate().
		Model((*pgmodel.Payment)(nil)).
		Set("qr_generating_at = clock_timestamp()").
		Where("id = ?", paymentID).
		Where("(qr_generating_at IS NULL OR qr_generating_at < clock_timestamp() - make_interval(secs => ?))",
			qrGenerationClaimTTL.Seconds()).
		Exec(ctx)
	if err != nil {
		return nil, false, apperror.Internal("failed to claim QR generation", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return nil, false, apperror.Internal("failed to claim QR generation", err)
	}
	if rows != 1 {
		return nil, false, nil
	}
	return &qrGenerationClaim{db: r.db, paymentID: paymentID}, true, nil
}

func (r *paymentRepository) SaveQRPayload(
	ctx context.Context,
	paymentID uuid.UUID,
	payload string,
) (bool, error) {
	var rowsAffected int64
	err := r.exec.Run(ctx, func(idb bun.IDB) error {
		result, err := idb.NewUpdate().
			Model((*pgmodel.Payment)(nil)).
			Set("qr_payload = ?", payload).
			Set("qr_generating_at = NULL").
			Set("updated_at = clock_timestamp()").
			Where("id = ?", paymentID).
			Where("status = ?", string(payment.StatusPending)).
			Where("qr_payload IS NULL").
			Where("qr_expire > clock_timestamp()").
			Where("EXISTS (SELECT 1 FROM orders AS o WHERE o.id = p.order_id AND o.status = ?)", payment.EligibleOrderStatus).
			Exec(ctx)
		if err != nil {
			return err
		}
		rowsAffected, err = result.RowsAffected()
		return err
	})
	if err != nil {
		return false, apperror.Internal("failed to persist payment QR", err)
	}
	return rowsAffected == 1, nil
}

func (r *paymentRepository) ListExpiredPaymentIDs(ctx context.Context, limit int) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.exec.Run(ctx, func(idb bun.IDB) error {
		return idb.NewSelect().
			TableExpr("payments AS p").
			Column("p.id").
			Where("p.status = ?", string(payment.StatusPending)).
			Where("p.qr_expire <= clock_timestamp()").
			Order("p.qr_expire ASC").
			Limit(limit).
			Scan(ctx, &ids)
	})
	if err != nil {
		return nil, apperror.Internal("failed to list expired payments", err)
	}
	return ids, nil
}

func (r *paymentRepository) ExpirePayment(
	ctx context.Context,
	paymentID uuid.UUID,
) (orderID uuid.UUID, cancelled bool, expired bool, err error) {
	err = r.tx.Transaction(ctx, func(ctx context.Context) error {
		return r.exec.Run(ctx, func(idb bun.IDB) error {
			var paymentOrderID uuid.UUID
			if err := idb.NewSelect().
				TableExpr("payments AS p").
				Column("p.order_id").
				Where("p.id = ?", paymentID).
				Scan(ctx, &paymentOrderID); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return nil
				}
				return err
			}

			lockedOrder := new(pgmodel.Order)
			if err := idb.NewSelect().
				Model(lockedOrder).
				Column("o.id", "o.status").
				Where("o.id = ?", paymentOrderID).
				For("UPDATE").
				Scan(ctx); err != nil {
				return err
			}

			lockedPayment := new(pgmodel.Payment)
			if err := idb.NewSelect().
				Model(lockedPayment).
				Column("p.id", "p.order_id", "p.status", "p.qr_expire").
				Where("p.id = ?", paymentID).
				For("UPDATE").
				Scan(ctx); err != nil {
				return err
			}
			if lockedPayment.Status != string(payment.StatusPending) || lockedPayment.QRExpire == nil {
				return nil
			}

			result, err := idb.NewUpdate().
				Model((*pgmodel.Payment)(nil)).
				Set("status = ?", string(payment.StatusExpired)).
				Set("qr_payload = NULL").
				Set("updated_at = clock_timestamp()").
				Where("id = ?", paymentID).
				Where("status = ?", string(payment.StatusPending)).
				Where("qr_expire <= clock_timestamp()").
				Exec(ctx)
			if err != nil {
				return err
			}
			rows, err := result.RowsAffected()
			if err != nil {
				return err
			}
			if rows == 0 {
				return nil
			}

			orderID = paymentOrderID
			expired = true
			if lockedOrder.Status != payment.EligibleOrderStatus {
				return nil
			}

			result, err = idb.NewUpdate().
				Model((*pgmodel.Order)(nil)).
				Set("status = ?", "CANCEL").
				Set("updated_at = clock_timestamp()").
				Where("id = ?", paymentOrderID).
				Where("status = ?", payment.EligibleOrderStatus).
				Exec(ctx)
			if err != nil {
				return err
			}
			rows, err = result.RowsAffected()
			if err != nil {
				return err
			}
			cancelled = rows == 1
			if cancelled {
				if _, err := idb.NewInsert().Model(&pgmodel.OrderCancellation{
					ID:          uuid.New(),
					OrderID:     paymentOrderID,
					Reason:      "OTHER",
					Explanation: "Automatically cancelled because the payment QR expired after 24 hours.",
					CreatedAt:   time.Now(),
				}).Exec(ctx); err != nil {
					return err
				}
			}
			return nil
		})
	})
	if err != nil {
		if isPaymentDomainError(err) {
			return orderID, false, false, err
		}
		return orderID, false, false, apperror.Internal("failed to expire payment QR", err)
	}
	return orderID, cancelled, expired, nil
}

func (r *paymentRepository) GetPaymentForVerification(
	ctx context.Context,
	customerID, orderID uuid.UUID,
) (*payment.PaymentForVerification, error) {
	var item struct {
		PaymentID    uuid.UUID  `bun:"payment_id"`
		OrderID      uuid.UUID  `bun:"order_id"`
		AmountSatang int64      `bun:"amount_satang"`
		Status       string     `bun:"payment_status"`
		OrderStatus  string     `bun:"order_status"`
		QRPayload    *string    `bun:"qr_payload"`
		QRExpire     *time.Time `bun:"qr_expire"`
		CreatedAt    time.Time  `bun:"created_at"`
	}
	err := r.exec.Run(ctx, func(idb bun.IDB) error {
		return idb.NewSelect().
			TableExpr("payments AS p").
			ColumnExpr("p.id AS payment_id").
			ColumnExpr("p.order_id").
			ColumnExpr("p.amount_satang").
			ColumnExpr("p.status AS payment_status").
			ColumnExpr("o.status AS order_status").
			ColumnExpr("p.qr_payload").
			ColumnExpr("p.qr_expire").
			ColumnExpr("p.created_at").
			Join("JOIN orders AS o ON o.id = p.order_id").
			Where("p.order_id = ?", orderID).
			Where("o.customer_id = ?", customerID).
			Scan(ctx, &item)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, payment.ErrOrderNotFound
		}
		return nil, apperror.Internal("failed to load payment for slip verification", err)
	}
	return &payment.PaymentForVerification{
		PaymentID:    item.PaymentID,
		OrderID:      item.OrderID,
		AmountSatang: item.AmountSatang,
		Status:       payment.Status(item.Status),
		OrderStatus:  item.OrderStatus,
		QRPayload:    item.QRPayload,
		QRExpire:     item.QRExpire,
		CreatedAt:    item.CreatedAt,
	}, nil
}

func (r *paymentRepository) ConfirmSlipPayment(
	ctx context.Context,
	customerID, orderID uuid.UUID,
	transRef string,
) (*payment.VerifiedPayment, error) {
	var verified payment.VerifiedPayment
	err := r.tx.Transaction(ctx, func(ctx context.Context) error {
		return r.exec.Run(ctx, func(idb bun.IDB) error {
			lockedOrder := new(pgmodel.Order)
			if err := idb.NewSelect().
				Model(lockedOrder).
				Column("o.id", "o.customer_id", "o.status").
				Where("o.id = ?", orderID).
				Where("o.customer_id = ?", customerID).
				For("UPDATE").
				Scan(ctx); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return payment.ErrOrderNotFound
				}
				return err
			}
			if lockedOrder.Status != payment.EligibleOrderStatus {
				return payment.ErrOrderNotPayable
			}

			lockedPayment := new(pgmodel.Payment)
			if err := idb.NewSelect().
				Model(lockedPayment).
				Column("p.id", "p.order_id", "p.amount_satang", "p.status", "p.qr_payload", "p.qr_expire").
				Where("p.order_id = ?", orderID).
				For("UPDATE").
				Scan(ctx); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return payment.ErrQRNotReady
				}
				return err
			}
			if lockedPayment.Status != string(payment.StatusPending) {
				return payment.ErrPaymentNotVerifiable
			}
			if lockedPayment.QRExpire == nil {
				return payment.ErrQRExpired
			}
			var isUnexpired bool
			if err := idb.NewSelect().
				ColumnExpr("? > clock_timestamp()", lockedPayment.QRExpire).
				Scan(ctx, &isUnexpired); err != nil {
				return err
			}
			if !isUnexpired {
				return payment.ErrQRExpired
			}
			if lockedPayment.QRPayload == nil || *lockedPayment.QRPayload == "" {
				return payment.ErrQRNotReady
			}

			update, err := idb.NewUpdate().
				Model((*pgmodel.Payment)(nil)).
				Set("status = ?", string(payment.StatusHeld)).
				Set("qr_payload = NULL").
				Set("provider_transaction_ref = ?", transRef).
				Set("updated_at = clock_timestamp()").
				Where("id = ?", lockedPayment.ID).
				Where("status = ?", string(payment.StatusPending)).
				Where("qr_expire > clock_timestamp()").
				Exec(ctx)
			if err != nil {
				return err
			}
			rows, err := update.RowsAffected()
			if err != nil {
				return err
			}
			if rows != 1 {
				return payment.ErrQRExpired
			}

			if _, err := idb.NewInsert().Model(&pgmodel.PaymentTransaction{
				ID:           uuid.New(),
				PaymentID:    lockedPayment.ID,
				Type:         "CHARGE",
				AmountSatang: lockedPayment.AmountSatang,
				CreatedAt:    time.Now(),
			}).Exec(ctx); err != nil {
				return err
			}

			update, err = idb.NewUpdate().
				Model((*pgmodel.Order)(nil)).
				Set("status = ?", "IN_PROCESS").
				Set("updated_at = clock_timestamp()").
				Where("id = ?", orderID).
				Where("customer_id = ?", customerID).
				Where("status = ?", payment.EligibleOrderStatus).
				Exec(ctx)
			if err != nil {
				return err
			}
			rows, err = update.RowsAffected()
			if err != nil {
				return err
			}
			if rows != 1 {
				return payment.ErrOrderNotPayable
			}

			verified = payment.VerifiedPayment{
				PaymentID: lockedPayment.ID,
				OrderID:   orderID,
				Status:    payment.StatusHeld,
			}
			return nil
		})
	})
	if err != nil {
		if isPaymentDomainError(err) {
			return nil, err
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == pgerrcode.UniqueViolation &&
			pgErr.ConstraintName == "payments_provider_transaction_ref_key" {
			return nil, payment.ErrDuplicateSlip
		}
		return nil, apperror.Internal("failed to confirm slip payment", err)
	}
	return &verified, nil
}

func isPaymentDomainError(err error) bool {
	var appErr *apperror.Error
	return errors.As(err, &appErr)
}

// qrGenerationClaimTTL must exceed the provider client timeout so a live
// generation is never superseded.
const qrGenerationClaimTTL = 30 * time.Second

type qrGenerationClaim struct {
	db        *bun.DB
	paymentID uuid.UUID
}

// Release clears the claim. If it fails, the claim simply expires after the TTL.
func (c *qrGenerationClaim) Release() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := c.db.NewUpdate().
		Model((*pgmodel.Payment)(nil)).
		Set("qr_generating_at = NULL").
		Where("id = ?", c.paymentID).
		Exec(ctx); err != nil {
		return fmt.Errorf("release QR generation claim: %w", err)
	}
	return nil
}

func paymentFromModel(m *pgmodel.Payment) *payment.Payment {
	return &payment.Payment{
		ID:                     m.ID,
		OrderID:                m.OrderID,
		AmountSatang:           m.AmountSatang,
		Status:                 payment.Status(m.Status),
		QRPayload:              m.QRPayload,
		QRExpire:               m.QRExpire,
		ProviderTransactionRef: m.ProviderTransactionRef,
		CreatedAt:              m.CreatedAt,
		UpdatedAt:              m.UpdatedAt,
	}
}
