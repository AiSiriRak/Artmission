package rest

import (
	"context"
	"io"
	"net/http"
	"strconv"
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
	huma.Get(api, "/artworks", h.searchArtworks,
		huma.OperationTags("artworks"),
		func(operation *huma.Operation) {
			operation.OperationID = "search-artworks"
			operation.Summary = "SearchArtworks"
			operation.Description = "Search artworks for the authenticated user's home page by artist name, optional category/style/price/review filters, sort, and page"
			operation.Middlewares = append(operation.Middlewares, requireAuth(api, h.authUsecase), requireRole(api, user.RoleCustomer))
		},
	)

	huma.Get(api, "/artists/{artist_id}/artworks", h.getArtistArtworks,
		huma.OperationTags("artists", "artworks"),
		func(operation *huma.Operation) {
			operation.OperationID = "get-artist-artworks"
			operation.Summary = "GetArtistArtworks"
			operation.Description = "Get every artwork created by an artist, newest first"
		},
	)

	huma.Get(api, "/artworks/{artwork_id}", h.getArtwork,
		huma.OperationTags("artworks"),
		func(o *huma.Operation) {
			o.OperationID = "get-artwork"
			o.Summary = "GetArtwork"
			o.Description = "Get the authenticated user artwork detail."
			o.Middlewares = append(
				o.Middlewares,
				requireAuth(api, h.authUsecase),
				requireAnyRole(api, user.RoleCustomer, user.RoleArtist),
			)
		})

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
			o.Description = "Update portfolio artwork owned by the authenticated artist"
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

	huma.Get(api, "/categories", h.listCategories,
		huma.OperationTags("artworks"),
		func(o *huma.Operation) {
			o.OperationID = "list-categories"
			o.Summary = "ListCategories"
			o.Description = "List artwork categories"
		},
	)

	huma.Get(api, "/styles", h.listStyles,
		huma.OperationTags("artworks"),
		func(o *huma.Operation) {
			o.OperationID = "list-styles"
			o.Summary = "ListStyles"
			o.Description = "List artwork styles"
		},
	)

}

type searchArtistView struct {
	ArtistID    uuid.UUID `json:"artist_id"`
	ArtistName  string    `json:"artist_name"`
	ProfileURL  *string   `json:"profile_url"`
	ReviewScore *float64  `json:"review_score" minimum:"1" maximum:"5"`
}

type searchArtworkView struct {
	ID                  uuid.UUID           `json:"id"`
	Name                string              `json:"name"`
	Category            string              `json:"category"`
	Styles              []string            `json:"styles"`
	Description         string              `json:"description"`
	ArtworkSamples      []artworkSampleView `json:"artwork_samples"`
	MinimumDeadlineDays int                 `json:"minimum_deadline_days"`
	PriceSatang         int64               `json:"price_satang"`
	CreatedAt           time.Time           `json:"created_at"`
	UpdatedAt           time.Time           `json:"updated_at"`
	Artist              searchArtistView    `json:"artist"`
}

type SearchArtworksInput struct {
	Q              string   `query:"q" doc:"Partial match on artist name."`
	Category       string   `query:"category" doc:"Filter by at most one artwork category."`
	Style          []string `query:"style,explode" doc:"Filter by one or more artwork styles. An artwork matches only when it has every requested style."`
	MinPriceSatang string   `query:"min_price_satang" doc:"Inclusive minimum artwork price in satang."`
	MaxPriceSatang string   `query:"max_price_satang" doc:"Inclusive maximum artwork price in satang."`
	MinReviewScore string   `query:"min_review_score" doc:"Inclusive minimum artist average review score (0-5)."`
	Sort           string   `query:"sort" enum:"name_asc,price_asc,price_desc,review_score_asc,review_score_desc" default:"name_asc" doc:"Default artwork name ascending. Price or review score, when set, is the primary sort; equal values fall back to artwork name ascending."`
	Page           int      `query:"page" minimum:"1" default:"1" doc:"1-based page of 20 artworks."`
}

type SearchArtworksOutput struct {
	Body struct {
		Artworks []searchArtworkView `json:"artworks"`
		Total    int                 `json:"total"`
		Page     int                 `json:"page"`
	}
}

func (h *ArtworkHandler) searchArtworks(ctx context.Context, input *SearchArtworksInput) (*SearchArtworksOutput, error) {
	if _, ok := authInfoFromContext(ctx); !ok {
		return nil, huma.Error401Unauthorized("missing authentication")
	}

	minPrice, err := parseOptionalInt64Query("min_price_satang", input.MinPriceSatang)
	if err != nil {
		return nil, err
	}
	maxPrice, err := parseOptionalInt64Query("max_price_satang", input.MaxPriceSatang)
	if err != nil {
		return nil, err
	}
	minScore, err := parseOptionalFloat64Query("min_review_score", input.MinReviewScore)
	if err != nil {
		return nil, err
	}

	page, err := h.artworkUsecase.Search(ctx, artwork.SearchQuery{
		ArtistName:     input.Q,
		Category:       input.Category,
		Styles:         input.Style,
		MinPriceSatang: minPrice,
		MaxPriceSatang: maxPrice,
		MinReviewScore: minScore,
		Sort:           artwork.SearchSort(input.Sort),
		Page:           input.Page,
	})
	if err != nil {
		return nil, mapAppError(err)
	}

	output := new(SearchArtworksOutput)
	output.Body.Artworks = make([]searchArtworkView, len(page.Items))
	output.Body.Total = page.Total
	output.Body.Page = page.Page
	for index, item := range page.Items {
		output.Body.Artworks[index] = newSearchArtworkView(item)
	}
	return output, nil
}

func parseOptionalInt64Query(name, raw string) (*int64, error) {
	if raw == "" {
		return nil, nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil, huma.Error400BadRequest(name + " must be an integer")
	}
	return &value, nil
}

func parseOptionalFloat64Query(name, raw string) (*float64, error) {
	if raw == "" {
		return nil, nil
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil, huma.Error400BadRequest(name + " must be a number")
	}
	return &value, nil
}

func newSearchArtworkView(item artwork.SearchItem) searchArtworkView {
	view := newArtworkView(&item.Artwork)
	return searchArtworkView{
		ID:                  view.ID,
		Name:                view.Name,
		Category:            view.Category,
		Styles:              view.Styles,
		Description:         view.Description,
		ArtworkSamples:      view.ArtworkSamples,
		MinimumDeadlineDays: view.MinimumDeadlineDays,
		PriceSatang:         view.PriceSatang,
		CreatedAt:           view.CreatedAt,
		UpdatedAt:           view.UpdatedAt,
		Artist: searchArtistView{
			ArtistID:    item.Artist.ID,
			ArtistName:  item.Artist.Name,
			ProfileURL:  item.Artist.ProfileURL,
			ReviewScore: item.Artist.ReviewScore,
		},
	}
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
	CategoryID          uuid.UUID       `form:"category_id" required:"true"`
	StyleIDs            []uuid.UUID     `form:"style_ids" contentType:"application/json" required:"true"`
	Description         string          `form:"description" minLength:"1" required:"true"`
	ArtworkSamples      []huma.FormFile `form:"artwork_samples" contentType:"image/jpeg,image/png,image/webp" required:"false"`
	MinimumDeadlineDays int             `form:"minimum_deadline_days" minimum:"1" required:"true"`
	PriceSatang         int64           `form:"price_satang" minimum:"0" required:"true"`
}

type updateArtworkForm struct {
	Name                string          `form:"name" minLength:"1" required:"true"`
	CategoryID          uuid.UUID       `form:"category_id" required:"true"`
	StyleIDs            []uuid.UUID     `form:"style_ids" contentType:"application/json" required:"true"`
	Description         string          `form:"description" minLength:"1" required:"true"`
	UploadedSamples     []huma.FormFile `form:"uploaded_samples" contentType:"image/jpeg,image/png,image/webp" required:"false"`
	DeletedSampleURLs   []string        `form:"deleted_sample_urls" contentType:"application/json" required:"false"`
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

	createInput := newArtworkCreateInput(input.RawBody.Data(), info.UserID)
	created, err := h.artworkUsecase.Create(ctx, createInput)
	if err != nil {
		return nil, mapAppError(err)
	}
	return &CreateArtworkOutput{Body: newArtworkView(created)}, nil
}

type UpdateArtworkInput struct {
	ArtworkID uuid.UUID `path:"artwork_id"`
	RawBody   huma.MultipartFormFiles[updateArtworkForm]
}

type UpdateArtworkOutput struct {
	Body artworkView
}

func (h *ArtworkHandler) updateArtwork(ctx context.Context, input *UpdateArtworkInput) (*UpdateArtworkOutput, error) {
	info, ok := authInfoFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("missing authentication")
	}

	updated, err := h.artworkUsecase.Update(ctx, newArtworkUpdateInput(input.RawBody.Data(), input.ArtworkID, info.UserID))
	if err != nil {
		return nil, mapAppError(err)
	}
	return &UpdateArtworkOutput{Body: newArtworkView(updated)}, nil
}

func newArtworkCreateInput(form *artworkForm, artistID uuid.UUID) artwork.CreateInput {
	return artwork.CreateInput{
		ArtistID:            artistID,
		Name:                form.Name,
		CategoryID:          form.CategoryID,
		StyleIDs:            form.StyleIDs,
		Description:         form.Description,
		SampleFiles:         artworkFileReaders(form.ArtworkSamples),
		MinimumDeadlineDays: form.MinimumDeadlineDays,
		PriceSatang:         form.PriceSatang,
	}
}

func newArtworkUpdateInput(form *updateArtworkForm, artworkID, artistID uuid.UUID) artwork.UpdateInput {
	return artwork.UpdateInput{
		ArtworkID:         artworkID,
		DeletedSampleURLs: form.DeletedSampleURLs,
		CreateInput: artwork.CreateInput{
			ArtistID:            artistID,
			Name:                form.Name,
			CategoryID:          form.CategoryID,
			StyleIDs:            form.StyleIDs,
			Description:         form.Description,
			SampleFiles:         artworkFileReaders(form.UploadedSamples),
			MinimumDeadlineDays: form.MinimumDeadlineDays,
			PriceSatang:         form.PriceSatang,
		},
	}
}

func artworkFileReaders(files []huma.FormFile) []io.Reader {
	readers := make([]io.Reader, len(files))
	for index := range files {
		readers[index] = files[index].File
	}
	return readers
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

type ListCategoriesInput struct{}

type ListCategoriesOutput struct {
	Body []artistReferenceView
}

type ListStylesInput struct{}

type ListStylesOutput struct {
	Body []artistReferenceView
}

func (h *ArtworkHandler) listCategories(ctx context.Context, _ *ListCategoriesInput) (*ListCategoriesOutput, error) {
	categories, err := h.artworkUsecase.ListAllCategories(ctx)
	if err != nil {
		return nil, mapAppError(err)
	}

	body := make([]artistReferenceView, len(categories))
	for index, category := range categories {
		body[index] = artistReferenceView{ID: category.ID, Label: category.Label}
	}
	return &ListCategoriesOutput{Body: body}, nil
}

func (h *ArtworkHandler) listStyles(ctx context.Context, _ *ListStylesInput) (*ListStylesOutput, error) {
	styles, err := h.artworkUsecase.ListAllStyles(ctx)
	if err != nil {
		return nil, mapAppError(err)
	}

	body := make([]artistReferenceView, len(styles))
	for index, style := range styles {
		body[index] = artistReferenceView{ID: style.ID, Label: style.Label}
	}
	return &ListStylesOutput{Body: body}, nil
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

type GetArtworkInput struct {
	ArtworkID uuid.UUID `path:"artwork_id"`
}

type GetArtworkOutput struct {
	Body artworkView
}

func (h *ArtworkHandler) getArtwork(ctx context.Context, input *GetArtworkInput) (*GetArtworkOutput, error) {
	detail, err := h.artworkUsecase.GetArtwork(ctx, input.ArtworkID)
	if err != nil {
		return nil, mapAppError(err)
	}

	categories, err := h.artworkUsecase.ListAllCategories(ctx)
	if err != nil {
		return nil, mapAppError(err)
	}
	styles, err := h.artworkUsecase.ListAllStyles(ctx)
	if err != nil {
		return nil, mapAppError(err)
	}

	categoryName := ""
	for _, cat := range categories {
		if cat.ID == detail.CategoryID {
			categoryName = cat.Label
			break
		}
	}

	styleMap := make(map[uuid.UUID]string, len(styles))
	for _, style := range styles {
		styleMap[style.ID] = style.Label
	}

	styleNames := make([]string, 0, len(detail.StyleIDs))
	for _, styleID := range detail.StyleIDs {
		if label, ok := styleMap[styleID]; ok {
			styleNames = append(styleNames, label)
		}
	}
	artwork := &artwork.Artwork{
		ID:                  detail.ID,
		ArtistID:            detail.ArtistID,
		Name:                detail.Name,
		Category:            categoryName,
		Styles:              styleNames,
		Description:         detail.Description,
		Samples:             detail.Samples,
		MinimumDeadlineDays: detail.MinimumDeadlineDays,
		PriceSatang:         detail.PriceSatang,
		CreatedAt:           detail.CreatedAt,
		UpdatedAt:           detail.UpdatedAt}
	return &GetArtworkOutput{
		Body: newArtworkView(artwork),
	}, nil
}
