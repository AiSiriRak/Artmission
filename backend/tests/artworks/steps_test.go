//go:build integration

package artworks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"time"

	"github.com/AiSiriRak/Artmission/backend/tests/internal/apptest"
	"github.com/cucumber/godog"
	"github.com/google/uuid"
)

type artworkSampleBody struct {
	ImageURL string `json:"image_url"`
}

type createArtworkBody struct {
	Name                string              `json:"name"`
	Category            string              `json:"category"`
	Styles              []string            `json:"styles"`
	Description         string              `json:"description"`
	ArtworkSamples      []artworkSampleBody `json:"artwork_samples"`
	MinimumDeadlineDays int                 `json:"minimum_deadline_days"`
	PriceSatang         int64               `json:"price_satang"`
}

type artworkResponse struct {
	ID                  string              `json:"id"`
	ArtistID            string              `json:"artist_id"`
	Name                string              `json:"name"`
	Category            string              `json:"category"`
	Styles              []string            `json:"styles"`
	Description         string              `json:"description"`
	ArtworkSamples      []artworkSampleBody `json:"artwork_samples"`
	MinimumDeadlineDays int                 `json:"minimum_deadline_days"`
	PriceSatang         int64               `json:"price_satang"`
	CreatedAt           time.Time           `json:"created_at"`
	UpdatedAt           time.Time           `json:"updated_at"`
}

type artworkImageRow struct {
	ImageURL  string `bun:"image_url"`
	SortOrder int    `bun:"sort_order"`
}

type artworkContext struct {
	client               *apptest.Client
	artist               apptest.Account
	accessToken          string
	artworkID            string
	otherArtwork         string
	response             *apptest.Response
	requestedBody        createArtworkBody
	currentSampleURLs    []string
	expectedRetainedURLs []string
	uploadedSampleCount  int
	createdAt            time.Time
	updatedAt            time.Time
}

func (a *artworkContext) registerArtist() error {
	account, err := apptest.RegisterArtist(app, a.client, "Portfolio artist")
	if err != nil {
		return err
	}
	a.artist = account
	return nil
}

func (a *artworkContext) loginArtist() error {
	token, err := apptest.Login(a.client, a.artist.Email, a.artist.Password)
	if err != nil {
		return err
	}
	a.accessToken = token
	return nil
}

func (a *artworkContext) createArtwork(token string) error {
	body := newArtworkBody()
	headers := map[string]string{}
	if token != "" {
		headers["Authorization"] = "Bearer " + token
	}
	response, err := a.client.DoMultipart(http.MethodPost, "/artworks", artworkFields(body), artworkFiles("artwork_samples", len(body.ArtworkSamples)), headers)
	if err != nil {
		return err
	}
	a.response = response
	a.requestedBody = body
	a.expectedRetainedURLs = nil
	a.uploadedSampleCount = len(body.ArtworkSamples)
	if response.StatusCode == http.StatusCreated {
		var created artworkResponse
		if err := response.JSON(&created); err != nil {
			return fmt.Errorf("decode created artwork: %w", err)
		}
		a.artworkID = created.ID
		a.createdAt = created.CreatedAt
		a.updatedAt = created.UpdatedAt
		a.currentSampleURLs = make([]string, len(created.ArtworkSamples))
		for index, sample := range created.ArtworkSamples {
			a.currentSampleURLs[index] = sample.ImageURL
		}
	}
	return nil
}

func (a *artworkContext) createArtworkForOtherArtist() error {
	otherClient := apptest.NewClient(app.BaseURL())
	other, err := apptest.RegisterArtist(app, otherClient, "Other portfolio artist")
	if err != nil {
		return err
	}
	token, err := apptest.Login(otherClient, other.Email, other.Password)
	if err != nil {
		return err
	}
	body := newArtworkBody()
	response, err := otherClient.DoMultipart(http.MethodPost, "/artworks", artworkFields(body), artworkFiles("artwork_samples", len(body.ArtworkSamples)), map[string]string{"Authorization": "Bearer " + token})
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusCreated {
		return fmt.Errorf("create other artwork: expected 201, got %d: %s", response.StatusCode, response.Body)
	}
	var created artworkResponse
	if err := response.JSON(&created); err != nil {
		return err
	}
	a.otherArtwork = created.ID
	return nil
}

func (a *artworkContext) deleteArtwork(id, token string) error {
	headers := map[string]string{}
	if token != "" {
		headers["Authorization"] = "Bearer " + token
	}
	response, err := a.client.Do(http.MethodDelete, "/artworks/"+id, nil, headers)
	if err != nil {
		return err
	}
	a.response = response
	return nil
}

func (a *artworkContext) updateArtwork(id, token string, body createArtworkBody, deletedSampleURLs []string) error {
	headers := map[string]string{}
	if token != "" {
		headers["Authorization"] = "Bearer " + token
	}
	fields := artworkFields(body)
	if deletedSampleURLs != nil {
		encoded, _ := json.Marshal(deletedSampleURLs)
		fields["deleted_sample_urls"] = string(encoded)
	}
	response, err := a.client.DoMultipart(http.MethodPut, "/artworks/"+id, fields, artworkFiles("uploaded_samples", len(body.ArtworkSamples)), headers)
	if err != nil {
		return err
	}
	a.response = response
	a.requestedBody = body
	deleted := make(map[string]struct{}, len(deletedSampleURLs))
	for _, imageURL := range deletedSampleURLs {
		deleted[imageURL] = struct{}{}
	}
	a.expectedRetainedURLs = a.expectedRetainedURLs[:0]
	for _, imageURL := range a.currentSampleURLs {
		if _, remove := deleted[imageURL]; !remove {
			a.expectedRetainedURLs = append(a.expectedRetainedURLs, imageURL)
		}
	}
	a.uploadedSampleCount = len(body.ArtworkSamples)
	return nil
}

func (a *artworkContext) assertStoredArtwork() error {
	if a.response.StatusCode != http.StatusCreated {
		return fmt.Errorf("expected 201, got %d: %s", a.response.StatusCode, a.response.Body)
	}
	var item artworkResponse
	if err := a.response.JSON(&item); err != nil {
		return err
	}
	return a.assertPersistedArtwork(item)
}

func (a *artworkContext) assertUpdatedArtwork() error {
	if a.response.StatusCode != http.StatusOK {
		return fmt.Errorf("expected 200, got %d: %s", a.response.StatusCode, a.response.Body)
	}
	var item artworkResponse
	if err := a.response.JSON(&item); err != nil {
		return err
	}
	if !item.CreatedAt.Equal(a.createdAt.Truncate(time.Microsecond)) || !item.UpdatedAt.After(a.updatedAt) {
		return fmt.Errorf("timestamps created_at=%v updated_at=%v, want created_at=%v and updated_at after %v", item.CreatedAt, item.UpdatedAt, a.createdAt, a.updatedAt)
	}
	return a.assertPersistedArtwork(item)
}

func (a *artworkContext) assertPersistedArtwork(item artworkResponse) error {
	if item.ID != a.artworkID || item.ArtistID != a.artist.ID || item.Name != a.requestedBody.Name || item.Category != a.requestedBody.Category || item.Description != a.requestedBody.Description || item.MinimumDeadlineDays != a.requestedBody.MinimumDeadlineDays || item.PriceSatang != a.requestedBody.PriceSatang || !reflect.DeepEqual(item.Styles, a.requestedBody.Styles) {
		return fmt.Errorf("unexpected artwork: %+v", item)
	}
	wantSampleCount := len(a.expectedRetainedURLs) + a.uploadedSampleCount
	if len(item.ArtworkSamples) != wantSampleCount {
		return fmt.Errorf("response sample count = %d, want %d", len(item.ArtworkSamples), wantSampleCount)
	}
	for index, sample := range item.ArtworkSamples {
		parsed, err := url.Parse(sample.ImageURL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return fmt.Errorf("response sample %d has invalid public URL %q", index, sample.ImageURL)
		}
	}
	for index, imageURL := range a.expectedRetainedURLs {
		if item.ArtworkSamples[index].ImageURL != imageURL {
			return fmt.Errorf("retained sample %d = %q, want %q", index, item.ArtworkSamples[index].ImageURL, imageURL)
		}
	}

	artworkID, err := uuid.Parse(item.ID)
	if err != nil {
		return err
	}
	ctx := context.Background()
	var category string
	if err := app.DB.NewSelect().
		TableExpr("categories AS c").
		ColumnExpr("c.label").
		Join("JOIN artworks AS a ON a.category_id = c.id").
		Where("a.id = ?", artworkID).
		Scan(ctx, &category); err != nil {
		return fmt.Errorf("read artwork category: %w", err)
	}
	if category != a.requestedBody.Category {
		return fmt.Errorf("stored category = %q, want %q", category, a.requestedBody.Category)
	}
	categoryCount, err := app.DB.NewSelect().Table("categories").Where("label = ?", a.requestedBody.Category).Count(ctx)
	if err != nil {
		return err
	}
	if categoryCount != 1 {
		return fmt.Errorf("category %q row count = %d, want 1", a.requestedBody.Category, categoryCount)
	}

	styles := make([]string, 0)
	if err := app.DB.NewSelect().
		TableExpr("styles AS s").
		ColumnExpr("s.label").
		Join("JOIN artwork_styles AS aws ON aws.style_id = s.id").
		Where("aws.artwork_id = ?", artworkID).
		OrderExpr("s.label ASC").
		Scan(ctx, &styles); err != nil {
		return fmt.Errorf("read artwork styles: %w", err)
	}
	wantStyles := append([]string{}, a.requestedBody.Styles...)
	sort.Strings(wantStyles)
	if !reflect.DeepEqual(styles, wantStyles) {
		return fmt.Errorf("stored styles = %v, want %v", styles, wantStyles)
	}
	for _, label := range wantStyles {
		styleCount, err := app.DB.NewSelect().Table("styles").Where("label = ?", label).Count(ctx)
		if err != nil {
			return err
		}
		if styleCount != 1 {
			return fmt.Errorf("style %q row count = %d, want 1", label, styleCount)
		}
	}

	images := make([]artworkImageRow, 0)
	if err := app.DB.NewSelect().
		Table("artwork_images").
		Column("image_url", "sort_order").
		Where("artwork_id = ?", artworkID).
		OrderExpr("sort_order ASC").
		Scan(ctx, &images); err != nil {
		return fmt.Errorf("read artwork samples: %w", err)
	}
	if len(images) != wantSampleCount {
		return fmt.Errorf("stored sample count = %d, want %d", len(images), wantSampleCount)
	}
	for i, image := range images {
		if image.ImageURL != item.ArtworkSamples[i].ImageURL || image.SortOrder != i {
			return fmt.Errorf("stored sample %d = %+v", i, image)
		}
	}
	a.currentSampleURLs = make([]string, len(item.ArtworkSamples))
	for index, sample := range item.ArtworkSamples {
		a.currentSampleURLs[index] = sample.ImageURL
	}
	return nil
}

func (a *artworkContext) assertDeletedArtwork() error {
	if a.response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("expected 204, got %d: %s", a.response.StatusCode, a.response.Body)
	}
	artworkID, err := uuid.Parse(a.artworkID)
	if err != nil {
		return err
	}
	for _, table := range []string{"artworks", "artwork_images", "artwork_styles"} {
		var count int
		query := app.DB.NewSelect().Table(table)
		if table == "artworks" {
			query = query.Where("id = ?", artworkID)
		} else {
			query = query.Where("artwork_id = ?", artworkID)
		}
		count, err = query.Count(context.Background())
		if err != nil {
			return err
		}
		if count != 0 {
			return fmt.Errorf("%s still has %d rows for deleted artwork", table, count)
		}
	}
	return nil
}

func (a *artworkContext) assertOtherArtworkWasNotDeleted() error {
	if a.response.StatusCode != http.StatusNotFound {
		return fmt.Errorf("expected 404, got %d: %s", a.response.StatusCode, a.response.Body)
	}
	artworkID, err := uuid.Parse(a.otherArtwork)
	if err != nil {
		return err
	}
	exists, err := app.DB.NewSelect().Table("artworks").Where("id = ?", artworkID).Exists(context.Background())
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("other artist's artwork was deleted")
	}
	return nil
}

func (a *artworkContext) assertOtherArtworkWasNotUpdated() error {
	if a.response.StatusCode != http.StatusNotFound {
		return fmt.Errorf("expected 404, got %d: %s", a.response.StatusCode, a.response.Body)
	}
	artworkID, err := uuid.Parse(a.otherArtwork)
	if err != nil {
		return err
	}
	var name string
	if err := app.DB.NewSelect().Table("artworks").Column("name").Where("id = ?", artworkID).Scan(context.Background(), &name); err != nil {
		return err
	}
	if name != newArtworkBody().Name {
		return fmt.Errorf("other artist artwork name = %q, want unchanged %q", name, newArtworkBody().Name)
	}
	return nil
}

func (a *artworkContext) assertNotFound() error {
	if a.response.StatusCode != http.StatusNotFound {
		return fmt.Errorf("expected 404, got %d: %s", a.response.StatusCode, a.response.Body)
	}
	return nil
}

func (a *artworkContext) assertInvalidSampleDeletion() error {
	if a.response.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("expected 400, got %d: %s", a.response.StatusCode, a.response.Body)
	}
	artworkID, err := uuid.Parse(a.artworkID)
	if err != nil {
		return err
	}
	var name string
	if err := app.DB.NewSelect().Table("artworks").Column("name").Where("id = ?", artworkID).Scan(context.Background(), &name); err != nil {
		return err
	}
	imageCount, err := app.DB.NewSelect().Table("artwork_images").Where("artwork_id = ?", artworkID).Count(context.Background())
	if err != nil {
		return err
	}
	if name != newArtworkBody().Name || imageCount != len(newArtworkBody().ArtworkSamples) {
		return fmt.Errorf("artwork changed after rejected sample deletion: name=%q images=%d", name, imageCount)
	}
	return nil
}

func newArtworkBody() createArtworkBody {
	return createArtworkBody{
		Name:        "Book Cover",
		Category:    "Book",
		Styles:      []string{"Pixel Art", "Cartoon"},
		Description: "A colorful book-cover commission",
		ArtworkSamples: []artworkSampleBody{
			{ImageURL: "https://example.com/book-cover-one.png"},
			{ImageURL: "https://example.com/book-cover-two.png"},
		},
		MinimumDeadlineDays: 7,
		PriceSatang:         250000,
	}
}

func newUpdatedArtworkBody() createArtworkBody {
	return createArtworkBody{
		Name:        "Updated Book Cover",
		Category:    "Book",
		Styles:      []string{"Cartoon", "Watercolor"},
		Description: "An updated colorful book-cover commission",
		ArtworkSamples: []artworkSampleBody{
			{ImageURL: "https://example.com/updated-book-cover.png"},
		},
		MinimumDeadlineDays: 10,
		PriceSatang:         300000,
	}
}

func newClearedArtworkBody() createArtworkBody {
	body := newUpdatedArtworkBody()
	body.Styles = []string{}
	body.ArtworkSamples = []artworkSampleBody{}
	return body
}

func artworkFields(body createArtworkBody) map[string]string {
	styles, _ := json.Marshal(body.Styles)
	return map[string]string{
		"name":                  body.Name,
		"category":              body.Category,
		"styles":                string(styles),
		"description":           body.Description,
		"minimum_deadline_days": strconv.Itoa(body.MinimumDeadlineDays),
		"price_satang":          strconv.FormatInt(body.PriceSatang, 10),
	}
}

func artworkFiles(fieldName string, count int) []apptest.MultipartFile {
	files := make([]apptest.MultipartFile, count)
	for index := range files {
		files[index] = apptest.MultipartFile{
			FieldName:   fieldName,
			Filename:    fmt.Sprintf("sample-%d.png", index+1),
			ContentType: "image/png",
			Data:        []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, byte(index)},
		}
	}
	return files
}

func InitializeScenario(scenario *godog.ScenarioContext) {
	var state *artworkContext

	scenario.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		state = &artworkContext{client: apptest.NewClient(app.BaseURL())}
		return ctx, nil
	})

	scenario.Step(`^an artist is registered and logged in$`, func() error {
		if err := state.registerArtist(); err != nil {
			return err
		}
		return state.loginArtist()
	})
	scenario.Step(`^the artist creates artwork with samples$`, func() error { return state.createArtwork(state.accessToken) })
	scenario.Step(`^the artwork is stored with its category, styles, and ordered samples$`, func() error { return state.assertStoredArtwork() })
	scenario.Step(`^the artist has created artwork with samples$`, func() error {
		if err := state.createArtwork(state.accessToken); err != nil {
			return err
		}
		if state.response.StatusCode != http.StatusCreated {
			return fmt.Errorf("create artwork: expected 201, got %d: %s", state.response.StatusCode, state.response.Body)
		}
		return nil
	})
	scenario.Step(`^the artist deletes their artwork$`, func() error { return state.deleteArtwork(state.artworkID, state.accessToken) })
	scenario.Step(`^the artwork and its relations are deleted$`, func() error { return state.assertDeletedArtwork() })
	scenario.Step(`^another artist has created artwork with samples$`, func() error { return state.createArtworkForOtherArtist() })
	scenario.Step(`^the artist deletes the other artist's artwork$`, func() error { return state.deleteArtwork(state.otherArtwork, state.accessToken) })
	scenario.Step(`^the system reports the artwork was not found and keeps it$`, func() error { return state.assertOtherArtworkWasNotDeleted() })
	scenario.Step(`^an unauthenticated caller creates artwork with samples$`, func() error { return state.createArtwork("") })
	scenario.Step(`^the system requires the caller to log in$`, func() error {
		if state.response.StatusCode != http.StatusUnauthorized {
			return fmt.Errorf("expected 401, got %d: %s", state.response.StatusCode, state.response.Body)
		}
		return nil
	})
	scenario.Step(`^a customer is registered and logged in$`, func() error {
		account, err := apptest.RegisterCustomer(app, state.client)
		if err != nil {
			return err
		}
		token, err := apptest.Login(state.client, account.Email, account.Password)
		if err != nil {
			return err
		}
		state.accessToken = token
		return nil
	})
	scenario.Step(`^the customer creates artwork with samples$`, func() error { return state.createArtwork(state.accessToken) })
	scenario.Step(`^the system denies artwork creation$`, func() error {
		if state.response.StatusCode != http.StatusForbidden {
			return fmt.Errorf("expected 403, got %d: %s", state.response.StatusCode, state.response.Body)
		}
		return nil
	})
	scenario.Step(`^the artist updates their artwork details and selected samples$`, func() error {
		return state.updateArtwork(state.artworkID, state.accessToken, newUpdatedArtworkBody(), state.currentSampleURLs[:1])
	})
	scenario.Step(`^the artwork contains the updated values and retained samples$`, func() error {
		return state.assertUpdatedArtwork()
	})
	scenario.Step(`^the artist replaces their artwork with empty styles and samples$`, func() error {
		return state.updateArtwork(state.artworkID, state.accessToken, newClearedArtworkBody(), append([]string{}, state.currentSampleURLs...))
	})
	scenario.Step(`^the artwork has no styles or samples$`, func() error {
		return state.assertUpdatedArtwork()
	})
	scenario.Step(`^the artist updates the artwork with an unknown deleted sample URL$`, func() error {
		body := newUpdatedArtworkBody()
		body.ArtworkSamples = nil
		return state.updateArtwork(state.artworkID, state.accessToken, body, []string{"https://storage.example.com/not-owned.png"})
	})
	scenario.Step(`^the system rejects the sample deletion and leaves the artwork unchanged$`, func() error {
		return state.assertInvalidSampleDeletion()
	})
	scenario.Step(`^the artist updates the other artist's artwork$`, func() error {
		return state.updateArtwork(state.otherArtwork, state.accessToken, newUpdatedArtworkBody(), nil)
	})
	scenario.Step(`^the system reports the artwork was not found and leaves it unchanged$`, func() error {
		return state.assertOtherArtworkWasNotUpdated()
	})
	scenario.Step(`^the artist updates a missing artwork$`, func() error {
		return state.updateArtwork(uuid.NewString(), state.accessToken, newUpdatedArtworkBody(), nil)
	})
	scenario.Step(`^the system reports the artwork was not found$`, func() error {
		return state.assertNotFound()
	})
	scenario.Step(`^an unauthenticated caller updates the artwork$`, func() error {
		return state.updateArtwork(state.artworkID, "", newUpdatedArtworkBody(), nil)
	})
	scenario.Step(`^the customer updates the artwork$`, func() error {
		return state.updateArtwork(state.artworkID, state.accessToken, newUpdatedArtworkBody(), nil)
	})
	scenario.Step(`^the system denies artwork updates$`, func() error {
		if state.response.StatusCode != http.StatusForbidden {
			return fmt.Errorf("expected 403, got %d: %s", state.response.StatusCode, state.response.Body)
		}
		return nil
	})
}
