package rest

import (
	"context"
	"io"

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
		func(operation *huma.Operation) {
			operation.OperationID = "get-artist-profile"
			operation.Summary = "GetArtistProfile"
			operation.Description = "Get a public artist profile with artwork-derived details and paginated reviews"
		},
	)

	huma.Put(api, "/artists/me", h.updateMyArtistProfile,
		huma.OperationTags("artists"),
		func(operation *huma.Operation) {
			operation.OperationID = "update-my-artist-profile"
			operation.Summary = "UpdateMyArtistProfile"
			operation.Description = "Update the authenticated artist's description or public profile image"
			operation.Middlewares = append(operation.Middlewares, requireAuth(api, h.authUsecase), requireRole(api, user.RoleArtist))
		},
	)
}

type artistReferenceView struct {
	ID    uuid.UUID `json:"id"`
	Label string    `json:"label"`
}

type artistReviewView struct {
	Username string `json:"username"`
	Order    string `json:"order"`
	Rating   int    `json:"rating" minimum:"1" maximum:"5"`
}

type artistProfileView struct {
	ArtistID       uuid.UUID             `json:"artist_id"`
	ArtistName     string                `json:"artist_name"`
	ProfileURL     *string               `json:"profile_url"`
	Description    *string               `json:"description"`
	Categories     []artistReferenceView `json:"categories"`
	Styles         []artistReferenceView `json:"styles"`
	MinPriceSatang *int64                `json:"min_price_satang"`
	MaxPriceSatang *int64                `json:"max_price_satang"`
	ReviewScore    *float64              `json:"review_score" minimum:"1" maximum:"5"`
	Reviews        []artistReviewView    `json:"reviews"`
	Total          int                   `json:"total" minimum:"0"`
}

type GetArtistProfileInput struct {
	ArtistID uuid.UUID `path:"artist_id"`
	Limit    int       `query:"limit" minimum:"1" maximum:"100" default:"20" doc:"Maximum number of reviews to return."`
	Offset   int       `query:"offset" minimum:"0" default:"0" doc:"Number of reviews to skip."`
}

type GetArtistProfileOutput struct {
	Body artistProfileView
}

func (h *ArtistHandler) getArtistProfile(ctx context.Context, in *GetArtistProfileInput) (*GetArtistProfileOutput, error) {
	profile, err := h.profileUsecase.GetProfile(ctx, in.ArtistID, artist.ProfileQuery{Limit: in.Limit, Offset: in.Offset})
	if err != nil {
		return nil, mapAppError(err)
	}
	return &GetArtistProfileOutput{Body: newArtistProfileView(profile)}, nil
}

type updateMyArtistProfileForm struct {
	Description        string        `form:"description" required:"false"`
	ProfileImage       huma.FormFile `form:"profile_image" contentType:"image/jpeg,image/png,image/webp" required:"false"`
	RemoveProfileImage bool          `form:"remove_profile_image" required:"false"`
}

type updateMyArtistProfileInput struct {
	RawBody huma.MultipartFormFiles[updateMyArtistProfileForm]
}

type UpdateMyArtistProfileOutput struct {
	Body artistProfileView
}

func (h *ArtistHandler) updateMyArtistProfile(ctx context.Context, in *updateMyArtistProfileInput) (*UpdateMyArtistProfileOutput, error) {
	info, ok := authInfoFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("missing authentication")
	}

	form := in.RawBody.Data()
	for field := range in.RawBody.Form.Value {
		if field != "description" && field != "remove_profile_image" {
			return nil, huma.Error422UnprocessableEntity("unexpected multipart field: " + field)
		}
	}
	for field := range in.RawBody.Form.File {
		if field != "profile_image" {
			return nil, huma.Error422UnprocessableEntity("unexpected multipart field: " + field)
		}
	}
	var description *string
	if _, present := in.RawBody.Form.Value["description"]; present {
		description = &form.Description
	}
	var profileImage io.Reader
	if form.ProfileImage.IsSet {
		profileImage = form.ProfileImage.File
	}
	profile, err := h.profileUsecase.UpdateProfile(ctx, info.UserID, artist.UpdateProfileInput{
		Description:        description,
		ProfileImage:       profileImage,
		RemoveProfileImage: form.RemoveProfileImage,
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
	reviews := make([]artistReviewView, len(profile.Reviews))
	for i, review := range profile.Reviews {
		reviews[i] = artistReviewView{Username: review.Username, Order: review.Order, Rating: review.Rating}
	}

	return artistProfileView{
		ArtistID:       profile.UserID,
		ArtistName:     profile.ArtistName,
		ProfileURL:     profile.ProfileURL,
		Description:    profile.Description,
		Categories:     categories,
		Styles:         styles,
		MinPriceSatang: profile.MinPriceSatang,
		MaxPriceSatang: profile.MaxPriceSatang,
		ReviewScore:    profile.ReviewScore,
		Reviews:        reviews,
		Total:          profile.Total,
	}
}
