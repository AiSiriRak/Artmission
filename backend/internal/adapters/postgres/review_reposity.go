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
}

var _ review.ReviewRepository = (*reviewRepository)(nil)

// NewReviewRepository creates a PostgreSQL-backed review repository.
func NewReviewRepository(db *bun.DB) review.ReviewRepository {
	return &reviewRepository{
		base: baserepo.NewBaseRepo[pgmodel.Review](db, "review"),
		exec: baserepo.NewExecutor(db),
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

// Create inserts a review and translates duplicate-order conflicts.
func (r *reviewRepository) Create(
	ctx context.Context,
	item *review.Review,
) error {
	err := r.base.Create(ctx, newReviewModel(item))
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) &&
		pgErr.Code == pgerrcode.UniqueViolation &&
		pgErr.ConstraintName == "reviews_order_id_key" {
		return review.ErrReviewAlreadyExists
	}

	return apperror.Internal("failed to create review", err)
}