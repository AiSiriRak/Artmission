//go:build integration

package payments

import (
	"context"
	"math"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/adapters/postgres"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/payment"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/config"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/database"
	appmigrations "github.com/AiSiriRak/Artmission/backend/internal/pkg/migrations"
	"github.com/AiSiriRak/Artmission/backend/tests/internal/apptest"
	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
	"github.com/uptrace/bun"
)

type fixture struct {
	customer, artist, orderID, paymentID uuid.UUID
	db                                   *bun.DB
	wallet                               payment.WalletRepository
	orders                               order.OrderRepository
}

func mustExec(t *testing.T, db *bun.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.NewRaw(query, args...).Exec(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func newFixture(t *testing.T, db *bun.DB, amount int64) fixture {
	t.Helper()
	f := fixture{customer: uuid.New(), artist: uuid.New(), orderID: uuid.New(), paymentID: uuid.New(), db: db,
		wallet: postgres.NewWalletRepository(db), orders: postgres.NewOrderRepository(db)}
	for _, u := range []struct {
		id   uuid.UUID
		role string
	}{{f.customer, "customer"}, {f.artist, "artist"}} {
		mustExec(t, db, `INSERT INTO users(id,username,email,password_hash,role) VALUES(?,'wallet test',?,'test',?)`, u.id, u.id.String()+"@test.local", u.role)
	}
	mustExec(t, db, `INSERT INTO artist_profiles(user_id) VALUES(?)`, f.artist)
	mustExec(t, db, `INSERT INTO orders(id,customer_id,artist_id,name,customer_description,price_satang_order,status,deadline_at)
		VALUES(?,?,?,'wallet order','',?,'IN_PROCESS',NOW()+INTERVAL '7 days')`, f.orderID, f.customer, f.artist, amount)
	mustExec(t, db, `INSERT INTO payments(id,order_id,amount_satang,status) VALUES(?,?,?,'HELD')`, f.paymentID, f.orderID, amount)
	return f
}

func (f fixture) cancel(ctx context.Context, participant order.Participant) error {
	id := f.customer
	if participant == order.ParticipantArtist {
		id = f.artist
	}
	return f.orders.CancelOrder(ctx, participant, id, &order.OrderCancellation{ID: uuid.New(), OrderID: f.orderID, Reason: "OTHER", CreatedAt: time.Now()})
}

func (f fixture) summary(t *testing.T) *payment.Summary {
	t.Helper()
	page, err := f.orders.ListOrders(context.Background(), order.ListQuery{Participant: order.ParticipantCustomer, ParticipantID: f.customer, Sort: order.DefaultSort, Order: order.DefaultOrder, Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	for _, o := range page.Orders {
		if o.ID == f.orderID {
			return o.Payment
		}
	}
	t.Fatal("order missing")
	return nil
}

func assertWallet(t *testing.T, f fixture, id uuid.UUID, amount int64, count int) {
	t.Helper()
	ctx := context.Background()
	wallet, err := f.wallet.GetWallet(ctx, id)
	if err != nil || wallet.UserID != id || wallet.BalanceSatang != amount {
		t.Fatalf("wallet %+v want %d err %v", wallet, amount, err)
	}
	page, err := f.wallet.ListWalletTransactions(ctx, id, payment.TransactionQuery{Limit: 100})
	if err != nil || page.Total != count || len(page.Transactions) != count {
		t.Fatalf("history %+v want %d err %v", page, count, err)
	}
	if count == 0 && wallet.UpdatedAt != nil {
		t.Fatal("empty wallet invented a credit timestamp")
	}
	if count > 0 && wallet.UpdatedAt == nil {
		t.Fatal("credited wallet lacks timestamp")
	}
	var sum int64
	for _, item := range page.Transactions {
		sum += item.AmountSatang
		if item.AmountSatang <= 0 || item.OrderName != "wallet order" || item.CreatedAt.IsZero() {
			t.Fatalf("invalid credit %+v", item)
		}
		var recipient uuid.UUID
		if err := f.db.NewRaw(`SELECT recipient_user_id FROM payment_transactions WHERE id=?`, item.ID).Scan(ctx, &recipient); err != nil || recipient != id {
			t.Fatalf("recipient %s want %s err %v", recipient, id, err)
		}
	}
	if sum != amount {
		t.Fatalf("history sum %d differs from wallet %d", sum, amount)
	}
}

func TestAutomaticWalletSettlement(t *testing.T) {
	ctx := context.Background()
	pg := apptest.StartPostgres(ctx, t)
	db, err := database.NewPostgresDB(config.Database{DSN: pg.DSN})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	for _, outcome := range []payment.SettlementOutcome{payment.Completed, payment.ArtistCancelled, payment.CustomerCancelled, payment.DeadlineExpired} {
		t.Run(string(outcome), func(t *testing.T) {
			f := newFixture(t, db, 10001)
			switch outcome {
			case payment.Completed:
				id := uuid.New()
				mustExec(t, db, `INSERT INTO order_deliverables(id,order_id,version,original_image_key,preview_image_key) VALUES(?,?,1,'original','preview')`, id, f.orderID)
				if err := f.orders.ConfirmDeliverable(ctx, f.customer, f.orderID, id, order.DeliverableDecisionApproved, ""); err != nil {
					t.Fatal(err)
				}
				if err := f.orders.ConfirmDeliverable(ctx, f.customer, f.orderID, id, order.DeliverableDecisionApproved, ""); err == nil {
					t.Fatal("completed twice")
				}
			case payment.DeadlineExpired:
				mustExec(t, db, `UPDATE orders SET deadline_at=NOW()-INTERVAL '1 day' WHERE id=?`, f.orderID)
				for range 2 {
					if _, err := f.orders.CancelExpired(ctx, time.Now().Add(-24*time.Hour), time.Now()); err != nil {
						t.Fatal(err)
					}
				}
			default:
				participant := order.ParticipantCustomer
				if outcome == payment.ArtistCancelled {
					participant = order.ParticipantArtist
				}
				if err := f.cancel(ctx, participant); err != nil {
					t.Fatal(err)
				}
				if err := f.cancel(ctx, participant); err == nil {
					t.Fatal("cancelled twice")
				}
			}
			p := f.summary(t)
			refund, artist, _ := payment.CalculateSettlement(outcome, 10001)
			if p.Status != "SETTLED" || p.SettlementOutcome == nil || *p.SettlementOutcome != outcome || p.SettledAt == nil || p.RefundSatang != refund || p.ArtistPayoutSatang != artist {
				t.Fatalf("summary %+v", p)
			}
			for _, a := range []struct {
				id     uuid.UUID
				amount int64
				status *payment.TransferStatus
			}{{f.customer, refund, p.RefundTransferStatus}, {f.artist, artist, p.ArtistPayoutTransferStatus}} {
				count := 0
				if a.amount > 0 {
					count = 1
					if a.status == nil || *a.status != payment.TransferSuccess {
						t.Fatal("wallet credit not completed")
					}
				} else if a.status != nil {
					t.Fatal("zero transfer row")
				}
				assertWallet(t, f, a.id, a.amount, count)
			}
			var charges int
			if err := db.NewRaw(`SELECT COUNT(*) FROM payment_transactions WHERE payment_id=? AND type='CHARGE'`, f.paymentID).Scan(ctx, &charges); err != nil || charges != 0 {
				t.Fatal("fabricated historical charge")
			}
		})
	}

	t.Run("existing slip payment credits refund and excludes charge history", func(t *testing.T) {
		f := newFixture(t, db, 10001)
		mustExec(t, db, `UPDATE orders SET status='NOT_PAID' WHERE id=?`, f.orderID)
		mustExec(t, db, `UPDATE payments SET status='PENDING',qr_payload='test-qr',qr_expire=NOW()+INTERVAL '1 hour' WHERE id=?`, f.paymentID)
		if _, err := postgres.NewPaymentRepository(db).ConfirmSlipPayment(ctx, f.customer, f.orderID, uuid.NewString()); err != nil {
			t.Fatal(err)
		}
		if err := f.cancel(ctx, order.ParticipantArtist); err != nil {
			t.Fatal(err)
		}
		assertWallet(t, f, f.customer, 10001, 1)
		var count int
		if err := db.NewRaw(`SELECT COUNT(*) FROM payment_transactions WHERE payment_id=?`, f.paymentID).Scan(ctx, &count); err != nil || count != 2 {
			t.Fatalf("expected charge and refund: %d %v", count, err)
		}
	})
	t.Run("empty wallet and pending cancellation", func(t *testing.T) {
		f := newFixture(t, db, 10001)
		assertWallet(t, f, f.customer, 0, 0)
		mustExec(t, db, `UPDATE payments SET status='PENDING',qr_payload='pending-qr' WHERE id=?`, f.paymentID)
		if err := f.cancel(ctx, order.ParticipantCustomer); err != nil {
			t.Fatal(err)
		}
		assertWallet(t, f, f.customer, 0, 0)
		assertWallet(t, f, f.artist, 0, 0)
		p := f.summary(t)
		if p.Status != "CANCELLED" || p.SettledAt != nil {
			t.Fatal("settled unpaid payment")
		}
		var qr *string
		if err := db.NewRaw(`SELECT qr_payload FROM payments WHERE id=?`, f.paymentID).Scan(ctx, &qr); err != nil || qr != nil {
			t.Fatal("pending QR cleanup lost")
		}
	})
	t.Run("zero and one satang", func(t *testing.T) {
		for _, amount := range []int64{0, 1} {
			f := newFixture(t, db, amount)
			if err := f.cancel(ctx, order.ParticipantCustomer); err != nil {
				t.Fatal(err)
			}
			assertWallet(t, f, f.customer, 0, 0)
			count := 0
			if amount > 0 {
				count = 1
			}
			assertWallet(t, f, f.artist, amount, count)
		}
	})
	t.Run("concurrent cancellation credits once", func(t *testing.T) {
		f := newFixture(t, db, 10001)
		successes := runConcurrently(8, func() error { return f.cancel(ctx, order.ParticipantCustomer) })
		if successes != 1 {
			t.Fatalf("successful cancellations %d", successes)
		}
		assertWallet(t, f, f.customer, 5000, 1)
		assertWallet(t, f, f.artist, 5001, 1)
	})
	t.Run("ledger failure rolls back order and both wallets", func(t *testing.T) {
		f := newFixture(t, db, 10001)
		mustExec(t, db, `INSERT INTO payment_transactions(id,payment_id,type,amount_satang,transfer_status,recipient_user_id) VALUES(?,?,'REFUND_TO_CUSTOMER',1,'PENDING',?)`, uuid.New(), f.paymentID, f.customer)
		if err := f.cancel(ctx, order.ParticipantCustomer); err == nil {
			t.Fatal("expected duplicate ledger failure")
		}
		assertWallet(t, f, f.customer, 0, 0)
		assertWallet(t, f, f.artist, 0, 0)
		p := f.summary(t)
		if p.Status != "HELD" || p.SettledAt != nil || p.ArtistPayoutTransactionID != nil {
			t.Fatal("payment changes did not roll back")
		}
		var status string
		if err := db.NewRaw(`SELECT status FROM orders WHERE id=?`, f.orderID).Scan(ctx, &status); err != nil || status != "IN_PROCESS" {
			t.Fatal("order did not roll back")
		}
	})
	t.Run("wallet update failure rolls back credit", func(t *testing.T) {
		f := newFixture(t, db, 10001)
		mustExec(t, db, `INSERT INTO wallets(user_id,balance_satang) VALUES(?,?)`, f.customer, int64(math.MaxInt64))
		if err := f.cancel(ctx, order.ParticipantArtist); err == nil {
			t.Fatal("expected balance overflow")
		}
		p := f.summary(t)
		if p.Status != "HELD" || p.RefundTransactionID != nil {
			t.Fatal("payment or ledger did not roll back")
		}
		wallet, err := f.wallet.GetWallet(ctx, f.customer)
		if err != nil || wallet.BalanceSatang != math.MaxInt64 {
			t.Fatal("wallet changed on failure")
		}
		var count int
		if err := db.NewRaw(`SELECT COUNT(*) FROM wallet_entries WHERE user_id=?`, f.customer).Scan(ctx, &count); err != nil || count != 0 {
			t.Fatal("credit entry survived rollback")
		}
	})
	t.Run("last version automatically credits artist", func(t *testing.T) {
		f := newFixture(t, db, 10001)
		for version := 1; version < order.MaxDeliverableVersions; version++ {
			mustExec(t, db, `INSERT INTO order_deliverables(id,order_id,version,decision,comment,original_image_key,preview_image_key) VALUES(?,?,?,'REJECTED','revise','original','preview')`, uuid.New(), f.orderID, version)
		}
		if _, err := f.orders.CreateDeliverable(ctx, f.artist, f.orderID, &order.Deliverable{ID: uuid.New(), OriginalImageKey: "original", PreviewImageKey: "preview", CreatedAt: time.Now(), UpdatedAt: time.Now()}); err != nil {
			t.Fatal(err)
		}
		assertWallet(t, f, f.artist, 10001, 1)
	})
	t.Run("auto approval batch credits every order", func(t *testing.T) {
		f := newFixture(t, db, 10001)
		second := uuid.New()
		mustExec(t, db, `INSERT INTO orders(id,customer_id,artist_id,name,customer_description,price_satang_order,status,deadline_at) VALUES(?,?,?,'wallet order','',10001,'IN_PROCESS',NOW()+INTERVAL '7 days')`, second, f.customer, f.artist)
		mustExec(t, db, `INSERT INTO payments(id,order_id,amount_satang,status) VALUES(?,?,10001,'HELD')`, uuid.New(), second)
		for _, id := range []uuid.UUID{f.orderID, second} {
			mustExec(t, db, `INSERT INTO order_deliverables(id,order_id,version,original_image_key,preview_image_key,created_at) VALUES(?,?,1,'original','preview',NOW()-INTERVAL '4 days')`, uuid.New(), id)
		}
		for range 2 {
			if _, err := f.orders.AutoApproveExpiredDeliverables(ctx, time.Now().Add(-72*time.Hour), time.Now()); err != nil {
				t.Fatal(err)
			}
		}
		assertWallet(t, f, f.artist, 20002, 2)
	})
	t.Run("shared recipient accumulates concurrent order credits", func(t *testing.T) {
		f := newFixture(t, db, 10001)
		second := uuid.New()
		mustExec(t, db, `INSERT INTO orders(id,customer_id,artist_id,name,customer_description,price_satang_order,status,deadline_at) VALUES(?,?,?,'wallet order','',10001,'IN_PROCESS',NOW()+INTERVAL '7 days')`, second, f.customer, f.artist)
		mustExec(t, db, `INSERT INTO payments(id,order_id,amount_satang,status) VALUES(?,?,10001,'HELD')`, uuid.New(), second)
		start := make(chan struct{})
		results := make(chan error, 2)
		for _, id := range []uuid.UUID{f.orderID, second} {
			go func() {
				<-start
				results <- f.orders.CancelOrder(ctx, order.ParticipantArtist, f.artist, &order.OrderCancellation{ID: uuid.New(), OrderID: id, Reason: "OTHER", CreatedAt: time.Now()})
			}()
		}
		close(start)
		for range 2 {
			if err := <-results; err != nil {
				t.Fatal(err)
			}
		}
		assertWallet(t, f, f.customer, 20002, 2)
		page, err := f.wallet.ListWalletTransactions(ctx, f.customer, payment.TransactionQuery{Limit: 1, Offset: 1})
		if err != nil || page.Total != 2 || len(page.Transactions) != 1 {
			t.Fatalf("pagination %+v %v", page, err)
		}
		empty, err := f.wallet.ListWalletTransactions(ctx, f.customer, payment.TransactionQuery{Limit: 1, Offset: 50})
		if err != nil || empty.Total != 2 || len(empty.Transactions) != 0 {
			t.Fatalf("empty page %+v %v", empty, err)
		}
		assertWallet(t, f, uuid.New(), 0, 0)
	})
}

func runConcurrently(n int, fn func() error) int {
	start := make(chan struct{})
	results := make(chan error, n)
	var wg sync.WaitGroup
	for range n {
		wg.Add(1)
		go func() { defer wg.Done(); <-start; results <- fn() }()
	}
	close(start)
	wg.Wait()
	close(results)
	successes := 0
	for err := range results {
		if err == nil {
			successes++
		}
	}
	return successes
}

func TestWalletCreditMigration(t *testing.T) {
	ctx := context.Background()
	pg := apptest.StartPostgres(ctx, t)
	db, err := database.NewPostgresDB(config.Database{DSN: pg.DSN})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	goose.SetBaseFS(appmigrations.Migrations)
	if err := goose.DownToContext(ctx, db.DB, ".", 20261103000000); err != nil {
		t.Fatal(err)
	}
	f := newFixture(t, db, 3)
	tid := uuid.New()
	mustExec(t, db, `UPDATE payments SET status='SETTLED' WHERE id=?`, f.paymentID)
	mustExec(t, db, `INSERT INTO payment_transactions(id,payment_id,type,amount_satang,transfer_status,transfer_ref,transferred_at) VALUES(?,?,'REFUND_TO_CUSTOMER',3,'SUCCESS','mock-wallet-old',NOW())`, tid, f.paymentID)
	mustExec(t, db, `INSERT INTO wallet_entries(id,user_id,payment_transaction_id,kind,amount_satang) VALUES(?,?,?,'REFUND_TO_CUSTOMER',3)`, uuid.New(), f.customer, tid)
	if err := goose.UpContext(ctx, db.DB, "."); err != nil {
		t.Fatal(err)
	}
	assertWallet(t, f, f.customer, 3, 1)
	// Reach the wallet migration without touching its data, then test its own atomic guard.
	if err := goose.DownToContext(ctx, db.DB, ".", 20261104000000); err != nil {
		t.Fatal(err)
	}
	before := snapshotWalletState(t, db)
	if err := goose.DownContext(ctx, db.DB, "."); err == nil || !strings.Contains(err.Error(), "Cannot roll back wallet credit migration") {
		t.Fatalf("expected wallet rollback guard, got %v", err)
	}
	if after := snapshotWalletState(t, db); after != before {
		t.Fatal("rejected wallet rollback changed schema, version or financial data")
	}
	assertWallet(t, f, f.customer, 3, 1)
	if err := goose.UpContext(ctx, db.DB, "."); err != nil {
		t.Fatal(err)
	}
	assertWallet(t, f, f.customer, 3, 1)
	// A rejected rollback must leave the application able to settle subsequent orders.
	next := newFixture(t, db, 10001)
	if err := next.cancel(ctx, order.ParticipantCustomer); err != nil {
		t.Fatal(err)
	}
	assertWallet(t, next, next.customer, 5000, 1)
	assertWallet(t, next, next.artist, 5001, 1)
	assertWallet(t, f, f.customer, 3, 1)
}
