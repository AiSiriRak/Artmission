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
	order          *review.OrderForReview
	created        *review.Review
	reviews        []review.ReviewDetail
	listArtistID   uuid.UUID
	listErr        error
	orderReview    *review.ReviewDetail
	orderReviewErr error
	getReviewCalls int
	getOrderID     uuid.UUID
	artworkReviews []review.ReviewDetail
	listArtworkID  uuid.UUID
	artworkErr     error
	listQuery      review.ListQuery
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
	query review.ListQuery,
) (*review.ReviewPage, error) {
	f.listArtistID = artistID
	f.listQuery = query

	if f.listErr != nil {
		return nil, f.listErr
	}

	items := f.reviews
	if items == nil {
		items = []review.ReviewDetail{}
	}

	return &review.ReviewPage{
		Reviews: items,
		Total:   len(items),
	}, nil
}

func (f *fakeRepo) GetByOrderID(
	_ context.Context,
	orderID uuid.UUID,
) (*review.ReviewDetail, error) {
	f.getReviewCalls++
	f.getOrderID = orderID
	return f.orderReview, f.orderReviewErr
}

func (f *fakeRepo) ListByArtworkID(
	_ context.Context,
	artworkID uuid.UUID,
	query review.ListQuery,
) (*review.ReviewPage, error) {
	f.listArtworkID = artworkID
	f.listQuery = query

	if f.artworkErr != nil {
		return nil, f.artworkErr
	}

	items := f.artworkReviews
	if items == nil {
		items = []review.ReviewDetail{}
	}

	return &review.ReviewPage{
		Reviews: items,
		Total:   len(items),
	}, nil
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

	result, err := uc.GetArtistReviews(context.Background(), artistID, review.ListQuery{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.listArtistID != artistID {
		t.Fatal("repository received the wrong artist ID")
	}
	if len(result.Reviews) != 1 {
		t.Fatalf("review count = %d, want 1", len(result.Reviews))
	}

	got := result.Reviews[0]
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

	result, err := uc.GetArtistReviews(context.Background(), uuid.New(), review.ListQuery{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected an empty list, not nil")
	}
	if len(result.Reviews) != 0 {
		t.Fatalf("review count = %d, want 0", len(result.Reviews))
	}
}

// propagates a repository failure.
func TestGetArtistReviews_RepositoryError(t *testing.T) {
	expectedErr := errors.New("database unavailable")
	repo := &fakeRepo{
		listErr: expectedErr,
	}
	uc := review.NewReviewUsecase(repo)

	result, err := uc.GetArtistReviews(context.Background(), uuid.New(), review.ListQuery{})

	if !errors.Is(err, expectedErr) {
		t.Fatalf("error = %v, want %v", err, expectedErr)
	}
	if result != nil {
		t.Fatal("expected no reviews when retrieval fails")
	}
}

// TestGetOrderReview_Participants allows both parties to read the order's review.
func TestGetOrderReview_Participants(t *testing.T) {
	customerID, artistID, orderID := uuid.New(), uuid.New(), uuid.New()
	for _, userID := range []uuid.UUID{customerID, artistID} {
		t.Run(userID.String(), func(t *testing.T) {
			item := &review.ReviewDetail{
				Review:   review.Review{ID: uuid.New(), OrderID: orderID, Rating: 5},
				Username: "Customer", Order: "Cat artwork",
			}
			repo := &fakeRepo{
				order:       &review.OrderForReview{ID: orderID, CustomerID: customerID, ArtistID: artistID},
				orderReview: item,
			}
			result, err := review.NewReviewUsecase(repo).GetOrderReview(context.Background(), userID, orderID)
			if err != nil || result != item || repo.getOrderID != orderID {
				t.Fatalf("expected the order's review, got %v, %v", result, err)
			}
		})
	}
}

// TestGetOrderReview_Outsider rejects other users before reading the review.
func TestGetOrderReview_Outsider(t *testing.T) {
	orderID := uuid.New()
	repo := &fakeRepo{order: &review.OrderForReview{
		ID: orderID, CustomerID: uuid.New(), ArtistID: uuid.New(),
	}}
	result, err := review.NewReviewUsecase(repo).GetOrderReview(context.Background(), uuid.New(), orderID)
	if !errors.Is(err, review.ErrNotOrderParticipant) || result != nil {
		t.Fatalf("expected ErrNotOrderParticipant, got %v, %v", result, err)
	}
	if repo.getReviewCalls != 0 {
		t.Fatal("must not read the review for an unrelated user")
	}
}

// TestGetOrderReview_MissingReview reports when the order has no review yet.
func TestGetOrderReview_MissingReview(t *testing.T) {
	customerID, orderID := uuid.New(), uuid.New()
	repo := &fakeRepo{order: &review.OrderForReview{ID: orderID, CustomerID: customerID}}
	result, err := review.NewReviewUsecase(repo).GetOrderReview(context.Background(), customerID, orderID)
	if !errors.Is(err, review.ErrReviewNotFound) || result != nil {
		t.Fatalf("expected ErrReviewNotFound, got %v, %v", result, err)
	}
}

// TestGetOrderReview_RepositoryError preserves errors returned while reading a review.
func TestGetOrderReview_RepositoryError(t *testing.T) {
	customerID, orderID := uuid.New(), uuid.New()
	expectedErr := errors.New("review query failed")
	repo := &fakeRepo{
		order:          &review.OrderForReview{ID: orderID, CustomerID: customerID},
		orderReviewErr: expectedErr,
	}
	result, err := review.NewReviewUsecase(repo).GetOrderReview(context.Background(), customerID, orderID)
	if !errors.Is(err, expectedErr) || result != nil {
		t.Fatalf("expected repository error, got %v, %v", result, err)
	}
}

// TestGetReviewsByArtwork_Success requests reviews for the selected artwork.
func TestGetReviewsByArtwork_Success(t *testing.T) {
	artworkID, reviewID := uuid.New(), uuid.New()
	repo := &fakeRepo{artworkReviews: []review.ReviewDetail{
		{Review: review.Review{ID: reviewID, Rating: 4}, Username: "Customer", Order: "Cat artwork"},
	}}
	result, err := review.NewReviewUsecase(repo).GetReviewsByArtwork(context.Background(), artworkID, review.ListQuery{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.listArtworkID != artworkID || len(result.Reviews) != 1 || result.Reviews[0].ID != reviewID {
		t.Fatal("expected reviews requested for the selected artwork")
	}
}

// TestGetReviewsByArtwork_Empty returns an empty list when there are no reviews.
func TestGetReviewsByArtwork_Empty(t *testing.T) {
	result, err := review.NewReviewUsecase(&fakeRepo{}).GetReviewsByArtwork(context.Background(), uuid.New(), review.ListQuery{})
	if err != nil || result == nil || len(result.Reviews) != 0 {
		t.Fatalf("expected an empty list, got %v, %v", result, err)
	}
}

// TestGetReviewsByArtwork_RepositoryError preserves errors instead of returning an empty list.
func TestGetReviewsByArtwork_RepositoryError(t *testing.T) {
	expectedErr := errors.New("artwork review query failed")
	repo := &fakeRepo{artworkErr: expectedErr}
	result, err := review.NewReviewUsecase(repo).GetReviewsByArtwork(context.Background(), uuid.New(), review.ListQuery{})
	if !errors.Is(err, expectedErr) || result != nil {
		t.Fatalf("expected repository error, got %v, %v", result, err)
	}
}

// TestSubmitReview_RatingRange accepts only ratings from 1 to 5.
func TestSubmitReview_RatingRange(t *testing.T) {
	cases := []struct {
		name    string
		rating  int
		invalid bool
	}{
		{name: "negative", rating: -1, invalid: true},
		{name: "zero", rating: 0, invalid: true},
		{name: "minimum", rating: 1},
		{name: "maximum", rating: 5},
		{name: "above maximum", rating: 6, invalid: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			customerID, orderID := uuid.New(), uuid.New()
			repo := &fakeRepo{
				order: &review.OrderForReview{
					ID:         orderID,
					CustomerID: customerID,
					ArtistID:   uuid.New(),
					Status:     order.StatusSuccess,
				},
			}

			result, err := review.NewReviewUsecase(repo).SubmitReview(
				context.Background(),
				customerID,
				review.SubmitReviewInput{
					OrderID: orderID,
					Rating:  tc.rating,
				},
			)

			if tc.invalid {
				if !errors.Is(err, review.ErrInvalidRating) {
					t.Fatalf("error = %v, want ErrInvalidRating", err)
				}
				if result != nil || repo.created != nil {
					t.Fatal("an invalid rating must not be saved")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result == nil || repo.created == nil {
				t.Fatal("expected a saved review")
			}
			if repo.created.Rating != tc.rating {
				t.Fatalf("saved rating = %d, want %d", repo.created.Rating, tc.rating)
			}
		})
	}
}

// TestSubmitReview_Comment accepts optional comments and trims surrounding whitespace.
func TestSubmitReview_Comment(t *testing.T) {
	cases := []struct {
		name    string
		comment string
		want    string
	}{
		{name: "empty", comment: "", want: ""},
		{name: "spaces only", comment: "   ", want: ""},
		{name: "with text", comment: "  Great work!  ", want: "Great work!"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			customerID, orderID := uuid.New(), uuid.New()
			repo := &fakeRepo{
				order: &review.OrderForReview{
					ID:         orderID,
					CustomerID: customerID,
					ArtistID:   uuid.New(),
					Status:     order.StatusSuccess,
				},
			}

			result, err := review.NewReviewUsecase(repo).SubmitReview(
				context.Background(),
				customerID,
				review.SubmitReviewInput{
					OrderID: orderID,
					Rating:  5,
					Comment: &tc.comment,
				},
			)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result == nil || repo.created == nil {
				t.Fatal("expected a saved review")
			}

			if tc.want == "" {
				if repo.created.Comment != nil {
					t.Fatal("an empty comment should be stored as nil")
				}
			} else if repo.created.Comment == nil ||
				*repo.created.Comment != tc.want {
				t.Fatal("saved comment does not match the trimmed text")
			}
		})
	}
}
