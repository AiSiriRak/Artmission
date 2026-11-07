//go:build integration

package payments

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/adapters/postgres"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/payment"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/config"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/database"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/objectstorage"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/seed"
	"github.com/AiSiriRak/Artmission/backend/tests/internal/apptest"
	"github.com/google/uuid"
)

func TestQRGenerationClaimDoesNotHoldPoolConnection(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	pg := apptest.StartPostgres(ctx, t)
	s3Config := apptest.StartObjectStorage(ctx, t)
	db, err := database.NewPostgresDB(config.Database{DSN: pg.DSN})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	storage, err := objectstorage.NewS3Client(ctx, s3Config)
	if err != nil {
		t.Fatal(err)
	}
	if err := seed.Up(ctx, db, storage); err != nil {
		t.Fatalf("seed fixtures: %v", err)
	}

	var orderID uuid.UUID
	if err := db.NewSelect().Table("orders").Column("id").
		Where("name = ?", "Anime avatar").Where("status = ?", payment.EligibleOrderStatus).
		Scan(ctx, &orderID); err != nil {
		t.Fatalf("load unpaid order: %v", err)
	}
	repo := postgres.NewPaymentRepository(db)
	item, err := repo.GetQRForOrder(ctx, orderID)
	if err != nil {
		t.Fatalf("load payment: %v", err)
	}

	first, acquired, err := repo.TryLockQRGeneration(ctx, item.ID)
	if err != nil || !acquired {
		t.Fatalf("first claim = %v, %v; want acquired", acquired, err)
	}
	// A claim is just a row update, so other queries still work while it is held.
	if _, err := repo.GetQRForOrder(ctx, orderID); err != nil {
		t.Fatalf("query while claim is held: %v", err)
	}
	if _, acquired, err := repo.TryLockQRGeneration(ctx, item.ID); err != nil || acquired {
		t.Fatalf("second claim = %v, %v; want not acquired", acquired, err)
	}
	if err := first.Release(); err != nil {
		t.Fatalf("release: %v", err)
	}
	again, acquired, err := repo.TryLockQRGeneration(ctx, item.ID)
	if err != nil || !acquired {
		t.Fatalf("claim after release = %v, %v; want acquired", acquired, err)
	}

	// SaveQRPayload clears the claim.
	if saved, err := repo.SaveQRPayload(ctx, item.ID, "claimed-payload"); err != nil || !saved {
		t.Fatalf("SaveQRPayload() = %v, %v", saved, err)
	}
	var generatingAt *time.Time
	if err := db.NewSelect().Table("payments").Column("qr_generating_at").
		Where("id = ?", item.ID).Scan(ctx, &generatingAt); err != nil {
		t.Fatal(err)
	}
	if generatingAt != nil {
		t.Fatalf("qr_generating_at = %v after save, want NULL", generatingAt)
	}
	_ = again.Release()

	// A crashed holder is superseded after the TTL.
	if _, err := db.NewUpdate().Table("payments").
		Set("qr_payload = NULL").
		Set("qr_generating_at = clock_timestamp() - INTERVAL '1 hour'").
		Where("id = ?", item.ID).Exec(ctx); err != nil {
		t.Fatal(err)
	}
	if _, acquired, err := repo.TryLockQRGeneration(ctx, item.ID); err != nil || !acquired {
		t.Fatalf("claim over stale holder = %v, %v; want acquired", acquired, err)
	}
}

func TestPaymentRepositoryLifecycle(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	pg := apptest.StartPostgres(ctx, t)
	s3Config := apptest.StartObjectStorage(ctx, t)
	db, err := database.NewPostgresDB(config.Database{DSN: pg.DSN})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	storage, err := objectstorage.NewS3Client(ctx, s3Config)
	if err != nil {
		t.Fatal(err)
	}
	if err := seed.Up(ctx, db, storage); err != nil {
		t.Fatalf("seed fixtures: %v", err)
	}

	var pendingOrder struct {
		ID       uuid.UUID `bun:"id"`
		ArtistID uuid.UUID `bun:"artist_id"`
		Price    int64     `bun:"price_satang_order"`
	}
	if err := db.NewSelect().
		Table("orders").
		Column("id", "artist_id", "price_satang_order").
		Where("name = ?", "Portrait for mom's birthday").
		Where("status = ?", "PENDING").
		Scan(ctx, &pendingOrder); err != nil {
		t.Fatalf("load pending order: %v", err)
	}

	orderRepo := postgres.NewOrderRepository(db)
	if err := orderRepo.ConfirmOrder(ctx, pendingOrder.ArtistID, pendingOrder.ID, order.StatusNotPaid); err != nil {
		t.Fatalf("transition order to NOT_PAID: %v", err)
	}
	if err := orderRepo.ConfirmOrder(ctx, pendingOrder.ArtistID, pendingOrder.ID, order.StatusNotPaid); !errors.Is(err, order.ErrInvalidOrderStatus) {
		t.Fatalf("duplicate NOT_PAID transition error = %v, want invalid status", err)
	}
	var paymentCount int
	if err := db.NewSelect().
		Table("payments").
		ColumnExpr("COUNT(*)").
		Where("order_id = ?", pendingOrder.ID).
		Scan(ctx, &paymentCount); err != nil {
		t.Fatalf("count payments after duplicate transition: %v", err)
	}
	if paymentCount != 1 {
		t.Fatalf("duplicate transition created %d payment rows, want 1", paymentCount)
	}

	var transitionAt time.Time
	if err := db.NewSelect().
		Table("orders").
		Column("updated_at").
		Where("id = ?", pendingOrder.ID).
		Scan(ctx, &transitionAt); err != nil {
		t.Fatalf("load transition timestamp: %v", err)
	}
	paymentRepo := postgres.NewPaymentRepository(db)
	createdPayment, err := paymentRepo.GetQRForOrder(ctx, pendingOrder.ID)
	if err != nil {
		t.Fatalf("load payment created with order transition: %v", err)
	}
	if createdPayment.AmountSatang != pendingOrder.Price ||
		createdPayment.QRExpire == nil ||
		!createdPayment.QRExpire.Equal(transitionAt.Add(payment.QRValidityPeriod)) ||
		!createdPayment.CreatedAt.Equal(transitionAt) {
		t.Fatalf("payment = %+v, not based on the committed NOT_PAID transition", createdPayment)
	}

	saved, err := paymentRepo.SaveQRPayload(ctx, createdPayment.ID, "saved-payload")
	if err != nil || !saved {
		t.Fatalf("SaveQRPayload() = %v, %v; want true, nil", saved, err)
	}
	saved, err = paymentRepo.SaveQRPayload(ctx, createdPayment.ID, "second-payload")
	if err != nil || saved {
		t.Fatalf("second SaveQRPayload() = %v, %v; want false, nil", saved, err)
	}

	var unpaidOrder struct {
		ID         uuid.UUID `bun:"id"`
		CustomerID uuid.UUID `bun:"customer_id"`
	}
	if err := db.NewSelect().
		Table("orders").
		Column("id", "customer_id").
		Where("name = ?", "Anime avatar").
		Where("status = ?", payment.EligibleOrderStatus).
		Scan(ctx, &unpaidOrder); err != nil {
		t.Fatalf("load seeded unpaid order: %v", err)
	}
	unpaidPayment, err := paymentRepo.GetQRForOrder(ctx, unpaidOrder.ID)
	if err != nil {
		t.Fatalf("load seeded payment: %v", err)
	}
	if _, err := db.NewUpdate().
		Table("payments").
		Set("qr_payload = ?", "integration-test-payload").
		Set("qr_expire = clock_timestamp() + INTERVAL '20 milliseconds'").
		Where("id = ?", unpaidPayment.ID).
		Exec(ctx); err != nil {
		t.Fatalf("prepare expiring payment: %v", err)
	}

	start := make(chan struct{})
	expiryDone := make(chan error, 1)
	verificationDone := make(chan error, 1)
	go func() {
		<-start
		_, _, _, err := paymentRepo.ExpirePayment(ctx, unpaidPayment.ID)
		expiryDone <- err
	}()
	go func() {
		<-start
		_, err := paymentRepo.ConfirmSlipPayment(ctx, unpaidOrder.CustomerID, unpaidOrder.ID, "integration-bank-ref")
		verificationDone <- err
	}()
	close(start)

	expiryErr := <-expiryDone
	verifyErr := <-verificationDone
	if expiryErr != nil {
		t.Fatalf("concurrent expiry: %v", expiryErr)
	}
	if verifyErr != nil && !errors.Is(verifyErr, payment.ErrQRExpired) &&
		!errors.Is(verifyErr, payment.ErrOrderNotPayable) {
		t.Fatalf("concurrent verification returned unexpected error: %v", verifyErr)
	}

	time.Sleep(25 * time.Millisecond)
	if _, _, _, err := paymentRepo.ExpirePayment(ctx, unpaidPayment.ID); err != nil {
		t.Fatalf("repeat expiry: %v", err)
	}

	var finalOrderStatus string
	if err := db.NewSelect().
		Table("orders").
		Column("status").
		Where("id = ?", unpaidOrder.ID).
		Scan(ctx, &finalOrderStatus); err != nil {
		t.Fatalf("load final order status: %v", err)
	}
	var finalPayment struct {
		Status    string  `bun:"status"`
		QRPayload *string `bun:"qr_payload"`
	}
	if err := db.NewSelect().
		Table("payments").
		Column("status", "qr_payload").
		Where("order_id = ?", unpaidOrder.ID).
		Scan(ctx, &finalPayment); err != nil {
		t.Fatalf("load final payment: %v", err)
	}

	if verifyErr == nil {
		if finalOrderStatus != "IN_PROCESS" || finalPayment.Status != string(payment.StatusHeld) || finalPayment.QRPayload != nil {
			t.Fatalf("verified payment was invalidated: order=%s payment=%+v", finalOrderStatus, finalPayment)
		}
		return
	}
	if finalOrderStatus != "CANCEL" || finalPayment.Status != string(payment.StatusExpired) || finalPayment.QRPayload != nil {
		t.Fatalf("expired payment did not cancel and clear QR: order=%s payment=%+v", finalOrderStatus, finalPayment)
	}
}
