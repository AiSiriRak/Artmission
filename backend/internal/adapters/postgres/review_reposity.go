package postgres

import (
	"context"
	"database/sql"
	"errors"

	pgmodel "github.com/AiSiriRak/Artmission/backend/internal/adapters/postgres/model"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/review"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/baserepo"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/uptrace/bun"
)

type reviewRepository struct {
	base baserepo.BaseRepo[pgmodel.Review]
	exec baserepo.Executor
	tx   baserepo.Transactioner
}

var _ review.ReviewRepository = (*reviewRepository)(nil)

// NewReviewRepository creates a PostgreSQL-backed review repository.
func NewReviewRepository(db *bun.DB) review.ReviewRepository {
	return &reviewRepository{
		base: baserepo.NewBaseRepo[pgmodel.Review](db, "review"),
		exec: baserepo.NewExecutor(db),
		tx:   baserepo.NewTransactioner(db),
	}
}

// GetOrderForReview loads the order's participants and current status.
func (r *reviewRepository) GetOrderForReview(
	ctx context.Context,
	orderID uuid.UUID,
) (*review.OrderForReview, error) {
	row := new(pgmodel.Order)

	err := r.exec.Run(ctx, func(idb bun.IDB) error {
		return idb.NewSelect().
			Model(row).
			Column("id", "customer_id", "artist_id", "status").
			Where("o.id = ?", orderID).
			Scan(ctx)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, review.ErrOrderNotFound
		}
		return nil, apperror.Internal("failed to get order for review", err)
	}

	return &review.OrderForReview{
		ID:         row.ID,
		CustomerID: row.CustomerID,
		ArtistID:   row.ArtistID,
		Status:     order.Status(row.Status),
	}, nil
}

// ExistsByOrderID checks whether the order already has a review.
func (r *reviewRepository) ExistsByOrderID(
	ctx context.Context,
	orderID uuid.UUID,
) (bool, error) {
	var exists bool

	err := r.exec.Run(ctx, func(idb bun.IDB) error {
		var err error
		exists, err = idb.NewSelect().
			Model(new(pgmodel.Review)).
			Where("r.order_id = ?", orderID).
			Exists(ctx)
		return err
	})
	if err != nil {
		return false, apperror.Internal("failed to check existing review", err)
	}

	return exists, nil
}

// newReviewModel converts a domain review into a database model.
func newReviewModel(item *review.Review) *pgmodel.Review {
	return &pgmodel.Review{
		ID:         item.ID,
		OrderID:    item.OrderID,
		CustomerID: item.CustomerID,
		ArtistID:   item.ArtistID,
		Rating:     item.Rating,
		Comment:    item.Comment,
		CreatedAt:  item.CreatedAt,
	}
}

// Create saves the review and refreshes the artist score in one transaction.
func (r *reviewRepository) Create(
	ctx context.Context,
	item *review.Review,
) error {
	err := r.tx.Transaction(ctx, func(ctx context.Context) error {
		return r.exec.Run(ctx, func(idb bun.IDB) error {
			var artistID uuid.UUID
			
			// serialize review submission for the same artist
			err := idb.NewRaw(`
				SELECT user_id
				FROM artist_profiles
				WHERE user_id = ?
				FOR UPDATE
			`, item.ArtistID).Scan(ctx, &artistID)

			if err != nil {
				return err
			}

			if err := r.base.Create(ctx, newReviewModel(item)); err != nil {
				return err
			}

			_,err = idb.NewRaw(`
				UPDATE artist_profiles
				SET review_score = (
					SELECT ROUND(AVG(rating)::numeric, 1)
					FROM reviews
					WHERE artist_id = ?
				)
				WHERE user_id = ?
			`, item.ArtistID, item.ArtistID).Exec(ctx)

			return err
		})
	})

	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) &&
		pgErr.Code == pgerrcode.UniqueViolation &&
		pgErr.ConstraintName == "reviews_order_id_key" {
		return review.ErrReviewAlreadyExists
	}

	return apperror.Internal("failed to create review and update artist score", err)
}

type reviewDetailRow struct {
	pgmodel.Review
	Username  string `bun:"username"`
	OrderName string `bun:"order_name"`
}

// newReviewDetailQuery selects reviews with customer and artwork names.
func newReviewDetailQuery(idb bun.IDB) *bun.SelectQuery {
	return idb.NewSelect().
		TableExpr("reviews AS r").
		ColumnExpr("r.*").
		ColumnExpr("customer.username AS username").
		ColumnExpr("COALESCE(NULLIF(o.artwork_snapshot->>'artwork_name', ''), a.name, o.name) AS order_name").
		Join("JOIN users AS customer ON customer.id = r.customer_id").
		Join("JOIN orders AS o ON o.id = r.order_id").
		Join("LEFT JOIN artworks AS a ON a.id = o.artwork_id")
}

// toDomain converts a database result into review details.
func (row reviewDetailRow) toDomain() review.ReviewDetail {
	return review.ReviewDetail{
		Review: review.Review{
			ID:         row.ID,
			OrderID:    row.OrderID,
			CustomerID: row.CustomerID,
			ArtistID:   row.ArtistID,
			Rating:     row.Rating,
			Comment:    row.Comment,
			CreatedAt:  row.CreatedAt,
		},
		Username: row.Username,
		Order:    row.OrderName,
	}
}

// ListByArtistID retrieves a page of artist reviews and the total count.
func (r *reviewRepository) ListByArtistID(
	ctx context.Context,
	artistID uuid.UUID,
	query review.ListQuery,
) (*review.ReviewPage, error) {
	var rows []reviewDetailRow
	var total int

	err := r.exec.Run(ctx, func(idb bun.IDB) error {
		var err error
		total, err = newReviewDetailQuery(idb).
			Where("r.artist_id = ?", artistID).
			OrderExpr("r.created_at DESC, r.id DESC").
			Limit(query.Limit).
			Offset(query.Offset).
			ScanAndCount(ctx, &rows)

		return err
	})
	if err != nil {
		return nil, apperror.Internal("failed to list artist reviews", err)
	}

	items := make([]review.ReviewDetail, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.toDomain())
	}

	return &review.ReviewPage{
		Reviews: items,
		Total:   total,
	}, nil
}

// GetByOrderID retrieves the review associated with an order.
func (r *reviewRepository) GetByOrderID(
	ctx context.Context,
	orderID uuid.UUID,
) (*review.ReviewDetail, error) {
	var row reviewDetailRow

	err := r.exec.Run(ctx, func(idb bun.IDB) error {
		return newReviewDetailQuery(idb).
			Where("r.order_id = ?", orderID).
			Scan(ctx, &row)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, review.ErrReviewNotFound
		}
		return nil, apperror.Internal("failed to get order review", err)
	}

	item := row.toDomain()
	return &item, nil
}

// ListByArtworkID retrieves a page of artwork reviews and the total count.
func (r *reviewRepository) ListByArtworkID(
	ctx context.Context,
	artworkID uuid.UUID,
	query review.ListQuery,
) (*review.ReviewPage, error) {
	var rows []reviewDetailRow
	var total int

	err := r.exec.Run(ctx, func(idb bun.IDB) error {
		var err error
		total, err = newReviewDetailQuery(idb).
			Where("o.artwork_id = ?", artworkID).
			OrderExpr("r.created_at DESC, r.id DESC").
			Limit(query.Limit).
			Offset(query.Offset).
			ScanAndCount(ctx, &rows)

		return err
	})
	if err != nil {
		return nil, apperror.Internal("failed to list artwork reviews", err)
	}

	items := make([]review.ReviewDetail, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.toDomain())
	}

	return &review.ReviewPage{
		Reviews: items,
		Total:   total,
	}, nil
}