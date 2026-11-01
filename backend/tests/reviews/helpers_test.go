//go:build integration

package reviews

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/AiSiriRak/Artmission/backend/tests/internal/apptest"
	"github.com/google/uuid"
)

type reviewResponse struct {
	ID         uuid.UUID `json:"id"`
	OrderID    uuid.UUID `json:"order_id"`
	CustomerID uuid.UUID `json:"customer_id"`
	ArtistID   uuid.UUID `json:"artist_id"`
	Rating     int       `json:"rating"`
	Comment    *string   `json:"comment"`
	CreatedAt  time.Time `json:"created_at"`
	Username   string    `json:"username"`
	Order      string    `json:"order"`
}

type reviewsContext struct {
	client          *apptest.Client
	customer        apptest.Account
	artist          apptest.Account
	customerToken   string
	artistToken     string
	orderID         uuid.UUID
	artworkID       uuid.UUID
	categoryID      uuid.UUID
	response        *apptest.Response
	existingReviews int
	expectedIDs     []uuid.UUID
	expectedDates   []time.Time
	concurrent      []submissionResult
}

type submissionResult struct {
	response *apptest.Response
	err      error
}

// prepareCommission seeds an order state without replaying the unrelated payment flow.
func (c *reviewsContext) prepareCommission(status string) error {
	var err error
	c.customer, err = apptest.RegisterCustomer(app, c.client)
	if err != nil {
		return err
	}
	c.artist, err = apptest.RegisterArtist(app, c.client, "Review test artist")
	if err != nil {
		return err
	}
	c.customerToken, err = apptest.Login(c.client, c.customer.Email, c.customer.Password)
	if err != nil {
		return err
	}
	c.artistToken, err = apptest.Login(c.client, c.artist.Email, c.artist.Password)
	if err != nil {
		return err
	}
	c.categoryID, c.artworkID = uuid.New(), uuid.New()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := app.DB.ExecContext(ctx,
		"INSERT INTO categories (id, label) VALUES (?, ?)",
		c.categoryID, "Review category "+apptest.UniqueSuffix(),
	); err != nil {
		return err
	}
	if _, err := app.DB.ExecContext(ctx, `
		INSERT INTO artworks (id, artist_id, category_id, name, description, price_satang, minimum_deadline_days)
		VALUES (?, ?, ?, 'Current artwork name', 'Review fixture', 1000, 7)
	`, c.artworkID, c.artist.ID, c.categoryID); err != nil {
		return err
	}
	c.orderID, err = c.createOrder(status, time.Now().UTC())
	return err
}

func (c *reviewsContext) createOrder(status string, createdAt time.Time) (uuid.UUID, error) {
	id := uuid.New()
	snapshot, err := json.Marshal(map[string]any{
		"artwork_name": "Purchased artwork name",
		"category_id":  c.categoryID,
		"style_ids":    []string{},
	})
	if err != nil {
		return uuid.Nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = app.DB.ExecContext(ctx, `
		INSERT INTO orders (
			id, customer_id, artist_id, artwork_id, artwork_snapshot, price_satang_order,
			name, customer_description, deadline_at, status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?::jsonb, 1000, 'Customer commission title', 'Review fixture', ?, ?, ?, ?)
	`, id, c.customer.ID, c.artist.ID, c.artworkID, string(snapshot), createdAt.AddDate(0, 0, 7), status, createdAt, createdAt)
	return id, err
}

func (c *reviewsContext) submit(orderID uuid.UUID, body any, token string) error {
	headers := map[string]string{}
	if token != "" {
		headers["Authorization"] = "Bearer " + token
	}
	var err error
	c.response, err = c.client.Do(http.MethodPost, "/orders/"+orderID.String()+"/review", body, headers)
	return err
}

func (c *reviewsContext) listReviews(token string) error {
	headers := map[string]string{}
	if token != "" {
		headers["Authorization"] = "Bearer " + token
	}
	var err error
	c.response, err = c.client.Do(http.MethodGet, "/artists/"+c.artist.ID+"/reviews", nil, headers)
	return err
}

func (c *reviewsContext) requireStatus(want int) error {
	if c.response == nil {
		return fmt.Errorf("no HTTP response")
	}
	if c.response.StatusCode != want {
		return fmt.Errorf("expected status %d, got %d: %s", want, c.response.StatusCode, c.response.Body)
	}
	return nil
}

func (c *reviewsContext) countOrderReviews() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return app.DB.NewSelect().Table("reviews").Where("order_id = ?", c.orderID).Count(ctx)
}

func (c *reviewsContext) requireAverage(want *float64) error {
	response, err := c.client.Do(http.MethodGet, "/artists/"+c.artist.ID, nil, nil)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("artist profile returned %d: %s", response.StatusCode, response.Body)
	}
	var profile struct {
		Score *float64 `json:"review_score"`
	}
	if err := response.JSON(&profile); err != nil {
		return err
	}
	if want == nil {
		if profile.Score != nil {
			return fmt.Errorf("expected no average score, got %v", *profile.Score)
		}
	} else if profile.Score == nil || *profile.Score != *want {
		return fmt.Errorf("expected average %v, got %v", *want, profile.Score)
	}
	return nil
}
