//go:build integration

package notifications

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/adapters/postgres"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/notification"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/baserepo"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/config"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/database"
	"github.com/AiSiriRak/Artmission/backend/tests/internal/apptest"
	"github.com/google/uuid"
)

type failingNotifications struct{}

func (failingNotifications) Create(context.Context, notification.CreateInput) (*notification.Notification, error) {
	return nil, errors.New("notification storage unavailable")
}

func TestPendingAndNotPaidNotifications(t *testing.T) {
	ctx := context.Background()
	pg := apptest.StartPostgres(ctx, t)
	db, err := database.NewPostgresDB(config.Database{DSN: pg.DSN})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := db.ExecContext(ctx, query, args...); err != nil {
			t.Fatal(err)
		}
	}
	artist, customer := uuid.New(), uuid.New()
	exec("INSERT INTO users (id,username,email,password_hash,role) VALUES (?, 'Artist','artist@test.local','hash','artist'), (?, 'Customer','customer@test.local','hash','customer')", artist, customer)
	exec("INSERT INTO artist_profiles (user_id) VALUES (?)", artist)
	raw := postgres.NewOrderRepository(db)
	tx := baserepo.NewTransactioner(db)
	uc := notification.NewNotificationUsecase(postgres.NewNotificationRepository(db))
	repo := notification.NewNotifyingRepository(raw, uc, tx)
	now := time.Now()
	item := &order.Order{ID: uuid.New(), ArtistID: artist, CustomerID: customer, Name: "Portrait",
		PriceSatangOrder: 10000, Status: order.StatusPending, DeadlineAt: now.Add(7 * 24 * time.Hour), CreatedAt: now, UpdatedAt: now}
	list := func(owner uuid.UUID) notification.Page {
		t.Helper()
		page, err := uc.ListMine(ctx, owner, notification.ListQuery{Limit: 100})
		if err != nil {
			t.Fatal(err)
		}
		return page
	}
	if err := repo.Create(ctx, item); err != nil {
		t.Fatal(err)
	}
	artistPage := list(artist)
	if artistPage.Total != 1 {
		t.Fatalf("expected new-order notification, got %+v", artistPage)
	}
	n := artistPage.Items[0]
	if n.EventType != notification.EventOrderCreated || n.OrderID == nil || *n.OrderID != item.ID ||
		n.OrderStatus == nil || *n.OrderStatus != order.StatusPending || n.DeadlineAt == nil ||
		n.DeadlineAt.Sub(item.DeadlineAt).Abs() > time.Microsecond {
		t.Fatalf("pending card: %+v", n)
	}
	if strings.Contains(n.Message, "Order:") || strings.Contains(n.Message, "Status:") {
		t.Fatal("technical metadata in display message")
	}
	if list(customer).Total != 0 {
		t.Fatal("new-order notification sent to customer")
	}

	if err := repo.ConfirmOrder(ctx, artist, item.ID, order.StatusNotPaid); err != nil {
		t.Fatal(err)
	}
	customerPage := list(customer)
	if customerPage.Total != 1 || customerPage.Items[0].EventType != notification.EventOrderAccepted ||
		customerPage.Items[0].OrderStatus == nil || *customerPage.Items[0].OrderStatus != order.StatusNotPaid {
		t.Fatalf("accepted card: %+v", customerPage)
	}
	if err := repo.ConfirmOrder(ctx, artist, item.ID, order.StatusNotPaid); err == nil {
		t.Fatal("duplicate acceptance should fail")
	}
	if list(customer).Total != 1 {
		t.Fatal("failed action produced duplicate notification")
	}
	if err := uc.MarkRead(ctx, n.ID, customer); err != notification.ErrNotFound {
		t.Fatalf("cross-account read: %v", err)
	}
	for i := 0; i < 2; i++ {
		if err := uc.MarkRead(ctx, n.ID, artist); err != nil {
			t.Fatal(err)
		}
	}
	if count, err := uc.CountUnread(ctx, artist); err != nil || count != 0 {
		t.Fatalf("unread = %d: %v", count, err)
	}
	empty, err := uc.ListMine(ctx, customer, notification.ListQuery{Limit: 20, Offset: 1000})
	if err != nil || empty.Items == nil || len(empty.Items) != 0 || empty.Total != 1 {
		t.Fatalf("empty page: %+v %v", empty, err)
	}

	t.Run("later actions keep working without notifications", func(t *testing.T) {
		beforeArtist, beforeCustomer := list(artist).Total, list(customer).Total
		if err := repo.BypassOrderPayment(ctx, customer, item.ID); err != nil {
			t.Fatal(err)
		}
		_, err := repo.CreateDeliverable(ctx, artist, item.ID, &order.Deliverable{
			ID: uuid.New(), OriginalImageKey: "original.png", PreviewImageKey: "preview.png", CreatedAt: now, UpdatedAt: now,
		})
		if err != nil {
			t.Fatal(err)
		}
		if list(artist).Total != beforeArtist || list(customer).Total != beforeCustomer {
			t.Fatal("in-process actions must not generate notifications")
		}
	})

	t.Run("rejection does not notify", func(t *testing.T) {
		rejected := *item
		rejected.ID = uuid.New()
		if err := raw.Create(ctx, &rejected); err != nil {
			t.Fatal(err)
		}
		before := list(customer).Total
		if err := repo.ConfirmOrder(ctx, artist, rejected.ID, order.StatusCancel); err != nil {
			t.Fatal(err)
		}
		if list(customer).Total != before {
			t.Fatal("only accepted orders should notify")
		}
	})

	t.Run("failed notification rolls back acceptance", func(t *testing.T) {
		pending := *item
		pending.ID = uuid.New()
		if err := raw.Create(ctx, &pending); err != nil {
			t.Fatal(err)
		}
		failing := notification.NewNotifyingRepository(raw, failingNotifications{}, tx)
		if err := failing.ConfirmOrder(ctx, artist, pending.ID, order.StatusNotPaid); err == nil {
			t.Fatal("expected failure")
		}
		detail, err := raw.GetOrderByID(ctx, order.ParticipantArtist, artist, pending.ID)
		if err != nil || detail.Order.Status != order.StatusPending {
			t.Fatalf("acceptance was not rolled back: %+v %v", detail, err)
		}
	})

	t.Run("ledger does not create notifications in this increment", func(t *testing.T) {
		beforeArtist, beforeCustomer := list(artist).Total, list(customer).Total
		paymentID := uuid.New()
		exec("INSERT INTO payments (id,order_id,amount_satang,status) VALUES (?,?,10000,'HELD')", paymentID, item.ID)
		exec("INSERT INTO payment_transactions (id,payment_id,type,amount_satang) VALUES (?,?,'CHARGE',10000)", uuid.New(), paymentID)
		if list(artist).Total != beforeArtist || list(customer).Total != beforeCustomer {
			t.Fatal("payment notification hook is still active")
		}
	})
}
