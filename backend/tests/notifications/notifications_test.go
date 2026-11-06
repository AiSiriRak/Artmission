//go:build integration

package notifications

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/notification"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/objectstorage"
	"github.com/AiSiriRak/Artmission/backend/tests/internal/apptest"
	"github.com/cucumber/godog"
	"github.com/google/uuid"
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
