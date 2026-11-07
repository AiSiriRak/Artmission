//go:build integration

package payments

import (
	"context"
	"strings"
	"testing"

	"github.com/AiSiriRak/Artmission/backend/internal/pkg/config"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/database"
	appmigrations "github.com/AiSiriRak/Artmission/backend/internal/pkg/migrations"
	"github.com/AiSiriRak/Artmission/backend/tests/internal/apptest"
	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
	"github.com/uptrace/bun"
)

type walletMigrationSnapshot struct {
	version                int64
	data, balances, schema string
}

func snapshotWalletState(t *testing.T, db *bun.DB) walletMigrationSnapshot {
	t.Helper()
	ctx := context.Background()
	s := walletMigrationSnapshot{}
	var err error
	s.version, err = goose.GetDBVersionContext(ctx, db.DB)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.NewRaw(`SELECT jsonb_build_object(
		'payments',(SELECT jsonb_agg(to_jsonb(p) ORDER BY id) FROM payments p),
		'transactions',(SELECT jsonb_agg(to_jsonb(t) ORDER BY id) FROM payment_transactions t),
		'entries',(SELECT jsonb_agg(to_jsonb(w) ORDER BY id) FROM wallet_entries w)
	)::text`).Scan(ctx, &s.data); err != nil {
		t.Fatal(err)
	}
	if err := db.NewRaw(`SELECT jsonb_agg(to_jsonb(c) ORDER BY table_name,ordinal_position)::text
		FROM information_schema.columns c WHERE table_schema='public'
		AND table_name IN ('payments','payment_transactions','wallet_entries','wallets')`).Scan(ctx, &s.schema); err != nil {
		t.Fatal(err)
	}
	var hasWallets bool
	if err := db.NewRaw(`SELECT to_regclass('public.wallets') IS NOT NULL`).Scan(ctx, &hasWallets); err != nil {
		t.Fatal(err)
	}
	if hasWallets {
		if err := db.NewRaw(`SELECT COALESCE(jsonb_agg(to_jsonb(w) ORDER BY user_id),'[]'::jsonb)::text FROM wallets w`).Scan(ctx, &s.balances); err != nil {
			t.Fatal(err)
		}
	}
	return s
}

func TestWalletLedgerRollbackSafety(t *testing.T) {
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
	f := newFixture(t, db, 10001)
	mustExec(t, db, `UPDATE payments SET status='SETTLED',settlement_outcome='CUSTOMER_CANCELLED',settled_at=NOW() WHERE id=?`, f.paymentID)
	for _, c := range []struct {
		userID uuid.UUID
		kind   string
		amount int64
	}{
		{f.customer, "REFUND_TO_CUSTOMER", 5000}, {f.artist, "RELEASE_TO_ARTIST", 5001},
	} {
		tid := uuid.New()
		mustExec(t, db, `INSERT INTO payment_transactions(id,payment_id,type,amount_satang,transfer_status,transfer_ref,transferred_at)
			VALUES(?,?,?,?,'SUCCESS',?,NOW())`, tid, f.paymentID, c.kind, c.amount, "wallet-credit-"+tid.String())
		mustExec(t, db, `INSERT INTO wallet_entries(id,user_id,payment_transaction_id,kind,amount_satang) VALUES(?,?,?,?,?)`, uuid.New(), c.userID, tid, c.kind, c.amount)
	}
	before := snapshotWalletState(t, db)
	if err := goose.DownContext(ctx, db.DB, "."); err == nil || !strings.Contains(err.Error(), "Cannot roll back wallet ledger migration") {
		t.Fatalf("expected ledger rollback guard, got %v", err)
	}
	if after := snapshotWalletState(t, db); after != before {
		t.Fatal("rejected ledger rollback changed schema, version or financial history")
	}
	if err := goose.UpContext(ctx, db.DB, "."); err != nil {
		t.Fatal(err)
	}
	assertWallet(t, f, f.customer, 5000, 1)
	assertWallet(t, f, f.artist, 5001, 1)
}

func TestWalletRollbackWithoutFinancialData(t *testing.T) {
	ctx := context.Background()
	pg := apptest.StartPostgres(ctx, t)
	db, err := database.NewPostgresDB(config.Database{DSN: pg.DSN})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	goose.SetBaseFS(appmigrations.Migrations)
	if err := goose.DownToContext(ctx, db.DB, ".", 20261102100000); err != nil {
		t.Fatal(err)
	}
	var journalExists bool
	if err := db.NewRaw(`SELECT to_regclass('public.wallet_entries') IS NOT NULL`).Scan(ctx, &journalExists); err != nil || journalExists {
		t.Fatalf("empty ledger rollback: exists=%t err=%v", journalExists, err)
	}
	if err := goose.UpContext(ctx, db.DB, "."); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.NewRaw(`SELECT (SELECT COUNT(*) FROM wallet_entries)+(SELECT COUNT(*) FROM wallets)`).Scan(ctx, &count); err != nil || count != 0 {
		t.Fatalf("fresh wallet schema count=%d err=%v", count, err)
	}
}

func TestWalletRollbackWithBalanceWithoutJournal(t *testing.T) {
	ctx := context.Background()
	pg := apptest.StartPostgres(ctx, t)
	db, err := database.NewPostgresDB(config.Database{DSN: pg.DSN})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	goose.SetBaseFS(appmigrations.Migrations)
	f := newFixture(t, db, 10001)
	// Even inconsistent/manual balances must not disappear just because the journal is empty.
	mustExec(t, db, `INSERT INTO wallets(user_id,balance_satang) VALUES(?,17)`, f.customer)
	if err := goose.DownToContext(ctx, db.DB, ".", 20261104000000); err != nil {
		t.Fatal(err)
	}
	before := snapshotWalletState(t, db)
	if err := goose.DownContext(ctx, db.DB, "."); err == nil || !strings.Contains(err.Error(), "Cannot roll back wallet credit migration") {
		t.Fatalf("expected balance rollback guard, got %v", err)
	}
	if after := snapshotWalletState(t, db); after != before {
		t.Fatal("rejected balance rollback changed schema, version or balances")
	}
	wallet, err := f.wallet.GetWallet(ctx, f.customer)
	if err != nil || wallet.BalanceSatang != 17 {
		t.Fatalf("wallet %+v err %v", wallet, err)
	}
}
