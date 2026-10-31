package review_test

import (
	"context"
	"errors"
	"testing"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/review"
	"github.com/google/uuid"
)

type fakeRepo struct {
	order        *review.OrderForReview
	created      *review.Review
	reviews      []review.ReviewDetail
	listArtistID uuid.UUID
	listErr      error
}

func (f *fakeRepo) GetOrderForReview(
	_ context.Context,
	orderID uuid.UUID,
) (*review.OrderForReview, error) {
	if f.order == nil || f.order.ID != orderID {
		return nil, review.ErrOrderNotFound
	}

	return f.order, nil
}

func (f *fakeRepo) ExistsByOrderID(
	_ context.Context,
	orderID uuid.UUID,
) (bool, error) {
	return f.created != nil && f.created.OrderID == orderID, nil
}

func (f *fakeRepo) Create(
	_ context.Context,
	item *review.Review,
) error {
	f.created = item
	return nil
}

func (f *fakeRepo) ListByArtistID(
	_ context.Context,
	artistID uuid.UUID,
) ([]review.ReviewDetail, error) {
	f.listArtistID = artistID
	return f.reviews, f.listErr
}

// verifies that a rating alone can create a review
func TestSubmitReview_WithoutComment(t *testing.T) {
	customerID := uuid.New()
	artistID := uuid.New()
	orderID := uuid.New()

	repo := &fakeRepo{
		order: &review.OrderForReview{
			ID:         orderID,
			CustomerID: customerID,
			ArtistID:   artistID,
			Status:     order.StatusSuccess,
		},
	}
	uc := review.NewReviewUsecase(repo)

	result, err := uc.SubmitReview(
		context.Background(),
		customerID,
		review.SubmitReviewInput{
			OrderID: orderID,
			Rating:  5,
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || repo.created == nil {
		t.Fatal("expected a saved review")
	}
	if result.ID == uuid.Nil || result.CreatedAt.IsZero() {
		t.Fatal("expected a generated ID and creation time")
	}
	if result.OrderID != orderID ||
		result.CustomerID != customerID ||
		result.ArtistID != artistID {
		t.Fatal("review does not match the order's participants")
	}
	if result.Rating != 5 {
		t.Errorf("rating = %d, want 5", result.Rating)
	}
	if result.Comment != nil {
		t.Error("expected no comment")
	}
}

// verifies that a review without a rating is rejected
func TestSubmitReview_MissingRating(t *testing.T) {
	repo := &fakeRepo{}
	uc := review.NewReviewUsecase(repo)

	result, err := uc.SubmitReview(
		context.Background(),
		uuid.New(),
		review.SubmitReviewInput{
			OrderID: uuid.New(),
		},
	)

	if !errors.Is(err, review.ErrInvalidRating) {
		t.Fatalf("error = %v, want ErrInvalidRating", err)
	}
	if result != nil || repo.created != nil {
		t.Fatal("an invalid review must not be saved")
	}
}

// rejects reviews from another customer
func TestSubmitReview_NotOrderOwner(t *testing.T) {
	ownerID := uuid.New()
	anotherCustomerID := uuid.New()
	orderID := uuid.New()

	repo := &fakeRepo{
		order: &review.OrderForReview{
			ID:         orderID,
			CustomerID: ownerID,
			ArtistID:   uuid.New(),
			Status:     order.StatusSuccess,
		},
	}
	uc := review.NewReviewUsecase(repo)

	result, err := uc.SubmitReview(
		context.Background(),
		anotherCustomerID,
		review.SubmitReviewInput{
			OrderID: orderID,
			Rating:  5,
		},
	)

	if !errors.Is(err, review.ErrNotOrderOwner) {
		t.Fatalf("error = %v, want ErrNotOrderOwner", err)
	}
	if result != nil || repo.created != nil {
		t.Fatal("another customer's review must not be saved")
	}
}

// rejects reviews for unfinished orders
func TestSubmitReview_OrderNotCompleted(t *testing.T) {
	customerID := uuid.New()
	orderID := uuid.New()

	repo := &fakeRepo{
		order: &review.OrderForReview{
			ID:         orderID,
			CustomerID: customerID,
			ArtistID:   uuid.New(),
			Status:     order.StatusInProcess,
		},
	}
	uc := review.NewReviewUsecase(repo)

	result, err := uc.SubmitReview(
		context.Background(),
		customerID,
		review.SubmitReviewInput{
			OrderID: orderID,
			Rating:  5,
		},
	)

	if !errors.Is(err, review.ErrOrderNotCompleted) {
		t.Fatalf("error = %v, want ErrOrderNotCompleted", err)
	}
	if result != nil || repo.created != nil {
		t.Fatal("a review for an unfinished order must not be saved")
	}
}

// rejects a second review for the same order
func TestSubmitReview_Duplicate(t *testing.T) {
	customerID := uuid.New()
	orderID := uuid.New()

	existingReview := &review.Review{
		ID:         uuid.New(),
		OrderID:    orderID,
		CustomerID: customerID,
		Rating:     4,
	}

	repo := &fakeRepo{
		order: &review.OrderForReview{
			ID:         orderID,
			CustomerID: customerID,
			ArtistID:   uuid.New(),
			Status:     order.StatusSuccess,
		},
		created: existingReview,
	}
	uc := review.NewReviewUsecase(repo)

	result, err := uc.SubmitReview(
		context.Background(),
		customerID,
		review.SubmitReviewInput{
			OrderID: orderID,
			Rating:  5,
		},
	)

	if !errors.Is(err, review.ErrReviewAlreadyExists) {
		t.Fatalf("error = %v, want ErrReviewAlreadyExists", err)
	}
	if result != nil {
		t.Fatal("a duplicate review must not be returned")
	}
	if repo.created != existingReview {
		t.Fatal("the existing review must not be replaced")
	}
}

// verifies that reviews are retrieved for the requested artist
func TestGetArtistReviews_Success(t *testing.T) {
	artistID := uuid.New()
	reviewID := uuid.New()
	comment := "Great work!"

	repo := &fakeRepo{
		reviews: []review.ReviewDetail{
			{
				Review: review.Review{
					ID:       reviewID,
					ArtistID: artistID,
					Rating:   5,
					Comment:  &comment,
				},
				Username: "Customer",
				Order:    "Cat artwork",
			},
		},
	}
	uc := review.NewReviewUsecase(repo)

	result, err := uc.GetArtistReviews(context.Background(), artistID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.listArtistID != artistID {
		t.Fatal("repository received the wrong artist ID")
	}
	if len(result) != 1 {
		t.Fatalf("review count = %d, want 1", len(result))
	}

	got := result[0]
	if got.ID != reviewID ||
		got.Username != "Customer" ||
		got.Order != "Cat artwork" ||
		got.Rating != 5 ||
		got.Comment == nil ||
		*got.Comment != comment {
		t.Fatal("returned review details do not match")
	}
}

// returns an empty list when there are no reviews.
func TestGetArtistReviews_Empty(t *testing.T) {
	repo := &fakeRepo{}
	uc := review.NewReviewUsecase(repo)

	result, err := uc.GetArtistReviews(context.Background(), uuid.New())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected an empty list, not nil")
	}
	if len(result) != 0 {
		t.Fatalf("review count = %d, want 0", len(result))
	}
}

// propagates a repository failure.
func TestGetArtistReviews_RepositoryError(t *testing.T) {
	expectedErr := errors.New("database unavailable")
	repo := &fakeRepo{
		listErr: expectedErr,
	}
	uc := review.NewReviewUsecase(repo)

	result, err := uc.GetArtistReviews(context.Background(), uuid.New())

	if !errors.Is(err, expectedErr) {
		t.Fatalf("error = %v, want %v", err, expectedErr)
	}
	if result != nil {
		t.Fatal("expected no reviews when retrieval fails")
	}
}