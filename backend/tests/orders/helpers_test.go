//go:build integration

package orders

import (
	"context"
	"time"

	"github.com/AiSiriRak/Artmission/backend/tests/internal/apptest"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

const (
	seedArtworkName         = "Portrait commission"
	seedArtworkDescription  = "A hand-painted portrait"
	seedPriceSatang         = int64(10000)
	seedMinimumDeadlineDays = 7
	seedOrderName           = "Anniversary portrait"
	seedCustomerDescription = "Please use a blue background."
)

// orderRow is a minimal bun model for the orders table, defined locally
// rather than imported from internal/adapters/postgres (whose orderModel
// is unexported). It exists purely to seed fixture rows — there is no
// create-order HTTP endpoint yet (see internal/modules/order/port.go) — so
// this is the one deliberate exception to "every precondition goes through
// the real HTTP API" in this suite.
type orderRow struct {
	bun.BaseModel `bun:"table:orders"`

	ID                  uuid.UUID  `bun:"id,pk"`
	CustomerID          uuid.UUID  `bun:"customer_id"`
	ArtistID            uuid.UUID  `bun:"artist_id"`
	ArtworkID           uuid.UUID  `bun:"artwork_id"`
	Name                string     `bun:"name"`
	PriceSatangOrder    int64      `bun:"price_satang_order"`
	CustomerDescription string     `bun:"customer_description"`
	DeadlineAt          *time.Time `bun:"deadline_at"`
	Status              string     `bun:"status"`
	CreatedAt           time.Time  `bun:"created_at"`
	UpdatedAt           time.Time  `bun:"updated_at"`
}

// orderSeed describes one fixture order row's controllable fields. Every
// field left zero gets seedOrder's default (see below), so a scenario only
// has to name the field it actually cares about — e.g. an explicit
// UpdatedAt to control ViewOrders' default sort order deterministically.
// The artwork and order descriptions are not test inputs: every seeded
// order shares the fixed seed* constants above.
type orderSeed struct {
	CustomerID       string
	ArtistID         string
	Status           string
	PriceSatangOrder int64
	DeadlineAt       *time.Time
	UpdatedAt        time.Time
}

// seedOrder inserts one order row directly against the database and
// returns its ID (as returned by the API: uuid.String()).
func seedOrder(seed orderSeed) (string, error) {
	now := time.Now()

	status := seed.Status
	if status == "" {
		status = "PENDING"
	}
	price := seed.PriceSatangOrder
	if price == 0 {
		price = seedPriceSatang
	}
	categoryID, artworkID := uuid.New(), uuid.New()
	artistID := uuid.MustParse(seed.ArtistID)
	if _, err := app.DB.ExecContext(context.Background(), `
		INSERT INTO categories (id, label) VALUES (?, ?)
	`, categoryID, "Order fixture "+categoryID.String()); err != nil {
		return "", err
	}
	if _, err := app.DB.ExecContext(context.Background(), `
		INSERT INTO artworks (id, artist_id, category_id, name, description, price_satang, minimum_deadline_days)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, artworkID, artistID, categoryID, seedArtworkName, seedArtworkDescription, price, seedMinimumDeadlineDays); err != nil {
		return "", err
	}
	updatedAt := seed.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = now
	}
	deadlineAt := seed.DeadlineAt
	if deadlineAt == nil {
		defaultDeadline := now.AddDate(0, 0, seedMinimumDeadlineDays)
		deadlineAt = &defaultDeadline
	}

	row := &orderRow{
		ID:                  uuid.New(),
		CustomerID:          uuid.MustParse(seed.CustomerID),
		ArtistID:            artistID,
		ArtworkID:           artworkID,
		Name:                seedOrderName,
		PriceSatangOrder:    price,
		CustomerDescription: seedCustomerDescription,
		DeadlineAt:          deadlineAt,
		Status:              status,
		CreatedAt:           now,
		UpdatedAt:           updatedAt,
	}
	if _, err := app.DB.NewInsert().Model(row).Exec(context.Background()); err != nil {
		return "", err
	}
	return row.ID.String(), nil
}

// orderDeliverableRow is a minimal bun model for order_deliverables,
// mirroring orderRow's role: the one way this suite creates deliverable
// fixtures, since no submit-deliverable HTTP endpoint exists yet.
type orderDeliverableRow struct {
	bun.BaseModel `bun:"table:order_deliverables"`

	ID               uuid.UUID `bun:"id,pk"`
	OrderID          uuid.UUID `bun:"order_id"`
	Version          int       `bun:"version"`
	Decision         *string   `bun:"decision"`
	Comment          *string   `bun:"comment"`
	OriginalImageKey string    `bun:"original_image_key"`
	PreviewImageKey  string    `bun:"preview_image_key"`
	CreatedAt        time.Time `bun:"created_at"`
}

// seedDeliverable inserts one order_deliverables row directly against the
// database for orderID at version, with previewKey as its
// preview_image_key. decision is nil for a pending (undecided) row, or
// "APPROVED"/"REJECTED" for a terminal one — the order_deliverables_
// order_id_pending_key unique index allows at most one pending row per
// order, so every version before the latest in a multi-version fixture
// must pass a non-nil decision.
func seedDeliverable(orderID string, version int, previewKey string, decision *string) error {
	var comment *string
	if decision != nil && *decision == "REJECTED" {
		comment = new("Please revise this version.")
	}
	row := &orderDeliverableRow{
		ID:               uuid.New(),
		OrderID:          uuid.MustParse(orderID),
		Version:          version,
		Decision:         decision,
		Comment:          comment,
		OriginalImageKey: previewKey + ".original",
		PreviewImageKey:  previewKey,
		CreatedAt:        time.Now(),
	}
	_, err := app.DB.NewInsert().Model(row).Exec(context.Background())
	return err
}

// sharedCounterpart returns the account backing the other side of every
// order seeded for o.account in this scenario: an artist when the account
// under test is a customer, a customer when it's an artist. It's an FK
// requirement (orders.customer_id/artist_id), not itself the point under
// test, so one is created lazily and reused for the whole scenario.
func (o *ordersContext) sharedCounterpart() (apptest.Account, error) {
	if o.counterpart.ID != "" {
		return o.counterpart, nil
	}

	var (
		counterpart apptest.Account
		err         error
	)
	if o.accountRole == "artist" {
		counterpart, err = apptest.RegisterCustomer(app, apptest.NewClient(app.BaseURL()))
	} else {
		counterpart, err = apptest.RegisterArtist(app, apptest.NewClient(app.BaseURL()), "I paint custom portraits.")
	}
	if err != nil {
		return apptest.Account{}, err
	}
	o.counterpart = counterpart
	return counterpart, nil
}

// seedForAccount seeds one order with o.account on whichever side
// o.accountRole names and counterpart on the other, applying seed's
// remaining fields as-is.
func (o *ordersContext) seedForAccount(seed orderSeed, counterpart apptest.Account) (string, error) {
	if o.accountRole == "artist" {
		seed.CustomerID = counterpart.ID
		seed.ArtistID = o.account.ID
	} else {
		seed.CustomerID = o.account.ID
		seed.ArtistID = counterpart.ID
	}
	return seedOrder(seed)
}
