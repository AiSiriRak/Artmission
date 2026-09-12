package rest

import (
	"context"
	"io"
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

	huma.Put(api, "/artworks/{artwork_id}", h.updateArtwork,
		huma.OperationTags("artworks"),
		func(o *huma.Operation) {
			o.OperationID = "update-artwork"
			o.Summary = "UpdateArtwork"
			o.Description = "Replace portfolio artwork owned by the authenticated artist"
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

type artworkForm struct {
	Name                string          `form:"name" minLength:"1" required:"true"`
	Category            string          `form:"category" minLength:"1" required:"true"`
	Styles              []string        `form:"styles" contentType:"application/json" required:"true"`
	Description         string          `form:"description" minLength:"1" required:"true"`
	ArtworkSamples      []huma.FormFile `form:"artwork_samples" contentType:"image/jpeg,image/png,image/webp" required:"false"`
	MinimumDeadlineDays int             `form:"minimum_deadline_days" minimum:"1" required:"true"`
	PriceSatang         int64           `form:"price_satang" minimum:"0" required:"true"`
}

type CreateArtworkInput struct {
	RawBody huma.MultipartFormFiles[artworkForm]
}

type CreateArtworkOutput struct {
	Body artworkView
}

func (h *ArtworkHandler) createArtwork(ctx context.Context, input *CreateArtworkInput) (*CreateArtworkOutput, error) {
	info, ok := authInfoFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("missing authentication")
	}

	createInput, err := newArtworkCreateInput(&input.RawBody, info.UserID)
	if err != nil {
		return nil, err
	}
	created, err := h.artworkUsecase.Create(ctx, createInput)
	if err != nil {
		return nil, mapAppError(err)
	}
	return &CreateArtworkOutput{Body: newArtworkView(created)}, nil
}

type UpdateArtworkInput struct {
	ArtworkID uuid.UUID `path:"artwork_id"`
	RawBody   huma.MultipartFormFiles[artworkForm]
}

type UpdateArtworkOutput struct {
	Body artworkView
}

func (h *ArtworkHandler) updateArtwork(ctx context.Context, input *UpdateArtworkInput) (*UpdateArtworkOutput, error) {
	info, ok := authInfoFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("missing authentication")
	}

	createInput, err := newArtworkCreateInput(&input.RawBody, info.UserID)
	if err != nil {
		return nil, err
	}
	updated, err := h.artworkUsecase.Update(ctx, artwork.UpdateInput{
		ArtworkID:   input.ArtworkID,
		CreateInput: createInput,
	})
	if err != nil {
		return nil, mapAppError(err)
	}
	return &UpdateArtworkOutput{Body: newArtworkView(updated)}, nil
}

func newArtworkCreateInput(raw *huma.MultipartFormFiles[artworkForm], artistID uuid.UUID) (artwork.CreateInput, error) {
	allowedValues := map[string]struct{}{
		"name": {}, "category": {}, "styles": {}, "description": {},
		"minimum_deadline_days": {}, "price_satang": {},
	}
	for field := range raw.Form.Value {
		if _, allowed := allowedValues[field]; !allowed {
			return artwork.CreateInput{}, huma.Error422UnprocessableEntity("unexpected multipart field: " + field)
		}
	}
	for field := range raw.Form.File {
		if field != "artwork_samples" {
			return artwork.CreateInput{}, huma.Error422UnprocessableEntity("unexpected multipart field: " + field)
		}
	}

	form := raw.Data()
	files := make([]io.Reader, len(form.ArtworkSamples))
	for index := range form.ArtworkSamples {
		files[index] = form.ArtworkSamples[index].File
	}
	return artwork.CreateInput{
		ArtistID:            artistID,
		Name:                form.Name,
		Category:            form.Category,
		Styles:              form.Styles,
		Description:         form.Description,
		SampleFiles:         files,
		MinimumDeadlineDays: form.MinimumDeadlineDays,
		PriceSatang:         form.PriceSatang,
	}, nil
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
