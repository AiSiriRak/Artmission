//go:build integration

package payments

import (
	"context"
	"testing"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/adapters/postgres"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/payment"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/config"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/database"
	"github.com/AiSiriRak/Artmission/backend/tests/internal/apptest"
	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
)

func TestTransferScenarios(t *testing.T) {
	ctx := context.Background()
	pg := apptest.StartPostgres(ctx, t)
	db, err := database.NewPostgresDB(config.Database{DSN: pg.DSN})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	cases := []struct {
		name           string
		initial        order.Status
		participant    order.Participant
		action         string
		paymentStatus  payment.Status
		outcome        payment.SettlementOutcome
		refund, payout int64
	}{
		{"TestTransfer01_ArtistRejectsPendingOrder", order.StatusPending, order.ParticipantArtist, "reject", "", "", 0, 0},
		{"TestTransfer02_CustomerCancelsPendingOrder", order.StatusPending, order.ParticipantCustomer, "cancel", "", "", 0, 0},
		{"TestTransfer03_ArtistCancelsNotPaidOrder", order.StatusNotPaid, order.ParticipantArtist, "cancel", payment.StatusCancelled, "", 0, 0},
		{"TestTransfer04_CustomerCancelsNotPaidOrder", order.StatusNotPaid, order.ParticipantCustomer, "cancel", payment.StatusCancelled, "", 0, 0},
		{"TestTransfer05_ArtistCancelsInProcessOrder", order.StatusInProcess, order.ParticipantArtist, "cancel", payment.StatusSettled, payment.ArtistCancelled, 10001, 0},
		{"TestTransfer06_CustomerCancelsInProcessOrder", order.StatusInProcess, order.ParticipantCustomer, "cancel", payment.StatusSettled, payment.CustomerCancelled, 5000, 5001},
		{"TestTransfer07_SuccessfulOrder", order.StatusInProcess, order.ParticipantCustomer, "approve", payment.StatusSettled, payment.Completed, 0, 10001},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t, db, 10001)
			mustExec(t, db, `UPDATE orders SET status=? WHERE id=?`, tc.initial, f.orderID)
			switch tc.initial {
			case order.StatusPending:
				mustExec(t, db, `DELETE FROM payments WHERE id=?`, f.paymentID)
			case order.StatusNotPaid:
				mustExec(t, db, `UPDATE payments SET status='PENDING',qr_payload='test-qr',qr_expire=NOW()+INTERVAL '1 hour',qr_generating_at=NOW() WHERE id=?`, f.paymentID)
			default:
				mustExec(t, db, `INSERT INTO payment_transactions(id,payment_id,type,amount_satang) VALUES(?,?,'CHARGE',10001)`, uuid.New(), f.paymentID)
			}
			u := order.NewOrderUsecase(f.orders, nil)
			deliverableID := uuid.New()
			if tc.action == "approve" {
				mustExec(t, db, `INSERT INTO order_deliverables(id,order_id,version,original_image_key,preview_image_key) VALUES(?,?,1,'original','preview')`, deliverableID, f.orderID)
			}
			act := func() error {
				switch tc.action {
				case "reject":
					_, err := u.ConfirmOrder(ctx, f.artist, f.orderID, order.ConfirmOrderInput{Accept: false})
					return err
				case "approve":
					_, err := u.ConfirmDeliverable(ctx, f.customer, order.ConfirmDeliverableInput{OrderID: f.orderID, DeliverableID: deliverableID, Accept: true})
					return err
				default:
					id := f.customer
					if tc.participant == order.ParticipantArtist {
						id = f.artist
					}
					_, err := u.CancelOrder(ctx, tc.participant, id, order.CancelInput{OrderID: f.orderID, Reason: order.CancelReasonOther, Explanation: "Test cancellation"})
					return err
				}
			}
			if err := act(); err != nil {
				t.Fatal(err)
			}
			if err := act(); err == nil {
				t.Fatal("terminal operation succeeded twice")
			}
			var status string
			if err := db.NewRaw(`SELECT status FROM orders WHERE id=?`, f.orderID).Scan(ctx, &status); err != nil {
				t.Fatal(err)
			}
			wantOrder := order.StatusCancel
			if tc.action == "approve" {
				wantOrder = order.StatusSuccess
			}
			if status != string(wantOrder) {
				t.Fatalf("order %s want %s", status, wantOrder)
			}
			p := f.summary(t)
			if tc.paymentStatus == "" {
				if p != nil {
					t.Fatal("fabricated payment for unaccepted order")
				}
			} else {
				if p == nil || p.Status != string(tc.paymentStatus) || p.RefundSatang != tc.refund || p.ArtistPayoutSatang != tc.payout {
					t.Fatalf("payment %+v", p)
				}
				if tc.outcome == "" {
					if p.SettlementOutcome != nil || p.SettledAt != nil {
						t.Fatal("settled unpaid payment")
					}
				} else if p.SettlementOutcome == nil || *p.SettlementOutcome != tc.outcome || p.SettledAt == nil {
					t.Fatalf("wrong outcome %+v", p)
				}
			}
			refundCount, payoutCount := 0, 0
			if tc.refund > 0 {
				refundCount = 1
			}
			if tc.payout > 0 {
				payoutCount = 1
			}
			assertWallet(t, f, f.customer, tc.refund, refundCount)
			assertWallet(t, f, f.artist, tc.payout, payoutCount)
			var ledgerCount int
			if err := db.NewRaw(`SELECT COUNT(*) FROM payment_transactions WHERE payment_id=?`, f.paymentID).Scan(ctx, &ledgerCount); err != nil {
				t.Fatal(err)
			}
			wantLedger := refundCount + payoutCount
			if tc.initial == order.StatusInProcess {
				wantLedger++
			}
			if ledgerCount != wantLedger {
				t.Fatalf("ledger rows %d want %d", ledgerCount, wantLedger)
			}
			if tc.paymentStatus == payment.StatusCancelled {
				var qr struct {
					QRPayload      *string
					QRGeneratingAt *time.Time
				}
				if err := db.NewRaw(`SELECT qr_payload,qr_generating_at FROM payments WHERE id=?`, f.paymentID).Scan(ctx, &qr); err != nil || qr.QRPayload != nil || qr.QRGeneratingAt != nil {
					t.Fatalf("QR not cleared %+v %v", qr, err)
				}
				repo := postgres.NewPaymentRepository(db)
				if _, err := repo.ConfirmSlipPayment(ctx, f.customer, f.orderID, uuid.NewString()); err == nil {
					t.Fatal("cancelled payment accepted slip")
				}
				mustExec(t, db, `UPDATE payments SET qr_expire=NOW()-INTERVAL '1 day' WHERE id=?`, f.paymentID)
				if _, _, expired, err := repo.ExpirePayment(ctx, f.paymentID); err != nil || expired {
					t.Fatalf("cancelled payment expired again: %t %v", expired, err)
				}
				if f.summary(t).Status != "CANCELLED" {
					t.Fatal("expiry overwrote cancellation")
				}
			}
		})
	}
	t.Run("TimeExpiryRemainsExpired", func(t *testing.T) {
		f := newFixture(t, db, 10001)
		mustExec(t, db, `UPDATE orders SET status='NOT_PAID',deadline_at=NOW()-INTERVAL '1 day' WHERE id=?`, f.orderID)
		mustExec(t, db, `UPDATE payments SET status='PENDING',qr_payload='test-qr',qr_generating_at=NOW() WHERE id=?`, f.paymentID)
		if _, err := f.orders.CancelExpired(ctx, time.Now().Add(-72*time.Hour), time.Now()); err != nil {
			t.Fatal(err)
		}
		p := f.summary(t)
		if p.Status != "EXPIRED" || p.SettledAt != nil {
			t.Fatalf("time expiry wrongly classified %+v", p)
		}
		assertWallet(t, f, f.customer, 0, 0)
		assertWallet(t, f, f.artist, 0, 0)
	})
	t.Run("UnpaidCancellationRollsBackOnPaymentFailure", func(t *testing.T) {
		f := newFixture(t, db, 10001)
		mustExec(t, db, `UPDATE orders SET status='NOT_PAID' WHERE id=?`, f.orderID)
		mustExec(t, db, `UPDATE payments SET status='PENDING',qr_payload='test-qr' WHERE id=?`, f.paymentID)
		mustExec(t, db, `ALTER TABLE payments ADD CONSTRAINT test_reject_cancelled_payment CHECK(status <> 'CANCELLED') NOT VALID`)
		t.Cleanup(func() { mustExec(t, db, `ALTER TABLE payments DROP CONSTRAINT test_reject_cancelled_payment`) })
		if err := f.cancel(ctx, order.ParticipantArtist); err == nil {
			t.Fatal("expected payment update failure")
		}
		var status string
		if err := db.NewRaw(`SELECT status FROM orders WHERE id=?`, f.orderID).Scan(ctx, &status); err != nil || status != "NOT_PAID" {
			t.Fatalf("order did not roll back %s %v", status, err)
		}
		var p struct {
			Status    string
			QRPayload *string
		}
		if err := db.NewRaw(`SELECT status,qr_payload FROM payments WHERE id=?`, f.paymentID).Scan(ctx, &p); err != nil || p.Status != "PENDING" || p.QRPayload == nil {
			t.Fatalf("payment did not roll back %+v %v", p, err)
		}
		assertWallet(t, f, f.customer, 0, 0)
		assertWallet(t, f, f.artist, 0, 0)
	})
	t.Run("CancelledStatusRollbackPreservesData", func(t *testing.T) {
		// Target the payment migration even when newer migrations exist.
		if err := goose.DownToContext(ctx, db.DB, ".", 20261105000000); err != nil {
			t.Fatal(err)
		}
		if err := goose.DownContext(ctx, db.DB, "."); err == nil {
			t.Fatal("rollback silently removed CANCELLED status")
		}
		version, err := goose.GetDBVersionContext(ctx, db.DB)
		if err != nil || version != 20261105000000 {
			t.Fatalf("migration version %d err %v", version, err)
		}
	})
}
