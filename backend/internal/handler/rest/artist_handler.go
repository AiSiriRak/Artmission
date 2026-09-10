package rest

import (
	"context"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/artist"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/auth"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/user"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type ArtistHandler struct {
	profileUsecase artist.ProfileUsecase
	authUsecase    auth.AuthUsecase
}

func NewArtistHandler(profileUsecase artist.ProfileUsecase, authUsecase auth.AuthUsecase) *ArtistHandler {
	return &ArtistHandler{profileUsecase: profileUsecase, authUsecase: authUsecase}
}

func (h *ArtistHandler) Register(api huma.API) {
	huma.Get(api, "/artists/{artist_id}", h.getArtistProfile,
		huma.OperationTags("artists"),
		func(o *huma.Operation) {
			o.OperationID = "get-artist-profile"
			o.Summary = "GetArtistProfile"
			o.Description = "Get a public artist profile"
		},
	)

	huma.Put(api, "/artists/me", h.updateMyArtistProfile,
		huma.OperationTags("artists"),
		func(o *huma.Operation) {
			o.OperationID = "update-my-artist-profile"
			o.Summary = "UpdateMyArtistProfile"
			o.Description = "Replace the authenticated artist's editable profile details"
			o.Middlewares = append(o.Middlewares, requireAuth(api, h.authUsecase), requireRole(api, user.RoleArtist))
		},
	)
}

type artistReferenceView struct {
	ID    uuid.UUID `json:"id"`
	Label string    `json:"label"`
}

type artistProfileView struct {
	ArtistID       uuid.UUID             `json:"artist_id"`
	ArtistName     string                `json:"artist_name"`
	Description    *string               `json:"description"`
	Categories     []artistReferenceView `json:"categories"`
	Styles         []artistReferenceView `json:"styles"`
	MinPriceSatang *int64                `json:"min_price_satang"`
	MaxPriceSatang *int64                `json:"max_price_satang"`
	ReviewScore    *float64              `json:"review_score"`
}

type GetArtistProfileInput struct {
	ArtistID uuid.UUID `path:"artist_id"`
}

type GetArtistProfileOutput struct {
	Body artistProfileView
}

func (h *ArtistHandler) getArtistProfile(ctx context.Context, in *GetArtistProfileInput) (*GetArtistProfileOutput, error) {
	profile, err := h.profileUsecase.GetProfile(ctx, in.ArtistID)
	if err != nil {
		return nil, mapAppError(err)
	}
	return &GetArtistProfileOutput{Body: newArtistProfileView(profile)}, nil
}

type updateMyArtistProfileInput struct {
	Body struct {
		Description    *string     `json:"description"`
		StyleIDs       []uuid.UUID `json:"style_ids"`
		MinPriceSatang int64       `json:"min_price_satang" minimum:"0"`
		MaxPriceSatang int64       `json:"max_price_satang" minimum:"0"`
	}
}

type UpdateMyArtistProfileOutput struct {
	Body artistProfileView
}

func (h *ArtistHandler) updateMyArtistProfile(ctx context.Context, in *updateMyArtistProfileInput) (*UpdateMyArtistProfileOutput, error) {
	info, ok := authInfoFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("missing authentication")
	}

	profile, err := h.profileUsecase.UpdateProfile(ctx, info.UserID, artist.UpdateProfileInput{
		Description:    in.Body.Description,
		StyleIDs:       in.Body.StyleIDs,
		MinPriceSatang: in.Body.MinPriceSatang,
		MaxPriceSatang: in.Body.MaxPriceSatang,
	})
	if err != nil {
		return nil, mapAppError(err)
	}
	return &UpdateMyArtistProfileOutput{Body: newArtistProfileView(profile)}, nil
}

func newArtistProfileView(profile *artist.Profile) artistProfileView {
	categories := make([]artistReferenceView, len(profile.Categories))
	for i, category := range profile.Categories {
		categories[i] = artistReferenceView{ID: category.ID, Label: category.Label}
	}
	styles := make([]artistReferenceView, len(profile.Styles))
	for i, style := range profile.Styles {
		styles[i] = artistReferenceView{ID: style.ID, Label: style.Label}
	}

	return artistProfileView{
		ArtistID:       profile.UserID,
		ArtistName:     profile.ArtistName,
		Description:    profile.Description,
		Categories:     categories,
		Styles:         styles,
		MinPriceSatang: profile.MinPriceSatang,
		MaxPriceSatang: profile.MaxPriceSatang,
		ReviewScore:    profile.ReviewScore,
	}
}
