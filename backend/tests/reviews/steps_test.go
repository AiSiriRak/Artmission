//go:build integration

package reviews

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/AiSiriRak/Artmission/backend/tests/internal/apptest"
	"github.com/cucumber/godog"
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
	orderID         uuid.UUID
	artworkID       uuid.UUID
	categoryID      uuid.UUID
	response        *apptest.Response
	existingReviews int
	expectedIDs     []uuid.UUID
	expectedDates   []time.Time
}

func InitializeScenario(sc *godog.ScenarioContext) {
	var c *reviewsContext
	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		c = &reviewsContext{client: apptest.NewClient(app.BaseURL())}
		return ctx, nil
	})
	sc.Step(`^a commission in status "([^"]+)"$`, func(status string) error {
		return c.prepareCommission(status)
	})
	sc.Step(`^the customer submits this review:$`, func(body *godog.DocString) error {
		var payload any
		if err := json.Unmarshal([]byte(body.Content), &payload); err != nil {
			return err
		}
		return c.submit(c.orderID, payload, c.customerToken)
	})
	sc.Step(`^the review is saved with rating (\d+) and comment "([^"]*)"$`, func(rating int, comment string) error {
		if err := c.requireStatus(http.StatusCreated); err != nil {
			return err
		}
		var result reviewResponse
		if err := c.response.JSON(&result); err != nil {
			return err
		}
		if result.ID == uuid.Nil || result.OrderID != c.orderID || result.CustomerID.String() != c.customer.ID || result.ArtistID.String() != c.artist.ID || result.Rating != rating || result.CreatedAt.IsZero() {
			return fmt.Errorf("unexpected created review: %+v", result)
		}
		if err := checkComment(result.Comment, comment); err != nil {
			return err
		}
		var stored reviewResponse
		if err := app.DB.NewSelect().Table("reviews").Where("order_id = ?", c.orderID).Scan(context.Background(), &stored); err != nil {
			return err
		}
		if stored.ID != result.ID || stored.OrderID != c.orderID || stored.CustomerID != result.CustomerID || stored.ArtistID != result.ArtistID || stored.Rating != rating || stored.CreatedAt.Sub(result.CreatedAt).Abs() > time.Microsecond {
			return fmt.Errorf("database review does not match response: %+v", stored)
		}
		return checkComment(stored.Comment, comment)
	})
	sc.Step(`^the artist average score is ([0-9.]+)$`, func(want float64) error {
		return c.requireAverage(&want)
	})
	sc.Step(`^the review is rejected as "([^"]+)"$`, func(reason string) error {
		statuses := map[string]int{
			"invalid request":       http.StatusUnprocessableEntity,
			"unfinished commission": http.StatusConflict,
			"duplicate review":      http.StatusConflict,
			"forbidden":             http.StatusForbidden,
		}
		status, ok := statuses[reason]
		if !ok {
			return fmt.Errorf("unknown rejection reason %q", reason)
		}
		if err := c.requireStatus(status); err != nil {
			return err
		}
		count, err := c.countOrderReviews()
		if err != nil {
			return err
		}
		if count != c.existingReviews {
			return fmt.Errorf("a rejected request changed the review count: got %d, want %d", count, c.existingReviews)
		}
		return nil
	})
	sc.Step(`^"([^"]+)" submits a review$`, func(caller string) error {
		token, err := c.tokenFor(caller)
		if err != nil {
			return err
		}
		return c.submit(c.orderID, map[string]any{"rating": 5}, token)
	})
	sc.Step(`^the customer has already reviewed the commission$`, func() error {
		if err := c.submit(c.orderID, map[string]any{"rating": 5, "comment": "Original review"}, c.customerToken); err != nil {
			return err
		}
		if err := c.requireStatus(http.StatusCreated); err != nil {
			return err
		}
		c.existingReviews = 1
		return nil
	})
	sc.Step(`^the original review remains unchanged$`, func() error {
		var row reviewResponse
		if err := app.DB.NewSelect().Table("reviews").Where("order_id = ?", c.orderID).Scan(context.Background(), &row); err != nil {
			return err
		}
		if row.Rating != 5 {
			return fmt.Errorf("duplicate submission replaced the rating")
		}
		return checkComment(row.Comment, "Original review")
	})
	sc.Step(`^the customer reviews three completed commissions with ratings 5, 4, and 4$`, func() error {
		for _, rating := range []int{5, 4, 4} {
			orderID, err := c.createOrder("SUCCESS", time.Now().UTC())
			if err != nil {
				return err
			}
			if err := c.submit(orderID, map[string]any{"rating": rating}, c.customerToken); err != nil {
				return err
			}
			if err := c.requireStatus(http.StatusCreated); err != nil {
				return err
			}
		}
		return nil
	})
	sc.Step(`^the artist has reviews whose order dates differ from their review dates$`, func() error {
		return c.prepareReviewHistory()
	})
	sc.Step(`^"([^"]+)" requests the artist reviews$`, func(caller string) error {
		token, err := c.tokenFor(caller)
		if err != nil {
			return err
		}
		return c.listReviews(token)
	})
	sc.Step(`^the reviews contain the customer and order names$`, func() error {
		items, err := c.readReviewList()
		if err != nil {
			return err
		}
		if len(items) == 0 {
			return fmt.Errorf("expected review details")
		}
		for _, item := range items {
			if item.Username != c.customer.Username || item.Order != "Customer commission title" || item.Rating != 5 || item.Comment == nil || *item.Comment != "Original review" || item.CreatedAt.IsZero() || item.ArtistID.String() != c.artist.ID {
				return fmt.Errorf("incorrect review details: %+v", item)
			}
		}
		return nil
	})
	sc.Step(`^the artist reviews are returned newest first$`, func() error {
		items, err := c.readReviewList()
		if err != nil {
			return err
		}
		if len(items) != len(c.expectedIDs) {
			return fmt.Errorf("expected %d reviews, got %d", len(c.expectedIDs), len(items))
		}
		for i, item := range items {
			if item.ID != c.expectedIDs[i] || !item.CreatedAt.Equal(c.expectedDates[i]) {
				return fmt.Errorf("review %d is out of order: got %s at %s, want %s at %s", i, item.ID, item.CreatedAt, c.expectedIDs[i], c.expectedDates[i])
			}
		}
		return nil
	})
	sc.Step(`^reviews from other artists are excluded$`, func() error {
		items, err := c.readReviewList()
		if err != nil {
			return err
		}
		if len(items) != 3 {
			return fmt.Errorf("expected only this artist's three reviews, got %d", len(items))
		}
		for _, item := range items {
			if item.ArtistID.String() != c.artist.ID {
				return fmt.Errorf("a review belonging to another artist was returned")
			}
		}
		return nil
	})
	sc.Step(`^the system returns an empty review list and no average score$`, func() error {
		items, err := c.readReviewList()
		if err != nil {
			return err
		}
		if items == nil || len(items) != 0 {
			return fmt.Errorf("expected reviews: [], got %+v", items)
		}
		return c.requireAverage(nil)
	})

	sc.Step(
		`^the API receives a "([^"]+)" review request with limit (\d+) and offset (\d+)$`,
		func(resource string, limit, offset int) error {
			var path string

			switch resource {
			case "artist":
				path = "/artists/" + c.artist.ID + "/reviews"
			case "artwork":
				path = "/artworks/" + c.artworkID.String() + "/reviews"
			default:
				return fmt.Errorf("unknown review resource %q", resource)
			}

			path += fmt.Sprintf("?limit=%d&offset=%d", limit, offset)

			var err error
			c.response, err = c.client.Do(
				http.MethodGet,
				path,
				nil,
				map[string]string{
					"Authorization": "Bearer " + c.customerToken,
				},
			)
			return err
		},
	)

	sc.Step(
		`^the review page contains (\d+) reviews with total (\d+)$`,
		func(count, total int) error {
			if err := c.requireStatus(http.StatusOK); err != nil {
				return err
			}

			var body struct {
				Reviews []reviewResponse `json:"reviews"`
				Total   *int             `json:"total"`
			}
			if err := c.response.JSON(&body); err != nil {
				return err
			}

			if body.Reviews == nil || len(body.Reviews) != count {
				return fmt.Errorf("expected an array of %d reviews, got %v", count, body.Reviews)
			}
			if body.Total == nil || *body.Total != total {
				return fmt.Errorf("expected total %d, got %v", total, body.Total)
			}

			return nil
		},
	)

	sc.Step(
		`^the review page matches the expected order from offset (\d+)$`,
		func(offset int) error {
			items, err := c.readReviewList()
			if err != nil {
				return err
			}

			for i, item := range items {
				position := offset + i
				if position >= len(c.expectedIDs) {
					return fmt.Errorf("unexpected review at position %d", position)
				}
				if item.ID != c.expectedIDs[position] {
					return fmt.Errorf("incorrect review at position %d", position)
				}
			}

			return nil
		},
	)

	sc.Step(`^"([^"]+)" requests the order review$`, func(caller string) error {
		token, err := c.tokenFor(caller)
		if err != nil {
			return err
		}

		c.response, err = c.client.Do(
			http.MethodGet,
			"/orders/"+c.orderID.String()+"/review",
			nil,
			map[string]string{
				"Authorization": "Bearer " + token,
			},
		)
		return err
	})

	sc.Step(`^the order review response has status (\d+)$`, func(status int) error {
		if err := c.requireStatus(status); err != nil {
			return err
		}

		if status != http.StatusOK {
			return nil
		}

		var item reviewResponse
		if err := c.response.JSON(&item); err != nil {
			return err
		}

		if item.ID == uuid.Nil ||
			item.OrderID != c.orderID ||
			item.CustomerID.String() != c.customer.ID ||
			item.ArtistID.String() != c.artist.ID ||
			item.Rating != 5 ||
			item.CreatedAt.IsZero() {
			return fmt.Errorf("incorrect order review: %+v", item)
		}

		return checkComment(item.Comment, "Original review")
	})

	sc.Step(`^another artwork by the same artist has a review$`, func() error {
		other := *c
		other.artworkID = uuid.New()

		_, err := app.DB.ExecContext(context.Background(), `
		INSERT INTO artworks (
			id, artist_id, category_id, name, description,
			price_satang, minimum_deadline_days
		)
		VALUES (?, ?, ?, 'Another artwork', 'Review fixture', 1000, 7)
	`, other.artworkID, c.artist.ID, c.categoryID)
		if err != nil {
			return err
		}

		orderID, err := other.createOrder("SUCCESS", time.Now().UTC())
		if err != nil {
			return err
		}

		if err := other.submit(
			orderID,
			map[string]any{"rating": 1},
			c.customerToken,
		); err != nil {
			return err
		}

		return other.requireStatus(http.StatusCreated)
	})
}

func checkComment(got *string, want string) error {
	if want == "NULL" {
		if got != nil {
			return fmt.Errorf("expected a NULL comment, got %q", *got)
		}
	} else if got == nil || *got != want {
		return fmt.Errorf("expected comment %q, got %v", want, got)
	}
	return nil
}

func (c *reviewsContext) tokenFor(caller string) (string, error) {
	switch caller {
	case "the customer":
		return c.customerToken, nil
	case "the artist":
		return apptest.Login(c.client, c.artist.Email, c.artist.Password)
	case "another customer":
		other, err := apptest.RegisterCustomer(app, c.client)
		if err != nil {
			return "", err
		}
		return apptest.Login(c.client, other.Email, other.Password)
	default:
		return "", fmt.Errorf("unknown caller %q", caller)
	}
}

func (c *reviewsContext) readReviewList() ([]reviewResponse, error) {
	if err := c.requireStatus(http.StatusOK); err != nil {
		return nil, err
	}
	var body struct {
		Reviews []reviewResponse `json:"reviews"`
		Total   *int             `json:"total"`
	}
	if err := c.response.JSON(&body); err != nil {
		return nil, err
	}
	if body.Total == nil || *body.Total < len(body.Reviews) {
		return nil, fmt.Errorf("total must be present and not less than the page size")
	}
	return body.Reviews, nil
}

// prepareReviewHistory makes order chronology disagree with review chronology.
func (c *reviewsContext) prepareReviewHistory() error {
	base := time.Date(2026, 10, 1, 10, 0, 0, 0, time.UTC)
	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	dates := []time.Time{base, base.Add(2 * time.Hour), base.Add(time.Hour)}
	for i := range ids {
		orderID, err := c.createOrder("SUCCESS", base.Add(time.Duration(3-i)*time.Hour))
		if err != nil {
			return err
		}
		if err := c.submit(orderID, map[string]any{"rating": 5, "comment": "Original review"}, c.customerToken); err != nil {
			return err
		}
		if err := c.requireStatus(http.StatusCreated); err != nil {
			return err
		}
		var item reviewResponse
		if err := c.response.JSON(&item); err != nil {
			return err
		}
		ids[i] = item.ID
		// Fixed review times keep the sorting check independent of machine speed.
		if _, err := app.DB.ExecContext(context.Background(), "UPDATE reviews SET created_at = ? WHERE order_id = ?", dates[i], orderID); err != nil {
			return err
		}
	}
	c.expectedIDs = []uuid.UUID{ids[1], ids[2], ids[0]}
	c.expectedDates = []time.Time{dates[1], dates[2], dates[0]}

	other := &reviewsContext{client: apptest.NewClient(app.BaseURL())}
	if err := other.prepareCommission("SUCCESS"); err != nil {
		return err
	}
	if err := other.submit(other.orderID, map[string]any{"rating": 1}, other.customerToken); err != nil {
		return err
	}
	return other.requireStatus(http.StatusCreated)
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
	var stored struct {
		Score *float64 `bun:"review_score"`
	}

	if err := app.DB.NewSelect().
		Table("artist_profiles").
		Column("review_score").
		Where("user_id = ?", c.artist.ID).
		Scan(context.Background(), &stored); err != nil {
		return err
	}

	if want == nil {
		if stored.Score != nil {
			return fmt.Errorf("expected stored score to be NULL, got %v", *stored.Score)
		}
	} else {
		if stored.Score == nil {
			return fmt.Errorf("expected stored score %v, got NULL", *want)
		}
		if *stored.Score != *want {
			return fmt.Errorf("stored score = %v, want %v", *stored.Score, *want)
		}
	}

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
