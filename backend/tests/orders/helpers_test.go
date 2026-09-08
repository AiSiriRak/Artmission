//go:build integration

package orders

import (
	"context"
	"time"

	"github.com/AiSiriRak/Artmission/backend/tests/internal/apptest"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// orderRow is a minimal bun model for the orders table, defined locally
// rather than imported from internal/adapters/postgres (whose orderModel
// is unexported). It exists purely to seed fixture rows — there is no
// create-order HTTP endpoint yet (see internal/modules/order/port.go) — so
// this is the one deliberate exception to "every precondition goes through
// the real HTTP API" in this suite.
type orderRow struct {
	bun.BaseModel `bun:"table:orders"`

	ID          uuid.UUID `bun:"id,pk"`
	CustomerID  uuid.UUID `bun:"customer_id"`
	ArtistID    uuid.UUID `bun:"artist_id"`
	Description string    `bun:"description"`
	Price       float64   `bun:"price"`
	Status      string    `bun:"status"`
	CreatedAt   time.Time `bun:"created_at"`
	UpdatedAt   time.Time `bun:"updated_at"`
}

// seedOrder inserts one order row directly against the database and
// returns its ID (as returned by the API: uuid.String()) and seeded status.
func seedOrder(customerID, artistID string) (string, string, error) {
	now := time.Now()
	row := &orderRow{
		ID:          uuid.New(),
		CustomerID:  uuid.MustParse(customerID),
		ArtistID:    uuid.MustParse(artistID),
		Description: "A custom commission",
		Price:       100,
		Status:      "PENDING",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if _, err := app.DB.NewInsert().Model(row).Exec(context.Background()); err != nil {
		return "", "", err
	}
	return row.ID.String(), row.Status, nil
}

// sharedArtist returns an artist account backing every order seeded in
// this scenario — an FK requirement (orders.artist_id), not itself the
// point under test, so one is created lazily and reused.
func (o *ordersContext) sharedArtist() (apptest.Account, error) {
	if o.artist.ID != "" {
		return o.artist, nil
	}
	artist, err := apptest.RegisterArtist(app, apptest.NewClient(app.BaseURL()), "I paint custom portraits.")
	if err != nil {
		return apptest.Account{}, err
	}
	o.artist = artist
	return artist, nil
}
