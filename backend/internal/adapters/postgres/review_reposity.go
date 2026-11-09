package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

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

// SubmitReview checks the order, saves the review, and refreshes the score atomically.
func (r *reviewRepository) SubmitReview(
	ctx context.Context,
	customerID uuid.UUID,
	input review.SubmitReviewInput,
) (*review.Review, error) {
	var result *review.Review

	err := r.tx.Transaction(ctx, func(ctx context.Context) error {
		return r.exec.Run(ctx, func(idb bun.IDB) error {
			row := new(pgmodel.Order)
			// Keep ownership and status stable until the review is committed.
			err := idb.NewSelect().
				Model(row).
				Column("id", "customer_id", "artist_id", "status").
				Where("o.id = ?", input.OrderID).
				For("UPDATE").
				Scan(ctx)
			if errors.Is(err, sql.ErrNoRows) {
				return review.ErrOrderNotFound
			}
			if err != nil {
				return err
			}
			if row.CustomerID != customerID {
				return review.ErrNotOrderOwner
			}
			if order.Status(row.Status) != order.StatusSuccess {
				return review.ErrOrderNotCompleted
			}

			exists, err := idb.NewSelect().
				Model(new(pgmodel.Review)).
				Where("r.order_id = ?", input.OrderID).
				Exists(ctx)
			if err != nil {
				return err
			}
			if exists {
				return review.ErrReviewAlreadyExists
			}

			var artistID uuid.UUID

			// Serialize score updates from different orders for the same artist.
			err = idb.NewRaw(`
				SELECT user_id
				FROM artist_profiles
				WHERE user_id = ?
				FOR UPDATE
			`, row.ArtistID).Scan(ctx, &artistID)

			if err != nil {
				return err
			}

			result = &review.Review{
				ID:         uuid.New(),
				OrderID:    row.ID,
				CustomerID: customerID,
				ArtistID:   row.ArtistID,
				Rating:     input.Rating,
				Comment:    input.Comment,
				CreatedAt:  time.Now(),
			}
			if err := r.base.Create(ctx, newReviewModel(result)); err != nil {
				return err
			}

			_, err = idb.NewRaw(`
				UPDATE artist_profiles
				SET review_score = (
					SELECT ROUND(AVG(rating)::numeric, 1)
					FROM reviews
					WHERE artist_id = ?
				)
				WHERE user_id = ?
			`, row.ArtistID, row.ArtistID).Exec(ctx)

			return err
		})
	})

	if err == nil {
		return result, nil
	}

	var appErr *apperror.Error
	if errors.As(err, &appErr) {
		return nil, appErr
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) &&
		pgErr.Code == pgerrcode.UniqueViolation &&
		pgErr.ConstraintName == "reviews_order_id_key" {
		return nil, review.ErrReviewAlreadyExists
	}

	return nil, apperror.Internal("failed to submit review and update artist score", err)
}

type reviewDetailRow struct {
	pgmodel.Review
	Username        string  `bun:"username"`
	ProfileImageKey *string `bun:"profile_image_key"`
	OrderName       string  `bun:"order_name"`
	ArtworkName     string  `bun:"artwork_name"`
}

// newReviewDetailQuery selects reviews with customer and order names.
func newReviewDetailQuery(idb bun.IDB) *bun.SelectQuery {
	return idb.NewSelect().
		TableExpr("reviews AS r").
		ColumnExpr("r.*").
		ColumnExpr("customer.username AS username").
		ColumnExpr("customer.profile_image_key AS profile_image_key").
		ColumnExpr("o.name AS order_name").
		Join("JOIN users AS customer ON customer.id = r.customer_id").
		Join("JOIN orders AS o ON o.id = r.order_id")
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
		Username:        row.Username,
		ProfileImageKey: row.ProfileImageKey,
		Order:           row.OrderName,
		ArtworkName:     row.ArtworkName,
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
			ColumnExpr("o.artwork_snapshot->>'artwork_name' AS artwork_name").
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

// GetReviewByOrderID retrieves an order's review for its customer or artist.
func (r *reviewRepository) GetReviewByOrderID(
	ctx context.Context,
	userID uuid.UUID,
	orderID uuid.UUID,
) (*review.ReviewDetail, error) {
	var row struct {
		pgmodel.Review
		OrderCustomerID uuid.UUID `bun:"order_customer_id"`
		OrderArtistID   uuid.UUID `bun:"order_artist_id"`
		Username        string    `bun:"username"`
		ProfileImageKey *string   `bun:"profile_image_key"`
		OrderName       string    `bun:"order_name"`
	}

	err := r.exec.Run(ctx, func(idb bun.IDB) error {
		// Keep the order when no review exists so access and missing-review errors stay distinct.
		return idb.NewSelect().
			TableExpr("orders AS o").
			ColumnExpr("r.*").
			ColumnExpr("o.customer_id AS order_customer_id, o.artist_id AS order_artist_id").
			ColumnExpr("customer.username AS username").
			ColumnExpr("customer.profile_image_key AS profile_image_key").
			ColumnExpr("o.name AS order_name").
			Join("LEFT JOIN reviews AS r ON r.order_id = o.id").
			Join("JOIN users AS customer ON customer.id = o.customer_id").
			Where("o.id = ?", orderID).
			Scan(ctx, &row)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, review.ErrOrderNotFound
		}
		return nil, apperror.Internal("failed to get order review", err)
	}

	if userID != row.OrderCustomerID && userID != row.OrderArtistID {
		return nil, review.ErrNotOrderParticipant
	}
	if row.ID == uuid.Nil {
		return nil, review.ErrReviewNotFound
	}

	item := reviewDetailRow{
		Review:    row.Review,
		Username:        row.Username,
		ProfileImageKey: row.ProfileImageKey,
		OrderName:       row.OrderName,
	}.toDomain()
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
