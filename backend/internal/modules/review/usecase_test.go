package review_test

import (
	"context"
	"errors"
	"testing"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/review"
	"github.com/google/uuid"
)

type fakeRepo struct {
	created        *review.Review
	submitCalls    int
	submitErr      error
	reviews        []review.ReviewDetail
	listArtistID   uuid.UUID
	listErr        error
	orderReview    *review.ReviewDetail
	orderReviewErr error
	getReviewCalls int
	getOrderID     uuid.UUID
	getUserID      uuid.UUID
	artworkReviews []review.ReviewDetail
	listArtworkID  uuid.UUID
	artworkErr     error
	listQuery      review.ListQuery
}

type fakeProfileImageURLResolver string

func (resolver fakeProfileImageURLResolver) PublicURL(key string) string {
	return string(resolver) + key
}

func (f *fakeRepo) SubmitReview(
	_ context.Context,
	customerID uuid.UUID,
	input review.SubmitReviewInput,
) (*review.Review, error) {
	f.submitCalls++
	if f.submitErr != nil {
		return nil, f.submitErr
	}
	f.created = &review.Review{
		OrderID:    input.OrderID,
		CustomerID: customerID,
		Rating:     input.Rating,
		Comment:    input.Comment,
	}
	return f.created, nil
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

func (f *fakeRepo) GetReviewByOrderID(
	_ context.Context,
	userID uuid.UUID,
	orderID uuid.UUID,
) (*review.ReviewDetail, error) {
	f.getReviewCalls++
	f.getOrderID = orderID
	f.getUserID = userID
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
	orderID := uuid.New()

	repo := &fakeRepo{}
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
	if repo.submitCalls != 1 || result != repo.created ||
		result.OrderID != orderID || result.CustomerID != customerID {
		t.Fatal("expected one repository submission with the customer and order IDs")
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
	if result != nil || repo.created != nil || repo.submitCalls != 0 {
		t.Fatal("an invalid review must not be saved")
	}
}

// TestSubmitReview_RepositoryError preserves eligibility and storage errors.
func TestSubmitReview_RepositoryError(t *testing.T) {
	cases := []struct {
		name string
		err  error
	}{
		{name: "missing order", err: review.ErrOrderNotFound},
		{name: "not the owner", err: review.ErrNotOrderOwner},
		{name: "unfinished order", err: review.ErrOrderNotCompleted},
		{name: "duplicate review", err: review.ErrReviewAlreadyExists},
		{name: "database failure", err: errors.New("database unavailable")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRepo{submitErr: tc.err}
			result, err := review.NewReviewUsecase(repo).SubmitReview(
				context.Background(),
				uuid.New(),
				review.SubmitReviewInput{OrderID: uuid.New(), Rating: 5},
			)
			if !errors.Is(err, tc.err) || result != nil {
				t.Fatalf("expected repository error %v, got %v, %v", tc.err, result, err)
			}
			if repo.submitCalls != 1 || repo.created != nil {
				t.Fatal("expected one unsuccessful repository submission")
			}
		})
	}
}

// TestSubmitReview_MissingIDs rejects missing identities before querying the repository.
func TestSubmitReview_MissingIDs(t *testing.T) {
	cases := []struct {
		name       string
		customerID uuid.UUID
		orderID    uuid.UUID
	}{
		{name: "missing customer", orderID: uuid.New()},
		{name: "missing order ID", customerID: uuid.New()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRepo{}
			result, err := review.NewReviewUsecase(repo).SubmitReview(
				context.Background(),
				tc.customerID,
				review.SubmitReviewInput{OrderID: tc.orderID, Rating: 5},
			)
			if err == nil || result != nil || repo.submitCalls != 0 {
				t.Fatal("missing IDs must be rejected before the repository is called")
			}
		})
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

func TestGetArtistReviews_ResolvesProfileImageURL(t *testing.T) {
	key := "profiles/customer.png"
	repo := &fakeRepo{reviews: []review.ReviewDetail{
		{ProfileImageKey: &key},
	}}
	uc := review.NewReviewUsecase(
		repo,
		review.WithProfileImageURLResolver(fakeProfileImageURLResolver("https://cdn.test/")),
	)

	result, err := uc.GetArtistReviews(context.Background(), uuid.New(), review.ListQuery{})
	if err != nil {
		t.Fatalf("GetArtistReviews() error = %v", err)
	}
	if result.Reviews[0].ProfileImageURL == nil ||
		*result.Reviews[0].ProfileImageURL != "https://cdn.test/"+key {
		t.Fatalf("ProfileImageURL = %v, want %q", result.Reviews[0].ProfileImageURL, "https://cdn.test/"+key)
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

// TestGetOrderReview_Success forwards the caller and order IDs and returns the repository result.
func TestGetOrderReview_Success(t *testing.T) {
	userID, orderID := uuid.New(), uuid.New()
	item := &review.ReviewDetail{
		Review:   review.Review{ID: uuid.New(), OrderID: orderID, Rating: 5},
		Username: "Customer", Order: "Cat artwork",
	}
	repo := &fakeRepo{orderReview: item}

	result, err := review.NewReviewUsecase(repo).GetOrderReview(context.Background(), userID, orderID)
	if err != nil || result != item {
		t.Fatalf("expected the repository review, got %v, %v", result, err)
	}
	if repo.getReviewCalls != 1 || repo.getUserID != userID || repo.getOrderID != orderID {
		t.Fatal("expected one repository call with the user and order IDs")
	}
}

// TestGetOrderReview_RepositoryError preserves access, missing-data, and storage errors.
func TestGetOrderReview_RepositoryError(t *testing.T) {
	cases := []struct {
		name string
		err  error
	}{
		{name: "missing order", err: review.ErrOrderNotFound},
		{name: "not a participant", err: review.ErrNotOrderParticipant},
		{name: "missing review", err: review.ErrReviewNotFound},
		{name: "database failure", err: errors.New("review query failed")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRepo{orderReviewErr: tc.err}
			result, err := review.NewReviewUsecase(repo).GetOrderReview(context.Background(), uuid.New(), uuid.New())
			if !errors.Is(err, tc.err) || result != nil {
				t.Fatalf("expected error %v, got %v, %v", tc.err, result, err)
			}
			if repo.getReviewCalls != 1 {
				t.Fatal("expected one repository call")
			}
		})
	}
}

// TestGetOrderReview_MissingIDs rejects missing identities before querying the repository.
func TestGetOrderReview_MissingIDs(t *testing.T) {
	cases := []struct {
		name    string
		userID  uuid.UUID
		orderID uuid.UUID
	}{
		{name: "missing user", orderID: uuid.New()},
		{name: "missing order ID", userID: uuid.New()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRepo{}
			result, err := review.NewReviewUsecase(repo).GetOrderReview(context.Background(), tc.userID, tc.orderID)
			if err == nil || result != nil || repo.getReviewCalls != 0 {
				t.Fatal("missing IDs must be rejected before the repository is called")
			}
		})
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
			repo := &fakeRepo{}

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
				if result != nil || repo.created != nil || repo.submitCalls != 0 {
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
			repo := &fakeRepo{}

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
