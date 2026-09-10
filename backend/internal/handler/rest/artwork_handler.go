package rest

import (
	"context"
	"net/http"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/auth"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/user"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type artworkSampleView struct {
	ImageURL string `json:"image_url" format:"uri"`
}

type artworkView struct {
	Name                string              `json:"name"`
	Category            string              `json:"category"`
	Styles              []string            `json:"styles"`
	Description         string              `json:"description"`
	ArtworkSamples      []artworkSampleView `json:"artwork_samples"`
	MinimumDeadlineDays int                 `json:"minimum_deadline_days"`
	PriceSatang         int64               `json:"price_satang"`
}

// ArtworkHandler is a transport-only stub until artwork persistence is implemented.
type ArtworkHandler struct {
	authUsecase auth.AuthUsecase
}

func NewArtworkHandler(authUsecase auth.AuthUsecase) *ArtworkHandler {
	return &ArtworkHandler{authUsecase: authUsecase}
}

func (h *ArtworkHandler) Register(api huma.API) {
	huma.Post(api, "/artworks", h.createArtwork,
		huma.OperationTags("artworks"),
		func(o *huma.Operation) {
			o.OperationID = "create-artwork"
			o.Summary = "CreateArtwork"
			o.Description = "Return a fixed artwork response for frontend integration"
			o.DefaultStatus = http.StatusCreated
			o.Middlewares = append(o.Middlewares, requireAuth(api, h.authUsecase), requireRole(api, user.RoleArtist))
		},
	)

	huma.Delete(api, "/artworks/{artwork_id}", h.deleteArtwork,
		huma.OperationTags("artworks"),
		func(o *huma.Operation) {
			o.OperationID = "delete-artwork"
			o.Summary = "DeleteArtwork"
			o.Description = "Return a successful deletion response for frontend integration"
			o.DefaultStatus = http.StatusNoContent
			o.Middlewares = append(o.Middlewares, requireAuth(api, h.authUsecase), requireRole(api, user.RoleArtist))
		},
	)
}

type CreateArtworkInput struct {
	Body struct {
		Name                string              `json:"name" minLength:"1"`
		Category            string              `json:"category" minLength:"1"`
		Styles              []string            `json:"styles"`
		Description         string              `json:"description" minLength:"1"`
		ArtworkSamples      []artworkSampleView `json:"artwork_samples"`
		MinimumDeadlineDays int                 `json:"minimum_deadline_days" minimum:"1"`
		PriceSatang         int64               `json:"price_satang" minimum:"0"`
	}
}

type CreateArtworkOutput struct {
	Body artworkView
}

func (h *ArtworkHandler) createArtwork(context.Context, *CreateArtworkInput) (*CreateArtworkOutput, error) {
	return &CreateArtworkOutput{Body: dummyArtwork()}, nil
}

type DeleteArtworkInput struct {
	ArtworkID uuid.UUID `path:"artwork_id"`
}

type DeleteArtworkOutput struct{}

func (h *ArtworkHandler) deleteArtwork(context.Context, *DeleteArtworkInput) (*DeleteArtworkOutput, error) {
	return &DeleteArtworkOutput{}, nil
}

func dummyArtwork() artworkView {
	return artworkView{
		Name:                "Book Cover",
		Category:            "Book",
		Styles:              []string{"Pixel Art", "Cartoon"},
		Description:         "A colorful book-cover commission",
		ArtworkSamples:      []artworkSampleView{{ImageURL: "https://example.com/sample.png"}},
		MinimumDeadlineDays: 7,
		PriceSatang:         250000,
	}
}
