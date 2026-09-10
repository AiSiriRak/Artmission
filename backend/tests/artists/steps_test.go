//go:build integration

package artists

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	"github.com/AiSiriRak/Artmission/backend/tests/internal/apptest"
	"github.com/cucumber/godog"
	"github.com/google/uuid"
)

type referenceResponse struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type artistProfileResponse struct {
	ArtistID       string              `json:"artist_id"`
	ArtistName     string              `json:"artist_name"`
	Description    *string             `json:"description"`
	Categories     []referenceResponse `json:"categories"`
	Styles         []referenceResponse `json:"styles"`
	MinPriceSatang *int64              `json:"min_price_satang"`
	MaxPriceSatang *int64              `json:"max_price_satang"`
	ReviewScore    *float64            `json:"review_score"`
}

type updateArtistProfileBody struct {
	Description    *string  `json:"description"`
	StyleIDs       []string `json:"style_ids"`
	MinPriceSatang int64    `json:"min_price_satang"`
	MaxPriceSatang int64    `json:"max_price_satang"`
}

type artistsContext struct {
	client      *apptest.Client
	artist      apptest.Account
	accessToken string
	styleIDs    []uuid.UUID
	resp        *apptest.Response
	original    *artistProfileResponse
}

func (a *artistsContext) registerArtist() error {
	account, err := apptest.RegisterArtist(app, a.client, "Original description")
	if err != nil {
		return err
	}
	a.artist = account
	return nil
}

func (a *artistsContext) registerArtistWithoutDescription() error {
	account, err := apptest.RegisterArtist(app, a.client, "")
	if err != nil {
		return err
	}
	a.artist = account
	return nil
}

func (a *artistsContext) loginArtist() error {
	token, err := apptest.Login(a.client, a.artist.Email, a.artist.Password)
	if err != nil {
		return err
	}
	a.accessToken = token
	return nil
}

func (a *artistsContext) seedStyles() error {
	a.styleIDs = []uuid.UUID{uuid.New(), uuid.New()}
	_, err := app.DB.ExecContext(context.Background(), `
		INSERT INTO styles (id, label) VALUES (?, ?), (?, ?)
	`, a.styleIDs[0], "Realism-"+apptest.UniqueSuffix(), a.styleIDs[1], "Anime-"+apptest.UniqueSuffix())
	return err
}

func (a *artistsContext) seedArtworkCategories() error {
	artistID, err := uuid.Parse(a.artist.ID)
	if err != nil {
		return err
	}
	categoryA, categoryB := uuid.New(), uuid.New()
	suffix := apptest.UniqueSuffix()
	if _, err := app.DB.ExecContext(context.Background(), `
		INSERT INTO categories (id, label) VALUES (?, ?), (?, ?)
	`, categoryA, "Landscape-"+suffix, categoryB, "Portrait-"+suffix); err != nil {
		return err
	}
	_, err = app.DB.ExecContext(context.Background(), `
		INSERT INTO artworks (id, artist_id, category_id, name, description, price_satang, minimum_deadline_days)
		VALUES
			(?, ?, ?, 'First', 'First artwork', 1000, 7),
			(?, ?, ?, 'Second', 'Second artwork', 2000, 7),
			(?, ?, ?, 'Third', 'Third artwork', 3000, 7)
	`, uuid.New(), artistID, categoryA, uuid.New(), artistID, categoryA, uuid.New(), artistID, categoryB)
	return err
}

func (a *artistsContext) seedReviewScore() error {
	_, err := app.DB.ExecContext(context.Background(), "UPDATE artist_profiles SET review_score = 4.5 WHERE user_id = ?", a.artist.ID)
	return err
}

func (a *artistsContext) getArtistProfile(artistID string) error {
	resp, err := a.client.Do(http.MethodGet, "/artists/"+artistID, nil, nil)
	if err != nil {
		return err
	}
	a.resp = resp
	return nil
}

func (a *artistsContext) updateArtistProfile(body updateArtistProfileBody, token string) error {
	headers := map[string]string{}
	if token != "" {
		headers["Authorization"] = "Bearer " + token
	}
	resp, err := a.client.Do(http.MethodPut, "/artists/me", body, headers)
	if err != nil {
		return err
	}
	a.resp = resp
	return nil
}

func (a *artistsContext) validUpdate() updateArtistProfileBody {
	return updateArtistProfileBody{
		Description:    stringPointer("  Updated commission profile  "),
		StyleIDs:       []string{a.styleIDs[0].String(), a.styleIDs[1].String(), a.styleIDs[0].String()},
		MinPriceSatang: 10_000,
		MaxPriceSatang: 50_000,
	}
}

func (a *artistsContext) configureProfile() error {
	if err := a.registerArtist(); err != nil {
		return err
	}
	if err := a.seedStyles(); err != nil {
		return err
	}
	if err := a.loginArtist(); err != nil {
		return err
	}
	if err := a.updateArtistProfile(a.validUpdate(), a.accessToken); err != nil {
		return err
	}
	if a.resp.StatusCode != http.StatusOK {
		return fmt.Errorf("configure profile: expected 200, got %d: %s", a.resp.StatusCode, a.resp.Body)
	}
	profile, err := decodeProfile(a.resp)
	if err != nil {
		return err
	}
	a.original = profile
	return nil
}

func (a *artistsContext) assertInitialProfile() error {
	profile, err := a.requireProfile(http.StatusOK)
	if err != nil {
		return err
	}
	if profile.ArtistID != a.artist.ID || profile.ArtistName != a.artist.Username || profile.Description == nil || *profile.Description != "Original description" {
		return fmt.Errorf("unexpected initial profile: %+v", profile)
	}
	if profile.MinPriceSatang != nil || profile.MaxPriceSatang != nil || profile.ReviewScore != nil {
		return fmt.Errorf("expected nullable initial values, got %+v", profile)
	}
	if profile.Categories == nil || len(profile.Categories) != 0 || profile.Styles == nil || len(profile.Styles) != 0 {
		return fmt.Errorf("expected empty category/style arrays, got %+v", profile)
	}
	return nil
}

func (a *artistsContext) assertCategoriesAndReview() error {
	profile, err := a.requireProfile(http.StatusOK)
	if err != nil {
		return err
	}
	if len(profile.Categories) != 2 || profile.Categories[0].Label > profile.Categories[1].Label {
		return fmt.Errorf("expected two distinct sorted categories, got %+v", profile.Categories)
	}
	if profile.ReviewScore == nil || *profile.ReviewScore != 4.5 {
		return fmt.Errorf("expected review score 4.5, got %+v", profile.ReviewScore)
	}
	return nil
}

func (a *artistsContext) assertSavedProfile() error {
	profile, err := a.requireProfile(http.StatusOK)
	if err != nil {
		return err
	}
	if profile.Description == nil || *profile.Description != "Updated commission profile" || profile.MinPriceSatang == nil || *profile.MinPriceSatang != 10_000 || profile.MaxPriceSatang == nil || *profile.MaxPriceSatang != 50_000 {
		return fmt.Errorf("unexpected updated profile: %+v", profile)
	}
	if len(profile.Styles) != 2 || profile.Styles[0].Label > profile.Styles[1].Label {
		return fmt.Errorf("expected two unique sorted styles, got %+v", profile.Styles)
	}
	if err := a.getArtistProfile(a.artist.ID); err != nil {
		return err
	}
	public, err := a.requireProfile(http.StatusOK)
	if err != nil {
		return err
	}
	if !sameOptionalString(public.Description, profile.Description) || !reflect.DeepEqual(public.Styles, profile.Styles) {
		return fmt.Errorf("public profile does not reflect update: %+v", public)
	}
	return nil
}

func (a *artistsContext) assertUnchangedAfterRejection() error {
	if a.resp.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("expected 400, got %d: %s", a.resp.StatusCode, a.resp.Body)
	}
	if err := a.getArtistProfile(a.artist.ID); err != nil {
		return err
	}
	profile, err := a.requireProfile(http.StatusOK)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(profile, a.original) {
		return fmt.Errorf("profile changed after rejected update: before=%+v after=%+v", a.original, profile)
	}
	return nil
}

func (a *artistsContext) requireProfile(status int) (*artistProfileResponse, error) {
	if a.resp.StatusCode != status {
		return nil, fmt.Errorf("expected %d, got %d: %s", status, a.resp.StatusCode, a.resp.Body)
	}
	return decodeProfile(a.resp)
}

func decodeProfile(resp *apptest.Response) (*artistProfileResponse, error) {
	var profile artistProfileResponse
	if err := resp.JSON(&profile); err != nil {
		return nil, fmt.Errorf("decode artist profile: %w", err)
	}
	return &profile, nil
}

func InitializeScenario(sc *godog.ScenarioContext) {
	var a *artistsContext

	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		a = &artistsContext{client: apptest.NewClient(app.BaseURL())}
		return ctx, nil
	})

	sc.Step(`^the artist has a registered account$`, func() error { return a.registerArtist() })
	sc.Step(`^the artist has a registered account without a description$`, func() error { return a.registerArtistWithoutDescription() })
	sc.Step(`^reference styles are available$`, func() error { return a.seedStyles() })
	sc.Step(`^the artist has logged in$`, func() error { return a.loginArtist() })
	sc.Step(`^the artist has samples in multiple categories$`, func() error { return a.seedArtworkCategories() })
	sc.Step(`^the artist has a review score$`, func() error { return a.seedReviewScore() })
	sc.Step(`^the artist account has been deleted$`, func() error {
		_, err := app.DB.ExecContext(context.Background(), "UPDATE users SET deleted_at = now() WHERE id = ?", a.artist.ID)
		return err
	})
	sc.Step(`^the artist has a configured profile$`, func() error { return a.configureProfile() })
	sc.Step(`^the customer has logged in$`, func() error {
		account, err := apptest.RegisterCustomer(app, a.client)
		if err != nil {
			return err
		}
		a.accessToken, err = apptest.Login(a.client, account.Email, account.Password)
		return err
	})
	sc.Step(`^a visitor requests the artist profile$`, func() error { return a.getArtistProfile(a.artist.ID) })
	sc.Step(`^a visitor requests an artist profile that does not exist$`, func() error { return a.getArtistProfile(uuid.NewString()) })
	sc.Step(`^the artist updates their profile with valid details$`, func() error { return a.updateArtistProfile(a.validUpdate(), a.accessToken) })
	sc.Step(`^the artist clears their selected styles$`, func() error {
		return a.updateArtistProfile(updateArtistProfileBody{Description: stringPointer("No selected styles"), StyleIDs: []string{}, MinPriceSatang: 100, MaxPriceSatang: 200}, a.accessToken)
	})
	sc.Step(`^the artist updates their profile with an invalid price range$`, func() error {
		body := a.validUpdate()
		body.Description = stringPointer("must not persist")
		body.MinPriceSatang, body.MaxPriceSatang = 500, 100
		return a.updateArtistProfile(body, a.accessToken)
	})
	sc.Step(`^the artist updates their profile with an unknown style$`, func() error {
		body := a.validUpdate()
		body.Description = stringPointer("must not persist")
		body.StyleIDs = []string{uuid.NewString()}
		return a.updateArtistProfile(body, a.accessToken)
	})
	sc.Step(`^someone updates the artist profile without logging in$`, func() error {
		return a.updateArtistProfile(updateArtistProfileBody{Description: stringPointer("new"), StyleIDs: []string{}, MaxPriceSatang: 1}, "")
	})
	sc.Step(`^the customer updates the artist profile$`, func() error {
		return a.updateArtistProfile(updateArtistProfileBody{Description: stringPointer("new"), StyleIDs: []string{}, MaxPriceSatang: 1}, a.accessToken)
	})
	sc.Step(`^the system returns the initial public artist profile$`, func() error { return a.assertInitialProfile() })
	sc.Step(`^the system returns a null artist description$`, func() error {
		profile, err := a.requireProfile(http.StatusOK)
		if err != nil {
			return err
		}
		if profile.Description != nil {
			return fmt.Errorf("description = %q, want null", *profile.Description)
		}
		return nil
	})
	sc.Step(`^the system returns distinct artwork categories and the review score$`, func() error { return a.assertCategoriesAndReview() })
	sc.Step(`^the system saves and returns the complete artist profile$`, func() error { return a.assertSavedProfile() })
	sc.Step(`^the system returns an empty style selection$`, func() error {
		profile, err := a.requireProfile(http.StatusOK)
		if err != nil {
			return err
		}
		if profile.Styles == nil || len(profile.Styles) != 0 {
			return fmt.Errorf("expected empty styles, got %+v", profile.Styles)
		}
		return nil
	})
	sc.Step(`^the system rejects the artist update without changing the profile$`, func() error { return a.assertUnchangedAfterRejection() })
	sc.Step(`^the system requires the user to log in$`, func() error {
		if a.resp.StatusCode != http.StatusUnauthorized {
			return fmt.Errorf("expected 401, got %d: %s", a.resp.StatusCode, a.resp.Body)
		}
		return nil
	})
	sc.Step(`^the system denies the artist profile update$`, func() error {
		if a.resp.StatusCode != http.StatusForbidden {
			return fmt.Errorf("expected 403, got %d: %s", a.resp.StatusCode, a.resp.Body)
		}
		return nil
	})
	sc.Step(`^the system reports that the artist profile was not found$`, func() error {
		if a.resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("expected 404, got %d: %s", a.resp.StatusCode, a.resp.Body)
		}
		return nil
	})
}

func stringPointer(value string) *string { return &value }

func sameOptionalString(left, right *string) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}
