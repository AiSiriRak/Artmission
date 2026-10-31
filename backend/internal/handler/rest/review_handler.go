package rest

import (
	"context"
	"net/http"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/auth"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/review"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/user"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type ReviewHandler struct {
	reviewUsecase review.ReviewUsecase
	authUsecase   auth.AuthUsecase
}

// NewReviewHandler creates a handler for customer reviews.
func NewReviewHandler(
	reviewUsecase review.ReviewUsecase,
	authUsecase auth.AuthUsecase,
) *ReviewHandler {
	return &ReviewHandler{
		reviewUsecase: reviewUsecase,
		authUsecase:   authUsecase,
	}
}

// Register adds the customer review endpoint to the API.
func (h *ReviewHandler) Register(api huma.API) {
	huma.Post(api, "/orders/{order_id}/review", h.submitReview,
		huma.OperationTags("reviews"),
		func(o *huma.Operation) {
			o.OperationID = "submit-review"
			o.Summary = "SubmitReview"
			o.Description = "Submit a review for a successful order as its customer."
			o.DefaultStatus = http.StatusCreated
			o.Middlewares = append(
				o.Middlewares,
				requireAuth(api, h.authUsecase),
				requireRole(api, user.RoleCustomer),
			)
		},
	)
	huma.Get(api, "/artists/{artist_id}/reviews", h.getArtistReviews,
		huma.OperationTags("reviews"),
		func(o *huma.Operation) {
			o.OperationID = "get-artist-reviews"
			o.Summary = "GetArtistReviews"
			o.Description = "Get an artist's reviews from newest to oldest."
			o.Middlewares = append(
				o.Middlewares,
				requireAuth(api, h.authUsecase),
				requireAnyRole(api, user.RoleCustomer, user.RoleArtist),
			)
		},
	)
}

type SubmitReviewInput struct {
	OrderID uuid.UUID `path:"order_id"`
	Body    struct {
		Rating  int     `json:"rating" minimum:"1" maximum:"5"`
		Comment *string `json:"comment,omitempty"`
	}
}

type reviewView struct {
	ID         uuid.UUID `json:"id"`
	OrderID    uuid.UUID `json:"order_id"`
	CustomerID uuid.UUID `json:"customer_id"`
	ArtistID   uuid.UUID `json:"artist_id"`
	Rating     int       `json:"rating"`
	Comment    *string   `json:"comment"`
	CreatedAt  time.Time `json:"created_at"`
}

type SubmitReviewOutput struct {
	Body reviewView
}

type GetArtistReviewsInput struct {
	ArtistID uuid.UUID `path:"artist_id"`
}

type reviewDetailView struct {
	reviewView
	Username string `json:"username"`
	Order    string `json:"order"`
}

type GetArtistReviewsOutput struct {
	Body struct {
		Reviews []reviewDetailView `json:"reviews"`
	}
}

// submitReview reads the authenticated customer and returns the created review.
func (h *ReviewHandler) submitReview(
	ctx context.Context,
	input *SubmitReviewInput,
) (*SubmitReviewOutput, error) {
	info, ok := authInfoFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("missing authentication")
	}

	result, err := h.reviewUsecase.SubmitReview(
		ctx,
		info.UserID,
		review.SubmitReviewInput{
			OrderID: input.OrderID,
			Rating:  input.Body.Rating,
			Comment: input.Body.Comment,
		},
	)
	if err != nil {
		return nil, mapAppError(err)
	}

	return &SubmitReviewOutput{
		Body: reviewView{
			ID:         result.ID,
			OrderID:    result.OrderID,
			CustomerID: result.CustomerID,
			ArtistID:   result.ArtistID,
			Rating:     result.Rating,
			Comment:    result.Comment,
			CreatedAt:  result.CreatedAt,
		},
	}, nil
}

// getArtistReviews returns an artist's reviews with customer and artwork names.
func (h *ReviewHandler) getArtistReviews(
	ctx context.Context,
	input *GetArtistReviewsInput,
) (*GetArtistReviewsOutput, error) {
	items, err := h.reviewUsecase.GetArtistReviews(ctx, input.ArtistID)
	if err != nil {
		return nil, mapAppError(err)
	}

	views := make([]reviewDetailView, 0, len(items))

	for _, item := range items {
		views = append(views, reviewDetailView{
			reviewView: reviewView{
				ID:         item.ID,
				OrderID:    item.OrderID,
				CustomerID: item.CustomerID,
				ArtistID:   item.ArtistID,
				Rating:     item.Rating,
				Comment:    item.Comment,
				CreatedAt:  item.CreatedAt,
			},
			Username: item.Username,
			Order:    item.Order,
		})
	}

	out := &GetArtistReviewsOutput{}
	out.Body.Reviews = views

	return out, nil
}