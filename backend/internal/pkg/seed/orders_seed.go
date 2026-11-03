package seed

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	pgmodel "github.com/AiSiriRak/Artmission/backend/internal/adapters/postgres/model"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

const paymentQRValidityPeriod = 24 * time.Hour

type orderSeed struct {
	Key                 string
	CustomerKey         string
	ArtistKey           string
	ArtworkKey          string
	Name                string
	CustomerDescription string
	Status              string
	Completed           bool // sets completed_at; only meaningful for SUCCESS
	DeadlineDays        int  // days from now; 0 = the artwork's minimum deadline
}

var orders = []orderSeed{
	{
		Key: "order-1", CustomerKey: "customer-1", ArtistKey: "artist-1", ArtworkKey: "artwork-1",
		Name: "Portrait for mom's birthday", CustomerDescription: "Please use a warm color palette.",
		Status: "PENDING",
	},
	{
		Key: "order-2", CustomerKey: "customer-1", ArtistKey: "artist-1", ArtworkKey: "artwork-2",
		Name: "Anime avatar", CustomerDescription: "Blue hair, casual outfit.",
		Status: "NOT_PAID",
	},
	{
		Key: "order-3", CustomerKey: "customer-1", ArtistKey: "artist-2", ArtworkKey: "artwork-3",
		Name: "Dream landscape", CustomerDescription: "A mountain valley at sunrise.",
		Status: "IN_PROCESS",
	},
	{
		Key: "order-4", CustomerKey: "customer-1", ArtistKey: "artist-2", ArtworkKey: "artwork-4",
		Name: "Living room wall art", CustomerDescription: "Calm colors, minimal detail.",
		Status: "SUCCESS", Completed: true,
	},
	{
		Key: "order-5", CustomerKey: "customer-1", ArtistKey: "artist-3", ArtworkKey: "artwork-5",
		Name: "Line art self-portrait", CustomerDescription: "Simple black line art.",
		Status: "CANCEL",
	},
	// Orders for the automated UI tests in frontend/e2e. Every test that changes an
	// order gets its own, so run `seed reseed` before each e2e run. All belong to
	// artist-2 with a 14-day deadline.
	{
		Key: "e2e-postpone-send", CustomerKey: "customer-1", ArtistKey: "artist-2", ArtworkKey: "artwork-4",
		Name: "E2E postpone - send request", CustomerDescription: "Order for the deadline extension request test.",
		Status: "IN_PROCESS", DeadlineDays: 14,
	},
	{
		Key: "e2e-postpone-validation", CustomerKey: "customer-1", ArtistKey: "artist-2", ArtworkKey: "artwork-4",
		Name: "E2E postpone - validation", CustomerDescription: "Order for the deadline extension validation tests.",
		Status: "IN_PROCESS", DeadlineDays: 14,
	},
	{
		Key: "e2e-postpone-pending", CustomerKey: "customer-1", ArtistKey: "artist-2", ArtworkKey: "artwork-4",
		Name: "E2E postpone - request pending", CustomerDescription: "Order that already has an extension request awaiting the customer.",
		Status: "IN_PROCESS", DeadlineDays: 14,
	},
	{
		Key: "e2e-draft-jpg", CustomerKey: "customer-1", ArtistKey: "artist-2", ArtworkKey: "artwork-4",
		Name: "E2E draft - JPEG", CustomerDescription: "Order for submitting a JPEG draft.",
		Status: "IN_PROCESS", DeadlineDays: 14,
	},
	{
		Key: "e2e-draft-png", CustomerKey: "customer-1", ArtistKey: "artist-2", ArtworkKey: "artwork-4",
		Name: "E2E draft - PNG", CustomerDescription: "Order for submitting a PNG draft.",
		Status: "IN_PROCESS", DeadlineDays: 14,
	},
	{
		Key: "e2e-draft-webp", CustomerKey: "customer-1", ArtistKey: "artist-2", ArtworkKey: "artwork-4",
		Name: "E2E draft - WebP", CustomerDescription: "Order for submitting a WebP draft.",
		Status: "IN_PROCESS", DeadlineDays: 14,
	},
	{
		Key: "e2e-draft-validation", CustomerKey: "customer-1", ArtistKey: "artist-2", ArtworkKey: "artwork-4",
		Name: "E2E draft - validation", CustomerDescription: "Order for the draft upload validation tests.",
		Status: "IN_PROCESS", DeadlineDays: 14,
	},
	{
		Key: "e2e-cancelled", CustomerKey: "customer-1", ArtistKey: "artist-2", ArtworkKey: "artwork-4",
		Name: "E2E cancelled order", CustomerDescription: "Already cancelled order, for read-only status checks.",
		Status: "CANCEL", DeadlineDays: 14,
	},
}

func orderID(key string) uuid.UUID { return id("order", key) }

func findArtwork(key string) artworkSeed {
	for _, a := range artworks {
		if a.Key == key {
			return a
		}
	}
	// Only reachable from a typo in this file's own seed data.
	panic("seed: unknown artwork key " + key)
}

type deliverableSeed struct {
	OrderKey string
	Version  int
	Decision string // "WAIT" | "APPROVED" | "REJECTED"
}

var deliverables = []deliverableSeed{
	{OrderKey: "order-3", Version: 1, Decision: "WAIT"},
	{OrderKey: "order-4", Version: 1, Decision: "APPROVED"},
}

// Every order seeded with status CANCEL needs a row here, mirroring what
// orderRepository.CancelOrder writes (one cancellation per order).
type cancellationSeed struct {
	OrderKey    string
	Reason      string
	Explanation string
}

var cancellations = []cancellationSeed{
	{OrderKey: "order-5", Reason: "NO_LONGER_NEEDED", Explanation: "I changed my mind about the style."},
	{OrderKey: "e2e-cancelled", Reason: "NO_LONGER_NEEDED", Explanation: "Cancelled for the automated tests."},
}

type delayRequestSeed struct {
	OrderKey    string
	ExtraDays   int // requested deadline = the order's deadline + ExtraDays
	Reason      string
	Explanation string
	Decision    string // "WAIT" | "APPROVED" | "REJECTED"
}

var delayRequests = []delayRequestSeed{
	{
		OrderKey: "e2e-postpone-pending", ExtraDays: 7,
		Reason: "NEED_MORE_TIME", Explanation: "I need more time for details.", Decision: "WAIT",
	},
}

func deliverableImageKeys(orderKey string, version int) (original, preview string) {
	return fmt.Sprintf("orders/%s/v%d/original.png", orderKey, version),
		fmt.Sprintf("orders/%s/v%d/preview.png", orderKey, version)
}

func seedOrdersUp(ctx context.Context, d deps) error {
	now := time.Now()
	rows := make([]pgmodel.Order, len(orders))
	for i, o := range orders {
		artwork := findArtwork(o.ArtworkKey)
		styleIDs := make([]uuid.UUID, len(artwork.StyleKeys))
		for index, styleKey := range artwork.StyleKeys {
			styleIDs[index] = styleID(styleKey)
		}
		var completedAt *time.Time
		if o.Completed {
			completedAt = &now
		}
		artworkKeyID := artworkID(o.ArtworkKey)
		rows[i] = pgmodel.Order{
			ID:         orderID(o.Key),
			CustomerID: userID(o.CustomerKey),
			ArtistID:   userID(o.ArtistKey),
			ArtworkID:  &artworkKeyID,
			ArtworkSnapshot: pgmodel.ArtworkSnapshot{
				ArtworkName: artwork.Name,
				CategoryID:  categoryID(artwork.CategoryKey),
				StyleIDs:    styleIDs,
			},
			Name:                o.Name,
			PriceSatangOrder:    artwork.PriceSatang,
			CustomerDescription: o.CustomerDescription,
			DeadlineAt:          now.AddDate(0, 0, deadlineDays(o, artwork)),
			Status:              o.Status,
			CompletedAt:         completedAt,
			CreatedAt:           now,
			UpdatedAt:           now,
		}
	}
	return seedTable(ctx, d, rows, "id")
}

func seedPaymentsUp(ctx context.Context, d deps) error {
	var seededOrder struct {
		ID           uuid.UUID `bun:"id"`
		AmountSatang int64     `bun:"price_satang_order"`
		TransitionAt time.Time `bun:"updated_at"`
	}
	if err := d.db.NewSelect().
		Table("orders").
		Column("id", "price_satang_order", "updated_at").
		Where("id = ?", orderID("order-2")).
		Where("status = ?", "NOT_PAID").
		Scan(ctx, &seededOrder); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("load NOT_PAID seed order for payment: %w", err)
	}

	expiresAt := seededOrder.TransitionAt.Add(paymentQRValidityPeriod)
	return seedTable(ctx, d, []pgmodel.Payment{{
		ID:           id("payment", "order-2"),
		OrderID:      seededOrder.ID,
		AmountSatang: seededOrder.AmountSatang,
		Status:       "PENDING",
		QRExpire:     &expiresAt,
		CreatedAt:    seededOrder.TransitionAt,
		UpdatedAt:    seededOrder.TransitionAt,
	}}, "order_id")
}

// orderIDsOfSeededUsers returns the IDs of every order placed by or assigned to
// a seeded user, including orders created later through the app. The down
// steps must clear all of them: orders.customer_id and orders.artist_id are
// ON DELETE RESTRICT, so one leftover order blocks deleting the seeded users.
func orderIDsOfSeededUsers(ctx context.Context, d deps) ([]uuid.UUID, error) {
	userIDs := make([]uuid.UUID, len(users))
	for i, u := range users {
		userIDs[i] = userID(u.Key)
	}
	var ids []uuid.UUID
	if err := d.db.NewSelect().
		Table("orders").
		Column("id").
		Where("customer_id IN (?) OR artist_id IN (?)", bun.List(userIDs), bun.List(userIDs)).
		Scan(ctx, &ids); err != nil {
		return nil, fmt.Errorf("find orders of seeded users: %w", err)
	}
	return ids, nil
}

func seedOrdersDown(ctx context.Context, d deps) error {
	ids, err := orderIDsOfSeededUsers(ctx, d)
	if err != nil {
		return err
	}
	return deleteModelRows[pgmodel.Order](ctx, d, "id", ids)
}

func seedReviewsDown(ctx context.Context, d deps) error {
	ids, err := orderIDsOfSeededUsers(ctx, d)
	if err != nil {
		return err
	}
	return deleteTableRows(ctx, d, "reviews", "order_id", ids)
}

func seedOrderCancellationsDown(ctx context.Context, d deps) error {
	ids, err := orderIDsOfSeededUsers(ctx, d)
	if err != nil {
		return err
	}
	return deleteTableRows(ctx, d, "order_cancellations", "order_id", ids)
}

func seedPaymentsDown(ctx context.Context, d deps) error {
	ids, err := orderIDsOfSeededUsers(ctx, d)
	if err != nil {
		return err
	}
	return deleteTableRows(ctx, d, "payments", "order_id", ids)
}

func seedWalletEntriesDown(ctx context.Context, d deps) error {
	userIDs := make([]uuid.UUID, len(users))
	for i, u := range users {
		userIDs[i] = userID(u.Key)
	}
	_, err := d.db.NewRaw(`WITH removed AS (DELETE FROM wallet_entries WHERE user_id IN (?) OR payment_transaction_id IN (
		SELECT t.id FROM payment_transactions t JOIN payments p ON p.id = t.payment_id
		JOIN orders o ON o.id = p.order_id WHERE o.customer_id IN (?) OR o.artist_id IN (?)
	) RETURNING user_id, amount_satang), totals AS (
		SELECT user_id, SUM(amount_satang) AS amount FROM removed GROUP BY user_id
	) UPDATE wallets w SET balance_satang = w.balance_satang - totals.amount, updated_at = NOW()
	FROM totals WHERE w.user_id = totals.user_id`, bun.List(userIDs), bun.List(userIDs), bun.List(userIDs)).Exec(ctx)
	if err != nil {
		return err
	}
	_, err = d.db.NewRaw(`DELETE FROM wallets WHERE user_id IN (?)`, bun.List(userIDs)).Exec(ctx)
	return err
}

func seedOrderDelayRequestsDown(ctx context.Context, d deps) error {
	ids, err := orderIDsOfSeededUsers(ctx, d)
	if err != nil {
		return err
	}
	return deleteTableRows(ctx, d, "order_delay_requests", "order_id", ids)
}

func seedOrderDeliverablesUp(ctx context.Context, d deps) error {
	now := time.Now()
	rows := make([]pgmodel.OrderDeliverable, len(deliverables))
	for i, del := range deliverables {
		seedKey := fmt.Sprintf("%s:v%d", del.OrderKey, del.Version)
		original, preview := deliverableImageKeys(del.OrderKey, del.Version)
		if err := uploadPlaceholderPair(ctx, d.storage.Private, seedKey, original, preview); err != nil {
			return fmt.Errorf("upload deliverable image for %q: %w", seedKey, err)
		}
		rows[i] = pgmodel.OrderDeliverable{
			ID:               id("order_deliverable", seedKey),
			OrderID:          orderID(del.OrderKey),
			Version:          del.Version,
			Decision:         del.Decision,
			OriginalImageKey: original,
			PreviewImageKey:  preview,
			CreatedAt:        now,
		}
	}
	return seedTable(ctx, d, rows, "id")
}

func seedOrderDeliverablesDown(ctx context.Context, d deps) error {
	ids, err := orderIDsOfSeededUsers(ctx, d)
	if err != nil {
		return err
	}
	if err := deleteModelRows[pgmodel.OrderDeliverable](ctx, d, "order_id", ids); err != nil {
		return err
	}
	for _, del := range deliverables {
		original, preview := deliverableImageKeys(del.OrderKey, del.Version)
		if err := deletePlaceholderPair(ctx, d.storage.Private, original, preview); err != nil {
			return fmt.Errorf("delete deliverable image for %s v%d: %w", del.OrderKey, del.Version, err)
		}
	}
	return nil
}

func seedOrderCancellationsUp(ctx context.Context, d deps) error {
	now := time.Now()
	rows := make([]pgmodel.OrderCancellation, len(cancellations))
	for i, c := range cancellations {
		rows[i] = pgmodel.OrderCancellation{
			ID:          id("order_cancellation", c.OrderKey),
			OrderID:     orderID(c.OrderKey),
			Reason:      c.Reason,
			Explanation: c.Explanation,
			CreatedAt:   now,
		}
	}
	return seedTable(ctx, d, rows, "id")
}

func seedOrderDelayRequestsUp(ctx context.Context, d deps) error {
	now := time.Now()
	rows := make([]pgmodel.OrderDelayRequest, len(delayRequests))
	for i, r := range delayRequests {
		var deadline time.Time
		if err := d.db.NewSelect().
			Table("orders").
			Column("deadline_at").
			Where("id = ?", orderID(r.OrderKey)).
			Scan(ctx, &deadline); err != nil {
			return fmt.Errorf("load deadline of seeded order %q: %w", r.OrderKey, err)
		}
		rows[i] = pgmodel.OrderDelayRequest{
			ID:                 id("order_delay_request", r.OrderKey),
			OrderID:            orderID(r.OrderKey),
			OriginalDeadlineAt: deadline,
			DeadlineRequestAt:  deadline.AddDate(0, 0, r.ExtraDays),
			Reason:             r.Reason,
			Explanation:        r.Explanation,
			CustomerDecision:   r.Decision,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
	}
	return seedTable(ctx, d, rows, "id")
}

// deadlineDays returns how many days from now the seeded order is due.
func deadlineDays(o orderSeed, artwork artworkSeed) int {
	if o.DeadlineDays > 0 {
		return o.DeadlineDays
	}
	return artwork.MinimumDeadlineDays
}
