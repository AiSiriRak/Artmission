//go:build integration

package notifications

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/notification"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/baserepo"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/objectstorage"
	"github.com/AiSiriRak/Artmission/backend/internal/wiring"
	"github.com/AiSiriRak/Artmission/backend/tests/internal/apptest"
	"github.com/cucumber/godog"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

var app *apptest.App
var storage *objectstorage.Client

func TestNotificationsFeatures(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	pg := apptest.StartPostgres(ctx, t)
	s3Config := apptest.StartObjectStorage(ctx, t)
	app = apptest.NewAppWithS3(t, pg.DSN, s3Config, notificationPaymentProvider{})
	var err error
	storage, err = objectstorage.NewS3Client(ctx, s3Config)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll("../../reports/json", 0755); err != nil {
		t.Fatal(err)
	}
	suite := godog.TestSuite{ScenarioInitializer: InitializeScenario, Options: &godog.Options{
		Format: "pretty,cucumber:../../reports/json/cucumber-notifications.json", Paths: []string{"features"}, TestingT: t,
	}}
	if suite.Run() != 0 {
		t.Fatal("notification feature scenarios failed")
	}
	t.Run("WalletCredits", testWalletCredits)
	t.Run("WiredWorkers", testWiredWorkers)
	t.Run("DelayRequestConcurrency", testDelayRequestConcurrency)
	t.Run("UnreadLatestTwenty", testUnreadLatestTwenty)
}

func testUnreadLatestTwenty(t *testing.T) {
	ctx := context.Background()
	n := &notificationsContext{}
	mustNotificationStep(t, n.registerAccounts())
	owner, other := uuid.MustParse(n.accounts["artist"].ID), uuid.MustParse(n.accounts["customer"].ID)
	ids := make([]uuid.UUID, 24)
	prefix := uuid.New()
	created := time.Now().Add(-time.Hour)
	for i := range ids {
		ids[i] = prefix
		ids[i][15] = byte(i)
		// Equal timestamps exercise the same ID tie-breaker as the list endpoint.
		mustNotificationStep(t, n.exec("INSERT INTO notifications (id, user_id, title, message, type, is_read, created_at) VALUES (?, ?, 'Test', 'Unread count', 'GENERAL', ?, ?)", ids[i], owner, i >= 11, created))
	}
	_, err := notificationActions().Create(ctx, notification.CreateInput{UserID: other, Title: "Other account", Message: "Must not affect the artist", Type: notification.TypeGeneral})
	mustNotificationStep(t, err)
	checkCount := func(want int) {
		t.Helper()
		mustNotificationStep(t, n.request("artist", "GET", "/notifications/unread-count", nil, 200))
		var result struct {
			Count int `json:"count"`
		}
		mustNotificationStep(t, n.resp.JSON(&result))
		if result.Count != want {
			t.Fatalf("unread count %d want %d", result.Count, want)
		}
	}
	// Four old unread cards must not count; seven of the latest twenty are unread.
	checkCount(7)
	page, err := n.feed("artist", "?limit=20&offset=0")
	mustNotificationStep(t, err)
	if len(page.Items) != 20 || page.Total != 24 {
		t.Fatalf("list pagination changed: %+v", page)
	}
	unread := 0
	for _, card := range page.Items {
		if !card.IsRead {
			unread++
		}
	}
	if unread != 7 {
		t.Fatalf("feed and count disagree: %d", unread)
	}
	mustNotificationStep(t, notificationActions().MarkRead(ctx, ids[10], owner))
	checkCount(6)
	mustNotificationStep(t, notificationActions().MarkRead(ctx, ids[0], owner))
	checkCount(6)
	mustNotificationStep(t, n.exec("UPDATE notifications SET is_read = true WHERE id IN (?)", bun.List(ids[4:])))
	checkCount(0)
	_, err = notificationActions().Create(ctx, notification.CreateInput{UserID: owner, Title: "Newest", Message: "Enters the latest twenty", Type: notification.TypeGeneral})
	mustNotificationStep(t, err)
	checkCount(1)
}

func testWiredWorkers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	orderContext, paymentContext := &notificationsContext{}, &notificationsContext{}
	for _, n := range []*notificationsContext{orderContext, paymentContext} {
		mustNotificationStep(t, n.registerAccounts())
	}
	mustNotificationStep(t, orderContext.orderWithStatus("PENDING"))
	mustNotificationStep(t, orderContext.exec("UPDATE orders SET created_at = ? WHERE id = ?", time.Now().Add(-order.PendingConfirmationTTL-time.Hour), orderContext.orderID))
	mustNotificationStep(t, paymentContext.orderWithStatus("NOT_PAID"))
	mustNotificationStep(t, paymentContext.exec("UPDATE payments SET qr_expire = ? WHERE order_id = ?", time.Now().Add(-time.Hour), paymentContext.orderID))
	workers := wiring.WireWorkers(wiring.Config{DB: app.DB, ObjectStorage: storage,
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), PaymentProvider: notificationPaymentProvider{}})
	done := make(chan struct{}, len(workers))
	for _, w := range workers {
		go func() { w.Run(ctx); done <- struct{}{} }()
	}
	t.Cleanup(func() {
		cancel()
		for range workers {
			<-done
		}
	})
	for {
		var count int
		mustNotificationStep(t, app.DB.NewRaw("SELECT COUNT(*) FROM notifications WHERE order_id IN (?, ?) AND type = 'ORDER_CANCELED'", orderContext.orderID, paymentContext.orderID).Scan(context.Background(), &count))
		if count == 4 {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatalf("wired workers did not notify both recipients: got %d cards", count)
		case <-time.After(25 * time.Millisecond):
		}
	}
	for _, n := range []*notificationsContext{orderContext, paymentContext} {
		mustNotificationStep(t, n.receives("artist", "ORDER_CANCELED"))
		mustNotificationStep(t, n.receives("customer", "ORDER_CANCELED"))
	}
}

func testWalletCredits(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		name, initial, action string
		refund, payout        int64
	}{
		{"unpaid pending cancellation", "PENDING", "artist", 0, 0},
		{"unpaid accepted cancellation", "NOT_PAID", "customer", 0, 0},
		{"artist cancellation", "IN_PROCESS", "artist", 10000, 0},
		{"customer cancellation", "IN_PROCESS", "customer", 5000, 5000},
		{"customer acceptance", "IN_PROCESS", "accept", 0, 10000},
		{"final draft auto approval", "IN_PROCESS", "final", 0, 10000},
		{"review window expiry", "IN_PROCESS", "auto", 0, 10000},
		{"deadline expiry", "IN_PROCESS", "deadline", 10000, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			n := &notificationsContext{}
			mustNotificationStep(t, n.registerAccounts())
			mustNotificationStep(t, n.orderWithStatus(tc.initial))
			switch tc.action {
			case "artist", "customer":
				mustNotificationStep(t, n.request(tc.action, "PUT", "/orders/"+n.orderID+"/cancel", map[string]string{"reason": "OTHER", "explanation": "Test wallet notification"}, 200))
			case "accept":
				mustNotificationStep(t, n.submitDraft())
				mustNotificationStep(t, n.confirmDraft("accepts"))
			case "final":
				for i := 0; i < 2; i++ {
					mustNotificationStep(t, n.submitDraft())
					mustNotificationStep(t, n.confirmDraft("rejects"))
				}
				mustNotificationStep(t, n.submitDraft())
			case "auto":
				mustNotificationStep(t, n.submitDraft())
				mustNotificationStep(t, n.exec("UPDATE order_deliverables SET created_at = ? WHERE id = ?", time.Now().Add(-15*24*time.Hour), n.draftID))
				mustNotificationStep(t, n.runExpiry())
				mustNotificationStep(t, n.runExpiry())
			case "deadline":
				mustNotificationStep(t, n.exec("UPDATE orders SET deadline_at = ? WHERE id = ?", time.Now().Add(-time.Hour), n.orderID))
				mustNotificationStep(t, n.runExpiry())
				mustNotificationStep(t, n.runExpiry())
			}
			for _, credit := range []struct {
				kind, role string
				amount     int64
			}{{"REFUND_CREDITED", "customer", tc.refund}, {"WALLET_CREDITED", "artist", tc.payout}} {
				var cards []struct {
					ID      uuid.UUID
					Message string
					IsRead  bool
				}
				mustNotificationStep(t, app.DB.NewRaw("SELECT id, message, is_read FROM notifications WHERE order_id = ? AND user_id = ? AND type = ?", n.orderID, n.accounts[credit.role].ID, credit.kind).Scan(ctx, &cards))
				want := 0
				if credit.amount > 0 {
					want = 1
				}
				if len(cards) != want {
					t.Fatalf("%s: got %d cards, want %d", credit.kind, len(cards), want)
				}
				if want == 0 {
					continue
				}
				if !strings.Contains(cards[0].Message, fmt.Sprintf("%d.%02d THB", credit.amount/100, credit.amount%100)) {
					t.Fatalf("incorrect amount: %s", cards[0].Message)
				}
				var balance int64
				mustNotificationStep(t, app.DB.NewRaw("SELECT balance_satang FROM wallets WHERE user_id = ?", n.accounts[credit.role].ID).Scan(ctx, &balance))
				if balance != credit.amount {
					t.Fatalf("wallet balance %d want %d", balance, credit.amount)
				}
				mustNotificationStep(t, n.receives(credit.role, credit.kind))

			}
		})
	}
	t.Run("refund notification failure rolls back both credits", func(t *testing.T) {
		n := &notificationsContext{}
		mustNotificationStep(t, n.registerAccounts())
		mustNotificationStep(t, n.orderWithStatus("IN_PROCESS"))
		mustNotificationStep(t, n.rememberTotals())
		actions := orderActions(failRefundNotification{creator: notificationActions()})
		_, err := actions.CancelOrder(ctx, order.ParticipantCustomer, uuid.MustParse(n.accounts["customer"].ID), order.CancelInput{OrderID: uuid.MustParse(n.orderID), Reason: order.CancelReasonOther, Explanation: "Test rollback"})
		if err == nil {
			t.Fatal("expected refund notification failure")
		}
		var state struct{ OrderStatus, PaymentStatus string }
		mustNotificationStep(t, app.DB.NewRaw("SELECT o.status AS order_status, p.status AS payment_status FROM orders o JOIN payments p ON p.order_id = o.id WHERE o.id = ?", n.orderID).Scan(ctx, &state))
		if state.OrderStatus != "IN_PROCESS" || state.PaymentStatus != "HELD" {
			t.Fatalf("state not rolled back: %+v", state)
		}
		var count int
		mustNotificationStep(t, app.DB.NewRaw("SELECT COUNT(*) FROM payment_transactions pt JOIN payments p ON p.id = pt.payment_id WHERE p.order_id = ? AND pt.type <> 'CHARGE'", n.orderID).Scan(ctx, &count))
		if count != 0 {
			t.Fatal("settlement ledger did not roll back")
		}
		mustNotificationStep(t, app.DB.NewRaw("SELECT COUNT(*) FROM wallets WHERE user_id IN (?, ?) AND balance_satang <> 0", n.accounts["customer"].ID, n.accounts["artist"].ID).Scan(ctx, &count))
		if count != 0 {
			t.Fatal("wallet credits did not roll back")
		}
		mustNotificationStep(t, n.unchanged())
		// Retrying with working storage commits both credits and their cards.
		_, err = orderActions(notificationActions()).CancelOrder(ctx, order.ParticipantCustomer, uuid.MustParse(n.accounts["customer"].ID), order.CancelInput{OrderID: uuid.MustParse(n.orderID), Reason: order.CancelReasonOther, Explanation: "Retry"})
		mustNotificationStep(t, err)
		mustNotificationStep(t, n.receives("artist", "WALLET_CREDITED"))
		mustNotificationStep(t, n.receives("customer", "REFUND_CREDITED"))
		// The terminal order transition rejects a repeated cancellation.
		mustNotificationStep(t, n.rememberTotals())
		_, err = orderActions(notificationActions()).CancelOrder(ctx, order.ParticipantCustomer, uuid.MustParse(n.accounts["customer"].ID), order.CancelInput{OrderID: uuid.MustParse(n.orderID), Reason: order.CancelReasonOther, Explanation: "Repeat"})
		if err == nil {
			t.Fatal("repeated cancellation should fail")
		}
		mustNotificationStep(t, n.unchanged())

	})
}

func mustNotificationStep(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

type failRefundNotification struct {
	creator notification.NotificationCreator
}

func (f failRefundNotification) Create(ctx context.Context, in notification.CreateInput) (*notification.Notification, error) {
	if in.Type == notification.TypeRefundCredited {
		return nil, errors.New("refund notification unavailable")
	}
	return f.creator.Create(ctx, in)
}

func testDelayRequestConcurrency(t *testing.T) {
	for _, customerFirst := range []bool{true, false} {
		name := "expiry first"
		if customerFirst {
			name = "customer first"
		}
		t.Run(name, func(t *testing.T) {
			n := &notificationsContext{}
			mustNotificationStep(t, n.registerAccounts())
			mustNotificationStep(t, n.orderWithStatus("IN_PROCESS"))
			mustNotificationStep(t, n.requestExtension())
			mustNotificationStep(t, n.exec("UPDATE order_delay_requests SET created_at = ? WHERE id = ?", time.Now().Add(-order.DelayRequestTTL-time.Hour), n.requestID))
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			var deadlines struct{ Original, Requested time.Time }
			mustNotificationStep(t, app.DB.NewRaw("SELECT original_deadline_at AS original, deadline_request_at AS requested FROM order_delay_requests WHERE id = ?", n.requestID).Scan(ctx, &deadlines))
			tx, executor := baserepo.NewTransactioner(app.DB), baserepo.NewExecutor(app.DB)
			actions := orderActions(notificationActions())
			confirm := func(ctx context.Context) error {
				_, err := actions.ConfirmDelayRequest(ctx, uuid.MustParse(n.accounts["customer"].ID), uuid.MustParse(n.orderID), order.ConfirmDelayRequestInput{DelayRequestID: uuid.MustParse(n.requestID), Accept: true})
				return err
			}
			var expired []uuid.UUID
			expire := func(ctx context.Context) error {
				var err error
				expired, err = actions.CancelExpiredDelayRequest(ctx)
				return err
			}
			first, second := confirm, expire
			if !customerFirst {
				first, second = expire, confirm
			}
			secondPID := make(chan int, 1)
			secondDone := make(chan error, 1)
			firstErr := tx.Transaction(ctx, func(firstCtx context.Context) error {
				var firstPID int
				if err := executor.Run(firstCtx, func(db bun.IDB) error {
					// Hold the parent lock as ConfirmDelayRequest does, then force a real overlap.
					var id uuid.UUID
					if err := db.NewRaw("SELECT id FROM orders WHERE id = ? FOR UPDATE", n.orderID).Scan(firstCtx, &id); err != nil {
						return err
					}
					return db.NewRaw("SELECT pg_backend_pid()").Scan(firstCtx, &firstPID)
				}); err != nil {
					return err
				}
				go func() {
					err := tx.Transaction(ctx, func(secondCtx context.Context) error {
						var pid int
						if err := executor.Run(secondCtx, func(db bun.IDB) error { return db.NewRaw("SELECT pg_backend_pid()").Scan(secondCtx, &pid) }); err != nil {
							secondPID <- 0
							return err
						}
						secondPID <- pid
						return second(secondCtx)
					})
					secondDone <- err
				}()
				var waitingPID int
				select {
				case waitingPID = <-secondPID:
				case <-ctx.Done():
					return ctx.Err()
				}
				if waitingPID == 0 {
					return errors.New("second transaction did not start")
				}
				// Observe PostgreSQL's actual wait graph, rather than relying on sleep ordering.
				for {
					var blocked bool
					if err := app.DB.NewRaw("SELECT ? = ANY(pg_blocking_pids(?))", firstPID, waitingPID).Scan(ctx, &blocked); err != nil {
						return err
					}
					if blocked {
						break
					}
					select {
					case <-ctx.Done():
						return fmt.Errorf("transactions did not overlap: %w", ctx.Err())
					case <-time.After(10 * time.Millisecond):
					}
				}
				return first(firstCtx)
			})
			var secondErr error
			select {
			case secondErr = <-secondDone:
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
			confirmErr, expiryErr := firstErr, secondErr
			if !customerFirst {
				confirmErr, expiryErr = secondErr, firstErr
			}
			if expiryErr != nil {
				t.Fatalf("expiry failed during concurrent confirmation: %v", expiryErr)
			}
			if customerFirst && confirmErr != nil {
				t.Fatalf("confirmation failed while holding the parent order: %v", confirmErr)
			}
			if !customerFirst && !errors.Is(confirmErr, order.ErrInvalidrDelayRequestStatus) {
				t.Fatalf("expected already-decided request, got %v", confirmErr)
			}
			var state struct {
				Decision string
				Deadline time.Time
			}
			mustNotificationStep(t, app.DB.NewRaw("SELECT r.customer_decision AS decision, o.deadline_at AS deadline FROM order_delay_requests r JOIN orders o ON o.id = r.order_id WHERE r.id = ?", n.requestID).Scan(ctx, &state))
			wantDecision, wantDeadline, approved, rejected := "REJECTED", deadlines.Original, 0, 2
			if customerFirst {
				wantDecision, wantDeadline, approved, rejected = "APPROVED", deadlines.Requested, 1, 0
			}
			if state.Decision != wantDecision || !state.Deadline.Equal(wantDeadline) {
				t.Fatalf("inconsistent state: %+v", state)
			}
			wantExpired := 1
			if customerFirst {
				wantExpired = 0
			}
			if len(expired) != wantExpired || (len(expired) == 1 && expired[0] != uuid.MustParse(n.requestID)) {
				t.Fatalf("incorrect expiry result: %v", expired)
			}
			for kind, want := range map[string]int{"DEADLINE_EXTENSION_APPROVED": approved, "DEADLINE_EXTENSION_REJECTED": rejected} {
				var count int
				mustNotificationStep(t, app.DB.NewRaw("SELECT COUNT(*) FROM notifications WHERE delay_request_id = ? AND type = ?", n.requestID, kind).Scan(ctx, &count))
				if count != want {
					t.Fatalf("%s count %d want %d", kind, count, want)
				}
			}
		})
	}
}
