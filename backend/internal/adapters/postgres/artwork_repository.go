package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	pgmodel "github.com/AiSiriRak/Artmission/backend/internal/adapters/postgres/model"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/artwork"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/baserepo"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

func newArtworkModel(item *artwork.Artwork, categoryID uuid.UUID) *pgmodel.Artwork {
	return &pgmodel.Artwork{
		ID:                  item.ID,
		ArtistID:            item.ArtistID,
		CategoryID:          categoryID,
		Name:                item.Name,
		Description:         item.Description,
		PriceSatang:         item.PriceSatang,
		MinimumDeadlineDays: item.MinimumDeadlineDays,
		CreatedAt:           item.CreatedAt,
		UpdatedAt:           item.UpdatedAt,
	}
}

func artworkModelToDomain(model *pgmodel.Artwork) artwork.Artwork {
	return artwork.Artwork{
		ID:                  model.ID,
		ArtistID:            model.ArtistID,
		Name:                model.Name,
		Category:            model.Category,
		Styles:              make([]string, 0),
		Description:         model.Description,
		Samples:             make([]artwork.Sample, 0),
		MinimumDeadlineDays: model.MinimumDeadlineDays,
		PriceSatang:         model.PriceSatang,
		CreatedAt:           model.CreatedAt,
		UpdatedAt:           model.UpdatedAt,
	}
}

type artworkStyleModel struct {
	ArtworkID uuid.UUID `bun:"artwork_id"`
	Label     string    `bun:"label"`
}

type artworkImageModel struct {
	ArtworkID uuid.UUID `bun:"artwork_id"`
	ImageURL  string    `bun:"image_url"`
}

type catalogIDLabel struct {
	ID    uuid.UUID `bun:"id"`
	Label string    `bun:"label"`
}

func assignCatalogLabels(ctx context.Context, idb bun.IDB, item *artwork.Artwork, categoryID uuid.UUID, styleIDs []uuid.UUID) error {
	var categoryLabel string
	if err := idb.NewSelect().
		Table("categories").
		Column("label").
		Where("id = ?", categoryID).
		Scan(ctx, &categoryLabel); err != nil {
		return err
	}
	item.Category = categoryLabel

	item.Styles = make([]string, 0, len(styleIDs))
	if len(styleIDs) == 0 {
		return nil
	}

	rows := make([]catalogIDLabel, 0, len(styleIDs))
	if err := idb.NewSelect().
		Table("styles").
		Column("id", "label").
		Where("id IN (?)", bun.List(styleIDs)).
		Scan(ctx, &rows); err != nil {
		return err
	}
	labelByID := make(map[uuid.UUID]string, len(rows))
	for _, row := range rows {
		labelByID[row.ID] = row.Label
	}
	for _, styleID := range styleIDs {
		label, ok := labelByID[styleID]
		if !ok {
			return sql.ErrNoRows
		}
		item.Styles = append(item.Styles, label)
	}
	return nil
}

type artworkRepository struct {
	exec baserepo.Executor
}

var _ artwork.Repository = (*artworkRepository)(nil)

func NewArtworkRepository(db *bun.DB) artwork.Repository {
	return &artworkRepository{exec: baserepo.NewExecutor(db)}
}

func catalogRefsExist(ctx context.Context, idb bun.IDB, categoryID uuid.UUID, styleIDs []uuid.UUID) error {
	categoryExists, err := idb.NewSelect().
		Table("categories").
		Where("id = ?", categoryID).
		Exists(ctx)
	if err != nil {
		return err
	}
	if !categoryExists {
		return artwork.ErrCategoryNotFound
	}
	if len(styleIDs) == 0 {
		return nil
	}

	found := make([]uuid.UUID, 0, len(styleIDs))
	if err := idb.NewSelect().
		Table("styles").
		Column("id").
		Where("id IN (?)", bun.List(styleIDs)).
		Scan(ctx, &found); err != nil {
		return err
	}
	foundSet := make(map[uuid.UUID]struct{}, len(found))
	for _, id := range found {
		foundSet[id] = struct{}{}
	}
	for _, styleID := range styleIDs {
		if _, ok := foundSet[styleID]; !ok {
			return artwork.ErrStyleNotFound
		}
	}
	return nil
}

func (repo *artworkRepository) Create(ctx context.Context, item *artwork.Artwork, categoryID uuid.UUID, styleIDs []uuid.UUID) error {
	err := repo.exec.Run(ctx, func(idb bun.IDB) error {
		if err := catalogRefsExist(ctx, idb, categoryID, styleIDs); err != nil {
			return err
		}

		if _, err := idb.NewInsert().Model(newArtworkModel(item, categoryID)).Exec(ctx); err != nil {
			return err
		}

		if len(styleIDs) > 0 {
			styles := make([]pgmodel.ArtworkStyle, len(styleIDs))
			for index, styleID := range styleIDs {
				styles[index] = pgmodel.ArtworkStyle{ArtworkID: item.ID, StyleID: styleID}
			}
			if _, err := idb.NewInsert().Model(&styles).Exec(ctx); err != nil {
				return err
			}
		}

		if len(item.Samples) > 0 {
			samples := make([]pgmodel.ArtworkImage, len(item.Samples))
			for index, sample := range item.Samples {
				samples[index] = pgmodel.ArtworkImage{
					ID:        uuid.New(),
					ArtworkID: item.ID,
					ImageURL:  sample.ImageURL,
					SortOrder: sample.SortOrder,
				}
			}
			if _, err := idb.NewInsert().Model(&samples).Exec(ctx); err != nil {
				return err
			}
		}
		return assignCatalogLabels(ctx, idb, item, categoryID, styleIDs)
	})
	if err != nil {
		if errors.Is(err, artwork.ErrCategoryNotFound) || errors.Is(err, artwork.ErrStyleNotFound) {
			return err
		}
		return apperror.Internal("failed to create artwork", err)
	}
	return nil
}

func (repo *artworkRepository) UpdateOwnedBy(ctx context.Context, item *artwork.Artwork, categoryID uuid.UUID, styleIDs []uuid.UUID, deletedSampleURLs []string) ([]string, error) {
	deletedURLs := make([]string, 0, len(deletedSampleURLs))
	timestamps := struct {
		CreatedAt time.Time `bun:"created_at"`
		UpdatedAt time.Time `bun:"updated_at"`
	}{}
	model := newArtworkModel(item, categoryID)
	err := repo.exec.Run(ctx, func(idb bun.IDB) error {
		if err := catalogRefsExist(ctx, idb, categoryID, styleIDs); err != nil {
			return err
		}

		err := idb.NewUpdate().
			Model(model).
			Column("category_id", "name", "description", "price_satang", "minimum_deadline_days", "updated_at").
			Where("id = ? AND artist_id = ?", item.ID, item.ArtistID).
			Returning("created_at, updated_at").
			Scan(ctx, &timestamps)
		if errors.Is(err, sql.ErrNoRows) {
			return artwork.ErrArtworkNotFound
		}
		if err != nil {
			return err
		}
		existingImages := make([]pgmodel.ArtworkImage, 0)
		if err := idb.NewSelect().
			Model(&existingImages).
			Column("id", "artwork_id", "image_url", "sort_order").
			Where("artwork_id = ?", item.ID).
			OrderExpr("sort_order ASC, id ASC").
			Scan(ctx); err != nil {
			return err
		}

		deleteSet := make(map[string]struct{}, len(deletedSampleURLs))
		for _, imageURL := range deletedSampleURLs {
			deleteSet[imageURL] = struct{}{}
		}
		retainedSamples := make([]artwork.Sample, 0, len(existingImages))
		for _, image := range existingImages {
			if _, deleted := deleteSet[image.ImageURL]; deleted {
				deletedURLs = append(deletedURLs, image.ImageURL)
				delete(deleteSet, image.ImageURL)
				continue
			}
			retainedSamples = append(retainedSamples, artwork.Sample{ImageURL: image.ImageURL})
		}
		if len(deleteSet) > 0 {
			return artwork.ErrSampleNotOwned
		}

		uploadedSamples := append([]artwork.Sample(nil), item.Samples...)
		item.Samples = append(retainedSamples, uploadedSamples...)
		for index := range item.Samples {
			item.Samples[index].SortOrder = index
		}

		if _, err := idb.NewDelete().Model(new(pgmodel.ArtworkStyle)).Where("artwork_id = ?", item.ID).Exec(ctx); err != nil {
			return err
		}
		if _, err := idb.NewDelete().Model(new(pgmodel.ArtworkImage)).Where("artwork_id = ?", item.ID).Exec(ctx); err != nil {
			return err
		}

		if len(styleIDs) > 0 {
			styles := make([]pgmodel.ArtworkStyle, len(styleIDs))
			for index, styleID := range styleIDs {
				styles[index] = pgmodel.ArtworkStyle{ArtworkID: item.ID, StyleID: styleID}
			}
			if _, err := idb.NewInsert().Model(&styles).Exec(ctx); err != nil {
				return err
			}
		}
		if len(item.Samples) > 0 {
			samples := make([]pgmodel.ArtworkImage, len(item.Samples))
			for index, sample := range item.Samples {
				samples[index] = pgmodel.ArtworkImage{
					ID:        uuid.New(),
					ArtworkID: item.ID,
					ImageURL:  sample.ImageURL,
					SortOrder: sample.SortOrder,
				}
			}
			if _, err := idb.NewInsert().Model(&samples).Exec(ctx); err != nil {
				return err
			}
		}

		return assignCatalogLabels(ctx, idb, item, categoryID, styleIDs)
	})
	if err != nil {
		if errors.Is(err, artwork.ErrArtworkNotFound) {
			return nil, artwork.ErrArtworkNotFound
		}
		if errors.Is(err, artwork.ErrSampleNotOwned) {
			return nil, artwork.ErrSampleNotOwned
		}
		if errors.Is(err, artwork.ErrCategoryNotFound) || errors.Is(err, artwork.ErrStyleNotFound) {
			return nil, err
		}
		return nil, apperror.Internal("failed to update artwork", err)
	}
	item.CreatedAt = timestamps.CreatedAt
	item.UpdatedAt = timestamps.UpdatedAt
	return deletedURLs, nil
}

func (repo *artworkRepository) DeleteOwnedBy(ctx context.Context, artworkID, artistID uuid.UUID) error {
	var result sql.Result
	err := repo.exec.Run(ctx, func(idb bun.IDB) error {
		var err error
		result, err = idb.NewDelete().
			Model(new(pgmodel.Artwork)).
			Where("id = ? AND artist_id = ?", artworkID, artistID).
			Exec(ctx)
		return err
	})
	if err != nil {
		return apperror.Internal("failed to delete artwork", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return apperror.Internal("failed to inspect artwork deletion", err)
	}
	if rows == 0 {
		return artwork.ErrArtworkNotFound
	}
	return nil
}

func (repo *artworkRepository) ListAllCategories(ctx context.Context) ([]artwork.Category, error) {
	models := make([]pgmodel.Category, 0)
	err := repo.exec.Run(ctx, func(idb bun.IDB) error {
		return idb.NewSelect().
			Model(&models).
			OrderExpr("cat.label ASC, cat.id ASC").
			Scan(ctx)
	})
	if err != nil {
		return nil, apperror.Internal("failed to list categories", err)
	}

	categories := make([]artwork.Category, len(models))
	for index, model := range models {
		categories[index] = artwork.Category{ID: model.ID, Label: model.Label}
	}
	return categories, nil
}

func (repo *artworkRepository) ListAllStyles(ctx context.Context) ([]artwork.Style, error) {
	models := make([]pgmodel.Style, 0)
	err := repo.exec.Run(ctx, func(idb bun.IDB) error {
		return idb.NewSelect().
			Model(&models).
			OrderExpr("sty.label ASC, sty.id ASC").
			Scan(ctx)
	})
	if err != nil {
		return nil, apperror.Internal("failed to list styles", err)
	}

	styles := make([]artwork.Style, len(models))
	for index, model := range models {
		styles[index] = artwork.Style{ID: model.ID, Label: model.Label}
	}
	return styles, nil
}

func (repo *artworkRepository) ListByArtistID(ctx context.Context, artistID uuid.UUID) ([]artwork.Artwork, error) {
	models := make([]pgmodel.Artwork, 0)
	var styleModels []artworkStyleModel
	var imageModels []artworkImageModel

	err := repo.exec.Run(ctx, func(idb bun.IDB) error {
		exists, err := idb.NewSelect().
			TableExpr("artist_profiles AS ap").
			Join("JOIN users AS u ON u.id = ap.user_id AND u.deleted_at IS NULL").
			Where("ap.user_id = ?", artistID).
			Exists(ctx)
		if err != nil {
			return err
		}
		if !exists {
			return artwork.ErrArtistNotFound
		}

		if err := idb.NewSelect().
			Model(&models).
			ColumnExpr("art.id, art.artist_id, art.name, art.description, art.price_satang, art.minimum_deadline_days, art.created_at, art.updated_at").
			ColumnExpr("c.label AS category").
			Join("JOIN categories AS c ON c.id = art.category_id").
			Where("art.artist_id = ?", artistID).
			OrderExpr("art.created_at DESC, art.id DESC").
			Scan(ctx); err != nil {
			return err
		}
		if len(models) == 0 {
			return nil
		}

		artworkIDs := make([]uuid.UUID, len(models))
		for index := range models {
			artworkIDs[index] = models[index].ID
		}

		var loadErr error
		styleModels, imageModels, loadErr = loadArtworkStylesAndImages(ctx, idb, artworkIDs)
		return loadErr
	})
	if err != nil {
		if errors.Is(err, artwork.ErrArtistNotFound) {
			return nil, artwork.ErrArtistNotFound
		}
		return nil, apperror.Internal("failed to list artist artworks", err)
	}

	artworks := make([]artwork.Artwork, len(models))
	indexByID := make(map[uuid.UUID]int, len(models))
	for index := range models {
		artworks[index] = artworkModelToDomain(&models[index])
		indexByID[models[index].ID] = index
	}
	for _, style := range styleModels {
		index := indexByID[style.ArtworkID]
		artworks[index].Styles = append(artworks[index].Styles, style.Label)
	}
	for _, image := range imageModels {
		index := indexByID[image.ArtworkID]
		artworks[index].Samples = append(artworks[index].Samples, artwork.Sample{ImageURL: image.ImageURL})
	}
	return artworks, nil
}

func (repo *artworkRepository) GetByID(ctx context.Context, artworkID uuid.UUID) (*artwork.Artwork, error) {
	var model pgmodel.Artwork
	var styleNames []string
	var samples []artwork.Sample

	err := repo.exec.Run(ctx, func(idb bun.IDB) error {
		// Get artwork with category label
		err := idb.NewSelect().
			Model(&model).
			Column("art.id",

				"art.artist_id",
				"art.name",
				"art.category_id",
				"cat.label AS category",
				"art.description",
				"art.price_satang",
				"art.minimum_deadline_days",
				"art.created_at",
				"art.updated_at").
			Join("JOIN categories AS cat ON cat.id = art.category_id").
			Where("art.id = ?", artworkID).
			Scan(ctx)

		if err != nil {
			return err
		}

		// Get style labels
		if err := idb.NewSelect().
			TableExpr("artwork_styles AS aws").
			Join("JOIN styles AS sty ON sty.id = aws.style_id").
			Column("sty.label").
			Where("aws.artwork_id = ?", artworkID).
			OrderExpr("sty.label ASC").
			Scan(ctx, &styleNames); err != nil {
			return err
		}

		// Get samples
		var imageModels []artworkImageModel
		if err := idb.NewSelect().
			TableExpr("artwork_images AS ai").
			Column("ai.image_url").
			Where("ai.artwork_id = ?", artworkID).
			OrderExpr("ai.id ASC").
			Scan(ctx, &imageModels); err != nil {
			return err
		}

		samples = make([]artwork.Sample, len(imageModels))
		for i, image := range imageModels {
			samples[i] = artwork.Sample{
				ImageURL:  image.ImageURL,
				SortOrder: i,
			}
		}

		return nil
	})

	if errors.Is(err, sql.ErrNoRows) {
		return nil, artwork.ErrArtworkNotFound
	}

	if err != nil {
		return nil, apperror.Internal("failed to get artwork", err)
	}

	return &artwork.Artwork{
		ID:                  model.ID,
		ArtistID:            model.ArtistID,
		Name:                model.Name,
		Category:            model.Category,
		Styles:              styleNames,
		Description:         model.Description,
		Samples:             samples,
		MinimumDeadlineDays: model.MinimumDeadlineDays,
		PriceSatang:         model.PriceSatang,
		CreatedAt:           model.CreatedAt,
		UpdatedAt:           model.UpdatedAt,
	}, nil
}

type artworkSearchModel struct {
	bun.BaseModel `bun:"table:artworks,alias:a"`

	ID                  uuid.UUID `bun:"id,pk"`
	ArtistID            uuid.UUID `bun:"artist_id"`
	Name                string    `bun:"name"`
	Category            string    `bun:"category,scanonly"`
	Description         string    `bun:"description"`
	PriceSatang         int64     `bun:"price_satang"`
	MinimumDeadlineDays int       `bun:"minimum_deadline_days"`
	CreatedAt           time.Time `bun:"created_at"`
	UpdatedAt           time.Time `bun:"updated_at"`
	ArtistName          string    `bun:"artist_name,scanonly"`
	ProfileImageKey     *string   `bun:"profile_image_key,scanonly"`
	ReviewScore         *float64  `bun:"review_score,scanonly"`
}

var artworkSearchOrderExpr = map[artwork.SearchSort]string{
	artwork.SearchSortNameAsc:         "a.name ASC, a.id ASC",
	artwork.SearchSortPriceAsc:        "a.price_satang ASC, a.name ASC, a.id ASC",
	artwork.SearchSortPriceDesc:       "a.price_satang DESC, a.name ASC, a.id ASC",
	artwork.SearchSortReviewScoreAsc:  "rs.review_score ASC NULLS FIRST, a.name ASC, a.id ASC",
	artwork.SearchSortReviewScoreDesc: "rs.review_score DESC NULLS LAST, a.name ASC, a.id ASC",
}

func (repo *artworkRepository) Search(ctx context.Context, query artwork.SearchQuery) (artwork.SearchPage, error) {
	orderExpr, ok := artworkSearchOrderExpr[query.Sort]
	if !ok {
		return artwork.SearchPage{}, apperror.Internal("unsupported search sort", nil)
	}

	page, err := baserepo.Paginate[artworkSearchModel](ctx, repo.exec, func(q *bun.SelectQuery) *bun.SelectQuery {
		q = q.
			ColumnExpr("a.id, a.artist_id, a.name, a.description, a.price_satang, a.minimum_deadline_days, a.created_at, a.updated_at").
			ColumnExpr("c.label AS category").
			ColumnExpr("u.username AS artist_name").
			ColumnExpr("u.profile_image_key").
			ColumnExpr("rs.review_score").
			Join("JOIN artist_profiles AS ap ON ap.user_id = a.artist_id").
			Join("JOIN users AS u ON u.id = ap.user_id AND u.deleted_at IS NULL").
			Join("JOIN categories AS c ON c.id = a.category_id").
			Join("LEFT JOIN (SELECT artist_id, ROUND(AVG(rating)::numeric, 1)::double precision AS review_score FROM reviews GROUP BY artist_id) AS rs ON rs.artist_id = a.artist_id")
		q = applyArtworkSearchFilters(q, query)
		return q.OrderExpr(orderExpr)
	}, baserepo.PaginationInput{Limit: query.Limit, Offset: query.Offset})
	if err != nil {
		if _, ok := errors.AsType[*apperror.Error](err); ok {
			return artwork.SearchPage{}, err
		}
		return artwork.SearchPage{}, apperror.Internal("failed to search artworks", err)
	}

	items := make([]artwork.SearchItem, len(page.Items))
	artworkIDs := make([]uuid.UUID, len(page.Items))
	indexByID := make(map[uuid.UUID]int, len(page.Items))
	for index, model := range page.Items {
		items[index] = artwork.SearchItem{
			Artwork: artwork.Artwork{
				ID:                  model.ID,
				ArtistID:            model.ArtistID,
				Name:                model.Name,
				Category:            model.Category,
				Styles:              make([]string, 0),
				Description:         model.Description,
				Samples:             make([]artwork.Sample, 0),
				MinimumDeadlineDays: model.MinimumDeadlineDays,
				PriceSatang:         model.PriceSatang,
				CreatedAt:           model.CreatedAt,
				UpdatedAt:           model.UpdatedAt,
			},
			Artist: artwork.ArtistSummary{
				ID:              model.ArtistID,
				Name:            model.ArtistName,
				ProfileImageKey: model.ProfileImageKey,
				ReviewScore:     model.ReviewScore,
			},
		}
		artworkIDs[index] = model.ID
		indexByID[model.ID] = index
	}
	if len(artworkIDs) == 0 {
		return artwork.SearchPage{Items: items, Total: page.Total}, nil
	}

	var styleModels []artworkStyleModel
	var imageModels []artworkImageModel
	err = repo.exec.Run(ctx, func(idb bun.IDB) error {
		var err error
		styleModels, imageModels, err = loadArtworkStylesAndImages(ctx, idb, artworkIDs)
		return err
	})
	if err != nil {
		return artwork.SearchPage{}, apperror.Internal("failed to load searched artwork details", err)
	}
	for _, style := range styleModels {
		index := indexByID[style.ArtworkID]
		items[index].Artwork.Styles = append(items[index].Artwork.Styles, style.Label)
	}
	for _, image := range imageModels {
		index := indexByID[image.ArtworkID]
		items[index].Artwork.Samples = append(items[index].Artwork.Samples, artwork.Sample{ImageURL: image.ImageURL})
	}
	return artwork.SearchPage{Items: items, Total: page.Total}, nil
}

func applyArtworkSearchFilters(q *bun.SelectQuery, query artwork.SearchQuery) *bun.SelectQuery {
	if query.ArtistName != "" {
		q = q.Where("u.username ILIKE ? ESCAPE '\\'", likeContains(query.ArtistName))
	}
	if query.Category != "" {
		q = q.Where("c.label = ?", query.Category)
	}
	if len(query.Styles) > 0 {
		q = q.Where(`a.id IN (
			SELECT aws.artwork_id
			FROM artwork_styles AS aws
			JOIN styles AS s ON s.id = aws.style_id
			WHERE s.label IN (?)
			GROUP BY aws.artwork_id
			HAVING COUNT(DISTINCT s.label) = ?
		)`, bun.In(query.Styles), len(query.Styles))
	}
	if query.MinPriceSatang != nil {
		q = q.Where("a.price_satang >= ?", *query.MinPriceSatang)
	}
	if query.MaxPriceSatang != nil {
		q = q.Where("a.price_satang <= ?", *query.MaxPriceSatang)
	}
	if query.MinReviewScore != nil {
		q = q.Where("rs.review_score >= ?", *query.MinReviewScore)
	}
	return q
}

func likeContains(value string) string {
	escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(value)
	return "%" + escaped + "%"
}

func loadArtworkStylesAndImages(ctx context.Context, idb bun.IDB, artworkIDs []uuid.UUID) ([]artworkStyleModel, []artworkImageModel, error) { //error
	styleModels := make([]artworkStyleModel, 0)
	imageModels := make([]artworkImageModel, 0)
	if len(artworkIDs) == 0 {
		return styleModels, imageModels, nil
	}
	if err := idb.NewSelect().
		TableExpr("artwork_styles AS aws").
		ColumnExpr("aws.artwork_id").
		ColumnExpr("s.label").
		Join("JOIN styles AS s ON s.id = aws.style_id").
		Where("aws.artwork_id IN (?)", bun.List(artworkIDs)).
		OrderExpr("aws.artwork_id ASC, s.label ASC, s.id ASC").
		Scan(ctx, &styleModels); err != nil {
		return nil, nil, err
	}
	if err := idb.NewSelect().
		TableExpr("artwork_images AS ai").
		ColumnExpr("ai.artwork_id, ai.image_url").
		Where("ai.artwork_id IN (?)", bun.List(artworkIDs)).
		OrderExpr("ai.artwork_id ASC, ai.sort_order ASC, ai.id ASC").
		Scan(ctx, &imageModels); err != nil {
		return nil, nil, err
	}
	return styleModels, imageModels, nil
}
