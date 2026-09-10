package rest

import (
	"context"
	"net/http"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/artwork"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/auth"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/user"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type artworkSampleView struct {
	ImageURL string `json:"image_url" format:"uri"`
}

type artworkView struct {
	ID                  uuid.UUID           `json:"id"`
	ArtistID            uuid.UUID           `json:"artist_id"`
	Name                string              `json:"name"`
	Category            string              `json:"category"`
	Styles              []string            `json:"styles"`
	Description         string              `json:"description"`
	ArtworkSamples      []artworkSampleView `json:"artwork_samples"`
	MinimumDeadlineDays int                 `json:"minimum_deadline_days"`
	PriceSatang         int64               `json:"price_satang"`
	CreatedAt           time.Time           `json:"created_at"`
	UpdatedAt           time.Time           `json:"updated_at"`
}

type artistArtworkView struct {
	ArtworkID           uuid.UUID           `json:"artwork_id"`
	Name                string              `json:"name"`
	Category            string              `json:"category"`
	Styles              []string            `json:"styles"`
	Description         string              `json:"description"`
	ArtworkSamples      []artworkSampleView `json:"artwork_samples"`
	MinimumDeadlineDays int                 `json:"minimum_deadline_days"`
	PriceSatang         int64               `json:"price_satang"`
}

type ArtworkHandler struct {
	artworkUsecase artwork.Usecase
	authUsecase    auth.AuthUsecase
}

func NewArtworkHandler(artworkUsecase artwork.Usecase, authUsecase auth.AuthUsecase) *ArtworkHandler {
	return &ArtworkHandler{artworkUsecase: artworkUsecase, authUsecase: authUsecase}
}

func (h *ArtworkHandler) Register(api huma.API) {
	huma.Get(api, "/artists/{artist_id}/artworks", h.getArtistArtworks,
		huma.OperationTags("artists", "artworks"),
		func(operation *huma.Operation) {
			operation.OperationID = "get-artist-artworks"
			operation.Summary = "GetArtistArtworks"
			operation.Description = "Get every artwork created by an artist, newest first"
		},
	)

	huma.Post(api, "/artworks", h.createArtwork,
		huma.OperationTags("artworks"),
		func(o *huma.Operation) {
			o.OperationID = "create-artwork"
			o.Summary = "CreateArtwork"
			o.Description = "Create portfolio artwork owned by the authenticated artist"
			o.DefaultStatus = http.StatusCreated
			o.Middlewares = append(o.Middlewares, requireAuth(api, h.authUsecase), requireRole(api, user.RoleArtist))
		},
	)

	huma.Delete(api, "/artworks/{artwork_id}", h.deleteArtwork,
		huma.OperationTags("artworks"),
		func(o *huma.Operation) {
			o.OperationID = "delete-artwork"
			o.Summary = "DeleteArtwork"
			o.Description = "Delete portfolio artwork owned by the authenticated artist"
			o.DefaultStatus = http.StatusNoContent
			o.Middlewares = append(o.Middlewares, requireAuth(api, h.authUsecase), requireRole(api, user.RoleArtist))
		},
	)
}

type GetArtistArtworksInput struct {
	ArtistID uuid.UUID `path:"artist_id"`
}

type GetArtistArtworksOutput struct {
	Body struct {
		Artworks []artistArtworkView `json:"artworks"`
	}
}

func (h *ArtworkHandler) getArtistArtworks(ctx context.Context, input *GetArtistArtworksInput) (*GetArtistArtworksOutput, error) {
	artworks, err := h.artworkUsecase.ListByArtistID(ctx, input.ArtistID)
	if err != nil {
		return nil, mapAppError(err)
	}

	output := new(GetArtistArtworksOutput)
	output.Body.Artworks = make([]artistArtworkView, len(artworks))
	for index, item := range artworks {
		samples := make([]artworkSampleView, len(item.Samples))
		for sampleIndex, sample := range item.Samples {
			samples[sampleIndex] = artworkSampleView{ImageURL: sample.ImageURL}
		}
		output.Body.Artworks[index] = artistArtworkView{
			ArtworkID:           item.ID,
			Name:                item.Name,
			Category:            item.Category,
			Styles:              item.Styles,
			Description:         item.Description,
			ArtworkSamples:      samples,
			MinimumDeadlineDays: item.MinimumDeadlineDays,
			PriceSatang:         item.PriceSatang,
		}
	}
	return output, nil
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

func (h *ArtworkHandler) createArtwork(ctx context.Context, input *CreateArtworkInput) (*CreateArtworkOutput, error) {
	info, ok := authInfoFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("missing authentication")
	}

	samples := make([]artwork.Sample, len(input.Body.ArtworkSamples))
	for index, sample := range input.Body.ArtworkSamples {
		samples[index] = artwork.Sample{ImageURL: sample.ImageURL}
	}
	created, err := h.artworkUsecase.Create(ctx, artwork.CreateInput{
		ArtistID:            info.UserID,
		Name:                input.Body.Name,
		Category:            input.Body.Category,
		Styles:              input.Body.Styles,
		Description:         input.Body.Description,
		Samples:             samples,
		MinimumDeadlineDays: input.Body.MinimumDeadlineDays,
		PriceSatang:         input.Body.PriceSatang,
	})
	if err != nil {
		return nil, mapAppError(err)
	}
	return &CreateArtworkOutput{Body: newArtworkView(created)}, nil
}

type DeleteArtworkInput struct {
	ArtworkID uuid.UUID `path:"artwork_id"`
}

type DeleteArtworkOutput struct{}

func (h *ArtworkHandler) deleteArtwork(ctx context.Context, input *DeleteArtworkInput) (*DeleteArtworkOutput, error) {
	info, ok := authInfoFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("missing authentication")
	}
	if err := h.artworkUsecase.Delete(ctx, info.UserID, input.ArtworkID); err != nil {
		return nil, mapAppError(err)
	}
	return &DeleteArtworkOutput{}, nil
}

func newArtworkView(item *artwork.Artwork) artworkView {
	samples := make([]artworkSampleView, len(item.Samples))
	for index, sample := range item.Samples {
		samples[index] = artworkSampleView{ImageURL: sample.ImageURL}
	}
	return artworkView{
		ID:                  item.ID,
		ArtistID:            item.ArtistID,
		Name:                item.Name,
		Category:            item.Category,
		Styles:              item.Styles,
		Description:         item.Description,
		ArtworkSamples:      samples,
		MinimumDeadlineDays: item.MinimumDeadlineDays,
		PriceSatang:         item.PriceSatang,
		CreatedAt:           item.CreatedAt,
		UpdatedAt:           item.UpdatedAt,
	}
}
