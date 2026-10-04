package seed

import (
	"context"
	"fmt"
	"time"

	pgmodel "github.com/AiSiriRak/Artmission/backend/internal/adapters/postgres/model"
	"github.com/google/uuid"
)

type categorySeed struct {
	Key   string
	Label string
	ID    uuid.UUID
}

var categories = []categorySeed{
	{Key: "novel-cover", Label: "Novel Cover"},
	{Key: "illustration", Label: "Illustration"},
	{Key: "character-design", Label: "Character Design"},
	{Key: "background-design", Label: "Background Design"},
	{Key: "storyboard", Label: "Storyboard"},
	{Key: "cartoon-panel", Label: "Cartoon Panel"},
	{Key: "logo-design", Label: "Logo Design"},
	{Key: "social-media-post", Label: "Social Media Post"},
}

func categoryID(key string) uuid.UUID { return id("category", key) }

func (c categorySeed) uuid() uuid.UUID {
	if c.ID != uuid.Nil {
		return c.ID
	}
	return categoryID(c.Key)
}

type styleSeed struct {
	Key   string
	Label string
	ID    uuid.UUID
}

var styles = []styleSeed{
	{Key: "cartoon", Label: "Cartoon"},
	{Key: "realism", Label: "Realism"},
	{Key: "semi-realism", Label: "Semi-Realism"},
	{Key: "watercolor", Label: "Watercolor"},
	{Key: "acrylic", Label: "Acrylic"},
	{Key: "sketch", Label: "Sketch"},
	{Key: "abstract", Label: "Abstract"},
	{Key: "pixel-art", Label: "Pixel Art"},
	{Key: "fantasy", Label: "Fantasy"},
}

func styleID(key string) uuid.UUID { return id("style", key) }

func (s styleSeed) uuid() uuid.UUID {
	if s.ID != uuid.Nil {
		return s.ID
	}
	return styleID(s.Key)
}

// ArtistKey/CategoryKey/StyleKeys reference userSeed.Key/categorySeed.Key/styleSeed.Key by name rather than a literal ID.
type artworkSeed struct {
	Key                 string
	ArtistKey           string
	CategoryKey         string
	StyleKeys           []string
	Name                string
	Description         string
	PriceSatang         int64
	MinimumDeadlineDays int
}

var artworks = []artworkSeed{
	/*{
		Key: "artwork-1", ArtistKey: "artist-1", CategoryKey: "portrait", StyleKeys: []string{"realism"},
		Name: "Custom Portrait", Description: "A hand-painted-style digital portrait from your photo.",
		PriceSatang: 150000, MinimumDeadlineDays: 5,
	},
	{
		Key: "artwork-2", ArtistKey: "artist-1", CategoryKey: "portrait", StyleKeys: []string{"anime"},
		Name: "Anime Portrait", Description: "Portrait commission in anime style.",
		PriceSatang: 120000, MinimumDeadlineDays: 4,
	},
	{
		Key: "artwork-3", ArtistKey: "artist-2", CategoryKey: "landscape", StyleKeys: []string{"realism"},
		Name: "Fantasy Landscape", Description: "A detailed fantasy landscape concept piece.",
		PriceSatang: 250000, MinimumDeadlineDays: 7,
	},
	{
		Key: "artwork-4", ArtistKey: "artist-2", CategoryKey: "landscape", StyleKeys: []string{"realism", "minimalist"},
		Name: "Minimalist Vista", Description: "A calm, minimalist landscape piece.",
		PriceSatang: 100000, MinimumDeadlineDays: 3,
	},
	{
		Key: "artwork-5", ArtistKey: "artist-3", CategoryKey: "portrait", StyleKeys: []string{"minimalist"},
		Name: "Minimalist Line Portrait", Description: "Single-line-art style portrait.",
		PriceSatang: 80000, MinimumDeadlineDays: 2,
	},
	{
		Key: "artwork-6", ArtistKey: "artist-3", CategoryKey: "landscape", StyleKeys: []string{"minimalist"},
		Name: "Minimalist Skyline", Description: "Clean minimalist city skyline illustration.",
		PriceSatang: 90000, MinimumDeadlineDays: 3,
	},*/
	{
		Key: "artwork-1", ArtistKey: "artist-1", CategoryKey: "illustration", StyleKeys: []string{"realism", "watercolor"},
		Name: "Custom Portrait", Description: "A hand-painted-style digital portrait from your photo.",
		PriceSatang: 150000, MinimumDeadlineDays: 5,
	},
	{
		Key: "artwork-2", ArtistKey: "artist-1", CategoryKey: "character-design", StyleKeys: []string{"cartoon", "semi-realism"},
		Name: "Anime Portrait", Description: "Portrait commission in anime style.",
		PriceSatang: 120000, MinimumDeadlineDays: 4,
	},
	{
		Key: "artwork-3", ArtistKey: "artist-2", CategoryKey: "background-design", StyleKeys: []string{"fantasy", "realism"},
		Name: "Fantasy Landscape", Description: "A detailed fantasy landscape concept piece.",
		PriceSatang: 250000, MinimumDeadlineDays: 7,
	},
	{
		Key: "artwork-4", ArtistKey: "artist-2", CategoryKey: "background-design", StyleKeys: []string{"realism", "sketch"},
		Name: "Minimalist Vista", Description: "A calm, minimalist landscape piece.",
		PriceSatang: 100000, MinimumDeadlineDays: 3,
	},
	{
		Key: "artwork-5", ArtistKey: "artist-3", CategoryKey: "character-design", StyleKeys: []string{"sketch", "abstract"},
		Name: "Minimalist Line Portrait", Description: "Single-line-art style portrait.",
		PriceSatang: 80000, MinimumDeadlineDays: 2,
	},
	{
		Key: "artwork-6", ArtistKey: "artist-3", CategoryKey: "background-design", StyleKeys: []string{"pixel-art"},
		Name: "Minimalist Skyline", Description: "Clean minimalist city skyline illustration.",
		PriceSatang: 90000, MinimumDeadlineDays: 3,
	},
}

func artworkID(key string) uuid.UUID { return id("artwork", key) }

// artworkImageKeys returns the deterministic public-bucket object keys
// for artwork's single seeded image.
func artworkImageKeys(artworkKey string) (original, preview string) {
	return fmt.Sprintf("artworks/%s/original.png", artworkKey),
		fmt.Sprintf("artworks/%s/preview.png", artworkKey)
}

func seedCategoriesUp(ctx context.Context, d deps) error {
	now := time.Now()
	rows := make([]pgmodel.Category, len(categories))
	for i, c := range categories {
		rows[i] = pgmodel.Category{ID: c.uuid(), Label: c.Label, CreatedAt: now}
	}
	return seedTable(ctx, d, rows, "id")
}

func seedStylesUp(ctx context.Context, d deps) error {
	now := time.Now()
	rows := make([]pgmodel.Style, len(styles))
	for i, s := range styles {
		rows[i] = pgmodel.Style{ID: s.uuid(), Label: s.Label, CreatedAt: now}
	}
	return seedTable(ctx, d, rows, "id")
}

func seedArtworksUp(ctx context.Context, d deps) error {
	now := time.Now()
	rows := make([]pgmodel.Artwork, len(artworks))
	for i, a := range artworks {
		rows[i] = pgmodel.Artwork{
			ID:                  artworkID(a.Key),
			ArtistID:            userID(a.ArtistKey),
			CategoryID:          categoryID(a.CategoryKey),
			Name:                a.Name,
			Description:         a.Description,
			PriceSatang:         a.PriceSatang,
			MinimumDeadlineDays: a.MinimumDeadlineDays,
			CreatedAt:           now,
			UpdatedAt:           now,
		}
	}
	return seedTable(ctx, d, rows, "id")
}

func seedArtworkImagesUp(ctx context.Context, d deps) error {
	now := time.Now()
	rows := make([]pgmodel.ArtworkImage, len(artworks))
	for i, a := range artworks {
		original, preview := artworkImageKeys(a.Key)
		if err := uploadPlaceholderPair(ctx, d.storage.Public, "artwork:"+a.Key, original, preview); err != nil {
			return fmt.Errorf("upload image for artwork %q: %w", a.Key, err)
		}
		rows[i] = pgmodel.ArtworkImage{
			ID:        id("artwork_image", a.Key),
			ArtworkID: artworkID(a.Key),
			ImageURL:  d.storage.Public.PublicURL(preview),
			SortOrder: 0,
			CreatedAt: now,
		}
	}
	return seedTable(ctx, d, rows, "id")
}

func seedArtworkStylesUp(ctx context.Context, d deps) error {
	var rows []pgmodel.ArtworkStyle
	for _, a := range artworks {
		for _, styleKey := range a.StyleKeys {
			rows = append(rows, pgmodel.ArtworkStyle{ArtworkID: artworkID(a.Key), StyleID: styleID(styleKey)})
		}
	}
	return seedTable(ctx, d, rows, "artwork_id, style_id")
}

func seedArtworkImagesDown(ctx context.Context, d deps) error {
	for _, a := range artworks {
		original, preview := artworkImageKeys(a.Key)
		if err := deletePlaceholderPair(ctx, d.storage.Public, original, preview); err != nil {
			return fmt.Errorf("delete image for artwork %q: %w", a.Key, err)
		}
	}
	return nil
}
