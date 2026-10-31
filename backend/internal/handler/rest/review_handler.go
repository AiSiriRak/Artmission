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