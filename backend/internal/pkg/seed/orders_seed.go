package seed

import (
	"context"
	"fmt"
	"time"

	pgmodel "github.com/AiSiriRak/Artmission/backend/internal/adapters/postgres/model"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type orderSeed struct {
	Key                 string
	CustomerKey         string
	ArtistKey           string
	ArtworkKey          string
	Name                string
	CustomerDescription string
	Status              string
	Completed           bool // sets completed_at; only meaningful for SUCCESS
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
	Decision *string // nil | "APPROVED" | "REJECTED"
}

var deliverables = []deliverableSeed{
	{OrderKey: "order-3", Version: 1, Decision: nil},
	{OrderKey: "order-4", Version: 1, Decision: new("APPROVED")},
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
		var completedAt *time.Time
		if o.Completed {
			completedAt = &now
		}
		artworkKeyID := artworkID(o.ArtworkKey)
		rows[i] = pgmodel.Order{
			ID:                          orderID(o.Key),
			CustomerID:                  userID(o.CustomerKey),
			ArtistID:                    userID(o.ArtistKey),
			ArtworkID:                   &artworkKeyID,
			Name:                        o.Name,
			ArtworkNameSnapshot:         artwork.Name,
			ArtworkDescriptionSnapshot:  artwork.Description,
			PriceSatangSnapshot:         artwork.PriceSatang,
			MinimumDeadlineDaysSnapshot: artwork.MinimumDeadlineDays,
			CustomerDescription:         o.CustomerDescription,
			Status:                      o.Status,
			CompletedAt:                 completedAt,
			CreatedAt:                   now,
			UpdatedAt:                   now,
		}
	}
	return seedTable(ctx, d, rows, "id")
}

func seedOrdersDown(ctx context.Context, d deps) error {
	ids := make([]uuid.UUID, len(orders))
	for i, o := range orders {
		ids[i] = orderID(o.Key)
	}
	return deleteByIDs[pgmodel.Order](ctx, d, "id", ids)
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
	ids := make([]uuid.UUID, len(orders))
	for i, o := range orders {
		ids[i] = orderID(o.Key)
	}
	if len(ids) == 0 {
		return nil
	}
	if _, err := d.db.NewDelete().Model((*pgmodel.OrderDeliverable)(nil)).Where("order_id IN (?)", bun.List(ids)).Exec(ctx); err != nil {
		return fmt.Errorf("delete seeded order deliverables: %w", err)
	}
	for _, del := range deliverables {
		original, preview := deliverableImageKeys(del.OrderKey, del.Version)
		if err := deletePlaceholderPair(ctx, d.storage.Private, original, preview); err != nil {
			return fmt.Errorf("delete deliverable image for %s v%d: %w", del.OrderKey, del.Version, err)
		}
	}
	return nil
}
