//go:build integration

package artists

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/AiSiriRak/Artmission/backend/tests/internal/apptest"
	"github.com/cucumber/godog"
	"github.com/google/uuid"
)

type referenceResponse struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type reviewResponse struct {
	Username string `json:"username"`
	Order    string `json:"order"`
	Rating   int    `json:"rating"`
}

type artistProfileResponse struct {
	ArtistID       string              `json:"artist_id"`
	ArtistName     string              `json:"artist_name"`
	ProfileURL     *string             `json:"profile_url"`
	Description    *string             `json:"description"`
	Categories     []referenceResponse `json:"categories"`
	Styles         []referenceResponse `json:"styles"`
	MinPriceSatang *int64              `json:"min_price_satang"`
	MaxPriceSatang *int64              `json:"max_price_satang"`
	ReviewScore    *float64            `json:"review_score"`
	Reviews        []reviewResponse    `json:"reviews"`
	Total          int                 `json:"total"`
}

type artworkSampleResponse struct {
	ImageURL string `json:"image_url"`
}

type artistArtworkResponse struct {
	ArtworkID           string                  `json:"artwork_id"`
	Name                string                  `json:"name"`
	Category            string                  `json:"category"`
	Styles              []string                `json:"styles"`
	Description         string                  `json:"description"`
	ArtworkSamples      []artworkSampleResponse `json:"artwork_samples"`
	MinimumDeadlineDays int                     `json:"minimum_deadline_days"`
	PriceSatang         int64                   `json:"price_satang"`
}

type artistArtworksResponse struct {
	Artworks []artistArtworkResponse `json:"artworks"`
}

type artistsContext struct {
	client              *apptest.Client
	artist              apptest.Account
	accessToken         string
	resp                *apptest.Response
	original            *artistProfileResponse
	artworkIDs          []uuid.UUID
	expectedReviewers   []string
	previousProfileURL  string
	currentProfileURL   string
	currentProfileImage []byte
}

func (artistContext *artistsContext) registerArtist() error {
	account, err := apptest.RegisterArtist(app, artistContext.client, "Original description")
	if err != nil {
		return err
	}
	artistContext.artist = account
	return nil
}

func (artistContext *artistsContext) loginArtist() error {
	token, err := apptest.Login(artistContext.client, artistContext.artist.Email, artistContext.artist.Password)
	if err != nil {
		return err
	}
	artistContext.accessToken = token
	return nil
}

func (artistContext *artistsContext) seedArtworkMetadata() error {
	artistID, err := uuid.Parse(artistContext.artist.ID)
	if err != nil {
		return err
	}
	categoryA, categoryB := uuid.New(), uuid.New()
	styleA, styleB := uuid.New(), uuid.New()
	artistContext.artworkIDs = []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	suffix := apptest.UniqueSuffix()
	oldest := time.Now().Add(-3 * time.Hour)

	if _, err := app.DB.ExecContext(context.Background(), `
		INSERT INTO categories (id, label) VALUES (?, ?), (?, ?)
	`, categoryA, "Landscape-"+suffix, categoryB, "Portrait-"+suffix); err != nil {
		return err
	}
	if _, err := app.DB.ExecContext(context.Background(), `
		INSERT INTO styles (id, label) VALUES (?, ?), (?, ?)
	`, styleA, "Anime-"+suffix, styleB, "Realism-"+suffix); err != nil {
		return err
	}
	if _, err := app.DB.ExecContext(context.Background(), `
		INSERT INTO artworks (id, artist_id, category_id, name, description, price_satang, minimum_deadline_days, created_at)
		VALUES
			(?, ?, ?, 'First', 'First artwork', 1000, 7, ?),
			(?, ?, ?, 'Second', 'Second artwork', 3000, 7, ?),
			(?, ?, ?, 'Third', 'Third artwork', 2000, 7, ?)
	`, artistContext.artworkIDs[0], artistID, categoryA, oldest,
		artistContext.artworkIDs[1], artistID, categoryA, oldest.Add(time.Hour),
		artistContext.artworkIDs[2], artistID, categoryB, oldest.Add(2*time.Hour)); err != nil {
		return err
	}
	if _, err = app.DB.ExecContext(context.Background(), `
		INSERT INTO artwork_styles (artwork_id, style_id)
		VALUES (?, ?), (?, ?), (?, ?), (?, ?)
	`, artistContext.artworkIDs[0], styleA, artistContext.artworkIDs[1], styleA, artistContext.artworkIDs[1], styleB, artistContext.artworkIDs[2], styleB); err != nil {
		return err
	}
	_, err = app.DB.ExecContext(context.Background(), `
		INSERT INTO artwork_images (id, artwork_id, image_url, sort_order)
		VALUES
			(?, ?, ?, 0),
			(?, ?, ?, 0),
			(?, ?, ?, 0)
	`, uuid.New(), artistContext.artworkIDs[0], "https://example.com/artworks/first/preview.webp",
		uuid.New(), artistContext.artworkIDs[1], "https://example.com/artworks/second/preview.webp",
		uuid.New(), artistContext.artworkIDs[2], "https://example.com/artworks/third/preview.webp")
	return err
}

func (artistContext *artistsContext) seedReviews() error {
	artistID, err := uuid.Parse(artistContext.artist.ID)
	if err != nil {
		return err
	}
	orderNames := []string{"Family Portrait", "Book Cover", "Pet Portrait"}
	ratings := []int{5, 3, 4}
	artistContext.expectedReviewers = make([]string, len(orderNames))
	baseTime := time.Now().Add(-time.Hour)

	for index := range orderNames {
		customer, err := apptest.RegisterCustomer(app, apptest.NewClient(app.BaseURL()))
		if err != nil {
			return err
		}
		customerID, err := uuid.Parse(customer.ID)
		if err != nil {
			return err
		}
		artistContext.expectedReviewers[index] = customer.Username
		orderID := uuid.New()
		createdAt := baseTime.Add(time.Duration(index) * time.Minute)
		if _, err := app.DB.ExecContext(context.Background(), `
			INSERT INTO orders (
				id, customer_id, artist_id, artwork_id, artwork_name_snapshot,
				artwork_description_snapshot, price_satang_snapshot,
				minimum_deadline_days_snapshot, name, customer_description,
				status, created_at, updated_at
			) VALUES (?, ?, ?, NULL, ?, 'Historical artwork', 1000, 7, ?, 'Please create this', 'SUCCESS', ?, ?)
		`, orderID, customerID, artistID, orderNames[index], orderNames[index], createdAt, createdAt); err != nil {
			return err
		}
		if _, err := app.DB.ExecContext(context.Background(), `
			INSERT INTO reviews (id, order_id, customer_id, artist_id, rating, created_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`, uuid.New(), orderID, customerID, artistID, ratings[index], createdAt); err != nil {
			return err
		}
	}
	return nil
}

func (artistContext *artistsContext) getArtistProfile(pathSuffix string) error {
	response, err := artistContext.client.Do(http.MethodGet, "/artists/"+artistContext.artist.ID+pathSuffix, nil, nil)
	if err != nil {
		return err
	}
	artistContext.resp = response
	return nil
}

func (artistContext *artistsContext) getArtistArtworks() error {
	response, err := artistContext.client.Do(http.MethodGet, "/artists/"+artistContext.artist.ID+"/artworks", nil, nil)
	if err != nil {
		return err
	}
	artistContext.resp = response
	return nil
}

func (artistContext *artistsContext) updateArtistProfile(fields map[string]string, file *apptest.MultipartFile, token string) error {
	headers := map[string]string{}
	if token != "" {
		headers["Authorization"] = "Bearer " + token
	}
	files := []apptest.MultipartFile{}
	if file != nil {
		files = append(files, *file)
	}
	response, err := artistContext.client.DoMultipart(http.MethodPut, "/artists/me", fields, files, headers)
	if err != nil {
		return err
	}
	artistContext.resp = response
	return nil
}

func (artistContext *artistsContext) requireProfile(status int) (*artistProfileResponse, error) {
	if artistContext.resp.StatusCode != status {
		return nil, fmt.Errorf("expected %d, got %d: %s", status, artistContext.resp.StatusCode, artistContext.resp.Body)
	}
	var profile artistProfileResponse
	if err := artistContext.resp.JSON(&profile); err != nil {
		return nil, fmt.Errorf("decode artist profile: %w", err)
	}
	return &profile, nil
}

func (artistContext *artistsContext) capturePublicProfile() error {
	if err := artistContext.getArtistProfile(""); err != nil {
		return err
	}
	profile, err := artistContext.requireProfile(http.StatusOK)
	if err != nil {
		return err
	}
	artistContext.original = profile
	return nil
}

func assertPublicImage(rawURL string, expected []byte) error {
	client := &http.Client{Timeout: 10 * time.Second}
	response, err := client.Get(rawURL)
	if err != nil {
		return fmt.Errorf("get public profile image: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("get public profile image: expected 200, got %d", response.StatusCode)
	}
	content, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(content, expected) {
		return fmt.Errorf("public profile image bytes differ: got %v want %v", content, expected)
	}
	return nil
}

func InitializeScenario(scenario *godog.ScenarioContext) {
	var artistContext *artistsContext

	scenario.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		artistContext = &artistsContext{client: apptest.NewClient(app.BaseURL())}
		return ctx, nil
	})

	scenario.Step(`^the artist has a registered account$`, func() error { return artistContext.registerArtist() })
	scenario.Step(`^the artist has logged in$`, func() error { return artistContext.loginArtist() })
	scenario.Step(`^the artist has artworks in multiple categories and styles$`, func() error { return artistContext.seedArtworkMetadata() })
	scenario.Step(`^the artist has three customer reviews$`, func() error { return artistContext.seedReviews() })
	scenario.Step(`^the artist account has been deleted$`, func() error {
		_, err := app.DB.ExecContext(context.Background(), "UPDATE users SET deleted_at = now() WHERE id = ?", artistContext.artist.ID)
		return err
	})
	scenario.Step(`^the customer has logged in$`, func() error {
		account, err := apptest.RegisterCustomer(app, artistContext.client)
		if err != nil {
			return err
		}
		artistContext.accessToken, err = apptest.Login(artistContext.client, account.Email, account.Password)
		return err
	})

	scenario.Step(`^a visitor requests the artist profile$`, func() error { return artistContext.getArtistProfile("") })
	scenario.Step(`^a visitor requests the artist artworks$`, func() error { return artistContext.getArtistArtworks() })
	scenario.Step(`^a visitor requests artworks for an artist that does not exist$`, func() error {
		artistContext.artist.ID = uuid.NewString()
		return artistContext.getArtistArtworks()
	})
	scenario.Step(`^a visitor requests two artist reviews$`, func() error { return artistContext.getArtistProfile("?limit=2&offset=0") })
	scenario.Step(`^a visitor requests an artist profile that does not exist$`, func() error {
		artistContext.artist.ID = uuid.NewString()
		return artistContext.getArtistProfile("")
	})
	scenario.Step(`^the artist updates only their description$`, func() error {
		return artistContext.updateArtistProfile(map[string]string{"description": "  Updated commission profile  "}, nil, artistContext.accessToken)
	})
	scenario.Step(`^the artist updates their description with blank text$`, func() error {
		return artistContext.updateArtistProfile(map[string]string{"description": "   "}, nil, artistContext.accessToken)
	})
	scenario.Step(`^the artist uploads a profile image$`, func() error {
		artistContext.currentProfileImage = pngImage()
		return artistContext.updateArtistProfile(nil, &apptest.MultipartFile{
			FieldName: "profile_image", Filename: "avatar.png", ContentType: "image/png", Data: artistContext.currentProfileImage,
		}, artistContext.accessToken)
	})
	scenario.Step(`^the artist replaces the profile image$`, func() error {
		artistContext.previousProfileURL = artistContext.currentProfileURL
		artistContext.currentProfileImage = webPImage()
		return artistContext.updateArtistProfile(nil, &apptest.MultipartFile{
			FieldName: "profile_image", Filename: "avatar.webp", ContentType: "image/webp", Data: artistContext.currentProfileImage,
		}, artistContext.accessToken)
	})
	scenario.Step(`^the artist removes the profile image$`, func() error {
		artistContext.previousProfileURL = artistContext.currentProfileURL
		return artistContext.updateArtistProfile(map[string]string{"remove_profile_image": "true"}, nil, artistContext.accessToken)
	})
	scenario.Step(`^the artist uploads invalid profile image content$`, func() error {
		if err := artistContext.capturePublicProfile(); err != nil {
			return err
		}
		return artistContext.updateArtistProfile(nil, &apptest.MultipartFile{
			FieldName: "profile_image", Filename: "fake.png", ContentType: "image/png", Data: []byte("not an image"),
		}, artistContext.accessToken)
	})
	scenario.Step(`^the artist tries to update an artwork-derived price$`, func() error {
		return artistContext.updateArtistProfile(map[string]string{"min_price_satang": "100"}, nil, artistContext.accessToken)
	})
	scenario.Step(`^the artist submits no profile changes$`, func() error {
		return artistContext.updateArtistProfile(nil, nil, artistContext.accessToken)
	})
	scenario.Step(`^someone updates the artist profile without logging in$`, func() error {
		return artistContext.updateArtistProfile(map[string]string{"description": "new"}, nil, "")
	})
	scenario.Step(`^the customer updates the artist profile$`, func() error {
		return artistContext.updateArtistProfile(map[string]string{"description": "new"}, nil, artistContext.accessToken)
	})

	scenario.Step(`^the system returns an empty initial artist profile$`, func() error {
		profile, err := artistContext.requireProfile(http.StatusOK)
		if err != nil {
			return err
		}
		if profile.ArtistID != artistContext.artist.ID || profile.ArtistName != artistContext.artist.Username || profile.Description == nil || *profile.Description != "Original description" {
			return fmt.Errorf("unexpected initial profile: %+v", profile)
		}
		if profile.ProfileURL != nil || profile.MinPriceSatang != nil || profile.MaxPriceSatang != nil || profile.ReviewScore != nil {
			return fmt.Errorf("expected nullable initial values, got %+v", profile)
		}
		if profile.Categories == nil || profile.Styles == nil || profile.Reviews == nil || len(profile.Categories) != 0 || len(profile.Styles) != 0 || len(profile.Reviews) != 0 || profile.Total != 0 {
			return fmt.Errorf("expected empty derived arrays and reviews, got %+v", profile)
		}
		return nil
	})
	scenario.Step(`^the system returns distinct artwork metadata and its price range$`, func() error {
		profile, err := artistContext.requireProfile(http.StatusOK)
		if err != nil {
			return err
		}
		if len(profile.Categories) != 2 || profile.Categories[0].Label > profile.Categories[1].Label {
			return fmt.Errorf("expected two distinct sorted categories, got %+v", profile.Categories)
		}
		if len(profile.Styles) != 2 || profile.Styles[0].Label > profile.Styles[1].Label {
			return fmt.Errorf("expected two distinct sorted styles, got %+v", profile.Styles)
		}
		if profile.MinPriceSatang == nil || *profile.MinPriceSatang != 1000 || profile.MaxPriceSatang == nil || *profile.MaxPriceSatang != 3000 {
			return fmt.Errorf("unexpected derived price range: min=%v max=%v", profile.MinPriceSatang, profile.MaxPriceSatang)
		}
		return nil
	})
	scenario.Step(`^the system returns every artist artwork newest first$`, func() error {
		if artistContext.resp.StatusCode != http.StatusOK {
			return fmt.Errorf("expected 200, got %d: %s", artistContext.resp.StatusCode, artistContext.resp.Body)
		}
		var response artistArtworksResponse
		if err := artistContext.resp.JSON(&response); err != nil {
			return err
		}
		if response.Artworks == nil || len(response.Artworks) != 3 {
			return fmt.Errorf("expected three artworks, got %+v", response.Artworks)
		}
		wantNames := []string{"Third", "Second", "First"}
		wantPrices := []int64{2000, 3000, 1000}
		for index, item := range response.Artworks {
			if item.ArtworkID != artistContext.artworkIDs[2-index].String() || item.Name != wantNames[index] || item.PriceSatang != wantPrices[index] {
				return fmt.Errorf("artwork %d = %+v", index, item)
			}
			if item.Category == "" || item.Styles == nil || item.ArtworkSamples == nil || len(item.ArtworkSamples) != 1 {
				return fmt.Errorf("artwork metadata is incomplete: %+v", item)
			}
			if !strings.HasSuffix(item.ArtworkSamples[0].ImageURL, "/artworks/"+strings.ToLower(item.Name)+"/preview.webp") {
				return fmt.Errorf("unexpected artwork image URL: %q", item.ArtworkSamples[0].ImageURL)
			}
		}
		return nil
	})
	scenario.Step(`^the system returns an empty artwork list$`, func() error {
		if artistContext.resp.StatusCode != http.StatusOK {
			return fmt.Errorf("expected 200, got %d: %s", artistContext.resp.StatusCode, artistContext.resp.Body)
		}
		var response artistArtworksResponse
		if err := artistContext.resp.JSON(&response); err != nil {
			return err
		}
		if response.Artworks == nil || len(response.Artworks) != 0 {
			return fmt.Errorf("expected an empty non-null artwork list, got %+v", response.Artworks)
		}
		return nil
	})
	scenario.Step(`^the system reports that the artist artworks were not found$`, func() error {
		if artistContext.resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("expected 404, got %d: %s", artistContext.resp.StatusCode, artistContext.resp.Body)
		}
		return nil
	})
	scenario.Step(`^the system returns the newest two reviews and the complete average$`, func() error {
		profile, err := artistContext.requireProfile(http.StatusOK)
		if err != nil {
			return err
		}
		if profile.Total != 3 || len(profile.Reviews) != 2 {
			return fmt.Errorf("reviews page = %+v total=%d", profile.Reviews, profile.Total)
		}
		if profile.ReviewScore == nil || *profile.ReviewScore != 4.0 {
			return fmt.Errorf("review score = %v, want 4.0", profile.ReviewScore)
		}
		if profile.Reviews[0] != (reviewResponse{Username: artistContext.expectedReviewers[2], Order: "Pet Portrait", Rating: 4}) ||
			profile.Reviews[1] != (reviewResponse{Username: artistContext.expectedReviewers[1], Order: "Book Cover", Rating: 3}) {
			return fmt.Errorf("unexpected review ordering/content: %+v", profile.Reviews)
		}
		return nil
	})
	scenario.Step(`^the system trims and saves the artist description$`, func() error {
		profile, err := artistContext.requireProfile(http.StatusOK)
		if err != nil {
			return err
		}
		if profile.Description == nil || *profile.Description != "Updated commission profile" {
			return fmt.Errorf("description = %#v", profile.Description)
		}
		if err := artistContext.getArtistProfile(""); err != nil {
			return err
		}
		public, err := artistContext.requireProfile(http.StatusOK)
		if err != nil || public.Description == nil || *public.Description != "Updated commission profile" {
			return fmt.Errorf("public description was not updated: %+v, error=%v", public, err)
		}
		return nil
	})
	scenario.Step(`^the system returns a null artist description$`, func() error {
		profile, err := artistContext.requireProfile(http.StatusOK)
		if err != nil {
			return err
		}
		if profile.Description != nil {
			return fmt.Errorf("description = %q, want null", *profile.Description)
		}
		return nil
	})
	scenario.Step(`^the profile image is publicly accessible$`, func() error {
		profile, err := artistContext.requireProfile(http.StatusOK)
		if err != nil {
			return err
		}
		if profile.ProfileURL == nil {
			return fmt.Errorf("profile_url is null")
		}
		artistContext.currentProfileURL = *profile.ProfileURL
		return assertPublicImage(artistContext.currentProfileURL, artistContext.currentProfileImage)
	})
	scenario.Step(`^the replacement profile image is publicly accessible and has a new URL$`, func() error {
		profile, err := artistContext.requireProfile(http.StatusOK)
		if err != nil {
			return err
		}
		if profile.ProfileURL == nil || *profile.ProfileURL == artistContext.previousProfileURL {
			return fmt.Errorf("replacement profile_url = %v, previous=%q", profile.ProfileURL, artistContext.previousProfileURL)
		}
		artistContext.currentProfileURL = *profile.ProfileURL
		return assertPublicImage(artistContext.currentProfileURL, artistContext.currentProfileImage)
	})
	scenario.Step(`^the artist profile has no profile image$`, func() error {
		profile, err := artistContext.requireProfile(http.StatusOK)
		if err != nil {
			return err
		}
		if profile.ProfileURL != nil {
			return fmt.Errorf("profile_url = %q, want null", *profile.ProfileURL)
		}
		return nil
	})
	scenario.Step(`^the system rejects the artist update without changing the profile$`, func() error {
		if artistContext.resp.StatusCode != http.StatusBadRequest {
			return fmt.Errorf("expected 400, got %d: %s", artistContext.resp.StatusCode, artistContext.resp.Body)
		}
		if err := artistContext.getArtistProfile(""); err != nil {
			return err
		}
		profile, err := artistContext.requireProfile(http.StatusOK)
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(profile, artistContext.original) {
			return fmt.Errorf("profile changed after rejection: before=%+v after=%+v", artistContext.original, profile)
		}
		return nil
	})
	scenario.Step(`^the system rejects the unsupported artist field$`, func() error {
		if artistContext.resp.StatusCode != http.StatusUnprocessableEntity {
			return fmt.Errorf("expected 422, got %d: %s", artistContext.resp.StatusCode, artistContext.resp.Body)
		}
		return nil
	})
	scenario.Step(`^the system rejects the empty artist update$`, func() error {
		if artistContext.resp.StatusCode != http.StatusBadRequest {
			return fmt.Errorf("expected 400, got %d: %s", artistContext.resp.StatusCode, artistContext.resp.Body)
		}
		return nil
	})
	scenario.Step(`^the system requires the user to log in$`, func() error {
		if artistContext.resp.StatusCode != http.StatusUnauthorized {
			return fmt.Errorf("expected 401, got %d: %s", artistContext.resp.StatusCode, artistContext.resp.Body)
		}
		return nil
	})
	scenario.Step(`^the system denies the artist profile update$`, func() error {
		if artistContext.resp.StatusCode != http.StatusForbidden {
			return fmt.Errorf("expected 403, got %d: %s", artistContext.resp.StatusCode, artistContext.resp.Body)
		}
		return nil
	})
	scenario.Step(`^the system reports that the artist profile was not found$`, func() error {
		if artistContext.resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("expected 404, got %d: %s", artistContext.resp.StatusCode, artistContext.resp.Body)
		}
		return nil
	})
}

func pngImage() []byte {
	return []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00}
}

func webPImage() []byte {
	return []byte{'R', 'I', 'F', 'F', 4, 0, 0, 0, 'W', 'E', 'B', 'P', 'd', 'a', 't', 'a'}
}
