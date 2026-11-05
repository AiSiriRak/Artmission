//go:build integration

package payments

import (
	"context"
	"testing"

	"github.com/AiSiriRak/Artmission/backend/internal/pkg/config"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/database"
	appmigrations "github.com/AiSiriRak/Artmission/backend/internal/pkg/migrations"
	"github.com/AiSiriRak/Artmission/backend/tests/internal/apptest"
	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
)

func TestSettlementMigrationFromLegacySchema(t *testing.T) {
	ctx := context.Background()
	pg := apptest.StartPostgres(ctx, t)
	db, err := database.NewPostgresDB(config.Database{DSN: pg.DSN})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	goose.SetBaseFS(appmigrations.Migrations)
	if err := goose.DownToContext(ctx, db.DB, ".", 20261013100000); err != nil {
		t.Fatal(err)
	}

	// Reproduce the recorded legacy migration and schema inspected in the local database.
	mustExec(t, db, `ALTER TABLE payments
		ADD COLUMN settled_at timestamptz,
		ADD COLUMN settlement_outcome text CHECK (settlement_outcome IS NULL OR settlement_outcome IN ('COMPLETED','DEADLINE_EXPIRED','ARTIST_CANCELLED','CUSTOMER_CANCELLED'));
		ALTER TABLE payment_transactions
		ADD COLUMN transfer_status text CHECK (transfer_status IS NULL OR transfer_status IN ('PENDING','SUCCESS','FAILED')),
		ADD COLUMN transfer_ref text, ADD COLUMN transferred_at timestamptz, ADD COLUMN failure_reason text;
		ALTER TABLE payment_transactions ADD CONSTRAINT payment_transactions_transfer_status_by_type_check
		CHECK ((type='CHARGE' AND transfer_status IS NULL) OR (type<>'CHARGE' AND transfer_status IS NOT NULL)) NOT VALID;
		CREATE UNIQUE INDEX payment_transactions_payment_id_type_key ON payment_transactions(payment_id,type);
		INSERT INTO goose_db_version(version_id,is_applied) VALUES(20261101100000,true);`)
	// Match the real chronology: legacy settlement was applied before QR generation claims.
	if err := goose.UpToContext(ctx, db.DB, ".", 20261102100000); err != nil {
		t.Fatal(err)
	}
	f := newFixture(t, db, 10001)
	mustExec(t, db, `UPDATE payments SET status='SETTLED',settled_at='2026-11-01T08:00:00Z',settlement_outcome='COMPLETED' WHERE id=?`, f.paymentID)
	mustExec(t, db, `INSERT INTO payment_transactions(id,payment_id,type,amount_satang,transfer_status,transfer_ref,transferred_at)
		VALUES(?,?,'RELEASE_TO_ARTIST',10001,'SUCCESS','legacy-success','2026-11-01T08:00:00Z')`, uuid.New(), f.paymentID)
	mustExec(t, db, `INSERT INTO payment_transactions(id,payment_id,type,amount_satang,transfer_status,failure_reason)
		VALUES(?,?,'REFUND_TO_CUSTOMER',1,'FAILED','legacy failure')`, uuid.New(), f.paymentID)
	partial := newFixture(t, db, 3)
	// Legacy SUCCESS did not require a reference/timestamp; preserve rather than invent them.
	mustExec(t, db, `INSERT INTO payment_transactions(id,payment_id,type,amount_satang,transfer_status)
		VALUES(?,?,'RELEASE_TO_ARTIST',3,'SUCCESS')`, uuid.New(), partial.paymentID)
	snapshot := func() string {
		t.Helper()
		var data string
		if err := db.NewRaw(`SELECT jsonb_build_object(
			'payments',(SELECT jsonb_agg(to_jsonb(p)-'source' ORDER BY p.id) FROM payments p),
			'transactions',(SELECT jsonb_agg(to_jsonb(t)-'recipient_user_id' ORDER BY t.id) FROM payment_transactions t)
		)::text`).Scan(ctx, &data); err != nil {
			t.Fatal(err)
		}
		return data
	}
	before := snapshot()
	for range 2 {
		if err := goose.UpContext(ctx, db.DB, "."); err != nil {
			t.Fatal(err)
		}
	}
	if after := snapshot(); after != before {
		t.Fatalf("legacy records changed\nbefore: %s\nafter: %s", before, after)
	}
	var indexes int
	if err := db.NewRaw(`SELECT COUNT(*) FROM pg_indexes WHERE tablename='payment_transactions'
		AND indexname IN ('payment_transactions_payment_id_type_key','payment_transactions_payment_type_key')`).Scan(ctx, &indexes); err != nil || indexes != 1 {
		t.Fatalf("unique index count %d err %v", indexes, err)
	}
	assertWallet(t, f, f.artist, 0, 0)
	assertWallet(t, f, f.customer, 0, 0)
	var validated bool
	if err := db.NewRaw(`SELECT convalidated FROM pg_constraint WHERE conrelid='payment_transactions'::regclass AND conname='payment_transactions_transfer_check'`).Scan(ctx, &validated); err != nil || validated {
		t.Fatalf("historical constraints should remain NOT VALID: %t %v", validated, err)
	}
	if _, err := db.NewRaw(`INSERT INTO payment_transactions(id,payment_id,type,amount_satang,transfer_status,recipient_user_id)
		VALUES(?,?,'REFUND_TO_CUSTOMER',1,'SUCCESS',?)`, uuid.New(), partial.paymentID, partial.customer).Exec(ctx); err == nil {
		t.Fatal("new SUCCESS accepted missing transfer metadata")
	}
	if err := goose.DownToContext(ctx, db.DB, ".", 20261102100000); err != nil {
		t.Fatal(err)
	}
	if after := snapshot(); after != before {
		t.Fatal("rollback changed legacy settlement columns or records")
	}
	if err := goose.UpContext(ctx, db.DB, "."); err != nil {
		t.Fatal(err)
	}
	if after := snapshot(); after != before {
		t.Fatal("reapplying migrations changed history")
	}
}
